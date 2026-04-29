package token

import "time"

// Claims represents normalized authentication claims used inside the service.
// Claims 是服务内部统一使用的认证声明，屏蔽 JWT 具体实现细节。
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
// Pair 是返回给接入系统的访问令牌与刷新令牌组合。
type Pair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// RefreshToken represents a persisted refresh token record.
// RefreshToken 表示已持久化的刷新令牌记录，数据库中只保存哈希值。
type RefreshToken struct {
	ID        string
	UserID    string
	ClientID  string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

// IsExpired reports whether the refresh token is past its expiration time.
// IsExpired 判断刷新令牌是否已经过期。
func (t RefreshToken) IsExpired(now time.Time) bool {
	return !t.ExpiresAt.After(now)
}

// IsRevoked reports whether the refresh token has been explicitly revoked.
// IsRevoked 判断刷新令牌是否已经被主动吊销。
func (t RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

// JWKS is the JSON Web Key Set returned from the public discovery endpoint.
// JWKS 是公开发现端点返回的 JSON Web Key Set。
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK describes a public signing key in JSON Web Key format.
// JWK 用 JSON Web Key 格式描述一个公开签名密钥。
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}
