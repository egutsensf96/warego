package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetProducts returns all products belonging to the authenticated tenant
func GetProducts(c *gin.Context) {
	db := database.GetDB()
	tenantID := c.MustGet("tenantID").(string)

	var products []models.Product
	if err := db.Where("tenant_id = ?", tenantID).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

// CreateProduct adds a new product linked to the tenant
func CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Inject TenantID from context to ensure isolation
	tenantIDStr := c.MustGet("tenantID").(string)
	product.TenantID = uuid.MustParse(tenantIDStr)

	db := database.GetDB()
	if err := db.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create product"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// UpdateProduct modifies an existing product if it belongs to the tenant
func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.MustGet("tenantID").(string)
	db := database.GetDB()

	var product models.Product
	// Ensure the product exists AND belongs to the tenant
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.Save(&product)
	c.JSON(http.StatusOK, product)
}

// DeleteProduct performs a soft delete on a product
func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.MustGet("tenantID").(string)
	db := database.GetDB()

	result := db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Product{})

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found or unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
