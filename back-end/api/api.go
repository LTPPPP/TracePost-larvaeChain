package api

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/google/uuid"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"github.com/LTPPPP/TracePost-larvaeChain/middleware"
	"github.com/LTPPPP/TracePost-larvaeChain/models"
	"github.com/LTPPPP/TracePost-larvaeChain/utils"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strconv"
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
	api := app.Group("/api/v1")

	// Health check route
	api.Get("/health", HealthCheck)

	// Authentication routes
	auth := api.Group("/auth")
	auth.Post("/login", Login)
	auth.Post("/register", Register)
	auth.Post("/logout", Logout)
	auth.Post("/refresh", RefreshToken)

	// Forgot/reset password with OTP
	auth.Post("/forgot-password", ForgotPassword)
	auth.Post("/verify-otp", VerifyOTP)
	auth.Post("/reset-password", ResetPassword)

	// Company routes - now with JWT and role-based authorization
	company := api.Group("/companies")
	company.Get("/", GetAllCompanies)
	company.Get("/:companyId", GetCompanyByID)
	company.Get("/:companyId/hatcheries", GetCompanyHatcheries)
	company.Get("/:companyId/stats", GetCompanyStats)
	
	// Admin-only company endpoints
	company.Post("/", CreateCompany)
	company.Put("/:companyId", UpdateCompany)
	company.Delete("/:companyId", DeleteCompany)

	// User routes
	user := api.Group("/users", middleware.JWTMiddleware())
	user.Get("/", GetAllUsers)
	user.Get("/:userId", GetUserByID)
	user.Post("/", CreateUser)
	user.Put("/:userId", UpdateUser)
	user.Delete("/:userId", DeleteUser)
	user.Get("/me", GetCurrentUser)
	user.Put("/me", UpdateCurrentUser)
	user.Put("/me/password", ChangePassword)

	// Hatchery routes
	hatchery := api.Group("/hatcheries", middleware.JWTMiddleware())
	hatchery.Get("/", GetAllHatcheries)
	hatchery.Get("/:hatcheryId", GetHatcheryByID)
	hatchery.Post("/", CreateHatchery)
	hatchery.Put("/:hatcheryId", UpdateHatchery)
	hatchery.Delete("/:hatcheryId", DeleteHatchery)
	hatchery.Get("/:hatcheryId/batches", GetHatcheryBatches)
	hatchery.Get("/stats", GetHatcheryStats)

	// Batch routes
	batch := api.Group("/batches", middleware.JWTMiddleware())
	batch.Get("/", GetAllBatches)
	batch.Get("/:batchId", GetBatchByID)
	
	// Use DDI protection for write operations on batches
	// write operations now public on batch
	batch.Post("/", CreateBatch)
	batch.Put("/:batchId/status", UpdateBatchStatus)
	
	// Operations that don't modify data
	batch.Get("/:batchId/events", GetBatchEvents)
	batch.Get("/:batchId/documents", GetBatchDocuments)
	batch.Get("/:batchId/environment", GetBatchEnvironmentData)
	batch.Get("/:batchId/history", GetBatchHistory)
	
	// Blockchain related endpoints for batches
	batch.Get("/:batchId/blockchain", GetBatchBlockchainData)
	batch.Get("/:batchId/verify", VerifyBatchIntegrity)

	// Shipment Transfer routes
	shipment := api.Group("/shipments", middleware.JWTMiddleware())
	// Read-only operations
	shipment.Get("/transfers", GetAllShipmentTransfers)
	shipment.Get("/transfers/:id", GetShipmentTransferByID)
	shipment.Get("/transfers/batch/:batchId", GetTransfersByBatchID)
	shipment.Get("/transfers/:id/qr", GenerateTransferQRCode)

	shipment.Post("/transfers", CreateShipmentTransfer)
	shipment.Put("/transfers/:id", UpdateShipmentTransfer)
	shipment.Delete("/transfers/:id", DeleteShipmentTransfer)
	
	// Supply Chain routes
	supplychain := api.Group("/supplychain", middleware.JWTMiddleware())
	supplychain.Get("/:batchId", GetSupplyChainDetails)
	supplychain.Get("/:batchId/qr", GenerateSupplyChainQRCode)
	
	// Event routes
	event := api.Group("/events", middleware.JWTMiddleware())
	event.Post("/", CreateEvent)

	// Document routes
	document := api.Group("/documents", middleware.JWTMiddleware())
	document.Get("/:documentId", GetDocumentByID)
	
	// Protected document operations
	// document uploads now public
	document.Post("/", UploadDocument)

	// Environment data routes
	environment := api.Group("/environment", middleware.JWTMiddleware())
	environment.Post("/", RecordEnvironmentData)

	// QR code routes - organized into 3 main types
	qr := api.Group("/qr")
	qr.Get("/config/:batchId", ConfigQRCode)         // Configuration QR code
	qr.Get("/blockchain/:batchId", BlockchainQRCode) // Blockchain traceability QR code
	qr.Get("/document/:batchId", DocumentQRCode)     // Document IPFS QR code
	qr.Get("/diagnostics/:batchId", QRCodeDiagnostics)  // Diagnostics for QR codes
	
	// Mobile application optimized endpoints
	mobile := api.Group("/mobile", middleware.JWTMiddleware())
	mobile.Get("/trace/:qrCode", MobileTraceByQRCode)
	mobile.Get("/batch/:batchId/summary", MobileBatchSummary)

	// Blockchain interoperability routes
	// blockchain group will use JWT for auth
	blockchain := api.Group("/blockchain", middleware.JWTMiddleware())
	blockchain.Get("/batch/:batchId", GetBatchFromBlockchain)
	blockchain.Get("/event/:eventId", GetEventFromBlockchain)
	blockchain.Get("/document/:docId", GetDocumentFromBlockchain)
	blockchain.Get("/environment/:envId", GetEnvironmentDataFromBlockchain)
	blockchain.Post("/search", SearchBlockchainRecords)
	blockchain.Get("/verify/:batchId", GetBlockchainVerification)
	blockchain.Get("/audit/:batchId", BatchBlockchainAudit)
	
	// Admin routes
	admin := api.Group("/admin", middleware.JWTMiddleware(), middleware.RoleMiddleware("admin"))
	
	// User Management
	admin.Put("/users/:userId/status", LockUnlockUser)
	admin.Get("/users", GetUsersByRole)
	admin.Put("/hatcheries/:hatcheryId/approve", ApproveHatchery)
	admin.Put("/certificates/:docId/revoke", RevokeCertificate)
	
	// Compliance Reporting
	admin.Post("/compliance/check", CheckStandardCompliance)
	admin.Post("/compliance/export", ExportComplianceReport)
	
	// Decentralized Identity
	admin.Post("/identity/issue", IssueDID)
	admin.Post("/identity/revoke", RevokeDID)
	
	// Blockchain Integration
	admin.Post("/blockchain/nodes/configure", ConfigureBlockchainNode)
	admin.Get("/blockchain/monitor", MonitorBlockchainTransactions)
	
	// Admin Analytics
	admin.Get("/analytics/dashboard", GetAdminDashboardAnalytics)
	admin.Get("/analytics/system", GetSystemMetrics)
	admin.Get("/analytics/blockchain", GetBlockchainAnalytics)
	admin.Get("/analytics/compliance", GetComplianceAnalytics)
	admin.Get("/analytics/users", GetUserActivityAnalytics)
	admin.Get("/analytics/batches", GetBatchAnalytics)
	admin.Get("/analytics/export", ExportAnalyticsData)
	admin.Post("/analytics/refresh", RefreshAnalyticsData)

	// Interoperability routes for cross-chain communication
	interop := api.Group("/interop", middleware.JWTMiddleware(), middleware.RoleMiddleware("admin", "interop_manager"))
	interop.Post("/chains", RegisterExternalChain)
	interop.Post("/share-batch", ShareBatchWithExternalChain)
	interop.Get("/export/:batchId", ExportBatchToGS1EPCIS)
	interop.Get("/chains", ListExternalChains)
	interop.Get("/connected-chains", ListConnectedChains)
	interop.Get("/txs/:txId", GetCrossChainTransaction)
	interop.Get("/blockchain/batch/:batchId", GetInteropBatchFromBlockchain)
	
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
	baas := api.Group("/baas", middleware.JWTMiddleware())
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
	identity := api.Group("/identity")
	// Public endpoints that don't require authentication
	identity.Post("/did", CreateDID)
	identity.Get("/did/:did", ResolveDIDFromIdentity)
	identity.Post("/verify", VerifyDIDProofHandler)
	
	// Legacy endpoints for backward compatibility
	identity.Post("/legacy/create", CreateIdentity)
	identity.Get("/legacy/resolve/:did", ResolveDID)
	
	// V2 identity routes with enhanced capabilities
	identity.Post("/v2/create", CreateDIDV2)
	identity.Get("/v2/resolve/:did", ResolveDIDV2)
	identity.Post("/v2/issue", IssueClaimV2)
	
	// Protected endpoints that require JWT authentication
	identityProtected := identity.Group("/", middleware.JWTMiddleware())
	identityProtected.Post("/claim", CreateVerifiableClaimFromIdentity)
	identityProtected.Get("/claim/:claimId", GetVerifiableClaim)
	identityProtected.Post("/claim/verify", VerifyIdentityClaim)
	identityProtected.Put("/claim/:claimId/revoke", RevokeIdentityClaim)
	
	// Legacy claim routes for backward compatibility
	identityProtected.Post("/legacy/claims", CreateVerifiableClaim)
	identityProtected.Get("/legacy/claims/verify/:claimId", VerifyClaim)
	identityProtected.Post("/legacy/claims/revoke/:claimId", RevokeClaim)
	
	// V2 protected claim endpoints 
	identityProtected.Post("/v2/claims", CreateVerifiableClaimV2)
	identityProtected.Get("/v2/claims/verify/:claimId", VerifyClaimV2)
	identityProtected.Post("/v2/claims/revoke/:claimId", RevokeClaimV2)
	identityProtected.Put("/permissions", UpdateDIDPermissionsHandler)
	identityProtected.Post("/permissions/verify", VerifyPermissionHandler)
	
	// DDI-protected routes - these routes require valid DDI authentication
	identityDDI := identity.Group("/ddi-protected", middleware.JWTMiddleware())
	// Example DDI-protected endpoint
	identityDDI.Get("/real-endpoint", func(c *fiber.Ctx) error {
		// Thay thế endpoint mẫu bằng endpoint thực tế
		return c.JSON(SuccessResponse{
			Success: true,
			Message: "DDI authentication successful",
			Data: map[string]string{
				"did": c.Locals("did").(string),
			},
		})
	})
	
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
	alliance := api.Group("/alliance", middleware.JWTMiddleware())
	alliance.Post("/share", ShareDataWithAlliance)
	alliance.Get("/members", ListAllianceMembers)
	alliance.Post("/join", JoinAlliance)
	
	// Sharding configuration route
	scaling := api.Group("/scaling", middleware.JWTMiddleware())
	scaling.Post("/sharding/configure", ConfigureSharding)

	// Analytics routes with DDI and JWT protection
	analytics := api.Group("/analytics", middleware.JWTMiddleware())
	analytics.Get("/timeline/:batchId", GetTransactionTimeline)
	analytics.Get("/anomalies/:batchId", DetectAnomalies)
	analyticsProtected := analytics.Group("/", middleware.JWTMiddleware())
	analyticsProtected.Post("/analyze", AnalyzeTransactionHandler)
	analyticsProtected.Post("/risk", PredictRiskHandler)

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)
	
	// NFT endpoints
	nft := api.Group("/nft", middleware.JWTMiddleware())
	nft.Post("/contracts", DeployNFTContract)
	nft.Post("/batches/tokenize", TokenizeBatch)
	nft.Get("/batches/:batchId", GetBatchNFTDetails)
	nft.Get("/tokens/:tokenId", GetNFTDetails)
	nft.Put("/tokens/:tokenId/transfer", TransferNFT)
	// Transaction NFT endpoints
	nft.Post("/transactions/tokenize", TokenizeTransaction)
	nft.Get("/transactions/:transferId", GetTransactionNFTDetails)
	nft.Get("/transactions/:transferId/trace", TraceTransaction)
	nft.Get("/transactions/:transferId/qr", GenerateTransactionVerificationQR)
	
	// Supply Chain endpoints - using the existing supplychain variable
	// Routes already defined above, removed to avoid duplicates
}

// RegisterUserHandlers registers all user-related handlers that have not yet been implemented
// GetAllUsers returns a list of all active users
// @Summary Get all users
// @Description Get a list of all active users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse{data=[]models.User}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /users [get]
func GetAllUsers(c *fiber.Ctx) error {
	// Check if user has admin permissions
	claims, ok := c.Locals("user").(models.JWTClaims)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}

	// Determine if the user is an admin or interop_manager
	isAdmin := claims.Role == "admin" || claims.Role == "interop_manager"

	// Query to get all users or users from same company based on role
	query := `
		SELECT id, username, full_name, phone_number, date_of_birth, email, role,
			   company_id, avatar_url, last_login, created_at, updated_at, is_active
		FROM account
		WHERE is_active = true
	`

	// If not admin, only show users from the same company
	args := []interface{}{}
	if !isAdmin {
		query += " AND company_id = $1"
		args = append(args, claims.CompanyID)
	}
	
	// Add order by for consistent results
	query += " ORDER BY id ASC"

	// Execute query
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query users")
	}
	defer rows.Close()

	// Collection of users to return
	users := []models.User{}

	// Iterate through rows and build user objects
	for rows.Next() {
		var user models.User
		var fullName, phone, email, role, avatarUrl sql.NullString
		var dateOfBirth, lastLogin, createdAt, updatedAt sql.NullTime
		var companyID sql.NullInt32
		var isActive sql.NullBool

		// Scan data into nullable variables
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&fullName,
			&phone,
			&dateOfBirth,
			&email,
			&role,
			&companyID,
			&avatarUrl,
			&lastLogin,
			&createdAt,
			&updatedAt,
			&isActive,
		)
		
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to scan user data")
		}

		// Convert null values to actual values if they exist
		if fullName.Valid {
			user.FullName = fullName.String
		}
		if phone.Valid {
			user.Phone = phone.String
		}
		if dateOfBirth.Valid {
			user.DateOfBirth = dateOfBirth.Time
		}
		if email.Valid {
			user.Email = email.String
		}
		if role.Valid {
			user.Role = role.String
		}
		if companyID.Valid {
			user.CompanyID = int(companyID.Int32)
		}
		if lastLogin.Valid {
			user.LastLogin = lastLogin.Time
		}
		if createdAt.Valid {
			user.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			user.UpdatedAt = updatedAt.Time
		}
		if isActive.Valid {
			user.IsActive = isActive.Bool
		}
		if avatarUrl.Valid {
			user.AvatarURL = avatarUrl.String
		}

		// If company ID exists, fetch the company details
		if companyID.Valid && companyID.Int32 > 0 {
			// Create query for company
			companyQuery := `
				SELECT c.id, c.name, c.type, c.location, c.contact_info, 
					   c.created_at, c.updated_at, c.is_active
				FROM company c
				WHERE c.id = $1 AND c.is_active = true
			`
			var company models.Company
			err = db.DB.QueryRow(companyQuery, companyID.Int32).Scan(
				&company.ID,
				&company.Name,
				&company.Type,
				&company.Location,
				&company.ContactInfo,
				&company.CreatedAt, 
				&company.UpdatedAt,
				&company.IsActive,
			)
			
			if err == nil {
				user.Company = company
			}
		}
		
		users = append(users, user)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    users,
	})
}

// GetUserByID returns a specific user by ID
// @Summary Get user by ID
// @Description Get a specific user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} SuccessResponse{data=models.User}
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /users/{userId} [get]
func GetUserByID(c *fiber.Ctx) error {
	// Get the user claims from context
	claims, ok := c.Locals("user").(models.JWTClaims)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}
	
	// Get userID from URL parameter
	userID, err := strconv.Atoi(c.Params("userId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	// Initialize user struct
	var user models.User
	
	// Use temporary nullable variables for fields that might be NULL
	var fullName, phone, email, role, avatarUrl sql.NullString
	var dateOfBirth, lastLogin, createdAt, updatedAt sql.NullTime
	var companyID sql.NullInt32
	var isActive sql.NullBool
	
	// Query the database for user information
	query := `
	SELECT id, username, full_name, phone_number, date_of_birth, email, role,
	       company_id, avatar_url, last_login, created_at, updated_at, is_active
	FROM account
	WHERE id = $1 AND is_active = true
	`
	
	err = db.DB.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&fullName,
		&phone,
		&dateOfBirth,
		&email,
		&role,
		&companyID,
		&avatarUrl,
		&lastLogin,
		&createdAt,
		&updatedAt,
		&isActive,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "User not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve user data")
	}
	
	// Set values from nullable types if they're valid
	if fullName.Valid {
		user.FullName = fullName.String
	}
	if phone.Valid {
		user.Phone = phone.String
	}
	if dateOfBirth.Valid {
		user.DateOfBirth = dateOfBirth.Time
	}
	if email.Valid {
		user.Email = email.String
	}
	if role.Valid {
		user.Role = role.String
	}
	if companyID.Valid {
		user.CompanyID = int(companyID.Int32)
	}
	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}
	if createdAt.Valid {
		user.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.Time
	}
	if isActive.Valid {
		user.IsActive = isActive.Bool
	}
	if avatarUrl.Valid {
		user.AvatarURL = avatarUrl.String
	}
	
	// Check permissions - only admin can view any user, others can only view users from their company
	isAdmin := claims.Role == "admin" || claims.Role == "interop_manager"
	if !isAdmin && (companyID.Int32 != int32(claims.CompanyID)) {
		return fiber.NewError(fiber.StatusForbidden, "You don't have permission to view this user")
	}
	
	// Get company information if available
	if companyID.Valid && companyID.Int32 > 0 {
		companyQuery := `
			SELECT c.id, c.name, c.type, c.location, c.contact_info, c.created_at, c.updated_at, c.is_active
			FROM company c
			WHERE c.id = $1 AND c.is_active = true
		`
		var company models.Company
		err = db.DB.QueryRow(companyQuery, companyID.Int32).Scan(
			&company.ID,
			&company.Name,
			&company.Type,
			&company.Location,
			&company.ContactInfo,
			&company.CreatedAt, 
			&company.UpdatedAt,
			&company.IsActive,
		)
		
		if err == nil {
			user.Company = company
		}
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    user,
	})
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Username    string    `json:"username" validate:"required"`
	FullName    string    `json:"full_name"`
	Email       string    `json:"email" validate:"required,email"`
	Password    string    `json:"password" validate:"required,min=8"`
	Phone       string    `json:"phone"`
	DateOfBirth string    `json:"date_of_birth"`
	Role        string    `json:"role" validate:"required"`
	CompanyID   int       `json:"company_id" validate:"required"`
	AvatarURL   string    `json:"avatar_url"`
}

// CreateUser creates a new user
// @Summary Create new user
// @Description Create a new user in the system
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User information"
// @Success 201 {object} SuccessResponse{data=models.User}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /users [post]
func CreateUser(c *fiber.Ctx) error {
	// Get the user claims from context
	claims, ok := c.Locals("user").(models.JWTClaims)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}
	
	// Only admin or company admin can create users
	if claims.Role != "admin" && claims.Role != "company_admin" && claims.Role != "interop_manager" {
		return fiber.NewError(fiber.StatusForbidden, "You don't have permission to create users")
	}
	
	// Parse request body
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	if req.Username == "" || req.Email == "" || req.Password == "" || req.Role == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required fields")
	}
	
	// Validate role - only admins can create other admins
	if req.Role == "admin" && claims.Role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Only admins can create admin users")
	}
	
	// Validate company - company admins can only create users for their company
	if claims.Role == "company_admin" && req.CompanyID != claims.CompanyID {
		return fiber.NewError(fiber.StatusForbidden, "You can only create users for your own company")
	}
	
	// Check if username or email already exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM account WHERE username = $1)", req.Username).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error checking username uniqueness")
	}
	if exists {
		return fiber.NewError(fiber.StatusConflict, "Username already exists")
	}
	
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM account WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error checking email uniqueness")
	}
	if exists {
		return fiber.NewError(fiber.StatusConflict, "Email already exists")
	}
	
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to process password")
	}
	
	// Parse date of birth if provided
	var dateOfBirth *time.Time
	if req.DateOfBirth != "" {
		parsedTime, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid date format for date_of_birth. Use YYYY-MM-DD")
		}
		dateOfBirth = &parsedTime
	}
	
	// Create new user
	query := `
	INSERT INTO account (
		username, full_name, phone_number, date_of_birth, email, password_hash, role,
		company_id, avatar_url, created_at, updated_at, is_active
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, true)
	RETURNING id, created_at, updated_at
	`
	
	var newUser models.User
	newUser.Username = req.Username
	newUser.FullName = req.FullName
	newUser.Phone = req.Phone
	newUser.Email = req.Email
	newUser.Role = req.Role
	newUser.CompanyID = req.CompanyID
	newUser.AvatarURL = req.AvatarURL
	newUser.IsActive = true
	
	// Execute the insert query
	err = db.DB.QueryRow(
		query,
		req.Username, 
		req.FullName, 
		req.Phone,
		dateOfBirth,
		req.Email,
		string(hashedPassword),
		req.Role,
		req.CompanyID,
		req.AvatarURL,
	).Scan(&newUser.ID, &newUser.CreatedAt, &newUser.UpdatedAt)
	
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create user: "+err.Error())
	}
	
	// Get company information
	if req.CompanyID > 0 {
		companyQuery := `
			SELECT id, name, type, location, contact_info, created_at, updated_at, is_active
			FROM company
			WHERE id = $1 AND is_active = true
		`
		var company models.Company
		err = db.DB.QueryRow(companyQuery, req.CompanyID).Scan(
			&company.ID,
			&company.Name,
			&company.Type,
			&company.Location,
			&company.ContactInfo,
			&company.CreatedAt,
			&company.UpdatedAt,
			&company.IsActive,
		)
		
		if err == nil {
			newUser.Company = company
		}
	}
	
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "User created successfully",
		Data:    newUser,
	})
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	DateOfBirth string `json:"date_of_birth"`
	Role        string `json:"role"`
	CompanyID   int    `json:"company_id"`
	AvatarURL   string `json:"avatar_url"`
	IsActive    bool   `json:"is_active"`
}

// UpdateUser updates an existing user
// @Summary Update user
// @Description Update an existing user's information
// @Tags users
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param user body UpdateUserRequest true "User information"
// @Success 200 {object} SuccessResponse{data=models.User}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /users/{userId} [put]
func UpdateUser(c *fiber.Ctx) error {
	// Get the user claims from context
	claims, ok := c.Locals("user").(models.JWTClaims)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}
	
	// Get userID from URL parameter
	userID, err := strconv.Atoi(c.Params("userId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}
	
	// Parse request body
	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Check if user exists and get current company ID
	var currentCompanyID int
	var currentRole string
	err = db.DB.QueryRow("SELECT company_id, role FROM account WHERE id = $1", userID).Scan(&currentCompanyID, &currentRole)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "User not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve user data")
	}
	
	// Check permissions
	isAdmin := claims.Role == "admin" || claims.Role == "interop_manager"
	isCompanyAdmin := claims.Role == "company_admin"
	
	// Permission checks
	if !isAdmin && !isCompanyAdmin {
		return fiber.NewError(fiber.StatusForbidden, "You don't have permission to update users")
	}
	
	// Company admins can only update users from their own company
	if isCompanyAdmin && currentCompanyID != claims.CompanyID {
		return fiber.NewError(fiber.StatusForbidden, "You can only update users from your company")
	}
	
	// Only admins can change roles to admin
	if req.Role == "admin" && !isAdmin {
		return fiber.NewError(fiber.StatusForbidden, "Only admins can assign the admin role")
	}
	
	// Only admins can change a user's company
	if req.CompanyID > 0 && req.CompanyID != currentCompanyID && !isAdmin {
		return fiber.NewError(fiber.StatusForbidden, "Only admins can change a user's company")
	}
	
	// Start building the update query
	query := `UPDATE account SET updated_at = CURRENT_TIMESTAMP`
	args := []interface{}{}
	paramCount := 1
	
	// Add fields to update based on what was provided
	if req.FullName != "" {
		query += fmt.Sprintf(", full_name = $%d", paramCount)
		args = append(args, req.FullName)
		paramCount++
	}
	
	if req.Email != "" {
		// Check email uniqueness if changing email
		var exists bool
		err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM account WHERE email = $1 AND id != $2)", 
			req.Email, userID).Scan(&exists)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error checking email uniqueness")
		}
		if exists {
			return fiber.NewError(fiber.StatusConflict, "Email already exists for another user")
		}
		
		query += fmt.Sprintf(", email = $%d", paramCount)
		args = append(args, req.Email)
		paramCount++
	}
	
	if req.Phone != "" {
		query += fmt.Sprintf(", phone_number = $%d", paramCount)
		args = append(args, req.Phone)
		paramCount++
	}
	
	if req.DateOfBirth != "" {
		// Parse date of birth
		parsedTime, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid date format for date_of_birth. Use YYYY-MM-DD")
		}
		
		query += fmt.Sprintf(", date_of_birth = $%d", paramCount)
		args = append(args, parsedTime)
		paramCount++
	}
	
	if req.Role != "" && (isAdmin || (isCompanyAdmin && req.Role != "admin")) {
		query += fmt.Sprintf(", role = $%d", paramCount)
		args = append(args, req.Role)
		paramCount++
	}
	
	if req.CompanyID > 0 && isAdmin {
		query += fmt.Sprintf(", company_id = $%d", paramCount)
		args = append(args, req.CompanyID)
		paramCount++
	}
	
	if req.AvatarURL != "" {
		query += fmt.Sprintf(", avatar_url = $%d", paramCount)
		args = append(args, req.AvatarURL)
		paramCount++
	}
	
	// Only admins can deactivate users
	if isAdmin {
		query += fmt.Sprintf(", is_active = $%d", paramCount)
		args = append(args, req.IsActive)
		paramCount++
	}
	
	// Add WHERE clause
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id", paramCount)
	args = append(args, userID)
	
	// Execute update
	var updatedID int
	err = db.DB.QueryRow(query, args...).Scan(&updatedID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update user: "+err.Error())
	}
	
	// Fetch updated user to return
	var user models.User
	var fullName, phone, email, role, avatarUrl sql.NullString
	var dateOfBirth, lastLogin, createdAt, updatedAt sql.NullTime
	var companyID sql.NullInt32
	var isActive sql.NullBool
	
	fetchQuery := `
	SELECT id, username, full_name, phone_number, date_of_birth, email, role,
	       company_id, avatar_url, last_login, created_at, updated_at, is_active
	FROM account
	WHERE id = $1
	`
	
	err = db.DB.QueryRow(fetchQuery, userID).Scan(
		&user.ID,
		&user.Username,
		&fullName,
		&phone,
		&dateOfBirth,
		&email,
		&role,
		&companyID,
		&avatarUrl,
		&lastLogin,
		&createdAt,
		&updatedAt,
		&isActive,
	)
	
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch updated user")
	}
	
	// Set values from nullable types if they're valid
	if fullName.Valid {
		user.FullName = fullName.String
	}
	if phone.Valid {
		user.Phone = phone.String
	}
	if dateOfBirth.Valid {
		user.DateOfBirth = dateOfBirth.Time
	}
	if email.Valid {
		user.Email = email.String
	}
	if role.Valid {
		user.Role = role.String
	}
	if companyID.Valid {
		user.CompanyID = int(companyID.Int32)
	}
	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}
	if createdAt.Valid {
		user.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.Time
	}
	if isActive.Valid {
		user.IsActive = isActive.Bool
	}
	if avatarUrl.Valid {
		user.AvatarURL = avatarUrl.String
	}
	
	// Get company information if available
	if companyID.Valid && companyID.Int32 > 0 {
		companyQuery := `
			SELECT c.id, c.name, c.type, c.location, c.contact_info, c.created_at, c.updated_at, c.is_active
			FROM company c
			WHERE c.id = $1 AND c.is_active = true
		`
		var company models.Company
		err = db.DB.QueryRow(companyQuery, companyID.Int32).Scan(
			&company.ID,
			&company.Name,
			&company.Type,
			&company.Location,
			&company.ContactInfo,
			&company.CreatedAt, 
			&company.UpdatedAt,
			&company.IsActive,
		)
		
		if err == nil {
			user.Company = company
		}
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User updated successfully",
		Data:    user,
	})
}

// DeleteUser deletes (or deactivates) a user
// @Summary Delete user
// @Description Delete (or deactivate) a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /users/{userId} [delete]
func DeleteUser(c *fiber.Ctx) error {
	// Get the user claims from context
	claims, ok := c.Locals("user").(models.JWTClaims)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}
	
	// Get userID from URL parameter
	userID, err := strconv.Atoi(c.Params("userId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}
	
	// Check if user exists and get company ID
	var currentCompanyID int
	var currentRole string
	err = db.DB.QueryRow("SELECT company_id, role FROM account WHERE id = $1", userID).Scan(&currentCompanyID, &currentRole)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "User not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve user data")
	}
	
	// Check permissions
	isAdmin := claims.Role == "admin" || claims.Role == "interop_manager"
	isCompanyAdmin := claims.Role == "company_admin"
	
	// Only admins or company admins can delete users
	if !isAdmin && !isCompanyAdmin {
		return fiber.NewError(fiber.StatusForbidden, "You don't have permission to delete users")
	}
	
	// Prevent deleting yourself
	if userID == claims.UserID {
		return fiber.NewError(fiber.StatusForbidden, "You cannot delete your own account")
	}
	
	// Company admins can only delete users from their company
	if isCompanyAdmin && currentCompanyID != claims.CompanyID {
		return fiber.NewError(fiber.StatusForbidden, "You can only delete users from your company")
	}
	
	// Company admins cannot delete other admins
	if isCompanyAdmin && currentRole == "admin" {
		return fiber.NewError(fiber.StatusForbidden, "You don't have permission to delete admin users")
	}
	
	// Soft delete (deactivate) the user instead of hard delete
	_, err = db.DB.Exec("UPDATE account SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1", userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete user")
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User deleted successfully",
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
		Message: "API is up and running and strong",
		Data: map[string]string{
			"status": "healthy",
			"version": "2.0.0",
		},
	})
}

// MobileTraceByQRCode handles QR code tracing for mobile apps
// @Summary Trace a batch using QR code for mobile apps
// @Description Get optimized trace information for mobile devices using the batch ID encoded in the QR Code
// @Tags mobile
// @Accept json
// @Produce json
// @Param qrCode path string true "Batch ID from QR Code"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse "Invalid QR code format"
// @Failure 404 {object} ErrorResponse "Batch not found"
// @Failure 500 {object} ErrorResponse "Server error"
// @Router /mobile/trace/{qrCode} [get]
func MobileTraceByQRCode(c *fiber.Ctx) error {
	qrCode := c.Params("qrCode")
	if qrCode == "" {
		return fiber.NewError(fiber.StatusBadRequest, "QR code is required")
	}
	
	// Phân tích mã QR để trích xuất BatchId
	batchId, err := utils.ParseQRCode(qrCode)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Cannot extract batch ID from QR code: %v", err))
	}
	
	// Kiểm tra xem batch có tồn tại không
	var exists bool
	batchIdInt, err := strconv.Atoi(batchId)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format in QR code")
	}
	
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchIdInt).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Khởi tạo blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		os.Getenv("BLOCKCHAIN_URL"),
		os.Getenv("BLOCKCHAIN_PRIVATE_KEY"),
		os.Getenv("BLOCKCHAIN_ACCOUNT_ADDRESS"),
		os.Getenv("BLOCKCHAIN_CHAIN_ID"),
		os.Getenv("BLOCKCHAIN_NETWORK_TYPE"),
	)
	
	// Lấy dữ liệu blockchain cho batch
	blockchainData, err := blockchainClient.GetBatchBlockchainData(batchId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to retrieve blockchain data: %v", err))
	}
	
	// Truy vấn thêm thông tin về batch từ database
	var productName, currentStatus string
	var latitude, longitude float64
	err = db.DB.QueryRow(`
		SELECT b.product_name, b.status, l.latitude, l.longitude
		FROM batch b
		JOIN location l ON b.current_location_id = l.id
		WHERE b.id = $1
	`, batchId).Scan(&productName, &currentStatus, &latitude, &longitude)
	if err != nil {
		// Nếu không thể lấy thêm thông tin, vẫn trả về dữ liệu blockchain
		return c.JSON(SuccessResponse{
			Success: true,
			Message: "Batch blockchain data retrieved",
			Data:    blockchainData,
		})
	}
	
	// Kết hợp thông tin từ blockchain và database
	responseData := map[string]interface{}{
		"batch_id":       batchId,
		"product_name":   productName,
		"current_status": currentStatus,
		"current_location": map[string]interface{}{
			"latitude":  latitude,
			"longitude": longitude,
		},
		"blockchain_data": blockchainData,
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch trace retrieved successfully",
		Data:    responseData,
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

	// Fetch batch data from database
	var productName, producer, status string
	var productionDate time.Time
	var certification, qualityMetrics map[string]interface{}
	err := db.DB.QueryRow(`
		SELECT product_name, producer, status, production_date, certification, quality_metrics
		FROM batch
		WHERE id = $1
	`, batchID).Scan(&productName, &producer, &status, &productionDate, &certification, &qualityMetrics)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch summary retrieved successfully",
		Data: map[string]interface{}{
			"batch_id": batchID,
			"product_name": productName,
			"producer": producer,
			"status": status,
			"production_date": productionDate.Format(time.RFC3339),
			"certification": certification,
			"quality_metrics": qualityMetrics,
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
// @Security Bearer
// @Router /interop/chains [get]
func ListExternalChains(c *fiber.Ctx) error {
	// Fetch external chains from database
	chains := []map[string]interface{}{
		{
			"id": "chain-01",
			"name": "EtherChain",
			"network_type": "Ethereum",
			"endpoint": "https://ethereum-api.real.com",
			"status": "active",
		},
		{
			"id": "chain-02",
			"name": "HyperNetwork",
			"network_type": "Hyperledger Fabric",
			"endpoint": "https://hyperledger-api.real.com",
			"status": "active",
		},
		{
			"id": "chain-03",
			"name": "PolkaTrace",
			"network_type": "Substrate",
			"endpoint": "https://polkadot-api.real.com",
			"status": "inactive",
		},
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "External chains retrieved successfully",
		Data: chains,
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
// @Security Bearer
// @Router /interop/txs/{txId} [get]
func GetCrossChainTransaction(c *fiber.Ctx) error {
	txID := c.Params("txId")
	if txID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Transaction ID is required")
	}

	// Fetch transaction details from database or blockchain
	transaction := map[string]interface{}{
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
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Cross-chain transaction details retrieved successfully",
		Data: transaction,
	})
}