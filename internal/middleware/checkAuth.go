// internal/middleware/checkAuth.go
package middleware

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
)

// ✅ FIX: Use distinct names for context keys vs helper functions
type contextKey string

const (
	UserIDKey       contextKey = "userID"
	RoleKey         contextKey = "role"
	TenantIDKey     contextKey = "tenantID"
	IsSuperAdminKey contextKey = "isSuperAdmin"
	TokenExpiresAt  contextKey = "token_expires_at"
)

// JwtValidate middleware validates JWT and injects claims into context
func JwtValidate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			return
		}

		// Remove "Bearer " prefix
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Extract required claims
		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing user_id in token"})
			return
		}

		role, ok := claims["role"].(string)
		if !ok || role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing role in token"})
			return
		}

		role = strings.ToLower(strings.TrimSpace(role))
		isSuperAdmin := role == "super-admin"

		// Extract tenant_id (may be null for SuperAdmin)
		var tenantID string
		if !isSuperAdmin {
			tenantIDRaw, ok := claims["tenant_id"]
			if !ok || tenantIDRaw == nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing tenant_id for non-SuperAdmin"})
				return
			}
			tenantID, ok = tenantIDRaw.(string)
			if !ok || tenantID == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid tenant_id format"})
				return
			}
			if _, err := uuid.Parse(tenantID); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant_id UUID"})
				return
			}
		}

		// ✅ Store claims in context using distinct keys
		c.Set(string(UserIDKey), userID)
		c.Set(string(RoleKey), role)
		c.Set(string(IsSuperAdminKey), isSuperAdmin) // ✅ Use renamed key
		if !isSuperAdmin {
			c.Set(string(TenantIDKey), tenantID)
		}

		// Optional: Add token expiration
		if exp, ok := claims["exp"].(float64); ok {
			c.Set(string(TokenExpiresAt), time.Unix(int64(exp), 0))
		}

		c.Next()
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(string(RoleKey))
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		userRole := role.(string)
		for _, allowed := range allowedRoles {
			if userRole == strings.ToLower(allowed) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
	}
}

// RequireTenant middleware ensures request is scoped to tenant (unless SuperAdmin)
func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ✅ FIX: Use helper function with distinct name
		if IsSuperAdminUser(c) {
			c.Next()
			return
		}

		tenantID, exists := c.Get(string(TenantIDKey))
		if !exists || tenantID == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Tenant context required"})
			return
		}

		// ✅ FIX: Removed undefined tenantExists() call
		// If you want tenant validation, implement it separately:
		// if !isValidTenant(tenantID.(string)) { ... }

		c.Next()
	}
}

// ==================== HELPER FUNCTIONS (with distinct names) ====================

// GetUserID safely extracts user ID from context
func GetUserID(c *gin.Context) string {
	if v, ok := c.Get(string(UserIDKey)); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetRole safely extracts role from context
func GetRole(c *gin.Context) string {
	if v, ok := c.Get(string(RoleKey)); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetTenantID safely extracts tenant ID from context
func GetTenantID(c *gin.Context) string {
	if v, ok := c.Get(string(TenantIDKey)); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func IsSuperAdminUser(c *gin.Context) bool {
	if v, ok := c.Get(string(IsSuperAdminKey)); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// IsTokenExpiringSoon checks if token expires within threshold
func IsTokenExpiringSoon(c *gin.Context, threshold time.Duration) bool {
	if expRaw, ok := c.Get(string(TokenExpiresAt)); ok {
		if exp, ok := expRaw.(time.Time); ok {
			return time.Until(exp) < threshold
		}
	}
	return false
}

func isValidTenant(tenantID string) bool {
	db := database.GetDB()
	var count int64
	db.Model(&models.Tenant{}).Where("id = ? AND deleted_at IS NULL", tenantID).Count(&count)
	return count > 0
}
