package client

import "time"

type Status string

const (
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

type Client struct {
	ID                  string
	ClientID            string
	Name                string
	JWTAudience         string
	AllowedOrigins      []string
	AllowedRedirectURIs []string
	Status              Status
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (c Client) IsActive() bool {
	return c.Status == StatusActive
}
