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
	walletusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/wallet"
)

// WalletHandler exposes wallet challenge and signature-login endpoints.
// WalletHandler 暴露钱包挑战值和签名登录接口。
type WalletHandler struct {
	wallet *walletusecase.Service
}

// NewWalletHandler creates a WalletHandler bound to the wallet usecase service.
// NewWalletHandler 创建绑定钱包用例服务的 HTTP handler。
func NewWalletHandler(wallet *walletusecase.Service) *WalletHandler {
	return &WalletHandler{wallet: wallet}
}

// Nonce creates the SIWE-compatible message that the browser wallet signs.
// Nonce 创建浏览器钱包需要签名的 SIWE 兼容消息。
func (h *WalletHandler) Nonce(c *gin.Context) {
	var req dto.WalletNonceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, walletusecase.ErrInvalidNonce, "invalid request")
		return
	}

	result, err := h.wallet.CreateNonce(c.Request.Context(), walletusecase.NonceRequest{
		Address: req.Address,
		Domain:  req.Domain,
		ChainID: req.ChainID,
	})
	if err != nil {
		writeWalletError(c, err)
		return
	}

	response.OK(c, dto.WalletNonceResponse{
		Nonce:     result.Nonce,
		Message:   result.Message,
		ExpiresAt: result.ExpiresAt.Format(timeFormatRFC3339),
	})
}

// Verify checks the wallet signature and returns an auth token pair.
// Verify 校验钱包签名并返回认证 token。
func (h *WalletHandler) Verify(c *gin.Context) {
	var req dto.WalletVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, walletusecase.ErrInvalidSignature, "invalid request")
		return
	}

	result, err := h.wallet.VerifySignature(c.Request.Context(), walletusecase.VerifyRequest{
		ClientID:  req.ClientID,
		Address:   req.Address,
		Nonce:     req.Nonce,
		Signature: req.Signature,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		writeWalletError(c, err)
		return
	}

	response.OK(c, dto.WalletAuthResponse{
		User: dto.AuthUser{
			ID:       result.UserID,
			Username: result.Username,
			Email:    result.Email,
		},
		Wallets: result.Wallets,
		Token: dto.TokenPair{
			AccessToken:  result.Token.AccessToken,
			RefreshToken: result.Token.RefreshToken,
			ExpiresAt:    result.Token.ExpiresAt.Format(timeFormatRFC3339),
			TokenType:    "Bearer",
		},
	})
}

// Bind verifies a wallet signature and binds the wallet to the current user.
// Bind 校验钱包签名，并把钱包绑定到当前登录用户。
func (h *WalletHandler) Bind(c *gin.Context) {
	authClaims, ok := currentWalletAuthClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid authorization token")
		return
	}

	var req dto.WalletBindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, walletusecase.ErrInvalidSignature, "invalid request")
		return
	}

	result, err := h.wallet.BindWallet(c.Request.Context(), walletusecase.BindRequest{
		UserID:    authClaims.UserID,
		Address:   req.Address,
		Nonce:     req.Nonce,
		Signature: req.Signature,
	})
	if err != nil {
		writeWalletError(c, err)
		return
	}

	response.OK(c, dto.WalletBindResponse{
		WalletID:   result.WalletID,
		Address:    result.Address,
		ChainType:  string(result.ChainType),
		VerifiedAt: result.VerifiedAt.Format(timeFormatRFC3339),
	})
}

// currentWalletAuthClaims reads token claims from Gin context.
// currentWalletAuthClaims 从 Gin 上下文读取认证 claims。
func currentWalletAuthClaims(c *gin.Context) (*token.Claims, bool) {
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

// writeWalletError maps wallet usecase errors to HTTP responses.
// writeWalletError 将钱包用例错误映射为 HTTP 响应。
func writeWalletError(c *gin.Context, err error) {
	var appErr *domain.Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case walletusecase.ErrInvalidClient, walletusecase.ErrInvalidNonce, walletusecase.ErrInvalidWalletAddress:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		case walletusecase.ErrInvalidSignature:
			response.Error(c, http.StatusUnauthorized, appErr.Code, appErr.Message)
		case walletusecase.ErrWalletAlreadyBound:
			response.Error(c, http.StatusConflict, appErr.Code, appErr.Message)
		case walletusecase.ErrRateLimited:
			response.Error(c, http.StatusTooManyRequests, appErr.Code, appErr.Message)
		default:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		}
		return
	}

	response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
