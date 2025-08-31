// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'proposal.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

Proposal _$ProposalFromJson(Map<String, dynamic> json) => Proposal(
  id: json['id'] as String,
  title: json['title'] as String,
  description: json['description'] as String,
  creator: json['creator'] as String,
  type: $enumDecode(_$ProposalTypeEnumMap, json['type']),
  votingType: $enumDecode(_$VotingTypeEnumMap, json['votingType']),
  status: $enumDecode(_$ProposalStatusEnumMap, json['status']),
  startTime: DateTime.parse(json['startTime'] as String),
  endTime: DateTime.parse(json['endTime'] as String),
  threshold: (json['threshold'] as num).toInt(),
  results: json['results'] == null
      ? null
      : VoteResults.fromJson(json['results'] as Map<String, dynamic>),
  metadataHash: json['metadataHash'] as String?,
);

Map<String, dynamic> _$ProposalToJson(Proposal instance) => <String, dynamic>{
  'id': instance.id,
  'title': instance.title,
  'description': instance.description,
  'creator': instance.creator,
  'type': _$ProposalTypeEnumMap[instance.type]!,
  'votingType': _$VotingTypeEnumMap[instance.votingType]!,
  'status': _$ProposalStatusEnumMap[instance.status]!,
  'startTime': instance.startTime.toIso8601String(),
  'endTime': instance.endTime.toIso8601String(),
  'threshold': instance.threshold,
  'results': instance.results,
  'metadataHash': instance.metadataHash,
};

const _$ProposalTypeEnumMap = {
  ProposalType.general: 'general',
  ProposalType.treasury: 'treasury',
  ProposalType.technical: 'technical',
  ProposalType.parameter: 'parameter',
};

const _$VotingTypeEnumMap = {
  VotingType.simple: 'simple',
  VotingType.quadratic: 'quadratic',
  VotingType.weighted: 'weighted',
  VotingType.reputation: 'reputation',
};

const _$ProposalStatusEnumMap = {
  ProposalStatus.pending: 'pending',
  ProposalStatus.active: 'active',
  ProposalStatus.passed: 'passed',
  ProposalStatus.rejected: 'rejected',
  ProposalStatus.executed: 'executed',
  ProposalStatus.cancelled: 'cancelled',
};

VoteResults _$VoteResultsFromJson(Map<String, dynamic> json) => VoteResults(
  yesVotes: (json['yesVotes'] as num).toInt(),
  noVotes: (json['noVotes'] as num).toInt(),
  abstainVotes: (json['abstainVotes'] as num).toInt(),
  totalVoters: (json['totalVoters'] as num).toInt(),
  quorum: (json['quorum'] as num).toInt(),
  passed: json['passed'] as bool,
);

Map<String, dynamic> _$VoteResultsToJson(VoteResults instance) =>
    <String, dynamic>{
      'yesVotes': instance.yesVotes,
      'noVotes': instance.noVotes,
      'abstainVotes': instance.abstainVotes,
      'totalVoters': instance.totalVoters,
      'quorum': instance.quorum,
      'passed': instance.passed,
    };
