package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/dto"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	authusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/auth"
)

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
		IP:              c.ClientIP(),
		UserAgent:       c.Request.UserAgent(),
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
		IP:          c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
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
		UserID:    authClaims.UserID,
		Email:     req.Email,
		Code:      req.Code,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
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
		UserID:    authClaims.UserID,
		Phone:     req.Phone,
		Code:      req.Code,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
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
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
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
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
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
