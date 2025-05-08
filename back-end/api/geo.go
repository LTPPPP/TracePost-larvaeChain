package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/config"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"time"
)

// GeoLocation represents a geographic location
type GeoLocation struct {
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Altitude   float64   `json:"altitude,omitempty"`
	Accuracy   float64   `json:"accuracy,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	LocationType string   `json:"location_type"` // "hatchery", "farm", "processing", "storage", "transportation", "export", etc.
	Description string    `json:"description,omitempty"`
}

// RecordGeoLocationRequest represents a request to record a geographic location
type RecordGeoLocationRequest struct {
	BatchID     string    `json:"batch_id"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Altitude    float64   `json:"altitude,omitempty"`
	Accuracy    float64   `json:"accuracy,omitempty"`
	LocationType string    `json:"location_type"`
	Description string     `json:"description,omitempty"`
}

// BatchJourneyResponse represents a batch's geographic journey
type BatchJourneyResponse struct {
	BatchID     string        `json:"batch_id"`
	Locations   []GeoLocation `json:"locations"`
	TotalDistance float64      `json:"total_distance"`
	StartTime   string        `json:"start_time"`
	CurrentTime string        `json:"current_time"`
	TransitTime string        `json:"transit_time"`
}

// CurrentLocationResponse represents a batch's current location
type CurrentLocationResponse struct {
	BatchID     string      `json:"batch_id"`
	Location    GeoLocation `json:"location"`
	LastUpdated string      `json:"last_updated"`
	Status      string      `json:"status"`
}

// RecordGeoLocation records a geographic location for a batch
// @Summary Record geographic location
// @Description Record a geographic location for a batch
// @Tags geo
// @Accept json
// @Produce json
// @Param request body RecordGeoLocationRequest true "Location details"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /geo/location [post]
func RecordGeoLocation(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Parse request
	var req RecordGeoLocationRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.BatchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	if req.Latitude < -90 || req.Latitude > 90 {
		return fiber.NewError(fiber.StatusBadRequest, "Latitude must be between -90 and 90")
	}
	
	if req.Longitude < -180 || req.Longitude > 180 {
		return fiber.NewError(fiber.StatusBadRequest, "Longitude must be between -180 and 180")
	}
	
	if req.LocationType == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Location type is required")
	}
	
	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1)", req.BatchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}
	
	// Record location in database
	now := time.Now()
	_, err = db.DB.Exec(`
		INSERT INTO geo_locations (batch_id, latitude, longitude, altitude, accuracy, location_type, description, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		req.BatchID,
		req.Latitude,
		req.Longitude,
		req.Altitude,
		req.Accuracy,
		req.LocationType,
		req.Description,
		now,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record location: "+err.Error())
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Record on blockchain
	_, err = blockchainClient.SubmitTransaction("GEO_LOCATION", map[string]interface{}{
		"batch_id":       req.BatchID,
		"latitude":       req.Latitude,
		"longitude":      req.Longitude,
		"altitude":       req.Altitude,
		"accuracy":       req.Accuracy,
		"location_type":  req.LocationType,
		"description":    req.Description,
		"timestamp":      now,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record location on blockchain: "+err.Error())
	}
	
	// Return response
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Geolocation recorded successfully",
		Data: GeoLocation{
			Latitude:      req.Latitude,
			Longitude:     req.Longitude,
			Altitude:      req.Altitude,
			Accuracy:      req.Accuracy,
			Timestamp:     now,
			LocationType:  req.LocationType,
			Description:   req.Description,
		},
	})
}

// GetBatchJourney gets the geographic journey of a batch
// @Summary Get batch journey
// @Description Get the geographic journey of a batch
// @Tags geo
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=BatchJourneyResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /geo/batch/{batchId}/journey [get]
func GetBatchJourney(c *fiber.Ctx) error {
	// Get batch ID from path
	batchID := c.Params("batchId")
	if batchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}
	
	// Get locations from database
	rows, err := db.DB.Query(`
		SELECT latitude, longitude, altitude, accuracy, location_type, description, recorded_at
		FROM geo_locations
		WHERE batch_id = $1
		ORDER BY recorded_at ASC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()
	
	// Process locations
	var locations []GeoLocation
	var startTime, endTime time.Time
	var totalDistance float64 = 0
	var prevLat, prevLng float64
	var firstLocation = true
	
	for rows.Next() {
		var location GeoLocation
		var recordedAt time.Time
		
		err := rows.Scan(
			&location.Latitude,
			&location.Longitude,
			&location.Altitude,
			&location.Accuracy,
			&location.LocationType,
			&location.Description,
			&recordedAt,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error parsing location data")
		}
		
		location.Timestamp = recordedAt
		
		// Update start and end times
		if firstLocation {
			startTime = recordedAt
			firstLocation = false
		}
		endTime = recordedAt
		
		// Calculate distance from previous location
		if len(locations) > 0 {
			distance := calculateDistance(prevLat, prevLng, location.Latitude, location.Longitude)
			totalDistance += distance
		}
		
		prevLat = location.Latitude
		prevLng = location.Longitude
		
		locations = append(locations, location)
	}
	
	// Calculate transit time
	transitTime := endTime.Sub(startTime)
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch journey retrieved successfully",
		Data: BatchJourneyResponse{
			BatchID:       batchID,
			Locations:     locations,
			TotalDistance: totalDistance,
			StartTime:     startTime.Format(time.RFC3339),
			CurrentTime:   endTime.Format(time.RFC3339),
			TransitTime:   formatDuration(transitTime),
		},
	})
}

// GetBatchCurrentLocation gets the current location of a batch
// @Summary Get batch current location
// @Description Get the current location of a batch
// @Tags geo
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=CurrentLocationResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /geo/batch/{batchId}/current-location [get]
func GetBatchCurrentLocation(c *fiber.Ctx) error {
	// Get batch ID from path
	batchID := c.Params("batchId")
	if batchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	// Check if batch exists and get status
	var exists bool
	var status string
	err := db.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1), 
		       (SELECT status FROM batches WHERE batch_id = $1)
	`, batchID).Scan(&exists, &status)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}
	
	// Get most recent location from database
	var location GeoLocation
	var recordedAt time.Time
	
	err = db.DB.QueryRow(`
		SELECT latitude, longitude, altitude, accuracy, location_type, description, recorded_at
		FROM geo_locations
		WHERE batch_id = $1
		ORDER BY recorded_at DESC
		LIMIT 1
	`, batchID).Scan(
		&location.Latitude,
		&location.Longitude,
		&location.Altitude,
		&location.Accuracy,
		&location.LocationType,
		&location.Description,
		&recordedAt,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "No location data found for batch")
	}
	
	location.Timestamp = recordedAt
	
	// Return response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Current location retrieved successfully",
		Data: CurrentLocationResponse{
			BatchID:     batchID,
			Location:    location,
			LastUpdated: recordedAt.Format(time.RFC3339),
			Status:      status,
		},
	})
}

// Helper function to calculate distance between two coordinates using Haversine formula
func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	// Haversine formula implementation
	// For simplicity, we'll return a mock distance value
	// In a real implementation, this would calculate the actual distance
	return 10.5 // Example distance in kilometers
}

// Helper function to format duration
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	
	result := ""
	if days > 0 {
		result += string(days) + " days, "
	}
	if hours > 0 {
		result += string(hours) + " hours, "
	}
	result += string(minutes) + " minutes"
	
	return result
}