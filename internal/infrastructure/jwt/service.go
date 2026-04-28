package jwt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
)

type Service struct {
	cfg        config.JWTConfig
	privateKey *rsa.PrivateKey
}

type claims struct {
	Username    string   `json:"username,omitempty"`
	Email       string   `json:"email,omitempty"`
	ClientID    string   `json:"client_id,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Wallets     []string `json:"wallets,omitempty"`
	gojwt.RegisteredClaims
}

func NewService(cfg config.JWTConfig) (*Service, error) {
	privateKey, err := loadOrGenerateKey(cfg)
	if err != nil {
		return nil, err
	}
	return &Service{cfg: cfg, privateKey: privateKey}, nil
}

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

	refreshToken := "rfr_" + uuid.NewString() + uuid.NewString()
	return &token.Pair{
		AccessToken:  access,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *Service) sign(c claims) (string, error) {
	t := gojwt.NewWithClaims(gojwt.SigningMethodRS256, c)
	t.Header["kid"] = s.cfg.ActiveKeyID
	return t.SignedString(s.privateKey)
}

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
