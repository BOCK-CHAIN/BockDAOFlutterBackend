// WebSocket client for real-time DAO events
class DAOWebSocket {
    constructor(url = 'ws://localhost:9000/dao/events') {
        this.url = url;
        this.ws = null;
        this.isConnected = false;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000; // Start with 1 second
        this.maxReconnectDelay = 30000; // Max 30 seconds
        this.heartbeatInterval = null;
        this.heartbeatTimeout = null;
        
        // Event callbacks
        this.onEvent = null;
        this.onConnectionChange = null;
        this.onError = null;
        
        // Message queue for when disconnected
        this.messageQueue = [];
        this.maxQueueSize = 100;
    }

    connect() {
        if (this.ws && (this.ws.readyState === WebSocket.CONNECTING || this.ws.readyState === WebSocket.OPEN)) {
            return;
        }

        try {
            console.log('Connecting to WebSocket:', this.url);
            this.ws = new WebSocket(this.url);
            
            this.ws.onopen = (event) => {
                console.log('WebSocket connected');
                this.isConnected = true;
                this.reconnectAttempts = 0;
                this.reconnectDelay = 1000;
                
                this.startHeartbeat();
                this.processMessageQueue();
                
                if (this.onConnectionChange) {
                    this.onConnectionChange(true);
                }
            };

            this.ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.handleMessage(data);
                } catch (error) {
                    console.error('Error parsing WebSocket message:', error);
                }
            };

            this.ws.onclose = (event) => {
                console.log('WebSocket disconnected:', event.code, event.reason);
                this.isConnected = false;
                this.stopHeartbeat();
                
                if (this.onConnectionChange) {
                    this.onConnectionChange(false);
                }

                // Attempt to reconnect unless it was a clean close
                if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
                    this.scheduleReconnect();
                }
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                
                if (this.onError) {
                    this.onError(error);
                }
            };

        } catch (error) {
            console.error('Failed to create WebSocket connection:', error);
            this.scheduleReconnect();
        }
    }

    disconnect() {
        if (this.ws) {
            this.ws.close(1000, 'Client disconnect');
        }
        this.stopHeartbeat();
        this.isConnected = false;
        this.reconnectAttempts = this.maxReconnectAttempts; // Prevent reconnection
    }

    scheduleReconnect() {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.log('Max reconnection attempts reached');
            return;
        }

        this.reconnectAttempts++;
        const delay = Math.min(this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1), this.maxReconnectDelay);
        
        console.log(`Scheduling reconnection attempt ${this.reconnectAttempts} in ${delay}ms`);
        
        setTimeout(() => {
            if (!this.isConnected) {
                this.connect();
            }
        }, delay);
    }

    handleMessage(data) {
        // Handle different message types
        switch (data.type) {
            case 'ping':
                this.sendPong();
                break;
            case 'pong':
                this.handlePong();
                break;
            default:
                // Regular event message
                if (this.onEvent) {
                    this.onEvent(data);
                }
                break;
        }
    }

    send(data) {
        if (this.isConnected && this.ws.readyState === WebSocket.OPEN) {
            try {
                this.ws.send(JSON.stringify(data));
                return true;
            } catch (error) {
                console.error('Error sending WebSocket message:', error);
                this.queueMessage(data);
                return false;
            }
        } else {
            this.queueMessage(data);
            return false;
        }
    }

    queueMessage(data) {
        if (this.messageQueue.length >= this.maxQueueSize) {
            this.messageQueue.shift(); // Remove oldest message
        }
        this.messageQueue.push(data);
    }

    processMessageQueue() {
        while (this.messageQueue.length > 0 && this.isConnected) {
            const message = this.messageQueue.shift();
            this.send(message);
        }
    }

    // Heartbeat mechanism to detect connection issues
    startHeartbeat() {
        this.heartbeatInterval = setInterval(() => {
            if (this.isConnected) {
                this.sendPing();
            }
        }, 30000); // Send ping every 30 seconds
    }

    stopHeartbeat() {
        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
            this.heartbeatInterval = null;
        }
        
        if (this.heartbeatTimeout) {
            clearTimeout(this.heartbeatTimeout);
            this.heartbeatTimeout = null;
        }
    }

    sendPing() {
        this.send({ type: 'ping', timestamp: Date.now() });
        
        // Set timeout for pong response
        this.heartbeatTimeout = setTimeout(() => {
            console.log('Heartbeat timeout - connection may be dead');
            if (this.ws) {
                this.ws.close();
            }
        }, 10000); // Wait 10 seconds for pong
    }

    sendPong() {
        this.send({ type: 'pong', timestamp: Date.now() });
    }

    handlePong() {
        if (this.heartbeatTimeout) {
            clearTimeout(this.heartbeatTimeout);
            this.heartbeatTimeout = null;
        }
    }

    // Event subscription methods
    subscribeToProposalEvents() {
        this.send({
            type: 'subscribe',
            events: ['proposal_created', 'proposal_updated', 'proposal_executed']
        });
    }

    subscribeToVotingEvents() {
        this.send({
            type: 'subscribe',
            events: ['vote_cast', 'voting_ended']
        });
    }

    subscribeToTreasuryEvents() {
        this.send({
            type: 'subscribe',
            events: ['treasury_transaction', 'treasury_signature']
        });
    }

    subscribeToAllEvents() {
        this.send({
            type: 'subscribe',
            events: ['*'] // Subscribe to all events
        });
    }

    unsubscribeFromEvents(events) {
        this.send({
            type: 'unsubscribe',
            events: events
        });
    }

    // Connection status
    getConnectionStatus() {
        return {
            isConnected: this.isConnected,
            readyState: this.ws ? this.ws.readyState : WebSocket.CLOSED,
            reconnectAttempts: this.reconnectAttempts,
            queuedMessages: this.messageQueue.length
        };
    }

    // Event handlers setup
    onProposalCreated(callback) {
        const originalOnEvent = this.onEvent;
        this.onEvent = (event) => {
            if (event.type === 'proposal_created') {
                callback(event);
            }
            if (originalOnEvent) {
                originalOnEvent(event);
            }
        };
    }

    onVoteCast(callback) {
        const originalOnEvent = this.onEvent;
        this.onEvent = (event) => {
            if (event.type === 'vote_cast') {
                callback(event);
            }
            if (originalOnEvent) {
                originalOnEvent(event);
            }
        };
    }

    onTreasuryTransaction(callback) {
        const originalOnEvent = this.onEvent;
        this.onEvent = (event) => {
            if (event.type === 'treasury_transaction') {
                callback(event);
            }
            if (originalOnEvent) {
                originalOnEvent(event);
            }
        };
    }

    onDelegationUpdate(callback) {
        const originalOnEvent = this.onEvent;
        this.onEvent = (event) => {
            if (event.type === 'delegation_updated') {
                callback(event);
            }
            if (originalOnEvent) {
                originalOnEvent(event);
            }
        };
    }

    // Utility methods
    formatEventForDisplay(event) {
        const timestamp = new Date(event.timestamp * 1000).toLocaleString();
        
        switch (event.type) {
            case 'proposal_created':
                return {
                    title: 'New Proposal',
                    message: `"${event.data.title}" created by ${this.formatAddress(event.data.creator)}`,
                    timestamp,
                    icon: 'fas fa-file-alt',
                    color: 'primary'
                };
            
            case 'vote_cast':
                const choice = ['', 'Yes', 'No', 'Abstain'][event.data.choice] || 'Unknown';
                return {
                    title: 'Vote Cast',
                    message: `${this.formatAddress(event.data.voter)} voted ${choice}`,
                    timestamp,
                    icon: 'fas fa-vote-yea',
                    color: 'success'
                };
            
            case 'proposal_passed':
                return {
                    title: 'Proposal Passed',
                    message: `Proposal "${event.data.title}" has passed`,
                    timestamp,
                    icon: 'fas fa-check-circle',
                    color: 'success'
                };
            
            case 'proposal_rejected':
                return {
                    title: 'Proposal Rejected',
                    message: `Proposal "${event.data.title}" was rejected`,
                    timestamp,
                    icon: 'fas fa-times-circle',
                    color: 'danger'
                };
            
            case 'treasury_transaction':
                return {
                    title: 'Treasury Transaction',
                    message: `${event.data.amount} PX transaction created`,
                    timestamp,
                    icon: 'fas fa-coins',
                    color: 'warning'
                };
            
            case 'delegation_updated':
                const action = event.data.action === 'delegate' ? 'delegated to' : 'revoked delegation';
                return {
                    title: 'Delegation Updated',
                    message: `${this.formatAddress(event.data.delegator)} ${action} ${this.formatAddress(event.data.delegate)}`,
                    timestamp,
                    icon: 'fas fa-users',
                    color: 'info'
                };
            
            default:
                return {
                    title: 'DAO Event',
                    message: `${event.type} event occurred`,
                    timestamp,
                    icon: 'fas fa-bell',
                    color: 'secondary'
                };
        }
    }

    formatAddress(address) {
        if (!address) return '';
        return `${address.slice(0, 6)}...${address.slice(-4)}`;
    }

    // Event history management
    constructor(url = 'ws://localhost:9000/dao/events') {
        // ... existing constructor code ...
        this.eventHistory = [];
        this.maxHistorySize = 1000;
    }

    handleMessage(data) {
        // Store event in history
        if (data.type !== 'ping' && data.type !== 'pong') {
            this.addToHistory(data);
        }

        // ... existing handleMessage code ...
    }

    addToHistory(event) {
        if (this.eventHistory.length >= this.maxHistorySize) {
            this.eventHistory.shift(); // Remove oldest event
        }
        
        this.eventHistory.push({
            ...event,
            receivedAt: Date.now()
        });
    }

    getEventHistory(filter = null) {
        if (!filter) {
            return this.eventHistory;
        }

        return this.eventHistory.filter(event => {
            if (filter.type && event.type !== filter.type) {
                return false;
            }
            if (filter.since && event.receivedAt < filter.since) {
                return false;
            }
            if (filter.until && event.receivedAt > filter.until) {
                return false;
            }
            return true;
        });
    }

    clearHistory() {
        this.eventHistory = [];
    }

    // Statistics
    getConnectionStats() {
        return {
            totalReconnectAttempts: this.reconnectAttempts,
            eventsReceived: this.eventHistory.length,
            queuedMessages: this.messageQueue.length,
            isConnected: this.isConnected,
            uptime: this.isConnected ? Date.now() - this.connectionStartTime : 0
        };
    }
}