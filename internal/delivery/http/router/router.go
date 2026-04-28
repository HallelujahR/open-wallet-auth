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

	v1 := engine.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Health)
		if deps.Auth != nil {
			auth := v1.Group("/auth")
			{
				auth.POST("/register", deps.Auth.Register)
				auth.POST("/login", deps.Auth.Login)
			}
		}
	}

	return engine
}
