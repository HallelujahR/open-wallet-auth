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
