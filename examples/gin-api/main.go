package main

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const authClaimsKey = "auth_claims"

// AuthClaims is the normalized identity payload used by business handlers.
type AuthClaims struct {
	UserID      string
	ClientID    string
	Username    string
	Email       string
	Wallets     []string
	Roles       []string
	Permissions []string
}

type jwtClaims struct {
	Username    string   `json:"username,omitempty"`
	Email       string   `json:"email,omitempty"`
	ClientID    string   `json:"client_id,omitempty"`
	Wallets     []string `json:"wallets,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// JWKSVerifier fetches and caches Open Wallet Auth public signing keys.
type JWKSVerifier struct {
	jwksURL    string
	issuer     string
	audience   string
	httpClient *http.Client

	mu        sync.RWMutex
	keyByID   map[string]*rsa.PublicKey
	fetchedAt time.Time
	ttl       time.Duration
}

// NewJWKSVerifier creates a verifier for one business application audience.
func NewJWKSVerifier(jwksURL string, issuer string, audience string) *JWKSVerifier {
	return &JWKSVerifier{
		jwksURL:    jwksURL,
		issuer:     issuer,
		audience:   audience,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		keyByID:    map[string]*rsa.PublicKey{},
		ttl:        5 * time.Minute,
	}
}

// Verify checks token signature and standard claims, then returns business-friendly claims.
func (v *JWKSVerifier) Verify(ctx context.Context, rawToken string) (*AuthClaims, error) {
	claims := &jwtClaims{}
	parsed, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected jwt signing method")
		}
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("missing jwt key id")
		}
		return v.key(ctx, kid)
	}, jwt.WithIssuer(v.issuer), jwt.WithAudience(v.audience))
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("invalid jwt")
	}

	return &AuthClaims{
		UserID:      claims.Subject,
		ClientID:    claims.ClientID,
		Username:    claims.Username,
		Email:       claims.Email,
		Wallets:     claims.Wallets,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
	}, nil
}

func (v *JWKSVerifier) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	key := v.keyByID[kid]
	fresh := time.Since(v.fetchedAt) < v.ttl
	v.mu.RUnlock()
	if key != nil && fresh {
		return key, nil
	}

	if err := v.refresh(ctx); err != nil {
		return nil, err
	}

	v.mu.RLock()
	defer v.mu.RUnlock()
	key = v.keyByID[kid]
	if key == nil {
		return nil, errors.New("jwt key id not found")
	}
	return key, nil
}

func (v *JWKSVerifier) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return err
	}
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("jwks endpoint returned non-200 status")
	}

	var set struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			Alg string `json:"alg"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
		return err
	}

	keys := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, jwk := range set.Keys {
		if jwk.Kty != "RSA" || jwk.Alg != "RS256" || jwk.Kid == "" {
			continue
		}
		key, err := rsaPublicKey(jwk.N, jwk.E)
		if err != nil {
			return err
		}
		keys[jwk.Kid] = key
	}

	v.mu.Lock()
	v.keyByID = keys
	v.fetchedAt = time.Now()
	v.mu.Unlock()
	return nil
}

func rsaPublicKey(encodedN string, encodedE string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(encodedN)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(encodedE)
	if err != nil {
		return nil, err
	}
	e := new(big.Int).SetBytes(eBytes).Int64()
	if e <= 0 {
		return nil, errors.New("invalid rsa exponent")
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(e),
	}, nil
}

// JWTMiddleware protects business API routes with Open Wallet Auth JWTs.
func JWTMiddleware(verifier *JWKSVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		claims, err := verifier.Verify(c.Request.Context(), rawToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(authClaimsKey, claims)
		c.Next()
	}
}

func bearerToken(header string) (string, bool) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return strings.TrimSpace(parts[1]), true
}

// MustAuthClaims returns verified identity claims from the Gin context.
func MustAuthClaims(c *gin.Context) *AuthClaims {
	claims, ok := c.MustGet(authClaimsKey).(*AuthClaims)
	if !ok {
		panic("auth claims missing from context")
	}
	return claims
}

func main() {
	jwksURL := env("OWA_JWKS_URL", "http://localhost:8080/.well-known/jwks.json")
	issuer := env("OWA_ISSUER", "open-wallet-auth")
	audience := env("OWA_AUDIENCE", "default")
	addr := env("APP_ADDR", ":8090")

	verifier := NewJWKSVerifier(jwksURL, issuer, audience)
	router := gin.Default()

	router.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	protected := router.Group("", JWTMiddleware(verifier))
	protected.GET("/profile", func(c *gin.Context) {
		claims := MustAuthClaims(c)
		c.JSON(http.StatusOK, gin.H{
			"auth_user_id": claims.UserID,
			"client_id":    claims.ClientID,
			"username":     claims.Username,
			"email":        claims.Email,
			"wallets":      claims.Wallets,
		})
	})

	if err := router.Run(addr); err != nil {
		panic(err)
	}
}

func env(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
