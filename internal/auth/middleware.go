package auth

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	jwtManager *JWTManager
	dbManager  *DatabaseManager
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtManager *JWTManager, dbManager *DatabaseManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		dbManager:  dbManager,
	}
}

// JWTAuth middleware for JWT token authentication
func (am *AuthMiddleware) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := am.extractTokenFromHeader(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		claims, err := am.jwtManager.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Get user from database to ensure they're still active
		user, err := am.dbManager.GetUserByID(claims.UserID)
		if err != nil || !user.IsActive {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User account is disabled",
				"code":  "ACCOUNT_DISABLED",
			})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)

		c.Next()
	}
}

// APIKeyAuth middleware for API key authentication
func (am *AuthMiddleware) APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key required",
				"code":  "API_KEY_REQUIRED",
			})
			c.Abort()
			return
		}

		user, err := am.dbManager.ValidateAPIKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
				"code":  "INVALID_API_KEY",
			})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)

		c.Next()
	}
}

// OptionalAuth middleware that allows both authenticated and unauthenticated access
func (am *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try JWT first
		token := am.extractTokenFromHeader(c)
		if token != "" {
			if claims, err := am.jwtManager.ValidateToken(token); err == nil {
				if user, err := am.dbManager.GetUserByID(claims.UserID); err == nil && user.IsActive {
					c.Set("user", user)
					c.Set("user_id", user.ID)
					c.Set("user_role", user.Role)
					c.Next()
					return
				}
			}
		}

		// Try API key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			if user, err := am.dbManager.ValidateAPIKey(apiKey); err == nil {
				c.Set("user", user)
				c.Set("user_id", user.ID)
				c.Set("user_role", user.Role)
				c.Next()
				return
			}
		}

		// No authentication provided, continue as anonymous
		c.Next()
	}
}

// RequireRole middleware that requires a specific role
func (am *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		role := userRole.(string)
		for _, requiredRole := range roles {
			if role == requiredRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
			"code":  "INSUFFICIENT_PERMISSIONS",
		})
		c.Abort()
	}
}

// RequireFileOwnership middleware that checks file ownership
func (am *AuthMiddleware) RequireFileOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		fileID := c.Param("id")
		if fileID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "File ID required",
				"code":  "FILE_ID_REQUIRED",
			})
			c.Abort()
			return
		}

		// Check if user is admin (admins can access all files)
		userRole, _ := c.Get("user_role")
		if userRole == RoleAdmin {
			c.Next()
			return
		}

		// Check file ownership
		_, err := am.dbManager.CheckFileOwnership(fileID, userID.(uint))
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "File not found or access denied",
				"code":  "FILE_ACCESS_DENIED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuditLog middleware that logs user actions
func (am *AuthMiddleware) AuditLog(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Execute the request first
		c.Next()

		// Log after request completion
		userID, exists := c.Get("user_id")
		if !exists {
			return // Skip logging for unauthenticated requests
		}

		resource := c.Request.URL.Path
		ipAddress := c.ClientIP()
		userAgent := c.Request.UserAgent()
		success := c.Writer.Status() < 400

		details := ""
		if !success {
			details = "HTTP " + strconv.Itoa(c.Writer.Status())
		}

		am.dbManager.LogAudit(
			userID.(uint),
			action,
			resource,
			ipAddress,
			userAgent,
			success,
			details,
		)
	}
}

// extractTokenFromHeader extracts JWT token from Authorization header
func (am *AuthMiddleware) extractTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

// GetCurrentUser helper function to get current user from context
func GetCurrentUser(c *gin.Context) (*User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	return user.(*User), true
}

// GetCurrentUserID helper function to get current user ID from context
func GetCurrentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(uint), true
}

// RequireAuth middleware that requires authentication (JWT or API key)
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// IsAdmin helper function to check if current user is admin
func IsAdmin(c *gin.Context) bool {
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}
	return userRole.(string) == RoleAdmin
}