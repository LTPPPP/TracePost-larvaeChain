package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vietchain/tracepost-larvae/blockchain"
	"github.com/vietchain/tracepost-larvae/config"
	"github.com/vietchain/tracepost-larvae/db"
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

// VerifiableClaimRequest represents a request to create a verifiable claim
type VerifiableClaimRequest struct {
	IssuerDID  string                 `json:"issuer_did"`
	SubjectDID string                 `json:"subject_did"`
	ClaimType  string                 `json:"claim_type"`
	Claims     map[string]interface{} `json:"claims"`
	ExpiryDays int                    `json:"expiry_days"`
}