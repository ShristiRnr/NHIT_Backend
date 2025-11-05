package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// RoleRepo implements ports.RoleRepository using sqlc-generated db.
type RoleRepo struct {
	q *db.Queries
}

// NewRoleRepo creates a new Role repository instance.
func NewRoleRepo(q *db.Queries) ports.RoleRepository {
	return &RoleRepo{q: q}
}

// Create inserts a new role for a tenant.
func (r *RoleRepo) Create(ctx context.Context, arg db.CreateRoleParams) (db.Role, error) {
	return r.q.CreateRole(ctx, arg)
}

// Get retrieves a role by its ID.
func (r *RoleRepo) Get(ctx context.Context, roleID uuid.UUID) (db.Role, error) {
	return r.q.GetRole(ctx, roleID)
}

// List returns all roles for a given tenant.
func (r *RoleRepo) List(ctx context.Context, tenantID uuid.UUID) ([]db.Role, error) {
	return r.q.ListRolesByTenant(ctx, tenantID)
}

// Update modifies an existing role.
func (r *RoleRepo) Update(ctx context.Context, arg db.UpdateRoleParams) (db.Role, error) {
	return r.q.UpdateRole(ctx, arg)
}

// Delete removes a role by its ID.
func (r *RoleRepo) Delete(ctx context.Context, roleID uuid.UUID) error {
	return r.q.DeleteRole(ctx, roleID)
}

// AssignRoleToUser assigns a role to a user.
func (r *RoleRepo) AssignRoleToUser(ctx context.Context, arg db.AssignRoleToUserParams) error {
	return r.q.AssignRoleToUser(ctx, arg)
}

// AssignPermissionToRole assigns a permission to a role.
func (r *RoleRepo) AssignPermissionToRole(ctx context.Context, arg db.AssignPermissionToRoleParams) error {
	return r.q.AssignPermissionToRole(ctx, arg)
}

// ListRolesOfUser returns all roles assigned to a user.
func (r *RoleRepo) ListRolesOfUser(ctx context.Context, userID uuid.UUID) ([]db.Role, error) {
	return r.q.ListRolesOfUser(ctx, userID)
}

// ListPermissionsOfUserViaRoles returns all permissions a user has via roles.
func (r *RoleRepo) ListPermissionsOfUserViaRoles(ctx context.Context, userID uuid.UUID) ([]db.Permission, error) {
	return r.q.ListPermissionsOfUserViaRoles(ctx, userID)
}
