package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const demoBearerToken = "test-token"

// appleReceiver is the Apple side of this public-protocol example. It uses
// transmitter metadata supplied by the IdP; it does not model Apple's private
// implementation or use an Apple production endpoint.
type appleReceiver struct {
	pollEndpoint     string
	bearerToken      string
	expectedIssuer   string
	expectedAudience string
	keyID            string
	publicKey        ed25519.PublicKey
}

type transmitterMetadata struct {
	Issuer                   string   `json:"issuer"`
	JWKSURI                  string   `json:"jwks_uri"`
	DeliveryMethodsSupported []string `json:"delivery_methods_supported"`
	ConfigurationEndpoint    string   `json:"configuration_endpoint"`
}

type jwksDocument struct {
	Keys []struct {
		KeyType string `json:"kty"`
		Curve   string `json:"crv"`
		Use     string `json:"use"`
		Alg     string `json:"alg"`
		KeyID   string `json:"kid"`
		X       string `json:"x"`
	} `json:"keys"`
}

func TestAppleReceiverPollsIdentityProviderUntilAcknowledged(t *testing.T) {
	db, err := openDatabase(t.TempDir() + "/ssf.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_787_337_600, 0)

	// IDP SIDE: the company's identity provider creates, signs, queues, and
	// serves SETs. In the Apple Business example, this is the SSF transmitter.
	identityProvider := &server{
		db:         db,
		privateKey: privateKey,
		config: config{
			audience:        demoAppleAudience,
			bearerToken:     demoBearerToken,
			keyID:           "test-key",
			retryAfter:      30 * time.Second,
			longPollTimeout: time.Second,
		},
		now: func() time.Time { return now },
	}
	idpHTTPServer := httptest.NewServer(identityProvider.routes())
	t.Cleanup(idpHTTPServer.Close)
	identityProvider.config.issuer = idpHTTPServer.URL

	// APPLE SIDE: Apple starts with the IdP's SSF configuration URL. It uses
	// discovery to learn the stream endpoint and download the IdP's public key.
	apple := connectAppleReceiver(
		t,
		idpHTTPServer.URL+"/.well-known/ssf-configuration",
		demoBearerToken,
	)

	// An IdP account service reports that the employee session was revoked.
	// /emit is an internal demo input; Apple never calls it.
	emit := postJSON(
		t,
		idpHTTPServer.URL+"/emit",
		demoBearerToken,
		`{"subject_id":"session-123"}`,
	)
	if emit.StatusCode != http.StatusCreated {
		t.Fatalf("emit status = %d", emit.StatusCode)
	}
	var emitted struct {
		JTI string `json:"jti"`
	}
	decodeResponse(t, emit, &emitted)

	// Apple polls the IdP, then verifies the SET before applying local policy.
	first := apple.poll(t, `{"maxEvents":10,"returnImmediately":true}`)
	token := first.Sets[emitted.JTI]
	if token == "" {
		t.Fatalf("first poll did not include %s", emitted.JTI)
	}
	apple.verifySET(t, token, emitted.JTI, "session-123")

	// The IdP keeps the SET until Apple acknowledges it, so a missed response
	// does not lose the security event.
	assertQueueSize(t, db, 1)
	now = now.Add(31 * time.Second)
	second := apple.poll(t, `{"maxEvents":10,"returnImmediately":true}`)
	if second.Sets[emitted.JTI] != token {
		t.Fatal("unacknowledged SET was not redelivered")
	}

	ackBody := `{"ack":["` + emitted.JTI + `"],"maxEvents":0,"returnImmediately":true}`
	ack := apple.poll(t, ackBody)
	if len(ack.Sets) != 0 {
		t.Fatalf("ack-only response contained %d SETs", len(ack.Sets))
	}
	assertQueueSize(t, db, 0)

	empty := apple.poll(t, `{"maxEvents":10,"returnImmediately":true}`)
	if len(empty.Sets) != 0 {
		t.Fatalf("poll after ack contained %d SETs", len(empty.Sets))
	}
}

func connectAppleReceiver(t *testing.T, configurationURL, bearerToken string) appleReceiver {
	t.Helper()

	metadataResponse := get(t, configurationURL, "")
	if metadataResponse.StatusCode != http.StatusOK {
		t.Fatalf("discovery status = %d", metadataResponse.StatusCode)
	}
	var metadata transmitterMetadata
	decodeResponse(t, metadataResponse, &metadata)
	if metadata.Issuer == "" || metadata.JWKSURI == "" || metadata.ConfigurationEndpoint == "" {
		t.Fatalf("incomplete transmitter metadata: %#v", metadata)
	}
	if len(metadata.DeliveryMethodsSupported) != 1 || metadata.DeliveryMethodsSupported[0] != pollDeliveryMethod {
		t.Fatalf("unexpected delivery methods: %#v", metadata.DeliveryMethodsSupported)
	}

	jwksResponse := get(t, metadata.JWKSURI, "")
	if jwksResponse.StatusCode != http.StatusOK {
		t.Fatalf("JWKS status = %d", jwksResponse.StatusCode)
	}
	var jwks jwksDocument
	decodeResponse(t, jwksResponse, &jwks)
	if len(jwks.Keys) != 1 {
		t.Fatalf("JWKS contains %d keys", len(jwks.Keys))
	}
	key := jwks.Keys[0]
	if key.KeyType != "OKP" || key.Curve != "Ed25519" || key.Use != "sig" || key.Alg != "EdDSA" {
		t.Fatalf("unexpected JWK: %#v", key)
	}
	publicKey, err := base64.RawURLEncoding.DecodeString(key.X)
	if err != nil {
		t.Fatal(err)
	}
	if len(publicKey) != ed25519.PublicKeySize {
		t.Fatalf("public key has %d bytes", len(publicKey))
	}

	streamResponse := get(t, metadata.ConfigurationEndpoint, bearerToken)
	if streamResponse.StatusCode != http.StatusOK {
		t.Fatalf("stream configuration status = %d", streamResponse.StatusCode)
	}
	var streams []streamConfig
	decodeResponse(t, streamResponse, &streams)
	if len(streams) != 1 {
		t.Fatalf("stream configuration contains %d streams", len(streams))
	}
	stream := streams[0]
	if stream.Issuer != metadata.Issuer {
		t.Fatalf("stream issuer = %q, want %q", stream.Issuer, metadata.Issuer)
	}
	if len(stream.Audience) != 1 || stream.Audience[0] != demoAppleAudience {
		t.Fatalf("stream audience = %#v", stream.Audience)
	}
	if stream.Delivery["method"] != pollDeliveryMethod {
		t.Fatalf("delivery method = %v", stream.Delivery["method"])
	}
	pollEndpoint, ok := stream.Delivery["endpoint_url"].(string)
	if !ok || pollEndpoint == "" {
		t.Fatalf("missing poll endpoint: %#v", stream.Delivery)
	}

	return appleReceiver{
		pollEndpoint:     pollEndpoint,
		bearerToken:      bearerToken,
		expectedIssuer:   metadata.Issuer,
		expectedAudience: demoAppleAudience,
		keyID:            key.KeyID,
		publicKey:        ed25519.PublicKey(publicKey),
	}
}

func (a appleReceiver) poll(t *testing.T, body string) pollResponse {
	t.Helper()
	response := postJSON(t, a.pollEndpoint, a.bearerToken, body)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("poll status = %d", response.StatusCode)
	}
	var result pollResponse
	decodeResponse(t, response, &result)
	return result
}

func (a appleReceiver) verifySET(t *testing.T, token, expectedJTI, expectedSubjectID string) {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("SET has %d parts", len(parts))
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatal(err)
	}
	if !ed25519.Verify(a.publicKey, []byte(parts[0]+"."+parts[1]), signature) {
		t.Fatal("SET signature is invalid")
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatal(err)
	}
	var header map[string]any
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		t.Fatal(err)
	}
	if header["typ"] != "secevent+jwt" || header["alg"] != "EdDSA" || header["kid"] != a.keyID {
		t.Fatalf("unexpected JOSE header: %#v", header)
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(claimsJSON))
	decoder.UseNumber()
	var claims map[string]any
	if err := decoder.Decode(&claims); err != nil {
		t.Fatal(err)
	}
	if claims["iss"] != a.expectedIssuer {
		t.Fatalf("iss = %v, want %q", claims["iss"], a.expectedIssuer)
	}
	if claims["aud"] != a.expectedAudience {
		t.Fatalf("aud = %v, want %q", claims["aud"], a.expectedAudience)
	}
	if claims["jti"] != expectedJTI {
		t.Fatalf("jti = %v", claims["jti"])
	}
	if _, exists := claims["sub"]; exists {
		t.Fatal("SSF SET must not contain sub")
	}
	if _, exists := claims["exp"]; exists {
		t.Fatal("SSF SET must not contain exp")
	}
	subject, ok := claims["sub_id"].(map[string]any)
	if !ok || subject["format"] != "opaque" || subject["id"] != expectedSubjectID {
		t.Fatalf("unexpected subject: %#v", claims["sub_id"])
	}
	events, ok := claims["events"].(map[string]any)
	if !ok || events[caepSessionRevoked] == nil {
		t.Fatalf("missing CAEP session-revoked event: %#v", claims["events"])
	}
}

func get(t *testing.T, url, bearerToken string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	if bearerToken != "" {
		request.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func postJSON(t *testing.T, url, bearerToken, body string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+bearerToken)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func decodeResponse(t *testing.T, response *http.Response, destination any) {
	t.Helper()
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		t.Fatal(err)
	}
}

func assertQueueSize(t *testing.T, db interface {
	QueryRow(query string, args ...any) *sql.Row
}, expected int) {
	t.Helper()
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM ssf_events`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != expected {
		t.Fatalf("queue size = %d, want %d", count, expected)
	}
}
