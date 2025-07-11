package api

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/config"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"strconv"
	"time"
)

// ShardingConfigRequest represents a request to configure sharding
type ShardingConfigRequest struct {
	Enabled       bool                   `json:"enabled"`
	ShardCount    int                    `json:"shard_count"`
	ShardingType  string                 `json:"sharding_type"`
	ShardStrategy string                 `json:"shard_strategy"`
	ConfigParams  map[string]interface{} `json:"config_params,omitempty"`
	Description   string                 `json:"description,omitempty"`
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

	// Configure sharding
	now := time.Now()

	if req.Enabled {
		// Create initial shards
		for i := 1; i <= req.ShardCount; i++ {
			shardID := fmt.Sprintf("shard-%d", i)
			_, err := db.DB.Exec(`
				INSERT INTO shards (shard_id, shard_number, shard_type, allocation_strategy, status, created_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`,
				shardID,
				i,
				req.ShardingType,
				req.ShardStrategy,
				"active",
				now,
			)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to create shard "+shardID+": "+err.Error())
			}
		}
	}

	// Update configuration in memory
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