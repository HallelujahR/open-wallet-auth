package dto

// RegisterRequest is the HTTP request body for registration.
type RegisterRequest struct {
	ClientID string `json:"client_id"`
	Username string `json:"username"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest is the HTTP request body for password login.
type LoginRequest struct {
	ClientID string `json:"client_id"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest is the HTTP request body for token rotation.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest is the HTTP request body for logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest is the HTTP body for changing the current password.
// ChangePasswordRequest 是修改当前用户密码的 HTTP 请求体。
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// ResetPasswordRequest is the HTTP body for resetting a password with email code.
// ResetPasswordRequest 是使用邮箱验证码重置密码的 HTTP 请求体。
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// BindEmailRequest is the HTTP body for binding an email to current user.
// BindEmailRequest 是当前用户绑定邮箱的 HTTP 请求体。
type BindEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

// BindPhoneRequest is the HTTP body for binding a phone number to current user.
// BindPhoneRequest 是当前用户绑定手机号的 HTTP 请求体。
type BindPhoneRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// BindContactResponse is returned after binding an email or phone number.
// BindContactResponse 是邮箱或手机号绑定成功后的响应体。
type BindContactResponse struct {
	UserID string `json:"user_id"`
	Value  string `json:"value"`
}

// AuthUser is the user payload returned by auth endpoints.
type AuthUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// TokenPair is the token payload returned by auth endpoints.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
	TokenType    string `json:"token_type"`
}

// AuthResponse is the successful response payload for auth endpoints.
type AuthResponse struct {
	User  AuthUser  `json:"user"`
	Token TokenPair `json:"token"`
}
