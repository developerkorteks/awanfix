package auth

import (
	"time"

	"github.com/gin-gonic/gin"
)

// AuthManager manages all authentication components
type AuthManager struct {
	DatabaseManager *DatabaseManager
	JWTManager      *JWTManager
	Middleware      *AuthMiddleware
	Handlers        *AuthHandlers
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(dbPath, jwtSecret string) (*AuthManager, error) {
	// Initialize database manager
	dbManager, err := NewDatabaseManager(dbPath)
	if err != nil {
		return nil, err
	}

	// Initialize JWT manager (1 hour token duration)
	jwtManager := NewJWTManager(jwtSecret, time.Hour)

	// Initialize middleware
	middleware := NewAuthMiddleware(jwtManager, dbManager)

	// Initialize handlers
	handlers := NewAuthHandlers(jwtManager, dbManager)

	return &AuthManager{
		DatabaseManager: dbManager,
		JWTManager:      jwtManager,
		Middleware:      middleware,
		Handlers:        handlers,
	}, nil
}

// SetupAuthRoutes sets up authentication routes
func (am *AuthManager) SetupAuthRoutes(r *gin.Engine) {
	// Public authentication routes
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", am.Handlers.Register)
		auth.POST("/login", am.Handlers.Login)
		auth.POST("/refresh", am.Handlers.RefreshToken)
	}

	// Protected user routes - Support both JWT and API key
	user := r.Group("/api/user")
	user.Use(am.Middleware.OptionalAuth())
	user.Use(am.Middleware.RequireAuth())
	{
		user.GET("/profile", am.Handlers.GetProfile)
		user.POST("/api-keys", am.Handlers.CreateAPIKey)
		user.GET("/api-keys", am.Handlers.ListAPIKeys)
		user.DELETE("/api-keys/:id", am.Handlers.DeleteAPIKey)
	}

	// Admin-only routes - Support both JWT and API key
	admin := r.Group("/api/admin")
	admin.Use(am.Middleware.OptionalAuth())
	admin.Use(am.Middleware.RequireAuth())
	admin.Use(am.Middleware.RequireRole(RoleAdmin))
	{
		admin.GET("/users", am.Handlers.ListUsers)
		admin.GET("/users/:id", am.Handlers.GetUser)
		admin.POST("/users", am.Handlers.Register) // Admin can create users
	}
}

// Close closes the authentication manager
func (am *AuthManager) Close() error {
	return am.DatabaseManager.Close()
}