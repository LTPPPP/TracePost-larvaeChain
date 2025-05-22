package api

import (
	"time"
	"fmt"
	"crypto/rand"
	// "encoding/hex"
	"strings"
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/components"
	"github.com/LTPPPP/TracePost-larvaeChain/config"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"github.com/LTPPPP/TracePost-larvaeChain/middleware"
	"github.com/LTPPPP/TracePost-larvaeChain/models"
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
	CompanyID string `json:"company_id,omitempty"` // Optional for user role
	Role      string `json:"role"`
}

func (r *RegisterRequest) Validate() error {
	if r.Username == "" {
		return fmt.Errorf("username is required")
	}
	if r.Password == "" {
		return fmt.Errorf("password is required")
	}
	if r.Email == "" {
		return fmt.Errorf("email is required")
	}
	if r.Role == "" {
		return fmt.Errorf("role is required")
	}
	
	// Validate company_id requirement based on role
	if r.Role != "user" && r.CompanyID == "" {
		return fmt.Errorf("company_id is required for role: %s", r.Role)
	}
	
	return nil
}

// TokenResponse represents the token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// RefreshTokenRequest represents the refresh token request body
type RefreshTokenRequest struct {
	AccessToken string `json:"access_token"`
}

// ForgotPasswordRequest represents the forgot password request body
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// VerifyOTPRequest represents the OTP verification request body
type VerifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

// ResetPasswordRequest represents the reset password request body
type ResetPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
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
	if req.Username == "" || req.Password == "" || req.Email == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Username, password, email are required")
	}

	// Set default role if not provided
	if req.Role == "" {
		req.Role = "user"
	}

	// Auto-generate default profile information
	// Extract name from email if no username specified
	if req.Username == req.Email {
		// Extract part before @ to use as username
		parts := strings.Split(req.Email, "@")
		if len(parts) > 0 {
			req.Username = parts[0]
		}
	}

	// Validate company_id only for non-consumer roles
	if strings.ToLower(req.Role) != "consumer" && req.CompanyID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Company ID is required for this role")
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

	// Prepare company_id for DB (nil if not provided)
	var companyID interface{}
	if strings.ToLower(req.Role) == "consumer" || req.CompanyID == "" {
		companyID = nil
	} else {
		id, err := strconv.Atoi(req.CompanyID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid company ID")
		}
		companyID = id
	}

	// Generate default profile information based on email and role
	fullName := ""
	if parts := strings.Split(req.Email, "@"); len(parts) > 0 {
		// Replace dots and underscores with spaces and capitalize words
		namePart := strings.ReplaceAll(parts[0], ".", " ")
		namePart = strings.ReplaceAll(namePart, "_", " ")
		namePart = strings.Title(strings.ToLower(namePart))
		fullName = namePart
	}

	// Insert user into database with profile information
	query := `
	INSERT INTO account (username, password_hash, email, role, company_id, full_name, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	RETURNING id
	`
	var userID int
	err = db.DB.QueryRow(query, req.Username, string(hashedPassword), req.Email, req.Role, companyID, fullName).Scan(&userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create user")
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "User registered successfully",
		Data: map[string]interface{}{
			"user_id": userID,
		},
	})
}

// generateJWTToken generates a JWT token for a user
func generateJWTToken(user models.User) (string, int, error) {
	// Get configuration
	cfg := config.GetConfig()
	
	// Get JWT secret with fallback
	secretKey, err := config.GetJWTSecret()
	if err != nil {
		// Log error and use default
		fmt.Printf("Error loading JWT secret: %v, using default value\n", err)
		secretKey = cfg.JWTSecret
	}
	
	// Set expiration time based on config (hours)
	expirationTime := time.Now().Add(time.Duration(cfg.JWTExpiration) * time.Hour)
	expiresIn := int(expirationTime.Sub(time.Now()).Seconds())

	// Create claims with proper fields
	claims := models.JWTClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CompanyID: user.CompanyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.JWTIssuer,
			Subject:   fmt.Sprintf("%d", user.ID),
			ID:        generateTokenID(), // Unique token ID for revocation if needed
		},
	}

	// Create token with HMAC-SHA256 signing method (more secure than default)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Sign token with secret key from config
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", 0, err
	}

	return signedToken, expiresIn, nil
}

// generateTokenID creates a unique ID for each token
func generateTokenID() string {
	// Generate a random token ID (UUID)
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	
	return fmt.Sprintf("%x-%x-%x-%x-%x", 
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
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
// @Router /identity/legacy/create [post]
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
// @Router /identity/legacy/resolve/{did} [get]
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
// @Router /identity/legacy/claims [post]
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
// @Router /identity/legacy/claims/verify/{claimId} [get]
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
// @Router /identity/legacy/claims/revoke/{claimId} [post]
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
// @Security Bearer
// @Success 200 {object} SuccessResponse
// @Router /auth/logout [post]
func Logout(c *fiber.Ctx) error {
	// Get token from request
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.JSON(SuccessResponse{
			Success: true,
			Message: "Successfully logged out",
		})
	}
	
	// Extract token
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		// Parse token to get claims
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Get configuration
		cfg := config.GetConfig()
		
		// Get JWT secret with fallback
		secretKey, err := config.GetJWTSecret()
		if err != nil {
			// Log error and use default
			fmt.Printf("Error loading JWT secret: %v, using default value\n", err)
			secretKey = cfg.JWTSecret
		}
		
		return []byte(secretKey), nil
	})
	
	// If token is valid, add it to blacklist
	if err == nil && token.Valid {
		claims, ok := token.Claims.(*models.JWTClaims)
		if ok && claims.ID != "" {
			// Add token to blacklist
			expirationTime := time.Unix(claims.ExpiresAt.Unix(), 0)
			middleware.RevokeToken(claims.ID, expirationTime)
		}
	}
	
	// Clear the JWT cookie if using cookie-based auth
	c.ClearCookie("token")
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Successfully logged out",
	})
}

// RefreshToken refreshes an existing JWT token
// @Summary Refresh JWT token
// @Description Refresh an existing JWT token before it expires
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Token refresh request"
// @Success 200 {object} SuccessResponse{data=TokenResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func RefreshToken(c *fiber.Ctx) error {
	// Get configuration
	cfg := config.GetConfig()
	
	// Parse request body
	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	
	// Validate input
	if req.AccessToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Access token is required")
	}
	
	// Parse the token to get claims
	token, err := jwt.ParseWithClaims(req.AccessToken, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		
			// Get JWT secret with fallback
		secretKey, err := config.GetJWTSecret()
		if err != nil {
			// Log error and use default
			fmt.Printf("Error loading JWT secret: %v, using default value\n", err)
			secretKey = cfg.JWTSecret
		}
		
		return []byte(secretKey), nil
	})
	
	if err != nil {
		// Only allow refresh for expired tokens, not for invalid tokens
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors == jwt.ValidationErrorExpired {
				// Continue with refresh for expired tokens
			} else {
				// Return error for other validation issues
				return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
			}
		} else {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
		}
	}
	
	// Extract claims
	var claims *models.JWTClaims
	if token.Valid {
		// Token is still valid, extract claims
		var ok bool
		claims, ok = token.Claims.(*models.JWTClaims)
		if !ok {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse token claims")
		}
	} else {
		// Token is expired, extract claims ignoring expiration
		claims, _ = token.Claims.(*models.JWTClaims)
		if claims == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token claims")
		}
	}
	
	// Look up user in database
	var user models.User
	query := "SELECT id, username, role, company_id FROM account WHERE id = $1"
	err = db.DB.QueryRow(query, claims.UserID).Scan(&user.ID, &user.Username, &user.Role, &user.CompanyID)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "User not found")
	}
	
	// Generate new JWT token
	newToken, expiresIn, err := generateJWTToken(user)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate token")
	}
	
	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Token refreshed successfully",
		Data: TokenResponse{
			AccessToken: newToken,
			TokenType:   "bearer",
			ExpiresIn:   expiresIn,
		},
	})
}

// ForgotPassword handles forgot password requests
// @Summary Forgot password
// @Description Send OTP to user's email for password reset
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Forgot password details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /auth/forgot-password [post]
func ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	if req.Email == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Email is required")
	}
	// Check if user exists
	var userID int
	err := db.DB.QueryRow("SELECT id FROM account WHERE email = $1 AND is_active = true", req.Email).Scan(&userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Email not found")
	}
	// Generate OTP
	otp := generateOTP(6)
	expiry := 10 * time.Minute
	// Store OTP in Redis
	ctx := context.Background()
	redisKey := db.OTPKey(req.Email)
	err = db.Redis.Set(ctx, redisKey, otp, expiry).Err()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to store OTP")
	}
	// Send OTP via email
	subject := "Your OTP for Password Reset"
	body := fmt.Sprintf("Your OTP code is: %s\nIt expires in 10 minutes.", otp)
	err = components.SendEmail(req.Email, subject, body)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to send OTP email")
	}
	return c.JSON(SuccessResponse{Success: true, Message: "OTP sent to email"})
}

// VerifyOTP handles OTP verification
// @Summary Verify OTP
// @Description Verify OTP for password reset
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyOTPRequest true "Verify OTP details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /auth/verify-otp [post]
func VerifyOTP(c *fiber.Ctx) error {
	var req VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	if req.Email == "" || req.OTP == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Email and OTP are required")
	}
	ctx := context.Background()
	redisKey := db.OTPKey(req.Email)
	val, err := db.Redis.Get(ctx, redisKey).Result()
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "OTP not found or expired")
	}
	if req.OTP != val {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid OTP")
	}
	return c.JSON(SuccessResponse{Success: true, Message: "OTP verified"})
}

// ResetPassword handles password reset
// @Summary Reset password
// @Description Reset password using OTP
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset password details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /auth/reset-password [post]
func ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	if req.Email == "" || req.OTP == "" || req.NewPassword == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Email, OTP, and new password are required")
	}
	ctx := context.Background()
	redisKey := db.OTPKey(req.Email)
	val, err := db.Redis.Get(ctx, redisKey).Result()
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "OTP not found or expired")
	}
	if req.OTP != val {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid OTP")
	}
	// Check if user exists
	var userID int
	err = db.DB.QueryRow("SELECT id FROM account WHERE email = $1 AND is_active = true", req.Email).Scan(&userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Email not found")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to hash password")
	}
	_, err = db.DB.Exec("UPDATE account SET password_hash = $1, updated_at = NOW() WHERE id = $2", hashedPassword, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update password")
	}
	// Invalidate OTP
	_ = db.Redis.Del(ctx, redisKey).Err()
	return c.JSON(SuccessResponse{Success: true, Message: "Password reset successful"})
}

// generateOTP generates a random numeric OTP of given length
func generateOTP(length int) string {
	const digits = "0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := 0; i < length; i++ {
		b[i] = digits[int(b[i])%10]
	}
	return string(b)
}