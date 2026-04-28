package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
)

// HealthHandler exposes liveness and readiness endpoints.
type HealthHandler struct {
	cfg       *config.Config
	startedAt time.Time
}

// NewHealthHandler creates a health handler.
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		cfg:       cfg,
		startedAt: time.Now().UTC(),
	}
}

// Health returns basic service liveness information.
func (h *HealthHandler) Health(c *gin.Context) {
	response.OK(c, gin.H{
		"service":    h.cfg.App.Name,
		"env":        h.cfg.App.Env,
		"status":     "ok",
		"started_at": h.startedAt,
	})
}

// Ready returns readiness information.
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
