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

var (
	tokenBlacklist = make(map[string]time.Time)
	blacklistMutex sync.RWMutex
)

func init() {
	go cleanupBlacklist()
}

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

func RevokeToken(tokenID string, expiryTime time.Time) {
	blacklistMutex.Lock()
	defer blacklistMutex.Unlock()
	
	tokenBlacklist[tokenID] = expiryTime
}

func IsTokenRevoked(tokenID string) bool {
	blacklistMutex.RLock()
	defer blacklistMutex.RUnlock()
	
	_, found := tokenBlacklist[tokenID]
	return found
}

func JWTMiddleware() fiber.Handler {
	cfg := config.GetConfig()
	issuer := cfg.JWTIssuer
	secretKey, err := config.GetJWTSecret()
	if err != nil {
		fmt.Printf("Error loading JWT secret: %v, using fallback\n", err)
		
		envSecret := os.Getenv("JWT_SECRET")
		if envSecret != "" && !strings.HasPrefix(envSecret, "file:") {
			secretKey = envSecret
		} else {
			secretKey = fmt.Sprintf("TEMP_KEY_%d", time.Now().UnixNano())
			fmt.Printf("WARNING: Using temporary JWT key. Authentication will be reset on server restart.\n")
		}
	}
	
	secretKeyBytes := []byte(secretKey)

	return func(c *fiber.Ctx) error {
		if c.Method() == "OPTIONS" {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Authorization header is required. Please include a Bearer token.")
		}

		if (!strings.HasPrefix(authHeader, "Bearer ")) {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid authorization format. Format should be 'Bearer your-token'.")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
				token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			
			return secretKeyBytes, nil
		})
		
		if err != nil {
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
		
		if !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
		}
		
		claims, ok := token.Claims.(*models.JWTClaims)
		if !ok {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse token claims")
		}
		
		if issuer != "" && claims.Issuer != issuer {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token issuer")
		}
		
		if IsTokenRevoked(claims.ID) {
			return fiber.NewError(fiber.StatusUnauthorized, "Token has been revoked")
		}
		
		c.Locals("userID", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)
		c.Locals("companyID", claims.CompanyID)
		c.Locals("user", claims)
		
		return c.Next()
	}
}

func RoleMiddleware(requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		username, okUsername := c.Locals("username").(string)
		role, okRole := c.Locals("role").(string)
		
		if !okRole {
			return fiber.NewError(fiber.StatusUnauthorized, "User role not found. Authentication may be incomplete.")
		}
		
		readableRoles := strings.Join(requiredRoles, "', '")
		readableRoles = "'" + readableRoles + "'"
		
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
		
		return c.Next()
	}
}

func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		
		err := c.Next()
		
		duration := time.Since(start)
		
		statusCode := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		ip := c.IP()
		userAgent := c.Get("User-Agent")
		
		logEntry := map[string]interface{}{
			"timestamp":  time.Now().Format(time.RFC3339),
			"duration":   duration.String(),
			"status":     statusCode,
			"method":     method,
			"path":       path,
			"ip":         ip,
			"user_agent": userAgent,
		}
		
		if userId, ok := c.Locals("userId").(int); ok {
			logEntry["user_id"] = userId
		}
		
		return err
	}
}

func RateLimitMiddleware() fiber.Handler {
	cfg := config.GetConfig()
	maxRequests := cfg.RateLimitRequests
	windowDuration := time.Duration(cfg.RateLimitDuration) * time.Second
	
	type client struct {
		count     int
		lastReset time.Time
	}
	
	var (
		clients = make(map[string]*client)
		mu      sync.Mutex
	)
	
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
		ip := c.IP()
		
		mu.Lock()
		defer mu.Unlock()
		
		cl, exists := clients[ip]
		if !exists {
			clients[ip] = &client{
				count:     0,
				lastReset: time.Now(),
			}
			cl = clients[ip]
		}
		
		if time.Since(cl.lastReset) > windowDuration {
			cl.count = 0
			cl.lastReset = time.Now()
		}
		
		cl.count++
		
		if cl.count > maxRequests {
			c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("Retry-After", fmt.Sprintf("%d", int(windowDuration.Seconds() - time.Since(cl.lastReset).Seconds())))
			
			return fiber.NewError(fiber.StatusTooManyRequests, "Rate limit exceeded")
		}
		
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", maxRequests-cl.count))
		
		return c.Next()
	}
}