package seeders

import (
	"log"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func SeedAll() {
	db, err := database.IntialDB() // Ensure this matches your package (InitialDB?)
	if err != nil {
		log.Fatal("Could not connect to DB for seeding")
	}

	// 1. SEED TENANT
	tenant := models.Tenant{
		Name:  "TechCorp Solutions",
		TaxID: "J-12345678-9",
	}
	db.Where(models.Tenant{Name: "TechCorp Solutions"}).FirstOrCreate(&tenant)

	// 2. SEED ROLES
	superRole := models.Role{
		Name: "Superadmin",
	}
	superRole.TenantID = tenant.ID // Setting after init to avoid struct literal issues
	db.Where(models.Role{Name: "Superadmin", TenantID: tenant.ID}).FirstOrCreate(&superRole)

	// 3. SEED MASTER USER
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Warego2026!"), 12)
	adminUser := models.User{
		Username:     "admin",
		Email:        "admin@techcorp.com",
		PasswordHash: string(hashedPassword),
		RoleID:       superRole.ID,
	}
	adminUser.TenantID = tenant.ID
	db.Where(models.User{Email: "admin@techcorp.com"}).FirstOrCreate(&adminUser)

	// 4. SEED WAREHOUSE (Required for locations below)
	warehouse := models.Warehouse{
		Name: "Main DC Warehouse",
	}
	warehouse.TenantID = tenant.ID
	db.Where(models.Warehouse{Name: "Main DC Warehouse", TenantID: tenant.ID}).FirstOrCreate(&warehouse)

	// 5. SEED LOCATIONS
	internalLoc := models.Location{
		WarehouseID:  warehouse.ID,
		Name:         "Shelf A1",
		LocationType: "internal",
	}
	internalLoc.TenantID = tenant.ID
	db.Where(models.Location{Name: "Shelf A1", WarehouseID: warehouse.ID}).FirstOrCreate(&internalLoc)

	inventoryLoc := models.Location{
		WarehouseID:  warehouse.ID,
		Name:         "Initial Inventory Adjustment",
		LocationType: "inventory",
	}
	inventoryLoc.TenantID = tenant.ID
	db.Where(models.Location{Name: "Initial Inventory Adjustment"}).FirstOrCreate(&inventoryLoc)

	// 6. SEED PRODUCT
	template := models.ProductTemplate{
		Name: "HPE ProLiant DL20 Gen10",
		Type: "storable",
	}
	template.TenantID = tenant.ID
	db.Where(models.ProductTemplate{Name: "HPE ProLiant DL20 Gen10", TenantID: tenant.ID}).FirstOrCreate(&template)

	variant := models.ProductVariant{
		TemplateID: template.ID,
		SKU:        "HPE-DL20-G10",
		CostPrice:  1200.50,
		ListPrice:  1500.00,
		Active:     true,
	}
	variant.TenantID = tenant.ID
	db.Where(models.ProductVariant{SKU: "HPE-DL20-G10", TenantID: tenant.ID}).FirstOrCreate(&variant)

	// 7. SEED INITIAL STOCK MOVE
	var moveCount int64
	// Use = here if moveCount was already declared, but since it's new, we use :=
	db.Model(&models.StockMove{}).Where("reference = ? AND tenant_id = ?", "INIT-001", tenant.ID).Count(&moveCount)

	if moveCount == 0 {
		stockMove := models.StockMove{
			VariantID:      variant.ID,
			SrcLocationID:  inventoryLoc.ID,
			DestLocationID: internalLoc.ID,
			Qty:            50,
			UserID:         adminUser.ID,
			Reference:      "INIT-001",
		}
		stockMove.TenantID = tenant.ID
		db.Create(&stockMove)
	}

	log.Println(">>> Seeding completed successfully!")
}
