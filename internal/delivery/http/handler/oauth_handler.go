package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/contextkey"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/dto"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
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

// BindStart creates an OAuth authorization URL for binding to the current user.
// BindStart 为当前登录用户绑定 OAuth 账号创建第三方授权地址。
func (h *OAuthHandler) BindStart(c *gin.Context) {
	authClaims, ok := currentOAuthAuthClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid authorization token")
		return
	}
	result, err := h.oauth.Start(c.Request.Context(), oauthusecase.StartRequest{
		Provider:    c.Param("provider"),
		ClientID:    c.Query("client_id"),
		RedirectURI: c.Query("redirect_uri"),
		BindUserID:  authClaims.UserID,
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

// currentOAuthAuthClaims reads token claims from Gin context.
// currentOAuthAuthClaims 从 Gin 上下文读取认证 claims。
func currentOAuthAuthClaims(c *gin.Context) (*token.Claims, bool) {
	claims, ok := c.Get(contextkey.AuthClaims)
	if !ok {
		return nil, false
	}
	authClaims, ok := claims.(*token.Claims)
	if !ok {
		return nil, false
	}
	return authClaims, true
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
		case oauthusecase.ErrOAuthBound:
			response.Error(c, http.StatusConflict, appErr.Code, appErr.Message)
		default:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		}
		return
	}
	response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
