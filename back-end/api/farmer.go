package api

import (
	"strconv"
	"time"
	
	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/config"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
)

// Farm represents a farming facility in the supply chain
type Farm struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CompanyID   int       `json:"company_id"`
	Location    string    `json:"location"`
	Contact     string    `json:"contact"`
	FarmType    string    `json:"farm_type"` // "pond", "tank", "cage", etc.
	AreaSize    float64   `json:"area_size"` // in hectares or square meters
	Capacity    int       `json:"capacity"`  // maximum capacity
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsActive    bool      `json:"is_active"`
	Coordinates string    `json:"coordinates,omitempty"` // GPS coordinates
	DID         string    `json:"did,omitempty"`         // Decentralized Identity
}

// FarmingRecord represents a record of farming activities
type FarmingRecord struct {
	ID            string                 `json:"id"`
	FarmID        string                 `json:"farm_id"`
	BatchID       string                 `json:"batch_id"`
	RecordType    string                 `json:"record_type"` // "feeding", "treatment", "monitoring", etc.
	RecordedAt    time.Time              `json:"recorded_at"`
	RecordedBy    string                 `json:"recorded_by"`
	Description   string                 `json:"description"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	BlockchainTxID string                `json:"blockchain_tx_id,omitempty"`
}

// CreateFarmRequest represents a request to create a new farm
type CreateFarmRequest struct {
	Name        string  `json:"name"`
	CompanyID   int     `json:"company_id"`
	Location    string  `json:"location"`
	Contact     string  `json:"contact"`
	FarmType    string  `json:"farm_type"`
	AreaSize    float64 `json:"area_size,omitempty"`
	Capacity    int     `json:"capacity,omitempty"`
	Coordinates string  `json:"coordinates,omitempty"`
}

// UpdateFarmRequest represents a request to update a farm
type UpdateFarmRequest struct {
	Name        string  `json:"name,omitempty"`
	Location    string  `json:"location,omitempty"`
	Contact     string  `json:"contact,omitempty"`
	FarmType    string  `json:"farm_type,omitempty"`
	AreaSize    float64 `json:"area_size,omitempty"`
	Capacity    int     `json:"capacity,omitempty"`
	Coordinates string  `json:"coordinates,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// CreateFarmingRecordRequest represents a request to create a farming record
type CreateFarmingRecordRequest struct {
	FarmID      string                 `json:"farm_id"`
	BatchID     string                 `json:"batch_id"`
	RecordType  string                 `json:"record_type"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ReceiveBatchRequest represents a request to receive a batch at a farm
type ReceiveBatchRequest struct {
	BatchID      string                 `json:"batch_id"`
	Quantity     int                    `json:"quantity"`
	ReceiptNotes string                 `json:"receipt_notes,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// TransferBatchRequest represents a request to transfer a batch from a farm
type TransferBatchRequest struct {
	BatchID       string                 `json:"batch_id"`
	Quantity      int                    `json:"quantity"`
	Destination   string                 `json:"destination"`
	TransferNotes string                 `json:"transfer_notes,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// GetAllFarms gets all farms
// @Summary Get all farms
// @Description Retrieve all farming facilities
// @Tags farms
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse{data=[]Farm}
// @Failure 500 {object} ErrorResponse
// @Router /farms [get]
func GetAllFarms(c *fiber.Ctx) error {
	// Query all farms from the database
	rows, err := db.DB.Query(`
		SELECT id, name, company_id, location, contact, farm_type, 
		       area_size, capacity, created_at, updated_at, is_active, coordinates, did
		FROM farms
		WHERE is_active = true
		ORDER BY name
	`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	defer rows.Close()
	
	// Process farms
	var farms []Farm
	for rows.Next() {
		var farm Farm
		err := rows.Scan(
			&farm.ID,
			&farm.Name,
			&farm.CompanyID,
			&farm.Location,
			&farm.Contact,
			&farm.FarmType,
			&farm.AreaSize,
			&farm.Capacity,
			&farm.CreatedAt,
			&farm.UpdatedAt,
			&farm.IsActive,
			&farm.Coordinates,
			&farm.DID,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error parsing farm data: "+err.Error())
		}
		
		farms = append(farms, farm)
	}
	
	// Return farms
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Farms retrieved successfully",
		Data:    farms,
	})
}

// GetFarmByID gets a farm by ID
// @Summary Get farm by ID
// @Description Retrieve a farming facility by its ID
// @Tags farms
// @Accept json
// @Produce json
// @Param farmId path string true "Farm ID"
// @Success 200 {object} SuccessResponse{data=Farm}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /farms/{farmId} [get]
func GetFarmByID(c *fiber.Ctx) error {
	// Get farm ID from path
	farmID := c.Params("farmId")
	if farmID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Farm ID is required")
	}
	
	// Query farm from database
	var farm Farm
	err := db.DB.QueryRow(`
		SELECT id, name, company_id, location, contact, farm_type, 
		       area_size, capacity, created_at, updated_at, is_active, coordinates, did
		FROM farms
		WHERE id = $1
	`, farmID).Scan(
		&farm.ID,
		&farm.Name,
		&farm.CompanyID,
		&farm.Location,
		&farm.Contact,
		&farm.FarmType,
		&farm.AreaSize,
		&farm.Capacity,
		&farm.CreatedAt,
		&farm.UpdatedAt,
		&farm.IsActive,
		&farm.Coordinates,
		&farm.DID,
	)
	
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return fiber.NewError(fiber.StatusNotFound, "Farm not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	
	// Return farm
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Farm retrieved successfully",
		Data:    farm,
	})
}

// CreateFarm creates a new farm
// @Summary Create a new farm
// @Description Create a new farming facility
// @Tags farms
// @Accept json
// @Produce json
// @Param request body CreateFarmRequest true "Farm creation details"
// @Success 201 {object} SuccessResponse{data=Farm}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /farms [post]
func CreateFarm(c *fiber.Ctx) error {
	// Parse request
	var req CreateFarmRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format: "+err.Error())
	}
	
	// Validate required fields
	if req.Name == "" || req.CompanyID == 0 || req.Location == "" || req.FarmType == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Name, company ID, location, and farm type are required")
	}
	
	// Check if company exists
	var companyExists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1)", req.CompanyID).Scan(&companyExists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !companyExists {
		return fiber.NewError(fiber.StatusBadRequest, "Company does not exist")
	}
	
	// Generate farm ID
	farmID := "farm-" + time.Now().Format("20060102150405")
	now := time.Now()
	
	// Initialize blockchain client
	cfg := config.GetConfig()
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Create DID for farm if identity is enabled
	var did string
	if cfg.IdentityEnabled {
		identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)
		didResult, err := identityClient.CreateDecentralizedID("farm", req.Name, map[string]interface{}{
			"company_id": req.CompanyID,
			"location":   req.Location,
			"farm_type":  req.FarmType,
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to create DID: "+err.Error())
		}
		did = didResult.DID
	}
	
	// Insert farm into database
	_, err = db.DB.Exec(`
		INSERT INTO farms (
			id, name, company_id, location, contact, farm_type, 
			area_size, capacity, created_at, updated_at, is_active, coordinates, did
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`,
		farmID,
		req.Name,
		req.CompanyID,
		req.Location,
		req.Contact,
		req.FarmType,
		req.AreaSize,
		req.Capacity,
		now,
		now,
		true,
		req.Coordinates,
		did,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create farm: "+err.Error())
	}
	
	// Create farm on blockchain
	_, err = blockchainClient.SubmitTransaction("CREATE_FARM", map[string]interface{}{
		"farm_id":     farmID,
		"name":        req.Name,
		"company_id":  req.CompanyID,
		"location":    req.Location,
		"farm_type":   req.FarmType,
		"did":         did,
		"coordinates": req.Coordinates,
		"timestamp":   now,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record farm on blockchain: "+err.Error())
	}
	
	// Return the created farm
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Farm created successfully",
		Data: Farm{
			ID:          farmID,
			Name:        req.Name,
			CompanyID:   req.CompanyID,
			Location:    req.Location,
			Contact:     req.Contact,
			FarmType:    req.FarmType,
			AreaSize:    req.AreaSize,
			Capacity:    req.Capacity,
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
			Coordinates: req.Coordinates,
			DID:         did,
		},
	})
}

// UpdateFarm updates a farm
// @Summary Update a farm
// @Description Update a farming facility
// @Tags farms
// @Accept json
// @Produce json
// @Param farmId path string true "Farm ID"
// @Param request body UpdateFarmRequest true "Farm update details"
// @Success 200 {object} SuccessResponse{data=Farm}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /farms/{farmId} [put]
func UpdateFarm(c *fiber.Ctx) error {
	// Get farm ID from path
	farmID := c.Params("farmId")
	if farmID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Farm ID is required")
	}
	
	// Parse request
	var req UpdateFarmRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format: "+err.Error())
	}
	
	// Check if farm exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM farms WHERE id = $1)", farmID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Farm not found")
	}
	
	// Get existing farm data
	var farm Farm
	err = db.DB.QueryRow(`
		SELECT id, name, company_id, location, contact, farm_type, 
		       area_size, capacity, created_at, updated_at, is_active, coordinates, did
		FROM farms
		WHERE id = $1
	`, farmID).Scan(
		&farm.ID,
		&farm.Name,
		&farm.CompanyID,
		&farm.Location,
		&farm.Contact,
		&farm.FarmType,
		&farm.AreaSize,
		&farm.Capacity,
		&farm.CreatedAt,
		&farm.UpdatedAt,
		&farm.IsActive,
		&farm.Coordinates,
		&farm.DID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve farm: "+err.Error())
	}
	
	// Update fields if provided
	updateSQL := `
		UPDATE farms SET 
			updated_at = $1
	`
	params := []interface{}{time.Now()}
	paramCount := 1
	
	updateData := make(map[string]interface{})
	
	if req.Name != "" {
		paramCount++
		updateSQL += ", name = $" + strconv.Itoa(paramCount)
		params = append(params, req.Name)
		updateData["name"] = req.Name
		farm.Name = req.Name
	}
	
	if req.Location != "" {
		paramCount++
		updateSQL += ", location = $" + strconv.Itoa(paramCount)
		params = append(params, req.Location)
		updateData["location"] = req.Location
		farm.Location = req.Location
	}
	
	if req.Contact != "" {
		paramCount++
		updateSQL += ", contact = $" + strconv.Itoa(paramCount)
		params = append(params, req.Contact)
		updateData["contact"] = req.Contact
		farm.Contact = req.Contact
	}
	
	if req.FarmType != "" {
		paramCount++
		updateSQL += ", farm_type = $" + strconv.Itoa(paramCount)
		params = append(params, req.FarmType)
		updateData["farm_type"] = req.FarmType
		farm.FarmType = req.FarmType
	}
	
	if req.AreaSize > 0 {
		paramCount++
		updateSQL += ", area_size = $" + strconv.Itoa(paramCount)
		params = append(params, req.AreaSize)
		updateData["area_size"] = req.AreaSize
		farm.AreaSize = req.AreaSize
	}
	
	if req.Capacity > 0 {
		paramCount++
		updateSQL += ", capacity = $" + strconv.Itoa(paramCount)
		params = append(params, req.Capacity)
		updateData["capacity"] = req.Capacity
		farm.Capacity = req.Capacity
	}
	
	if req.Coordinates != "" {
		paramCount++
		updateSQL += ", coordinates = $" + strconv.Itoa(paramCount)
		params = append(params, req.Coordinates)
		updateData["coordinates"] = req.Coordinates
		farm.Coordinates = req.Coordinates
	}
	
	if req.IsActive != nil {
		paramCount++
		updateSQL += ", is_active = $" + strconv.Itoa(paramCount)
		params = append(params, *req.IsActive)
		updateData["is_active"] = *req.IsActive
		farm.IsActive = *req.IsActive
	}
	
	// Add WHERE clause
	updateSQL += " WHERE id = $" + strconv.Itoa(paramCount+1)
	params = append(params, farmID)
	
	// Execute update
	_, err = db.DB.Exec(updateSQL, params...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update farm: "+err.Error())
	}
	
	// Update farm on blockchain
	cfg := config.GetConfig()
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	updateData["farm_id"] = farmID
	updateData["timestamp"] = time.Now()
	
	_, err = blockchainClient.SubmitTransaction("UPDATE_FARM", updateData)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record farm update on blockchain: "+err.Error())
	}
	
	// Return updated farm
	farm.UpdatedAt = time.Now()
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Farm updated successfully",
		Data:    farm,
	})
}

// DeleteFarm deletes a farm (soft delete)
// @Summary Delete a farm
// @Description Delete a farming facility (soft delete)
// @Tags farms
// @Accept json
// @Produce json
// @Param farmId path string true "Farm ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /farms/{farmId} [delete]
func DeleteFarm(c *fiber.Ctx) error {
	// Get farm ID from path
	farmID := c.Params("farmId")
	if farmID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Farm ID is required")
	}
	
	// Check if farm exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM farms WHERE id = $1)", farmID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Farm not found")
	}
	
	// Soft delete (update is_active flag)
	_, err = db.DB.Exec("UPDATE farms SET is_active = false, updated_at = $1 WHERE id = $2", time.Now(), farmID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete farm: "+err.Error())
	}
	
	// Record on blockchain
	cfg := config.GetConfig()
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	_, err = blockchainClient.SubmitTransaction("DELETE_FARM", map[string]interface{}{
		"farm_id":   farmID,
		"is_active": false,
		"timestamp": time.Now(),
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record farm deletion on blockchain: "+err.Error())
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Farm deleted successfully",
	})
}

// GetFarmBatches gets all batches at a farm
// @Summary Get farm batches
// @Description Retrieve all batches at a farming facility
// @Tags farms
// @Accept json
// @Produce json
// @Param farmId path string true "Farm ID"
// @Success 200 {object} SuccessResponse{data=[]models.Batch}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /farms/{farmId}/batches [get]
func GetFarmBatches(c *fiber.Ctx) error {
	// Get farm ID from path
	farmID := c.Params("farmId")
	if farmID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Farm ID is required")
	}
	
	// Check if farm exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM farms WHERE id = $1)", farmID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Farm not found")
	}
	
	// Query batches at the farm
	rows, err := db.DB.Query(`
		SELECT b.id, b.species, b.quantity, b.status, b.created_at, b.updated_at, b.is_active, b.hatchery_id
		FROM batches b
		JOIN farm_batches fb ON b.id = fb.batch_id
		WHERE fb.farm_id = $1
		ORDER BY b.created_at DESC
	`, farmID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve batches: "+err.Error())
	}
	defer rows.Close()
	
	// Process batches
	var batches []map[string]interface{}
	for rows.Next() {
		var batchID string
		var species string
		var quantity int
		var status string
		var createdAt time.Time
		var updatedAt time.Time
		var isActive bool
		var hatcheryID int
		
		err := rows.Scan(&batchID, &species, &quantity, &status, &createdAt, &updatedAt, &isActive, &hatcheryID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error parsing batch data: "+err.Error())
		}
		
		batch := map[string]interface{}{
			"id":          batchID,
			"species":     species,
			"quantity":    quantity,
			"status":      status,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
			"is_active":   isActive,
			"hatchery_id": hatcheryID,
		}
		
		batches = append(batches, batch)
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Farm batches retrieved successfully",
		Data:    batches,
	})
}

// CreateFarmingRecord creates a new farming record
// @Summary Create a farming record
// @Description Create a new record of farming activities
// @Tags farms
// @Accept json
// @Produce json
// @Param request body CreateFarmingRecordRequest true "Farming record details"
// @Success 201 {object} SuccessResponse{data=FarmingRecord}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /farms/records [post]
func CreateFarmingRecord(c *fiber.Ctx) error {
	// Parse request
	var req CreateFarmingRecordRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format: "+err.Error())
	}
	
	// Validate required fields
	if req.FarmID == "" || req.BatchID == "" || req.RecordType == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Farm ID, batch ID, and record type are required")
	}
	
	// Check if farm exists
	var farmExists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM farms WHERE id = $1)", req.FarmID).Scan(&farmExists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !farmExists {
		return fiber.NewError(fiber.StatusNotFound, "Farm not found")
	}
	
	// Check if batch exists
	var batchExists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE id = $1)", req.BatchID).Scan(&batchExists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !batchExists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}
	
	// Check if batch is at this farm
	var batchAtFarm bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM farm_batches WHERE farm_id = $1 AND batch_id = $2)", 
		req.FarmID, req.BatchID).Scan(&batchAtFarm)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !batchAtFarm {
		return fiber.NewError(fiber.StatusBadRequest, "Batch is not at this farm")
	}
	
	// Get user ID from token
	userID := c.Locals("user_id").(string)
	
	// Generate record ID
	recordID := "frec-" + time.Now().Format("20060102150405")
	now := time.Now()
	
	// Initialize blockchain client
	cfg := config.GetConfig()
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Record on blockchain
	txID, err := blockchainClient.SubmitTransaction("FARM_RECORD", map[string]interface{}{
		"record_id":    recordID,
		"farm_id":      req.FarmID,
		"batch_id":     req.BatchID,
		"record_type":  req.RecordType,
		"description":  req.Description,
		"recorded_by":  userID,
		"metadata":     req.Metadata,
		"timestamp":    now,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record on blockchain: "+err.Error())
	}
	
	// Insert into database
	_, err = db.DB.Exec(`
		INSERT INTO farming_records (
			id, farm_id, batch_id, record_type, recorded_at, 
			recorded_by, description, metadata, blockchain_tx_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		recordID,
		req.FarmID,
		req.BatchID,
		req.RecordType,
		now,
		userID,
		req.Description,
		req.Metadata,
		txID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create farming record: "+err.Error())
	}
	
	// Return the created record
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Farming record created successfully",
		Data: FarmingRecord{
			ID:            recordID,
			FarmID:        req.FarmID,
			BatchID:       req.BatchID,
			RecordType:    req.RecordType,
			RecordedAt:    now,
			RecordedBy:    userID,
			Description:   req.Description,
			Metadata:      req.Metadata,
			BlockchainTxID: txID,
		},
	})
}

// GetFarmRecords gets all farming records for a farm
// @Summary Get farm records
// @Description Retrieve all farming records for a farm
// @Tags farms
// @Accept json
// @Produce json
// @Param farmId path string true "Farm ID"
// @Success 200 {object} SuccessResponse{data=[]FarmingRecord}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /farms/{farmId}/records [get]
func GetFarmRecords(c *fiber.Ctx) error {
	// Get farm ID from path
	farmID := c.Params("farmId")
	if farmID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Farm ID is required")
	}
	
	// Check if farm exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM farms WHERE id = $1)", farmID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Farm not found")
	}
	
	// Query records for the farm
	rows, err := db.DB.Query(`
		SELECT id, farm_id, batch_id, record_type, recorded_at, 
		       recorded_by, description, metadata, blockchain_tx_id
		FROM farming_records
		WHERE farm_id = $1
		ORDER BY recorded_at DESC
	`, farmID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve farming records: "+err.Error())
	}
	defer rows.Close()
	
	// Process records
	var records []FarmingRecord
	for rows.Next() {
		var record FarmingRecord
		err := rows.Scan(
			&record.ID,
			&record.FarmID,
			&record.BatchID,
			&record.RecordType,
			&record.RecordedAt,
			&record.RecordedBy,
			&record.Description,
			&record.Metadata,
			&record.BlockchainTxID,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error parsing farming record data: "+err.Error())
		}
		
		records = append(records, record)
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Farm records retrieved successfully",
		Data:    records,
	})
}

// ReceiveBatch handles receiving a batch at a farm
// @Summary Receive batch at farm
// @Description Handle the receipt of a batch at a farming facility
// @Tags farms
// @Accept json
// @Produce json
// @Param farmId path string true "Farm ID"
// @Param request body ReceiveBatchRequest true "Batch receipt details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /farms/{farmId}/receive-batch [post]
func ReceiveBatch(c *fiber.Ctx) error {
	// Get farm ID from path
	farmID := c.Params("farmId")
	if farmID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Farm ID is required")
	}
	
	// Parse request
	var req ReceiveBatchRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format: "+err.Error())
	}
	
	// Validate required fields
	if req.BatchID == "" || req.Quantity <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID and quantity are required")
	}
	
	// Check if farm exists
	var farmExists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM farms WHERE id = $1)", farmID).Scan(&farmExists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !farmExists {
		return fiber.NewError(fiber.StatusNotFound, "Farm not found")
	}
	
	// Check if batch exists
	var batchExists bool
	var currentStatus string
	var currentQuantity int
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE id = $1), (SELECT status FROM batches WHERE id = $1), (SELECT quantity FROM batches WHERE id = $1)", 
		req.BatchID).Scan(&batchExists, &currentStatus, &currentQuantity)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !batchExists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}
	
	// Check if batch is already at this farm
	var batchAtFarm bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM farm_batches WHERE farm_id = $1 AND batch_id = $2)", 
		farmID, req.BatchID).Scan(&batchAtFarm)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if batchAtFarm {
		return fiber.NewError(fiber.StatusBadRequest, "Batch is already at this farm")
	}
	
	// Get user ID from token
	userID := c.Locals("user_id").(string)
	
	// Start a transaction
	tx, err := db.DB.Begin()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to start database transaction: "+err.Error())
	}
	
	// Associate batch with farm
	receiptID := "rcpt-" + time.Now().Format("20060102150405")
	now := time.Now()
	
	_, err = tx.Exec(`
		INSERT INTO farm_batches (farm_id, batch_id, received_at, received_by, quantity, receipt_notes)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, farmID, req.BatchID, now, userID, req.Quantity, req.ReceiptNotes)
	if err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to associate batch with farm: "+err.Error())
	}
	
	// Update batch status
	_, err = tx.Exec(`
		UPDATE batches 
		SET status = 'at_farm', updated_at = $1
		WHERE id = $2
	`, now, req.BatchID)
	if err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update batch status: "+err.Error())
	}
	
	// Create batch event
	_, err = tx.Exec(`
		INSERT INTO events (id, batch_id, event_type, location, actor_id, timestamp, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, 
		"evt-" + time.Now().Format("20060102150405"), 
		req.BatchID, 
		"batch_received", 
		farmID, 
		userID, 
		now, 
		req.Metadata,
	)
	if err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create batch event: "+err.Error())
	}
	
	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
	}
	
	// Record on blockchain
	cfg := config.GetConfig()
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	_, err = blockchainClient.SubmitTransaction("BATCH_RECEIVED", map[string]interface{}{
		"receipt_id":    receiptID,
		"farm_id":       farmID,
		"batch_id":      req.BatchID,
		"quantity":      req.Quantity,
		"received_by":   userID,
		"receipt_notes": req.ReceiptNotes,
		"metadata":      req.Metadata,
		"timestamp":     now,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record receipt on blockchain: "+err.Error())
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch received successfully at farm",
		Data: map[string]interface{}{
			"receipt_id":  receiptID,
			"farm_id":     farmID,
			"batch_id":    req.BatchID,
			"quantity":    req.Quantity,
			"received_at": now,
			"received_by": userID,
		},
	})
}

// TransferBatch handles transferring a batch from a farm
// @Summary Transfer batch from farm
// @Description Handle the transfer of a batch from a farming facility
// @Tags farms
// @Accept json
// @Produce json
// @Param farmId path string true "Farm ID"
// @Param request body TransferBatchRequest true "Batch transfer details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /farms/{farmId}/transfer-batch [post]
func TransferBatch(c *fiber.Ctx) error {
	// Get farm ID from path
	farmID := c.Params("farmId")
	if farmID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Farm ID is required")
	}
	
	// Parse request
	var req TransferBatchRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format: "+err.Error())
	}
	
	// Validate required fields
	if req.BatchID == "" || req.Quantity <= 0 || req.Destination == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID, quantity, and destination are required")
	}
	
	// Check if farm exists
	var farmExists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM farms WHERE id = $1)", farmID).Scan(&farmExists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !farmExists {
		return fiber.NewError(fiber.StatusNotFound, "Farm not found")
	}
	
	// Check if batch is at this farm
	var batchAtFarm bool
	var currentQuantity int
	err = db.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM farm_batches WHERE farm_id = $1 AND batch_id = $2),
		       (SELECT quantity FROM farm_batches WHERE farm_id = $1 AND batch_id = $2)
	`, farmID, req.BatchID).Scan(&batchAtFarm, &currentQuantity)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !batchAtFarm {
		return fiber.NewError(fiber.StatusBadRequest, "Batch is not at this farm")
	}
	
	// Check if quantity is valid
	if req.Quantity > currentQuantity {
		return fiber.NewError(fiber.StatusBadRequest, "Transfer quantity exceeds available quantity")
	}
	
	// Get user ID from token
	userID := c.Locals("user_id").(string)
	
	// Start a transaction
	tx, err := db.DB.Begin()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to start database transaction: "+err.Error())
	}
	
	// Generate transfer ID
	transferID := "tran-" + time.Now().Format("20060102150405")
	now := time.Now()
	
	// If transferring all, remove the batch from the farm
	if req.Quantity == currentQuantity {
		_, err = tx.Exec(`
			DELETE FROM farm_batches 
			WHERE farm_id = $1 AND batch_id = $2
		`, farmID, req.BatchID)
		if err != nil {
			tx.Rollback()
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to remove batch from farm: "+err.Error())
		}
	} else {
		// Otherwise, update the quantity
		_, err = tx.Exec(`
			UPDATE farm_batches 
			SET quantity = quantity - $1 
			WHERE farm_id = $2 AND batch_id = $3
		`, req.Quantity, farmID, req.BatchID)
		if err != nil {
			tx.Rollback()
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to update batch quantity: "+err.Error())
		}
	}
	
	// Create transfer record
	_, err = tx.Exec(`
		INSERT INTO batch_transfers (
			id, batch_id, source_id, destination, quantity, 
			transferred_at, transferred_by, transfer_notes, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		transferID,
		req.BatchID,
		farmID,
		req.Destination,
		req.Quantity,
		now,
		userID,
		req.TransferNotes,
		req.Metadata,
	)
	if err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create transfer record: "+err.Error())
	}
	
	// Update batch status
	_, err = tx.Exec(`
		UPDATE batches 
		SET status = 'in_transit', updated_at = $1
		WHERE id = $2
	`, now, req.BatchID)
	if err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update batch status: "+err.Error())
	}
	
	// Create batch event
	_, err = tx.Exec(`
		INSERT INTO events (id, batch_id, event_type, location, actor_id, timestamp, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, 
		"evt-" + time.Now().Format("20060102150405"), 
		req.BatchID, 
		"batch_transferred", 
		farmID, 
		userID, 
		now, 
		req.Metadata,
	)
	if err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create batch event: "+err.Error())
	}
	
	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
	}
	
	// Record on blockchain
	cfg := config.GetConfig()
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	_, err = blockchainClient.SubmitTransaction("BATCH_TRANSFERRED", map[string]interface{}{
		"transfer_id":     transferID,
		"batch_id":        req.BatchID,
		"source_id":       farmID,
		"destination":     req.Destination,
		"quantity":        req.Quantity,
		"transferred_by":  userID,
		"transfer_notes":  req.TransferNotes,
		"metadata":        req.Metadata,
		"timestamp":       now,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record transfer on blockchain: "+err.Error())
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch transferred successfully from farm",
		Data: map[string]interface{}{
			"transfer_id":     transferID,
			"batch_id":        req.BatchID,
			"source_id":       farmID,
			"destination":     req.Destination,
			"quantity":        req.Quantity,
			"transferred_at":  now,
			"transferred_by":  userID,
		},
	})
}

// GetFarmStats gets statistics for a farm
// @Summary Get farm statistics
// @Description Retrieve statistics for a farming facility
// @Tags farms
// @Accept json
// @Produce json
// @Param farmId path string true "Farm ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /farms/{farmId}/stats [get]
func GetFarmStats(c *fiber.Ctx) error {
	// Get farm ID from path
	farmID := c.Params("farmId")
	if farmID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Farm ID is required")
	}
	
	// Check if farm exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM farms WHERE id = $1)", farmID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Farm not found")
	}
	
	// Get statistics
	var totalBatches int
	var activeBatches int
	var totalQuantity int
	var batchesLastMonth int
	
	// Total batches
	err = db.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM farm_batches 
		WHERE farm_id = $1
	`, farmID).Scan(&totalBatches)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get farm statistics: "+err.Error())
	}
	
	// Active batches
	err = db.DB.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(quantity), 0)
		FROM farm_batches fb
		JOIN batches b ON fb.batch_id = b.id
		WHERE fb.farm_id = $1 AND b.is_active = true
	`, farmID).Scan(&activeBatches, &totalQuantity)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get farm statistics: "+err.Error())
	}
	
	// Batches received in the last month
	err = db.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM farm_batches 
		WHERE farm_id = $1 AND received_at > NOW() - INTERVAL '30 days'
	`, farmID).Scan(&batchesLastMonth)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get farm statistics: "+err.Error())
	}
	
	// Get batch distribution by species
	rows, err := db.DB.Query(`
		SELECT b.species, COUNT(*), SUM(fb.quantity)
		FROM farm_batches fb
		JOIN batches b ON fb.batch_id = b.id
		WHERE fb.farm_id = $1 AND b.is_active = true
		GROUP BY b.species
	`, farmID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get species distribution: "+err.Error())
	}
	defer rows.Close()
	
	speciesDistribution := make(map[string]map[string]interface{})
	for rows.Next() {
		var species string
		var count int
		var quantity int
		
		err := rows.Scan(&species, &count, &quantity)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error parsing species data: "+err.Error())
		}
		
		speciesDistribution[species] = map[string]interface{}{
			"batch_count": count,
			"quantity":    quantity,
		}
	}
	
	// Get recent records
	rows, err = db.DB.Query(`
		SELECT id, record_type, recorded_at
		FROM farming_records
		WHERE farm_id = $1
		ORDER BY recorded_at DESC
		LIMIT 5
	`, farmID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get recent records: "+err.Error())
	}
	defer rows.Close()
	
	recentRecords := []map[string]interface{}{}
	for rows.Next() {
		var id string
		var recordType string
		var recordedAt time.Time
		
		err := rows.Scan(&id, &recordType, &recordedAt)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error parsing record data: "+err.Error())
		}
		
		recentRecords = append(recentRecords, map[string]interface{}{
			"id":          id,
			"record_type": recordType,
			"recorded_at": recordedAt,
		})
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Farm statistics retrieved successfully",
		Data: map[string]interface{}{
			"total_batches":        totalBatches,
			"active_batches":       activeBatches,
			"total_quantity":       totalQuantity,
			"batches_last_month":   batchesLastMonth,
			"species_distribution": speciesDistribution,
			"recent_records":       recentRecords,
		},
	})
}