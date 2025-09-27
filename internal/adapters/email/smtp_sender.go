package email

import (
	"context"
	"fmt"
	"net/smtp"
)

type SMTPSender struct {
	addr    string
	auth    smtp.Auth
	from    string
	appName string
}

func NewSMTPSender(host string, port int, username, password, from, appName string) *SMTPSender {
	addr := fmt.Sprintf("%s:%d", host, port)
	auth := smtp.PlainAuth("", username, password, host)
	return &SMTPSender{addr: addr, auth: auth, from: from, appName: appName}
}

func (s *SMTPSender) SendVerificationEmail(ctx context.Context, to string, link string, expiresAt string) error {
	subject := fmt.Sprintf("[%s] Verify Your Email", s.appName)
	body := fmt.Sprintf(
		"Hello,\n\nPlease verify your email by clicking the link below:\n\n%s\n\nThis link expires at %s.\n\nThanks,\n%s Team",
		link, expiresAt, s.appName,
	)

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body))
	return smtp.SendMail(s.addr, s.auth, s.from, []string{to}, msg)
}
