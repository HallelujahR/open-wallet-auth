package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
)

// JWKSProvider exposes public signing keys.
type JWKSProvider interface {
	JWKS() token.JWKS
}

// JWKSHandler serves the public JWKS endpoint.
type JWKSHandler struct {
	provider JWKSProvider
}

// NewJWKSHandler creates a JWKS handler.
func NewJWKSHandler(provider JWKSProvider) *JWKSHandler {
	return &JWKSHandler{provider: provider}
}

// JWKS writes the public key set.
func (h *JWKSHandler) JWKS(c *gin.Context) {
	c.JSON(200, h.provider.JWKS())
}
