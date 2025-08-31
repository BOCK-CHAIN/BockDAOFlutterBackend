# ProjectX VM Governance Instructions

This document describes the governance instructions implemented in the ProjectX Virtual Machine (VM) for decentralized autonomous organization (DAO) operations.

## Overview

The ProjectX VM has been extended with 12 governance-specific instructions that enable on-chain DAO operations including proposal creation, voting, delegation, treasury management, and token operations.

## Instruction Set

### Core Governance Instructions

| Instruction | Opcode | Description |
|-------------|--------|-------------|
| `InstrCreateProposal` | 0x20 (32) | Create a new governance proposal |
| `InstrCastVote` | 0x21 (33) | Cast a vote on an existing proposal |
| `InstrDelegate` | 0x22 (34) | Delegate voting power to another address |
| `InstrCalculateQuorum` | 0x23 (35) | Calculate if quorum is met for a proposal |
| `InstrExecuteProposal` | 0x24 (36) | Execute a passed proposal |
| `InstrQuadraticVote` | 0x25 (37) | Cast a quadratic vote |
| `InstrTreasuryTransfer` | 0x26 (38) | Execute treasury fund transfer |
| `InstrMintTokens` | 0x27 (39) | Mint governance tokens |
| `InstrBurnTokens` | 0x28 (40) | Burn governance tokens |
| `InstrGetProposal` | 0x29 (41) | Retrieve proposal information |
| `InstrGetVote` | 0x2a (42) | Retrieve vote information |
| `InstrGetDelegation` | 0x2b (43) | Retrieve delegation information |

## Instruction Details

### InstrCreateProposal (0x20)

Creates a new governance proposal.

**Stack Input (bottom to top):**
- `title` (string) - Proposal title
- `description` (string) - Proposal description
- `proposalType` (ProposalType) - Type of proposal
- `votingType` (VotingType) - Voting mechanism to use
- `startTime` (int64) - Voting start timestamp
- `endTime` (int64) - Voting end timestamp
- `threshold` (uint64) - Voting threshold
- `metadataHash` ([]byte) - IPFS hash for additional metadata

**Stack Output:**
- `proposalID` ([]byte) - Unique proposal identifier

**Errors:**
- `ErrInvalidTimeframeError` - Start time >= end time

### InstrCastVote (0x21)

Cast a vote on an existing proposal.

**Stack Input:**
- `proposalID` ([]byte) - Proposal identifier
- `choice` (VoteChoice) - Vote choice (Yes/No/Abstain)
- `weight` (uint64) - Vote weight
- `reason` (string) - Voting reason

**Stack Output:**
- `success` (bool) - Vote success status

**Errors:**
- `ErrProposalNotFoundError` - Proposal doesn't exist
- `ErrVotingNotStarted` - Voting period hasn't started
- `ErrVotingPeriodClosed` - Voting period has ended
- `ErrDuplicateVoteError` - User already voted
- `ErrInvalidVoteChoiceError` - Invalid vote choice

### InstrDelegate (0x22)

Delegate voting power to another address.

**Stack Input:**
- `delegate` ([]byte|PublicKey) - Delegate address
- `duration` (int64) - Delegation duration in seconds
- `revoke` (bool) - Whether to revoke existing delegation

**Stack Output:**
- `success` (bool) - Delegation success status

### InstrCalculateQuorum (0x23)

Calculate if quorum is met for a proposal.

**Stack Input:**
- `proposalID` ([]byte) - Proposal identifier

**Stack Output:**
- `quorumMet` (bool) - Whether quorum threshold is met

**Errors:**
- `ErrProposalNotFoundError` - Proposal doesn't exist

### InstrExecuteProposal (0x24)

Execute a passed proposal.

**Stack Input:**
- `proposalID` ([]byte) - Proposal identifier

**Stack Output:**
- `success` (bool) - Execution success status

**Errors:**
- `ErrProposalNotFoundError` - Proposal doesn't exist
- `ErrQuorumNotMetError` - Quorum not met

### InstrQuadraticVote (0x25)

Cast a quadratic vote where cost increases quadratically.

**Stack Input:**
- `proposalID` ([]byte) - Proposal identifier
- `choice` (VoteChoice) - Vote choice
- `voteCount` (uint64) - Number of votes (cost = voteCount²)
- `reason` (string) - Voting reason

**Stack Output:**
- `tokenCost` (uint64) - Token cost for the vote
- `success` (bool) - Vote success status

**Errors:**
- `ErrProposalNotFoundError` - Proposal doesn't exist
- `ErrVotingNotStarted` - Voting period hasn't started
- `ErrVotingPeriodClosed` - Voting period has ended
- `ErrDuplicateVoteError` - User already voted

### InstrTreasuryTransfer (0x26)

Execute a treasury fund transfer with multi-signature validation.

**Stack Input:**
- `recipient` ([]byte|PublicKey) - Recipient address
- `amount` (uint64) - Transfer amount
- `purpose` (string) - Transfer purpose
- `signatures` ([]byte) - JSON-encoded signatures
- `requiredSigs` (uint8) - Required signature count

**Stack Output:**
- `transactionID` ([]byte) - Transaction identifier

**Errors:**
- `ErrTreasuryInsufficientFunds` - Insufficient treasury balance
- `ErrInvalidSignature` - Invalid or insufficient signatures

### InstrMintTokens (0x27)

Mint governance tokens to an address.

**Stack Input:**
- `recipient` ([]byte) - Recipient address
- `amount` (uint64) - Amount to mint
- `reason` (string) - Minting reason

**Stack Output:**
- `success` (bool) - Minting success status

### InstrBurnTokens (0x28)

Burn governance tokens from caller's balance.

**Stack Input:**
- `amount` (uint64) - Amount to burn
- `reason` (string) - Burning reason

**Stack Output:**
- `success` (bool) - Burning success status

**Errors:**
- `ErrInsufficientTokensForVote` - Insufficient token balance

### InstrGetProposal (0x29)

Retrieve proposal information.

**Stack Input:**
- `proposalID` ([]byte) - Proposal identifier

**Stack Output:**
- `proposalData` ([]byte) - JSON-encoded proposal data (nil if not found)

### InstrGetVote (0x2a)

Retrieve vote information for a specific voter and proposal.

**Stack Input:**
- `proposalID` ([]byte) - Proposal identifier
- `voter` ([]byte) - Voter address

**Stack Output:**
- `voteData` ([]byte) - JSON-encoded vote data (nil if not found)

### InstrGetDelegation (0x2b)

Retrieve delegation information for a delegator.

**Stack Input:**
- `delegator` ([]byte) - Delegator address

**Stack Output:**
- `delegationData` ([]byte) - JSON-encoded delegation data (nil if not found)

## Data Types

### ProposalType
- `ProposalTypeGeneral` (0x01) - General governance
- `ProposalTypeTreasury` (0x02) - Treasury spending
- `ProposalTypeTechnical` (0x03) - Protocol changes
- `ProposalTypeParameter` (0x04) - Parameter updates

### VotingType
- `VotingTypeSimple` (0x01) - Simple majority
- `VotingTypeQuadratic` (0x02) - Quadratic voting
- `VotingTypeWeighted` (0x03) - Token-weighted
- `VotingTypeReputation` (0x04) - Reputation-based

### VoteChoice
- `VoteChoiceYes` (0x01) - Support the proposal
- `VoteChoiceNo` (0x02) - Oppose the proposal
- `VoteChoiceAbstain` (0x03) - Abstain from voting

### ProposalStatus
- `ProposalStatusPending` (0x01) - Proposal created, voting not started
- `ProposalStatusActive` (0x02) - Voting period active
- `ProposalStatusPassed` (0x03) - Proposal passed
- `ProposalStatusRejected` (0x04) - Proposal rejected
- `ProposalStatusExecuted` (0x05) - Proposal executed
- `ProposalStatusCancelled` (0x06) - Proposal cancelled

## VM State Management

The governance instructions operate on a `GovernanceState` that includes:

- **Proposals**: Map of proposal IDs to proposal data
- **Votes**: Map of proposal IDs to voter votes
- **Delegations**: Map of delegator addresses to delegation data
- **TokenHolders**: Map of addresses to token holder information
- **Treasury**: Multi-signature treasury state
- **Config**: DAO configuration parameters

## Default Configuration

- **MinProposalThreshold**: 1000 tokens
- **VotingPeriod**: 86400 seconds (24 hours)
- **QuorumThreshold**: 2000 votes
- **PassingThreshold**: 5100 basis points (51%)
- **TreasuryThreshold**: 5000 tokens

## Security Features

1. **Time-based validation**: Proposals have start/end times
2. **Duplicate vote prevention**: Users can only vote once per proposal
3. **Multi-signature treasury**: Requires multiple signatures for fund transfers
4. **Type safety**: Strong typing for all governance operations
5. **Error handling**: Comprehensive error codes and messages

## Usage Example

See `vm_governance_example.go` for a complete example of using all governance instructions.

## Testing

The governance instructions are thoroughly tested with:
- Unit tests for each instruction
- Error handling tests
- Integration tests for complete workflows
- Edge case validation

Run tests with:
```bash
go test -v ./core -run TestVMGovernance
```

## Integration

The governance VM instructions integrate with:
- ProjectX blockchain transaction system
- JSON-RPC API server
- IPFS for metadata storage
- Multi-platform frontend applications

This implementation provides a complete foundation for decentralized governance operations within the ProjectX ecosystem.