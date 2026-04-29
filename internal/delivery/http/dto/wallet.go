package dto

// WalletNonceRequest is the HTTP request body for wallet login challenges.
type WalletNonceRequest struct {
	Address string `json:"address" binding:"required"`
	Domain  string `json:"domain" binding:"required"`
	ChainID int64  `json:"chain_id"`
}

// WalletNonceResponse is returned after creating a wallet login challenge.
type WalletNonceResponse struct {
	Nonce     string `json:"nonce"`
	Message   string `json:"message"`
	ExpiresAt string `json:"expires_at"`
}

// WalletVerifyRequest is the HTTP request body for wallet signature login.
type WalletVerifyRequest struct {
	ClientID  string `json:"client_id"`
	Address   string `json:"address" binding:"required"`
	Nonce     string `json:"nonce" binding:"required"`
	Signature string `json:"signature" binding:"required"`
}

// WalletAuthResponse is returned after a successful wallet signature login.
type WalletAuthResponse struct {
	User    AuthUser  `json:"user"`
	Wallets []string  `json:"wallets"`
	Token   TokenPair `json:"token"`
}
