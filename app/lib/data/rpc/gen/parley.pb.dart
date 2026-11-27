// This is a generated file - do not edit.
//
// Generated from parley.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'parley.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'parley.pbenum.dart';

class Empty extends $pb.GeneratedMessage {
  factory Empty() => Empty._();

  Empty._();

  factory Empty.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Empty()..mergeFromBuffer(data, registry);
  factory Empty.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Empty()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Empty',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Empty.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Empty clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Empty copyWith(void Function(Empty) updates) =>
      super.copyWith((message) => updates(message as Empty)) as Empty;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Empty() / Empty.new instead')
  static Empty create() => Empty._();
  static $pb.GeneratedMessage $_createMessage() => Empty._();
  @$core.override
  Empty createEmptyInstance() => Empty._();
  @$core.pragma('dart2js:noInline')
  static Empty getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Empty>(Empty.$_createMessage);
  static Empty? _defaultInstance;
}

class IdRequest extends $pb.GeneratedMessage {
  factory IdRequest({
    $core.String? id,
  }) {
    final result = IdRequest._();
    if (id != null) result.id = id;
    return result;
  }

  IdRequest._();

  factory IdRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      IdRequest()..mergeFromBuffer(data, registry);
  factory IdRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      IdRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'IdRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: IdRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  IdRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  IdRequest copyWith(void Function(IdRequest) updates) =>
      super.copyWith((message) => updates(message as IdRequest)) as IdRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use IdRequest() / IdRequest.new instead')
  static IdRequest create() => IdRequest._();
  static $pb.GeneratedMessage $_createMessage() => IdRequest._();
  @$core.override
  IdRequest createEmptyInstance() => IdRequest._();
  @$core.pragma('dart2js:noInline')
  static IdRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<IdRequest>(IdRequest.$_createMessage);
  static IdRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class PasswordRequest extends $pb.GeneratedMessage {
  factory PasswordRequest({
    $core.String? password,
  }) {
    final result = PasswordRequest._();
    if (password != null) result.password = password;
    return result;
  }

  PasswordRequest._();

  factory PasswordRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PasswordRequest()..mergeFromBuffer(data, registry);
  factory PasswordRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PasswordRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PasswordRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: PasswordRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'password')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PasswordRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PasswordRequest copyWith(void Function(PasswordRequest) updates) =>
      super.copyWith((message) => updates(message as PasswordRequest))
          as PasswordRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use PasswordRequest() / PasswordRequest.new instead')
  static PasswordRequest create() => PasswordRequest._();
  static $pb.GeneratedMessage $_createMessage() => PasswordRequest._();
  @$core.override
  PasswordRequest createEmptyInstance() => PasswordRequest._();
  @$core.pragma('dart2js:noInline')
  static PasswordRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PasswordRequest>(
          PasswordRequest.$_createMessage);
  static PasswordRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get password => $_getSZ(0);
  @$pb.TagNumber(1)
  set password($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPassword() => $_has(0);
  @$pb.TagNumber(1)
  void clearPassword() => $_clearField(1);
}

class VaultPathRequest extends $pb.GeneratedMessage {
  factory VaultPathRequest({
    $core.String? path,
    $core.bool? move,
  }) {
    final result = VaultPathRequest._();
    if (path != null) result.path = path;
    if (move != null) result.move = move;
    return result;
  }

  VaultPathRequest._();

  factory VaultPathRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VaultPathRequest()..mergeFromBuffer(data, registry);
  factory VaultPathRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VaultPathRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'VaultPathRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: VaultPathRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'path')
    ..aOB(2, _omitFieldNames ? '' : 'move')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VaultPathRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VaultPathRequest copyWith(void Function(VaultPathRequest) updates) =>
      super.copyWith((message) => updates(message as VaultPathRequest))
          as VaultPathRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use VaultPathRequest() / VaultPathRequest.new instead')
  static VaultPathRequest create() => VaultPathRequest._();
  static $pb.GeneratedMessage $_createMessage() => VaultPathRequest._();
  @$core.override
  VaultPathRequest createEmptyInstance() => VaultPathRequest._();
  @$core.pragma('dart2js:noInline')
  static VaultPathRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<VaultPathRequest>(
          VaultPathRequest.$_createMessage);
  static VaultPathRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get path => $_getSZ(0);
  @$pb.TagNumber(1)
  set path($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPath() => $_has(0);
  @$pb.TagNumber(1)
  void clearPath() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get move => $_getBF(1);
  @$pb.TagNumber(2)
  set move($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMove() => $_has(1);
  @$pb.TagNumber(2)
  void clearMove() => $_clearField(2);
}

class ChangePasswordRequest extends $pb.GeneratedMessage {
  factory ChangePasswordRequest({
    $core.String? oldPassword,
    $core.String? newPassword,
  }) {
    final result = ChangePasswordRequest._();
    if (oldPassword != null) result.oldPassword = oldPassword;
    if (newPassword != null) result.newPassword = newPassword;
    return result;
  }

  ChangePasswordRequest._();

  factory ChangePasswordRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChangePasswordRequest()..mergeFromBuffer(data, registry);
  factory ChangePasswordRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChangePasswordRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ChangePasswordRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ChangePasswordRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'oldPassword')
    ..aOS(2, _omitFieldNames ? '' : 'newPassword')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChangePasswordRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChangePasswordRequest copyWith(
          void Function(ChangePasswordRequest) updates) =>
      super.copyWith((message) => updates(message as ChangePasswordRequest))
          as ChangePasswordRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ChangePasswordRequest() / ChangePasswordRequest.new instead')
  static ChangePasswordRequest create() => ChangePasswordRequest._();
  static $pb.GeneratedMessage $_createMessage() => ChangePasswordRequest._();
  @$core.override
  ChangePasswordRequest createEmptyInstance() => ChangePasswordRequest._();
  @$core.pragma('dart2js:noInline')
  static ChangePasswordRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ChangePasswordRequest>(
          ChangePasswordRequest.$_createMessage);
  static ChangePasswordRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get oldPassword => $_getSZ(0);
  @$pb.TagNumber(1)
  set oldPassword($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasOldPassword() => $_has(0);
  @$pb.TagNumber(1)
  void clearOldPassword() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get newPassword => $_getSZ(1);
  @$pb.TagNumber(2)
  set newPassword($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasNewPassword() => $_has(1);
  @$pb.TagNumber(2)
  void clearNewPassword() => $_clearField(2);
}

class State extends $pb.GeneratedMessage {
  factory State({
    $core.bool? vaultExists,
    $core.bool? unlocked,
    $core.String? vaultPath,
    $core.String? platform,
    $core.String? coreVersion,
    $core.String? coreBuild,
    $core.int? accounts,
    $core.int? accountsOnline,
    $core.bool? canRemember,
    $core.bool? remembered,
    ShardState? shard,
  }) {
    final result = State._();
    if (vaultExists != null) result.vaultExists = vaultExists;
    if (unlocked != null) result.unlocked = unlocked;
    if (vaultPath != null) result.vaultPath = vaultPath;
    if (platform != null) result.platform = platform;
    if (coreVersion != null) result.coreVersion = coreVersion;
    if (coreBuild != null) result.coreBuild = coreBuild;
    if (accounts != null) result.accounts = accounts;
    if (accountsOnline != null) result.accountsOnline = accountsOnline;
    if (canRemember != null) result.canRemember = canRemember;
    if (remembered != null) result.remembered = remembered;
    if (shard != null) result.shard = shard;
    return result;
  }

  State._();

  factory State.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      State()..mergeFromBuffer(data, registry);
  factory State.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      State()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'State',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: State.$_createMessage)
    ..aOB(1, _omitFieldNames ? '' : 'vaultExists')
    ..aOB(2, _omitFieldNames ? '' : 'unlocked')
    ..aOS(3, _omitFieldNames ? '' : 'vaultPath')
    ..aOS(4, _omitFieldNames ? '' : 'platform')
    ..aOS(5, _omitFieldNames ? '' : 'coreVersion')
    ..aOS(6, _omitFieldNames ? '' : 'coreBuild')
    ..aI(7, _omitFieldNames ? '' : 'accounts')
    ..aI(8, _omitFieldNames ? '' : 'accountsOnline')
    ..aOB(10, _omitFieldNames ? '' : 'canRemember')
    ..aOB(11, _omitFieldNames ? '' : 'remembered')
    ..aOM<ShardState>(12, _omitFieldNames ? '' : 'shard',
        subBuilder: ShardState.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  State clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  State copyWith(void Function(State) updates) =>
      super.copyWith((message) => updates(message as State)) as State;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use State() / State.new instead')
  static State create() => State._();
  static $pb.GeneratedMessage $_createMessage() => State._();
  @$core.override
  State createEmptyInstance() => State._();
  @$core.pragma('dart2js:noInline')
  static State getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<State>(State.$_createMessage);
  static State? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get vaultExists => $_getBF(0);
  @$pb.TagNumber(1)
  set vaultExists($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasVaultExists() => $_has(0);
  @$pb.TagNumber(1)
  void clearVaultExists() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get unlocked => $_getBF(1);
  @$pb.TagNumber(2)
  set unlocked($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUnlocked() => $_has(1);
  @$pb.TagNumber(2)
  void clearUnlocked() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get vaultPath => $_getSZ(2);
  @$pb.TagNumber(3)
  set vaultPath($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasVaultPath() => $_has(2);
  @$pb.TagNumber(3)
  void clearVaultPath() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get platform => $_getSZ(3);
  @$pb.TagNumber(4)
  set platform($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPlatform() => $_has(3);
  @$pb.TagNumber(4)
  void clearPlatform() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get coreVersion => $_getSZ(4);
  @$pb.TagNumber(5)
  set coreVersion($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCoreVersion() => $_has(4);
  @$pb.TagNumber(5)
  void clearCoreVersion() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get coreBuild => $_getSZ(5);
  @$pb.TagNumber(6)
  set coreBuild($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCoreBuild() => $_has(5);
  @$pb.TagNumber(6)
  void clearCoreBuild() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.int get accounts => $_getIZ(6);
  @$pb.TagNumber(7)
  set accounts($core.int value) => $_setSignedInt32(6, value);
  @$pb.TagNumber(7)
  $core.bool hasAccounts() => $_has(6);
  @$pb.TagNumber(7)
  void clearAccounts() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.int get accountsOnline => $_getIZ(7);
  @$pb.TagNumber(8)
  set accountsOnline($core.int value) => $_setSignedInt32(7, value);
  @$pb.TagNumber(8)
  $core.bool hasAccountsOnline() => $_has(7);
  @$pb.TagNumber(8)
  void clearAccountsOnline() => $_clearField(8);

  @$pb.TagNumber(10)
  $core.bool get canRemember => $_getBF(8);
  @$pb.TagNumber(10)
  set canRemember($core.bool value) => $_setBool(8, value);
  @$pb.TagNumber(10)
  $core.bool hasCanRemember() => $_has(8);
  @$pb.TagNumber(10)
  void clearCanRemember() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.bool get remembered => $_getBF(9);
  @$pb.TagNumber(11)
  set remembered($core.bool value) => $_setBool(9, value);
  @$pb.TagNumber(11)
  $core.bool hasRemembered() => $_has(9);
  @$pb.TagNumber(11)
  void clearRemembered() => $_clearField(11);

  @$pb.TagNumber(12)
  ShardState get shard => $_getN(10);
  @$pb.TagNumber(12)
  set shard(ShardState value) => $_setField(12, value);
  @$pb.TagNumber(12)
  $core.bool hasShard() => $_has(10);
  @$pb.TagNumber(12)
  void clearShard() => $_clearField(12);
  @$pb.TagNumber(12)
  ShardState ensureShard() => $_ensure(10);
}

class ShardState extends $pb.GeneratedMessage {
  factory ShardState({
    ShardState_Status? status,
    $core.bool? enabled,
    $core.bool? available,
    $core.String? pairingCode,
    $core.String? message,
    $fixnum.Int64? syncedMs,
    $core.String? deviceName,
    $core.bool? canMerge,
    $core.bool? unpaired,
    $core.bool? canOpen,
  }) {
    final result = ShardState._();
    if (status != null) result.status = status;
    if (enabled != null) result.enabled = enabled;
    if (available != null) result.available = available;
    if (pairingCode != null) result.pairingCode = pairingCode;
    if (message != null) result.message = message;
    if (syncedMs != null) result.syncedMs = syncedMs;
    if (deviceName != null) result.deviceName = deviceName;
    if (canMerge != null) result.canMerge = canMerge;
    if (unpaired != null) result.unpaired = unpaired;
    if (canOpen != null) result.canOpen = canOpen;
    return result;
  }

  ShardState._();

  factory ShardState.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ShardState()..mergeFromBuffer(data, registry);
  factory ShardState.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ShardState()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ShardState',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ShardState.$_createMessage)
    ..aE<ShardState_Status>(1, _omitFieldNames ? '' : 'status',
        enumValues: ShardState_Status.values)
    ..aOB(2, _omitFieldNames ? '' : 'enabled')
    ..aOB(3, _omitFieldNames ? '' : 'available')
    ..aOS(4, _omitFieldNames ? '' : 'pairingCode')
    ..aOS(5, _omitFieldNames ? '' : 'message')
    ..aInt64(6, _omitFieldNames ? '' : 'syncedMs')
    ..aOS(7, _omitFieldNames ? '' : 'deviceName')
    ..aOB(8, _omitFieldNames ? '' : 'canMerge')
    ..aOB(9, _omitFieldNames ? '' : 'unpaired')
    ..aOB(10, _omitFieldNames ? '' : 'canOpen')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ShardState clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ShardState copyWith(void Function(ShardState) updates) =>
      super.copyWith((message) => updates(message as ShardState)) as ShardState;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ShardState() / ShardState.new instead')
  static ShardState create() => ShardState._();
  static $pb.GeneratedMessage $_createMessage() => ShardState._();
  @$core.override
  ShardState createEmptyInstance() => ShardState._();
  @$core.pragma('dart2js:noInline')
  static ShardState getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ShardState>(ShardState.$_createMessage);
  static ShardState? _defaultInstance;

  @$pb.TagNumber(1)
  ShardState_Status get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(ShardState_Status value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get enabled => $_getBF(1);
  @$pb.TagNumber(2)
  set enabled($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEnabled() => $_has(1);
  @$pb.TagNumber(2)
  void clearEnabled() => $_clearField(2);

  /// Shard answers here (status is not OFF/NOT_RUNNING), so the switch
  /// can be shown; with no_vault false it can be turned on.
  @$pb.TagNumber(3)
  $core.bool get available => $_getBF(2);
  @$pb.TagNumber(3)
  set available($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAvailable() => $_has(2);
  @$pb.TagNumber(3)
  void clearAvailable() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get pairingCode => $_getSZ(3);
  @$pb.TagNumber(4)
  set pairingCode($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPairingCode() => $_has(3);
  @$pb.TagNumber(4)
  void clearPairingCode() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get message => $_getSZ(4);
  @$pb.TagNumber(5)
  set message($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasMessage() => $_has(4);
  @$pb.TagNumber(5)
  void clearMessage() => $_clearField(5);

  @$pb.TagNumber(6)
  $fixnum.Int64 get syncedMs => $_getI64(5);
  @$pb.TagNumber(6)
  set syncedMs($fixnum.Int64 value) => $_setInt64(5, value);
  @$pb.TagNumber(6)
  $core.bool hasSyncedMs() => $_has(5);
  @$pb.TagNumber(6)
  void clearSyncedMs() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get deviceName => $_getSZ(6);
  @$pb.TagNumber(7)
  set deviceName($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasDeviceName() => $_has(6);
  @$pb.TagNumber(7)
  void clearDeviceName() => $_clearField(7);

  /// CONFLICT: merging is possible (the same vault on both sides).
  @$pb.TagNumber(8)
  $core.bool get canMerge => $_getBF(7);
  @$pb.TagNumber(8)
  set canMerge($core.bool value) => $_setBool(7, value);
  @$pb.TagNumber(8)
  $core.bool hasCanMerge() => $_has(7);
  @$pb.TagNumber(8)
  void clearCanMerge() => $_clearField(8);

  /// The pairing was revoked or denied: sync went off, offer to pair again.
  @$pb.TagNumber(9)
  $core.bool get unpaired => $_getBF(8);
  @$pb.TagNumber(9)
  set unpaired($core.bool value) => $_setBool(8, value);
  @$pb.TagNumber(9)
  $core.bool hasUnpaired() => $_has(8);
  @$pb.TagNumber(9)
  void clearUnpaired() => $_clearField(9);

  /// The platform can open Shard (Android): offer it for LOCKED/PAIRING.
  @$pb.TagNumber(10)
  $core.bool get canOpen => $_getBF(9);
  @$pb.TagNumber(10)
  set canOpen($core.bool value) => $_setBool(9, value);
  @$pb.TagNumber(10)
  $core.bool hasCanOpen() => $_has(9);
  @$pb.TagNumber(10)
  void clearCanOpen() => $_clearField(10);
}

class ResolveShardRequest extends $pb.GeneratedMessage {
  factory ResolveShardRequest({
    ResolveShardRequest_Choice? choice,
  }) {
    final result = ResolveShardRequest._();
    if (choice != null) result.choice = choice;
    return result;
  }

  ResolveShardRequest._();

  factory ResolveShardRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResolveShardRequest()..mergeFromBuffer(data, registry);
  factory ResolveShardRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResolveShardRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ResolveShardRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ResolveShardRequest.$_createMessage)
    ..aE<ResolveShardRequest_Choice>(1, _omitFieldNames ? '' : 'choice',
        enumValues: ResolveShardRequest_Choice.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResolveShardRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResolveShardRequest copyWith(void Function(ResolveShardRequest) updates) =>
      super.copyWith((message) => updates(message as ResolveShardRequest))
          as ResolveShardRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use ResolveShardRequest() / ResolveShardRequest.new instead')
  static ResolveShardRequest create() => ResolveShardRequest._();
  static $pb.GeneratedMessage $_createMessage() => ResolveShardRequest._();
  @$core.override
  ResolveShardRequest createEmptyInstance() => ResolveShardRequest._();
  @$core.pragma('dart2js:noInline')
  static ResolveShardRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ResolveShardRequest>(
          ResolveShardRequest.$_createMessage);
  static ResolveShardRequest? _defaultInstance;

  @$pb.TagNumber(1)
  ResolveShardRequest_Choice get choice => $_getN(0);
  @$pb.TagNumber(1)
  set choice(ResolveShardRequest_Choice value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasChoice() => $_has(0);
  @$pb.TagNumber(1)
  void clearChoice() => $_clearField(1);
}

class DisableShardRequest extends $pb.GeneratedMessage {
  factory DisableShardRequest({
    $core.bool? removeSynced,
  }) {
    final result = DisableShardRequest._();
    if (removeSynced != null) result.removeSynced = removeSynced;
    return result;
  }

  DisableShardRequest._();

  factory DisableShardRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DisableShardRequest()..mergeFromBuffer(data, registry);
  factory DisableShardRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DisableShardRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DisableShardRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: DisableShardRequest.$_createMessage)
    ..aOB(1, _omitFieldNames ? '' : 'removeSynced')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DisableShardRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DisableShardRequest copyWith(void Function(DisableShardRequest) updates) =>
      super.copyWith((message) => updates(message as DisableShardRequest))
          as DisableShardRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use DisableShardRequest() / DisableShardRequest.new instead')
  static DisableShardRequest create() => DisableShardRequest._();
  static $pb.GeneratedMessage $_createMessage() => DisableShardRequest._();
  @$core.override
  DisableShardRequest createEmptyInstance() => DisableShardRequest._();
  @$core.pragma('dart2js:noInline')
  static DisableShardRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DisableShardRequest>(
          DisableShardRequest.$_createMessage);
  static DisableShardRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get removeSynced => $_getBF(0);
  @$pb.TagNumber(1)
  set removeSynced($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasRemoveSynced() => $_has(0);
  @$pb.TagNumber(1)
  void clearRemoveSynced() => $_clearField(1);
}

class Field extends $pb.GeneratedMessage {
  factory Field({
    $core.String? key,
    $core.String? label,
    $core.String? help,
    $core.bool? secret,
    $core.bool? optional,
    $core.String? defaultValue,
    $core.bool? multiline,
  }) {
    final result = Field._();
    if (key != null) result.key = key;
    if (label != null) result.label = label;
    if (help != null) result.help = help;
    if (secret != null) result.secret = secret;
    if (optional != null) result.optional = optional;
    if (defaultValue != null) result.defaultValue = defaultValue;
    if (multiline != null) result.multiline = multiline;
    return result;
  }

  Field._();

  factory Field.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Field()..mergeFromBuffer(data, registry);
  factory Field.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Field()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Field',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Field.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'key')
    ..aOS(2, _omitFieldNames ? '' : 'label')
    ..aOS(3, _omitFieldNames ? '' : 'help')
    ..aOB(4, _omitFieldNames ? '' : 'secret')
    ..aOB(5, _omitFieldNames ? '' : 'optional')
    ..aOS(6, _omitFieldNames ? '' : 'defaultValue')
    ..aOB(7, _omitFieldNames ? '' : 'multiline')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Field clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Field copyWith(void Function(Field) updates) =>
      super.copyWith((message) => updates(message as Field)) as Field;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Field() / Field.new instead')
  static Field create() => Field._();
  static $pb.GeneratedMessage $_createMessage() => Field._();
  @$core.override
  Field createEmptyInstance() => Field._();
  @$core.pragma('dart2js:noInline')
  static Field getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Field>(Field.$_createMessage);
  static Field? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get key => $_getSZ(0);
  @$pb.TagNumber(1)
  set key($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasKey() => $_has(0);
  @$pb.TagNumber(1)
  void clearKey() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get label => $_getSZ(1);
  @$pb.TagNumber(2)
  set label($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasLabel() => $_has(1);
  @$pb.TagNumber(2)
  void clearLabel() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get help => $_getSZ(2);
  @$pb.TagNumber(3)
  set help($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasHelp() => $_has(2);
  @$pb.TagNumber(3)
  void clearHelp() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get secret => $_getBF(3);
  @$pb.TagNumber(4)
  set secret($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSecret() => $_has(3);
  @$pb.TagNumber(4)
  void clearSecret() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get optional => $_getBF(4);
  @$pb.TagNumber(5)
  set optional($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasOptional() => $_has(4);
  @$pb.TagNumber(5)
  void clearOptional() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get defaultValue => $_getSZ(5);
  @$pb.TagNumber(6)
  set defaultValue($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasDefaultValue() => $_has(5);
  @$pb.TagNumber(6)
  void clearDefaultValue() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.bool get multiline => $_getBF(6);
  @$pb.TagNumber(7)
  set multiline($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasMultiline() => $_has(6);
  @$pb.TagNumber(7)
  void clearMultiline() => $_clearField(7);
}

class AuthMethod extends $pb.GeneratedMessage {
  factory AuthMethod({
    $core.String? id,
    $core.String? name,
    $core.String? description,
  }) {
    final result = AuthMethod._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (description != null) result.description = description;
    return result;
  }

  AuthMethod._();

  factory AuthMethod.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AuthMethod()..mergeFromBuffer(data, registry);
  factory AuthMethod.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AuthMethod()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AuthMethod',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: AuthMethod.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'description')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AuthMethod clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AuthMethod copyWith(void Function(AuthMethod) updates) =>
      super.copyWith((message) => updates(message as AuthMethod)) as AuthMethod;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AuthMethod() / AuthMethod.new instead')
  static AuthMethod create() => AuthMethod._();
  static $pb.GeneratedMessage $_createMessage() => AuthMethod._();
  @$core.override
  AuthMethod createEmptyInstance() => AuthMethod._();
  @$core.pragma('dart2js:noInline')
  static AuthMethod getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AuthMethod>(AuthMethod.$_createMessage);
  static AuthMethod? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get description => $_getSZ(2);
  @$pb.TagNumber(3)
  set description($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDescription() => $_has(2);
  @$pb.TagNumber(3)
  void clearDescription() => $_clearField(3);
}

class ProviderInfo extends $pb.GeneratedMessage {
  factory ProviderInfo({
    $core.String? type,
    $core.String? name,
    $core.Iterable<AuthMethod>? authMethods,
    $core.Iterable<Capability>? capabilities,
  }) {
    final result = ProviderInfo._();
    if (type != null) result.type = type;
    if (name != null) result.name = name;
    if (authMethods != null) result.authMethods.addAll(authMethods);
    if (capabilities != null) result.capabilities.addAll(capabilities);
    return result;
  }

  ProviderInfo._();

  factory ProviderInfo.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProviderInfo()..mergeFromBuffer(data, registry);
  factory ProviderInfo.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProviderInfo()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProviderInfo',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ProviderInfo.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'type')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..pPM<AuthMethod>(3, _omitFieldNames ? '' : 'authMethods',
        subBuilder: AuthMethod.$_createMessage)
    ..pc<Capability>(
        4, _omitFieldNames ? '' : 'capabilities', $pb.PbFieldType.KE,
        valueOf: Capability.valueOf,
        enumValues: Capability.values,
        defaultEnumValue: Capability.CAPABILITY_UNSPECIFIED)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProviderInfo clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProviderInfo copyWith(void Function(ProviderInfo) updates) =>
      super.copyWith((message) => updates(message as ProviderInfo))
          as ProviderInfo;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ProviderInfo() / ProviderInfo.new instead')
  static ProviderInfo create() => ProviderInfo._();
  static $pb.GeneratedMessage $_createMessage() => ProviderInfo._();
  @$core.override
  ProviderInfo createEmptyInstance() => ProviderInfo._();
  @$core.pragma('dart2js:noInline')
  static ProviderInfo getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ProviderInfo>(
          ProviderInfo.$_createMessage);
  static ProviderInfo? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get type => $_getSZ(0);
  @$pb.TagNumber(1)
  set type($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $pb.PbList<AuthMethod> get authMethods => $_getList(2);

  @$pb.TagNumber(4)
  $pb.PbList<Capability> get capabilities => $_getList(3);
}

class ProvidersReply extends $pb.GeneratedMessage {
  factory ProvidersReply({
    $core.Iterable<ProviderInfo>? providers,
  }) {
    final result = ProvidersReply._();
    if (providers != null) result.providers.addAll(providers);
    return result;
  }

  ProvidersReply._();

  factory ProvidersReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProvidersReply()..mergeFromBuffer(data, registry);
  factory ProvidersReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProvidersReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProvidersReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ProvidersReply.$_createMessage)
    ..pPM<ProviderInfo>(1, _omitFieldNames ? '' : 'providers',
        subBuilder: ProviderInfo.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProvidersReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProvidersReply copyWith(void Function(ProvidersReply) updates) =>
      super.copyWith((message) => updates(message as ProvidersReply))
          as ProvidersReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ProvidersReply() / ProvidersReply.new instead')
  static ProvidersReply create() => ProvidersReply._();
  static $pb.GeneratedMessage $_createMessage() => ProvidersReply._();
  @$core.override
  ProvidersReply createEmptyInstance() => ProvidersReply._();
  @$core.pragma('dart2js:noInline')
  static ProvidersReply getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ProvidersReply>(
          ProvidersReply.$_createMessage);
  static ProvidersReply? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<ProviderInfo> get providers => $_getList(0);
}

class StartLoginRequest extends $pb.GeneratedMessage {
  factory StartLoginRequest({
    $core.String? provider,
    $core.String? method,
  }) {
    final result = StartLoginRequest._();
    if (provider != null) result.provider = provider;
    if (method != null) result.method = method;
    return result;
  }

  StartLoginRequest._();

  factory StartLoginRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StartLoginRequest()..mergeFromBuffer(data, registry);
  factory StartLoginRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StartLoginRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StartLoginRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: StartLoginRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'provider')
    ..aOS(2, _omitFieldNames ? '' : 'method')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartLoginRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartLoginRequest copyWith(void Function(StartLoginRequest) updates) =>
      super.copyWith((message) => updates(message as StartLoginRequest))
          as StartLoginRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use StartLoginRequest() / StartLoginRequest.new instead')
  static StartLoginRequest create() => StartLoginRequest._();
  static $pb.GeneratedMessage $_createMessage() => StartLoginRequest._();
  @$core.override
  StartLoginRequest createEmptyInstance() => StartLoginRequest._();
  @$core.pragma('dart2js:noInline')
  static StartLoginRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<StartLoginRequest>(
          StartLoginRequest.$_createMessage);
  static StartLoginRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get provider => $_getSZ(0);
  @$pb.TagNumber(1)
  set provider($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasProvider() => $_has(0);
  @$pb.TagNumber(1)
  void clearProvider() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get method => $_getSZ(1);
  @$pb.TagNumber(2)
  set method($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMethod() => $_has(1);
  @$pb.TagNumber(2)
  void clearMethod() => $_clearField(2);
}

class LoginStepRequest extends $pb.GeneratedMessage {
  factory LoginStepRequest({
    $core.String? loginId,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? values,
  }) {
    final result = LoginStepRequest._();
    if (loginId != null) result.loginId = loginId;
    if (values != null) result.values.addEntries(values);
    return result;
  }

  LoginStepRequest._();

  factory LoginStepRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LoginStepRequest()..mergeFromBuffer(data, registry);
  factory LoginStepRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LoginStepRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LoginStepRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: LoginStepRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'loginId')
    ..m<$core.String, $core.String>(2, _omitFieldNames ? '' : 'values',
        entryClassName: 'LoginStepRequest.ValuesEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('selfhold.parley.v1'))
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginStepRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginStepRequest copyWith(void Function(LoginStepRequest) updates) =>
      super.copyWith((message) => updates(message as LoginStepRequest))
          as LoginStepRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LoginStepRequest() / LoginStepRequest.new instead')
  static LoginStepRequest create() => LoginStepRequest._();
  static $pb.GeneratedMessage $_createMessage() => LoginStepRequest._();
  @$core.override
  LoginStepRequest createEmptyInstance() => LoginStepRequest._();
  @$core.pragma('dart2js:noInline')
  static LoginStepRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LoginStepRequest>(
          LoginStepRequest.$_createMessage);
  static LoginStepRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get loginId => $_getSZ(0);
  @$pb.TagNumber(1)
  set loginId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasLoginId() => $_has(0);
  @$pb.TagNumber(1)
  void clearLoginId() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbMap<$core.String, $core.String> get values => $_getMap(1);
}

class LoginStepReply extends $pb.GeneratedMessage {
  factory LoginStepReply({
    $core.String? loginId,
    $core.String? title,
    $core.String? message,
    $core.String? error,
    $core.Iterable<Field>? fields,
    $core.String? openUrl,
    $core.bool? done,
    Account? account,
    $core.String? qr,
  }) {
    final result = LoginStepReply._();
    if (loginId != null) result.loginId = loginId;
    if (title != null) result.title = title;
    if (message != null) result.message = message;
    if (error != null) result.error = error;
    if (fields != null) result.fields.addAll(fields);
    if (openUrl != null) result.openUrl = openUrl;
    if (done != null) result.done = done;
    if (account != null) result.account = account;
    if (qr != null) result.qr = qr;
    return result;
  }

  LoginStepReply._();

  factory LoginStepReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LoginStepReply()..mergeFromBuffer(data, registry);
  factory LoginStepReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LoginStepReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LoginStepReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: LoginStepReply.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'loginId')
    ..aOS(2, _omitFieldNames ? '' : 'title')
    ..aOS(3, _omitFieldNames ? '' : 'message')
    ..aOS(4, _omitFieldNames ? '' : 'error')
    ..pPM<Field>(5, _omitFieldNames ? '' : 'fields',
        subBuilder: Field.$_createMessage)
    ..aOS(6, _omitFieldNames ? '' : 'openUrl')
    ..aOB(7, _omitFieldNames ? '' : 'done')
    ..aOM<Account>(8, _omitFieldNames ? '' : 'account',
        subBuilder: Account.$_createMessage)
    ..aOS(9, _omitFieldNames ? '' : 'qr')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginStepReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginStepReply copyWith(void Function(LoginStepReply) updates) =>
      super.copyWith((message) => updates(message as LoginStepReply))
          as LoginStepReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LoginStepReply() / LoginStepReply.new instead')
  static LoginStepReply create() => LoginStepReply._();
  static $pb.GeneratedMessage $_createMessage() => LoginStepReply._();
  @$core.override
  LoginStepReply createEmptyInstance() => LoginStepReply._();
  @$core.pragma('dart2js:noInline')
  static LoginStepReply getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LoginStepReply>(
          LoginStepReply.$_createMessage);
  static LoginStepReply? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get loginId => $_getSZ(0);
  @$pb.TagNumber(1)
  set loginId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasLoginId() => $_has(0);
  @$pb.TagNumber(1)
  void clearLoginId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get title => $_getSZ(1);
  @$pb.TagNumber(2)
  set title($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTitle() => $_has(1);
  @$pb.TagNumber(2)
  void clearTitle() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get message => $_getSZ(2);
  @$pb.TagNumber(3)
  set message($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasMessage() => $_has(2);
  @$pb.TagNumber(3)
  void clearMessage() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get error => $_getSZ(3);
  @$pb.TagNumber(4)
  set error($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasError() => $_has(3);
  @$pb.TagNumber(4)
  void clearError() => $_clearField(4);

  @$pb.TagNumber(5)
  $pb.PbList<Field> get fields => $_getList(4);

  /// Open this URL in a browser on this device (OAuth); the step may finish
  /// by itself (wait with LoginStep and no values) or ask for the result.
  @$pb.TagNumber(6)
  $core.String get openUrl => $_getSZ(5);
  @$pb.TagNumber(6)
  set openUrl($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasOpenUrl() => $_has(5);
  @$pb.TagNumber(6)
  void clearOpenUrl() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.bool get done => $_getBF(6);
  @$pb.TagNumber(7)
  set done($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasDone() => $_has(6);
  @$pb.TagNumber(7)
  void clearDone() => $_clearField(7);

  @$pb.TagNumber(8)
  Account get account => $_getN(7);
  @$pb.TagNumber(8)
  set account(Account value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasAccount() => $_has(7);
  @$pb.TagNumber(8)
  void clearAccount() => $_clearField(8);
  @$pb.TagNumber(8)
  Account ensureAccount() => $_ensure(7);

  /// Show this text as a QR code to scan with another device (Telegram:
  /// Settings → Devices → Link Desktop Device). Wait like for open_url; the
  /// code changes every ~30 s, each reply carries the current one.
  @$pb.TagNumber(9)
  $core.String get qr => $_getSZ(8);
  @$pb.TagNumber(9)
  set qr($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasQr() => $_has(8);
  @$pb.TagNumber(9)
  void clearQr() => $_clearField(9);
}

class Account extends $pb.GeneratedMessage {
  factory Account({
    $core.String? id,
    $core.String? provider,
    $core.String? name,
    $core.String? login,
    AccountStatus? status,
    $core.String? statusMessage,
    $core.bool? enabled,
    User? self,
    $core.int? unread,
    $core.Iterable<Capability>? capabilities,
    $core.int? mailUnread,
    $core.Iterable<ChatAction>? chatActions,
  }) {
    final result = Account._();
    if (id != null) result.id = id;
    if (provider != null) result.provider = provider;
    if (name != null) result.name = name;
    if (login != null) result.login = login;
    if (status != null) result.status = status;
    if (statusMessage != null) result.statusMessage = statusMessage;
    if (enabled != null) result.enabled = enabled;
    if (self != null) result.self = self;
    if (unread != null) result.unread = unread;
    if (capabilities != null) result.capabilities.addAll(capabilities);
    if (mailUnread != null) result.mailUnread = mailUnread;
    if (chatActions != null) result.chatActions.addAll(chatActions);
    return result;
  }

  Account._();

  factory Account.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Account()..mergeFromBuffer(data, registry);
  factory Account.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Account()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Account',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Account.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'provider')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOS(4, _omitFieldNames ? '' : 'login')
    ..aE<AccountStatus>(5, _omitFieldNames ? '' : 'status',
        enumValues: AccountStatus.values)
    ..aOS(6, _omitFieldNames ? '' : 'statusMessage')
    ..aOB(7, _omitFieldNames ? '' : 'enabled')
    ..aOM<User>(8, _omitFieldNames ? '' : 'self',
        subBuilder: User.$_createMessage)
    ..aI(9, _omitFieldNames ? '' : 'unread')
    ..pc<Capability>(
        10, _omitFieldNames ? '' : 'capabilities', $pb.PbFieldType.KE,
        valueOf: Capability.valueOf,
        enumValues: Capability.values,
        defaultEnumValue: Capability.CAPABILITY_UNSPECIFIED)
    ..aI(11, _omitFieldNames ? '' : 'mailUnread')
    ..pc<ChatAction>(
        12, _omitFieldNames ? '' : 'chatActions', $pb.PbFieldType.KE,
        valueOf: ChatAction.valueOf,
        enumValues: ChatAction.values,
        defaultEnumValue: ChatAction.CHAT_ACTION_UNSPECIFIED)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Account clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Account copyWith(void Function(Account) updates) =>
      super.copyWith((message) => updates(message as Account)) as Account;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Account() / Account.new instead')
  static Account create() => Account._();
  static $pb.GeneratedMessage $_createMessage() => Account._();
  @$core.override
  Account createEmptyInstance() => Account._();
  @$core.pragma('dart2js:noInline')
  static Account getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Account>(Account.$_createMessage);
  static Account? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get provider => $_getSZ(1);
  @$pb.TagNumber(2)
  set provider($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasProvider() => $_has(1);
  @$pb.TagNumber(2)
  void clearProvider() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get name => $_getSZ(2);
  @$pb.TagNumber(3)
  set name($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get login => $_getSZ(3);
  @$pb.TagNumber(4)
  set login($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasLogin() => $_has(3);
  @$pb.TagNumber(4)
  void clearLogin() => $_clearField(4);

  @$pb.TagNumber(5)
  AccountStatus get status => $_getN(4);
  @$pb.TagNumber(5)
  set status(AccountStatus value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasStatus() => $_has(4);
  @$pb.TagNumber(5)
  void clearStatus() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get statusMessage => $_getSZ(5);
  @$pb.TagNumber(6)
  set statusMessage($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasStatusMessage() => $_has(5);
  @$pb.TagNumber(6)
  void clearStatusMessage() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.bool get enabled => $_getBF(6);
  @$pb.TagNumber(7)
  set enabled($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasEnabled() => $_has(6);
  @$pb.TagNumber(7)
  void clearEnabled() => $_clearField(7);

  @$pb.TagNumber(8)
  User get self => $_getN(7);
  @$pb.TagNumber(8)
  set self(User value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasSelf() => $_has(7);
  @$pb.TagNumber(8)
  void clearSelf() => $_clearField(8);
  @$pb.TagNumber(8)
  User ensureSelf() => $_ensure(7);

  @$pb.TagNumber(9)
  $core.int get unread => $_getIZ(8);
  @$pb.TagNumber(9)
  set unread($core.int value) => $_setSignedInt32(8, value);
  @$pb.TagNumber(9)
  $core.bool hasUnread() => $_has(8);
  @$pb.TagNumber(9)
  void clearUnread() => $_clearField(9);

  @$pb.TagNumber(10)
  $pb.PbList<Capability> get capabilities => $_getList(9);

  @$pb.TagNumber(11)
  $core.int get mailUnread => $_getIZ(10);
  @$pb.TagNumber(11)
  set mailUnread($core.int value) => $_setSignedInt32(10, value);
  @$pb.TagNumber(11)
  $core.bool hasMailUnread() => $_has(10);
  @$pb.TagNumber(11)
  void clearMailUnread() => $_clearField(11);

  @$pb.TagNumber(12)
  $pb.PbList<ChatAction> get chatActions => $_getList(11);
}

class AccountsReply extends $pb.GeneratedMessage {
  factory AccountsReply({
    $core.Iterable<Account>? accounts,
  }) {
    final result = AccountsReply._();
    if (accounts != null) result.accounts.addAll(accounts);
    return result;
  }

  AccountsReply._();

  factory AccountsReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AccountsReply()..mergeFromBuffer(data, registry);
  factory AccountsReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AccountsReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AccountsReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: AccountsReply.$_createMessage)
    ..pPM<Account>(1, _omitFieldNames ? '' : 'accounts',
        subBuilder: Account.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AccountsReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AccountsReply copyWith(void Function(AccountsReply) updates) =>
      super.copyWith((message) => updates(message as AccountsReply))
          as AccountsReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AccountsReply() / AccountsReply.new instead')
  static AccountsReply create() => AccountsReply._();
  static $pb.GeneratedMessage $_createMessage() => AccountsReply._();
  @$core.override
  AccountsReply createEmptyInstance() => AccountsReply._();
  @$core.pragma('dart2js:noInline')
  static AccountsReply getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountsReply>(
          AccountsReply.$_createMessage);
  static AccountsReply? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Account> get accounts => $_getList(0);
}

class ReorderAccountsRequest extends $pb.GeneratedMessage {
  factory ReorderAccountsRequest({
    $core.Iterable<$core.String>? ids,
  }) {
    final result = ReorderAccountsRequest._();
    if (ids != null) result.ids.addAll(ids);
    return result;
  }

  ReorderAccountsRequest._();

  factory ReorderAccountsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReorderAccountsRequest()..mergeFromBuffer(data, registry);
  factory ReorderAccountsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReorderAccountsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReorderAccountsRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ReorderAccountsRequest.$_createMessage)
    ..pPS(1, _omitFieldNames ? '' : 'ids')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReorderAccountsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReorderAccountsRequest copyWith(
          void Function(ReorderAccountsRequest) updates) =>
      super.copyWith((message) => updates(message as ReorderAccountsRequest))
          as ReorderAccountsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ReorderAccountsRequest() / ReorderAccountsRequest.new instead')
  static ReorderAccountsRequest create() => ReorderAccountsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ReorderAccountsRequest._();
  @$core.override
  ReorderAccountsRequest createEmptyInstance() => ReorderAccountsRequest._();
  @$core.pragma('dart2js:noInline')
  static ReorderAccountsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ReorderAccountsRequest>(
          ReorderAccountsRequest.$_createMessage);
  static ReorderAccountsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$core.String> get ids => $_getList(0);
}

class RenameRequest extends $pb.GeneratedMessage {
  factory RenameRequest({
    $core.String? id,
    $core.String? name,
  }) {
    final result = RenameRequest._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    return result;
  }

  RenameRequest._();

  factory RenameRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RenameRequest()..mergeFromBuffer(data, registry);
  factory RenameRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RenameRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RenameRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: RenameRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RenameRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RenameRequest copyWith(void Function(RenameRequest) updates) =>
      super.copyWith((message) => updates(message as RenameRequest))
          as RenameRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use RenameRequest() / RenameRequest.new instead')
  static RenameRequest create() => RenameRequest._();
  static $pb.GeneratedMessage $_createMessage() => RenameRequest._();
  @$core.override
  RenameRequest createEmptyInstance() => RenameRequest._();
  @$core.pragma('dart2js:noInline')
  static RenameRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RenameRequest>(
          RenameRequest.$_createMessage);
  static RenameRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);
}

class SetEnabledRequest extends $pb.GeneratedMessage {
  factory SetEnabledRequest({
    $core.String? id,
    $core.bool? enabled,
  }) {
    final result = SetEnabledRequest._();
    if (id != null) result.id = id;
    if (enabled != null) result.enabled = enabled;
    return result;
  }

  SetEnabledRequest._();

  factory SetEnabledRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetEnabledRequest()..mergeFromBuffer(data, registry);
  factory SetEnabledRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetEnabledRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetEnabledRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: SetEnabledRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOB(2, _omitFieldNames ? '' : 'enabled')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetEnabledRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetEnabledRequest copyWith(void Function(SetEnabledRequest) updates) =>
      super.copyWith((message) => updates(message as SetEnabledRequest))
          as SetEnabledRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use SetEnabledRequest() / SetEnabledRequest.new instead')
  static SetEnabledRequest create() => SetEnabledRequest._();
  static $pb.GeneratedMessage $_createMessage() => SetEnabledRequest._();
  @$core.override
  SetEnabledRequest createEmptyInstance() => SetEnabledRequest._();
  @$core.pragma('dart2js:noInline')
  static SetEnabledRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SetEnabledRequest>(
          SetEnabledRequest.$_createMessage);
  static SetEnabledRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get enabled => $_getBF(1);
  @$pb.TagNumber(2)
  set enabled($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEnabled() => $_has(1);
  @$pb.TagNumber(2)
  void clearEnabled() => $_clearField(2);
}

class User extends $pb.GeneratedMessage {
  factory User({
    $core.String? id,
    $core.String? name,
    $core.String? username,
    $core.String? avatarUrl,
    $core.bool? online,
  }) {
    final result = User._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (username != null) result.username = username;
    if (avatarUrl != null) result.avatarUrl = avatarUrl;
    if (online != null) result.online = online;
    return result;
  }

  User._();

  factory User.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      User()..mergeFromBuffer(data, registry);
  factory User.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      User()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'User',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: User.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'username')
    ..aOS(4, _omitFieldNames ? '' : 'avatarUrl')
    ..aOB(5, _omitFieldNames ? '' : 'online')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  User clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  User copyWith(void Function(User) updates) =>
      super.copyWith((message) => updates(message as User)) as User;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use User() / User.new instead')
  static User create() => User._();
  static $pb.GeneratedMessage $_createMessage() => User._();
  @$core.override
  User createEmptyInstance() => User._();
  @$core.pragma('dart2js:noInline')
  static User getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<User>(User.$_createMessage);
  static User? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get username => $_getSZ(2);
  @$pb.TagNumber(3)
  set username($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasUsername() => $_has(2);
  @$pb.TagNumber(3)
  void clearUsername() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get avatarUrl => $_getSZ(3);
  @$pb.TagNumber(4)
  set avatarUrl($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasAvatarUrl() => $_has(3);
  @$pb.TagNumber(4)
  void clearAvatarUrl() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get online => $_getBF(4);
  @$pb.TagNumber(5)
  set online($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasOnline() => $_has(4);
  @$pb.TagNumber(5)
  void clearOnline() => $_clearField(5);
}

class Chat extends $pb.GeneratedMessage {
  factory Chat({
    $core.String? accountId,
    $core.String? id,
    ChatKind? kind,
    $core.String? title,
    $core.String? avatarUrl,
    $core.int? unread,
    $core.bool? muted,
    Message? lastMessage,
    $fixnum.Int64? updatedMs,
    $core.bool? pinned,
    $core.bool? readOnly,
    $core.bool? forum,
    $core.bool? online,
    $core.bool? saved,
    $core.bool? archived,
    $core.Iterable<$core.String>? folders,
  }) {
    final result = Chat._();
    if (accountId != null) result.accountId = accountId;
    if (id != null) result.id = id;
    if (kind != null) result.kind = kind;
    if (title != null) result.title = title;
    if (avatarUrl != null) result.avatarUrl = avatarUrl;
    if (unread != null) result.unread = unread;
    if (muted != null) result.muted = muted;
    if (lastMessage != null) result.lastMessage = lastMessage;
    if (updatedMs != null) result.updatedMs = updatedMs;
    if (pinned != null) result.pinned = pinned;
    if (readOnly != null) result.readOnly = readOnly;
    if (forum != null) result.forum = forum;
    if (online != null) result.online = online;
    if (saved != null) result.saved = saved;
    if (archived != null) result.archived = archived;
    if (folders != null) result.folders.addAll(folders);
    return result;
  }

  Chat._();

  factory Chat.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Chat()..mergeFromBuffer(data, registry);
  factory Chat.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Chat()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Chat',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Chat.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'id')
    ..aE<ChatKind>(3, _omitFieldNames ? '' : 'kind',
        enumValues: ChatKind.values)
    ..aOS(4, _omitFieldNames ? '' : 'title')
    ..aOS(5, _omitFieldNames ? '' : 'avatarUrl')
    ..aI(6, _omitFieldNames ? '' : 'unread')
    ..aOB(7, _omitFieldNames ? '' : 'muted')
    ..aOM<Message>(8, _omitFieldNames ? '' : 'lastMessage',
        subBuilder: Message.$_createMessage)
    ..aInt64(9, _omitFieldNames ? '' : 'updatedMs')
    ..aOB(10, _omitFieldNames ? '' : 'pinned')
    ..aOB(11, _omitFieldNames ? '' : 'readOnly')
    ..aOB(12, _omitFieldNames ? '' : 'forum')
    ..aOB(13, _omitFieldNames ? '' : 'online')
    ..aOB(14, _omitFieldNames ? '' : 'saved')
    ..aOB(15, _omitFieldNames ? '' : 'archived')
    ..pPS(16, _omitFieldNames ? '' : 'folders')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Chat clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Chat copyWith(void Function(Chat) updates) =>
      super.copyWith((message) => updates(message as Chat)) as Chat;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Chat() / Chat.new instead')
  static Chat create() => Chat._();
  static $pb.GeneratedMessage $_createMessage() => Chat._();
  @$core.override
  Chat createEmptyInstance() => Chat._();
  @$core.pragma('dart2js:noInline')
  static Chat getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Chat>(Chat.$_createMessage);
  static Chat? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get id => $_getSZ(1);
  @$pb.TagNumber(2)
  set id($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasId() => $_has(1);
  @$pb.TagNumber(2)
  void clearId() => $_clearField(2);

  @$pb.TagNumber(3)
  ChatKind get kind => $_getN(2);
  @$pb.TagNumber(3)
  set kind(ChatKind value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasKind() => $_has(2);
  @$pb.TagNumber(3)
  void clearKind() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get title => $_getSZ(3);
  @$pb.TagNumber(4)
  set title($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasTitle() => $_has(3);
  @$pb.TagNumber(4)
  void clearTitle() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get avatarUrl => $_getSZ(4);
  @$pb.TagNumber(5)
  set avatarUrl($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAvatarUrl() => $_has(4);
  @$pb.TagNumber(5)
  void clearAvatarUrl() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.int get unread => $_getIZ(5);
  @$pb.TagNumber(6)
  set unread($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasUnread() => $_has(5);
  @$pb.TagNumber(6)
  void clearUnread() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.bool get muted => $_getBF(6);
  @$pb.TagNumber(7)
  set muted($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasMuted() => $_has(6);
  @$pb.TagNumber(7)
  void clearMuted() => $_clearField(7);

  @$pb.TagNumber(8)
  Message get lastMessage => $_getN(7);
  @$pb.TagNumber(8)
  set lastMessage(Message value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasLastMessage() => $_has(7);
  @$pb.TagNumber(8)
  void clearLastMessage() => $_clearField(8);
  @$pb.TagNumber(8)
  Message ensureLastMessage() => $_ensure(7);

  @$pb.TagNumber(9)
  $fixnum.Int64 get updatedMs => $_getI64(8);
  @$pb.TagNumber(9)
  set updatedMs($fixnum.Int64 value) => $_setInt64(8, value);
  @$pb.TagNumber(9)
  $core.bool hasUpdatedMs() => $_has(8);
  @$pb.TagNumber(9)
  void clearUpdatedMs() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.bool get pinned => $_getBF(9);
  @$pb.TagNumber(10)
  set pinned($core.bool value) => $_setBool(9, value);
  @$pb.TagNumber(10)
  $core.bool hasPinned() => $_has(9);
  @$pb.TagNumber(10)
  void clearPinned() => $_clearField(10);

  /// The account can only read this chat (someone's broadcast channel,
  /// writing banned): no composer.
  @$pb.TagNumber(11)
  $core.bool get readOnly => $_getBF(10);
  @$pb.TagNumber(11)
  set readOnly($core.bool value) => $_setBool(10, value);
  @$pb.TagNumber(11)
  $core.bool hasReadOnly() => $_has(10);
  @$pb.TagNumber(11)
  void clearReadOnly() => $_clearField(11);

  /// A forum: the chat is split into topics (Topics).
  @$pb.TagNumber(12)
  $core.bool get forum => $_getBF(11);
  @$pb.TagNumber(12)
  set forum($core.bool value) => $_setBool(11, value);
  @$pb.TagNumber(12)
  $core.bool hasForum() => $_has(11);
  @$pb.TagNumber(12)
  void clearForum() => $_clearField(12);

  /// A direct chat whose other person is online now (live, not cached;
  /// a change comes as KIND_CHAT).
  @$pb.TagNumber(13)
  $core.bool get online => $_getBF(12);
  @$pb.TagNumber(13)
  set online($core.bool value) => $_setBool(12, value);
  @$pb.TagNumber(13)
  $core.bool hasOnline() => $_has(12);
  @$pb.TagNumber(13)
  void clearOnline() => $_clearField(13);

  /// The account's own notes (Telegram's Saved Messages, the chat with
  /// oneself): the app names it.
  @$pb.TagNumber(14)
  $core.bool get saved => $_getBF(13);
  @$pb.TagNumber(14)
  set saved($core.bool value) => $_setBool(13, value);
  @$pb.TagNumber(14)
  $core.bool hasSaved() => $_has(13);
  @$pb.TagNumber(14)
  void clearSaved() => $_clearField(14);

  /// In the messenger's archive: listed apart from the other chats.
  @$pb.TagNumber(15)
  $core.bool get archived => $_getBF(14);
  @$pb.TagNumber(15)
  set archived($core.bool value) => $_setBool(14, value);
  @$pb.TagNumber(15)
  $core.bool hasArchived() => $_has(14);
  @$pb.TagNumber(15)
  void clearArchived() => $_clearField(15);

  /// Ids of the account's chat folders the chat is in by their rules
  /// (see ChatFolder).
  @$pb.TagNumber(16)
  $pb.PbList<$core.String> get folders => $_getList(15);
}

class ManageChatRequest extends $pb.GeneratedMessage {
  factory ManageChatRequest({
    $core.String? accountId,
    $core.String? chatId,
    ChatAction? action,
  }) {
    final result = ManageChatRequest._();
    if (accountId != null) result.accountId = accountId;
    if (chatId != null) result.chatId = chatId;
    if (action != null) result.action = action;
    return result;
  }

  ManageChatRequest._();

  factory ManageChatRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ManageChatRequest()..mergeFromBuffer(data, registry);
  factory ManageChatRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ManageChatRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ManageChatRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ManageChatRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'chatId')
    ..aE<ChatAction>(3, _omitFieldNames ? '' : 'action',
        enumValues: ChatAction.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ManageChatRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ManageChatRequest copyWith(void Function(ManageChatRequest) updates) =>
      super.copyWith((message) => updates(message as ManageChatRequest))
          as ManageChatRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ManageChatRequest() / ManageChatRequest.new instead')
  static ManageChatRequest create() => ManageChatRequest._();
  static $pb.GeneratedMessage $_createMessage() => ManageChatRequest._();
  @$core.override
  ManageChatRequest createEmptyInstance() => ManageChatRequest._();
  @$core.pragma('dart2js:noInline')
  static ManageChatRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ManageChatRequest>(
          ManageChatRequest.$_createMessage);
  static ManageChatRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get chatId => $_getSZ(1);
  @$pb.TagNumber(2)
  set chatId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChatId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChatId() => $_clearField(2);

  @$pb.TagNumber(3)
  ChatAction get action => $_getN(2);
  @$pb.TagNumber(3)
  set action(ChatAction value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasAction() => $_has(2);
  @$pb.TagNumber(3)
  void clearAction() => $_clearField(3);
}

/// A folder of an account's chat list. A chat of the account is in it when
/// its id is in always, or Chat.folders names the folder and no exclude_*
/// applies to it.
class ChatFolder extends $pb.GeneratedMessage {
  factory ChatFolder({
    $core.String? accountId,
    $core.String? id,
    $core.String? title,
    $core.String? emoji,
    $core.Iterable<$core.String>? always,
    $core.Iterable<$core.String>? pinned,
    $core.bool? excludeMuted,
    $core.bool? excludeRead,
    $core.bool? excludeArchived,
  }) {
    final result = ChatFolder._();
    if (accountId != null) result.accountId = accountId;
    if (id != null) result.id = id;
    if (title != null) result.title = title;
    if (emoji != null) result.emoji = emoji;
    if (always != null) result.always.addAll(always);
    if (pinned != null) result.pinned.addAll(pinned);
    if (excludeMuted != null) result.excludeMuted = excludeMuted;
    if (excludeRead != null) result.excludeRead = excludeRead;
    if (excludeArchived != null) result.excludeArchived = excludeArchived;
    return result;
  }

  ChatFolder._();

  factory ChatFolder.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChatFolder()..mergeFromBuffer(data, registry);
  factory ChatFolder.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChatFolder()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ChatFolder',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ChatFolder.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'id')
    ..aOS(3, _omitFieldNames ? '' : 'title')
    ..aOS(4, _omitFieldNames ? '' : 'emoji')
    ..pPS(5, _omitFieldNames ? '' : 'always')
    ..pPS(6, _omitFieldNames ? '' : 'pinned')
    ..aOB(7, _omitFieldNames ? '' : 'excludeMuted')
    ..aOB(8, _omitFieldNames ? '' : 'excludeRead')
    ..aOB(9, _omitFieldNames ? '' : 'excludeArchived')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChatFolder clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChatFolder copyWith(void Function(ChatFolder) updates) =>
      super.copyWith((message) => updates(message as ChatFolder)) as ChatFolder;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ChatFolder() / ChatFolder.new instead')
  static ChatFolder create() => ChatFolder._();
  static $pb.GeneratedMessage $_createMessage() => ChatFolder._();
  @$core.override
  ChatFolder createEmptyInstance() => ChatFolder._();
  @$core.pragma('dart2js:noInline')
  static ChatFolder getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ChatFolder>(ChatFolder.$_createMessage);
  static ChatFolder? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get id => $_getSZ(1);
  @$pb.TagNumber(2)
  set id($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasId() => $_has(1);
  @$pb.TagNumber(2)
  void clearId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get title => $_getSZ(2);
  @$pb.TagNumber(3)
  set title($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTitle() => $_has(2);
  @$pb.TagNumber(3)
  void clearTitle() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get emoji => $_getSZ(3);
  @$pb.TagNumber(4)
  set emoji($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasEmoji() => $_has(3);
  @$pb.TagNumber(4)
  void clearEmoji() => $_clearField(4);

  @$pb.TagNumber(5)
  $pb.PbList<$core.String> get always => $_getList(4);

  @$pb.TagNumber(6)
  $pb.PbList<$core.String> get pinned => $_getList(5);

  @$pb.TagNumber(7)
  $core.bool get excludeMuted => $_getBF(6);
  @$pb.TagNumber(7)
  set excludeMuted($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasExcludeMuted() => $_has(6);
  @$pb.TagNumber(7)
  void clearExcludeMuted() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.bool get excludeRead => $_getBF(7);
  @$pb.TagNumber(8)
  set excludeRead($core.bool value) => $_setBool(7, value);
  @$pb.TagNumber(8)
  $core.bool hasExcludeRead() => $_has(7);
  @$pb.TagNumber(8)
  void clearExcludeRead() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.bool get excludeArchived => $_getBF(8);
  @$pb.TagNumber(9)
  set excludeArchived($core.bool value) => $_setBool(8, value);
  @$pb.TagNumber(9)
  $core.bool hasExcludeArchived() => $_has(8);
  @$pb.TagNumber(9)
  void clearExcludeArchived() => $_clearField(9);
}

class ChatFoldersReply extends $pb.GeneratedMessage {
  factory ChatFoldersReply({
    $core.Iterable<ChatFolder>? folders,
    $core.Iterable<AccountError>? failed,
  }) {
    final result = ChatFoldersReply._();
    if (folders != null) result.folders.addAll(folders);
    if (failed != null) result.failed.addAll(failed);
    return result;
  }

  ChatFoldersReply._();

  factory ChatFoldersReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChatFoldersReply()..mergeFromBuffer(data, registry);
  factory ChatFoldersReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChatFoldersReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ChatFoldersReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ChatFoldersReply.$_createMessage)
    ..pPM<ChatFolder>(1, _omitFieldNames ? '' : 'folders',
        subBuilder: ChatFolder.$_createMessage)
    ..pPM<AccountError>(2, _omitFieldNames ? '' : 'failed',
        subBuilder: AccountError.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChatFoldersReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChatFoldersReply copyWith(void Function(ChatFoldersReply) updates) =>
      super.copyWith((message) => updates(message as ChatFoldersReply))
          as ChatFoldersReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ChatFoldersReply() / ChatFoldersReply.new instead')
  static ChatFoldersReply create() => ChatFoldersReply._();
  static $pb.GeneratedMessage $_createMessage() => ChatFoldersReply._();
  @$core.override
  ChatFoldersReply createEmptyInstance() => ChatFoldersReply._();
  @$core.pragma('dart2js:noInline')
  static ChatFoldersReply getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ChatFoldersReply>(
          ChatFoldersReply.$_createMessage);
  static ChatFoldersReply? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<ChatFolder> get folders => $_getList(0);

  @$pb.TagNumber(2)
  $pb.PbList<AccountError> get failed => $_getList(1);
}

class ChatsRequest extends $pb.GeneratedMessage {
  factory ChatsRequest({
    $core.Iterable<$core.String>? accountIds,
    $core.int? limit,
  }) {
    final result = ChatsRequest._();
    if (accountIds != null) result.accountIds.addAll(accountIds);
    if (limit != null) result.limit = limit;
    return result;
  }

  ChatsRequest._();

  factory ChatsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChatsRequest()..mergeFromBuffer(data, registry);
  factory ChatsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChatsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ChatsRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ChatsRequest.$_createMessage)
    ..pPS(1, _omitFieldNames ? '' : 'accountIds')
    ..aI(2, _omitFieldNames ? '' : 'limit')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChatsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChatsRequest copyWith(void Function(ChatsRequest) updates) =>
      super.copyWith((message) => updates(message as ChatsRequest))
          as ChatsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ChatsRequest() / ChatsRequest.new instead')
  static ChatsRequest create() => ChatsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ChatsRequest._();
  @$core.override
  ChatsRequest createEmptyInstance() => ChatsRequest._();
  @$core.pragma('dart2js:noInline')
  static ChatsRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ChatsRequest>(
          ChatsRequest.$_createMessage);
  static ChatsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$core.String> get accountIds => $_getList(0);

  @$pb.TagNumber(2)
  $core.int get limit => $_getIZ(1);
  @$pb.TagNumber(2)
  set limit($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasLimit() => $_has(1);
  @$pb.TagNumber(2)
  void clearLimit() => $_clearField(2);
}

class ChatsReply extends $pb.GeneratedMessage {
  factory ChatsReply({
    $core.Iterable<Chat>? chats,
  }) {
    final result = ChatsReply._();
    if (chats != null) result.chats.addAll(chats);
    return result;
  }

  ChatsReply._();

  factory ChatsReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChatsReply()..mergeFromBuffer(data, registry);
  factory ChatsReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChatsReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ChatsReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ChatsReply.$_createMessage)
    ..pPM<Chat>(1, _omitFieldNames ? '' : 'chats',
        subBuilder: Chat.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChatsReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChatsReply copyWith(void Function(ChatsReply) updates) =>
      super.copyWith((message) => updates(message as ChatsReply)) as ChatsReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ChatsReply() / ChatsReply.new instead')
  static ChatsReply create() => ChatsReply._();
  static $pb.GeneratedMessage $_createMessage() => ChatsReply._();
  @$core.override
  ChatsReply createEmptyInstance() => ChatsReply._();
  @$core.pragma('dart2js:noInline')
  static ChatsReply getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ChatsReply>(ChatsReply.$_createMessage);
  static ChatsReply? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Chat> get chats => $_getList(0);
}

class Attachment extends $pb.GeneratedMessage {
  factory Attachment({
    AttachmentKind? kind,
    $core.String? name,
    $core.String? mime,
    $fixnum.Int64? size,
    $core.String? ref,
    $core.String? thumb,
    $core.int? width,
    $core.int? height,
    $core.int? durationMs,
  }) {
    final result = Attachment._();
    if (kind != null) result.kind = kind;
    if (name != null) result.name = name;
    if (mime != null) result.mime = mime;
    if (size != null) result.size = size;
    if (ref != null) result.ref = ref;
    if (thumb != null) result.thumb = thumb;
    if (width != null) result.width = width;
    if (height != null) result.height = height;
    if (durationMs != null) result.durationMs = durationMs;
    return result;
  }

  Attachment._();

  factory Attachment.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Attachment()..mergeFromBuffer(data, registry);
  factory Attachment.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Attachment()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Attachment',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Attachment.$_createMessage)
    ..aE<AttachmentKind>(1, _omitFieldNames ? '' : 'kind',
        enumValues: AttachmentKind.values)
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'mime')
    ..aInt64(4, _omitFieldNames ? '' : 'size')
    ..aOS(5, _omitFieldNames ? '' : 'ref')
    ..aOS(6, _omitFieldNames ? '' : 'thumb')
    ..aI(7, _omitFieldNames ? '' : 'width')
    ..aI(8, _omitFieldNames ? '' : 'height')
    ..aI(9, _omitFieldNames ? '' : 'durationMs')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Attachment clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Attachment copyWith(void Function(Attachment) updates) =>
      super.copyWith((message) => updates(message as Attachment)) as Attachment;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Attachment() / Attachment.new instead')
  static Attachment create() => Attachment._();
  static $pb.GeneratedMessage $_createMessage() => Attachment._();
  @$core.override
  Attachment createEmptyInstance() => Attachment._();
  @$core.pragma('dart2js:noInline')
  static Attachment getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Attachment>(Attachment.$_createMessage);
  static Attachment? _defaultInstance;

  @$pb.TagNumber(1)
  AttachmentKind get kind => $_getN(0);
  @$pb.TagNumber(1)
  set kind(AttachmentKind value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasKind() => $_has(0);
  @$pb.TagNumber(1)
  void clearKind() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get mime => $_getSZ(2);
  @$pb.TagNumber(3)
  set mime($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasMime() => $_has(2);
  @$pb.TagNumber(3)
  void clearMime() => $_clearField(3);

  @$pb.TagNumber(4)
  $fixnum.Int64 get size => $_getI64(3);
  @$pb.TagNumber(4)
  set size($fixnum.Int64 value) => $_setInt64(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSize() => $_has(3);
  @$pb.TagNumber(4)
  void clearSize() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get ref => $_getSZ(4);
  @$pb.TagNumber(5)
  set ref($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasRef() => $_has(4);
  @$pb.TagNumber(5)
  void clearRef() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get thumb => $_getSZ(5);
  @$pb.TagNumber(6)
  set thumb($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasThumb() => $_has(5);
  @$pb.TagNumber(6)
  void clearThumb() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.int get width => $_getIZ(6);
  @$pb.TagNumber(7)
  set width($core.int value) => $_setSignedInt32(6, value);
  @$pb.TagNumber(7)
  $core.bool hasWidth() => $_has(6);
  @$pb.TagNumber(7)
  void clearWidth() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.int get height => $_getIZ(7);
  @$pb.TagNumber(8)
  set height($core.int value) => $_setSignedInt32(7, value);
  @$pb.TagNumber(8)
  $core.bool hasHeight() => $_has(7);
  @$pb.TagNumber(8)
  void clearHeight() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.int get durationMs => $_getIZ(8);
  @$pb.TagNumber(9)
  set durationMs($core.int value) => $_setSignedInt32(8, value);
  @$pb.TagNumber(9)
  $core.bool hasDurationMs() => $_has(8);
  @$pb.TagNumber(9)
  void clearDurationMs() => $_clearField(9);
}

class Reaction extends $pb.GeneratedMessage {
  factory Reaction({
    $core.String? emoji,
    $core.int? count,
    $core.bool? mine,
  }) {
    final result = Reaction._();
    if (emoji != null) result.emoji = emoji;
    if (count != null) result.count = count;
    if (mine != null) result.mine = mine;
    return result;
  }

  Reaction._();

  factory Reaction.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Reaction()..mergeFromBuffer(data, registry);
  factory Reaction.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Reaction()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Reaction',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Reaction.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'emoji')
    ..aI(2, _omitFieldNames ? '' : 'count')
    ..aOB(3, _omitFieldNames ? '' : 'mine')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Reaction clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Reaction copyWith(void Function(Reaction) updates) =>
      super.copyWith((message) => updates(message as Reaction)) as Reaction;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Reaction() / Reaction.new instead')
  static Reaction create() => Reaction._();
  static $pb.GeneratedMessage $_createMessage() => Reaction._();
  @$core.override
  Reaction createEmptyInstance() => Reaction._();
  @$core.pragma('dart2js:noInline')
  static Reaction getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Reaction>(Reaction.$_createMessage);
  static Reaction? _defaultInstance;

  /// The emoji, or ":name:" for a custom one the app cannot draw.
  @$pb.TagNumber(1)
  $core.String get emoji => $_getSZ(0);
  @$pb.TagNumber(1)
  set emoji($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEmoji() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmoji() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get count => $_getIZ(1);
  @$pb.TagNumber(2)
  set count($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCount() => $_has(1);
  @$pb.TagNumber(2)
  void clearCount() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get mine => $_getBF(2);
  @$pb.TagNumber(3)
  set mine($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasMine() => $_has(2);
  @$pb.TagNumber(3)
  void clearMine() => $_clearField(3);
}

class Thread extends $pb.GeneratedMessage {
  factory Thread({
    $core.String? id,
    $core.int? replies,
    $fixnum.Int64? lastMs,
    $core.Iterable<User>? users,
  }) {
    final result = Thread._();
    if (id != null) result.id = id;
    if (replies != null) result.replies = replies;
    if (lastMs != null) result.lastMs = lastMs;
    if (users != null) result.users.addAll(users);
    return result;
  }

  Thread._();

  factory Thread.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Thread()..mergeFromBuffer(data, registry);
  factory Thread.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Thread()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Thread',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Thread.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aI(2, _omitFieldNames ? '' : 'replies')
    ..aInt64(3, _omitFieldNames ? '' : 'lastMs')
    ..pPM<User>(4, _omitFieldNames ? '' : 'users',
        subBuilder: User.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Thread clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Thread copyWith(void Function(Thread) updates) =>
      super.copyWith((message) => updates(message as Thread)) as Thread;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Thread() / Thread.new instead')
  static Thread create() => Thread._();
  static $pb.GeneratedMessage $_createMessage() => Thread._();
  @$core.override
  Thread createEmptyInstance() => Thread._();
  @$core.pragma('dart2js:noInline')
  static Thread getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Thread>(Thread.$_createMessage);
  static Thread? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get replies => $_getIZ(1);
  @$pb.TagNumber(2)
  set replies($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasReplies() => $_has(1);
  @$pb.TagNumber(2)
  void clearReplies() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get lastMs => $_getI64(2);
  @$pb.TagNumber(3)
  set lastMs($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasLastMs() => $_has(2);
  @$pb.TagNumber(3)
  void clearLastMs() => $_clearField(3);

  @$pb.TagNumber(4)
  $pb.PbList<User> get users => $_getList(3);
}

class Message extends $pb.GeneratedMessage {
  factory Message({
    $core.String? accountId,
    $core.String? chatId,
    $core.String? id,
    User? sender,
    $core.String? text,
    $fixnum.Int64? timeMs,
    $fixnum.Int64? editedMs,
    $core.bool? outgoing,
    $core.String? replyToId,
    $core.Iterable<Attachment>? attachments,
    $core.bool? deleted,
    $core.bool? service,
    $core.Iterable<Reaction>? reactions,
    Thread? thread,
    $core.String? inThread,
    $core.String? group,
    $core.bool? mentioned,
    $core.bool? forwarded,
    $core.String? forwardedFrom,
    $core.String? quote,
    $core.Iterable<TextSpan>? format,
    $core.String? markdown,
  }) {
    final result = Message._();
    if (accountId != null) result.accountId = accountId;
    if (chatId != null) result.chatId = chatId;
    if (id != null) result.id = id;
    if (sender != null) result.sender = sender;
    if (text != null) result.text = text;
    if (timeMs != null) result.timeMs = timeMs;
    if (editedMs != null) result.editedMs = editedMs;
    if (outgoing != null) result.outgoing = outgoing;
    if (replyToId != null) result.replyToId = replyToId;
    if (attachments != null) result.attachments.addAll(attachments);
    if (deleted != null) result.deleted = deleted;
    if (service != null) result.service = service;
    if (reactions != null) result.reactions.addAll(reactions);
    if (thread != null) result.thread = thread;
    if (inThread != null) result.inThread = inThread;
    if (group != null) result.group = group;
    if (mentioned != null) result.mentioned = mentioned;
    if (forwarded != null) result.forwarded = forwarded;
    if (forwardedFrom != null) result.forwardedFrom = forwardedFrom;
    if (quote != null) result.quote = quote;
    if (format != null) result.format.addAll(format);
    if (markdown != null) result.markdown = markdown;
    return result;
  }

  Message._();

  factory Message.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Message()..mergeFromBuffer(data, registry);
  factory Message.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Message()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Message',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Message.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'chatId')
    ..aOS(3, _omitFieldNames ? '' : 'id')
    ..aOM<User>(4, _omitFieldNames ? '' : 'sender',
        subBuilder: User.$_createMessage)
    ..aOS(5, _omitFieldNames ? '' : 'text')
    ..aInt64(6, _omitFieldNames ? '' : 'timeMs')
    ..aInt64(7, _omitFieldNames ? '' : 'editedMs')
    ..aOB(8, _omitFieldNames ? '' : 'outgoing')
    ..aOS(9, _omitFieldNames ? '' : 'replyToId')
    ..pPM<Attachment>(10, _omitFieldNames ? '' : 'attachments',
        subBuilder: Attachment.$_createMessage)
    ..aOB(11, _omitFieldNames ? '' : 'deleted')
    ..aOB(12, _omitFieldNames ? '' : 'service')
    ..pPM<Reaction>(13, _omitFieldNames ? '' : 'reactions',
        subBuilder: Reaction.$_createMessage)
    ..aOM<Thread>(14, _omitFieldNames ? '' : 'thread',
        subBuilder: Thread.$_createMessage)
    ..aOS(15, _omitFieldNames ? '' : 'inThread')
    ..aOS(16, _omitFieldNames ? '' : 'group')
    ..aOB(17, _omitFieldNames ? '' : 'mentioned')
    ..aOB(18, _omitFieldNames ? '' : 'forwarded')
    ..aOS(19, _omitFieldNames ? '' : 'forwardedFrom')
    ..aOS(20, _omitFieldNames ? '' : 'quote')
    ..pPM<TextSpan>(21, _omitFieldNames ? '' : 'format',
        subBuilder: TextSpan.$_createMessage)
    ..aOS(22, _omitFieldNames ? '' : 'markdown')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Message clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Message copyWith(void Function(Message) updates) =>
      super.copyWith((message) => updates(message as Message)) as Message;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Message() / Message.new instead')
  static Message create() => Message._();
  static $pb.GeneratedMessage $_createMessage() => Message._();
  @$core.override
  Message createEmptyInstance() => Message._();
  @$core.pragma('dart2js:noInline')
  static Message getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Message>(Message.$_createMessage);
  static Message? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get chatId => $_getSZ(1);
  @$pb.TagNumber(2)
  set chatId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChatId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChatId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get id => $_getSZ(2);
  @$pb.TagNumber(3)
  set id($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasId() => $_has(2);
  @$pb.TagNumber(3)
  void clearId() => $_clearField(3);

  @$pb.TagNumber(4)
  User get sender => $_getN(3);
  @$pb.TagNumber(4)
  set sender(User value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasSender() => $_has(3);
  @$pb.TagNumber(4)
  void clearSender() => $_clearField(4);
  @$pb.TagNumber(4)
  User ensureSender() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.String get text => $_getSZ(4);
  @$pb.TagNumber(5)
  set text($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasText() => $_has(4);
  @$pb.TagNumber(5)
  void clearText() => $_clearField(5);

  @$pb.TagNumber(6)
  $fixnum.Int64 get timeMs => $_getI64(5);
  @$pb.TagNumber(6)
  set timeMs($fixnum.Int64 value) => $_setInt64(5, value);
  @$pb.TagNumber(6)
  $core.bool hasTimeMs() => $_has(5);
  @$pb.TagNumber(6)
  void clearTimeMs() => $_clearField(6);

  @$pb.TagNumber(7)
  $fixnum.Int64 get editedMs => $_getI64(6);
  @$pb.TagNumber(7)
  set editedMs($fixnum.Int64 value) => $_setInt64(6, value);
  @$pb.TagNumber(7)
  $core.bool hasEditedMs() => $_has(6);
  @$pb.TagNumber(7)
  void clearEditedMs() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.bool get outgoing => $_getBF(7);
  @$pb.TagNumber(8)
  set outgoing($core.bool value) => $_setBool(7, value);
  @$pb.TagNumber(8)
  $core.bool hasOutgoing() => $_has(7);
  @$pb.TagNumber(8)
  void clearOutgoing() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get replyToId => $_getSZ(8);
  @$pb.TagNumber(9)
  set replyToId($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasReplyToId() => $_has(8);
  @$pb.TagNumber(9)
  void clearReplyToId() => $_clearField(9);

  @$pb.TagNumber(10)
  $pb.PbList<Attachment> get attachments => $_getList(9);

  @$pb.TagNumber(11)
  $core.bool get deleted => $_getBF(10);
  @$pb.TagNumber(11)
  set deleted($core.bool value) => $_setBool(10, value);
  @$pb.TagNumber(11)
  $core.bool hasDeleted() => $_has(10);
  @$pb.TagNumber(11)
  void clearDeleted() => $_clearField(11);

  /// Service message (joined, title changed, call): shown centred.
  @$pb.TagNumber(12)
  $core.bool get service => $_getBF(11);
  @$pb.TagNumber(12)
  set service($core.bool value) => $_setBool(11, value);
  @$pb.TagNumber(12)
  $core.bool hasService() => $_has(11);
  @$pb.TagNumber(12)
  void clearService() => $_clearField(12);

  @$pb.TagNumber(13)
  $pb.PbList<Reaction> get reactions => $_getList(12);

  @$pb.TagNumber(14)
  Thread get thread => $_getN(13);
  @$pb.TagNumber(14)
  set thread(Thread value) => $_setField(14, value);
  @$pb.TagNumber(14)
  $core.bool hasThread() => $_has(13);
  @$pb.TagNumber(14)
  void clearThread() => $_clearField(14);
  @$pb.TagNumber(14)
  Thread ensureThread() => $_ensure(13);

  @$pb.TagNumber(15)
  $core.String get inThread => $_getSZ(14);
  @$pb.TagNumber(15)
  set inThread($core.String value) => $_setString(14, value);
  @$pb.TagNumber(15)
  $core.bool hasInThread() => $_has(14);
  @$pb.TagNumber(15)
  void clearInThread() => $_clearField(15);

  /// Messages sent together as one album share it (Telegram's grouped_id);
  /// neighbours with the same group are shown as one message.
  @$pb.TagNumber(16)
  $core.String get group => $_getSZ(15);
  @$pb.TagNumber(16)
  set group($core.String value) => $_setString(15, value);
  @$pb.TagNumber(16)
  $core.bool hasGroup() => $_has(15);
  @$pb.TagNumber(16)
  void clearGroup() => $_clearField(16);

  /// The message mentions this account's user (or replies to them, in
  /// Telegram): muted chats still notify about it.
  @$pb.TagNumber(17)
  $core.bool get mentioned => $_getBF(16);
  @$pb.TagNumber(17)
  set mentioned($core.bool value) => $_setBool(16, value);
  @$pb.TagNumber(17)
  $core.bool hasMentioned() => $_has(16);
  @$pb.TagNumber(17)
  void clearMentioned() => $_clearField(17);

  /// A forwarded message: who wrote the original (a name; may be empty
  /// when the messenger hides it, then forwarded is still set).
  @$pb.TagNumber(18)
  $core.bool get forwarded => $_getBF(17);
  @$pb.TagNumber(18)
  set forwarded($core.bool value) => $_setBool(17, value);
  @$pb.TagNumber(18)
  $core.bool hasForwarded() => $_has(17);
  @$pb.TagNumber(18)
  void clearForwarded() => $_clearField(18);

  @$pb.TagNumber(19)
  $core.String get forwardedFrom => $_getSZ(18);
  @$pb.TagNumber(19)
  set forwardedFrom($core.String value) => $_setString(18, value);
  @$pb.TagNumber(19)
  $core.bool hasForwardedFrom() => $_has(18);
  @$pb.TagNumber(19)
  void clearForwardedFrom() => $_clearField(19);

  /// A reply that quotes a fragment of reply_to_id's text (Telegram).
  @$pb.TagNumber(20)
  $core.String get quote => $_getSZ(19);
  @$pb.TagNumber(20)
  set quote($core.String value) => $_setString(19, value);
  @$pb.TagNumber(20)
  $core.bool hasQuote() => $_has(19);
  @$pb.TagNumber(20)
  void clearQuote() => $_clearField(20);

  /// Styles of text, which holds no markup itself.
  @$pb.TagNumber(21)
  $pb.PbList<TextSpan> get format => $_getList(20);

  /// The text as Markdown for the composer when it differs from text (the
  /// account's own messages only).
  @$pb.TagNumber(22)
  $core.String get markdown => $_getSZ(21);
  @$pb.TagNumber(22)
  set markdown($core.String value) => $_setString(21, value);
  @$pb.TagNumber(22)
  $core.bool hasMarkdown() => $_has(21);
  @$pb.TagNumber(22)
  void clearMarkdown() => $_clearField(22);
}

/// A styled part of a text, in UTF-16 units. Spans may nest and overlap.
class TextSpan extends $pb.GeneratedMessage {
  factory TextSpan({
    TextStyle? style,
    $core.int? offset,
    $core.int? length,
    $core.String? url,
    $core.String? language,
  }) {
    final result = TextSpan._();
    if (style != null) result.style = style;
    if (offset != null) result.offset = offset;
    if (length != null) result.length = length;
    if (url != null) result.url = url;
    if (language != null) result.language = language;
    return result;
  }

  TextSpan._();

  factory TextSpan.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TextSpan()..mergeFromBuffer(data, registry);
  factory TextSpan.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TextSpan()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TextSpan',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: TextSpan.$_createMessage)
    ..aE<TextStyle>(1, _omitFieldNames ? '' : 'style',
        enumValues: TextStyle.values)
    ..aI(2, _omitFieldNames ? '' : 'offset')
    ..aI(3, _omitFieldNames ? '' : 'length')
    ..aOS(4, _omitFieldNames ? '' : 'url')
    ..aOS(5, _omitFieldNames ? '' : 'language')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TextSpan clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TextSpan copyWith(void Function(TextSpan) updates) =>
      super.copyWith((message) => updates(message as TextSpan)) as TextSpan;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use TextSpan() / TextSpan.new instead')
  static TextSpan create() => TextSpan._();
  static $pb.GeneratedMessage $_createMessage() => TextSpan._();
  @$core.override
  TextSpan createEmptyInstance() => TextSpan._();
  @$core.pragma('dart2js:noInline')
  static TextSpan getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TextSpan>(TextSpan.$_createMessage);
  static TextSpan? _defaultInstance;

  @$pb.TagNumber(1)
  TextStyle get style => $_getN(0);
  @$pb.TagNumber(1)
  set style(TextStyle value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStyle() => $_has(0);
  @$pb.TagNumber(1)
  void clearStyle() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get offset => $_getIZ(1);
  @$pb.TagNumber(2)
  set offset($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasOffset() => $_has(1);
  @$pb.TagNumber(2)
  void clearOffset() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get length => $_getIZ(2);
  @$pb.TagNumber(3)
  set length($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasLength() => $_has(2);
  @$pb.TagNumber(3)
  void clearLength() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get url => $_getSZ(3);
  @$pb.TagNumber(4)
  set url($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasUrl() => $_has(3);
  @$pb.TagNumber(4)
  void clearUrl() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get language => $_getSZ(4);
  @$pb.TagNumber(5)
  set language($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasLanguage() => $_has(4);
  @$pb.TagNumber(5)
  void clearLanguage() => $_clearField(5);
}

class ThreadRequest extends $pb.GeneratedMessage {
  factory ThreadRequest({
    $core.String? accountId,
    $core.String? chatId,
    $core.String? threadId,
    $core.String? beforeId,
    $core.int? limit,
  }) {
    final result = ThreadRequest._();
    if (accountId != null) result.accountId = accountId;
    if (chatId != null) result.chatId = chatId;
    if (threadId != null) result.threadId = threadId;
    if (beforeId != null) result.beforeId = beforeId;
    if (limit != null) result.limit = limit;
    return result;
  }

  ThreadRequest._();

  factory ThreadRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ThreadRequest()..mergeFromBuffer(data, registry);
  factory ThreadRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ThreadRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ThreadRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ThreadRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'chatId')
    ..aOS(3, _omitFieldNames ? '' : 'threadId')
    ..aOS(4, _omitFieldNames ? '' : 'beforeId')
    ..aI(5, _omitFieldNames ? '' : 'limit')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ThreadRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ThreadRequest copyWith(void Function(ThreadRequest) updates) =>
      super.copyWith((message) => updates(message as ThreadRequest))
          as ThreadRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ThreadRequest() / ThreadRequest.new instead')
  static ThreadRequest create() => ThreadRequest._();
  static $pb.GeneratedMessage $_createMessage() => ThreadRequest._();
  @$core.override
  ThreadRequest createEmptyInstance() => ThreadRequest._();
  @$core.pragma('dart2js:noInline')
  static ThreadRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ThreadRequest>(
          ThreadRequest.$_createMessage);
  static ThreadRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get chatId => $_getSZ(1);
  @$pb.TagNumber(2)
  set chatId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChatId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChatId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get threadId => $_getSZ(2);
  @$pb.TagNumber(3)
  set threadId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasThreadId() => $_has(2);
  @$pb.TagNumber(3)
  void clearThreadId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get beforeId => $_getSZ(3);
  @$pb.TagNumber(4)
  set beforeId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasBeforeId() => $_has(3);
  @$pb.TagNumber(4)
  void clearBeforeId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.int get limit => $_getIZ(4);
  @$pb.TagNumber(5)
  set limit($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasLimit() => $_has(4);
  @$pb.TagNumber(5)
  void clearLimit() => $_clearField(5);
}

class TopicsRequest extends $pb.GeneratedMessage {
  factory TopicsRequest({
    $core.String? accountId,
    $core.String? chatId,
  }) {
    final result = TopicsRequest._();
    if (accountId != null) result.accountId = accountId;
    if (chatId != null) result.chatId = chatId;
    return result;
  }

  TopicsRequest._();

  factory TopicsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TopicsRequest()..mergeFromBuffer(data, registry);
  factory TopicsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TopicsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TopicsRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: TopicsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'chatId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TopicsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TopicsRequest copyWith(void Function(TopicsRequest) updates) =>
      super.copyWith((message) => updates(message as TopicsRequest))
          as TopicsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use TopicsRequest() / TopicsRequest.new instead')
  static TopicsRequest create() => TopicsRequest._();
  static $pb.GeneratedMessage $_createMessage() => TopicsRequest._();
  @$core.override
  TopicsRequest createEmptyInstance() => TopicsRequest._();
  @$core.pragma('dart2js:noInline')
  static TopicsRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TopicsRequest>(
          TopicsRequest.$_createMessage);
  static TopicsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get chatId => $_getSZ(1);
  @$pb.TagNumber(2)
  set chatId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChatId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChatId() => $_clearField(2);
}

class Topic extends $pb.GeneratedMessage {
  factory Topic({
    $core.String? id,
    $core.String? title,
    $core.int? unread,
    $core.int? color,
    $core.bool? pinned,
    $core.bool? closed,
    $core.bool? hidden,
    Message? lastMessage,
    $fixnum.Int64? updatedMs,
  }) {
    final result = Topic._();
    if (id != null) result.id = id;
    if (title != null) result.title = title;
    if (unread != null) result.unread = unread;
    if (color != null) result.color = color;
    if (pinned != null) result.pinned = pinned;
    if (closed != null) result.closed = closed;
    if (hidden != null) result.hidden = hidden;
    if (lastMessage != null) result.lastMessage = lastMessage;
    if (updatedMs != null) result.updatedMs = updatedMs;
    return result;
  }

  Topic._();

  factory Topic.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Topic()..mergeFromBuffer(data, registry);
  factory Topic.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Topic()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Topic',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Topic.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'title')
    ..aI(3, _omitFieldNames ? '' : 'unread')
    ..aI(4, _omitFieldNames ? '' : 'color')
    ..aOB(5, _omitFieldNames ? '' : 'pinned')
    ..aOB(6, _omitFieldNames ? '' : 'closed')
    ..aOB(7, _omitFieldNames ? '' : 'hidden')
    ..aOM<Message>(8, _omitFieldNames ? '' : 'lastMessage',
        subBuilder: Message.$_createMessage)
    ..aInt64(9, _omitFieldNames ? '' : 'updatedMs')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Topic clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Topic copyWith(void Function(Topic) updates) =>
      super.copyWith((message) => updates(message as Topic)) as Topic;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Topic() / Topic.new instead')
  static Topic create() => Topic._();
  static $pb.GeneratedMessage $_createMessage() => Topic._();
  @$core.override
  Topic createEmptyInstance() => Topic._();
  @$core.pragma('dart2js:noInline')
  static Topic getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Topic>(Topic.$_createMessage);
  static Topic? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get title => $_getSZ(1);
  @$pb.TagNumber(2)
  set title($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTitle() => $_has(1);
  @$pb.TagNumber(2)
  void clearTitle() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get unread => $_getIZ(2);
  @$pb.TagNumber(3)
  set unread($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasUnread() => $_has(2);
  @$pb.TagNumber(3)
  void clearUnread() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get color => $_getIZ(3);
  @$pb.TagNumber(4)
  set color($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasColor() => $_has(3);
  @$pb.TagNumber(4)
  void clearColor() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get pinned => $_getBF(4);
  @$pb.TagNumber(5)
  set pinned($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasPinned() => $_has(4);
  @$pb.TagNumber(5)
  void clearPinned() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.bool get closed => $_getBF(5);
  @$pb.TagNumber(6)
  set closed($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasClosed() => $_has(5);
  @$pb.TagNumber(6)
  void clearClosed() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.bool get hidden => $_getBF(6);
  @$pb.TagNumber(7)
  set hidden($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasHidden() => $_has(6);
  @$pb.TagNumber(7)
  void clearHidden() => $_clearField(7);

  @$pb.TagNumber(8)
  Message get lastMessage => $_getN(7);
  @$pb.TagNumber(8)
  set lastMessage(Message value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasLastMessage() => $_has(7);
  @$pb.TagNumber(8)
  void clearLastMessage() => $_clearField(8);
  @$pb.TagNumber(8)
  Message ensureLastMessage() => $_ensure(7);

  @$pb.TagNumber(9)
  $fixnum.Int64 get updatedMs => $_getI64(8);
  @$pb.TagNumber(9)
  set updatedMs($fixnum.Int64 value) => $_setInt64(8, value);
  @$pb.TagNumber(9)
  $core.bool hasUpdatedMs() => $_has(8);
  @$pb.TagNumber(9)
  void clearUpdatedMs() => $_clearField(9);
}

class TopicsReply extends $pb.GeneratedMessage {
  factory TopicsReply({
    $core.Iterable<Topic>? topics,
  }) {
    final result = TopicsReply._();
    if (topics != null) result.topics.addAll(topics);
    return result;
  }

  TopicsReply._();

  factory TopicsReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TopicsReply()..mergeFromBuffer(data, registry);
  factory TopicsReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TopicsReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TopicsReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: TopicsReply.$_createMessage)
    ..pPM<Topic>(1, _omitFieldNames ? '' : 'topics',
        subBuilder: Topic.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TopicsReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TopicsReply copyWith(void Function(TopicsReply) updates) =>
      super.copyWith((message) => updates(message as TopicsReply))
          as TopicsReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use TopicsReply() / TopicsReply.new instead')
  static TopicsReply create() => TopicsReply._();
  static $pb.GeneratedMessage $_createMessage() => TopicsReply._();
  @$core.override
  TopicsReply createEmptyInstance() => TopicsReply._();
  @$core.pragma('dart2js:noInline')
  static TopicsReply getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TopicsReply>(
          TopicsReply.$_createMessage);
  static TopicsReply? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Topic> get topics => $_getList(0);
}

class ReactRequest extends $pb.GeneratedMessage {
  factory ReactRequest({
    $core.String? accountId,
    $core.String? chatId,
    $core.String? messageId,
    $core.String? emoji,
    $core.bool? on,
  }) {
    final result = ReactRequest._();
    if (accountId != null) result.accountId = accountId;
    if (chatId != null) result.chatId = chatId;
    if (messageId != null) result.messageId = messageId;
    if (emoji != null) result.emoji = emoji;
    if (on != null) result.on = on;
    return result;
  }

  ReactRequest._();

  factory ReactRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReactRequest()..mergeFromBuffer(data, registry);
  factory ReactRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReactRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReactRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ReactRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'chatId')
    ..aOS(3, _omitFieldNames ? '' : 'messageId')
    ..aOS(4, _omitFieldNames ? '' : 'emoji')
    ..aOB(5, _omitFieldNames ? '' : 'on')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReactRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReactRequest copyWith(void Function(ReactRequest) updates) =>
      super.copyWith((message) => updates(message as ReactRequest))
          as ReactRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ReactRequest() / ReactRequest.new instead')
  static ReactRequest create() => ReactRequest._();
  static $pb.GeneratedMessage $_createMessage() => ReactRequest._();
  @$core.override
  ReactRequest createEmptyInstance() => ReactRequest._();
  @$core.pragma('dart2js:noInline')
  static ReactRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ReactRequest>(
          ReactRequest.$_createMessage);
  static ReactRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get chatId => $_getSZ(1);
  @$pb.TagNumber(2)
  set chatId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChatId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChatId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get messageId => $_getSZ(2);
  @$pb.TagNumber(3)
  set messageId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasMessageId() => $_has(2);
  @$pb.TagNumber(3)
  void clearMessageId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get emoji => $_getSZ(3);
  @$pb.TagNumber(4)
  set emoji($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasEmoji() => $_has(3);
  @$pb.TagNumber(4)
  void clearEmoji() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get on => $_getBF(4);
  @$pb.TagNumber(5)
  set on($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasOn() => $_has(4);
  @$pb.TagNumber(5)
  void clearOn() => $_clearField(5);
}

class EditRequest extends $pb.GeneratedMessage {
  factory EditRequest({
    $core.String? accountId,
    $core.String? chatId,
    $core.String? messageId,
    $core.String? text,
    $core.String? threadId,
  }) {
    final result = EditRequest._();
    if (accountId != null) result.accountId = accountId;
    if (chatId != null) result.chatId = chatId;
    if (messageId != null) result.messageId = messageId;
    if (text != null) result.text = text;
    if (threadId != null) result.threadId = threadId;
    return result;
  }

  EditRequest._();

  factory EditRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      EditRequest()..mergeFromBuffer(data, registry);
  factory EditRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      EditRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EditRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: EditRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'chatId')
    ..aOS(3, _omitFieldNames ? '' : 'messageId')
    ..aOS(4, _omitFieldNames ? '' : 'text')
    ..aOS(5, _omitFieldNames ? '' : 'threadId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EditRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EditRequest copyWith(void Function(EditRequest) updates) =>
      super.copyWith((message) => updates(message as EditRequest))
          as EditRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use EditRequest() / EditRequest.new instead')
  static EditRequest create() => EditRequest._();
  static $pb.GeneratedMessage $_createMessage() => EditRequest._();
  @$core.override
  EditRequest createEmptyInstance() => EditRequest._();
  @$core.pragma('dart2js:noInline')
  static EditRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<EditRequest>(
          EditRequest.$_createMessage);
  static EditRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get chatId => $_getSZ(1);
  @$pb.TagNumber(2)
  set chatId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChatId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChatId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get messageId => $_getSZ(2);
  @$pb.TagNumber(3)
  set messageId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasMessageId() => $_has(2);
  @$pb.TagNumber(3)
  void clearMessageId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get text => $_getSZ(3);
  @$pb.TagNumber(4)
  set text($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasText() => $_has(3);
  @$pb.TagNumber(4)
  void clearText() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get threadId => $_getSZ(4);
  @$pb.TagNumber(5)
  set threadId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasThreadId() => $_has(4);
  @$pb.TagNumber(5)
  void clearThreadId() => $_clearField(5);
}

class DeleteRequest extends $pb.GeneratedMessage {
  factory DeleteRequest({
    $core.String? accountId,
    $core.String? chatId,
    $core.Iterable<$core.String>? messageIds,
    $core.bool? forAll,
    $core.String? threadId,
  }) {
    final result = DeleteRequest._();
    if (accountId != null) result.accountId = accountId;
    if (chatId != null) result.chatId = chatId;
    if (messageIds != null) result.messageIds.addAll(messageIds);
    if (forAll != null) result.forAll = forAll;
    if (threadId != null) result.threadId = threadId;
    return result;
  }

  DeleteRequest._();

  factory DeleteRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteRequest()..mergeFromBuffer(data, registry);
  factory DeleteRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: DeleteRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'chatId')
    ..pPS(3, _omitFieldNames ? '' : 'messageIds')
    ..aOB(4, _omitFieldNames ? '' : 'forAll')
    ..aOS(5, _omitFieldNames ? '' : 'threadId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteRequest copyWith(void Function(DeleteRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteRequest))
          as DeleteRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use DeleteRequest() / DeleteRequest.new instead')
  static DeleteRequest create() => DeleteRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteRequest._();
  @$core.override
  DeleteRequest createEmptyInstance() => DeleteRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DeleteRequest>(
          DeleteRequest.$_createMessage);
  static DeleteRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get chatId => $_getSZ(1);
  @$pb.TagNumber(2)
  set chatId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChatId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChatId() => $_clearField(2);

  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get messageIds => $_getList(2);

  @$pb.TagNumber(4)
  $core.bool get forAll => $_getBF(3);
  @$pb.TagNumber(4)
  set forAll($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasForAll() => $_has(3);
  @$pb.TagNumber(4)
  void clearForAll() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get threadId => $_getSZ(4);
  @$pb.TagNumber(5)
  set threadId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasThreadId() => $_has(4);
  @$pb.TagNumber(5)
  void clearThreadId() => $_clearField(5);
}

/// A span of SendRequest.text naming a person; offsets in UTF-16 code units.
class Mention extends $pb.GeneratedMessage {
  factory Mention({
    $core.int? offset,
    $core.int? length,
    User? user,
  }) {
    final result = Mention._();
    if (offset != null) result.offset = offset;
    if (length != null) result.length = length;
    if (user != null) result.user = user;
    return result;
  }

  Mention._();

  factory Mention.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Mention()..mergeFromBuffer(data, registry);
  factory Mention.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Mention()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Mention',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Mention.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'offset')
    ..aI(2, _omitFieldNames ? '' : 'length')
    ..aOM<User>(3, _omitFieldNames ? '' : 'user',
        subBuilder: User.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Mention clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Mention copyWith(void Function(Mention) updates) =>
      super.copyWith((message) => updates(message as Mention)) as Mention;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Mention() / Mention.new instead')
  static Mention create() => Mention._();
  static $pb.GeneratedMessage $_createMessage() => Mention._();
  @$core.override
  Mention createEmptyInstance() => Mention._();
  @$core.pragma('dart2js:noInline')
  static Mention getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Mention>(Mention.$_createMessage);
  static Mention? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get offset => $_getIZ(0);
  @$pb.TagNumber(1)
  set offset($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasOffset() => $_has(0);
  @$pb.TagNumber(1)
  void clearOffset() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get length => $_getIZ(1);
  @$pb.TagNumber(2)
  set length($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasLength() => $_has(1);
  @$pb.TagNumber(2)
  void clearLength() => $_clearField(2);

  @$pb.TagNumber(3)
  User get user => $_getN(2);
  @$pb.TagNumber(3)
  set user(User value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasUser() => $_has(2);
  @$pb.TagNumber(3)
  void clearUser() => $_clearField(3);
  @$pb.TagNumber(3)
  User ensureUser() => $_ensure(2);
}

class MembersRequest extends $pb.GeneratedMessage {
  factory MembersRequest({
    $core.String? accountId,
    $core.String? chatId,
    $core.String? query,
    $core.int? limit,
  }) {
    final result = MembersRequest._();
    if (accountId != null) result.accountId = accountId;
    if (chatId != null) result.chatId = chatId;
    if (query != null) result.query = query;
    if (limit != null) result.limit = limit;
    return result;
  }

  MembersRequest._();

  factory MembersRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MembersRequest()..mergeFromBuffer(data, registry);
  factory MembersRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MembersRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MembersRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: MembersRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'chatId')
    ..aOS(3, _omitFieldNames ? '' : 'query')
    ..aI(4, _omitFieldNames ? '' : 'limit')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MembersRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MembersRequest copyWith(void Function(MembersRequest) updates) =>
      super.copyWith((message) => updates(message as MembersRequest))
          as MembersRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MembersRequest() / MembersRequest.new instead')
  static MembersRequest create() => MembersRequest._();
  static $pb.GeneratedMessage $_createMessage() => MembersRequest._();
  @$core.override
  MembersRequest createEmptyInstance() => MembersRequest._();
  @$core.pragma('dart2js:noInline')
  static MembersRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MembersRequest>(
          MembersRequest.$_createMessage);
  static MembersRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get chatId => $_getSZ(1);
  @$pb.TagNumber(2)
  set chatId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChatId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChatId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get query => $_getSZ(2);
  @$pb.TagNumber(3)
  set query($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasQuery() => $_has(2);
  @$pb.TagNumber(3)
  void clearQuery() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get limit => $_getIZ(3);
  @$pb.TagNumber(4)
  set limit($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasLimit() => $_has(3);
  @$pb.TagNumber(4)
  void clearLimit() => $_clearField(4);
}

class MembersReply extends $pb.GeneratedMessage {
  factory MembersReply({
    $core.Iterable<User>? users,
  }) {
    final result = MembersReply._();
    if (users != null) result.users.addAll(users);
    return result;
  }

  MembersReply._();

  factory MembersReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MembersReply()..mergeFromBuffer(data, registry);
  factory MembersReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MembersReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MembersReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: MembersReply.$_createMessage)
    ..pPM<User>(1, _omitFieldNames ? '' : 'users',
        subBuilder: User.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MembersReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MembersReply copyWith(void Function(MembersReply) updates) =>
      super.copyWith((message) => updates(message as MembersReply))
          as MembersReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MembersReply() / MembersReply.new instead')
  static MembersReply create() => MembersReply._();
  static $pb.GeneratedMessage $_createMessage() => MembersReply._();
  @$core.override
  MembersReply createEmptyInstance() => MembersReply._();
  @$core.pragma('dart2js:noInline')
  static MembersReply getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MembersReply>(
          MembersReply.$_createMessage);
  static MembersReply? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<User> get users => $_getList(0);
}

class LinkPreviewRequest extends $pb.GeneratedMessage {
  factory LinkPreviewRequest({
    $core.String? url,
  }) {
    final result = LinkPreviewRequest._();
    if (url != null) result.url = url;
    return result;
  }

  LinkPreviewRequest._();

  factory LinkPreviewRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LinkPreviewRequest()..mergeFromBuffer(data, registry);
  factory LinkPreviewRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LinkPreviewRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LinkPreviewRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: LinkPreviewRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'url')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LinkPreviewRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LinkPreviewRequest copyWith(void Function(LinkPreviewRequest) updates) =>
      super.copyWith((message) => updates(message as LinkPreviewRequest))
          as LinkPreviewRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LinkPreviewRequest() / LinkPreviewRequest.new instead')
  static LinkPreviewRequest create() => LinkPreviewRequest._();
  static $pb.GeneratedMessage $_createMessage() => LinkPreviewRequest._();
  @$core.override
  LinkPreviewRequest createEmptyInstance() => LinkPreviewRequest._();
  @$core.pragma('dart2js:noInline')
  static LinkPreviewRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LinkPreviewRequest>(
          LinkPreviewRequest.$_createMessage);
  static LinkPreviewRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get url => $_getSZ(0);
  @$pb.TagNumber(1)
  set url($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearUrl() => $_clearField(1);
}

class LinkPreviewReply extends $pb.GeneratedMessage {
  factory LinkPreviewReply({
    $core.String? url,
    $core.String? siteName,
    $core.String? title,
    $core.String? description,
    $core.List<$core.int>? image,
    $core.String? imageMime,
  }) {
    final result = LinkPreviewReply._();
    if (url != null) result.url = url;
    if (siteName != null) result.siteName = siteName;
    if (title != null) result.title = title;
    if (description != null) result.description = description;
    if (image != null) result.image = image;
    if (imageMime != null) result.imageMime = imageMime;
    return result;
  }

  LinkPreviewReply._();

  factory LinkPreviewReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LinkPreviewReply()..mergeFromBuffer(data, registry);
  factory LinkPreviewReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LinkPreviewReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LinkPreviewReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: LinkPreviewReply.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'url')
    ..aOS(2, _omitFieldNames ? '' : 'siteName')
    ..aOS(3, _omitFieldNames ? '' : 'title')
    ..aOS(4, _omitFieldNames ? '' : 'description')
    ..a<$core.List<$core.int>>(
        5, _omitFieldNames ? '' : 'image', $pb.PbFieldType.OY)
    ..aOS(6, _omitFieldNames ? '' : 'imageMime')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LinkPreviewReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LinkPreviewReply copyWith(void Function(LinkPreviewReply) updates) =>
      super.copyWith((message) => updates(message as LinkPreviewReply))
          as LinkPreviewReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LinkPreviewReply() / LinkPreviewReply.new instead')
  static LinkPreviewReply create() => LinkPreviewReply._();
  static $pb.GeneratedMessage $_createMessage() => LinkPreviewReply._();
  @$core.override
  LinkPreviewReply createEmptyInstance() => LinkPreviewReply._();
  @$core.pragma('dart2js:noInline')
  static LinkPreviewReply getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LinkPreviewReply>(
          LinkPreviewReply.$_createMessage);
  static LinkPreviewReply? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get url => $_getSZ(0);
  @$pb.TagNumber(1)
  set url($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearUrl() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get siteName => $_getSZ(1);
  @$pb.TagNumber(2)
  set siteName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSiteName() => $_has(1);
  @$pb.TagNumber(2)
  void clearSiteName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get title => $_getSZ(2);
  @$pb.TagNumber(3)
  set title($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTitle() => $_has(2);
  @$pb.TagNumber(3)
  void clearTitle() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get description => $_getSZ(3);
  @$pb.TagNumber(4)
  set description($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDescription() => $_has(3);
  @$pb.TagNumber(4)
  void clearDescription() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.List<$core.int> get image => $_getN(4);
  @$pb.TagNumber(5)
  set image($core.List<$core.int> value) => $_setBytes(4, value);
  @$pb.TagNumber(5)
  $core.bool hasImage() => $_has(4);
  @$pb.TagNumber(5)
  void clearImage() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get imageMime => $_getSZ(5);
  @$pb.TagNumber(6)
  set imageMime($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasImageMime() => $_has(5);
  @$pb.TagNumber(6)
  void clearImageMime() => $_clearField(6);
}

class ProfileRequest extends $pb.GeneratedMessage {
  factory ProfileRequest({
    $core.String? accountId,
    $core.String? chatId,
  }) {
    final result = ProfileRequest._();
    if (accountId != null) result.accountId = accountId;
    if (chatId != null) result.chatId = chatId;
    return result;
  }

  ProfileRequest._();

  factory ProfileRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileRequest()..mergeFromBuffer(data, registry);
  factory ProfileRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProfileRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ProfileRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'chatId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileRequest copyWith(void Function(ProfileRequest) updates) =>
      super.copyWith((message) => updates(message as ProfileRequest))
          as ProfileRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ProfileRequest() / ProfileRequest.new instead')
  static ProfileRequest create() => ProfileRequest._();
  static $pb.GeneratedMessage $_createMessage() => ProfileRequest._();
  @$core.override
  ProfileRequest createEmptyInstance() => ProfileRequest._();
  @$core.pragma('dart2js:noInline')
  static ProfileRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ProfileRequest>(
          ProfileRequest.$_createMessage);
  static ProfileRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get chatId => $_getSZ(1);
  @$pb.TagNumber(2)
  set chatId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChatId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChatId() => $_clearField(2);
}

class ChatProfile extends $pb.GeneratedMessage {
  factory ChatProfile({
    $core.String? title,
    $core.String? avatarUrl,
    $core.String? username,
    $core.String? about,
    $core.String? phone,
    $core.String? email,
    $core.String? position,
    $core.String? link,
    $core.int? members,
    $core.Iterable<User>? memberList,
    $fixnum.Int64? lastSeenMs,
    $core.bool? online,
    $core.bool? bot,
  }) {
    final result = ChatProfile._();
    if (title != null) result.title = title;
    if (avatarUrl != null) result.avatarUrl = avatarUrl;
    if (username != null) result.username = username;
    if (about != null) result.about = about;
    if (phone != null) result.phone = phone;
    if (email != null) result.email = email;
    if (position != null) result.position = position;
    if (link != null) result.link = link;
    if (members != null) result.members = members;
    if (memberList != null) result.memberList.addAll(memberList);
    if (lastSeenMs != null) result.lastSeenMs = lastSeenMs;
    if (online != null) result.online = online;
    if (bot != null) result.bot = bot;
    return result;
  }

  ChatProfile._();

  factory ChatProfile.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChatProfile()..mergeFromBuffer(data, registry);
  factory ChatProfile.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChatProfile()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ChatProfile',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ChatProfile.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'title')
    ..aOS(2, _omitFieldNames ? '' : 'avatarUrl')
    ..aOS(3, _omitFieldNames ? '' : 'username')
    ..aOS(4, _omitFieldNames ? '' : 'about')
    ..aOS(5, _omitFieldNames ? '' : 'phone')
    ..aOS(6, _omitFieldNames ? '' : 'email')
    ..aOS(7, _omitFieldNames ? '' : 'position')
    ..aOS(8, _omitFieldNames ? '' : 'link')
    ..aI(9, _omitFieldNames ? '' : 'members')
    ..pPM<User>(10, _omitFieldNames ? '' : 'memberList',
        subBuilder: User.$_createMessage)
    ..aInt64(11, _omitFieldNames ? '' : 'lastSeenMs')
    ..aOB(12, _omitFieldNames ? '' : 'online')
    ..aOB(13, _omitFieldNames ? '' : 'bot')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChatProfile clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChatProfile copyWith(void Function(ChatProfile) updates) =>
      super.copyWith((message) => updates(message as ChatProfile))
          as ChatProfile;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ChatProfile() / ChatProfile.new instead')
  static ChatProfile create() => ChatProfile._();
  static $pb.GeneratedMessage $_createMessage() => ChatProfile._();
  @$core.override
  ChatProfile createEmptyInstance() => ChatProfile._();
  @$core.pragma('dart2js:noInline')
  static ChatProfile getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ChatProfile>(
          ChatProfile.$_createMessage);
  static ChatProfile? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get title => $_getSZ(0);
  @$pb.TagNumber(1)
  set title($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTitle() => $_has(0);
  @$pb.TagNumber(1)
  void clearTitle() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get avatarUrl => $_getSZ(1);
  @$pb.TagNumber(2)
  set avatarUrl($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAvatarUrl() => $_has(1);
  @$pb.TagNumber(2)
  void clearAvatarUrl() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get username => $_getSZ(2);
  @$pb.TagNumber(3)
  set username($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasUsername() => $_has(2);
  @$pb.TagNumber(3)
  void clearUsername() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get about => $_getSZ(3);
  @$pb.TagNumber(4)
  set about($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasAbout() => $_has(3);
  @$pb.TagNumber(4)
  void clearAbout() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get phone => $_getSZ(4);
  @$pb.TagNumber(5)
  set phone($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasPhone() => $_has(4);
  @$pb.TagNumber(5)
  void clearPhone() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get email => $_getSZ(5);
  @$pb.TagNumber(6)
  set email($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasEmail() => $_has(5);
  @$pb.TagNumber(6)
  void clearEmail() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get position => $_getSZ(6);
  @$pb.TagNumber(7)
  set position($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasPosition() => $_has(6);
  @$pb.TagNumber(7)
  void clearPosition() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get link => $_getSZ(7);
  @$pb.TagNumber(8)
  set link($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasLink() => $_has(7);
  @$pb.TagNumber(8)
  void clearLink() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.int get members => $_getIZ(8);
  @$pb.TagNumber(9)
  set members($core.int value) => $_setSignedInt32(8, value);
  @$pb.TagNumber(9)
  $core.bool hasMembers() => $_has(8);
  @$pb.TagNumber(9)
  void clearMembers() => $_clearField(9);

  @$pb.TagNumber(10)
  $pb.PbList<User> get memberList => $_getList(9);

  @$pb.TagNumber(11)
  $fixnum.Int64 get lastSeenMs => $_getI64(10);
  @$pb.TagNumber(11)
  set lastSeenMs($fixnum.Int64 value) => $_setInt64(10, value);
  @$pb.TagNumber(11)
  $core.bool hasLastSeenMs() => $_has(10);
  @$pb.TagNumber(11)
  void clearLastSeenMs() => $_clearField(11);

  @$pb.TagNumber(12)
  $core.bool get online => $_getBF(11);
  @$pb.TagNumber(12)
  set online($core.bool value) => $_setBool(11, value);
  @$pb.TagNumber(12)
  $core.bool hasOnline() => $_has(11);
  @$pb.TagNumber(12)
  void clearOnline() => $_clearField(12);

  @$pb.TagNumber(13)
  $core.bool get bot => $_getBF(12);
  @$pb.TagNumber(13)
  set bot($core.bool value) => $_setBool(12, value);
  @$pb.TagNumber(13)
  $core.bool hasBot() => $_has(12);
  @$pb.TagNumber(13)
  void clearBot() => $_clearField(13);
}

class MediaRequest extends $pb.GeneratedMessage {
  factory MediaRequest({
    $core.String? accountId,
    $core.String? ref,
    $fixnum.Int64? maxBytes,
    $core.bool? prefetch,
  }) {
    final result = MediaRequest._();
    if (accountId != null) result.accountId = accountId;
    if (ref != null) result.ref = ref;
    if (maxBytes != null) result.maxBytes = maxBytes;
    if (prefetch != null) result.prefetch = prefetch;
    return result;
  }

  MediaRequest._();

  factory MediaRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MediaRequest()..mergeFromBuffer(data, registry);
  factory MediaRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MediaRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MediaRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: MediaRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'ref')
    ..aInt64(3, _omitFieldNames ? '' : 'maxBytes')
    ..aOB(4, _omitFieldNames ? '' : 'prefetch')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MediaRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MediaRequest copyWith(void Function(MediaRequest) updates) =>
      super.copyWith((message) => updates(message as MediaRequest))
          as MediaRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MediaRequest() / MediaRequest.new instead')
  static MediaRequest create() => MediaRequest._();
  static $pb.GeneratedMessage $_createMessage() => MediaRequest._();
  @$core.override
  MediaRequest createEmptyInstance() => MediaRequest._();
  @$core.pragma('dart2js:noInline')
  static MediaRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MediaRequest>(
          MediaRequest.$_createMessage);
  static MediaRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get ref => $_getSZ(1);
  @$pb.TagNumber(2)
  set ref($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRef() => $_has(1);
  @$pb.TagNumber(2)
  void clearRef() => $_clearField(2);

  /// Larger files fail with RESOURCE_EXHAUSTED; 0: the core's default (32 MB).
  @$pb.TagNumber(3)
  $fixnum.Int64 get maxBytes => $_getI64(2);
  @$pb.TagNumber(3)
  set maxBytes($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasMaxBytes() => $_has(2);
  @$pb.TagNumber(3)
  void clearMaxBytes() => $_clearField(3);

  /// Only fetch it into the core's cache; the reply carries no data.
  @$pb.TagNumber(4)
  $core.bool get prefetch => $_getBF(3);
  @$pb.TagNumber(4)
  set prefetch($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPrefetch() => $_has(3);
  @$pb.TagNumber(4)
  void clearPrefetch() => $_clearField(4);
}

class MediaReply extends $pb.GeneratedMessage {
  factory MediaReply({
    $core.List<$core.int>? data,
    $core.String? mime,
  }) {
    final result = MediaReply._();
    if (data != null) result.data = data;
    if (mime != null) result.mime = mime;
    return result;
  }

  MediaReply._();

  factory MediaReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MediaReply()..mergeFromBuffer(data, registry);
  factory MediaReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MediaReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MediaReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: MediaReply.$_createMessage)
    ..a<$core.List<$core.int>>(
        1, _omitFieldNames ? '' : 'data', $pb.PbFieldType.OY)
    ..aOS(2, _omitFieldNames ? '' : 'mime')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MediaReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MediaReply copyWith(void Function(MediaReply) updates) =>
      super.copyWith((message) => updates(message as MediaReply)) as MediaReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MediaReply() / MediaReply.new instead')
  static MediaReply create() => MediaReply._();
  static $pb.GeneratedMessage $_createMessage() => MediaReply._();
  @$core.override
  MediaReply createEmptyInstance() => MediaReply._();
  @$core.pragma('dart2js:noInline')
  static MediaReply getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MediaReply>(MediaReply.$_createMessage);
  static MediaReply? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get data => $_getN(0);
  @$pb.TagNumber(1)
  set data($core.List<$core.int> value) => $_setBytes(0, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get mime => $_getSZ(1);
  @$pb.TagNumber(2)
  set mime($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMime() => $_has(1);
  @$pb.TagNumber(2)
  void clearMime() => $_clearField(2);
}

class HistoryRequest extends $pb.GeneratedMessage {
  factory HistoryRequest({
    $core.String? accountId,
    $core.String? chatId,
    $core.String? beforeId,
    $core.int? limit,
    $core.bool? cachedOnly,
  }) {
    final result = HistoryRequest._();
    if (accountId != null) result.accountId = accountId;
    if (chatId != null) result.chatId = chatId;
    if (beforeId != null) result.beforeId = beforeId;
    if (limit != null) result.limit = limit;
    if (cachedOnly != null) result.cachedOnly = cachedOnly;
    return result;
  }

  HistoryRequest._();

  factory HistoryRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HistoryRequest()..mergeFromBuffer(data, registry);
  factory HistoryRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HistoryRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'HistoryRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: HistoryRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'chatId')
    ..aOS(3, _omitFieldNames ? '' : 'beforeId')
    ..aI(4, _omitFieldNames ? '' : 'limit')
    ..aOB(5, _omitFieldNames ? '' : 'cachedOnly')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HistoryRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HistoryRequest copyWith(void Function(HistoryRequest) updates) =>
      super.copyWith((message) => updates(message as HistoryRequest))
          as HistoryRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use HistoryRequest() / HistoryRequest.new instead')
  static HistoryRequest create() => HistoryRequest._();
  static $pb.GeneratedMessage $_createMessage() => HistoryRequest._();
  @$core.override
  HistoryRequest createEmptyInstance() => HistoryRequest._();
  @$core.pragma('dart2js:noInline')
  static HistoryRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<HistoryRequest>(
          HistoryRequest.$_createMessage);
  static HistoryRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get chatId => $_getSZ(1);
  @$pb.TagNumber(2)
  set chatId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChatId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChatId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get beforeId => $_getSZ(2);
  @$pb.TagNumber(3)
  set beforeId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasBeforeId() => $_has(2);
  @$pb.TagNumber(3)
  void clearBeforeId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get limit => $_getIZ(3);
  @$pb.TagNumber(4)
  set limit($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasLimit() => $_has(3);
  @$pb.TagNumber(4)
  void clearLimit() => $_clearField(4);

  /// Only what the cache has, at once, possibly with gaps (no network): to
  /// show while the real page loads. before_id is ignored.
  @$pb.TagNumber(5)
  $core.bool get cachedOnly => $_getBF(4);
  @$pb.TagNumber(5)
  set cachedOnly($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCachedOnly() => $_has(4);
  @$pb.TagNumber(5)
  void clearCachedOnly() => $_clearField(5);
}

class HistoryReply extends $pb.GeneratedMessage {
  factory HistoryReply({
    $core.Iterable<Message>? messages,
    $core.bool? more,
  }) {
    final result = HistoryReply._();
    if (messages != null) result.messages.addAll(messages);
    if (more != null) result.more = more;
    return result;
  }

  HistoryReply._();

  factory HistoryReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HistoryReply()..mergeFromBuffer(data, registry);
  factory HistoryReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HistoryReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'HistoryReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: HistoryReply.$_createMessage)
    ..pPM<Message>(1, _omitFieldNames ? '' : 'messages',
        subBuilder: Message.$_createMessage)
    ..aOB(2, _omitFieldNames ? '' : 'more')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HistoryReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HistoryReply copyWith(void Function(HistoryReply) updates) =>
      super.copyWith((message) => updates(message as HistoryReply))
          as HistoryReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use HistoryReply() / HistoryReply.new instead')
  static HistoryReply create() => HistoryReply._();
  static $pb.GeneratedMessage $_createMessage() => HistoryReply._();
  @$core.override
  HistoryReply createEmptyInstance() => HistoryReply._();
  @$core.pragma('dart2js:noInline')
  static HistoryReply getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<HistoryReply>(
          HistoryReply.$_createMessage);
  static HistoryReply? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Message> get messages => $_getList(0);

  @$pb.TagNumber(2)
  $core.bool get more => $_getBF(1);
  @$pb.TagNumber(2)
  set more($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMore() => $_has(1);
  @$pb.TagNumber(2)
  void clearMore() => $_clearField(2);
}

class SendRequest extends $pb.GeneratedMessage {
  factory SendRequest({
    $core.String? accountId,
    $core.String? chatId,
    $core.String? text,
    $core.String? replyToId,
    $core.String? threadId,
    $core.String? quote,
    $core.Iterable<OutFile>? files,
    $core.Iterable<Mention>? mentions,
  }) {
    final result = SendRequest._();
    if (accountId != null) result.accountId = accountId;
    if (chatId != null) result.chatId = chatId;
    if (text != null) result.text = text;
    if (replyToId != null) result.replyToId = replyToId;
    if (threadId != null) result.threadId = threadId;
    if (quote != null) result.quote = quote;
    if (files != null) result.files.addAll(files);
    if (mentions != null) result.mentions.addAll(mentions);
    return result;
  }

  SendRequest._();

  factory SendRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SendRequest()..mergeFromBuffer(data, registry);
  factory SendRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SendRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SendRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: SendRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'chatId')
    ..aOS(3, _omitFieldNames ? '' : 'text')
    ..aOS(4, _omitFieldNames ? '' : 'replyToId')
    ..aOS(5, _omitFieldNames ? '' : 'threadId')
    ..aOS(6, _omitFieldNames ? '' : 'quote')
    ..pPM<OutFile>(7, _omitFieldNames ? '' : 'files',
        subBuilder: OutFile.$_createMessage)
    ..pPM<Mention>(8, _omitFieldNames ? '' : 'mentions',
        subBuilder: Mention.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SendRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SendRequest copyWith(void Function(SendRequest) updates) =>
      super.copyWith((message) => updates(message as SendRequest))
          as SendRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use SendRequest() / SendRequest.new instead')
  static SendRequest create() => SendRequest._();
  static $pb.GeneratedMessage $_createMessage() => SendRequest._();
  @$core.override
  SendRequest createEmptyInstance() => SendRequest._();
  @$core.pragma('dart2js:noInline')
  static SendRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SendRequest>(
          SendRequest.$_createMessage);
  static SendRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get chatId => $_getSZ(1);
  @$pb.TagNumber(2)
  set chatId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChatId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChatId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get text => $_getSZ(2);
  @$pb.TagNumber(3)
  set text($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasText() => $_has(2);
  @$pb.TagNumber(3)
  void clearText() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get replyToId => $_getSZ(3);
  @$pb.TagNumber(4)
  set replyToId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasReplyToId() => $_has(3);
  @$pb.TagNumber(4)
  void clearReplyToId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get threadId => $_getSZ(4);
  @$pb.TagNumber(5)
  set threadId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasThreadId() => $_has(4);
  @$pb.TagNumber(5)
  void clearThreadId() => $_clearField(5);

  /// A fragment of reply_to_id's text this reply quotes: a native quote
  /// where the messenger has one, otherwise "> " lines before the text.
  @$pb.TagNumber(6)
  $core.String get quote => $_getSZ(5);
  @$pb.TagNumber(6)
  set quote($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasQuote() => $_has(5);
  @$pb.TagNumber(6)
  void clearQuote() => $_clearField(6);

  /// Files to send (CAPABILITY_FILES), the text being their caption; at
  /// most 100 MB together. Several files may become several messages (an
  /// album): the reply is the last one, the others come as events. The
  /// reply has an empty id when the messenger does not echo what was sent
  /// (it comes as an event then).
  @$pb.TagNumber(7)
  $pb.PbList<OutFile> get files => $_getList(6);

  /// People picked after "@" (CAPABILITY_MENTIONS): spans of text the
  /// messenger turns into its mentions.
  @$pb.TagNumber(8)
  $pb.PbList<Mention> get mentions => $_getList(7);
}

class OutFile extends $pb.GeneratedMessage {
  factory OutFile({
    $core.String? name,
    $core.String? mime,
    $core.List<$core.int>? data,
  }) {
    final result = OutFile._();
    if (name != null) result.name = name;
    if (mime != null) result.mime = mime;
    if (data != null) result.data = data;
    return result;
  }

  OutFile._();

  factory OutFile.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      OutFile()..mergeFromBuffer(data, registry);
  factory OutFile.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      OutFile()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'OutFile',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: OutFile.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'mime')
    ..a<$core.List<$core.int>>(
        3, _omitFieldNames ? '' : 'data', $pb.PbFieldType.OY)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  OutFile clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  OutFile copyWith(void Function(OutFile) updates) =>
      super.copyWith((message) => updates(message as OutFile)) as OutFile;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use OutFile() / OutFile.new instead')
  static OutFile create() => OutFile._();
  static $pb.GeneratedMessage $_createMessage() => OutFile._();
  @$core.override
  OutFile createEmptyInstance() => OutFile._();
  @$core.pragma('dart2js:noInline')
  static OutFile getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<OutFile>(OutFile.$_createMessage);
  static OutFile? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get mime => $_getSZ(1);
  @$pb.TagNumber(2)
  set mime($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMime() => $_has(1);
  @$pb.TagNumber(2)
  void clearMime() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.List<$core.int> get data => $_getN(2);
  @$pb.TagNumber(3)
  set data($core.List<$core.int> value) => $_setBytes(2, value);
  @$pb.TagNumber(3)
  $core.bool hasData() => $_has(2);
  @$pb.TagNumber(3)
  void clearData() => $_clearField(3);
}

/// Forwards messages of one chat to another chat of the same account
/// (CAPABILITY_FORWARD), attachments and the "forwarded" mark included.
class ForwardRequest extends $pb.GeneratedMessage {
  factory ForwardRequest({
    $core.String? accountId,
    $core.String? chatId,
    $core.Iterable<$core.String>? messageIds,
    $core.String? toChatId,
  }) {
    final result = ForwardRequest._();
    if (accountId != null) result.accountId = accountId;
    if (chatId != null) result.chatId = chatId;
    if (messageIds != null) result.messageIds.addAll(messageIds);
    if (toChatId != null) result.toChatId = toChatId;
    return result;
  }

  ForwardRequest._();

  factory ForwardRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ForwardRequest()..mergeFromBuffer(data, registry);
  factory ForwardRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ForwardRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ForwardRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ForwardRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'chatId')
    ..pPS(3, _omitFieldNames ? '' : 'messageIds')
    ..aOS(4, _omitFieldNames ? '' : 'toChatId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ForwardRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ForwardRequest copyWith(void Function(ForwardRequest) updates) =>
      super.copyWith((message) => updates(message as ForwardRequest))
          as ForwardRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ForwardRequest() / ForwardRequest.new instead')
  static ForwardRequest create() => ForwardRequest._();
  static $pb.GeneratedMessage $_createMessage() => ForwardRequest._();
  @$core.override
  ForwardRequest createEmptyInstance() => ForwardRequest._();
  @$core.pragma('dart2js:noInline')
  static ForwardRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ForwardRequest>(
          ForwardRequest.$_createMessage);
  static ForwardRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get chatId => $_getSZ(1);
  @$pb.TagNumber(2)
  set chatId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChatId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChatId() => $_clearField(2);

  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get messageIds => $_getList(2);

  @$pb.TagNumber(4)
  $core.String get toChatId => $_getSZ(3);
  @$pb.TagNumber(4)
  set toChatId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasToChatId() => $_has(3);
  @$pb.TagNumber(4)
  void clearToChatId() => $_clearField(4);
}

class ForwardReply extends $pb.GeneratedMessage {
  factory ForwardReply({
    $core.Iterable<Message>? messages,
  }) {
    final result = ForwardReply._();
    if (messages != null) result.messages.addAll(messages);
    return result;
  }

  ForwardReply._();

  factory ForwardReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ForwardReply()..mergeFromBuffer(data, registry);
  factory ForwardReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ForwardReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ForwardReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: ForwardReply.$_createMessage)
    ..pPM<Message>(1, _omitFieldNames ? '' : 'messages',
        subBuilder: Message.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ForwardReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ForwardReply copyWith(void Function(ForwardReply) updates) =>
      super.copyWith((message) => updates(message as ForwardReply))
          as ForwardReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ForwardReply() / ForwardReply.new instead')
  static ForwardReply create() => ForwardReply._();
  static $pb.GeneratedMessage $_createMessage() => ForwardReply._();
  @$core.override
  ForwardReply createEmptyInstance() => ForwardReply._();
  @$core.pragma('dart2js:noInline')
  static ForwardReply getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ForwardReply>(
          ForwardReply.$_createMessage);
  static ForwardReply? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Message> get messages => $_getList(0);
}

class MarkReadRequest extends $pb.GeneratedMessage {
  factory MarkReadRequest({
    $core.String? accountId,
    $core.String? chatId,
    $core.String? messageId,
    $core.String? threadId,
  }) {
    final result = MarkReadRequest._();
    if (accountId != null) result.accountId = accountId;
    if (chatId != null) result.chatId = chatId;
    if (messageId != null) result.messageId = messageId;
    if (threadId != null) result.threadId = threadId;
    return result;
  }

  MarkReadRequest._();

  factory MarkReadRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MarkReadRequest()..mergeFromBuffer(data, registry);
  factory MarkReadRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MarkReadRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MarkReadRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: MarkReadRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'chatId')
    ..aOS(3, _omitFieldNames ? '' : 'messageId')
    ..aOS(4, _omitFieldNames ? '' : 'threadId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MarkReadRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MarkReadRequest copyWith(void Function(MarkReadRequest) updates) =>
      super.copyWith((message) => updates(message as MarkReadRequest))
          as MarkReadRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MarkReadRequest() / MarkReadRequest.new instead')
  static MarkReadRequest create() => MarkReadRequest._();
  static $pb.GeneratedMessage $_createMessage() => MarkReadRequest._();
  @$core.override
  MarkReadRequest createEmptyInstance() => MarkReadRequest._();
  @$core.pragma('dart2js:noInline')
  static MarkReadRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MarkReadRequest>(
          MarkReadRequest.$_createMessage);
  static MarkReadRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get chatId => $_getSZ(1);
  @$pb.TagNumber(2)
  set chatId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChatId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChatId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get messageId => $_getSZ(2);
  @$pb.TagNumber(3)
  set messageId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasMessageId() => $_has(2);
  @$pb.TagNumber(3)
  void clearMessageId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get threadId => $_getSZ(3);
  @$pb.TagNumber(4)
  set threadId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasThreadId() => $_has(3);
  @$pb.TagNumber(4)
  void clearThreadId() => $_clearField(4);
}

class Event extends $pb.GeneratedMessage {
  factory Event({
    Event_Kind? kind,
    Account? account,
    Chat? chat,
    Message? message,
    $core.String? accountId,
    Call? call,
    CallSignal? signal,
    Mail? mail,
    $core.bool? notify,
  }) {
    final result = Event._();
    if (kind != null) result.kind = kind;
    if (account != null) result.account = account;
    if (chat != null) result.chat = chat;
    if (message != null) result.message = message;
    if (accountId != null) result.accountId = accountId;
    if (call != null) result.call = call;
    if (signal != null) result.signal = signal;
    if (mail != null) result.mail = mail;
    if (notify != null) result.notify = notify;
    return result;
  }

  Event._();

  factory Event.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Event()..mergeFromBuffer(data, registry);
  factory Event.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Event()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Event',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Event.$_createMessage)
    ..aE<Event_Kind>(1, _omitFieldNames ? '' : 'kind',
        enumValues: Event_Kind.values)
    ..aOM<Account>(2, _omitFieldNames ? '' : 'account',
        subBuilder: Account.$_createMessage)
    ..aOM<Chat>(3, _omitFieldNames ? '' : 'chat',
        subBuilder: Chat.$_createMessage)
    ..aOM<Message>(4, _omitFieldNames ? '' : 'message',
        subBuilder: Message.$_createMessage)
    ..aOS(5, _omitFieldNames ? '' : 'accountId')
    ..aOM<Call>(6, _omitFieldNames ? '' : 'call',
        subBuilder: Call.$_createMessage)
    ..aOM<CallSignal>(7, _omitFieldNames ? '' : 'signal',
        subBuilder: CallSignal.$_createMessage)
    ..aOM<Mail>(8, _omitFieldNames ? '' : 'mail',
        subBuilder: Mail.$_createMessage)
    ..aOB(9, _omitFieldNames ? '' : 'notify')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Event clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Event copyWith(void Function(Event) updates) =>
      super.copyWith((message) => updates(message as Event)) as Event;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Event() / Event.new instead')
  static Event create() => Event._();
  static $pb.GeneratedMessage $_createMessage() => Event._();
  @$core.override
  Event createEmptyInstance() => Event._();
  @$core.pragma('dart2js:noInline')
  static Event getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Event>(Event.$_createMessage);
  static Event? _defaultInstance;

  @$pb.TagNumber(1)
  Event_Kind get kind => $_getN(0);
  @$pb.TagNumber(1)
  set kind(Event_Kind value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasKind() => $_has(0);
  @$pb.TagNumber(1)
  void clearKind() => $_clearField(1);

  @$pb.TagNumber(2)
  Account get account => $_getN(1);
  @$pb.TagNumber(2)
  set account(Account value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasAccount() => $_has(1);
  @$pb.TagNumber(2)
  void clearAccount() => $_clearField(2);
  @$pb.TagNumber(2)
  Account ensureAccount() => $_ensure(1);

  @$pb.TagNumber(3)
  Chat get chat => $_getN(2);
  @$pb.TagNumber(3)
  set chat(Chat value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasChat() => $_has(2);
  @$pb.TagNumber(3)
  void clearChat() => $_clearField(3);
  @$pb.TagNumber(3)
  Chat ensureChat() => $_ensure(2);

  @$pb.TagNumber(4)
  Message get message => $_getN(3);
  @$pb.TagNumber(4)
  set message(Message value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasMessage() => $_has(3);
  @$pb.TagNumber(4)
  void clearMessage() => $_clearField(4);
  @$pb.TagNumber(4)
  Message ensureMessage() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.String get accountId => $_getSZ(4);
  @$pb.TagNumber(5)
  set accountId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAccountId() => $_has(4);
  @$pb.TagNumber(5)
  void clearAccountId() => $_clearField(5);

  @$pb.TagNumber(6)
  Call get call => $_getN(5);
  @$pb.TagNumber(6)
  set call(Call value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasCall() => $_has(5);
  @$pb.TagNumber(6)
  void clearCall() => $_clearField(6);
  @$pb.TagNumber(6)
  Call ensureCall() => $_ensure(5);

  @$pb.TagNumber(7)
  CallSignal get signal => $_getN(6);
  @$pb.TagNumber(7)
  set signal(CallSignal value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasSignal() => $_has(6);
  @$pb.TagNumber(7)
  void clearSignal() => $_clearField(7);
  @$pb.TagNumber(7)
  CallSignal ensureSignal() => $_ensure(6);

  @$pb.TagNumber(8)
  Mail get mail => $_getN(7);
  @$pb.TagNumber(8)
  set mail(Mail value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasMail() => $_has(7);
  @$pb.TagNumber(8)
  void clearMail() => $_clearField(8);
  @$pb.TagNumber(8)
  Mail ensureMail() => $_ensure(7);

  /// KIND_MESSAGE_NEW / KIND_THREAD_MESSAGE worth a notification: from
  /// someone else, in a chat that is not muted or mentioning the user
  /// (in a thread: only a mention, once). The app adds whether it is seen.
  @$pb.TagNumber(9)
  $core.bool get notify => $_getBF(8);
  @$pb.TagNumber(9)
  set notify($core.bool value) => $_setBool(8, value);
  @$pb.TagNumber(9)
  $core.bool hasNotify() => $_has(8);
  @$pb.TagNumber(9)
  void clearNotify() => $_clearField(9);
}

class LogsRequest extends $pb.GeneratedMessage {
  factory LogsRequest({
    $core.String? accountId,
    $core.int? limit,
  }) {
    final result = LogsRequest._();
    if (accountId != null) result.accountId = accountId;
    if (limit != null) result.limit = limit;
    return result;
  }

  LogsRequest._();

  factory LogsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogsRequest()..mergeFromBuffer(data, registry);
  factory LogsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LogsRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: LogsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aI(2, _omitFieldNames ? '' : 'limit')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogsRequest copyWith(void Function(LogsRequest) updates) =>
      super.copyWith((message) => updates(message as LogsRequest))
          as LogsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LogsRequest() / LogsRequest.new instead')
  static LogsRequest create() => LogsRequest._();
  static $pb.GeneratedMessage $_createMessage() => LogsRequest._();
  @$core.override
  LogsRequest createEmptyInstance() => LogsRequest._();
  @$core.pragma('dart2js:noInline')
  static LogsRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LogsRequest>(
          LogsRequest.$_createMessage);
  static LogsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get limit => $_getIZ(1);
  @$pb.TagNumber(2)
  set limit($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasLimit() => $_has(1);
  @$pb.TagNumber(2)
  void clearLimit() => $_clearField(2);
}

class LogEntry extends $pb.GeneratedMessage {
  factory LogEntry({
    $fixnum.Int64? seq,
    $fixnum.Int64? timeMs,
    $fixnum.Int64? firstMs,
    $core.String? level,
    $core.String? source,
    $core.String? accountId,
    $core.String? message,
    $core.int? count,
  }) {
    final result = LogEntry._();
    if (seq != null) result.seq = seq;
    if (timeMs != null) result.timeMs = timeMs;
    if (firstMs != null) result.firstMs = firstMs;
    if (level != null) result.level = level;
    if (source != null) result.source = source;
    if (accountId != null) result.accountId = accountId;
    if (message != null) result.message = message;
    if (count != null) result.count = count;
    return result;
  }

  LogEntry._();

  factory LogEntry.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogEntry()..mergeFromBuffer(data, registry);
  factory LogEntry.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogEntry()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LogEntry',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: LogEntry.$_createMessage)
    ..a<$fixnum.Int64>(1, _omitFieldNames ? '' : 'seq', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..aInt64(2, _omitFieldNames ? '' : 'timeMs')
    ..aInt64(3, _omitFieldNames ? '' : 'firstMs')
    ..aOS(4, _omitFieldNames ? '' : 'level')
    ..aOS(5, _omitFieldNames ? '' : 'source')
    ..aOS(6, _omitFieldNames ? '' : 'accountId')
    ..aOS(7, _omitFieldNames ? '' : 'message')
    ..aI(8, _omitFieldNames ? '' : 'count')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogEntry clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogEntry copyWith(void Function(LogEntry) updates) =>
      super.copyWith((message) => updates(message as LogEntry)) as LogEntry;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LogEntry() / LogEntry.new instead')
  static LogEntry create() => LogEntry._();
  static $pb.GeneratedMessage $_createMessage() => LogEntry._();
  @$core.override
  LogEntry createEmptyInstance() => LogEntry._();
  @$core.pragma('dart2js:noInline')
  static LogEntry getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LogEntry>(LogEntry.$_createMessage);
  static LogEntry? _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get seq => $_getI64(0);
  @$pb.TagNumber(1)
  set seq($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSeq() => $_has(0);
  @$pb.TagNumber(1)
  void clearSeq() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get timeMs => $_getI64(1);
  @$pb.TagNumber(2)
  set timeMs($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTimeMs() => $_has(1);
  @$pb.TagNumber(2)
  void clearTimeMs() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get firstMs => $_getI64(2);
  @$pb.TagNumber(3)
  set firstMs($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasFirstMs() => $_has(2);
  @$pb.TagNumber(3)
  void clearFirstMs() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get level => $_getSZ(3);
  @$pb.TagNumber(4)
  set level($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasLevel() => $_has(3);
  @$pb.TagNumber(4)
  void clearLevel() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get source => $_getSZ(4);
  @$pb.TagNumber(5)
  set source($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSource() => $_has(4);
  @$pb.TagNumber(5)
  void clearSource() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get accountId => $_getSZ(5);
  @$pb.TagNumber(6)
  set accountId($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasAccountId() => $_has(5);
  @$pb.TagNumber(6)
  void clearAccountId() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get message => $_getSZ(6);
  @$pb.TagNumber(7)
  set message($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasMessage() => $_has(6);
  @$pb.TagNumber(7)
  void clearMessage() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.int get count => $_getIZ(7);
  @$pb.TagNumber(8)
  set count($core.int value) => $_setSignedInt32(7, value);
  @$pb.TagNumber(8)
  $core.bool hasCount() => $_has(7);
  @$pb.TagNumber(8)
  void clearCount() => $_clearField(8);
}

class LogsReply extends $pb.GeneratedMessage {
  factory LogsReply({
    $core.Iterable<LogEntry>? entries,
  }) {
    final result = LogsReply._();
    if (entries != null) result.entries.addAll(entries);
    return result;
  }

  LogsReply._();

  factory LogsReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogsReply()..mergeFromBuffer(data, registry);
  factory LogsReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogsReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LogsReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: LogsReply.$_createMessage)
    ..pPM<LogEntry>(1, _omitFieldNames ? '' : 'entries',
        subBuilder: LogEntry.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogsReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogsReply copyWith(void Function(LogsReply) updates) =>
      super.copyWith((message) => updates(message as LogsReply)) as LogsReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LogsReply() / LogsReply.new instead')
  static LogsReply create() => LogsReply._();
  static $pb.GeneratedMessage $_createMessage() => LogsReply._();
  @$core.override
  LogsReply createEmptyInstance() => LogsReply._();
  @$core.pragma('dart2js:noInline')
  static LogsReply getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LogsReply>(LogsReply.$_createMessage);
  static LogsReply? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<LogEntry> get entries => $_getList(0);
}

class IceServer extends $pb.GeneratedMessage {
  factory IceServer({
    $core.Iterable<$core.String>? urls,
    $core.String? username,
    $core.String? credential,
  }) {
    final result = IceServer._();
    if (urls != null) result.urls.addAll(urls);
    if (username != null) result.username = username;
    if (credential != null) result.credential = credential;
    return result;
  }

  IceServer._();

  factory IceServer.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      IceServer()..mergeFromBuffer(data, registry);
  factory IceServer.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      IceServer()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'IceServer',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: IceServer.$_createMessage)
    ..pPS(1, _omitFieldNames ? '' : 'urls')
    ..aOS(2, _omitFieldNames ? '' : 'username')
    ..aOS(3, _omitFieldNames ? '' : 'credential')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  IceServer clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  IceServer copyWith(void Function(IceServer) updates) =>
      super.copyWith((message) => updates(message as IceServer)) as IceServer;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use IceServer() / IceServer.new instead')
  static IceServer create() => IceServer._();
  static $pb.GeneratedMessage $_createMessage() => IceServer._();
  @$core.override
  IceServer createEmptyInstance() => IceServer._();
  @$core.pragma('dart2js:noInline')
  static IceServer getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<IceServer>(IceServer.$_createMessage);
  static IceServer? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$core.String> get urls => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get username => $_getSZ(1);
  @$pb.TagNumber(2)
  set username($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUsername() => $_has(1);
  @$pb.TagNumber(2)
  void clearUsername() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get credential => $_getSZ(2);
  @$pb.TagNumber(3)
  set credential($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCredential() => $_has(2);
  @$pb.TagNumber(3)
  void clearCredential() => $_clearField(3);
}

class CallMedia extends $pb.GeneratedMessage {
  factory CallMedia({
    $core.bool? muted,
    $core.bool? video,
    $core.bool? screen,
  }) {
    final result = CallMedia._();
    if (muted != null) result.muted = muted;
    if (video != null) result.video = video;
    if (screen != null) result.screen = screen;
    return result;
  }

  CallMedia._();

  factory CallMedia.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CallMedia()..mergeFromBuffer(data, registry);
  factory CallMedia.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CallMedia()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CallMedia',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: CallMedia.$_createMessage)
    ..aOB(1, _omitFieldNames ? '' : 'muted')
    ..aOB(2, _omitFieldNames ? '' : 'video')
    ..aOB(3, _omitFieldNames ? '' : 'screen')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CallMedia clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CallMedia copyWith(void Function(CallMedia) updates) =>
      super.copyWith((message) => updates(message as CallMedia)) as CallMedia;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CallMedia() / CallMedia.new instead')
  static CallMedia create() => CallMedia._();
  static $pb.GeneratedMessage $_createMessage() => CallMedia._();
  @$core.override
  CallMedia createEmptyInstance() => CallMedia._();
  @$core.pragma('dart2js:noInline')
  static CallMedia getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CallMedia>(CallMedia.$_createMessage);
  static CallMedia? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get muted => $_getBF(0);
  @$pb.TagNumber(1)
  set muted($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMuted() => $_has(0);
  @$pb.TagNumber(1)
  void clearMuted() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get video => $_getBF(1);
  @$pb.TagNumber(2)
  set video($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasVideo() => $_has(1);
  @$pb.TagNumber(2)
  void clearVideo() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get screen => $_getBF(2);
  @$pb.TagNumber(3)
  set screen($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasScreen() => $_has(2);
  @$pb.TagNumber(3)
  void clearScreen() => $_clearField(3);
}

class Call extends $pb.GeneratedMessage {
  factory Call({
    $core.String? accountId,
    $core.String? id,
    $core.String? chatId,
    User? peer,
    $core.bool? outgoing,
    $core.bool? video,
    CallState? state,
    CallEnd? end,
    $core.String? error,
    $fixnum.Int64? startedMs,
    $core.Iterable<$core.String>? verification,
    $core.Iterable<IceServer>? ice,
    $core.bool? relayOnly,
    $core.bool? polite,
    $core.String? channel,
    $core.bool? openChannel,
    $core.bool? offer,
    CallMedia? remote,
  }) {
    final result = Call._();
    if (accountId != null) result.accountId = accountId;
    if (id != null) result.id = id;
    if (chatId != null) result.chatId = chatId;
    if (peer != null) result.peer = peer;
    if (outgoing != null) result.outgoing = outgoing;
    if (video != null) result.video = video;
    if (state != null) result.state = state;
    if (end != null) result.end = end;
    if (error != null) result.error = error;
    if (startedMs != null) result.startedMs = startedMs;
    if (verification != null) result.verification.addAll(verification);
    if (ice != null) result.ice.addAll(ice);
    if (relayOnly != null) result.relayOnly = relayOnly;
    if (polite != null) result.polite = polite;
    if (channel != null) result.channel = channel;
    if (openChannel != null) result.openChannel = openChannel;
    if (offer != null) result.offer = offer;
    if (remote != null) result.remote = remote;
    return result;
  }

  Call._();

  factory Call.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Call()..mergeFromBuffer(data, registry);
  factory Call.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Call()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Call',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Call.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'id')
    ..aOS(3, _omitFieldNames ? '' : 'chatId')
    ..aOM<User>(4, _omitFieldNames ? '' : 'peer',
        subBuilder: User.$_createMessage)
    ..aOB(5, _omitFieldNames ? '' : 'outgoing')
    ..aOB(6, _omitFieldNames ? '' : 'video')
    ..aE<CallState>(7, _omitFieldNames ? '' : 'state',
        enumValues: CallState.values)
    ..aE<CallEnd>(8, _omitFieldNames ? '' : 'end', enumValues: CallEnd.values)
    ..aOS(9, _omitFieldNames ? '' : 'error')
    ..aInt64(10, _omitFieldNames ? '' : 'startedMs')
    ..pPS(11, _omitFieldNames ? '' : 'verification')
    ..pPM<IceServer>(12, _omitFieldNames ? '' : 'ice',
        subBuilder: IceServer.$_createMessage)
    ..aOB(13, _omitFieldNames ? '' : 'relayOnly')
    ..aOB(14, _omitFieldNames ? '' : 'polite')
    ..aOS(15, _omitFieldNames ? '' : 'channel')
    ..aOB(16, _omitFieldNames ? '' : 'openChannel')
    ..aOB(17, _omitFieldNames ? '' : 'offer')
    ..aOM<CallMedia>(18, _omitFieldNames ? '' : 'remote',
        subBuilder: CallMedia.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Call clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Call copyWith(void Function(Call) updates) =>
      super.copyWith((message) => updates(message as Call)) as Call;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Call() / Call.new instead')
  static Call create() => Call._();
  static $pb.GeneratedMessage $_createMessage() => Call._();
  @$core.override
  Call createEmptyInstance() => Call._();
  @$core.pragma('dart2js:noInline')
  static Call getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Call>(Call.$_createMessage);
  static Call? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get id => $_getSZ(1);
  @$pb.TagNumber(2)
  set id($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasId() => $_has(1);
  @$pb.TagNumber(2)
  void clearId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get chatId => $_getSZ(2);
  @$pb.TagNumber(3)
  set chatId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasChatId() => $_has(2);
  @$pb.TagNumber(3)
  void clearChatId() => $_clearField(3);

  @$pb.TagNumber(4)
  User get peer => $_getN(3);
  @$pb.TagNumber(4)
  set peer(User value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasPeer() => $_has(3);
  @$pb.TagNumber(4)
  void clearPeer() => $_clearField(4);
  @$pb.TagNumber(4)
  User ensurePeer() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.bool get outgoing => $_getBF(4);
  @$pb.TagNumber(5)
  set outgoing($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasOutgoing() => $_has(4);
  @$pb.TagNumber(5)
  void clearOutgoing() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.bool get video => $_getBF(5);
  @$pb.TagNumber(6)
  set video($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasVideo() => $_has(5);
  @$pb.TagNumber(6)
  void clearVideo() => $_clearField(6);

  @$pb.TagNumber(7)
  CallState get state => $_getN(6);
  @$pb.TagNumber(7)
  set state(CallState value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasState() => $_has(6);
  @$pb.TagNumber(7)
  void clearState() => $_clearField(7);

  @$pb.TagNumber(8)
  CallEnd get end => $_getN(7);
  @$pb.TagNumber(8)
  set end(CallEnd value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasEnd() => $_has(7);
  @$pb.TagNumber(8)
  void clearEnd() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get error => $_getSZ(8);
  @$pb.TagNumber(9)
  set error($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasError() => $_has(8);
  @$pb.TagNumber(9)
  void clearError() => $_clearField(9);

  @$pb.TagNumber(10)
  $fixnum.Int64 get startedMs => $_getI64(9);
  @$pb.TagNumber(10)
  set startedMs($fixnum.Int64 value) => $_setInt64(9, value);
  @$pb.TagNumber(10)
  $core.bool hasStartedMs() => $_has(9);
  @$pb.TagNumber(10)
  void clearStartedMs() => $_clearField(10);

  @$pb.TagNumber(11)
  $pb.PbList<$core.String> get verification => $_getList(10);

  @$pb.TagNumber(12)
  $pb.PbList<IceServer> get ice => $_getList(11);

  @$pb.TagNumber(13)
  $core.bool get relayOnly => $_getBF(12);
  @$pb.TagNumber(13)
  set relayOnly($core.bool value) => $_setBool(12, value);
  @$pb.TagNumber(13)
  $core.bool hasRelayOnly() => $_has(12);
  @$pb.TagNumber(13)
  void clearRelayOnly() => $_clearField(13);

  /// Perfect negotiation: polite peers roll back their offer on a collision.
  @$pb.TagNumber(14)
  $core.bool get polite => $_getBF(13);
  @$pb.TagNumber(14)
  set polite($core.bool value) => $_setBool(13, value);
  @$pb.TagNumber(14)
  $core.bool hasPolite() => $_has(13);
  @$pb.TagNumber(14)
  void clearPolite() => $_clearField(14);

  @$pb.TagNumber(15)
  $core.String get channel => $_getSZ(14);
  @$pb.TagNumber(15)
  set channel($core.String value) => $_setString(14, value);
  @$pb.TagNumber(15)
  $core.bool hasChannel() => $_has(14);
  @$pb.TagNumber(15)
  void clearChannel() => $_clearField(15);

  @$pb.TagNumber(16)
  $core.bool get openChannel => $_getBF(15);
  @$pb.TagNumber(16)
  set openChannel($core.bool value) => $_setBool(15, value);
  @$pb.TagNumber(16)
  $core.bool hasOpenChannel() => $_has(15);
  @$pb.TagNumber(16)
  void clearOpenChannel() => $_clearField(16);

  @$pb.TagNumber(17)
  $core.bool get offer => $_getBF(16);
  @$pb.TagNumber(17)
  set offer($core.bool value) => $_setBool(16, value);
  @$pb.TagNumber(17)
  $core.bool hasOffer() => $_has(16);
  @$pb.TagNumber(17)
  void clearOffer() => $_clearField(17);

  @$pb.TagNumber(18)
  CallMedia get remote => $_getN(17);
  @$pb.TagNumber(18)
  set remote(CallMedia value) => $_setField(18, value);
  @$pb.TagNumber(18)
  $core.bool hasRemote() => $_has(17);
  @$pb.TagNumber(18)
  void clearRemote() => $_clearField(18);
  @$pb.TagNumber(18)
  CallMedia ensureRemote() => $_ensure(17);
}

class CallSignal extends $pb.GeneratedMessage {
  factory CallSignal({
    SignalKind? kind,
    $core.String? sdp,
    $core.String? mid,
    $core.int? mline,
    $core.String? data,
    $fixnum.Int64? seq,
  }) {
    final result = CallSignal._();
    if (kind != null) result.kind = kind;
    if (sdp != null) result.sdp = sdp;
    if (mid != null) result.mid = mid;
    if (mline != null) result.mline = mline;
    if (data != null) result.data = data;
    if (seq != null) result.seq = seq;
    return result;
  }

  CallSignal._();

  factory CallSignal.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CallSignal()..mergeFromBuffer(data, registry);
  factory CallSignal.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CallSignal()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CallSignal',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: CallSignal.$_createMessage)
    ..aE<SignalKind>(1, _omitFieldNames ? '' : 'kind',
        enumValues: SignalKind.values)
    ..aOS(2, _omitFieldNames ? '' : 'sdp')
    ..aOS(3, _omitFieldNames ? '' : 'mid')
    ..aI(4, _omitFieldNames ? '' : 'mline')
    ..aOS(5, _omitFieldNames ? '' : 'data')
    ..aInt64(6, _omitFieldNames ? '' : 'seq')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CallSignal clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CallSignal copyWith(void Function(CallSignal) updates) =>
      super.copyWith((message) => updates(message as CallSignal)) as CallSignal;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CallSignal() / CallSignal.new instead')
  static CallSignal create() => CallSignal._();
  static $pb.GeneratedMessage $_createMessage() => CallSignal._();
  @$core.override
  CallSignal createEmptyInstance() => CallSignal._();
  @$core.pragma('dart2js:noInline')
  static CallSignal getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CallSignal>(CallSignal.$_createMessage);
  static CallSignal? _defaultInstance;

  @$pb.TagNumber(1)
  SignalKind get kind => $_getN(0);
  @$pb.TagNumber(1)
  set kind(SignalKind value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasKind() => $_has(0);
  @$pb.TagNumber(1)
  void clearKind() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get sdp => $_getSZ(1);
  @$pb.TagNumber(2)
  set sdp($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSdp() => $_has(1);
  @$pb.TagNumber(2)
  void clearSdp() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get mid => $_getSZ(2);
  @$pb.TagNumber(3)
  set mid($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasMid() => $_has(2);
  @$pb.TagNumber(3)
  void clearMid() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get mline => $_getIZ(3);
  @$pb.TagNumber(4)
  set mline($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasMline() => $_has(3);
  @$pb.TagNumber(4)
  void clearMline() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get data => $_getSZ(4);
  @$pb.TagNumber(5)
  set data($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasData() => $_has(4);
  @$pb.TagNumber(5)
  void clearData() => $_clearField(5);

  @$pb.TagNumber(6)
  $fixnum.Int64 get seq => $_getI64(5);
  @$pb.TagNumber(6)
  set seq($fixnum.Int64 value) => $_setInt64(5, value);
  @$pb.TagNumber(6)
  $core.bool hasSeq() => $_has(5);
  @$pb.TagNumber(6)
  void clearSeq() => $_clearField(6);
}

class StartCallRequest extends $pb.GeneratedMessage {
  factory StartCallRequest({
    $core.String? accountId,
    $core.String? chatId,
    $core.bool? video,
  }) {
    final result = StartCallRequest._();
    if (accountId != null) result.accountId = accountId;
    if (chatId != null) result.chatId = chatId;
    if (video != null) result.video = video;
    return result;
  }

  StartCallRequest._();

  factory StartCallRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StartCallRequest()..mergeFromBuffer(data, registry);
  factory StartCallRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StartCallRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StartCallRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: StartCallRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'chatId')
    ..aOB(3, _omitFieldNames ? '' : 'video')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartCallRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartCallRequest copyWith(void Function(StartCallRequest) updates) =>
      super.copyWith((message) => updates(message as StartCallRequest))
          as StartCallRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use StartCallRequest() / StartCallRequest.new instead')
  static StartCallRequest create() => StartCallRequest._();
  static $pb.GeneratedMessage $_createMessage() => StartCallRequest._();
  @$core.override
  StartCallRequest createEmptyInstance() => StartCallRequest._();
  @$core.pragma('dart2js:noInline')
  static StartCallRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<StartCallRequest>(
          StartCallRequest.$_createMessage);
  static StartCallRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get chatId => $_getSZ(1);
  @$pb.TagNumber(2)
  set chatId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChatId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChatId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get video => $_getBF(2);
  @$pb.TagNumber(3)
  set video($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasVideo() => $_has(2);
  @$pb.TagNumber(3)
  void clearVideo() => $_clearField(3);
}

class CallRequest extends $pb.GeneratedMessage {
  factory CallRequest({
    $core.String? accountId,
    $core.String? callId,
    $core.bool? video,
  }) {
    final result = CallRequest._();
    if (accountId != null) result.accountId = accountId;
    if (callId != null) result.callId = callId;
    if (video != null) result.video = video;
    return result;
  }

  CallRequest._();

  factory CallRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CallRequest()..mergeFromBuffer(data, registry);
  factory CallRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CallRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CallRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: CallRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'callId')
    ..aOB(3, _omitFieldNames ? '' : 'video')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CallRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CallRequest copyWith(void Function(CallRequest) updates) =>
      super.copyWith((message) => updates(message as CallRequest))
          as CallRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CallRequest() / CallRequest.new instead')
  static CallRequest create() => CallRequest._();
  static $pb.GeneratedMessage $_createMessage() => CallRequest._();
  @$core.override
  CallRequest createEmptyInstance() => CallRequest._();
  @$core.pragma('dart2js:noInline')
  static CallRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CallRequest>(
          CallRequest.$_createMessage);
  static CallRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get callId => $_getSZ(1);
  @$pb.TagNumber(2)
  set callId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCallId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCallId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get video => $_getBF(2);
  @$pb.TagNumber(3)
  set video($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasVideo() => $_has(2);
  @$pb.TagNumber(3)
  void clearVideo() => $_clearField(3);
}

class EndCallRequest extends $pb.GeneratedMessage {
  factory EndCallRequest({
    $core.String? accountId,
    $core.String? callId,
    CallEnd? reason,
  }) {
    final result = EndCallRequest._();
    if (accountId != null) result.accountId = accountId;
    if (callId != null) result.callId = callId;
    if (reason != null) result.reason = reason;
    return result;
  }

  EndCallRequest._();

  factory EndCallRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      EndCallRequest()..mergeFromBuffer(data, registry);
  factory EndCallRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      EndCallRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EndCallRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: EndCallRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'callId')
    ..aE<CallEnd>(3, _omitFieldNames ? '' : 'reason',
        enumValues: CallEnd.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EndCallRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EndCallRequest copyWith(void Function(EndCallRequest) updates) =>
      super.copyWith((message) => updates(message as EndCallRequest))
          as EndCallRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use EndCallRequest() / EndCallRequest.new instead')
  static EndCallRequest create() => EndCallRequest._();
  static $pb.GeneratedMessage $_createMessage() => EndCallRequest._();
  @$core.override
  EndCallRequest createEmptyInstance() => EndCallRequest._();
  @$core.pragma('dart2js:noInline')
  static EndCallRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<EndCallRequest>(
          EndCallRequest.$_createMessage);
  static EndCallRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get callId => $_getSZ(1);
  @$pb.TagNumber(2)
  set callId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCallId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCallId() => $_clearField(2);

  @$pb.TagNumber(3)
  CallEnd get reason => $_getN(2);
  @$pb.TagNumber(3)
  set reason(CallEnd value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasReason() => $_has(2);
  @$pb.TagNumber(3)
  void clearReason() => $_clearField(3);
}

class CallSignalRequest extends $pb.GeneratedMessage {
  factory CallSignalRequest({
    $core.String? accountId,
    $core.String? callId,
    CallSignal? signal,
  }) {
    final result = CallSignalRequest._();
    if (accountId != null) result.accountId = accountId;
    if (callId != null) result.callId = callId;
    if (signal != null) result.signal = signal;
    return result;
  }

  CallSignalRequest._();

  factory CallSignalRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CallSignalRequest()..mergeFromBuffer(data, registry);
  factory CallSignalRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CallSignalRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CallSignalRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: CallSignalRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'callId')
    ..aOM<CallSignal>(3, _omitFieldNames ? '' : 'signal',
        subBuilder: CallSignal.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CallSignalRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CallSignalRequest copyWith(void Function(CallSignalRequest) updates) =>
      super.copyWith((message) => updates(message as CallSignalRequest))
          as CallSignalRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CallSignalRequest() / CallSignalRequest.new instead')
  static CallSignalRequest create() => CallSignalRequest._();
  static $pb.GeneratedMessage $_createMessage() => CallSignalRequest._();
  @$core.override
  CallSignalRequest createEmptyInstance() => CallSignalRequest._();
  @$core.pragma('dart2js:noInline')
  static CallSignalRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CallSignalRequest>(
          CallSignalRequest.$_createMessage);
  static CallSignalRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get callId => $_getSZ(1);
  @$pb.TagNumber(2)
  set callId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCallId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCallId() => $_clearField(2);

  @$pb.TagNumber(3)
  CallSignal get signal => $_getN(2);
  @$pb.TagNumber(3)
  set signal(CallSignal value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasSignal() => $_has(2);
  @$pb.TagNumber(3)
  void clearSignal() => $_clearField(3);
  @$pb.TagNumber(3)
  CallSignal ensureSignal() => $_ensure(2);
}

class CallMediaRequest extends $pb.GeneratedMessage {
  factory CallMediaRequest({
    $core.String? accountId,
    $core.String? callId,
    CallMedia? media,
  }) {
    final result = CallMediaRequest._();
    if (accountId != null) result.accountId = accountId;
    if (callId != null) result.callId = callId;
    if (media != null) result.media = media;
    return result;
  }

  CallMediaRequest._();

  factory CallMediaRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CallMediaRequest()..mergeFromBuffer(data, registry);
  factory CallMediaRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CallMediaRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CallMediaRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: CallMediaRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'callId')
    ..aOM<CallMedia>(3, _omitFieldNames ? '' : 'media',
        subBuilder: CallMedia.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CallMediaRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CallMediaRequest copyWith(void Function(CallMediaRequest) updates) =>
      super.copyWith((message) => updates(message as CallMediaRequest))
          as CallMediaRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CallMediaRequest() / CallMediaRequest.new instead')
  static CallMediaRequest create() => CallMediaRequest._();
  static $pb.GeneratedMessage $_createMessage() => CallMediaRequest._();
  @$core.override
  CallMediaRequest createEmptyInstance() => CallMediaRequest._();
  @$core.pragma('dart2js:noInline')
  static CallMediaRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CallMediaRequest>(
          CallMediaRequest.$_createMessage);
  static CallMediaRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get callId => $_getSZ(1);
  @$pb.TagNumber(2)
  set callId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCallId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCallId() => $_clearField(2);

  @$pb.TagNumber(3)
  CallMedia get media => $_getN(2);
  @$pb.TagNumber(3)
  set media(CallMedia value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasMedia() => $_has(2);
  @$pb.TagNumber(3)
  void clearMedia() => $_clearField(3);
  @$pb.TagNumber(3)
  CallMedia ensureMedia() => $_ensure(2);
}

class CallsReply extends $pb.GeneratedMessage {
  factory CallsReply({
    $core.Iterable<Call>? calls,
  }) {
    final result = CallsReply._();
    if (calls != null) result.calls.addAll(calls);
    return result;
  }

  CallsReply._();

  factory CallsReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CallsReply()..mergeFromBuffer(data, registry);
  factory CallsReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CallsReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CallsReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: CallsReply.$_createMessage)
    ..pPM<Call>(1, _omitFieldNames ? '' : 'calls',
        subBuilder: Call.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CallsReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CallsReply copyWith(void Function(CallsReply) updates) =>
      super.copyWith((message) => updates(message as CallsReply)) as CallsReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CallsReply() / CallsReply.new instead')
  static CallsReply create() => CallsReply._();
  static $pb.GeneratedMessage $_createMessage() => CallsReply._();
  @$core.override
  CallsReply createEmptyInstance() => CallsReply._();
  @$core.pragma('dart2js:noInline')
  static CallsReply getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CallsReply>(CallsReply.$_createMessage);
  static CallsReply? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Call> get calls => $_getList(0);
}

class CallSignalsRequest extends $pb.GeneratedMessage {
  factory CallSignalsRequest({
    $core.String? accountId,
    $core.String? callId,
    $fixnum.Int64? after,
  }) {
    final result = CallSignalsRequest._();
    if (accountId != null) result.accountId = accountId;
    if (callId != null) result.callId = callId;
    if (after != null) result.after = after;
    return result;
  }

  CallSignalsRequest._();

  factory CallSignalsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CallSignalsRequest()..mergeFromBuffer(data, registry);
  factory CallSignalsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CallSignalsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CallSignalsRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: CallSignalsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'callId')
    ..aInt64(3, _omitFieldNames ? '' : 'after')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CallSignalsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CallSignalsRequest copyWith(void Function(CallSignalsRequest) updates) =>
      super.copyWith((message) => updates(message as CallSignalsRequest))
          as CallSignalsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CallSignalsRequest() / CallSignalsRequest.new instead')
  static CallSignalsRequest create() => CallSignalsRequest._();
  static $pb.GeneratedMessage $_createMessage() => CallSignalsRequest._();
  @$core.override
  CallSignalsRequest createEmptyInstance() => CallSignalsRequest._();
  @$core.pragma('dart2js:noInline')
  static CallSignalsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CallSignalsRequest>(
          CallSignalsRequest.$_createMessage);
  static CallSignalsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get callId => $_getSZ(1);
  @$pb.TagNumber(2)
  set callId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCallId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCallId() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get after => $_getI64(2);
  @$pb.TagNumber(3)
  set after($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAfter() => $_has(2);
  @$pb.TagNumber(3)
  void clearAfter() => $_clearField(3);
}

class CallSignalsReply extends $pb.GeneratedMessage {
  factory CallSignalsReply({
    $core.Iterable<CallSignal>? signals,
  }) {
    final result = CallSignalsReply._();
    if (signals != null) result.signals.addAll(signals);
    return result;
  }

  CallSignalsReply._();

  factory CallSignalsReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CallSignalsReply()..mergeFromBuffer(data, registry);
  factory CallSignalsReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CallSignalsReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CallSignalsReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: CallSignalsReply.$_createMessage)
    ..pPM<CallSignal>(1, _omitFieldNames ? '' : 'signals',
        subBuilder: CallSignal.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CallSignalsReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CallSignalsReply copyWith(void Function(CallSignalsReply) updates) =>
      super.copyWith((message) => updates(message as CallSignalsReply))
          as CallSignalsReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CallSignalsReply() / CallSignalsReply.new instead')
  static CallSignalsReply create() => CallSignalsReply._();
  static $pb.GeneratedMessage $_createMessage() => CallSignalsReply._();
  @$core.override
  CallSignalsReply createEmptyInstance() => CallSignalsReply._();
  @$core.pragma('dart2js:noInline')
  static CallSignalsReply getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CallSignalsReply>(
          CallSignalsReply.$_createMessage);
  static CallSignalsReply? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<CallSignal> get signals => $_getList(0);
}

class AccountsRequest extends $pb.GeneratedMessage {
  factory AccountsRequest({
    $core.Iterable<$core.String>? accountIds,
  }) {
    final result = AccountsRequest._();
    if (accountIds != null) result.accountIds.addAll(accountIds);
    return result;
  }

  AccountsRequest._();

  factory AccountsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AccountsRequest()..mergeFromBuffer(data, registry);
  factory AccountsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AccountsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AccountsRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: AccountsRequest.$_createMessage)
    ..pPS(1, _omitFieldNames ? '' : 'accountIds')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AccountsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AccountsRequest copyWith(void Function(AccountsRequest) updates) =>
      super.copyWith((message) => updates(message as AccountsRequest))
          as AccountsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AccountsRequest() / AccountsRequest.new instead')
  static AccountsRequest create() => AccountsRequest._();
  static $pb.GeneratedMessage $_createMessage() => AccountsRequest._();
  @$core.override
  AccountsRequest createEmptyInstance() => AccountsRequest._();
  @$core.pragma('dart2js:noInline')
  static AccountsRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountsRequest>(
          AccountsRequest.$_createMessage);
  static AccountsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$core.String> get accountIds => $_getList(0);
}

/// A failed part of a merged read.
class AccountError extends $pb.GeneratedMessage {
  factory AccountError({
    $core.String? accountId,
    $core.String? error,
  }) {
    final result = AccountError._();
    if (accountId != null) result.accountId = accountId;
    if (error != null) result.error = error;
    return result;
  }

  AccountError._();

  factory AccountError.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AccountError()..mergeFromBuffer(data, registry);
  factory AccountError.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AccountError()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AccountError',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: AccountError.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'error')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AccountError clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AccountError copyWith(void Function(AccountError) updates) =>
      super.copyWith((message) => updates(message as AccountError))
          as AccountError;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AccountError() / AccountError.new instead')
  static AccountError create() => AccountError._();
  static $pb.GeneratedMessage $_createMessage() => AccountError._();
  @$core.override
  AccountError createEmptyInstance() => AccountError._();
  @$core.pragma('dart2js:noInline')
  static AccountError getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AccountError>(
          AccountError.$_createMessage);
  static AccountError? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get error => $_getSZ(1);
  @$pb.TagNumber(2)
  set error($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasError() => $_has(1);
  @$pb.TagNumber(2)
  void clearError() => $_clearField(2);
}

class Folder extends $pb.GeneratedMessage {
  factory Folder({
    $core.String? accountId,
    $core.String? id,
    $core.String? name,
    FolderRole? role,
    $core.int? unread,
    $core.int? total,
  }) {
    final result = Folder._();
    if (accountId != null) result.accountId = accountId;
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (role != null) result.role = role;
    if (unread != null) result.unread = unread;
    if (total != null) result.total = total;
    return result;
  }

  Folder._();

  factory Folder.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Folder()..mergeFromBuffer(data, registry);
  factory Folder.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Folder()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Folder',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Folder.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'id')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aE<FolderRole>(4, _omitFieldNames ? '' : 'role',
        enumValues: FolderRole.values)
    ..aI(5, _omitFieldNames ? '' : 'unread')
    ..aI(6, _omitFieldNames ? '' : 'total')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Folder clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Folder copyWith(void Function(Folder) updates) =>
      super.copyWith((message) => updates(message as Folder)) as Folder;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Folder() / Folder.new instead')
  static Folder create() => Folder._();
  static $pb.GeneratedMessage $_createMessage() => Folder._();
  @$core.override
  Folder createEmptyInstance() => Folder._();
  @$core.pragma('dart2js:noInline')
  static Folder getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Folder>(Folder.$_createMessage);
  static Folder? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get id => $_getSZ(1);
  @$pb.TagNumber(2)
  set id($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasId() => $_has(1);
  @$pb.TagNumber(2)
  void clearId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get name => $_getSZ(2);
  @$pb.TagNumber(3)
  set name($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearName() => $_clearField(3);

  @$pb.TagNumber(4)
  FolderRole get role => $_getN(3);
  @$pb.TagNumber(4)
  set role(FolderRole value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasRole() => $_has(3);
  @$pb.TagNumber(4)
  void clearRole() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.int get unread => $_getIZ(4);
  @$pb.TagNumber(5)
  set unread($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasUnread() => $_has(4);
  @$pb.TagNumber(5)
  void clearUnread() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.int get total => $_getIZ(5);
  @$pb.TagNumber(6)
  set total($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasTotal() => $_has(5);
  @$pb.TagNumber(6)
  void clearTotal() => $_clearField(6);
}

class FoldersReply extends $pb.GeneratedMessage {
  factory FoldersReply({
    $core.Iterable<Folder>? folders,
    $core.Iterable<AccountError>? failed,
  }) {
    final result = FoldersReply._();
    if (folders != null) result.folders.addAll(folders);
    if (failed != null) result.failed.addAll(failed);
    return result;
  }

  FoldersReply._();

  factory FoldersReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      FoldersReply()..mergeFromBuffer(data, registry);
  factory FoldersReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      FoldersReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'FoldersReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: FoldersReply.$_createMessage)
    ..pPM<Folder>(1, _omitFieldNames ? '' : 'folders',
        subBuilder: Folder.$_createMessage)
    ..pPM<AccountError>(2, _omitFieldNames ? '' : 'failed',
        subBuilder: AccountError.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FoldersReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FoldersReply copyWith(void Function(FoldersReply) updates) =>
      super.copyWith((message) => updates(message as FoldersReply))
          as FoldersReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use FoldersReply() / FoldersReply.new instead')
  static FoldersReply create() => FoldersReply._();
  static $pb.GeneratedMessage $_createMessage() => FoldersReply._();
  @$core.override
  FoldersReply createEmptyInstance() => FoldersReply._();
  @$core.pragma('dart2js:noInline')
  static FoldersReply getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<FoldersReply>(
          FoldersReply.$_createMessage);
  static FoldersReply? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Folder> get folders => $_getList(0);

  @$pb.TagNumber(2)
  $pb.PbList<AccountError> get failed => $_getList(1);
}

class Address extends $pb.GeneratedMessage {
  factory Address({
    $core.String? name,
    $core.String? email,
  }) {
    final result = Address._();
    if (name != null) result.name = name;
    if (email != null) result.email = email;
    return result;
  }

  Address._();

  factory Address.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Address()..mergeFromBuffer(data, registry);
  factory Address.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Address()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Address',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Address.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'email')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Address clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Address copyWith(void Function(Address) updates) =>
      super.copyWith((message) => updates(message as Address)) as Address;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Address() / Address.new instead')
  static Address create() => Address._();
  static $pb.GeneratedMessage $_createMessage() => Address._();
  @$core.override
  Address createEmptyInstance() => Address._();
  @$core.pragma('dart2js:noInline')
  static Address getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Address>(Address.$_createMessage);
  static Address? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get email => $_getSZ(1);
  @$pb.TagNumber(2)
  set email($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEmail() => $_has(1);
  @$pb.TagNumber(2)
  void clearEmail() => $_clearField(2);
}

class Mail extends $pb.GeneratedMessage {
  factory Mail({
    $core.String? accountId,
    $core.String? id,
    $core.String? folderId,
    $core.String? threadId,
    Address? from,
    $core.Iterable<Address>? to,
    $core.Iterable<Address>? cc,
    $core.Iterable<Address>? replyTo,
    $core.String? subject,
    $core.String? snippet,
    $fixnum.Int64? timeMs,
    $core.bool? unread,
    $core.bool? starred,
    $core.bool? hasAttachments,
    $core.String? text,
    $core.String? html,
    $core.Iterable<Attachment>? attachments,
    $core.Iterable<TextSpan>? format,
  }) {
    final result = Mail._();
    if (accountId != null) result.accountId = accountId;
    if (id != null) result.id = id;
    if (folderId != null) result.folderId = folderId;
    if (threadId != null) result.threadId = threadId;
    if (from != null) result.from = from;
    if (to != null) result.to.addAll(to);
    if (cc != null) result.cc.addAll(cc);
    if (replyTo != null) result.replyTo.addAll(replyTo);
    if (subject != null) result.subject = subject;
    if (snippet != null) result.snippet = snippet;
    if (timeMs != null) result.timeMs = timeMs;
    if (unread != null) result.unread = unread;
    if (starred != null) result.starred = starred;
    if (hasAttachments != null) result.hasAttachments = hasAttachments;
    if (text != null) result.text = text;
    if (html != null) result.html = html;
    if (attachments != null) result.attachments.addAll(attachments);
    if (format != null) result.format.addAll(format);
    return result;
  }

  Mail._();

  factory Mail.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Mail()..mergeFromBuffer(data, registry);
  factory Mail.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Mail()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Mail',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Mail.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'id')
    ..aOS(3, _omitFieldNames ? '' : 'folderId')
    ..aOS(4, _omitFieldNames ? '' : 'threadId')
    ..aOM<Address>(5, _omitFieldNames ? '' : 'from',
        subBuilder: Address.$_createMessage)
    ..pPM<Address>(6, _omitFieldNames ? '' : 'to',
        subBuilder: Address.$_createMessage)
    ..pPM<Address>(7, _omitFieldNames ? '' : 'cc',
        subBuilder: Address.$_createMessage)
    ..pPM<Address>(8, _omitFieldNames ? '' : 'replyTo',
        subBuilder: Address.$_createMessage)
    ..aOS(9, _omitFieldNames ? '' : 'subject')
    ..aOS(10, _omitFieldNames ? '' : 'snippet')
    ..aInt64(11, _omitFieldNames ? '' : 'timeMs')
    ..aOB(12, _omitFieldNames ? '' : 'unread')
    ..aOB(13, _omitFieldNames ? '' : 'starred')
    ..aOB(14, _omitFieldNames ? '' : 'hasAttachments')
    ..aOS(15, _omitFieldNames ? '' : 'text')
    ..aOS(16, _omitFieldNames ? '' : 'html')
    ..pPM<Attachment>(17, _omitFieldNames ? '' : 'attachments',
        subBuilder: Attachment.$_createMessage)
    ..pPM<TextSpan>(18, _omitFieldNames ? '' : 'format',
        subBuilder: TextSpan.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Mail clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Mail copyWith(void Function(Mail) updates) =>
      super.copyWith((message) => updates(message as Mail)) as Mail;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Mail() / Mail.new instead')
  static Mail create() => Mail._();
  static $pb.GeneratedMessage $_createMessage() => Mail._();
  @$core.override
  Mail createEmptyInstance() => Mail._();
  @$core.pragma('dart2js:noInline')
  static Mail getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Mail>(Mail.$_createMessage);
  static Mail? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get id => $_getSZ(1);
  @$pb.TagNumber(2)
  set id($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasId() => $_has(1);
  @$pb.TagNumber(2)
  void clearId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get folderId => $_getSZ(2);
  @$pb.TagNumber(3)
  set folderId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasFolderId() => $_has(2);
  @$pb.TagNumber(3)
  void clearFolderId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get threadId => $_getSZ(3);
  @$pb.TagNumber(4)
  set threadId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasThreadId() => $_has(3);
  @$pb.TagNumber(4)
  void clearThreadId() => $_clearField(4);

  @$pb.TagNumber(5)
  Address get from => $_getN(4);
  @$pb.TagNumber(5)
  set from(Address value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasFrom() => $_has(4);
  @$pb.TagNumber(5)
  void clearFrom() => $_clearField(5);
  @$pb.TagNumber(5)
  Address ensureFrom() => $_ensure(4);

  @$pb.TagNumber(6)
  $pb.PbList<Address> get to => $_getList(5);

  @$pb.TagNumber(7)
  $pb.PbList<Address> get cc => $_getList(6);

  @$pb.TagNumber(8)
  $pb.PbList<Address> get replyTo => $_getList(7);

  @$pb.TagNumber(9)
  $core.String get subject => $_getSZ(8);
  @$pb.TagNumber(9)
  set subject($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasSubject() => $_has(8);
  @$pb.TagNumber(9)
  void clearSubject() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.String get snippet => $_getSZ(9);
  @$pb.TagNumber(10)
  set snippet($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasSnippet() => $_has(9);
  @$pb.TagNumber(10)
  void clearSnippet() => $_clearField(10);

  @$pb.TagNumber(11)
  $fixnum.Int64 get timeMs => $_getI64(10);
  @$pb.TagNumber(11)
  set timeMs($fixnum.Int64 value) => $_setInt64(10, value);
  @$pb.TagNumber(11)
  $core.bool hasTimeMs() => $_has(10);
  @$pb.TagNumber(11)
  void clearTimeMs() => $_clearField(11);

  @$pb.TagNumber(12)
  $core.bool get unread => $_getBF(11);
  @$pb.TagNumber(12)
  set unread($core.bool value) => $_setBool(11, value);
  @$pb.TagNumber(12)
  $core.bool hasUnread() => $_has(11);
  @$pb.TagNumber(12)
  void clearUnread() => $_clearField(12);

  @$pb.TagNumber(13)
  $core.bool get starred => $_getBF(12);
  @$pb.TagNumber(13)
  set starred($core.bool value) => $_setBool(12, value);
  @$pb.TagNumber(13)
  $core.bool hasStarred() => $_has(12);
  @$pb.TagNumber(13)
  void clearStarred() => $_clearField(13);

  @$pb.TagNumber(14)
  $core.bool get hasAttachments => $_getBF(13);
  @$pb.TagNumber(14)
  set hasAttachments($core.bool value) => $_setBool(13, value);
  @$pb.TagNumber(14)
  $core.bool hasHasAttachments() => $_has(13);
  @$pb.TagNumber(14)
  void clearHasAttachments() => $_clearField(14);

  /// GetMail only. text is the visible text (of the HTML part when there is
  /// one) and format its styles; html is the original HTML part, if any.
  /// Remote content in it is not loaded by the core.
  @$pb.TagNumber(15)
  $core.String get text => $_getSZ(14);
  @$pb.TagNumber(15)
  set text($core.String value) => $_setString(14, value);
  @$pb.TagNumber(15)
  $core.bool hasText() => $_has(14);
  @$pb.TagNumber(15)
  void clearText() => $_clearField(15);

  @$pb.TagNumber(16)
  $core.String get html => $_getSZ(15);
  @$pb.TagNumber(16)
  set html($core.String value) => $_setString(15, value);
  @$pb.TagNumber(16)
  $core.bool hasHtml() => $_has(15);
  @$pb.TagNumber(16)
  void clearHtml() => $_clearField(16);

  @$pb.TagNumber(17)
  $pb.PbList<Attachment> get attachments => $_getList(16);

  @$pb.TagNumber(18)
  $pb.PbList<TextSpan> get format => $_getList(17);
}

class MailListRequest extends $pb.GeneratedMessage {
  factory MailListRequest({
    $core.Iterable<$core.String>? accountIds,
    FolderRole? role,
    $core.String? folderId,
    $core.String? cursor,
    $core.int? limit,
  }) {
    final result = MailListRequest._();
    if (accountIds != null) result.accountIds.addAll(accountIds);
    if (role != null) result.role = role;
    if (folderId != null) result.folderId = folderId;
    if (cursor != null) result.cursor = cursor;
    if (limit != null) result.limit = limit;
    return result;
  }

  MailListRequest._();

  factory MailListRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MailListRequest()..mergeFromBuffer(data, registry);
  factory MailListRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MailListRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MailListRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: MailListRequest.$_createMessage)
    ..pPS(1, _omitFieldNames ? '' : 'accountIds')
    ..aE<FolderRole>(2, _omitFieldNames ? '' : 'role',
        enumValues: FolderRole.values)
    ..aOS(3, _omitFieldNames ? '' : 'folderId')
    ..aOS(4, _omitFieldNames ? '' : 'cursor')
    ..aI(5, _omitFieldNames ? '' : 'limit')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MailListRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MailListRequest copyWith(void Function(MailListRequest) updates) =>
      super.copyWith((message) => updates(message as MailListRequest))
          as MailListRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MailListRequest() / MailListRequest.new instead')
  static MailListRequest create() => MailListRequest._();
  static $pb.GeneratedMessage $_createMessage() => MailListRequest._();
  @$core.override
  MailListRequest createEmptyInstance() => MailListRequest._();
  @$core.pragma('dart2js:noInline')
  static MailListRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MailListRequest>(
          MailListRequest.$_createMessage);
  static MailListRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$core.String> get accountIds => $_getList(0);

  /// The folder of each account with this role (all inboxes); or, with
  /// exactly one account, the folder folder_id.
  @$pb.TagNumber(2)
  FolderRole get role => $_getN(1);
  @$pb.TagNumber(2)
  set role(FolderRole value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasRole() => $_has(1);
  @$pb.TagNumber(2)
  void clearRole() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get folderId => $_getSZ(2);
  @$pb.TagNumber(3)
  set folderId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasFolderId() => $_has(2);
  @$pb.TagNumber(3)
  void clearFolderId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get cursor => $_getSZ(3);
  @$pb.TagNumber(4)
  set cursor($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCursor() => $_has(3);
  @$pb.TagNumber(4)
  void clearCursor() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.int get limit => $_getIZ(4);
  @$pb.TagNumber(5)
  set limit($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasLimit() => $_has(4);
  @$pb.TagNumber(5)
  void clearLimit() => $_clearField(5);
}

class MailListReply extends $pb.GeneratedMessage {
  factory MailListReply({
    $core.Iterable<Mail>? mails,
    $core.bool? more,
    $core.String? next,
    $core.Iterable<AccountError>? failed,
  }) {
    final result = MailListReply._();
    if (mails != null) result.mails.addAll(mails);
    if (more != null) result.more = more;
    if (next != null) result.next = next;
    if (failed != null) result.failed.addAll(failed);
    return result;
  }

  MailListReply._();

  factory MailListReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MailListReply()..mergeFromBuffer(data, registry);
  factory MailListReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MailListReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MailListReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: MailListReply.$_createMessage)
    ..pPM<Mail>(1, _omitFieldNames ? '' : 'mails',
        subBuilder: Mail.$_createMessage)
    ..aOB(2, _omitFieldNames ? '' : 'more')
    ..aOS(3, _omitFieldNames ? '' : 'next')
    ..pPM<AccountError>(4, _omitFieldNames ? '' : 'failed',
        subBuilder: AccountError.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MailListReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MailListReply copyWith(void Function(MailListReply) updates) =>
      super.copyWith((message) => updates(message as MailListReply))
          as MailListReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MailListReply() / MailListReply.new instead')
  static MailListReply create() => MailListReply._();
  static $pb.GeneratedMessage $_createMessage() => MailListReply._();
  @$core.override
  MailListReply createEmptyInstance() => MailListReply._();
  @$core.pragma('dart2js:noInline')
  static MailListReply getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MailListReply>(
          MailListReply.$_createMessage);
  static MailListReply? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Mail> get mails => $_getList(0);

  @$pb.TagNumber(2)
  $core.bool get more => $_getBF(1);
  @$pb.TagNumber(2)
  set more($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMore() => $_has(1);
  @$pb.TagNumber(2)
  void clearMore() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get next => $_getSZ(2);
  @$pb.TagNumber(3)
  set next($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasNext() => $_has(2);
  @$pb.TagNumber(3)
  void clearNext() => $_clearField(3);

  @$pb.TagNumber(4)
  $pb.PbList<AccountError> get failed => $_getList(3);
}

class GetMailRequest extends $pb.GeneratedMessage {
  factory GetMailRequest({
    $core.String? accountId,
    $core.String? id,
  }) {
    final result = GetMailRequest._();
    if (accountId != null) result.accountId = accountId;
    if (id != null) result.id = id;
    return result;
  }

  GetMailRequest._();

  factory GetMailRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMailRequest()..mergeFromBuffer(data, registry);
  factory GetMailRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMailRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetMailRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: GetMailRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMailRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMailRequest copyWith(void Function(GetMailRequest) updates) =>
      super.copyWith((message) => updates(message as GetMailRequest))
          as GetMailRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetMailRequest() / GetMailRequest.new instead')
  static GetMailRequest create() => GetMailRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetMailRequest._();
  @$core.override
  GetMailRequest createEmptyInstance() => GetMailRequest._();
  @$core.pragma('dart2js:noInline')
  static GetMailRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetMailRequest>(
          GetMailRequest.$_createMessage);
  static GetMailRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get id => $_getSZ(1);
  @$pb.TagNumber(2)
  set id($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasId() => $_has(1);
  @$pb.TagNumber(2)
  void clearId() => $_clearField(2);
}

class SendMailRequest extends $pb.GeneratedMessage {
  factory SendMailRequest({
    $core.String? accountId,
    $core.Iterable<Address>? to,
    $core.Iterable<Address>? cc,
    $core.Iterable<Address>? bcc,
    $core.String? subject,
    $core.String? text,
    $core.String? replyToId,
  }) {
    final result = SendMailRequest._();
    if (accountId != null) result.accountId = accountId;
    if (to != null) result.to.addAll(to);
    if (cc != null) result.cc.addAll(cc);
    if (bcc != null) result.bcc.addAll(bcc);
    if (subject != null) result.subject = subject;
    if (text != null) result.text = text;
    if (replyToId != null) result.replyToId = replyToId;
    return result;
  }

  SendMailRequest._();

  factory SendMailRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SendMailRequest()..mergeFromBuffer(data, registry);
  factory SendMailRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SendMailRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SendMailRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: SendMailRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..pPM<Address>(2, _omitFieldNames ? '' : 'to',
        subBuilder: Address.$_createMessage)
    ..pPM<Address>(3, _omitFieldNames ? '' : 'cc',
        subBuilder: Address.$_createMessage)
    ..pPM<Address>(4, _omitFieldNames ? '' : 'bcc',
        subBuilder: Address.$_createMessage)
    ..aOS(5, _omitFieldNames ? '' : 'subject')
    ..aOS(6, _omitFieldNames ? '' : 'text')
    ..aOS(7, _omitFieldNames ? '' : 'replyToId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SendMailRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SendMailRequest copyWith(void Function(SendMailRequest) updates) =>
      super.copyWith((message) => updates(message as SendMailRequest))
          as SendMailRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use SendMailRequest() / SendMailRequest.new instead')
  static SendMailRequest create() => SendMailRequest._();
  static $pb.GeneratedMessage $_createMessage() => SendMailRequest._();
  @$core.override
  SendMailRequest createEmptyInstance() => SendMailRequest._();
  @$core.pragma('dart2js:noInline')
  static SendMailRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SendMailRequest>(
          SendMailRequest.$_createMessage);
  static SendMailRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<Address> get to => $_getList(1);

  @$pb.TagNumber(3)
  $pb.PbList<Address> get cc => $_getList(2);

  @$pb.TagNumber(4)
  $pb.PbList<Address> get bcc => $_getList(3);

  @$pb.TagNumber(5)
  $core.String get subject => $_getSZ(4);
  @$pb.TagNumber(5)
  set subject($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSubject() => $_has(4);
  @$pb.TagNumber(5)
  void clearSubject() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get text => $_getSZ(5);
  @$pb.TagNumber(6)
  set text($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasText() => $_has(5);
  @$pb.TagNumber(6)
  void clearText() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get replyToId => $_getSZ(6);
  @$pb.TagNumber(7)
  set replyToId($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasReplyToId() => $_has(6);
  @$pb.TagNumber(7)
  void clearReplyToId() => $_clearField(7);
}

class UpdateMailRequest extends $pb.GeneratedMessage {
  factory UpdateMailRequest({
    $core.String? accountId,
    $core.Iterable<$core.String>? ids,
    MailAction? action,
  }) {
    final result = UpdateMailRequest._();
    if (accountId != null) result.accountId = accountId;
    if (ids != null) result.ids.addAll(ids);
    if (action != null) result.action = action;
    return result;
  }

  UpdateMailRequest._();

  factory UpdateMailRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateMailRequest()..mergeFromBuffer(data, registry);
  factory UpdateMailRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateMailRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateMailRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: UpdateMailRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..pPS(2, _omitFieldNames ? '' : 'ids')
    ..aE<MailAction>(3, _omitFieldNames ? '' : 'action',
        enumValues: MailAction.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateMailRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateMailRequest copyWith(void Function(UpdateMailRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateMailRequest))
          as UpdateMailRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use UpdateMailRequest() / UpdateMailRequest.new instead')
  static UpdateMailRequest create() => UpdateMailRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateMailRequest._();
  @$core.override
  UpdateMailRequest createEmptyInstance() => UpdateMailRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateMailRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UpdateMailRequest>(
          UpdateMailRequest.$_createMessage);
  static UpdateMailRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get ids => $_getList(1);

  @$pb.TagNumber(3)
  MailAction get action => $_getN(2);
  @$pb.TagNumber(3)
  set action(MailAction value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasAction() => $_has(2);
  @$pb.TagNumber(3)
  void clearAction() => $_clearField(3);
}

class MoveMailRequest extends $pb.GeneratedMessage {
  factory MoveMailRequest({
    $core.String? accountId,
    $core.Iterable<$core.String>? ids,
    $core.String? fromFolderId,
    $core.String? toFolderId,
  }) {
    final result = MoveMailRequest._();
    if (accountId != null) result.accountId = accountId;
    if (ids != null) result.ids.addAll(ids);
    if (fromFolderId != null) result.fromFolderId = fromFolderId;
    if (toFolderId != null) result.toFolderId = toFolderId;
    return result;
  }

  MoveMailRequest._();

  factory MoveMailRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MoveMailRequest()..mergeFromBuffer(data, registry);
  factory MoveMailRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MoveMailRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MoveMailRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: MoveMailRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..pPS(2, _omitFieldNames ? '' : 'ids')
    ..aOS(3, _omitFieldNames ? '' : 'fromFolderId')
    ..aOS(4, _omitFieldNames ? '' : 'toFolderId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MoveMailRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MoveMailRequest copyWith(void Function(MoveMailRequest) updates) =>
      super.copyWith((message) => updates(message as MoveMailRequest))
          as MoveMailRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MoveMailRequest() / MoveMailRequest.new instead')
  static MoveMailRequest create() => MoveMailRequest._();
  static $pb.GeneratedMessage $_createMessage() => MoveMailRequest._();
  @$core.override
  MoveMailRequest createEmptyInstance() => MoveMailRequest._();
  @$core.pragma('dart2js:noInline')
  static MoveMailRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MoveMailRequest>(
          MoveMailRequest.$_createMessage);
  static MoveMailRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get ids => $_getList(1);

  @$pb.TagNumber(3)
  $core.String get fromFolderId => $_getSZ(2);
  @$pb.TagNumber(3)
  set fromFolderId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasFromFolderId() => $_has(2);
  @$pb.TagNumber(3)
  void clearFromFolderId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get toFolderId => $_getSZ(3);
  @$pb.TagNumber(4)
  set toFolderId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasToFolderId() => $_has(3);
  @$pb.TagNumber(4)
  void clearToFolderId() => $_clearField(4);
}

class Calendar extends $pb.GeneratedMessage {
  factory Calendar({
    $core.String? accountId,
    $core.String? id,
    $core.String? name,
    $core.int? color,
    $core.bool? readOnly,
    $core.bool? primary,
  }) {
    final result = Calendar._();
    if (accountId != null) result.accountId = accountId;
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (color != null) result.color = color;
    if (readOnly != null) result.readOnly = readOnly;
    if (primary != null) result.primary = primary;
    return result;
  }

  Calendar._();

  factory Calendar.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Calendar()..mergeFromBuffer(data, registry);
  factory Calendar.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Calendar()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Calendar',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Calendar.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'id')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aI(4, _omitFieldNames ? '' : 'color')
    ..aOB(5, _omitFieldNames ? '' : 'readOnly')
    ..aOB(6, _omitFieldNames ? '' : 'primary')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Calendar clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Calendar copyWith(void Function(Calendar) updates) =>
      super.copyWith((message) => updates(message as Calendar)) as Calendar;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Calendar() / Calendar.new instead')
  static Calendar create() => Calendar._();
  static $pb.GeneratedMessage $_createMessage() => Calendar._();
  @$core.override
  Calendar createEmptyInstance() => Calendar._();
  @$core.pragma('dart2js:noInline')
  static Calendar getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Calendar>(Calendar.$_createMessage);
  static Calendar? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get id => $_getSZ(1);
  @$pb.TagNumber(2)
  set id($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasId() => $_has(1);
  @$pb.TagNumber(2)
  void clearId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get name => $_getSZ(2);
  @$pb.TagNumber(3)
  set name($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get color => $_getIZ(3);
  @$pb.TagNumber(4)
  set color($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasColor() => $_has(3);
  @$pb.TagNumber(4)
  void clearColor() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get readOnly => $_getBF(4);
  @$pb.TagNumber(5)
  set readOnly($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasReadOnly() => $_has(4);
  @$pb.TagNumber(5)
  void clearReadOnly() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.bool get primary => $_getBF(5);
  @$pb.TagNumber(6)
  set primary($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPrimary() => $_has(5);
  @$pb.TagNumber(6)
  void clearPrimary() => $_clearField(6);
}

class CalendarsReply extends $pb.GeneratedMessage {
  factory CalendarsReply({
    $core.Iterable<Calendar>? calendars,
    $core.Iterable<AccountError>? failed,
  }) {
    final result = CalendarsReply._();
    if (calendars != null) result.calendars.addAll(calendars);
    if (failed != null) result.failed.addAll(failed);
    return result;
  }

  CalendarsReply._();

  factory CalendarsReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CalendarsReply()..mergeFromBuffer(data, registry);
  factory CalendarsReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CalendarsReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CalendarsReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: CalendarsReply.$_createMessage)
    ..pPM<Calendar>(1, _omitFieldNames ? '' : 'calendars',
        subBuilder: Calendar.$_createMessage)
    ..pPM<AccountError>(2, _omitFieldNames ? '' : 'failed',
        subBuilder: AccountError.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CalendarsReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CalendarsReply copyWith(void Function(CalendarsReply) updates) =>
      super.copyWith((message) => updates(message as CalendarsReply))
          as CalendarsReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CalendarsReply() / CalendarsReply.new instead')
  static CalendarsReply create() => CalendarsReply._();
  static $pb.GeneratedMessage $_createMessage() => CalendarsReply._();
  @$core.override
  CalendarsReply createEmptyInstance() => CalendarsReply._();
  @$core.pragma('dart2js:noInline')
  static CalendarsReply getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CalendarsReply>(
          CalendarsReply.$_createMessage);
  static CalendarsReply? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Calendar> get calendars => $_getList(0);

  @$pb.TagNumber(2)
  $pb.PbList<AccountError> get failed => $_getList(1);
}

class Attendee extends $pb.GeneratedMessage {
  factory Attendee({
    Address? address,
    Rsvp? status,
    $core.bool? self,
  }) {
    final result = Attendee._();
    if (address != null) result.address = address;
    if (status != null) result.status = status;
    if (self != null) result.self = self;
    return result;
  }

  Attendee._();

  factory Attendee.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Attendee()..mergeFromBuffer(data, registry);
  factory Attendee.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Attendee()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Attendee',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: Attendee.$_createMessage)
    ..aOM<Address>(1, _omitFieldNames ? '' : 'address',
        subBuilder: Address.$_createMessage)
    ..aE<Rsvp>(2, _omitFieldNames ? '' : 'status', enumValues: Rsvp.values)
    ..aOB(3, _omitFieldNames ? '' : 'self')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Attendee clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Attendee copyWith(void Function(Attendee) updates) =>
      super.copyWith((message) => updates(message as Attendee)) as Attendee;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Attendee() / Attendee.new instead')
  static Attendee create() => Attendee._();
  static $pb.GeneratedMessage $_createMessage() => Attendee._();
  @$core.override
  Attendee createEmptyInstance() => Attendee._();
  @$core.pragma('dart2js:noInline')
  static Attendee getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Attendee>(Attendee.$_createMessage);
  static Attendee? _defaultInstance;

  @$pb.TagNumber(1)
  Address get address => $_getN(0);
  @$pb.TagNumber(1)
  set address(Address value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearAddress() => $_clearField(1);
  @$pb.TagNumber(1)
  Address ensureAddress() => $_ensure(0);

  @$pb.TagNumber(2)
  Rsvp get status => $_getN(1);
  @$pb.TagNumber(2)
  set status(Rsvp value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasStatus() => $_has(1);
  @$pb.TagNumber(2)
  void clearStatus() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get self => $_getBF(2);
  @$pb.TagNumber(3)
  set self($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSelf() => $_has(2);
  @$pb.TagNumber(3)
  void clearSelf() => $_clearField(3);
}

class CalendarEvent extends $pb.GeneratedMessage {
  factory CalendarEvent({
    $core.String? accountId,
    $core.String? calendarId,
    $core.String? id,
    $core.String? title,
    $core.String? location,
    $core.String? description,
    $fixnum.Int64? startMs,
    $fixnum.Int64? endMs,
    $core.bool? allDay,
    $core.bool? recurring,
    Address? organizer,
    $core.Iterable<Attendee>? attendees,
    $core.String? link,
    $core.String? recurrence,
    RecurrenceScope? scope,
  }) {
    final result = CalendarEvent._();
    if (accountId != null) result.accountId = accountId;
    if (calendarId != null) result.calendarId = calendarId;
    if (id != null) result.id = id;
    if (title != null) result.title = title;
    if (location != null) result.location = location;
    if (description != null) result.description = description;
    if (startMs != null) result.startMs = startMs;
    if (endMs != null) result.endMs = endMs;
    if (allDay != null) result.allDay = allDay;
    if (recurring != null) result.recurring = recurring;
    if (organizer != null) result.organizer = organizer;
    if (attendees != null) result.attendees.addAll(attendees);
    if (link != null) result.link = link;
    if (recurrence != null) result.recurrence = recurrence;
    if (scope != null) result.scope = scope;
    return result;
  }

  CalendarEvent._();

  factory CalendarEvent.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CalendarEvent()..mergeFromBuffer(data, registry);
  factory CalendarEvent.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CalendarEvent()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CalendarEvent',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: CalendarEvent.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'calendarId')
    ..aOS(3, _omitFieldNames ? '' : 'id')
    ..aOS(4, _omitFieldNames ? '' : 'title')
    ..aOS(5, _omitFieldNames ? '' : 'location')
    ..aOS(6, _omitFieldNames ? '' : 'description')
    ..aInt64(7, _omitFieldNames ? '' : 'startMs')
    ..aInt64(8, _omitFieldNames ? '' : 'endMs')
    ..aOB(9, _omitFieldNames ? '' : 'allDay')
    ..aOB(10, _omitFieldNames ? '' : 'recurring')
    ..aOM<Address>(11, _omitFieldNames ? '' : 'organizer',
        subBuilder: Address.$_createMessage)
    ..pPM<Attendee>(12, _omitFieldNames ? '' : 'attendees',
        subBuilder: Attendee.$_createMessage)
    ..aOS(13, _omitFieldNames ? '' : 'link')
    ..aOS(14, _omitFieldNames ? '' : 'recurrence')
    ..aE<RecurrenceScope>(15, _omitFieldNames ? '' : 'scope',
        enumValues: RecurrenceScope.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CalendarEvent clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CalendarEvent copyWith(void Function(CalendarEvent) updates) =>
      super.copyWith((message) => updates(message as CalendarEvent))
          as CalendarEvent;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CalendarEvent() / CalendarEvent.new instead')
  static CalendarEvent create() => CalendarEvent._();
  static $pb.GeneratedMessage $_createMessage() => CalendarEvent._();
  @$core.override
  CalendarEvent createEmptyInstance() => CalendarEvent._();
  @$core.pragma('dart2js:noInline')
  static CalendarEvent getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CalendarEvent>(
          CalendarEvent.$_createMessage);
  static CalendarEvent? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get calendarId => $_getSZ(1);
  @$pb.TagNumber(2)
  set calendarId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCalendarId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCalendarId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get id => $_getSZ(2);
  @$pb.TagNumber(3)
  set id($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasId() => $_has(2);
  @$pb.TagNumber(3)
  void clearId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get title => $_getSZ(3);
  @$pb.TagNumber(4)
  set title($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasTitle() => $_has(3);
  @$pb.TagNumber(4)
  void clearTitle() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get location => $_getSZ(4);
  @$pb.TagNumber(5)
  set location($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasLocation() => $_has(4);
  @$pb.TagNumber(5)
  void clearLocation() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get description => $_getSZ(5);
  @$pb.TagNumber(6)
  set description($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasDescription() => $_has(5);
  @$pb.TagNumber(6)
  void clearDescription() => $_clearField(6);

  /// End is exclusive. All-day events: midnights UTC of their dates.
  @$pb.TagNumber(7)
  $fixnum.Int64 get startMs => $_getI64(6);
  @$pb.TagNumber(7)
  set startMs($fixnum.Int64 value) => $_setInt64(6, value);
  @$pb.TagNumber(7)
  $core.bool hasStartMs() => $_has(6);
  @$pb.TagNumber(7)
  void clearStartMs() => $_clearField(7);

  @$pb.TagNumber(8)
  $fixnum.Int64 get endMs => $_getI64(7);
  @$pb.TagNumber(8)
  set endMs($fixnum.Int64 value) => $_setInt64(7, value);
  @$pb.TagNumber(8)
  $core.bool hasEndMs() => $_has(7);
  @$pb.TagNumber(8)
  void clearEndMs() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.bool get allDay => $_getBF(8);
  @$pb.TagNumber(9)
  set allDay($core.bool value) => $_setBool(8, value);
  @$pb.TagNumber(9)
  $core.bool hasAllDay() => $_has(8);
  @$pb.TagNumber(9)
  void clearAllDay() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.bool get recurring => $_getBF(9);
  @$pb.TagNumber(10)
  set recurring($core.bool value) => $_setBool(9, value);
  @$pb.TagNumber(10)
  $core.bool hasRecurring() => $_has(9);
  @$pb.TagNumber(10)
  void clearRecurring() => $_clearField(10);

  @$pb.TagNumber(11)
  Address get organizer => $_getN(10);
  @$pb.TagNumber(11)
  set organizer(Address value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasOrganizer() => $_has(10);
  @$pb.TagNumber(11)
  void clearOrganizer() => $_clearField(11);
  @$pb.TagNumber(11)
  Address ensureOrganizer() => $_ensure(10);

  @$pb.TagNumber(12)
  $pb.PbList<Attendee> get attendees => $_getList(11);

  @$pb.TagNumber(13)
  $core.String get link => $_getSZ(12);
  @$pb.TagNumber(13)
  set link($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasLink() => $_has(12);
  @$pb.TagNumber(13)
  void clearLink() => $_clearField(13);

  /// The repeating event's rule: an RFC 5545 RRULE value without "RRULE:"
  /// ("FREQ=WEEKLY;BYDAY=MO"); empty: it does not repeat. Occurrences carry
  /// their series' rule. In SaveEvent it sets the rule of a new event, of an
  /// event that did not repeat, of the new series (scope FOLLOWING; the old
  /// rule goes on with what was left of a COUNT), or of the series (scope
  /// ALL); empty keeps the series' rule.
  @$pb.TagNumber(14)
  $core.String get recurrence => $_getSZ(13);
  @$pb.TagNumber(14)
  set recurrence($core.String value) => $_setString(13, value);
  @$pb.TagNumber(14)
  $core.bool hasRecurrence() => $_has(13);
  @$pb.TagNumber(14)
  void clearRecurrence() => $_clearField(14);

  /// SaveEvent only: what of a repeating event an occurrence's change touches.
  @$pb.TagNumber(15)
  RecurrenceScope get scope => $_getN(14);
  @$pb.TagNumber(15)
  set scope(RecurrenceScope value) => $_setField(15, value);
  @$pb.TagNumber(15)
  $core.bool hasScope() => $_has(14);
  @$pb.TagNumber(15)
  void clearScope() => $_clearField(15);
}

class CalendarRef extends $pb.GeneratedMessage {
  factory CalendarRef({
    $core.String? accountId,
    $core.String? calendarId,
  }) {
    final result = CalendarRef._();
    if (accountId != null) result.accountId = accountId;
    if (calendarId != null) result.calendarId = calendarId;
    return result;
  }

  CalendarRef._();

  factory CalendarRef.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CalendarRef()..mergeFromBuffer(data, registry);
  factory CalendarRef.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CalendarRef()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CalendarRef',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: CalendarRef.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'calendarId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CalendarRef clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CalendarRef copyWith(void Function(CalendarRef) updates) =>
      super.copyWith((message) => updates(message as CalendarRef))
          as CalendarRef;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CalendarRef() / CalendarRef.new instead')
  static CalendarRef create() => CalendarRef._();
  static $pb.GeneratedMessage $_createMessage() => CalendarRef._();
  @$core.override
  CalendarRef createEmptyInstance() => CalendarRef._();
  @$core.pragma('dart2js:noInline')
  static CalendarRef getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CalendarRef>(
          CalendarRef.$_createMessage);
  static CalendarRef? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get calendarId => $_getSZ(1);
  @$pb.TagNumber(2)
  set calendarId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCalendarId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCalendarId() => $_clearField(2);
}

class EventsRequest extends $pb.GeneratedMessage {
  factory EventsRequest({
    $core.Iterable<$core.String>? accountIds,
    $core.Iterable<CalendarRef>? calendars,
    $fixnum.Int64? fromMs,
    $fixnum.Int64? toMs,
  }) {
    final result = EventsRequest._();
    if (accountIds != null) result.accountIds.addAll(accountIds);
    if (calendars != null) result.calendars.addAll(calendars);
    if (fromMs != null) result.fromMs = fromMs;
    if (toMs != null) result.toMs = toMs;
    return result;
  }

  EventsRequest._();

  factory EventsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      EventsRequest()..mergeFromBuffer(data, registry);
  factory EventsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      EventsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EventsRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: EventsRequest.$_createMessage)
    ..pPS(1, _omitFieldNames ? '' : 'accountIds')
    ..pPM<CalendarRef>(2, _omitFieldNames ? '' : 'calendars',
        subBuilder: CalendarRef.$_createMessage)
    ..aInt64(3, _omitFieldNames ? '' : 'fromMs')
    ..aInt64(4, _omitFieldNames ? '' : 'toMs')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EventsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EventsRequest copyWith(void Function(EventsRequest) updates) =>
      super.copyWith((message) => updates(message as EventsRequest))
          as EventsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use EventsRequest() / EventsRequest.new instead')
  static EventsRequest create() => EventsRequest._();
  static $pb.GeneratedMessage $_createMessage() => EventsRequest._();
  @$core.override
  EventsRequest createEmptyInstance() => EventsRequest._();
  @$core.pragma('dart2js:noInline')
  static EventsRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<EventsRequest>(
          EventsRequest.$_createMessage);
  static EventsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$core.String> get accountIds => $_getList(0);

  @$pb.TagNumber(2)
  $pb.PbList<CalendarRef> get calendars => $_getList(1);

  @$pb.TagNumber(3)
  $fixnum.Int64 get fromMs => $_getI64(2);
  @$pb.TagNumber(3)
  set fromMs($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasFromMs() => $_has(2);
  @$pb.TagNumber(3)
  void clearFromMs() => $_clearField(3);

  @$pb.TagNumber(4)
  $fixnum.Int64 get toMs => $_getI64(3);
  @$pb.TagNumber(4)
  set toMs($fixnum.Int64 value) => $_setInt64(3, value);
  @$pb.TagNumber(4)
  $core.bool hasToMs() => $_has(3);
  @$pb.TagNumber(4)
  void clearToMs() => $_clearField(4);
}

class EventsReply extends $pb.GeneratedMessage {
  factory EventsReply({
    $core.Iterable<CalendarEvent>? events,
    $core.Iterable<AccountError>? failed,
  }) {
    final result = EventsReply._();
    if (events != null) result.events.addAll(events);
    if (failed != null) result.failed.addAll(failed);
    return result;
  }

  EventsReply._();

  factory EventsReply.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      EventsReply()..mergeFromBuffer(data, registry);
  factory EventsReply.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      EventsReply()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EventsReply',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: EventsReply.$_createMessage)
    ..pPM<CalendarEvent>(1, _omitFieldNames ? '' : 'events',
        subBuilder: CalendarEvent.$_createMessage)
    ..pPM<AccountError>(2, _omitFieldNames ? '' : 'failed',
        subBuilder: AccountError.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EventsReply clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EventsReply copyWith(void Function(EventsReply) updates) =>
      super.copyWith((message) => updates(message as EventsReply))
          as EventsReply;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use EventsReply() / EventsReply.new instead')
  static EventsReply create() => EventsReply._();
  static $pb.GeneratedMessage $_createMessage() => EventsReply._();
  @$core.override
  EventsReply createEmptyInstance() => EventsReply._();
  @$core.pragma('dart2js:noInline')
  static EventsReply getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<EventsReply>(
          EventsReply.$_createMessage);
  static EventsReply? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<CalendarEvent> get events => $_getList(0);

  @$pb.TagNumber(2)
  $pb.PbList<AccountError> get failed => $_getList(1);
}

class DeleteEventRequest extends $pb.GeneratedMessage {
  factory DeleteEventRequest({
    $core.String? accountId,
    $core.String? calendarId,
    $core.String? id,
    RecurrenceScope? scope,
  }) {
    final result = DeleteEventRequest._();
    if (accountId != null) result.accountId = accountId;
    if (calendarId != null) result.calendarId = calendarId;
    if (id != null) result.id = id;
    if (scope != null) result.scope = scope;
    return result;
  }

  DeleteEventRequest._();

  factory DeleteEventRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteEventRequest()..mergeFromBuffer(data, registry);
  factory DeleteEventRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteEventRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteEventRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'selfhold.parley.v1'),
      createEmptyInstance: DeleteEventRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..aOS(2, _omitFieldNames ? '' : 'calendarId')
    ..aOS(3, _omitFieldNames ? '' : 'id')
    ..aE<RecurrenceScope>(4, _omitFieldNames ? '' : 'scope',
        enumValues: RecurrenceScope.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteEventRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteEventRequest copyWith(void Function(DeleteEventRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteEventRequest))
          as DeleteEventRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use DeleteEventRequest() / DeleteEventRequest.new instead')
  static DeleteEventRequest create() => DeleteEventRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteEventRequest._();
  @$core.override
  DeleteEventRequest createEmptyInstance() => DeleteEventRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteEventRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteEventRequest>(
          DeleteEventRequest.$_createMessage);
  static DeleteEventRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get calendarId => $_getSZ(1);
  @$pb.TagNumber(2)
  set calendarId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCalendarId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCalendarId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get id => $_getSZ(2);
  @$pb.TagNumber(3)
  set id($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasId() => $_has(2);
  @$pb.TagNumber(3)
  void clearId() => $_clearField(3);

  @$pb.TagNumber(4)
  RecurrenceScope get scope => $_getN(3);
  @$pb.TagNumber(4)
  set scope(RecurrenceScope value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasScope() => $_has(3);
  @$pb.TagNumber(4)
  void clearScope() => $_clearField(4);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
