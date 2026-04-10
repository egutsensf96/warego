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
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	r := gin.Default()

	// 1. GLOBAL MIDDLEWARE
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(cors.Default())

	// Security Headers Middleware
	r.Use(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	})

	// 2. AUTHENTICATION (Public)
	auth := r.Group("/auth")
	{
		auth.POST("/login", controller.SingIn)
		auth.POST("/register", controller.SignUp)
	}

	// 3. ERP MODULES (Requires JWT & Tenant Context)
	// Every route here will require 'X-Tenant-ID' in the header
	api := r.Group("/api/v1")
	api.Use(middleware.JwtValidate)
	api.Use(TenantContextMiddleware())
	{
		// INVENTORY MODULE
		inventory := api.Group("/inventory")
		{
			inventory.GET("/products", controller.GetProducts)
			inventory.POST("/products", controller.CreateProduct)
			inventory.GET("/categories", controller.GetCategories)
			inventory.GET("/stock", controller.GetStockLevels)
			inventory.POST("/draw", controller.ProcessStockDraw) // The 'Draw' logic
		}

		// HUMAN RESOURCES MODULE
		hr := api.Group("/hr")
		{
			hr.GET("/employees", controller.GetAllEmployees)
			hr.POST("/employees", controller.CreateEmployee)
			hr.GET("/contracts", controller.GetEmployeeContracts)
			hr.POST("/contracts", controller.SignNewContract)
		}

		// ADMINISTRATION / CORE
		admin := api.Group("/admin")
		{
			admin.GET("/company", controller.GetCompanyProfile)
			admin.GET("/roles", controller.GetAllRole)
			admin.POST("/roles", controller.AddRole)
			admin.GET("/tracker", controller.GetAuditLogs) // The 'Tracker' logic
		}
	}

	// 4. DATABASE SYNC
	// Trigger AutoMigrate for: Company, User, Role, Product, Category,
	// Stock, Warehouse, Employee, Contract, Tracker
	r.GET("/sync", middleware.JwtValidate, middleware.IsSuperAdmin(), migrations.SchemaMigrations)
	if os.Getenv("SEED_DB") == "true" {
		seeders.SeedAll()
	}

	r.Run(":8080")
}

// TenantContextMiddleware validates that the request is scoped to a specific instance
func TenantContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Multi-tenancy error: X-Tenant-ID header is missing",
			})
			return
		}
		// Inject into Gin context for controllers to use in GORM queries
		c.Set("tenantID", tenantID)
		c.Next()
	}
}
