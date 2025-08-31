package dao

import (
	"encoding/hex"
	"testing"
	"time"
)

// Use the existing randomHash function from dao_test.go

func TestWalletIntegrationService_ConnectWallet(t *testing.T) {
	service := NewWalletIntegrationService()

	// Generate test keys
	_, publicKey, _, err := GenerateTestWallet()
	if err != nil {
		t.Fatalf("Failed to generate test wallet: %v", err)
	}

	address := publicKey

	// Test connecting a MetaMask wallet
	connection, err := service.ConnectWallet(WalletProviderMetaMask, address, publicKey, "0x1")
	if err != nil {
		t.Fatalf("Failed to connect wallet: %v", err)
	}

	if connection.Provider != WalletProviderMetaMask {
		t.Errorf("Expected provider %s, got %s", WalletProviderMetaMask, connection.Provider)
	}

	if connection.Address.String() != address.String() {
		t.Errorf("Expected address %s, got %s", address.String(), connection.Address.String())
	}

	if !connection.IsActive {
		t.Error("Expected wallet to be active")
	}

	// Test connecting the same wallet again
	connection2, err := service.ConnectWallet(WalletProviderMetaMask, address, publicKey, "0x1")
	if err != nil {
		t.Fatalf("Failed to reconnect wallet: %v", err)
	}

	if connection2 != connection {
		t.Error("Expected same connection instance for reconnection")
	}
}

func TestWalletIntegrationService_DisconnectWallet(t *testing.T) {
	service := NewWalletIntegrationService()

	// Generate test keys
	privateKey, publicKey, address, err := GenerateTestWallet()
	if err != nil {
		t.Fatalf("Failed to generate test wallet: %v", err)
	}

	// Connect wallet
	_, err = service.ConnectWallet(WalletProviderManual, publicKey, publicKey, "")
	if err != nil {
		t.Fatalf("Failed to connect wallet: %v", err)
	}

	// Disconnect wallet
	err = service.DisconnectWallet(publicKey)
	if err != nil {
		t.Fatalf("Failed to disconnect wallet: %v", err)
	}

	// Try to get connection after disconnect
	_, err = service.GetConnection(publicKey)
	if err == nil {
		t.Error("Expected error when getting disconnected wallet")
	}

	_ = privateKey
	_ = address
}

func TestWalletIntegrationService_SignTransaction(t *testing.T) {
	service := NewWalletIntegrationService()

	// Generate test keys
	privateKey, publicKey, _, err := GenerateTestWallet()
	if err != nil {
		t.Fatalf("Failed to generate test wallet: %v", err)
	}

	// Connect wallet
	_, err = service.ConnectWallet(WalletProviderManual, publicKey, publicKey, "")
	if err != nil {
		t.Fatalf("Failed to connect wallet: %v", err)
	}

	// Create test transaction
	tx := &ProposalTx{
		Fee:         1000,
		Title:       "Test Proposal",
		Description: "This is a test proposal",
		StartTime:   time.Now().Unix(),
		EndTime:     time.Now().Add(24 * time.Hour).Unix(),
	}

	// Sign transaction
	signer := NewTransactionSigner(privateKey)
	signature, err := signer.SignDAOTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	// Verify signed transaction
	signedTx, err := service.SignTransaction(publicKey, tx, *signature)
	if err != nil {
		t.Fatalf("Failed to create signed transaction: %v", err)
	}

	if signedTx.Signer.String() != publicKey.String() {
		t.Errorf("Expected signer %s, got %s", publicKey.String(), signedTx.Signer.String())
	}

	if signedTx.SigningMethod != string(WalletProviderManual) {
		t.Errorf("Expected signing method %s, got %s", WalletProviderManual, signedTx.SigningMethod)
	}
}

func TestWalletIntegrationService_VerifySignedTransaction(t *testing.T) {
	service := NewWalletIntegrationService()

	// Generate test keys
	privateKey, publicKey, _, err := GenerateTestWallet()
	if err != nil {
		t.Fatalf("Failed to generate test wallet: %v", err)
	}

	// Connect wallet
	_, err = service.ConnectWallet(WalletProviderManual, publicKey, publicKey, "")
	if err != nil {
		t.Fatalf("Failed to connect wallet: %v", err)
	}

	// Create and sign transaction
	tx := &VoteTx{
		Fee:        500,
		ProposalID: randomHash(),
		Choice:     VoteChoiceYes,
		Weight:     100,
		Reason:     "I support this proposal",
	}

	signer := NewTransactionSigner(privateKey)
	signature, err := signer.SignDAOTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	signedTx, err := service.SignTransaction(publicKey, tx, *signature)
	if err != nil {
		t.Fatalf("Failed to create signed transaction: %v", err)
	}

	// Verify the signed transaction
	err = service.VerifySignedTransaction(signedTx)
	if err != nil {
		t.Fatalf("Failed to verify signed transaction: %v", err)
	}
}

func TestWalletIntegrationService_GetActiveConnections(t *testing.T) {
	service := NewWalletIntegrationService()

	// Generate multiple test wallets
	var connections []*WalletConnection
	for i := 0; i < 3; i++ {
		_, publicKey, _, err := GenerateTestWallet()
		if err != nil {
			t.Fatalf("Failed to generate test wallet %d: %v", i, err)
		}

		connection, err := service.ConnectWallet(WalletProviderManual, publicKey, publicKey, "")
		if err != nil {
			t.Fatalf("Failed to connect wallet %d: %v", i, err)
		}
		connections = append(connections, connection)
	}

	// Get active connections
	active := service.GetActiveConnections()
	if len(active) != 3 {
		t.Errorf("Expected 3 active connections, got %d", len(active))
	}

	// Disconnect one wallet
	err := service.DisconnectWallet(connections[0].Address)
	if err != nil {
		t.Fatalf("Failed to disconnect wallet: %v", err)
	}

	// Check active connections again
	active = service.GetActiveConnections()
	if len(active) != 2 {
		t.Errorf("Expected 2 active connections after disconnect, got %d", len(active))
	}
}

func TestWalletIntegrationService_CleanupInactiveConnections(t *testing.T) {
	service := NewWalletIntegrationService()

	// Generate test wallet
	_, publicKey, _, err := GenerateTestWallet()
	if err != nil {
		t.Fatalf("Failed to generate test wallet: %v", err)
	}

	// Connect and immediately disconnect
	_, err = service.ConnectWallet(WalletProviderManual, publicKey, publicKey, "")
	if err != nil {
		t.Fatalf("Failed to connect wallet: %v", err)
	}

	err = service.DisconnectWallet(publicKey)
	if err != nil {
		t.Fatalf("Failed to disconnect wallet: %v", err)
	}

	// Manually set last active time to past
	addressStr := publicKey.String()
	if connection, exists := service.connections[addressStr]; exists {
		connection.LastActive = time.Now().Add(-2 * time.Hour)
	}

	// Cleanup connections older than 1 hour
	service.CleanupInactiveConnections(1 * time.Hour)

	// Check that connection was removed
	if _, exists := service.connections[addressStr]; exists {
		t.Error("Expected inactive connection to be cleaned up")
	}
}

func TestMetaMaskValidator(t *testing.T) {
	validator := &MetaMaskValidator{}

	// Create test transaction
	tx := &ProposalTx{
		Fee:         1000,
		Title:       "Test Proposal",
		Description: "This is a test proposal",
	}

	// Test transaction formatting
	txData, err := validator.FormatTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to format transaction: %v", err)
	}

	if len(txData) == 0 {
		t.Error("Expected non-empty transaction data")
	}

	// Generate test signature
	privateKey, publicKey, _, err := GenerateTestWallet()
	if err != nil {
		t.Fatalf("Failed to generate test wallet: %v", err)
	}

	signer := NewTransactionSigner(privateKey)
	signature, err := signer.SignDAOTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	// Test signature validation
	err = validator.ValidateSignature(tx, *signature, publicKey)
	if err != nil {
		t.Fatalf("Failed to validate signature: %v", err)
	}
}

func TestWalletConnectionManager(t *testing.T) {
	manager := NewWalletConnectionManager()

	// Generate test wallet
	privateKey, publicKey, address, err := GenerateTestWallet()
	if err != nil {
		t.Fatalf("Failed to generate test wallet: %v", err)
	}

	addressHex := hex.EncodeToString(address[:])
	publicKeyHex := hex.EncodeToString(publicKey)

	// Test wallet connection
	connection, err := manager.HandleWalletConnection(WalletProviderManual, addressHex, publicKeyHex, "")
	if err != nil {
		t.Fatalf("Failed to handle wallet connection: %v", err)
	}

	if connection.Provider != WalletProviderManual {
		t.Errorf("Expected provider %s, got %s", WalletProviderManual, connection.Provider)
	}

	// Test getting wallet info
	info, err := manager.GetWalletInfo(addressHex)
	if err != nil {
		t.Fatalf("Failed to get wallet info: %v", err)
	}

	if info.Address.String() != publicKey.String() {
		t.Errorf("Expected address %s, got %s", publicKey.String(), info.Address.String())
	}

	// Test transaction signing
	tx := &VoteTx{
		Fee:        500,
		ProposalID: randomHash(),
		Choice:     VoteChoiceYes,
		Weight:     100,
	}

	signer := NewTransactionSigner(privateKey)
	signature, err := signer.SignDAOTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	signatureHex := hex.EncodeToString(append(signature.R.Bytes(), signature.S.Bytes()...))

	signedTx, err := manager.HandleTransactionSigning(addressHex, tx, signatureHex)
	if err != nil {
		t.Fatalf("Failed to handle transaction signing: %v", err)
	}

	if signedTx.SigningMethod != string(WalletProviderManual) {
		t.Errorf("Expected signing method %s, got %s", WalletProviderManual, signedTx.SigningMethod)
	}

	// Test wallet disconnection
	err = manager.DisconnectWallet(addressHex)
	if err != nil {
		t.Fatalf("Failed to disconnect wallet: %v", err)
	}
}

func TestGenerateTestWallet(t *testing.T) {
	privateKey, publicKey, address, err := GenerateTestWallet()
	if err != nil {
		t.Fatalf("Failed to generate test wallet: %v", err)
	}

	// Verify that keys are related
	derivedPublicKey := privateKey.PublicKey()
	if derivedPublicKey.String() != publicKey.String() {
		t.Error("Public key doesn't match derived public key from private key")
	}

	// Verify that address is derived from public key
	derivedAddress := publicKey.Address()
	if derivedAddress.String() != address.String() {
		t.Error("Address doesn't match derived address from public key")
	}

	// Test signing with generated wallet
	testData := []byte("test message")
	signature, err := privateKey.Sign(testData)
	if err != nil {
		t.Fatalf("Failed to sign test data: %v", err)
	}

	// Verify signature
	if !signature.Verify(publicKey, testData) {
		t.Error("Generated wallet signature verification failed")
	}
}

func TestTransactionSigner(t *testing.T) {
	// Generate test wallet
	privateKey, publicKey, _, err := GenerateTestWallet()
	if err != nil {
		t.Fatalf("Failed to generate test wallet: %v", err)
	}

	signer := NewTransactionSigner(privateKey)

	// Test signing different transaction types
	testCases := []interface{}{
		&ProposalTx{
			Fee:         1000,
			Title:       "Test Proposal",
			Description: "Test Description",
		},
		&VoteTx{
			Fee:        500,
			ProposalID: randomHash(),
			Choice:     VoteChoiceYes,
			Weight:     100,
		},
		&DelegationTx{
			Fee:      250,
			Delegate: publicKey,
			Duration: 86400, // 1 day
		},
		&TreasuryTx{
			Fee:       750,
			Recipient: publicKey,
			Amount:    5000,
			Purpose:   "Test payment",
		},
	}

	for i, tx := range testCases {
		signature, err := signer.SignDAOTransaction(tx)
		if err != nil {
			t.Fatalf("Failed to sign transaction %d: %v", i, err)
		}

		if signature == nil {
			t.Errorf("Expected non-nil signature for transaction %d", i)
		}

		if signature.R == nil || signature.S == nil {
			t.Errorf("Expected valid R and S values in signature for transaction %d", i)
		}
	}
}

// Benchmark tests
func BenchmarkWalletConnection(b *testing.B) {
	service := NewWalletIntegrationService()

	// Generate test keys
	_, publicKey, _, err := GenerateTestWallet()
	if err != nil {
		b.Fatalf("Failed to generate test wallet: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ConnectWallet(WalletProviderManual, publicKey, publicKey, "")
		if err != nil {
			b.Fatalf("Failed to connect wallet: %v", err)
		}
	}
}

func BenchmarkTransactionSigning(b *testing.B) {
	// Generate test wallet
	privateKey, publicKey, _, err := GenerateTestWallet()
	if err != nil {
		b.Fatalf("Failed to generate test wallet: %v", err)
	}

	signer := NewTransactionSigner(privateKey)
	tx := &VoteTx{
		Fee:        500,
		ProposalID: randomHash(),
		Choice:     VoteChoiceYes,
		Weight:     100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := signer.SignDAOTransaction(tx)
		if err != nil {
			b.Fatalf("Failed to sign transaction: %v", err)
		}
	}

	_ = publicKey
}
