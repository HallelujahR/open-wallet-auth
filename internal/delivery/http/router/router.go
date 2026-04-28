package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/handler"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/middleware"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
)

type Dependencies struct {
	Config *config.Config
	Logger *zap.Logger
	Auth   *handler.AuthHandler
	Token  middleware.TokenVerifier
	JWKS   *handler.JWKSHandler
}

func New(deps Dependencies) *gin.Engine {
	if deps.Config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(middleware.RequestID())
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
				if deps.Token != nil {
					auth.GET("/me", middleware.Authenticate(deps.Token, "default"), deps.Auth.Me)
				}
			}
		}
	}

	return engine
}
