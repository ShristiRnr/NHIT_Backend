package repository

import (
	"context"

	sqlc "github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
	"github.com/google/uuid"
)

type roleRepository struct {
	queries *sqlc.Queries
}

func NewRoleRepository(queries *sqlc.Queries) ports.RoleRepository {
	return &roleRepository{queries: queries}
}

func (r *roleRepository) Create(ctx context.Context, role *domain.Role) (*domain.Role, error) {
	parentOrgID := uuid.NullUUID{}
	if role.OrgID != nil {
		parentOrgID = uuid.NullUUID{UUID: *role.OrgID, Valid: true}
	}

	params := sqlc.CreateRoleParams{
		TenantID:     role.TenantID,
		ParentOrgID:  parentOrgID,
		Name:         role.Name,
		Description:  role.Description,
		Permissions:  role.Permissions,
		IsSystemRole: role.IsSystemRole,
		CreatedBy:    &role.CreatedBy,
	}

	dbRole, err := r.queries.CreateRole(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainRole(dbRole), nil
}

func (r *roleRepository) GetByID(ctx context.Context, roleID uuid.UUID) (*domain.Role, error) {
	dbRole, err := r.queries.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return toDomainRole(dbRole), nil
}

func (r *roleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error) {
	dbRoles, err := r.queries.ListRolesByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	roles := make([]*domain.Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		roles[i] = toDomainRole(dbRole)
	}
	return roles, nil
}

func (r *roleRepository) ListByTenantAndOrg(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID) ([]*domain.Role, error) {
	parentOrgID := uuid.NullUUID{}
	if orgID != nil {
		parentOrgID = uuid.NullUUID{UUID: *orgID, Valid: true}
	}

	params := sqlc.ListRolesByTenantAndOrgParams{
		TenantID:    tenantID,
		ParentOrgID: parentOrgID,
	}

	dbRoles, err := r.queries.ListRolesByTenantAndOrg(ctx, params)
	if err != nil {
		return nil, err
	}

	roles := make([]*domain.Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		roles[i] = toDomainRole(dbRole)
	}
	return roles, nil
}

func (r *roleRepository) ListByTenantAndOrgIncludingSystem(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID) ([]*domain.Role, error) {
	parentOrgID := uuid.NullUUID{}
	if orgID != nil {
		parentOrgID = uuid.NullUUID{UUID: *orgID, Valid: true}
	}

	params := sqlc.ListRolesByOrganizationIncludingSystemParams{
		TenantID:    tenantID,
		ParentOrgID: parentOrgID,
	}

	dbRoles, err := r.queries.ListRolesByOrganizationIncludingSystem(ctx, params)
	if err != nil {
		return nil, err
	}

	roles := make([]*domain.Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		roles[i] = toDomainRole(dbRole)
	}
	return roles, nil
}

func (r *roleRepository) Update(ctx context.Context, role *domain.Role) (*domain.Role, error) {
	params := sqlc.UpdateRoleParams{
		RoleID:      role.RoleID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: role.Permissions,
	}

	dbRole, err := r.queries.UpdateRole(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainRole(dbRole), nil
}

func (r *roleRepository) Delete(ctx context.Context, roleID uuid.UUID) error {
	return r.queries.DeleteRole(ctx, roleID)
}

func toDomainRole(dbRole *sqlc.Role) *domain.Role {
	var orgID *uuid.UUID
	if dbRole.ParentOrgID.Valid {
		id := dbRole.ParentOrgID.UUID
		orgID = &id
	}

	var createdBy string
	if dbRole.CreatedBy != nil {
		createdBy = *dbRole.CreatedBy
	}

	return &domain.Role{
		RoleID:       dbRole.RoleID,
		TenantID:     dbRole.TenantID,
		OrgID:        orgID,
		Name:         dbRole.Name,
		Description:  dbRole.Description,
		Permissions:  dbRole.Permissions,
		IsSystemRole: dbRole.IsSystemRole,
		CreatedBy:    createdBy,
		CreatedAt:    dbRole.CreatedAt.Time,
		UpdatedAt:    dbRole.UpdatedAt.Time,
	}
}
