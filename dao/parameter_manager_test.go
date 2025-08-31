package dao

import (
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParameterManager(t *testing.T) {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)

	pm := NewParameterManager(governanceState, tokenState)

	assert.NotNil(t, pm)
	assert.NotNil(t, pm.parameterConfig)
	assert.NotNil(t, pm.parameterHistory)
	assert.Equal(t, uint64(1000), pm.parameterConfig.MinProposalThreshold)
}

func TestNewDefaultParameterConfig(t *testing.T) {
	config := NewDefaultParameterConfig()

	assert.NotNil(t, config)
	assert.Equal(t, uint64(1000), config.MinProposalThreshold)
	assert.Equal(t, int64(86400), config.VotingPeriod)
	assert.Equal(t, uint64(2000), config.QuorumThreshold)
	assert.Equal(t, uint64(5100), config.PassingThreshold)
	assert.Equal(t, uint64(5000), config.TreasuryThreshold)
	assert.True(t, config.TokenBurningEnabled)
	assert.True(t, config.DelegationEnabled)
	assert.True(t, config.ReputationEnabled)
}

func TestCreateParameterProposal(t *testing.T) {
	// Setup
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	pm := NewParameterManager(governanceState, tokenState)

	creator := crypto.GeneratePrivateKey().PublicKey()
	creatorStr := creator.String()

	// Give creator sufficient tokens
	tokenState.Mint(creatorStr, 2000)

	// Test data
	parameterChanges := map[string]interface{}{
		"voting_period":    int64(172800), // 48 hours
		"quorum_threshold": uint64(3000),
	}
	justification := "Increase voting period for better participation"
	effectiveTime := time.Now().Unix() + 7200 // 2 hours from now
	startTime := time.Now().Unix() + 3600     // 1 hour from now
	endTime := time.Now().Unix() + 7200       // 2 hours from now
	threshold := uint64(1000)

	// Create parameter proposal
	proposalID, err := pm.CreateParameterProposal(
		creator,
		parameterChanges,
		justification,
		effectiveTime,
		VotingTypeSimple,
		startTime,
		endTime,
		threshold,
	)

	require.NoError(t, err)
	assert.NotEqual(t, types.Hash{}, proposalID)

	// Verify proposal was created
	proposal, exists := governanceState.Proposals[proposalID]
	require.True(t, exists)
	assert.Equal(t, creator, proposal.Creator)
	assert.Equal(t, ProposalTypeParameter, proposal.ProposalType)
	assert.Equal(t, justification, proposal.Description)
	assert.Equal(t, ProposalStatusPending, proposal.Status)
}

func TestCreateParameterProposalInsufficientTokens(t *testing.T) {
	// Setup
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	pm := NewParameterManager(governanceState, tokenState)

	creator := crypto.GeneratePrivateKey().PublicKey()
	creatorStr := creator.String()

	// Give creator insufficient tokens
	tokenState.Mint(creatorStr, 500) // Less than minimum threshold

	parameterChanges := map[string]interface{}{
		"voting_period": int64(172800),
	}

	// Attempt to create parameter proposal
	_, err := pm.CreateParameterProposal(
		creator,
		parameterChanges,
		"Test proposal",
		time.Now().Unix()+7200,
		VotingTypeSimple,
		time.Now().Unix()+3600,
		time.Now().Unix()+7200,
		1000,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient tokens")
}

func TestValidateParameterChanges(t *testing.T) {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	tokenState.TotalSupply = 1000000 // Set total supply for validation
	pm := NewParameterManager(governanceState, tokenState)

	tests := []struct {
		name    string
		changes map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid voting period change",
			changes: map[string]interface{}{
				"voting_period": int64(172800),
			},
			wantErr: false,
		},
		{
			name: "valid quorum threshold change",
			changes: map[string]interface{}{
				"quorum_threshold": uint64(3000),
			},
			wantErr: false,
		},
		{
			name: "invalid voting period - too short",
			changes: map[string]interface{}{
				"voting_period": int64(1800), // 30 minutes, less than minimum
			},
			wantErr: true,
			errMsg:  "voting period must be between",
		},
		{
			name: "invalid min proposal threshold - zero",
			changes: map[string]interface{}{
				"min_proposal_threshold": uint64(0),
			},
			wantErr: true,
			errMsg:  "minimum proposal threshold must be greater than zero",
		},
		{
			name: "invalid min proposal threshold - too high",
			changes: map[string]interface{}{
				"min_proposal_threshold": uint64(600000), // More than 50% of supply
			},
			wantErr: true,
			errMsg:  "minimum proposal threshold cannot exceed 50%",
		},
		{
			name: "invalid passing threshold - too high",
			changes: map[string]interface{}{
				"passing_threshold": uint64(15000), // More than 100%
			},
			wantErr: true,
			errMsg:  "passing threshold must be between 1 and 10000 basis points",
		},
		{
			name: "invalid parameter type",
			changes: map[string]interface{}{
				"voting_period": "invalid_type",
			},
			wantErr: true,
			errMsg:  "voting_period must be int64",
		},
		{
			name: "unknown parameter",
			changes: map[string]interface{}{
				"unknown_param": uint64(100),
			},
			wantErr: true,
			errMsg:  "unknown parameter",
		},
		{
			name: "valid boolean parameter",
			changes: map[string]interface{}{
				"delegation_enabled": false,
			},
			wantErr: false,
		},
		{
			name: "invalid boolean parameter type",
			changes: map[string]interface{}{
				"delegation_enabled": "not_boolean",
			},
			wantErr: true,
			errMsg:  "delegation_enabled must be bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.ValidateParameterChanges(tt.changes)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecuteParameterChanges(t *testing.T) {
	// Setup
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	pm := NewParameterManager(governanceState, tokenState)

	creator := crypto.GeneratePrivateKey().PublicKey()
	executor := crypto.GeneratePrivateKey().PublicKey()

	// Create a passed parameter proposal
	proposalID := types.Hash{1, 2, 3}
	proposal := &Proposal{
		ID:           proposalID,
		Creator:      creator,
		Title:        "Parameter Change Proposal",
		Description:  "Test parameter change",
		ProposalType: ProposalTypeParameter,
		VotingType:   VotingTypeSimple,
		StartTime:    time.Now().Unix() - 7200,
		EndTime:      time.Now().Unix() - 3600,
		Status:       ProposalStatusPassed,
		Threshold:    1000,
		Results:      &VoteResults{Passed: true},
	}

	governanceState.Proposals[proposalID] = proposal

	// Store original values
	originalVotingPeriod := pm.parameterConfig.VotingPeriod
	_ = pm.parameterConfig.QuorumThreshold // Store for potential verification

	// Execute parameter changes
	err := pm.ExecuteParameterChanges(proposalID, executor)
	require.NoError(t, err)

	// Verify proposal status updated
	assert.Equal(t, ProposalStatusExecuted, proposal.Status)

	// Verify parameter history was recorded
	history := pm.GetParameterHistory("voting_period")
	assert.Len(t, history, 1)
	assert.Equal(t, originalVotingPeriod, history[0].OldValue)
	assert.Equal(t, executor, history[0].ChangedBy)
	assert.Equal(t, proposalID, history[0].ProposalID)
}

func TestExecuteParameterChangesInvalidProposal(t *testing.T) {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	pm := NewParameterManager(governanceState, tokenState)

	executor := crypto.GeneratePrivateKey().PublicKey()
	nonExistentID := types.Hash{9, 9, 9}

	// Try to execute non-existent proposal
	err := pm.ExecuteParameterChanges(nonExistentID, executor)
	assert.Error(t, err)
	assert.Equal(t, ErrProposalNotFoundError, err)
}

func TestExecuteParameterChangesNotPassed(t *testing.T) {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	pm := NewParameterManager(governanceState, tokenState)

	creator := crypto.GeneratePrivateKey().PublicKey()
	executor := crypto.GeneratePrivateKey().PublicKey()

	// Create a rejected parameter proposal
	proposalID := types.Hash{1, 2, 3}
	proposal := &Proposal{
		ID:           proposalID,
		Creator:      creator,
		ProposalType: ProposalTypeParameter,
		Status:       ProposalStatusRejected,
	}

	governanceState.Proposals[proposalID] = proposal

	// Try to execute rejected proposal
	err := pm.ExecuteParameterChanges(proposalID, executor)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "proposal has not passed")
}

func TestGetParameterValue(t *testing.T) {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	pm := NewParameterManager(governanceState, tokenState)

	// Test getting existing parameter
	value, err := pm.GetParameterValue("voting_period")
	require.NoError(t, err)
	assert.Equal(t, int64(86400), value)

	// Test getting non-existent parameter
	_, err = pm.GetParameterValue("non_existent_param")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown parameter")
}

func TestListAllParameters(t *testing.T) {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	pm := NewParameterManager(governanceState, tokenState)

	params := pm.ListAllParameters()

	assert.NotEmpty(t, params)
	assert.Contains(t, params, "min_proposal_threshold")
	assert.Contains(t, params, "voting_period")
	assert.Contains(t, params, "quorum_threshold")
	assert.Contains(t, params, "passing_threshold")
	assert.Contains(t, params, "delegation_enabled")

	// Verify values
	assert.Equal(t, float64(1000), params["min_proposal_threshold"])
	assert.Equal(t, float64(86400), params["voting_period"])
	assert.Equal(t, true, params["delegation_enabled"])
}

func TestIsParameterChangeAllowed(t *testing.T) {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	tokenState.TotalSupply = 1000000
	governanceState.Treasury.Balance = 50000
	pm := NewParameterManager(governanceState, tokenState)

	tests := []struct {
		name      string
		parameter string
		newValue  interface{}
		allowed   bool
		reason    string
	}{
		{
			name:      "valid voting period change",
			parameter: "voting_period",
			newValue:  int64(172800),
			allowed:   true,
			reason:    "",
		},
		{
			name:      "invalid min proposal threshold - too high",
			parameter: "min_proposal_threshold",
			newValue:  uint64(600000), // More than 50% of supply
			allowed:   false,
			reason:    "minimum proposal threshold cannot exceed 50% of total supply",
		},
		{
			name:      "invalid max treasury withdraw - exceeds balance",
			parameter: "max_treasury_withdraw",
			newValue:  uint64(100000), // More than treasury balance
			allowed:   false,
			reason:    "max treasury withdraw cannot exceed current treasury balance",
		},
		{
			name:      "valid max treasury withdraw",
			parameter: "max_treasury_withdraw",
			newValue:  uint64(30000), // Less than treasury balance
			allowed:   true,
			reason:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, reason := pm.IsParameterChangeAllowed(tt.parameter, tt.newValue)

			assert.Equal(t, tt.allowed, allowed)
			if !tt.allowed {
				assert.Equal(t, tt.reason, reason)
			}
		})
	}
}

func TestGetParameterConstraints(t *testing.T) {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	tokenState.TotalSupply = 1000000
	governanceState.Treasury.Balance = 50000
	pm := NewParameterManager(governanceState, tokenState)

	tests := []struct {
		name      string
		parameter string
		wantType  string
		hasMin    bool
		hasMax    bool
	}{
		{
			name:      "voting period constraints",
			parameter: "voting_period",
			wantType:  "int64",
			hasMin:    true,
			hasMax:    true,
		},
		{
			name:      "min proposal threshold constraints",
			parameter: "min_proposal_threshold",
			wantType:  "uint64",
			hasMin:    true,
			hasMax:    true,
		},
		{
			name:      "passing threshold constraints",
			parameter: "passing_threshold",
			wantType:  "uint64",
			hasMin:    true,
			hasMax:    true,
		},
		{
			name:      "max treasury withdraw constraints",
			parameter: "max_treasury_withdraw",
			wantType:  "uint64",
			hasMin:    true,
			hasMax:    true,
		},
		{
			name:      "unknown parameter",
			parameter: "unknown_param",
			wantType:  "unknown",
			hasMin:    false,
			hasMax:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraints := pm.GetParameterConstraints(tt.parameter)

			assert.Equal(t, tt.wantType, constraints["type"])

			if tt.hasMin {
				assert.Contains(t, constraints, "min")
			}

			if tt.hasMax {
				assert.Contains(t, constraints, "max")
			}
		})
	}
}

func TestParameterChangeHistory(t *testing.T) {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	pm := NewParameterManager(governanceState, tokenState)

	executor := crypto.GeneratePrivateKey().PublicKey()
	proposalID := types.Hash{1, 2, 3}

	// Simulate a parameter change
	change := &ParameterChange{
		Parameter:  "voting_period",
		OldValue:   int64(86400),
		NewValue:   int64(172800),
		ChangedBy:  executor,
		ChangedAt:  time.Now().Unix(),
		ProposalID: proposalID,
		Reason:     "Increase voting period for better participation",
	}

	pm.parameterHistory["voting_period"] = []*ParameterChange{change}

	// Test getting parameter history
	history := pm.GetParameterHistory("voting_period")
	assert.Len(t, history, 1)
	assert.Equal(t, change, history[0])

	// Test getting all parameter history
	allHistory := pm.GetAllParameterHistory()
	assert.Contains(t, allHistory, "voting_period")
	assert.Len(t, allHistory["voting_period"], 1)

	// Test getting history for non-existent parameter
	emptyHistory := pm.GetParameterHistory("non_existent")
	assert.Nil(t, emptyHistory)
}

func TestApplyParameterChange(t *testing.T) {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	pm := NewParameterManager(governanceState, tokenState)

	// Test applying various parameter changes
	tests := []struct {
		name      string
		parameter string
		value     interface{}
		wantErr   bool
	}{
		{
			name:      "apply voting period change",
			parameter: "voting_period",
			value:     int64(172800),
			wantErr:   false,
		},
		{
			name:      "apply quorum threshold change",
			parameter: "quorum_threshold",
			value:     uint64(3000),
			wantErr:   false,
		},
		{
			name:      "apply boolean parameter change",
			parameter: "delegation_enabled",
			value:     false,
			wantErr:   false,
		},
		{
			name:      "apply unknown parameter",
			parameter: "unknown_param",
			value:     uint64(100),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.applyParameterChange(tt.parameter, tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the parameter was actually changed
				currentValue := pm.getCurrentParameterValue(tt.parameter)
				assert.Equal(t, tt.value, currentValue)
			}
		})
	}
}

func TestValidateParameterProposal(t *testing.T) {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	pm := NewParameterManager(governanceState, tokenState)

	creator := crypto.GeneratePrivateKey().PublicKey()
	creatorStr := creator.String()

	// Test with insufficient tokens
	tokenState.Mint(creatorStr, 500) // Less than minimum threshold

	parameterChanges := map[string]interface{}{
		"voting_period": int64(172800),
	}

	err := pm.ValidateParameterProposal(creator, parameterChanges)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient tokens")

	// Test with sufficient tokens
	tokenState.Mint(creatorStr, 1500) // Now has enough tokens

	err = pm.ValidateParameterProposal(creator, parameterChanges)
	assert.NoError(t, err)

	// Test with invalid parameter changes
	invalidChanges := map[string]interface{}{
		"voting_period": "invalid_type",
	}

	err = pm.ValidateParameterProposal(creator, invalidChanges)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "voting_period must be int64")
}
