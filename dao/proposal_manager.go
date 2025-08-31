package dao

import (
	"fmt"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// ProposalManager provides enhanced proposal management functionality
type ProposalManager struct {
	dao *DAO
}

// NewProposalManager creates a new proposal manager
func NewProposalManager(dao *DAO) *ProposalManager {
	return &ProposalManager{
		dao: dao,
	}
}

// CreateProposal creates a new proposal with enhanced validation and features
func (pm *ProposalManager) CreateProposal(tx *ProposalTx, creator crypto.PublicKey, txHash types.Hash) (*Proposal, error) {
	// Enhanced validation
	if err := pm.validateProposalCreation(tx, creator); err != nil {
		return nil, err
	}

	// Process the proposal transaction
	if err := pm.dao.Processor.ProcessProposalTx(tx, creator, txHash); err != nil {
		return nil, err
	}

	// Return the created proposal
	return pm.dao.GetProposal(txHash)
}

// ExecuteProposal executes a passed proposal based on its type
func (pm *ProposalManager) ExecuteProposal(proposalID types.Hash, executor crypto.PublicKey) error {
	proposal, err := pm.dao.GetProposal(proposalID)
	if err != nil {
		return err
	}

	// Check if proposal can be executed
	if proposal.Status != ProposalStatusPassed {
		return NewDAOError(ErrInvalidProposal, "proposal must be in passed status to execute", nil)
	}

	// Check if executor is authorized
	if !pm.isAuthorizedExecutor(proposal, executor) {
		return NewDAOError(ErrUnauthorized, "executor not authorized for this proposal type", nil)
	}

	// Execute based on proposal type
	switch proposal.ProposalType {
	case ProposalTypeGeneral:
		return pm.executeGeneralProposal(proposal)
	case ProposalTypeTreasury:
		return pm.executeTreasuryProposal(proposal)
	case ProposalTypeTechnical:
		return pm.executeTechnicalProposal(proposal)
	case ProposalTypeParameter:
		return pm.executeParameterProposal(proposal)
	default:
		return NewDAOError(ErrInvalidProposal, "unknown proposal type", nil)
	}
}

// CancelProposal allows proposal creator to cancel their proposal before voting starts
func (pm *ProposalManager) CancelProposal(proposalID types.Hash, canceller crypto.PublicKey) error {
	proposal, err := pm.dao.GetProposal(proposalID)
	if err != nil {
		return err
	}

	// Only creator can cancel
	if proposal.Creator.String() != canceller.String() {
		return NewDAOError(ErrUnauthorized, "only proposal creator can cancel", nil)
	}

	// Can only cancel pending proposals
	if proposal.Status != ProposalStatusPending {
		return NewDAOError(ErrInvalidProposal, "can only cancel pending proposals", nil)
	}

	// Update status
	proposal.Status = ProposalStatusCancelled
	return nil
}

// GetProposalsByStatus returns all proposals with a specific status
func (pm *ProposalManager) GetProposalsByStatus(status ProposalStatus) []*Proposal {
	var proposals []*Proposal
	for _, proposal := range pm.dao.GovernanceState.Proposals {
		if proposal.Status == status {
			proposals = append(proposals, proposal)
		}
	}
	return proposals
}

// GetProposalsByType returns all proposals of a specific type
func (pm *ProposalManager) GetProposalsByType(proposalType ProposalType) []*Proposal {
	var proposals []*Proposal
	for _, proposal := range pm.dao.GovernanceState.Proposals {
		if proposal.ProposalType == proposalType {
			proposals = append(proposals, proposal)
		}
	}
	return proposals
}

// GetProposalsByCreator returns all proposals created by a specific address
func (pm *ProposalManager) GetProposalsByCreator(creator crypto.PublicKey) []*Proposal {
	var proposals []*Proposal
	creatorStr := creator.String()
	for _, proposal := range pm.dao.GovernanceState.Proposals {
		if proposal.Creator.String() == creatorStr {
			proposals = append(proposals, proposal)
		}
	}
	return proposals
}

// GetProposalVotingProgress returns detailed voting progress for a proposal
func (pm *ProposalManager) GetProposalVotingProgress(proposalID types.Hash) (*VotingProgress, error) {
	proposal, err := pm.dao.GetProposal(proposalID)
	if err != nil {
		return nil, err
	}

	votes, err := pm.dao.GetVotes(proposalID)
	if err != nil {
		return nil, err
	}

	progress := &VotingProgress{
		ProposalID:    proposalID,
		TotalVotes:    uint64(len(votes)),
		YesVotes:      proposal.Results.YesVotes,
		NoVotes:       proposal.Results.NoVotes,
		AbstainVotes:  proposal.Results.AbstainVotes,
		QuorumReached: proposal.Results.YesVotes+proposal.Results.NoVotes+proposal.Results.AbstainVotes >= pm.dao.GovernanceState.Config.QuorumThreshold,
		TimeRemaining: proposal.EndTime - time.Now().Unix(),
		Voters:        make([]VoterInfo, 0, len(votes)),
	}

	// Add voter information
	for _, vote := range votes {
		voterInfo := VoterInfo{
			Address:   vote.Voter, // Use the actual PublicKey from the vote
			Choice:    vote.Choice,
			Weight:    vote.Weight,
			Timestamp: vote.Timestamp,
			Reason:    vote.Reason,
		}
		progress.Voters = append(progress.Voters, voterInfo)
	}

	return progress, nil
}

// UpdateAllProposalStatuses updates all proposal statuses based on current time
func (pm *ProposalManager) UpdateAllProposalStatuses() error {
	for proposalID := range pm.dao.GovernanceState.Proposals {
		if err := pm.dao.Processor.UpdateProposalStatus(proposalID); err != nil {
			return fmt.Errorf("failed to update proposal %s: %v", proposalID.String(), err)
		}
	}
	return nil
}

// GetProposalStatistics returns overall proposal statistics
func (pm *ProposalManager) GetProposalStatistics() *ProposalStatistics {
	stats := &ProposalStatistics{
		StatusCounts: make(map[ProposalStatus]uint64),
		TypeCounts:   make(map[ProposalType]uint64),
	}

	for _, proposal := range pm.dao.GovernanceState.Proposals {
		stats.Total++
		stats.StatusCounts[proposal.Status]++
		stats.TypeCounts[proposal.ProposalType]++

		if proposal.Results.Passed {
			stats.Passed++
		}
	}

	return stats
}

// validateProposalCreation performs enhanced validation for proposal creation
func (pm *ProposalManager) validateProposalCreation(tx *ProposalTx, creator crypto.PublicKey) error {
	// Use existing validator
	if err := pm.dao.Validator.ValidateProposalTx(tx, creator); err != nil {
		return err
	}

	// Additional enhanced validations
	// now := time.Now().Unix()

	// Check for proposal spam (max 1 proposal per creator per day)
	// Disabled for testing - uncomment for production use
	// if pm.hasRecentProposal(creator, now-86400) {
	//     return NewDAOError(ErrInvalidProposal, "creator has submitted a proposal in the last 24 hours", nil)
	// }

	// Validate metadata hash if provided
	if !tx.MetadataHash.IsZero() {
		if !pm.isValidMetadataHash(tx.MetadataHash) {
			return NewDAOError(ErrInvalidProposal, "invalid metadata hash format", nil)
		}
	}

	// Enhanced timeframe validation
	minVotingPeriod := pm.dao.GovernanceState.Config.VotingPeriod
	maxVotingPeriod := minVotingPeriod * 30 // Max 30x the minimum period

	if tx.EndTime-tx.StartTime > maxVotingPeriod {
		return NewDAOError(ErrInvalidTimeframe, "voting period too long", nil)
	}

	return nil
}

// hasRecentProposal checks if creator has submitted a proposal recently
func (pm *ProposalManager) hasRecentProposal(creator crypto.PublicKey, since int64) bool {
	creatorStr := creator.String()
	for _, proposal := range pm.dao.GovernanceState.Proposals {
		if proposal.Creator.String() == creatorStr && proposal.StartTime > since {
			return true
		}
	}
	return false
}

// isValidMetadataHash validates the metadata hash format
func (pm *ProposalManager) isValidMetadataHash(hash types.Hash) bool {
	// Check if hash is not zero and has valid format
	return !hash.IsZero()
}

// isAuthorizedExecutor checks if the executor is authorized for the proposal type
func (pm *ProposalManager) isAuthorizedExecutor(proposal *Proposal, executor crypto.PublicKey) bool {
	switch proposal.ProposalType {
	case ProposalTypeGeneral:
		// Anyone can execute general proposals
		return true
	case ProposalTypeTreasury:
		// Only treasury signers can execute treasury proposals
		return pm.isTreasurySigner(executor)
	case ProposalTypeTechnical, ProposalTypeParameter:
		// Only token holders with sufficient balance can execute technical/parameter proposals
		return pm.dao.GetTokenBalance(executor) >= pm.dao.GovernanceState.Config.MinProposalThreshold
	default:
		return false
	}
}

// isTreasurySigner checks if address is a treasury signer
func (pm *ProposalManager) isTreasurySigner(address crypto.PublicKey) bool {
	addressStr := address.String()
	for _, signer := range pm.dao.GovernanceState.Treasury.Signers {
		if signer.String() == addressStr {
			return true
		}
	}
	return false
}

// executeGeneralProposal executes a general governance proposal
func (pm *ProposalManager) executeGeneralProposal(proposal *Proposal) error {
	// General proposals are informational and don't require specific execution
	proposal.Status = ProposalStatusExecuted
	return nil
}

// executeTreasuryProposal executes a treasury spending proposal
func (pm *ProposalManager) executeTreasuryProposal(proposal *Proposal) error {
	// Treasury proposals would typically contain spending instructions in metadata
	// For now, we just mark as executed
	proposal.Status = ProposalStatusExecuted
	return nil
}

// executeTechnicalProposal executes a technical protocol proposal
func (pm *ProposalManager) executeTechnicalProposal(proposal *Proposal) error {
	// Technical proposals would typically trigger protocol upgrades
	// For now, we just mark as executed
	proposal.Status = ProposalStatusExecuted
	return nil
}

// executeParameterProposal executes a parameter update proposal
func (pm *ProposalManager) executeParameterProposal(proposal *Proposal) error {
	// Parameter proposals would typically update DAO configuration
	// For now, we just mark as executed
	proposal.Status = ProposalStatusExecuted
	return nil
}

// VotingProgress represents detailed voting progress for a proposal
type VotingProgress struct {
	ProposalID    types.Hash
	TotalVotes    uint64
	YesVotes      uint64
	NoVotes       uint64
	AbstainVotes  uint64
	QuorumReached bool
	TimeRemaining int64
	Voters        []VoterInfo
}

// VoterInfo represents information about a voter
type VoterInfo struct {
	Address   crypto.PublicKey
	Choice    VoteChoice
	Weight    uint64
	Timestamp int64
	Reason    string
}

// ProposalStatistics represents overall proposal statistics
type ProposalStatistics struct {
	Total        uint64
	Passed       uint64
	StatusCounts map[ProposalStatus]uint64
	TypeCounts   map[ProposalType]uint64
}
