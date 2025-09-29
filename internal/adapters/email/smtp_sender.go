package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
)

type SMTPTLSender struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	AppName  string
}

// sendMailTLS sends an email via Gmail using TLS (port 465)
func (s *SMTPTLSender) sendMailTLS(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         s.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return err
	}
	defer client.Quit()

	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)
	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(s.From); err != nil {
		return err
	}

	if err = client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body))
	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	if err = w.Close(); err != nil {
		return err
	}

	return nil
}

func (s *SMTPTLSender) SendVerificationEmail(ctx context.Context, to, link, expiresAt string) error {
	subject := fmt.Sprintf("[%s] Verify Your Email", s.AppName)
	body := fmt.Sprintf("Hello,\n\nPlease verify your email by clicking the link below:\n\n%s\n\nThis link expires at %s.\n\nThanks,\n%s Team",
		link, expiresAt, s.AppName)
	return s.sendMailTLS(to, subject, body)
}

func (s *SMTPTLSender) SendResetPasswordEmail(ctx context.Context, to, link, expiresAt string) error {
	subject := fmt.Sprintf("[%s] Reset Your Password", s.AppName)
	body := fmt.Sprintf("Hello,\n\nYou requested to reset your password. Click the link below:\n\n%s\n\nThis link expires at %s.\n\nIf you did not request this, please ignore.\n\nThanks,\n%s Team",
		link, expiresAt, s.AppName)
	return s.sendMailTLS(to, subject, body)
}