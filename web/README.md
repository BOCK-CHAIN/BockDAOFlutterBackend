# Block DAO Web Interface

A responsive web interface for interacting with the Block DAO (Decentralized Autonomous Organization). This interface provides a user-friendly way to participate in governance, manage treasury funds, and interact with the DAO ecosystem.

## Features

### 🏛️ Governance
- **Proposal Management**: Create, view, and vote on governance proposals
- **Multiple Voting Types**: Support for simple majority, quadratic, weighted, and reputation-based voting
- **Real-time Updates**: Live updates via WebSocket connections
- **Proposal Filtering**: Filter proposals by status, type, and other criteria

### 💰 Treasury Management
- **Multi-signature Treasury**: Secure fund management with multiple signature requirements
- **Transaction History**: Complete audit trail of all treasury operations
- **Balance Tracking**: Real-time treasury balance monitoring
- **Spending Proposals**: Create and approve treasury spending requests

### 👥 Member Management
- **Member Directory**: View all DAO members and their statistics
- **Token Balances**: Check governance token holdings
- **Reputation System**: Track member reputation and participation
- **Delegation**: Delegate voting power to trusted representatives

### 🔗 Wallet Integration
- **Multiple Wallet Support**: Compatible with MetaMask and manual key input
- **Secure Transactions**: All transactions are signed locally
- **Balance Monitoring**: Real-time token balance updates
- **Transaction History**: Track your DAO participation

## Getting Started

### Prerequisites
- Block DAO blockchain node running
- DAO API server running on port 9000
- Modern web browser with JavaScript enabled

### Installation

1. **Start the Block DAO Server**:
   ```bash
   cd projectx
   go run main.go
   ```

2. **Access the Web Interface**:
   Open your browser and navigate to:
   ```
   http://localhost:9000
   ```

3. **Connect Your Wallet**:
   - Click "Connect Wallet" in the top navigation
   - Choose between MetaMask or manual key input
   - For testing, use the demo credentials provided

### Demo Mode

For testing and development, visit the demo page:
```
http://localhost:9000/demo.html
```

The demo page includes:
- Test credentials for wallet connection
- API endpoint testing tools
- Sample data generators
- UI component testing

## File Structure

```
web/
├── index.html          # Main application interface
├── demo.html           # Demo and testing interface
├── styles.css          # Complete CSS styling
├── js/
│   ├── app.js          # Main application logic
│   ├── dao-api.js      # API client for DAO operations
│   ├── wallet.js       # Wallet management
│   └── websocket.js    # Real-time event handling
└── README.md           # This file
```

## Usage Guide

### Connecting Your Wallet

1. **MetaMask Connection**:
   - Ensure MetaMask is installed
   - Click "Connect Wallet" → Select MetaMask
   - Approve the connection request

2. **Manual Connection** (for testing):
   - Click "Connect Wallet" → Select Manual
   - Enter your private key and address
   - Use demo credentials for testing

### Creating Proposals

1. Navigate to the "Proposals" section
2. Click "Create Proposal"
3. Fill in the proposal details:
   - **Title**: Brief description of the proposal
   - **Description**: Detailed explanation
   - **Type**: General, Treasury, Technical, or Parameter
   - **Voting Type**: Choose voting mechanism
   - **Duration**: How long voting remains open
   - **Threshold**: Percentage needed to pass

### Voting on Proposals

1. Browse active proposals in the "Proposals" section
2. Click on a proposal to view details
3. Click "Yes", "No", or "Abstain"
4. Specify your vote weight (up to your token balance)
5. Optionally add a reason for your vote
6. Confirm the transaction

### Treasury Operations

1. Navigate to the "Treasury" section
2. View current balance and multi-sig configuration
3. To create a spending request:
   - Click "New Transaction"
   - Enter recipient address and amount
   - Provide a purpose description
   - Submit for multi-sig approval

### Managing Delegations

1. Go to your member profile
2. Click "Delegate Voting Power"
3. Enter the delegate's address
4. Specify delegation duration
5. Confirm the delegation transaction

## API Integration

The web interface communicates with the ProjectX DAO API server. Key endpoints include:

### Proposals
- `GET /dao/proposals` - List all proposals
- `POST /dao/proposal` - Create new proposal
- `POST /dao/vote` - Cast vote on proposal

### Treasury
- `GET /dao/treasury` - Get treasury status
- `POST /dao/treasury/transaction` - Create treasury transaction
- `POST /dao/treasury/sign` - Sign treasury transaction

### Members
- `GET /dao/members` - List DAO members
- `GET /dao/member/:address` - Get member details
- `GET /dao/token/balance/:address` - Get token balance

### Real-time Events
- `WebSocket /dao/events` - Real-time DAO events

## Customization

### Styling
The interface uses CSS custom properties for easy theming:

```css
:root {
  --primary-color: #667eea;
  --secondary-color: #764ba2;
  --success-color: #28a745;
  --danger-color: #dc3545;
  --warning-color: #ffc107;
}
```

### Configuration
Update the API base URL in `js/dao-api.js`:

```javascript
class DAOAPI {
    constructor(baseURL = 'http://localhost:9000') {
        this.baseURL = baseURL;
        // ...
    }
}
```

## Security Considerations

### Development vs Production

⚠️ **Important**: This interface is designed for development and testing. For production use:

1. **Never expose private keys** in the browser
2. **Use proper wallet integrations** (MetaMask, WalletConnect, etc.)
3. **Implement proper authentication** and session management
4. **Use HTTPS** for all communications
5. **Validate all inputs** on both client and server side

### Best Practices

- Always verify transaction details before signing
- Use hardware wallets for significant amounts
- Keep your private keys secure and never share them
- Regularly update your browser and wallet software
- Be cautious of phishing attempts

## Troubleshooting

### Common Issues

1. **Wallet Connection Failed**:
   - Ensure MetaMask is installed and unlocked
   - Check that you're on the correct network
   - Try refreshing the page

2. **API Errors**:
   - Verify the DAO server is running on port 9000
   - Check browser console for detailed error messages
   - Ensure CORS is properly configured

3. **Transaction Failures**:
   - Check your token balance
   - Verify you have sufficient tokens for fees
   - Ensure the proposal/transaction is still active

4. **WebSocket Connection Issues**:
   - Check firewall settings
   - Verify WebSocket support in your browser
   - Try refreshing the connection

### Debug Mode

Enable debug logging by opening browser console and running:
```javascript
localStorage.setItem('dao-debug', 'true');
```

## Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Contributing

To contribute to the web interface:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly using the demo interface
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions:
- Check the troubleshooting section above
- Review the demo interface for examples
- Consult the ProjectX DAO documentation
- Open an issue in the project repository