package blockchain

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ShardingConfig contains configuration for blockchain sharding
type ShardingConfig struct {
	Enabled         bool               // Whether sharding is enabled
	ShardCount      int                // Number of shards
	NodesPerShard   int                // Number of validator nodes per shard
	ShardAssignments map[string]string  // Assignment of nodes to shards
	CrossShardTxs   bool               // Whether cross-shard transactions are enabled
	ShardRebalance  bool               // Whether automatic shard rebalancing is enabled
	CrossShardProtocol string           // Protocol for cross-shard communication
	ShardRebalanceInterval int          // Interval for shard rebalancing
}

// ShardingManager manages the sharding of the blockchain
type ShardingManager struct {
	config  ShardingConfig
	shards  map[int]*Shard
	mutex   sync.RWMutex
	metrics *ShardingMetrics
}

// Shard represents a single shard in the blockchain
type Shard struct {
	ID            int
	ValidatorIDs  []string
	BlockHeight   int64
	Transactions  int64
	LoadFactor    float64
	CrossShardTxs map[int]int64 // Transactions to other shards, by shard ID
}

// ShardingMetrics contains metrics for sharding performance
type ShardingMetrics struct {
	TotalTransactions      int64
	CrossShardTransactions int64
	ThroughputPerShard     map[int]float64
	LatencyPerShard        map[int]float64
	LoadImbalance          float64
	LastMeasured           time.Time
}

// NewShardingManager creates a new sharding manager
func NewShardingManager(config ShardingConfig) *ShardingManager {
	manager := &ShardingManager{
		config: config,
		shards: make(map[int]*Shard),
		metrics: &ShardingMetrics{
			ThroughputPerShard: make(map[int]float64),
			LatencyPerShard:    make(map[int]float64),
			LastMeasured:       time.Now(),
		},
	}

	// Initialize shards if enabled
	if config.Enabled {
		manager.initializeShards()
	}

	return manager
}

// initializeShards initializes the shards based on configuration
func (sm *ShardingManager) initializeShards() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Create shards
	for i := 0; i < sm.config.ShardCount; i++ {
		sm.shards[i] = &Shard{
			ID:            i,
			ValidatorIDs:  make([]string, 0, sm.config.NodesPerShard),
			BlockHeight:   0,
			Transactions:  0,
			LoadFactor:    0.0,
			CrossShardTxs: make(map[int]int64),
		}
	}	// Assign validators to shards
	if len(sm.config.ShardAssignments) > 0 {
		// Use predefined assignment if available
		for nodeID, shardID := range sm.config.ShardAssignments {
			shardIDInt := 0
			fmt.Sscanf(shardID, "shard-%d", &shardIDInt)
			if shard, ok := sm.shards[shardIDInt]; ok {
				shard.ValidatorIDs = append(shard.ValidatorIDs, nodeID)
			}
		}
	} else {
		// Random assignment for simulation
		nodeIDs := generateSimulatedNodeIDs(sm.config.ShardCount * sm.config.NodesPerShard)
		for i, nodeID := range nodeIDs {
			shardID := i % sm.config.ShardCount
			if shard, ok := sm.shards[shardID]; ok {
				shard.ValidatorIDs = append(shard.ValidatorIDs, nodeID)
			}
		}
	}
}

// GetShardForTransaction determines which shard should handle a transaction
func (sm *ShardingManager) GetShardForTransaction(txID string, accountAddr string) int {
	if !sm.config.Enabled {
		return 0 // No sharding, use default shard
	}

	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Simple sharding: use account address to determine shard
	// In a production system, this would use a more sophisticated algorithm
	// such as consistent hashing or account-based sharding
	
	// Use the first few bytes of the account address as a deterministic way to assign a shard
	var shardIndex int
	if len(accountAddr) > 0 {
		// Use the first character of the address to determine the shard
		// This is a simplified approach for demonstration
		shardIndex = int(accountAddr[0]) % sm.config.ShardCount
	} else if len(txID) > 0 {
		// Fall back to transaction ID if account address not available
		shardIndex = int(txID[0]) % sm.config.ShardCount
	} else {
		// If neither is available, use a random shard
		shardIndex = rand.Intn(sm.config.ShardCount)
	}

	return shardIndex
}

// RecordTransaction records a transaction in the appropriate shard
func (sm *ShardingManager) RecordTransaction(txID string, accountAddr string, crossShardAccounts []string) {
	if !sm.config.Enabled {
		return
	}

	// Determine primary shard for this transaction
	primaryShardID := sm.GetShardForTransaction(txID, accountAddr)

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Record transaction in primary shard
	if shard, ok := sm.shards[primaryShardID]; ok {
		shard.Transactions++
		sm.metrics.TotalTransactions++

		// Update load factor
		shard.LoadFactor = float64(shard.Transactions) / float64(sm.metrics.TotalTransactions)

		// Process cross-shard aspects of transaction
		if len(crossShardAccounts) > 0 && sm.config.CrossShardTxs {
			sm.metrics.CrossShardTransactions++

			// Record cross-shard communication for each affected account
			for _, crossAccAddr := range crossShardAccounts {
				crossShardID := sm.GetShardForTransaction("", crossAccAddr)
				if crossShardID != primaryShardID {
					shard.CrossShardTxs[crossShardID]++
				}
			}
		}
	}

	// Check if rebalancing is needed
	if sm.config.ShardRebalance {
		sm.checkRebalancing()
	}
}

// GetShardMetrics returns metrics for all shards
func (sm *ShardingManager) GetShardMetrics() *ShardingMetrics {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Update metrics
	totalTxs := sm.metrics.TotalTransactions
	if totalTxs == 0 {
		totalTxs = 1 // Avoid division by zero
	}

	// Calculate load imbalance (standard deviation of load factors)
	loadFactors := make([]float64, sm.config.ShardCount)
	avgLoad := 1.0 / float64(sm.config.ShardCount)
	var sumSquaredDiff float64

	for i, shard := range sm.shards {
		loadFactors[i] = shard.LoadFactor
		diff := shard.LoadFactor - avgLoad
		sumSquaredDiff += diff * diff
	}

	sm.metrics.LoadImbalance = 0
	if sm.config.ShardCount > 1 {
		sm.metrics.LoadImbalance = float64(sumSquaredDiff) / float64(sm.config.ShardCount)
	}

	sm.metrics.LastMeasured = time.Now()
	return sm.metrics
}

// checkRebalancing checks if shards need to be rebalanced
func (sm *ShardingManager) checkRebalancing() {
	// Only rebalance if we have enough data
	if sm.metrics.TotalTransactions < 1000 {
		return
	}

	// Check load imbalance
	metrics := sm.GetShardMetrics()
	if metrics.LoadImbalance > 0.2 { // Threshold for rebalancing
		sm.rebalanceShards()
	}
}

// rebalanceShards rebalances the shards to even out the load
func (sm *ShardingManager) rebalanceShards() {
	// Find most and least loaded shards
	var mostLoadedID, leastLoadedID int
	var maxLoad, minLoad float64 = -1, 2.0

	for id, shard := range sm.shards {
		if shard.LoadFactor > maxLoad {
			maxLoad = shard.LoadFactor
			mostLoadedID = id
		}
		if shard.LoadFactor < minLoad {
			minLoad = shard.LoadFactor
			leastLoadedID = id
		}
	}

	// Only rebalance if there's a significant imbalance
	if maxLoad-minLoad < 0.2 {
		return
	}

	// In a real implementation, we would reallocate validator nodes
	// For this simulation, we'll just log the rebalancing
	fmt.Printf("Rebalancing shards: moving validator from shard %d to shard %d\n", 
		mostLoadedID, leastLoadedID)
	
	// Simulate moving a validator from most loaded to least loaded shard
	if len(sm.shards[mostLoadedID].ValidatorIDs) > sm.config.NodesPerShard/2 {
		// Only move if the shard has enough validators
		validatorToMove := sm.shards[mostLoadedID].ValidatorIDs[0]
		sm.shards[mostLoadedID].ValidatorIDs = sm.shards[mostLoadedID].ValidatorIDs[1:]
		sm.shards[leastLoadedID].ValidatorIDs = append(sm.shards[leastLoadedID].ValidatorIDs, validatorToMove)
	}
}

// generateSimulatedNodeIDs generates simulated node IDs for testing
func generateSimulatedNodeIDs(count int) []string {
	nodeIDs := make([]string, count)
	for i := 0; i < count; i++ {
		nodeIDs[i] = fmt.Sprintf("node%d", i+1)
	}
	return nodeIDs
}

// GetShardStatus returns the status of a specific shard
func (sm *ShardingManager) GetShardStatus(shardID int) (*Shard, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	shard, ok := sm.shards[shardID]
	if !ok {
		return nil, fmt.Errorf("shard %d not found", shardID)
	}

	return shard, nil
}

// GetShardCount returns the number of shards
func (sm *ShardingManager) GetShardCount() int {
	if !sm.config.Enabled {
		return 1
	}
	return sm.config.ShardCount
}

// UpdateShardBlockHeight updates the block height of a shard
func (sm *ShardingManager) UpdateShardBlockHeight(shardID int, height int64) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	shard, ok := sm.shards[shardID]
	if !ok {
		return fmt.Errorf("shard %d not found", shardID)
	}

	shard.BlockHeight = height
	return nil
}
