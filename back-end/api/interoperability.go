package api

import (
	"fmt"
	"strconv"
	"time"
	
	"github.com/gofiber/fiber/v2"
	"github.com/vietchain/tracepost-larvae/blockchain"
	"github.com/vietchain/tracepost-larvae/config"
	"github.com/vietchain/tracepost-larvae/db"
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

// GetBatchFromBlockchain returns batch data from blockchain
// @Summary Get batch from blockchain
// @Description Retrieve batch data directly from the blockchain
// @Tags blockchain
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /blockchain/batch/{batchId} [get]
func GetBatchFromBlockchain(c *fiber.Ctx) error {
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