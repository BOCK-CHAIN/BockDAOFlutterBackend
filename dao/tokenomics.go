package dao

import (
	"fmt"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
)

// TokenomicsManager manages token distribution, vesting, and staking
type TokenomicsManager struct {
	governanceState  *GovernanceState
	tokenState       *GovernanceToken
	distributions    map[string]*TokenDistribution
	vestingSchedules map[string]*VestingSchedule
	stakingPools     map[string]*StakingPool
	config           *TokenomicsConfig
}

// TokenDistribution represents a token allocation category
type TokenDistribution struct {
	Category    DistributionCategory
	Allocation  uint64 // Total tokens allocated
	Distributed uint64 // Tokens already distributed
	Recipients  map[string]*DistributionRecipient
	VestingType VestingType
	StartTime   int64
	CliffPeriod int64 // Time before any tokens can be claimed
	Duration    int64 // Total vesting duration
}

// DistributionRecipient represents an individual recipient in a distribution
type DistributionRecipient struct {
	Address     crypto.PublicKey
	Allocation  uint64 // Total tokens allocated to this recipient
	Claimed     uint64 // Tokens already claimed
	LastClaimed int64  // Timestamp of last claim
	VestingID   string // Reference to vesting schedule
}

// VestingSchedule manages time-locked token releases
type VestingSchedule struct {
	ID          string
	Beneficiary crypto.PublicKey
	TotalAmount uint64
	StartTime   int64
	CliffTime   int64
	Duration    int64
	Released    uint64
	Revoked     bool
	VestingType VestingType
}

// StakingPool represents a staking mechanism
type StakingPool struct {
	ID             string
	Name           string
	TotalStaked    uint64
	RewardRate     uint64 // Rewards per second per token
	LastUpdateTime int64
	RewardPerToken uint64
	MinStakeAmount uint64
	LockupPeriod   int64 // Minimum staking duration
	Stakers        map[string]*StakerInfo
	Active         bool
}

// StakerInfo represents an individual staker's information
type StakerInfo struct {
	Address            crypto.PublicKey
	StakedAmount       uint64
	RewardPerTokenPaid uint64
	Rewards            uint64
	StakeTime          int64
	UnlockTime         int64
}

// TokenomicsConfig contains configuration for tokenomics
type TokenomicsConfig struct {
	TotalSupply            uint64
	FounderAllocation      uint64 // Percentage in basis points (10000 = 100%)
	TeamAllocation         uint64
	CommunityAllocation    uint64
	TreasuryAllocation     uint64
	EcosystemAllocation    uint64
	DefaultVestingCliff    int64  // Default cliff period in seconds
	DefaultVestingDuration int64  // Default vesting duration in seconds
	StakingRewardRate      uint64 // Default staking reward rate
	MinStakeAmount         uint64 // Minimum amount to stake
}

// NewTokenomicsManager creates a new tokenomics manager
func NewTokenomicsManager(governanceState *GovernanceState, tokenState *GovernanceToken) *TokenomicsManager {
	return &TokenomicsManager{
		governanceState:  governanceState,
		tokenState:       tokenState,
		distributions:    make(map[string]*TokenDistribution),
		vestingSchedules: make(map[string]*VestingSchedule),
		stakingPools:     make(map[string]*StakingPool),
		config:           NewDefaultTokenomicsConfig(),
	}
}

// NewDefaultTokenomicsConfig creates default tokenomics configuration
func NewDefaultTokenomicsConfig() *TokenomicsConfig {
	return &TokenomicsConfig{
		TotalSupply:            1000000000,          // 1 billion tokens
		FounderAllocation:      2000,                // 20%
		TeamAllocation:         1500,                // 15%
		CommunityAllocation:    3000,                // 30%
		TreasuryAllocation:     2000,                // 20%
		EcosystemAllocation:    1500,                // 15%
		DefaultVestingCliff:    365 * 24 * 3600,     // 1 year cliff
		DefaultVestingDuration: 4 * 365 * 24 * 3600, // 4 year vesting
		StakingRewardRate:      100,                 // 100 tokens per second per token (adjust as needed)
		MinStakeAmount:         1000,                // Minimum 1000 tokens to stake
	}
}

// InitializeTokenDistribution sets up initial token distribution
func (tm *TokenomicsManager) InitializeTokenDistribution() error {
	config := tm.config

	// Calculate allocations based on total supply
	founderTokens := (config.TotalSupply * config.FounderAllocation) / 10000
	teamTokens := (config.TotalSupply * config.TeamAllocation) / 10000
	communityTokens := (config.TotalSupply * config.CommunityAllocation) / 10000
	treasuryTokens := (config.TotalSupply * config.TreasuryAllocation) / 10000
	ecosystemTokens := (config.TotalSupply * config.EcosystemAllocation) / 10000

	// Create distribution categories
	distributions := map[DistributionCategory]uint64{
		DistributionFounders:  founderTokens,
		DistributionTeam:      teamTokens,
		DistributionCommunity: communityTokens,
		DistributionTreasury:  treasuryTokens,
		DistributionEcosystem: ecosystemTokens,
	}

	now := time.Now().Unix()

	for category, allocation := range distributions {
		categoryName := tm.getCategoryName(category)

		distribution := &TokenDistribution{
			Category:    category,
			Allocation:  allocation,
			Distributed: 0,
			Recipients:  make(map[string]*DistributionRecipient),
			VestingType: tm.getDefaultVestingType(category),
			StartTime:   now,
			CliffPeriod: tm.getDefaultCliffPeriod(category),
			Duration:    tm.getDefaultDuration(category),
		}

		tm.distributions[categoryName] = distribution
	}

	// Update total supply in token state
	tm.tokenState.TotalSupply = config.TotalSupply

	return nil
}

// AddDistributionRecipient adds a recipient to a distribution category
func (tm *TokenomicsManager) AddDistributionRecipient(category DistributionCategory, recipient crypto.PublicKey, amount uint64) error {
	categoryName := tm.getCategoryName(category)
	distribution, exists := tm.distributions[categoryName]
	if !exists {
		return NewDAOError(ErrInvalidProposal, "distribution category not found", nil)
	}

	// Check if allocation would exceed category limit
	if distribution.Distributed+amount > distribution.Allocation {
		return NewDAOError(ErrInsufficientTokens, "allocation exceeds category limit", nil)
	}

	recipientStr := recipient.String()

	// Create vesting schedule if needed
	var vestingID string
	if distribution.VestingType != VestingTypeImmediate {
		vestingID = tm.createVestingSchedule(recipient, amount, distribution.VestingType,
			distribution.StartTime, distribution.CliffPeriod, distribution.Duration)
	}

	// Add recipient
	distribution.Recipients[recipientStr] = &DistributionRecipient{
		Address:     recipient,
		Allocation:  amount,
		Claimed:     0,
		LastClaimed: 0,
		VestingID:   vestingID,
	}

	distribution.Distributed += amount

	// If immediate vesting, distribute tokens now
	if distribution.VestingType == VestingTypeImmediate {
		return tm.distributeTokensImmediately(recipient, amount)
	}

	return nil
}

// createVestingSchedule creates a new vesting schedule
func (tm *TokenomicsManager) createVestingSchedule(beneficiary crypto.PublicKey, amount uint64, vestingType VestingType, startTime, cliffPeriod, duration int64) string {
	vestingID := fmt.Sprintf("vesting_%s_%d", beneficiary.String()[:8], time.Now().UnixNano())

	schedule := &VestingSchedule{
		ID:          vestingID,
		Beneficiary: beneficiary,
		TotalAmount: amount,
		StartTime:   startTime,
		CliffTime:   startTime + cliffPeriod,
		Duration:    duration,
		Released:    0,
		Revoked:     false,
		VestingType: vestingType,
	}

	tm.vestingSchedules[vestingID] = schedule
	return vestingID
}

// distributeTokensImmediately distributes tokens without vesting
func (tm *TokenomicsManager) distributeTokensImmediately(recipient crypto.PublicKey, amount uint64) error {
	recipientStr := recipient.String()

	// Mint tokens to recipient
	if err := tm.tokenState.Mint(recipientStr, amount); err != nil {
		return err
	}

	// Update token holder record
	if holder, exists := tm.governanceState.TokenHolders[recipientStr]; exists {
		holder.Balance += amount
	} else {
		tm.governanceState.TokenHolders[recipientStr] = &TokenHolder{
			Address:    recipient,
			Balance:    amount,
			Staked:     0,
			Reputation: 0,
			JoinedAt:   time.Now().Unix(),
			LastActive: time.Now().Unix(),
		}
	}

	return nil
}

// ClaimVestedTokens allows a beneficiary to claim vested tokens
func (tm *TokenomicsManager) ClaimVestedTokens(vestingID string, beneficiary crypto.PublicKey) (uint64, error) {
	schedule, exists := tm.vestingSchedules[vestingID]
	if !exists {
		return 0, NewDAOError(ErrProposalNotFound, "vesting schedule not found", nil)
	}

	if schedule.Beneficiary.String() != beneficiary.String() {
		return 0, NewDAOError(ErrUnauthorized, "not authorized to claim from this vesting schedule", nil)
	}

	if schedule.Revoked {
		return 0, NewDAOError(ErrInvalidProposal, "vesting schedule has been revoked", nil)
	}

	// Calculate vested amount
	vestedAmount := tm.calculateVestedAmount(schedule)
	claimableAmount := vestedAmount - schedule.Released

	if claimableAmount == 0 {
		return 0, NewDAOError(ErrInsufficientTokens, "no tokens available to claim", nil)
	}

	// Mint tokens to beneficiary
	beneficiaryStr := beneficiary.String()
	if err := tm.tokenState.Mint(beneficiaryStr, claimableAmount); err != nil {
		return 0, err
	}

	// Update vesting schedule
	schedule.Released += claimableAmount

	// Update token holder record
	if holder, exists := tm.governanceState.TokenHolders[beneficiaryStr]; exists {
		holder.Balance += claimableAmount
		holder.LastActive = time.Now().Unix()
	} else {
		tm.governanceState.TokenHolders[beneficiaryStr] = &TokenHolder{
			Address:    beneficiary,
			Balance:    claimableAmount,
			Staked:     0,
			Reputation: 0,
			JoinedAt:   time.Now().Unix(),
			LastActive: time.Now().Unix(),
		}
	}

	return claimableAmount, nil
}

// calculateVestedAmount calculates the amount of tokens vested at current time
func (tm *TokenomicsManager) calculateVestedAmount(schedule *VestingSchedule) uint64 {
	now := time.Now().Unix()

	// If before cliff, no tokens are vested
	if now < schedule.CliffTime {
		return 0
	}

	// If after full vesting period, all tokens are vested
	if now >= schedule.StartTime+schedule.Duration {
		return schedule.TotalAmount
	}

	// Calculate linear vesting
	switch schedule.VestingType {
	case VestingTypeLinear:
		elapsed := now - schedule.CliffTime
		totalVestingTime := schedule.Duration - (schedule.CliffTime - schedule.StartTime)
		return (schedule.TotalAmount * uint64(elapsed)) / uint64(totalVestingTime)

	case VestingTypeCliff:
		// All tokens vest at cliff time
		if now >= schedule.CliffTime {
			return schedule.TotalAmount
		}
		return 0

	default:
		// Default to linear vesting
		elapsed := now - schedule.CliffTime
		totalVestingTime := schedule.Duration - (schedule.CliffTime - schedule.StartTime)
		return (schedule.TotalAmount * uint64(elapsed)) / uint64(totalVestingTime)
	}
}

// CreateStakingPool creates a new staking pool
func (tm *TokenomicsManager) CreateStakingPool(poolID, name string, rewardRate, minStakeAmount, lockupPeriod uint64) error {
	if _, exists := tm.stakingPools[poolID]; exists {
		return NewDAOError(ErrInvalidProposal, "staking pool already exists", nil)
	}

	pool := &StakingPool{
		ID:             poolID,
		Name:           name,
		TotalStaked:    0,
		RewardRate:     rewardRate,
		LastUpdateTime: time.Now().Unix(),
		RewardPerToken: 0,
		MinStakeAmount: minStakeAmount,
		LockupPeriod:   int64(lockupPeriod),
		Stakers:        make(map[string]*StakerInfo),
		Active:         true,
	}

	tm.stakingPools[poolID] = pool
	return nil
}

// StakeTokens stakes tokens in a staking pool
func (tm *TokenomicsManager) StakeTokens(poolID string, staker crypto.PublicKey, amount uint64, lockDuration int64) error {
	pool, exists := tm.stakingPools[poolID]
	if !exists {
		return NewDAOError(ErrProposalNotFound, "staking pool not found", nil)
	}

	if !pool.Active {
		return NewDAOError(ErrInvalidProposal, "staking pool is not active", nil)
	}

	if amount < pool.MinStakeAmount {
		return NewDAOError(ErrInsufficientTokens, "stake amount below minimum", nil)
	}

	stakerStr := staker.String()

	// Check if user has sufficient balance
	if tm.tokenState.GetBalance(stakerStr) < amount {
		return NewDAOError(ErrInsufficientTokens, "insufficient balance to stake", nil)
	}

	// Update pool rewards before staking
	tm.updatePoolRewards(pool)

	// Transfer tokens from user balance to staked
	tm.tokenState.Balances[stakerStr] -= amount

	// Update or create staker info
	if stakerInfo, exists := pool.Stakers[stakerStr]; exists {
		// Claim existing rewards before adding more stake
		tm.claimStakingRewards(pool, stakerInfo)
		stakerInfo.StakedAmount += amount
		stakerInfo.RewardPerTokenPaid = pool.RewardPerToken
		if lockDuration > 0 {
			stakerInfo.UnlockTime = time.Now().Unix() + lockDuration
		}
	} else {
		unlockTime := int64(0)
		if lockDuration > 0 {
			unlockTime = time.Now().Unix() + lockDuration
		} else if pool.LockupPeriod > 0 {
			unlockTime = time.Now().Unix() + pool.LockupPeriod
		}

		pool.Stakers[stakerStr] = &StakerInfo{
			Address:            staker,
			StakedAmount:       amount,
			RewardPerTokenPaid: pool.RewardPerToken,
			Rewards:            0,
			StakeTime:          time.Now().Unix(),
			UnlockTime:         unlockTime,
		}
	}

	// Update pool total
	pool.TotalStaked += amount

	// Update token holder staked amount
	if holder, exists := tm.governanceState.TokenHolders[stakerStr]; exists {
		holder.Staked += amount
		holder.Balance -= amount
		holder.LastActive = time.Now().Unix()
	}

	return nil
}

// UnstakeTokens unstakes tokens from a staking pool
func (tm *TokenomicsManager) UnstakeTokens(poolID string, staker crypto.PublicKey, amount uint64) error {
	pool, exists := tm.stakingPools[poolID]
	if !exists {
		return NewDAOError(ErrProposalNotFound, "staking pool not found", nil)
	}

	stakerStr := staker.String()
	stakerInfo, exists := pool.Stakers[stakerStr]
	if !exists {
		return NewDAOError(ErrProposalNotFound, "staker not found in pool", nil)
	}

	if stakerInfo.StakedAmount < amount {
		return NewDAOError(ErrInsufficientTokens, "insufficient staked amount", nil)
	}

	// Check lockup period
	now := time.Now().Unix()
	if stakerInfo.UnlockTime > 0 && now < stakerInfo.UnlockTime {
		return NewDAOError(ErrInvalidProposal, "tokens are still locked", nil)
	}

	// Update pool rewards before unstaking
	tm.updatePoolRewards(pool)

	// Claim rewards before unstaking
	tm.claimStakingRewards(pool, stakerInfo)

	// Update staker info
	stakerInfo.StakedAmount -= amount
	stakerInfo.RewardPerTokenPaid = pool.RewardPerToken

	// Update pool total
	pool.TotalStaked -= amount

	// Return tokens to user balance
	tm.tokenState.Balances[stakerStr] += amount

	// Update token holder record
	if holder, exists := tm.governanceState.TokenHolders[stakerStr]; exists {
		holder.Staked -= amount
		holder.Balance += amount
		holder.LastActive = now
	}

	// Remove staker if no tokens left
	if stakerInfo.StakedAmount == 0 {
		delete(pool.Stakers, stakerStr)
	}

	return nil
}

// ClaimStakingRewards claims accumulated staking rewards
func (tm *TokenomicsManager) ClaimStakingRewards(poolID string, staker crypto.PublicKey) (uint64, error) {
	pool, exists := tm.stakingPools[poolID]
	if !exists {
		return 0, NewDAOError(ErrProposalNotFound, "staking pool not found", nil)
	}

	stakerStr := staker.String()
	stakerInfo, exists := pool.Stakers[stakerStr]
	if !exists {
		return 0, NewDAOError(ErrProposalNotFound, "staker not found in pool", nil)
	}

	// Update pool rewards
	tm.updatePoolRewards(pool)

	// Calculate and claim rewards
	rewards := tm.claimStakingRewards(pool, stakerInfo)

	if rewards > 0 {
		// Mint reward tokens
		if err := tm.tokenState.Mint(stakerStr, rewards); err != nil {
			return 0, err
		}

		// Update token holder record
		if holder, exists := tm.governanceState.TokenHolders[stakerStr]; exists {
			holder.Balance += rewards
			holder.LastActive = time.Now().Unix()
		}
	}

	return rewards, nil
}

// updatePoolRewards updates the reward calculations for a staking pool
func (tm *TokenomicsManager) updatePoolRewards(pool *StakingPool) {
	now := time.Now().Unix()

	if pool.TotalStaked > 0 {
		timeDiff := now - pool.LastUpdateTime
		rewardIncrease := uint64(timeDiff) * pool.RewardRate
		pool.RewardPerToken += rewardIncrease / pool.TotalStaked
	}

	pool.LastUpdateTime = now
}

// claimStakingRewards calculates and claims rewards for a staker
func (tm *TokenomicsManager) claimStakingRewards(pool *StakingPool, stakerInfo *StakerInfo) uint64 {
	earned := stakerInfo.StakedAmount * (pool.RewardPerToken - stakerInfo.RewardPerTokenPaid)
	totalRewards := stakerInfo.Rewards + earned

	stakerInfo.Rewards = 0
	stakerInfo.RewardPerTokenPaid = pool.RewardPerToken

	return totalRewards
}

// Helper methods

func (tm *TokenomicsManager) getCategoryName(category DistributionCategory) string {
	switch category {
	case DistributionFounders:
		return "founders"
	case DistributionTeam:
		return "team"
	case DistributionCommunity:
		return "community"
	case DistributionTreasury:
		return "treasury"
	case DistributionEcosystem:
		return "ecosystem"
	default:
		return "unknown"
	}
}

func (tm *TokenomicsManager) getDefaultVestingType(category DistributionCategory) VestingType {
	switch category {
	case DistributionFounders:
		return VestingTypeLinear
	case DistributionTeam:
		return VestingTypeLinear
	case DistributionCommunity:
		return VestingTypeImmediate
	case DistributionTreasury:
		return VestingTypeImmediate
	case DistributionEcosystem:
		return VestingTypeLinear
	default:
		return VestingTypeLinear
	}
}

func (tm *TokenomicsManager) getDefaultCliffPeriod(category DistributionCategory) int64 {
	switch category {
	case DistributionFounders:
		return tm.config.DefaultVestingCliff
	case DistributionTeam:
		return tm.config.DefaultVestingCliff / 2 // 6 months for team
	case DistributionCommunity:
		return 0
	case DistributionTreasury:
		return 0
	case DistributionEcosystem:
		return tm.config.DefaultVestingCliff / 4 // 3 months for ecosystem
	default:
		return tm.config.DefaultVestingCliff
	}
}

func (tm *TokenomicsManager) getDefaultDuration(category DistributionCategory) int64 {
	switch category {
	case DistributionFounders:
		return tm.config.DefaultVestingDuration
	case DistributionTeam:
		return tm.config.DefaultVestingDuration / 2 // 2 years for team
	case DistributionCommunity:
		return 0
	case DistributionTreasury:
		return 0
	case DistributionEcosystem:
		return tm.config.DefaultVestingDuration
	default:
		return tm.config.DefaultVestingDuration
	}
}

// Getter methods

// GetDistribution returns a distribution by category
func (tm *TokenomicsManager) GetDistribution(category DistributionCategory) (*TokenDistribution, bool) {
	categoryName := tm.getCategoryName(category)
	distribution, exists := tm.distributions[categoryName]
	return distribution, exists
}

// GetVestingSchedule returns a vesting schedule by ID
func (tm *TokenomicsManager) GetVestingSchedule(vestingID string) (*VestingSchedule, bool) {
	schedule, exists := tm.vestingSchedules[vestingID]
	return schedule, exists
}

// GetStakingPool returns a staking pool by ID
func (tm *TokenomicsManager) GetStakingPool(poolID string) (*StakingPool, bool) {
	pool, exists := tm.stakingPools[poolID]
	return pool, exists
}

// GetStakerInfo returns staker information for a specific pool
func (tm *TokenomicsManager) GetStakerInfo(poolID string, staker crypto.PublicKey) (*StakerInfo, bool) {
	pool, exists := tm.stakingPools[poolID]
	if !exists {
		return nil, false
	}

	stakerInfo, exists := pool.Stakers[staker.String()]
	return stakerInfo, exists
}

// GetTokenomicsConfig returns the current tokenomics configuration
func (tm *TokenomicsManager) GetTokenomicsConfig() *TokenomicsConfig {
	return tm.config
}

// UpdateTokenomicsConfig updates the tokenomics configuration
func (tm *TokenomicsManager) UpdateTokenomicsConfig(newConfig *TokenomicsConfig) error {
	// Validate configuration
	totalAllocation := newConfig.FounderAllocation + newConfig.TeamAllocation +
		newConfig.CommunityAllocation + newConfig.TreasuryAllocation + newConfig.EcosystemAllocation

	if totalAllocation != 10000 {
		return NewDAOError(ErrInvalidProposal, "total allocation must equal 100% (10000 basis points)", nil)
	}

	tm.config = newConfig
	return nil
}

// ListAllDistributions returns all token distributions
func (tm *TokenomicsManager) ListAllDistributions() map[string]*TokenDistribution {
	return tm.distributions
}

// ListAllVestingSchedules returns all vesting schedules
func (tm *TokenomicsManager) ListAllVestingSchedules() map[string]*VestingSchedule {
	return tm.vestingSchedules
}

// ListAllStakingPools returns all staking pools
func (tm *TokenomicsManager) ListAllStakingPools() map[string]*StakingPool {
	return tm.stakingPools
}

// GetVestingSchedulesByBeneficiary returns all vesting schedules for a beneficiary
func (tm *TokenomicsManager) GetVestingSchedulesByBeneficiary(beneficiary crypto.PublicKey) []*VestingSchedule {
	var schedules []*VestingSchedule
	beneficiaryStr := beneficiary.String()

	for _, schedule := range tm.vestingSchedules {
		if schedule.Beneficiary.String() == beneficiaryStr {
			schedules = append(schedules, schedule)
		}
	}

	return schedules
}

// GetTotalStakedByUser returns total staked amount across all pools for a user
func (tm *TokenomicsManager) GetTotalStakedByUser(user crypto.PublicKey) uint64 {
	userStr := user.String()
	totalStaked := uint64(0)

	for _, pool := range tm.stakingPools {
		if stakerInfo, exists := pool.Stakers[userStr]; exists {
			totalStaked += stakerInfo.StakedAmount
		}
	}

	return totalStaked
}

// GetTotalRewardsByUser returns total pending rewards across all pools for a user
func (tm *TokenomicsManager) GetTotalRewardsByUser(user crypto.PublicKey) uint64 {
	userStr := user.String()
	totalRewards := uint64(0)

	for _, pool := range tm.stakingPools {
		if stakerInfo, exists := pool.Stakers[userStr]; exists {
			// Update pool rewards to get current values
			tm.updatePoolRewards(pool)
			earned := stakerInfo.StakedAmount * (pool.RewardPerToken - stakerInfo.RewardPerTokenPaid)
			totalRewards += stakerInfo.Rewards + earned
		}
	}

	return totalRewards
}
