package sessions

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const sessionContextKey = "my_media_app"

type SessionManager struct {
	store              SessionStore
	idleExpiration     time.Duration
	absoluteExpiration time.Duration
	gcInterval         time.Duration
	cookieName         string
	secureCookie       bool

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type SessionConfig struct {
	IdleExpiration     time.Duration
	AbsoluteExpiration time.Duration
	GCInterval         time.Duration
	CookieName         string
	SecureCookie       bool
}

func NewSessionManager(
	store SessionStore,
	config SessionConfig,
) (*SessionManager, error) {
	if store == nil {
		return nil, errors.New("session store cannot be nil")
	}

	if config.IdleExpiration <= 0 || config.AbsoluteExpiration <= 0 || config.GCInterval <= 0 {
		return nil, errors.New("session expiration and GC intervals must be positive")
	}

	if config.CookieName == "" {
		config.CookieName = "session_id"
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &SessionManager{
		store:       store,
		idleExpiration: config.IdleExpiration,
		absoluteExpiration: config.AbsoluteExpiration,
		gcInterval:  config.GCInterval,
		cookieName:  config.CookieName,
		secureCookie:  config.SecureCookie,
		cancel:   cancel,
	}

	m.wg.Add(1)
	go m.gc(ctx)

	return m, nil
}

func (m *SessionManager) gc(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(m.gcInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := m.store.GC(
				ctx,
				m.idleExpiration,
				m.absoluteExpiration,
			)

			if err != nil && ctx.Err() == nil {
				log.Printf("session garbage collection failed: %v", err)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (m *SessionManager) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var session *Session

		cookie, err := c.Request.Cookie(m.cookieName)

		if err == nil {
			session, err = m.store.Read(ctx, cookie.Value)

			if err != nil {
				log.Printf("session read failed: %v", err)
				c.AbortWithStatus(http.StatusServiceUnavailable)
				return
			}
		}

		// Reject expired sessions
		if session != nil && session.expired(
			m.idleExpiration,
			m.absoluteExpiration,
		) {
			if err := m.store.Destroy(ctx, session.ID); err != nil {
				log.Printf("expired session deletion failed: %v", err)
				c.AbortWithStatus(http.StatusServiceUnavailable)
				return
			}

			session = nil
		}

		// Create and persist a new session
		if session == nil {
			session, err = NewSession()
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			if err := m.store.Write(ctx, session); err != nil {
				log.Printf("new session creation failed: %v", err)
				c.AbortWithStatus(http.StatusServiceUnavailable)
				return
			}

			m.setCookie(c, session.ID)
		}

		session.touch()

		c.Set(sessionContextKey, session)

		c.Next()

		// Persist session changes after request processing
		if err := m.store.Write(ctx, session); err != nil {
			log.Printf("session save failed: %v", err)
		}
	}
}

func (m *SessionManager) setCookie(c *gin.Context, id string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     m.cookieName,
		Value:    id,
		Path:     "/",
		MaxAge:   int(m.absoluteExpiration.Seconds()),
		HttpOnly: true,
		Secure:   m.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})
}

func (m *SessionManager) Get(c *gin.Context) (*Session, bool) {
	value, exists := c.Get(sessionContextKey)
	if !exists {
		return nil, false
	}

	session, ok := value.(*Session)
	return session, ok
}

// Destroy the current session and clear its cookie
func (m *SessionManager) Destroy(c *gin.Context) error {
	session, ok := m.Get(c)
	if !ok {
		return nil
	}

	if err := m.store.Destroy(
		c.Request.Context(),
		session.ID,
	); err != nil {
		return err
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:         m.cookieName,
		Value:        "",
		Path:         "/",
		MaxAge:       -1,
		HttpOnly:     true,
		Secure:       m.secureCookie,
		SameSite:     http.SameSiteLaxMode,
	})

	session.Clear()
	c.Set(sessionContextKey, (*Session)(nil))

	return nil
}

// Stop the garbage collector
func (m *SessionManager) Close() {
	m.cancel()
	m.wg.Wait()
}