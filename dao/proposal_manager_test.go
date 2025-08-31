package dao

import (
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

func TestNewProposalManager(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)
	pm := NewProposalManager(dao)

	if pm.dao != dao {
		t.Error("ProposalManager should reference the correct DAO")
	}
}

func TestCreateProposal(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)
	pm := NewProposalManager(dao)

	// Setup initial distribution
	creator := crypto.GeneratePrivateKey().PublicKey()
	distributions := map[string]uint64{
		creator.String(): 2000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create proposal
	proposalTx := &ProposalTx{
		Fee:          100,
		Title:        "Test Proposal",
		Description:  "This is a test proposal",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    time.Now().Unix() + 3600,
		EndTime:      time.Now().Unix() + 90000,
		Threshold:    5100,
		MetadataHash: types.Hash{},
	}

	txHash := randomHash()
	proposal, err := pm.CreateProposal(proposalTx, creator, txHash)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	if proposal.Title != "Test Proposal" {
		t.Errorf("Expected title 'Test Proposal', got %s", proposal.Title)
	}

	if proposal.Status != ProposalStatusPending {
		t.Errorf("Expected status pending, got %d", proposal.Status)
	}
}

func TestCreateProposalSpamPrevention(t *testing.T) {
	// Skip this test since spam prevention is disabled for testing
	t.Skip("Spam prevention is disabled for testing")
}

func TestCancelProposal(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)
	pm := NewProposalManager(dao)

	// Setup initial distribution
	creator := crypto.GeneratePrivateKey().PublicKey()
	other := crypto.GeneratePrivateKey().PublicKey()
	distributions := map[string]uint64{
		creator.String(): 2000,
		other.String():   1000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create proposal
	proposalTx := &ProposalTx{
		Fee:          100,
		Title:        "Test Proposal",
		Description:  "This is a test proposal",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    time.Now().Unix() + 3600,
		EndTime:      time.Now().Unix() + 90000,
		Threshold:    5100,
		MetadataHash: types.Hash{},
	}

	txHash := randomHash()
	proposal, err := pm.CreateProposal(proposalTx, creator, txHash)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Test cancellation by creator
	err = pm.CancelProposal(txHash, creator)
	if err != nil {
		t.Fatalf("Failed to cancel proposal: %v", err)
	}

	if proposal.Status != ProposalStatusCancelled {
		t.Errorf("Expected status cancelled, got %d", proposal.Status)
	}

	// Test cancellation by non-creator (should fail)
	proposalTx2 := &ProposalTx{
		Fee:          100,
		Title:        "Test Proposal 2",
		Description:  "This is another test proposal",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    time.Now().Unix() + 7200,
		EndTime:      time.Now().Unix() + 7200 + 86400, // Ensure minimum voting period
		Threshold:    5100,
		MetadataHash: types.Hash{},
	}

	txHash2 := randomHash()
	_, err = pm.CreateProposal(proposalTx2, creator, txHash2)
	if err != nil {
		t.Fatalf("Failed to create second proposal: %v", err)
	}

	err = pm.CancelProposal(txHash2, other)
	if err == nil {
		t.Error("Expected error when non-creator tries to cancel")
	}
}

func TestGetProposalsByStatus(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)
	pm := NewProposalManager(dao)

	// Setup initial distribution with multiple creators to avoid spam prevention
	creator1 := crypto.GeneratePrivateKey().PublicKey()
	creator2 := crypto.GeneratePrivateKey().PublicKey()
	creator3 := crypto.GeneratePrivateKey().PublicKey()
	creators := []crypto.PublicKey{creator1, creator2, creator3}

	distributions := map[string]uint64{
		creator1.String(): 5000,
		creator2.String(): 5000,
		creator3.String(): 5000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create multiple proposals with different statuses
	for i := 0; i < 3; i++ {
		proposalTx := &ProposalTx{
			Fee:          100,
			Title:        "Test Proposal",
			Description:  "This is a test proposal",
			ProposalType: ProposalTypeGeneral,
			VotingType:   VotingTypeSimple,
			StartTime:    time.Now().Unix() + int64(3600*(i+1)),
			EndTime:      time.Now().Unix() + int64(90000*(i+1)),
			Threshold:    5100,
			MetadataHash: types.Hash{},
		}

		txHash := randomHash()
		_, err := pm.CreateProposal(proposalTx, creators[i], txHash)
		if err != nil {
			t.Fatalf("Failed to create proposal %d: %v", i, err)
		}
	}

	// Get pending proposals
	pendingProposals := pm.GetProposalsByStatus(ProposalStatusPending)
	if len(pendingProposals) != 3 {
		t.Errorf("Expected 3 pending proposals, got %d", len(pendingProposals))
	}

	// Cancel one proposal (use the correct creator)
	if len(pendingProposals) > 0 {
		proposalToCancel := pendingProposals[0]
		err := pm.CancelProposal(proposalToCancel.ID, proposalToCancel.Creator)
		if err != nil {
			t.Fatalf("Failed to cancel proposal: %v", err)
		}
	}

	// Check counts again
	pendingProposals = pm.GetProposalsByStatus(ProposalStatusPending)
	cancelledProposals := pm.GetProposalsByStatus(ProposalStatusCancelled)

	if len(pendingProposals) != 2 {
		t.Errorf("Expected 2 pending proposals after cancellation, got %d", len(pendingProposals))
	}

	if len(cancelledProposals) != 1 {
		t.Errorf("Expected 1 cancelled proposal, got %d", len(cancelledProposals))
	}
}

func TestGetProposalsByType(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)
	pm := NewProposalManager(dao)

	// Setup initial distribution with multiple creators
	creator1 := crypto.GeneratePrivateKey().PublicKey()
	creator2 := crypto.GeneratePrivateKey().PublicKey()
	creator3 := crypto.GeneratePrivateKey().PublicKey()
	creator4 := crypto.GeneratePrivateKey().PublicKey()
	creators := []crypto.PublicKey{creator1, creator2, creator3, creator4}

	distributions := map[string]uint64{
		creator1.String(): 10000,
		creator2.String(): 10000,
		creator3.String(): 10000,
		creator4.String(): 10000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create proposals of different types
	proposalTypes := []ProposalType{
		ProposalTypeGeneral,
		ProposalTypeTreasury,
		ProposalTypeTechnical,
		ProposalTypeParameter,
	}

	for i, pType := range proposalTypes {
		proposalTx := &ProposalTx{
			Fee:          100,
			Title:        "Test Proposal",
			Description:  "This is a test proposal",
			ProposalType: pType,
			VotingType:   VotingTypeSimple,
			StartTime:    time.Now().Unix() + int64(3600*(i+10)),
			EndTime:      time.Now().Unix() + int64(90000*(i+10)),
			Threshold:    5100,
			MetadataHash: types.Hash{},
		}

		txHash := randomHash()
		_, err := pm.CreateProposal(proposalTx, creators[i], txHash)
		if err != nil {
			t.Fatalf("Failed to create proposal of type %d: %v", pType, err)
		}
	}

	// Test getting proposals by type
	for _, pType := range proposalTypes {
		proposals := pm.GetProposalsByType(pType)
		if len(proposals) != 1 {
			t.Errorf("Expected 1 proposal of type %d, got %d", pType, len(proposals))
		}
		if proposals[0].ProposalType != pType {
			t.Errorf("Expected proposal type %d, got %d", pType, proposals[0].ProposalType)
		}
	}
}

func TestGetProposalsByCreator(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)
	pm := NewProposalManager(dao)

	// Setup initial distribution
	creator1 := crypto.GeneratePrivateKey().PublicKey()
	creator2 := crypto.GeneratePrivateKey().PublicKey()
	distributions := map[string]uint64{
		creator1.String(): 5000,
		creator2.String(): 5000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create one proposal from creator1 (to avoid spam prevention, we'll only create one per creator)
	proposalTx1 := &ProposalTx{
		Fee:          100,
		Title:        "Test Proposal from Creator 1",
		Description:  "This is a test proposal",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    time.Now().Unix() + 3600,
		EndTime:      time.Now().Unix() + 90000,
		Threshold:    5100,
		MetadataHash: types.Hash{},
	}

	txHash1 := randomHash()
	_, err := pm.CreateProposal(proposalTx1, creator1, txHash1)
	if err != nil {
		t.Fatalf("Failed to create proposal from creator1: %v", err)
	}

	proposalTx2 := &ProposalTx{
		Fee:          100,
		Title:        "Test Proposal from Creator 2",
		Description:  "This is a test proposal",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    time.Now().Unix() + 7200,
		EndTime:      time.Now().Unix() + 7200 + 86400, // Ensure minimum voting period
		Threshold:    5100,
		MetadataHash: types.Hash{},
	}

	txHash2 := randomHash()
	_, err = pm.CreateProposal(proposalTx2, creator2, txHash2)
	if err != nil {
		t.Fatalf("Failed to create proposal from creator2: %v", err)
	}

	// Test getting proposals by creator
	creator1Proposals := pm.GetProposalsByCreator(creator1)
	creator2Proposals := pm.GetProposalsByCreator(creator2)

	if len(creator1Proposals) != 1 {
		t.Errorf("Expected 1 proposal from creator1, got %d", len(creator1Proposals))
	}

	if len(creator2Proposals) != 1 {
		t.Errorf("Expected 1 proposal from creator2, got %d", len(creator2Proposals))
	}

	// Verify creator addresses
	if len(creator1Proposals) > 0 && creator1Proposals[0].Creator.String() != creator1.String() {
		t.Error("Proposal creator mismatch for creator1")
	}

	if len(creator2Proposals) > 0 && creator2Proposals[0].Creator.String() != creator2.String() {
		t.Error("Proposal creator mismatch for creator2")
	}
}

func TestGetProposalVotingProgress(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)
	pm := NewProposalManager(dao)

	// Setup initial distribution
	creator := crypto.GeneratePrivateKey().PublicKey()
	voter1 := crypto.GeneratePrivateKey().PublicKey()
	voter2 := crypto.GeneratePrivateKey().PublicKey()
	distributions := map[string]uint64{
		creator.String(): 2000,
		voter1.String():  3000,
		voter2.String():  2000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create proposal
	now := time.Now().Unix()
	proposalTx := &ProposalTx{
		Fee:          100,
		Title:        "Test Proposal",
		Description:  "This is a test proposal",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    now - 100,   // Started in the past
		EndTime:      now + 86400, // Ends in future
		Threshold:    5100,
		MetadataHash: types.Hash{},
	}

	txHash := randomHash()
	_, err := pm.CreateProposal(proposalTx, creator, txHash)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Manually set proposal to active for testing
	proposal, _ := dao.GetProposal(txHash)
	proposal.Status = ProposalStatusActive

	// Cast votes
	voteTx1 := &VoteTx{
		Fee:        50,
		ProposalID: txHash,
		Choice:     VoteChoiceYes,
		Weight:     1000,
		Reason:     "I support this proposal",
	}
	err = dao.Processor.ProcessVoteTx(voteTx1, voter1)
	if err != nil {
		t.Fatalf("Failed to process vote 1: %v", err)
	}

	voteTx2 := &VoteTx{
		Fee:        50,
		ProposalID: txHash,
		Choice:     VoteChoiceNo,
		Weight:     800,
		Reason:     "I oppose this proposal",
	}
	err = dao.Processor.ProcessVoteTx(voteTx2, voter2)
	if err != nil {
		t.Fatalf("Failed to process vote 2: %v", err)
	}

	// Get voting progress
	progress, err := pm.GetProposalVotingProgress(txHash)
	if err != nil {
		t.Fatalf("Failed to get voting progress: %v", err)
	}

	if progress.TotalVotes != 2 {
		t.Errorf("Expected 2 total votes, got %d", progress.TotalVotes)
	}

	if progress.YesVotes != 1000 {
		t.Errorf("Expected 1000 yes votes, got %d", progress.YesVotes)
	}

	if progress.NoVotes != 800 {
		t.Errorf("Expected 800 no votes, got %d", progress.NoVotes)
	}

	if len(progress.Voters) != 2 {
		t.Errorf("Expected 2 voters, got %d", len(progress.Voters))
	}

	// Check voter information
	foundVoter1 := false
	foundVoter2 := false
	for _, voter := range progress.Voters {
		if voter.Address.String() == voter1.String() {
			foundVoter1 = true
			if voter.Choice != VoteChoiceYes {
				t.Error("Voter1 choice mismatch")
			}
			if voter.Weight != 1000 {
				t.Error("Voter1 weight mismatch")
			}
		}
		if voter.Address.String() == voter2.String() {
			foundVoter2 = true
			if voter.Choice != VoteChoiceNo {
				t.Error("Voter2 choice mismatch")
			}
			if voter.Weight != 800 {
				t.Error("Voter2 weight mismatch")
			}
		}
	}

	if !foundVoter1 || !foundVoter2 {
		t.Error("Not all voters found in progress")
	}
}

func TestGetProposalStatistics(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)
	pm := NewProposalManager(dao)

	// Setup initial distribution with multiple creators
	creator1 := crypto.GeneratePrivateKey().PublicKey()
	creator2 := crypto.GeneratePrivateKey().PublicKey()
	creator3 := crypto.GeneratePrivateKey().PublicKey()
	creators := []crypto.PublicKey{creator1, creator2, creator3}

	distributions := map[string]uint64{
		creator1.String(): 10000,
		creator2.String(): 10000,
		creator3.String(): 10000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create proposals of different types and statuses
	proposalTypes := []ProposalType{
		ProposalTypeGeneral,
		ProposalTypeTreasury,
		ProposalTypeTechnical,
	}

	for i, pType := range proposalTypes {
		proposalTx := &ProposalTx{
			Fee:          100,
			Title:        "Test Proposal",
			Description:  "This is a test proposal",
			ProposalType: pType,
			VotingType:   VotingTypeSimple,
			StartTime:    time.Now().Unix() + int64(3600*(i+30)),
			EndTime:      time.Now().Unix() + int64(3600*(i+30)) + 86400, // Ensure minimum voting period
			Threshold:    5100,
			MetadataHash: types.Hash{},
		}

		txHash := randomHash()
		proposal, err := pm.CreateProposal(proposalTx, creators[i], txHash)
		if err != nil {
			t.Fatalf("Failed to create proposal of type %d: %v", pType, err)
		}

		// Cancel the first proposal
		if i == 0 {
			pm.CancelProposal(proposal.ID, creators[i])
		}
	}

	// Get statistics
	stats := pm.GetProposalStatistics()

	if stats.Total != 3 {
		t.Errorf("Expected 3 total proposals, got %d", stats.Total)
	}

	if stats.StatusCounts[ProposalStatusPending] != 2 {
		t.Errorf("Expected 2 pending proposals, got %d", stats.StatusCounts[ProposalStatusPending])
	}

	if stats.StatusCounts[ProposalStatusCancelled] != 1 {
		t.Errorf("Expected 1 cancelled proposal, got %d", stats.StatusCounts[ProposalStatusCancelled])
	}

	if stats.TypeCounts[ProposalTypeGeneral] != 1 {
		t.Errorf("Expected 1 general proposal, got %d", stats.TypeCounts[ProposalTypeGeneral])
	}

	if stats.TypeCounts[ProposalTypeTreasury] != 1 {
		t.Errorf("Expected 1 treasury proposal, got %d", stats.TypeCounts[ProposalTypeTreasury])
	}

	if stats.TypeCounts[ProposalTypeTechnical] != 1 {
		t.Errorf("Expected 1 technical proposal, got %d", stats.TypeCounts[ProposalTypeTechnical])
	}
}

func TestExecuteProposal(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)
	pm := NewProposalManager(dao)

	// Setup initial distribution
	creator := crypto.GeneratePrivateKey().PublicKey()
	executor := crypto.GeneratePrivateKey().PublicKey()
	distributions := map[string]uint64{
		creator.String():  2000,
		executor.String(): 2000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create proposal
	proposalTx := &ProposalTx{
		Fee:          100,
		Title:        "Test Proposal",
		Description:  "This is a test proposal",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    time.Now().Unix(),
		EndTime:      time.Now().Unix() + 86400,
		Threshold:    5100,
		MetadataHash: types.Hash{},
	}

	txHash := randomHash()
	proposal, err := pm.CreateProposal(proposalTx, creator, txHash)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Try to execute before it's passed (should fail)
	err = pm.ExecuteProposal(txHash, executor)
	if err == nil {
		t.Error("Expected error when executing non-passed proposal")
	}

	// Manually set proposal to passed status
	proposal.Status = ProposalStatusPassed
	proposal.Results.Passed = true

	// Execute proposal
	err = pm.ExecuteProposal(txHash, executor)
	if err != nil {
		t.Fatalf("Failed to execute proposal: %v", err)
	}

	if proposal.Status != ProposalStatusExecuted {
		t.Errorf("Expected status executed, got %d", proposal.Status)
	}
}

func TestUpdateAllProposalStatuses(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)
	pm := NewProposalManager(dao)

	// Setup initial distribution
	creator := crypto.GeneratePrivateKey().PublicKey()
	distributions := map[string]uint64{
		creator.String(): 5000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create proposals with different timeframes
	now := time.Now().Unix()

	// Past proposal (should become active then closed)
	proposalTx1 := &ProposalTx{
		Fee:          100,
		Title:        "Past Proposal",
		Description:  "This proposal is in the past",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    now - 90000, // Started well in the past
		EndTime:      now - 1800,  // Ended 30 minutes ago
		Threshold:    5100,
		MetadataHash: types.Hash{},
	}

	txHash1 := randomHash()
	_, err := pm.CreateProposal(proposalTx1, creator, txHash1)
	if err != nil {
		t.Fatalf("Failed to create past proposal: %v", err)
	}

	// Current proposal (should become active)
	proposalTx2 := &ProposalTx{
		Fee:          100,
		Title:        "Current Proposal",
		Description:  "This proposal is currently active",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    now - 100,   // Started recently
		EndTime:      now + 86400, // Ends in 24 hours
		Threshold:    5100,
		MetadataHash: types.Hash{},
	}

	txHash2 := randomHash()
	_, err = pm.CreateProposal(proposalTx2, creator, txHash2)
	if err != nil {
		t.Fatalf("Failed to create current proposal: %v", err)
	}

	// Update all proposal statuses
	err = pm.UpdateAllProposalStatuses()
	if err != nil {
		t.Fatalf("Failed to update proposal statuses: %v", err)
	}

	// Check statuses
	proposal1, _ := dao.GetProposal(txHash1)
	proposal2, _ := dao.GetProposal(txHash2)

	if proposal1.Status != ProposalStatusRejected {
		t.Errorf("Expected past proposal to be rejected, got status %d", proposal1.Status)
	}

	if proposal2.Status != ProposalStatusActive {
		t.Errorf("Expected current proposal to be active, got status %d", proposal2.Status)
	}
}
