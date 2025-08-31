package dao

import (
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
)

func TestDelegationCreation(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	delegator := crypto.GeneratePrivateKey().PublicKey()
	delegate := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		delegator.String(): 2000,
		delegate.String():  1000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create delegation transaction
	delegationTx := &DelegationTx{
		Fee:      100,
		Delegate: delegate,
		Duration: 86400, // 1 day
		Revoke:   false,
	}

	err := dao.Processor.ProcessDelegationTx(delegationTx, delegator)
	if err != nil {
		t.Fatalf("Failed to process delegation: %v", err)
	}

	// Verify delegation was created
	delegation, exists := dao.GetDelegation(delegator)
	if !exists {
		t.Fatal("Delegation not found")
	}

	if delegation.Delegate.String() != delegate.String() {
		t.Errorf("Expected delegate %s, got %s", delegate.String(), delegation.Delegate.String())
	}

	if !delegation.Active {
		t.Error("Expected delegation to be active")
	}

	// Verify fee was deducted
	if dao.TokenState.Balances[delegator.String()] != 1900 {
		t.Errorf("Expected delegator balance 1900, got %d", dao.TokenState.Balances[delegator.String()])
	}
}

func TestDelegationRevocation(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	delegator := crypto.GeneratePrivateKey().PublicKey()
	delegate := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		delegator.String(): 2000,
		delegate.String():  1000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create delegation first
	delegationTx := &DelegationTx{
		Fee:      100,
		Delegate: delegate,
		Duration: 86400,
		Revoke:   false,
	}

	err := dao.Processor.ProcessDelegationTx(delegationTx, delegator)
	if err != nil {
		t.Fatalf("Failed to create delegation: %v", err)
	}

	// Verify delegation is active
	delegation, exists := dao.GetDelegation(delegator)
	if !exists || !delegation.Active {
		t.Fatal("Delegation should be active")
	}

	// Revoke delegation
	revokeTx := &DelegationTx{
		Fee:      50,
		Delegate: delegate, // This field is ignored for revocation
		Duration: 0,        // This field is ignored for revocation
		Revoke:   true,
	}

	err = dao.Processor.ProcessDelegationTx(revokeTx, delegator)
	if err != nil {
		t.Fatalf("Failed to revoke delegation: %v", err)
	}

	// Verify delegation is revoked
	delegation, exists = dao.GetDelegation(delegator)
	if !exists {
		t.Fatal("Delegation record should still exist")
	}

	if delegation.Active {
		t.Error("Expected delegation to be revoked")
	}

	// Verify revocation fee was deducted
	expectedBalance := 2000 - 100 - 50 // initial - creation fee - revocation fee
	if dao.TokenState.Balances[delegator.String()] != uint64(expectedBalance) {
		t.Errorf("Expected delegator balance %d, got %d", expectedBalance, dao.TokenState.Balances[delegator.String()])
	}
}

func TestDelegationVotingPower(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	delegator := crypto.GeneratePrivateKey().PublicKey()
	delegate := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		delegator.String(): 2000,
		delegate.String():  1000,
	}
	dao.InitialTokenDistribution(distributions)

	// Check initial voting power
	delegatorPower := dao.GetEffectiveVotingPower(delegator)
	delegatePower := dao.GetEffectiveVotingPower(delegate)

	if delegatorPower != 2000 {
		t.Errorf("Expected delegator power 2000, got %d", delegatorPower)
	}

	if delegatePower != 1000 {
		t.Errorf("Expected delegate power 1000, got %d", delegatePower)
	}

	// Create delegation
	delegationTx := &DelegationTx{
		Fee:      100,
		Delegate: delegate,
		Duration: 86400,
		Revoke:   false,
	}

	err := dao.Processor.ProcessDelegationTx(delegationTx, delegator)
	if err != nil {
		t.Fatalf("Failed to create delegation: %v", err)
	}

	// Check voting power after delegation
	delegatorPowerAfter := dao.GetEffectiveVotingPower(delegator)
	delegatePowerAfter := dao.GetEffectiveVotingPower(delegate)

	if delegatorPowerAfter != 0 {
		t.Errorf("Expected delegator power 0 after delegation, got %d", delegatorPowerAfter)
	}

	// Delegate should have their own power + delegated power
	expectedDelegatePower := 1000 + (2000 - 100) // own + delegated (minus fee)
	if delegatePowerAfter != uint64(expectedDelegatePower) {
		t.Errorf("Expected delegate power %d, got %d", expectedDelegatePower, delegatePowerAfter)
	}

	// Test individual power calculations
	ownPower := dao.GetOwnVotingPower(delegator)
	if ownPower != 0 {
		t.Errorf("Expected delegator own power 0, got %d", ownPower)
	}

	delegatedPower := dao.GetDelegatedPower(delegate)
	if delegatedPower != 1900 { // 2000 - 100 fee
		t.Errorf("Expected delegated power to delegate 1900, got %d", delegatedPower)
	}
}

func TestDelegationValidation(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	delegator := crypto.GeneratePrivateKey().PublicKey()
	delegate := crypto.GeneratePrivateKey().PublicKey()
	nonExistentDelegate := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		delegator.String(): 2000,
		delegate.String():  1000,
	}
	dao.InitialTokenDistribution(distributions)

	// Test self-delegation (should fail)
	selfDelegationTx := &DelegationTx{
		Fee:      100,
		Delegate: delegator, // Same as delegator
		Duration: 86400,
		Revoke:   false,
	}

	err := dao.Processor.ProcessDelegationTx(selfDelegationTx, delegator)
	if err == nil {
		t.Error("Expected error for self-delegation")
	}

	// Test delegation to non-existent address (should fail)
	nonExistentTx := &DelegationTx{
		Fee:      100,
		Delegate: nonExistentDelegate,
		Duration: 86400,
		Revoke:   false,
	}

	err = dao.Processor.ProcessDelegationTx(nonExistentTx, delegator)
	if err == nil {
		t.Error("Expected error for delegation to non-existent address")
	}

	// Test zero duration (should fail)
	zeroDurationTx := &DelegationTx{
		Fee:      100,
		Delegate: delegate,
		Duration: 0,
		Revoke:   false,
	}

	err = dao.Processor.ProcessDelegationTx(zeroDurationTx, delegator)
	if err == nil {
		t.Error("Expected error for zero duration")
	}

	// Test negative duration (should fail)
	negativeDurationTx := &DelegationTx{
		Fee:      100,
		Delegate: delegate,
		Duration: -86400,
		Revoke:   false,
	}

	err = dao.Processor.ProcessDelegationTx(negativeDurationTx, delegator)
	if err == nil {
		t.Error("Expected error for negative duration")
	}

	// Test insufficient tokens for fee (should fail)
	poorDelegator := crypto.GeneratePrivateKey().PublicKey()
	distributions2 := map[string]uint64{
		poorDelegator.String(): 50, // Less than fee
	}
	dao.InitialTokenDistribution(distributions2)

	insufficientTx := &DelegationTx{
		Fee:      100,
		Delegate: delegate,
		Duration: 86400,
		Revoke:   false,
	}

	err = dao.Processor.ProcessDelegationTx(insufficientTx, poorDelegator)
	if err == nil {
		t.Error("Expected error for insufficient tokens")
	}
}

func TestDelegationExpiration(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	delegator := crypto.GeneratePrivateKey().PublicKey()
	delegate := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		delegator.String(): 2000,
		delegate.String():  1000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create short-duration delegation (1 second)
	delegationTx := &DelegationTx{
		Fee:      100,
		Delegate: delegate,
		Duration: 1, // 1 second
		Revoke:   false,
	}

	err := dao.Processor.ProcessDelegationTx(delegationTx, delegator)
	if err != nil {
		t.Fatalf("Failed to create delegation: %v", err)
	}

	// Check voting power immediately after delegation
	delegatePower := dao.GetEffectiveVotingPower(delegate)
	expectedPower := 1000 + 1900 // own + delegated (minus fee)
	if delegatePower != uint64(expectedPower) {
		t.Errorf("Expected delegate power %d, got %d", expectedPower, delegatePower)
	}

	// Wait for delegation to expire
	time.Sleep(2 * time.Second)

	// Check voting power after expiration
	delegatePowerAfter := dao.GetEffectiveVotingPower(delegate)
	if delegatePowerAfter != 1000 { // Only own power
		t.Errorf("Expected delegate power 1000 after expiration, got %d", delegatePowerAfter)
	}

	delegatorPowerAfter := dao.GetEffectiveVotingPower(delegator)
	if delegatorPowerAfter != 1900 { // Own power restored (minus fee)
		t.Errorf("Expected delegator power 1900 after expiration, got %d", delegatorPowerAfter)
	}
}

func TestMultipleDelegations(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	delegator1 := crypto.GeneratePrivateKey().PublicKey()
	delegator2 := crypto.GeneratePrivateKey().PublicKey()
	delegate := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		delegator1.String(): 2000,
		delegator2.String(): 1500,
		delegate.String():   1000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create first delegation
	delegation1Tx := &DelegationTx{
		Fee:      100,
		Delegate: delegate,
		Duration: 86400,
		Revoke:   false,
	}

	err := dao.Processor.ProcessDelegationTx(delegation1Tx, delegator1)
	if err != nil {
		t.Fatalf("Failed to create first delegation: %v", err)
	}

	// Create second delegation
	delegation2Tx := &DelegationTx{
		Fee:      50,
		Delegate: delegate,
		Duration: 86400,
		Revoke:   false,
	}

	err = dao.Processor.ProcessDelegationTx(delegation2Tx, delegator2)
	if err != nil {
		t.Fatalf("Failed to create second delegation: %v", err)
	}

	// Check delegate's total voting power
	delegatePower := dao.GetEffectiveVotingPower(delegate)
	expectedPower := 1000 + (2000 - 100) + (1500 - 50) // own + delegated1 + delegated2
	if delegatePower != uint64(expectedPower) {
		t.Errorf("Expected delegate power %d, got %d", expectedPower, delegatePower)
	}

	// Check delegated power specifically
	delegatedPower := dao.GetDelegatedPower(delegate)
	expectedDelegated := (2000 - 100) + (1500 - 50) // delegated1 + delegated2
	if delegatedPower != uint64(expectedDelegated) {
		t.Errorf("Expected delegated power %d, got %d", expectedDelegated, delegatedPower)
	}

	// Check delegators have no voting power
	if dao.GetEffectiveVotingPower(delegator1) != 0 {
		t.Error("Delegator1 should have no voting power")
	}

	if dao.GetEffectiveVotingPower(delegator2) != 0 {
		t.Error("Delegator2 should have no voting power")
	}
}

func TestDelegationRevocationValidation(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	delegator := crypto.GeneratePrivateKey().PublicKey()
	delegate := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		delegator.String(): 2000,
		delegate.String():  1000,
	}
	dao.InitialTokenDistribution(distributions)

	// Try to revoke non-existent delegation (should fail)
	revokeTx := &DelegationTx{
		Fee:      50,
		Delegate: delegate,
		Duration: 0,
		Revoke:   true,
	}

	err := dao.Processor.ProcessDelegationTx(revokeTx, delegator)
	if err == nil {
		t.Error("Expected error for revoking non-existent delegation")
	}

	// Create delegation first
	delegationTx := &DelegationTx{
		Fee:      100,
		Delegate: delegate,
		Duration: 86400,
		Revoke:   false,
	}

	err = dao.Processor.ProcessDelegationTx(delegationTx, delegator)
	if err != nil {
		t.Fatalf("Failed to create delegation: %v", err)
	}

	// Now revoke it (should succeed)
	err = dao.Processor.ProcessDelegationTx(revokeTx, delegator)
	if err != nil {
		t.Fatalf("Failed to revoke delegation: %v", err)
	}

	// Try to revoke again (should fail)
	err = dao.Processor.ProcessDelegationTx(revokeTx, delegator)
	if err == nil {
		t.Error("Expected error for revoking already revoked delegation")
	}
}

func TestDuplicateDelegationValidation(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	delegator := crypto.GeneratePrivateKey().PublicKey()
	delegate1 := crypto.GeneratePrivateKey().PublicKey()
	delegate2 := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		delegator.String(): 3000,
		delegate1.String(): 1000,
		delegate2.String(): 1000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create first delegation
	delegation1Tx := &DelegationTx{
		Fee:      100,
		Delegate: delegate1,
		Duration: 86400,
		Revoke:   false,
	}

	err := dao.Processor.ProcessDelegationTx(delegation1Tx, delegator)
	if err != nil {
		t.Fatalf("Failed to create first delegation: %v", err)
	}

	// Try to create second delegation while first is active (should fail)
	delegation2Tx := &DelegationTx{
		Fee:      100,
		Delegate: delegate2,
		Duration: 86400,
		Revoke:   false,
	}

	err = dao.Processor.ProcessDelegationTx(delegation2Tx, delegator)
	if err == nil {
		t.Error("Expected error for creating duplicate delegation")
	}
}

func TestDelegationListingFunctions(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	delegator1 := crypto.GeneratePrivateKey().PublicKey()
	delegator2 := crypto.GeneratePrivateKey().PublicKey()
	delegate := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		delegator1.String(): 2000,
		delegator2.String(): 1500,
		delegate.String():   1000,
	}
	dao.InitialTokenDistribution(distributions)

	// Create delegations
	delegation1Tx := &DelegationTx{
		Fee:      100,
		Delegate: delegate,
		Duration: 86400,
		Revoke:   false,
	}

	delegation2Tx := &DelegationTx{
		Fee:      50,
		Delegate: delegate,
		Duration: 86400,
		Revoke:   false,
	}

	dao.Processor.ProcessDelegationTx(delegation1Tx, delegator1)
	dao.Processor.ProcessDelegationTx(delegation2Tx, delegator2)

	// Test ListDelegations
	activeDelegations := dao.ListDelegations()
	if len(activeDelegations) != 2 {
		t.Errorf("Expected 2 active delegations, got %d", len(activeDelegations))
	}

	// Test GetDelegationsByDelegate
	delegationsForDelegate := dao.GetDelegationsByDelegate(delegate)
	if len(delegationsForDelegate) != 2 {
		t.Errorf("Expected 2 delegations for delegate, got %d", len(delegationsForDelegate))
	}

	// Verify the delegations are correct
	found1, found2 := false, false
	for _, delegation := range delegationsForDelegate {
		if delegation.Delegator.String() == delegator1.String() {
			found1 = true
		}
		if delegation.Delegator.String() == delegator2.String() {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("Not all expected delegations found")
	}
}

func TestMaximumDelegationDuration(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	delegator := crypto.GeneratePrivateKey().PublicKey()
	delegate := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		delegator.String(): 2000,
		delegate.String():  1000,
	}
	dao.InitialTokenDistribution(distributions)

	// Test maximum duration (1 year + 1 second, should fail)
	maxDurationTx := &DelegationTx{
		Fee:      100,
		Delegate: delegate,
		Duration: 365*24*3600 + 1, // 1 year + 1 second
		Revoke:   false,
	}

	err := dao.Processor.ProcessDelegationTx(maxDurationTx, delegator)
	if err == nil {
		t.Error("Expected error for exceeding maximum duration")
	}

	// Test exactly maximum duration (should succeed)
	validDurationTx := &DelegationTx{
		Fee:      100,
		Delegate: delegate,
		Duration: 365 * 24 * 3600, // Exactly 1 year
		Revoke:   false,
	}

	err = dao.Processor.ProcessDelegationTx(validDurationTx, delegator)
	if err != nil {
		t.Fatalf("Failed to create delegation with maximum duration: %v", err)
	}
}
