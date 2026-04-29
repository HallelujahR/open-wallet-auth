package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
)

const RequestIDHeader = "X-Request-ID"

// RequestID ensures every request has a stable request id.
// RequestID 确保每个请求都有稳定的链路 ID。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set(response.RequestIDKey, requestID)
		c.Writer.Header().Set(RequestIDHeader, requestID)
		c.Next()
	}
}
