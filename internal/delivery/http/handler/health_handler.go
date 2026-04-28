package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
)

type HealthHandler struct {
	cfg       *config.Config
	startedAt time.Time
}

func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		cfg:       cfg,
		startedAt: time.Now().UTC(),
	}
}

func (h *HealthHandler) Health(c *gin.Context) {
	response.OK(c, gin.H{
		"service":    h.cfg.App.Name,
		"env":        h.cfg.App.Env,
		"status":     "ok",
		"started_at": h.startedAt,
	})
}

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
