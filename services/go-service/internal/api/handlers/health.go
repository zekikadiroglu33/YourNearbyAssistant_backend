package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler creates a HealthHandler with a database connection for liveness checks.
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Check handles GET /api/v1/health.
// Pings the database and returns 200 if healthy, 503 if the DB is unreachable.
func (h *HealthHandler) Check(c *gin.Context) {
	if err := h.db.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"service": "go-service",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "go-service",
	})
}
