package githubapp

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const DefaultAPIBaseURL = "https://api.github.com"

// Config holds every value the app needs to talk to GitHub.
type Config struct {
	// ClientID is the issuer of the app JWT.
	ClientID string
	// PrivateKeyPEM is the RSA private key downloaded from the app settings.
	PrivateKeyPEM []byte
	// Org is the installation account displayed by the demo.
	Org        string
	APIBaseURL string
	StaticDir  string
	// Logger receives debug and error logs. If nil, New installs a discard
	// logger. ConfigFromEnv builds one from LOG_LEVEL.
	Logger *slog.Logger
}

// ConfigFromEnv reads configuration from a local .env file and the
// environment, then fails fast when a required value is missing. godotenv does
// not override variables that are already set, so a real environment value
// still wins over the file.
func ConfigFromEnv() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		ClientID:   os.Getenv("GITHUB_CLIENT_ID"),
		Org:        os.Getenv("GITHUB_ORG"),
		APIBaseURL: envOr("GITHUB_API_BASE_URL", DefaultAPIBaseURL),
		StaticDir:  envOr("GITHUB_STATIC_DIR", "static"),
		Logger:     newLogger(envOr("LOG_LEVEL", "info")),
	}

	keyPath := os.Getenv("GITHUB_PRIVATE_KEY_PATH")
	if keyPath == "" {
		return Config{}, fmt.Errorf("GITHUB_PRIVATE_KEY_PATH is required")
	}
	pemBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return Config{}, fmt.Errorf("read private key: %w", err)
	}
	cfg.PrivateKeyPEM = pemBytes

	if cfg.ClientID == "" {
		return Config{}, fmt.Errorf("GITHUB_CLIENT_ID is required")
	}
	return cfg, nil
}

// newLogger builds a stderr logger at the requested level. Set LOG_LEVEL=debug
// to see token activity.
func newLogger(level string) *slog.Logger {
	var parsed slog.Level
	switch strings.ToLower(level) {
	case "debug":
		parsed = slog.LevelDebug
	case "warn":
		parsed = slog.LevelWarn
	case "error":
		parsed = slog.LevelError
	default:
		parsed = slog.LevelInfo
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: parsed}))
}

// loggerOrDefault keeps tests quiet when no logger is supplied.
func loggerOrDefault(logger *slog.Logger) *slog.Logger {
	if logger != nil {
		return logger
	}
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
