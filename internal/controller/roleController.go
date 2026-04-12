package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
)

func AddRole(c *gin.Context) {
	// 1. Get TenantID from Context (Ensuring it's an integer)
	tenantID, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context missing"})
		return
	}

	var body struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role name is required"})
		return
	}

	db, _ := database.IntialDB()

	role := models.Role{
		TenantID: tenantID.(int),
		Name:     body.Name,
	}

	if result := db.Create(&role); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": role})
}

func GetAllRole(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	db, _ := database.IntialDB()

	var roles []models.Role

	// Scoped search
	result := db.Where("tenant_id = ?", tenantID).Find(&roles)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": roles})
}

func UpdateRole(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	roleID := c.Param("id")

	var body struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	db, _ := database.IntialDB()

	// Update strictly scoped by TenantID
	result := db.Model(&models.Role{}).
		Where("id = ? AND tenant_id = ?", roleID, tenantID).
		Update("name", body.Name)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found or unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Update successful"})
}

func DeleteRole(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	roleID := c.Param("id")

	db, _ := database.IntialDB()

	// Guard against deleting roles from other tenants
	result := db.Where("id = ? AND tenant_id = ?", roleID, tenantID).Delete(&models.Role{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Role deleted successfully"})
}
