package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
)

// JWKSProvider exposes public signing keys.
// JWKSProvider 暴露公开签名密钥集合。
type JWKSProvider interface {
	JWKS() token.JWKS
}

// JWKSHandler serves the public JWKS endpoint.
// JWKSHandler 提供公开 JWKS 接口，供业务系统校验 JWT。
type JWKSHandler struct {
	provider JWKSProvider
}

// NewJWKSHandler creates a JWKS handler.
// NewJWKSHandler 创建 JWKS HTTP handler。
func NewJWKSHandler(provider JWKSProvider) *JWKSHandler {
	return &JWKSHandler{provider: provider}
}

// JWKS writes the public key set.
// JWKS 输出当前服务的公开签名密钥集合。
func (h *JWKSHandler) JWKS(c *gin.Context) {
	c.JSON(200, h.provider.JWKS())
}
