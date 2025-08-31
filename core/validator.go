package core

import (
	"errors"
	"fmt"

	"github.com/BOCK-CHAIN/BockChain/dao"
)

var ErrBlockKnown = errors.New("block already known")

type Validator interface {
	ValidateBlock(*Block) error
}

type BlockValidator struct {
	bc *Blockchain
}

func NewBlockValidator(bc *Blockchain) *BlockValidator {
	return &BlockValidator{
		bc: bc,
	}
}

func (v *BlockValidator) ValidateBlock(b *Block) error {
	if v.bc.HasBlock(b.Height) {
		// return fmt.Errorf("chain already contains block (%d) with hash (%s)", b.Height, b.Hash(BlockHasher{}))
		return ErrBlockKnown
	}

	if b.Height != v.bc.Height()+1 {
		return fmt.Errorf("block (%s) with height (%d) is too high => current height (%d)", b.Hash(BlockHasher{}), b.Height, v.bc.Height())
	}

	prevHeader, err := v.bc.GetHeader(b.Height - 1)
	if err != nil {
		return err
	}

	hash := BlockHasher{}.Hash(prevHeader)
	if hash != b.PrevBlockHash {
		return fmt.Errorf("the hash of the previous block (%s) is invalid", b.PrevBlockHash)
	}

	if err := b.Verify(); err != nil {
		return err
	}

	// Validate DAO transactions in the block
	for _, tx := range b.Transactions {
		if err := v.validateDAOTransaction(tx); err != nil {
			return fmt.Errorf("invalid DAO transaction in block: %w", err)
		}
	}

	return nil
}

// validateDAOTransaction validates DAO-specific transactions
func (v *BlockValidator) validateDAOTransaction(tx *Transaction) error {
	if tx.TxInner == nil {
		return nil // Not a DAO transaction
	}

	// Get DAO validator from blockchain
	daoValidator := dao.NewDAOValidator(v.bc.GetDAOState(), v.bc.GetDAOTokenState())

	switch t := tx.TxInner.(type) {
	case dao.ProposalTx:
		return daoValidator.ValidateProposalTx(&t, tx.From)

	case dao.VoteTx:
		return daoValidator.ValidateVoteTx(&t, tx.From)

	case dao.DelegationTx:
		return daoValidator.ValidateDelegationTx(&t, tx.From)

	case dao.TreasuryTx:
		return daoValidator.ValidateTreasuryTx(&t)

	case dao.TokenMintTx:
		return daoValidator.ValidateTokenMintTx(&t, tx.From)

	case dao.TokenBurnTx:
		return daoValidator.ValidateTokenBurnTx(&t, tx.From)

	case dao.TokenTransferTx:
		return daoValidator.ValidateTokenTransferTx(&t, tx.From)

	case dao.TokenApproveTx:
		return daoValidator.ValidateTokenApproveTx(&t, tx.From)

	case dao.TokenTransferFromTx:
		return daoValidator.ValidateTokenTransferFromTx(&t, tx.From)

	default:
		// Not a DAO transaction, no validation needed
		return nil
	}

	return nil
}
