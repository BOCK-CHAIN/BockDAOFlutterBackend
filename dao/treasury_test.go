package dao

import (
	"crypto/rand"
	"testing"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// randomTreasuryHash generates a random hash for testing
func randomTreasuryHash() types.Hash {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return types.HashFromBytes(bytes)
}

func TestTreasuryManager_CreateTreasuryTransaction(t *testing.T) {
	// Setup
	dao := NewDAO("GOV", "Governance Token", 18)

	// Initialize treasury with signers
	signer1 := crypto.GeneratePrivateKey()
	signer2 := crypto.GeneratePrivateKey()
	signers := []crypto.PublicKey{signer1.PublicKey(), signer2.PublicKey()}

	err := dao.InitializeTreasury(signers, 2)
	if err != nil {
		t.Fatalf("Failed to initialize treasury: %v", err)
	}

	// Add funds to treasury
	dao.AddTreasuryFunds(10000)

	// Create recipient
	recipient := crypto.GeneratePrivateKey().PublicKey()

	// Create treasury transaction
	tx := &TreasuryTx{
		Fee:          100,
		Recipient:    recipient,
		Amount:       5000,
		Purpose:      "Development funding",
		Signatures:   []crypto.Signature{},
		RequiredSigs: 2,
	}

	txHash := randomTreasuryHash()

	// Test successful creation
	err = dao.CreateTreasuryTransaction(tx, txHash)
	if err != nil {
		t.Fatalf("Failed to create treasury transaction: %v", err)
	}

	// Verify transaction was stored
	pendingTx, exists := dao.GetTreasuryTransaction(txHash)
	if !exists {
		t.Fatal("Treasury transaction was not stored")
	}

	if pendingTx.Amount != 5000 {
		t.Errorf("Expected amount 5000, got %d", pendingTx.Amount)
	}

	if pendingTx.Purpose != "Development funding" {
		t.Errorf("Expected purpose 'Development funding', got %s", pendingTx.Purpose)
	}

	if pendingTx.Executed {
		t.Error("Transaction should not be executed yet")
	}
}

func TestTreasuryManager_CreateTreasuryTransaction_InsufficientFunds(t *testing.T) {
	// Setup
	dao := NewDAO("GOV", "Governance Token", 18)

	// Initialize treasury with signers
	signer1 := crypto.GeneratePrivateKey()
	signer2 := crypto.GeneratePrivateKey()
	signers := []crypto.PublicKey{signer1.PublicKey(), signer2.PublicKey()}

	err := dao.InitializeTreasury(signers, 2)
	if err != nil {
		t.Fatalf("Failed to initialize treasury: %v", err)
	}

	// Add insufficient funds to treasury
	dao.AddTreasuryFunds(1000)

	// Create recipient
	recipient := crypto.GeneratePrivateKey().PublicKey()

	// Create treasury transaction with amount exceeding balance
	tx := &TreasuryTx{
		Fee:          100,
		Recipient:    recipient,
		Amount:       5000, // More than treasury balance
		Purpose:      "Development funding",
		Signatures:   []crypto.Signature{},
		RequiredSigs: 2,
	}

	txHash := randomTreasuryHash()

	// Test should fail due to insufficient funds
	err = dao.CreateTreasuryTransaction(tx, txHash)
	if err == nil {
		t.Fatal("Expected error for insufficient treasury funds")
	}

	if err != ErrTreasuryInsufficientFunds {
		t.Errorf("Expected ErrTreasuryInsufficientFunds, got %v", err)
	}
}

func TestTreasuryManager_SignTreasuryTransaction(t *testing.T) {
	// Setup
	dao := NewDAO("GOV", "Governance Token", 18)

	// Initialize treasury with signers
	signer1 := crypto.GeneratePrivateKey()
	signer2 := crypto.GeneratePrivateKey()
	signers := []crypto.PublicKey{signer1.PublicKey(), signer2.PublicKey()}

	err := dao.InitializeTreasury(signers, 2)
	if err != nil {
		t.Fatalf("Failed to initialize treasury: %v", err)
	}

	// Add funds to treasury
	dao.AddTreasuryFunds(10000)

	// Create recipient
	recipient := crypto.GeneratePrivateKey().PublicKey()

	// Create treasury transaction
	tx := &TreasuryTx{
		Fee:          100,
		Recipient:    recipient,
		Amount:       5000,
		Purpose:      "Development funding",
		Signatures:   []crypto.Signature{},
		RequiredSigs: 2,
	}

	txHash := randomTreasuryHash()

	// Create transaction
	err = dao.CreateTreasuryTransaction(tx, txHash)
	if err != nil {
		t.Fatalf("Failed to create treasury transaction: %v", err)
	}

	// Sign with first signer
	err = dao.SignTreasuryTransaction(txHash, signer1)
	if err != nil {
		t.Fatalf("Failed to sign treasury transaction: %v", err)
	}

	// Verify signature was added
	pendingTx, _ := dao.GetTreasuryTransaction(txHash)
	if len(pendingTx.Signatures) != 1 {
		t.Errorf("Expected 1 signature, got %d", len(pendingTx.Signatures))
	}

	// Transaction should not be executed yet (need 2 signatures)
	if pendingTx.Executed {
		t.Error("Transaction should not be executed with only 1 signature")
	}

	// Sign with second signer (should trigger execution)
	err = dao.SignTreasuryTransaction(txHash, signer2)
	if err != nil {
		t.Fatalf("Failed to sign treasury transaction with second signer: %v", err)
	}

	// Verify transaction was executed
	pendingTx, _ = dao.GetTreasuryTransaction(txHash)
	if !pendingTx.Executed {
		t.Error("Transaction should be executed after sufficient signatures")
	}

	// Verify treasury balance was reduced
	if dao.GetTreasuryBalance() != 5000 { // 10000 - 5000
		t.Errorf("Expected treasury balance 5000, got %d", dao.GetTreasuryBalance())
	}

	// Verify recipient received tokens
	recipientBalance := dao.GetTokenBalance(recipient)
	if recipientBalance != 5000 {
		t.Errorf("Expected recipient balance 5000, got %d", recipientBalance)
	}
}
