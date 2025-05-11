// anomaly_detection.go
package blockchain

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"
)

// AnomalyDetectionService provides methods for detecting anomalies in supply chain data
type AnomalyDetectionService struct {
	BlockchainClient *BlockchainClient
}

// AnomalyType represents the type of anomaly detected
type AnomalyType string

const (
	// AnomalyTypeTimeGap represents an unusual time gap between events
	AnomalyTypeTimeGap AnomalyType = "TIME_GAP"
	
	// AnomalyTypeLocation represents an unusual location change
	AnomalyTypeLocation AnomalyType = "LOCATION"
	
	// AnomalyTypeTemperature represents unusual temperature readings
	AnomalyTypeTemperature AnomalyType = "TEMPERATURE"
	
	// AnomalyTypeHumidity represents unusual humidity readings
	AnomalyTypeHumidity AnomalyType = "HUMIDITY"
	
	// AnomalyTypeAuthorization represents an authorization issue
	AnomalyTypeAuthorization AnomalyType = "AUTHORIZATION"
	
	// AnomalyTypeRouteDeviation represents a deviation from expected route
	AnomalyTypeRouteDeviation AnomalyType = "ROUTE_DEVIATION"
)

// AnomalyDetectionResult represents the result of an anomaly detection
type AnomalyDetectionResult struct {
	AnomalyType     AnomalyType              `json:"anomalyType"`
	BatchID         string                   `json:"batchId"`
	TransferID      string                   `json:"transferId,omitempty"`
	Timestamp       time.Time                `json:"timestamp"`
	Confidence      float64                  `json:"confidence"`
	ExpectedValue   interface{}              `json:"expectedValue,omitempty"`
	ActualValue     interface{}              `json:"actualValue,omitempty"`
	Description     string                   `json:"description"`
	RelatedEvents   []string                 `json:"relatedEvents,omitempty"`
	Recommendations []string                 `json:"recommendations,omitempty"`
	Metadata        map[string]interface{}   `json:"metadata,omitempty"`
}

// BatchEventData represents event data for a batch
type BatchEventData struct {
	BatchID         string                   `json:"batchId"`
	Events          []map[string]interface{} `json:"events"`
	Transfers       []map[string]interface{} `json:"transfers"`
	EnvironmentData []map[string]interface{} `json:"environmentData"`
}

// NewAnomalyDetectionService creates a new anomaly detection service
func NewAnomalyDetectionService(client *BlockchainClient) *AnomalyDetectionService {
	return &AnomalyDetectionService{
		BlockchainClient: client,
	}
}

// DetectAnomaliesForBatch detects anomalies for a batch
func (s *AnomalyDetectionService) DetectAnomaliesForBatch(batchID string) ([]AnomalyDetectionResult, error) {
	// Get batch data from blockchain
	batchData, err := s.getBatchData(batchID)
	if err != nil {
		return nil, err
	}
	
	var anomalies []AnomalyDetectionResult
	
	// Detect time gap anomalies
	timeGapAnomalies, err := s.detectTimeGapAnomalies(batchData)
	if err != nil {
		return nil, err
	}
	anomalies = append(anomalies, timeGapAnomalies...)
	
	// Detect temperature anomalies
	tempAnomalies, err := s.detectTemperatureAnomalies(batchData)
	if err != nil {
		return nil, err
	}
	anomalies = append(anomalies, tempAnomalies...)
	
	// Detect humidity anomalies
	humidityAnomalies, err := s.detectHumidityAnomalies(batchData)
	if err != nil {
		return nil, err
	}
	anomalies = append(anomalies, humidityAnomalies...)
	
	// Detect location anomalies
	locationAnomalies, err := s.detectLocationAnomalies(batchData)
	if err != nil {
		return nil, err
	}
	anomalies = append(anomalies, locationAnomalies...)
	
	// Detect authorization anomalies
	authAnomalies, err := s.detectAuthorizationAnomalies(batchData)
	if err != nil {
		return nil, err
	}
	anomalies = append(anomalies, authAnomalies...)
	
	// Sort anomalies by timestamp
	sort.Slice(anomalies, func(i, j int) bool {
		return anomalies[i].Timestamp.Before(anomalies[j].Timestamp)
	})
	
	return anomalies, nil
}

// getBatchData retrieves batch data from the blockchain
func (s *AnomalyDetectionService) getBatchData(batchID string) (*BatchEventData, error) {
	// Call smart contract to get batch events
	eventsResult, err := s.BlockchainClient.CallContract(
		"", // Contract address should be configured
		"getBatchEvents(string)",
		[]interface{}{batchID},
	)
	if err != nil {
		return nil, err
	}
	
	// Call smart contract to get batch transfers
	transfersResult, err := s.BlockchainClient.CallContract(
		"", // Contract address should be configured
		"getBatchTransfers(string)",
		[]interface{}{batchID},
	)
	if err != nil {
		return nil, err
	}
	
	// Call smart contract to get environment data
	envDataResult, err := s.BlockchainClient.CallContract(
		"", // Contract address should be configured
		"getBatchEnvironmentData(string)",
		[]interface{}{batchID},
	)
	if err != nil {
		return nil, err
	}
	
	// Convert results to appropriate types
	eventsBytes, ok := eventsResult.([]byte)
	if !ok {
		return nil, fmt.Errorf("unexpected events result type")
	}
	
	transfersBytes, ok := transfersResult.([]byte)
	if !ok {
		return nil, fmt.Errorf("unexpected transfers result type")
	}
	
	envDataBytes, ok := envDataResult.([]byte)
	if !ok {
		return nil, fmt.Errorf("unexpected environment data result type")
	}
	
	// Parse JSON
	var events []map[string]interface{}
	if err := json.Unmarshal(eventsBytes, &events); err != nil {
		return nil, err
	}
	
	var transfers []map[string]interface{}
	if err := json.Unmarshal(transfersBytes, &transfers); err != nil {
		return nil, err
	}
	
	var envData []map[string]interface{}
	if err := json.Unmarshal(envDataBytes, &envData); err != nil {
		return nil, err
	}
	
	return &BatchEventData{
		BatchID:         batchID,
		Events:          events,
		Transfers:       transfers,
		EnvironmentData: envData,
	}, nil
}

// detectTimeGapAnomalies detects unusual time gaps between events
func (s *AnomalyDetectionService) detectTimeGapAnomalies(batchData *BatchEventData) ([]AnomalyDetectionResult, error) {
	var anomalies []AnomalyDetectionResult
	
	// Combine events and transfers, sort by timestamp
	type timeEvent struct {
		Type      string
		ID        string
		Timestamp time.Time
		Data      map[string]interface{}
	}
	
	var timeEvents []timeEvent
	
	// Add events
	for _, event := range batchData.Events {
		timestampStr, ok := event["timestamp"].(string)
		if !ok {
			continue
		}
		
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			continue
		}
		
		id, _ := event["id"].(string)
		
		timeEvents = append(timeEvents, timeEvent{
			Type:      "event",
			ID:        id,
			Timestamp: timestamp,
			Data:      event,
		})
	}
	
	// Add transfers
	for _, transfer := range batchData.Transfers {
		timestampStr, ok := transfer["timestamp"].(string)
		if !ok {
			continue
		}
		
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			continue
		}
		
		id, _ := transfer["id"].(string)
		
		timeEvents = append(timeEvents, timeEvent{
			Type:      "transfer",
			ID:        id,
			Timestamp: timestamp,
			Data:      transfer,
		})
	}
	
	// Sort by timestamp
	sort.Slice(timeEvents, func(i, j int) bool {
		return timeEvents[i].Timestamp.Before(timeEvents[j].Timestamp)
	})
	
	// Check for unusual time gaps
	if len(timeEvents) < 2 {
		return anomalies, nil
	}
	
	// Calculate average time gap
	var totalGap time.Duration
	for i := 1; i < len(timeEvents); i++ {
		gap := timeEvents[i].Timestamp.Sub(timeEvents[i-1].Timestamp)
		totalGap += gap
	}
	avgGap := totalGap / time.Duration(len(timeEvents)-1)
	
	// Check each gap
	for i := 1; i < len(timeEvents); i++ {
		gap := timeEvents[i].Timestamp.Sub(timeEvents[i-1].Timestamp)
		
		// If gap is more than 3 times the average, flag as anomaly
		if gap > avgGap*3 {
			anomaly := AnomalyDetectionResult{
				AnomalyType:   AnomalyTypeTimeGap,
				BatchID:       batchData.BatchID,
				Timestamp:     timeEvents[i].Timestamp,
				Confidence:    calculateConfidence(float64(gap), float64(avgGap), 3),
				ExpectedValue: avgGap.String(),
				ActualValue:   gap.String(),
				Description:   fmt.Sprintf("Unusual time gap of %s detected between events", gap),
				RelatedEvents: []string{timeEvents[i-1].ID, timeEvents[i].ID},
				Recommendations: []string{
					"Check for missing events during this time period",
					"Verify if there was any issue with data recording",
				},
				Metadata: map[string]interface{}{
					"avgGap":     avgGap.String(),
					"prevEvent":  timeEvents[i-1].Type,
					"nextEvent":  timeEvents[i].Type,
					"gapSeconds": gap.Seconds(),
				},
			}
			
			anomalies = append(anomalies, anomaly)
		}
	}
	
	return anomalies, nil
}

// detectTemperatureAnomalies detects unusual temperature readings
func (s *AnomalyDetectionService) detectTemperatureAnomalies(batchData *BatchEventData) ([]AnomalyDetectionResult, error) {
	var anomalies []AnomalyDetectionResult
	
	var temperatures []float64
	var timestamps []time.Time
	
	// Extract temperature data
	for _, envData := range batchData.EnvironmentData {
		tempValue, ok := envData["temperature"].(float64)
		if !ok {
			continue
		}
		
		timestampStr, ok := envData["timestamp"].(string)
		if !ok {
			continue
		}
		
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			continue
		}
		
		temperatures = append(temperatures, tempValue)
		timestamps = append(timestamps, timestamp)
	}
	
	if len(temperatures) < 3 {
		return anomalies, nil
	}
	
	// Calculate mean and standard deviation
	mean, stdDev := calculateStats(temperatures)
	
	// Check for outliers (values more than 3 std deviations from mean)
	for i, temp := range temperatures {
		if math.Abs(temp-mean) > 3*stdDev {
			anomaly := AnomalyDetectionResult{
				AnomalyType:   AnomalyTypeTemperature,
				BatchID:       batchData.BatchID,
				Timestamp:     timestamps[i],
				Confidence:    calculateConfidence(math.Abs(temp-mean), 3*stdDev, 1),
				ExpectedValue: fmt.Sprintf("%.2f°C ± %.2f°C", mean, stdDev),
				ActualValue:   fmt.Sprintf("%.2f°C", temp),
				Description:   fmt.Sprintf("Unusual temperature reading: %.2f°C (expected range: %.2f°C to %.2f°C)", temp, mean-2*stdDev, mean+2*stdDev),
				Recommendations: []string{
					"Verify temperature sensor calibration",
					"Check if environmental controls were functioning properly",
					"Review handling procedures during this time",
				},
				Metadata: map[string]interface{}{
					"mean":        mean,
					"stdDev":      stdDev,
					"zScore":      math.Abs(temp-mean) / stdDev,
					"sensorType":  "temperature",
					"unitOfMeasure": "Celsius",
				},
			}
			
			anomalies = append(anomalies, anomaly)
		}
	}
	
	return anomalies, nil
}

// detectHumidityAnomalies detects unusual humidity readings
func (s *AnomalyDetectionService) detectHumidityAnomalies(batchData *BatchEventData) ([]AnomalyDetectionResult, error) {
	var anomalies []AnomalyDetectionResult
	
	var humidities []float64
	var timestamps []time.Time
	
	// Extract humidity data
	for _, envData := range batchData.EnvironmentData {
		humidityValue, ok := envData["humidity"].(float64)
		if !ok {
			continue
		}
		
		timestampStr, ok := envData["timestamp"].(string)
		if !ok {
			continue
		}
		
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			continue
		}
		
		humidities = append(humidities, humidityValue)
		timestamps = append(timestamps, timestamp)
	}
	
	if len(humidities) < 3 {
		return anomalies, nil
	}
	
	// Calculate mean and standard deviation
	mean, stdDev := calculateStats(humidities)
	
	// Check for outliers (values more than 3 std deviations from mean)
	for i, humidity := range humidities {
		if math.Abs(humidity-mean) > 3*stdDev {
			anomaly := AnomalyDetectionResult{
				AnomalyType:   AnomalyTypeHumidity,
				BatchID:       batchData.BatchID,
				Timestamp:     timestamps[i],
				Confidence:    calculateConfidence(math.Abs(humidity-mean), 3*stdDev, 1),
				ExpectedValue: fmt.Sprintf("%.2f%% ± %.2f%%", mean, stdDev),
				ActualValue:   fmt.Sprintf("%.2f%%", humidity),
				Description:   fmt.Sprintf("Unusual humidity reading: %.2f%% (expected range: %.2f%% to %.2f%%)", humidity, mean-2*stdDev, mean+2*stdDev),
				Recommendations: []string{
					"Verify humidity sensor calibration",
					"Check if environmental controls were functioning properly",
					"Review storage conditions during this time",
				},
				Metadata: map[string]interface{}{
					"mean":        mean,
					"stdDev":      stdDev,
					"zScore":      math.Abs(humidity-mean) / stdDev,
					"sensorType":  "humidity",
					"unitOfMeasure": "percent",
				},
			}
			
			anomalies = append(anomalies, anomaly)
		}
	}
	
	return anomalies, nil
}

// detectLocationAnomalies detects unusual location changes
func (s *AnomalyDetectionService) detectLocationAnomalies(batchData *BatchEventData) ([]AnomalyDetectionResult, error) {
	var anomalies []AnomalyDetectionResult
	
	type locationEvent struct {
		Latitude   float64
		Longitude  float64
		Timestamp  time.Time
		Type       string
		ID         string
	}
	
	var locationEvents []locationEvent
	
	// Extract location data from transfers
	for _, transfer := range batchData.Transfers {
		locationMap, ok := transfer["location"].(map[string]interface{})
		if !ok {
			continue
		}
		
		lat, latOk := locationMap["latitude"].(float64)
		lng, lngOk := locationMap["longitude"].(float64)
		if !latOk || !lngOk {
			continue
		}
		
		timestampStr, ok := transfer["timestamp"].(string)
		if !ok {
			continue
		}
		
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			continue
		}
		
		id, _ := transfer["id"].(string)
		
		locationEvents = append(locationEvents, locationEvent{
			Latitude:   lat,
			Longitude:  lng,
			Timestamp:  timestamp,
			Type:       "transfer",
			ID:         id,
		})
	}
	
	// Sort by timestamp
	sort.Slice(locationEvents, func(i, j int) bool {
		return locationEvents[i].Timestamp.Before(locationEvents[j].Timestamp)
	})
	
	// Need at least two locations to detect anomalies
	if len(locationEvents) < 2 {
		return anomalies, nil
	}
	
	// Check for unusually large distances between consecutive locations
	for i := 1; i < len(locationEvents); i++ {
		prev := locationEvents[i-1]
		curr := locationEvents[i]
		
		distance := calculateHaversineDistance(
			prev.Latitude, prev.Longitude,
			curr.Latitude, curr.Longitude,
		)
		
		timeDiff := curr.Timestamp.Sub(prev.Timestamp)
		speedKmh := distance / (float64(timeDiff.Hours()))
		
		// Flag if speed exceeds 150 km/h (unusually fast for logistics)
		if speedKmh > 150 {
			anomaly := AnomalyDetectionResult{
				AnomalyType:   AnomalyTypeLocation,
				BatchID:       batchData.BatchID,
				TransferID:    curr.ID,
				Timestamp:     curr.Timestamp,
				Confidence:    calculateConfidence(speedKmh, 150, 0.5),
				ExpectedValue: "< 150 km/h",
				ActualValue:   fmt.Sprintf("%.2f km/h", speedKmh),
				Description:   fmt.Sprintf("Unusual speed between locations: %.2f km/h over %.2f km", speedKmh, distance),
				RelatedEvents: []string{prev.ID, curr.ID},
				Recommendations: []string{
					"Verify location data accuracy",
					"Check if there could be data input errors",
					"Investigate transport method used",
				},
				Metadata: map[string]interface{}{
					"distance":    distance,
					"timeElapsed": timeDiff.String(),
					"prevLat":     prev.Latitude,
					"prevLng":     prev.Longitude,
					"currLat":     curr.Latitude,
					"currLng":     curr.Longitude,
				},
			}
			
			anomalies = append(anomalies, anomaly)
		}
	}
	
	return anomalies, nil
}

// detectAuthorizationAnomalies detects authorization-related anomalies
func (s *AnomalyDetectionService) detectAuthorizationAnomalies(batchData *BatchEventData) ([]AnomalyDetectionResult, error) {
	var anomalies []AnomalyDetectionResult
	
	// Extract authorization events
	for _, event := range batchData.Events {
		eventType, ok := event["eventType"].(string)
		if !ok {
			continue
		}
		
		// Look for authorization-related events
		if eventType == "AUTHORIZATION_FAILURE" || eventType == "UNAUTHORIZED_ACCESS" {
			timestampStr, ok := event["timestamp"].(string)
			if !ok {
				continue
			}
			
			timestamp, err := time.Parse(time.RFC3339, timestampStr)
			if err != nil {
				continue
			}
			
			id, _ := event["id"].(string)
			description, _ := event["description"].(string)
			actor, _ := event["actor"].(string)
			
			anomaly := AnomalyDetectionResult{
				AnomalyType:   AnomalyTypeAuthorization,
				BatchID:       batchData.BatchID,
				Timestamp:     timestamp,
				Confidence:    1.0, // High confidence since this is a direct report
				Description:   description,
				RelatedEvents: []string{id},
				Recommendations: []string{
					"Review access controls and permissions",
					"Verify identity of the actor",
					"Check for potential security breaches",
				},
				Metadata: map[string]interface{}{
					"eventType": eventType,
					"actor":     actor,
				},
			}
			
			anomalies = append(anomalies, anomaly)
		}
	}
	
	return anomalies, nil
}

// Helper functions

// calculateStats calculates mean and standard deviation
func calculateStats(values []float64) (mean, stdDev float64) {
	if len(values) == 0 {
		return 0, 0
	}
	
	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))
	
	// Calculate standard deviation
	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}
	stdDev = math.Sqrt(sumSquaredDiff / float64(len(values)))
	
	return mean, stdDev
}

// calculateHaversineDistance calculates the great-circle distance between two points in km
func calculateHaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const r = 6371.0 // Earth radius in km
	
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	return r * c
}

// calculateConfidence calculates a confidence score between 0 and 1
func calculateConfidence(actual, threshold, steepness float64) float64 {
	// Higher ratio means more confidence
	ratio := actual / threshold
	
	// Use sigmoid function to scale the confidence
	confidence := 1.0 / (1.0 + math.Exp(-steepness*(ratio-1.0)))
	
	// Ensure the value is between 0 and 1
	if confidence < 0 {
		confidence = 0
	} else if confidence > 1 {
		confidence = 1
	}
	
	return confidence
}
