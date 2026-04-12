package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/gin-gonic/gin"
)

// GetStockLevels retrieves current stock quantities across all internal locations for the tenant
func GetStockLevels(c *gin.Context) {
	// 1. Get TenantID from context (Ensure it's an int from middleware)
	tenantID, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context missing"})
		return
	}

	db, err := database.IntialDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// In the Odoo-style model, we calculate stock based on locations.
	// This struct helps format the response for the frontend.
	type StockReport struct {
		ProductID   int     `json:"product_id"`
		ProductName string  `json:"product_name"`
		SKU         string  `json:"sku"`
		Location    string  `json:"location_name"`
		Warehouse   string  `json:"warehouse_name"`
		Quantity    float64 `json:"quantity"`
	}

	var report []StockReport

	// 2. Aggregate Query
	// We calculate: SUM(qty at destination) - SUM(qty at source)
	// filtered by tenant and grouped by location/variant.
	query := `
		SELECT 
			pv.id as product_id, 
			pt.name as product_name, 
			pv.sku, 
			l.name as location_name, 
			w.name as warehouse_name,
			SUM(CASE WHEN sm.dest_location_id = l.id THEN sm.qty ELSE -sm.qty END) as quantity
		FROM stock_moves sm
		JOIN product_variants pv ON sm.variant_id = pv.id
		JOIN product_templates pt ON pv.template_id = pt.id
		JOIN locations l ON (sm.dest_location_id = l.id OR sm.src_location_id = l.id)
		JOIN warehouses w ON l.warehouse_id = w.id
		WHERE sm.tenant_id = ? AND l.location_type = 'internal'
		GROUP BY pv.id, pt.name, pv.sku, l.id, w.name
		HAVING SUM(CASE WHEN sm.dest_location_id = l.id THEN sm.qty ELSE -sm.qty END) > 0
	`

	result := db.Raw(query, tenantID).Scan(&report)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate stock levels"})
		return
	}

	// 3. Response
	c.JSON(http.StatusOK, gin.H{
		"count":  len(report),
		"result": report,
	})
}
