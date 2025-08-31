package dao

import (
	"math"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// ReputationSystem manages reputation tracking and calculation
type ReputationSystem struct {
	governanceState *GovernanceState
	tokenState      *GovernanceToken
	config          *ReputationConfig
}

// ReputationConfig contains configuration for reputation calculations
type ReputationConfig struct {
	BaseReputation          uint64  // Initial reputation for new members
	ProposalCreationBonus   uint64  // Reputation gained for creating proposals
	VotingParticipation     uint64  // Reputation gained per vote cast
	ProposalPassedBonus     uint64  // Bonus for creating passed proposals
	ProposalRejectedPenalty uint64  // Penalty for creating rejected proposals
	InactivityDecayRate     float64 // Daily decay rate for inactive users (0.01 = 1% per day)
	MaxReputation           uint64  // Maximum reputation cap
	MinReputation           uint64  // Minimum reputation floor
	DecayPeriodDays         int64   // Days of inactivity before decay starts
}

// NewReputationSystem creates a new reputation system
func NewReputationSystem(governanceState *GovernanceState, tokenState *GovernanceToken) *ReputationSystem {
	return &ReputationSystem{
		governanceState: governanceState,
		tokenState:      tokenState,
		config:          NewReputationConfig(),
	}
}

// NewReputationConfig creates default reputation configuration
func NewReputationConfig() *ReputationConfig {
	return &ReputationConfig{
		BaseReputation:          100,
		ProposalCreationBonus:   50,
		VotingParticipation:     10,
		ProposalPassedBonus:     100,
		ProposalRejectedPenalty: 25,
		InactivityDecayRate:     0.005, // 0.5% per day
		MaxReputation:           10000,
		MinReputation:           10,
		DecayPeriodDays:         30, // Start decay after 30 days of inactivity
	}
}

// InitializeReputation sets initial reputation for a token holder
func (rs *ReputationSystem) InitializeReputation(address crypto.PublicKey, tokenBalance uint64) {
	addressStr := address.String()

	if holder, exists := rs.governanceState.TokenHolders[addressStr]; exists {
		// Calculate initial reputation based on token balance and base reputation
		initialReputation := rs.config.BaseReputation + (tokenBalance / 100) // 1 reputation per 100 tokens
		if initialReputation > rs.config.MaxReputation {
			initialReputation = rs.config.MaxReputation
		}

		holder.Reputation = initialReputation
		holder.JoinedAt = time.Now().Unix()
		holder.LastActive = time.Now().Unix()
	}
}

// UpdateReputationForProposalCreation updates reputation when a user creates a proposal
func (rs *ReputationSystem) UpdateReputationForProposalCreation(creator crypto.PublicKey) {
	creatorStr := creator.String()

	if holder, exists := rs.governanceState.TokenHolders[creatorStr]; exists {
		// Add proposal creation bonus
		newReputation := holder.Reputation + rs.config.ProposalCreationBonus
		if newReputation > rs.config.MaxReputation {
			newReputation = rs.config.MaxReputation
		}

		holder.Reputation = newReputation
		holder.LastActive = time.Now().Unix()
	}
}

// UpdateReputationForVoting updates reputation when a user votes
func (rs *ReputationSystem) UpdateReputationForVoting(voter crypto.PublicKey, proposalID types.Hash) {
	voterStr := voter.String()

	if holder, exists := rs.governanceState.TokenHolders[voterStr]; exists {
		// Add voting participation bonus
		newReputation := holder.Reputation + rs.config.VotingParticipation
		if newReputation > rs.config.MaxReputation {
			newReputation = rs.config.MaxReputation
		}

		holder.Reputation = newReputation
		holder.LastActive = time.Now().Unix()
	}
}

// UpdateReputationForProposalOutcome updates reputation based on proposal outcomes
func (rs *ReputationSystem) UpdateReputationForProposalOutcome(proposalID types.Hash) {
	proposal, exists := rs.governanceState.Proposals[proposalID]
	if !exists {
		return
	}

	creatorStr := proposal.Creator.String()
	holder, exists := rs.governanceState.TokenHolders[creatorStr]
	if !exists {
		return
	}

	switch proposal.Status {
	case ProposalStatusPassed:
		// Bonus for successful proposal
		newReputation := holder.Reputation + rs.config.ProposalPassedBonus
		if newReputation > rs.config.MaxReputation {
			newReputation = rs.config.MaxReputation
		}
		holder.Reputation = newReputation

	case ProposalStatusRejected:
		// Penalty for rejected proposal (but not below minimum)
		if holder.Reputation > rs.config.ProposalRejectedPenalty {
			newReputation := holder.Reputation - rs.config.ProposalRejectedPenalty
			if newReputation < rs.config.MinReputation {
				newReputation = rs.config.MinReputation
			}
			holder.Reputation = newReputation
		}
	}
}

// ApplyInactivityDecay applies reputation decay for inactive users
func (rs *ReputationSystem) ApplyInactivityDecay() {
	now := time.Now().Unix()
	decayThreshold := now - (rs.config.DecayPeriodDays * 24 * 3600) // Convert days to seconds

	for _, holder := range rs.governanceState.TokenHolders {
		if holder.LastActive < decayThreshold {
			// Calculate days of inactivity beyond threshold
			inactiveDays := float64(now-holder.LastActive) / (24 * 3600)
			if inactiveDays > float64(rs.config.DecayPeriodDays) {
				excessDays := inactiveDays - float64(rs.config.DecayPeriodDays)

				// Apply exponential decay
				decayFactor := math.Pow(1-rs.config.InactivityDecayRate, excessDays)
				newReputation := uint64(float64(holder.Reputation) * decayFactor)

				if newReputation < rs.config.MinReputation {
					newReputation = rs.config.MinReputation
				}

				holder.Reputation = newReputation
			}
		}
	}
}

// CalculateReputationWeight calculates voting weight based on reputation
func (rs *ReputationSystem) CalculateReputationWeight(voter crypto.PublicKey, requestedWeight uint64) (uint64, error) {
	voterStr := voter.String()
	holder, exists := rs.governanceState.TokenHolders[voterStr]
	if !exists {
		return 0, NewDAOError(ErrUnauthorized, "voter not found in token holders", nil)
	}

	// Maximum voting weight is limited by reputation
	maxWeight := holder.Reputation
	if requestedWeight > maxWeight {
		return 0, NewDAOError(ErrInsufficientTokens, "requested weight exceeds reputation", nil)
	}

	return requestedWeight, nil
}

// CalculateReputationBasedVotingCost calculates the token cost for reputation-based voting
func (rs *ReputationSystem) CalculateReputationBasedVotingCost(voter crypto.PublicKey, weight uint64) (uint64, error) {
	voterStr := voter.String()
	holder, exists := rs.governanceState.TokenHolders[voterStr]
	if !exists {
		return 0, NewDAOError(ErrUnauthorized, "voter not found in token holders", nil)
	}

	voterBalance := rs.tokenState.GetBalance(voterStr)

	// Cost is proportional to the percentage of reputation being used
	// Formula: cost = (weight / reputation) * balance * cost_multiplier
	if holder.Reputation == 0 {
		return 0, NewDAOError(ErrInsufficientTokens, "voter has no reputation", nil)
	}

	// Cost multiplier to make reputation voting meaningful but not prohibitive
	costMultiplier := float64(0.1) // 10% of proportional balance
	reputationRatio := float64(weight) / float64(holder.Reputation)
	cost := uint64(float64(voterBalance) * reputationRatio * costMultiplier)

	// Minimum cost of 1 token to prevent zero-cost voting
	if cost == 0 {
		cost = 1
	}

	return cost, nil
}

// GetReputationRanking returns users sorted by reputation (highest first)
func (rs *ReputationSystem) GetReputationRanking() []*TokenHolder {
	var holders []*TokenHolder

	for _, holder := range rs.governanceState.TokenHolders {
		holders = append(holders, holder)
	}

	// Sort by reputation (descending)
	for i := 0; i < len(holders)-1; i++ {
		for j := i + 1; j < len(holders); j++ {
			if holders[i].Reputation < holders[j].Reputation {
				holders[i], holders[j] = holders[j], holders[i]
			}
		}
	}

	return holders
}

// GetReputationStats returns statistics about the reputation system
func (rs *ReputationSystem) GetReputationStats() *ReputationStats {
	var totalReputation uint64
	var activeUsers uint64
	var maxReputation uint64
	var minReputation uint64 = rs.config.MaxReputation // Start with max for comparison

	now := time.Now().Unix()
	activeThreshold := now - (7 * 24 * 3600) // Active in last 7 days

	for _, holder := range rs.governanceState.TokenHolders {
		totalReputation += holder.Reputation

		if holder.LastActive >= activeThreshold {
			activeUsers++
		}

		if holder.Reputation > maxReputation {
			maxReputation = holder.Reputation
		}

		if holder.Reputation < minReputation {
			minReputation = holder.Reputation
		}
	}

	totalUsers := uint64(len(rs.governanceState.TokenHolders))
	var averageReputation uint64
	if totalUsers > 0 {
		averageReputation = totalReputation / totalUsers
	}

	return &ReputationStats{
		TotalUsers:        totalUsers,
		ActiveUsers:       activeUsers,
		TotalReputation:   totalReputation,
		AverageReputation: averageReputation,
		MaxReputation:     maxReputation,
		MinReputation:     minReputation,
	}
}

// ReputationStats contains statistics about the reputation system
type ReputationStats struct {
	TotalUsers        uint64
	ActiveUsers       uint64
	TotalReputation   uint64
	AverageReputation uint64
	MaxReputation     uint64
	MinReputation     uint64
}

// UpdateReputationConfig updates the reputation system configuration
func (rs *ReputationSystem) UpdateReputationConfig(newConfig *ReputationConfig) error {
	// Validate configuration
	if newConfig.BaseReputation == 0 {
		return NewDAOError(ErrInvalidProposal, "base reputation must be greater than zero", nil)
	}

	if newConfig.MaxReputation <= newConfig.MinReputation {
		return NewDAOError(ErrInvalidProposal, "max reputation must be greater than min reputation", nil)
	}

	if newConfig.InactivityDecayRate < 0 || newConfig.InactivityDecayRate > 1 {
		return NewDAOError(ErrInvalidProposal, "inactivity decay rate must be between 0 and 1", nil)
	}

	if newConfig.DecayPeriodDays <= 0 {
		return NewDAOError(ErrInvalidProposal, "decay period must be positive", nil)
	}

	rs.config = newConfig
	return nil
}

// GetReputationConfig returns the current reputation configuration
func (rs *ReputationSystem) GetReputationConfig() *ReputationConfig {
	return rs.config
}

// RecalculateAllReputation recalculates reputation for all users based on their activity
func (rs *ReputationSystem) RecalculateAllReputation() {
	// This is a comprehensive recalculation that can be run periodically
	// to ensure reputation scores are accurate

	for addressStr, holder := range rs.governanceState.TokenHolders {
		// Reset to base reputation
		holder.Reputation = rs.config.BaseReputation

		// Add token-based reputation
		tokenBonus := holder.Balance / 100 // 1 reputation per 100 tokens
		holder.Reputation += tokenBonus

		// Count proposals created
		proposalsCreated := 0
		proposalsPassed := 0
		proposalsRejected := 0

		for _, proposal := range rs.governanceState.Proposals {
			if proposal.Creator.String() == addressStr {
				proposalsCreated++
				if proposal.Status == ProposalStatusPassed {
					proposalsPassed++
				} else if proposal.Status == ProposalStatusRejected {
					proposalsRejected++
				}
			}
		}

		// Add proposal bonuses/penalties
		holder.Reputation += uint64(proposalsCreated) * rs.config.ProposalCreationBonus
		holder.Reputation += uint64(proposalsPassed) * rs.config.ProposalPassedBonus
		if holder.Reputation > uint64(proposalsRejected)*rs.config.ProposalRejectedPenalty {
			holder.Reputation -= uint64(proposalsRejected) * rs.config.ProposalRejectedPenalty
		}

		// Count votes cast
		votesCast := 0
		for _, votes := range rs.governanceState.Votes {
			if _, voted := votes[addressStr]; voted {
				votesCast++
			}
		}

		// Add voting participation bonus
		holder.Reputation += uint64(votesCast) * rs.config.VotingParticipation

		// Apply caps
		if holder.Reputation > rs.config.MaxReputation {
			holder.Reputation = rs.config.MaxReputation
		}
		if holder.Reputation < rs.config.MinReputation {
			holder.Reputation = rs.config.MinReputation
		}
	}

	// Apply inactivity decay
	rs.ApplyInactivityDecay()
}

// GetUserReputationHistory returns reputation-affecting events for a user
func (rs *ReputationSystem) GetUserReputationHistory(user crypto.PublicKey) *UserReputationHistory {
	userStr := user.String()
	holder, exists := rs.governanceState.TokenHolders[userStr]
	if !exists {
		return nil
	}

	history := &UserReputationHistory{
		User:              user,
		CurrentReputation: holder.Reputation,
		JoinedAt:          holder.JoinedAt,
		LastActive:        holder.LastActive,
		Events:            make([]*ReputationEvent, 0),
	}

	// Count proposals
	for _, proposal := range rs.governanceState.Proposals {
		if proposal.Creator.String() == userStr {
			eventType := ReputationEventProposalCreated
			impact := int64(rs.config.ProposalCreationBonus)

			if proposal.Status == ProposalStatusPassed {
				impact += int64(rs.config.ProposalPassedBonus)
			} else if proposal.Status == ProposalStatusRejected {
				impact -= int64(rs.config.ProposalRejectedPenalty)
			}

			history.Events = append(history.Events, &ReputationEvent{
				Type:       eventType,
				Timestamp:  proposal.StartTime, // Use proposal start time as event time
				Impact:     impact,
				ProposalID: &proposal.ID,
			})
		}
	}

	// Count votes
	for proposalID, votes := range rs.governanceState.Votes {
		if vote, voted := votes[userStr]; voted {
			history.Events = append(history.Events, &ReputationEvent{
				Type:       ReputationEventVoteCast,
				Timestamp:  vote.Timestamp,
				Impact:     int64(rs.config.VotingParticipation),
				ProposalID: &proposalID,
			})
		}
	}

	return history
}

// ReputationEventType represents different types of reputation events
type ReputationEventType byte

const (
	ReputationEventProposalCreated ReputationEventType = 0x01
	ReputationEventVoteCast        ReputationEventType = 0x02
	ReputationEventInactivityDecay ReputationEventType = 0x03
)

// ReputationEvent represents a single reputation-affecting event
type ReputationEvent struct {
	Type       ReputationEventType
	Timestamp  int64
	Impact     int64 // Positive for gains, negative for losses
	ProposalID *types.Hash
}

// UserReputationHistory contains the reputation history for a user
type UserReputationHistory struct {
	User              crypto.PublicKey
	CurrentReputation uint64
	JoinedAt          int64
	LastActive        int64
	Events            []*ReputationEvent
}
