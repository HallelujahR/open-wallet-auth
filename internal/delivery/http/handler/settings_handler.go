package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	settingsusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/settings"
)

// SettingsHandler exposes runtime provider configuration APIs.
// SettingsHandler 暴露运行期服务商配置管理接口。
type SettingsHandler struct {
	settings *settingsusecase.Service
}

// NewSettingsHandler creates a settings management handler.
// NewSettingsHandler 创建系统配置管理 HTTP handler。
func NewSettingsHandler(settings *settingsusecase.Service) *SettingsHandler {
	return &SettingsHandler{settings: settings}
}

// Get returns redacted runtime provider settings.
// Get 返回脱敏后的运行期服务商配置。
func (h *SettingsHandler) Get(c *gin.Context) {
	result, err := h.settings.Public(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "SETTINGS_READ_FAILED", "read settings failed")
		return
	}
	response.OK(c, result)
}

// Update persists runtime provider settings from the admin console.
// Update 保存管理后台提交的运行期服务商配置。
func (h *SettingsHandler) Update(c *gin.Context) {
	var req settingsusecase.Snapshot
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "SETTINGS_INVALID_INPUT", "invalid settings payload")
		return
	}
	result, err := h.settings.Update(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "SETTINGS_UPDATE_FAILED", "update settings failed")
		return
	}
	response.OK(c, result)
}
