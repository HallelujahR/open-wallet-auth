package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/dto"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	oauthusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/oauth"
)

// OAuthHandler exposes third-party OAuth start and callback endpoints.
// OAuthHandler 暴露第三方 OAuth 发起与回调接口。
type OAuthHandler struct {
	oauth *oauthusecase.Service
}

// NewOAuthHandler creates an OAuthHandler bound to the OAuth usecase service.
// NewOAuthHandler 创建绑定 OAuth 用例服务的 HTTP handler。
func NewOAuthHandler(oauth *oauthusecase.Service) *OAuthHandler {
	return &OAuthHandler{oauth: oauth}
}

// Start creates a provider authorization URL.
// Start 创建第三方服务商授权地址。
func (h *OAuthHandler) Start(c *gin.Context) {
	result, err := h.oauth.Start(c.Request.Context(), oauthusecase.StartRequest{
		Provider:    c.Param("provider"),
		ClientID:    c.Query("client_id"),
		RedirectURI: c.Query("redirect_uri"),
	})
	if err != nil {
		writeOAuthError(c, err)
		return
	}
	response.OK(c, dto.OAuthStartResponse{Provider: result.Provider, AuthURL: result.AuthURL, State: result.State})
}

// Callback completes provider login and returns a token pair.
// Callback 完成第三方登录回调并返回 token 组合。
func (h *OAuthHandler) Callback(c *gin.Context) {
	result, err := h.oauth.Callback(c.Request.Context(), oauthusecase.CallbackRequest{
		Provider:  c.Param("provider"),
		Code:      c.Query("code"),
		State:     c.Query("state"),
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		writeOAuthError(c, err)
		return
	}
	response.OK(c, dto.AuthResponse{
		User: dto.AuthUser{
			ID:       result.UserID,
			Username: result.Username,
			Email:    result.Email,
		},
		Token: dto.TokenPair{
			AccessToken:  result.Token.AccessToken,
			RefreshToken: result.Token.RefreshToken,
			ExpiresAt:    result.Token.ExpiresAt.Format(timeFormatRFC3339),
			TokenType:    "Bearer",
		},
	})
}

// writeOAuthError maps OAuth usecase errors to HTTP responses.
// writeOAuthError 将 OAuth 用例错误映射为 HTTP 响应。
func writeOAuthError(c *gin.Context, err error) {
	var appErr *domain.Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case oauthusecase.ErrProviderFailed:
			response.Error(c, http.StatusServiceUnavailable, appErr.Code, appErr.Message)
		case oauthusecase.ErrInvalidState:
			response.Error(c, http.StatusUnauthorized, appErr.Code, appErr.Message)
		default:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		}
		return
	}
	response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
