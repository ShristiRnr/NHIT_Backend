package services

import (
	"context"
	"time"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt" 
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)


type PasswordResetService struct {
	repo ports.PasswordResetRepository
	userRepo  ports.UserRepository
	tokenTTL  time.Duration
	emailSender ports.EmailSender
}

func NewPasswordResetService(repo ports.PasswordResetRepository, userRepo ports.UserRepository, ttl time.Duration, emailSender ports.EmailSender) *PasswordResetService {
	return &PasswordResetService{repo: repo, userRepo: userRepo, tokenTTL: ttl, emailSender: emailSender}
}

// CreateToken generates a new password reset token for an email
func (s *PasswordResetService) CreateToken(ctx context.Context, email string) (uuid.UUID, error) {
    user, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        return uuid.Nil, errors.New("if this email is registered, you will receive a password reset link")
    }

    token := uuid.New()
    expiresAt := time.Now().Add(s.tokenTTL)

    _, err = s.repo.Create(ctx, email, token, expiresAt)
    if err != nil {
        return uuid.Nil, err
    }

    // Password reset link
    resetLink := fmt.Sprintf("https://yourapp.com/reset-password/%s", token.String())

    // Yaha mail bhejna hai
    go func() {
        _ = s.emailSender.SendResetPasswordEmail(ctx, user.Email, resetLink, expiresAt.Format(time.RFC1123))
    }()

    return token, nil
}

// GetToken retrieves a password reset token by token UUID
func (s *PasswordResetService) GetToken(ctx context.Context, token uuid.UUID) (*uuid.UUID, error) {
	pr, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if pr.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return &token, nil
}


// DeleteToken removes a password reset token
func (s *PasswordResetService) DeleteToken(ctx context.Context, token uuid.UUID) error {
	return s.repo.Delete(ctx, token)
}

func (s *PasswordResetService) ResetPassword(ctx context.Context, token uuid.UUID, newPassword string) error {
	pr, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return errors.New("invalid token")
	}

	if pr.ExpiresAt.Before(time.Now()) {
		return errors.New("token expired")
	}

	// Fetch user
	user, err := s.userRepo.GetByEmail(ctx, pr.Email)
	if err != nil {
		return errors.New("user not found")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update user password
	_, err = s.userRepo.UpdatePassword(ctx, user.UserID, string(hashedPassword))
	if err != nil {
		return err
	}

	// Delete the token
	return s.repo.Delete(ctx, token)
}