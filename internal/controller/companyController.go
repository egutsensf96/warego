package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetCompanyProfile returns the details of the current tenant's company
func GetCompanyProfile(c *gin.Context) {
	tenantIDStr := c.GetString("tenantID")
	db, _ := database.IntialDB()

	var company models.Company

	// We search by ID using the tenantID because in a multi-tenant
	// setup, the TenantID is the ID of the company record itself.
	result := db.Where("id = ?", tenantIDStr).First(&company)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company profile not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": company})
}

// UpdateCompany updates the corporate information for the current instance
func UpdateCompany(c *gin.Context) {
	tenantIDStr := c.GetString("tenantID")

	var body struct {
		Name        string `json:"name"`
		TaxID       string `json:"tax_id"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	db, _ := database.IntialDB()

	// Scoped update to ensure a user only updates their own company
	result := db.Model(&models.Company{}).
		Where("id = ?", tenantIDStr).
		Updates(models.Company{
			Name:        body.Name,
			TaxID:       body.TaxID,
			Description: body.Description,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company information"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company record not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Company updated successfully"})
}

// CreateCompany is usually a super-admin function used during "Sign Up"
// or when onboarding a new business instance.
func CreateCompany(c *gin.Context) {
	var body struct {
		Name  string `json:"name" binding:"required"`
		TaxID string `json:"tax_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and TaxID are required"})
		return
	}

	db, _ := database.IntialDB()

	// In a new company creation, we generate a new TenantID
	newTenantID := uuid.New()

	company := models.Company{
		Base: models.Base{
			ID:       newTenantID, // The Company ID IS the TenantID
			TenantID: newTenantID,
		},
		Name:  body.Name,
		TaxID: body.TaxID,
	}

	if err := db.Create(&company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company instance"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Company instance created",
		"tenant_id": newTenantID,
	})
}
