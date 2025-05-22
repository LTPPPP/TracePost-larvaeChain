package api

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"github.com/LTPPPP/TracePost-larvaeChain/models"
)

// CreateCompanyRequest represents a request to create a new company
type CreateCompanyRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Location    string `json:"location"`
	ContactInfo string `json:"contact_info"`
}

// UpdateCompanyRequest represents a request to update a company
type UpdateCompanyRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Location    string `json:"location"`
	ContactInfo string `json:"contact_info"`
}

// GetAllCompanies returns all companies
// @Summary Get all companies
// @Description Retrieve all companies in the system
// @Tags companies
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse{data=[]models.Company}
// @Failure 500 {object} ErrorResponse
// @Router /companies [get]
func GetAllCompanies(c *fiber.Ctx) error {
	// Initialize companies slice
	var companies []models.Company

	// Get all companies from the database
	query := `
		SELECT id, name, type, location, contact_info, created_at, updated_at, is_active
		FROM company
		WHERE is_active = true
		ORDER BY name ASC
	`
	rows, err := db.DB.Query(query)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Iterate through rows and build companies slice
	for rows.Next() {
		var company models.Company
		err := rows.Scan(
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
			return fiber.NewError(fiber.StatusInternalServerError, "Error parsing company data")
		}
		companies = append(companies, company)
	}

	// Return success response with companies data
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Success: true,
		Message: "Companies retrieved successfully",
		Data:    companies,
	})
}

// GetCompanyByID returns a company by ID
// @Summary Get company by ID
// @Description Retrieve a company by its ID
// @Tags companies
// @Accept json
// @Produce json
// @Param companyId path int true "Company ID"
// @Success 200 {object} SuccessResponse{data=models.Company}
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /companies/{companyId} [get]
func GetCompanyByID(c *fiber.Ctx) error {
	// Parse company ID from parameters
	companyID, err := strconv.Atoi(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid company ID")
	}

	// Get company from database
	var company models.Company
	query := `
		SELECT id, name, type, location, contact_info, created_at, updated_at, is_active
		FROM company
		WHERE id = $1 AND is_active = true
	`
	err = db.DB.QueryRow(query, companyID).Scan(
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
		return fiber.NewError(fiber.StatusNotFound, "Company not found")
	}
	// Get hatcheries for this company
	hatcheriesQuery := `
		SELECT id, name, company_id, created_at, updated_at, is_active
		FROM hatchery
		WHERE company_id = $1 AND is_active = true
	`
	hatcheryRows, err := db.DB.Query(hatcheriesQuery, companyID)
	if err == nil {
		defer hatcheryRows.Close()
		
		var hatcheries []models.Hatchery
		for hatcheryRows.Next() {
			var hatchery models.Hatchery			
			err := hatcheryRows.Scan(
				&hatchery.ID,
				&hatchery.Name,
				&hatchery.CompanyID,
				&hatchery.CreatedAt,
				&hatchery.UpdatedAt,
				&hatchery.IsActive,
			)
			if err == nil {
				hatcheries = append(hatcheries, hatchery)
			}
		}
		company.Hatcheries = hatcheries
	}

	// Return success response with company data
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Success: true,
		Message: "Company retrieved successfully",
		Data:    company,
	})
}

// CreateCompany creates a new company
// @Summary Create a new company
// @Description Create a new company in the system
// @Tags companies
// @Accept json
// @Produce json
// @Param request body CreateCompanyRequest true "Company creation details"
// @Success 201 {object} SuccessResponse{data=models.Company}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /companies [post]
func CreateCompany(c *fiber.Ctx) error {
	// Parse request body
	var req CreateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate input
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Company name is required")
	}

	// Initialize blockchain client for blockchain record
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Create company record in database
	var company models.Company
	company.Name = req.Name
	company.Type = req.Type
	company.Location = req.Location
	company.ContactInfo = req.ContactInfo
	company.IsActive = true

	query := `
		INSERT INTO company (name, type, location, contact_info, created_at, updated_at, is_active)
		VALUES ($1, $2, $3, $4, NOW(), NOW(), $5)
		RETURNING id, created_at, updated_at
	`
	err := db.DB.QueryRow(
		query,
		company.Name,
		company.Type, 
		company.Location,
		company.ContactInfo,
		company.IsActive,
	).Scan(
		&company.ID,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create company")
	}

	// Record company creation on blockchain using custom transaction
	companyData := map[string]interface{}{
		"company_id":   strconv.Itoa(company.ID),
		"name":         company.Name,
		"type":         company.Type,
		"location":     company.Location,
		"contact_info": company.ContactInfo,
		"created_at":   company.CreatedAt.Format(time.RFC3339),
	}
	
	txID, err := blockchainClient.SubmitGenericTransaction("CREATE_COMPANY", companyData)

	// If blockchain recording is successful, save the blockchain record
	if err == nil && txID != "" {
		_, _ = db.DB.Exec(
			"INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active) VALUES ($1, $2, $3, $4, NOW(), NOW(), true)",
			"company",
			company.ID,
			txID,
			"", // Metadata hash would be calculated in a real implementation
		)
	}

	// Return success response with company data
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Company created successfully",
		Data:    company,
	})
}

// UpdateCompany updates an existing company
// @Summary Update an existing company
// @Description Update an existing company in the system
// @Tags companies
// @Accept json
// @Produce json
// @Param companyId path int true "Company ID"
// @Param request body UpdateCompanyRequest true "Company update details"
// @Success 200 {object} SuccessResponse{data=models.Company}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /companies/{companyId} [put]
func UpdateCompany(c *fiber.Ctx) error {
	// Parse company ID from parameters
	companyID, err := strconv.Atoi(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid company ID")
	}

	// Parse request body
	var req UpdateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Verify company exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM company WHERE id = $1 AND is_active = true)", companyID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Company not found")
	}

	// Get existing company data
	var company models.Company
	query := `
		SELECT id, name, type, location, contact_info, created_at, updated_at, is_active
		FROM company
		WHERE id = $1 AND is_active = true
	`
	err = db.DB.QueryRow(query, companyID).Scan(
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
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Update company fields if provided
	if req.Name != "" {
		company.Name = req.Name
	}
	if req.Type != "" {
		company.Type = req.Type
	}
	if req.Location != "" {
		company.Location = req.Location
	}
	if req.ContactInfo != "" {
		company.ContactInfo = req.ContactInfo
	}

	// Update company in database
	updateQuery := `
		UPDATE company
		SET name = $1, type = $2, location = $3, contact_info = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`
	err = db.DB.QueryRow(
		updateQuery,
		company.Name,
		company.Type,
		company.Location,
		company.ContactInfo,
		company.ID,
	).Scan(&company.UpdatedAt)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update company")
	}

	// Initialize blockchain client for blockchain record
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Record company update on blockchain using custom transaction
	companyData := map[string]interface{}{
		"company_id":   strconv.Itoa(company.ID),
		"name":         company.Name,
		"type":         company.Type,
		"location":     company.Location,
		"contact_info": company.ContactInfo,
		"updated_at":   company.UpdatedAt.Format(time.RFC3339),
	}
	
	txID, err := blockchainClient.SubmitGenericTransaction("UPDATE_COMPANY", companyData)

	// If blockchain recording is successful, save the blockchain record
	if err == nil && txID != "" {
		_, _ = db.DB.Exec(
			"INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active) VALUES ($1, $2, $3, $4, NOW(), NOW(), true)",
			"company",
			company.ID,
			txID,
			"", // Metadata hash would be calculated in a real implementation
		)
	}

	// Return success response with updated company data
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Success: true,
		Message: "Company updated successfully",
		Data:    company,
	})
}

// DeleteCompany soft-deletes a company by ID
// @Summary Delete a company
// @Description Soft-delete a company by its ID
// @Tags companies
// @Accept json
// @Produce json
// @Param companyId path int true "Company ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /companies/{companyId} [delete]
func DeleteCompany(c *fiber.Ctx) error {
	// Parse company ID from parameters
	companyID, err := strconv.Atoi(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid company ID")
	}

	// Verify company exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM company WHERE id = $1 AND is_active = true)", companyID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Company not found")
	}

	// Soft delete the company (set is_active to false)
	_, err = db.DB.Exec(
		"UPDATE company SET is_active = false, updated_at = NOW() WHERE id = $1",
		companyID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete company")
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Record company deletion on blockchain using custom transaction
	deleteData := map[string]interface{}{
		"company_id":     strconv.Itoa(companyID),
		"status":         "inactive",
		"deactivated_at": time.Now().Format(time.RFC3339),
	}
	
	txID, err := blockchainClient.SubmitGenericTransaction("DELETE_COMPANY", deleteData)

	// If blockchain recording is successful, save the blockchain record
	if err == nil && txID != "" {
		_, _ = db.DB.Exec(
			"INSERT INTO blockchain_record (related_table, related_id, tx_id, metadata_hash, created_at, updated_at, is_active) VALUES ($1, $2, $3, $4, NOW(), NOW(), true)",
			"company",
			companyID,
			txID,
			"", // Metadata hash would be calculated in a real implementation
		)
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Success: true,
		Message: "Company deleted successfully",
	})
}

// GetCompanyHatcheries returns all hatcheries for a specific company
// @Summary Get company hatcheries
// @Description Retrieve all hatcheries for a specific company
// @Tags companies
// @Accept json
// @Produce json
// @Param companyId path int true "Company ID"
// @Success 200 {object} SuccessResponse{data=[]models.Hatchery}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /companies/{companyId}/hatcheries [get]
func GetCompanyHatcheries(c *fiber.Ctx) error {
	// Parse company ID from parameters
	companyID, err := strconv.Atoi(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid company ID")
	}

	// Verify company exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM company WHERE id = $1 AND is_active = true)", companyID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Company not found")
	}

	// Get hatcheries for this company
	var hatcheries []models.Hatchery
	query := `
		SELECT id, name, location, contact, company_id, created_at, updated_at, is_active
		FROM hatchery
		WHERE company_id = $1 AND is_active = true
		ORDER BY name ASC
	`
	rows, err := db.DB.Query(query, companyID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Iterate through rows and build hatcheries slice
	for rows.Next() {
		var hatchery models.Hatchery		
		err := rows.Scan(
			&hatchery.ID,
			&hatchery.Name,
			&hatchery.CompanyID,
			&hatchery.CreatedAt,
			&hatchery.UpdatedAt,
			&hatchery.IsActive,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error parsing hatchery data")
		}
		hatcheries = append(hatcheries, hatchery)
	}

	// Return success response with hatcheries data
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Success: true,
		Message: "Company hatcheries retrieved successfully",
		Data:    hatcheries,
	})
}

// GetCompanyStats returns statistics for a company
// @Summary Get company statistics
// @Description Retrieve statistics for a specific company
// @Tags companies
// @Accept json
// @Produce json
// @Param companyId path int true "Company ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /companies/{companyId}/stats [get]
func GetCompanyStats(c *fiber.Ctx) error {
	// Parse company ID from parameters
	companyID, err := strconv.Atoi(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid company ID")
	}

	// Verify company exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM company WHERE id = $1 AND is_active = true)", companyID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Company not found")
	}

	// Get company statistics
	var stats struct {
		TotalHatcheries int `json:"total_hatcheries"`
		TotalBatches    int `json:"total_batches"`
		TotalEvents     int `json:"total_events"`
		TotalUsers      int `json:"total_users"`
	}

	// Count hatcheries
	err = db.DB.QueryRow("SELECT COUNT(*) FROM hatchery WHERE company_id = $1 AND is_active = true", companyID).Scan(&stats.TotalHatcheries)
	if err != nil {
		stats.TotalHatcheries = 0
	}

	// Count batches (from all company hatcheries)
	err = db.DB.QueryRow(`
		SELECT COUNT(b.id) 
		FROM batch b 
		JOIN hatchery h ON b.hatchery_id = h.id 
		WHERE h.company_id = $1 AND b.is_active = true AND h.is_active = true
	`, companyID).Scan(&stats.TotalBatches)
	if err != nil {
		stats.TotalBatches = 0
	}

	// Count events (from all company batches)
	err = db.DB.QueryRow(`
		SELECT COUNT(e.id) 
		FROM event e 
		JOIN batch b ON e.batch_id = b.id 
		JOIN hatchery h ON b.hatchery_id = h.id 
		WHERE h.company_id = $1 AND e.is_active = true AND b.is_active = true AND h.is_active = true
	`, companyID).Scan(&stats.TotalEvents)
	if err != nil {
		stats.TotalEvents = 0
	}

	// Count users
	err = db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE company_id = $1 AND is_active = true", companyID).Scan(&stats.TotalUsers)
	if err != nil {
		stats.TotalUsers = 0
	}

	// Return success response with statistics
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Success: true,
		Message: "Company statistics retrieved successfully",
		Data:    stats,
	})
}