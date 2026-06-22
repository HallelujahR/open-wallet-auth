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
	WhitelistEnabled    bool
	Status              Status
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// IsActive reports whether the client is allowed to issue or verify tokens.
// IsActive 判断该业务系统是否仍允许签发或校验 token。
func (c Client) IsActive() bool {
	return c.Status == StatusActive
}

// MemberStatus is the lifecycle state of a user's access to one application.
// MemberStatus 表示用户对某个业务系统的准入授权状态。
type MemberStatus string

const (
	MemberStatusActive   MemberStatus = "active"
	MemberStatusDisabled MemberStatus = "disabled"
)

// Member represents an allow-list entry for one application client.
// Member 表示某个统一身份用户对一个业务系统的访问授权。
type Member struct {
	ID          string
	ClientID    string
	UserID      string
	Role        string
	Permissions []string
	Status      MemberStatus
	Remark      string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Display fields are populated for management APIs only.
	// 展示字段仅供管理接口返回，不参与授权判断。
	Username string
	Email    string
	Phone    string
}

// IsActive reports whether this membership can grant access.
// IsActive 判断该成员授权是否允许进入对应业务系统。
func (m Member) IsActive() bool {
	return m.Status == MemberStatusActive
}
