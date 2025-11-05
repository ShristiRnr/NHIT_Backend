package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// SessionRepo implements ports.SessionRepository using sqlc-generated queries.
type SessionRepo struct {
	q *db.Queries
}

// NewSessionRepo creates a new Session repository instance.
func NewSessionRepo(q *db.Queries) ports.SessionRepository {
	return &SessionRepo{q: q}
}

// Create inserts a new session for a user.
func (r *SessionRepo) Create(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (db.Session, error) {
	params := db.CreateSessionParams{
		UserID:    uuid.NullUUID{UUID: userID, Valid: true},
		SessionToken:     token,
		ExpiresAt: sql.NullTime{Time: expiresAt, Valid: true},
	}
	return r.q.CreateSession(ctx, params)
}

// Get retrieves a session by its ID.
func (r *SessionRepo) Get(ctx context.Context, sessionID uuid.UUID) (db.Session, error) {
	return r.q.GetSession(ctx, sessionID)
}

// Delete removes a session by its ID.
func (r *SessionRepo) Delete(ctx context.Context, sessionID uuid.UUID) error {
	return r.q.DeleteSession(ctx, sessionID)
}
