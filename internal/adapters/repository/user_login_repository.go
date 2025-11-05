package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// UserLoginRepo implements UserLoginRepository using sqlc-generated db.Queries
type UserLoginRepo struct {
	q *db.Queries
}

// NewUserLoginRepo creates a new UserLoginRepo instance.
func NewUserLoginRepo(q *db.Queries) ports.UserLoginRepository {
	return &UserLoginRepo{q: q}
}

// Create inserts a new login history record
func (r *UserLoginRepo) Create(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) (db.UserLoginHistory, error) {
	params := db.CreateLoginHistoryParams{
		UserID:    uuid.NullUUID{UUID: userID, Valid: true},
		IpAddress: sql.NullString{String: ipAddress, Valid: true},
		UserAgent: sql.NullString{String: userAgent, Valid: true},
	}
	return r.q.CreateLoginHistory(ctx, params)
}

// List returns paginated login history for a user
func (r *UserLoginRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]db.UserLoginHistory, error) {
	params := db.ListUserLoginHistoriesParams{
		UserID: uuid.NullUUID{UUID: userID, Valid: true},
		Limit:  limit,
		Offset: offset,
	}
	return r.q.ListUserLoginHistories(ctx, params)
}
