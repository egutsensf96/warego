// internal/controller/warehouseController.go
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

// GetWarehouses obtiene todos los almacenes del tenant actual
func GetWarehouses(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var warehouses []models.Warehouse
	if err := db.Where("tenant_id = ?", tenantID).Order("name ASC").Find(&warehouses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al obtener almacenes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"warehouses": warehouses,
	})
}

// GetWarehouse obtiene un almacén específico por ID
func GetWarehouse(c *gin.Context) {
	id := c.Param("id")
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var warehouse models.Warehouse
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&warehouse).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Almacén no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al buscar el almacén"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "warehouse": warehouse})
}

// CreateWarehouse crea un nuevo almacén para la organización
func CreateWarehouse(c *gin.Context) {
	tenantIDStr := middleware.GetTenantID(c)
	tenantUUID, _ := uuid.Parse(tenantIDStr)

	var req struct {
		Name     string `json:"name" binding:"required"`
		Location string `json:"location"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "El nombre es obligatorio"})
		return
	}

	warehouse := models.Warehouse{
		Name:     strings.TrimSpace(req.Name),
		Location: strings.TrimSpace(req.Location),
		TenantID: tenantUUID,
	}

	if err := database.GetDB().Create(&warehouse).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al crear el almacén"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "warehouse": warehouse})
}

// UpdateWarehouse actualiza la información de un almacén
func UpdateWarehouse(c *gin.Context) {
	id := c.Param("id")
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var warehouse models.Warehouse
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&warehouse).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Almacén no encontrado"})
		return
	}

	var req struct {
		Name     string `json:"name"`
		Location string `json:"location"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Datos inválidos"})
		return
	}

	if req.Name != "" {
		warehouse.Name = strings.TrimSpace(req.Name)
	}
	warehouse.Location = strings.TrimSpace(req.Location)

	if err := db.Save(&warehouse).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al actualizar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "warehouse": warehouse})
}

// DeleteWarehouse elimina un almacén si no tiene inventario activo
func DeleteWarehouse(c *gin.Context) {
	id := c.Param("id")
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	// 1. Verificar existencia y pertenencia
	var warehouse models.Warehouse
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&warehouse).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Almacén no encontrado"})
		return
	}

	// 2. Verificar si hay stock registrado en este almacén (usando el modelo Tracker o Stock)
	var stockCount int64
	db.Model(&models.Tracker{}).Where("warehouse_id = ? AND tenant_id = ? AND quantity > 0", id, tenantID).Count(&stockCount)

	if stockCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "No se puede eliminar: el almacén todavía contiene productos con stock",
		})
		return
	}

	// 3. Eliminar (Soft Delete)
	if err := db.Delete(&warehouse).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al eliminar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Almacén eliminado correctamente"})
}
