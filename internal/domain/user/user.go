package user

import "time"

// Status is the lifecycle state of a user account.
// Status 表示用户账号的生命周期状态。
type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusDeleted   Status = "deleted"
)

// User is the core account entity owned by the auth service.
// User 是认证服务拥有的核心账号实体。
type User struct {
	ID           string
	Username     string
	Email        string
	Phone        string
	PasswordHash string
	Avatar       string
	Status       Status
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// IsActive reports whether the user can authenticate and receive tokens.
// IsActive 判断用户是否可以登录并获得 token。
func (u User) IsActive() bool {
	return u.Status == StatusActive
}
