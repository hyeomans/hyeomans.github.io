package auth

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"time"
)

type sessionHandler func(http.ResponseWriter, *http.Request, Session)

func (a *App) requireSession(next sessionHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "sign in first")
			return
		}

		a.store.mu.RLock()
		session, ok := a.store.sessions[cookie.Value]
		a.store.mu.RUnlock()
		if !ok || time.Now().After(session.ExpiresAt) {
			writeError(w, http.StatusUnauthorized, "session is missing or expired")
			return
		}
		next(w, r, session)
	}
}

func (a *App) requireCSRF(next sessionHandler) sessionHandler {
	return func(w http.ResponseWriter, r *http.Request, session Session) {
		got := r.Header.Get("X-CSRF-Token")
		if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(session.CSRFToken)) != 1 {
			writeError(w, http.StatusForbidden, "missing or invalid CSRF token")
			return
		}
		next(w, r, session)
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; base-uri 'none'; frame-ancestors 'none'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}
