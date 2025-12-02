package notifier

import (
	"context"
	"fmt"
	"log"
)

// MockNotificationClient provides a mock implementation of the notification client
type MockNotificationClient struct {
	IsEnabled bool
}

// NewMockNotificationClient creates a new mock notification client
func NewMockNotificationClient(enabled bool) *MockNotificationClient {
	return &MockNotificationClient{
		IsEnabled: enabled,
	}
}

// SendPasswordResetEmail sends a password reset email using the mock client
func (c *MockNotificationClient) SendPasswordResetEmail(ctx context.Context, email, name, resetToken string) error {
	if !c.IsEnabled {
		return fmt.Errorf("mock notification client is disabled")
	}

	log.Printf("MOCK: Sending password reset email to %s (%s) with token %s", name, email, resetToken)
	log.Printf("MOCK: Password reset link would be: http://localhost:3000/reset-password?token=%s", resetToken)

	return nil
}

// SendEmailUpdateNotification sends an email update notification using the mock client
func (c *MockNotificationClient) SendEmailUpdateNotification(ctx context.Context, email, name string) error {
	if !c.IsEnabled {
		return fmt.Errorf("mock notification client is disabled")
	}

	log.Printf("MOCK: Sending email update notification to %s (%s)", name, email)

	return nil
}

// SendOtpPasswordResetEmail sends a password reset email with OTP using the mock client
func (c *MockNotificationClient) SendOtpPasswordResetEmail(ctx context.Context, email, name, otp string) error {
	if !c.IsEnabled {
		return fmt.Errorf("mock notification client is disabled")
	}

	log.Printf("MOCK: Sending OTP password reset email to %s (%s) with OTP: %s", name, email, otp)

	return nil
}
