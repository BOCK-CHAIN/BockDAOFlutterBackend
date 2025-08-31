package dao

import (
	"fmt"
	"log"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// TreasuryExample demonstrates the multi-signature treasury system
func TreasuryExample() {
	fmt.Println("=== ProjectX DAO Treasury System Example ===")

	// Create a new DAO
	dao := NewDAO("GOVX", "GovernanceX Token", 18)

	// Create treasury signers (multi-sig setup)
	signer1 := crypto.GeneratePrivateKey()
	signer2 := crypto.GeneratePrivateKey()
	signer3 := crypto.GeneratePrivateKey()
	signers := []crypto.PublicKey{signer1.PublicKey(), signer2.PublicKey(), signer3.PublicKey()}

	fmt.Printf("\n--- Treasury Initialization ---\n")
	fmt.Printf("Setting up treasury with 3 signers, requiring 2 signatures\n")

	// Initialize treasury with 2-of-3 multi-sig
	err := dao.InitializeTreasury(signers, 2)
	if err != nil {
		log.Fatalf("Failed to initialize treasury: %v", err)
	}

	fmt.Printf("Treasury initialized successfully\n")
	fmt.Printf("Signers: %d\n", len(dao.GetTreasurySigners()))
	fmt.Printf("Required signatures: %d\n", dao.GetRequiredSignatures())

	// Add initial funds to treasury
	initialFunds := uint64(1000000) // 1M tokens
	dao.AddTreasuryFunds(initialFunds)
	fmt.Printf("Added %d tokens to treasury\n", initialFunds)
	fmt.Printf("Treasury balance: %d tokens\n", dao.GetTreasuryBalance())

	// Create recipients for treasury disbursements
	developer := crypto.GeneratePrivateKey().PublicKey()
	marketing := crypto.GeneratePrivateKey().PublicKey()
	community := crypto.GeneratePrivateKey().PublicKey()

	fmt.Printf("\n--- Treasury Transaction Creation ---\n")

	// Create first treasury transaction (development funding)
	devTx := &TreasuryTx{
		Fee:          500,
		Recipient:    developer,
		Amount:       100000,
		Purpose:      "Q1 Development milestone payment",
		Signatures:   []crypto.Signature{},
		RequiredSigs: 2,
	}

	devTxHash := generateExampleHash("dev-funding-q1")
	err = dao.CreateTreasuryTransaction(devTx, devTxHash)
	if err != nil {
		log.Fatalf("Failed to create development treasury transaction: %v", err)
	}

	fmt.Printf("Created development funding transaction: %s\n", devTxHash.String()[:16]+"...")
	fmt.Printf("Amount: %d tokens\n", devTx.Amount)
	fmt.Printf("Purpose: %s\n", devTx.Purpose)

	// Create second treasury transaction (marketing funding)
	marketingTx := &TreasuryTx{
		Fee:          300,
		Recipient:    marketing,
		Amount:       50000,
		Purpose:      "Marketing campaign for Q1 launch",
		Signatures:   []crypto.Signature{},
		RequiredSigs: 2,
	}

	marketingTxHash := generateExampleHash("marketing-q1")
	err = dao.CreateTreasuryTransaction(marketingTx, marketingTxHash)
	if err != nil {
		log.Fatalf("Failed to create marketing treasury transaction: %v", err)
	}

	fmt.Printf("Created marketing funding transaction: %s\n", marketingTxHash.String()[:16]+"...")
	fmt.Printf("Amount: %d tokens\n", marketingTx.Amount)
	fmt.Printf("Purpose: %s\n", marketingTx.Purpose)

	// Show pending transactions
	fmt.Printf("\n--- Pending Treasury Transactions ---\n")
	pending := dao.GetPendingTreasuryTransactions()
	fmt.Printf("Pending transactions: %d\n", len(pending))

	for txHash, tx := range pending {
		fmt.Printf("- %s: %d tokens to %s\n",
			txHash.String()[:16]+"...",
			tx.Amount,
			tx.Recipient.String()[:16]+"...")
		fmt.Printf("  Purpose: %s\n", tx.Purpose)
		fmt.Printf("  Signatures: %d/%d\n", len(tx.Signatures), dao.GetRequiredSignatures())
	}

	fmt.Printf("\n--- Multi-Signature Process ---\n")

	// Sign development transaction with first signer
	fmt.Printf("Signer 1 signing development transaction...\n")
	err = dao.SignTreasuryTransaction(devTxHash, signer1)
	if err != nil {
		log.Fatalf("Failed to sign with signer 1: %v", err)
	}

	// Check transaction status
	devPendingTx, _ := dao.GetTreasuryTransaction(devTxHash)
	fmt.Printf("Development transaction signatures: %d/%d\n",
		len(devPendingTx.Signatures), dao.GetRequiredSignatures())
	fmt.Printf("Executed: %t\n", devPendingTx.Executed)

	// Sign development transaction with second signer (should execute)
	fmt.Printf("Signer 2 signing development transaction...\n")
	err = dao.SignTreasuryTransaction(devTxHash, signer2)
	if err != nil {
		log.Fatalf("Failed to sign with signer 2: %v", err)
	}

	// Check if transaction was executed
	devPendingTx, _ = dao.GetTreasuryTransaction(devTxHash)
	fmt.Printf("Development transaction executed: %t\n", devPendingTx.Executed)

	// Show updated balances
	fmt.Printf("\n--- Updated Balances ---\n")
	fmt.Printf("Treasury balance: %d tokens\n", dao.GetTreasuryBalance())
	fmt.Printf("Developer balance: %d tokens\n", dao.GetTokenBalance(developer))

	// Sign marketing transaction with different signers
	fmt.Printf("\n--- Marketing Transaction Signing ---\n")
	fmt.Printf("Signer 1 signing marketing transaction...\n")
	err = dao.SignTreasuryTransaction(marketingTxHash, signer1)
	if err != nil {
		log.Fatalf("Failed to sign marketing tx with signer 1: %v", err)
	}

	fmt.Printf("Signer 3 signing marketing transaction...\n")
	err = dao.SignTreasuryTransaction(marketingTxHash, signer3)
	if err != nil {
		log.Fatalf("Failed to sign marketing tx with signer 3: %v", err)
	}

	// Check marketing transaction execution
	marketingPendingTx, _ := dao.GetTreasuryTransaction(marketingTxHash)
	fmt.Printf("Marketing transaction executed: %t\n", marketingPendingTx.Executed)

	// Final balances
	fmt.Printf("\n--- Final Treasury State ---\n")
	fmt.Printf("Treasury balance: %d tokens\n", dao.GetTreasuryBalance())
	fmt.Printf("Developer balance: %d tokens\n", dao.GetTokenBalance(developer))
	fmt.Printf("Marketing balance: %d tokens\n", dao.GetTokenBalance(marketing))

	// Show executed transactions
	executed := dao.GetExecutedTreasuryTransactions()
	fmt.Printf("Executed transactions: %d\n", len(executed))

	for txHash, tx := range executed {
		fmt.Printf("- %s: %d tokens to %s\n",
			txHash.String()[:16]+"...",
			tx.Amount,
			tx.Recipient.String()[:16]+"...")
		fmt.Printf("  Purpose: %s\n", tx.Purpose)
	}

	// Demonstrate treasury signer updates
	fmt.Printf("\n--- Treasury Signer Management ---\n")
	fmt.Printf("Current signers: %d\n", len(dao.GetTreasurySigners()))

	// Add a new signer
	newSigner := crypto.GeneratePrivateKey()
	updatedSigners := append(signers, newSigner.PublicKey())

	err = dao.UpdateTreasurySigners(updatedSigners, 3) // Now require 3 of 4
	if err != nil {
		log.Fatalf("Failed to update treasury signers: %v", err)
	}

	fmt.Printf("Updated signers: %d\n", len(dao.GetTreasurySigners()))
	fmt.Printf("Required signatures: %d\n", dao.GetRequiredSignatures())

	// Create a transaction requiring 3 signatures
	communityTx := &TreasuryTx{
		Fee:          200,
		Recipient:    community,
		Amount:       25000,
		Purpose:      "Community rewards program",
		Signatures:   []crypto.Signature{},
		RequiredSigs: 3,
	}

	communityTxHash := generateExampleHash("community-rewards")
	err = dao.CreateTreasuryTransaction(communityTx, communityTxHash)
	if err != nil {
		log.Fatalf("Failed to create community treasury transaction: %v", err)
	}

	fmt.Printf("Created community transaction requiring 3 signatures\n")

	// Sign with 3 different signers
	fmt.Printf("Collecting 3 signatures...\n")
	err = dao.SignTreasuryTransaction(communityTxHash, signer1)
	if err != nil {
		log.Fatalf("Failed to sign with signer 1: %v", err)
	}

	err = dao.SignTreasuryTransaction(communityTxHash, signer2)
	if err != nil {
		log.Fatalf("Failed to sign with signer 2: %v", err)
	}

	err = dao.SignTreasuryTransaction(communityTxHash, newSigner)
	if err != nil {
		log.Fatalf("Failed to sign with new signer: %v", err)
	}

	// Check final execution
	communityPendingTx, _ := dao.GetTreasuryTransaction(communityTxHash)
	fmt.Printf("Community transaction executed: %t\n", communityPendingTx.Executed)

	// Final summary
	fmt.Printf("\n--- Treasury System Summary ---\n")
	fmt.Printf("Total treasury transactions: %d\n", len(dao.GetTreasuryHistory()))
	fmt.Printf("Executed transactions: %d\n", len(dao.GetExecutedTreasuryTransactions()))
	fmt.Printf("Pending transactions: %d\n", len(dao.GetPendingTreasuryTransactions()))
	fmt.Printf("Final treasury balance: %d tokens\n", dao.GetTreasuryBalance())
	fmt.Printf("Total disbursed: %d tokens\n",
		initialFunds-dao.GetTreasuryBalance())

	fmt.Printf("\n=== Treasury Example Complete ===\n")
}

// generateExampleHash creates a deterministic hash for examples
func generateExampleHash(seed string) types.Hash {
	bytes := make([]byte, 32)
	seedBytes := []byte(seed)

	for i := 0; i < 32; i++ {
		if i < len(seedBytes) {
			bytes[i] = seedBytes[i]
		} else {
			bytes[i] = byte(i + len(seedBytes))
		}
	}

	return types.HashFromBytes(bytes)
}
