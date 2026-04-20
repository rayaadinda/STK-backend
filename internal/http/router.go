package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"stk-backend/internal/config"
	"stk-backend/internal/http/handlers"
	"stk-backend/internal/response"
)

func NewRouter(cfg config.Config, db *gorm.DB) *gin.Engine {
	setGinMode(cfg.App.Env)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	healthHandler := handlers.NewHealthHandler(db)

	api := router.Group("/api")
	{
		api.GET("/health", healthHandler.Ping)
		registerMenuPlaceholders(api)
	}

	return router
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

func registerMenuPlaceholders(api *gin.RouterGroup) {
	menus := api.Group("/menus")
	{
		menus.GET("", notImplemented)
		menus.GET("/:id", notImplemented)
		menus.POST("", notImplemented)
		menus.PUT("/:id", notImplemented)
		menus.DELETE("/:id", notImplemented)
		menus.PATCH("/:id/move", notImplemented)
		menus.PATCH("/:id/reorder", notImplemented)
	}
}

func notImplemented(c *gin.Context) {
	response.JSONError(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "endpoint is not implemented yet", nil)
}
