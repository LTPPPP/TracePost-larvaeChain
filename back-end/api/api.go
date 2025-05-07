package api

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// ErrorHandler handles API errors
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Default status code
	code := fiber.StatusInternalServerError

	// Check if it's a Fiber error
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	// Return JSON error response
	return c.Status(code).JSON(ErrorResponse{
		Success: false,
		Message: "An error occurred while processing your request",
		Error:   err.Error(),
	})
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SetupAPI sets up the API server
func SetupAPI(app *fiber.App) {
	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// API routes
	api := app.Group("/api/v1")  // Changed from "/api" to "/api/v1" to match Swagger documentation

	// Health check route
	api.Get("/health", HealthCheck)

	// Authentication routes
	auth := api.Group("/auth")
	auth.Post("/login", Login)
	auth.Post("/register", Register)
	auth.Post("/logout", Logout)

	// Company routes
	company := api.Group("/companies")
	company.Get("/", GetAllCompanies)
	company.Get("/:companyId", GetCompanyByID)
	company.Post("/", CreateCompany)
	company.Put("/:companyId", UpdateCompany)
	company.Delete("/:companyId", DeleteCompany)
	company.Get("/:companyId/hatcheries", GetCompanyHatcheries)
	company.Get("/:companyId/stats", GetCompanyStats)

	// User routes
	user := api.Group("/users")
	user.Get("/", GetAllUsers)
	user.Get("/:userId", GetUserByID)
	user.Post("/", CreateUser)
	user.Put("/:userId", UpdateUser)
	user.Delete("/:userId", DeleteUser)
	user.Get("/me", GetCurrentUser)
	user.Put("/me", UpdateCurrentUser)
	user.Put("/me/password", ChangePassword)

	// Hatchery routes
	hatchery := api.Group("/hatcheries")
	hatchery.Get("/", GetAllHatcheries)
	hatchery.Get("/:hatcheryId", GetHatcheryByID)
	hatchery.Post("/", CreateHatchery)
	hatchery.Put("/:hatcheryId", UpdateHatchery)
	hatchery.Delete("/:hatcheryId", DeleteHatchery)
	hatchery.Get("/:hatcheryId/batches", GetHatcheryBatches)
	hatchery.Get("/stats", GetHatcheryStats)

	// Batch routes
	batch := api.Group("/batches")
	batch.Get("/", GetAllBatches)
	batch.Get("/:batchId", GetBatchByID)
	batch.Post("/", CreateBatch)
	batch.Put("/:batchId/status", UpdateBatchStatus)
	batch.Get("/:batchId/qr", GenerateBatchQRCode)
	batch.Get("/:batchId/events", GetBatchEvents)
	batch.Get("/:batchId/documents", GetBatchDocuments)
	batch.Get("/:batchId/environment", GetBatchEnvironmentData)
	batch.Get("/:batchId/history", GetBatchHistory)

	// Event routes
	event := api.Group("/events")
	event.Post("/", CreateEvent)

	// Document routes
	document := api.Group("/documents")
	document.Post("/", UploadDocument)
	document.Get("/:documentId", GetDocumentByID)

	// Environment data routes
	environment := api.Group("/environment")
	environment.Post("/", RecordEnvironmentData)

	// QR code routes
	qr := api.Group("/qr")
	qr.Get("/:batchId", TraceByQRCode)
	qr.Get("/gateway/:batchId", GenerateGatewayQRCode)

	// Blockchain interoperability routes
	blockchain := api.Group("/blockchain")
	blockchain.Get("/batch/:batchId", GetBatchFromBlockchain)
	blockchain.Get("/event/:eventId", GetEventFromBlockchain)
	blockchain.Get("/document/:docId", GetDocumentFromBlockchain)
	blockchain.Get("/environment/:envId", GetEnvironmentDataFromBlockchain)

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)
}

// RegisterUserHandlers registers all user-related handlers that have not yet been implemented
func GetAllUsers(c *fiber.Ctx) error {
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Feature not yet implemented",
	})
}

func GetUserByID(c *fiber.Ctx) error {
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Feature not yet implemented",
	})
}

func CreateUser(c *fiber.Ctx) error {
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Feature not yet implemented",
	})
}

func UpdateUser(c *fiber.Ctx) error {
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Feature not yet implemented",
	})
}

func DeleteUser(c *fiber.Ctx) error {
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Feature not yet implemented",
	})
}

// HealthCheck handles the health check endpoint
// @Summary Health check
// @Description Check if the API is running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Router /health [get]
func HealthCheck(c *fiber.Ctx) error {
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "API is up and running",
		Data: map[string]string{
			"status": "healthy",
			"version": "1.0.0",
		},
	})
}