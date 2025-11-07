package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/domain"
)

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page     int32
	PageSize int32
}

// PaginationResult holds pagination result metadata
type PaginationResult struct {
	CurrentPage int32
	PageSize    int32
	TotalItems  int32
	TotalPages  int32
}

// OrganizationRepository defines the interface for organization data operations
type OrganizationRepository interface {
	// Create creates a new organization
	Create(ctx context.Context, org *domain.Organization) (*domain.Organization, error)
	
	// GetByID retrieves an organization by ID
	GetByID(ctx context.Context, orgID uuid.UUID) (*domain.Organization, error)
	
	// GetByCode retrieves an organization by code
	GetByCode(ctx context.Context, code string) (*domain.Organization, error)
	
	// Update updates an existing organization
	Update(ctx context.Context, org *domain.Organization) (*domain.Organization, error)
	
	// Delete deletes an organization by ID
	Delete(ctx context.Context, orgID uuid.UUID) error
	
	// ListByTenant retrieves all organizations for a tenant with pagination
	ListByTenant(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*domain.Organization, *PaginationResult, error)
	
	// ListAccessibleByUser retrieves all organizations accessible by a user with pagination
	ListAccessibleByUser(ctx context.Context, userID uuid.UUID, pagination PaginationParams) ([]*domain.Organization, *PaginationResult, error)
	
	// CodeExists checks if an organization code already exists (excluding a specific org ID if provided)
	CodeExists(ctx context.Context, code string, excludeOrgID *uuid.UUID) (bool, error)
	
	// CountByTenant counts organizations for a tenant
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int32, error)
}

// UserOrganizationRepository defines the interface for user-organization operations
type UserOrganizationRepository interface {
	// AddUserToOrganization adds a user to an organization with a specific role
	AddUserToOrganization(ctx context.Context, userOrg *domain.UserOrganization) error
	
	// RemoveUserFromOrganization removes a user from an organization
	RemoveUserFromOrganization(ctx context.Context, userID, orgID uuid.UUID) error
	
	// ListUsersByOrganization retrieves all user IDs in an organization
	ListUsersByOrganization(ctx context.Context, orgID uuid.UUID) ([]uuid.UUID, error)
	
	// ListOrganizationsByUser retrieves all organizations for a user
	ListOrganizationsByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Organization, error)
	
	// GetUserOrganization retrieves a specific user-organization relationship
	GetUserOrganization(ctx context.Context, userID, orgID uuid.UUID) (*domain.UserOrganization, error)
	
	// UpdateUserOrganization updates a user-organization relationship
	UpdateUserOrganization(ctx context.Context, userOrg *domain.UserOrganization) error
	
	// SetCurrentOrganization sets an organization as the current context for a user
	SetCurrentOrganization(ctx context.Context, userID, orgID uuid.UUID) error
	
	// GetCurrentOrganization retrieves the current organization for a user
	GetCurrentOrganization(ctx context.Context, userID uuid.UUID) (*domain.Organization, error)
	
	// UserHasAccessToOrganization checks if a user has access to an organization
	UserHasAccessToOrganization(ctx context.Context, userID, orgID uuid.UUID) (bool, error)
	
	// CountOrganizationsByUser counts organizations for a user
	CountOrganizationsByUser(ctx context.Context, userID uuid.UUID) (int32, error)
}
