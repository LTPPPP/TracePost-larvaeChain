package models

import (
	"database/sql/driver"
	"errors"
	"time"
	
	"github.com/golang-jwt/jwt/v4"
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

	// Relationships
	Users     []User     `json:"users,omitempty" gorm:"foreignKey:CompanyID"`
	Hatcheries []Hatchery `json:"hatcheries,omitempty" gorm:"foreignKey:CompanyID"`
}

// User represents a system user (account in DB)
type User struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"uniqueIndex"`
	Email        string    `json:"email" gorm:"uniqueIndex"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CompanyID    int       `json:"company_id"`
	Company      Company   `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	LastLogin    time.Time `json:"last_login"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsActive     bool      `json:"is_active"`
}

// Hatchery represents a shrimp hatchery
type Hatchery struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Contact   string    `json:"contact"`
	CompanyID int       `json:"company_id"`
	Company   Company   `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`

	// Relationships
	Batches []Batch `json:"batches,omitempty" gorm:"foreignKey:HatcheryID"`
}

// Batch represents a batch of shrimp larvae
type Batch struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	HatcheryID int       `json:"hatchery_id"` // Foreign key to Hatchery
	Hatchery   Hatchery  `json:"hatchery,omitempty" gorm:"foreignKey:HatcheryID"`
	Species    string    `json:"species"`
	Quantity   int       `json:"quantity"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	IsActive   bool      `json:"is_active"`

	// Relationships
	Events          []Event           `json:"events,omitempty" gorm:"foreignKey:BatchID"`
	Documents       []Document        `json:"documents,omitempty" gorm:"foreignKey:BatchID"`
	EnvironmentData []EnvironmentData `json:"environment_data,omitempty" gorm:"foreignKey:BatchID"`
	BlockchainRecords []BlockchainRecord `json:"blockchain_records,omitempty" gorm:"polymorphic:Related;polymorphicValue:batch"`
}

// Event represents a traceability event for a batch
type Event struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	BatchID   int       `json:"batch_id"` // Refers to Batch.ID
	EventType string    `json:"event_type"`
	ActorID   int       `json:"actor_id"` // Refers to User.ID
	Actor     User      `json:"actor,omitempty" gorm:"foreignKey:ActorID"`
	Location  string    `json:"location"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  JSONB     `json:"metadata"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`

	// Related blockchain records
	BlockchainRecords []BlockchainRecord `json:"blockchain_records,omitempty" gorm:"polymorphic:Related;polymorphicValue:event"`
}

// Document represents a document or certificate associated with a batch
type Document struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	BatchID    int       `json:"batch_id"` // Refers to Batch.ID
	DocType    string    `json:"doc_type"`
	IPFSHash   string    `json:"ipfs_hash"`
	UploadedBy int       `json:"uploaded_by"` // Refers to User.ID
	Uploader   User      `json:"uploader,omitempty" gorm:"foreignKey:UploadedBy"`
	UploadedAt time.Time `json:"uploaded_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	IsActive   bool      `json:"is_active"`

	// Related blockchain records
	BlockchainRecords []BlockchainRecord `json:"blockchain_records,omitempty" gorm:"polymorphic:Related;polymorphicValue:document"`
}

// EnvironmentData represents environmental parameters for a batch (environment in DB)
type EnvironmentData struct {
	ID              int       `json:"id" gorm:"primaryKey"`
	BatchID         int       `json:"batch_id"` // Refers to Batch.ID
	Temperature     float64   `json:"temperature"`
	PH              float64   `json:"ph"`
	Salinity        float64   `json:"salinity"`
	DissolvedOxygen float64   `json:"dissolved_oxygen"`
	Timestamp       time.Time `json:"timestamp"`
	UpdatedAt       time.Time `json:"updated_at"`
	IsActive        bool      `json:"is_active"`

	// Related blockchain records
	BlockchainRecords []BlockchainRecord `json:"blockchain_records,omitempty" gorm:"polymorphic:Related;polymorphicValue:environment"`
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
	Data interface{} `json:"data"`
}
