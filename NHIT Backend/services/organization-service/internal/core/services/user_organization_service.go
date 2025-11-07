package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/ports"
)

var (
	// ErrUserOrganizationNotFound is returned when user-organization relationship is not found
	ErrUserOrganizationNotFound = errors.New("user-organization relationship not found")
	
	// ErrUserAlreadyInOrganization is returned when user is already in an organization
	ErrUserAlreadyInOrganization = errors.New("user is already a member of this organization")
)

type userOrganizationService struct {
	orgRepo     ports.OrganizationRepository
	userOrgRepo ports.UserOrganizationRepository
}

// NewUserOrganizationService creates a new user-organization service instance
func NewUserOrganizationService(
	orgRepo ports.OrganizationRepository,
	userOrgRepo ports.UserOrganizationRepository,
) ports.UserOrganizationService {
	return &userOrganizationService{
		orgRepo:     orgRepo,
		userOrgRepo: userOrgRepo,
	}
}

// AddUserToOrganization adds a user to an organization with a role
func (s *userOrganizationService) AddUserToOrganization(
	ctx context.Context,
	userID, orgID, roleID uuid.UUID,
) error {
	log.Printf("Adding user to organization: user=%s, org=%s, role=%s", 
		userID.String(), orgID.String(), roleID.String())
	
	// Validate organization exists and is active
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}
	
	if org == nil {
		return ErrOrganizationNotFound
	}
	
	if !org.CanBeAccessed() {
		return domain.ErrOrganizationNotActive
	}
	
	// Check if user already exists in organization
	hasAccess, err := s.userOrgRepo.UserHasAccessToOrganization(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to check user access: %w", err)
	}
	
	if hasAccess {
		return ErrUserAlreadyInOrganization
	}
	
	// Create user-organization relationship
	userOrg, err := domain.NewUserOrganization(userID, orgID, roleID)
	if err != nil {
		return fmt.Errorf("failed to create user-organization: %w", err)
	}
	
	// Persist relationship
	if err := s.userOrgRepo.AddUserToOrganization(ctx, userOrg); err != nil {
		return fmt.Errorf("failed to add user to organization: %w", err)
	}
	
	log.Printf("User added to organization successfully: user=%s, org=%s", 
		userID.String(), orgID.String())
	
	return nil
}

// RemoveUserFromOrganization removes a user from an organization
func (s *userOrganizationService) RemoveUserFromOrganization(
	ctx context.Context,
	userID, orgID uuid.UUID,
) error {
	log.Printf("Removing user from organization: user=%s, org=%s", 
		userID.String(), orgID.String())
	
	// Check if relationship exists
	hasAccess, err := s.userOrgRepo.UserHasAccessToOrganization(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to check user access: %w", err)
	}
	
	if !hasAccess {
		return ErrUserOrganizationNotFound
	}
	
	// Business rule: Check if this is the user's current organization
	currentOrg, err := s.userOrgRepo.GetCurrentOrganization(ctx, userID)
	if err == nil && currentOrg != nil && currentOrg.OrgID == orgID {
		// In production, you might want to prevent removal or switch to another org first
		log.Printf("Warning: Removing user from their current organization context")
	}
	
	// Remove relationship
	if err := s.userOrgRepo.RemoveUserFromOrganization(ctx, userID, orgID); err != nil {
		return fmt.Errorf("failed to remove user from organization: %w", err)
	}
	
	log.Printf("User removed from organization successfully: user=%s, org=%s", 
		userID.String(), orgID.String())
	
	return nil
}

// SwitchUserOrganization switches a user's current organization context
func (s *userOrganizationService) SwitchUserOrganization(
	ctx context.Context,
	userID, orgID uuid.UUID,
) error {
	log.Printf("Switching user organization context: user=%s, to_org=%s", 
		userID.String(), orgID.String())
	
	// Validate user has access to the organization
	hasAccess, err := s.userOrgRepo.UserHasAccessToOrganization(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to check user access: %w", err)
	}
	
	if !hasAccess {
		return fmt.Errorf("user does not have access to organization")
	}
	
	// Validate organization is active
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}
	
	if !org.CanBeAccessed() {
		return domain.ErrOrganizationNotActive
	}
	
	// Set as current organization
	if err := s.userOrgRepo.SetCurrentOrganization(ctx, userID, orgID); err != nil {
		return fmt.Errorf("failed to switch organization: %w", err)
	}
	
	log.Printf("User organization context switched successfully: user=%s, org=%s, database=%s", 
		userID.String(), orgID.String(), org.DatabaseName)
	
	return nil
}

// GetUserCurrentOrganization retrieves a user's current organization
func (s *userOrganizationService) GetUserCurrentOrganization(
	ctx context.Context,
	userID uuid.UUID,
) (*domain.Organization, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user ID cannot be empty")
	}
	
	org, err := s.userOrgRepo.GetCurrentOrganization(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current organization: %w", err)
	}
	
	return org, nil
}

// GetUserOrganizations retrieves all organizations for a user
func (s *userOrganizationService) GetUserOrganizations(
	ctx context.Context,
	userID uuid.UUID,
) ([]*domain.Organization, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user ID cannot be empty")
	}
	
	orgs, err := s.userOrgRepo.ListOrganizationsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}
	
	// Filter out inactive organizations
	activeOrgs := make([]*domain.Organization, 0)
	for _, org := range orgs {
		if org.CanBeAccessed() {
			activeOrgs = append(activeOrgs, org)
		}
	}
	
	return activeOrgs, nil
}

// UpdateUserRoleInOrganization updates a user's role within an organization
func (s *userOrganizationService) UpdateUserRoleInOrganization(
	ctx context.Context,
	userID, orgID, roleID uuid.UUID,
) error {
	log.Printf("Updating user role in organization: user=%s, org=%s, new_role=%s", 
		userID.String(), orgID.String(), roleID.String())
	
	// Get existing relationship
	userOrg, err := s.userOrgRepo.GetUserOrganization(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to get user-organization: %w", err)
	}
	
	if userOrg == nil {
		return ErrUserOrganizationNotFound
	}
	
	// Update role
	if err := userOrg.UpdateRole(roleID); err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	
	// Persist update
	if err := s.userOrgRepo.UpdateUserOrganization(ctx, userOrg); err != nil {
		return fmt.Errorf("failed to persist role update: %w", err)
	}
	
	log.Printf("User role updated successfully: user=%s, org=%s, role=%s", 
		userID.String(), orgID.String(), roleID.String())
	
	return nil
}

// ListUsersInOrganization retrieves all users in an organization
func (s *userOrganizationService) ListUsersInOrganization(
	ctx context.Context,
	orgID uuid.UUID,
) ([]uuid.UUID, error) {
	if orgID == uuid.Nil {
		return nil, errors.New("organization ID cannot be empty")
	}
	
	userIDs, err := s.userOrgRepo.ListUsersByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list users in organization: %w", err)
	}
	
	return userIDs, nil
}

// GetUserOrganizationRole retrieves a user's role in an organization
func (s *userOrganizationService) GetUserOrganizationRole(
	ctx context.Context,
	userID, orgID uuid.UUID,
) (uuid.UUID, error) {
	if userID == uuid.Nil || orgID == uuid.Nil {
		return uuid.Nil, errors.New("user ID and organization ID cannot be empty")
	}
	
	userOrg, err := s.userOrgRepo.GetUserOrganization(ctx, userID, orgID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get user-organization: %w", err)
	}
	
	if userOrg == nil {
		return uuid.Nil, ErrUserOrganizationNotFound
	}
	
	return userOrg.RoleID, nil
}
