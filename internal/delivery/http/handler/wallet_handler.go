package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/dto"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	walletusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/wallet"
)

// WalletHandler exposes wallet challenge and signature-login endpoints.
type WalletHandler struct {
	wallet *walletusecase.Service
}

// NewWalletHandler creates a WalletHandler bound to the wallet usecase service.
func NewWalletHandler(wallet *walletusecase.Service) *WalletHandler {
	return &WalletHandler{wallet: wallet}
}

// Nonce creates the SIWE-compatible message that the browser wallet signs.
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

func writeWalletError(c *gin.Context, err error) {
	var appErr *domain.Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case walletusecase.ErrInvalidClient, walletusecase.ErrInvalidNonce, walletusecase.ErrInvalidWalletAddress:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		case walletusecase.ErrInvalidSignature:
			response.Error(c, http.StatusUnauthorized, appErr.Code, appErr.Message)
		default:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		}
		return
	}

	response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
