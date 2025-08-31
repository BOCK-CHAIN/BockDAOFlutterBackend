package dao

import (
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
)

// TestVotingMechanismsIntegration tests all voting mechanisms in a comprehensive scenario
func TestVotingMechanismsIntegration(t *testing.T) {
	dao := NewDAO("PXGOV", "ProjectX Governance Token", 18)

	// Setup diverse group of token holders
	founder := crypto.GeneratePrivateKey().PublicKey()
	developer1 := crypto.GeneratePrivateKey().PublicKey()
	developer2 := crypto.GeneratePrivateKey().PublicKey()
	community1 := crypto.GeneratePrivateKey().PublicKey()
	community2 := crypto.GeneratePrivateKey().PublicKey()
	community3 := crypto.GeneratePrivateKey().PublicKey()

	// Initial token distribution
	distributions := map[string]uint64{
		founder.String():    10000, // High token holder
		developer1.String(): 5000,  // Medium token holder
		developer2.String(): 3000,  // Medium token holder
		community1.String(): 2000,  // Small token holder
		community2.String(): 1500,  // Small token holder
		community3.String(): 1000,  // Small token holder
	}

	err := dao.InitialTokenDistribution(distributions)
	if err != nil {
		t.Fatalf("Failed to distribute tokens: %v", err)
	}

	// Adjust reputation scores to create different voting scenarios
	dao.GovernanceState.TokenHolders[founder.String()].Reputation = 8000    // High reputation
	dao.GovernanceState.TokenHolders[developer1.String()].Reputation = 6000 // High reputation
	dao.GovernanceState.TokenHolders[developer2.String()].Reputation = 4000 // Medium reputation
	dao.GovernanceState.TokenHolders[community1.String()].Reputation = 1000 // Low reputation
	dao.GovernanceState.TokenHolders[community2.String()].Reputation = 800  // Low reputation
	dao.GovernanceState.TokenHolders[community3.String()].Reputation = 500  // Low reputation

	t.Run("SimpleVotingScenario", func(t *testing.T) {
		testSimpleVotingScenario(t, dao, founder, developer1, developer2, community1)
	})

	t.Run("QuadraticVotingScenario", func(t *testing.T) {
		testQuadraticVotingScenario(t, dao, founder, developer1, community1, community2)
	})

	t.Run("WeightedVotingScenario", func(t *testing.T) {
		testWeightedVotingScenario(t, dao, founder, developer1, developer2, community1, community2)
	})

	t.Run("ReputationVotingScenario", func(t *testing.T) {
		testReputationVotingScenario(t, dao, founder, developer1, developer2, community1, community2, community3)
	})

	t.Run("MixedVotingValidation", func(t *testing.T) {
		// Create a fresh DAO for validation tests to avoid token balance issues
		validationDAO := NewDAO("PXGOV", "ProjectX Governance Token", 18)
		testVoter := crypto.GeneratePrivateKey().PublicKey()
		validationDistributions := map[string]uint64{
			testVoter.String(): 10000,
		}
		validationDAO.InitialTokenDistribution(validationDistributions)
		testMixedVotingValidation(t, validationDAO, testVoter, testVoter, testVoter)
	})
}

func testSimpleVotingScenario(t *testing.T, dao *DAO, founder, developer1, developer2, community1 crypto.PublicKey) {
	// Create simple majority proposal
	proposalTx := &ProposalTx{
		Fee:          200,
		Title:        "Simple Majority Test",
		Description:  "Testing simple majority voting mechanism",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    time.Now().Unix() - 3600,
		EndTime:      time.Now().Unix() + 86400,
		Threshold:    5100, // 51%
		MetadataHash: randomHash(),
	}

	proposalHash := randomHash()
	err := dao.Processor.ProcessProposalTx(proposalTx, founder, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create simple voting proposal: %v", err)
	}

	// Set proposal as active
	dao.GovernanceState.Proposals[proposalHash].Status = ProposalStatusActive

	// Simulate voting with different weights
	votes := []struct {
		voter  crypto.PublicKey
		choice VoteChoice
		weight uint64
	}{
		{founder, VoteChoiceYes, 2000},       // Uses 2000 tokens
		{developer1, VoteChoiceNo, 1500},     // Uses 1500 tokens
		{developer2, VoteChoiceYes, 1000},    // Uses 1000 tokens
		{community1, VoteChoiceAbstain, 500}, // Uses 500 tokens
	}

	for _, vote := range votes {
		voteTx := &VoteTx{
			Fee:        100,
			ProposalID: proposalHash,
			Choice:     vote.choice,
			Weight:     vote.weight,
			Reason:     "Simple voting test",
		}

		err := dao.Processor.ProcessVoteTx(voteTx, vote.voter)
		if err != nil {
			t.Errorf("Failed to process vote from %s: %v", vote.voter.String()[:8], err)
		}
	}

	// Verify results
	proposal := dao.GovernanceState.Proposals[proposalHash]
	if proposal.Results.YesVotes != 3000 { // 2000 + 1000
		t.Errorf("Expected 3000 yes votes, got %d", proposal.Results.YesVotes)
	}
	if proposal.Results.NoVotes != 1500 {
		t.Errorf("Expected 1500 no votes, got %d", proposal.Results.NoVotes)
	}
	if proposal.Results.AbstainVotes != 500 {
		t.Errorf("Expected 500 abstain votes, got %d", proposal.Results.AbstainVotes)
	}
	if proposal.Results.TotalVoters != 4 {
		t.Errorf("Expected 4 voters, got %d", proposal.Results.TotalVoters)
	}

	t.Logf("✓ Simple voting: Yes=%d, No=%d, Abstain=%d, Voters=%d",
		proposal.Results.YesVotes, proposal.Results.NoVotes,
		proposal.Results.AbstainVotes, proposal.Results.TotalVoters)
}

func testQuadraticVotingScenario(t *testing.T, dao *DAO, founder, developer1, community1, community2 crypto.PublicKey) {
	// Create quadratic voting proposal
	proposalTx := &ProposalTx{
		Fee:          200,
		Title:        "Quadratic Voting Test",
		Description:  "Testing quadratic voting mechanism to prevent plutocracy",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeQuadratic,
		StartTime:    time.Now().Unix() - 3600,
		EndTime:      time.Now().Unix() + 86400,
		Threshold:    5100,
		MetadataHash: randomHash(),
	}

	proposalHash := randomHash()
	err := dao.Processor.ProcessProposalTx(proposalTx, founder, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create quadratic voting proposal: %v", err)
	}

	// Set proposal as active
	dao.GovernanceState.Proposals[proposalHash].Status = ProposalStatusActive

	// Simulate quadratic voting - cost = weight^2
	votes := []struct {
		voter        crypto.PublicKey
		choice       VoteChoice
		weight       uint64
		expectedCost uint64
	}{
		{founder, VoteChoiceYes, 30, 900},        // 30^2 = 900 tokens
		{developer1, VoteChoiceNo, 20, 400},      // 20^2 = 400 tokens
		{community1, VoteChoiceYes, 15, 225},     // 15^2 = 225 tokens
		{community2, VoteChoiceAbstain, 10, 100}, // 10^2 = 100 tokens
	}

	for _, vote := range votes {
		initialBalance := dao.TokenState.Balances[vote.voter.String()]

		voteTx := &VoteTx{
			Fee:        100,
			ProposalID: proposalHash,
			Choice:     vote.choice,
			Weight:     vote.weight,
			Reason:     "Quadratic voting test",
		}

		err := dao.Processor.ProcessVoteTx(voteTx, vote.voter)
		if err != nil {
			t.Errorf("Failed to process quadratic vote from %s: %v", vote.voter.String()[:8], err)
		}

		// Verify cost was deducted correctly
		expectedBalance := initialBalance - vote.expectedCost - 100 // cost + fee
		actualBalance := dao.TokenState.Balances[vote.voter.String()]
		if actualBalance != expectedBalance {
			t.Errorf("Expected balance %d after quadratic vote, got %d", expectedBalance, actualBalance)
		}
	}

	// Verify results
	proposal := dao.GovernanceState.Proposals[proposalHash]
	if proposal.Results.YesVotes != 45 { // 30 + 15
		t.Errorf("Expected 45 yes votes, got %d", proposal.Results.YesVotes)
	}
	if proposal.Results.NoVotes != 20 {
		t.Errorf("Expected 20 no votes, got %d", proposal.Results.NoVotes)
	}
	if proposal.Results.AbstainVotes != 10 {
		t.Errorf("Expected 10 abstain votes, got %d", proposal.Results.AbstainVotes)
	}

	t.Logf("✓ Quadratic voting: Yes=%d, No=%d, Abstain=%d (costs: 900+400+225+100 tokens)",
		proposal.Results.YesVotes, proposal.Results.NoVotes, proposal.Results.AbstainVotes)
}

func testWeightedVotingScenario(t *testing.T, dao *DAO, founder, developer1, developer2, community1, community2 crypto.PublicKey) {
	// Create weighted voting proposal
	proposalTx := &ProposalTx{
		Fee:          200,
		Title:        "Weighted Voting Test",
		Description:  "Testing token-weighted voting mechanism",
		ProposalType: ProposalTypeTreasury,
		VotingType:   VotingTypeWeighted,
		StartTime:    time.Now().Unix() - 3600,
		EndTime:      time.Now().Unix() + 86400,
		Threshold:    5100,
		MetadataHash: randomHash(),
	}

	proposalHash := randomHash()
	err := dao.Processor.ProcessProposalTx(proposalTx, founder, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create weighted voting proposal: %v", err)
	}

	// Set proposal as active
	dao.GovernanceState.Proposals[proposalHash].Status = ProposalStatusActive

	// Simulate weighted voting - each voter uses different proportions of their balance
	votes := []struct {
		voter  crypto.PublicKey
		choice VoteChoice
		weight uint64
	}{
		{founder, VoteChoiceYes, 3000},       // Uses 3000 of remaining tokens
		{developer1, VoteChoiceNo, 2000},     // Uses 2000 tokens
		{developer2, VoteChoiceYes, 1500},    // Uses 1500 tokens
		{community1, VoteChoiceNo, 800},      // Uses 800 tokens
		{community2, VoteChoiceAbstain, 600}, // Uses 600 tokens
	}

	for _, vote := range votes {
		voteTx := &VoteTx{
			Fee:        100,
			ProposalID: proposalHash,
			Choice:     vote.choice,
			Weight:     vote.weight,
			Reason:     "Weighted voting test",
		}

		err := dao.Processor.ProcessVoteTx(voteTx, vote.voter)
		if err != nil {
			t.Errorf("Failed to process weighted vote from %s: %v", vote.voter.String()[:8], err)
		}
	}

	// Verify results
	proposal := dao.GovernanceState.Proposals[proposalHash]
	if proposal.Results.YesVotes != 4500 { // 3000 + 1500
		t.Errorf("Expected 4500 yes votes, got %d", proposal.Results.YesVotes)
	}
	if proposal.Results.NoVotes != 2800 { // 2000 + 800
		t.Errorf("Expected 2800 no votes, got %d", proposal.Results.NoVotes)
	}
	if proposal.Results.AbstainVotes != 600 {
		t.Errorf("Expected 600 abstain votes, got %d", proposal.Results.AbstainVotes)
	}

	t.Logf("✓ Weighted voting: Yes=%d, No=%d, Abstain=%d",
		proposal.Results.YesVotes, proposal.Results.NoVotes, proposal.Results.AbstainVotes)
}

func testReputationVotingScenario(t *testing.T, dao *DAO, founder, developer1, developer2, community1, community2, community3 crypto.PublicKey) {
	// Create reputation-based voting proposal
	proposalTx := &ProposalTx{
		Fee:          200,
		Title:        "Reputation Voting Test",
		Description:  "Testing reputation-based voting for technical decisions",
		ProposalType: ProposalTypeTechnical,
		VotingType:   VotingTypeReputation,
		StartTime:    time.Now().Unix() - 3600,
		EndTime:      time.Now().Unix() + 86400,
		Threshold:    5100,
		MetadataHash: randomHash(),
	}

	proposalHash := randomHash()
	err := dao.Processor.ProcessProposalTx(proposalTx, founder, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create reputation voting proposal: %v", err)
	}

	// Set proposal as active
	dao.GovernanceState.Proposals[proposalHash].Status = ProposalStatusActive

	// Simulate reputation-based voting
	votes := []struct {
		voter  crypto.PublicKey
		choice VoteChoice
		weight uint64
	}{
		{founder, VoteChoiceYes, 4000},       // Uses 4000 of 8000 reputation
		{developer1, VoteChoiceYes, 3000},    // Uses 3000 of 6000 reputation
		{developer2, VoteChoiceNo, 2000},     // Uses 2000 of 4000 reputation
		{community1, VoteChoiceAbstain, 500}, // Uses 500 of 1000 reputation
		{community2, VoteChoiceNo, 400},      // Uses 400 of 800 reputation
		{community3, VoteChoiceYes, 300},     // Uses 300 of 500 reputation
	}

	for _, vote := range votes {
		voteTx := &VoteTx{
			Fee:        50, // Lower fee for reputation voting
			ProposalID: proposalHash,
			Choice:     vote.choice,
			Weight:     vote.weight,
			Reason:     "Reputation voting test",
		}

		err := dao.Processor.ProcessVoteTx(voteTx, vote.voter)
		if err != nil {
			t.Errorf("Failed to process reputation vote from %s: %v", vote.voter.String()[:8], err)
		}
	}

	// Verify results
	proposal := dao.GovernanceState.Proposals[proposalHash]
	if proposal.Results.YesVotes != 7300 { // 4000 + 3000 + 300
		t.Errorf("Expected 7300 yes votes, got %d", proposal.Results.YesVotes)
	}
	if proposal.Results.NoVotes != 2400 { // 2000 + 400
		t.Errorf("Expected 2400 no votes, got %d", proposal.Results.NoVotes)
	}
	if proposal.Results.AbstainVotes != 500 {
		t.Errorf("Expected 500 abstain votes, got %d", proposal.Results.AbstainVotes)
	}

	t.Logf("✓ Reputation voting: Yes=%d, No=%d, Abstain=%d (based on reputation scores)",
		proposal.Results.YesVotes, proposal.Results.NoVotes, proposal.Results.AbstainVotes)
}

func testMixedVotingValidation(t *testing.T, dao *DAO, proposalCreator, voter1, voter2 crypto.PublicKey) {
	// Test various validation scenarios across different voting types

	// Create a simple voting proposal for validation tests
	proposalTx := &ProposalTx{
		Fee:          200,
		Title:        "Validation Test",
		Description:  "Testing various validation scenarios",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    time.Now().Unix() - 3600,
		EndTime:      time.Now().Unix() + 86400,
		Threshold:    5100,
		MetadataHash: randomHash(),
	}

	proposalHash := randomHash()
	err := dao.Processor.ProcessProposalTx(proposalTx, proposalCreator, proposalHash)
	if err != nil {
		t.Fatalf("Failed to create validation test proposal: %v", err)
	}

	// Set proposal as active
	dao.GovernanceState.Proposals[proposalHash].Status = ProposalStatusActive

	// Test 1: Valid vote
	validVote := &VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     VoteChoiceYes,
		Weight:     1000,
		Reason:     "Valid vote test",
	}

	err = dao.Processor.ProcessVoteTx(validVote, proposalCreator)
	if err != nil {
		t.Errorf("Valid vote should succeed: %v", err)
	}

	// Test 2: Double voting prevention
	duplicateVote := &VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     VoteChoiceNo,
		Weight:     500,
		Reason:     "Duplicate vote test",
	}

	err = dao.Processor.ProcessVoteTx(duplicateVote, proposalCreator)
	if err == nil {
		t.Error("Duplicate vote should fail")
	}

	// Test 3: Insufficient balance
	highWeightVote := &VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     VoteChoiceYes,
		Weight:     50000, // More than any user has
		Reason:     "High weight test",
	}

	err = dao.Validator.ValidateVoteTx(highWeightVote, proposalCreator)
	if err == nil {
		t.Error("High weight vote should fail validation")
	}

	// Test 4: Invalid vote choice
	invalidChoiceVote := &VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     VoteChoice(99),
		Weight:     100,
		Reason:     "Invalid choice test",
	}

	err = dao.Validator.ValidateVoteTx(invalidChoiceVote, proposalCreator)
	if err == nil {
		t.Error("Invalid choice vote should fail validation")
	}

	// Test 5: Zero weight vote
	zeroWeightVote := &VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     VoteChoiceYes,
		Weight:     0,
		Reason:     "Zero weight test",
	}

	err = dao.Validator.ValidateVoteTx(zeroWeightVote, proposalCreator)
	if err == nil {
		t.Error("Zero weight vote should fail validation")
	}

	t.Log("✓ Mixed validation tests completed successfully")
}

// TestVotingMechanismPerformance tests the performance of different voting mechanisms
func TestVotingMechanismPerformance(t *testing.T) {
	dao := NewDAO("PXGOV", "ProjectX Governance Token", 18)

	// Setup many voters for performance testing
	numVoters := 100
	voters := make([]crypto.PublicKey, numVoters)
	distributions := make(map[string]uint64)

	for i := 0; i < numVoters; i++ {
		voters[i] = crypto.GeneratePrivateKey().PublicKey()
		distributions[voters[i].String()] = uint64(1000 + i*10) // Varying balances
	}

	err := dao.InitialTokenDistribution(distributions)
	if err != nil {
		t.Fatalf("Failed to distribute tokens: %v", err)
	}

	// Test performance of simple voting with many voters
	proposalTx := &ProposalTx{
		Fee:          200,
		Title:        "Performance Test",
		Description:  "Testing voting performance with many participants",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    time.Now().Unix() - 3600,
		EndTime:      time.Now().Unix() + 86400,
		Threshold:    5100,
		MetadataHash: randomHash(),
	}

	proposalHash := randomHash()
	err = dao.Processor.ProcessProposalTx(proposalTx, voters[0], proposalHash)
	if err != nil {
		t.Fatalf("Failed to create performance test proposal: %v", err)
	}

	// Set proposal as active
	dao.GovernanceState.Proposals[proposalHash].Status = ProposalStatusActive

	// Measure voting performance
	start := time.Now()

	for i := 1; i < numVoters; i++ { // Skip voter[0] who created the proposal
		choice := VoteChoiceYes
		if i%3 == 0 {
			choice = VoteChoiceNo
		} else if i%5 == 0 {
			choice = VoteChoiceAbstain
		}

		voteTx := &VoteTx{
			Fee:        100,
			ProposalID: proposalHash,
			Choice:     choice,
			Weight:     uint64(100 + i),
			Reason:     "Performance test vote",
		}

		err := dao.Processor.ProcessVoteTx(voteTx, voters[i])
		if err != nil {
			t.Errorf("Failed to process vote %d: %v", i, err)
		}
	}

	duration := time.Since(start)

	// Verify all votes were processed
	proposal := dao.GovernanceState.Proposals[proposalHash]
	if proposal.Results.TotalVoters != uint64(numVoters-1) { // -1 for proposal creator
		t.Errorf("Expected %d voters, got %d", numVoters-1, proposal.Results.TotalVoters)
	}

	t.Logf("✓ Performance test: Processed %d votes in %v (%.2f votes/ms)",
		numVoters-1, duration, float64(numVoters-1)/float64(duration.Milliseconds()))
}
