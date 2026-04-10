package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AddCategory(c *gin.Context) {
	// 1. Get TenantID from context
	tenantIDStr := c.GetString("tenantID")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Tenant Context"})
		return
	}

	var body struct {
		Description string `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category description is required"})
		return
	}

	db, _ := database.IntialDB()

	category := models.Category{
		Base: models.Base{
			TenantID: tenantID,
		},
		Description: body.Description,
	}

	if result := db.Create(&category); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create category"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": category})
}

func GetCategories(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	db, _ := database.IntialDB()

	var categories []models.Category

	// Ensure we only retrieve categories for this specific tenant
	if err := db.Where("tenant_id = ?", tenantID).Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": categories})
}

func UpdateCategory(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	id := c.Param("id")

	var body struct {
		Description string `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Description is required"})
		return
	}

	db, _ := database.IntialDB()

	// Scoped update: Must match both ID and TenantID
	result := db.Model(&models.Category{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Update("description", body.Description)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found in this instance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Category updated successfully"})
}

func DeleteCategory(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	id := c.Param("id")

	db, _ := database.IntialDB()

	// Using GORM's soft delete (updates deleted_at)
	result := db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Category{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Category deleted successfully"})
}
