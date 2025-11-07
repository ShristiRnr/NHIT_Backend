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
	// ErrOrganizationNotFound is returned when organization is not found
	ErrOrganizationNotFound = errors.New("organization not found")
	
	// ErrUnauthorizedAccess is returned when user is not authorized to perform an action
	ErrUnauthorizedAccess = errors.New("unauthorized access to organization")
	
	// ErrOrganizationHasUsers is returned when trying to delete an organization with users
	ErrOrganizationHasUsers = errors.New("cannot delete organization with existing users")
)

type organizationService struct {
	orgRepo     ports.OrganizationRepository
	userOrgRepo ports.UserOrganizationRepository
}

// NewOrganizationService creates a new organization service instance
func NewOrganizationService(
	orgRepo ports.OrganizationRepository,
	userOrgRepo ports.UserOrganizationRepository,
) ports.OrganizationService {
	return &organizationService{
		orgRepo:     orgRepo,
		userOrgRepo: userOrgRepo,
	}
}

// CreateOrganization creates a new organization with business validation
func (s *organizationService) CreateOrganization(
	ctx context.Context,
	tenantID uuid.UUID,
	name, code, description, logo string,
	createdBy uuid.UUID,
) (*domain.Organization, error) {
	log.Printf("Creating organization: name=%s, code=%s, tenant=%s, creator=%s", 
		name, code, tenantID.String(), createdBy.String())
	
	// Check if code already exists
	exists, err := s.orgRepo.CodeExists(ctx, code, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check code existence: %w", err)
	}
	
	if exists {
		return nil, domain.ErrDuplicateOrganizationCode
	}
	
	// Create organization domain object with validation
	org, err := domain.NewOrganization(tenantID, name, code, description, logo, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}
	
	// Persist organization
	createdOrg, err := s.orgRepo.Create(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to persist organization: %w", err)
	}
	
	// Automatically add the creator to the organization (this would be done via UserOrganizationService in production)
	log.Printf("Organization created successfully: id=%s, database=%s", 
		createdOrg.OrgID.String(), createdOrg.DatabaseName)
	
	return createdOrg, nil
}

// GetOrganization retrieves an organization by ID
func (s *organizationService) GetOrganization(ctx context.Context, orgID uuid.UUID) (*domain.Organization, error) {
	if orgID == uuid.Nil {
		return nil, errors.New("organization ID cannot be empty")
	}
	
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve organization: %w", err)
	}
	
	if org == nil {
		return nil, ErrOrganizationNotFound
	}
	
	return org, nil
}

// GetOrganizationByCode retrieves an organization by code
func (s *organizationService) GetOrganizationByCode(ctx context.Context, code string) (*domain.Organization, error) {
	if code == "" {
		return nil, errors.New("organization code cannot be empty")
	}
	
	org, err := s.orgRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve organization by code: %w", err)
	}
	
	if org == nil {
		return nil, ErrOrganizationNotFound
	}
	
	return org, nil
}

// UpdateOrganization updates an existing organization
func (s *organizationService) UpdateOrganization(
	ctx context.Context,
	orgID uuid.UUID,
	name, code, description, logo string,
	isActive bool,
) (*domain.Organization, error) {
	log.Printf("Updating organization: id=%s, name=%s, code=%s", orgID.String(), name, code)
	
	// Retrieve existing organization
	org, err := s.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}
	
	// Check if code is being changed and if new code already exists
	if org.Code != code {
		exists, err := s.orgRepo.CodeExists(ctx, code, &orgID)
		if err != nil {
			return nil, fmt.Errorf("failed to check code existence: %w", err)
		}
		
		if exists {
			return nil, domain.ErrDuplicateOrganizationCode
		}
	}
	
	// Update organization fields
	if err := org.Update(name, code, description, logo, isActive); err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}
	
	// Persist updated organization
	updatedOrg, err := s.orgRepo.Update(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to persist organization update: %w", err)
	}
	
	log.Printf("Organization updated successfully: id=%s", updatedOrg.OrgID.String())
	
	return updatedOrg, nil
}

// DeleteOrganization deletes an organization with business rules
func (s *organizationService) DeleteOrganization(ctx context.Context, orgID, requestedBy uuid.UUID) error {
	log.Printf("Deleting organization: id=%s, requested_by=%s", orgID.String(), requestedBy.String())
	
	// Retrieve organization
	org, err := s.GetOrganization(ctx, orgID)
	if err != nil {
		return err
	}
	
	// Business rule: Check if organization has users
	users, err := s.userOrgRepo.ListUsersByOrganization(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to check organization users: %w", err)
	}
	
	if len(users) > 0 {
		return ErrOrganizationHasUsers
	}
	
	// Note: In production, you might want to add more business rules:
	// - Only superadmin or creator can delete
	// - Check if organization has data
	// - Soft delete instead of hard delete
	// - Delete associated database
	
	if err := s.orgRepo.Delete(ctx, orgID); err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}
	
	log.Printf("Organization deleted successfully: id=%s", org.OrgID.String())
	
	return nil
}

// ListOrganizationsByTenant retrieves all organizations for a tenant
func (s *organizationService) ListOrganizationsByTenant(
	ctx context.Context,
	tenantID uuid.UUID,
	pagination ports.PaginationParams,
) ([]*domain.Organization, *ports.PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, nil, errors.New("tenant ID cannot be empty")
	}
	
	// Set default pagination if not provided
	if pagination.PageSize == 0 {
		pagination.PageSize = 10
	}
	if pagination.Page == 0 {
		pagination.Page = 1
	}
	
	orgs, paginationResult, err := s.orgRepo.ListByTenant(ctx, tenantID, pagination)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list organizations by tenant: %w", err)
	}
	
	return orgs, paginationResult, nil
}

// ListAccessibleOrganizations retrieves all organizations accessible by a user
func (s *organizationService) ListAccessibleOrganizations(
	ctx context.Context,
	userID uuid.UUID,
	pagination ports.PaginationParams,
) ([]*domain.Organization, *ports.PaginationResult, error) {
	if userID == uuid.Nil {
		return nil, nil, errors.New("user ID cannot be empty")
	}
	
	// Set default pagination if not provided
	if pagination.PageSize == 0 {
		pagination.PageSize = 10
	}
	if pagination.Page == 0 {
		pagination.Page = 1
	}
	
	orgs, paginationResult, err := s.orgRepo.ListAccessibleByUser(ctx, userID, pagination)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list accessible organizations: %w", err)
	}
	
	return orgs, paginationResult, nil
}

// ToggleOrganizationStatus toggles the active status of an organization
func (s *organizationService) ToggleOrganizationStatus(
	ctx context.Context,
	orgID, requestedBy uuid.UUID,
) (*domain.Organization, error) {
	log.Printf("Toggling organization status: id=%s, requested_by=%s", 
		orgID.String(), requestedBy.String())
	
	// Retrieve organization
	org, err := s.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}
	
	// Business rule: In production, you might want to check permissions here
	// For example, only superadmin or creator can toggle status
	
	// Toggle status
	org.ToggleStatus()
	
	// Persist updated organization
	updatedOrg, err := s.orgRepo.Update(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to toggle organization status: %w", err)
	}
	
	status := "deactivated"
	if updatedOrg.IsActive {
		status = "activated"
	}
	log.Printf("Organization %s successfully: id=%s", status, updatedOrg.OrgID.String())
	
	return updatedOrg, nil
}

// CheckOrganizationCode checks if an organization code is available
func (s *organizationService) CheckOrganizationCode(
	ctx context.Context,
	code string,
	excludeOrgID *uuid.UUID,
) (bool, error) {
	if code == "" {
		return false, errors.New("organization code cannot be empty")
	}
	
	exists, err := s.orgRepo.CodeExists(ctx, code, excludeOrgID)
	if err != nil {
		return false, fmt.Errorf("failed to check code availability: %w", err)
	}
	
	return !exists, nil
}

// ValidateOrganizationAccess validates if a user can access an organization
func (s *organizationService) ValidateOrganizationAccess(
	ctx context.Context,
	userID, orgID uuid.UUID,
) error {
	if userID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}
	
	if orgID == uuid.Nil {
		return errors.New("organization ID cannot be empty")
	}
	
	// Check if organization exists and is active
	org, err := s.GetOrganization(ctx, orgID)
	if err != nil {
		return err
	}
	
	if !org.CanBeAccessed() {
		return domain.ErrOrganizationNotActive
	}
	
	// Check if user has access to organization
	hasAccess, err := s.userOrgRepo.UserHasAccessToOrganization(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to validate user access: %w", err)
	}
	
	if !hasAccess {
		return ErrUnauthorizedAccess
	}
	
	return nil
}
