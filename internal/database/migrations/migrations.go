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
	// Optional: Check for a secret 'Migration-Key' header as a final fail-safe
	masterKey := c.GetHeader("X-Migration-Key")
	if masterKey != os.Getenv("MIGRATION_SECRET") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Migration Key"})
		return
	}

	db, err := database.IntialDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// 2. Run AutoMigrate
	// Order: Independent tables (Company/Role) -> Dependent tables (User/Employee) -> Transactional tables
	err = db.AutoMigrate(
		// Core Module
		&models.Company{},
		&models.Role{},
		&models.User{},

		// Inventory Module
		&models.Category{},
		&models.Warehouse{},
		&models.Product{},
		&models.Stock{},
		&models.Draw{}, // The withdrawal records

		// HR Module
		&models.Employee{},
		&models.Contract{},

		// Audit Module
		&models.Tracker{},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Migration failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Database schema synchronized successfully",
	})
}
