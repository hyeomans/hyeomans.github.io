package githubapp

import "net/http"

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/app", a.handleAppMetadata)
	mux.HandleFunc("GET /api/installation/repositories", a.handleRepositories)
	mux.HandleFunc("GET /api/organization", a.handleOrganization)
	mux.Handle("/", http.FileServer(http.Dir(a.cfg.StaticDir)))
	return securityHeaders(mux)
}

func (a *App) handleAppMetadata(w http.ResponseWriter, r *http.Request) {
	metadata, err := a.getAppMetadata(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, metadata)
}

func (a *App) handleRepositories(w http.ResponseWriter, r *http.Request) {
	repositories, err := a.listInstallationRepositories(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"repositories": repositories})
}

func (a *App) handleOrganization(w http.ResponseWriter, r *http.Request) {
	if a.cfg.Org == "" {
		writeError(w, http.StatusBadRequest, "set GITHUB_ORG to load organization metadata")
		return
	}
	org, err := a.getOrganization(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, org)
}
