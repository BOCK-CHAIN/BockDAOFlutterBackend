package tests

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/BOCK-CHAIN/BockChain/core"
	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/dao"
	"github.com/go-kit/log"
)

// SystemValidationRunner orchestrates comprehensive system validation
type SystemValidationRunner struct {
	logger    log.Logger
	results   map[string]*ValidationResult
	startTime time.Time
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
}

// ValidationResult represents the result of a validation test
type ValidationResult struct {
	TestName string
	Passed   bool
	Duration time.Duration
	Error    error
	Metrics  map[string]interface{}
	Details  string
}

// NewSystemValidationRunner creates a new validation runner
func NewSystemValidationRunner() *SystemValidationRunner {
	logger := log.NewLogfmtLogger(os.Stdout)
	ctx, cancel := context.WithCancel(context.Background())

	return &SystemValidationRunner{
		logger:    logger,
		results:   make(map[string]*ValidationResult),
		startTime: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// RunCompleteSystemValidation executes all system validation tests
func (r *SystemValidationRunner) RunCompleteSystemValidation() error {
	r.logger.Log("msg", "Starting complete system validation")

	// Run validation tests in sequence
	validationTests := []struct {
		name string
		fn   func() error
	}{
		{"ComponentInitialization", r.validateComponentInitialization},
		{"CoreFunctionality", r.validateCoreFunctionality},
		{"IntegrationPoints", r.validateIntegrationPoints},
		{"PerformanceMetrics", r.validatePerformanceMetrics},
		{"SecurityMeasures", r.validateSecurityMeasures},
		{"ErrorHandling", r.validateErrorHandling},
		{"DataIntegrity", r.validateDataIntegrity},
		{"ScalabilityLimits", r.validateScalabilityLimits},
		{"ResourceUtilization", r.validateResourceUtilization},
		{"SystemResilience", r.validateSystemResilience},
	}

	for _, test := range validationTests {
		r.runValidationTest(test.name, test.fn)
	}

	// Generate final report
	return r.generateValidationReport()
}

// runValidationTest executes a single validation test with error handling
func (r *SystemValidationRunner) runValidationTest(name string, testFn func() error) {
	start := time.Now()

	r.logger.Log("msg", "Running validation test", "test", name)

	result := &ValidationResult{
		TestName: name,
		Metrics:  make(map[string]interface{}),
	}

	defer func() {
		if rec := recover(); rec != nil {
			result.Passed = false
			result.Error = fmt.Errorf("test panicked: %v", rec)
			result.Details = "Test encountered a panic during execution"
		}

		result.Duration = time.Since(start)

		r.mu.Lock()
		r.results[name] = result
		r.mu.Unlock()

		status := "PASSED"
		if !result.Passed {
			status = "FAILED"
		}

		r.logger.Log("msg", "Validation test completed", "test", name, "status", status, "duration", result.Duration)
	}()

	err := testFn()
	if err != nil {
		result.Passed = false
		result.Error = err
		result.Details = err.Error()
	} else {
		result.Passed = true
		result.Details = "Test completed successfully"
	}
}

// validateComponentInitialization validates that all components initialize correctly
func (r *SystemValidationRunner) validateComponentInitialization() error {
	// Create test blockchain
	logger := log.NewNopLogger()
	genesis := r.createTestGenesisBlock()
	blockchain, err := core.NewBlockchain(logger, genesis)
	if err != nil {
		return fmt.Errorf("blockchain initialization failed: %w", err)
	}

	// Create DAO instance
	daoInstance := dao.NewDAO("VALIDATION", "Validation Token", 18)
	if daoInstance == nil {
		return fmt.Errorf("DAO instance creation failed")
	}

	// Initialize token distribution
	testDistribution := map[string]uint64{
		"validation_treasury": 1000000,
	}
	err = daoInstance.InitialTokenDistribution(testDistribution)
	if err != nil {
		return fmt.Errorf("token distribution initialization failed: %w", err)
	}

	// Validate component states
	if blockchain.Height() == 0 {
		return fmt.Errorf("blockchain height should be greater than 0")
	}

	if daoInstance.GetTotalSupply() == 0 {
		return fmt.Errorf("total token supply should be greater than 0")
	}

	r.results["ComponentInitialization"].Metrics["blockchain_height"] = blockchain.Height()
	r.results["ComponentInitialization"].Metrics["total_supply"] = daoInstance.GetTotalSupply()

	return nil
}

// validateCoreFunctionality validates core DAO functionality
func (r *SystemValidationRunner) validateCoreFunctionality() error {
	// Setup test environment
	daoInstance := dao.NewDAO("CORE_TEST", "Core Test Token", 18)

	testDistribution := map[string]uint64{
		"core_treasury": 10000000,
	}
	err := daoInstance.InitialTokenDistribution(testDistribution)
	if err != nil {
		return fmt.Errorf("failed to initialize test distribution: %w", err)
	}

	// Test user setup
	user1 := crypto.GeneratePrivateKey()
	user2 := crypto.GeneratePrivateKey()

	// Test token minting
	err = daoInstance.MintTokens(user1.PublicKey(), 10000)
	if err != nil {
		return fmt.Errorf("token minting failed: %w", err)
	}

	err = daoInstance.MintTokens(user2.PublicKey(), 10000)
	if err != nil {
		return fmt.Errorf("token minting failed: %w", err)
	}

	// Test token transfer
	err = daoInstance.TransferTokens(user1.PublicKey(), user2.PublicKey(), 1000)
	if err != nil {
		return fmt.Errorf("token transfer failed: %w", err)
	}

	// Verify balances
	balance1 := daoInstance.GetTokenBalance(user1.PublicKey())
	balance2 := daoInstance.GetTokenBalance(user2.PublicKey())

	if balance1 != 9000 {
		return fmt.Errorf("user1 balance incorrect: expected 9000, got %d", balance1)
	}

	if balance2 != 11000 {
		return fmt.Errorf("user2 balance incorrect: expected 11000, got %d", balance2)
	}

	// Test proposal creation
	proposalTx := &dao.ProposalTx{
		Fee:          200,
		Title:        "Core Functionality Test",
		Description:  "Testing core proposal functionality",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() - 100,
		EndTime:      time.Now().Unix() + 3600,
		Threshold:    1000,
		MetadataHash: r.generateRandomHash(),
	}

	proposalHash := r.generateTxHash(proposalTx, user1)
	err = daoInstance.ProcessDAOTransaction(proposalTx, user1.PublicKey(), proposalHash)
	if err != nil {
		return fmt.Errorf("proposal creation failed: %w", err)
	}

	// Verify proposal exists
	proposal, err := daoInstance.GetProposal(proposalHash)
	if err != nil {
		return fmt.Errorf("proposal retrieval failed: %w", err)
	}

	if proposal.Title != proposalTx.Title {
		return fmt.Errorf("proposal title mismatch")
	}

	// Test voting
	voteTx := &dao.VoteTx{
		Fee:        100,
		ProposalID: proposalHash,
		Choice:     dao.VoteChoiceYes,
		Weight:     1000,
		Reason:     "Core functionality test vote",
	}

	voteHash := r.generateTxHash(voteTx, user2)
	err = daoInstance.ProcessDAOTransaction(voteTx, user2.PublicKey(), voteHash)
	if err != nil {
		return fmt.Errorf("voting failed: %w", err)
	}

	// Verify vote was recorded
	votes, err := daoInstance.GetVotes(proposalHash)
	if err != nil {
		return fmt.Errorf("vote retrieval failed: %w", err)
	}

	if len(votes) != 1 {
		return fmt.Errorf("expected 1 vote, got %d", len(votes))
	}

	r.results["CoreFunctionality"].Metrics["proposals_created"] = 1
	r.results["CoreFunctionality"].Metrics["votes_cast"] = len(votes)
	r.results["CoreFunctionality"].Metrics["token_transfers"] = 1

	return nil
}

// validateIntegrationPoints validates integration between components
func (r *SystemValidationRunner) validateIntegrationPoints() error {
	// Test DAO-Blockchain integration
	logger := log.NewNopLogger()
	genesis := r.createTestGenesisBlock()
	blockchain, err := core.NewBlockchain(logger, genesis)
	if err != nil {
		return fmt.Errorf("blockchain creation failed: %w", err)
	}

	daoInstance := dao.NewDAO("INTEGRATION", "Integration Token", 18)
	testDistribution := map[string]uint64{
		"integration_treasury": 5000000,
	}
	err = daoInstance.InitialTokenDistribution(testDistribution)
	if err != nil {
		return fmt.Errorf("DAO initialization failed: %w", err)
	}

	// Test transaction processing integration
	user := crypto.GeneratePrivateKey()
	err = daoInstance.MintTokens(user.PublicKey(), 10000)
	if err != nil {
		return fmt.Errorf("token minting failed: %w", err)
	}

	// Create a DAO transaction
	proposalTx := &dao.ProposalTx{
		Fee:          200,
		Title:        "Integration Test Proposal",
		Description:  "Testing DAO-Blockchain integration",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() - 100,
		EndTime:      time.Now().Unix() + 3600,
		Threshold:    1000,
		MetadataHash: r.generateRandomHash(),
	}

	proposalHash := r.generateTxHash(proposalTx, user)
	err = daoInstance.ProcessDAOTransaction(proposalTx, user.PublicKey(), proposalHash)
	if err != nil {
		return fmt.Errorf("DAO transaction processing failed: %w", err)
	}

	// Create corresponding blockchain transaction
	blockchainTx := &core.Transaction{
		TxInner: proposalTx,
		From:    user.PublicKey(),
		Value:   0,
	}
	blockchainTx.Sign(user)

	// Create block with transaction
	block := r.createBlockWithTransaction(blockchain, blockchainTx)
	err = blockchain.AddBlock(block)
	if err != nil {
		return fmt.Errorf("blockchain transaction processing failed: %w", err)
	}

	// Verify integration
	if blockchain.Height() <= 1 {
		return fmt.Errorf("blockchain height should have increased")
	}

	proposal, err := daoInstance.GetProposal(proposalHash)
	if err != nil {
		return fmt.Errorf("proposal should exist in DAO: %w", err)
	}

	if proposal.Status != dao.ProposalStatusActive {
		return fmt.Errorf("proposal should be active")
	}

	r.results["IntegrationPoints"].Metrics["blockchain_height"] = blockchain.Height()
	r.results["IntegrationPoints"].Metrics["dao_proposals"] = len(daoInstance.ListAllProposals())

	return nil
}

// validatePerformanceMetrics validates system performance under normal load
func (r *SystemValidationRunner) validatePerformanceMetrics() error {
	daoInstance := dao.NewDAO("PERF", "Performance Token", 18)

	testDistribution := map[string]uint64{
		"perf_treasury": 50000000,
	}
	err := daoInstance.InitialTokenDistribution(testDistribution)
	if err != nil {
		return fmt.Errorf("performance test setup failed: %w", err)
	}

	// Setup test users
	numUsers := 50
	users := make([]crypto.PrivateKey, numUsers)
	for i := range users {
		users[i] = crypto.GeneratePrivateKey()
		err = daoInstance.MintTokens(users[i].PublicKey(), 10000)
		if err != nil {
			return fmt.Errorf("user setup failed: %w", err)
		}
	}

	// Performance test: Proposal creation
	start := time.Now()
	numProposals := 20

	for i := 0; i < numProposals; i++ {
		creator := users[i%numUsers]
		proposalTx := &dao.ProposalTx{
			Fee:          200,
			Title:        fmt.Sprintf("Performance Test Proposal %d", i),
			Description:  "Performance testing proposal",
			ProposalType: dao.ProposalTypeGeneral,
			VotingType:   dao.VotingTypeSimple,
			StartTime:    time.Now().Unix() - 100,
			EndTime:      time.Now().Unix() + 3600,
			Threshold:    1000,
			MetadataHash: r.generateRandomHash(),
		}

		proposalHash := r.generateTxHash(proposalTx, creator)
		err := daoInstance.ProcessDAOTransaction(proposalTx, creator.PublicKey(), proposalHash)
		if err != nil {
			return fmt.Errorf("proposal creation failed during performance test: %w", err)
		}
	}

	proposalDuration := time.Since(start)
	proposalThroughput := float64(numProposals) / proposalDuration.Seconds()

	// Performance test: Token transfers
	start = time.Now()
	numTransfers := 100

	for i := 0; i < numTransfers; i++ {
		from := users[i%numUsers]
		to := users[(i+1)%numUsers]
		err := daoInstance.TransferTokens(from.PublicKey(), to.PublicKey(), 10)
		if err != nil {
			return fmt.Errorf("token transfer failed during performance test: %w", err)
		}
	}

	transferDuration := time.Since(start)
	transferThroughput := float64(numTransfers) / transferDuration.Seconds()

	// Performance assertions
	minProposalThroughput := 5.0  // 5 proposals/sec minimum
	minTransferThroughput := 20.0 // 20 transfers/sec minimum

	if proposalThroughput < minProposalThroughput {
		return fmt.Errorf("proposal throughput too low: %.2f < %.2f", proposalThroughput, minProposalThroughput)
	}

	if transferThroughput < minTransferThroughput {
		return fmt.Errorf("transfer throughput too low: %.2f < %.2f", transferThroughput, minTransferThroughput)
	}

	r.results["PerformanceMetrics"].Metrics["proposal_throughput"] = proposalThroughput
	r.results["PerformanceMetrics"].Metrics["transfer_throughput"] = transferThroughput
	r.results["PerformanceMetrics"].Metrics["proposal_duration_ms"] = proposalDuration.Milliseconds()
	r.results["PerformanceMetrics"].Metrics["transfer_duration_ms"] = transferDuration.Milliseconds()

	return nil
}

// validateSecurityMeasures validates security controls and access restrictions
func (r *SystemValidationRunner) validateSecurityMeasures() error {
	daoInstance := dao.NewDAO("SECURITY", "Security Token", 18)

	testDistribution := map[string]uint64{
		"security_treasury": 10000000,
	}
	err := daoInstance.InitialTokenDistribution(testDistribution)
	if err != nil {
		return fmt.Errorf("security test setup failed: %w", err)
	}

	// Setup test users
	admin := crypto.GeneratePrivateKey()
	normalUser := crypto.GeneratePrivateKey()
	attacker := crypto.GeneratePrivateKey()

	err = daoInstance.MintTokens(admin.PublicKey(), 50000)
	if err != nil {
		return fmt.Errorf("admin setup failed: %w", err)
	}

	err = daoInstance.MintTokens(normalUser.PublicKey(), 10000)
	if err != nil {
		return fmt.Errorf("normal user setup failed: %w", err)
	}

	// Don't give attacker any tokens initially

	// Initialize security roles
	err = daoInstance.InitializeFounderRoles([]crypto.PublicKey{admin.PublicKey()})
	if err != nil {
		return fmt.Errorf("founder role initialization failed: %w", err)
	}

	// Test access control
	hasPermission := daoInstance.HasPermission(admin.PublicKey(), dao.PermissionManageRoles)
	if !hasPermission {
		return fmt.Errorf("admin should have manage roles permission")
	}

	hasPermission = daoInstance.HasPermission(attacker.PublicKey(), dao.PermissionManageRoles)
	if hasPermission {
		return fmt.Errorf("attacker should not have manage roles permission")
	}

	// Test insufficient balance protection
	expensiveProposal := &dao.ProposalTx{
		Fee:          1000000, // Very high fee
		Title:        "Expensive Attack Proposal",
		Description:  "Attempting to create expensive proposal without funds",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() - 100,
		EndTime:      time.Now().Unix() + 3600,
		Threshold:    1000,
		MetadataHash: r.generateRandomHash(),
	}

	attackHash := r.generateTxHash(expensiveProposal, attacker)
	err = daoInstance.ProcessDAOTransaction(expensiveProposal, attacker.PublicKey(), attackHash)
	if err == nil {
		return fmt.Errorf("expensive proposal should have been rejected due to insufficient funds")
	}

	// Test emergency pause mechanism
	err = daoInstance.ActivateEmergency(admin.PublicKey(), "Security validation test", dao.SecurityLevelCritical, []string{"voting"})
	if err != nil {
		return fmt.Errorf("emergency activation failed: %w", err)
	}

	isActive := daoInstance.IsEmergencyActive()
	if !isActive {
		return fmt.Errorf("emergency should be active")
	}

	// Test that operations are restricted during emergency
	normalProposal := &dao.ProposalTx{
		Fee:          200,
		Title:        "Normal Proposal During Emergency",
		Description:  "This should be blocked during emergency",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() - 100,
		EndTime:      time.Now().Unix() + 3600,
		Threshold:    1000,
		MetadataHash: r.generateRandomHash(),
	}

	normalHash := r.generateTxHash(normalProposal, normalUser)
	err = daoInstance.ProcessDAOTransaction(normalProposal, normalUser.PublicKey(), normalHash)
	if err == nil {
		return fmt.Errorf("proposal should have been blocked during emergency")
	}

	// Deactivate emergency
	err = daoInstance.DeactivateEmergency(admin.PublicKey())
	if err != nil {
		return fmt.Errorf("emergency deactivation failed: %w", err)
	}

	r.results["SecurityMeasures"].Metrics["access_control_tests"] = 2
	r.results["SecurityMeasures"].Metrics["emergency_tests"] = 1
	r.results["SecurityMeasures"].Metrics["insufficient_balance_tests"] = 1

	return nil
}

// validateErrorHandling validates system error handling and recovery
func (r *SystemValidationRunner) validateErrorHandling() error {
	daoInstance := dao.NewDAO("ERROR", "Error Test Token", 18)

	testDistribution := map[string]uint64{
		"error_treasury": 5000000,
	}
	err := daoInstance.InitialTokenDistribution(testDistribution)
	if err != nil {
		return fmt.Errorf("error test setup failed: %w", err)
	}

	user := crypto.GeneratePrivateKey()
	err = daoInstance.MintTokens(user.PublicKey(), 10000)
	if err != nil {
		return fmt.Errorf("user setup failed: %w", err)
	}

	// Test invalid proposal handling
	invalidProposal := &dao.ProposalTx{
		Fee:          200,
		Title:        "", // Invalid empty title
		Description:  "Invalid proposal test",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() + 3600, // Invalid: start time in future
		EndTime:      time.Now().Unix() - 100,  // Invalid: end time in past
		Threshold:    1000,
		MetadataHash: r.generateRandomHash(),
	}

	invalidHash := r.generateTxHash(invalidProposal, user)
	err = daoInstance.ProcessDAOTransaction(invalidProposal, user.PublicKey(), invalidHash)
	if err == nil {
		return fmt.Errorf("invalid proposal should have been rejected")
	}

	// Test system recovery - valid operation after error
	validProposal := &dao.ProposalTx{
		Fee:          200,
		Title:        "Valid Recovery Proposal",
		Description:  "Testing system recovery after error",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() - 100,
		EndTime:      time.Now().Unix() + 3600,
		Threshold:    1000,
		MetadataHash: r.generateRandomHash(),
	}

	validHash := r.generateTxHash(validProposal, user)
	err = daoInstance.ProcessDAOTransaction(validProposal, user.PublicKey(), validHash)
	if err != nil {
		return fmt.Errorf("valid proposal should succeed after error recovery: %w", err)
	}

	// Verify system state is consistent
	proposals := daoInstance.ListAllProposals()
	if len(proposals) != 1 {
		return fmt.Errorf("should have exactly 1 valid proposal, got %d", len(proposals))
	}

	r.results["ErrorHandling"].Metrics["invalid_proposals_rejected"] = 1
	r.results["ErrorHandling"].Metrics["recovery_operations"] = 1

	return nil
}

// validateDataIntegrity validates data consistency and integrity
func (r *SystemValidationRunner) validateDataIntegrity() error {
	daoInstance := dao.NewDAO("INTEGRITY", "Integrity Token", 18)

	testDistribution := map[string]uint64{
		"integrity_treasury": 20000000,
	}
	err := daoInstance.InitialTokenDistribution(testDistribution)
	if err != nil {
		return fmt.Errorf("integrity test setup failed: %w", err)
	}

	// Setup test users
	users := make([]crypto.PrivateKey, 5)
	for i := range users {
		users[i] = crypto.GeneratePrivateKey()
		err = daoInstance.MintTokens(users[i].PublicKey(), 10000)
		if err != nil {
			return fmt.Errorf("user %d setup failed: %w", i, err)
		}
	}

	// Test token balance integrity
	initialTotalSupply := daoInstance.GetTotalSupply()

	// Perform multiple token operations
	for i := 0; i < 4; i++ {
		from := users[i]
		to := users[i+1]
		err := daoInstance.TransferTokens(from.PublicKey(), to.PublicKey(), 500)
		if err != nil {
			return fmt.Errorf("token transfer %d failed: %w", i, err)
		}
	}

	// Verify total supply integrity
	finalTotalSupply := daoInstance.GetTotalSupply()
	if initialTotalSupply != finalTotalSupply {
		return fmt.Errorf("total supply integrity violated: %d != %d", initialTotalSupply, finalTotalSupply)
	}

	// Test proposal-vote integrity
	creator := users[0]
	proposalTx := &dao.ProposalTx{
		Fee:          200,
		Title:        "Integrity Test Proposal",
		Description:  "Testing data integrity",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix() - 100,
		EndTime:      time.Now().Unix() + 3600,
		Threshold:    1000,
		MetadataHash: r.generateRandomHash(),
	}

	proposalHash := r.generateTxHash(proposalTx, creator)
	err = daoInstance.ProcessDAOTransaction(proposalTx, creator.PublicKey(), proposalHash)
	if err != nil {
		return fmt.Errorf("proposal creation failed: %w", err)
	}

	// Cast votes and verify integrity
	expectedVotes := 0
	expectedWeight := uint64(0)

	for i, voter := range users[1:] {
		weight := uint64((i + 1) * 100)
		voteTx := &dao.VoteTx{
			Fee:        100,
			ProposalID: proposalHash,
			Choice:     dao.VoteChoiceYes,
			Weight:     weight,
			Reason:     fmt.Sprintf("Integrity test vote %d", i),
		}

		voteHash := r.generateTxHash(voteTx, voter)
		err := daoInstance.ProcessDAOTransaction(voteTx, voter.PublicKey(), voteHash)
		if err != nil {
			return fmt.Errorf("vote %d failed: %w", i, err)
		}

		expectedVotes++
		expectedWeight += weight
	}

	// Verify vote integrity
	votes, err := daoInstance.GetVotes(proposalHash)
	if err != nil {
		return fmt.Errorf("vote retrieval failed: %w", err)
	}

	if len(votes) != expectedVotes {
		return fmt.Errorf("vote count integrity violated: expected %d, got %d", expectedVotes, len(votes))
	}

	// Update proposal and verify result integrity
	daoInstance.UpdateAllProposalStatuses()

	proposal, err := daoInstance.GetProposal(proposalHash)
	if err != nil {
		return fmt.Errorf("proposal retrieval failed: %w", err)
	}

	if proposal.Results == nil {
		return fmt.Errorf("proposal results should not be nil")
	}

	if proposal.Results.YesVotes != expectedWeight {
		return fmt.Errorf("vote weight integrity violated: expected %d, got %d", expectedWeight, proposal.Results.YesVotes)
	}

	r.results["DataIntegrity"].Metrics["token_operations"] = 4
	r.results["DataIntegrity"].Metrics["vote_operations"] = expectedVotes
	r.results["DataIntegrity"].Metrics["total_supply_consistent"] = true

	return nil
}

// validateScalabilityLimits validates system behavior under increased load
func (r *SystemValidationRunner) validateScalabilityLimits() error {
	daoInstance := dao.NewDAO("SCALE", "Scalability Token", 18)

	testDistribution := map[string]uint64{
		"scale_treasury": 100000000,
	}
	err := daoInstance.InitialTokenDistribution(testDistribution)
	if err != nil {
		return fmt.Errorf("scalability test setup failed: %w", err)
	}

	// Test with larger user base
	numUsers := 200
	users := make([]crypto.PrivateKey, numUsers)

	start := time.Now()
	for i := range users {
		users[i] = crypto.GeneratePrivateKey()
		err = daoInstance.MintTokens(users[i].PublicKey(), 5000)
		if err != nil {
			return fmt.Errorf("user %d setup failed: %w", i, err)
		}
	}
	setupDuration := time.Since(start)

	// Test proposal creation scalability
	start = time.Now()
	numProposals := 50

	for i := 0; i < numProposals; i++ {
		creator := users[i%numUsers]
		proposalTx := &dao.ProposalTx{
			Fee:          200,
			Title:        fmt.Sprintf("Scalability Test Proposal %d", i),
			Description:  "Testing scalability limits",
			ProposalType: dao.ProposalTypeGeneral,
			VotingType:   dao.VotingTypeSimple,
			StartTime:    time.Now().Unix() - 100,
			EndTime:      time.Now().Unix() + 3600,
			Threshold:    1000,
			MetadataHash: r.generateRandomHash(),
		}

		proposalHash := r.generateTxHash(proposalTx, creator)
		err := daoInstance.ProcessDAOTransaction(proposalTx, creator.PublicKey(), proposalHash)
		if err != nil {
			return fmt.Errorf("proposal %d creation failed: %w", i, err)
		}
	}
	proposalDuration := time.Since(start)

	// Test voting scalability
	proposals := daoInstance.ListAllProposals()
	if len(proposals) < numProposals {
		return fmt.Errorf("not all proposals were created: expected %d, got %d", numProposals, len(proposals))
	}

	start = time.Now()
	voteCount := 0

	// Have multiple users vote on multiple proposals
	for i := 0; i < 20; i++ {
		voter := users[i]
		proposal := proposals[i%len(proposals)]

		voteTx := &dao.VoteTx{
			Fee:        100,
			ProposalID: proposal.ID,
			Choice:     dao.VoteChoiceYes,
			Weight:     100,
			Reason:     fmt.Sprintf("Scalability vote %d", i),
		}

		voteHash := r.generateTxHash(voteTx, voter)
		err := daoInstance.ProcessDAOTransaction(voteTx, voter.PublicKey(), voteHash)
		if err != nil {
			return fmt.Errorf("vote %d failed: %w", i, err)
		}
		voteCount++
	}
	votingDuration := time.Since(start)

	// Performance thresholds for scalability
	maxSetupTime := 10 * time.Second
	maxProposalTime := 30 * time.Second
	maxVotingTime := 10 * time.Second

	if setupDuration > maxSetupTime {
		return fmt.Errorf("user setup took too long: %v > %v", setupDuration, maxSetupTime)
	}

	if proposalDuration > maxProposalTime {
		return fmt.Errorf("proposal creation took too long: %v > %v", proposalDuration, maxProposalTime)
	}

	if votingDuration > maxVotingTime {
		return fmt.Errorf("voting took too long: %v > %v", votingDuration, maxVotingTime)
	}

	r.results["ScalabilityLimits"].Metrics["users_created"] = numUsers
	r.results["ScalabilityLimits"].Metrics["proposals_created"] = numProposals
	r.results["ScalabilityLimits"].Metrics["votes_cast"] = voteCount
	r.results["ScalabilityLimits"].Metrics["setup_duration_ms"] = setupDuration.Milliseconds()
	r.results["ScalabilityLimits"].Metrics["proposal_duration_ms"] = proposalDuration.Milliseconds()
	r.results["ScalabilityLimits"].Metrics["voting_duration_ms"] = votingDuration.Milliseconds()

	return nil
}

// validateResourceUtilization validates system resource usage
func (r *SystemValidationRunner) validateResourceUtilization() error {
	// Get initial memory stats
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)
	runtime.GC()

	daoInstance := dao.NewDAO("RESOURCE", "Resource Token", 18)

	testDistribution := map[string]uint64{
		"resource_treasury": 50000000,
	}
	err := daoInstance.InitialTokenDistribution(testDistribution)
	if err != nil {
		return fmt.Errorf("resource test setup failed: %w", err)
	}

	// Perform memory-intensive operations
	numUsers := 100
	users := make([]crypto.PrivateKey, numUsers)
	for i := range users {
		users[i] = crypto.GeneratePrivateKey()
		err = daoInstance.MintTokens(users[i].PublicKey(), 10000)
		if err != nil {
			return fmt.Errorf("user %d setup failed: %w", i, err)
		}
		daoInstance.InitializeUserReputation(users[i].PublicKey(), uint64(1000+i))
	}

	// Create proposals and votes
	for i := 0; i < 20; i++ {
		creator := users[i%numUsers]
		proposalTx := &dao.ProposalTx{
			Fee:          200,
			Title:        fmt.Sprintf("Resource Test Proposal %d", i),
			Description:  "Testing resource utilization",
			ProposalType: dao.ProposalTypeGeneral,
			VotingType:   dao.VotingTypeSimple,
			StartTime:    time.Now().Unix() - 100,
			EndTime:      time.Now().Unix() + 3600,
			Threshold:    1000,
			MetadataHash: r.generateRandomHash(),
		}

		proposalHash := r.generateTxHash(proposalTx, creator)
		err := daoInstance.ProcessDAOTransaction(proposalTx, creator.PublicKey(), proposalHash)
		if err != nil {
			return fmt.Errorf("proposal %d creation failed: %w", i, err)
		}
	}

	// Get final memory stats
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Calculate memory usage
	memoryUsed := m2.Alloc - m1.Alloc
	memoryPerUser := memoryUsed / uint64(numUsers)

	// Resource utilization thresholds
	maxMemoryPerUser := uint64(50000)  // 50KB per user
	maxTotalMemory := uint64(10000000) // 10MB total

	if memoryPerUser > maxMemoryPerUser {
		return fmt.Errorf("memory per user too high: %d > %d bytes", memoryPerUser, maxMemoryPerUser)
	}

	if memoryUsed > maxTotalMemory {
		return fmt.Errorf("total memory usage too high: %d > %d bytes", memoryUsed, maxTotalMemory)
	}

	r.results["ResourceUtilization"].Metrics["memory_used_bytes"] = memoryUsed
	r.results["ResourceUtilization"].Metrics["memory_per_user_bytes"] = memoryPerUser
	r.results["ResourceUtilization"].Metrics["gc_cycles"] = m2.NumGC - m1.NumGC
	r.results["ResourceUtilization"].Metrics["heap_objects"] = m2.HeapObjects

	return nil
}

// validateSystemResilience validates system resilience and recovery
func (r *SystemValidationRunner) validateSystemResilience() error {
	daoInstance := dao.NewDAO("RESILIENCE", "Resilience Token", 18)

	testDistribution := map[string]uint64{
		"resilience_treasury": 10000000,
	}
	err := daoInstance.InitialTokenDistribution(testDistribution)
	if err != nil {
		return fmt.Errorf("resilience test setup failed: %w", err)
	}

	user := crypto.GeneratePrivateKey()
	err = daoInstance.MintTokens(user.PublicKey(), 20000)
	if err != nil {
		return fmt.Errorf("user setup failed: %w", err)
	}

	// Test system resilience under various error conditions
	errorScenarios := []struct {
		name string
		test func() error
	}{
		{"InvalidTransactionRecovery", func() error {
			// Try invalid transaction, then valid one
			invalidTx := &dao.ProposalTx{
				Fee:          -100, // Invalid negative fee
				Title:        "Invalid Proposal",
				Description:  "This should fail",
				ProposalType: dao.ProposalTypeGeneral,
				VotingType:   dao.VotingTypeSimple,
				StartTime:    time.Now().Unix() - 100,
				EndTime:      time.Now().Unix() + 3600,
				Threshold:    1000,
				MetadataHash: r.generateRandomHash(),
			}

			invalidHash := r.generateTxHash(invalidTx, user)
			err := daoInstance.ProcessDAOTransaction(invalidTx, user.PublicKey(), invalidHash)
			if err == nil {
				return fmt.Errorf("invalid transaction should have failed")
			}

			// Now try valid transaction
			validTx := &dao.ProposalTx{
				Fee:          200,
				Title:        "Valid Recovery Proposal",
				Description:  "This should succeed after error",
				ProposalType: dao.ProposalTypeGeneral,
				VotingType:   dao.VotingTypeSimple,
				StartTime:    time.Now().Unix() - 100,
				EndTime:      time.Now().Unix() + 3600,
				Threshold:    1000,
				MetadataHash: r.generateRandomHash(),
			}

			validHash := r.generateTxHash(validTx, user)
			return daoInstance.ProcessDAOTransaction(validTx, user.PublicKey(), validHash)
		}},
		{"StateConsistencyAfterErrors", func() error {
			// Verify system state is consistent after errors
			proposals := daoInstance.ListAllProposals()
			if len(proposals) == 0 {
				return fmt.Errorf("should have at least one proposal after recovery")
			}

			totalSupply := daoInstance.GetTotalSupply()
			if totalSupply == 0 {
				return fmt.Errorf("total supply should be greater than 0")
			}

			balance := daoInstance.GetTokenBalance(user.PublicKey())
			if balance == 0 {
				return fmt.Errorf("user balance should be greater than 0")
			}

			return nil
		}},
	}

	for _, scenario := range errorScenarios {
		err := scenario.test()
		if err != nil {
			return fmt.Errorf("resilience test '%s' failed: %w", scenario.name, err)
		}
	}

	r.results["SystemResilience"].Metrics["error_scenarios_tested"] = len(errorScenarios)
	r.results["SystemResilience"].Metrics["recovery_successful"] = true

	return nil
}

// generateValidationReport generates a comprehensive validation report
func (r *SystemValidationRunner) generateValidationReport() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	totalDuration := time.Since(r.startTime)
	passedTests := 0
	failedTests := 0

	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("PROJECTX DAO SYSTEM VALIDATION REPORT\n")
	fmt.Printf(strings.Repeat("=", 80) + "\n")
	fmt.Printf("Total Validation Duration: %v\n", totalDuration)
	fmt.Printf("Timestamp: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("\n")

	// Test Results Summary
	fmt.Printf("TEST RESULTS SUMMARY:\n")
	fmt.Printf(strings.Repeat("-", 40) + "\n")

	for testName, result := range r.results {
		status := "PASSED"
		if !result.Passed {
			status = "FAILED"
			failedTests++
		} else {
			passedTests++
		}

		fmt.Printf("%-30s: %s (%v)\n", testName, status, result.Duration)
		if !result.Passed && result.Error != nil {
			fmt.Printf("  Error: %s\n", result.Error.Error())
		}
	}

	fmt.Printf("\n")
	fmt.Printf("Total Tests: %d\n", passedTests+failedTests)
	fmt.Printf("Passed: %d\n", passedTests)
	fmt.Printf("Failed: %d\n", failedTests)
	fmt.Printf("Success Rate: %.1f%%\n", float64(passedTests)/float64(passedTests+failedTests)*100)

	// Detailed Metrics
	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("DETAILED METRICS:\n")
	fmt.Printf(strings.Repeat("=", 80) + "\n")

	for testName, result := range r.results {
		if len(result.Metrics) > 0 {
			fmt.Printf("\n%s:\n", testName)
			fmt.Printf(strings.Repeat("-", len(testName)) + "\n")
			for metric, value := range result.Metrics {
				fmt.Printf("  %-25s: %v\n", metric, value)
			}
		}
	}

	// System Information
	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("SYSTEM INFORMATION:\n")
	fmt.Printf(strings.Repeat("=", 80) + "\n")
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("CPUs: %d\n", runtime.NumCPU())

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Memory Allocated: %d KB\n", m.Alloc/1024)
	fmt.Printf("Total Allocations: %d\n", m.TotalAlloc/1024)
	fmt.Printf("GC Cycles: %d\n", m.NumGC)

	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")

	if failedTests > 0 {
		return fmt.Errorf("validation failed: %d out of %d tests failed", failedTests, passedTests+failedTests)
	}

	fmt.Printf("✅ ALL VALIDATION TESTS PASSED\n")
	fmt.Printf("System is ready for production deployment\n")
	fmt.Printf(strings.Repeat("=", 80) + "\n")

	return nil
}

// Helper methods

func (r *SystemValidationRunner) createTestGenesisBlock() *core.Block {
	privKey := crypto.GeneratePrivateKey()

	genesisTx := &core.Transaction{
		TxInner: core.CollectionTx{
			Fee:      0,
			MetaData: []byte("Validation Genesis Block"),
		},
		From:  privKey.PublicKey(),
		To:    privKey.PublicKey(),
		Value: 1000000000,
	}
	genesisTx.Sign(privKey)

	header := &core.Header{
		Version:       1,
		PrevBlockHash: [32]byte{},
		Height:        0,
		Timestamp:     time.Now().UnixNano(),
	}

	block, _ := core.NewBlock(header, []*core.Transaction{genesisTx})
	dataHash, _ := core.CalculateDataHash(block.Transactions)
	block.Header.DataHash = dataHash
	block.Sign(privKey)

	return block
}

func (r *SystemValidationRunner) createBlockWithTransaction(bc *core.Blockchain, tx *core.Transaction) *core.Block {
	privKey := crypto.GeneratePrivateKey()

	prevBlock, _ := bc.GetBlock(bc.Height())

	header := &core.Header{
		Version:       1,
		PrevBlockHash: prevBlock.Hash(core.BlockHasher{}),
		Height:        bc.Height() + 1,
		Timestamp:     time.Now().UnixNano(),
	}

	block, _ := core.NewBlock(header, []*core.Transaction{tx})
	dataHash, _ := core.CalculateDataHash(block.Transactions)
	block.Header.DataHash = dataHash
	block.Sign(privKey)

	return block
}

func (r *SystemValidationRunner) generateTxHash(tx interface{}, signer crypto.PrivateKey) [32]byte {
	data := fmt.Sprintf("%v%s%d", tx, signer.PublicKey().String(), time.Now().UnixNano())
	hash := [32]byte{}
	copy(hash[:], []byte(data)[:32])
	return hash
}

func (r *SystemValidationRunner) generateRandomHash() [32]byte {
	hash := [32]byte{}
	for i := range hash {
		hash[i] = byte(i % 256)
	}
	return hash
}

// RunSystemValidation is the main entry point for system validation
func RunSystemValidation() error {
	runner := NewSystemValidationRunner()
	defer runner.cancel()

	return runner.RunCompleteSystemValidation()
}
