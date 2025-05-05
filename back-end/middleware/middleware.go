package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/vietchain/tracepost-larvae/models"
)

// JWTMiddleware is a middleware that verifies JWT tokens
func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")
		
		// Check if authorization header exists
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Authorization header is required")
		}

		// Check if authorization header has correct format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid authorization header format")
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		
		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token signing method")
			}
			
			// Return secret key
			// In a real application, this would be properly configured
			return []byte("your-secret-key"), nil
		})
		
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid or expired token")
		}
		
		// Check if token is valid
		if !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
		}
		
		// Extract claims
		claims, ok := token.Claims.(*models.JWTClaims)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token claims")
		}
		
		// Check if token is expired
		if claims.ExpiresAt.Time.Before(time.Now()) {
			return fiber.NewError(fiber.StatusUnauthorized, "Token is expired")
		}
		
		// Set user data in context
		c.Locals("userId", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)
		c.Locals("companyId", claims.CompanyID)
		
		// Continue
		return c.Next()
	}
}

// RoleMiddleware is a middleware that checks if a user has the required role
func RoleMiddleware(requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user role from context
		role, ok := c.Locals("role").(string)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "User role not found")
		}
		
		// Check if user has required role
		hasRole := false
		for _, requiredRole := range requiredRoles {
			if role == requiredRole {
				hasRole = true
				break
			}
		}
		
		if !hasRole {
			return fiber.NewError(fiber.StatusForbidden, "Insufficient permissions")
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