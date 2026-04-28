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
