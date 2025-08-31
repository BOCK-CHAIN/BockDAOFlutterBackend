package dao

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// randomHash generates a random hash for testing
func randomHash() types.Hash {
	b := make([]byte, 32)
	rand.Read(b)
	return types.HashFromBytes(b)
}

func TestNewDAO(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	if dao.TokenState.Symbol != "GOV" {
		t.Errorf("Expected token symbol 'GOV', got %s", dao.TokenState.Symbol)
	}

	if dao.TokenState.Name != "Governance Token" {
		t.Errorf("Expected token name 'Governance Token', got %s", dao.TokenState.Name)
	}

	if dao.TokenState.Decimals != 18 {
		t.Errorf("Expected decimals 18, got %d", dao.TokenState.Decimals)
	}

	if dao.TokenState.TotalSupply != 0 {
		t.Errorf("Expected initial supply 0, got %d", dao.TokenState.TotalSupply)
	}
}

func TestInitialTokenDistribution(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Create test addresses
	addr1 := crypto.GeneratePrivateKey().PublicKey()
	addr2 := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		addr1.String(): 1000,
		addr2.String(): 2000,
	}

	err := dao.InitialTokenDistribution(distributions)
	if err != nil {
		t.Fatalf("Failed to distribute initial tokens: %v", err)
	}

	if dao.TokenState.TotalSupply != 3000 {
		t.Errorf("Expected total supply 3000, got %d", dao.TokenState.TotalSupply)
	}

	if dao.TokenState.Balances[addr1.String()] != 1000 {
		t.Errorf("Expected addr1 balance 1000, got %d", dao.TokenState.Balances[addr1.String()])
	}

	if dao.TokenState.Balances[addr2.String()] != 2000 {
		t.Errorf("Expected addr2 balance 2000, got %d", dao.TokenState.Balances[addr2.String()])
	}

	// Check token holder records
	holder1, exists := dao.GetTokenHolder(addr1)
	if !exists {
		t.Error("Token holder record not created for addr1")
	}
	if holder1.Balance != 1000 {
		t.Errorf("Expected holder1 balance 1000, got %d", holder1.Balance)
	}
	if holder1.Reputation != 110 { // base(100) + tokens(1000/100=10) = 110
		t.Errorf("Expected holder1 reputation 110, got %d", holder1.Reputation)
	}
}

func TestInitializeTreasury(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Create test signers
	signer1 := crypto.GeneratePrivateKey().PublicKey()
	signer2 := crypto.GeneratePrivateKey().PublicKey()
	signers := []crypto.PublicKey{signer1, signer2}

	err := dao.InitializeTreasury(signers, 2)
	if err != nil {
		t.Fatalf("Failed to initialize treasury: %v", err)
	}

	if len(dao.GovernanceState.Treasury.Signers) != 2 {
		t.Errorf("Expected 2 signers, got %d", len(dao.GovernanceState.Treasury.Signers))
	}

	if dao.GovernanceState.Treasury.RequiredSigs != 2 {
		t.Errorf("Expected required sigs 2, got %d", dao.GovernanceState.Treasury.RequiredSigs)
	}

	// Test invalid cases
	err = dao.InitializeTreasury([]crypto.PublicKey{}, 1)
	if err == nil {
		t.Error("Expected error for empty signers")
	}

	err = dao.InitializeTreasury(signers, 0)
	if err == nil {
		t.Error("Expected error for zero required sigs")
	}

	err = dao.InitializeTreasury(signers, 3)
	if err == nil {
		t.Error("Expected error for required sigs > signers")
	}
}

func TestProposalCreation(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	creator := crypto.GeneratePrivateKey().PublicKey()
	distributions := map[string]uint64{
		creator.String(): 2000, // Above minimum threshold
	}
	dao.InitialTokenDistribution(distributions)

	// Create proposal transaction
	proposalTx := &ProposalTx{
		Fee:          100,
		Title:        "Test Proposal",
		Description:  "This is a test proposal for governance",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    time.Now().Unix() + 3600,  // 1 hour from now
		EndTime:      time.Now().Unix() + 90000, // 25 hours from now
		Threshold:    5100,                      // 51%
		MetadataHash: randomHash(),
	}

	txHash := randomHash()

	err := dao.Processor.ProcessProposalTx(proposalTx, creator, txHash)
	if err != nil {
		t.Fatalf("Failed to process proposal: %v", err)
	}

	// Verify proposal was created
	proposal, err := dao.GetProposal(txHash)
	if err != nil {
		t.Fatalf("Failed to get proposal: %v", err)
	}

	if proposal.Title != "Test Proposal" {
		t.Errorf("Expected title 'Test Proposal', got %s", proposal.Title)
	}

	if proposal.Status != ProposalStatusPending {
		t.Errorf("Expected status pending, got %d", proposal.Status)
	}

	// Verify fee was deducted
	if dao.TokenState.Balances[creator.String()] != 1900 {
		t.Errorf("Expected creator balance 1900, got %d", dao.TokenState.Balances[creator.String()])
	}
}

func TestValidationErrors(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Test insufficient tokens for proposal
	creator := crypto.GeneratePrivateKey().PublicKey()
	distributions := map[string]uint64{
		creator.String(): 500, // Below minimum threshold of 1000
	}
	dao.InitialTokenDistribution(distributions)

	proposalTx := &ProposalTx{
		Fee:          100,
		Title:        "Test Proposal",
		Description:  "This is a test proposal",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    time.Now().Unix() + 3600,
		EndTime:      time.Now().Unix() + 90000,
		Threshold:    5100,
		MetadataHash: randomHash(),
	}

	txHash := randomHash()
	err := dao.Processor.ProcessProposalTx(proposalTx, creator, txHash)
	if err == nil {
		t.Error("Expected error for insufficient tokens")
	}

	daoErr, ok := err.(*DAOError)
	if !ok {
		t.Error("Expected DAOError type")
	}

	if daoErr.Code != ErrInsufficientTokens {
		t.Errorf("Expected error code %d, got %d", ErrInsufficientTokens, daoErr.Code)
	}
}

func TestTokenMinting(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution for minter
	minter := crypto.GeneratePrivateKey().PublicKey()
	recipient := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		minter.String(): 2000, // Enough for fees
	}
	dao.InitialTokenDistribution(distributions)

	// Test token minting
	mintTx := &TokenMintTx{
		Fee:       100,
		Recipient: recipient,
		Amount:    1000,
		Reason:    "Initial distribution",
	}

	err := dao.Processor.ProcessTokenMintTx(mintTx, minter)
	if err != nil {
		t.Fatalf("Failed to mint tokens: %v", err)
	}

	// Verify recipient balance
	if dao.TokenState.Balances[recipient.String()] != 1000 {
		t.Errorf("Expected recipient balance 1000, got %d", dao.TokenState.Balances[recipient.String()])
	}

	// Verify total supply increased
	if dao.TokenState.TotalSupply != 3000 { // 2000 initial + 1000 minted
		t.Errorf("Expected total supply 3000, got %d", dao.TokenState.TotalSupply)
	}

	// Verify minter fee was deducted
	if dao.TokenState.Balances[minter.String()] != 1900 { // 2000 - 100 fee
		t.Errorf("Expected minter balance 1900, got %d", dao.TokenState.Balances[minter.String()])
	}

	// Verify token holder record was created
	holder, exists := dao.GetTokenHolder(recipient)
	if !exists {
		t.Error("Token holder record not created for recipient")
	}
	if holder.Balance != 1000 {
		t.Errorf("Expected holder balance 1000, got %d", holder.Balance)
	}
}

func TestTokenBurning(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	burner := crypto.GeneratePrivateKey().PublicKey()
	distributions := map[string]uint64{
		burner.String(): 2000,
	}
	dao.InitialTokenDistribution(distributions)

	// Test token burning
	burnTx := &TokenBurnTx{
		Fee:    100,
		Amount: 500,
		Reason: "Deflationary mechanism",
	}

	err := dao.Processor.ProcessTokenBurnTx(burnTx, burner)
	if err != nil {
		t.Fatalf("Failed to burn tokens: %v", err)
	}

	// Verify burner balance decreased
	if dao.TokenState.Balances[burner.String()] != 1400 { // 2000 - 500 - 100 fee
		t.Errorf("Expected burner balance 1400, got %d", dao.TokenState.Balances[burner.String()])
	}

	// Verify total supply decreased
	if dao.TokenState.TotalSupply != 1500 { // 2000 - 500 burned
		t.Errorf("Expected total supply 1500, got %d", dao.TokenState.TotalSupply)
	}
}

func TestTokenTransfer(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	sender := crypto.GeneratePrivateKey().PublicKey()
	recipient := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		sender.String(): 2000,
	}
	dao.InitialTokenDistribution(distributions)

	// Test token transfer
	transferTx := &TokenTransferTx{
		Fee:       100,
		Recipient: recipient,
		Amount:    500,
	}

	err := dao.Processor.ProcessTokenTransferTx(transferTx, sender)
	if err != nil {
		t.Fatalf("Failed to transfer tokens: %v", err)
	}

	// Verify sender balance decreased
	if dao.TokenState.Balances[sender.String()] != 1400 { // 2000 - 500 - 100 fee
		t.Errorf("Expected sender balance 1400, got %d", dao.TokenState.Balances[sender.String()])
	}

	// Verify recipient balance increased
	if dao.TokenState.Balances[recipient.String()] != 500 {
		t.Errorf("Expected recipient balance 500, got %d", dao.TokenState.Balances[recipient.String()])
	}

	// Verify total supply unchanged
	if dao.TokenState.TotalSupply != 2000 {
		t.Errorf("Expected total supply 2000, got %d", dao.TokenState.TotalSupply)
	}

	// Verify token holder record was created for recipient
	holder, exists := dao.GetTokenHolder(recipient)
	if !exists {
		t.Error("Token holder record not created for recipient")
	}
	if holder.Balance != 500 {
		t.Errorf("Expected holder balance 500, got %d", holder.Balance)
	}
}

func TestTokenApproval(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	owner := crypto.GeneratePrivateKey().PublicKey()
	spender := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		owner.String(): 2000,
	}
	dao.InitialTokenDistribution(distributions)

	// Test token approval
	approveTx := &TokenApproveTx{
		Fee:     100,
		Spender: spender,
		Amount:  500,
	}

	err := dao.Processor.ProcessTokenApproveTx(approveTx, owner)
	if err != nil {
		t.Fatalf("Failed to approve tokens: %v", err)
	}

	// Verify allowance was set
	allowance := dao.GetTokenAllowance(owner, spender)
	if allowance != 500 {
		t.Errorf("Expected allowance 500, got %d", allowance)
	}

	// Verify owner fee was deducted
	if dao.TokenState.Balances[owner.String()] != 1900 { // 2000 - 100 fee
		t.Errorf("Expected owner balance 1900, got %d", dao.TokenState.Balances[owner.String()])
	}
}

func TestTokenTransferFrom(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	owner := crypto.GeneratePrivateKey().PublicKey()
	spender := crypto.GeneratePrivateKey().PublicKey()
	recipient := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		owner.String():   2000,
		spender.String(): 1000, // For fees
	}
	dao.InitialTokenDistribution(distributions)

	// First approve spender
	dao.ApproveTokens(owner, spender, 500)

	// Test transferFrom
	transferFromTx := &TokenTransferFromTx{
		Fee:       100,
		From:      owner,
		Recipient: recipient,
		Amount:    300,
	}

	err := dao.Processor.ProcessTokenTransferFromTx(transferFromTx, spender)
	if err != nil {
		t.Fatalf("Failed to transferFrom tokens: %v", err)
	}

	// Verify owner balance decreased
	if dao.TokenState.Balances[owner.String()] != 1700 { // 2000 - 300
		t.Errorf("Expected owner balance 1700, got %d", dao.TokenState.Balances[owner.String()])
	}

	// Verify recipient balance increased
	if dao.TokenState.Balances[recipient.String()] != 300 {
		t.Errorf("Expected recipient balance 300, got %d", dao.TokenState.Balances[recipient.String()])
	}

	// Verify spender fee was deducted
	if dao.TokenState.Balances[spender.String()] != 900 { // 1000 - 100 fee
		t.Errorf("Expected spender balance 900, got %d", dao.TokenState.Balances[spender.String()])
	}

	// Verify allowance was reduced
	allowance := dao.GetTokenAllowance(owner, spender)
	if allowance != 200 { // 500 - 300
		t.Errorf("Expected allowance 200, got %d", allowance)
	}
}

func TestTokenTransferValidation(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	sender := crypto.GeneratePrivateKey().PublicKey()
	recipient := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		sender.String(): 500, // Not enough for transfer + fee
	}
	dao.InitialTokenDistribution(distributions)

	// Test insufficient balance
	transferTx := &TokenTransferTx{
		Fee:       100,
		Recipient: recipient,
		Amount:    500, // Would leave nothing for fee
	}

	err := dao.Processor.ProcessTokenTransferTx(transferTx, sender)
	if err == nil {
		t.Error("Expected error for insufficient balance")
	}

	// Test self-transfer
	selfTransferTx := &TokenTransferTx{
		Fee:       100,
		Recipient: sender, // Same as sender
		Amount:    100,
	}

	err = dao.Processor.ProcessTokenTransferTx(selfTransferTx, sender)
	if err == nil {
		t.Error("Expected error for self-transfer")
	}

	// Test zero amount
	zeroTransferTx := &TokenTransferTx{
		Fee:       100,
		Recipient: recipient,
		Amount:    0,
	}

	err = dao.Processor.ProcessTokenTransferTx(zeroTransferTx, sender)
	if err == nil {
		t.Error("Expected error for zero amount transfer")
	}
}

func TestTokenAllowanceValidation(t *testing.T) {
	dao := NewDAO("GOV", "Governance Token", 18)

	// Setup initial distribution
	owner := crypto.GeneratePrivateKey().PublicKey()
	spender := crypto.GeneratePrivateKey().PublicKey()
	recipient := crypto.GeneratePrivateKey().PublicKey()

	distributions := map[string]uint64{
		owner.String():   1000,
		spender.String(): 500,
	}
	dao.InitialTokenDistribution(distributions)

	// Test transferFrom without approval
	transferFromTx := &TokenTransferFromTx{
		Fee:       100,
		From:      owner,
		Recipient: recipient,
		Amount:    300,
	}

	err := dao.Processor.ProcessTokenTransferFromTx(transferFromTx, spender)
	if err == nil {
		t.Error("Expected error for insufficient allowance")
	}

	// Approve small amount
	dao.ApproveTokens(owner, spender, 200)

	// Test transferFrom exceeding allowance
	err = dao.Processor.ProcessTokenTransferFromTx(transferFromTx, spender)
	if err == nil {
		t.Error("Expected error for exceeding allowance")
	}
}
func TestTokenSystemIntegration(t *testing.T) {
	// This test runs the complete token example to ensure all functionality works together
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Token example panicked: %v", r)
		}
	}()

	// Run the token example (this will panic if anything fails)
	TokenExample()

	// If we get here, the example ran successfully
	t.Log("Token system integration test passed")
}
