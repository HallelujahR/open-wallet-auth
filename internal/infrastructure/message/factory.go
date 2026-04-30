package message

import (
	"strings"

	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
	emailusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/email"
	phoneusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/phone"
)

// NewProvider creates a message provider from runtime configuration.
// NewProvider 根据运行配置创建短信/邮件发送适配器。
func NewProvider(cfg config.MessageProviderConfig) (phoneusecase.SMSProvider, emailusecase.EmailProvider) {
	switch strings.ToLower(strings.TrimSpace(cfg.Type)) {
	case "webhook":
		provider := NewWebhookProvider(WebhookConfig{
			URL:         cfg.Webhook.URL,
			BearerToken: cfg.Webhook.BearerToken,
			Headers:     cfg.Headers,
		})
		return provider, provider
	case "smtp":
		emailProvider := NewSMTPProvider(SMTPConfig{
			Host:     cfg.SMTP.Host,
			Port:     cfg.SMTP.Port,
			Username: cfg.SMTP.Username,
			Password: cfg.SMTP.Password,
			From:     cfg.SMTP.From,
		})
		return NoopProvider{}, emailProvider
	case "aliyun_sms":
		smsProvider := NewAliyunSMSProvider(AliyunSMSConfig{
			AccessKeyID:     cfg.AliyunSMS.AccessKeyID,
			AccessKeySecret: cfg.AliyunSMS.AccessKeySecret,
			SignName:        cfg.AliyunSMS.SignName,
			TemplateCode:    cfg.AliyunSMS.TemplateCode,
			RegionID:        cfg.AliyunSMS.RegionID,
			Endpoint:        cfg.AliyunSMS.Endpoint,
		})
		return smsProvider, NoopProvider{}
	default:
		provider := NoopProvider{}
		return provider, provider
	}
}
