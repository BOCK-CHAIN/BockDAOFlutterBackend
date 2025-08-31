package dao

import (
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParameterManagementIntegration(t *testing.T) {
	// Create a full DAO instance
	dao := NewDAO("GOVTOKEN", "Governance Token", 18)

	// Create test users
	creator := crypto.GeneratePrivateKey()
	voter1 := crypto.GeneratePrivateKey()
	voter2 := crypto.GeneratePrivateKey()
	executor := crypto.GeneratePrivateKey()

	// Initialize token distribution
	distributions := map[string]uint64{
		creator.PublicKey().String(): 10000,
		voter1.PublicKey().String():  5000,
		voter2.PublicKey().String():  3000,
	}

	err := dao.InitialTokenDistribution(distributions)
	require.NoError(t, err)

	// Test 1: Create a parameter change proposal
	t.Run("CreateParameterProposal", func(t *testing.T) {
		parameterChanges := map[string]interface{}{
			"voting_period":     int64(172800), // 48 hours
			"quorum_threshold":  uint64(3000),
			"passing_threshold": uint64(6000), // 60%
		}

		proposalID, err := dao.CreateParameterProposal(
			creator.PublicKey(),
			parameterChanges,
			"Improve governance parameters for better participation",
			time.Now().Unix()+7200, // Effective in 2 hours
			VotingTypeSimple,
			time.Now().Unix()+600,  // Start in 10 minutes
			time.Now().Unix()+3600, // End in 1 hour
			2000,
		)

		require.NoError(t, err)
		assert.NotEqual(t, types.Hash{}, proposalID)

		// Verify proposal exists
		proposal, err := dao.GetProposal(proposalID)
		require.NoError(t, err)
		assert.Equal(t, ProposalTypeParameter, proposal.ProposalType)
		assert.Equal(t, creator.PublicKey(), proposal.Creator)
	})

	// Test 2: Vote on parameter proposal
	t.Run("VoteOnParameterProposal", func(t *testing.T) {
		// First create a proposal that's already active
		parameterChanges := map[string]interface{}{
			"min_proposal_threshold": uint64(1500),
		}

		proposalID, err := dao.CreateParameterProposal(
			creator.PublicKey(),
			parameterChanges,
			"Adjust minimum proposal threshold",
			time.Now().Unix()+7200,
			VotingTypeSimple,
			time.Now().Unix()-600,  // Started 10 minutes ago
			time.Now().Unix()+3600, // Ends in 1 hour
			2000,
		)
		require.NoError(t, err)

		// Update proposal status to active
		dao.UpdateAllProposalStatuses()

		// Vote on the proposal
		voteTx1 := &VoteTx{
			Fee:        100,
			ProposalID: proposalID,
			Choice:     VoteChoiceYes,
			Weight:     3000,
			Reason:     "Support the threshold adjustment",
		}

		err = dao.ProcessDAOTransaction(voteTx1, voter1.PublicKey(), types.Hash{1})
		require.NoError(t, err)

		voteTx2 := &VoteTx{
			Fee:        100,
			ProposalID: proposalID,
			Choice:     VoteChoiceYes,
			Weight:     2000,
			Reason:     "Agree with the change",
		}

		err = dao.ProcessDAOTransaction(voteTx2, voter2.PublicKey(), types.Hash{2})
		require.NoError(t, err)

		// Verify votes were recorded
		votes, err := dao.GetVotes(proposalID)
		require.NoError(t, err)
		assert.Len(t, votes, 2)

		// Check proposal results
		proposal, err := dao.GetProposal(proposalID)
		require.NoError(t, err)
		assert.Equal(t, uint64(5000), proposal.Results.YesVotes)
		assert.Equal(t, uint64(2), proposal.Results.TotalVoters)
	})

	// Test 3: Execute parameter changes
	t.Run("ExecuteParameterChanges", func(t *testing.T) {
		// Create a proposal that will pass
		parameterChanges := map[string]interface{}{
			"treasury_threshold": uint64(7000),
		}

		proposalID, err := dao.CreateParameterProposal(
			creator.PublicKey(),
			parameterChanges,
			"Increase treasury threshold for security",
			time.Now().Unix()+1800, // Effective in 30 minutes
			VotingTypeSimple,
			time.Now().Unix()-3600, // Started 1 hour ago
			time.Now().Unix()-600,  // Ended 10 minutes ago
			2000,
		)
		require.NoError(t, err)

		// Simulate voting that passes the proposal
		proposal, err := dao.GetProposal(proposalID)
		require.NoError(t, err)

		// Manually set proposal as passed for testing
		proposal.Status = ProposalStatusPassed
		proposal.Results.YesVotes = 8000
		proposal.Results.NoVotes = 1000
		proposal.Results.TotalVoters = 3
		proposal.Results.Passed = true

		// Store original treasury threshold
		originalThreshold := dao.GetParameterConfig().TreasuryThreshold

		// Execute parameter changes
		err = dao.ExecuteParameterChanges(proposalID, executor.PublicKey())
		require.NoError(t, err)

		// Verify parameter was changed
		newConfig := dao.GetParameterConfig()
		assert.Equal(t, uint64(7000), newConfig.TreasuryThreshold)
		assert.NotEqual(t, originalThreshold, newConfig.TreasuryThreshold)

		// Verify proposal status updated
		proposal, err = dao.GetProposal(proposalID)
		require.NoError(t, err)
		assert.Equal(t, ProposalStatusExecuted, proposal.Status)

		// Verify parameter history was recorded
		history := dao.GetParameterHistory("treasury_threshold")
		assert.Len(t, history, 1)
		assert.Equal(t, originalThreshold, history[0].OldValue)
		assert.Equal(t, uint64(7000), history[0].NewValue)
		assert.Equal(t, executor.PublicKey(), history[0].ChangedBy)
	})

	// Test 4: Parameter validation and constraints
	t.Run("ParameterValidationAndConstraints", func(t *testing.T) {
		// Test parameter constraints
		constraints := dao.GetParameterConstraints("voting_period")
		assert.Equal(t, "int64", constraints["type"])
		assert.Contains(t, constraints, "min")
		assert.Contains(t, constraints, "max")

		// Test parameter value retrieval
		value, err := dao.GetParameterValue("min_proposal_threshold")
		require.NoError(t, err)
		assert.Equal(t, uint64(1000), value)

		// Test listing all parameters
		allParams := dao.ListAllParameters()
		assert.Contains(t, allParams, "min_proposal_threshold")
		assert.Contains(t, allParams, "voting_period")
		assert.Contains(t, allParams, "quorum_threshold")

		// Test parameter change validation
		invalidChanges := map[string]interface{}{
			"min_proposal_threshold": uint64(0), // Invalid: zero threshold
		}

		err = dao.ValidateParameterProposal(creator.PublicKey(), invalidChanges)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "minimum proposal threshold must be greater than zero")

		// Test valid parameter changes
		validChanges := map[string]interface{}{
			"voting_period": int64(259200), // 72 hours
		}

		err = dao.ValidateParameterProposal(creator.PublicKey(), validChanges)
		assert.NoError(t, err)
	})

	// Test 5: Parameter change permissions and security
	t.Run("ParameterChangePermissions", func(t *testing.T) {
		// Initialize security system
		founders := []crypto.PublicKey{creator.PublicKey()}
		err := dao.InitializeFounderRoles(founders)
		require.NoError(t, err)

		parameterChanges := map[string]interface{}{
			"emergency_pause_enabled": false,
		}

		// Create parameter proposal transaction
		paramTx := &ParameterProposalTx{
			Fee:              200,
			ParameterChanges: parameterChanges,
			Justification:    "Disable emergency pause for testing",
			EffectiveTime:    time.Now().Unix() + 7200,
			ProposalType:     ProposalTypeParameter,
			VotingType:       VotingTypeSimple,
			StartTime:        time.Now().Unix() + 600,
			EndTime:          time.Now().Unix() + 3600,
			Threshold:        2000,
		}

		txHash := types.Hash{9, 9, 9}

		// Process with security validation
		err = dao.SecureProcessDAOTransaction(paramTx, creator.PublicKey(), txHash)
		require.NoError(t, err)

		// Verify proposal was created
		proposal, err := dao.GetProposal(txHash)
		require.NoError(t, err)
		assert.Equal(t, ProposalTypeParameter, proposal.ProposalType)
	})

	// Test 6: Parameter change history and auditing
	t.Run("ParameterChangeHistory", func(t *testing.T) {
		// Get all parameter history
		allHistory := dao.GetAllParameterHistory()

		// Should have history from previous tests
		assert.NotEmpty(t, allHistory)

		// Check specific parameter history
		if treasuryHistory, exists := allHistory["treasury_threshold"]; exists {
			assert.NotEmpty(t, treasuryHistory)

			// Verify history entry structure
			change := treasuryHistory[0]
			assert.NotEmpty(t, change.Parameter)
			assert.NotNil(t, change.OldValue)
			assert.NotNil(t, change.NewValue)
			assert.NotEqual(t, crypto.PublicKey{}, change.ChangedBy)
			assert.Greater(t, change.ChangedAt, int64(0))
			assert.NotEqual(t, types.Hash{}, change.ProposalID)
		}
	})

	// Test 7: Parameter change restrictions
	t.Run("ParameterChangeRestrictions", func(t *testing.T) {
		// Test parameter change allowance
		allowed, reason := dao.IsParameterChangeAllowed("min_proposal_threshold", uint64(600000))
		assert.False(t, allowed)
		assert.Contains(t, reason, "minimum proposal threshold cannot exceed 50% of total supply")

		// Test valid parameter change
		allowed, reason = dao.IsParameterChangeAllowed("voting_period", int64(172800))
		assert.True(t, allowed)
		assert.Empty(t, reason)

		// Test treasury-related restrictions
		dao.AddTreasuryFunds(25000) // Add some treasury funds

		allowed, reason = dao.IsParameterChangeAllowed("max_treasury_withdraw", uint64(50000))
		assert.False(t, allowed)
		assert.Contains(t, reason, "max treasury withdraw cannot exceed current treasury balance")

		allowed, reason = dao.IsParameterChangeAllowed("max_treasury_withdraw", uint64(20000))
		assert.True(t, allowed)
		assert.Empty(t, reason)
	})
}

func TestParameterManagementWithQuadraticVoting(t *testing.T) {
	dao := NewDAO("GOVTOKEN", "Governance Token", 18)

	// Create test users
	creator := crypto.GeneratePrivateKey()
	voter1 := crypto.GeneratePrivateKey()
	voter2 := crypto.GeneratePrivateKey()

	// Initialize token distribution
	distributions := map[string]uint64{
		creator.PublicKey().String(): 15000,
		voter1.PublicKey().String():  10000,
		voter2.PublicKey().String():  8000,
	}

	err := dao.InitialTokenDistribution(distributions)
	require.NoError(t, err)

	// Create parameter proposal with quadratic voting
	parameterChanges := map[string]interface{}{
		"quadratic_voting_cost": uint64(2), // Increase quadratic voting cost
	}

	proposalID, err := dao.CreateParameterProposal(
		creator.PublicKey(),
		parameterChanges,
		"Adjust quadratic voting cost",
		time.Now().Unix()+7200,
		VotingTypeQuadratic, // Use quadratic voting
		time.Now().Unix()-600,
		time.Now().Unix()+3600,
		3000,
	)
	require.NoError(t, err)

	// Update proposal status to active
	dao.UpdateAllProposalStatuses()

	// Cast quadratic votes
	voteTx1 := &VoteTx{
		Fee:        100,
		ProposalID: proposalID,
		Choice:     VoteChoiceYes,
		Weight:     10, // Cost will be 100 tokens (10^2)
		Reason:     "Support quadratic voting adjustment",
	}

	err = dao.ProcessDAOTransaction(voteTx1, voter1.PublicKey(), types.Hash{1})
	require.NoError(t, err)

	voteTx2 := &VoteTx{
		Fee:        100,
		ProposalID: proposalID,
		Choice:     VoteChoiceNo,
		Weight:     8, // Cost will be 64 tokens (8^2)
		Reason:     "Current cost is fine",
	}

	err = dao.ProcessDAOTransaction(voteTx2, voter2.PublicKey(), types.Hash{2})
	require.NoError(t, err)

	// Verify votes were recorded with correct weights
	proposal, err := dao.GetProposal(proposalID)
	require.NoError(t, err)
	assert.Equal(t, uint64(10), proposal.Results.YesVotes)
	assert.Equal(t, uint64(8), proposal.Results.NoVotes)

	// Verify token costs were deducted correctly
	voter1Balance := dao.GetTokenBalance(voter1.PublicKey())
	voter2Balance := dao.GetTokenBalance(voter2.PublicKey())

	// voter1: 10000 - 100 (quadratic cost) - 100 (fee) = 9800
	assert.Equal(t, uint64(9800), voter1Balance)

	// voter2: 8000 - 64 (quadratic cost) - 100 (fee) = 7836
	assert.Equal(t, uint64(7836), voter2Balance)
}

func TestParameterManagementErrorCases(t *testing.T) {
	dao := NewDAO("GOVTOKEN", "Governance Token", 18)

	creator := crypto.GeneratePrivateKey()

	// Initialize with minimal tokens
	distributions := map[string]uint64{
		creator.PublicKey().String(): 500, // Less than minimum threshold
	}

	err := dao.InitialTokenDistribution(distributions)
	require.NoError(t, err)

	// Test 1: Insufficient tokens for parameter proposal
	t.Run("InsufficientTokensForProposal", func(t *testing.T) {
		parameterChanges := map[string]interface{}{
			"voting_period": int64(172800),
		}

		_, err := dao.CreateParameterProposal(
			creator.PublicKey(),
			parameterChanges,
			"Test proposal",
			time.Now().Unix()+7200,
			VotingTypeSimple,
			time.Now().Unix()+600,
			time.Now().Unix()+3600,
			1000,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient tokens")
	})

	// Test 2: Invalid parameter values
	t.Run("InvalidParameterValues", func(t *testing.T) {
		// Give creator enough tokens
		dao.MintTokens(creator.PublicKey(), 2000)

		invalidChanges := map[string]interface{}{
			"voting_period":          int64(0),      // Invalid: zero period
			"min_proposal_threshold": uint64(0),     // Invalid: zero threshold
			"passing_threshold":      uint64(15000), // Invalid: > 100%
			"unknown_parameter":      uint64(100),   // Invalid: unknown param
		}

		for param, value := range invalidChanges {
			changes := map[string]interface{}{param: value}

			_, err := dao.CreateParameterProposal(
				creator.PublicKey(),
				changes,
				"Invalid test proposal",
				time.Now().Unix()+7200,
				VotingTypeSimple,
				time.Now().Unix()+600,
				time.Now().Unix()+3600,
				1000,
			)

			assert.Error(t, err, "Expected error for parameter: %s", param)
		}
	})

	// Test 3: Invalid timing
	t.Run("InvalidTiming", func(t *testing.T) {
		parameterChanges := map[string]interface{}{
			"voting_period": int64(172800),
		}

		// Start time after end time
		_, err := dao.CreateParameterProposal(
			creator.PublicKey(),
			parameterChanges,
			"Invalid timing proposal",
			time.Now().Unix()+7200,
			VotingTypeSimple,
			time.Now().Unix()+3600, // Start after end
			time.Now().Unix()+1800, // End before start
			1000,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start time must be before end time")

		// Effective time before end time
		_, err = dao.CreateParameterProposal(
			creator.PublicKey(),
			parameterChanges,
			"Invalid effective time proposal",
			time.Now().Unix()+1800, // Effective before end
			VotingTypeSimple,
			time.Now().Unix()+600,
			time.Now().Unix()+3600,
			1000,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "effective time must be after voting ends")
	})

	// Test 4: Execute non-existent proposal
	t.Run("ExecuteNonExistentProposal", func(t *testing.T) {
		nonExistentID := types.Hash{9, 9, 9}

		err := dao.ExecuteParameterChanges(nonExistentID, creator.PublicKey())
		assert.Error(t, err)
		assert.Equal(t, ErrProposalNotFoundError, err)
	})

	// Test 5: Execute non-parameter proposal
	t.Run("ExecuteNonParameterProposal", func(t *testing.T) {
		// Create a regular proposal
		regularTx := &ProposalTx{
			Fee:          200,
			Title:        "Regular Proposal",
			Description:  "Not a parameter proposal",
			ProposalType: ProposalTypeGeneral,
			VotingType:   VotingTypeSimple,
			StartTime:    time.Now().Unix() - 3600,
			EndTime:      time.Now().Unix() - 600,
			Threshold:    1000,
		}

		proposalID := types.Hash{8, 8, 8}
		err := dao.ProcessDAOTransaction(regularTx, creator.PublicKey(), proposalID)
		require.NoError(t, err)

		// Try to execute as parameter proposal
		err = dao.ExecuteParameterChanges(proposalID, creator.PublicKey())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "proposal is not a parameter change proposal")
	})
}
