package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// UserRoleService provides business logic for managing user-role assignments.
type UserRoleService struct {
	repo ports.UserRoleRepository
}

// NewUserRoleService creates a new UserRoleService.
func NewUserRoleService(repo ports.UserRoleRepository) *UserRoleService {
	return &UserRoleService{repo: repo}
}

// AssignRoles assigns multiple roles to a user.
func (s *UserRoleService) AssignRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	for _, roleID := range roleIDs {
		params := db.AssignRoleToUserParams{
			UserID: userID,
			RoleID: roleID,
		}
		if err := s.repo.AssignRole(ctx, params); err != nil {
			return err
		}
	}
	return nil
}

// ListRolesOfUser fetches all roles and permissions assigned to a user.
func (s *UserRoleService) ListRolesOfUser(ctx context.Context, userID uuid.UUID) ([]db.ListRolesForUserRow, error) {
	return s.repo.ListRoles(ctx, userID)
}