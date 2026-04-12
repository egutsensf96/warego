package controller

import (
	"errors"
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ProcessStockDraw(c *gin.Context) {
	// 1. Get Context (Assumes Middleware sets these as int and models.User)
	tenantID, _ := c.Get("tenantID")
	val, _ := c.Get("user")
	currentUser := val.(models.User)

	var body struct {
		VariantID   int     `json:"variant_id" binding:"required"`
		SourceLocID int     `json:"source_location_id" binding:"required"`
		DestLocID   int     `json:"dest_location_id" binding:"required"` // e.g., a 'Customer' or 'Scrap' location
		Quantity    float64 `json:"quantity" binding:"required,gt=0"`
		Reference   string  `json:"reference"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, _ := database.IntialDB()

	// 2. Transactional Operation
	err := db.Transaction(func(tx *gorm.DB) error {
		// Calculate current stock at source to ensure availability
		// Sum(dest_qty) - Sum(src_qty)
		var incoming, outgoing float64
		tx.Model(&models.StockMove{}).Where("variant_id = ? AND dest_location_id = ? AND tenant_id = ?",
			body.VariantID, body.SourceLocID, tenantID).Select("COALESCE(SUM(qty), 0)").Scan(&incoming)
		tx.Model(&models.StockMove{}).Where("variant_id = ? AND src_location_id = ? AND tenant_id = ?",
			body.VariantID, body.SourceLocID, tenantID).Select("COALESCE(SUM(qty), 0)").Scan(&outgoing)

		currentStock := incoming - outgoing

		if currentStock < body.Quantity {
			return errors.New("insufficient stock at source location")
		}

		// Create the StockMove record (This is your lifecycle audit)
		move := models.StockMove{
			TenantID:       tenantID.(int),
			VariantID:      body.VariantID,
			SrcLocationID:  body.SourceLocID,
			DestLocationID: body.DestLocID,
			UserID:         currentUser.ID,
			Qty:            body.Quantity,
			Reference:      body.Reference,
		}

		if err := tx.Create(&move).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stock movement processed successfully"})
}

func GetDraws(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	db, _ := database.IntialDB()

	var moves []models.StockMove

	// Preload Variant and User for a detailed UI audit log
	result := db.Preload("Variant").Preload("User").
		Where("tenant_id = ?", tenantID).
		Order("created_at desc").
		Find(&moves)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": moves})
}
