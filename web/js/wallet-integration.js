// Enhanced Wallet Integration for ProjectX DAO
// Supports MetaMask, WalletConnect, and manual key management

class WalletIntegration {
    constructor() {
        this.supportedWallets = {
            metamask: 'MetaMask',
            walletconnect: 'WalletConnect',
            manual: 'Manual Key Input',
            ledger: 'Ledger Hardware Wallet'
        };
        
        this.currentWallet = null;
        this.isConnected = false;
        this.address = null;
        this.publicKey = null;
        this.chainId = null;
        this.balance = 0;
        
        // Event listeners
        this.listeners = {
            connect: [],
            disconnect: [],
            accountChange: [],
            chainChange: [],
            balanceUpdate: []
        };
        
        this.init();
    }

    async init() {
        // Check for existing connections
        await this.checkExistingConnections();
        
        // Set up event listeners for supported wallets
        this.setupEventListeners();
    }

    // Wallet Detection
    detectAvailableWallets() {
        const available = {};
        
        // MetaMask detection
        if (typeof window.ethereum !== 'undefined' && window.ethereum.isMetaMask) {
            available.metamask = {
                name: 'MetaMask',
                icon: 'https://metamask.io/images/metamask-logo.png',
                installed: true
            };
        }
        
        // WalletConnect detection
        if (typeof window.WalletConnect !== 'undefined') {
            available.walletconnect = {
                name: 'WalletConnect',
                icon: 'https://walletconnect.org/walletconnect-logo.svg',
                installed: true
            };
        }
        
        // Manual is always available
        available.manual = {
            name: 'Manual Key Input',
            icon: '🔑',
            installed: true
        };
        
        // Ledger detection (simplified)
        available.ledger = {
            name: 'Ledger Hardware Wallet',
            icon: 'https://www.ledger.com/favicon.ico',
            installed: typeof window.TransportWebUSB !== 'undefined'
        };
        
        return available;
    }

    // Connection Methods
    async connect(walletType) {
        try {
            switch (walletType) {
                case 'metamask':
                    return await this.connectMetaMask();
                case 'walletconnect':
                    return await this.connectWalletConnect();
                case 'manual':
                    return await this.connectManual();
                case 'ledger':
                    return await this.connectLedger();
                default:
                    throw new Error(`Unsupported wallet type: ${walletType}`);
            }
        } catch (error) {
            console.error('Wallet connection error:', error);
            throw error;
        }
    }

    async connectMetaMask() {
        if (typeof window.ethereum === 'undefined' || !window.ethereum.isMetaMask) {
            throw new Error('MetaMask is not installed. Please install MetaMask to continue.');
        }

        try {
            // Request account access
            const accounts = await window.ethereum.request({
                method: 'eth_requestAccounts'
            });

            if (accounts.length === 0) {
                throw new Error('No accounts found in MetaMask');
            }

            // Get chain ID
            const chainId = await window.ethereum.request({
                method: 'eth_chainId'
            });

            // Set up connection
            this.currentWallet = 'metamask';
            this.isConnected = true;
            this.address = accounts[0];
            this.chainId = chainId;

            // Get balance
            await this.updateBalance();

            // Store connection info
            localStorage.setItem('dao_wallet_type', 'metamask');
            localStorage.setItem('dao_wallet_address', this.address);

            this.emit('connect', {
                wallet: 'metamask',
                address: this.address,
                chainId: this.chainId
            });

            return {
                success: true,
                wallet: 'metamask',
                address: this.address,
                chainId: this.chainId
            };

        } catch (error) {
            throw new Error(`MetaMask connection failed: ${error.message}`);
        }
    }

    async connectWalletConnect() {
        try {
            // Initialize WalletConnect
            const WalletConnect = (await import('@walletconnect/client')).default;
            
            const connector = new WalletConnect({
                bridge: 'https://bridge.walletconnect.org',
                qrcodeModal: {
                    open: (uri, cb) => {
                        this.showQRCodeModal(uri, cb);
                    },
                    close: () => {
                        this.hideQRCodeModal();
                    }
                }
            });

            // Check if already connected
            if (!connector.connected) {
                await connector.createSession();
            }

            const { accounts, chainId } = connector;

            this.currentWallet = 'walletconnect';
            this.isConnected = true;
            this.address = accounts[0];
            this.chainId = chainId;
            this.walletConnectConnector = connector;

            // Set up event listeners
            connector.on('session_update', (error, payload) => {
                if (error) {
                    throw error;
                }
                const { accounts, chainId } = payload.params[0];
                this.address = accounts[0];
                this.chainId = chainId;
                this.emit('accountChange', { address: this.address, chainId: this.chainId });
            });

            connector.on('disconnect', (error, payload) => {
                if (error) {
                    throw error;
                }
                this.disconnect();
            });

            await this.updateBalance();

            localStorage.setItem('dao_wallet_type', 'walletconnect');
            localStorage.setItem('dao_wallet_address', this.address);

            this.emit('connect', {
                wallet: 'walletconnect',
                address: this.address,
                chainId: this.chainId
            });

            return {
                success: true,
                wallet: 'walletconnect',
                address: this.address,
                chainId: this.chainId
            };

        } catch (error) {
            throw new Error(`WalletConnect connection failed: ${error.message}`);
        }
    }

    async connectManual() {
        return new Promise((resolve, reject) => {
            const modal = this.createManualWalletModal();
            document.body.appendChild(modal);

            const form = modal.querySelector('#manualWalletForm');
            const closeBtn = modal.querySelector('.close');
            const cancelBtn = modal.querySelector('#cancelManualWallet');
            const generateBtn = modal.querySelector('#generateTestWallet');

            const cleanup = () => {
                document.body.removeChild(modal);
            };

            // Generate test wallet
            generateBtn.addEventListener('click', () => {
                const testWallet = this.generateTestWallet();
                document.getElementById('manualPrivateKey').value = testWallet.privateKey;
                document.getElementById('manualAddress').value = testWallet.address;
            });

            form.addEventListener('submit', async (e) => {
                e.preventDefault();
                
                const privateKeyInput = document.getElementById('manualPrivateKey').value.trim();
                const addressInput = document.getElementById('manualAddress').value.trim();

                try {
                    if (!privateKeyInput || !addressInput) {
                        throw new Error('Both private key and address are required');
                    }

                    if (!this.isValidPrivateKey(privateKeyInput)) {
                        throw new Error('Invalid private key format (must be 64 hex characters)');
                    }

                    if (!this.isValidAddress(addressInput)) {
                        throw new Error('Invalid address format');
                    }

                    // Derive public key from private key
                    const publicKey = await this.derivePublicKey(privateKeyInput);

                    this.currentWallet = 'manual';
                    this.isConnected = true;
                    this.address = addressInput;
                    this.publicKey = publicKey;
                    this.privateKey = privateKeyInput;

                    // Store securely (for development only)
                    sessionStorage.setItem('dao_wallet_private_key', privateKeyInput);
                    localStorage.setItem('dao_wallet_type', 'manual');
                    localStorage.setItem('dao_wallet_address', this.address);

                    await this.updateBalance();

                    cleanup();

                    this.emit('connect', {
                        wallet: 'manual',
                        address: this.address,
                        publicKey: this.publicKey
                    });

                    resolve({
                        success: true,
                        wallet: 'manual',
                        address: this.address,
                        publicKey: this.publicKey
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
        });
    }

    async connectLedger() {
        try {
            // Import Ledger libraries
            const TransportWebUSB = (await import('@ledgerhq/hw-transport-webusb')).default;
            const AppEth = (await import('@ledgerhq/hw-app-eth')).default;

            const transport = await TransportWebUSB.create();
            const eth = new AppEth(transport);

            // Get address from Ledger
            const result = await eth.getAddress("44'/60'/0'/0/0");
            
            this.currentWallet = 'ledger';
            this.isConnected = true;
            this.address = result.address;
            this.publicKey = result.publicKey;
            this.ledgerTransport = transport;
            this.ledgerApp = eth;

            await this.updateBalance();

            localStorage.setItem('dao_wallet_type', 'ledger');
            localStorage.setItem('dao_wallet_address', this.address);

            this.emit('connect', {
                wallet: 'ledger',
                address: this.address,
                publicKey: this.publicKey
            });

            return {
                success: true,
                wallet: 'ledger',
                address: this.address,
                publicKey: this.publicKey
            };

        } catch (error) {
            throw new Error(`Ledger connection failed: ${error.message}`);
        }
    }

    // Transaction Signing
    async signTransaction(transactionData) {
        if (!this.isConnected) {
            throw new Error('Wallet not connected');
        }

        switch (this.currentWallet) {
            case 'metamask':
                return await this.signWithMetaMask(transactionData);
            case 'walletconnect':
                return await this.signWithWalletConnect(transactionData);
            case 'manual':
                return await this.signWithPrivateKey(transactionData);
            case 'ledger':
                return await this.signWithLedger(transactionData);
            default:
                throw new Error('Unknown wallet type');
        }
    }

    async signWithMetaMask(transactionData) {
        try {
            // For DAO transactions, we need to format them properly
            const formattedTx = this.formatTransactionForSigning(transactionData);
            
            // Use eth_signTypedData_v4 for structured data
            const signature = await window.ethereum.request({
                method: 'eth_signTypedData_v4',
                params: [this.address, JSON.stringify(formattedTx)]
            });

            return {
                signature,
                transactionHash: this.calculateTransactionHash(transactionData),
                signer: this.address,
                method: 'metamask'
            };

        } catch (error) {
            throw new Error(`MetaMask signing failed: ${error.message}`);
        }
    }

    async signWithWalletConnect(transactionData) {
        try {
            const formattedTx = this.formatTransactionForSigning(transactionData);
            
            const signature = await this.walletConnectConnector.signTypedData([
                this.address,
                JSON.stringify(formattedTx)
            ]);

            return {
                signature,
                transactionHash: this.calculateTransactionHash(transactionData),
                signer: this.address,
                method: 'walletconnect'
            };

        } catch (error) {
            throw new Error(`WalletConnect signing failed: ${error.message}`);
        }
    }

    async signWithPrivateKey(transactionData) {
        try {
            // Use the ProjectX crypto library approach
            const transactionHash = this.calculateTransactionHash(transactionData);
            const signature = await this.signHashWithPrivateKey(transactionHash, this.privateKey);

            return {
                signature,
                transactionHash,
                signer: this.address,
                method: 'manual'
            };

        } catch (error) {
            throw new Error(`Private key signing failed: ${error.message}`);
        }
    }

    async signWithLedger(transactionData) {
        try {
            const transactionHash = this.calculateTransactionHash(transactionData);
            
            // Sign with Ledger
            const signature = await this.ledgerApp.signPersonalMessage(
                "44'/60'/0'/0/0",
                Buffer.from(transactionHash, 'hex')
            );

            return {
                signature: signature.r + signature.s + signature.v.toString(16),
                transactionHash,
                signer: this.address,
                method: 'ledger'
            };

        } catch (error) {
            throw new Error(`Ledger signing failed: ${error.message}`);
        }
    }

    // Transaction Broadcasting
    async broadcastTransaction(signedTransaction) {
        try {
            // Use the DAO API to broadcast the transaction
            const daoAPI = new DAOAPI();
            const result = await daoAPI.broadcastTransaction(signedTransaction);
            
            return {
                success: true,
                transactionHash: result.hash,
                blockHeight: result.blockHeight
            };

        } catch (error) {
            throw new Error(`Transaction broadcast failed: ${error.message}`);
        }
    }

    // Utility Methods
    formatTransactionForSigning(transactionData) {
        // Format transaction data according to EIP-712 standard
        return {
            types: {
                EIP712Domain: [
                    { name: 'name', type: 'string' },
                    { name: 'version', type: 'string' },
                    { name: 'chainId', type: 'uint256' }
                ],
                Transaction: [
                    { name: 'to', type: 'address' },
                    { name: 'value', type: 'uint256' },
                    { name: 'data', type: 'bytes' },
                    { name: 'nonce', type: 'uint256' }
                ]
            },
            primaryType: 'Transaction',
            domain: {
                name: 'ProjectX DAO',
                version: '1',
                chainId: parseInt(this.chainId || '1', 16)
            },
            message: {
                to: transactionData.to || '0x0000000000000000000000000000000000000000',
                value: transactionData.value || '0',
                data: transactionData.data || '0x',
                nonce: transactionData.nonce || Date.now()
            }
        };
    }

    calculateTransactionHash(transactionData) {
        // Simple hash calculation (in production, use proper cryptographic hashing)
        const dataString = JSON.stringify(transactionData);
        return this.sha256(dataString);
    }

    async signHashWithPrivateKey(hash, privateKey) {
        // Simplified signing (in production, use proper ECDSA signing)
        const crypto = await import('crypto');
        const sign = crypto.createSign('SHA256');
        sign.update(hash);
        return sign.sign(privateKey, 'hex');
    }

    sha256(data) {
        // Simple SHA256 implementation for browser
        const encoder = new TextEncoder();
        const dataBuffer = encoder.encode(data);
        return crypto.subtle.digest('SHA-256', dataBuffer).then(hashBuffer => {
            const hashArray = Array.from(new Uint8Array(hashBuffer));
            return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
        });
    }

    async derivePublicKey(privateKey) {
        // Simplified public key derivation
        // In production, use proper elliptic curve cryptography
        return privateKey.slice(0, 64); // Placeholder
    }

    generateTestWallet() {
        const privateKey = Array.from(crypto.getRandomValues(new Uint8Array(32)))
            .map(b => b.toString(16).padStart(2, '0'))
            .join('');
        
        const address = Array.from(crypto.getRandomValues(new Uint8Array(20)))
            .map(b => b.toString(16).padStart(2, '0'))
            .join('');

        return { privateKey, address };
    }

    isValidPrivateKey(key) {
        return /^[0-9a-fA-F]{64}$/.test(key);
    }

    isValidAddress(address) {
        return /^[0-9a-fA-F]{40}$/.test(address);
    }

    // Balance Management
    async updateBalance() {
        if (!this.isConnected) return;

        try {
            const daoAPI = new DAOAPI();
            const balance = await daoAPI.getTokenBalance(this.address);
            const oldBalance = this.balance;
            this.balance = balance;

            if (oldBalance !== balance) {
                this.emit('balanceUpdate', { balance, oldBalance });
            }

            return balance;
        } catch (error) {
            console.error('Balance update failed:', error);
            return this.balance;
        }
    }

    // Connection Management
    async disconnect() {
        if (this.currentWallet === 'walletconnect' && this.walletConnectConnector) {
            await this.walletConnectConnector.killSession();
        }

        if (this.currentWallet === 'ledger' && this.ledgerTransport) {
            await this.ledgerTransport.close();
        }

        this.currentWallet = null;
        this.isConnected = false;
        this.address = null;
        this.publicKey = null;
        this.privateKey = null;
        this.balance = 0;

        // Clear storage
        localStorage.removeItem('dao_wallet_type');
        localStorage.removeItem('dao_wallet_address');
        sessionStorage.removeItem('dao_wallet_private_key');

        this.emit('disconnect');
    }

    async checkExistingConnections() {
        const walletType = localStorage.getItem('dao_wallet_type');
        const address = localStorage.getItem('dao_wallet_address');

        if (walletType && address) {
            try {
                if (walletType === 'manual') {
                    const privateKey = sessionStorage.getItem('dao_wallet_private_key');
                    if (privateKey) {
                        this.currentWallet = 'manual';
                        this.isConnected = true;
                        this.address = address;
                        this.privateKey = privateKey;
                        this.publicKey = await this.derivePublicKey(privateKey);
                        await this.updateBalance();
                    }
                } else if (walletType === 'metamask' && window.ethereum) {
                    const accounts = await window.ethereum.request({ method: 'eth_accounts' });
                    if (accounts.includes(address)) {
                        this.currentWallet = 'metamask';
                        this.isConnected = true;
                        this.address = address;
                        await this.updateBalance();
                    }
                }
            } catch (error) {
                console.error('Auto-connection failed:', error);
                this.disconnect();
            }
        }
    }

    setupEventListeners() {
        // MetaMask event listeners
        if (window.ethereum) {
            window.ethereum.on('accountsChanged', (accounts) => {
                if (this.currentWallet === 'metamask') {
                    if (accounts.length === 0) {
                        this.disconnect();
                    } else {
                        this.address = accounts[0];
                        this.emit('accountChange', { address: this.address });
                        this.updateBalance();
                    }
                }
            });

            window.ethereum.on('chainChanged', (chainId) => {
                if (this.currentWallet === 'metamask') {
                    this.chainId = chainId;
                    this.emit('chainChange', { chainId });
                }
            });

            window.ethereum.on('disconnect', () => {
                if (this.currentWallet === 'metamask') {
                    this.disconnect();
                }
            });
        }
    }

    // Event System
    on(event, callback) {
        if (this.listeners[event]) {
            this.listeners[event].push(callback);
        }
    }

    off(event, callback) {
        if (this.listeners[event]) {
            const index = this.listeners[event].indexOf(callback);
            if (index > -1) {
                this.listeners[event].splice(index, 1);
            }
        }
    }

    emit(event, data) {
        if (this.listeners[event]) {
            this.listeners[event].forEach(callback => callback(data));
        }
    }

    // UI Helpers
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
                            <input type="password" id="manualPrivateKey" placeholder="Enter your private key (64 hex characters)" required>
                            <small class="text-muted">Your private key will be stored temporarily in session storage</small>
                        </div>
                        <div class="form-group">
                            <label for="manualAddress">Address</label>
                            <input type="text" id="manualAddress" placeholder="Enter your address (40 hex characters)" required>
                            <small class="text-muted">Your public address for the DAO</small>
                        </div>
                        <div class="form-group">
                            <button type="button" id="generateTestWallet" class="btn btn-outline">Generate Test Wallet</button>
                            <small class="text-muted">Generate a test wallet for development purposes</small>
                        </div>
                        <div id="manualWalletError" class="text-danger" style="display: none; margin-bottom: 1rem;"></div>
                        <div class="form-actions">
                            <button type="button" id="cancelManualWallet" class="btn btn-secondary">Cancel</button>
                            <button type="submit" class="btn btn-primary">Connect</button>
                        </div>
                        <div style="margin-top: 1rem; padding: 1rem; background: #fff3cd; border-radius: 8px; font-size: 0.9rem;">
                            <strong>⚠️ Security Warning:</strong> This is a development interface. Never enter real private keys in production applications.
                        </div>
                    </div>
                </form>
            </div>
        `;
        return modal;
    }

    showQRCodeModal(uri, callback) {
        const modal = document.createElement('div');
        modal.className = 'modal active';
        modal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h3>Connect with WalletConnect</h3>
                    <span class="close">&times;</span>
                </div>
                <div style="padding: 1.5rem; text-align: center;">
                    <p>Scan this QR code with your wallet app:</p>
                    <div id="qrcode" style="margin: 1rem 0;"></div>
                    <p><small>Or copy the connection URI:</small></p>
                    <input type="text" value="${uri}" readonly style="width: 100%; margin-bottom: 1rem;">
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // Generate QR code (you'd need to include a QR code library)
        const qrDiv = modal.querySelector('#qrcode');
        qrDiv.innerHTML = `<p>QR Code would appear here<br><small>${uri}</small></p>`;
        
        modal.querySelector('.close').addEventListener('click', () => {
            document.body.removeChild(modal);
            callback();
        });
        
        this.qrModal = modal;
    }

    hideQRCodeModal() {
        if (this.qrModal) {
            document.body.removeChild(this.qrModal);
            this.qrModal = null;
        }
    }

    // Wallet Info
    getWalletInfo() {
        return {
            isConnected: this.isConnected,
            wallet: this.currentWallet,
            address: this.address,
            publicKey: this.publicKey,
            balance: this.balance,
            chainId: this.chainId
        };
    }

    formatAddress(address) {
        if (!address) return '';
        return `${address.slice(0, 6)}...${address.slice(-4)}`;
    }
}

// Export for use in other modules
window.WalletIntegration = WalletIntegration;