package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetCategories retrieves all categories for the authenticated tenant
func GetCategories(c *gin.Context) {
	db := database.GetDB()
	tenantID := c.MustGet("tenantID").(string)

	var categories []models.Category
	if err := db.Where("tenant_id = ?", tenantID).Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// AddCategory creates a new product category
func AddCategory(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Lock the category to the current tenant
	tenantIDStr := c.MustGet("tenantID").(string)
	category.TenantID = uuid.MustParse(tenantIDStr)

	db := database.GetDB()
	if err := db.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create category"})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// UpdateCategory updates a specific category name or details
func UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.MustGet("tenantID").(string)
	db := database.GetDB()

	var category models.Category
	// Verify ownership before updating
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.Save(&category)
	c.JSON(http.StatusOK, category)
}

// DeleteCategory removes a category (Soft Delete)
func DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.MustGet("tenantID").(string)
	db := database.GetDB()

	// Check if any products are still using this category before deleting (optional but recommended)
	var productCount int64
	db.Model(&models.Product{}).Where("category_id = ? AND tenant_id = ?", id, tenantID).Count(&productCount)
	if productCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete category while it still contains products"})
		return
	}

	result := db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Category{})

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found or unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
