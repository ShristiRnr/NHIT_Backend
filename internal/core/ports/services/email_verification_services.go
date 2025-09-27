package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

type EmailVerificationService struct {
	repo        ports.EmailVerificationRepository
	userRepo    ports.UserRepository
	emailSender ports.EmailSender
	baseURL     string
}

func NewEmailVerificationService(
	repo ports.EmailVerificationRepository,
	userRepo ports.UserRepository,
	emailSender ports.EmailSender,
	baseURL string,
) *EmailVerificationService {
	return &EmailVerificationService{
		repo:        repo,
		userRepo:    userRepo,
		emailSender: emailSender,
		baseURL:     baseURL,
	}
}

// Create token & send email
func (s *EmailVerificationService) SendVerificationEmail(ctx context.Context, userID uuid.UUID, email string) (uuid.UUID, error) {
	token := uuid.New()

	_, err := s.repo.Insert(ctx, userID, token)
	if err != nil {
		return uuid.Nil, err
	}

	link := fmt.Sprintf("%s/verify-email?token=%s", s.baseURL, token.String())
	expiry := time.Now().Add(24 * time.Hour).Format(time.RFC1123)

	if err := s.emailSender.SendVerificationEmail(ctx, email, link, expiry); err != nil {
		return uuid.Nil, err
	}

	return token, nil
}

// Verify user with token
func (s *EmailVerificationService) VerifyEmail(ctx context.Context, token uuid.UUID) error {
	verification, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return err
	}

	// mark user as verified
	if err := s.userRepo.MarkEmailVerified(ctx, verification.UserID.UUID); err != nil {
		return err
	}

	// delete token after successful verification
	return s.repo.DeleteByUserID(ctx, verification.UserID.UUID)
}
