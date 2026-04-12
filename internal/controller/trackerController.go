package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
)

// GetAuditLogs retrieves the history of actions performed within the tenant's instance
func GetAuditLogs(c *gin.Context) {
	// 1. Get TenantID from the middleware context (Ensuring it's an int)
	tenantID, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context missing"})
		return
	}

	db, err := database.IntialDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}

	var logs []models.Tracker

	// 2. Query with Preload and Scoping
	// We cast tenantID to (int) to match the database column type
	result := db.Preload("User").
		Where("tenant_id = ?", tenantID.(int)).
		Order("created_at desc"). // Standard ERP practice: Newest first
		Limit(200).               // Increased limit slightly for better visibility
		Find(&logs)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit logs"})
		return
	}

	// 3. Logic check for empty logs
	if len(logs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No activity recorded yet",
			"result":  []models.Tracker{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":  len(logs),
		"result": logs,
	})
}
