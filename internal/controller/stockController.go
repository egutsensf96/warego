package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
)

// GetStockLevels retrieves current stock quantities across all warehouses for the tenant
func GetStockLevels(c *gin.Context) {
	// 1. Get TenantID from context (Middleware)
	tenantID := c.GetString("tenantID")

	db, err := database.IntialDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	var stocks []models.Stock

	// 2. Query with Preloads
	// We join Product and Warehouse data so the frontend shows names, not just UUIDs
	result := db.Preload("Product").
		Preload("Warehouse").
		Where("tenant_id = ?", tenantID).
		Find(&stocks)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stock levels"})
		return
	}

	// 3. Logic check: If no stock records exist yet
	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No stock records found for this instance",
			"result":  []models.Stock{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":  result.RowsAffected,
		"result": stocks,
	})
}
