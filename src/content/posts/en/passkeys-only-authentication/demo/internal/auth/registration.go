package auth

import (
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

func (a *App) beginRegistration(w http.ResponseWriter, r *http.Request) {
	if !isJSON(r) {
		writeError(w, http.StatusUnsupportedMediaType, "use application/json")
		return
	}

	email, err := readEmail(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := a.store.userForRegistration(email)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	options, data, err := a.webAuthn.BeginRegistration(
		user,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		webauthn.WithExclusions(webauthn.Credentials(user.Credentials).CredentialDescriptors()),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not start registration")
		return
	}

	a.saveCeremony(w, email, "registration", *data)
	writeJSON(w, http.StatusOK, options)
}

func (a *App) finishRegistration(w http.ResponseWriter, r *http.Request) {
	if !isJSON(r) {
		writeError(w, http.StatusUnsupportedMediaType, "use application/json")
		return
	}

	ceremony, err := a.takeCeremony(r, "registration")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	user := a.store.userByEmail(ceremony.Email)
	if user == nil {
		writeError(w, http.StatusBadRequest, "registration expired")
		return
	}

	credential, err := a.webAuthn.FinishRegistration(user, ceremony.Data, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "passkey registration could not be verified")
		return
	}

	if err := a.store.addFirstCredential(ceremony.Email, *credential); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message":    "passkey registered",
		"credential": summarizeCredential(credential),
	})
}
