import 'package:json_annotation/json_annotation.dart';

part 'proposal.g.dart';

@JsonSerializable()
class Proposal {
  final String id;
  final String title;
  final String description;
  final String creator;
  final ProposalType type;
  final VotingType votingType;
  final ProposalStatus status;
  final DateTime startTime;
  final DateTime endTime;
  final int threshold;
  final VoteResults? results;
  final String? metadataHash;

  const Proposal({
    required this.id,
    required this.title,
    required this.description,
    required this.creator,
    required this.type,
    required this.votingType,
    required this.status,
    required this.startTime,
    required this.endTime,
    required this.threshold,
    this.results,
    this.metadataHash,
  });

  factory Proposal.fromJson(Map<String, dynamic> json) => _$ProposalFromJson(json);
  Map<String, dynamic> toJson() => _$ProposalToJson(this);
}

@JsonSerializable()
class VoteResults {
  final int yesVotes;
  final int noVotes;
  final int abstainVotes;
  final int totalVoters;
  final int quorum;
  final bool passed;

  const VoteResults({
    required this.yesVotes,
    required this.noVotes,
    required this.abstainVotes,
    required this.totalVoters,
    required this.quorum,
    required this.passed,
  });

  factory VoteResults.fromJson(Map<String, dynamic> json) => _$VoteResultsFromJson(json);
  Map<String, dynamic> toJson() => _$VoteResultsToJson(this);
}

enum ProposalType {
  @JsonValue('general')
  general,
  @JsonValue('treasury')
  treasury,
  @JsonValue('technical')
  technical,
  @JsonValue('parameter')
  parameter,
}

enum VotingType {
  @JsonValue('simple')
  simple,
  @JsonValue('quadratic')
  quadratic,
  @JsonValue('weighted')
  weighted,
  @JsonValue('reputation')
  reputation,
}

enum ProposalStatus {
  @JsonValue('pending')
  pending,
  @JsonValue('active')
  active,
  @JsonValue('passed')
  passed,
  @JsonValue('rejected')
  rejected,
  @JsonValue('executed')
  executed,
  @JsonValue('cancelled')
  cancelled,
}