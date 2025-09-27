package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

type RoleService struct {
	repo ports.RoleRepository
}

func NewRoleService(repo ports.RoleRepository) *RoleService {
	return &RoleService{repo: repo}
}

// Create a new role
func (s *RoleService) CreateRole(ctx context.Context, tenantID uuid.UUID, name string) (db.Role, error) {
	params := db.CreateRoleParams{
		TenantID: tenantID,
		Name:     name,
	}
	return s.repo.Create(ctx, params)
}

// Get a role by ID
func (s *RoleService) GetRole(ctx context.Context, roleID uuid.UUID) (db.Role, error) {
	return s.repo.Get(ctx, roleID)
}

// List roles for a tenant
func (s *RoleService) ListRoles(ctx context.Context, tenantID uuid.UUID) ([]db.Role, error) {
	return s.repo.List(ctx, tenantID)
}

// Update a role
func (s *RoleService) UpdateRole(ctx context.Context, roleID uuid.UUID, name string) (db.Role, error) {
	params := db.UpdateRoleParams{
		RoleID: roleID,
		Name:   name,
	}
	return s.repo.Update(ctx, params)
}

// Delete a role
func (s *RoleService) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	return s.repo.Delete(ctx, roleID)
}

// Assign role to user
func (s *RoleService) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	params := db.AssignRoleToUserParams{
		UserID: userID,
		RoleID: roleID,
	}
	return s.repo.AssignRoleToUser(ctx, params)
}

// Assign permission to role
func (s *RoleService) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	params := db.AssignPermissionToRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	}
	return s.repo.AssignPermissionToRole(ctx, params)
}

// List roles of a user
func (s *RoleService) ListRolesOfUser(ctx context.Context, userID uuid.UUID) ([]db.Role, error) {
	return s.repo.ListRolesOfUser(ctx, userID)
}

// List permissions of a user via roles
func (s *RoleService) ListPermissionsOfUser(ctx context.Context, userID uuid.UUID) ([]db.Permission, error) {
	return s.repo.ListPermissionsOfUserViaRoles(ctx, userID)
}
