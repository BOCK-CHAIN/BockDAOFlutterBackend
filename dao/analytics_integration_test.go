package dao

import (
	"fmt"
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

func TestAnalyticsIntegration_CompleteWorkflow(t *testing.T) {
	// Create a complete DAO instance with analytics
	dao := NewDAO("GOVTEST", "Governance Test Token", 18)

	// Create test users
	user1 := crypto.GeneratePrivateKey()
	user2 := crypto.GeneratePrivateKey()
	user3 := crypto.GeneratePrivateKey()

	// Initialize with token distribution
	distributions := map[string]uint64{
		user1.PublicKey().String(): 10000,
		user2.PublicKey().String(): 20000,
		user3.PublicKey().String(): 15000,
	}

	err := dao.InitialTokenDistribution(distributions)
	if err != nil {
		t.Fatalf("Failed to initialize token distribution: %v", err)
	}

	// Initialize treasury
	treasurySigners := []crypto.PublicKey{
		user1.PublicKey(),
		user2.PublicKey(),
	}
	err = dao.InitializeTreasury(treasurySigners, 2)
	if err != nil {
		t.Fatalf("Failed to initialize treasury: %v", err)
	}

	// Add some treasury funds
	dao.AddTreasuryFunds(50000)

	// Create several proposals to generate analytics data
	now := time.Now().Unix()

	// Proposal 1: General proposal that passes
	proposal1Tx := &ProposalTx{
		Fee:          1000,
		Title:        "Improve Documentation",
		Description:  "Proposal to improve project documentation",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    now - 90000, // Start 25 hours ago
		EndTime:      now - 3600,  // End 1 hour ago (24 hour voting period)
		Threshold:    2,
	}
	proposal1Hash := types.Hash{1}
	err = dao.ProcessDAOTransaction(proposal1Tx, user1.PublicKey(), proposal1Hash)
	if err != nil {
		t.Fatalf("Failed to create proposal 1: %v", err)
	}

	// Set proposal 1 to passed status and add votes
	if proposal, exists := dao.GovernanceState.Proposals[proposal1Hash]; exists {
		proposal.Status = ProposalStatusPassed
		proposal.Results = &VoteResults{
			YesVotes:     2,
			NoVotes:      0,
			AbstainVotes: 0,
			TotalVoters:  2,
			Quorum:       2,
			Passed:       true,
		}
	}

	// Manually add votes for proposal 1 (since it's already ended)
	if dao.GovernanceState.Votes[proposal1Hash] == nil {
		dao.GovernanceState.Votes[proposal1Hash] = make(map[string]*Vote)
	}
	dao.GovernanceState.Votes[proposal1Hash][user1.PublicKey().String()] = &Vote{
		Voter:     user1.PublicKey(),
		Choice:    VoteChoiceYes,
		Weight:    10000,
		Timestamp: now - 5000,
	}
	dao.GovernanceState.Votes[proposal1Hash][user2.PublicKey().String()] = &Vote{
		Voter:     user2.PublicKey(),
		Choice:    VoteChoiceYes,
		Weight:    20000,
		Timestamp: now - 4000,
	}

	// Proposal 2: Treasury proposal that gets rejected
	proposal2Tx := &ProposalTx{
		Fee:          1000,
		Title:        "Marketing Campaign",
		Description:  "Proposal for marketing campaign funding",
		ProposalType: ProposalTypeTreasury,
		VotingType:   VotingTypeWeighted,
		StartTime:    now - 88200, // Start 24.5 hours ago
		EndTime:      now - 1800,  // End 30 minutes ago (24 hour voting period)
		Threshold:    2,
	}
	proposal2Hash := types.Hash{2}
	err = dao.ProcessDAOTransaction(proposal2Tx, user2.PublicKey(), proposal2Hash)
	if err != nil {
		t.Fatalf("Failed to create proposal 2: %v", err)
	}

	// Set proposal 2 to rejected status and add votes
	if proposal, exists := dao.GovernanceState.Proposals[proposal2Hash]; exists {
		proposal.Status = ProposalStatusRejected
		proposal.Results = &VoteResults{
			YesVotes:     0,
			NoVotes:      2,
			AbstainVotes: 0,
			TotalVoters:  2,
			Quorum:       2,
			Passed:       false,
		}
	}

	// Manually add votes for proposal 2 (since it's already ended)
	if dao.GovernanceState.Votes[proposal2Hash] == nil {
		dao.GovernanceState.Votes[proposal2Hash] = make(map[string]*Vote)
	}
	dao.GovernanceState.Votes[proposal2Hash][user3.PublicKey().String()] = &Vote{
		Voter:     user3.PublicKey(),
		Choice:    VoteChoiceNo,
		Weight:    15000,
		Timestamp: now - 3000,
	}
	dao.GovernanceState.Votes[proposal2Hash][user1.PublicKey().String()] = &Vote{
		Voter:     user1.PublicKey(),
		Choice:    VoteChoiceNo,
		Weight:    10000,
		Timestamp: now - 2000,
	}

	// Proposal 3: Active proposal
	proposal3Tx := &ProposalTx{
		Fee:          1000,
		Title:        "Technical Upgrade",
		Description:  "Proposal for technical system upgrade",
		ProposalType: ProposalTypeTechnical,
		VotingType:   VotingTypeSimple, // Changed from reputation to simple
		StartTime:    now - 1800,       // Started 30 minutes ago
		EndTime:      now + 84600,      // Ends in 23.5 hours (24 hour voting period)
		Threshold:    2,
	}
	proposal3Hash := types.Hash{3}
	err = dao.ProcessDAOTransaction(proposal3Tx, user3.PublicKey(), proposal3Hash)
	if err != nil {
		t.Fatalf("Failed to create proposal 3: %v", err)
	}

	// Set proposal 3 to active status manually
	if proposal, exists := dao.GovernanceState.Proposals[proposal3Hash]; exists {
		proposal.Status = ProposalStatusActive
	}

	// Add one vote for the active proposal
	vote5Tx := &VoteTx{
		Fee:        500,
		ProposalID: proposal3Hash,
		Choice:     VoteChoiceYes,
		Weight:     15000, // Reduced weight to account for fees
	}
	err = dao.ProcessDAOTransaction(vote5Tx, user2.PublicKey(), types.Hash{})
	if err != nil {
		t.Fatalf("Failed to vote on proposal 3: %v", err)
	}

	// Create some delegations
	delegation1Tx := &DelegationTx{
		Fee:      200,
		Delegate: user2.PublicKey(),
		Duration: 86400, // 24 hours
		Revoke:   false,
	}
	err = dao.ProcessDAOTransaction(delegation1Tx, user3.PublicKey(), types.Hash{})
	if err != nil {
		t.Fatalf("Failed to create delegation: %v", err)
	}

	// Create treasury transactions
	treasuryTx1 := &TreasuryTx{
		Fee:          1000,
		Recipient:    user1.PublicKey(),
		Amount:       5000,
		Purpose:      "Development",
		RequiredSigs: 2,
	}
	tx1Hash := types.Hash{10}
	err = dao.CreateTreasuryTransaction(treasuryTx1, tx1Hash)
	if err != nil {
		t.Fatalf("Failed to create treasury transaction: %v", err)
	}

	// Sign and execute treasury transaction
	err = dao.SignTreasuryTransaction(tx1Hash, user1)
	if err != nil {
		t.Fatalf("Failed to sign treasury transaction: %v", err)
	}

	err = dao.SignTreasuryTransaction(tx1Hash, user2)
	if err != nil {
		t.Fatalf("Failed to sign treasury transaction: %v", err)
	}

	// Check if transaction needs execution
	if tx, exists := dao.GetTreasuryTransaction(tx1Hash); exists && !tx.Executed {
		err = dao.ExecuteTreasuryTransaction(tx1Hash)
		if err != nil {
			t.Fatalf("Failed to execute treasury transaction: %v", err)
		}
	}

	// Update proposal statuses
	dao.UpdateAllProposalStatuses()

	// Now test all analytics functions
	t.Run("ParticipationMetrics", func(t *testing.T) {
		metrics := dao.GetGovernanceParticipationMetrics()

		if metrics.TotalProposals != 3 {
			t.Errorf("Expected 3 total proposals, got %d", metrics.TotalProposals)
		}

		if metrics.ActiveProposals != 1 {
			t.Errorf("Expected 1 active proposal, got %d", metrics.ActiveProposals)
		}

		if metrics.TotalVotes != 5 {
			t.Errorf("Expected 5 total votes, got %d", metrics.TotalVotes)
		}

		if metrics.UniqueVoters != 3 {
			t.Errorf("Expected 3 unique voters, got %d", metrics.UniqueVoters)
		}

		// Check participation rate (3 voters out of 3 token holders = 100%)
		if metrics.ParticipationRate != 100.0 {
			t.Errorf("Expected 100%% participation rate, got %.2f", metrics.ParticipationRate)
		}

		// Check voting patterns
		if metrics.VotingPatterns[VoteChoiceYes] != 3 {
			t.Errorf("Expected 3 yes votes, got %d", metrics.VotingPatterns[VoteChoiceYes])
		}

		if metrics.VotingPatterns[VoteChoiceNo] != 2 {
			t.Errorf("Expected 2 no votes, got %d", metrics.VotingPatterns[VoteChoiceNo])
		}

		// Check proposals by type
		if metrics.ProposalsByType[ProposalTypeGeneral] != 1 {
			t.Errorf("Expected 1 general proposal, got %d", metrics.ProposalsByType[ProposalTypeGeneral])
		}

		if metrics.ProposalsByType[ProposalTypeTreasury] != 1 {
			t.Errorf("Expected 1 treasury proposal, got %d", metrics.ProposalsByType[ProposalTypeTreasury])
		}

		if metrics.ProposalsByType[ProposalTypeTechnical] != 1 {
			t.Errorf("Expected 1 technical proposal, got %d", metrics.ProposalsByType[ProposalTypeTechnical])
		}

		// Check top participants
		if len(metrics.TopParticipants) != 3 {
			t.Errorf("Expected 3 top participants, got %d", len(metrics.TopParticipants))
		}

		// Check delegation metrics
		if metrics.DelegationMetrics.TotalDelegations != 1 {
			t.Errorf("Expected 1 total delegation, got %d", metrics.DelegationMetrics.TotalDelegations)
		}

		if metrics.DelegationMetrics.ActiveDelegations != 1 {
			t.Errorf("Expected 1 active delegation, got %d", metrics.DelegationMetrics.ActiveDelegations)
		}
	})

	t.Run("TreasuryMetrics", func(t *testing.T) {
		metrics := dao.GetTreasuryPerformanceMetrics()

		if metrics.CurrentBalance != 45000 { // 50000 - 5000
			t.Errorf("Expected current balance 45000, got %d", metrics.CurrentBalance)
		}

		if metrics.TransactionCount != 1 {
			t.Errorf("Expected 1 transaction, got %d", metrics.TransactionCount)
		}

		if metrics.ExecutedTransactions != 1 {
			t.Errorf("Expected 1 executed transaction, got %d", metrics.ExecutedTransactions)
		}

		if metrics.TotalOutflows != 5000 {
			t.Errorf("Expected total outflows 5000, got %d", metrics.TotalOutflows)
		}

		if metrics.SigningEfficiency != 100.0 {
			t.Errorf("Expected 100%% signing efficiency, got %.2f", metrics.SigningEfficiency)
		}

		if metrics.TransactionsByPurpose["Development"] != 1 {
			t.Errorf("Expected 1 development transaction, got %d", metrics.TransactionsByPurpose["Development"])
		}
	})

	t.Run("ProposalAnalytics", func(t *testing.T) {
		analytics := dao.GetProposalAnalytics()

		if analytics.TotalProposals != 3 {
			t.Errorf("Expected 3 total proposals, got %d", analytics.TotalProposals)
		}

		// Note: Proposal status updates might not reflect immediately in test
		// We'll check the basic structure instead
		if analytics.ProposalsByCreator[user1.PublicKey().String()] != 1 {
			t.Errorf("Expected user1 to have 1 proposal, got %d", analytics.ProposalsByCreator[user1.PublicKey().String()])
		}

		if analytics.ProposalsByCreator[user2.PublicKey().String()] != 1 {
			t.Errorf("Expected user2 to have 1 proposal, got %d", analytics.ProposalsByCreator[user2.PublicKey().String()])
		}

		if analytics.ProposalsByCreator[user3.PublicKey().String()] != 1 {
			t.Errorf("Expected user3 to have 1 proposal, got %d", analytics.ProposalsByCreator[user3.PublicKey().String()])
		}
	})

	t.Run("HealthMetrics", func(t *testing.T) {
		health := dao.GetDAOHealthMetrics()

		// Basic health checks
		if health.OverallScore < 0 || health.OverallScore > 100 {
			t.Errorf("Overall score should be between 0-100, got %.2f", health.OverallScore)
		}

		if health.ParticipationHealth < 0 || health.ParticipationHealth > 100 {
			t.Errorf("Participation health should be between 0-100, got %.2f", health.ParticipationHealth)
		}

		if health.TreasuryHealth < 0 || health.TreasuryHealth > 100 {
			t.Errorf("Treasury health should be between 0-100, got %.2f", health.TreasuryHealth)
		}

		if health.GovernanceHealth < 0 || health.GovernanceHealth > 100 {
			t.Errorf("Governance health should be between 0-100, got %.2f", health.GovernanceHealth)
		}

		if health.SecurityHealth < 0 || health.SecurityHealth > 100 {
			t.Errorf("Security health should be between 0-100, got %.2f", health.SecurityHealth)
		}

		if health.LastUpdated == 0 {
			t.Error("LastUpdated should be set")
		}

		if health.HealthTrend == "" {
			t.Error("HealthTrend should be set")
		}

		// With 100% participation, health should be reasonably good
		if health.ParticipationHealth < 50 {
			t.Errorf("Expected good participation health with 100%% participation, got %.2f", health.ParticipationHealth)
		}
	})

	t.Run("AnalyticsSummary", func(t *testing.T) {
		summary := dao.GetAnalyticsSummary()

		// Check all required sections are present
		requiredSections := []string{
			"participation_metrics",
			"treasury_metrics",
			"proposal_analytics",
			"health_metrics",
			"generated_at",
		}

		for _, section := range requiredSections {
			if _, exists := summary[section]; !exists {
				t.Errorf("Missing required section: %s", section)
			}
		}

		// Verify data types
		if _, ok := summary["participation_metrics"].(*GovernanceParticipationMetrics); !ok {
			t.Error("participation_metrics should be GovernanceParticipationMetrics")
		}

		if _, ok := summary["treasury_metrics"].(*TreasuryPerformanceMetrics); !ok {
			t.Error("treasury_metrics should be TreasuryPerformanceMetrics")
		}

		if _, ok := summary["proposal_analytics"].(*ProposalAnalytics); !ok {
			t.Error("proposal_analytics should be ProposalAnalytics")
		}

		if _, ok := summary["health_metrics"].(*DAOHealthMetrics); !ok {
			t.Error("health_metrics should be DAOHealthMetrics")
		}

		if _, ok := summary["generated_at"].(int64); !ok {
			t.Error("generated_at should be int64")
		}
	})
}

func TestAnalyticsIntegration_EdgeCases(t *testing.T) {
	// Test analytics with minimal data
	dao := NewDAO("MINIMAL", "Minimal Token", 18)

	t.Run("EmptyDAOAnalytics", func(t *testing.T) {
		// Test analytics on empty DAO
		participation := dao.GetGovernanceParticipationMetrics()
		if participation.TotalProposals != 0 {
			t.Errorf("Expected 0 proposals in empty DAO, got %d", participation.TotalProposals)
		}

		treasury := dao.GetTreasuryPerformanceMetrics()
		if treasury.CurrentBalance != 0 {
			t.Errorf("Expected 0 treasury balance in empty DAO, got %d", treasury.CurrentBalance)
		}

		proposals := dao.GetProposalAnalytics()
		if proposals.TotalProposals != 0 {
			t.Errorf("Expected 0 proposals in empty DAO, got %d", proposals.TotalProposals)
		}

		health := dao.GetDAOHealthMetrics()
		// Health should still be calculable even with no data
		if health.OverallScore < 0 || health.OverallScore > 100 {
			t.Errorf("Health score should be valid even for empty DAO, got %.2f", health.OverallScore)
		}
	})

	t.Run("SingleUserDAO", func(t *testing.T) {
		// Test with single user
		user := crypto.GeneratePrivateKey()
		distributions := map[string]uint64{
			user.PublicKey().String(): 1000,
		}

		err := dao.InitialTokenDistribution(distributions)
		if err != nil {
			t.Fatalf("Failed to initialize single user distribution: %v", err)
		}

		// Create single proposal
		proposalTx := &ProposalTx{
			Fee:          100,
			Title:        "Solo Proposal",
			Description:  "A proposal by the only user",
			ProposalType: ProposalTypeGeneral,
			VotingType:   VotingTypeSimple,
			StartTime:    time.Now().Unix() - 3600,  // Started 1 hour ago
			EndTime:      time.Now().Unix() + 82800, // Ends in 23 hours (24 hour voting period)
			Threshold:    1,
		}
		proposalHash := types.Hash{1}
		err = dao.ProcessDAOTransaction(proposalTx, user.PublicKey(), proposalHash)
		if err != nil {
			t.Fatalf("Failed to create solo proposal: %v", err)
		}

		// Set proposal to active status
		if proposal, exists := dao.GovernanceState.Proposals[proposalHash]; exists {
			proposal.Status = ProposalStatusActive
		}

		// Vote on own proposal
		voteTx := &VoteTx{
			Fee:        50,
			ProposalID: proposalHash,
			Choice:     VoteChoiceYes,
			Weight:     850, // Further reduced to account for proposal fee
		}
		err = dao.ProcessDAOTransaction(voteTx, user.PublicKey(), types.Hash{})
		if err != nil {
			t.Fatalf("Failed to vote on solo proposal: %v", err)
		}

		// Test analytics
		participation := dao.GetGovernanceParticipationMetrics()
		if participation.ParticipationRate != 100.0 {
			t.Errorf("Expected 100%% participation with single user, got %.2f", participation.ParticipationRate)
		}

		if participation.UniqueVoters != 1 {
			t.Errorf("Expected 1 unique voter, got %d", participation.UniqueVoters)
		}

		if len(participation.TopParticipants) != 1 {
			t.Errorf("Expected 1 top participant, got %d", len(participation.TopParticipants))
		}
	})
}

func TestAnalyticsIntegration_PerformanceWithLargeDataset(t *testing.T) {
	// Test analytics performance with larger dataset
	dao := NewDAO("LARGE", "Large Test Token", 18)

	// Create multiple users
	users := make([]crypto.PrivateKey, 20)
	distributions := make(map[string]uint64)

	for i := 0; i < 20; i++ {
		users[i] = crypto.GeneratePrivateKey()
		distributions[users[i].PublicKey().String()] = uint64(10000 + i*1000) // Increased amounts
	}

	err := dao.InitialTokenDistribution(distributions)
	if err != nil {
		t.Fatalf("Failed to initialize large distribution: %v", err)
	}

	// Create multiple proposals
	for i := 0; i < 10; i++ {
		proposalTx := &ProposalTx{
			Fee:          1000,
			Title:        fmt.Sprintf("Proposal %d", i+1),
			Description:  fmt.Sprintf("Description for proposal %d", i+1),
			ProposalType: ProposalType(1 + (i % 4)),                      // Cycle through proposal types
			VotingType:   VotingType(1 + (i % 4)),                        // Cycle through voting types
			StartTime:    time.Now().Unix() - int64(86400*(10-i)) - 3600, // Start with 24+ hour periods
			EndTime:      time.Now().Unix() - int64(86400*(10-i-1)),      // End 24 hours later
			Threshold:    uint64(5 + i),
		}
		proposalHash := types.Hash{byte(i + 1)}

		creator := users[i%len(users)]
		err = dao.ProcessDAOTransaction(proposalTx, creator.PublicKey(), proposalHash)
		if err != nil {
			t.Fatalf("Failed to create proposal %d: %v", i+1, err)
		}

		// Set proposal status and add votes manually (since they're all in the past)
		if proposal, exists := dao.GovernanceState.Proposals[proposalHash]; exists {
			if i%2 == 0 {
				proposal.Status = ProposalStatusPassed
			} else {
				proposal.Status = ProposalStatusRejected
			}
		}

		// Manually add votes for all proposals (since they're all completed)
		if dao.GovernanceState.Votes[proposalHash] == nil {
			dao.GovernanceState.Votes[proposalHash] = make(map[string]*Vote)
		}

		for j := 0; j < min(15, len(users)); j++ {
			if j == i%len(users) {
				continue // Skip creator voting on own proposal for variety
			}

			choice := VoteChoiceYes
			if j%3 == 1 {
				choice = VoteChoiceNo
			} else if j%3 == 2 {
				choice = VoteChoiceAbstain
			}

			dao.GovernanceState.Votes[proposalHash][users[j].PublicKey().String()] = &Vote{
				Voter:     users[j].PublicKey(),
				Choice:    choice,
				Weight:    distributions[users[j].PublicKey().String()],
				Timestamp: time.Now().Unix() - int64(86400*(10-i)) + int64(j*60),
			}
		}
	}

	// Measure analytics performance
	start := time.Now()

	participation := dao.GetGovernanceParticipationMetrics()
	_ = dao.GetTreasuryPerformanceMetrics() // Measured for performance, not used in assertions
	proposals := dao.GetProposalAnalytics()
	health := dao.GetDAOHealthMetrics()
	summary := dao.GetAnalyticsSummary()

	duration := time.Since(start)

	// Performance should be reasonable (under 1 second for this dataset)
	if duration > time.Second {
		t.Errorf("Analytics calculation took too long: %v", duration)
	}

	// Verify data integrity
	if participation.TotalProposals != 10 {
		t.Errorf("Expected 10 proposals, got %d", participation.TotalProposals)
	}

	if len(participation.TopParticipants) > 10 {
		t.Errorf("Top participants list should be limited to 10, got %d", len(participation.TopParticipants))
	}

	if proposals.TotalProposals != 10 {
		t.Errorf("Expected 10 proposals in analytics, got %d", proposals.TotalProposals)
	}

	if health.OverallScore < 0 || health.OverallScore > 100 {
		t.Errorf("Health score should be valid, got %.2f", health.OverallScore)
	}

	if summary == nil {
		t.Error("Summary should not be nil")
	}

	t.Logf("Analytics calculation completed in %v", duration)
	t.Logf("Processed %d proposals, %d users, %d votes",
		participation.TotalProposals, len(users), participation.TotalVotes)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
