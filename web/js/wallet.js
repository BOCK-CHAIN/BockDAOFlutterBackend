// Wallet management for ProjectX DAO
class WalletManager {
    constructor() {
        this.isConnected = false;
        this.address = null;
        this.privateKey = null;
        this.publicKey = null;
        this.balance = 0;
        this.supportedWallets = ['metamask', 'manual'];
        this.currentWallet = null;
    }

    async connect(walletType = 'manual') {
        try {
            switch (walletType) {
                case 'metamask':
                    return await this.connectMetaMask();
                case 'manual':
                    return await this.connectManual();
                default:
                    throw new Error('Unsupported wallet type');
            }
        } catch (error) {
            console.error('Wallet connection error:', error);
            return { success: false, error: error.message };
        }
    }

    async connectMetaMask() {
        if (typeof window.ethereum === 'undefined') {
            throw new Error('MetaMask is not installed');
        }

        try {
            // Request account access
            const accounts = await window.ethereum.request({
                method: 'eth_requestAccounts'
            });

            if (accounts.length === 0) {
                throw new Error('No accounts found');
            }

            this.address = accounts[0];
            this.isConnected = true;
            this.currentWallet = 'metamask';

            // Listen for account changes
            window.ethereum.on('accountsChanged', (accounts) => {
                if (accounts.length === 0) {
                    this.disconnect();
                } else {
                    this.address = accounts[0];
                }
            });

            // Listen for chain changes
            window.ethereum.on('chainChanged', () => {
                window.location.reload();
            });

            return {
                success: true,
                address: this.address,
                wallet: 'metamask'
            };
        } catch (error) {
            throw new Error(`MetaMask connection failed: ${error.message}`);
        }
    }

    async connectManual() {
        return new Promise((resolve, reject) => {
            // Create modal for manual key input
            const modal = this.createManualWalletModal();
            document.body.appendChild(modal);

            const form = modal.querySelector('#manualWalletForm');
            const closeBtn = modal.querySelector('.close');
            const cancelBtn = modal.querySelector('#cancelManualWallet');

            const cleanup = () => {
                document.body.removeChild(modal);
            };

            form.addEventListener('submit', async (e) => {
                e.preventDefault();
                
                const privateKeyInput = document.getElementById('manualPrivateKey').value;
                const addressInput = document.getElementById('manualAddress').value;

                try {
                    // Validate inputs
                    if (!privateKeyInput || !addressInput) {
                        throw new Error('Both private key and address are required');
                    }

                    if (!this.isValidPrivateKey(privateKeyInput)) {
                        throw new Error('Invalid private key format');
                    }

                    if (!this.isValidAddress(addressInput)) {
                        throw new Error('Invalid address format');
                    }

                    // Store wallet info
                    this.privateKey = privateKeyInput;
                    this.address = addressInput;
                    this.isConnected = true;
                    this.currentWallet = 'manual';

                    // Store in session storage (not recommended for production)
                    sessionStorage.setItem('dao_wallet_private_key', privateKeyInput);
                    sessionStorage.setItem('dao_wallet_address', addressInput);

                    cleanup();
                    resolve({
                        success: true,
                        address: this.address,
                        wallet: 'manual'
                    });
                } catch (error) {
                    const errorDiv = modal.querySelector('#manualWalletError');
                    errorDiv.textContent = error.message;
                    errorDiv.style.display = 'block';
                }
            });

            closeBtn.addEventListener('click', () => {
                cleanup();
                reject(new Error('User cancelled wallet connection'));
            });

            cancelBtn.addEventListener('click', () => {
                cleanup();
                reject(new Error('User cancelled wallet connection'));
            });

            // Check if we have stored credentials
            const storedPrivateKey = sessionStorage.getItem('dao_wallet_private_key');
            const storedAddress = sessionStorage.getItem('dao_wallet_address');

            if (storedPrivateKey && storedAddress) {
                document.getElementById('manualPrivateKey').value = storedPrivateKey;
                document.getElementById('manualAddress').value = storedAddress;
            }
        });
    }

    createManualWalletModal() {
        const modal = document.createElement('div');
        modal.className = 'modal active';
        modal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h3>Connect Wallet Manually</h3>
                    <span class="close">&times;</span>
                </div>
                <form id="manualWalletForm">
                    <div style="padding: 1.5rem;">
                        <div class="form-group">
                            <label for="manualPrivateKey">Private Key</label>
                            <input type="password" id="manualPrivateKey" placeholder="Enter your private key (hex format)" required>
                            <small class="text-muted">Your private key will be stored temporarily in session storage</small>
                        </div>
                        <div class="form-group">
                            <label for="manualAddress">Address</label>
                            <input type="text" id="manualAddress" placeholder="Enter your address (hex format)" required>
                            <small class="text-muted">Your public address for the DAO</small>
                        </div>
                        <div id="manualWalletError" class="text-danger" style="display: none; margin-bottom: 1rem;"></div>
                        <div class="form-actions">
                            <button type="button" id="cancelManualWallet" class="btn btn-secondary">Cancel</button>
                            <button type="submit" class="btn btn-primary">Connect</button>
                        </div>
                        <div style="margin-top: 1rem; padding: 1rem; background: #fff3cd; border-radius: 8px; font-size: 0.9rem;">
                            <strong>⚠️ Security Warning:</strong> This is a development interface. Never enter real private keys in production applications.
                            For testing, you can generate a test key pair or use the demo credentials.
                        </div>
                    </div>
                </form>
            </div>
        `;
        return modal;
    }

    async disconnect() {
        this.isConnected = false;
        this.address = null;
        this.privateKey = null;
        this.publicKey = null;
        this.balance = 0;
        this.currentWallet = null;

        // Clear session storage
        sessionStorage.removeItem('dao_wallet_private_key');
        sessionStorage.removeItem('dao_wallet_address');

        // If MetaMask, we don't need to do anything special
        // The user can disconnect from MetaMask directly
    }

    async getPrivateKey() {
        if (!this.isConnected) {
            throw new Error('Wallet not connected');
        }

        if (this.currentWallet === 'manual') {
            return this.privateKey;
        } else if (this.currentWallet === 'metamask') {
            // For MetaMask, we can't access the private key directly
            // In a real implementation, you'd use MetaMask's signing methods
            throw new Error('Cannot access private key from MetaMask. Use signing methods instead.');
        }

        throw new Error('Unknown wallet type');
    }

    async signTransaction(transactionData) {
        if (!this.isConnected) {
            throw new Error('Wallet not connected');
        }

        if (this.currentWallet === 'manual') {
            // For manual wallet, we'd implement signing here
            // This is a simplified version - in production you'd use proper crypto libraries
            return this.signWithPrivateKey(transactionData, this.privateKey);
        } else if (this.currentWallet === 'metamask') {
            // For MetaMask, use their signing methods
            return await this.signWithMetaMask(transactionData);
        }

        throw new Error('Unknown wallet type');
    }

    async signWithPrivateKey(data, privateKey) {
        // This is a placeholder implementation
        // In a real application, you'd use proper cryptographic signing
        const message = JSON.stringify(data);
        const signature = `signed_${message}_with_${privateKey.slice(0, 8)}`;
        
        return {
            signature,
            message,
            signer: this.address
        };
    }

    async signWithMetaMask(data) {
        try {
            const message = JSON.stringify(data);
            const signature = await window.ethereum.request({
                method: 'personal_sign',
                params: [message, this.address]
            });

            return {
                signature,
                message,
                signer: this.address
            };
        } catch (error) {
            throw new Error(`MetaMask signing failed: ${error.message}`);
        }
    }

    async getBalance() {
        if (!this.isConnected) {
            return 0;
        }

        try {
            // In a real implementation, you'd query the blockchain for the balance
            // For now, we'll use the DAO API
            const daoAPI = new DAOAPI();
            this.balance = await daoAPI.getTokenBalance(this.address);
            return this.balance;
        } catch (error) {
            console.error('Error fetching balance:', error);
            return 0;
        }
    }

    isValidPrivateKey(key) {
        // Basic validation for hex private key (64 characters)
        const hexRegex = /^[0-9a-fA-F]{64}$/;
        return hexRegex.test(key);
    }

    isValidAddress(address) {
        // Basic validation for hex address
        const hexRegex = /^[0-9a-fA-F]+$/;
        return hexRegex.test(address) && address.length >= 40;
    }

    // Utility methods
    formatAddress(address) {
        if (!address) return '';
        return `${address.slice(0, 6)}...${address.slice(-4)}`;
    }

    // Generate a test wallet for development
    generateTestWallet() {
        // This is for development/testing only
        const privateKey = this.generateRandomHex(64);
        const address = this.generateRandomHex(40);
        
        return {
            privateKey,
            address,
            isTest: true
        };
    }

    generateRandomHex(length) {
        const chars = '0123456789abcdef';
        let result = '';
        for (let i = 0; i < length; i++) {
            result += chars.charAt(Math.floor(Math.random() * chars.length));
        }
        return result;
    }

    // Auto-connect if previously connected
    async autoConnect() {
        const storedPrivateKey = sessionStorage.getItem('dao_wallet_private_key');
        const storedAddress = sessionStorage.getItem('dao_wallet_address');

        if (storedPrivateKey && storedAddress) {
            try {
                this.privateKey = storedPrivateKey;
                this.address = storedAddress;
                this.isConnected = true;
                this.currentWallet = 'manual';

                return {
                    success: true,
                    address: this.address,
                    wallet: 'manual'
                };
            } catch (error) {
                console.error('Auto-connect failed:', error);
                this.disconnect();
            }
        }

        return { success: false };
    }

    // Wallet info
    getWalletInfo() {
        return {
            isConnected: this.isConnected,
            address: this.address,
            balance: this.balance,
            walletType: this.currentWallet
        };
    }

    // Event handling
    onConnect(callback) {
        this.onConnectCallback = callback;
    }

    onDisconnect(callback) {
        this.onDisconnectCallback = callback;
    }

    onBalanceChange(callback) {
        this.onBalanceChangeCallback = callback;
    }

    // Trigger events
    triggerConnect() {
        if (this.onConnectCallback) {
            this.onConnectCallback(this.getWalletInfo());
        }
    }

    triggerDisconnect() {
        if (this.onDisconnectCallback) {
            this.onDisconnectCallback();
        }
    }

    triggerBalanceChange(newBalance) {
        const oldBalance = this.balance;
        this.balance = newBalance;
        
        if (this.onBalanceChangeCallback && oldBalance !== newBalance) {
            this.onBalanceChangeCallback(newBalance, oldBalance);
        }
    }

    // Periodic balance updates
    startBalanceUpdates(interval = 30000) {
        if (this.balanceUpdateInterval) {
            clearInterval(this.balanceUpdateInterval);
        }

        this.balanceUpdateInterval = setInterval(async () => {
            if (this.isConnected) {
                try {
                    const newBalance = await this.getBalance();
                    this.triggerBalanceChange(newBalance);
                } catch (error) {
                    console.error('Balance update failed:', error);
                }
            }
        }, interval);
    }

    stopBalanceUpdates() {
        if (this.balanceUpdateInterval) {
            clearInterval(this.balanceUpdateInterval);
            this.balanceUpdateInterval = null;
        }
    }
}