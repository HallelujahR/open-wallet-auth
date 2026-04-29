package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
)

const AdminTokenHeader = "X-Admin-Token"

// RequireAdminToken protects management endpoints with a configured shared token.
// RequireAdminToken 使用配置的管理 token 保护管理类接口。
func RequireAdminToken(adminToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if adminToken == "" {
			response.Error(c, http.StatusForbidden, "MANAGEMENT_DISABLED", "management API is disabled")
			c.Abort()
			return
		}
		if c.GetHeader(AdminTokenHeader) != adminToken {
			response.Error(c, http.StatusUnauthorized, "MANAGEMENT_UNAUTHORIZED", "invalid management token")
			c.Abort()
			return
		}
		c.Next()
	}
}
