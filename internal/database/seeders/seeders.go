// internal/database/seeders/seeders.go
package seeders

import (
	"errors"
	"log"
	"strings"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedAll creates all necessary data for the application.
// It is safe to run multiple times (checks for existing records).
func SeedAll() {
	db := database.GetDB()
	if db == nil {
		log.Fatal("🔴 Database connection is nil. Ensure DB is initialized first.")
	}

	log.Println("🌱 Starting database seeding...")

	// 1. Seed Tenant
	tenant := seedTenant(db, "DemoCorp", "democorp.local")
	if tenant == nil {
		log.Fatal("❌ Failed to seed tenant - cannot continue")
	}

	// 2. Seed Roles
	// Create a Global SuperAdmin role
	superAdminRole := seedGlobalRole(db, "SuperAdmin", "Global System Administrator")
	if superAdminRole == nil {
		log.Fatal("❌ Failed to seed SuperAdmin role - cannot continue")
	}

	// Create Tenant-specific roles
	adminRole := seedTenantRole(db, tenant.ID, "Admin", "Tenant Administrator")
	if adminRole == nil {
		log.Fatal("❌ Failed to seed Admin role - cannot continue")
	}

	staffRole := seedTenantRole(db, tenant.ID, "Staff", "Inventory Staff")
	if staffRole == nil {
		log.Fatal("❌ Failed to seed Staff role - cannot continue")
	}

	// 3. Seed Users
	// Password for all seed users: password123
	hashedPassword := hashPassword("password123")

	// SuperAdmin User (assigned to DemoCorp for this seed)
	_ = seedUser(db, "superadmin@warego.com", hashedPassword, tenant.ID, superAdminRole.ID, "Super Admin", true)

	// Admin User
	_ = seedUser(db, "admin@democorp.com", hashedPassword, tenant.ID, adminRole.ID, "John Admin", true)

	// Staff User
	_ = seedUser(db, "staff@democorp.com", hashedPassword, tenant.ID, staffRole.ID, "Jane Staff", true)

	// 4. Seed Suppliers
	_ = seedSupplier(db, tenant.ID, "TechParts Global", "orders@techparts.com", "+1-555-0101")
	_ = seedSupplier(db, tenant.ID, "Office Supplies Co.", "sales@officesupplies.com", "+1-555-0102")

	// 5. Seed Categories
	catElectronics := seedCategory(db, tenant.ID, "Electronics", "Computers, gadgets, and components")
	if catElectronics == nil {
		log.Fatal("❌ Failed to seed Electronics category - cannot continue")
	}
	_ = seedCategory(db, tenant.ID, "Office Supplies", "Paper, pens, and furniture")

	// 6. Seed Warehouses
	whMain := seedWarehouse(db, tenant.ID, "Main Warehouse", "Building A - Floor 1")
	if whMain == nil {
		log.Fatal("❌ Failed to seed Main Warehouse - cannot continue")
	}
	_ = seedWarehouse(db, tenant.ID, "Overflow Storage", "Building B - Basement")

	// 7. Seed Products
	laptop := seedProduct(db, tenant.ID, "MacBook Pro 16\"", "LAP-MBP-16", 50, catElectronics.ID)
	mouse := seedProduct(db, tenant.ID, "Logitech MX Master", "MOU-LOG-MX", 100, catElectronics.ID)
	_ = seedProduct(db, tenant.ID, "A4 Copy Paper (Box)", "PAP-A4-BOX", 200, catElectronics.ID)

	// 8. Seed Stock Trackers (Current Inventory Levels)
	seedTracker(db, tenant.ID, laptop.ID, whMain.ID, 40)
	seedTracker(db, tenant.ID, mouse.ID, whMain.ID, 100)

	// 9. Seed Stock Transactions (Audit Trail)
	// Record the initial stock entry using the SuperAdmin user
	superAdminUser := seedUser(db, "audit@warego.com", hashedPassword, tenant.ID, superAdminRole.ID, "Audit Bot", true)
	if superAdminUser != nil {
		seedTransaction(db, tenant.ID, laptop.ID, whMain.ID, superAdminUser.ID, models.TypeInitial, 40, "Initial stock entry")
		seedTransaction(db, tenant.ID, mouse.ID, whMain.ID, superAdminUser.ID, models.TypeInitial, 100, "Initial stock entry")
	}

	// 10. Seed Draws (Stock Withdrawals)
	seedDraw(db, tenant.ID, mouse.ID, 5, "Q1 Employee Giveaway", "pending")
	seedDraw(db, tenant.ID, laptop.ID, 2, "Executive Allocation", "pending")

	log.Println("✅ Database seeding completed successfully!")
	log.Printf("   📦 Tenant: %s\n   👤 SuperAdmin: superadmin@warego.com\n   👤 Admin: admin@democorp.com\n   🔑 Password: password123", tenant.Name)
}

// ================= HELPER FUNCTIONS =================

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	return string(bytes)
}

func seedTenant(db *gorm.DB, name, domain string) *models.Tenant {
	var tenant models.Tenant
	if err := db.Where("domain = ?", strings.ToLower(domain)).First(&tenant).Error; err != nil {
		tenant = models.Tenant{
			Name:     name,
			Domain:   strings.ToLower(domain),
			IsActive: true,
		}
		if err := db.Create(&tenant).Error; err != nil {
			log.Printf("⚠️ Failed to create tenant %s: %v", name, err)
			return nil
		}
		log.Printf("✅ Created Tenant: %s", name)
	}
	return &tenant
}

func seedGlobalRole(db *gorm.DB, name, desc string) *models.Role {
	var role models.Role
	// Check if exists (Global roles are not tied to tenant)
	if err := db.Where("name = ? AND is_global = ?", name, true).First(&role).Error; err != nil {
		role = models.Role{
			Name:        name,
			Description: desc,
			IsGlobal:    true,
			// ✅ FIX: Don't set TenantID for global roles (leave as nil)
		}
		if err := db.Create(&role).Error; err != nil {
			log.Printf("⚠️ Failed to create global role %s: %v", name, err)
			return nil
		}
		log.Printf("✅ Created Global Role: %s", name)
	}
	return &role
}
func seedTenantRole(db *gorm.DB, tenantID uuid.UUID, name, desc string) *models.Role {
	var role models.Role
	if err := db.Where("name = ? AND tenant_id = ?", name, tenantID).First(&role).Error; err != nil {
		role = models.Role{
			Name:        name,
			Description: desc,
			TenantID:    &tenantID,
			IsGlobal:    false,
		}
		if err := db.Create(&role).Error; err != nil {
			log.Printf("⚠️ Failed to create role %s: %v", name, err)
			return nil
		}
		log.Printf("✅ Created Role: %s", name)
	}
	return &role
}

func seedUser(db *gorm.DB, email, password string, tenantID, roleID uuid.UUID, name string, isActive bool) *models.User {
	var user models.User
	if err := db.Where("email = ?", strings.ToLower(email)).First(&user).Error; err != nil {
		user = models.User{
			Name:     name,
			Email:    strings.ToLower(email),
			Password: password,
			RoleID:   roleID,
			TenantID: tenantID,
			IsActive: isActive,
		}
		if err := db.Create(&user).Error; err != nil {
			log.Printf("⚠️ Failed to create user %s: %v", email, err)
			return nil
		}
		log.Printf("✅ Created User: %s", email)
	}
	return &user
}

func seedSupplier(db *gorm.DB, tenantID uuid.UUID, name, email, phone string) *models.Supplier {
	var supplier models.Supplier

	// 1. Limpiar y preparar el email
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	var emailPtr string
	if cleanEmail != "" {
		emailPtr = cleanEmail
	} else {
		emailPtr = "" // O podrías dejarlo como nil si tu modelo lo permite
	}

	// 2. Definir criterio de búsqueda
	// Si hay email, buscamos por email. Si no, buscamos por nombre (dentro del mismo tenant)
	query := db.Where("tenant_id = ?", tenantID)
	if emailPtr != "" {
		query = query.Where("email = ?", emailPtr)
	} else {
		query = query.Where("name = ?", name)
	}

	// 3. Buscar o Crear
	err := query.First(&supplier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			supplier = models.Supplier{
				Name:     name,
				Email:    emailPtr, // Asignamos el puntero (*string)
				Phone:    phone,
				TenantID: tenantID,
			}
			if err := db.Create(&supplier).Error; err != nil {
				log.Printf("❌ Error al crear Supplier %s: %v", name, err)
				return nil
			}
			log.Printf("✅ Created Supplier: %s", name)
		} else {
			log.Printf("❌ Error al buscar Supplier %s: %v", name, err)
			return nil
		}
	}

	return &supplier
}

func seedCategory(db *gorm.DB, tenantID uuid.UUID, name, desc string) *models.Category {
	var category models.Category
	if err := db.Where("name = ? AND tenant_id = ?", name, tenantID).First(&category).Error; err != nil {
		category = models.Category{
			Name:        name,
			Description: desc,
			TenantID:    tenantID,
		}
		if err := db.Create(&category).Error; err != nil {
			log.Printf("⚠️ Failed to create category %s: %v", name, err)
			return nil
		}
		log.Printf("✅ Created Category: %s", name)
	}
	return &category
}

func seedWarehouse(db *gorm.DB, tenantID uuid.UUID, name, location string) *models.Warehouse {
	var warehouse models.Warehouse
	if err := db.Where("name = ? AND tenant_id = ?", name, tenantID).First(&warehouse).Error; err != nil {
		warehouse = models.Warehouse{
			Name:     name,
			Location: location,
			TenantID: tenantID,
		}
		if err := db.Create(&warehouse).Error; err != nil {
			log.Printf("⚠️ Failed to create warehouse %s: %v", name, err)
			return nil
		}
		log.Printf("✅ Created Warehouse: %s", name)
	}
	return &warehouse
}

func seedProduct(db *gorm.DB, tenantID uuid.UUID, name, sku string, qty int, categoryID uuid.UUID) *models.Product {
	var product models.Product
	if err := db.Where("sku = ?", strings.ToUpper(sku)).First(&product).Error; err != nil {
		product = models.Product{
			Name:       name,
			SKU:        strings.ToUpper(sku),
			Quantity:   qty,
			CategoryID: categoryID,
			TenantID:   tenantID,
		}
		if err := db.Create(&product).Error; err != nil {
			log.Printf("⚠️ Failed to create product %s: %v", name, err)
			return nil
		}
		log.Printf("✅ Created Product: %s", name)
	}
	return &product
}

func seedTracker(db *gorm.DB, tenantID, productID, warehouseID uuid.UUID, quantity int) {
	var tracker models.Tracker
	if err := db.Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).First(&tracker).Error; err != nil {
		tracker = models.Tracker{
			ProductID:   productID,
			WarehouseID: warehouseID,
			Quantity:    quantity,
			TenantID:    tenantID,
		}
		db.Create(&tracker)
	}
}

func seedTransaction(db *gorm.DB, tenantID, productID, warehouseID, userID uuid.UUID, txType models.StockTransactionType, qtyChange int, notes string) {
	txn := models.StockTransaction{
		ProductID:      productID,
		WarehouseID:    warehouseID,
		TenantID:       tenantID,
		QuantityChange: qtyChange,
		Type:           txType,
		UserID:         &userID,
		Notes:          notes,
	}
	db.Create(&txn)
}

func seedDraw(db *gorm.DB, tenantID, productID uuid.UUID, quantity int, name, status string) {
	draw := models.Draw{
		Name:      name,
		ProductID: productID,
		Quantity:  quantity,
		TenantID:  tenantID,
		Status:    status,
	}
	db.Create(&draw)
	log.Printf("✅ Created Draw: %s", name)
}
