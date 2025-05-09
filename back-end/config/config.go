package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents the application configuration
type Config struct {
	// Server configuration
	ServerPort    string
	ServerTimeout int
	ServerHost    string
	BaseURL       string

	// Database configuration
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBSSLMode            string
	DBMaxConnections     int
	DBMaxIdleConnections int
	DBConnectionLifetime int

	// Blockchain configuration
	BlockchainNodeURL     string
	BlockchainChainID     string
	BlockchainAccount     string
	BlockchainKeyFile     string
	BlockchainConsensus   string
	BlockchainContractAddr string
	BlockchainPrivateKey  string
	BlockchainNetworkID   string

	// Interoperability configuration
	InteropEnabled        bool
	InteropRelayEndpoint  string
	InteropAllowedChains  []string
	InteropDefaultStandard string

	// Identity configuration
	IdentityEnabled       bool
	IdentityRegistryAddr  string
	IdentityResolverURL   string
	IdentityRegistryContract string

	// IPFS configuration
	IPFSNodeURL   string
	IPFSGatewayURL string
	IPFSAPIKey    string
	// JWT configuration
	JWTSecret     string
	JWTExpiration int
	JWTIssuer     string
	// Rate limiting configuration
	RateLimitRequests int
	RateLimitDuration int

	// Logging configuration
	LogLevel  string
	LogFormat string
	LogFile   string

	// Metrics
	EnableMetrics bool
	MetricsPort   string

	// Environment
	Environment string
}

// Load loads the configuration from environment variables
func Load() *Config {
	return &Config{
		// Server configuration
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		ServerTimeout: getEnvAsInt("SERVER_TIMEOUT", 30),
		ServerHost:    getEnv("SERVER_HOST", "0.0.0.0"),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),

		// Database configuration
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPassword:           getEnv("DB_PASSWORD", "postgres"),
		DBName:               getEnv("DB_NAME", "tracepost"),
		DBSSLMode:            getEnv("DB_SSLMODE", "disable"),
		DBMaxConnections:     getEnvAsInt("DB_MAX_CONNECTIONS", 20),
		DBMaxIdleConnections: getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 5),
		DBConnectionLifetime: getEnvAsInt("DB_CONNECTION_LIFETIME", 300),
		// Blockchain configuration
		BlockchainNodeURL:     getEnv("BLOCKCHAIN_NODE_URL", "http://localhost:26657"),
		BlockchainChainID:     getEnv("BLOCKCHAIN_CHAIN_ID", "tracepost-chain"),
		BlockchainAccount:     getEnv("BLOCKCHAIN_ACCOUNT", "tracepost"),
		BlockchainKeyFile:     getEnv("BLOCKCHAIN_KEY_FILE", ""),
		BlockchainConsensus:   getEnv("BLOCKCHAIN_CONSENSUS", "poa"),
		BlockchainContractAddr: getEnv("BLOCKCHAIN_CONTRACT_ADDRESS", ""),
		BlockchainPrivateKey:   getEnv("BLOCKCHAIN_PRIVATE_KEY", ""),
		BlockchainNetworkID:    getEnv("BLOCKCHAIN_NETWORK_ID", "tracepost-network"),

		// Interoperability configuration
		InteropEnabled:        getEnvAsBool("INTEROP_ENABLED", false),
		InteropRelayEndpoint:  getEnv("INTEROP_RELAY_ENDPOINT", ""),
		InteropAllowedChains:  getEnvAsStringSlice("INTEROP_ALLOWED_CHAINS", []string{}),
		InteropDefaultStandard: getEnv("INTEROP_DEFAULT_STANDARD", "GS1-EPCIS"),

		// Identity configuration
		IdentityEnabled:          getEnvAsBool("IDENTITY_ENABLED", false),
		IdentityRegistryAddr:     getEnv("IDENTITY_REGISTRY_ADDRESS", ""),
		IdentityResolverURL:      getEnv("IDENTITY_RESOLVER_URL", ""),
		IdentityRegistryContract: getEnv("IDENTITY_REGISTRY_CONTRACT", ""),

		// IPFS configuration
		IPFSNodeURL:    getEnv("IPFS_NODE_URL", "http://localhost:5001"),
		IPFSGatewayURL: getEnv("IPFS_GATEWAY_URL", "http://localhost:8080"),
		IPFSAPIKey:     getEnv("IPFS_API_KEY", ""),

		// JWT configuration
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpiration: getEnvAsInt("JWT_EXPIRATION", 24),
		JWTIssuer:     getEnv("JWT_ISSUER", "tracepost-larvae-api"),

		// Logging configuration
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),
		LogFile:   getEnv("LOG_FILE", "app.log"),

		// Rate limiting
		RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitDuration: getEnvAsInt("RATE_LIMIT_DURATION", 60),

		// Metrics
		EnableMetrics: getEnvAsBool("ENABLE_METRICS", true),
		MetricsPort:   getEnv("METRICS_PORT", "9090"),

		// Environment
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsBool gets an environment variable as a boolean or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsStringSlice gets an environment variable as a string slice or returns a default value
func getEnvAsStringSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}

// GetConfig returns the application configuration
func GetConfig() *Config {
	return Load()
}

// UpdateConfig updates the configuration with new values
func (c *Config) UpdateConfig(updates map[string]interface{}) {
	// This is a simplified implementation that only handles string values
	// In a real application, this would need to handle different types
	for key, value := range updates {
		switch key {
		case "BaseURL":
			if strVal, ok := value.(string); ok {
				c.BaseURL = strVal
			}
		case "BlockchainPrivateKey":
			if strVal, ok := value.(string); ok {
				c.BlockchainPrivateKey = strVal
			}
		case "IdentityRegistryContract":
			if strVal, ok := value.(string); ok {
				c.IdentityRegistryContract = strVal
			}
		case "LayerTwoEnabled":
			if boolVal, ok := value.(bool); ok {
				c.InteropEnabled = boolVal
			}
		case "ShardingEnabled":
			if boolVal, ok := value.(bool); ok {
				c.InteropEnabled = boolVal
			}
		// Add more cases as needed
		}
	}
}

// GetJWTSecret retrieves the JWT secret from the configured source
func GetJWTSecret() (string, error) {
	cfg := GetConfig()
	secret := cfg.JWTSecret
	
	// Check if secret is a file path
	if strings.HasPrefix(secret, "file:") {
		// Extract file path
		filePath := strings.TrimPrefix(secret, "file:")
		
		// Read secret from file
		data, err := os.ReadFile(filePath)
		if err != nil {
			// Fallback to environment variable if file doesn't exist
			envSecret := os.Getenv("JWT_SECRET_VALUE")
			if envSecret != "" {
				return envSecret, nil
			}
			return "", fmt.Errorf("failed to read JWT secret from file %s: %v", filePath, err)
		}
		
		// Trim whitespace and return
		return strings.TrimSpace(string(data)), nil
	}
	
	return secret, nil
}