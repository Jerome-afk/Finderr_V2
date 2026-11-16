// This file contains middleware functions for the Gin web framework.
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"time"	
)

// Defining the session structure
type Session struct {
	createdAt    time.Time
	lastActivity time.Time
	id           string
	data         map[string]any
}

// Session storage medium
type SessionStore interface {
	read(id string) (*Session, error)
	write(session *Session) error
	destroy(id string) error
	gc(idleExpiration, absoluteExpiration time.Duration) error
}

// SessionManager manages user sessions
type SessionManager struct {
	store              SessionStore
	idleExpiration     time.Duration
	absoluteExpiration time.Duration
	cookieName         string
}

// Main function include handling sessions among other things.
func generateSessionID() string {
	id := make([]byte, 32)

	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		panic("failed to generate session ID")
	}

	return base64.RawURLEncoding.EncodeToString(id)
}

// Creating a new session
func NewSession() *Session {
	return &Session{
		id:           generateSessionID(),
		data:         make(map[string]any),
		createdAt:    time.Now(),
		lastActivity: time.Now(),
	}
}

func (m *SessionManager) gc(d time.Duration) {
	ticker := time.NewTicker(d)

	for range ticker.C {
		m.store.gc(m.idleExpiration, m.absoluteExpiration)
	}
}

// Creating a new session manager
func NewSessionManager(
	store SessionStore,
	gcInterval,
	idleExpiration,
	absoluteExpiration time.Duration,
	cookieName string,
) *SessionManager {
	m := &SessionManager{
		store:              store,
		idleExpiration:     idleExpiration,
		absoluteExpiration: absoluteExpiration,
		cookieName:         cookieName,
	}

	go m.gc(gcInterval)

	return m
}

func (s *Session) Get(key string) any {
	return s.data[key]
}

func (s *Session) Put(key string, value any) {
	s.data[key] = value
}

func (s *Session) Delete(key string) {
	delete(s.data, key)
}

// Session validation
func (m *SessionManager) validate(session *Session) bool {
	if time.Since(session.createdAt) > m.absoluteExpiration || time.Since(session.lastActivity) > m.idleExpiration {

		// Delete the session from the store if it has expired
		err := m.store.destroy(session.id)
		if err != nil {
			// Log the error if session deletion fails
			panic("failed to delete expired session: " + err.Error())
		}

		return false
	}

	return true
}

func (m *SessionManager) Start(r *http.Request) (*Session, *http.Request) {
	var session *Session

	// Check if the session cookie exists
	cookie, err := r.Cookie(m.cookieName)
	if err == nil {
		session, err = m.store.read(cookie.Value)
		if err != nil {
			log.Printf("failed to read session: %v", err)
		}
	}

	// Generate a new session
	if session == nil || !m.validate(session) {
		session = NewSession()
	}

	// Attach the session to the request context
	ctx := context.WithValue(r.Context(), sessionContextKey{}, session)
	r = r.WithContext(ctx)

	// Update the last activity timestamp
	session.lastActivity = time.Now()

	return session, r
}

func (m *SessionManager) Save(session *Session) error {
	session.lastActivity = time.Now()

	err := m.store.write(session)
	if err != nil {
		return err
	}

	return nil
}

func (m *SessionManager) migrate(session *Session) error {
	session.mu.Lock()
	defer session.mu.Unlock()

	err := m.store.destroy(session.id)
	if err != nil {
		return err
	}

	session.id = generateSessionID()

	return nil
}
