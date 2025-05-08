package blockchain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"encoding/base64"
	"time"
	
	"github.com/LTPPPP/TracePost-larvaeChain/config"
)

// BaaSService provides Blockchain-as-a-Service functionality
type BaaSService struct {
	Config     *config.BaaSConfig
	HTTPClient *http.Client
	Networks   map[string]*BaaSNetwork
}

// BaaSNetwork represents a connected blockchain network
type BaaSNetwork struct {
	Config          NetworkConfig
	ActiveEndpoint  string
	ConnectionState string // "connected", "disconnected", "error"
	LastSync        time.Time
	NodeInfo        map[string]interface{}
	BlockHeight     int64
}

// NetworkConfig represents the configuration for a blockchain network
type NetworkConfig struct {
	NetworkID       string
	ChainType       string
	NodeEndpoints   []string
	RPCEndpoint     string
	IsMainnet       bool
	IBCEnabled      bool
	XCMEnabled      bool
	NetworkParams   map[string]interface{}
	ExplorerURL     string
}

// NewBaaSService creates a new BaaS service instance
func NewBaaSService() *BaaSService {
	cfg, err := config.LoadBaaSConfig("config/baas-config.json")
	if err != nil {
		// If config file not found, create a default one
		cfg = config.CreateDefaultConfig()
	}
	
	// Initialize HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(cfg.APIConfig.RequestTimeout) * time.Second,
	}
	
	// Create the service instance
	service := &BaaSService{
		Config:     cfg,
		HTTPClient: client,
		Networks:   make(map[string]*BaaSNetwork),
	}
	
	// Initialize networks
	for _, netCfg := range cfg.Networks {
		// Skip disabled networks
		if !netCfg.Enabled {
			continue
		}
		
		// Select first endpoint as active by default
		activeEndpoint := ""
		if len(netCfg.Endpoints) > 0 {
			activeEndpoint = netCfg.Endpoints[0]
		}
		
		service.Networks[netCfg.NetworkID] = &BaaSNetwork{
			Config: NetworkConfig{
				NetworkID:       netCfg.NetworkID,
				ChainType:       netCfg.NetworkType,
				NodeEndpoints:   netCfg.Endpoints,
				RPCEndpoint:     activeEndpoint,
				IsMainnet:       true, // Default to mainnet unless specified otherwise
				IBCEnabled:      netCfg.NetworkType == "cosmos", // Enable IBC for Cosmos chains by default
				XCMEnabled:      netCfg.NetworkType == "polkadot" || netCfg.NetworkType == "substrate", // Enable XCM for Polkadot chains
				NetworkParams:   netCfg.NetworkParams,
				ExplorerURL:     "", // Will be populated later if available
			},
			ActiveEndpoint:  activeEndpoint,
			ConnectionState: "disconnected",
			NodeInfo:        make(map[string]interface{}),
		}
	}
	
	return service
}

// ConnectToNetwork establishes a connection to a blockchain network
func (s *BaaSService) ConnectToNetwork(networkID string) error {
	network, exists := s.Networks[networkID]
	if !exists {
		return fmt.Errorf("network %s not configured", networkID)
	}
	
	// If already connected, just return
	if network.ConnectionState == "connected" {
		return nil
	}
	
	// Try to connect to each endpoint until success
	var lastError error
	for _, endpoint := range network.Config.NodeEndpoints {
		// Set active endpoint
		network.ActiveEndpoint = endpoint
		
		// Try to get node status
		nodeInfo, blockHeight, err := s.getNodeStatus(network)
		if err == nil {
			// Connection successful
			network.ConnectionState = "connected"
			network.LastSync = time.Now()
			network.NodeInfo = nodeInfo
			network.BlockHeight = blockHeight
			return nil
		}
		
		lastError = err
	}
	
	// Set to error state if all endpoints failed
	network.ConnectionState = "error"
	return fmt.Errorf("failed to connect to network %s: %v", networkID, lastError)
}

// GetNodeStatus retrieves the current status of a node
func (s *BaaSService) getNodeStatus(network *BaaSNetwork) (map[string]interface{}, int64, error) {
	// Construct the URL based on chain type
	var url string
	switch network.Config.ChainType {
	case "cosmos":
		url = fmt.Sprintf("%s/status", network.ActiveEndpoint)
	case "substrate", "polkadot":
		url = fmt.Sprintf("%s/system/health", network.ActiveEndpoint)
	default:
		url = fmt.Sprintf("%s/status", network.ActiveEndpoint)
	}
	
	// Send request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	
	// Try to get API key for the network
	networkConfig, err := s.Config.GetNetworkConfig(network.Config.NetworkID)
	if err == nil && networkConfig.ApiKeys != nil {
		if apiKey, ok := networkConfig.ApiKeys["baas"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("node returned status: %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}
	
	// Extract block height based on chain type
	var blockHeight int64
	
	switch network.Config.ChainType {
	case "cosmos":
		if syncInfo, ok := result["sync_info"].(map[string]interface{}); ok {
			if height, ok := syncInfo["latest_block_height"].(string); ok {
				fmt.Sscanf(height, "%d", &blockHeight)
			}
		}
	case "substrate", "polkadot":
		if result["isSyncing"] == false {
			// For substrate, we need to make another call to get the block height
			blockHeight, _ = s.getSubstrateBlockHeight(network)
		}
	}
	
	return result, blockHeight, nil
}

// getSubstrateBlockHeight gets the current block height for a Substrate/Polkadot chain
func (s *BaaSService) getSubstrateBlockHeight(network *BaaSNetwork) (int64, error) {
	// Construct request for getting block number
	payload := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "chain_getBlockNumber",
		"params":  []interface{}{},
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	
	// Send request
	resp, err := http.Post(
		network.Config.RPCEndpoint,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	// Parse response
	var result struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	
	// Convert hex to int64
	var blockHeight int64
	fmt.Sscanf(result.Result, "0x%x", &blockHeight)
	
	return blockHeight, nil
}

// CreateCosmosIBCClient creates an IBC client for cross-chain communication in Cosmos
func (s *BaaSService) CreateCosmosIBCClient(networkID, targetNetworkID string) (string, error) {
	// Check if networks exist
	sourceNetwork, exists := s.Networks[networkID]
	if !exists {
		return "", fmt.Errorf("source network %s not configured", networkID)
	}
	
	targetNetwork, exists := s.Networks[targetNetworkID]
	if !exists {
		return "", fmt.Errorf("target network %s not configured", targetNetworkID)
	}
	
	// Check if the networks support IBC
	if sourceNetwork.Config.ChainType != "cosmos" || !sourceNetwork.Config.IBCEnabled {
		return "", fmt.Errorf("source network %s does not support IBC", networkID)
	}
	
	if targetNetwork.Config.ChainType != "cosmos" || !targetNetwork.Config.IBCEnabled {
		return "", fmt.Errorf("target network %s does not support IBC", targetNetworkID)
	}
	
	// Prepare IBC client creation request
	clientRequest := map[string]interface{}{
		"client_type":    "07-tendermint", // Default for Cosmos chains
		"chain_id":       sourceNetwork.Config.NetworkParams["chain_id"],
		"target_chain_id": targetNetwork.Config.NetworkParams["chain_id"],
		"unbonding_period": 1814400, // 3 weeks in seconds (typical value)
		"trusting_period": 604800,   // 1 week in seconds
		"max_clock_drift": 10,       // 10 seconds
	}
	
	// Get network endpoint for the BaaS API
	baasEndpoint := sourceNetwork.ActiveEndpoint
	if len(sourceNetwork.Config.NodeEndpoints) > 0 {
		// Use the first node endpoint as a fallback
		baasEndpoint = sourceNetwork.Config.NodeEndpoints[0]
	}
	
	// Construct URL
	url := fmt.Sprintf("%s/ibc/clients", baasEndpoint)
	
	// Convert to JSON
	jsonData, err := json.Marshal(clientRequest)
	if err != nil {
		return "", err
	}
	
	// Send request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	
	// Try to get API key for the network
	networkConfig, err := s.Config.GetNetworkConfig(networkID)
	if err == nil && networkConfig.ApiKeys != nil {
		if apiKey, ok := networkConfig.ApiKeys["baas"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create IBC client: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	// Extract client ID
	clientID, ok := result["client_id"].(string)
	if !ok {
		return "", errors.New("client ID not found in response")
	}
	
	return clientID, nil
}

// CreateIBCConnection creates an IBC connection between two chains
func (s *BaaSService) CreateIBCConnection(sourceNetworkID, targetNetworkID, sourceClientID, targetClientID string) (string, error) {
	// Get source network for endpoint
	sourceNetwork, exists := s.Networks[sourceNetworkID]
	if !exists {
		return "", fmt.Errorf("source network %s not configured", sourceNetworkID)
	}
	
	// Get network endpoint for the BaaS API
	baasEndpoint := sourceNetwork.ActiveEndpoint
	if len(sourceNetwork.Config.NodeEndpoints) > 0 {
		// Use the first node endpoint as a fallback
		baasEndpoint = sourceNetwork.Config.NodeEndpoints[0]
	}
	
	// Prepare IBC connection creation request
	connectionRequest := map[string]interface{}{
		"source_chain_id":  sourceNetworkID,
		"target_chain_id":  targetNetworkID,
		"source_client_id": sourceClientID,
		"target_client_id": targetClientID,
	}
	
	// Construct URL
	url := fmt.Sprintf("%s/ibc/connections", baasEndpoint)
	
	// Convert to JSON
	jsonData, err := json.Marshal(connectionRequest)
	if err != nil {
		return "", err
	}
	
	// Send request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	
	// Try to get API key for the network
	networkConfig, err := s.Config.GetNetworkConfig(sourceNetworkID)
	if err == nil && networkConfig.ApiKeys != nil {
		if apiKey, ok := networkConfig.ApiKeys["baas"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create IBC connection: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	// Extract connection ID
	connectionID, ok := result["connection_id"].(string)
	if !ok {
		return "", errors.New("connection ID not found in response")
	}
	
	return connectionID, nil
}

// CreateIBCChannel creates an IBC channel between two chains
func (s *BaaSService) CreateIBCChannel(
	sourceNetworkID string,
	targetNetworkID string,
	connectionID string,
	sourcePort string,
	targetPort string,
	version string,
	ordering string,
) (string, error) {
	// Get source network for endpoint
	sourceNetwork, exists := s.Networks[sourceNetworkID]
	if !exists {
		return "", fmt.Errorf("source network %s not configured", sourceNetworkID)
	}
	
	// Get network endpoint for the BaaS API
	baasEndpoint := sourceNetwork.ActiveEndpoint
	if len(sourceNetwork.Config.NodeEndpoints) > 0 {
		// Use the first node endpoint as a fallback
		baasEndpoint = sourceNetwork.Config.NodeEndpoints[0]
	}
	
	// Prepare IBC channel creation request
	channelRequest := map[string]interface{}{
		"source_chain_id": sourceNetworkID,
		"target_chain_id": targetNetworkID,
		"connection_id":   connectionID,
		"source_port":     sourcePort,
		"target_port":     targetPort,
		"version":         version,
		"ordering":        ordering, // "ORDERED" or "UNORDERED"
	}
	
	// Construct URL
	url := fmt.Sprintf("%s/ibc/channels", baasEndpoint)
	
	// Convert to JSON
	jsonData, err := json.Marshal(channelRequest)
	if err != nil {
		return "", err
	}
	
	// Send request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	
	// Try to get API key for the network
	networkConfig, err := s.Config.GetNetworkConfig(sourceNetworkID)
	if err == nil && networkConfig.ApiKeys != nil {
		if apiKey, ok := networkConfig.ApiKeys["baas"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create IBC channel: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	// Extract channel ID
	channelID, ok := result["channel_id"].(string)
	if !ok {
		return "", errors.New("channel ID not found in response")
	}
	
	return channelID, nil
}

// SendIBCPacket sends an IBC packet through the specified channel
func (s *BaaSService) SendIBCPacket(
	networkID string,
	channelID string,
	portID string,
	data map[string]interface{},
	timeoutHeight int64,
	timeoutTimestamp int64,
) (string, error) {
	// Get network configuration
	network, exists := s.Networks[networkID]
	if !exists {
		return "", fmt.Errorf("network %s not configured", networkID)
	}

	// Prepare IBC packet request
	packetRequest := map[string]interface{}{
		"source_chain_id": networkID,
		"source_channel": channelID,
		"source_port":    portID,
		"data":           data,
		"timeout_height": map[string]interface{}{
			"revision_number": 0,
			"revision_height": timeoutHeight,
		},
		"timeout_timestamp": timeoutTimestamp,
	}
	
	// Get network endpoint for the BaaS API
	// For a real service, this would be a central BaaS API endpoint
	baasEndpoint := network.ActiveEndpoint
	if len(network.Config.NodeEndpoints) > 0 {
		// Use the first node endpoint as a fallback
		baasEndpoint = network.Config.NodeEndpoints[0]
	}
	
	// Construct URL
	url := fmt.Sprintf("%s/ibc/packets", baasEndpoint)
	
	// Convert to JSON
	jsonData, err := json.Marshal(packetRequest)
	if err != nil {
		return "", err
	}
	
	// Send request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	
	// Try to get API key for the network
	networkConfig, err := s.Config.GetNetworkConfig(networkID)
	if err == nil && networkConfig.ApiKeys != nil {
		if apiKey, ok := networkConfig.ApiKeys["baas"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("failed to send IBC packet: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	// Extract transaction hash
	txHash, ok := result["tx_hash"].(string)
	if !ok {
		return "", errors.New("transaction hash not found in response")
	}
	
	return txHash, nil
}

// CreatePolkadotXCMConnection creates an XCM connection for Polkadot chains
func (s *BaaSService) CreatePolkadotXCMConnection(sourceNetworkID, targetNetworkID string) (string, error) {
	// Check if networks exist
	sourceNetwork, exists := s.Networks[sourceNetworkID]
	if (!exists) {
		return "", fmt.Errorf("source network %s not configured", sourceNetworkID)
	}
	
	targetNetwork, exists := s.Networks[targetNetworkID]
	if (!exists) {
		return "", fmt.Errorf("target network %s not configured", targetNetworkID)
	}
	
	// Check if the networks support XCM
	if sourceNetwork.Config.ChainType != "substrate" && sourceNetwork.Config.ChainType != "polkadot" || !sourceNetwork.Config.XCMEnabled {
		return "", fmt.Errorf("source network %s does not support XCM", sourceNetworkID)
	}
	
	if targetNetwork.Config.ChainType != "substrate" && targetNetwork.Config.ChainType != "polkadot" || !targetNetwork.Config.XCMEnabled {
		return "", fmt.Errorf("target network %s does not support XCM", targetNetworkID)
	}
	
	// Prepare XCM connection request
	xcmRequest := map[string]interface{}{
		"source_network_id": sourceNetworkID,
		"target_network_id": targetNetworkID,
		"source_parachain_id": sourceNetwork.Config.NetworkParams["parachain_id"],
		"target_parachain_id": targetNetwork.Config.NetworkParams["parachain_id"],
		"relay_chain":       sourceNetwork.Config.NetworkParams["relay_chain"],
	}
	
	// Construct URL
	url := fmt.Sprintf("%s/xcm/connections", sourceNetwork.Config.NodeEndpoints[0])
	
	// Convert to JSON
	jsonData, err := json.Marshal(xcmRequest)
	if err != nil {
		return "", err
	}
	
	// Send request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	
	// Try to get API key for the network
	networkConfig, err := s.Config.GetNetworkConfig(sourceNetworkID)
	if err == nil && networkConfig.ApiKeys != nil {
		if apiKey, ok := networkConfig.ApiKeys["baas"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create XCM connection: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	// Extract connection ID
	connectionID, ok := result["connection_id"].(string)
	if !ok {
		return "", errors.New("connection ID not found in response")
	}
	
	return connectionID, nil
}

// SendXCMMessage sends an XCM message through a polkadot connection
func (s *BaaSService) SendXCMMessage(
	sourceNetworkID string,
	targetNetworkID string,
	connectionID string,
	messageType string,
	payload map[string]interface{},
) (string, error) {
	// Prepare XCM message request
	xcmRequest := map[string]interface{}{
		"source_network_id": sourceNetworkID,
		"target_network_id": targetNetworkID,
		"connection_id":     connectionID,
		"message_type":      messageType,
		"payload":           payload,
	}
	
	// Construct URL
	url := fmt.Sprintf("%s/xcm/messages", s.Networks[sourceNetworkID].Config.NodeEndpoints[0])
	
	// Convert to JSON
	jsonData, err := json.Marshal(xcmRequest)
	if err != nil {
		return "", err
	}
	
	// Send request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	
	// Try to get API key for the network
	networkConfig, err := s.Config.GetNetworkConfig(sourceNetworkID)
	if err == nil && networkConfig.ApiKeys != nil {
		if apiKey, ok := networkConfig.ApiKeys["baas"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("failed to send XCM message: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	// Extract transaction hash
	txHash, ok := result["tx_hash"].(string)
	if !ok {
		return "", errors.New("transaction hash not found in response")
	}
	
	return txHash, nil
}

// GetNetworkStatus gets the status of a network
func (s *BaaSService) GetNetworkStatus(networkID string) (map[string]interface{}, error) {
	network, exists := s.Networks[networkID]
	if !exists {
		return nil, fmt.Errorf("network %s not configured", networkID)
	}
	
	// Check if we need to refresh status
	// Refresh if last sync was more than 1 minute ago
	if time.Since(network.LastSync) > time.Minute {
		nodeInfo, blockHeight, err := s.getNodeStatus(network)
		if err != nil {
			// Don't update if error
			return map[string]interface{}{
				"network_id":       networkID,
				"chain_type":       network.Config.ChainType,
				"connection_state": network.ConnectionState,
				"last_sync":        network.LastSync,
				"block_height":     network.BlockHeight,
				"error":            err.Error(),
			}, nil
		}
		
		// Update network status
		network.ConnectionState = "connected"
		network.LastSync = time.Now()
		network.NodeInfo = nodeInfo
		network.BlockHeight = blockHeight
	}
	
	// Return status
	return map[string]interface{}{
		"network_id":       networkID,
		"chain_type":       network.Config.ChainType,
		"connection_state": network.ConnectionState,
		"last_sync":        network.LastSync.Format(time.RFC3339),
		"block_height":     network.BlockHeight,
		"node_info":        network.NodeInfo,
	}, nil
}

// VerifyTransaction verifies a transaction on a blockchain network
func (s *BaaSService) VerifyTransaction(networkID, txHash string) (bool, map[string]interface{}, error) {
	network, exists := s.Networks[networkID]
	if (!exists) {
		return false, nil, fmt.Errorf("network %s not configured", networkID)
	}
	
	// Construct URL based on chain type
	var url string
	switch network.Config.ChainType {
	case "cosmos":
		url = fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", network.ActiveEndpoint, txHash)
	case "substrate", "polkadot":
		// For Substrate/Polkadot chains, we need to use a different approach
		return s.verifySubstrateTransaction(network, txHash)
	default:
		url = fmt.Sprintf("%s/tx/%s", network.ActiveEndpoint, txHash)
	}
	
	// Send request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, nil, err
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	
	// Try to get API key for the network
	networkConfig, err := s.Config.GetNetworkConfig(networkID)
	if err == nil && networkConfig.ApiKeys != nil {
		if apiKey, ok := networkConfig.ApiKeys["baas"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return false, nil, err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return false, nil, fmt.Errorf("failed to verify transaction: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, nil, err
	}
	
	// Extract verification result based on chain type
	var isVerified bool
	
	switch network.Config.ChainType {
	case "cosmos":
		txResponse, ok := result["tx_response"].(map[string]interface{})
		if !ok {
			return false, nil, errors.New("tx_response not found in response")
		}
		
		// Check if transaction was successful
		code, ok := txResponse["code"].(float64)
		if !ok {
			return false, nil, errors.New("code not found in tx_response")
		}
		
		isVerified = code == 0
	default:
		// For other chains, assume if we got a response, it's verified
		isVerified = true
	}
	
	return isVerified, result, nil
}

// verifySubstrateTransaction verifies a transaction on a Substrate/Polkadot chain
func (s *BaaSService) verifySubstrateTransaction(network *BaaSNetwork, txHash string) (bool, map[string]interface{}, error) {
	// Construct request
	payload := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "chain_getBlock",
		"params":  []interface{}{txHash},
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false, nil, err
	}
	
	// Send request
	resp, err := http.Post(
		network.Config.RPCEndpoint,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return false, nil, err
	}
	defer resp.Body.Close()
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, nil, err
	}
	
	// Check for error
	if _, ok := result["error"]; ok {
		return false, result, nil
	}
	
	// If we got a result, the transaction is verified
	return true, result, nil
}
// GetAvailableNetworks returns the list of available networks with their status
func (s *BaaSService) GetAvailableNetworks() []map[string]interface{} {
	networks := make([]map[string]interface{}, 0, len(s.Networks))
	
	for id, network := range s.Networks {
		networks = append(networks, map[string]interface{}{
			"network_id":       id,
			"chain_type":       network.Config.ChainType,
			"connection_state": network.ConnectionState,
			"is_mainnet":       network.Config.IsMainnet,
			"ibc_enabled":      network.Config.IBCEnabled,
			"xcm_enabled":      network.Config.XCMEnabled,
			"explorer_url":     network.Config.ExplorerURL,
		})
	}
	
	return networks
}

// ReceiveIBCPacket handles receiving an IBC packet from another chain
func (s *BaaSService) ReceiveIBCPacket(
	networkID string,
	sourceChainID string,
	sourceChannel string,
	destinationChannel string,
	packet map[string]interface{},
	proof string,
	proofHeight map[string]interface{},
) (string, error) {
	// Prepare receive packet request
	receiveRequest := map[string]interface{}{
		"source_chain_id":     sourceChainID,
		"destination_chain_id": networkID,
		"source_channel":      sourceChannel,
		"destination_channel": destinationChannel,
		"packet":              packet,
		"proof":               proof,
		"proof_height":        proofHeight,
	}
	
	// Construct URL
	url := fmt.Sprintf("%s/ibc/packets/receive", s.Networks[networkID].Config.NodeEndpoints[0])
	
	// Convert to JSON
	jsonData, err := json.Marshal(receiveRequest)
	if err != nil {
		return "", err
	}
	
	// Send request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	
	// Try to get API key for the network
	networkConfig, err := s.Config.GetNetworkConfig(networkID)
	if err == nil && networkConfig.ApiKeys != nil {
		if apiKey, ok := networkConfig.ApiKeys["baas"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("failed to receive IBC packet: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	// Extract transaction hash
	txHash, ok := result["tx_hash"].(string)
	if !ok {
		return "", errors.New("transaction hash not found in response")
	}
	
	return txHash, nil
}

// QueryIBCChannels queries all IBC channels for a network
func (s *BaaSService) QueryIBCChannels(networkID string) ([]map[string]interface{}, error) {
	network, exists := s.Networks[networkID]
	if !exists {
		return nil, fmt.Errorf("network %s not configured", networkID)
	}
	
	// Check if the network supports IBC
	if network.Config.ChainType != "cosmos" || !network.Config.IBCEnabled {
		return nil, fmt.Errorf("network %s does not support IBC", networkID)
	}
	
	// Construct URL
	url := fmt.Sprintf("%s/ibc/core/channel/v1/channels", network.Config.NodeEndpoints[0])
	
	// Send request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	// Add headers
	req.Header.Set("X-API-Key", s.Config.APIKey)
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to query IBC channels: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	// Extract channels
	channelsData, ok := result["channels"].([]interface{})
	if !ok {
		return nil, errors.New("channels not found in response")
	}
	
	// Convert to standard format
	channels := make([]map[string]interface{}, 0, len(channelsData))
	for _, channelData := range channelsData {
		channel, ok := channelData.(map[string]interface{})
		if !ok {
			continue
		}
		channels = append(channels, channel)
	}
	
	return channels, nil
}

// QueryIBCConnections queries all IBC connections for a network
func (s *BaaSService) QueryIBCConnections(networkID string) ([]map[string]interface{}, error) {
	network, exists := s.Networks[networkID]
	if !exists {
		return nil, fmt.Errorf("network %s not configured", networkID)
	}
	
	// Check if the network supports IBC
	if network.Config.ChainType != "cosmos" || !network.Config.IBCEnabled {
		return nil, fmt.Errorf("network %s does not support IBC", networkID)
	}
	
	// Construct URL
	url := fmt.Sprintf("%s/ibc/core/connection/v1/connections", network.Config.NodeEndpoints[0])
	
	// Send request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	// Add headers
	req.Header.Set("X-API-Key", s.Config.APIKey)
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to query IBC connections: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	// Define `connections` variable
	connections := []map[string]interface{}{}
	
	return connections, nil
}

// GetIBCDenomTrace retrieves the trace information for an IBC token
func (s *BaaSService) GetIBCDenomTrace(networkID, denom string) (map[string]interface{}, error) {
	network, exists := s.Networks[networkID]
	if !exists {
		return nil, fmt.Errorf("network %s not configured", networkID)
	}
	
	// Check if the network supports IBC
	if network.Config.ChainType != "cosmos" || !network.Config.IBCEnabled {
		return nil, fmt.Errorf("network %s does not support IBC", networkID)
	}
	
	// For IBC denoms, extract the hash
	var hash string
	if strings.HasPrefix(denom, "ibc/") {
		hash = strings.TrimPrefix(denom, "ibc/")
	} else {
		// For non-IBC denoms, return basic info
		return map[string]interface{}{
			"denom": denom,
			"path": "",
			"base_denom": denom,
			"is_ibc_token": false,
		}, nil
	}
	
	// Construct URL
	url := fmt.Sprintf("%s/ibc/apps/transfer/v1/denom_traces/%s", network.Config.NodeEndpoints[0], hash)
	
	// Send request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	// Add headers
	req.Header.Set("X-API-Key", s.Config.APIKey)
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get IBC denom trace: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	// Extract denom trace
	denomTrace, ok := result["denom_trace"].(map[string]interface{})
	if !ok {
		return nil, errors.New("denom_trace not found in response")
	}
	
	// Add additional info
	denomTrace["is_ibc_token"] = true
	
	return denomTrace, nil
}

// CreateInterChainAccount creates an interchain account (ICA) for cross-chain interactions
func (s *BaaSService) CreateInterChainAccount(
	networkID string,
	targetNetworkID string,
	connectionID string,
	owner string,
) (string, error) {
	// Prepare ICA creation request
	icaRequest := map[string]interface{}{
		"source_chain_id": networkID,
		"target_chain_id": targetNetworkID,
		"connection_id":   connectionID,
		"owner":           owner,
	}
	
	// Construct URL
	url := fmt.Sprintf("%s/ibc/interchain_accounts", s.Networks[networkID].Config.NodeEndpoints[0])
	
	// Convert to JSON
	jsonData, err := json.Marshal(icaRequest)
	if err != nil {
		return "", err
	}
	
	// Send request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	
	// Try to get API key for the network
	networkConfig, err := s.Config.GetNetworkConfig(networkID)
	if err == nil && networkConfig.ApiKeys != nil {
		if apiKey, ok := networkConfig.ApiKeys["baas"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create interchain account: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	// Extract account address
	accountAddress, ok := result["account_address"].(string)
	if !ok {
		return "", errors.New("account address not found in response")
	}
	
	return accountAddress, nil
}

// SendInterChainAccountTx sends a transaction from an interchain account
func (s *BaaSService) SendInterChainAccountTx(
	networkID string,
	targetNetworkID string,
	connectionID string,
	owner string,
	msgs []map[string]interface{},
	memo string,
) (string, error) {
	// Prepare ICA tx request
	icaTxRequest := map[string]interface{}{
		"source_chain_id": networkID,
		"target_chain_id": targetNetworkID,
		"connection_id":   connectionID,
		"owner":           owner,
		"messages":        msgs,
		"memo":            memo,
	}
	
	// Construct URL
	url := fmt.Sprintf("%s/ibc/interchain_accounts/tx", s.Networks[networkID].Config.NodeEndpoints[0])
	
	// Convert to JSON
	jsonData, err := json.Marshal(icaTxRequest)
	if err != nil {
		return "", err
	}
	
	// Send request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	
	// Try to get API key for the network
	networkConfig, err := s.Config.GetNetworkConfig(networkID)
	if err == nil && networkConfig.ApiKeys != nil {
		if apiKey, ok := networkConfig.ApiKeys["baas"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("failed to send interchain account transaction: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	// Extract transaction hash
	txHash, ok := result["tx_hash"].(string)
	if (!ok) {
		return "", errors.New("transaction hash not found in response")
	}
	
	return txHash, nil
}

// CreateCustomChain creates a custom blockchain in the BaaS platform
func (s *BaaSService) CreateCustomChain(
	name string,
	chainType string,
	consensus string,
	networkParams map[string]string,
	validators []map[string]interface{},
) (string, error) {
	// Prepare chain creation request
	chainRequest := map[string]interface{}{
		"name":           name,
		"chain_type":     chainType,
		"consensus":      consensus,
		"network_params": networkParams,
		"validators":     validators,
	}
	
	// Construct URL
	url := fmt.Sprintf("%s/chains", s.Networks[name].Config.NodeEndpoints[0])
	
	// Convert to JSON
	jsonData, err := json.Marshal(chainRequest)
	if err != nil {
		return "", err
	}
	
	// Send request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	
	// Try to get API key for the network
	networkConfig, err := s.Config.GetNetworkConfig(name)
	if err == nil && networkConfig.ApiKeys != nil {
		if apiKey, ok := networkConfig.ApiKeys["baas"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create custom chain: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	// Extract chain ID
	chainID, ok := result["chain_id"].(string)
	if !ok {
		return "", errors.New("chain ID not found in response")
	}
	
	return chainID, nil
}

// DeploySmartContract deploys a smart contract to a blockchain
func (s *BaaSService) DeploySmartContract(
	networkID string,
	contractType string,
	contractName string,
	contractCode string,
	initArgs map[string]interface{},
) (string, error) {
	// Prepare contract deployment request
	deployRequest := map[string]interface{}{
		"network_id":     networkID,
		"contract_type":  contractType,
		"contract_name":  contractName,
		"contract_code":  contractCode,
		"init_args":      initArgs,
	}
	
	// Construct URL
	url := fmt.Sprintf("%s/contracts", s.Networks[networkID].Config.NodeEndpoints[0])
	
	// Convert to JSON
	jsonData, err := json.Marshal(deployRequest)
	if err != nil {
		return "", err
	}
	
	// Send request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	
	// Try to get API key for the network
	networkConfig, err := s.Config.GetNetworkConfig(networkID)
	if err == nil && networkConfig.ApiKeys != nil {
		if apiKey, ok := networkConfig.ApiKeys["baas"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to deploy smart contract: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	// Extract contract address
	contractAddress, ok := result["contract_address"].(string)
	if !ok {
		return "", errors.New("contract address not found in response")
	}
	
	return contractAddress, nil
}

// QueryContractState queries the state of a smart contract
func (s *BaaSService) QueryContractState(
	networkID string,
	contractAddress string,
	queryData map[string]interface{},
) (map[string]interface{}, error) {
	network, exists := s.Networks[networkID]
	if !exists {
		return nil, fmt.Errorf("network %s not configured", networkID)
	}
	
	// Determine URL based on chain type
	var url string
	switch network.Config.ChainType {
	case "cosmos":
		url = fmt.Sprintf("%s/cosmwasm/wasm/v1/contract/%s/smart/%s", 
			network.ActiveEndpoint, contractAddress, encodeQueryToBase64(queryData))
	case "substrate", "polkadot":
		// For substrate chains, use RPC endpoint
		return s.querySubstrateContractState(network, contractAddress, queryData)
	default:
		url = fmt.Sprintf("%s/contracts/%s/query", network.ActiveEndpoint, contractAddress)
	}
	
	// Send request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	// Add headers
	req.Header.Set("X-API-Key", s.Config.APIKey)
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to query contract state: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result, nil
}

// encodeQueryToBase64 encodes a query object to base64 (helper for CosmWasm queries)
func encodeQueryToBase64(queryData map[string]interface{}) string {
	jsonData, _ := json.Marshal(queryData)
	return base64.StdEncoding.EncodeToString(jsonData)
}

// querySubstrateContractState queries a Substrate contract state
func (s *BaaSService) querySubstrateContractState(
	network *BaaSNetwork,
	contractAddress string,
	queryData map[string]interface{},
) (map[string]interface{}, error) {
	// Prepare RPC request
	rpcRequest := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "contracts_call",
		"params": []interface{}{
			map[string]interface{}{
				"dest":      contractAddress,
				"value":     "0",
				"gas_limit": -1,
				"input_data": queryData,
			},
		},
	}
	
	// Convert to JSON
	jsonData, err := json.Marshal(rpcRequest)
	if err != nil {
		return nil, err
	}
	
	// Send request
	resp, err := http.Post(
		network.Config.RPCEndpoint,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	// Check for error
	if errData, ok := result["error"]; ok {
		errMsg, _ := json.Marshal(errData)
		return nil, fmt.Errorf("RPC error: %s", string(errMsg))
	}
	
	// Extract result
	if rpcResult, ok := result["result"].(map[string]interface{}); ok {
		return rpcResult, nil
	}
	
	return nil, errors.New("invalid RPC response format")
}

// CreateCrossChainBridge creates a cross-chain bridge between two networks
func (s *BaaSService) CreateCrossChainBridge(
	sourceNetworkID string,
	targetNetworkID string,
	bridgeType string,
	bridgeConfig map[string]interface{},
) (string, error) {
	// Validate networks
	_, sourceExists := s.Networks[sourceNetworkID]
	if !sourceExists {
		return "", fmt.Errorf("source network %s not configured", sourceNetworkID)
	}
	
	_, targetExists := s.Networks[targetNetworkID]
	if !targetExists {
		return "", fmt.Errorf("target network %s not configured", targetNetworkID)
	}
	
	// Prepare bridge creation request
	bridgeRequest := map[string]interface{}{
		"source_network_id": sourceNetworkID,
		"target_network_id": targetNetworkID,
		"bridge_type":       bridgeType,
		"bridge_config":     bridgeConfig,
	}
	
	// Construct URL
	url := fmt.Sprintf("%s/bridges", s.Networks[sourceNetworkID].Config.NodeEndpoints[0])
	
	// Convert to JSON
	jsonData, err := json.Marshal(bridgeRequest)
	if err != nil {
		return "", err
	}
	
	// Send request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	
	// Try to get API key for the network
	networkConfig, err := s.Config.GetNetworkConfig(sourceNetworkID)
	if err == nil && networkConfig.ApiKeys != nil {
		if apiKey, ok := networkConfig.ApiKeys["baas"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create cross-chain bridge: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	// Extract bridge ID
	bridgeID, ok := result["bridge_id"].(string)
	if !ok {
		return "", errors.New("bridge ID not found in response")
	}
	
	return bridgeID, nil
}

// TransferAssetAcrossChains transfers an asset from one chain to another using a bridge
func (s *BaaSService) TransferAssetAcrossChains(
	sourceNetworkID string,
	targetNetworkID string,
	bridgeID string,
	assetID string,
	amount string,
	sender string,
	recipient string,
) (string, error) {
	// Prepare transfer request
	transferRequest := map[string]interface{}{
		"source_network_id": sourceNetworkID,
		"target_network_id": targetNetworkID,
		"bridge_id":         bridgeID,
		"asset_id":          assetID,
		"amount":            amount,
		"sender":            sender,
		"recipient":         recipient,
	}
	
	// Construct URL
	url := fmt.Sprintf("%s/bridges/%s/transfer", s.Networks[sourceNetworkID].Config.NodeEndpoints[0], bridgeID)
	
	// Convert to JSON
	jsonData, err := json.Marshal(transferRequest)
	if err != nil {
		return "", err
	}
	
	// Send request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	
	// Try to get API key for the network
	networkConfig, err := s.Config.GetNetworkConfig(sourceNetworkID)
	if err == nil && networkConfig.ApiKeys != nil {
		if apiKey, ok := networkConfig.ApiKeys["baas"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("failed to transfer asset across chains: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	// Extract transaction hash
	txHash, ok := result["tx_hash"].(string)
	if !ok {
		return "", errors.New("transaction hash not found in response")
	}
	
	return txHash, nil
}

// GetBridgeTransactions gets the list of transactions for a cross-chain bridge
func (s *BaaSService) GetBridgeTransactions(
	bridgeID string,
	limit int,
	offset int,
) ([]map[string]interface{}, error) {
	// Construct URL
	url := fmt.Sprintf("%s/bridges/%s/transactions?limit=%d&offset=%d", 
		s.Networks[bridgeID].Config.NodeEndpoints[0], bridgeID, limit, offset)
	
	// Send request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	// Add headers
	req.Header.Set("X-API-Key", s.Config.APIKey)
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get bridge transactions: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	// Extract transactions
	transactions, ok := result["transactions"].([]interface{})
	if !ok {
		return nil, errors.New("transactions not found in response")
	}
	
	// Convert to standard format
	txList := make([]map[string]interface{}, 0, len(transactions))
	for _, txData := range transactions {
		tx, ok := txData.(map[string]interface{})
		if !ok {
			continue
		}
		txList = append(txList, tx)
	}
	
	return txList, nil
}

// GetBridgeById gets details of a specific bridge
func (s *BaaSService) GetBridgeById(bridgeID string) (map[string]interface{}, error) {
	// Construct URL
	url := fmt.Sprintf("%s/bridges/%s", s.Networks[bridgeID].Config.NodeEndpoints[0], bridgeID)
	
	// Send request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	// Add headers
	req.Header.Set("X-API-Key", s.Config.APIKey)
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get bridge details: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result, nil
}

// CallContractMethod invokes a method on a smart contract
func (s *BaaSService) CallContractMethod(
	networkID string,
	contractAddress string,
	methodData map[string]interface{},
) (map[string]interface{}, error) {
	// Validate required fields
	if _, ok := methodData["method"]; !ok {
		return nil, errors.New("method name is required")
	}
	
	// Prepare contract call request
	callRequest := map[string]interface{}{
		"contract_address": contractAddress,
		"method":           methodData["method"],
		"params":           methodData["params"],
	}
	
	// Construct URL for the network node
	network, exists := s.Networks[networkID]
	if !exists || len(network.Config.NodeEndpoints) == 0 {
		return nil, fmt.Errorf("network '%s' not found or has no endpoints", networkID)
	}
	
	url := fmt.Sprintf("%s/contracts/%s/call", network.Config.NodeEndpoints[0], contractAddress)
	
	// Convert to JSON
	jsonData, err := json.Marshal(callRequest)
	if err != nil {
		return nil, err
	}
	
	// Send request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	
	// Try to get API key for the network
	networkConfig, err := s.Config.GetNetworkConfig(networkID)
	if err == nil && networkConfig.ApiKeys != nil {
		if apiKey, ok := networkConfig.ApiKeys["baas"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
	
	// Execute request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to call contract method: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result, nil
}