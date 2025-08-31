import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/dao_provider.dart';

class WalletScreen extends StatefulWidget {
  const WalletScreen({super.key});

  @override
  State<WalletScreen> createState() => _WalletScreenState();
}

class _WalletScreenState extends State<WalletScreen> {
  final _addressController = TextEditingController();
  final _publicKeyController = TextEditingController();
  final _privateKeyController = TextEditingController();
  bool _isConnecting = false;

  @override
  void dispose() {
    _addressController.dispose();
    _publicKeyController.dispose();
    _privateKeyController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Wallet'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
      ),
      body: Consumer<DAOProvider>(
        builder: (context, daoProvider, child) {
          if (daoProvider.isWalletConnected) {
            return _buildConnectedWallet(daoProvider);
          } else {
            return _buildWalletConnection();
          }
        },
      ),
    );
  }

  Widget _buildConnectedWallet(DAOProvider daoProvider) {
    final wallet = daoProvider.walletInfo!;
    
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Wallet Status Card
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(Icons.check_circle, color: Colors.green),
                      const SizedBox(width: 8),
                      Text(
                        'Wallet Connected',
                        style: Theme.of(context).textTheme.headlineSmall,
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  _buildInfoRow('Address', _truncateAddress(wallet.address)),
                  _buildInfoRow('Public Key', _truncateAddress(wallet.publicKey)),
                  _buildInfoRow('Joined', _formatDate(wallet.joinedAt)),
                  _buildInfoRow('Last Active', _formatDate(wallet.lastActive)),
                ],
              ),
            ),
          ),
          
          const SizedBox(height: 16),
          
          // Token Balance Card
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Token Balance',
                    style: Theme.of(context).textTheme.headlineSmall,
                  ),
                  const SizedBox(height: 16),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            '${wallet.tokenBalance}',
                            style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                              fontWeight: FontWeight.bold,
                              color: Colors.blue,
                            ),
                          ),
                          const Text('Available Tokens'),
                        ],
                      ),
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.end,
                        children: [
                          Text(
                            '${wallet.stakedBalance}',
                            style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                              fontWeight: FontWeight.bold,
                              color: Colors.green,
                            ),
                          ),
                          const Text('Staked Tokens'),
                        ],
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          
          const SizedBox(height: 16),
          
          // Reputation Card
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Reputation',
                    style: Theme.of(context).textTheme.headlineSmall,
                  ),
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      Icon(Icons.star, color: Colors.amber, size: 32),
                      const SizedBox(width: 8),
                      Text(
                        '${wallet.reputation}',
                        style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                          fontWeight: FontWeight.bold,
                          color: Colors.amber,
                        ),
                      ),
                      const SizedBox(width: 8),
                      Text(
                        'points',
                        style: Theme.of(context).textTheme.bodyLarge,
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  LinearProgressIndicator(
                    value: (wallet.reputation % 1000) / 1000,
                    backgroundColor: Colors.grey[300],
                    valueColor: AlwaysStoppedAnimation<Color>(Colors.amber),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    'Next level: ${1000 - (wallet.reputation % 1000)} points',
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                ],
              ),
            ),
          ),
          
          const SizedBox(height: 24),
          
          // Disconnect Button
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: () async {
                await daoProvider.disconnectWallet();
                if (mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Wallet disconnected')),
                  );
                }
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.red,
                foregroundColor: Colors.white,
              ),
              child: const Text('Disconnect Wallet'),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildWalletConnection() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Connect Your Wallet',
            style: Theme.of(context).textTheme.headlineMedium,
          ),
          const SizedBox(height: 8),
          Text(
            'Enter your wallet credentials to connect to the ProjectX DAO.',
            style: Theme.of(context).textTheme.bodyMedium,
          ),
          const SizedBox(height: 24),
          
          // Address Field
          TextField(
            controller: _addressController,
            decoration: const InputDecoration(
              labelText: 'Wallet Address',
              border: OutlineInputBorder(),
              prefixIcon: Icon(Icons.account_balance_wallet),
            ),
          ),
          const SizedBox(height: 16),
          
          // Public Key Field
          TextField(
            controller: _publicKeyController,
            decoration: const InputDecoration(
              labelText: 'Public Key',
              border: OutlineInputBorder(),
              prefixIcon: Icon(Icons.key),
            ),
          ),
          const SizedBox(height: 16),
          
          // Private Key Field
          TextField(
            controller: _privateKeyController,
            decoration: const InputDecoration(
              labelText: 'Private Key',
              border: OutlineInputBorder(),
              prefixIcon: Icon(Icons.lock),
            ),
            obscureText: true,
          ),
          const SizedBox(height: 24),
          
          // Connect Button
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: _isConnecting ? null : _connectWallet,
              child: _isConnecting
                  ? const Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        ),
                        SizedBox(width: 8),
                        Text('Connecting...'),
                      ],
                    )
                  : const Text('Connect Wallet'),
            ),
          ),
          
          const SizedBox(height: 16),
          
          // Demo Button
          SizedBox(
            width: double.infinity,
            child: OutlinedButton(
              onPressed: _connectDemoWallet,
              child: const Text('Use Demo Wallet'),
            ),
          ),
          
          const SizedBox(height: 24),
          
          // Security Notice
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.amber[50],
              border: Border.all(color: Colors.amber),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Row(
              children: [
                Icon(Icons.warning, color: Colors.amber[800]),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Never share your private key with anyone. This is a demo implementation.',
                    style: TextStyle(color: Colors.amber[800]),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildInfoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 100,
            child: Text(
              '$label:',
              style: const TextStyle(fontWeight: FontWeight.w500),
            ),
          ),
          Expanded(
            child: Text(value),
          ),
        ],
      ),
    );
  }

  String _truncateAddress(String address) {
    if (address.length <= 16) return address;
    return '${address.substring(0, 8)}...${address.substring(address.length - 8)}';
  }

  String _formatDate(DateTime date) {
    return '${date.day}/${date.month}/${date.year}';
  }

  Future<void> _connectWallet() async {
    if (_addressController.text.isEmpty ||
        _publicKeyController.text.isEmpty ||
        _privateKeyController.text.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please fill in all fields')),
      );
      return;
    }

    setState(() {
      _isConnecting = true;
    });

    try {
      await context.read<DAOProvider>().connectWallet(
        address: _addressController.text,
        publicKey: _publicKeyController.text,
        privateKey: _privateKeyController.text,
      );
      
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Wallet connected successfully!')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to connect wallet: $e')),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _isConnecting = false;
        });
      }
    }
  }

  void _connectDemoWallet() {
    _addressController.text = 'demo_address_1234567890abcdef';
    _publicKeyController.text = 'demo_public_key_abcdef1234567890';
    _privateKeyController.text = 'demo_private_key_1234567890abcdef';
    _connectWallet();
  }
}