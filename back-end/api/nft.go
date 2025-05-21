package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"github.com/LTPPPP/TracePost-larvaeChain/ipfs"
	"github.com/skip2/go-qrcode"
)

// NFTContractDeployRequest represents a request to deploy an NFT contract
type NFTContractDeployRequest struct {
	NetworkID        string                 `json:"network_id"`
	ContractName     string                 `json:"contract_name"`
	ContractSymbol   string                 `json:"contract_symbol"`
	LogisticsAddress string                 `json:"logistics_address,omitempty"`
	InitArgs         map[string]interface{} `json:"init_args,omitempty"`
}

// TokenizeBatchRequest represents a request to tokenize a batch as an NFT
type TokenizeBatchRequest struct {
	BatchID          string `json:"batch_id"`
	NetworkID        string `json:"network_id"`
	ContractAddress  string `json:"contract_address"`
	RecipientAddress string `json:"recipient_address"`
	TransferID       string `json:"transfer_id,omitempty"` // Optional transfer ID to associate with NFT
}

// TransferNFTRequest represents a request to transfer an NFT to a new owner
type TransferNFTRequest struct {
	ContractAddress string `json:"contract_address"`
	NetworkID       string `json:"network_id"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
}

// TokenizeTransactionRequest represents a request to tokenize a transaction as an NFT
type TokenizeTransactionRequest struct {
	TransferID       string                 `json:"transfer_id" binding:"required"`     // Required shipment transfer ID
	NetworkID        string                 `json:"network_id" binding:"required"`      // Blockchain network ID
	ContractAddress  string                 `json:"contract_address" binding:"required"` // NFT contract address
	RecipientAddress string                 `json:"recipient_address" binding:"required"` // Address to receive the NFT
	Metadata         map[string]interface{} `json:"metadata"`                           // Additional metadata for the NFT
}

// TransactionNFTResponse represents the response when querying a transaction NFT
type TransactionNFTResponse struct {
	TokenID         string                 `json:"token_id"`
	ContractAddress string                 `json:"contract_address"`
	TransferID      string                 `json:"transfer_id"`
	BatchID         string                 `json:"batch_id"`
	OwnerAddress    string                 `json:"owner_address"`
	TokenURI        string                 `json:"token_uri"`
	QRCodeURL       string                 `json:"qr_code_url"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
}

// TransactionTraceResponse represents the response when tracing a transaction
type TransactionTraceResponse struct {
	TransferID    string                   `json:"transfer_id"`
	BatchID       string                   `json:"batch_id"`
	TokenID       string                   `json:"token_id"`
	History       []map[string]interface{} `json:"history"`
	CurrentOwner  string                   `json:"current_owner"`
	IsVerified    bool                     `json:"is_verified"`
	VerifiedData  map[string]interface{}   `json:"verified_data,omitempty"`
}

// Declare variables for database query results
var (
	batchID   string
	recipient string
	tokenURI  string
	createdAt time.Time
)

// DeployNFTContract deploys an NFT contract for batch traceability
// @Summary Deploy NFT contract
// @Description Deploy a new NFT contract for batch tokenization
// @Tags nft
// @Accept json
// @Produce json
// @Param request body NFTContractDeployRequest true "NFT contract deployment details"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /nft/contracts [post]
func DeployNFTContract(c *fiber.Ctx) error {
	// Parse request
	var req NFTContractDeployRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate required fields
	if req.NetworkID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "network_id is required")
	}
	
	if req.ContractName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "contract_name is required")
	}
	
	if req.ContractSymbol == "" {
		return fiber.NewError(fiber.StatusBadRequest, "contract_symbol is required")
	}
	
	// Load NFT contract code
	contractPath := filepath.Join("contracts", "LogisticsTraceabilityNFT.sol")
	contractCode, err := ioutil.ReadFile(contractPath)
	if (err != nil) {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to read NFT contract code: "+err.Error())
	}
	
	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}
	
	// Prepare initialization arguments
	if req.InitArgs == nil {
		req.InitArgs = make(map[string]interface{})
	}
	
	req.InitArgs["name"] = req.ContractName
	req.InitArgs["symbol"] = req.ContractSymbol
	req.InitArgs["logistics_contract"] = req.LogisticsAddress
	
	// Deploy the NFT contract
	contractAddress, err := baasService.DeploySmartContract(
		req.NetworkID,
		"nft",
		"LogisticsTraceabilityNFT",
		string(contractCode),
		req.InitArgs,
	)
	
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to deploy NFT contract: "+err.Error())
	}
	
	// Record contract deployment in the database
	db.DB.Exec(`
		INSERT INTO contract_deployments (
			network_id, contract_type, contract_name, contract_address, created_at
		) VALUES (
			$1, $2, $3, $4, $5
		)
	`, req.NetworkID, "nft", req.ContractName, contractAddress, time.Now())
	
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "NFT contract deployed successfully",
		Data: map[string]interface{}{
			"network_id":       req.NetworkID,
			"contract_type":    "nft",
			"contract_name":    req.ContractName,
			"contract_symbol":  req.ContractSymbol,
			"contract_address": contractAddress,
		},
	})
}

// TokenizeBatch mints an NFT for a specific batch
// @Summary Tokenize batch
// @Description Create an NFT token representing a batch
// @Tags nft
// @Accept json
// @Produce json
// @Param request body TokenizeBatchRequest true "Batch tokenization details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /nft/batches/tokenize [post]
func TokenizeBatch(c *fiber.Ctx) error {
	// Parse request
	var req TokenizeBatchRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate required fields
	if req.BatchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "batch_id is required")
	}
	
	if req.NetworkID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "network_id is required")
	}
	
	if req.ContractAddress == "" {
		return fiber.NewError(fiber.StatusBadRequest, "contract_address is required")
	}
	
	if req.RecipientAddress == "" {
		return fiber.NewError(fiber.StatusBadRequest, "recipient_address is required")
	}
	
	// Check if batch exists in database
	var batchExists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1)", req.BatchID).Scan(&batchExists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	
	if !batchExists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}
	
	// Get batch details to include in metadata
	var species, hatcheryID string
	var createdAt time.Time
	err = db.DB.QueryRow(`
		SELECT species, hatchery_id, created_at 
		FROM batch 
		WHERE id = $1
	`, req.BatchID).Scan(&species, &hatcheryID, &createdAt)
	
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve batch details: "+err.Error())
	}
	
	// Check if a hatchery with the ID exists
	var hatcheryName, location string
	err = db.DB.QueryRow(`
		SELECT name, location
		FROM hatchery
		WHERE id = $1
	`, hatcheryID).Scan(&hatcheryName, &location)
	
	if err != nil {
		hatcheryName = "Unknown"
		location = "Unknown"
	}
		// Generate QR code URL for this batch
	qrCodeURL := "https://trace.viechain.com/api/v1/batches/" + req.BatchID + "/qr"
	
	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}
	
	// Get transfer information if a transfer ID is provided
	var transferInfo map[string]interface{}
	if req.TransferID != "" {
		// Check if the transfer exists and relates to this batch
		var transferExists bool
		batchIDInt, err := strconv.Atoi(req.BatchID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid batch ID format")
		}
		
		err = db.DB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM shipment_transfer 
				WHERE id = $1 AND batch_id = $2
			)
		`, req.TransferID, batchIDInt).Scan(&transferExists)
		
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Database error checking transfer: "+err.Error())
		}
		
		if !transferExists {
			return fiber.NewError(fiber.StatusBadRequest, "Transfer ID does not exist or is not associated with this batch")
		}
		
		// Get transfer details to include in the token metadata
		var sourceType, destinationID, destinationType, status string
		var quantity int
		var transferredAt time.Time
		
		err = db.DB.QueryRow(`
			SELECT source_type, destination_id, destination_type, 
				   quantity, transferred_at, status
			FROM shipment_transfer
			WHERE id = $1
		`, req.TransferID).Scan(
			&sourceType,
			&destinationID,
			&destinationType,
			&quantity,
			&transferredAt,
			&status,
		)
		
		if err == nil {
			transferInfo = map[string]interface{}{
				"transfer_id":       req.TransferID,
				"source":            fmt.Sprintf("%s (%s)", sourceType),
				"destination":       fmt.Sprintf("%s (%s)", destinationID, destinationType),
				"quantity":          quantity,
				"transferred_at":    transferredAt.Format(time.RFC3339),
				"status":            status,
			}
			
			// Use transfer verification URL instead
			qrCodeURL = fmt.Sprintf("https://trace.viechain.com/api/v1/shipments/transfers/%s/qr", req.TransferID)
		}
	}
		// Prepare the contract call
	contractMethods := map[string]interface{}{
		"method": "mintBatchNFT",
		"params": []interface{}{
			req.BatchID,
			req.RecipientAddress,
			"", // Will be overridden with generated URI below
		},
	}
	
	// Add additional metadata for token URI generation
	metadataParams := []interface{}{
		req.BatchID,
		species,
		location,
		createdAt.Unix(),
		qrCodeURL,
	}
	
	// Add transfer info to metadata if available
	if transferInfo != nil {
		metadataParams = append(metadataParams, transferInfo)
	}
	
	// First generate the token URI using the contract's generateTokenURI method
	tokenURIResult, err := baasService.QueryContractState(
    req.NetworkID,
    req.ContractAddress,
    map[string]interface{}{
        "method": "generateTokenURI",
        "params": metadataParams,
    	},
	)
	
	tokenURI, ok := tokenURIResult["result"].(string)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "Invalid token URI format")
	}
	
	// Update the method params with the token URI
	params := contractMethods["params"].([]interface{})
	params[2] = tokenURI
	contractMethods["params"] = params
	
	// Make the contract call to mint the NFT
	result, err := baasService.CallContractMethod(
		req.NetworkID,
		req.ContractAddress,
		contractMethods,
	)
	
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to tokenize batch: "+err.Error())
	}
	
	// Get the token ID from the result
	tokenID, ok := result["token_id"].(float64)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "Invalid token ID in response")
	}
		// Record the NFT in the database
	_, err = db.DB.Exec(`
		INSERT INTO batch_nft (
			batch_id, network_id, contract_address, token_id, recipient, token_uri, transfer_id, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`, req.BatchID, req.NetworkID, req.ContractAddress, int(tokenID), req.RecipientAddress, tokenURI, req.TransferID, time.Now())
	
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record NFT in database: "+err.Error())
	}
	
	// Update the batch record to mark it as tokenized
	_, err = db.DB.Exec(`
		UPDATE batch 
		SET is_tokenized = true, nft_token_id = $1, nft_contract = $2
		WHERE id = $3
	`, int(tokenID), req.ContractAddress, req.BatchID)
	
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update batch record: "+err.Error())
	}
	
	// If this was associated with a transfer, update the transfer record too
	if req.TransferID != "" {
		_, err = db.DB.Exec(`
			UPDATE shipment_transfer 
			SET nft_token_id = $1, nft_contract_address = $2
			WHERE id = $3
		`, int(tokenID), req.ContractAddress, req.TransferID)
		
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to update transfer record: "+err.Error())
		}
	}
		return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch successfully tokenized as NFT",
		Data: map[string]interface{}{
			"batch_id":         req.BatchID,
			"token_id":         int(tokenID),
			"network_id":       req.NetworkID,
			"contract_address": req.ContractAddress,
			"recipient":        req.RecipientAddress,
			"token_uri":        tokenURI,
			"transfer_id":      req.TransferID,
			"transfer_info":    transferInfo,
			"verification_url": qrCodeURL,
		},
	})
}

// GetBatchNFTDetails returns NFT details for a batch
// @Summary Get batch NFT details
// @Description Retrieve NFT details for a tokenized batch
// @Tags nft
// @Accept json
// @Produce json
// @Param batchId path string true "Batch ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /nft/batches/{batchId} [get]
func GetBatchNFTDetails(c *fiber.Ctx) error {
	batchID := c.Params("batchId")
	if batchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Batch ID is required")
	}
	
	// Check if batch exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1)", batchID).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Batch not found")
	}
	
	// Get NFT details for the batch
	var isTokenized bool
	var nftTokenID sql.NullInt64
	var nftContract sql.NullString
	var tokenURI sql.NullString
	var recipient sql.NullString
	var createdAt sql.NullTime
	var networkID sql.NullString
	
	err = db.DB.QueryRow(`
		SELECT b.is_tokenized, b.nft_token_id, b.nft_contract, 
		       n.token_uri, n.recipient, n.created_at, n.network_id
		FROM batch b
		LEFT JOIN batch_nft n ON b.id = n.batch_id
		WHERE b.id = $1
	`, batchID).Scan(
		&isTokenized, 
		&nftTokenID, 
		&nftContract, 
		&tokenURI, 
		&recipient, 
		&createdAt, 
		&networkID,
	)
	
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve batch NFT details")
	}
	
	// If not tokenized, return basic info
	if !isTokenized {
		return c.JSON(SuccessResponse{
			Success: true,
			Message: "Batch NFT details retrieved successfully",
			Data: map[string]interface{}{
				"batch_id":     batchID,
				"is_tokenized": false,
			},
		})
	}
	
	// Get batch detailed information
	var species, status, hatcheryID string
	var batchCreatedAt time.Time
	
	err = db.DB.QueryRow(`
		SELECT species, status, hatchery_id, created_at
		FROM batch
		WHERE id = $1
	`, batchID).Scan(&species, &status, &hatcheryID, &batchCreatedAt)
	
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve batch details")
	}
	
	// Get current owner of the NFT if available
	var currentOwner string
	
	if nftContract.Valid && nftTokenID.Valid && networkID.Valid {
				// Initialize BaaS service
		baasService := blockchain.NewBaaSService()
		if baasService != nil {
			// Query the NFT contract for the current owner
			ownerResult, err := baasService.QueryContractState(
				networkID.String,
				nftContract.String,
				map[string]interface{}{
					"method": "ownerOf",
					"params": []interface{}{nftTokenID.Int64},
				},
			)
			
			if err == nil {
				if owner, ok := ownerResult["result"].(string); ok {
					currentOwner = owner
				}
			}
		}
	}
	
	// If no owner found from blockchain, use the recipient from database
	if currentOwner == "" && recipient.Valid {
		currentOwner = recipient.String
	}
	
	// Generate NFT metadata
	nftMetadata := map[string]interface{}{
		"batch_id":        batchID,
		"is_tokenized":    true,
		"token_id":        nftTokenID.Int64,
		"contract":        nftContract.String,
		"token_uri":       tokenURI.String,
		"creator":         recipient.String,
		"current_owner":   currentOwner,
		"created_at":      createdAt.Time.Format(time.RFC3339),
		"network_id":      networkID.String,
		"species":         species,
		"status":          status,
		"hatchery_id":     hatcheryID,
		"batch_created_at": batchCreatedAt.Format(time.RFC3339),
		"marketplace_url": fmt.Sprintf("https://marketplace.viechain.com/token/%s/%d", 
			nftContract.String, nftTokenID.Int64),
	}
	
	// Generate QR code URL for verification
	nftMetadata["verification_url"] = fmt.Sprintf("https://trace.viechain.com/verify/nft/%s/%d", 
		nftContract.String, nftTokenID.Int64)
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch NFT details retrieved successfully",
		Data:    nftMetadata,
	})
}

// GetNFTDetails returns details of an NFT by token ID
// @Summary Get NFT details
// @Description Retrieve details of an NFT by token ID
// @Tags nft
// @Accept json
// @Produce json
// @Param tokenId path string true "Token ID"
// @Param contract query string true "Contract address"
// @Param network query string true "Network ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /nft/tokens/{tokenId} [get]
func GetNFTDetails(c *fiber.Ctx) error {
	tokenID := c.Params("tokenId")
	if tokenID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Token ID is required")
	}
	
	contractAddress := c.Query("contract")
	if contractAddress == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Contract address is required")
	}
	
	networkID := c.Query("network")
	if networkID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Network ID is required")
	}
	
	// Convert token ID to integer
	tokenIDInt, err := strconv.ParseInt(tokenID, 10, 64)
	err = db.DB.QueryRow(`
		SELECT batch_id, recipient, token_uri, created_at
		FROM batch_nft
		WHERE token_id = $1 AND contract_address = $2 AND network_id = $3
	`, tokenIDInt, contractAddress, networkID).Scan(&batchID, &recipient, &tokenURI, &createdAt)
	
	if err != nil && err != sql.ErrNoRows {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	
	// Initialize BaaS service to query the blockchain
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize blockchain service")
	}
	
	// Query the token owner from the contract
	ownerResult, err := baasService.QueryContractState(
		networkID,
		contractAddress,
		map[string]interface{}{
			"method": "ownerOf",
			"params": []interface{}{tokenIDInt},
		},
	)
	
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query token owner: "+err.Error())
	}
	
	owner, ok := ownerResult["result"].(string)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "Invalid token owner format")
	}
	
	// Query token URI if not available from database
	var nftTokenURI string
	if (tokenURI != "") {
		nftTokenURI = tokenURI
	} else {
		tokenURIResult, err := baasService.QueryContractState(
			networkID,
			contractAddress,
			map[string]interface{}{
				"method": "tokenURI",
				"params": []interface{}{tokenIDInt},
			},
		)
		
		if err == nil {
			if uri, ok := tokenURIResult["result"].(string); ok {
				nftTokenURI = uri
			}
		}
	}
	
	// Query batch ID if not available from database
	var nftBatchID string
	if batchID != "" {
		nftBatchID = batchID
	} else {
		batchIDResult, err := baasService.QueryContractState(
			networkID,
			contractAddress,
			map[string]interface{}{
				"method": "getBatchId",
				"params": []interface{}{tokenIDInt},
			},
		)
		
		if err == nil {
			if id, ok := batchIDResult["result"].(string); ok {
				nftBatchID = id
			}
		}
	}
	
	// If batch ID is available, get batch details from database
	var batchDetails map[string]interface{}
	if nftBatchID != "" {
		var species, status, hatcheryID string
		var batchCreatedAt time.Time
		
		err = db.DB.QueryRow(`
			SELECT species, status, hatchery_id, created_at
			FROM batch
			WHERE id = $1
		`, nftBatchID).Scan(&species, &status, &hatcheryID, &batchCreatedAt)
		
		if err == nil {
			batchDetails = map[string]interface{}{
				"batch_id":      nftBatchID,
				"species":       species,
				"status":        status,
				"hatchery_id":   hatcheryID,
				"created_at":    batchCreatedAt.Format(time.RFC3339),
			}
		}
	}
	
	// Construct NFT details response
	nftDetails := map[string]interface{}{
		"token_id":       tokenIDInt,
		"contract":       contractAddress,
		"network_id":     networkID,
		"owner":          owner,
		"creator":        recipient,
		"token_uri":      nftTokenURI,
		"created_at":     createdAt.Format(time.RFC3339),
		"marketplace_url": fmt.Sprintf("https://marketplace.viechain.com/token/%s/%d", 
			contractAddress, tokenIDInt),
	}
	
	// Add batch details if available
	if batchDetails != nil {
		nftDetails["batch"] = batchDetails
	} else if nftBatchID != "" {
		nftDetails["batch_id"] = nftBatchID
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "NFT details retrieved successfully",
		Data:    nftDetails,
	})
}

// TransferNFT transfers an NFT to a new owner
// @Summary Transfer NFT
// @Description Transfer an NFT to a new owner
// @Tags nft
// @Accept json
// @Produce json
// @Param tokenId path string true "Token ID"
// @Param request body TransferNFTRequest true "Transfer details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /nft/tokens/{tokenId}/transfer [put]
func TransferNFT(c *fiber.Ctx) error {
	// Get token ID from path
	tokenID := c.Params("tokenId")
	if tokenID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Token ID is required")
	}
	
	// Parse request
	var req TransferNFTRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate required fields
	if req.NetworkID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "network_id is required")
	}
	
	if req.ContractAddress == "" {
		return fiber.NewError(fiber.StatusBadRequest, "contract_address is required")
	}
	
	if req.ToAddress == "" {
		return fiber.NewError(fiber.StatusBadRequest, "to_address is required")
	}
	
	// Convert token ID to integer
	tokenIDInt, err := strconv.ParseInt(tokenID, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid token ID format")
	}
	
	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}
	
	// Get current owner if not provided in request
	fromAddress := req.FromAddress
	if fromAddress == "" {
		ownerResult, err := baasService.QueryContractState(
			req.NetworkID,
			req.ContractAddress,
			map[string]interface{}{
				"method": "ownerOf",
				"params": []interface{}{tokenIDInt},
			},
		)
		
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to query token owner: "+err.Error())
		}
		
		owner, ok := ownerResult["result"].(string)
		if !ok {
			return fiber.NewError(fiber.StatusInternalServerError, "Invalid token owner format")
		}
		
		fromAddress = owner
	}
	
	// Get batch ID for this token
	batchIDResult, err := baasService.QueryContractState(
		req.NetworkID,
		req.ContractAddress,
		map[string]interface{}{
			"method": "getBatchId",
			"params": []interface{}{tokenIDInt},
		},
	)
	
	var batchID string
	if err == nil {
		if id, ok := batchIDResult["result"].(string); ok {
			batchID = id
		}
	}
	
	// Prepare the contract call to transfer the token
	contractMethods := map[string]interface{}{
		"method": "transferBatch",
		"params": []interface{}{
			req.ToAddress,
			batchID,
		},
	}
	
	// Execute the transfer
	result, err := baasService.CallContractMethod(
		req.NetworkID,
		req.ContractAddress,
		contractMethods,
	)
	
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to transfer NFT: "+err.Error())
	}
	
	// Record the transfer in the database
	_, err = db.DB.Exec(`
		INSERT INTO nft_transfers (
			token_id, contract_address, network_id, from_address, to_address, transferred_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`, tokenIDInt, req.ContractAddress, req.NetworkID, fromAddress, req.ToAddress, time.Now())
	
	if err != nil {
		// Log the error but continue as the blockchain transfer was successful
		fmt.Printf("Failed to record NFT transfer in database: %v\n", err)
	}
	
	// Update batch ownership in the database if we have a batch ID
	if batchID != "" {
		_, err = db.DB.Exec(`
			UPDATE batch_nft
			SET owner = $1, updated_at = $2
			WHERE batch_id = $3 AND contract_address = $4
		`, req.ToAddress, time.Now(), batchID, req.ContractAddress)
		
		if err != nil {
			fmt.Printf("Failed to update batch ownership in database: %v\n", err)
		}
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: "NFT transferred successfully",
		Data: map[string]interface{}{
			"token_id":        tokenIDInt,
			"contract":        req.ContractAddress,
			"network_id":      req.NetworkID,
			"from":            fromAddress,
			"to":              req.ToAddress,
			"transferred_at":  time.Now().Format(time.RFC3339),
			"batch_id":        batchID,
			"transaction_hash": result["tx_hash"],
		},
	})
}

// TokenizeTransaction creates an NFT for a specific transaction/shipment transfer
// @Summary Tokenize transaction
// @Description Create an NFT token representing a transaction in the supply chain
// @Tags nft
// @Accept json
// @Produce json
// @Param request body TokenizeTransactionRequest true "Transaction tokenization details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /nft/transactions/tokenize [post]
func TokenizeTransaction(c *fiber.Ctx) error {
	// Parse request
	var req TokenizeTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate required fields
	if req.TransferID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "transfer_id is required")
	}
	
	if req.NetworkID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "network_id is required")
	}
	
	if req.ContractAddress == "" {
		return fiber.NewError(fiber.StatusBadRequest, "contract_address is required")
	}
	
	if req.RecipientAddress == "" {
		return fiber.NewError(fiber.StatusBadRequest, "recipient_address is required")
	}
	
	// Check if transfer exists in database
	var transferExists bool
	var batchID, sourceType, destinationID, destinationType, status string
	var transferredAt time.Time
	var quantity int
	
	err := db.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM shipment_transfer WHERE id = $1),
		       batch_id, source_type, destination_id, 
			   destination_type, status, transferred_at, quantity
		FROM shipment_transfer 
		WHERE id = $1
	`, req.TransferID).Scan(
		&transferExists, &batchID, &sourceType, 
		&destinationID, &destinationType, &status, &transferredAt, &quantity,
	)
	
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	
	if !transferExists {
		return fiber.NewError(fiber.StatusNotFound, "Shipment transfer not found")
	}
	
	// Check if transfer is already tokenized
	var tokenized bool
	err = db.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM transaction_nft WHERE shipment_transfer_id = $1)
	`, req.TransferID).Scan(&tokenized)
	
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	
	if tokenized {
		return fiber.NewError(fiber.StatusBadRequest, "This transaction is already tokenized")
	}
	
	// Get batch details
	var species, hatcheryID string
	var batchCreatedAt time.Time
	err = db.DB.QueryRow(`
		SELECT species, hatchery_id, created_at 
		FROM batch 
		WHERE id = $1
	`, batchID).Scan(&species, &hatcheryID, &batchCreatedAt)
	
	if err != nil && err != sql.ErrNoRows {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve batch details: "+err.Error())
	}
	
	// Initialize BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}
	
	// Prepare metadata for NFT
	metadata := map[string]interface{}{
		"type":             "transaction",
		"transfer_id":      req.TransferID,
		"batch_id":         batchID,
		"source_type":      sourceType,
		"destination_id":   destinationID,
		"destination_type": destinationType,
		"status":           status,
		"quantity":         quantity,
		"transferred_at":   transferredAt.Format(time.RFC3339),
		"created_at":       time.Now().Format(time.RFC3339),
	}
	
	// Add batch details if available
	if species != "" {
		metadata["species"] = species
		metadata["hatchery_id"] = hatcheryID
		metadata["batch_created_at"] = batchCreatedAt.Format(time.RFC3339)
	}
	
	// Merge with user-provided metadata
	for k, v := range req.Metadata {
		metadata[k] = v
	}
	
	// Create IPFS metadata
	ipfsService := ipfs.NewIPFSService()
	metadataJSON, err := ipfsService.StoreJSON(metadata)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to store metadata on IPFS: "+err.Error())
	}
	
	// Generate a unique token ID using batch ID and transfer ID
	tokenIdSuffix := fmt.Sprintf("tx_%s", req.TransferID)
	
	// Prepare contract methods for minting
	contractMethods := map[string]interface{}{
		"method": "mintTransactionNFT",
		"params": []interface{}{
			req.TransferID,
			batchID,
			req.RecipientAddress,
			metadataJSON.URI,
			tokenIdSuffix,
		},
	}
	
	// Call the contract to mint the NFT
	result, err := baasService.CallSmartContract(
		req.NetworkID,
		req.ContractAddress,
		"mintTransactionNFT",
		contractMethods,
	)
	
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to mint transaction NFT: "+err.Error())
	}
	
	// Extract token ID from result
	tokenID, ok := result["token_id"].(string)
	if !ok {
		tokenID = fmt.Sprintf("%v", result["token_id"])
	}
	
	// Generate QR code for the NFT
	qrService := NewQRCodeService()
	qrData := fmt.Sprintf("https://tracepost.app/verify?transfer=%s&token=%s&contract=%s", 
		req.TransferID, tokenID, req.ContractAddress)
	
	qrCode, err := qrService.GenerateQRCode(qrData)
	if err != nil {
		// Log the error but continue as it's not critical
		fmt.Printf("Failed to generate QR code: %v\n", err)
	}
	
	// Store QR code in IPFS
	qrCodeURI := ""
	if qrCode != nil {
		qrCodeIPFS, err := ipfsService.StoreFile(qrCode, fmt.Sprintf("qr_tx_%s.png", req.TransferID))
		if err != nil {
			fmt.Printf("Failed to store QR code on IPFS: %v\n", err)
		} else {
			qrCodeURI = qrCodeIPFS.URI
		}
	}
	
	// Record the NFT in the database
	_, err = db.DB.Exec(`
		INSERT INTO transaction_nft (
			tx_id, shipment_transfer_id, token_id, contract_address,
			token_uri, qr_code_url, owner_address, metadata,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $9
		)
	`, 
		result["tx_hash"], req.TransferID, tokenID, req.ContractAddress,
		metadataJSON.URI, qrCodeURI, req.RecipientAddress, metadataJSON.JSON,
		time.Now(),
	)
	
	if err != nil {
		fmt.Printf("Failed to record transaction NFT in database: %v\n", err)
	}
	
	// Update the shipment_transfer record with NFT information
	_, err = db.DB.Exec(`
		UPDATE shipment_transfer 
		SET nft_token_id = $1, nft_contract_address = $2, updated_at = $3
		WHERE id = $4
	`, tokenID, req.ContractAddress, time.Now(), req.TransferID)
	
	if err != nil {
		fmt.Printf("Failed to update shipment transfer with NFT info: %v\n", err)
	}
	
	// Record blockchain transaction
	_, err = db.DB.Exec(`
		INSERT INTO blockchain_record (
			related_table, related_id, tx_id, metadata_hash,
			created_at, updated_at
		) VALUES (
			'transaction_nft', $1, $2, $3, $4, $4
		)
	`, req.TransferID, result["tx_hash"], metadataJSON.CID, time.Now())
	
	if err != nil {
		fmt.Printf("Failed to record blockchain transaction: %v\n", err)
	}
	
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Transaction tokenized successfully",
		Data: map[string]interface{}{
			"transfer_id":      req.TransferID,
			"batch_id":         batchID,
			"token_id":         tokenID,
			"contract_address": req.ContractAddress,
			"network_id":       req.NetworkID,
			"owner":            req.RecipientAddress,
			"metadata_uri":     metadataJSON.URI,
			"qr_code_uri":      qrCodeURI,
			"transaction_hash": result["tx_hash"],
		},
	})
}

// GetTransactionNFTDetails retrieves the details of a transaction NFT
// @Summary Get transaction NFT details
// @Description Retrieve NFT details for a tokenized transaction
// @Tags nft
// @Accept json
// @Produce json
// @Param transferId path string true "Transaction/Transfer ID"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /nft/transactions/{transferId} [get]
func GetTransactionNFTDetails(c *fiber.Ctx) error {
	transferId := c.Params("transferId")
	if transferId == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Transfer ID is required")
	}
	
	// Query the database for the transaction NFT
	var nft TransactionNFTResponse
	var metadataJSON []byte
	var createdAt time.Time
	
	err := db.DB.QueryRow(`
		SELECT tn.token_id, tn.contract_address, tn.shipment_transfer_id,
		       st.batch_id, tn.owner_address, tn.token_uri, tn.qr_code_url,
			   tn.metadata, tn.created_at
		FROM transaction_nft tn
		JOIN shipment_transfer st ON tn.shipment_transfer_id = st.id
		WHERE tn.shipment_transfer_id = $1 AND tn.is_active = true
	`, transferId).Scan(
		&nft.TokenID, &nft.ContractAddress, &nft.TransferID,
		&nft.BatchID, &nft.OwnerAddress, &nft.TokenURI, &nft.QRCodeURL,
		&metadataJSON, &createdAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "Transaction NFT not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
		// Parse the metadata
	if metadataJSON != nil && len(metadataJSON) > 0 {
		var metadata map[string]interface{}
		err = json.Unmarshal(metadataJSON, &metadata)
		if err == nil {
			nft.Metadata = metadata
		}
	}
	
	nft.CreatedAt = createdAt
	
	return c.JSON(SuccessResponse{
		Success: true,
		Data:    nft,
	})
}

// TraceTransaction traces the history of a transaction on the blockchain
// @Summary Trace transaction
// @Description Verify and trace the history of a transaction on the blockchain
// @Tags nft
// @Accept json
// @Produce json
// @Param transferId path string true "Transaction/Transfer ID"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /nft/transactions/{transferId}/trace [get]
func TraceTransaction(c *fiber.Ctx) error {
	transferId := c.Params("transferId")
	if transferId == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Transfer ID is required")
	}
	
	// Fetch transaction NFT details
	var tokenID, contractAddress, batchID, ownerAddress string
	var networkID string
	
	err := db.DB.QueryRow(`
		SELECT tn.token_id, tn.contract_address, st.batch_id, 
		       tn.owner_address, br.network_id
		FROM transaction_nft tn
		JOIN shipment_transfer st ON tn.shipment_transfer_id = st.id
		LEFT JOIN blockchain_record br ON br.related_table = 'transaction_nft' AND br.related_id = tn.shipment_transfer_id
		WHERE tn.shipment_transfer_id = $1 AND tn.is_active = true
	`, transferId).Scan(
		&tokenID, &contractAddress, &batchID, &ownerAddress, &networkID,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "Transaction NFT not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	
	// Initialize response
	response := TransactionTraceResponse{
		TransferID:   transferId,
		BatchID:      batchID,
		TokenID:      tokenID,
		CurrentOwner: ownerAddress,
		History:      []map[string]interface{}{},
		IsVerified:   false,
	}
	
	// Fetch transaction history from database
	rows, err := db.DB.Query(`
		SELECT e.event_type, e.actor_id, a.username, a.company_id, c.name as company_name,
		       e.location, e.timestamp, e.metadata
		FROM event e
		LEFT JOIN account a ON e.actor_id = a.id
		LEFT JOIN company c ON a.company_id = c.id
		WHERE e.batch_id = $1 AND e.is_active = true
		ORDER BY e.timestamp
	`, batchID)
	
	if err != nil && err != sql.ErrNoRows {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error fetching events: "+err.Error())
	}
	
	if rows != nil {
		defer rows.Close()
		
		for rows.Next() {
			var eventType, actorID, username, companyID, companyName, location string
			var timestamp time.Time
			var metadataJSON []byte
			var metadataMap map[string]interface{}
			
			err = rows.Scan(&eventType, &actorID, &username, &companyID, &companyName, &location, &timestamp, &metadataJSON)
			if err != nil {
				continue
			}
			
			// Parse metadata
			if metadataJSON != nil {
				err = json.Unmarshal(metadataJSON, &metadataMap)
				if err != nil {
					metadataMap = make(map[string]interface{})
				}
			} else {
				metadataMap = make(map[string]interface{})
			}
			
			// Create event record
			event := map[string]interface{}{
				"event_type":   eventType,
				"actor_id":     actorID,
				"username":     username,
				"company_id":   companyID,
				"company_name": companyName,
				"location":     location,
				"timestamp":    timestamp.Format(time.RFC3339),
				"metadata":     metadataMap,
			}
			
			response.History = append(response.History, event)
		}
	}
	
	// Fetch blockchain verification data
	baasService := blockchain.NewBaaSService()
	if baasService != nil && networkID != "" && contractAddress != "" && tokenID != "" {
		// Get token data from blockchain
		contractMethods := map[string]interface{}{
			"method": "getTokenInfo",
			"params": []interface{}{tokenID},
		}
		
		result, err := baasService.CallSmartContract(
			networkID,
			contractAddress,
			"verifyTransaction",
			contractMethods,
		)
		
		if err == nil && result != nil {
			response.IsVerified = true
			response.VerifiedData = result
		}
	}
	
	// Get shipment transfers related to this batch
	shipmentRows, err := db.DB.Query(`
		SELECT id, source_type, destination_id, destination_type,
		       status, transferred_at, transferred_by, blockchain_tx_id, nft_token_id
		FROM shipment_transfer
		WHERE batch_id = $1 AND is_active = true
		ORDER BY transferred_at
	`, batchID)
	
	if err == nil && shipmentRows != nil {
		defer shipmentRows.Close()
		
		for shipmentRows.Next() {
			var id, sourceType, destinationID, destinationType, status string
			var transferredBy, blockchainTxID, nftTokenID sql.NullString
			var transferredAt time.Time
			
			err = shipmentRows.Scan(
				&id, &sourceType, &destinationID, &destinationType,
				&status, &transferredAt, &transferredBy, &blockchainTxID, &nftTokenID,
			)
			
			if err != nil {
				continue
			}
			
			transfer := map[string]interface{}{
				"event_type":        "transfer",
				"transfer_id":       id,
				"source_type":       sourceType,
				"destination_id":    destinationID,
				"destination_type":  destinationType,
				"status":            status,
				"transferred_at":    transferredAt.Format(time.RFC3339),
				"has_nft":           nftTokenID.Valid,
				"blockchain_tx_id":  blockchainTxID.String,
				"metadata": map[string]interface{}{
					"is_current_transfer": id == transferId,
				},
			}
			
			if transferredBy.Valid {
				transfer["transferred_by"] = transferredBy.String
			}
			
			if nftTokenID.Valid {
				transfer["nft_token_id"] = nftTokenID.String
			}
			
			response.History = append(response.History, transfer)
		}
	}
	
	// Sort history by timestamp
	sort.Slice(response.History, func(i, j int) bool {
		timeI, okI := response.History[i]["timestamp"].(string)
		if !okI {
			timeI, okI = response.History[i]["transferred_at"].(string)
			if !okI {
				return false
			}
		}
		
		timeJ, okJ := response.History[j]["timestamp"].(string)
		if !okJ {
			timeJ, okJ = response.History[j]["transferred_at"].(string)
			if !okJ {
				return true
			}
		}
		
		return timeI < timeJ
	})
	
	return c.JSON(SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// QRCodeService provides functionality for generating QR codes
type QRCodeService struct{}

// NewQRCodeService creates a new QR code service
func NewQRCodeService() *QRCodeService {
	return &QRCodeService{}
}

// GenerateQRCode generates a QR code from the given data
func (s *QRCodeService) GenerateQRCode(data string) ([]byte, error) {
	qr, err := qrcode.Encode(data, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}
	return qr, nil
}

// GenerateVerificationQRCode generates a QR code for verification with custom settings
func (s *QRCodeService) GenerateVerificationQRCode(data string) ([]byte, error) {
	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return nil, err
	}
	qr.DisableBorder = false
	
	// Generate PNG data
	pngData, err := qr.PNG(256)
	if err != nil {
		return nil, err
	}
	
	return pngData, nil
}

// GenerateTransactionVerificationQR generates a QR code for transaction verification
// @Summary Generate transaction verification QR
// @Description Generate a QR code for transaction verification that links to a verification page
// @Tags nft
// @Accept json
// @Produce json
// @Param transferId path string true "Transaction/Transfer ID"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /nft/transactions/{transferId}/qr [get]
func GenerateTransactionVerificationQR(c *fiber.Ctx) error {
	transferId := c.Params("transferId")
	if transferId == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Transfer ID is required")
	}
	
	// Check if transaction exists and has NFT
	var tokenID, contractAddress string
	var exists bool
	
	err := db.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM transaction_nft WHERE shipment_transfer_id = $1 AND is_active = true
		),
		token_id, contract_address
		FROM transaction_nft
		WHERE shipment_transfer_id = $1 AND is_active = true
	`, transferId).Scan(&exists, &tokenID, &contractAddress)
	
	if err != nil && err != sql.ErrNoRows {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	
	if !exists {
		// If no NFT exists, check if the transfer exists
		err = db.DB.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM shipment_transfer WHERE id = $1)
		`, transferId).Scan(&exists)
		
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Database error: "+err.Error())
		}
		
		if !exists {
			return fiber.NewError(fiber.StatusNotFound, "Transaction not found")
		}
		
		// Transfer exists but has no NFT
		return fiber.NewError(fiber.StatusNotFound, "Transaction has not been tokenized yet")
	}
	
	// Generate verification URL
	baseURL := "https://tracepost.app/verify"
	verificationURL := fmt.Sprintf("%s?transfer=%s&token=%s&contract=%s", 
		baseURL, transferId, tokenID, contractAddress)
	
	// Generate QR code for the verification URL
	qrService := NewQRCodeService()
	qrCode, err := qrService.GenerateQRCode(verificationURL)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR code: "+err.Error())
	}
	
	// Set the content type and send the PNG image
	c.Set(fiber.HeaderContentType, "image/png")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="tx_%s_verification_qr.png"`, transferId))
	return c.Send(qrCode)
}

// Add a new handler for interacting with LogisticsTraceabilityNFT contract
func InteractWithLogisticsNFTContract(c *fiber.Ctx) error {
	type Request struct {
		ContractAddress string                 `json:"contract_address"`
		FunctionName    string                 `json:"function_name"`
		Args            map[string]interface{} `json:"args"`
		NetworkID       string                 `json:"network_id"` // Add this line
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.ContractAddress == "" {
		return fiber.NewError(fiber.StatusBadRequest, "contract_address is required")
	}

	if req.FunctionName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "function_name is required")
	}

	// Initialize the BaaS service
	baasService := blockchain.NewBaaSService()
	if baasService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to initialize BaaS service")
	}

	// Interact with the contract
	result, err := baasService.CallSmartContract(
		req.NetworkID,
		req.ContractAddress,
		req.FunctionName,
		req.Args,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to interact with contract: "+err.Error())
	}

	// Return the result
	return c.JSON(fiber.Map{
		"result": result,
	})
}