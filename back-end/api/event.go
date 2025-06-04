package api

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"github.com/LTPPPP/TracePost-larvaeChain/models"
)

// UpdateEventRequest represents a request to update an event
type UpdateEventRequest struct {
	EventType string                 `json:"event_type"`
	Location  string                 `json:"location"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// GetAllEvents retrieves all event records
// @Summary Get all events
// @Description Retrieve all event records with optional filtering
// @Tags events
// @Accept json
// @Produce json
// @Param batch_id query int false "Filter by batch ID"
// @Param event_type query string false "Filter by event type"
// @Param limit query int false "Limit number of results (default: 50)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Success 200 {object} SuccessResponse{data=[]models.Event}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /events [get]
func GetAllEvents(c *fiber.Ctx) error {
	// Parse query parameters
	batchIDStr := c.Query("batch_id")
	eventType := c.Query("event_type")
	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100 // Max limit to prevent abuse
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}
	// Build query
	query := `
		SELECT 
			e.id, e.batch_id, e.event_type, e.location, 
			e.timestamp, e.updated_at, e.is_active, e.metadata,
			b.species, b.quantity, b.status,
			h.name AS hatchery_name,
			c.name AS company_name
		FROM event e
		INNER JOIN batch b ON e.batch_id = b.id
		INNER JOIN hatchery h ON b.hatchery_id = h.id
		INNER JOIN company c ON h.company_id = c.id
		WHERE e.is_active = true
	`
	
	args := []interface{}{}
	argIndex := 1

	// Add batch_id filter if provided
	if batchIDStr != "" {
		batchID, err := strconv.Atoi(batchIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid batch_id format")
		}
		query += fmt.Sprintf(" AND e.batch_id = $%d", argIndex)
		args = append(args, batchID)
		argIndex++
	}

	// Add event_type filter if provided
	if eventType != "" {
		query += fmt.Sprintf(" AND e.event_type = $%d", argIndex)
		args = append(args, eventType)
		argIndex++
	}

	query += " ORDER BY e.timestamp DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve events")
	}
	defer rows.Close()

	// Parse results
	var eventList []map[string]interface{}
	for rows.Next() {
		var event models.Event
		var species, status, hatcheryName, companyName string
		var quantity int
		var metadata sql.NullString
		err := rows.Scan(
			&event.ID,
			&event.BatchID,
			&event.EventType,
			&event.Location,
			&event.Timestamp,
			&event.UpdatedAt,
			&event.IsActive,
			&metadata,
			&species,
			&quantity,
			&status,
			&hatcheryName,
			&companyName,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse event data")
		}

		// Parse metadata if available
		var metadataMap map[string]interface{}
		if metadata.Valid && metadata.String != "" {
			// You might want to implement JSON parsing here
			metadataMap = make(map[string]interface{})
		}
		// Create comprehensive data structure
		eventEntry := map[string]interface{}{
			"id":          event.ID,
			"batch_id":    event.BatchID,
			"event_type":  event.EventType,
			"location":    event.Location,
			"timestamp":   event.Timestamp,
			"updated_at":  event.UpdatedAt,
			"is_active":   event.IsActive,
			"metadata":    metadataMap,
			"batch_info": map[string]interface{}{
				"species":  species,
				"quantity": quantity,
				"status":   status,
			},
			"facility_info": map[string]interface{}{
				"hatchery_name": hatcheryName,
				"company_name":  companyName,
			},
		}

		eventList = append(eventList, eventEntry)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Events retrieved successfully",
		Data:    eventList,
	})
}

// GetEventByID retrieves a specific event record by ID
// @Summary Get event by ID
// @Description Retrieve a specific event record by its ID
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} SuccessResponse{data=models.Event}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /events/{id} [get]
func GetEventByID(c *fiber.Ctx) error {
	// Get event ID from params
	eventIDStr := c.Params("id")
	if eventIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Event ID is required")
	}

	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid event ID format")
	}
	// Query event with related information
	query := `
		SELECT 
			e.id, e.batch_id, e.event_type, e.location, 
			e.timestamp, e.updated_at, e.is_active, e.metadata,
			b.species, b.quantity, b.status,
			h.name AS hatchery_name,
			c.name AS company_name, c.location AS company_location,
			br.tx_id AS blockchain_tx_id,
			br.metadata_hash AS blockchain_metadata
		FROM event e
		INNER JOIN batch b ON e.batch_id = b.id
		INNER JOIN hatchery h ON b.hatchery_id = h.id
		INNER JOIN company c ON h.company_id = c.id
		LEFT JOIN blockchain_record br ON br.related_table = 'event' AND br.related_id = e.id
		WHERE e.id = $1 AND e.is_active = true
	`

	var event models.Event
	var species, status, hatcheryName, companyName, companyLocation string
	var quantity int
	var metadata, blockchainTxID, blockchainMetadata sql.NullString
	err = db.DB.QueryRow(query, eventID).Scan(
		&event.ID,
		&event.BatchID,
		&event.EventType,
		&event.Location,
		&event.Timestamp,
		&event.UpdatedAt,
		&event.IsActive,
		&metadata,
		&species,
		&quantity,
		&status,
		&hatcheryName,
		&companyName,
		&companyLocation,
		&blockchainTxID,
		&blockchainMetadata,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "Event not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve event")
	}

	// Parse metadata if available
	var metadataMap map[string]interface{}
	if metadata.Valid && metadata.String != "" {
		// You might want to implement JSON parsing here
		metadataMap = make(map[string]interface{})
	}
	// Create comprehensive response
	response := map[string]interface{}{
		"id":          event.ID,
		"batch_id":    event.BatchID,
		"event_type":  event.EventType,
		"location":    event.Location,
		"timestamp":   event.Timestamp,
		"updated_at":  event.UpdatedAt,
		"is_active":   event.IsActive,
		"metadata":    metadataMap,
		"batch_info": map[string]interface{}{
			"species":  species,
			"quantity": quantity,
			"status":   status,
		},
		"facility_info": map[string]interface{}{
			"hatchery_name":    hatcheryName,
			"company_name":     companyName,
			"company_location": companyLocation,
		},
	}

	// Add blockchain verification if available
	if blockchainTxID.Valid {
		response["blockchain_verification"] = map[string]interface{}{
			"tx_id":         blockchainTxID.String,
			"metadata_hash": blockchainMetadata.String,
			"explorer_url":  fmt.Sprintf("https://explorer.viechain.com/tx/%s", blockchainTxID.String),
		}
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Event retrieved successfully",
		Data:    response,
	})
}

// UpdateEvent updates an existing event record
// @Summary Update event
// @Description Update an existing event record
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param request body UpdateEventRequest true "Event update details"
// @Success 200 {object} SuccessResponse{data=models.Event}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /events/{id} [put]
func UpdateEvent(c *fiber.Ctx) error {
	// Get event ID from params
	eventIDStr := c.Params("id")
	if eventIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Event ID is required")
	}

	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid event ID format")
	}

	// Parse request body
	var req UpdateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Check if event exists
	var exists bool
	var batchID int
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM event WHERE id = $1 AND is_active = true), batch_id FROM event WHERE id = $1", eventID).Scan(&exists, &batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Event not found")
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		os.Getenv("BLOCKCHAIN_NODE_URL"),
		os.Getenv("BLOCKCHAIN_PRIVATE_KEY"),
		os.Getenv("BLOCKCHAIN_ACCOUNT"),
		os.Getenv("BLOCKCHAIN_CHAIN_ID"),
		os.Getenv("BLOCKCHAIN_CONSENSUS"),
	)
	// Record update on blockchain
	updateData := map[string]interface{}{
		"event_id":    eventID,
		"batch_id":    batchID,
		"event_type":  req.EventType,
		"location":    req.Location,
		"metadata":    req.Metadata,
		"updated_at":  time.Now(),
	}
	txID, err := blockchainClient.RecordEvent(
		strconv.Itoa(batchID),
		"event_update",
		"facility",
		"system",
		updateData,
	)
	if err != nil {
		fmt.Printf("Warning: Failed to record event update on blockchain: %v\n", err)
	}

	// Convert metadata to string (you might want to use JSON encoding)
	metadataStr := ""
	if req.Metadata != nil {
		// Implement JSON marshaling here if needed
		metadataStr = "{}"
	}
	// Update event in database
	query := `
		UPDATE event 
		SET event_type = $1, location = $2, metadata = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, batch_id, event_type, location, timestamp, updated_at, is_active, metadata
	`

	var event models.Event
	var metadata sql.NullString
	err = db.DB.QueryRow(
		query,
		req.EventType,
		req.Location,
		metadataStr,
		eventID,
	).Scan(
		&event.ID,
		&event.BatchID,
		&event.EventType,
		&event.Location,
		&event.Timestamp,
		&event.UpdatedAt,
		&event.IsActive,
		&metadata,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update event")
	}

	// Record blockchain transaction if successful
	if txID != "" {
		metadataHash, err := blockchainClient.HashData(updateData)
		if err != nil {
			fmt.Printf("Warning: Failed to generate metadata hash: %v\n", err)
		}

		_, err = db.DB.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "event", event.ID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Event updated successfully",
		Data:    event,
	})
}

// DeleteEvent soft deletes an event record
// @Summary Delete event
// @Description Soft delete an event record (sets is_active to false)
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /events/{id} [delete]
func DeleteEvent(c *fiber.Ctx) error {
	// Get event ID from params
	eventIDStr := c.Params("id")
	if eventIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Event ID is required")
	}

	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid event ID format")
	}

	// Check if event exists
	var exists bool
	var batchID int
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM event WHERE id = $1 AND is_active = true), batch_id FROM event WHERE id = $1", eventID).Scan(&exists, &batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Event not found")
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		os.Getenv("BLOCKCHAIN_NODE_URL"),
		os.Getenv("BLOCKCHAIN_PRIVATE_KEY"),
		os.Getenv("BLOCKCHAIN_ACCOUNT"),
		os.Getenv("BLOCKCHAIN_CHAIN_ID"),
		os.Getenv("BLOCKCHAIN_CONSENSUS"),
	)

	// Record deletion on blockchain
	deletionData := map[string]interface{}{
		"event_id":   eventID,
		"batch_id":   batchID,
		"action":     "soft_delete",
		"deleted_at": time.Now(),
	}
	txID, err := blockchainClient.RecordEvent(
		strconv.Itoa(batchID),
		"event_deletion",
		"facility",
		"system",
		deletionData,
	)
	if err != nil {
		fmt.Printf("Warning: Failed to record event deletion on blockchain: %v\n", err)
	}

	// Soft delete event
	_, err = db.DB.Exec("UPDATE event SET is_active = false, updated_at = NOW() WHERE id = $1", eventID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete event")
	}

	// Record blockchain transaction if successful
	if txID != "" {
		metadataHash, err := blockchainClient.HashData(deletionData)
		if err != nil {
			fmt.Printf("Warning: Failed to generate metadata hash: %v\n", err)
		}

		_, err = db.DB.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "event", eventID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Event deleted successfully",
	})
}
