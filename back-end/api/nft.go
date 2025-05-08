package api

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
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
	BatchID         string `json:"batch_id"`
	NetworkID       string `json:"network_id"`
	ContractAddress string `json:"contract_address"`
	RecipientAddress string `json:"recipient_address"`
}

// TransferNFTRequest represents a request to transfer an NFT to a new owner
type TransferNFTRequest struct {
	ContractAddress string `json:"contract_address"`
	NetworkID       string `json:"network_id"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
}

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
	if err != nil {
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
	
	// Prepare the contract call
	contractMethods := map[string]interface{}{
		"method": "mintBatchNFT",
		"params": []interface{}{
			req.BatchID,
			req.RecipientAddress,
			"", // Will be overridden with generated URI below
		},
	}
	
	// First generate the token URI using the contract's generateTokenURI method
	tokenURIResult, err := baasService.QueryContractState(
    req.NetworkID,
    req.ContractAddress,
    map[string]interface{}{
        "method": "generateTokenURI",
        "params": []interface{}{
            req.BatchID,
            species,
            location,
            createdAt.Unix(),
            qrCodeURL,
      	 },
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
			batch_id, network_id, contract_address, token_id, recipient, token_uri, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`, req.BatchID, req.NetworkID, req.ContractAddress, int(tokenID), req.RecipientAddress, tokenURI, time.Now())
	
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
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid token ID format")
	}
	
	// Check if token exists in the database
	var batchID string
	var recipient, tokenURI string
	var createdAt time.Time
	
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