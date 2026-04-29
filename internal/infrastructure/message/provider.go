package message

import (
	"context"

	emailusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/email"
	phoneusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/phone"
)

// NoopProvider accepts messages without sending them.
// NoopProvider 本地开发时吞掉消息，不调用真实短信或邮件服务商。
type NoopProvider struct{}

func (NoopProvider) SendSMS(ctx context.Context, msg phoneusecase.SMSMessage) error {
	return nil
}

func (NoopProvider) SendEmail(ctx context.Context, msg emailusecase.EmailMessage) error {
	return nil
}
