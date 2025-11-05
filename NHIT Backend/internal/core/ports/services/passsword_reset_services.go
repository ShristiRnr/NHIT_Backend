package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

type PasswordResetService struct {
	repo ports.PasswordResetRepository
}

func NewPasswordResetService(repo ports.PasswordResetRepository) *PasswordResetService {
	return &PasswordResetService{repo: repo}
}

// CreateToken generates a new password reset token for a user
func (s *PasswordResetService) CreateToken(ctx context.Context, userID uuid.UUID, token uuid.UUID, expiresAt time.Time) (db.PasswordReset, error) {
	params := db.CreatePasswordResetTokenParams{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	createdToken, err := s.repo.Create(ctx, params)
	if err != nil {
		return db.PasswordReset{}, err
	}
	return createdToken, nil
}

// GetToken retrieves a password reset token by token UUID
func (s *PasswordResetService) GetToken(ctx context.Context, token uuid.UUID) (db.PasswordReset, error) {
	return s.repo.Get(ctx, token)
}

// DeleteToken removes a password reset token
func (s *PasswordResetService) DeleteToken(ctx context.Context, token uuid.UUID) error {
	return s.repo.Delete(ctx, token)
}
