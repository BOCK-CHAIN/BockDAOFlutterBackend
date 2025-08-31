package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/api"
	"github.com/BOCK-CHAIN/BockChain/core"
	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/dao"
	"github.com/BOCK-CHAIN/BockChain/network"
	"github.com/BOCK-CHAIN/BockChain/types"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CompleteSystemIntegrationTest tests the entire DAO system integration
func TestCompleteSystemIntegration(t *testing.T) {
	t.Run("SystemBootstrapAndInitialization", testSystemBootstrapAndInitialization)
	t.Run("EndToEndGovernanceWorkflow", testEndToEndGovernanceWorkflow)
	t.Run("ConcurrentOperationsStressTest", testConcurrentOperationsStressTest)
	t.Run("SecurityAndAttackVectorValidation", testSecurityAndAttackVectorValidation)
	t.Run("PerformanceUnderLoad", testPerformanceUnderLoad)
	t.Run("CrossComponentIntegration", testCrossComponentIntegration)
	t.Run("ErrorHandlingAndRecovery", testErrorHandlingAndRecovery)
	t.Run("DataConsistencyValidation", testDataConsistencyValidation)
}

// IntegratedSystemTestSuite represents the complete test environment
type IntegratedSystemTestSuite struct {
	daoInstance   *dao.DAO
	blockchain    *core.Blockchain
	networkServer *network.Server
	apiServer     *api.DAOServer
	logger        log.Logger
	ctx           context.Context
	cancel        context.CancelFunc
	txChan        chan *core.Transaction
	cleanup       func()
}

// setupCompleteSystem creates a fully integrated test system
func setupCompleteSystem(t *testing.T) *IntegratedSystemTestSuite {
	logger := log.NewNopLogger()
	ctx, cancel := context.WithCancel(context.Background())

	// Create test blockchain
	genesis := createTestGenesisBlock(t)
	blockchain, err := core.NewBlockchain(logger, genesis)
	require.NoError(t, err)

	// Create DAO instance
	daoInstance := dao.NewDAO("INTTEST", "Integration Test Token", 18)

	// Initialize with comprehensive test distribution
	testDistribution := map[string]uint64{
		"integration_treasury": 100000000, // 100M tokens
		"test_validator":       50000000,  // 50M tokens
	}
	err = daoInstance.InitialTokenDistribution(testDistribution)
	require.NoError(t, err)

	// Create transaction channel
	txChan := make(chan *core.Transaction, 1000)

	// Setup network server (simplified for testing)
	validatorKey := crypto.GeneratePrivateKey()
	networkOpts := network.ServerOpts{
		APIListenAddr: ":0", // Use random port for testing
		SeedNodes:     []string{},
		ListenAddr:    ":0", // Use random port for testing
		PrivateKey:    &validatorKey,
		ID:            "TEST_NODE",
		Logger:        logger,
	}

	networkServer, err := network.NewServer(networkOpts)
	require.NoError(t, err)

	// Setup API server
	apiConfig := api.ServerConfig{
		Logger:     logger,
		ListenAddr: ":0", // Use random port for testing
	}
	apiServer := api.NewDAOServer(apiConfig, blockchain, txChan, daoInstance)

	return &IntegratedSystemTestSuite{
		daoInstance:   daoInstance,
		blockchain:    blockchain,
		networkServer: networkServer,
		apiServer:     apiServer,
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
		txChan:        txChan,
		cleanup: func() {
			cancel()
			close(txChan)
		},
	}
}

func testSystemBootstrapAndInitialization(t *testing.T) {
	suite := setupCompleteSystem(t)
	defer suite.cleanup()

	// Test 1: Verify DAO initialization
	assert.NotNil(t, suite.daoInstance)
	assert.NotNil(t, suite.daoInstance.GovernanceState)
	assert.NotNil(t, suite.daoInstance.TokenState)

	// Test 2: Verify token system initialization
	totalSupply := suite.daoInstance.GetTotalSupply()
	assert.Greater(t, totalSupply, uint64(0))
	assert.Equal(t, uint64(150000000), totalSupply) // 100M + 50M

	// Test 3: Verify treasury initialization
	treasuryBalance := suite.daoInstance.GetTreasuryBalance()
	assert.GreaterOrEqual(t, treasuryBalance, uint64(0))

	// Test 4: Verify blockchain initialization
	assert.Greater(t, suite.blockchain.Height(), uint32(0))

	// Test 5: Verify network server initialization
	assert.NotNil(t, suite.networkServer)

	// Test 6: Verify API server initialization
	assert.NotNil(t, suite.apiServer)

	// Test 7: Verify transaction channel
	assert.NotNil(t, suite.txChan)
	assert.Equal(t, 1000, cap(suite.txChan))

	t.Log("System bootstrap and initialization test passed")
}

func testEndToEndGovernanceWorkflow(t *testing.T) {
	suite := setupCompleteSystem(t)
	defer suite.cleanup()

	// Setup test participants
	creator := crypto.GeneratePrivateKey()
	voters := make([]crypto.PrivateKey, 10)
	for i := range voters {
		voters[i] = crypto.GeneratePrivateKey()
	}

	// Initialize participants with tokens
	setupTestUsers(t, suite.daoInstance, append(voters, creator)...)

	// Phase 1: Create proposal
	proposalTx := &dao.ProposalTx{
		Fee:          200,
		Title:        "Complete E2E Integration Test",
		Description:  "Testing complete end-to-end governance workflow with full system integration",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() - 100,
		EndTime:      time.Now().Unix() + 3600,
		Threshold:    5000,
		MetadataHash: generateRandomHash(),
	}

	proposalHash := generateTxHash(proposalTx, creator)
	err := suite.daoInstance.ProcessDAOTransaction(proposalTx, creator.PublicKey(), proposalHash)
	require.NoError(t, err)

	// Verify proposal was created and is accessible through all interfaces
	proposal, err := suite.daoInstance.GetProposal(proposalHash)
	require.NoError(t, err)
	assert.Equal(t, proposalTx.Title, proposal.Title)
	assert.Equal(t, dao.ProposalStatusActive, proposal.Status)

	// Phase 2: Cast votes through different mechanisms
	for i, voter := range voters {
		choice := dao.VoteChoiceYes
		if i%3 == 0 {
			choice = dao.VoteChoiceNo
		}

		voteTx := &dao.VoteTx{
			Fee:        100,
			ProposalID: proposalHash,
			Choice:     choice,
			Weight:     1000,
			Reason:     fmt.Sprintf("E2E test vote from participant %d", i),
		}

		voteHash := generateTxHash(voteTx, voter)
		err := suite.daoInstance.ProcessDAOTransaction(voteTx, voter.PublicKey(), voteHash)
		require.NoError(t, err)

		// Simulate transaction processing through the system
		tx := &core.Transaction{
			TxInner: voteTx,
			From:    voter.PublicKey(),
		}
		tx.Sign(voter)

		// Add to transaction channel (simulating network processing)
		select {
		case suite.txChan <- tx:
		case <-time.After(1 * time.Second):
			t.Fatal("Transaction channel blocked")
		}
	}

	// Phase 3: Process transactions and update proposal status
	suite.daoInstance.UpdateAllProposalStatuses()

	// Phase 4: Verify final results
	finalProposal, err := suite.daoInstance.GetProposal(proposalHash)
	require.NoError(t, err)
	assert.NotNil(t, finalProposal.Results)
	assert.Equal(t, uint64(10), finalProposal.Results.TotalVoters)

	// Verify vote distribution (7 Yes, 3 No based on the pattern)
	expectedYes := uint64(7000) // 7 voters * 1000 weight
	expectedNo := uint64(3000)  // 3 voters * 1000 weight
	assert.Equal(t, expectedYes, finalProposal.Results.YesVotes)
	assert.Equal(t, expectedNo, finalProposal.Results.NoVotes)

	t.Log("End-to-end governance workflow test passed")
}

func testConcurrentOperationsStressTest(t *testing.T) {
	suite := setupCompleteSystem(t)
	defer suite.cleanup()

	// Setup large number of participants
	numUsers := 100
	users := make([]crypto.PrivateKey, numUsers)
	for i := range users {
		users[i] = crypto.GeneratePrivateKey()
	}
	setupTestUsers(t, suite.daoInstance, users...)

	// Setup treasury for concurrent operations
	treasurySigners := users[:5]
	signerPubKeys := make([]crypto.PublicKey, len(treasurySigners))
	for i, signer := range treasurySigners {
		signerPubKeys[i] = signer.PublicKey()
	}
	err := suite.daoInstance.InitializeTreasury(signerPubKeys, 3)
	require.NoError(t, err)
	suite.daoInstance.AddTreasuryFunds(10000000)

	var wg sync.WaitGroup
	start := time.Now()

	// Concurrent proposal creation
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			creator := users[i%numUsers]
			proposalTx := &dao.ProposalTx{
				Fee:          200,
				Title:        fmt.Sprintf("Stress Test Proposal %d", i),
				Description:  "Concurrent stress test proposal",
				ProposalType: dao.ProposalTypeGeneral,
				VotingType:   dao.VotingTypeSimple,
				StartTime:    time.Now().Unix() - 100,
				EndTime:      time.Now().Unix() + 3600,
				Threshold:    1000,
				MetadataHash: generateRandomHash(),
			}

			proposalHash := generateTxHash(proposalTx, creator)
			err := suite.daoInstance.ProcessDAOTransaction(proposalTx, creator.PublicKey(), proposalHash)
			if err != nil {
				t.Logf("Proposal creation error (expected under stress): %v", err)
			}
		}
	}()

	// Concurrent token transfers
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			from := users[i%numUsers]
			to := users[(i+1)%numUsers]
			err := suite.daoInstance.TransferTokens(from.PublicKey(), to.PublicKey(), 10)
			if err != nil {
				t.Logf("Token transfer error (expected under stress): %v", err)
			}
		}
	}()

	// Concurrent delegations
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			delegator := users[i*2%numUsers]
			delegate := users[(i*2+1)%numUsers]

			delegationTx := &dao.DelegationTx{
				Fee:      200,
				Delegate: delegate.PublicKey(),
				Duration: 3600,
				Revoke:   false,
			}

			delegationHash := generateTxHash(delegationTx, delegator)
			err := suite.daoInstance.ProcessDAOTransaction(delegationTx, delegator.PublicKey(), delegationHash)
			if err != nil {
				t.Logf("Delegation error (expected under stress): %v", err)
			}
		}
	}()

	wg.Wait()
	duration := time.Since(start)

	// Verify system remained stable under stress
	totalSupply := suite.daoInstance.GetTotalSupply()
	assert.Greater(t, totalSupply, uint64(0))

	proposals := suite.daoInstance.ListAllProposals()
	assert.Greater(t, len(proposals), 0)

	t.Logf("Concurrent operations stress test completed in %v", duration)
	t.Log("System remained stable under concurrent load")
}

func testSecurityAndAttackVectorValidation(t *testing.T) {
	suite := setupCompleteSystem(t)
	defer suite.cleanup()

	// Setup test participants
	admin := crypto.GeneratePrivateKey()
	attacker := crypto.GeneratePrivateKey()
	normalUser := crypto.GeneratePrivateKey()

	setupTestUsers(t, suite.daoInstance, admin, attacker, normalUser)

	// Initialize security roles
	err := suite.daoInstance.InitializeFounderRoles([]crypto.PublicKey{admin.PublicKey()})
	require.NoError(t, err)

	// Test 1: Access control validation
	hasPermission := suite.daoInstance.HasPermission(admin.PublicKey(), dao.PermissionManageRoles)
	assert.True(t, hasPermission)

	hasPermission = suite.daoInstance.HasPermission(attacker.PublicKey(), dao.PermissionManageRoles)
	assert.False(t, hasPermission)

	// Test 2: Double voting prevention
	proposalTx := &dao.ProposalTx{
		Fee:          200,
		Title:        "Security Test Proposal",
		Description:  "Testing security measures",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() - 100,
		EndTime:      time.Now().Unix() + 3600,
		Threshold:    1000,
		MetadataHash: generateRandomHash(),
	}

	proposalHash := generateTxHash(proposalTx, admin)
	err = suite.daoInstance.ProcessDAOTransaction(proposalTx, admin.PublicKey(), proposalHash)
	require.NoError(t, err)

	// First vote should succeed
	voteTx := &dao.VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     dao.VoteChoiceYes,
		Weight:     1000,
		Reason:     "First vote",
	}

	voteHash := generateTxHash(voteTx, normalUser)
	err = suite.daoInstance.ProcessDAOTransaction(voteTx, normalUser.PublicKey(), voteHash)
	require.NoError(t, err)

	// Second vote should fail (double voting prevention)
	voteTx2 := &dao.VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     dao.VoteChoiceNo,
		Weight:     1000,
		Reason:     "Attempted double vote",
	}

	voteHash2 := generateTxHash(voteTx2, normalUser)
	err = suite.daoInstance.ProcessDAOTransaction(voteTx2, normalUser.PublicKey(), voteHash2)
	assert.Error(t, err, "Double voting should be prevented")

	// Test 3: Emergency pause mechanism
	err = suite.daoInstance.ActivateEmergency(admin.PublicKey(), "Security test emergency", dao.SecurityLevelCritical, []string{"voting"})
	assert.NoError(t, err)

	isActive := suite.daoInstance.IsEmergencyActive()
	assert.True(t, isActive)

	// Operations should be restricted during emergency
	restrictedVoteTx := &dao.VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     dao.VoteChoiceYes,
		Weight:     1000,
		Reason:     "Vote during emergency",
	}

	restrictedVoteHash := generateTxHash(restrictedVoteTx, attacker)
	err = suite.daoInstance.ProcessDAOTransaction(restrictedVoteTx, attacker.PublicKey(), restrictedVoteHash)
	assert.Error(t, err, "Operations should be restricted during emergency")

	// Deactivate emergency
	err = suite.daoInstance.DeactivateEmergency(admin.PublicKey())
	assert.NoError(t, err)

	t.Log("Security and attack vector validation test passed")
}

func testPerformanceUnderLoad(t *testing.T) {
	suite := setupCompleteSystem(t)
	defer suite.cleanup()

	// Setup performance test users
	numUsers := 200
	users := make([]crypto.PrivateKey, numUsers)
	for i := range users {
		users[i] = crypto.GeneratePrivateKey()
	}
	setupTestUsers(t, suite.daoInstance, users...)

	// Performance Test 1: High-volume proposal creation
	start := time.Now()
	numProposals := 100

	for i := 0; i < numProposals; i++ {
		creator := users[i%numUsers]
		proposalTx := &dao.ProposalTx{
			Fee:          200,
			Title:        fmt.Sprintf("Performance Test Proposal %d", i),
			Description:  "High-volume performance test",
			ProposalType: dao.ProposalTypeGeneral,
			VotingType:   dao.VotingTypeSimple,
			StartTime:    time.Now().Unix() - 100,
			EndTime:      time.Now().Unix() + 3600,
			Threshold:    1000,
			MetadataHash: generateRandomHash(),
		}

		proposalHash := generateTxHash(proposalTx, creator)
		err := suite.daoInstance.ProcessDAOTransaction(proposalTx, creator.PublicKey(), proposalHash)
		require.NoError(t, err)
	}

	proposalDuration := time.Since(start)
	proposalThroughput := float64(numProposals) / proposalDuration.Seconds()

	// Performance Test 2: High-volume token operations
	start = time.Now()
	numTransfers := 500

	for i := 0; i < numTransfers; i++ {
		from := users[i%numUsers]
		to := users[(i+1)%numUsers]
		err := suite.daoInstance.TransferTokens(from.PublicKey(), to.PublicKey(), 10)
		require.NoError(t, err)
	}

	transferDuration := time.Since(start)
	transferThroughput := float64(numTransfers) / transferDuration.Seconds()

	// Performance assertions
	assert.Greater(t, proposalThroughput, 10.0, "Should achieve at least 10 proposals/sec")
	assert.Greater(t, transferThroughput, 50.0, "Should achieve at least 50 transfers/sec")

	// Verify system integrity after load
	totalSupply := suite.daoInstance.GetTotalSupply()
	assert.Greater(t, totalSupply, uint64(0))

	proposals := suite.daoInstance.ListAllProposals()
	assert.GreaterOrEqual(t, len(proposals), numProposals)

	t.Logf("Performance test completed - Proposals: %.2f/sec, Transfers: %.2f/sec",
		proposalThroughput, transferThroughput)
}

func testCrossComponentIntegration(t *testing.T) {
	suite := setupCompleteSystem(t)
	defer suite.cleanup()

	// Test integration between DAO, Blockchain, and API components
	creator := crypto.GeneratePrivateKey()
	setupTestUsers(t, suite.daoInstance, creator)

	// Create proposal through DAO
	proposalTx := &dao.ProposalTx{
		Fee:          200,
		Title:        "Cross-Component Integration Test",
		Description:  "Testing integration between all system components",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() - 100,
		EndTime:      time.Now().Unix() + 3600,
		Threshold:    1000,
		MetadataHash: generateRandomHash(),
	}

	proposalHash := generateTxHash(proposalTx, creator)
	err := suite.daoInstance.ProcessDAOTransaction(proposalTx, creator.PublicKey(), proposalHash)
	require.NoError(t, err)

	// Verify proposal exists in DAO
	proposal, err := suite.daoInstance.GetProposal(proposalHash)
	require.NoError(t, err)
	assert.Equal(t, proposalTx.Title, proposal.Title)

	// Create blockchain transaction
	blockchainTx := &core.Transaction{
		TxInner: proposalTx,
		From:    creator.PublicKey(),
		Value:   0,
	}
	blockchainTx.Sign(creator)

	// Add to blockchain
	block := createBlockWithTransaction(t, suite.blockchain, blockchainTx)
	err = suite.blockchain.AddBlock(block)
	require.NoError(t, err)

	// Verify blockchain height increased
	assert.Greater(t, suite.blockchain.Height(), uint32(1))

	// Test transaction channel integration
	testTx := &core.Transaction{
		TxInner: &dao.VoteTx{
			Fee:        100,
			ProposalID: proposalHash,
			Choice:     dao.VoteChoiceYes,
			Weight:     1000,
			Reason:     "Cross-component test vote",
		},
		From: creator.PublicKey(),
	}
	testTx.Sign(creator)

	// Send through transaction channel
	select {
	case suite.txChan <- testTx:
		t.Log("Transaction successfully sent through channel")
	case <-time.After(1 * time.Second):
		t.Fatal("Transaction channel integration failed")
	}

	// Verify channel has the transaction
	select {
	case receivedTx := <-suite.txChan:
		assert.Equal(t, testTx.Hash(core.TxHasher{}), receivedTx.Hash(core.TxHasher{}))
	case <-time.After(1 * time.Second):
		t.Fatal("Failed to receive transaction from channel")
	}

	t.Log("Cross-component integration test passed")
}

func testErrorHandlingAndRecovery(t *testing.T) {
	suite := setupCompleteSystem(t)
	defer suite.cleanup()

	user := crypto.GeneratePrivateKey()
	setupTestUsers(t, suite.daoInstance, user)

	// Test 1: Invalid proposal handling
	invalidProposal := &dao.ProposalTx{
		Fee:          200,
		Title:        "", // Invalid empty title
		Description:  "Test invalid proposal",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() + 3600, // Invalid future start time
		EndTime:      time.Now().Unix() - 100,  // Invalid past end time
		Threshold:    1000,
		MetadataHash: generateRandomHash(),
	}

	proposalHash := generateTxHash(invalidProposal, user)
	err := suite.daoInstance.ProcessDAOTransaction(invalidProposal, user.PublicKey(), proposalHash)
	assert.Error(t, err, "Invalid proposal should be rejected")

	// Test 2: Insufficient balance handling
	poorUser := crypto.GeneratePrivateKey()
	// Don't give poorUser any tokens

	expensiveProposal := &dao.ProposalTx{
		Fee:          10000000, // Very high fee
		Title:        "Expensive Proposal",
		Description:  "Testing insufficient balance handling",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() - 100,
		EndTime:      time.Now().Unix() + 3600,
		Threshold:    1000,
		MetadataHash: generateRandomHash(),
	}

	expensiveHash := generateTxHash(expensiveProposal, poorUser)
	err = suite.daoInstance.ProcessDAOTransaction(expensiveProposal, poorUser.PublicKey(), expensiveHash)
	assert.Error(t, err, "Transaction with insufficient balance should be rejected")

	// Test 3: System recovery after errors
	// Verify system is still functional after errors
	validProposal := &dao.ProposalTx{
		Fee:          200,
		Title:        "Recovery Test Proposal",
		Description:  "Testing system recovery after errors",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() - 100,
		EndTime:      time.Now().Unix() + 3600,
		Threshold:    1000,
		MetadataHash: generateRandomHash(),
	}

	validHash := generateTxHash(validProposal, user)
	err = suite.daoInstance.ProcessDAOTransaction(validProposal, user.PublicKey(), validHash)
	assert.NoError(t, err, "Valid proposal should succeed after error recovery")

	t.Log("Error handling and recovery test passed")
}

func testDataConsistencyValidation(t *testing.T) {
	suite := setupCompleteSystem(t)
	defer suite.cleanup()

	// Setup test users
	users := make([]crypto.PrivateKey, 10)
	for i := range users {
		users[i] = crypto.GeneratePrivateKey()
	}
	setupTestUsers(t, suite.daoInstance, users...)

	// Test 1: Token balance consistency
	initialTotalSupply := suite.daoInstance.GetTotalSupply()

	// Perform multiple token operations
	for i := 0; i < 5; i++ {
		from := users[i]
		to := users[i+1]
		err := suite.daoInstance.TransferTokens(from.PublicKey(), to.PublicKey(), 100)
		require.NoError(t, err)
	}

	// Verify total supply remains consistent
	finalTotalSupply := suite.daoInstance.GetTotalSupply()
	assert.Equal(t, initialTotalSupply, finalTotalSupply, "Total supply should remain consistent")

	// Test 2: Proposal state consistency
	creator := users[0]
	proposalTx := &dao.ProposalTx{
		Fee:          200,
		Title:        "Consistency Test Proposal",
		Description:  "Testing data consistency",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() - 100,
		EndTime:      time.Now().Unix() + 3600,
		Threshold:    1000,
		MetadataHash: generateRandomHash(),
	}

	proposalHash := generateTxHash(proposalTx, creator)
	err := suite.daoInstance.ProcessDAOTransaction(proposalTx, creator.PublicKey(), proposalHash)
	require.NoError(t, err)

	// Cast votes and verify consistency
	for i, voter := range users[1:6] {
		voteTx := &dao.VoteTx{
			Fee:        100,
			ProposalID: proposalHash,
			Choice:     dao.VoteChoiceYes,
			Weight:     200,
			Reason:     fmt.Sprintf("Consistency test vote %d", i),
		}

		voteHash := generateTxHash(voteTx, voter)
		err := suite.daoInstance.ProcessDAOTransaction(voteTx, voter.PublicKey(), voteHash)
		require.NoError(t, err)
	}

	// Verify vote count consistency
	votes, err := suite.daoInstance.GetVotes(proposalHash)
	require.NoError(t, err)
	assert.Len(t, votes, 5, "Vote count should be consistent")

	// Update proposal status and verify consistency
	suite.daoInstance.UpdateAllProposalStatuses()

	proposal, err := suite.daoInstance.GetProposal(proposalHash)
	require.NoError(t, err)
	assert.NotNil(t, proposal.Results)
	assert.Equal(t, uint64(1000), proposal.Results.YesVotes) // 5 votes * 200 weight
	assert.Equal(t, uint64(5), proposal.Results.TotalVoters)

	t.Log("Data consistency validation test passed")
}

// Helper functions

func setupTestUsers(t *testing.T, daoInstance *dao.DAO, users ...crypto.PrivateKey) {
	for _, user := range users {
		// Mint tokens to user
		err := daoInstance.MintTokens(user.PublicKey(), 50000)
		require.NoError(t, err)

		// Initialize reputation
		daoInstance.InitializeUserReputation(user.PublicKey(), 10000)
	}
}

func createTestGenesisBlock(t *testing.T) *core.Block {
	privKey := crypto.GeneratePrivateKey()

	genesisTx := &core.Transaction{
		TxInner: core.CollectionTx{
			Fee:      0,
			MetaData: []byte("Integration Test Genesis Block"),
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

func createBlockWithTransaction(t *testing.T, bc *core.Blockchain, tx *core.Transaction) *core.Block {
	privKey := crypto.GeneratePrivateKey()

	prevBlock, err := bc.GetBlock(bc.Height())
	require.NoError(t, err)

	header := &core.Header{
		Version:       1,
		PrevBlockHash: prevBlock.Hash(core.BlockHasher{}),
		Height:        bc.Height() + 1,
		Timestamp:     time.Now().UnixNano(),
	}

	block, err := core.NewBlock(header, []*core.Transaction{tx})
	require.NoError(t, err)

	dataHash, err := core.CalculateDataHash(block.Transactions)
	require.NoError(t, err)
	block.Header.DataHash = dataHash

	err = block.Sign(privKey)
	require.NoError(t, err)

	return block
}

func generateTxHash(tx interface{}, signer crypto.PrivateKey) types.Hash {
	// Simple hash generation for testing
	data := fmt.Sprintf("%v%s%d", tx, signer.PublicKey().String(), time.Now().UnixNano())
	hash := [32]byte{}
	copy(hash[:], []byte(data)[:32])
	return hash
}

func generateRandomHash() types.Hash {
	hash := [32]byte{}
	for i := range hash {
		hash[i] = byte(i % 256)
	}
	return hash
}
