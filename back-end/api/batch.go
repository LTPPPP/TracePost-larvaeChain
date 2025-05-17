package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/skip2/go-qrcode"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"github.com/LTPPPP/TracePost-larvaeChain/models"
)

// CreateBatchRequest represents a request to create a new batch
type CreateBatchRequest struct {
	HatcheryID int    `json:"hatchery_id"`
	Species    string `json:"species"`
	Quantity   int    `json:"quantity"`
}

// UpdateBatchStatusRequest represents a request to update a batch status
type UpdateBatchStatusRequest struct {
	Status string `json:"status"`
}

// GetAllBatches returns all batches
// @Summary Get all batches
// @Description Retrieve all shrimp larvae batches
// @Tags batches
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse{data=[]models.Batch}
// @Failure 500 {object} ErrorResponse
// @Router /batches [get]
func GetAllBatches(c *fiber.Ctx) error {
	// Query batches from database with hatchery and company information
	rows, err := db.DB.Query(`
		SELECT 
			b.id, b.hatchery_id, b.species, b.quantity, b.status, b.created_at, b.updated_at, b.is_active,
			h.id, h.name, h.location, h.contact, h.company_id, h.created_at, h.updated_at, h.is_active,
			c.id, c.name, c.type, c.location, c.contact_info, c.created_at, c.updated_at, c.is_active
		FROM batch b
		INNER JOIN hatchery h ON b.hatchery_id = h.id AND h.is_active = true
		INNER JOIN company c ON h.company_id = c.id AND c.is_active = true 
		WHERE b.is_active = true
		ORDER BY b.created_at DESC
	`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse batches
	var batches []models.Batch
	for rows.Next() {
		var batch models.Batch
		var hatchery models.Hatchery
		var company models.Company
		err := rows.Scan(
			&batch.ID,
			&batch.HatcheryID,
			&batch.Species,
			&batch.Quantity,
			&batch.Status,
			&batch.CreatedAt,
			&batch.UpdatedAt,
			&batch.IsActive,
			&hatchery.ID,
			&hatchery.Name,
			&hatchery.Location,
			&hatchery.Contact,
			&hatchery.CompanyID,
			&hatchery.CreatedAt,
			&hatchery.UpdatedAt,
			&hatchery.IsActive,
			&company.ID,
			&company.Name,
			&company.Type,
			&company.Location,
			&company.ContactInfo,
			&company.CreatedAt,
			&company.UpdatedAt,
			&company.IsActive,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse batch data")
		}

		// Set relationships
		hatchery.Company = company
		batch.Hatchery = hatchery
		batches = append(batches, batch)
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batches retrieved successfully",
		Data:    batches,
	})
}

// GetBatchByID returns a batch by ID
// @Summary Get batch by ID
// @Description Retrieve a shrimp larvae batch by its ID
// @Tags batches
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=models.Batch}
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /batches/{batchId} [get]
func GetBatchByID(c *fiber.Ctx) error {
	// Get batch ID from params
	batchIDStr := c.Params("batchId")
	if batchIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	batchID, err := strconv.Atoi(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
	}

	// Query batch from database with hatchery and company information
	var batch models.Batch
	var hatchery models.Hatchery
	var company models.Company
	query := `
		SELECT 
			b.id, b.hatchery_id, b.species, b.quantity, b.status, b.created_at, b.updated_at, b.is_active,
			h.id, h.name, h.location, h.contact, h.company_id, h.created_at, h.updated_at, h.is_active,
			c.id, c.name, c.type, c.location, c.contact_info, c.created_at, c.updated_at, c.is_active
		FROM batch b
		INNER JOIN hatchery h ON b.hatchery_id = h.id AND h.is_active = true
		INNER JOIN company c ON h.company_id = c.id AND c.is_active = true
		WHERE b.id = $1 AND b.is_active = true
	`
	err = db.DB.QueryRow(query, batchID).Scan(
		&batch.ID,
		&batch.HatcheryID,
		&batch.Species,
		&batch.Quantity,
		&batch.Status,
		&batch.CreatedAt,
		&batch.UpdatedAt,
		&batch.IsActive,
		&hatchery.ID,
		&hatchery.Name,
		&hatchery.Location,
		&hatchery.Contact,
		&hatchery.CompanyID,
		&hatchery.CreatedAt,
		&hatchery.UpdatedAt,
		&hatchery.IsActive,
		&company.ID,
		&company.Name,
		&company.Type,
		&company.Location,
		&company.ContactInfo,
		&company.CreatedAt,
		&company.UpdatedAt,
		&company.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "Batch not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	
	// Set relationships
	hatchery.Company = company
	batch.Hatchery = hatchery

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch retrieved successfully",
		Data:    batch,
	})
}

// CreateBatch creates a new batch
// @Summary Create a new batch
// @Description Create a new shrimp larvae batch
// @Tags batches
// @Accept json
// @Produce json
// @Param request body CreateBatchRequest true "Batch creation details"
// @Success 201 {object} SuccessResponse{data=models.Batch}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /batches [post]
func CreateBatch(c *fiber.Ctx) error {
	// Parse request body
	var req CreateBatchRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate input
	if req.HatcheryID <= 0 || req.Species == "" || req.Quantity <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Hatchery ID, species, and quantity are required")
	}

	// Check if hatchery exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM hatchery WHERE id = $1 AND is_active = true)", req.HatcheryID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusBadRequest, "Hatchery not found")
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Get hatchery information first with company details
	var hatchery models.Hatchery
	hatcheryQuery := `
		SELECT h.id, h.name, h.location, h.contact, h.company_id, h.created_at, h.updated_at, h.is_active,
			   c.id, c.name, c.type, c.location, c.contact_info, c.created_at, c.updated_at, c.is_active
		FROM hatchery h
		INNER JOIN company c ON h.company_id = c.id AND c.is_active = true
		WHERE h.id = $1 AND h.is_active = true
	`
	var company models.Company
	err = db.DB.QueryRow(hatcheryQuery, req.HatcheryID).Scan(
		&hatchery.ID,
		&hatchery.Name,
		&hatchery.Location,
		&hatchery.Contact,
		&hatchery.CompanyID,
		&hatchery.CreatedAt,
		&hatchery.UpdatedAt,
		&hatchery.IsActive,
		&company.ID,
		&company.Name,
		&company.Type,
		&company.Location,
		&company.ContactInfo,
		&company.CreatedAt,
		&company.UpdatedAt,
		&company.IsActive,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get hatchery information")
	}
	hatchery.Company = company

	// Insert batch into database
	query := `
		INSERT INTO batch (hatchery_id, species, quantity, status, created_at, updated_at, is_active)
		VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		RETURNING id, created_at, updated_at
	`
	var batch models.Batch
	batch.HatcheryID = req.HatcheryID
	batch.Species = req.Species
	batch.Quantity = req.Quantity
	batch.Status = "created"
	batch.IsActive = true
	batch.Hatchery = hatchery

	err = db.DB.QueryRow(
		query,
		batch.HatcheryID,
		batch.Species,
		batch.Quantity,
		batch.Status,
	).Scan(&batch.ID, &batch.CreatedAt, &batch.UpdatedAt)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save batch to database")
	}

	// Create batch on blockchain
	txID, err := blockchainClient.CreateBatch(
		strconv.Itoa(batch.ID),
		strconv.Itoa(req.HatcheryID),
		req.Species,
		req.Quantity,
	)
	if err != nil {
		// Log the error but continue - blockchain is secondary to database
		// In a production environment, we might want to handle this differently
		fmt.Printf("Warning: Failed to record batch on blockchain: %v\n", err)
	}

	// Record blockchain transaction
	if txID != "" {
		// Generate metadata hash
		metadata := map[string]interface{}{
			"batch_id":     batch.ID,
			"hatchery_id":  req.HatcheryID,
			"species":      req.Species,
			"quantity":     req.Quantity,
			"created_at":   batch.CreatedAt,
		}
		metadataHash, err := blockchainClient.HashData(metadata)
		if err != nil {
			fmt.Printf("Warning: Failed to generate metadata hash: %v\n", err)
		}

		// Save blockchain record
		_, err = db.DB.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "batch", batch.ID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Batch created successfully",
		Data:    batch,
	})
}

// UpdateBatchStatus updates the status of a batch
// @Summary Update batch status
// @Description Update the status of a shrimp larvae batch
// @Tags batches
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Param request body UpdateBatchStatusRequest true "Status update details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /batches/{batchId}/status [put]
func UpdateBatchStatus(c *fiber.Ctx) error {
	// Get batch ID from params
	batchIDStr := c.Params("batchId")
	if batchIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	batchID, err := strconv.Atoi(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
	}

	// Parse request body
	var req UpdateBatchStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate input
	if req.Status == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Status is required")
	}

	// Check if batch exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchID).Scan(&exists)
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

	// Update batch status in database
	_, err = db.DB.Exec(
		"UPDATE batch SET status = $1, updated_at = NOW() WHERE id = $2",
		req.Status,
		batchID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update batch status in database")
	}

	// Update batch status on blockchain
	txID, err := blockchainClient.UpdateBatchStatus(strconv.Itoa(batchID), req.Status)
	if err != nil {
		// Log the error but continue - blockchain is secondary to database
		fmt.Printf("Warning: Failed to update batch status on blockchain: %v\n", err)
	}

	// Record blockchain transaction
	if txID != "" {
		// Generate metadata hash
		metadata := map[string]interface{}{
			"batch_id": batchID,
			"status":   req.Status,
			"updated_at": time.Now(),
		}
		metadataHash, err := blockchainClient.HashData(metadata)
		if err != nil {
			fmt.Printf("Warning: Failed to generate metadata hash: %v\n", err)
		}

		// Save blockchain record
		_, err = db.DB.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "batch", batchID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch status updated successfully",
	})
}

// GenerateBatchQRCode generates a QR code for a batch
// @Summary Generate batch QR code
// @Description Generate a QR code for a shrimp larvae batch
// @Tags batches
// @Accept json
// @Produce image/png
// @Param batchId path string true "Batch ID"
// @Param gateway query string false "IPFS gateway to use (e.g., ipfs.io)"
// @Param format query string false "QR code format: 'ipfs', 'gateway', or 'trace' (default: 'trace')"
// @Param size query int false "QR code size in pixels (default: 256)"
// @Success 200 {file} byte[] "QR code as PNG image"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /batches/{batchId}/qr [get]
func GenerateBatchQRCode(c *fiber.Ctx) error {
	// Get batch ID from params
	batchIDStr := c.Params("batchId")
	if batchIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	batchID, err := strconv.Atoi(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
	}

	// Check if batch exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Get QR code format (ipfs, gateway, or trace)
	format := c.Query("format", "trace")
	if format != "ipfs" && format != "gateway" && format != "trace" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid format. Must be 'ipfs', 'gateway', or 'trace'")
	}

	// Get QR code size
	sizeStr := c.Query("size", "256")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 128 || size > 1024 {
		size = 256 // Default to 256px if invalid
	}

	// Generate QR data based on format
	var qrData string
	
	switch format {
	case "ipfs":
		// Use the standard IPFS URL format
		qrData = fmt.Sprintf("http://ipfs:/%d", batchID)
	case "gateway":
		// Check if a gateway is specified
		gateway := c.Query("gateway", "ipfs.io")
		// Create a gateway URL format
		qrData = fmt.Sprintf("https://%s/ipfs/%d", gateway, batchID)
	case "trace":
		// Create a web-friendly traceability URL
		// This should point to the frontend app that will display the traceability data
		baseURL := c.Query("baseURL", "https://trace.viechain.com")
		qrData = fmt.Sprintf("%s/trace/%d", baseURL, batchID)
	}

	// Generate QR code
	qrCode, err := qrcode.Encode(qrData, qrcode.Medium, size)
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

// GetBatchEvents returns all events for a batch
// @Summary Get batch events
// @Description Retrieve all events for a shrimp larvae batch
// @Tags batches
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=[]models.Event}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /batches/{batchId}/events [get]
func GetBatchEvents(c *fiber.Ctx) error {
	// Get batch ID from params
	batchIDStr := c.Params("batchId")
	if batchIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	batchID, err := strconv.Atoi(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
	}

	// Check if batch exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Query events from database
	rows, err := db.DB.Query(`
		SELECT id, batch_id, event_type, actor_id, location, timestamp, metadata, updated_at, is_active
		FROM event
		WHERE batch_id = $1 AND is_active = true
		ORDER BY timestamp DESC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse events
	var events []models.Event
	for rows.Next() {
		var event models.Event
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
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse event data")
		}
		events = append(events, event)
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Events retrieved successfully",
		Data:    events,
	})
}

// GetBatchDocuments returns all documents for a batch
// @Summary Get batch documents
// @Description Retrieve all documents for a shrimp larvae batch
// @Tags batches
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=[]models.Document}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /batches/{batchId}/documents [get]
func GetBatchDocuments(c *fiber.Ctx) error {
	// Get batch ID from params
	batchIDStr := c.Params("batchId")
	if batchIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	batchID, err := strconv.Atoi(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
	}

	// Check if batch exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Query documents from database
	rows, err := db.DB.Query(`
		SELECT id, batch_id, doc_type, ipfs_hash, uploaded_by, uploaded_at, updated_at, is_active
		FROM document
		WHERE batch_id = $1 AND is_active = true
		ORDER BY uploaded_at DESC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse documents
	var documents []models.Document
	for rows.Next() {
		var doc models.Document
		err := rows.Scan(
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

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Documents retrieved successfully",
		Data:    documents,
	})
}

// GetBatchEnvironmentData returns all environment data for a batch
// @Summary Get batch environment data
// @Description Retrieve all environment data for a shrimp larvae batch
// @Tags batches
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=[]models.EnvironmentData}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /batches/{batchId}/environment [get]
func GetBatchEnvironmentData(c *fiber.Ctx) error {
	// Get batch ID from params
	batchIDStr := c.Params("batchId")
	if batchIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	batchID, err := strconv.Atoi(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
	}

	// Check if batch exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Query environment data from database with related information
	rows, err := db.DB.Query(`
		SELECT 
			e.id, e.batch_id, e.temperature, e.pH, e.salinity, e.density, e.age, e.timestamp, e.updated_at, e.is_active,
			b.species, b.quantity, b.status,
			h.name AS hatchery_name, h.location AS hatchery_location,
			c.name AS company_name,
			u.username AS recorded_by,
			br.tx_id AS blockchain_tx_id,
			br.metadata_hash AS blockchain_metadata
		FROM environment_data e
		INNER JOIN batch b ON e.batch_id = b.id
		INNER JOIN hatchery h ON b.hatchery_id = h.id
		INNER JOIN company c ON h.company_id = c.id
		LEFT JOIN account u ON e.recorded_by = u.id
		LEFT JOIN blockchain_record br ON br.related_table = 'environment' AND br.related_id = e.id
		WHERE e.batch_id = $1 AND e.is_active = true
		ORDER BY e.timestamp DESC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse environment data
	var envDataList []map[string]interface{}
	for rows.Next() {
		var (
			envData models.EnvironmentData
			species, status, hatcheryName, hatcheryLocation, companyName, recordedBy string
			blockchainTxID, blockchainMetadata sql.NullString
			quantity int
		)
		err := rows.Scan(
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
			&species,
			&quantity,
			&status,
			&hatcheryName,
			&hatcheryLocation,
			&companyName,
			&recordedBy,
			&blockchainTxID,
			&blockchainMetadata,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse environment data")
		}

		// Create comprehensive data structure
		envDataEntry := map[string]interface{}{
			"id": envData.ID,
			"environment_data": map[string]interface{}{
				"temperature": envData.Temperature,
				"ph":         envData.PH,
				"salinity":   envData.Salinity,
				"density":    envData.Density,
				"age":        envData.Age,
				"timestamp":  envData.Timestamp,
				"updated_at": envData.UpdatedAt,
				"is_active":  envData.IsActive,
			},
			"batch_info": map[string]interface{}{
				"id":       envData.BatchID,
				"species":  species,
				"quantity": quantity,
				"status":   status,
			},
			"facility_info": map[string]interface{}{
				"hatchery_name":     hatcheryName,
				"hatchery_location": hatcheryLocation,
				"company_name":      companyName,
			},
			"metadata": map[string]interface{}{
				"recorded_by": recordedBy,
			},
		}

		// Add blockchain verification if available
		if blockchainTxID.Valid {
			envDataEntry["blockchain_verification"] = map[string]interface{}{
				"tx_id":         blockchainTxID.String,
				"metadata_hash": blockchainMetadata.String,
				"explorer_url": fmt.Sprintf("https://explorer.viechain.com/tx/%s", blockchainTxID.String),
			}
		}

		envDataList = append(envDataList, envDataEntry)
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Environment data retrieved successfully",
		Data:    envDataList,
	})
}

// GetBatchHistory returns the blockchain transactions for a batch
// @Summary Get batch blockchain history
// @Description Retrieve the blockchain history for a shrimp larvae batch
// @Tags batches
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=[]models.BlockchainRecord}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /batches/{batchId}/history [get]
func GetBatchHistory(c *fiber.Ctx) error {
	// Get batch ID from params
	batchIDStr := c.Params("batchId")
	if batchIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	batchID, err := strconv.Atoi(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
	}

	// Check if batch exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Query blockchain records from database
	rows, err := db.DB.Query(`
		SELECT id, related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active
		FROM blockchain_record
		WHERE (related_table = 'batch' AND related_id = $1) OR 
              EXISTS (SELECT 1 FROM event WHERE id = related_id AND related_table = 'event' AND batch_id = $1) OR
              EXISTS (SELECT 1 FROM document WHERE id = related_id AND related_table = 'document' AND batch_id = $1) OR
              EXISTS (SELECT 1 FROM environment WHERE id = related_id AND related_table = 'environment' AND batch_id = $1)
		ORDER BY created_at DESC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse blockchain records
	var records []models.BlockchainRecord
	for rows.Next() {
		var record models.BlockchainRecord
		err := rows.Scan(
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
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse blockchain record data")
		}
		records = append(records, record)
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch blockchain history retrieved successfully",
		Data:    records,
	})
}

// GetBatchQRCode returns a QR code for a batch
// @Summary Get batch QR code
// @Description Generate a QR code for a batch that contains blockchain verification data
// @Tags batches
// @Accept json
// @Produce png,json
// @Param batchId path string true "Batch ID"
// @Param format query string false "Response format (png or json)"
// @Success 200 {file} binary "QR code image"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /batches/{batchId}/qr [get]
func GetBatchQRCode(c *fiber.Ctx) error {
	batchID := c.Params("batchId")
	if batchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	format := c.Query("format", "png")
	if format != "png" && format != "json" {
		return fiber.NewError(fiber.StatusBadRequest, "Format must be png or json")
	}
	
	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}
	
	// Get batch details including NFT information
	var species, status, hatcheryID string
	var isTokenized bool
	var nftTokenID sql.NullInt64
	var nftContract sql.NullString
	var createdAt time.Time
	
	err = db.DB.QueryRow(`
		SELECT species, status, hatchery_id, created_at, is_tokenized, nft_token_id, nft_contract 
		FROM batch 
		WHERE id = $1
	`, batchID).Scan(&species, &status, &hatcheryID, &createdAt, &isTokenized, &nftTokenID, &nftContract)
	
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve batch details")
	}
	
	// Get blockchain information for this batch
	var blockchainTxID, metadataHash sql.NullString
	err = db.DB.QueryRow(`
		SELECT tx_id, metadata_hash
		FROM blockchain_record
		WHERE related_table = 'batch' AND related_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, batchID).Scan(&blockchainTxID, &metadataHash)
	
	// Create QR data with verification information
	qrData := map[string]interface{}{
		"batch_id":            batchID,
		"species":             species,
		"status":              status,
		"created_at":          createdAt.Format(time.RFC3339),
		"verification_url":    fmt.Sprintf("https://trace.viechain.com/verify/%s", batchID),
		"blockchain_verified": blockchainTxID.Valid,
	}
	
	// Get transfer history for this batch
	rows, err := db.DB.Query(`
		SELECT id, source_type, destination_id, destination_type, 
		       quantity, transferred_at, status, blockchain_tx_id
		FROM shipment_transfer
		WHERE batch_id = $1 AND is_active = true
		ORDER BY transferred_at DESC
	`, batchID)
	
	if err == nil {
		defer rows.Close()
		
		var transfers []map[string]interface{}
		for rows.Next() {
			var transferID, sourceType, destinationID, destinationType, status, blockchainTxID string
			var quantity int
			var transferredAt time.Time
			
			err := rows.Scan(
				&transferID,
				&sourceType,
				&destinationID,
				&destinationType,
				&quantity,
				&transferredAt,
				&status,
				&blockchainTxID,
			)
			
			if err == nil {
				transfers = append(transfers, map[string]interface{}{
					"transfer_id":       transferID,
					"source":            fmt.Sprintf("%s (%s)", sourceType),
					"destination":       fmt.Sprintf("%s (%s)", destinationID, destinationType),
					"quantity":          quantity,
					"transferred_at":    transferredAt.Format(time.RFC3339),
					"status":            status,
					"blockchain_verified": blockchainTxID != "",
				})
			}
		}
		
		if len(transfers) > 0 {
			qrData["transfer_history"] = transfers
		}
	}
	
	// Add NFT information if tokenized
	if isTokenized && nftTokenID.Valid && nftContract.Valid {
		qrData["nft"] = map[string]interface{}{
			"is_tokenized":    true,
			"token_id":        nftTokenID.Int64,
			"contract":        nftContract.String,
			"marketplace_url": fmt.Sprintf("https://marketplace.viechain.com/token/%s/%d", 
				nftContract.String, nftTokenID.Int64),
		}
	} else {
		qrData["nft"] = map[string]interface{}{
			"is_tokenized": false,
		}
	}
	
	// Add blockchain verification data if available
	if blockchainTxID.Valid {
		qrData["blockchain"] = map[string]interface{}{
			"tx_id":        blockchainTxID.String,
			"metadata_hash": metadataHash.String,
			"explorer_url": fmt.Sprintf("https://explorer.viechain.com/tx/%s", blockchainTxID.String),
		}
	}
	
	// If JSON format is requested, return data directly
	if format == "json" {
		return c.JSON(qrData)
	}
	
	// For PNG format, generate QR code
	// Convert data to JSON string
	jsonData, err := json.Marshal(qrData)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR data")
	}
	
	// Generate QR code
	qr, err := qrcode.New(string(jsonData), qrcode.Medium)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR code")
	}
	
	// Set QR code size
	qr.DisableBorder = false
	
	// Create PNG image
	png, err := qr.PNG(256)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR image")
	}
	
	// Set content type and return image
	c.Set("Content-Type", "image/png")
	return c.Send(png)
}

// Helper function to convert string to int
func convertToInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}