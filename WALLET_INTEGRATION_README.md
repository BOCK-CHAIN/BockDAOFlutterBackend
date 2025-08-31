# ProjectX DAO Wallet Integration

This document describes the comprehensive wallet integration system implemented for the ProjectX DAO, supporting multiple wallet providers and secure transaction signing across web, mobile, and backend platforms.

## Overview

The wallet integration system provides:
- **Multi-platform support**: Web (JavaScript), Mobile (Flutter), and Backend (Go)
- **Multiple wallet providers**: MetaMask, WalletConnect, Manual key input, and Ledger hardware wallets
- **Secure transaction signing**: Proper cryptographic signing for all DAO transaction types
- **Comprehensive testing**: Unit tests, integration tests, and security tests
- **API integration**: RESTful endpoints for wallet operations

## Architecture

### Components

1. **Web Integration** (`projectx/web/js/wallet-integration.js`)
   - Browser-based wallet connections
   - MetaMask integration
   - WalletConnect support
   - Manual key management
   - Transaction signing and broadcasting

2. **Mobile Integration** (`projectx/mobile/lib/services/wallet_service.dart`)
   - Flutter-based wallet service
   - Secure storage for keys
   - Cross-platform wallet support
   - DAO transaction signing

3. **Backend Integration** (`projectx/dao/wallet_integration.go`)
   - Go-based wallet management
   - Transaction validation
   - Signature verification
   - Connection management

4. **API Endpoints** (`projectx/api/dao_server.go`)
   - RESTful wallet operations
   - Transaction broadcasting
   - Real-time WebSocket events

## Supported Wallets

### 1. MetaMask
- **Platform**: Web browsers
- **Features**: Account management, transaction signing, chain switching
- **Security**: Browser extension security model
- **Usage**: Automatic detection and connection

### 2. WalletConnect
- **Platform**: Mobile and desktop wallets
- **Features**: QR code connection, deep linking, mobile wallet support
- **Security**: End-to-end encryption
- **Usage**: QR code scanning or deep links

### 3. Manual Key Input
- **Platform**: All platforms (development only)
- **Features**: Direct private key input, test wallet generation
- **Security**: Session storage (development only)
- **Usage**: Development and testing purposes

### 4. Ledger Hardware Wallet
- **Platform**: Web and desktop (planned)
- **Features**: Hardware security, transaction confirmation
- **Security**: Hardware-based key storage
- **Status**: Framework implemented, full integration pending

## API Endpoints

### Wallet Management
```
POST /dao/wallet/connect          - Connect a wallet
POST /dao/wallet/disconnect       - Disconnect a wallet
GET  /dao/wallet/info/:address    - Get wallet information
GET  /dao/wallet/connections      - List active connections
```

### Transaction Operations
```
POST /dao/wallet/sign             - Sign a transaction
POST /dao/wallet/broadcast        - Broadcast signed transaction
POST /dao/wallet/verify           - Verify transaction signature
```

### Utilities
```
POST /dao/wallet/generate-test    - Generate test wallet
GET  /dao/wallet/supported        - List supported wallets
```

### WebSocket Events
```
ws://localhost:9000/dao/events    - Real-time wallet events
```

## Usage Examples

### Web Integration

```javascript
// Initialize wallet integration
const walletIntegration = new WalletIntegration();

// Connect MetaMask
const result = await walletIntegration.connect('metamask');
if (result.success) {
    console.log('Connected to:', result.address);
}

// Sign a DAO transaction
const transaction = {
    type: 'proposal',
    fee: 1000,
    title: 'Test Proposal',
    description: 'This is a test proposal'
};

const signedTx = await walletIntegration.signTransaction(transaction);
console.log('Signed transaction:', signedTx);

// Broadcast transaction
const broadcastResult = await walletIntegration.broadcastTransaction(signedTx);
console.log('Transaction hash:', broadcastResult.transactionHash);
```

### Mobile Integration

```dart
// Initialize wallet service
final walletService = WalletService();
await walletService.initialize();

// Connect manual wallet
final testWallet = walletService.generateTestWallet();
final walletInfo = await walletService.connectWallet(
  WalletType.manual,
  params: {
    'privateKey': testWallet['privateKey']!,
    'address': testWallet['address']!,
  },
);

// Sign DAO transaction
final transaction = {
  'type': 'vote',
  'fee': 500,
  'proposalId': 'proposal_123',
  'choice': 'yes',
};

final signedTx = await walletService.signDAOTransaction(transaction);
print('Signed transaction: $signedTx');

// Broadcast transaction
final result = await walletService.broadcastTransaction(signedTx);
print('Broadcast result: $result');
```

### Backend Integration

```go
// Initialize wallet connection manager
manager := dao.NewWalletConnectionManager()

// Handle wallet connection
connection, err := manager.HandleWalletConnection(
    dao.WalletProviderManual,
    "address_hex",
    "public_key_hex",
    "",
)

// Sign transaction
signedTx, err := manager.HandleTransactionSigning(
    "address_hex",
    transaction,
    "signature_hex",
)

// Verify signed transaction
service := dao.NewWalletIntegrationService()
err = service.VerifySignedTransaction(signedTx)
```

## Security Features

### 1. Key Management
- **Secure Storage**: Private keys stored in secure storage (mobile) or session storage (web development)
- **Key Validation**: Strict validation of private key and address formats
- **Memory Protection**: Keys cleared from memory on disconnect

### 2. Transaction Security
- **Signature Verification**: All transactions verified before processing
- **Hash Integrity**: Transaction hashes calculated and verified
- **Replay Protection**: Nonce-based replay attack prevention

### 3. Connection Security
- **Provider Validation**: Wallet provider verification
- **Session Management**: Secure session handling
- **Event Logging**: Comprehensive audit logging

### 4. Development vs Production
- **Environment Separation**: Clear distinction between development and production features
- **Warning Messages**: Security warnings for development-only features
- **Key Exposure Prevention**: No private key exposure in production builds

## Testing

### Test Coverage
- **Unit Tests**: Individual component testing
- **Integration Tests**: End-to-end workflow testing
- **Security Tests**: Attack vector validation
- **Performance Tests**: Scalability and performance validation

### Running Tests

#### Go Tests
```bash
cd projectx
go test -v ./dao -run TestWalletIntegration
```

#### Flutter Tests
```bash
cd projectx/mobile
flutter test test/wallet_service_test.dart
```

#### JavaScript Tests
```javascript
// In browser console
const tester = new WalletIntegrationTests();
await tester.runAllTests();
```

## Configuration

### Web Configuration
```javascript
// Update API base URL in dao-api.js
const daoAPI = new DAOAPI('http://localhost:9000');

// Enable debug mode
localStorage.setItem('dao-debug', 'true');
```

### Mobile Configuration
```yaml
# pubspec.yaml dependencies
dependencies:
  walletconnect_dart: ^0.0.11
  web3dart: ^2.7.1
  pointycastle: ^3.7.4
  flutter_secure_storage: ^9.0.0
```

### Backend Configuration
```go
// Initialize with custom settings
service := dao.NewWalletIntegrationService()
manager := dao.NewWalletConnectionManager()

// Set up cleanup routine
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    for range ticker.C {
        service.CleanupInactiveConnections(24 * time.Hour)
    }
}()
```

## Error Handling

### Common Errors
- `Wallet not connected`: Ensure wallet is connected before operations
- `Invalid signature`: Check transaction format and signing process
- `Unsupported wallet type`: Verify wallet provider is supported
- `Transaction validation failed`: Check transaction structure and required fields

### Error Codes
- `4001`: Insufficient tokens
- `4002`: Proposal not found
- `4003`: Voting closed
- `4004`: Unauthorized
- `4005`: Invalid signature
- `4006`: Quorum not met
- `4007`: Treasury insufficient

## Development Guidelines

### Adding New Wallet Providers
1. Implement provider-specific connection logic
2. Add transaction signing methods
3. Create validation functions
4. Add comprehensive tests
5. Update documentation

### Security Best Practices
1. Never expose private keys in logs
2. Use secure storage for sensitive data
3. Validate all inputs thoroughly
4. Implement proper error handling
5. Regular security audits

### Testing Requirements
1. Unit tests for all components
2. Integration tests for workflows
3. Security tests for attack vectors
4. Performance tests for scalability
5. Cross-platform compatibility tests

## Troubleshooting

### Connection Issues
1. Check wallet provider availability
2. Verify network connectivity
3. Ensure correct chain/network
4. Check browser permissions (web)

### Signing Issues
1. Verify wallet is connected
2. Check transaction format
3. Ensure sufficient balance
4. Validate signature format

### API Issues
1. Check server status
2. Verify endpoint URLs
3. Check CORS settings
4. Validate request format

## Future Enhancements

### Planned Features
1. **Hardware Wallet Support**: Full Ledger integration
2. **Multi-Signature Wallets**: Enhanced security for treasury operations
3. **Biometric Authentication**: Mobile biometric unlock
4. **Offline Signing**: Air-gapped transaction signing
5. **Advanced Analytics**: Wallet usage analytics

### Integration Roadmap
1. **Phase 1**: Core wallet providers (✅ Completed)
2. **Phase 2**: Hardware wallet integration
3. **Phase 3**: Advanced security features
4. **Phase 4**: Analytics and monitoring
5. **Phase 5**: Cross-chain support

## Support

For issues and questions:
1. Check the troubleshooting section
2. Review test files for examples
3. Consult API documentation
4. Open GitHub issues for bugs
5. Join community discussions

## License

This wallet integration system is part of the ProjectX DAO project and follows the same licensing terms as the main project.