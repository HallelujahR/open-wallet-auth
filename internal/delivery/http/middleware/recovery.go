package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
)

// Recovery converts panics into safe HTTP 500 responses and structured logs.
// Recovery 将 panic 转为安全的 500 响应，并记录结构化错误日志。
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error("panic recovered",
			zap.Any("panic", recovered),
			zap.String("request_id", response.RequestID(c)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	})
}
