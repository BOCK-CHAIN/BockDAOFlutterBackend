package dao

import (
	"fmt"
	"log"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// ParameterManagementExample demonstrates the complete parameter management workflow
func ParameterManagementExample() {
	fmt.Println("=== ProjectX DAO Parameter Management Example ===")

	// Create a new DAO instance
	dao := NewDAO("GOVTOKEN", "Governance Token", 18)

	// Create test users
	founder := crypto.GeneratePrivateKey()
	proposer := crypto.GeneratePrivateKey()
	voter1 := crypto.GeneratePrivateKey()
	voter2 := crypto.GeneratePrivateKey()
	voter3 := crypto.GeneratePrivateKey()

	fmt.Printf("Created test users:\n")
	fmt.Printf("- Founder: %s\n", founder.PublicKey().String()[:16]+"...")
	fmt.Printf("- Proposer: %s\n", proposer.PublicKey().String()[:16]+"...")
	fmt.Printf("- Voter 1: %s\n", voter1.PublicKey().String()[:16]+"...")
	fmt.Printf("- Voter 2: %s\n", voter2.PublicKey().String()[:16]+"...")
	fmt.Printf("- Voter 3: %s\n", voter3.PublicKey().String()[:16]+"...")

	// Initialize token distribution
	distributions := map[string]uint64{
		founder.PublicKey().String():  20000,
		proposer.PublicKey().String(): 15000,
		voter1.PublicKey().String():   10000,
		voter2.PublicKey().String():   8000,
		voter3.PublicKey().String():   5000,
	}

	err := dao.InitialTokenDistribution(distributions)
	if err != nil {
		log.Fatalf("Failed to initialize token distribution: %v", err)
	}

	fmt.Printf("\n✓ Initial token distribution completed\n")
	fmt.Printf("Total supply: %d tokens\n", dao.GetTotalSupply())

	// Initialize security roles
	founders := []crypto.PublicKey{founder.PublicKey()}
	err = dao.InitializeFounderRoles(founders)
	if err != nil {
		log.Fatalf("Failed to initialize founder roles: %v", err)
	}

	fmt.Printf("✓ Security roles initialized\n")

	// Step 1: Display current parameter configuration
	fmt.Printf("\n=== Step 1: Current Parameter Configuration ===\n")
	currentConfig := dao.GetParameterConfig()
	fmt.Printf("Min Proposal Threshold: %d tokens\n", currentConfig.MinProposalThreshold)
	fmt.Printf("Voting Period: %d seconds (%d hours)\n", currentConfig.VotingPeriod, currentConfig.VotingPeriod/3600)
	fmt.Printf("Quorum Threshold: %d votes\n", currentConfig.QuorumThreshold)
	fmt.Printf("Passing Threshold: %d basis points (%.1f%%)\n", currentConfig.PassingThreshold, float64(currentConfig.PassingThreshold)/100)
	fmt.Printf("Treasury Threshold: %d tokens\n", currentConfig.TreasuryThreshold)
	fmt.Printf("Delegation Enabled: %t\n", currentConfig.DelegationEnabled)
	fmt.Printf("Token Burning Enabled: %t\n", currentConfig.TokenBurningEnabled)

	// Step 2: Create a parameter change proposal
	fmt.Printf("\n=== Step 2: Creating Parameter Change Proposal ===\n")

	parameterChanges := map[string]interface{}{
		"voting_period":      int64(172800), // Increase to 48 hours
		"quorum_threshold":   uint64(3000),  // Reduce quorum requirement
		"passing_threshold":  uint64(6000),  // Increase to 60%
		"treasury_threshold": uint64(7500),  // Increase treasury threshold
		"delegation_enabled": true,          // Ensure delegation is enabled
	}

	justification := `
	Proposal to improve governance parameters:
	1. Increase voting period to 48 hours for better participation
	2. Reduce quorum threshold to make proposals more viable
	3. Increase passing threshold to 60% for stronger consensus
	4. Increase treasury threshold for better security
	5. Ensure delegation remains enabled for flexibility
	`

	startTime := time.Now().Unix() + 300      // Start in 5 minutes
	endTime := time.Now().Unix() + 3900       // End in 65 minutes
	effectiveTime := time.Now().Unix() + 4200 // Effective in 70 minutes

	proposalID, err := dao.CreateParameterProposal(
		proposer.PublicKey(),
		parameterChanges,
		justification,
		effectiveTime,
		VotingTypeSimple,
		startTime,
		endTime,
		3000, // Threshold for this proposal
	)

	if err != nil {
		log.Fatalf("Failed to create parameter proposal: %v", err)
	}

	fmt.Printf("✓ Parameter proposal created with ID: %x\n", proposalID[:8])
	fmt.Printf("Proposed changes:\n")
	for param, value := range parameterChanges {
		fmt.Printf("  - %s: %v\n", param, value)
	}

	// Step 3: Validate parameter constraints
	fmt.Printf("\n=== Step 3: Parameter Validation and Constraints ===\n")

	// Show constraints for key parameters
	keyParams := []string{"voting_period", "min_proposal_threshold", "passing_threshold", "treasury_threshold"}

	for _, param := range keyParams {
		constraints := dao.GetParameterConstraints(param)
		fmt.Printf("%s constraints:\n", param)
		fmt.Printf("  Type: %s\n", constraints["type"])
		if min, exists := constraints["min"]; exists {
			fmt.Printf("  Min: %v\n", min)
		}
		if max, exists := constraints["max"]; exists {
			fmt.Printf("  Max: %v\n", max)
		}
		if unit, exists := constraints["unit"]; exists {
			fmt.Printf("  Unit: %s\n", unit)
		}
		fmt.Println()
	}

	// Test parameter change validation
	fmt.Printf("Testing parameter change validation:\n")

	// Valid change
	validChanges := map[string]interface{}{
		"voting_period": int64(259200), // 72 hours
	}
	err = dao.ValidateParameterProposal(proposer.PublicKey(), validChanges)
	if err != nil {
		fmt.Printf("  ✗ Valid change rejected: %v\n", err)
	} else {
		fmt.Printf("  ✓ Valid change accepted\n")
	}

	// Invalid change
	invalidChanges := map[string]interface{}{
		"min_proposal_threshold": uint64(0), // Invalid: zero threshold
	}
	err = dao.ValidateParameterProposal(proposer.PublicKey(), invalidChanges)
	if err != nil {
		fmt.Printf("  ✓ Invalid change correctly rejected: %v\n", err)
	} else {
		fmt.Printf("  ✗ Invalid change incorrectly accepted\n")
	}

	// Step 4: Simulate proposal activation and voting
	fmt.Printf("\n=== Step 4: Voting on Parameter Proposal ===\n")

	// Manually activate the proposal for demonstration
	proposal, err := dao.GetProposal(proposalID)
	if err != nil {
		log.Fatalf("Failed to get proposal: %v", err)
	}
	proposal.Status = ProposalStatusActive
	proposal.StartTime = time.Now().Unix() - 300 // Started 5 minutes ago

	fmt.Printf("Proposal is now active for voting\n")

	// Cast votes
	votes := []struct {
		voter  crypto.PrivateKey
		choice VoteChoice
		weight uint64
		reason string
	}{
		{voter1, VoteChoiceYes, 5000, "Support improved governance parameters"},
		{voter2, VoteChoiceYes, 4000, "Agree with longer voting period"},
		{voter3, VoteChoiceNo, 2000, "Current parameters are fine"},
	}

	for i, vote := range votes {
		voteTx := &VoteTx{
			Fee:        100,
			ProposalID: proposalID,
			Choice:     vote.choice,
			Weight:     vote.weight,
			Reason:     vote.reason,
		}

		err = dao.ProcessDAOTransaction(voteTx, vote.voter.PublicKey(), types.Hash{byte(i + 1)})
		if err != nil {
			fmt.Printf("  ✗ Vote from %s failed: %v\n", vote.voter.PublicKey().String()[:16]+"...", err)
		} else {
			choiceStr := map[VoteChoice]string{
				VoteChoiceYes:     "YES",
				VoteChoiceNo:      "NO",
				VoteChoiceAbstain: "ABSTAIN",
			}[vote.choice]
			fmt.Printf("  ✓ Vote cast: %s - %s (%d tokens)\n",
				vote.voter.PublicKey().String()[:16]+"...", choiceStr, vote.weight)
		}
	}

	// Display voting results
	proposal, _ = dao.GetProposal(proposalID)
	fmt.Printf("\nVoting Results:\n")
	fmt.Printf("  Yes: %d tokens\n", proposal.Results.YesVotes)
	fmt.Printf("  No: %d tokens\n", proposal.Results.NoVotes)
	fmt.Printf("  Abstain: %d tokens\n", proposal.Results.AbstainVotes)
	fmt.Printf("  Total Voters: %d\n", proposal.Results.TotalVoters)

	// Step 5: Finalize proposal and execute parameter changes
	fmt.Printf("\n=== Step 5: Finalizing and Executing Parameter Changes ===\n")

	// Manually set proposal as passed for demonstration
	totalActiveVotes := proposal.Results.YesVotes + proposal.Results.NoVotes
	if totalActiveVotes >= dao.GetParameterConfig().QuorumThreshold {
		passPercentage := (proposal.Results.YesVotes * 10000) / totalActiveVotes
		if passPercentage >= dao.GetParameterConfig().PassingThreshold {
			proposal.Status = ProposalStatusPassed
			proposal.Results.Passed = true
			fmt.Printf("✓ Proposal PASSED (%.1f%% approval)\n", float64(passPercentage)/100)
		} else {
			proposal.Status = ProposalStatusRejected
			proposal.Results.Passed = false
			fmt.Printf("✗ Proposal REJECTED (%.1f%% approval, needed %.1f%%)\n",
				float64(passPercentage)/100, float64(dao.GetParameterConfig().PassingThreshold)/100)
		}
	} else {
		proposal.Status = ProposalStatusRejected
		proposal.Results.Passed = false
		fmt.Printf("✗ Proposal REJECTED (quorum not met: %d < %d)\n",
			totalActiveVotes, dao.GetParameterConfig().QuorumThreshold)
	}

	if proposal.Status == ProposalStatusPassed {
		// Execute parameter changes
		fmt.Printf("\nExecuting parameter changes...\n")

		// Store original values for comparison
		originalConfig := dao.GetParameterConfig()
		originalValues := map[string]interface{}{
			"voting_period":      originalConfig.VotingPeriod,
			"quorum_threshold":   originalConfig.QuorumThreshold,
			"passing_threshold":  originalConfig.PassingThreshold,
			"treasury_threshold": originalConfig.TreasuryThreshold,
		}

		err = dao.ExecuteParameterChanges(proposalID, founder.PublicKey())
		if err != nil {
			fmt.Printf("✗ Failed to execute parameter changes: %v\n", err)
		} else {
			fmt.Printf("✓ Parameter changes executed successfully\n")

			// Display changes
			newConfig := dao.GetParameterConfig()
			fmt.Printf("\nParameter Changes Applied:\n")

			changes := map[string]struct{ old, new interface{} }{
				"voting_period":      {originalValues["voting_period"], newConfig.VotingPeriod},
				"quorum_threshold":   {originalValues["quorum_threshold"], newConfig.QuorumThreshold},
				"passing_threshold":  {originalValues["passing_threshold"], newConfig.PassingThreshold},
				"treasury_threshold": {originalValues["treasury_threshold"], newConfig.TreasuryThreshold},
			}

			for param, change := range changes {
				if change.old != change.new {
					fmt.Printf("  %s: %v → %v\n", param, change.old, change.new)
				}
			}
		}
	}

	// Step 6: Display parameter change history
	fmt.Printf("\n=== Step 6: Parameter Change History ===\n")

	allHistory := dao.GetAllParameterHistory()
	if len(allHistory) == 0 {
		fmt.Printf("No parameter changes recorded yet\n")
	} else {
		fmt.Printf("Parameter change history:\n")
		for param, history := range allHistory {
			fmt.Printf("\n%s:\n", param)
			for i, change := range history {
				fmt.Printf("  %d. %v → %v\n", i+1, change.OldValue, change.NewValue)
				fmt.Printf("     Changed by: %s\n", change.ChangedBy.String()[:16]+"...")
				fmt.Printf("     Reason: %s\n", change.Reason)
				fmt.Printf("     Date: %s\n", time.Unix(change.ChangedAt, 0).Format("2006-01-02 15:04:05"))
			}
		}
	}

	// Step 7: Demonstrate parameter constraints and validation
	fmt.Printf("\n=== Step 7: Advanced Parameter Management ===\n")

	// Test parameter change restrictions
	fmt.Printf("Testing parameter change restrictions:\n")

	restrictionTests := []struct {
		param    string
		value    interface{}
		expected bool
		desc     string
	}{
		{"min_proposal_threshold", uint64(30000), false, "threshold too high (>50% supply)"},
		{"min_proposal_threshold", uint64(2000), true, "reasonable threshold"},
		{"voting_period", int64(1800), false, "voting period too short"},
		{"voting_period", int64(259200), true, "reasonable voting period"},
		{"passing_threshold", uint64(15000), false, "passing threshold >100%"},
		{"passing_threshold", uint64(7500), true, "reasonable passing threshold"},
	}

	for _, test := range restrictionTests {
		allowed, reason := dao.IsParameterChangeAllowed(test.param, test.value)
		status := "✓"
		if allowed != test.expected {
			status = "✗"
		}

		fmt.Printf("  %s %s: %s", status, test.desc, test.param)
		if !allowed {
			fmt.Printf(" (blocked: %s)", reason)
		}
		fmt.Println()
	}

	// Display final configuration
	fmt.Printf("\n=== Final Configuration Summary ===\n")
	_ = dao.GetParameterConfig() // Get config for potential future use
	allParams := dao.ListAllParameters()

	fmt.Printf("Current DAO Parameters:\n")
	keyParameters := []string{
		"min_proposal_threshold", "voting_period", "quorum_threshold",
		"passing_threshold", "treasury_threshold", "delegation_enabled",
		"token_burning_enabled", "reputation_enabled",
	}

	for _, param := range keyParameters {
		if value, exists := allParams[param]; exists {
			fmt.Printf("  %s: %v\n", param, value)
		}
	}

	fmt.Printf("\nTreasury Status:\n")
	fmt.Printf("  Balance: %d tokens\n", dao.GetTreasuryBalance())
	fmt.Printf("  Required Signatures: %d\n", dao.GetRequiredSignatures())
	fmt.Printf("  Authorized Signers: %d\n", len(dao.GetTreasurySigners()))

	fmt.Printf("\nToken Status:\n")
	fmt.Printf("  Total Supply: %d tokens\n", dao.GetTotalSupply())
	fmt.Printf("  Active Token Holders: %d\n", len(dao.GovernanceState.TokenHolders))

	fmt.Printf("\nGovernance Status:\n")
	fmt.Printf("  Total Proposals: %d\n", len(dao.ListAllProposals()))
	fmt.Printf("  Active Proposals: %d\n", len(dao.ListActiveProposals()))
	fmt.Printf("  Parameter Changes: %d\n", len(allHistory))

	fmt.Printf("\n=== Parameter Management Example Complete ===\n")
}

// QuickParameterChangeExample demonstrates a simple parameter change
func QuickParameterChangeExample() {
	fmt.Println("=== Quick Parameter Change Example ===")

	// Setup
	dao := NewDAO("QUICK", "Quick Token", 18)
	admin := crypto.GeneratePrivateKey()

	// Initialize
	distributions := map[string]uint64{
		admin.PublicKey().String(): 10000,
	}
	dao.InitialTokenDistribution(distributions)

	fmt.Printf("Initial voting period: %d seconds\n", dao.GetParameterConfig().VotingPeriod)

	// Create and execute parameter change
	changes := map[string]interface{}{
		"voting_period": int64(172800), // 48 hours
	}

	proposalID, err := dao.CreateParameterProposal(
		admin.PublicKey(),
		changes,
		"Quick voting period adjustment",
		time.Now().Unix()+3600,
		VotingTypeSimple,
		time.Now().Unix()-600,
		time.Now().Unix()-300,
		1000,
	)

	if err != nil {
		fmt.Printf("Error creating proposal: %v\n", err)
		return
	}

	// Simulate passed proposal
	proposal, _ := dao.GetProposal(proposalID)
	proposal.Status = ProposalStatusPassed
	proposal.Results.Passed = true

	// Execute changes
	err = dao.ExecuteParameterChanges(proposalID, admin.PublicKey())
	if err != nil {
		fmt.Printf("Error executing changes: %v\n", err)
		return
	}

	fmt.Printf("New voting period: %d seconds\n", dao.GetParameterConfig().VotingPeriod)
	fmt.Printf("✓ Parameter change completed successfully\n")
}
