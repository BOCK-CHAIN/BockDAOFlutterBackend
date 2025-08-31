package dao

import (
	"fmt"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// ExampleDAOUsage demonstrates how to use the DAO system
func ExampleDAOUsage() {
	fmt.Println("=== ProjectX DAO Example ===")

	// 1. Create a new DAO
	dao := NewDAO("PXGOV", "ProjectX Governance Token", 18)
	fmt.Println("✓ Created new DAO with governance token PXGOV")

	// 2. Set up initial token distribution
	founder1 := crypto.GeneratePrivateKey()
	founder2 := crypto.GeneratePrivateKey()
	community := crypto.GeneratePrivateKey()

	distributions := map[string]uint64{
		founder1.PublicKey().String():  10000, // 10,000 tokens
		founder2.PublicKey().String():  8000,  // 8,000 tokens
		community.PublicKey().String(): 5000,  // 5,000 tokens
	}

	err := dao.InitialTokenDistribution(distributions)
	if err != nil {
		fmt.Printf("✗ Failed to distribute tokens: %v\n", err)
		return
	}
	fmt.Printf("✓ Distributed %d total tokens to %d addresses\n", dao.GetTotalSupply(), len(distributions))

	// 3. Initialize treasury with multi-sig
	treasurySigners := []crypto.PublicKey{
		founder1.PublicKey(),
		founder2.PublicKey(),
	}
	err = dao.InitializeTreasury(treasurySigners, 2) // Require both signatures
	if err != nil {
		fmt.Printf("✗ Failed to initialize treasury: %v\n", err)
		return
	}
	fmt.Println("✓ Initialized treasury with 2-of-2 multi-signature")

	// 4. Add some funds to treasury
	dao.AddTreasuryFunds(50000) // 50,000 units
	fmt.Printf("✓ Added %d units to treasury\n", dao.GetTreasuryBalance())

	// 5. Create a governance proposal
	proposalTx := &ProposalTx{
		Fee:          100,
		Title:        "Upgrade Protocol to v2.0",
		Description:  "This proposal suggests upgrading the ProjectX protocol to version 2.0 with enhanced features including improved consensus mechanism and better scalability.",
		ProposalType: ProposalTypeTechnical,
		VotingType:   VotingTypeSimple,
		StartTime:    time.Now().Unix(),
		EndTime:      time.Now().Unix() + 86400, // 24 hours
		Threshold:    5100,                      // 51%
		MetadataHash: types.Hash{},              // Empty for example
	}

	// Generate a mock transaction hash
	proposalHash := types.Hash{}
	copy(proposalHash[:], "proposal_hash_example_123456")

	err = dao.Processor.ProcessProposalTx(proposalTx, founder1.PublicKey(), proposalHash)
	if err != nil {
		fmt.Printf("✗ Failed to create proposal: %v\n", err)
		return
	}
	fmt.Printf("✓ Created proposal: '%s'\n", proposalTx.Title)

	// 6. Update proposal status to active
	dao.Processor.UpdateProposalStatus(proposalHash)
	proposal, _ := dao.GetProposal(proposalHash)
	fmt.Printf("✓ Proposal status: %d (Active=2)\n", proposal.Status)

	// 7. Cast votes
	// Founder2 votes YES
	voteTx1 := &VoteTx{
		Fee:        50,
		ProposalID: proposalHash,
		Choice:     VoteChoiceYes,
		Weight:     3000,
		Reason:     "I support this upgrade for better scalability",
	}
	err = dao.Processor.ProcessVoteTx(voteTx1, founder2.PublicKey())
	if err != nil {
		fmt.Printf("✗ Failed to process vote: %v\n", err)
		return
	}
	fmt.Printf("✓ Founder2 voted YES with weight %d\n", voteTx1.Weight)

	// Community votes NO
	voteTx2 := &VoteTx{
		Fee:        50,
		ProposalID: proposalHash,
		Choice:     VoteChoiceNo,
		Weight:     2000,
		Reason:     "Need more testing before upgrade",
	}
	err = dao.Processor.ProcessVoteTx(voteTx2, community.PublicKey())
	if err != nil {
		fmt.Printf("✗ Failed to process vote: %v\n", err)
		return
	}
	fmt.Printf("✓ Community voted NO with weight %d\n", voteTx2.Weight)

	// 8. Check voting results
	proposal, _ = dao.GetProposal(proposalHash)
	fmt.Printf("✓ Voting results: YES=%d, NO=%d, ABSTAIN=%d\n",
		proposal.Results.YesVotes,
		proposal.Results.NoVotes,
		proposal.Results.AbstainVotes)

	// 9. Demonstrate delegation
	delegationTx := &DelegationTx{
		Fee:      25,
		Delegate: founder1.PublicKey(),
		Duration: 86400 * 7, // 7 days
	}
	err = dao.Processor.ProcessDelegationTx(delegationTx, community.PublicKey())
	if err != nil {
		fmt.Printf("✗ Failed to create delegation: %v\n", err)
		return
	}
	fmt.Println("✓ Community delegated voting power to Founder1")

	// Check effective voting power
	effectivePower := dao.Processor.GetEffectiveVotingPower(founder1.PublicKey())
	fmt.Printf("✓ Founder1's effective voting power: %d (includes delegated power)\n", effectivePower)

	// 10. Demonstrate token operations
	// Mint new tokens
	mintTx := &TokenMintTx{
		Fee:       100,
		Recipient: community.PublicKey(),
		Amount:    1000,
		Reason:    "Community contribution reward",
	}
	err = dao.Processor.ProcessTokenMintTx(mintTx, founder1.PublicKey())
	if err != nil {
		fmt.Printf("✗ Failed to mint tokens: %v\n", err)
		return
	}
	fmt.Printf("✓ Minted %d tokens for community\n", mintTx.Amount)

	// 11. Show final balances
	fmt.Println("\n=== Final Token Balances ===")
	fmt.Printf("Founder1: %d tokens\n", dao.GetTokenBalance(founder1.PublicKey()))
	fmt.Printf("Founder2: %d tokens\n", dao.GetTokenBalance(founder2.PublicKey()))
	fmt.Printf("Community: %d tokens\n", dao.GetTokenBalance(community.PublicKey()))
	fmt.Printf("Total Supply: %d tokens\n", dao.GetTotalSupply())
	fmt.Printf("Treasury Balance: %d units\n", dao.GetTreasuryBalance())

	// 12. Show DAO statistics
	fmt.Println("\n=== DAO Statistics ===")
	allProposals := dao.ListAllProposals()
	activeProposals := dao.ListActiveProposals()
	fmt.Printf("Total Proposals: %d\n", len(allProposals))
	fmt.Printf("Active Proposals: %d\n", len(activeProposals))

	fmt.Println("\n✓ DAO example completed successfully!")
}
