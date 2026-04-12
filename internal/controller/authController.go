package controller

import (
	"net/http"
	"os"
	"time"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func SignUp(c *gin.Context) {
	var body struct {
		Email       string    `json:"email" binding:"required,email"`
		Password    string    `json:"password" binding:"required,min=8"`
		TenantID    uuid.UUID `json:"tenant_id" binding:"required"`
		RoleID      uuid.UUID `json:"role_id" binding:"required"`
		ImageBase64 string    `json:"image_base64"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	db := database.GetDB()

	user := models.User{
		Email:       body.Email,
		Password:    string(hash),
		TenantID:    body.TenantID,
		RoleID:      body.RoleID,
		ImageBase64: body.ImageBase64,
	}

	if result := db.Create(&user); result.Error != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists or invalid tenant/role"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user_id": user.ID,
	})
}

func SingIn(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password required"})
		return
	}

	db := database.GetDB()
	var user models.User

	// 1. Find user by email
	if err := db.Where("email = ?", body.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// 2. Compare password with hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// 3. Generate JWT with Tenant Context
	// We convert UUIDs to strings for the JWT claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":       user.ID.String(),
		"tenant_id": user.TenantID.String(), // Critical for multi-tenancy middleware
		"role_id":   user.RoleID.String(),
		"exp":       time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRETKEY")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// 4. Set Cookie & Response
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*7, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"token":     tokenString,
		"tenant_id": user.TenantID,
		"user": gin.H{
			"email":        user.Email,
			"image_base64": user.ImageBase64,
		},
	})
}

func CheckAuth(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
