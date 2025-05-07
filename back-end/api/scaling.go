package api

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/vietchain/tracepost-larvae/blockchain"
	"github.com/vietchain/tracepost-larvae/config"
	"github.com/vietchain/tracepost-larvae/db"
	"strconv"
	"time"
)

// ScalingConfigRequest represents a request to configure Layer 2 scaling
type ScalingConfigRequest struct {
	Enabled     bool   `json:"enabled"`
	Provider    string `json:"provider"` // "optimism", "arbitrum", "zksync", etc.
	ChainID     string `json:"chain_id,omitempty"`
	Endpoint    string `json:"endpoint,omitempty"`
	BridgeType  string `json:"bridge_type,omitempty"` // "canonical", "custom"
	Description string `json:"description,omitempty"`
}

// ShardingConfigRequest represents a request to configure sharding
type ShardingConfigRequest struct {
	Enabled       bool                   `json:"enabled"`
	ShardCount    int                    `json:"shard_count"`
	ShardingType  string                 `json:"sharding_type"` // "state", "transaction", "data"
	ShardStrategy string                 `json:"shard_strategy"` // "geographic", "batch-type", "timestamp"
	ConfigParams  map[string]interface{} `json:"config_params,omitempty"`
	Description   string                 `json:"description,omitempty"`
}

// ScalingStatusResponse represents the status of Layer 2 scaling
type ScalingStatusResponse struct {
	Layer2Enabled  bool                   `json:"layer2_enabled"`
	Provider       string                 `json:"provider,omitempty"`
	ChainID        string                 `json:"chain_id,omitempty"`
	BridgeAddress  string                 `json:"bridge_address,omitempty"`
	LastSync       string                 `json:"last_sync,omitempty"`
	TxCount        int                    `json:"tx_count"`
	GasReduction   float64                `json:"gas_reduction,omitempty"`
	Performance    map[string]interface{} `json:"performance,omitempty"`
	ConfiguredAt   string                 `json:"configured_at,omitempty"`
	ShardingStatus *ShardingStatus        `json:"sharding_status,omitempty"`
}

// ShardingStatus represents the status of sharding
type ShardingStatus struct {
	Enabled       bool                   `json:"enabled"`
	ShardCount    int                    `json:"shard_count"`
	ActiveShards  int                    `json:"active_shards"`
	ShardingType  string                 `json:"sharding_type"`
	ShardStrategy string                 `json:"shard_strategy"`
	ConfigParams  map[string]interface{} `json:"config_params,omitempty"`
	ConfiguredAt  string                 `json:"configured_at,omitempty"`
	ShardStats    map[string]interface{} `json:"shard_stats,omitempty"`
}

// EnableLayer2Scaling enables or disables Layer 2 scaling
// @Summary Enable Layer 2 scaling
// @Description Enable or disable Layer 2 scaling for improved performance and sustainability
// @Tags scaling
// @Accept json
// @Produce json
// @Param request body ScalingConfigRequest true "Layer 2 scaling configuration"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /scaling/l2/enable [post]
func EnableLayer2Scaling(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Parse request
	var req ScalingConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.Enabled && (req.Provider == "" || req.ChainID == "" || req.Endpoint == "") {
		return fiber.NewError(fiber.StatusBadRequest, "Provider, Chain ID, and Endpoint are required when enabling Layer 2")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Configure Layer 2 scaling
	now := time.Now()
	var bridgeAddress string
	
	if req.Enabled {
		// In a real implementation, this would:
		// 1. Deploy or connect to a Layer 2 bridge contract
		// 2. Configure the blockchain client to use the Layer 2 chain
		// 3. Migrate or synchronize necessary state
		
		// For this example, we'll simulate a successful Layer 2 configuration
		bridgeAddress = "0x1234567890123456789012345678901234567890" // Example bridge address
	} else {
		// Disable Layer 2 scaling
		// In a real implementation, this would reconfigure the system to use the base layer only
	}
	
	// Record configuration in blockchain
	_, err := blockchainClient.SubmitTransaction("LAYER2_CONFIG", map[string]interface{}{
		"enabled":       req.Enabled,
		"provider":      req.Provider,
		"chain_id":      req.ChainID,
		"endpoint":      req.Endpoint,
		"bridge_type":   req.BridgeType,
		"bridge_address": bridgeAddress,
		"description":   req.Description,
		"timestamp":     now,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record Layer 2 configuration on blockchain: "+err.Error())
	}
	
	// Update configuration in database
	_, err = db.DB.Exec(`
		INSERT INTO layer2_config (enabled, provider, chain_id, endpoint, bridge_type, bridge_address, description, configured_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (provider) DO UPDATE SET 
		enabled = EXCLUDED.enabled,
		chain_id = EXCLUDED.chain_id,
		endpoint = EXCLUDED.endpoint,
		bridge_type = EXCLUDED.bridge_type,
		bridge_address = EXCLUDED.bridge_address,
		description = EXCLUDED.description,
		configured_at = EXCLUDED.configured_at
	`,
		req.Enabled,
		req.Provider,
		req.ChainID,
		req.Endpoint,
		req.BridgeType,
		bridgeAddress,
		req.Description,
		now,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update Layer 2 configuration in database: "+err.Error())
	}
	
	// Update config in memory
	configUpdate := map[string]interface{}{
		"LayerTwoEnabled":   req.Enabled,
		"LayerTwoProvider":  req.Provider,
		"LayerTwoChainID":   req.ChainID,
		"LayerTwoEndpoint":  req.Endpoint,
		"LayerTwoBridgeType": req.BridgeType,
	}
	cfg.UpdateConfig(configUpdate)
	
	// Return response
	message := "Layer 2 scaling has been disabled"
	if req.Enabled {
		message = "Layer 2 scaling has been enabled"
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: message,
		Data: map[string]interface{}{
			"enabled":        req.Enabled,
			"provider":       req.Provider,
			"chain_id":       req.ChainID,
			"bridge_address": bridgeAddress,
			"configured_at":  now.Format(time.RFC3339),
		},
	})
}

// GetLayer2Status gets the status of Layer 2 scaling
// @Summary Get Layer 2 status
// @Description Get the current status of Layer 2 scaling
// @Tags scaling
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse{data=ScalingStatusResponse}
// @Failure 500 {object} ErrorResponse
// @Router /scaling/l2/status [get]
func GetLayer2Status(c *fiber.Ctx) error {
	// Get Layer 2 configuration from database
	var config struct {
		Enabled       bool
		Provider      string
		ChainID       string
		BridgeAddress string
		ConfiguredAt  time.Time
	}
	
	err := db.DB.QueryRow(`
		SELECT enabled, provider, chain_id, bridge_address, configured_at
		FROM layer2_config
		ORDER BY configured_at DESC
		LIMIT 1
	`).Scan(
		&config.Enabled,
		&config.Provider,
		&config.ChainID,
		&config.BridgeAddress,
		&config.ConfiguredAt,
	)
	if err != nil {
		// No configuration found, return default values
		return c.JSON(SuccessResponse{
			Success: true,
			Message: "Layer 2 scaling is not configured",
			Data: ScalingStatusResponse{
				Layer2Enabled: false,
				TxCount:       0,
			},
		})
	}
	
	// Get sharding status
	var shardingStatus ShardingStatus
	var shardConfiguredAt time.Time
	err = db.DB.QueryRow(`
		SELECT enabled, shard_count, sharding_type, shard_strategy, config_params, configured_at
		FROM sharding_config
		ORDER BY configured_at DESC
		LIMIT 1
	`).Scan(
		&shardingStatus.Enabled,
		&shardingStatus.ShardCount,
		&shardingStatus.ShardingType,
		&shardingStatus.ShardStrategy,
		&shardingStatus.ConfigParams,
			&shardConfiguredAt,
	)
	if err == nil {
		// Get active shard count
		err = db.DB.QueryRow(`
			SELECT COUNT(*) FROM shards WHERE status = 'active'
		`).Scan(&shardingStatus.ActiveShards)
		if err != nil {
			shardingStatus.ActiveShards = 0
		}
		
		shardingStatus.ConfiguredAt = shardConfiguredAt.Format(time.RFC3339)
		
		// Get shard statistics
		shardingStatus.ShardStats = getShardStats()
	} else {
		shardingStatus.Enabled = false
		shardingStatus.ShardCount = 0
		shardingStatus.ActiveShards = 0
	}
	
	// Get performance metrics
	performance := make(map[string]interface{})
	if config.Enabled {
		// In a real implementation, this would query actual performance metrics
		// For this example, we'll provide some sample data
		performance["tps"] = 500
		performance["latency_ms"] = 150
		performance["block_time_sec"] = 2
		
		// Calculate gas reduction
		gasReduction := 95.5 // 95.5% gas reduction compared to base layer
		
		// Get transaction count
		var txCount int
		err = db.DB.QueryRow(`
			SELECT COUNT(*) FROM blockchain_record WHERE layer = 'l2'
		`).Scan(&txCount)
		if err != nil {
			txCount = 0
		}
		
		// Get last sync time
		var lastSync time.Time
		err = db.DB.QueryRow(`
			SELECT MAX(timestamp) FROM layer2_sync
		`).Scan(&lastSync)
		
		// Return response with Layer 2 enabled
		return c.JSON(SuccessResponse{
			Success: true,
			Message: "Layer 2 scaling status retrieved successfully",
			Data: ScalingStatusResponse{
				Layer2Enabled:  true,
				Provider:       config.Provider,
				ChainID:        config.ChainID,
				BridgeAddress:  config.BridgeAddress,
				LastSync:       lastSync.Format(time.RFC3339),
				TxCount:        txCount,
				GasReduction:   gasReduction,
				Performance:    performance,
				ConfiguredAt:   config.ConfiguredAt.Format(time.RFC3339),
				ShardingStatus: &shardingStatus,
			},
		})
	} else {
		// Return response with Layer 2 disabled
		return c.JSON(SuccessResponse{
			Success: true,
			Message: "Layer 2 scaling is disabled",
			Data: ScalingStatusResponse{
				Layer2Enabled:  false,
				TxCount:        0,
				ConfiguredAt:   config.ConfiguredAt.Format(time.RFC3339),
				ShardingStatus: &shardingStatus,
			},
		})
	}
}

// ConfigureSharding configures sharding for the TracePost-larvaeChain
// @Summary Configure sharding
// @Description Configure sharding for improved scalability and performance
// @Tags scaling
// @Accept json
// @Produce json
// @Param request body ShardingConfigRequest true "Sharding configuration"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /scaling/sharding/configure [post]
func ConfigureSharding(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	
	// Parse request
	var req ShardingConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}
	
	// Validate request
	if req.Enabled && (req.ShardCount <= 0 || req.ShardingType == "" || req.ShardStrategy == "") {
		return fiber.NewError(fiber.StatusBadRequest, "Shard count, sharding type, and shard strategy are required when enabling sharding")
	}
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		cfg.BlockchainPrivateKey,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Configure sharding
	now := time.Now()
	
	if req.Enabled {
		// In a real implementation, this would:
		// 1. Set up the necessary sharding infrastructure
		// 2. Deploy shard contracts or configure shard nodes
		// 3. Initialize shard allocation strategy
		
		// For this example, we'll simulate a successful sharding configuration
	} else {
		// Disable sharding
		// In a real implementation, this would reconfigure the system to use a single shard
	}
	
	// Record configuration in blockchain
	_, err := blockchainClient.SubmitTransaction("SHARDING_CONFIG", map[string]interface{}{
		"enabled":        req.Enabled,
		"shard_count":    req.ShardCount,
		"sharding_type":  req.ShardingType,
		"shard_strategy": req.ShardStrategy,
		"config_params":  req.ConfigParams,
		"description":    req.Description,
		"timestamp":      now,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record sharding configuration on blockchain: "+err.Error())
	}
	
	// Update configuration in database
	_, err = db.DB.Exec(`
		INSERT INTO sharding_config (enabled, shard_count, sharding_type, shard_strategy, config_params, description, configured_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		req.Enabled,
		req.ShardCount,
		req.ShardingType,
		req.ShardStrategy,
		req.ConfigParams,
		req.Description,
		now,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update sharding configuration in database: "+err.Error())
	}
	
	// If sharding is enabled, create initial shards
	if req.Enabled {
		for i := 1; i <= req.ShardCount; i++ {
			shardID := fmt.Sprintf("shard-%d", i)
			
			// Insert shard record
			_, err = db.DB.Exec(`
				INSERT INTO shards (shard_id, shard_number, shard_type, allocation_strategy, status, created_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`,
				shardID,
				i,
				req.ShardingType,
				req.ShardStrategy,
				"active", // All shards start as active
				now,
			)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to create shard "+shardID+": "+err.Error())
			}
		}
	}
	
	// Update config in memory
	configUpdate := map[string]interface{}{
		"ShardingEnabled":    req.Enabled,
		"ShardCount":         req.ShardCount,
		"ShardingType":       req.ShardingType,
		"ShardStrategy":      req.ShardStrategy,
	}
	cfg.UpdateConfig(configUpdate)
	
	// Return response
	message := "Sharding has been disabled"
	if req.Enabled {
		message = "Sharding has been enabled with " + strconv.Itoa(req.ShardCount) + " shards"
	}
	
	return c.JSON(SuccessResponse{
		Success: true,
		Message: message,
		Data: map[string]interface{}{
			"enabled":        req.Enabled,
			"shard_count":    req.ShardCount,
			"sharding_type":  req.ShardingType,
			"shard_strategy": req.ShardStrategy,
			"configured_at":  now.Format(time.RFC3339),
		},
	})
}

// Helper function to get shard statistics
func getShardStats() map[string]interface{} {
	// In a real implementation, this would query actual shard statistics
	// For this example, we'll provide some sample data
	return map[string]interface{}{
		"avg_tx_per_shard":      1200,
		"avg_block_time_sec":    2.5,
		"cross_shard_tx_ratio":  0.15, // 15% of transactions cross shards
		"shard_size_variation":  0.05, // 5% variation in size between shards
		"rebalancing_threshold": 0.2,  // Trigger rebalancing when imbalance exceeds 20%
	}
}