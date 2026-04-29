package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
)

// HealthHandler exposes liveness and readiness endpoints.
// HealthHandler 暴露存活检查和就绪检查接口。
type HealthHandler struct {
	cfg       *config.Config
	startedAt time.Time
}

// NewHealthHandler creates a health handler.
// NewHealthHandler 创建健康检查 HTTP handler。
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		cfg:       cfg,
		startedAt: time.Now().UTC(),
	}
}

// Health returns basic service liveness information.
// Health 返回服务存活状态。
func (h *HealthHandler) Health(c *gin.Context) {
	response.OK(c, gin.H{
		"service":    h.cfg.App.Name,
		"env":        h.cfg.App.Env,
		"status":     "ok",
		"started_at": h.startedAt,
	})
}

// Ready returns readiness information.
// Ready 返回服务就绪状态。
func (h *HealthHandler) Ready(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":       "OK",
		"message":    "ready",
		"request_id": response.RequestID(c),
		"data": gin.H{
			"status": "ready",
		},
	})
}
