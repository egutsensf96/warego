package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
)

func AddCategory(c *gin.Context) {
	// 1. Get TenantID from context (Middleware should set this as int)
	tenantID, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context missing"})
		return
	}

	var body struct {
		Name     string `json:"name" binding:"required"`
		ParentID *int   `json:"parent_id"` // Optional for hierarchical categories
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, _ := database.IntialDB()

	category := models.Category{
		TenantID: tenantID.(int),
		Name:     body.Name,
		ParentID: body.ParentID,
	}

	if result := db.Create(&category); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create category"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": category})
}

func GetCategories(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	db, _ := database.IntialDB()

	var categories []models.Category

	// Scoped to Tenant and preloading SubCategories if you implemented the slice in the model
	if err := db.Where("tenant_id = ?", tenantID).Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": categories})
}

func UpdateCategory(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	id := c.Param("id")

	var body struct {
		Name     string `json:"name" binding:"required"`
		ParentID *int   `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	db, _ := database.IntialDB()

	// Update logic using the struct to handle the tenant scoping
	result := db.Model(&models.Category{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(models.Category{
			Name:     body.Name,
			ParentID: body.ParentID,
		})

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
	tenantID, _ := c.Get("tenantID")
	id := c.Param("id")

	db, _ := database.IntialDB()

	// Scoped delete to ensure a user can't delete another tenant's category by guessing the ID
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
