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

// UpdateEnvironmentDataRequest represents a request to update environment data
type UpdateEnvironmentDataRequest struct {
	Temperature float64 `json:"temperature"`
	PH          float64 `json:"ph"`
	Salinity    float64 `json:"salinity"`
	Density     float64 `json:"density"`
	Age         int     `json:"age"`
}

// GetAllEnvironmentData retrieves all environment data records
// @Summary Get all environment data
// @Description Retrieve all environment data records with optional filtering
// @Tags environment
// @Accept json
// @Produce json
// @Param batch_id query int false "Filter by batch ID"
// @Param limit query int false "Limit number of results (default: 50)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Success 200 {object} SuccessResponse{data=[]models.EnvironmentData}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /environment [get]
func GetAllEnvironmentData(c *fiber.Ctx) error {
	// Parse query parameters
	batchIDStr := c.Query("batch_id")
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
			e.id, e.batch_id, e.temperature, e.ph, e.salinity, e.density, e.age, 
			e.timestamp, e.updated_at, e.is_active,
			b.species, b.quantity, b.status,
			h.name AS hatchery_name,
			c.name AS company_name
		FROM environment_data e
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

	query += " ORDER BY e.timestamp DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve environment data")
	}
	defer rows.Close()

	// Parse results
	var environmentDataList []map[string]interface{}
	for rows.Next() {
		var envData models.EnvironmentData
		var species, status, hatcheryName, companyName string
		var quantity int

		err := rows.Scan(
			&envData.ID,
			&envData.BatchID,
			&envData.Temperature,
			&envData.PH,
			&envData.Salinity,
			&envData.Density,
			&envData.Age,
			&envData.Timestamp,
			&envData.UpdatedAt,
			&envData.IsActive,
			&species,
			&quantity,
			&status,
			&hatcheryName,
			&companyName,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse environment data")
		}

		// Create comprehensive data structure
		envDataEntry := map[string]interface{}{
			"id":          envData.ID,
			"batch_id":    envData.BatchID,
			"temperature": envData.Temperature,
			"ph":          envData.PH,
			"salinity":    envData.Salinity,
			"density":     envData.Density,
			"age":         envData.Age,
			"timestamp":   envData.Timestamp,
			"updated_at":  envData.UpdatedAt,
			"is_active":   envData.IsActive,
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

		environmentDataList = append(environmentDataList, envDataEntry)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Environment data retrieved successfully",
		Data:    environmentDataList,
	})
}

// GetEnvironmentDataByID retrieves a specific environment data record by ID
// @Summary Get environment data by ID
// @Description Retrieve a specific environment data record by its ID
// @Tags environment
// @Accept json
// @Produce json
// @Param id path string true "Environment Data ID"
// @Success 200 {object} SuccessResponse{data=models.EnvironmentData}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /environment/{id} [get]
func GetEnvironmentDataByID(c *fiber.Ctx) error {
	// Get environment data ID from params
	envIDStr := c.Params("id")
	if envIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Environment data ID is required")
	}

	envID, err := strconv.Atoi(envIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid environment data ID format")
	}

	// Query environment data with related information
	query := `
		SELECT 
			e.id, e.batch_id, e.temperature, e.ph, e.salinity, e.density, e.age, 
			e.timestamp, e.updated_at, e.is_active,
			b.species, b.quantity, b.status,
			h.name AS hatchery_name,
			c.name AS company_name, c.location AS company_location,
			br.tx_id AS blockchain_tx_id,
			br.metadata_hash AS blockchain_metadata
		FROM environment_data e
		INNER JOIN batch b ON e.batch_id = b.id
		INNER JOIN hatchery h ON b.hatchery_id = h.id
		INNER JOIN company c ON h.company_id = c.id
		LEFT JOIN blockchain_record br ON br.related_table = 'environment_data' AND br.related_id = e.id
		WHERE e.id = $1 AND e.is_active = true
	`

	var envData models.EnvironmentData
	var species, status, hatcheryName, companyName, companyLocation string
	var quantity int
	var blockchainTxID, blockchainMetadata sql.NullString

	err = db.DB.QueryRow(query, envID).Scan(
		&envData.ID,
		&envData.BatchID,
		&envData.Temperature,
		&envData.PH,
		&envData.Salinity,
		&envData.Density,
		&envData.Age,
		&envData.Timestamp,
		&envData.UpdatedAt,
		&envData.IsActive,
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
			return fiber.NewError(fiber.StatusNotFound, "Environment data not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve environment data")
	}

	// Create comprehensive response
	response := map[string]interface{}{
		"id":          envData.ID,
		"batch_id":    envData.BatchID,
		"temperature": envData.Temperature,
		"ph":          envData.PH,
		"salinity":    envData.Salinity,
		"density":     envData.Density,
		"age":         envData.Age,
		"timestamp":   envData.Timestamp,
		"updated_at":  envData.UpdatedAt,
		"is_active":   envData.IsActive,
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
		Message: "Environment data retrieved successfully",
		Data:    response,
	})
}

// UpdateEnvironmentData updates an existing environment data record
// @Summary Update environment data
// @Description Update an existing environment data record
// @Tags environment
// @Accept json
// @Produce json
// @Param id path string true "Environment Data ID"
// @Param request body UpdateEnvironmentDataRequest true "Environment data update details"
// @Success 200 {object} SuccessResponse{data=models.EnvironmentData}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /environment/{id} [put]
func UpdateEnvironmentData(c *fiber.Ctx) error {
	// Get environment data ID from params
	envIDStr := c.Params("id")
	if envIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Environment data ID is required")
	}

	envID, err := strconv.Atoi(envIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid environment data ID format")
	}

	// Parse request body
	var req UpdateEnvironmentDataRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Check if environment data exists
	var exists bool
	var batchID int
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM environment_data WHERE id = $1 AND is_active = true), batch_id FROM environment_data WHERE id = $1", envID).Scan(&exists, &batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Environment data not found")
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
		"environment_id": envID,
		"batch_id":       batchID,
		"temperature":    req.Temperature,
		"ph":             req.PH,
		"salinity":       req.Salinity,
		"density":        req.Density,
		"age":            req.Age,
		"updated_at":     time.Now(),
	}
	txID, err := blockchainClient.RecordEvent(
		strconv.Itoa(batchID),
		"environment_update",
		"facility",
		"system",
		updateData,
	)
	if err != nil {
		fmt.Printf("Warning: Failed to record environment update on blockchain: %v\n", err)
	}

	// Update environment data in database
	query := `
		UPDATE environment_data 
		SET temperature = $1, ph = $2, salinity = $3, density = $4, age = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING id, batch_id, temperature, ph, salinity, density, age, timestamp, updated_at, is_active
	`

	var envData models.EnvironmentData
	err = db.DB.QueryRow(
		query,
		req.Temperature,
		req.PH,
		req.Salinity,
		req.Density,
		req.Age,
		envID,
	).Scan(
		&envData.ID,
		&envData.BatchID,
		&envData.Temperature,
		&envData.PH,
		&envData.Salinity,
		&envData.Density,
		&envData.Age,
		&envData.Timestamp,
		&envData.UpdatedAt,
		&envData.IsActive,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update environment data")
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
		`, "environment_data", envData.ID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Environment data updated successfully",
		Data:    envData,
	})
}

// DeleteEnvironmentData soft deletes an environment data record
// @Summary Delete environment data
// @Description Soft delete an environment data record (sets is_active to false)
// @Tags environment
// @Accept json
// @Produce json
// @Param id path string true "Environment Data ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /environment/{id} [delete]
func DeleteEnvironmentData(c *fiber.Ctx) error {
	// Get environment data ID from params
	envIDStr := c.Params("id")
	if envIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Environment data ID is required")
	}

	envID, err := strconv.Atoi(envIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid environment data ID format")
	}

	// Check if environment data exists
	var exists bool
	var batchID int
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM environment_data WHERE id = $1 AND is_active = true), batch_id FROM environment_data WHERE id = $1", envID).Scan(&exists, &batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Environment data not found")
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
		"environment_id": envID,
		"batch_id":       batchID,
		"action":         "soft_delete",
		"deleted_at":     time.Now(),
	}
	txID, err := blockchainClient.RecordEvent(
		strconv.Itoa(batchID),
		"environment_deletion",
		"facility",
		"system",
		deletionData,
	)
	if err != nil {
		fmt.Printf("Warning: Failed to record environment deletion on blockchain: %v\n", err)
	}

	// Soft delete environment data
	_, err = db.DB.Exec("UPDATE environment_data SET is_active = false, updated_at = NOW() WHERE id = $1", envID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete environment data")
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
		`, "environment_data", envID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Environment data deleted successfully",
	})
}
