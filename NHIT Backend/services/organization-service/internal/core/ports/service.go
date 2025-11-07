package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/domain"
)

// OrganizationService defines the interface for organization business logic
type OrganizationService interface {
	// CreateOrganization creates a new organization with business validation
	CreateOrganization(ctx context.Context, tenantID uuid.UUID, name, code, description, logo string, createdBy uuid.UUID) (*domain.Organization, error)
	
	// GetOrganization retrieves an organization by ID
	GetOrganization(ctx context.Context, orgID uuid.UUID) (*domain.Organization, error)
	
	// GetOrganizationByCode retrieves an organization by code
	GetOrganizationByCode(ctx context.Context, code string) (*domain.Organization, error)
	
	// UpdateOrganization updates an existing organization
	UpdateOrganization(ctx context.Context, orgID uuid.UUID, name, code, description, logo string, isActive bool) (*domain.Organization, error)
	
	// DeleteOrganization deletes an organization (with business rules)
	DeleteOrganization(ctx context.Context, orgID uuid.UUID, requestedBy uuid.UUID) error
	
	// ListOrganizationsByTenant retrieves all organizations for a tenant
	ListOrganizationsByTenant(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*domain.Organization, *PaginationResult, error)
	
	// ListAccessibleOrganizations retrieves all organizations accessible by a user
	ListAccessibleOrganizations(ctx context.Context, userID uuid.UUID, pagination PaginationParams) ([]*domain.Organization, *PaginationResult, error)
	
	// ToggleOrganizationStatus toggles the active status of an organization
	ToggleOrganizationStatus(ctx context.Context, orgID, requestedBy uuid.UUID) (*domain.Organization, error)
	
	// CheckOrganizationCode checks if an organization code is available
	CheckOrganizationCode(ctx context.Context, code string, excludeOrgID *uuid.UUID) (bool, error)
	
	// ValidateOrganizationAccess validates if a user can access an organization
	ValidateOrganizationAccess(ctx context.Context, userID, orgID uuid.UUID) error
}

// UserOrganizationService defines the interface for user-organization relationship business logic
type UserOrganizationService interface {
	// AddUserToOrganization adds a user to an organization with a role
	AddUserToOrganization(ctx context.Context, userID, orgID, roleID uuid.UUID) error
	
	// RemoveUserFromOrganization removes a user from an organization
	RemoveUserFromOrganization(ctx context.Context, userID, orgID uuid.UUID) error
	
	// SwitchUserOrganization switches a user's current organization context
	SwitchUserOrganization(ctx context.Context, userID, orgID uuid.UUID) error
	
	// GetUserCurrentOrganization retrieves a user's current organization
	GetUserCurrentOrganization(ctx context.Context, userID uuid.UUID) (*domain.Organization, error)
	
	// GetUserOrganizations retrieves all organizations for a user
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*domain.Organization, error)
	
	// UpdateUserRoleInOrganization updates a user's role within an organization
	UpdateUserRoleInOrganization(ctx context.Context, userID, orgID, roleID uuid.UUID) error
	
	// ListUsersInOrganization retrieves all users in an organization
	ListUsersInOrganization(ctx context.Context, orgID uuid.UUID) ([]uuid.UUID, error)
	
	// GetUserOrganizationRole retrieves a user's role in an organization
	GetUserOrganizationRole(ctx context.Context, userID, orgID uuid.UUID) (uuid.UUID, error)
}
