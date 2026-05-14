// internal/controller/authController.go
package controller

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// SignUp handles POST /api/v1/auth/register
func SignUp(c *gin.Context) {
	var body struct {
		Name        string    `json:"name" binding:"required,min=2,max=100"`
		Email       string    `json:"email" binding:"required,email"`
		Password    string    `json:"password" binding:"required,min=8,max=128"`
		TenantID    uuid.UUID `json:"tenant_id" binding:"required"`
		RoleID      uuid.UUID `json:"role_id" binding:"required"`
		ImageBase64 string    `json:"image_base64,omitempty"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	db := database.GetDB()

	// Check for duplicate email within tenant
	var existing models.User
	if err := db.Where("email = ? AND tenant_id = ?", strings.ToLower(body.Email), body.TenantID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists in this tenant"})
		return
	}

	// Validate role belongs to tenant
	var role models.Role
	if err := db.Where("id = ? AND tenant_id = ?", body.RoleID, body.TenantID).First(&role).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role for this tenant"})
		return
	}

	user := models.User{
		Name:        strings.TrimSpace(body.Name),
		Email:       strings.ToLower(strings.TrimSpace(body.Email)),
		Password:    string(hash),
		TenantID:    body.TenantID,
		RoleID:      body.RoleID,
		ImageBase64: body.ImageBase64,
		IsActive:    true,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User created successfully",
		"user_id": user.ID,
	})
}

// SignIn handles POST /api/v1/auth/login
func SignIn(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var user models.User
	if err := db.Preload("Role").Preload("Tenant").
		Where("email = ? AND is_active = true AND deleted_at IS NULL",
			strings.ToLower(req.Email)).
		First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Update last login timestamp (async)
	go func() {
		now := time.Now()
		db.Model(&user).Update("last_login", &now)
	}()

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   user.ID.String(),
		"email":     strings.ToLower(strings.TrimSpace(req.Email)),
		"name":      user.Name,
		"role":      user.Role.Name,
		"tenant_id": user.TenantID.String(),
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	})

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-jwt-secret-change-in-production"
	}

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"token":   tokenString,
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"name":      user.Name,
			"role":      user.Role.Name,
			"tenant_id": user.TenantID,
			"avatar":    user.ImageBase64,
		},
	})
}

// CheckAuth handles GET /api/v1/auth/check
func CheckAuth(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	db := database.GetDB()
	var user models.User
	if err := db.Preload("Role").Preload("Tenant").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Password = "" // Hide password
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user": user,
		},
	})
}
