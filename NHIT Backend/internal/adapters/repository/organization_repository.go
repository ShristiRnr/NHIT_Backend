package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// OrganizationRepo implements ports.OrganizationRepository using sqlc-generated db.
type OrganizationRepo struct {
	q *db.Queries
}

// NewOrganizationRepo creates a new repository instance.
func NewOrganizationRepo(q *db.Queries) ports.OrganizationRepository {
	return &OrganizationRepo{q: q}
}

// Create inserts a new organization.
func (r *OrganizationRepo) Create(ctx context.Context, tenantID uuid.UUID, name string) (db.Organization, error) {
	arg := db.CreateOrganizationParams{
		TenantID: tenantID,
		Name:     name,
	}
	return r.q.CreateOrganization(ctx, arg)
}

// Get retrieves an organization by its ID.
func (r *OrganizationRepo) Get(ctx context.Context, orgID uuid.UUID) (db.Organization, error) {
	return r.q.GetOrganization(ctx, orgID)
}

// List returns all organizations for a tenant.
func (r *OrganizationRepo) List(ctx context.Context, tenantID uuid.UUID) ([]db.Organization, error) {
	return r.q.ListOrganizationsByTenant(ctx, tenantID)
}

// Update modifies an existing organization.
func (r *OrganizationRepo) Update(ctx context.Context, orgID uuid.UUID, name string) (db.Organization, error) {
	arg := db.UpdateOrganizationParams{
		OrgID: orgID,
		Name:  name,
	}
	return r.q.UpdateOrganization(ctx, arg)
}

// Delete removes an organization by its ID.
func (r *OrganizationRepo) Delete(ctx context.Context, orgID uuid.UUID) error {
	return r.q.DeleteOrganization(ctx, orgID)
}