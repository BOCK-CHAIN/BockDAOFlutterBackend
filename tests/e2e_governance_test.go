package tests

import (
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/core"
	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/dao"
	"github.com/BOCK-CHAIN/BockChain/types"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompleteGovernanceFlows tests end-to-end governance workflows
func TestCompleteGovernanceFlows(t *testing.T) {
	t.Run("CompleteProposalLifecycle", testCompleteProposalLifecycle)
	t.Run("DelegatedVotingFlow", testDelegatedVotingFlow)
	t.Run("TreasuryManagementFlow", testTreasuryManagementFlow)
	t.Run("QuadraticVotingFlow", testQuadraticVotingFlow)
	t.Run("MultiProposalConcurrentVoting", testMultiProposalConcurrentVoting)
}

func testCompleteProposalLifecycle(t *testing.T) {
	// Setup test blockchain
	bc, cleanup := setupTestBlockchain(t)
	defer cleanup()

	// Setup test users with tokens
	creator := crypto.GeneratePrivateKey()
	voters := make([]crypto.PrivateKey, 5)
	for i := range voters {
		voters[i] = crypto.GeneratePrivateKey()
	}

	setupTestUsers(t, bc, append(voters, creator)...)

	// Phase 1: Proposal Creation
	proposalTx := &core.Transaction{
		TxInner: dao.ProposalTx{
			Fee:          200,
			Title:        "E2E Test Proposal",
			Description:  "End-to-end testing of complete proposal lifecycle",
			ProposalType: dao.ProposalTypeGeneral,
			VotingType:   dao.VotingTypeSimple,
			StartTime:    time.Now().Unix() + 10,   // Start in 10 seconds
			EndTime:      time.Now().Unix() + 3600, // End in 1 hour
			Threshold:    5100,                     // 51%
			MetadataHash: randomHash(),
		},
		From: creator.PublicKey(),
	}
	proposalTx.Sign(creator)

	// Add proposal to blockchain
	block := createBlockWithTxs(t, bc, []*core.Transaction{proposalTx})
	err := bc.AddBlock(block)
	require.NoError(t, err)

	proposalID := proposalTx.Hash(core.TxHasher{})

	// Verify proposal creation
	proposal, err := bc.GetProposal(proposalID)
	require.NoError(t, err)
	assert.Equal(t, "E2E Test Proposal", proposal.Title)
	assert.Equal(t, dao.ProposalStatusPending, proposal.Status)

	// Phase 2: Wait for proposal to become active
	time.Sleep(11 * time.Second)
	err = bc.UpdateProposalStatuses()
	require.NoError(t, err)

	proposal, err = bc.GetProposal(proposalID)
	require.NoError(t, err)
	assert.Equal(t, dao.ProposalStatusActive, proposal.Status)

	// Phase 3: Voting
	var voteTxs []*core.Transaction
	for i, voter := range voters {
		choice := dao.VoteChoiceYes
		if i%2 == 0 {
			choice = dao.VoteChoiceNo
		}

		voteTx := &core.Transaction{
			TxInner: dao.VoteTx{
				Fee:        100,
				ProposalID: proposalID,
				Choice:     choice,
				Weight:     1000,
				Reason:     fmt.Sprintf("Vote from voter %d", i),
			},
			From: voter.PublicKey(),
		}
		voteTx.Sign(voter)
		voteTxs = append(voteTxs, voteTx)
	}

	// Add votes to blockchain
	voteBlock := createBlockWithTxs(t, bc, voteTxs)
	err = bc.AddBlock(voteBlock)
	require.NoError(t, err)

	// Phase 4: Verify voting results
	votes, err := bc.GetVotes(proposalID)
	require.NoError(t, err)
	assert.Len(t, votes, 5)

	proposal, err = bc.GetProposal(proposalID)
	require.NoError(t, err)

	// Verify vote tallies (3 No votes, 2 Yes votes)
	assert.Equal(t, uint64(3000), proposal.Results.NoVotes)
	assert.Equal(t, uint64(2000), proposal.Results.YesVotes)
	assert.Equal(t, uint64(5), proposal.Results.TotalVoters)

	t.Logf("Complete proposal lifecycle test passed - Proposal %s processed successfully", proposalID.String()[:8])
}

func testDelegatedVotingFlow(t *testing.T) {
	bc, cleanup := setupTestBlockchain(t)
	defer cleanup()

	// Setup delegator and delegate
	delegator := crypto.GeneratePrivateKey()
	delegate := crypto.GeneratePrivateKey()

	setupTestUsers(t, bc, delegator, delegate)

	// Create delegation
	delegationTx := &core.Transaction{
		TxInner: dao.DelegationTx{
			Fee:      100,
			Delegate: delegate.PublicKey(),
			Duration: 3600,
			Revoke:   false,
		},
		From: delegator.PublicKey(),
	}
	delegationTx.Sign(delegator)

	block := createBlockWithTxs(t, bc, []*core.Transaction{delegationTx})
	err := bc.AddBlock(block)
	require.NoError(t, err)

	// Verify delegation
	delegation, err := bc.GetDelegation(delegator.PublicKey())
	require.NoError(t, err)
	assert.Equal(t, delegate.PublicKey(), delegation.Delegate)
	assert.True(t, delegation.Active)

	// Test effective voting power
	processor := bc.GetDAOProcessor()
	delegatorPower := processor.GetEffectiveVotingPower(delegator.PublicKey())
	delegatePower := processor.GetEffectiveVotingPower(delegate.PublicKey())

	assert.Equal(t, uint64(0), delegatorPower)      // Delegated away
	assert.Greater(t, delegatePower, uint64(10000)) // Has own + delegated power

	t.Log("Delegated voting flow test passed")
}

func testTreasuryManagementFlow(t *testing.T) {
	bc, cleanup := setupTestBlockchain(t)
	defer cleanup()

	// Setup treasury signers
	signer1 := crypto.GeneratePrivateKey()
	signer2 := crypto.GeneratePrivateKey()
	recipient := crypto.GeneratePrivateKey()

	setupTestUsers(t, bc, signer1, signer2, recipient)

	// Initialize treasury
	treasuryState := bc.GetTreasuryState()
	treasuryState.Balance = 100000
	treasuryState.Signers = []crypto.PublicKey{signer1.PublicKey(), signer2.PublicKey()}
	treasuryState.RequiredSigs = 2

	// Create treasury transaction with signatures
	txData := []byte("treasury-disbursement-data")
	sig1, err := signer1.Sign(txData)
	require.NoError(t, err)
	sig2, err := signer2.Sign(txData)
	require.NoError(t, err)

	treasuryTx := &core.Transaction{
		TxInner: dao.TreasuryTx{
			Fee:          200,
			Recipient:    recipient.PublicKey(),
			Amount:       10000,
			Purpose:      "Development grant disbursement",
			Signatures:   []crypto.Signature{*sig1, *sig2},
			RequiredSigs: 2,
		},
		From: signer1.PublicKey(),
	}
	treasuryTx.Sign(signer1)

	block := createBlockWithTxs(t, bc, []*core.Transaction{treasuryTx})
	err = bc.AddBlock(block)
	require.NoError(t, err)

	// Verify treasury transaction
	updatedTreasury := bc.GetTreasuryState()
	assert.Equal(t, uint64(90000), updatedTreasury.Balance) // 100000 - 10000

	// Verify recipient received tokens
	recipientBalance := bc.GetTokenBalance(recipient.PublicKey())
	assert.Greater(t, recipientBalance, uint64(10000))

	t.Log("Treasury management flow test passed")
}

func testQuadraticVotingFlow(t *testing.T) {
	bc, cleanup := setupTestBlockchain(t)
	defer cleanup()

	creator := crypto.GeneratePrivateKey()
	voter := crypto.GeneratePrivateKey()

	setupTestUsers(t, bc, creator, voter)

	// Create quadratic voting proposal
	proposalTx := &core.Transaction{
		TxInner: dao.ProposalTx{
			Fee:          200,
			Title:        "Quadratic Voting Test",
			Description:  "Testing quadratic voting mechanism",
			ProposalType: dao.ProposalTypeGeneral,
			VotingType:   dao.VotingTypeQuadratic,
			StartTime:    time.Now().Unix() - 100,
			EndTime:      time.Now().Unix() + 3600,
			Threshold:    5100,
			MetadataHash: randomHash(),
		},
		From: creator.PublicKey(),
	}
	proposalTx.Sign(creator)

	block := createBlockWithTxs(t, bc, []*core.Transaction{proposalTx})
	err := bc.AddBlock(block)
	require.NoError(t, err)

	proposalID := proposalTx.Hash(core.TxHasher{})

	// Cast quadratic vote
	initialBalance := bc.GetTokenBalance(voter.PublicKey())

	quadraticVoteTx := &core.Transaction{
		TxInner: dao.VoteTx{
			Fee:        100,
			ProposalID: proposalID,
			Choice:     dao.VoteChoiceYes,
			Weight:     15, // Cost will be 15^2 = 225 tokens
			Reason:     "Quadratic vote test",
		},
		From: voter.PublicKey(),
	}
	quadraticVoteTx.Sign(voter)

	voteBlock := createBlockWithTxs(t, bc, []*core.Transaction{quadraticVoteTx})
	err = bc.AddBlock(voteBlock)
	require.NoError(t, err)

	// Verify quadratic cost
	newBalance := bc.GetTokenBalance(voter.PublicKey())
	expectedCost := uint64(225 + 100) // 15^2 + fee
	assert.Equal(t, initialBalance-expectedCost, newBalance)

	// Verify vote weight
	votes, err := bc.GetVotes(proposalID)
	require.NoError(t, err)
	vote := votes[voter.PublicKey().String()]
	assert.Equal(t, uint64(15), vote.Weight)

	t.Log("Quadratic voting flow test passed")
}

func testMultiProposalConcurrentVoting(t *testing.T) {
	bc, cleanup := setupTestBlockchain(t)
	defer cleanup()

	creator := crypto.GeneratePrivateKey()
	voters := make([]crypto.PrivateKey, 10)
	for i := range voters {
		voters[i] = crypto.GeneratePrivateKey()
	}

	setupTestUsers(t, bc, append(voters, creator)...)

	// Create multiple proposals
	var proposalTxs []*core.Transaction
	var proposalIDs []types.Hash

	for i := 0; i < 3; i++ {
		proposalTx := &core.Transaction{
			TxInner: dao.ProposalTx{
				Fee:          200,
				Title:        fmt.Sprintf("Concurrent Proposal %d", i+1),
				Description:  fmt.Sprintf("Testing concurrent voting on proposal %d", i+1),
				ProposalType: dao.ProposalTypeGeneral,
				VotingType:   dao.VotingTypeSimple,
				StartTime:    time.Now().Unix() - 100,
				EndTime:      time.Now().Unix() + 3600,
				Threshold:    5100,
				MetadataHash: randomHash(),
			},
			From: creator.PublicKey(),
		}
		proposalTx.Sign(creator)
		proposalTxs = append(proposalTxs, proposalTx)
		proposalIDs = append(proposalIDs, proposalTx.Hash(core.TxHasher{}))
	}

	// Add all proposals
	proposalBlock := createBlockWithTxs(t, bc, proposalTxs)
	err := bc.AddBlock(proposalBlock)
	require.NoError(t, err)

	// Concurrent voting on all proposals
	var allVoteTxs []*core.Transaction

	for voterIdx, voter := range voters {
		for propIdx, proposalID := range proposalIDs {
			// Vary voting patterns
			choice := dao.VoteChoiceYes
			if (voterIdx+propIdx)%2 == 0 {
				choice = dao.VoteChoiceNo
			}

			voteTx := &core.Transaction{
				TxInner: dao.VoteTx{
					Fee:        100,
					ProposalID: proposalID,
					Choice:     choice,
					Weight:     500,
					Reason:     fmt.Sprintf("Vote from voter %d on proposal %d", voterIdx, propIdx),
				},
				From: voter.PublicKey(),
			}
			voteTx.Sign(voter)
			allVoteTxs = append(allVoteTxs, voteTx)
		}
	}

	// Add all votes in one block
	voteBlock := createBlockWithTxs(t, bc, allVoteTxs)
	err = bc.AddBlock(voteBlock)
	require.NoError(t, err)

	// Verify all proposals received votes
	for i, proposalID := range proposalIDs {
		votes, err := bc.GetVotes(proposalID)
		require.NoError(t, err)
		assert.Len(t, votes, 10, "Proposal %d should have 10 votes", i+1)

		proposal, err := bc.GetProposal(proposalID)
		require.NoError(t, err)
		assert.Equal(t, uint64(10), proposal.Results.TotalVoters)
	}

	t.Log("Multi-proposal concurrent voting test passed")
}

// Helper functions

func setupTestBlockchain(t *testing.T) (*core.Blockchain, func()) {
	logger := log.NewNopLogger()
	genesis := createTestGenesisBlock(t)
	bc, err := core.NewBlockchain(logger, genesis)
	require.NoError(t, err)

	return bc, func() {
		// Cleanup if needed
	}
}

func createTestGenesisBlock(t *testing.T) *core.Block {
	privKey := crypto.GeneratePrivateKey()
	header := &core.Header{
		Version:       1,
		PrevBlockHash: types.Hash{},
		Height:        0,
		Timestamp:     time.Now().UnixNano(),
	}

	block, err := core.NewBlock(header, []*core.Transaction{})
	require.NoError(t, err)

	dataHash, err := core.CalculateDataHash(block.Transactions)
	require.NoError(t, err)
	block.Header.DataHash = dataHash

	err = block.Sign(privKey)
	require.NoError(t, err)

	return block
}

func setupTestUsers(t *testing.T, bc *core.Blockchain, users ...crypto.PrivateKey) {
	// Initialize DAO token state directly for testing
	tokenState := bc.GetDAOTokenState()

	// Create an admin user with initial tokens
	admin := crypto.GeneratePrivateKey()
	adminStr := admin.PublicKey().String()

	// Mint initial supply to admin
	err := tokenState.Mint(adminStr, 1000000) // 1M tokens
	require.NoError(t, err)

	// Distribute tokens to test users
	for _, user := range users {
		userStr := user.PublicKey().String()
		err := tokenState.Transfer(adminStr, userStr, 10000)
		require.NoError(t, err)

		// Verify token balance
		balance := bc.GetTokenBalance(user.PublicKey())
		assert.Equal(t, uint64(10000), balance)
	}
}

func createBlockWithTxs(t *testing.T, bc *core.Blockchain, txs []*core.Transaction) *core.Block {
	privKey := crypto.GeneratePrivateKey()
	prevHeader, err := bc.GetHeader(bc.Height())
	require.NoError(t, err)
	prevBlockHash := core.BlockHasher{}.Hash(prevHeader)

	header := &core.Header{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		Height:        bc.Height() + 1,
		Timestamp:     time.Now().UnixNano(),
	}

	block, err := core.NewBlock(header, txs)
	require.NoError(t, err)

	dataHash, err := core.CalculateDataHash(block.Transactions)
	require.NoError(t, err)
	block.Header.DataHash = dataHash

	err = block.Sign(privKey)
	require.NoError(t, err)

	return block
}

func randomHash() types.Hash {
	var hash types.Hash
	rand.Read(hash[:])
	return hash
}
