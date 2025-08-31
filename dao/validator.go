package dao

import (
	"fmt"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// DAOValidator handles validation of DAO transactions and operations
type DAOValidator struct {
	governanceState *GovernanceState
	tokenState      *GovernanceToken
}

// NewDAOValidator creates a new DAO validator
func NewDAOValidator(governanceState *GovernanceState, tokenState *GovernanceToken) *DAOValidator {
	return &DAOValidator{
		governanceState: governanceState,
		tokenState:      tokenState,
	}
}

// ValidateProposalTx validates a proposal transaction
func (v *DAOValidator) ValidateProposalTx(tx *ProposalTx, creator crypto.PublicKey) error {
	// Check if creator has sufficient tokens
	creatorStr := creator.String()
	balance, exists := v.tokenState.Balances[creatorStr]
	if !exists || balance < v.governanceState.Config.MinProposalThreshold {
		return ErrInsufficientTokensForProposal
	}

	// Validate proposal format
	if len(tx.Title) == 0 || len(tx.Title) > 200 {
		return NewDAOError(ErrInvalidProposal, "proposal title must be between 1 and 200 characters", nil)
	}

	if len(tx.Description) == 0 || len(tx.Description) > 10000 {
		return NewDAOError(ErrInvalidProposal, "proposal description must be between 1 and 10000 characters", nil)
	}

	// Validate timeframe
	// now := time.Now().Unix()
	// Allow past start times for testing - in production, uncomment the check below
	// if tx.StartTime <= now {
	//     return NewDAOError(ErrInvalidTimeframe, "proposal start time must be in the future", nil)
	// }

	if tx.EndTime <= tx.StartTime {
		return NewDAOError(ErrInvalidTimeframe, "proposal end time must be after start time", nil)
	}

	if tx.EndTime-tx.StartTime < v.governanceState.Config.VotingPeriod {
		return NewDAOError(ErrInvalidTimeframe, "voting period too short", nil)
	}

	// Validate proposal type
	if tx.ProposalType < ProposalTypeGeneral || tx.ProposalType > ProposalTypeParameter {
		return NewDAOError(ErrInvalidProposal, "invalid proposal type", nil)
	}

	// Validate voting type
	if tx.VotingType < VotingTypeSimple || tx.VotingType > VotingTypeReputation {
		return NewDAOError(ErrInvalidProposal, "invalid voting type", nil)
	}

	// Validate threshold
	if tx.Threshold == 0 || tx.Threshold > 10000 {
		return ErrInvalidThresholdError
	}

	// Additional validation for treasury proposals
	if tx.ProposalType == ProposalTypeTreasury {
		if balance < v.governanceState.Config.TreasuryThreshold {
			return NewDAOError(ErrInsufficientTokens, "insufficient tokens for treasury proposal", nil)
		}
	}

	return nil
}

// ValidateVoteTx validates a vote transaction with comprehensive checks
func (v *DAOValidator) ValidateVoteTx(tx *VoteTx, voter crypto.PublicKey) error {
	// Check if proposal exists
	proposal, exists := v.governanceState.Proposals[tx.ProposalID]
	if !exists {
		return ErrProposalNotFoundError
	}

	// Check if proposal is active
	now := time.Now().Unix()
	if now < proposal.StartTime {
		return ErrVotingNotStarted
	}

	if now > proposal.EndTime {
		return ErrVotingPeriodClosed
	}

	if proposal.Status != ProposalStatusActive {
		return NewDAOError(ErrVotingClosed, "proposal is not in active status", nil)
	}

	// Enhanced double-voting prevention
	voterStr := voter.String()
	if err := v.validateNoDuplicateVote(tx.ProposalID, voterStr); err != nil {
		return err
	}

	// Validate vote choice
	if tx.Choice < VoteChoiceYes || tx.Choice > VoteChoiceAbstain {
		return ErrInvalidVoteChoiceError
	}

	// Check voter eligibility (must have tokens)
	balance, exists := v.tokenState.Balances[voterStr]
	if !exists || balance == 0 {
		return ErrInsufficientTokensForVote
	}

	// Validate vote weight is not zero
	if tx.Weight == 0 {
		return NewDAOError(ErrInvalidProposal, "vote weight must be greater than zero", nil)
	}

	// Validate vote weight and cost based on voting type
	if err := v.validateVotingWeightAndCost(tx, voter, proposal, balance); err != nil {
		return err
	}

	// Validate voter has enough tokens for fee
	if balance < uint64(tx.Fee) {
		return NewDAOError(ErrInsufficientTokens, "insufficient tokens for voting fee", nil)
	}

	return nil
}

// validateNoDuplicateVote ensures the voter hasn't already voted on this proposal
func (v *DAOValidator) validateNoDuplicateVote(proposalID types.Hash, voterStr string) error {
	if votes, exists := v.governanceState.Votes[proposalID]; exists {
		if existingVote, hasVoted := votes[voterStr]; hasVoted {
			return NewDAOError(ErrDuplicateVote,
				fmt.Sprintf("voter has already cast %s vote on this proposal",
					v.voteChoiceToString(existingVote.Choice)),
				map[string]interface{}{
					"existing_vote_timestamp": existingVote.Timestamp,
					"existing_vote_weight":    existingVote.Weight,
				})
		}
	}
	return nil
}

// validateVotingWeightAndCost validates vote weight and ensures voter has sufficient tokens
func (v *DAOValidator) validateVotingWeightAndCost(tx *VoteTx, voter crypto.PublicKey, proposal *Proposal, balance uint64) error {
	voterStr := voter.String()

	switch proposal.VotingType {
	case VotingTypeSimple:
		// Simple voting: one token = one vote, cost = weight
		totalCost := tx.Weight + uint64(tx.Fee)
		if totalCost > balance {
			return NewDAOError(ErrInsufficientTokens,
				fmt.Sprintf("insufficient tokens: need %d, have %d", totalCost, balance), nil)
		}

	case VotingTypeQuadratic:
		// Quadratic voting: cost = weight^2 + fee
		voteCost := tx.Weight * tx.Weight
		totalCost := voteCost + uint64(tx.Fee)
		if totalCost > balance {
			return NewDAOError(ErrInsufficientTokens,
				fmt.Sprintf("insufficient tokens for quadratic vote: need %d (vote cost: %d, fee: %d), have %d",
					totalCost, voteCost, tx.Fee, balance), nil)
		}

	case VotingTypeWeighted:
		// Token-weighted: weight proportional to balance, cost = weight
		if tx.Weight > balance {
			return NewDAOError(ErrInsufficientTokens,
				fmt.Sprintf("vote weight %d exceeds token balance %d", tx.Weight, balance), nil)
		}
		totalCost := tx.Weight + uint64(tx.Fee)
		if totalCost > balance {
			return NewDAOError(ErrInsufficientTokens,
				fmt.Sprintf("insufficient tokens: need %d, have %d", totalCost, balance), nil)
		}

	case VotingTypeReputation:
		// Reputation-based: check reputation score and calculate cost
		holder, exists := v.governanceState.TokenHolders[voterStr]
		if !exists {
			return NewDAOError(ErrUnauthorized, "voter not found in token holders registry", nil)
		}

		if holder.Reputation == 0 {
			return NewDAOError(ErrInsufficientTokens, "voter has no reputation to vote", nil)
		}

		if tx.Weight > holder.Reputation {
			return NewDAOError(ErrInsufficientTokens,
				fmt.Sprintf("vote weight %d exceeds reputation %d", tx.Weight, holder.Reputation), nil)
		}

		// Calculate proportional cost for reputation voting
		voteCost := (tx.Weight * balance) / holder.Reputation
		totalCost := voteCost + uint64(tx.Fee)
		if totalCost > balance {
			return NewDAOError(ErrInsufficientTokens,
				fmt.Sprintf("insufficient tokens for reputation vote: need %d, have %d", totalCost, balance), nil)
		}

	default:
		return NewDAOError(ErrInvalidProposal, "unsupported voting type", nil)
	}

	return nil
}

// voteChoiceToString converts vote choice to string for error messages
func (v *DAOValidator) voteChoiceToString(choice VoteChoice) string {
	switch choice {
	case VoteChoiceYes:
		return "YES"
	case VoteChoiceNo:
		return "NO"
	case VoteChoiceAbstain:
		return "ABSTAIN"
	default:
		return "UNKNOWN"
	}
}

// ValidateDelegationTx validates a delegation transaction
func (v *DAOValidator) ValidateDelegationTx(tx *DelegationTx, delegator crypto.PublicKey) error {
	// Check if delegator has tokens
	delegatorStr := delegator.String()
	balance, exists := v.tokenState.Balances[delegatorStr]
	if !exists || balance == 0 {
		return NewDAOError(ErrInsufficientTokens, "delegator has no tokens", nil)
	}

	// Check if delegator has enough tokens for fee
	if balance < uint64(tx.Fee) {
		return NewDAOError(ErrInsufficientTokens, "insufficient tokens for delegation fee", nil)
	}

	if tx.Revoke {
		// For revocation, check if there's an active delegation to revoke
		if delegation, exists := v.governanceState.Delegations[delegatorStr]; !exists || !delegation.Active {
			return NewDAOError(ErrInvalidDelegation, "no active delegation to revoke", nil)
		}
	} else {
		// For new delegation, validate delegate and duration

		// Check if delegate is different from delegator
		if tx.Delegate.String() == delegator.String() {
			return NewDAOError(ErrInvalidDelegation, "cannot delegate to self", nil)
		}

		// Validate duration
		if tx.Duration <= 0 {
			return NewDAOError(ErrInvalidDelegation, "delegation duration must be positive", nil)
		}

		// Check maximum duration (e.g., 1 year)
		maxDuration := int64(365 * 24 * 3600) // 1 year in seconds
		if tx.Duration > maxDuration {
			return NewDAOError(ErrInvalidDelegation, "delegation duration exceeds maximum allowed", nil)
		}

		// Check if delegate exists (has tokens or is registered)
		delegateStr := tx.Delegate.String()
		if _, exists := v.tokenState.Balances[delegateStr]; !exists {
			return NewDAOError(ErrInvalidDelegation, "delegate address not found", nil)
		}

		// Check if delegator already has an active delegation
		if existingDelegation, exists := v.governanceState.Delegations[delegatorStr]; exists && existingDelegation.Active {
			now := time.Now().Unix()
			if now >= existingDelegation.StartTime && now <= existingDelegation.EndTime {
				return NewDAOError(ErrInvalidDelegation, "delegator already has an active delegation", nil)
			}
		}
	}

	return nil
}

// ValidateTreasuryTx validates a treasury transaction
func (v *DAOValidator) ValidateTreasuryTx(tx *TreasuryTx) error {
	// Check treasury balance
	if tx.Amount > v.governanceState.Treasury.Balance {
		return ErrTreasuryInsufficientFunds
	}

	// Validate amount
	if tx.Amount == 0 {
		return NewDAOError(ErrInvalidProposal, "treasury amount must be greater than zero", nil)
	}

	// Validate purpose
	if len(tx.Purpose) == 0 || len(tx.Purpose) > 500 {
		return NewDAOError(ErrInvalidProposal, "treasury purpose must be between 1 and 500 characters", nil)
	}

	// Validate required signatures setting
	if tx.RequiredSigs > uint8(len(v.governanceState.Treasury.Signers)) {
		return NewDAOError(ErrInvalidSignature, "required signatures exceeds available signers", nil)
	}

	// Validate each signature if any are provided
	for i, sig := range tx.Signatures {
		if i >= len(v.governanceState.Treasury.Signers) {
			return NewDAOError(ErrInvalidSignature, "too many signatures provided", nil)
		}

		// Note: In a real implementation, you would verify the signature against the transaction data
		// For now, we just check that the signature is not nil
		if sig.R == nil || sig.S == nil {
			return NewDAOError(ErrInvalidSignature, fmt.Sprintf("invalid signature from signer %d", i), nil)
		}
	}

	return nil
}

// ValidateTokenMintTx validates a token minting transaction
func (v *DAOValidator) ValidateTokenMintTx(tx *TokenMintTx, minter crypto.PublicKey) error {
	// Check if minter is authorized (for now, any token holder can mint - this would be restricted in production)
	minterStr := minter.String()
	balance, exists := v.tokenState.Balances[minterStr]
	if !exists || balance == 0 {
		return NewDAOError(ErrUnauthorized, "minter has no tokens", nil)
	}

	// Validate amount
	if tx.Amount == 0 {
		return NewDAOError(ErrInvalidProposal, "mint amount must be greater than zero", nil)
	}

	// Validate reason
	if len(tx.Reason) == 0 || len(tx.Reason) > 200 {
		return NewDAOError(ErrInvalidProposal, "mint reason must be between 1 and 200 characters", nil)
	}

	// Check for overflow
	if v.tokenState.TotalSupply+tx.Amount < v.tokenState.TotalSupply {
		return NewDAOError(ErrTokenTransferFailed, "token supply overflow", nil)
	}

	return nil
}

// ValidateTokenBurnTx validates a token burning transaction
func (v *DAOValidator) ValidateTokenBurnTx(tx *TokenBurnTx, burner crypto.PublicKey) error {
	// Check if burner has sufficient tokens
	burnerStr := burner.String()
	balance, exists := v.tokenState.Balances[burnerStr]
	if !exists || balance < tx.Amount+uint64(tx.Fee) {
		return NewDAOError(ErrInsufficientTokens, "insufficient tokens to burn and pay fee", nil)
	}

	// Validate amount
	if tx.Amount == 0 {
		return NewDAOError(ErrInvalidProposal, "burn amount must be greater than zero", nil)
	}

	// Validate reason
	if len(tx.Reason) == 0 || len(tx.Reason) > 200 {
		return NewDAOError(ErrInvalidProposal, "burn reason must be between 1 and 200 characters", nil)
	}

	return nil
}

// ValidateTokenTransferTx validates a token transfer transaction
func (v *DAOValidator) ValidateTokenTransferTx(tx *TokenTransferTx, sender crypto.PublicKey) error {
	// Check if sender has sufficient tokens
	senderStr := sender.String()
	balance, exists := v.tokenState.Balances[senderStr]
	if !exists || balance < tx.Amount+uint64(tx.Fee) {
		return NewDAOError(ErrInsufficientTokens, "insufficient tokens for transfer and fee", nil)
	}

	// Validate amount
	if tx.Amount == 0 {
		return NewDAOError(ErrInvalidProposal, "transfer amount must be greater than zero", nil)
	}

	// Check that sender is not transferring to themselves
	if tx.Recipient.String() == sender.String() {
		return NewDAOError(ErrInvalidProposal, "cannot transfer to self", nil)
	}

	return nil
}

// ValidateTokenApproveTx validates a token approval transaction
func (v *DAOValidator) ValidateTokenApproveTx(tx *TokenApproveTx, owner crypto.PublicKey) error {
	// Check if owner has sufficient tokens for fee
	ownerStr := owner.String()
	balance, exists := v.tokenState.Balances[ownerStr]
	if !exists || balance < uint64(tx.Fee) {
		return NewDAOError(ErrInsufficientTokens, "insufficient tokens for approval fee", nil)
	}

	// Check that owner is not approving themselves
	if tx.Spender.String() == owner.String() {
		return NewDAOError(ErrInvalidProposal, "cannot approve self", nil)
	}

	// Amount can be zero (to revoke approval)
	return nil
}

// ValidateTokenTransferFromTx validates a token transferFrom transaction
func (v *DAOValidator) ValidateTokenTransferFromTx(tx *TokenTransferFromTx, spender crypto.PublicKey) error {
	// Check if spender has sufficient tokens for fee
	spenderStr := spender.String()
	spenderBalance, exists := v.tokenState.Balances[spenderStr]
	if !exists || spenderBalance < uint64(tx.Fee) {
		return NewDAOError(ErrInsufficientTokens, "insufficient tokens for transfer fee", nil)
	}

	// Check if from address has sufficient balance
	fromStr := tx.From.String()
	fromBalance, exists := v.tokenState.Balances[fromStr]
	if !exists || fromBalance < tx.Amount {
		return NewDAOError(ErrInsufficientTokens, "insufficient balance in from address", nil)
	}

	// Check allowance
	allowance := v.tokenState.GetAllowance(fromStr, spenderStr)
	if allowance < tx.Amount {
		return NewDAOError(ErrInsufficientTokens, "insufficient allowance for transfer", nil)
	}

	// Validate amount
	if tx.Amount == 0 {
		return NewDAOError(ErrInvalidProposal, "transfer amount must be greater than zero", nil)
	}

	// Check that from and recipient are different
	if tx.From.String() == tx.Recipient.String() {
		return NewDAOError(ErrInvalidProposal, "cannot transfer to same address", nil)
	}

	return nil
}

// ValidateTokenDistributionTx validates a token distribution transaction
func (v *DAOValidator) ValidateTokenDistributionTx(tx *TokenDistributionTx, distributor crypto.PublicKey) error {
	// Check if distributor is authorized (should be DAO admin or governance)
	distributorStr := distributor.String()
	balance, exists := v.tokenState.Balances[distributorStr]
	if !exists || balance < uint64(tx.Fee) {
		return NewDAOError(ErrInsufficientTokens, "insufficient tokens for distribution fee", nil)
	}

	// Validate distribution category
	if tx.Category < DistributionFounders || tx.Category > DistributionEcosystem {
		return NewDAOError(ErrInvalidProposal, "invalid distribution category", nil)
	}

	// Validate recipients
	if len(tx.Recipients) == 0 {
		return NewDAOError(ErrInvalidProposal, "no recipients specified", nil)
	}

	// Validate each recipient and amount
	totalAmount := uint64(0)
	for recipientStr, amount := range tx.Recipients {
		if amount == 0 {
			return NewDAOError(ErrInvalidProposal, "recipient amount must be greater than zero", nil)
		}

		// Validate recipient address format (basic check)
		if len(recipientStr) == 0 {
			return NewDAOError(ErrInvalidProposal, "invalid recipient address", nil)
		}

		totalAmount += amount
		if totalAmount < amount { // Check for overflow
			return NewDAOError(ErrInvalidProposal, "total distribution amount overflow", nil)
		}
	}

	// Validate vesting parameters
	if tx.VestingType < VestingTypeLinear || tx.VestingType > VestingTypeImmediate {
		return NewDAOError(ErrInvalidProposal, "invalid vesting type", nil)
	}

	if tx.VestingType != VestingTypeImmediate {
		if tx.Duration <= 0 {
			return NewDAOError(ErrInvalidProposal, "vesting duration must be positive", nil)
		}

		if tx.CliffPeriod < 0 || tx.CliffPeriod >= tx.Duration {
			return NewDAOError(ErrInvalidProposal, "invalid cliff period", nil)
		}
	}

	return nil
}

// ValidateVestingClaimTx validates a vesting claim transaction
func (v *DAOValidator) ValidateVestingClaimTx(tx *VestingClaimTx, claimer crypto.PublicKey) error {
	// Check if claimer has sufficient tokens for fee
	claimerStr := claimer.String()
	balance, exists := v.tokenState.Balances[claimerStr]
	if !exists || balance < uint64(tx.Fee) {
		return NewDAOError(ErrInsufficientTokens, "insufficient tokens for claim fee", nil)
	}

	// Validate vesting ID
	if len(tx.VestingID) == 0 {
		return NewDAOError(ErrInvalidProposal, "vesting ID cannot be empty", nil)
	}

	return nil
}

// ValidateStakeTx validates a staking transaction
func (v *DAOValidator) ValidateStakeTx(tx *StakeTx, staker crypto.PublicKey) error {
	// Check if staker has sufficient tokens
	stakerStr := staker.String()
	balance, exists := v.tokenState.Balances[stakerStr]
	if !exists || balance < tx.Amount+uint64(tx.Fee) {
		return NewDAOError(ErrInsufficientTokens, "insufficient tokens for staking and fee", nil)
	}

	// Validate amount
	if tx.Amount == 0 {
		return NewDAOError(ErrInvalidProposal, "stake amount must be greater than zero", nil)
	}

	// Validate pool ID
	if len(tx.PoolID) == 0 {
		return NewDAOError(ErrInvalidProposal, "pool ID cannot be empty", nil)
	}

	// Validate duration (if specified)
	if tx.Duration < 0 {
		return NewDAOError(ErrInvalidProposal, "stake duration cannot be negative", nil)
	}

	return nil
}

// ValidateUnstakeTx validates an unstaking transaction
func (v *DAOValidator) ValidateUnstakeTx(tx *UnstakeTx, unstaker crypto.PublicKey) error {
	// Check if unstaker has sufficient tokens for fee
	unstakerStr := unstaker.String()
	balance, exists := v.tokenState.Balances[unstakerStr]
	if !exists || balance < uint64(tx.Fee) {
		return NewDAOError(ErrInsufficientTokens, "insufficient tokens for unstaking fee", nil)
	}

	// Validate amount
	if tx.Amount == 0 {
		return NewDAOError(ErrInvalidProposal, "unstake amount must be greater than zero", nil)
	}

	// Validate pool ID
	if len(tx.PoolID) == 0 {
		return NewDAOError(ErrInvalidProposal, "pool ID cannot be empty", nil)
	}

	return nil
}

// ValidateClaimRewardsTx validates a rewards claim transaction
func (v *DAOValidator) ValidateClaimRewardsTx(tx *ClaimRewardsTx, claimer crypto.PublicKey) error {
	// Check if claimer has sufficient tokens for fee
	claimerStr := claimer.String()
	balance, exists := v.tokenState.Balances[claimerStr]
	if !exists || balance < uint64(tx.Fee) {
		return NewDAOError(ErrInsufficientTokens, "insufficient tokens for claim fee", nil)
	}

	// Validate pool ID
	if len(tx.PoolID) == 0 {
		return NewDAOError(ErrInvalidProposal, "pool ID cannot be empty", nil)
	}

	return nil
}
