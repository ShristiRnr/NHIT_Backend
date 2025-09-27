package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// PaginationRepo implements ports.PaginationRepository using sqlc-generated db.
type PaginationRepo struct {
	q *db.Queries
}

// NewPaginationRepo creates a new repository instance.
func NewPaginationRepo(q *db.Queries) ports.PaginationRepository {
	return &PaginationRepo{q: q}
}

// ListPaginated returns a paginated list of users for a tenant.
func (r *PaginationRepo) ListPaginated(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]db.User, error) {
	params := db.PaginatedUsersByTenantParams{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	}
	return r.q.PaginatedUsersByTenant(ctx, params)
}

// CountByTenant returns the total number of users for a tenant.
func (r *PaginationRepo) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return r.q.CountUsersByTenant(ctx, tenantID)
}
