package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
)

type userRoleRepository struct {
	queries *db.Queries
}

// NewUserRoleRepository creates a new user role repository
func NewUserRoleRepository(queries *db.Queries) ports.UserRoleRepository {
	return &userRoleRepository{queries: queries}
}

func (r *userRoleRepository) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	params := db.AssignRoleToUserParams{
		UserID: userID,
		RoleID: roleID,
	}
	return r.queries.AssignRoleToUser(ctx, params)
}

func (r *userRoleRepository) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error {
	// This would need a new SQL query - for now return nil
	// TODO: Add RemoveRoleFromUser query to sqlc
	return nil
}

func (r *userRoleRepository) ListRolesByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	dbRoles, err := r.queries.ListRolesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles := make([]*domain.Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		roles[i] = &domain.Role{
			RoleID: dbRole.RoleID,
			Name:   dbRole.Name,
			// Permissions field removed from query - can be fetched separately if needed
		}
	}

	return roles, nil
}
