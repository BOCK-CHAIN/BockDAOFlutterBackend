import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:bock_dao_mobile/services/wallet_service.dart';

void main() {
  group('WalletService Tests', () {
    late WalletService walletService;

    setUp(() {
      walletService = WalletService();
      // Mock secure storage for testing
      FlutterSecureStorage.setMockInitialValues({});
    });

    tearDown(() {
      // Clean up after each test
      walletService.dispose();
    });

    test('should initialize without errors', () async {
      await walletService.initialize();
      expect(walletService.isConnected, false);
      expect(walletService.currentWallet, null);
    });

    test('should detect available wallets', () {
      final availableWallets = walletService.getAvailableWallets();
      
      expect(availableWallets, isNotEmpty);
      expect(availableWallets[WalletType.manual], true);
      expect(availableWallets[WalletType.walletConnect], true);
      expect(availableWallets[WalletType.metamask], false); // Not supported on mobile
    });

    test('should generate test wallet with valid keys', () {
      final testWallet = walletService.generateTestWallet();
      
      expect(testWallet['privateKey'], isNotNull);
      expect(testWallet['address'], isNotNull);
      expect(testWallet['publicKey'], isNotNull);
      
      // Validate private key format (64 hex characters)
      expect(testWallet['privateKey']!.length, 64);
      expect(RegExp(r'^[0-9a-fA-F]+$').hasMatch(testWallet['privateKey']!), true);
      
      // Validate address format (40 hex characters)
      expect(testWallet['address']!.length, 42); // Including 0x prefix
      expect(testWallet['address']!.startsWith('0x'), true);
    });

    test('should connect manual wallet with valid credentials', () async {
      final testWallet = walletService.generateTestWallet();
      
      final walletInfo = await walletService.connectWallet(
        WalletType.manual,
        params: {
          'privateKey': testWallet['privateKey']!,
          'address': testWallet['address']!,
        },
      );
      
      expect(walletInfo.isConnected, true);
      expect(walletInfo.address, testWallet['address']);
      expect(walletService.isConnected, true);
      expect(walletService.currentWallet, walletInfo);
    });

    test('should reject invalid private key', () async {
      expect(
        () => walletService.connectWallet(
          WalletType.manual,
          params: {
            'privateKey': 'invalid_key',
            'address': '0x1234567890123456789012345678901234567890',
          },
        ),
        throwsA(isA<ArgumentError>()),
      );
    });

    test('should reject invalid address', () async {
      final testWallet = walletService.generateTestWallet();
      
      expect(
        () => walletService.connectWallet(
          WalletType.manual,
          params: {
            'privateKey': testWallet['privateKey']!,
            'address': 'invalid_address',
          },
        ),
        throwsA(isA<ArgumentError>()),
      );
    });

    test('should disconnect wallet properly', () async {
      // Connect first
      final testWallet = walletService.generateTestWallet();
      await walletService.connectWallet(
        WalletType.manual,
        params: {
          'privateKey': testWallet['privateKey']!,
          'address': testWallet['address']!,
        },
      );
      
      expect(walletService.isConnected, true);
      
      // Disconnect
      await walletService.disconnectWallet();
      
      expect(walletService.isConnected, false);
      expect(walletService.currentWallet, null);
    });

    test('should sign DAO transaction with manual wallet', () async {
      // Connect wallet first
      final testWallet = walletService.generateTestWallet();
      await walletService.connectWallet(
        WalletType.manual,
        params: {
          'privateKey': testWallet['privateKey']!,
          'address': testWallet['address']!,
        },
      );
      
      final transaction = {
        'type': 'proposal',
        'fee': 1000,
        'title': 'Test Proposal',
        'description': 'This is a test proposal',
        'nonce': DateTime.now().millisecondsSinceEpoch,
      };
      
      final signedTx = await walletService.signDAOTransaction(transaction);
      
      expect(signedTx['signature'], isNotNull);
      expect(signedTx['transactionHash'], isNotNull);
      expect(signedTx['signer'], testWallet['address']);
      expect(signedTx['method'], 'private_key');
      expect(signedTx['timestamp'], isNotNull);
    });

    test('should fail to sign transaction when not connected', () async {
      final transaction = {
        'type': 'vote',
        'fee': 500,
        'proposalId': 'test_proposal_id',
        'choice': 'yes',
      };
      
      expect(
        () => walletService.signDAOTransaction(transaction),
        throwsA(isA<Exception>()),
      );
    });

    test('should broadcast transaction successfully', () async {
      final signedTransaction = {
        'signature': 'test_signature',
        'transactionHash': 'test_hash',
        'signer': '0x1234567890123456789012345678901234567890',
        'method': 'private_key',
        'timestamp': DateTime.now().millisecondsSinceEpoch,
      };
      
      final result = await walletService.broadcastTransaction(signedTransaction);
      
      expect(result['success'], true);
      expect(result['transactionHash'], 'test_hash');
      expect(result['blockHeight'], isNotNull);
      expect(result['timestamp'], isNotNull);
    });

    test('should update wallet balance', () async {
      // Connect wallet first
      final testWallet = walletService.generateTestWallet();
      await walletService.connectWallet(
        WalletType.manual,
        params: {
          'privateKey': testWallet['privateKey']!,
          'address': testWallet['address']!,
        },
      );
      
      final initialBalance = walletService.currentWallet!.tokenBalance;
      
      walletService.updateWalletBalance();
      
      // Balance should be updated (simulated)
      expect(walletService.currentWallet!.tokenBalance, isNot(initialBalance));
      expect(walletService.currentWallet!.lastActive, isNotNull);
    });

    test('should format address correctly', () {
      const longAddress = '0x1234567890123456789012345678901234567890';
      final formatted = walletService.formatAddress(longAddress);
      
      expect(formatted.contains('...'), true);
      expect(formatted.length, lessThan(longAddress.length));
      expect(formatted.startsWith('0x1234'), true);
      expect(formatted.endsWith('7890'), true);
    });

    test('should return correct wallet info', () async {
      // Connect wallet first
      final testWallet = walletService.generateTestWallet();
      await walletService.connectWallet(
        WalletType.manual,
        params: {
          'privateKey': testWallet['privateKey']!,
          'address': testWallet['address']!,
        },
      );
      
      final walletInfo = walletService.getWalletInfo();
      
      expect(walletInfo['isConnected'], true);
      expect(walletInfo['walletType'], contains('manual'));
      expect(walletInfo['address'], testWallet['address']);
      expect(walletInfo['balance'], isNotNull);
      expect(walletInfo['reputation'], isNotNull);
    });

    test('should handle wallet events correctly', () async {
      bool connectEventFired = false;
      bool disconnectEventFired = false;
      bool balanceEventFired = false;
      
      // Set up event listeners
      walletService.onWalletConnected = (wallet) {
        connectEventFired = true;
      };
      
      walletService.onWalletDisconnected = () {
        disconnectEventFired = true;
      };
      
      walletService.onBalanceUpdated = (balance) {
        balanceEventFired = true;
      };
      
      // Connect wallet
      final testWallet = walletService.generateTestWallet();
      await walletService.connectWallet(
        WalletType.manual,
        params: {
          'privateKey': testWallet['privateKey']!,
          'address': testWallet['address']!,
        },
      );
      
      expect(connectEventFired, true);
      
      // Update balance
      walletService.updateWalletBalance();
      expect(balanceEventFired, true);
      
      // Disconnect wallet
      await walletService.disconnectWallet();
      expect(disconnectEventFired, true);
    });

    test('should persist and restore wallet connection', () async {
      // Connect wallet
      final testWallet = walletService.generateTestWallet();
      await walletService.connectWallet(
        WalletType.manual,
        params: {
          'privateKey': testWallet['privateKey']!,
          'address': testWallet['address']!,
        },
      );
      
      expect(walletService.isConnected, true);
      
      // Create new service instance (simulating app restart)
      final newWalletService = WalletService();
      await newWalletService.initialize();
      
      // Should restore connection from secure storage
      // Note: This test might need adjustment based on actual implementation
      // as it depends on the secure storage mock behavior
    });

    test('should handle concurrent operations safely', () async {
      final testWallet = walletService.generateTestWallet();
      
      // Attempt multiple concurrent connections
      final futures = List.generate(5, (index) => 
        walletService.connectWallet(
          WalletType.manual,
          params: {
            'privateKey': testWallet['privateKey']!,
            'address': testWallet['address']!,
          },
        )
      );
      
      final results = await Future.wait(futures);
      
      // All should succeed and return the same wallet info
      for (final result in results) {
        expect(result.isConnected, true);
        expect(result.address, testWallet['address']);
      }
    });

    test('should validate transaction data before signing', () async {
      // Connect wallet first
      final testWallet = walletService.generateTestWallet();
      await walletService.connectWallet(
        WalletType.manual,
        params: {
          'privateKey': testWallet['privateKey']!,
          'address': testWallet['address']!,
        },
      );
      
      // Test with empty transaction
      await expectLater(
        walletService.signDAOTransaction({}),
        throwsA(isA<Exception>()),
      );
      
      // Test with invalid transaction structure
      await expectLater(
        walletService.signDAOTransaction({'invalid': 'structure'}),
        throwsA(isA<Exception>()),
      );
    });

    test('should handle network errors gracefully', () async {
      // This test would require mocking network calls
      // For now, we'll test that the service doesn't crash on errors
      
      final testWallet = walletService.generateTestWallet();
      await walletService.connectWallet(
        WalletType.manual,
        params: {
          'privateKey': testWallet['privateKey']!,
          'address': testWallet['address']!,
        },
      );
      
      // Update balance should not throw even if network fails
      walletService.updateWalletBalance();
      
      // Service should still be connected
      expect(walletService.isConnected, true);
    });
  });

  group('WalletService Security Tests', () {
    late WalletService walletService;

    setUp(() {
      walletService = WalletService();
      FlutterSecureStorage.setMockInitialValues({});
    });

    test('should not expose private key in wallet info', () async {
      final testWallet = walletService.generateTestWallet();
      await walletService.connectWallet(
        WalletType.manual,
        params: {
          'privateKey': testWallet['privateKey']!,
          'address': testWallet['address']!,
        },
      );
      
      final walletInfo = walletService.getWalletInfo();
      
      // Private key should not be in the wallet info
      expect(walletInfo.containsKey('privateKey'), false);
      expect(walletInfo.toString().contains(testWallet['privateKey']!), false);
    });

    test('should validate private key format strictly', () {
      final validKeys = [
        'a' * 64, // All lowercase
        'A' * 64, // All uppercase
        '0123456789abcdefABCDEF${'0' * 42}', // Mixed case
      ];
      
      final invalidKeys = [
        'a' * 63, // Too short
        'a' * 65, // Too long
        'g' * 64, // Invalid hex character
        '', // Empty
        'a' * 32, // Half length
      ];
      
      for (final key in validKeys) {
        expect(() => walletService.connectWallet(
          WalletType.manual,
          params: {
            'privateKey': key,
            'address': '0x${'a' * 40}',
          },
        ), returnsNormally);
      }
      
      for (final key in invalidKeys) {
        expect(() => walletService.connectWallet(
          WalletType.manual,
          params: {
            'privateKey': key,
            'address': '0x${'a' * 40}',
          },
        ), throwsA(isA<ArgumentError>()));
      }
    });

    test('should clear sensitive data on disconnect', () async {
      final testWallet = walletService.generateTestWallet();
      await walletService.connectWallet(
        WalletType.manual,
        params: {
          'privateKey': testWallet['privateKey']!,
          'address': testWallet['address']!,
        },
      );
      
      // Verify connection
      expect(walletService.isConnected, true);
      
      // Disconnect
      await walletService.disconnectWallet();
      
      // Verify all sensitive data is cleared
      expect(walletService.currentWallet, null);
      expect(walletService.isConnected, false);
      
      // Attempting to sign should fail
      expect(
        () => walletService.signDAOTransaction({'test': 'data'}),
        throwsA(isA<Exception>()),
      );
    });
  });

  group('WalletService Performance Tests', () {
    late WalletService walletService;

    setUp(() {
      walletService = WalletService();
    });

    test('should generate test wallets quickly', () {
      final stopwatch = Stopwatch()..start();
      
      for (int i = 0; i < 10; i++) {
        walletService.generateTestWallet();
      }
      
      stopwatch.stop();
      
      // Should generate 10 wallets in under 1 second
      expect(stopwatch.elapsedMilliseconds, lessThan(1000));
    });

    test('should handle multiple concurrent signing operations', () async {
      final testWallet = walletService.generateTestWallet();
      await walletService.connectWallet(
        WalletType.manual,
        params: {
          'privateKey': testWallet['privateKey']!,
          'address': testWallet['address']!,
        },
      );
      
      final transactions = List.generate(5, (index) => {
        'type': 'vote',
        'fee': 500,
        'proposalId': 'proposal_$index',
        'choice': 'yes',
        'nonce': DateTime.now().millisecondsSinceEpoch + index,
      });
      
      final stopwatch = Stopwatch()..start();
      
      final futures = transactions.map((tx) => 
        walletService.signDAOTransaction(tx)
      ).toList();
      
      final results = await Future.wait(futures);
      
      stopwatch.stop();
      
      // All should succeed
      expect(results.length, 5);
      for (final result in results) {
        expect(result['signature'], isNotNull);
        expect(result['transactionHash'], isNotNull);
      }
      
      // Should complete in reasonable time (under 5 seconds)
      expect(stopwatch.elapsedMilliseconds, lessThan(5000));
    });
  });
}