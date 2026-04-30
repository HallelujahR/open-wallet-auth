package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/handler"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/middleware"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
)

type Dependencies struct {
	Config           *config.Config
	Logger           *zap.Logger
	Auth             *handler.AuthHandler
	Wallet           *handler.WalletHandler
	Phone            *handler.PhoneHandler
	Email            *handler.EmailHandler
	OAuth            *handler.OAuthHandler
	Client           *handler.ClientHandler
	Admin            *handler.AdminHandler
	Token            middleware.TokenVerifier
	AudienceResolver middleware.ClientAudienceResolver
	JWKS             *handler.JWKSHandler
	AdminToken       string
}

// New creates the HTTP router and registers public and authenticated routes.
// New 创建 HTTP 路由并注册公开、认证和管理接口。
func New(deps Dependencies) *gin.Engine {
	if deps.Config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(middleware.RequestID())
	engine.Use(middleware.CORS(deps.Config.HTTP.CORSAllowedOrigins))
	engine.Use(middleware.Recovery(deps.Logger))
	engine.Use(middleware.AccessLog(deps.Logger))

	healthHandler := handler.NewHealthHandler(deps.Config)

	engine.GET("/healthz", healthHandler.Health)
	engine.GET("/readyz", healthHandler.Ready)
	if deps.JWKS != nil {
		engine.GET("/.well-known/jwks.json", deps.JWKS.JWKS)
	}

	v1 := engine.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Health)
		if deps.Auth != nil {
			auth := v1.Group("/auth")
			{
				auth.POST("/register", deps.Auth.Register)
				auth.POST("/login", deps.Auth.Login)
				auth.POST("/refresh", deps.Auth.Refresh)
				auth.POST("/logout", deps.Auth.Logout)
				auth.POST("/password/reset", deps.Auth.ResetPassword)
				if deps.Token != nil && deps.AudienceResolver != nil {
					authenticated := auth.Group("", middleware.AuthenticateClient(deps.Token, deps.AudienceResolver))
					{
						authenticated.GET("/me", deps.Auth.Me)
						authenticated.GET("/profile", deps.Auth.Profile)
						authenticated.PATCH("/profile", deps.Auth.UpdateProfile)
						authenticated.PATCH("/password", deps.Auth.ChangePassword)
						authenticated.POST("/bind/email", deps.Auth.BindEmail)
						authenticated.POST("/bind/phone", deps.Auth.BindPhone)
						authenticated.DELETE("/bind/email", deps.Auth.UnbindEmail)
						authenticated.DELETE("/bind/phone", deps.Auth.UnbindPhone)
						authenticated.DELETE("/wallets/:wallet_id", deps.Auth.UnbindWallet)
						authenticated.DELETE("/oauth-accounts/:account_id", deps.Auth.UnbindOAuthAccount)
					}
				}
			}
			if deps.Token != nil && deps.AudienceResolver != nil {
				profile := v1.Group("/profile", middleware.AuthenticateClient(deps.Token, deps.AudienceResolver))
				{
					profile.GET("", deps.Auth.Profile)
					profile.PATCH("", deps.Auth.UpdateProfile)
				}
			}
		}
		if deps.Wallet != nil {
			wallet := v1.Group("/wallet")
			{
				wallet.POST("/nonce", deps.Wallet.Nonce)
				wallet.POST("/verify", deps.Wallet.Verify)
				if deps.Token != nil && deps.AudienceResolver != nil {
					wallet.POST("/bind", middleware.AuthenticateClient(deps.Token, deps.AudienceResolver), deps.Wallet.Bind)
				}
			}
		}
		if deps.Phone != nil {
			phone := v1.Group("/phone")
			{
				phone.POST("/code", deps.Phone.Code)
				phone.POST("/login", deps.Phone.Login)
			}
		}
		if deps.Email != nil {
			email := v1.Group("/email")
			{
				email.POST("/code", deps.Email.Code)
				email.POST("/verify", deps.Email.Verify)
			}
		}
		if deps.OAuth != nil {
			oauth := v1.Group("/oauth")
			{
				oauth.GET("/:provider/start", deps.OAuth.Start)
				if deps.Token != nil && deps.AudienceResolver != nil {
					oauth.GET("/:provider/bind/start", middleware.AuthenticateClient(deps.Token, deps.AudienceResolver), deps.OAuth.BindStart)
				}
				oauth.GET("/:provider/callback", deps.OAuth.Callback)
			}
		}
		if deps.Client != nil {
			clients := v1.Group("/clients", middleware.RequireAdminToken(deps.AdminToken))
			{
				clients.POST("", deps.Client.Create)
				clients.GET("", deps.Client.List)
			}
		}
		if deps.Admin != nil {
			admin := v1.Group("/admin", middleware.RequireAdminToken(deps.AdminToken))
			{
				admin.GET("/users", deps.Admin.ListUsers)
				admin.GET("/users/:user_id", deps.Admin.GetUser)
				admin.PATCH("/users/:user_id/status", deps.Admin.UpdateUserStatus)
				admin.DELETE("/users/:user_id/sessions", deps.Admin.RevokeUserSessions)
				admin.DELETE("/users/:user_id/wallets/:wallet_id", deps.Admin.UnbindWallet)
				admin.DELETE("/users/:user_id/oauth-accounts/:account_id", deps.Admin.UnbindOAuthAccount)
				admin.GET("/login-logs", deps.Admin.ListLoginLogs)
				admin.GET("/sessions", deps.Admin.ListSessions)
				admin.DELETE("/sessions/:session_id", deps.Admin.RevokeSession)
				if deps.Client != nil {
					admin.POST("/clients", deps.Client.Create)
					admin.GET("/clients", deps.Client.List)
				}
			}
		}
	}

	return engine
}
