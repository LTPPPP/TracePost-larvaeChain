package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"context"
)

var (
	DB       *sql.DB
	Redis    *redis.Client
	dbInitMu sync.Mutex
	dbInitialized bool
)

// InitDB initializes the database connection with optimal settings
func InitDB() error {
	// Use mutex to prevent concurrent initialization
	dbInitMu.Lock()
	defer dbInitMu.Unlock()

	// Skip if already initialized
	if dbInitialized && DB != nil {
		return nil
	}

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
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s application_name=tracepost-larvae-api connect_timeout=10",
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
		DB = nil // Reset DB if connection failed
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	fmt.Printf("Successfully connected to database %s at %s:%s\n", dbname, host, port)
	// Create tables if they don't exist
	if err = createTables(); err != nil {
		DB = nil // Reset DB if table creation failed
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// Initialize Redis
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)
	Redis = redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})
	if err := Redis.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	fmt.Printf("Successfully connected to Redis at %s\n", redisAddr)

	// Mark as initialized
	dbInitialized = true
	
	return nil
}

// createTables creates the necessary tables if they don't exist
func createTables() error {
	// Define table creation queries
	tableQueries := map[string]string{
		"company": `
			CREATE TABLE IF NOT EXISTS company (
				id SERIAL PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				type VARCHAR(100),
				location TEXT,
				contact_info TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				is_active BOOLEAN DEFAULT TRUE
			);
		`,
		"account": `
			CREATE TABLE IF NOT EXISTS account (
				id SERIAL PRIMARY KEY,
				username VARCHAR(255) UNIQUE NOT NULL,
				company_id INTEGER REFERENCES company(id),
				full_name VARCHAR(255),
				email VARCHAR(255) UNIQUE NOT NULL,
				password_hash VARCHAR(255) NOT NULL,
				role VARCHAR(50) NOT NULL,
				phone_number VARCHAR(50),
				date_of_birth DATE,
				avatar_url TEXT,
				is_active BOOLEAN DEFAULT TRUE,
				last_login TIMESTAMP,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
		`,
		"hatchery": `
			CREATE TABLE IF NOT EXISTS hatchery (
				id SERIAL PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				company_id INTEGER REFERENCES company(id),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				is_active BOOLEAN DEFAULT TRUE
			);
		`,
		"batch": `
			CREATE TABLE IF NOT EXISTS batch (
				id SERIAL PRIMARY KEY,
				hatchery_id INTEGER REFERENCES hatchery(id),
				species VARCHAR(100),
				quantity INTEGER,
				status VARCHAR(50),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				is_active BOOLEAN DEFAULT TRUE
			);
		`,
		"api_logs": `
			CREATE TABLE IF NOT EXISTS api_logs (
			id SERIAL PRIMARY KEY,
			endpoint VARCHAR(255) NOT NULL,
			method VARCHAR(10) NOT NULL,
			user_id INTEGER,
			status_code INTEGER,
			response_time FLOAT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		`,
		"event": `
			CREATE TABLE IF NOT EXISTS event (
				id SERIAL PRIMARY KEY,
				batch_id INTEGER REFERENCES batch(id),
				event_type VARCHAR(100),
				actor_id INTEGER REFERENCES account(id),
				location TEXT,
				timestamp TIMESTAMP,
				metadata JSONB,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				is_active BOOLEAN DEFAULT TRUE
			);
		`,
		"environment_data": `
			CREATE TABLE IF NOT EXISTS environment_data (
				id SERIAL PRIMARY KEY,
				batch_id INTEGER REFERENCES batch(id),
				temperature FLOAT,
				ph FLOAT,
				salinity FLOAT,
				density FLOAT,
				age INTEGER,
				timestamp TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				is_active BOOLEAN DEFAULT TRUE
			);
		`,
		"document": `
			CREATE TABLE IF NOT EXISTS document (
				id SERIAL PRIMARY KEY,
				batch_id INTEGER REFERENCES batch(id),
				doc_type VARCHAR(100),
				file_name TEXT,
				file_size INTEGER,
				ipfs_hash TEXT,
				ipfs_uri TEXT,
				uploaded_by INTEGER REFERENCES account(id),
				expiry_date TIMESTAMP,
				uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				is_active BOOLEAN DEFAULT TRUE
			);
		`,
		"certificates": `
			CREATE TABLE IF NOT EXISTS certificates (
				id SERIAL PRIMARY KEY,
				batch_id INTEGER REFERENCES batch(id),
				company_id INTEGER REFERENCES company(id),
				certificate_type VARCHAR(100) NOT NULL,
				issuer VARCHAR(255) NOT NULL,
				issue_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				expiry_date TIMESTAMP,
				status VARCHAR(50) NOT NULL,
				document_id INTEGER REFERENCES document(id),
				metadata JSONB,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				is_active BOOLEAN DEFAULT TRUE
			);
		`,
		"blockchain_record": `
			CREATE TABLE IF NOT EXISTS blockchain_record (
				id SERIAL PRIMARY KEY,
				related_table VARCHAR(100),
				related_id INTEGER,
				tx_id TEXT,
				metadata_hash TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				is_active BOOLEAN DEFAULT TRUE
			);
		`,
		"blockchain_nodes": `
			CREATE TABLE IF NOT EXISTS blockchain_nodes (
				id SERIAL PRIMARY KEY,
				node_name VARCHAR(255) NOT NULL,
				endpoint_url TEXT NOT NULL,
				node_type VARCHAR(100),
				network_id VARCHAR(100),
				provider VARCHAR(100),
				status VARCHAR(50),
				last_heartbeat TIMESTAMP,
				performance_metrics JSONB,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				is_active BOOLEAN DEFAULT TRUE
			);
		`,
		"shipment_transfer": `
			CREATE TABLE IF NOT EXISTS shipment_transfer (
				id SERIAL PRIMARY KEY,
				batch_id INTEGER REFERENCES batch(id),
				sender_id INTEGER REFERENCES account(id),
				receiver_id INTEGER REFERENCES account(id),
				transfer_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				status VARCHAR(50),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				is_active BOOLEAN DEFAULT TRUE
			);
		`,
		"transaction_nft": `
			CREATE TABLE IF NOT EXISTS transaction_nft (				
				id SERIAL PRIMARY KEY,
				token_id TEXT NOT NULL,
				batch_id INTEGER REFERENCES batch(id),
				shipment_transfer_id INTEGER REFERENCES shipment_transfer(id),
				owner_address VARCHAR(255) NOT NULL,
				status VARCHAR(50) NOT NULL,
				metadata JSONB,
				contract_address TEXT,
				is_active BOOLEAN DEFAULT true,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
		`,
		"transaction_nft_history": `
			CREATE TABLE IF NOT EXISTS transaction_nft_history (
				id SERIAL PRIMARY KEY,
				nft_id INTEGER REFERENCES transaction_nft(id),
				previous_status VARCHAR(50),
				new_status VARCHAR(50),
				previous_owner VARCHAR(255),
				new_owner VARCHAR(255),
				action_type VARCHAR(50),
				metadata_change JSONB,
				changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
		`,
		"company_compliance": `
			CREATE TABLE IF NOT EXISTS company_compliance (
				id SERIAL PRIMARY KEY,
				company_id INTEGER REFERENCES company(id),
				compliance_type VARCHAR(100) NOT NULL,
				status VARCHAR(50) NOT NULL,
				last_audit_date TIMESTAMP,
				next_audit_date TIMESTAMP,
				compliance_score FLOAT,
				findings TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				is_active BOOLEAN DEFAULT TRUE
			);
		`,
		"analytics_data": `
			CREATE TABLE IF NOT EXISTS analytics_data (
				id SERIAL PRIMARY KEY,
				metric_name VARCHAR(100) NOT NULL,
				metric_value FLOAT,
				metric_type VARCHAR(50),
				timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				entity_type VARCHAR(50),
				entity_id INTEGER,
				notes TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
		`,
		"identities": `
			CREATE TABLE IF NOT EXISTS identities (
				id SERIAL PRIMARY KEY,
				did VARCHAR(255) UNIQUE NOT NULL,
				entity_type VARCHAR(100) NOT NULL,
				entity_name VARCHAR(255) NOT NULL,
				public_key TEXT NOT NULL,
				metadata JSONB NOT NULL,
				status VARCHAR(50) NOT NULL,
				created_at TIMESTAMP NOT NULL,
				updated_at TIMESTAMP NOT NULL
			);
		`,
	}

	// Table creation order to satisfy foreign key constraints
	tableOrder := []string{
		"company",
		"account",
		"api_logs",
		"hatchery",
		"batch",
		"event",
		"environment_data",
		"document",
		"certificates",
		"blockchain_record",
		"blockchain_nodes",
		"shipment_transfer",
		"transaction_nft",
		"transaction_nft_history",
		"company_compliance",
		"analytics_data",
		"identities",
	}

	for _, tableName := range tableOrder {
		query := tableQueries[tableName]
		if _, err := DB.Exec(query); err != nil {
			return fmt.Errorf("failed to create table %s: %w", tableName, err)
		}
		fmt.Printf("Table %s created\n", tableName)
	}

	// Create triggers after all tables have been created
	if err := createTriggers(); err != nil {
		return fmt.Errorf("failed to create triggers: %w", err)
	}

	return nil
}

// createTriggers creates necessary database triggers
func createTriggers() error {
	// Check if triggers already exist to avoid unnecessary recreation
	var triggerExists bool
	err := DB.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'transaction_nft_history_trigger')").Scan(&triggerExists)
	if err != nil {
		// Just log the error but continue, as this is not fatal
		fmt.Printf("Warning: failed to check if triggers exist: %v\n", err)
	}
	
	if triggerExists {
		fmt.Println("Database triggers already exist, skipping creation")
		return nil
	}
	
	// Trigger function to track NFT status changes
	nftHistoryTriggerFn := `
	CREATE OR REPLACE FUNCTION track_transaction_nft_changes()
	RETURNS TRIGGER AS $$
	BEGIN
		IF (TG_OP = 'UPDATE') THEN
			-- Only insert history record if status, owner_address, or metadata has changed
			IF (OLD.status <> NEW.status OR OLD.owner_address <> NEW.owner_address OR OLD.metadata <> NEW.metadata) THEN
				INSERT INTO transaction_nft_history(
					nft_id, 
					previous_status, 
					new_status, 
					previous_owner, 
					new_owner, 
					action_type,
					metadata_change
				) VALUES (
					NEW.id,
					OLD.status,
					NEW.status,
					OLD.owner_address,
					NEW.owner_address,
					CASE 
						WHEN OLD.status <> NEW.status THEN 'status_change'
						WHEN OLD.owner_address <> NEW.owner_address THEN 'ownership_change'
						ELSE 'metadata_update'
					END,
					CASE 
						WHEN OLD.metadata <> NEW.metadata THEN 
							jsonb_build_object('old', OLD.metadata, 'new', NEW.metadata)
						ELSE NULL
					END
				);
			END IF;
		END IF;
		RETURN NULL;
	END;
	$$ LANGUAGE plpgsql;
	`
	
	// Trigger to track NFT changes
	nftHistoryTrigger := `
	DROP TRIGGER IF EXISTS transaction_nft_history_trigger ON transaction_nft;
	CREATE TRIGGER transaction_nft_history_trigger
	AFTER UPDATE ON transaction_nft
	FOR EACH ROW
	EXECUTE FUNCTION track_transaction_nft_changes();
	`
	
	// Soft delete trigger function
	softDeleteTriggerFn := `
	CREATE OR REPLACE FUNCTION handle_soft_delete()
	RETURNS TRIGGER AS $$
	BEGIN
		-- Update is_active to FALSE instead of actually deleting
		UPDATE transaction_nft SET is_active = FALSE, updated_at = NOW() WHERE id = OLD.id;
		RETURN NULL;
	END;
	$$ LANGUAGE plpgsql;
	`
	
	// Soft delete trigger
	softDeleteTrigger := `
	DROP TRIGGER IF EXISTS before_delete_transaction_nft ON transaction_nft;
	CREATE TRIGGER before_delete_transaction_nft
	BEFORE DELETE ON transaction_nft
	FOR EACH ROW
	EXECUTE FUNCTION handle_soft_delete();
	`
	
	// Execute trigger queries
	triggerQueries := []string{
		nftHistoryTriggerFn,
		nftHistoryTrigger,
		softDeleteTriggerFn,
		softDeleteTrigger,
	}
	
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for creating triggers: %w", err)
	}
	
	// Wrap in a transaction for better atomicity
	for _, query := range triggerQueries {
		_, err := tx.Exec(query)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create trigger: %w", err)
		}
	}
	
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for triggers: %w", err)
	}
	
	fmt.Println("Database triggers created successfully")
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

// OTPKey returns the Redis key for storing OTP for a given email
func OTPKey(email string) string {
	return "otp:reset:" + email
}

// Close closes the database connection
func Close() {
	dbInitMu.Lock()
	defer dbInitMu.Unlock()
	
	if DB != nil {
		if err := DB.Close(); err != nil {
			fmt.Printf("Error closing database connection: %v\n", err)
		} else {
			fmt.Println("Database connection closed successfully")
		}
		DB = nil
		dbInitialized = false
	}

	if Redis != nil {
		if err := Redis.Close(); err != nil {
			fmt.Printf("Error closing Redis connection: %v\n", err)
		} else {
			fmt.Println("Redis connection closed successfully")
		}
		Redis = nil
	}
}