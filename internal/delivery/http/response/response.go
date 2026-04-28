package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const RequestIDKey = "request_id"

// Body is the standard HTTP response envelope.
type Body struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      any    `json:"data"`
}

// OK writes a successful response envelope.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{
		Code:      "OK",
		Message:   "success",
		RequestID: RequestID(c),
		Data:      data,
	})
}

// Error writes an error response envelope.
func Error(c *gin.Context, status int, code string, message string) {
	c.JSON(status, Body{
		Code:      code,
		Message:   message,
		RequestID: RequestID(c),
		Data:      nil,
	})
}

// RequestID returns the request id stored in the Gin context.
func RequestID(c *gin.Context) string {
	if v, ok := c.Get(RequestIDKey); ok {
		if requestID, ok := v.(string); ok {
			return requestID
		}
	}
	return ""
}
