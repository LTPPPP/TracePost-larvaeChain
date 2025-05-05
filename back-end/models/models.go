package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
	
	"github.com/golang-jwt/jwt/v4"
)

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CompanyID string `json:"company_id"`
	jwt.RegisteredClaims
}

// Hatcheries represents a shrimp hatchery
type Hatcheries struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Contact   string    `json:"contact"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Batches []Batch `json:"batches" gorm:"foreignKey:HatcheryID"`
}

// Batch represents a batch of shrimp larvae
type Batch struct {
	ID             int       `json:"id" gorm:"primaryKey"`
	BatchID        string    `json:"batch_id" gorm:"uniqueIndex"`
	HatcheryID     int       `json:"hatchery_id"` // Foreign key to Hatcheries
	Hatchery       Hatcheries `json:"hatchery" gorm:"foreignKey:HatcheryID"`
	CreationDate   time.Time `json:"creation_date"`
	Species        string    `json:"species"`
	Quantity       int       `json:"quantity"`
	Status         string    `json:"status"`
	BlockchainTxID string    `json:"blockchain_tx_id"`
	MetadataHash   string    `json:"metadata_hash"`

	// Relationships
	Events          []Event           `json:"events" gorm:"foreignKey:BatchID;references:BatchID"`
	Documents       []Document        `json:"documents" gorm:"foreignKey:BatchID;references:BatchID"`
	EnvironmentData []EnvironmentData `json:"environment_data" gorm:"foreignKey:BatchID;references:BatchID"`
}

// Event represents a traceability event for a batch
type Event struct {
	ID             int       `json:"id" gorm:"primaryKey"`
	BatchID        string    `json:"batch_id"` // Refers to Batch.BatchID
	EventType      string    `json:"event_type"`
	Timestamp      time.Time `json:"timestamp"`
	Location       string    `json:"location"`
	ActorID        int       `json:"actor_id"` // Refers to User.ID
	Actor          User      `json:"actor" gorm:"foreignKey:ActorID"`
	Details        JSONB     `json:"details"`
	BlockchainTxID string    `json:"blockchain_tx_id"`
	MetadataHash   string    `json:"metadata_hash"`
}

// Document represents a document or certificate associated with a batch
type Document struct {
	ID             int       `json:"id" gorm:"primaryKey"`
	BatchID        string    `json:"batch_id"` // Refers to Batch.BatchID
	DocumentType   string    `json:"document_type"`
	IPFSHash       string    `json:"ipfs_hash"`
	UploadDate     time.Time `json:"upload_date"`
	Issuer         string    `json:"issuer"`
	IsVerified     bool      `json:"is_verified"`
	BlockchainTxID string    `json:"blockchain_tx_id"`
}

// EnvironmentData represents environmental parameters for a batch
type EnvironmentData struct {
	ID              int       `json:"id" gorm:"primaryKey"`
	BatchID         string    `json:"batch_id"` // Refers to Batch.BatchID
	Timestamp       time.Time `json:"timestamp"`
	Temperature     float64   `json:"temperature"`
	PH              float64   `json:"ph"`
	Salinity        float64   `json:"salinity"`
	DissolvedOxygen float64   `json:"dissolved_oxygen"`
	OtherParams     JSONB     `json:"other_params"`
	BlockchainTxID  string    `json:"blockchain_tx_id"`
}

// User represents a system user
type User struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"unique"`
	PasswordHash string    `json:"-"` // Don't expose in JSON
	Role         string    `json:"role"`
	CompanyID    string    `json:"company_id"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	LastLogin    time.Time `json:"last_login"`

	// Relationships
	Events []Event `json:"events" gorm:"foreignKey:ActorID"`
}

// JSONB is a wrapper around json.RawMessage to implement SQL scanner interface
type JSONB json.RawMessage

// Value returns JSONB value for saving to the database
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// Scan scans a value from the database into JSONB
func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	if len(bytes) == 0 {
		*j = JSONB("null")
		return nil
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = JSONB(result)
	return err
}

// MarshalJSON returns the JSON encoding of JSONB
func (j JSONB) MarshalJSON() ([]byte, error) {
	return json.RawMessage(j).MarshalJSON()
}

// UnmarshalJSON sets *j to a copy of data
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("JSONB: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}
