package main

import (
	"log"
	"net/http"
	"os"

	"github.com/egutsenf96/warego/internal/controller"
	"github.com/egutsenf96/warego/internal/database/migrations"
	"github.com/egutsenf96/warego/internal/database/seeders"
	"github.com/egutsenf96/warego/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid" // Added for UUID validation
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
		auth.POST("/company/onboard", controller.CreateCompany)
	}

	// 3. PRIVATE ERP ROUTES (Requires Valid JWT)
	api := r.Group("/api/v1")
	api.Use(middleware.JwtValidate)
	api.Use(TenantContextGuard())
	{
		// --- INVENTORY MODULE ---
		inventory := api.Group("/inventory")
		{
			// Products (Now supports Base64 images in controller logic)
			inventory.GET("/products", controller.GetProducts)
			inventory.POST("/products", controller.CreateProduct)
			inventory.PUT("/products/:id", controller.UpdateProduct)
			inventory.DELETE("/products/:id", controller.DeleteProduct)

			// Categories
			inventory.GET("/categories", controller.GetCategories)
			inventory.POST("/categories", controller.AddCategory)
			inventory.PUT("/categories/:id", controller.UpdateCategory)
			inventory.DELETE("/categories/:id", controller.DeleteCategory)

			// Stock & Draws (Now includes Retrieval logic)
			inventory.GET("/stock", controller.GetStockLevels)
			inventory.POST("/move", controller.ProcessStockDraw)
			inventory.GET("/draws", controller.GetDraws)
			inventory.PATCH("/draws/:id/retire", controller.RetireProduct) // New retrieval endpoint
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

			// Warehouse & Audit
			admin.GET("/warehouses", controller.GetWarehouses)
			admin.GET("/tracker", controller.GetAuditLogs)
		}
	}

	// 4. SYSTEM ROUTES
	r.GET("/sync", migrations.SchemaMigrations)

	if os.Getenv("SEED_DB") == "true" {
		seeders.SeedAll()
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Warego ERP running on port %s", port)
	r.Run(":" + port)
}

// TenantContextGuard ensures every request is scoped to a valid UUID Tenant
func TenantContextGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("tenantID")

		if !exists {
			// Check header as fallback for dev
			headerID := c.GetHeader("X-Tenant-ID")
			if headerID != "" {
				if _, err := uuid.Parse(headerID); err == nil {
					c.Set("tenantID", headerID)
					c.Next()
					return
				}
			}

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Multi-tenancy context missing. Access denied.",
			})
			return
		}

		// Validate that the context value is a valid string-format UUID
		strID, ok := val.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Tenant context corrupted: expected UUID string",
			})
			return
		}

		if _, err := uuid.Parse(strID); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Tenant UUID format",
			})
			return
		}

		c.Next()
	}
}
