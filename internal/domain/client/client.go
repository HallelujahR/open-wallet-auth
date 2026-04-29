package client

import "time"

// Status is the lifecycle state of an application client.
// Status 表示接入应用 client 的生命周期状态。
type Status string

const (
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

// Client represents an application allowed to request tokens.
// Client 表示一个允许向认证服务申请 token 的业务系统。
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

// IsActive reports whether the client is allowed to issue or verify tokens.
// IsActive 判断该业务系统是否仍允许签发或校验 token。
func (c Client) IsActive() bool {
	return c.Status == StatusActive
}
