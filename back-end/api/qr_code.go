package api

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/skip2/go-qrcode"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"strconv"
	"time"
)

// UnifiedTraceByQRCode is a single API that generates a QR code containing all information about a batch
// including its complete transport history and blockchain verification
// @Summary Unified batch QR code traceability
// @Description Generate a QR code with complete batch information from blockchain, including all transport history
// @Tags qr
// @Accept json
// @Produce image/png,application/json
// @Param batchId path string true "Batch ID"
// @Param format query string false "Format: 'png' or 'json' (default: 'png')"
// @Success 200 {file} byte[] "QR code image or JSON data"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /qr/unified/{batchId} [get]
func UnifiedTraceByQRCode(c *fiber.Ctx) error {
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
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// 1. Get batch details with hatchery information
	var batchInfo struct {
		ID               int       `json:"id"`
		HatcheryID       string    `json:"hatchery_id"`
		HatcheryName     string    `json:"hatchery_name"`
		HatcheryLocation string    `json:"hatchery_location"`
		Species          string    `json:"species"`
		Quantity         int       `json:"quantity"`
		Status           string    `json:"status"`
		CreatedAt        time.Time `json:"created_at"`
		IsTokenized      bool      `json:"is_tokenized"`
		TokenID          *int64    `json:"token_id,omitempty"`
		ContractAddress  *string   `json:"contract_address,omitempty"`
	}
	
	err = db.DB.QueryRow(`
		SELECT b.id, b.hatchery_id, h.name, h.location, b.species, b.quantity, b.status, 
		       b.created_at, b.is_tokenized, b.nft_token_id, b.nft_contract
		FROM batch b
		JOIN hatchery h ON b.hatchery_id = h.id
		WHERE b.id = $1 AND b.is_active = true
	`, batchID).Scan(
		&batchInfo.ID,
		&batchInfo.HatcheryID,
		&batchInfo.HatcheryName,
		&batchInfo.HatcheryLocation,
		&batchInfo.Species,
		&batchInfo.Quantity,
		&batchInfo.Status,
		&batchInfo.CreatedAt,
		&batchInfo.IsTokenized,
		&batchInfo.TokenID,
		&batchInfo.ContractAddress,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve batch data")
	}

	// 2. Get all events with actor information
	rows, err := db.DB.Query(`
		SELECT e.id, e.event_type, e.actor_id, e.location, e.timestamp, e.metadata,
			   a.username, a.role
		FROM event e
		JOIN account a ON e.actor_id = a.id
		WHERE e.batch_id = $1 AND e.is_active = true
		ORDER BY e.timestamp
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve events")
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		event := map[string]interface{}{}
		var id int
		var eventType, location, actorName, actorRole string
		var actorID int
		var timestamp time.Time
		var metadata []byte
		
		err := rows.Scan(
			&id,
			&eventType,
			&actorID,
			&location,
			&timestamp,
			&metadata,
			&actorName,
			&actorRole,
		)
		if err != nil {
			continue
		}

		event["id"] = id
		event["event_type"] = eventType
		event["actor_id"] = actorID
		event["location"] = location
		event["timestamp"] = timestamp.Format(time.RFC3339)
		event["actor_name"] = actorName
		event["actor_role"] = actorRole

		// Parse metadata if available
		if len(metadata) > 0 {
			var metadataMap map[string]interface{}
			if json.Unmarshal(metadata, &metadataMap) == nil {
				event["details"] = metadataMap
			}
		}

		events = append(events, event)
	}

	// 3. Get transfer history (logistics chain)
	rows, err = db.DB.Query(`
		SELECT id, source_type, destination_id, destination_type, 
			   quantity, transferred_at, status
		FROM shipment_transfer
		WHERE batch_id = $1 AND is_active = true
		ORDER BY transferred_at
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve transfer history")
	}
	defer rows.Close()

	var transfers []map[string]interface{}
	for rows.Next() {
		var transferID int
		var sourceType, destinationID, destinationType, status string
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
				"destination_id":   destinationID,
				"destination_type": destinationType,
				"destination":      fmt.Sprintf("%s (%s)", destinationID, destinationType),
				"quantity":         quantity,
				"transferred_at":   transferredAt.Format(time.RFC3339),
				"status":           status,
			})
		}
	}

	// 4. Get environment data
	rows, err = db.DB.Query(`
		SELECT id, temperature, pH, salinity, dissolved_oxygen, timestamp
		FROM environment
		WHERE batch_id = $1 AND is_active = true
		ORDER BY timestamp DESC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve environment data")
	}
	defer rows.Close()

	var environmentData []map[string]interface{}
	for rows.Next() {
		var id string
		var temperature, pH, salinity, dissolvedOxygen float64
		var timestamp time.Time
		
		err := rows.Scan(
			&id,
			&temperature,
			&pH,
			&salinity,
			&dissolvedOxygen,
			&timestamp,
		)
		
		if err == nil {
			environmentData = append(environmentData, map[string]interface{}{
				"id":                id,
				"temperature":       temperature,
				"pH":                pH,
				"salinity":          salinity,
				"dissolved_oxygen":  dissolvedOxygen,
				"timestamp":         timestamp.Format(time.RFC3339),
			})
		}
	}

	// 5. Get blockchain verification records
	blockchainRecords, err := getBlockchainRecordsForBatch(batchID)
	if err != nil {
		// Just log the error but continue, as blockchain records are not critical
		fmt.Printf("Warning: Failed to retrieve blockchain records: %v\n", err)
	}

	// 6. Get documents
	rows, err = db.DB.Query(`
		SELECT id, doc_type, ipfs_hash, uploaded_by, uploaded_at
		FROM document
		WHERE batch_id = $1 AND is_active = true
		ORDER BY uploaded_at DESC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve documents")
	}
	defer rows.Close()

	var documents []map[string]interface{}
	for rows.Next() {
		var id, docType, ipfsHash, uploadedBy string
		var uploadedAt time.Time
		
		err := rows.Scan(
			&id,
			&docType,
			&ipfsHash,
			&uploadedBy,
			&uploadedAt,
		)
		
		if err == nil {
			documents = append(documents, map[string]interface{}{
				"id":          id,
				"doc_type":    docType,
				"ipfs_hash":   ipfsHash,
				"uploaded_by": uploadedBy,
				"uploaded_at": uploadedAt.Format(time.RFC3339),
				"view_url":    fmt.Sprintf("https://ipfs.io/ipfs/%s", ipfsHash),
			})
		}
	}

	// Determine current location from transfers if available
	var currentLocation string
	if len(transfers) > 0 {
		lastTransfer := transfers[len(transfers)-1]
		if lastTransfer["status"] == "completed" {
			currentLocation = lastTransfer["destination"].(string)
		}
	}
	if currentLocation == "" {
		currentLocation = batchInfo.HatcheryLocation
	}

	// Create the complete response object
	response := map[string]interface{}{
		"batch": map[string]interface{}{
			"id":               batchInfo.ID,
			"species":          batchInfo.Species,
			"quantity":         batchInfo.Quantity,
			"status":           batchInfo.Status,
			"created_at":       batchInfo.CreatedAt.Format(time.RFC3339),
			"current_location": currentLocation,
			"origin": map[string]interface{}{
				"hatchery_id":   batchInfo.HatcheryID,
				"hatchery_name": batchInfo.HatcheryName,
				"location":      batchInfo.HatcheryLocation,
			},
		},
		"events":          events,
		"logistics":       transfers,
		"environment":     environmentData,
		"documents":       documents,
		"blockchain":      blockchainRecords,
		"verification_url": fmt.Sprintf("https://trace.viechain.com/verify/%d", batchID),
	}

	// Add NFT information if tokenized
	if batchInfo.IsTokenized && batchInfo.TokenID != nil && batchInfo.ContractAddress != nil {
		response["nft"] = map[string]interface{}{
			"is_tokenized":    true,
			"token_id":        *batchInfo.TokenID,
			"contract":        *batchInfo.ContractAddress,
			"marketplace_url": fmt.Sprintf("https://marketplace.viechain.com/token/%s/%d", 
				*batchInfo.ContractAddress, *batchInfo.TokenID),
		}
	} else {
		response["nft"] = map[string]interface{}{
			"is_tokenized": false,
		}
	}

	// If JSON format is requested, return data directly
	if format == "json" {
		return c.JSON(response)
	}
	
	// For PNG format, generate QR code
	// Convert data to JSON string
	jsonData, err := json.Marshal(response)
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

// Helper function to get blockchain records for a batch
func getBlockchainRecordsForBatch(batchID int) ([]map[string]interface{}, error) {
	// Query blockchain records for this batch and related entities
	rows, err := db.DB.Query(`
		SELECT id, related_table, related_id, tx_id, metadata_hash, created_at
		FROM blockchain_record
		WHERE (related_table = 'batch' AND related_id = $1) OR 
			  EXISTS (SELECT 1 FROM event WHERE id = related_id AND related_table = 'event' AND batch_id = $1) OR
			  EXISTS (SELECT 1 FROM document WHERE id = related_id AND related_table = 'document' AND batch_id = $1) OR
			  EXISTS (SELECT 1 FROM environment WHERE id = related_id AND related_table = 'environment' AND batch_id = $1)
		ORDER BY created_at DESC
	`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var id, relatedTable, relatedID, txID, metadataHash string
		var createdAt time.Time
		
		err := rows.Scan(
			&id,
			&relatedTable,
			&relatedID,
			&txID,
			&metadataHash,
			&createdAt,
		)
		
		if err == nil {
			records = append(records, map[string]interface{}{
				"id":            id,
				"related_table": relatedTable,
				"related_id":    relatedID,
				"tx_id":         txID,
				"metadata_hash": metadataHash,
				"created_at":    createdAt.Format(time.RFC3339),
				"explorer_url":  fmt.Sprintf("https://explorer.viechain.com/tx/%s", txID),
			})
		}
	}

	return records, nil
}
