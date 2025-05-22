package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	// "io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/skip2/go-qrcode"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"github.com/LTPPPP/TracePost-larvaeChain/ipfs"
	"github.com/LTPPPP/TracePost-larvaeChain/models"
)

// CreateEventRequest represents a request to create a new event
type CreateEventRequest struct {
	BatchID   int                    `json:"batch_id"`
	EventType string                 `json:"event_type"`
	Location  string                 `json:"location"`
	ActorID   int                    `json:"actor_id"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// RecordEnvironmentDataRequest represents a request to record environment data
type RecordEnvironmentDataRequest struct {
	BatchID     int     `json:"batch_id"`
	Temperature float64 `json:"temperature"`
	PH          float64 `json:"ph"`
	Salinity    float64 `json:"salinity"`
	Density     float64 `json:"density"`
	Age         int     `json:"age"`
}

// UploadDocumentRequest represents a request to upload a document
type UploadDocumentRequest struct {
	BatchID   int    `form:"batch_id"`
	DocType   string `form:"doc_type"`
	UploadedBy int    `form:"uploaded_by"`
}

// UploadAvatarRequest represents a request to upload a profile image
// TraceByQRCodeResponse represents the response for QR code tracing
type TraceByQRCodeResponse struct {
	Batch           models.BatchWithHatchery  `json:"batch"`
	Events          []models.EventWithActor   `json:"events"`
	Documents       []models.Document         `json:"documents"`
	EnvironmentData []models.EnvironmentData  `json:"environment_data"`
	LogisticsChain  []models.LogisticsEvent   `json:"logistics_chain"`
	BlockchainInfo  []models.BlockchainRecord `json:"blockchain_info"`
}

// CreateEvent creates a new event for a batch
// @Summary Create a new event
// @Description Create a new event for a shrimp larvae batch
// @Tags events
// @Accept json
// @Produce json
// @Param request body CreateEventRequest true "Event creation details"
// @Success 201 {object} SuccessResponse{data=models.Event}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /events [post]
func CreateEvent(c *fiber.Ctx) error {
	// Parse request body
	var req CreateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate input
	if req.BatchID <= 0 || req.EventType == "" || req.Location == "" || req.ActorID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID, event type, location, and actor ID are required")
	}

	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", req.BatchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Check if actor exists
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM account WHERE id = $1 AND is_active = true)", req.ActorID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Actor not found")
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(req.Metadata)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to serialize metadata")
	}
	var metadataJSONB models.JSONB
	err = json.Unmarshal(metadataJSON, &metadataJSONB)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to convert metadata to JSONB")
	}

	// Record event on blockchain
	txID, err := blockchainClient.RecordEvent(
		strconv.Itoa(req.BatchID),
		req.EventType,
		req.Location,
		strconv.Itoa(req.ActorID),
		req.Metadata,
	)
	if err != nil {
		// Log error but continue - blockchain is secondary to database
		fmt.Printf("Warning: Failed to record event on blockchain: %v\n", err)
	}

	// Insert event into database
	query := `
		INSERT INTO event (batch_id, event_type, actor_id, location, timestamp, metadata, updated_at, is_active)
		VALUES ($1, $2, $3, $4, NOW(), $5, NOW(), true)
		RETURNING id, timestamp
	`
	var event models.Event
	event.BatchID = req.BatchID
	event.EventType = req.EventType
	event.ActorID = req.ActorID
	event.Location = req.Location
	event.Metadata = metadataJSONB
	event.IsActive = true

	err = db.DB.QueryRow(
		query,
		event.BatchID,
		event.EventType,
		event.ActorID,
		event.Location,
		event.Metadata,
	).Scan(&event.ID, &event.Timestamp)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save event to database")
	}

	// Record blockchain transaction
	if txID != "" {
		// Generate metadata hash
		metadataForHash := map[string]interface{}{
			"event_id":   event.ID,
			"batch_id":   req.BatchID,
			"event_type": req.EventType,
			"location":   req.Location,
			"actor_id":   req.ActorID,
			"metadata":   req.Metadata,
			"timestamp":  event.Timestamp,
		}
		metadataHash, err := blockchainClient.HashData(metadataForHash)
		if err != nil {
			fmt.Printf("Warning: Failed to generate metadata hash: %v\n", err)
		}

		// Save blockchain record
		_, err = db.DB.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "event", event.ID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}

	// If event type is 'status_change', update batch status
	if event.EventType == "status_change" {
		// Get the new status from the event metadata
		newStatus, ok := req.Metadata["new_status"].(string)
		if ok && newStatus != "" {
			// Update batch status in database
			_, err = db.DB.Exec(
				"UPDATE batch SET status = $1, updated_at = NOW() WHERE id = $2",
				newStatus,
				event.BatchID,
			)
			if err != nil {
				// Log error but don't fail the request
				fmt.Printf("Warning: Failed to update batch status: %v\n", err)
			}
		}
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Event created successfully",
		Data:    event,
	})
}

// RecordEnvironmentData records environment data for a batch
// @Summary Record environment data
// @Description Record environment data for a shrimp larvae batch
// @Tags environment
// @Accept json
// @Produce json
// @Param request body RecordEnvironmentDataRequest true "Environment data details"
// @Success 201 {object} SuccessResponse{data=models.EnvironmentData}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /environment [post]
func RecordEnvironmentData(c *fiber.Ctx) error {
	// Parse request body
	var req RecordEnvironmentDataRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate input
	if req.BatchID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}

	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", req.BatchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Record environment data on blockchain
	otherParams := map[string]interface{}{
		"density": req.Density,
		"age":    req.Age,
	}
	txID, err := blockchainClient.RecordEnvironmentData(
		strconv.Itoa(req.BatchID),
		req.Temperature,
		req.PH,
		req.Salinity,
		0,
		otherParams,
	)
	if err != nil {
		// Log error but continue - blockchain is secondary to database
		fmt.Printf("Warning: Failed to record environment data on blockchain: %v\n", err)
	}

	// Insert environment data into database
	query := `
		INSERT INTO environment_data (batch_id, temperature, ph, salinity, density, age, timestamp, updated_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW(), true)
		RETURNING id, timestamp
	`
	var envData models.EnvironmentData
	envData.BatchID = req.BatchID
	envData.Temperature = req.Temperature
	envData.PH = req.PH
	envData.Salinity = req.Salinity
	envData.Density = req.Density
	envData.Age = req.Age
	envData.IsActive = true

	err = db.DB.QueryRow(
		query,
		envData.BatchID,
		envData.Temperature,
		envData.PH,
		envData.Salinity,
		envData.Density,
		envData.Age,
	).Scan(&envData.ID, &envData.Timestamp)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save environment data to database")
	}

	// Record blockchain transaction
	if txID != "" {
		// Generate metadata hash
		metadataForHash := map[string]interface{}{
			"environment_id": envData.ID,
			"batch_id":      req.BatchID,
			"temperature":   req.Temperature,
			"ph":           req.PH,
			"salinity":     req.Salinity,
			"density":      req.Density,
			"age":          req.Age,
			"timestamp":    envData.Timestamp,
		}
		metadataHash, err := blockchainClient.HashData(metadataForHash)
		if err != nil {
			fmt.Printf("Warning: Failed to generate metadata hash: %v\n", err)
		}

		// Save blockchain record
		_, err = db.DB.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "environment_data", envData.ID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Environment data recorded successfully",
		Data:    envData,
	})
}

// UploadDocument uploads a document for a batch
// @Summary Upload a document
// @Description Upload a document for a shrimp larvae batch
// @Tags documents
// @Accept multipart/form-data
// @Produce json
// @Param batch_id formData int true "Batch ID"
// @Param doc_type formData string true "Document type"
// @Param uploaded_by formData int true "Uploader ID"
// @Param file formData file true "Document file"
// @Success 201 {object} SuccessResponse{data=models.Document}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /documents [post]
func UploadDocument(c *fiber.Ctx) error {
	// Parse form with file size limit
	form, err := c.MultipartForm()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid multipart form")
	}

	// Get form values
	batchIDs := form.Value["batch_id"]
	docTypes := form.Value["doc_type"]
	uploaderIDs := form.Value["uploaded_by"]

	// Validate input
	if len(batchIDs) == 0 || len(docTypes) == 0 || len(uploaderIDs) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID, document type, and uploader ID are required")
	}
	
	batchIDStr := batchIDs[0]
	docType := docTypes[0]
	uploaderIDStr := uploaderIDs[0]
	
	// Convert string IDs to integers
	batchID, err := strconv.Atoi(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
	}
	
	uploaderID, err := strconv.Atoi(uploaderIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid uploader ID format")
	}

	// Check if batch exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error checking batch")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found or inactive")
	}

	// Check if uploader exists
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM account WHERE id = $1 AND is_active = true)", uploaderID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error checking uploader")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Uploader not found or inactive")
	}

	// Get file
	files := form.File["file"]
	if len(files) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "File is required")
	}
	file := files[0]

	// Validate file size (e.g., 10MB limit)
	if file.Size > 10*1024*1024 {
		return fiber.NewError(fiber.StatusBadRequest, "File size exceeds 10MB limit")
	}

	// Open file
	fileHandle, err := file.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to open file")
	}
	defer fileHandle.Close()

	// Initialize IPFS+Pinata service with connection pooling
	ipfsPinataService := ipfs.NewIPFSPinataService()

	// Define metadata for Pinata
	metadata := map[string]string{
		"batch_id":     batchIDStr,
		"document_type": docType,
		"uploader_id":   uploaderIDStr,
		"app":           "TracePost-larvaeChain",
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	// Upload file to IPFS and pin to Pinata with retries and timeouts
	ipfsResult, err := ipfsPinataService.UploadFile(fileHandle, file.Filename, metadata, true)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to upload file: %v", err))
	}

	// Initialize blockchain client with configuration from environment
	blockchainClient := blockchain.NewBlockchainClient(
		os.Getenv("BLOCKCHAIN_NODE_URL"),
		os.Getenv("BLOCKCHAIN_CHAIN_ID"),
		os.Getenv("BLOCKCHAIN_ACCOUNT"),
		os.Getenv("BLOCKCHAIN_CONTRACT_ADDRESS"),
		os.Getenv("BLOCKCHAIN_CONSENSUS"),
	)

	// Record document on blockchain
	txID, err := blockchainClient.RecordDocument(strconv.Itoa(batchID), docType, ipfsResult.CID, strconv.Itoa(uploaderID))
	if err != nil {
		// Log error but continue - blockchain is secondary to database
		fmt.Printf("Warning: Failed to record document on blockchain: %v\n", err)
	}

	// Insert document into database
	query := `
		INSERT INTO document (batch_id, doc_type, ipfs_hash, ipfs_uri, file_name, file_size, uploaded_by, uploaded_at, updated_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW(), true)
		RETURNING id, uploaded_at
	`
	var doc models.Document
	doc.BatchID = batchID
	doc.DocType = docType
	doc.IPFSHash = ipfsResult.CID
	
	// Use Pinata URI if available, otherwise use standard IPFS URI
	if ipfsResult.PinataSuccess && ipfsResult.PinataUri != "" {
		doc.IPFSURI = ipfsResult.PinataUri
	} else {
		doc.IPFSURI = ipfsResult.IPFSUri
	}
	
	doc.FileName = ipfsResult.Name
	doc.FileSize = ipfsResult.Size
	doc.UploadedBy = uploaderID
	doc.IsActive = true

	// Debugging: Log the query and parameters before execution
	fmt.Printf("Executing query: %s\n", query)
	fmt.Printf("Parameters: BatchID=%d, DocType=%s, IPFSHash=%s, IPFSURI=%s, FileName=%s, FileSize=%d, UploadedBy=%d\n",
		doc.BatchID, doc.DocType, doc.IPFSHash, doc.IPFSURI, doc.FileName, doc.FileSize, doc.UploadedBy)

	// Execute the query
	err = db.DB.QueryRow(
		query,
		doc.BatchID,
		doc.DocType,
		doc.IPFSHash,
		doc.IPFSURI,
		doc.FileName,
		doc.FileSize,
		doc.UploadedBy,
	).Scan(&doc.ID, &doc.UploadedAt)
	if err != nil {
		// Log the error for debugging
		fmt.Printf("Database error: %v\n", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save document to database")
	}

	// Record blockchain transaction
	if txID != "" {
		// Generate metadata hash
		metadataForHash := map[string]interface{}{
			"document_id": doc.ID,
			"batch_id":    batchID,
			"doc_type":    docType,
			"ipfs_hash":   ipfsResult.CID,
			"ipfs_uri":    doc.IPFSURI,
			"file_name":   ipfsResult.Name,
			"file_size":   ipfsResult.Size,
			"uploaded_by": uploaderID,
			"uploaded_at": doc.UploadedAt,
			"pinata_pinned": ipfsResult.PinataSuccess,
		}
		metadataHash, err := blockchainClient.HashData(metadataForHash)
		if err != nil {
			fmt.Printf("Warning: Failed to generate metadata hash: %v", err)
		}

		// Save blockchain record
		_, err = db.DB.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "document", doc.ID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}

	// Get uploader information before returning response
	var uploader models.Account
	
	// Use temporary nullable variables for fields that might be NULL
	var fullName, phone, email, role sql.NullString
	var dateOfBirth, lastLogin, createdAt, updatedAt sql.NullTime
	var companyID sql.NullInt32
	var isActive sql.NullBool

	uploaderQuery := `
		SELECT u.id, u.username, u.full_name, u.phone_number as phone, u.date_of_birth, u.email, u.role,
		       u.company_id, u.last_login, u.created_at, u.updated_at, u.is_active
		FROM "account" u
		WHERE u.id = $1 AND u.is_active = true
	`
	err = db.DB.QueryRow(uploaderQuery, doc.UploadedBy).Scan(
		&uploader.ID,
		&uploader.Username,
		&fullName,
		&phone,
		&dateOfBirth,
		&email,
		&role,
		&companyID,
		&lastLogin,
		&createdAt,
		&updatedAt,
		&isActive,
	)
	
	// Set values from nullable types if they're valid
	if fullName.Valid {
		uploader.FullName = fullName.String
	}
	if phone.Valid {
		uploader.Phone = phone.String
	}
	if dateOfBirth.Valid {
		uploader.DateOfBirth = dateOfBirth.Time
	}
	if email.Valid {
		uploader.Email = email.String
	}
	if role.Valid {
		uploader.Role = role.String
	}
	if companyID.Valid {
		uploader.CompanyID = int(companyID.Int32)
	}
	if lastLogin.Valid {
		uploader.LastLogin = lastLogin.Time
	}
	if createdAt.Valid {
		uploader.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		uploader.UpdatedAt = updatedAt.Time
	}
	if isActive.Valid {
		uploader.IsActive = isActive.Bool
	}
	if err == nil {
		doc.Uploader = uploader
		
		// Get company information 
		var company models.Company
		companyQuery := `
			SELECT c.id, c.name, c.type, c.location, c.contact_info, c.created_at, c.updated_at, c.is_active
			FROM company c
			WHERE c.id = $1 AND c.is_active = true
		`
		err = db.DB.QueryRow(companyQuery, uploader.CompanyID).Scan(
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
			doc.Uploader.Company = company
			doc.Company = company
		} else {
			fmt.Printf("Warning: Failed to get company data: %v\n", err)
		}
	} else {
		fmt.Printf("Warning: Failed to get uploader data: %v\n", err)
	}

	// Return success response with information about Pinata pinning
	var message string
	if ipfsResult.PinataSuccess {
		message = "Document uploaded successfully and pinned to Pinata"
	} else {
		message = "Document uploaded successfully to IPFS but not pinned to Pinata"
	}
	
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: message,
		Data:    doc,
	})
}

// GetDocumentByID returns a document by ID
// @Summary Get document by ID
// @Description Retrieve a document by its ID
// @Tags documents
// @Accept json
// @Produce json
// @Param documentId path string true "Document ID"
// @Success 200 {object} SuccessResponse{data=models.Document}
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /documents/{documentId} [get]
func GetDocumentByID(c *fiber.Ctx) error {
	// Get document ID from params
	documentIDStr := c.Params("documentId")
	if documentIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Document ID is required")
	}
	
	documentID, err := strconv.Atoi(documentIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid document ID format")
	}

	// Query document from database with all necessary fields
	var doc models.Document
	query := `
		SELECT d.id, d.batch_id, d.doc_type, d.ipfs_hash, d.file_name, d.file_size, 
		       d.uploaded_by, d.uploaded_at, d.updated_at, d.is_active
		FROM document d
		WHERE d.id = $1 AND d.is_active = true
	`
	err = db.DB.QueryRow(query, documentID).Scan(
		&doc.ID,
		&doc.BatchID,
		&doc.DocType,
		&doc.IPFSHash,
		&doc.FileName,
		&doc.FileSize,
		&doc.UploadedBy,
		&doc.UploadedAt,
		&doc.UpdatedAt,
		&doc.IsActive,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return fiber.NewError(fiber.StatusNotFound, "Document not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: " + err.Error())
	}

	// Get IPFS gateway URL from environment or use default
	ipfsGatewayURL := os.Getenv("IPFS_GATEWAY_URL")
	if ipfsGatewayURL == "" {
		ipfsGatewayURL = "https://ipfs.io/ipfs"
	}
	
	// Create IPFS URI
	ipfsClient := ipfs.NewIPFSClient(os.Getenv("IPFS_NODE_URL"))
	doc.IPFSURI = ipfsClient.CreateIPFSURL(doc.IPFSHash, ipfsGatewayURL)
	
	// Get uploader information
	var uploader models.Account
	
	// Use temporary nullable variables for fields that might be NULL
	var fullName, phone, email, role sql.NullString
	var dateOfBirth, lastLogin, createdAt, updatedAt sql.NullTime
	var companyID sql.NullInt32
	var isActive sql.NullBool

	uploaderQuery := `
		SELECT u.id, u.full_name, u.phone_number as phone, u.date_of_birth, u.email, u.role,
		       u.company_id, u.last_login, u.created_at, u.updated_at, u.is_active
		FROM "account" u
		WHERE u.id = $1 AND u.is_active = true
	`
	err = db.DB.QueryRow(uploaderQuery, doc.UploadedBy).Scan(
		&uploader.ID,
		&fullName,
		&phone,
		&dateOfBirth,
		&email,
		&role,
		&companyID,
		&lastLogin,
		&createdAt,
		&updatedAt,
		&isActive,
	)
	
	// Set values from nullable types if they're valid
	if fullName.Valid {
		uploader.FullName = fullName.String
	}
	if phone.Valid {
		uploader.Phone = phone.String
	}
	if dateOfBirth.Valid {
		uploader.DateOfBirth = dateOfBirth.Time
	}
	if email.Valid {
		uploader.Email = email.String
	}
	if role.Valid {
		uploader.Role = role.String
	}
	if companyID.Valid {
		uploader.CompanyID = int(companyID.Int32)
	}
	if lastLogin.Valid {
		uploader.LastLogin = lastLogin.Time
	}
	if createdAt.Valid {
		uploader.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		uploader.UpdatedAt = updatedAt.Time
	}
	if isActive.Valid {
		uploader.IsActive = isActive.Bool
	}
	if err == nil {
		doc.Uploader = uploader
		
		// Get company information 
		var company models.Company
		companyQuery := `
			SELECT c.id, c.name, c.type, c.location, c.contact_info, c.created_at, c.updated_at, c.is_active
			FROM company c
			WHERE c.id = $1 AND c.is_active = true
		`
		err = db.DB.QueryRow(companyQuery, uploader.CompanyID).Scan(
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
			doc.Uploader.Company = company
			doc.Company = company
		}
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Document retrieved successfully",
		Data:    doc,
	})
}

// TraceByQRCode traces a batch by QR code
// @Summary Trace by QR code
// @Description Trace a shrimp larvae batch by QR code, including complete logistics tracking
// @Tags qr
// @Accept json
// @Produce json
// @Param batchID path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=TraceByQRCodeResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /qr/{batchID} [get]
func TraceByQRCode(c *fiber.Ctx) error {
    // Get batchID from params
    batchIDStr := c.Params("batchID")
    if batchIDStr == "" {
        return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
    }
    
    // Convert to integer
    batchID, err := strconv.Atoi(batchIDStr)
    if err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
    }

    // Check if batch exists in database
    var exists bool
    err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchID).Scan(&exists)
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Database error")
    }
    if !exists {
        return fiber.NewError(fiber.StatusNotFound, "Batch not found")
    }

    // Get batch details with hatchery information
    var batchWithHatchery models.BatchWithHatchery
    query := `
        SELECT b.id, b.hatchery_id, b.species, b.quantity, b.status, b.created_at, b.updated_at, b.is_active,
               h.name, h.location, h.contact
        FROM batch b
        JOIN hatchery h ON b.hatchery_id = h.id
        WHERE b.id = $1 AND b.is_active = true
    `
    err = db.DB.QueryRow(query, batchID).Scan(
        &batchWithHatchery.ID,
        &batchWithHatchery.HatcheryID,
        &batchWithHatchery.Species,
        &batchWithHatchery.Quantity,
        &batchWithHatchery.Status,
        &batchWithHatchery.CreatedAt,
        &batchWithHatchery.UpdatedAt,
        &batchWithHatchery.IsActive,
        &batchWithHatchery.HatcheryName,
        &batchWithHatchery.HatcheryLocation,
        &batchWithHatchery.HatcheryContact,
    )
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve batch data")
    }

    // Get events with actor information
    rows, err := db.DB.Query(`
        SELECT e.id, e.batch_id, e.event_type, e.actor_id, e.location, e.timestamp, e.metadata, e.updated_at, e.is_active,
               a.username, a.role, a.email
        FROM event e
        JOIN account a ON e.actor_id = a.id
        WHERE e.batch_id = $1 AND e.is_active = true
        ORDER BY e.timestamp DESC
    `, batchID)
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve events")
    }
    defer rows.Close()

    var eventsWithActor []models.EventWithActor
    for rows.Next() {
        var event models.EventWithActor
        err := rows.Scan(
            &event.ID,
            &event.BatchID,
            &event.EventType,
            &event.ActorID,
            &event.Location,
            &event.Timestamp,
            &event.Metadata,
            &event.UpdatedAt,
            &event.IsActive,
            &event.ActorName,
            &event.ActorRole,
            &event.ActorEmail,
        )
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse event data")
        }
        eventsWithActor = append(eventsWithActor, event)
    }

    // Get documents
    docRows, err := db.DB.Query(`
        SELECT id, batch_id, doc_type, ipfs_hash, uploaded_by, uploaded_at, updated_at, is_active
        FROM document
        WHERE batch_id = $1 AND is_active = true
        ORDER BY uploaded_at DESC
    `, batchID)
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve documents")
    }
    defer docRows.Close()

    var documents []models.Document
    for docRows.Next() {
        var doc models.Document
        err := docRows.Scan(
            &doc.ID,
            &doc.BatchID,
            &doc.DocType,
            &doc.IPFSHash,
            &doc.UploadedBy,
            &doc.UploadedAt,
            &doc.UpdatedAt,
            &doc.IsActive,
        )
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse document data")
        }
        documents = append(documents, doc)
    }

    // Get environment data
    envRows, err := db.DB.Query(`
        SELECT id, batch_id, temperature, pH, salinity, dissolved_oxygen, timestamp, updated_at, is_active
        FROM environment
        WHERE batch_id = $1 AND is_active = true
        ORDER BY timestamp DESC
    `, batchID)
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve environment data")
    }
    defer envRows.Close()

    var envDataList []models.EnvironmentData
    for envRows.Next() {
        var envData models.EnvironmentData
        err := envRows.Scan(
            &envData.ID,
            &envData.BatchID,
            &envData.Temperature,
            &envData.PH,
            &envData.Salinity,
            &envData.Density,
            &envData.Age,
            &envData.Timestamp,
            &envData.UpdatedAt,
            &envData.IsActive,
        )
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse environment data")
        }
        envDataList = append(envDataList, envData)
    }

    // Extract logistics chain from events
    // This builds a chronological chain of transfer/transport events
    var logisticsChain []models.LogisticsEvent
    for _, event := range eventsWithActor {
        // Only include logistics-related events
        if event.EventType == "transfer" || event.EventType == "transport" || 
           event.EventType == "shipping" || event.EventType == "receiving" {
            
            // Extract logistics data from event metadata
            var fromLocation, toLocation, transporterName string
            var departureTime, arrivalTime time.Time
            var status string
            
            // Parse metadata from JSON
            var metadata map[string]interface{}
            if len(event.Metadata) > 0 {
                err := json.Unmarshal(event.Metadata, &metadata)
                if err == nil {
                    // Extract logistics fields from metadata if they exist
                    if val, ok := metadata["from_location"].(string); ok {
                        fromLocation = val
                    } else {
                        fromLocation = event.Location // fallback to event location
                    }
                    
                    if val, ok := metadata["to_location"].(string); ok {
                        toLocation = val
                    }
                    
                    if val, ok := metadata["transporter_name"].(string); ok {
                        transporterName = val
                    } else if event.ActorRole == "transporter" {
                        transporterName = event.ActorName // fallback to actor name if role is transporter
                    }
                    
                    if val, ok := metadata["departure_time"].(string); ok {
                        departureTime, _ = time.Parse(time.RFC3339, val)
                    }
                    
                    if val, ok := metadata["arrival_time"].(string); ok {
                        arrivalTime, _ = time.Parse(time.RFC3339, val)
                    }
                    
                    if val, ok := metadata["status"].(string); ok {
                        status = val
                    } else {
                        status = "completed" // default status
                    }
                }
            }
            
            // Create logistics event
            logisticsEvent := models.LogisticsEvent{
                ID:              event.ID,
                BatchID:         event.BatchID,
                EventType:       event.EventType,
                FromLocation:    fromLocation,
                ToLocation:      toLocation,
                TransporterName: transporterName,
                DepartureTime:   departureTime,
                ArrivalTime:     arrivalTime,
                Status:          status,
                Metadata:        event.Metadata,
                Timestamp:       event.Timestamp,
            }
            
            logisticsChain = append(logisticsChain, logisticsEvent)
        }
    }
    
    // Sort logistics chain by timestamp (oldest first)
    sort.Slice(logisticsChain, func(i, j int) bool {
        return logisticsChain[i].Timestamp.Before(logisticsChain[j].Timestamp)
    })

    // Get blockchain records for this batch
    blockchainRows, err := db.DB.Query(`
        SELECT id, related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active
        FROM blockchain_record
        WHERE (related_table = 'batch' AND related_id = $1) OR 
              EXISTS (SELECT 1 FROM event WHERE id = related_id AND related_table = 'event' AND batch_id = $1) OR
              EXISTS (SELECT 1 FROM document WHERE id = related_id AND related_table = 'document' AND batch_id = $1) OR
              EXISTS (SELECT 1 FROM environment WHERE id = related_id AND related_table = 'environment' AND batch_id = $1)
        ORDER BY created_at DESC
    `, batchID)
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve blockchain records")
    }
    defer blockchainRows.Close()

    var blockchainRecords []models.BlockchainRecord
    for blockchainRows.Next() {
        var record models.BlockchainRecord
        err := blockchainRows.Scan(
            &record.ID,
            &record.RelatedTable,
            &record.RelatedID,
            &record.TxID,
            &record.MetadataHash,
            &record.CreatedAt,
            &record.UpdatedAt,
            &record.IsActive,
        )
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse blockchain record")
        }
        blockchainRecords = append(blockchainRecords, record)
    }

    // Create response with all data
    response := TraceByQRCodeResponse{
        Batch:           batchWithHatchery,
        Events:          eventsWithActor,
        Documents:       documents,
        EnvironmentData: envDataList,
        LogisticsChain:  logisticsChain,
        BlockchainInfo:  blockchainRecords,
    }

    // Return success response
    return c.JSON(SuccessResponse{
        Success: true,
        Message: "Batch traced successfully",
        Data:    response,
    })
}

// GetCurrentUser returns the current user
// @Summary Get current user
// @Description Retrieve the current user's information
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse{data=models.User}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/me [get]
func GetCurrentUser(c *fiber.Ctx) error {
	// Get the user claims from context
	claims, ok := c.Locals("user").(models.JWTClaims)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
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
	       company_id, last_login, created_at, updated_at, is_active, avatar_url
	FROM account
	WHERE id = $1 AND is_active = true
	`
	
	err := db.DB.QueryRow(query, claims.UserID).Scan(
		&user.ID,
		&user.Username,
		&fullName,
		&phone,
		&dateOfBirth,
		&email,
		&role,
		&companyID,
		&lastLogin,
		&createdAt,
		&updatedAt,
		&isActive,
		&avatarUrl,
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
	
	// Don't forget to include company information if companyID is valid
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

	// Calculate profile completion percentage
	completionFields := 0
	totalFields := 5 // Count fields that contribute to completion
	
	if user.FullName != "" {
		completionFields++
	}
	if user.Phone != "" {
		completionFields++
	}
	if !user.DateOfBirth.IsZero() {
		completionFields++
	}
	if user.Email != "" {
		completionFields++
	}
	if user.AvatarURL != "" {
		completionFields++
	}
	
	completionPercentage := int((float64(completionFields) / float64(totalFields)) * 100)
	
	// Return success response with user data and profile completion
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data: map[string]interface{}{
			"user": user,
			"profile_completion": completionPercentage,
		},
	})
}

// UpdateProfileRequest represents the update profile request body
type UpdateProfileRequest struct {
	FullName    string     `json:"full_name,omitempty"`
	Phone       string     `json:"phone,omitempty"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	Email       string     `json:"email,omitempty"`
	Avatar      string     `json:"avatar,omitempty"` // Base64 encoded image
}

// UpdateCurrentUser updates the current user
// @Summary Update current user
// @Description Update the current user's information
// @Tags users
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Profile update details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/me [put]
func UpdateCurrentUser(c *fiber.Ctx) error {
	// Get the user claims from context
	claims, ok := c.Locals("user").(models.JWTClaims)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}

	// Parse request body
	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate email format if provided
	if req.Email != "" {
		// Check if email already exists for another user
		var count int
		err := db.DB.QueryRow("SELECT COUNT(*) FROM account WHERE email = $1 AND id != $2", req.Email, claims.UserID).Scan(&count)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Database error")
		}
		if count > 0 {
			return fiber.NewError(fiber.StatusConflict, "Email already in use by another user")
		}
	}
	
	// Validate phone number format if provided
	if req.Phone != "" {
		// Basic validation - check length and that it contains only digits, +, -, (, )
		validPhone := true
		for _, c := range req.Phone {
			if (c < '0' || c > '9') && c != '+' && c != '-' && c != '(' && c != ')' && c != ' ' {
				validPhone = false
				break
			}
		}
		
		if !validPhone || len(req.Phone) < 8 || len(req.Phone) > 20 {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid phone number format")
		}
	}

	// Construct SQL update query dynamically based on provided fields
	setFields := []string{}
	args := []interface{}{}
	argPos := 1

	if req.FullName != "" {
		setFields = append(setFields, fmt.Sprintf("full_name = $%d", argPos))
		args = append(args, req.FullName)
		argPos++
	}

	if req.Phone != "" {
		setFields = append(setFields, fmt.Sprintf("phone_number = $%d", argPos))
		args = append(args, req.Phone)
		argPos++
	}

	if req.DateOfBirth != nil {
		setFields = append(setFields, fmt.Sprintf("date_of_birth = $%d", argPos))
		args = append(args, req.DateOfBirth)
		argPos++
	}

	if req.Email != "" {
		setFields = append(setFields, fmt.Sprintf("email = $%d", argPos))
		args = append(args, req.Email)
		argPos++
	}

	// Process avatar image upload if provided
	if req.Avatar != "" {
		// Initialize IPFS client with URL from environment variable or use default
		ipfsNodeURL := os.Getenv("IPFS_NODE_URL")
		if ipfsNodeURL == "" {
			ipfsNodeURL = "http://ipfs:5001" // Default IPFS node URL
		}
		ipfsClient := ipfs.NewIPFSClient(ipfsNodeURL)
		
		// Upload the image to IPFS - convert string data to a file-like reader
		reader := bytes.NewReader([]byte(req.Avatar))
		cid, err := ipfsClient.Shell.Add(reader)
		if err != nil {
			fmt.Printf("Error uploading avatar to IPFS: %v\n", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to upload avatar")
		}

		// Generate IPFS URL
		ipfsURL := fmt.Sprintf("ipfs://%s", cid)
		
		setFields = append(setFields, fmt.Sprintf("avatar_url = $%d", argPos))
		args = append(args, ipfsURL)
		argPos++
	}

	// Always update the updated_at timestamp
	setFields = append(setFields, "updated_at = NOW()")

	// If no fields to update
	if len(setFields) == 1 { // Only updated_at
		return fiber.NewError(fiber.StatusBadRequest, "No fields to update")
	}

	// Construct and execute the query
	query := fmt.Sprintf("UPDATE account SET %s WHERE id = $%d", strings.Join(setFields, ", "), argPos)
	args = append(args, claims.UserID)
	
	_, err := db.DB.Exec(query, args...)
	if err != nil {
		fmt.Printf("Error updating user profile: %v\n", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update profile")
	}
	
	// Get updated user data to return in the response
	var user models.User
	
	// Use temporary nullable variables for fields that might be NULL
	var fullName, phone, email, role, avatarUrl sql.NullString
	var dateOfBirth, lastLogin, createdAt, updatedAt sql.NullTime
	var companyID sql.NullInt32
	var isActive sql.NullBool
	
	// Query the database for the updated user information
	queryUser := `
	SELECT id, username, full_name, phone_number, date_of_birth, email, role,
	       company_id, last_login, created_at, updated_at, is_active, avatar_url
	FROM account
	WHERE id = $1 AND is_active = true
	`
	
	err = db.DB.QueryRow(queryUser, claims.UserID).Scan(
		&user.ID,
		&user.Username,
		&fullName,
		&phone,
		&dateOfBirth,
		&email,
		&role,
		&companyID,
		&lastLogin,
		&createdAt,
		&updatedAt,
		&isActive,
		&avatarUrl,
	)
	
	if err != nil {
		// Even if this fails, the profile was updated
		return c.JSON(SuccessResponse{
			Success: true,
			Message: "Profile updated successfully, but unable to retrieve updated data",
		})
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
		// Add avatar URL to user object
		user.AvatarURL = avatarUrl.String
	}
	
	// Don't forget to include company information if companyID is valid
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

	// Calculate profile completion percentage
	completionFields := 0
	totalFields := 5 // Count fields that contribute to completion
	
	if user.FullName != "" {
		completionFields++
	}
	if user.Phone != "" {
		completionFields++
	}
	if !user.DateOfBirth.IsZero() {
		completionFields++
	}
	if user.Email != "" {
		completionFields++
	}
	if user.AvatarURL != "" {
		completionFields++
	}
	
	completionPercentage := int((float64(completionFields) / float64(totalFields)) * 100)

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Profile updated successfully",
		Data: map[string]interface{}{
			"user": user,
			"profile_completion": completionPercentage,
		},
	})
}

// ChangePassword changes the current user's password
// @Summary Change password
// @Description Change the current user's password
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/me/password [put]
func ChangePassword(c *fiber.Ctx) error {
	// This is a placeholder - in a real application, you would update the user's password in the database
	// For now, we'll just return a success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Password changed successfully",
	})
}

// GenerateGatewayQRCode generates a QR code for a batch with a public gateway URL
// @Summary Generate gateway QR code
// @Description Generate a QR code for a batch with a public IPFS gateway URL
// @Tags qr
// @Accept json
// @Produce image/png
// @Param batchId path string true "Batch ID"
// @Param gateway query string false "IPFS gateway to use (default: ipfs.io)"
// @Success 200 {file} byte[] "QR code as PNG image"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /qr/gateway/{batchId} [get]
func GenerateGatewayQRCode(c *fiber.Ctx) error {
	// Get batch ID from params
	batchIDStr := c.Params("batchId")
	if batchIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	batchID, err := strconv.Atoi(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
	}

	// Check if batch exists in database
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Get gateway from query parameter, default to ipfs.io
	gateway := c.Query("gateway", "ipfs.io")
	
	// Create gateway URL format
	qrData := fmt.Sprintf("https://%s/ipfs/%d", gateway, batchID)

	// Generate QR code
	qrCode, err := qrcode.Encode(qrData, qrcode.Medium, 256)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR code")
	}

	// Set necessary headers for image display
	c.Response().Header.Set("Content-Type", "image/png")
	c.Response().Header.Set("Content-Length", fmt.Sprintf("%d", len(qrCode)))
	c.Response().Header.Set("Cache-Control", "public, max-age=86400")
	
	// Send the binary data directly to the client
	return c.Send(qrCode)
}

// Removed UploadAvatar function as the functionality is now integrated into UpdateCurrentUser