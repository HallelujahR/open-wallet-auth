package token

import "time"

// Claims represents normalized authentication claims used inside the service.
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

// Pair contains an access token and refresh token returned to clients.
type Pair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// RefreshToken represents a persisted refresh token record.
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

// JWKS is the JSON Web Key Set returned from the public discovery endpoint.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK describes a public signing key in JSON Web Key format.
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}
