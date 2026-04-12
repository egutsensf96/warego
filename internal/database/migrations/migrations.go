package migrations

import (
	"net/http"
	"os"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
)

// SchemaMigrations handles the GORM AutoMigrate process
func SchemaMigrations(c *gin.Context) {
	// 1. Security Check
	masterKey := c.GetHeader("X-Migration-Key")
	if masterKey == "" || masterKey != os.Getenv("MIGRATION_SECRET") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing Migration Key"})
		return
	}

	db, err := database.IntialDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// 2. Run AutoMigrate in Logical Order
	// We migrate 'Tenants' and 'Roles' first so 'Users' can reference them.
	err = db.AutoMigrate(
		// Core Infrastructure
		&models.Tenant{},
		&models.Role{},
		&models.User{},

		// Inventory Foundation (Odoo Structure)
		&models.Category{},
		&models.Warehouse{},
		&models.Location{}, // New: Supports internal vs virtual locations

		// Product Management
		&models.ProductTemplate{}, // General info + Image
		&models.ProductVariant{},  // SKU + Specific pricing

		// Transactional Data
		&models.StockMove{}, // Replaces 'Draw' and 'Stock' for audit-safe inventory

		// HR & Administration

		&models.Tracker{}, // System-wide audit logs
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Migration failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Database schema synchronized successfully with Double-Entry support",
	})
}
