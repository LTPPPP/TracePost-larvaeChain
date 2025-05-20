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

// QRCodeDiagnostics provides diagnostic information about QR code generation
// @Summary QR Code Diagnostics
// @Description Get diagnostic information about QR code generation for troubleshooting scanning issues
// @Tags qr
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /qr/diagnostics/{batchId} [get]
func QRCodeDiagnostics(c *fiber.Ctx) error {
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
		return fiber.NewError(fiber.StatusInternalServerError, "Database error checking batch existence")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}
	
	// Get server base URL
	serverHost := os.Getenv("SERVER_HOST")
	serverPort := os.Getenv("SERVER_PORT")
	baseURL := fmt.Sprintf("http://%s:%s", serverHost, serverPort)
	if serverHost == "" || serverPort == "" {
		baseURL = os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8080"
		}
	}
		// Test generating all three QR code types and measure data size
	// 1. Configuration QR
	configResponse, err := getConfigQRData(batchID, baseURL)
	if err != nil {
		fmt.Printf("Error getting configuration QR data: %v\n", err)
		configResponse = map[string]interface{}{
			"error": err.Error(),
		}
	}
	
	configJSON, err := json.Marshal(configResponse)
	configSize := len(configJSON)
	configScannable := configSize < 2000 // Rough estimate of max reliable QR data size
	
	// 2. Blockchain Traceability QR
	blockchainResponse, err := getBlockchainQRData(batchID, baseURL)
	if err != nil {
		fmt.Printf("Error getting blockchain QR data: %v\n", err)
		blockchainResponse = map[string]interface{}{
			"error": err.Error(),
		}
	}
	
	blockchainJSON, err := json.Marshal(blockchainResponse)
	blockchainSize := len(blockchainJSON)
	blockchainScannable := blockchainSize < 2500 // Blockchain QR should work up to around 2.5KB
	
	// 3. Document QR
	documentResponse, err := getDocumentQRData(batchID, baseURL)
	if err != nil {
		fmt.Printf("Error getting document QR data: %v\n", err)
		documentResponse = map[string]interface{}{
			"error": err.Error(),
		}
	}
	
	documentJSON, err := json.Marshal(documentResponse)
	documentSize := len(documentJSON)
	documentScannable := documentSize < 3000 // Document QR should be reliable
	
	// Get QR code libraries version
	qrLibraryVersion := "skip2/go-qrcode"
	
	// Run test with each recovery level and measure
	var levelResults []map[string]interface{}
	
	for _, level := range []qrcode.RecoveryLevel{qrcode.Low, qrcode.Medium, qrcode.High, qrcode.Highest} {
		var levelName string
		switch level {
		case qrcode.Low:
			levelName = "Low"
		case qrcode.Medium:
			levelName = "Medium"
		case qrcode.High:
			levelName = "High"
		case qrcode.Highest:
			levelName = "Highest"
		}
		
		// Create test string of increasing complexity
		testData := fmt.Sprintf(`{"id":%d,"test":"value","timestamp":"%s"}`, 
			batchID, time.Now().Format(time.RFC3339))
			
		// Try to create QR code
		_, err := qrcode.New(testData, level)
		
		levelResults = append(levelResults, map[string]interface{}{
			"level":     levelName,
			"test_data": testData,
			"success":   err == nil,
			"error":     func() interface{} {
				if err != nil {
					return err.Error()
				}
				return nil
			}(),
		})
	}
	// Build response with all diagnostic data - consolidated to 3 main QR APIs
	diagnosticResponse := map[string]interface{}{
		"batch_id": batchID,
		"server_info": map[string]interface{}{
			"base_url":   baseURL,
			"qr_library": qrLibraryVersion,
		},
		"qr_types": map[string]interface{}{
			"config_qr": map[string]interface{}{
				"description":      "Configuration information QR code",
				"data_size_bytes": configSize,
				"likely_scannable": configScannable,
				"url": fmt.Sprintf("%s/api/v1/qr/config/%d", baseURL, batchID),
				"example_data": map[string]interface{}{
					"batch_id": batchID,
					"config": map[string]interface{}{
						"species_name": "Sample Species",
						"farming_type": "Intensive",
						"feed_type": "Premium",
					},
				},
			},
			"blockchain_qr": map[string]interface{}{
				"description":      "Blockchain traceability QR code",
				"data_size_bytes": blockchainSize,
				"likely_scannable": blockchainScannable,
				"url": fmt.Sprintf("%s/api/v1/qr/blockchain/%d", baseURL, batchID),
				"example_data": map[string]interface{}{
					"batch_id": batchID,
					"blockchain_verification": map[string]interface{}{
						"tx_hash": "0x1234...5678",
						"verified": true,
						"timestamp": time.Now().Format(time.RFC3339),
					},
				},
			},
			"document_qr": map[string]interface{}{
				"description":      "Document IPFS link QR code",
				"data_size_bytes": documentSize,
				"likely_scannable": documentScannable,
				"url": fmt.Sprintf("%s/api/v1/qr/document/%d", baseURL, batchID),
				"example_data": map[string]interface{}{
					"batch_id": batchID,
					"document_url": "https://ipfs.io/ipfs/QmX5J3jvgQKmTJjz2h6brY4zsXmiNMRa3678MXCTexy1B",
				},
			},
		},
		"recovery_level_tests": levelResults,
		"recommendations": []string{
			func() string {
				if configScannable {
					return "Configuration QR code should work for detailed settings"
				}
				return "Configuration QR code contains too much data, consider simplifying configuration data"
			}(),
			func() string {
				if blockchainScannable {
					return "Blockchain traceability QR is recommended for supply chain verification"
				}
				return "Blockchain QR has complex data, consider using document QR with links to blockchain info"
			}(),
			func() string {
				if documentScannable {
					return "Document QR code links directly to IPFS documents and should work reliably"
				}
				return "Document linking may be unreliable, check IPFS gateway settings"
			}(),
			"For best scanning results, use size=800 parameter",
			"Make sure your device has a good camera and proper lighting for scanning",
			"The document QR is recommended for most end-user scanning scenarios",
		},
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "QR code diagnostics completed",
		Data:    diagnosticResponse,
	})
}

// Helper function to get configuration QR data
func getConfigQRData(batchID int, baseURL string) (map[string]interface{}, error) {
	// Get batch details with configuration information
	var batchInfo struct {
		ID           int       
		Species      string    
		Quantity     int       
		Status       string    
		CreatedAt    time.Time 
		HatcheryName string    
	}
	
	err := db.DB.QueryRow(`
		SELECT b.id, b.species, b.quantity, b.status, b.created_at, h.name
		FROM batch b
		JOIN hatchery h ON b.hatchery_id = h.id
		WHERE b.id = $1 AND b.is_active = true
	`, batchID).Scan(
		&batchInfo.ID,
		&batchInfo.Species,
		&batchInfo.Quantity,
		&batchInfo.Status,
		&batchInfo.CreatedAt,
		&batchInfo.HatcheryName,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve batch data: %v", err)
	}
	
	// Get environment data for configuration details
	var environmentData map[string]interface{}
	rows, err := db.DB.Query(`
		SELECT id, temperature, ph, salinity, density, age, timestamp
		FROM environment_data
		WHERE batch_id = $1 AND is_active = true
		ORDER BY timestamp DESC
		LIMIT 1
	`, batchID)
	
	if err == nil {
		defer rows.Close()
		
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
	}
	
	// Create configuration response
	response := map[string]interface{}{
		"batch_id":     batchInfo.ID,
		"species":      batchInfo.Species,
		"origin":       batchInfo.HatcheryName,
		"quantity":     batchInfo.Quantity,
		"created_at":   batchInfo.CreatedAt.Format(time.RFC3339),
		"config":       environmentData,
		"verification": fmt.Sprintf("%s/api/v1/batches/%d/verify", baseURL, batchID),
	}
	
	return response, nil
}

// Helper function to get blockchain QR data
func getBlockchainQRData(batchID int, baseURL string) (map[string]interface{}, error) {
	// Get batch basic details
	var batchInfo struct {
		ID           int       
		Species      string    
		Status       string    
		HatcheryName string    
		CreatedAt    time.Time
	}
	
	err := db.DB.QueryRow(`
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
		return nil, fmt.Errorf("failed to retrieve batch data: %v", err)
	}
	
	// Get blockchain verification records
	blockchainRecords, _ := getBlockchainRecordsForBatch(batchID)
	
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
	
	// Create blockchain traceability response
	response := map[string]interface{}{
		"batch_id":     batchInfo.ID,
		"species":      batchInfo.Species,
		"status":       batchInfo.Status,
		"origin":       batchInfo.HatcheryName,
		"location":     currentLocation,
		"created_at":   batchInfo.CreatedAt.Format(time.RFC3339),
		"blockchain":   blockchainRecords,
		"verification": fmt.Sprintf("%s/api/v1/batches/%d/verify", baseURL, batchID),
	}
	
	return response, nil
}

// Helper function to get document QR data
func getDocumentQRData(batchID int, baseURL string) (map[string]interface{}, error) {
	// Get basic batch info
	var batchInfo struct {
		ID      int    
		Species string 
	}
	
	err := db.DB.QueryRow(`
		SELECT id, species
		FROM batch
		WHERE id = $1 AND is_active = true
	`, batchID).Scan(&batchInfo.ID, &batchInfo.Species)
	
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve batch data: %v", err)
	}
	
	// Get documents for this batch
	var documents []map[string]interface{}
	rows, err := db.DB.Query(`
		SELECT id, doc_type, ipfs_hash, uploaded_by, uploaded_at
		FROM document
		WHERE batch_id = $1 AND is_active = true
		ORDER BY uploaded_at DESC
	`, batchID)
	
	if err == nil {
		defer rows.Close()
		
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
	}
	
	// Create document response
	response := map[string]interface{}{
		"batch_id":     batchInfo.ID,
		"species":      batchInfo.Species,
		"documents":    documents,
		"verification": fmt.Sprintf("%s/api/v1/batches/%d/verify", baseURL, batchID),
	}
	
	return response, nil
}
