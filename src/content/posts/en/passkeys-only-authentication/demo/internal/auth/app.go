package auth

import (
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

const (
	ceremonyCookie = "passkey_ceremony"
	sessionCookie  = "session"
	ceremonyTTL    = 5 * time.Minute
	sessionTTL     = 24 * time.Hour
)

// Config holds the relying-party settings for one deployment.
type Config struct {
	RPDisplayName string
	RPID          string
	RPOrigins     []string
	StaticDir     string
}

// App wires the WebAuthn library to the HTTP handlers.
type App struct {
	webAuthn  *webauthn.WebAuthn
	store     *Store
	staticDir string
}

// New builds an App from config.
func New(cfg Config) (*App, error) {
	web, err := webauthn.New(&webauthn.Config{
		RPDisplayName:         cfg.RPDisplayName,
		RPID:                  cfg.RPID,
		RPOrigins:             cfg.RPOrigins,
		AttestationPreference: protocol.PreferNoAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			ResidentKey:      protocol.ResidentKeyRequirementRequired,
			UserVerification: protocol.VerificationRequired,
		},
	})
	if err != nil {
		return nil, err
	}

	if cfg.StaticDir == "" {
		cfg.StaticDir = "static"
	}

	return &App{webAuthn: web, store: newStore(), staticDir: cfg.StaticDir}, nil
}

// Routes returns the full HTTP handler.
func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/passkeys/registration/begin", a.beginRegistration)
	mux.HandleFunc("POST /api/passkeys/registration/finish", a.finishRegistration)
	mux.HandleFunc("POST /api/passkeys/login/begin", a.beginLogin)
	mux.HandleFunc("POST /api/passkeys/login/finish", a.finishLogin)
	mux.HandleFunc("GET /api/session", a.requireSession(a.getSession))
	mux.HandleFunc("POST /api/profile", a.requireSession(a.requireCSRF(a.updateProfile)))
	mux.HandleFunc("POST /api/logout", a.requireSession(a.requireCSRF(a.logout)))
	mux.Handle("/", http.FileServer(http.Dir(a.staticDir)))
	return securityHeaders(mux)
}
