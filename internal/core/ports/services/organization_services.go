package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// OrganizationService handles business logic for organizations.
type OrganizationService struct {
	repo ports.OrganizationRepository
}

// NewOrganizationService creates a new OrganizationService instance.
func NewOrganizationService(repo ports.OrganizationRepository) *OrganizationService {
	return &OrganizationService{repo: repo}
}

// CreateOrganization creates a new organization under a tenant.
func (s *OrganizationService) CreateOrganization(ctx context.Context, tenantID uuid.UUID, name string) (db.Organization, error) {
	return s.repo.Create(ctx, tenantID, name)
}

// GetOrganization fetches an organization by ID.
func (s *OrganizationService) GetOrganization(ctx context.Context, orgID uuid.UUID) (db.Organization, error) {
	return s.repo.Get(ctx, orgID)
}

// ListOrganizations lists all organizations for a tenant.
func (s *OrganizationService) ListOrganizations(ctx context.Context, tenantID uuid.UUID) ([]db.Organization, error) {
	return s.repo.List(ctx, tenantID)
}

// UpdateOrganization updates an organization's name.
func (s *OrganizationService) UpdateOrganization(ctx context.Context, orgID uuid.UUID, name string) (db.Organization, error) {
	return s.repo.Update(ctx, orgID, name)
}

// DeleteOrganization removes an organization.
func (s *OrganizationService) DeleteOrganization(ctx context.Context, orgID uuid.UUID) error {
	return s.repo.Delete(ctx, orgID)
}
