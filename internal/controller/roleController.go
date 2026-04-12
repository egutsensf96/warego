package controller

import (
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
)

// GetAllRole retrieves all roles available for the tenant's users
func GetAllRole(c *gin.Context) {
	db := database.GetDB()
	var roles []models.Role

	// Depending on your design, roles might be global or per-tenant.
	// This query assumes roles are global or assigned to the tenant's users.
	if err := db.Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch roles"})
		return
	}

	c.JSON(http.StatusOK, roles)
}

// AddRole creates a new role in the system
func AddRole(c *gin.Context) {
	var role models.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	if err := db.Create(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create role"})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// UpdateRole modifies role descriptions or names
func UpdateRole(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()
	var role models.Role

	if err := db.First(&role, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.Save(&role)
	c.JSON(http.StatusOK, role)
}

// DeleteRole removes a role from the system
func DeleteRole(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()

	// Safety check: Don't delete a role if users are still assigned to it
	var userCount int64
	db.Model(&models.User{}).Where("role_id = ?", id).Count(&userCount)
	if userCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete role because users are currently assigned to it"})
		return
	}

	if err := db.Delete(&models.Role{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}
