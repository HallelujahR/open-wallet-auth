package handler

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/dto"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
)

// ManagementHandler exposes management-console authentication endpoints.
// ManagementHandler 暴露认证管理后台自身的登录接口。
type ManagementHandler struct {
	cfg config.ManagementConfig
}

// NewManagementHandler creates a management authentication handler.
// NewManagementHandler 创建管理后台认证 HTTP handler。
func NewManagementHandler(cfg config.ManagementConfig) *ManagementHandler {
	return &ManagementHandler{cfg: cfg}
}

// Login verifies configured admin credentials and returns a management token.
// Login 校验配置中的后台管理员账号密码，并返回管理接口调用凭证。
func (h *ManagementHandler) Login(c *gin.Context) {
	var req dto.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "MANAGEMENT_INVALID_LOGIN", "invalid request")
		return
	}
	if h.cfg.AdminToken == "" {
		response.Error(c, http.StatusForbidden, "MANAGEMENT_DISABLED", "management API is disabled")
		return
	}
	if !constantTimeEqual(req.Username, h.cfg.AdminUsername) || !constantTimeEqual(req.Password, h.cfg.AdminPassword) {
		response.Error(c, http.StatusUnauthorized, "MANAGEMENT_UNAUTHORIZED", "invalid username or password")
		return
	}
	response.OK(c, dto.AdminLoginResponse{
		TokenType:  "Management",
		AdminToken: h.cfg.AdminToken,
	})
}

// constantTimeEqual compares small credential strings without early exit.
// constantTimeEqual 使用固定时间比较小型凭证字符串，避免明显的提前返回差异。
func constantTimeEqual(got string, want string) bool {
	if got == "" || want == "" {
		return false
	}
	if len(got) != len(want) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}
