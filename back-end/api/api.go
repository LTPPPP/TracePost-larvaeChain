package api

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/google/uuid"
	"github.com/vietchain/tracepost-larvae/middleware"
	"time"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	Error       string `json:"error,omitempty"`
	StatusCode  int    `json:"status_code,omitempty"`
	Path        string `json:"path,omitempty"`
	Method      string `json:"method,omitempty"`
	RequestID   string `json:"request_id,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
	ErrorType   string `json:"error_type,omitempty"`
	ErrorDetail string `json:"error_detail,omitempty"`
}

// ErrorHandler handles API errors
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Default status code
	code := fiber.StatusInternalServerError
	errorType := "InternalServerError"
	errorDetail := "An unexpected error occurred on the server"

	// Check if it's a Fiber error
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
		
		// Determine error type based on status code
		switch code {
		case fiber.StatusBadRequest:
			errorType = "BadRequest"
			errorDetail = "The request was invalid or cannot be served"
		case fiber.StatusUnauthorized:
			errorType = "Unauthorized"
			errorDetail = "Authentication is required and has failed or has not been provided"
		case fiber.StatusForbidden:
			errorType = "Forbidden"
			errorDetail = "The request was valid, but you don't have permission to access the requested resource"
		case fiber.StatusNotFound:
			errorType = "NotFound"
			errorDetail = "The requested resource could not be found"
		case fiber.StatusMethodNotAllowed:
			errorType = "MethodNotAllowed"
			errorDetail = "The method specified in the request is not allowed for the resource"
		case fiber.StatusConflict:
			errorType = "Conflict"
			errorDetail = "The request could not be completed due to a conflict with the current state of the resource"
		case fiber.StatusUnprocessableEntity:
			errorType = "UnprocessableEntity"
			errorDetail = "The request was well-formed but was unable to be processed due to semantic errors"
		case fiber.StatusTooManyRequests:
			errorType = "TooManyRequests"
			errorDetail = "You have sent too many requests in a given amount of time"
		}
	}

	// Get detailed message from error
	errorMessage := err.Error()
	
	// Create a request ID if not present
	requestID := c.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}

	// Return enhanced JSON error response
	return c.Status(code).JSON(ErrorResponse{
		Success:     false,
		Message:     "An error occurred while processing your request",
		Error:       errorMessage,
		StatusCode:  code,
		Path:        c.Path(),
		Method:      c.Method(),
		RequestID:   requestID,
		Timestamp:   time.Now().Format(time.RFC3339),
		ErrorType:   errorType,
		ErrorDetail: errorDetail,
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

	// Company routes - now with JWT and role-based authorization
	company := api.Group("/companies", middleware.JWTMiddleware())
	company.Get("/", GetAllCompanies)
	company.Get("/:companyId", GetCompanyByID)
	company.Get("/:companyId/hatcheries", GetCompanyHatcheries)
	company.Get("/:companyId/stats", GetCompanyStats)
	
	// Admin-only company endpoints
	company.Post("/", middleware.RoleMiddleware("admin"), CreateCompany)
	company.Put("/:companyId", middleware.RoleMiddleware("admin"), UpdateCompany)
	company.Delete("/:companyId", middleware.RoleMiddleware("admin"), DeleteCompany)

	// User routes
	user := api.Group("/users", middleware.JWTMiddleware())
	user.Get("/", middleware.RoleMiddleware("admin"), GetAllUsers)
	user.Get("/:userId", middleware.RoleMiddleware("admin"), GetUserByID)
	user.Post("/", middleware.RoleMiddleware("admin"), CreateUser)
	user.Put("/:userId", middleware.RoleMiddleware("admin"), UpdateUser)
	user.Delete("/:userId", middleware.RoleMiddleware("admin"), DeleteUser)
	user.Get("/me", GetCurrentUser)
	user.Put("/me", UpdateCurrentUser)
	user.Put("/me/password", ChangePassword)

	// Hatchery routes
	hatchery := api.Group("/hatcheries", middleware.JWTMiddleware())
	hatchery.Get("/", GetAllHatcheries)
	hatchery.Get("/:hatcheryId", GetHatcheryByID)
	hatchery.Post("/", middleware.RoleMiddleware("admin", "hatchery_manager"), CreateHatchery)
	hatchery.Put("/:hatcheryId", middleware.RoleMiddleware("admin", "hatchery_manager"), UpdateHatchery)
	hatchery.Delete("/:hatcheryId", middleware.RoleMiddleware("admin"), DeleteHatchery)
	hatchery.Get("/:hatcheryId/batches", GetHatcheryBatches)
	hatchery.Get("/stats", GetHatcheryStats)

	// Batch routes
	batch := api.Group("/batches", middleware.JWTMiddleware())
	batch.Get("/", GetAllBatches)
	batch.Get("/:batchId", GetBatchByID)
	batch.Post("/", middleware.RoleMiddleware("admin", "hatchery_manager", "farm_manager"), CreateBatch)
	batch.Put("/:batchId/status", middleware.RoleMiddleware("admin", "hatchery_manager", "farm_manager"), UpdateBatchStatus)
	batch.Get("/:batchId/qr", GenerateBatchQRCode)
	batch.Get("/:batchId/events", GetBatchEvents)
	batch.Get("/:batchId/documents", GetBatchDocuments)
	batch.Get("/:batchId/environment", GetBatchEnvironmentData)
	batch.Get("/:batchId/history", GetBatchHistory)

	// Event routes
	event := api.Group("/events", middleware.JWTMiddleware())
	event.Post("/", CreateEvent)

	// Document routes
	document := api.Group("/documents", middleware.JWTMiddleware())
	document.Post("/", UploadDocument)
	document.Get("/:documentId", GetDocumentByID)

	// Environment data routes
	environment := api.Group("/environment", middleware.JWTMiddleware())
	environment.Post("/", RecordEnvironmentData)

	// QR code routes - public access
	qr := api.Group("/qr")
	qr.Get("/:batchId", TraceByQRCode)
	qr.Get("/gateway/:batchId", GenerateGatewayQRCode)

	// Blockchain interoperability routes
	blockchain := api.Group("/blockchain", middleware.JWTMiddleware())
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