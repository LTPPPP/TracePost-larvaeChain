package middleware

import (
	"fmt"
	"os"
	"strings"
	"time"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/LTPPPP/TracePost-larvaeChain/config"
	"github.com/LTPPPP/TracePost-larvaeChain/models"
)

// tokenBlacklist stores revoked token IDs with their expiry time
var (
	tokenBlacklist = make(map[string]time.Time)
	blacklistMutex sync.RWMutex
)

// init function starts a background goroutine to clean up expired tokens from the blacklist
func init() {
	go cleanupBlacklist()
}

// cleanupBlacklist runs every hour to remove expired tokens from the blacklist
func cleanupBlacklist() {
	for {
		time.Sleep(1 * time.Hour)
		
		blacklistMutex.Lock()
		now := time.Now()
		for tokenID, expiry := range tokenBlacklist {
			if now.After(expiry) {
				delete(tokenBlacklist, tokenID)
			}
		}
		blacklistMutex.Unlock()
	}
}

// RevokeToken adds a token to the blacklist
// Should be called when a user logs out or changes password
func RevokeToken(tokenID string, expiryTime time.Time) {
	blacklistMutex.Lock()
	defer blacklistMutex.Unlock()
	
	tokenBlacklist[tokenID] = expiryTime
}

// IsTokenRevoked checks if a token is in the blacklist
func IsTokenRevoked(tokenID string) bool {
	blacklistMutex.RLock()
	defer blacklistMutex.RUnlock()
	
	_, found := tokenBlacklist[tokenID]
	return found
}

// JWTMiddleware is a middleware that verifies JWT tokens
func JWTMiddleware() fiber.Handler {
	// Get configuration
	cfg := config.GetConfig()
	issuer := cfg.JWTIssuer
	// Get secret key
	secretKey, err := config.GetJWTSecret()
	if err != nil {
		// Log error and use fallback mechanism
		fmt.Printf("Error loading JWT secret: %v, using fallback\n", err)
		
		// Try environment variable directly
		envSecret := os.Getenv("JWT_SECRET")
		if envSecret != "" && !strings.HasPrefix(envSecret, "file:") {
			secretKey = envSecret
		} else {
			// Last resort - use a generated secret (not secure for production)
			secretKey = fmt.Sprintf("TEMP_KEY_%d", time.Now().UnixNano())
			fmt.Printf("WARNING: Using temporary JWT key. Authentication will be reset on server restart.\n")
		}
	}
	
	secretKeyBytes := []byte(secretKey)

	return func(c *fiber.Ctx) error {
		// Skip auth for OPTIONS requests (CORS preflight)
		if c.Method() == "OPTIONS" {
			return c.Next()
		}

		// Get authorization header
		authHeader := c.Get("Authorization")
		
		// Check if authorization header exists
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Authorization header is required. Please include a Bearer token.")
		}

		// Check if authorization header has correct format
		if (!strings.HasPrefix(authHeader, "Bearer ")) {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid authorization format. Format should be 'Bearer your-token'.")
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
				// Parse and validate token with claims
		token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			
			// Return secret key from config
			return secretKeyBytes, nil
		})
		
		if err != nil {
			// Provide detailed error messages for different validation errors
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorMalformed != 0 {
					return fiber.NewError(fiber.StatusUnauthorized, "Token is malformed")
				} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
					return fiber.NewError(fiber.StatusUnauthorized, "Token has expired or is not yet valid")
				} else if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
					return fiber.NewError(fiber.StatusUnauthorized, "Token signature is invalid")
				} else {
					return fiber.NewError(fiber.StatusUnauthorized, "Token validation error")
				}
			}
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
		}
		
		// Check if token is valid
		if !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
		}
		
		// Type assert claims
		claims, ok := token.Claims.(*models.JWTClaims)
		if !ok {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse token claims")
		}
		
		// Verify issuer if configured
		if issuer != "" && claims.Issuer != issuer {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token issuer")
		}
		
		// Check if token is revoked
		if IsTokenRevoked(claims.ID) {
			return fiber.NewError(fiber.StatusUnauthorized, "Token has been revoked")
		}
		
		// Set claims to context for use in handlers
		c.Locals("userID", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)
		c.Locals("companyID", claims.CompanyID)
		c.Locals("user", claims)
		
		// Proceed to handler
		return c.Next()
	}
}

// RoleMiddleware is a middleware that checks if a user has the required role
func RoleMiddleware(requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user info from context
		username, okUsername := c.Locals("username").(string)
		role, okRole := c.Locals("role").(string)
		
		// Check if user role is available
		if !okRole {
			return fiber.NewError(fiber.StatusUnauthorized, "User role not found. Authentication may be incomplete.")
		}
		
		// Convert required roles to a readable format for error messages
		readableRoles := strings.Join(requiredRoles, "', '")
		readableRoles = "'" + readableRoles + "'"
		
		// Check if user has required role
		hasRole := false
		for _, requiredRole := range requiredRoles {
			if role == requiredRole {
				hasRole = true
				break
			}
		}
		
		if !hasRole {
			userInfo := ""
			if okUsername {
				userInfo = "User '" + username + "'"
			} else {
				userInfo = "Current user"
			}
			
			return fiber.NewError(
				fiber.StatusForbidden, 
				fmt.Sprintf("%s with role '%s' does not have sufficient permissions. Required role(s): %s.", 
					userInfo, role, readableRoles),
			)
		}
		
		// Continue
		return c.Next()
	}
}

// LoggerMiddleware is a middleware that logs requests
func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Record start time
		start := time.Now()
		
		// Process request
		err := c.Next()
		
		// Calculate request duration
		duration := time.Since(start)
		
		// Log request
		// In a real application, this would use a proper logging framework
		// and would be integrated with OpenTelemetry
		statusCode := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		ip := c.IP()
		userAgent := c.Get("User-Agent")
		
		// Create structured log entry (simplified for this example)
		logEntry := map[string]interface{}{
			"timestamp":  time.Now().Format(time.RFC3339),
			"duration":   duration.String(),
			"status":     statusCode,
			"method":     method,
			"path":       path,
			"ip":         ip,
			"user_agent": userAgent,
		}
		
		// Add user ID if available
		if userId, ok := c.Locals("userId").(int); ok {
			logEntry["user_id"] = userId
		}
		
		// Log entry
		// In a real application, this would use a proper logging framework
		// For now, we'll just print it to stdout
		// fmt.Printf("%+v\n", logEntry)
		
		return err
	}
}

// RateLimitMiddleware implements rate limiting for API endpoints
func RateLimitMiddleware() fiber.Handler {
	// Get configuration
	cfg := config.GetConfig()
	maxRequests := cfg.RateLimitRequests
	windowDuration := time.Duration(cfg.RateLimitDuration) * time.Second
	
	// IP-based request counters with expiration
	type client struct {
		count     int
		lastReset time.Time
	}
	
	// Thread-safe map for storing client request counts
	var (
		clients = make(map[string]*client)
		mu      sync.Mutex
	)
	
	// Start a background goroutine to clean up expired clients
	// This prevents memory leaks from storing too many IP addresses
	go func() {
		for {
			time.Sleep(time.Minute)
			
			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastReset) > windowDuration*2 {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
	
	return func(c *fiber.Ctx) error {
		// Get client IP (considering trusted proxies)
		ip := c.IP()
		
		mu.Lock()
		defer mu.Unlock()
		
		// Get or create client
		cl, exists := clients[ip]
		if !exists {
			clients[ip] = &client{
				count:     0,
				lastReset: time.Now(),
			}
			cl = clients[ip]
		}
		
		// Check if window expired and reset if needed
		if time.Since(cl.lastReset) > windowDuration {
			cl.count = 0
			cl.lastReset = time.Now()
		}
		
		// Increment request count
		cl.count++
		
		// Check if rate limit exceeded
		if cl.count > maxRequests {
			// Return 429 Too Many Requests with proper headers
			c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("Retry-After", fmt.Sprintf("%d", int(windowDuration.Seconds() - time.Since(cl.lastReset).Seconds())))
			
			return fiber.NewError(fiber.StatusTooManyRequests, "Rate limit exceeded")
		}
		
		// Set rate limit headers
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", maxRequests-cl.count))
		
		return c.Next()
	}
}