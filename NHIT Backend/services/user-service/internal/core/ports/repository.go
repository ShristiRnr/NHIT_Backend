package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) (*domain.User, error)
	Delete(ctx context.Context, userID uuid.UUID) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]*domain.User, error)
}

// UserRoleRepository defines the interface for user-role operations
type UserRoleRepository interface {
	AssignRole(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error
	ListRolesByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error)
}
