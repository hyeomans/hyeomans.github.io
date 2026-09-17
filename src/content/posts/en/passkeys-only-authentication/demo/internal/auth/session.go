package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func (a *App) getSession(w http.ResponseWriter, _ *http.Request, session Session) {
	user := a.store.userByEmail(session.Email)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "session user no longer exists")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"email":       user.Email,
		"profileName": user.ProfileName,
		"csrfToken":   session.CSRFToken,
	})
}

func (a *App) updateProfile(w http.ResponseWriter, r *http.Request, session Session) {
	if !isJSON(r) {
		writeError(w, http.StatusUnsupportedMediaType, "use application/json")
		return
	}

	var body struct {
		ProfileName string `json:"profileName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	body.ProfileName = strings.TrimSpace(body.ProfileName)
	if body.ProfileName == "" || len(body.ProfileName) > 80 {
		writeError(w, http.StatusBadRequest, "profile name must be 1 to 80 characters")
		return
	}

	if !a.store.updateProfileName(session.Email, body.ProfileName) {
		writeError(w, http.StatusUnauthorized, "session user no longer exists")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "profile updated"})
}

func (a *App) logout(w http.ResponseWriter, r *http.Request, _ Session) {
	cookie, _ := r.Cookie(sessionCookie)
	if cookie != nil {
		a.store.mu.Lock()
		delete(a.store.sessions, cookie.Value)
		a.store.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	writeJSON(w, http.StatusOK, map[string]string{"message": "signed out"})
}

func (a *App) createSession(email string) (string, Session) {
	token := randomToken(32)
	session := Session{Email: email, CSRFToken: randomToken(32), ExpiresAt: time.Now().Add(sessionTTL)}
	a.store.mu.Lock()
	a.store.sessions[token] = session
	a.store.mu.Unlock()
	return token, session
}
