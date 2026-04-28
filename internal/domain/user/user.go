package user

import "time"

type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusDeleted   Status = "deleted"
)

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
