package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// PaginationService handles paginated data fetching
type PaginationService struct {
	repo ports.PaginationRepository
}

// NewPaginationService creates a new PaginationService instance
func NewPaginationService(repo ports.PaginationRepository) *PaginationService {
	return &PaginationService{repo: repo}
}

// ListUsers returns a paginated list of users for a tenant
func (s *PaginationService) ListUsers(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]db.User, error) {
	return s.repo.ListPaginated(ctx, tenantID, limit, offset)
}

// CountUsers returns the total number of users for a tenant
func (s *PaginationService) CountUsers(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return s.repo.CountByTenant(ctx, tenantID)
}
