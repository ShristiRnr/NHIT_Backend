package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

type TenantService struct {
	repo ports.TenantRepository
}

func NewTenantService(repo ports.TenantRepository) *TenantService {
	return &TenantService{repo: repo}
}

// CreateTenant creates a new tenant
func (s *TenantService) CreateTenant(ctx context.Context, name string, superAdminID *uuid.UUID) (db.Tenant, error) {
	tenant := db.CreateTenantParams{
		Name: name,
	}

	if superAdminID != nil {
		tenant.SuperAdminUserID = uuid.NullUUID{UUID: *superAdminID, Valid: true}
	} else {
		tenant.SuperAdminUserID = uuid.NullUUID{Valid: false}
	}

	return s.repo.Create(ctx, tenant)
}

// GetTenant retrieves a tenant by ID
func (s *TenantService) GetTenant(ctx context.Context, tenantID uuid.UUID) (db.Tenant, error) {
	return s.repo.Get(ctx, tenantID)
}
