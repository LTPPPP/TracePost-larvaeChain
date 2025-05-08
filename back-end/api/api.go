package api

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/google/uuid"
	"github.com/LTPPPP/TracePost-larvaeChain/middleware"
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
	
	// Mobile application optimized endpoints
	mobile := api.Group("/mobile")
	mobile.Get("/trace/:qrCode", MobileTraceByQRCode)
	mobile.Get("/batch/:batchId/summary", MobileBatchSummary)

	// Blockchain interoperability routes
	blockchain := api.Group("/blockchain", middleware.JWTMiddleware())
	blockchain.Get("/batch/:batchId", GetBatchFromBlockchain)
	blockchain.Get("/event/:eventId", GetEventFromBlockchain)
	blockchain.Get("/document/:docId", GetDocumentFromBlockchain)
	blockchain.Get("/environment/:envId", GetEnvironmentDataFromBlockchain)
	
	// Interoperability routes for cross-chain communication
	interop := api.Group("/interop", middleware.JWTMiddleware(), middleware.RoleMiddleware("admin", "interop_manager"))
	interop.Post("/chains", RegisterExternalChain)
	interop.Post("/share-batch", ShareBatchWithExternalChain)
	interop.Get("/export/:batchId", ExportBatchToGS1EPCIS)
	interop.Get("/chains", ListExternalChains)
	interop.Get("/txs/:txId", GetCrossChainTransaction)
	
	// Cosmos SDK Integration routes
	interop.Post("/bridges/cosmos", CreateCosmosBridge)
	interop.Post("/bridges/cosmos/channels", AddIBCChannel)
	interop.Post("/ibc/send", SendIBCPacket)
	interop.Get("/protocols", GetSupportedProtocols)
	interop.Get("/status/:protocol/:sourceChainId/:txId", GetTransactionStatus)
	interop.Post("/verify", VerifyTransaction)
	
	// Polkadot integration routes
	interop.Post("/bridges/polkadot", CreatePolkadotBridge)
	interop.Post("/xcm/send", SendXCMMessage)
	
	// Blockchain-as-a-Service (BaaS) routes
	baas := api.Group("/baas", middleware.JWTMiddleware(), middleware.RoleMiddleware("admin", "baas_manager"))
	baas.Post("/networks", CreateBlockchainNetwork)
	baas.Get("/networks", ListBlockchainNetworks)
	baas.Get("/networks/:networkId", GetBlockchainNetwork)
	baas.Put("/networks/:networkId", UpdateBlockchainNetwork)
	baas.Delete("/networks/:networkId", DeleteBlockchainNetwork)
	baas.Post("/networks/:networkId/nodes", AddNodeToNetwork)
	baas.Get("/templates", ListBlockchainTemplates)
	baas.Post("/deployments", DeployBlockchainContract)
	baas.Get("/deployments", ListContractDeployments)
	baas.Get("/deployments/:deploymentId", GetContractDeployment)
	
	// Decentralized Digital Identity (DDI) routes
	identity := api.Group("/identity", middleware.JWTMiddleware())
	identity.Post("/did", CreateDID)
	identity.Get("/did/:did", ResolveDIDFromIdentity)
	identity.Post("/claim", CreateVerifiableClaimFromIdentity)
	identity.Get("/claim/:claimId", GetVerifiableClaim)
	identity.Post("/claim/verify", VerifyIdentityClaim)
	identity.Put("/claim/:claimId/revoke", RevokeIdentityClaim)
	
	// Compliance and regulation routes
	compliance := api.Group("/compliance", middleware.JWTMiddleware())
	compliance.Get("/check/:batchId", CheckBatchCompliance)
	compliance.Get("/report/:batchId", GenerateComplianceReport)
	compliance.Get("/standards", ListComplianceStandards)
	compliance.Post("/validate", ValidateAgainstStandard)
	
	// Geospatial tracking routes
	geo := api.Group("/geo", middleware.JWTMiddleware())
	geo.Post("/location", RecordGeoLocation)
	geo.Get("/batch/:batchId/journey", GetBatchJourney)
	geo.Get("/batch/:batchId/current-location", GetBatchCurrentLocation)
	
	// Industry alliance routes
	alliance := api.Group("/alliance", middleware.JWTMiddleware(), middleware.RoleMiddleware("admin", "alliance_manager"))
	alliance.Post("/share", ShareDataWithAlliance)
	alliance.Get("/members", ListAllianceMembers)
	alliance.Post("/join", JoinAlliance)
	
	// Layer 2 scaling and sustainability routes
	scaling := api.Group("/scaling", middleware.JWTMiddleware(), middleware.RoleMiddleware("admin"))
	scaling.Post("/l2/enable", EnableLayer2Scaling)
	scaling.Get("/l2/status", GetLayer2Status)
	scaling.Post("/sharding/configure", ConfigureSharding)

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Identity API routes
	app.Post("/api/v1/identity/did", CreateDID)
	app.Get("/api/v1/identity/did/:did", ResolveDIDFromIdentity)
	app.Post("/api/v1/identity/claim", CreateVerifiableClaimFromIdentity)
	app.Get("/api/v1/identity/claim/:claimId", GetVerifiableClaim)
	app.Post("/api/v1/identity/claim/verify", VerifyIdentityClaim)
	app.Put("/api/v1/identity/claim/:claimId/revoke", RevokeIdentityClaim)
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

// MobileTraceByQRCode handles QR code tracing for mobile apps
// @Summary Trace a batch using QR code for mobile apps
// @Description Get optimized trace information for mobile devices
// @Tags mobile
// @Accept json
// @Produce json
// @Param qrCode path string true "QR Code"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /mobile/trace/{qrCode} [get]
func MobileTraceByQRCode(c *fiber.Ctx) error {
	qrCode := c.Params("qrCode")
	if qrCode == "" {
		return fiber.NewError(fiber.StatusBadRequest, "QR code is required")
	}

	// This is a placeholder implementation
	// In a real implementation, you would decode the QR code and fetch the relevant data
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch trace retrieved successfully",
		Data: map[string]interface{}{
			"batch_id": "sample-batch-" + qrCode,
			"product_name": "Sample Product",
			"current_status": "Processing",
			"current_location": map[string]interface{}{
				"name": "Processing Plant",
				"latitude": 10.78,
				"longitude": 106.69,
			},
			"journey_summary": []map[string]interface{}{
				{
					"event": "Created",
					"location": "Hatchery ABC",
					"timestamp": time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339),
				},
				{
					"event": "Shipped",
					"location": "Farm XYZ",
					"timestamp": time.Now().Add(-15 * 24 * time.Hour).Format(time.RFC3339),
				},
				{
					"event": "Processing",
					"location": "Processing Plant",
					"timestamp": time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339),
				},
			},
		},
	})
}

// MobileBatchSummary provides a mobile-optimized summary of a batch
// @Summary Get batch summary for mobile apps
// @Description Get a mobile-optimized summary of a batch
// @Tags mobile
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /mobile/batch/{batchId}/summary [get]
func MobileBatchSummary(c *fiber.Ctx) error {
	batchID := c.Params("batchId")
	if batchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}

	// This is a placeholder implementation
	// In a real implementation, you would fetch the batch data from the database
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch summary retrieved successfully",
		Data: map[string]interface{}{
			"batch_id": batchID,
			"product_name": "Sample Product",
			"producer": "Sample Producer",
			"status": "Processing",
			"production_date": time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339),
			"certification": map[string]interface{}{
				"organic": true,
				"antibiotic_free": true,
				"sustainable": true,
			},
			"quality_metrics": map[string]interface{}{
				"health_index": 92,
				"growth_rate": "Above average",
				"sustainability_score": 87,
			},
		},
	})
}

// ListExternalChains lists external blockchain networks available for interoperability
// @Summary List external blockchain networks
// @Description Get a list of registered external blockchain networks for interoperability
// @Tags interoperability
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /interop/chains [get]
func ListExternalChains(c *fiber.Ctx) error {
	// This is a placeholder implementation
	// In a real implementation, you would fetch the external chains from the database
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "External chains retrieved successfully",
		Data: []map[string]interface{}{
			{
				"id": "chain-01",
				"name": "EtherChain",
				"network_type": "Ethereum",
				"endpoint": "https://ethereum-api.example.com",
				"status": "active",
			},
			{
				"id": "chain-02",
				"name": "HyperNetwork",
				"network_type": "Hyperledger Fabric",
				"endpoint": "https://hyperledger-api.example.com",
				"status": "active",
			},
			{
				"id": "chain-03",
				"name": "PolkaTrace",
				"network_type": "Substrate",
				"endpoint": "https://polkadot-api.example.com",
				"status": "inactive",
			},
		},
	})
}

// GetCrossChainTransaction gets details of a cross-chain transaction
// @Summary Get cross-chain transaction details
// @Description Get details of a transaction that spans multiple blockchain networks
// @Tags interoperability
// @Accept json
// @Produce json
// @Param txId path string true "Transaction ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /interop/txs/{txId} [get]
func GetCrossChainTransaction(c *fiber.Ctx) error {
	txID := c.Params("txId")
	if txID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Transaction ID is required")
	}

	// This is a placeholder implementation
	// In a real implementation, you would fetch the transaction details from the database or blockchain
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Cross-chain transaction details retrieved successfully",
		Data: map[string]interface{}{
			"tx_id": txID,
			"status": "completed",
			"created_at": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"completed_at": time.Now().Add(-23 * time.Hour).Format(time.RFC3339),
			"source_chain": map[string]interface{}{
				"id": "chain-01",
				"name": "EtherChain",
				"tx_hash": "0x" + txID + "a1b2c3d4e5f6",
				"block_number": 12345678,
			},
			"destination_chain": map[string]interface{}{
				"id": "chain-02",
				"name": "HyperNetwork",
				"tx_hash": "hyper-" + txID + "-9z8y7x",
				"block_id": "block98765",
			},
			"asset": map[string]interface{}{
				"type": "batch_data",
				"id": "batch-123456",
				"name": "Organic Shrimp Batch #123456",
			},
		},
	})
}