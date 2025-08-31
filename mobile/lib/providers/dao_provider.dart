import 'package:flutter/foundation.dart';
import '../models/proposal.dart';
import '../models/vote.dart';
import '../models/wallet.dart';
import '../services/api_service.dart';
import '../services/wallet_service.dart';
import '../services/websocket_service.dart';

class DAOProvider extends ChangeNotifier {
  final ApiService _apiService = ApiService();
  final WalletService _walletService = WalletService();
  final WebSocketService _webSocketService = WebSocketService();

  List<Proposal> _proposals = [];
  WalletInfo? _walletInfo;
  Map<String, dynamic>? _treasuryStatus;
  bool _isLoading = false;
  String? _error;

  // Getters
  List<Proposal> get proposals => _proposals;
  WalletInfo? get walletInfo => _walletInfo;
  Map<String, dynamic>? get treasuryStatus => _treasuryStatus;
  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isWalletConnected => _walletService.isConnected;

  // Initialize the provider
  Future<void> initialize() async {
    await _walletService.initialize();
    _walletInfo = _walletService.currentWallet;
    
    if (_walletInfo != null) {
      await _connectWebSocket();
      await loadProposals();
      await loadTreasuryStatus();
    }
    
    notifyListeners();
  }

  // Wallet operations
  Future<void> connectWallet({
    required String address,
    required String publicKey,
    required String privateKey,
  }) async {
    _setLoading(true);
    try {
      _walletInfo = await _walletService.connectWallet(
        WalletType.manual,
        params: {
          'address': address,
          'publicKey': publicKey,
          'privateKey': privateKey,
        },
      );
      
      // Load wallet info from API (simulate for now)
      try {
        final apiWalletInfo = await _apiService.getWalletInfo(address);
        _walletService.updateWalletBalance(
          tokenBalance: apiWalletInfo.tokenBalance,
          stakedBalance: apiWalletInfo.stakedBalance,
          reputation: apiWalletInfo.reputation,
        );
      } catch (e) {
        // If API call fails, use default values
        _walletService.updateWalletBalance(
          tokenBalance: 1000,
          stakedBalance: 0,
          reputation: 100,
        );
      }
      _walletInfo = _walletService.currentWallet;
      
      await _connectWebSocket();
      await loadProposals();
      await loadTreasuryStatus();
      
      _clearError();
    } catch (e) {
      _setError('Failed to connect wallet: $e');
    } finally {
      _setLoading(false);
    }
  }

  Future<void> disconnectWallet() async {
    await _walletService.disconnectWallet();
    await _webSocketService.disconnect();
    _walletInfo = null;
    _proposals.clear();
    _treasuryStatus = null;
    notifyListeners();
  }

  // Proposal operations
  Future<void> loadProposals({
    ProposalStatus? status,
    ProposalType? type,
    String? creator,
  }) async {
    _setLoading(true);
    try {
      _proposals = await _apiService.getProposals(
        status: status,
        type: type,
        creator: creator,
      );
      _clearError();
    } catch (e) {
      _setError('Failed to load proposals: $e');
    } finally {
      _setLoading(false);
    }
  }

  Future<void> createProposal({
    required String title,
    required String description,
    required ProposalType type,
    required VotingType votingType,
    required DateTime startTime,
    required DateTime endTime,
    required int threshold,
    String? metadataHash,
  }) async {
    if (!isWalletConnected) {
      _setError('Wallet not connected');
      return;
    }

    _setLoading(true);
    try {
      await _apiService.createProposal(
        title: title,
        description: description,
        type: type,
        votingType: votingType,
        startTime: startTime,
        endTime: endTime,
        threshold: threshold,
        metadataHash: metadataHash,
      );
      
      // Reload proposals to include the new one
      await loadProposals();
      _clearError();
    } catch (e) {
      _setError('Failed to create proposal: $e');
    } finally {
      _setLoading(false);
    }
  }

  // Voting operations
  Future<void> castVote({
    required String proposalId,
    required VoteChoice choice,
    required int weight,
    String? reason,
  }) async {
    if (!isWalletConnected) {
      _setError('Wallet not connected');
      return;
    }

    _setLoading(true);
    try {
      await _apiService.castVote(
        proposalId: proposalId,
        choice: choice,
        weight: weight,
        reason: reason,
      );
      
      // Reload proposals to update vote counts
      await loadProposals();
      _clearError();
    } catch (e) {
      _setError('Failed to cast vote: $e');
    } finally {
      _setLoading(false);
    }
  }

  // Delegation operations
  Future<void> delegateVoting({
    required String delegate,
    required Duration duration,
  }) async {
    if (!isWalletConnected) {
      _setError('Wallet not connected');
      return;
    }

    _setLoading(true);
    try {
      await _apiService.delegateVoting(
        delegate: delegate,
        duration: duration,
      );
      _clearError();
    } catch (e) {
      _setError('Failed to delegate voting: $e');
    } finally {
      _setLoading(false);
    }
  }

  // Treasury operations
  Future<void> loadTreasuryStatus() async {
    _setLoading(true);
    try {
      _treasuryStatus = await _apiService.getTreasuryStatus();
      _clearError();
    } catch (e) {
      _setError('Failed to load treasury status: $e');
    } finally {
      _setLoading(false);
    }
  }

  // WebSocket connection
  Future<void> _connectWebSocket() async {
    await _webSocketService.connect();
    
    // Subscribe to relevant events
    _webSocketService.subscribeToProposalEvents('all');
    _webSocketService.subscribeToVotingEvents();
    _webSocketService.subscribeToTreasuryEvents();
    
    // Listen to WebSocket events
    _webSocketService.events.listen((event) {
      _handleWebSocketEvent(event);
    });
  }

  // Handle WebSocket events
  void _handleWebSocketEvent(Map<String, dynamic> event) {
    switch (event['type']) {
      case 'proposal_created':
      case 'proposal_updated':
      case 'vote_cast':
        // Reload proposals when there are updates
        loadProposals();
        break;
      case 'treasury_updated':
        // Reload treasury status
        loadTreasuryStatus();
        break;
    }
  }

  // Helper methods
  void _setLoading(bool loading) {
    _isLoading = loading;
    notifyListeners();
  }

  void _setError(String error) {
    _error = error;
    notifyListeners();
  }

  void _clearError() {
    _error = null;
    notifyListeners();
  }

  @override
  void dispose() {
    _apiService.dispose();
    _webSocketService.dispose();
    super.dispose();
  }
}