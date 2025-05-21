package db

import (
	"database/sql"
	"fmt"
	"strings"
	"encoding/json"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// NFTMetadataSchema defines the expected structure of NFT metadata
type NFTMetadataSchema struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Image         string   `json:"image"`
	ExternalURL   string   `json:"external_url,omitempty"`
	Attributes    []NFTAttribute `json:"attributes"`
	BatchID       int      `json:"batch_id"`
	ShipmentID    string   `json:"shipment_id"`
	Origin        string   `json:"origin"`
	Destination   string   `json:"destination,omitempty"`
	CreationDate  string   `json:"creation_date"`
	ExpiryDate    string   `json:"expiry_date,omitempty"`
	BatchQuantity int      `json:"batch_quantity"`
}

type NFTAttribute struct {
	TraitType   string      `json:"trait_type"`
	Value       interface{} `json:"value"`
	DisplayType string      `json:"display_type,omitempty"`
}

// TransactionNFT represents the transaction_nft table row
type TransactionNFT struct {
	ID                 int
	TxID               string
	ShipmentTransferID string
	TokenID            string
	ContractAddress    string
	TokenURI           sql.NullString
	QRCodeURL          sql.NullString
	OwnerAddress       string
	Status             string
	BlockchainRecordID sql.NullInt64
	BatchID            sql.NullInt64
	Metadata           []byte
	MetadataSchema     sql.NullString
	DigestHash         sql.NullString
	CreatedAt          string
	UpdatedAt          string
	IsActive           bool
}

// ValidateNFTMetadata validates the structure of NFT metadata against expected schema
func ValidateNFTMetadata(metadata []byte) error {
	var data NFTMetadataSchema
	
	// Attempt to unmarshal the metadata
	err := json.Unmarshal(metadata, &data)
	if err != nil {
		return fmt.Errorf("invalid metadata JSON format: %w", err)
	}
	
	// Perform basic validation
	if data.Name == "" {
		return errors.New("metadata missing required field: name")
	}
	
	if data.Description == "" {
		return errors.New("metadata missing required field: description")
	}
	
	if data.Image == "" {
		return errors.New("metadata missing required field: image")
	}
	
	if data.BatchID <= 0 {
		return errors.New("metadata has invalid batch_id")
	}
	
	if data.ShipmentID == "" {
		return errors.New("metadata missing required field: shipment_id")
	}
	
	if data.BatchQuantity <= 0 {
		return errors.New("metadata has invalid batch_quantity")
	}
	
	return nil
}

// GenerateDigestHash creates a hash from NFT data for data integrity verification
func GenerateDigestHash(nft *TransactionNFT) (string, error) {
	// Combine critical fields into a string for hashing
	dataToHash := fmt.Sprintf("%s:%s:%s:%s:%s:%d",
		nft.TxID,
		nft.ShipmentTransferID,
		nft.TokenID,
		nft.ContractAddress,
		nft.OwnerAddress,
		nft.BatchID.Int64,
	)
	
	// Add metadata if available
	if len(nft.Metadata) > 0 {
		dataToHash += ":" + string(nft.Metadata)
	}
	
	// Generate SHA-256 hash
	hash := sha256.Sum256([]byte(dataToHash))
	return hex.EncodeToString(hash[:]), nil
}

// VerifyNFTDataIntegrity checks if an NFT's data is consistent and valid
func VerifyNFTDataIntegrity(nftID int) (bool, string, error) {
	// Get NFT data
	var nft TransactionNFT
	var metadata []byte
	
	err := DB.QueryRow(`
		SELECT id, tx_id, shipment_transfer_id, token_id, contract_address, 
		       token_uri, qr_code_url, owner_address, status, blockchain_record_id, 
		       batch_id, metadata, metadata_schema, digest_hash
		FROM transaction_nft 
		WHERE id = $1 AND is_active = true
	`, nftID).Scan(
		&nft.ID, &nft.TxID, &nft.ShipmentTransferID, &nft.TokenID, &nft.ContractAddress,
		&nft.TokenURI, &nft.QRCodeURL, &nft.OwnerAddress, &nft.Status, &nft.BlockchainRecordID,
		&nft.BatchID, &metadata, &nft.MetadataSchema, &nft.DigestHash,
	)
	
	if err != nil {
		return false, "", fmt.Errorf("failed to retrieve NFT data: %w", err)
	}
	
	nft.Metadata = metadata
	
	// Verify referenced entities exist
	if nft.BatchID.Valid {
		var batchExists bool
		err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM batch WHERE id = $1)", nft.BatchID.Int64).Scan(&batchExists)
		if err != nil {
			return false, "", fmt.Errorf("failed to verify batch existence: %w", err)
		}
		if !batchExists {
			return false, "Referenced batch does not exist", nil
		}
	}
	
	// Verify blockchain record exists
	if nft.BlockchainRecordID.Valid {
		var recordExists bool
		err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM blockchain_record WHERE id = $1)", nft.BlockchainRecordID.Int64).Scan(&recordExists)
		if err != nil {
			return false, "", fmt.Errorf("failed to verify blockchain record: %w", err)
		}
		if !recordExists {
			return false, "Referenced blockchain record does not exist", nil
		}
	}
	
	// Verify shipment transfer exists
	var shipmentExists bool
	err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM shipment_transfer WHERE id = $1)", nft.ShipmentTransferID).Scan(&shipmentExists)
	if err != nil {
		return false, "", fmt.Errorf("failed to verify shipment transfer: %w", err)
	}
	if !shipmentExists {
		return false, "Referenced shipment transfer does not exist", nil
	}
	
	// Validate metadata structure if exists
	if len(nft.Metadata) > 0 {
		if err := ValidateNFTMetadata(nft.Metadata); err != nil {
			return false, fmt.Sprintf("Metadata validation error: %v", err), nil
		}
	}
	
	// Check data integrity with hash
	if nft.DigestHash.Valid && nft.DigestHash.String != "" {
		calculatedHash, err := GenerateDigestHash(&nft)
		if err != nil {
			return false, "", fmt.Errorf("failed to calculate digest hash: %w", err)
		}
		
		if calculatedHash != nft.DigestHash.String {
			return false, "Digest hash mismatch, data integrity compromised", nil
		}
	}
	
	return true, "Data integrity verified successfully", nil
}

// FindDuplicateNFTs checks for duplicate NFTs in the system
func FindDuplicateNFTs() ([]string, error) {
	duplicates := []string{}
	
	// Check for duplicate token_ids within the same contract
	rows, err := DB.Query(`
		SELECT token_id, contract_address, COUNT(*) 
		FROM transaction_nft 
		WHERE is_active = true 
		GROUP BY token_id, contract_address 
		HAVING COUNT(*) > 1
	`)
	
	if err != nil {
		return nil, fmt.Errorf("failed to query for duplicate tokens: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var tokenID, contractAddress string
		var count int
		if err := rows.Scan(&tokenID, &contractAddress, &count); err != nil {
			return nil, fmt.Errorf("error scanning duplicate row: %w", err)
		}
		
		duplicates = append(duplicates, fmt.Sprintf("Token ID %s on contract %s has %d duplicate entries", 
			tokenID, contractAddress, count))
	}
	
	// Check for duplicate shipment transfers with different NFTs
	dupShipmentRows, err := DB.Query(`
		SELECT shipment_transfer_id, COUNT(*) 
		FROM transaction_nft 
		WHERE is_active = true 
		GROUP BY shipment_transfer_id 
		HAVING COUNT(*) > 1
	`)
	
	if err != nil {
		return nil, fmt.Errorf("failed to query for duplicate shipments: %w", err)
	}
	defer dupShipmentRows.Close()
	
	for dupShipmentRows.Next() {
		var shipmentID string
		var count int
		if err := dupShipmentRows.Scan(&shipmentID, &count); err != nil {
			return nil, fmt.Errorf("error scanning duplicate shipment row: %w", err)
		}
		
		duplicates = append(duplicates, fmt.Sprintf("Shipment ID %s has %d NFTs associated with it", 
			shipmentID, count))
	}
	
	return duplicates, nil
}

// EncryptSensitiveData encrypts sensitive data in the NFT table
func EncryptSensitiveData(nftID int, field string, value string) (string, error) {
	// This is a placeholder - in a real implementation you would:
	// 1. Use encryption libraries like AES from crypto/aes
	// 2. Use key management systems for proper key handling
	// 3. Implement proper authentication and authorization
	
	// Simple obfuscation for example purposes (NOT for production use)
	h := sha256.New()
	h.Write([]byte(value))
	hashedValue := hex.EncodeToString(h.Sum(nil))
	
	// First 6 characters + last 4 characters of the original + hash
	var maskedValue string
	if len(value) > 10 {
		maskedValue = value[:6] + "..." + value[len(value)-4:] + ":" + hashedValue
	} else {
		maskedValue = "***:" + hashedValue
	}
	
	return maskedValue, nil
}

// GenerateQRCode creates a QR code URL for an NFT
func GenerateQRCode(nftID int) (string, error) {
	// This is a placeholder - in a real implementation you would:
	// 1. Generate an actual QR code using a library
	// 2. Store the QR code image on IPFS or another storage
	// 3. Return the URL to the stored QR code
	
	// For now, we'll just create a mock URL
	var txID, tokenID string
	err := DB.QueryRow("SELECT tx_id, token_id FROM transaction_nft WHERE id = $1", nftID).Scan(&txID, &tokenID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve NFT data for QR code: %w", err)
	}
	
	// This would be replaced with actual QR code generation
	return fmt.Sprintf("https://tracepost-api.com/qr/%s/%s", strings.ToLower(txID[:8]), tokenID), nil
}
