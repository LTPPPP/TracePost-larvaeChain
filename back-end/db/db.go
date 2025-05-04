package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB() error {
	// Read environment variables
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "tracepost")
	sslmode := getEnv("DB_SSLMODE", "disable")

	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Open connection
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	// Check connection
	if err = DB.Ping(); err != nil {
		return err
	}

	// Create tables if they don't exist
	if err = createTables(); err != nil {
		return err
	}

	return nil
}

// createTables creates the necessary tables if they don't exist
func createTables() error {
	// Batch table - stores batch information for shrimp larvae
	batchTableQuery := `
	CREATE TABLE IF NOT EXISTS batches (
		id SERIAL PRIMARY KEY,
		batch_id VARCHAR(50) UNIQUE NOT NULL,
		hatchery_id VARCHAR(50) NOT NULL,
		creation_date TIMESTAMP NOT NULL DEFAULT NOW(),
		species VARCHAR(100) NOT NULL,
		quantity INT NOT NULL,
		status VARCHAR(50) NOT NULL,
		blockchain_tx_id VARCHAR(100) NOT NULL,
		metadata_hash VARCHAR(100) NOT NULL
	);`

	// Event table - stores events related to each batch
	eventTableQuery := `
	CREATE TABLE IF NOT EXISTS events (
		id SERIAL PRIMARY KEY,
		batch_id VARCHAR(50) REFERENCES batches(batch_id),
		event_type VARCHAR(50) NOT NULL,
		timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
		location VARCHAR(100) NOT NULL,
		actor_id VARCHAR(50) NOT NULL,
		details JSONB NOT NULL,
		blockchain_tx_id VARCHAR(100) NOT NULL,
		metadata_hash VARCHAR(100) NOT NULL
	);`

	// Document table - stores document/certificate references
	documentTableQuery := `
	CREATE TABLE IF NOT EXISTS documents (
		id SERIAL PRIMARY KEY,
		batch_id VARCHAR(50) REFERENCES batches(batch_id),
		document_type VARCHAR(50) NOT NULL,
		ipfs_hash VARCHAR(100) NOT NULL,
		upload_date TIMESTAMP NOT NULL DEFAULT NOW(),
		issuer VARCHAR(100) NOT NULL,
		is_verified BOOLEAN NOT NULL DEFAULT FALSE,
		blockchain_tx_id VARCHAR(100) NOT NULL
	);`

	// Environment data table - stores environment parameters
	envDataTableQuery := `
	CREATE TABLE IF NOT EXISTS environment_data (
		id SERIAL PRIMARY KEY,
		batch_id VARCHAR(50) REFERENCES batches(batch_id),
		timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
		temperature DECIMAL(5,2) NOT NULL,
		ph DECIMAL(4,2) NOT NULL,
		salinity DECIMAL(5,2) NOT NULL,
		dissolved_oxygen DECIMAL(5,2) NOT NULL,
		other_params JSONB,
		blockchain_tx_id VARCHAR(100) NOT NULL
	);`

	// User table - for API access and authentication
	userTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password_hash VARCHAR(100) NOT NULL,
		role VARCHAR(20) NOT NULL,
		company_id VARCHAR(50) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		last_login TIMESTAMP
	);`

	// Execute the queries
	queries := []string{
		batchTableQuery,
		eventTableQuery,
		documentTableQuery,
		envDataTableQuery,
		userTableQuery,
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

// getEnv retrieves an environment variable or returns a default value if not set
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Close closes the database connection
func Close() error {
	return DB.Close()
}