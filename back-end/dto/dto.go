package dto

import (
	"time"
	"encoding/json"
	"github.com/LTPPPP/TracePost-larvaeChain/models"
)

// DTO models for Swagger to prevent recursion issues

type CompanyDTO struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Location    string    `json:"location"`
	ContactInfo string    `json:"contact_info"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsActive    bool      `json:"is_active"`
	UserIDs      []int `json:"user_ids,omitempty"`
	HatcheryIDs  []int `json:"hatchery_ids,omitempty"`
}

type UserDTO struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	FullName     string    `json:"full_name"`
	Phone        string    `json:"phone"`
	DateOfBirth  time.Time `json:"date_of_birth"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	CompanyID    int       `json:"company_id"`
	CompanyName  string    `json:"company_name,omitempty"`
	LastLogin    time.Time `json:"last_login"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsActive     bool      `json:"is_active"`
}

type HatcheryDTO struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Contact   string    `json:"contact"`
	CompanyID int       `json:"company_id"`
	CompanyName string   `json:"company_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
	BatchIDs    []int    `json:"batch_ids,omitempty"`
}

type BatchDTO struct {
	ID         int       `json:"id"`
	HatcheryID int       `json:"hatchery_id"`
	HatcheryName string   `json:"hatchery_name,omitempty"`
	Species    string    `json:"species"`
	Quantity   int       `json:"quantity"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	IsActive   bool      `json:"is_active"`
	EventCount          int    `json:"event_count,omitempty"`
	DocumentCount       int    `json:"document_count,omitempty"`
	EnvironmentDataCount int   `json:"environment_data_count,omitempty"`
	BlockchainRecordCount int  `json:"blockchain_record_count,omitempty"`
}

type EventDTO struct {
	ID        int                    `json:"id"`
	BatchID   int                    `json:"batch_id"`
	EventType string                 `json:"event_type"`
	ActorID   int                    `json:"actor_id"`
	ActorName string                 `json:"actor_name,omitempty"`
	Location  string                 `json:"location"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata" swaggertype:"object"`
	UpdatedAt time.Time              `json:"updated_at"`
	IsActive  bool                   `json:"is_active"`
}

type DocumentDTO struct {
	ID         int       `json:"id"`
	BatchID    int       `json:"batch_id"`
	DocType    string    `json:"doc_type"`
	IPFSHash   string    `json:"ipfs_hash"`
	IPFSURI    string    `json:"ipfs_uri"`
	FileName   string    `json:"file_name"`
	FileSize   int64     `json:"file_size"`
	UploadedBy int       `json:"uploaded_by"`
	UploaderName string  `json:"uploader_name,omitempty"`
	CompanyName string   `json:"company_name,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	IsActive   bool      `json:"is_active"`
}

type EnvironmentDataDTO struct {
	ID          int       `json:"id"`
	BatchID     int       `json:"batch_id"`
	Temperature float64   `json:"temperature"`
	PH          float64   `json:"ph"`
	Salinity    float64   `json:"salinity"`
	Density     float64   `json:"density"`
	Age         int       `json:"age"`
	Timestamp   time.Time `json:"timestamp"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsActive    bool      `json:"is_active"`
}

// Convert full models to DTOs

func ToCompanyDTO(company models.Company) CompanyDTO {
	dto := CompanyDTO{
		ID:          company.ID,
		Name:        company.Name,
		Type:        company.Type,
		Location:    company.Location,
		ContactInfo: company.ContactInfo,
		CreatedAt:   company.CreatedAt,
		UpdatedAt:   company.UpdatedAt,
		IsActive:    company.IsActive,
	}
	
	// Convert relationships to IDs
	if len(company.Users) > 0 {
		dto.UserIDs = make([]int, len(company.Users))
		for i, user := range company.Users {
			dto.UserIDs[i] = user.ID
		}
	}
	
	if len(company.Hatcheries) > 0 {
		dto.HatcheryIDs = make([]int, len(company.Hatcheries))
		for i, hatchery := range company.Hatcheries {
			dto.HatcheryIDs[i] = hatchery.ID
		}
	}
	
	return dto
}

func ToHatcheryDTO(hatchery models.Hatchery) HatcheryDTO {
	dto := HatcheryDTO{
		ID:        hatchery.ID,
		Name:      hatchery.Name,
		CompanyID: hatchery.CompanyID,
		CreatedAt: hatchery.CreatedAt,
		UpdatedAt: hatchery.UpdatedAt,
		IsActive:  hatchery.IsActive,
	}
	
	// Add company name if available
	if hatchery.Company.ID != 0 {
		dto.CompanyName = hatchery.Company.Name
	}
	
	// Convert batch relationships to IDs
	if len(hatchery.Batches) > 0 {
		dto.BatchIDs = make([]int, len(hatchery.Batches))
		for i, batch := range hatchery.Batches {
			dto.BatchIDs[i] = batch.ID
		}
	}
	
	return dto
}

func ToBatchDTO(batch models.Batch) BatchDTO {
	dto := BatchDTO{
		ID:         batch.ID,
		HatcheryID: batch.HatcheryID,
		Species:    batch.Species,
		Quantity:   batch.Quantity,
		Status:     batch.Status,
		CreatedAt:  batch.CreatedAt,
		UpdatedAt:  batch.UpdatedAt,
		IsActive:   batch.IsActive,
	}
	
	// Add hatchery name if available
	if batch.Hatchery.ID != 0 {
		dto.HatcheryName = batch.Hatchery.Name
	}
	
	// Add counts instead of full objects
	dto.EventCount = len(batch.Events)
	dto.DocumentCount = len(batch.Documents)
	dto.EnvironmentDataCount = len(batch.EnvironmentData)
	dto.BlockchainRecordCount = len(batch.BlockchainRecords)
	
	return dto
}

// JSONBToSwaggerFormat converts JSONB to a map for Swagger documentation
func JSONBToSwaggerFormat(data models.JSONB) map[string]interface{} {
	var result map[string]interface{}
	if len(data) == 0 {
		return nil
	}
	
	// Try to unmarshal the JSONB data into a map
	err := json.Unmarshal(data, &result)
	if err != nil {
		// If there's an error, return a placeholder
		return map[string]interface{}{"error": "Invalid JSON format"}
	}
	
	return result
}

// ToEventDTO converts an Event model to EventDTO
func ToEventDTO(event models.Event) EventDTO {
	dto := EventDTO{
		ID:        event.ID,
		BatchID:   event.BatchID,
		EventType: event.EventType,
		ActorID:   event.ActorID,
		Location:  event.Location,
		Timestamp: event.Timestamp,
		UpdatedAt: event.UpdatedAt,
		IsActive:  event.IsActive,
		Metadata:  JSONBToSwaggerFormat(event.Metadata),
	}
	
	// Add actor name if available
	if event.Actor.ID != 0 {
		dto.ActorName = event.Actor.FullName
	}
	
	return dto
}

// ToDocumentDTO converts a Document model to DocumentDTO
func ToDocumentDTO(doc models.Document) DocumentDTO {
	dto := DocumentDTO{
		ID:         doc.ID,
		BatchID:    doc.BatchID,
		DocType:    doc.DocType,
		IPFSHash:   doc.IPFSHash,
		IPFSURI:    doc.IPFSURI,
		FileName:   doc.FileName,
		FileSize:   doc.FileSize,
		UploadedBy: doc.UploadedBy,
		UploadedAt: doc.UploadedAt,
		UpdatedAt:  doc.UpdatedAt,
		IsActive:   doc.IsActive,
	}
	
	// Add uploader name if available
	if doc.Uploader.ID != 0 {
		dto.UploaderName = doc.Uploader.FullName
	}
	
	// Add company name if available
	if doc.Company.ID != 0 {
		dto.CompanyName = doc.Company.Name
	}
	
	return dto
}

// ToEnvironmentDataDTO converts an EnvironmentData model to EnvironmentDataDTO
func ToEnvironmentDataDTO(data models.EnvironmentData) EnvironmentDataDTO {
	return EnvironmentDataDTO{
		ID:          data.ID,
		BatchID:     data.BatchID,
		Temperature: data.Temperature,
		PH:          data.PH,
		Salinity:    data.Salinity,
		Density:     data.Density,
		Age:         data.Age,
		Timestamp:   data.Timestamp,
		UpdatedAt:   data.UpdatedAt,
		IsActive:    data.IsActive,
	}
}
