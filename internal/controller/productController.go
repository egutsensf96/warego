// internal/controller/productController.go
package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/middleware"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetProducts retrieves products (SuperAdmin sees all, others see tenant-only)
func GetProducts(c *gin.Context) {
	db := database.GetDB()
	tenantID := middleware.GetTenantID(c)

	// ✅ Iniciamos la query explícitamente sobre el modelo Product
	query := db.Model(&models.Product{})

	// 1. Filtro base por tenant (evita ambigüedad especificando la tabla)
	if !middleware.IsSuperAdminUser(c) {
		query = query.Where("products.tenant_id = ?", tenantID)
	}

	// 2. Filtro opcional por almacén
	if warehouseID := c.Query("warehouse_id"); warehouseID != "" {
		if wid, err := uuid.Parse(warehouseID); err == nil {
			// ✅ JOIN explícito + DISTINCT para evitar filas duplicadas
			query = query.Joins(
				"INNER JOIN trackers ON trackers.product_id = products.id AND trackers.warehouse_id = ? AND trackers.tenant_id = ?",
				wid, tenantID,
			).Distinct("products.*")
		}
	}

	var products []models.Product
	if err := query.Preload("Category").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Could not fetch products"})
		return
	}

	// ✅ Claves JSON limpias (sin espacios trailing que rompían Flutter)
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"count":    len(products),
		"products": products,
	})
}

// CreateProductInput represents the payload for creating a product
type CreateProductInput struct {
	models.Product
	InitialWarehouseID uuid.UUID `json:"warehouse_id" binding:"required"`
	// SupplierID *uuid.UUID `json:"supplier_id,omitempty"` // Descomenta si agregaste el campo a models.Product
}

// CreateProduct adds a new product linked to the tenant + logs initial movement
func CreateProduct(c *gin.Context) {
	var input CreateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	tenantID := middleware.GetTenantID(c)
	input.Product.TenantID = uuid.MustParse(tenantID)

	db := database.GetDB()

	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. Crear Producto
		if err := tx.Create(&input.Product).Error; err != nil {
			return err
		}

		// 2. Crear Tracker (Stock actual por almacén)
		tracker := models.Tracker{
			ProductID:   input.Product.ID,
			WarehouseID: input.InitialWarehouseID,
			Quantity:    input.Product.Quantity,
			TenantID:    input.Product.TenantID,
		}
		if err := tx.Create(&tracker).Error; err != nil {
			return err
		}

		// 3. Log de movimiento inicial
		userIDStr, _ := c.Get("userID")
		var userID *uuid.UUID
		if userIDStr != nil {
			if parsed, err := uuid.Parse(userIDStr.(string)); err == nil {
				userID = &parsed
			}
		}

		movement := models.StockTransaction{
			ProductID:      input.Product.ID,
			WarehouseID:    input.InitialWarehouseID,
			TenantID:       input.Product.TenantID,
			Type:           models.TypeInitial,
			QuantityChange: input.Product.Quantity,
			Notes:          "Product registered with initial stock",
			UserID:         userID,
			// SupplierID:     input.SupplierID, // Descomenta si el campo existe
		}
		return tx.Create(&movement).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Product created successfully",
		"product": input.Product,
	})
}

func UpdateProductWithAudit(c *gin.Context) {
	idParam := c.Param("id")
	productID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid product ID"})
		return
	}

	tenantID := middleware.GetTenantID(c)
	userIDStr, _ := c.Get("userID")
	userID := uuid.Nil
	if userIDStr != nil {
		userID, _ = uuid.Parse(userIDStr.(string))
	}

	db := database.GetDB()

	// ✅ Corregido typo en nombre de variable
	var originalProduct models.Product
	if err := db.Where("id = ? AND tenant_id = ?", productID, tenantID).First(&originalProduct).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Product not found"})
		return
	}

	var req struct {
		Name        string     `json:"name,omitempty"`
		SKU         string     `json:"sku,omitempty"`
		Quantity    *int       `json:"quantity,omitempty"`
		CategoryID  *uuid.UUID `json:"category_id,omitempty"`
		ImageBase64 string     `json:"image_base64,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		updates := make(map[string]interface{})
		if req.Name != "" {
			updates["name"] = strings.TrimSpace(req.Name)
		}
		if req.SKU != "" {
			updates["sku"] = strings.ToUpper(strings.TrimSpace(req.SKU))
		}
		if req.CategoryID != nil {
			updates["category_id"] = *req.CategoryID
		}
		if req.ImageBase64 != "" {
			updates["image_base64"] = req.ImageBase64
		}

		var quantityChanged bool
		var oldQty, newQty int
		if req.Quantity != nil {
			oldQty = originalProduct.Quantity // ✅ Corregido typo
			newQty = *req.Quantity
			quantityChanged = oldQty != newQty
			updates["quantity"] = newQty
		}

		if len(updates) > 0 {
			if err := tx.Model(&originalProduct).Where("id = ?", productID).Updates(updates).Error; err != nil {
				return err
			}
		}

		// Log de cambio de cantidad a StockTransaction
		if quantityChanged && userID != uuid.Nil {
			var trackers []models.Tracker
			if err := tx.Where("product_id = ? AND tenant_id = ?", productID, tenantID).Find(&trackers).Error; err != nil {
				return err
			}

			for _, tracker := range trackers {
				change := newQty - oldQty
				txn := models.StockTransaction{
					ProductID:      productID,
					WarehouseID:    tracker.WarehouseID,
					TenantID:       uuid.MustParse(tenantID),
					QuantityChange: change,
					Type:           models.TypeAdjust,
					UserID:         &userID,
					Notes:          fmt.Sprintf("Product edit: quantity %d → %d", oldQty, newQty),
				}
				if err := tx.Create(&txn).Error; err != nil {
					return err
				}

				// Actualizar tracker atómicamente
				if err := tx.Model(&tracker).Update("quantity", gorm.Expr("quantity + ?", change)).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update product: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Product updated with audit log",
		"product": originalProduct,
	})
}

func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	result := db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Product{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Product not found or unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Product deleted successfully",
	})
}
