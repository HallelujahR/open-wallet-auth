package message

import (
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
)

// NewProvider creates a message provider from runtime configuration.
func NewProvider(cfg config.MessageProviderConfig) (SMSProvider, EmailProvider) {
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
