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
		// 1. Get user from context (Set by JwtValidate)
		val, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		user := val.(models.User)

		// 2. Check Role (Assuming your "Admin" role has this description)
		// You can also check if their TenantID matches a master System Tenant
		if user.Role.Description != "Superadmin" && user.Role.Description != "Admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied: Superadmin privileges required",
			})
			return
		}

		c.Next()
	}
}

func JwtValidate(c *gin.Context) {
	// 1. Get token from Cookie
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No authorization cookie found"})
		return
	}

	// 2. Parse and Validate Token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// HMAC secret must be []byte
		return []byte(os.Getenv("SECRETKEY")), nil
	})

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// 3. Extract Claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Check expiration
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			return
		}

		// Initialize DB
		db, err := database.IntialDB()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
			return
		}

		// 4. Find User and verify Tenant
		var user models.User
		// We use 'sub' for the user ID and 'tenant_id' from the token claims
		userID := claims["sub"]
		tenantID := claims["tenant_id"]

		result := db.Where("id = ? AND tenant_id = ?", userID, tenantID).First(&user)
		if result.Error != nil || user.ID.String() == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User no longer exists or tenant mismatch"})
			return
		}

		// 5. Context Injection
		// Save the full user object for use in controllers (CheckAuth, etc.)
		c.Set("user", user)

		// Optional: Automatically set the tenantID in context so controllers don't have to read headers
		c.Set("tenantID", user.TenantID.String())

		c.Next()
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
