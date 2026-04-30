package message

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	emailusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/email"
)

// SMTPConfig configures an SMTP email provider.
// SMTPConfig 配置 SMTP 邮件服务商。
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// SMTPProvider sends verification emails through an SMTP server.
// SMTPProvider 通过 SMTP 服务器发送邮箱验证码。
type SMTPProvider struct {
	cfg SMTPConfig
}

// NewSMTPProvider creates an SMTP email provider.
// NewSMTPProvider 创建 SMTP 邮件发送适配器。
func NewSMTPProvider(cfg SMTPConfig) *SMTPProvider {
	return &SMTPProvider{cfg: cfg}
}

// SendEmail sends a plain-text verification-code email.
// SendEmail 发送纯文本邮箱验证码邮件。
func (p *SMTPProvider) SendEmail(ctx context.Context, msg emailusecase.EmailMessage) error {
	if err := p.validate(); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	from := p.cfg.From
	if from == "" {
		from = p.cfg.Username
	}
	addr := fmt.Sprintf("%s:%d", p.cfg.Host, p.cfg.Port)
	auth := smtp.PlainAuth("", p.cfg.Username, p.cfg.Password, p.cfg.Host)
	raw := buildEmail(from, msg.Email, msg.Subject, fmt.Sprintf("Your verification code is: %s\n", msg.Code))
	return smtp.SendMail(addr, auth, from, []string{msg.Email}, []byte(raw))
}

// validate checks required SMTP settings before sending.
// validate 在发送前检查必要的 SMTP 配置。
func (p *SMTPProvider) validate() error {
	if strings.TrimSpace(p.cfg.Host) == "" || p.cfg.Port <= 0 {
		return errors.New("smtp host and port are required")
	}
	if strings.TrimSpace(p.cfg.Username) == "" || strings.TrimSpace(p.cfg.Password) == "" {
		return errors.New("smtp username and password are required")
	}
	return nil
}

// buildEmail creates a minimal RFC 5322 style text email.
// buildEmail 构造最小可用的 RFC 5322 文本邮件。
func buildEmail(from string, to string, subject string, body string) string {
	return strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")
}
