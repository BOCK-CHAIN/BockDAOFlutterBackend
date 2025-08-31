package dao

import (
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

func TestSecurityIntegration_BasicFlow(t *testing.T) {
	dao := NewDAO("TEST", "Test Token", 18)

	// Create test users
	founder := crypto.GeneratePrivateKey().PublicKey()
	member := crypto.GeneratePrivateKey().PublicKey()

	// Initialize founder roles
	err := dao.InitializeFounderRoles([]crypto.PublicKey{founder})
	if err != nil {
		t.Fatal("should initialize founder roles")
	}

	// Grant member role
	err = dao.GrantRole(member, RoleMember, founder, 0)
	if err != nil {
		t.Fatal("should grant member role")
	}

	// Distribute tokens
	distributions := map[string]uint64{
		member.String(): 10000, // Increased to cover multiple transactions
	}
	err = dao.InitialTokenDistribution(distributions)
	if err != nil {
		t.Fatal("should distribute tokens")
	}

	// Test authorized proposal creation
	startTime := time.Now().Unix()
	proposalTx := &ProposalTx{
		Fee:          100,
		Title:        "Test Proposal",
		Description:  "A test proposal",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    startTime,
		EndTime:      startTime + 86400, // 24 hours
		Threshold:    500,
	}

	txHash := types.Hash{1, 2, 3}
	err = dao.SecureProcessDAOTransaction(proposalTx, member, txHash)
	if err != nil {
		t.Fatalf("member should be able to create proposal: %v", err)
	}

	// Test unauthorized proposal creation
	unauthorized := crypto.GeneratePrivateKey().PublicKey()
	err = dao.SecureProcessDAOTransaction(proposalTx, unauthorized, txHash)
	if err == nil {
		t.Fatal("unauthorized user should not create proposal")
	}

	// Verify audit log
	auditEntries, err := dao.GetAuditLog(founder, 10, 0, SecurityLevelPublic)
	if err != nil {
		t.Fatal("founder should access audit log")
	}
	if len(auditEntries) == 0 {
		t.Fatal("should have audit entries")
	}
}

func TestSecurityIntegration_EmergencyMode(t *testing.T) {
	dao := NewDAO("TEST", "Test Token", 18)

	// Create test users
	founder := crypto.GeneratePrivateKey().PublicKey()
	member := crypto.GeneratePrivateKey().PublicKey()

	// Initialize founder roles
	err := dao.InitializeFounderRoles([]crypto.PublicKey{founder})
	if err != nil {
		t.Fatal("should initialize founder roles")
	}

	// Grant member role
	err = dao.GrantRole(member, RoleMember, founder, 0)
	if err != nil {
		t.Fatal("should grant member role")
	}

	// Distribute tokens
	distributions := map[string]uint64{
		member.String(): 10000, // Increased to cover multiple transactions
	}
	err = dao.InitialTokenDistribution(distributions)
	if err != nil {
		t.Fatal("should distribute tokens")
	}

	// Test normal operation
	startTime := time.Now().Unix()
	proposalTx := &ProposalTx{
		Fee:          100,
		Title:        "Test Proposal",
		Description:  "A test proposal",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    startTime,
		EndTime:      startTime + 86400, // 24 hours
		Threshold:    500,
	}

	txHash := types.Hash{1, 2, 3}
	err = dao.SecureProcessDAOTransaction(proposalTx, member, txHash)
	if err != nil {
		t.Fatalf("should create proposal normally: %v", err)
	}

	// Activate emergency mode
	err = dao.ActivateEmergency(founder, "Security incident", SecurityLevelCritical, []string{"CreateProposal"})
	if err != nil {
		t.Fatal("founder should activate emergency")
	}

	// Test blocked operation during emergency
	txHash2 := types.Hash{4, 5, 6}
	err = dao.SecureProcessDAOTransaction(proposalTx, member, txHash2)
	if err == nil {
		t.Fatal("proposal creation should be blocked during emergency")
	}

	// Deactivate emergency
	err = dao.DeactivateEmergency(founder)
	if err != nil {
		t.Fatal("founder should deactivate emergency")
	}

	// Test normal operation restored
	txHash3 := types.Hash{10, 11, 12}
	err = dao.SecureProcessDAOTransaction(proposalTx, member, txHash3)
	if err != nil {
		t.Fatalf("proposal creation should work after emergency: %v", err)
	}
}
