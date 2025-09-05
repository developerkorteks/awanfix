package auth

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthHandlers handles authentication-related HTTP requests
type AuthHandlers struct {
	jwtManager *JWTManager
	dbManager  *DatabaseManager
}

// NewAuthHandlers creates new authentication handlers
func NewAuthHandlers(jwtManager *JWTManager, dbManager *DatabaseManager) *AuthHandlers {
	return &AuthHandlers{
		jwtManager: jwtManager,
		dbManager:  dbManager,
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role,omitempty"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

// UserInfo represents user information (without sensitive data)
type UserInfo struct {
	ID           uint    `json:"id"`
	Email        string  `json:"email"`
	Role         string  `json:"role"`
	StorageUsed  int64   `json:"storage_used"`
	StorageQuota int64   `json:"storage_quota"`
	UsagePercent float64 `json:"usage_percent"`
	CreatedAt    string  `json:"created_at"`
}

// APIKeyRequest represents an API key creation request
type APIKeyRequest struct {
	Name string `json:"name" binding:"required"`
}

// APIKeyResponse represents an API key response
type APIKeyResponse struct {
	ID        uint      `json:"id"`
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Register handles user registration
// @Summary User registration
// @Description Register a new user account
// @Tags authentication
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration data"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Router /../auth/register [post]
func (ah *AuthHandlers) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Set default role if not specified
	if req.Role == "" {
		req.Role = RoleUser
	}

	// Only admins can create admin users
	if req.Role == RoleAdmin {
		if !IsAdmin(c) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Only administrators can create admin users",
			})
			return
		}
	}

	// Create user
	user, err := ah.dbManager.CreateUser(req.Email, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user": UserInfo{
			ID:           user.ID,
			Email:        user.Email,
			Role:         user.Role,
			StorageUsed:  user.StorageUsed,
			StorageQuota: user.StorageQuota,
			UsagePercent: user.GetStorageUsagePercent(),
			CreatedAt:    user.CreatedAt.Format(time.RFC3339),
		},
	})
}

// Login handles user login
// @Summary User login
// @Description Authenticate user and get JWT token
// @Tags authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "User credentials"
// @Success 200 {object} LoginResponse "Login successful"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Router /../auth/login [post]
func (ah *AuthHandlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Authenticate user
	user, err := ah.dbManager.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// Generate JWT token
	token, err := ah.jwtManager.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour),
		User: UserInfo{
			ID:           user.ID,
			Email:        user.Email,
			Role:         user.Role,
			StorageUsed:  user.StorageUsed,
			StorageQuota: user.StorageQuota,
			UsagePercent: user.GetStorageUsagePercent(),
			CreatedAt:    user.CreatedAt.Format(time.RFC3339),
		},
	})
}

// RefreshToken handles token refresh
func (ah *AuthHandlers) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Authorization header required",
		})
		return
	}

	token := authHeader[7:] // Remove "Bearer " prefix
	newToken, err := ah.jwtManager.RefreshToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Cannot refresh token",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      newToken,
		"expires_at": time.Now().Add(time.Hour),
	})
}

// GetProfile returns current user profile
// @Summary Get user profile
// @Description Get current user's profile information
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserInfo "User profile"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /../user/profile [get]
func (ah *AuthHandlers) GetProfile(c *gin.Context) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	c.JSON(http.StatusOK, UserInfo{
		ID:           user.ID,
		Email:        user.Email,
		Role:         user.Role,
		StorageUsed:  user.StorageUsed,
		StorageQuota: user.StorageQuota,
		UsagePercent: user.GetStorageUsagePercent(),
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
	})
}

// CreateAPIKey creates a new API key
// @Summary Create API key
// @Description Create a new API key for the current user
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param apikey body APIKeyRequest true "API key data"
// @Success 201 {object} APIKeyResponse "API key created successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /../user/api-keys [post]
func (ah *AuthHandlers) CreateAPIKey(c *gin.Context) {
	var req APIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	userID, exists := GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	apiKey, err := ah.dbManager.CreateAPIKey(userID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create API key",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, APIKeyResponse{
		ID:        apiKey.ID,
		Key:       apiKey.Key,
		Name:      apiKey.Name,
		CreatedAt: apiKey.CreatedAt,
	})
}

// ListAPIKeys lists user's API keys
// @Summary List API keys
// @Description Get all API keys for the current user
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "List of API keys"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /../user/api-keys [get]
func (ah *AuthHandlers) ListAPIKeys(c *gin.Context) {
	userID, exists := GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	apiKeys, err := ah.dbManager.ListAPIKeys(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list API keys",
		})
		return
	}

	var response []gin.H
	for _, key := range apiKeys {
		response = append(response, gin.H{
			"id":         key.ID,
			"name":       key.Name,
			"last_used":  key.LastUsed,
			"created_at": key.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"api_keys": response,
		"total":    len(response),
	})
}

// ChangePassword changes user password
// @Summary Change user password
// @Description Change current user's password
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]string true "Password change request"
// @Success 200 {object} map[string]interface{} "Password changed successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /../user/change-password [post]
func (ah *AuthHandlers) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var request struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Get current user
	var user User
	if err := ah.dbManager.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Create password manager
	pm := NewPasswordManager()

	// Verify current password
	if err := pm.CheckPassword(request.CurrentPassword, user.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password (validation is done inside HashPassword)
	hashedPassword, err := pm.HashPassword(request.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password
	if err := ah.dbManager.db.Model(&user).Update("password", hashedPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// DeleteAPIKey deletes an API key
// @Summary Delete API key
// @Description Delete an API key by ID
// @Tags user
// @Produce json
// @Security BearerAuth
// @Param id path int true "API Key ID"
// @Success 200 {object} map[string]interface{} "API key deleted successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "API key not found"
// @Router /../user/api-keys/{id} [delete]
func (ah *AuthHandlers) DeleteAPIKey(c *gin.Context) {
	keyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid API key ID",
		})
		return
	}

	userID, exists := GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	if err := ah.dbManager.DeleteAPIKey(uint(keyID), userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "API key not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API key deleted successfully",
	})
}

// ListUsers lists all users (admin only)
// @Summary List all users
// @Description Get list of all users (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "List of users"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /../admin/users [get]
func (ah *AuthHandlers) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	users, total, err := ah.dbManager.ListUsers(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list users",
		})
		return
	}

	var response []UserInfo
	for _, user := range users {
		response = append(response, UserInfo{
			ID:           user.ID,
			Email:        user.Email,
			Role:         user.Role,
			StorageUsed:  user.StorageUsed,
			StorageQuota: user.StorageQuota,
			UsagePercent: user.GetStorageUsagePercent(),
			CreatedAt:    user.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"users": response,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetUser gets a specific user (admin only)
func (ah *AuthHandlers) GetUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	user, err := ah.dbManager.GetUserByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, UserInfo{
		ID:           user.ID,
		Email:        user.Email,
		Role:         user.Role,
		StorageUsed:  user.StorageUsed,
		StorageQuota: user.StorageQuota,
		UsagePercent: user.GetStorageUsagePercent(),
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
	})
}