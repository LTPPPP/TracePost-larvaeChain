package api

import (
	"strconv"
	"time"
	"fmt"
	"strings"
	
	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain/bridges"
	"github.com/LTPPPP/TracePost-larvaeChain/config"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
)

// InteroperabilityRegisterChainRequest represents a request to register an external blockchain
type InteroperabilityRegisterChainRequest struct {
	ChainID   string `json:"chain_id"`
	ChainType string `json:"chain_type"`
	Endpoint  string `json:"endpoint"`
}

// InteroperabilityShareBatchRequest represents a request to share a batch with an external blockchain
type InteroperabilityShareBatchRequest struct {
	BatchID      string `json:"batch_id"`
	DestChainID  string `json:"dest_chain_id"`
	DataStandard string `json:"data_standard"`
}

// CrossChainTransactionResponse represents a response for a cross-chain transaction
type CrossChainTransactionResponse struct {
	SourceTxID      string                 `json:"source_tx_id"`
	DestinationTxID string                 `json:"destination_tx_id"`
	SourceChainID   string                 `json:"source_chain_id"`
	DestChainID     string                 `json:"dest_chain_id"`
	Status          string                 `json:"status"`
	Timestamp       string                 `json:"timestamp"`
	Payload         map[string]interface{} `json:"payload,omitempty"`
}

// PolkadotBridgeRequest represents a request to create a Polkadot bridge
type PolkadotBridgeRequest struct {
	ChainID       string `json:"chain_id"`
	RelayEndpoint string `json:"relay_endpoint"`
	RelayChainID  string `json:"relay_chain_id"`
	ParachainID   string `json:"parachain_id"`
	APIKey        string `json:"api_key"`
}

// CosmosBridgeRequest represents a request to create a Cosmos bridge
type CosmosBridgeRequest struct {
	ChainID         string `json:"chain_id"`
	NodeEndpoint    string `json:"node_endpoint"`
	APIKey          string `json:"api_key"`
	AccountAddress  string `json:"account_address"`
}

// IBCChannelRequest represents a request to add an IBC channel
type IBCChannelRequest struct {
	ChainID               string `json:"chain_id"`
	ChannelID             string `json:"channel_id"`
	PortID                string `json:"port_id"`
	CounterpartyChannelID string `json:"counterparty_channel_id"`
	CounterpartyPortID    string `json:"counterparty_port_id"`
	ConnectionID          string `json:"connection_id"`
}

// XCMMessageRequest represents a request to send an XCM message
type XCMMessageRequest struct {
	SourceChainID string                 `json:"source_chain_id"`
	DestChainID   string                 `json:"dest_chain_id"`
	MessageType   string                 `json:"message_type"`
	Payload       map[string]interface{} `json:"payload"`
}

// IBCPacketRequest represents a request to send an IBC packet
type IBCPacketRequest struct {
	SourceChainID    string                 `json:"source_chain_id"`
	DestChainID      string                 `json:"dest_chain_id"`
	ChannelID        string                 `json:"channel_id"`
	Payload          map[string]interface{} `json:"payload"`
	TimeoutInMinutes int                    `json:"timeout_in_minutes"`
}

// VerifyTransactionRequest represents a request to verify a cross-chain transaction
type VerifyTransactionRequest struct {
	TxID          string `json:"tx_id"`
	Protocol      string `json:"protocol"`
	SourceChainID string `json:"source_chain_id"`
	DestChainID   string `json:"dest_chain_id"`
}

// RegisterExternalChain registers an external blockchain for interoperability
// @Summary Register an external blockchain
// @Description Register an external blockchain for cross-chain communication
// @Tags interoperability
// @Accept json
// @Produce json
// @Param request body InteroperabilityRegisterChainRequest true "Chain registration details"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/chains [post]
func RegisterExternalChain(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Parse request
	var req InteroperabilityRegisterChainRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.ChainID == "" || req.ChainType == "" || req.Endpoint == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required fields")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Register chain
	connectionID, err := blockchainClient.InteropClient.RegisterChain(req.ChainID, req.ChainType, req.Endpoint)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to register chain: "+err.Error())
	}
	
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Chain registered successfully",
		Data: map[string]string{
			"connection_id": connectionID,
		},
	})
}

// ShareBatchWithExternalChain shares a batch with an external blockchain
// @Summary Share a batch with external blockchain
// @Description Share a batch with an external blockchain using the specified data standard
// @Tags interoperability
// @Accept json
// @Produce json
// @Param request body InteroperabilityShareBatchRequest true "Batch sharing details"
// @Success 200 {object} SuccessResponse{data=CrossChainTransactionResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/share-batch [post]
func ShareBatchWithExternalChain(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Parse request
	var req InteroperabilityShareBatchRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.BatchID == "" || req.DestChainID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required fields")
	}
	
	// Use default data standard if not specified
	if req.DataStandard == "" {
		req.DataStandard = cfg.InteropDefaultStandard
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batches WHERE batch_id = $1)", req.BatchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}
	
	// Share batch with external chain
	destTxID, err := blockchainClient.ShareBatchWithExternalChain(req.BatchID, req.DestChainID, req.DataStandard)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to share batch: "+err.Error())
	}
	
	// Construct response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch shared successfully",
		Data: CrossChainTransactionResponse{
			SourceTxID:      "local-tx-" + destTxID[:8], // Simplified for example
			DestinationTxID: destTxID,
			SourceChainID:   cfg.BlockchainChainID,
			DestChainID:     req.DestChainID,
			Status:          "completed",
			Timestamp:       time.Now().Format(time.RFC3339),
		},
	})
}

// ExportBatchToGS1EPCIS exports a batch to GS1 EPCIS format
// @Summary Export batch to GS1 EPCIS
// @Description Export a batch to GS1 EPCIS format for interoperability
// @Tags interoperability
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/export/{batchId} [get]
func ExportBatchToGS1EPCIS(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
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
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Export batch to GS1 EPCIS
	epcisData, err := blockchainClient.ExportBatchToGS1EPCIS(batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to export batch: "+err.Error())
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch exported successfully",
		Data:    epcisData,
	})
}

// GetInteropBatchFromBlockchain returns batch data from blockchain through interoperability layer
// @Summary Get batch from blockchain via interoperability layer
// @Description Retrieve batch data directly from the blockchain using the interoperability layer
// @Tags blockchain,interoperability
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/blockchain/batch/{batchId} [get]
func GetInteropBatchFromBlockchain(c *fiber.Ctx) error {
	// Get batch ID from params
	batchIDStr := c.Params("batchId")
	if batchIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	batchID, err := strconv.Atoi(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
	}

	// Check if batch exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Get batch transactions from blockchain
	blockchainTxs, err := blockchainClient.GetBatchTransactions(strconv.Itoa(batchID))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve batch data from blockchain")
	}

	// Get blockchain records from database
	rows, err := db.DB.Query(`
		SELECT tx_id, metadata_hash, created_at
		FROM blockchain_record
		WHERE related_table = 'batch' AND related_id = $1
		ORDER BY created_at ASC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse blockchain records
	type BlockchainTxRecord struct {
		TxID         string    `json:"tx_id"`
		MetadataHash string    `json:"metadata_hash"`
		Timestamp    string    `json:"timestamp"`
		BlockchainTx interface{} `json:"blockchain_tx,omitempty"`
	}

	var records []BlockchainTxRecord
	for rows.Next() {
		var record BlockchainTxRecord
		var created string
		err := rows.Scan(
			&record.TxID,
			&record.MetadataHash,
			&created,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse blockchain record data")
		}
		record.Timestamp = created

		// Find matching transaction from blockchain
		for _, tx := range blockchainTxs {
			if tx.TxID == record.TxID {
				record.BlockchainTx = tx
				break
			}
		}

		records = append(records, record)
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch blockchain data retrieved successfully",
		Data:    records,
	})
}

// GetEventFromBlockchain returns event data from blockchain
// @Summary Get event from blockchain
// @Description Retrieve event data directly from the blockchain
// @Tags blockchain
// @Accept json
// @Produce json
// @Param eventId path string true "Event ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /blockchain/event/{eventId} [get]
func GetEventFromBlockchain(c *fiber.Ctx) error {
	// Get event ID from params
	eventIDStr := c.Params("eventId")
	if eventIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Event ID is required")
	}
	
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid event ID format")
	}

	// Check if event exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM event WHERE id = $1)", eventID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Event not found")
	}

	// Get event data from database
	var batchID int
	err = db.DB.QueryRow("SELECT batch_id FROM event WHERE id = $1", eventID).Scan(&batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Get event transactions from blockchain
	blockchainTxs, err := blockchainClient.GetEventTransactions(strconv.Itoa(eventID))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve event data from blockchain")
	}

	// Get blockchain records from database
	rows, err := db.DB.Query(`
		SELECT tx_id, metadata_hash, created_at
		FROM blockchain_record
		WHERE related_table = 'event' AND related_id = $1
		ORDER BY created_at ASC
	`, eventID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse blockchain records
	type BlockchainTxRecord struct {
		TxID         string    `json:"tx_id"`
		MetadataHash string    `json:"metadata_hash"`
		Timestamp    string    `json:"timestamp"`
		BlockchainTx interface{} `json:"blockchain_tx,omitempty"`
	}

	var records []BlockchainTxRecord
	for rows.Next() {
		var record BlockchainTxRecord
		var created string
		err := rows.Scan(
			&record.TxID,
			&record.MetadataHash,
			&created,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse blockchain record data")
		}
		record.Timestamp = created

		// Find matching transaction from blockchain
		for _, tx := range blockchainTxs {
			if tx.TxID == record.TxID {
				record.BlockchainTx = tx
				break
			}
		}

		records = append(records, record)
	}

	// Return success response with additional context
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Event blockchain data retrieved successfully",
		Data: map[string]interface{}{
			"event_id": eventID,
			"batch_id": batchID,
			"records":  records,
		},
	})
}

// GetDocumentFromBlockchain returns document data from blockchain
// @Summary Get document from blockchain
// @Description Retrieve document data directly from the blockchain
// @Tags blockchain
// @Accept json
// @Produce json
// @Param docId path string true "Document ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /blockchain/document/{docId} [get]
func GetDocumentFromBlockchain(c *fiber.Ctx) error {
	// Get document ID from params
	docIDStr := c.Params("docId")
	if docIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Document ID is required")
	}
	
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid document ID format")
	}

	// Check if document exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM document WHERE id = $1)", docID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Document not found")
	}

	// Get document data from database
	var batchID int
	var ipfsHash string
	err = db.DB.QueryRow("SELECT batch_id, ipfs_hash FROM document WHERE id = $1", docID).Scan(&batchID, &ipfsHash)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Get document transactions from blockchain
	blockchainTxs, err := blockchainClient.GetDocumentTransactions(strconv.Itoa(docID))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve document data from blockchain")
	}

	// Get blockchain records from database
	rows, err := db.DB.Query(`
		SELECT tx_id, metadata_hash, created_at
		FROM blockchain_record
		WHERE related_table = 'document' AND related_id = $1
		ORDER BY created_at ASC
	`, docID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse blockchain records
	type BlockchainTxRecord struct {
		TxID         string    `json:"tx_id"`
		MetadataHash string    `json:"metadata_hash"`
		Timestamp    string    `json:"timestamp"`
		BlockchainTx interface{} `json:"blockchain_tx,omitempty"`
	}

	var records []BlockchainTxRecord
	for rows.Next() {
		var record BlockchainTxRecord
		var created string
		err := rows.Scan(
			&record.TxID,
			&record.MetadataHash,
			&created,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse blockchain record data")
		}
		record.Timestamp = created

		// Find matching transaction from blockchain
		for _, tx := range blockchainTxs {
			if tx.TxID == record.TxID {
				record.BlockchainTx = tx
				break
			}
		}

		records = append(records, record)
	}

	// Return success response with additional context
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Document blockchain data retrieved successfully",
		Data: map[string]interface{}{
			"document_id": docID,
			"batch_id":    batchID,
			"ipfs_hash":   ipfsHash,
			"records":     records,
		},
	})
}

// GetEnvironmentDataFromBlockchain returns environment data from blockchain
// @Summary Get environment data from blockchain
// @Description Retrieve environment data directly from the blockchain
// @Tags blockchain
// @Accept json
// @Produce json
// @Param envId path string true "Environment Data ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /blockchain/environment/{envId} [get]
func GetEnvironmentDataFromBlockchain(c *fiber.Ctx) error {
	// Get environment data ID from params
	envIDStr := c.Params("envId")
	if envIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Environment data ID is required")
	}
	
	envID, err := strconv.Atoi(envIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid environment data ID format")
	}

	// Check if environment data exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM environment WHERE id = $1)", envID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Environment data not found")
	}

	// Get environment data from database
	var batchID int
	err = db.DB.QueryRow("SELECT batch_id FROM environment WHERE id = $1", envID).Scan(&batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		"http://localhost:26657",
		"private-key",
		"account-address",
		"tracepost-chain",
		"poa",
	)

	// Get environment data transactions from blockchain
	blockchainTxs, err := blockchainClient.GetEnvironmentDataTransactions(strconv.Itoa(envID))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve environment data from blockchain")
	}

	// Get blockchain records from database
	rows, err := db.DB.Query(`
		SELECT tx_id, metadata_hash, created_at
		FROM blockchain_record
		WHERE related_table = 'environment' AND related_id = $1
		ORDER BY created_at ASC
	`, envID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Parse blockchain records
	type BlockchainTxRecord struct {
		TxID         string    `json:"tx_id"`
		MetadataHash string    `json:"metadata_hash"`
		Timestamp    string    `json:"timestamp"`
		BlockchainTx interface{} `json:"blockchain_tx,omitempty"`
	}

	var records []BlockchainTxRecord
	for rows.Next() {
		var record BlockchainTxRecord
		var created string
		err := rows.Scan(
			&record.TxID,
			&record.MetadataHash,
			&created,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse blockchain record data")
		}
		record.Timestamp = created

		// Find matching transaction from blockchain
		for _, tx := range blockchainTxs {
			if tx.TxID == record.TxID {
				record.BlockchainTx = tx
				break
			}
		}

		records = append(records, record)
	}

	// Return success response with additional context
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Environment data blockchain records retrieved successfully",
		Data: map[string]interface{}{
			"environment_id": envID,
			"batch_id":       batchID,
			"records":        records,
		},
	})
}

// CreatePolkadotBridge creates a new Polkadot bridge
// @Summary Create a Polkadot bridge
// @Description Create a Polkadot bridge for cross-chain communication
// @Tags interoperability
// @Accept json
// @Produce json
// @Param request body PolkadotBridgeRequest true "Polkadot bridge details"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/bridges/polkadot [post]
func CreatePolkadotBridge(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Parse request
	var req PolkadotBridgeRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.ChainID == "" || req.RelayEndpoint == "" || req.RelayChainID == "" || req.ParachainID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required fields")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Create the Polkadot bridge
	err := blockchainClient.InteropClient.CreatePolkadotBridge(
		req.ChainID, 
		req.RelayEndpoint, 
		req.RelayChainID, 
		req.ParachainID, 
		req.APIKey,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create Polkadot bridge: "+err.Error())
	}
	
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Polkadot bridge created successfully",
		Data: map[string]string{
			"chain_id":      req.ChainID,
			"parachain_id":  req.ParachainID,
			"relay_chain_id": req.RelayChainID,
		},
	})
}

// CreateCosmosBridge creates a new Cosmos bridge
// @Summary Create a Cosmos bridge
// @Description Create a Cosmos bridge for cross-chain communication
// @Tags interoperability
// @Accept json
// @Produce json
// @Param request body CosmosBridgeRequest true "Cosmos bridge details"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/bridges/cosmos [post]
func CreateCosmosBridge(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Parse request
	var req CosmosBridgeRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.ChainID == "" || req.NodeEndpoint == "" || req.AccountAddress == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required fields")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Create the Cosmos bridge
	err := blockchainClient.InteropClient.CreateCosmosBridge(
		req.ChainID,
		req.NodeEndpoint,
		req.APIKey,
		req.AccountAddress,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create Cosmos bridge: "+err.Error())
	}
	
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Cosmos bridge created successfully",
		Data: map[string]string{
			"chain_id":        req.ChainID,
			"node_endpoint":   req.NodeEndpoint,
			"account_address": req.AccountAddress,
		},
	})
}

// AddIBCChannel adds an IBC channel to a Cosmos bridge
// @Summary Add an IBC channel
// @Description Add an IBC channel to a Cosmos bridge
// @Tags interoperability
// @Accept json
// @Produce json
// @Param request body IBCChannelRequest true "IBC channel details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/bridges/cosmos/channels [post]
func AddIBCChannel(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Parse request
	var req IBCChannelRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.ChainID == "" || req.ChannelID == "" || req.PortID == "" || 
	   req.CounterpartyChannelID == "" || req.CounterpartyPortID == "" || req.ConnectionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required fields")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Check if a Cosmos bridge exists for the chain
	cosmosBridge, exists := blockchainClient.InteropClient.CosmosBridges[req.ChainID]
	if !exists {
		return fiber.NewError(fiber.StatusBadRequest, "No Cosmos bridge found for the specified chain ID")
	}
	
	// Add the IBC channel
	cosmosBridge.AddIBCChannel(
		req.ChannelID,
		req.PortID,
		req.CounterpartyChannelID,
		req.CounterpartyPortID,
		req.ConnectionID,
	)
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "IBC channel added successfully",
		Data: map[string]string{
			"chain_id":   req.ChainID,
			"channel_id": req.ChannelID,
		},
	})
}

// SendXCMMessage sends an XCM message to a Polkadot chain
// @Summary Send XCM message
// @Description Send a cross-consensus message (XCM) to a Polkadot-based chain
// @Tags interoperability
// @Accept json
// @Produce json
// @Param request body XCMMessageRequest true "XCM message details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interoperability/xcm/message [post]
func SendXCMMessage(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Check if Substrate protocol is enabled
	if !cfg.SubstrateEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Substrate protocol is not enabled")
	}
	
	// Parse request
	var req XCMMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format: "+err.Error())
	}
	
	// Validate request
	if req.SourceChainID == "" || req.DestChainID == "" || req.MessageType == "" || req.Payload == nil {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required fields")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Create XCM message
	xcmMessage := bridges.XCMMessage{
		MessageID:          fmt.Sprintf("xcm-%s", time.Now().Format("20060102150405")),
		SourceChainID:      req.SourceChainID,
		DestinationChainID: req.DestChainID,
		MessageType:        req.MessageType,
		Payload:            req.Payload,
		Timestamp:          time.Now().Unix(),
		Status:             "pending",
		Version:            "v2",
	}
	
	// Send XCM message
	messageID, err := blockchainClient.InteropClient.SendXCMMessage(xcmMessage)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to send XCM message: "+err.Error())
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "XCM message sent successfully",
		Data: map[string]interface{}{
			"message_id": messageID,
			"source_chain_id": req.SourceChainID,
			"destination_chain_id": req.DestChainID,
			"status": "pending",
		},
	})
}

// SendIBCPacket sends an IBC packet to a Cosmos chain
// @Summary Send IBC packet
// @Description Send an Inter-Blockchain Communication (IBC) packet to a Cosmos-based chain
// @Tags interoperability
// @Accept json
// @Produce json
// @Param request body IBCPacketRequest true "IBC packet details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interoperability/ibc/packet [post]
func SendIBCPacket(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Check if IBC protocol is enabled
	if !cfg.IBCEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "IBC protocol is not enabled")
	}
	
	// Parse request
	var req IBCPacketRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format: "+err.Error())
	}
	
	// Validate request
	if req.SourceChainID == "" || req.DestChainID == "" || req.ChannelID == "" || req.Payload == nil {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required fields")
	}
	
	// Set default timeout if not specified
	if req.TimeoutInMinutes <= 0 {
		req.TimeoutInMinutes = 30 // Default 30 minutes timeout
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Get channel info
	channelInfo, found := blockchainClient.InteropClient.IBCChannels[req.ChannelID]
	if !found {
		return fiber.NewError(fiber.StatusBadRequest, "Channel not found")
	}
	
	// Create IBC message with packet data
	ibcMessage := bridges.IBCMessage{
		MessageID:          fmt.Sprintf("ibc-%s", time.Now().Format("20060102150405")),
		SourceChainID:      req.SourceChainID,
		DestinationChainID: req.DestChainID,
		SourceChannel:      req.ChannelID,
		DestinationChannel: channelInfo.CounterpartyChannelID,
		SourcePort:         channelInfo.PortID,
		DestinationPort:    channelInfo.CounterpartyPortID,
		Payload:            req.Payload,
		Timestamp:          time.Now().Unix(),
		Status:             "pending",
		TimeoutTimestamp:   time.Now().Add(time.Duration(req.TimeoutInMinutes) * time.Minute).Unix(),
	}
	
	// Send IBC packet
	packetID, err := blockchainClient.InteropClient.SendIBCPacket(ibcMessage)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to send IBC packet: "+err.Error())
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "IBC packet sent successfully",
		Data: map[string]interface{}{
			"packet_id": packetID,
			"source_chain_id": req.SourceChainID,
			"destination_chain_id": req.DestChainID,
			"channel_id": req.ChannelID,
			"status": "pending",
		},
	})
}

// VerifyTransaction verifies a cross-chain transaction
// @Summary Verify a cross-chain transaction
// @Description Verify a cross-chain transaction on the destination chain
// @Tags interoperability
// @Accept json
// @Produce json
// @Param request body VerifyTransactionRequest true "Transaction verification details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/verify [post]
func VerifyTransaction(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Parse request
	var req VerifyTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.TxID == "" || req.Protocol == "" || req.SourceChainID == "" || req.DestChainID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required fields")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Verify the transaction
	verified, err := blockchainClient.InteropClient.VerifyTransaction(
		req.TxID,
		req.Protocol,
		req.SourceChainID,
		req.DestChainID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to verify transaction: "+err.Error())
	}
	
	var message string
	if verified {
		message = "Transaction verified successfully"
	} else {
		message = "Transaction verification failed"
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: message,
		Data: map[string]interface{}{
			"verified":        verified,
			"tx_id":           req.TxID,
			"protocol":        req.Protocol,
			"source_chain_id": req.SourceChainID,
			"dest_chain_id":   req.DestChainID,
		},
	})
}

// GetTransactionStatus gets the status of a cross-chain transaction
// @Summary Get transaction status
// @Description Get the status of a cross-chain transaction
// @Tags interoperability
// @Accept json
// @Produce json
// @Param txId path string true "Transaction ID"
// @Param protocol path string true "Protocol (ibc, substrate, bridge)"
// @Param sourceChainId path string true "Source Chain ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/status/{protocol}/{sourceChainId}/{txId} [get]
func GetTransactionStatus(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Get parameters from path
	txID := c.Params("txId")
	protocol := c.Params("protocol")
	sourceChainID := c.Params("sourceChainId")
	
	// Validate parameters
	if txID == "" || protocol == "" || sourceChainID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required parameters")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Get the transaction status
	status, err := blockchainClient.InteropClient.GetTransactionStatus(
		txID,
		protocol,
		sourceChainID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get transaction status: "+err.Error())
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Transaction status retrieved successfully",
		Data: map[string]interface{}{
			"tx_id":          txID,
			"protocol":       protocol,
			"source_chain_id": sourceChainID,
			"status":         status,
		},
	})
}

// GetSupportedProtocols gets the list of supported cross-chain protocols
// @Summary Get supported protocols
// @Description Get the list of supported cross-chain protocols
// @Tags interoperability
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /interop/protocols [get]
func GetSupportedProtocols(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Get the supported protocols
	protocols := blockchainClient.InteropClient.GetSupportedProtocols()
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Supported protocols retrieved successfully",
		Data: map[string]interface{}{
			"protocols": protocols,
		},
	})
}

// ListConnectedChains lists all connected external blockchains
// @Summary List connected chains
// @Description List all connected external blockchains
// @Tags interoperability
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /interop/connected-chains [get]
func ListConnectedChains(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}
	
	// Get available networks
	networks := baasService.GetAvailableNetworks()
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Connected chains retrieved successfully",
		Data: map[string]interface{}{
			"chains": networks,
		},
	})
}

// GetChainStatus gets the status of a connected blockchain
// @Summary Get chain status
// @Description Get the status of a connected blockchain
// @Tags interoperability
// @Accept json
// @Produce json
// @Param chainId path string true "Chain ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/chains/{chainId}/status [get]
func GetChainStatus(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Get chain ID from path
	chainID := c.Params("chainId")
	if chainID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Chain ID is required")
	}
	
	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}
	
	// Get network status
	status, err := baasService.GetNetworkStatus(chainID)
	if err != nil {
		if err.Error() == fmt.Sprintf("network %s not configured", chainID) {
			return fiber.NewError(fiber.StatusNotFound, "Chain not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get chain status: "+err.Error())
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Chain status retrieved successfully",
		Data:    status,
	})
}

// GetCrossChainTransactions gets cross-chain transactions between chains
// @Summary Get cross-chain transactions
// @Description Get cross-chain transactions between chains
// @Tags interoperability
// @Accept json
// @Produce json
// @Param sourceChainId path string true "Source Chain ID"
// @Param destChainId path string true "Destination Chain ID"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset results"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/transactions/{sourceChainId}/{destChainId} [get]
func GetCrossChainTransactions(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Get chain IDs from path
	sourceChainID := c.Params("sourceChainId")
	destChainID := c.Params("destChainId")
	
	// Validate parameters
	if sourceChainID == "" || destChainID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Source and destination chain IDs are required")
	}
	
	// Get limit and offset from query params
	limitStr := c.Query("limit", "10")
	offsetStr := c.Query("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}
	
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}
	
	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}
	
	// Check if the bridge exists
	bridgeID := fmt.Sprintf("bridge_%s_%s", sourceChainID, destChainID)
	
	// Get bridge transactions
	transactions, err := baasService.GetBridgeTransactions(bridgeID, limit, offset)
	if err != nil {
		// Try the reverse direction if this bridge doesn't exist
		if strings.Contains(err.Error(), "not found") {
			bridgeID = fmt.Sprintf("bridge_%s_%s", destChainID, sourceChainID)
			transactions, err = baasService.GetBridgeTransactions(bridgeID, limit, offset)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to get cross-chain transactions: "+err.Error())
			}
		} else {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to get cross-chain transactions: "+err.Error())
		}
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Cross-chain transactions retrieved successfully",
		Data: map[string]interface{}{
			"source_chain_id":  sourceChainID,
			"dest_chain_id":    destChainID,
			"bridge_id":        bridgeID,
			"transactions":     transactions,
			"limit":            limit,
			"offset":           offset,
			"total_count":      len(transactions), // This should be the total count, not just the returned count
		},
	})
}

// CreateCrossChainBridge creates a cross-chain bridge
// @Summary Create cross-chain bridge
// @Description Create a cross-chain bridge between two chains
// @Tags interoperability
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Bridge creation details"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/bridges [post]
func CreateCrossChainBridge(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Parse request
	var req map[string]interface{}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate required fields
	sourceNetworkID, ok := req["source_network_id"].(string)
	if !ok || sourceNetworkID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "source_network_id is required")
	}
	
	targetNetworkID, ok := req["target_network_id"].(string)
	if !ok || targetNetworkID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "target_network_id is required")
	}
	
	bridgeType, ok := req["bridge_type"].(string)
	if !ok || bridgeType == "" {
		return fiber.NewError(fiber.StatusBadRequest, "bridge_type is required")
	}
	
	bridgeConfig, ok := req["bridge_config"].(map[string]interface{})
	if !ok {
		bridgeConfig = make(map[string]interface{})
	}
	
	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}
	
	// Create the cross-chain bridge
	bridgeID, err := baasService.CreateCrossChainBridge(
		sourceNetworkID,
		targetNetworkID,
		bridgeType,
		bridgeConfig,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create cross-chain bridge: "+err.Error())
	}
	
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Cross-chain bridge created successfully",
		Data: map[string]interface{}{
			"bridge_id":          bridgeID,
			"source_network_id":  sourceNetworkID,
			"target_network_id":  targetNetworkID,
			"bridge_type":        bridgeType,
		},
	})
}

// TransferAssetAcrossChains transfers an asset across chains
// @Summary Transfer asset across chains
// @Description Transfer an asset from one chain to another using a bridge
// @Tags interoperability
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Asset transfer details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/bridges/transfer [post]
func TransferAssetAcrossChains(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Parse request
	var req map[string]interface{}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate required fields
	sourceNetworkID, ok := req["source_network_id"].(string)
	if !ok || sourceNetworkID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "source_network_id is required")
	}
	
	targetNetworkID, ok := req["target_network_id"].(string)
	if !ok || targetNetworkID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "target_network_id is required")
	}
	
	bridgeID, ok := req["bridge_id"].(string)
	if !ok || bridgeID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "bridge_id is required")
	}
	
	assetID, ok := req["asset_id"].(string)
	if !ok || assetID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "asset_id is required")
	}
	
	amount, ok := req["amount"].(string)
	if !ok || amount == "" {
		return fiber.NewError(fiber.StatusBadRequest, "amount is required")
	}
	
	sender, ok := req["sender"].(string)
	if !ok || sender == "" {
		return fiber.NewError(fiber.StatusBadRequest, "sender is required")
	}
	
	recipient, ok := req["recipient"].(string)
	if !ok || recipient == "" {
		return fiber.NewError(fiber.StatusBadRequest, "recipient is required")
	}
	
	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}
	
	// Transfer the asset
	txHash, err := baasService.TransferAssetAcrossChains(
		sourceNetworkID,
		targetNetworkID,
		bridgeID,
		assetID,
		amount,
		sender,
		recipient,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to transfer asset: "+err.Error())
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Asset transfer initiated successfully",
		Data: map[string]interface{}{
			"tx_hash":            txHash,
			"source_network_id":  sourceNetworkID,
			"target_network_id":  targetNetworkID,
			"bridge_id":          bridgeID,
			"asset_id":           assetID,
			"amount":             amount,
			"sender":             sender,
			"recipient":          recipient,
			"status":             "pending",
		},
	})
}

// QueryIBCChannels queries IBC channels for a Cosmos chain
// @Summary Query IBC channels
// @Description Query IBC channels for a Cosmos chain
// @Tags interoperability
// @Accept json
// @Produce json
// @Param chainId path string true "Chain ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/ibc/channels/{chainId} [get]
func QueryIBCChannels(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Get chain ID from path
	chainID := c.Params("chainId")
	if chainID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Chain ID is required")
	}
	
	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}
	
	// Query IBC channels
	channels, err := baasService.QueryIBCChannels(chainID)
	if err != nil {
		if strings.Contains(err.Error(), "not configured") {
			return fiber.NewError(fiber.StatusNotFound, "Chain not found")
		}
		if strings.Contains(err.Error(), "does not support IBC") {
			return fiber.NewError(fiber.StatusBadRequest, "Chain does not support IBC")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query IBC channels: "+err.Error())
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "IBC channels retrieved successfully",
		Data: map[string]interface{}{
			"chain_id": chainID,
			"channels": channels,
		},
	})
}

// QueryXCMAssets queries XCM assets for a Polkadot chain
// @Summary Query XCM assets
// @Description Query XCM assets for a Polkadot chain
// @Tags interoperability
// @Accept json
// @Produce json
// @Param chainId path string true "Chain ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/xcm/assets/{chainId} [get]
func QueryXCMAssets(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Get chain ID from path
	chainID := c.Params("chainId")
	if chainID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Chain ID is required")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Check if a Polkadot bridge exists for the chain
	polkadotBridge, exists := blockchainClient.InteropClient.PolkadotBridges[chainID]
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "No Polkadot bridge found for the specified chain ID")
	}
	
	// Get registered assets
	assets := polkadotBridge.RegisteredAssets
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "XCM assets retrieved successfully",
		Data: map[string]interface{}{
			"chain_id": chainID,
			"assets":   assets,
		},
	})
}

// TraceIBCDenom traces an IBC token's origin
// @Summary Trace IBC token origin
// @Description Trace the origin of an IBC token
// @Tags interoperability
// @Accept json
// @Produce json
// @Param chainId path string true "Chain ID"
// @Param denom path string true "Token Denom"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/ibc/trace/{chainId}/{denom} [get]
func TraceIBCDenom(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Get parameters from path
	chainID := c.Params("chainId")
	denom := c.Params("denom")
	
	// Validate parameters
	if chainID == "" || denom == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Chain ID and denom are required")
	}
	
	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}
	
	// Trace IBC denom
	denomTrace, err := baasService.GetIBCDenomTrace(chainID, denom)
	if err != nil {
		if strings.Contains(err.Error(), "not configured") {
			return fiber.NewError(fiber.StatusNotFound, "Chain not found")
		}
		if strings.Contains(err.Error(), "does not support IBC") {
			return fiber.NewError(fiber.StatusBadRequest, "Chain does not support IBC")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to trace IBC denom: "+err.Error())
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "IBC denom trace retrieved successfully",
		Data: map[string]interface{}{
			"chain_id":    chainID,
			"denom":       denom,
			"denom_trace": denomTrace,
		},
	})
}

// TraceXCMAsset traces an XCM asset's origin
// @Summary Trace XCM asset origin
// @Description Trace the origin of an XCM asset
// @Tags interoperability
// @Accept json
// @Produce json
// @Param chainId path string true "Chain ID"
// @Param assetId path string true "Asset ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/xcm/trace/{chainId}/{assetId} [get]
func TraceXCMAsset(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Get parameters from path
	chainID := c.Params("chainId")
	assetID := c.Params("assetId")
	
	// Validate parameters
	if chainID == "" || assetID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Chain ID and asset ID are required")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Check if a Polkadot bridge exists for the chain
	polkadotBridge, exists := blockchainClient.InteropClient.PolkadotBridges[chainID]
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "No Polkadot bridge found for the specified chain ID")
	}
	
	// Trace XCM asset
	assetDetails, err := polkadotBridge.TraceXCMAsset(assetID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to trace XCM asset: "+err.Error())
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "XCM asset trace retrieved successfully",
		Data: map[string]interface{}{
			"chain_id":      chainID,
			"asset_id":      assetID,
			"asset_details": assetDetails,
		},
	})
}

// GetBridgeById gets details of a specific bridge
// @Summary Get bridge details
// @Description Get details of a specific bridge
// @Tags interoperability
// @Accept json
// @Produce json
// @Param bridgeId path string true "Bridge ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/bridges/{bridgeId} [get]
func GetBridgeById(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Get bridge ID from path
	bridgeID := c.Params("bridgeId")
	if bridgeID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Bridge ID is required")
	}
	
	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}
	
	// Get bridge details
	bridge, err := baasService.GetBridgeById(bridgeID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fiber.NewError(fiber.StatusNotFound, "Bridge not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get bridge details: "+err.Error())
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Bridge details retrieved successfully",
		Data:    bridge,
	})
}

// DeploySmartContract deploys a smart contract to a blockchain
// @Summary Deploy smart contract
// @Description Deploy a smart contract to a blockchain
// @Tags interoperability
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Contract deployment details"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/contracts [post]
func DeploySmartContract(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Parse request
	var req map[string]interface{}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate required fields
	networkID, ok := req["network_id"].(string)
	if !ok || networkID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "network_id is required")
	}
	
	contractType, ok := req["contract_type"].(string)
	if !ok || contractType == "" {
		return fiber.NewError(fiber.StatusBadRequest, "contract_type is required")
	}
	
	contractName, ok := req["contract_name"].(string)
	if !ok || contractName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "contract_name is required")
	}
	
	contractCode, ok := req["contract_code"].(string)
	if !ok || contractCode == "" {
		return fiber.NewError(fiber.StatusBadRequest, "contract_code is required")
	}
	
	initArgs, ok := req["init_args"].(map[string]interface{})
	if !ok {
		initArgs = make(map[string]interface{})
	}
	
	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}
	
	// Deploy the smart contract
	contractAddress, err := baasService.DeploySmartContract(
		networkID,
		contractType,
		contractName,
		contractCode,
		initArgs,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to deploy smart contract: "+err.Error())
	}
	
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Smart contract deployed successfully",
		Data: map[string]interface{}{
			"network_id":        networkID,
			"contract_type":     contractType,
			"contract_name":     contractName,
			"contract_address":  contractAddress,
		},
	})
}

// QueryContractState queries the state of a smart contract
// @Summary Query contract state
// @Description Query the state of a smart contract
// @Tags interoperability
// @Accept json
// @Produce json
// @Param networkId path string true "Network ID"
// @Param contractAddress path string true "Contract Address"
// @Param request body map[string]interface{} true "Query data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/contracts/{networkId}/{contractAddress}/query [post]
func QueryContractState(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Get parameters from path
	networkID := c.Params("networkId")
	contractAddress := c.Params("contractAddress")
	
	// Validate parameters
	if networkID == "" || contractAddress == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Network ID and contract address are required")
	}
	
	// Parse request
	var queryData map[string]interface{}
	if err := c.BodyParser(&queryData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}
	
	// Query the contract state
	state, err := baasService.QueryContractState(
		networkID,
		contractAddress,
		queryData,
	)
	if err != nil {
		if strings.Contains(err.Error(), "not configured") {
			return fiber.NewError(fiber.StatusNotFound, "Network not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query contract state: "+err.Error())
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Contract state queried successfully",
		Data: map[string]interface{}{
			"network_id":        networkID,
			"contract_address":  contractAddress,
			"query":             queryData,
			"result":            state,
		},
	})
}

// These are supplementary structs to support cross-chain account management

// InterChainAccountRequest represents a request to create an interchain account
type InterChainAccountRequest struct {
	SourceChainID  string `json:"source_chain_id"`
	TargetChainID  string `json:"target_chain_id"`
	ConnectionID   string `json:"connection_id"`
	Owner          string `json:"owner"`
}

// InterChainAccountTxRequest represents a request to send a transaction from an interchain account
type InterChainAccountTxRequest struct {
	SourceChainID  string                   `json:"source_chain_id"`
	TargetChainID  string                   `json:"target_chain_id"`
	ConnectionID   string                   `json:"connection_id"`
	Owner          string                   `json:"owner"`
	Messages       []map[string]interface{} `json:"messages"`
	Memo           string                   `json:"memo"`
}

// CreateInterChainAccount creates an interchain account
// @Summary Create interchain account
// @Description Create an interchain account for cross-chain interaction
// @Tags interoperability
// @Accept json
// @Produce json
// @Param request body InterChainAccountRequest true "Interchain account creation details"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/ibc/accounts [post]
func CreateInterChainAccount(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Parse request
	var req InterChainAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.SourceChainID == "" || req.TargetChainID == "" || req.ConnectionID == "" || req.Owner == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required fields")
	}
	
	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}
	
	// Create the interchain account
	accountAddress, err := baasService.CreateInterChainAccount(
		req.SourceChainID,
		req.TargetChainID,
		req.ConnectionID,
		req.Owner,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create interchain account: "+err.Error())
	}
	
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Interchain account created successfully",
		Data: map[string]interface{}{
			"source_chain_id":  req.SourceChainID,
			"target_chain_id":  req.TargetChainID,
			"connection_id":    req.ConnectionID,
			"owner":            req.Owner,
			"account_address":  accountAddress,
		},
	})
}

// SendInterChainAccountTx sends a transaction from an interchain account
// @Summary Send interchain account transaction
// @Description Send a transaction from an interchain account
// @Tags interoperability
// @Accept json
// @Produce json
// @Param request body InterChainAccountTxRequest true "Interchain account transaction details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interop/ibc/accounts/tx [post]
func SendInterChainAccountTx(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Parse request
	var req InterChainAccountTxRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.SourceChainID == "" || req.TargetChainID == "" || req.ConnectionID == "" || 
	   req.Owner == "" || len(req.Messages) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required fields")
	}
	
	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}
	
	// Send the interchain account transaction
	txHash, err := baasService.SendInterChainAccountTx(
		req.SourceChainID,
		req.TargetChainID,
		req.ConnectionID,
		req.Owner,
		req.Messages,
		req.Memo,
			)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to send interchain account transaction: "+err.Error())
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Interchain account transaction sent successfully",
		Data: map[string]interface{}{
			"source_chain_id":  req.SourceChainID,
			"target_chain_id":  req.TargetChainID,
			"tx_hash":          txHash,
			"status":           "pending",
		},
	})
}

// VerifyInteropTransaction verifies a cross-chain transaction
// @Summary Verify cross-chain transaction
// @Description Verify the status and integrity of a cross-chain transaction
// @Tags interoperability
// @Accept json
// @Produce json
// @Param tx_id query string true "Transaction ID"
// @Param source_chain_id query string true "Source Chain ID"
// @Param dest_chain_id query string true "Destination Chain ID" 
// @Param protocol query string false "Protocol (ibc, xcm, bridge)" 
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /interoperability/transactions/verify [get]
func VerifyInteropTransaction(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Check if interoperability is enabled
	if !cfg.InteropEnabled {
		return fiber.NewError(fiber.StatusBadRequest, "Interoperability is not enabled")
	}
	
	// Get query parameters
	txID := c.Query("tx_id")
	sourceChainID := c.Query("source_chain_id")
	destChainID := c.Query("dest_chain_id")
	protocol := c.Query("protocol", "auto") // Default to auto-detect
	
	// Validate parameters
	if txID == "" || sourceChainID == "" || destChainID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required query parameters")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for now
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Check cache first
	cacheKey := fmt.Sprintf("%s-%s-%s", txID, sourceChainID, destChainID)
	if cachedResult, found := blockchainClient.InteropClient.VerificationCache[cacheKey]; found {
		// Only use cache if less than 5 minutes old
		if time.Since(cachedResult.Timestamp) < 5*time.Minute {
			return c.JSON(SuccessResponse{
				Success: true,
				Message: "Transaction verification result (cached)",
				Data: map[string]interface{}{
					"tx_id": txID,
					"source_chain_id": sourceChainID,
					"destination_chain_id": destChainID,
					"verified": cachedResult.Verified,
					"proof_data": cachedResult.ProofData,
					"cached_at": cachedResult.Timestamp.Format(time.RFC3339),
				},
			})
		}
	}
	
	// Determine which verification method to use based on protocol
	var verified bool
	var proofData string
	var err error
	
	switch strings.ToLower(protocol) {
	case "ibc":
		verified, proofData, err = blockchainClient.InteropClient.VerifyIBCTransaction(txID, sourceChainID, destChainID)
	case "xcm":
		verified, proofData, err = blockchainClient.InteropClient.VerifyXCMTransaction(txID, sourceChainID, destChainID)
	case "bridge":
		verified, proofData, err = blockchainClient.InteropClient.VerifyBridgeTransaction(txID, sourceChainID, destChainID)
	default:
		// Auto-detect based on chain IDs
		if strings.Contains(strings.ToLower(sourceChainID), "cosmos") || 
		   strings.Contains(strings.ToLower(destChainID), "cosmos") {
			verified, proofData, err = blockchainClient.InteropClient.VerifyIBCTransaction(txID, sourceChainID, destChainID)
		} else if strings.Contains(strings.ToLower(sourceChainID), "dot") || 
				  strings.Contains(strings.ToLower(destChainID), "dot") {
			verified, proofData, err = blockchainClient.InteropClient.VerifyXCMTransaction(txID, sourceChainID, destChainID)
		} else {
			verified, proofData, err = blockchainClient.InteropClient.VerifyBridgeTransaction(txID, sourceChainID, destChainID)
		}
	}
	
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Transaction verification failed: "+err.Error())
	}
	
	// Cache result
	blockchainClient.InteropClient.VerificationCache[cacheKey] = blockchain.InteropVerificationResult{
		Verified:  verified,
		Timestamp: time.Now(),
		ProofData: proofData,
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Transaction verification completed",
		Data: map[string]interface{}{
			"tx_id": txID,
			"source_chain_id": sourceChainID,
			"destination_chain_id": destChainID,
			"verified": verified,
			"proof_data": proofData,
		},
	})
}