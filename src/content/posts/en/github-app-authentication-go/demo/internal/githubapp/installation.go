package githubapp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// appMetadata is the public identity GitHub returns for the authenticated app.
type appMetadata struct {
	ID          int64             `json:"id"`
	ClientID    string            `json:"client_id"`
	Slug        string            `json:"slug"`
	Name        string            `json:"name"`
	HTMLURL     string            `json:"html_url"`
	Permissions map[string]string `json:"permissions"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
}

// getAppMetadata calls an app-level endpoint with the short-lived JWT.
func (a *App) getAppMetadata(ctx context.Context) (appMetadata, error) {
	appJWT, err := signAppJWT(a.cfg.ClientID, a.privateKey, time.Now())
	if err != nil {
		return appMetadata{}, err
	}
	data, err := a.http.json(ctx, "GET", a.http.apiBase+"/app", appJWT, nil)
	if err != nil {
		return appMetadata{}, err
	}
	var metadata appMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return appMetadata{}, fmt.Errorf("decode app metadata: %w", err)
	}
	return metadata, nil
}

// installation is one install of the app on a user or organization account.
type installation struct {
	ID      int64 `json:"id"`
	Account struct {
		Login string `json:"login"`
	} `json:"account"`
}

// installationID discovers which installation to act as. It is resolved once
// and cached, because a webhook would normally supply this value directly.
func (a *App) installationID(ctx context.Context) (int64, error) {
	a.mu.Lock()
	if a.cachedInstallationID != 0 {
		id := a.cachedInstallationID
		a.mu.Unlock()
		return id, nil
	}
	a.mu.Unlock()

	appJWT, err := signAppJWT(a.cfg.ClientID, a.privateKey, time.Now())
	if err != nil {
		return 0, err
	}
	data, err := a.http.json(ctx, "GET", a.http.apiBase+"/app/installations", appJWT, nil)
	if err != nil {
		return 0, err
	}

	var installations []installation
	if err := json.Unmarshal(data, &installations); err != nil {
		return 0, fmt.Errorf("decode installations: %w", err)
	}
	if len(installations) == 0 {
		return 0, fmt.Errorf("the app has no installations")
	}

	// When GITHUB_ORG is empty, the workshop uses the first installation.
	// Production apps should persist installation IDs from webhook payloads and
	// select the installation associated with the account making the request.
	chosen := installations[0]
	if a.cfg.Org != "" {
		found := false
		for _, candidate := range installations {
			if strings.EqualFold(candidate.Account.Login, a.cfg.Org) {
				chosen = candidate
				found = true
				break
			}
		}
		if !found {
			return 0, fmt.Errorf("app is not installed on organization %q", a.cfg.Org)
		}
	}

	a.mu.Lock()
	a.cachedInstallationID = chosen.ID
	a.mu.Unlock()
	a.log.Debug("discovered installations",
		"count", len(installations),
		"installation_id", chosen.ID,
		"account", chosen.Account.Login)
	return chosen.ID, nil
}

// installationToken returns a cached installation token, minting a new one when
// the cached token is within five minutes of expiry. The token is never written
// to disk; it is short-lived by design.
func (a *App) installationToken(ctx context.Context) (string, error) {
	id, err := a.installationID(ctx)
	if err != nil {
		return "", err
	}

	a.mu.Lock()
	if cached, ok := a.installationTokens[id]; ok && time.Until(cached.ExpiresAt) > 5*time.Minute {
		a.mu.Unlock()
		a.log.Debug("using cached installation token",
			"installation_id", id,
			"expires_at", cached.ExpiresAt)
		return cached.Token, nil
	}
	a.mu.Unlock()

	appJWT, err := signAppJWT(a.cfg.ClientID, a.privateKey, time.Now())
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/app/installations/%d/access_tokens", a.http.apiBase, id)
	data, err := a.http.json(ctx, "POST", url, appJWT, nil)
	if err != nil {
		return "", err
	}

	var minted installationToken
	if err := json.Unmarshal(data, &minted); err != nil {
		return "", fmt.Errorf("decode installation token: %w", err)
	}

	a.mu.Lock()
	a.installationTokens[id] = minted
	a.mu.Unlock()
	a.log.Debug("minted installation token",
		"installation_id", id,
		"expires_at", minted.ExpiresAt)
	return minted.Token, nil
}

// repository is the subset of a repository the workshop displays.
type repository struct {
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	HTMLURL  string `json:"html_url"`
}

func (a *App) listInstallationRepositories(ctx context.Context) ([]repository, error) {
	token, err := a.installationToken(ctx)
	if err != nil {
		return nil, err
	}
	data, err := a.http.json(ctx, "GET", a.http.apiBase+"/installation/repositories", token, nil)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Repositories []repository `json:"repositories"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("decode repositories: %w", err)
	}
	return payload.Repositories, nil
}
