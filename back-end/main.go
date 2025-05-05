package main

import (
	"fmt"
	"log"
	"os"
	"time"
	
	// Import Swagger docs
	_ "github.com/vietchain/tracepost-larvae/docs"
	
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
	"github.com/vietchain/tracepost-larvae/api"
	"github.com/vietchain/tracepost-larvae/config"
	"github.com/vietchain/tracepost-larvae/db"
	"github.com/vietchain/tracepost-larvae/middleware"
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

	// Create a new Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "TracePost-larvaeChain",
		ErrorHandler: api.ErrorHandler,
		ReadTimeout:  time.Duration(cfg.ServerTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.ServerTimeout) * time.Second,
	})

	// Use global middlewares
	app.Use(recover.New())
	app.Use(middleware.LoggerMiddleware())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	// Setup Swagger
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL:         "/swagger/doc.json",
		DeepLinking: true,
	}))

	// Setup API routes
	api.SetupRoutes(app)

	// Print startup message
	startupMessage(cfg)

	// Start the server
	log.Fatal(app.Listen(":" + cfg.ServerPort))
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