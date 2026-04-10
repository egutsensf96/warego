package seeders

import (
	"log"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func SeedAll() {
	db, err := database.IntialDB()
	if err != nil {
		log.Fatal("Could not connect to DB for seeding")
	}

	// --- 1. SEED COMPANY (TENANT) ---
	tenantID := uuid.New()
	company := models.Company{
		Base:  models.Base{ID: tenantID, TenantID: tenantID},
		Name:  "TechCorp Solutions",
		TaxID: "J-12345678-9",
	}
	db.FirstOrCreate(&company, models.Company{Name: "TechCorp Solutions"})

	// --- 2. SEED ROLES ---
	superRole := models.Role{
		Base:        models.Base{TenantID: tenantID},
		Description: "Superadmin",
	}
	userRole := models.Role{
		Base:        models.Base{TenantID: tenantID},
		Description: "Warehouse Operator",
	}
	db.FirstOrCreate(&superRole, models.Role{Description: "Superadmin", Base: models.Base{TenantID: tenantID}})
	db.FirstOrCreate(&userRole, models.Role{Description: "Warehouse Operator", Base: models.Base{TenantID: tenantID}})

	// --- 3. SEED MASTER USER ---
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Warego2026!"), 12)
	adminUser := models.User{
		Base:      models.Base{TenantID: tenantID},
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@techcorp.com",
		Password:  string(hashedPassword),
		RoleID:    superRole.ID,
	}
	db.Where("email = ?", "admin@techcorp.com").FirstOrCreate(&adminUser)

	// --- 4. SEED INVENTORY CATEGORIES ---
	catHardware := models.Category{
		Base:        models.Base{TenantID: tenantID},
		Description: "Server Hardware",
	}
	db.FirstOrCreate(&catHardware, models.Category{Description: "Server Hardware", Base: models.Base{TenantID: tenantID}})

	// --- 5. SEED WAREHOUSE ---
	warehouse := models.Warehouse{
		Base:     models.Base{TenantID: tenantID},
		Name:     "Main DC Warehouse",
		Location: "Building A - Sector 4",
	}
	db.FirstOrCreate(&warehouse, models.Warehouse{Name: "Main DC Warehouse", Base: models.Base{TenantID: tenantID}})

	// --- 6. SEED PRODUCTS ---
	product := models.Product{
		Base:        models.Base{TenantID: tenantID},
		SKU:         "HPE-DL20-G10",
		Description: "HPE ProLiant DL20 Gen10 Server",
		Cost:        1200.50,
		CategoryID:  catHardware.ID,
		UserID:      adminUser.ID,
	}
	db.Where("sku = ?", "HPE-DL20-G10").FirstOrCreate(&product)

	// --- 7. SEED INITIAL STOCK ---
	stock := models.Stock{
		Base:        models.Base{TenantID: tenantID},
		ProductID:   product.ID,
		WarehouseID: warehouse.ID,
		Quantity:    25,
	}
	db.Where("product_id = ? AND warehouse_id = ?", product.ID, warehouse.ID).FirstOrCreate(&stock)

	log.Println(">>> Seeding completed successfully!")
	log.Printf(">>> Tenant ID: %s", tenantID)
	log.Println(">>> Credentials: admin@techcorp.com / Warego2026!")
}
