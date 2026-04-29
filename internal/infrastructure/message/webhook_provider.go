package message

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// WebhookConfig configures a generic HTTP message provider.
type WebhookConfig struct {
	URL         string
	BearerToken string
	Headers     map[string]string
}

// WebhookProvider sends SMS and email messages to a user-defined webhook.
type WebhookProvider struct {
	cfg        WebhookConfig
	httpClient *http.Client
}

// NewWebhookProvider creates a webhook-backed message provider.
func NewWebhookProvider(cfg WebhookConfig) *WebhookProvider {
	return &WebhookProvider{cfg: cfg, httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func (p *WebhookProvider) SendSMS(ctx context.Context, msg SMSMessage) error {
	return p.send(ctx, map[string]any{"type": "sms", "phone": msg.Phone, "code": msg.Code})
}

func (p *WebhookProvider) SendEmail(ctx context.Context, msg EmailMessage) error {
	return p.send(ctx, map[string]any{"type": "email", "email": msg.Email, "subject": msg.Subject, "code": msg.Code})
}

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
