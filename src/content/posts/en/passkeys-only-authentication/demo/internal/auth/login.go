package auth

import (
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

func (a *App) beginLogin(w http.ResponseWriter, r *http.Request) {
	if !isJSON(r) {
		writeError(w, http.StatusUnsupportedMediaType, "use application/json")
		return
	}

	email, err := readEmail(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	user := a.store.userByEmail(email)
	if user == nil || len(user.Credentials) == 0 {
		writeError(w, http.StatusUnauthorized, "sign-in could not be started")
		return
	}

	options, data, err := a.webAuthn.BeginLogin(
		user,
		webauthn.WithUserVerification(protocol.VerificationRequired),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not start sign-in")
		return
	}

	a.saveCeremony(w, email, "login", *data)
	writeJSON(w, http.StatusOK, options)
}

func (a *App) finishLogin(w http.ResponseWriter, r *http.Request) {
	if !isJSON(r) {
		writeError(w, http.StatusUnsupportedMediaType, "use application/json")
		return
	}

	ceremony, err := a.takeCeremony(r, "login")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	user := a.store.userByEmail(ceremony.Email)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "sign-in failed")
		return
	}

	credential, err := a.webAuthn.FinishLogin(user, ceremony.Data, r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "passkey assertion could not be verified")
		return
	}

	if !a.store.replaceCredential(user.Email, *credential) {
		writeError(w, http.StatusInternalServerError, "credential state could not be saved")
		return
	}

	token, _ := a.createSession(user.Email)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set true behind HTTPS in production.
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"message":    "signed in",
		"credential": summarizeCredential(credential),
	})
}
