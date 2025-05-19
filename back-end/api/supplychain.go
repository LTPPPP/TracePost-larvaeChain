package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"github.com/LTPPPP/TracePost-larvaeChain/models"
	"github.com/skip2/go-qrcode"
)

// SupplyChainDetails represents the request for getting full supply chain details
type SupplyChainDetails struct {
	ID                string                      `json:"id"`
	Species           string                      `json:"species"`
	Status            string                      `json:"status"`
	CreatedAt         time.Time                   `json:"created_at"`
	HatcheryDetails   map[string]interface{}      `json:"hatchery_details,omitempty"`
	FarmDetails       []map[string]interface{}    `json:"farm_details,omitempty"`
	ProcessorDetails  []map[string]interface{}    `json:"processor_details,omitempty"`
	ExporterDetails   []map[string]interface{}    `json:"exporter_details,omitempty"`
	NFTDetails        map[string]interface{}      `json:"nft_details,omitempty"`
	TransferHistory   []models.ShipmentTransfer   `json:"transfer_history,omitempty"`
	BlockchainRecords []map[string]interface{}    `json:"blockchain_records,omitempty"`
	Events            []map[string]interface{}    `json:"events,omitempty"`
}

// GetSupplyChainDetails retrieves the complete supply chain journey for a batch
// @Summary Get complete supply chain details
// @Description Retrieve the complete journey of a batch through the supply chain
// @Tags supplychain
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=SupplyChainDetails}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /supplychain/{batchId} [get]
func GetSupplyChainDetails(c *fiber.Ctx) error {
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
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}
	
	// Initialize supply chain details
	var details SupplyChainDetails
	details.ID = batchIDStr
	
	// Get batch basic information
	err = db.DB.QueryRow(`
		SELECT species, status, created_at, is_tokenized, nft_token_id, nft_contract
		FROM batch
		WHERE id = $1
	`, batchID).Scan(
		&details.Species,
		&details.Status,
		&details.CreatedAt,
		&exists, // is_tokenized
		&details.NFTDetails,
		&details.NFTDetails,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve batch details")
	}
	
	// If tokenized, format NFT details
	if exists {
		// NFT details already populated from the query above
	} else {
		details.NFTDetails = map[string]interface{}{"is_tokenized": false}
	}
	
	// 1. Get Hatchery Details
	var hatcheryID string
	err = db.DB.QueryRow("SELECT hatchery_id FROM batch WHERE id = $1", batchID).Scan(&hatcheryID)
	if err == nil && hatcheryID != "" {
		var hatcheryName, location, contact string
		err = db.DB.QueryRow(`
			SELECT name, location, contact
			FROM hatchery
			WHERE id = $1
		`, hatcheryID).Scan(&hatcheryName, &location, &contact)
		
		if err == nil {
			details.HatcheryDetails = map[string]interface{}{
				"id":       hatcheryID,
				"name":     hatcheryName,
				"location": location,
				"contact":  contact,
			}
			
			// Get hatchery-specific data
			rows, err := db.DB.Query(`
				SELECT recorded_at, record_type, description
				FROM hatchery_records
				WHERE batch_id = $1
				ORDER BY recorded_at
			`, batchID)
			
			if err == nil {
				defer rows.Close()
				
				var hatcheryRecords []map[string]interface{}
				for rows.Next() {
					var recordedAt time.Time
					var recordType, description string
					
					err := rows.Scan(&recordedAt, &recordType, &description)
					if err == nil {
						hatcheryRecords = append(hatcheryRecords, map[string]interface{}{
							"recorded_at": recordedAt,
							"type":        recordType,
							"description": description,
						})
					}
				}
				
				if len(hatcheryRecords) > 0 {
					details.HatcheryDetails["records"] = hatcheryRecords
				}
			}
		}
	}
	
	// 2. Get Farm Details
	rows, err := db.DB.Query(`
		SELECT f.id, f.name, f.location, fb.received_at, fb.quantity
		FROM farms f
		JOIN farm_batches fb ON f.id = fb.farm_id
		WHERE fb.batch_id = $1
		ORDER BY fb.received_at
	`, batchID)
	
	if err == nil {
		defer rows.Close()
		
		var farmDetails []map[string]interface{}
		for rows.Next() {
			var farmID, farmName, location string
			var receivedAt time.Time
			var quantity int
			
			err := rows.Scan(&farmID, &farmName, &location, &receivedAt, &quantity)
			if err == nil {
				farmDetail := map[string]interface{}{
					"id":          farmID,
					"name":        farmName,
					"location":    location,
					"received_at": receivedAt,
					"quantity":    quantity,
				}
				
				// Get farm records for this batch
				farmRecords, err := db.DB.Query(`
					SELECT id, record_type, recorded_at, description
					FROM farming_records
					WHERE farm_id = $1 AND batch_id = $2
					ORDER BY recorded_at
				`, farmID, batchID)
				
				if err == nil {
					defer farmRecords.Close()
					
					var records []map[string]interface{}
					for farmRecords.Next() {
						var recordID, recordType, description string
						var recordedAt time.Time
						
						err := farmRecords.Scan(&recordID, &recordType, &recordedAt, &description)
						if err == nil {
							records = append(records, map[string]interface{}{
								"id":          recordID,
								"type":        recordType,
								"recorded_at": recordedAt,
								"description": description,
							})
						}
					}
					
					if len(records) > 0 {
						farmDetail["records"] = records
					}
				}
				
				farmDetails = append(farmDetails, farmDetail)
			}
		}
		
		if len(farmDetails) > 0 {
			details.FarmDetails = farmDetails
		}
	}
	
	// 3. Get Processor Details
	rows, err = db.DB.Query(`
		SELECT p.id, p.name, p.location, pb.received_at, pb.quantity
		FROM processors p
		JOIN processor_batches pb ON p.id = pb.processor_id
		WHERE pb.batch_id = $1
		ORDER BY pb.received_at
	`, batchID)
	
	if err == nil {
		defer rows.Close()
		
		var processorDetails []map[string]interface{}
		for rows.Next() {
			var processorID, processorName, location string
			var receivedAt time.Time
			var quantity int
			
			err := rows.Scan(&processorID, &processorName, &location, &receivedAt, &quantity)
			if err == nil {
				processorDetail := map[string]interface{}{
					"id":          processorID,
					"name":        processorName,
					"location":    location,
					"received_at": receivedAt,
					"quantity":    quantity,
				}
				
				// Get processing records for this batch
				processingRecords, err := db.DB.Query(`
					SELECT id, process_type, processed_at, description
					FROM processing_records
					WHERE processor_id = $1 AND batch_id = $2
					ORDER BY processed_at
				`, processorID, batchID)
				
				if err == nil {
					defer processingRecords.Close()
					
					var records []map[string]interface{}
					for processingRecords.Next() {
						var recordID, processType, description string
						var processedAt time.Time
						
						err := processingRecords.Scan(&recordID, &processType, &processedAt, &description)
						if err == nil {
							records = append(records, map[string]interface{}{
								"id":           recordID,
								"type":         processType,
								"processed_at": processedAt,
								"description":  description,
							})
						}
					}
					
					if len(records) > 0 {
						processorDetail["records"] = records
					}
				}
				
				processorDetails = append(processorDetails, processorDetail)
			}
		}
		
		if len(processorDetails) > 0 {
			details.ProcessorDetails = processorDetails
		}
	}
	
	// 4. Get Transfer History
	rows, err = db.DB.Query(`
		SELECT id, batch_id, source_type, destination_id, destination_type, 
			   quantity, transferred_at, transferred_by, status, blockchain_tx_id,
			   nft_token_id, nft_contract_address, transfer_notes, metadata, 
			   created_at, updated_at, is_active
		FROM shipment_transfer
		WHERE batch_id = $1 AND is_active = true
		ORDER BY transferred_at DESC
	`, batchID)
	
	if err == nil {
		defer rows.Close()
		
		var transfers []models.ShipmentTransfer
		for rows.Next() {
			var transfer models.ShipmentTransfer
			err := rows.Scan(
				&transfer.ID,
				&transfer.BatchID,
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
			if err == nil {
				transfers = append(transfers, transfer)
			}
		}
		
		if len(transfers) > 0 {
			details.TransferHistory = transfers
		}
	}
	
	// 5. Get Blockchain Records
	rows, err = db.DB.Query(`
		SELECT tx_id, metadata_hash, created_at
		FROM blockchain_record
		WHERE related_table = 'batch' AND related_id = $1
		ORDER BY created_at DESC
	`, batchID)
	
	if err == nil {
		defer rows.Close()
		
		var blockchainRecords []map[string]interface{}
		for rows.Next() {
			var txID, metadataHash string
			var createdAt time.Time
			
			err := rows.Scan(&txID, &metadataHash, &createdAt)
			if err == nil {
				blockchainRecords = append(blockchainRecords, map[string]interface{}{
					"tx_id":          txID,
					"metadata_hash":  metadataHash,
					"created_at":     createdAt,
					"explorer_url":   fmt.Sprintf("https://explorer.viechain.com/tx/%s", txID),
				})
			}
		}
		
		if len(blockchainRecords) > 0 {
			details.BlockchainRecords = blockchainRecords
		}
	}
	
	// 6. Get Events Timeline
	rows, err = db.DB.Query(`
		SELECT id, event_type, location, actor_id, timestamp, metadata
		FROM events
		WHERE batch_id = $1
		ORDER BY timestamp DESC
	`, batchID)
	
	if err == nil {
		defer rows.Close()
		
		var events []map[string]interface{}
		for rows.Next() {
			var eventID, eventType, location, actorID string
			var timestamp time.Time
			var metadata map[string]interface{}
			
			err := rows.Scan(&eventID, &eventType, &location, &actorID, &timestamp, &metadata)
			if err == nil {
				events = append(events, map[string]interface{}{
					"id":        eventID,
					"type":      eventType,
					"location":  location,
					"actor_id":  actorID,
					"timestamp": timestamp,
					"metadata":  metadata,
				})
			}
		}
		
		if len(events) > 0 {
			details.Events = events
		}
	}
	
	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Supply chain details retrieved successfully",
		Data:    details,
	})
}

// GenerateSupplyChainQRCode generates a QR code for the complete supply chain journey
// @Summary Generate supply chain QR code
// @Description Generate a QR code with the complete supply chain journey data
// @Tags supplychain
// @Accept json
// @Produce image/png,application/json
// @Param batchId path string true "Batch ID"
// @Param format query string false "Format: 'png' or 'json' (default: 'png')"
// @Success 200 {file} byte[] "QR code image or JSON data"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /supplychain/{batchId}/qr [get]
func GenerateSupplyChainQRCode(c *fiber.Ctx) error {
	// Get batch ID from path
	batchIDStr := c.Params("batchId")
	if batchIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	batchID, err := strconv.Atoi(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
	}
	
	// Check format (png or json)
	format := c.Query("format", "png")
	if format != "png" && format != "json" {
		return fiber.NewError(fiber.StatusBadRequest, "Format must be png or json")
	}
	
	// Check if batch exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}
	
	// Get batch basic information
	var species, status, hatcheryID string
	var isTokenized bool
	var createdAt time.Time
	
	err = db.DB.QueryRow(`
		SELECT species, status, hatchery_id, created_at, is_tokenized
		FROM batch
		WHERE id = $1
	`, batchID).Scan(&species, &status, &hatcheryID, &createdAt, &isTokenized)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve batch details")
	}
	
	// Build QR data structure with comprehensive supply chain info
	qrData := map[string]interface{}{
		"batch_id":            batchID,
		"species":             species,
		"status":              status,
		"created_at":          createdAt.Format(time.RFC3339),
		"verification_url":    fmt.Sprintf("https://trace.viechain.com/verify/%d", batchID),
		"is_tokenized":        isTokenized,
	}
	
	// Get origin information
	if hatcheryID != "" {
		var hatcheryName, location string
		err = db.DB.QueryRow("SELECT name, location FROM hatchery WHERE id = $1", hatcheryID).Scan(&hatcheryName, &location)
		if err == nil {
			qrData["origin"] = map[string]interface{}{
				"type":     "hatchery",
				"id":       hatcheryID,
				"name":     hatcheryName,
				"location": location,
			}
		}
	}
	
	// Get transfer history
	rows, err := db.DB.Query(`
		SELECT id, source_type, destination_id, destination_type, 
		       quantity, transferred_at, status
		FROM shipment_transfer
		WHERE batch_id = $1 AND is_active = true
		ORDER BY transferred_at
	`, batchID)
	
	if err == nil {
		defer rows.Close()
		
		var transfers []map[string]interface{}
		for rows.Next() {
			var transferID, sourceType, destinationID, destinationType, status string
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
			)
			
			if err == nil {
				transfers = append(transfers, map[string]interface{}{
					"id":               transferID,
					"source":           sourceType,
					"destination":      fmt.Sprintf("%s (%s)", destinationID, destinationType),
					"quantity":         quantity,
					"transferred_at":   transferredAt.Format(time.RFC3339),
					"status":           status,
				})
			}
		}
		
		if len(transfers) > 0 {
			qrData["transfers"] = transfers
			
			// Current location is the last destination in the transfer history
			if len(transfers) > 0 && status == "transferred" {
				lastTransfer := transfers[len(transfers)-1]
				qrData["current_location"] = lastTransfer["destination"]
			}
		}
	}
	
	// Get blockchain verification information
	var txID, metadataHash string
	err = db.DB.QueryRow(`
		SELECT tx_id, metadata_hash
		FROM blockchain_record
		WHERE related_table = 'batch' AND related_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, batchID).Scan(&txID, &metadataHash)
	
	if err == nil {
		qrData["blockchain"] = map[string]interface{}{
			"verified":      true,
			"tx_id":         txID,
			"metadata_hash": metadataHash,
			"explorer_url": fmt.Sprintf("https://explorer.viechain.com/tx/%s", txID),
		}
	} else {
		qrData["blockchain"] = map[string]interface{}{
			"verified": false,
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
