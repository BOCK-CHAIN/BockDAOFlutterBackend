package dao

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// RunAnalyticsExample demonstrates the analytics and reporting system
func RunAnalyticsExample() {
	fmt.Println("=== ProjectX DAO Analytics and Reporting System Demo ===")

	// Create a DAO instance
	dao := NewDAO("ANALYTICS", "Analytics Demo Token", 18)

	// Create test users
	alice := crypto.GeneratePrivateKey()
	bob := crypto.GeneratePrivateKey()
	charlie := crypto.GeneratePrivateKey()
	diana := crypto.GeneratePrivateKey()

	fmt.Println("1. Setting up DAO with initial token distribution...")

	// Initialize token distribution
	distributions := map[string]uint64{
		alice.PublicKey().String():   25000,
		bob.PublicKey().String():     20000,
		charlie.PublicKey().String(): 15000,
		diana.PublicKey().String():   10000,
	}

	err := dao.InitialTokenDistribution(distributions)
	if err != nil {
		fmt.Printf("Error initializing distribution: %v\n", err)
		return
	}

	// Initialize treasury
	treasurySigners := []crypto.PublicKey{
		alice.PublicKey(),
		bob.PublicKey(),
	}
	err = dao.InitializeTreasury(treasurySigners, 2)
	if err != nil {
		fmt.Printf("Error initializing treasury: %v\n", err)
		return
	}

	// Add treasury funds
	dao.AddTreasuryFunds(100000)

	fmt.Println("✓ DAO initialized with 4 members and 100,000 treasury tokens")

	// Create sample proposals and votes
	fmt.Println("2. Creating sample governance activity...")

	now := time.Now().Unix()

	// Proposal 1: Passed proposal
	proposal1 := &Proposal{
		ID:           types.Hash{1},
		Creator:      alice.PublicKey(),
		Title:        "Improve Documentation",
		Description:  "Proposal to enhance project documentation",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		Status:       ProposalStatusPassed,
		StartTime:    now - 172800, // 2 days ago
		EndTime:      now - 86400,  // 1 day ago
		Threshold:    3,
		Results: &VoteResults{
			YesVotes:     3,
			NoVotes:      1,
			AbstainVotes: 0,
			TotalVoters:  4,
			Quorum:       3,
			Passed:       true,
		},
	}
	dao.GovernanceState.Proposals[proposal1.ID] = proposal1

	// Add votes for proposal 1
	dao.GovernanceState.Votes[proposal1.ID] = map[string]*Vote{
		alice.PublicKey().String(): {
			Voter:     alice.PublicKey(),
			Choice:    VoteChoiceYes,
			Weight:    25000,
			Timestamp: now - 150000,
		},
		bob.PublicKey().String(): {
			Voter:     bob.PublicKey(),
			Choice:    VoteChoiceYes,
			Weight:    20000,
			Timestamp: now - 140000,
		},
		charlie.PublicKey().String(): {
			Voter:     charlie.PublicKey(),
			Choice:    VoteChoiceYes,
			Weight:    15000,
			Timestamp: now - 130000,
		},
		diana.PublicKey().String(): {
			Voter:     diana.PublicKey(),
			Choice:    VoteChoiceNo,
			Weight:    10000,
			Timestamp: now - 120000,
		},
	}

	// Proposal 2: Rejected treasury proposal
	proposal2 := &Proposal{
		ID:           types.Hash{2},
		Creator:      bob.PublicKey(),
		Title:        "Marketing Campaign Funding",
		Description:  "Proposal to fund a marketing campaign",
		ProposalType: ProposalTypeTreasury,
		VotingType:   VotingTypeWeighted,
		Status:       ProposalStatusRejected,
		StartTime:    now - 129600, // 1.5 days ago
		EndTime:      now - 43200,  // 12 hours ago
		Threshold:    3,
		Results: &VoteResults{
			YesVotes:     1,
			NoVotes:      2,
			AbstainVotes: 1,
			TotalVoters:  4,
			Quorum:       3,
			Passed:       false,
		},
	}
	dao.GovernanceState.Proposals[proposal2.ID] = proposal2

	// Add votes for proposal 2
	dao.GovernanceState.Votes[proposal2.ID] = map[string]*Vote{
		alice.PublicKey().String(): {
			Voter:     alice.PublicKey(),
			Choice:    VoteChoiceNo,
			Weight:    25000,
			Timestamp: now - 100000,
		},
		bob.PublicKey().String(): {
			Voter:     bob.PublicKey(),
			Choice:    VoteChoiceYes,
			Weight:    20000,
			Timestamp: now - 90000,
		},
		charlie.PublicKey().String(): {
			Voter:     charlie.PublicKey(),
			Choice:    VoteChoiceNo,
			Weight:    15000,
			Timestamp: now - 80000,
		},
		diana.PublicKey().String(): {
			Voter:     diana.PublicKey(),
			Choice:    VoteChoiceAbstain,
			Weight:    10000,
			Timestamp: now - 70000,
		},
	}

	// Proposal 3: Active proposal
	proposal3 := &Proposal{
		ID:           types.Hash{3},
		Creator:      charlie.PublicKey(),
		Title:        "Technical Upgrade",
		Description:  "Proposal for system technical upgrade",
		ProposalType: ProposalTypeTechnical,
		VotingType:   VotingTypeSimple,
		Status:       ProposalStatusActive,
		StartTime:    now - 7200,  // 2 hours ago
		EndTime:      now + 79200, // 22 hours from now
		Threshold:    3,
	}
	dao.GovernanceState.Proposals[proposal3.ID] = proposal3

	// Add some votes for active proposal
	dao.GovernanceState.Votes[proposal3.ID] = map[string]*Vote{
		alice.PublicKey().String(): {
			Voter:     alice.PublicKey(),
			Choice:    VoteChoiceYes,
			Weight:    25000,
			Timestamp: now - 3600,
		},
		bob.PublicKey().String(): {
			Voter:     bob.PublicKey(),
			Choice:    VoteChoiceYes,
			Weight:    20000,
			Timestamp: now - 1800,
		},
	}

	// Create delegation
	dao.GovernanceState.Delegations[diana.PublicKey().String()] = &Delegation{
		Delegator: diana.PublicKey(),
		Delegate:  alice.PublicKey(),
		StartTime: now - 86400,
		EndTime:   now + 86400,
		Active:    true,
	}

	// Create treasury transactions
	tx1 := &PendingTx{
		ID:        types.Hash{10},
		Recipient: charlie.PublicKey(),
		Amount:    5000,
		Purpose:   "Development",
		CreatedAt: now - 7200,
		Executed:  true,
	}
	dao.GovernanceState.Treasury.Transactions[tx1.ID] = tx1

	tx2 := &PendingTx{
		ID:        types.Hash{11},
		Recipient: diana.PublicKey(),
		Amount:    3000,
		Purpose:   "Design",
		CreatedAt: now - 3600,
		ExpiresAt: now + 86400,
		Executed:  false,
	}
	dao.GovernanceState.Treasury.Transactions[tx2.ID] = tx2

	// Update treasury balance
	dao.GovernanceState.Treasury.Balance = 92000 // 100000 - 5000 - 3000

	fmt.Println("✓ Created 3 proposals (1 passed, 1 rejected, 1 active)")
	fmt.Println("✓ Added voting activity from all members")
	fmt.Println("✓ Created delegation and treasury transactions")

	// Demonstrate analytics
	fmt.Println("3. Generating Analytics Reports...")

	// Participation Metrics
	fmt.Println("=== GOVERNANCE PARTICIPATION METRICS ===")
	participation := dao.GetGovernanceParticipationMetrics()

	fmt.Printf("Total Proposals: %d\n", participation.TotalProposals)
	fmt.Printf("Active Proposals: %d\n", participation.ActiveProposals)
	fmt.Printf("Total Votes Cast: %d\n", participation.TotalVotes)
	fmt.Printf("Unique Voters: %d\n", participation.UniqueVoters)
	fmt.Printf("Participation Rate: %.1f%%\n", participation.ParticipationRate)
	fmt.Printf("Average Votes per User: %.1f\n", participation.AverageVotesPerUser)

	fmt.Println("\nVoting Patterns:")
	fmt.Printf("  Yes Votes: %d\n", participation.VotingPatterns[VoteChoiceYes])
	fmt.Printf("  No Votes: %d\n", participation.VotingPatterns[VoteChoiceNo])
	fmt.Printf("  Abstain Votes: %d\n", participation.VotingPatterns[VoteChoiceAbstain])

	fmt.Println("\nProposals by Type:")
	typeNames := map[ProposalType]string{
		ProposalTypeGeneral:   "General",
		ProposalTypeTreasury:  "Treasury",
		ProposalTypeTechnical: "Technical",
		ProposalTypeParameter: "Parameter",
	}
	for pType, count := range participation.ProposalsByType {
		fmt.Printf("  %s: %d\n", typeNames[pType], count)
	}

	fmt.Println("\nTop Participants:")
	for i, participant := range participation.TopParticipants {
		if i >= 3 { // Show top 3
			break
		}
		fmt.Printf("  #%d: %s (Votes: %d, Participation: %.1f%%)\n",
			i+1, truncateAddress(participant.Address), participant.VotesCast, participant.ParticipationRate)
	}

	fmt.Printf("\nDelegation Rate: %.1f%% (%d active delegations)\n",
		participation.DelegationMetrics.DelegationRate, participation.DelegationMetrics.ActiveDelegations)

	// Treasury Metrics
	fmt.Println("\n=== TREASURY PERFORMANCE METRICS ===")
	treasury := dao.GetTreasuryPerformanceMetrics()

	fmt.Printf("Current Balance: %s tokens\n", formatTokenAmount(treasury.CurrentBalance))
	fmt.Printf("Total Outflows: %s tokens\n", formatTokenAmount(treasury.TotalOutflows))
	fmt.Printf("Transaction Count: %d\n", treasury.TransactionCount)
	fmt.Printf("Executed Transactions: %d\n", treasury.ExecutedTransactions)
	fmt.Printf("Pending Transactions: %d\n", treasury.PendingTransactions)
	fmt.Printf("Signing Efficiency: %.1f%%\n", treasury.SigningEfficiency)

	if len(treasury.TransactionsByPurpose) > 0 {
		fmt.Println("\nTransactions by Purpose:")
		for purpose, count := range treasury.TransactionsByPurpose {
			fmt.Printf("  %s: %d\n", purpose, count)
		}
	}

	// Proposal Analytics
	fmt.Println("\n=== PROPOSAL ANALYTICS ===")
	proposals := dao.GetProposalAnalytics()

	fmt.Printf("Total Proposals: %d\n", proposals.TotalProposals)
	fmt.Printf("Passed Proposals: %d\n", proposals.PassedProposals)
	fmt.Printf("Rejected Proposals: %d\n", proposals.RejectedProposals)
	fmt.Printf("Pending Proposals: %d\n", proposals.PendingProposals)
	fmt.Printf("Success Rate: %.1f%%\n", proposals.SuccessRate)
	fmt.Printf("Quorum Achievement Rate: %.1f%%\n", proposals.QuorumAchievementRate)

	fmt.Println("\nProposals by Creator:")
	for creator, count := range proposals.ProposalsByCreator {
		fmt.Printf("  %s: %d\n", truncateAddress(creator), count)
	}

	if len(proposals.SuccessRateByType) > 0 {
		fmt.Println("\nSuccess Rate by Type:")
		for pType, rate := range proposals.SuccessRateByType {
			fmt.Printf("  %s: %.1f%%\n", typeNames[pType], rate)
		}
	}

	// Health Metrics
	fmt.Println("\n=== DAO HEALTH METRICS ===")
	health := dao.GetDAOHealthMetrics()

	fmt.Printf("Overall Health Score: %.1f/100\n", health.OverallScore)
	fmt.Printf("Participation Health: %.1f/100\n", health.ParticipationHealth)
	fmt.Printf("Treasury Health: %.1f/100\n", health.TreasuryHealth)
	fmt.Printf("Governance Health: %.1f/100\n", health.GovernanceHealth)
	fmt.Printf("Security Health: %.1f/100\n", health.SecurityHealth)
	fmt.Printf("Health Trend: %s\n", health.HealthTrend)

	if len(health.RiskIndicators) > 0 {
		fmt.Println("\nRisk Indicators:")
		for _, risk := range health.RiskIndicators {
			fmt.Printf("  ⚠️  %s (%s): %s\n", risk.Type, risk.Severity, risk.Description)
		}
	}

	if len(health.Recommendations) > 0 {
		fmt.Println("\nRecommendations:")
		for _, rec := range health.Recommendations {
			fmt.Printf("  💡 %s\n", rec)
		}
	}

	// JSON Export Example
	fmt.Println("\n=== ANALYTICS SUMMARY (JSON) ===")
	summary := dao.GetAnalyticsSummary()

	// Pretty print a subset of the summary
	summarySubset := map[string]interface{}{
		"participation_rate":   participation.ParticipationRate,
		"total_proposals":      participation.TotalProposals,
		"treasury_balance":     treasury.CurrentBalance,
		"success_rate":         proposals.SuccessRate,
		"overall_health_score": health.OverallScore,
		"generated_at":         summary["generated_at"],
	}

	jsonData, err := json.MarshalIndent(summarySubset, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
	} else {
		fmt.Println(string(jsonData))
	}

	fmt.Println("\n=== Analytics Demo Complete ===")
	fmt.Println("The analytics system provides comprehensive insights into:")
	fmt.Println("• Governance participation and voting patterns")
	fmt.Println("• Treasury performance and transaction efficiency")
	fmt.Println("• Proposal success rates and trends")
	fmt.Println("• Overall DAO health and risk indicators")
	fmt.Println("• Real-time metrics for dashboard display")
}

// Helper functions for formatting
func truncateAddress(address string) string {
	if len(address) < 10 {
		return address
	}
	return fmt.Sprintf("%s...%s", address[:6], address[len(address)-4:])
}

func formatTokenAmount(amount uint64) string {
	if amount >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(amount)/1000000)
	} else if amount >= 1000 {
		return fmt.Sprintf("%.1fK", float64(amount)/1000)
	}
	return fmt.Sprintf("%d", amount)
}
