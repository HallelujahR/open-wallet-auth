package message

import (
	"context"
	"sync"

	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
	emailusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/email"
	phoneusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/phone"
	settingsusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/settings"
)

// SettingsProvider supplies the latest runtime provider settings.
// SettingsProvider 提供最新的运行期服务商配置。
type SettingsProvider interface {
	Get(ctx context.Context) (settingsusecase.Snapshot, error)
}

// DynamicSMSProvider resolves the SMS provider from editable settings on each send.
// DynamicSMSProvider 每次发送短信前读取最新可视化配置。
type DynamicSMSProvider struct {
	settings SettingsProvider
	fallback config.MessageProviderConfig
	mu       sync.Mutex
}

// NewDynamicSMSProvider creates a dynamic SMS provider.
// NewDynamicSMSProvider 创建动态短信服务商适配器。
func NewDynamicSMSProvider(settings SettingsProvider, fallback config.MessageProviderConfig) *DynamicSMSProvider {
	return &DynamicSMSProvider{settings: settings, fallback: fallback}
}

// SendSMS sends an SMS through the currently configured provider.
// SendSMS 使用当前配置的短信服务商发送验证码。
func (p *DynamicSMSProvider) SendSMS(ctx context.Context, msg phoneusecase.SMSMessage) error {
	cfg := p.fallback
	if p.settings != nil {
		if snapshot, err := p.settings.Get(ctx); err == nil {
			cfg = messageConfig(snapshot.Phone.Provider)
		}
	}
	sms, _ := NewProvider(cfg)
	p.mu.Lock()
	defer p.mu.Unlock()
	return sms.SendSMS(ctx, msg)
}

// DynamicEmailProvider resolves the email provider from editable settings on each send.
// DynamicEmailProvider 每次发送邮件前读取最新可视化配置。
type DynamicEmailProvider struct {
	settings SettingsProvider
	fallback config.MessageProviderConfig
	mu       sync.Mutex
}

// NewDynamicEmailProvider creates a dynamic email provider.
// NewDynamicEmailProvider 创建动态邮件服务商适配器。
func NewDynamicEmailProvider(settings SettingsProvider, fallback config.MessageProviderConfig) *DynamicEmailProvider {
	return &DynamicEmailProvider{settings: settings, fallback: fallback}
}

// SendEmail sends an email through the currently configured provider.
// SendEmail 使用当前配置的邮件服务商发送验证码。
func (p *DynamicEmailProvider) SendEmail(ctx context.Context, msg emailusecase.EmailMessage) error {
	cfg := p.fallback
	if p.settings != nil {
		if snapshot, err := p.settings.Get(ctx); err == nil {
			cfg = messageConfig(snapshot.Email.Provider)
		}
	}
	_, email := NewProvider(cfg)
	p.mu.Lock()
	defer p.mu.Unlock()
	return email.SendEmail(ctx, msg)
}

// messageConfig converts management settings into infrastructure config.
// messageConfig 将管理配置转换为 infrastructure provider 配置。
func messageConfig(input settingsusecase.MessageProviderSettings) config.MessageProviderConfig {
	return config.MessageProviderConfig{
		Type: input.Type,
		Webhook: config.WebhookConfig{
			URL:         input.Webhook.URL,
			BearerToken: input.Webhook.BearerToken,
		},
		SMTP: config.SMTPConfig{
			Host:     input.SMTP.Host,
			Port:     input.SMTP.Port,
			Username: input.SMTP.Username,
			Password: input.SMTP.Password,
			From:     input.SMTP.From,
		},
		AliyunSMS: config.AliyunSMSConfig{
			AccessKeyID:     input.AliyunSMS.AccessKeyID,
			AccessKeySecret: input.AliyunSMS.AccessKeySecret,
			SignName:        input.AliyunSMS.SignName,
			TemplateCode:    input.AliyunSMS.TemplateCode,
			RegionID:        input.AliyunSMS.RegionID,
			Endpoint:        input.AliyunSMS.Endpoint,
		},
		Headers: input.Headers,
	}
}
