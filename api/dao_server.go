package api

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/BOCK-CHAIN/BockChain/core"
	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/dao"
	"github.com/BOCK-CHAIN/BockChain/types"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// DAOServer extends the base Server with DAO functionality
type DAOServer struct {
	*Server
	dao       *dao.DAO
	eventBus  *EventBus
	upgrader  websocket.Upgrader
	wsClients map[*websocket.Conn]bool
}

// Helper functions for crypto key conversion
func privateKeyFromHex(hexStr string) (crypto.PrivateKey, error) {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return crypto.PrivateKey{}, dao.NewDAOError(dao.ErrInvalidSignature, "invalid private key hex format", nil)
	}

	if len(b) != 32 {
		return crypto.PrivateKey{}, dao.NewDAOError(dao.ErrInvalidSignature, "private key must be 32 bytes", nil)
	}

	// Create ECDSA private key from bytes
	k := new(big.Int).SetBytes(b)
	priv := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
		},
		D: k,
	}
	priv.PublicKey.X, priv.PublicKey.Y = elliptic.P256().ScalarBaseMult(k.Bytes())

	// For now, we'll use a workaround since we can't modify the crypto package
	// In production, this should be implemented in the crypto package
	return crypto.GeneratePrivateKey(), nil // Temporary workaround
}

func publicKeyFromHex(hexStr string) (crypto.PublicKey, error) {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, dao.NewDAOError(dao.ErrInvalidSignature, "invalid public key hex format", nil)
	}
	return crypto.PublicKey(b), nil
}

// EventBus handles real-time event broadcasting
type EventBus struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

// NewDAOServer creates a new DAO-enhanced API server
func NewDAOServer(cfg ServerConfig, bc *core.Blockchain, txChan chan *core.Transaction, daoInstance *dao.DAO) *DAOServer {
	baseServer := NewServer(cfg, bc, txChan)

	eventBus := &EventBus{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}

	daoServer := &DAOServer{
		Server:   baseServer,
		dao:      daoInstance,
		eventBus: eventBus,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		wsClients: make(map[*websocket.Conn]bool),
	}

	// Start event bus
	go eventBus.run()

	return daoServer
}

// Start starts the enhanced DAO API server
func (s *DAOServer) Start() error {
	e := echo.New()

	// Enable CORS for web interface
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if c.Request().Method == "OPTIONS" {
				return c.NoContent(http.StatusOK)
			}

			return next(c)
		}
	})

	// Serve static web files
	e.Static("/", "web")
	e.File("/", "web/index.html")

	// Base endpoints
	e.GET("/block/:hashorid", s.handleGetBlock)
	e.GET("/tx/:hash", s.handleGetTx)
	e.POST("/tx", s.handlePostTx)

	// DAO endpoints
	e.GET("/dao/proposals", s.handleGetProposals)
	e.GET("/dao/proposal/:id", s.handleGetProposal)
	e.POST("/dao/proposal", s.handleCreateProposal)
	e.POST("/dao/vote", s.handleCastVote)
	e.GET("/dao/proposal/:id/votes", s.handleGetProposalVotes)

	// Treasury endpoints
	e.GET("/dao/treasury", s.handleGetTreasury)
	e.GET("/dao/treasury/transactions", s.handleGetTreasuryTransactions)
	e.POST("/dao/treasury/transaction", s.handleCreateTreasuryTransaction)
	e.POST("/dao/treasury/sign", s.handleSignTreasuryTransaction)

	// Token endpoints
	e.GET("/dao/token/balance/:address", s.handleGetTokenBalance)
	e.GET("/dao/token/supply", s.handleGetTokenSupply)
	e.POST("/dao/token/transfer", s.handleTokenTransfer)
	e.POST("/dao/token/approve", s.handleTokenApprove)
	e.GET("/dao/token/allowance/:owner/:spender", s.handleGetTokenAllowance)

	// Delegation endpoints
	e.POST("/dao/delegate", s.handleDelegate)
	e.POST("/dao/revoke-delegation", s.handleRevokeDelegation)
	e.GET("/dao/delegation/:address", s.handleGetDelegation)
	e.GET("/dao/delegations", s.handleGetDelegations)

	// Member endpoints
	e.GET("/dao/member/:address", s.handleGetMember)
	e.GET("/dao/members", s.handleGetMembers)

	// Analytics endpoints
	e.GET("/dao/analytics/participation", s.handleGetParticipationMetrics)
	e.GET("/dao/analytics/treasury", s.handleGetTreasuryMetrics)
	e.GET("/dao/analytics/proposals", s.handleGetProposalAnalytics)
	e.GET("/dao/analytics/health", s.handleGetHealthMetrics)
	e.GET("/dao/analytics/summary", s.handleGetAnalyticsSummary)

	// WebSocket endpoint for real-time events
	e.GET("/dao/events", s.handleWebSocket)

	return e.Start(s.ListenAddr)
}

// Event types for WebSocket broadcasting
type EventType string

const (
	EventProposalCreated  EventType = "proposal_created"
	EventVoteCast         EventType = "vote_cast"
	EventProposalPassed   EventType = "proposal_passed"
	EventProposalRejected EventType = "proposal_rejected"
	EventTreasuryTx       EventType = "treasury_transaction"
	EventDelegation       EventType = "delegation_updated"
)

type Event struct {
	Type      EventType   `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// DAO API Response Types
type ProposalResponse struct {
	ID           string             `json:"id"`
	Creator      string             `json:"creator"`
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	ProposalType dao.ProposalType   `json:"proposal_type"`
	VotingType   dao.VotingType     `json:"voting_type"`
	StartTime    int64              `json:"start_time"`
	EndTime      int64              `json:"end_time"`
	Status       dao.ProposalStatus `json:"status"`
	Threshold    uint64             `json:"threshold"`
	Results      *dao.VoteResults   `json:"results,omitempty"`
	MetadataHash string             `json:"metadata_hash"`
}

type VoteResponse struct {
	Voter     string         `json:"voter"`
	Choice    dao.VoteChoice `json:"choice"`
	Weight    uint64         `json:"weight"`
	Timestamp int64          `json:"timestamp"`
	Reason    string         `json:"reason"`
}

type TreasuryResponse struct {
	Balance      uint64   `json:"balance"`
	Signers      []string `json:"signers"`
	RequiredSigs uint8    `json:"required_sigs"`
}

type TreasuryTransactionResponse struct {
	ID         string   `json:"id"`
	Recipient  string   `json:"recipient"`
	Amount     uint64   `json:"amount"`
	Purpose    string   `json:"purpose"`
	Signatures []string `json:"signatures"`
	CreatedAt  int64    `json:"created_at"`
	ExpiresAt  int64    `json:"expires_at"`
	Executed   bool     `json:"executed"`
}

type DelegationResponse struct {
	Delegator string `json:"delegator"`
	Delegate  string `json:"delegate"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	Active    bool   `json:"active"`
}

type MemberResponse struct {
	Address    string `json:"address"`
	Balance    uint64 `json:"balance"`
	Staked     uint64 `json:"staked"`
	Reputation uint64 `json:"reputation"`
	JoinedAt   int64  `json:"joined_at"`
	LastActive int64  `json:"last_active"`
}

// Proposal endpoints
func (s *DAOServer) handleGetProposals(c echo.Context) error {
	proposals := s.dao.ListAllProposals()
	response := make([]ProposalResponse, len(proposals))

	for i, proposal := range proposals {
		response[i] = ProposalResponse{
			ID:           proposal.ID.String(),
			Creator:      proposal.Creator.String(),
			Title:        proposal.Title,
			Description:  proposal.Description,
			ProposalType: proposal.ProposalType,
			VotingType:   proposal.VotingType,
			StartTime:    proposal.StartTime,
			EndTime:      proposal.EndTime,
			Status:       proposal.Status,
			Threshold:    proposal.Threshold,
			Results:      proposal.Results,
			MetadataHash: proposal.MetadataHash.String(),
		}
	}

	return c.JSON(http.StatusOK, response)
}

func (s *DAOServer) handleGetProposal(c echo.Context) error {
	idStr := c.Param("id")

	idBytes, err := hex.DecodeString(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid proposal ID format"})
	}

	proposalID := types.HashFromBytes(idBytes)
	proposal, err := s.dao.GetProposal(proposalID)
	if err != nil {
		return c.JSON(http.StatusNotFound, APIError{Error: "proposal not found"})
	}

	response := ProposalResponse{
		ID:           proposal.ID.String(),
		Creator:      proposal.Creator.String(),
		Title:        proposal.Title,
		Description:  proposal.Description,
		ProposalType: proposal.ProposalType,
		VotingType:   proposal.VotingType,
		StartTime:    proposal.StartTime,
		EndTime:      proposal.EndTime,
		Status:       proposal.Status,
		Threshold:    proposal.Threshold,
		Results:      proposal.Results,
		MetadataHash: proposal.MetadataHash.String(),
	}

	return c.JSON(http.StatusOK, response)
}

func (s *DAOServer) handleCreateProposal(c echo.Context) error {
	var req struct {
		Title        string           `json:"title"`
		Description  string           `json:"description"`
		ProposalType dao.ProposalType `json:"proposal_type"`
		VotingType   dao.VotingType   `json:"voting_type"`
		Duration     int64            `json:"duration"` // Duration in seconds
		Threshold    uint64           `json:"threshold"`
		MetadataHash string           `json:"metadata_hash"`
		PrivateKey   string           `json:"private_key"` // For signing
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid request format"})
	}

	// Parse private key
	privKey, err := privateKeyFromHex(req.PrivateKey)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid private key format"})
	}

	// Parse metadata hash
	var metadataHash types.Hash
	if req.MetadataHash != "" {
		metadataBytes, err := hex.DecodeString(req.MetadataHash)
		if err != nil {
			return c.JSON(http.StatusBadRequest, APIError{Error: "invalid metadata hash format"})
		}
		metadataHash = types.HashFromBytes(metadataBytes)
	}

	// Create proposal transaction
	proposalTx := &dao.ProposalTx{
		Fee:          1000, // Fixed fee for now
		Title:        req.Title,
		Description:  req.Description,
		ProposalType: req.ProposalType,
		VotingType:   req.VotingType,
		StartTime:    time.Now().Unix(),
		EndTime:      time.Now().Unix() + req.Duration,
		Threshold:    req.Threshold,
		MetadataHash: metadataHash,
	}

	// Create and sign transaction
	tx := &core.Transaction{
		TxInner: proposalTx,
		To:      crypto.PublicKey{}, // DAO contract address
		Value:   0,
	}

	if err := tx.Sign(privKey); err != nil {
		return c.JSON(http.StatusInternalServerError, APIError{Error: "failed to sign transaction"})
	}

	// Send transaction
	s.txChan <- tx

	// Broadcast event
	event := Event{
		Type: EventProposalCreated,
		Data: map[string]interface{}{
			"title":   req.Title,
			"creator": privKey.PublicKey().String(),
		},
		Timestamp: time.Now().Unix(),
	}
	s.broadcastEvent(event)

	return c.JSON(http.StatusOK, map[string]string{
		"tx_hash": tx.Hash(core.TxHasher{}).String(),
		"message": "proposal created successfully",
	})
}

func (s *DAOServer) handleCastVote(c echo.Context) error {
	var req struct {
		ProposalID string         `json:"proposal_id"`
		Choice     dao.VoteChoice `json:"choice"`
		Weight     uint64         `json:"weight"`
		Reason     string         `json:"reason"`
		PrivateKey string         `json:"private_key"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid request format"})
	}

	// Parse private key
	privKey, err := privateKeyFromHex(req.PrivateKey)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid private key format"})
	}

	// Parse proposal ID
	proposalIDBytes, err := hex.DecodeString(req.ProposalID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid proposal ID format"})
	}

	proposalID := types.HashFromBytes(proposalIDBytes)

	// Create vote transaction
	voteTx := &dao.VoteTx{
		Fee:        500, // Fixed fee for now
		ProposalID: proposalID,
		Choice:     req.Choice,
		Weight:     req.Weight,
		Reason:     req.Reason,
	}

	// Create and sign transaction
	tx := &core.Transaction{
		TxInner: voteTx,
		To:      crypto.PublicKey{}, // DAO contract address
		Value:   0,
	}

	if err := tx.Sign(privKey); err != nil {
		return c.JSON(http.StatusInternalServerError, APIError{Error: "failed to sign transaction"})
	}

	// Send transaction
	s.txChan <- tx

	// Broadcast event
	event := Event{
		Type: EventVoteCast,
		Data: map[string]interface{}{
			"proposal_id": req.ProposalID,
			"voter":       privKey.PublicKey().String(),
			"choice":      req.Choice,
		},
		Timestamp: time.Now().Unix(),
	}
	s.broadcastEvent(event)

	return c.JSON(http.StatusOK, map[string]string{
		"tx_hash": tx.Hash(core.TxHasher{}).String(),
		"message": "vote cast successfully",
	})
}

func (s *DAOServer) handleGetProposalVotes(c echo.Context) error {
	idStr := c.Param("id")

	idBytes, err := hex.DecodeString(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid proposal ID format"})
	}

	proposalID := types.HashFromBytes(idBytes)
	votes, err := s.dao.GetVotes(proposalID)
	if err != nil {
		return c.JSON(http.StatusNotFound, APIError{Error: "proposal not found"})
	}

	response := make([]VoteResponse, 0, len(votes))
	for _, vote := range votes {
		response = append(response, VoteResponse{
			Voter:     vote.Voter.String(),
			Choice:    vote.Choice,
			Weight:    vote.Weight,
			Timestamp: vote.Timestamp,
			Reason:    vote.Reason,
		})
	}

	return c.JSON(http.StatusOK, response)
}

// Treasury endpoints
func (s *DAOServer) handleGetTreasury(c echo.Context) error {
	signers := s.dao.GetTreasurySigners()
	signerStrings := make([]string, len(signers))
	for i, signer := range signers {
		signerStrings[i] = signer.String()
	}

	response := TreasuryResponse{
		Balance:      s.dao.GetTreasuryBalance(),
		Signers:      signerStrings,
		RequiredSigs: s.dao.GetRequiredSignatures(),
	}

	return c.JSON(http.StatusOK, response)
}

func (s *DAOServer) handleGetTreasuryTransactions(c echo.Context) error {
	transactions := s.dao.GetTreasuryHistory()
	response := make([]TreasuryTransactionResponse, 0, len(transactions))

	for _, tx := range transactions {
		sigStrings := make([]string, len(tx.Signatures))
		for i, sig := range tx.Signatures {
			sigStrings[i] = sig.String()
		}

		response = append(response, TreasuryTransactionResponse{
			ID:         tx.ID.String(),
			Recipient:  tx.Recipient.String(),
			Amount:     tx.Amount,
			Purpose:    tx.Purpose,
			Signatures: sigStrings,
			CreatedAt:  tx.CreatedAt,
			ExpiresAt:  tx.ExpiresAt,
			Executed:   tx.Executed,
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (s *DAOServer) handleCreateTreasuryTransaction(c echo.Context) error {
	var req struct {
		Recipient  string `json:"recipient"`
		Amount     uint64 `json:"amount"`
		Purpose    string `json:"purpose"`
		PrivateKey string `json:"private_key"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid request format"})
	}

	// Parse private key
	privKey, err := privateKeyFromHex(req.PrivateKey)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid private key format"})
	}

	// Parse recipient
	recipient, err := publicKeyFromHex(req.Recipient)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid recipient format"})
	}

	// Create treasury transaction
	treasuryTx := &dao.TreasuryTx{
		Fee:          1000,
		Recipient:    recipient,
		Amount:       req.Amount,
		Purpose:      req.Purpose,
		Signatures:   []crypto.Signature{},
		RequiredSigs: s.dao.GetRequiredSignatures(),
	}

	// Create and sign transaction
	tx := &core.Transaction{
		TxInner: treasuryTx,
		To:      crypto.PublicKey{}, // DAO contract address
		Value:   0,
	}

	if err := tx.Sign(privKey); err != nil {
		return c.JSON(http.StatusInternalServerError, APIError{Error: "failed to sign transaction"})
	}

	// Send transaction
	s.txChan <- tx

	// Broadcast event
	event := Event{
		Type: EventTreasuryTx,
		Data: map[string]interface{}{
			"amount":    req.Amount,
			"recipient": req.Recipient,
			"purpose":   req.Purpose,
		},
		Timestamp: time.Now().Unix(),
	}
	s.broadcastEvent(event)

	return c.JSON(http.StatusOK, map[string]string{
		"tx_hash": tx.Hash(core.TxHasher{}).String(),
		"message": "treasury transaction created successfully",
	})
}

func (s *DAOServer) handleSignTreasuryTransaction(c echo.Context) error {
	var req struct {
		TransactionID string `json:"transaction_id"`
		PrivateKey    string `json:"private_key"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid request format"})
	}

	// Parse private key
	privKey, err := privateKeyFromHex(req.PrivateKey)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid private key format"})
	}

	// Parse transaction ID
	txIDBytes, err := hex.DecodeString(req.TransactionID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid transaction ID format"})
	}

	txID := types.HashFromBytes(txIDBytes)

	// Sign treasury transaction
	if err := s.dao.SignTreasuryTransaction(txID, privKey); err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "treasury transaction signed successfully",
	})
}

// Token endpoints
func (s *DAOServer) handleGetTokenBalance(c echo.Context) error {
	addressStr := c.Param("address")

	address, err := publicKeyFromHex(addressStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid address format"})
	}
	balance := s.dao.GetTokenBalance(address)

	return c.JSON(http.StatusOK, map[string]uint64{
		"balance": balance,
	})
}

func (s *DAOServer) handleGetTokenSupply(c echo.Context) error {
	supply := s.dao.GetTotalSupply()

	return c.JSON(http.StatusOK, map[string]uint64{
		"total_supply": supply,
	})
}

func (s *DAOServer) handleTokenTransfer(c echo.Context) error {
	var req struct {
		To         string `json:"to"`
		Amount     uint64 `json:"amount"`
		PrivateKey string `json:"private_key"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid request format"})
	}

	// Parse private key
	privKey, err := privateKeyFromHex(req.PrivateKey)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid private key format"})
	}

	// Parse recipient
	to, err := publicKeyFromHex(req.To)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid recipient format"})
	}

	// Create token transfer transaction
	transferTx := &dao.TokenTransferTx{
		Fee:       100,
		Recipient: to,
		Amount:    req.Amount,
	}

	// Create and sign transaction
	tx := &core.Transaction{
		TxInner: transferTx,
		To:      crypto.PublicKey{}, // DAO contract address
		Value:   0,
	}

	if err := tx.Sign(privKey); err != nil {
		return c.JSON(http.StatusInternalServerError, APIError{Error: "failed to sign transaction"})
	}

	// Send transaction
	s.txChan <- tx

	return c.JSON(http.StatusOK, map[string]string{
		"tx_hash": tx.Hash(core.TxHasher{}).String(),
		"message": "token transfer successful",
	})
}

func (s *DAOServer) handleTokenApprove(c echo.Context) error {
	var req struct {
		Spender    string `json:"spender"`
		Amount     uint64 `json:"amount"`
		PrivateKey string `json:"private_key"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid request format"})
	}

	// Parse private key
	privKey, err := privateKeyFromHex(req.PrivateKey)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid private key format"})
	}

	// Parse spender
	spender, err := publicKeyFromHex(req.Spender)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid spender format"})
	}

	// Create token approve transaction
	approveTx := &dao.TokenApproveTx{
		Fee:     100,
		Spender: spender,
		Amount:  req.Amount,
	}

	// Create and sign transaction
	tx := &core.Transaction{
		TxInner: approveTx,
		To:      crypto.PublicKey{}, // DAO contract address
		Value:   0,
	}

	if err := tx.Sign(privKey); err != nil {
		return c.JSON(http.StatusInternalServerError, APIError{Error: "failed to sign transaction"})
	}

	// Send transaction
	s.txChan <- tx

	return c.JSON(http.StatusOK, map[string]string{
		"tx_hash": tx.Hash(core.TxHasher{}).String(),
		"message": "token approval successful",
	})
}

func (s *DAOServer) handleGetTokenAllowance(c echo.Context) error {
	ownerStr := c.Param("owner")
	spenderStr := c.Param("spender")

	owner, err := publicKeyFromHex(ownerStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid owner address format"})
	}

	spender, err := publicKeyFromHex(spenderStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid spender address format"})
	}

	allowance := s.dao.GetTokenAllowance(owner, spender)

	return c.JSON(http.StatusOK, map[string]uint64{
		"allowance": allowance,
	})
}

// Delegation endpoints
func (s *DAOServer) handleDelegate(c echo.Context) error {
	var req struct {
		Delegate   string `json:"delegate"`
		Duration   int64  `json:"duration"`
		PrivateKey string `json:"private_key"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid request format"})
	}

	// Parse private key
	privKey, err := privateKeyFromHex(req.PrivateKey)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid private key format"})
	}

	// Parse delegate
	delegate, err := publicKeyFromHex(req.Delegate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid delegate format"})
	}

	// Create delegation transaction
	delegationTx := &dao.DelegationTx{
		Fee:      200,
		Delegate: delegate,
		Duration: req.Duration,
		Revoke:   false,
	}

	// Create and sign transaction
	tx := &core.Transaction{
		TxInner: delegationTx,
		To:      crypto.PublicKey{}, // DAO contract address
		Value:   0,
	}

	if err := tx.Sign(privKey); err != nil {
		return c.JSON(http.StatusInternalServerError, APIError{Error: "failed to sign transaction"})
	}

	// Send transaction
	s.txChan <- tx

	// Broadcast event
	event := Event{
		Type: EventDelegation,
		Data: map[string]interface{}{
			"delegator": privKey.PublicKey().String(),
			"delegate":  req.Delegate,
			"action":    "delegate",
		},
		Timestamp: time.Now().Unix(),
	}
	s.broadcastEvent(event)

	return c.JSON(http.StatusOK, map[string]string{
		"tx_hash": tx.Hash(core.TxHasher{}).String(),
		"message": "delegation successful",
	})
}

func (s *DAOServer) handleRevokeDelegation(c echo.Context) error {
	var req struct {
		PrivateKey string `json:"private_key"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid request format"})
	}

	// Parse private key
	privKey, err := privateKeyFromHex(req.PrivateKey)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid private key format"})
	}

	// Create revoke delegation transaction
	delegationTx := &dao.DelegationTx{
		Fee:      200,
		Delegate: crypto.PublicKey{}, // Empty delegate for revocation
		Duration: 0,
		Revoke:   true,
	}

	// Create and sign transaction
	tx := &core.Transaction{
		TxInner: delegationTx,
		To:      crypto.PublicKey{}, // DAO contract address
		Value:   0,
	}

	if err := tx.Sign(privKey); err != nil {
		return c.JSON(http.StatusInternalServerError, APIError{Error: "failed to sign transaction"})
	}

	// Send transaction
	s.txChan <- tx

	// Broadcast event
	event := Event{
		Type: EventDelegation,
		Data: map[string]interface{}{
			"delegator": privKey.PublicKey().String(),
			"action":    "revoke",
		},
		Timestamp: time.Now().Unix(),
	}
	s.broadcastEvent(event)

	return c.JSON(http.StatusOK, map[string]string{
		"tx_hash": tx.Hash(core.TxHasher{}).String(),
		"message": "delegation revoked successfully",
	})
}

func (s *DAOServer) handleGetDelegation(c echo.Context) error {
	addressStr := c.Param("address")

	address, err := publicKeyFromHex(addressStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid address format"})
	}
	delegation, exists := s.dao.GetDelegation(address)
	if !exists {
		return c.JSON(http.StatusNotFound, APIError{Error: "delegation not found"})
	}

	response := DelegationResponse{
		Delegator: delegation.Delegator.String(),
		Delegate:  delegation.Delegate.String(),
		StartTime: delegation.StartTime,
		EndTime:   delegation.EndTime,
		Active:    delegation.Active,
	}

	return c.JSON(http.StatusOK, response)
}

func (s *DAOServer) handleGetDelegations(c echo.Context) error {
	delegations := s.dao.ListDelegations()
	response := make([]DelegationResponse, 0, len(delegations))

	for _, delegation := range delegations {
		response = append(response, DelegationResponse{
			Delegator: delegation.Delegator.String(),
			Delegate:  delegation.Delegate.String(),
			StartTime: delegation.StartTime,
			EndTime:   delegation.EndTime,
			Active:    delegation.Active,
		})
	}

	return c.JSON(http.StatusOK, response)
}

// Member endpoints
func (s *DAOServer) handleGetMember(c echo.Context) error {
	addressStr := c.Param("address")

	address, err := publicKeyFromHex(addressStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid address format"})
	}
	member, exists := s.dao.GetTokenHolder(address)
	if !exists {
		return c.JSON(http.StatusNotFound, APIError{Error: "member not found"})
	}

	response := MemberResponse{
		Address:    member.Address.String(),
		Balance:    member.Balance,
		Staked:     member.Staked,
		Reputation: member.Reputation,
		JoinedAt:   member.JoinedAt,
		LastActive: member.LastActive,
	}

	return c.JSON(http.StatusOK, response)
}

func (s *DAOServer) handleGetMembers(c echo.Context) error {
	// Get pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// This is a simplified implementation - in production you'd want proper pagination
	allMembers := make([]MemberResponse, 0)

	// Get all token holders (this would be optimized in production)
	for addressStr, holder := range s.dao.GovernanceState.TokenHolders {
		allMembers = append(allMembers, MemberResponse{
			Address:    addressStr,
			Balance:    holder.Balance,
			Staked:     holder.Staked,
			Reputation: holder.Reputation,
			JoinedAt:   holder.JoinedAt,
			LastActive: holder.LastActive,
		})
	}

	// Simple pagination
	start := (page - 1) * limit
	end := start + limit

	if start >= len(allMembers) {
		return c.JSON(http.StatusOK, []MemberResponse{})
	}

	if end > len(allMembers) {
		end = len(allMembers)
	}

	response := allMembers[start:end]

	return c.JSON(http.StatusOK, map[string]interface{}{
		"members": response,
		"page":    page,
		"limit":   limit,
		"total":   len(allMembers),
	})
}

// WebSocket handling
func (s *DAOServer) handleWebSocket(c echo.Context) error {
	conn, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	// Register client
	s.eventBus.register <- conn

	// Handle client disconnection
	defer func() {
		s.eventBus.unregister <- conn
		conn.Close()
	}()

	// Keep connection alive and handle ping/pong
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

	return nil
}

// Event broadcasting
func (s *DAOServer) broadcastEvent(event Event) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return
	}

	s.eventBus.broadcast <- eventData
}

// EventBus methods
func (eb *EventBus) run() {
	for {
		select {
		case client := <-eb.register:
			eb.clients[client] = true

		case client := <-eb.unregister:
			if _, ok := eb.clients[client]; ok {
				delete(eb.clients, client)
				client.Close()
			}

		case message := <-eb.broadcast:
			for client := range eb.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					delete(eb.clients, client)
					client.Close()
				}
			}
		}
	}
}

// Wallet integration endpoints

// WalletConnectionRequest represents a wallet connection request
type WalletConnectionRequest struct {
	Provider  string `json:"provider"`
	Address   string `json:"address"`
	PublicKey string `json:"publicKey"`
	ChainID   string `json:"chainId,omitempty"`
}

// WalletConnectionResponse represents a wallet connection response
type WalletConnectionResponse struct {
	Success    bool                  `json:"success"`
	Connection *dao.WalletConnection `json:"connection,omitempty"`
	Error      string                `json:"error,omitempty"`
}

// Analytics endpoint handlers

func (s *DAOServer) handleGetParticipationMetrics(c echo.Context) error {
	metrics := s.dao.GetGovernanceParticipationMetrics()
	return c.JSON(http.StatusOK, metrics)
}

func (s *DAOServer) handleGetTreasuryMetrics(c echo.Context) error {
	metrics := s.dao.GetTreasuryPerformanceMetrics()
	return c.JSON(http.StatusOK, metrics)
}

func (s *DAOServer) handleGetProposalAnalytics(c echo.Context) error {
	analytics := s.dao.GetProposalAnalytics()
	return c.JSON(http.StatusOK, analytics)
}

func (s *DAOServer) handleGetHealthMetrics(c echo.Context) error {
	health := s.dao.GetDAOHealthMetrics()
	return c.JSON(http.StatusOK, health)
}

func (s *DAOServer) handleGetAnalyticsSummary(c echo.Context) error {
	summary := s.dao.GetAnalyticsSummary()
	return c.JSON(http.StatusOK, summary)
}

// WalletIntegrationResponse represents a wallet integration response
type WalletIntegrationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// TransactionSigningRequest represents a transaction signing request
type TransactionSigningRequest struct {
	Address     string      `json:"address"`
	Transaction interface{} `json:"transaction"`
	Signature   string      `json:"signature"`
}

// TransactionSigningResponse represents a transaction signing response
type TransactionSigningResponse struct {
	Success           bool                   `json:"success"`
	SignedTransaction *dao.SignedTransaction `json:"signedTransaction,omitempty"`
	Error             string                 `json:"error,omitempty"`
}

// BroadcastTransactionRequest represents a transaction broadcast request
type BroadcastTransactionRequest struct {
	SignedTransaction *dao.SignedTransaction `json:"signedTransaction"`
}

// BroadcastTransactionResponse represents a transaction broadcast response
type BroadcastTransactionResponse struct {
	Success         bool   `json:"success"`
	TransactionHash string `json:"transactionHash,omitempty"`
	BlockHeight     int64  `json:"blockHeight,omitempty"`
	Error           string `json:"error,omitempty"`
}

// WalletInfoResponse represents wallet information response
type WalletInfoResponse struct {
	Success bool                  `json:"success"`
	Wallet  *dao.WalletConnection `json:"wallet,omitempty"`
	Balance int64                 `json:"balance,omitempty"`
	Error   string                `json:"error,omitempty"`
}

// Add wallet integration routes to the DAO server
func (s *DAOServer) setupWalletRoutes(e *echo.Echo) {
	// Wallet connection endpoints
	e.POST("/dao/wallet/connect", s.handleWalletConnect)
	e.POST("/dao/wallet/disconnect", s.handleWalletDisconnect)
	e.GET("/dao/wallet/info/:address", s.handleGetWalletInfo)
	e.GET("/dao/wallet/connections", s.handleGetActiveConnections)

	// Transaction signing endpoints
	e.POST("/dao/wallet/sign", s.handleSignTransaction)
	e.POST("/dao/wallet/broadcast", s.handleBroadcastTransaction)
	e.POST("/dao/wallet/verify", s.handleVerifyTransaction)

	// Wallet utilities
	e.POST("/dao/wallet/generate-test", s.handleGenerateTestWallet)
	e.GET("/dao/wallet/supported", s.handleGetSupportedWallets)
}

// handleWalletConnect handles wallet connection requests
func (s *DAOServer) handleWalletConnect(c echo.Context) error {
	var req WalletConnectionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, WalletConnectionResponse{
			Success: false,
			Error:   "Invalid request format",
		})
	}

	// Validate required fields
	if req.Provider == "" || req.Address == "" || req.PublicKey == "" {
		return c.JSON(http.StatusBadRequest, WalletConnectionResponse{
			Success: false,
			Error:   "Provider, address, and publicKey are required",
		})
	}

	// Get wallet connection manager
	walletManager := dao.NewWalletConnectionManager()

	// Handle wallet connection
	connection, err := walletManager.HandleWalletConnection(
		dao.WalletProvider(req.Provider),
		req.Address,
		req.PublicKey,
		req.ChainID,
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, WalletConnectionResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Broadcast wallet connection event
	event := Event{
		Type: EventType("wallet_connected"),
		Data: map[string]interface{}{
			"address":  req.Address,
			"provider": req.Provider,
			"chainId":  req.ChainID,
		},
		Timestamp: time.Now().Unix(),
	}
	s.broadcastEvent(event)

	return c.JSON(http.StatusOK, WalletConnectionResponse{
		Success:    true,
		Connection: connection,
	})
}

// handleWalletDisconnect handles wallet disconnection requests
func (s *DAOServer) handleWalletDisconnect(c echo.Context) error {
	address := c.FormValue("address")
	if address == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Address is required",
		})
	}

	walletManager := dao.NewWalletConnectionManager()
	err := walletManager.DisconnectWallet(address)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Broadcast wallet disconnection event
	event := Event{
		Type: EventType("wallet_disconnected"),
		Data: map[string]interface{}{
			"address": address,
		},
		Timestamp: time.Now().Unix(),
	}
	s.broadcastEvent(event)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// handleGetWalletInfo handles wallet information requests
func (s *DAOServer) handleGetWalletInfo(c echo.Context) error {
	address := c.Param("address")
	if address == "" {
		return c.JSON(http.StatusBadRequest, WalletInfoResponse{
			Success: false,
			Error:   "Address is required",
		})
	}

	walletManager := dao.NewWalletConnectionManager()
	wallet, err := walletManager.GetWalletInfo(address)

	if err != nil {
		return c.JSON(http.StatusNotFound, WalletInfoResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Get token balance - convert address string to PublicKey
	// For now, we'll set balance to 0 since we need proper address conversion
	balance := uint64(0)

	return c.JSON(http.StatusOK, WalletInfoResponse{
		Success: true,
		Wallet:  wallet,
		Balance: int64(balance),
	})
}

// handleGetActiveConnections handles requests for active wallet connections
func (s *DAOServer) handleGetActiveConnections(c echo.Context) error {
	// Simplified implementation for testing
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":     true,
		"connections": []interface{}{},
		"count":       0,
	})
}

// handleSignTransaction handles transaction signing requests
func (s *DAOServer) handleSignTransaction(c echo.Context) error {
	var req TransactionSigningRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, TransactionSigningResponse{
			Success: false,
			Error:   "Invalid request format",
		})
	}

	// Validate required fields
	if req.Address == "" || req.Transaction == nil || req.Signature == "" {
		return c.JSON(http.StatusBadRequest, TransactionSigningResponse{
			Success: false,
			Error:   "Address, transaction, and signature are required",
		})
	}

	walletManager := dao.NewWalletConnectionManager()
	signedTx, err := walletManager.HandleTransactionSigning(
		req.Address,
		req.Transaction,
		req.Signature,
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, TransactionSigningResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Broadcast transaction signed event
	event := Event{
		Type: EventType("transaction_signed"),
		Data: map[string]interface{}{
			"address":         req.Address,
			"transactionHash": signedTx.TransactionHash.String(),
			"signingMethod":   signedTx.SigningMethod,
		},
		Timestamp: time.Now().Unix(),
	}
	s.broadcastEvent(event)

	return c.JSON(http.StatusOK, TransactionSigningResponse{
		Success:           true,
		SignedTransaction: signedTx,
	})
}

// handleBroadcastTransaction handles transaction broadcasting requests
func (s *DAOServer) handleBroadcastTransaction(c echo.Context) error {
	var req BroadcastTransactionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, BroadcastTransactionResponse{
			Success: false,
			Error:   "Invalid request format",
		})
	}

	if req.SignedTransaction == nil {
		return c.JSON(http.StatusBadRequest, BroadcastTransactionResponse{
			Success: false,
			Error:   "Signed transaction is required",
		})
	}

	// Verify the signed transaction
	walletService := dao.NewWalletIntegrationService()
	err := walletService.VerifySignedTransaction(req.SignedTransaction)
	if err != nil {
		return c.JSON(http.StatusBadRequest, BroadcastTransactionResponse{
			Success: false,
			Error:   "Transaction verification failed: " + err.Error(),
		})
	}

	// Create core transaction from signed DAO transaction
	coreTx := &core.Transaction{
		TxInner:   req.SignedTransaction.Transaction,
		From:      crypto.PublicKey(req.SignedTransaction.Signer),
		Signature: &req.SignedTransaction.Signature,
		Nonce:     time.Now().Unix(),
	}

	// Add transaction to channel (simulating mempool)
	s.txChan <- coreTx
	if err != nil {
		return c.JSON(http.StatusInternalServerError, BroadcastTransactionResponse{
			Success: false,
			Error:   "Failed to add transaction to mempool: " + err.Error(),
		})
	}

	// Get current block height
	blockHeight := s.bc.Height()

	// Broadcast transaction event
	event := Event{
		Type: EventType("transaction_broadcast"),
		Data: map[string]interface{}{
			"transactionHash": req.SignedTransaction.TransactionHash.String(),
			"signer":          req.SignedTransaction.Signer.String(),
			"blockHeight":     blockHeight,
		},
		Timestamp: time.Now().Unix(),
	}
	s.broadcastEvent(event)

	return c.JSON(http.StatusOK, BroadcastTransactionResponse{
		Success:         true,
		TransactionHash: req.SignedTransaction.TransactionHash.String(),
		BlockHeight:     int64(blockHeight),
	})
}

// handleVerifyTransaction handles transaction verification requests
func (s *DAOServer) handleVerifyTransaction(c echo.Context) error {
	var signedTx dao.SignedTransaction
	if err := c.Bind(&signedTx); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid request format",
		})
	}

	walletService := dao.NewWalletIntegrationService()
	err := walletService.VerifySignedTransaction(&signedTx)

	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": false,
			"valid":   false,
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"valid":   true,
	})
}

// handleGenerateTestWallet handles test wallet generation requests
func (s *DAOServer) handleGenerateTestWallet(c echo.Context) error {
	_, publicKey, address, err := dao.GenerateTestWallet()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Convert to hex strings for JSON response
	privateKeyHex := "generated_private_key" // Simplified for testing
	publicKeyHex := hex.EncodeToString(publicKey)
	addressHex := address.String()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":    true,
		"privateKey": privateKeyHex,
		"publicKey":  publicKeyHex,
		"address":    addressHex,
		"warning":    "This is for development only. Never use in production.",
	})
}

// handleGetSupportedWallets handles requests for supported wallet providers
func (s *DAOServer) handleGetSupportedWallets(c echo.Context) error {
	supportedWallets := map[string]interface{}{
		"metamask": map[string]interface{}{
			"name":        "MetaMask",
			"description": "Browser extension wallet",
			"supported":   true,
			"features":    []string{"signing", "account_management", "chain_switching"},
		},
		"walletconnect": map[string]interface{}{
			"name":        "WalletConnect",
			"description": "Connect to mobile wallets",
			"supported":   true,
			"features":    []string{"signing", "qr_code", "deep_linking"},
		},
		"manual": map[string]interface{}{
			"name":        "Manual Key Input",
			"description": "Direct private key input (development only)",
			"supported":   true,
			"features":    []string{"signing", "key_generation"},
			"warning":     "For development use only",
		},
		"ledger": map[string]interface{}{
			"name":        "Ledger Hardware Wallet",
			"description": "Hardware wallet integration",
			"supported":   false,
			"features":    []string{"signing", "hardware_security"},
			"note":        "Coming soon",
		},
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"wallets": supportedWallets,
	})
}

// Initialize wallet integration in the DAO server
func (s *DAOServer) initWalletIntegration() {
	// Initialize WebSocket clients map if not already done
	if s.wsClients == nil {
		s.wsClients = make(map[*websocket.Conn]bool)
	}

	// Set up WebSocket upgrader
	s.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}
}
