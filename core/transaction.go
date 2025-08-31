package core

import (
	"encoding/gob"
	"fmt"
	"math/rand"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/dao"
	"github.com/BOCK-CHAIN/BockChain/types"
)

type TxType byte

const (
	TxTypeCollection TxType = iota // 0x0
	TxTypeMint                     // 0x01
	// DAO transaction types start from 0x10
	TxTypeDAOProposal          = 0x10
	TxTypeDAOVote              = 0x11
	TxTypeDAODelegation        = 0x12
	TxTypeDAOTreasury          = 0x13
	TxTypeDAOTokenMint         = 0x14
	TxTypeDAOTokenBurn         = 0x15
	TxTypeDAOTokenTransfer     = 0x16
	TxTypeDAOTokenApprove      = 0x17
	TxTypeDAOTokenTransferFrom = 0x18
	TxTypeDAOParameter         = 0x19
)

type CollectionTx struct {
	Fee      int64
	MetaData []byte
}

type MintTx struct {
	Fee             int64
	NFT             types.Hash
	Collection      types.Hash
	MetaData        []byte
	CollectionOwner crypto.PublicKey
	Signature       crypto.Signature
}

type Transaction struct {
	// Only used for native NFT logic
	TxInner any
	// Any arbitrary data for the VM
	Data      []byte
	To        crypto.PublicKey
	Value     uint64
	From      crypto.PublicKey
	Signature *crypto.Signature
	Nonce     int64

	// cached version of the tx data hash
	hash types.Hash
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data:  data,
		Nonce: rand.Int63n(1000000000000000),
	}
}

func (tx *Transaction) Hash(hasher Hasher[*Transaction]) types.Hash {
	if tx.hash.IsZero() {
		tx.hash = hasher.Hash(tx)
	}
	return tx.hash
}

func (tx *Transaction) Sign(privKey crypto.PrivateKey) error {
	hash := tx.Hash(TxHasher{})
	sig, err := privKey.Sign(hash.ToSlice())
	if err != nil {
		return err
	}

	tx.From = privKey.PublicKey()
	tx.Signature = sig

	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil {
		return fmt.Errorf("transaction has no signature")
	}

	hash := tx.Hash(TxHasher{})
	if !tx.Signature.Verify(tx.From, hash.ToSlice()) {
		return fmt.Errorf("invalid transaction signature")
	}

	return nil
}

func (tx *Transaction) Decode(dec Decoder[*Transaction]) error {
	return dec.Decode(tx)
}

func (tx *Transaction) Encode(enc Encoder[*Transaction]) error {
	return enc.Encode(tx)
}

func init() {
	gob.Register(CollectionTx{})
	gob.Register(MintTx{})
	// Register DAO transaction types
	gob.Register(dao.ProposalTx{})
	gob.Register(dao.VoteTx{})
	gob.Register(dao.DelegationTx{})
	gob.Register(dao.TreasuryTx{})
	gob.Register(dao.TokenMintTx{})
	gob.Register(dao.TokenBurnTx{})
	gob.Register(dao.TokenTransferTx{})
	gob.Register(dao.TokenApproveTx{})
	gob.Register(dao.TokenTransferFromTx{})
	gob.Register(dao.ParameterProposalTx{})
}
