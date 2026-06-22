package dto

// CreateClientRequest is the HTTP request body for creating an application client.
type CreateClientRequest struct {
	ClientID            string   `json:"client_id" binding:"required"`
	Name                string   `json:"name" binding:"required"`
	JWTAudience         string   `json:"jwt_audience"`
	AllowedOrigins      []string `json:"allowed_origins"`
	AllowedRedirectURIs []string `json:"allowed_redirect_uris"`
	WhitelistEnabled    bool     `json:"whitelist_enabled"`
}

// ClientResponse is the HTTP representation of an application client.
type ClientResponse struct {
	ID                  string   `json:"id"`
	ClientID            string   `json:"client_id"`
	Name                string   `json:"name"`
	JWTAudience         string   `json:"jwt_audience"`
	AllowedOrigins      []string `json:"allowed_origins"`
	AllowedRedirectURIs []string `json:"allowed_redirect_uris"`
	WhitelistEnabled    bool     `json:"whitelist_enabled"`
	Status              string   `json:"status"`
	CreatedAt           string   `json:"created_at"`
}

// UpdateClientAccessPolicyRequest is the HTTP request body for client access policy.
type UpdateClientAccessPolicyRequest struct {
	WhitelistEnabled bool `json:"whitelist_enabled"`
}

// ClientMemberRequest is the HTTP request body for application allow-list members.
type ClientMemberRequest struct {
	UserID      string   `json:"user_id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	Status      string   `json:"status"`
	Remark      string   `json:"remark"`
}

// ClientMemberResponse is the HTTP representation of an application allow-list member.
type ClientMemberResponse struct {
	ID          string   `json:"id"`
	ClientID    string   `json:"client_id"`
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	Status      string   `json:"status"`
	Remark      string   `json:"remark"`
	CreatedBy   string   `json:"created_by"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// PublicClientResponse is the safe client payload exposed to hosted login pages.
// PublicClientResponse 是统一登录页可读取的安全业务系统信息。
type PublicClientResponse struct {
	ClientID string `json:"client_id"`
	Name     string `json:"name"`
}

// LoginPageSettingsResponse is the public hosted login-page configuration.
// LoginPageSettingsResponse 是公开的统一登录页配置。
type LoginPageSettingsResponse struct {
	BrandName      string `json:"brand_name"`
	BrandMark      string `json:"brand_mark"`
	Subtitle       string `json:"subtitle"`
	EnableRegister bool   `json:"enable_register"`
	EnablePhone    bool   `json:"enable_phone"`
	EnableGitHub   bool   `json:"enable_github"`
	EnableGoogle   bool   `json:"enable_google"`
	EnableWallet   bool   `json:"enable_wallet"`
}

// LoginConfigResponse contains everything the hosted login page needs to render.
// LoginConfigResponse 包含统一登录页渲染所需的公开配置。
type LoginConfigResponse struct {
	Client PublicClientResponse      `json:"client"`
	Login  LoginPageSettingsResponse `json:"login"`
}
