package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB initializes the database connection with optimal settings
func InitDB() error {
	// Read environment variables
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "tracepost")
	sslmode := getEnv("DB_SSLMODE", "disable")
	maxConn := getEnvAsInt("DB_MAX_CONNECTIONS", 20)
	maxIdleConn := getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 5)
	connLifetime := getEnvAsInt("DB_CONNECTION_LIFETIME", 300)

	// Create connection string with additional parameters for performance
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s application_name=tracepost-larvae-api",
		host, port, user, password, dbname, sslmode)

	// Open connection
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(maxConn)
	DB.SetMaxIdleConns(maxIdleConn)
	DB.SetConnMaxLifetime(time.Duration(connLifetime) * time.Second)

	// Check connection with detailed error logging
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	fmt.Printf("Successfully connected to database %s at %s:%s\n", dbname, host, port)

	// Create tables if they don't exist
	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// createTables creates the necessary tables if they don't exist
func createTables() error {
	// Company table - stores organization information
	companyTableQuery := `
	CREATE TABLE IF NOT EXISTS company (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		type VARCHAR(50),
		location TEXT,
		contact_info TEXT,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		is_active BOOLEAN DEFAULT TRUE
	);`

	// Account table - stores user account information
	accountTableQuery := `
	CREATE TABLE IF NOT EXISTS account (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role VARCHAR(20) CHECK (role IN ('admin', 'operator', 'viewer')) NOT NULL,
		company_id INT REFERENCES company(id) ON DELETE CASCADE,
		last_login TIMESTAMP,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		is_active BOOLEAN DEFAULT TRUE
	);`

	// Hatchery table - stores information about shrimp hatcheries
	hatcheryTableQuery := `
	CREATE TABLE IF NOT EXISTS hatchery (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		location TEXT,
		contact TEXT,
		company_id INT REFERENCES company(id),
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		is_active BOOLEAN DEFAULT TRUE
	);`

	// Batch table - stores batch information for shrimp larvae
	batchTableQuery := `
	CREATE TABLE IF NOT EXISTS batch (
		id SERIAL PRIMARY KEY,
		hatchery_id INT REFERENCES hatchery(id) ON DELETE CASCADE,
		species TEXT,
		quantity INT,
		status VARCHAR(30),
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		is_active BOOLEAN DEFAULT TRUE
	);`

	// Event table - stores events related to each batch
	eventTableQuery := `
	CREATE TABLE IF NOT EXISTS event (
		id SERIAL PRIMARY KEY,
		batch_id INT REFERENCES batch(id) ON DELETE CASCADE,
		event_type VARCHAR(50),
		actor_id INT REFERENCES account(id),
		location TEXT,
		timestamp TIMESTAMP DEFAULT NOW(),
		metadata JSONB,
		updated_at TIMESTAMP DEFAULT NOW(),
		is_active BOOLEAN DEFAULT TRUE
	);`
	// Environment data table - stores environmental parameters
	environmentTableQuery := `
	CREATE TABLE IF NOT EXISTS environment (
		id SERIAL PRIMARY KEY,
		batch_id INT REFERENCES batch(id) ON DELETE CASCADE NOT NULL,
		temperature DECIMAL(5,2) NOT NULL,
		pH DECIMAL(4,2) NOT NULL,
		salinity DECIMAL(5,2) NOT NULL,
		dissolved_oxygen DECIMAL(5,2) NOT NULL,
		timestamp TIMESTAMP DEFAULT NOW() NOT NULL,
		updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
		is_active BOOLEAN DEFAULT TRUE NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_environment_batch_id ON environment(batch_id);
	`;
	// Document table - stores document/certificate references
	documentTableQuery := `
	CREATE TABLE IF NOT EXISTS document (
		id SERIAL PRIMARY KEY,
		batch_id INT REFERENCES batch(id) ON DELETE CASCADE NOT NULL,
		doc_type VARCHAR(50) NOT NULL,
		ipfs_hash TEXT NOT NULL,
		uploaded_by INT REFERENCES account(id) NOT NULL,
		uploaded_at TIMESTAMP DEFAULT NOW() NOT NULL,
		updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
		is_active BOOLEAN DEFAULT TRUE NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_document_ipfs_hash ON document(ipfs_hash);
	CREATE INDEX IF NOT EXISTS idx_document_batch_id ON document(batch_id);
	`;	// Blockchain record table - stores blockchain transaction records
	blockchainRecordTableQuery := `
	CREATE TABLE IF NOT EXISTS blockchain_record (
		id SERIAL PRIMARY KEY,
		related_table TEXT NOT NULL,
		related_id INT NOT NULL,
		tx_id TEXT NOT NULL,
		metadata_hash TEXT,
		created_at TIMESTAMP DEFAULT NOW() NOT NULL,
		updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
		is_active BOOLEAN DEFAULT TRUE NOT NULL,
		CONSTRAINT valid_relation CHECK (related_table IS NOT NULL AND related_id IS NOT NULL)
	);
	CREATE INDEX IF NOT EXISTS idx_blockchain_record_tx_id ON blockchain_record(tx_id);
	CREATE INDEX IF NOT EXISTS idx_blockchain_record_related ON blockchain_record(related_table, related_id);
	`;	// Shipment transfer table - stores batch transfer information
	shipmentTransferTableQuery := `
	CREATE TABLE IF NOT EXISTS shipment_transfer (
		id TEXT PRIMARY KEY,
		batch_id INT REFERENCES batch(id) ON DELETE CASCADE NOT NULL,
		source_id TEXT NOT NULL,
		source_type TEXT NOT NULL,
		destination_id TEXT,
		destination_type TEXT,
		quantity INT NOT NULL CHECK (quantity > 0),
		transferred_at TIMESTAMP DEFAULT NOW() NOT NULL,
		transferred_by INT REFERENCES account(id) NOT NULL,
		status VARCHAR(30) DEFAULT 'initiated' NOT NULL CHECK (status IN ('initiated', 'in_transit', 'delivered', 'rejected', 'cancelled')),
		blockchain_tx_id TEXT,
		nft_token_id INT,
		nft_contract_address TEXT,
		transfer_notes TEXT,
		metadata JSONB,
		created_at TIMESTAMP DEFAULT NOW() NOT NULL,
		updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
		is_active BOOLEAN DEFAULT TRUE NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_shipment_transfer_batch_id ON shipment_transfer(batch_id);
	CREATE INDEX IF NOT EXISTS idx_shipment_transfer_status ON shipment_transfer(status);
	`;	// Transaction NFT table - stores NFT information for each transaction
	transactionNFTTableQuery := `
	CREATE TABLE IF NOT EXISTS transaction_nft (
		id SERIAL PRIMARY KEY,
		tx_id TEXT UNIQUE NOT NULL,
		shipment_transfer_id TEXT REFERENCES shipment_transfer(id) ON DELETE CASCADE NOT NULL,
		token_id TEXT NOT NULL,
		contract_address TEXT NOT NULL,
		token_uri TEXT,
		qr_code_url TEXT,
		owner_address TEXT NOT NULL,
		metadata JSONB,
		created_at TIMESTAMP DEFAULT NOW() NOT NULL,
		updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
		is_active BOOLEAN DEFAULT TRUE NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_transaction_nft_token_id ON transaction_nft(token_id);
	CREATE INDEX IF NOT EXISTS idx_transaction_nft_owner ON transaction_nft(owner_address);
	`;

	// Execute the queries
	queries := map[string]string{
		"company":            companyTableQuery,
		"account":            accountTableQuery,
		"hatchery":           hatcheryTableQuery,
		"batch":              batchTableQuery,
		"event":              eventTableQuery,
		"environment":        environmentTableQuery,
		"document":           documentTableQuery,
		"blockchain_record":  blockchainRecordTableQuery,
		"shipment_transfer":  shipmentTransferTableQuery,
		"transaction_nft":    transactionNFTTableQuery,
	}

	for tableName, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to create table %s: %w", tableName, err)
		}
		fmt.Printf("Table %s created or already exists\n", tableName)
	}

	return nil
}

// getEnv retrieves an environment variable or returns a default value if not set
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	} else {
		fmt.Printf("Warning: environment variable %s with value '%s' is not a valid integer, using default value %d\n", 
			key, valueStr, defaultValue)
		return defaultValue
	}
}

// Close closes the database connection
func Close() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			fmt.Printf("Error closing database connection: %v\n", err)
		} else {
			fmt.Println("Database connection closed successfully")
		}
	}
}