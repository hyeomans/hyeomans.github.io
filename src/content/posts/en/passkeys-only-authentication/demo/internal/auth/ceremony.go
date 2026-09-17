package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

func (a *App) saveCeremony(w http.ResponseWriter, email, kind string, data webauthn.SessionData) {
	id := randomToken(32)
	a.store.mu.Lock()
	a.store.ceremonies[id] = Ceremony{Email: email, Kind: kind, Data: data, ExpiresAt: time.Now().Add(ceremonyTTL)}
	a.store.mu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name:     ceremonyCookie,
		Value:    id,
		Path:     "/api/passkeys/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(ceremonyTTL.Seconds()),
	})
}

func (a *App) takeCeremony(r *http.Request, kind string) (Ceremony, error) {
	cookie, err := r.Cookie(ceremonyCookie)
	if err != nil {
		return Ceremony{}, errors.New("ceremony cookie is missing")
	}
	a.store.mu.Lock()
	ceremony, ok := a.store.ceremonies[cookie.Value]
	delete(a.store.ceremonies, cookie.Value)
	a.store.mu.Unlock()
	if !ok || ceremony.Kind != kind || time.Now().After(ceremony.ExpiresAt) {
		return Ceremony{}, errors.New("ceremony is missing, expired, or already used")
	}
	return ceremony, nil
}
