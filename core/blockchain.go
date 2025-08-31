package core

import (
	"fmt"
	"sync"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/dao"
	"github.com/BOCK-CHAIN/BockChain/types"
	"github.com/go-kit/log"
)

type Blockchain struct {
	logger log.Logger
	store  Storage
	// TODO: double check this!
	lock       sync.RWMutex
	headers    []*Header
	blocks     []*Block
	txStore    map[types.Hash]*Transaction
	blockStore map[types.Hash]*Block

	accountState *AccountState

	stateLock       sync.RWMutex
	collectionState map[types.Hash]*CollectionTx
	mintState       map[types.Hash]*MintTx
	validator       Validator
	// TODO: make this an interface.
	contractState *State

	// DAO state management
	daoState      *dao.GovernanceState
	daoTokenState *dao.GovernanceToken
	daoProcessor  *dao.DAOProcessor
}

func NewBlockchain(l log.Logger, genesis *Block) (*Blockchain, error) {
	// We should create all states inside the scope of the newblockchain.

	// TODO: read this from disk later on
	accountState := NewAccountState()

	coinbase := crypto.PublicKey{}
	accountState.CreateAccount(coinbase.Address())

	// Initialize DAO state
	daoState := dao.NewGovernanceState()
	daoTokenState := dao.NewGovernanceToken("GOVX", "ProjectX Governance Token", 18)
	daoProcessor := dao.NewDAOProcessor(daoState, daoTokenState)

	bc := &Blockchain{
		contractState:   NewState(),
		headers:         []*Header{},
		store:           NewMemorystore(),
		logger:          l,
		accountState:    accountState,
		collectionState: make(map[types.Hash]*CollectionTx),
		mintState:       make(map[types.Hash]*MintTx),
		blockStore:      make(map[types.Hash]*Block),
		txStore:         make(map[types.Hash]*Transaction),
		daoState:        daoState,
		daoTokenState:   daoTokenState,
		daoProcessor:    daoProcessor,
	}
	bc.validator = NewBlockValidator(bc)
	err := bc.addBlockWithoutValidation(genesis)

	return bc, err
}

func (bc *Blockchain) SetValidator(v Validator) {
	bc.validator = v
}

func (bc *Blockchain) AddBlock(b *Block) error {
	if err := bc.validator.ValidateBlock(b); err != nil {
		return err
	}

	return bc.addBlockWithoutValidation(b)
}

func (bc *Blockchain) handleNativeTransfer(tx *Transaction) error {
	bc.logger.Log(
		"msg", "handle native token transfer",
		"from", tx.From,
		"to", tx.To,
		"value", tx.Value)

	return bc.accountState.Transfer(tx.From.Address(), tx.To.Address(), tx.Value)
}

func (bc *Blockchain) handleNativeNFT(tx *Transaction) error {
	hash := tx.Hash(TxHasher{})

	switch t := tx.TxInner.(type) {
	case CollectionTx:
		bc.collectionState[hash] = &t
		bc.logger.Log("msg", "created new NFT collection", "hash", hash)
	case MintTx:
		_, ok := bc.collectionState[t.Collection]
		if !ok {
			return fmt.Errorf("collection (%s) does not exist on the blockchain", t.Collection)
		}
		bc.mintState[hash] = &t

		bc.logger.Log("msg", "created new NFT mint", "NFT", t.NFT, "collection", t.Collection)
	default:
		return fmt.Errorf("unsupported tx type %v", t)
	}

	return nil
}

// handleDAOTransaction processes DAO-specific transactions
func (bc *Blockchain) handleDAOTransaction(tx *Transaction) error {
	hash := tx.Hash(TxHasher{})

	switch t := tx.TxInner.(type) {
	case dao.ProposalTx:
		if err := bc.daoProcessor.ProcessProposalTx(&t, tx.From, hash); err != nil {
			return fmt.Errorf("failed to process proposal transaction: %w", err)
		}
		bc.logger.Log("msg", "processed DAO proposal", "hash", hash, "title", t.Title)

	case dao.VoteTx:
		if err := bc.daoProcessor.ProcessVoteTx(&t, tx.From); err != nil {
			return fmt.Errorf("failed to process vote transaction: %w", err)
		}
		bc.logger.Log("msg", "processed DAO vote", "hash", hash, "proposal", t.ProposalID, "choice", t.Choice)

	case dao.DelegationTx:
		if err := bc.daoProcessor.ProcessDelegationTx(&t, tx.From); err != nil {
			return fmt.Errorf("failed to process delegation transaction: %w", err)
		}
		bc.logger.Log("msg", "processed DAO delegation", "hash", hash, "delegator", tx.From, "delegate", t.Delegate)

	case dao.TreasuryTx:
		if err := bc.daoProcessor.ProcessTreasuryTx(&t, hash); err != nil {
			return fmt.Errorf("failed to process treasury transaction: %w", err)
		}
		bc.logger.Log("msg", "processed DAO treasury", "hash", hash, "recipient", t.Recipient, "amount", t.Amount)

	case dao.TokenMintTx:
		if err := bc.daoProcessor.ProcessTokenMintTx(&t, tx.From); err != nil {
			return fmt.Errorf("failed to process token mint transaction: %w", err)
		}
		bc.logger.Log("msg", "processed DAO token mint", "hash", hash, "recipient", t.Recipient, "amount", t.Amount)

	case dao.TokenBurnTx:
		if err := bc.daoProcessor.ProcessTokenBurnTx(&t, tx.From); err != nil {
			return fmt.Errorf("failed to process token burn transaction: %w", err)
		}
		bc.logger.Log("msg", "processed DAO token burn", "hash", hash, "burner", tx.From, "amount", t.Amount)

	case dao.TokenTransferTx:
		if err := bc.daoProcessor.ProcessTokenTransferTx(&t, tx.From); err != nil {
			return fmt.Errorf("failed to process token transfer transaction: %w", err)
		}
		bc.logger.Log("msg", "processed DAO token transfer", "hash", hash, "from", tx.From, "to", t.Recipient, "amount", t.Amount)

	case dao.TokenApproveTx:
		if err := bc.daoProcessor.ProcessTokenApproveTx(&t, tx.From); err != nil {
			return fmt.Errorf("failed to process token approve transaction: %w", err)
		}
		bc.logger.Log("msg", "processed DAO token approve", "hash", hash, "owner", tx.From, "spender", t.Spender, "amount", t.Amount)

	case dao.TokenTransferFromTx:
		if err := bc.daoProcessor.ProcessTokenTransferFromTx(&t, tx.From); err != nil {
			return fmt.Errorf("failed to process token transferFrom transaction: %w", err)
		}
		bc.logger.Log("msg", "processed DAO token transferFrom", "hash", hash, "spender", tx.From, "from", t.From, "to", t.Recipient, "amount", t.Amount)

	default:
		return fmt.Errorf("unsupported DAO transaction type %T", t)
	}

	return nil
}

func (bc *Blockchain) GetBlockByHash(hash types.Hash) (*Block, error) {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	block, ok := bc.blockStore[hash]
	if !ok {
		return nil, fmt.Errorf("block with hash (%s) not found", hash)
	}

	return block, nil
}

func (bc *Blockchain) GetBlock(height uint32) (*Block, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("given height (%d) too high", height)
	}

	bc.lock.Lock()
	defer bc.lock.Unlock()

	return bc.blocks[height], nil
}

func (bc *Blockchain) GetHeader(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("given height (%d) too high", height)
	}

	bc.lock.Lock()
	defer bc.lock.Unlock()

	return bc.headers[height], nil
}

func (bc *Blockchain) GetTxByHash(hash types.Hash) (*Transaction, error) {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	tx, ok := bc.txStore[hash]
	if !ok {
		return nil, fmt.Errorf("could not find tx with hash (%s)", hash)
	}

	return tx, nil
}

func (bc *Blockchain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

// [0, 1, 2 ,3] => 4 len
// [0, 1, 2 ,3] => 3 height
func (bc *Blockchain) Height() uint32 {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return uint32(len(bc.headers) - 1)
}

func (bc *Blockchain) handleTransaction(tx *Transaction) error {
	// If we have data inside execute that data on the VM.
	if len(tx.Data) > 0 {
		bc.logger.Log("msg", "executing code", "len", len(tx.Data), "hash", tx.Hash(&TxHasher{}))

		vm := NewVM(tx.Data, bc.contractState)
		if err := vm.Run(); err != nil {
			return err
		}
	}

	// If the txInner of the transaction is not nil we need to handle
	// the native NFT implementation or DAO transactions.
	if tx.TxInner != nil {
		// Check if it's a DAO transaction
		if bc.isDAOTransaction(tx.TxInner) {
			if err := bc.handleDAOTransaction(tx); err != nil {
				return err
			}
		} else {
			// Handle native NFT transactions
			if err := bc.handleNativeNFT(tx); err != nil {
				return err
			}
		}
	}

	// Handle the native transaction here
	if tx.Value > 0 {
		if err := bc.handleNativeTransfer(tx); err != nil {
			return err
		}
	}

	return nil
}

// isDAOTransaction checks if a transaction inner type is a DAO transaction
func (bc *Blockchain) isDAOTransaction(txInner any) bool {
	switch txInner.(type) {
	case dao.ProposalTx, dao.VoteTx, dao.DelegationTx, dao.TreasuryTx,
		dao.TokenMintTx, dao.TokenBurnTx, dao.TokenTransferTx,
		dao.TokenApproveTx, dao.TokenTransferFromTx:
		return true
	default:
		return false
	}
}

func (bc *Blockchain) addBlockWithoutValidation(b *Block) error {
	bc.stateLock.Lock()
	for i := 0; i < len(b.Transactions); i++ {
		if err := bc.handleTransaction(b.Transactions[i]); err != nil {
			bc.logger.Log("error", err.Error())

			b.Transactions[i] = b.Transactions[len(b.Transactions)-1]
			b.Transactions = b.Transactions[:len(b.Transactions)-1]

			continue
		}
	}
	bc.stateLock.Unlock()

	// fmt.Println("========ACCOUNT STATE==============")
	// fmt.Printf("%+v\n", bc.accountState.accounts)
	// fmt.Println("========ACCOUNT STATE==============")

	bc.lock.Lock()
	bc.headers = append(bc.headers, b.Header)
	bc.blocks = append(bc.blocks, b)
	bc.blockStore[b.Hash(BlockHasher{})] = b

	for _, tx := range b.Transactions {
		bc.txStore[tx.Hash(TxHasher{})] = tx
	}
	bc.lock.Unlock()

	bc.logger.Log(
		"msg", "new block",
		"hash", b.Hash(BlockHasher{}),
		"height", b.Height,
		"transactions", len(b.Transactions),
	)

	return bc.store.Put(b)
}

// GetDAOState returns the current DAO governance state
func (bc *Blockchain) GetDAOState() *dao.GovernanceState {
	bc.stateLock.RLock()
	defer bc.stateLock.RUnlock()
	return bc.daoState
}

// GetDAOTokenState returns the current DAO token state
func (bc *Blockchain) GetDAOTokenState() *dao.GovernanceToken {
	bc.stateLock.RLock()
	defer bc.stateLock.RUnlock()
	return bc.daoTokenState
}

// GetDAOProcessor returns the DAO transaction processor
func (bc *Blockchain) GetDAOProcessor() *dao.DAOProcessor {
	return bc.daoProcessor
}

// GetProposal returns a specific proposal by ID
func (bc *Blockchain) GetProposal(proposalID types.Hash) (*dao.Proposal, error) {
	bc.stateLock.RLock()
	defer bc.stateLock.RUnlock()

	proposal, exists := bc.daoState.Proposals[proposalID]
	if !exists {
		return nil, fmt.Errorf("proposal with ID (%s) not found", proposalID)
	}

	return proposal, nil
}

// GetProposals returns all proposals
func (bc *Blockchain) GetProposals() map[types.Hash]*dao.Proposal {
	bc.stateLock.RLock()
	defer bc.stateLock.RUnlock()

	// Return a copy to prevent external modification
	proposals := make(map[types.Hash]*dao.Proposal)
	for id, proposal := range bc.daoState.Proposals {
		proposals[id] = proposal
	}

	return proposals
}

// GetVotes returns all votes for a specific proposal
func (bc *Blockchain) GetVotes(proposalID types.Hash) (map[string]*dao.Vote, error) {
	bc.stateLock.RLock()
	defer bc.stateLock.RUnlock()

	votes, exists := bc.daoState.Votes[proposalID]
	if !exists {
		return nil, fmt.Errorf("no votes found for proposal (%s)", proposalID)
	}

	// Return a copy to prevent external modification
	votesCopy := make(map[string]*dao.Vote)
	for voter, vote := range votes {
		votesCopy[voter] = vote
	}

	return votesCopy, nil
}

// GetTokenBalance returns the governance token balance for an address
func (bc *Blockchain) GetTokenBalance(address crypto.PublicKey) uint64 {
	bc.stateLock.RLock()
	defer bc.stateLock.RUnlock()

	return bc.daoTokenState.GetBalance(address.String())
}

// GetTreasuryState returns the current treasury state
func (bc *Blockchain) GetTreasuryState() *dao.TreasuryState {
	bc.stateLock.RLock()
	defer bc.stateLock.RUnlock()
	return bc.daoState.Treasury
}

// GetDelegation returns the delegation for a specific delegator
func (bc *Blockchain) GetDelegation(delegator crypto.PublicKey) (*dao.Delegation, error) {
	bc.stateLock.RLock()
	defer bc.stateLock.RUnlock()

	delegation, exists := bc.daoState.Delegations[delegator.String()]
	if !exists {
		return nil, fmt.Errorf("no delegation found for address (%s)", delegator.String())
	}

	return delegation, nil
}

// UpdateProposalStatuses updates the status of all active proposals based on current time
func (bc *Blockchain) UpdateProposalStatuses() error {
	bc.stateLock.Lock()
	defer bc.stateLock.Unlock()

	for proposalID := range bc.daoState.Proposals {
		if err := bc.daoProcessor.UpdateProposalStatus(proposalID); err != nil {
			bc.logger.Log("error", "failed to update proposal status", "proposal", proposalID, "err", err)
		}
	}

	return nil
}
