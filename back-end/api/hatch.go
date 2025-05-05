package api

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vietchain/tracepost-larvae/blockchain"
	"github.com/vietchain/tracepost-larvae/db"
	"github.com/vietchain/tracepost-larvae/models"
)

// CreateHatcheryRequest represents a request to create a new hatchery
type CreateHatcheryRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Contact  string `json:"contact"`
}

// UpdateHatcheryRequest represents a request to update a hatchery
type UpdateHatcheryRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Contact  string `json:"contact"`
}

// GetAllHatcheries returns all hatcheries
// @Summary Get all hatcheries
// @Description Retrieve all shrimp hatcheries
// @Tags hatcheries
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse{data=[]models.Hatcheries}
// @Failure 500 {object} ErrorResponse
// @Router /hatcheries [get]
func GetAllHatcheries(c *fiber.Ctx) error {
	// Query hatcheries from database
	rows, err := db.DB.Query(`
		SELECT id, name, location, contact, created_at, updated_at
		FROM hatcheries
		ORDER BY name ASC
	`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse hatcheries
	var hatcheries []models.Hatcheries
	for rows.Next() {
		var hatchery models.Hatcheries
		err := rows.Scan(
			&hatchery.ID,
			&hatchery.Name,
			&hatchery.Location,
			&hatchery.Contact,
			&hatchery.CreatedAt,
			&hatchery.UpdatedAt,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse hatchery data")
		}
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
// @Param hatcheryId path int true "Hatchery ID"
// @Success 200 {object} SuccessResponse{data=models.Hatcheries}
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /hatcheries/{hatcheryId} [get]
func GetHatcheryByID(c *fiber.Ctx) error {
	// Get hatchery ID from params
	hatcheryID := c.Params("hatcheryId")
	if hatcheryID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Hatchery ID is required")
	}

	// Query hatchery from database
	var hatchery models.Hatcheries
	query := `
		SELECT id, name, location, contact, created_at, updated_at
		FROM hatcheries
		WHERE id = $1
	`
	err := db.DB.QueryRow(query, hatcheryID).Scan(
		&hatchery.ID,
		&hatchery.Name,
		&hatchery.Location,
		&hatchery.Contact,
		&hatchery.CreatedAt,
		&hatchery.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "Hatchery not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Get batches for this hatchery
	rows, err := db.DB.Query(`
		SELECT id, batch_id, creation_date, species, quantity, status, blockchain_tx_id, metadata_hash
		FROM batches
		WHERE hatchery_id = $1
		ORDER BY creation_date DESC
	`, hatchery.ID)
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
			&batch.BatchID,
			&batch.CreationDate,
			&batch.Species,
			&batch.Quantity,
			&batch.Status,
			&batch.BlockchainTxID,
			&batch.MetadataHash,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse batch data")
		}
		batch.HatcheryID = hatchery.ID
		batches = append(batches, batch)
	}

	// Assign batches to hatchery
	hatchery.Batches = batches

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
// @Success 201 {object} SuccessResponse{data=models.Hatcheries}
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
	if req.Name == "" || req.Location == "" || req.Contact == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Name, location, and contact are required")
	}

	// Insert hatchery into database
	query := `
		INSERT INTO hatcheries (name, location, contact, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	var hatchery models.Hatcheries
	hatchery.Name = req.Name
	hatchery.Location = req.Location
	hatchery.Contact = req.Contact

	err := db.DB.QueryRow(
		query,
		hatchery.Name,
		hatchery.Location,
		hatchery.Contact,
	).Scan(&hatchery.ID, &hatchery.CreatedAt, &hatchery.UpdatedAt)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save hatchery to database")
	}

	// Initialize blockchain client for identity creation if enabled
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain", // Add chainID parameter
		"poa",         // Add consensusType parameter
	)
	
	// Create DID for hatchery (if identity system is enabled)
	metadata := map[string]interface{}{
		"name":     hatchery.Name,
		"location": hatchery.Location,
		"contact":  hatchery.Contact,
	}
	
	// We're just logging this information rather than requiring it for the hatchery creation
	// In a production environment, this would be properly handled
	_, err = blockchainClient.IdentityClient.CreateDecentralizedID("hatchery", hatchery.Name, metadata)
	if err != nil {
		// Log error but continue since this is not critical
		fmt.Printf("Warning: Failed to create DID for hatchery: %v\n", err)
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
// @Description Update an existing shrimp hatchery
// @Tags hatcheries
// @Accept json
// @Produce json
// @Param hatcheryId path int true "Hatchery ID"
// @Param request body UpdateHatcheryRequest true "Hatchery update details"
// @Success 200 {object} SuccessResponse{data=models.Hatcheries}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /hatcheries/{hatcheryId} [put]
func UpdateHatchery(c *fiber.Ctx) error {
	// Get hatchery ID from params
	hatcheryID := c.Params("hatcheryId")
	if hatcheryID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Hatchery ID is required")
	}

	// Parse request body
	var req UpdateHatcheryRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Check if at least one field is provided
	if req.Name == "" && req.Location == "" && req.Contact == "" {
		return fiber.NewError(fiber.StatusBadRequest, "At least one field (name, location, or contact) must be provided")
	}

	// Check if hatchery exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM hatcheries WHERE id = $1)", hatcheryID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Hatchery not found")
	}

	// Get current hatchery data
	var hatchery models.Hatcheries
	err = db.DB.QueryRow(`
		SELECT name, location, contact
		FROM hatcheries
		WHERE id = $1
	`, hatcheryID).Scan(&hatchery.Name, &hatchery.Location, &hatchery.Contact)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Update fields if provided
	if req.Name != "" {
		hatchery.Name = req.Name
	}
	if req.Location != "" {
		hatchery.Location = req.Location
	}
	if req.Contact != "" {
		hatchery.Contact = req.Contact
	}

	// Update hatchery in database
	_, err = db.DB.Exec(`
		UPDATE hatcheries
		SET name = $1, location = $2, contact = $3, updated_at = NOW()
		WHERE id = $4
	`, hatchery.Name, hatchery.Location, hatchery.Contact, hatcheryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update hatchery in database")
	}

	// Get updated hatchery
	err = db.DB.QueryRow(`
		SELECT id, name, location, contact, created_at, updated_at
		FROM hatcheries
		WHERE id = $1
	`, hatcheryID).Scan(
		&hatchery.ID,
		&hatchery.Name,
		&hatchery.Location,
		&hatchery.Contact,
		&hatchery.CreatedAt,
		&hatchery.UpdatedAt,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Hatchery updated successfully",
		Data:    hatchery,
	})
}

// DeleteHatchery deletes a hatchery
// @Summary Delete a hatchery
// @Description Delete an existing shrimp hatchery
// @Tags hatcheries
// @Accept json
// @Produce json
// @Param hatcheryId path int true "Hatchery ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /hatcheries/{hatcheryId} [delete]
func DeleteHatchery(c *fiber.Ctx) error {
	// Get hatchery ID from params
	hatcheryID := c.Params("hatcheryId")
	if hatcheryID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Hatchery ID is required")
	}

	// Check if hatchery exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM hatcheries WHERE id = $1)", hatcheryID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Hatchery not found")
	}

	// Check if hatchery has any batches
	var batchCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM batches WHERE hatchery_id = $1", hatcheryID).Scan(&batchCount)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if batchCount > 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot delete hatchery with associated batches")
	}

	// Delete hatchery from database
	_, err = db.DB.Exec("DELETE FROM hatcheries WHERE id = $1", hatcheryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete hatchery from database")
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Hatchery deleted successfully",
	})
}

// GetHatcheryBatches returns all batches for a hatchery
// @Summary Get hatchery batches
// @Description Retrieve all batches for a specific hatchery
// @Tags hatcheries
// @Accept json
// @Produce json
// @Param hatcheryId path int true "Hatchery ID"
// @Success 200 {object} SuccessResponse{data=[]models.Batch}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /hatcheries/{hatcheryId}/batches [get]
func GetHatcheryBatches(c *fiber.Ctx) error {
	// Get hatchery ID from params
	hatcheryID := c.Params("hatcheryId")
	if hatcheryID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Hatchery ID is required")
	}

	// Check if hatchery exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM hatcheries WHERE id = $1)", hatcheryID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Hatchery not found")
	}

	// Query batches from database
	rows, err := db.DB.Query(`
		SELECT id, batch_id, creation_date, species, quantity, status, blockchain_tx_id, metadata_hash
		FROM batches
		WHERE hatchery_id = $1
		ORDER BY creation_date DESC
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
			&batch.BatchID,
			&batch.CreationDate,
			&batch.Species,
			&batch.Quantity,
			&batch.Status,
			&batch.BlockchainTxID,
			&batch.MetadataHash,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse batch data")
		}
		batch.HatcheryID, _ = convertToInt(hatcheryID)
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
// @Description Retrieve statistics for a specific hatchery
// @Tags hatcheries
// @Accept json
// @Produce json
// @Param hatcheryId path int true "Hatchery ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /hatcheries/{hatcheryId}/stats [get]
func GetHatcheryStats(c *fiber.Ctx) error {
	// Get hatchery ID from params
	hatcheryID := c.Params("hatcheryId")
	if hatcheryID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Hatchery ID is required")
	}

	// Check if hatchery exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM hatcheries WHERE id = $1)", hatcheryID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Hatchery not found")
	}

	// Get total number of batches
	var totalBatches int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM batches WHERE hatchery_id = $1", hatcheryID).Scan(&totalBatches)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Get total quantity of larvae
	var totalQuantity int
	err = db.DB.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM batches WHERE hatchery_id = $1", hatcheryID).Scan(&totalQuantity)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Get counts by status
	rows, err := db.DB.Query(`
		SELECT status, COUNT(*) as count
		FROM batches
		WHERE hatchery_id = $1
		GROUP BY status
	`, hatcheryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse status counts
	statusCounts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		err := rows.Scan(&status, &count)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse status count data")
		}
		statusCounts[status] = count
	}

	// Get counts by species
	rows, err = db.DB.Query(`
		SELECT species, COUNT(*) as count, SUM(quantity) as total_quantity
		FROM batches
		WHERE hatchery_id = $1
		GROUP BY species
	`, hatcheryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse species counts
	speciesCounts := make([]map[string]interface{}, 0)
	for rows.Next() {
		var species string
		var count int
		var totalSpeciesQuantity int
		err := rows.Scan(&species, &count, &totalSpeciesQuantity)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse species count data")
		}
		speciesCounts = append(speciesCounts, map[string]interface{}{
			"species":        species,
			"count":          count,
			"total_quantity": totalSpeciesQuantity,
		})
	}

	// Get recent batches
	rows, err = db.DB.Query(`
		SELECT id, batch_id, creation_date, species, quantity, status
		FROM batches
		WHERE hatchery_id = $1
		ORDER BY creation_date DESC
		LIMIT 5
	`, hatcheryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse recent batches
	recentBatches := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var batchID string
		var creationDate time.Time
		var species string
		var quantity int
		var status string
		err := rows.Scan(&id, &batchID, &creationDate, &species, &quantity, &status)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse recent batch data")
		}
		recentBatches = append(recentBatches, map[string]interface{}{
			"id":            id,
			"batch_id":      batchID,
			"creation_date": creationDate,
			"species":       species,
			"quantity":      quantity,
			"status":        status,
		})
	}

	// Return success response with statistics
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Hatchery statistics retrieved successfully",
		Data: map[string]interface{}{
			"total_batches":  totalBatches,
			"total_quantity": totalQuantity,
			"status_counts":  statusCounts,
			"species_counts": speciesCounts,
			"recent_batches": recentBatches,
		},
	})
}

// Helper function to convert string to int
func convertToInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}