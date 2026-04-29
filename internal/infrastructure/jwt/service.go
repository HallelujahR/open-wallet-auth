package jwt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
)

// Service signs, verifies, and exposes JWT/JWKS data.
// Service 负责 JWT 签发、校验和 JWKS 暴露，是 token 端口的基础设施实现。
type Service struct {
	cfg        config.JWTConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// claims is the JWT-specific representation of normalized domain claims.
// claims 是领域 claims 在 JWT 中的具体承载结构。
type claims struct {
	Username    string   `json:"username,omitempty"`
	Email       string   `json:"email,omitempty"`
	ClientID    string   `json:"client_id,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Wallets     []string `json:"wallets,omitempty"`
	gojwt.RegisteredClaims
}

// NewService creates a JWT service from configured or generated RSA keys.
// NewService 使用配置文件中的 RSA 私钥或临时生成密钥创建 JWT 服务。
func NewService(cfg config.JWTConfig) (*Service, error) {
	privateKey, err := loadOrGenerateKey(cfg)
	if err != nil {
		return nil, err
	}
	return &Service{cfg: cfg, privateKey: privateKey, publicKey: &privateKey.PublicKey}, nil
}

// IssuePair returns a signed access token and an opaque refresh token.
// IssuePair 签发访问令牌，并生成不透明刷新令牌。
func (s *Service) IssuePair(ctx context.Context, input token.Claims) (*token.Pair, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.cfg.AccessTokenTTL)

	access, err := s.sign(claims{
		Username:    input.Username,
		Email:       input.Email,
		ClientID:    input.ClientID,
		Roles:       input.Roles,
		Permissions: input.Permissions,
		Wallets:     input.Wallets,
		RegisteredClaims: gojwt.RegisteredClaims{
			Issuer:    s.cfg.Issuer,
			Subject:   input.UserID,
			Audience:  gojwt.ClaimStrings{input.Audience},
			ExpiresAt: gojwt.NewNumericDate(expiresAt),
			IssuedAt:  gojwt.NewNumericDate(now),
			NotBefore: gojwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	})
	if err != nil {
		return nil, err
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	return &token.Pair{
		AccessToken:  access,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// IssueAccessToken signs a single access token without creating a refresh token.
// IssueAccessToken 只签发访问令牌，不生成刷新令牌。
func (s *Service) IssueAccessToken(ctx context.Context, input token.Claims) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.cfg.AccessTokenTTL)
	access, err := s.sign(claims{
		Username:    input.Username,
		Email:       input.Email,
		ClientID:    input.ClientID,
		Roles:       input.Roles,
		Permissions: input.Permissions,
		Wallets:     input.Wallets,
		RegisteredClaims: gojwt.RegisteredClaims{
			Issuer:    s.cfg.Issuer,
			Subject:   input.UserID,
			Audience:  gojwt.ClaimStrings{input.Audience},
			ExpiresAt: gojwt.NewNumericDate(expiresAt),
			IssuedAt:  gojwt.NewNumericDate(now),
			NotBefore: gojwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	})
	return access, expiresAt, err
}

// RefreshTokenTTL returns the configured refresh token lifetime.
// RefreshTokenTTL 返回配置中的刷新令牌有效期。
func (s *Service) RefreshTokenTTL() time.Duration {
	return s.cfg.RefreshTokenTTL
}

// GenerateRefreshToken creates a cryptographically random opaque refresh token.
// GenerateRefreshToken 创建密码学安全的不透明刷新令牌。
func GenerateRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "rfr_" + base64.RawURLEncoding.EncodeToString(buf), nil
}

// sign signs JWT claims with the active RSA private key and key id.
// sign 使用当前 RSA 私钥和 kid 签名 JWT claims。
func (s *Service) sign(c claims) (string, error) {
	t := gojwt.NewWithClaims(gojwt.SigningMethodRS256, c)
	t.Header["kid"] = s.cfg.ActiveKeyID
	return t.SignedString(s.privateKey)
}

// Verify validates a JWT and returns normalized claims.
// Verify 校验 JWT 签名、issuer、audience，并返回归一化 claims。
func (s *Service) Verify(ctx context.Context, tokenString string, audience string) (*token.Claims, error) {
	parsed, err := gojwt.ParseWithClaims(tokenString, &claims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected jwt signing method")
		}
		return s.publicKey, nil
	}, gojwt.WithIssuer(s.cfg.Issuer), gojwt.WithAudience(audience))
	if err != nil {
		return nil, err
	}

	c, ok := parsed.Claims.(*claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid jwt claims")
	}

	return &token.Claims{
		UserID:      c.Subject,
		ClientID:    c.ClientID,
		Audience:    firstAudience(c.Audience),
		Username:    c.Username,
		Email:       c.Email,
		Roles:       c.Roles,
		Permissions: c.Permissions,
		Wallets:     c.Wallets,
		Issuer:      c.Issuer,
		ExpiresAt:   c.ExpiresAt.Time,
		IssuedAt:    c.IssuedAt.Time,
	}, nil
}

// JWKS returns the public signing keys in JSON Web Key Set format.
// JWKS 以 JSON Web Key Set 格式返回公开签名密钥。
func (s *Service) JWKS() token.JWKS {
	return token.JWKS{
		Keys: []token.JWK{{
			Kty: "RSA",
			Use: "sig",
			Kid: s.cfg.ActiveKeyID,
			Alg: "RS256",
			N:   base64.RawURLEncoding.EncodeToString(s.publicKey.N.Bytes()),
			E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(s.publicKey.E)).Bytes()),
		}},
	}
}

// firstAudience returns the first audience claim used by this service.
// firstAudience 返回本服务当前使用的第一个 audience 声明。
func firstAudience(aud gojwt.ClaimStrings) string {
	if len(aud) == 0 {
		return ""
	}
	return aud[0]
}

// loadOrGenerateKey loads a configured RSA private key or creates an ephemeral one.
// loadOrGenerateKey 加载配置的 RSA 私钥；本地开发缺失时生成临时密钥。
func loadOrGenerateKey(cfg config.JWTConfig) (*rsa.PrivateKey, error) {
	if cfg.PrivateKeyPath == "" {
		return rsa.GenerateKey(rand.Reader, 2048)
	}

	raw, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return rsa.GenerateKey(rand.Reader, 2048)
		}
		return nil, err
	}

	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("invalid jwt private key pem")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("jwt private key is not rsa")
	}
	return rsaKey, nil
}
