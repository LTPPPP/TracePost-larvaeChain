package api

import (
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