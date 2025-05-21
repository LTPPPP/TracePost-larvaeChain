package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/config"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"github.com/LTPPPP/TracePost-larvaeChain/models"
	"github.com/skip2/go-qrcode"
)

// CreateShipmentTransferRequest represents a request to create a shipment transfer
type CreateShipmentTransferRequest struct {
	BatchID      int       `json:"batch_id"`
	SenderID     int       `json:"sender_id"`
	ReceiverID   int       `json:"receiver_id"`
	TransferTime time.Time `json:"transfer_time,omitempty"`
	Status       string    `json:"status,omitempty"`
}

// UpdateShipmentTransferRequest represents a request to update a shipment transfer

type UpdateShipmentTransferRequest struct {
	ReceiverID   int       `json:"receiver_id,omitempty"`
	TransferTime time.Time `json:"transfer_time,omitempty"`
	Status       string    `json:"status,omitempty"`
}

// GetAllShipmentTransfers retrieves all shipment transfers
// @Summary Get all shipment transfers
// @Description Retrieve all shipment transfers
// @Tags shipments
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse{data=[]models.ShipmentTransfer}
// @Failure 500 {object} ErrorResponse
// @Router /shipments/transfers [get]
func GetAllShipmentTransfers(c *fiber.Ctx) error {
	// Query transfers from database
	rows, err := db.DB.Query(`
		SELECT id, batch_id, sender_id, receiver_id, transfer_time, status,
			   created_at, updated_at, is_active
		FROM shipment_transfer
		WHERE is_active = true
		ORDER BY transfer_time DESC
	`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	defer rows.Close()

	// Parse transfers
	var transfers []models.ShipmentTransfer
	for rows.Next() {
		var transfer models.ShipmentTransfer
		err := rows.Scan(
			&transfer.ID,
			&transfer.BatchID,
			&transfer.SenderID,
			&transfer.ReceiverID,
			&transfer.TransferTime,
			&transfer.Status,
			&transfer.CreatedAt,
			&transfer.UpdatedAt,
			&transfer.IsActive,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse transfer data: "+err.Error())
		}
		transfers = append(transfers, transfer)
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Shipment transfers retrieved successfully",
		Data:    transfers,
	})
}

// GetShipmentTransferByID retrieves a specific shipment transfer by ID
// @Summary Get shipment transfer by ID
// @Description Retrieve a shipment transfer by its ID
// @Tags shipments
// @Accept json
// @Produce json
// @Param id path string true "Transfer ID"
// @Success 200 {object} SuccessResponse{data=models.ShipmentTransfer}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /shipments/transfers/{id} [get]
func GetShipmentTransferByID(c *fiber.Ctx) error {
	// Get transfer ID from path
	transferID := c.Params("id")
	if transferID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Transfer ID is required")
	}

	// Query transfer from database
	var transfer models.ShipmentTransfer
	err := db.DB.QueryRow(`
		SELECT id, batch_id, sender_id, receiver_id, transfer_time, status,
			   created_at, updated_at, is_active
		FROM shipment_transfer
		WHERE id = $1 AND is_active = true
	`, transferID).Scan(
		&transfer.ID,
		&transfer.BatchID,
		&transfer.SenderID,
		&transfer.ReceiverID,
		&transfer.TransferTime,
		&transfer.Status,
		&transfer.CreatedAt,
		&transfer.UpdatedAt,
		&transfer.IsActive,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Transfer not found")
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Shipment transfer retrieved successfully",
		Data:    transfer,
	})
}

// GetTransfersByBatchID retrieves all transfers for a specific batch
// @Summary Get transfers by batch ID
// @Description Retrieve all shipment transfers for a specific batch
// @Tags shipments
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=[]models.ShipmentTransfer}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /shipments/transfers/batch/{batchId} [get]
func GetTransfersByBatchID(c *fiber.Ctx) error {
	// Get batch ID from path
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

	// Query transfers from database
	rows, err := db.DB.Query(`
		SELECT id, batch_id, sender_id, receiver_id, transfer_time, status,
			   created_at, updated_at, is_active
		FROM shipment_transfer
		WHERE batch_id = $1 AND is_active = true
		ORDER BY transfer_time DESC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	defer rows.Close()

	// Parse transfers
	var transfers []models.ShipmentTransfer
	for rows.Next() {
		var transfer models.ShipmentTransfer
		err := rows.Scan(
			&transfer.ID,
			&transfer.BatchID,
			&transfer.SenderID,
			&transfer.ReceiverID,
			&transfer.TransferTime,
			&transfer.Status,
			&transfer.CreatedAt,
			&transfer.UpdatedAt,
			&transfer.IsActive,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse transfer data: "+err.Error())
		}
		transfers = append(transfers, transfer)
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch transfers retrieved successfully",
		Data:    transfers,
	})
}

// CreateShipmentTransfer creates a new shipment transfer
// @Summary Create a shipment transfer
// @Description Create a new shipment transfer between a sender and receiver
// @Tags shipments
// @Accept json
// @Produce json
// @Param request body CreateShipmentTransferRequest true "Transfer details"
// @Success 201 {object} SuccessResponse{data=models.ShipmentTransfer}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /shipments/transfers [post]
func CreateShipmentTransfer(c *fiber.Ctx) error {
	// Parse request
	var req CreateShipmentTransferRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format: "+err.Error())
	}

	// Validate required fields
	if req.BatchID <= 0 || req.SenderID <= 0 || req.ReceiverID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID, sender ID, and receiver ID are required")
	}

	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", req.BatchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Check if sender exists
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM account WHERE id = $1 AND is_active = true)", req.SenderID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Sender not found")
	}

	// Check if receiver exists
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM account WHERE id = $1 AND is_active = true)", req.ReceiverID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Receiver not found")
	}

	now := time.Now()
	transferTime := req.TransferTime
	if transferTime.IsZero() {
		transferTime = now
	}

	status := req.Status
	if status == "" {
		status = "pending" // Default status
	}

	// Start a transaction
	tx, err := db.DB.Begin()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to start database transaction: "+err.Error())
	}

	// Insert transfer record
	var transferID int
	err = tx.QueryRow(`
		INSERT INTO shipment_transfer (
			batch_id, sender_id, receiver_id, transfer_time, status, 
			created_at, updated_at, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING id
	`,
		req.BatchID,
		req.SenderID,
		req.ReceiverID,
		transferTime,
		status,
		now,
		now,
		true,
	).Scan(&transferID)

	if err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create transfer record: "+err.Error())
	}

	// Create batch event - let the database generate the ID using SERIAL
	_, err = tx.Exec(`
		INSERT INTO event (batch_id, event_type, actor_id, location, timestamp, metadata, updated_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, 
		req.BatchID, 
		"batch_transfer_initiated", 
		req.SenderID, 
		"", // Location could be added as a parameter if needed
		now, 
		nil, // Metadata is not needed here
		now,
		true,
	)
	if err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create event record: "+err.Error())
	}

	// Update batch status
	_, err = tx.Exec("UPDATE batch SET status = 'in_transfer', updated_at = $1 WHERE id = $2", now, req.BatchID)
	if err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update batch status: "+err.Error())
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
	}

	// Get the created transfer
	var transfer models.ShipmentTransfer
	err = db.DB.QueryRow(`
		SELECT id, batch_id, sender_id, receiver_id, transfer_time, status, 
			   created_at, updated_at, is_active
		FROM shipment_transfer
		WHERE id = $1
	`, transferID).Scan(
		&transfer.ID,
		&transfer.BatchID,
		&transfer.SenderID,
		&transfer.ReceiverID,
		&transfer.TransferTime,
		&transfer.Status,
		&transfer.CreatedAt,
		&transfer.UpdatedAt,
		&transfer.IsActive,
	)
	if err != nil {
		return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
			Success: true,
			Message: "Shipment transfer created successfully, but failed to retrieve details",
			Data: map[string]interface{}{
				"id":            transferID,
				"batch_id":      req.BatchID,
				"sender_id":     req.SenderID,
				"receiver_id":   req.ReceiverID,
				"transfer_time": transferTime,
				"status":        status,
			},
		})
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Shipment transfer created successfully",
		Data:    transfer,
	})
}

// UpdateShipmentTransfer updates an existing shipment transfer
// @Summary Update a shipment transfer
// @Description Update an existing shipment transfer status and metadata
// @Tags shipments
// @Accept json
// @Produce json
// @Param id path string true "Transfer ID"
// @Param request body UpdateShipmentTransferRequest true "Transfer update details"
// @Success 200 {object} SuccessResponse{data=models.ShipmentTransfer}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /shipments/transfers/{id} [put]
func UpdateShipmentTransfer(c *fiber.Ctx) error {
	// Get transfer ID from path
	transferID := c.Params("id")
	if transferID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Transfer ID is required")
	}

	// Parse request
	var req UpdateShipmentTransferRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format: "+err.Error())
	}

	// Check if at least one field is provided for update
	if req.Status == "" && req.ReceiverID == 0 && req.TransferTime.IsZero() {
		return fiber.NewError(fiber.StatusBadRequest, "At least one field to update is required")
	}

	// Check if transfer exists
	var exists bool
	var batchID int
	var currentStatus string
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM shipment_transfer WHERE id = $1 AND is_active = true), batch_id, status FROM shipment_transfer WHERE id = $1", transferID).Scan(&exists, &batchID, &currentStatus)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Transfer not found")
	}

	// Get user ID from token, with fallback if not found
	var userIDStr string
	userIDValue := c.Locals("user_id")
	if userIDValue != nil {
		userIDStr = fmt.Sprintf("%v", userIDValue)
	} else {
		userIDStr = "system" // Or use some appropriate default value
	}
	
	// Try to convert userID to integer
	var userID int
	if userIDStr != "system" {
		userID, err = strconv.Atoi(userIDStr)
		if err != nil {
			// If conversion fails, just log it but continue with the update
			fmt.Printf("Warning: unable to convert user_id %s to integer: %v\n", userIDStr, err)
		}
	}
	
	now := time.Now()

	// Start a transaction
	tx, err := db.DB.Begin()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to start database transaction: "+err.Error())
	}

	// Update transfer record
	// Dynamically build the update query based on provided fields
	updateQuery := "UPDATE shipment_transfer SET updated_at = $1"
	updateParams := []interface{}{now}
	paramCounter := 2

	if req.Status != "" {
		updateQuery += fmt.Sprintf(", status = $%d", paramCounter)
		updateParams = append(updateParams, req.Status)
		paramCounter++
	}

	if req.ReceiverID != 0 {
		updateQuery += fmt.Sprintf(", receiver_id = $%d", paramCounter)
		updateParams = append(updateParams, req.ReceiverID)
		paramCounter++
	}

	if !req.TransferTime.IsZero() {
		updateQuery += fmt.Sprintf(", transfer_time = $%d", paramCounter)
		updateParams = append(updateParams, req.TransferTime)
		paramCounter++
	}

	updateQuery += " WHERE id = $" + strconv.Itoa(paramCounter)
	updateParams = append(updateParams, transferID)

	_, err = tx.Exec(updateQuery, updateParams...)
	if err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update transfer record: "+err.Error())
	}

	// Create event for status change if provided
	if req.Status != "" && req.Status != currentStatus {
		// Update batch status based on transfer status
		var batchStatus string
		switch req.Status {
		case "completed":
			batchStatus = "transferred"
		case "in_transit":
			batchStatus = "in_transit"
		case "rejected":
			batchStatus = "transfer_rejected"
		default:
			batchStatus = "in_transfer"
		}

		_, err = tx.Exec("UPDATE batch SET status = $1, updated_at = $2 WHERE id = $3", batchStatus, now, batchID)
		if err != nil {
			tx.Rollback()
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to update batch status: "+err.Error())
		}

		// Create batch event
		eventMetadata := map[string]interface{}{
			"old_status": currentStatus,
			"new_status": req.Status,
		}
		
		// Convert event metadata to JSON
		eventMetadataJSON, err := json.Marshal(eventMetadata)
		if err != nil {
			tx.Rollback()
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to marshal event metadata: "+err.Error())
		}
		
		// Let the database generate the ID using SERIAL
		_, err = tx.Exec(`
			INSERT INTO event (batch_id, event_type, actor_id, location, timestamp, metadata, updated_at, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, 
			batchID, 
			"batch_transfer_status_changed", 
			userID, 
			"", 
			now, 
			eventMetadataJSON,
			now,
			true,
		)
		if err != nil {
			tx.Rollback()
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to create event record: "+err.Error())
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
	}

	// Record on blockchain if status was updated
	if req.Status != "" && req.Status != currentStatus {
		cfg := config.GetConfig()
		blockchainClient := blockchain.NewBlockchainClient(
			cfg.BlockchainNodeURL,
			cfg.BlockchainPrivateKey,
			cfg.BlockchainAccount,
			cfg.BlockchainChainID,
			cfg.BlockchainConsensus,
		)
		txResult, err := blockchainClient.SubmitTransaction("SHIPMENT_TRANSFER_UPDATED", map[string]interface{}{
			"transfer_id":    transferID,
			"batch_id":       batchID,
			"old_status":     currentStatus,
			"new_status":     req.Status,
			"updated_by":     userIDStr,
			"timestamp":      now,
		})

		if err == nil && txResult != "" {
			// Update blockchain record
			_, err = db.DB.Exec(
				"INSERT INTO blockchain_record (related_table, related_id, tx_id, created_at, updated_at, is_active) VALUES ($1, $2, $3, $4, $5, $6)",
				"shipment_transfer",
				transferID,
				txResult,
				now,
				now,
				true,
			)
			if err != nil {
				// Just log the error but continue
				fmt.Printf("Failed to record blockchain transaction: %v\n", err)
			}
		}
	}

	// Get the updated transfer
	var transfer models.ShipmentTransfer
	err = db.DB.QueryRow(`
		SELECT id, batch_id, sender_id, receiver_id, transfer_time, status, 
		       created_at, updated_at, is_active
		FROM shipment_transfer
		WHERE id = $1
	`, transferID).Scan(
		&transfer.ID,
		&transfer.BatchID,
		&transfer.SenderID,
		&transfer.ReceiverID,
		&transfer.TransferTime,
		&transfer.Status,
		&transfer.CreatedAt,
		&transfer.UpdatedAt,
		&transfer.IsActive,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve updated transfer: "+err.Error())
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Shipment transfer updated successfully",
		Data:    transfer,
	})
}

// DeleteShipmentTransfer soft deletes a shipment transfer
// @Summary Delete a shipment transfer
// @Description Soft delete a shipment transfer (mark as inactive)
// @Tags shipments
// @Accept json
// @Produce json
// @Param id path string true "Transfer ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /shipments/transfers/{id} [delete]
func DeleteShipmentTransfer(c *fiber.Ctx) error {
	// Get transfer ID from path
	transferID := c.Params("id")
	if transferID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Transfer ID is required")
	}

	// Check if transfer exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM shipment_transfer WHERE id = $1 AND is_active = true)", transferID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Transfer not found")
	}

	// Soft delete the transfer
	_, err = db.DB.Exec("UPDATE shipment_transfer SET is_active = false, updated_at = $1 WHERE id = $2", time.Now(), transferID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete transfer: "+err.Error())
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Shipment transfer deleted successfully",
	})
}

// GenerateTransferQRCode generates a QR code for a shipment transfer
// @Summary Generate transfer QR code
// @Description Generate a QR code for a shipment transfer with embedded traceability data
// @Tags shipments
// @Accept json
// @Produce image/png,application/json
// @Param id path string true "Transfer ID"
// @Param format query string false "Format: 'png' or 'json' (default: 'png')"
// @Success 200 {file} byte[] "QR code image or JSON data"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /shipments/transfers/{id}/qr [get]
func GenerateTransferQRCode(c *fiber.Ctx) error {
	// Get transfer ID from path
	transferID := c.Params("id")
	if transferID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Transfer ID is required")
	}

	// Check format (png or json)
	format := c.Query("format", "png")
	if format != "png" && format != "json" {
		return fiber.NewError(fiber.StatusBadRequest, "Format must be png or json")
	}

	// Check if transfer exists and get details
	var transfer models.ShipmentTransfer
	err := db.DB.QueryRow(`
		SELECT id, batch_id, sender_id, receiver_id, transfer_time, status,
		       created_at, updated_at, is_active
		FROM shipment_transfer
		WHERE id = $1 AND is_active = true
	`, transferID).Scan(
		&transfer.ID,
		&transfer.BatchID,
		&transfer.SenderID,
		&transfer.ReceiverID,
		&transfer.TransferTime,
		&transfer.Status,
		&transfer.CreatedAt,
		&transfer.UpdatedAt,
		&transfer.IsActive,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Transfer not found")
	}

	// Get batch details
	var species, batchStatus string
	err = db.DB.QueryRow(`
		SELECT species, status
		FROM batch
		WHERE id = $1
	`, transfer.BatchID).Scan(&species, &batchStatus)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve batch details")
	}

	// Get sender and receiver details
	var senderName, receiverName string
	err = db.DB.QueryRow(`
		SELECT username
		FROM account
		WHERE id = $1
	`, transfer.SenderID).Scan(&senderName)
	if err != nil {
		senderName = "Unknown Sender"
	}

	err = db.DB.QueryRow(`
		SELECT username
		FROM account
		WHERE id = $1
	`, transfer.ReceiverID).Scan(&receiverName)
	if err != nil {
		receiverName = "Unknown Receiver"
	}

	// Check for blockchain record
	var blockchainTxID string
	db.DB.QueryRow(`
		SELECT tx_id
		FROM blockchain_record
		WHERE related_table = 'shipment_transfer' AND related_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, transferID).Scan(&blockchainTxID)

	// Construct QR data with traceability information
	qrData := map[string]interface{}{
		"transfer_id":        transfer.ID,
		"batch_id":           transfer.BatchID,
		"sender_id":          transfer.SenderID,
		"sender_name":        senderName,
		"receiver_id":        transfer.ReceiverID,
		"receiver_name":      receiverName,
		"status":             transfer.Status,
		"transfer_time":      transfer.TransferTime.Format(time.RFC3339),
		"species":            species,
		"verification_url":   fmt.Sprintf("https://trace.viechain.com/verify/transfer/%s", transferID),
		"blockchain_verified": blockchainTxID != "",
	}

	// Add blockchain verification data if available
	if blockchainTxID != "" {
		qrData["blockchain"] = map[string]interface{}{
			"tx_id":        blockchainTxID,
			"explorer_url": fmt.Sprintf("https://explorer.viechain.com/tx/%s", blockchainTxID),
		}
	}

	// If JSON format is requested, return data directly
	if format == "json" {
		return c.JSON(qrData)
	}

	// For PNG format, generate QR code
	jsonData, err := json.Marshal(qrData)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR data")
	}

	qrCode, err := qrcode.Encode(string(jsonData), qrcode.Medium, 256)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR code")
	}

	// Set content type and return image
	c.Set("Content-Type", "image/png")
	return c.Send(qrCode)
}
