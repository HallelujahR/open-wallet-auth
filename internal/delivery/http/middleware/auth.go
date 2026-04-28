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
