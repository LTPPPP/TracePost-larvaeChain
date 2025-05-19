package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/config"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"time"
)

// AllianceMember represents a member of the industry alliance
type AllianceMember struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	DID            string    `json:"did"`
	MemberType     string    `json:"member_type"` // "producer", "processor", "exporter", "regulator", etc.
	JoinDate       time.Time `json:"join_date"`
	Status         string    `json:"status"` // "active", "pending", "suspended"
	Location       string    `json:"location,omitempty"`
	Website        string    `json:"website,omitempty"`
	ContactDetails string    `json:"contact_details,omitempty"`
}

// ShareDataRequest represents a request to share data with alliance members
type ShareDataRequest struct {
	BatchID      string   `json:"batch_id"`
	DataType     string   `json:"data_type"` // "origin", "quality", "certification", etc.
	Recipients   []string `json:"recipients,omitempty"` // Member IDs, if empty share with all members
	Description  string   `json:"description,omitempty"`
	Permissions  string   `json:"permissions"` // "read_only", "read_write", etc.
}

// JoinAllianceRequest represents a request to join the industry alliance
type JoinAllianceRequest struct {
	Name           string                 `json:"name"`
	DID            string                 `json:"did"`
	MemberType     string                 `json:"member_type"`
	Location       string                 `json:"location,omitempty"`
	Website        string                 `json:"website,omitempty"`
	ContactDetails string                 `json:"contact_details,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// SharedDataResponse represents a response for shared data
type SharedDataResponse struct {
	ID           string                 `json:"id"`
	BatchID      string                 `json:"batch_id"`
	DataType     string                 `json:"data_type"`
	SharedBy     string                 `json:"shared_by"`
	SharedWith   []string               `json:"shared_with"`
	SharedAt     string                 `json:"shared_at"`
	Description  string                 `json:"description,omitempty"`
	Permissions  string                 `json:"permissions"`
	AccessURL    string                 `json:"access_url,omitempty"`
	AccessToken  string                 `json:"access_token,omitempty"`
	DataSummary  map[string]interface{} `json:"data_summary,omitempty"`
}

// ShareDataWithAlliance shares batch data with alliance members
// @Summary Share data with alliance
// @Description Share batch data with industry alliance members
// @Tags alliance
// @Accept json
// @Produce json
// @Param request body ShareDataRequest true "Data sharing details"
// @Success 201 {object} SuccessResponse{data=SharedDataResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /alliance/share [post]
func ShareDataWithAlliance(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Parse request
	var req ShareDataRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.BatchID == "" || req.DataType == "" || req.Permissions == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID, data type, and permissions are required")
	}
	
	// Get user details from token
	userDID := c.Locals("user_did").(string)
	if userDID == "" {
		return fiber.NewError(fiber.StatusForbidden, "User not authenticated")
	}
	
	// Check if batch exists and user has permission to access it
	var exists bool
	var batchOwnerID string
	err := db.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1),
		       (SELECT owner_id FROM batches WHERE batch_id = $1)
	`, req.BatchID).Scan(&exists, &batchOwnerID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}
	
	// Check if user has permission to share this batch
	var userID string
	err = db.DB.QueryRow("SELECT id FROM account WHERE did = $1", userDID).Scan(&userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve user information")
	}
	
	if userID != batchOwnerID {
		// Check if user has been delegated permission
		var hasPermission bool
		err = db.DB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM batch_permissions 
				WHERE batch_id = $1 AND user_id = $2 AND permission_type = 'share'
			)
		`, req.BatchID, userID).Scan(&hasPermission)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Database error")
		}
		if !hasPermission {
			return fiber.NewError(fiber.StatusForbidden, "You don't have permission to share this batch")
		}
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Prepare recipients
	var recipients []string
	if len(req.Recipients) == 0 {
		// Share with all active alliance members
		rows, err := db.DB.Query(`
			SELECT id FROM alliance_members
			WHERE status = 'active'
		`)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Database error")
		}
		defer rows.Close()
		
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Error parsing alliance members")
			}
			recipients = append(recipients, id)
		}
	} else {
		// Validate that all specified recipients are active alliance members
		for _, recipientID := range req.Recipients {
			var exists bool
			var status string
			err := db.DB.QueryRow(`
				SELECT EXISTS(SELECT 1 FROM alliance_members WHERE id = $1),
				       (SELECT status FROM alliance_members WHERE id = $1)
			`, recipientID).Scan(&exists, &status)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Database error")
			}
			if !exists {
				return fiber.NewError(fiber.StatusBadRequest, "Recipient "+recipientID+" is not an alliance member")
			}
			if status != "active" {
				return fiber.NewError(fiber.StatusBadRequest, "Recipient "+recipientID+" is not an active alliance member")
			}
			
			recipients = append(recipients, recipientID)
		}
	}
	
	// Generate sharing ID
	shareID := "share-" + req.BatchID + "-" + time.Now().Format("20060102150405")
	
	// Record on blockchain
	txID, err := blockchainClient.SubmitTransaction("ALLIANCE_SHARE", map[string]interface{}{
		"share_id":     shareID,
		"batch_id":     req.BatchID,
		"data_type":    req.DataType,
		"shared_by":    userDID,
		"shared_with":  recipients,
		"permissions":  req.Permissions,
		"description":  req.Description,
		"timestamp":    time.Now(),
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record share on blockchain: "+err.Error())
	}
	
	// Record in database
	now := time.Now()
	_, err = db.DB.Exec(`
		INSERT INTO shared_data (
			id, batch_id, data_type, shared_by, shared_with, permissions, 
			description, shared_at, blockchain_tx_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		shareID,
		req.BatchID,
		req.DataType,
		userID,
		recipients,
		req.Permissions,
		req.Description,
		now,
		txID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record share in database: "+err.Error())
	}
	
	// Generate access URL and token
	accessURL := cfg.BaseURL + "/api/v1/alliance/data/" + shareID
	accessToken := generateAccessToken(shareID, recipients)
	
	// Get batch data summary
	batchData, err := blockchainClient.GetBatchData(req.BatchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve batch data: "+err.Error())
	}
	
	// Create data summary based on requested data type
	dataSummary := extractDataSummary(batchData, req.DataType)
	
	// Return response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Data shared successfully with alliance members",
		Data: SharedDataResponse{
			ID:          shareID,
			BatchID:     req.BatchID,
			DataType:    req.DataType,
			SharedBy:    userID,
			SharedWith:  recipients,
			SharedAt:    now.Format(time.RFC3339),
			Description: req.Description,
			Permissions: req.Permissions,
			AccessURL:   accessURL,
			AccessToken: accessToken,
			DataSummary: dataSummary,
		},
	})
}

// ListAllianceMembers lists all members of the industry alliance
// @Summary List alliance members
// @Description List all members of the industry alliance
// @Tags alliance
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse{data=[]AllianceMember}
// @Failure 500 {object} ErrorResponse
// @Router /alliance/members [get]
func ListAllianceMembers(c *fiber.Ctx) error {
	// Get all alliance members from database
	rows, err := db.DB.Query(`
		SELECT id, name, did, member_type, join_date, status, location, website, contact_details
		FROM alliance_members
		ORDER BY name
	`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()
	
	// Process members
	var members []AllianceMember
	
	for rows.Next() {
		var member AllianceMember
		err := rows.Scan(
			&member.ID,
			&member.Name,
			&member.DID,
			&member.MemberType,
			&member.JoinDate,
			&member.Status,
			&member.Location,
			&member.Website,
			&member.ContactDetails,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error parsing alliance member data")
		}
		
		members = append(members, member)
	}
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Alliance members retrieved successfully",
		Data:    members,
	})
}

// JoinAlliance submits a request to join the industry alliance
// @Summary Join alliance
// @Description Submit a request to join the industry alliance
// @Tags alliance
// @Accept json
// @Produce json
// @Param request body JoinAllianceRequest true "Alliance join request details"
// @Success 201 {object} SuccessResponse{data=AllianceMember}
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /alliance/join [post]
func JoinAlliance(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Parse request
	var req JoinAllianceRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.Name == "" || req.DID == "" || req.MemberType == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Name, DID, and member type are required")
	}
	
	// Check if already a member
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM alliance_members WHERE did = $1)", req.DID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if exists {
		return fiber.NewError(fiber.StatusConflict, "Already a member of the alliance")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Verify DID exists
	identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)
	_, err = identityClient.ResolveDID(req.DID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid DID: "+err.Error())
	}
	
	// Generate member ID
	memberID := "member-" + time.Now().Format("20060102150405")
	now := time.Now()
	
	// Record on blockchain
	_, err = blockchainClient.SubmitTransaction("ALLIANCE_JOIN", map[string]interface{}{
		"member_id":   memberID,
		"name":        req.Name,
		"did":         req.DID,
		"member_type": req.MemberType,
		"location":    req.Location,
		"website":     req.Website,
		"contact":     req.ContactDetails,
		"metadata":    req.Metadata,
		"timestamp":   now,
		"status":      "pending", // New members start as pending until approved
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record join request on blockchain: "+err.Error())
	}
	
	// Record in database
	_, err = db.DB.Exec(`
		INSERT INTO alliance_members (
			id, name, did, member_type, join_date, status, 
			location, website, contact_details, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,
		memberID,
		req.Name,
		req.DID,
		req.MemberType,
		now,
		"pending", // New members start as pending until approved
		req.Location,
		req.Website,
		req.ContactDetails,
		req.Metadata,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save alliance member: "+err.Error())
	}
	
	// Return response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Alliance join request submitted successfully",
		Data: AllianceMember{
			ID:             memberID,
			Name:           req.Name,
			DID:            req.DID,
			MemberType:     req.MemberType,
			JoinDate:       now,
			Status:         "pending",
			Location:       req.Location,
			Website:        req.Website,
			ContactDetails: req.ContactDetails,
		},
	})
}

// Helper function to generate an access token for shared data
func generateAccessToken(shareID string, recipients []string) string {
	// In a real implementation, this would generate a secure JWT or similar token
	// For this example, we'll just return a mock token
	return "alliance-access-" + shareID + "-" + time.Now().Format("20060102")
}

// Helper function to extract data summary based on data type
func extractDataSummary(batchData map[string]interface{}, dataType string) map[string]interface{} {
	summary := make(map[string]interface{})
	
	switch dataType {
	case "origin":
		if origin, ok := batchData["origin"].(map[string]interface{}); ok {
			summary["location"] = origin["location"]
			summary["hatchery"] = origin["hatchery"]
			summary["production_date"] = origin["production_date"]
		}
	case "quality":
		if quality, ok := batchData["quality"].(map[string]interface{}); ok {
			summary["grade"] = quality["grade"]
			summary["certification"] = quality["certification"]
			summary["inspection_date"] = quality["inspection_date"]
		}
	case "health":
		if health, ok := batchData["health"].(map[string]interface{}); ok {
			summary["health_status"] = health["health_status"]
			summary["disease_free"] = health["disease_free"]
			summary["treatment_history"] = health["treatment_history"]
		}
	case "all":
		summary = batchData
	default:
		// Return specific field if it exists
		if field, ok := batchData[dataType].(map[string]interface{}); ok {
			summary = field
		}
	}
	
	return summary
}