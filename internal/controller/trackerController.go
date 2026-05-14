package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/middleware"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetAuditLogs retrieves stock transactions (audit trail) for the authenticated tenant
// Endpoint: GET /api/v1/admin/tracker
func GetAuditLogs(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	// Start query with tenant isolation
	query := db.Where("tenant_id = ?", tenantID).
		Preload("Product").
		Preload("Warehouse").
		Preload("User").
		Preload("Supplier")

	// Optional filtering via query params
	actionFilter := c.Query("type") // e.g., "INITIAL", "ADJUST", "DRAW"
	if actionFilter != "" {
		query = query.Where("type = ?", models.StockTransactionType(actionFilter))
	}

	productFilter := c.Query("product_id")
	if productFilter != "" {
		if pid, err := uuid.Parse(productFilter); err == nil {
			query = query.Where("product_id = ?", pid)
		}
	}

	// Date range filter
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate != "" {
		if start, err := time.Parse(time.RFC3339, startDate); err == nil {
			query = query.Where("created_at >= ?", start)
		}
	}
	if endDate != "" {
		if end, err := time.Parse(time.RFC3339, endDate); err == nil {
			query = query.Where("created_at <= ?", end)
		}
	}

	// Pagination
	page := 1
	limit := 20
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
		if limit > 100 {
			limit = 100 // Cap max limit
		}
	}
	offset := (page - 1) * limit

	// Get total count
	var total int64
	query.Model(&models.StockTransaction{}).Count(&total)

	// Fetch paginated results
	var transactions []models.StockTransaction
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"transactions": transactions,
			"pagination": gin.H{
				"page":     page,
				"limit":    limit,
				"total":    total,
				"pages":    (int(total) + limit - 1) / limit,
				"has_next": offset+limit < int(total),
				"has_prev": page > 1,
			},
		},
	})
}
