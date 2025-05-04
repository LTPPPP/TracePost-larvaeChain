package api

import (
	"errors"
	// "net/http"

	"github.com/gofiber/fiber/v2"
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

// SetupRoutes sets up all API routes
func SetupRoutes(app *fiber.App) {
	// API versioning
	api := app.Group("/api/v1")

	// Health check route
	api.Get("/health", HealthCheck)

	// Authentication routes
	auth := api.Group("/auth")
	auth.Post("/login", Login)
	auth.Post("/register", Register)

	// Batch routes
	batches := api.Group("/batches")
	batches.Get("/", GetAllBatches)
	batches.Get("/:batchId", GetBatchByID)
	batches.Post("/", CreateBatch)
	batches.Put("/:batchId/status", UpdateBatchStatus)
	batches.Get("/:batchId/events", GetBatchEvents)
	batches.Get("/:batchId/documents", GetBatchDocuments)
	batches.Get("/:batchId/environment", GetBatchEnvironmentData)
	batches.Get("/:batchId/qr", GenerateBatchQRCode)
	batches.Get("/:batchId/history", GetBatchHistory)

	// Event routes
	events := api.Group("/events")
	events.Post("/", CreateEvent)

	// Environment data routes
	environment := api.Group("/environment")
	environment.Post("/", RecordEnvironmentData)

	// Document routes
	documents := api.Group("/documents")
	documents.Post("/", UploadDocument)
	documents.Get("/:documentId", GetDocumentByID)

	// QR code routes
	qr := api.Group("/qr")
	qr.Get("/:code", TraceByQRCode)

	// User routes
	users := api.Group("/users")
	users.Get("/me", GetCurrentUser)
	users.Put("/me", UpdateCurrentUser)
	users.Put("/me/password", ChangePassword)
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