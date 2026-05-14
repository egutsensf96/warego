// internal/controller/tenantController.go
package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/middleware"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetTenants - Solo SuperAdmin: Lista todos los tenants registrados
func GetTenants(c *gin.Context) {
	if !middleware.IsSuperAdminUser(c) {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Acceso restringido a SuperAdmin"})
		return
	}

	db := database.GetDB()
	var tenants []models.Tenant

	// Opcional: Filtrado por nombre o dominio si se pasan por query params
	query := db.Order("created_at DESC")
	if search := c.Query("search"); search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(domain) LIKE ?", searchTerm, searchTerm)
	}

	if err := query.Find(&tenants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al obtener tenants"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "tenants": tenants})
}

// GetTenant - Obtiene los detalles de un tenant por ID
func GetTenant(c *gin.Context) {
	if !middleware.IsSuperAdminUser(c) {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Acceso restringido"})
		return
	}

	id := c.Param("id")
	db := database.GetDB()
	var tenant models.Tenant

	if err := db.First(&tenant, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Tenant no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al buscar tenant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "tenant": tenant})
}

// CreateTenant - Crea una nueva empresa/entidad en el sistema
func CreateTenant(c *gin.Context) {
	if !middleware.IsSuperAdminUser(c) {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Acceso restringido"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Domain      string `json:"domain" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Nombre y dominio son obligatorios"})
		return
	}

	db := database.GetDB()
	domainClean := strings.ToLower(strings.TrimSpace(req.Domain))

	// Verificar si el dominio ya existe
	var exists int64
	db.Model(&models.Tenant{}).Where("domain = ?", domainClean).Count(&exists)
	if exists > 0 {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "El dominio ya está registrado"})
		return
	}

	tenant := models.Tenant{
		Name:        strings.TrimSpace(req.Name),
		Domain:      domainClean,
		Description: strings.TrimSpace(req.Description),
		IsActive:    true,
	}

	if err := db.Create(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al crear el tenant"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "tenant": tenant})
}

// UpdateTenant - Actualiza la configuración de un tenant
func UpdateTenant(c *gin.Context) {
	if !middleware.IsSuperAdminUser(c) {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Acceso restringido"})
		return
	}

	id := c.Param("id")
	db := database.GetDB()
	var tenant models.Tenant

	if err := db.First(&tenant, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Tenant no encontrado"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		IsActive    *bool  `json:"is_active"` // Puntero para detectar si viene el booleano
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Datos inválidos"})
		return
	}

	if req.Name != "" {
		tenant.Name = strings.TrimSpace(req.Name)
	}
	if req.IsActive != nil {
		tenant.IsActive = *req.IsActive
	}
	tenant.Description = strings.TrimSpace(req.Description)

	if err := db.Save(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al actualizar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "tenant": tenant})
}

// DeleteTenant - Elimina un tenant (Soft Delete)
func DeleteTenant(c *gin.Context) {
	if !middleware.IsSuperAdminUser(c) {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Acceso restringido"})
		return
	}

	id := c.Param("id")
	db := database.GetDB()

	// Opcional: Impedir borrado si hay usuarios activos en este tenant
	var userCount int64
	db.Model(&models.User{}).Where("tenant_id = ?", id).Count(&userCount)
	if userCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "No se puede eliminar: el tenant tiene usuarios asociados. Desactívelos primero.",
		})
		return
	}

	if err := db.Delete(&models.Tenant{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al eliminar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Tenant eliminado correctamente"})
}
