package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
)

type userRoleRepository struct {
	queries *sqlc.Queries
}

// NewUserRoleRepository creates a new user role repository
func NewUserRoleRepository(queries *sqlc.Queries) ports.UserRoleRepository {
	return &userRoleRepository{queries: queries}
}

func (r *userRoleRepository) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	params := sqlc.AssignRoleToUserParams{
		UserID: userID,
		RoleID: roleID,
	}
	return r.queries.AssignRoleToUser(ctx, params)
}

func (r *userRoleRepository) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error {
	params := sqlc.RemoveRoleFromUserParams{
		UserID: userID,
		RoleID: roleID,
	}
	return r.queries.RemoveRoleFromUser(ctx, params)
}

func (r *userRoleRepository) ListRolesByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	roleIDs, err := r.queries.ListRolesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles := make([]*domain.Role, len(roleIDs))
	for i, roleID := range roleIDs {
		roles[i] = &domain.Role{
			RoleID: roleID,
			// Name and other fields would need to be fetched from role service
		}
	}

	return roles, nil
}
