package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/contextkey"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
)

type TokenVerifier interface {
	Verify(ctx context.Context, tokenString string, audience string) (*token.Claims, error)
}

type ClientAudienceResolver interface {
	ResolveAudience(ctx context.Context, clientID string) (string, error)
}

// Authenticate validates a Bearer token and stores claims in the request context.
func Authenticate(verifier TokenVerifier, audience string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "AUTH_MISSING_TOKEN", "missing authorization token")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid authorization token")
			c.Abort()
			return
		}

		claims, err := verifier.Verify(c.Request.Context(), parts[1], audience)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid authorization token")
			c.Abort()
			return
		}

		c.Set(contextkey.AuthClaims, claims)
		c.Next()
	}
}

// AuthenticateClient validates a token against the audience of the requested client.
func AuthenticateClient(verifier TokenVerifier, resolver ClientAudienceResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.GetHeader("X-Client-ID")
		if clientID == "" {
			clientID = c.Query("client_id")
		}
		if clientID == "" {
			clientID = "default"
		}

		audience, err := resolver.ResolveAudience(c.Request.Context(), clientID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "CLIENT_INVALID", "invalid client")
			c.Abort()
			return
		}

		Authenticate(verifier, audience)(c)
	}
}
