package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSOriginResolver supplies runtime-editable browser origins.
// CORSOriginResolver 提供运行期可编辑的浏览器来源配置。
type CORSOriginResolver interface {
	CORSAllowedOrigins(ctx context.Context) ([]string, error)
}

// CORS allows browser-based clients to call the auth service from configured origins.
// CORS 允许配置中的浏览器来源调用认证服务。
func CORS(allowedOrigins []string, resolver CORSOriginResolver) gin.HandlerFunc {
	staticOrigins := allowedOrigins
	return func(c *gin.Context) {
		origins := staticOrigins
		if resolver != nil {
			if current, err := resolver.CORSAllowedOrigins(c.Request.Context()); err == nil && len(current) > 0 {
				origins = current
			}
		}
		applyCORS(c, origins)
	}
}

// applyCORS evaluates one request against allowed origins.
// applyCORS 根据允许来源处理单个请求。
func applyCORS(c *gin.Context, allowedOrigins []string) {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	allowAll := false
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowAll = true
			continue
		}
		if origin != "" {
			allowed[origin] = struct{}{}
		}
	}

	origin := c.GetHeader("Origin")
	if origin != "" {
		if allowAll {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if _, ok := allowed[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Vary", "Origin")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Client-ID, X-Admin-Token, X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	}

	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}
