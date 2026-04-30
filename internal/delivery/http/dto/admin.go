package dto

// AdminUserResponse is the management view of an identity user.
// AdminUserResponse 是身份用户的管理视图。
type AdminUserResponse struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	Status      string `json:"status"`
	LastLoginAt string `json:"last_login_at,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// AdminUserListResponse is the paginated identity-user list response.
// AdminUserListResponse 是分页身份用户列表响应。
type AdminUserListResponse struct {
	Items    []AdminUserResponse `json:"items"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// AdminUpdateUserStatusRequest is the request body for changing identity status.
// AdminUpdateUserStatusRequest 是修改身份用户状态的请求体。
type AdminUpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// AdminUserClientResponse describes one business system a user has logged into.
// AdminUserClientResponse 描述用户登录过的一个业务系统。
type AdminUserClientResponse struct {
	ClientID     string `json:"client_id"`
	FirstLoginAt string `json:"first_login_at"`
	LastLoginAt  string `json:"last_login_at"`
	LoginCount   int64  `json:"login_count"`
	Status       string `json:"status"`
}

// AdminWalletResponse describes one wallet binding.
// AdminWalletResponse 描述一个钱包绑定关系。
type AdminWalletResponse struct {
	ID         string `json:"id"`
	ChainType  string `json:"chain_type"`
	Address    string `json:"address"`
	IsPrimary  bool   `json:"is_primary"`
	VerifiedAt string `json:"verified_at"`
	CreatedAt  string `json:"created_at"`
}

// AdminOAuthAccountResponse describes one third-party account binding.
// AdminOAuthAccountResponse 描述一个第三方账号绑定关系。
type AdminOAuthAccountResponse struct {
	ID                string `json:"id"`
	Provider          string `json:"provider"`
	ProviderSubject   string `json:"provider_subject"`
	ProviderEmail     string `json:"provider_email,omitempty"`
	ProviderUsername  string `json:"provider_username,omitempty"`
	ProviderAvatarURL string `json:"provider_avatar_url,omitempty"`
	CreatedAt         string `json:"created_at"`
}

// AdminUserDetailResponse aggregates identity, client, wallet, and OAuth data.
// AdminUserDetailResponse 聚合身份、业务系统、钱包和第三方账号数据。
type AdminUserDetailResponse struct {
	User     AdminUserResponse           `json:"user"`
	Clients  []AdminUserClientResponse   `json:"clients"`
	Wallets  []AdminWalletResponse       `json:"wallets"`
	Accounts []AdminOAuthAccountResponse `json:"accounts"`
	Sessions []AdminSessionResponse      `json:"sessions"`
}

// AdminSessionResponse is the management view of one refresh-token session.
// AdminSessionResponse 是单个刷新令牌会话的管理视图。
type AdminSessionResponse struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	ClientID   string `json:"client_id"`
	IP         string `json:"ip,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
	Active     bool   `json:"active"`
	ExpiresAt  string `json:"expires_at"`
	RevokedAt  string `json:"revoked_at,omitempty"`
	LastUsedAt string `json:"last_used_at,omitempty"`
	CreatedAt  string `json:"created_at"`
}

// AdminSessionListResponse is the token-session list response.
// AdminSessionListResponse 是 token 会话列表响应。
type AdminSessionListResponse struct {
	Items []AdminSessionResponse `json:"items"`
}

// AdminRevokeSessionsResponse describes revoked session count.
// AdminRevokeSessionsResponse 描述已吊销的会话数量。
type AdminRevokeSessionsResponse struct {
	Revoked int64 `json:"revoked"`
}

// AdminLoginLogResponse is the management view of one login audit event.
// AdminLoginLogResponse 是单条登录审计事件的管理视图。
type AdminLoginLogResponse struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	ClientID    string `json:"client_id"`
	LoginMethod string `json:"login_method"`
	IP          string `json:"ip,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	Success     bool   `json:"success"`
	FailureCode string `json:"failure_code,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// AdminLoginLogListResponse is the paginated login-log response.
// AdminLoginLogListResponse 是分页登录日志响应。
type AdminLoginLogListResponse struct {
	Items    []AdminLoginLogResponse `json:"items"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}

// AdminSecurityEventResponse is the management view of one sensitive-operation event.
// AdminSecurityEventResponse 是单条敏感操作审计事件的管理视图。
type AdminSecurityEventResponse struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	EventType   string `json:"event_type"`
	TargetType  string `json:"target_type,omitempty"`
	TargetID    string `json:"target_id,omitempty"`
	IP          string `json:"ip,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	Success     bool   `json:"success"`
	FailureCode string `json:"failure_code,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// AdminSecurityEventListResponse is the paginated security-event response.
// AdminSecurityEventListResponse 是分页安全操作审计响应。
type AdminSecurityEventListResponse struct {
	Items    []AdminSecurityEventResponse `json:"items"`
	Total    int64                        `json:"total"`
	Page     int                          `json:"page"`
	PageSize int                          `json:"page_size"`
}
