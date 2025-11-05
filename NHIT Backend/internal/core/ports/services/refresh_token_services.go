package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

type RefreshTokenService struct {
	repo ports.RefreshTokenRepository
}

func NewRefreshTokenService(repo ports.RefreshTokenRepository) *RefreshTokenService {
	return &RefreshTokenService{repo: repo}
}

// CreateToken inserts a new refresh token for a user
func (s *RefreshTokenService) CreateToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	return s.repo.Create(ctx, userID, token, expiresAt)
}

// GetUserIDByToken returns the user ID associated with a refresh token
func (s *RefreshTokenService) GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error) {
	return s.repo.GetUserIDByToken(ctx, token)
}

// DeleteToken removes a refresh token
func (s *RefreshTokenService) DeleteToken(ctx context.Context, token string) error {
	return s.repo.Delete(ctx, token)
}
