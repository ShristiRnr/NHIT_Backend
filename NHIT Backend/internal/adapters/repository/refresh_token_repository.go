package repository

import (
	"context"
	"errors"
	"time"
	"fmt"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// RefreshRepo implements ports.RefreshTokenRepository using sqlc-generated queries.
type RefreshRepo struct {
	q *db.Queries
}

// NewRefreshRepo creates a new RefreshRepo instance.
func NewRefreshRepo(q *db.Queries) ports.RefreshTokenRepository {
	return &RefreshRepo{q: q}
}

// Create inserts a new refresh token
func (r *RefreshRepo) Create(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	params := db.CreateRefreshTokenParams{
		Token:     token,
		UserID:    uuid.NullUUID{UUID: userID, Valid: true},
		ExpiresAt: expiresAt,
	}
	return r.q.CreateRefreshToken(ctx, params)
}

// GetUserIDByToken retrieves the user ID associated with a refresh token.
func (r *RefreshRepo) GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error) {
	nullID, err := r.q.GetUserIDByRefreshToken(ctx, token)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get user_id for token %s: %w", token, err)
	}
	if !nullID.Valid {
		return uuid.Nil, errors.New("refresh token not linked to a user")
	}
	return nullID.UUID, nil
}

// Delete removes a refresh token by its token string.
func (r *RefreshRepo) Delete(ctx context.Context, token string) error {
	return r.q.DeleteRefreshToken(ctx, token)
}
