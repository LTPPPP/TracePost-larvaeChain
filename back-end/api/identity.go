package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/config"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CreateDID creates a new decentralized identity
// @Summary Create a new decentralized identity (DID)
// @Description Create a new decentralized identity for an entity in the supply chain
// @Tags identity
// @Accept json
// @Produce json
// @Param request body CreateIdentityRequest true "DID creation details"
// @Success 201 {object} SuccessResponse{data=DecentralizedIDResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /identity/did [post]
func CreateDID(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Parse request
	var req CreateIdentityRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.EntityType == "" || req.EntityName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Entity type and name are required")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Create identity client
	identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)
	
	// Create DID
	did, err := identityClient.CreateDecentralizedID(req.EntityType, req.EntityName, req.Metadata)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create DID: "+err.Error())
	}
	
	// Save DID to database for future reference
	_, err = db.DB.Exec(`
		INSERT INTO identities (did, entity_type, entity_name, public_key, metadata, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, 
		did.DID, 
		req.EntityType, 
		req.EntityName, 
		did.PublicKey, 
		did.MetaData, 
		did.Status, 
		did.Created, 
		did.Updated,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save DID to database: "+err.Error())
	}
	
	// Return response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Decentralized identity created successfully",
		Data: DecentralizedIDResponse{
			DID:           did.DID,
			ControllerDID: did.ControllerDID,
			PublicKey:     did.PublicKey,
			MetaData:      did.MetaData,
			Status:        did.Status,
			Created:       did.Created,
			Updated:       did.Updated,
		},
	})
}

// ResolveDIDFromIdentity resolves a DID to retrieve the associated DID document
// @Summary Resolve a DID
// @Description Resolve a DID to retrieve the associated DID document
// @Tags identity
// @Accept json
// @Produce json
// @Param did path string true "Decentralized Identifier (DID)"
// @Success 200 {object} SuccessResponse{data=DIDResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /identity/did/{did} [get]
func ResolveDIDFromIdentity(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Get DID from path
	didStr := c.Params("did")
	if didStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "DID is required")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Create identity client
	identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)
	
	// Resolve DID
	did, err := identityClient.ResolveDID(didStr)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "DID not found: "+err.Error())
	}
	
	// Convert proof to map for JSON response
	var proofMap map[string]interface{}
	if did.Proof != nil {
		proofMap = map[string]interface{}{
			"type":                did.Proof.Type,
			"created":             did.Proof.Created.Format(time.RFC3339),
			"verification_method": did.Proof.VerificationMethod,
			"proof_purpose":       did.Proof.ProofPurpose,
			"proof_value":         did.Proof.ProofValue,
		}
	}
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "DID resolved successfully",
		Data: DIDResponse{
			DID:           did.DID,
			ControllerDID: did.ControllerDID,
			PublicKey:     did.PublicKey,
			MetaData:      did.MetaData,
			Status:        did.Status,
			Created:       did.Created.Format(time.RFC3339),
			Updated:       did.Updated.Format(time.RFC3339),
			Proof:         proofMap,
		},
	})
}

// CreateVerifiableClaimFromIdentity creates a verifiable claim about an identity
// @Summary Create a verifiable claim
// @Description Create a verifiable claim about an identity
// @Tags identity
// @Accept json
// @Produce json
// @Param request body VerifiableClaimRequest true "Verifiable claim details"
// @Success 201 {object} SuccessResponse{data=VerifiableClaimResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /identity/claim [post]
func CreateVerifiableClaimFromIdentity(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Parse request
	var req VerifiableClaimRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.IssuerDID == "" || req.SubjectDID == "" || req.ClaimType == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Issuer DID, subject DID, and claim type are required")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Create identity client
	identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)
	
	// Create claim
	claim, err := identityClient.CreateVerifiableClaim(
		req.IssuerDID,
		req.SubjectDID,
		req.ClaimType,
		req.Claims,
		req.ExpiryDays,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create verifiable claim: "+err.Error())
	}
	
	// Save claim to database
	_, err = db.DB.Exec(`
		INSERT INTO verifiable_claims (claim_id, claim_type, issuer_did, subject_did, claims, issuance_date, expiry_date, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		claim.ID,
		claim.Type,
		claim.Issuer,
		claim.Subject,
		claim.Claims,
		claim.IssuanceDate,
		claim.ExpiryDate,
		claim.Status,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save claim to database: "+err.Error())
	}
	
	// Return response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Verifiable claim created successfully",
		Data: VerifiableClaimResponse{
			ID:           claim.ID,
			Type:         claim.Type,
			Issuer:       claim.Issuer,
			Subject:      claim.Subject,
			IssuanceDate: claim.IssuanceDate.Format(time.RFC3339),
			ExpiryDate:   claim.ExpiryDate.Format(time.RFC3339),
			Claims:       claim.Claims,
			Status:       claim.Status,
		},
	})
}

// GetVerifiableClaim gets a verifiable claim by ID
// @Summary Get a verifiable claim
// @Description Get a verifiable claim by ID
// @Tags identity
// @Accept json
// @Produce json
// @Param claimId path string true "Claim ID"
// @Success 200 {object} SuccessResponse{data=VerifiableClaimResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /identity/claim/{claimId} [get]
func GetVerifiableClaim(c *fiber.Ctx) error {
	// Get claim ID from path
	claimID := c.Params("claimId")
	if claimID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Claim ID is required")
	}
	
	// Get claim from database
	var claim struct {
		ID           string
		Type         string
		Issuer       string
		Subject      string
		Claims       map[string]interface{}
		IssuanceDate time.Time
		ExpiryDate   time.Time
		Status       string
	}
	
	err := db.DB.QueryRow(`
		SELECT claim_id, claim_type, issuer_did, subject_did, claims, issuance_date, expiry_date, status
		FROM verifiable_claims
		WHERE claim_id = $1
	`, claimID).Scan(
		&claim.ID,
		&claim.Type,
		&claim.Issuer,
		&claim.Subject,
		&claim.Claims,
		&claim.IssuanceDate,
		&claim.ExpiryDate,
		&claim.Status,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Claim not found")
	}
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Verifiable claim retrieved successfully",
		Data: VerifiableClaimResponse{
			ID:           claim.ID,
			Type:         claim.Type,
			Issuer:       claim.Issuer,
			Subject:      claim.Subject,
			IssuanceDate: claim.IssuanceDate.Format(time.RFC3339),
			ExpiryDate:   claim.ExpiryDate.Format(time.RFC3339),
			Claims:       claim.Claims,
			Status:       claim.Status,
		},
	})
}

// VerifyIdentityClaim verifies a claim
// @Summary Verify a claim
// @Description Verify a claim's validity
// @Tags identity
// @Accept json
// @Produce json
// @Param request body VerifyClaimRequest true "Verification request"
// @Success 200 {object} SuccessResponse{data=VerificationResultResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /identity/claim/verify [post]
func VerifyIdentityClaim(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Parse request
	var req VerifyClaimRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.ClaimID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Claim ID is required")
	}
	
	// Get claim from database
	var claim blockchain.IdentityClaim
	var issuanceDate, expiryDate time.Time
	var status string
	
	err := db.DB.QueryRow(`
		SELECT claim_id, claim_type, issuer_did, subject_did, claims, issuance_date, expiry_date, status
		FROM verifiable_claims
		WHERE claim_id = $1
	`, req.ClaimID).Scan(
		&claim.ID,
		&claim.Type,
		&claim.Issuer,
		&claim.Subject,
		&claim.Claims,
		&issuanceDate,
		&expiryDate,
		&status,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Claim not found")
	}
	
	// Set time fields
	claim.IssuanceDate = issuanceDate
	claim.ExpiryDate = expiryDate
	claim.Status = status
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Create identity client
	identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)
	
	// Verify claim
	result, err := identityClient.VerifyClaim(&claim)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to verify claim: "+err.Error())
	}
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Claim verification completed",
		Data: VerificationResultResponse{
			IsValid:        result.IsValid,
			ValidationTime: result.ValidationTime,
			Errors:         result.Errors,
		},
	})
}

// RevokeIdentityClaim revokes a claim
// @Summary Revoke a claim
// @Description Revoke a previously issued claim
// @Tags identity
// @Accept json
// @Produce json
// @Param claimId path string true "Claim ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /identity/claim/{claimId}/revoke [put]
func RevokeIdentityClaim(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Get claim ID from path
	claimID := c.Params("claimId")
	if claimID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Claim ID is required")
	}
	
	// Get user DID from JWT token
	userDID := c.Locals("user_did").(string)
	if userDID == "" {
		return fiber.NewError(fiber.StatusForbidden, "User DID not found in token")
	}
	
	// Check if claim exists and user is the issuer
	var issuerDID string
	err := db.DB.QueryRow(`
		SELECT issuer_did
		FROM verifiable_claims
		WHERE claim_id = $1
	`, claimID).Scan(&issuerDID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Claim not found")
	}
	
	// Check if user is the issuer
	if issuerDID != userDID {
		return fiber.NewError(fiber.StatusForbidden, "Only the issuer can revoke a claim")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Create identity client
	identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)
	
	// Revoke claim
	err = identityClient.RevokeClaim(claimID, userDID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to revoke claim: "+err.Error())
	}
	
	// Update claim status in database
	_, err = db.DB.Exec(`
		UPDATE verifiable_claims
		SET status = 'revoked'
		WHERE claim_id = $1
	`, claimID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update claim status in database")
	}
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Claim revoked successfully",
	})
}

// Helper function to resolve external DIDs (like did:web, did:ethr, etc.)
func resolveExternalDID(did string, cfg *config.Config) (map[string]interface{}, error) {
	// Parse DID to determine which method to use
	parts := strings.Split(did, ":")
	if len(parts) < 3 {
		return nil, errors.New("invalid DID format")
	}
	
	method := parts[1]
	
	// Determine resolver URL based on method
	var resolverURL string
	switch method {
	case "web":
		resolverURL = cfg.IdentityResolverURL + "/1.0/identifiers/" + did
	case "ethr":
		resolverURL = cfg.IdentityResolverURL + "/1.0/identifiers/" + did
	case "key":
		resolverURL = cfg.IdentityResolverURL + "/1.0/identifiers/" + did
	default:
		resolverURL = cfg.IdentityResolverURL + "/1.0/identifiers/" + did
	}
	
	// Make HTTP request to universal resolver
	resp, err := http.Get(resolverURL)
	if err != nil {
		return nil, fmt.Errorf("resolver error: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("resolver returned status %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode resolver response: %v", err)
	}
	
	return result, nil
}

// VerifiableClaimRequest represents the request to create a verifiable claim
type VerifiableClaimRequest struct {
	IssuerDID  string                 `json:"issuer_did"`
	SubjectDID string                 `json:"subject_did"`
	ClaimType  string                 `json:"claim_type"`
	Claims     map[string]interface{} `json:"claims"`
	ExpiryDays int                    `json:"expiry_days"`
}

// CreateDIDV2 creates a new decentralized identity with improved capabilities
// @Summary Create decentralized identity
// @Description Create a new decentralized identity (DID) for an entity
// @Tags identity
// @Accept json
// @Produce json
// @Param request body CreateIdentityRequest true "Identity creation details"
// @Success 201 {object} SuccessResponse{data=DecentralizedIDResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /identity/create [post]
func CreateDIDV2(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Parse request
	var req CreateIdentityRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.EntityType == "" || req.EntityName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Entity type and name are required")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Create identity client
	identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)
	
	// Add additional metadata for enhanced DID
	if req.Metadata == nil {
		req.Metadata = make(map[string]interface{})
	}
	
	// Add industry-specific metadata based on entity type
	switch req.EntityType {
	case "hatchery":
		req.Metadata["industry"] = "aquaculture"
		req.Metadata["entity_role"] = "producer"
	case "processor":
		req.Metadata["industry"] = "aquaculture"
		req.Metadata["entity_role"] = "processor"
	case "distributor":
		req.Metadata["industry"] = "aquaculture"
		req.Metadata["entity_role"] = "distributor"
	case "retailer":
		req.Metadata["industry"] = "retail"
		req.Metadata["entity_role"] = "seller"
	case "certifier":
		req.Metadata["industry"] = "certification"
		req.Metadata["entity_role"] = "verifier"
	}
	
	// Add timestamp for creation
	req.Metadata["created_timestamp"] = time.Now().Format(time.RFC3339)
	
	// Create DID
	did, err := identityClient.CreateDecentralizedID(req.EntityType, req.EntityName, req.Metadata)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create DID: "+err.Error())
	}
	
	// Save DID to database for future reference
	_, err = db.DB.Exec(`
		INSERT INTO identities (did, entity_type, entity_name, public_key, metadata, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, 
		did.DID, 
		req.EntityType, 
		req.EntityName, 
		did.PublicKey, 
		did.MetaData, 
		did.Status, 
		did.Created, 
		did.Updated,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save DID to database: "+err.Error())
	}
	
	// Return response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Decentralized identity created successfully",
		Data: DecentralizedIDResponse{
			DID:           did.DID,
			ControllerDID: did.ControllerDID,
			PublicKey:     did.PublicKey,
			MetaData:      did.MetaData,
			Status:        did.Status,
			Created:       did.Created,
			Updated:       did.Updated,
		},
	})
}

// ResolveDIDV2 resolves a DID to retrieve the associated DID document with improved capabilities
// @Summary Resolve decentralized identity
// @Description Resolve a DID to retrieve its DID document
// @Tags identity
// @Accept json
// @Produce json
// @Param did path string true "Decentralized Identifier (DID)"
// @Success 200 {object} SuccessResponse{data=DecentralizedIDResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /identity/resolve/{did} [get]
func ResolveDIDV2(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Get DID from path
	didStr := c.Params("did")
	if didStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "DID is required")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Create identity client
	identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)
	
	// Resolve DID from local DB first for performance
	var didDoc struct {
		DID           string
		EntityType    string
		EntityName    string
		PublicKey     string
		Metadata      map[string]interface{}
		Status        string
		CreatedAt     time.Time
		UpdatedAt     time.Time
		ControllerDID string
	}
	
	err := db.DB.QueryRow(`
		SELECT did, entity_type, entity_name, public_key, metadata, status, created_at, updated_at, 
			   COALESCE(metadata->>'controller_did', '') as controller_did
		FROM identities
		WHERE did = $1
	`, didStr).Scan(
		&didDoc.DID,
		&didDoc.EntityType,
		&didDoc.EntityName,
		&didDoc.PublicKey,
		&didDoc.Metadata,
		&didDoc.Status,
		&didDoc.CreatedAt,
		&didDoc.UpdatedAt,
		&didDoc.ControllerDID,
	)
	
	if err == nil {
		// Found in local DB, return it
		return c.JSON(SuccessResponse{
			Success: true,
			Message: "DID resolved successfully from local database",
			Data: DecentralizedIDResponse{
				DID:           didDoc.DID,
				ControllerDID: didDoc.ControllerDID,
				PublicKey:     didDoc.PublicKey,
				MetaData:      didDoc.Metadata,
				Status:        didDoc.Status,
				Created:       didDoc.CreatedAt,
				Updated:       didDoc.UpdatedAt,
			},
		})
	}
	
	// Not found in local DB, try to resolve from blockchain
	did, err := identityClient.ResolveDID(didStr)
	if err != nil {
		// Try to resolve via interoperability if it's an external DID
		if strings.HasPrefix(didStr, "did:") && !strings.HasPrefix(didStr, "did:tracepost:") {
			// Try to resolve from other DID systems via interoperability
			externalDid, err := resolveExternalDID(didStr, cfg)
			if err != nil {
				return fiber.NewError(fiber.StatusNotFound, "DID not found: "+err.Error())
			}
			
			return c.JSON(SuccessResponse{
				Success: true,
				Message: "External DID resolved successfully",
				Data:    externalDid,
			})
		}
		
		return fiber.NewError(fiber.StatusNotFound, "DID not found: "+err.Error())
	}
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "DID resolved successfully from blockchain",
		Data: DecentralizedIDResponse{
			DID:           did.DID,
			ControllerDID: did.ControllerDID,
			PublicKey:     did.PublicKey,
			MetaData:      did.MetaData,
			Status:        did.Status,
			Created:       did.Created,
			Updated:       did.Updated,
		},
	})
}

// ListDIDs lists all DIDs matching certain criteria
// @Summary List decentralized identities
// @Description List all DIDs that match given criteria
// @Tags identity
// @Accept json
// @Produce json
// @Param entity_type query string false "Filter by entity type"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of items per page (default: 20)"
// @Success 200 {object} SuccessResponse{data=DIDListResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /identity/list [get]
func ListDIDs(c *fiber.Ctx) error {
	// Get query parameters
	entityType := c.Query("entity_type")
	status := c.Query("status", "active") // Default to active DIDs
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	// Validate parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	// Build query
	query := `
		SELECT did, entity_type, entity_name, status, created_at
		FROM identities
		WHERE 1=1
	`
	countQuery := `
		SELECT COUNT(*)
		FROM identities
		WHERE 1=1
	`
	
	var args []interface{}
	var argIndex int = 1
	
	if entityType != "" {
		query += fmt.Sprintf(" AND entity_type = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND entity_type = $%d", argIndex)
		args = append(args, entityType)
		argIndex++
	}
	
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}
	
	// Add pagination
	query += " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
	args = append(args, limit, (page-1)*limit)
	
	// Get total count
	var total int
	err := db.DB.QueryRow(countQuery, args[:argIndex-1]...).Scan(&total)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	
	// Execute query
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	defer rows.Close()
	
	// Parse results
	var dids []DIDSummary
	for rows.Next() {
		var did DIDSummary
		err := rows.Scan(&did.DID, &did.EntityType, &did.EntityName, &did.Status, &did.Created)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error scanning DID: "+err.Error())
		}
		dids = append(dids, did)
	}
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "DIDs retrieved successfully",
		Data: DIDListResponse{
			DIDs:  dids,
			Total: total,
			Page:  page,
			Limit: limit,
		},
	})
}

// CreateVerifiableClaimV2 creates a verifiable claim with enhanced capabilities
// @Summary Create verifiable claim
// @Description Create a verifiable claim about a decentralized identity
// @Tags identity
// @Accept json
// @Produce json
// @Param request body CreateVerifiableClaimRequest true "Claim creation details"
// @Success 201 {object} SuccessResponse{data=VerifiableClaimResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /identity/claims [post]
func CreateVerifiableClaimV2(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Parse request
	var req CreateVerifiableClaimRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.IssuerDID == "" || req.SubjectDID == "" || req.ClaimType == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Issuer DID, subject DID, and claim type are required")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Create identity client
	identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)
	
	// Set default expiry days if not provided
	if req.ExpiryDays <= 0 {
		req.ExpiryDays = 365 // Default to one year
	}
	
	// Create claim
	claim, err := identityClient.CreateVerifiableClaim(
		req.IssuerDID,
		req.SubjectDID,
		req.ClaimType,
		req.Claims,
		req.ExpiryDays,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create verifiable claim: "+err.Error())
	}
	
	// Save claim to database
	_, err = db.DB.Exec(`
		INSERT INTO verifiable_claims (claim_id, claim_type, issuer_did, subject_did, claims, issuance_date, expiry_date, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		claim.ID,
		claim.Type,
		claim.Issuer,
		claim.Subject,
		claim.Claims,
		claim.IssuanceDate,
		claim.ExpiryDate,
		claim.Status,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save claim to database: "+err.Error())
	}
	
	// Return response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Verifiable claim created successfully",
		Data: VerifiableClaimResponse{
			ID:           claim.ID,
			Type:         claim.Type,
			Issuer:       claim.Issuer,
			Subject:      claim.Subject,
			IssuanceDate: claim.IssuanceDate.Format(time.RFC3339),
			ExpiryDate:   claim.ExpiryDate.Format(time.RFC3339),
			Claims:       claim.Claims,
			Status:       claim.Status,
		},
	})
}

// VerifyClaimV2 verifies a claim with enhanced validation
// @Summary Verify claim
// @Description Verify a claim about a decentralized identity
// @Tags identity
// @Accept json
// @Produce json
// @Param claimId path string true "Claim ID"
// @Success 200 {object} SuccessResponse{data=VerificationResultResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /identity/claims/verify/{claimId} [get]
func VerifyClaimV2(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Get claim ID from path
	claimID := c.Params("claimId")
	if claimID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Claim ID is required")
	}
	
	// Get claim from database
	var claim blockchain.IdentityClaim
	var issuanceDate, expiryDate time.Time
	var status string
	
	err := db.DB.QueryRow(`
		SELECT claim_id, claim_type, issuer_did, subject_did, claims, issuance_date, expiry_date, status
		FROM verifiable_claims
		WHERE claim_id = $1
	`, claimID).Scan(
		&claim.ID,
		&claim.Type,
		&claim.Issuer,
		&claim.Subject,
		&claim.Claims,
		&issuanceDate,
		&expiryDate,
		&status,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Claim not found")
	}
	
	// Set time fields
	claim.IssuanceDate = issuanceDate
	claim.ExpiryDate = expiryDate
	claim.Status = status
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Create identity client
	identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)
	
	// Add enhanced validation
	var verificationErrors []string
	
	// Check if claim has expired
	if time.Now().After(claim.ExpiryDate) {
		verificationErrors = append(verificationErrors, "Claim has expired")
	}
	
	// Check if claim has been revoked
	if claim.Status == "revoked" {
		verificationErrors = append(verificationErrors, "Claim has been revoked")
	}
	
	// Verify claim using blockchain verification
	blockchainResult, err := identityClient.VerifyClaim(&claim)
	if err != nil {
		verificationErrors = append(verificationErrors, "Blockchain verification failed: "+err.Error())
	} else if !blockchainResult.IsValid {
		verificationErrors = append(verificationErrors, blockchainResult.Errors...)
	}
	
	// Determine final validation result
	isValid := len(verificationErrors) == 0
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Claim verification completed",
		Data: VerificationResultResponse{
			IsValid:        isValid,
			ValidationTime: time.Now(),
			Errors:         verificationErrors,
		},
	})
}

// RevokeClaimV2 revokes a claim with enhanced security
// @Summary Revoke claim
// @Description Revoke a verifiable claim
// @Tags identity
// @Accept json
// @Produce json
// @Param claimId path string true "Claim ID"
// @Param issuerDid query string true "Issuer DID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /identity/claims/revoke/{claimId} [post]
func RevokeClaimV2(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Get claim ID from path
	claimID := c.Params("claimId")
	if claimID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Claim ID is required")
	}
	
	// Get issuer DID from query
	issuerDID := c.Query("issuerDid")
	if issuerDID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Issuer DID is required")
	}
	
	// Check if claim exists and user is the issuer
	var dbIssuerDID string
	err := db.DB.QueryRow(`
		SELECT issuer_did
		FROM verifiable_claims
		WHERE claim_id = $1
	`, claimID).Scan(&dbIssuerDID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Claim not found")
	}
	
	// Check if user is the issuer
	if dbIssuerDID != issuerDID {
		return fiber.NewError(fiber.StatusForbidden, "Only the issuer can revoke a claim")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Create identity client
	identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)
	
	// Revoke claim
	err = identityClient.RevokeClaim(claimID, issuerDID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to revoke claim: "+err.Error())
	}
	
	// Update claim status in database
	_, err = db.DB.Exec(`
		UPDATE verifiable_claims
		SET status = 'revoked'
		WHERE claim_id = $1
	`, claimID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update claim status in database")
	}
	
	// Record revocation event on blockchain
	_, err = blockchainClient.SubmitTransaction(
		"CLAIM_REVOKED",
		map[string]interface{}{
			"claim_id":   claimID,
			"issuer_did": issuerDID,
			"revoked_at": time.Now(),
		},
	)
	if err != nil {
		// Log error but continue, as the claim is already revoked in the database
		// In a production environment, you would implement a retry mechanism
		fmt.Printf("Failed to record revocation on blockchain: %v\n", err)
	}
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Claim revoked successfully",
	})
}

// Do not duplicate struct declarations that already exist in auth.go

// DIDResponse represents a DID document response
type DIDResponse struct {
	DID           string                 `json:"did"`
	ControllerDID string                 `json:"controller_did,omitempty"`
	PublicKey     string                 `json:"public_key"`
	MetaData      map[string]interface{} `json:"metadata"`
	Status        string                 `json:"status"`
	Created       string                 `json:"created"`
	Updated       string                 `json:"updated"`
	Proof         map[string]interface{} `json:"proof,omitempty"`
}

// DIDListResponse represents a list of DIDs
type DIDListResponse struct {
	DIDs  []DIDSummary `json:"dids"`
	Total int          `json:"total"`
	Page  int          `json:"page"`
	Limit int          `json:"limit"`
}

// DIDSummary represents a summary of a DID
type DIDSummary struct {
	DID        string    `json:"did"`
	EntityType string    `json:"entity_type"`
	EntityName string    `json:"entity_name"`
	Status     string    `json:"status"`
	Created    time.Time `json:"created"`
}