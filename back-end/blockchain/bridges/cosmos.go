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
	"time"
)

// CosmosBridge implements cross-chain functionality with Cosmos IBC protocol
type CosmosBridge struct {
	NodeEndpoint     string
	ChainID          string
	APIKey           string
	IBCClientID      string
	IBCConnectionID  string
	IBCChannels      map[string]IBCChannel
	AccountAddress   string
	LastBlockHeight  int64
	// Enhanced fields for improved IBC support
	SupportedIBCVersions  []string
	RegisteredTokens      map[string]IBCTokenDetails
	IBCClientState        map[string]interface{}
	IBCConsensusState     map[string]interface{}
	TrustedChains         map[string]TrustedChainDetails
}

// TrustedChainDetails stores information about a trusted chain in IBC
type TrustedChainDetails struct {
	ChainID            string
	TrustingPeriod     int64 // in seconds
	MaxClockDrift      int64 // in seconds
	ClientType         string // e.g., "07-tendermint", "08-wasm", etc.
	LatestHeightVerified int64
	TrustLevel         float64 // e.g., 1/3 for Tendermint chains
}

// IBCTokenDetails holds details about a token that can be transferred via IBC
type IBCTokenDetails struct {
	Denom       string `json:"denom"`
	BaseDenom   string `json:"base_denom"` // original denomination
	ChannelID   string `json:"channel_id"` // source channel if this is a voucher token
	DisplayName string `json:"display_name"`
	Symbol      string `json:"symbol"`
	Decimals    int    `json:"decimals"`
	Origin      string `json:"origin"`     // original chain
	Path        string `json:"path"`       // trace path if this is an IBC token
}

// IBCChannel represents an IBC channel configuration
type IBCChannel struct {
	ChannelID             string
	PortID                string
	CounterpartyChannelID string
	CounterpartyPortID    string
	State                 string
	Version               string
	ConnectionHops        []string
	// Enhanced channel fields
	Ordering              string            // "ORDERED", "UNORDERED"
	ExtendedState         map[string]string // Additional state information
	PacketSequence        uint64            // Next sequence to send
	LastAckSequence       uint64            // Last acknowledged sequence
	TimeoutHeight         IBCHeight         // Default timeout height for this channel
	TimeoutTimestamp      int64             // Default timeout timestamp for this channel
}

// IBCMessage represents a cross-chain message in the Cosmos ecosystem
type IBCMessage struct {
	MessageID          string                 `json:"message_id"`
	SourceChainID      string                 `json:"source_chain_id"`
	DestinationChainID string                 `json:"destination_chain_id"`
	SourceChannel      string                 `json:"source_channel"`
	DestinationChannel string                 `json:"destination_channel"`
	SourcePort         string                 `json:"source_port"`
	DestinationPort    string                 `json:"destination_port"`
	Payload            map[string]interface{} `json:"payload"`
	Timestamp          int64                  `json:"timestamp"`
	Status             string                 `json:"status"`
	TimeoutHeight      IBCHeight              `json:"timeout_height"`
	TimeoutTimestamp   int64                  `json:"timeout_timestamp"`
	Proof              string                 `json:"proof,omitempty"`
	// Enhanced message fields
	PacketSequence      uint64                 `json:"packet_sequence,omitempty"`
	TimeoutBlock        uint64                 `json:"timeout_block,omitempty"`
	MaxGas              uint64                 `json:"max_gas,omitempty"`
	Memo                string                 `json:"memo,omitempty"`
	IBCVersion          string                 `json:"ibc_version"`
	AppVersion          string                 `json:"app_version,omitempty"`
	CallbackAddress     string                 `json:"callback_address,omitempty"`
	FeeMetadata         map[string]interface{} `json:"fee_metadata,omitempty"`
	RelayerAddress      string                 `json:"relayer_address,omitempty"`
}

// IBCHeight represents an IBC height with revision number and height
type IBCHeight struct {
	RevisionNumber uint64 `json:"revision_number"`
	RevisionHeight uint64 `json:"revision_height"`
}

// NewCosmosBridge creates a new Cosmos bridge instance
func NewCosmosBridge(nodeEndpoint, chainID, apiKey, accountAddress string) *CosmosBridge {
	return &CosmosBridge{
		NodeEndpoint:      nodeEndpoint,
		ChainID:           chainID,
		APIKey:            apiKey,
		IBCChannels:       make(map[string]IBCChannel),
		AccountAddress:    accountAddress,
		LastBlockHeight:   0,
		SupportedIBCVersions: []string{"1.0.0", "1.1.0", "1.2.0"},
		RegisteredTokens: make(map[string]IBCTokenDetails),
		IBCClientState:   make(map[string]interface{}),
		IBCConsensusState: make(map[string]interface{}),
		TrustedChains:    make(map[string]TrustedChainDetails),
	}
}

// AddIBCChannel adds an IBC channel to the bridge
func (b *CosmosBridge) AddIBCChannel(channelID, portID, counterpartyChannelID, counterpartyPortID, connectionID string) {
	b.IBCChannels[channelID] = IBCChannel{
		ChannelID:             channelID,
		PortID:                portID,
		CounterpartyChannelID: counterpartyChannelID,
		CounterpartyPortID:    counterpartyPortID,
		State:                 "OPEN", // Default state
		Version:               "ics20-1",
		ConnectionHops:        []string{connectionID},
		Ordering:              "UNORDERED", // Default ordering
		ExtendedState:         make(map[string]string),
		PacketSequence:        1,
		LastAckSequence:       0,
	}
}

// SetIBCConnectionDetails sets the IBC connection details
func (b *CosmosBridge) SetIBCConnectionDetails(clientID, connectionID string) {
	b.IBCClientID = clientID
	b.IBCConnectionID = connectionID
}

// RegisterIBCToken registers a token for IBC transfers
func (b *CosmosBridge) RegisterIBCToken(token IBCTokenDetails) {
	b.RegisteredTokens[token.Denom] = token
}

// AddTrustedChain adds a chain to the trusted chains list
func (b *CosmosBridge) AddTrustedChain(details TrustedChainDetails) {
	b.TrustedChains[details.ChainID] = details
}

// SendIBCPacket sends an IBC packet to another Cosmos chain
func (b *CosmosBridge) SendIBCPacket(destinationChainID, channelID string, payload map[string]interface{}, timeoutInMinutes int) (string, error) {
	// Check if the channel exists
	channel, exists := b.IBCChannels[channelID]
	if !exists {
		return "", fmt.Errorf("IBC channel %s does not exist", channelID)
	}

	// Create a unique message ID
	payloadJSON, _ := json.Marshal(payload)
	hash := sha256.Sum256(payloadJSON)
	messageID := hex.EncodeToString(hash[:])

	// Calculate timeout
	timeoutTimestamp := time.Now().Add(time.Duration(timeoutInMinutes) * time.Minute).UnixNano()
	
	// Get the current height for timeout calculation
	currentHeight, err := b.GetLastBlockHeight()
	if err != nil {
		return "", fmt.Errorf("failed to get current block height: %v", err)
	}
	
	// Create default timeout height
	timeoutHeight := IBCHeight{
		RevisionNumber: 0,
		RevisionHeight: uint64(currentHeight) + 1000, // Default 1000 blocks timeout
	}
	
	// Get current sequence
	sequence := channel.PacketSequence
	
	// Create the IBC message
	message := IBCMessage{
		MessageID:          messageID,
		SourceChainID:      b.ChainID,
		DestinationChainID: destinationChainID,
		SourceChannel:      channelID,
		DestinationChannel: channel.CounterpartyChannelID,
		SourcePort:         channel.PortID,
		DestinationPort:    channel.CounterpartyPortID,
		Payload:            payload,
		Timestamp:          time.Now().UnixNano(),
		Status:             "pending",
		TimeoutHeight:      timeoutHeight,
		TimeoutTimestamp:   timeoutTimestamp,
		PacketSequence:     sequence,
		IBCVersion:         "1.2.0", // Latest IBC version
		AppVersion:         "ics20-1", // ICS-20 for token transfers
		Memo:               fmt.Sprintf("IBC transfer from %s to %s", b.ChainID, destinationChainID),
	}

	// Increment the sequence for next use
	b.IBCChannels[channelID] = IBCChannel{
		ChannelID:             channel.ChannelID,
		PortID:                channel.PortID,
		CounterpartyChannelID: channel.CounterpartyChannelID,
		CounterpartyPortID:    channel.CounterpartyPortID,
		State:                 channel.State,
		Version:               channel.Version,
		ConnectionHops:        channel.ConnectionHops,
		Ordering:              channel.Ordering,
		ExtendedState:         channel.ExtendedState,
		PacketSequence:        sequence + 1,
		LastAckSequence:       channel.LastAckSequence,
		TimeoutHeight:         channel.TimeoutHeight,
		TimeoutTimestamp:      channel.TimeoutTimestamp,
	}

	// Prepare the IBC packet request
	packetRequest := map[string]interface{}{
		"source_port":    channel.PortID,
		"source_channel": channelID,
		"token":          payload, // For ICS-20 transfers
		"sender":         b.AccountAddress,
		"timeout_height": map[string]interface{}{
			"revision_number": message.TimeoutHeight.RevisionNumber,
			"revision_height": message.TimeoutHeight.RevisionHeight,
		},
		"timeout_timestamp": message.TimeoutTimestamp,
		"packet_sequence":   sequence,
		"memo":              message.Memo,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(packetRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal IBC packet request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", b.NodeEndpoint+"/ibc/packets", bytes.NewBuffer(jsonData))
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
		return "", fmt.Errorf("failed to send IBC packet: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		var errResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			if errMsg, ok := errResp["error"].(string); ok {
				return "", fmt.Errorf("IBC packet send failed: %s", errMsg)
			}
		}
		return "", fmt.Errorf("IBC packet send failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	// Extract the transaction hash
	txHash, ok := response["tx_hash"].(string)
	if !ok {
		return messageID, nil // Fallback to message ID if hash not available
	}

	return txHash, nil
}

// GetIBCPacketStatus gets the status of an IBC packet
func (b *CosmosBridge) GetIBCPacketStatus(txHash string) (string, error) {
	// Create the request URL
	url := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", b.NodeEndpoint, txHash)

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

	// Extract tx_response if it exists
	txResponse, ok := txResult["tx_response"].(map[string]interface{})
	if !ok {
		return "unknown", nil
	}

	// Check if the transaction was successful
	code, ok := txResponse["code"].(float64)
	if !ok {
		return "unknown", nil
	}

	if code == 0 {
		return "success", nil
	} else {
		rawLog, _ := txResult["raw_log"].(string)
		return "failed", fmt.Errorf("transaction failed with code %d: %s", int(code), rawLog)
	}
}

// VerifyIBCPacket verifies an IBC packet on the destination chain
func (b *CosmosBridge) VerifyIBCPacket(sourceChainID, sourceChannel, destinationChannel, packetSequence string) (bool, error) {
	// Create the verification request
	verificationRequest := map[string]interface{}{
		"source_chain_id":      sourceChainID,
		"source_channel":       sourceChannel,
		"destination_channel":  destinationChannel,
		"packet_sequence":      packetSequence,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(verificationRequest)
	if err != nil {
		return false, fmt.Errorf("failed to marshal verification request: %v", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/ibc/packets/verify", b.NodeEndpoint)
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
		return false, fmt.Errorf("failed to verify IBC packet: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to verify IBC packet: HTTP %d", resp.StatusCode)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("failed to decode verification response: %v", err)
	}

	// Check the verification result
	received, ok := response["received"].(bool)
	if !ok {
		return false, errors.New("verification response did not contain received status")
	}

	return received, nil
}

// QueryIBCChannels queries all IBC channels on the chain
func (b *CosmosBridge) QueryIBCChannels() ([]IBCChannel, error) {
	// Create the request URL
	url := fmt.Sprintf("%s/ibc/core/channel/v1/channels", b.NodeEndpoint)

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
		return nil, fmt.Errorf("failed to query IBC channels: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to query IBC channels: HTTP %d", resp.StatusCode)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode channels response: %v", err)
	}

	// Extract channels
	channelsData, ok := response["channels"].([]interface{})
	if !ok {
		return nil, errors.New("response does not contain channels")
	}

	// Process the channels
	channels := []IBCChannel{}
	for _, channelData := range channelsData {
		channelMap, ok := channelData.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract channel details
		channelID, _ := channelMap["channel_id"].(string)
		portID, _ := channelMap["port_id"].(string)
		counterpartyData, _ := channelMap["counterparty"].(map[string]interface{})
		counterpartyChannelID, _ := counterpartyData["channel_id"].(string)
		counterpartyPortID, _ := counterpartyData["port_id"].(string)
		state, _ := channelMap["state"].(string)
		version, _ := channelMap["version"].(string)
		ordering, _ := channelMap["ordering"].(string)

		// Extract connection hops
		connectionHops := []string{}
		connectionHopsData, ok := channelMap["connection_hops"].([]interface{})
		if ok {
			for _, hop := range connectionHopsData {
				if hopStr, ok := hop.(string); ok {
					connectionHops = append(connectionHops, hopStr)
				}
			}
		}

		// Create and add the channel
		channel := IBCChannel{
			ChannelID:             channelID,
			PortID:                portID,
			CounterpartyChannelID: counterpartyChannelID,
			CounterpartyPortID:    counterpartyPortID,
			State:                 state,
			Version:               version,
			ConnectionHops:        connectionHops,
			Ordering:              ordering,
			ExtendedState:         make(map[string]string),
			PacketSequence:        1,
			LastAckSequence:       0,
		}

		channels = append(channels, channel)
		
		// Also update our internal channels map
		b.IBCChannels[channelID] = channel
	}

	return channels, nil
}

// GetLastBlockHeight gets the latest block height from the chain
func (b *CosmosBridge) GetLastBlockHeight() (int64, error) {
	// If we have a cached value that's recent, return it
	if b.LastBlockHeight > 0 && time.Now().Unix()-b.LastBlockHeight < 60 {
		return b.LastBlockHeight, nil
	}

	// Create the request URL
	url := fmt.Sprintf("%s/blocks/latest", b.NodeEndpoint)

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

	// Extract block data
	blockData, ok := blockResponse["block"].(map[string]interface{})
	if !ok {
		return 0, errors.New("response does not contain block data")
	}

	// Extract header
	header, ok := blockData["header"].(map[string]interface{})
	if !ok {
		return 0, errors.New("block data does not contain header")
	}

	// Extract height
	heightStr, ok := header["height"].(string)
	if !ok {
		return 0, errors.New("header does not contain height")
	}

	// Parse height
	var height int64
	_, err = fmt.Sscanf(heightStr, "%d", &height)
	if err != nil {
		return 0, fmt.Errorf("failed to parse height: %v", err)
	}

	// Update cached height
	b.LastBlockHeight = height

	return height, nil
}

// QueryIBCDenoms queries all IBC denominations on the chain
func (b *CosmosBridge) QueryIBCDenoms() ([]IBCTokenDetails, error) {
	// Create the request URL
	url := fmt.Sprintf("%s/ibc/apps/transfer/v1/denom_traces", b.NodeEndpoint)

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
		return nil, fmt.Errorf("failed to query IBC denoms: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to query IBC denoms: HTTP %d", resp.StatusCode)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode denoms response: %v", err)
	}

	// Extract denom traces
	denomTraces, ok := response["denom_traces"].([]interface{})
	if !ok {
		return nil, errors.New("response does not contain denom_traces")
	}

	// Process the denom traces
	tokens := []IBCTokenDetails{}
	for _, traceData := range denomTraces {
		traceMap, ok := traceData.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract denom details
		path, _ := traceMap["path"].(string)
		baseDenom, _ := traceMap["base_denom"].(string)
		
		// For IBC tokens, the denom is constructed as "ibc/{hash}"
		// We'll set some reasonable defaults for display info
		tokens = append(tokens, IBCTokenDetails{
			Denom:       fmt.Sprintf("ibc/%s", sha256.Sum256([]byte(path+"/"+baseDenom))),
			BaseDenom:   baseDenom,
			DisplayName: fmt.Sprintf("IBC %s", baseDenom),
			Symbol:      baseDenom,
			Decimals:    6, // Most Cosmos tokens have 6 decimals
			Origin:      "unknown", // Would need to parse the path to determine
			Path:        path,
		})
	}

	return tokens, nil
}

// GetChannelPacketCommitment gets the commitment for a specific packet
func (b *CosmosBridge) GetChannelPacketCommitment(portID, channelID string, sequence uint64) (string, error) {
	// Create the request URL
	url := fmt.Sprintf("%s/ibc/core/channel/v1/channels/%s/ports/%s/packet_commitments/%d", 
		b.NodeEndpoint, channelID, portID, sequence)

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
		return "", fmt.Errorf("failed to get packet commitment: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get packet commitment: HTTP %d", resp.StatusCode)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode commitment response: %v", err)
	}

	// Extract commitment
	commitment, ok := response["commitment"].(string)
	if !ok {
		return "", errors.New("response does not contain commitment")
	}

	return commitment, nil
}