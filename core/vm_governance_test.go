package core

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/dao"
	"github.com/BOCK-CHAIN/BockChain/types"
)

func TestVMGovernanceInstructions(t *testing.T) {
	// Setup test environment
	state := NewState()
	governanceState := dao.NewGovernanceState()
	privateKey := crypto.GeneratePrivateKey()
	publicKey := privateKey.PublicKey()

	// Use fixed timestamp for consistent testing
	testTimestamp := int64(1000000)

	// Create VM with governance support
	vm := NewVMWithGovernanceAndTimestamp([]byte{}, state, governanceState, publicKey, testTimestamp)

	t.Run("CreateProposal", func(t *testing.T) {
		// Setup stack for proposal creation
		vm.stack.Push("Test Proposal")
		vm.stack.Push("This is a test proposal description")
		vm.stack.Push(dao.ProposalTypeGeneral)
		vm.stack.Push(dao.VotingTypeSimple)
		vm.stack.Push(testTimestamp + 100)     // Start time
		vm.stack.Push(testTimestamp + 1000)    // End time
		vm.stack.Push(uint64(5000))            // Threshold
		vm.stack.Push([]byte("metadata-hash")) // Metadata hash

		// Execute create proposal instruction
		err := vm.Exec(InstrCreateProposal)
		if err != nil {
			t.Fatalf("Failed to create proposal: %v", err)
		}

		// Verify proposal was created
		if len(vm.governanceState.Proposals) != 1 {
			t.Fatalf("Expected 1 proposal, got %d", len(vm.governanceState.Proposals))
		}

		// Verify stack contains proposal ID
		proposalIDBytes := vm.stack.Pop().([]byte)
		if len(proposalIDBytes) == 0 {
			t.Fatal("Expected proposal ID on stack")
		}
	})

	t.Run("CastVote", func(t *testing.T) {
		// First create a proposal
		vm.stack.Push("Vote Test Proposal")
		vm.stack.Push("Test proposal for voting")
		vm.stack.Push(dao.ProposalTypeGeneral)
		vm.stack.Push(dao.VotingTypeSimple)
		vm.stack.Push(testTimestamp - 10)   // Start time (in past)
		vm.stack.Push(testTimestamp + 1000) // End time (in future)
		vm.stack.Push(uint64(5000))
		vm.stack.Push([]byte("metadata-hash"))

		err := vm.Exec(InstrCreateProposal)
		if err != nil {
			t.Fatalf("Failed to create proposal for voting test: %v", err)
		}

		proposalIDBytes := vm.stack.Pop().([]byte)

		// Cast a vote
		vm.stack.Push(proposalIDBytes)
		vm.stack.Push(dao.VoteChoiceYes)
		vm.stack.Push(uint64(100)) // Weight
		vm.stack.Push("I support this proposal")

		err = vm.Exec(InstrCastVote)
		if err != nil {
			t.Fatalf("Failed to cast vote: %v", err)
		}

		// Verify vote was recorded
		success := vm.stack.Pop().(bool)
		if !success {
			t.Fatal("Expected vote to be successful")
		}

		// Verify vote exists in governance state
		var proposalID types.Hash
		copy(proposalID[:], proposalIDBytes)

		voterKey := string(vm.caller[:])
		if _, exists := vm.governanceState.Votes[proposalID][voterKey]; !exists {
			t.Fatal("Vote was not recorded in governance state")
		}
	})

	t.Run("Delegate", func(t *testing.T) {
		// Create delegate keypair
		delegatePrivateKey := crypto.GeneratePrivateKey()
		delegatePublicKey := delegatePrivateKey.PublicKey()

		// Setup delegation
		vm.stack.Push([]byte(delegatePublicKey))
		vm.stack.Push(int64(86400)) // 24 hours
		vm.stack.Push(false)        // Not revoking

		err := vm.Exec(InstrDelegate)
		if err != nil {
			t.Fatalf("Failed to delegate: %v", err)
		}

		// Verify delegation was successful
		success := vm.stack.Pop().(bool)
		if !success {
			t.Fatal("Expected delegation to be successful")
		}

		// Verify delegation exists
		delegatorKey := string(vm.caller[:])
		delegation, exists := vm.governanceState.Delegations[delegatorKey]
		if !exists {
			t.Fatal("Delegation was not recorded")
		}

		if !delegation.Active {
			t.Fatal("Delegation should be active")
		}
	})

	t.Run("CalculateQuorum", func(t *testing.T) {
		// Create a proposal with votes
		vm.stack.Push("Quorum Test Proposal")
		vm.stack.Push("Test proposal for quorum calculation")
		vm.stack.Push(dao.ProposalTypeGeneral)
		vm.stack.Push(dao.VotingTypeSimple)
		vm.stack.Push(testTimestamp + 10)
		vm.stack.Push(testTimestamp + 1000)
		vm.stack.Push(uint64(5000))
		vm.stack.Push([]byte("metadata-hash"))

		err := vm.Exec(InstrCreateProposal)
		if err != nil {
			t.Fatalf("Failed to create proposal: %v", err)
		}

		proposalIDBytes := vm.stack.Pop().([]byte)
		var proposalID types.Hash
		copy(proposalID[:], proposalIDBytes)

		// Manually add votes to meet quorum
		proposal := vm.governanceState.Proposals[proposalID]
		proposal.Results.YesVotes = 1500
		proposal.Results.NoVotes = 800
		proposal.Results.AbstainVotes = 200

		// Calculate quorum
		vm.stack.Push(proposalIDBytes)
		err = vm.Exec(InstrCalculateQuorum)
		if err != nil {
			t.Fatalf("Failed to calculate quorum: %v", err)
		}

		// Check if quorum is met
		quorumMet := vm.stack.Pop().(bool)
		if !quorumMet {
			t.Fatal("Expected quorum to be met")
		}
	})

	t.Run("QuadraticVote", func(t *testing.T) {
		// Create a quadratic voting proposal
		vm.stack.Push("Quadratic Vote Test")
		vm.stack.Push("Test quadratic voting")
		vm.stack.Push(dao.ProposalTypeGeneral)
		vm.stack.Push(dao.VotingTypeQuadratic)
		vm.stack.Push(testTimestamp - 10)   // Start time (in past)
		vm.stack.Push(testTimestamp + 1000) // End time (in future)
		vm.stack.Push(uint64(5000))
		vm.stack.Push([]byte("metadata-hash"))

		err := vm.Exec(InstrCreateProposal)
		if err != nil {
			t.Fatalf("Failed to create quadratic proposal: %v", err)
		}

		proposalIDBytes := vm.stack.Pop().([]byte)

		// Cast quadratic vote
		vm.stack.Push(proposalIDBytes)
		vm.stack.Push(dao.VoteChoiceYes)
		vm.stack.Push(uint64(5)) // Vote count (cost will be 25)
		vm.stack.Push("Quadratic vote reason")

		err = vm.Exec(InstrQuadraticVote)
		if err != nil {
			t.Fatalf("Failed to cast quadratic vote: %v", err)
		}

		// Check results - stack is actually FIFO, so tokenCost comes first, then success
		tokenCost := vm.stack.Pop().(uint64)
		success := vm.stack.Pop().(bool)

		if !success {
			t.Fatal("Expected quadratic vote to be successful")
		}

		if tokenCost != 25 { // 5^2 = 25
			t.Fatalf("Expected token cost of 25, got %d", tokenCost)
		}
	})

	t.Run("MintTokens", func(t *testing.T) {
		recipientPrivateKey := crypto.GeneratePrivateKey()
		recipientPublicKey := recipientPrivateKey.PublicKey()

		vm.stack.Push([]byte(recipientPublicKey))
		vm.stack.Push(uint64(1000))
		vm.stack.Push("Initial token allocation")

		err := vm.Exec(InstrMintTokens)
		if err != nil {
			t.Fatalf("Failed to mint tokens: %v", err)
		}

		success := vm.stack.Pop().(bool)
		if !success {
			t.Fatal("Expected token minting to be successful")
		}

		// Verify tokens were minted
		recipientKey := string(recipientPublicKey[:])
		holder, exists := vm.governanceState.TokenHolders[recipientKey]
		if !exists {
			t.Fatal("Token holder was not created")
		}

		if holder.Balance != 1000 {
			t.Fatalf("Expected balance of 1000, got %d", holder.Balance)
		}
	})

	t.Run("BurnTokens", func(t *testing.T) {
		// First mint tokens to caller
		callerKey := string(vm.caller[:])
		vm.governanceState.TokenHolders[callerKey] = &dao.TokenHolder{
			Address:    vm.caller,
			Balance:    500,
			Staked:     0,
			Reputation: 0,
			JoinedAt:   time.Now().Unix(),
			LastActive: time.Now().Unix(),
		}

		vm.stack.Push(uint64(200))
		vm.stack.Push("Burning excess tokens")

		err := vm.Exec(InstrBurnTokens)
		if err != nil {
			t.Fatalf("Failed to burn tokens: %v", err)
		}

		success := vm.stack.Pop().(bool)
		if !success {
			t.Fatal("Expected token burning to be successful")
		}

		// Verify tokens were burned
		holder := vm.governanceState.TokenHolders[callerKey]
		if holder.Balance != 300 {
			t.Fatalf("Expected balance of 300 after burning, got %d", holder.Balance)
		}
	})

	t.Run("GetProposal", func(t *testing.T) {
		// Create a proposal first
		vm.stack.Push("Get Proposal Test")
		vm.stack.Push("Test proposal retrieval")
		vm.stack.Push(dao.ProposalTypeGeneral)
		vm.stack.Push(dao.VotingTypeSimple)
		vm.stack.Push(time.Now().Unix() + 100)
		vm.stack.Push(time.Now().Unix() + 1000)
		vm.stack.Push(uint64(5000))
		vm.stack.Push([]byte("metadata-hash"))

		err := vm.Exec(InstrCreateProposal)
		if err != nil {
			t.Fatalf("Failed to create proposal: %v", err)
		}

		proposalIDBytes := vm.stack.Pop().([]byte)

		// Get the proposal
		vm.stack.Push(proposalIDBytes)
		err = vm.Exec(InstrGetProposal)
		if err != nil {
			t.Fatalf("Failed to get proposal: %v", err)
		}

		// Verify proposal data
		proposalData := vm.stack.Pop().([]byte)
		if proposalData == nil {
			t.Fatal("Expected proposal data, got nil")
		}

		var proposal dao.Proposal
		err = json.Unmarshal(proposalData, &proposal)
		if err != nil {
			t.Fatalf("Failed to unmarshal proposal data: %v", err)
		}

		if proposal.Title != "Get Proposal Test" {
			t.Fatalf("Expected title 'Get Proposal Test', got '%s'", proposal.Title)
		}
	})

	t.Run("TreasuryTransfer", func(t *testing.T) {
		// Setup treasury with funds
		vm.governanceState.Treasury.Balance = 10000

		recipientPrivateKey := crypto.GeneratePrivateKey()
		recipientPublicKey := recipientPrivateKey.PublicKey()
		signatures := []crypto.Signature{} // Empty for test
		signaturesBytes, _ := json.Marshal(signatures)

		vm.stack.Push([]byte(recipientPublicKey[:]))
		vm.stack.Push(uint64(1000))
		vm.stack.Push("Test treasury transfer")
		vm.stack.Push(signaturesBytes)
		vm.stack.Push(uint8(0)) // No signatures required for test

		err := vm.Exec(InstrTreasuryTransfer)
		if err != nil {
			t.Fatalf("Failed to execute treasury transfer: %v", err)
		}

		// Verify transfer
		txIDBytes := vm.stack.Pop().([]byte)
		if len(txIDBytes) == 0 {
			t.Fatal("Expected transaction ID")
		}

		// Verify treasury balance was reduced
		if vm.governanceState.Treasury.Balance != 9000 {
			t.Fatalf("Expected treasury balance of 9000, got %d", vm.governanceState.Treasury.Balance)
		}
	})
}

func TestVMGovernanceErrorHandling(t *testing.T) {
	state := NewState()
	governanceState := dao.NewGovernanceState()
	privateKey := crypto.GeneratePrivateKey()
	publicKey := privateKey.PublicKey()
	testTimestamp := int64(1000000)
	vm := NewVMWithGovernanceAndTimestamp([]byte{}, state, governanceState, publicKey, testTimestamp)

	t.Run("VoteOnNonexistentProposal", func(t *testing.T) {
		// Try to vote on non-existent proposal
		fakeProposalID := types.Hash{}
		copy(fakeProposalID[:], "fake-proposal-id")

		vm.stack.Push(fakeProposalID[:])
		vm.stack.Push(dao.VoteChoiceYes)
		vm.stack.Push(uint64(100))
		vm.stack.Push("Vote on fake proposal")

		err := vm.Exec(InstrCastVote)
		if err == nil {
			t.Fatal("Expected error when voting on non-existent proposal")
		}

		if err != dao.ErrProposalNotFoundError {
			t.Fatalf("Expected ErrProposalNotFoundError, got %v", err)
		}
	})

	t.Run("DuplicateVote", func(t *testing.T) {
		// Create proposal
		vm.stack.Push("Duplicate Vote Test")
		vm.stack.Push("Test duplicate voting")
		vm.stack.Push(dao.ProposalTypeGeneral)
		vm.stack.Push(dao.VotingTypeSimple)
		vm.stack.Push(testTimestamp - 100)  // Start time (in past)
		vm.stack.Push(testTimestamp + 1000) // End time (in future)
		vm.stack.Push(uint64(5000))
		vm.stack.Push([]byte("metadata-hash"))

		err := vm.Exec(InstrCreateProposal)
		if err != nil {
			t.Fatalf("Failed to create proposal: %v", err)
		}

		proposalIDBytes := vm.stack.Pop().([]byte)

		// Cast first vote
		vm.stack.Push(proposalIDBytes)
		vm.stack.Push(dao.VoteChoiceYes)
		vm.stack.Push(uint64(100))
		vm.stack.Push("First vote")

		err = vm.Exec(InstrCastVote)
		if err != nil {
			t.Fatalf("Failed to cast first vote: %v", err)
		}

		vm.stack.Pop() // Remove success result

		// Try to cast second vote (should fail)
		vm.stack.Push(proposalIDBytes)
		vm.stack.Push(dao.VoteChoiceNo)
		vm.stack.Push(uint64(50))
		vm.stack.Push("Second vote")

		err = vm.Exec(InstrCastVote)
		if err == nil {
			t.Fatal("Expected error when casting duplicate vote")
		}

		if err != dao.ErrDuplicateVoteError {
			t.Fatalf("Expected ErrDuplicateVoteError, got %v", err)
		}
	})

	t.Run("InvalidTimeframe", func(t *testing.T) {
		// Try to create proposal with invalid timeframe
		vm.stack.Push("Invalid Timeframe Test")
		vm.stack.Push("Test invalid timeframe")
		vm.stack.Push(dao.ProposalTypeGeneral)
		vm.stack.Push(dao.VotingTypeSimple)
		vm.stack.Push(testTimestamp + 1000) // Start time after end time
		vm.stack.Push(testTimestamp + 100)  // End time before start time
		vm.stack.Push(uint64(5000))
		vm.stack.Push([]byte("metadata-hash"))

		err := vm.Exec(InstrCreateProposal)
		if err == nil {
			t.Fatal("Expected error for invalid timeframe")
		}

		if err != dao.ErrInvalidTimeframeError {
			t.Fatalf("Expected ErrInvalidTimeframeError, got %v", err)
		}
	})

	t.Run("InsufficientTreasuryFunds", func(t *testing.T) {
		// Set treasury balance to low amount
		vm.governanceState.Treasury.Balance = 100

		recipientPrivateKey := crypto.GeneratePrivateKey()
		recipientPublicKey := recipientPrivateKey.PublicKey()
		signatures := []crypto.Signature{}
		signaturesBytes, _ := json.Marshal(signatures)

		vm.stack.Push([]byte(recipientPublicKey[:]))
		vm.stack.Push(uint64(1000)) // More than treasury balance
		vm.stack.Push("Test insufficient funds")
		vm.stack.Push(signaturesBytes)
		vm.stack.Push(uint8(0))

		err := vm.Exec(InstrTreasuryTransfer)
		if err == nil {
			t.Fatal("Expected error for insufficient treasury funds")
		}

		if err != dao.ErrTreasuryInsufficientFunds {
			t.Fatalf("Expected ErrTreasuryInsufficientFunds, got %v", err)
		}
	})

	t.Run("BurnInsufficientTokens", func(t *testing.T) {
		// Set caller balance to low amount
		callerKey := string(vm.caller[:])
		vm.governanceState.TokenHolders[callerKey] = &dao.TokenHolder{
			Address: vm.caller,
			Balance: 50,
		}

		vm.stack.Push(uint64(100)) // More than balance
		vm.stack.Push("Test insufficient tokens")

		err := vm.Exec(InstrBurnTokens)
		if err == nil {
			t.Fatal("Expected error for insufficient tokens to burn")
		}

		if err != dao.ErrInsufficientTokensForVote {
			t.Fatalf("Expected ErrInsufficientTokensForVote, got %v", err)
		}
	})
}

func TestVMGovernanceIntegration(t *testing.T) {
	state := NewState()
	governanceState := dao.NewGovernanceState()
	privateKey := crypto.GeneratePrivateKey()
	publicKey := privateKey.PublicKey()
	testTimestamp := int64(1000000)
	vm := NewVMWithGovernanceAndTimestamp([]byte{}, state, governanceState, publicKey, testTimestamp)

	t.Run("CompleteProposalLifecycle", func(t *testing.T) {
		// 1. Create proposal
		vm.stack.Push("Integration Test Proposal")
		vm.stack.Push("Complete lifecycle test")
		vm.stack.Push(dao.ProposalTypeGeneral)
		vm.stack.Push(dao.VotingTypeSimple)
		vm.stack.Push(testTimestamp - 100)  // Start time (in past)
		vm.stack.Push(testTimestamp + 1000) // End time (in future)
		vm.stack.Push(uint64(1000))
		vm.stack.Push([]byte("metadata-hash"))

		err := vm.Exec(InstrCreateProposal)
		if err != nil {
			t.Fatalf("Failed to create proposal: %v", err)
		}

		proposalIDBytes := vm.stack.Pop().([]byte)

		// 2. Cast votes (need to meet quorum threshold of 2000)
		vm.stack.Push(proposalIDBytes)
		vm.stack.Push(dao.VoteChoiceYes)
		vm.stack.Push(uint64(2500)) // More than quorum threshold
		vm.stack.Push("I support this")

		err = vm.Exec(InstrCastVote)
		if err != nil {
			t.Fatalf("Failed to cast vote: %v", err)
		}

		vm.stack.Pop() // Remove success result

		// 3. Calculate quorum
		vm.stack.Push(proposalIDBytes)
		err = vm.Exec(InstrCalculateQuorum)
		if err != nil {
			t.Fatalf("Failed to calculate quorum: %v", err)
		}

		quorumMet := vm.stack.Pop().(bool)
		if !quorumMet {
			t.Fatal("Expected quorum to be met")
		}

		// 4. Execute proposal
		vm.stack.Push(proposalIDBytes)
		err = vm.Exec(InstrExecuteProposal)
		if err != nil {
			t.Fatalf("Failed to execute proposal: %v", err)
		}

		success := vm.stack.Pop().(bool)
		if !success {
			t.Fatal("Expected proposal execution to succeed")
		}

		// 5. Verify proposal status
		var proposalID types.Hash
		copy(proposalID[:], proposalIDBytes)

		proposal := vm.governanceState.Proposals[proposalID]
		if proposal.Status != dao.ProposalStatusExecuted {
			t.Fatalf("Expected proposal status to be executed, got %v", proposal.Status)
		}

		if !proposal.Results.Passed {
			t.Fatal("Expected proposal to be marked as passed")
		}
	})

	t.Run("DelegationAndVoting", func(t *testing.T) {
		// Create delegate
		delegatePrivateKey := crypto.GeneratePrivateKey()
		delegatePublicKey := delegatePrivateKey.PublicKey()
		delegateVM := NewVMWithGovernanceAndTimestamp([]byte{}, state, governanceState, delegatePublicKey, testTimestamp)

		// 1. Delegate voting power
		vm.stack.Push(delegatePublicKey[:])
		vm.stack.Push(int64(86400))
		vm.stack.Push(false)

		err := vm.Exec(InstrDelegate)
		if err != nil {
			t.Fatalf("Failed to delegate: %v", err)
		}

		vm.stack.Pop() // Remove success result

		// 2. Create proposal
		delegateVM.stack.Push("Delegation Test Proposal")
		delegateVM.stack.Push("Test delegated voting")
		delegateVM.stack.Push(dao.ProposalTypeGeneral)
		delegateVM.stack.Push(dao.VotingTypeSimple)
		delegateVM.stack.Push(testTimestamp - 100)  // Start time (in past)
		delegateVM.stack.Push(testTimestamp + 1000) // End time (in future)
		delegateVM.stack.Push(uint64(1000))
		delegateVM.stack.Push([]byte("metadata-hash"))

		err = delegateVM.Exec(InstrCreateProposal)
		if err != nil {
			t.Fatalf("Failed to create proposal: %v", err)
		}

		proposalIDBytes := delegateVM.stack.Pop().([]byte)

		// 3. Delegate votes on behalf of delegator
		delegateVM.stack.Push(proposalIDBytes)
		delegateVM.stack.Push(dao.VoteChoiceYes)
		delegateVM.stack.Push(uint64(500))
		delegateVM.stack.Push("Voting as delegate")

		err = delegateVM.Exec(InstrCastVote)
		if err != nil {
			t.Fatalf("Failed to cast delegated vote: %v", err)
		}

		success := delegateVM.stack.Pop().(bool)
		if !success {
			t.Fatal("Expected delegated vote to be successful")
		}

		// 4. Verify delegation exists
		delegatorKey := string(vm.caller[:])
		delegation, exists := vm.governanceState.Delegations[delegatorKey]
		if !exists {
			t.Fatal("Delegation should exist")
		}

		if !delegation.Active {
			t.Fatal("Delegation should be active")
		}
	})
}
