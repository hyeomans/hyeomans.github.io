package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type credentialSummary struct {
	IDBase64URL        string                            `json:"idBase64URL"`
	PublicKeyBase64URL string                            `json:"publicKeyBase64URL"`
	AttestationType    string                            `json:"attestationType"`
	AttestationFormat  string                            `json:"attestationFormat"`
	Transports         []protocol.AuthenticatorTransport `json:"transports"`
	Flags              webauthn.CredentialFlags          `json:"flags"`
	Authenticator      credentialAuthenticatorSummary    `json:"authenticator"`
}

type credentialAuthenticatorSummary struct {
	AAGUIDBase64URL string                           `json:"aaguidBase64URL"`
	Attachment      protocol.AuthenticatorAttachment `json:"attachment"`
	SignCount       uint32                           `json:"signCount"`
	CloneWarning    bool                             `json:"cloneWarning"`
}

func summarizeCredential(credential *webauthn.Credential) credentialSummary {
	return credentialSummary{
		IDBase64URL:        base64.RawURLEncoding.EncodeToString(credential.ID),
		PublicKeyBase64URL: base64.RawURLEncoding.EncodeToString(credential.PublicKey),
		AttestationType:    credential.AttestationType,
		AttestationFormat:  credential.AttestationFormat,
		Transports:         credential.Transport,
		Flags:              credential.Flags,
		Authenticator: credentialAuthenticatorSummary{
			AAGUIDBase64URL: base64.RawURLEncoding.EncodeToString(credential.Authenticator.AAGUID),
			Attachment:      credential.Authenticator.Attachment,
			SignCount:       credential.Authenticator.SignCount,
			CloneWarning:    credential.Authenticator.CloneWarning,
		},
	}
}

func readEmail(r *http.Request) (string, error) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return "", errors.New("invalid JSON")
	}
	email := strings.ToLower(strings.TrimSpace(body.Email))
	if email == "" || len(email) > 254 || !strings.Contains(email, "@") {
		return "", errors.New("enter a valid email address")
	}
	return email, nil
}

func isJSON(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && strings.EqualFold(mediaType, "application/json")
}

func randomBytes(size int) []byte {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Errorf("crypto/rand failed: %w", err))
	}
	return b
}

func randomToken(size int) string {
	return base64.RawURLEncoding.EncodeToString(randomBytes(size))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
