package auth

import (
	"bytes"
	"errors"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// User is one account. It implements the webauthn.User interface.
type User struct {
	ID          []byte
	Email       string
	ProfileName string
	Credentials []webauthn.Credential
}

func (u *User) WebAuthnID() []byte                         { return u.ID }
func (u *User) WebAuthnName() string                       { return u.Email }
func (u *User) WebAuthnDisplayName() string                { return u.Email }
func (u *User) WebAuthnCredentials() []webauthn.Credential { return u.Credentials }

// Ceremony is one in-flight registration or login exchange.
type Ceremony struct {
	Email     string
	Kind      string
	Data      webauthn.SessionData
	ExpiresAt time.Time
}

// Session is one authenticated browser session.
type Session struct {
	Email     string
	CSRFToken string
	ExpiresAt time.Time
}

// Store is an in-memory stand-in for a database.
type Store struct {
	mu         sync.RWMutex
	users      map[string]*User
	ceremonies map[string]Ceremony
	sessions   map[string]Session
}

func newStore() *Store {
	return &Store{
		users:      make(map[string]*User),
		ceremonies: make(map[string]Ceremony),
		sessions:   make(map[string]Session),
	}
}

func cloneUser(user *User) *User {
	if user == nil {
		return nil
	}
	copy := *user
	copy.ID = bytes.Clone(user.ID)
	copy.Credentials = append([]webauthn.Credential(nil), user.Credentials...)
	return &copy
}

func (s *Store) userByEmail(email string) *User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneUser(s.users[email])
}

func (s *Store) userForRegistration(email string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := s.users[email]
	if user == nil {
		user = &User{ID: randomBytes(32), Email: email, ProfileName: email}
		s.users[email] = user
	}
	if len(user.Credentials) > 0 {
		return nil, errors.New("this workshop account already has a passkey")
	}
	return cloneUser(user), nil
}

func (s *Store) addFirstCredential(email string, credential webauthn.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := s.users[email]
	if user == nil {
		return errors.New("registration expired")
	}
	if len(user.Credentials) > 0 {
		return errors.New("this workshop account already has a passkey")
	}
	for _, otherUser := range s.users {
		for _, existing := range otherUser.Credentials {
			if bytes.Equal(existing.ID, credential.ID) {
				return errors.New("credential is already registered")
			}
		}
	}
	user.Credentials = append(user.Credentials, credential)
	return nil
}

func (s *Store) replaceCredential(email string, credential webauthn.Credential) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := s.users[email]
	if user == nil {
		return false
	}
	for i := range user.Credentials {
		if bytes.Equal(user.Credentials[i].ID, credential.ID) {
			user.Credentials[i] = credential
			return true
		}
	}
	return false
}

func (s *Store) updateProfileName(email, profileName string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.users[email]
	if user == nil {
		return false
	}
	user.ProfileName = profileName
	return true
}
