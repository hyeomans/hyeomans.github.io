package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	caepSessionRevoked = "https://schemas.openid.net/secevent/caep/event-type/session-revoked"
	pollDeliveryMethod = "urn:ietf:rfc:8936"
	// This reserved example URL labels Apple as the receiver in the demo. It is
	// not an Apple production endpoint.
	demoAppleAudience = "https://apple.example/ssf"
)

type config struct {
	issuer          string
	audience        string
	bearerToken     string
	keyID           string
	retryAfter      time.Duration
	longPollTimeout time.Duration
}

// server is the company identity provider in the article's Apple example.
// The identity provider operates the SSF transmitter; Apple is the receiver.
type server struct {
	db         *sql.DB
	privateKey ed25519.PrivateKey
	config     config
	now        func() time.Time
}

type pollRequest struct {
	Ack               []string            `json:"ack,omitempty"`
	SetErrors         map[string]setError `json:"setErrs,omitempty"`
	MaxEvents         *int                `json:"maxEvents,omitempty"`
	ReturnImmediately bool                `json:"returnImmediately,omitempty"`
}

type setError struct {
	Error       string `json:"err"`
	Description string `json:"description,omitempty"`
}

type pollResponse struct {
	Sets          map[string]string `json:"sets"`
	MoreAvailable bool              `json:"moreAvailable,omitempty"`
}

type streamConfig struct {
	StreamID        string         `json:"stream_id"`
	Issuer          string         `json:"iss"`
	Audience        []string       `json:"aud"`
	EventsSupported []string       `json:"events_supported"`
	EventsRequested []string       `json:"events_requested"`
	EventsDelivered []string       `json:"events_delivered"`
	Delivery        map[string]any `json:"delivery"`
}

func main() {
	address := envOr("SSF_ADDR", ":8080")
	issuer := envOr("SSF_ISSUER", "http://localhost:8080")
	bearerToken := os.Getenv("SSF_BEARER_TOKEN")
	if bearerToken == "" {
		log.Fatal("SSF_BEARER_TOKEN is required")
	}

	db, err := openDatabase(envOr("SSF_DATABASE", "ssf.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	privateKey, err := loadOrCreatePrivateKey(envOr("SSF_SIGNING_KEY", "ssf-ed25519.key"))
	if err != nil {
		log.Fatal(err)
	}

	identityProvider := &server{
		db:         db,
		privateKey: privateKey,
		config: config{
			issuer:          strings.TrimRight(issuer, "/"),
			audience:        envOr("SSF_AUDIENCE", demoAppleAudience),
			bearerToken:     bearerToken,
			keyID:           "demo-key-1",
			retryAfter:      30 * time.Second,
			longPollTimeout: 15 * time.Second,
		},
		now: time.Now,
	}

	httpServer := &http.Server{
		Addr:              address,
		Handler:           identityProvider.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("identity-provider SSF transmitter listening on %s", address)
	log.Printf("issuer: %s", identityProvider.config.issuer)
	log.Fatal(httpServer.ListenAndServe())
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /.well-known/ssf-configuration", s.handleDiscovery)
	mux.HandleFunc("GET /jwks.json", s.handleJWKS)
	mux.HandleFunc("GET /ssf/stream", s.requireBearer(s.handleStream))
	mux.HandleFunc("POST /events", s.requireBearer(s.handlePoll))
	// /emit represents an internal IdP account or risk service, not an SSF API.
	mux.HandleFunc("POST /emit", s.requireBearer(s.handleEmit))
	return mux
}

func openDatabase(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// One connection keeps this small demo's queue operations easy to reason about.
	db.SetMaxOpenConns(1)

	const schema = `
CREATE TABLE IF NOT EXISTS ssf_events (
    jti              TEXT PRIMARY KEY,
    set_token        TEXT NOT NULL,
    status           TEXT NOT NULL DEFAULT 'queued'
                     CHECK (status IN ('queued', 'failed')),
    created_at       INTEGER NOT NULL,
    deliver_after    INTEGER NOT NULL DEFAULT 0,
    attempts         INTEGER NOT NULL DEFAULT 0,
    last_error       TEXT
);
CREATE INDEX IF NOT EXISTS ssf_events_ready
    ON ssf_events(status, deliver_after, created_at);`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create queue schema: %w", err)
	}
	return db, nil
}

func (s *server) handleDiscovery(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"spec_version":               "1_0",
		"issuer":                     s.config.issuer,
		"jwks_uri":                   s.config.issuer + "/jwks.json",
		"delivery_methods_supported": []string{pollDeliveryMethod},
		"configuration_endpoint":     s.config.issuer + "/ssf/stream",
	})
}

func (s *server) handleJWKS(w http.ResponseWriter, _ *http.Request) {
	publicKey := s.privateKey.Public().(ed25519.PublicKey)
	writeJSON(w, http.StatusOK, map[string]any{
		"keys": []map[string]string{{
			"kty": "OKP",
			"crv": "Ed25519",
			"use": "sig",
			"alg": "EdDSA",
			"kid": s.config.keyID,
			"x":   base64.RawURLEncoding.EncodeToString(publicKey),
		}},
	})
}

func (s *server) handleStream(w http.ResponseWriter, r *http.Request) {
	stream := streamConfig{
		StreamID:        "demo-stream",
		Issuer:          s.config.issuer,
		Audience:        []string{s.config.audience},
		EventsSupported: []string{caepSessionRevoked},
		EventsRequested: []string{caepSessionRevoked},
		EventsDelivered: []string{caepSessionRevoked},
		Delivery: map[string]any{
			"method":       pollDeliveryMethod,
			"endpoint_url": s.config.issuer + "/events",
		},
	}
	streamID := r.URL.Query().Get("stream_id")
	if streamID == "" {
		writeJSON(w, http.StatusOK, []streamConfig{stream})
		return
	}
	if streamID != stream.StreamID {
		http.Error(w, "stream not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, stream)
}

func (s *server) handleEmit(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SubjectID string `json:"subject_id"`
	}
	if err := decodeJSON(r.Body, &input); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(input.SubjectID) == "" {
		http.Error(w, "subject_id is required", http.StatusBadRequest)
		return
	}

	jti, token, err := s.makeSessionRevokedSET(input.SubjectID)
	if err != nil {
		http.Error(w, "create SET", http.StatusInternalServerError)
		return
	}
	if _, err := s.db.Exec(
		`INSERT INTO ssf_events (jti, set_token, created_at) VALUES (?, ?, ?)`,
		jti, token, s.now().Unix(),
	); err != nil {
		http.Error(w, "queue SET", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"jti": jti, "queued": true})
}

func (s *server) handlePoll(w http.ResponseWriter, r *http.Request) {
	var request pollRequest
	if err := decodeJSON(r.Body, &request); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := validatePollRequest(request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := s.poll(r.Context(), request)
	if err != nil {
		http.Error(w, "poll queue", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, response)
}

func (s *server) poll(ctx context.Context, request pollRequest) (pollResponse, error) {
	response, err := s.applyFeedbackAndClaim(ctx, request)
	if err != nil || len(response.Sets) > 0 || request.ReturnImmediately || maxEvents(request) == 0 {
		return response, err
	}

	deadline := time.NewTimer(s.config.longPollTimeout)
	defer deadline.Stop()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	request.Ack = nil
	request.SetErrors = nil
	for {
		select {
		case <-ctx.Done():
			return pollResponse{}, ctx.Err()
		case <-deadline.C:
			return pollResponse{Sets: map[string]string{}}, nil
		case <-ticker.C:
			response, err = s.applyFeedbackAndClaim(ctx, request)
			if err != nil || len(response.Sets) > 0 {
				return response, err
			}
		}
	}
}

func (s *server) applyFeedbackAndClaim(ctx context.Context, request pollRequest) (pollResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return pollResponse{}, err
	}
	defer tx.Rollback()

	for _, jti := range request.Ack {
		if _, err := tx.ExecContext(ctx, `DELETE FROM ssf_events WHERE jti = ?`, jti); err != nil {
			return pollResponse{}, err
		}
	}
	for jti, setErr := range request.SetErrors {
		message := strings.TrimSpace(setErr.Error + ": " + setErr.Description)
		if _, err := tx.ExecContext(ctx,
			`UPDATE ssf_events SET status = 'failed', last_error = ? WHERE jti = ?`,
			message, jti,
		); err != nil {
			return pollResponse{}, err
		}
	}

	sets := map[string]string{}
	limit := maxEvents(request)
	if limit > 0 {
		now := s.now()
		rows, err := tx.QueryContext(ctx, `
WITH ready AS (
    SELECT jti
    FROM ssf_events
    WHERE status = 'queued' AND deliver_after <= ?
    ORDER BY created_at, jti
    LIMIT ?
)
UPDATE ssf_events
SET deliver_after = ?, attempts = attempts + 1
WHERE jti IN (SELECT jti FROM ready)
RETURNING jti, set_token`, now.Unix(), limit, now.Add(s.config.retryAfter).Unix())
		if err != nil {
			return pollResponse{}, err
		}
		for rows.Next() {
			var jti, token string
			if err := rows.Scan(&jti, &token); err != nil {
				rows.Close()
				return pollResponse{}, err
			}
			sets[jti] = token
		}
		if err := rows.Close(); err != nil {
			return pollResponse{}, err
		}
	}

	var readyCount int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM ssf_events WHERE status = 'queued' AND deliver_after <= ?`,
		s.now().Unix(),
	).Scan(&readyCount); err != nil {
		return pollResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return pollResponse{}, err
	}
	return pollResponse{Sets: sets, MoreAvailable: readyCount > 0}, nil
}

func maxEvents(request pollRequest) int {
	if request.MaxEvents == nil {
		return 100 // A transmitter-side batch cap keeps responses bounded.
	}
	return *request.MaxEvents
}

func validatePollRequest(request pollRequest) error {
	if request.MaxEvents != nil && (*request.MaxEvents < 0 || *request.MaxEvents > 100) {
		return errors.New("maxEvents must be between 0 and 100")
	}
	seen := make(map[string]struct{}, len(request.Ack))
	for _, jti := range request.Ack {
		if jti == "" {
			return errors.New("ack values cannot be empty")
		}
		seen[jti] = struct{}{}
	}
	for jti, setErr := range request.SetErrors {
		if jti == "" || setErr.Error == "" {
			return errors.New("setErrs requires a jti and err value")
		}
		if _, exists := seen[jti]; exists {
			return errors.New("the same jti cannot appear in ack and setErrs")
		}
	}
	return nil
}

func (s *server) makeSessionRevokedSET(subjectID string) (string, string, error) {
	jti, err := randomID()
	if err != nil {
		return "", "", err
	}
	txn, err := randomID()
	if err != nil {
		return "", "", err
	}
	now := s.now().Unix()
	header := map[string]any{
		"alg": "EdDSA",
		"kid": s.config.keyID,
		"typ": "secevent+jwt",
	}
	claims := map[string]any{
		"iss": s.config.issuer,
		"aud": s.config.audience,
		"iat": now,
		"jti": jti,
		"txn": txn,
		"sub_id": map[string]string{
			"format": "opaque",
			"id":     subjectID,
		},
		"events": map[string]any{
			caepSessionRevoked: map[string]any{"event_timestamp": now},
		},
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", "", err
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", "", err
	}
	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedClaims := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signingInput := encodedHeader + "." + encodedClaims
	signature := ed25519.Sign(s.privateKey, []byte(signingInput))
	token := signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
	return jti, token, nil
}

func (s *server) requireBearer(next http.HandlerFunc) http.HandlerFunc {
	expected := []byte("Bearer " + s.config.bearerToken)
	return func(w http.ResponseWriter, r *http.Request) {
		provided := []byte(r.Header.Get("Authorization"))
		if len(provided) != len(expected) || subtle.ConstantTimeCompare(provided, expected) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func loadOrCreatePrivateKey(path string) (ed25519.PrivateKey, error) {
	key, err := os.ReadFile(path)
	if err == nil {
		if len(key) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("%s does not contain an Ed25519 private key", path)
		}
		return ed25519.PrivateKey(key), nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read signing key: %w", err)
	}
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate signing key: %w", err)
	}
	if err := os.WriteFile(path, privateKey, 0o600); err != nil {
		return nil, fmt.Errorf("save signing key: %w", err)
	}
	return privateKey, nil
}

func randomID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}

func decodeJSON(body io.Reader, destination any) error {
	decoder := json.NewDecoder(io.LimitReader(body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode JSON response: %v", err)
	}
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
