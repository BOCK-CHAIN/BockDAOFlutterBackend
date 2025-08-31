package api

import (
	"log"

	"github.com/BOCK-CHAIN/BockChain/core"
	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/dao"
)

// Example demonstrates how to set up and use the enhanced DAO API server
func ExampleDAOServer() {
	// 1. Create a blockchain instance (simplified for example)
	bc := &core.Blockchain{}

	// 2. Create a DAO instance
	daoInstance := dao.NewDAO("GOVTOKEN", "Governance Token", 18)

	// 3. Initialize treasury with multi-sig setup
	signer1 := crypto.GeneratePrivateKey()
	signer2 := crypto.GeneratePrivateKey()
	signer3 := crypto.GeneratePrivateKey()

	signers := []crypto.PublicKey{
		signer1.PublicKey(),
		signer2.PublicKey(),
		signer3.PublicKey(),
	}

	err := daoInstance.InitializeTreasury(signers, 2) // Require 2 out of 3 signatures
	if err != nil {
		log.Fatal("Failed to initialize treasury:", err)
	}

	// 4. Distribute initial tokens to founding members
	founder1 := crypto.GeneratePrivateKey()
	founder2 := crypto.GeneratePrivateKey()
	founder3 := crypto.GeneratePrivateKey()

	initialDistribution := map[string]uint64{
		founder1.PublicKey().String(): 100000, // 100k tokens
		founder2.PublicKey().String(): 75000,  // 75k tokens
		founder3.PublicKey().String(): 50000,  // 50k tokens
	}

	err = daoInstance.InitialTokenDistribution(initialDistribution)
	if err != nil {
		log.Fatal("Failed to distribute initial tokens:", err)
	}

	// 5. Add funds to treasury
	daoInstance.AddTreasuryFunds(1000000) // 1M units

	// 6. Create transaction channel for processing
	txChan := make(chan *core.Transaction, 1000)

	// 7. Set up server configuration
	cfg := ServerConfig{
		Logger:     nil, // Use nil for example
		ListenAddr: ":8080",
	}

	// 8. Create and start the enhanced DAO server
	server := NewDAOServer(cfg, bc, txChan, daoInstance)

	log.Println("Starting DAO API server on :8080")
	log.Println("Available endpoints:")
	log.Println("  GET  /dao/proposals - List all proposals")
	log.Println("  GET  /dao/proposal/:id - Get specific proposal")
	log.Println("  POST /dao/proposal - Create new proposal")
	log.Println("  POST /dao/vote - Cast vote")
	log.Println("  GET  /dao/treasury - Get treasury status")
	log.Println("  POST /dao/treasury/transaction - Create treasury transaction")
	log.Println("  POST /dao/treasury/sign - Sign treasury transaction")
	log.Println("  GET  /dao/token/balance/:address - Get token balance")
	log.Println("  GET  /dao/token/supply - Get total token supply")
	log.Println("  POST /dao/delegate - Delegate voting power")
	log.Println("  GET  /dao/member/:address - Get member information")
	log.Println("  GET  /dao/events - WebSocket for real-time events")

	// Start processing transactions in background
	go func() {
		for tx := range txChan {
			log.Printf("Processing transaction: %s", tx.Hash(core.TxHasher{}).String())
			// In a real implementation, you would process the transaction
			// through the blockchain and update the DAO state accordingly
		}
	}()

	// Start the server
	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// ExampleAPIUsage demonstrates how to interact with the DAO API
func ExampleAPIUsage() {
	// Example API requests that can be made to the DAO server:

	// 1. Create a proposal
	/*
		POST /dao/proposal
		{
			"title": "Increase Marketing Budget",
			"description": "Proposal to allocate 50,000 tokens for marketing initiatives",
			"proposal_type": 2,
			"voting_type": 1,
			"duration": 604800,
			"threshold": 10000,
			"metadata_hash": "",
			"private_key": "your_private_key_hex"
		}
	*/

	// 2. Cast a vote
	/*
		POST /dao/vote
		{
			"proposal_id": "proposal_hash_hex",
			"choice": 1,
			"weight": 1000,
			"reason": "I support this initiative",
			"private_key": "your_private_key_hex"
		}
	*/

	// 3. Create treasury transaction
	/*
		POST /dao/treasury/transaction
		{
			"recipient": "recipient_public_key_hex",
			"amount": 50000,
			"purpose": "Marketing budget allocation",
			"private_key": "signer_private_key_hex"
		}
	*/

	// 4. Delegate voting power
	/*
		POST /dao/delegate
		{
			"delegate": "delegate_public_key_hex",
			"duration": 2592000,
			"private_key": "your_private_key_hex"
		}
	*/

	// 5. WebSocket connection for real-time events
	/*
		WebSocket connection to: ws://localhost:8080/dao/events

		Events you'll receive:
		- proposal_created: When new proposals are submitted
		- vote_cast: When votes are cast
		- proposal_passed/proposal_rejected: When proposals conclude
		- treasury_transaction: When treasury operations occur
		- delegation_updated: When delegations change
	*/
}

// ExampleWebSocketClient shows how to connect to real-time events
func ExampleWebSocketClient() {
	/*
		JavaScript example for connecting to WebSocket events:

		const ws = new WebSocket('ws://localhost:8080/dao/events');

		ws.onopen = function(event) {
			console.log('Connected to DAO events');
		};

		ws.onmessage = function(event) {
			const eventData = JSON.parse(event.data);
			console.log('DAO Event:', eventData);

			switch(eventData.type) {
				case 'proposal_created':
					console.log('New proposal:', eventData.data.title);
					break;
				case 'vote_cast':
					console.log('Vote cast on proposal:', eventData.data.proposal_id);
					break;
				case 'treasury_transaction':
					console.log('Treasury transaction:', eventData.data.amount);
					break;
			}
		};

		ws.onclose = function(event) {
			console.log('Disconnected from DAO events');
		};
	*/
}

// ExampleIntegrationWithFrontend shows how frontend applications can integrate
func ExampleIntegrationWithFrontend() {
	/*
		Frontend Integration Examples:

		1. React/JavaScript Web App:
		```javascript
		// Create proposal
		const createProposal = async (proposalData) => {
			const response = await fetch('/dao/proposal', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(proposalData)
			});
			return response.json();
		};

		// Get proposals
		const getProposals = async () => {
			const response = await fetch('/dao/proposals');
			return response.json();
		};

		// Cast vote
		const castVote = async (voteData) => {
			const response = await fetch('/dao/vote', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(voteData)
			});
			return response.json();
		};
		```

		2. Flutter Mobile App:
		```dart
		// HTTP client for API calls
		class DAOApiClient {
			static const String baseUrl = 'http://localhost:8080';

			Future<List<Proposal>> getProposals() async {
				final response = await http.get(Uri.parse('$baseUrl/dao/proposals'));
				if (response.statusCode == 200) {
					final List<dynamic> data = json.decode(response.body);
					return data.map((json) => Proposal.fromJson(json)).toList();
				}
				throw Exception('Failed to load proposals');
			}

			Future<void> createProposal(ProposalRequest request) async {
				final response = await http.post(
					Uri.parse('$baseUrl/dao/proposal'),
					headers: {'Content-Type': 'application/json'},
					body: json.encode(request.toJson()),
				);
				if (response.statusCode != 200) {
					throw Exception('Failed to create proposal');
				}
			}
		}
		```

		3. WebSocket Integration:
		```javascript
		class DAOEventManager {
			constructor(wsUrl) {
				this.ws = new WebSocket(wsUrl);
				this.eventHandlers = {};
			}

			on(eventType, handler) {
				if (!this.eventHandlers[eventType]) {
					this.eventHandlers[eventType] = [];
				}
				this.eventHandlers[eventType].push(handler);
			}

			connect() {
				this.ws.onmessage = (event) => {
					const data = JSON.parse(event.data);
					const handlers = this.eventHandlers[data.type] || [];
					handlers.forEach(handler => handler(data));
				};
			}
		}
		```
	*/
}
