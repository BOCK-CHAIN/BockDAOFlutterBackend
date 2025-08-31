package dao

import (
	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// DAOTxType represents different types of DAO transactions
type DAOTxType byte

const (
	TxTypeProposal          DAOTxType = 0x10
	TxTypeVote              DAOTxType = 0x11
	TxTypeDelegation        DAOTxType = 0x12
	TxTypeTreasury          DAOTxType = 0x13
	TxTypeTokenMint         DAOTxType = 0x14
	TxTypeTokenBurn         DAOTxType = 0x15
	TxTypeTokenDistribution DAOTxType = 0x16
	TxTypeVestingClaim      DAOTxType = 0x17
	TxTypeStake             DAOTxType = 0x18
	TxTypeParameter         DAOTxType = 0x19
	TxTypeUnstake           DAOTxType = 0x1A
	TxTypeClaimRewards      DAOTxType = 0x1B
)

// ProposalType represents different categories of proposals
type ProposalType byte

const (
	ProposalTypeGeneral   ProposalType = 0x01 // General governance
	ProposalTypeTreasury  ProposalType = 0x02 // Treasury spending
	ProposalTypeTechnical ProposalType = 0x03 // Protocol changes
	ProposalTypeParameter ProposalType = 0x04 // Parameter updates
)

// ProposalStatus represents the current state of a proposal
type ProposalStatus byte

const (
	ProposalStatusPending   ProposalStatus = 0x01
	ProposalStatusActive    ProposalStatus = 0x02
	ProposalStatusPassed    ProposalStatus = 0x03
	ProposalStatusRejected  ProposalStatus = 0x04
	ProposalStatusExecuted  ProposalStatus = 0x05
	ProposalStatusCancelled ProposalStatus = 0x06
)

// VotingType represents different voting mechanisms
type VotingType byte

const (
	VotingTypeSimple     VotingType = 0x01 // Simple majority
	VotingTypeQuadratic  VotingType = 0x02 // Quadratic voting
	VotingTypeWeighted   VotingType = 0x03 // Token-weighted
	VotingTypeReputation VotingType = 0x04 // Reputation-based
)

// VoteChoice represents the voting options
type VoteChoice byte

const (
	VoteChoiceYes     VoteChoice = 0x01
	VoteChoiceNo      VoteChoice = 0x02
	VoteChoiceAbstain VoteChoice = 0x03
)

// ProposalTx represents a governance proposal transaction
type ProposalTx struct {
	Fee          int64
	Title        string
	Description  string
	ProposalType ProposalType
	VotingType   VotingType
	StartTime    int64
	EndTime      int64
	Threshold    uint64
	MetadataHash types.Hash // IPFS hash for large content
}

// VoteTx represents a voting transaction
type VoteTx struct {
	Fee        int64
	ProposalID types.Hash
	Choice     VoteChoice
	Weight     uint64
	Reason     string
}

// DelegationTx represents a delegation transaction
type DelegationTx struct {
	Fee      int64
	Delegate crypto.PublicKey
	Duration int64
	Revoke   bool // If true, revokes existing delegation
}

// TreasuryTx represents a treasury operation transaction
type TreasuryTx struct {
	Fee          int64
	Recipient    crypto.PublicKey
	Amount       uint64
	Purpose      string
	Signatures   []crypto.Signature
	RequiredSigs uint8
}

// TokenMintTx represents a governance token minting transaction
type TokenMintTx struct {
	Fee       int64
	Recipient crypto.PublicKey
	Amount    uint64
	Reason    string
}

// TokenBurnTx represents a governance token burning transaction
type TokenBurnTx struct {
	Fee    int64
	Amount uint64
	Reason string
}

// TokenTransferTx represents a governance token transfer transaction
type TokenTransferTx struct {
	Fee       int64
	Recipient crypto.PublicKey
	Amount    uint64
}

// TokenApproveTx represents a governance token approval transaction
type TokenApproveTx struct {
	Fee     int64
	Spender crypto.PublicKey
	Amount  uint64
}

// TokenTransferFromTx represents a governance token transferFrom transaction
type TokenTransferFromTx struct {
	Fee       int64
	From      crypto.PublicKey
	Recipient crypto.PublicKey
	Amount    uint64
}

// TokenDistributionTx represents a token distribution transaction
type TokenDistributionTx struct {
	Fee         int64
	Category    DistributionCategory
	Recipients  map[string]uint64 // address -> amount
	VestingType VestingType
	CliffPeriod int64
	Duration    int64
}

// VestingClaimTx represents a vesting claim transaction
type VestingClaimTx struct {
	Fee       int64
	VestingID string
}

// StakeTx represents a staking transaction
type StakeTx struct {
	Fee      int64
	PoolID   string
	Amount   uint64
	Duration int64 // Optional lock duration
}

// UnstakeTx represents an unstaking transaction
type UnstakeTx struct {
	Fee    int64
	PoolID string
	Amount uint64
}

// ClaimRewardsTx represents a rewards claim transaction
type ClaimRewardsTx struct {
	Fee    int64
	PoolID string
}

// DistributionCategory represents different token allocation categories
type DistributionCategory byte

const (
	DistributionFounders  DistributionCategory = 0x01
	DistributionTeam      DistributionCategory = 0x02
	DistributionCommunity DistributionCategory = 0x03
	DistributionTreasury  DistributionCategory = 0x04
	DistributionEcosystem DistributionCategory = 0x05
)

// VestingType represents different vesting mechanisms
type VestingType byte

const (
	VestingTypeLinear    VestingType = 0x01 // Linear vesting over time
	VestingTypeCliff     VestingType = 0x02 // Cliff vesting (all at once after period)
	VestingTypeMilestone VestingType = 0x03 // Milestone-based vesting
	VestingTypeImmediate VestingType = 0x04 // Immediate (no vesting)
)
