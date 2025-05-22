package ipfs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// PinataService represents a client for interacting with Pinata Cloud
type PinataService struct {
	JWT            string
	APIKey         string
	APISecret      string
	BaseURL        string
	GatewayURL     string
	PinTimeout     time.Duration
	UseGatewayCheck bool
	GatewayCheckAttempts int
	APIVersion     string
}

// PinataPinResponse represents a response from Pinata pinning API
type PinataPinResponse struct {
	IpfsHash    string `json:"IpfsHash"`
	PinSize     int    `json:"PinSize"`
	Timestamp   string `json:"Timestamp"`
	Status      string `json:"status,omitempty"`
	Message     string `json:"message,omitempty"`
	IsSuccess   bool   `json:"isSuccess,omitempty"`
}

// PinataPinnedListResponse represents a response from Pinata pinned files list API
type PinataPinnedListResponse struct {
	Count   int         `json:"count"`
	Rows    []PinataPin `json:"rows"`
	Status  string      `json:"status,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PinataPin represents a pinned file in Pinata
type PinataPin struct {
	ID            string   `json:"id"`
	IpfsHash      string   `json:"ipfs_pin_hash"`
	Name          string   `json:"metadata"`
	Date          string   `json:"date_pinned"`
	Size          int      `json:"size"`
	UserProvidedName string `json:"user_provided_name,omitempty"`
}

// PinataMetadata represents metadata for pinning
type PinataMetadata struct {
	Name       string            `json:"name"`
	KeyValues  map[string]string `json:"keyvalues"`
}

// PinataPinOptions represents options for pinning
type PinataPinOptions struct {
	CIDVersion    int  `json:"cidVersion,omitempty"`
	WrapWithDir   bool `json:"wrapWithDirectory,omitempty"`
	CustomPinPolicy *PinataCustomPinPolicy `json:"customPinPolicy,omitempty"`
}

// PinataCustomPinPolicy represents a custom pin policy
type PinataCustomPinPolicy struct {
	Regions []PinataRegion `json:"regions"`
}

// PinataRegion represents a region for pinning
type PinataRegion struct {
	ID                      string `json:"id"`
	DesiredReplicationCount int    `json:"desiredReplicationCount"`
}

// NewPinataService creates a new Pinata service
func NewPinataService() *PinataService {
	// Get Pinata JWT token from environment
	jwt := os.Getenv("PINATA_JWT")
	apiKey := os.Getenv("PINATA_API_KEY")
	apiSecret := os.Getenv("PINATA_API_SECRET")
	
	// Get Pinata Gateway URL
	gatewayURL := os.Getenv("PINATA_GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "https://gateway.pinata.cloud"
	}
	
	// Get Pinata Pin Timeout
	pinTimeout := 180 * time.Second
	if timeoutStr := os.Getenv("PINATA_PIN_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			pinTimeout = time.Duration(timeout) * time.Second
		}
	}
	
	// Get API Version
	apiVersion := os.Getenv("PINATA_API_VERSION")
	if apiVersion == "" {
		apiVersion = "v1"
	}
	
	// Get gateway check preferences
	useGatewayCheck := true
	if checkStr := os.Getenv("PINATA_USE_GATEWAY_CHECK"); checkStr == "false" {
		useGatewayCheck = false
	}
	
	// Get gateway check attempts
	gatewayCheckAttempts := 3
	if attemptsStr := os.Getenv("PINATA_GATEWAY_CHECK_ATTEMPTS"); attemptsStr != "" {
		if attempts, err := strconv.Atoi(attemptsStr); err == nil {
			gatewayCheckAttempts = attempts
		}
	}
	
	return &PinataService{
		JWT:            jwt,
		APIKey:         apiKey,
		APISecret:      apiSecret,
		BaseURL:        "https://api.pinata.cloud",
		GatewayURL:     gatewayURL,
		PinTimeout:     pinTimeout,
		UseGatewayCheck: useGatewayCheck,
		GatewayCheckAttempts: gatewayCheckAttempts,
		APIVersion:     apiVersion,
	}
}

// PinFile pins a file to Pinata Cloud
func (p *PinataService) PinFile(file multipart.File, filename string, metadata map[string]string) (*PinataPinResponse, error) {
	endpoint := fmt.Sprintf("%s/pinning/pinFileToIPFS", p.BaseURL)
	
	// Create a new multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Add the file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %v", err)
	}
	
	// Copy the file content to the part
	if _, err = io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %v", err)
	}
	
	// Add metadata if provided
	if metadata != nil && len(metadata) > 0 {
		// Create Pinata metadata structure
		pinataMeta := PinataMetadata{
			Name:      filename,
			KeyValues: metadata,
		}
		
		// Marshal to JSON
		metadataJSON, err := json.Marshal(pinataMeta)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %v", err)
		}
		
		// Add to form
		if err := writer.WriteField("pinataMetadata", string(metadataJSON)); err != nil {
			return nil, fmt.Errorf("failed to add metadata field: %v", err)
		}
	}
	
	// Close the writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %v", err)
	}
	
	// Create request
	req, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	// Use JWT if available, otherwise use API key/secret
	if p.JWT != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.JWT))
	} else {
		req.Header.Set("pinata_api_key", p.APIKey)
		req.Header.Set("pinata_secret_api_key", p.APISecret)
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), p.PinTimeout)
	defer cancel()
	
	// Execute the request with the context
	req = req.WithContext(ctx)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	
	// Check for error response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pinning failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	
	// Parse the response
	var pinResponse PinataPinResponse
	if err := json.Unmarshal(respBody, &pinResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}
	
	// Verify the file is accessible via gateway if enabled
	if p.UseGatewayCheck && pinResponse.IpfsHash != "" {
		if err := p.verifyPinGatewayAccess(pinResponse.IpfsHash); err != nil {
			return &pinResponse, fmt.Errorf("file pinned but gateway access verification failed: %v", err)
		}
	}
	
	return &pinResponse, nil
}

// PinJSON pins JSON data to Pinata Cloud
func (p *PinataService) PinJSON(data interface{}, name string, metadata map[string]string) (*PinataPinResponse, error) {
	endpoint := fmt.Sprintf("%s/pinning/pinJSONToIPFS", p.BaseURL)
	
	// Create the request body
	requestBody := map[string]interface{}{
		"pinataContent": data,
	}
	
	// Add metadata if provided
	if metadata != nil && len(metadata) > 0 {
		pinataMeta := PinataMetadata{
			Name:      name,
			KeyValues: metadata,
		}
		requestBody["pinataMetadata"] = pinataMeta
	}
	
	// Marshal the request body
	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}
	
	// Create request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	
	// Use JWT if available, otherwise use API key/secret
	if p.JWT != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.JWT))
	} else {
		req.Header.Set("pinata_api_key", p.APIKey)
		req.Header.Set("pinata_secret_api_key", p.APISecret)
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), p.PinTimeout)
	defer cancel()
	
	// Execute the request with the context
	req = req.WithContext(ctx)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()
		// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	
	// Log the response for debugging
	fmt.Printf("Pinata API Response (status %d): %s\n", resp.StatusCode, string(respBody))
	
	// Check for error response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pinning JSON failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	
	// Parse the response
	var pinResponse PinataPinResponse
	if err := json.Unmarshal(respBody, &pinResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}
	
	// Verify the file is accessible via gateway if enabled
	if p.UseGatewayCheck && pinResponse.IpfsHash != "" {
		if err := p.verifyPinGatewayAccess(pinResponse.IpfsHash); err != nil {
			return &pinResponse, fmt.Errorf("JSON pinned but gateway access verification failed: %v", err)
		}
	}
	
	return &pinResponse, nil
}

// PinByCID pins an existing IPFS CID to Pinata Cloud
func (p *PinataService) PinByCID(cid string, name string, metadata map[string]string) (*PinataPinResponse, error) {
	endpoint := fmt.Sprintf("%s/pinning/pinByHash", p.BaseURL)
	
	// Create the request body
	requestBody := map[string]interface{}{
		"hashToPin": cid,
	}
	
	// Add metadata if provided
	if metadata != nil && len(metadata) > 0 {
		pinataMeta := PinataMetadata{
			Name:      name,
			KeyValues: metadata,
		}
		requestBody["pinataMetadata"] = pinataMeta
	}
	
	// Marshal the request body
	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}
	
	// Create request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	
	// Use JWT if available, otherwise use API key/secret
	if p.JWT != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.JWT))
	} else {
		req.Header.Set("pinata_api_key", p.APIKey)
		req.Header.Set("pinata_secret_api_key", p.APISecret)
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), p.PinTimeout)
	defer cancel()
	
	// Execute the request with the context
	req = req.WithContext(ctx)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	
	// Check for error response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pinning by CID failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	
	// For pinByHash, we create our own response since Pinata's response format differs
	pinResponse := &PinataPinResponse{
		IpfsHash:  cid,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	
	// Verify the file is accessible via gateway if enabled
	if p.UseGatewayCheck && cid != "" {
		if err := p.verifyPinGatewayAccess(cid); err != nil {
			return pinResponse, fmt.Errorf("CID pinned but gateway access verification failed: %v", err)
		}
	}
	
	return pinResponse, nil
}

// UnpinByCID unpins a file from Pinata Cloud
func (p *PinataService) UnpinByCID(cid string) error {
	endpoint := fmt.Sprintf("%s/pinning/unpin/%s", p.BaseURL, cid)
	
	// Create request
	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	
	// Set headers
	// Use JWT if available, otherwise use API key/secret
	if p.JWT != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.JWT))
	} else {
		req.Header.Set("pinata_api_key", p.APIKey)
		req.Header.Set("pinata_secret_api_key", p.APISecret)
	}
	
	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()
	
	// Check for error response
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unpinning failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	
	return nil
}

// GetPinnedCIDs retrieves a list of pinned CIDs from Pinata Cloud
func (p *PinataService) GetPinnedCIDs(metadata map[string]string) (*PinataPinnedListResponse, error) {
	endpoint := fmt.Sprintf("%s/data/pinList", p.BaseURL)
	
	// Add query parameters for metadata
	if metadata != nil && len(metadata) > 0 {
		endpoint += "?"
		for key, value := range metadata {
			endpoint += fmt.Sprintf("metadata[keyvalues][%s]=%s&", key, value)
		}
		// Remove trailing '&'
		endpoint = endpoint[:len(endpoint)-1]
	}
	
	// Create request
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// Set headers
	// Use JWT if available, otherwise use API key/secret
	if p.JWT != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.JWT))
	} else {
		req.Header.Set("pinata_api_key", p.APIKey)
		req.Header.Set("pinata_secret_api_key", p.APISecret)
	}
	
	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	
	// Check for error response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getting pinned list failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	
	// Parse the response
	var pinnedResponse PinataPinnedListResponse
	if err := json.Unmarshal(respBody, &pinnedResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}
	
	return &pinnedResponse, nil
}

// CreatePinataGatewayURL creates a URL for accessing a file on Pinata gateway
func (p *PinataService) CreatePinataGatewayURL(cid string) string {
	if cid == "" {
		return ""
	}
	
	gatewayURL := p.GatewayURL
	// Make sure we're using the public gateway URL
	if gatewayURL == "" || !strings.Contains(gatewayURL, "gateway.pinata.cloud") {
		gatewayURL = "https://gateway.pinata.cloud"
	}
	
	// Ensure proper URL format for IPFS gateway
	if !strings.HasSuffix(gatewayURL, "/") && !strings.HasSuffix(gatewayURL, "/ipfs/") {
		gatewayURL = gatewayURL + "/ipfs/"
	} else if strings.HasSuffix(gatewayURL, "/") {
		gatewayURL = gatewayURL + "ipfs/"
	}
	
	// Final URL construction
	url := gatewayURL + cid
	fmt.Printf("Created Pinata gateway URL: %s\n", url)
	return url
}

// verifyPinGatewayAccess verifies that a CID is accessible via the Pinata gateway
func (p *PinataService) verifyPinGatewayAccess(cid string) error {
	gatewayURL := p.CreatePinataGatewayURL(cid)
	
	// Retry multiple times with backoff
	for attempt := 0; attempt < p.GatewayCheckAttempts; attempt++ {
		// If not first attempt, wait with exponential backoff
		if attempt > 0 {
			backoffTime := time.Duration(2<<uint(attempt-1)) * time.Second
			time.Sleep(backoffTime)
		}
		
		// Create request
		req, err := http.NewRequest("HEAD", gatewayURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create gateway check request: %v", err)
		}
		
		// Execute the request
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			continue // Retry on error
		}
		defer resp.Body.Close()
		
		// If successful, return nil
		if resp.StatusCode == http.StatusOK {
			return nil
		}
	}
	
	return fmt.Errorf("failed to verify gateway access after %d attempts", p.GatewayCheckAttempts)
}

// TestPinataConnection tests the connection to Pinata Cloud
func (p *PinataService) TestPinataConnection() error {
	// Use JWT for authentication if available
	if p.JWT != "" || (p.APIKey != "" && p.APISecret != "") {
		// Test access by getting user data
		endpoint := fmt.Sprintf("%s/data/userPinnedDataTotal", p.BaseURL)
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			return fmt.Errorf("failed to create test request: %v", err)
		}
		
		// Set authorization headers
		if p.JWT != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.JWT))
		} else {
			req.Header.Set("pinata_api_key", p.APIKey)
			req.Header.Set("pinata_secret_api_key", p.APISecret)
		}
		
		// Execute the request with a timeout
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to connect to Pinata Cloud: %v", err)
		}
		defer resp.Body.Close()
		
		// Check if the response is successful
		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("Pinata authentication failed with status %d: %s", resp.StatusCode, string(respBody))
		}
		
		return nil
	}
	
	return fmt.Errorf("Pinata JWT or API Key/Secret not configured")
}
