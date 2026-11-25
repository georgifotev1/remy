package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Session struct {
	Plaintext string // sent to the client in the cookie, never stored
	Hash      []byte // sha256 of Plaintext, stored in the db
	UserID    uuid.UUID
	Expiry    time.Time
}

// NewSession generates a random session token that expires after ttl.
// UserID is left empty; set it before inserting.
func NewSession(ttl time.Duration) (*Session, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	plaintext := base64.RawURLEncoding.EncodeToString(b)

	s := &Session{
		Plaintext: plaintext,
		Hash:      hashToken(plaintext),
		Expiry:    time.Now().Add(ttl),
	}
	return s, nil
}

func hashToken(plaintext string) []byte {
	hash := sha256.Sum256([]byte(plaintext))
	return hash[:]
}

func insertSession(ctx context.Context, q DBTX, s *Session) error {
	const query = `
		INSERT INTO sessions (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)`

	_, err := q.Exec(ctx, query, s.Hash, s.UserID, s.Expiry)
	return err
}

type SessionModel struct {
	db *pgxpool.Pool
}

func (m *SessionModel) Insert(ctx context.Context, s *Session) error {
	return insertSession(ctx, m.db, s)
}

func (m *SessionModel) DeleteByToken(ctx context.Context, plaintext string) error {
	const query = `
		DELETE FROM sessions
		WHERE token_hash = $1`

	_, err := m.db.Exec(ctx, query, hashToken(plaintext))
	return err
}

func (m *SessionModel) GetByToken(ctx, token string) {}
