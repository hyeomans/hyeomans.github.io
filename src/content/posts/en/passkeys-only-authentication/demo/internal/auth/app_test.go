package auth

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

func TestSummarizeCredential(t *testing.T) {
	credential := &webauthn.Credential{
		ID:                []byte("credential-id"),
		PublicKey:         []byte("public-key"),
		AttestationType:   "none",
		AttestationFormat: "none",
		Transport:         []protocol.AuthenticatorTransport{"internal"},
		Flags:             webauthn.NewCredentialFlags(protocol.FlagUserPresent | protocol.FlagUserVerified | protocol.FlagBackupEligible),
		Authenticator: webauthn.Authenticator{
			AAGUID:       []byte("aaguid"),
			Attachment:   protocol.Platform,
			SignCount:    7,
			CloneWarning: true,
		},
	}

	summary := summarizeCredential(credential)

	if summary.IDBase64URL != base64.RawURLEncoding.EncodeToString(credential.ID) {
		t.Fatalf("IDBase64URL = %q", summary.IDBase64URL)
	}
	if summary.PublicKeyBase64URL != base64.RawURLEncoding.EncodeToString(credential.PublicKey) {
		t.Fatalf("PublicKeyBase64URL = %q", summary.PublicKeyBase64URL)
	}
	if !summary.Flags.UserVerified || !summary.Flags.BackupEligible {
		t.Fatalf("Flags = %+v", summary.Flags)
	}
	if summary.Authenticator.SignCount != 7 || !summary.Authenticator.CloneWarning {
		t.Fatalf("Authenticator = %+v", summary.Authenticator)
	}
}

func TestProfileRequiresSessionAndCSRFToken(t *testing.T) {
	app, err := New(Config{
		RPDisplayName: "Passkey Workshop",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8080"},
	})
	if err != nil {
		t.Fatal(err)
	}
	app.store.users["reader@example.com"] = &User{
		ID:          randomBytes(32),
		Email:       "reader@example.com",
		ProfileName: "Reader",
	}
	sessionToken, session := app.createSession("reader@example.com")
	handler := app.Routes()

	tests := []struct {
		name       string
		cookie     bool
		csrfToken  string
		wantStatus int
	}{
		{name: "no session", wantStatus: http.StatusUnauthorized},
		{name: "session but no CSRF token", cookie: true, wantStatus: http.StatusForbidden},
		{name: "wrong CSRF token", cookie: true, csrfToken: "wrong", wantStatus: http.StatusForbidden},
		{name: "matching CSRF token", cookie: true, csrfToken: session.CSRFToken, wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/profile", strings.NewReader(`{"profileName":"New name"}`))
			req.Header.Set("Content-Type", "application/json")
			if tt.cookie {
				req.AddCookie(&http.Cookie{Name: sessionCookie, Value: sessionToken})
			}
			if tt.csrfToken != "" {
				req.Header.Set("X-CSRF-Token", tt.csrfToken)
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, req)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, tt.wantStatus, response.Body.String())
			}
		})
	}
}

func TestBeginLoginRequiresExactJSONMediaType(t *testing.T) {
	app, err := New(Config{
		RPDisplayName: "Passkey Workshop",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8080"},
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/passkeys/login/begin", strings.NewReader(`{"email":"reader@example.com"}`))
	req.Header.Set("Content-Type", "application/jsonp")
	response := httptest.NewRecorder()

	app.Routes().ServeHTTP(response, req)

	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusUnsupportedMediaType, response.Body.String())
	}
}

func TestStoreRejectsCredentialOwnedByAnotherUser(t *testing.T) {
	store := newStore()
	store.users["one@example.com"] = &User{ID: randomBytes(32), Email: "one@example.com"}
	store.users["two@example.com"] = &User{
		ID:          randomBytes(32),
		Email:       "two@example.com",
		Credentials: []webauthn.Credential{{ID: []byte("same-credential")}},
	}

	err := store.addFirstCredential("one@example.com", webauthn.Credential{ID: []byte("same-credential")})
	if err == nil {
		t.Fatal("expected duplicate credential to be rejected")
	}
}
