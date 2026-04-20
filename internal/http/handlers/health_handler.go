package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"stk-backend/internal/response"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Ping(c *gin.Context) {
	databaseStatus := "up"
	status := "ok"
	statusCode := http.StatusOK

	if h.db == nil {
		databaseStatus = "not_configured"
	} else {
		sqlDB, err := h.db.DB()
		if err != nil {
			databaseStatus = "down"
			status = "degraded"
			statusCode = http.StatusServiceUnavailable
		} else {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()

			if err := sqlDB.PingContext(ctx); err != nil {
				databaseStatus = "down"
				status = "degraded"
				statusCode = http.StatusServiceUnavailable
			}
		}
	}

	response.JSONSuccess(c, statusCode, gin.H{
		"service":   "stk-backend",
		"status":    status,
		"database":  databaseStatus,
		"timestamp": time.Now().UTC(),
	}, "service health status")
}
