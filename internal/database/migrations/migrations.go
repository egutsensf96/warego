package migrations

import (
	"log"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models" // Import your models here
)

// SchemaMigrations handles the GORM AutoMigrate process
func SchemaMigrations() {

	db := database.InitDB()
	if db == nil {
		log.Fatal("🔴 Failed to initialize database")
	}

	// Run AutoMigrate for all models
	if err := db.AutoMigrate(
		&models.Tenant{},
		&models.Role{},
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.Supplier{},
		&models.Warehouse{},
		&models.Tracker{},
		&models.StockTransaction{},
		&models.AuditLog{},
	); err != nil {
		log.Fatalf("🔴 AutoMigrate failed: %v", err)
	}
	log.Println("✅ Database tables initialized")
}
