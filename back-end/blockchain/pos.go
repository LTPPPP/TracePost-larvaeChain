package blockchain

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// PoSConfig contains Proof of Stake specific configuration
type PoSConfig struct {
	MinimumStake     *big.Int        // Minimum stake required to be a validator
	SlashingEnabled  bool            // Whether slashing is enabled
	SlashingRatio    float64         // Percentage of stake to slash on violation
	RewardPerBlock   *big.Int        // Reward per block in the native token
	MaxValidators    int             // Maximum number of validators
	EpochLength      int             // Number of blocks in an epoch
	CooldownPeriod   time.Duration   // Required time after unstaking before stake can be withdrawn
	DelegationEnabled bool           // Whether delegation is enabled
}

// Validator represents a validator in the PoS consensus mechanism
type Validator struct {
	Address       string     // Blockchain address of the validator
	PublicKey     string     // Public key of the validator
	Stake         *big.Int   // Current stake amount
	DelegatedStake *big.Int  // Stake delegated to this validator
	TotalStake    *big.Int   // Total stake (own + delegated)
	Commission    float64    // Commission percentage for delegators
	Uptime        float64    // Percentage of time the validator has been active
	BlocksProposed int64     // Number of blocks proposed by this validator
	BlocksValidated int64    // Number of blocks validated by this validator
	JailUntil     time.Time  // Time until which the validator is jailed (if jailed)
	IsActive      bool       // Whether the validator is currently active
}

// PosValidatorSet manages the set of validators for PoS
type PosValidatorSet struct {
	Validators     map[string]*Validator
	ActiveSet      []string  // List of addresses in the active set
	Config         PoSConfig
	TotalStake     *big.Int   // Total stake across all validators
	LastUpdateTime time.Time  // Last time the validator set was updated
	mutex          sync.RWMutex
}

// Delegation represents a delegation from a delegator to a validator
type Delegation struct {
	DelegatorAddress string     // Address of the delegator
	ValidatorAddress string     // Address of the validator
	Amount           *big.Int   // Amount delegated
	Rewards          *big.Int   // Accumulated rewards
	Since            time.Time  // Time when delegation started
}

// NewValidatorSet creates a new validator set
func NewValidatorSet(config PoSConfig) *PosValidatorSet {
	return &PosValidatorSet{
		Validators:     make(map[string]*Validator),
		ActiveSet:      make([]string, 0),
		Config:         config,
		TotalStake:     big.NewInt(0),
		LastUpdateTime: time.Now(),
	}
}

// AddValidator adds a new validator to the set
func (vs *PosValidatorSet) AddValidator(address, publicKey string, stake *big.Int) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	// Check if validator already exists
	if _, exists := vs.Validators[address]; exists {
		return fmt.Errorf("validator with address %s already exists", address)
	}

	// Check minimum stake
	if stake.Cmp(vs.Config.MinimumStake) < 0 {
		return fmt.Errorf("stake amount %s is below minimum required %s", stake.String(), vs.Config.MinimumStake.String())
	}

	// Add validator
	validator := &Validator{
		Address:        address,
		PublicKey:      publicKey,
		Stake:          new(big.Int).Set(stake),
		DelegatedStake: big.NewInt(0),
		TotalStake:     new(big.Int).Set(stake),
		Commission:     0.1, // Default 10% commission
		Uptime:         100.0, // Start with 100% uptime
		BlocksProposed: 0,
		BlocksValidated: 0,
		IsActive:       true,
	}
	vs.Validators[address] = validator

	// Update total stake
	vs.TotalStake.Add(vs.TotalStake, stake)

	// Check if validator should be in active set
	if len(vs.ActiveSet) < vs.Config.MaxValidators {
		vs.ActiveSet = append(vs.ActiveSet, address)
	} else {
		// Check if stake is higher than any current active validator
		vs.updateActiveSet(address)
	}

	vs.LastUpdateTime = time.Now()
	return nil
}

// UpdateStake updates the stake of an existing validator
func (vs *PosValidatorSet) UpdateStake(address string, newStake *big.Int) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	// Check if validator exists
	validator, exists := vs.Validators[address]
	if !exists {
		return fmt.Errorf("validator with address %s does not exist", address)
	}

	// Check minimum stake
	if newStake.Cmp(vs.Config.MinimumStake) < 0 {
		return fmt.Errorf("stake amount %s is below minimum required %s", newStake.String(), vs.Config.MinimumStake.String())
	}

	// Update total stake
	stakeDiff := new(big.Int).Sub(newStake, validator.Stake)
	vs.TotalStake.Add(vs.TotalStake, stakeDiff)

	// Update validator stake
	validator.Stake = new(big.Int).Set(newStake)
	validator.TotalStake = new(big.Int).Add(validator.Stake, validator.DelegatedStake)

	// Update active set
	vs.updateActiveSet(address)

	vs.LastUpdateTime = time.Now()
	return nil
}

// RemoveValidator removes a validator from the set
func (vs *PosValidatorSet) RemoveValidator(address string) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	// Check if validator exists
	validator, exists := vs.Validators[address]
	if !exists {
		return fmt.Errorf("validator with address %s does not exist", address)
	}

	// Update total stake
	vs.TotalStake.Sub(vs.TotalStake, validator.TotalStake)

	// Remove from active set if present
	for i, addr := range vs.ActiveSet {
		if addr == address {
			vs.ActiveSet = append(vs.ActiveSet[:i], vs.ActiveSet[i+1:]...)
			break
		}
	}

	// Remove validator
	delete(vs.Validators, address)

	vs.LastUpdateTime = time.Now()
	return nil
}

// GetValidator returns a validator by address
func (vs *PosValidatorSet) GetValidator(address string) (*Validator, error) {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	validator, exists := vs.Validators[address]
	if !exists {
		return nil, fmt.Errorf("validator with address %s does not exist", address)
	}

	return validator, nil
}

// GetActiveValidators returns the list of active validators
func (vs *PosValidatorSet) GetActiveValidators() []*Validator {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	validators := make([]*Validator, 0, len(vs.ActiveSet))
	for _, address := range vs.ActiveSet {
		if validator, exists := vs.Validators[address]; exists {
			validators = append(validators, validator)
		}
	}

	return validators
}

// GetAllValidators returns all validators
func (vs *PosValidatorSet) GetAllValidators() []*Validator {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	validators := make([]*Validator, 0, len(vs.Validators))
	for _, validator := range vs.Validators {
		validators = append(validators, validator)
	}

	return validators
}

// IsActiveValidator checks if a validator is in the active set
func (vs *PosValidatorSet) IsActiveValidator(address string) bool {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	for _, addr := range vs.ActiveSet {
		if addr == address {
			return true
		}
	}

	return false
}

// SlashValidator slashes a validator for misbehavior
func (vs *PosValidatorSet) SlashValidator(address string, reason string) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	// Check if slashing is enabled
	if !vs.Config.SlashingEnabled {
		return errors.New("slashing is not enabled")
	}

	// Check if validator exists
	validator, exists := vs.Validators[address]
	if !exists {
		return fmt.Errorf("validator with address %s does not exist", address)
	}

	// Calculate slash amount
	slashAmount := new(big.Int).Mul(validator.Stake, big.NewInt(int64(vs.Config.SlashingRatio*100)))
	slashAmount = slashAmount.Div(slashAmount, big.NewInt(100))

	// Update validator stake
	validator.Stake.Sub(validator.Stake, slashAmount)
	validator.TotalStake.Sub(validator.TotalStake, slashAmount)

	// Update total stake
	vs.TotalStake.Sub(vs.TotalStake, slashAmount)

	// Jail validator
	validator.JailUntil = time.Now().Add(24 * time.Hour) // Jail for 24 hours
	validator.IsActive = false

	// Remove from active set
	for i, addr := range vs.ActiveSet {
		if addr == address {
			vs.ActiveSet = append(vs.ActiveSet[:i], vs.ActiveSet[i+1:]...)
			break
		}
	}

	vs.LastUpdateTime = time.Now()
	return nil
}

// UnjailValidator removes a validator from jail
func (vs *PosValidatorSet) UnjailValidator(address string) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	// Check if validator exists
	validator, exists := vs.Validators[address]
	if !exists {
		return fmt.Errorf("validator with address %s does not exist", address)
	}

	// Check if validator is jailed
	if validator.IsActive || time.Now().Before(validator.JailUntil) {
		return fmt.Errorf("validator is not jailed or jail period has not ended")
	}

	// Unjail validator
	validator.IsActive = true
	validator.JailUntil = time.Time{}

	// Update active set
	vs.updateActiveSet(address)

	vs.LastUpdateTime = time.Now()
	return nil
}

// Delegate delegates stake from a delegator to a validator
func (vs *PosValidatorSet) Delegate(delegatorAddress, validatorAddress string, amount *big.Int) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	// Check if delegation is enabled
	if !vs.Config.DelegationEnabled {
		return errors.New("delegation is not enabled")
	}

	// Check if validator exists
	validator, exists := vs.Validators[validatorAddress]
	if !exists {
		return fmt.Errorf("validator with address %s does not exist", validatorAddress)
	}

	// Check if validator is active
	if !validator.IsActive {
		return fmt.Errorf("validator with address %s is not active", validatorAddress)
	}

	// Update validator delegated stake
	validator.DelegatedStake.Add(validator.DelegatedStake, amount)
	validator.TotalStake.Add(validator.TotalStake, amount)

	// Update total stake
	vs.TotalStake.Add(vs.TotalStake, amount)

	// Update active set
	vs.updateActiveSet(validatorAddress)

	vs.LastUpdateTime = time.Now()
	return nil
}

// Undelegate removes delegated stake from a validator
func (vs *PosValidatorSet) Undelegate(delegatorAddress, validatorAddress string, amount *big.Int) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	// Check if delegation is enabled
	if !vs.Config.DelegationEnabled {
		return errors.New("delegation is not enabled")
	}

	// Check if validator exists
	validator, exists := vs.Validators[validatorAddress]
	if !exists {
		return fmt.Errorf("validator with address %s does not exist", validatorAddress)
	}

	// Check if amount is valid
	if validator.DelegatedStake.Cmp(amount) < 0 {
		return fmt.Errorf("requested undelegation amount exceeds delegated stake")
	}

	// Update validator delegated stake
	validator.DelegatedStake.Sub(validator.DelegatedStake, amount)
	validator.TotalStake.Sub(validator.TotalStake, amount)

	// Update total stake
	vs.TotalStake.Sub(vs.TotalStake, amount)

	// Update active set
	vs.updateActiveSet(validatorAddress)

	vs.LastUpdateTime = time.Now()
	return nil
}

// updateActiveSet updates the active validator set based on stake
func (vs *PosValidatorSet) updateActiveSet(validatorAddress string) {
	// Check if validator is already in active set
	inActiveSet := false
	for _, addr := range vs.ActiveSet {
		if addr == validatorAddress {
			inActiveSet = true
			break
		}
	}

	validator := vs.Validators[validatorAddress]

	// If not active, don't add to active set
	if !validator.IsActive {
		if inActiveSet {
			// Remove from active set
			for i, addr := range vs.ActiveSet {
				if addr == validatorAddress {
					vs.ActiveSet = append(vs.ActiveSet[:i], vs.ActiveSet[i+1:]...)
					break
				}
			}
		}
		return
	}

	// If already in active set, nothing to do
	if inActiveSet {
		return
	}

	// If active set not full, add validator
	if len(vs.ActiveSet) < vs.Config.MaxValidators {
		vs.ActiveSet = append(vs.ActiveSet, validatorAddress)
		return
	}

	// Find validator with lowest stake in active set
	lowestStake := new(big.Int).Set(validator.TotalStake)
	lowestStakeAddr := validatorAddress
	for _, addr := range vs.ActiveSet {
		activeValidator := vs.Validators[addr]
		if activeValidator.TotalStake.Cmp(lowestStake) < 0 {
			lowestStake = activeValidator.TotalStake
			lowestStakeAddr = addr
		}
	}

	// If validator has higher stake than lowest in active set, replace it
	if lowestStakeAddr != validatorAddress {
		for i, addr := range vs.ActiveSet {
			if addr == lowestStakeAddr {
				vs.ActiveSet[i] = validatorAddress
				break
			}
		}
	}
}

// SelectProposer selects a validator to propose the next block
func (vs *PosValidatorSet) SelectProposer(seed int64) (string, error) {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	if len(vs.ActiveSet) == 0 {
		return "", errors.New("no active validators available")
	}

	// Simple round-robin selection based on seed
	// In a real implementation, would use VRF (Verifiable Random Function)
	index := seed % int64(len(vs.ActiveSet))
	return vs.ActiveSet[index], nil
}

// RewardValidators distributes rewards to validators for a finalized block
func (vs *PosValidatorSet) RewardValidators(blockProposer string, blockValidators []string) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	// Check if proposer exists
	proposer, exists := vs.Validators[blockProposer]
	if !exists {
		return fmt.Errorf("proposer with address %s does not exist", blockProposer)
	}

	// Update proposer metrics
	proposer.BlocksProposed++

	// Reward proposer (40% of block reward)
	proposerReward := new(big.Int).Mul(vs.Config.RewardPerBlock, big.NewInt(40))
	proposerReward = proposerReward.Div(proposerReward, big.NewInt(100))
	proposer.Stake.Add(proposer.Stake, proposerReward)
	proposer.TotalStake.Add(proposer.TotalStake, proposerReward)

	// Distribute remaining rewards to validators
	// 60% of block reward split among validators
	validatorCount := len(blockValidators)
	if validatorCount > 0 {
		totalValidatorReward := new(big.Int).Mul(vs.Config.RewardPerBlock, big.NewInt(60))
		totalValidatorReward = totalValidatorReward.Div(totalValidatorReward, big.NewInt(100))
		rewardPerValidator := new(big.Int).Div(totalValidatorReward, big.NewInt(int64(validatorCount)))

		for _, validatorAddr := range blockValidators {
			validator, validatorExists := vs.Validators[validatorAddr]
			if validatorExists {
				validator.BlocksValidated++
				validator.Stake.Add(validator.Stake, rewardPerValidator)
				validator.TotalStake.Add(validator.TotalStake, rewardPerValidator)
			}
		}
	}

	// Update total stake
	rewardTotal := new(big.Int).Set(vs.Config.RewardPerBlock)
	vs.TotalStake.Add(vs.TotalStake, rewardTotal)

	vs.LastUpdateTime = time.Now()
	return nil
}

// GetTopValidators returns the validators with the highest stake
func (vs *PosValidatorSet) GetTopValidators(count int) []*Validator {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	// Copy validators to a slice for sorting
	validators := make([]*Validator, 0, len(vs.Validators))
	for _, validator := range vs.Validators {
		validators = append(validators, validator)
	}

	// Sort validators by total stake (descending)
	for i := 0; i < len(validators); i++ {
		for j := i + 1; j < len(validators); j++ {
			if validators[j].TotalStake.Cmp(validators[i].TotalStake) > 0 {
				validators[i], validators[j] = validators[j], validators[i]
			}
		}
	}

	// Return top count validators
	if count > len(validators) {
		count = len(validators)
	}
	return validators[:count]
}
