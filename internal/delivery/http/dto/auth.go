package dto

type RegisterRequest struct {
	ClientID string `json:"client_id"`
	Username string `json:"username"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	ClientID string `json:"client_id"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
	TokenType    string `json:"token_type"`
}

type AuthResponse struct {
	User  AuthUser  `json:"user"`
	Token TokenPair `json:"token"`
}
