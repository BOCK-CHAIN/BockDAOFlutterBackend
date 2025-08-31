package dao

import (
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// DAOProcessor handles the processing of DAO transactions
type DAOProcessor struct {
	governanceState *GovernanceState
	tokenState      *GovernanceToken
	validator       *DAOValidator
}

// NewDAOProcessor creates a new DAO transaction processor
func NewDAOProcessor(governanceState *GovernanceState, tokenState *GovernanceToken) *DAOProcessor {
	validator := NewDAOValidator(governanceState, tokenState)
	return &DAOProcessor{
		governanceState: governanceState,
		tokenState:      tokenState,
		validator:       validator,
	}
}

// ProcessProposalTx processes a proposal transaction
func (p *DAOProcessor) ProcessProposalTx(tx *ProposalTx, creator crypto.PublicKey, txHash types.Hash) error {
	// Validate the transaction
	if err := p.validator.ValidateProposalTx(tx, creator); err != nil {
		return err
	}

	// Create the proposal
	proposal := &Proposal{
		ID:           txHash,
		Creator:      creator,
		Title:        tx.Title,
		Description:  tx.Description,
		ProposalType: tx.ProposalType,
		VotingType:   tx.VotingType,
		StartTime:    tx.StartTime,
		EndTime:      tx.EndTime,
		Status:       ProposalStatusPending,
		Threshold:    tx.Threshold,
		Results:      &VoteResults{},
		MetadataHash: tx.MetadataHash,
	}

	// Store the proposal
	p.governanceState.Proposals[txHash] = proposal

	// Initialize vote tracking for this proposal
	p.governanceState.Votes[txHash] = make(map[string]*Vote)

	// Deduct fee from creator's balance
	creatorStr := creator.String()
	p.tokenState.Balances[creatorStr] -= uint64(tx.Fee)

	// Update reputation for proposal creation
	p.updateReputationForProposalCreation(creator)

	return nil
}

// ProcessVoteTx processes a vote transaction with enhanced voting mechanisms
func (p *DAOProcessor) ProcessVoteTx(tx *VoteTx, voter crypto.PublicKey) error {
	// Validate the transaction
	if err := p.validator.ValidateVoteTx(tx, voter); err != nil {
		return err
	}

	// Get the proposal to determine voting type
	proposal := p.governanceState.Proposals[tx.ProposalID]
	if proposal == nil {
		return ErrProposalNotFoundError
	}

	// Calculate effective voting power and cost based on voting type
	effectiveWeight, cost, err := p.calculateVotingWeightAndCost(tx, voter, proposal)
	if err != nil {
		return err
	}

	// Create the vote with calculated effective weight
	vote := &Vote{
		Voter:     voter,
		Choice:    tx.Choice,
		Weight:    effectiveWeight,
		Timestamp: time.Now().Unix(),
		Reason:    tx.Reason,
	}

	// Store the vote
	voterStr := voter.String()
	if p.governanceState.Votes[tx.ProposalID] == nil {
		p.governanceState.Votes[tx.ProposalID] = make(map[string]*Vote)
	}
	p.governanceState.Votes[tx.ProposalID][voterStr] = vote

	// Update vote results with effective weight
	if proposal.Results == nil {
		proposal.Results = &VoteResults{}
	}

	switch tx.Choice {
	case VoteChoiceYes:
		proposal.Results.YesVotes += effectiveWeight
	case VoteChoiceNo:
		proposal.Results.NoVotes += effectiveWeight
	case VoteChoiceAbstain:
		proposal.Results.AbstainVotes += effectiveWeight
	}
	proposal.Results.TotalVoters++

	// Deduct voting cost from voter's balance
	p.tokenState.Balances[voterStr] -= cost

	// Deduct transaction fee
	p.tokenState.Balances[voterStr] -= uint64(tx.Fee)

	// Update reputation for voting participation
	p.updateReputationForVoting(voter, tx.ProposalID)

	return nil
}

// calculateVotingWeightAndCost calculates the effective voting weight and token cost based on voting type
func (p *DAOProcessor) calculateVotingWeightAndCost(tx *VoteTx, voter crypto.PublicKey, proposal *Proposal) (uint64, uint64, error) {
	voterStr := voter.String()
	voterBalance := p.tokenState.Balances[voterStr]

	switch proposal.VotingType {
	case VotingTypeSimple:
		// Simple majority: 1 token = 1 vote, cost = weight
		if tx.Weight > voterBalance {
			return 0, 0, NewDAOError(ErrInsufficientTokens, "vote weight exceeds token balance", nil)
		}
		return tx.Weight, tx.Weight, nil

	case VotingTypeQuadratic:
		// Quadratic voting: cost = weight^2, effective weight = weight
		cost := tx.Weight * tx.Weight
		if cost > voterBalance {
			return 0, 0, NewDAOError(ErrInsufficientTokens, "insufficient tokens for quadratic vote cost", nil)
		}
		return tx.Weight, cost, nil

	case VotingTypeWeighted:
		// Token-weighted: voting power proportional to token balance, cost = weight
		maxWeight := voterBalance
		if tx.Weight > maxWeight {
			return 0, 0, NewDAOError(ErrInsufficientTokens, "vote weight exceeds available balance", nil)
		}
		return tx.Weight, tx.Weight, nil

	case VotingTypeReputation:
		// Reputation-based: voting power based on reputation score
		// Use reputation system for calculation
		effectiveWeight, err := p.calculateReputationWeight(voter, tx.Weight)
		if err != nil {
			return 0, 0, err
		}

		cost, err := p.calculateReputationBasedVotingCost(voter, tx.Weight)
		if err != nil {
			return 0, 0, err
		}

		return effectiveWeight, cost, nil

	default:
		return 0, 0, NewDAOError(ErrInvalidProposal, "unsupported voting type", nil)
	}
}

// ProcessDelegationTx processes a delegation transaction
func (p *DAOProcessor) ProcessDelegationTx(tx *DelegationTx, delegator crypto.PublicKey) error {
	// Validate the transaction
	if err := p.validator.ValidateDelegationTx(tx, delegator); err != nil {
		return err
	}

	delegatorStr := delegator.String()

	if tx.Revoke {
		// Revoke existing delegation
		if existingDelegation, exists := p.governanceState.Delegations[delegatorStr]; exists {
			existingDelegation.Active = false
			existingDelegation.EndTime = time.Now().Unix()
		}
		// Note: We still store the revoked delegation for historical purposes
	} else {
		// Create or update delegation
		delegation := &Delegation{
			Delegator: delegator,
			Delegate:  tx.Delegate,
			StartTime: time.Now().Unix(),
			EndTime:   time.Now().Unix() + tx.Duration,
			Active:    true,
		}

		// Store the delegation
		p.governanceState.Delegations[delegatorStr] = delegation
	}

	// Deduct fee
	p.tokenState.Balances[delegatorStr] -= uint64(tx.Fee)

	return nil
}

// ProcessTreasuryTx processes a treasury transaction
func (p *DAOProcessor) ProcessTreasuryTx(tx *TreasuryTx, txHash types.Hash) error {
	// Create treasury manager
	treasuryManager := NewTreasuryManager(p.governanceState, p.tokenState)

	// Create the treasury transaction
	if err := treasuryManager.CreateTreasuryTransaction(tx, txHash); err != nil {
		return err
	}

	// If signatures are provided, add them and try to execute
	if len(tx.Signatures) > 0 {
		// Store signatures in the pending transaction
		pendingTx := p.governanceState.Treasury.Transactions[txHash]
		pendingTx.Signatures = tx.Signatures

		// Try to execute if we have enough signatures
		if len(tx.Signatures) >= int(tx.RequiredSigs) {
			return treasuryManager.ExecuteTreasuryTransaction(txHash)
		}
	}

	return nil
}

// ProcessTokenMintTx processes a token minting transaction
func (p *DAOProcessor) ProcessTokenMintTx(tx *TokenMintTx, minter crypto.PublicKey) error {
	// Validate the transaction
	if err := p.validator.ValidateTokenMintTx(tx, minter); err != nil {
		return err
	}

	// Mint tokens using the token state method
	recipientStr := tx.Recipient.String()
	if err := p.tokenState.Mint(recipientStr, tx.Amount); err != nil {
		return err
	}

	// Deduct fee from minter
	minterStr := minter.String()
	p.tokenState.Balances[minterStr] -= uint64(tx.Fee)

	// Update token holder record
	p.updateTokenHolderRecord(recipientStr)

	return nil
}

// ProcessTokenBurnTx processes a token burning transaction
func (p *DAOProcessor) ProcessTokenBurnTx(tx *TokenBurnTx, burner crypto.PublicKey) error {
	// Validate the transaction
	if err := p.validator.ValidateTokenBurnTx(tx, burner); err != nil {
		return err
	}

	// Burn tokens using the token state method
	burnerStr := burner.String()
	if err := p.tokenState.Burn(burnerStr, tx.Amount); err != nil {
		return err
	}

	// Deduct fee
	p.tokenState.Balances[burnerStr] -= uint64(tx.Fee)

	return nil
}

// ProcessTokenTransferTx processes a token transfer transaction
func (p *DAOProcessor) ProcessTokenTransferTx(tx *TokenTransferTx, sender crypto.PublicKey) error {
	// Validate the transaction
	if err := p.validator.ValidateTokenTransferTx(tx, sender); err != nil {
		return err
	}

	// Transfer tokens
	senderStr := sender.String()
	recipientStr := tx.Recipient.String()

	if err := p.tokenState.Transfer(senderStr, recipientStr, tx.Amount); err != nil {
		return err
	}

	// Deduct fee
	p.tokenState.Balances[senderStr] -= uint64(tx.Fee)

	// Update token holder records
	p.updateTokenHolderRecord(senderStr)
	p.updateTokenHolderRecord(recipientStr)

	return nil
}

// ProcessTokenApproveTx processes a token approval transaction
func (p *DAOProcessor) ProcessTokenApproveTx(tx *TokenApproveTx, owner crypto.PublicKey) error {
	// Validate the transaction
	if err := p.validator.ValidateTokenApproveTx(tx, owner); err != nil {
		return err
	}

	// Approve spender
	ownerStr := owner.String()
	spenderStr := tx.Spender.String()

	if err := p.tokenState.Approve(ownerStr, spenderStr, tx.Amount); err != nil {
		return err
	}

	// Deduct fee
	p.tokenState.Balances[ownerStr] -= uint64(tx.Fee)

	return nil
}

// ProcessTokenTransferFromTx processes a token transferFrom transaction
func (p *DAOProcessor) ProcessTokenTransferFromTx(tx *TokenTransferFromTx, spender crypto.PublicKey) error {
	// Validate the transaction
	if err := p.validator.ValidateTokenTransferFromTx(tx, spender); err != nil {
		return err
	}

	// Transfer tokens using allowance
	spenderStr := spender.String()
	fromStr := tx.From.String()
	recipientStr := tx.Recipient.String()

	if err := p.tokenState.TransferFrom(spenderStr, fromStr, recipientStr, tx.Amount); err != nil {
		return err
	}

	// Deduct fee from spender
	p.tokenState.Balances[spenderStr] -= uint64(tx.Fee)

	// Update token holder records
	p.updateTokenHolderRecord(fromStr)
	p.updateTokenHolderRecord(recipientStr)

	return nil
}

// ProcessParameterProposalTx processes a parameter change proposal transaction
func (p *DAOProcessor) ProcessParameterProposalTx(tx *ParameterProposalTx, creator crypto.PublicKey, txHash types.Hash) error {
	// Create parameter manager for validation and processing
	parameterManager := NewParameterManager(p.governanceState, p.tokenState)

	// Validate the parameter proposal
	if err := parameterManager.ValidateParameterProposal(creator, tx.ParameterChanges); err != nil {
		return err
	}

	// Validate timing
	if tx.StartTime >= tx.EndTime {
		return NewDAOError(ErrInvalidTimeframe, "start time must be before end time", nil)
	}

	if tx.EffectiveTime < tx.EndTime {
		return NewDAOError(ErrInvalidTimeframe, "effective time must be after voting ends", nil)
	}

	// Create the parameter proposal
	proposal := &Proposal{
		ID:           txHash,
		Creator:      creator,
		Title:        "Parameter Change Proposal",
		Description:  tx.Justification,
		ProposalType: ProposalTypeParameter,
		VotingType:   tx.VotingType,
		StartTime:    tx.StartTime,
		EndTime:      tx.EndTime,
		Status:       ProposalStatusPending,
		Threshold:    tx.Threshold,
		Results:      &VoteResults{},
		MetadataHash: types.Hash{}, // Could store parameter changes in IPFS
	}

	// Store the proposal
	p.governanceState.Proposals[txHash] = proposal

	// Initialize vote tracking for this proposal
	p.governanceState.Votes[txHash] = make(map[string]*Vote)

	// Deduct fee from creator's balance
	creatorStr := creator.String()
	p.tokenState.Balances[creatorStr] -= uint64(tx.Fee)

	// Update reputation for proposal creation
	p.updateReputationForProposalCreation(creator)

	return nil
}

// updateTokenHolderRecord updates or creates a token holder record
func (p *DAOProcessor) updateTokenHolderRecord(address string) {
	balance := p.tokenState.GetBalance(address)

	if holder, exists := p.governanceState.TokenHolders[address]; exists {
		holder.Balance = balance
		holder.LastActive = time.Now().Unix()
	} else if balance > 0 {
		// Create new token holder record
		pubKey := crypto.PublicKey(address) // Convert string back to PublicKey
		p.governanceState.TokenHolders[address] = &TokenHolder{
			Address:    pubKey,
			Balance:    balance,
			Staked:     0,
			Reputation: balance / 10, // Initial reputation based on balance
			JoinedAt:   time.Now().Unix(),
			LastActive: time.Now().Unix(),
		}
	}
}

// UpdateProposalStatus updates proposal status based on current time and voting results
func (p *DAOProcessor) UpdateProposalStatus(proposalID types.Hash) error {
	proposal, exists := p.governanceState.Proposals[proposalID]
	if !exists {
		return ErrProposalNotFoundError
	}

	now := time.Now().Unix()

	// Check if voting period has started
	if now >= proposal.StartTime && proposal.Status == ProposalStatusPending {
		proposal.Status = ProposalStatusActive
	}

	// Check if voting period has ended
	if now > proposal.EndTime && proposal.Status == ProposalStatusActive {
		// Calculate if proposal passed
		totalVotes := proposal.Results.YesVotes + proposal.Results.NoVotes + proposal.Results.AbstainVotes

		// Check quorum
		if totalVotes >= p.governanceState.Config.QuorumThreshold {
			proposal.Results.Quorum = totalVotes

			// Check if passed (excluding abstain votes from calculation)
			activeVotes := proposal.Results.YesVotes + proposal.Results.NoVotes
			if activeVotes > 0 {
				passPercentage := (proposal.Results.YesVotes * 10000) / activeVotes
				if passPercentage >= p.governanceState.Config.PassingThreshold {
					proposal.Status = ProposalStatusPassed
					proposal.Results.Passed = true
				} else {
					proposal.Status = ProposalStatusRejected
					proposal.Results.Passed = false
				}
			} else {
				// No active votes, proposal rejected
				proposal.Status = ProposalStatusRejected
				proposal.Results.Passed = false
			}
		} else {
			// Quorum not met
			proposal.Status = ProposalStatusRejected
			proposal.Results.Passed = false
		}

		// Update reputation based on proposal outcome
		p.updateReputationForProposalOutcome(proposalID)
	}

	return nil
}

// GetEffectiveVotingPower calculates the effective voting power for a user, including delegations
func (p *DAOProcessor) GetEffectiveVotingPower(user crypto.PublicKey) uint64 {
	userStr := user.String()
	now := time.Now().Unix()

	// Check if user has delegated their voting power
	if delegation, exists := p.governanceState.Delegations[userStr]; exists && delegation.Active {
		if now >= delegation.StartTime && now <= delegation.EndTime {
			// User has delegated their power, so they have no direct voting power
			return 0
		}
	}

	// Start with user's own balance
	power := p.tokenState.Balances[userStr]

	// Add delegated power from others
	for delegatorStr, delegation := range p.governanceState.Delegations {
		if delegation.Active && delegation.Delegate.String() == userStr {
			if now >= delegation.StartTime && now <= delegation.EndTime {
				power += p.tokenState.Balances[delegatorStr]
			}
		}
	}

	return power
}

// GetDelegatedPower returns the total voting power delegated to a user
func (p *DAOProcessor) GetDelegatedPower(delegate crypto.PublicKey) uint64 {
	delegateStr := delegate.String()
	now := time.Now().Unix()
	delegatedPower := uint64(0)

	for delegatorStr, delegation := range p.governanceState.Delegations {
		if delegation.Active && delegation.Delegate.String() == delegateStr {
			if now >= delegation.StartTime && now <= delegation.EndTime {
				delegatedPower += p.tokenState.Balances[delegatorStr]
			}
		}
	}

	return delegatedPower
}

// GetOwnVotingPower returns the user's own voting power (excluding delegations)
func (p *DAOProcessor) GetOwnVotingPower(user crypto.PublicKey) uint64 {
	userStr := user.String()
	now := time.Now().Unix()

	// Check if user has delegated their voting power
	if delegation, exists := p.governanceState.Delegations[userStr]; exists && delegation.Active {
		if now >= delegation.StartTime && now <= delegation.EndTime {
			// User has delegated their power
			return 0
		}
	}

	return p.tokenState.Balances[userStr]
}

// RevokeDelegation revokes an active delegation
func (p *DAOProcessor) RevokeDelegation(delegator crypto.PublicKey) error {
	delegatorStr := delegator.String()

	delegation, exists := p.governanceState.Delegations[delegatorStr]
	if !exists || !delegation.Active {
		return NewDAOError(ErrInvalidDelegation, "no active delegation to revoke", nil)
	}

	delegation.Active = false
	delegation.EndTime = time.Now().Unix()

	return nil
}

// Reputation-related helper methods

// updateReputationForProposalCreation updates reputation when a user creates a proposal
func (p *DAOProcessor) updateReputationForProposalCreation(creator crypto.PublicKey) {
	creatorStr := creator.String()

	if holder, exists := p.governanceState.TokenHolders[creatorStr]; exists {
		// Create a temporary reputation system to access the config
		reputationSystem := NewReputationSystem(p.governanceState, p.tokenState)
		config := reputationSystem.GetReputationConfig()

		// Add proposal creation bonus
		newReputation := holder.Reputation + config.ProposalCreationBonus
		if newReputation > config.MaxReputation {
			newReputation = config.MaxReputation
		}

		holder.Reputation = newReputation
		holder.LastActive = time.Now().Unix()
	}
}

// updateReputationForVoting updates reputation when a user votes
func (p *DAOProcessor) updateReputationForVoting(voter crypto.PublicKey, proposalID types.Hash) {
	voterStr := voter.String()

	if holder, exists := p.governanceState.TokenHolders[voterStr]; exists {
		// Create a temporary reputation system to access the config
		reputationSystem := NewReputationSystem(p.governanceState, p.tokenState)
		config := reputationSystem.GetReputationConfig()

		// Add voting participation bonus
		newReputation := holder.Reputation + config.VotingParticipation
		if newReputation > config.MaxReputation {
			newReputation = config.MaxReputation
		}

		holder.Reputation = newReputation
		holder.LastActive = time.Now().Unix()
	}
}

// calculateReputationWeight calculates voting weight based on reputation
func (p *DAOProcessor) calculateReputationWeight(voter crypto.PublicKey, requestedWeight uint64) (uint64, error) {
	voterStr := voter.String()
	holder, exists := p.governanceState.TokenHolders[voterStr]
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

// calculateReputationBasedVotingCost calculates the token cost for reputation-based voting
func (p *DAOProcessor) calculateReputationBasedVotingCost(voter crypto.PublicKey, weight uint64) (uint64, error) {
	voterStr := voter.String()
	holder, exists := p.governanceState.TokenHolders[voterStr]
	if !exists {
		return 0, NewDAOError(ErrUnauthorized, "voter not found in token holders", nil)
	}

	voterBalance := p.tokenState.GetBalance(voterStr)

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

// updateReputationForProposalOutcome updates reputation based on proposal outcomes
func (p *DAOProcessor) updateReputationForProposalOutcome(proposalID types.Hash) {
	proposal, exists := p.governanceState.Proposals[proposalID]
	if !exists {
		return
	}

	creatorStr := proposal.Creator.String()
	holder, exists := p.governanceState.TokenHolders[creatorStr]
	if !exists {
		return
	}

	// Create a temporary reputation system to access the config
	reputationSystem := NewReputationSystem(p.governanceState, p.tokenState)
	config := reputationSystem.GetReputationConfig()

	switch proposal.Status {
	case ProposalStatusPassed:
		// Bonus for successful proposal
		newReputation := holder.Reputation + config.ProposalPassedBonus
		if newReputation > config.MaxReputation {
			newReputation = config.MaxReputation
		}
		holder.Reputation = newReputation

	case ProposalStatusRejected:
		// Penalty for rejected proposal (but not below minimum)
		if holder.Reputation > config.ProposalRejectedPenalty {
			newReputation := holder.Reputation - config.ProposalRejectedPenalty
			if newReputation < config.MinReputation {
				newReputation = config.MinReputation
			}
			holder.Reputation = newReputation
		}
	}
}

// ProcessTokenDistributionTx processes a token distribution transaction
func (p *DAOProcessor) ProcessTokenDistributionTx(tx *TokenDistributionTx, distributor crypto.PublicKey) error {
	// Validate the transaction
	if err := p.validator.ValidateTokenDistributionTx(tx, distributor); err != nil {
		return err
	}

	// Create tokenomics manager
	tokenomicsManager := NewTokenomicsManager(p.governanceState, p.tokenState)

	// Process each recipient
	for recipientStr, amount := range tx.Recipients {
		// Convert string to PublicKey
		recipient := crypto.PublicKey(recipientStr)

		// Add distribution recipient
		if err := tokenomicsManager.AddDistributionRecipient(tx.Category, recipient, amount); err != nil {
			return err
		}
	}

	// Deduct fee from distributor
	distributorStr := distributor.String()
	p.tokenState.Balances[distributorStr] -= uint64(tx.Fee)

	return nil
}

// ProcessVestingClaimTx processes a vesting claim transaction
func (p *DAOProcessor) ProcessVestingClaimTx(tx *VestingClaimTx, claimer crypto.PublicKey) error {
	// Validate the transaction
	if err := p.validator.ValidateVestingClaimTx(tx, claimer); err != nil {
		return err
	}

	// Create tokenomics manager
	tokenomicsManager := NewTokenomicsManager(p.governanceState, p.tokenState)

	// Claim vested tokens
	claimedAmount, err := tokenomicsManager.ClaimVestedTokens(tx.VestingID, claimer)
	if err != nil {
		return err
	}

	// Deduct fee from claimer
	claimerStr := claimer.String()
	p.tokenState.Balances[claimerStr] -= uint64(tx.Fee)

	// Update token holder record
	p.updateTokenHolderRecord(claimerStr)

	// Log claimed amount (could be used for events)
	_ = claimedAmount

	return nil
}

// ProcessStakeTx processes a staking transaction
func (p *DAOProcessor) ProcessStakeTx(tx *StakeTx, staker crypto.PublicKey) error {
	// Validate the transaction
	if err := p.validator.ValidateStakeTx(tx, staker); err != nil {
		return err
	}

	// Create tokenomics manager
	tokenomicsManager := NewTokenomicsManager(p.governanceState, p.tokenState)

	// Stake tokens
	if err := tokenomicsManager.StakeTokens(tx.PoolID, staker, tx.Amount, tx.Duration); err != nil {
		return err
	}

	// Deduct fee from staker
	stakerStr := staker.String()
	p.tokenState.Balances[stakerStr] -= uint64(tx.Fee)

	return nil
}

// ProcessUnstakeTx processes an unstaking transaction
func (p *DAOProcessor) ProcessUnstakeTx(tx *UnstakeTx, unstaker crypto.PublicKey) error {
	// Validate the transaction
	if err := p.validator.ValidateUnstakeTx(tx, unstaker); err != nil {
		return err
	}

	// Create tokenomics manager
	tokenomicsManager := NewTokenomicsManager(p.governanceState, p.tokenState)

	// Unstake tokens
	if err := tokenomicsManager.UnstakeTokens(tx.PoolID, unstaker, tx.Amount); err != nil {
		return err
	}

	// Deduct fee from unstaker
	unstakerStr := unstaker.String()
	p.tokenState.Balances[unstakerStr] -= uint64(tx.Fee)

	return nil
}

// ProcessClaimRewardsTx processes a rewards claim transaction
func (p *DAOProcessor) ProcessClaimRewardsTx(tx *ClaimRewardsTx, claimer crypto.PublicKey) error {
	// Validate the transaction
	if err := p.validator.ValidateClaimRewardsTx(tx, claimer); err != nil {
		return err
	}

	// Create tokenomics manager
	tokenomicsManager := NewTokenomicsManager(p.governanceState, p.tokenState)

	// Claim staking rewards
	rewardAmount, err := tokenomicsManager.ClaimStakingRewards(tx.PoolID, claimer)
	if err != nil {
		return err
	}

	// Deduct fee from claimer
	claimerStr := claimer.String()
	p.tokenState.Balances[claimerStr] -= uint64(tx.Fee)

	// Update token holder record
	p.updateTokenHolderRecord(claimerStr)

	// Log reward amount (could be used for events)
	_ = rewardAmount

	return nil
}
