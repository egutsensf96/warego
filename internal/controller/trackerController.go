package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
)

// GetAuditLogs retrieves the history of actions performed within the tenant's instance
func GetAuditLogs(c *gin.Context) {
	// 1. Get TenantID from the middleware context
	tenantID := c.GetString("tenantID")

	db, err := database.IntialDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}

	var logs []models.Tracker

	// 2. Query with Preload
	// We preload "User" so the UI shows "John Doe" instead of just a UUID
	result := db.Preload("User").
		Where("tenant_id = ?", tenantID).
		Order("created_at desc"). // Show newest actions first
		Limit(100).               // Limit to last 100 entries for performance
		Find(&logs)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit logs"})
		return
	}

	// 3. Logic check for empty logs
	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No activity recorded yet",
			"result":  []models.Tracker{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":  result.RowsAffected,
		"result": logs,
	})
}
