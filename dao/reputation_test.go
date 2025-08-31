package dao

import (
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
)

// TestReputationSystemInitialization tests the initialization of the reputation system
func TestReputationSystemInitialization(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Test that reputation system is properly initialized
	if dao.ReputationSystem == nil {
		t.Fatal("ReputationSystem should be initialized")
	}

	config := dao.ReputationSystem.GetReputationConfig()
	if config == nil {
		t.Fatal("ReputationConfig should be initialized")
	}

	// Test default configuration values
	if config.BaseReputation != 100 {
		t.Errorf("Expected BaseReputation 100, got %d", config.BaseReputation)
	}

	if config.ProposalCreationBonus != 50 {
		t.Errorf("Expected ProposalCreationBonus 50, got %d", config.ProposalCreationBonus)
	}

	if config.VotingParticipation != 10 {
		t.Errorf("Expected VotingParticipation 10, got %d", config.VotingParticipation)
	}
}

// TestInitializeUserReputation tests reputation initialization for new users
func TestInitializeUserReputation(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	user1 := crypto.GeneratePrivateKey().PublicKey()
	user2 := crypto.GeneratePrivateKey().PublicKey()

	// Initialize users with different token balances
	distributions := map[string]uint64{
		user1.String(): 1000,
		user2.String(): 5000,
	}
	dao.InitialTokenDistribution(distributions)

	// Check reputation initialization
	reputation1 := dao.GetUserReputation(user1)
	reputation2 := dao.GetUserReputation(user2)

	// user1: base(100) + tokens(1000/100=10) = 110
	expectedReputation1 := uint64(110)
	if reputation1 != expectedReputation1 {
		t.Errorf("Expected user1 reputation %d, got %d", expectedReputation1, reputation1)
	}

	// user2: base(100) + tokens(5000/100=50) = 150
	expectedReputation2 := uint64(150)
	if reputation2 != expectedReputation2 {
		t.Errorf("Expected user2 reputation %d, got %d", expectedReputation2, reputation2)
	}
}

// TestReputationForProposalCreation tests reputation updates when creating proposals
func TestReputationForProposalCreation(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	creator := crypto.GeneratePrivateKey().PublicKey()
	distributions := map[string]uint64{
		creator.String(): 2000,
	}
	dao.InitialTokenDistribution(distributions)

	initialReputation := dao.GetUserReputation(creator)

	// Create a proposal
	proposalTx := createTestProposal(VotingTypeSimple)
	proposalHash := randomHash()

	err := dao.Processor.ProcessProposalTx(proposalTx, creator, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Check reputation increase
	newReputation := dao.GetUserReputation(creator)
	expectedIncrease := dao.ReputationSystem.GetReputationConfig().ProposalCreationBonus

	if newReputation != initialReputation+expectedIncrease {
		t.Errorf("Expected reputation increase of %d, got %d", expectedIncrease, newReputation-initialReputation)
	}
}

// TestReputationForVoting tests reputation updates when voting
func TestReputationForVoting(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	creator := crypto.GeneratePrivateKey().PublicKey()
	voter := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		creator.String(): 2000,
		voter.String():   1500,
	}
	dao.InitialTokenDistribution(distributions)

	// Create a proposal
	proposalTx := createTestProposal(VotingTypeSimple)
	proposalHash := randomHash()

	err := dao.Processor.ProcessProposalTx(proposalTx, creator, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Activate the proposal
	proposal := dao.GovernanceState.Proposals[proposalHash]
	proposal.Status = ProposalStatusActive

	initialVoterReputation := dao.GetUserReputation(voter)

	// Vote on the proposal
	voteTx := &VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     VoteChoiceYes,
		Weight:     500,
		Reason:     "Test vote",
	}

	err = dao.Processor.ProcessVoteTx(voteTx, voter)
	if err != nil {
		t.Fatalf("Failed to vote: %v", err)
	}

	// Check reputation increase
	newVoterReputation := dao.GetUserReputation(voter)
	expectedIncrease := dao.ReputationSystem.GetReputationConfig().VotingParticipation

	if newVoterReputation != initialVoterReputation+expectedIncrease {
		t.Errorf("Expected reputation increase of %d, got %d", expectedIncrease, newVoterReputation-initialVoterReputation)
	}
}

// TestReputationBasedVotingMechanism tests the reputation-based voting mechanism
func TestReputationBasedVotingMechanism(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	creator := crypto.GeneratePrivateKey().PublicKey()
	voter1 := crypto.GeneratePrivateKey().PublicKey()
	voter2 := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		creator.String(): 3000,
		voter1.String():  2000,
		voter2.String():  1500,
	}
	dao.InitialTokenDistribution(distributions)

	// Manually adjust reputation scores for testing
	dao.GovernanceState.TokenHolders[voter1.String()].Reputation = 500
	dao.GovernanceState.TokenHolders[voter2.String()].Reputation = 200

	// Create a reputation-based voting proposal
	proposalTx := createTestProposal(VotingTypeReputation)
	proposalHash := randomHash()

	err := dao.Processor.ProcessProposalTx(proposalTx, creator, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Activate the proposal
	proposal := dao.GovernanceState.Proposals[proposalHash]
	proposal.Status = ProposalStatusActive

	// Test voting with reputation limits
	testCases := []struct {
		voter      crypto.PublicKey
		choice     VoteChoice
		weight     uint64
		shouldFail bool
		reason     string
	}{
		{voter1, VoteChoiceYes, 300, false, "voter1 uses 300 of 500 reputation"},
		{voter2, VoteChoiceNo, 150, false, "voter2 uses 150 of 200 reputation"},
		{voter1, VoteChoiceYes, 300, true, "voter1 already voted"},
		{voter2, VoteChoiceNo, 300, true, "voter2 exceeds reputation (200)"},
	}

	for _, tc := range testCases {
		voteTx := &VoteTx{
			Fee:        50,
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
	if proposal.Results.NoVotes != 150 {
		t.Errorf("Expected 150 no votes, got %d", proposal.Results.NoVotes)
	}
}

// TestReputationForProposalOutcome tests reputation updates based on proposal outcomes
func TestReputationForProposalOutcome(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	creator1 := crypto.GeneratePrivateKey().PublicKey()
	creator2 := crypto.GeneratePrivateKey().PublicKey()
	voter := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		creator1.String(): 3000,
		creator2.String(): 3000,
		voter.String():    5000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create two proposals
	proposalTx1 := createTestProposal(VotingTypeSimple)
	proposalHash1 := randomHash()
	err := dao.Processor.ProcessProposalTx(proposalTx1, creator1, proposalHash1)
	if err != nil {
		t.Fatalf("Failed to create proposal 1: %v", err)
	}

	proposalTx2 := createTestProposal(VotingTypeSimple)
	proposalHash2 := randomHash()
	err = dao.Processor.ProcessProposalTx(proposalTx2, creator2, proposalHash2)
	if err != nil {
		t.Fatalf("Failed to create proposal 2: %v", err)
	}

	// Activate proposals
	dao.GovernanceState.Proposals[proposalHash1].Status = ProposalStatusActive
	dao.GovernanceState.Proposals[proposalHash2].Status = ProposalStatusActive

	initialReputation1 := dao.GetUserReputation(creator1)
	initialReputation2 := dao.GetUserReputation(creator2)

	// Vote to pass proposal 1
	voteTx1 := &VoteTx{
		Fee:        100,
		ProposalID: proposalHash1,
		Choice:     VoteChoiceYes,
		Weight:     2000, // Enough to pass
		Reason:     "Pass proposal 1",
	}
	err = dao.Processor.ProcessVoteTx(voteTx1, voter)
	if err != nil {
		t.Fatalf("Failed to vote on proposal 1: %v", err)
	}

	// Vote to reject proposal 2
	voteTx2 := &VoteTx{
		Fee:        100,
		ProposalID: proposalHash2,
		Choice:     VoteChoiceNo,
		Weight:     2000, // Enough to reject
		Reason:     "Reject proposal 2",
	}
	err = dao.Processor.ProcessVoteTx(voteTx2, voter)
	if err != nil {
		t.Fatalf("Failed to vote on proposal 2: %v", err)
	}

	// Update proposal statuses (simulate end of voting period)
	dao.GovernanceState.Proposals[proposalHash1].EndTime = time.Now().Unix() - 1
	dao.GovernanceState.Proposals[proposalHash2].EndTime = time.Now().Unix() - 1

	err = dao.Processor.UpdateProposalStatus(proposalHash1)
	if err != nil {
		t.Fatalf("Failed to update proposal 1 status: %v", err)
	}

	err = dao.Processor.UpdateProposalStatus(proposalHash2)
	if err != nil {
		t.Fatalf("Failed to update proposal 2 status: %v", err)
	}

	// Check reputation changes
	newReputation1 := dao.GetUserReputation(creator1)
	newReputation2 := dao.GetUserReputation(creator2)

	config := dao.ReputationSystem.GetReputationConfig()

	// Creator1 should get bonus for passed proposal
	expectedReputation1 := initialReputation1 + config.ProposalPassedBonus
	if newReputation1 != expectedReputation1 {
		t.Errorf("Expected creator1 reputation %d, got %d", expectedReputation1, newReputation1)
	}

	// Creator2 should get penalty for rejected proposal
	expectedReputation2 := initialReputation2 - config.ProposalRejectedPenalty
	if newReputation2 != expectedReputation2 {
		t.Errorf("Expected creator2 reputation %d, got %d", expectedReputation2, newReputation2)
	}
}

// TestReputationRanking tests the reputation ranking functionality
func TestReputationRanking(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	user1 := crypto.GeneratePrivateKey().PublicKey()
	user2 := crypto.GeneratePrivateKey().PublicKey()
	user3 := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		user1.String(): 1000,
		user2.String(): 2000,
		user3.String(): 1500,
	}
	dao.InitialTokenDistribution(distributions)

	// Manually set different reputation scores
	dao.GovernanceState.TokenHolders[user1.String()].Reputation = 300
	dao.GovernanceState.TokenHolders[user2.String()].Reputation = 500
	dao.GovernanceState.TokenHolders[user3.String()].Reputation = 200

	// Get ranking
	ranking := dao.GetReputationRanking()

	if len(ranking) != 3 {
		t.Fatalf("Expected 3 users in ranking, got %d", len(ranking))
	}

	// Should be sorted by reputation (highest first)
	if ranking[0].Reputation != 500 {
		t.Errorf("Expected highest reputation 500, got %d", ranking[0].Reputation)
	}
	if ranking[1].Reputation != 300 {
		t.Errorf("Expected second highest reputation 300, got %d", ranking[1].Reputation)
	}
	if ranking[2].Reputation != 200 {
		t.Errorf("Expected lowest reputation 200, got %d", ranking[2].Reputation)
	}
}

// TestReputationStats tests the reputation statistics functionality
func TestReputationStats(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	user1 := crypto.GeneratePrivateKey().PublicKey()
	user2 := crypto.GeneratePrivateKey().PublicKey()
	user3 := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		user1.String(): 1000,
		user2.String(): 2000,
		user3.String(): 1500,
	}
	dao.InitialTokenDistribution(distributions)

	// Set last active times (user2 is inactive)
	now := time.Now().Unix()
	dao.GovernanceState.TokenHolders[user1.String()].LastActive = now - 3600      // 1 hour ago (active)
	dao.GovernanceState.TokenHolders[user2.String()].LastActive = now - 8*24*3600 // 8 days ago (inactive)
	dao.GovernanceState.TokenHolders[user3.String()].LastActive = now - 1800      // 30 minutes ago (active)

	stats := dao.GetReputationStats()

	if stats.TotalUsers != 3 {
		t.Errorf("Expected 3 total users, got %d", stats.TotalUsers)
	}

	if stats.ActiveUsers != 2 {
		t.Errorf("Expected 2 active users, got %d", stats.ActiveUsers)
	}

	// Check that stats are reasonable
	if stats.TotalReputation == 0 {
		t.Error("Expected non-zero total reputation")
	}

	if stats.AverageReputation == 0 {
		t.Error("Expected non-zero average reputation")
	}
}

// TestInactivityDecay tests the inactivity decay functionality
func TestInactivityDecay(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	user := crypto.GeneratePrivateKey().PublicKey()
	distributions := map[string]uint64{
		user.String(): 2000,
	}
	dao.InitialTokenDistribution(distributions)

	// Set user as inactive for more than decay period
	now := time.Now().Unix()
	decayPeriod := dao.ReputationSystem.GetReputationConfig().DecayPeriodDays
	dao.GovernanceState.TokenHolders[user.String()].LastActive = now - (decayPeriod+10)*24*3600

	initialReputation := dao.GetUserReputation(user)

	// Apply decay
	dao.ApplyInactivityDecay()

	newReputation := dao.GetUserReputation(user)

	// Reputation should have decreased
	if newReputation >= initialReputation {
		t.Errorf("Expected reputation to decrease from %d, but got %d", initialReputation, newReputation)
	}

	// Should not go below minimum
	minReputation := dao.ReputationSystem.GetReputationConfig().MinReputation
	if newReputation < minReputation {
		t.Errorf("Reputation should not go below minimum %d, got %d", minReputation, newReputation)
	}
}

// TestReputationConfigUpdate tests updating reputation configuration
func TestReputationConfigUpdate(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Test valid config update
	newConfig := &ReputationConfig{
		BaseReputation:          200,
		ProposalCreationBonus:   75,
		VotingParticipation:     15,
		ProposalPassedBonus:     150,
		ProposalRejectedPenalty: 30,
		InactivityDecayRate:     0.01,
		MaxReputation:           15000,
		MinReputation:           20,
		DecayPeriodDays:         45,
	}

	err := dao.UpdateReputationConfig(newConfig)
	if err != nil {
		t.Fatalf("Failed to update reputation config: %v", err)
	}

	// Verify config was updated
	currentConfig := dao.GetReputationConfig()
	if currentConfig.BaseReputation != 200 {
		t.Errorf("Expected BaseReputation 200, got %d", currentConfig.BaseReputation)
	}

	// Test invalid config (max <= min)
	invalidConfig := &ReputationConfig{
		BaseReputation:          100,
		ProposalCreationBonus:   50,
		VotingParticipation:     10,
		ProposalPassedBonus:     100,
		ProposalRejectedPenalty: 25,
		InactivityDecayRate:     0.005,
		MaxReputation:           100,
		MinReputation:           200, // Invalid: min > max
		DecayPeriodDays:         30,
	}

	err = dao.UpdateReputationConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid config (min > max)")
	}
}

// TestReputationRecalculation tests the comprehensive reputation recalculation
func TestReputationRecalculation(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	creator := crypto.GeneratePrivateKey().PublicKey()
	voter := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		creator.String(): 3000,
		voter.String():   2000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create and pass a proposal
	proposalTx := createTestProposal(VotingTypeSimple)
	proposalHash := randomHash()

	err := dao.Processor.ProcessProposalTx(proposalTx, creator, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Activate and vote
	dao.GovernanceState.Proposals[proposalHash].Status = ProposalStatusActive

	voteTx := &VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     VoteChoiceYes,
		Weight:     1500,
		Reason:     "Test vote",
	}

	err = dao.Processor.ProcessVoteTx(voteTx, voter)
	if err != nil {
		t.Fatalf("Failed to vote: %v", err)
	}

	// End voting and update status
	dao.GovernanceState.Proposals[proposalHash].EndTime = time.Now().Unix() - 1
	err = dao.Processor.UpdateProposalStatus(proposalHash)
	if err != nil {
		t.Fatalf("Failed to update proposal status: %v", err)
	}

	// Store current reputations
	creatorRepBefore := dao.GetUserReputation(creator)
	voterRepBefore := dao.GetUserReputation(voter)

	// Recalculate all reputation
	dao.RecalculateAllReputation()

	// Check that reputations are still reasonable
	creatorRepAfter := dao.GetUserReputation(creator)
	voterRepAfter := dao.GetUserReputation(voter)

	if creatorRepAfter == 0 {
		t.Error("Creator reputation should not be zero after recalculation")
	}

	if voterRepAfter == 0 {
		t.Error("Voter reputation should not be zero after recalculation")
	}

	// Both should have some reputation from their activities
	config := dao.ReputationSystem.GetReputationConfig()
	if creatorRepAfter < config.BaseReputation {
		t.Errorf("Creator reputation should be at least base reputation %d, got %d", config.BaseReputation, creatorRepAfter)
	}

	if voterRepAfter < config.BaseReputation {
		t.Errorf("Voter reputation should be at least base reputation %d, got %d", config.BaseReputation, voterRepAfter)
	}

	t.Logf("Creator reputation: before=%d, after=%d", creatorRepBefore, creatorRepAfter)
	t.Logf("Voter reputation: before=%d, after=%d", voterRepBefore, voterRepAfter)
}

// TestUserReputationHistory tests the user reputation history functionality
func TestUserReputationHistory(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	user := crypto.GeneratePrivateKey().PublicKey()
	distributions := map[string]uint64{
		user.String(): 2000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create a proposal (should generate reputation event)
	proposalTx := createTestProposal(VotingTypeSimple)
	proposalHash := randomHash()

	err := dao.Processor.ProcessProposalTx(proposalTx, user, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Get reputation history
	history := dao.GetUserReputationHistory(user)
	if history == nil {
		t.Fatal("Expected reputation history, got nil")
	}

	if history.User.String() != user.String() {
		t.Errorf("Expected user %s, got %s", user.String(), history.User.String())
	}

	if history.CurrentReputation == 0 {
		t.Error("Expected non-zero current reputation")
	}

	if len(history.Events) == 0 {
		t.Error("Expected at least one reputation event")
	}

	// Check that we have a proposal creation event
	foundProposalEvent := false
	for _, event := range history.Events {
		if event.Type == ReputationEventProposalCreated {
			foundProposalEvent = true
			if event.Impact <= 0 {
				t.Error("Expected positive impact for proposal creation")
			}
			break
		}
	}

	if !foundProposalEvent {
		t.Error("Expected to find proposal creation event in history")
	}
}

// Helper function to create a random hash for testing (using existing randomHash from dao_test.go)
