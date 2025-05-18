package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"
	
	"github.com/golang-jwt/jwt/v4"
	"github.com/ipfs/go-ipfs-api"
	"os"
	"log"
)

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CompanyID int    `json:"company_id"`
	jwt.RegisteredClaims
}

// Company represents a company in the system
type Company struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Location    string    `json:"location"`
	ContactInfo string    `json:"contact_info"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsActive    bool      `json:"is_active"`

	// Relationships - use swaggertype to avoid recursion
	Users     []User     `json:"users,omitempty" gorm:"foreignKey:CompanyID" swaggertype:"array,object"`
	Hatcheries []Hatchery `json:"hatcheries,omitempty" gorm:"foreignKey:CompanyID" swaggertype:"array,object"`
}

// User represents a system user (user in DB)
type User struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"uniqueIndex"`
	FullName     string    `json:"full_name"`
	Phone        string    `json:"phone"`
	DateOfBirth  time.Time `json:"date_of_birth"`
	Email        string    `json:"email" gorm:"uniqueIndex"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CompanyID    int       `json:"company_id" gorm:"foreignKey:CompanyID"`
	Company      Company   `json:"company,omitempty" gorm:"foreignKey:CompanyID" swaggertype:"object"`
	LastLogin    time.Time `json:"last_login"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsActive     bool      `json:"is_active"`
}

// Define Account as an alias for User if Account is intended to represent a user
type Account = User

// Hatchery represents a shrimp hatchery
type Hatchery struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Contact   string    `json:"contact"`
	CompanyID int       `json:"company_id"`
	Company   Company   `json:"company,omitempty" gorm:"foreignKey:CompanyID" swaggertype:"object"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`

	// Relationships
	Batches []Batch `json:"batches,omitempty" gorm:"foreignKey:HatcheryID" swaggertype:"array,object"`
}

// Batch represents a batch of shrimp larvae
type Batch struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	HatcheryID int       `json:"hatchery_id"` // Foreign key to Hatchery
	Hatchery   Hatchery  `json:"hatchery,omitempty" gorm:"foreignKey:HatcheryID" swaggertype:"object"`
	Species    string    `json:"species"`
	Quantity   int       `json:"quantity"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	IsActive   bool      `json:"is_active"`

	// Relationships
	Events          []Event           `json:"events,omitempty" gorm:"foreignKey:BatchID" swaggertype:"array,object"`
	Documents       []Document        `json:"documents,omitempty" gorm:"foreignKey:BatchID" swaggertype:"array,object"`
	EnvironmentData []EnvironmentData `json:"environment_data,omitempty" gorm:"foreignKey:BatchID" swaggertype:"array,object"`
	BlockchainRecords []BlockchainRecord `json:"blockchain_records,omitempty" gorm:"polymorphic:Related;polymorphicValue:batch" swaggertype:"array,object"`
}

// Event represents a traceability event for a batch
type Event struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	BatchID   int       `json:"batch_id"` // Refers to Batch.ID
	EventType string    `json:"event_type"`
	ActorID   int       `json:"actor_id"` // Refers to User.ID
	Actor     User      `json:"actor,omitempty" gorm:"foreignKey:ActorID" swaggertype:"object"`
	Location  string    `json:"location"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  JSONB     `json:"metadata"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`

	// Related blockchain records
	BlockchainRecords []BlockchainRecord `json:"blockchain_records,omitempty" gorm:"polymorphic:Related;polymorphicValue:event" swaggertype:"array,object"`
}

// Document represents a document or certificate associated with a batch
type Document struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	BatchID    int       `json:"batch_id"` // Refers to Batch.ID
	DocType    string    `json:"doc_type"`
	IPFSHash   string    `json:"ipfs_hash"`
	IPFSURI    string    `json:"ipfs_uri"`
	FileName   string    `json:"file_name"`
	FileSize   int64     `json:"file_size"`
	UploadedBy int       `json:"uploaded_by"` // Refers to User.ID
	Uploader   User      `json:"uploader,omitempty" gorm:"foreignKey:UploadedBy" swaggertype:"object"`
	UploadedAt time.Time `json:"uploaded_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	IsActive   bool      `json:"is_active"`
	Company      Company   `json:"company,omitempty" gorm:"foreignKey:CompanyID" swaggertype:"object"`

	// Related blockchain records
	BlockchainRecords []BlockchainRecord `json:"blockchain_records,omitempty" gorm:"polymorphic:Related;polymorphicValue:document" swaggertype:"array,object"`
}

// EnvironmentData represents environmental parameters for a batch (environment in DB)
type EnvironmentData struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	BatchID     int       `json:"batch_id"` // Refers to Batch.ID
	Temperature float64   `json:"temperature"`
	PH          float64   `json:"ph"`
	Salinity    float64   `json:"salinity"`
	Density     float64   `json:"density"`
	Age         int       `json:"age"`
	Timestamp   time.Time `json:"timestamp"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsActive    bool      `json:"is_active"`

	// Related blockchain records
	BlockchainRecords []BlockchainRecord `json:"blockchain_records,omitempty" gorm:"polymorphic:Related;polymorphicValue:environment" swaggertype:"array,object"`
}

// BlockchainRecord represents a blockchain transaction record
type BlockchainRecord struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	RelatedTable string    `json:"related_table"`
	RelatedID    int       `json:"related_id"`
	TxID         string    `json:"tx_id"`
	MetadataHash string    `json:"metadata_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsActive     bool      `json:"is_active"`
}

// JSONB is a wrapper around json.RawMessage to implement SQL scanner interface
type JSONB []byte

// Value returns JSONB value for saving to the database
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

// Scan scans a value from the database into JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("invalid scan source for JSONB")
	}
	
	*j = bytes
	return nil
}

// MarshalJSON returns the JSON encoding of JSONB
func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON sets *j to a copy of data
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[:0], data...)
	return nil
}

// SwaggerUIJsonRawMessage is for documentation purposes only
// to fix the issue with Swagger not recognizing json.RawMessage
type SwaggerUIJsonRawMessage struct {
	// Can be any valid JSON value
	RawMessage map[string]interface{} `json:"rawMessage,omitempty" example:"{\"key\":\"value\"}"`
}

// BatchWithHatchery represents a batch with its associated hatchery information
type BatchWithHatchery struct {
	Batch
	HatcheryName     string `json:"hatchery_name"`
	HatcheryLocation string `json:"hatchery_location"`
	HatcheryContact  string `json:"hatchery_contact"`
}

// EventWithActor represents an event with its associated actor information
type EventWithActor struct {
	Event
	ActorName  string `json:"actor_name"`
	ActorRole  string `json:"actor_role"`
	ActorEmail string `json:"actor_email"`
}

// LogisticsEvent represents a logistics event in the supply chain
type LogisticsEvent struct {
	ID              int       `json:"id"`
	BatchID         int       `json:"batch_id"`
	EventType       string    `json:"event_type"`
	FromLocation    string    `json:"from_location"`
	ToLocation      string    `json:"to_location"`
	TransporterName string    `json:"transporter_name"`
	DepartureTime   time.Time `json:"departure_time"`
	ArrivalTime     time.Time `json:"arrival_time"`
	Status          string    `json:"status"`
	Metadata        JSONB     `json:"metadata"`
	Timestamp       time.Time `json:"timestamp"`
}

// ShipmentTransfer represents a transfer of a batch between supply chain participants
type ShipmentTransfer struct {
	ID                int       `json:"id" gorm:"primaryKey"` // Transfer ID as primary key
	BatchID           int       `json:"batch_id"`             // Reference to the batch being transferred
	SourceID          string    `json:"source_id"`            // ID of the source entity
	SourceType        string    `json:"source_type"`          // Type of the source entity
	DestinationID     string    `json:"destination_id"`       // ID of the destination entity
	DestinationType   string    `json:"destination_type"`     // Type of the destination entity
	Quantity          int       `json:"quantity"`             // Quantity being transferred
	TransferredAt     time.Time `json:"transferred_at"`       // Time of transfer
	TransferredBy     string    `json:"transferred_by"`       // User who initiated the transfer
	Status            string    `json:"status"`               // Status of transfer
	BlockchainTxID    string    `json:"blockchain_tx_id"`     // Blockchain transaction ID
	NFTTokenID        int       `json:"nft_token_id"`         // NFT token ID if tokenized
	NFTContractAddress string    `json:"nft_contract_address"` // NFT contract address
	TransferNotes     string    `json:"transfer_notes"`       // Notes about the transfer
	Metadata          JSONB     `json:"metadata"`             // Additional metadata
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	IsActive          bool      `json:"is_active"`
}

// SaveDocumentToIPFS uploads a document to IPFS and returns the CID and URI
func SaveDocumentToIPFS(filePath string) (string, string, error) {
	// Connect to IPFS node
	ipfsNodeURL := os.Getenv("IPFS_NODE_URL")
	if ipfsNodeURL == "" {
		ipfsNodeURL = "http://localhost:5001" // Default to local IPFS node
	}
	sh := shell.NewShell(ipfsNodeURL)

	// Open the file
	file, err := os.Open(filePath)
	if (err != nil) {
		log.Printf("Failed to open file: %v", err)
		return "", "", err
	}
	defer file.Close()

	// Add the file to IPFS
	cid, err := sh.Add(file)
	if err != nil {
		log.Printf("Failed to upload file to IPFS: %v", err)
		return "", "", err
	}

	// Construct the IPFS URI
	gatewayURL := os.Getenv("IPFS_GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8080" // Default to local gateway
	}
	
	// Remove trailing slash if present
	gatewayURL = strings.TrimSuffix(gatewayURL, "/")
	
	// If the gateway URL already ends with /ipfs, don't add it again
	ipfsURI := ""
	if strings.HasSuffix(gatewayURL, "/ipfs") {
		ipfsURI = fmt.Sprintf("%s/%s", gatewayURL, cid)
	} else {
		ipfsURI = fmt.Sprintf("%s/ipfs/%s", gatewayURL, cid)
	}

	return cid, ipfsURI, nil
}
