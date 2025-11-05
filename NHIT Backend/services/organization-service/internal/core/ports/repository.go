package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/domain"
)

// OrganizationRepository defines the interface for organization data operations
type OrganizationRepository interface {
	Create(ctx context.Context, org *domain.Organization) (*domain.Organization, error)
	GetByID(ctx context.Context, orgID uuid.UUID) (*domain.Organization, error)
	Update(ctx context.Context, org *domain.Organization) (*domain.Organization, error)
	Delete(ctx context.Context, orgID uuid.UUID) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.Organization, error)
}

// UserOrganizationRepository defines the interface for user-organization operations
type UserOrganizationRepository interface {
	AddUserToOrganization(ctx context.Context, userID, orgID, roleID uuid.UUID) error
	RemoveUserFromOrganization(ctx context.Context, userID, orgID uuid.UUID) error
	ListUsersByOrganization(ctx context.Context, orgID uuid.UUID) ([]uuid.UUID, error)
	ListOrganizationsByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Organization, error)
}

// TenantRepository defines the interface for tenant operations
type TenantRepository interface {
	Create(ctx context.Context, tenant *domain.Tenant) (*domain.Tenant, error)
	GetByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error)
}
