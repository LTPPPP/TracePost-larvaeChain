package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Batch represents a batch of shrimp larvae
type Batch struct {
	ID             int       `json:"id"`
	BatchID        string    `json:"batch_id"`
	HatcheryID     string    `json:"hatchery_id"`
	CreationDate   time.Time `json:"creation_date"`
	Species        string    `json:"species"`
	Quantity       int       `json:"quantity"`
	Status         string    `json:"status"`
	BlockchainTxID string    `json:"blockchain_tx_id"`
	MetadataHash   string    `json:"metadata_hash"`
}

// Event represents a traceability event for a batch
type Event struct {
	ID             int       `json:"id"`
	BatchID        string    `json:"batch_id"`
	EventType      string    `json:"event_type"`
	Timestamp      time.Time `json:"timestamp"`
	Location       string    `json:"location"`
	ActorID        string    `json:"actor_id"`
	Details        JSONB     `json:"details"`
	BlockchainTxID string    `json:"blockchain_tx_id"`
	MetadataHash   string    `json:"metadata_hash"`
}

// Document represents a document or certificate associated with a batch
type Document struct {
	ID             int       `json:"id"`
	BatchID        string    `json:"batch_id"`
	DocumentType   string    `json:"document_type"`
	IPFSHash       string    `json:"ipfs_hash"`
	UploadDate     time.Time `json:"upload_date"`
	Issuer         string    `json:"issuer"`
	IsVerified     bool      `json:"is_verified"`
	BlockchainTxID string    `json:"blockchain_tx_id"`
}

// EnvironmentData represents environmental parameters for a batch
type EnvironmentData struct {
	ID               int       `json:"id"`
	BatchID          string    `json:"batch_id"`
	Timestamp        time.Time `json:"timestamp"`
	Temperature      float64   `json:"temperature"`
	PH               float64   `json:"ph"`
	Salinity         float64   `json:"salinity"`
	DissolvedOxygen  float64   `json:"dissolved_oxygen"`
	OtherParams      JSONB     `json:"other_params"`
	BlockchainTxID   string    `json:"blockchain_tx_id"`
}

// User represents a system user
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Don't expose in JSON
	Role         string    `json:"role"`
	CompanyID    string    `json:"company_id"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	LastLogin    time.Time `json:"last_login"`
}

// JSONB is a wrapper around json.RawMessage to implement SQL scanner interface
type JSONB json.RawMessage

// Value returns JSONB value
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// Scan scans value into JSONB
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