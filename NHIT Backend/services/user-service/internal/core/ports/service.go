package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
)

// UserService defines the interface for user business logic
type UserService interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	ListUsersByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]*domain.User, error)
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error)
}
