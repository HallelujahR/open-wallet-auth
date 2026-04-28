package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const RequestIDKey = "request_id"

type Body struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      any    `json:"data"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{
		Code:      "OK",
		Message:   "success",
		RequestID: RequestID(c),
		Data:      data,
	})
}

func Error(c *gin.Context, status int, code string, message string) {
	c.JSON(status, Body{
		Code:      code,
		Message:   message,
		RequestID: RequestID(c),
		Data:      nil,
	})
}

func RequestID(c *gin.Context) string {
	if v, ok := c.Get(RequestIDKey); ok {
		if requestID, ok := v.(string); ok {
			return requestID
		}
	}
	return ""
}
