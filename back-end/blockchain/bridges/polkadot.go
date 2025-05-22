// Package bridges provides implementations of cross-chain bridges
package bridges

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// PolkadotBridge implements cross-chain functionality with Polkadot's XCM protocol
type PolkadotBridge struct {
	RelayEndpoint    string
	RelayChainID     string
	ParachainID      string
	ChainID          string
	APIKey           string
	RegisteredAssets map[string]XCMAssetDetails
	XCMRoutes        map[string]XCMRouteDetails
	LastBlockNumber  uint64
	RococoMode       bool
}

// XCMAssetDetails holds details about an asset that can be transferred via XCM
type XCMAssetDetails struct {
	AssetID          string `json:"asset_id"`
	Name             string `json:"name"`
	Symbol           string `json:"symbol"`
	Decimals         int    `json:"decimals"`
	OriginChain      string `json:"origin_chain"`
	OriginLocation   string `json:"origin_location"`
	MultiLocation    map[string]interface{} `json:"multi_location"`
	MetadataURI      string `json:"metadata_uri,omitempty"`
	AssetProcessor   string `json:"asset_processor,omitempty"`
}

// XCMRouteDetails holds details about an XCM route between chains
type XCMRouteDetails struct {
	SourceChainID      string `json:"source_chain_id"`
	DestinationChainID string `json:"destination_chain_id"`
	RouteType          string `json:"route_type"`
	Hops               []XCMHop `json:"hops,omitempty"`
	Status             string `json:"status"`
	Fee                string `json:"fee,omitempty"`
	FeeAsset           string `json:"fee_asset,omitempty"`
	Teleporter         string `json:"teleporter,omitempty"`
}

// XCMHop represents a single hop in an XCM route
type XCMHop struct {
	ChainID        string `json:"chain_id"`
	ParachainID    string `json:"parachain_id,omitempty"`
	RelayChain     string `json:"relay_chain,omitempty"`
	BridgeContract string `json:"bridge_contract,omitempty"`
}

// XCMMessage represents a cross-chain message in the Polkadot ecosystem
type XCMMessage struct {
	MessageID          string                 `json:"message_id"`
	SourceChainID      string                 `json:"source_chain_id"`
	DestinationChainID string                 `json:"destination_chain_id"`
	MessageType        string                 `json:"message_type"`
	Payload            map[string]interface{} `json:"payload"`
	Timestamp          int64                  `json:"timestamp"`
	Status             string                 `json:"status"`
	Version            string                 `json:"version"`
	Instructions       []map[string]interface{} `json:"instructions,omitempty"`
	Weight             uint64                 `json:"weight"`
	Fee                string                 `json:"fee,omitempty"`
	BatchID            string                 `json:"batch_id,omitempty"`
	ParachainID        string                 `json:"parachain_id,omitempty"`
	RelayChainID       string                 `json:"relay_chain_id,omitempty"`
}

// NewPolkadotBridge creates a new Polkadot bridge instance
func NewPolkadotBridge(relayEndpoint, relayChainID, parachainID, chainID, apiKey string) *PolkadotBridge {
	return &PolkadotBridge{
		RelayEndpoint:    relayEndpoint,
		RelayChainID:     relayChainID,
		ParachainID:      parachainID,
		ChainID:          chainID,
		APIKey:           apiKey,
		RegisteredAssets: make(map[string]XCMAssetDetails),
		XCMRoutes:        make(map[string]XCMRouteDetails),
		RococoMode:       strings.Contains(strings.ToLower(relayChainID), "rococo"),
	}
}

// RegisterXCMAsset registers an asset for XCM transfers
func (b *PolkadotBridge) RegisterXCMAsset(asset XCMAssetDetails) {
	b.RegisteredAssets[asset.AssetID] = asset
}

// AddXCMRoute adds an XCM route between chains
func (b *PolkadotBridge) AddXCMRoute(route XCMRouteDetails) {
	routeID := fmt.Sprintf("%s-%s", route.SourceChainID, route.DestinationChainID)
	b.XCMRoutes[routeID] = route
}

// SendXCMMessage sends an XCM message to another Polkadot-based chain
func (b *PolkadotBridge) SendXCMMessage(destinationChainID, messageType string, payload map[string]interface{}) (string, error) {
	// Create a unique message ID
	payloadJSON, _ := json.Marshal(payload)
	hash := sha256.Sum256(payloadJSON)
	messageID := hex.EncodeToString(hash[:])
	
	// Find the route to the destination chain
	routeID := fmt.Sprintf("%s-%s", b.ChainID, destinationChainID)
	route, exists := b.XCMRoutes[routeID]
	if (!exists) {
		return "", fmt.Errorf("no XCM route found from %s to %s", b.ChainID, destinationChainID)
	}
	
	// Determine the destination parachain ID if it's a Polkadot ecosystem chain
	var destParachainID string
	if len(route.Hops) > 0 && route.RouteType == "ViaRelay" {
		for _, hop := range route.Hops {
			if hop.ChainID == destinationChainID {
				destParachainID = hop.ParachainID
				break
			}
		}
	}
	
	// Create the XCM message
	message := XCMMessage{
		MessageID:          messageID,
		SourceChainID:      b.ChainID,
		DestinationChainID: destinationChainID,
		MessageType:        messageType,
		Payload:            payload,
		Timestamp:          time.Now().UnixNano(),
		Status:             "pending",
		Version:            "V3", // Default to XCM V3
		Weight:             10000000, // Default weight
		ParachainID:        b.ParachainID,
		RelayChainID:       b.RelayChainID,
	}
	
	// If destination is another parachain, include its parachain ID
	if destParachainID != "" {
		message.ParachainID = destParachainID
	}
	
	// Create appropriate instructions based on message type
	instructions := []map[string]interface{}{}
	
	if messageType == "Transfer" {
		// Extract required fields from payload
		recipient, ok := payload["recipient"].(string)
		if !ok {
			return "", errors.New("recipient is required for Transfer messages")
		}
		
		assetID, ok := payload["asset_id"].(string)
		if !ok {
			return "", errors.New("asset_id is required for Transfer messages")
		}
		
		amount, ok := payload["amount"].(string)
		if (!ok) {
			return "", errors.New("amount is required for Transfer messages")
		}
		
		// Get asset details - FIX: Removed unused variable declaration
		assetDetails, exists := b.RegisteredAssets[assetID]
		if !exists {
			return "", fmt.Errorf("asset %s is not registered", assetID)
		}
		
		// Build XCM transfer instructions
		instructions = []map[string]interface{}{
			{
				"ReserveAssetDeposited": []map[string]interface{}{
					{
						"assets": []map[string]interface{}{
							{
								"id": map[string]interface{}{
									"Concrete": assetDetails.MultiLocation,
								},
								"fun": map[string]interface{}{
									"Fungible": amount,
								},
							},
						},
					},
				},
			},
			{
				"ClearOrigin": map[string]interface{}{},
			},
			{
				"BuyExecution": map[string]interface{}{
					"fees": map[string]interface{}{
						"id": map[string]interface{}{
							"Concrete": assetDetails.MultiLocation,
						},
						"fun": map[string]interface{}{
							"Fungible": "1000000", // Hardcoded fee amount for example
						},
					},
					"weightLimit": "Unlimited",
				},
			},
			{
				"DepositAsset": map[string]interface{}{
					"assets": map[string]interface{}{
						"Wild": "All",
					},
					"beneficiary": map[string]interface{}{
						"parents": 0,
						"interior": map[string]interface{}{
							"X1": map[string]interface{}{
								"AccountId32": map[string]interface{}{
									"id": recipient,
									"network": "Any",
								},
							},
						},
					},
				},
			},
		}
	} else if messageType == "Call" {
		// Extract required fields from payload
		callData, ok := payload["call_data"].(string)
		if (!ok) {
			return "", errors.New("call_data is required for Call messages")
		}
		
		// Build XCM call instructions
		instructions = []map[string]interface{}{
			{
				"Transact": map[string]interface{}{
					"originType": "SovereignAccount",
					"requireWeightAtMost": message.Weight,
					"call": map[string]interface{}{
						"encoded": callData,
					},
				},
			},
		}
	}
	
	// Add instructions to the message
	message.Instructions = instructions
	
	// Prepare the XCM message request
	xcmRequest := map[string]interface{}{
		"destination_chain_id": destinationChainID,
		"message_type": messageType,
		"version": message.Version,
		"instructions": instructions,
		"payload": payload,
	}
	
	// For parachain to parachain communication
	if destParachainID != "" && b.ParachainID != "" {
		xcmRequest["source_parachain_id"] = b.ParachainID
		xcmRequest["destination_parachain_id"] = destParachainID
		xcmRequest["relay_chain_id"] = b.RelayChainID
	}
	
	// Convert to JSON
	jsonData, err := json.Marshal(xcmRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal XCM message request: %v", err)
	}
	
	// Determine the endpoint based on route type
	var endpoint string
	if route.RouteType == "ViaRelay" {
		endpoint = b.RelayEndpoint
	} else {
		// Direct routes use the chain's endpoint
		endpoint = fmt.Sprintf("https://%s-api.polkadot.io", strings.ToLower(b.ChainID))
	}
	
	// Create HTTP request
	req, err := http.NewRequest("POST", endpoint+"/xcm/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	if b.APIKey != "" {
		req.Header.Set("X-API-Key", b.APIKey)
	}
	
	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send XCM message: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		var errResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			if errMsg, ok := errResp["error"].(string); ok {
				return "", fmt.Errorf("XCM message send failed: %s", errMsg)
			}
		}
		return "", fmt.Errorf("XCM message send failed with status: %d", resp.StatusCode)
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}
	
	// Extract the transaction hash
	txHash, ok := response["tx_hash"].(string)
	if (!ok) {
		return messageID, nil // Fallback to message ID if hash not available
	}
	
	return txHash, nil
}

// GetXCMMessageStatus gets the status of an XCM message
func (b *PolkadotBridge) GetXCMMessageStatus(txHash string) (string, error) {
	// Create the request URL
	url := fmt.Sprintf("%s/tx/%s", b.RelayEndpoint, txHash)
	
	// Send the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	
	// Add headers
	if b.APIKey != "" {
		req.Header.Set("X-API-Key", b.APIKey)
	}
	
	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get transaction status: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get transaction status: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var txResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&txResult); err != nil {
		return "", fmt.Errorf("failed to decode transaction status: %v", err)
	}
	
	// Extract transaction details
	status, ok := txResult["success"].(bool)
	if (!ok) {
		return "unknown", nil
	}
	
	if status {
		return "success", nil
	} else {
		errorMsg, _ := txResult["error"].(string)
		return "failed", fmt.Errorf("transaction failed: %s", errorMsg)
	}
}

// VerifyXCMMessage verifies an XCM message on the destination chain
func (b *PolkadotBridge) VerifyXCMMessage(sourceChainID, messageID, txHash string) (bool, error) {
	// Create the verification request
	verificationRequest := map[string]interface{}{
		"source_chain_id": sourceChainID,
		"message_id":      messageID,
		"tx_hash":         txHash,
	}
	
	// Convert to JSON
	jsonData, err := json.Marshal(verificationRequest)
	if err != nil {
		return false, fmt.Errorf("failed to marshal verification request: %v", err)
	}
	
	// Create HTTP request
	url := fmt.Sprintf("%s/xcm/verify", b.RelayEndpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to create verification request: %v", err)
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	if b.APIKey != "" {
		req.Header.Set("X-API-Key", b.APIKey)
	}
	
	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to verify XCM message: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to verify XCM message: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("failed to decode verification response: %v", err)
	}
	
	// Check the verification result
	verified, ok := response["verified"].(bool)
	if (!ok) {
		return false, errors.New("verification response did not contain verified status")
	}
	
	return verified, nil
}

// QueryXCMRoutes queries available XCM routes for this chain
func (b *PolkadotBridge) QueryXCMRoutes() ([]XCMRouteDetails, error) {
	// Create the request URL
	url := fmt.Sprintf("%s/xcm/routes?parachain_id=%s", b.RelayEndpoint, b.ParachainID)
	
	// Send the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// Add headers
	if b.APIKey != "" {
		req.Header.Set("X-API-Key", b.APIKey)
	}
	
	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query XCM routes: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to query XCM routes: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode routes response: %v", err)
	}
	
	// Extract routes
	routesData, ok := response["routes"].([]interface{})
	if (!ok) {
		return nil, errors.New("response does not contain routes")
	}
	
	// Process the routes
	routes := []XCMRouteDetails{}
	for _, routeData := range routesData {
		routeMap, ok := routeData.(map[string]interface{})
		if (!ok) {
			continue
		}
		
		// Extract route details
		sourceChainID, _ := routeMap["source_chain_id"].(string)
		destinationChainID, _ := routeMap["destination_chain_id"].(string)
		routeType, _ := routeMap["route_type"].(string)
		status, _ := routeMap["status"].(string)
		fee, _ := routeMap["fee"].(string)
		feeAsset, _ := routeMap["fee_asset"].(string)
		
		// Extract hops
		hops := []XCMHop{}
		hopsData, ok := routeMap["hops"].([]interface{})
		if ok {
			for _, hopData := range hopsData {
				hopMap, ok := hopData.(map[string]interface{})
				if (!ok) {
					continue
				}
				
				// Extract hop details
				chainID, _ := hopMap["chain_id"].(string)
				parachainID, _ := hopMap["parachain_id"].(string)
				relayChain, _ := hopMap["relay_chain"].(string)
				bridgeContract, _ := hopMap["bridge_contract"].(string)
				
				// Create and add the hop
				hop := XCMHop{
					ChainID:        chainID,
					ParachainID:    parachainID,
					RelayChain:     relayChain,
					BridgeContract: bridgeContract,
				}
				
				hops = append(hops, hop)
			}
		}
		
		// Create and add the route
		route := XCMRouteDetails{
			SourceChainID:      sourceChainID,
			DestinationChainID: destinationChainID,
			RouteType:          routeType,
			Hops:               hops,
			Status:             status,
			Fee:                fee,
			FeeAsset:           feeAsset,
		}
		
		routes = append(routes, route)
		
		// Also update our internal routes map
		routeID := fmt.Sprintf("%s-%s", sourceChainID, destinationChainID)
		b.XCMRoutes[routeID] = route
	}
	
	return routes, nil
}

// QueryXCMAssets queries available XCM assets for this chain
func (b *PolkadotBridge) QueryXCMAssets() ([]XCMAssetDetails, error) {
	// Create the request URL
	url := fmt.Sprintf("%s/xcm/assets?parachain_id=%s", b.RelayEndpoint, b.ParachainID)
	
	// Send the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// Add headers
	if b.APIKey != "" {
		req.Header.Set("X-API-Key", b.APIKey)
	}
	
	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query XCM assets: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to query XCM assets: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode assets response: %v", err)
	}
	
	// Extract assets
	assetsData, ok := response["assets"].([]interface{})
	if (!ok) {
		return nil, errors.New("response does not contain assets")
	}
	
	// Process the assets
	assets := []XCMAssetDetails{}
	for _, assetData := range assetsData {
		assetMap, ok := assetData.(map[string]interface{})
		if (!ok) {
			continue
		}
		
		// Extract asset details
		assetID, _ := assetMap["asset_id"].(string)
		name, _ := assetMap["name"].(string)
		symbol, _ := assetMap["symbol"].(string)
		decimals, _ := assetMap["decimals"].(float64)
		originChain, _ := assetMap["origin_chain"].(string)
		originLocation, _ := assetMap["origin_location"].(string)
		multiLocation, _ := assetMap["multi_location"].(map[string]interface{})
		metadataURI, _ := assetMap["metadata_uri"].(string)
		assetProcessor, _ := assetMap["asset_processor"].(string)
		
		// Create and add the asset
		asset := XCMAssetDetails{
			AssetID:         assetID,
			Name:            name,
			Symbol:          symbol,
			Decimals:        int(decimals),
			OriginChain:     originChain,
			OriginLocation:  originLocation,
			MultiLocation:   multiLocation,
			MetadataURI:     metadataURI,
			AssetProcessor:  assetProcessor,
		}
		
		assets = append(assets, asset)
		
		// Also update our internal assets map
		b.RegisteredAssets[assetID] = asset
	}
	
	return assets, nil
}

// GetLastBlockNumber gets the latest block number from the chain
func (b *PolkadotBridge) GetLastBlockNumber() (uint64, error) {
	// Create the request URL
	url := fmt.Sprintf("%s/blocks/head", b.RelayEndpoint)
	
	// Send the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %v", err)
	}
	
	// Add headers
	if b.APIKey != "" {
		req.Header.Set("X-API-Key", b.APIKey)
	}
	
	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get latest block: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var blockResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&blockResponse); err != nil {
		return 0, fmt.Errorf("failed to decode block response: %v", err)
	}
	
	// Extract block number
	blockNumber, ok := blockResponse["number"].(float64)
	if (!ok) {
		return 0, errors.New("response does not contain block number")
	}
	
	// Update cached block number
	b.LastBlockNumber = uint64(blockNumber)
	
	return uint64(blockNumber), nil
}

// TransferXCMAsset transfers an asset via XCM from this chain to another chain
func (b *PolkadotBridge) TransferXCMAsset(recipient, assetID string, amount string, destinationChainID string) (string, error) {
	// Check if the asset is registered
	_, exists := b.RegisteredAssets[assetID]
	if (!exists) {
		return "", fmt.Errorf("asset %s is not registered", assetID)
	}
	
	// Create the payload for an XCM transfer
	payload := map[string]interface{}{
		"recipient": recipient,
		"asset_id":  assetID,
		"amount":    amount,
	}
	
	// Send as an XCM message
	return b.SendXCMMessage(destinationChainID, "Transfer", payload)
}

// TraceXCMAsset traces an XCM asset's origin and path
func (b *PolkadotBridge) TraceXCMAsset(assetID string) (map[string]interface{}, error) {
	// Check if we have this asset registered locally
	asset, exists := b.RegisteredAssets[assetID]
	if (exists) {
		return map[string]interface{}{
			"asset_id":        asset.AssetID,
			"name":            asset.Name,
			"symbol":          asset.Symbol,
			"decimals":        asset.Decimals,
			"origin_chain":    asset.OriginChain,
			"origin_location": asset.OriginLocation,
			"multi_location":  asset.MultiLocation,
			"metadata_uri":    asset.MetadataURI,
			"asset_processor": asset.AssetProcessor,
		}, nil
	}
	
	// If not locally registered, query the chain
	url := fmt.Sprintf("%s/xcm/assets/%s?parachain_id=%s", b.RelayEndpoint, assetID, b.ParachainID)
	
	// Send the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// Add headers
	if b.APIKey != "" {
		req.Header.Set("X-API-Key", b.APIKey)
	}
	
	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to trace XCM asset: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to trace XCM asset: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode asset trace response: %v", err)
	}
	
	// Extract asset details
	assetDetails, ok := response["asset"].(map[string]interface{})
	if (!ok) {
		return nil, errors.New("response does not contain asset details")
	}
	
	// Extract and convert properties
	name, _ := assetDetails["name"].(string)
	symbol, _ := assetDetails["symbol"].(string)
	decimals, _ := assetDetails["decimals"].(float64)
	originChain, _ := assetDetails["origin_chain"].(string)
	originLocation, _ := assetDetails["origin_location"].(string)
	multiLocation, _ := assetDetails["multi_location"].(map[string]interface{})
	metadataURI, _ := assetDetails["metadata_uri"].(string)
	assetProcessor, _ := assetDetails["asset_processor"].(string)
	
	// Register this asset for future reference
	b.RegisteredAssets[assetID] = XCMAssetDetails{
		AssetID:        assetID,
		Name:           name,
		Symbol:         symbol,
		Decimals:       int(decimals),
		OriginChain:    originChain,
		OriginLocation: originLocation,
		MultiLocation:  multiLocation,
		MetadataURI:    metadataURI,
		AssetProcessor: assetProcessor,
	}
	
	return assetDetails, nil
}

// ReceiveXCMMessage handles an XCM message received from another chain
func (b *PolkadotBridge) ReceiveXCMMessage(sourceChainID string, message map[string]interface{}) (string, error) {
	// Create the receive message request
	receiveRequest := map[string]interface{}{
		"source_chain_id": sourceChainID,
		"message":         message,
	}
	
	// Convert to JSON
	jsonData, err := json.Marshal(receiveRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal receive message request: %v", err)
	}
	
	// Create HTTP request
	url := fmt.Sprintf("%s/xcm/receive", b.RelayEndpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create receive request: %v", err)
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	if b.APIKey != "" {
		req.Header.Set("X-API-Key", b.APIKey)
	}
	
	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to receive XCM message: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("failed to receive XCM message: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode receive response: %v", err)
	}
	
	// Extract the transaction hash
	txHash, ok := response["tx_hash"].(string)
	if (!ok) {
		return "", errors.New("transaction hash not found in response")
	}
	
	return txHash, nil
}

// ExecuteXCMCall executes a remote call on a target chain via XCM
func (b *PolkadotBridge) ExecuteXCMCall(destinationChainID, callData string, weight uint64) (string, error) {
	// Create the payload for an XCM call
	payload := map[string]interface{}{
		"call_data": callData,
		"weight":    weight,
	}
	
	// Send as an XCM message
	return b.SendXCMMessage(destinationChainID, "Call", payload)
}

// QueryCrossChainStatus queries the status of cross-chain operations
func (b *PolkadotBridge) QueryCrossChainStatus(opType string, limit, offset int) ([]map[string]interface{}, error) {
	// Create the request URL
	url := fmt.Sprintf("%s/xcm/operations?parachain_id=%s&type=%s&limit=%d&offset=%d", 
		b.RelayEndpoint, b.ParachainID, opType, limit, offset)
	
	// Send the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// Add headers
	if b.APIKey != "" {
		req.Header.Set("X-API-Key", b.APIKey)
	}
	
	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query cross-chain operations: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to query cross-chain operations: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode operations response: %v", err)
	}
	
	// Extract operations
	operations, ok := response["operations"].([]interface{})
	if (!ok) {
		return nil, errors.New("response does not contain operations")
	}
	
	// Process operations
	result := []map[string]interface{}{}
	for _, op := range operations {
		if opMap, ok := op.(map[string]interface{}); ok {
			result = append(result, opMap)
		}
	}
	
	return result, nil
}

// GetRelayChainStatus gets the status of the relay chain
func (b *PolkadotBridge) GetRelayChainStatus() (map[string]interface{}, error) {
	// Create the request URL
	url := fmt.Sprintf("%s/status", b.RelayEndpoint)
	
	// Send the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// Add headers
	if b.APIKey != "" {
		req.Header.Set("X-API-Key", b.APIKey)
	}
	
	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get relay chain status: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get relay chain status: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode status response: %v", err)
	}
	
	return status, nil
}

// GetRegisteredParachains gets a list of registered parachains on the relay chain
func (b *PolkadotBridge) GetRegisteredParachains() ([]map[string]interface{}, error) {
	// Create the request URL
	url := fmt.Sprintf("%s/parachains", b.RelayEndpoint)
	
	// Send the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// Add headers
	if b.APIKey != "" {
		req.Header.Set("X-API-Key", b.APIKey)
	}
	
	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get parachains: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get parachains: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode parachains response: %v", err)
	}
	
	// Extract parachains
	parachains, ok := response["parachains"].([]interface{})
	if (!ok) {
		return nil, errors.New("response does not contain parachains")
	}
	
	// Process parachains
	result := []map[string]interface{}{}
	for _, parachain := range parachains {
		if parachainMap, ok := parachain.(map[string]interface{}); ok {
			result = append(result, parachainMap)
		}
	}
	
	return result, nil
}

// RegisterParachain registers this chain as a parachain on the relay chain (administrative function)
func (b *PolkadotBridge) RegisterParachain(parachainID, wasmRuntime, wasmCode string) (string, error) {
	// Create the registration request
	registrationRequest := map[string]interface{}{
		"parachain_id": parachainID,
		"wasm_runtime": wasmRuntime,
		"wasm_code":    wasmCode,
	}
	
	// Convert to JSON
	jsonData, err := json.Marshal(registrationRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal registration request: %v", err)
	}
	
	// Create HTTP request
	url := fmt.Sprintf("%s/admin/parachains", b.RelayEndpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create registration request: %v", err)
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	if b.APIKey != "" {
		req.Header.Set("X-API-Key", b.APIKey)
	}
	
	// Send the request
	client := &http.Client{Timeout: 120 * time.Second} // Longer timeout for parachain registration
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to register parachain: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("failed to register parachain: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode registration response: %v", err)
	}
	
	// Extract the transaction hash
	txHash, ok := response["tx_hash"].(string)
	if (!ok) {
		return "", errors.New("transaction hash not found in response")
	}
	
	return txHash, nil
}

// SetMultiLocationForAsset sets the multilocation for an asset
func (b *PolkadotBridge) SetMultiLocationForAsset(assetID string, multiLocation map[string]interface{}) error {
	// Check if the asset is registered
	asset, exists := b.RegisteredAssets[assetID]
	if (!exists) {
		return fmt.Errorf("asset %s is not registered", assetID)
	}
	
	// Update the multilocation
	asset.MultiLocation = multiLocation
	b.RegisteredAssets[assetID] = asset
	
	return nil
}

// GetParachainId returns the parachain ID
func (b *PolkadotBridge) GetParachainId() string {
	return b.ParachainID
}

// GetRelayChainId returns the relay chain ID
func (b *PolkadotBridge) GetRelayChainId() string {
	return b.RelayChainID
}

// GetXCMVersion queries the supported XCM version for a destination chain
func (b *PolkadotBridge) GetXCMVersion(destinationChainID string) (string, error) {
	// Create the request URL
	url := fmt.Sprintf("%s/xcm/version?destination=%s", b.RelayEndpoint, destinationChainID)
	
	// Send the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	
	// Add headers
	if b.APIKey != "" {
		req.Header.Set("X-API-Key", b.APIKey)
	}
	
	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get XCM version: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get XCM version: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode version response: %v", err)
	}
	
	// Extract version
	version, ok := response["version"].(string)
	if (!ok) {
		return "V2", nil // Default to V2 if not specified
	}
	
	return version, nil
}

// CreateXCMAsset creates a new XCM asset on the chain
func (b *PolkadotBridge) CreateXCMAsset(name, symbol string, decimals int, multiLocation map[string]interface{}) (string, error) {
	// Create the asset request
	assetRequest := map[string]interface{}{
		"name":           name,
		"symbol":         symbol,
		"decimals":       decimals,
		"multi_location": multiLocation,
	}
	
	// Convert to JSON
	jsonData, err := json.Marshal(assetRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal asset request: %v", err)
	}
	
	// Create HTTP request
	url := fmt.Sprintf("%s/assets", b.RelayEndpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create asset request: %v", err)
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	if b.APIKey != "" {
		req.Header.Set("X-API-Key", b.APIKey)
	}
	
	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create XCM asset: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("failed to create XCM asset: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode asset response: %v", err)
	}
	
	// Extract the asset ID
	assetID, ok := response["asset_id"].(string)
	if (!ok) {
		return "", errors.New("asset ID not found in response")
	}
	
	// Register the new asset
	b.RegisteredAssets[assetID] = XCMAssetDetails{
		AssetID:        assetID,
		Name:           name,
		Symbol:         symbol,
		Decimals:       decimals,
		OriginChain:    b.ChainID,
		OriginLocation: "",
		MultiLocation:  multiLocation,
	}
	
	return assetID, nil
}