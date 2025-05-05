package api

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/skip2/go-qrcode"
	"github.com/vietchain/tracepost-larvae/blockchain"
	"github.com/vietchain/tracepost-larvae/db"
	"github.com/vietchain/tracepost-larvae/models"
)

// CreateBatchRequest represents a request to create a new batch
type CreateBatchRequest struct {
	HatcheryID string `json:"hatchery_id"`
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
		SELECT id, batch_id, hatchery_id, creation_date, species, quantity, status, blockchain_tx_id, metadata_hash
		FROM batches
		ORDER BY creation_date DESC
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
	batchID := c.Params("batchId")
	if batchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}

	// Query batch from database
	var batch models.Batch
	query := `
		SELECT id, batch_id, hatchery_id, creation_date, species, quantity, status, blockchain_tx_id, metadata_hash
		FROM batches
		WHERE batch_id = $1
	`
	err := db.DB.QueryRow(query, batchID).Scan(
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
	if req.HatcheryID == "" || req.Species == "" || req.Quantity <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Hatchery ID, species, and quantity are required")
	}

	// Generate a unique batch ID
	timestamp := time.Now().Unix()
	batchID := fmt.Sprintf("BATCH-%s-%d", req.HatcheryID, timestamp)

	// Initialize blockchain client
	// In a real application, this would be properly configured
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain", // Add chainID parameter
		"poa",
	)

	// Create batch on blockchain
	txID, err := blockchainClient.CreateBatch(batchID, req.HatcheryID, req.Species, req.Quantity)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record batch on blockchain")
	}

	// Generate metadata hash
	metadata := map[string]interface{}{
		"batch_id":     batchID,
		"hatchery_id":  req.HatcheryID,
		"species":      req.Species,
		"quantity":     req.Quantity,
		"created_at":   time.Now(),
	}
	metadataHash, err := blockchainClient.HashData(metadata)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate metadata hash")
	}

	// Convert string hatcheryID to int
	hatcheryIDInt, err := convertToInt(req.HatcheryID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid hatchery ID format, must be an integer")
	}

	// Insert batch into database
	query := `
		INSERT INTO batches (batch_id, hatchery_id, creation_date, species, quantity, status, blockchain_tx_id, metadata_hash)
		VALUES ($1, $2, NOW(), $3, $4, $5, $6, $7)
		RETURNING id, creation_date
	`
	var batch models.Batch
	batch.BatchID = batchID
	batch.HatcheryID = hatcheryIDInt
	batch.Species = req.Species
	batch.Quantity = req.Quantity
	batch.Status = "created"
	batch.BlockchainTxID = txID
	batch.MetadataHash = metadataHash

	err = db.DB.QueryRow(
		query,
		batch.BatchID,
		batch.HatcheryID,
		batch.Species,
		batch.Quantity,
		batch.Status,
		batch.BlockchainTxID,
		batch.MetadataHash,
	).Scan(&batch.ID, &batch.CreationDate)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save batch to database")
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
	batchID := c.Params("batchId")
	if batchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
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
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1)", batchID).Scan(&exists)
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
		"tracepost-chain", // Add chainID parameter
		"poa",             // Add consensusType parameter (proof of authority)
	)

	// Update batch status on blockchain
	txID, err := blockchainClient.UpdateBatchStatus(batchID, req.Status)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update batch status on blockchain")
	}

	// Update batch status in database
	_, err = db.DB.Exec(
		"UPDATE batches SET status = $1, blockchain_tx_id = $2 WHERE batch_id = $3",
		req.Status,
		txID,
		batchID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update batch status in database")
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
	batchID := c.Params("batchId")
	if batchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}

	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Generate QR code data
	// This would typically be a URL to a public tracing page
	qrData := fmt.Sprintf("https://tracepost.example.com/trace/%s", batchID)

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
	batchID := c.Params("batchId")
	if batchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}

	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Query events from database
	rows, err := db.DB.Query(`
		SELECT id, batch_id, event_type, timestamp, location, actor_id, details, blockchain_tx_id, metadata_hash
		FROM events
		WHERE batch_id = $1
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
	batchID := c.Params("batchId")
	if batchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}

	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Query documents from database
	rows, err := db.DB.Query(`
		SELECT id, batch_id, document_type, ipfs_hash, upload_date, issuer, is_verified, blockchain_tx_id
		FROM documents
		WHERE batch_id = $1
		ORDER BY upload_date DESC
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
	batchID := c.Params("batchId")
	if batchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}

	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Query environment data from database
	rows, err := db.DB.Query(`
		SELECT id, batch_id, timestamp, temperature, ph, salinity, dissolved_oxygen, other_params, blockchain_tx_id
		FROM environment_data
		WHERE batch_id = $1
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

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Environment data retrieved successfully",
		Data:    envDataList,
	})
}

// GetBatchHistory returns the blockchain history for a batch
// @Summary Get batch blockchain history
// @Description Retrieve the blockchain history for a shrimp larvae batch
// @Tags batches
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=[]blockchain.Transaction}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /batches/{batchId}/history [get]
func GetBatchHistory(c *fiber.Ctx) error {
	// Get batch ID from params
	batchID := c.Params("batchId")
	if batchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}

	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1)", batchID).Scan(&exists)
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
		"tracepost-chain", // Add chainID parameter
		"poa",             // Add consensusType parameter (proof of authority)
	)

	// Get batch history from blockchain
	transactions, err := blockchainClient.GetBatchHistory(batchID)
	if err != nil {
		// This is just a mock implementation, in a real app we'd handle this more gracefully
		return c.JSON(SuccessResponse{
			Success: true,
			Message: "Blockchain history is not yet available in this version",
			Data:    []blockchain.Transaction{}, // Return empty array for now
		})
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch history retrieved successfully",
		Data:    transactions,
	})
}