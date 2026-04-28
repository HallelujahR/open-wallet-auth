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

type RefreshToken struct {
	ID        string
	UserID    string
	ClientID  string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (t RefreshToken) IsExpired(now time.Time) bool {
	return !t.ExpiresAt.After(now)
}

func (t RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}
