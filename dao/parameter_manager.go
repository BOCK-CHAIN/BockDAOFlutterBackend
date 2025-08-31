package dao

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// ParameterManager handles governance parameter management
type ParameterManager struct {
	governanceState  *GovernanceState
	tokenState       *GovernanceToken
	parameterConfig  *ParameterConfig
	parameterHistory map[string][]*ParameterChange
}

// ParameterConfig defines configurable DAO parameters
type ParameterConfig struct {
	// Proposal parameters
	MinProposalThreshold uint64 `json:"min_proposal_threshold"`
	VotingPeriod         int64  `json:"voting_period"`
	QuorumThreshold      uint64 `json:"quorum_threshold"`
	PassingThreshold     uint64 `json:"passing_threshold"`
	TreasuryThreshold    uint64 `json:"treasury_threshold"`

	// Voting parameters
	MaxVotingPeriod     int64  `json:"max_voting_period"`
	MinVotingPeriod     int64  `json:"min_voting_period"`
	QuadraticVotingCost uint64 `json:"quadratic_voting_cost"`

	// Token parameters
	MaxTokenSupply      uint64 `json:"max_token_supply"`
	TokenMintingRate    uint64 `json:"token_minting_rate"`
	TokenBurningEnabled bool   `json:"token_burning_enabled"`

	// Treasury parameters
	MaxTreasuryWithdraw uint64 `json:"max_treasury_withdraw"`
	TreasurySignersMin  uint8  `json:"treasury_signers_min"`
	TreasurySignersMax  uint8  `json:"treasury_signers_max"`

	// Delegation parameters
	MaxDelegationPeriod int64 `json:"max_delegation_period"`
	MinDelegationPeriod int64 `json:"min_delegation_period"`
	DelegationEnabled   bool  `json:"delegation_enabled"`

	// Reputation parameters
	ReputationEnabled   bool   `json:"reputation_enabled"`
	ReputationDecayRate uint64 `json:"reputation_decay_rate"`
	ReputationBoostRate uint64 `json:"reputation_boost_rate"`

	// Security parameters
	EmergencyPauseEnabled bool  `json:"emergency_pause_enabled"`
	MultiSigRequired      bool  `json:"multi_sig_required"`
	AuditLogRetention     int64 `json:"audit_log_retention"`
}

// ParameterChange represents a parameter change event
type ParameterChange struct {
	Parameter  string           `json:"parameter"`
	OldValue   interface{}      `json:"old_value"`
	NewValue   interface{}      `json:"new_value"`
	ChangedBy  crypto.PublicKey `json:"changed_by"`
	ChangedAt  int64            `json:"changed_at"`
	ProposalID types.Hash       `json:"proposal_id"`
	Reason     string           `json:"reason"`
}

// ParameterProposalTx represents a parameter change proposal transaction
type ParameterProposalTx struct {
	Fee              int64                  `json:"fee"`
	ParameterChanges map[string]interface{} `json:"parameter_changes"`
	Justification    string                 `json:"justification"`
	EffectiveTime    int64                  `json:"effective_time"`
	ProposalType     ProposalType           `json:"proposal_type"`
	VotingType       VotingType             `json:"voting_type"`
	StartTime        int64                  `json:"start_time"`
	EndTime          int64                  `json:"end_time"`
	Threshold        uint64                 `json:"threshold"`
}

// NewParameterManager creates a new parameter manager
func NewParameterManager(governanceState *GovernanceState, tokenState *GovernanceToken) *ParameterManager {
	return &ParameterManager{
		governanceState:  governanceState,
		tokenState:       tokenState,
		parameterConfig:  NewDefaultParameterConfig(),
		parameterHistory: make(map[string][]*ParameterChange),
	}
}

// NewDefaultParameterConfig creates default parameter configuration
func NewDefaultParameterConfig() *ParameterConfig {
	return &ParameterConfig{
		// Proposal parameters
		MinProposalThreshold: 1000,
		VotingPeriod:         86400, // 24 hours
		QuorumThreshold:      2000,
		PassingThreshold:     5100, // 51%
		TreasuryThreshold:    5000,

		// Voting parameters
		MaxVotingPeriod:     604800, // 7 days
		MinVotingPeriod:     3600,   // 1 hour
		QuadraticVotingCost: 1,

		// Token parameters
		MaxTokenSupply:      1000000000, // 1 billion
		TokenMintingRate:    10000,      // 10k per mint
		TokenBurningEnabled: true,

		// Treasury parameters
		MaxTreasuryWithdraw: 100000,
		TreasurySignersMin:  2,
		TreasurySignersMax:  10,

		// Delegation parameters
		MaxDelegationPeriod: 2592000, // 30 days
		MinDelegationPeriod: 86400,   // 1 day
		DelegationEnabled:   true,

		// Reputation parameters
		ReputationEnabled:   true,
		ReputationDecayRate: 5,  // 5% per period
		ReputationBoostRate: 10, // 10% boost for participation

		// Security parameters
		EmergencyPauseEnabled: true,
		MultiSigRequired:      true,
		AuditLogRetention:     2592000, // 30 days
	}
}

// CreateParameterProposal creates a new parameter change proposal
func (pm *ParameterManager) CreateParameterProposal(creator crypto.PublicKey, parameterChanges map[string]interface{}, justification string, effectiveTime int64, votingType VotingType, startTime, endTime int64, threshold uint64) (types.Hash, error) {
	// Validate parameter changes
	if err := pm.ValidateParameterChanges(parameterChanges); err != nil {
		return types.Hash{}, fmt.Errorf("invalid parameter changes: %w", err)
	}

	// Validate timing
	if startTime >= endTime {
		return types.Hash{}, NewDAOError(ErrInvalidTimeframe, "start time must be before end time", nil)
	}

	if effectiveTime < endTime {
		return types.Hash{}, NewDAOError(ErrInvalidTimeframe, "effective time must be after voting ends", nil)
	}

	// Check creator has sufficient tokens
	creatorBalance := pm.tokenState.GetBalance(creator.String())
	if creatorBalance < pm.parameterConfig.MinProposalThreshold {
		return types.Hash{}, NewDAOError(ErrInsufficientTokens, "insufficient tokens to create parameter proposal", nil)
	}

	// Create proposal transaction
	paramTx := &ParameterProposalTx{
		Fee:              200,
		ParameterChanges: parameterChanges,
		Justification:    justification,
		EffectiveTime:    effectiveTime,
		ProposalType:     ProposalTypeParameter,
		VotingType:       votingType,
		StartTime:        startTime,
		EndTime:          endTime,
		Threshold:        threshold,
	}

	// Generate proposal ID
	proposalID := pm.generateParameterProposalID(paramTx, creator)

	// Create proposal
	proposal := &Proposal{
		ID:           proposalID,
		Creator:      creator,
		Title:        "Parameter Change Proposal",
		Description:  justification,
		ProposalType: ProposalTypeParameter,
		VotingType:   votingType,
		StartTime:    startTime,
		EndTime:      endTime,
		Status:       ProposalStatusPending,
		Threshold:    threshold,
		Results:      &VoteResults{},
		MetadataHash: types.Hash{}, // Could be extended to store detailed changes in IPFS
	}

	// Store proposal
	pm.governanceState.Proposals[proposalID] = proposal
	pm.governanceState.Votes[proposalID] = make(map[string]*Vote)

	return proposalID, nil
}

// ValidateParameterChanges validates proposed parameter changes
func (pm *ParameterManager) ValidateParameterChanges(changes map[string]interface{}) error {
	for param, value := range changes {
		if err := pm.validateSingleParameter(param, value); err != nil {
			return fmt.Errorf("invalid parameter %s: %w", param, err)
		}
	}
	return nil
}

// validateSingleParameter validates a single parameter change
func (pm *ParameterManager) validateSingleParameter(param string, value interface{}) error {
	switch param {
	case "min_proposal_threshold":
		if v, ok := value.(uint64); ok {
			if v == 0 {
				return fmt.Errorf("minimum proposal threshold must be greater than zero")
			}
			if v > pm.tokenState.TotalSupply/2 {
				return fmt.Errorf("minimum proposal threshold cannot exceed 50%% of total supply")
			}
		} else {
			return fmt.Errorf("min_proposal_threshold must be uint64")
		}

	case "voting_period":
		if v, ok := value.(int64); ok {
			if v < pm.parameterConfig.MinVotingPeriod || v > pm.parameterConfig.MaxVotingPeriod {
				return fmt.Errorf("voting period must be between %d and %d seconds", pm.parameterConfig.MinVotingPeriod, pm.parameterConfig.MaxVotingPeriod)
			}
		} else {
			return fmt.Errorf("voting_period must be int64")
		}

	case "quorum_threshold":
		if v, ok := value.(uint64); ok {
			if v == 0 {
				return fmt.Errorf("quorum threshold must be greater than zero")
			}
			if v > pm.tokenState.TotalSupply {
				return fmt.Errorf("quorum threshold cannot exceed total supply")
			}
		} else {
			return fmt.Errorf("quorum_threshold must be uint64")
		}

	case "passing_threshold":
		if v, ok := value.(uint64); ok {
			if v == 0 || v > 10000 {
				return fmt.Errorf("passing threshold must be between 1 and 10000 basis points")
			}
		} else {
			return fmt.Errorf("passing_threshold must be uint64")
		}

	case "treasury_threshold":
		if v, ok := value.(uint64); ok {
			if v > pm.tokenState.TotalSupply {
				return fmt.Errorf("treasury threshold cannot exceed total supply")
			}
		} else {
			return fmt.Errorf("treasury_threshold must be uint64")
		}

	case "max_treasury_withdraw":
		if v, ok := value.(uint64); ok {
			if v > pm.governanceState.Treasury.Balance {
				return fmt.Errorf("max treasury withdraw cannot exceed current treasury balance")
			}
		} else {
			return fmt.Errorf("max_treasury_withdraw must be uint64")
		}

	case "treasury_signers_min", "treasury_signers_max":
		if v, ok := value.(uint8); ok {
			if v == 0 {
				return fmt.Errorf("treasury signers count must be greater than zero")
			}
			if param == "treasury_signers_min" && v > pm.parameterConfig.TreasurySignersMax {
				return fmt.Errorf("minimum signers cannot exceed maximum signers")
			}
			if param == "treasury_signers_max" && v < pm.parameterConfig.TreasurySignersMin {
				return fmt.Errorf("maximum signers cannot be less than minimum signers")
			}
		} else {
			return fmt.Errorf("%s must be uint8", param)
		}

	case "max_token_supply":
		if v, ok := value.(uint64); ok {
			if v < pm.tokenState.TotalSupply {
				return fmt.Errorf("max token supply cannot be less than current total supply")
			}
		} else {
			return fmt.Errorf("max_token_supply must be uint64")
		}

	case "token_burning_enabled", "delegation_enabled", "reputation_enabled", "emergency_pause_enabled", "multi_sig_required":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%s must be bool", param)
		}

	case "max_delegation_period", "min_delegation_period", "audit_log_retention":
		if v, ok := value.(int64); ok {
			if v <= 0 {
				return fmt.Errorf("%s must be positive", param)
			}
		} else {
			return fmt.Errorf("%s must be int64", param)
		}

	case "reputation_decay_rate", "reputation_boost_rate", "quadratic_voting_cost", "token_minting_rate":
		if v, ok := value.(uint64); ok {
			if param == "reputation_decay_rate" || param == "reputation_boost_rate" {
				if v > 100 {
					return fmt.Errorf("%s cannot exceed 100%%", param)
				}
			}
		} else {
			return fmt.Errorf("%s must be uint64", param)
		}

	default:
		return fmt.Errorf("unknown parameter: %s", param)
	}

	return nil
}

// ExecuteParameterChanges executes approved parameter changes
func (pm *ParameterManager) ExecuteParameterChanges(proposalID types.Hash, executor crypto.PublicKey) error {
	proposal, exists := pm.governanceState.Proposals[proposalID]
	if !exists {
		return ErrProposalNotFoundError
	}

	if proposal.ProposalType != ProposalTypeParameter {
		return NewDAOError(ErrInvalidProposal, "proposal is not a parameter change proposal", nil)
	}

	if proposal.Status != ProposalStatusPassed {
		return NewDAOError(ErrInvalidProposal, "proposal has not passed", nil)
	}

	// Find the parameter changes from proposal metadata
	// In a real implementation, this would be stored in the proposal or IPFS
	// For now, we'll simulate retrieving the changes
	parameterChanges, err := pm.getParameterChangesFromProposal(proposalID)
	if err != nil {
		return fmt.Errorf("failed to retrieve parameter changes: %w", err)
	}

	// Apply parameter changes
	for param, newValue := range parameterChanges {
		oldValue := pm.getCurrentParameterValue(param)

		if err := pm.applyParameterChange(param, newValue); err != nil {
			return fmt.Errorf("failed to apply parameter change %s: %w", param, err)
		}

		// Record parameter change
		change := &ParameterChange{
			Parameter:  param,
			OldValue:   oldValue,
			NewValue:   newValue,
			ChangedBy:  executor,
			ChangedAt:  time.Now().Unix(),
			ProposalID: proposalID,
			Reason:     proposal.Description,
		}

		if pm.parameterHistory[param] == nil {
			pm.parameterHistory[param] = make([]*ParameterChange, 0)
		}
		pm.parameterHistory[param] = append(pm.parameterHistory[param], change)
	}

	// Update proposal status
	proposal.Status = ProposalStatusExecuted

	return nil
}

// applyParameterChange applies a single parameter change
func (pm *ParameterManager) applyParameterChange(param string, value interface{}) error {
	switch param {
	case "min_proposal_threshold":
		pm.parameterConfig.MinProposalThreshold = value.(uint64)
		pm.governanceState.Config.MinProposalThreshold = value.(uint64)
	case "voting_period":
		pm.parameterConfig.VotingPeriod = value.(int64)
		pm.governanceState.Config.VotingPeriod = value.(int64)
	case "quorum_threshold":
		pm.parameterConfig.QuorumThreshold = value.(uint64)
		pm.governanceState.Config.QuorumThreshold = value.(uint64)
	case "passing_threshold":
		pm.parameterConfig.PassingThreshold = value.(uint64)
		pm.governanceState.Config.PassingThreshold = value.(uint64)
	case "treasury_threshold":
		pm.parameterConfig.TreasuryThreshold = value.(uint64)
		pm.governanceState.Config.TreasuryThreshold = value.(uint64)
	case "max_voting_period":
		pm.parameterConfig.MaxVotingPeriod = value.(int64)
	case "min_voting_period":
		pm.parameterConfig.MinVotingPeriod = value.(int64)
	case "quadratic_voting_cost":
		pm.parameterConfig.QuadraticVotingCost = value.(uint64)
	case "max_token_supply":
		pm.parameterConfig.MaxTokenSupply = value.(uint64)
	case "token_minting_rate":
		pm.parameterConfig.TokenMintingRate = value.(uint64)
	case "token_burning_enabled":
		pm.parameterConfig.TokenBurningEnabled = value.(bool)
	case "max_treasury_withdraw":
		pm.parameterConfig.MaxTreasuryWithdraw = value.(uint64)
	case "treasury_signers_min":
		pm.parameterConfig.TreasurySignersMin = value.(uint8)
	case "treasury_signers_max":
		pm.parameterConfig.TreasurySignersMax = value.(uint8)
	case "max_delegation_period":
		pm.parameterConfig.MaxDelegationPeriod = value.(int64)
	case "min_delegation_period":
		pm.parameterConfig.MinDelegationPeriod = value.(int64)
	case "delegation_enabled":
		pm.parameterConfig.DelegationEnabled = value.(bool)
	case "reputation_enabled":
		pm.parameterConfig.ReputationEnabled = value.(bool)
	case "reputation_decay_rate":
		pm.parameterConfig.ReputationDecayRate = value.(uint64)
	case "reputation_boost_rate":
		pm.parameterConfig.ReputationBoostRate = value.(uint64)
	case "emergency_pause_enabled":
		pm.parameterConfig.EmergencyPauseEnabled = value.(bool)
	case "multi_sig_required":
		pm.parameterConfig.MultiSigRequired = value.(bool)
	case "audit_log_retention":
		pm.parameterConfig.AuditLogRetention = value.(int64)
	default:
		return fmt.Errorf("unknown parameter: %s", param)
	}

	return nil
}

// getCurrentParameterValue gets the current value of a parameter
func (pm *ParameterManager) getCurrentParameterValue(param string) interface{} {
	switch param {
	case "min_proposal_threshold":
		return pm.parameterConfig.MinProposalThreshold
	case "voting_period":
		return pm.parameterConfig.VotingPeriod
	case "quorum_threshold":
		return pm.parameterConfig.QuorumThreshold
	case "passing_threshold":
		return pm.parameterConfig.PassingThreshold
	case "treasury_threshold":
		return pm.parameterConfig.TreasuryThreshold
	case "max_voting_period":
		return pm.parameterConfig.MaxVotingPeriod
	case "min_voting_period":
		return pm.parameterConfig.MinVotingPeriod
	case "quadratic_voting_cost":
		return pm.parameterConfig.QuadraticVotingCost
	case "max_token_supply":
		return pm.parameterConfig.MaxTokenSupply
	case "token_minting_rate":
		return pm.parameterConfig.TokenMintingRate
	case "token_burning_enabled":
		return pm.parameterConfig.TokenBurningEnabled
	case "max_treasury_withdraw":
		return pm.parameterConfig.MaxTreasuryWithdraw
	case "treasury_signers_min":
		return pm.parameterConfig.TreasurySignersMin
	case "treasury_signers_max":
		return pm.parameterConfig.TreasurySignersMax
	case "max_delegation_period":
		return pm.parameterConfig.MaxDelegationPeriod
	case "min_delegation_period":
		return pm.parameterConfig.MinDelegationPeriod
	case "delegation_enabled":
		return pm.parameterConfig.DelegationEnabled
	case "reputation_enabled":
		return pm.parameterConfig.ReputationEnabled
	case "reputation_decay_rate":
		return pm.parameterConfig.ReputationDecayRate
	case "reputation_boost_rate":
		return pm.parameterConfig.ReputationBoostRate
	case "emergency_pause_enabled":
		return pm.parameterConfig.EmergencyPauseEnabled
	case "multi_sig_required":
		return pm.parameterConfig.MultiSigRequired
	case "audit_log_retention":
		return pm.parameterConfig.AuditLogRetention
	default:
		return nil
	}
}

// getParameterChangesFromProposal retrieves parameter changes from a proposal
// In a real implementation, this would be stored in IPFS or proposal metadata
func (pm *ParameterManager) getParameterChangesFromProposal(proposalID types.Hash) (map[string]interface{}, error) {
	// This is a placeholder implementation
	// In practice, you would store the parameter changes in the proposal metadata
	// or in IPFS and retrieve them here

	// For demonstration, return a sample parameter change
	return map[string]interface{}{
		"voting_period": int64(172800), // 48 hours
	}, nil
}

// GetParameterConfig returns the current parameter configuration
func (pm *ParameterManager) GetParameterConfig() *ParameterConfig {
	return pm.parameterConfig
}

// GetParameterHistory returns the change history for a parameter
func (pm *ParameterManager) GetParameterHistory(parameter string) []*ParameterChange {
	return pm.parameterHistory[parameter]
}

// GetAllParameterHistory returns the complete parameter change history
func (pm *ParameterManager) GetAllParameterHistory() map[string][]*ParameterChange {
	return pm.parameterHistory
}

// ValidateParameterProposal validates a parameter proposal before creation
func (pm *ParameterManager) ValidateParameterProposal(creator crypto.PublicKey, parameterChanges map[string]interface{}) error {
	// Check creator has sufficient tokens
	creatorBalance := pm.tokenState.GetBalance(creator.String())
	if creatorBalance < pm.parameterConfig.MinProposalThreshold {
		return NewDAOError(ErrInsufficientTokens, "insufficient tokens to create parameter proposal", nil)
	}

	// Validate parameter changes
	return pm.ValidateParameterChanges(parameterChanges)
}

// GetParameterValue returns the current value of a specific parameter
func (pm *ParameterManager) GetParameterValue(parameter string) (interface{}, error) {
	value := pm.getCurrentParameterValue(parameter)
	if value == nil {
		return nil, fmt.Errorf("unknown parameter: %s", parameter)
	}
	return value, nil
}

// ListAllParameters returns all configurable parameters and their current values
func (pm *ParameterManager) ListAllParameters() map[string]interface{} {
	params := make(map[string]interface{})

	// Use reflection or manual mapping to get all parameter values
	configBytes, _ := json.Marshal(pm.parameterConfig)
	json.Unmarshal(configBytes, &params)

	return params
}

// generateParameterProposalID generates a unique ID for a parameter proposal
func (pm *ParameterManager) generateParameterProposalID(tx *ParameterProposalTx, creator crypto.PublicKey) types.Hash {
	// Create a unique hash based on proposal content and creator
	data := fmt.Sprintf("param_%s_%d_%s", creator.String(), tx.StartTime, tx.Justification)
	hash := [32]byte{}
	copy(hash[:], []byte(data)[:32])
	return hash
}

// IsParameterChangeAllowed checks if a parameter change is allowed based on current state
func (pm *ParameterManager) IsParameterChangeAllowed(parameter string, newValue interface{}) (bool, string) {
	// Check if parameter changes are currently allowed
	if !pm.parameterConfig.EmergencyPauseEnabled {
		// If emergency pause is disabled, check if we're in emergency mode
		// This would require integration with the security manager
	}

	// Check specific parameter constraints
	switch parameter {
	case "min_proposal_threshold":
		if v, ok := newValue.(uint64); ok {
			if v > pm.tokenState.TotalSupply/2 {
				return false, "minimum proposal threshold cannot exceed 50% of total supply"
			}
		}
	case "max_treasury_withdraw":
		if v, ok := newValue.(uint64); ok {
			if v > pm.governanceState.Treasury.Balance {
				return false, "max treasury withdraw cannot exceed current treasury balance"
			}
		}
	}

	return true, ""
}

// GetParameterConstraints returns the constraints for a specific parameter
func (pm *ParameterManager) GetParameterConstraints(parameter string) map[string]interface{} {
	constraints := make(map[string]interface{})

	switch parameter {
	case "min_proposal_threshold":
		constraints["min"] = uint64(1)
		constraints["max"] = pm.tokenState.TotalSupply / 2
		constraints["type"] = "uint64"
	case "voting_period":
		constraints["min"] = pm.parameterConfig.MinVotingPeriod
		constraints["max"] = pm.parameterConfig.MaxVotingPeriod
		constraints["type"] = "int64"
	case "quorum_threshold":
		constraints["min"] = uint64(1)
		constraints["max"] = pm.tokenState.TotalSupply
		constraints["type"] = "uint64"
	case "passing_threshold":
		constraints["min"] = uint64(1)
		constraints["max"] = uint64(10000)
		constraints["type"] = "uint64"
		constraints["unit"] = "basis_points"
	case "treasury_threshold":
		constraints["min"] = uint64(0)
		constraints["max"] = pm.tokenState.TotalSupply
		constraints["type"] = "uint64"
	case "max_treasury_withdraw":
		constraints["min"] = uint64(0)
		constraints["max"] = pm.governanceState.Treasury.Balance
		constraints["type"] = "uint64"
	case "treasury_signers_min", "treasury_signers_max":
		constraints["min"] = uint8(1)
		constraints["max"] = uint8(255)
		constraints["type"] = "uint8"
	case "max_token_supply":
		constraints["min"] = pm.tokenState.TotalSupply
		constraints["max"] = ^uint64(0) // Max uint64
		constraints["type"] = "uint64"
	case "reputation_decay_rate", "reputation_boost_rate":
		constraints["min"] = uint64(0)
		constraints["max"] = uint64(100)
		constraints["type"] = "uint64"
		constraints["unit"] = "percentage"
	default:
		constraints["type"] = "unknown"
	}

	return constraints
}
