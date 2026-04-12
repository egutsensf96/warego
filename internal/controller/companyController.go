package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
)

// GetCompanyProfile returns the details of the current tenant's company
func GetCompanyProfile(c *gin.Context) {
	// 1. Retrieve the tenantID from context (Middleware should set this as int)
	tenantID, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No tenant context found"})
		return
	}

	db, _ := database.IntialDB()

	// In our schema, the model is named 'Tenant'
	var company models.Tenant

	// We search by ID because the Tenant's primary key is the identifier
	if err := db.First(&company, tenantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company profile not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": company})
}

// UpdateCompany updates the corporate information for the current instance
func UpdateCompany(c *gin.Context) {
	tenantID, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No tenant context found"})
		return
	}

	var body struct {
		Name  string `json:"name"`
		TaxID string `json:"tax_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	db, _ := database.IntialDB()

	// Perform a scoped update
	result := db.Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Updates(models.Tenant{
			Name:  body.Name,
			TaxID: body.TaxID,
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

// CreateCompany is used during onboarding a new business instance.
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

	// In an integer SERIAL setup, Postgres generates the ID automatically.
	// We don't need to manually generate a UUID.
	company := models.Tenant{
		Name:  body.Name,
		TaxID: body.TaxID,
	}

	if err := db.Create(&company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company instance"})
		return
	}

	// After Create, company.ID is populated by GORM/Postgres
	c.JSON(http.StatusCreated, gin.H{
		"message":   "Company instance created",
		"tenant_id": company.ID,
	})
}
