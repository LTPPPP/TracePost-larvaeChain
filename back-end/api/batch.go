package api

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/skip2/go-qrcode"
	"github.com/vietchain/tracepost-larvae/blockchain"
	"github.com/vietchain/tracepost-larvae/db"
	"github.com/vietchain/tracepost-larvae/models"
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
	// Query batches from database
	rows, err := db.DB.Query(`
		SELECT id, hatchery_id, species, quantity, status, created_at, updated_at, is_active
		FROM batch
		WHERE is_active = true
		ORDER BY created_at DESC
	`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse batches
	var batches []models.Batch
	for rows.Next() {
		var batch models.Batch
		err := rows.Scan(
			&batch.ID,
			&batch.HatcheryID,
			&batch.Species,
			&batch.Quantity,
			&batch.Status,
			&batch.CreatedAt,
			&batch.UpdatedAt,
			&batch.IsActive,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse batch data")
		}
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

	// Query batch from database
	var batch models.Batch
	query := `
		SELECT id, hatchery_id, species, quantity, status, created_at, updated_at, is_active
		FROM batch
		WHERE id = $1 AND is_active = true
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
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "Batch not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

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
// @Produce png
// @Param batchId path string true "Batch ID"
// @Success 200 {file} QR code
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

	// Generate QR code data
	// This would typically be a URL to a public tracing page
	qrData := fmt.Sprintf("https://tracepost.example.com/trace/%d", batchID)

	// Generate QR code
	qrCode, err := qrcode.Encode(qrData, qrcode.Medium, 256)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR code")
	}

	// Set content type and send QR code
	c.Set("Content-Type", "image/png")
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

	// Query environment data from database
	rows, err := db.DB.Query(`
		SELECT id, batch_id, temperature, pH, salinity, dissolved_oxygen, timestamp, updated_at, is_active
		FROM environment
		WHERE batch_id = $1 AND is_active = true
		ORDER BY timestamp DESC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse environment data
	var envDataList []models.EnvironmentData
	for rows.Next() {
		var envData models.EnvironmentData
		err := rows.Scan(
			&envData.ID,
			&envData.BatchID,
			&envData.Temperature,
			&envData.PH,
			&envData.Salinity,
			&envData.DissolvedOxygen,
			&envData.Timestamp,
			&envData.UpdatedAt,
			&envData.IsActive,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse environment data")
		}
		envDataList = append(envDataList, envData)
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

// Helper function to convert string to int
func convertToInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}