// internal/controller/companyController.go
package controller

import (
	"net/http"
	"strings"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/middleware"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type OnboardRequest struct {
	CompanyName string `json:"company_name" binding:"required,min=2,max=100"`
	Domain      string `json:"domain" binding:"required"`
	AdminEmail  string `json:"admin_email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8,max=128"`
}

// GetCompanyProfile retrieves the profile of the current tenant
func GetCompanyProfile(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var tenant models.Tenant
	if err := db.Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company profile not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tenant,
	})
}

// UpdateCompany updates the company name or domain for the current tenant
func UpdateCompany(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var tenant models.Tenant
	if err := db.Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company profile not found"})
		return
	}

	var updateData struct {
		Name   string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
		Domain string `json:"domain,omitempty" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if updateData.Name != "" {
		tenant.Name = strings.TrimSpace(updateData.Name)
	}
	if updateData.Domain != "" {
		domain := strings.ToLower(strings.TrimSpace(updateData.Domain))
		// Check for duplicate domain (excluding current tenant)
		var existing models.Tenant
		if err := db.Where("domain = ? AND id != ?", domain, tenantID).First(&existing).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Domain already in use by another tenant"})
			return
		}
		tenant.Domain = domain
	}

	if err := db.Save(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Company profile updated successfully",
		"data":    tenant,
	})
}

// CreateCompany handles the initial setup of a Tenant and its first Admin User
func CreateCompany(c *gin.Context) {
	var req OnboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Hash the admin password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process security credentials"})
		return
	}

	// Execute within a Transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		// Create the Tenant
		tenant := models.Tenant{
			Name:   strings.TrimSpace(req.CompanyName),
			Domain: strings.ToLower(strings.TrimSpace(req.Domain)),
		}
		if err := tx.Create(&tenant).Error; err != nil {
			return err
		}

		// Create or find the 'Admin' Role
		var adminRole models.Role
		if err := tx.Where("name = ? AND (tenant_id = ? OR is_global = ?)", "Admin", tenant.ID, true).First(&adminRole).Error; err != nil {
			// Create default Admin role if not found
			adminRole = models.Role{
				Name:     "Admin",
				TenantID: &tenant.ID, // ✅ FIX: Use pointer to uuid.UUID
			}
			if err := tx.Create(&adminRole).Error; err != nil {
				return err
			}
		}

		// Create the Admin User linked to the new Tenant
		adminUser := models.User{
			Name:     strings.Split(req.AdminEmail, "@")[0], // Use email prefix as name
			Email:    strings.ToLower(strings.TrimSpace(req.AdminEmail)),
			Password: string(hashedPassword),
			TenantID: tenant.ID,
			RoleID:   adminRole.ID,
			IsActive: true,
		}

		if err := tx.Create(&adminUser).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Company domain or admin email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create company: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Company onboarded successfully",
		"company": req.CompanyName,
	})
}
