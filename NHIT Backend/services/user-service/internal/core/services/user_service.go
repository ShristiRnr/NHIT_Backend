package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
)

type userService struct {
	userRepo     ports.UserRepository
	userRoleRepo ports.UserRoleRepository
}

// NewUserService creates a new user service instance
func NewUserService(userRepo ports.UserRepository, userRoleRepo ports.UserRoleRepository) ports.UserService {
	return &userService{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
	}
}

func (s *userService) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	// TODO: Hash password before storing (add bcrypt after module setup)
	// For now, store password as-is (NOT PRODUCTION READY)
	
	// Create user in repository
	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

func (s *userService) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	// TODO: Hash password if being updated (add bcrypt after module setup)
	
	updatedUser, err := s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updatedUser, nil
}

func (s *userService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (s *userService) ListUsersByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]*domain.User, error) {
	users, err := s.userRepo.ListByTenant(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (s *userService) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if err := s.userRoleRepo.AssignRole(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}
	return nil
}

func (s *userService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	roles, err := s.userRoleRepo.ListRolesByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return roles, nil
}
