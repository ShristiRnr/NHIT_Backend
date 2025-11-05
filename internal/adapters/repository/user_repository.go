package repository

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// UserRepo implements ports.UserRepository using sqlc-generated db.Queries
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

// GetByEmail retrieves a user by email
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (db.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

// GetUserByToken retrieves a user by session token
func (r *UserRepo) GetUserByToken(ctx context.Context, token string) (db.User, error) {
	return r.q.GetUserByToken(ctx, token)
}

// GetUserPermissions retrieves all permissions for a user
func (r *UserRepo) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return r.q.GetUserPermissions(ctx, userID)
}

// Update modifies an existing user
func (r *UserRepo) Update(ctx context.Context, user db.UpdateUserParams) (db.User, error) {
	return r.q.UpdateUser(ctx, user)
}

// UpdatePassword changes the user's password
func (r *UserRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) (db.User, error) {
	params := db.UpdateUserPasswordParams{
		UserID:   userID,
		Password: hashedPassword,
	}
	return r.q.UpdateUserPassword(ctx, params)
}

// ConfirmPassword checks if a password matches the stored hash
func (r *UserRepo) ConfirmPassword(ctx context.Context, userID uuid.UUID, password string) (bool, error) {
	userPassword, err := r.q.GetUserPassword(ctx, userID)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(password))
	if err != nil {
		return false, nil
	}

	return true, nil
}

// ListByTenant returns a paginated list of users for a tenant
func (r *UserRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit int32, offset int32) ([]db.User, error) {
	params := db.ListUsersByTenantParams{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	}
	return r.q.ListUsersByTenant(ctx, params)
}

// MarkEmailVerified sets email_verified_at = NOW()
func (r *UserRepo) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	return r.q.MarkEmailVerified(ctx, userID)
}

func (r *UserRepo) GetRolesAndPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// Use the Queries method that lists permissions via roles
	perms, err := r.q.ListPermissionsOfUserViaRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert db.Permission to string slice
	var permissions []string
	for _, p := range perms {
		permissions = append(permissions, p.Name)
	}
	return permissions, nil
}
