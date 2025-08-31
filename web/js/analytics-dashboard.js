/**
 * DAO Analytics Dashboard
 * Provides comprehensive analytics and reporting for DAO operations
 */

class AnalyticsDashboard {
    constructor(apiBaseUrl = '') {
        this.apiBaseUrl = apiBaseUrl;
        this.refreshInterval = 30000; // 30 seconds
        this.charts = {};
        this.init();
    }

    async init() {
        this.createDashboardHTML();
        await this.loadAllMetrics();
        this.startAutoRefresh();
        this.setupEventListeners();
    }

    createDashboardHTML() {
        const dashboardHTML = `
            <div id="analytics-dashboard" class="analytics-dashboard">
                <div class="dashboard-header">
                    <h2>DAO Analytics Dashboard</h2>
                    <div class="refresh-controls">
                        <button id="refresh-btn" class="btn btn-primary">Refresh</button>
                        <span id="last-updated" class="last-updated"></span>
                    </div>
                </div>

                <div class="metrics-grid">
                    <!-- Health Overview -->
                    <div class="metric-card health-overview">
                        <h3>DAO Health Overview</h3>
                        <div class="health-score">
                            <div class="score-circle">
                                <span id="overall-score">--</span>
                                <small>Overall Score</small>
                            </div>
                        </div>
                        <div class="health-breakdown">
                            <div class="health-item">
                                <span class="label">Participation</span>
                                <div class="progress-bar">
                                    <div id="participation-progress" class="progress-fill"></div>
                                </div>
                                <span id="participation-score" class="score">--</span>
                            </div>
                            <div class="health-item">
                                <span class="label">Treasury</span>
                                <div class="progress-bar">
                                    <div id="treasury-progress" class="progress-fill"></div>
                                </div>
                                <span id="treasury-score" class="score">--</span>
                            </div>
                            <div class="health-item">
                                <span class="label">Governance</span>
                                <div class="progress-bar">
                                    <div id="governance-progress" class="progress-fill"></div>
                                </div>
                                <span id="governance-score" class="score">--</span>
                            </div>
                            <div class="health-item">
                                <span class="label">Security</span>
                                <div class="progress-bar">
                                    <div id="security-progress" class="progress-fill"></div>
                                </div>
                                <span id="security-score" class="score">--</span>
                            </div>
                        </div>
                        <div id="health-trend" class="health-trend"></div>
                    </div>

                    <!-- Participation Metrics -->
                    <div class="metric-card participation-metrics">
                        <h3>Governance Participation</h3>
                        <div class="key-stats">
                            <div class="stat">
                                <span class="value" id="total-proposals">--</span>
                                <span class="label">Total Proposals</span>
                            </div>
                            <div class="stat">
                                <span class="value" id="active-proposals">--</span>
                                <span class="label">Active Proposals</span>
                            </div>
                            <div class="stat">
                                <span class="value" id="unique-voters">--</span>
                                <span class="label">Unique Voters</span>
                            </div>
                            <div class="stat">
                                <span class="value" id="participation-rate">--%</span>
                                <span class="label">Participation Rate</span>
                            </div>
                        </div>
                        <div class="chart-container">
                            <canvas id="voting-patterns-chart"></canvas>
                        </div>
                    </div>

                    <!-- Treasury Performance -->
                    <div class="metric-card treasury-metrics">
                        <h3>Treasury Performance</h3>
                        <div class="key-stats">
                            <div class="stat">
                                <span class="value" id="treasury-balance">--</span>
                                <span class="label">Current Balance</span>
                            </div>
                            <div class="stat">
                                <span class="value" id="total-outflows">--</span>
                                <span class="label">Total Outflows</span>
                            </div>
                            <div class="stat">
                                <span class="value" id="transaction-count">--</span>
                                <span class="label">Transactions</span>
                            </div>
                            <div class="stat">
                                <span class="value" id="signing-efficiency">--%</span>
                                <span class="label">Signing Efficiency</span>
                            </div>
                        </div>
                        <div class="chart-container">
                            <canvas id="treasury-flow-chart"></canvas>
                        </div>
                    </div>

                    <!-- Proposal Analytics -->
                    <div class="metric-card proposal-analytics">
                        <h3>Proposal Analytics</h3>
                        <div class="key-stats">
                            <div class="stat">
                                <span class="value" id="success-rate">--%</span>
                                <span class="label">Success Rate</span>
                            </div>
                            <div class="stat">
                                <span class="value" id="passed-proposals">--</span>
                                <span class="label">Passed</span>
                            </div>
                            <div class="stat">
                                <span class="value" id="rejected-proposals">--</span>
                                <span class="label">Rejected</span>
                            </div>
                            <div class="stat">
                                <span class="value" id="quorum-rate">--%</span>
                                <span class="label">Quorum Rate</span>
                            </div>
                        </div>
                        <div class="chart-container">
                            <canvas id="proposal-types-chart"></canvas>
                        </div>
                    </div>

                    <!-- Top Participants -->
                    <div class="metric-card top-participants">
                        <h3>Top Participants</h3>
                        <div id="participants-list" class="participants-list">
                            <!-- Populated dynamically -->
                        </div>
                    </div>

                    <!-- Risk Indicators -->
                    <div class="metric-card risk-indicators">
                        <h3>Risk Indicators</h3>
                        <div id="risk-list" class="risk-list">
                            <!-- Populated dynamically -->
                        </div>
                    </div>

                    <!-- Recommendations -->
                    <div class="metric-card recommendations">
                        <h3>Recommendations</h3>
                        <div id="recommendations-list" class="recommendations-list">
                            <!-- Populated dynamically -->
                        </div>
                    </div>

                    <!-- Delegation Analytics -->
                    <div class="metric-card delegation-analytics">
                        <h3>Delegation Analytics</h3>
                        <div class="key-stats">
                            <div class="stat">
                                <span class="value" id="total-delegations">--</span>
                                <span class="label">Total Delegations</span>
                            </div>
                            <div class="stat">
                                <span class="value" id="active-delegations">--</span>
                                <span class="label">Active</span>
                            </div>
                            <div class="stat">
                                <span class="value" id="delegation-rate">--%</span>
                                <span class="label">Delegation Rate</span>
                            </div>
                        </div>
                        <div id="top-delegates" class="delegates-list">
                            <!-- Populated dynamically -->
                        </div>
                    </div>
                </div>
            </div>
        `;

        // Insert dashboard into the page
        const container = document.getElementById('analytics-container') || document.body;
        container.innerHTML = dashboardHTML;
    }

    async loadAllMetrics() {
        try {
            const [healthMetrics, participationMetrics, treasuryMetrics, proposalAnalytics] = await Promise.all([
                this.fetchHealthMetrics(),
                this.fetchParticipationMetrics(),
                this.fetchTreasuryMetrics(),
                this.fetchProposalAnalytics()
            ]);

            this.updateHealthOverview(healthMetrics);
            this.updateParticipationMetrics(participationMetrics);
            this.updateTreasuryMetrics(treasuryMetrics);
            this.updateProposalAnalytics(proposalAnalytics);

            this.updateLastRefreshed();
        } catch (error) {
            console.error('Error loading metrics:', error);
            this.showError('Failed to load analytics data');
        }
    }

    async fetchHealthMetrics() {
        const response = await fetch(`${this.apiBaseUrl}/dao/analytics/health`);
        if (!response.ok) throw new Error('Failed to fetch health metrics');
        return response.json();
    }

    async fetchParticipationMetrics() {
        const response = await fetch(`${this.apiBaseUrl}/dao/analytics/participation`);
        if (!response.ok) throw new Error('Failed to fetch participation metrics');
        return response.json();
    }

    async fetchTreasuryMetrics() {
        const response = await fetch(`${this.apiBaseUrl}/dao/analytics/treasury`);
        if (!response.ok) throw new Error('Failed to fetch treasury metrics');
        return response.json();
    }

    async fetchProposalAnalytics() {
        const response = await fetch(`${this.apiBaseUrl}/dao/analytics/proposals`);
        if (!response.ok) throw new Error('Failed to fetch proposal analytics');
        return response.json();
    }

    updateHealthOverview(health) {
        document.getElementById('overall-score').textContent = Math.round(health.overall_score);
        document.getElementById('participation-score').textContent = Math.round(health.participation_health);
        document.getElementById('treasury-score').textContent = Math.round(health.treasury_health);
        document.getElementById('governance-score').textContent = Math.round(health.governance_health);
        document.getElementById('security-score').textContent = Math.round(health.security_health);

        // Update progress bars
        this.updateProgressBar('participation-progress', health.participation_health);
        this.updateProgressBar('treasury-progress', health.treasury_health);
        this.updateProgressBar('governance-progress', health.governance_health);
        this.updateProgressBar('security-progress', health.security_health);

        // Update health trend
        const trendElement = document.getElementById('health-trend');
        trendElement.textContent = `Trend: ${health.health_trend}`;
        trendElement.className = `health-trend ${health.health_trend.toLowerCase()}`;

        // Update risk indicators
        this.updateRiskIndicators(health.risk_indicators);

        // Update recommendations
        this.updateRecommendations(health.recommendations);
    }

    updateParticipationMetrics(participation) {
        document.getElementById('total-proposals').textContent = participation.total_proposals;
        document.getElementById('active-proposals').textContent = participation.active_proposals;
        document.getElementById('unique-voters').textContent = participation.unique_voters;
        document.getElementById('participation-rate').textContent = `${Math.round(participation.participation_rate)}%`;

        // Update top participants
        this.updateTopParticipants(participation.top_participants);

        // Update delegation analytics
        this.updateDelegationAnalytics(participation.delegation_metrics);

        // Create voting patterns chart
        this.createVotingPatternsChart(participation.voting_patterns);
    }

    updateTreasuryMetrics(treasury) {
        document.getElementById('treasury-balance').textContent = this.formatTokenAmount(treasury.current_balance);
        document.getElementById('total-outflows').textContent = this.formatTokenAmount(treasury.total_outflows);
        document.getElementById('transaction-count').textContent = treasury.transaction_count;
        document.getElementById('signing-efficiency').textContent = `${Math.round(treasury.signing_efficiency)}%`;

        // Create treasury flow chart if data available
        if (treasury.monthly_flows && treasury.monthly_flows.length > 0) {
            this.createTreasuryFlowChart(treasury.monthly_flows);
        }
    }

    updateProposalAnalytics(proposals) {
        document.getElementById('success-rate').textContent = `${Math.round(proposals.success_rate)}%`;
        document.getElementById('passed-proposals').textContent = proposals.passed_proposals;
        document.getElementById('rejected-proposals').textContent = proposals.rejected_proposals;
        document.getElementById('quorum-rate').textContent = `${Math.round(proposals.quorum_achievement_rate)}%`;

        // Create proposal types chart
        this.createProposalTypesChart(proposals.success_rate_by_type);
    }

    updateProgressBar(elementId, value) {
        const progressBar = document.getElementById(elementId);
        if (progressBar) {
            progressBar.style.width = `${Math.min(value, 100)}%`;
            
            // Color coding based on value
            if (value >= 75) {
                progressBar.className = 'progress-fill good';
            } else if (value >= 50) {
                progressBar.className = 'progress-fill okay';
            } else {
                progressBar.className = 'progress-fill poor';
            }
        }
    }

    updateTopParticipants(participants) {
        const container = document.getElementById('participants-list');
        if (!container || !participants) return;

        container.innerHTML = participants.slice(0, 5).map((participant, index) => `
            <div class="participant-item">
                <div class="rank">#${index + 1}</div>
                <div class="participant-info">
                    <div class="address">${this.truncateAddress(participant.address)}</div>
                    <div class="stats">
                        <span>${participant.votes_cast} votes</span>
                        <span>${participant.proposals_created} proposals</span>
                        <span>${Math.round(participant.participation_rate)}% participation</span>
                    </div>
                </div>
                <div class="reputation">${participant.reputation}</div>
            </div>
        `).join('');
    }

    updateDelegationAnalytics(delegation) {
        if (!delegation) return;

        document.getElementById('total-delegations').textContent = delegation.total_delegations;
        document.getElementById('active-delegations').textContent = delegation.active_delegations;
        document.getElementById('delegation-rate').textContent = `${Math.round(delegation.delegation_rate)}%`;

        // Update top delegates
        const container = document.getElementById('top-delegates');
        if (container && delegation.top_delegates) {
            container.innerHTML = delegation.top_delegates.slice(0, 3).map((delegate, index) => `
                <div class="delegate-item">
                    <div class="rank">#${index + 1}</div>
                    <div class="delegate-info">
                        <div class="address">${this.truncateAddress(delegate.address)}</div>
                        <div class="stats">
                            <span>${delegate.delegators_count} delegators</span>
                            <span>${this.formatTokenAmount(delegate.total_voting_power)} power</span>
                        </div>
                    </div>
                </div>
            `).join('');
        }
    }

    updateRiskIndicators(risks) {
        const container = document.getElementById('risk-list');
        if (!container || !risks) return;

        if (risks.length === 0) {
            container.innerHTML = '<div class="no-risks">No significant risks detected</div>';
            return;
        }

        container.innerHTML = risks.map(risk => `
            <div class="risk-item ${risk.severity.toLowerCase()}">
                <div class="risk-header">
                    <span class="risk-type">${risk.type}</span>
                    <span class="risk-severity">${risk.severity}</span>
                </div>
                <div class="risk-description">${risk.description}</div>
                <div class="risk-mitigation">${risk.mitigation}</div>
            </div>
        `).join('');
    }

    updateRecommendations(recommendations) {
        const container = document.getElementById('recommendations-list');
        if (!container || !recommendations) return;

        if (recommendations.length === 0) {
            container.innerHTML = '<div class="no-recommendations">No recommendations at this time</div>';
            return;
        }

        container.innerHTML = recommendations.map(rec => `
            <div class="recommendation-item">
                <i class="icon-lightbulb"></i>
                <span>${rec}</span>
            </div>
        `).join('');
    }

    createVotingPatternsChart(patterns) {
        const ctx = document.getElementById('voting-patterns-chart');
        if (!ctx || !patterns) return;

        if (this.charts.votingPatterns) {
            this.charts.votingPatterns.destroy();
        }

        this.charts.votingPatterns = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: ['Yes', 'No', 'Abstain'],
                datasets: [{
                    data: [patterns['1'] || 0, patterns['2'] || 0, patterns['3'] || 0],
                    backgroundColor: ['#4CAF50', '#F44336', '#FF9800']
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom'
                    }
                }
            }
        });
    }

    createTreasuryFlowChart(flows) {
        const ctx = document.getElementById('treasury-flow-chart');
        if (!ctx || !flows) return;

        if (this.charts.treasuryFlow) {
            this.charts.treasuryFlow.destroy();
        }

        this.charts.treasuryFlow = new Chart(ctx, {
            type: 'line',
            data: {
                labels: flows.map(f => new Date(f.timestamp * 1000).toLocaleDateString()),
                datasets: [{
                    label: 'Balance',
                    data: flows.map(f => f.balance),
                    borderColor: '#2196F3',
                    backgroundColor: 'rgba(33, 150, 243, 0.1)',
                    fill: true
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    }

    createProposalTypesChart(successRates) {
        const ctx = document.getElementById('proposal-types-chart');
        if (!ctx || !successRates) return;

        if (this.charts.proposalTypes) {
            this.charts.proposalTypes.destroy();
        }

        const typeNames = {
            '1': 'General',
            '2': 'Treasury',
            '3': 'Technical',
            '4': 'Parameter'
        };

        const labels = Object.keys(successRates).map(key => typeNames[key] || `Type ${key}`);
        const data = Object.values(successRates);

        this.charts.proposalTypes = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [{
                    label: 'Success Rate (%)',
                    data: data,
                    backgroundColor: '#4CAF50',
                    borderColor: '#388E3C',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        max: 100
                    }
                }
            }
        });
    }

    setupEventListeners() {
        document.getElementById('refresh-btn').addEventListener('click', () => {
            this.loadAllMetrics();
        });
    }

    startAutoRefresh() {
        setInterval(() => {
            this.loadAllMetrics();
        }, this.refreshInterval);
    }

    updateLastRefreshed() {
        const now = new Date();
        document.getElementById('last-updated').textContent = 
            `Last updated: ${now.toLocaleTimeString()}`;
    }

    formatTokenAmount(amount) {
        if (amount >= 1000000) {
            return `${(amount / 1000000).toFixed(1)}M`;
        } else if (amount >= 1000) {
            return `${(amount / 1000).toFixed(1)}K`;
        }
        return amount.toString();
    }

    truncateAddress(address) {
        if (!address || address.length < 10) return address;
        return `${address.slice(0, 6)}...${address.slice(-4)}`;
    }

    showError(message) {
        // Simple error display - could be enhanced with a proper notification system
        console.error(message);
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error-message';
        errorDiv.textContent = message;
        document.body.appendChild(errorDiv);
        
        setTimeout(() => {
            document.body.removeChild(errorDiv);
        }, 5000);
    }
}

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Check if Chart.js is available
    if (typeof Chart === 'undefined') {
        console.warn('Chart.js not found. Charts will not be displayed.');
    }
    
    // Initialize the dashboard
    window.analyticsDashboard = new AnalyticsDashboard();
});

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = AnalyticsDashboard;
}