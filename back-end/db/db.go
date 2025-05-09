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
		batch_id INT REFERENCES batch(id),
		temperature DECIMAL(5,2),
		pH DECIMAL(4,2),
		salinity DECIMAL(5,2),
		dissolved_oxygen DECIMAL(5,2),
		timestamp TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		is_active BOOLEAN DEFAULT TRUE
	);`

	// Document table - stores document/certificate references
	documentTableQuery := `
	CREATE TABLE IF NOT EXISTS document (
		id SERIAL PRIMARY KEY,
		batch_id INT REFERENCES batch(id) ON DELETE CASCADE,
		doc_type VARCHAR(50),
		ipfs_hash TEXT NOT NULL,
		uploaded_by INT REFERENCES account(id),
		uploaded_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		is_active BOOLEAN DEFAULT TRUE
	);`
	// Blockchain record table - stores blockchain transaction records
	blockchainRecordTableQuery := `
	CREATE TABLE IF NOT EXISTS blockchain_record (
		id SERIAL PRIMARY KEY,
		related_table TEXT,
		related_id INT,
		tx_id TEXT NOT NULL,
		metadata_hash TEXT,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		is_active BOOLEAN DEFAULT TRUE
	);`

	// Shipment transfer table - stores batch transfer information
	shipmentTransferTableQuery := `
	CREATE TABLE IF NOT EXISTS shipment_transfer (
		id TEXT PRIMARY KEY,
		batch_id INT REFERENCES batch(id) ON DELETE CASCADE,
		source_id TEXT NOT NULL,
		source_type TEXT NOT NULL,
		destination_id TEXT,
		destination_type TEXT,
		quantity INT NOT NULL,
		transferred_at TIMESTAMP DEFAULT NOW(),
		transferred_by TEXT REFERENCES account(id),
		status VARCHAR(30) DEFAULT 'initiated',
		blockchain_tx_id TEXT,
		nft_token_id INT,
		nft_contract_address TEXT,
		transfer_notes TEXT,
		metadata JSONB,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		is_active BOOLEAN DEFAULT TRUE
	);`
	// Transaction NFT table - stores NFT information for each transaction
	transactionNFTTableQuery := `
	CREATE TABLE IF NOT EXISTS transaction_nft (
		id SERIAL PRIMARY KEY,
		tx_id TEXT UNIQUE NOT NULL,
		shipment_transfer_id TEXT REFERENCES shipment_transfer(id) ON DELETE CASCADE,
		token_id TEXT NOT NULL,
		contract_address TEXT NOT NULL,
		token_uri TEXT,
		qr_code_url TEXT,
		owner_address TEXT NOT NULL,
		metadata JSONB,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		is_active BOOLEAN DEFAULT TRUE
	);`

	// Execute the queries
	queries := []string{
		companyTableQuery,
		accountTableQuery,
		hatcheryTableQuery,
		batchTableQuery,
		eventTableQuery,
		environmentTableQuery,
		documentTableQuery,
		blockchainRecordTableQuery,
		shipmentTransferTableQuery,
		transactionNFTTableQuery,
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