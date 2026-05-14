// internal/controller/stockController.go
package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/middleware"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetStockLevels retrieves current stock levels per warehouse (tenant-scoped)
func GetStockLevels(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var stock []models.Tracker
	if err := db.Preload("Product").Preload("Warehouse").Where("tenant_id = ?", tenantID).Find(&stock).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Could not fetch stock levels"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stock":   stock,
	})
}

// ProcessStockDraw handles stock withdrawal for a named event (e.g., giveaway, allocation)
func ProcessStockDraw(c *gin.Context) {
	var req struct {
		Name        string    `json:"name" binding:"required"`
		ProductID   uuid.UUID `json:"product_id" binding:"required"`
		WarehouseID uuid.UUID `json:"warehouse_id" binding:"required"`
		WinnerID    uuid.UUID `json:"winner_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	tenantID := middleware.GetTenantID(c)

	// Extract UserID from JWT context
	userIDStr, ok := c.Get("userID")
	if !ok || userIDStr == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "User ID missing in token"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID format"})
		return
	}

	db := database.GetDB()

	err = db.Transaction(func(tx *gorm.DB) error {
		// 1. Check & Reduce Stock
		var tracker models.Tracker
		if err := tx.Where("product_id = ? AND warehouse_id = ? AND tenant_id = ?", req.ProductID, req.WarehouseID, tenantID).First(&tracker).Error; err != nil {
			return gorm.ErrRecordNotFound
		}
		if tracker.Quantity <= 0 {
			return errors.New("insufficient stock")
		}

		// Atomic decrement
		if err := tx.Model(&tracker).Update("quantity", gorm.Expr("quantity - ?", 1)).Error; err != nil {
			return err
		}

		// 2. Create Draw Record (Business event)
		draw := models.Draw{
			Name:      strings.TrimSpace(req.Name),
			ProductID: req.ProductID,
			WinnerID:  &req.WinnerID,
			TenantID:  uuid.MustParse(tenantID),
			Status:    string(models.StatusPending),
		}
		if err := tx.Create(&draw).Error; err != nil {
			return err
		}

		// 3. Create StockTransaction for Audit Trail
		txn := models.StockTransaction{
			ProductID:      req.ProductID,
			WarehouseID:    req.WarehouseID,
			TenantID:       uuid.MustParse(tenantID),
			QuantityChange: -1,
			Type:           models.TypeDraw,
			UserID:         &userID,
			ReferenceID:    &draw.ID,
			Notes:          "Stock drawn for winner: " + req.Name,
		}
		return tx.Create(&txn).Error
	})

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "Transaction failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Stock moved and draw created successfully",
	})
}

// GetDraws retrieves all draw/retirement events for the tenant
func GetDraws(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var draws []models.Draw
	err := db.Preload("Product").
		Preload("Winner").
		Preload("Winner.Role").
		Preload("RetrievedBy").
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&draws).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Could not fetch history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"draws":   draws,
	})
}

// RetireProduct marks a Draw as completed (product retrieved by winner)
func RetireProduct(c *gin.Context) {
	idStr := c.Param("id")
	drawID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid draw ID"})
		return
	}

	tenantID := middleware.GetTenantID(c)
	userIDStr, _ := c.Get("userID")

	db := database.GetDB()
	var draw models.Draw

	if userIDStr != nil {
		if uid, err := uuid.Parse(userIDStr.(string)); err == nil {
			draw.RetrievedBy = &uid // ✅ Correct field name
		}
	}
	if err := db.Where("id = ? AND tenant_id = ?", drawID, tenantID).First(&draw).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Draw not found"})
		return
	}

	if draw.Status == string(models.StatusRetrieved) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Product already retired"})
		return
	}

	// ✅ FIX: Use MarkAsRetrieved helper method
	draw.MarkAsRetrieved(nil) // Pass userID if you want to track who confirmed retrieval

	if err := db.Save(&draw).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update draw"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "Product retired successfully",
		"retrieved_at": draw.RetrievedAt,
	})
}
