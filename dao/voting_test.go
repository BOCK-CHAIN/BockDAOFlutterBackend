package dao

import (
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
)

// TestSimpleMajorityVoting tests the simple majority voting mechanism
func TestSimpleMajorityVoting(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup voters with different token amounts
	voter1 := crypto.GeneratePrivateKey().PublicKey()
	voter2 := crypto.GeneratePrivateKey().PublicKey()
	voter3 := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		voter1.String(): 1000,
		voter2.String(): 2000,
		voter3.String(): 1500,
	}
	dao.InitialTokenDistribution(distributions)

	// Create a simple majority proposal
	proposalTx := createTestProposal(VotingTypeSimple)
	proposalHash := randomHash()

	err := dao.Processor.ProcessProposalTx(proposalTx, voter1, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Update proposal status to active
	proposal := dao.GovernanceState.Proposals[proposalHash]
	proposal.Status = ProposalStatusActive

	// Test voting with different weights
	testCases := []struct {
		voter      crypto.PublicKey
		choice     VoteChoice
		weight     uint64
		shouldFail bool
		reason     string
	}{
		{voter1, VoteChoiceYes, 500, false, "valid yes vote"},
		{voter2, VoteChoiceNo, 800, false, "valid no vote"},
		{voter3, VoteChoiceAbstain, 300, false, "valid abstain vote"},
		{voter1, VoteChoiceYes, 600, true, "duplicate vote should fail"},
		{voter2, VoteChoiceNo, 1500, true, "weight exceeds remaining balance"},
	}

	for _, tc := range testCases {
		voteTx := &VoteTx{
			Fee:        100,
			ProposalID: proposalHash,
			Choice:     tc.choice,
			Weight:     tc.weight,
			Reason:     tc.reason,
		}

		err := dao.Processor.ProcessVoteTx(voteTx, tc.voter)
		if tc.shouldFail && err == nil {
			t.Errorf("Expected error for %s, but got none", tc.reason)
		} else if !tc.shouldFail && err != nil {
			t.Errorf("Unexpected error for %s: %v", tc.reason, err)
		}
	}

	// Verify vote results
	if proposal.Results.YesVotes != 500 {
		t.Errorf("Expected 500 yes votes, got %d", proposal.Results.YesVotes)
	}
	if proposal.Results.NoVotes != 800 {
		t.Errorf("Expected 800 no votes, got %d", proposal.Results.NoVotes)
	}
	if proposal.Results.AbstainVotes != 300 {
		t.Errorf("Expected 300 abstain votes, got %d", proposal.Results.AbstainVotes)
	}
	if proposal.Results.TotalVoters != 3 {
		t.Errorf("Expected 3 total voters, got %d", proposal.Results.TotalVoters)
	}
}

// TestQuadraticVoting tests the quadratic voting mechanism
func TestQuadraticVoting(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup voters with sufficient tokens for quadratic voting
	voter1 := crypto.GeneratePrivateKey().PublicKey()
	voter2 := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		voter1.String(): 10000, // Enough for high quadratic costs
		voter2.String(): 5000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create a quadratic voting proposal
	proposalTx := createTestProposal(VotingTypeQuadratic)
	proposalHash := randomHash()

	err := dao.Processor.ProcessProposalTx(proposalTx, voter1, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Update proposal status to active
	proposal := dao.GovernanceState.Proposals[proposalHash]
	proposal.Status = ProposalStatusActive

	// Test quadratic voting costs
	testCases := []struct {
		voter        crypto.PublicKey
		choice       VoteChoice
		weight       uint64
		expectedCost uint64
		shouldFail   bool
		reason       string
	}{
		{voter1, VoteChoiceYes, 10, 100, false, "weight 10 costs 100 tokens"}, // 10^2 = 100
		{voter2, VoteChoiceNo, 20, 400, false, "weight 20 costs 400 tokens"},  // 20^2 = 400
		{voter1, VoteChoiceYes, 100, 10000, true, "weight 100 would cost 10000 + fee, exceeding balance"},
		{voter2, VoteChoiceNo, 80, 6400, true, "weight 80 would cost 6400, exceeding balance"},
	}

	for _, tc := range testCases {
		initialBalance := dao.TokenState.Balances[tc.voter.String()]

		voteTx := &VoteTx{
			Fee:        100,
			ProposalID: proposalHash,
			Choice:     tc.choice,
			Weight:     tc.weight,
			Reason:     tc.reason,
		}

		err := dao.Processor.ProcessVoteTx(voteTx, tc.voter)
		if tc.shouldFail && err == nil {
			t.Errorf("Expected error for %s, but got none", tc.reason)
		} else if !tc.shouldFail && err != nil {
			t.Errorf("Unexpected error for %s: %v", tc.reason, err)
		} else if !tc.shouldFail {
			// Verify cost was deducted correctly
			expectedBalance := initialBalance - tc.expectedCost - 100 // cost + fee
			actualBalance := dao.TokenState.Balances[tc.voter.String()]
			if actualBalance != expectedBalance {
				t.Errorf("Expected balance %d after %s, got %d", expectedBalance, tc.reason, actualBalance)
			}
		}
	}

	// Verify vote results (only successful votes)
	if proposal.Results.YesVotes != 10 {
		t.Errorf("Expected 10 yes votes, got %d", proposal.Results.YesVotes)
	}
	if proposal.Results.NoVotes != 20 {
		t.Errorf("Expected 20 no votes, got %d", proposal.Results.NoVotes)
	}
}

// TestWeightedVoting tests the token-weighted voting mechanism
func TestWeightedVoting(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup voters with different token amounts
	voter1 := crypto.GeneratePrivateKey().PublicKey()
	voter2 := crypto.GeneratePrivateKey().PublicKey()
	voter3 := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		voter1.String(): 1000,
		voter2.String(): 5000,
		voter3.String(): 2000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create a weighted voting proposal
	proposalTx := createTestProposal(VotingTypeWeighted)
	proposalHash := randomHash()

	err := dao.Processor.ProcessProposalTx(proposalTx, voter1, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Update proposal status to active
	proposal := dao.GovernanceState.Proposals[proposalHash]
	proposal.Status = ProposalStatusActive

	// Test weighted voting - each voter votes once
	// Note: voter1 created the proposal, so they have 1000 - 200 (proposal fee) = 800 tokens left
	votes := []struct {
		voter  crypto.PublicKey
		choice VoteChoice
		weight uint64
		reason string
	}{
		{voter1, VoteChoiceYes, 600, "voter1 uses 600 of remaining 800 tokens"}, // 800 - 100 fee - 600 weight = 100 left
		{voter2, VoteChoiceNo, 4000, "voter2 uses 4000 of 5000 tokens"},
		{voter3, VoteChoiceAbstain, 1500, "voter3 uses 1500 of 2000 tokens"},
	}

	for _, vote := range votes {
		voteTx := &VoteTx{
			Fee:        100,
			ProposalID: proposalHash,
			Choice:     vote.choice,
			Weight:     vote.weight,
			Reason:     vote.reason,
		}

		err := dao.Processor.ProcessVoteTx(voteTx, vote.voter)
		if err != nil {
			t.Errorf("Unexpected error for %s: %v", vote.reason, err)
		}
	}

	// Test double voting prevention
	doubleVote := &VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     VoteChoiceNo,
		Weight:     100,
		Reason:     "voter1 tries to vote again",
	}

	err = dao.Processor.ProcessVoteTx(doubleVote, voter1)
	if err == nil {
		t.Error("Expected error for double voting")
	}

	// Verify vote results
	if proposal.Results.YesVotes != 600 {
		t.Errorf("Expected 600 yes votes, got %d", proposal.Results.YesVotes)
	}
	if proposal.Results.NoVotes != 4000 {
		t.Errorf("Expected 4000 no votes, got %d", proposal.Results.NoVotes)
	}
	if proposal.Results.AbstainVotes != 1500 {
		t.Errorf("Expected 1500 abstain votes, got %d", proposal.Results.AbstainVotes)
	}
}

// TestReputationBasedVoting tests the reputation-based voting mechanism
func TestReputationBasedVoting(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup voters with different token amounts and reputation
	voter1 := crypto.GeneratePrivateKey().PublicKey()
	voter2 := crypto.GeneratePrivateKey().PublicKey()
	voter3 := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		voter1.String(): 1000,
		voter2.String(): 2000,
		voter3.String(): 1500,
	}
	dao.InitialTokenDistribution(distributions)

	// Manually set different reputation scores
	dao.GovernanceState.TokenHolders[voter1.String()].Reputation = 500 // High reputation relative to tokens
	dao.GovernanceState.TokenHolders[voter2.String()].Reputation = 100 // Low reputation relative to tokens
	dao.GovernanceState.TokenHolders[voter3.String()].Reputation = 200 // Medium reputation

	// Create a reputation-based voting proposal
	proposalTx := createTestProposal(VotingTypeReputation)
	proposalHash := randomHash()

	err := dao.Processor.ProcessProposalTx(proposalTx, voter1, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Update proposal status to active
	proposal := dao.GovernanceState.Proposals[proposalHash]
	proposal.Status = ProposalStatusActive

	// Test reputation-based voting
	testCases := []struct {
		voter      crypto.PublicKey
		choice     VoteChoice
		weight     uint64
		shouldFail bool
		reason     string
	}{
		{voter1, VoteChoiceYes, 300, false, "voter1 uses 300 of 500 reputation"},
		{voter2, VoteChoiceNo, 50, false, "voter2 uses 50 of 100 reputation"},
		{voter3, VoteChoiceAbstain, 150, false, "voter3 uses 150 of 200 reputation"},
		{voter1, VoteChoiceYes, 300, true, "voter1 already voted"},
		{voter2, VoteChoiceNo, 200, true, "voter2 exceeds reputation"},
	}

	for _, tc := range testCases {
		voteTx := &VoteTx{
			Fee:        50, // Lower fee for reputation voting
			ProposalID: proposalHash,
			Choice:     tc.choice,
			Weight:     tc.weight,
			Reason:     tc.reason,
		}

		err := dao.Processor.ProcessVoteTx(voteTx, tc.voter)
		if tc.shouldFail && err == nil {
			t.Errorf("Expected error for %s, but got none", tc.reason)
		} else if !tc.shouldFail && err != nil {
			t.Errorf("Unexpected error for %s: %v", tc.reason, err)
		}
	}

	// Verify vote results
	if proposal.Results.YesVotes != 300 {
		t.Errorf("Expected 300 yes votes, got %d", proposal.Results.YesVotes)
	}
	if proposal.Results.NoVotes != 50 {
		t.Errorf("Expected 50 no votes, got %d", proposal.Results.NoVotes)
	}
	if proposal.Results.AbstainVotes != 150 {
		t.Errorf("Expected 150 abstain votes, got %d", proposal.Results.AbstainVotes)
	}
}

// TestDoubleVotingPrevention tests that double voting is properly prevented
func TestDoubleVotingPrevention(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup voter
	voter := crypto.GeneratePrivateKey().PublicKey()
	distributions := map[string]uint64{
		voter.String(): 5000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create proposal
	proposalTx := createTestProposal(VotingTypeSimple)
	proposalHash := randomHash()

	err := dao.Processor.ProcessProposalTx(proposalTx, voter, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Update proposal status to active
	proposal := dao.GovernanceState.Proposals[proposalHash]
	proposal.Status = ProposalStatusActive

	// First vote should succeed
	firstVote := &VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     VoteChoiceYes,
		Weight:     1000,
		Reason:     "First vote",
	}

	err = dao.Processor.ProcessVoteTx(firstVote, voter)
	if err != nil {
		t.Fatalf("First vote should succeed: %v", err)
	}

	// Second vote should fail with duplicate vote error
	secondVote := &VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     VoteChoiceNo,
		Weight:     500,
		Reason:     "Second vote (should fail)",
	}

	err = dao.Processor.ProcessVoteTx(secondVote, voter)
	if err == nil {
		t.Error("Second vote should fail with duplicate vote error")
	}

	daoErr, ok := err.(*DAOError)
	if !ok {
		t.Errorf("Expected DAOError, got %T", err)
	} else if daoErr.Code != ErrDuplicateVote {
		t.Errorf("Expected duplicate vote error, got %d", daoErr.Code)
	}

	// Verify only first vote was recorded
	if proposal.Results.YesVotes != 1000 {
		t.Errorf("Expected 1000 yes votes, got %d", proposal.Results.YesVotes)
	}
	if proposal.Results.NoVotes != 0 {
		t.Errorf("Expected 0 no votes, got %d", proposal.Results.NoVotes)
	}
	if proposal.Results.TotalVoters != 1 {
		t.Errorf("Expected 1 total voter, got %d", proposal.Results.TotalVoters)
	}
}

// TestVotingValidation tests various validation scenarios
func TestVotingValidation(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup voter
	voter := crypto.GeneratePrivateKey().PublicKey()
	distributions := map[string]uint64{
		voter.String(): 1000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create proposal
	proposalTx := createTestProposal(VotingTypeSimple)
	proposalHash := randomHash()

	err := dao.Processor.ProcessProposalTx(proposalTx, voter, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Test voting on non-existent proposal
	nonExistentHash := randomHash()
	invalidProposalVote := &VoteTx{
		Fee:        100,
		ProposalID: nonExistentHash,
		Choice:     VoteChoiceYes,
		Weight:     100,
		Reason:     "Vote on non-existent proposal",
	}

	err = dao.Processor.ProcessVoteTx(invalidProposalVote, voter)
	if err == nil {
		t.Error("Expected error for voting on non-existent proposal")
	}

	// Test voting with invalid choice
	invalidChoiceVote := &VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     VoteChoice(99), // Invalid choice
		Weight:     100,
		Reason:     "Invalid choice",
	}

	err = dao.Validator.ValidateVoteTx(invalidChoiceVote, voter)
	if err == nil {
		t.Error("Expected error for invalid vote choice")
	}

	// Test voting with zero weight
	zeroWeightVote := &VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     VoteChoiceYes,
		Weight:     0,
		Reason:     "Zero weight vote",
	}

	err = dao.Validator.ValidateVoteTx(zeroWeightVote, voter)
	if err == nil {
		t.Error("Expected error for zero weight vote")
	}

	// Test voting with insufficient tokens for fee
	highFeeVote := &VoteTx{
		Fee:        2000, // More than voter's balance
		ProposalID: proposalHash,
		Choice:     VoteChoiceYes,
		Weight:     100,
		Reason:     "High fee vote",
	}

	err = dao.Validator.ValidateVoteTx(highFeeVote, voter)
	if err == nil {
		t.Error("Expected error for insufficient tokens for fee")
	}
}

// TestVotingPeriodValidation tests voting period restrictions
func TestVotingPeriodValidation(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup voter
	voter := crypto.GeneratePrivateKey().PublicKey()
	distributions := map[string]uint64{
		voter.String(): 5000,
	}
	dao.InitialTokenDistribution(distributions)

	now := time.Now().Unix()

	// Create proposal with future voting period
	proposalTx := &ProposalTx{
		Fee:          200,
		Title:        "Future Voting Test",
		Description:  "Test proposal with future voting period",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    now + 3600,  // 1 hour from now
		EndTime:      now + 90000, // 25 hours from now (meets minimum voting period)
		Threshold:    5100,
		MetadataHash: randomHash(),
	}

	proposalHash := randomHash()
	err := dao.Processor.ProcessProposalTx(proposalTx, voter, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Test voting before start time
	earlyVote := &VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     VoteChoiceYes,
		Weight:     100,
		Reason:     "Early vote",
	}

	err = dao.Validator.ValidateVoteTx(earlyVote, voter)
	if err == nil {
		t.Error("Expected error for voting before start time")
	}

	// Create proposal with past voting period
	pastProposalTx := &ProposalTx{
		Fee:          200,
		Title:        "Past Voting Test",
		Description:  "Test proposal with past voting period",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    now - 90000, // 25 hours ago
		EndTime:      now - 3600,  // 1 hour ago (meets minimum voting period)
		Threshold:    5100,
		MetadataHash: randomHash(),
	}

	pastProposalHash := randomHash()
	err = dao.Processor.ProcessProposalTx(pastProposalTx, voter, pastProposalHash)
	if err != nil {
		t.Fatalf("Failed to create past proposal: %v", err)
	}

	// Test voting after end time
	lateVote := &VoteTx{
		Fee:        100,
		ProposalID: pastProposalHash,
		Choice:     VoteChoiceYes,
		Weight:     100,
		Reason:     "Late vote",
	}

	err = dao.Validator.ValidateVoteTx(lateVote, voter)
	if err == nil {
		t.Error("Expected error for voting after end time")
	}
}

// Helper function to create a test proposal
func createTestProposal(votingType VotingType) *ProposalTx {
	now := time.Now().Unix()
	return &ProposalTx{
		Fee:          200,
		Title:        "Test Proposal",
		Description:  "This is a test proposal for voting mechanisms",
		ProposalType: ProposalTypeGeneral,
		VotingType:   votingType,
		StartTime:    now - 3600,  // 1 hour ago (active)
		EndTime:      now + 86400, // 24 hours from now
		Threshold:    5100,        // 51%
		MetadataHash: randomHash(),
	}
}
