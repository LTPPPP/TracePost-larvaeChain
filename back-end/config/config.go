package config

import (
	"os"
	"strconv"
)

// Config represents the application configuration
type Config struct {
	// Server configuration
	ServerPort    string
	ServerTimeout int

	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Blockchain configuration
	BlockchainNodeURL string
	BlockchainChainID string
	BlockchainAccount string
	BlockchainKeyFile string

	// IPFS configuration
	IPFSNodeURL string

	// JWT configuration
	JWTSecret     string
	JWTExpiration int

	// Logging configuration
	LogLevel  string
	LogFormat string
}

// Load loads the configuration from environment variables
func Load() *Config {
	return &Config{
		// Server configuration
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		ServerTimeout: getEnvAsInt("SERVER_TIMEOUT", 30),

		// Database configuration
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "tracepost"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// Blockchain configuration
		BlockchainNodeURL: getEnv("BLOCKCHAIN_NODE_URL", "http://localhost:26657"),
		BlockchainChainID: getEnv("BLOCKCHAIN_CHAIN_ID", "tracepost-chain"),
		BlockchainAccount: getEnv("BLOCKCHAIN_ACCOUNT", "tracepost"),
		BlockchainKeyFile: getEnv("BLOCKCHAIN_KEY_FILE", ""),

		// IPFS configuration
		IPFSNodeURL: getEnv("IPFS_NODE_URL", "http://localhost:5001"),

		// JWT configuration
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpiration: getEnvAsInt("JWT_EXPIRATION", 24),

		// Logging configuration
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),
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

// GetConfig returns the application configuration
func GetConfig() *Config {
	return Load()
}