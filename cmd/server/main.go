package main

import (
	"log"
	"net/http"

	"mrs_sendemail_be/internal/config"
	"mrs_sendemail_be/internal/handlers"
	"mrs_sendemail_be/internal/middleware"
	"mrs_sendemail_be/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize services
	redisService := services.NewRedisService(cfg)
	smtpService := services.NewSMTPService(cfg)

	// Test connections at startup
	log.Println("Testing service connections...")
	if err := redisService.Ping(nil); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	} else {
		log.Println("✓ Redis connection successful")
	}

	if err := smtpService.TestConnection(); err != nil {
		log.Printf("Warning: SMTP connection failed: %v", err)
	} else {
		log.Println("✓ SMTP connection successful")
	}

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(redisService, smtpService)
	generateHandler := handlers.NewGenerateHandler(cfg, redisService, smtpService)
	verifyHandler := handlers.NewVerifyHandler(redisService)

	// Setup Gin router
	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS middleware (cho phép cross-origin requests)
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, x-api-key")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Public routes (không cần API key)
	router.GET("/health", healthHandler.HealthCheck)

	// Protected routes (cần API key)
	protected := router.Group("/")
	protected.Use(middleware.APIKeyAuth(cfg))
	{
		// Generate endpoint với rate limiting
		generateGroup := protected.Group("/")
		generateGroup.Use(middleware.RateLimit(redisService))
		generateGroup.Use(middleware.EmailRateLimit(redisService))
		generateGroup.POST("/generate", generateHandler.Generate)

		// Verify endpoint (chỉ cần API key, không cần rate limiting)
		protected.POST("/verify", verifyHandler.Verify)
	}

	// Start server
	address := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting server on %s", address)
	log.Printf("Health check: GET http://%s/health", address)
	log.Printf("Generate code: POST http://%s/generate", address)
	log.Printf("Verify code: POST http://%s/verify", address)

	if err := router.Run(address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
