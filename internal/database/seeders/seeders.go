package seeders

import (
	"log"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func SeedAll() {
	db := database.GetDB()
	log.Println("🌱 Seeding database with test data...")

	// 1. Create a Tenant
	tenant := models.Tenant{
		Name:   "Acme Corp ERP",
		Domain: "acme.warego.io",
	}
	db.FirstOrCreate(&tenant, models.Tenant{Domain: "acme.warego.io"})

	// 2. Create Roles
	adminRole := models.Role{Name: "Admin", Description: "Full system access"}
	staffRole := models.Role{Name: "Staff", Description: "Inventory management only"}
	db.FirstOrCreate(&adminRole, models.Role{Name: "Admin"})
	db.FirstOrCreate(&staffRole, models.Role{Name: "Staff"})

	// 3. Create an Admin User
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	adminUser := models.User{
		Email:    "admin@acme.com",
		Password: string(hash),
		TenantID: tenant.ID,
		RoleID:   adminRole.ID,
	}
	db.FirstOrCreate(&adminUser, models.User{Email: "admin@acme.com"})

	// 4. Create a Category
	category := models.Category{
		Name:     "Electronics",
		TenantID: tenant.ID,
	}
	db.FirstOrCreate(&category, models.Category{Name: "Electronics", TenantID: tenant.ID})

	// 5. Create a Product
	product := models.Product{
		Name:       "Industrial Sensor X1",
		SKU:        "SNSR-001",
		CategoryID: category.ID,
		TenantID:   tenant.ID,
	}
	db.FirstOrCreate(&product, models.Product{SKU: "SNSR-001", TenantID: tenant.ID})

	// 6. Create a Warehouse
	warehouse := models.Warehouse{
		Name:     "Main Distribution Center",
		Location: "North Sector",
		TenantID: tenant.ID,
	}
	db.FirstOrCreate(&warehouse, models.Warehouse{Name: "Main Distribution Center", TenantID: tenant.ID})

	// 7. Initialize Stock (Tracker)
	tracker := models.Tracker{
		ProductID:   product.ID,
		WarehouseID: warehouse.ID,
		Quantity:    150,
		TenantID:    tenant.ID,
	}
	db.FirstOrCreate(&tracker, models.Tracker{ProductID: product.ID, WarehouseID: warehouse.ID})

	log.Println("✅ Seeding finished successfully!")
}
