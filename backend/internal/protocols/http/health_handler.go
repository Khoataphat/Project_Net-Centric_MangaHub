package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler returns the current status of the server.
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "up",
		"message": "MangaHub Backend is running",
	})
}
