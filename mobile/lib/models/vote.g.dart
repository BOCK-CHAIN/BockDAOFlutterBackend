// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'vote.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

Vote _$VoteFromJson(Map<String, dynamic> json) => Vote(
  proposalId: json['proposalId'] as String,
  voter: json['voter'] as String,
  choice: $enumDecode(_$VoteChoiceEnumMap, json['choice']),
  weight: (json['weight'] as num).toInt(),
  timestamp: DateTime.parse(json['timestamp'] as String),
  reason: json['reason'] as String?,
);

Map<String, dynamic> _$VoteToJson(Vote instance) => <String, dynamic>{
  'proposalId': instance.proposalId,
  'voter': instance.voter,
  'choice': _$VoteChoiceEnumMap[instance.choice]!,
  'weight': instance.weight,
  'timestamp': instance.timestamp.toIso8601String(),
  'reason': instance.reason,
};

const _$VoteChoiceEnumMap = {
  VoteChoice.yes: 'yes',
  VoteChoice.no: 'no',
  VoteChoice.abstain: 'abstain',
};
