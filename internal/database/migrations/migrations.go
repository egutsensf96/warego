package migrations

import (
	"log"
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models" // Import your models here
	"github.com/gin-gonic/gin"
)

// SchemaMigrations handles the GORM AutoMigrate process
func SchemaMigrations(c *gin.Context) {
	db := database.GetDB()

	log.Println("🔄 Starting schema sync...")

	// Automigrate using the structs from the models package
	err := db.AutoMigrate(
		&models.Tenant{},
		&models.Role{},
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.Warehouse{},
		&models.Tracker{},
		&models.Draw{},
	)

	if err != nil {
		log.Printf("❌ Migration failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to sync database schema",
		})
		return
	}

	log.Println("✅ Database schema is up to date.")
	c.JSON(http.StatusOK, gin.H{
		"message": "Schema synchronized successfully",
	})
}
