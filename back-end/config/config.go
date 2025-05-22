package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents the application configuration
type Config struct {
	ServerPort    string
	ServerTimeout int
	ServerHost    string
	BaseURL       string

	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBSSLMode            string
	DBMaxConnections     int
	DBMaxIdleConnections int
	DBConnectionLifetime int

	BlockchainNodeURL     string
	BlockchainChainID     string
	BlockchainAccount     string
	BlockchainKeyFile     string
	BlockchainConsensus   string
	BlockchainContractAddr string
	BlockchainPrivateKey  string
	BlockchainNetworkID   string

	InteropEnabled        bool
	InteropRelayEndpoint  string
	InteropAllowedChains  []string
	InteropDefaultStandard string
	IBCEnabled            bool
	SubstrateEnabled      bool

	IdentityEnabled       bool
	IdentityRegistryAddr  string
	IdentityResolverURL   string
	IdentityRegistryContract string

	IPFSNodeURL   string
	IPFSGatewayURL string
	IPFSAPIKey    string
	JWTSecret     string
	JWTExpiration int
	JWTIssuer     string
	RateLimitRequests int
	RateLimitDuration int

	LogLevel  string
	LogFormat string
	LogFile   string

	EnableMetrics bool
	MetricsPort   string

	Environment string
}

// Load loads the configuration from environment variables
func Load() *Config {
	return &Config{
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		ServerTimeout: getEnvAsInt("SERVER_TIMEOUT", 30),
		ServerHost:    getEnv("SERVER_HOST", "0.0.0.0"),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),

		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPassword:           getEnv("DB_PASSWORD", "postgres"),
		DBName:               getEnv("DB_NAME", "tracepost"),
		DBSSLMode:            getEnv("DB_SSLMODE", "disable"),
		DBMaxConnections:     getEnvAsInt("DB_MAX_CONNECTIONS", 20),
		DBMaxIdleConnections: getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 5),
		DBConnectionLifetime: getEnvAsInt("DB_CONNECTION_LIFETIME", 300),
		BlockchainNodeURL:     getEnv("BLOCKCHAIN_NODE_URL", "http://localhost:26657"),
		BlockchainChainID:     getEnv("BLOCKCHAIN_CHAIN_ID", "tracepost-chain"),
		BlockchainAccount:     getEnv("BLOCKCHAIN_ACCOUNT", "tracepost"),
		BlockchainKeyFile:     getEnv("BLOCKCHAIN_KEY_FILE", ""),
		BlockchainConsensus:   getEnv("BLOCKCHAIN_CONSENSUS", "poa"),
		BlockchainContractAddr: getEnv("BLOCKCHAIN_CONTRACT_ADDRESS", ""),
		BlockchainPrivateKey:   getEnv("BLOCKCHAIN_PRIVATE_KEY", ""),
		BlockchainNetworkID:    getEnv("BLOCKCHAIN_NETWORK_ID", "tracepost-network"),

		InteropEnabled:        getEnvAsBool("INTEROP_ENABLED", false),
		InteropRelayEndpoint:  getEnv("INTEROP_RELAY_ENDPOINT", ""),
		InteropAllowedChains:  getEnvAsStringSlice("INTEROP_ALLOWED_CHAINS", []string{}),
		InteropDefaultStandard: getEnv("INTEROP_DEFAULT_STANDARD", "GS1-EPCIS"),
		IBCEnabled:            getEnvAsBool("IBC_ENABLED", false),
		SubstrateEnabled:      getEnvAsBool("SUBSTRATE_ENABLED", false),

		IdentityEnabled:          getEnvAsBool("IDENTITY_ENABLED", false),
		IdentityRegistryAddr:     getEnv("IDENTITY_REGISTRY_ADDRESS", ""),
		IdentityResolverURL:      getEnv("IDENTITY_RESOLVER_URL", ""),
		IdentityRegistryContract: getEnv("IDENTITY_REGISTRY_CONTRACT", ""),

		IPFSNodeURL:    getEnv("IPFS_NODE_URL", "http://localhost:5001"),
		IPFSGatewayURL: getEnv("IPFS_GATEWAY_URL", "http://localhost:8080"),
		IPFSAPIKey:     getEnv("IPFS_API_KEY", ""),

		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpiration: getEnvAsInt("JWT_EXPIRATION", 24),
		JWTIssuer:     getEnv("JWT_ISSUER", "tracepost-larvae-api"),

		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),
		LogFile:   getEnv("LOG_FILE", "app.log"),

		RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitDuration: getEnvAsInt("RATE_LIMIT_DURATION", 60),

		EnableMetrics: getEnvAsBool("ENABLE_METRICS", true),
		MetricsPort:   getEnv("METRICS_PORT", "9090"),

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
		case "ShardingEnabled":
			if boolVal, ok := value.(bool); ok {
				c.InteropEnabled = boolVal
			}
		}
	}
}

// GetJWTSecret retrieves the JWT secret from the configured source
func GetJWTSecret() (string, error) {
	cfg := GetConfig()
	secret := cfg.JWTSecret
	
	if strings.HasPrefix(secret, "file:") {
		filePath := strings.TrimPrefix(secret, "file:")
		
		data, err := os.ReadFile(filePath)
		if err != nil {
			envSecret := os.Getenv("JWT_SECRET_VALUE")
			if envSecret != "" {
				return envSecret, nil
			}
			return "", fmt.Errorf("failed to read JWT secret from file %s: %v", filePath, err)
		}
		
		return strings.TrimSpace(string(data)), nil
	}
	
	return secret, nil
}