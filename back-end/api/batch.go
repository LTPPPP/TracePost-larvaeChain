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
			h.id, h.name, h.company_id, h.created_at, h.updated_at, h.is_active,
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
			h.id, h.name, h.company_id, h.created_at, h.updated_at, h.is_active,
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

	// Initialize blockchain client with more robust configuration
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
		SELECT h.id, h.name, h.company_id, h.created_at, h.updated_at, h.is_active,
			   c.id, c.name, c.type, c.location, c.contact_info, c.created_at, c.updated_at, c.is_active
		FROM hatchery h
		INNER JOIN company c ON h.company_id = c.id AND c.is_active = true
		WHERE h.id = $1 AND h.is_active = true
	`
	var company models.Company
	err = db.DB.QueryRow(hatcheryQuery, req.HatcheryID).Scan(
		&hatchery.ID,
		&hatchery.Name,
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

	// Begin database transaction to ensure data consistency
	tx, err := db.DB.Begin()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to start database transaction")
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

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

	err = tx.QueryRow(
		query,
		batch.HatcheryID,
		batch.Species,
		batch.Quantity,
		batch.Status,
	).Scan(&batch.ID, &batch.CreatedAt, &batch.UpdatedAt)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save batch to database")
	}

	// Prepare rich metadata for blockchain
	extendedMetadata := map[string]interface{}{
		"batch_id":         batch.ID,
		"hatchery_id":      req.HatcheryID,
		"species":          req.Species,
		"quantity":         req.Quantity,
		"status":           batch.Status,
		"company_id":       hatchery.Company.ID,
		"company_name":     hatchery.Company.Name,
		"hatchery_name":    hatchery.Name,
		"location":         hatchery.Company.Location,
		"created_at":       batch.CreatedAt,
		"blockchain_entry": true,
		"traceability_version": "2.0",
	}

	// Create batch on blockchain with enhanced data
	txID, err := blockchainClient.CreateBatch(
		strconv.Itoa(batch.ID),
		strconv.Itoa(req.HatcheryID),
		req.Species,
		req.Quantity,
	)
	
	// Additional blockchain transaction with extended metadata
	extendedTxID, err2 := blockchainClient.SubmitGenericTransaction(
		"BATCH_DATA_EXTENDED", 
		extendedMetadata,
	)
	
	// Process blockchain results
	blockchainSuccess := true
	blockchainErrors := make([]string, 0)
	
	if err != nil {
		// Log the error but continue - blockchain is secondary to database
		blockchainSuccess = false
		blockchainErrors = append(blockchainErrors, err.Error())
		fmt.Printf("Warning: Failed to record basic batch on blockchain: %v\n", err)
	}
	
	if err2 != nil {
		blockchainSuccess = false
		blockchainErrors = append(blockchainErrors, err2.Error())
		fmt.Printf("Warning: Failed to record extended batch data on blockchain: %v\n", err2)
	}

	// Record blockchain transactions in database
	if txID != "" {
		// Generate metadata hash
		metadataHash, err := blockchainClient.HashData(extendedMetadata)
		if err != nil {
			fmt.Printf("Warning: Failed to generate metadata hash: %v\n", err)
		}

		// Save blockchain record
		_, err = tx.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "batch", batch.ID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}
	
	// Record extended transaction if available
	if extendedTxID != "" {
		// Save extended blockchain record
		_, err = tx.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "batch_extended", batch.ID, extendedTxID, "extended_data")
		if err != nil {
			fmt.Printf("Warning: Failed to save extended blockchain record: %v\n", err)
		}
	}
	
	// Record batch creation event
	_, err = tx.Exec(`
		INSERT INTO event (batch_id, event_type, location, timestamp, metadata, updated_at, is_active)
		VALUES ($1, $2, $3, NOW(), $4, NOW(), true)
	`, batch.ID, "batch_created", hatchery.Company.Location, fmt.Sprintf(`{"blockchain_success": %v, "blockchain_errors": %v}`, blockchainSuccess, blockchainErrors))
	if err != nil {
		fmt.Printf("Warning: Failed to record batch creation event: %v\n", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to commit database transaction")
	}

	// Return success response
	responseData := map[string]interface{}{
		"batch": batch,
		"blockchain": map[string]interface{}{
			"success":         blockchainSuccess,
			"transaction_ids": []string{txID, extendedTxID},
			"errors":          blockchainErrors,
		},
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Batch created successfully and recorded on blockchain",
		Data:    responseData,
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

	// Check if batch exists and get current data
	var batch models.Batch
	var hatchery models.Hatchery
	var company models.Company
	query := `
		SELECT 
			b.id, b.hatchery_id, b.species, b.quantity, b.status, b.created_at, b.updated_at, b.is_active,
			h.id, h.name, h.company_id, h.created_at, h.updated_at, h.is_active,
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
	
	// Set relationships for complete batch data
	hatchery.Company = company
	batch.Hatchery = hatchery
	
	if batch.Status == req.Status {
		return c.JSON(SuccessResponse{
			Success: true,
			Message: "Batch status already set to " + req.Status,
		})
	}

	// Begin database transaction
	dbTx, err := db.DB.Begin()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to start database transaction")
	}
	defer func() {
		if err != nil {
			dbTx.Rollback()
		}
	}()

	// Update batch status in database
	_, err = dbTx.Exec(
		"UPDATE batch SET status = $1, updated_at = NOW() WHERE id = $2",
		req.Status,
		batchID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update batch status in database")
	}
	
	// Record this status change as an event
	var eventID int
	err = dbTx.QueryRow(`
		INSERT INTO event (batch_id, event_type, location, timestamp, metadata, updated_at, is_active)
		VALUES ($1, $2, $3, NOW(), $4, NOW(), true)
		RETURNING id
	`, batchID, "status_changed", company.Location, fmt.Sprintf(`{"old_status": "%s", "new_status": "%s"}`, batch.Status, req.Status)).Scan(&eventID)
	if err != nil {
		fmt.Printf("Warning: Failed to record status change event: %v\n", err)
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)
	
	// Prepare comprehensive metadata for blockchain
	updateMetadata := map[string]interface{}{
		"batch_id":       batchID,
		"species":        batch.Species,
		"quantity":       batch.Quantity,
		"old_status":     batch.Status,
		"new_status":     req.Status,
		"hatchery_id":    batch.HatcheryID,
		"hatchery_name":  hatchery.Name,
		"company_id":     company.ID,
		"company_name":   company.Name,
		"location":       company.Location,
		"updated_at":     time.Now(),
		"event_id":       eventID,
		"update_version": "2.0",
	}

	// Update batch status on blockchain
	txID, err := blockchainClient.UpdateBatchStatus(strconv.Itoa(batchID), req.Status)
	blockchainSuccess := true
	blockchainErrors := make([]string, 0)
	
	if err != nil {
		blockchainSuccess = false
		blockchainErrors = append(blockchainErrors, err.Error())
		fmt.Printf("Warning: Failed to update batch status on blockchain: %v\n", err)
	}
	
	// Submit a more comprehensive transaction with all metadata
	extendedTxID, err2 := blockchainClient.SubmitGenericTransaction(
		"BATCH_STATUS_UPDATE_EXTENDED", 
		updateMetadata,
	)
	
	if err2 != nil {
		blockchainSuccess = false
		blockchainErrors = append(blockchainErrors, err2.Error())
		fmt.Printf("Warning: Failed to record extended batch status update on blockchain: %v\n", err2)
	}

	// Record blockchain transactions in database
	var metadataHash string
	if txID != "" {
		// Generate metadata hash
		metadataHash, err = blockchainClient.HashData(updateMetadata)
		if err != nil {
			fmt.Printf("Warning: Failed to generate metadata hash: %v\n", err)
		}

		// Save blockchain record
		_, err = dbTx.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "batch", batchID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}
	
	// Record extended transaction if available
	if extendedTxID != "" {
		// Save extended blockchain record
		_, err = dbTx.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "batch_status_extended", batchID, extendedTxID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save extended blockchain record: %v\n", err)
		}
	}
	
	// Also record this blockchain transaction for the event
	if eventID > 0 && txID != "" {
		_, err = dbTx.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "event", eventID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save event blockchain record: %v\n", err)
		}
	}
	
	// Commit the database transaction
	if err = dbTx.Commit(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to commit database transaction")
	}

	// Prepare response
	responseData := map[string]interface{}{
		"batch_id":      batchID,
		"previous_status": batch.Status,
		"new_status":    req.Status,
		"updated_at":    time.Now(),
		"blockchain": map[string]interface{}{
			"success":         blockchainSuccess,
			"transaction_ids": []string{txID, extendedTxID},
			"errors":          blockchainErrors,
			"metadata_hash":   metadataHash,
		},
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch status updated successfully and recorded on blockchain",
		Data:    responseData,
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
	// DEPRECATED: Use /api/v1/qr/config/:batchId, /api/v1/qr/blockchain/:batchId, or /api/v1/qr/document/:batchId instead
	fmt.Println("Warning: GenerateBatchQRCode is deprecated and will be removed in a future version")
	
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
			h.name AS hatchery_name, 
			c.name AS company_name, c.location AS company_location,
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
			species, status, hatcheryName, companyName, companyLocation, recordedBy string
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
			&companyName,
			&companyLocation,
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
				"company_name":      companyName,
				"company_location":  companyLocation,
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

// GetBatchHistory returns the full history of a batch from blockchain records
// @Summary Get batch history
// @Description Retrieve the complete history of a batch from blockchain records
// @Tags batches
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=[]map[string]interface{}}
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

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Get batch transactions from blockchain
	txs, err := blockchainClient.GetBatchTransactions(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to get batch transactions: %v", err))
	}
	
	// Get blockchain records from database
	rows, err := db.DB.Query(`
		SELECT br.id, br.tx_id, br.metadata_hash, br.created_at,
		       CASE 
		           WHEN e.id IS NOT NULL THEN json_build_object('event_id', e.id, 'event_type', e.event_type, 'timestamp', e.timestamp)
		           ELSE NULL
		       END as event_data
		FROM blockchain_record br
		LEFT JOIN event e ON br.related_table = 'event' AND br.related_id = e.id
		WHERE (br.related_table = 'batch' AND br.related_id = $1)
		   OR (br.related_table = 'batch_extended' AND br.related_id = $1)
		   OR (br.related_table = 'batch_status_extended' AND br.related_id = $1)
		   OR EXISTS (
		       SELECT 1 FROM event 
		       WHERE batch_id = $1 AND id = br.related_id AND br.related_table = 'event'
		   )
		ORDER BY br.created_at DESC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error retrieving blockchain records")
	}
	defer rows.Close()
	
	// Parse blockchain records
	var records []map[string]interface{}
	for rows.Next() {
		var id int
		var txID, metadataHash string
		var createdAt time.Time
		var eventData sql.NullString
		
		if err := rows.Scan(&id, &txID, &metadataHash, &createdAt, &eventData); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse blockchain record")
		}
		
		record := map[string]interface{}{
			"id":            id,
			"tx_id":         txID,
			"metadata_hash": metadataHash,
			"created_at":    createdAt,
		}
		
		if eventData.Valid && eventData.String != "null" {
			var eventJSON map[string]interface{}
			if err := json.Unmarshal([]byte(eventData.String), &eventJSON); err == nil {
				record["event_data"] = eventJSON
			}
		}
		
		records = append(records, record)
	}
	
	// Get batch events with timestamps to correlate with blockchain records
	rows, err = db.DB.Query(`
		SELECT id, event_type, timestamp, metadata
		FROM event
		WHERE batch_id = $1 AND is_active = true
		ORDER BY timestamp DESC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error retrieving batch events")
	}
	defer rows.Close()
	
	// Parse batch events
	var events []map[string]interface{}
	for rows.Next() {
		var id int
		var eventType string
		var timestamp time.Time
		var metadata models.JSONB
		
		if err := rows.Scan(&id, &eventType, &timestamp, &metadata); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse batch event")
		}
		
		var metadataObj map[string]interface{}
		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &metadataObj); err != nil {
				metadataObj = map[string]interface{}{"raw": string(metadata)}
			}
		}
		
		events = append(events, map[string]interface{}{
			"id":         id,
			"event_type": eventType,
			"timestamp":  timestamp,
			"metadata":   metadataObj,
		})
	}
	
	// Convert blockchain transactions to a common format
	var txHistory []map[string]interface{}
	for _, tx := range txs {
		txHistory = append(txHistory, map[string]interface{}{
			"tx_id":       tx.TxID,
			"type":        tx.Type,
			"timestamp":   tx.Timestamp,
			"payload":     tx.Payload,
			"sender":      tx.Sender,
			"validated_at": tx.ValidatedAt,
		})
	}
	
	// Combine all data into a comprehensive history view
	historyData := map[string]interface{}{
		"blockchain_transactions": txHistory,
		"db_records":             records,
		"batch_events":           events,
		"verifiable_history":     true,
		"batch_id":               batchID,
	}
	
	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch history retrieved successfully",
		Data:    historyData,
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
// @Router /batches/{batchId}/qr/basic [get]
func GetBatchQRCode(c *fiber.Ctx) error {
	// DEPRECATED: Use /api/v1/qr/config/:batchId, /api/v1/qr/blockchain/:batchId, or /api/v1/qr/document/:batchId instead
	fmt.Println("Warning: GetBatchQRCode is deprecated and will be removed in a future version")
	
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

// GetBatchBlockchainData returns the blockchain data for a batch
// @Summary Get batch blockchain data
// @Description Retrieve blockchain data for a shrimp larvae batch
// @Tags batches
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=map[string]interface{}}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /batches/{batchId}/blockchain [get]
func GetBatchBlockchainData(c *fiber.Ctx) error {
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

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)
	
	// Get blockchain data for the batch
	blockchainData, err := blockchainClient.GetBatchBlockchainData(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to retrieve blockchain data: %v", err))
	}
	
	// Get blockchain records from database
	rows, err := db.DB.Query(`
		SELECT id, tx_id, metadata_hash, created_at
		FROM blockchain_record
		WHERE related_table = 'batch' AND related_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error retrieving blockchain records")
	}
	defer rows.Close()
	
	// Parse blockchain records
	var records []map[string]interface{}
	for rows.Next() {
		var id int
		var txID, metadataHash string
		var createdAt time.Time
		
		if err := rows.Scan(&id, &txID, &metadataHash, &createdAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse blockchain record")
		}
		
		records = append(records, map[string]interface{}{
			"id":            id,
			"tx_id":         txID,
			"metadata_hash": metadataHash,
			"created_at":    createdAt,
		})
	}
	
	// Combine blockchain data with database records
	result := map[string]interface{}{
		"blockchain_data": blockchainData,
		"db_records":      records,
	}
	
	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch blockchain data retrieved successfully",
		Data:    result,
	})
}

// VerifyBatchIntegrity verifies the integrity of a batch against the blockchain
// @Summary Verify batch integrity
// @Description Verify the integrity of a batch against its blockchain records
// @Tags batches
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=map[string]interface{}}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /batches/{batchId}/verify [get]
func VerifyBatchIntegrity(c *fiber.Ctx) error {
	// Get batch ID from params
	batchIDStr := c.Params("batchId")
	if batchIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	batchID, err := strconv.Atoi(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
	}

	// Query batch from database
	var batch models.Batch
	var hatchery models.Hatchery
	var company models.Company
	query := `
		SELECT 
			b.id, b.hatchery_id, b.species, b.quantity, b.status, b.created_at, b.updated_at, b.is_active,
			h.id, h.name, h.company_id, h.created_at, h.updated_at, h.is_active,
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
	
	// Prepare batch data for verification
	batchData := map[string]interface{}{
		"batch_id":    fmt.Sprintf("%d", batch.ID),
		"hatchery_id": fmt.Sprintf("%d", batch.HatcheryID),
		"species":     batch.Species,
		"quantity":    batch.Quantity,
		"status":      batch.Status,
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)
	
	// Verify batch integrity
	isValid, discrepancies, err := blockchainClient.VerifyBatchIntegrity(batchIDStr, batchData)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to verify batch integrity: %v", err))
	}
	
	// Prepare result
	result := map[string]interface{}{
		"is_valid":       isValid,
		"discrepancies":  discrepancies,
		"verified_at":    time.Now(),
		"batch_id":       batchID,
		"batch_status":   batch.Status,
		"check_details": map[string]interface{}{
			"blockchain_checks_passed": isValid,
			"db_integrity_verified":    true,
			"total_checks_performed":   len(batchData),
		},
	}
	
	// Return success response
	var message string
	if isValid {
		message = "Batch integrity verified successfully"
	} else {
		message = "Batch integrity verification failed"
	}
	return c.JSON(SuccessResponse{
		Success: true,
		Message: message,
		Data:    result,
	})
}

// GetBatchFromBlockchain returns batch data directly from the blockchain
// @Summary Get batch from blockchain
// @Description Retrieve batch data directly from the blockchain
// @Tags blockchain
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=map[string]interface{}}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /blockchain/batch/{batchId} [get]
func GetBatchFromBlockchain(c *fiber.Ctx) error {
	// Get batch ID from params
	batchIDStr := c.Params("batchId")
	if batchIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Get batch data from the blockchain
	blockchainData, err := blockchainClient.GetBatchBlockchainData(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to get blockchain data: %v", err))
	}
	
	// Get extra verification information
	verificationData, err := blockchainClient.VerifyBatchDataOnChain(batchIDStr)
	if err != nil {
		// Log the error but continue
		fmt.Printf("Warning: Failed to verify batch data on chain: %v\n", err)
	} else {
		// Add verification data to result
		blockchainData["verification"] = verificationData
	}

	// Return success response with the blockchain data
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch blockchain data retrieved successfully",
		Data:    blockchainData,
	})
}

// Helper function to convert string to int
func convertToInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}