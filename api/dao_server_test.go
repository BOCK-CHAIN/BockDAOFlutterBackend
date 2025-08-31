package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/core"
	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/dao"
	"github.com/BOCK-CHAIN/BockChain/types"
	"github.com/go-kit/log"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDAOServer() (*DAOServer, *dao.DAO, chan *core.Transaction) {
	// Create test blockchain
	bc := &core.Blockchain{} // Simplified for testing

	// Create test DAO
	testDAO := dao.NewDAO("TEST", "Test Token", 18)

	// Create transaction channel
	txChan := make(chan *core.Transaction, 100)

	// Create server config
	cfg := ServerConfig{
		Logger:     log.NewNopLogger(),
		ListenAddr: ":0", // Use random port for testing
	}

	// Create DAO server
	server := NewDAOServer(cfg, bc, txChan, testDAO)

	return server, testDAO, txChan
}

func TestDAOServer_GetProposals(t *testing.T) {
	server, testDAO, _ := setupTestDAOServer()

	// Create test proposal
	privKey := crypto.GeneratePrivateKey()
	proposalID := types.Hash{1, 2, 3} // Simple test hash

	proposal := &dao.Proposal{
		ID:           proposalID,
		Creator:      privKey.PublicKey(),
		Title:        "Test Proposal",
		Description:  "Test Description",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix(),
		EndTime:      time.Now().Unix() + 3600,
		Status:       dao.ProposalStatusActive,
		Threshold:    1000,
		Results:      nil,
		MetadataHash: types.Hash{},
	}

	testDAO.GovernanceState.Proposals[proposalID] = proposal

	// Create test request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/dao/proposals", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := server.handleGetProposals(c)
	require.NoError(t, err)

	// Check response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response []ProposalResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response, 1)
	assert.Equal(t, "Test Proposal", response[0].Title)
	assert.Equal(t, "Test Description", response[0].Description)
}

func TestDAOServer_GetProposal(t *testing.T) {
	server, testDAO, _ := setupTestDAOServer()

	// Create test proposal
	privKey := crypto.GeneratePrivateKey()
	proposalID := types.Hash{1, 2, 3}

	proposal := &dao.Proposal{
		ID:           proposalID,
		Creator:      privKey.PublicKey(),
		Title:        "Test Proposal",
		Description:  "Test Description",
		ProposalType: dao.ProposalTypeGeneral,
		VotingType:   dao.VotingTypeSimple,
		StartTime:    time.Now().Unix(),
		EndTime:      time.Now().Unix() + 3600,
		Status:       dao.ProposalStatusActive,
		Threshold:    1000,
		Results:      nil,
		MetadataHash: types.Hash{},
	}

	testDAO.GovernanceState.Proposals[proposalID] = proposal

	// Create test request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/dao/proposal/"+proposalID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(proposalID.String())

	// Execute handler
	err := server.handleGetProposal(c)
	require.NoError(t, err)

	// Check response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response ProposalResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Test Proposal", response.Title)
	assert.Equal(t, "Test Description", response.Description)
	assert.Equal(t, proposalID.String(), response.ID)
}

func TestDAOServer_CreateProposal(t *testing.T) {
	server, testDAO, txChan := setupTestDAOServer()

	// Initialize test token distribution
	privKey := crypto.GeneratePrivateKey()
	testDAO.InitialTokenDistribution(map[string]uint64{
		privKey.PublicKey().String(): 10000,
	})

	// Create test request
	reqBody := map[string]interface{}{
		"title":         "Test Proposal",
		"description":   "Test Description",
		"proposal_type": dao.ProposalTypeGeneral,
		"voting_type":   dao.VotingTypeSimple,
		"duration":      3600,
		"threshold":     1000,
		"metadata_hash": "",
		"private_key":   hex.EncodeToString([]byte("test_private_key_32_bytes_long!!")), // 32 bytes
	}

	reqJSON, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/dao/proposal", bytes.NewReader(reqJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := server.handleCreateProposal(c)
	require.NoError(t, err)

	// Check response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check that transaction was sent
	select {
	case tx := <-txChan:
		assert.NotNil(t, tx)
		assert.NotNil(t, tx.TxInner)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected transaction to be sent to channel")
	}
}

func TestDAOServer_GetTreasury(t *testing.T) {
	server, testDAO, _ := setupTestDAOServer()

	// Initialize treasury
	privKey1 := crypto.GeneratePrivateKey()
	privKey2 := crypto.GeneratePrivateKey()
	signers := []crypto.PublicKey{privKey1.PublicKey(), privKey2.PublicKey()}

	err := testDAO.InitializeTreasury(signers, 2)
	require.NoError(t, err)

	testDAO.AddTreasuryFunds(50000)

	// Create test request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/dao/treasury", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err = server.handleGetTreasury(c)
	require.NoError(t, err)

	// Check response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response TreasuryResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, uint64(50000), response.Balance)
	assert.Equal(t, uint8(2), response.RequiredSigs)
	assert.Len(t, response.Signers, 2)
}

func TestDAOServer_GetTokenBalance(t *testing.T) {
	server, testDAO, _ := setupTestDAOServer()

	// Initialize test token distribution
	privKey := crypto.GeneratePrivateKey()
	testDAO.InitialTokenDistribution(map[string]uint64{
		privKey.PublicKey().String(): 10000,
	})

	// Create test request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/dao/token/balance/"+privKey.PublicKey().String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("address")
	c.SetParamValues(privKey.PublicKey().String())

	// Execute handler
	err := server.handleGetTokenBalance(c)
	require.NoError(t, err)

	// Check response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]uint64
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, uint64(10000), response["balance"])
}

func TestDAOServer_GetTokenSupply(t *testing.T) {
	server, testDAO, _ := setupTestDAOServer()

	// Initialize test token distribution
	privKey := crypto.GeneratePrivateKey()
	testDAO.InitialTokenDistribution(map[string]uint64{
		privKey.PublicKey().String(): 10000,
	})

	// Create test request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/dao/token/supply", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := server.handleGetTokenSupply(c)
	require.NoError(t, err)

	// Check response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]uint64
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, uint64(10000), response["total_supply"])
}

func TestDAOServer_WebSocketConnection(t *testing.T) {
	server, _, _ := setupTestDAOServer()

	// This is a simplified test for WebSocket setup
	// In a real test, you'd use gorilla/websocket test utilities
	assert.NotNil(t, server.eventBus)
	assert.NotNil(t, server.upgrader)
}

func TestDAOServer_EventBroadcast(t *testing.T) {
	server, _, _ := setupTestDAOServer()

	// Test event creation and broadcasting
	event := Event{
		Type: EventProposalCreated,
		Data: map[string]interface{}{
			"title":   "Test Proposal",
			"creator": "test_creator",
		},
		Timestamp: time.Now().Unix(),
	}

	// This should not panic
	server.broadcastEvent(event)
}

// Integration test for complete proposal flow
func TestDAOServer_ProposalFlow(t *testing.T) {
	server, testDAO, txChan := setupTestDAOServer()

	// Initialize test environment
	privKey := crypto.GeneratePrivateKey()
	testDAO.InitialTokenDistribution(map[string]uint64{
		privKey.PublicKey().String(): 10000,
	})

	// 1. Create proposal
	reqBody := map[string]interface{}{
		"title":         "Integration Test Proposal",
		"description":   "Test Description for Integration",
		"proposal_type": dao.ProposalTypeGeneral,
		"voting_type":   dao.VotingTypeSimple,
		"duration":      3600,
		"threshold":     1000,
		"metadata_hash": "",
		"private_key":   hex.EncodeToString([]byte("test_private_key_32_bytes_long!!")),
	}

	reqJSON, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/dao/proposal", bytes.NewReader(reqJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := server.handleCreateProposal(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify transaction was created
	select {
	case tx := <-txChan:
		assert.NotNil(t, tx)
		proposalTx, ok := tx.TxInner.(*dao.ProposalTx)
		assert.True(t, ok)
		assert.Equal(t, "Integration Test Proposal", proposalTx.Title)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected transaction to be sent to channel")
	}

	// 2. Get proposals (should include the one we just created)
	req2 := httptest.NewRequest(http.MethodGet, "/dao/proposals", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = server.handleGetProposals(c2)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	// Note: In a real integration test, we'd process the transaction through the DAO
	// and then verify the proposal appears in the list
}
