package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type OnboardRequest struct {
	CompanyName string `json:"company_name" binding:"required"`
	Domain      string `json:"domain"`
	AdminEmail  string `json:"admin_email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
}

// GetCompanyProfile retrieves the profile of the current tenant
func GetCompanyProfile(c *gin.Context) {
	db := database.GetDB()
	tenantID := c.MustGet("tenantID").(string)

	var tenant models.Tenant
	if err := db.Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company profile not found"})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// UpdateCompany updates the company name or domain for the current tenant
func UpdateCompany(c *gin.Context) {
	db := database.GetDB()
	tenantID := c.MustGet("tenantID").(string)

	var tenant models.Tenant
	// Find the existing record first
	if err := db.Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company profile not found"})
		return
	}

	// Create a temporary struct for binding to avoid overwriting sensitive fields like ID
	var updateData struct {
		Name   string `json:"name"`
		Domain string `json:"domain"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Apply updates
	if updateData.Name != "" {
		tenant.Name = updateData.Name
	}
	if updateData.Domain != "" {
		tenant.Domain = updateData.Domain
	}

	if err := db.Save(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company profile"})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// CreateCompany handles the initial setup of a Tenant and its first Admin User
func CreateCompany(c *gin.Context) {
	var req OnboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// 1. Hash the admin password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process security credentials"})
		return
	}

	// 2. Execute within a Transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		// Create the Tenant
		tenant := models.Tenant{
			Name:   req.CompanyName,
			Domain: req.Domain,
		}
		if err := tx.Create(&tenant).Error; err != nil {
			return err
		}

		// Create the 'Admin' Role for this tenant (or find existing global role)
		// Usually, roles are seeded, but we'll assume a standard 'Admin' role ID here
		// or fetch the first role named 'Admin'
		var adminRole models.Role
		if err := tx.Where("name = ?", "Admin").First(&adminRole).Error; err != nil {
			return err // Ensure roles are seeded before onboarding
		}

		// Create the Admin User linked to the new Tenant
		adminUser := models.User{
			Email:    req.AdminEmail,
			Password: string(hashedPassword),
			TenantID: tenant.ID,
			RoleID:   adminRole.ID,
		}

		if err := tx.Create(&adminUser).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Could not create company. Domain or Email might already exist."})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Company onboarded successfully",
		"company": req.CompanyName,
	})
}
