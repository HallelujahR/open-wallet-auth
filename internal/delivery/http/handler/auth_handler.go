package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/contextkey"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/dto"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	authusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/auth"
)

// AuthHandler exposes authentication usecases as HTTP endpoints.
// AuthHandler 将认证用例暴露为 HTTP 接口。
type AuthHandler struct {
	auth *authusecase.Service
}

// NewAuthHandler creates an AuthHandler bound to the auth usecase service.
// NewAuthHandler 创建绑定认证用例服务的 HTTP handler。
func NewAuthHandler(auth *authusecase.Service) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register handles email/password registration and returns a token pair.
// Register 处理邮箱密码注册请求，并返回 token 组合。
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}

	result, err := h.auth.Register(c.Request.Context(), authusecase.RegisterRequest{
		ClientID:  req.ClientID,
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		writeAuthError(c, err)
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

// Login handles email/password authentication and returns a token pair.
// Login 处理邮箱密码登录请求，并返回 token 组合。
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}

	result, err := h.auth.Login(c.Request.Context(), authusecase.LoginRequest{
		ClientID:  req.ClientID,
		Email:     req.Email,
		Password:  req.Password,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		writeAuthError(c, err)
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

// Refresh rotates a refresh token and returns a new token pair.
// Refresh 轮换刷新令牌，并返回新的 token 组合。
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}

	result, err := h.auth.Refresh(c.Request.Context(), authusecase.RefreshRequest{
		RefreshToken: req.RefreshToken,
		IP:           c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})
	if err != nil {
		writeAuthError(c, err)
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

// Logout revokes a refresh token.
// Logout 吊销刷新令牌。
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}

	if err := h.auth.Logout(c.Request.Context(), authusecase.LogoutRequest{
		RefreshToken: req.RefreshToken,
	}); err != nil {
		writeAuthError(c, err)
		return
	}

	response.OK(c, gin.H{"logged_out": true})
}

// Profile returns the current user's persisted identity profile.
// Profile 返回当前用户的持久化身份资料。
func (h *AuthHandler) Profile(c *gin.Context) {
	authClaims, ok := currentAuthClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid authorization token")
		return
	}
	result, err := h.auth.GetProfile(c.Request.Context(), authClaims.UserID)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.OK(c, profileResponse(result))
}

// UpdateProfile updates display-only profile fields for the current user.
// UpdateProfile 更新当前用户的展示型身份资料字段。
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	authClaims, ok := currentAuthClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid authorization token")
		return
	}
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}
	result, err := h.auth.UpdateProfile(c.Request.Context(), authusecase.UpdateProfileRequest{
		UserID:   authClaims.UserID,
		Username: req.Username,
		Avatar:   req.Avatar,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.OK(c, profileResponse(result))
}

// ChangePassword updates the authenticated user's password.
// ChangePassword 修改当前登录用户的密码。
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	authClaims, ok := currentAuthClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid authorization token")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}

	if err := h.auth.ChangePassword(c.Request.Context(), authusecase.ChangePasswordRequest{
		UserID:          authClaims.UserID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}); err != nil {
		writeAuthError(c, err)
		return
	}

	response.OK(c, gin.H{"password_changed": true})
}

// ResetPassword resets a password with a verified email code.
// ResetPassword 使用邮箱验证码重置用户密码。
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}

	if err := h.auth.ResetPassword(c.Request.Context(), authusecase.ResetPasswordRequest{
		Email:       req.Email,
		Code:        req.Code,
		NewPassword: req.NewPassword,
	}); err != nil {
		writeAuthError(c, err)
		return
	}

	response.OK(c, gin.H{"password_reset": true})
}

// BindEmail verifies an email code and binds the email to the current user.
// BindEmail 校验邮箱验证码，并把邮箱绑定到当前登录用户。
func (h *AuthHandler) BindEmail(c *gin.Context) {
	authClaims, ok := currentAuthClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid authorization token")
		return
	}

	var req dto.BindEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}
	result, err := h.auth.BindEmail(c.Request.Context(), authusecase.BindEmailRequest{
		UserID: authClaims.UserID,
		Email:  req.Email,
		Code:   req.Code,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.OK(c, dto.BindContactResponse{UserID: result.UserID, Value: result.Value})
}

// BindPhone verifies a phone code and binds the phone number to the current user.
// BindPhone 校验手机号验证码，并把手机号绑定到当前登录用户。
func (h *AuthHandler) BindPhone(c *gin.Context) {
	authClaims, ok := currentAuthClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid authorization token")
		return
	}

	var req dto.BindPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}
	result, err := h.auth.BindPhone(c.Request.Context(), authusecase.BindPhoneRequest{
		UserID: authClaims.UserID,
		Phone:  req.Phone,
		Code:   req.Code,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.OK(c, dto.BindContactResponse{UserID: result.UserID, Value: result.Value})
}

// UnbindEmail removes the current user's email binding when another method remains.
// UnbindEmail 在仍保留其他登录方式时解绑当前用户邮箱。
func (h *AuthHandler) UnbindEmail(c *gin.Context) {
	authClaims, ok := currentAuthClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid authorization token")
		return
	}
	if err := h.auth.UnbindEmail(c.Request.Context(), authClaims.UserID); err != nil {
		writeAuthError(c, err)
		return
	}
	response.OK(c, gin.H{"email_unbound": true})
}

// UnbindPhone removes the current user's phone binding when another method remains.
// UnbindPhone 在仍保留其他登录方式时解绑当前用户手机号。
func (h *AuthHandler) UnbindPhone(c *gin.Context) {
	authClaims, ok := currentAuthClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid authorization token")
		return
	}
	if err := h.auth.UnbindPhone(c.Request.Context(), authClaims.UserID); err != nil {
		writeAuthError(c, err)
		return
	}
	response.OK(c, gin.H{"phone_unbound": true})
}

// UnbindWallet removes one wallet binding owned by the current user.
// UnbindWallet 解绑当前用户拥有的一个钱包。
func (h *AuthHandler) UnbindWallet(c *gin.Context) {
	authClaims, ok := currentAuthClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid authorization token")
		return
	}
	if err := h.auth.UnbindWallet(c.Request.Context(), authusecase.UnbindRequest{
		UserID:    authClaims.UserID,
		BindingID: c.Param("wallet_id"),
	}); err != nil {
		writeAuthError(c, err)
		return
	}
	response.OK(c, gin.H{"wallet_unbound": true})
}

// UnbindOAuthAccount removes one OAuth binding owned by the current user.
// UnbindOAuthAccount 解绑当前用户拥有的一个 OAuth 账号。
func (h *AuthHandler) UnbindOAuthAccount(c *gin.Context) {
	authClaims, ok := currentAuthClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid authorization token")
		return
	}
	if err := h.auth.UnbindOAuthAccount(c.Request.Context(), authusecase.UnbindRequest{
		UserID:    authClaims.UserID,
		BindingID: c.Param("account_id"),
	}); err != nil {
		writeAuthError(c, err)
		return
	}
	response.OK(c, gin.H{"oauth_account_unbound": true})
}

// Me returns the authenticated user claims injected by auth middleware.
// Me 返回认证中间件注入的当前用户 claims。
func (h *AuthHandler) Me(c *gin.Context) {
	authClaims, ok := currentAuthClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid authorization token")
		return
	}

	response.OK(c, dto.AuthUser{
		ID:       authClaims.UserID,
		Username: authClaims.Username,
		Email:    authClaims.Email,
	})
}

// currentAuthClaims reads token claims from Gin context.
// currentAuthClaims 从 Gin 上下文读取认证 claims。
func currentAuthClaims(c *gin.Context) (*token.Claims, bool) {
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

// profileResponse converts a usecase profile result into an HTTP response.
// profileResponse 将用例资料结果转换为 HTTP 响应。
func profileResponse(result *authusecase.ProfileResult) dto.ProfileResponse {
	wallets := make([]dto.ProfileWalletResponse, 0, len(result.Wallets))
	for _, wallet := range result.Wallets {
		wallets = append(wallets, dto.ProfileWalletResponse{
			ID:         wallet.ID,
			ChainType:  string(wallet.ChainType),
			Address:    wallet.Address,
			IsPrimary:  wallet.IsPrimary,
			VerifiedAt: wallet.VerifiedAt.Format(timeFormatRFC3339),
			CreatedAt:  wallet.CreatedAt.Format(timeFormatRFC3339),
		})
	}
	accounts := make([]dto.ProfileOAuthAccountResponse, 0, len(result.Accounts))
	for _, account := range result.Accounts {
		accounts = append(accounts, dto.ProfileOAuthAccountResponse{
			ID:                account.ID,
			Provider:          account.Provider,
			ProviderSubject:   account.ProviderSubject,
			ProviderEmail:     account.ProviderEmail,
			ProviderUsername:  account.ProviderUsername,
			ProviderAvatarURL: account.ProviderAvatarURL,
			CreatedAt:         account.CreatedAt.Format(timeFormatRFC3339),
		})
	}
	return dto.ProfileResponse{
		ID:           result.User.ID,
		Username:     result.User.Username,
		Email:        result.User.Email,
		Phone:        result.User.Phone,
		Avatar:       result.User.Avatar,
		Status:       string(result.User.Status),
		LoginMethods: result.LoginMethods,
		Wallets:      wallets,
		Accounts:     accounts,
		LastLoginAt:  formatAuthOptionalTime(result.User.LastLoginAt),
		CreatedAt:    result.User.CreatedAt.Format(timeFormatRFC3339),
		UpdatedAt:    result.User.UpdatedAt.Format(timeFormatRFC3339),
	}
}

// formatOptionalTime returns an RFC3339 string for nullable timestamps.
// formatOptionalTime 将可空时间格式化为 RFC3339 字符串。
func formatAuthOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(timeFormatRFC3339)
}

// writeAuthError maps usecase/domain errors to stable HTTP responses.
// writeAuthError 将用例/领域错误映射为稳定的 HTTP 响应。
func writeAuthError(c *gin.Context, err error) {
	var appErr *domain.Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case authusecase.ErrEmailAlreadyExists:
			response.Error(c, http.StatusConflict, appErr.Code, appErr.Message)
		case authusecase.ErrEmailAlreadyBound, authusecase.ErrPhoneAlreadyBound:
			response.Error(c, http.StatusConflict, appErr.Code, appErr.Message)
		case authusecase.ErrInvalidClient:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		case authusecase.ErrInvalidCode:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		case authusecase.ErrInvalidCredentials:
			response.Error(c, http.StatusUnauthorized, appErr.Code, appErr.Message)
		case authusecase.ErrInvalidRefreshToken:
			response.Error(c, http.StatusUnauthorized, appErr.Code, appErr.Message)
		case authusecase.ErrBindingNotFound:
			response.Error(c, http.StatusNotFound, appErr.Code, appErr.Message)
		case authusecase.ErrLastLoginMethod:
			response.Error(c, http.StatusConflict, appErr.Code, appErr.Message)
		case authusecase.ErrRateLimited:
			response.Error(c, http.StatusTooManyRequests, appErr.Code, appErr.Message)
		default:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		}
		return
	}

	response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}

const timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"
