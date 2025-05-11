// consensus.go
package blockchain

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ValidatorSet represents a set of validators
type ValidatorSet struct {
	Validators       []string // List of validator node IDs
	CurrentValidator int      // Index of current primary validator
	NextRotation     time.Time // When to rotate to next validator
}

// ConsensusState represents the state of the consensus
type ConsensusState struct {
	Height           int64
	Round            int
	Step             string
	LockedRound      int
	LockedBlock      string
	ValidRound       int
	ValidBlock       string
}

// ConsensusMetrics tracks performance metrics for the consensus engine
type ConsensusMetrics struct {
	TotalBlocksProduced    int
	BlockProductionRate    float64  // blocks per minute
	AverageBlockSize       int      // in bytes
	AvgTransactionsPerBlock int
	ValidationTimeMs       int64    // average validation time in ms
	ConsensusTimeMs        int64    // average consensus time in ms
	LastUpdated            time.Time
}

// ConsensusEngine represents a consensus engine for the blockchain
type ConsensusEngine struct {
	Config          ConsensusConfig
	ValidatorSet    *ValidatorSet
	ShardingManager *ShardingManager
	State           ConsensusState
	
	// Advanced consensus parameters
	CurrentLeader   string
	LeaderChangeTime time.Time
	VoteResults     map[string]int // Track votes for each proposed block
	CommittedBlocks map[string]bool // Track committed blocks
	
	// Performance metrics
	Metrics ConsensusMetrics
	
	// DPoS specific fields
	Delegates      []*Delegate
	ActiveValidators []string
	ValidatorVotes map[string]int
	blockValidationChannel chan BlockValidationRequest
	electionTicker *time.Ticker
	ShardConfig    *ShardingConfig
	
	// Mutex for concurrent access
	mutex sync.RWMutex
}

// Delegate represents a DPoS delegate
type Delegate struct {
	NodeID      string
	Address     string
	PublicKey   string
	VoteCount   int
	IsActive    bool
	Performance DelegatePerformance
}

// DelegatePerformance tracks delegate performance metrics
type DelegatePerformance struct {
	BlocksProduced      int
	BlocksMissed        int
	AvgResponseTime     int64 // in milliseconds
	LastHeartbeat       time.Time
	SuccessfulValidations int
	FailedValidations   int
}

// We'll use the ShardingConfig from sharding.go

// BlockValidationRequest represents a request to validate a block
type BlockValidationRequest struct {
	Block        *Block
	ResponderChan chan BlockValidationResponse
}

// BlockValidationResponse represents the response to a block validation request
type BlockValidationResponse struct {
	IsValid      bool
	ValidatorID  string
	Signature    string
	ErrorMessage string
}

// Block represents a blockchain block with sharding support
type Block struct {
	Header     BlockHeader
	Transactions []Transaction
	ShardID    string
	CrossShardRefs []string // References to blocks in other shards
	Signatures map[string]string // ValidatorID -> Signature
}

// BlockHeader contains the header information for a block
type BlockHeader struct {
	Height      int64
	PrevHash    string
	MerkleRoot  string
	Timestamp   time.Time
	Producer    string
	ShardID     string
	Difficulty  int
}

// NewConsensusEngine creates a new consensus engine instance
func NewConsensusEngine(config ConsensusConfig) *ConsensusEngine {
	engine := &ConsensusEngine{
		Config:         config,
		Delegates:      make([]*Delegate, 0),
		ActiveValidators: make([]string, 0),
		ValidatorVotes: make(map[string]int),
		blockValidationChannel: make(chan BlockValidationRequest, 100),
	}
		// Initialize sharding configuration if enabled
	if config.Type == "dpos" || config.Type == "hybrid" {
		engine.ShardConfig = &ShardingConfig{
			Enabled:      true,
			ShardCount:   3, // Default to 3 shards: producers, farmers, processors
			NodesPerShard: 5,
			ShardAssignments: make(map[string]string),
			CrossShardProtocol: "relay",
			ShardRebalanceInterval: 100,
		}
	}
	
	// Set up delegates for DPoS
	if config.Type == "dpos" || config.Type == "hybrid" {
		for _, validatorNode := range config.ValidatorNodes {
			delegate := &Delegate{
				NodeID:    validatorNode,
				IsActive:  true,
				Performance: DelegatePerformance{
					LastHeartbeat: time.Now(),
				},
			}
			engine.Delegates = append(engine.Delegates, delegate)
		}
		
		// Set initial active validators
		engine.ActiveValidators = make([]string, 0, config.MinValidations)
		for i := 0; i < min(config.MinValidations, len(engine.Delegates)); i++ {
			engine.ActiveValidators = append(engine.ActiveValidators, engine.Delegates[i].NodeID)
		}
	}
	
	return engine
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Start starts the consensus engine
func (ce *ConsensusEngine) Start() error {
	// Start block validation worker
	go ce.blockValidationWorker()
	
	// If DPoS is enabled, start the delegate election process
	if ce.Config.Type == "dpos" || ce.Config.Type == "hybrid" {
		ce.electionTicker = time.NewTicker(time.Duration(ce.Config.EpochLength) * time.Second)
		go ce.delegateElectionWorker()
	}
	
	return nil
}

// Stop stops the consensus engine
func (ce *ConsensusEngine) Stop() {
	// Stop election ticker if running
	if ce.electionTicker != nil {
		ce.electionTicker.Stop()
	}
	
	// Close validation channel
	close(ce.blockValidationChannel)
}

// blockValidationWorker processes block validation requests
func (ce *ConsensusEngine) blockValidationWorker() {
	for req := range ce.blockValidationChannel {
		// Simulate validation based on consensus type
		response := BlockValidationResponse{
			IsValid:     true,
			ValidatorID: "node-" + fmt.Sprint(rand.Intn(1000)),
			Signature:   hex.EncodeToString([]byte("signature")),
		}
		
		// Reply with validation result
		req.ResponderChan <- response
	}
}

// delegateElectionWorker handles periodic delegate elections
func (ce *ConsensusEngine) delegateElectionWorker() {
	for range ce.electionTicker.C {
		ce.electDelegates()
	}
}

// electDelegates performs the DPoS delegate election
func (ce *ConsensusEngine) electDelegates() {
	ce.mutex.Lock()
	defer ce.mutex.Unlock()
	
	// Sort delegates by vote count (simple implementation)
	// In a real system, this would include stake-weighted voting
	sortedDelegates := make([]*Delegate, len(ce.Delegates))
	copy(sortedDelegates, ce.Delegates)
	
	// Simple bubble sort for demonstration
	for i := 0; i < len(sortedDelegates); i++ {
		for j := 0; j < len(sortedDelegates)-i-1; j++ {
			if sortedDelegates[j].VoteCount < sortedDelegates[j+1].VoteCount {
				sortedDelegates[j], sortedDelegates[j+1] = sortedDelegates[j+1], sortedDelegates[j]
			}
		}
	}
	
	// Clear active validators
	ce.ActiveValidators = make([]string, 0, ce.Config.MinValidations)
	
	// Select top delegates as active validators
	for i := 0; i < min(ce.Config.MinValidations, len(sortedDelegates)); i++ {
		sortedDelegates[i].IsActive = true
		ce.ActiveValidators = append(ce.ActiveValidators, sortedDelegates[i].NodeID)
	}
	
	// Mark remaining delegates as inactive
	for i := ce.Config.MinValidations; i < len(sortedDelegates); i++ {
		sortedDelegates[i].IsActive = false
	}
}

// ValidateBlock validates a block according to the consensus rules
func (ce *ConsensusEngine) ValidateBlock(block *Block) (bool, error) {
	// Create response channel
	respChan := make(chan BlockValidationResponse)
	
	// Submit validation request
	ce.blockValidationChannel <- BlockValidationRequest{
		Block:        block,
		ResponderChan: respChan,
	}
	
	// Wait for response
	response := <-respChan
	
	if !response.IsValid {
		return false, errors.New(response.ErrorMessage)
	}
	
	return true, nil
}

// AssignShardToNode assigns a shard to a node
func (ce *ConsensusEngine) AssignShardToNode(nodeID string) string {
	if !ce.ShardConfig.Enabled {
		return "default"
	}
	
	ce.mutex.Lock()
	defer ce.mutex.Unlock()
	
	// Check if node already has a shard assignment
	if shardID, exists := ce.ShardConfig.ShardAssignments[nodeID]; exists {
		return shardID
	}
	
	// Count nodes per shard
	shardCounts := make(map[string]int)
	for _, shardID := range ce.ShardConfig.ShardAssignments {
		shardCounts[shardID]++
	}
	
	// Find shard with fewest nodes
	minCount := ce.ShardConfig.NodesPerShard
	targetShard := fmt.Sprintf("shard-%d", rand.Intn(ce.ShardConfig.ShardCount))
	
	for i := 0; i < ce.ShardConfig.ShardCount; i++ {
		shardID := fmt.Sprintf("shard-%d", i)
		count := shardCounts[shardID]
		
		if count < minCount {
			minCount = count
			targetShard = shardID
		}
	}
		// Assign node to shard
	ce.ShardConfig.ShardAssignments[nodeID] = targetShard
	
	return targetShard
}

// VoteForDelegate allows voting for delegates in DPoS
func (ce *ConsensusEngine) VoteForDelegate(voterID, delegateID string, voteWeight int) error {
	ce.mutex.Lock()
	defer ce.mutex.Unlock()
	
	// Find delegate
	var targetDelegate *Delegate
	for _, delegate := range ce.Delegates {
		if delegate.NodeID == delegateID {
			targetDelegate = delegate
			break
		}
	}
	
	if targetDelegate == nil {
		return errors.New("delegate not found")
	}
	
	// Add votes
	targetDelegate.VoteCount += voteWeight
	
	return nil
}

// GetActiveValidators returns the current set of active validators
func (ce *ConsensusEngine) GetActiveValidators() []string {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()
	
	// Return a copy to prevent external modification
	validators := make([]string, len(ce.ActiveValidators))
	copy(validators, ce.ActiveValidators)
	
	return validators
}

// GetDelegatePerformance returns performance metrics for a delegate
func (ce *ConsensusEngine) GetDelegatePerformance(delegateID string) (*DelegatePerformance, error) {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()
	
	// Find delegate
	for _, delegate := range ce.Delegates {
		if delegate.NodeID == delegateID {
			// Return a copy to prevent external modification
			perfCopy := delegate.Performance
			return &perfCopy, nil
		}
	}
	
	return nil, errors.New("delegate not found")
}

// Returns the shard ID for a given entity type
func (ce *ConsensusEngine) GetShardForEntityType(entityType string) string {
	switch entityType {
	case "hatchery":
		return "shard-0"
	case "farmer":
		return "shard-1"
	case "processor":
		return "shard-2"
	default:
		return "shard-0"
	}
}
