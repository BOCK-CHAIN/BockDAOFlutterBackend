package dao

import (
	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// GovernanceState manages the overall state of the DAO
type GovernanceState struct {
	Proposals    map[types.Hash]*Proposal
	Votes        map[types.Hash]map[string]*Vote
	Delegations  map[string]*Delegation
	TokenHolders map[string]*TokenHolder
	Treasury     *TreasuryState
	Config       *DAOConfig
}

// NewGovernanceState creates a new governance state instance
func NewGovernanceState() *GovernanceState {
	return &GovernanceState{
		Proposals:    make(map[types.Hash]*Proposal),
		Votes:        make(map[types.Hash]map[string]*Vote),
		Delegations:  make(map[string]*Delegation),
		TokenHolders: make(map[string]*TokenHolder),
		Treasury:     NewTreasuryState(),
		Config:       NewDAOConfig(),
	}
}

// Proposal represents a governance proposal
type Proposal struct {
	ID           types.Hash
	Creator      crypto.PublicKey
	Title        string
	Description  string
	ProposalType ProposalType
	VotingType   VotingType
	StartTime    int64
	EndTime      int64
	Status       ProposalStatus
	Threshold    uint64
	Results      *VoteResults
	MetadataHash types.Hash
}

// Vote represents a cast vote
type Vote struct {
	Voter     crypto.PublicKey
	Choice    VoteChoice
	Weight    uint64
	Timestamp int64
	Reason    string
}

// Delegation represents voting power delegation
type Delegation struct {
	Delegator crypto.PublicKey
	Delegate  crypto.PublicKey
	StartTime int64
	EndTime   int64
	Active    bool
}

// TokenHolder represents a governance token holder
type TokenHolder struct {
	Address    crypto.PublicKey
	Balance    uint64
	Staked     uint64
	Reputation uint64
	JoinedAt   int64
	LastActive int64
}

// VoteResults contains the results of a proposal vote
type VoteResults struct {
	YesVotes     uint64
	NoVotes      uint64
	AbstainVotes uint64
	TotalVoters  uint64
	Quorum       uint64
	Passed       bool
}

// TreasuryState manages the DAO treasury
type TreasuryState struct {
	Balance      uint64
	Signers      []crypto.PublicKey
	RequiredSigs uint8
	Transactions map[types.Hash]*PendingTx
}

// NewTreasuryState creates a new treasury state
func NewTreasuryState() *TreasuryState {
	return &TreasuryState{
		Balance:      0,
		Signers:      make([]crypto.PublicKey, 0),
		RequiredSigs: 1,
		Transactions: make(map[types.Hash]*PendingTx),
	}
}

// PendingTx represents a pending treasury transaction
type PendingTx struct {
	ID         types.Hash
	Recipient  crypto.PublicKey
	Amount     uint64
	Purpose    string
	Signatures []crypto.Signature
	CreatedAt  int64
	ExpiresAt  int64
	Executed   bool
}

// DAOConfig contains DAO configuration parameters
type DAOConfig struct {
	MinProposalThreshold uint64 // Minimum tokens required to create proposal
	VotingPeriod         int64  // Duration of voting period in seconds
	QuorumThreshold      uint64 // Minimum participation for valid vote
	PassingThreshold     uint64 // Percentage required to pass (basis points)
	TreasuryThreshold    uint64 // Minimum tokens for treasury proposals
}

// NewDAOConfig creates default DAO configuration
func NewDAOConfig() *DAOConfig {
	return &DAOConfig{
		MinProposalThreshold: 1000,  // 1000 tokens minimum
		VotingPeriod:         86400, // 24 hours
		QuorumThreshold:      2000,  // 20% participation
		PassingThreshold:     5100,  // 51% to pass
		TreasuryThreshold:    5000,  // 5000 tokens for treasury proposals
	}
}

// GovernanceToken manages the governance token state
type GovernanceToken struct {
	Symbol      string
	Name        string
	TotalSupply uint64
	Decimals    uint8
	Balances    map[string]uint64
	Allowances  map[string]map[string]uint64
}

// NewGovernanceToken creates a new governance token
func NewGovernanceToken(symbol, name string, decimals uint8) *GovernanceToken {
	return &GovernanceToken{
		Symbol:      symbol,
		Name:        name,
		TotalSupply: 0,
		Decimals:    decimals,
		Balances:    make(map[string]uint64),
		Allowances:  make(map[string]map[string]uint64),
	}
}

// Transfer transfers tokens from one address to another
func (gt *GovernanceToken) Transfer(from, to string, amount uint64) error {
	if gt.Balances[from] < amount {
		return NewDAOError(ErrInsufficientTokens, "insufficient balance for transfer", nil)
	}

	gt.Balances[from] -= amount
	if gt.Balances[to] == 0 {
		gt.Balances[to] = amount
	} else {
		gt.Balances[to] += amount
	}

	return nil
}

// Approve approves a spender to spend tokens on behalf of the owner
func (gt *GovernanceToken) Approve(owner, spender string, amount uint64) error {
	if gt.Allowances[owner] == nil {
		gt.Allowances[owner] = make(map[string]uint64)
	}
	gt.Allowances[owner][spender] = amount
	return nil
}

// TransferFrom transfers tokens from one address to another using allowance
func (gt *GovernanceToken) TransferFrom(spender, from, to string, amount uint64) error {
	// Check allowance
	if gt.Allowances[from] == nil || gt.Allowances[from][spender] < amount {
		return NewDAOError(ErrInsufficientTokens, "insufficient allowance for transfer", nil)
	}

	// Check balance
	if gt.Balances[from] < amount {
		return NewDAOError(ErrInsufficientTokens, "insufficient balance for transfer", nil)
	}

	// Perform transfer
	gt.Balances[from] -= amount
	if gt.Balances[to] == 0 {
		gt.Balances[to] = amount
	} else {
		gt.Balances[to] += amount
	}

	// Reduce allowance
	gt.Allowances[from][spender] -= amount

	return nil
}

// GetBalance returns the balance of an address
func (gt *GovernanceToken) GetBalance(address string) uint64 {
	return gt.Balances[address]
}

// GetAllowance returns the allowance between owner and spender
func (gt *GovernanceToken) GetAllowance(owner, spender string) uint64 {
	if gt.Allowances[owner] == nil {
		return 0
	}
	return gt.Allowances[owner][spender]
}

// Mint creates new tokens and assigns them to an address
func (gt *GovernanceToken) Mint(to string, amount uint64) error {
	// Check for overflow
	if gt.TotalSupply+amount < gt.TotalSupply {
		return NewDAOError(ErrTokenTransferFailed, "token supply overflow", nil)
	}

	gt.TotalSupply += amount
	if gt.Balances[to] == 0 {
		gt.Balances[to] = amount
	} else {
		gt.Balances[to] += amount
	}

	return nil
}

// Burn destroys tokens from an address
func (gt *GovernanceToken) Burn(from string, amount uint64) error {
	if gt.Balances[from] < amount {
		return NewDAOError(ErrInsufficientTokens, "insufficient balance to burn", nil)
	}

	gt.Balances[from] -= amount
	gt.TotalSupply -= amount

	return nil
}
