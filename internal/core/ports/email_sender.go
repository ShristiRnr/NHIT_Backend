package ports

import "context"

type EmailSender interface {
	SendVerificationEmail(ctx context.Context, to string, link string, expiresAt string) error
}
