package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const RequestIDKey = "request_id"

// Body is the standard HTTP response envelope.
// Body 是 HTTP 接口统一响应包裹结构，方便接入方稳定解析。
type Body struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      any    `json:"data"`
}

// OK writes a successful response envelope.
// OK 写入成功响应，并自动带上 request_id。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{
		Code:      "OK",
		Message:   "success",
		RequestID: RequestID(c),
		Data:      data,
	})
}

// Error writes an error response envelope.
// Error 写入失败响应，错误码保持机器可读。
func Error(c *gin.Context, status int, code string, message string) {
	c.JSON(status, Body{
		Code:      code,
		Message:   message,
		RequestID: RequestID(c),
		Data:      nil,
	})
}

// RequestID returns the request id stored in the Gin context.
// RequestID 从 Gin 上下文中取出请求链路 ID。
func RequestID(c *gin.Context) string {
	if v, ok := c.Get(RequestIDKey); ok {
		if requestID, ok := v.(string); ok {
			return requestID
		}
	}
	return ""
}
