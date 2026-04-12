package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/egutsenf96/warego/internal/controller"
	"github.com/egutsenf96/warego/internal/database/migrations"
	"github.com/egutsenf96/warego/internal/database/seeders"
	"github.com/egutsenf96/warego/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system env")
	}

	r := gin.Default()

	// 1. GLOBAL MIDDLEWARE
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(cors.Default())

	// Security Headers
	r.Use(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	})

	// 2. PUBLIC ROUTES (Authentication & Onboarding)
	auth := r.Group("/auth")
	{
		auth.POST("/login", controller.SingIn)
		auth.POST("/register", controller.SignUp)
		auth.POST("/company/onboard", controller.CreateCompany) // Create new Tenant instance
	}

	// 3. PRIVATE ERP ROUTES (Requires Valid JWT)
	api := r.Group("/api/v1")
	api.Use(middleware.JwtValidate)
	api.Use(TenantContextGuard()) // Ensures the tenantID is locked and valid
	{
		// --- INVENTORY MODULE ---
		inventory := api.Group("/inventory")
		{
			// Products & Templates
			inventory.GET("/products", controller.GetProducts)
			inventory.POST("/products", controller.CreateProduct)
			inventory.PUT("/products/:id", controller.UpdateProduct)
			inventory.DELETE("/products/:id", controller.DeleteProduct)

			// Categories
			inventory.GET("/categories", controller.GetCategories)
			inventory.POST("/categories", controller.AddCategory)
			inventory.PUT("/categories/:id", controller.UpdateCategory)
			inventory.DELETE("/categories/:id", controller.DeleteCategory)

			// Stock Operations (Double-Entry)
			inventory.GET("/stock", controller.GetStockLevels)
			inventory.POST("/move", controller.ProcessStockDraw) // Stock movements/draws
			inventory.GET("/history", controller.GetDraws)       // Movement history
		}

		// --- ADMINISTRATION / CORE ---
		admin := api.Group("/admin")
		{
			admin.GET("/company", controller.GetCompanyProfile)
			admin.PUT("/company", controller.UpdateCompany)

			// Role Management
			admin.GET("/roles", controller.GetAllRole)
			admin.POST("/roles", controller.AddRole)
			admin.PUT("/roles/:id", controller.UpdateRole)
			admin.DELETE("/roles/:id", controller.DeleteRole)

			// Audit Logs
			admin.GET("/tracker", controller.GetAuditLogs)
		}
	}

	// 4. SYSTEM ROUTES
	// Note: migrations.SchemaMigrations handles your gorm auto-migrations
	r.GET("/sync", middleware.JwtValidate, migrations.SchemaMigrations)

	if os.Getenv("SEED_DB") == "true" {
		seeders.SeedAll()
	}

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Warego ERP running on port %s", port)
	r.Run(":" + port)
}

// TenantContextGuard is a safety layer that validates the TenantID
// passed from the JWT middleware.
func TenantContextGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Attempt to get the tenantID set by JwtValidate middleware
		val, exists := c.Get("tenantID")

		if !exists {
			// Optional: Fallback check for dev headers (remove in production)
			headerID := c.GetHeader("X-Tenant-ID")
			if headerID != "" {
				id, err := strconv.Atoi(headerID)
				if err == nil {
					c.Set("tenantID", id)
					c.Next()
					return
				}
			}

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Multi-tenancy context missing. Access denied.",
			})
			return
		}

		// Ensure the type is int to prevent GORM query errors
		if _, ok := val.(int); !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Tenant context corrupted: expected integer ID",
			})
			return
		}

		c.Next()
	}
}
