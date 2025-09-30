package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"
)

// SMTPTLSender is a secure SMTP sender for production
type SMTPTLSender struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	AppName  string
	Timeout  time.Duration // optional timeout for sending emails
}

// sendMailTLS sends an email securely via TLS (supports plain text and HTML)
func (s *SMTPTLSender) sendMailTLS(ctx context.Context, to, subject, body, mimeType string) error {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	// Dial with context for cancellation/timeout
	dialer := &net.Dialer{}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
		ServerName: s.Host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Wrap the connection in a context-aware client
	done := make(chan error, 1)
	go func() {
		client, err := smtp.NewClient(conn, s.Host)
		if err != nil {
			done <- fmt.Errorf("failed to create SMTP client: %w", err)
			return
		}
		defer client.Quit()

		auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)
		if err := client.Auth(auth); err != nil {
			done <- fmt.Errorf("SMTP auth failed: %w", err)
			return
		}

		if err := client.Mail(s.From); err != nil {
			done <- fmt.Errorf("MAIL FROM failed: %w", err)
			return
		}
		if err := client.Rcpt(to); err != nil {
			done <- fmt.Errorf("RCPT TO failed: %w", err)
			return
		}

		w, err := client.Data()
		if err != nil {
			done <- fmt.Errorf("failed to get SMTP writer: %w", err)
			return
		}

		msg := fmt.Sprintf(
			"From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: %s; charset=\"UTF-8\"\r\n\r\n%s",
			s.AppName, s.From, to, subject, mimeType, body,
		)
		if _, err := w.Write([]byte(msg)); err != nil {
			done <- fmt.Errorf("failed to write email body: %w", err)
			return
		}
		if err := w.Close(); err != nil {
			done <- fmt.Errorf("failed to finalize email: %w", err)
			return
		}
		done <- nil
	}()

	select {
	case <-ctx.Done():
		return ctx.Err() // respect cancellation
	case err := <-done:
		return err
	}
}

// SendVerificationEmail sends an HTML/Plain email for verification
func (s *SMTPTLSender) SendVerificationEmail(ctx context.Context, to, link string, expiresAt string) error {
	subject := fmt.Sprintf("[%s] Verify Your Email", s.AppName)
	body := fmt.Sprintf("Hello,<br><br>Please verify your email by clicking the link below:<br><a href=\"%s\">Verify Email</a><br><br>This link expires at %s.<br><br>Thanks,<br>%s Team",
		link, expiresAt, s.AppName)
	return s.sendMailTLS(ctx, to, subject, body, "text/html")
}

// SendResetPasswordEmail sends an HTML/Plain email for password reset
func (s *SMTPTLSender) SendResetPasswordEmail(ctx context.Context, to, link string, expiresAt string) error {
	subject := fmt.Sprintf("[%s] Reset Your Password", s.AppName)
	body := fmt.Sprintf("Hello,<br><br>You requested to reset your password. Click the link below:<br><a href=\"%s\">Reset Password</a><br><br>This link expires at %s.<br>If you did not request this, please ignore.<br><br>Thanks,<br>%s Team",
		link, expiresAt, s.AppName)
	return s.sendMailTLS(ctx, to, subject, body, "text/html")
}

// Optional: Send email asynchronously (non-blocking)
func (s *SMTPTLSender) SendAsync(ctx context.Context, to, subject, body, mimeType string) {
	go func() {
		if err := s.sendMailTLS(ctx, to, subject, body, mimeType); err != nil {
			fmt.Printf("Failed to send email to %s: %v\n", to, err)
		}
	}()
}
