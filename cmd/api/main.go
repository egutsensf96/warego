// cmd/api/main.go
package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/egutsenf96/warego/internal/controller"
	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/database/seeders"
	"github.com/egutsenf96/warego/internal/middleware"
	"github.com/egutsenf96/warego/internal/models"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ .env file not found, using system env vars")
	}

	// Set Gin mode
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)

	// Initialize database
	db := database.InitDB()
	if db == nil {
		log.Fatal("🔴 Failed to initialize database")
	}

	// ✅ Run AutoMigrate for ALL models
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
		&models.Draw{},
		&models.AuditLog{},
	); err != nil {
		log.Fatalf("🔴 AutoMigrate failed: %v", err)
	}
	log.Println("✅ Database tables initialized")

	// Run seeders if enabled
	if os.Getenv("SEED_DB") == "true" {
		log.Println("🌱 Running database seeders...")
		seeders.SeedAll()
	}

	// Setup Gin router
	r := gin.Default()

	// ✅ Middleware stack
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     getAllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Tenant-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ==================== 🌐 PUBLIC ROUTES (No Auth Required) ====================
	public := r.Group("/api/v1")
	{
		// Health check
		public.GET("/health", healthHandler)

		// 🔐 Authentication
		auth := public.Group("/auth")
		{
			auth.POST("/login", controller.SignIn)          // User login → JWT
			auth.POST("/register", controller.SignUp)       // Register user in existing tenant
			auth.POST("/onboard", controller.CreateCompany) // Create new tenant + admin user
			auth.GET("/check", controller.CheckAuth)        // Verify token validity (optional)
		}
	}

	// ==================== 🔐 PROTECTED ROUTES (JWT Required) ====================
	api := r.Group("/api/v1")
	api.Use(middleware.JwtValidate()) // ✅ Validate JWT on all protected routes
	{
		// ==================== 🌍 SUPERADMIN ROUTES (Global Access) ====================
		superAdmin := api.Group("/superadmin")
		superAdmin.Use(middleware.RequireRole("super-admin")) // ✅ Only SuperAdmin can access
		{
			// 🔹 Tenant Management (Global)
			superAdmin.GET("/tenants", controller.GetTenants)
			superAdmin.GET("/tenants/:id", controller.GetTenant)
			superAdmin.POST("/tenants", controller.CreateTenant)
			superAdmin.PUT("/tenants/:id", controller.UpdateTenant)
			superAdmin.DELETE("/tenants/:id", controller.DeleteTenant)

			// 🔹 Global User Management
			superAdmin.GET("/users", controller.GetUsers)          // List ALL users across ALL tenants
			superAdmin.GET("/users/:id", controller.GetUser)       // Get any user by ID
			superAdmin.PUT("/users/:id", controller.UpdateUser)    // Update any user
			superAdmin.DELETE("/users/:id", controller.DeleteUser) // Delete any user (soft/hard)

			// 🔹 Global Audit Logs (All user actions across all tenants)
			superAdmin.GET("/audit-logs", controller.GetAuditLogsGlobal) // Fetch ALL audit logs
		}

		// ==================== 🏢 TENANT-SCOPED ADMIN ROUTES ====================
		admin := api.Group("/admin")
		admin.Use(middleware.RequireTenant())                     // ✅ Ensure tenant_id in context
		admin.Use(middleware.RequireRole("admin", "super-admin")) // ✅ Admin + SuperAdmin only
		{
			// 🔹 User Management (Tenant-Scoped)
			admin.GET("/users", controller.GetUsers)
			admin.GET("/users/:id", controller.GetUser)
			admin.POST("/users", controller.CreateUser)
			admin.PUT("/users/:id", controller.UpdateUser)
			admin.DELETE("/users/:id", controller.DeleteUser)

			// 🔹 Role Management
			admin.GET("/roles", controller.GetRoles)
			admin.POST("/roles", controller.AddRole)
			admin.PUT("/roles/:id", controller.UpdateRole)
			admin.DELETE("/roles/:id", controller.DeleteRole)

			// 🔹 Warehouse Management
			admin.GET("/warehouses", controller.GetWarehouses)
			admin.POST("/warehouses", controller.CreateWarehouse)
			admin.PUT("/warehouses/:id", controller.UpdateWarehouse)
			admin.DELETE("/warehouses/:id", controller.DeleteWarehouse)

			// 🔹 Audit Trail / Stock Transactions (Tenant-Scoped)
			admin.GET("/tracker", controller.GetAuditLogs) // Fetch StockTransaction logs for tenant
		}

		// ==================== 📦 INVENTORY MODULE (Tenant-Scoped) ====================
		inventory := api.Group("/inventory")
		inventory.Use(middleware.RequireTenant()) // ✅ Tenant isolation
		{
			// 🔹 Products CRUD
			inventory.GET("/products", controller.GetProducts)
			inventory.POST("/products", controller.CreateProduct)
			inventory.PUT("/products/:id", controller.UpdateProductWithAudit)
			inventory.DELETE("/products/:id", controller.DeleteProduct)

			// 🔹 Categories CRUD
			inventory.GET("/categories", controller.GetCategories)
			inventory.POST("/categories", controller.AddCategory)
			inventory.PUT("/categories/:id", controller.UpdateCategory)
			inventory.DELETE("/categories/:id", controller.DeleteCategory)

			// 🔹 Stock Levels & Movements
			inventory.GET("/stock", controller.GetStockLevels)      // Current stock per warehouse
			inventory.GET("/transactions", controller.GetAuditLogs) // Alias for /admin/tracker

			// 🔹 Draws / Giveaways (Legacy Flow)
			inventory.GET("/draws", controller.GetDraws)
			inventory.POST("/draws", controller.ProcessStockDraw)
			inventory.POST("/draws/:id/retrieve", controller.RetireProduct) // Mark draw as completed
		}

		// ==================== 🤝 SUPPLIERS MODULE (Tenant-Scoped) ====================
		suppliers := api.Group("/suppliers")
		suppliers.Use(middleware.RequireTenant())
		{
			suppliers.GET("/", controller.GetSuppliers)
			suppliers.POST("/", controller.CreateSupplier)
			suppliers.PUT("/:id", controller.UpdateSupplier)
			suppliers.DELETE("/:id", controller.DeleteSupplier)
		}

		// ==================== 🔄 SYNC MODULE (Offline-First Support) ====================
		sync := api.Group("/sync")
		sync.Use(middleware.RequireTenant())
		{
			sync.POST("/push", controller.PushSync) // Send local changes to server
			sync.GET("/pull", controller.PullSync)  // Fetch server changes since timestamp
		}

		// ==================== 👤 USER PROFILE ROUTES (All Authenticated Users) ====================
		user := api.Group("/user")
		{
			user.GET("/profile", controller.GetUserProfile)    // Get my profile + permissions
			user.PUT("/profile", controller.UpdateUserProfile) // Update my profile
			user.POST("/logout", controller.Logout)            // Client-side logout hint
		}

		// ==================== 🏢 COMPANY PROFILE (Tenant-Scoped) ====================
		company := api.Group("/company")
		company.Use(middleware.RequireTenant())
		{
			company.GET("/profile", controller.GetCompanyProfile) // Get current tenant details
			company.PUT("/profile", controller.UpdateCompany)     // Update tenant name/domain
		}
	}

	// ==================== 🚀 START SERVER ====================
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s (%s mode)", port, ginMode)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("🔴 Server failed to start: %v", err)
	}
}

// ==================== 🛠️ HELPER FUNCTIONS ====================

// healthHandler returns service status
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"service":   "warego-erp",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC(),
		"database":  checkDatabaseHealth(),
	})
}

// checkDatabaseHealth verifies DB connection
func checkDatabaseHealth() string {
	db := database.GetDB()
	if db == nil {
		return "disconnected"
	}
	sqlDB, err := db.DB()
	if err != nil {
		return "error"
	}
	if err := sqlDB.Ping(); err != nil {
		return "unhealthy"
	}
	return "connected"
}

// getAllowedOrigins returns CORS allowed origins from env
func getAllowedOrigins() []string {
	origins := os.Getenv("CORS_ORIGINS")
	if origins == "" {
		// Default for development
		return []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://192.168.1.100:3000", // Physical device testing
		}
	}
	return parseCommaSeparated(origins)
}

// parseCommaSeparated splits comma-separated string into slice
func parseCommaSeparated(s string) []string {
	var result []string
	for _, item := range strings.Split(s, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}
