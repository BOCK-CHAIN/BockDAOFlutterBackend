import 'dart:convert';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:crypto/crypto.dart';
import '../models/wallet.dart';

enum WalletType { manual, walletConnect, metamask, ledger }

class WalletService {
  static const _storage = FlutterSecureStorage();
  static const String _walletKey = 'wallet_info';
  static const String _privateKeyKey = 'private_key';
  static const String _walletTypeKey = 'wallet_type';
  
  WalletInfo? _currentWallet;
  WalletType? _currentWalletType;
  String? _privateKey;
  
  WalletInfo? get currentWallet => _currentWallet;
  bool get isConnected => _currentWallet?.isConnected ?? false;
  WalletType? get walletType => _currentWalletType;

  // Event callbacks
  Function(WalletInfo)? onWalletConnected;
  Function()? onWalletDisconnected;
  Function(int balance)? onBalanceUpdated;

  // Initialize wallet service and load stored wallet
  Future<void> initialize() async {
    await _loadStoredWallet();
  }

  // Detect available wallets
  Map<WalletType, bool> getAvailableWallets() {
    return {
      WalletType.manual: true, // Always available
      WalletType.walletConnect: true, // WalletConnect is available
      WalletType.metamask: false, // Not directly available on mobile
      WalletType.ledger: false, // Would require additional setup
    };
  }

  // Connect wallet with different providers
  Future<WalletInfo> connectWallet(WalletType walletType, {Map<String, dynamic>? params}) async {
    switch (walletType) {
      case WalletType.manual:
        return await _connectManualWallet(params!);
      case WalletType.walletConnect:
        return await _connectWalletConnect();
      case WalletType.metamask:
        throw UnsupportedError('MetaMask not supported on mobile');
      case WalletType.ledger:
        throw UnsupportedError('Ledger not implemented yet');
    }
  }

  // Manual wallet connection
  Future<WalletInfo> _connectManualWallet(Map<String, dynamic> params) async {
    final privateKeyHex = params['privateKey'] as String;
    final address = params['address'] as String;

    if (!_isValidPrivateKey(privateKeyHex)) {
      throw ArgumentError('Invalid private key format');
    }

    if (!_isValidAddress(address)) {
      throw ArgumentError('Invalid address format');
    }

    // Store private key
    _privateKey = privateKeyHex;
    final publicKey = address; // Use provided address

    final walletInfo = WalletInfo(
      address: address,
      publicKey: publicKey,
      tokenBalance: 0,
      stakedBalance: 0,
      reputation: 0,
      joinedAt: DateTime.now(),
      lastActive: DateTime.now(),
      isConnected: true,
    );

    // Store wallet info securely
    await _storage.write(key: _walletKey, value: json.encode(walletInfo.toJson()));
    await _storage.write(key: _privateKeyKey, value: privateKeyHex);
    await _storage.write(key: _walletTypeKey, value: walletType.toString());
    
    _currentWallet = walletInfo;
    _currentWalletType = WalletType.manual;

    onWalletConnected?.call(walletInfo);
    return walletInfo;
  }

  // WalletConnect integration (simplified for now)
  Future<WalletInfo> _connectWalletConnect() async {
    throw UnsupportedError('WalletConnect integration not fully implemented yet');
  }

  // Disconnect wallet
  Future<void> disconnectWallet() async {
    await _storage.delete(key: _walletKey);
    await _storage.delete(key: _privateKeyKey);
    await _storage.delete(key: _walletTypeKey);
    
    _currentWallet = null;
    _currentWalletType = null;
    _privateKey = null;

    onWalletDisconnected?.call();
  }

  // Sign DAO transaction
  Future<Map<String, dynamic>> signDAOTransaction(Map<String, dynamic> transaction) async {
    if (!isConnected) {
      throw Exception('Wallet not connected');
    }

    // Validate transaction data
    if (transaction.isEmpty) {
      throw Exception('Transaction data cannot be empty');
    }

    // Check for required fields based on transaction type
    if (!transaction.containsKey('type') && !transaction.containsKey('fee')) {
      throw Exception('Transaction must contain type or fee field');
    }

    switch (_currentWalletType) {
      case WalletType.manual:
        return await _signWithPrivateKey(transaction);
      case WalletType.walletConnect:
        return await _signWithWalletConnect(transaction);
      default:
        throw Exception('Unsupported wallet type for signing');
    }
  }

  // Sign with private key (manual wallet)
  Future<Map<String, dynamic>> _signWithPrivateKey(Map<String, dynamic> transaction) async {
    if (_privateKey == null) {
      throw Exception('Private key not available');
    }

    try {
      // Create transaction hash
      final transactionHash = _calculateTransactionHash(transaction);
      
      // Sign the hash
      final signature = await _signHash(transactionHash, _privateKey!);
      
      return {
        'signature': signature,
        'transactionHash': transactionHash,
        'signer': _currentWallet!.address,
        'method': 'private_key',
        'timestamp': DateTime.now().millisecondsSinceEpoch,
      };

    } catch (e) {
      throw Exception('Private key signing failed: $e');
    }
  }

  // Sign with WalletConnect (simplified for now)
  Future<Map<String, dynamic>> _signWithWalletConnect(Map<String, dynamic> transaction) async {
    throw UnsupportedError('WalletConnect signing not fully implemented yet');
  }

  // Calculate transaction hash
  String _calculateTransactionHash(Map<String, dynamic> transaction) {
    final transactionString = json.encode(transaction);
    final bytes = utf8.encode(transactionString);
    final digest = sha256.convert(bytes);
    return digest.toString();
  }

  // Sign hash with private key (simplified implementation)
  Future<String> _signHash(String hash, String privateKey) async {
    // Simplified signing - in production use proper ECDSA
    final combined = hash + privateKey;
    final bytes = utf8.encode(combined);
    final digest = sha256.convert(bytes);
    return digest.toString();
  }

  // Generate test wallet for development
  Map<String, String> generateTestWallet() {
    // Generate random private key (32 bytes = 64 hex chars)
    final random = List.generate(32, (i) => (DateTime.now().millisecondsSinceEpoch + i) % 256);
    final privateKeyHex = random.map((b) => b.toRadixString(16).padLeft(2, '0')).join('');
    
    // Generate random address (20 bytes = 40 hex chars)
    final addressRandom = List.generate(20, (i) => (DateTime.now().microsecondsSinceEpoch + i) % 256);
    final address = '0x${addressRandom.map((b) => b.toRadixString(16).padLeft(2, '0')).join('')}';
    
    // Generate public key (simplified)
    final publicKey = privateKeyHex.substring(0, 64);

    return {
      'privateKey': privateKeyHex,
      'address': address,
      'publicKey': publicKey,
    };
  }

  // Broadcast transaction to network
  Future<Map<String, dynamic>> broadcastTransaction(Map<String, dynamic> signedTransaction) async {
    try {
      // In a real implementation, this would send to the ProjectX DAO API
      // For now, simulate the broadcast
      await Future.delayed(const Duration(seconds: 1));
      
      return {
        'success': true,
        'transactionHash': signedTransaction['transactionHash'],
        'blockHeight': DateTime.now().millisecondsSinceEpoch,
        'timestamp': DateTime.now().toIso8601String(),
      };

    } catch (e) {
      throw Exception('Transaction broadcast failed: $e');
    }
  }

  // Update wallet balance from API
  Future<void> refreshWalletBalance() async {
    if (!isConnected) return;

    try {
      // In a real implementation, fetch from API
      // For now, simulate balance update
      final newBalance = DateTime.now().millisecondsSinceEpoch % 10000;
      
      if (_currentWallet != null) {
        _currentWallet = WalletInfo(
          address: _currentWallet!.address,
          publicKey: _currentWallet!.publicKey,
          tokenBalance: newBalance,
          stakedBalance: _currentWallet!.stakedBalance,
          reputation: _currentWallet!.reputation,
          joinedAt: _currentWallet!.joinedAt,
          lastActive: DateTime.now(),
          isConnected: _currentWallet!.isConnected,
        );

        // Update stored wallet info
        await _storage.write(key: _walletKey, value: json.encode(_currentWallet!.toJson()));
        
        onBalanceUpdated?.call(newBalance);
      }

    } catch (e) {
      print('Balance update failed: $e');
    }
  }

  // Load stored wallet from secure storage
  Future<void> _loadStoredWallet() async {
    final walletData = await _storage.read(key: _walletKey);
    final walletTypeData = await _storage.read(key: _walletTypeKey);
    
    if (walletData != null && walletTypeData != null) {
      try {
        final walletJson = json.decode(walletData);
        _currentWallet = WalletInfo.fromJson(walletJson);
        
        // Parse wallet type
        final walletTypeString = walletTypeData.split('.').last;
        _currentWalletType = WalletType.values.firstWhere(
          (type) => type.toString().split('.').last == walletTypeString,
        );

        // Restore private key for manual wallet
        if (_currentWalletType == WalletType.manual) {
          final privateKeyHex = await _storage.read(key: _privateKeyKey);
          if (privateKeyHex != null) {
            _privateKey = privateKeyHex;
          }
        }

        // For WalletConnect, we'd need to restore the session
        // This is more complex and would require session persistence

      } catch (e) {
        print('Failed to load stored wallet: $e');
        await disconnectWallet();
      }
    }
  }

  // Update wallet address (for WalletConnect account changes)
  void _updateWalletAddress(String newAddress) {
    if (_currentWallet != null) {
      _currentWallet = WalletInfo(
        address: newAddress,
        publicKey: _currentWallet!.publicKey,
        tokenBalance: _currentWallet!.tokenBalance,
        stakedBalance: _currentWallet!.stakedBalance,
        reputation: _currentWallet!.reputation,
        joinedAt: _currentWallet!.joinedAt,
        lastActive: DateTime.now(),
        isConnected: _currentWallet!.isConnected,
      );
    }
  }

  // Show WalletConnect URI (implement in UI)
  void _showWalletConnectUri(String uri) {
    // This would be implemented in the UI layer
    // Could show QR code or deep link
    print('WalletConnect URI: $uri');
  }

  // Validation helpers
  bool _isValidPrivateKey(String key) {
    return RegExp(r'^[0-9a-fA-F]{64}$').hasMatch(key);
  }

  bool _isValidAddress(String address) {
    return RegExp(r'^[0-9a-fA-F]{40}$').hasMatch(address.replaceFirst('0x', ''));
  }

  // Update wallet balance with specific values
  void updateWalletBalance({
    required int tokenBalance,
    required int stakedBalance,
    required int reputation,
  }) {
    if (_currentWallet != null) {
      _currentWallet = WalletInfo(
        address: _currentWallet!.address,
        publicKey: _currentWallet!.publicKey,
        tokenBalance: tokenBalance,
        stakedBalance: stakedBalance,
        reputation: reputation,
        joinedAt: _currentWallet!.joinedAt,
        lastActive: DateTime.now(),
        isConnected: _currentWallet!.isConnected,
      );
      onBalanceUpdated?.call(tokenBalance);
    }
  }

  // Utility methods
  String formatAddress(String address) {
    if (address.length < 10) return address;
    return '${address.substring(0, 6)}...${address.substring(address.length - 4)}';
  }

  // Get wallet info
  Map<String, dynamic> getWalletInfo() {
    return {
      'isConnected': isConnected,
      'walletType': _currentWalletType?.toString(),
      'address': _currentWallet?.address,
      'balance': _currentWallet?.tokenBalance ?? 0,
      'reputation': _currentWallet?.reputation ?? 0,
    };
  }

  // Cleanup resources
  void dispose() {
    // Clean up any resources
  }
}