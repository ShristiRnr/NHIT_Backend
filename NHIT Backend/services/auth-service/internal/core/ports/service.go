package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/domain"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Authentication
	Register(ctx context.Context, tenantID uuid.UUID, name, email, password string, roles []string) (*domain.LoginResponse, error)
	Login(ctx context.Context, email, password string, tenantID uuid.UUID, orgID *uuid.UUID) (*domain.LoginResponse, error)
	Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string, tenantID uuid.UUID, orgID *uuid.UUID) (*domain.LoginResponse, error)
	ValidateToken(ctx context.Context, token string) (*domain.TokenValidation, error)
	
	// Email Verification
	SendVerificationEmail(ctx context.Context, userID uuid.UUID) error
	VerifyEmail(ctx context.Context, userID uuid.UUID, token string) error
	
	// Password Reset
	ForgotPassword(ctx context.Context, email string) error
	ResetPasswordByToken(ctx context.Context, token, newPassword string) error
	
	// Session Management
	InvalidateAllSessions(ctx context.Context, userID uuid.UUID) error
	GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error)
}
