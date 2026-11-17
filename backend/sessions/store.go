package sessions

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type SessionStore interface {
	Read(ctx context.Context, id string) (*Session, error)
	Write(ctx context.Context, session *Session) error
	Destroy(ctx context.Context, id string) error
	GC(ctx context.Context, idle, absolute time.Duration) error
}

type PostgreSQLStore struct {
	db *sql.DB
}

func NewPostgreSQLStore(db *sql.DB) *PostgreSQLStore {
	return &PostgreSQLStore{db: db}
}

func (p *PostgreSQLStore) Read(
	ctx context.Context,
	id string,
) (*Session, error) {
	var (
		data         []byte
		createdAt    time.Time
		lastActivity time.Time
	)

	err := p.db.QueryRowContext(ctx, `
	SELECT data, created_at, last_activity
	FROM sessions
	WHERE id = $1
	`, id).Scan(&data, &createdAt, &lastActivity)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return restoreSession(id, data, createdAt, lastActivity)
}

func (p *PostgreSQLStore) Write(
	ctx context.Context,
	session *Session,
) error {
	data, createdAt, lastActivity := session.snapshot()

	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = p.db.ExecContext(ctx, `
	INSERT INTO sessions (
	id, data, created_at, last_activity
	)
	VALUES ($1, $2::jsonb, $3, $4)
	ON CONFLICT (id)
	DO UPDATE SET
	    data = EXCLUDED.data,
		last_activity = EXCLUDED.last_activity
	`, session.ID, encoded, createdAt, lastActivity)

	return err
}

func (p *PostgreSQLStore) Destroy(
	ctx context.Context,
	id string,
) error {
	_, err := p.db.ExecContext(ctx, `
	DELETE FROM sessions WHERE id = $1
	`, id)

	return err
}

func (p *PostgreSQLStore) GC(
	ctx context.Context,
	idle, absolute time.Duration,
) error {
	_, err := p.db.ExecContext(ctx, `
	DELETE FROM sessions
	WHERE last_activity < NOW() - ($1 * INTERVAL '1 second')
	    OR created_at < NOW() - ($2 * INTERVAL '1 second')
		`,
		idle.Seconds(),
		absolute.Seconds(),
	)

	return err
}
