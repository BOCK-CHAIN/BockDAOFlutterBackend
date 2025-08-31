import 'package:json_annotation/json_annotation.dart';

part 'vote.g.dart';

@JsonSerializable()
class Vote {
  final String proposalId;
  final String voter;
  final VoteChoice choice;
  final int weight;
  final DateTime timestamp;
  final String? reason;

  const Vote({
    required this.proposalId,
    required this.voter,
    required this.choice,
    required this.weight,
    required this.timestamp,
    this.reason,
  });

  factory Vote.fromJson(Map<String, dynamic> json) => _$VoteFromJson(json);
  Map<String, dynamic> toJson() => _$VoteToJson(this);
}

enum VoteChoice {
  @JsonValue('yes')
  yes,
  @JsonValue('no')
  no,
  @JsonValue('abstain')
  abstain,
}