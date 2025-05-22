package dto

import (
	"time"
)

// BatchBlockchainDataResponse represents the blockchain data response for a batch
type BatchBlockchainDataResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    BatchBlockchainDataDTO `json:"data"`
}

// BatchBlockchainDataDTO represents the blockchain data for a batch
type BatchBlockchainDataDTO struct {
	BatchID   string              `json:"batch_id"`
	FirstTx   time.Time           `json:"first_tx"`
	LatestTx  time.Time           `json:"latest_tx"`
	State     BatchBlockchainState `json:"state"`
	TxCount   int                 `json:"tx_count"`
	Txs       []BlockchainTxDTO   `json:"txs"`
}

// BatchBlockchainState represents the current state of a batch in the blockchain
type BatchBlockchainState struct {
	BatchID    string `json:"batch_id"`
	HatcheryID string `json:"hatchery_id"`
	Quantity   int    `json:"quantity"`
	Species    string `json:"species"`
	Status     string `json:"status"`
}

// BlockchainTxDTO represents a blockchain transaction
type BlockchainTxDTO struct {
	TxID        string                 `json:"tx_id"`
	Type        string                 `json:"type"`
	Timestamp   time.Time              `json:"timestamp"`
	Payload     map[string]interface{} `json:"payload"`
	ValidatedAt time.Time              `json:"validated_at"`
}