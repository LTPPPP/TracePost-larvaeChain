package api

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/skip2/go-qrcode"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"os"
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
// @Param size query int false "QR code size in pixels (default: 512)"
// @Param simplified query bool false "Generate a simplified QR code with only essential data (default: false)"
// @Success 200 {file} byte[] "QR code image or JSON data"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /qr/unified/{batchId} [get]
func UnifiedTraceByQRCode(c *fiber.Ctx) error {
	// DEPRECATED: This endpoint is deprecated. Please use the new specialized QR code endpoints:
	// - /api/v1/qr/config/:batchId - For configuration information
	// - /api/v1/qr/blockchain/:batchId - For blockchain traceability information
	// - /api/v1/qr/document/:batchId - For document IPFS links
	fmt.Println("Warning: UnifiedTraceByQRCode is deprecated and will be removed in a future version")
	
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
	
	// Get QR code size if provided
	sizeStr := c.Query("size", "512")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 || size > 2048 {
		// Default to 512 if invalid
		size = 512
	}
	
	// Check if simplified QR code is requested
	simplified := c.QueryBool("simplified", false)
	
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
		// NFT fields are removed since they don't exist in the database
	}
	err = db.DB.QueryRow(`
		SELECT b.id, b.hatchery_id, h.name, c.location, b.species, b.quantity, b.status, 
		       b.created_at
		FROM batch b
		JOIN hatchery h ON b.hatchery_id = h.id
		JOIN company c ON h.company_id = c.id
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
	)
	if err != nil {
		fmt.Printf("Database error retrieving batch %d: %v\n", batchID, err)
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to retrieve batch data: %v", err))
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
		SELECT id, sender_id, receiver_id, transfer_time, 
			   status
		FROM shipment_transfer
		WHERE batch_id = $1 AND is_active = true
		ORDER BY transfer_time
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve transfer history")
	}
	defer rows.Close()

	var transfers []map[string]interface{}
	for rows.Next() {
		var transferID int
		var senderID, receiverID, status string
		var transferTime time.Time
		
		err := rows.Scan(
			&transferID,
			&senderID,
			&receiverID,
			&transferTime,
			&status,
		)
				if err == nil {
			transfers = append(transfers, map[string]interface{}{
				"id":               transferID,
				"sender_id":        senderID,
				"receiver_id":      receiverID,
				"destination":      receiverID,
				"transferred_at":   transferTime.Format(time.RFC3339),
				"status":           status,
			})
		}
	}
	// 4. Get environment data
	rows, err = db.DB.Query(`
		SELECT id, temperature, ph, salinity, density, age, timestamp
		FROM environment_data
		WHERE batch_id = $1 AND is_active = true
		ORDER BY timestamp DESC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve environment data")
	}
	defer rows.Close()

	var environmentData []map[string]interface{}
	for rows.Next() {
		var id int
		var temperature, ph, salinity, density float64
		var age int
		var timestamp time.Time
		
		err := rows.Scan(
			&id,
			&temperature,
			&ph,
			&salinity,
			&density,
			&age,
			&timestamp,
		)
				if err == nil {
			environmentData = append(environmentData, map[string]interface{}{
				"id":                id,
				"temperature":       temperature,
				"ph":                ph,
				"salinity":          salinity,
				"density":           density,
				"age":               age,
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
			// Use Pinata gateway URL if available, otherwise fall back to IPFS gateway
			gatewayURL := os.Getenv("PINATA_GATEWAY_URL")
			if gatewayURL == "" {
				gatewayURL = os.Getenv("IPFS_GATEWAY_URL")
			}
			if gatewayURL == "" {
				gatewayURL = "https://gateway.pinata.cloud"
			}
			
			documents = append(documents, map[string]interface{}{
				"id":          id,
				"doc_type":    docType,
				"ipfs_hash":   ipfsHash,
				"uploaded_by": uploadedBy,
				"uploaded_at": uploadedAt.Format(time.RFC3339),
				"view_url":    fmt.Sprintf("%s/ipfs/%s", gatewayURL, ipfsHash),
			})
		}
	}
	// Determine current location from transfers if available
	var currentLocation string
	if len(transfers) > 0 {
		lastTransfer := transfers[len(transfers)-1]
		if lastTransfer["status"] == "completed" {
			currentLocation = lastTransfer["receiver_id"].(string)
		}
	}
	if currentLocation == "" {
		currentLocation = batchInfo.HatcheryLocation
	}

	// Create the complete response object	// Get server base URL from environment or use a default
	serverHost := os.Getenv("SERVER_HOST")
	serverPort := os.Getenv("SERVER_PORT")
	baseURL := fmt.Sprintf("http://%s:%s", serverHost, serverPort)
	if serverHost == "" || serverPort == "" {
		baseURL = os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8080"
		}
	}
	var response map[string]interface{}
	
	if simplified {
		// Create a simplified response with only essential data
		response = map[string]interface{}{
			"batch_id": batchInfo.ID,
			"species": batchInfo.Species,
			"status": batchInfo.Status,
			"origin": batchInfo.HatcheryName,
			"verification_url": fmt.Sprintf("%s/api/v1/batches/%d/verify", baseURL, batchID),
		}
		
		// Add current location if available
		if currentLocation != "" {
			response["location"] = currentLocation
		}
	} else {
		// Create the complete response object
		response = map[string]interface{}{
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
			"verification_url": fmt.Sprintf("%s/api/v1/batches/%d/verify", baseURL, batchID),
		}
		
		// Since the NFT columns don't exist in the database, we'll set default NFT information
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
		fmt.Printf("Error marshaling QR data: %v\n", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR data")
	}
	
	// Check data size and handle accordingly
	dataSize := len(jsonData)
	fmt.Printf("QR code data size: %d bytes\n", dataSize)
	
	// Force simplified mode if data is extremely large
	if dataSize > 2000 && !simplified {
		fmt.Printf("Warning: QR code data is too large (%d bytes). Automatically switching to simplified mode.\n", dataSize)
		// Create a simplified version instead of full data
		simplifiedData := map[string]interface{}{
			"batch_id":         batchInfo.ID,
			"species":          batchInfo.Species,
			"status":           batchInfo.Status,
			"origin":           batchInfo.HatcheryName,
			"current_location": currentLocation,
			"verification_url": fmt.Sprintf("%s/api/v1/batches/%d/verify", baseURL, batchID),
		}
		
		jsonData, err = json.Marshal(simplifiedData)
		if err != nil {
			fmt.Printf("Error marshaling simplified QR data: %v\n", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR data")
		}
		fmt.Printf("Simplified QR code data size: %d bytes\n", len(jsonData))
	} else if dataSize > 1000 && !simplified {
		fmt.Printf("Warning: QR code data is large (%d bytes). Consider using simplified=true for better scanning.\n", dataSize)
	}
	
	// Select appropriate QR level based on data size
	var qrLevel qrcode.RecoveryLevel
	if len(jsonData) < 500 {
		qrLevel = qrcode.Low // Low level for small data
	} else if len(jsonData) < 1000 {
		qrLevel = qrcode.Medium // Medium level for moderate data
	} else {
		qrLevel = qrcode.Highest // Highest error correction for complex data
	}
		fmt.Printf("Using QR error correction level: %v\n", qrLevel)
	
	// Try to limit data size if it's extremely large
	if len(jsonData) > 3000 {
		jsonData = limitJSONSize(jsonData, 2500)
		fmt.Printf("Trimmed QR data to %d bytes\n", len(jsonData))
	}
	
	// Generate QR code with safety checks
	qr, err := qrcode.New(string(jsonData), qrLevel)
	if err != nil {
		fmt.Printf("Error generating QR code: %v (data size: %d bytes)\n", err, len(jsonData))
		// Try with a different error correction level as fallback
		qr, err = qrcode.New(string(jsonData), qrcode.Low)
		if err != nil {
			fmt.Printf("First fallback QR generation failed: %v\n", err)
			
			// Final fallback - bare minimum QR code
			minimalData := fmt.Sprintf("{\"id\":%d,\"url\":\"%s/api/v1/batches/%d\"}", 
				batchID, baseURL, batchID)
			
			qr, err = qrcode.New(minimalData, qrcode.Low)
			if err != nil {
				fmt.Printf("Final fallback QR generation failed: %v\n", err)
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR code")
			}
		}
	}
	
	// Set QR code options
	qr.DisableBorder = false
	
	// Create PNG image with requested size
	png, err := qr.PNG(size)
	if err != nil {
		fmt.Printf("Error generating QR PNG: %v (size: %d)\n", err, size)
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
			  EXISTS (SELECT 1 FROM environment_data WHERE id = related_id AND related_table = 'environment_data' AND batch_id = $1)
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
			// Get IPFS gateway URL from environment or use default
			gatewayURL := os.Getenv("IPFS_GATEWAY_URL")
			if gatewayURL == "" {
				gatewayURL = "https://ipfs.io"
			}
			
			records = append(records, map[string]interface{}{
				"id":            id,
				"related_table": relatedTable,
				"related_id":    relatedID,
				"tx_id":         txID,
				"metadata_hash": metadataHash,
				"created_at":    createdAt.Format(time.RFC3339),
				"ipfs_url":      fmt.Sprintf("%s/ipfs/%s", gatewayURL, metadataHash),
			})
		}
	}

	return records, nil
}

// limitJSONSize trims JSON data to stay under a maximum byte size while preserving valid JSON format
func limitJSONSize(originalJSON []byte, maxSize int) []byte {
	if len(originalJSON) <= maxSize {
		return originalJSON
	}
	
	// Parse original JSON
	var data map[string]interface{}
	if err := json.Unmarshal(originalJSON, &data); err != nil {
		// If we can't parse, just truncate with a basic approach (less ideal)
		if len(originalJSON) > maxSize-2 {
			return append(originalJSON[:maxSize-2], []byte("{}")...)
		}
		return originalJSON
	}
	
	// Remove large arrays first to reduce size
	for key, value := range data {
		// Check if value is an array/slice
		if arr, ok := value.([]interface{}); ok {
			if len(arr) > 3 {
				// Keep only first 3 elements
				data[key] = arr[:3]
			}
		}
		
		// Check if value is a nested map
		if nestedMap, ok := value.(map[string]interface{}); ok {
			// Remove less important nested fields
			for nestedKey := range nestedMap {
				if nestedKey != "id" && nestedKey != "status" && nestedKey != "type" {
					delete(nestedMap, nestedKey)
				}
			}
		}
	}
	
	// Try marshaling the reduced data
	reduced, err := json.Marshal(data)
	if err != nil || len(reduced) > maxSize {
		// Further simplification - keep only the most critical fields
		minimal := map[string]interface{}{
			"id": data["id"],
			"batch": map[string]interface{}{
				"id": data["batch"].(map[string]interface{})["id"],
			},
			"verification_url": data["verification_url"],
		}
		
		reduced, _ = json.Marshal(minimal)
		if len(reduced) > maxSize {
			// Last resort: Create a minimal JSON with just an ID
			minimal = map[string]interface{}{
				"id": data["id"],
			}
			reduced, _ = json.Marshal(minimal)
		}
	}
	
	return reduced
}
