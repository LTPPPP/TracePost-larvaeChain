package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vietchain/tracepost-larvae/blockchain"
	"github.com/vietchain/tracepost-larvae/config"
	"github.com/vietchain/tracepost-larvae/db"
	"time"
)

// ComplianceStandard represents a regulatory compliance standard
type ComplianceStandard struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Issuer      string    `json:"issuer"`
	ValidFrom   time.Time `json:"valid_from"`
	ValidTo     time.Time `json:"valid_to"`
	Requirements []string  `json:"requirements"`
	Region      string    `json:"region"`
}

// ComplianceCheckRequest represents a request to validate against a standard
type ComplianceCheckRequest struct {
	BatchID     string `json:"batch_id"`
	StandardID  string `json:"standard_id"`
}

// ComplianceCheckResult represents the result of a compliance check
type ComplianceCheckResult struct {
	BatchID          string                 `json:"batch_id"`
	StandardID       string                 `json:"standard_id"`
	StandardName     string                 `json:"standard_name"`
	IsCompliant      bool                   `json:"is_compliant"`
	ComplianceScore  float64                `json:"compliance_score"`
	CheckedAt        string                 `json:"checked_at"`
	Issues           []ComplianceIssue      `json:"issues,omitempty"`
	RequirementsMet  map[string]interface{} `json:"requirements_met"`
}

// ComplianceIssue represents a compliance issue found during a check
type ComplianceIssue struct {
	Requirement  string `json:"requirement"`
	Description  string `json:"description"`
	Severity     string `json:"severity"` // "critical", "major", "minor"
	Recommendation string `json:"recommendation"`
}

// ComplianceReport represents a detailed compliance report
type ComplianceReport struct {
	BatchID         string                 `json:"batch_id"`
	GeneratedAt     string                 `json:"generated_at"`
	Standards       []string               `json:"standards"`
	OverallStatus   string                 `json:"overall_status"`
	Details         map[string]interface{} `json:"details"`
	RecommendedActions []string            `json:"recommended_actions,omitempty"`
}

// CheckBatchCompliance checks if a batch complies with regulatory standards
// @Summary Check batch compliance
// @Description Check if a batch complies with regulatory standards (EU DR, US FDA, ASC, etc.)
// @Tags compliance
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=ComplianceCheckResult}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /compliance/check/{batchId} [get]
func CheckBatchCompliance(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Get batch ID from path
	batchID := c.Params("batchId")
	if batchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	// Get standard ID from query
	standardID := c.Query("standard")
	if standardID == "" {
		// Default to EU DR standard if not specified
		standardID = "eu-dr-2022"
	}
	
	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}
	
	// Get standard details
	var standard ComplianceStandard
	err = db.DB.QueryRow(`
		SELECT id, name, description, version, issuer, valid_from, valid_to, requirements, region
		FROM compliance_standards
		WHERE id = $1
	`, standardID).Scan(
		&standard.ID,
		&standard.Name,
		&standard.Description,
		&standard.Version,
		&standard.Issuer,
		&standard.ValidFrom,
		&standard.ValidTo,
		&standard.Requirements,
		&standard.Region,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Compliance standard not found")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Check batch compliance
	// In a real implementation, this would analyze batch data against standard requirements
	// For this example, we'll simulate a compliance check
	
	// Get batch data from blockchain
	batchData, err := blockchainClient.GetBatchData(batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get batch data: "+err.Error())
	}
	
	// Perform compliance check
	result := performComplianceCheck(batchID, standard, batchData)
	
	// Save compliance check result to database
	_, err = db.DB.Exec(`
		INSERT INTO compliance_checks (batch_id, standard_id, is_compliant, compliance_score, checked_at, issues, requirements_met)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		result.BatchID,
		result.StandardID,
		result.IsCompliant,
		result.ComplianceScore,
		time.Now(),
		result.Issues,
		result.RequirementsMet,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save compliance check result: "+err.Error())
	}
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch compliance check completed",
		Data:    result,
	})
}

// GenerateComplianceReport generates a detailed compliance report for a batch
// @Summary Generate compliance report
// @Description Generate a detailed compliance report for a batch against multiple standards
// @Tags compliance
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=ComplianceReport}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /compliance/report/{batchId} [get]
func GenerateComplianceReport(c *fiber.Ctx) error {
	// Get batch ID from path
	batchID := c.Params("batchId")
	if batchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}
	
	// Get all compliance check results for this batch
	rows, err := db.DB.Query(`
		SELECT c.batch_id, c.standard_id, c.is_compliant, c.compliance_score, c.checked_at, c.issues, c.requirements_met,
		       s.name as standard_name
		FROM compliance_checks c
		JOIN compliance_standards s ON c.standard_id = s.id
		WHERE c.batch_id = $1
		ORDER BY c.checked_at DESC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()
	
	// Process results and generate report
	var report ComplianceReport
	report.BatchID = batchID
	report.GeneratedAt = time.Now().Format(time.RFC3339)
	report.Details = make(map[string]interface{})
	
	var standards []string
	var overallCompliant = true
	var checkResults []ComplianceCheckResult
	
	for rows.Next() {
		var result ComplianceCheckResult
		var standardName string
		var checkedAt time.Time
		
		err := rows.Scan(
			&result.BatchID,
			&result.StandardID,
			&result.IsCompliant,
			&result.ComplianceScore,
			&checkedAt,
			&result.Issues,
			&result.RequirementsMet,
			&standardName,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error parsing compliance check data")
		}
		
		result.StandardName = standardName
		result.CheckedAt = checkedAt.Format(time.RFC3339)
		
		standards = append(standards, result.StandardID)
		report.Details[result.StandardID] = result
		
		// If any standard is not compliant, the overall status is not compliant
		if !result.IsCompliant {
			overallCompliant = false
		}
		
		checkResults = append(checkResults, result)
	}
	
	report.Standards = standards
	if overallCompliant {
		report.OverallStatus = "compliant"
	} else {
		report.OverallStatus = "non-compliant"
		
		// Generate recommended actions for non-compliant standards
		for _, result := range checkResults {
			if !result.IsCompliant {
				for _, issue := range result.Issues {
					report.RecommendedActions = append(report.RecommendedActions, 
						"For "+result.StandardName+": "+issue.Recommendation)
				}
			}
		}
	}
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Compliance report generated successfully",
		Data:    report,
	})
}

// ListComplianceStandards lists all available compliance standards
// @Summary List compliance standards
// @Description List all available compliance standards
// @Tags compliance
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Router /compliance/standards [get]
func ListComplianceStandards(c *fiber.Ctx) error {
	// Get all compliance standards from database
	rows, err := db.DB.Query(`
		SELECT id, name, description, version, issuer, valid_from, valid_to, requirements, region
		FROM compliance_standards
		ORDER BY name
	`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()
	
	// Process standards
	var standards []ComplianceStandard
	
	for rows.Next() {
		var standard ComplianceStandard
		err := rows.Scan(
			&standard.ID,
			&standard.Name,
			&standard.Description,
			&standard.Version,
			&standard.Issuer,
			&standard.ValidFrom,
			&standard.ValidTo,
			&standard.Requirements,
			&standard.Region,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error parsing standards data")
		}
		
		standards = append(standards, standard)
	}
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Compliance standards retrieved successfully",
		Data:    standards,
	})
}

// ValidateAgainstStandard validates a batch against a specific standard
// @Summary Validate against standard
// @Description Validate a batch against a specific regulatory standard
// @Tags compliance
// @Accept json
// @Produce json
// @Param request body ComplianceCheckRequest true "Validation request"
// @Success 200 {object} SuccessResponse{data=ComplianceCheckResult}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /compliance/validate [post]
func ValidateAgainstStandard(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Parse request
	var req ComplianceCheckRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.BatchID == "" || req.StandardID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID and Standard ID are required")
	}
	
	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1)", req.BatchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}
	
	// Get standard details
	var standard ComplianceStandard
	err = db.DB.QueryRow(`
		SELECT id, name, description, version, issuer, valid_from, valid_to, requirements, region
		FROM compliance_standards
		WHERE id = $1
	`, req.StandardID).Scan(
		&standard.ID,
		&standard.Name,
		&standard.Description,
		&standard.Version,
		&standard.Issuer,
		&standard.ValidFrom,
		&standard.ValidTo,
		&standard.Requirements,
		&standard.Region,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Compliance standard not found")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Get batch data from blockchain
	batchData, err := blockchainClient.GetBatchData(req.BatchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get batch data: "+err.Error())
	}
	
	// Perform compliance check
	result := performComplianceCheck(req.BatchID, standard, batchData)
	
	// Save compliance check result to database
	_, err = db.DB.Exec(`
		INSERT INTO compliance_checks (batch_id, standard_id, is_compliant, compliance_score, checked_at, issues, requirements_met)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		result.BatchID,
		result.StandardID,
		result.IsCompliant,
		result.ComplianceScore,
		time.Now(),
		result.Issues,
		result.RequirementsMet,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save compliance check result: "+err.Error())
	}
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Validation against standard completed successfully",
		Data:    result,
	})
}

// performComplianceCheck performs a compliance check of batch data against a standard
// This is a simplified implementation for demonstration purposes
func performComplianceCheck(batchID string, standard ComplianceStandard, batchData map[string]interface{}) ComplianceCheckResult {
	// Initialize result
	result := ComplianceCheckResult{
		BatchID:         batchID,
		StandardID:      standard.ID,
		StandardName:    standard.Name,
		CheckedAt:       time.Now().Format(time.RFC3339),
		RequirementsMet: make(map[string]interface{}),
	}
	
	// For each requirement in the standard, check if batch data satisfies it
	// In a real implementation, this would be a complex logic specific to each standard
	// For this example, we'll simulate a compliance check
	
	// Example checks for EU DR (European Delegated Regulation)
	if standard.ID == "eu-dr-2022" {
		// Check for batch origin documentation
		if origin, ok := batchData["origin"].(map[string]interface{}); ok {
			if _, hasLocation := origin["location"]; hasLocation {
				result.RequirementsMet["origin_location"] = true
			} else {
				result.Issues = append(result.Issues, ComplianceIssue{
					Requirement:     "Origin location",
					Description:     "The batch origin location is not documented",
					Severity:        "major",
					Recommendation:  "Add batch origin location information",
				})
				result.RequirementsMet["origin_location"] = false
			}
		}
		
		// Check for health certificate
		if documents, ok := batchData["documents"].([]interface{}); ok {
			hasHealthCert := false
			for _, doc := range documents {
				if docMap, ok := doc.(map[string]interface{}); ok {
					if docType, ok := docMap["type"].(string); ok && docType == "health_certificate" {
						hasHealthCert = true
						break
					}
				}
			}
			
			if hasHealthCert {
				result.RequirementsMet["health_certificate"] = true
			} else {
				result.Issues = append(result.Issues, ComplianceIssue{
					Requirement:     "Health certificate",
					Description:     "No health certificate found for the batch",
					Severity:        "critical",
					Recommendation:  "Upload the required health certificate",
				})
				result.RequirementsMet["health_certificate"] = false
			}
		}
		
		// Check for antibiotic usage
		if treatments, ok := batchData["treatments"].([]interface{}); ok {
			hasAntibiotics := false
			for _, treatment := range treatments {
				if treatmentMap, ok := treatment.(map[string]interface{}); ok {
					if treatmentType, ok := treatmentMap["type"].(string); ok && treatmentType == "antibiotic" {
						hasAntibiotics = true
						break
					}
				}
			}
			
			if !hasAntibiotics {
				result.RequirementsMet["antibiotic_free"] = true
			} else {
				result.Issues = append(result.Issues, ComplianceIssue{
					Requirement:     "Antibiotic usage",
					Description:     "The batch has been treated with antibiotics",
					Severity:        "minor",
					Recommendation:  "For EU export, ensure antibiotics are used according to EU regulations",
				})
				result.RequirementsMet["antibiotic_free"] = false
			}
		}
	}
	
	// Example checks for US FDA
	if standard.ID == "us-fda-seafood" {
		// Check for HACCP compliance
		if haccp, ok := batchData["haccp_compliant"].(bool); ok && haccp {
			result.RequirementsMet["haccp_compliant"] = true
		} else {
			result.Issues = append(result.Issues, ComplianceIssue{
				Requirement:     "HACCP compliance",
				Description:     "No evidence of HACCP compliance",
				Severity:        "critical",
				Recommendation:  "Implement and document HACCP procedures",
			})
			result.RequirementsMet["haccp_compliant"] = false
		}
		
		// Check for contaminant testing
		if tests, ok := batchData["tests"].([]interface{}); ok {
			hasContaminantTest := false
			for _, test := range tests {
				if testMap, ok := test.(map[string]interface{}); ok {
					if testType, ok := testMap["type"].(string); ok && testType == "contaminant" {
						hasContaminantTest = true
						break
					}
				}
			}
			
			if hasContaminantTest {
				result.RequirementsMet["contaminant_testing"] = true
			} else {
				result.Issues = append(result.Issues, ComplianceIssue{
					Requirement:     "Contaminant testing",
					Description:     "No contaminant testing results found",
					Severity:        "major",
					Recommendation:  "Perform and document required contaminant testing",
				})
				result.RequirementsMet["contaminant_testing"] = false
			}
		}
	}
	
	// Example checks for ASC (Aquaculture Stewardship Council)
	if standard.ID == "asc-shrimp" {
		// Check for environmental impact assessment
		if eia, ok := batchData["environmental_impact_assessment"].(bool); ok && eia {
			result.RequirementsMet["environmental_impact"] = true
		} else {
			result.Issues = append(result.Issues, ComplianceIssue{
				Requirement:     "Environmental impact assessment",
				Description:     "No environmental impact assessment found",
				Severity:        "major",
				Recommendation:  "Conduct and document an environmental impact assessment",
			})
			result.RequirementsMet["environmental_impact"] = false
		}
		
		// Check for feed sustainability
		if feed, ok := batchData["feed"].(map[string]interface{}); ok {
			if sustainable, ok := feed["sustainable"].(bool); ok && sustainable {
				result.RequirementsMet["sustainable_feed"] = true
			} else {
				result.Issues = append(result.Issues, ComplianceIssue{
					Requirement:     "Sustainable feed",
					Description:     "No evidence of sustainable feed usage",
					Severity:        "major",
					Recommendation:  "Use and document sustainable feed sources",
				})
				result.RequirementsMet["sustainable_feed"] = false
			}
		}
	}
	
	// Calculate compliance score
	totalRequirements := len(result.RequirementsMet)
	if totalRequirements > 0 {
		metCount := 0
		for _, met := range result.RequirementsMet {
			if metBool, ok := met.(bool); ok && metBool {
				metCount++
			}
		}
		result.ComplianceScore = float64(metCount) / float64(totalRequirements) * 100
	}
	
	// Determine overall compliance
	result.IsCompliant = len(result.Issues) == 0
	
	return result
}