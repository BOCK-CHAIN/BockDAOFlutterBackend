import 'dart:convert';
import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:web_socket_channel/status.dart' as status;

class WebSocketService {
  static const String wsUrl = 'ws://10.0.2.2:9000/dao/events';
  
  WebSocketChannel? _channel;
  StreamController<Map<String, dynamic>>? _eventController;
  bool _isConnected = false;
  
  Stream<Map<String, dynamic>> get events => _eventController?.stream ?? const Stream.empty();
  bool get isConnected => _isConnected;

  // Connect to WebSocket
  Future<void> connect() async {
    try {
      _channel = WebSocketChannel.connect(Uri.parse(wsUrl));
      _eventController = StreamController<Map<String, dynamic>>.broadcast();
      
      _channel!.stream.listen(
        (data) {
          try {
            final event = json.decode(data);
            _eventController?.add(event);
          } catch (e) {
            // Log error parsing WebSocket message
            debugPrint('Error parsing WebSocket message: $e');
          }
        },
        onError: (error) {
          // Log WebSocket error
          debugPrint('WebSocket error: $error');
          _handleDisconnection();
        },
        onDone: () {
          // Log WebSocket connection closed
          debugPrint('WebSocket connection closed');
          _handleDisconnection();
        },
      );
      
      _isConnected = true;
      debugPrint('WebSocket connected successfully');
    } catch (e) {
      debugPrint('Failed to connect to WebSocket: $e');
      _handleDisconnection();
    }
  }

  // Disconnect from WebSocket
  Future<void> disconnect() async {
    if (_channel != null) {
      await _channel!.sink.close(status.goingAway);
    }
    _handleDisconnection();
  }

  // Subscribe to specific event types
  void subscribeToProposalEvents(String proposalId) {
    _sendMessage({
      'type': 'subscribe',
      'event': 'proposal_updates',
      'proposalId': proposalId,
    });
  }

  void subscribeToVotingEvents() {
    _sendMessage({
      'type': 'subscribe',
      'event': 'voting_updates',
    });
  }

  void subscribeToTreasuryEvents() {
    _sendMessage({
      'type': 'subscribe',
      'event': 'treasury_updates',
    });
  }

  // Send message to WebSocket
  void _sendMessage(Map<String, dynamic> message) {
    if (_isConnected && _channel != null) {
      _channel!.sink.add(json.encode(message));
    }
  }

  // Handle disconnection
  void _handleDisconnection() {
    _isConnected = false;
    _eventController?.close();
    _eventController = null;
    _channel = null;
  }

  // Dispose resources
  void dispose() {
    disconnect();
  }
}