package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// Ensure interface compliance at compile time
var _ ports.PasswordResetRepository = &PasswordResetRepo{}

// PasswordResetRepo implements ports.PasswordResetRepository using sqlc-generated queries.
type PasswordResetRepo struct {
	q *db.Queries
}

// NewPasswordResetRepo creates a new repository instance.
func NewPasswordResetRepo(q *db.Queries) ports.PasswordResetRepository {
	return &PasswordResetRepo{q: q}
}

// Create inserts a new password reset token.
func (r *PasswordResetRepo) Create(ctx context.Context, email string, token uuid.UUID, expiresAt time.Time) (db.PasswordReset, error) {
    return r.q.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
        Email:     email,
        Token:     token,
        ExpiresAt: expiresAt,
    })
}

// Get retrieves a password reset token by its token string.
func (r *PasswordResetRepo) GetByToken(ctx context.Context, token uuid.UUID) (db.PasswordReset, error) {
	return r.q.GetPasswordResetByToken(ctx, token)
}

// Delete removes a password reset token by its token string.
func (r *PasswordResetRepo) Delete(ctx context.Context, token uuid.UUID) error {
	return r.q.DeletePasswordResetToken(ctx, token)
}
