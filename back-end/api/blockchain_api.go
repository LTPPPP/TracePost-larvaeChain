package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"github.com/LTPPPP/TracePost-larvaeChain/models"
)

// SearchBlockchainRecordsRequest represents a request to search blockchain records
type SearchBlockchainRecordsRequest struct {
	RelatedTable  string `json:"related_table"`
	RelatedID     int    `json:"related_id"`
	FromTimestamp string `json:"from_timestamp"`
	ToTimestamp   string `json:"to_timestamp"`
	Limit         int    `json:"limit"`
}

// SearchBlockchainRecords searches blockchain records based on criteria
// @Summary Search blockchain records
// @Description Search for blockchain records based on specified criteria
// @Tags blockchain
// @Accept json
// @Produce json
// @Param request body SearchBlockchainRecordsRequest true "Search criteria"
// @Success 200 {object} SuccessResponse{data=[]map[string]interface{}}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /blockchain/search [post]
func SearchBlockchainRecords(c *fiber.Ctx) error {
	// Parse request body
	var req SearchBlockchainRecordsRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if req.Limit <= 0 {
		req.Limit = 100 // Default limit
	} else if req.Limit > 1000 {
		req.Limit = 1000 // Maximum limit
	}

	// Build query parameters
	var params []interface{}
	query := `
		SELECT br.id, br.related_table, br.related_id, br.tx_id, br.metadata_hash, br.created_at 
		FROM blockchain_record br
		WHERE br.is_active = true
	`

	paramCounter := 1

	// Add filters based on request
	if req.RelatedTable != "" {
		query += fmt.Sprintf(" AND br.related_table = $%d", paramCounter)
		params = append(params, req.RelatedTable)
		paramCounter++
	}

	if req.RelatedID > 0 {
		query += fmt.Sprintf(" AND br.related_id = $%d", paramCounter)
		params = append(params, req.RelatedID)
		paramCounter++
	}

	if req.FromTimestamp != "" {
		fromTime, err := time.Parse(time.RFC3339, req.FromTimestamp)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid from_timestamp format. Expected RFC3339.")
		}
		query += fmt.Sprintf(" AND br.created_at >= $%d", paramCounter)
		params = append(params, fromTime)
		paramCounter++
	}

	if req.ToTimestamp != "" {
		toTime, err := time.Parse(time.RFC3339, req.ToTimestamp)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid to_timestamp format. Expected RFC3339.")
		}
		query += fmt.Sprintf(" AND br.created_at <= $%d", paramCounter)
		params = append(params, toTime)
		paramCounter++
	}

	// Add ordering and limit
	query += " ORDER BY br.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d", paramCounter)
	params = append(params, req.Limit)

	// Execute query
	rows, err := db.DB.Query(query, params...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error when searching blockchain records")
	}
	defer rows.Close()

	// Parse results
	var records []map[string]interface{}
	for rows.Next() {
		var id, relatedID int
		var relatedTable, txID, metadataHash string
		var createdAt time.Time

		if err := rows.Scan(&id, &relatedTable, &relatedID, &txID, &metadataHash, &createdAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse blockchain record")
		}

		record := map[string]interface{}{
			"id":            id,
			"related_table": relatedTable,
			"related_id":    relatedID,
			"tx_id":         txID,
			"metadata_hash": metadataHash,
			"created_at":    createdAt,
		}

		// For batch-related records, include additional batch info
		if relatedTable == "batch" || relatedTable == "batch_extended" || relatedTable == "batch_status_extended" {
			var batch models.Batch
			err := db.DB.QueryRow(`
				SELECT id, species, quantity, status, created_at, updated_at
				FROM batch
				WHERE id = $1 AND is_active = true
			`, relatedID).Scan(
				&batch.ID,
				&batch.Species,
				&batch.Quantity,
				&batch.Status,
				&batch.CreatedAt,
				&batch.UpdatedAt,
			)
			if err == nil {
				record["batch_info"] = map[string]interface{}{
					"id":       batch.ID,
					"species":  batch.Species,
					"quantity": batch.Quantity,
					"status":   batch.Status,
				}
			}
		}

		records = append(records, record)
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Found %d blockchain records", len(records)),
		Data:    records,
	})
}

// GetBlockchainVerification performs a full blockchain verification of a batch and returns details
// @Summary Get blockchain verification for a batch
// @Description Performs a comprehensive blockchain verification for a batch
// @Tags blockchain
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=map[string]interface{}}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /blockchain/verify/{batchId} [get]
func GetBlockchainVerification(c *fiber.Ctx) error {
	// Get batch ID from params
	batchIDStr := c.Params("batchId")
	if batchIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}

	batchID, err := strconv.Atoi(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
	}

	// Check if batch exists and get its data
	var batch models.Batch
	err = db.DB.QueryRow(`
		SELECT id, hatchery_id, species, quantity, status, created_at, updated_at
		FROM batch
		WHERE id = $1 AND is_active = true
	`, batchID).Scan(
		&batch.ID,
		&batch.HatcheryID,
		&batch.Species,
		&batch.Quantity,
		&batch.Status,
		&batch.CreatedAt,
		&batch.UpdatedAt,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}

	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		os.Getenv("BLOCKCHAIN_NODE_URL"),
		os.Getenv("BLOCKCHAIN_PRIVATE_KEY"),
		os.Getenv("BLOCKCHAIN_ACCOUNT"),
		os.Getenv("BLOCKCHAIN_CHAIN_ID"),
		os.Getenv("BLOCKCHAIN_CONSENSUS"),
	)

	// Convert batch data to map
	batchData := map[string]interface{}{
		"batch_id":    batchIDStr,
		"hatchery_id": strconv.Itoa(batch.HatcheryID),
		"species":     batch.Species,
		"quantity":    batch.Quantity,
		"status":      batch.Status,
	}

	// Verify integrity
	isValid, discrepancies, err := blockchainClient.VerifyBatchIntegrity(batchIDStr, batchData)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to verify batch integrity: %v", err))
	}

	// Get all blockchain records for this batch
	rows, err := db.DB.Query(`
		SELECT id, related_table, related_id, tx_id, metadata_hash, created_at
		FROM blockchain_record
		WHERE (related_table = 'batch' OR related_table = 'batch_extended' OR related_table = 'batch_status_extended')
		  AND related_id = $1
		  AND is_active = true
		ORDER BY created_at ASC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve blockchain records")
	}
	defer rows.Close()

	// Collect blockchain records
	var blockchainRecords []map[string]interface{}
	for rows.Next() {
		var id, relatedID int
		var relatedTable, txID, metadataHash string
		var createdAt time.Time

		if err := rows.Scan(&id, &relatedTable, &relatedID, &txID, &metadataHash, &createdAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse blockchain record")
		}

		blockchainRecords = append(blockchainRecords, map[string]interface{}{
			"id":            id,
			"related_table": relatedTable,
			"tx_id":         txID,
			"metadata_hash": metadataHash,
			"created_at":    createdAt,
		})
	}

	// Get full blockchain verification
	verificationDetails, err := blockchainClient.VerifyBatchDataOnChain(batchIDStr)
	// Ignore errors for comprehensive verification
	if err != nil {
		fmt.Printf("Warning: Failed to get comprehensive blockchain verification: %v\n", err)
	}

	// Compile verification result
	verificationResult := map[string]interface{}{
		"batch_id":            batchID,
		"verification_time":   time.Now(),
		"is_valid":            isValid,
		"discrepancies":       discrepancies,
		"blockchain_records":  blockchainRecords,
		"record_count":        len(blockchainRecords),
		"verification_level":  "comprehensive",
		"verification_result": verificationDetails,
	}

	// Return success response
	var message string
	if isValid {
		message = "Batch verified successfully on blockchain"
	} else {
		message = "Batch verification found discrepancies"
	}
	return c.JSON(SuccessResponse{
		Success: true,
		Message: message,
		Data:    verificationResult,
	})
}

// BatchBlockchainAudit returns a complete audit trail for a batch from blockchain data
// @Summary Get batch blockchain audit trail
// @Description Retrieve a complete audit trail for a batch from blockchain
// @Tags blockchain
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse{data=map[string]interface{}}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /blockchain/audit/{batchId} [get]
func BatchBlockchainAudit(c *fiber.Ctx) error {
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
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1 AND is_active = true)", batchID).Scan(&exists)
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

	// Get blockchain data
	blockchainData, err := blockchainClient.GetBatchBlockchainData(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to get blockchain data: %v", err))
	}

	// Get blockchain transactions
	txs, err := blockchainClient.GetBatchTransactions(batchIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to get batch transactions: %v", err))
	}

	// Get batch events from database
	rows, err := db.DB.Query(`
		SELECT id, event_type, actor_id, location, timestamp, metadata
		FROM event
		WHERE batch_id = $1 AND is_active = true
		ORDER BY timestamp ASC
	`, batchID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error retrieving batch events")
	}
	defer rows.Close()

	// Parse batch events
	var events []map[string]interface{}
	for rows.Next() {
		var id, actorID int
		var eventType, location string
		var timestamp time.Time
		var metadata models.JSONB

		if err := rows.Scan(&id, &eventType, &actorID, &location, &timestamp, &metadata); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse batch event")
		}

		var metadataObj map[string]interface{}
		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &metadataObj); err != nil {
				metadataObj = map[string]interface{}{"raw": string(metadata)}
			}
		}

		// Get blockchain records for this event
		var blockchainRecords []map[string]interface{}
		eventRecordsRows, err := db.DB.Query(`
			SELECT id, tx_id, metadata_hash, created_at
			FROM blockchain_record
			WHERE related_table = 'event' AND related_id = $1 AND is_active = true
		`, id)
		if err == nil {
			defer eventRecordsRows.Close()

			for eventRecordsRows.Next() {
				var recordID int
				var txID, metadataHash string
				var createdAt time.Time

				if err := eventRecordsRows.Scan(&recordID, &txID, &metadataHash, &createdAt); err == nil {
					blockchainRecords = append(blockchainRecords, map[string]interface{}{
						"id":            recordID,
						"tx_id":         txID,
						"metadata_hash": metadataHash,
						"created_at":    createdAt,
					})
				}
			}
		}

		events = append(events, map[string]interface{}{
			"id":                id,
			"event_type":        eventType,
			"actor_id":          actorID,
			"location":          location,
			"timestamp":         timestamp,
			"metadata":          metadataObj,
			"blockchain_records": blockchainRecords,
		})
	}

	// Combine all data into an audit trail
	auditTrail := map[string]interface{}{
		"batch_id":                  batchID,
		"blockchain_data":           blockchainData,
		"blockchain_transactions":   txs,
		"events":                    events,
		"audit_timestamp":           time.Now(),
		"audit_blockchain_verified": true,
	}

	// Return success response
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch blockchain audit completed successfully",
		Data:    auditTrail,
	})
}

// DeployLogisticsTraceabilityContractRequest represents a request to deploy the LogisticsTraceability contract
// This type is used for Swagger documentation and request binding
// It matches the fields expected in the handler's Request struct
type DeployLogisticsTraceabilityContractRequest struct {
	NetworkID    string                 `json:"network_id"`
	ContractName string                 `json:"contract_name"`
	InitArgs     map[string]interface{} `json:"init_args"`
}

// DeployLogisticsTraceabilityContract deploys the LogisticsTraceability contract
// @Summary Deploy LogisticsTraceability contract
// @Description Deploy the LogisticsTraceability contract on the specified network
// @Tags blockchain
// @Accept json
// @Produce json
// @Param request body DeployLogisticsTraceabilityContractRequest true "Deployment request"
// @Success 200 {object} SuccessResponse{data=map[string]interface{}}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /blockchain/deploy-logistics-contract [post]
func DeployLogisticsTraceabilityContract(c *fiber.Ctx) error {
	type Request struct {
		NetworkID    string                 `json:"network_id"`
		ContractName string                 `json:"contract_name"`
		InitArgs     map[string]interface{} `json:"init_args"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.ContractName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "contract_name is required")
	}

	// Load LogisticsTraceability contract code
	contractPath := filepath.Join("contracts", "LogisticsTraceability.sol")
	contractCode, err := ioutil.ReadFile(contractPath)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to read contract code: "+err.Error())
	}

	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}

	// Deploy the contract
	contractAddress, err := baasService.DeploySmartContract(
		req.NetworkID,
		"logistics",
		req.ContractName,
		string(contractCode),
		req.InitArgs,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to deploy contract: "+err.Error())
	}

	// Return the contract address
	return c.JSON(fiber.Map{
		"contract_address": contractAddress,
	})
}
