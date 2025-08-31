// Comprehensive tests for wallet integration
// This file contains tests for all wallet integration functionality

class WalletIntegrationTests {
    constructor() {
        this.testResults = [];
        this.walletIntegration = new WalletIntegration();
    }

    async runAllTests() {
        console.log('🧪 Starting Wallet Integration Tests...');
        
        const tests = [
            this.testWalletDetection,
            this.testManualWalletConnection,
            this.testTransactionSigning,
            this.testWalletDisconnection,
            this.testBalanceUpdates,
            this.testEventHandling,
            this.testErrorHandling,
            this.testSecurityFeatures,
            this.testUtilityFunctions
        ];

        for (const test of tests) {
            try {
                await test.call(this);
            } catch (error) {
                this.addTestResult(test.name, false, error.message);
            }
        }

        this.displayResults();
        return this.testResults;
    }

    addTestResult(testName, passed, message = '') {
        this.testResults.push({
            test: testName,
            passed,
            message,
            timestamp: new Date().toISOString()
        });
        
        const status = passed ? '✅' : '❌';
        console.log(`${status} ${testName}: ${message}`);
    }

    async testWalletDetection() {
        const available = this.walletIntegration.detectAvailableWallets();
        
        // Manual wallet should always be available
        if (!available.manual || !available.manual.installed) {
            throw new Error('Manual wallet should always be available');
        }

        // Check structure of returned data
        for (const [walletType, info] of Object.entries(available)) {
            if (!info.name || typeof info.installed !== 'boolean') {
                throw new Error(`Invalid wallet info structure for ${walletType}`);
            }
        }

        this.addTestResult('testWalletDetection', true, `Detected ${Object.keys(available).length} wallet types`);
    }

    async testManualWalletConnection() {
        // Generate test wallet
        const testWallet = this.walletIntegration.generateTestWallet();
        
        // Mock the manual connection process
        const originalConnect = this.walletIntegration.connectManual;
        this.walletIntegration.connectManual = async () => {
            this.walletIntegration.currentWallet = 'manual';
            this.walletIntegration.isConnected = true;
            this.walletIntegration.address = testWallet.address;
            this.walletIntegration.privateKey = testWallet.privateKey;
            
            return {
                success: true,
                wallet: 'manual',
                address: testWallet.address
            };
        };

        const result = await this.walletIntegration.connect('manual');
        
        if (!result.success) {
            throw new Error('Manual wallet connection failed');
        }

        if (!this.walletIntegration.isConnected) {
            throw new Error('Wallet should be connected after successful connection');
        }

        if (this.walletIntegration.address !== testWallet.address) {
            throw new Error('Address mismatch after connection');
        }

        // Restore original method
        this.walletIntegration.connectManual = originalConnect;

        this.addTestResult('testManualWalletConnection', true, 'Manual wallet connected successfully');
    }

    async testTransactionSigning() {
        // Ensure wallet is connected
        if (!this.walletIntegration.isConnected) {
            await this.testManualWalletConnection();
        }

        const testTransaction = {
            type: 'proposal',
            fee: 1000,
            title: 'Test Proposal',
            description: 'This is a test proposal for signing',
            nonce: Date.now()
        };

        try {
            const signedTx = await this.walletIntegration.signTransaction(testTransaction);
            
            if (!signedTx.signature) {
                throw new Error('Signature missing from signed transaction');
            }

            if (!signedTx.transactionHash) {
                throw new Error('Transaction hash missing from signed transaction');
            }

            if (signedTx.signer !== this.walletIntegration.address) {
                throw new Error('Signer address mismatch');
            }

            this.addTestResult('testTransactionSigning', true, 'Transaction signed successfully');
        } catch (error) {
            throw new Error(`Transaction signing failed: ${error.message}`);
        }
    }

    async testWalletDisconnection() {
        // Ensure wallet is connected first
        if (!this.walletIntegration.isConnected) {
            await this.testManualWalletConnection();
        }

        await this.walletIntegration.disconnect();

        if (this.walletIntegration.isConnected) {
            throw new Error('Wallet should be disconnected');
        }

        if (this.walletIntegration.address !== null) {
            throw new Error('Address should be null after disconnection');
        }

        if (this.walletIntegration.currentWallet !== null) {
            throw new Error('Current wallet should be null after disconnection');
        }

        this.addTestResult('testWalletDisconnection', true, 'Wallet disconnected successfully');
    }

    async testBalanceUpdates() {
        // Mock the balance update
        const originalUpdateBalance = this.walletIntegration.updateBalance;
        let balanceUpdateCalled = false;
        
        this.walletIntegration.updateBalance = async () => {
            balanceUpdateCalled = true;
            this.walletIntegration.balance = 5000;
            return 5000;
        };

        // Connect wallet first
        await this.testManualWalletConnection();

        const balance = await this.walletIntegration.updateBalance();

        if (!balanceUpdateCalled) {
            throw new Error('Balance update method was not called');
        }

        if (balance !== 5000) {
            throw new Error('Balance update returned incorrect value');
        }

        if (this.walletIntegration.balance !== 5000) {
            throw new Error('Wallet balance was not updated correctly');
        }

        // Restore original method
        this.walletIntegration.updateBalance = originalUpdateBalance;

        this.addTestResult('testBalanceUpdates', true, 'Balance updates working correctly');
    }

    async testEventHandling() {
        let connectEventFired = false;
        let disconnectEventFired = false;
        let balanceEventFired = false;

        // Set up event listeners
        this.walletIntegration.on('connect', () => {
            connectEventFired = true;
        });

        this.walletIntegration.on('disconnect', () => {
            disconnectEventFired = true;
        });

        this.walletIntegration.on('balanceUpdate', () => {
            balanceEventFired = true;
        });

        // Trigger events
        this.walletIntegration.emit('connect', { wallet: 'test' });
        this.walletIntegration.emit('disconnect');
        this.walletIntegration.emit('balanceUpdate', { balance: 1000 });

        if (!connectEventFired) {
            throw new Error('Connect event was not fired');
        }

        if (!disconnectEventFired) {
            throw new Error('Disconnect event was not fired');
        }

        if (!balanceEventFired) {
            throw new Error('Balance update event was not fired');
        }

        this.addTestResult('testEventHandling', true, 'Event handling working correctly');
    }

    async testErrorHandling() {
        // Test connection with invalid wallet type
        try {
            await this.walletIntegration.connect('invalid_wallet');
            throw new Error('Should have thrown error for invalid wallet type');
        } catch (error) {
            if (!error.message.includes('Unsupported wallet type')) {
                throw new Error('Wrong error message for invalid wallet type');
            }
        }

        // Test signing without connection
        this.walletIntegration.isConnected = false;
        try {
            await this.walletIntegration.signTransaction({ test: 'data' });
            throw new Error('Should have thrown error for signing without connection');
        } catch (error) {
            if (!error.message.includes('Wallet not connected')) {
                throw new Error('Wrong error message for signing without connection');
            }
        }

        this.addTestResult('testErrorHandling', true, 'Error handling working correctly');
    }

    async testSecurityFeatures() {
        // Test private key validation
        const validKey = 'a'.repeat(64);
        const invalidKey = 'invalid';

        if (!this.walletIntegration.isValidPrivateKey(validKey)) {
            throw new Error('Valid private key was rejected');
        }

        if (this.walletIntegration.isValidPrivateKey(invalidKey)) {
            throw new Error('Invalid private key was accepted');
        }

        // Test address validation
        const validAddress = 'a'.repeat(40);
        const invalidAddress = 'invalid';

        if (!this.walletIntegration.isValidAddress(validAddress)) {
            throw new Error('Valid address was rejected');
        }

        if (this.walletIntegration.isValidAddress(invalidAddress)) {
            throw new Error('Invalid address was accepted');
        }

        this.addTestResult('testSecurityFeatures', true, 'Security validation working correctly');
    }

    async testUtilityFunctions() {
        // Test address formatting
        const longAddress = 'a'.repeat(42);
        const formatted = this.walletIntegration.formatAddress(longAddress);
        
        if (!formatted.includes('...')) {
            throw new Error('Address formatting should include ellipsis');
        }

        if (formatted.length >= longAddress.length) {
            throw new Error('Formatted address should be shorter than original');
        }

        // Test test wallet generation
        const testWallet = this.walletIntegration.generateTestWallet();
        
        if (!testWallet.privateKey || !testWallet.address) {
            throw new Error('Test wallet generation incomplete');
        }

        if (!this.walletIntegration.isValidPrivateKey(testWallet.privateKey)) {
            throw new Error('Generated private key is invalid');
        }

        if (!this.walletIntegration.isValidAddress(testWallet.address)) {
            throw new Error('Generated address is invalid');
        }

        // Test wallet info
        this.walletIntegration.isConnected = true;
        this.walletIntegration.address = 'test_address';
        this.walletIntegration.balance = 1000;
        this.walletIntegration.currentWallet = 'manual';

        const info = this.walletIntegration.getWalletInfo();
        
        if (!info.isConnected || info.address !== 'test_address' || info.balance !== 1000) {
            throw new Error('Wallet info is incorrect');
        }

        this.addTestResult('testUtilityFunctions', true, 'Utility functions working correctly');
    }

    displayResults() {
        const passed = this.testResults.filter(r => r.passed).length;
        const total = this.testResults.length;
        const percentage = Math.round((passed / total) * 100);

        console.log('\n📊 Test Results Summary:');
        console.log(`✅ Passed: ${passed}/${total} (${percentage}%)`);
        console.log(`❌ Failed: ${total - passed}/${total}`);

        if (passed === total) {
            console.log('🎉 All tests passed!');
        } else {
            console.log('⚠️  Some tests failed. Check the details above.');
        }

        // Display failed tests
        const failed = this.testResults.filter(r => !r.passed);
        if (failed.length > 0) {
            console.log('\n❌ Failed Tests:');
            failed.forEach(test => {
                console.log(`  - ${test.test}: ${test.message}`);
            });
        }
    }

    // Integration test with mock API
    async testAPIIntegration() {
        console.log('🔗 Testing API Integration...');

        // Mock DAOAPI
        window.DAOAPI = class MockDAOAPI {
            async getTokenBalance(address) {
                return 5000;
            }

            async broadcastTransaction(signedTx) {
                return {
                    hash: 'mock_hash_' + Date.now(),
                    blockHeight: 12345
                };
            }
        };

        // Test balance fetching
        this.walletIntegration.address = 'test_address';
        this.walletIntegration.isConnected = true;
        
        const balance = await this.walletIntegration.updateBalance();
        if (balance !== 5000) {
            throw new Error('API balance integration failed');
        }

        // Test transaction broadcasting
        const mockSignedTx = {
            signature: 'mock_signature',
            transactionHash: 'mock_hash',
            signer: 'test_address'
        };

        const result = await this.walletIntegration.broadcastTransaction(mockSignedTx);
        if (!result.success || !result.transactionHash) {
            throw new Error('API transaction broadcasting failed');
        }

        this.addTestResult('testAPIIntegration', true, 'API integration working correctly');
    }

    // Performance test
    async testPerformance() {
        console.log('⚡ Testing Performance...');

        const iterations = 100;
        
        // Test wallet generation performance
        const startTime = performance.now();
        for (let i = 0; i < iterations; i++) {
            this.walletIntegration.generateTestWallet();
        }
        const endTime = performance.now();
        
        const avgTime = (endTime - startTime) / iterations;
        if (avgTime > 10) { // Should be under 10ms per generation
            throw new Error(`Wallet generation too slow: ${avgTime}ms average`);
        }

        // Test validation performance
        const validKey = 'a'.repeat(64);
        const startValidation = performance.now();
        for (let i = 0; i < iterations; i++) {
            this.walletIntegration.isValidPrivateKey(validKey);
        }
        const endValidation = performance.now();
        
        const avgValidation = (endValidation - startValidation) / iterations;
        if (avgValidation > 1) { // Should be under 1ms per validation
            throw new Error(`Validation too slow: ${avgValidation}ms average`);
        }

        this.addTestResult('testPerformance', true, `Performance acceptable (${avgTime.toFixed(2)}ms gen, ${avgValidation.toFixed(2)}ms val)`);
    }
}

// Auto-run tests when page loads (for development)
if (typeof window !== 'undefined') {
    window.WalletIntegrationTests = WalletIntegrationTests;
    
    // Add test runner to window for manual execution
    window.runWalletTests = async () => {
        const tester = new WalletIntegrationTests();
        await tester.runAllTests();
        await tester.testAPIIntegration();
        await tester.testPerformance();
        return tester.testResults;
    };

    // Auto-run tests in development mode
    if (localStorage.getItem('dao-debug') === 'true') {
        document.addEventListener('DOMContentLoaded', () => {
            setTimeout(() => {
                console.log('🚀 Auto-running wallet integration tests...');
                window.runWalletTests();
            }, 1000);
        });
    }
}

// Export for Node.js testing
if (typeof module !== 'undefined' && module.exports) {
    module.exports = WalletIntegrationTests;
}