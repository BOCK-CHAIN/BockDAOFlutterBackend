package dao

import (
	"fmt"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// ProposalManagementExample demonstrates the enhanced proposal management system
func ProposalManagementExample() {
	fmt.Println("=== ProjectX DAO Enhanced Proposal Management Example ===")

	// 1. Create a new DAO
	dao := NewDAO("PXGOV", "ProjectX Governance Token", 18)
	fmt.Println("✓ Created new DAO with enhanced proposal management")

	// 2. Set up initial token distribution
	founder := crypto.GeneratePrivateKey()
	developer := crypto.GeneratePrivateKey()
	community1 := crypto.GeneratePrivateKey()
	community2 := crypto.GeneratePrivateKey()

	distributions := map[string]uint64{
		founder.PublicKey().String():    15000, // 15,000 tokens
		developer.PublicKey().String():  10000, // 10,000 tokens
		community1.PublicKey().String(): 8000,  // 8,000 tokens
		community2.PublicKey().String(): 7000,  // 7,000 tokens
	}

	err := dao.InitialTokenDistribution(distributions)
	if err != nil {
		fmt.Printf("✗ Failed to distribute tokens: %v\n", err)
		return
	}
	fmt.Printf("✓ Distributed %d total tokens to %d addresses\n", dao.GetTotalSupply(), len(distributions))

	// 3. Initialize treasury
	treasurySigners := []crypto.PublicKey{
		founder.PublicKey(),
		developer.PublicKey(),
	}
	err = dao.InitializeTreasury(treasurySigners, 2)
	if err != nil {
		fmt.Printf("✗ Failed to initialize treasury: %v\n", err)
		return
	}
	dao.AddTreasuryFunds(100000)
	fmt.Printf("✓ Initialized treasury with %d units\n", dao.GetTreasuryBalance())

	// 4. Create multiple proposals of different types
	fmt.Println("\n--- Creating Various Proposal Types ---")

	now := time.Now().Unix()
	proposalHashes := make([]types.Hash, 0)

	// General governance proposal
	generalProposal := &ProposalTx{
		Fee:          100,
		Title:        "Community Guidelines Update",
		Description:  "Update community guidelines to include new participation rules and code of conduct for better governance.",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    now + 300,         // 5 minutes from now
		EndTime:      now + 300 + 86400, // 24+ hours from start
		Threshold:    5100,              // 51%
		MetadataHash: types.Hash{},
	}

	txHash1 := generateMockHash("general_proposal")
	_, err = dao.ProposalManager.CreateProposal(generalProposal, founder.PublicKey(), txHash1)
	if err != nil {
		fmt.Printf("✗ Failed to create general proposal: %v\n", err)
		return
	}
	proposalHashes = append(proposalHashes, txHash1)
	fmt.Printf("✓ Created general proposal: '%s'\n", generalProposal.Title)

	// Treasury proposal
	treasuryProposal := &ProposalTx{
		Fee:          200,
		Title:        "Development Fund Allocation",
		Description:  "Allocate 50,000 units from treasury for core development team compensation and infrastructure costs.",
		ProposalType: ProposalTypeTreasury,
		VotingType:   VotingTypeWeighted,
		StartTime:    now + 600,         // 10 minutes from now
		EndTime:      now + 600 + 86400, // 24+ hours from start
		Threshold:    6000,              // 60%
		MetadataHash: types.Hash{},
	}

	txHash2 := generateMockHash("treasury_proposal")
	_, err = dao.ProposalManager.CreateProposal(treasuryProposal, developer.PublicKey(), txHash2)
	if err != nil {
		fmt.Printf("✗ Failed to create treasury proposal: %v\n", err)
		return
	}
	proposalHashes = append(proposalHashes, txHash2)
	fmt.Printf("✓ Created treasury proposal: '%s'\n", treasuryProposal.Title)

	// Technical proposal
	technicalProposal := &ProposalTx{
		Fee:          150,
		Title:        "Protocol Upgrade v3.0",
		Description:  "Implement major protocol upgrade including consensus improvements, gas optimization, and new VM features.",
		ProposalType: ProposalTypeTechnical,
		VotingType:   VotingTypeReputation,
		StartTime:    now + 900,         // 15 minutes from now
		EndTime:      now + 900 + 86400, // 24+ hours from start
		Threshold:    7000,              // 70%
		MetadataHash: types.Hash{},
	}

	txHash3 := generateMockHash("technical_proposal")
	_, err = dao.ProposalManager.CreateProposal(technicalProposal, founder.PublicKey(), txHash3)
	if err != nil {
		fmt.Printf("✗ Failed to create technical proposal: %v\n", err)
		return
	}
	proposalHashes = append(proposalHashes, txHash3)
	fmt.Printf("✓ Created technical proposal: '%s'\n", technicalProposal.Title)

	// 5. Demonstrate proposal filtering and statistics
	fmt.Println("\n--- Proposal Management Features ---")

	// Get proposals by status
	pendingProposals := dao.ProposalManager.GetProposalsByStatus(ProposalStatusPending)
	fmt.Printf("✓ Found %d pending proposals\n", len(pendingProposals))

	// Get proposals by type
	treasuryProposals := dao.ProposalManager.GetProposalsByType(ProposalTypeTreasury)
	technicalProposals := dao.ProposalManager.GetProposalsByType(ProposalTypeTechnical)
	fmt.Printf("✓ Found %d treasury proposals and %d technical proposals\n",
		len(treasuryProposals), len(technicalProposals))

	// Get proposals by creator
	founderProposals := dao.ProposalManager.GetProposalsByCreator(founder.PublicKey())
	developerProposals := dao.ProposalManager.GetProposalsByCreator(developer.PublicKey())
	fmt.Printf("✓ Founder created %d proposals, Developer created %d proposals\n",
		len(founderProposals), len(developerProposals))

	// 6. Demonstrate proposal cancellation
	fmt.Println("\n--- Proposal Cancellation ---")

	// Create a proposal to cancel
	cancelProposal := &ProposalTx{
		Fee:          100,
		Title:        "Proposal to Cancel",
		Description:  "This proposal will be cancelled to demonstrate the feature",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    now + 3600,         // 1 hour from now
		EndTime:      now + 3600 + 86400, // 24+ hours from start
		Threshold:    5100,
		MetadataHash: types.Hash{},
	}

	cancelHash := generateMockHash("cancel_proposal")
	_, err = dao.ProposalManager.CreateProposal(cancelProposal, community1.PublicKey(), cancelHash)
	if err != nil {
		fmt.Printf("✗ Failed to create proposal to cancel: %v\n", err)
		return
	}
	fmt.Printf("✓ Created proposal to cancel: '%s'\n", cancelProposal.Title)

	// Cancel the proposal
	err = dao.ProposalManager.CancelProposal(cancelHash, community1.PublicKey())
	if err != nil {
		fmt.Printf("✗ Failed to cancel proposal: %v\n", err)
		return
	}
	fmt.Printf("✓ Successfully cancelled proposal\n")

	// 7. Simulate voting and demonstrate voting progress
	fmt.Println("\n--- Voting Simulation ---")

	// Update proposal statuses to make them active
	dao.ProposalManager.UpdateAllProposalStatuses()

	// Find an active proposal to vote on
	activeProposals := dao.ProposalManager.GetProposalsByStatus(ProposalStatusActive)
	if len(activeProposals) > 0 {
		activeProposal := activeProposals[0]
		fmt.Printf("✓ Found active proposal: '%s'\n", activeProposal.Title)

		// Cast votes from different users
		voteTx1 := &VoteTx{
			Fee:        50,
			ProposalID: activeProposal.ID,
			Choice:     VoteChoiceYes,
			Weight:     5000,
			Reason:     "I strongly support this initiative",
		}
		err = dao.Processor.ProcessVoteTx(voteTx1, developer.PublicKey())
		if err != nil {
			fmt.Printf("✗ Failed to process developer vote: %v\n", err)
		} else {
			fmt.Printf("✓ Developer voted YES with weight %d\n", voteTx1.Weight)
		}

		voteTx2 := &VoteTx{
			Fee:        50,
			ProposalID: activeProposal.ID,
			Choice:     VoteChoiceNo,
			Weight:     3000,
			Reason:     "Need more discussion before implementation",
		}
		err = dao.Processor.ProcessVoteTx(voteTx2, community1.PublicKey())
		if err != nil {
			fmt.Printf("✗ Failed to process community1 vote: %v\n", err)
		} else {
			fmt.Printf("✓ Community1 voted NO with weight %d\n", voteTx2.Weight)
		}

		voteTx3 := &VoteTx{
			Fee:        50,
			ProposalID: activeProposal.ID,
			Choice:     VoteChoiceAbstain,
			Weight:     2000,
			Reason:     "Neutral on this topic",
		}
		err = dao.Processor.ProcessVoteTx(voteTx3, community2.PublicKey())
		if err != nil {
			fmt.Printf("✗ Failed to process community2 vote: %v\n", err)
		} else {
			fmt.Printf("✓ Community2 abstained with weight %d\n", voteTx3.Weight)
		}

		// Get detailed voting progress
		progress, err := dao.ProposalManager.GetProposalVotingProgress(activeProposal.ID)
		if err != nil {
			fmt.Printf("✗ Failed to get voting progress: %v\n", err)
		} else {
			fmt.Printf("✓ Voting Progress: %d total votes, %d YES, %d NO, %d ABSTAIN\n",
				progress.TotalVotes, progress.YesVotes, progress.NoVotes, progress.AbstainVotes)
			fmt.Printf("  Quorum reached: %t, Time remaining: %d seconds\n",
				progress.QuorumReached, progress.TimeRemaining)

			// Show individual voter details
			fmt.Println("  Voter details:")
			for _, voter := range progress.Voters {
				choiceStr := map[VoteChoice]string{
					VoteChoiceYes:     "YES",
					VoteChoiceNo:      "NO",
					VoteChoiceAbstain: "ABSTAIN",
				}[voter.Choice]
				fmt.Printf("    - %s: %s (weight: %d) - %s\n",
					voter.Address.String()[:8]+"...", choiceStr, voter.Weight, voter.Reason)
			}
		}
	}

	// 8. Show comprehensive statistics
	fmt.Println("\n--- DAO Statistics ---")

	stats := dao.ProposalManager.GetProposalStatistics()
	fmt.Printf("✓ Total proposals: %d (Passed: %d)\n", stats.Total, stats.Passed)

	fmt.Println("  Status breakdown:")
	statusNames := map[ProposalStatus]string{
		ProposalStatusPending:   "Pending",
		ProposalStatusActive:    "Active",
		ProposalStatusPassed:    "Passed",
		ProposalStatusRejected:  "Rejected",
		ProposalStatusExecuted:  "Executed",
		ProposalStatusCancelled: "Cancelled",
	}
	for status, count := range stats.StatusCounts {
		if count > 0 {
			fmt.Printf("    - %s: %d\n", statusNames[status], count)
		}
	}

	fmt.Println("  Type breakdown:")
	typeNames := map[ProposalType]string{
		ProposalTypeGeneral:   "General",
		ProposalTypeTreasury:  "Treasury",
		ProposalTypeTechnical: "Technical",
		ProposalTypeParameter: "Parameter",
	}
	for pType, count := range stats.TypeCounts {
		if count > 0 {
			fmt.Printf("    - %s: %d\n", typeNames[pType], count)
		}
	}

	// 9. Demonstrate proposal execution
	fmt.Println("\n--- Proposal Execution ---")

	// Find a passed proposal to execute (we'll manually set one to passed for demo)
	if len(activeProposals) > 0 {
		demoProposal := activeProposals[0]
		// Manually set to passed for demonstration
		demoProposal.Status = ProposalStatusPassed
		demoProposal.Results.Passed = true

		err = dao.ProposalManager.ExecuteProposal(demoProposal.ID, founder.PublicKey())
		if err != nil {
			fmt.Printf("✗ Failed to execute proposal: %v\n", err)
		} else {
			fmt.Printf("✓ Successfully executed proposal: '%s'\n", demoProposal.Title)
		}
	}

	// 10. Final summary
	fmt.Println("\n--- Final Summary ---")
	finalStats := dao.ProposalManager.GetProposalStatistics()
	fmt.Printf("✓ DAO now has %d total proposals across all types\n", finalStats.Total)
	fmt.Printf("✓ Token holders: %d addresses with tokens\n", len(dao.TokenState.Balances))
	fmt.Printf("✓ Treasury balance: %d units\n", dao.GetTreasuryBalance())

	fmt.Println("\n✓ Enhanced Proposal Management Example completed successfully!")
}

// generateMockHash creates a mock hash for demonstration purposes
func generateMockHash(seed string) types.Hash {
	var hash types.Hash
	copy(hash[:], seed)
	// Fill remaining bytes with a pattern
	for i := len(seed); i < 32; i++ {
		hash[i] = byte(i % 256)
	}
	return hash
}
