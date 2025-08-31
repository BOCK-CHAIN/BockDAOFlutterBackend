# ProjectX - Modular Blockchain with DAO Governance

A community-driven, modular, production-ready blockchain built from scratch with comprehensive DAO (Decentralized Autonomous Organization) governance capabilities. Every line of code has been developed through live coding sessions, ensuring complete transparency and community involvement.

## 🎥 Live Development Sessions

All development sessions are recorded and available on [YouTube](https://www.youtube.com/channel/UCIjIAXXsX4YMYeFj-LP42-Q). Watch the entire blockchain being built from the ground up!

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [System Requirements](#system-requirements)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Usage Examples](#usage-examples)
- [API Documentation](#api-documentation)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## 🌟 Overview

ProjectX is a next-generation blockchain platform that combines:

- **Custom Blockchain Engine**: Built from scratch with Go for maximum performance
- **DAO Governance System**: Complete decentralized governance with proposals, voting, and treasury management
- **Multi-Platform Support**: Web, mobile (Flutter), and backend integrations
- **Wallet Integration**: Support for MetaMask, WalletConnect, hardware wallets, and more
- **NFT & Token Support**: Native support for NFTs, collections, and custom tokens
- **IPFS Integration**: Decentralized storage for metadata and content
- **Real-time API**: WebSocket and REST API for real-time blockchain interaction

## ✨ Features

### Core Blockchain Features
- ✅ **Custom Consensus Mechanism**: Efficient block validation and consensus
- ✅ **Transaction Pool**: Optimized transaction management and processing
- ✅ **State Management**: Comprehensive blockchain state tracking
- ✅ **Cryptographic Security**: Robust signature verification and hashing
- ✅ **Network Layer**: P2P networking with TCP transport
- ✅ **Block Explorer**: Built-in blockchain exploration capabilities

### DAO Governance Features
- ✅ **Proposal Management**: Create, vote on, and execute governance proposals
- ✅ **Token-based Voting**: Weighted voting based on token holdings
- ✅ **Delegation System**: Delegate voting power to trusted representatives
- ✅ **Treasury Management**: Multi-signature treasury with fund management
- ✅ **Reputation System**: Track and reward community participation
- ✅ **Security Controls**: Role-based access control and emergency mechanisms
- ✅ **Analytics Dashboard**: Comprehensive governance analytics

### Integration Features
- ✅ **Multi-Wallet Support**: MetaMask, WalletConnect, hardware wallets
- ✅ **Cross-Platform**: Web, mobile (Flutter), and backend APIs
- ✅ **IPFS Storage**: Decentralized content and metadata storage
- ✅ **Real-time Events**: WebSocket-based real-time updates
- ✅ **RESTful API**: Complete REST API for all blockchain operations

## 🔧 System Requirements

### Minimum Requirements
- **Operating System**: Windows 10+, macOS 10.15+, or Linux (Ubuntu 18.04+)
- **Go Version**: 1.18 or higher
- **RAM**: 4GB minimum, 8GB recommended
- **Storage**: 10GB free space
- **Network**: Stable internet connection for P2P networking

### Development Requirements
- **Go**: 1.18+ with modules enabled
- **Make**: For build automation
- **Git**: For version control
- **Node.js**: 16+ (for web components)
- **Flutter**: 3.0+ (for mobile development)

### Optional Dependencies
- **Docker**: For containerized deployment
- **IPFS**: For decentralized storage features
- **PostgreSQL**: For advanced analytics (optional)

## 🚀 Installation

### 1. Clone the Repository

```bash
git clone https://github.com/BOCK-CHAIN/BockChain.git
cd BockChain
```

### 2. Install Go Dependencies

```bash
go mod download
go mod verify
```

### 3. Build the Project

```bash
make build
```

This creates the executable in `./bin/projectx`.

### 4. Verify Installation

```bash
./bin/projectx --version
```

## ⚡ Quick Start

### 1. Start a Local Network

```bash
# Build and run the blockchain network
make run
```

This starts a multi-node network with:
- **Local Node**: `:3000` (validator node with API on `:9000`)
- **Remote Node**: `:4000` (peer node)
- **Remote Node B**: `:5000` (peer node)
- **Late Node**: `:6000` (joins after 11 seconds)

### 2. Access the Web Interface

Open your browser and navigate to:
```
http://localhost:9000
```

### 3. Connect a Wallet

1. Open the web interface
2. Click "Connect Wallet"
3. Choose your preferred wallet (MetaMask, WalletConnect, or Manual)
4. Follow the connection prompts

### 4. Interact with the DAO

```bash
# Create a proposal
curl -X POST http://localhost:9000/dao/proposals \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Proposal",
    "description": "This is a test proposal",
    "fee": 1000
  }'

# Vote on a proposal
curl -X POST http://localhost:9000/dao/vote \
  -H "Content-Type: application/json" \
  -d '{
    "proposalId": "proposal_id_here",
    "choice": "yes",
    "fee": 500
  }'
```

## 🏗️ Architecture

### System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │  Mobile Client  │    │  API Clients    │
│                 │    │                 │    │                 │
│ • React/JS      │    │ • Flutter       │    │ • REST API      │
│ • Wallet UI     │    │ • Native UI     │    │ • WebSocket     │
│ • Real-time     │    │ • Secure Store  │    │ • Integration   │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   API Server    │
                    │                 │
                    │ • REST API      │
                    │ • WebSocket     │
                    │ • Authentication│
                    │ • Rate Limiting │
                    └─────────┬───────┘
                              │
          ┌───────────────────┼───────────────────┐
          │                   │                   │
┌─────────▼───────┐  ┌────────▼────────┐  ┌──────▼──────┐
│   DAO System    │  │   Blockchain    │  │   Network   │
│                 │  │                 │  │             │
│ • Governance    │  │ • Transactions  │  │ • P2P       │
│ • Proposals     │  │ • Blocks        │  │ • Transport │
│ • Voting        │  │ • State         │  │ • Discovery │
│ • Treasury      │  │ • Validation    │  │ • Consensus │
│ • Tokens        │  │ • Storage       │  │ • Security  │
└─────────────────┘  └─────────────────┘  └─────────────┘
```

### Core Components

1. **Blockchain Core** (`/core`)
   - Transaction processing and validation
   - Block creation and validation
   - State management and storage
   - Virtual machine for smart contracts

2. **DAO System** (`/dao`)
   - Governance proposals and voting
   - Token management and distribution
   - Treasury and fund management
   - Reputation and delegation systems

3. **Network Layer** (`/network`)
   - P2P networking and transport
   - Message routing and discovery
   - Transaction pool management
   - Consensus mechanisms

4. **API Layer** (`/api`)
   - REST API endpoints
   - WebSocket real-time events
   - Authentication and authorization
   - Rate limiting and security

5. **Wallet Integration** (`/web`, `/mobile`)
   - Multi-wallet support
   - Transaction signing
   - Secure key management
   - Cross-platform compatibility

## 💡 Usage Examples

### Creating and Managing Proposals

```go
// Create a new DAO proposal
proposal := &dao.Proposal{
    Title:       "Increase Block Rewards",
    Description: "Proposal to increase mining rewards by 10%",
    Fee:         1000,
    Metadata:    []byte(`{"category": "economic"}`),
}

// Submit the proposal
proposalID, err := daoSystem.CreateProposal(proposal, creatorAddress)
if err != nil {
    log.Fatal(err)
}

// Vote on the proposal
vote := &dao.Vote{
    ProposalID: proposalID,
    Choice:     dao.VoteYes,
    Fee:        500,
}

err = daoSystem.CastVote(vote, voterAddress)
if err != nil {
    log.Fatal(err)
}
```

### Token Operations

```go
// Transfer tokens between accounts
transfer := &core.Transaction{
    To:    recipientAddress,
    Value: 1000, // Amount in smallest unit
    Fee:   10,
}

err = transfer.Sign(senderPrivateKey)
if err != nil {
    log.Fatal(err)
}

// Broadcast the transaction
txHash, err := blockchain.ProcessTransaction(transfer)
if err != nil {
    log.Fatal(err)
}
```

### NFT Creation and Management

```go
// Create an NFT collection
collection := &core.CollectionTx{
    Fee:      200,
    MetaData: []byte("My NFT Collection"),
}

collectionTx := core.NewTransaction(collection)
collectionTx.Sign(ownerPrivateKey)

// Mint an NFT in the collection
nftMetadata := map[string]interface{}{
    "name":        "Rare Digital Art",
    "description": "A unique piece of digital art",
    "image":       "ipfs://QmHash...",
    "attributes": []map[string]interface{}{
        {"trait_type": "Color", "value": "Blue"},
        {"trait_type": "Rarity", "value": "Legendary"},
    },
}

mintTx := &core.MintTx{
    Fee:             200,
    NFT:             util.RandomHash(),
    MetaData:        jsonEncode(nftMetadata),
    Collection:      collectionHash,
    CollectionOwner: ownerPublicKey,
}

nftTx := core.NewTransaction(mintTx)
nftTx.Sign(ownerPrivateKey)
```

## 📚 API Documentation

### REST API Endpoints

#### Blockchain Operations
```
GET    /blocks                    - Get recent blocks
GET    /blocks/:hash              - Get specific block
GET    /transactions              - Get recent transactions
GET    /transactions/:hash        - Get specific transaction
POST   /transactions              - Submit new transaction
GET    /accounts/:address         - Get account information
```

#### DAO Operations
```
GET    /dao/proposals             - List all proposals
POST   /dao/proposals             - Create new proposal
GET    /dao/proposals/:id         - Get specific proposal
POST   /dao/vote                  - Cast a vote
GET    /dao/votes/:proposalId     - Get votes for proposal
GET    /dao/treasury              - Get treasury information
POST   /dao/delegate              - Delegate voting power
```

#### Wallet Operations
```
POST   /dao/wallet/connect        - Connect wallet
POST   /dao/wallet/disconnect     - Disconnect wallet
POST   /dao/wallet/sign           - Sign transaction
POST   /dao/wallet/broadcast      - Broadcast signed transaction
GET    /dao/wallet/info/:address  - Get wallet information
```

### WebSocket Events

Connect to `ws://localhost:9000/dao/events` for real-time updates:

```javascript
const ws = new WebSocket('ws://localhost:9000/dao/events');

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    
    switch(data.type) {
        case 'new_block':
            console.log('New block:', data.block);
            break;
        case 'new_transaction':
            console.log('New transaction:', data.transaction);
            break;
        case 'proposal_created':
            console.log('New proposal:', data.proposal);
            break;
        case 'vote_cast':
            console.log('Vote cast:', data.vote);
            break;
    }
};
```

## 🛠️ Development

### Project Structure

```
projectx/
├── api/                 # REST API and WebSocket servers
├── core/               # Blockchain core (transactions, blocks, state)
├── crypto/             # Cryptographic utilities
├── dao/                # DAO governance system
├── network/            # P2P networking and transport
├── types/              # Common types and data structures
├── util/               # Utility functions
├── web/                # Web interface and JavaScript client
├── mobile/             # Flutter mobile application
├── tests/              # Integration and system tests
├── main.go             # Main application entry point
├── Makefile            # Build automation
└── README.md           # This file
```

### Building from Source

```bash
# Install dependencies
go mod download

# Run tests
make test

# Build for development
make build

# Build for production
go build -ldflags="-s -w" -o ./bin/projectx

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o ./bin/projectx-linux
GOOS=windows GOARCH=amd64 go build -o ./bin/projectx-windows.exe
GOOS=darwin GOARCH=amd64 go build -o ./bin/projectx-macos
```

### Development Workflow

1. **Fork and Clone**: Fork the repository and clone your fork
2. **Create Branch**: Create a feature branch for your changes
3. **Develop**: Make your changes with comprehensive tests
4. **Test**: Run all tests and ensure they pass
5. **Document**: Update documentation as needed
6. **Submit**: Create a pull request with detailed description

### Code Style Guidelines

- Follow Go best practices and conventions
- Use `gofmt` for code formatting
- Write comprehensive tests for new features
- Document public APIs with clear comments
- Use meaningful variable and function names
- Keep functions focused and modular

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./core -v
go test ./dao -v
go test ./network -v

# Run integration tests
go test ./tests -v

# Run tests with coverage
go test -cover ./...

# Run performance tests
go test -bench=. ./...
```

### Test Categories

1. **Unit Tests**: Test individual components in isolation
2. **Integration Tests**: Test component interactions
3. **System Tests**: Test complete system workflows
4. **Performance Tests**: Test system performance and scalability
5. **Security Tests**: Test security mechanisms and attack resistance

### Test Examples

```bash
# Test DAO functionality
go test -v ./dao -run TestDAOProposalCreation
go test -v ./dao -run TestVotingMechanism
go test -v ./dao -run TestTreasuryManagement

# Test blockchain operations
go test -v ./core -run TestTransactionValidation
go test -v ./core -run TestBlockCreation
go test -v ./core -run TestStateManagement

# Test wallet integration
go test -v ./dao -run TestWalletIntegration
```

## 🚀 Deployment

### Local Development Deployment

```bash
# Start local network
make run

# Or start individual components
./bin/projectx --node-id=LOCAL_NODE --listen=:3000 --api=:9000
./bin/projectx --node-id=REMOTE_NODE --listen=:4000 --peers=:3000
```

### Production Deployment

#### Using Docker

```dockerfile
# Dockerfile
FROM golang:1.18-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o projectx

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/projectx .
EXPOSE 3000 9000
CMD ["./projectx"]
```

```bash
# Build and run with Docker
docker build -t projectx .
docker run -p 3000:3000 -p 9000:9000 projectx
```

#### Using Docker Compose

```yaml
# docker-compose.yml
version: '3.8'
services:
  validator:
    build: .
    ports:
      - "3000:3000"
      - "9000:9000"
    environment:
      - NODE_ID=VALIDATOR
      - LISTEN_ADDR=:3000
      - API_ADDR=:9000
    
  peer1:
    build: .
    ports:
      - "4000:4000"
    environment:
      - NODE_ID=PEER1
      - LISTEN_ADDR=:4000
      - PEERS=validator:3000
    depends_on:
      - validator
```

```bash
docker-compose up -d
```

### Cloud Deployment

#### AWS Deployment
1. Create EC2 instances for validator and peer nodes
2. Configure security groups for P2P communication
3. Set up load balancer for API endpoints
4. Configure auto-scaling for peer nodes
5. Set up monitoring and logging

#### Kubernetes Deployment
```yaml
# k8s-deployment.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: projectx-validator
spec:
  replicas: 1
  selector:
    matchLabels:
      app: projectx-validator
  template:
    metadata:
      labels:
        app: projectx-validator
    spec:
      containers:
      - name: projectx
        image: projectx:latest
        ports:
        - containerPort: 3000
        - containerPort: 9000
        env:
        - name: NODE_ID
          value: "VALIDATOR"
        - name: LISTEN_ADDR
          value: ":3000"
        - name: API_ADDR
          value: ":9000"
```

## 🔍 Troubleshooting

### Common Issues

#### Build Issues

**Problem**: `go build` fails with module errors
```bash
# Solution: Clean and rebuild modules
go clean -modcache
go mod download
go mod verify
make build
```

**Problem**: Missing dependencies
```bash
# Solution: Update dependencies
go mod tidy
go mod download
```

#### Runtime Issues

**Problem**: Network connection failures
```bash
# Check if ports are available
netstat -an | grep :3000
netstat -an | grep :9000

# Check firewall settings
# Windows: Windows Defender Firewall
# macOS: System Preferences > Security & Privacy > Firewall
# Linux: ufw status
```

**Problem**: API endpoints not responding
```bash
# Check if API server is running
curl http://localhost:9000/health

# Check logs for errors
./bin/projectx --log-level=debug
```

#### DAO Issues

**Problem**: Wallet connection fails
1. Ensure MetaMask is installed and unlocked
2. Check network configuration (should be on correct chain)
3. Verify API server is running on correct port
4. Check browser console for JavaScript errors

**Problem**: Transaction signing fails
1. Verify wallet has sufficient balance for fees
2. Check transaction format and required fields
3. Ensure private key is valid and properly formatted
4. Verify signature algorithm compatibility

#### Performance Issues

**Problem**: Slow transaction processing
1. Check system resources (CPU, memory, disk)
2. Verify network connectivity and latency
3. Review transaction pool size and limits
4. Consider increasing hardware resources

**Problem**: High memory usage
1. Monitor Go garbage collection metrics
2. Check for memory leaks in long-running processes
3. Adjust Go runtime parameters (GOGC, GOMEMLIMIT)
4. Profile memory usage with `go tool pprof`

### Debug Mode

Enable debug logging for detailed troubleshooting:

```bash
# Enable debug logging
export LOG_LEVEL=debug
./bin/projectx

# Or use command line flag
./bin/projectx --log-level=debug

# Enable Go runtime debugging
export GODEBUG=gctrace=1
./bin/projectx
```

### Getting Help

1. **Check Documentation**: Review this README and inline code documentation
2. **Search Issues**: Check GitHub issues for similar problems
3. **Enable Debug Logging**: Use debug mode to get detailed error information
4. **Community Support**: Join community discussions and forums
5. **Create Issue**: If problem persists, create a detailed GitHub issue

### Performance Monitoring

```bash
# Monitor system resources
top -p $(pgrep projectx)

# Monitor network connections
netstat -an | grep projectx

# Monitor Go runtime metrics
curl http://localhost:9000/debug/pprof/
```

## 🤝 Contributing

We welcome contributions from the community! Here's how to get involved:

### Ways to Contribute

1. **Code Contributions**: Bug fixes, new features, performance improvements
2. **Documentation**: Improve documentation, tutorials, and examples
3. **Testing**: Write tests, report bugs, test new features
4. **Community**: Help other users, participate in discussions
5. **Feedback**: Provide feedback on features and user experience

### Contribution Process

1. **Fork the Repository**: Create your own fork of the project
2. **Create a Branch**: Create a feature branch for your changes
3. **Make Changes**: Implement your changes with tests and documentation
4. **Test Thoroughly**: Ensure all tests pass and add new tests as needed
5. **Submit Pull Request**: Create a detailed pull request with description
6. **Code Review**: Participate in the code review process
7. **Merge**: Once approved, your changes will be merged

### Development Guidelines

- Follow Go best practices and project conventions
- Write comprehensive tests for new functionality
- Update documentation for any API changes
- Use clear, descriptive commit messages
- Keep pull requests focused and atomic
- Participate constructively in code reviews

### Code of Conduct

- Be respectful and inclusive in all interactions
- Focus on constructive feedback and collaboration
- Help create a welcoming environment for all contributors
- Follow project guidelines and best practices
- Report any inappropriate behavior to project maintainers

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

### Third-Party Licenses

This project uses several third-party libraries and tools:

- **Go Standard Library**: BSD-style license
- **Gorilla WebSocket**: BSD-2-Clause license
- **Echo Framework**: MIT license
- **Logrus**: MIT license
- **Testify**: MIT license
- **IPFS Go API**: MIT license

See `go.mod` for a complete list of dependencies and their versions.

---

## 📞 Support and Community

- **GitHub Issues**: [Report bugs and request features](https://github.com/BOCK-CHAIN/BockChain/issues)
- **Discussions**: [Community discussions and Q&A](https://github.com/BOCK-CHAIN/BockChain/discussions)
- **Documentation**: [Comprehensive project documentation](https://github.com/BOCK-CHAIN/BockChain/wiki)

---

**Built with ❤️ by the ProjectX community**

*Last updated: August 28, 2025*
