import 'package:json_annotation/json_annotation.dart';

part 'wallet.g.dart';

@JsonSerializable()
class WalletInfo {
  final String address;
  final String publicKey;
  final int tokenBalance;
  final int stakedBalance;
  final int reputation;
  final DateTime joinedAt;
  final DateTime lastActive;
  final bool isConnected;

  const WalletInfo({
    required this.address,
    required this.publicKey,
    required this.tokenBalance,
    required this.stakedBalance,
    required this.reputation,
    required this.joinedAt,
    required this.lastActive,
    required this.isConnected,
  });

  factory WalletInfo.fromJson(Map<String, dynamic> json) => _$WalletInfoFromJson(json);
  Map<String, dynamic> toJson() => _$WalletInfoToJson(this);
}

@JsonSerializable()
class Delegation {
  final String delegator;
  final String delegate;
  final DateTime startTime;
  final DateTime endTime;
  final bool active;

  const Delegation({
    required this.delegator,
    required this.delegate,
    required this.startTime,
    required this.endTime,
    required this.active,
  });

  factory Delegation.fromJson(Map<String, dynamic> json) => _$DelegationFromJson(json);
  Map<String, dynamic> toJson() => _$DelegationToJson(this);
}