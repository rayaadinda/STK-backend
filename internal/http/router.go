package httpserver

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"stk-backend/internal/config"
	"stk-backend/internal/http/handlers"
	"stk-backend/internal/menu"
)

func NewRouter(cfg config.Config, db *gorm.DB) *gin.Engine {
	setGinMode(cfg.App.Env)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), docsEntryMiddleware(), corsMiddleware(cfg.App.CORSAllowOrigin))
	router.StaticFile("/openapi.yaml", "./docs/openapi.yaml")

	healthHandler := handlers.NewHealthHandler(db)
	menuRepository := menu.NewRepository(db)
	menuService := menu.NewService(menuRepository)
	menuHandler := handlers.NewMenuHandler(menuService)
	swaggerHandler := ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/openapi.yaml"),
	)

	api := router.Group("/api")
	{
		api.GET("/health", healthHandler.Ping)
		api.GET("/docs/*any", swaggerHandler)

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

func docsEntryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		if c.Request.Method == http.MethodHead && strings.HasPrefix(path, "/api/docs") {
			c.Status(http.StatusOK)
			c.Abort()
			return
		}

		if path == "/api/docs" || path == "/api/docs/" {
			c.Redirect(http.StatusTemporaryRedirect, "/api/docs/index.html")
			c.Abort()
			return
		}

		c.Next()
	}
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
