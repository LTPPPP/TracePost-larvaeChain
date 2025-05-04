package api

import (
	"encoding/json"
	// "io/ioutil"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vietchain/tracepost-larvae/blockchain"
	"github.com/vietchain/tracepost-larvae/db"
	"github.com/vietchain/tracepost-larvae/ipfs"
	"github.com/vietchain/tracepost-larvae/models"
)

// CreateEventRequest represents a request to create a new event
type CreateEventRequest struct {
	BatchID   string                 `json:"batch_id"`
	EventType string                 `json:"event_type"`
	Location  string                 `json:"location"`
	ActorID   string                 `json:"actor_id"`
	Details   map[string]interface{} `json:"details"`
}

// RecordEnvironmentDataRequest represents a request to record environment data
type RecordEnvironmentDataRequest struct {
	BatchID          string                 `json:"batch_id"`
	Temperature      float64                `json:"temperature"`
	PH               float64                `json:"ph"`
	Salinity         float64                `json:"salinity"`
	DissolvedOxygen  float64                `json:"dissolved_oxygen"`
	OtherParams      map[string]interface{} `json:"other_params,omitempty"`
}

// UploadDocumentRequest represents a request to upload a document
type UploadDocumentRequest struct {
	BatchID      string `form:"batch_id"`
	DocumentType string `form:"document_type"`
	Issuer       string `form:"issuer"`
}

// TraceByQRCodeResponse represents the response for QR code tracing
type TraceByQRCodeResponse struct {
	Batch          models.Batch            `json:"batch"`
	Events         []models.Event          `json:"events"`
	Documents      []models.Document       `json:"documents"`
	EnvironmentData []models.EnvironmentData `json:"environment_data"`
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
	if req.BatchID == "" || req.EventType == "" || req.Location == "" || req.ActorID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID, event type, location, and actor ID are required")
	}

	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1)", req.BatchID).Scan(&exists)
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
	)

	// Record event on blockchain
	txID, err := blockchainClient.RecordEvent(req.BatchID, req.EventType, req.Location, req.ActorID, req.Details)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record event on blockchain")
	}

	// Generate metadata hash
	metadata := map[string]interface{}{
		"batch_id":    req.BatchID,
		"event_type":  req.EventType,
		"location":    req.Location,
		"actor_id":    req.ActorID,
		"details":     req.Details,
		"recorded_at": time.Now(),
	}
	metadataHash, err := blockchainClient.HashData(metadata)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate metadata hash")
	}

	// Convert details to JSONB
	detailsJSON, err := json.Marshal(req.Details)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to serialize details")
	}
	var detailsJSONB models.JSONB
	err = json.Unmarshal(detailsJSON, &detailsJSONB)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to convert details to JSONB")
	}

	// Insert event into database
	query := `
		INSERT INTO events (batch_id, event_type, timestamp, location, actor_id, details, blockchain_tx_id, metadata_hash)
		VALUES ($1, $2, NOW(), $3, $4, $5, $6, $7)
		RETURNING id, timestamp
	`
	var event models.Event
	event.BatchID = req.BatchID
	event.EventType = req.EventType
	event.Location = req.Location
	event.ActorID = req.ActorID
	event.Details = detailsJSONB
	event.BlockchainTxID = txID
	event.MetadataHash = metadataHash

	err = db.DB.QueryRow(
		query,
		event.BatchID,
		event.EventType,
		event.Location,
		event.ActorID,
		event.Details,
		event.BlockchainTxID,
		event.MetadataHash,
	).Scan(&event.ID, &event.Timestamp)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save event to database")
	}

	// If event type is 'status_change', update batch status
	if event.EventType == "status_change" {
		// Get the new status from the event details
		newStatus, ok := req.Details["new_status"].(string)
		if ok && newStatus != "" {
			// Update batch status in database
			_, err = db.DB.Exec(
				"UPDATE batches SET status = $1 WHERE batch_id = $2",
				newStatus,
				event.BatchID,
			)
			if err != nil {
				// Log error but don't fail the request
				// In a real application, this would be properly logged
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
	if req.BatchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}

	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1)", req.BatchID).Scan(&exists)
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
	)

	// Record environment data on blockchain
	txID, err := blockchainClient.RecordEnvironmentData(
		req.BatchID,
		req.Temperature,
		req.PH,
		req.Salinity,
		req.DissolvedOxygen,
		req.OtherParams,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record environment data on blockchain")
	}

	// Convert other params to JSONB
	var otherParamsJSONB models.JSONB
	if req.OtherParams != nil {
		otherParamsJSON, err := json.Marshal(req.OtherParams)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to serialize other params")
		}
		err = json.Unmarshal(otherParamsJSON, &otherParamsJSONB)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to convert other params to JSONB")
		}
	}

	// Insert environment data into database
	query := `
		INSERT INTO environment_data (batch_id, timestamp, temperature, ph, salinity, dissolved_oxygen, other_params, blockchain_tx_id)
		VALUES ($1, NOW(), $2, $3, $4, $5, $6, $7)
		RETURNING id, timestamp
	`
	var envData models.EnvironmentData
	envData.BatchID = req.BatchID
	envData.Temperature = req.Temperature
	envData.PH = req.PH
	envData.Salinity = req.Salinity
	envData.DissolvedOxygen = req.DissolvedOxygen
	envData.OtherParams = otherParamsJSONB
	envData.BlockchainTxID = txID

	err = db.DB.QueryRow(
		query,
		envData.BatchID,
		envData.Temperature,
		envData.PH,
		envData.Salinity,
		envData.DissolvedOxygen,
		envData.OtherParams,
		envData.BlockchainTxID,
	).Scan(&envData.ID, &envData.Timestamp)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save environment data to database")
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
// @Param batch_id formData string true "Batch ID"
// @Param document_type formData string true "Document type"
// @Param issuer formData string true "Issuer"
// @Param file formData file true "Document file"
// @Success 201 {object} SuccessResponse{data=models.Document}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /documents [post]
func UploadDocument(c *fiber.Ctx) error {
	// Parse form
	form, err := c.MultipartForm()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid multipart form")
	}

	// Get form values
	batchIDs := form.Value["batch_id"]
	docTypes := form.Value["document_type"]
	issuers := form.Value["issuer"]

	// Validate input
	if len(batchIDs) == 0 || len(docTypes) == 0 || len(issuers) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID, document type, and issuer are required")
	}
	batchID := batchIDs[0]
	docType := docTypes[0]
	issuer := issuers[0]

	// Check if batch exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Get file
	files := form.File["file"]
	if len(files) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "File is required")
	}
	file := files[0]

	// Open file
	fileHandle, err := file.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to open file")
	}
	defer fileHandle.Close()

	// Initialize IPFS client
	ipfsClient := ipfs.NewIPFSClient("http://localhost:5001")

	// Upload file to IPFS
	ipfsHash, err := ipfsClient.UploadFile(fileHandle)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to upload file to IPFS")
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
	)

	// Record document on blockchain
	txID, err := blockchainClient.RecordDocument(batchID, docType, ipfsHash, issuer)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record document on blockchain")
	}

	// Insert document into database
	query := `
		INSERT INTO documents (batch_id, document_type, ipfs_hash, upload_date, issuer, is_verified, blockchain_tx_id)
		VALUES ($1, $2, $3, NOW(), $4, $5, $6)
		RETURNING id, upload_date
	`
	var doc models.Document
	doc.BatchID = batchID
	doc.DocumentType = docType
	doc.IPFSHash = ipfsHash
	doc.Issuer = issuer
	doc.IsVerified = false
	doc.BlockchainTxID = txID

	err = db.DB.QueryRow(
		query,
		doc.BatchID,
		doc.DocumentType,
		doc.IPFSHash,
		doc.Issuer,
		doc.IsVerified,
		doc.BlockchainTxID,
	).Scan(&doc.ID, &doc.UploadDate)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save document to database")
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Document uploaded successfully",
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
	documentID := c.Params("documentId")
	if documentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Document ID is required")
	}

	// Query document from database
	var doc models.Document
	query := `
		SELECT id, batch_id, document_type, ipfs_hash, upload_date, issuer, is_verified, blockchain_tx_id
		FROM documents
		WHERE id = $1
	`
	err := db.DB.QueryRow(query, documentID).Scan(
		&doc.ID,
		&doc.BatchID,
		&doc.DocumentType,
		&doc.IPFSHash,
		&doc.UploadDate,
		&doc.Issuer,
		&doc.IsVerified,
		&doc.BlockchainTxID,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return fiber.NewError(fiber.StatusNotFound, "Document not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
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
// @Description Trace a shrimp larvae batch by QR code
// @Tags qr
// @Accept json
// @Produce json
// @Param code path string true "QR Code"
// @Success 200 {object} SuccessResponse{data=TraceByQRCodeResponse}
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /qr/{code} [get]
func TraceByQRCode(c *fiber.Ctx) error {
	// Get code from params
	code := c.Params("code")
	if code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "QR code is required")
	}

	// Query batch from database
	var batch models.Batch
	batchQuery := `
		SELECT id, batch_id, hatchery_id, creation_date, species, quantity, status, blockchain_tx_id, metadata_hash
		FROM batches
		WHERE batch_id = $1
	`
	err := db.DB.QueryRow(batchQuery, code).Scan(
		&batch.ID,
		&batch.BatchID,
		&batch.HatcheryID,
		&batch.CreationDate,
		&batch.Species,
		&batch.Quantity,
		&batch.Status,
		&batch.BlockchainTxID,
		&batch.MetadataHash,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return fiber.NewError(fiber.StatusNotFound, "Batch not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Query events from database
	eventRows, err := db.DB.Query(`
		SELECT id, batch_id, event_type, timestamp, location, actor_id, details, blockchain_tx_id, metadata_hash
		FROM events
		WHERE batch_id = $1
		ORDER BY timestamp DESC
	`, batch.BatchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer eventRows.Close()

	// Parse events
	var events []models.Event
	for eventRows.Next() {
		var event models.Event
		err := eventRows.Scan(
			&event.ID,
			&event.BatchID,
			&event.EventType,
			&event.Timestamp,
			&event.Location,
			&event.ActorID,
			&event.Details,
			&event.BlockchainTxID,
			&event.MetadataHash,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse event data")
		}
		events = append(events, event)
	}

	// Query documents from database
	docRows, err := db.DB.Query(`
		SELECT id, batch_id, document_type, ipfs_hash, upload_date, issuer, is_verified, blockchain_tx_id
		FROM documents
		WHERE batch_id = $1
		ORDER BY upload_date DESC
	`, batch.BatchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer docRows.Close()

	// Parse documents
	var documents []models.Document
	for docRows.Next() {
		var doc models.Document
		err := docRows.Scan(
			&doc.ID,
			&doc.BatchID,
			&doc.DocumentType,
			&doc.IPFSHash,
			&doc.UploadDate,
			&doc.Issuer,
			&doc.IsVerified,
			&doc.BlockchainTxID,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse document data")
		}
		documents = append(documents, doc)
	}

	// Query environment data from database
	envRows, err := db.DB.Query(`
		SELECT id, batch_id, timestamp, temperature, ph, salinity, dissolved_oxygen, other_params, blockchain_tx_id
		FROM environment_data
		WHERE batch_id = $1
		ORDER BY timestamp DESC
	`, batch.BatchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer envRows.Close()

	// Parse environment data
	var envDataList []models.EnvironmentData
	for envRows.Next() {
		var envData models.EnvironmentData
		err := envRows.Scan(
			&envData.ID,
			&envData.BatchID,
			&envData.Timestamp,
			&envData.Temperature,
			&envData.PH,
			&envData.Salinity,
			&envData.DissolvedOxygen,
			&envData.OtherParams,
			&envData.BlockchainTxID,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse environment data")
		}
		envDataList = append(envDataList, envData)
	}

	// Create response
	response := TraceByQRCodeResponse{
		Batch:          batch,
		Events:         events,
		Documents:      documents,
		EnvironmentData: envDataList,
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
	// This is a placeholder - in a real application, you would get the user ID from the JWT token
	// For now, we'll return a mock user
	user := models.User{
		ID:        1,
		Username:  "john.doe",
		Role:      "admin",
		CompanyID: "company-1",
		Email:     "john.doe@example.com",
		CreatedAt: time.Now(),
		LastLogin: time.Now(),
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    user,
	})
}

// UpdateCurrentUser updates the current user
// @Summary Update current user
// @Description Update the current user's information
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/me [put]
func UpdateCurrentUser(c *fiber.Ctx) error {
	// This is a placeholder - in a real application, you would update the user in the database
	// For now, we'll just return a success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User updated successfully",
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