# ProjectX DAO API Server

The ProjectX DAO API Server provides a comprehensive REST API and WebSocket interface for interacting with the decentralized autonomous organization built on the ProjectX blockchain.

## Features

- **Proposal Management**: Create, view, and manage governance proposals
- **Voting System**: Cast votes with multiple voting mechanisms (simple, quadratic, weighted, reputation-based)
- **Treasury Operations**: Multi-signature treasury management with secure fund disbursement
- **Token Management**: Governance token operations including transfers, approvals, and balance queries
- **Delegation System**: Delegate voting power to trusted representatives
- **Real-time Events**: WebSocket support for live governance event updates
- **Member Management**: Track DAO member information and participation

## API Endpoints

### Proposal Endpoints

#### GET /dao/proposals
List all governance proposals.

**Response:**
```json
[
  {
    "id": "proposal_hash",
    "creator": "creator_public_key",
    "title": "Proposal Title",
    "description": "Proposal Description",
    "proposal_type": 1,
    "voting_type": 1,
    "start_time": 1640995200,
    "end_time": 1641081600,
    "status": 2,
    "threshold": 1000,
    "results": {
      "yes_votes": 5000,
      "no_votes": 2000,
      "abstain_votes": 500,
      "total_voters": 75,
      "quorum": 7500,
      "passed": true
    },
    "metadata_hash": "ipfs_hash"
  }
]
```

#### GET /dao/proposal/:id
Get a specific proposal by ID.

**Parameters:**
- `id`: Proposal hash (hex string)

#### POST /dao/proposal
Create a new governance proposal.

**Request Body:**
```json
{
  "title": "Proposal Title",
  "description": "Detailed proposal description",
  "proposal_type": 1,
  "voting_type": 1,
  "duration": 604800,
  "threshold": 1000,
  "metadata_hash": "optional_ipfs_hash",
  "private_key": "creator_private_key_hex"
}
```

**Proposal Types:**
- `1`: General governance
- `2`: Treasury spending
- `3`: Technical/protocol changes
- `4`: Parameter updates

**Voting Types:**
- `1`: Simple majority
- `2`: Quadratic voting
- `3`: Token-weighted
- `4`: Reputation-based

#### POST /dao/vote
Cast a vote on a proposal.

**Request Body:**
```json
{
  "proposal_id": "proposal_hash_hex",
  "choice": 1,
  "weight": 1000,
  "reason": "Optional voting reason",
  "private_key": "voter_private_key_hex"
}
```

**Vote Choices:**
- `1`: Yes
- `2`: No
- `3`: Abstain

#### GET /dao/proposal/:id/votes
Get all votes for a specific proposal.

### Treasury Endpoints

#### GET /dao/treasury
Get treasury status and information.

**Response:**
```json
{
  "balance": 1000000,
  "signers": ["signer1_pubkey", "signer2_pubkey", "signer3_pubkey"],
  "required_sigs": 2
}
```

#### GET /dao/treasury/transactions
Get treasury transaction history.

#### POST /dao/treasury/transaction
Create a new treasury transaction.

**Request Body:**
```json
{
  "recipient": "recipient_public_key_hex",
  "amount": 50000,
  "purpose": "Marketing budget allocation",
  "private_key": "signer_private_key_hex"
}
```

#### POST /dao/treasury/sign
Sign a pending treasury transaction.

**Request Body:**
```json
{
  "transaction_id": "transaction_hash_hex",
  "private_key": "signer_private_key_hex"
}
```

### Token Endpoints

#### GET /dao/token/balance/:address
Get token balance for an address.

**Parameters:**
- `address`: Public key (hex string)

**Response:**
```json
{
  "balance": 10000
}
```

#### GET /dao/token/supply
Get total token supply.

**Response:**
```json
{
  "total_supply": 1000000
}
```

#### POST /dao/token/transfer
Transfer tokens between addresses.

**Request Body:**
```json
{
  "to": "recipient_public_key_hex",
  "amount": 1000,
  "private_key": "sender_private_key_hex"
}
```

#### POST /dao/token/approve
Approve a spender to use tokens on your behalf.

**Request Body:**
```json
{
  "spender": "spender_public_key_hex",
  "amount": 5000,
  "private_key": "owner_private_key_hex"
}
```

#### GET /dao/token/allowance/:owner/:spender
Get allowance between owner and spender.

### Delegation Endpoints

#### POST /dao/delegate
Delegate voting power to another address.

**Request Body:**
```json
{
  "delegate": "delegate_public_key_hex",
  "duration": 2592000,
  "private_key": "delegator_private_key_hex"
}
```

#### POST /dao/revoke-delegation
Revoke existing delegation.

**Request Body:**
```json
{
  "private_key": "delegator_private_key_hex"
}
```

#### GET /dao/delegation/:address
Get delegation information for an address.

#### GET /dao/delegations
Get all active delegations.

### Member Endpoints

#### GET /dao/member/:address
Get member information.

**Response:**
```json
{
  "address": "member_public_key",
  "balance": 10000,
  "staked": 5000,
  "reputation": 1000,
  "joined_at": 1640995200,
  "last_active": 1641081600
}
```

#### GET /dao/members
Get all DAO members with pagination.

**Query Parameters:**
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 50, max: 100)

## WebSocket Events

### Connection
Connect to `ws://localhost:8080/dao/events` for real-time governance events.

### Event Types

#### proposal_created
Fired when a new proposal is created.
```json
{
  "type": "proposal_created",
  "data": {
    "title": "Proposal Title",
    "creator": "creator_public_key"
  },
  "timestamp": 1641081600
}
```

#### vote_cast
Fired when a vote is cast.
```json
{
  "type": "vote_cast",
  "data": {
    "proposal_id": "proposal_hash",
    "voter": "voter_public_key",
    "choice": 1
  },
  "timestamp": 1641081600
}
```

#### proposal_passed / proposal_rejected
Fired when a proposal concludes.
```json
{
  "type": "proposal_passed",
  "data": {
    "proposal_id": "proposal_hash",
    "final_results": { ... }
  },
  "timestamp": 1641081600
}
```

#### treasury_transaction
Fired when treasury operations occur.
```json
{
  "type": "treasury_transaction",
  "data": {
    "amount": 50000,
    "recipient": "recipient_public_key",
    "purpose": "Marketing budget"
  },
  "timestamp": 1641081600
}
```

#### delegation_updated
Fired when delegations change.
```json
{
  "type": "delegation_updated",
  "data": {
    "delegator": "delegator_public_key",
    "delegate": "delegate_public_key",
    "action": "delegate"
  },
  "timestamp": 1641081600
}
```

## Usage Examples

### JavaScript/React Integration

```javascript
// Create DAO API client
class DAOClient {
  constructor(baseUrl) {
    this.baseUrl = baseUrl;
  }

  async getProposals() {
    const response = await fetch(`${this.baseUrl}/dao/proposals`);
    return response.json();
  }

  async createProposal(proposalData) {
    const response = await fetch(`${this.baseUrl}/dao/proposal`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(proposalData)
    });
    return response.json();
  }

  async castVote(voteData) {
    const response = await fetch(`${this.baseUrl}/dao/vote`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(voteData)
    });
    return response.json();
  }

  connectToEvents(onEvent) {
    const ws = new WebSocket(`ws://localhost:8080/dao/events`);
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      onEvent(data);
    };
    return ws;
  }
}

// Usage
const daoClient = new DAOClient('http://localhost:8080');

// Get proposals
const proposals = await daoClient.getProposals();

// Create proposal
await daoClient.createProposal({
  title: "Marketing Budget Increase",
  description: "Allocate additional funds for Q2 marketing",
  proposal_type: 2,
  voting_type: 1,
  duration: 604800,
  threshold: 10000,
  private_key: "your_private_key_hex"
});

// Listen to events
const ws = daoClient.connectToEvents((event) => {
  console.log('DAO Event:', event);
});
```

### Flutter/Dart Integration

```dart
import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:web_socket_channel/web_socket_channel.dart';

class DAOClient {
  final String baseUrl;
  
  DAOClient(this.baseUrl);
  
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
  
  WebSocketChannel connectToEvents() {
    return WebSocketChannel.connect(
      Uri.parse('ws://localhost:8080/dao/events'),
    );
  }
}
```

## Error Handling

The API returns standard HTTP status codes and JSON error responses:

```json
{
  "error": "Error description"
}
```

Common error codes:
- `400`: Bad Request - Invalid request format or parameters
- `404`: Not Found - Resource not found
- `500`: Internal Server Error - Server-side error

## Security Considerations

1. **Private Key Handling**: Private keys are sent in request bodies for transaction signing. In production, consider using secure key management solutions.

2. **CORS**: The WebSocket upgrader currently allows all origins for development. Configure appropriate CORS policies for production.

3. **Rate Limiting**: Implement rate limiting to prevent abuse of API endpoints.

4. **Input Validation**: All inputs are validated, but additional sanitization may be needed for production use.

5. **HTTPS**: Use HTTPS in production to encrypt all API communications.

## Development and Testing

### Running Tests
```bash
cd projectx
go test ./api -v
```

### Building
```bash
go build ./api
```

### Starting the Server
```go
// See dao_server_example.go for complete setup example
server := NewDAOServer(cfg, blockchain, txChan, daoInstance)
server.Start()
```

## Integration with ProjectX Blockchain

The DAO API server integrates seamlessly with the ProjectX blockchain:

1. **Transaction Processing**: All DAO operations create blockchain transactions
2. **State Management**: DAO state is maintained consistently with blockchain state
3. **Event Broadcasting**: Real-time events are triggered by blockchain state changes
4. **Security**: All operations are cryptographically signed and verified

## Future Enhancements

- **IPFS Integration**: Store large proposal documents on IPFS
- **Layer-2 Scaling**: Implement off-chain computation for complex operations
- **Advanced Analytics**: Add governance participation and treasury performance metrics
- **Mobile SDK**: Native mobile SDKs for iOS and Android
- **GraphQL API**: Alternative GraphQL interface for complex queries