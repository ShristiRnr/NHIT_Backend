package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/ports"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/utils"
)

// authService implements the AuthService interface
type authService struct {
	userRepo                ports.UserRepository
	sessionRepo             ports.SessionRepository
	refreshTokenRepo        ports.RefreshTokenRepository
	passwordResetRepo       ports.PasswordResetRepository
	emailVerificationRepo   ports.EmailVerificationRepository
	jwtManager              *utils.JWTManager
	emailService            utils.EmailService
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo ports.UserRepository,
	sessionRepo ports.SessionRepository,
	refreshTokenRepo ports.RefreshTokenRepository,
	passwordResetRepo ports.PasswordResetRepository,
	emailVerificationRepo ports.EmailVerificationRepository,
	jwtManager *utils.JWTManager,
	emailService utils.EmailService,
) ports.AuthService {
	return &authService{
		userRepo:              userRepo,
		sessionRepo:           sessionRepo,
		refreshTokenRepo:      refreshTokenRepo,
		passwordResetRepo:     passwordResetRepo,
		emailVerificationRepo: emailVerificationRepo,
		jwtManager:            jwtManager,
		emailService:          emailService,
	}
}

// Register registers a new user
func (s *authService) Register(ctx context.Context, tenantID uuid.UUID, name, email, password string, roles []string) (*domain.LoginResponse, error) {
	// Validate password strength
	if err := utils.ValidatePasswordStrength(password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, tenantID, email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	// Create user (this should be done via User Service gRPC call in production)
	// For now, we'll assume the user is created and return a mock response
	userID := uuid.New()
	
	// Store hashed password (via User Service)
	// In production, this would be a gRPC call to User Service
	_ = hashedPassword

	// Send verification email
	verificationToken, err := s.emailVerificationRepo.Create(ctx, userID, time.Now().Add(24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("failed to create verification token: %w", err)
	}

	// Send email with error handling
	if err := s.emailService.SendVerificationEmail(email, name, verificationToken.Token.String()); err != nil {
		// If email fails, notify user to update email
		fmt.Printf("⚠️  Failed to send verification email to %s: %v\n", email, err)
		if err := s.emailService.SendEmailUpdateNotification(email, name); err != nil {
			fmt.Printf("⚠️  Failed to send email update notification: %v\n", err)
		}
	}

	// Generate tokens
	accessToken, accessExpiresAt, err := s.jwtManager.GenerateAccessToken(
		userID.String(), email, name, tenantID.String(), "", roles, []string{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(userID.String(), tenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	if err := s.refreshTokenRepo.Create(ctx, &domain.RefreshToken{
		Token:     refreshToken,
		UserID:    userID,
		ExpiresAt: time.Unix(refreshExpiresAt, 0),
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Create session
	session := &domain.Session{
		SessionID:    uuid.New(),
		UserID:       userID,
		SessionToken: accessToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Unix(accessExpiresAt, 0),
	}
	if _, err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &domain.LoginResponse{
		Token:            accessToken,
		RefreshToken:     refreshToken,
		UserID:           userID,
		Email:            email,
		Name:             name,
		Roles:            roles,
		Permissions:      []string{},
		LastLoginAt:      time.Now(),
		LastLoginIP:      "",
		TenantID:         tenantID,
		OrgID:            nil,
		TokenExpiresAt:   accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// Login authenticates a user and returns tokens
func (s *authService) Login(ctx context.Context, email, password string, tenantID uuid.UUID, orgID *uuid.UUID) (*domain.LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Verify password
	if err := utils.VerifyPassword(user.Password, password); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check if email is verified
	if user.EmailVerifiedAt == nil {
		return nil, fmt.Errorf("email not verified. Please verify your email before logging in")
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.UserID, "", ""); err != nil {
		fmt.Printf("⚠️  Failed to update last login: %v\n", err)
	}

	// Generate tokens
	orgIDStr := ""
	if orgID != nil {
		orgIDStr = orgID.String()
	}

	accessToken, accessExpiresAt, err := s.jwtManager.GenerateAccessToken(
		user.UserID.String(), user.Email, user.Name, user.TenantID.String(), orgIDStr, []string{}, []string{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(user.UserID.String(), user.TenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	if err := s.refreshTokenRepo.Create(ctx, &domain.RefreshToken{
		Token:     refreshToken,
		UserID:    user.UserID,
		ExpiresAt: time.Unix(refreshExpiresAt, 0),
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Create session
	session := &domain.Session{
		SessionID:    uuid.New(),
		UserID:       user.UserID,
		SessionToken: accessToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Unix(accessExpiresAt, 0),
	}
	if _, err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &domain.LoginResponse{
		Token:            accessToken,
		RefreshToken:     refreshToken,
		UserID:           user.UserID,
		Email:            user.Email,
		Name:             user.Name,
		Roles:            []string{},
		Permissions:      []string{},
		LastLoginAt:      time.Now(),
		LastLoginIP:      "",
		TenantID:         user.TenantID,
		OrgID:            orgID,
		TokenExpiresAt:   accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// Logout logs out a user by invalidating their refresh token and session
func (s *authService) Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	// Delete refresh token
	if err := s.refreshTokenRepo.Delete(ctx, refreshToken); err != nil {
		fmt.Printf("⚠️  Failed to delete refresh token: %v\n", err)
	}

	// Invalidate all sessions for this user
	if err := s.InvalidateAllSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to invalidate sessions: %w", err)
	}

	return nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string, tenantID uuid.UUID, orgID *uuid.UUID) (*domain.LoginResponse, error) {
	// Validate refresh token
	userIDStr, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	// Check if refresh token exists in database
	storedUserID, err := s.refreshTokenRepo.GetUserIDByToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found or expired")
	}

	if storedUserID != userID {
		return nil, fmt.Errorf("refresh token user ID mismatch")
	}

	// Get user details
	user, err := s.userRepo.GetByEmail(ctx, tenantID, "")
	if err != nil {
		// In production, we'd have a GetByID method
		return nil, fmt.Errorf("user not found")
	}

	// Generate new tokens
	orgIDStr := ""
	if orgID != nil {
		orgIDStr = orgID.String()
	}

	accessToken, accessExpiresAt, err := s.jwtManager.GenerateAccessToken(
		user.UserID.String(), user.Email, user.Name, user.TenantID.String(), orgIDStr, []string{}, []string{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(user.UserID.String(), user.TenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Delete old refresh token
	if err := s.refreshTokenRepo.Delete(ctx, refreshToken); err != nil {
		fmt.Printf("⚠️  Failed to delete old refresh token: %v\n", err)
	}

	// Store new refresh token
	if err := s.refreshTokenRepo.Create(ctx, &domain.RefreshToken{
		Token:     newRefreshToken,
		UserID:    user.UserID,
		ExpiresAt: time.Unix(refreshExpiresAt, 0),
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	// Create new session
	session := &domain.Session{
		SessionID:    uuid.New(),
		UserID:       user.UserID,
		SessionToken: accessToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Unix(accessExpiresAt, 0),
	}
	if _, err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &domain.LoginResponse{
		Token:            accessToken,
		RefreshToken:     newRefreshToken,
		UserID:           user.UserID,
		Email:            user.Email,
		Name:             user.Name,
		Roles:            []string{},
		Permissions:      []string{},
		LastLoginAt:      time.Now(),
		LastLoginIP:      "",
		TenantID:         user.TenantID,
		OrgID:            orgID,
		TokenExpiresAt:   accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// ValidateToken validates an access token
func (s *authService) ValidateToken(ctx context.Context, token string) (*domain.TokenValidation, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return &domain.TokenValidation{Valid: false}, fmt.Errorf("invalid token: %w", err)
	}

	// Check if session exists and is valid
	session, err := s.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return &domain.TokenValidation{Valid: false}, fmt.Errorf("session not found")
	}

	if session.ExpiresAt.Before(time.Now()) {
		return &domain.TokenValidation{Valid: false}, fmt.Errorf("session expired")
	}

	userID, _ := uuid.Parse(claims.UserID)
	tenantID, _ := uuid.Parse(claims.TenantID)
	var orgID *uuid.UUID
	if claims.OrgID != "" {
		oid, _ := uuid.Parse(claims.OrgID)
		orgID = &oid
	}

	expiresAt := time.Now()
	if claims.RegisteredClaims.ExpiresAt != nil {
		expiresAt = claims.RegisteredClaims.ExpiresAt.Time
	}

	return &domain.TokenValidation{
		Valid:       true,
		UserID:      userID,
		Email:       claims.Email,
		Name:        claims.Name,
		TenantID:    tenantID,
		OrgID:       orgID,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
		ExpiresAt:   expiresAt,
	}, nil
}

// SendVerificationEmail sends a verification email to the user
func (s *authService) SendVerificationEmail(ctx context.Context, userID uuid.UUID) error {
	// Get user details (in production, call User Service)
	// For now, we'll use mock data
	email := "user@example.com"
	name := "User"

	// Create verification token
	verificationToken, err := s.emailVerificationRepo.Create(ctx, userID, time.Now().Add(24*time.Hour))
	if err != nil {
		return fmt.Errorf("failed to create verification token: %w", err)
	}

	// Send email with error handling
	if err := s.emailService.SendVerificationEmail(email, name, verificationToken.Token.String()); err != nil {
		fmt.Printf("⚠️  Failed to send verification email: %v\n", err)
		// Send notification about email failure
		if err := s.emailService.SendEmailUpdateNotification(email, name); err != nil {
			fmt.Printf("⚠️  Failed to send email update notification: %v\n", err)
		}
		return fmt.Errorf("failed to send verification email. Please update your email address")
	}

	return nil
}

// VerifyEmail verifies a user's email address
func (s *authService) VerifyEmail(ctx context.Context, userID uuid.UUID, token string) error {
	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		return fmt.Errorf("invalid verification token format")
	}

	// Verify token
	valid, err := s.emailVerificationRepo.Verify(ctx, userID, tokenUUID)
	if err != nil {
		return fmt.Errorf("failed to verify token: %w", err)
	}

	if !valid {
		return fmt.Errorf("invalid or expired verification token")
	}

	// Update user email verified status
	if err := s.userRepo.VerifyEmail(ctx, userID); err != nil {
		return fmt.Errorf("failed to update email verification status: %w", err)
	}

	// Delete verification token
	if err := s.emailVerificationRepo.Delete(ctx, userID); err != nil {
		fmt.Printf("⚠️  Failed to delete verification token: %v\n", err)
	}

	return nil
}

// ForgotPassword initiates a password reset flow
func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	// Get user by email (need tenant ID in production)
	// For now, using a mock tenant ID
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	
	user, err := s.userRepo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		// Don't reveal if user exists or not for security
		return nil
	}

	// Create password reset token
	resetToken := uuid.New()
	_, err = s.passwordResetRepo.Create(ctx, user.UserID, resetToken, time.Now().Add(1*time.Hour))
	if err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	// Send password reset email with error handling
	if err := s.emailService.SendPasswordResetEmail(email, user.Name, resetToken.String()); err != nil {
		fmt.Printf("⚠️  Failed to send password reset email: %v\n", err)
		// Send notification about email failure
		if err := s.emailService.SendEmailUpdateNotification(email, user.Name); err != nil {
			fmt.Printf("⚠️  Failed to send email update notification: %v\n", err)
		}
		return fmt.Errorf("failed to send password reset email. Please update your email address")
	}

	return nil
}

// ResetPasswordByToken resets a user's password using a reset token
func (s *authService) ResetPasswordByToken(ctx context.Context, token, newPassword string) error {
	// Validate password strength
	if err := utils.ValidatePasswordStrength(newPassword); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		return fmt.Errorf("invalid reset token format")
	}

	// Get password reset record
	resetRecord, err := s.passwordResetRepo.GetByToken(ctx, tokenUUID)
	if err != nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	// Check if token is expired
	if resetRecord.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("reset token has expired")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, resetRecord.UserID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Delete reset token
	if err := s.passwordResetRepo.Delete(ctx, tokenUUID); err != nil {
		fmt.Printf("⚠️  Failed to delete reset token: %v\n", err)
	}

	// Invalidate all sessions for security
	if err := s.InvalidateAllSessions(ctx, resetRecord.UserID); err != nil {
		fmt.Printf("⚠️  Failed to invalidate sessions: %v\n", err)
	}

	return nil
}

// InvalidateAllSessions invalidates all sessions for a user
func (s *authService) InvalidateAllSessions(ctx context.Context, userID uuid.UUID) error {
	// Get all sessions for user
	sessions, err := s.GetActiveSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get sessions: %w", err)
	}

	// Delete each session
	for _, session := range sessions {
		if err := s.sessionRepo.Delete(ctx, session.SessionID); err != nil {
			fmt.Printf("⚠️  Failed to delete session %s: %v\n", session.SessionID, err)
		}
	}

	return nil
}

// GetActiveSessions gets all active sessions for a user
func (s *authService) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	// This would need a new repository method
	// For now, return empty list
	return []*domain.Session{}, nil
}
