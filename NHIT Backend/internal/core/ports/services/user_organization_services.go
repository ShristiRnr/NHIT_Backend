package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

type UserOrganizationService struct {
	repo ports.UserOrganizationRepository
}

func NewUserOrganizationService(repo ports.UserOrganizationRepository) *UserOrganizationService {
	return &UserOrganizationService{repo: repo}
}

// AddUserToOrganization assigns a user to an organization with a role
func (s *UserOrganizationService) AddUserToOrganization(ctx context.Context, userID, orgID, roleID uuid.UUID) error {
	params := db.AddUserToOrganizationParams{
		UserID: userID,
		OrgID:  orgID,
		RoleID: roleID,
	}
	return s.repo.AddUserToOrganization(ctx, params)
}

// ListUsersByOrganization retrieves all users assigned to a given organization
func (s *UserOrganizationService) ListUsersByOrganization(ctx context.Context, orgID uuid.UUID) ([]db.ListUsersByOrganizationRow, error) {
	return s.repo.ListUsersByOrganization(ctx, orgID)
}
