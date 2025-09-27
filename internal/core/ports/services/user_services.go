package services

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
	"github.com/google/uuid"
)

type AuthUser struct {
    ID    uuid.UUID
    Name  string
    Email string
    Roles []string
}

type UserService struct {
	repo ports.UserRepository
}

func NewUserService(repo ports.UserRepository) *UserService {
    return &UserService{repo: repo}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, user db.CreateUserParams) (db.User, error) {
	return s.repo.Create(ctx, user)
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (db.User, error) {
	return s.repo.Get(ctx, userID)
}

// UpdateUser modifies an existing user
func (s *UserService) UpdateUser(ctx context.Context, user db.UpdateUserParams) (db.User, error) {
	return s.repo.Update(ctx, user)
}

// DeleteUser removes a user by ID
func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.repo.Delete(ctx, userID)
}

// ListUsers returns paginated users for a tenant
func (s *UserService) ListUsers(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]db.User, error) {
	return s.repo.ListByTenant(ctx, tenantID, limit, offset)
}

// GetUserPermissions retrieves a user's permissions via roles
func (s *UserService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.repo.GetRolesAndPermissions(ctx, userID)
}

func (s *UserService) GetUserFromToken(ctx context.Context, token string) (*AuthUser, error) {
    u, err := s.repo.GetUserByToken(ctx, token)
    if err != nil {
        return nil, err
    }

    roles, _ := s.repo.GetRolesAndPermissions(ctx, u.UserID)

    return &AuthUser{
        ID:    u.UserID,
        Name:  u.Name,
        Email: u.Email,
        Roles: roles,
    }, nil
}

func (s *UserService) UserHasPermission(ctx context.Context, userID uuid.UUID, permission string) bool {
    perms, _ := s.repo.GetUserPermissions(ctx, userID)
    for _, p := range perms {
        if p == permission {
            return true
        }
    }
    return false
}


