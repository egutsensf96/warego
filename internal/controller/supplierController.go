// internal/controller/supplierController.go
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

// GetSuppliers obtiene todos los proveedores del tenant actual
func GetSuppliers(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var suppliers []models.Supplier
	if err := db.Where("tenant_id = ?", tenantID).Order("name ASC").Find(&suppliers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al obtener proveedores"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"suppliers": suppliers,
	})
}

// GetSupplier obtiene un proveedor específico por ID
func GetSupplier(c *gin.Context) {
	id := c.Param("id")
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var supplier models.Supplier
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&supplier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Proveedor no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error en el servidor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "supplier": supplier})
}

// CreateSupplier crea un nuevo proveedor
func CreateSupplier(c *gin.Context) {
	tenantIDStr := middleware.GetTenantID(c)
	tenantUUID, _ := uuid.Parse(tenantIDStr)

	var req struct {
		Name    string `json:"name" binding:"required"`
		Contact string `json:"contact"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Address string `json:"address"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "El nombre es obligatorio"})
		return
	}

	db := database.GetDB()

	// Validar si el email ya existe para este tenant (opcional)
	if req.Email != "" {
		var count int64
		emailClean := strings.ToLower(strings.TrimSpace(req.Email))
		db.Model(&models.Supplier{}).Where("email = ? AND tenant_id = ?", emailClean, tenantUUID).Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"success": false, "error": "Este correo ya está registrado para otro proveedor"})
			return
		}
	}

	supplier := models.Supplier{
		Name:     strings.TrimSpace(req.Name),
		Contact:  strings.TrimSpace(req.Contact),
		Email:    strings.ToLower(strings.TrimSpace(req.Email)),
		Phone:    strings.TrimSpace(req.Phone),
		Address:  strings.TrimSpace(req.Address),
		TenantID: tenantUUID,
	}

	if err := db.Create(&supplier).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al crear el proveedor"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "supplier": supplier})
}

// UpdateSupplier actualiza los datos de un proveedor
func UpdateSupplier(c *gin.Context) {
	id := c.Param("id")
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var supplier models.Supplier
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&supplier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Proveedor no encontrado"})
		return
	}

	var req struct {
		Name    string `json:"name"`
		Contact string `json:"contact"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Address string `json:"address"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Datos inválidos"})
		return
	}

	// Actualización selectiva
	if req.Name != "" {
		supplier.Name = strings.TrimSpace(req.Name)
	}
	if req.Email != "" {
		supplier.Email = strings.ToLower(strings.TrimSpace(req.Email))
	}
	supplier.Contact = strings.TrimSpace(req.Contact)
	supplier.Phone = strings.TrimSpace(req.Phone)
	supplier.Address = strings.TrimSpace(req.Address)

	if err := db.Save(&supplier).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al actualizar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "supplier": supplier})
}

// DeleteSupplier elimina un proveedor si no tiene productos vinculados
func DeleteSupplier(c *gin.Context) {
	id := c.Param("id")
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	// 1. Verificar existencia y pertenencia al tenant
	var supplier models.Supplier
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&supplier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Proveedor no encontrado"})
		return
	}

	// 2. Verificar integridad (si tiene productos asociados)
	var productCount int64
	db.Model(&models.Product{}).Where("supplier_id = ? AND tenant_id = ?", id, tenantID).Count(&productCount)
	if productCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "No se puede eliminar el proveedor porque tiene productos asociados",
		})
		return
	}

	// 3. Eliminación (GORM usará Soft Delete si DeletedAt está en el modelo)
	if err := db.Delete(&supplier).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al eliminar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Proveedor eliminado correctamente"})
}
