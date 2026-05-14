// internal/controller/syncController.go
package controller

import (
	"net/http"
	"time"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/middleware"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PushSync handles offline→online sync from Flutter
func PushSync(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var req struct {
		Products  []map[string]interface{} `json:"products"`
		Timestamp string                   `json:"timestamp"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	var syncedIDs []string

	for _, p := range req.Products {
		// 1. Validate tenant isolation
		if p["tenant_id"] != tenantID {
			continue
		}

		// 2. Map local fields to Product model
		product := models.Product{
			Name:        getString(p, "name"),
			SKU:         getString(p, "sku"),
			Quantity:    getInt(p, "quantity"),
			ImageBase64: getString(p, "image_base64"),
			TenantID:    uuid.MustParse(tenantID),
		}

		// Optional fields
		if catID, ok := p["category_id"].(string); ok && catID != "" {
			product.CategoryID = uuid.MustParse(catID)
		}

		// 3. Handle existing vs new product (upsert logic)
		if serverID, ok := p["server_id"].(string); ok && serverID != "" {
			product.ID = uuid.MustParse(serverID)
			if err := db.Model(&product).Where("id = ? AND tenant_id = ?", product.ID, tenantID).Updates(product).Error; err != nil {
				continue
			}
		} else {
			if err := db.Create(&product).Error; err != nil {
				// Handle duplicate SKU gracefully
				if err.Error() == `pq: duplicate key value violates unique constraint "products_sku_key"` {
					var existing models.Product
					if db.Where("sku = ? AND tenant_id = ?", product.SKU, tenantID).First(&existing).Error == nil {
						product.ID = existing.ID
						db.Model(&product).Where("id = ?", existing.ID).Updates(product)
					}
				}
				continue
			}
		}

		// 4. Update/create Tracker if warehouse_id provided
		if whIDStr, ok := p["warehouse_id"].(string); ok && whIDStr != "" {
			whID := uuid.MustParse(whIDStr)
			var tracker models.Tracker
			if err := db.Where("product_id = ? AND warehouse_id = ?", product.ID, whID).
				FirstOrCreate(&tracker, models.Tracker{
					ProductID:   product.ID,
					WarehouseID: whID,
					TenantID:    product.TenantID,
					Quantity:    0,
				}).Error; err != nil {
				continue
			}
			if tracker.Quantity != product.Quantity {
				db.Model(&tracker).Update("quantity", product.Quantity)
			}
		}

		syncedIDs = append(syncedIDs, product.ID.String())
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"synced_ids": syncedIDs,
		"count":      len(syncedIDs),
	})
}

// PullSync returns changes since last sync timestamp
func PullSync(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	lastSyncStr := c.Query("last_sync_at")

	lastSync := time.Unix(0, 0) // Default to 1970-01-01
	if lastSyncStr != "" {
		if parsed, err := time.Parse(time.RFC3339, lastSyncStr); err == nil {
			lastSync = parsed
		}
	}

	db := database.GetDB()
	var products []models.Product

	if err := db.Preload("Category").
		Where("tenant_id = ? AND updated_at > ?", tenantID, lastSync).
		Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch updates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"products":  products,
		"count":     len(products),
		"last_sync": time.Now().UTC().Format(time.RFC3339),
	})
}

// ==================== HELPERS ====================

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return 0
}
