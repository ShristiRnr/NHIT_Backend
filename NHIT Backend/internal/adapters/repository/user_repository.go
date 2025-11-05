package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// UserRepo implements UserRepository using sqlc-generated db.Queries
type UserRepo struct {
	q *db.Queries
}

// NewUserRepo creates a new UserRepo
func NewUserRepo(q *db.Queries) ports.UserRepository {
	return &UserRepo{q: q}
}

// Create inserts a new user
func (r *UserRepo) Create(ctx context.Context, user db.CreateUserParams) (db.User, error) {
	return r.q.CreateUser(ctx, user)
}

// Delete removes a user by ID
func (r *UserRepo) Delete(ctx context.Context, userID uuid.UUID) error {
	return r.q.DeleteUser(ctx, userID)
}

// Get retrieves a user by ID
func (r *UserRepo) Get(ctx context.Context, userID uuid.UUID) (db.User, error) {
	return r.q.GetUser(ctx, userID)
}

// GetRolesAndPermissions retrieves a user's permissions via roles
func (r *UserRepo) GetRolesAndPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	perms, err := r.q.ListPermissionsOfUserViaRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	var permissions []string
	for _, p := range perms {
		permissions = append(permissions, p.Name)
	}
	return permissions, nil
}

// ListByTenant retrieves a paginated list of users for a tenant
func (r *UserRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit int32, offset int32) ([]db.User, error) {
	params := db.ListUsersByTenantParams{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	}
	return r.q.ListUsersByTenant(ctx, params)
}

// Update modifies an existing user
func (r *UserRepo) Update(ctx context.Context, user db.UpdateUserParams) (db.User, error) {
	return r.q.UpdateUser(ctx, user)
}