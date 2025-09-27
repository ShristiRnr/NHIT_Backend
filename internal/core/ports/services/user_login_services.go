package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

type UserLoginService struct {
	repo ports.UserLoginRepository
}

func NewUserLoginService(repo ports.UserLoginRepository) *UserLoginService {
	return &UserLoginService{repo: repo}
}

// RecordLogin creates a new login history record for a user
func (s *UserLoginService) RecordLogin(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) (db.UserLoginHistory, error) {
	return s.repo.Create(ctx, userID, ipAddress, userAgent)
}

// GetLoginHistory returns paginated login history for a user
func (s *UserLoginService) GetLoginHistory(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]db.UserLoginHistory, error) {
	return s.repo.List(ctx, userID, limit, offset)
}
