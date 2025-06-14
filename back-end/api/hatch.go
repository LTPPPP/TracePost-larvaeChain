package api

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"github.com/LTPPPP/TracePost-larvaeChain/models"
)

// CreateHatcheryRequest represents a request to create a new hatchery
type CreateHatcheryRequest struct {
	Name      string `json:"name"`
	CompanyID int    `json:"company_id"`
}

// UpdateHatcheryRequest represents a request to update a hatchery
type UpdateHatcheryRequest struct {
	Name     string `json:"name"`
}

// GetAllHatcheries returns all hatcheries
// @Summary Get all hatcheries
// @Description Retrieve all shrimp hatcheries
// @Tags hatcheries
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse{data=[]models.Hatchery}
// @Failure 500 {object} ErrorResponse
// @Router /hatcheries [get]
func GetAllHatcheries(c *fiber.Ctx) error {
	// Query hatcheries from database with company information
	rows, err := db.DB.Query(`
		SELECT h.id, h.name, h.company_id, h.created_at, h.updated_at, h.is_active,
			   c.id, c.name, c.type, c.location, c.contact_info, c.created_at, c.updated_at, c.is_active
		FROM hatchery h
		LEFT JOIN company c ON h.company_id = c.id
		WHERE h.is_active = true
		ORDER BY h.created_at DESC
	`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse hatcheries
	var hatcheries []models.Hatchery
	for rows.Next() {
		var hatchery models.Hatchery
		var company models.Company
		err := rows.Scan(
			&hatchery.ID,
			&hatchery.Name,
			&hatchery.CompanyID,
			&hatchery.CreatedAt,
			&hatchery.UpdatedAt,
			&hatchery.IsActive,
			&company.ID,
			&company.Name,
			&company.Type,
			&company.Location,
			&company.ContactInfo,
			&company.CreatedAt,
			&company.UpdatedAt,
			&company.IsActive,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error parsing hatchery data")
		}
		hatchery.Company = company
		hatcheries = append(hatcheries, hatchery)
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Hatcheries retrieved successfully",
		Data:    hatcheries,
	})
}

// GetHatcheryByID returns a hatchery by ID
// @Summary Get hatchery by ID 
// @Description Retrieve a shrimp hatchery by its ID
// @Tags hatcheries
// @Accept json
// @Produce json
// @Param hatcheryId path string true "Hatchery ID"
// @Success 200 {object} SuccessResponse{data=models.Hatchery}
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /hatcheries/{hatcheryId} [get]
func GetHatcheryByID(c *fiber.Ctx) error {
	// Get hatchery ID from params
	hatcheryIDStr := c.Params("hatcheryId")
	if hatcheryIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Hatchery ID is required")
	}
	
	hatcheryID, err := strconv.Atoi(hatcheryIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid hatchery ID format")
	}

	// Query hatchery from database with company information
	var hatchery models.Hatchery
	var company models.Company
	query := `
		SELECT h.id, h.name, h.company_id, h.created_at, h.updated_at, h.is_active,
			   c.id, c.name, c.type, c.location, c.contact_info, c.created_at, c.updated_at, c.is_active
		FROM hatchery h
		LEFT JOIN company c ON h.company_id = c.id
		WHERE h.id = $1 AND h.is_active = true
	`
	err = db.DB.QueryRow(query, hatcheryID).Scan(
		&hatchery.ID,
		&hatchery.Name,
		&hatchery.CompanyID,
		&hatchery.CreatedAt,
		&hatchery.UpdatedAt,
		&hatchery.IsActive,
		&company.ID,
		&company.Name,
		&company.Type,
		&company.Location,
		&company.ContactInfo,
		&company.CreatedAt,
		&company.UpdatedAt,
		&company.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "Hatchery not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	
	// Set company information
	hatchery.Company = company

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Hatchery retrieved successfully",
		Data:    hatchery,
	})
}

// CreateHatchery creates a new hatchery
// @Summary Create a new hatchery
// @Description Create a new shrimp hatchery
// @Tags hatcheries
// @Accept json
// @Produce json
// @Param request body CreateHatcheryRequest true "Hatchery creation details"
// @Success 201 {object} SuccessResponse{data=models.Hatchery}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /hatcheries [post]
func CreateHatchery(c *fiber.Ctx) error {
	// Parse request body
	var req CreateHatcheryRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate input
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Hatchery name is required")
	}

	// Check if company exists
	if req.CompanyID > 0 {
		var exists bool
		err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM company WHERE id = $1 AND is_active = true)", req.CompanyID).Scan(&exists)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Database error")
		}
		if !exists {
			return fiber.NewError(fiber.StatusBadRequest, "Company not found")
		}
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Insert hatchery into database
	query := `
		INSERT INTO hatchery (name, company_id, created_at, updated_at, is_active)
		VALUES ($1, $2, NOW(), NOW(), true)
		RETURNING id, created_at, updated_at
	`
	var hatchery models.Hatchery
	hatchery.Name = req.Name
	hatchery.CompanyID = req.CompanyID
	hatchery.IsActive = true

	err := db.DB.QueryRow(
		query,
		hatchery.Name,
		hatchery.CompanyID,
	).Scan(&hatchery.ID, &hatchery.CreatedAt, &hatchery.UpdatedAt)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save hatchery to database")
	}

	// Get company information
	var company models.Company
	err = db.DB.QueryRow(`
		SELECT id, name, type, location, contact_info, created_at, updated_at, is_active 
		FROM company 
		WHERE id = $1 AND is_active = true`,
		hatchery.CompanyID,
	).Scan(
		&company.ID,
		&company.Name,
		&company.Type,
		&company.Location,
		&company.ContactInfo,
		&company.CreatedAt,
		&company.UpdatedAt,
		&company.IsActive,
	)
	if err != nil && err != sql.ErrNoRows {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get company information")
	}
	
	// Set company information
	hatchery.Company = company

	// Create hatchery on blockchain
	txID, err := blockchainClient.CreateHatchery(
		strconv.Itoa(hatchery.ID),
		hatchery.Name,
		company.Location,
		company.ContactInfo,
		strconv.Itoa(hatchery.CompanyID),
	)
	if err != nil {
		// Log the error but continue - blockchain is secondary to database
		fmt.Printf("Warning: Failed to record hatchery on blockchain: %v\n", err)
	}

	// Record blockchain transaction
	if txID != "" {
		// Generate metadata hash
		metadata := map[string]interface{}{
			"hatchery_id": hatchery.ID,
			"name":        hatchery.Name,
			"company_id":  hatchery.CompanyID,
			"created_at":  hatchery.CreatedAt,
		}
		metadataHash, err := blockchainClient.HashData(metadata)
		if err != nil {
			fmt.Printf("Warning: Failed to generate metadata hash: %v\n", err)
		}

		// Save blockchain record
		_, err = db.DB.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "hatchery", hatchery.ID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Hatchery created successfully",
		Data:    hatchery,
	})
}

// UpdateHatchery updates a hatchery
// @Summary Update a hatchery
// @Description Update a shrimp hatchery
// @Tags hatcheries
// @Accept json
// @Produce json
// @Param hatcheryId path string true "Hatchery ID"
// @Param request body UpdateHatcheryRequest true "Hatchery update details"
// @Success 200 {object} SuccessResponse{data=models.Hatchery}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /hatcheries/{hatcheryId} [put]
func UpdateHatchery(c *fiber.Ctx) error {
	// Get hatchery ID from params
	hatcheryIDStr := c.Params("hatcheryId")
	if hatcheryIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Hatchery ID is required")
	}
	
	hatcheryID, err := strconv.Atoi(hatcheryIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid hatchery ID format")
	}

	// Parse request body
	var req UpdateHatcheryRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Check if hatchery exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM hatchery WHERE id = $1 AND is_active = true)", hatcheryID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Hatchery not found")
	}

	// Get existing hatchery data
	var hatchery models.Hatchery
	query := `
		SELECT id, name, company_id, created_at, updated_at, is_active
		FROM hatchery
		WHERE id = $1 AND is_active = true
	`
	err = db.DB.QueryRow(query, hatcheryID).Scan(
		&hatchery.ID,
		&hatchery.Name,
		&hatchery.CompanyID,
		&hatchery.CreatedAt,
		&hatchery.UpdatedAt,
		&hatchery.IsActive,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Update hatchery fields if provided
	if req.Name != "" {
		hatchery.Name = req.Name
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Update hatchery in database
	updateQuery := `
		UPDATE hatchery 
		SET name = $1, updated_at = NOW() 
		WHERE id = $2 AND is_active = true
		RETURNING updated_at
	`
	err = db.DB.QueryRow(
		updateQuery,
		hatchery.Name,
		hatchery.ID,
	).Scan(&hatchery.UpdatedAt)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update hatchery in database")
	}

	// Get company information for the blockchain record
	var companyInfo models.Company
	err = db.DB.QueryRow(`SELECT location, contact_info FROM company WHERE id = $1 AND is_active = true`, 
		hatchery.CompanyID).Scan(&companyInfo.Location, &companyInfo.ContactInfo)
	if err != nil && err != sql.ErrNoRows {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get company information")
	}
	
	// Update hatchery on blockchain
	txID, err := blockchainClient.UpdateHatchery(
		strconv.Itoa(hatchery.ID),
		hatchery.Name,
		companyInfo.Location,
		companyInfo.ContactInfo,
		strconv.Itoa(hatchery.CompanyID),
	)
	if err != nil {
		// Log the error but continue - blockchain is secondary to database
		fmt.Printf("Warning: Failed to update hatchery on blockchain: %v\n", err)
	}

	// Record blockchain transaction
	if txID != "" {
		// Generate metadata hash
		metadata := map[string]interface{}{
			"hatchery_id": hatchery.ID,
			"name":        hatchery.Name,
			"company_id":  hatchery.CompanyID,
			"updated_at":  hatchery.UpdatedAt,
		}
		metadataHash, err := blockchainClient.HashData(metadata)
		if err != nil {
			fmt.Printf("Warning: Failed to generate metadata hash: %v\n", err)
		}

		// Save blockchain record
		_, err = db.DB.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "hatchery", hatchery.ID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Hatchery updated successfully",
		Data:    hatchery,
	})
}

// DeleteHatchery soft-deletes a hatchery
// @Summary Delete a hatchery
// @Description Delete a shrimp hatchery (soft delete)
// @Tags hatcheries
// @Accept json
// @Produce json
// @Param hatcheryId path string true "Hatchery ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /hatcheries/{hatcheryId} [delete]
func DeleteHatchery(c *fiber.Ctx) error {
	// Get hatchery ID from params
	hatcheryIDStr := c.Params("hatcheryId")
	if hatcheryIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Hatchery ID is required")
	}
	
	hatcheryID, err := strconv.Atoi(hatcheryIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid hatchery ID format")
	}

	// Check if hatchery exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM hatchery WHERE id = $1 AND is_active = true)", hatcheryID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Hatchery not found")
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Soft delete hatchery in database
	_, err = db.DB.Exec(
		"UPDATE hatchery SET is_active = false, updated_at = NOW() WHERE id = $1",
		hatcheryID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete hatchery from database")
	}

	// Record deletion on blockchain
	txID, err := blockchainClient.DeleteHatchery(strconv.Itoa(hatcheryID))
	if err != nil {
		// Log the error but continue - blockchain is secondary to database
		fmt.Printf("Warning: Failed to record hatchery deletion on blockchain: %v\n", err)
	}

	// Record blockchain transaction
	if txID != "" {
		// Generate metadata hash
		metadata := map[string]interface{}{
			"hatchery_id": hatcheryID,
			"deleted_at":  time.Now(),
		}
		metadataHash, err := blockchainClient.HashData(metadata)
		if err != nil {
			fmt.Printf("Warning: Failed to generate metadata hash: %v\n", err)
		}

		// Save blockchain record
		_, err = db.DB.Exec(`
			INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), true)
		`, "hatchery", hatcheryID, txID, metadataHash)
		if err != nil {
			fmt.Printf("Warning: Failed to save blockchain record: %v\n", err)
		}
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Hatchery deleted successfully",
	})
}

// GetHatcheryBatches returns all batches for a hatchery
// @Summary Get hatchery batches
// @Description Retrieve all batches for a shrimp hatchery
// @Tags hatcheries
// @Accept json
// @Produce json
// @Param hatcheryId path string true "Hatchery ID"
// @Success 200 {object} SuccessResponse{data=[]models.Batch}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /hatcheries/{hatcheryId}/batches [get]
func GetHatcheryBatches(c *fiber.Ctx) error {
	// Get hatchery ID from params
	hatcheryIDStr := c.Params("hatcheryId")
	if hatcheryIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Hatchery ID is required")
	}
	
	hatcheryID, err := strconv.Atoi(hatcheryIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid hatchery ID format")
	}

	// Check if hatchery exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM hatchery WHERE id = $1 AND is_active = true)", hatcheryID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Hatchery not found")
	}

	// Query batches from database
	rows, err := db.DB.Query(`
		SELECT id, hatchery_id, species, quantity, status, created_at, updated_at, is_active
		FROM batch
		WHERE hatchery_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`, hatcheryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse batches
	var batches []models.Batch
	for rows.Next() {
		var batch models.Batch
		err := rows.Scan(
			&batch.ID,
			&batch.HatcheryID,
			&batch.Species,
			&batch.Quantity,
			&batch.Status,
			&batch.CreatedAt,
			&batch.UpdatedAt,
			&batch.IsActive,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse batch data")
		}
		batches = append(batches, batch)
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Hatchery batches retrieved successfully",
		Data:    batches,
	})
}

// GetHatcheryStats returns statistics for a hatchery
// @Summary Get hatchery statistics
// @Description Retrieve statistics for all shrimp hatcheries
// @Tags hatcheries
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Router /hatcheries/stats [get]
func GetHatcheryStats(c *fiber.Ctx) error {
	// Query batch statistics from database grouped by hatchery
	rows, err := db.DB.Query(`
		SELECT h.id, h.name, COUNT(b.id) as batch_count, SUM(b.quantity) as total_quantity
		FROM hatchery h
		LEFT JOIN batch b ON h.id = b.hatchery_id AND b.is_active = true
		WHERE h.is_active = true
		GROUP BY h.id, h.name
		ORDER BY h.name
	`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse statistics
	type HatcheryStat struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		BatchCount    int    `json:"batch_count"`
		TotalQuantity int    `json:"total_quantity"`
	}

	var stats []HatcheryStat
	for rows.Next() {
		var stat HatcheryStat
		var batchCount sql.NullInt64
		var totalQuantity sql.NullInt64

		err := rows.Scan(
			&stat.ID,
			&stat.Name,
			&batchCount,
			&totalQuantity,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse hatchery statistics")
		}

		if batchCount.Valid {
			stat.BatchCount = int(batchCount.Int64)
		}
		if totalQuantity.Valid {
			stat.TotalQuantity = int(totalQuantity.Int64)
		}

		stats = append(stats, stat)
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Hatchery statistics retrieved successfully",
		Data:    stats,
	})
}

// Helper function to convert string to int
// Moved to batch.go to avoid duplication