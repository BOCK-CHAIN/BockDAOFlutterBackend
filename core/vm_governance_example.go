package core

import (
	"fmt"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/dao"
)

// GovernanceVMExample demonstrates how to use the VM governance instructions
func GovernanceVMExample() {
	// Create a new VM with governance support
	state := NewState()
	governanceState := dao.NewGovernanceState()

	// Create a user keypair
	privateKey := crypto.GeneratePrivateKey()
	publicKey := privateKey.PublicKey()

	// Initialize VM with governance state
	vm := NewVMWithGovernance([]byte{}, state, governanceState, publicKey)

	fmt.Println("=== ProjectX DAO VM Governance Instructions Example ===")

	// Example 1: Create a governance proposal
	fmt.Println("\n1. Creating a governance proposal...")

	vm.stack.Push("Increase Block Reward")
	vm.stack.Push("Proposal to increase block reward from 10 to 15 tokens")
	vm.stack.Push(dao.ProposalTypeTechnical)
	vm.stack.Push(dao.VotingTypeSimple)
	vm.stack.Push(time.Now().Unix() - 100)  // Start time (past)
	vm.stack.Push(time.Now().Unix() + 3600) // End time (1 hour from now)
	vm.stack.Push(uint64(1000))             // Threshold
	vm.stack.Push([]byte("ipfs-hash-123"))  // Metadata hash

	err := vm.Exec(InstrCreateProposal)
	if err != nil {
		fmt.Printf("Error creating proposal: %v\n", err)
		return
	}

	proposalIDBytes := vm.stack.Pop().([]byte)
	fmt.Printf("✓ Proposal created with ID: %x\n", proposalIDBytes)

	// Example 2: Cast a vote on the proposal
	fmt.Println("\n2. Casting a vote...")

	vm.stack.Push(proposalIDBytes)
	vm.stack.Push(dao.VoteChoiceYes)
	vm.stack.Push(uint64(500))
	vm.stack.Push("I support increasing the block reward")

	err = vm.Exec(InstrCastVote)
	if err != nil {
		fmt.Printf("Error casting vote: %v\n", err)
		return
	}

	success := vm.stack.Pop().(bool)
	if success {
		fmt.Println("✓ Vote cast successfully")
	}

	// Example 3: Mint governance tokens
	fmt.Println("\n3. Minting governance tokens...")

	vm.stack.Push([]byte(publicKey[:]))
	vm.stack.Push(uint64(1000))
	vm.stack.Push("Initial token allocation")

	err = vm.Exec(InstrMintTokens)
	if err != nil {
		fmt.Printf("Error minting tokens: %v\n", err)
		return
	}

	success = vm.stack.Pop().(bool)
	if success {
		fmt.Println("✓ Tokens minted successfully")
	}

	// Example 4: Delegate voting power
	fmt.Println("\n4. Delegating voting power...")

	// Create a delegate
	delegatePrivateKey := crypto.GeneratePrivateKey()
	delegatePublicKey := delegatePrivateKey.PublicKey()

	vm.stack.Push([]byte(delegatePublicKey[:]))
	vm.stack.Push(int64(86400)) // 24 hours
	vm.stack.Push(false)        // Not revoking

	err = vm.Exec(InstrDelegate)
	if err != nil {
		fmt.Printf("Error delegating: %v\n", err)
		return
	}

	success = vm.stack.Pop().(bool)
	if success {
		fmt.Println("✓ Voting power delegated successfully")
	}

	// Example 5: Quadratic voting
	fmt.Println("\n5. Creating quadratic voting proposal...")

	vm.stack.Push("Quadratic Vote Test")
	vm.stack.Push("Test quadratic voting mechanism")
	vm.stack.Push(dao.ProposalTypeGeneral)
	vm.stack.Push(dao.VotingTypeQuadratic)
	vm.stack.Push(time.Now().Unix() - 100)
	vm.stack.Push(time.Now().Unix() + 3600)
	vm.stack.Push(uint64(1000))
	vm.stack.Push([]byte("ipfs-hash-456"))

	err = vm.Exec(InstrCreateProposal)
	if err != nil {
		fmt.Printf("Error creating quadratic proposal: %v\n", err)
		return
	}

	quadraticProposalIDBytes := vm.stack.Pop().([]byte)
	fmt.Printf("✓ Quadratic proposal created with ID: %x\n", quadraticProposalIDBytes)

	// Cast quadratic vote
	fmt.Println("\n6. Casting quadratic vote...")

	vm.stack.Push(quadraticProposalIDBytes)
	vm.stack.Push(dao.VoteChoiceYes)
	vm.stack.Push(uint64(5)) // Vote count (cost will be 25)
	vm.stack.Push("Quadratic vote for better governance")

	err = vm.Exec(InstrQuadraticVote)
	if err != nil {
		fmt.Printf("Error casting quadratic vote: %v\n", err)
		return
	}

	tokenCost := vm.stack.Pop().(uint64)
	success = vm.stack.Pop().(bool)

	if success {
		fmt.Printf("✓ Quadratic vote cast successfully (cost: %d tokens)\n", tokenCost)
	}

	// Example 6: Treasury operations
	fmt.Println("\n7. Treasury operations...")

	// Set up treasury with funds
	vm.governanceState.Treasury.Balance = 10000

	recipientPrivateKey := crypto.GeneratePrivateKey()
	recipientPublicKey := recipientPrivateKey.PublicKey()

	vm.stack.Push([]byte(recipientPublicKey[:]))
	vm.stack.Push(uint64(1000))
	vm.stack.Push("Development grant")
	vm.stack.Push([]byte("[]")) // Empty signatures for example
	vm.stack.Push(uint8(0))     // No signatures required for example

	err = vm.Exec(InstrTreasuryTransfer)
	if err != nil {
		fmt.Printf("Error executing treasury transfer: %v\n", err)
		return
	}

	txIDBytes := vm.stack.Pop().([]byte)
	fmt.Printf("✓ Treasury transfer executed with TX ID: %x\n", txIDBytes)

	// Example 7: Get proposal information
	fmt.Println("\n8. Retrieving proposal information...")

	vm.stack.Push(proposalIDBytes)
	err = vm.Exec(InstrGetProposal)
	if err != nil {
		fmt.Printf("Error getting proposal: %v\n", err)
		return
	}

	proposalData := vm.stack.Pop().([]byte)
	if proposalData != nil {
		fmt.Printf("✓ Proposal data retrieved (%d bytes)\n", len(proposalData))
	}

	fmt.Println("\n=== All governance instructions executed successfully! ===")

	// Print final state summary
	fmt.Printf("\nFinal State Summary:\n")
	fmt.Printf("- Total proposals: %d\n", len(vm.governanceState.Proposals))
	fmt.Printf("- Total delegations: %d\n", len(vm.governanceState.Delegations))
	fmt.Printf("- Total token holders: %d\n", len(vm.governanceState.TokenHolders))
	fmt.Printf("- Treasury balance: %d\n", vm.governanceState.Treasury.Balance)
	fmt.Printf("- Treasury transactions: %d\n", len(vm.governanceState.Treasury.Transactions))
}
