package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"stk-backend/internal/config"
	"stk-backend/internal/http/handlers"
	"stk-backend/internal/menu"
)

func NewRouter(cfg config.Config, db *gorm.DB) *gin.Engine {
	setGinMode(cfg.App.Env)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), corsMiddleware(cfg.App.CORSAllowOrigin))
	router.StaticFile("/openapi.yaml", "./docs/openapi.yaml")

	healthHandler := handlers.NewHealthHandler(db)
	menuRepository := menu.NewRepository(db)
	menuService := menu.NewService(menuRepository)
	menuHandler := handlers.NewMenuHandler(menuService)

	api := router.Group("/api")
	{
		api.GET("/health", healthHandler.Ping)

		menus := api.Group("/menus")
		{
			menus.GET("", menuHandler.GetAll)
			menus.GET("/:id", menuHandler.GetByID)
			menus.POST("", menuHandler.Create)
			menus.PUT("/:id", menuHandler.Update)
			menus.DELETE("/:id", menuHandler.Delete)
			menus.PATCH("/:id/move", menuHandler.Move)
			menus.PATCH("/:id/reorder", menuHandler.Reorder)
		}
	}

	return router
}

func corsMiddleware(allowOrigin string) gin.HandlerFunc {
	if allowOrigin == "" {
		allowOrigin = "http://localhost:3000"
	}

	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Vary", "Origin")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func setGinMode(environment string) {
	switch environment {
	case "production":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}
