package audit

import "time"

// LoginMethod describes how a user authenticated.
type LoginMethod string

const (
	LoginMethodPassword LoginMethod = "password"
	LoginMethodRefresh  LoginMethod = "refresh"
	LoginMethodWallet   LoginMethod = "wallet"
	LoginMethodPhone    LoginMethod = "phone"
	LoginMethodOAuth    LoginMethod = "oauth"
)

// SecurityEventType identifies account-sensitive operations.
// SecurityEventType 标识账号安全相关操作类型。
type SecurityEventType string

const (
	SecurityEventChangePassword   SecurityEventType = "change_password"
	SecurityEventResetPassword    SecurityEventType = "reset_password"
	SecurityEventAdminSetPassword SecurityEventType = "admin_set_password"
	SecurityEventBindEmail        SecurityEventType = "bind_email"
	SecurityEventBindPhone        SecurityEventType = "bind_phone"
	SecurityEventUnbindEmail      SecurityEventType = "unbind_email"
	SecurityEventUnbindPhone      SecurityEventType = "unbind_phone"
	SecurityEventUnbindWallet     SecurityEventType = "unbind_wallet"
	SecurityEventUnbindOAuth      SecurityEventType = "unbind_oauth"
)

// LoginLog records one authentication attempt.
type LoginLog struct {
	ID          string
	UserID      string
	ClientID    string
	LoginMethod LoginMethod
	IP          string
	UserAgent   string
	Success     bool
	FailureCode string
	CreatedAt   time.Time
}

// SecurityEvent records a sensitive account operation for later audit.
// SecurityEvent 记录敏感账号操作，便于后续安全审计。
type SecurityEvent struct {
	ID          string
	UserID      string
	EventType   SecurityEventType
	TargetType  string
	TargetID    string
	IP          string
	UserAgent   string
	Success     bool
	FailureCode string
	CreatedAt   time.Time
}

// UserClient records a user's relationship with an application client.
type UserClient struct {
	ID           string
	UserID       string
	ClientID     string
	FirstLoginAt time.Time
	LastLoginAt  time.Time
	LoginCount   int64
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
