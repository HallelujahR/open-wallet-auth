package user

import "time"

// Status is the lifecycle state of a user account.
type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusDeleted   Status = "deleted"
)

// User is the core account entity owned by the auth service.
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

func (u User) IsActive() bool {
	return u.Status == StatusActive
}
