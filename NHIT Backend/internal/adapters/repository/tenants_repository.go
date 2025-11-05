package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// TenantRepo implements ports.TenantRepository using sqlc-generated db.Queries
type TenantRepo struct {
	q *db.Queries
}

// NewTenantRepo creates a new repository instance
func NewTenantRepo(q *db.Queries) ports.TenantRepository {
	return &TenantRepo{q: q}
}

// Create inserts a new tenant
func (r *TenantRepo) Create(ctx context.Context, tenant db.CreateTenantParams) (db.Tenant, error) {
	return r.q.CreateTenant(ctx, tenant)
}

// Get retrieves a tenant by ID
func (r *TenantRepo) Get(ctx context.Context, tenantID uuid.UUID) (db.Tenant, error) {
	return r.q.GetTenant(ctx, tenantID)
}
