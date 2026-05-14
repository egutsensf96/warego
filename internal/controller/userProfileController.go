// internal/controller/userProfileController.go
package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/middleware"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GetUserProfile - Get current authenticated user's profile
// Endpoint: GET /api/v1/user/profile
func GetUserProfile(c *gin.Context) {
	// ✅ Extract user ID from JWT context (set by JwtValidate middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	db := database.GetDB()
	var user models.User

	// ✅ Fetch user with relations, scoped to their tenant
	if err := db.Preload("Role").Preload("Tenant").Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile"})
		return
	}

	// ✅ Update last login timestamp (async, non-blocking)
	go func() {
		now := time.Now()
		db.Model(&user).Update("last_login", &now)
	}()

	// ✅ Hide sensitive fields before response
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user":        user,
			"permissions": getPermissionsForRole(user.Role),
		},
	})
}

// UpdateUserProfile - Update current user's profile details
// Endpoint: PUT /api/v1/user/profile
func UpdateUserProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	db := database.GetDB()
	var user models.User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// ✅ Bind and validate request body
	var req struct {
		Name            string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
		Email           string `json:"email,omitempty" binding:"omitempty,email"`
		CurrentPassword string `json:"current_password,omitempty"` // Required to change password/email
		NewPassword     string `json:"new_password,omitempty" binding:"omitempty,min=6,max=128"`
		AvatarBase64    string `json:"avatar_base64,omitempty"` // Optional: store as text or upload to S3
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ✅ If changing sensitive fields (email/password), require current password
	isSensitiveChange := req.Email != "" || req.NewPassword != ""
	if isSensitiveChange && req.CurrentPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password required to change sensitive fields"})
		return
	}

	// ✅ Verify current password if provided
	if req.CurrentPassword != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
			return
		}
	}

	// ✅ Validate email uniqueness (if changing)
	if req.Email != "" {
		email := strings.ToLower(strings.TrimSpace(req.Email))
		var existing models.User
		if err := db.Where("email = ? AND id != ? AND tenant_id = ?", email, userID, user.TenantID).First(&existing).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already in use by another user in your organization"})
			return
		}
		user.Email = email
	}

	// ✅ Hash and update new password (if provided)
	if req.NewPassword != "" {
		hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
			return
		}
		user.Password = string(hashedPass)
	}

	// ✅ Update other fields
	if req.Name != "" {
		user.Name = strings.TrimSpace(req.Name)
	}
	if req.AvatarBase64 != "" {
		// Optional: Validate size, store in separate table or S3
		// For now, store directly (consider limits in production)
		user.ImageBase64 = req.AvatarBase64 // Assuming you add this field to User model
	}

	// ✅ Save changes
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// ✅ Hide password before response
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile updated successfully",
		"data": gin.H{
			"user": user,
		},
	})
}

// Endpoint: POST /api/v1/user/logout
func Logout(c *gin.Context) {

	_ = c.GetHeader("Authorization") // Token extraction for optional blacklist (currently unused)

	// ✅ Optional: Add token to blacklist (requires Redis/DB storage)
	// For stateless JWT, logout is client-side (delete token from storage)
	// If you want server-side invalidation, implement a blacklist:

	// Example with Redis (pseudo-code):
	/*
		authHeader := c.GetHeader("Authorization")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString != "" {
			// Parse token to get expiration
			token, _ := jwt.Parse(tokenString, ...)
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if exp, ok := claims["exp"].(float64); ok {
					expiry := time.Unix(int64(exp), 0)
					// Store token JTI or full token in Redis with TTL = expiry - now
					// redis.Set("blacklist:"+tokenString, "1", expiry.Sub(time.Now()))
				}
			}
		}
	*/

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully. Please delete the token from your client storage.",
	})
}

// Helper: Return permissions based on role name
func getPermissionsForRole(role *models.Role) []string {
	if role == nil {
		return []string{}
	}

	// Define permissions per role (extend as needed)
	permissions := map[string][]string{
		"super-admin": {
			"tenant:read", "tenant:write", "tenant:delete",
			"user:read:global", "user:write:global", "user:delete:global",
			"product:read:global", "inventory:manage:global",
		},
		"admin": {
			"user:read", "user:write", // within tenant
			"product:read", "product:write", "inventory:manage",
			"category:manage", "warehouse:manage", "supplier:manage",
		},
		"staff": {
			"product:read", "inventory:view",
			"draw:create", "draw:read",
		},
	}

	if perms, ok := permissions[strings.ToLower(role.Name)]; ok {
		return perms
	}
	return []string{} // Default: no permissions
}
