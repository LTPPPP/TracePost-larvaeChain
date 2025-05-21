package blockchain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// IBMBlockchainClient represents a client for interacting with IBM Blockchain Platform
type IBMBlockchainClient struct {
	Config     IBMBlockchainConfig
	HTTPClient *http.Client
}

// IBMBlockchainConfig contains configuration for connecting to IBM Blockchain Platform
type IBMBlockchainConfig struct {
	APIEndpoint      string
	APIKey           string
	OrganizationID   string
	ServiceInstanceID string
	BearerToken      string
	TokenExpiry      time.Time
	NetworkID        string
	ChannelName      string
	ChaincodeName    string
}

// NewIBMBlockchainClient creates a new IBM Blockchain Platform client
func NewIBMBlockchainClient(config IBMBlockchainConfig) *IBMBlockchainClient {
	return &IBMBlockchainClient{
		Config: config,
		HTTPClient: &http.Client{
			Timeout: time.Duration(30) * time.Second,
		},
	}
}

// Authenticate authenticates with the IBM Blockchain Platform API
func (ibc *IBMBlockchainClient) Authenticate() error {
	// Check if we already have a valid token
	if ibc.Config.BearerToken != "" && time.Now().Before(ibc.Config.TokenExpiry) {
		return nil
	}
	
	// Prepare authentication request
	authURL := fmt.Sprintf("%s/auth/token", ibc.Config.APIEndpoint)
	authPayload := map[string]string{
		"api_key": ibc.Config.APIKey,
	}
	
	// Convert to JSON
	authPayloadJSON, err := json.Marshal(authPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal authentication payload: %w", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequest("POST", authURL, bytes.NewBuffer(authPayloadJSON))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	
	// Send request
	resp, err := ibc.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status code: %d", resp.StatusCode)
	}
	
	// Parse response
	var authResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	if err != nil {
		return fmt.Errorf("failed to decode authentication response: %w", err)
	}
	
	// Store token and expiry
	ibc.Config.BearerToken = authResponse.AccessToken
	ibc.Config.TokenExpiry = time.Now().Add(time.Duration(authResponse.ExpiresIn) * time.Second)
	
	return nil
}

// InvokeChaincode invokes a chaincode function on the IBM Blockchain Platform
func (ibc *IBMBlockchainClient) InvokeChaincode(ctx context.Context, functionName string, args []string) (string, error) {
	// Authenticate if needed
	err := ibc.Authenticate()
	if err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}
	
	// Prepare invoke request
	invokeURL := fmt.Sprintf(
		"%s/networks/%s/channels/%s/chaincodes/%s/invoke",
		ibc.Config.APIEndpoint,
		ibc.Config.NetworkID,
		ibc.Config.ChannelName,
		ibc.Config.ChaincodeName,
	)
	
	invokePayload := map[string]interface{}{
		"function": functionName,
		"args":     args,
	}
	
	// Convert to JSON
	invokePayloadJSON, err := json.Marshal(invokePayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal invoke payload: %w", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequest("POST", invokeURL, bytes.NewBuffer(invokePayloadJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ibc.Config.BearerToken)
	
	// Send request
	resp, err := ibc.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("chaincode invocation failed with status code: %d", resp.StatusCode)
	}
	
	// Parse response
	var invokeResponse struct {
		TxID    string `json:"txid"`
		Result  string `json:"result"`
		Message string `json:"message"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&invokeResponse)
	if err != nil {
		return "", fmt.Errorf("failed to decode invoke response: %w", err)
	}
	
	return invokeResponse.TxID, nil
}

// QueryChaincode queries a chaincode function on the IBM Blockchain Platform
func (ibc *IBMBlockchainClient) QueryChaincode(ctx context.Context, functionName string, args []string) (map[string]interface{}, error) {
	// Authenticate if needed
	err := ibc.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	
	// Prepare query request
	queryURL := fmt.Sprintf(
		"%s/networks/%s/channels/%s/chaincodes/%s/query",
		ibc.Config.APIEndpoint,
		ibc.Config.NetworkID,
		ibc.Config.ChannelName,
		ibc.Config.ChaincodeName,
	)
	
	queryPayload := map[string]interface{}{
		"function": functionName,
		"args":     args,
	}
	
	// Convert to JSON
	queryPayloadJSON, err := json.Marshal(queryPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query payload: %w", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequest("POST", queryURL, bytes.NewBuffer(queryPayloadJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ibc.Config.BearerToken)
	
	// Send request
	resp, err := ibc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chaincode query failed with status code: %d", resp.StatusCode)
	}
	
	// Parse response
	var queryResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&queryResult)
	if err != nil {
		return nil, fmt.Errorf("failed to decode query response: %w", err)
	}
	
	return queryResult, nil
}

// GetNetworkStatus gets the status of the blockchain network
func (ibc *IBMBlockchainClient) GetNetworkStatus(ctx context.Context) (map[string]interface{}, error) {
	// Authenticate if needed
	err := ibc.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	
	// Prepare status request
	statusURL := fmt.Sprintf(
		"%s/networks/%s/status",
		ibc.Config.APIEndpoint,
		ibc.Config.NetworkID,
	)
	
	// Create HTTP request
	req, err := http.NewRequest("GET", statusURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Authorization", "Bearer "+ibc.Config.BearerToken)
	
	// Send request
	resp, err := ibc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get network status failed with status code: %d", resp.StatusCode)
	}
	
	// Parse response
	var statusResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&statusResult)
	if err != nil {
		return nil, fmt.Errorf("failed to decode status response: %w", err)
	}
	
	return statusResult, nil
}
