package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

var ErrSessionExpired = errors.New("session expired")

type Session struct {
	mu           sync.RWMutex
	ID           string
	CreatedAt    time.Time
	LastActivity time.Time
	Data         map[string]any
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func NewSession() (*Session, error) {
	id, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	return &Session{
		ID:           id,
		CreatedAt:    now,
		LastActivity: now,
		Data:         make(map[string]any),
	}, nil
}

func (s *Session) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.Data[key]
	return value, exists
}

func (s *Session) Put(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data[key] = value
}

func (s *Session) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.Data, key)
}

func (s *Session) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data = make(map[string]any)
}

func (s *Session) snapshot() (map[string]any, time.Time, time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := make(map[string]any, len(s.Data))

	for k, v := range s.Data {
		data[k] = v
	}

	return data, s.CreatedAt, s.LastActivity
}

func (s *Session) touch() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.LastActivity = time.Now().UTC()
}

func (s *Session) expired(idle, absolute time.Duration) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()

	return now.Sub(s.LastActivity) > idle || now.Sub(s.CreatedAt) > absolute
}

func (s *Session) regenerateID() error {
	id, err := generateSessionID()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.ID = id
	s.CreatedAt = time.Now().UTC()
	s.LastActivity = s.CreatedAt

	return nil
}

// Restore session data from Postgre
func restoreSession(
	id string,
	data []byte,
	createdAt time.Time,
	lastActivity time.Time,
) (*Session, error) {
	values := make(map[string]any)

	if err := json.Unmarshal(data, &values); err != nil {
		return nil, err
	}

	return &Session{
		ID:           id,
		CreatedAt:    createdAt,
		LastActivity: lastActivity,
		Data:         values,
	}, nil
}
