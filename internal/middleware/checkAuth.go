package middleware

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func IsSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		user := val.(models.User)

		// Use a local variable to make it cleaner for the linter
		// and check if the Name is empty (meaning Preload failed)
		roleName := user.Role.Name

		if roleName != "Superadmin" && roleName != "Admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied: Administrative privileges required",
			})
			return
		}

		c.Next()
	}
}

func JwtValidate(c *gin.Context) {
	// 1. Get token from Cookie (or Authorization Header)
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No authorization cookie found"})
		return
	}

	// 2. Parse and Validate Token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRETKEY")), nil
	})

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// 3. Extract Claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Check expiration
		if exp, ok := claims["exp"].(float64); ok {
			if float64(time.Now().Unix()) > exp {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
				return
			}
		}

		db, err := database.IntialDB()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// 4. Find User and Preload Role (CRITICAL for IsSuperAdmin)
		var user models.User

		// In Go, claims numbers are float64 by default. Convert to int.
		userID := int(claims["sub"].(float64))
		tenantID := int(claims["tenant_id"].(float64))

		// Preload("Role") is required so the Role.Name field isn't empty in IsSuperAdmin
		result := db.Preload("Role").Where("id = ? AND tenant_id = ?", userID, tenantID).First(&user)

		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User session invalid or tenant mismatch"})
			return
		}

		// 5. Context Injection
		// Injecting the full object for controllers
		c.Set("user", user)

		// Injecting the ID as an INT for the TenantContextGuard in main.go
		c.Set("tenantID", user.TenantID)

		c.Next()
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
