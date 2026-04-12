package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
)

func CreateProduct(c *gin.Context) {
	// 1. Get Context (Assumes Middleware sets these as int and models.User)
	tenantID, _ := c.Get("tenantID")

	var body struct {
		Name        string  `json:"name" binding:"required"`
		CategoryID  int     `json:"category_id" binding:"required"`
		SKU         string  `json:"sku" binding:"required"`
		CostPrice   float64 `json:"cost_price" binding:"required,gte=0"`
		ListPrice   float64 `json:"list_price" binding:"required,gte=0"`
		ImageBase64 string  `json:"image_base64"` // Requirement: Codec Base64
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, _ := database.IntialDB()

	// 2. Build the Product Template and Variant (Odoo Structure)
	template := models.ProductTemplate{
		TenantID:    tenantID.(int),
		CategoryID:  body.CategoryID,
		Name:        body.Name,
		ImageBase64: body.ImageBase64,
		Type:        "storable", // Default for inventory
	}

	// Save template first to get ID
	if err := db.Create(&template).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Invalid category or template data"})
		return
	}

	variant := models.ProductVariant{
		TemplateID: template.ID,
		TenantID:   tenantID.(int),
		SKU:        body.SKU,
		CostPrice:  body.CostPrice,
		ListPrice:  body.ListPrice,
		Active:     true,
	}

	if err := db.Create(&variant).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "SKU already exists for this tenant"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product and Variant created",
		"data":    variant,
	})
}

func GetProducts(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	db, _ := database.IntialDB()

	var products []models.ProductTemplate

	// Preload Category and Variants (Hierarchical Odoo style)
	result := db.Preload("Category").Preload("Variants").
		Where("tenant_id = ?", tenantID).
		Find(&products)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": products})
}

func UpdateProduct(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	variantID := c.Param("id")

	var body struct {
		CostPrice float64 `json:"cost_price"`
		ListPrice float64 `json:"list_price"`
		Active    bool    `json:"active"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, _ := database.IntialDB()

	// Scoped update for Variant
	result := db.Model(&models.ProductVariant{}).
		Where("id = ? AND tenant_id = ?", variantID, tenantID).
		Updates(models.ProductVariant{
			CostPrice: body.CostPrice,
			ListPrice: body.ListPrice,
			Active:    body.Active,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product variant not found in your instance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Product variant updated successfully"})
}

func DeleteProduct(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	templateID := c.Param("id")

	db, _ := database.IntialDB()

	// Scoped delete for Template (Cascades to Variants depending on DB setup)
	result := db.Where("id = ? AND tenant_id = ?", templateID, tenantID).Delete(&models.ProductTemplate{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Product template deleted successfully"})
}
