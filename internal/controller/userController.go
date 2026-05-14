// internal/controller/userController.go
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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GetUsers obtiene todos los usuarios del tenant actual
func GetUsers(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var users []models.User
	// Preload("Role") permite traer los detalles del rol en la misma consulta
	if err := db.Where("tenant_id = ?", tenantID).Preload("Role").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al obtener usuarios"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"users":   users,
	})
}

// GetUser obtiene un usuario por ID (validando tenant)
func GetUser(c *gin.Context) {
	id := c.Param("id")
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var user models.User
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).Preload("Role").First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Usuario no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al buscar usuario"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "user": user})
}

// CreateUser crea un nuevo usuario con contraseña encriptada
func CreateUser(c *gin.Context) {
	tenantIDStr := middleware.GetTenantID(c)
	tenantUUID, _ := uuid.Parse(tenantIDStr)

	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		RoleID   string `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Datos inválidos o contraseña muy corta"})
		return
	}

	db := database.GetDB()
	emailClean := strings.ToLower(strings.TrimSpace(req.Email))

	// 1. Verificar si el email ya existe
	var exists int64
	db.Model(&models.User{}).Where("email = ?", emailClean).Count(&exists)
	if exists > 0 {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "El correo electrónico ya está registrado"})
		return
	}

	// 2. Encriptar contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al procesar credenciales"})
		return
	}

	roleUUID, _ := uuid.Parse(req.RoleID)
	user := models.User{
		Name:     strings.TrimSpace(req.Name),
		Email:    emailClean,
		Password: string(hashedPassword),
		RoleID:   roleUUID,
		TenantID: tenantUUID,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al crear usuario"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "user": user})
}

// UpdateUser actualiza datos del usuario
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	tenantID := middleware.GetTenantID(c)
	db := database.GetDB()

	var user models.User
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Usuario no encontrado"})
		return
	}

	var req struct {
		Name     string `json:"name"`
		RoleID   string `json:"role_id"`
		Password string `json:"password"` // Opcional
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Datos inválidos"})
		return
	}

	if req.Name != "" {
		user.Name = strings.TrimSpace(req.Name)
	}
	if req.RoleID != "" {
		roleUUID, _ := uuid.Parse(req.RoleID)
		user.RoleID = roleUUID
	}
	if req.Password != "" {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		user.Password = string(hashed)
	}

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al actualizar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "user": user})
}

// DeleteUser elimina (Soft Delete) un usuario
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	tenantID := middleware.GetTenantID(c)
	currentUserID := middleware.GetUserID(c) // No permitir borrarse a sí mismo

	if id == currentUserID {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "No puedes eliminar tu propia cuenta"})
		return
	}

	db := database.GetDB()
	result := db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.User{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al eliminar"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Usuario no encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Usuario eliminado correctamente"})
}
