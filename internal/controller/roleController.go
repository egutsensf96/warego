package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AddRole(c *gin.Context) {
	// 1. Get TenantID from Context (Set by Middleware)
	tenantIDStr := c.GetString("tenantID")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Tenant ID"})
		return
	}

	var body struct {
		Description string `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Description is required"})
		return
	}

	// 2. Initialize DB
	db, _ := database.IntialDB()

	role := models.Role{
		Base: models.Base{
			TenantID: tenantID,
		},
		Description: body.Description,
	}

	if result := db.Create(&role); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": role})
}

func GetAllRole(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	db, _ := database.IntialDB()

	var roles []models.Role

	// CRITICAL: Always filter by TenantID
	result := db.Where("tenant_id = ?", tenantID).Find(&roles)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": roles})
}

func UpdateRole(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	roleID := c.Param("id")

	var body struct {
		Description string `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Description is required"})
		return
	}

	db, _ := database.IntialDB()

	// Ensure the role belongs to the tenant before updating
	result := db.Model(&models.Role{}).
		Where("id = ? AND tenant_id = ?", roleID, tenantID).
		Update("description", body.Description)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found in this instance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Update successful"})
}

func DeleteRole(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	roleID := c.Param("id")

	db, _ := database.IntialDB()

	// Soft delete (GORM handles this via DeletedAt in Base struct)
	result := db.Where("id = ? AND tenant_id = ?", roleID, tenantID).Delete(&models.Role{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Role deleted successfully"})
}
