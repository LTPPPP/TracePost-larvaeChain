package api

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/vietchain/tracepost-larvae/middleware"
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
	// API group with /api/v1 prefix
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", HealthCheck)

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/login", Login)
	auth.Post("/register", Register)

	// User routes (authenticated)
	user := api.Group("/users", middleware.JWTMiddleware())
	user.Get("/me", GetCurrentUser)
	user.Put("/me", UpdateCurrentUser)
	user.Put("/me/password", ChangePassword)

	// Batch routes
	batch := api.Group("/batches")
	batch.Get("/", GetAllBatches)
	batch.Get("/:batchId", GetBatchByID)
	batch.Post("/", middleware.JWTMiddleware(), CreateBatch)
	batch.Put("/:batchId/status", middleware.JWTMiddleware(), UpdateBatchStatus)
	batch.Get("/:batchId/events", GetBatchEvents)
	batch.Get("/:batchId/documents", GetBatchDocuments)
	batch.Get("/:batchId/environment", GetBatchEnvironmentData)
	batch.Get("/:batchId/qr", GenerateBatchQRCode)
	batch.Get("/:batchId/history", GetBatchBlockchainHistory)

	// Event routes
	event := api.Group("/events")
	event.Post("/", middleware.JWTMiddleware(), CreateEvent)

	// Environment routes
	env := api.Group("/environment")
	env.Post("/", middleware.JWTMiddleware(), RecordEnvironmentData)

	// Document routes
	doc := api.Group("/documents")
	doc.Post("/", middleware.JWTMiddleware(), UploadDocument)
	doc.Get("/:documentId", GetDocumentByID)

	// QR code routes
	qr := api.Group("/qr")
	qr.Get("/:code", TraceByQRCode)
	
	// Interoperability routes (new for 2025)
	interop := api.Group("/interop", middleware.JWTMiddleware(), middleware.RoleMiddleware("admin", "manager"))
	interop.Post("/chains", RegisterExternalChain)
	interop.Post("/share-batch", ShareBatchWithExternalChain)
	interop.Get("/export/:batchId", ExportBatchToGS1EPCIS)
	
	// Identity routes (new for 2025)
	identity := api.Group("/identity")
	identity.Post("/create", middleware.JWTMiddleware(), middleware.RoleMiddleware("admin", "manager"), CreateIdentity)
	identity.Get("/resolve/:did", ResolveDID)
	
	// Identity claims routes
	claims := identity.Group("/claims")
	claims.Post("/", middleware.JWTMiddleware(), middleware.RoleMiddleware("admin", "verifier", "authority"), CreateVerifiableClaim)
	claims.Get("/verify/:claimId", VerifyClaim)
	claims.Post("/revoke/:claimId", middleware.JWTMiddleware(), middleware.RoleMiddleware("admin", "verifier", "authority"), RevokeClaim)
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