package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
)

type JWKSProvider interface {
	JWKS() token.JWKS
}

type JWKSHandler struct {
	provider JWKSProvider
}

func NewJWKSHandler(provider JWKSProvider) *JWKSHandler {
	return &JWKSHandler{provider: provider}
}

func (h *JWKSHandler) JWKS(c *gin.Context) {
	c.JSON(200, h.provider.JWKS())
}
