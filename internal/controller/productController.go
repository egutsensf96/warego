package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateProduct(c *gin.Context) {
	// 1. Get Context from Middleware
	tenantIDStr := c.GetString("tenantID")
	tenantID, _ := uuid.Parse(tenantIDStr)

	val, _ := c.Get("user")
	currentUser := val.(models.User)

	var body struct {
		SKU         string    `json:"sku" binding:"required"`
		Description string    `json:"description" binding:"required"`
		Cost        float64   `json:"cost" binding:"required,gte=0"`
		CategoryID  uuid.UUID `json:"category_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, _ := database.IntialDB()

	// 2. Build product model
	// Note: Field changed from User_Id to UserID to match the struct
	product := models.Product{
		Base: models.Base{
			TenantID: tenantID,
		},
		SKU:         body.SKU,
		Description: body.Description,
		Cost:        body.Cost,
		CategoryID:  body.CategoryID,
		UserID:      currentUser.ID,
	}

	// 3. Save to DB
	if result := db.Create(&product); result.Error != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "SKU already exists for this tenant or invalid category"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": product})
}

func GetProducts(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	db, _ := database.IntialDB()

	var products []models.Product

	// Preload Category relationship
	result := db.Preload("Category").Where("tenant_id = ?", tenantID).Find(&products)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": products})
}

func UpdateProduct(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	id := c.Param("id")

	var body struct {
		Description string    `json:"description"`
		Cost        float64   `json:"cost"`
		CategoryID  uuid.UUID `json:"category_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, _ := database.IntialDB()

	// Safe update scoped by ID and TenantID
	result := db.Model(&models.Product{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(models.Product{
			Description: body.Description,
			Cost:        body.Cost,
			CategoryID:  body.CategoryID,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found in your instance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Product updated successfully"})
}

func DeleteProduct(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	id := c.Param("id")

	db, _ := database.IntialDB()

	// Soft delete
	result := db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Product{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Product deleted successfully"})
}
