// internal/controller/auditController.go
package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/middleware"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
)

// GetAuditLogsGlobal - SuperAdmin only: fetch audit logs across ALL tenants
// Endpoint: GET /api/v1/superadmin/audit-logs
func GetAuditLogsGlobal(c *gin.Context) {
	if !middleware.IsSuperAdminUser(c) {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "SuperAdmin access required"})
		return
	}

	db := database.GetDB()
	query := db.Model(&models.AuditLog{})

	// ✅ OPTIONAL filters (removed by default to show EVERYTHING)
	if tid := c.Query("tenant_id"); tid != "" {
		query = query.Where("tenant_id = ?", tid)
	}
	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
	}
	if targetType := c.Query("target_type"); targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to count logs"})
		return
	}

	// Fetch logs
	var logs []models.AuditLog
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"logs": logs,
			"pagination": gin.H{
				"page":     page,
				"limit":    limit,
				"total":    total,
				"pages":    (int(total) + limit - 1) / limit,
				"has_next": offset+limit < int(total),
			},
		},
	})
}

// GetAuditLogsTenant - Admin only: fetch audit logs for current tenant only
// Endpoint: GET /api/v1/admin/audit-logs
func GetAuditLogsTenant(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	// Scope query to tenant
	query := db.Where("tenant_id = ?", tenantID).Preload("Changes")

	// Optional filtering (same as global but tenant-scoped)
	actionFilter := c.Query("action")
	if actionFilter != "" {
		query = query.Where("action = ?", models.AuditLogAction(actionFilter))
	}

	targetType := c.Query("target_type")
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}

	search := c.Query("search")
	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where(
			"description ILIKE ? OR target_name ILIKE ? OR user_email ILIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
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
			limit = 100
		}
	}
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.AuditLog{}).Count(&total)

	var logs []models.AuditLog
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"logs": logs,
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
