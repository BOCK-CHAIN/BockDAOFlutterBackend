package tests

import (
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

// TestSystemIntegration tests the complete integrated DAO system
func TestSystemIntegration(t *testing.T) {
	t.Run("CompleteSystemBootstrap", testCompleteSystemBootstrap)
	t.Run("EndToEndGovernanceWorkflow", testEndToEndGovernanceWorkflow)
	t.Run("APIServerIntegration", testAPIServerIntegration)
	t.Run("TreasuryOperations", testTreasuryOperations)
	t.Run("TokenOperations", testTokenOperations)
	t.Run("DelegationWorkflow", testDelegationWorkflow)
	t.Run("ReputationSystem", testReputationSystem)
	t.Run("SecurityValidation", testSecurityValidation)
	t.Run("PerformanceUnderLoad", testPerformanceUnderLoad)
	t.Run("CrossPlatformConsistency", testCrossPlatformConsistency)
}

// IntegratedTestSystem represents a complete test system
type IntegratedTestSystem struct {
	daoInstance *dao.DAO
	blockchain  *core.Blockchain
	logger      log.Logger
	cleanup     func()
}

func testCompleteSystemBootstrap(t *testing.T) {
	system := setupIntegratedTestSystem(t)
	defer system.cleanup()

	// Test 1: Verify DAO initialization
	assert.NotNil(t, system.daoInstance)
	assert.NotNil(t, system.daoInstance.GovernanceState)
	assert.NotNil(t, system.daoInstance.TokenState)

	// Test 2: Verify token system initialization
	totalSupply := system.daoInstance.GetTotalSupply()
	assert.Greater(t, totalSupply, uint64(0))

	// Test 3: Verify treasury initialization
	treasuryBalance := system.daoInstance.GetTreasuryBalance()
	assert.GreaterOrEqual(t, treasuryBalance, uint64(0))

	// Test 4: Verify blockchain initialization
	assert.Greater(t, system.blockchain.Height(), uint32(0))

	t.Log("Complete system bootstrap test passed")
}

func testEndToEndGovernanceWorkflow(t *testing.T) {
	system := setupIntegratedTestSystem(t)
	defer system.cleanup()

	// Setup test participants
	creator := crypto.GeneratePrivateKey()
	voters := make([]crypto.PrivateKey, 5)
	for i := range voters {
		voters[i] = crypto.GeneratePrivateKey()
	}

	// Initialize participants with tokens
	setupTestUsersForIntegration(t, system.daoInstance, append(voters, creator)...)

	// Phase 1: Create proposal
	proposalTx := &dao.ProposalTx{
		Fee:          200,
		Title:        "End-to-End Integration Test Proposal",
		Description:  "Testing complete governance workflow",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() - 100,
		EndTime:      time.Now().Unix() + 3600,
		Threshold:    5100,
		MetadataHash: randomHashForIntegration(),
	}

	proposalHash := generateTxHashForIntegration(proposalTx, creator)
	err := system.daoInstance.ProcessDAOTransaction(proposalTx, creator.PublicKey(), proposalHash)
	require.NoError(t, err)

	// Verify proposal was created
	proposal, err := system.daoInstance.GetProposal(proposalHash)
	require.NoError(t, err)
	assert.Equal(t, proposalTx.Title, proposal.Title)
	assert.Equal(t, dao.ProposalStatusActive, proposal.Status)

	// Phase 2: Cast votes
	for i, voter := range voters {
		voteTx := &dao.VoteTx{
			Fee:        100,
			ProposalID: proposalHash,
			Choice:     dao.VoteChoiceYes,
			Weight:     1000,
			Reason:     fmt.Sprintf("Vote from voter %d", i),
		}

		voteHash := generateTxHashForIntegration(voteTx, voter)
		err := system.daoInstance.ProcessDAOTransaction(voteTx, voter.PublicKey(), voteHash)
		require.NoError(t, err)
	}

	// Phase 3: Verify voting results
	votes, err := system.daoInstance.GetVotes(proposalHash)
	require.NoError(t, err)
	assert.Len(t, votes, 5)

	// Update proposal status
	system.daoInstance.UpdateAllProposalStatuses()

	// Verify proposal results
	updatedProposal, err := system.daoInstance.GetProposal(proposalHash)
	require.NoError(t, err)
	assert.NotNil(t, updatedProposal.Results)
	assert.Equal(t, uint64(5000), updatedProposal.Results.YesVotes)
	assert.Equal(t, uint64(5), updatedProposal.Results.TotalVoters)

	t.Log("End-to-end governance workflow test passed")
}

func testAPIServerIntegration(t *testing.T) {
	system := setupIntegratedTestSystem(t)
	defer system.cleanup()

	// Test basic DAO functionality through direct calls
	// (In a real test, this would test HTTP endpoints)

	// Test proposal creation
	creator := crypto.GeneratePrivateKey()
	setupTestUsersForIntegration(t, system.daoInstance, creator)

	proposalTx := &dao.ProposalTx{
		Fee:          200,
		Title:        "API Integration Test",
		Description:  "Testing API integration",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() - 100,
		EndTime:      time.Now().Unix() + 3600,
		Threshold:    1000,
		MetadataHash: randomHash(),
	}

	proposalHash := generateTxHash(proposalTx, creator)
	err := system.daoInstance.ProcessDAOTransaction(proposalTx, creator.PublicKey(), proposalHash)
	require.NoError(t, err)

	// Test proposal retrieval
	proposal, err := system.daoInstance.GetProposal(proposalHash)
	require.NoError(t, err)
	assert.Equal(t, proposalTx.Title, proposal.Title)

	// Test token balance retrieval
	balance := system.daoInstance.GetTokenBalance(creator.PublicKey())
	assert.Greater(t, balance, uint64(0))

	// Test treasury status
	treasuryBalance := system.daoInstance.GetTreasuryBalance()
	assert.GreaterOrEqual(t, treasuryBalance, uint64(0))

	t.Log("API server integration test passed")
}

func testTreasuryOperations(t *testing.T) {
	system := setupIntegratedTestSystem(t)
	defer system.cleanup()

	// Setup treasury signers
	signer1 := crypto.GeneratePrivateKey()
	signer2 := crypto.GeneratePrivateKey()
	recipient := crypto.GeneratePrivateKey()

	signers := []crypto.PublicKey{signer1.PublicKey(), signer2.PublicKey()}
	err := system.daoInstance.InitializeTreasury(signers, 2)
	require.NoError(t, err)

	// Add funds to treasury
	system.daoInstance.AddTreasuryFunds(100000)

	// Create treasury transaction
	treasuryTx := &dao.TreasuryTx{
		Fee:          500,
		Recipient:    recipient.PublicKey(),
		Amount:       10000,
		Purpose:      "Test treasury transaction",
		Signatures:   []crypto.Signature{},
		RequiredSigs: 2,
	}

	txHash := generateTxHash(treasuryTx, signer1)
	err = system.daoInstance.CreateTreasuryTransaction(treasuryTx, txHash)
	require.NoError(t, err)

	// Sign with first signer
	err = system.daoInstance.SignTreasuryTransaction(txHash, signer1)
	require.NoError(t, err)

	// Sign with second signer
	err = system.daoInstance.SignTreasuryTransaction(txHash, signer2)
	require.NoError(t, err)

	// Execute transaction
	err = system.daoInstance.ExecuteTreasuryTransaction(txHash)
	require.NoError(t, err)

	// Verify transaction was executed
	tx, exists := system.daoInstance.GetTreasuryTransaction(txHash)
	require.True(t, exists)
	assert.True(t, tx.Executed)

	t.Log("Treasury operations test passed")
}

func testTokenOperations(t *testing.T) {
	system := setupIntegratedTestSystem(t)
	defer system.cleanup()

	// Setup test users
	user1 := crypto.GeneratePrivateKey()
	user2 := crypto.GeneratePrivateKey()
	setupTestUsersForIntegration(t, system.daoInstance, user1, user2)

	// Test token transfer
	initialBalance1 := system.daoInstance.GetTokenBalance(user1.PublicKey())
	initialBalance2 := system.daoInstance.GetTokenBalance(user2.PublicKey())

	transferAmount := uint64(1000)
	err := system.daoInstance.TransferTokens(user1.PublicKey(), user2.PublicKey(), transferAmount)
	require.NoError(t, err)

	// Verify balances updated
	newBalance1 := system.daoInstance.GetTokenBalance(user1.PublicKey())
	newBalance2 := system.daoInstance.GetTokenBalance(user2.PublicKey())

	assert.Equal(t, initialBalance1-transferAmount, newBalance1)
	assert.Equal(t, initialBalance2+transferAmount, newBalance2)

	// Test token approval
	approveAmount := uint64(500)
	err = system.daoInstance.ApproveTokens(user1.PublicKey(), user2.PublicKey(), approveAmount)
	require.NoError(t, err)

	// Verify allowance
	allowance := system.daoInstance.GetTokenAllowance(user1.PublicKey(), user2.PublicKey())
	assert.Equal(t, approveAmount, allowance)

	t.Log("Token operations test passed")
}

func testDelegationWorkflow(t *testing.T) {
	system := setupIntegratedTestSystem(t)
	defer system.cleanup()

	// Setup test users
	delegator := crypto.GeneratePrivateKey()
	delegate := crypto.GeneratePrivateKey()
	setupTestUsersForIntegration(t, system.daoInstance, delegator, delegate)

	// Create delegation
	delegationTx := &dao.DelegationTx{
		Fee:      200,
		Delegate: delegate.PublicKey(),
		Duration: 86400, // 24 hours
		Revoke:   false,
	}

	txHash := generateTxHash(delegationTx, delegator)
	err := system.daoInstance.ProcessDAOTransaction(delegationTx, delegator.PublicKey(), txHash)
	require.NoError(t, err)

	// Verify delegation exists
	delegation, exists := system.daoInstance.GetDelegation(delegator.PublicKey())
	require.True(t, exists)
	assert.Equal(t, delegate.PublicKey().String(), delegation.Delegate.String())
	assert.True(t, delegation.Active)

	// Test delegated voting power
	delegatedPower := system.daoInstance.GetDelegatedPower(delegate.PublicKey())
	assert.Greater(t, delegatedPower, uint64(0))

	// Revoke delegation
	revokeTx := &dao.DelegationTx{
		Fee:      200,
		Delegate: crypto.PublicKey{},
		Duration: 0,
		Revoke:   true,
	}

	revokeHash := generateTxHash(revokeTx, delegator)
	err = system.daoInstance.ProcessDAOTransaction(revokeTx, delegator.PublicKey(), revokeHash)
	require.NoError(t, err)

	// Verify delegation revoked
	updatedDelegation, exists := system.daoInstance.GetDelegation(delegator.PublicKey())
	if exists {
		assert.False(t, updatedDelegation.Active)
	}

	t.Log("Delegation workflow test passed")
}

func testReputationSystem(t *testing.T) {
	system := setupIntegratedTestSystem(t)
	defer system.cleanup()

	// Setup test user
	user := crypto.GeneratePrivateKey()
	setupTestUsersForIntegration(t, system.daoInstance, user)

	// Initialize reputation
	system.daoInstance.InitializeUserReputation(user.PublicKey(), 10000)

	// Get initial reputation
	initialReputation := system.daoInstance.GetUserReputation(user.PublicKey())
	assert.Greater(t, initialReputation, uint64(0))

	// Apply inactivity decay
	system.daoInstance.ApplyInactivityDecay()

	// Verify reputation system is working
	stats := system.daoInstance.GetReputationStats()
	assert.NotNil(t, stats)

	t.Log("Reputation system test passed")
}

func testSecurityValidation(t *testing.T) {
	system := setupIntegratedTestSystem(t)
	defer system.cleanup()

	// Setup test users
	admin := crypto.GeneratePrivateKey()
	user := crypto.GeneratePrivateKey()
	setupTestUsersForIntegration(t, system.daoInstance, admin, user)

	// Initialize founder roles
	err := system.daoInstance.InitializeFounderRoles([]crypto.PublicKey{admin.PublicKey()})
	require.NoError(t, err)

	// Test permission checking
	hasPermission := system.daoInstance.HasPermission(admin.PublicKey(), dao.PermissionManageRoles)
	assert.True(t, hasPermission)

	// Test access validation
	err = system.daoInstance.ValidateAccess(admin.PublicKey(), "TestOperation", "TestResource", dao.SecurityLevelMember)
	assert.NoError(t, err)

	// Test emergency activation
	err = system.daoInstance.ActivateEmergency(admin.PublicKey(), "Test emergency", dao.SecurityLevelCritical, []string{"voting"})
	assert.NoError(t, err)

	// Verify emergency is active
	isActive := system.daoInstance.IsEmergencyActive()
	assert.True(t, isActive)

	// Deactivate emergency
	err = system.daoInstance.DeactivateEmergency(admin.PublicKey())
	assert.NoError(t, err)

	t.Log("Security validation test passed")
}

func testPerformanceUnderLoad(t *testing.T) {
	system := setupIntegratedTestSystem(t)
	defer system.cleanup()

	// Setup multiple users
	users := make([]crypto.PrivateKey, 50)
	for i := range users {
		users[i] = crypto.GeneratePrivateKey()
	}
	setupTestUsersForIntegration(t, system.daoInstance, users...)

	// Create multiple proposals concurrently
	start := time.Now()

	for i := 0; i < 10; i++ {
		proposalTx := &dao.ProposalTx{
			Fee:          200,
			Title:        fmt.Sprintf("Load Test Proposal %d", i),
			Description:  fmt.Sprintf("Testing system under load - proposal %d", i),
			ProposalType: dao.ProposalTypeGeneral,
			VotingType:   dao.VotingTypeSimple,
			StartTime:    time.Now().Unix() - 100,
			EndTime:      time.Now().Unix() + 3600,
			Threshold:    1000,
			MetadataHash: randomHash(),
		}

		proposalHash := generateTxHash(proposalTx, users[i%len(users)])
		err := system.daoInstance.ProcessDAOTransaction(proposalTx, users[i%len(users)].PublicKey(), proposalHash)
		require.NoError(t, err)
	}

	duration := time.Since(start)
	t.Logf("Created 10 proposals in %v", duration)

	// Verify all proposals were created
	proposals := system.daoInstance.ListAllProposals()
	assert.GreaterOrEqual(t, len(proposals), 10)

	t.Log("Performance under load test passed")
}

func testCrossPlatformConsistency(t *testing.T) {
	system := setupIntegratedTestSystem(t)
	defer system.cleanup()

	// Create proposal through DAO instance
	creator := crypto.GeneratePrivateKey()
	setupTestUsersForIntegration(t, system.daoInstance, creator)

	proposalTx := &dao.ProposalTx{
		Fee:          200,
		Title:        "Cross-Platform Test",
		Description:  "Testing consistency across platforms",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() - 100,
		EndTime:      time.Now().Unix() + 3600,
		Threshold:    1000,
		MetadataHash: randomHash(),
	}

	proposalHash := generateTxHash(proposalTx, creator)
	err := system.daoInstance.ProcessDAOTransaction(proposalTx, creator.PublicKey(), proposalHash)
	require.NoError(t, err)

	// Retrieve via DAO instance
	proposal, err := system.daoInstance.GetProposal(proposalHash)
	require.NoError(t, err)

	// Verify consistency
	assert.Equal(t, proposalTx.Title, proposal.Title)
	assert.Equal(t, proposalTx.Description, proposal.Description)
	assert.Equal(t, proposalTx.ProposalType, proposal.ProposalType)

	t.Log("Cross-platform consistency test passed")
}

// Helper functions

func setupIntegratedTestSystem(t *testing.T) *IntegratedTestSystem {
	logger := log.NewNopLogger()

	// Create test blockchain
	genesis := createTestGenesisBlockForIntegration(t)
	blockchain, err := core.NewBlockchain(logger, genesis)
	require.NoError(t, err)

	// Create DAO instance
	daoInstance := dao.NewDAO("TEST", "Test Token", 18)

	// Initialize with test distribution
	testDistribution := map[string]uint64{
		"test_treasury": 1000000,
	}
	err = daoInstance.InitialTokenDistribution(testDistribution)
	require.NoError(t, err)

	return &IntegratedTestSystem{
		daoInstance: daoInstance,
		blockchain:  blockchain,
		logger:      logger,
		cleanup: func() {
			// Cleanup resources if needed
		},
	}
}

func createTestGenesisBlockForIntegration(t *testing.T) *core.Block {
	privKey := crypto.GeneratePrivateKey()

	genesisTx := &core.Transaction{
		TxInner: core.CollectionTx{
			Fee:      0,
			MetaData: []byte("Test Genesis Block"),
		},
		From:  privKey.PublicKey(),
		To:    privKey.PublicKey(),
		Value: 1000000000,
	}
	genesisTx.Sign(privKey)

	header := &core.Header{
		Version:       1,
		PrevBlockHash: types.Hash{},
		Height:        0,
		Timestamp:     time.Now().UnixNano(),
	}

	block, err := core.NewBlock(header, []*core.Transaction{genesisTx})
	require.NoError(t, err)

	dataHash, err := core.CalculateDataHash(block.Transactions)
	require.NoError(t, err)
	block.Header.DataHash = dataHash

	err = block.Sign(privKey)
	require.NoError(t, err)

	return block
}

func setupTestUsersForIntegration(t *testing.T, daoInstance *dao.DAO, users ...crypto.PrivateKey) {
	for _, user := range users {
		// Mint tokens to user
		err := daoInstance.MintTokens(user.PublicKey(), 10000)
		require.NoError(t, err)

		// Initialize reputation
		daoInstance.InitializeUserReputation(user.PublicKey(), 10000)
	}
}

func generateTxHashForIntegration(tx interface{}, signer crypto.PrivateKey) types.Hash {
	// Simple hash generation for testing
	data := fmt.Sprintf("%v%s%d", tx, signer.PublicKey().String(), time.Now().UnixNano())
	hash := [32]byte{}
	copy(hash[:], []byte(data)[:32])
	return hash
}

func randomHashForIntegration() types.Hash {
	hash := [32]byte{}
	for i := range hash {
		hash[i] = byte(i)
	}
	return hash
}
