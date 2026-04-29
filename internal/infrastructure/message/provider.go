package message

import "context"

// SMSMessage describes a phone verification message.
type SMSMessage struct {
	Phone string
	Code  string
}

// EmailMessage describes an email verification message.
type EmailMessage struct {
	Email   string
	Subject string
	Code    string
}

// SMSProvider sends phone verification messages.
type SMSProvider interface {
	SendSMS(ctx context.Context, msg SMSMessage) error
}

// EmailProvider sends email verification messages.
type EmailProvider interface {
	SendEmail(ctx context.Context, msg EmailMessage) error
}

// NoopProvider accepts messages without sending them.
type NoopProvider struct{}

func (NoopProvider) SendSMS(ctx context.Context, msg SMSMessage) error {
	return nil
}

func (NoopProvider) SendEmail(ctx context.Context, msg EmailMessage) error {
	return nil
}
