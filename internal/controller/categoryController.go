// internal/controller/categoryController.go
package controller

import (
	"net/http"
	"strings"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/middleware"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetCategories obtiene todas las categorías del tenant actual
func GetCategories(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var categories []models.Category
	if err := db.Where("tenant_id = ?", tenantID).Order("name ASC").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al obtener categorías"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"categories": categories,
	})
}

// GetCategory obtiene una categoría específica por ID
func GetCategory(c *gin.Context) {
	id := c.Param("id")
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var category models.Category
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Categoría no encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al buscar categoría"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "category": category})
}

// AddCategory crea una nueva categoría (Sujeto al tenant actual)
func AddCategory(c *gin.Context) {
	tenantIDStr := middleware.GetTenantID(c)
	tenantUUID, _ := uuid.Parse(tenantIDStr)

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "El nombre es obligatorio"})
		return
	}

	category := models.Category{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		TenantID:    tenantUUID,
	}

	if err := database.GetDB().Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al guardar la categoría"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "category": category})
}

// UpdateCategory actualiza una categoría existente
func UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var category models.Category
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Categoría no encontrada"})
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
		category.Name = strings.TrimSpace(req.Name)
	}
	category.Description = strings.TrimSpace(req.Description)

	if err := db.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al actualizar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "category": category})
}

// DeleteCategory elimina una categoría si no tiene productos asociados
func DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	// Validación de seguridad: Verificar si la categoría pertenece al tenant
	var category models.Category
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Categoría no encontrada"})
		return
	}

	// Integridad Referencial: No borrar si hay productos usándola
	var productCount int64
	db.Model(&models.Product{}).Where("category_id = ? AND tenant_id = ?", id, tenantID).Count(&productCount)
	if productCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "No se puede eliminar: existen productos asociados a esta categoría",
		})
		return
	}

	if err := db.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al eliminar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Categoría eliminada correctamente"})
}
