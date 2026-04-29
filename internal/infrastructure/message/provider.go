package message

import (
	"context"

	emailusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/email"
	phoneusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/phone"
)

// NoopProvider accepts messages without sending them.
// NoopProvider 本地开发时吞掉消息，不调用真实短信或邮件服务商。
type NoopProvider struct{}

// SendSMS accepts an SMS message and intentionally performs no external call.
// SendSMS 接收短信消息但不调用外部服务，适合本地调试。
func (NoopProvider) SendSMS(ctx context.Context, msg phoneusecase.SMSMessage) error {
	return nil
}

// SendEmail accepts an email message and intentionally performs no external call.
// SendEmail 接收邮件消息但不调用外部服务，适合本地调试。
func (NoopProvider) SendEmail(ctx context.Context, msg emailusecase.EmailMessage) error {
	return nil
}
