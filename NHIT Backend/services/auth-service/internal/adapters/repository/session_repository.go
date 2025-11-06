package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/domain"
)

type sessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *sessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) (*domain.Session, error) {
	query := `
		INSERT INTO sessions (session_id, user_id, session_token, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING session_id, user_id, session_token, created_at, expires_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		session.SessionID,
		session.UserID,
		session.SessionToken,
		session.CreatedAt,
		session.ExpiresAt,
	).Scan(
		&session.SessionID,
		&session.UserID,
		&session.SessionToken,
		&session.CreatedAt,
		&session.ExpiresAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func (r *sessionRepository) GetByID(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error) {
	query := `
		SELECT session_id, user_id, session_token, created_at, expires_at
		FROM sessions
		WHERE session_id = $1 AND expires_at > NOW()
	`

	session := &domain.Session{}
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.SessionID,
		&session.UserID,
		&session.SessionToken,
		&session.CreatedAt,
		&session.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

func (r *sessionRepository) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	query := `
		SELECT session_id, user_id, session_token, created_at, expires_at
		FROM sessions
		WHERE session_token = $1 AND expires_at > NOW()
	`

	session := &domain.Session{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&session.SessionID,
		&session.UserID,
		&session.SessionToken,
		&session.CreatedAt,
		&session.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

func (r *sessionRepository) Delete(ctx context.Context, sessionID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE session_id = $1`

	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}

	return nil
}
