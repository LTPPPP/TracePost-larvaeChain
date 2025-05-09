package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"strings"
	"strconv"
	
	// Import Swagger docs
	_ "github.com/LTPPPP/TracePost-larvaeChain/docs"
	
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
	"github.com/LTPPPP/TracePost-larvaeChain/api"
	"github.com/LTPPPP/TracePost-larvaeChain/config"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"github.com/LTPPPP/TracePost-larvaeChain/middleware"
)

// @title TracePost-larvaeChain API
// @version 1.0
// @description Traceability system for shrimp larvae using blockchain technology
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@vietchain.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default environment variables")
	}

	// Load configuration
	cfg := config.GetConfig()

	// Initialize database connection
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create a new Fiber app with optimized configuration
	app := fiber.New(fiber.Config{
		AppName:               "TracePost-larvaeChain",
		ErrorHandler:          api.ErrorHandler,
		ReadTimeout:           time.Duration(cfg.ServerTimeout) * time.Second,
		WriteTimeout:          time.Duration(cfg.ServerTimeout) * time.Second,
		IdleTimeout:           time.Duration(getEnvAsInt("SERVER_IDLE_TIMEOUT", 60)) * time.Second,
		BodyLimit:             getEnvAsInt("SERVER_BODY_LIMIT", 10) * 1024 * 1024, // Default 10MB
		Concurrency:           getEnvAsInt("SERVER_CONCURRENCY", 256 * 1024),      // Default 256K
		DisableStartupMessage: getEnvAsBool("DISABLE_STARTUP_MESSAGE", false),
		EnablePrintRoutes:     getEnvAsBool("ENABLE_PRINT_ROUTES", false),
		Prefork:               getEnvAsBool("SERVER_PREFORK", false),
		ReduceMemoryUsage:     getEnvAsBool("SERVER_REDUCE_MEMORY", true),
		EnableTrustedProxyCheck: getEnvAsBool("TRUSTED_PROXY_ENABLED", true),
		TrustedProxies:       strings.Split(getEnv("TRUSTED_PROXIES", "127.0.0.1"), ","),
		GETOnly:              getEnvAsBool("SERVER_GET_ONLY", false),
		CompressedFileSuffix: ".gz",
	})

	// Use global middlewares
	app.Use(recover.New())
	app.Use(middleware.LoggerMiddleware())
	
	// Security middleware
	app.Use(func(c *fiber.Ctx) error {
		// Add security headers
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Set("Content-Security-Policy", "default-src 'self'")
		c.Set("Referrer-Policy", "no-referrer")
		c.Set("Feature-Policy", "camera 'none'; microphone 'none'")
		c.Set("X-DNS-Prefetch-Control", "off")
		
		return c.Next()
	})
	
	// CORS configuration
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-DID, X-DID-Proof",
		ExposeHeaders:    "Content-Length, Authorization",
		AllowCredentials: true,
	}))

	// Setup Swagger
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL:         "/swagger/doc.json",
		DeepLinking: true,
	}))

	// Setup API routes
	api.SetupAPI(app)

	// Print startup message
	startupMessage(cfg)

	// Start the server
	log.Fatal(app.Listen(":" + cfg.ServerPort))
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// startupMessage prints a startup message with the server configuration
func startupMessage(cfg *config.Config) {
	fmt.Println("┌─────────────────────────────────────────────────────┐")
	fmt.Println("│                 TracePost-larvaeChain               │")
	fmt.Println("├─────────────────────────────────────────────────────┤")
	fmt.Println("│ Shrimp Larvae Traceability System                   │")
	fmt.Println("│ Built with Go, Fiber, and Blockchain Technology     │")
	fmt.Println("├─────────────────────────────────────────────────────┤")
	fmt.Printf("│ HTTP Server running on port %-24s │\n", cfg.ServerPort)
	fmt.Printf("│ Swagger UI available at http://localhost:%s/swagger  │\n", cfg.ServerPort)
	fmt.Println("├─────────────────────────────────────────────────────┤")
	fmt.Printf("│ Environment: %-38s │\n", os.Getenv("GO_ENV"))
	fmt.Println("└─────────────────────────────────────────────────────┘")
}