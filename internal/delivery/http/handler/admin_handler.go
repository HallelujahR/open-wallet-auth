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
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/wallet"
	adminusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/admin"
)

// AdminHandler exposes identity-management APIs for internal operations.
// AdminHandler 暴露公司内部使用的统一身份管理接口。
type AdminHandler struct {
	admin *adminusecase.Service
}

// NewAdminHandler creates an AdminHandler bound to the admin usecase service.
// NewAdminHandler 创建绑定身份管理用例服务的 HTTP handler。
func NewAdminHandler(admin *adminusecase.Service) *AdminHandler {
	return &AdminHandler{admin: admin}
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
		items = append(items, toAdminUserResponse(u))
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

// writeAdminError maps admin usecase errors to HTTP responses.
// writeAdminError 将身份管理用例错误映射为 HTTP 响应。
func writeAdminError(c *gin.Context, err error) {
	var appErr *domain.Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case adminusecase.ErrUserNotFound:
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
	wallets := make([]dto.AdminWalletResponse, 0, len(result.Wallets))
	for _, w := range result.Wallets {
		wallets = append(wallets, toAdminWalletResponse(w))
	}
	accounts := make([]dto.AdminOAuthAccountResponse, 0, len(result.Accounts))
	for _, account := range result.Accounts {
		accounts = append(accounts, toAdminOAuthAccountResponse(account))
	}
	return dto.AdminUserDetailResponse{
		User:     toAdminUserResponse(result.User),
		Clients:  clients,
		Wallets:  wallets,
		Accounts: accounts,
	}
}

// toAdminUserResponse converts a domain user to a management DTO.
// toAdminUserResponse 将领域用户转换为管理接口 DTO。
func toAdminUserResponse(u user.User) dto.AdminUserResponse {
	return dto.AdminUserResponse{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		Phone:       u.Phone,
		Avatar:      u.Avatar,
		Status:      string(u.Status),
		LastLoginAt: formatOptionalTime(u.LastLoginAt),
		CreatedAt:   formatTime(u.CreatedAt),
		UpdatedAt:   formatTime(u.UpdatedAt),
	}
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

// intQuery reads a positive integer query parameter with a fallback.
// intQuery 读取正整数查询参数，缺失或非法时使用默认值。
func intQuery(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil || value <= 0 {
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
