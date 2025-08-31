// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'wallet.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

WalletInfo _$WalletInfoFromJson(Map<String, dynamic> json) => WalletInfo(
  address: json['address'] as String,
  publicKey: json['publicKey'] as String,
  tokenBalance: (json['tokenBalance'] as num).toInt(),
  stakedBalance: (json['stakedBalance'] as num).toInt(),
  reputation: (json['reputation'] as num).toInt(),
  joinedAt: DateTime.parse(json['joinedAt'] as String),
  lastActive: DateTime.parse(json['lastActive'] as String),
  isConnected: json['isConnected'] as bool,
);

Map<String, dynamic> _$WalletInfoToJson(WalletInfo instance) =>
    <String, dynamic>{
      'address': instance.address,
      'publicKey': instance.publicKey,
      'tokenBalance': instance.tokenBalance,
      'stakedBalance': instance.stakedBalance,
      'reputation': instance.reputation,
      'joinedAt': instance.joinedAt.toIso8601String(),
      'lastActive': instance.lastActive.toIso8601String(),
      'isConnected': instance.isConnected,
    };

Delegation _$DelegationFromJson(Map<String, dynamic> json) => Delegation(
  delegator: json['delegator'] as String,
  delegate: json['delegate'] as String,
  startTime: DateTime.parse(json['startTime'] as String),
  endTime: DateTime.parse(json['endTime'] as String),
  active: json['active'] as bool,
);

Map<String, dynamic> _$DelegationToJson(Delegation instance) =>
    <String, dynamic>{
      'delegator': instance.delegator,
      'delegate': instance.delegate,
      'startTime': instance.startTime.toIso8601String(),
      'endTime': instance.endTime.toIso8601String(),
      'active': instance.active,
    };
