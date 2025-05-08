package api

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/skip2/go-qrcode"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"github.com/LTPPPP/TracePost-larvaeChain/ipfs"
	"github.com/LTPPPP/TracePost-larvaeChain/models"
)

// CreateEventRequest represents a request to create a new event
type CreateEventRequest struct {
	BatchID   int                    `json:"batch_id"`
	EventType string                 `json:"event_type"`
	Location  string                 `json:"location"`
	ActorID   int                    `json:"actor_id"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// RecordEnvironmentDataRequest represents a request to record environment data
type RecordEnvironmentDataRequest struct {
	BatchID          int     `json:"batch_id"`
	Temperature      float64 `json:"temperature"`
	PH               float64 `json:"ph"`
	Salinity         float64 `json:"salinity"`
	DissolvedOxygen  float64 `json:"dissolved_oxygen"`
}

// UploadDocumentRequest represents a request to upload a document
type UploadDocumentRequest struct {
	BatchID   int    `form:"batch_id"`
	DocType   string `form:"doc_type"`
	UploadedBy int    `form:"uploaded_by"`
}

// TraceByQRCodeResponse represents the response for QR code tracing
type TraceByQRCodeResponse struct {
	Batch           models.BatchWithHatchery  `json:"batch"`
	Events          []models.EventWithActor   `json:"events"`
	Documents       []models.Document         `json:"documents"`
	EnvironmentData []models.EnvironmentData  `json:"environment_data"`
	LogisticsChain  []models.LogisticsEvent   `json:"logistics_chain"`
	BlockchainInfo  []models.BlockchainRecord `json:"blockchain_info"`
}

// CreateEvent creates a new event for a batch
// @Summary Create a new event
// @Description Create a new event for a shrimp larvae batch
// @Tags events
// @Accept json
// @Produce json
// @Param request body CreateEventRequest true "Event creation details"
// @Success 201 {object} SuccessResponse{data=models.Event}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /events [post]
func CreateEvent(c *fiber.Ctx) error {
	// Parse request body
	var req CreateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate input
	if req.BatchID <= 0 || req.EventType == "" || req.Location == "" || req.ActorID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID, event type, location, and actor ID are required")
	}

	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", req.BatchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Check if actor exists
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM account WHERE id = $1 AND is_active = true)", req.ActorID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Actor not found")
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(req.Metadata)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to serialize metadata")
	}
	var metadataJSONB models.JSONB
	err = json.Unmarshal(metadataJSON, &metadataJSONB)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to convert metadata to JSONB")
	}

	// Record event on blockchain
	txID, err := blockchainClient.RecordEvent(
		strconv.Itoa(req.BatchID),
		req.EventType,
		req.Location,
		strconv.Itoa(req.ActorID),
		req.Metadata,
	)
	if err != nil {
		// Log error but continue - blockchain is secondary to database
		fmt.Printf("Warning: Failed to record event on blockchain: %v\n", err)
	}

	// Insert event into database
	query := `
		INSERT INTO event (batch_id, event_type, actor_id, location, timestamp, metadata, updated_at, is_active)
		VALUES ($1, $2, $3, $4, NOW(), $5, NOW(), true)
		RETURNING id, timestamp
	`
	var event models.Event
	event.BatchID = req.BatchID
	event.EventType = req.EventType
	event.ActorID = req.ActorID
	event.Location = req.Location
	event.Metadata = metadataJSONB
	event.IsActive = true

	err = db.DB.QueryRow(
		query,
		event.BatchID,
		event.EventType,
		event.ActorID,
		event.Location,
		event.Metadata,
	).Scan(&event.ID, &event.Timestamp)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save event to database")
	}

	// Record blockchain transaction
	if txID != "" {
		// Generate metadata hash
		metadataForHash := map[string]interface{}{
			"event_id":   event.ID,
			"batch_id":   req.BatchID,
			"event_type": req.EventType,
			"location":   req.Location,
			"actor_id":   req.ActorID,
			"metadata":   req.Metadata,
			"timestamp":  event.Timestamp,
		}
		metadataHash, err := blockchainClient.HashData(metadataForHash)
		if err != nil {
			fmt.Printf("Warning: Failed to generate metadata hash: %v\n", err)
		}

		// Save blockchain record
		_, err = db.DB.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "event", event.ID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}

	// If event type is 'status_change', update batch status
	if event.EventType == "status_change" {
		// Get the new status from the event metadata
		newStatus, ok := req.Metadata["new_status"].(string)
		if ok && newStatus != "" {
			// Update batch status in database
			_, err = db.DB.Exec(
				"UPDATE batch SET status = $1, updated_at = NOW() WHERE id = $2",
				newStatus,
				event.BatchID,
			)
			if err != nil {
				// Log error but don't fail the request
				fmt.Printf("Warning: Failed to update batch status: %v\n", err)
			}
		}
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Event created successfully",
		Data:    event,
	})
}

// RecordEnvironmentData records environment data for a batch
// @Summary Record environment data
// @Description Record environment data for a shrimp larvae batch
// @Tags environment
// @Accept json
// @Produce json
// @Param request body RecordEnvironmentDataRequest true "Environment data details"
// @Success 201 {object} SuccessResponse{data=models.EnvironmentData}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /environment [post]
func RecordEnvironmentData(c *fiber.Ctx) error {
	// Parse request body
	var req RecordEnvironmentDataRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate input
	if req.BatchID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}

	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", req.BatchID).Scan(&exists)
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

	// Record environment data on blockchain
	txID, err := blockchainClient.RecordEnvironmentData(
		strconv.Itoa(req.BatchID),
		req.Temperature,
		req.PH,
		req.Salinity,
		req.DissolvedOxygen,
		nil, // No other params in new schema
	)
	if err != nil {
		// Log error but continue - blockchain is secondary to database
		fmt.Printf("Warning: Failed to record environment data on blockchain: %v\n", err)
	}

	// Insert environment data into database
	query := `
		INSERT INTO environment (batch_id, temperature, pH, salinity, dissolved_oxygen, timestamp, updated_at, is_active)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW(), true)
		RETURNING id, timestamp
	`
	var envData models.EnvironmentData
	envData.BatchID = req.BatchID
	envData.Temperature = req.Temperature
	envData.PH = req.PH
	envData.Salinity = req.Salinity
	envData.DissolvedOxygen = req.DissolvedOxygen
	envData.IsActive = true

	err = db.DB.QueryRow(
		query,
		envData.BatchID,
		envData.Temperature,
		envData.PH,
		envData.Salinity,
		envData.DissolvedOxygen,
	).Scan(&envData.ID, &envData.Timestamp)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save environment data to database")
	}

	// Record blockchain transaction
	if txID != "" {
		// Generate metadata hash
		metadataForHash := map[string]interface{}{
			"environment_id":   envData.ID,
			"batch_id":         req.BatchID,
			"temperature":      req.Temperature,
			"ph":               req.PH,
			"salinity":         req.Salinity,
			"dissolved_oxygen": req.DissolvedOxygen,
			"timestamp":        envData.Timestamp,
		}
		metadataHash, err := blockchainClient.HashData(metadataForHash)
		if err != nil {
			fmt.Printf("Warning: Failed to generate metadata hash: %v\n", err)
		}

		// Save blockchain record
		_, err = db.DB.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "environment", envData.ID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Environment data recorded successfully",
		Data:    envData,
	})
}

// UploadDocument uploads a document for a batch
// @Summary Upload a document
// @Description Upload a document for a shrimp larvae batch
// @Tags documents
// @Accept multipart/form-data
// @Produce json
// @Param batch_id formData int true "Batch ID"
// @Param doc_type formData string true "Document type"
// @Param uploaded_by formData int true "Uploader ID"
// @Param file formData file true "Document file"
// @Success 201 {object} SuccessResponse{data=models.Document}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /documents [post]
func UploadDocument(c *fiber.Ctx) error {
	// Parse form
	form, err := c.MultipartForm()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid multipart form")
	}

	// Get form values
	batchIDs := form.Value["batch_id"]
	docTypes := form.Value["doc_type"]
	uploaderIDs := form.Value["uploaded_by"]

	// Validate input
	if len(batchIDs) == 0 || len(docTypes) == 0 || len(uploaderIDs) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID, document type, and uploader ID are required")
	}
	
	batchIDStr := batchIDs[0]
	docType := docTypes[0]
	uploaderIDStr := uploaderIDs[0]
	
	// Convert string IDs to integers
	batchID, err := strconv.Atoi(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
	}
	
	uploaderID, err := strconv.Atoi(uploaderIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid uploader ID format")
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

	// Check if uploader exists
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM account WHERE id = $1 AND is_active = true)", uploaderID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Uploader not found")
	}

	// Get file
	files := form.File["file"]
	if len(files) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "File is required")
	}
	file := files[0]

	// Open file
	fileHandle, err := file.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to open file")
	}
	defer fileHandle.Close()

	// Initialize IPFS client
	ipfsClient := ipfs.NewIPFSClient("http://localhost:5001")

	// Upload file to IPFS
	ipfsHash, err := ipfsClient.UploadFile(fileHandle)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to upload file to IPFS")
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Record document on blockchain
	txID, err := blockchainClient.RecordDocument(strconv.Itoa(batchID), docType, ipfsHash, strconv.Itoa(uploaderID))
	if err != nil {
		// Log error but continue - blockchain is secondary to database
		fmt.Printf("Warning: Failed to record document on blockchain: %v\n", err)
	}

	// Insert document into database
	query := `
		INSERT INTO document (batch_id, doc_type, ipfs_hash, uploaded_by, uploaded_at, updated_at, is_active)
		VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		RETURNING id, uploaded_at
	`
	var doc models.Document
	doc.BatchID = batchID
	doc.DocType = docType
	doc.IPFSHash = ipfsHash
	doc.UploadedBy = uploaderID
	doc.IsActive = true

	err = db.DB.QueryRow(
		query,
		doc.BatchID,
		doc.DocType,
		doc.IPFSHash,
		doc.UploadedBy,
	).Scan(&doc.ID, &doc.UploadedAt)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save document to database")
	}

	// Record blockchain transaction
	if txID != "" {
		// Generate metadata hash
		metadataForHash := map[string]interface{}{
			"document_id": doc.ID,
			"batch_id":    batchID,
			"doc_type":    docType,
			"ipfs_hash":   ipfsHash,
			"uploaded_by": uploaderID,
			"uploaded_at": doc.UploadedAt,
		}
		metadataHash, err := blockchainClient.HashData(metadataForHash)
		if err != nil {
			fmt.Printf("Warning: Failed to generate metadata hash: %v\n", err)
		}

		// Save blockchain record
		_, err = db.DB.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "document", doc.ID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Document uploaded successfully",
		Data:    doc,
	})
}

// GetDocumentByID returns a document by ID
// @Summary Get document by ID
// @Description Retrieve a document by its ID
// @Tags documents
// @Accept json
// @Produce json
// @Param documentId path string true "Document ID"
// @Success 200 {object} SuccessResponse{data=models.Document}
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /documents/{documentId} [get]
func GetDocumentByID(c *fiber.Ctx) error {
	// Get document ID from params
	documentIDStr := c.Params("documentId")
	if documentIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Document ID is required")
	}
	
	documentID, err := strconv.Atoi(documentIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid document ID format")
	}

	// Query document from database
	var doc models.Document
	query := `
		SELECT id, batch_id, doc_type, ipfs_hash, uploaded_by, uploaded_at, updated_at, is_active
		FROM document
		WHERE id = $1 AND is_active = true
	`
	err = db.DB.QueryRow(query, documentID).Scan(
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
		if err.Error() == "sql: no rows in result set" {
			return fiber.NewError(fiber.StatusNotFound, "Document not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Document retrieved successfully",
		Data:    doc,
	})
}

// TraceByQRCode traces a batch by QR code
// @Summary Trace by QR code
// @Description Trace a shrimp larvae batch by QR code, including complete logistics tracking
// @Tags qr
// @Accept json
// @Produce json
// @Param batchID path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=TraceByQRCodeResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /qr/{batchID} [get]
func TraceByQRCode(c *fiber.Ctx) error {
    // Get batchID from params
    batchIDStr := c.Params("batchID")
    if batchIDStr == "" {
        return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
    }
    
    // Convert to integer
    batchID, err := strconv.Atoi(batchIDStr)
    if err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
    }

    // Check if batch exists in database
    var exists bool
    err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchID).Scan(&exists)
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Database error")
    }
    if !exists {
        return fiber.NewError(fiber.StatusNotFound, "Batch not found")
    }

    // Get batch details with hatchery information
    var batchWithHatchery models.BatchWithHatchery
    query := `
        SELECT b.id, b.hatchery_id, b.species, b.quantity, b.status, b.created_at, b.updated_at, b.is_active,
               h.name, h.location, h.contact
        FROM batch b
        JOIN hatchery h ON b.hatchery_id = h.id
        WHERE b.id = $1 AND b.is_active = true
    `
    err = db.DB.QueryRow(query, batchID).Scan(
        &batchWithHatchery.ID,
        &batchWithHatchery.HatcheryID,
        &batchWithHatchery.Species,
        &batchWithHatchery.Quantity,
        &batchWithHatchery.Status,
        &batchWithHatchery.CreatedAt,
        &batchWithHatchery.UpdatedAt,
        &batchWithHatchery.IsActive,
        &batchWithHatchery.HatcheryName,
        &batchWithHatchery.HatcheryLocation,
        &batchWithHatchery.HatcheryContact,
    )
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve batch data")
    }

    // Get events with actor information
    rows, err := db.DB.Query(`
        SELECT e.id, e.batch_id, e.event_type, e.actor_id, e.location, e.timestamp, e.metadata, e.updated_at, e.is_active,
               a.username, a.role, a.email
        FROM event e
        JOIN account a ON e.actor_id = a.id
        WHERE e.batch_id = $1 AND e.is_active = true
        ORDER BY e.timestamp DESC
    `, batchID)
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve events")
    }
    defer rows.Close()

    var eventsWithActor []models.EventWithActor
    for rows.Next() {
        var event models.EventWithActor
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
            &event.ActorName,
            &event.ActorRole,
            &event.ActorEmail,
        )
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse event data")
        }
        eventsWithActor = append(eventsWithActor, event)
    }

    // Get documents
    docRows, err := db.DB.Query(`
        SELECT id, batch_id, doc_type, ipfs_hash, uploaded_by, uploaded_at, updated_at, is_active
        FROM document
        WHERE batch_id = $1 AND is_active = true
        ORDER BY uploaded_at DESC
    `, batchID)
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve documents")
    }
    defer docRows.Close()

    var documents []models.Document
    for docRows.Next() {
        var doc models.Document
        err := docRows.Scan(
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

    // Get environment data
    envRows, err := db.DB.Query(`
        SELECT id, batch_id, temperature, pH, salinity, dissolved_oxygen, timestamp, updated_at, is_active
        FROM environment
        WHERE batch_id = $1 AND is_active = true
        ORDER BY timestamp DESC
    `, batchID)
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve environment data")
    }
    defer envRows.Close()

    var envDataList []models.EnvironmentData
    for envRows.Next() {
        var envData models.EnvironmentData
        err := envRows.Scan(
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

    // Extract logistics chain from events
    // This builds a chronological chain of transfer/transport events
    var logisticsChain []models.LogisticsEvent
    for _, event := range eventsWithActor {
        // Only include logistics-related events
        if event.EventType == "transfer" || event.EventType == "transport" || 
           event.EventType == "shipping" || event.EventType == "receiving" {
            
            // Extract logistics data from event metadata
            var fromLocation, toLocation, transporterName string
            var departureTime, arrivalTime time.Time
            var status string
            
            // Parse metadata from JSON
            var metadata map[string]interface{}
            if len(event.Metadata) > 0 {
                err := json.Unmarshal(event.Metadata, &metadata)
                if err == nil {
                    // Extract logistics fields from metadata if they exist
                    if val, ok := metadata["from_location"].(string); ok {
                        fromLocation = val
                    } else {
                        fromLocation = event.Location // fallback to event location
                    }
                    
                    if val, ok := metadata["to_location"].(string); ok {
                        toLocation = val
                    }
                    
                    if val, ok := metadata["transporter_name"].(string); ok {
                        transporterName = val
                    } else if event.ActorRole == "transporter" {
                        transporterName = event.ActorName // fallback to actor name if role is transporter
                    }
                    
                    if val, ok := metadata["departure_time"].(string); ok {
                        departureTime, _ = time.Parse(time.RFC3339, val)
                    }
                    
                    if val, ok := metadata["arrival_time"].(string); ok {
                        arrivalTime, _ = time.Parse(time.RFC3339, val)
                    }
                    
                    if val, ok := metadata["status"].(string); ok {
                        status = val
                    } else {
                        status = "completed" // default status
                    }
                }
            }
            
            // Create logistics event
            logisticsEvent := models.LogisticsEvent{
                ID:              event.ID,
                BatchID:         event.BatchID,
                EventType:       event.EventType,
                FromLocation:    fromLocation,
                ToLocation:      toLocation,
                TransporterName: transporterName,
                DepartureTime:   departureTime,
                ArrivalTime:     arrivalTime,
                Status:          status,
                Metadata:        event.Metadata,
                Timestamp:       event.Timestamp,
            }
            
            logisticsChain = append(logisticsChain, logisticsEvent)
        }
    }
    
    // Sort logistics chain by timestamp (oldest first)
    sort.Slice(logisticsChain, func(i, j int) bool {
        return logisticsChain[i].Timestamp.Before(logisticsChain[j].Timestamp)
    })

    // Get blockchain records for this batch
    blockchainRows, err := db.DB.Query(`
        SELECT id, related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active
        FROM blockchain_record
        WHERE (related_table = 'batch' AND related_id = $1) OR 
              EXISTS (SELECT 1 FROM event WHERE id = related_id AND related_table = 'event' AND batch_id = $1) OR
              EXISTS (SELECT 1 FROM document WHERE id = related_id AND related_table = 'document' AND batch_id = $1) OR
              EXISTS (SELECT 1 FROM environment WHERE id = related_id AND related_table = 'environment' AND batch_id = $1)
        ORDER BY created_at DESC
    `, batchID)
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve blockchain records")
    }
    defer blockchainRows.Close()

    var blockchainRecords []models.BlockchainRecord
    for blockchainRows.Next() {
        var record models.BlockchainRecord
        err := blockchainRows.Scan(
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
            return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse blockchain record")
        }
        blockchainRecords = append(blockchainRecords, record)
    }

    // Create response with all data
    response := TraceByQRCodeResponse{
        Batch:           batchWithHatchery,
        Events:          eventsWithActor,
        Documents:       documents,
        EnvironmentData: envDataList,
        LogisticsChain:  logisticsChain,
        BlockchainInfo:  blockchainRecords,
    }

    // Return success response
    return c.JSON(SuccessResponse{
        Success: true,
        Message: "Batch traced successfully",
        Data:    response,
    })
}

// GetCurrentUser returns the current user
// @Summary Get current user
// @Description Retrieve the current user's information
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse{data=models.User}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/me [get]
func GetCurrentUser(c *fiber.Ctx) error {
	// This is a placeholder - in a real application, you would get the user ID from the JWT token
	// For now, we'll return a mock user
	user := models.User{
		ID:           1,
		Username:     "john.doe",
		Email:        "john.doe@example.com",
		PasswordHash: "", // Don't expose this
		Role:         "admin",
		CompanyID:    1,
		LastLogin:    time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    user,
	})
}

// UpdateCurrentUser updates the current user
// @Summary Update current user
// @Description Update the current user's information
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/me [put]
func UpdateCurrentUser(c *fiber.Ctx) error {
	// This is a placeholder - in a real application, you would update the user in the database
	// For now, we'll just return a success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User updated successfully",
	})
}

// ChangePassword changes the current user's password
// @Summary Change password
// @Description Change the current user's password
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/me/password [put]
func ChangePassword(c *fiber.Ctx) error {
	// This is a placeholder - in a real application, you would update the user's password in the database
	// For now, we'll just return a success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Password changed successfully",
	})
}

// GenerateGatewayQRCode generates a QR code for a batch with a public gateway URL
// @Summary Generate gateway QR code
// @Description Generate a QR code for a batch with a public IPFS gateway URL
// @Tags qr
// @Accept json
// @Produce image/png
// @Param batchId path string true "Batch ID"
// @Param gateway query string false "IPFS gateway to use (default: ipfs.io)"
// @Success 200 {file} byte[] "QR code as PNG image"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /qr/gateway/{batchId} [get]
func GenerateGatewayQRCode(c *fiber.Ctx) error {
	// Get batch ID from params
	batchIDStr := c.Params("batchId")
	if batchIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	batchID, err := strconv.Atoi(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
	}

	// Check if batch exists in database
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Get gateway from query parameter, default to ipfs.io
	gateway := c.Query("gateway", "ipfs.io")
	
	// Create gateway URL format
	qrData := fmt.Sprintf("https://%s/ipfs/%d", gateway, batchID)

	// Generate QR code
	qrCode, err := qrcode.Encode(qrData, qrcode.Medium, 256)
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