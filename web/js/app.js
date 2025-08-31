// Main application controller
class DAOApp {
    constructor() {
        this.currentSection = 'dashboard';
        this.isWalletConnected = false;
        this.walletAddress = null;
        this.tokenBalance = 0;
        this.daoAPI = new DAOAPI();
        this.websocket = new DAOWebSocket();
        this.wallet = new WalletManager();
        
        this.init();
    }

    async init() {
        this.setupEventListeners();
        this.setupWebSocket();
        await this.loadInitialData();
        this.startPeriodicUpdates();
    }

    setupEventListeners() {
        // Navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const section = e.target.dataset.section;
                this.navigateToSection(section);
            });
        });

        // Wallet connection
        document.getElementById('connectWallet').addEventListener('click', () => {
            this.connectWallet();
        });

        // Modal controls
        document.querySelectorAll('.close').forEach(closeBtn => {
            closeBtn.addEventListener('click', (e) => {
                const modal = e.target.closest('.modal');
                this.closeModal(modal.id);
            });
        });

        // Click outside modal to close
        document.querySelectorAll('.modal').forEach(modal => {
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    this.closeModal(modal.id);
                }
            });
        });

        // Create proposal button
        document.getElementById('createProposalBtn').addEventListener('click', () => {
            this.openModal('createProposalModal');
        });

        // Create treasury transaction button
        document.getElementById('createTreasuryTxBtn').addEventListener('click', () => {
            this.openModal('treasuryTxModal');
        });

        // Form submissions
        document.getElementById('createProposalForm').addEventListener('submit', (e) => {
            this.handleCreateProposal(e);
        });

        document.getElementById('treasuryTxForm').addEventListener('submit', (e) => {
            this.handleCreateTreasuryTransaction(e);
        });

        // Filters
        document.getElementById('statusFilter').addEventListener('change', () => {
            this.filterProposals();
        });

        document.getElementById('typeFilter').addEventListener('change', () => {
            this.filterProposals();
        });

        // Member search
        document.getElementById('memberSearch').addEventListener('input', (e) => {
            this.searchMembers(e.target.value);
        });
    }

    setupWebSocket() {
        this.websocket.onEvent = (event) => {
            this.handleWebSocketEvent(event);
        };

        this.websocket.onConnectionChange = (connected) => {
            this.updateConnectionStatus(connected);
        };

        this.websocket.connect();
    }

    async loadInitialData() {
        try {
            await Promise.all([
                this.loadDashboardData(),
                this.loadProposals(),
                this.loadTreasuryData(),
                this.loadMembers()
            ]);
        } catch (error) {
            console.error('Error loading initial data:', error);
            this.showNotification('Error loading data', 'error');
        }
    }

    startPeriodicUpdates() {
        // Update dashboard every 30 seconds
        setInterval(() => {
            if (this.currentSection === 'dashboard') {
                this.loadDashboardData();
            }
        }, 30000);

        // Update proposals every 60 seconds
        setInterval(() => {
            if (this.currentSection === 'proposals') {
                this.loadProposals();
            }
        }, 60000);

        // Update treasury every 60 seconds
        setInterval(() => {
            if (this.currentSection === 'treasury') {
                this.loadTreasuryData();
            }
        }, 60000);
    }

    navigateToSection(section) {
        // Update navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
        });
        document.querySelector(`[data-section="${section}"]`).classList.add('active');

        // Update content
        document.querySelectorAll('.content-section').forEach(sec => {
            sec.classList.remove('active');
        });
        document.getElementById(section).classList.add('active');

        this.currentSection = section;

        // Load section-specific data
        switch (section) {
            case 'dashboard':
                this.loadDashboardData();
                break;
            case 'proposals':
                this.loadProposals();
                break;
            case 'treasury':
                this.loadTreasuryData();
                break;
            case 'members':
                this.loadMembers();
                break;
        }
    }

    async connectWallet() {
        try {
            const result = await this.wallet.connect();
            if (result.success) {
                this.isWalletConnected = true;
                this.walletAddress = result.address;
                this.tokenBalance = await this.daoAPI.getTokenBalance(result.address);
                this.updateWalletUI();
                this.showNotification('Wallet connected successfully', 'success');
            } else {
                this.showNotification(result.error, 'error');
            }
        } catch (error) {
            console.error('Wallet connection error:', error);
            this.showNotification('Failed to connect wallet', 'error');
        }
    }

    updateWalletUI() {
        const connectBtn = document.getElementById('connectWallet');
        const walletInfo = document.getElementById('walletInfo');
        const walletAddress = document.getElementById('walletAddress');
        const tokenBalance = document.getElementById('tokenBalance');

        if (this.isWalletConnected) {
            connectBtn.style.display = 'none';
            walletInfo.style.display = 'block';
            walletAddress.textContent = this.formatAddress(this.walletAddress);
            tokenBalance.textContent = `${this.tokenBalance} PX`;
        } else {
            connectBtn.style.display = 'block';
            walletInfo.style.display = 'none';
        }
    }

    async loadDashboardData() {
        try {
            const [proposals, treasury, members] = await Promise.all([
                this.daoAPI.getProposals(),
                this.daoAPI.getTreasury(),
                this.daoAPI.getMembers()
            ]);

            // Update stats
            document.getElementById('totalProposals').textContent = proposals.length;
            document.getElementById('activeProposals').textContent = 
                proposals.filter(p => p.status === 2).length;
            document.getElementById('treasuryBalance').textContent = treasury.balance;
            document.getElementById('totalMembers').textContent = members.total || members.length;

            // Update recent activity
            this.updateRecentActivity(proposals.slice(0, 5));
        } catch (error) {
            console.error('Error loading dashboard data:', error);
        }
    }

    updateRecentActivity(recentProposals) {
        const activityList = document.getElementById('recentActivity');
        
        if (recentProposals.length === 0) {
            activityList.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-inbox"></i>
                    <h3>No Recent Activity</h3>
                    <p>Create your first proposal to get started!</p>
                </div>
            `;
            return;
        }

        activityList.innerHTML = recentProposals.map(proposal => `
            <div class="activity-item">
                <div class="activity-icon">
                    <i class="fas fa-file-alt"></i>
                </div>
                <div class="activity-content">
                    <div class="activity-title">${proposal.title}</div>
                    <div class="activity-time">${this.formatTime(proposal.start_time)}</div>
                </div>
            </div>
        `).join('');
    }

    async loadProposals() {
        try {
            const proposals = await this.daoAPI.getProposals();
            this.renderProposals(proposals);
        } catch (error) {
            console.error('Error loading proposals:', error);
            this.showNotification('Error loading proposals', 'error');
        }
    }

    renderProposals(proposals) {
        const proposalsList = document.getElementById('proposalsList');
        
        if (proposals.length === 0) {
            proposalsList.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-vote-yea"></i>
                    <h3>No Proposals Yet</h3>
                    <p>Be the first to create a governance proposal!</p>
                </div>
            `;
            return;
        }

        proposalsList.innerHTML = proposals.map(proposal => `
            <div class="proposal-card" data-id="${proposal.id}">
                <div class="proposal-header">
                    <div>
                        <div class="proposal-title">${proposal.title}</div>
                        <div class="proposal-meta">
                            <span>By ${this.formatAddress(proposal.creator)}</span>
                            <span>•</span>
                            <span>${this.formatTime(proposal.start_time)}</span>
                            <span>•</span>
                            <span>${this.getProposalTypeText(proposal.proposal_type)}</span>
                        </div>
                    </div>
                    <div class="proposal-status status-${this.getStatusClass(proposal.status)}">
                        ${this.getStatusText(proposal.status)}
                    </div>
                </div>
                <div class="proposal-description">
                    ${proposal.description}
                </div>
                ${this.renderVotingSection(proposal)}
            </div>
        `).join('');

        // Add event listeners for voting buttons
        this.setupVotingEventListeners();
    }

    renderVotingSection(proposal) {
        if (proposal.status !== 2) { // Not active
            return '';
        }

        const results = proposal.results || { yes_votes: 0, no_votes: 0, total_voters: 0 };
        const totalVotes = results.yes_votes + results.no_votes;
        const yesPercentage = totalVotes > 0 ? (results.yes_votes / totalVotes) * 100 : 0;

        return `
            <div class="proposal-voting">
                <div class="voting-progress">
                    <div class="progress-bar">
                        <div class="progress-fill" style="width: ${yesPercentage}%"></div>
                    </div>
                    <div class="voting-stats">
                        <span>Yes: ${results.yes_votes}</span>
                        <span>No: ${results.no_votes}</span>
                        <span>Total: ${results.total_voters}</span>
                    </div>
                </div>
                <div class="voting-actions">
                    <button class="btn btn-success vote-btn" data-proposal="${proposal.id}" data-choice="1">
                        <i class="fas fa-thumbs-up"></i> Yes
                    </button>
                    <button class="btn btn-danger vote-btn" data-proposal="${proposal.id}" data-choice="2">
                        <i class="fas fa-thumbs-down"></i> No
                    </button>
                    <button class="btn btn-secondary vote-btn" data-proposal="${proposal.id}" data-choice="3">
                        <i class="fas fa-minus"></i> Abstain
                    </button>
                </div>
            </div>
        `;
    }

    setupVotingEventListeners() {
        document.querySelectorAll('.vote-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const proposalId = e.target.dataset.proposal;
                const choice = parseInt(e.target.dataset.choice);
                this.openVoteModal(proposalId, choice);
            });
        });
    }

    openVoteModal(proposalId, choice) {
        if (!this.isWalletConnected) {
            this.showNotification('Please connect your wallet first', 'warning');
            return;
        }

        const modal = document.getElementById('voteModal');
        const content = document.getElementById('voteModalContent');
        
        const choiceText = ['', 'Yes', 'No', 'Abstain'][choice];
        const choiceClass = ['', 'success', 'danger', 'secondary'][choice];

        content.innerHTML = `
            <form id="voteForm">
                <div class="form-group">
                    <label>Your Vote</label>
                    <div class="btn btn-${choiceClass}" style="width: 100%; justify-content: center;">
                        ${choiceText}
                    </div>
                </div>
                <div class="form-group">
                    <label for="voteWeight">Vote Weight</label>
                    <input type="number" id="voteWeight" value="${this.tokenBalance}" max="${this.tokenBalance}" min="1" required>
                    <small class="text-muted">Available: ${this.tokenBalance} PX</small>
                </div>
                <div class="form-group">
                    <label for="voteReason">Reason (Optional)</label>
                    <textarea id="voteReason" rows="3" placeholder="Explain your vote..."></textarea>
                </div>
                <div class="form-actions">
                    <button type="button" class="btn btn-secondary" onclick="app.closeModal('voteModal')">Cancel</button>
                    <button type="submit" class="btn btn-primary">Cast Vote</button>
                </div>
            </form>
        `;

        document.getElementById('voteForm').addEventListener('submit', (e) => {
            this.handleVote(e, proposalId, choice);
        });

        this.openModal('voteModal');
    }

    async handleVote(e, proposalId, choice) {
        e.preventDefault();
        
        const weight = parseInt(document.getElementById('voteWeight').value);
        const reason = document.getElementById('voteReason').value;

        try {
            const result = await this.daoAPI.castVote({
                proposal_id: proposalId,
                choice: choice,
                weight: weight,
                reason: reason,
                private_key: await this.wallet.getPrivateKey()
            });

            this.closeModal('voteModal');
            this.showNotification('Vote cast successfully!', 'success');
            this.loadProposals(); // Refresh proposals
        } catch (error) {
            console.error('Error casting vote:', error);
            this.showNotification('Failed to cast vote', 'error');
        }
    }

    async handleCreateProposal(e) {
        e.preventDefault();

        if (!this.isWalletConnected) {
            this.showNotification('Please connect your wallet first', 'warning');
            return;
        }

        const formData = new FormData(e.target);
        const proposalData = {
            title: document.getElementById('proposalTitle').value,
            description: document.getElementById('proposalDescription').value,
            proposal_type: parseInt(document.getElementById('proposalType').value),
            voting_type: parseInt(document.getElementById('votingType').value),
            duration: parseInt(document.getElementById('proposalDuration').value) * 3600, // Convert to seconds
            threshold: parseInt(document.getElementById('proposalThreshold').value),
            metadata_hash: '',
            private_key: await this.wallet.getPrivateKey()
        };

        try {
            const result = await this.daoAPI.createProposal(proposalData);
            this.closeModal('createProposalModal');
            this.showNotification('Proposal created successfully!', 'success');
            this.loadProposals(); // Refresh proposals
            e.target.reset(); // Reset form
        } catch (error) {
            console.error('Error creating proposal:', error);
            this.showNotification('Failed to create proposal', 'error');
        }
    }

    async loadTreasuryData() {
        try {
            const [treasury, transactions] = await Promise.all([
                this.daoAPI.getTreasury(),
                this.daoAPI.getTreasuryTransactions()
            ]);

            // Update treasury overview
            document.getElementById('treasuryBalanceDetail').textContent = `${treasury.balance} PX`;
            document.getElementById('requiredSigs').textContent = treasury.required_sigs;
            document.getElementById('totalSigners').textContent = treasury.signers.length;

            // Update transactions
            this.renderTreasuryTransactions(transactions);
        } catch (error) {
            console.error('Error loading treasury data:', error);
            this.showNotification('Error loading treasury data', 'error');
        }
    }

    renderTreasuryTransactions(transactions) {
        const transactionsList = document.getElementById('treasuryTransactions');
        
        if (transactions.length === 0) {
            transactionsList.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-coins"></i>
                    <h3>No Transactions Yet</h3>
                    <p>Treasury transactions will appear here.</p>
                </div>
            `;
            return;
        }

        transactionsList.innerHTML = transactions.map(tx => `
            <div class="transaction-item">
                <div class="transaction-info">
                    <div class="transaction-amount">${tx.amount} PX</div>
                    <div class="transaction-purpose">${tx.purpose}</div>
                    <div class="text-muted">To: ${this.formatAddress(tx.recipient)}</div>
                </div>
                <div class="transaction-status tx-${tx.executed ? 'executed' : 'pending'}">
                    ${tx.executed ? 'Executed' : `${tx.signatures.length}/${this.requiredSigs} Signatures`}
                </div>
            </div>
        `).join('');
    }

    async handleCreateTreasuryTransaction(e) {
        e.preventDefault();

        if (!this.isWalletConnected) {
            this.showNotification('Please connect your wallet first', 'warning');
            return;
        }

        const txData = {
            recipient: document.getElementById('txRecipient').value,
            amount: parseInt(document.getElementById('txAmount').value),
            purpose: document.getElementById('txPurpose').value,
            private_key: await this.wallet.getPrivateKey()
        };

        try {
            const result = await this.daoAPI.createTreasuryTransaction(txData);
            this.closeModal('treasuryTxModal');
            this.showNotification('Treasury transaction created successfully!', 'success');
            this.loadTreasuryData(); // Refresh treasury data
            e.target.reset(); // Reset form
        } catch (error) {
            console.error('Error creating treasury transaction:', error);
            this.showNotification('Failed to create treasury transaction', 'error');
        }
    }

    async loadMembers() {
        try {
            const members = await this.daoAPI.getMembers();
            this.renderMembers(members.members || members);
        } catch (error) {
            console.error('Error loading members:', error);
            this.showNotification('Error loading members', 'error');
        }
    }

    renderMembers(members) {
        const membersList = document.getElementById('membersList');
        
        if (members.length === 0) {
            membersList.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-users"></i>
                    <h3>No Members Yet</h3>
                    <p>DAO members will appear here.</p>
                </div>
            `;
            return;
        }

        membersList.innerHTML = members.map(member => `
            <div class="member-card">
                <div class="member-header">
                    <div class="member-avatar">
                        ${this.getAddressInitials(member.address)}
                    </div>
                    <div class="member-info">
                        <h4>${this.formatAddress(member.address)}</h4>
                    </div>
                </div>
                <div class="member-stats">
                    <div class="member-stat">
                        <div class="member-stat-value">${member.balance}</div>
                        <div class="member-stat-label">Balance</div>
                    </div>
                    <div class="member-stat">
                        <div class="member-stat-value">${member.reputation}</div>
                        <div class="member-stat-label">Reputation</div>
                    </div>
                </div>
            </div>
        `).join('');
    }

    searchMembers(query) {
        const memberCards = document.querySelectorAll('.member-card');
        memberCards.forEach(card => {
            const address = card.querySelector('h4').textContent.toLowerCase();
            if (address.includes(query.toLowerCase())) {
                card.style.display = 'block';
            } else {
                card.style.display = 'none';
            }
        });
    }

    filterProposals() {
        const statusFilter = document.getElementById('statusFilter').value;
        const typeFilter = document.getElementById('typeFilter').value;
        const proposalCards = document.querySelectorAll('.proposal-card');

        proposalCards.forEach(card => {
            let show = true;
            
            if (statusFilter) {
                const status = card.querySelector('.proposal-status').className;
                if (!status.includes(`status-${this.getStatusClass(parseInt(statusFilter))}`)) {
                    show = false;
                }
            }
            
            if (typeFilter && show) {
                const typeText = card.querySelector('.proposal-meta').textContent;
                const expectedType = this.getProposalTypeText(parseInt(typeFilter));
                if (!typeText.includes(expectedType)) {
                    show = false;
                }
            }

            card.style.display = show ? 'block' : 'none';
        });
    }

    handleWebSocketEvent(event) {
        switch (event.type) {
            case 'proposal_created':
                this.showNotification(`New proposal: ${event.data.title}`, 'info');
                if (this.currentSection === 'proposals' || this.currentSection === 'dashboard') {
                    this.loadProposals();
                    this.loadDashboardData();
                }
                break;
            case 'vote_cast':
                if (this.currentSection === 'proposals') {
                    this.loadProposals();
                }
                break;
            case 'treasury_transaction':
                this.showNotification(`New treasury transaction: ${event.data.amount} PX`, 'info');
                if (this.currentSection === 'treasury') {
                    this.loadTreasuryData();
                }
                break;
        }
    }

    updateConnectionStatus(connected) {
        const status = document.getElementById('connectionStatus');
        const icon = status.querySelector('i');
        const text = status.querySelector('span');

        if (connected) {
            status.className = 'connection-status connected';
            text.textContent = 'Connected';
        } else {
            status.className = 'connection-status disconnected';
            text.textContent = 'Disconnected';
        }
    }

    openModal(modalId) {
        document.getElementById(modalId).classList.add('active');
    }

    closeModal(modalId) {
        document.getElementById(modalId).classList.remove('active');
    }

    showNotification(message, type = 'info') {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.innerHTML = `
            <i class="fas fa-${this.getNotificationIcon(type)}"></i>
            <span>${message}</span>
        `;

        // Add to page
        document.body.appendChild(notification);

        // Remove after 5 seconds
        setTimeout(() => {
            notification.remove();
        }, 5000);
    }

    getNotificationIcon(type) {
        const icons = {
            success: 'check-circle',
            error: 'exclamation-circle',
            warning: 'exclamation-triangle',
            info: 'info-circle'
        };
        return icons[type] || 'info-circle';
    }

    formatAddress(address) {
        if (!address) return '';
        return `${address.slice(0, 6)}...${address.slice(-4)}`;
    }

    getAddressInitials(address) {
        if (!address) return '??';
        return address.slice(2, 4).toUpperCase();
    }

    formatTime(timestamp) {
        return new Date(timestamp * 1000).toLocaleDateString();
    }

    getStatusText(status) {
        const statuses = {
            1: 'Pending',
            2: 'Active',
            3: 'Passed',
            4: 'Rejected',
            5: 'Executed',
            6: 'Cancelled'
        };
        return statuses[status] || 'Unknown';
    }

    getStatusClass(status) {
        const classes = {
            1: 'pending',
            2: 'active',
            3: 'passed',
            4: 'rejected',
            5: 'executed',
            6: 'cancelled'
        };
        return classes[status] || 'unknown';
    }

    getProposalTypeText(type) {
        const types = {
            1: 'General',
            2: 'Treasury',
            3: 'Technical',
            4: 'Parameter'
        };
        return types[type] || 'Unknown';
    }
}

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.app = new DAOApp();
});

// Global modal functions for HTML onclick handlers
function closeModal(modalId) {
    window.app.closeModal(modalId);
}