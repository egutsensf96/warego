package controller

import (
	"net/http"
	"time"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetStockLevels returns the current quantity of all products across warehouses
func GetStockLevels(c *gin.Context) {
	db := database.GetDB()
	tenantID := c.MustGet("tenantID").(string)

	var stock []models.Tracker
	// We use Preload to get Product and Warehouse details in the same response
	if err := db.Preload("Product").Preload("Warehouse").Where("tenant_id = ?", tenantID).Find(&stock).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch stock levels"})
		return
	}

	c.JSON(http.StatusOK, stock)
}

// ProcessStockDraw handles the creation of a Draw and reduces the Tracker quantity
func ProcessStockDraw(c *gin.Context) {
	var req struct {
		Name        string    `json:"name" binding:"required"`
		ProductID   uuid.UUID `json:"product_id" binding:"required"`
		WarehouseID uuid.UUID `json:"warehouse_id" binding:"required"`
		WinnerID    uuid.UUID `json:"winner_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantIDStr := c.MustGet("tenantID").(string)
	tenantID := uuid.MustParse(tenantIDStr)
	db := database.GetDB()

	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. Check and Reduce Stock
		var tracker models.Tracker
		if err := tx.Where("product_id = ? AND warehouse_id = ? AND tenant_id = ?", req.ProductID, req.WarehouseID, tenantID).First(&tracker).Error; err != nil {
			return gorm.ErrRecordNotFound
		}

		if tracker.Quantity <= 0 {
			return gorm.ErrInvalidData // Or custom "Out of Stock" error
		}

		tracker.Quantity--
		if err := tx.Save(&tracker).Error; err != nil {
			return err
		}

		// 2. Create Draw Record
		draw := models.Draw{
			Name:      req.Name,
			ProductID: req.ProductID,
			WinnerID:  &req.WinnerID,
			TenantID:  tenantID,
			Status:    "pending",
		}

		return tx.Create(&draw).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Stock moved and draw created successfully"})
}

// GetDraws retrieves history of all draws for the tenant
func GetDraws(c *gin.Context) {
	db := database.GetDB()
	tenantID := c.MustGet("tenantID").(string)

	var draws []models.Draw
	if err := db.Preload("Product").Where("tenant_id = ?", tenantID).Find(&draws).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch draws"})
		return
	}

	c.JSON(http.StatusOK, draws)
}

// RetireProduct marks a draw as physically collected
func RetireProduct(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.MustGet("tenantID").(string)

	// In a real app, this would come from the JWT claims (the admin logged in)
	adminIDStr, _ := c.Get("userID")

	db := database.GetDB()
	var draw models.Draw

	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&draw).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Draw not found"})
		return
	}

	if draw.Status == "retrieved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product already retired"})
		return
	}

	now := time.Now()
	draw.RetrievedAt = &now
	draw.Status = "retrieved"

	// If adminID is available in context, set it
	if adminIDStr != nil {
		uid := uuid.MustParse(adminIDStr.(string))
		draw.RetrievedByID = &uid
	}

	db.Save(&draw)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Product retired successfully",
		"retrieved_at": draw.RetrievedAt,
	})
}
