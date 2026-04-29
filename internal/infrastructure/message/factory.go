package message

import (
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
	emailusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/email"
	phoneusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/phone"
)

// NewProvider creates a message provider from runtime configuration.
// NewProvider 根据运行配置创建短信/邮件发送适配器。
func NewProvider(cfg config.MessageProviderConfig) (phoneusecase.SMSProvider, emailusecase.EmailProvider) {
	if cfg.Type == "webhook" {
		provider := NewWebhookProvider(WebhookConfig{
			URL:         cfg.WebhookURL,
			BearerToken: cfg.BearerToken,
			Headers:     cfg.Headers,
		})
		return provider, provider
	}
	provider := NoopProvider{}
	return provider, provider
}
