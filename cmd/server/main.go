// @title RcloneStorage API
// @version 1.0.0
// @description Cloud Storage API using Rclone with multiple providers support, authentication, and video streaming capabilities.
// @termsOfService http://swagger.io/terms/

// @contact.name RcloneStorage API Support
// @contact.email support@rclonestorage.local

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT token. Format: Bearer {token}

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API Key for authentication

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/nabilulilalbab/rclonestorage/docs"
	"github.com/nabilulilalbab/rclonestorage/internal/api"
	"github.com/nabilulilalbab/rclonestorage/internal/auth"
	"github.com/nabilulilalbab/rclonestorage/internal/config"
	"github.com/nabilulilalbab/rclonestorage/internal/monitoring"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize authentication system
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key-change-in-production"
		log.Println("Warning: Using default JWT secret. Set JWT_SECRET environment variable in production.")
	}

	authManager, err := auth.NewAuthManager("./data/auth.db", jwtSecret)
	if err != nil {
		log.Fatalf("Failed to initialize authentication: %v", err)
	}
	defer authManager.Close()

	// Setup Gin router
	r := gin.Default()

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	// Setup static file serving for web interface
	r.Static("/static", "./web/static")
	r.StaticFile("/", "./web/templates/index.html")
	r.StaticFile("/login.html", "./web/templates/login.html")
	r.StaticFile("/register.html", "./web/templates/register.html")
	r.StaticFile("/upload.html", "./web/templates/upload.html")
	r.StaticFile("/files.html", "./web/templates/files.html")
	r.StaticFile("/stream.html", "./web/templates/stream.html")
	r.StaticFile("/profile.html", "./web/templates/profile.html")
	r.StaticFile("/dashboard.html", "./web/templates/dashboard.html")

	// Setup authentication routes
	authManager.SetupAuthRoutes(r)

	// Setup API routes with authentication
	api.SetupRoutes(r, cfg, authManager)

	// Setup monitoring dashboard
	monitoringDashboard := monitoring.NewMonitoringDashboard(cfg, authManager)
	monitoringDashboard.SetupRoutes(r)

	// Setup Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Health check endpoint (public)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "rclonestorage",
			"version": "1.0.0",
			"features": []string{
				"multi-provider storage",
				"video streaming",
				"authentication",
				"api keys",
				"monitoring dashboard",
				"swagger documentation",
			},
		})
	})

	// Create data directory if not exists
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Start server
	port := cfg.Server.Port
	if port == "" {
		port = "5601"
	}

	log.Printf("Starting RcloneStorage server on port %s", port)
	log.Printf("Default admin credentials: admin@rclonestorage.local / Admin123!")
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
