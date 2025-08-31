package dao

import (
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
)

// TestProposalManagementIntegration tests the complete proposal management workflow
func TestProposalManagementIntegration(t *testing.T) {
	// Create DAO with enhanced proposal management
	dao := NewDAO("PXGOV", "ProjectX Governance Token", 18)

	// Setup participants
	founder := crypto.GeneratePrivateKey().PublicKey()
	developer := crypto.GeneratePrivateKey().PublicKey()
	community1 := crypto.GeneratePrivateKey().PublicKey()
	community2 := crypto.GeneratePrivateKey().PublicKey()

	// Initial token distribution
	distributions := map[string]uint64{
		founder.String():    15000,
		developer.String():  10000,
		community1.String(): 8000,
		community2.String(): 7000,
	}

	err := dao.InitialTokenDistribution(distributions)
	if err != nil {
		t.Fatalf("Failed to distribute tokens: %v", err)
	}

	// Initialize treasury
	treasurySigners := []crypto.PublicKey{founder, developer}
	err = dao.InitializeTreasury(treasurySigners, 2)
	if err != nil {
		t.Fatalf("Failed to initialize treasury: %v", err)
	}
	dao.AddTreasuryFunds(100000)

	pm := dao.ProposalManager

	// Test 1: Create different types of proposals
	t.Run("CreateDifferentProposalTypes", func(t *testing.T) {
		now := time.Now().Unix()

		// General proposal
		generalTx := &ProposalTx{
			Fee:          100,
			Title:        "Community Guidelines Update",
			Description:  "Update community guidelines for better governance",
			ProposalType: ProposalTypeGeneral,
			VotingType:   VotingTypeSimple,
			StartTime:    now + 100,
			EndTime:      now + 100 + 86400,
			Threshold:    5100,
		}

		generalHash := randomHash()
		_, err := pm.CreateProposal(generalTx, founder, generalHash)
		if err != nil {
			t.Fatalf("Failed to create general proposal: %v", err)
		}

		// Treasury proposal
		treasuryTx := &ProposalTx{
			Fee:          200,
			Title:        "Development Fund Allocation",
			Description:  "Allocate funds for development",
			ProposalType: ProposalTypeTreasury,
			VotingType:   VotingTypeWeighted,
			StartTime:    now + 200,
			EndTime:      now + 200 + 86400,
			Threshold:    6000,
		}

		treasuryHash := randomHash()
		_, err = pm.CreateProposal(treasuryTx, developer, treasuryHash)
		if err != nil {
			t.Fatalf("Failed to create treasury proposal: %v", err)
		}

		// Technical proposal
		technicalTx := &ProposalTx{
			Fee:          150,
			Title:        "Protocol Upgrade v3.0",
			Description:  "Major protocol upgrade",
			ProposalType: ProposalTypeTechnical,
			VotingType:   VotingTypeReputation,
			StartTime:    now + 300,
			EndTime:      now + 300 + 86400,
			Threshold:    7000,
		}

		technicalHash := randomHash()
		_, err = pm.CreateProposal(technicalTx, community1, technicalHash)
		if err != nil {
			t.Fatalf("Failed to create technical proposal: %v", err)
		}

		// Verify proposals were created
		allProposals := pm.dao.ListAllProposals()
		if len(allProposals) != 3 {
			t.Errorf("Expected 3 proposals, got %d", len(allProposals))
		}
	})

	// Test 2: Proposal filtering and statistics
	t.Run("ProposalFilteringAndStatistics", func(t *testing.T) {
		// Test filtering by status
		pendingProposals := pm.GetProposalsByStatus(ProposalStatusPending)
		if len(pendingProposals) != 3 {
			t.Errorf("Expected 3 pending proposals, got %d", len(pendingProposals))
		}

		// Test filtering by type
		generalProposals := pm.GetProposalsByType(ProposalTypeGeneral)
		treasuryProposals := pm.GetProposalsByType(ProposalTypeTreasury)
		technicalProposals := pm.GetProposalsByType(ProposalTypeTechnical)

		if len(generalProposals) != 1 {
			t.Errorf("Expected 1 general proposal, got %d", len(generalProposals))
		}
		if len(treasuryProposals) != 1 {
			t.Errorf("Expected 1 treasury proposal, got %d", len(treasuryProposals))
		}
		if len(technicalProposals) != 1 {
			t.Errorf("Expected 1 technical proposal, got %d", len(technicalProposals))
		}

		// Test filtering by creator
		founderProposals := pm.GetProposalsByCreator(founder)
		developerProposals := pm.GetProposalsByCreator(developer)
		community1Proposals := pm.GetProposalsByCreator(community1)

		if len(founderProposals) != 1 {
			t.Errorf("Expected 1 founder proposal, got %d", len(founderProposals))
		}
		if len(developerProposals) != 1 {
			t.Errorf("Expected 1 developer proposal, got %d", len(developerProposals))
		}
		if len(community1Proposals) != 1 {
			t.Errorf("Expected 1 community1 proposal, got %d", len(community1Proposals))
		}

		// Test statistics
		stats := pm.GetProposalStatistics()
		if stats.Total != 3 {
			t.Errorf("Expected 3 total proposals, got %d", stats.Total)
		}
		if stats.StatusCounts[ProposalStatusPending] != 3 {
			t.Errorf("Expected 3 pending proposals in stats, got %d", stats.StatusCounts[ProposalStatusPending])
		}
	})

	// Test 3: Proposal cancellation
	t.Run("ProposalCancellation", func(t *testing.T) {
		// Create a proposal to cancel
		now := time.Now().Unix()
		cancelTx := &ProposalTx{
			Fee:          100,
			Title:        "Proposal to Cancel",
			Description:  "This will be cancelled",
			ProposalType: ProposalTypeGeneral,
			VotingType:   VotingTypeSimple,
			StartTime:    now + 1000,
			EndTime:      now + 1000 + 86400,
			Threshold:    5100,
		}

		cancelHash := randomHash()
		_, err := pm.CreateProposal(cancelTx, community2, cancelHash)
		if err != nil {
			t.Fatalf("Failed to create proposal to cancel: %v", err)
		}

		// Cancel the proposal
		err = pm.CancelProposal(cancelHash, community2)
		if err != nil {
			t.Fatalf("Failed to cancel proposal: %v", err)
		}

		// Verify cancellation
		cancelledProposals := pm.GetProposalsByStatus(ProposalStatusCancelled)
		if len(cancelledProposals) != 1 {
			t.Errorf("Expected 1 cancelled proposal, got %d", len(cancelledProposals))
		}

		// Test unauthorized cancellation
		anotherTx := &ProposalTx{
			Fee:          100,
			Title:        "Another Proposal",
			Description:  "This won't be cancelled by wrong user",
			ProposalType: ProposalTypeGeneral,
			VotingType:   VotingTypeSimple,
			StartTime:    now + 2000,
			EndTime:      now + 2000 + 86400,
			Threshold:    5100,
		}

		anotherHash := randomHash()
		_, err = pm.CreateProposal(anotherTx, community2, anotherHash)
		if err != nil {
			t.Fatalf("Failed to create another proposal: %v", err)
		}

		// Try to cancel with wrong user (should fail)
		err = pm.CancelProposal(anotherHash, founder)
		if err == nil {
			t.Error("Expected error when wrong user tries to cancel")
		}
	})

	// Test 4: Voting and progress tracking
	t.Run("VotingAndProgressTracking", func(t *testing.T) {
		// Create an active proposal for voting
		now := time.Now().Unix()
		votingTx := &ProposalTx{
			Fee:          100,
			Title:        "Voting Test Proposal",
			Description:  "Test voting functionality",
			ProposalType: ProposalTypeGeneral,
			VotingType:   VotingTypeSimple,
			StartTime:    now - 100, // Already started
			EndTime:      now + 86400,
			Threshold:    5100,
		}

		votingHash := randomHash()
		_, err := pm.CreateProposal(votingTx, founder, votingHash)
		if err != nil {
			t.Fatalf("Failed to create voting proposal: %v", err)
		}

		// Set proposal to active
		proposal, _ := dao.GetProposal(votingHash)
		proposal.Status = ProposalStatusActive

		// Cast votes
		vote1 := &VoteTx{
			Fee:        50,
			ProposalID: votingHash,
			Choice:     VoteChoiceYes,
			Weight:     5000,
			Reason:     "I support this proposal",
		}
		err = dao.Processor.ProcessVoteTx(vote1, developer)
		if err != nil {
			t.Fatalf("Failed to process developer vote: %v", err)
		}

		vote2 := &VoteTx{
			Fee:        50,
			ProposalID: votingHash,
			Choice:     VoteChoiceNo,
			Weight:     3000,
			Reason:     "I oppose this proposal",
		}
		err = dao.Processor.ProcessVoteTx(vote2, community1)
		if err != nil {
			t.Fatalf("Failed to process community1 vote: %v", err)
		}

		vote3 := &VoteTx{
			Fee:        50,
			ProposalID: votingHash,
			Choice:     VoteChoiceAbstain,
			Weight:     2000,
			Reason:     "I'm neutral",
		}
		err = dao.Processor.ProcessVoteTx(vote3, community2)
		if err != nil {
			t.Fatalf("Failed to process community2 vote: %v", err)
		}

		// Check voting progress
		progress, err := pm.GetProposalVotingProgress(votingHash)
		if err != nil {
			t.Fatalf("Failed to get voting progress: %v", err)
		}

		if progress.TotalVotes != 3 {
			t.Errorf("Expected 3 total votes, got %d", progress.TotalVotes)
		}
		if progress.YesVotes != 5000 {
			t.Errorf("Expected 5000 yes votes, got %d", progress.YesVotes)
		}
		if progress.NoVotes != 3000 {
			t.Errorf("Expected 3000 no votes, got %d", progress.NoVotes)
		}
		if progress.AbstainVotes != 2000 {
			t.Errorf("Expected 2000 abstain votes, got %d", progress.AbstainVotes)
		}

		// Verify voter information
		if len(progress.Voters) != 3 {
			t.Errorf("Expected 3 voters, got %d", len(progress.Voters))
		}

		// Check individual voter details
		voterFound := make(map[string]bool)
		for _, voter := range progress.Voters {
			voterStr := voter.Address.String()
			voterFound[voterStr] = true

			if voterStr == developer.String() {
				if voter.Choice != VoteChoiceYes || voter.Weight != 5000 {
					t.Error("Developer vote details incorrect")
				}
			} else if voterStr == community1.String() {
				if voter.Choice != VoteChoiceNo || voter.Weight != 3000 {
					t.Error("Community1 vote details incorrect")
				}
			} else if voterStr == community2.String() {
				if voter.Choice != VoteChoiceAbstain || voter.Weight != 2000 {
					t.Error("Community2 vote details incorrect")
				}
			}
		}

		if !voterFound[developer.String()] || !voterFound[community1.String()] || !voterFound[community2.String()] {
			t.Error("Not all voters found in progress")
		}
	})

	// Test 5: Proposal execution
	t.Run("ProposalExecution", func(t *testing.T) {
		// Create a proposal for execution
		now := time.Now().Unix()
		execTx := &ProposalTx{
			Fee:          100,
			Title:        "Execution Test Proposal",
			Description:  "Test proposal execution",
			ProposalType: ProposalTypeGeneral,
			VotingType:   VotingTypeSimple,
			StartTime:    now - 200,
			EndTime:      now + 86400,
			Threshold:    5100,
		}

		execHash := randomHash()
		proposal, err := pm.CreateProposal(execTx, founder, execHash)
		if err != nil {
			t.Fatalf("Failed to create execution proposal: %v", err)
		}

		// Try to execute before it's passed (should fail)
		err = pm.ExecuteProposal(execHash, founder)
		if err == nil {
			t.Error("Expected error when executing non-passed proposal")
		}

		// Set proposal to passed
		proposal.Status = ProposalStatusPassed
		proposal.Results.Passed = true

		// Execute proposal
		err = pm.ExecuteProposal(execHash, founder)
		if err != nil {
			t.Fatalf("Failed to execute proposal: %v", err)
		}

		if proposal.Status != ProposalStatusExecuted {
			t.Errorf("Expected status executed, got %d", proposal.Status)
		}
	})

	// Test 6: Status updates
	t.Run("StatusUpdates", func(t *testing.T) {
		// Update all proposal statuses
		err := pm.UpdateAllProposalStatuses()
		if err != nil {
			t.Fatalf("Failed to update proposal statuses: %v", err)
		}

		// Verify some proposals became active
		activeProposals := pm.GetProposalsByStatus(ProposalStatusActive)
		if len(activeProposals) == 0 {
			t.Error("Expected some active proposals after status update")
		}
	})

	// Test 7: Final statistics
	t.Run("FinalStatistics", func(t *testing.T) {
		finalStats := pm.GetProposalStatistics()

		// Should have multiple proposals across different statuses
		if finalStats.Total < 5 {
			t.Errorf("Expected at least 5 total proposals, got %d", finalStats.Total)
		}

		// Should have at least one executed proposal
		if finalStats.StatusCounts[ProposalStatusExecuted] < 1 {
			t.Errorf("Expected at least 1 executed proposal, got %d", finalStats.StatusCounts[ProposalStatusExecuted])
		}

		// Should have proposals of different types
		if finalStats.TypeCounts[ProposalTypeGeneral] < 1 {
			t.Error("Expected at least 1 general proposal")
		}
		if finalStats.TypeCounts[ProposalTypeTreasury] < 1 {
			t.Error("Expected at least 1 treasury proposal")
		}
		if finalStats.TypeCounts[ProposalTypeTechnical] < 1 {
			t.Error("Expected at least 1 technical proposal")
		}
	})

	t.Log("✓ Proposal Management Integration Test completed successfully!")
}
