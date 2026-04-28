package token

import "time"

type Claims struct {
	UserID      string
	ClientID    string
	Audience    string
	Username    string
	Email       string
	Roles       []string
	Permissions []string
	Wallets     []string
	Issuer      string
	ExpiresAt   time.Time
	IssuedAt    time.Time
}

type Pair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}
