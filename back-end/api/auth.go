package api

import (
	"time"
	// "strings"
	"crypto/rand"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
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

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CompanyID string `json:"company_id"`
	jwt.RegisteredClaims
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
	query := "SELECT id, username, password_hash, role, company_id FROM users WHERE username = $1"
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
	_, err = db.DB.Exec("UPDATE users SET last_login = NOW() WHERE id = $1", user.ID)
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
	err := db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", req.Username).Scan(&count)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if count > 0 {
		return fiber.NewError(fiber.StatusConflict, "Username already exists")
	}

	// Check if email already exists
	err = db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&count)
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
	INSERT INTO users (username, password_hash, email, role, company_id, created_at)
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
	claims := JWTClaims{
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