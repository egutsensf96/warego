// internal/controller/roleController.go
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

// GetRoles obtiene todos los roles pertenecientes al tenant actual
func GetRoles(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var roles []models.Role
	// Buscamos roles donde el tenant_id coincida
	if err := db.Where("tenant_id = ?", tenantID).Order("name ASC").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al obtener los roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"roles":   roles,
	})
}

// GetRole obtiene un rol específico por ID (validando pertenencia al tenant)
func GetRole(c *gin.Context) {
	id := c.Param("id")
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var role models.Role
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Rol no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al buscar el rol"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "role": role})
}

// AddRole crea un nuevo rol para el tenant actual
func AddRole(c *gin.Context) {
	tenantIDStr := middleware.GetTenantID(c)
	tenantUUID, _ := uuid.Parse(tenantIDStr)

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "El nombre del rol es obligatorio"})
		return
	}

	db := database.GetDB()
	nameClean := strings.TrimSpace(req.Name)

	// Verificar si ya existe un rol con ese nombre para este tenant
	var count int64
	db.Model(&models.Role{}).Where("LOWER(name) = ? AND tenant_id = ?", strings.ToLower(nameClean), tenantUUID).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "Ya existe un rol con ese nombre en su organización"})
		return
	}

	role := models.Role{
		Name:        nameClean,
		Description: strings.TrimSpace(req.Description),
		TenantID:    &tenantUUID, // Puntero a UUID según tu modelo
	}

	if err := db.Create(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al crear el rol"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "role": role})
}

// UpdateRole actualiza el nombre o descripción de un rol
func UpdateRole(c *gin.Context) {
	id := c.Param("id")
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var role models.Role
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&role).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Rol no encontrado"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Datos inválidos"})
		return
	}

	if req.Name != "" {
		role.Name = strings.TrimSpace(req.Name)
	}
	role.Description = strings.TrimSpace(req.Description)

	if err := db.Save(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al actualizar el rol"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "role": role})
}

// DeleteRole elimina un rol siempre que no esté en uso por ningún usuario
func DeleteRole(c *gin.Context) {
	id := c.Param("id")
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	// 1. Validar existencia y propiedad
	var role models.Role
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&role).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Rol no encontrado"})
		return
	}

	// 2. Verificar si hay usuarios que dependen de este rol
	var userCount int64
	db.Model(&models.User{}).Where("role_id = ? AND tenant_id = ?", id, tenantID).Count(&userCount)
	if userCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "No se puede eliminar el rol porque está asignado a uno o más usuarios",
		})
		return
	}

	// 3. Eliminar
	if err := db.Delete(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al eliminar el rol"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Rol eliminado correctamente"})
}
