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

// SessionLoginRequest is the HTTP body for one-click login from the central session.
// SessionLoginRequest 是使用中台会话一键登录业务系统的 HTTP 请求体。
type SessionLoginRequest struct {
	ClientID string `json:"client_id"`
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

// UpdateProfileRequest is the HTTP body for updating display-only profile fields.
// UpdateProfileRequest 是更新展示型身份资料字段的 HTTP 请求体。
type UpdateProfileRequest struct {
	Username string `json:"username" binding:"required"`
	Avatar   string `json:"avatar"`
}

// ProfileWalletResponse describes one wallet binding in the profile view.
// ProfileWalletResponse 描述资料视图中的一个钱包绑定。
type ProfileWalletResponse struct {
	ID         string `json:"id"`
	ChainType  string `json:"chain_type"`
	Address    string `json:"address"`
	IsPrimary  bool   `json:"is_primary"`
	VerifiedAt string `json:"verified_at"`
	CreatedAt  string `json:"created_at"`
}

// ProfileOAuthAccountResponse describes one OAuth binding in the profile view.
// ProfileOAuthAccountResponse 描述资料视图中的一个 OAuth 绑定。
type ProfileOAuthAccountResponse struct {
	ID                string `json:"id"`
	Provider          string `json:"provider"`
	ProviderSubject   string `json:"provider_subject"`
	ProviderEmail     string `json:"provider_email,omitempty"`
	ProviderUsername  string `json:"provider_username,omitempty"`
	ProviderAvatarURL string `json:"provider_avatar_url,omitempty"`
	CreatedAt         string `json:"created_at"`
}

// ProfileResponse is the current user's persisted identity profile.
// ProfileResponse 是当前用户的持久化身份资料。
type ProfileResponse struct {
	ID           string                        `json:"id"`
	Username     string                        `json:"username"`
	Email        string                        `json:"email,omitempty"`
	Phone        string                        `json:"phone,omitempty"`
	Avatar       string                        `json:"avatar,omitempty"`
	Status       string                        `json:"status"`
	LoginMethods []string                      `json:"login_methods"`
	Wallets      []ProfileWalletResponse       `json:"wallets"`
	Accounts     []ProfileOAuthAccountResponse `json:"oauth_accounts"`
	LastLoginAt  string                        `json:"last_login_at,omitempty"`
	CreatedAt    string                        `json:"created_at"`
	UpdatedAt    string                        `json:"updated_at"`
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
