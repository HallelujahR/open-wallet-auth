package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/dto"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/audit"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/oauth"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/wallet"
	adminusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/admin"
	authusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/auth"
)

// AdminHandler exposes identity-management APIs for internal operations.
// AdminHandler 暴露公司内部使用的统一身份管理接口。
type AdminHandler struct {
	admin       *adminusecase.Service
	tokenHasher authusecase.TokenHasher
}

// NewAdminHandler creates an AdminHandler bound to the admin usecase service.
// NewAdminHandler 创建绑定身份管理用例服务的 HTTP handler。
func NewAdminHandler(admin *adminusecase.Service, tokenHasher authusecase.TokenHasher) *AdminHandler {
	return &AdminHandler{admin: admin, tokenHasher: tokenHasher}
}

// ListUsers returns paginated identity users.
// ListUsers 返回分页身份用户列表。
func (h *AdminHandler) ListUsers(c *gin.Context) {
	result, err := h.admin.ListUsers(c.Request.Context(), adminusecase.UserListRequest{
		Keyword:  c.Query("keyword"),
		Status:   c.Query("status"),
		Page:     intQuery(c, "page", 1),
		PageSize: intQuery(c, "page_size", 20),
	})
	if err != nil {
		writeAdminError(c, err)
		return
	}

	items := make([]dto.AdminUserResponse, 0, len(result.Users))
	for _, u := range result.Users {
		items = append(items, toAdminUserResponse(u, result.Wallets[u.ID], result.Accounts[u.ID]))
	}
	response.OK(c, dto.AdminUserListResponse{
		Items:    items,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// GetUser returns one identity user's detail and bindings.
// GetUser 返回单个身份用户详情和账号绑定信息。
func (h *AdminHandler) GetUser(c *gin.Context) {
	result, err := h.admin.GetUserDetail(c.Request.Context(), c.Param("user_id"))
	if err != nil {
		writeAdminError(c, err)
		return
	}
	response.OK(c, toAdminUserDetailResponse(result))
}

// UpdateUserStatus enables, suspends, or marks an identity as deleted.
// UpdateUserStatus 启用、禁用或标记删除身份用户。
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	var req dto.AdminUpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, adminusecase.ErrInvalidInput, "invalid request")
		return
	}
	if err := h.admin.UpdateUserStatus(c.Request.Context(), adminusecase.UpdateUserStatusRequest{
		UserID: c.Param("user_id"),
		Status: req.Status,
	}); err != nil {
		writeAdminError(c, err)
		return
	}
	response.OK(c, gin.H{"updated": true})
}

// ListSessions returns refresh-token sessions for management.
// ListSessions 返回管理端可见的刷新令牌会话列表。
func (h *AdminHandler) ListSessions(c *gin.Context) {
	result, err := h.admin.ListSessions(c.Request.Context(), adminusecase.SessionListRequest{
		UserID:     c.Query("user_id"),
		ClientID:   c.Query("client_id"),
		ActiveOnly: boolQuery(c, "active_only", true),
	})
	if err != nil {
		writeAdminError(c, err)
		return
	}
	items := make([]dto.AdminSessionResponse, 0, len(result.Sessions))
	currentHash := h.currentSessionHash(c)
	for _, session := range result.Sessions {
		items = append(items, toAdminSessionResponse(session, currentHash))
	}
	response.OK(c, dto.AdminSessionListResponse{Items: items})
}

// RevokeSession revokes one refresh-token session.
// RevokeSession 吊销单个刷新令牌会话。
func (h *AdminHandler) RevokeSession(c *gin.Context) {
	if err := h.admin.RevokeSession(c.Request.Context(), c.Param("session_id")); err != nil {
		writeAdminError(c, err)
		return
	}
	response.OK(c, gin.H{"revoked": true})
}

// RevokeUserSessions revokes all or client-scoped sessions for one user.
// RevokeUserSessions 吊销某个用户的全部或指定业务系统会话。
func (h *AdminHandler) RevokeUserSessions(c *gin.Context) {
	result, err := h.admin.RevokeUserSessions(c.Request.Context(), adminusecase.RevokeUserSessionsRequest{
		UserID:   c.Param("user_id"),
		ClientID: c.Query("client_id"),
	})
	if err != nil {
		writeAdminError(c, err)
		return
	}
	response.OK(c, dto.AdminRevokeSessionsResponse{Revoked: result.Revoked})
}

// UnbindWallet removes a wallet binding from one identity user.
// UnbindWallet 从身份用户上解绑一个钱包。
func (h *AdminHandler) UnbindWallet(c *gin.Context) {
	if err := h.admin.UnbindWallet(c.Request.Context(), adminusecase.UnbindRequest{
		UserID:    c.Param("user_id"),
		BindingID: c.Param("wallet_id"),
	}); err != nil {
		writeAdminError(c, err)
		return
	}
	response.OK(c, gin.H{"unbound": true})
}

// UnbindOAuthAccount removes an OAuth account binding from one identity user.
// UnbindOAuthAccount 从身份用户上解绑一个第三方账号。
func (h *AdminHandler) UnbindOAuthAccount(c *gin.Context) {
	if err := h.admin.UnbindOAuthAccount(c.Request.Context(), adminusecase.UnbindRequest{
		UserID:    c.Param("user_id"),
		BindingID: c.Param("account_id"),
	}); err != nil {
		writeAdminError(c, err)
		return
	}
	response.OK(c, gin.H{"unbound": true})
}

// ListLoginLogs returns paginated login audit events.
// ListLoginLogs 返回分页登录审计事件。
func (h *AdminHandler) ListLoginLogs(c *gin.Context) {
	result, err := h.admin.ListLoginLogs(c.Request.Context(), adminusecase.LoginLogListRequest{
		UserID:   c.Query("user_id"),
		ClientID: c.Query("client_id"),
		Page:     intQuery(c, "page", 1),
		PageSize: intQuery(c, "page_size", 20),
	})
	if err != nil {
		writeAdminError(c, err)
		return
	}

	items := make([]dto.AdminLoginLogResponse, 0, len(result.Logs))
	for _, log := range result.Logs {
		items = append(items, toAdminLoginLogResponse(log))
	}
	response.OK(c, dto.AdminLoginLogListResponse{
		Items:    items,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// ListSecurityEvents returns paginated sensitive-operation audit events.
// ListSecurityEvents 返回分页敏感操作审计事件。
func (h *AdminHandler) ListSecurityEvents(c *gin.Context) {
	result, err := h.admin.ListSecurityEvents(c.Request.Context(), adminusecase.SecurityEventListRequest{
		UserID:    c.Query("user_id"),
		EventType: c.Query("event_type"),
		Page:      intQuery(c, "page", 1),
		PageSize:  intQuery(c, "page_size", 20),
	})
	if err != nil {
		writeAdminError(c, err)
		return
	}

	items := make([]dto.AdminSecurityEventResponse, 0, len(result.Events))
	for _, event := range result.Events {
		items = append(items, toAdminSecurityEventResponse(event))
	}
	response.OK(c, dto.AdminSecurityEventListResponse{
		Items:    items,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// writeAdminError maps admin usecase errors to HTTP responses.
// writeAdminError 将身份管理用例错误映射为 HTTP 响应。
func writeAdminError(c *gin.Context, err error) {
	var appErr *domain.Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case adminusecase.ErrUserNotFound:
			response.Error(c, http.StatusNotFound, appErr.Code, appErr.Message)
		case adminusecase.ErrBindingNotFound:
			response.Error(c, http.StatusNotFound, appErr.Code, appErr.Message)
		default:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		}
		return
	}
	response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}

// toAdminUserDetailResponse converts aggregated usecase data to an HTTP DTO.
// toAdminUserDetailResponse 将聚合后的用例数据转换为 HTTP DTO。
func toAdminUserDetailResponse(result *adminusecase.UserDetailResult) dto.AdminUserDetailResponse {
	clients := make([]dto.AdminUserClientResponse, 0, len(result.Clients))
	for _, client := range result.Clients {
		clients = append(clients, toAdminUserClientResponse(client))
	}
	wallets := toAdminWalletResponses(result.Wallets)
	accounts := toAdminOAuthAccountResponses(result.Accounts)
	sessions := make([]dto.AdminSessionResponse, 0, len(result.Sessions))
	for _, session := range result.Sessions {
		sessions = append(sessions, toAdminSessionResponse(session, ""))
	}
	return dto.AdminUserDetailResponse{
		User:     toAdminUserResponse(result.User, result.Wallets, result.Accounts),
		Clients:  clients,
		Wallets:  wallets,
		Accounts: accounts,
		Sessions: sessions,
	}
}

// toAdminUserResponse converts a domain user to a management DTO.
// toAdminUserResponse 将领域用户转换为管理接口 DTO。
func toAdminUserResponse(u user.User, wallets []wallet.UserWallet, accounts []oauth.Account) dto.AdminUserResponse {
	return dto.AdminUserResponse{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		Phone:        u.Phone,
		Avatar:       u.Avatar,
		Status:       string(u.Status),
		LoginMethods: adminLoginMethods(u, wallets, accounts),
		Wallets:      toAdminWalletResponses(wallets),
		Accounts:     toAdminOAuthAccountResponses(accounts),
		LastLoginAt:  formatOptionalTime(u.LastLoginAt),
		CreatedAt:    formatTime(u.CreatedAt),
		UpdatedAt:    formatTime(u.UpdatedAt),
	}
}

// adminLoginMethods derives display login methods from bound identity factors.
// adminLoginMethods 根据已绑定的身份因子生成管理端展示用登录方式。
func adminLoginMethods(u user.User, wallets []wallet.UserWallet, accounts []oauth.Account) []string {
	methods := make([]string, 0, 4)
	if u.Email != "" {
		methods = append(methods, "email")
	}
	if u.Phone != "" {
		methods = append(methods, "phone")
	}
	if len(wallets) > 0 {
		methods = append(methods, "wallet")
	}
	for _, account := range accounts {
		methods = append(methods, account.Provider)
	}
	return methods
}

// toAdminUserClientResponse converts a user-client relation to a management DTO.
// toAdminUserClientResponse 将用户-client 关系转换为管理接口 DTO。
func toAdminUserClientResponse(client audit.UserClient) dto.AdminUserClientResponse {
	return dto.AdminUserClientResponse{
		ClientID:     client.ClientID,
		FirstLoginAt: formatTime(client.FirstLoginAt),
		LastLoginAt:  formatTime(client.LastLoginAt),
		LoginCount:   client.LoginCount,
		Status:       client.Status,
	}
}

// toAdminWalletResponse converts a wallet binding to a management DTO.
// toAdminWalletResponse 将钱包绑定转换为管理接口 DTO。
func toAdminWalletResponse(w wallet.UserWallet) dto.AdminWalletResponse {
	return dto.AdminWalletResponse{
		ID:         w.ID,
		ChainType:  string(w.ChainType),
		Address:    w.Address,
		IsPrimary:  w.IsPrimary,
		VerifiedAt: formatTime(w.VerifiedAt),
		CreatedAt:  formatTime(w.CreatedAt),
	}
}

// toAdminWalletResponses converts wallet bindings to management DTOs.
// toAdminWalletResponses 批量转换钱包绑定管理 DTO。
func toAdminWalletResponses(wallets []wallet.UserWallet) []dto.AdminWalletResponse {
	items := make([]dto.AdminWalletResponse, 0, len(wallets))
	for _, w := range wallets {
		items = append(items, toAdminWalletResponse(w))
	}
	return items
}

// toAdminOAuthAccountResponse converts an OAuth account to a management DTO.
// toAdminOAuthAccountResponse 将第三方账号绑定转换为管理接口 DTO。
func toAdminOAuthAccountResponse(account oauth.Account) dto.AdminOAuthAccountResponse {
	return dto.AdminOAuthAccountResponse{
		ID:                account.ID,
		Provider:          account.Provider,
		ProviderSubject:   account.ProviderSubject,
		ProviderEmail:     account.ProviderEmail,
		ProviderUsername:  account.ProviderUsername,
		ProviderAvatarURL: account.ProviderAvatarURL,
		CreatedAt:         formatTime(account.CreatedAt),
	}
}

// toAdminOAuthAccountResponses converts OAuth bindings to management DTOs.
// toAdminOAuthAccountResponses 批量转换第三方账号绑定管理 DTO。
func toAdminOAuthAccountResponses(accounts []oauth.Account) []dto.AdminOAuthAccountResponse {
	items := make([]dto.AdminOAuthAccountResponse, 0, len(accounts))
	for _, account := range accounts {
		items = append(items, toAdminOAuthAccountResponse(account))
	}
	return items
}

// toAdminLoginLogResponse converts a login audit event to a management DTO.
// toAdminLoginLogResponse 将登录审计事件转换为管理接口 DTO。
func toAdminLoginLogResponse(log audit.LoginLog) dto.AdminLoginLogResponse {
	return dto.AdminLoginLogResponse{
		ID:          log.ID,
		UserID:      log.UserID,
		ClientID:    log.ClientID,
		LoginMethod: string(log.LoginMethod),
		IP:          log.IP,
		UserAgent:   log.UserAgent,
		Success:     log.Success,
		FailureCode: log.FailureCode,
		CreatedAt:   formatTime(log.CreatedAt),
	}
}

// toAdminSecurityEventResponse converts a security audit event to a management DTO.
// toAdminSecurityEventResponse 将安全操作审计事件转换为管理接口 DTO。
func toAdminSecurityEventResponse(event audit.SecurityEvent) dto.AdminSecurityEventResponse {
	return dto.AdminSecurityEventResponse{
		ID:          event.ID,
		UserID:      event.UserID,
		EventType:   string(event.EventType),
		TargetType:  event.TargetType,
		TargetID:    event.TargetID,
		IP:          event.IP,
		UserAgent:   event.UserAgent,
		Success:     event.Success,
		FailureCode: event.FailureCode,
		CreatedAt:   formatTime(event.CreatedAt),
	}
}

// toAdminSessionResponse converts a refresh token to a management DTO.
// toAdminSessionResponse 将刷新令牌会话转换为管理接口 DTO。
func toAdminSessionResponse(session token.RefreshToken, currentHash string) dto.AdminSessionResponse {
	return dto.AdminSessionResponse{
		ID:         session.ID,
		UserID:     session.UserID,
		ClientID:   session.ClientID,
		IP:         session.IP,
		UserAgent:  session.UserAgent,
		Active:     !session.IsRevoked() && !session.IsExpired(time.Now().UTC()),
		Current:    currentHash != "" && session.TokenHash == currentHash,
		ExpiresAt:  formatTime(session.ExpiresAt),
		RevokedAt:  formatOptionalTime(session.RevokedAt),
		LastUsedAt: formatOptionalTime(session.LastUsedAt),
		CreatedAt:  formatTime(session.CreatedAt),
	}
}

// currentSessionHash hashes the auth-domain browser session cookie for comparison.
// currentSessionHash 哈希当前浏览器的中台会话 Cookie，用于在管理页标记“当前浏览器”。
func (h *AdminHandler) currentSessionHash(c *gin.Context) string {
	if h.tokenHasher == nil {
		return ""
	}
	sessionToken, ok := readSessionCookie(c)
	if !ok {
		return ""
	}
	return h.tokenHasher.HashToken(sessionToken)
}

// intQuery reads a positive integer query parameter with a fallback.
// intQuery 读取正整数查询参数，缺失或非法时使用默认值。
func intQuery(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

// boolQuery reads a boolean query parameter with a fallback.
// boolQuery 读取布尔查询参数，缺失或非法时使用默认值。
func boolQuery(c *gin.Context, key string, fallback bool) bool {
	raw := c.Query(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}

// formatOptionalTime formats a nullable time pointer for JSON responses.
// formatOptionalTime 格式化可空时间指针用于 JSON 响应。
func formatOptionalTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return formatTime(*value)
}

// formatTime formats time values consistently for HTTP responses.
// formatTime 为 HTTP 响应统一格式化时间。
func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(timeFormatRFC3339)
}
