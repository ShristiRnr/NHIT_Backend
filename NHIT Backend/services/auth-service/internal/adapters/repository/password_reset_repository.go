package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/domain"
)

type passwordResetRepository struct {
	db *sql.DB
}

func NewPasswordResetRepository(db *sql.DB) *passwordResetRepository {
	return &passwordResetRepository{db: db}
}

func (r *passwordResetRepository) Create(ctx context.Context, userID uuid.UUID, token uuid.UUID, expiresAt time.Time) (*domain.PasswordReset, error) {
	query := `
		INSERT INTO password_resets (token, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING token, user_id, expires_at, created_at
	`

	reset := &domain.PasswordReset{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		token,
		userID,
		expiresAt,
		time.Now(),
	).Scan(
		&reset.Token,
		&reset.UserID,
		&reset.ExpiresAt,
		&reset.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create password reset: %w", err)
	}

	return reset, nil
}

func (r *passwordResetRepository) GetByToken(ctx context.Context, token uuid.UUID) (*domain.PasswordReset, error) {
	query := `
		SELECT token, user_id, expires_at, created_at
		FROM password_resets
		WHERE token = $1 AND expires_at > NOW()
	`

	reset := &domain.PasswordReset{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&reset.Token,
		&reset.UserID,
		&reset.ExpiresAt,
		&reset.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("password reset token not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get password reset: %w", err)
	}

	return reset, nil
}

func (r *passwordResetRepository) Delete(ctx context.Context, token uuid.UUID) error {
	query := `DELETE FROM password_resets WHERE token = $1`

	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete password reset: %w", err)
	}

	return nil
}
