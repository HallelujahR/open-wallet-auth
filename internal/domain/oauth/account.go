package oauth

import "time"

// Account links a third-party OAuth identity to an auth-service user.
type Account struct {
	ID                string
	UserID            string
	Provider          string
	ProviderSubject   string
	ProviderEmail     string
	ProviderUsername  string
	ProviderAvatarURL string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
