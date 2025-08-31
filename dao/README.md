# ProjectX DAO Implementation

This package implements a comprehensive Decentralized Autonomous Organization (DAO) system for the ProjectX blockchain. The DAO provides governance capabilities through token-based voting, delegation, treasury management, and various governance mechanisms.

## Features

### Core Functionality
- **Governance Tokens**: ERC-20-like token system for voting rights
- **Proposal System**: Create and manage governance proposals
- **Voting Mechanisms**: Multiple voting types (simple, quadratic, weighted, reputation-based)
- **Delegation**: Delegate voting power to trusted representatives
- **Treasury Management**: Multi-signature treasury for fund management
- **Token Operations**: Mint and burn governance tokens

### Transaction Types
The DAO extends ProjectX's transaction system with new transaction types:
- `ProposalTx` (0x10): Create governance proposals
- `VoteTx` (0x11): Cast votes on proposals
- `DelegationTx` (0x12): Delegate voting power
- `TreasuryTx` (0x13): Treasury operations
- `TokenMintTx` (0x14): Mint governance tokens
- `TokenBurnTx` (0x15): Burn governance tokens

## Architecture

### Core Components

1. **DAO**: Main DAO system coordinator
2. **GovernanceState**: Manages proposals, votes, delegations, and token holders
3. **GovernanceToken**: Token state management
4. **DAOValidator**: Validates all DAO transactions
5. **DAOProcessor**: Processes DAO transactions and updates state

### Data Structures

#### Proposal
```go
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
```

#### Vote
```go
type Vote struct {
    Voter     crypto.PublicKey
    Choice    VoteChoice
    Weight    uint64
    Timestamp int64
    Reason    string
}
```

#### Delegation
```go
type Delegation struct {
    Delegator crypto.PublicKey
    Delegate  crypto.PublicKey
    StartTime int64
    EndTime   int64
    Active    bool
}
```

## Usage Examples

### Creating a DAO
```go
// Create new DAO
dao := NewDAO("GOV", "Governance Token", 18)

// Initialize token distribution
distributions := map[string]uint64{
    "address1": 10000,
    "address2": 5000,
}
dao.InitialTokenDistribution(distributions)

// Initialize treasury
signers := []crypto.PublicKey{signer1, signer2}
dao.InitializeTreasury(signers, 2) // 2-of-2 multisig
```

### Creating Proposals
```go
proposalTx := &ProposalTx{
    Fee:          100,
    Title:        "Protocol Upgrade",
    Description:  "Upgrade to version 2.0",
    ProposalType: ProposalTypeTechnical,
    VotingType:   VotingTypeSimple,
    StartTime:    time.Now().Unix(),
    EndTime:      time.Now().Unix() + 86400,
    Threshold:    5100, // 51%
    MetadataHash: ipfsHash,
}

err := dao.Processor.ProcessProposalTx(proposalTx, creator, txHash)
```

### Voting
```go
voteTx := &VoteTx{
    Fee:        50,
    ProposalID: proposalHash,
    Choice:     VoteChoiceYes,
    Weight:     1000,
    Reason:     "I support this proposal",
}

err := dao.Processor.ProcessVoteTx(voteTx, voter)
```

### Delegation
```go
delegationTx := &DelegationTx{
    Fee:      25,
    Delegate: delegateAddress,
    Duration: 86400 * 7, // 7 days
}

err := dao.Processor.ProcessDelegationTx(delegationTx, delegator)
```

## Voting Mechanisms

### Simple Voting
- One token = one vote
- Majority wins
- Cost = vote weight

### Quadratic Voting
- Cost = (vote weight)²
- Prevents plutocracy
- Encourages broader participation

### Weighted Voting
- Vote weight proportional to token balance
- Traditional token-weighted governance

### Reputation-Based Voting
- Vote weight based on reputation score
- Rewards active participation
- Technical expertise consideration

## Security Features

### Validation
- Token balance verification
- Proposal format validation
- Timeframe validation
- Signature verification
- Double-voting prevention

### Error Handling
- Comprehensive error types
- Detailed error messages
- Graceful failure handling

### Access Control
- Role-based permissions
- Multi-signature requirements
- Emergency pause mechanisms

## Configuration

### DAO Parameters
```go
type DAOConfig struct {
    MinProposalThreshold uint64 // Minimum tokens for proposals
    VotingPeriod         int64  // Voting duration
    QuorumThreshold      uint64 // Minimum participation
    PassingThreshold     uint64 // Percentage to pass
    TreasuryThreshold    uint64 // Minimum for treasury proposals
}
```

## Integration with ProjectX

The DAO system integrates seamlessly with ProjectX's existing infrastructure:

1. **Transaction System**: Extends existing transaction types
2. **Blockchain**: Stores all governance data on-chain
3. **VM**: Can be extended with governance instructions
4. **API**: Governance endpoints for external access

## Testing

Run the test suite:
```bash
go test ./dao -v
```

The test suite covers:
- DAO creation and initialization
- Token distribution
- Proposal creation and validation
- Voting mechanisms
- Delegation functionality
- Token minting and burning
- Error handling and edge cases

## Error Codes

| Code | Error | Description |
|------|-------|-------------|
| 4001 | ErrInsufficientTokens | Not enough tokens for operation |
| 4002 | ErrProposalNotFound | Proposal doesn't exist |
| 4003 | ErrVotingClosed | Voting period ended |
| 4004 | ErrUnauthorized | Unauthorized access |
| 4005 | ErrInvalidSignature | Invalid signature |
| 4006 | ErrQuorumNotMet | Insufficient participation |
| 4007 | ErrTreasuryInsufficient | Not enough treasury funds |
| 4008 | ErrInvalidProposal | Invalid proposal format |
| 4009 | ErrDuplicateVote | User already voted |
| 4010 | ErrInvalidDelegation | Invalid delegation parameters |

## Future Enhancements

- **IPFS Integration**: Store large proposal documents off-chain
- **Layer-2 Scaling**: Off-chain computation with on-chain verification
- **Advanced Voting**: Ranked choice, approval voting
- **Reputation System**: Dynamic reputation based on participation
- **Governance Analytics**: Participation metrics and insights
- **Mobile/Web Interfaces**: User-friendly governance interfaces

## License

This implementation is part of the ProjectX blockchain project.