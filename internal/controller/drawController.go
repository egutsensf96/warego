package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ProcessStockDraw(c *gin.Context) {
	// 1. Get Context from Middleware
	tenantIDStr := c.GetString("tenantID")
	tenantID, _ := uuid.Parse(tenantIDStr)

	// Get the authenticated user ID
	val, _ := c.Get("user")
	currentUser := val.(models.User)

	var body struct {
		ProductID   uuid.UUID `json:"product_id" binding:"required"`
		WarehouseID uuid.UUID `json:"warehouse_id" binding:"required"`
		Quantity    float64   `json:"quantity" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, _ := database.IntialDB()

	// 2. Transactional Operation
	// We use a transaction to ensure Stock is updated AND the Draw record is created together
	err := db.Transaction(func(tx *gorm.DB) error {
		var stock models.Stock

		// Find current stock in the specific warehouse
		if err := tx.Where("product_id = ? AND warehouse_id = ? AND tenant_id = ?",
			body.ProductID, body.WarehouseID, tenantID).First(&stock).Error; err != nil {
			return gorm.ErrRecordNotFound // Or "Insufficient Stock"
		}

		// Check if we have enough
		if stock.Quantity < body.Quantity {
			return gorm.ErrInvalidData // Custom error logic
		}

		// Update Stock Quantity
		stock.Quantity -= body.Quantity
		if err := tx.Save(&stock).Error; err != nil {
			return err
		}

		// Create the Draw record (Audit Trail)
		draw := models.Draw{
			Base: models.Base{
				TenantID: tenantID,
			},
			Product_Id: body.ProductID, // Adjusted to match your schema naming
			Stock:      float32(body.Quantity),
			User_Id:    currentUser.ID,
		}

		if err := tx.Create(&draw).Error; err != nil {
			return err
		}

		// Log into Tracker as well
		tracker := models.Tracker{
			Base:   models.Base{TenantID: tenantID},
			UserID: currentUser.ID,
			Event:  "STOCK_WITHDRAWAL",
		}
		tx.Create(&tracker)

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed: stock may be insufficient"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stock withdrawal successful"})
}

func GetDraws(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	db, _ := database.IntialDB()

	var draws []models.Draw

	// Preload Product and User data for the frontend
	result := db.Preload("Product").Preload("User").
		Where("tenant_id = ?", tenantID).
		Find(&draws)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch withdrawal logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": draws})
}
