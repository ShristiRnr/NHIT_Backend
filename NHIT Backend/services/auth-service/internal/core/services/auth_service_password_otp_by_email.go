package services

import (
	"context"
	"log"
)

// ForgotPasswordWithOTPByEmail initiates a password reset flow using an OTP
// It automatically fetches the tenant_id from the user's email
func (s *authService) ForgotPasswordWithOTPByEmail(ctx context.Context, email string) error {
	// Query user by email across all tenants to find their tenant_id
	user, err := s.userRepo.GetByEmailGlobal(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not for security
		log.Printf("User with email %s not found: %v", email, err)
		return nil  // Return success even if user doesn't exist for security
	}

	// Call the existing method with the fetched tenant_id
	return s.ForgotPasswordWithOTP(ctx, email, user.TenantID)
}

// VerifyOTPAndResetPasswordByEmail verifies an OTP and resets the user's password
// It automatically fetches the tenant_id from the user's email
func (s *authService) VerifyOTPAndResetPasswordByEmail(ctx context.Context, email, otp, newPassword string) error {
	// Query user by email across all tenants to find their tenant_id
	user, err := s.userRepo.GetByEmailGlobal(ctx, email)
	if err != nil {
		return err
	}

	// Call the existing method with the fetched tenant_id
	return s.VerifyOTPAndResetPassword(ctx, email, otp, newPassword, user.TenantID)
}
