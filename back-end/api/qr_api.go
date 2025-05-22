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

// ConfigQRCode generates a QR code containing configuration information for a batch
// @Summary Configuration QR Code
// @Description Generate a QR code with configuration information about a batch
// @Tags qr
// @Accept json
// @Produce image/png,application/json
// @Param batchId path string true "Batch ID"
// @Param format query string false "Format: 'png' or 'json' (default: 'png')"
// @Param size query int false "QR code size in pixels (default: 512)"
// @Success 200 {file} byte[] "QR code image or JSON data"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /qr/config/{batchId} [get]
func ConfigQRCode(c *fiber.Ctx) error {
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
	
	// Check if batch exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// 1. Get batch details with configuration information
	var batchInfo struct {
		ID               int       `json:"id"`
		HatcheryID       string    `json:"hatchery_id"`
		HatcheryName     string    `json:"hatchery_name"`
		Species          string    `json:"species"`
		Quantity         int       `json:"quantity"`
		Status           string    `json:"status"`
		CreatedAt        time.Time `json:"created_at"`
	}
	err = db.DB.QueryRow(`
		SELECT b.id, b.hatchery_id, h.name, b.species, b.quantity, b.status, 
		       b.created_at
		FROM batch b
		JOIN hatchery h ON b.hatchery_id = h.id
		WHERE b.id = $1 AND b.is_active = true
	`, batchID).Scan(
		&batchInfo.ID,
		&batchInfo.HatcheryID,
		&batchInfo.HatcheryName,
		&batchInfo.Species,
		&batchInfo.Quantity,
		&batchInfo.Status,
		&batchInfo.CreatedAt,
	)
	if err != nil {
		fmt.Printf("Database error retrieving batch %d: %v\n", batchID, err)
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to retrieve batch data: %v", err))
	}

	// 2. Get environment data for configuration details
	rows, err := db.DB.Query(`
		SELECT id, temperature, ph, salinity, density, age, timestamp
		FROM environment_data
		WHERE batch_id = $1 AND is_active = true
		ORDER BY timestamp DESC
		LIMIT 1
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve environment data")
	}
	defer rows.Close()

	var environmentData map[string]interface{}
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
			environmentData = map[string]interface{}{
				"id":          id,
				"temperature": temperature,
				"ph":          ph,
				"salinity":    salinity,
				"density":     density,
				"age":         age,
				"timestamp":   timestamp.Format(time.RFC3339),
			}
		}
	}

	// Get server base URL from environment or use a default
	serverHost := os.Getenv("SERVER_HOST")
	serverPort := os.Getenv("SERVER_PORT")
	baseURL := fmt.Sprintf("http://%s:%s", serverHost, serverPort)
	if serverHost == "" || serverPort == "" {
		baseURL = os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8080"
		}
	}

	// Create the configuration response object
	configResponse := map[string]interface{}{
		"batch_id":     batchInfo.ID,
		"species":      batchInfo.Species,
		"origin":       batchInfo.HatcheryName,
		"quantity":     batchInfo.Quantity,
		"created_at":   batchInfo.CreatedAt.Format(time.RFC3339),
		"config":       environmentData,
		"verification": fmt.Sprintf("%s/api/v1/batches/%d/verify", baseURL, batchID),
	}

	// If JSON format is requested, return data directly
	if format == "json" {
		return c.JSON(configResponse)
	}

	// For PNG format, generate QR code
	jsonData, err := json.Marshal(configResponse)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR data")
	}
	
	// Generate QR code with appropriate error correction level
	var qrLevel qrcode.RecoveryLevel
	dataSize := len(jsonData)
	
	if dataSize < 500 {
		qrLevel = qrcode.Low
	} else if dataSize < 1000 {
		qrLevel = qrcode.Medium
	} else {
		qrLevel = qrcode.High
	}
	
	qr, err := qrcode.New(string(jsonData), qrLevel)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR code")
	}
	
	// Set QR code options
	qr.DisableBorder = false
	
	// Create PNG image with requested size
	png, err := qr.PNG(size)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR image")
	}
	
	// Set content type and return image
	c.Set("Content-Type", "image/png")
	return c.Send(png)
}

// BlockchainQRCode generates a QR code containing blockchain traceability information
// @Summary Blockchain traceability QR Code
// @Description Generate a QR code with blockchain traceability information about a batch
// @Tags qr
// @Accept json
// @Produce image/png,application/json
// @Param batchId path string true "Batch ID"
// @Param format query string false "Format: 'png' or 'json' (default: 'png')"
// @Param size query int false "QR code size in pixels (default: 512)"
// @Success 200 {file} byte[] "QR code image or JSON data"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /qr/blockchain/{batchId} [get]
func BlockchainQRCode(c *fiber.Ctx) error {
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
	
	// Check if batch exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// 1. Get batch basic details
	var batchInfo struct {
		ID           int       `json:"id"`
		Species      string    `json:"species"`
		Status       string    `json:"status"`
		HatcheryName string    `json:"hatchery_name"`
		CreatedAt    time.Time `json:"created_at"`
	}
	
	err = db.DB.QueryRow(`
		SELECT b.id, b.species, b.status, h.name, b.created_at
		FROM batch b
		JOIN hatchery h ON b.hatchery_id = h.id
		WHERE b.id = $1 AND b.is_active = true
	`, batchID).Scan(
		&batchInfo.ID,
		&batchInfo.Species,
		&batchInfo.Status,
		&batchInfo.HatcheryName,
		&batchInfo.CreatedAt,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve batch data")
	}

	// 2. Get blockchain verification records
	blockchainRecords, err := getBlockchainRecordsForBatch(batchID)
	if err != nil {
		fmt.Printf("Warning: Failed to retrieve blockchain records: %v\n", err)
	}

	// Determine current location from transfers if available
	var currentLocation string
	err = db.DB.QueryRow(`
		SELECT CASE 
			WHEN t.status = 'completed' THEN t.receiver_id
			ELSE c.location
		END as current_location
		FROM batch b
		JOIN hatchery h ON b.hatchery_id = h.id
		JOIN company c ON h.company_id = c.id
		LEFT JOIN (
			SELECT * FROM shipment_transfer 
			WHERE batch_id = $1 AND is_active = true 
			ORDER BY transfer_time DESC LIMIT 1
		) t ON true
		WHERE b.id = $1
	`, batchID).Scan(&currentLocation)
	
	if err != nil {
		currentLocation = "Unknown"
	}

	// Get server base URL from environment or use a default
	serverHost := os.Getenv("SERVER_HOST")
	serverPort := os.Getenv("SERVER_PORT")
	baseURL := fmt.Sprintf("http://%s:%s", serverHost, serverPort)
	if serverHost == "" || serverPort == "" {
		baseURL = os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8080"
		}
	}

	// Create blockchain traceability response
	blockchainResponse := map[string]interface{}{
		"batch_id":     batchInfo.ID,
		"species":      batchInfo.Species,
		"status":       batchInfo.Status,
		"origin":       batchInfo.HatcheryName,
		"location":     currentLocation,
		"created_at":   batchInfo.CreatedAt.Format(time.RFC3339),
		"blockchain":   blockchainRecords,
		"verification": fmt.Sprintf("%s/api/v1/batches/%d/verify", baseURL, batchID),
	}

	// If JSON format is requested, return data directly
	if format == "json" {
		return c.JSON(blockchainResponse)
	}

	// For PNG format, generate QR code
	jsonData, err := json.Marshal(blockchainResponse)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR data")
	}
	
	// Check data size and adjust if necessary
	dataSize := len(jsonData)
	if dataSize > 2000 {
		// Simplify the data to reduce size
		simplifiedData := map[string]interface{}{
			"batch_id":     batchInfo.ID,
			"species":      batchInfo.Species,
			"status":       batchInfo.Status,
			"origin":       batchInfo.HatcheryName,
			"location":     currentLocation,
			"verification": fmt.Sprintf("%s/api/v1/batches/%d/verify", baseURL, batchID),
		}
		
		if len(blockchainRecords) > 0 {
			simplifiedData["latest_tx"] = blockchainRecords[0]["tx_id"]
		}
		
		jsonData, err = json.Marshal(simplifiedData)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR data")
		}
	}
	
	// Generate QR code with appropriate error correction level
	var qrLevel qrcode.RecoveryLevel
	if len(jsonData) < 500 {
		qrLevel = qrcode.Low
	} else if len(jsonData) < 1000 {
		qrLevel = qrcode.Medium
	} else {
		qrLevel = qrcode.High
	}
	
	qr, err := qrcode.New(string(jsonData), qrLevel)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR code")
	}
	
	// Set QR code options
	qr.DisableBorder = false
	
	// Create PNG image with requested size
	png, err := qr.PNG(size)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR image")
	}
	
	// Set content type and return image
	c.Set("Content-Type", "image/png")
	return c.Send(png)
}

// DocumentQRCode generates a QR code containing IPFS document links for a batch
// @Summary Document IPFS link QR Code
// @Description Generate a QR code with document IPFS links for a batch
// @Tags qr
// @Accept json
// @Produce image/png,application/json
// @Param batchId path string true "Batch ID"
// @Param format query string false "Format: 'png' or 'json' (default: 'png')"
// @Param size query int false "QR code size in pixels (default: 512)"
// @Success 200 {file} byte[] "QR code image or JSON data"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /qr/document/{batchId} [get]
func DocumentQRCode(c *fiber.Ctx) error {
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
	
	// Check if batch exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Get documents for this batch to find the most recent IPFS hash
	rows, err := db.DB.Query(`
		SELECT ipfs_hash
		FROM document
		WHERE batch_id = $1 AND is_active = true
		ORDER BY uploaded_at DESC
		LIMIT 1
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve documents")
	}
	defer rows.Close()

	// Get a single IPFS hash for this batch (if documents exist)
	var batchIpfsHash string
	if rows.Next() {
		err := rows.Scan(&batchIpfsHash)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve IPFS hash")
		}
	}
	
	if batchIpfsHash == "" {
		return fiber.NewError(fiber.StatusNotFound, "No documents found for this batch")
	}
	
	// Use Pinata gateway URL if available, otherwise use default
	gatewayURL := os.Getenv("PINATA_GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = os.Getenv("IPFS_GATEWAY_URL")
	}
	if gatewayURL == "" {
		gatewayURL = "https://gateway.pinata.cloud"
	}
	
	// Create the IPFS URI
	ipfsUri := fmt.Sprintf("%s/ipfs/%s", gatewayURL, batchIpfsHash)
	
	// If JSON format is requested, return simple response with just the URI
	if format == "json" {
		return c.JSON(map[string]string{"ipfs_uri": ipfsUri})
	}

	// For PNG format, generate QR code with just the IPFS URI
	qr, err := qrcode.New(ipfsUri, qrcode.Medium)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR code")
	}
	
	// Set QR code options
	qr.DisableBorder = false
	
	// Create PNG image with requested size
	png, err := qr.PNG(size)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR image")
	}
	
	// Set content type and return image
	c.Set("Content-Type", "image/png")
	return c.Send(png)
}
