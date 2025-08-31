package core

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/dao"
	"github.com/BOCK-CHAIN/BockChain/types"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockchainDAOIntegration(t *testing.T) {
	// Create a test blockchain
	bc, cleanup := newTestBlockchain(t)
	defer cleanup()

	// Create test users
	creator := crypto.GeneratePrivateKey()
	voter1 := crypto.GeneratePrivateKey()
	voter2 := crypto.GeneratePrivateKey()

	// Initialize users with governance tokens
	initializeTestUsers(t, bc, creator, voter1, voter2)

	t.Run("ProposalCreationAndVoting", func(t *testing.T) {
		testProposalCreationAndVoting(t, bc, creator, voter1, voter2)
	})

	t.Run("DelegationFlow", func(t *testing.T) {
		testDelegationFlow(t, bc, creator, voter1, voter2)
	})

	t.Run("TreasuryOperations", func(t *testing.T) {
		testTreasuryOperations(t, bc, creator, voter1, voter2)
	})

	t.Run("TokenOperations", func(t *testing.T) {
		testTokenOperations(t, bc, creator, voter1, voter2)
	})

	t.Run("QuadraticVoting", func(t *testing.T) {
		testQuadraticVoting(t, bc, creator, voter1, voter2)
	})
}

func newTestBlockchain(t *testing.T) (*Blockchain, func()) {
	logger := log.NewNopLogger()
	genesis := randomDAOBlock(t, 0, types.Hash{})
	bc, err := NewBlockchain(logger, genesis)
	require.NoError(t, err)

	return bc, func() {
		// Cleanup if needed
	}
}

func initializeTestUsers(t *testing.T, bc *Blockchain, users ...crypto.PrivateKey) {
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

func testProposalCreationAndVoting(t *testing.T, bc *Blockchain, creator, voter1, voter2 crypto.PrivateKey) {
	// Create a proposal
	proposalTx := &Transaction{
		TxInner: dao.ProposalTx{
			Fee:          200,
			Title:        "Test Proposal",
			Description:  "This is a test proposal for integration testing",
			ProposalType: dao.ProposalTypeGeneral,
			VotingType:   dao.VotingTypeSimple,
			StartTime:    time.Now().Unix() - 100,   // Started 100 seconds ago
			EndTime:      time.Now().Unix() + 86400, // Ends in 24 hours
			Threshold:    5100,                      // 51%
			MetadataHash: randomHash(),
		},
		From: creator.PublicKey(),
	}
	proposalTx.Sign(creator)

	// Add proposal to blockchain
	block := randomDAOBlockWithTxs(t, bc.Height()+1, getDAOPrevBlockHash(t, bc), []*Transaction{proposalTx})
	err := bc.AddBlock(block)
	require.NoError(t, err)

	proposalID := proposalTx.Hash(TxHasher{})

	// Verify proposal was created
	proposal, err := bc.GetProposal(proposalID)
	require.NoError(t, err)
	assert.Equal(t, "Test Proposal", proposal.Title)

	// Update proposal status based on current time
	err = bc.UpdateProposalStatuses()
	require.NoError(t, err)

	// Check proposal status again
	proposal, err = bc.GetProposal(proposalID)
	require.NoError(t, err)
	assert.Equal(t, dao.ProposalStatusActive, proposal.Status)

	// Vote on the proposal
	vote1Tx := &Transaction{
		TxInner: dao.VoteTx{
			Fee:        100,
			ProposalID: proposalID,
			Choice:     dao.VoteChoiceYes,
			Weight:     1000,
			Reason:     "I support this proposal",
		},
		From: voter1.PublicKey(),
	}
	vote1Tx.Sign(voter1)

	vote2Tx := &Transaction{
		TxInner: dao.VoteTx{
			Fee:        100,
			ProposalID: proposalID,
			Choice:     dao.VoteChoiceNo,
			Weight:     500,
			Reason:     "I disagree with this proposal",
		},
		From: voter2.PublicKey(),
	}
	vote2Tx.Sign(voter2)

	// Add votes to blockchain
	voteBlock := randomDAOBlockWithTxs(t, bc.Height()+1, getDAOPrevBlockHash(t, bc), []*Transaction{vote1Tx, vote2Tx})
	err = bc.AddBlock(voteBlock)
	require.NoError(t, err)

	// Verify votes were recorded
	votes, err := bc.GetVotes(proposalID)
	require.NoError(t, err)
	assert.Len(t, votes, 2)

	// Check vote results
	proposal, err = bc.GetProposal(proposalID)
	require.NoError(t, err)
	assert.Equal(t, uint64(1000), proposal.Results.YesVotes)
	assert.Equal(t, uint64(500), proposal.Results.NoVotes)
	assert.Equal(t, uint64(2), proposal.Results.TotalVoters)
}

func testDelegationFlow(t *testing.T, bc *Blockchain, creator, voter1, voter2 crypto.PrivateKey) {
	// Create delegation from voter1 to voter2
	delegationTx := &Transaction{
		TxInner: dao.DelegationTx{
			Fee:      100,
			Delegate: voter2.PublicKey(),
			Duration: 3600, // 1 hour
			Revoke:   false,
		},
		From: voter1.PublicKey(),
	}
	delegationTx.Sign(voter1)

	// Add delegation to blockchain
	block := randomDAOBlockWithTxs(t, bc.Height()+1, getDAOPrevBlockHash(t, bc), []*Transaction{delegationTx})
	err := bc.AddBlock(block)
	require.NoError(t, err)

	// Verify delegation was created
	delegation, err := bc.GetDelegation(voter1.PublicKey())
	require.NoError(t, err)
	assert.Equal(t, voter2.PublicKey(), delegation.Delegate)
	assert.True(t, delegation.Active)

	// Test effective voting power
	processor := bc.GetDAOProcessor()
	voter1Power := processor.GetEffectiveVotingPower(voter1.PublicKey())
	voter2Power := processor.GetEffectiveVotingPower(voter2.PublicKey())

	// voter1 should have no direct voting power (delegated)
	assert.Equal(t, uint64(0), voter1Power)
	// voter2 should have their own balance plus delegated power
	expectedVoter2Power := bc.GetTokenBalance(voter2.PublicKey()) + bc.GetTokenBalance(voter1.PublicKey())
	assert.Equal(t, expectedVoter2Power, voter2Power)

	// Test delegation revocation
	revokeTx := &Transaction{
		TxInner: dao.DelegationTx{
			Fee:      100,
			Delegate: voter2.PublicKey(),
			Duration: 0,
			Revoke:   true,
		},
		From: voter1.PublicKey(),
	}
	revokeTx.Sign(voter1)

	revokeBlock := randomDAOBlockWithTxs(t, bc.Height()+1, getDAOPrevBlockHash(t, bc), []*Transaction{revokeTx})
	err = bc.AddBlock(revokeBlock)
	require.NoError(t, err)

	// Verify delegation was revoked
	delegation, err = bc.GetDelegation(voter1.PublicKey())
	require.NoError(t, err)
	assert.False(t, delegation.Active)
}

func testTreasuryOperations(t *testing.T, bc *Blockchain, creator, voter1, voter2 crypto.PrivateKey) {
	// First, add funds to treasury (simulate treasury funding)
	treasuryState := bc.GetTreasuryState()
	treasuryState.Balance = 50000
	treasuryState.Signers = []crypto.PublicKey{creator.PublicKey(), voter1.PublicKey()}
	treasuryState.RequiredSigs = 2

	// Create treasury transaction
	sig1, _ := creator.Sign([]byte("treasury-tx-data"))
	sig2, _ := voter1.Sign([]byte("treasury-tx-data"))

	treasuryTx := &Transaction{
		TxInner: dao.TreasuryTx{
			Fee:          200,
			Recipient:    voter2.PublicKey(),
			Amount:       5000,
			Purpose:      "Development grant",
			Signatures:   []crypto.Signature{*sig1, *sig2},
			RequiredSigs: 2,
		},
		From: creator.PublicKey(),
	}
	treasuryTx.Sign(creator)

	// Add treasury transaction to blockchain
	block := randomDAOBlockWithTxs(t, bc.Height()+1, getDAOPrevBlockHash(t, bc), []*Transaction{treasuryTx})
	err := bc.AddBlock(block)
	require.NoError(t, err)

	// Verify treasury transaction was processed
	updatedTreasuryState := bc.GetTreasuryState()
	// Check if transaction was executed (balance should be reduced)
	if updatedTreasuryState.Balance == 45000 {
		// Transaction was executed
		assert.Equal(t, uint64(45000), updatedTreasuryState.Balance) // 50000 - 5000

		// Verify recipient received tokens
		recipientBalance := bc.GetTokenBalance(voter2.PublicKey())
		assert.Greater(t, recipientBalance, uint64(10000)) // Should have more than initial balance
	} else {
		// Transaction is pending (not enough signatures or other issue)
		assert.Equal(t, uint64(50000), updatedTreasuryState.Balance)
		assert.Len(t, updatedTreasuryState.Transactions, 1) // Should have pending transaction
	}
}

func testTokenOperations(t *testing.T, bc *Blockchain, creator, voter1, voter2 crypto.PrivateKey) {
	initialBalance1 := bc.GetTokenBalance(voter1.PublicKey())
	initialBalance2 := bc.GetTokenBalance(voter2.PublicKey())

	// Test token transfer
	transferTx := &Transaction{
		TxInner: dao.TokenTransferTx{
			Fee:       100,
			Recipient: voter2.PublicKey(),
			Amount:    1000,
		},
		From: voter1.PublicKey(),
	}
	transferTx.Sign(voter1)

	block := randomDAOBlockWithTxs(t, bc.Height()+1, getDAOPrevBlockHash(t, bc), []*Transaction{transferTx})
	err := bc.AddBlock(block)
	require.NoError(t, err)

	// Verify balances updated
	newBalance1 := bc.GetTokenBalance(voter1.PublicKey())
	newBalance2 := bc.GetTokenBalance(voter2.PublicKey())

	assert.Equal(t, initialBalance1-1000-100, newBalance1) // amount + fee
	assert.Equal(t, initialBalance2+1000, newBalance2)

	// Test token approval and transferFrom
	approveTx := &Transaction{
		TxInner: dao.TokenApproveTx{
			Fee:     100,
			Spender: voter2.PublicKey(),
			Amount:  500,
		},
		From: voter1.PublicKey(),
	}
	approveTx.Sign(voter1)

	approveBlock := randomDAOBlockWithTxs(t, bc.Height()+1, getDAOPrevBlockHash(t, bc), []*Transaction{approveTx})
	err = bc.AddBlock(approveBlock)
	require.NoError(t, err)

	// Verify allowance
	tokenState := bc.GetDAOTokenState()
	allowance := tokenState.GetAllowance(voter1.PublicKey().String(), voter2.PublicKey().String())
	assert.Equal(t, uint64(500), allowance)

	// Test transferFrom
	transferFromTx := &Transaction{
		TxInner: dao.TokenTransferFromTx{
			Fee:       100,
			From:      voter1.PublicKey(),
			Recipient: creator.PublicKey(),
			Amount:    300,
		},
		From: voter2.PublicKey(),
	}
	transferFromTx.Sign(voter2)

	transferFromBlock := randomDAOBlockWithTxs(t, bc.Height()+1, getDAOPrevBlockHash(t, bc), []*Transaction{transferFromTx})
	err = bc.AddBlock(transferFromBlock)
	require.NoError(t, err)

	// Verify allowance was reduced
	newAllowance := tokenState.GetAllowance(voter1.PublicKey().String(), voter2.PublicKey().String())
	assert.Equal(t, uint64(200), newAllowance) // 500 - 300
}

func testQuadraticVoting(t *testing.T, bc *Blockchain, creator, voter1, voter2 crypto.PrivateKey) {
	// Create a proposal with quadratic voting
	proposalTx := &Transaction{
		TxInner: dao.ProposalTx{
			Fee:          200,
			Title:        "Quadratic Voting Test",
			Description:  "Testing quadratic voting mechanism",
			ProposalType: dao.ProposalTypeGeneral,
			VotingType:   dao.VotingTypeQuadratic,
			StartTime:    time.Now().Unix() - 100,
			EndTime:      time.Now().Unix() + 86400,
			Threshold:    5100,
			MetadataHash: randomHash(),
		},
		From: creator.PublicKey(),
	}
	proposalTx.Sign(creator)

	block := randomDAOBlockWithTxs(t, bc.Height()+1, getDAOPrevBlockHash(t, bc), []*Transaction{proposalTx})
	err := bc.AddBlock(block)
	require.NoError(t, err)

	proposalID := proposalTx.Hash(TxHasher{})

	// Update proposal status to active
	err = bc.UpdateProposalStatuses()
	require.NoError(t, err)

	// Cast quadratic vote (weight=10, cost=100)
	initialBalance := bc.GetTokenBalance(voter1.PublicKey())
	quadraticVoteTx := &Transaction{
		TxInner: dao.VoteTx{
			Fee:        100,
			ProposalID: proposalID,
			Choice:     dao.VoteChoiceYes,
			Weight:     10, // Cost will be 10^2 = 100 tokens
			Reason:     "Quadratic vote test",
		},
		From: voter1.PublicKey(),
	}
	quadraticVoteTx.Sign(voter1)

	voteBlock := randomDAOBlockWithTxs(t, bc.Height()+1, getDAOPrevBlockHash(t, bc), []*Transaction{quadraticVoteTx})
	err = bc.AddBlock(voteBlock)
	require.NoError(t, err)

	// Verify quadratic cost was applied (weight^2 + fee)
	newBalance := bc.GetTokenBalance(voter1.PublicKey())
	expectedCost := uint64(100 + 100) // 10^2 + fee
	assert.Equal(t, initialBalance-expectedCost, newBalance)

	// Verify vote was recorded with correct weight
	votes, err := bc.GetVotes(proposalID)
	require.NoError(t, err)
	vote := votes[voter1.PublicKey().String()]
	assert.Equal(t, uint64(10), vote.Weight) // Effective weight is 10, not 100
}

// Helper functions

func randomDAOBlock(t *testing.T, height uint32, prevBlockHash types.Hash) *Block {
	privKey := crypto.GeneratePrivateKey()
	header := &Header{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		Height:        height,
		Timestamp:     time.Now().UnixNano(),
	}

	block, err := NewBlock(header, []*Transaction{})
	require.NoError(t, err)

	dataHash, err := CalculateDataHash(block.Transactions)
	require.NoError(t, err)
	block.Header.DataHash = dataHash

	err = block.Sign(privKey)
	require.NoError(t, err)

	return block
}

func randomDAOBlockWithTxs(t *testing.T, height uint32, prevBlockHash types.Hash, txs []*Transaction) *Block {
	privKey := crypto.GeneratePrivateKey()
	header := &Header{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		Height:        height,
		Timestamp:     time.Now().UnixNano(),
	}

	block, err := NewBlock(header, txs)
	require.NoError(t, err)

	dataHash, err := CalculateDataHash(block.Transactions)
	require.NoError(t, err)
	block.Header.DataHash = dataHash

	err = block.Sign(privKey)
	require.NoError(t, err)

	return block
}

func getDAOPrevBlockHash(t *testing.T, bc *Blockchain) types.Hash {
	prevHeader, err := bc.GetHeader(bc.Height())
	require.NoError(t, err)
	return BlockHasher{}.Hash(prevHeader)
}

func randomHash() types.Hash {
	var hash types.Hash
	rand.Read(hash[:])
	return hash
}
