package dao

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// DAO represents the main DAO system
type DAO struct {
	GovernanceState   *GovernanceState
	TokenState        *GovernanceToken
	Processor         *DAOProcessor
	Validator         *DAOValidator
	ProposalManager   *ProposalManager
	TreasuryManager   *TreasuryManager
	ParameterManager  *ParameterManager
	TokenomicsManager *TokenomicsManager
	IPFSClient        *IPFSClient
	ReputationSystem  *ReputationSystem
	SecurityManager   *SecurityManager
	AnalyticsSystem   *AnalyticsSystem
}

// NewDAO creates a new DAO instance
func NewDAO(tokenSymbol, tokenName string, decimals uint8) *DAO {
	governanceState := NewGovernanceState()
	tokenState := NewGovernanceToken(tokenSymbol, tokenName, decimals)
	processor := NewDAOProcessor(governanceState, tokenState)
	validator := NewDAOValidator(governanceState, tokenState)

	dao := &DAO{
		GovernanceState: governanceState,
		TokenState:      tokenState,
		Processor:       processor,
		Validator:       validator,
		IPFSClient:      NewIPFSClient(""), // Use default IPFS node
		SecurityManager: NewSecurityManager(),
	}

	// Initialize ProposalManager with the DAO instance
	dao.ProposalManager = NewProposalManager(dao)

	// Initialize TreasuryManager
	dao.TreasuryManager = NewTreasuryManager(governanceState, tokenState)

	// Initialize ReputationSystem
	dao.ReputationSystem = NewReputationSystem(governanceState, tokenState)

	// Initialize ParameterManager
	dao.ParameterManager = NewParameterManager(governanceState, tokenState)

	// Initialize AnalyticsSystem
	dao.AnalyticsSystem = NewAnalyticsSystem(governanceState, tokenState)

	// Initialize TokenomicsManager
	dao.TokenomicsManager = NewTokenomicsManager(governanceState, tokenState)

	return dao
}

// InitializeTreasury sets up the treasury with initial signers and requirements
func (d *DAO) InitializeTreasury(signers []crypto.PublicKey, requiredSigs uint8) error {
	if len(signers) == 0 {
		return NewDAOError(ErrInvalidProposal, "treasury must have at least one signer", nil)
	}

	if requiredSigs == 0 || requiredSigs > uint8(len(signers)) {
		return NewDAOError(ErrInvalidProposal, "invalid required signatures count", nil)
	}

	d.GovernanceState.Treasury.Signers = signers
	d.GovernanceState.Treasury.RequiredSigs = requiredSigs

	return nil
}

// InitialTokenDistribution distributes initial tokens to founding members
func (d *DAO) InitialTokenDistribution(distributions map[string]uint64) error {
	totalDistribution := uint64(0)

	// Calculate total distribution
	for _, amount := range distributions {
		totalDistribution += amount
	}

	// Update token state
	d.TokenState.TotalSupply = totalDistribution

	// Distribute tokens
	for recipientStr, amount := range distributions {
		d.TokenState.Balances[recipientStr] = amount

		// Create token holder record
		d.GovernanceState.TokenHolders[recipientStr] = &TokenHolder{
			Address:    crypto.PublicKey([]byte(recipientStr)), // Convert string to PublicKey
			Balance:    amount,
			Staked:     0,
			Reputation: 0, // Will be initialized by reputation system
			JoinedAt:   0, // Genesis time
			LastActive: 0,
		}

		// Initialize reputation using the reputation system
		// The recipientStr is a hex-encoded public key, so we need to decode it
		pubKeyBytes, err := hex.DecodeString(recipientStr)
		if err != nil {
			// If decoding fails, create a dummy PublicKey from the string
			pubKeyBytes = []byte(recipientStr)
			if len(pubKeyBytes) > 64 {
				pubKeyBytes = pubKeyBytes[:64]
			}
		}
		pubKey := crypto.PublicKey(pubKeyBytes)
		d.ReputationSystem.InitializeReputation(pubKey, amount)
	}

	return nil
}

// GetProposal retrieves a proposal by ID
func (d *DAO) GetProposal(proposalID types.Hash) (*Proposal, error) {
	proposal, exists := d.GovernanceState.Proposals[proposalID]
	if !exists {
		return nil, ErrProposalNotFoundError
	}
	return proposal, nil
}

// GetVotes retrieves all votes for a proposal
func (d *DAO) GetVotes(proposalID types.Hash) (map[string]*Vote, error) {
	votes, exists := d.GovernanceState.Votes[proposalID]
	if !exists {
		return nil, ErrProposalNotFoundError
	}
	return votes, nil
}

// GetTokenBalance retrieves the token balance for an address
func (d *DAO) GetTokenBalance(address crypto.PublicKey) uint64 {
	return d.TokenState.Balances[address.String()]
}

// GetTotalSupply returns the total token supply
func (d *DAO) GetTotalSupply() uint64 {
	return d.TokenState.TotalSupply
}

// GetTreasuryBalance returns the current treasury balance
func (d *DAO) GetTreasuryBalance() uint64 {
	return d.GovernanceState.Treasury.Balance
}

// GetDelegation retrieves delegation information for an address
func (d *DAO) GetDelegation(delegator crypto.PublicKey) (*Delegation, bool) {
	delegation, exists := d.GovernanceState.Delegations[delegator.String()]
	return delegation, exists
}

// GetEffectiveVotingPower returns the effective voting power for a user
func (d *DAO) GetEffectiveVotingPower(user crypto.PublicKey) uint64 {
	return d.Processor.GetEffectiveVotingPower(user)
}

// GetDelegatedPower returns the total voting power delegated to a user
func (d *DAO) GetDelegatedPower(delegate crypto.PublicKey) uint64 {
	return d.Processor.GetDelegatedPower(delegate)
}

// GetOwnVotingPower returns the user's own voting power (excluding delegations)
func (d *DAO) GetOwnVotingPower(user crypto.PublicKey) uint64 {
	return d.Processor.GetOwnVotingPower(user)
}

// RevokeDelegation revokes an active delegation
func (d *DAO) RevokeDelegation(delegator crypto.PublicKey) error {
	return d.Processor.RevokeDelegation(delegator)
}

// ListDelegations returns all active delegations
func (d *DAO) ListDelegations() map[string]*Delegation {
	activeDelegations := make(map[string]*Delegation)
	now := time.Now().Unix()

	for delegatorStr, delegation := range d.GovernanceState.Delegations {
		if delegation.Active && now >= delegation.StartTime && now <= delegation.EndTime {
			activeDelegations[delegatorStr] = delegation
		}
	}

	return activeDelegations
}

// GetDelegationsByDelegate returns all delegations for a specific delegate
func (d *DAO) GetDelegationsByDelegate(delegate crypto.PublicKey) []*Delegation {
	var delegations []*Delegation
	delegateStr := delegate.String()
	now := time.Now().Unix()

	for _, delegation := range d.GovernanceState.Delegations {
		if delegation.Active && delegation.Delegate.String() == delegateStr {
			if now >= delegation.StartTime && now <= delegation.EndTime {
				delegations = append(delegations, delegation)
			}
		}
	}

	return delegations
}

// GetTokenHolder retrieves token holder information
func (d *DAO) GetTokenHolder(address crypto.PublicKey) (*TokenHolder, bool) {
	holder, exists := d.GovernanceState.TokenHolders[address.String()]
	return holder, exists
}

// ListActiveProposals returns all currently active proposals
func (d *DAO) ListActiveProposals() []*Proposal {
	var activeProposals []*Proposal

	for _, proposal := range d.GovernanceState.Proposals {
		if proposal.Status == ProposalStatusActive {
			activeProposals = append(activeProposals, proposal)
		}
	}

	return activeProposals
}

// ListAllProposals returns all proposals
func (d *DAO) ListAllProposals() []*Proposal {
	var allProposals []*Proposal

	for _, proposal := range d.GovernanceState.Proposals {
		allProposals = append(allProposals, proposal)
	}

	return allProposals
}

// UpdateConfig updates DAO configuration parameters
func (d *DAO) UpdateConfig(newConfig *DAOConfig) error {
	// Validate new configuration
	if newConfig.MinProposalThreshold == 0 {
		return NewDAOError(ErrInvalidProposal, "minimum proposal threshold must be greater than zero", nil)
	}

	if newConfig.VotingPeriod <= 0 {
		return NewDAOError(ErrInvalidProposal, "voting period must be positive", nil)
	}

	if newConfig.QuorumThreshold == 0 {
		return NewDAOError(ErrInvalidProposal, "quorum threshold must be greater than zero", nil)
	}

	if newConfig.PassingThreshold == 0 || newConfig.PassingThreshold > 10000 {
		return NewDAOError(ErrInvalidProposal, "passing threshold must be between 1 and 10000 basis points", nil)
	}

	d.GovernanceState.Config = newConfig
	return nil
}

// AddTreasuryFunds adds funds to the treasury
func (d *DAO) AddTreasuryFunds(amount uint64) {
	d.TreasuryManager.AddTreasuryFunds(amount)
}

// CreateTreasuryTransaction creates a new treasury transaction
func (d *DAO) CreateTreasuryTransaction(tx *TreasuryTx, txHash types.Hash) error {
	return d.TreasuryManager.CreateTreasuryTransaction(tx, txHash)
}

// SignTreasuryTransaction adds a signature to a pending treasury transaction
func (d *DAO) SignTreasuryTransaction(txHash types.Hash, signer crypto.PrivateKey) error {
	return d.TreasuryManager.SignTreasuryTransaction(txHash, signer)
}

// ExecuteTreasuryTransaction executes a treasury transaction if it has sufficient signatures
func (d *DAO) ExecuteTreasuryTransaction(txHash types.Hash) error {
	return d.TreasuryManager.ExecuteTreasuryTransaction(txHash)
}

// GetPendingTreasuryTransactions returns all pending treasury transactions
func (d *DAO) GetPendingTreasuryTransactions() map[types.Hash]*PendingTx {
	return d.TreasuryManager.GetPendingTreasuryTransactions()
}

// GetTreasuryTransaction returns a specific treasury transaction
func (d *DAO) GetTreasuryTransaction(txHash types.Hash) (*PendingTx, bool) {
	return d.TreasuryManager.GetTreasuryTransaction(txHash)
}

// GetTreasurySigners returns the list of authorized treasury signers
func (d *DAO) GetTreasurySigners() []crypto.PublicKey {
	return d.TreasuryManager.GetTreasurySigners()
}

// GetRequiredSignatures returns the number of required signatures
func (d *DAO) GetRequiredSignatures() uint8 {
	return d.TreasuryManager.GetRequiredSignatures()
}

// UpdateTreasurySigners updates the treasury signers (requires governance approval)
func (d *DAO) UpdateTreasurySigners(signers []crypto.PublicKey, requiredSigs uint8) error {
	return d.TreasuryManager.UpdateTreasurySigners(signers, requiredSigs)
}

// CleanupExpiredTransactions removes expired treasury transactions
func (d *DAO) CleanupExpiredTransactions() int {
	return d.TreasuryManager.CleanupExpiredTransactions()
}

// GetTreasuryHistory returns all treasury transactions (executed and pending)
func (d *DAO) GetTreasuryHistory() map[types.Hash]*PendingTx {
	return d.TreasuryManager.GetTreasuryHistory()
}

// GetExecutedTreasuryTransactions returns only executed treasury transactions
func (d *DAO) GetExecutedTreasuryTransactions() map[types.Hash]*PendingTx {
	return d.TreasuryManager.GetExecutedTreasuryTransactions()
}

// ProcessDAOTransaction processes any DAO transaction type
func (d *DAO) ProcessDAOTransaction(txInner interface{}, from crypto.PublicKey, txHash types.Hash) error {
	switch tx := txInner.(type) {
	case *ProposalTx:
		return d.Processor.ProcessProposalTx(tx, from, txHash)
	case *VoteTx:
		return d.Processor.ProcessVoteTx(tx, from)
	case *DelegationTx:
		return d.Processor.ProcessDelegationTx(tx, from)
	case *TreasuryTx:
		return d.Processor.ProcessTreasuryTx(tx, txHash)
	case *TokenMintTx:
		return d.Processor.ProcessTokenMintTx(tx, from)
	case *TokenBurnTx:
		return d.Processor.ProcessTokenBurnTx(tx, from)
	case *TokenTransferTx:
		return d.Processor.ProcessTokenTransferTx(tx, from)
	case *TokenApproveTx:
		return d.Processor.ProcessTokenApproveTx(tx, from)
	case *TokenTransferFromTx:
		return d.Processor.ProcessTokenTransferFromTx(tx, from)
	case *ParameterProposalTx:
		return d.Processor.ProcessParameterProposalTx(tx, from, txHash)
	case *TokenDistributionTx:
		return d.Processor.ProcessTokenDistributionTx(tx, from)
	case *VestingClaimTx:
		return d.Processor.ProcessVestingClaimTx(tx, from)
	case *StakeTx:
		return d.Processor.ProcessStakeTx(tx, from)
	case *UnstakeTx:
		return d.Processor.ProcessUnstakeTx(tx, from)
	case *ClaimRewardsTx:
		return d.Processor.ProcessClaimRewardsTx(tx, from)
	default:
		return NewDAOError(ErrInvalidProposal, "unknown DAO transaction type", nil)
	}
}

// UpdateAllProposalStatuses updates the status of all proposals based on current time
func (d *DAO) UpdateAllProposalStatuses() {
	for proposalID := range d.GovernanceState.Proposals {
		d.Processor.UpdateProposalStatus(proposalID)
	}
}

// TransferTokens transfers tokens between addresses
func (d *DAO) TransferTokens(from, to crypto.PublicKey, amount uint64) error {
	return d.TokenState.Transfer(from.String(), to.String(), amount)
}

// ApproveTokens approves a spender to spend tokens on behalf of the owner
func (d *DAO) ApproveTokens(owner, spender crypto.PublicKey, amount uint64) error {
	return d.TokenState.Approve(owner.String(), spender.String(), amount)
}

// GetTokenAllowance returns the allowance between owner and spender
func (d *DAO) GetTokenAllowance(owner, spender crypto.PublicKey) uint64 {
	return d.TokenState.GetAllowance(owner.String(), spender.String())
}

// MintTokens mints new tokens to an address
func (d *DAO) MintTokens(to crypto.PublicKey, amount uint64) error {
	return d.TokenState.Mint(to.String(), amount)
}

// BurnTokens burns tokens from an address
func (d *DAO) BurnTokens(from crypto.PublicKey, amount uint64) error {
	return d.TokenState.Burn(from.String(), amount)
}

// IPFS-related methods

// CreateProposalWithMetadata creates a proposal with rich metadata stored on IPFS
func (d *DAO) CreateProposalWithMetadata(creator crypto.PublicKey, title, description, details string, documents []DocumentReference, links []LinkReference, tags []string, proposalType ProposalType, votingType VotingType, startTime, endTime int64, threshold uint64) (types.Hash, types.Hash, error) {
	// Upload metadata to IPFS
	_, metadataHash, err := d.IPFSClient.CreateProposalWithIPFS(title, description, details, documents, links, tags)
	if err != nil {
		return types.Hash{}, types.Hash{}, fmt.Errorf("failed to upload metadata to IPFS: %w", err)
	}

	// Pin the metadata to prevent garbage collection
	if err := d.IPFSClient.PinContent(metadataHash); err != nil {
		// Log warning but don't fail the proposal creation
		// In production, you might want to handle this differently
	}

	// Create the proposal transaction
	proposalTx := &ProposalTx{
		Fee:          200, // Standard fee
		Title:        title,
		Description:  description,
		ProposalType: proposalType,
		VotingType:   votingType,
		StartTime:    startTime,
		EndTime:      endTime,
		Threshold:    threshold,
		MetadataHash: metadataHash,
	}

	// Generate proposal hash
	proposalHash := d.generateProposalHash(proposalTx, creator)

	// Process the proposal
	if err := d.Processor.ProcessProposalTx(proposalTx, creator, proposalHash); err != nil {
		return types.Hash{}, types.Hash{}, fmt.Errorf("failed to process proposal: %w", err)
	}

	return proposalHash, metadataHash, nil
}

// GetProposalMetadata retrieves the full metadata for a proposal from IPFS
func (d *DAO) GetProposalMetadata(proposalID types.Hash) (*ProposalMetadata, error) {
	proposal, err := d.GetProposal(proposalID)
	if err != nil {
		return nil, err
	}

	if proposal.MetadataHash == (types.Hash{}) {
		return nil, fmt.Errorf("proposal has no metadata hash")
	}

	return d.IPFSClient.RetrieveProposalMetadata(proposal.MetadataHash)
}

// UpdateProposalMetadata updates the metadata for an existing proposal
func (d *DAO) UpdateProposalMetadata(proposalID types.Hash, updates *ProposalMetadata) (types.Hash, error) {
	proposal, err := d.GetProposal(proposalID)
	if err != nil {
		return types.Hash{}, err
	}

	if proposal.MetadataHash == (types.Hash{}) {
		return types.Hash{}, fmt.Errorf("proposal has no existing metadata")
	}

	// Update metadata on IPFS
	newMetadataHash, err := d.IPFSClient.UpdateProposalMetadata(proposal.MetadataHash, updates)
	if err != nil {
		return types.Hash{}, fmt.Errorf("failed to update metadata on IPFS: %w", err)
	}

	// Pin new metadata
	if err := d.IPFSClient.PinContent(newMetadataHash); err != nil {
		// Log warning but continue
	}

	// Unpin old metadata
	if err := d.IPFSClient.UnpinContent(proposal.MetadataHash); err != nil {
		// Log warning but continue
	}

	// Update proposal with new metadata hash
	proposal.MetadataHash = newMetadataHash

	return newMetadataHash, nil
}

// UploadProposalDocument uploads a document related to a proposal
func (d *DAO) UploadProposalDocument(name string, data []byte, mimeType string) (*DocumentReference, error) {
	return d.IPFSClient.UploadDocument(name, data, mimeType)
}

// RetrieveProposalDocument retrieves a document from IPFS
func (d *DAO) RetrieveProposalDocument(docRef *DocumentReference) ([]byte, error) {
	return d.IPFSClient.RetrieveDocument(docRef)
}

// VerifyProposalMetadata verifies that proposal metadata exists and is accessible
func (d *DAO) VerifyProposalMetadata(proposalID types.Hash) (bool, error) {
	proposal, err := d.GetProposal(proposalID)
	if err != nil {
		return false, err
	}

	if proposal.MetadataHash == (types.Hash{}) {
		return false, nil
	}

	return d.IPFSClient.VerifyContentExists(proposal.MetadataHash)
}

// GetIPFSNodeInfo returns information about the connected IPFS node
func (d *DAO) GetIPFSNodeInfo() (map[string]interface{}, error) {
	return d.IPFSClient.GetNodeInfo()
}

// ListPinnedContent returns all pinned IPFS content
func (d *DAO) ListPinnedContent() ([]types.Hash, error) {
	return d.IPFSClient.ListPinnedContent()
}

// CleanupUnusedMetadata unpins metadata for proposals that are no longer active
func (d *DAO) CleanupUnusedMetadata() error {
	// Get all pinned content
	pinnedHashes, err := d.IPFSClient.ListPinnedContent()
	if err != nil {
		return fmt.Errorf("failed to list pinned content: %w", err)
	}

	// Get all active proposals
	activeProposals := d.ListActiveProposals()
	activeMetadataHashes := make(map[types.Hash]bool)

	for _, proposal := range activeProposals {
		if proposal.MetadataHash != (types.Hash{}) {
			activeMetadataHashes[proposal.MetadataHash] = true
		}
	}

	// Unpin metadata that's not associated with active proposals
	for _, hash := range pinnedHashes {
		if !activeMetadataHashes[hash] {
			if err := d.IPFSClient.UnpinContent(hash); err != nil {
				// Log error but continue with other hashes
				continue
			}
		}
	}

	return nil
}

// Reputation-related methods

// InitializeUserReputation initializes reputation for a new token holder
func (d *DAO) InitializeUserReputation(address crypto.PublicKey, tokenBalance uint64) {
	d.ReputationSystem.InitializeReputation(address, tokenBalance)
}

// GetUserReputation returns the reputation score for a user
func (d *DAO) GetUserReputation(address crypto.PublicKey) uint64 {
	if holder, exists := d.GovernanceState.TokenHolders[address.String()]; exists {
		return holder.Reputation
	}
	return 0
}

// GetReputationRanking returns users sorted by reputation
func (d *DAO) GetReputationRanking() []*TokenHolder {
	return d.ReputationSystem.GetReputationRanking()
}

// GetReputationStats returns statistics about the reputation system
func (d *DAO) GetReputationStats() *ReputationStats {
	return d.ReputationSystem.GetReputationStats()
}

// UpdateReputationConfig updates the reputation system configuration
func (d *DAO) UpdateReputationConfig(newConfig *ReputationConfig) error {
	return d.ReputationSystem.UpdateReputationConfig(newConfig)
}

// GetReputationConfig returns the current reputation configuration
func (d *DAO) GetReputationConfig() *ReputationConfig {
	return d.ReputationSystem.GetReputationConfig()
}

// ApplyInactivityDecay applies reputation decay for inactive users
func (d *DAO) ApplyInactivityDecay() {
	d.ReputationSystem.ApplyInactivityDecay()
}

// RecalculateAllReputation recalculates reputation for all users
func (d *DAO) RecalculateAllReputation() {
	d.ReputationSystem.RecalculateAllReputation()
}

// GetUserReputationHistory returns reputation history for a user
func (d *DAO) GetUserReputationHistory(user crypto.PublicKey) *UserReputationHistory {
	return d.ReputationSystem.GetUserReputationHistory(user)
}

// generateProposalHash generates a hash for a proposal
func (d *DAO) generateProposalHash(tx *ProposalTx, creator crypto.PublicKey) types.Hash {
	// This is a simplified hash generation
	// In practice, you'd want to use the same hashing mechanism as the blockchain
	data := fmt.Sprintf("%s%s%d%d%s", tx.Title, tx.Description, tx.StartTime, tx.EndTime, creator.String())
	hash := [32]byte{}
	copy(hash[:], []byte(data)[:32])
	return hash
}

// Security-related methods

// GrantRole grants a role to a user with security logging
func (d *DAO) GrantRole(user crypto.PublicKey, role Role, grantedBy crypto.PublicKey, duration int64) error {
	err := d.SecurityManager.GrantRole(user, role, grantedBy, duration)
	if err != nil {
		return err
	}

	// Log the role grant in audit log
	d.SecurityManager.LogAuditEvent(grantedBy, "DAO_GRANT_ROLE", user.String(), "SUCCESS",
		map[string]interface{}{"role": role, "duration": duration}, SecurityLevelSensitive)

	return nil
}

// RevokeRole revokes a role from a user with security logging
func (d *DAO) RevokeRole(user crypto.PublicKey, revokedBy crypto.PublicKey) error {
	err := d.SecurityManager.RevokeRole(user, revokedBy)
	if err != nil {
		return err
	}

	// Log the role revocation in audit log
	d.SecurityManager.LogAuditEvent(revokedBy, "DAO_REVOKE_ROLE", user.String(), "SUCCESS",
		nil, SecurityLevelSensitive)

	return nil
}

// HasPermission checks if a user has a specific permission
func (d *DAO) HasPermission(user crypto.PublicKey, permission Permission) bool {
	return d.SecurityManager.HasPermission(user, permission)
}

// GetUserRole returns the role of a user
func (d *DAO) GetUserRole(user crypto.PublicKey) (Role, bool) {
	return d.SecurityManager.GetUserRole(user)
}

// ActivateEmergency activates emergency mode with security validation
func (d *DAO) ActivateEmergency(activatedBy crypto.PublicKey, reason string, level SecurityLevel, affectedFunctions []string) error {
	return d.SecurityManager.ActivateEmergency(activatedBy, reason, level, affectedFunctions)
}

// DeactivateEmergency deactivates emergency mode with security validation
func (d *DAO) DeactivateEmergency(deactivatedBy crypto.PublicKey) error {
	return d.SecurityManager.DeactivateEmergency(deactivatedBy)
}

// IsEmergencyActive returns whether emergency mode is active
func (d *DAO) IsEmergencyActive() bool {
	return d.SecurityManager.IsEmergencyActive()
}

// IsFunctionPaused checks if a specific function is paused
func (d *DAO) IsFunctionPaused(functionName string) bool {
	return d.SecurityManager.IsFunctionPaused(functionName)
}

// GetAuditLog returns audit log entries with permission validation
func (d *DAO) GetAuditLog(user crypto.PublicKey, limit int, offset int, minLevel SecurityLevel) ([]*AuditLogEntry, error) {
	return d.SecurityManager.GetAuditLog(user, limit, offset, minLevel)
}

// ValidateAccess validates access for a specific operation
func (d *DAO) ValidateAccess(user crypto.PublicKey, operation string, resource string, level SecurityLevel) error {
	return d.SecurityManager.ValidateAccess(user, operation, resource, level)
}

// GetSecurityConfig returns the current security configuration
func (d *DAO) GetSecurityConfig(requestedBy crypto.PublicKey) (*SecurityConfig, error) {
	return d.SecurityManager.GetSecurityConfig(requestedBy)
}

// UpdateSecurityConfig updates the security configuration
func (d *DAO) UpdateSecurityConfig(updatedBy crypto.PublicKey, newConfig *SecurityConfig) error {
	return d.SecurityManager.UpdateSecurityConfig(updatedBy, newConfig)
}

// ListActiveRoles returns all active role assignments
func (d *DAO) ListActiveRoles(requestedBy crypto.PublicKey) (map[string]*AccessControlEntry, error) {
	return d.SecurityManager.ListActiveRoles(requestedBy)
}

// GetEmergencyState returns the current emergency state
func (d *DAO) GetEmergencyState(requestedBy crypto.PublicKey) (*EmergencyState, error) {
	return d.SecurityManager.GetEmergencyState(requestedBy)
}

// AddEmergencyContact adds an emergency contact
func (d *DAO) AddEmergencyContact(contact crypto.PublicKey, addedBy crypto.PublicKey) error {
	return d.SecurityManager.AddEmergencyContact(contact, addedBy)
}

// GetEmergencyContacts returns the list of emergency contacts
func (d *DAO) GetEmergencyContacts(requestedBy crypto.PublicKey) ([]crypto.PublicKey, error) {
	return d.SecurityManager.GetEmergencyContacts(requestedBy)
}

// SecureProcessDAOTransaction processes DAO transactions with security validation
func (d *DAO) SecureProcessDAOTransaction(txInner interface{}, from crypto.PublicKey, txHash types.Hash) error {
	// Determine operation type and required permission
	var operation string
	var permission Permission
	var securityLevel SecurityLevel

	switch txInner.(type) {
	case *ProposalTx:
		operation = "CreateProposal"
		permission = PermissionCreateProposal
		securityLevel = SecurityLevelMember
	case *VoteTx:
		operation = "Vote"
		permission = PermissionVote
		securityLevel = SecurityLevelMember
	case *DelegationTx:
		operation = "Delegate"
		permission = PermissionDelegate
		securityLevel = SecurityLevelMember
	case *TreasuryTx:
		operation = "TreasuryOperation"
		permission = PermissionManageTreasury
		securityLevel = SecurityLevelSensitive
	case *TokenMintTx:
		operation = "MintTokens"
		permission = PermissionManageRoles // Only admins can mint
		securityLevel = SecurityLevelSensitive
	case *TokenBurnTx:
		operation = "BurnTokens"
		permission = PermissionVote // Users can burn their own tokens
		securityLevel = SecurityLevelMember
	case *ParameterProposalTx:
		operation = "ParameterProposal"
		permission = PermissionCreateProposal
		securityLevel = SecurityLevelSensitive
	default:
		d.SecurityManager.LogAuditEvent(from, "UNKNOWN_TRANSACTION", txHash.String(), "BLOCKED",
			map[string]interface{}{"type": fmt.Sprintf("%T", txInner)}, SecurityLevelCritical)
		return NewDAOError(ErrInvalidProposal, "unknown DAO transaction type", nil)
	}

	// Validate access
	if err := d.ValidateAccess(from, operation, txHash.String(), securityLevel); err != nil {
		return err
	}

	// Check permissions
	if !d.HasPermission(from, permission) {
		d.SecurityManager.LogAuditEvent(from, operation, txHash.String(), "DENIED",
			map[string]interface{}{"reason": "insufficient_permissions"}, securityLevel)
		return NewDAOError(ErrUnauthorized, fmt.Sprintf("insufficient permissions for %s", operation), nil)
	}

	// Process the transaction
	err := d.ProcessDAOTransaction(txInner, from, txHash)

	// Log the result
	result := "SUCCESS"
	if err != nil {
		result = "FAILURE"
	}

	d.SecurityManager.LogAuditEvent(from, operation, txHash.String(), result,
		map[string]interface{}{"error": err}, securityLevel)

	return err
}

// InitializeFounderRoles sets up initial roles for DAO founders
func (d *DAO) InitializeFounderRoles(founders []crypto.PublicKey) error {
	if len(founders) == 0 {
		return NewDAOError(ErrInvalidProposal, "must have at least one founder", nil)
	}

	// Grant super admin role to first founder
	firstFounder := founders[0]
	d.SecurityManager.accessControl[firstFounder.String()] = &AccessControlEntry{
		User:        firstFounder,
		Role:        RoleSuperAdmin,
		Permissions: d.SecurityManager.rolePermissions[RoleSuperAdmin],
		GrantedBy:   firstFounder,
		GrantedAt:   time.Now().Unix(),
		ExpiresAt:   0,
		Active:      true,
	}

	// Grant admin roles to other founders
	for i := 1; i < len(founders); i++ {
		err := d.GrantRole(founders[i], RoleAdmin, firstFounder, 0)
		if err != nil {
			return fmt.Errorf("failed to grant admin role to founder %s: %w", founders[i].String(), err)
		}
	}

	// Log the initialization
	d.SecurityManager.LogAuditEvent(firstFounder, "INITIALIZE_FOUNDER_ROLES", "system", "SUCCESS",
		map[string]interface{}{"founders": len(founders)}, SecurityLevelCritical)

	return nil
}

// PerformSecurityAudit performs a comprehensive security audit
func (d *DAO) PerformSecurityAudit(auditedBy crypto.PublicKey) (map[string]interface{}, error) {
	// Check if user has audit permissions
	if !d.HasPermission(auditedBy, PermissionAuditAccess) {
		return nil, NewDAOError(ErrUnauthorized, "insufficient permissions for security audit", nil)
	}

	auditResults := make(map[string]interface{})

	// Check active roles
	activeRoles, err := d.ListActiveRoles(auditedBy)
	if err != nil {
		return nil, err
	}
	auditResults["active_roles_count"] = len(activeRoles)

	// Check emergency state
	emergencyState, err := d.GetEmergencyState(auditedBy)
	if err != nil {
		return nil, err
	}
	auditResults["emergency_active"] = emergencyState.Active

	// Check recent audit log entries
	recentEntries, err := d.GetAuditLog(auditedBy, 100, 0, SecurityLevelPublic)
	if err != nil {
		return nil, err
	}
	auditResults["recent_audit_entries"] = len(recentEntries)

	// Count failed operations in recent entries
	failedOps := 0
	for _, entry := range recentEntries {
		if entry.Result == "FAILURE" || entry.Result == "BLOCKED" || entry.Result == "DENIED" {
			failedOps++
		}
	}
	auditResults["failed_operations"] = failedOps

	// Check security config
	securityConfig, err := d.GetSecurityConfig(auditedBy)
	if err != nil {
		return nil, err
	}
	auditResults["mfa_required"] = securityConfig.RequireMFA
	auditResults["strong_passwords_required"] = securityConfig.RequireStrongPasswords

	// Log the audit
	d.SecurityManager.LogAuditEvent(auditedBy, "SECURITY_AUDIT", "system", "SUCCESS",
		auditResults, SecurityLevelSensitive)

	return auditResults, nil
}

// Parameter Management Methods

// CreateParameterProposal creates a new parameter change proposal
func (d *DAO) CreateParameterProposal(creator crypto.PublicKey, parameterChanges map[string]interface{}, justification string, effectiveTime int64, votingType VotingType, startTime, endTime int64, threshold uint64) (types.Hash, error) {
	return d.ParameterManager.CreateParameterProposal(creator, parameterChanges, justification, effectiveTime, votingType, startTime, endTime, threshold)
}

// Analytics-related methods

// GetGovernanceParticipationMetrics returns comprehensive participation analytics
func (d *DAO) GetGovernanceParticipationMetrics() *GovernanceParticipationMetrics {
	return d.AnalyticsSystem.GetGovernanceParticipationMetrics()
}

// GetTreasuryPerformanceMetrics returns treasury performance analytics
func (d *DAO) GetTreasuryPerformanceMetrics() *TreasuryPerformanceMetrics {
	return d.AnalyticsSystem.GetTreasuryPerformanceMetrics()
}

// GetProposalAnalytics returns proposal success rates and patterns
func (d *DAO) GetProposalAnalytics() *ProposalAnalytics {
	return d.AnalyticsSystem.GetProposalAnalytics()
}

// GetDAOHealthMetrics returns overall DAO health indicators
func (d *DAO) GetDAOHealthMetrics() *DAOHealthMetrics {
	return d.AnalyticsSystem.GetDAOHealthMetrics()
}

// GetAnalyticsSummary returns a comprehensive analytics summary
func (d *DAO) GetAnalyticsSummary() map[string]interface{} {
	return d.AnalyticsSystem.GetAnalyticsSummary()
}

// ExecuteParameterChanges executes approved parameter changes
func (d *DAO) ExecuteParameterChanges(proposalID types.Hash, executor crypto.PublicKey) error {
	return d.ParameterManager.ExecuteParameterChanges(proposalID, executor)
}

// GetParameterConfig returns the current parameter configuration
func (d *DAO) GetParameterConfig() *ParameterConfig {
	return d.ParameterManager.GetParameterConfig()
}

// GetParameterHistory returns the change history for a parameter
func (d *DAO) GetParameterHistory(parameter string) []*ParameterChange {
	return d.ParameterManager.GetParameterHistory(parameter)
}

// GetAllParameterHistory returns the complete parameter change history
func (d *DAO) GetAllParameterHistory() map[string][]*ParameterChange {
	return d.ParameterManager.GetAllParameterHistory()
}

// ValidateParameterProposal validates a parameter proposal before creation
func (d *DAO) ValidateParameterProposal(creator crypto.PublicKey, parameterChanges map[string]interface{}) error {
	return d.ParameterManager.ValidateParameterProposal(creator, parameterChanges)
}

// GetParameterValue returns the current value of a specific parameter
func (d *DAO) GetParameterValue(parameter string) (interface{}, error) {
	return d.ParameterManager.GetParameterValue(parameter)
}

// ListAllParameters returns all configurable parameters and their current values
func (d *DAO) ListAllParameters() map[string]interface{} {
	return d.ParameterManager.ListAllParameters()
}

// IsParameterChangeAllowed checks if a parameter change is allowed based on current state
func (d *DAO) IsParameterChangeAllowed(parameter string, newValue interface{}) (bool, string) {
	return d.ParameterManager.IsParameterChangeAllowed(parameter, newValue)
}

// GetParameterConstraints returns the constraints for a specific parameter
func (d *DAO) GetParameterConstraints(parameter string) map[string]interface{} {
	return d.ParameterManager.GetParameterConstraints(parameter)
}
// Tokenomics-related methods

// InitializeTokenomics sets up the initial token distribution system
func (d *DAO) InitializeTokenomics() error {
	return d.TokenomicsManager.InitializeTokenDistribution()
}

// AddDistributionRecipient adds a recipient to a distribution category
func (d *DAO) AddDistributionRecipient(category DistributionCategory, recipient crypto.PublicKey, amount uint64) error {
	return d.TokenomicsManager.AddDistributionRecipient(category, recipient, amount)
}

// ClaimVestedTokens allows a beneficiary to claim vested tokens
func (d *DAO) ClaimVestedTokens(vestingID string, beneficiary crypto.PublicKey) (uint64, error) {
	return d.TokenomicsManager.ClaimVestedTokens(vestingID, beneficiary)
}

// CreateStakingPool creates a new staking pool
func (d *DAO) CreateStakingPool(poolID, name string, rewardRate, minStakeAmount, lockupPeriod uint64) error {
	return d.TokenomicsManager.CreateStakingPool(poolID, name, rewardRate, minStakeAmount, lockupPeriod)
}

// StakeTokens stakes tokens in a staking pool
func (d *DAO) StakeTokens(poolID string, staker crypto.PublicKey, amount uint64, lockDuration int64) error {
	return d.TokenomicsManager.StakeTokens(poolID, staker, amount, lockDuration)
}

// UnstakeTokens unstakes tokens from a staking pool
func (d *DAO) UnstakeTokens(poolID string, staker crypto.PublicKey, amount uint64) error {
	return d.TokenomicsManager.UnstakeTokens(poolID, staker, amount)
}

// ClaimStakingRewards claims accumulated staking rewards
func (d *DAO) ClaimStakingRewards(poolID string, staker crypto.PublicKey) (uint64, error) {
	return d.TokenomicsManager.ClaimStakingRewards(poolID, staker)
}

// GetDistribution returns a distribution by category
func (d *DAO) GetDistribution(category DistributionCategory) (*TokenDistribution, bool) {
	return d.TokenomicsManager.GetDistribution(category)
}

// Get
