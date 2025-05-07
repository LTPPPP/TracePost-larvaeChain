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

// PolkadotBridge implements cross-chain functionality with Polkadot parachains
type PolkadotBridge struct {
	RelayChainEndpoint string
	RelayChainID       string
	ParachainID        string
	ParachainEndpoint  string
	APIKey             string
	ProxyAddress       string
	LastSyncHeight     int64
	// Enhanced fields for XCM v3 support
	SupportedXCMVersions []string
	RegisteredAssets     map[string]XCMAssetDetails
	CrossChainFormats    []string
}

// XCMAssetDetails holds details about an asset that can be transferred via XCM
type XCMAssetDetails struct {
	AssetID        string `json:"asset_id"`
	Name           string `json:"name"`
	Decimals       int    `json:"decimals"`
	Symbol         string `json:"symbol"`
	LocationFormat string `json:"location_format"` // e.g., "Concrete", "Abstract"
	ParentChain    string `json:"parent_chain"`
	IsReserve      bool   `json:"is_reserve"`
	IsNative       bool   `json:"is_native"`
}

// XCMMessage represents a cross-chain message in the Polkadot ecosystem
type XCMMessage struct {
	SourceChainID      string                 `json:"source_chain_id"`
	DestinationChainID string                 `json:"destination_chain_id"`
	MessageID          string                 `json:"message_id"`
	MessageType        string                 `json:"message_type"`
	Payload            map[string]interface{} `json:"payload"`
	Timestamp          int64                  `json:"timestamp"`
	ProofData          string                 `json:"proof_data,omitempty"`
	Status             string                 `json:"status"`
	// Enhanced fields for XCM v3
	XCMVersion         string                 `json:"xcm_version"`
	Instructions       []map[string]interface{} `json:"instructions,omitempty"`
	TransactVersion    string                 `json:"transact_version,omitempty"`
}

// NewPolkadotBridge creates a new Polkadot bridge instance
func NewPolkadotBridge(relayEndpoint, relayChainID, parachainID, parachainEndpoint, apiKey string) *PolkadotBridge {
	return &PolkadotBridge{
		RelayChainEndpoint: relayEndpoint,
		RelayChainID:       relayChainID,
		ParachainID:        parachainID,
		ParachainEndpoint:  parachainEndpoint,
		APIKey:             apiKey,
		ProxyAddress:       "",
		LastSyncHeight:     0,
		SupportedXCMVersions: []string{"V2", "V3"},
		RegisteredAssets:   make(map[string]XCMAssetDetails),
		CrossChainFormats: []string{"xcm", "generic"},
	}
}

// SetProxyAddress sets the proxy address for sending cross-chain messages
func (b *PolkadotBridge) SetProxyAddress(proxyAddress string) {
	b.ProxyAddress = proxyAddress
}

// RegisterAsset registers an asset for XCM transfers
func (b *PolkadotBridge) RegisterAsset(assetDetails XCMAssetDetails) {
	b.RegisteredAssets[assetDetails.AssetID] = assetDetails
}

// SendXCMMessage sends a cross-chain message to a Polkadot parachain
func (b *PolkadotBridge) SendXCMMessage(destinationParachainID string, messageType string, payload map[string]interface{}) (string, error) {
	// Create a unique message ID
	payloadJSON, _ := json.Marshal(payload)
	hash := sha256.Sum256(payloadJSON)
	messageID := hex.EncodeToString(hash[:])

	// Create the XCM message
	message := XCMMessage{
		SourceChainID:      b.ParachainID,
		DestinationChainID: destinationParachainID,
		MessageID:          messageID,
		MessageType:        messageType,
		Payload:            payload,
		Timestamp:          time.Now().UnixNano(),
		Status:             "pending",
		XCMVersion:         "V3", // Use the latest XCM version
	}

	// For XCM V3, we add instructions if this is a complex message
	if messageType == "transfer" || messageType == "data_transfer" {
		// Add transfer instructions for XCM v3
		message.Instructions = buildXCMV3Instructions(messageType, payload, destinationParachainID)
	}

	// Prepare the API request
	xcmRequest := map[string]interface{}{
		"message":          message,
		"call":             "xcmPallet.send",
		"destination_type": "V3",
		"destination": map[string]interface{}{
			"parents":  1, // Relay chain is one level up
			"interior": map[string]interface{}{
				"X1": map[string]interface{}{
					"Parachain": destinationParachainID,
				},
			},
		},
		"beneficiary": b.ProxyAddress,
		"assets": []map[string]interface{}{
			{
				"id": map[string]interface{}{
					"Concrete": map[string]interface{}{
						"parents":  0,
						"interior": "Here",
					},
				},
				"fun": map[string]interface{}{
					"Fungible": 1000000000, // Amount in smallest units
				},
			},
		},
		"fee_asset_item": 0,
		"weight_limit":   "Unlimited",
	}

	// Convert to JSON
	jsonData, err := json.Marshal(xcmRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal XCM request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", b.ParachainEndpoint+"/api/xcm/send", bytes.NewBuffer(jsonData))
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
				return "", fmt.Errorf("XCM send failed: %s", errMsg)
			}
		}
		return "", fmt.Errorf("XCM send failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	// Extract the extrinsic hash
	extrinsicHash, ok := response["extrinsic_hash"].(string)
	if !ok {
		return messageID, nil // Fallback to message ID if hash not available
	}

	return extrinsicHash, nil
}

// buildXCMV3Instructions creates XCM V3 instructions based on message type and payload
func buildXCMV3Instructions(messageType string, payload map[string]interface{}, destinationParachainID string) []map[string]interface{} {
	instructions := []map[string]interface{}{}
	
	if messageType == "transfer" {
		// For token transfers
		amount, _ := payload["amount"].(float64)
		beneficiary, _ := payload["beneficiary"].(string)
		
		// Withdraw asset
		instructions = append(instructions, map[string]interface{}{
			"WithdrawAsset": []map[string]interface{}{
				{
					"id": map[string]interface{}{
						"Concrete": map[string]interface{}{
							"parents":  0,
							"interior": "Here",
						},
					},
					"fun": map[string]interface{}{
						"Fungible": amount,
					},
				},
			},
		})
		
		// Buy execution with asset
		instructions = append(instructions, map[string]interface{}{
			"BuyExecution": map[string]interface{}{
				"fees": map[string]interface{}{
					"id": map[string]interface{}{
						"Concrete": map[string]interface{}{
							"parents":  0,
							"interior": "Here",
						},
					},
					"fun": map[string]interface{}{
						"Fungible": amount * 0.01, // 1% for fees
					},
				},
				"weight_limit": "Unlimited",
			},
		})
		
		// Deposit asset
		instructions = append(instructions, map[string]interface{}{
			"DepositAsset": map[string]interface{}{
				"assets": []map[string]interface{}{
					{
						"id": map[string]interface{}{
							"Concrete": map[string]interface{}{
								"parents":  0,
								"interior": "Here",
							},
						},
						"fun": map[string]interface{}{
							"Fungible": amount * 0.99, // Remaining after fees
						},
					},
				},
				"beneficiary": map[string]interface{}{
					"parents": 0,
					"interior": map[string]interface{}{
						"X1": map[string]interface{}{
							"AccountId32": map[string]interface{}{
								"id": beneficiary,
								"network": "Any",
							},
						},
					},
				},
			},
		})
	} else if messageType == "data_transfer" {
		// For data transfers
		data, _ := payload["data"].(string)
		
		// Transact instruction for running a pallet call to store data
		instructions = append(instructions, map[string]interface{}{
			"Transact": map[string]interface{}{
				"origin_kind": "SovereignAccount",
				"require_weight_at_most": 1000000000,
				"call": map[string]interface{}{
					"encoded": data, // This would be properly encoded call data in the real implementation
				},
			},
		})
	}
	
	return instructions
}

// RetrieveXCMMessage retrieves a cross-chain message by its ID
func (b *PolkadotBridge) RetrieveXCMMessage(messageID string) (*XCMMessage, error) {
	// Create the request URL
	url := fmt.Sprintf("%s/api/xcm/message/%s", b.ParachainEndpoint, messageID)

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
		return nil, fmt.Errorf("failed to retrieve XCM message: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve XCM message: HTTP %d", resp.StatusCode)
	}

	// Parse the response
	var message XCMMessage
	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		return nil, fmt.Errorf("failed to decode message: %v", err)
	}

	return &message, nil
}

// VerifyXCMMessage verifies a cross-chain message on the destination chain
func (b *PolkadotBridge) VerifyXCMMessage(messageID string, destinationParachainID string) (bool, error) {
	// If destinationParachainID is this bridge's parachain, check locally
	if destinationParachainID == b.ParachainID {
		message, err := b.RetrieveXCMMessage(messageID)
		if err != nil {
			return false, err
		}
		return message.Status == "executed", nil
	}

	// Otherwise, we need to check with the destination parachain
	// This would typically involve querying that parachain's API
	// For this example, we'll just simulate a check
	
	// Create the verification request
	verificationRequest := map[string]interface{}{
		"message_id": messageID,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(verificationRequest)
	if err != nil {
		return false, fmt.Errorf("failed to marshal verification request: %v", err)
	}

	// Determine destination endpoint
	// In a real system, you would have a registry of parachains and their endpoints
	destinationEndpoint := strings.Replace(b.ParachainEndpoint, b.ParachainID, destinationParachainID, 1)
	
	// Create HTTP request
	url := fmt.Sprintf("%s/api/xcm/verify/%s", destinationEndpoint, messageID)
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
	if !ok {
		return false, errors.New("verification response did not contain verification status")
	}

	return verified, nil
}

// SyncWithRelayChain synchronizes the bridge with the relay chain
func (b *PolkadotBridge) SyncWithRelayChain() error {
	// Create the sync request
	syncRequest := map[string]interface{}{
		"parachain_id":      b.ParachainID,
		"last_sync_height":  b.LastSyncHeight,
		"max_blocks_to_sync": 100,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(syncRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal sync request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", b.RelayChainEndpoint+"/api/sync", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create sync request: %v", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	if b.APIKey != "" {
		req.Header.Set("X-API-Key", b.APIKey)
	}

	// Send the request
	client := &http.Client{Timeout: 60 * time.Second}  // Longer timeout for sync
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to sync with relay chain: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to sync with relay chain: HTTP %d", resp.StatusCode)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode sync response: %v", err)
	}

	// Update the last sync height
	if newHeight, ok := response["new_sync_height"].(float64); ok {
		b.LastSyncHeight = int64(newHeight)
	}

	return nil
}

// GetXCMMessageStatus gets the status of an XCM message
func (b *PolkadotBridge) GetXCMMessageStatus(messageID string) (string, error) {
	message, err := b.RetrieveXCMMessage(messageID)
	if err != nil {
		return "", err
	}
	return message.Status, nil
}

// RegisterParachain registers a parachain with the bridge
func (b *PolkadotBridge) RegisterParachain(parachainID, parachainEndpoint string) error {
	// Create the registration request
	registrationRequest := map[string]interface{}{
		"parachain_id":      parachainID,
		"parachain_endpoint": parachainEndpoint,
		"relay_chain_id":    b.RelayChainID,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(registrationRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal registration request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", b.RelayChainEndpoint+"/api/parachains/register", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create registration request: %v", err)
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
		return fmt.Errorf("failed to register parachain: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to register parachain: HTTP %d", resp.StatusCode)
	}

	return nil
}

// QueryChainState queries the state of a parachain
func (b *PolkadotBridge) QueryChainState(storageKey string) (map[string]interface{}, error) {
	// Create the query URL
	url := fmt.Sprintf("%s/api/state/storage/%s", b.ParachainEndpoint, storageKey)
	
	// Send the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create query request: %v", err)
	}
	
	// Add headers
	if b.APIKey != "" {
		req.Header.Set("X-API-Key", b.APIKey)
	}
	
	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query chain state: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to query chain state: HTTP %d", resp.StatusCode)
	}
	
	// Parse the response
	var state map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("failed to decode state response: %v", err)
	}
	
	return state, nil
}