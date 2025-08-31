package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/BOCK-CHAIN/BockChain/api"
	"github.com/BOCK-CHAIN/BockChain/core"
	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/dao"
	"github.com/BOCK-CHAIN/BockChain/network"
	"github.com/BOCK-CHAIN/BockChain/types"
	kitlog "github.com/go-kit/log"
)

// DAOSystemIntegration represents the complete integrated DAO system
type DAOSystemIntegration struct {
	blockchain    *core.Blockchain
	networkServer *network.Server
	daoServer     *api.DAOServer
	daoInstance   *dao.DAO
	logger        kitlog.Logger
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	txChan        chan *core.Transaction
}

// NewDAOSystemIntegration creates a new integrated DAO system
func NewDAOSystemIntegration(config *DAOSystemConfig) (*DAOSystemIntegration, error) {
	logger := kitlog.NewLogfmtLogger(os.Stdout)
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize blockchain with DAO support
	genesis := createDAOGenesisBlock()
	blockchain, err := core.NewBlockchain(logger, genesis)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create blockchain: %w", err)
	}

	// Create transaction channel for communication
	txChan := make(chan *core.Transaction, 1000)

	// Initialize DAO instance
	daoInstance := dao.NewDAO("GOVX", "ProjectX Governance Token", 18)

	// Initialize DAO with initial token distribution
	initialDistribution := map[string]uint64{
		config.ValidatorKey.PublicKey().String(): 1000000, // 1M tokens to validator
	}
	if err := daoInstance.InitialTokenDistribution(initialDistribution); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize token distribution: %w", err)
	}

	// Initialize treasury with validator as initial signer
	treasurySigners := []crypto.PublicKey{config.ValidatorKey.PublicKey()}
	if err := daoInstance.InitializeTreasury(treasurySigners, 1); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize treasury: %w", err)
	}

	// Initialize network server
	networkOpts := network.ServerOpts{
		APIListenAddr: config.APIListenAddr,
		SeedNodes:     config.SeedNodes,
		ListenAddr:    config.ListenAddr,
		PrivateKey:    &config.ValidatorKey,
		ID:            config.NodeID,
		Logger:        logger,
	}

	networkServer, err := network.NewServer(networkOpts)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create network server: %w", err)
	}

	// Initialize DAO server with enhanced API
	apiConfig := api.ServerConfig{
		Logger:     logger,
		ListenAddr: config.APIListenAddr,
	}
	daoServer := api.NewDAOServer(apiConfig, blockchain, txChan, daoInstance)

	return &DAOSystemIntegration{
		blockchain:    blockchain,
		networkServer: networkServer,
		daoServer:     daoServer,
		daoInstance:   daoInstance,
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
		txChan:        txChan,
	}, nil
}

// Start initializes and starts all system components
func (d *DAOSystemIntegration) Start() error {
	d.logger.Log("msg", "Starting DAO system integration")

	// Start transaction processor
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.processTxLoop()
	}()

	// Start network server
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.networkServer.Start()
	}()

	// Start DAO API server
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		if err := d.daoServer.Start(); err != nil {
			d.logger.Log("error", "DAO server failed", "err", err)
		}
	}()

	// Start background services
	d.startBackgroundServices()

	// Setup graceful shutdown
	d.setupGracefulShutdown()

	d.logger.Log("msg", "DAO system integration started successfully")
	return nil
}

// Stop gracefully shuts down all system components
func (d *DAOSystemIntegration) Stop() error {
	d.logger.Log("msg", "Stopping DAO system integration")

	d.cancel()
	d.wg.Wait()

	// Close transaction channel
	close(d.txChan)

	d.logger.Log("msg", "DAO system integration stopped")
	return nil
}

// processTxLoop processes transactions from the transaction channel
func (d *DAOSystemIntegration) processTxLoop() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case tx := <-d.txChan:
			if tx == nil {
				return // Channel closed
			}
			if err := d.processTransaction(tx); err != nil {
				d.logger.Log("error", "Failed to process transaction", "err", err, "tx_hash", tx.Hash(core.TxHasher{}).String())
			}
		}
	}
}

// processTransaction processes a single transaction
func (d *DAOSystemIntegration) processTransaction(tx *core.Transaction) error {
	txHash := tx.Hash(core.TxHasher{})

	// Process DAO transactions through the DAO instance
	if err := d.daoInstance.ProcessDAOTransaction(tx.TxInner, tx.From, txHash); err != nil {
		return fmt.Errorf("failed to process DAO transaction: %w", err)
	}

	// Add transaction to blockchain
	block := d.createBlockWithTx(tx)
	if err := d.blockchain.AddBlock(block); err != nil {
		return fmt.Errorf("failed to add block to blockchain: %w", err)
	}

	d.logger.Log("msg", "Transaction processed successfully", "tx_hash", txHash.String())
	return nil
}

// createBlockWithTx creates a new block containing the given transaction
func (d *DAOSystemIntegration) createBlockWithTx(tx *core.Transaction) *core.Block {
	prevBlock, _ := d.blockchain.GetBlock(d.blockchain.Height())

	header := &core.Header{
		Version:       1,
		PrevBlockHash: prevBlock.Hash(core.BlockHasher{}),
		Height:        d.blockchain.Height() + 1,
		Timestamp:     time.Now().UnixNano(),
	}

	block, err := core.NewBlock(header, []*core.Transaction{tx})
	if err != nil {
		d.logger.Log("error", "Failed to create block", "err", err)
		return nil
	}

	// Calculate data hash
	dataHash, err := core.CalculateDataHash(block.Transactions)
	if err != nil {
		d.logger.Log("error", "Failed to calculate data hash", "err", err)
		return nil
	}
	block.Header.DataHash = dataHash

	return block
}

// startBackgroundServices starts essential background services
func (d *DAOSystemIntegration) startBackgroundServices() {
	// Proposal status updater
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-d.ctx.Done():
				return
			case <-ticker.C:
				d.daoInstance.UpdateAllProposalStatuses()
			}
		}
	}()

	// Treasury management service
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-d.ctx.Done():
				return
			case <-ticker.C:
				if err := d.processPendingTreasuryTransactions(); err != nil {
					d.logger.Log("error", "Failed to process treasury transactions", "err", err)
				}
			}
		}
	}()

	// Reputation system updater
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-d.ctx.Done():
				return
			case <-ticker.C:
				d.daoInstance.ApplyInactivityDecay()
			}
		}
	}()
}

// setupGracefulShutdown sets up signal handling for graceful shutdown
func (d *DAOSystemIntegration) setupGracefulShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		select {
		case <-d.ctx.Done():
			return
		case sig := <-sigChan:
			d.logger.Log("msg", "Received shutdown signal", "signal", sig)
			d.Stop()
		}
	}()
}

// processPendingTreasuryTransactions processes pending multi-sig treasury transactions
func (d *DAOSystemIntegration) processPendingTreasuryTransactions() error {
	pendingTxs := d.daoInstance.GetPendingTreasuryTransactions()

	for txID, pendingTx := range pendingTxs {
		if pendingTx.Executed || time.Now().Unix() > pendingTx.ExpiresAt {
			continue
		}

		// Check if we have enough signatures
		requiredSigs := d.daoInstance.GetRequiredSignatures()
		if len(pendingTx.Signatures) >= int(requiredSigs) {
			// Execute the transaction
			if err := d.daoInstance.ExecuteTreasuryTransaction(txID); err != nil {
				d.logger.Log("error", "Failed to execute treasury transaction", "txID", txID, "err", err)
				continue
			}
			d.logger.Log("msg", "Executed treasury transaction", "txID", txID, "amount", pendingTx.Amount)
		}
	}

	return nil
}

// DAOSystemConfig holds configuration for the DAO system
type DAOSystemConfig struct {
	NodeID         string
	ListenAddr     string
	APIListenAddr  string
	SeedNodes      []string
	ValidatorKey   crypto.PrivateKey
	EnableSecurity bool
	EnableIPFS     bool
}

// createDAOGenesisBlock creates a genesis block with DAO initialization
func createDAOGenesisBlock() *core.Block {
	privKey := crypto.GeneratePrivateKey()

	// Create a simple genesis transaction
	genesisTx := &core.Transaction{
		TxInner: core.CollectionTx{
			Fee:      0,
			MetaData: []byte("Genesis DAO Block"),
		},
		From:  privKey.PublicKey(),
		To:    privKey.PublicKey(),
		Value: 1000000000,
	}
	genesisTx.Sign(privKey)

	header := &core.Header{
		Version:       1,
		PrevBlockHash: types.Hash{},
		Height:        0,
		Timestamp:     time.Now().UnixNano(),
	}

	block, err := core.NewBlock(header, []*core.Transaction{genesisTx})
	if err != nil {
		panic(fmt.Sprintf("Failed to create genesis block: %v", err))
	}

	dataHash, err := core.CalculateDataHash(block.Transactions)
	if err != nil {
		panic(fmt.Sprintf("Failed to calculate data hash: %v", err))
	}
	block.Header.DataHash = dataHash

	if err := block.Sign(privKey); err != nil {
		panic(fmt.Sprintf("Failed to sign genesis block: %v", err))
	}

	return block
}

// RunDAOSystem is the main entry point for running the integrated DAO system
func RunDAOSystem() error {
	config := &DAOSystemConfig{
		NodeID:         "DAO_NODE_1",
		ListenAddr:     ":3000",
		APIListenAddr:  ":9000",
		SeedNodes:      []string{},
		ValidatorKey:   crypto.GeneratePrivateKey(),
		EnableSecurity: true,
		EnableIPFS:     true,
	}

	system, err := NewDAOSystemIntegration(config)
	if err != nil {
		return fmt.Errorf("failed to create DAO system: %w", err)
	}

	if err := system.Start(); err != nil {
		return fmt.Errorf("failed to start DAO system: %w", err)
	}

	// Keep the system running
	select {}
}

// Health check endpoint for monitoring
func (d *DAOSystemIntegration) HealthCheck() *HealthStatus {
	return &HealthStatus{
		Blockchain:    d.blockchain.Height() > 0,
		NetworkServer: true, // Simplified check
		DAOServer:     true, // Simplified check
		Timestamp:     time.Now().Unix(),
	}
}

type HealthStatus struct {
	Blockchain    bool  `json:"blockchain"`
	NetworkServer bool  `json:"network_server"`
	DAOServer     bool  `json:"dao_server"`
	Timestamp     int64 `json:"timestamp"`
}

// GetDAOInstance returns the DAO instance for testing
func (d *DAOSystemIntegration) GetDAOInstance() *dao.DAO {
	return d.daoInstance
}

// GetBlockchain returns the blockchain instance for testing
func (d *DAOSystemIntegration) GetBlockchain() *core.Blockchain {
	return d.blockchain
}

// GetTxChan returns the transaction channel for testing
func (d *DAOSystemIntegration) GetTxChan() chan *core.Transaction {
	return d.txChan
}
