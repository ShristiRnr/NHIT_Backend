package repository

import (
	"context"

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
func (r *PasswordResetRepo) Create(ctx context.Context, arg db.CreatePasswordResetTokenParams) (db.PasswordReset, error) {
	token, err := r.q.CreatePasswordResetToken(ctx, arg)
	if err != nil {
		return db.PasswordReset{}, err
	}
	// Return the created token details
	return db.PasswordReset{
		Token:     token,
		UserID:    arg.UserID,
		ExpiresAt: arg.ExpiresAt,
	}, nil
}

// Get retrieves a password reset token by its token string.
func (r *PasswordResetRepo) Get(ctx context.Context, token uuid.UUID) (db.PasswordReset, error) {
	return r.q.GetPasswordResetToken(ctx, token)
}

// Delete removes a password reset token by its token string.
func (r *PasswordResetRepo) Delete(ctx context.Context, token uuid.UUID) error {
	return r.q.DeletePasswordResetToken(ctx, token)
}
