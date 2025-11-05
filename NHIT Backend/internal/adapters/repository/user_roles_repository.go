package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// UserRoleRepo is the SQLC adapter implementing UserRoleRepository.
type UserRoleRepo struct {
	q *db.Queries
}

// NewUserRoleRepo creates a new UserRoleRepo instance.
func NewUserRoleRepo(q *db.Queries) ports.UserRoleRepository {
	return &UserRoleRepo{q: q}
}

// AssignRole assigns a role to a user.
func (r *UserRoleRepo) AssignRole(ctx context.Context, params db.AssignRoleToUserParams) error {
	return r.q.AssignRoleToUser(ctx, params)
}

// ListRoles returns all roles assigned to a user.
func (r *UserRoleRepo) ListRoles(ctx context.Context, userID uuid.UUID) ([]db.ListRolesForUserRow, error) {
	return r.q.ListRolesForUser(ctx, userID)
}
