package githubapp

import (
	"crypto/rsa"
	"log/slog"
	"sync"
	"time"
)

// installationToken is the short-lived credential cached for one installation.
type installationToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// App wires GitHub authentication to the local HTTP API.
type App struct {
	cfg        Config
	http       *httpClient
	privateKey *rsa.PrivateKey
	log        *slog.Logger

	mu                   sync.Mutex
	cachedInstallationID int64
	installationTokens   map[int64]installationToken
}

func New(cfg Config) (*App, error) {
	if cfg.APIBaseURL == "" {
		cfg.APIBaseURL = DefaultAPIBaseURL
	}
	if cfg.StaticDir == "" {
		cfg.StaticDir = "static"
	}

	key, err := parsePrivateKey(cfg.PrivateKeyPEM)
	if err != nil {
		return nil, err
	}

	return &App{
		cfg:                cfg,
		http:               newHTTPClient(cfg.APIBaseURL),
		privateKey:         key,
		log:                loggerOrDefault(cfg.Logger),
		installationTokens: make(map[int64]installationToken),
	}, nil
}
