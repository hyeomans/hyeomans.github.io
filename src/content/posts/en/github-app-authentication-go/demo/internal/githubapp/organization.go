package githubapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// organization is the account information displayed by the workshop UI.
type organization struct {
	Login       string `json:"login"`
	Name        string `json:"name"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	AvatarURL   string `json:"avatar_url"`
	PublicRepos int    `json:"public_repos"`
}

func (a *App) getOrganization(ctx context.Context) (organization, error) {
	token, err := a.installationToken(ctx)
	if err != nil {
		return organization{}, err
	}
	endpoint := fmt.Sprintf("%s/orgs/%s", a.http.apiBase, url.PathEscape(a.cfg.Org))
	data, err := a.http.json(ctx, "GET", endpoint, token, nil)
	if err != nil {
		return organization{}, err
	}
	var org organization
	if err := json.Unmarshal(data, &org); err != nil {
		return organization{}, fmt.Errorf("decode organization: %w", err)
	}
	return org, nil
}
