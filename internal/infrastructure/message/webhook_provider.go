package message

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	emailusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/email"
	phoneusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/phone"
)

// WebhookConfig configures a generic HTTP message provider.
// WebhookConfig 用于把短信/邮件发送委托给用户自己的 HTTP 消息网关。
type WebhookConfig struct {
	URL         string
	BearerToken string
	Headers     map[string]string
}

// WebhookProvider sends SMS and email messages to a user-defined webhook.
// WebhookProvider 是通用服务商 Demo，可在外部网关里再接阿里云、腾讯云、SendGrid 等。
type WebhookProvider struct {
	cfg        WebhookConfig
	httpClient *http.Client
}

// NewWebhookProvider creates a webhook-backed message provider.
// NewWebhookProvider 根据配置创建 webhook 发送适配器。
func NewWebhookProvider(cfg WebhookConfig) *WebhookProvider {
	return &WebhookProvider{cfg: cfg, httpClient: &http.Client{Timeout: 10 * time.Second}}
}

// SendSMS forwards an SMS verification message to the configured webhook.
// SendSMS 将短信验证码消息转发到配置的 webhook。
func (p *WebhookProvider) SendSMS(ctx context.Context, msg phoneusecase.SMSMessage) error {
	return p.send(ctx, map[string]any{"type": "sms", "phone": msg.Phone, "code": msg.Code})
}

// SendEmail forwards an email verification message to the configured webhook.
// SendEmail 将邮箱验证码消息转发到配置的 webhook。
func (p *WebhookProvider) SendEmail(ctx context.Context, msg emailusecase.EmailMessage) error {
	return p.send(ctx, map[string]any{"type": "email", "email": msg.Email, "subject": msg.Subject, "code": msg.Code})
}

// send posts a provider-neutral JSON payload to the external message gateway.
// send 将通用 JSON 消息发送到外部消息网关。
func (p *WebhookProvider) send(ctx context.Context, body any) error {
	if p.cfg.URL == "" {
		return errors.New("message webhook url is required")
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.URL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.cfg.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.cfg.BearerToken)
	}
	for key, value := range p.cfg.Headers {
		req.Header.Set(key, value)
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("message webhook returned non-2xx status")
	}
	return nil
}
