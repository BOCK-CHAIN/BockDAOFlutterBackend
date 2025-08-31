package dao

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// WalletProvider represents different wallet providers
type WalletProvider string

const (
	WalletProviderMetaMask      WalletProvider = "metamask"
	WalletProviderWalletConnect WalletProvider = "walletconnect"
	WalletProviderManual        WalletProvider = "manual"
	WalletProviderLedger        WalletProvider = "ledger"
)

// WalletConnection represents a connected wallet
type WalletConnection struct {
	Provider    WalletProvider   `json:"provider"`
	Address     crypto.PublicKey `json:"address"`
	PublicKey   crypto.PublicKey `json:"publicKey"`
	ChainID     string           `json:"chainId,omitempty"`
	ConnectedAt time.Time        `json:"connectedAt"`
	LastActive  time.Time        `json:"lastActive"`
	IsActive    bool             `json:"isActive"`
}

// SignedTransaction represents a signed transaction
type SignedTransaction struct {
	Transaction     interface{}      `json:"transaction"`
	Signature       crypto.Signature `json:"signature"`
	TransactionHash types.Hash       `json:"transactionHash"`
	Signer          crypto.PublicKey `json:"signer"`
	SigningMethod   string           `json:"signingMethod"`
	Timestamp       time.Time        `json:"timestamp"`
}

// WalletIntegrationService handles wallet connections and transaction signing
type WalletIntegrationService struct {
	connections map[string]*WalletConnection
	validators  map[WalletProvider]TransactionValidator
}

// TransactionValidator interface for validating transactions from different wallet providers
type TransactionValidator interface {
	ValidateSignature(tx interface{}, signature crypto.Signature, publicKey crypto.PublicKey) error
	FormatTransaction(tx interface{}) ([]byte, error)
}

// NewWalletIntegrationService creates a new wallet integration service
func NewWalletIntegrationService() *WalletIntegrationService {
	service := &WalletIntegrationService{
		connections: make(map[string]*WalletConnection),
		validators:  make(map[WalletProvider]TransactionValidator),
	}

	// Register default validators
	service.validators[WalletProviderMetaMask] = &MetaMaskValidator{}
	service.validators[WalletProviderWalletConnect] = &WalletConnectValidator{}
	service.validators[WalletProviderManual] = &ManualWalletValidator{}
	service.validators[WalletProviderLedger] = &LedgerValidator{}

	return service
}

// ConnectWallet establishes a connection with a wallet
func (w *WalletIntegrationService) ConnectWallet(provider WalletProvider, address crypto.PublicKey, publicKey crypto.PublicKey, chainID string) (*WalletConnection, error) {
	addressStr := address.String()

	// Check if wallet is already connected
	if existing, exists := w.connections[addressStr]; exists && existing.IsActive {
		existing.LastActive = time.Now()
		return existing, nil
	}

	connection := &WalletConnection{
		Provider:    provider,
		Address:     address,
		PublicKey:   publicKey,
		ChainID:     chainID,
		ConnectedAt: time.Now(),
		LastActive:  time.Now(),
		IsActive:    true,
	}

	w.connections[addressStr] = connection
	return connection, nil
}

// DisconnectWallet disconnects a wallet
func (w *WalletIntegrationService) DisconnectWallet(address crypto.PublicKey) error {
	addressStr := address.String()

	if connection, exists := w.connections[addressStr]; exists {
		connection.IsActive = false
		connection.LastActive = time.Now()
	}

	return nil
}

// GetConnection retrieves a wallet connection
func (w *WalletIntegrationService) GetConnection(address crypto.PublicKey) (*WalletConnection, error) {
	addressStr := address.String()

	connection, exists := w.connections[addressStr]
	if !exists || !connection.IsActive {
		return nil, fmt.Errorf("wallet not connected: %s", addressStr)
	}

	connection.LastActive = time.Now()
	return connection, nil
}

// SignTransaction signs a transaction using the appropriate wallet provider
func (w *WalletIntegrationService) SignTransaction(address crypto.PublicKey, transaction interface{}, signature crypto.Signature) (*SignedTransaction, error) {
	connection, err := w.GetConnection(address)
	if err != nil {
		return nil, err
	}

	validator, exists := w.validators[connection.Provider]
	if !exists {
		return nil, fmt.Errorf("no validator for provider: %s", connection.Provider)
	}

	// Validate the signature
	if err := validator.ValidateSignature(transaction, signature, connection.PublicKey); err != nil {
		return nil, fmt.Errorf("signature validation failed: %w", err)
	}

	// Calculate transaction hash
	txData, err := validator.FormatTransaction(transaction)
	if err != nil {
		return nil, fmt.Errorf("transaction formatting failed: %w", err)
	}

	hash := sha256.Sum256(txData)
	txHash := types.HashFromBytes(hash[:])

	signedTx := &SignedTransaction{
		Transaction:     transaction,
		Signature:       signature,
		TransactionHash: txHash,
		Signer:          address,
		SigningMethod:   string(connection.Provider),
		Timestamp:       time.Now(),
	}

	return signedTx, nil
}

// VerifySignedTransaction verifies a signed transaction
func (w *WalletIntegrationService) VerifySignedTransaction(signedTx *SignedTransaction) error {
	connection, err := w.GetConnection(signedTx.Signer)
	if err != nil {
		return err
	}

	validator, exists := w.validators[connection.Provider]
	if !exists {
		return fmt.Errorf("no validator for provider: %s", connection.Provider)
	}

	return validator.ValidateSignature(signedTx.Transaction, signedTx.Signature, connection.PublicKey)
}

// GetActiveConnections returns all active wallet connections
func (w *WalletIntegrationService) GetActiveConnections() []*WalletConnection {
	var active []*WalletConnection

	for _, connection := range w.connections {
		if connection.IsActive {
			active = append(active, connection)
		}
	}

	return active
}

// CleanupInactiveConnections removes inactive connections older than the specified duration
func (w *WalletIntegrationService) CleanupInactiveConnections(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)

	for address, connection := range w.connections {
		if !connection.IsActive && connection.LastActive.Before(cutoff) {
			delete(w.connections, address)
		}
	}
}

// MetaMaskValidator validates MetaMask transactions
type MetaMaskValidator struct{}

func (m *MetaMaskValidator) ValidateSignature(tx interface{}, signature crypto.Signature, publicKey crypto.PublicKey) error {
	// Format transaction for MetaMask (EIP-712 style)
	txData, err := m.FormatTransaction(tx)
	if err != nil {
		return err
	}

	// Verify signature
	if !signature.Verify(publicKey, txData) {
		return fmt.Errorf("invalid MetaMask signature")
	}

	return nil
}

func (m *MetaMaskValidator) FormatTransaction(tx interface{}) ([]byte, error) {
	// Convert transaction to EIP-712 format
	eip712Data := map[string]interface{}{
		"types": map[string]interface{}{
			"EIP712Domain": []map[string]string{
				{"name": "name", "type": "string"},
				{"name": "version", "type": "string"},
				{"name": "chainId", "type": "uint256"},
			},
			"Transaction": []map[string]string{
				{"name": "to", "type": "address"},
				{"name": "value", "type": "uint256"},
				{"name": "data", "type": "bytes"},
				{"name": "nonce", "type": "uint256"},
			},
		},
		"primaryType": "Transaction",
		"domain": map[string]interface{}{
			"name":    "ProjectX DAO",
			"version": "1",
			"chainId": 1,
		},
		"message": tx,
	}

	return json.Marshal(eip712Data)
}

// WalletConnectValidator validates WalletConnect transactions
type WalletConnectValidator struct{}

func (w *WalletConnectValidator) ValidateSignature(tx interface{}, signature crypto.Signature, publicKey crypto.PublicKey) error {
	txData, err := w.FormatTransaction(tx)
	if err != nil {
		return err
	}

	if !signature.Verify(publicKey, txData) {
		return fmt.Errorf("invalid WalletConnect signature")
	}

	return nil
}

func (w *WalletConnectValidator) FormatTransaction(tx interface{}) ([]byte, error) {
	// Format for WalletConnect personal_sign
	return json.Marshal(tx)
}

// ManualWalletValidator validates manual wallet transactions
type ManualWalletValidator struct{}

func (m *ManualWalletValidator) ValidateSignature(tx interface{}, signature crypto.Signature, publicKey crypto.PublicKey) error {
	txData, err := m.FormatTransaction(tx)
	if err != nil {
		return err
	}

	if !signature.Verify(publicKey, txData) {
		return fmt.Errorf("invalid manual wallet signature")
	}

	return nil
}

func (m *ManualWalletValidator) FormatTransaction(tx interface{}) ([]byte, error) {
	// Simple JSON serialization for manual wallets
	return json.Marshal(tx)
}

// LedgerValidator validates Ledger hardware wallet transactions
type LedgerValidator struct{}

func (l *LedgerValidator) ValidateSignature(tx interface{}, signature crypto.Signature, publicKey crypto.PublicKey) error {
	txData, err := l.FormatTransaction(tx)
	if err != nil {
		return err
	}

	if !signature.Verify(publicKey, txData) {
		return fmt.Errorf("invalid Ledger signature")
	}

	return nil
}

func (l *LedgerValidator) FormatTransaction(tx interface{}) ([]byte, error) {
	// Format for Ledger signing (similar to manual but with specific encoding)
	return json.Marshal(tx)
}

// TransactionSigner provides utilities for transaction signing
type TransactionSigner struct {
	privateKey crypto.PrivateKey
}

// NewTransactionSigner creates a new transaction signer
func NewTransactionSigner(privateKey crypto.PrivateKey) *TransactionSigner {
	return &TransactionSigner{
		privateKey: privateKey,
	}
}

// SignDAOTransaction signs a DAO transaction
func (t *TransactionSigner) SignDAOTransaction(tx interface{}) (*crypto.Signature, error) {
	// Serialize transaction
	txData, err := json.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Sign the transaction data
	signature, err := t.privateKey.Sign(txData)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return signature, nil
}

// GenerateTestWallet generates a test wallet for development
func GenerateTestWallet() (crypto.PrivateKey, crypto.PublicKey, types.Address, error) {
	// Generate private key using the crypto package
	privateKey := crypto.GeneratePrivateKey()

	// Get public key
	publicKey := privateKey.PublicKey()

	// Generate address
	address := publicKey.Address()

	return privateKey, publicKey, address, nil
}

// WalletConnectionManager manages multiple wallet connections
type WalletConnectionManager struct {
	service *WalletIntegrationService
}

// NewWalletConnectionManager creates a new wallet connection manager
func NewWalletConnectionManager() *WalletConnectionManager {
	return &WalletConnectionManager{
		service: NewWalletIntegrationService(),
	}
}

// HandleWalletConnection handles a new wallet connection request
func (w *WalletConnectionManager) HandleWalletConnection(provider WalletProvider, address, publicKey string, chainID string) (*WalletConnection, error) {
	// Parse address and public key
	addressBytes, err := hex.DecodeString(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address format: %w", err)
	}

	publicKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid public key format: %w", err)
	}

	addr := crypto.PublicKey(addressBytes)
	pubKey := crypto.PublicKey(publicKeyBytes)

	return w.service.ConnectWallet(provider, addr, pubKey, chainID)
}

// HandleTransactionSigning handles transaction signing requests
func (w *WalletConnectionManager) HandleTransactionSigning(address string, transaction interface{}, signatureHex string) (*SignedTransaction, error) {
	// Parse address
	addressBytes, err := hex.DecodeString(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address format: %w", err)
	}

	// Parse signature
	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return nil, fmt.Errorf("invalid signature format: %w", err)
	}

	// Convert to crypto.Signature (simplified)
	if len(sigBytes) < 64 {
		return nil, fmt.Errorf("signature too short")
	}

	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:64])
	signature := crypto.Signature{R: r, S: s}

	addr := crypto.PublicKey(addressBytes)

	return w.service.SignTransaction(addr, transaction, signature)
}

// GetWalletInfo returns wallet information
func (w *WalletConnectionManager) GetWalletInfo(address string) (*WalletConnection, error) {
	addressBytes, err := hex.DecodeString(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address format: %w", err)
	}

	addr := crypto.PublicKey(addressBytes)
	return w.service.GetConnection(addr)
}

// DisconnectWallet disconnects a wallet
func (w *WalletConnectionManager) DisconnectWallet(address string) error {
	addressBytes, err := hex.DecodeString(address)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}

	addr := crypto.PublicKey(addressBytes)
	return w.service.DisconnectWallet(addr)
}
