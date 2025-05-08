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
	BatchID         int                    `json:"batch_id"`
	SourceID        string                 `json:"source_id"`
	SourceType      string                 `json:"source_type"`
	DestinationID   string                 `json:"destination_id"`
	DestinationType string                 `json:"destination_type"`
	Quantity        int                    `json:"quantity"`
	TransferNotes   string                 `json:"transfer_notes,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	AutoGenerateNFT bool                   `json:"auto_generate_nft,omitempty"`
}

// UpdateShipmentTransferRequest represents a request to update a shipment transfer
type UpdateShipmentTransferRequest struct {
	Status        string                 `json:"status,omitempty"`
	TransferNotes string                 `json:"transfer_notes,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
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
		SELECT id, batch_id, source_id, source_type, destination_id, destination_type, 
			   quantity, transferred_at, transferred_by, status, blockchain_tx_id,
			   nft_token_id, nft_contract_address, transfer_notes, metadata, 
			   created_at, updated_at, is_active
		FROM shipment_transfer
		WHERE is_active = true
		ORDER BY transferred_at DESC
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
			&transfer.SourceID,
			&transfer.SourceType,
			&transfer.DestinationID,
			&transfer.DestinationType,
			&transfer.Quantity,
			&transfer.TransferredAt,
			&transfer.TransferredBy,
			&transfer.Status,
			&transfer.BlockchainTxID,
			&transfer.NFTTokenID,
			&transfer.NFTContractAddress,
			&transfer.TransferNotes,
			&transfer.Metadata,
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
		SELECT id, batch_id, source_id, source_type, destination_id, destination_type, 
			   quantity, transferred_at, transferred_by, status, blockchain_tx_id,
			   nft_token_id, nft_contract_address, transfer_notes, metadata, 
			   created_at, updated_at, is_active
		FROM shipment_transfer
		WHERE id = $1 AND is_active = true
	`, transferID).Scan(
		&transfer.ID,
		&transfer.BatchID,
		&transfer.SourceID,
		&transfer.SourceType,
		&transfer.DestinationID,
		&transfer.DestinationType,
		&transfer.Quantity,
		&transfer.TransferredAt,
		&transfer.TransferredBy,
		&transfer.Status,
		&transfer.BlockchainTxID,
		&transfer.NFTTokenID,
		&transfer.NFTContractAddress,
		&transfer.TransferNotes,
		&transfer.Metadata,
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
		SELECT id, batch_id, source_id, source_type, destination_id, destination_type, 
			   quantity, transferred_at, transferred_by, status, blockchain_tx_id,
			   nft_token_id, nft_contract_address, transfer_notes, metadata, 
			   created_at, updated_at, is_active
		FROM shipment_transfer
		WHERE batch_id = $1 AND is_active = true
		ORDER BY transferred_at DESC
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
			&transfer.SourceID,
			&transfer.SourceType,
			&transfer.DestinationID,
			&transfer.DestinationType,
			&transfer.Quantity,
			&transfer.TransferredAt,
			&transfer.TransferredBy,
			&transfer.Status,
			&transfer.BlockchainTxID,
			&transfer.NFTTokenID,
			&transfer.NFTContractAddress,
			&transfer.TransferNotes,
			&transfer.Metadata,
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
// @Description Create a new shipment transfer with optional NFT generation
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
	if req.BatchID <= 0 || req.SourceID == "" || req.SourceType == "" || req.Quantity <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID, source ID, source type, and quantity are required")
	}

	// Check if batch exists
	var exists bool
	var batchStatus string
	var totalQuantity int
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true), status, quantity FROM batch WHERE id = $1", req.BatchID).Scan(&exists, &batchStatus, &totalQuantity)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Check if quantity is valid
	if req.Quantity > totalQuantity {
		return fiber.NewError(fiber.StatusBadRequest, "Transfer quantity exceeds available quantity")
	}

	// Get user ID from token
	userID := c.Locals("user_id").(string)
	now := time.Now()

	// Generate transfer ID
	transferID := fmt.Sprintf("tran-%s", now.Format("20060102150405"))

	// Start a transaction
	tx, err := db.DB.Begin()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to start database transaction: "+err.Error())
	}

	// Insert transfer record
	_, err = tx.Exec(`
		INSERT INTO shipment_transfer (
			id, batch_id, source_id, source_type, destination_id, destination_type, 
			quantity, transferred_at, transferred_by, status, transfer_notes, metadata, 
			created_at, updated_at, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`,
		transferID,
		req.BatchID,
		req.SourceID,
		req.SourceType,
		req.DestinationID,
		req.DestinationType,
		req.Quantity,
		now,
		userID,
		"initiated",
		req.TransferNotes,
		req.Metadata,
		now,
		now,
		true,
	)
	if err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create transfer record: "+err.Error())
	}

	// Create batch event
	_, err = tx.Exec(`
		INSERT INTO event (id, batch_id, event_type, actor_id, location, timestamp, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, 
		"evt-" + now.Format("20060102150405"), 
		req.BatchID, 
		"batch_transfer_initiated", 
		req.SourceID, 
		userID, 
		now, 
		req.Metadata,
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

	// Record on blockchain
	cfg := config.GetConfig()
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)

	txResult, err := blockchainClient.SubmitTransaction("SHIPMENT_TRANSFER_CREATED", map[string]interface{}{
		"transfer_id":      transferID,
		"batch_id":         req.BatchID,
		"source_id":        req.SourceID,
		"source_type":      req.SourceType,
		"destination_id":   req.DestinationID,
		"destination_type": req.DestinationType,
		"quantity":         req.Quantity,
		"transferred_by":   userID,
		"timestamp":        now,
	})

	var blockchainTxID string
	if err == nil && txResult != "" {
		blockchainTxID = txResult
		// Update transfer record with blockchain transaction ID
		db.DB.Exec("UPDATE shipment_transfer SET blockchain_tx_id = $1 WHERE id = $2", blockchainTxID, transferID)
	}

	// Generate NFT for this transfer if requested
	var nftTokenID int
	var nftContractAddress string
	if req.AutoGenerateNFT {
		// Get NFT contract configuration from database or config
		// In a real implementation, you would get this from a configuration or database
		var contractAddress string
		err := db.DB.QueryRow("SELECT contract_address FROM nft_contracts WHERE is_default = true AND is_active = true").Scan(&contractAddress)
		if err == nil && contractAddress != "" {
			// Initialize the BaaS service
			baasService := blockchain.NewBaaSService()
			if baasService != nil {
				// Generate QR code URL for this batch
				qrCodeURL := fmt.Sprintf("https://trace.viechain.com/api/v1/shipments/transfers/%s/qr", transferID)
				
				// Get batch details
				var species string
				err = db.DB.QueryRow("SELECT species FROM batch WHERE id = $1", req.BatchID).Scan(&species)
				if err == nil {
					// Get recipient address (use destination or a default)
					recipientAddress := "0x" + transferID // Use a proper recipient address in production
					
					// Prepare the contract call to mint NFT
					contractMethods := map[string]interface{}{
						"method": "mintBatchNFT",
						"params": []interface{}{
							transferID,
							recipientAddress,
							"", // Will be overridden with generated URI
						},
					}
					
					// First generate the token URI
					tokenURIResult, err := baasService.QueryContractState(
						cfg.BlockchainNetworkID,
						contractAddress,
						map[string]interface{}{
							"method": "generateTokenURI",
							"params": []interface{}{
								transferID,
								species,
								fmt.Sprintf("%s -> %s", req.SourceType, req.DestinationType),
								now.Unix(),
								qrCodeURL,
							},
						},
					)
					
					if err == nil {
						tokenURI, ok := tokenURIResult["result"].(string)
						if ok {
							// Update the method params with the token URI
							params := contractMethods["params"].([]interface{})
							params[2] = tokenURI
							contractMethods["params"] = params
							
							// Make the contract call to mint the NFT
							result, err := baasService.CallContractMethod(
								cfg.BlockchainNetworkID,
								contractAddress,
								contractMethods,
							)
							
							if err == nil {
								// Get the token ID from the result
								if tokenID, ok := result["token_id"].(float64); ok {
									nftTokenID = int(tokenID)
									nftContractAddress = contractAddress
									
									// Update the shipment transfer with NFT information
									db.DB.Exec(
										"UPDATE shipment_transfer SET nft_token_id = $1, nft_contract_address = $2 WHERE id = $3",
										nftTokenID,
										nftContractAddress,
										transferID,
									)
								}
							}
						}
					}
				}
			}
		}
	}

	// Get the created transfer
	var transfer models.ShipmentTransfer
	err = db.DB.QueryRow(`
		SELECT id, batch_id, source_id, source_type, destination_id, destination_type, 
			   quantity, transferred_at, transferred_by, status, blockchain_tx_id,
			   nft_token_id, nft_contract_address, transfer_notes, metadata, 
			   created_at, updated_at, is_active
		FROM shipment_transfer
		WHERE id = $1
	`, transferID).Scan(
		&transfer.ID,
		&transfer.BatchID,
		&transfer.SourceID,
		&transfer.SourceType,
		&transfer.DestinationID,
		&transfer.DestinationType,
		&transfer.Quantity,
		&transfer.TransferredAt,
		&transfer.TransferredBy,
		&transfer.Status,
		&transfer.BlockchainTxID,
		&transfer.NFTTokenID,
		&transfer.NFTContractAddress,
		&transfer.TransferNotes,
		&transfer.Metadata,
		&transfer.CreatedAt,
		&transfer.UpdatedAt,
		&transfer.IsActive,
	)
	if err != nil {
		// Return basic info if retrieval fails
		return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
			Success: true,
			Message: "Shipment transfer created successfully",
			Data: map[string]interface{}{
				"transfer_id":         transferID,
				"batch_id":            req.BatchID,
				"source_id":           req.SourceID,
				"destination_id":      req.DestinationID,
				"quantity":            req.Quantity,
				"transferred_at":      now,
				"blockchain_tx_id":    blockchainTxID,
				"nft_token_id":        nftTokenID,
				"nft_contract_address": nftContractAddress,
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
	if req.Status == "" && req.TransferNotes == "" && req.Metadata == nil {
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

	// Get user ID from token
	userID := c.Locals("user_id").(string)
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

	if req.TransferNotes != "" {
		updateQuery += fmt.Sprintf(", transfer_notes = $%d", paramCounter)
		updateParams = append(updateParams, req.TransferNotes)
		paramCounter++
	}

	if req.Metadata != nil {
		updateQuery += fmt.Sprintf(", metadata = $%d", paramCounter)
		updateParams = append(updateParams, req.Metadata)
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
		_, err = tx.Exec(`
			INSERT INTO event (id, batch_id, event_type, actor_id, location, timestamp, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, 
			"evt-" + now.Format("20060102150405"), 
			batchID, 
			"batch_transfer_status_changed", 
			userID, 
			"", 
			now, 
			map[string]interface{}{
				"old_status": currentStatus,
				"new_status": req.Status,
				"notes":      req.TransferNotes,
			},
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
			"updated_by":     userID,
			"timestamp":      now,
		})

		if err == nil && txResult != "" {
			// Update transfer record with blockchain transaction ID
			db.DB.Exec("UPDATE shipment_transfer SET blockchain_tx_id = $1 WHERE id = $2", txResult, transferID)
		}
	}

	// Get the updated transfer
	var transfer models.ShipmentTransfer
	err = db.DB.QueryRow(`
		SELECT id, batch_id, source_id, source_type, destination_id, destination_type, 
			   quantity, transferred_at, transferred_by, status, blockchain_tx_id,
			   nft_token_id, nft_contract_address, transfer_notes, metadata, 
			   created_at, updated_at, is_active
		FROM shipment_transfer
		WHERE id = $1
	`, transferID).Scan(
		&transfer.ID,
		&transfer.BatchID,
		&transfer.SourceID,
		&transfer.SourceType,
		&transfer.DestinationID,
		&transfer.DestinationType,
		&transfer.Quantity,
		&transfer.TransferredAt,
		&transfer.TransferredBy,
		&transfer.Status,
		&transfer.BlockchainTxID,
		&transfer.NFTTokenID,
		&transfer.NFTContractAddress,
		&transfer.TransferNotes,
		&transfer.Metadata,
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
		SELECT id, batch_id, source_id, source_type, destination_id, destination_type, 
			   quantity, transferred_at, transferred_by, status, blockchain_tx_id,
			   nft_token_id, nft_contract_address, transfer_notes
		FROM shipment_transfer
		WHERE id = $1 AND is_active = true
	`, transferID).Scan(
		&transfer.ID,
		&transfer.BatchID,
		&transfer.SourceID,
		&transfer.SourceType,
		&transfer.DestinationID,
		&transfer.DestinationType,
		&transfer.Quantity,
		&transfer.TransferredAt,
		&transfer.TransferredBy,
		&transfer.Status,
		&transfer.BlockchainTxID,
		&transfer.NFTTokenID,
		&transfer.NFTContractAddress,
		&transfer.TransferNotes,
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

	// Construct QR data with traceability information
	qrData := map[string]interface{}{
		"transfer_id":        transfer.ID,
		"batch_id":           transfer.BatchID,
		"source":             fmt.Sprintf("%s (%s)", transfer.SourceID, transfer.SourceType),
		"destination":        fmt.Sprintf("%s (%s)", transfer.DestinationID, transfer.DestinationType),
		"status":             transfer.Status,
		"quantity":           transfer.Quantity,
		"transferred_at":     transfer.TransferredAt.Format(time.RFC3339),
		"species":            species,
		"verification_url":   fmt.Sprintf("https://trace.viechain.com/verify/transfer/%s", transfer.ID),
		"blockchain_verified": transfer.BlockchainTxID != "",
	}

	// Add NFT information if tokenized
	if transfer.NFTTokenID > 0 && transfer.NFTContractAddress != "" {
		qrData["nft"] = map[string]interface{}{
			"token_id":        transfer.NFTTokenID,
			"contract":        transfer.NFTContractAddress,
			"marketplace_url": fmt.Sprintf("https://marketplace.viechain.com/token/%s/%d", 
				transfer.NFTContractAddress, transfer.NFTTokenID),
		}
	}

	// Add blockchain verification data if available
	if transfer.BlockchainTxID != "" {
		qrData["blockchain"] = map[string]interface{}{
			"tx_id":        transfer.BlockchainTxID,
			"explorer_url": fmt.Sprintf("https://explorer.viechain.com/tx/%s", transfer.BlockchainTxID),
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
