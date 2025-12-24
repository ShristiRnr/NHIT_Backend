package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/utils"
	"github.com/google/uuid"
)

// GenerateOTP generates a random numeric OTP of the specified length
func generateOTP(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("OTP length must be positive")
	}

	// Define characters for OTP
	const digits = "0123456789"
	maxNum := big.NewInt(int64(len(digits)))

	otp := make([]byte, length)
	for i := 0; i < length; i++ {
		// Generate random index
		num, err := rand.Int(rand.Reader, maxNum)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		// Select character at the random index
		otp[i] = digits[num.Int64()]
	}

	return string(otp), nil
}

// ForgotPasswordWithOTP initiates a password reset flow using an OTP
func (s *authService) ForgotPasswordWithOTP(ctx context.Context, email string, tenantID uuid.UUID) error {
	// Find user
	user, err := s.userRepo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		// Don't reveal if user exists or not for security
		log.Printf("User with email %s not found for tenant %s: %v", email, tenantID, err)
		return nil
	}

	// Generate OTP (5 digits)
	otp, err := generateOTP(5)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Create password reset token with OTP
	// Create password reset token with OTP
	expiresAt := time.Now().Add(5 * time.Minute) // OTPs expire in 5 minutes as requested
	_, err = s.passwordResetRepo.CreateWithOTP(ctx, user.UserID, otp, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create password reset OTP: %w", err)
	}

	// Send OTP via notification service
	if s.notificationClient != nil {
		if err := s.notificationClient.SendOtpPasswordResetEmail(ctx, email, user.Name, otp); err != nil {
			log.Printf("⚠️ Failed to send OTP email via notification service: %v", err)
			// Fall back to old method if available
			if s.emailService != nil {
				if err := s.emailService.SendPasswordResetEmail(email, user.Name, otp); err != nil {
					log.Printf("⚠️ Failed to send fallback OTP email: %v", err)
				}
			}
		}
	}

	// Log event
	if s.kafkaPublisher != nil {
		eventTimestamp := time.Now()
		eventData := map[string]interface{}{
			"user_id":      user.UserID.String(),
			"tenant_id":    user.TenantID.String(),
			"email":        user.Email,
			"name":         user.Name,
			"otp_sent":     true,
			"expires_at":   expiresAt.Format(time.RFC3339),
			"requested_at": eventTimestamp.Format(time.RFC3339),
		}
		if err := s.kafkaPublisher.Publish(ctx, "user.password_reset_otp_requested", eventData); err != nil {
			log.Printf("⚠️ Failed to publish password reset OTP event: %v", err)
		}
	}

	return nil
}

// VerifyOTPAndResetPassword verifies an OTP and resets the user's password
func (s *authService) VerifyOTPAndResetPassword(ctx context.Context, email, otp, newPassword string, tenantID uuid.UUID) error {
	// Find user
	user, err := s.userRepo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Validate password strength
	if err := utils.ValidatePasswordStrength(newPassword); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Verify OTP
	resetRecord, err := s.passwordResetRepo.GetByUserIDAndOTP(ctx, user.UserID, otp)
	if err != nil {
		return fmt.Errorf("invalid OTP")
	}

	// Check if OTP is expired
	if resetRecord.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("OTP has expired")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password in users table
	if err := s.userRepo.UpdatePassword(ctx, user.UserID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// TRIGGER SYNC: Update password in tenants and organizations tables if email matches
	// This ensures consistency across the platform as requested
	log.Printf("Syncing password for %s across platforms...", user.Email)
	
	if err := s.userRepo.UpdateTenantPassword(ctx, user.Email, hashedPassword); err != nil {
		log.Printf("⚠️  Failed to sync tenant password for %s: %v", user.Email, err)
		// We don't fail the whole operation if sync fails, but we log it
	}

	if err := s.userRepo.UpdateOrganizationSuperAdminPassword(ctx, user.Email, hashedPassword); err != nil {
		log.Printf("⚠️  Failed to sync organization super admin password for %s: %v", user.Email, err)
		// We don't fail the whole operation if sync fails, but we log it
	}

	// Delete used OTP
	if err := s.passwordResetRepo.Delete(ctx, resetRecord.ID); err != nil {
		log.Printf("⚠️ Failed to delete reset token: %v", err)
	}

	// Invalidate all sessions for security
	if err := s.InvalidateAllSessions(ctx, user.UserID); err != nil {
		log.Printf("⚠️ Failed to invalidate sessions: %v", err)
	}

	return nil
}
