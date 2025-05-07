package api

import (
	"time"
	// "strings"
	"crypto/rand"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/vietchain/tracepost-larvae/blockchain"
	"github.com/vietchain/tracepost-larvae/config"
	"github.com/vietchain/tracepost-larvae/db"
	"github.com/vietchain/tracepost-larvae/models"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	CompanyID string `json:"company_id"`
	Role      string `json:"role"`
}

// TokenResponse represents the token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Login handles user authentication
// @Summary User login
// @Description Authenticate a user and return a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} SuccessResponse{data=TokenResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func Login(c *fiber.Ctx) error {
	// Parse request body
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Username and password are required")
	}

	// Query user from database
	var user models.User
	query := "SELECT id, username, password_hash, role, company_id FROM account WHERE username = $1"
	err := db.DB.QueryRow(query, req.Username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CompanyID)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid username or password")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid username or password")
	}

	// Generate JWT token
	token, expiresIn, err := generateJWTToken(user)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate token")
	}

	// Update last login time
	_, err = db.DB.Exec("UPDATE account SET last_login = NOW() WHERE id = $1", user.ID)
	if err != nil {
		// Not critical, just log the error
		// In a real application, this would be logged properly
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Login successful",
		Data: TokenResponse{
			AccessToken: token,
			TokenType:   "bearer",
			ExpiresIn:   expiresIn,
		},
	})
}

// Register handles user registration
// @Summary User registration
// @Description Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /auth/register [post]
func Register(c *fiber.Ctx) error {
	// Parse request body
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate input
	if req.Username == "" || req.Password == "" || req.Email == "" || req.CompanyID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Username, password, email, and company ID are required")
	}

	// Check if username already exists
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM account WHERE username = $1", req.Username).Scan(&count)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if count > 0 {
		return fiber.NewError(fiber.StatusConflict, "Username already exists")
	}

	// Check if email already exists
	err = db.DB.QueryRow("SELECT COUNT(*) FROM account WHERE email = $1", req.Email).Scan(&count)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if count > 0 {
		return fiber.NewError(fiber.StatusConflict, "Email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to hash password")
	}

	// Set default role if not provided
	if req.Role == "" {
		req.Role = "user"
	}

	// Insert user into database
	query := `
	INSERT INTO account (username, password_hash, email, role, company_id, created_at)
	VALUES ($1, $2, $3, $4, $5, NOW())
	`
	_, err = db.DB.Exec(query, req.Username, string(hashedPassword), req.Email, req.Role, req.CompanyID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create user")
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "User registered successfully",
	})
}

// generateJWTToken generates a JWT token for a user
func generateJWTToken(user models.User) (string, int, error) {
	// Set expiration time (24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)
	expiresIn := int(expirationTime.Sub(time.Now()).Seconds())

	// Create claims
	claims := models.JWTClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CompanyID: user.CompanyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	// In a real application, this should be a secure environment variable
	secretKey := getSecretKey()
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", 0, err
	}

	return signedToken, expiresIn, nil
}

// getSecretKey gets the JWT secret key from environment or generates a random one
func getSecretKey() string {
	// In a real application, this should be a secure environment variable
	secretKey := "your-secret-key" // This should be stored securely in an environment variable

	// If no secret key is set, generate a random one
	// This is only for development purposes and should be replaced with a proper configuration
	if secretKey == "your-secret-key" {
		// Generate a random key
		bytes := make([]byte, 32)
		rand.Read(bytes)
		secretKey = hex.EncodeToString(bytes)
	}

	return secretKey
}

// CreateIdentityRequest represents a request to create a new decentralized identity
type CreateIdentityRequest struct {
	EntityType string                 `json:"entity_type"` // "company", "user", "hatchery", "farm", "processor", etc.
	EntityName string                 `json:"entity_name"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// CreateVerifiableClaimRequest represents a request to create a verifiable claim
type CreateVerifiableClaimRequest struct {
	IssuerDID  string                 `json:"issuer_did"`
	SubjectDID string                 `json:"subject_did"`
	ClaimType  string                 `json:"claim_type"` // "CertifiedHatchery", "OrganicFarm", "QualityProcessor", etc.
	Claims     map[string]interface{} `json:"claims"`
	ExpiryDays int                    `json:"expiry_days"`
}

// VerifyClaimRequest represents a request to verify a claim
type VerifyClaimRequest struct {
	ClaimID string `json:"claim_id"`
}

// DecentralizedIDResponse represents a response for a DID operation
type DecentralizedIDResponse struct {
	DID           string                 `json:"did"`
	ControllerDID string                 `json:"controller_did,omitempty"`
	PublicKey     string                 `json:"public_key"`
	MetaData      map[string]interface{} `json:"metadata"`
	Status        string                 `json:"status"`
	Created       time.Time              `json:"created"`
	Updated       time.Time              `json:"updated"`
}

// VerifiableClaimResponse represents a response for a verifiable claim
type VerifiableClaimResponse struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Issuer       string                 `json:"issuer"`
	Subject      string                 `json:"subject"`
	IssuanceDate string                 `json:"issuance_date"` // Changed from time.Time to string
	ExpiryDate   string                 `json:"expiry_date"`   // Changed from time.Time to string
	Claims       map[string]interface{} `json:"claims"`
	Status       string                 `json:"status"`
}

// VerificationResultResponse represents a response for a verification result
type VerificationResultResponse struct {
	IsValid        bool      `json:"is_valid"`
	ValidationTime time.Time `json:"validation_time"`
	Errors         []string  `json:"errors,omitempty"`
}

// CreateIdentity creates a new decentralized identity
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
func CreateIdentity(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if identity is enabled
	if !cfg.IdentityEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Decentralized identity is not enabled")
	}
	
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
	
	// Create identity
	identity, err := blockchainClient.IdentityClient.CreateDecentralizedID(req.EntityType, req.EntityName, req.Metadata)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create identity: "+err.Error())
	}
	
	// Return response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Decentralized identity created successfully",
		Data: DecentralizedIDResponse{
			DID:           identity.DID,
			ControllerDID: identity.ControllerDID,
			PublicKey:     identity.PublicKey,
			MetaData:      identity.MetaData,
			Status:        identity.Status,
			Created:       identity.Created,
			Updated:       identity.Updated,
		},
	})
}

// ResolveDID resolves a decentralized identity
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
func ResolveDID(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if identity is enabled
	if !cfg.IdentityEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Decentralized identity is not enabled")
	}
	
	// Get DID from path
	did := c.Params("did")
	if did == "" {
		return fiber.NewError(fiber.StatusBadRequest, "DID is required")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Resolve DID
	identity, err := blockchainClient.IdentityClient.ResolveDID(did)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "DID not found: "+err.Error())
	}
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "DID resolved successfully",
		Data: DecentralizedIDResponse{
			DID:           identity.DID,
			ControllerDID: identity.ControllerDID,
			PublicKey:     identity.PublicKey,
			MetaData:      identity.MetaData,
			Status:        identity.Status,
			Created:       identity.Created,
			Updated:       identity.Updated,
		},
	})
}

// CreateVerifiableClaim creates a verifiable claim about an identity
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
func CreateVerifiableClaim(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if identity is enabled
	if !cfg.IdentityEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Decentralized identity is not enabled")
	}
	
	// Parse request
	var req CreateVerifiableClaimRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.IssuerDID == "" || req.SubjectDID == "" || req.ClaimType == "" || len(req.Claims) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required fields")
	}
	
	// Set default expiry if not provided
	if req.ExpiryDays <= 0 {
		req.ExpiryDays = 365 // 1 year by default
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Create claim
	claim, err := blockchainClient.IdentityClient.CreateVerifiableClaim(
		req.IssuerDID,
		req.SubjectDID,
		req.ClaimType,
		req.Claims,
		req.ExpiryDays,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create claim: "+err.Error())
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

// VerifyClaim verifies a claim
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
func VerifyClaim(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if identity is enabled
	if !cfg.IdentityEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Decentralized identity is not enabled")
	}
	
	// Get claim ID from path
	claimID := c.Params("claimId")
	if claimID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Claim ID is required")
	}
	
	// In a real implementation, we would look up the claim in a database or on the blockchain
	// For this example, we'll create a mock claim
	claim := &blockchain.IdentityClaim{
		ID:           claimID,
		Type:         "CertifiedHatchery",
		Issuer:       "did:tracepost:authority:1234",
		Subject:      "did:tracepost:hatchery:5678",
		IssuanceDate: time.Now().AddDate(0, -1, 0), // 1 month ago
		ExpiryDate:   time.Now().AddDate(1, 0, 0),  // 1 year from now
		Claims: map[string]interface{}{
			"certification": "ISO9001",
			"issuedBy":      "Vietnam Fisheries Authority",
			"validFrom":     "2024-01-01",
			"validTo":       "2025-01-01",
		},
		Status: "valid",
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Verify claim
	result, err := blockchainClient.IdentityClient.VerifyClaim(claim)
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

// RevokeClaim revokes a claim
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
func RevokeClaim(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if identity is enabled
	if !cfg.IdentityEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Decentralized identity is not enabled")
	}
	
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
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Revoke claim
	err := blockchainClient.IdentityClient.RevokeClaim(claimID, issuerDID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to revoke claim: "+err.Error())
	}
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Claim revoked successfully",
	})
}

// Logout logs out a user
// @Summary Logout
// @Description Logout and invalidate the user's session
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Router /auth/logout [post]
func Logout(c *fiber.Ctx) error {
	// Clear the JWT cookie if using cookie-based auth
	c.ClearCookie("token")
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Successfully logged out",
	})
}