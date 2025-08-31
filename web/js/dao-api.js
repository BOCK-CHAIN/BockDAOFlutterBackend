// DAO API client for interacting with the ProjectX DAO backend
class DAOAPI {
    constructor(baseURL = 'http://localhost:9000') {
        this.baseURL = baseURL;
        this.headers = {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        };
    }

    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            headers: this.headers,
            ...options
        };

        try {
            const response = await fetch(url, config);
            
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
            }

            return await response.json();
        } catch (error) {
            console.error(`API request failed: ${endpoint}`, error);
            throw error;
        }
    }

    // Proposal endpoints
    async getProposals() {
        return this.request('/dao/proposals');
    }

    async getProposal(id) {
        return this.request(`/dao/proposal/${id}`);
    }

    async createProposal(proposalData) {
        return this.request('/dao/proposal', {
            method: 'POST',
            body: JSON.stringify(proposalData)
        });
    }

    async castVote(voteData) {
        return this.request('/dao/vote', {
            method: 'POST',
            body: JSON.stringify(voteData)
        });
    }

    async getProposalVotes(proposalId) {
        return this.request(`/dao/proposal/${proposalId}/votes`);
    }

    // Treasury endpoints
    async getTreasury() {
        return this.request('/dao/treasury');
    }

    async getTreasuryTransactions() {
        return this.request('/dao/treasury/transactions');
    }

    async createTreasuryTransaction(txData) {
        return this.request('/dao/treasury/transaction', {
            method: 'POST',
            body: JSON.stringify(txData)
        });
    }

    async signTreasuryTransaction(txId, privateKey) {
        return this.request('/dao/treasury/sign', {
            method: 'POST',
            body: JSON.stringify({
                transaction_id: txId,
                private_key: privateKey
            })
        });
    }

    // Token endpoints
    async getTokenBalance(address) {
        const response = await this.request(`/dao/token/balance/${address}`);
        return response.balance;
    }

    async getTokenSupply() {
        const response = await this.request('/dao/token/supply');
        return response.total_supply;
    }

    async transferTokens(transferData) {
        return this.request('/dao/token/transfer', {
            method: 'POST',
            body: JSON.stringify(transferData)
        });
    }

    async approveTokens(approvalData) {
        return this.request('/dao/token/approve', {
            method: 'POST',
            body: JSON.stringify(approvalData)
        });
    }

    async getTokenAllowance(owner, spender) {
        const response = await this.request(`/dao/token/allowance/${owner}/${spender}`);
        return response.allowance;
    }

    // Delegation endpoints
    async delegate(delegationData) {
        return this.request('/dao/delegate', {
            method: 'POST',
            body: JSON.stringify(delegationData)
        });
    }

    async revokeDelegation(privateKey) {
        return this.request('/dao/revoke-delegation', {
            method: 'POST',
            body: JSON.stringify({
                private_key: privateKey
            })
        });
    }

    async getDelegation(address) {
        return this.request(`/dao/delegation/${address}`);
    }

    async getDelegations() {
        return this.request('/dao/delegations');
    }

    // Member endpoints
    async getMember(address) {
        return this.request(`/dao/member/${address}`);
    }

    async getMembers(page = 1, limit = 50) {
        return this.request(`/dao/members?page=${page}&limit=${limit}`);
    }

    // Blockchain endpoints (from base API)
    async getBlock(hashOrId) {
        return this.request(`/block/${hashOrId}`);
    }

    async getTransaction(hash) {
        return this.request(`/tx/${hash}`);
    }

    async submitTransaction(txData) {
        return this.request('/tx', {
            method: 'POST',
            body: txData // This should be the encoded transaction
        });
    }

    // Utility methods
    async healthCheck() {
        try {
            await this.request('/dao/treasury');
            return { status: 'healthy', timestamp: Date.now() };
        } catch (error) {
            return { status: 'unhealthy', error: error.message, timestamp: Date.now() };
        }
    }

    // Batch operations
    async batchRequest(requests) {
        const promises = requests.map(req => 
            this.request(req.endpoint, req.options).catch(error => ({ error, ...req }))
        );
        
        return Promise.all(promises);
    }

    // Proposal filtering and search
    async getProposalsByStatus(status) {
        const proposals = await this.getProposals();
        return proposals.filter(p => p.status === status);
    }

    async getProposalsByType(type) {
        const proposals = await this.getProposals();
        return proposals.filter(p => p.proposal_type === type);
    }

    async getActiveProposals() {
        return this.getProposalsByStatus(2); // Active status
    }

    async searchProposals(query) {
        const proposals = await this.getProposals();
        const lowerQuery = query.toLowerCase();
        
        return proposals.filter(p => 
            p.title.toLowerCase().includes(lowerQuery) ||
            p.description.toLowerCase().includes(lowerQuery) ||
            p.creator.toLowerCase().includes(lowerQuery)
        );
    }

    // Treasury analytics
    async getTreasuryAnalytics() {
        const [treasury, transactions] = await Promise.all([
            this.getTreasury(),
            this.getTreasuryTransactions()
        ]);

        const totalSpent = transactions
            .filter(tx => tx.executed)
            .reduce((sum, tx) => sum + tx.amount, 0);

        const pendingAmount = transactions
            .filter(tx => !tx.executed)
            .reduce((sum, tx) => sum + tx.amount, 0);

        return {
            currentBalance: treasury.balance,
            totalSpent,
            pendingAmount,
            availableBalance: treasury.balance - pendingAmount,
            transactionCount: transactions.length,
            executedCount: transactions.filter(tx => tx.executed).length,
            pendingCount: transactions.filter(tx => !tx.executed).length
        };
    }

    // Member analytics
    async getMemberAnalytics() {
        const members = await this.getMembers();
        const memberList = members.members || members;

        const totalBalance = memberList.reduce((sum, m) => sum + m.balance, 0);
        const totalStaked = memberList.reduce((sum, m) => sum + m.staked, 0);
        const totalReputation = memberList.reduce((sum, m) => sum + m.reputation, 0);

        return {
            totalMembers: memberList.length,
            totalBalance,
            totalStaked,
            totalReputation,
            averageBalance: totalBalance / memberList.length,
            averageReputation: totalReputation / memberList.length,
            topHolders: memberList
                .sort((a, b) => b.balance - a.balance)
                .slice(0, 10)
        };
    }

    // Governance analytics
    async getGovernanceAnalytics() {
        const proposals = await this.getProposals();
        
        const statusCounts = proposals.reduce((acc, p) => {
            acc[p.status] = (acc[p.status] || 0) + 1;
            return acc;
        }, {});

        const typeCounts = proposals.reduce((acc, p) => {
            acc[p.proposal_type] = (acc[p.proposal_type] || 0) + 1;
            return acc;
        }, {});

        const passRate = proposals.length > 0 
            ? (statusCounts[3] || 0) / proposals.length * 100 
            : 0;

        return {
            totalProposals: proposals.length,
            statusCounts,
            typeCounts,
            passRate,
            activeProposals: statusCounts[2] || 0,
            recentProposals: proposals
                .sort((a, b) => b.start_time - a.start_time)
                .slice(0, 5)
        };
    }

    // Real-time data polling
    startPolling(callback, interval = 30000) {
        const poll = async () => {
            try {
                const data = await Promise.all([
                    this.getGovernanceAnalytics(),
                    this.getTreasuryAnalytics(),
                    this.getMemberAnalytics()
                ]);
                
                callback({
                    governance: data[0],
                    treasury: data[1],
                    members: data[2],
                    timestamp: Date.now()
                });
            } catch (error) {
                console.error('Polling error:', error);
                callback({ error: error.message, timestamp: Date.now() });
            }
        };

        // Initial poll
        poll();
        
        // Set up interval
        const intervalId = setInterval(poll, interval);
        
        // Return cleanup function
        return () => clearInterval(intervalId);
    }

    // Error handling utilities
    isNetworkError(error) {
        return error.message.includes('fetch') || 
               error.message.includes('network') ||
               error.message.includes('Failed to fetch');
    }

    isServerError(error) {
        return error.message.includes('HTTP 5');
    }

    isClientError(error) {
        return error.message.includes('HTTP 4');
    }

    // Retry mechanism
    async requestWithRetry(endpoint, options = {}, maxRetries = 3) {
        let lastError;
        
        for (let i = 0; i <= maxRetries; i++) {
            try {
                return await this.request(endpoint, options);
            } catch (error) {
                lastError = error;
                
                // Don't retry client errors (4xx)
                if (this.isClientError(error)) {
                    throw error;
                }
                
                // Wait before retry (exponential backoff)
                if (i < maxRetries) {
                    await new Promise(resolve => setTimeout(resolve, Math.pow(2, i) * 1000));
                }
            }
        }
        
        throw lastError;
    }

    // Cache management
    constructor(baseURL = 'http://localhost:9000') {
        this.baseURL = baseURL;
        this.headers = {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        };
        this.cache = new Map();
        this.cacheTimeout = 30000; // 30 seconds
    }

    getCacheKey(endpoint, options) {
        return `${endpoint}_${JSON.stringify(options)}`;
    }

    getCachedData(key) {
        const cached = this.cache.get(key);
        if (cached && Date.now() - cached.timestamp < this.cacheTimeout) {
            return cached.data;
        }
        return null;
    }

    setCachedData(key, data) {
        this.cache.set(key, {
            data,
            timestamp: Date.now()
        });
    }

    async requestWithCache(endpoint, options = {}, useCache = true) {
        if (useCache && options.method !== 'POST') {
            const cacheKey = this.getCacheKey(endpoint, options);
            const cached = this.getCachedData(cacheKey);
            
            if (cached) {
                return cached;
            }
            
            const data = await this.request(endpoint, options);
            this.setCachedData(cacheKey, data);
            return data;
        }
        
        return this.request(endpoint, options);
    }

    clearCache() {
        this.cache.clear();
    }
}