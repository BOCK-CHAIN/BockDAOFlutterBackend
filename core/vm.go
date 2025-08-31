package core

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/dao"
	"github.com/BOCK-CHAIN/BockChain/types"
)

type Instruction byte

const (
	InstrPushInt  Instruction = 0x0a // 10
	InstrAdd      Instruction = 0x0b // 11
	InstrPushByte Instruction = 0x0c
	InstrPack     Instruction = 0x0d
	InstrSub      Instruction = 0x0e
	InstrStore    Instruction = 0x0f

	// Governance instructions
	InstrCreateProposal   Instruction = 0x20 // 32
	InstrCastVote         Instruction = 0x21 // 33
	InstrDelegate         Instruction = 0x22 // 34
	InstrCalculateQuorum  Instruction = 0x23 // 35
	InstrExecuteProposal  Instruction = 0x24 // 36
	InstrQuadraticVote    Instruction = 0x25 // 37
	InstrTreasuryTransfer Instruction = 0x26 // 38
	InstrMintTokens       Instruction = 0x27 // 39
	InstrBurnTokens       Instruction = 0x28 // 40
	InstrGetProposal      Instruction = 0x29 // 41
	InstrGetVote          Instruction = 0x2a // 42
	InstrGetDelegation    Instruction = 0x2b // 43
)

type Stack struct {
	data []any
	sp   int
}

func NewStack(size int) *Stack {
	return &Stack{
		data: make([]any, size),
		sp:   0,
	}
}

func (s *Stack) Push(v any) {
	s.data[s.sp] = v
	s.sp++
}

func (s *Stack) Pop() any {
	value := s.data[0]
	s.data = append(s.data[:0], s.data[1:]...)
	s.sp--

	return value
}

type VM struct {
	data            []byte
	ip              int // instruction pointer
	stack           *Stack
	contractState   *State
	governanceState *dao.GovernanceState
	caller          crypto.PublicKey
	timestamp       int64
}

func NewVM(data []byte, contractState *State) *VM {
	return &VM{
		contractState:   contractState,
		data:            data,
		ip:              0,
		stack:           NewStack(128),
		governanceState: dao.NewGovernanceState(),
		timestamp:       time.Now().Unix(),
	}
}

func NewVMWithGovernance(data []byte, contractState *State, governanceState *dao.GovernanceState, caller crypto.PublicKey) *VM {
	return &VM{
		contractState:   contractState,
		governanceState: governanceState,
		data:            data,
		ip:              0,
		stack:           NewStack(128),
		caller:          caller,
		timestamp:       time.Now().Unix(),
	}
}

func NewVMWithGovernanceAndTimestamp(data []byte, contractState *State, governanceState *dao.GovernanceState, caller crypto.PublicKey, timestamp int64) *VM {
	return &VM{
		contractState:   contractState,
		governanceState: governanceState,
		data:            data,
		ip:              0,
		stack:           NewStack(128),
		caller:          caller,
		timestamp:       timestamp,
	}
}

func (vm *VM) Run() error {
	for {
		instr := Instruction(vm.data[vm.ip])

		if err := vm.Exec(instr); err != nil {
			return err
		}

		vm.ip++

		if vm.ip > len(vm.data)-1 {
			break
		}
	}

	return nil
}

func (vm *VM) Exec(instr Instruction) error {
	switch instr {
	case InstrStore:
		var (
			key             = vm.stack.Pop().([]byte)
			value           = vm.stack.Pop()
			serializedValue []byte
		)

		switch v := value.(type) {
		case int:
			serializedValue = serializeInt64(int64(v))
		default:
			panic("TODO: unknown type")
		}

		vm.contractState.Put(key, serializedValue)

	case InstrPushInt:
		vm.stack.Push(int(vm.data[vm.ip-1]))

	case InstrPushByte:
		vm.stack.Push(byte(vm.data[vm.ip-1]))

	case InstrPack:
		n := vm.stack.Pop().(int)
		b := make([]byte, n)

		for i := 0; i < n; i++ {
			b[i] = vm.stack.Pop().(byte)
		}

		vm.stack.Push(b)

	case InstrSub:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a - b
		vm.stack.Push(c)

	case InstrAdd:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a + b
		vm.stack.Push(c)

	// Governance instructions
	case InstrCreateProposal:
		return vm.execCreateProposal()
	case InstrCastVote:
		return vm.execCastVote()
	case InstrDelegate:
		return vm.execDelegate()
	case InstrCalculateQuorum:
		return vm.execCalculateQuorum()
	case InstrExecuteProposal:
		return vm.execExecuteProposal()
	case InstrQuadraticVote:
		return vm.execQuadraticVote()
	case InstrTreasuryTransfer:
		return vm.execTreasuryTransfer()
	case InstrMintTokens:
		return vm.execMintTokens()
	case InstrBurnTokens:
		return vm.execBurnTokens()
	case InstrGetProposal:
		return vm.execGetProposal()
	case InstrGetVote:
		return vm.execGetVote()
	case InstrGetDelegation:
		return vm.execGetDelegation()
	}

	return nil
}

func serializeInt64(value int64) []byte {
	buf := make([]byte, 8)

	binary.LittleEndian.PutUint64(buf, uint64(value))

	return buf
}

func deserializeInt64(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}

// Governance instruction handlers

// execCreateProposal handles proposal creation
func (vm *VM) execCreateProposal() error {
	// Stack: [title, description, proposalType, votingType, startTime, endTime, threshold, metadataHash]
	title := vm.stack.Pop().(string)
	description := vm.stack.Pop().(string)
	proposalType := vm.stack.Pop().(dao.ProposalType)
	votingType := vm.stack.Pop().(dao.VotingType)
	startTime := vm.stack.Pop().(int64)
	endTime := vm.stack.Pop().(int64)
	threshold := vm.stack.Pop().(uint64)
	metadataHashBytes := vm.stack.Pop().([]byte)

	// Validate timeframe
	if startTime >= endTime {
		return dao.ErrInvalidTimeframeError
	}

	// Create proposal ID
	proposalID := types.Hash{}
	copy(proposalID[:], fmt.Sprintf("%s-%d", title, vm.timestamp))

	// Create metadata hash
	var metadataHash types.Hash
	copy(metadataHash[:], metadataHashBytes)

	// Create proposal
	proposal := &dao.Proposal{
		ID:           proposalID,
		Creator:      vm.caller,
		Title:        title,
		Description:  description,
		ProposalType: proposalType,
		VotingType:   votingType,
		StartTime:    startTime,
		EndTime:      endTime,
		Status:       dao.ProposalStatusPending,
		Threshold:    threshold,
		Results:      &dao.VoteResults{},
		MetadataHash: metadataHash,
	}

	// Store proposal in governance state
	vm.governanceState.Proposals[proposalID] = proposal

	// Initialize vote map for this proposal
	vm.governanceState.Votes[proposalID] = make(map[string]*dao.Vote)

	// Push proposal ID to stack as result
	vm.stack.Push(proposalID[:])

	return nil
}

// execCastVote handles vote casting
func (vm *VM) execCastVote() error {
	// Stack: [proposalID, choice, weight, reason]
	proposalIDBytes := vm.stack.Pop().([]byte)
	choice := vm.stack.Pop().(dao.VoteChoice)
	weight := vm.stack.Pop().(uint64)
	reason := vm.stack.Pop().(string)

	var proposalID types.Hash
	copy(proposalID[:], proposalIDBytes)

	// Check if proposal exists
	proposal, exists := vm.governanceState.Proposals[proposalID]
	if !exists {
		return dao.ErrProposalNotFoundError
	}

	// Check if voting is active
	if vm.timestamp < proposal.StartTime {
		return dao.ErrVotingNotStarted
	}
	if vm.timestamp > proposal.EndTime {
		return dao.ErrVotingPeriodClosed
	}

	// Check for duplicate vote
	voterKey := string(vm.caller[:])
	if _, hasVoted := vm.governanceState.Votes[proposalID][voterKey]; hasVoted {
		return dao.ErrDuplicateVoteError
	}

	// Validate vote choice
	if choice != dao.VoteChoiceYes && choice != dao.VoteChoiceNo && choice != dao.VoteChoiceAbstain {
		return dao.ErrInvalidVoteChoiceError
	}

	// Create vote
	vote := &dao.Vote{
		Voter:     vm.caller,
		Choice:    choice,
		Weight:    weight,
		Timestamp: vm.timestamp,
		Reason:    reason,
	}

	// Store vote
	vm.governanceState.Votes[proposalID][voterKey] = vote

	// Update proposal results
	switch choice {
	case dao.VoteChoiceYes:
		proposal.Results.YesVotes += weight
	case dao.VoteChoiceNo:
		proposal.Results.NoVotes += weight
	case dao.VoteChoiceAbstain:
		proposal.Results.AbstainVotes += weight
	}
	proposal.Results.TotalVoters++

	// Push success result
	vm.stack.Push(true)

	return nil
}

// execDelegate handles voting power delegation
func (vm *VM) execDelegate() error {
	// Stack: [delegate, duration, revoke]
	delegateData := vm.stack.Pop()
	duration := vm.stack.Pop().(int64)
	revoke := vm.stack.Pop().(bool)

	var delegate crypto.PublicKey
	switch v := delegateData.(type) {
	case []byte:
		copy(delegate[:], v)
	case crypto.PublicKey:
		delegate = v
	default:
		return dao.NewDAOError(dao.ErrInvalidDelegation, "invalid delegate format", nil)
	}

	delegatorKey := string(vm.caller[:])

	if revoke {
		// Revoke existing delegation
		if delegation, exists := vm.governanceState.Delegations[delegatorKey]; exists {
			delegation.Active = false
		}
	} else {
		// Create new delegation
		delegation := &dao.Delegation{
			Delegator: vm.caller,
			Delegate:  delegate,
			StartTime: vm.timestamp,
			EndTime:   vm.timestamp + duration,
			Active:    true,
		}

		vm.governanceState.Delegations[delegatorKey] = delegation
	}

	// Push success result
	vm.stack.Push(true)

	return nil
}

// execCalculateQuorum calculates if quorum is met for a proposal
func (vm *VM) execCalculateQuorum() error {
	// Stack: [proposalID]
	proposalIDBytes := vm.stack.Pop().([]byte)

	var proposalID types.Hash
	copy(proposalID[:], proposalIDBytes)

	// Check if proposal exists
	proposal, exists := vm.governanceState.Proposals[proposalID]
	if !exists {
		return dao.ErrProposalNotFoundError
	}

	// Calculate total participation
	totalVotes := proposal.Results.YesVotes + proposal.Results.NoVotes + proposal.Results.AbstainVotes

	// Check against quorum threshold
	quorumMet := totalVotes >= vm.governanceState.Config.QuorumThreshold

	// Update proposal results
	proposal.Results.Quorum = totalVotes

	// Push result
	vm.stack.Push(quorumMet)

	return nil
}

// execExecuteProposal executes a passed proposal
func (vm *VM) execExecuteProposal() error {
	// Stack: [proposalID]
	proposalIDBytes := vm.stack.Pop().([]byte)

	var proposalID types.Hash
	copy(proposalID[:], proposalIDBytes)

	// Check if proposal exists
	proposal, exists := vm.governanceState.Proposals[proposalID]
	if !exists {
		return dao.ErrProposalNotFoundError
	}

	// Check if proposal has passed
	totalVotes := proposal.Results.YesVotes + proposal.Results.NoVotes
	if totalVotes == 0 {
		return dao.ErrQuorumNotMetError
	}

	passingThreshold := (proposal.Results.YesVotes * 10000) / totalVotes
	if passingThreshold < vm.governanceState.Config.PassingThreshold {
		proposal.Status = dao.ProposalStatusRejected
		vm.stack.Push(false)
		return nil
	}

	// Check quorum
	if proposal.Results.Quorum < vm.governanceState.Config.QuorumThreshold {
		return dao.ErrQuorumNotMetError
	}

	// Mark as executed
	proposal.Status = dao.ProposalStatusExecuted
	proposal.Results.Passed = true

	// Push success result
	vm.stack.Push(true)

	return nil
}

// execQuadraticVote handles quadratic voting
func (vm *VM) execQuadraticVote() error {
	// Stack: [proposalID, choice, voteCount, reason]
	proposalIDBytes := vm.stack.Pop().([]byte)
	choice := vm.stack.Pop().(dao.VoteChoice)
	voteCount := vm.stack.Pop().(uint64)
	reason := vm.stack.Pop().(string)

	var proposalID types.Hash
	copy(proposalID[:], proposalIDBytes)

	// Check if proposal exists and uses quadratic voting
	proposal, exists := vm.governanceState.Proposals[proposalID]
	if !exists {
		return dao.ErrProposalNotFoundError
	}

	if proposal.VotingType != dao.VotingTypeQuadratic {
		return dao.NewDAOError(dao.ErrInvalidProposal, "proposal does not use quadratic voting", nil)
	}

	// Calculate quadratic cost (voteCount^2)
	tokenCost := voteCount * voteCount

	// Check if voting is active
	if vm.timestamp < proposal.StartTime {
		return dao.ErrVotingNotStarted
	}
	if vm.timestamp > proposal.EndTime {
		return dao.ErrVotingPeriodClosed
	}

	// Check for duplicate vote
	voterKey := string(vm.caller[:])
	if _, hasVoted := vm.governanceState.Votes[proposalID][voterKey]; hasVoted {
		return dao.ErrDuplicateVoteError
	}

	// Create vote with quadratic weight
	vote := &dao.Vote{
		Voter:     vm.caller,
		Choice:    choice,
		Weight:    voteCount, // Effective voting power
		Timestamp: vm.timestamp,
		Reason:    reason,
	}

	// Store vote
	vm.governanceState.Votes[proposalID][voterKey] = vote

	// Update proposal results with effective voting power
	switch choice {
	case dao.VoteChoiceYes:
		proposal.Results.YesVotes += voteCount
	case dao.VoteChoiceNo:
		proposal.Results.NoVotes += voteCount
	case dao.VoteChoiceAbstain:
		proposal.Results.AbstainVotes += voteCount
	}
	proposal.Results.TotalVoters++

	// Push token cost and success
	vm.stack.Push(tokenCost)
	vm.stack.Push(true)

	return nil
}

// execTreasuryTransfer handles treasury fund transfers
func (vm *VM) execTreasuryTransfer() error {
	// Stack: [recipient, amount, purpose, signatures, requiredSigs]
	recipientData := vm.stack.Pop()
	amount := vm.stack.Pop().(uint64)
	purpose := vm.stack.Pop().(string)
	signaturesBytes := vm.stack.Pop().([]byte)
	requiredSigs := vm.stack.Pop().(uint8)

	var recipient crypto.PublicKey
	switch v := recipientData.(type) {
	case []byte:
		copy(recipient[:], v)
	case crypto.PublicKey:
		recipient = v
	default:
		return dao.NewDAOError(dao.ErrInvalidSignature, "invalid recipient format", nil)
	}

	// Deserialize signatures
	var signatures []crypto.Signature
	if err := json.Unmarshal(signaturesBytes, &signatures); err != nil {
		return dao.NewDAOError(dao.ErrInvalidSignature, "failed to deserialize signatures", nil)
	}

	// Check if treasury has sufficient funds
	if vm.governanceState.Treasury.Balance < amount {
		return dao.ErrTreasuryInsufficientFunds
	}

	// Validate signatures
	if len(signatures) < int(requiredSigs) {
		return dao.NewDAOError(dao.ErrInvalidSignature, "insufficient signatures", nil)
	}

	// Create transaction ID
	txID := types.Hash{}
	copy(txID[:], fmt.Sprintf("treasury-%d-%s", vm.timestamp, purpose))

	// Create pending transaction
	pendingTx := &dao.PendingTx{
		ID:         txID,
		Recipient:  recipient,
		Amount:     amount,
		Purpose:    purpose,
		Signatures: signatures,
		CreatedAt:  vm.timestamp,
		ExpiresAt:  vm.timestamp + 86400, // 24 hours
		Executed:   false,
	}

	// Store pending transaction
	vm.governanceState.Treasury.Transactions[txID] = pendingTx

	// Execute transfer
	vm.governanceState.Treasury.Balance -= amount
	pendingTx.Executed = true

	// Push transaction ID
	vm.stack.Push(txID[:])

	return nil
}

// execMintTokens handles governance token minting
func (vm *VM) execMintTokens() error {
	// Stack: [recipient, amount, reason]
	recipientBytes := vm.stack.Pop().([]byte)
	amount := vm.stack.Pop().(uint64)
	_ = vm.stack.Pop().(string) // reason - unused for now

	recipientKey := string(recipientBytes)

	// Check if token holder exists, create if not
	if _, exists := vm.governanceState.TokenHolders[recipientKey]; !exists {
		var recipientPubKey crypto.PublicKey
		copy(recipientPubKey[:], recipientBytes)

		vm.governanceState.TokenHolders[recipientKey] = &dao.TokenHolder{
			Address:    recipientPubKey,
			Balance:    0,
			Staked:     0,
			Reputation: 0,
			JoinedAt:   vm.timestamp,
			LastActive: vm.timestamp,
		}
	}

	// Mint tokens
	vm.governanceState.TokenHolders[recipientKey].Balance += amount

	// Push success result
	vm.stack.Push(true)

	return nil
}

// execBurnTokens handles governance token burning
func (vm *VM) execBurnTokens() error {
	// Stack: [amount, reason]
	amount := vm.stack.Pop().(uint64)
	_ = vm.stack.Pop().(string) // reason - unused for now

	callerKey := string(vm.caller[:])

	// Check if caller exists and has sufficient balance
	holder, exists := vm.governanceState.TokenHolders[callerKey]
	if !exists {
		return dao.ErrInsufficientTokensForVote
	}

	if holder.Balance < amount {
		return dao.ErrInsufficientTokensForVote
	}

	// Burn tokens
	holder.Balance -= amount

	// Push success result
	vm.stack.Push(true)

	return nil
}

// execGetProposal retrieves proposal information
func (vm *VM) execGetProposal() error {
	// Stack: [proposalID]
	proposalIDBytes := vm.stack.Pop().([]byte)

	var proposalID types.Hash
	copy(proposalID[:], proposalIDBytes)

	// Check if proposal exists
	proposal, exists := vm.governanceState.Proposals[proposalID]
	if !exists {
		vm.stack.Push(nil)
		return nil
	}

	// Serialize proposal data
	proposalData, err := json.Marshal(proposal)
	if err != nil {
		return dao.NewDAOError(dao.ErrInvalidProposal, "failed to serialize proposal", nil)
	}

	// Push proposal data
	vm.stack.Push(proposalData)

	return nil
}

// execGetVote retrieves vote information
func (vm *VM) execGetVote() error {
	// Stack: [proposalID, voter]
	proposalIDBytes := vm.stack.Pop().([]byte)
	voterBytes := vm.stack.Pop().([]byte)

	var proposalID types.Hash
	copy(proposalID[:], proposalIDBytes)

	voterKey := string(voterBytes)

	// Check if vote exists
	if votes, exists := vm.governanceState.Votes[proposalID]; exists {
		if vote, hasVoted := votes[voterKey]; hasVoted {
			// Serialize vote data
			voteData, err := json.Marshal(vote)
			if err != nil {
				return dao.NewDAOError(dao.ErrInvalidVoteChoice, "failed to serialize vote", nil)
			}
			vm.stack.Push(voteData)
			return nil
		}
	}

	// No vote found
	vm.stack.Push(nil)
	return nil
}

// execGetDelegation retrieves delegation information
func (vm *VM) execGetDelegation() error {
	// Stack: [delegator]
	delegatorBytes := vm.stack.Pop().([]byte)
	delegatorKey := string(delegatorBytes)

	// Check if delegation exists
	delegation, exists := vm.governanceState.Delegations[delegatorKey]
	if !exists {
		vm.stack.Push(nil)
		return nil
	}

	// Serialize delegation data
	delegationData, err := json.Marshal(delegation)
	if err != nil {
		return dao.NewDAOError(dao.ErrInvalidDelegation, "failed to serialize delegation", nil)
	}

	// Push delegation data
	vm.stack.Push(delegationData)

	return nil
}
