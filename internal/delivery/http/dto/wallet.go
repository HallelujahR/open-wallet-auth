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

// WalletBindRequest is the HTTP request body for binding a wallet to current user.
// WalletBindRequest 是为当前用户绑定钱包的 HTTP 请求体。
type WalletBindRequest struct {
	Address   string `json:"address" binding:"required"`
	Nonce     string `json:"nonce" binding:"required"`
	Signature string `json:"signature" binding:"required"`
}

// WalletBindResponse is returned after binding a wallet to current user.
// WalletBindResponse 是当前用户绑定钱包后的响应体。
type WalletBindResponse struct {
	WalletID   string `json:"wallet_id"`
	Address    string `json:"address"`
	ChainType  string `json:"chain_type"`
	VerifiedAt string `json:"verified_at"`
}

// WalletAuthResponse is returned after a successful wallet signature login.
type WalletAuthResponse struct {
	User    AuthUser  `json:"user"`
	Wallets []string  `json:"wallets"`
	Token   TokenPair `json:"token"`
}
