package dao

import (
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

func TestAnalyticsSystem_GetGovernanceParticipationMetrics(t *testing.T) {
	// Create test data
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	analytics := NewAnalyticsSystem(governanceState, tokenState)

	// Create test users
	user1 := crypto.GeneratePrivateKey().PublicKey()
	user2 := crypto.GeneratePrivateKey().PublicKey()
	user3 := crypto.GeneratePrivateKey().PublicKey()

	// Add token holders
	governanceState.TokenHolders[user1.String()] = &TokenHolder{
		Address:    user1,
		Balance:    1000,
		Reputation: 100,
		JoinedAt:   time.Now().Unix() - 86400,
	}
	governanceState.TokenHolders[user2.String()] = &TokenHolder{
		Address:    user2,
		Balance:    2000,
		Reputation: 200,
		JoinedAt:   time.Now().Unix() - 172800,
	}
	governanceState.TokenHolders[user3.String()] = &TokenHolder{
		Address:    user3,
		Balance:    500,
		Reputation: 50,
		JoinedAt:   time.Now().Unix() - 259200,
	}

	// Create test proposals
	proposal1ID := types.Hash{1}
	proposal2ID := types.Hash{2}

	governanceState.Proposals[proposal1ID] = &Proposal{
		ID:           proposal1ID,
		Creator:      user1,
		Title:        "Test Proposal 1",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		Status:       ProposalStatusPassed,
		StartTime:    time.Now().Unix() - 3600,
		EndTime:      time.Now().Unix(),
	}

	governanceState.Proposals[proposal2ID] = &Proposal{
		ID:           proposal2ID,
		Creator:      user2,
		Title:        "Test Proposal 2",
		ProposalType: ProposalTypeTreasury,
		VotingType:   VotingTypeQuadratic,
		Status:       ProposalStatusActive,
		StartTime:    time.Now().Unix() - 1800,
		EndTime:      time.Now().Unix() + 1800,
	}

	// Add votes
	governanceState.Votes[proposal1ID] = map[string]*Vote{
		user1.String(): {
			Voter:     user1,
			Choice:    VoteChoiceYes,
			Weight:    1000,
			Timestamp: time.Now().Unix() - 1800,
		},
		user2.String(): {
			Voter:     user2,
			Choice:    VoteChoiceNo,
			Weight:    2000,
			Timestamp: time.Now().Unix() - 1200,
		},
	}

	governanceState.Votes[proposal2ID] = map[string]*Vote{
		user1.String(): {
			Voter:     user1,
			Choice:    VoteChoiceYes,
			Weight:    1000,
			Timestamp: time.Now().Unix() - 600,
		},
	}

	// Get metrics
	metrics := analytics.GetGovernanceParticipationMetrics()

	// Verify results
	if metrics.TotalProposals != 2 {
		t.Errorf("Expected 2 total proposals, got %d", metrics.TotalProposals)
	}

	if metrics.ActiveProposals != 1 {
		t.Errorf("Expected 1 active proposal, got %d", metrics.ActiveProposals)
	}

	if metrics.TotalVotes != 3 {
		t.Errorf("Expected 3 total votes, got %d", metrics.TotalVotes)
	}

	if metrics.UniqueVoters != 2 {
		t.Errorf("Expected 2 unique voters, got %d", metrics.UniqueVoters)
	}

	expectedParticipationRate := float64(2) / float64(3) * 100 // 2 voters out of 3 token holders
	if metrics.ParticipationRate != expectedParticipationRate {
		t.Errorf("Expected participation rate %.2f, got %.2f", expectedParticipationRate, metrics.ParticipationRate)
	}

	if metrics.VotingPatterns[VoteChoiceYes] != 2 {
		t.Errorf("Expected 2 yes votes, got %d", metrics.VotingPatterns[VoteChoiceYes])
	}

	if metrics.VotingPatterns[VoteChoiceNo] != 1 {
		t.Errorf("Expected 1 no vote, got %d", metrics.VotingPatterns[VoteChoiceNo])
	}

	if metrics.ProposalsByType[ProposalTypeGeneral] != 1 {
		t.Errorf("Expected 1 general proposal, got %d", metrics.ProposalsByType[ProposalTypeGeneral])
	}

	if metrics.ProposalsByType[ProposalTypeTreasury] != 1 {
		t.Errorf("Expected 1 treasury proposal, got %d", metrics.ProposalsByType[ProposalTypeTreasury])
	}

	// Check top participants
	if len(metrics.TopParticipants) != 2 {
		t.Errorf("Expected 2 top participants, got %d", len(metrics.TopParticipants))
	}

	// User1 should be first (2 votes)
	if metrics.TopParticipants[0].Address != user1.String() {
		t.Errorf("Expected user1 to be top participant, got %s", metrics.TopParticipants[0].Address)
	}

	if metrics.TopParticipants[0].VotesCast != 2 {
		t.Errorf("Expected user1 to have 2 votes cast, got %d", metrics.TopParticipants[0].VotesCast)
	}
}

func TestAnalyticsSystem_GetTreasuryPerformanceMetrics(t *testing.T) {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	analytics := NewAnalyticsSystem(governanceState, tokenState)

	// Set treasury balance
	governanceState.Treasury.Balance = 10000

	// Create test treasury transactions
	tx1ID := types.Hash{1}
	tx2ID := types.Hash{2}
	tx3ID := types.Hash{3}

	now := time.Now().Unix()

	governanceState.Treasury.Transactions[tx1ID] = &PendingTx{
		ID:        tx1ID,
		Amount:    1000,
		Purpose:   "Development",
		CreatedAt: now - 3600,
		Executed:  true,
	}

	governanceState.Treasury.Transactions[tx2ID] = &PendingTx{
		ID:        tx2ID,
		Amount:    2000,
		Purpose:   "Marketing",
		CreatedAt: now - 1800,
		Executed:  true,
	}

	governanceState.Treasury.Transactions[tx3ID] = &PendingTx{
		ID:        tx3ID,
		Amount:    500,
		Purpose:   "Operations",
		CreatedAt: now - 900,
		ExpiresAt: now - 100, // Expired
		Executed:  false,
	}

	// Get metrics
	metrics := analytics.GetTreasuryPerformanceMetrics()

	// Verify results
	if metrics.CurrentBalance != 10000 {
		t.Errorf("Expected current balance 10000, got %d", metrics.CurrentBalance)
	}

	if metrics.TransactionCount != 3 {
		t.Errorf("Expected 3 transactions, got %d", metrics.TransactionCount)
	}

	if metrics.ExecutedTransactions != 2 {
		t.Errorf("Expected 2 executed transactions, got %d", metrics.ExecutedTransactions)
	}

	if metrics.ExpiredTransactions != 1 {
		t.Errorf("Expected 1 expired transaction, got %d", metrics.ExpiredTransactions)
	}

	if metrics.TotalOutflows != 3000 {
		t.Errorf("Expected total outflows 3000, got %d", metrics.TotalOutflows)
	}

	if metrics.AverageTransactionSize != 1500 {
		t.Errorf("Expected average transaction size 1500, got %d", metrics.AverageTransactionSize)
	}

	if metrics.LargestTransaction != 2000 {
		t.Errorf("Expected largest transaction 2000, got %d", metrics.LargestTransaction)
	}

	if metrics.SmallestTransaction != 1000 {
		t.Errorf("Expected smallest transaction 1000, got %d", metrics.SmallestTransaction)
	}

	expectedSigningEfficiency := float64(2) / float64(3) * 100
	if metrics.SigningEfficiency != expectedSigningEfficiency {
		t.Errorf("Expected signing efficiency %.2f, got %.2f", expectedSigningEfficiency, metrics.SigningEfficiency)
	}

	if metrics.TransactionsByPurpose["Development"] != 1 {
		t.Errorf("Expected 1 development transaction, got %d", metrics.TransactionsByPurpose["Development"])
	}

	if metrics.TransactionsByPurpose["Marketing"] != 1 {
		t.Errorf("Expected 1 marketing transaction, got %d", metrics.TransactionsByPurpose["Marketing"])
	}
}

func TestAnalyticsSystem_GetProposalAnalytics(t *testing.T) {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	analytics := NewAnalyticsSystem(governanceState, tokenState)

	// Create test users
	user1 := crypto.GeneratePrivateKey().PublicKey()
	user2 := crypto.GeneratePrivateKey().PublicKey()

	// Add token holders
	governanceState.TokenHolders[user1.String()] = &TokenHolder{Address: user1, Balance: 1000}
	governanceState.TokenHolders[user2.String()] = &TokenHolder{Address: user2, Balance: 2000}

	// Create test proposals
	proposal1ID := types.Hash{1}
	proposal2ID := types.Hash{2}
	proposal3ID := types.Hash{3}

	now := time.Now().Unix()

	governanceState.Proposals[proposal1ID] = &Proposal{
		ID:           proposal1ID,
		Creator:      user1,
		ProposalType: ProposalTypeGeneral,
		Status:       ProposalStatusPassed,
		StartTime:    now - 7200,
		EndTime:      now - 3600,
		Results:      &VoteResults{Quorum: 1, Passed: true},
	}

	governanceState.Proposals[proposal2ID] = &Proposal{
		ID:           proposal2ID,
		Creator:      user2,
		ProposalType: ProposalTypeGeneral,
		Status:       ProposalStatusRejected,
		StartTime:    now - 5400,
		EndTime:      now - 1800,
		Results:      &VoteResults{Quorum: 1, Passed: false},
	}

	governanceState.Proposals[proposal3ID] = &Proposal{
		ID:           proposal3ID,
		Creator:      user1,
		ProposalType: ProposalTypeTreasury,
		Status:       ProposalStatusActive,
		StartTime:    now - 1800,
		EndTime:      now + 1800,
	}

	// Add votes
	governanceState.Votes[proposal1ID] = map[string]*Vote{
		user1.String(): {Choice: VoteChoiceYes},
		user2.String(): {Choice: VoteChoiceYes},
	}

	governanceState.Votes[proposal2ID] = map[string]*Vote{
		user1.String(): {Choice: VoteChoiceNo},
	}

	// Get analytics
	analytics_result := analytics.GetProposalAnalytics()

	// Verify results
	if analytics_result.TotalProposals != 3 {
		t.Errorf("Expected 3 total proposals, got %d", analytics_result.TotalProposals)
	}

	if analytics_result.PassedProposals != 1 {
		t.Errorf("Expected 1 passed proposal, got %d", analytics_result.PassedProposals)
	}

	if analytics_result.RejectedProposals != 1 {
		t.Errorf("Expected 1 rejected proposal, got %d", analytics_result.RejectedProposals)
	}

	if analytics_result.PendingProposals != 1 {
		t.Errorf("Expected 1 pending proposal, got %d", analytics_result.PendingProposals)
	}

	expectedSuccessRate := float64(1) / float64(3) * 100
	if analytics_result.SuccessRate != expectedSuccessRate {
		t.Errorf("Expected success rate %.2f, got %.2f", expectedSuccessRate, analytics_result.SuccessRate)
	}

	if analytics_result.ProposalsByCreator[user1.String()] != 2 {
		t.Errorf("Expected user1 to have 2 proposals, got %d", analytics_result.ProposalsByCreator[user1.String()])
	}

	if analytics_result.ProposalsByCreator[user2.String()] != 1 {
		t.Errorf("Expected user2 to have 1 proposal, got %d", analytics_result.ProposalsByCreator[user2.String()])
	}

	// Check success rate by type
	expectedGeneralSuccessRate := float64(1) / float64(2) * 100 // 1 passed out of 2 general proposals
	if analytics_result.SuccessRateByType[ProposalTypeGeneral] != expectedGeneralSuccessRate {
		t.Errorf("Expected general proposal success rate %.2f, got %.2f",
			expectedGeneralSuccessRate, analytics_result.SuccessRateByType[ProposalTypeGeneral])
	}

	// Treasury proposals: 0 passed out of 1 (still active)
	if analytics_result.SuccessRateByType[ProposalTypeTreasury] != 0 {
		t.Errorf("Expected treasury proposal success rate 0, got %.2f",
			analytics_result.SuccessRateByType[ProposalTypeTreasury])
	}
}

func TestAnalyticsSystem_GetDAOHealthMetrics(t *testing.T) {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	analytics := NewAnalyticsSystem(governanceState, tokenState)

	// Create a healthy DAO scenario
	user1 := crypto.GeneratePrivateKey().PublicKey()
	user2 := crypto.GeneratePrivateKey().PublicKey()
	user3 := crypto.GeneratePrivateKey().PublicKey()

	// Add token holders (good participation base)
	governanceState.TokenHolders[user1.String()] = &TokenHolder{Address: user1, Balance: 1000}
	governanceState.TokenHolders[user2.String()] = &TokenHolder{Address: user2, Balance: 2000}
	governanceState.TokenHolders[user3.String()] = &TokenHolder{Address: user3, Balance: 1500}

	// Create proposals with good success rate
	for i := 0; i < 10; i++ {
		proposalID := types.Hash{byte(i)}
		status := ProposalStatusPassed
		if i > 6 { // 70% success rate
			status = ProposalStatusRejected
		}

		governanceState.Proposals[proposalID] = &Proposal{
			ID:           proposalID,
			Creator:      user1,
			ProposalType: ProposalTypeGeneral,
			Status:       status,
			Results:      &VoteResults{Quorum: 2, Passed: status == ProposalStatusPassed},
		}

		// Add votes for good participation
		governanceState.Votes[proposalID] = map[string]*Vote{
			user1.String(): {Choice: VoteChoiceYes},
			user2.String(): {Choice: VoteChoiceYes},
		}
	}

	// Set up healthy treasury
	governanceState.Treasury.Balance = 50000
	for i := 0; i < 5; i++ {
		txID := types.Hash{byte(i + 10)}
		governanceState.Treasury.Transactions[txID] = &PendingTx{
			ID:       txID,
			Amount:   1000,
			Executed: true,
		}
	}

	// Get health metrics
	health := analytics.GetDAOHealthMetrics()

	// Verify overall health is reasonable
	if health.OverallScore < 50 {
		t.Errorf("Expected overall score > 50 for healthy DAO, got %.2f", health.OverallScore)
	}

	if health.ParticipationHealth < 30 { // 2/3 participation = 66.67%
		t.Errorf("Expected participation health > 30, got %.2f", health.ParticipationHealth)
	}

	if health.TreasuryHealth < 50 {
		t.Errorf("Expected treasury health > 50, got %.2f", health.TreasuryHealth)
	}

	if health.GovernanceHealth < 50 {
		t.Errorf("Expected governance health > 50, got %.2f", health.GovernanceHealth)
	}

	if health.LastUpdated == 0 {
		t.Error("Expected LastUpdated to be set")
	}

	if health.HealthTrend == "" {
		t.Error("Expected HealthTrend to be set")
	}
}

func TestAnalyticsSystem_GetAnalyticsSummary(t *testing.T) {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	analytics := NewAnalyticsSystem(governanceState, tokenState)

	// Get summary
	summary := analytics.GetAnalyticsSummary()

	// Verify all sections are present
	if _, exists := summary["participation_metrics"]; !exists {
		t.Error("Expected participation_metrics in summary")
	}

	if _, exists := summary["treasury_metrics"]; !exists {
		t.Error("Expected treasury_metrics in summary")
	}

	if _, exists := summary["proposal_analytics"]; !exists {
		t.Error("Expected proposal_analytics in summary")
	}

	if _, exists := summary["health_metrics"]; !exists {
		t.Error("Expected health_metrics in summary")
	}

	if _, exists := summary["generated_at"]; !exists {
		t.Error("Expected generated_at in summary")
	}

	// Verify types
	if _, ok := summary["participation_metrics"].(*GovernanceParticipationMetrics); !ok {
		t.Error("Expected participation_metrics to be GovernanceParticipationMetrics")
	}

	if _, ok := summary["treasury_metrics"].(*TreasuryPerformanceMetrics); !ok {
		t.Error("Expected treasury_metrics to be TreasuryPerformanceMetrics")
	}

	if _, ok := summary["proposal_analytics"].(*ProposalAnalytics); !ok {
		t.Error("Expected proposal_analytics to be ProposalAnalytics")
	}

	if _, ok := summary["health_metrics"].(*DAOHealthMetrics); !ok {
		t.Error("Expected health_metrics to be DAOHealthMetrics")
	}
}

func TestDelegationAnalytics(t *testing.T) {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken("TEST", "Test Token", 18)
	analytics := NewAnalyticsSystem(governanceState, tokenState)

	// Create test users
	delegator1 := crypto.GeneratePrivateKey().PublicKey()
	delegator2 := crypto.GeneratePrivateKey().PublicKey()
	delegate1 := crypto.GeneratePrivateKey().PublicKey()

	now := time.Now().Unix()

	// Add token holders
	governanceState.TokenHolders[delegator1.String()] = &TokenHolder{Address: delegator1, Balance: 1000}
	governanceState.TokenHolders[delegator2.String()] = &TokenHolder{Address: delegator2, Balance: 2000}
	governanceState.TokenHolders[delegate1.String()] = &TokenHolder{Address: delegate1, Balance: 1500}

	// Create active delegations
	governanceState.Delegations[delegator1.String()] = &Delegation{
		Delegator: delegator1,
		Delegate:  delegate1,
		StartTime: now - 3600,
		EndTime:   now + 3600,
		Active:    true,
	}

	governanceState.Delegations[delegator2.String()] = &Delegation{
		Delegator: delegator2,
		Delegate:  delegate1,
		StartTime: now - 1800,
		EndTime:   now + 1800,
		Active:    true,
	}

	// Get delegation analytics
	delegationAnalytics := analytics.getDelegationAnalytics()

	// Verify results
	if delegationAnalytics.TotalDelegations != 2 {
		t.Errorf("Expected 2 total delegations, got %d", delegationAnalytics.TotalDelegations)
	}

	if delegationAnalytics.ActiveDelegations != 2 {
		t.Errorf("Expected 2 active delegations, got %d", delegationAnalytics.ActiveDelegations)
	}

	expectedDelegationRate := float64(2) / float64(3) * 100 // 2 delegations out of 3 token holders
	if delegationAnalytics.DelegationRate != expectedDelegationRate {
		t.Errorf("Expected delegation rate %.2f, got %.2f", expectedDelegationRate, delegationAnalytics.DelegationRate)
	}

	if len(delegationAnalytics.TopDelegates) != 1 {
		t.Errorf("Expected 1 top delegate, got %d", len(delegationAnalytics.TopDelegates))
	}

	if delegationAnalytics.TopDelegates[0].Address != delegate1.String() {
		t.Errorf("Expected delegate1 to be top delegate, got %s", delegationAnalytics.TopDelegates[0].Address)
	}

	if delegationAnalytics.TopDelegates[0].DelegatorsCount != 2 {
		t.Errorf("Expected delegate1 to have 2 delegators, got %d", delegationAnalytics.TopDelegates[0].DelegatorsCount)
	}

	if delegationAnalytics.DelegationDistribution[delegate1.String()] != 2 {
		t.Errorf("Expected delegate1 to have 2 in distribution, got %d", delegationAnalytics.DelegationDistribution[delegate1.String()])
	}
}
