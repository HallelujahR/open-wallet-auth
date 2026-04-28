package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
)

func TestServiceIssueVerifyAndJWKS(t *testing.T) {
	service, err := NewService(config.JWTConfig{
		Issuer:          "open-wallet-auth-test",
		AccessTokenTTL:  time.Minute,
		RefreshTokenTTL: time.Hour,
		ActiveKeyID:     "test-key",
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	pair, err := service.IssuePair(context.Background(), token.Claims{
		UserID:   "usr_1",
		ClientID: "default",
		Audience: "default",
		Username: "alice",
		Email:    "alice@example.com",
	})
	if err != nil {
		t.Fatalf("issue pair: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("expected token pair")
	}
	if service.RefreshTokenTTL() != time.Hour {
		t.Fatal("expected refresh token ttl")
	}

	claims, err := service.Verify(context.Background(), pair.AccessToken, "default")
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if claims.UserID != "usr_1" || claims.Email != "alice@example.com" {
		t.Fatalf("unexpected claims: %#v", claims)
	}

	jwks := service.JWKS()
	if len(jwks.Keys) != 1 {
		t.Fatalf("expected one jwk, got %d", len(jwks.Keys))
	}
	if jwks.Keys[0].Kid != "test-key" || jwks.Keys[0].Alg != "RS256" {
		t.Fatalf("unexpected jwk: %#v", jwks.Keys[0])
	}
}

func TestServiceVerifyRejectsWrongAudience(t *testing.T) {
	service, err := NewService(config.JWTConfig{
		Issuer:         "open-wallet-auth-test",
		AccessTokenTTL: time.Minute,
		ActiveKeyID:    "test-key",
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	pair, err := service.IssuePair(context.Background(), token.Claims{
		UserID:   "usr_1",
		ClientID: "default",
		Audience: "default",
	})
	if err != nil {
		t.Fatalf("issue pair: %v", err)
	}

	if _, err := service.Verify(context.Background(), pair.AccessToken, "blockx"); err == nil {
		t.Fatal("expected wrong audience to be rejected")
	}
}
