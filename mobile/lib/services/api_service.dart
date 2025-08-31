import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/proposal.dart';
import '../models/vote.dart';
import '../models/wallet.dart';

class ApiService {
  static const String baseUrl = 'http://10.0.2.2:9000'; // ProjectX API endpoint (Android emulator)
  
  final http.Client _client = http.Client();

  // Proposal endpoints
  Future<List<Proposal>> getProposals({
    ProposalStatus? status,
    ProposalType? type,
    String? creator,
  }) async {
    final queryParams = <String, String>{};
    if (status != null) queryParams['status'] = status.name;
    if (type != null) queryParams['type'] = type.name;
    if (creator != null) queryParams['creator'] = creator;

    final uri = Uri.parse('$baseUrl/dao/proposals').replace(queryParameters: queryParams);
    final response = await _client.get(uri);

    if (response.statusCode == 200) {
      final List<dynamic> data = json.decode(response.body);
      return data.map((json) => Proposal.fromJson(json)).toList();
    } else {
      throw Exception('Failed to load proposals: ${response.statusCode}');
    }
  }

  Future<Proposal> getProposal(String id) async {
    final response = await _client.get(Uri.parse('$baseUrl/dao/proposal/$id'));

    if (response.statusCode == 200) {
      return Proposal.fromJson(json.decode(response.body));
    } else {
      throw Exception('Failed to load proposal: ${response.statusCode}');
    }
  }

  Future<String> createProposal({
    required String title,
    required String description,
    required ProposalType type,
    required VotingType votingType,
    required DateTime startTime,
    required DateTime endTime,
    required int threshold,
    String? metadataHash,
  }) async {
    final response = await _client.post(
      Uri.parse('$baseUrl/dao/proposal'),
      headers: {'Content-Type': 'application/json'},
      body: json.encode({
        'title': title,
        'description': description,
        'type': type.name,
        'votingType': votingType.name,
        'startTime': startTime.toIso8601String(),
        'endTime': endTime.toIso8601String(),
        'threshold': threshold,
        if (metadataHash != null) 'metadataHash': metadataHash,
      }),
    );

    if (response.statusCode == 201) {
      final data = json.decode(response.body);
      return data['proposalId'];
    } else {
      throw Exception('Failed to create proposal: ${response.statusCode}');
    }
  }

  // Voting endpoints
  Future<void> castVote({
    required String proposalId,
    required VoteChoice choice,
    required int weight,
    String? reason,
  }) async {
    final response = await _client.post(
      Uri.parse('$baseUrl/dao/vote'),
      headers: {'Content-Type': 'application/json'},
      body: json.encode({
        'proposalId': proposalId,
        'choice': choice.name,
        'weight': weight,
        if (reason != null) 'reason': reason,
      }),
    );

    if (response.statusCode != 200) {
      throw Exception('Failed to cast vote: ${response.statusCode}');
    }
  }

  // Wallet endpoints
  Future<WalletInfo> getWalletInfo(String address) async {
    final response = await _client.get(Uri.parse('$baseUrl/dao/member/$address'));

    if (response.statusCode == 200) {
      return WalletInfo.fromJson(json.decode(response.body));
    } else {
      throw Exception('Failed to load wallet info: ${response.statusCode}');
    }
  }

  // Delegation endpoints
  Future<void> delegateVoting({
    required String delegate,
    required Duration duration,
  }) async {
    final response = await _client.post(
      Uri.parse('$baseUrl/dao/delegate'),
      headers: {'Content-Type': 'application/json'},
      body: json.encode({
        'delegate': delegate,
        'duration': duration.inSeconds,
      }),
    );

    if (response.statusCode != 200) {
      throw Exception('Failed to delegate voting: ${response.statusCode}');
    }
  }

  // Treasury endpoints
  Future<Map<String, dynamic>> getTreasuryStatus() async {
    final response = await _client.get(Uri.parse('$baseUrl/dao/treasury'));

    if (response.statusCode == 200) {
      return json.decode(response.body);
    } else {
      throw Exception('Failed to load treasury status: ${response.statusCode}');
    }
  }

  void dispose() {
    _client.close();
  }
}