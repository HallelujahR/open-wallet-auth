package audit

import "time"

// LoginMethod describes how a user authenticated.
type LoginMethod string

const (
	LoginMethodPassword LoginMethod = "password"
	LoginMethodRefresh  LoginMethod = "refresh"
	LoginMethodWallet   LoginMethod = "wallet"
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
