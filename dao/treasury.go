package dao

import (
	"crypto/sha256"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// TreasuryManager handles multi-signature treasury operations
type TreasuryManager struct {
	governanceState *GovernanceState
	tokenState      *GovernanceToken
	validator       *DAOValidator
}

// NewTreasuryManager creates a new treasury manager
func NewTreasuryManager(governanceState *GovernanceState, tokenState *GovernanceToken) *TreasuryManager {
	validator := NewDAOValidator(governanceState, tokenState)
	return &TreasuryManager{
		governanceState: governanceState,
		tokenState:      tokenState,
		validator:       validator,
	}
}

// CreateTreasuryTransaction creates a new treasury transaction
func (tm *TreasuryManager) CreateTreasuryTransaction(tx *TreasuryTx, txHash types.Hash) error {
	// Validate the transaction
	if err := tm.validator.ValidateTreasuryTx(tx); err != nil {
		return err
	}

	// Create pending treasury transaction
	pendingTx := &PendingTx{
		ID:         txHash,
		Recipient:  tx.Recipient,
		Amount:     tx.Amount,
		Purpose:    tx.Purpose,
		Signatures: make([]crypto.Signature, 0),
		CreatedAt:  time.Now().Unix(),
		ExpiresAt:  time.Now().Unix() + 86400, // 24 hours expiry
		Executed:   false,
	}

	// Store the pending transaction
	tm.governanceState.Treasury.Transactions[txHash] = pendingTx

	return nil
}

// SignTreasuryTransaction adds a signature to a pending treasury transaction
func (tm *TreasuryManager) SignTreasuryTransaction(txHash types.Hash, signer crypto.PrivateKey) error {
	// Get pending transaction
	pendingTx, exists := tm.governanceState.Treasury.Transactions[txHash]
	if !exists {
		return NewDAOError(ErrProposalNotFound, "treasury transaction not found", nil)
	}

	// Check if transaction has expired
	if time.Now().Unix() > pendingTx.ExpiresAt {
		return NewDAOError(ErrProposalExpired, "treasury transaction has expired", nil)
	}

	// Check if already executed
	if pendingTx.Executed {
		return NewDAOError(ErrInvalidProposal, "treasury transaction already executed", nil)
	}

	// Check if signer is authorized
	signerPubKey := signer.PublicKey()
	if !tm.isAuthorizedSigner(signerPubKey) {
		return NewDAOError(ErrUnauthorized, "signer not authorized for treasury operations", nil)
	}

	// Check if signer has already signed
	if tm.hasSignerSigned(pendingTx, signerPubKey) {
		return NewDAOError(ErrDuplicateVote, "signer has already signed this transaction", nil)
	}

	// Create transaction data for signing
	txData := tm.createTreasuryTxData(pendingTx)

	// Sign the transaction data
	signature, err := signer.Sign(txData)
	if err != nil {
		return NewDAOError(ErrInvalidSignature, "failed to sign transaction", nil)
	}

	// Add signature
	pendingTx.Signatures = append(pendingTx.Signatures, *signature)

	// Check if we have enough signatures to execute
	if len(pendingTx.Signatures) >= int(tm.governanceState.Treasury.RequiredSigs) {
		return tm.executeTreasuryTransaction(txHash)
	}

	return nil
}

// ExecuteTreasuryTransaction executes a treasury transaction if it has sufficient signatures
func (tm *TreasuryManager) ExecuteTreasuryTransaction(txHash types.Hash) error {
	pendingTx, exists := tm.governanceState.Treasury.Transactions[txHash]
	if !exists {
		return NewDAOError(ErrProposalNotFound, "treasury transaction not found", nil)
	}

	// Check if transaction has expired
	if time.Now().Unix() > pendingTx.ExpiresAt {
		return NewDAOError(ErrProposalExpired, "treasury transaction has expired", nil)
	}

	// Check if already executed
	if pendingTx.Executed {
		return NewDAOError(ErrInvalidProposal, "treasury transaction already executed", nil)
	}

	// Verify we have enough signatures
	if len(pendingTx.Signatures) < int(tm.governanceState.Treasury.RequiredSigs) {
		return NewDAOError(ErrInvalidSignature, "insufficient signatures for execution", nil)
	}

	// Verify all signatures
	if err := tm.verifyTreasurySignatures(pendingTx); err != nil {
		return err
	}

	return tm.executeTreasuryTransaction(txHash)
}

// executeTreasuryTransaction performs the actual treasury transaction execution
func (tm *TreasuryManager) executeTreasuryTransaction(txHash types.Hash) error {
	pendingTx := tm.governanceState.Treasury.Transactions[txHash]

	// Check treasury balance
	if tm.governanceState.Treasury.Balance < pendingTx.Amount {
		return ErrTreasuryInsufficientFunds
	}

	// Transfer funds from treasury
	tm.governanceState.Treasury.Balance -= pendingTx.Amount

	// Add to recipient's token balance
	recipientStr := pendingTx.Recipient.String()
	if tm.tokenState.Balances[recipientStr] == 0 {
		tm.tokenState.Balances[recipientStr] = pendingTx.Amount
	} else {
		tm.tokenState.Balances[recipientStr] += pendingTx.Amount
	}

	// Mark as executed
	pendingTx.Executed = true

	return nil
}

// GetPendingTreasuryTransactions returns all pending treasury transactions
func (tm *TreasuryManager) GetPendingTreasuryTransactions() map[types.Hash]*PendingTx {
	pending := make(map[types.Hash]*PendingTx)
	now := time.Now().Unix()

	for txHash, tx := range tm.governanceState.Treasury.Transactions {
		if !tx.Executed && now <= tx.ExpiresAt {
			pending[txHash] = tx
		}
	}

	return pending
}

// GetTreasuryTransaction returns a specific treasury transaction
func (tm *TreasuryManager) GetTreasuryTransaction(txHash types.Hash) (*PendingTx, bool) {
	tx, exists := tm.governanceState.Treasury.Transactions[txHash]
	return tx, exists
}

// AddTreasuryFunds adds funds to the treasury
func (tm *TreasuryManager) AddTreasuryFunds(amount uint64) {
	tm.governanceState.Treasury.Balance += amount
}

// GetTreasuryBalance returns the current treasury balance
func (tm *TreasuryManager) GetTreasuryBalance() uint64 {
	return tm.governanceState.Treasury.Balance
}

// GetTreasurySigners returns the list of authorized treasury signers
func (tm *TreasuryManager) GetTreasurySigners() []crypto.PublicKey {
	return tm.governanceState.Treasury.Signers
}

// GetRequiredSignatures returns the number of required signatures
func (tm *TreasuryManager) GetRequiredSignatures() uint8 {
	return tm.governanceState.Treasury.RequiredSigs
}

// UpdateTreasurySigners updates the treasury signers (requires governance approval)
func (tm *TreasuryManager) UpdateTreasurySigners(signers []crypto.PublicKey, requiredSigs uint8) error {
	if len(signers) == 0 {
		return NewDAOError(ErrInvalidProposal, "treasury must have at least one signer", nil)
	}

	if requiredSigs == 0 || requiredSigs > uint8(len(signers)) {
		return NewDAOError(ErrInvalidProposal, "invalid required signatures count", nil)
	}

	tm.governanceState.Treasury.Signers = signers
	tm.governanceState.Treasury.RequiredSigs = requiredSigs

	return nil
}

// CleanupExpiredTransactions removes expired treasury transactions
func (tm *TreasuryManager) CleanupExpiredTransactions() int {
	now := time.Now().Unix()
	cleaned := 0

	for txHash, tx := range tm.governanceState.Treasury.Transactions {
		if !tx.Executed && now > tx.ExpiresAt {
			delete(tm.governanceState.Treasury.Transactions, txHash)
			cleaned++
		}
	}

	return cleaned
}

// GetTreasuryHistory returns all treasury transactions (executed and pending)
func (tm *TreasuryManager) GetTreasuryHistory() map[types.Hash]*PendingTx {
	return tm.governanceState.Treasury.Transactions
}

// GetExecutedTreasuryTransactions returns only executed treasury transactions
func (tm *TreasuryManager) GetExecutedTreasuryTransactions() map[types.Hash]*PendingTx {
	executed := make(map[types.Hash]*PendingTx)

	for txHash, tx := range tm.governanceState.Treasury.Transactions {
		if tx.Executed {
			executed[txHash] = tx
		}
	}

	return executed
}

// isAuthorizedSigner checks if a public key is an authorized treasury signer
func (tm *TreasuryManager) isAuthorizedSigner(pubKey crypto.PublicKey) bool {
	pubKeyStr := pubKey.String()
	for _, signer := range tm.governanceState.Treasury.Signers {
		if signer.String() == pubKeyStr {
			return true
		}
	}
	return false
}

// hasSignerSigned checks if a signer has already signed a transaction
func (tm *TreasuryManager) hasSignerSigned(pendingTx *PendingTx, signer crypto.PublicKey) bool {
	txData := tm.createTreasuryTxData(pendingTx)

	for _, sig := range pendingTx.Signatures {
		if sig.Verify(signer, txData) {
			return true
		}
	}
	return false
}

// createTreasuryTxData creates the data to be signed for a treasury transaction
func (tm *TreasuryManager) createTreasuryTxData(pendingTx *PendingTx) []byte {
	// Create a deterministic hash of the transaction data
	hasher := sha256.New()
	hasher.Write(pendingTx.ID.ToSlice())
	hasher.Write([]byte(pendingTx.Recipient))
	hasher.Write([]byte{
		byte(pendingTx.Amount >> 56),
		byte(pendingTx.Amount >> 48),
		byte(pendingTx.Amount >> 40),
		byte(pendingTx.Amount >> 32),
		byte(pendingTx.Amount >> 24),
		byte(pendingTx.Amount >> 16),
		byte(pendingTx.Amount >> 8),
		byte(pendingTx.Amount),
	})
	hasher.Write([]byte(pendingTx.Purpose))
	hasher.Write([]byte{
		byte(pendingTx.CreatedAt >> 56),
		byte(pendingTx.CreatedAt >> 48),
		byte(pendingTx.CreatedAt >> 40),
		byte(pendingTx.CreatedAt >> 32),
		byte(pendingTx.CreatedAt >> 24),
		byte(pendingTx.CreatedAt >> 16),
		byte(pendingTx.CreatedAt >> 8),
		byte(pendingTx.CreatedAt),
	})

	return hasher.Sum(nil)
}

// verifyTreasurySignatures verifies all signatures on a treasury transaction
func (tm *TreasuryManager) verifyTreasurySignatures(pendingTx *PendingTx) error {
	txData := tm.createTreasuryTxData(pendingTx)
	validSignatures := 0

	// Check each signature against authorized signers
	for _, sig := range pendingTx.Signatures {
		signatureValid := false

		for _, signer := range tm.governanceState.Treasury.Signers {
			if sig.Verify(signer, txData) {
				signatureValid = true
				validSignatures++
				break
			}
		}

		if !signatureValid {
			return NewDAOError(ErrInvalidSignature, "invalid signature found in treasury transaction", nil)
		}
	}

	if validSignatures < int(tm.governanceState.Treasury.RequiredSigs) {
		return NewDAOError(ErrInvalidSignature, "insufficient valid signatures", nil)
	}

	return nil
}
