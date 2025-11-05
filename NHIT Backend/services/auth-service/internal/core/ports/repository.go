package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/domain"
)

// SessionRepository defines the interface for session operations
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) (*domain.Session, error)
	GetByID(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error)
	GetByToken(ctx context.Context, token string) (*domain.Session, error)
	Delete(ctx context.Context, sessionID uuid.UUID) error
}

// RefreshTokenRepository defines the interface for refresh token operations
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error)
	Delete(ctx context.Context, token string) error
}

// PasswordResetRepository defines the interface for password reset operations
type PasswordResetRepository interface {
	Create(ctx context.Context, userID uuid.UUID, token uuid.UUID, expiresAt time.Time) (*domain.PasswordReset, error)
	GetByToken(ctx context.Context, token uuid.UUID) (*domain.PasswordReset, error)
	Delete(ctx context.Context, token uuid.UUID) error
}

// EmailVerificationRepository defines the interface for email verification operations
type EmailVerificationRepository interface {
	Create(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (*domain.EmailVerificationToken, error)
	Verify(ctx context.Context, userID uuid.UUID, token uuid.UUID) (bool, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}

// UserRepository defines user-related operations needed by auth service
type UserRepository interface {
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*UserData, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) error
	VerifyEmail(ctx context.Context, userID uuid.UUID) error
}

// UserData represents user data needed for authentication
type UserData struct {
	UserID          uuid.UUID
	TenantID        uuid.UUID
	Email           string
	Name            string
	Password        string
	EmailVerifiedAt *time.Time
}
