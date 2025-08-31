package dao

import (
	"testing"

	"github.com/BOCK-CHAIN/BockChain/crypto"
)

func TestDelegationVotingIntegration(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	delegator := crypto.GeneratePrivateKey().PublicKey()
	delegate := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		delegator.String(): 5000,
		delegate.String():  3000,
	}
	dao.InitialTokenDistribution(distributions)

	// Test initial voting powers
	initialDelegatorPower := dao.GetEffectiveVotingPower(delegator)
	initialDelegatePower := dao.GetEffectiveVotingPower(delegate)

	if initialDelegatorPower != 5000 {
		t.Errorf("Expected initial delegator power 5000, got %d", initialDelegatorPower)
	}

	if initialDelegatePower != 3000 {
		t.Errorf("Expected initial delegate power 3000, got %d", initialDelegatePower)
	}

	// Create delegation
	delegationTx := &DelegationTx{
		Fee:      100,
		Delegate: delegate,
		Duration: 86400, // 1 day
		Revoke:   false,
	}

	err := dao.Processor.ProcessDelegationTx(delegationTx, delegator)
	if err != nil {
		t.Fatalf("Failed to create delegation: %v", err)
	}

	// Test voting powers after delegation
	delegatorPowerAfterDelegation := dao.GetEffectiveVotingPower(delegator)
	delegatePowerAfterDelegation := dao.GetEffectiveVotingPower(delegate)

	if delegatorPowerAfterDelegation != 0 {
		t.Errorf("Expected delegator power 0 after delegation, got %d", delegatorPowerAfterDelegation)
	}

	expectedDelegatePower := uint64(3000 + 4900) // own + delegated balance
	if delegatePowerAfterDelegation != expectedDelegatePower {
		t.Errorf("Expected delegate power %d after delegation, got %d", expectedDelegatePower, delegatePowerAfterDelegation)
	}

	// Test revocation
	revokeTx := &DelegationTx{
		Fee:      50,
		Delegate: delegate,
		Duration: 0,
		Revoke:   true,
	}

	err = dao.Processor.ProcessDelegationTx(revokeTx, delegator)
	if err != nil {
		t.Fatalf("Failed to revoke delegation: %v", err)
	}

	// Check powers after revocation
	delegatorPowerAfterRevocation := dao.GetEffectiveVotingPower(delegator)
	delegatePowerAfterRevocation := dao.GetEffectiveVotingPower(delegate)

	expectedDelegatorPowerAfterRevocation := uint64(5000 - 100 - 50) // original - fees
	if delegatorPowerAfterRevocation != expectedDelegatorPowerAfterRevocation {
		t.Errorf("Expected delegator power %d after revocation, got %d", expectedDelegatorPowerAfterRevocation, delegatorPowerAfterRevocation)
	}

	if delegatePowerAfterRevocation != 3000 {
		t.Errorf("Expected delegate power 3000 after revocation, got %d", delegatePowerAfterRevocation)
	}
}
