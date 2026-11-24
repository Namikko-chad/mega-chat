// This is a generated file - do not edit.
//
// Generated from megachat.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports
// ignore_for_file: unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use capabilityDescriptor instead')
const Capability$json = {
  '1': 'Capability',
  '2': [
    {'1': 'CAPABILITY_UNSPECIFIED', '2': 0},
    {'1': 'CAPABILITY_MESSAGES', '2': 1},
    {'1': 'CAPABILITY_CALLS', '2': 2},
    {'1': 'CAPABILITY_GROUP_CALLS', '2': 3},
    {'1': 'CAPABILITY_MAIL', '2': 4},
    {'1': 'CAPABILITY_CALENDAR', '2': 5},
    {'1': 'CAPABILITY_THREADS', '2': 6},
    {'1': 'CAPABILITY_REACTIONS', '2': 7},
    {'1': 'CAPABILITY_MEDIA', '2': 8},
    {'1': 'CAPABILITY_TOPICS', '2': 9},
    {'1': 'CAPABILITY_FORWARD', '2': 10},
    {'1': 'CAPABILITY_FILES', '2': 11},
    {'1': 'CAPABILITY_PROFILES', '2': 12},
    {'1': 'CAPABILITY_EDIT', '2': 13},
    {'1': 'CAPABILITY_DELETE', '2': 14},
    {'1': 'CAPABILITY_DELETE_FOR_ME', '2': 15},
    {'1': 'CAPABILITY_MENTIONS', '2': 16},
    {'1': 'CAPABILITY_CHAT_FOLDERS', '2': 17},
    {'1': 'CAPABILITY_MANAGE_CHATS', '2': 18},
  ],
};

/// Descriptor for `Capability`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List capabilityDescriptor = $convert.base64Decode(
    'CgpDYXBhYmlsaXR5EhoKFkNBUEFCSUxJVFlfVU5TUEVDSUZJRUQQABIXChNDQVBBQklMSVRZX0'
    '1FU1NBR0VTEAESFAoQQ0FQQUJJTElUWV9DQUxMUxACEhoKFkNBUEFCSUxJVFlfR1JPVVBfQ0FM'
    'TFMQAxITCg9DQVBBQklMSVRZX01BSUwQBBIXChNDQVBBQklMSVRZX0NBTEVOREFSEAUSFgoSQ0'
    'FQQUJJTElUWV9USFJFQURTEAYSGAoUQ0FQQUJJTElUWV9SRUFDVElPTlMQBxIUChBDQVBBQklM'
    'SVRZX01FRElBEAgSFQoRQ0FQQUJJTElUWV9UT1BJQ1MQCRIWChJDQVBBQklMSVRZX0ZPUldBUk'
    'QQChIUChBDQVBBQklMSVRZX0ZJTEVTEAsSFwoTQ0FQQUJJTElUWV9QUk9GSUxFUxAMEhMKD0NB'
    'UEFCSUxJVFlfRURJVBANEhUKEUNBUEFCSUxJVFlfREVMRVRFEA4SHAoYQ0FQQUJJTElUWV9ERU'
    'xFVEVfRk9SX01FEA8SFwoTQ0FQQUJJTElUWV9NRU5USU9OUxAQEhsKF0NBUEFCSUxJVFlfQ0hB'
    'VF9GT0xERVJTEBESGwoXQ0FQQUJJTElUWV9NQU5BR0VfQ0hBVFMQEg==');

@$core.Deprecated('Use chatActionDescriptor instead')
const ChatAction$json = {
  '1': 'ChatAction',
  '2': [
    {'1': 'CHAT_ACTION_UNSPECIFIED', '2': 0},
    {'1': 'CHAT_ACTION_MUTE', '2': 1},
    {'1': 'CHAT_ACTION_UNMUTE', '2': 2},
    {'1': 'CHAT_ACTION_PIN', '2': 3},
    {'1': 'CHAT_ACTION_UNPIN', '2': 4},
    {'1': 'CHAT_ACTION_ARCHIVE', '2': 5},
    {'1': 'CHAT_ACTION_UNARCHIVE', '2': 6},
    {'1': 'CHAT_ACTION_LEAVE', '2': 7},
  ],
};

/// Descriptor for `ChatAction`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List chatActionDescriptor = $convert.base64Decode(
    'CgpDaGF0QWN0aW9uEhsKF0NIQVRfQUNUSU9OX1VOU1BFQ0lGSUVEEAASFAoQQ0hBVF9BQ1RJT0'
    '5fTVVURRABEhYKEkNIQVRfQUNUSU9OX1VOTVVURRACEhMKD0NIQVRfQUNUSU9OX1BJThADEhUK'
    'EUNIQVRfQUNUSU9OX1VOUElOEAQSFwoTQ0hBVF9BQ1RJT05fQVJDSElWRRAFEhkKFUNIQVRfQU'
    'NUSU9OX1VOQVJDSElWRRAGEhUKEUNIQVRfQUNUSU9OX0xFQVZFEAc=');

@$core.Deprecated('Use accountStatusDescriptor instead')
const AccountStatus$json = {
  '1': 'AccountStatus',
  '2': [
    {'1': 'ACCOUNT_STATUS_UNSPECIFIED', '2': 0},
    {'1': 'ACCOUNT_STATUS_DISABLED', '2': 1},
    {'1': 'ACCOUNT_STATUS_CONNECTING', '2': 2},
    {'1': 'ACCOUNT_STATUS_ONLINE', '2': 3},
    {'1': 'ACCOUNT_STATUS_OFFLINE', '2': 4},
    {'1': 'ACCOUNT_STATUS_NEEDS_LOGIN', '2': 5},
  ],
};

/// Descriptor for `AccountStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List accountStatusDescriptor = $convert.base64Decode(
    'Cg1BY2NvdW50U3RhdHVzEh4KGkFDQ09VTlRfU1RBVFVTX1VOU1BFQ0lGSUVEEAASGwoXQUNDT1'
    'VOVF9TVEFUVVNfRElTQUJMRUQQARIdChlBQ0NPVU5UX1NUQVRVU19DT05ORUNUSU5HEAISGQoV'
    'QUNDT1VOVF9TVEFUVVNfT05MSU5FEAMSGgoWQUNDT1VOVF9TVEFUVVNfT0ZGTElORRAEEh4KGk'
    'FDQ09VTlRfU1RBVFVTX05FRURTX0xPR0lOEAU=');

@$core.Deprecated('Use chatKindDescriptor instead')
const ChatKind$json = {
  '1': 'ChatKind',
  '2': [
    {'1': 'CHAT_KIND_UNSPECIFIED', '2': 0},
    {'1': 'CHAT_KIND_DIRECT', '2': 1},
    {'1': 'CHAT_KIND_GROUP', '2': 2},
    {'1': 'CHAT_KIND_CHANNEL', '2': 3},
  ],
};

/// Descriptor for `ChatKind`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List chatKindDescriptor = $convert.base64Decode(
    'CghDaGF0S2luZBIZChVDSEFUX0tJTkRfVU5TUEVDSUZJRUQQABIUChBDSEFUX0tJTkRfRElSRU'
    'NUEAESEwoPQ0hBVF9LSU5EX0dST1VQEAISFQoRQ0hBVF9LSU5EX0NIQU5ORUwQAw==');

@$core.Deprecated('Use attachmentKindDescriptor instead')
const AttachmentKind$json = {
  '1': 'AttachmentKind',
  '2': [
    {'1': 'ATTACHMENT_KIND_UNSPECIFIED', '2': 0},
    {'1': 'ATTACHMENT_KIND_FILE', '2': 1},
    {'1': 'ATTACHMENT_KIND_IMAGE', '2': 2},
    {'1': 'ATTACHMENT_KIND_VIDEO', '2': 3},
    {'1': 'ATTACHMENT_KIND_AUDIO', '2': 4},
    {'1': 'ATTACHMENT_KIND_VOICE', '2': 5},
    {'1': 'ATTACHMENT_KIND_STICKER', '2': 6},
  ],
};

/// Descriptor for `AttachmentKind`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List attachmentKindDescriptor = $convert.base64Decode(
    'Cg5BdHRhY2htZW50S2luZBIfChtBVFRBQ0hNRU5UX0tJTkRfVU5TUEVDSUZJRUQQABIYChRBVF'
    'RBQ0hNRU5UX0tJTkRfRklMRRABEhkKFUFUVEFDSE1FTlRfS0lORF9JTUFHRRACEhkKFUFUVEFD'
    'SE1FTlRfS0lORF9WSURFTxADEhkKFUFUVEFDSE1FTlRfS0lORF9BVURJTxAEEhkKFUFUVEFDSE'
    '1FTlRfS0lORF9WT0lDRRAFEhsKF0FUVEFDSE1FTlRfS0lORF9TVElDS0VSEAY=');

@$core.Deprecated('Use textStyleDescriptor instead')
const TextStyle$json = {
  '1': 'TextStyle',
  '2': [
    {'1': 'TEXT_STYLE_UNSPECIFIED', '2': 0},
    {'1': 'TEXT_STYLE_BOLD', '2': 1},
    {'1': 'TEXT_STYLE_ITALIC', '2': 2},
    {'1': 'TEXT_STYLE_UNDERLINE', '2': 3},
    {'1': 'TEXT_STYLE_STRIKE', '2': 4},
    {'1': 'TEXT_STYLE_SPOILER', '2': 5},
    {'1': 'TEXT_STYLE_CODE', '2': 6},
    {'1': 'TEXT_STYLE_PRE', '2': 7},
    {'1': 'TEXT_STYLE_LINK', '2': 8},
    {'1': 'TEXT_STYLE_QUOTE', '2': 9},
    {'1': 'TEXT_STYLE_HEADING', '2': 10},
  ],
};

/// Descriptor for `TextStyle`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List textStyleDescriptor = $convert.base64Decode(
    'CglUZXh0U3R5bGUSGgoWVEVYVF9TVFlMRV9VTlNQRUNJRklFRBAAEhMKD1RFWFRfU1RZTEVfQk'
    '9MRBABEhUKEVRFWFRfU1RZTEVfSVRBTElDEAISGAoUVEVYVF9TVFlMRV9VTkRFUkxJTkUQAxIV'
    'ChFURVhUX1NUWUxFX1NUUklLRRAEEhYKElRFWFRfU1RZTEVfU1BPSUxFUhAFEhMKD1RFWFRfU1'
    'RZTEVfQ09ERRAGEhIKDlRFWFRfU1RZTEVfUFJFEAcSEwoPVEVYVF9TVFlMRV9MSU5LEAgSFAoQ'
    'VEVYVF9TVFlMRV9RVU9URRAJEhYKElRFWFRfU1RZTEVfSEVBRElORxAK');

@$core.Deprecated('Use callStateDescriptor instead')
const CallState$json = {
  '1': 'CallState',
  '2': [
    {'1': 'CALL_STATE_UNSPECIFIED', '2': 0},
    {'1': 'CALL_STATE_WAITING', '2': 1},
    {'1': 'CALL_STATE_RINGING', '2': 2},
    {'1': 'CALL_STATE_CONNECTING', '2': 3},
    {'1': 'CALL_STATE_ACTIVE', '2': 4},
    {'1': 'CALL_STATE_ENDED', '2': 5},
  ],
};

/// Descriptor for `CallState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List callStateDescriptor = $convert.base64Decode(
    'CglDYWxsU3RhdGUSGgoWQ0FMTF9TVEFURV9VTlNQRUNJRklFRBAAEhYKEkNBTExfU1RBVEVfV0'
    'FJVElORxABEhYKEkNBTExfU1RBVEVfUklOR0lORxACEhkKFUNBTExfU1RBVEVfQ09OTkVDVElO'
    'RxADEhUKEUNBTExfU1RBVEVfQUNUSVZFEAQSFAoQQ0FMTF9TVEFURV9FTkRFRBAF');

@$core.Deprecated('Use callEndDescriptor instead')
const CallEnd$json = {
  '1': 'CallEnd',
  '2': [
    {'1': 'CALL_END_UNSPECIFIED', '2': 0},
    {'1': 'CALL_END_HANGUP', '2': 1},
    {'1': 'CALL_END_DECLINED', '2': 2},
    {'1': 'CALL_END_MISSED', '2': 3},
    {'1': 'CALL_END_BUSY', '2': 4},
    {'1': 'CALL_END_FAILED', '2': 5},
  ],
};

/// Descriptor for `CallEnd`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List callEndDescriptor = $convert.base64Decode(
    'CgdDYWxsRW5kEhgKFENBTExfRU5EX1VOU1BFQ0lGSUVEEAASEwoPQ0FMTF9FTkRfSEFOR1VQEA'
    'ESFQoRQ0FMTF9FTkRfREVDTElORUQQAhITCg9DQUxMX0VORF9NSVNTRUQQAxIRCg1DQUxMX0VO'
    'RF9CVVNZEAQSEwoPQ0FMTF9FTkRfRkFJTEVEEAU=');

@$core.Deprecated('Use signalKindDescriptor instead')
const SignalKind$json = {
  '1': 'SignalKind',
  '2': [
    {'1': 'SIGNAL_KIND_UNSPECIFIED', '2': 0},
    {'1': 'SIGNAL_KIND_OFFER', '2': 1},
    {'1': 'SIGNAL_KIND_ANSWER', '2': 2},
    {'1': 'SIGNAL_KIND_CANDIDATE', '2': 3},
    {'1': 'SIGNAL_KIND_DATA', '2': 4},
    {'1': 'SIGNAL_KIND_CONNECTED', '2': 5},
    {'1': 'SIGNAL_KIND_FAILED', '2': 6},
  ],
};

/// Descriptor for `SignalKind`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List signalKindDescriptor = $convert.base64Decode(
    'CgpTaWduYWxLaW5kEhsKF1NJR05BTF9LSU5EX1VOU1BFQ0lGSUVEEAASFQoRU0lHTkFMX0tJTk'
    'RfT0ZGRVIQARIWChJTSUdOQUxfS0lORF9BTlNXRVIQAhIZChVTSUdOQUxfS0lORF9DQU5ESURB'
    'VEUQAxIUChBTSUdOQUxfS0lORF9EQVRBEAQSGQoVU0lHTkFMX0tJTkRfQ09OTkVDVEVEEAUSFg'
    'oSU0lHTkFMX0tJTkRfRkFJTEVEEAY=');

@$core.Deprecated('Use folderRoleDescriptor instead')
const FolderRole$json = {
  '1': 'FolderRole',
  '2': [
    {'1': 'FOLDER_ROLE_OTHER', '2': 0},
    {'1': 'FOLDER_ROLE_INBOX', '2': 1},
    {'1': 'FOLDER_ROLE_SENT', '2': 2},
    {'1': 'FOLDER_ROLE_DRAFTS', '2': 3},
    {'1': 'FOLDER_ROLE_TRASH', '2': 4},
    {'1': 'FOLDER_ROLE_SPAM', '2': 5},
    {'1': 'FOLDER_ROLE_ARCHIVE', '2': 6},
  ],
};

/// Descriptor for `FolderRole`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List folderRoleDescriptor = $convert.base64Decode(
    'CgpGb2xkZXJSb2xlEhUKEUZPTERFUl9ST0xFX09USEVSEAASFQoRRk9MREVSX1JPTEVfSU5CT1'
    'gQARIUChBGT0xERVJfUk9MRV9TRU5UEAISFgoSRk9MREVSX1JPTEVfRFJBRlRTEAMSFQoRRk9M'
    'REVSX1JPTEVfVFJBU0gQBBIUChBGT0xERVJfUk9MRV9TUEFNEAUSFwoTRk9MREVSX1JPTEVfQV'
    'JDSElWRRAG');

@$core.Deprecated('Use mailActionDescriptor instead')
const MailAction$json = {
  '1': 'MailAction',
  '2': [
    {'1': 'MAIL_ACTION_UNSPECIFIED', '2': 0},
    {'1': 'MAIL_ACTION_READ', '2': 1},
    {'1': 'MAIL_ACTION_UNREAD', '2': 2},
    {'1': 'MAIL_ACTION_STAR', '2': 3},
    {'1': 'MAIL_ACTION_UNSTAR', '2': 4},
    {'1': 'MAIL_ACTION_TRASH', '2': 5},
  ],
};

/// Descriptor for `MailAction`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List mailActionDescriptor = $convert.base64Decode(
    'CgpNYWlsQWN0aW9uEhsKF01BSUxfQUNUSU9OX1VOU1BFQ0lGSUVEEAASFAoQTUFJTF9BQ1RJT0'
    '5fUkVBRBABEhYKEk1BSUxfQUNUSU9OX1VOUkVBRBACEhQKEE1BSUxfQUNUSU9OX1NUQVIQAxIW'
    'ChJNQUlMX0FDVElPTl9VTlNUQVIQBBIVChFNQUlMX0FDVElPTl9UUkFTSBAF');

@$core.Deprecated('Use rsvpDescriptor instead')
const Rsvp$json = {
  '1': 'Rsvp',
  '2': [
    {'1': 'RSVP_UNSPECIFIED', '2': 0},
    {'1': 'RSVP_NEEDS_ACTION', '2': 1},
    {'1': 'RSVP_ACCEPTED', '2': 2},
    {'1': 'RSVP_TENTATIVE', '2': 3},
    {'1': 'RSVP_DECLINED', '2': 4},
  ],
};

/// Descriptor for `Rsvp`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List rsvpDescriptor = $convert.base64Decode(
    'CgRSc3ZwEhQKEFJTVlBfVU5TUEVDSUZJRUQQABIVChFSU1ZQX05FRURTX0FDVElPThABEhEKDV'
    'JTVlBfQUNDRVBURUQQAhISCg5SU1ZQX1RFTlRBVElWRRADEhEKDVJTVlBfREVDTElORUQQBA==');

@$core.Deprecated('Use emptyDescriptor instead')
const Empty$json = {
  '1': 'Empty',
};

/// Descriptor for `Empty`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List emptyDescriptor =
    $convert.base64Decode('CgVFbXB0eQ==');

@$core.Deprecated('Use idRequestDescriptor instead')
const IdRequest$json = {
  '1': 'IdRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `IdRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List idRequestDescriptor =
    $convert.base64Decode('CglJZFJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use passwordRequestDescriptor instead')
const PasswordRequest$json = {
  '1': 'PasswordRequest',
  '2': [
    {'1': 'password', '3': 1, '4': 1, '5': 9, '10': 'password'},
  ],
};

/// Descriptor for `PasswordRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List passwordRequestDescriptor = $convert.base64Decode(
    'Cg9QYXNzd29yZFJlcXVlc3QSGgoIcGFzc3dvcmQYASABKAlSCHBhc3N3b3Jk');

@$core.Deprecated('Use vaultPathRequestDescriptor instead')
const VaultPathRequest$json = {
  '1': 'VaultPathRequest',
  '2': [
    {'1': 'path', '3': 1, '4': 1, '5': 9, '10': 'path'},
    {'1': 'move', '3': 2, '4': 1, '5': 8, '10': 'move'},
  ],
};

/// Descriptor for `VaultPathRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List vaultPathRequestDescriptor = $convert.base64Decode(
    'ChBWYXVsdFBhdGhSZXF1ZXN0EhIKBHBhdGgYASABKAlSBHBhdGgSEgoEbW92ZRgCIAEoCFIEbW'
    '92ZQ==');

@$core.Deprecated('Use changePasswordRequestDescriptor instead')
const ChangePasswordRequest$json = {
  '1': 'ChangePasswordRequest',
  '2': [
    {'1': 'old_password', '3': 1, '4': 1, '5': 9, '10': 'oldPassword'},
    {'1': 'new_password', '3': 2, '4': 1, '5': 9, '10': 'newPassword'},
  ],
};

/// Descriptor for `ChangePasswordRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List changePasswordRequestDescriptor = $convert.base64Decode(
    'ChVDaGFuZ2VQYXNzd29yZFJlcXVlc3QSIQoMb2xkX3Bhc3N3b3JkGAEgASgJUgtvbGRQYXNzd2'
    '9yZBIhCgxuZXdfcGFzc3dvcmQYAiABKAlSC25ld1Bhc3N3b3Jk');

@$core.Deprecated('Use stateDescriptor instead')
const State$json = {
  '1': 'State',
  '2': [
    {'1': 'vault_exists', '3': 1, '4': 1, '5': 8, '10': 'vaultExists'},
    {'1': 'unlocked', '3': 2, '4': 1, '5': 8, '10': 'unlocked'},
    {'1': 'vault_path', '3': 3, '4': 1, '5': 9, '10': 'vaultPath'},
    {'1': 'platform', '3': 4, '4': 1, '5': 9, '10': 'platform'},
    {'1': 'core_version', '3': 5, '4': 1, '5': 9, '10': 'coreVersion'},
    {'1': 'core_build', '3': 6, '4': 1, '5': 9, '10': 'coreBuild'},
    {'1': 'accounts', '3': 7, '4': 1, '5': 5, '10': 'accounts'},
    {'1': 'accounts_online', '3': 8, '4': 1, '5': 5, '10': 'accountsOnline'},
    {'1': 'can_remember', '3': 10, '4': 1, '5': 8, '10': 'canRemember'},
    {'1': 'remembered', '3': 11, '4': 1, '5': 8, '10': 'remembered'},
  ],
  '9': [
    {'1': 9, '2': 10},
  ],
};

/// Descriptor for `State`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List stateDescriptor = $convert.base64Decode(
    'CgVTdGF0ZRIhCgx2YXVsdF9leGlzdHMYASABKAhSC3ZhdWx0RXhpc3RzEhoKCHVubG9ja2VkGA'
    'IgASgIUgh1bmxvY2tlZBIdCgp2YXVsdF9wYXRoGAMgASgJUgl2YXVsdFBhdGgSGgoIcGxhdGZv'
    'cm0YBCABKAlSCHBsYXRmb3JtEiEKDGNvcmVfdmVyc2lvbhgFIAEoCVILY29yZVZlcnNpb24SHQ'
    'oKY29yZV9idWlsZBgGIAEoCVIJY29yZUJ1aWxkEhoKCGFjY291bnRzGAcgASgFUghhY2NvdW50'
    'cxInCg9hY2NvdW50c19vbmxpbmUYCCABKAVSDmFjY291bnRzT25saW5lEiEKDGNhbl9yZW1lbW'
    'JlchgKIAEoCFILY2FuUmVtZW1iZXISHgoKcmVtZW1iZXJlZBgLIAEoCFIKcmVtZW1iZXJlZEoE'
    'CAkQCg==');

@$core.Deprecated('Use fieldDescriptor instead')
const Field$json = {
  '1': 'Field',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'label', '3': 2, '4': 1, '5': 9, '10': 'label'},
    {'1': 'help', '3': 3, '4': 1, '5': 9, '10': 'help'},
    {'1': 'secret', '3': 4, '4': 1, '5': 8, '10': 'secret'},
    {'1': 'optional', '3': 5, '4': 1, '5': 8, '10': 'optional'},
    {'1': 'default_value', '3': 6, '4': 1, '5': 9, '10': 'defaultValue'},
    {'1': 'multiline', '3': 7, '4': 1, '5': 8, '10': 'multiline'},
  ],
};

/// Descriptor for `Field`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List fieldDescriptor = $convert.base64Decode(
    'CgVGaWVsZBIQCgNrZXkYASABKAlSA2tleRIUCgVsYWJlbBgCIAEoCVIFbGFiZWwSEgoEaGVscB'
    'gDIAEoCVIEaGVscBIWCgZzZWNyZXQYBCABKAhSBnNlY3JldBIaCghvcHRpb25hbBgFIAEoCFII'
    'b3B0aW9uYWwSIwoNZGVmYXVsdF92YWx1ZRgGIAEoCVIMZGVmYXVsdFZhbHVlEhwKCW11bHRpbG'
    'luZRgHIAEoCFIJbXVsdGlsaW5l');

@$core.Deprecated('Use authMethodDescriptor instead')
const AuthMethod$json = {
  '1': 'AuthMethod',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'description', '3': 3, '4': 1, '5': 9, '10': 'description'},
  ],
};

/// Descriptor for `AuthMethod`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List authMethodDescriptor = $convert.base64Decode(
    'CgpBdXRoTWV0aG9kEg4KAmlkGAEgASgJUgJpZBISCgRuYW1lGAIgASgJUgRuYW1lEiAKC2Rlc2'
    'NyaXB0aW9uGAMgASgJUgtkZXNjcmlwdGlvbg==');

@$core.Deprecated('Use providerInfoDescriptor instead')
const ProviderInfo$json = {
  '1': 'ProviderInfo',
  '2': [
    {'1': 'type', '3': 1, '4': 1, '5': 9, '10': 'type'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {
      '1': 'auth_methods',
      '3': 3,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.AuthMethod',
      '10': 'authMethods'
    },
    {
      '1': 'capabilities',
      '3': 4,
      '4': 3,
      '5': 14,
      '6': '.megachat.v1.Capability',
      '10': 'capabilities'
    },
  ],
};

/// Descriptor for `ProviderInfo`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List providerInfoDescriptor = $convert.base64Decode(
    'CgxQcm92aWRlckluZm8SEgoEdHlwZRgBIAEoCVIEdHlwZRISCgRuYW1lGAIgASgJUgRuYW1lEj'
    'oKDGF1dGhfbWV0aG9kcxgDIAMoCzIXLm1lZ2FjaGF0LnYxLkF1dGhNZXRob2RSC2F1dGhNZXRo'
    'b2RzEjsKDGNhcGFiaWxpdGllcxgEIAMoDjIXLm1lZ2FjaGF0LnYxLkNhcGFiaWxpdHlSDGNhcG'
    'FiaWxpdGllcw==');

@$core.Deprecated('Use providersReplyDescriptor instead')
const ProvidersReply$json = {
  '1': 'ProvidersReply',
  '2': [
    {
      '1': 'providers',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.ProviderInfo',
      '10': 'providers'
    },
  ],
};

/// Descriptor for `ProvidersReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List providersReplyDescriptor = $convert.base64Decode(
    'Cg5Qcm92aWRlcnNSZXBseRI3Cglwcm92aWRlcnMYASADKAsyGS5tZWdhY2hhdC52MS5Qcm92aW'
    'RlckluZm9SCXByb3ZpZGVycw==');

@$core.Deprecated('Use startLoginRequestDescriptor instead')
const StartLoginRequest$json = {
  '1': 'StartLoginRequest',
  '2': [
    {'1': 'provider', '3': 1, '4': 1, '5': 9, '10': 'provider'},
    {'1': 'method', '3': 2, '4': 1, '5': 9, '10': 'method'},
  ],
};

/// Descriptor for `StartLoginRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List startLoginRequestDescriptor = $convert.base64Decode(
    'ChFTdGFydExvZ2luUmVxdWVzdBIaCghwcm92aWRlchgBIAEoCVIIcHJvdmlkZXISFgoGbWV0aG'
    '9kGAIgASgJUgZtZXRob2Q=');

@$core.Deprecated('Use loginStepRequestDescriptor instead')
const LoginStepRequest$json = {
  '1': 'LoginStepRequest',
  '2': [
    {'1': 'login_id', '3': 1, '4': 1, '5': 9, '10': 'loginId'},
    {
      '1': 'values',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.LoginStepRequest.ValuesEntry',
      '10': 'values'
    },
  ],
  '3': [LoginStepRequest_ValuesEntry$json],
};

@$core.Deprecated('Use loginStepRequestDescriptor instead')
const LoginStepRequest_ValuesEntry$json = {
  '1': 'ValuesEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `LoginStepRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List loginStepRequestDescriptor = $convert.base64Decode(
    'ChBMb2dpblN0ZXBSZXF1ZXN0EhkKCGxvZ2luX2lkGAEgASgJUgdsb2dpbklkEkEKBnZhbHVlcx'
    'gCIAMoCzIpLm1lZ2FjaGF0LnYxLkxvZ2luU3RlcFJlcXVlc3QuVmFsdWVzRW50cnlSBnZhbHVl'
    'cxo5CgtWYWx1ZXNFbnRyeRIQCgNrZXkYASABKAlSA2tleRIUCgV2YWx1ZRgCIAEoCVIFdmFsdW'
    'U6AjgB');

@$core.Deprecated('Use loginStepReplyDescriptor instead')
const LoginStepReply$json = {
  '1': 'LoginStepReply',
  '2': [
    {'1': 'login_id', '3': 1, '4': 1, '5': 9, '10': 'loginId'},
    {'1': 'title', '3': 2, '4': 1, '5': 9, '10': 'title'},
    {'1': 'message', '3': 3, '4': 1, '5': 9, '10': 'message'},
    {'1': 'error', '3': 4, '4': 1, '5': 9, '10': 'error'},
    {
      '1': 'fields',
      '3': 5,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Field',
      '10': 'fields'
    },
    {'1': 'open_url', '3': 6, '4': 1, '5': 9, '10': 'openUrl'},
    {'1': 'done', '3': 7, '4': 1, '5': 8, '10': 'done'},
    {
      '1': 'account',
      '3': 8,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.Account',
      '10': 'account'
    },
    {'1': 'qr', '3': 9, '4': 1, '5': 9, '10': 'qr'},
  ],
};

/// Descriptor for `LoginStepReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List loginStepReplyDescriptor = $convert.base64Decode(
    'Cg5Mb2dpblN0ZXBSZXBseRIZCghsb2dpbl9pZBgBIAEoCVIHbG9naW5JZBIUCgV0aXRsZRgCIA'
    'EoCVIFdGl0bGUSGAoHbWVzc2FnZRgDIAEoCVIHbWVzc2FnZRIUCgVlcnJvchgEIAEoCVIFZXJy'
    'b3ISKgoGZmllbGRzGAUgAygLMhIubWVnYWNoYXQudjEuRmllbGRSBmZpZWxkcxIZCghvcGVuX3'
    'VybBgGIAEoCVIHb3BlblVybBISCgRkb25lGAcgASgIUgRkb25lEi4KB2FjY291bnQYCCABKAsy'
    'FC5tZWdhY2hhdC52MS5BY2NvdW50UgdhY2NvdW50Eg4KAnFyGAkgASgJUgJxcg==');

@$core.Deprecated('Use accountDescriptor instead')
const Account$json = {
  '1': 'Account',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'provider', '3': 2, '4': 1, '5': 9, '10': 'provider'},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '10': 'name'},
    {'1': 'login', '3': 4, '4': 1, '5': 9, '10': 'login'},
    {
      '1': 'status',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.megachat.v1.AccountStatus',
      '10': 'status'
    },
    {'1': 'status_message', '3': 6, '4': 1, '5': 9, '10': 'statusMessage'},
    {'1': 'enabled', '3': 7, '4': 1, '5': 8, '10': 'enabled'},
    {
      '1': 'self',
      '3': 8,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.User',
      '10': 'self'
    },
    {'1': 'unread', '3': 9, '4': 1, '5': 5, '10': 'unread'},
    {
      '1': 'capabilities',
      '3': 10,
      '4': 3,
      '5': 14,
      '6': '.megachat.v1.Capability',
      '10': 'capabilities'
    },
    {'1': 'mail_unread', '3': 11, '4': 1, '5': 5, '10': 'mailUnread'},
    {
      '1': 'chat_actions',
      '3': 12,
      '4': 3,
      '5': 14,
      '6': '.megachat.v1.ChatAction',
      '10': 'chatActions'
    },
  ],
};

/// Descriptor for `Account`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List accountDescriptor = $convert.base64Decode(
    'CgdBY2NvdW50Eg4KAmlkGAEgASgJUgJpZBIaCghwcm92aWRlchgCIAEoCVIIcHJvdmlkZXISEg'
    'oEbmFtZRgDIAEoCVIEbmFtZRIUCgVsb2dpbhgEIAEoCVIFbG9naW4SMgoGc3RhdHVzGAUgASgO'
    'MhoubWVnYWNoYXQudjEuQWNjb3VudFN0YXR1c1IGc3RhdHVzEiUKDnN0YXR1c19tZXNzYWdlGA'
    'YgASgJUg1zdGF0dXNNZXNzYWdlEhgKB2VuYWJsZWQYByABKAhSB2VuYWJsZWQSJQoEc2VsZhgI'
    'IAEoCzIRLm1lZ2FjaGF0LnYxLlVzZXJSBHNlbGYSFgoGdW5yZWFkGAkgASgFUgZ1bnJlYWQSOw'
    'oMY2FwYWJpbGl0aWVzGAogAygOMhcubWVnYWNoYXQudjEuQ2FwYWJpbGl0eVIMY2FwYWJpbGl0'
    'aWVzEh8KC21haWxfdW5yZWFkGAsgASgFUgptYWlsVW5yZWFkEjoKDGNoYXRfYWN0aW9ucxgMIA'
    'MoDjIXLm1lZ2FjaGF0LnYxLkNoYXRBY3Rpb25SC2NoYXRBY3Rpb25z');

@$core.Deprecated('Use accountsReplyDescriptor instead')
const AccountsReply$json = {
  '1': 'AccountsReply',
  '2': [
    {
      '1': 'accounts',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Account',
      '10': 'accounts'
    },
  ],
};

/// Descriptor for `AccountsReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List accountsReplyDescriptor = $convert.base64Decode(
    'Cg1BY2NvdW50c1JlcGx5EjAKCGFjY291bnRzGAEgAygLMhQubWVnYWNoYXQudjEuQWNjb3VudF'
    'IIYWNjb3VudHM=');

@$core.Deprecated('Use reorderAccountsRequestDescriptor instead')
const ReorderAccountsRequest$json = {
  '1': 'ReorderAccountsRequest',
  '2': [
    {'1': 'ids', '3': 1, '4': 3, '5': 9, '10': 'ids'},
  ],
};

/// Descriptor for `ReorderAccountsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List reorderAccountsRequestDescriptor = $convert
    .base64Decode('ChZSZW9yZGVyQWNjb3VudHNSZXF1ZXN0EhAKA2lkcxgBIAMoCVIDaWRz');

@$core.Deprecated('Use renameRequestDescriptor instead')
const RenameRequest$json = {
  '1': 'RenameRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `RenameRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List renameRequestDescriptor = $convert.base64Decode(
    'Cg1SZW5hbWVSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZBISCgRuYW1lGAIgASgJUgRuYW1l');

@$core.Deprecated('Use setEnabledRequestDescriptor instead')
const SetEnabledRequest$json = {
  '1': 'SetEnabledRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'enabled', '3': 2, '4': 1, '5': 8, '10': 'enabled'},
  ],
};

/// Descriptor for `SetEnabledRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setEnabledRequestDescriptor = $convert.base64Decode(
    'ChFTZXRFbmFibGVkUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQSGAoHZW5hYmxlZBgCIAEoCFIHZW'
    '5hYmxlZA==');

@$core.Deprecated('Use userDescriptor instead')
const User$json = {
  '1': 'User',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'username', '3': 3, '4': 1, '5': 9, '10': 'username'},
    {'1': 'avatar_url', '3': 4, '4': 1, '5': 9, '10': 'avatarUrl'},
    {'1': 'online', '3': 5, '4': 1, '5': 8, '10': 'online'},
  ],
};

/// Descriptor for `User`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List userDescriptor = $convert.base64Decode(
    'CgRVc2VyEg4KAmlkGAEgASgJUgJpZBISCgRuYW1lGAIgASgJUgRuYW1lEhoKCHVzZXJuYW1lGA'
    'MgASgJUgh1c2VybmFtZRIdCgphdmF0YXJfdXJsGAQgASgJUglhdmF0YXJVcmwSFgoGb25saW5l'
    'GAUgASgIUgZvbmxpbmU=');

@$core.Deprecated('Use chatDescriptor instead')
const Chat$json = {
  '1': 'Chat',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'id', '3': 2, '4': 1, '5': 9, '10': 'id'},
    {
      '1': 'kind',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.megachat.v1.ChatKind',
      '10': 'kind'
    },
    {'1': 'title', '3': 4, '4': 1, '5': 9, '10': 'title'},
    {'1': 'avatar_url', '3': 5, '4': 1, '5': 9, '10': 'avatarUrl'},
    {'1': 'unread', '3': 6, '4': 1, '5': 5, '10': 'unread'},
    {'1': 'muted', '3': 7, '4': 1, '5': 8, '10': 'muted'},
    {
      '1': 'last_message',
      '3': 8,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.Message',
      '10': 'lastMessage'
    },
    {'1': 'updated_ms', '3': 9, '4': 1, '5': 3, '10': 'updatedMs'},
    {'1': 'pinned', '3': 10, '4': 1, '5': 8, '10': 'pinned'},
    {'1': 'read_only', '3': 11, '4': 1, '5': 8, '10': 'readOnly'},
    {'1': 'forum', '3': 12, '4': 1, '5': 8, '10': 'forum'},
    {'1': 'online', '3': 13, '4': 1, '5': 8, '10': 'online'},
    {'1': 'saved', '3': 14, '4': 1, '5': 8, '10': 'saved'},
    {'1': 'archived', '3': 15, '4': 1, '5': 8, '10': 'archived'},
    {'1': 'folders', '3': 16, '4': 3, '5': 9, '10': 'folders'},
  ],
};

/// Descriptor for `Chat`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List chatDescriptor = $convert.base64Decode(
    'CgRDaGF0Eh0KCmFjY291bnRfaWQYASABKAlSCWFjY291bnRJZBIOCgJpZBgCIAEoCVICaWQSKQ'
    'oEa2luZBgDIAEoDjIVLm1lZ2FjaGF0LnYxLkNoYXRLaW5kUgRraW5kEhQKBXRpdGxlGAQgASgJ'
    'UgV0aXRsZRIdCgphdmF0YXJfdXJsGAUgASgJUglhdmF0YXJVcmwSFgoGdW5yZWFkGAYgASgFUg'
    'Z1bnJlYWQSFAoFbXV0ZWQYByABKAhSBW11dGVkEjcKDGxhc3RfbWVzc2FnZRgIIAEoCzIULm1l'
    'Z2FjaGF0LnYxLk1lc3NhZ2VSC2xhc3RNZXNzYWdlEh0KCnVwZGF0ZWRfbXMYCSABKANSCXVwZG'
    'F0ZWRNcxIWCgZwaW5uZWQYCiABKAhSBnBpbm5lZBIbCglyZWFkX29ubHkYCyABKAhSCHJlYWRP'
    'bmx5EhQKBWZvcnVtGAwgASgIUgVmb3J1bRIWCgZvbmxpbmUYDSABKAhSBm9ubGluZRIUCgVzYX'
    'ZlZBgOIAEoCFIFc2F2ZWQSGgoIYXJjaGl2ZWQYDyABKAhSCGFyY2hpdmVkEhgKB2ZvbGRlcnMY'
    'ECADKAlSB2ZvbGRlcnM=');

@$core.Deprecated('Use manageChatRequestDescriptor instead')
const ManageChatRequest$json = {
  '1': 'ManageChatRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'chat_id', '3': 2, '4': 1, '5': 9, '10': 'chatId'},
    {
      '1': 'action',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.megachat.v1.ChatAction',
      '10': 'action'
    },
  ],
};

/// Descriptor for `ManageChatRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List manageChatRequestDescriptor = $convert.base64Decode(
    'ChFNYW5hZ2VDaGF0UmVxdWVzdBIdCgphY2NvdW50X2lkGAEgASgJUglhY2NvdW50SWQSFwoHY2'
    'hhdF9pZBgCIAEoCVIGY2hhdElkEi8KBmFjdGlvbhgDIAEoDjIXLm1lZ2FjaGF0LnYxLkNoYXRB'
    'Y3Rpb25SBmFjdGlvbg==');

@$core.Deprecated('Use chatFolderDescriptor instead')
const ChatFolder$json = {
  '1': 'ChatFolder',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'id', '3': 2, '4': 1, '5': 9, '10': 'id'},
    {'1': 'title', '3': 3, '4': 1, '5': 9, '10': 'title'},
    {'1': 'emoji', '3': 4, '4': 1, '5': 9, '10': 'emoji'},
    {'1': 'always', '3': 5, '4': 3, '5': 9, '10': 'always'},
    {'1': 'pinned', '3': 6, '4': 3, '5': 9, '10': 'pinned'},
    {'1': 'exclude_muted', '3': 7, '4': 1, '5': 8, '10': 'excludeMuted'},
    {'1': 'exclude_read', '3': 8, '4': 1, '5': 8, '10': 'excludeRead'},
    {'1': 'exclude_archived', '3': 9, '4': 1, '5': 8, '10': 'excludeArchived'},
  ],
};

/// Descriptor for `ChatFolder`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List chatFolderDescriptor = $convert.base64Decode(
    'CgpDaGF0Rm9sZGVyEh0KCmFjY291bnRfaWQYASABKAlSCWFjY291bnRJZBIOCgJpZBgCIAEoCV'
    'ICaWQSFAoFdGl0bGUYAyABKAlSBXRpdGxlEhQKBWVtb2ppGAQgASgJUgVlbW9qaRIWCgZhbHdh'
    'eXMYBSADKAlSBmFsd2F5cxIWCgZwaW5uZWQYBiADKAlSBnBpbm5lZBIjCg1leGNsdWRlX211dG'
    'VkGAcgASgIUgxleGNsdWRlTXV0ZWQSIQoMZXhjbHVkZV9yZWFkGAggASgIUgtleGNsdWRlUmVh'
    'ZBIpChBleGNsdWRlX2FyY2hpdmVkGAkgASgIUg9leGNsdWRlQXJjaGl2ZWQ=');

@$core.Deprecated('Use chatFoldersReplyDescriptor instead')
const ChatFoldersReply$json = {
  '1': 'ChatFoldersReply',
  '2': [
    {
      '1': 'folders',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.ChatFolder',
      '10': 'folders'
    },
    {
      '1': 'failed',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.AccountError',
      '10': 'failed'
    },
  ],
};

/// Descriptor for `ChatFoldersReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List chatFoldersReplyDescriptor = $convert.base64Decode(
    'ChBDaGF0Rm9sZGVyc1JlcGx5EjEKB2ZvbGRlcnMYASADKAsyFy5tZWdhY2hhdC52MS5DaGF0Rm'
    '9sZGVyUgdmb2xkZXJzEjEKBmZhaWxlZBgCIAMoCzIZLm1lZ2FjaGF0LnYxLkFjY291bnRFcnJv'
    'clIGZmFpbGVk');

@$core.Deprecated('Use chatsRequestDescriptor instead')
const ChatsRequest$json = {
  '1': 'ChatsRequest',
  '2': [
    {'1': 'account_ids', '3': 1, '4': 3, '5': 9, '10': 'accountIds'},
    {'1': 'limit', '3': 2, '4': 1, '5': 5, '10': 'limit'},
  ],
};

/// Descriptor for `ChatsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List chatsRequestDescriptor = $convert.base64Decode(
    'CgxDaGF0c1JlcXVlc3QSHwoLYWNjb3VudF9pZHMYASADKAlSCmFjY291bnRJZHMSFAoFbGltaX'
    'QYAiABKAVSBWxpbWl0');

@$core.Deprecated('Use chatsReplyDescriptor instead')
const ChatsReply$json = {
  '1': 'ChatsReply',
  '2': [
    {
      '1': 'chats',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Chat',
      '10': 'chats'
    },
  ],
};

/// Descriptor for `ChatsReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List chatsReplyDescriptor = $convert.base64Decode(
    'CgpDaGF0c1JlcGx5EicKBWNoYXRzGAEgAygLMhEubWVnYWNoYXQudjEuQ2hhdFIFY2hhdHM=');

@$core.Deprecated('Use attachmentDescriptor instead')
const Attachment$json = {
  '1': 'Attachment',
  '2': [
    {
      '1': 'kind',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.megachat.v1.AttachmentKind',
      '10': 'kind'
    },
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'mime', '3': 3, '4': 1, '5': 9, '10': 'mime'},
    {'1': 'size', '3': 4, '4': 1, '5': 3, '10': 'size'},
    {'1': 'ref', '3': 5, '4': 1, '5': 9, '10': 'ref'},
    {'1': 'thumb', '3': 6, '4': 1, '5': 9, '10': 'thumb'},
    {'1': 'width', '3': 7, '4': 1, '5': 5, '10': 'width'},
    {'1': 'height', '3': 8, '4': 1, '5': 5, '10': 'height'},
    {'1': 'duration_ms', '3': 9, '4': 1, '5': 5, '10': 'durationMs'},
  ],
};

/// Descriptor for `Attachment`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List attachmentDescriptor = $convert.base64Decode(
    'CgpBdHRhY2htZW50Ei8KBGtpbmQYASABKA4yGy5tZWdhY2hhdC52MS5BdHRhY2htZW50S2luZF'
    'IEa2luZBISCgRuYW1lGAIgASgJUgRuYW1lEhIKBG1pbWUYAyABKAlSBG1pbWUSEgoEc2l6ZRgE'
    'IAEoA1IEc2l6ZRIQCgNyZWYYBSABKAlSA3JlZhIUCgV0aHVtYhgGIAEoCVIFdGh1bWISFAoFd2'
    'lkdGgYByABKAVSBXdpZHRoEhYKBmhlaWdodBgIIAEoBVIGaGVpZ2h0Eh8KC2R1cmF0aW9uX21z'
    'GAkgASgFUgpkdXJhdGlvbk1z');

@$core.Deprecated('Use reactionDescriptor instead')
const Reaction$json = {
  '1': 'Reaction',
  '2': [
    {'1': 'emoji', '3': 1, '4': 1, '5': 9, '10': 'emoji'},
    {'1': 'count', '3': 2, '4': 1, '5': 5, '10': 'count'},
    {'1': 'mine', '3': 3, '4': 1, '5': 8, '10': 'mine'},
  ],
};

/// Descriptor for `Reaction`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List reactionDescriptor = $convert.base64Decode(
    'CghSZWFjdGlvbhIUCgVlbW9qaRgBIAEoCVIFZW1vamkSFAoFY291bnQYAiABKAVSBWNvdW50Eh'
    'IKBG1pbmUYAyABKAhSBG1pbmU=');

@$core.Deprecated('Use threadDescriptor instead')
const Thread$json = {
  '1': 'Thread',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'replies', '3': 2, '4': 1, '5': 5, '10': 'replies'},
    {'1': 'last_ms', '3': 3, '4': 1, '5': 3, '10': 'lastMs'},
    {
      '1': 'users',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.User',
      '10': 'users'
    },
  ],
};

/// Descriptor for `Thread`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List threadDescriptor = $convert.base64Decode(
    'CgZUaHJlYWQSDgoCaWQYASABKAlSAmlkEhgKB3JlcGxpZXMYAiABKAVSB3JlcGxpZXMSFwoHbG'
    'FzdF9tcxgDIAEoA1IGbGFzdE1zEicKBXVzZXJzGAQgAygLMhEubWVnYWNoYXQudjEuVXNlclIF'
    'dXNlcnM=');

@$core.Deprecated('Use messageDescriptor instead')
const Message$json = {
  '1': 'Message',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'chat_id', '3': 2, '4': 1, '5': 9, '10': 'chatId'},
    {'1': 'id', '3': 3, '4': 1, '5': 9, '10': 'id'},
    {
      '1': 'sender',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.User',
      '10': 'sender'
    },
    {'1': 'text', '3': 5, '4': 1, '5': 9, '10': 'text'},
    {'1': 'time_ms', '3': 6, '4': 1, '5': 3, '10': 'timeMs'},
    {'1': 'edited_ms', '3': 7, '4': 1, '5': 3, '10': 'editedMs'},
    {'1': 'outgoing', '3': 8, '4': 1, '5': 8, '10': 'outgoing'},
    {'1': 'reply_to_id', '3': 9, '4': 1, '5': 9, '10': 'replyToId'},
    {
      '1': 'attachments',
      '3': 10,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Attachment',
      '10': 'attachments'
    },
    {'1': 'deleted', '3': 11, '4': 1, '5': 8, '10': 'deleted'},
    {'1': 'service', '3': 12, '4': 1, '5': 8, '10': 'service'},
    {
      '1': 'reactions',
      '3': 13,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Reaction',
      '10': 'reactions'
    },
    {
      '1': 'thread',
      '3': 14,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.Thread',
      '10': 'thread'
    },
    {'1': 'in_thread', '3': 15, '4': 1, '5': 9, '10': 'inThread'},
    {'1': 'group', '3': 16, '4': 1, '5': 9, '10': 'group'},
    {'1': 'mentioned', '3': 17, '4': 1, '5': 8, '10': 'mentioned'},
    {'1': 'forwarded', '3': 18, '4': 1, '5': 8, '10': 'forwarded'},
    {'1': 'forwarded_from', '3': 19, '4': 1, '5': 9, '10': 'forwardedFrom'},
    {'1': 'quote', '3': 20, '4': 1, '5': 9, '10': 'quote'},
    {
      '1': 'format',
      '3': 21,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.TextSpan',
      '10': 'format'
    },
    {'1': 'markdown', '3': 22, '4': 1, '5': 9, '10': 'markdown'},
  ],
};

/// Descriptor for `Message`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List messageDescriptor = $convert.base64Decode(
    'CgdNZXNzYWdlEh0KCmFjY291bnRfaWQYASABKAlSCWFjY291bnRJZBIXCgdjaGF0X2lkGAIgAS'
    'gJUgZjaGF0SWQSDgoCaWQYAyABKAlSAmlkEikKBnNlbmRlchgEIAEoCzIRLm1lZ2FjaGF0LnYx'
    'LlVzZXJSBnNlbmRlchISCgR0ZXh0GAUgASgJUgR0ZXh0EhcKB3RpbWVfbXMYBiABKANSBnRpbW'
    'VNcxIbCgllZGl0ZWRfbXMYByABKANSCGVkaXRlZE1zEhoKCG91dGdvaW5nGAggASgIUghvdXRn'
    'b2luZxIeCgtyZXBseV90b19pZBgJIAEoCVIJcmVwbHlUb0lkEjkKC2F0dGFjaG1lbnRzGAogAy'
    'gLMhcubWVnYWNoYXQudjEuQXR0YWNobWVudFILYXR0YWNobWVudHMSGAoHZGVsZXRlZBgLIAEo'
    'CFIHZGVsZXRlZBIYCgdzZXJ2aWNlGAwgASgIUgdzZXJ2aWNlEjMKCXJlYWN0aW9ucxgNIAMoCz'
    'IVLm1lZ2FjaGF0LnYxLlJlYWN0aW9uUglyZWFjdGlvbnMSKwoGdGhyZWFkGA4gASgLMhMubWVn'
    'YWNoYXQudjEuVGhyZWFkUgZ0aHJlYWQSGwoJaW5fdGhyZWFkGA8gASgJUghpblRocmVhZBIUCg'
    'Vncm91cBgQIAEoCVIFZ3JvdXASHAoJbWVudGlvbmVkGBEgASgIUgltZW50aW9uZWQSHAoJZm9y'
    'd2FyZGVkGBIgASgIUglmb3J3YXJkZWQSJQoOZm9yd2FyZGVkX2Zyb20YEyABKAlSDWZvcndhcm'
    'RlZEZyb20SFAoFcXVvdGUYFCABKAlSBXF1b3RlEi0KBmZvcm1hdBgVIAMoCzIVLm1lZ2FjaGF0'
    'LnYxLlRleHRTcGFuUgZmb3JtYXQSGgoIbWFya2Rvd24YFiABKAlSCG1hcmtkb3du');

@$core.Deprecated('Use textSpanDescriptor instead')
const TextSpan$json = {
  '1': 'TextSpan',
  '2': [
    {
      '1': 'style',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.megachat.v1.TextStyle',
      '10': 'style'
    },
    {'1': 'offset', '3': 2, '4': 1, '5': 5, '10': 'offset'},
    {'1': 'length', '3': 3, '4': 1, '5': 5, '10': 'length'},
    {'1': 'url', '3': 4, '4': 1, '5': 9, '10': 'url'},
    {'1': 'language', '3': 5, '4': 1, '5': 9, '10': 'language'},
  ],
};

/// Descriptor for `TextSpan`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List textSpanDescriptor = $convert.base64Decode(
    'CghUZXh0U3BhbhIsCgVzdHlsZRgBIAEoDjIWLm1lZ2FjaGF0LnYxLlRleHRTdHlsZVIFc3R5bG'
    'USFgoGb2Zmc2V0GAIgASgFUgZvZmZzZXQSFgoGbGVuZ3RoGAMgASgFUgZsZW5ndGgSEAoDdXJs'
    'GAQgASgJUgN1cmwSGgoIbGFuZ3VhZ2UYBSABKAlSCGxhbmd1YWdl');

@$core.Deprecated('Use threadRequestDescriptor instead')
const ThreadRequest$json = {
  '1': 'ThreadRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'chat_id', '3': 2, '4': 1, '5': 9, '10': 'chatId'},
    {'1': 'thread_id', '3': 3, '4': 1, '5': 9, '10': 'threadId'},
    {'1': 'before_id', '3': 4, '4': 1, '5': 9, '10': 'beforeId'},
    {'1': 'limit', '3': 5, '4': 1, '5': 5, '10': 'limit'},
  ],
};

/// Descriptor for `ThreadRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List threadRequestDescriptor = $convert.base64Decode(
    'Cg1UaHJlYWRSZXF1ZXN0Eh0KCmFjY291bnRfaWQYASABKAlSCWFjY291bnRJZBIXCgdjaGF0X2'
    'lkGAIgASgJUgZjaGF0SWQSGwoJdGhyZWFkX2lkGAMgASgJUgh0aHJlYWRJZBIbCgliZWZvcmVf'
    'aWQYBCABKAlSCGJlZm9yZUlkEhQKBWxpbWl0GAUgASgFUgVsaW1pdA==');

@$core.Deprecated('Use topicsRequestDescriptor instead')
const TopicsRequest$json = {
  '1': 'TopicsRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'chat_id', '3': 2, '4': 1, '5': 9, '10': 'chatId'},
  ],
};

/// Descriptor for `TopicsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List topicsRequestDescriptor = $convert.base64Decode(
    'Cg1Ub3BpY3NSZXF1ZXN0Eh0KCmFjY291bnRfaWQYASABKAlSCWFjY291bnRJZBIXCgdjaGF0X2'
    'lkGAIgASgJUgZjaGF0SWQ=');

@$core.Deprecated('Use topicDescriptor instead')
const Topic$json = {
  '1': 'Topic',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'title', '3': 2, '4': 1, '5': 9, '10': 'title'},
    {'1': 'unread', '3': 3, '4': 1, '5': 5, '10': 'unread'},
    {'1': 'color', '3': 4, '4': 1, '5': 5, '10': 'color'},
    {'1': 'pinned', '3': 5, '4': 1, '5': 8, '10': 'pinned'},
    {'1': 'closed', '3': 6, '4': 1, '5': 8, '10': 'closed'},
    {'1': 'hidden', '3': 7, '4': 1, '5': 8, '10': 'hidden'},
    {
      '1': 'last_message',
      '3': 8,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.Message',
      '10': 'lastMessage'
    },
    {'1': 'updated_ms', '3': 9, '4': 1, '5': 3, '10': 'updatedMs'},
  ],
};

/// Descriptor for `Topic`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List topicDescriptor = $convert.base64Decode(
    'CgVUb3BpYxIOCgJpZBgBIAEoCVICaWQSFAoFdGl0bGUYAiABKAlSBXRpdGxlEhYKBnVucmVhZB'
    'gDIAEoBVIGdW5yZWFkEhQKBWNvbG9yGAQgASgFUgVjb2xvchIWCgZwaW5uZWQYBSABKAhSBnBp'
    'bm5lZBIWCgZjbG9zZWQYBiABKAhSBmNsb3NlZBIWCgZoaWRkZW4YByABKAhSBmhpZGRlbhI3Cg'
    'xsYXN0X21lc3NhZ2UYCCABKAsyFC5tZWdhY2hhdC52MS5NZXNzYWdlUgtsYXN0TWVzc2FnZRId'
    'Cgp1cGRhdGVkX21zGAkgASgDUgl1cGRhdGVkTXM=');

@$core.Deprecated('Use topicsReplyDescriptor instead')
const TopicsReply$json = {
  '1': 'TopicsReply',
  '2': [
    {
      '1': 'topics',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Topic',
      '10': 'topics'
    },
  ],
};

/// Descriptor for `TopicsReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List topicsReplyDescriptor = $convert.base64Decode(
    'CgtUb3BpY3NSZXBseRIqCgZ0b3BpY3MYASADKAsyEi5tZWdhY2hhdC52MS5Ub3BpY1IGdG9waW'
    'Nz');

@$core.Deprecated('Use reactRequestDescriptor instead')
const ReactRequest$json = {
  '1': 'ReactRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'chat_id', '3': 2, '4': 1, '5': 9, '10': 'chatId'},
    {'1': 'message_id', '3': 3, '4': 1, '5': 9, '10': 'messageId'},
    {'1': 'emoji', '3': 4, '4': 1, '5': 9, '10': 'emoji'},
    {'1': 'on', '3': 5, '4': 1, '5': 8, '10': 'on'},
  ],
};

/// Descriptor for `ReactRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List reactRequestDescriptor = $convert.base64Decode(
    'CgxSZWFjdFJlcXVlc3QSHQoKYWNjb3VudF9pZBgBIAEoCVIJYWNjb3VudElkEhcKB2NoYXRfaW'
    'QYAiABKAlSBmNoYXRJZBIdCgptZXNzYWdlX2lkGAMgASgJUgltZXNzYWdlSWQSFAoFZW1vamkY'
    'BCABKAlSBWVtb2ppEg4KAm9uGAUgASgIUgJvbg==');

@$core.Deprecated('Use editRequestDescriptor instead')
const EditRequest$json = {
  '1': 'EditRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'chat_id', '3': 2, '4': 1, '5': 9, '10': 'chatId'},
    {'1': 'message_id', '3': 3, '4': 1, '5': 9, '10': 'messageId'},
    {'1': 'text', '3': 4, '4': 1, '5': 9, '10': 'text'},
    {'1': 'thread_id', '3': 5, '4': 1, '5': 9, '10': 'threadId'},
  ],
};

/// Descriptor for `EditRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List editRequestDescriptor = $convert.base64Decode(
    'CgtFZGl0UmVxdWVzdBIdCgphY2NvdW50X2lkGAEgASgJUglhY2NvdW50SWQSFwoHY2hhdF9pZB'
    'gCIAEoCVIGY2hhdElkEh0KCm1lc3NhZ2VfaWQYAyABKAlSCW1lc3NhZ2VJZBISCgR0ZXh0GAQg'
    'ASgJUgR0ZXh0EhsKCXRocmVhZF9pZBgFIAEoCVIIdGhyZWFkSWQ=');

@$core.Deprecated('Use deleteRequestDescriptor instead')
const DeleteRequest$json = {
  '1': 'DeleteRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'chat_id', '3': 2, '4': 1, '5': 9, '10': 'chatId'},
    {'1': 'message_ids', '3': 3, '4': 3, '5': 9, '10': 'messageIds'},
    {'1': 'for_all', '3': 4, '4': 1, '5': 8, '10': 'forAll'},
    {'1': 'thread_id', '3': 5, '4': 1, '5': 9, '10': 'threadId'},
  ],
};

/// Descriptor for `DeleteRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteRequestDescriptor = $convert.base64Decode(
    'Cg1EZWxldGVSZXF1ZXN0Eh0KCmFjY291bnRfaWQYASABKAlSCWFjY291bnRJZBIXCgdjaGF0X2'
    'lkGAIgASgJUgZjaGF0SWQSHwoLbWVzc2FnZV9pZHMYAyADKAlSCm1lc3NhZ2VJZHMSFwoHZm9y'
    'X2FsbBgEIAEoCFIGZm9yQWxsEhsKCXRocmVhZF9pZBgFIAEoCVIIdGhyZWFkSWQ=');

@$core.Deprecated('Use mentionDescriptor instead')
const Mention$json = {
  '1': 'Mention',
  '2': [
    {'1': 'offset', '3': 1, '4': 1, '5': 5, '10': 'offset'},
    {'1': 'length', '3': 2, '4': 1, '5': 5, '10': 'length'},
    {
      '1': 'user',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.User',
      '10': 'user'
    },
  ],
};

/// Descriptor for `Mention`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List mentionDescriptor = $convert.base64Decode(
    'CgdNZW50aW9uEhYKBm9mZnNldBgBIAEoBVIGb2Zmc2V0EhYKBmxlbmd0aBgCIAEoBVIGbGVuZ3'
    'RoEiUKBHVzZXIYAyABKAsyES5tZWdhY2hhdC52MS5Vc2VyUgR1c2Vy');

@$core.Deprecated('Use membersRequestDescriptor instead')
const MembersRequest$json = {
  '1': 'MembersRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'chat_id', '3': 2, '4': 1, '5': 9, '10': 'chatId'},
    {'1': 'query', '3': 3, '4': 1, '5': 9, '10': 'query'},
    {'1': 'limit', '3': 4, '4': 1, '5': 5, '10': 'limit'},
  ],
};

/// Descriptor for `MembersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List membersRequestDescriptor = $convert.base64Decode(
    'Cg5NZW1iZXJzUmVxdWVzdBIdCgphY2NvdW50X2lkGAEgASgJUglhY2NvdW50SWQSFwoHY2hhdF'
    '9pZBgCIAEoCVIGY2hhdElkEhQKBXF1ZXJ5GAMgASgJUgVxdWVyeRIUCgVsaW1pdBgEIAEoBVIF'
    'bGltaXQ=');

@$core.Deprecated('Use membersReplyDescriptor instead')
const MembersReply$json = {
  '1': 'MembersReply',
  '2': [
    {
      '1': 'users',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.User',
      '10': 'users'
    },
  ],
};

/// Descriptor for `MembersReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List membersReplyDescriptor = $convert.base64Decode(
    'CgxNZW1iZXJzUmVwbHkSJwoFdXNlcnMYASADKAsyES5tZWdhY2hhdC52MS5Vc2VyUgV1c2Vycw'
    '==');

@$core.Deprecated('Use linkPreviewRequestDescriptor instead')
const LinkPreviewRequest$json = {
  '1': 'LinkPreviewRequest',
  '2': [
    {'1': 'url', '3': 1, '4': 1, '5': 9, '10': 'url'},
  ],
};

/// Descriptor for `LinkPreviewRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List linkPreviewRequestDescriptor = $convert
    .base64Decode('ChJMaW5rUHJldmlld1JlcXVlc3QSEAoDdXJsGAEgASgJUgN1cmw=');

@$core.Deprecated('Use linkPreviewReplyDescriptor instead')
const LinkPreviewReply$json = {
  '1': 'LinkPreviewReply',
  '2': [
    {'1': 'url', '3': 1, '4': 1, '5': 9, '10': 'url'},
    {'1': 'site_name', '3': 2, '4': 1, '5': 9, '10': 'siteName'},
    {'1': 'title', '3': 3, '4': 1, '5': 9, '10': 'title'},
    {'1': 'description', '3': 4, '4': 1, '5': 9, '10': 'description'},
    {'1': 'image', '3': 5, '4': 1, '5': 12, '10': 'image'},
    {'1': 'image_mime', '3': 6, '4': 1, '5': 9, '10': 'imageMime'},
  ],
};

/// Descriptor for `LinkPreviewReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List linkPreviewReplyDescriptor = $convert.base64Decode(
    'ChBMaW5rUHJldmlld1JlcGx5EhAKA3VybBgBIAEoCVIDdXJsEhsKCXNpdGVfbmFtZRgCIAEoCV'
    'IIc2l0ZU5hbWUSFAoFdGl0bGUYAyABKAlSBXRpdGxlEiAKC2Rlc2NyaXB0aW9uGAQgASgJUgtk'
    'ZXNjcmlwdGlvbhIUCgVpbWFnZRgFIAEoDFIFaW1hZ2USHQoKaW1hZ2VfbWltZRgGIAEoCVIJaW'
    '1hZ2VNaW1l');

@$core.Deprecated('Use profileRequestDescriptor instead')
const ProfileRequest$json = {
  '1': 'ProfileRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'chat_id', '3': 2, '4': 1, '5': 9, '10': 'chatId'},
  ],
};

/// Descriptor for `ProfileRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List profileRequestDescriptor = $convert.base64Decode(
    'Cg5Qcm9maWxlUmVxdWVzdBIdCgphY2NvdW50X2lkGAEgASgJUglhY2NvdW50SWQSFwoHY2hhdF'
    '9pZBgCIAEoCVIGY2hhdElk');

@$core.Deprecated('Use chatProfileDescriptor instead')
const ChatProfile$json = {
  '1': 'ChatProfile',
  '2': [
    {'1': 'title', '3': 1, '4': 1, '5': 9, '10': 'title'},
    {'1': 'avatar_url', '3': 2, '4': 1, '5': 9, '10': 'avatarUrl'},
    {'1': 'username', '3': 3, '4': 1, '5': 9, '10': 'username'},
    {'1': 'about', '3': 4, '4': 1, '5': 9, '10': 'about'},
    {'1': 'phone', '3': 5, '4': 1, '5': 9, '10': 'phone'},
    {'1': 'email', '3': 6, '4': 1, '5': 9, '10': 'email'},
    {'1': 'position', '3': 7, '4': 1, '5': 9, '10': 'position'},
    {'1': 'link', '3': 8, '4': 1, '5': 9, '10': 'link'},
    {'1': 'members', '3': 9, '4': 1, '5': 5, '10': 'members'},
    {
      '1': 'member_list',
      '3': 10,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.User',
      '10': 'memberList'
    },
    {'1': 'last_seen_ms', '3': 11, '4': 1, '5': 3, '10': 'lastSeenMs'},
    {'1': 'online', '3': 12, '4': 1, '5': 8, '10': 'online'},
    {'1': 'bot', '3': 13, '4': 1, '5': 8, '10': 'bot'},
  ],
};

/// Descriptor for `ChatProfile`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List chatProfileDescriptor = $convert.base64Decode(
    'CgtDaGF0UHJvZmlsZRIUCgV0aXRsZRgBIAEoCVIFdGl0bGUSHQoKYXZhdGFyX3VybBgCIAEoCV'
    'IJYXZhdGFyVXJsEhoKCHVzZXJuYW1lGAMgASgJUgh1c2VybmFtZRIUCgVhYm91dBgEIAEoCVIF'
    'YWJvdXQSFAoFcGhvbmUYBSABKAlSBXBob25lEhQKBWVtYWlsGAYgASgJUgVlbWFpbBIaCghwb3'
    'NpdGlvbhgHIAEoCVIIcG9zaXRpb24SEgoEbGluaxgIIAEoCVIEbGluaxIYCgdtZW1iZXJzGAkg'
    'ASgFUgdtZW1iZXJzEjIKC21lbWJlcl9saXN0GAogAygLMhEubWVnYWNoYXQudjEuVXNlclIKbW'
    'VtYmVyTGlzdBIgCgxsYXN0X3NlZW5fbXMYCyABKANSCmxhc3RTZWVuTXMSFgoGb25saW5lGAwg'
    'ASgIUgZvbmxpbmUSEAoDYm90GA0gASgIUgNib3Q=');

@$core.Deprecated('Use mediaRequestDescriptor instead')
const MediaRequest$json = {
  '1': 'MediaRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'ref', '3': 2, '4': 1, '5': 9, '10': 'ref'},
    {'1': 'max_bytes', '3': 3, '4': 1, '5': 3, '10': 'maxBytes'},
    {'1': 'prefetch', '3': 4, '4': 1, '5': 8, '10': 'prefetch'},
  ],
};

/// Descriptor for `MediaRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List mediaRequestDescriptor = $convert.base64Decode(
    'CgxNZWRpYVJlcXVlc3QSHQoKYWNjb3VudF9pZBgBIAEoCVIJYWNjb3VudElkEhAKA3JlZhgCIA'
    'EoCVIDcmVmEhsKCW1heF9ieXRlcxgDIAEoA1IIbWF4Qnl0ZXMSGgoIcHJlZmV0Y2gYBCABKAhS'
    'CHByZWZldGNo');

@$core.Deprecated('Use mediaReplyDescriptor instead')
const MediaReply$json = {
  '1': 'MediaReply',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 12, '10': 'data'},
    {'1': 'mime', '3': 2, '4': 1, '5': 9, '10': 'mime'},
  ],
};

/// Descriptor for `MediaReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List mediaReplyDescriptor = $convert.base64Decode(
    'CgpNZWRpYVJlcGx5EhIKBGRhdGEYASABKAxSBGRhdGESEgoEbWltZRgCIAEoCVIEbWltZQ==');

@$core.Deprecated('Use historyRequestDescriptor instead')
const HistoryRequest$json = {
  '1': 'HistoryRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'chat_id', '3': 2, '4': 1, '5': 9, '10': 'chatId'},
    {'1': 'before_id', '3': 3, '4': 1, '5': 9, '10': 'beforeId'},
    {'1': 'limit', '3': 4, '4': 1, '5': 5, '10': 'limit'},
    {'1': 'cached_only', '3': 5, '4': 1, '5': 8, '10': 'cachedOnly'},
  ],
};

/// Descriptor for `HistoryRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List historyRequestDescriptor = $convert.base64Decode(
    'Cg5IaXN0b3J5UmVxdWVzdBIdCgphY2NvdW50X2lkGAEgASgJUglhY2NvdW50SWQSFwoHY2hhdF'
    '9pZBgCIAEoCVIGY2hhdElkEhsKCWJlZm9yZV9pZBgDIAEoCVIIYmVmb3JlSWQSFAoFbGltaXQY'
    'BCABKAVSBWxpbWl0Eh8KC2NhY2hlZF9vbmx5GAUgASgIUgpjYWNoZWRPbmx5');

@$core.Deprecated('Use historyReplyDescriptor instead')
const HistoryReply$json = {
  '1': 'HistoryReply',
  '2': [
    {
      '1': 'messages',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Message',
      '10': 'messages'
    },
    {'1': 'more', '3': 2, '4': 1, '5': 8, '10': 'more'},
  ],
};

/// Descriptor for `HistoryReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List historyReplyDescriptor = $convert.base64Decode(
    'CgxIaXN0b3J5UmVwbHkSMAoIbWVzc2FnZXMYASADKAsyFC5tZWdhY2hhdC52MS5NZXNzYWdlUg'
    'htZXNzYWdlcxISCgRtb3JlGAIgASgIUgRtb3Jl');

@$core.Deprecated('Use sendRequestDescriptor instead')
const SendRequest$json = {
  '1': 'SendRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'chat_id', '3': 2, '4': 1, '5': 9, '10': 'chatId'},
    {'1': 'text', '3': 3, '4': 1, '5': 9, '10': 'text'},
    {'1': 'reply_to_id', '3': 4, '4': 1, '5': 9, '10': 'replyToId'},
    {'1': 'thread_id', '3': 5, '4': 1, '5': 9, '10': 'threadId'},
    {'1': 'quote', '3': 6, '4': 1, '5': 9, '10': 'quote'},
    {
      '1': 'files',
      '3': 7,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.OutFile',
      '10': 'files'
    },
    {
      '1': 'mentions',
      '3': 8,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Mention',
      '10': 'mentions'
    },
  ],
};

/// Descriptor for `SendRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List sendRequestDescriptor = $convert.base64Decode(
    'CgtTZW5kUmVxdWVzdBIdCgphY2NvdW50X2lkGAEgASgJUglhY2NvdW50SWQSFwoHY2hhdF9pZB'
    'gCIAEoCVIGY2hhdElkEhIKBHRleHQYAyABKAlSBHRleHQSHgoLcmVwbHlfdG9faWQYBCABKAlS'
    'CXJlcGx5VG9JZBIbCgl0aHJlYWRfaWQYBSABKAlSCHRocmVhZElkEhQKBXF1b3RlGAYgASgJUg'
    'VxdW90ZRIqCgVmaWxlcxgHIAMoCzIULm1lZ2FjaGF0LnYxLk91dEZpbGVSBWZpbGVzEjAKCG1l'
    'bnRpb25zGAggAygLMhQubWVnYWNoYXQudjEuTWVudGlvblIIbWVudGlvbnM=');

@$core.Deprecated('Use outFileDescriptor instead')
const OutFile$json = {
  '1': 'OutFile',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'mime', '3': 2, '4': 1, '5': 9, '10': 'mime'},
    {'1': 'data', '3': 3, '4': 1, '5': 12, '10': 'data'},
  ],
};

/// Descriptor for `OutFile`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List outFileDescriptor = $convert.base64Decode(
    'CgdPdXRGaWxlEhIKBG5hbWUYASABKAlSBG5hbWUSEgoEbWltZRgCIAEoCVIEbWltZRISCgRkYX'
    'RhGAMgASgMUgRkYXRh');

@$core.Deprecated('Use forwardRequestDescriptor instead')
const ForwardRequest$json = {
  '1': 'ForwardRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'chat_id', '3': 2, '4': 1, '5': 9, '10': 'chatId'},
    {'1': 'message_ids', '3': 3, '4': 3, '5': 9, '10': 'messageIds'},
    {'1': 'to_chat_id', '3': 4, '4': 1, '5': 9, '10': 'toChatId'},
  ],
};

/// Descriptor for `ForwardRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List forwardRequestDescriptor = $convert.base64Decode(
    'Cg5Gb3J3YXJkUmVxdWVzdBIdCgphY2NvdW50X2lkGAEgASgJUglhY2NvdW50SWQSFwoHY2hhdF'
    '9pZBgCIAEoCVIGY2hhdElkEh8KC21lc3NhZ2VfaWRzGAMgAygJUgptZXNzYWdlSWRzEhwKCnRv'
    'X2NoYXRfaWQYBCABKAlSCHRvQ2hhdElk');

@$core.Deprecated('Use forwardReplyDescriptor instead')
const ForwardReply$json = {
  '1': 'ForwardReply',
  '2': [
    {
      '1': 'messages',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Message',
      '10': 'messages'
    },
  ],
};

/// Descriptor for `ForwardReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List forwardReplyDescriptor = $convert.base64Decode(
    'CgxGb3J3YXJkUmVwbHkSMAoIbWVzc2FnZXMYASADKAsyFC5tZWdhY2hhdC52MS5NZXNzYWdlUg'
    'htZXNzYWdlcw==');

@$core.Deprecated('Use markReadRequestDescriptor instead')
const MarkReadRequest$json = {
  '1': 'MarkReadRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'chat_id', '3': 2, '4': 1, '5': 9, '10': 'chatId'},
    {'1': 'message_id', '3': 3, '4': 1, '5': 9, '10': 'messageId'},
    {'1': 'thread_id', '3': 4, '4': 1, '5': 9, '10': 'threadId'},
  ],
};

/// Descriptor for `MarkReadRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List markReadRequestDescriptor = $convert.base64Decode(
    'Cg9NYXJrUmVhZFJlcXVlc3QSHQoKYWNjb3VudF9pZBgBIAEoCVIJYWNjb3VudElkEhcKB2NoYX'
    'RfaWQYAiABKAlSBmNoYXRJZBIdCgptZXNzYWdlX2lkGAMgASgJUgltZXNzYWdlSWQSGwoJdGhy'
    'ZWFkX2lkGAQgASgJUgh0aHJlYWRJZA==');

@$core.Deprecated('Use eventDescriptor instead')
const Event$json = {
  '1': 'Event',
  '2': [
    {
      '1': 'kind',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.megachat.v1.Event.Kind',
      '10': 'kind'
    },
    {
      '1': 'account',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.Account',
      '10': 'account'
    },
    {
      '1': 'chat',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.Chat',
      '10': 'chat'
    },
    {
      '1': 'message',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.Message',
      '10': 'message'
    },
    {'1': 'account_id', '3': 5, '4': 1, '5': 9, '10': 'accountId'},
    {
      '1': 'call',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.Call',
      '10': 'call'
    },
    {
      '1': 'signal',
      '3': 7,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.CallSignal',
      '10': 'signal'
    },
    {
      '1': 'mail',
      '3': 8,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.Mail',
      '10': 'mail'
    },
    {'1': 'notify', '3': 9, '4': 1, '5': 8, '10': 'notify'},
  ],
  '4': [Event_Kind$json],
};

@$core.Deprecated('Use eventDescriptor instead')
const Event_Kind$json = {
  '1': 'Kind',
  '2': [
    {'1': 'KIND_UNSPECIFIED', '2': 0},
    {'1': 'KIND_STATE', '2': 1},
    {'1': 'KIND_ACCOUNT', '2': 2},
    {'1': 'KIND_CHAT', '2': 3},
    {'1': 'KIND_MESSAGE_NEW', '2': 4},
    {'1': 'KIND_MESSAGE_EDIT', '2': 5},
    {'1': 'KIND_MESSAGE_DELETE', '2': 6},
    {'1': 'KIND_TYPING', '2': 7},
    {'1': 'KIND_CHAT_REMOVED', '2': 8},
    {'1': 'KIND_THREAD_MESSAGE', '2': 9},
    {'1': 'KIND_CALL', '2': 10},
    {'1': 'KIND_CALL_SIGNAL', '2': 11},
    {'1': 'KIND_CHAT_FOLDERS', '2': 12},
    {'1': 'KIND_MAIL_NEW', '2': 20},
    {'1': 'KIND_MAIL', '2': 21},
    {'1': 'KIND_CALENDAR', '2': 22},
  ],
};

/// Descriptor for `Event`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List eventDescriptor = $convert.base64Decode(
    'CgVFdmVudBIrCgRraW5kGAEgASgOMhcubWVnYWNoYXQudjEuRXZlbnQuS2luZFIEa2luZBIuCg'
    'dhY2NvdW50GAIgASgLMhQubWVnYWNoYXQudjEuQWNjb3VudFIHYWNjb3VudBIlCgRjaGF0GAMg'
    'ASgLMhEubWVnYWNoYXQudjEuQ2hhdFIEY2hhdBIuCgdtZXNzYWdlGAQgASgLMhQubWVnYWNoYX'
    'QudjEuTWVzc2FnZVIHbWVzc2FnZRIdCgphY2NvdW50X2lkGAUgASgJUglhY2NvdW50SWQSJQoE'
    'Y2FsbBgGIAEoCzIRLm1lZ2FjaGF0LnYxLkNhbGxSBGNhbGwSLwoGc2lnbmFsGAcgASgLMhcubW'
    'VnYWNoYXQudjEuQ2FsbFNpZ25hbFIGc2lnbmFsEiUKBG1haWwYCCABKAsyES5tZWdhY2hhdC52'
    'MS5NYWlsUgRtYWlsEhYKBm5vdGlmeRgJIAEoCFIGbm90aWZ5IsUCCgRLaW5kEhQKEEtJTkRfVU'
    '5TUEVDSUZJRUQQABIOCgpLSU5EX1NUQVRFEAESEAoMS0lORF9BQ0NPVU5UEAISDQoJS0lORF9D'
    'SEFUEAMSFAoQS0lORF9NRVNTQUdFX05FVxAEEhUKEUtJTkRfTUVTU0FHRV9FRElUEAUSFwoTS0'
    'lORF9NRVNTQUdFX0RFTEVURRAGEg8KC0tJTkRfVFlQSU5HEAcSFQoRS0lORF9DSEFUX1JFTU9W'
    'RUQQCBIXChNLSU5EX1RIUkVBRF9NRVNTQUdFEAkSDQoJS0lORF9DQUxMEAoSFAoQS0lORF9DQU'
    'xMX1NJR05BTBALEhUKEUtJTkRfQ0hBVF9GT0xERVJTEAwSEQoNS0lORF9NQUlMX05FVxAUEg0K'
    'CUtJTkRfTUFJTBAVEhEKDUtJTkRfQ0FMRU5EQVIQFg==');

@$core.Deprecated('Use logsRequestDescriptor instead')
const LogsRequest$json = {
  '1': 'LogsRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'limit', '3': 2, '4': 1, '5': 5, '10': 'limit'},
  ],
};

/// Descriptor for `LogsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List logsRequestDescriptor = $convert.base64Decode(
    'CgtMb2dzUmVxdWVzdBIdCgphY2NvdW50X2lkGAEgASgJUglhY2NvdW50SWQSFAoFbGltaXQYAi'
    'ABKAVSBWxpbWl0');

@$core.Deprecated('Use logEntryDescriptor instead')
const LogEntry$json = {
  '1': 'LogEntry',
  '2': [
    {'1': 'seq', '3': 1, '4': 1, '5': 4, '10': 'seq'},
    {'1': 'time_ms', '3': 2, '4': 1, '5': 3, '10': 'timeMs'},
    {'1': 'first_ms', '3': 3, '4': 1, '5': 3, '10': 'firstMs'},
    {'1': 'level', '3': 4, '4': 1, '5': 9, '10': 'level'},
    {'1': 'source', '3': 5, '4': 1, '5': 9, '10': 'source'},
    {'1': 'account_id', '3': 6, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'message', '3': 7, '4': 1, '5': 9, '10': 'message'},
    {'1': 'count', '3': 8, '4': 1, '5': 5, '10': 'count'},
  ],
};

/// Descriptor for `LogEntry`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List logEntryDescriptor = $convert.base64Decode(
    'CghMb2dFbnRyeRIQCgNzZXEYASABKARSA3NlcRIXCgd0aW1lX21zGAIgASgDUgZ0aW1lTXMSGQ'
    'oIZmlyc3RfbXMYAyABKANSB2ZpcnN0TXMSFAoFbGV2ZWwYBCABKAlSBWxldmVsEhYKBnNvdXJj'
    'ZRgFIAEoCVIGc291cmNlEh0KCmFjY291bnRfaWQYBiABKAlSCWFjY291bnRJZBIYCgdtZXNzYW'
    'dlGAcgASgJUgdtZXNzYWdlEhQKBWNvdW50GAggASgFUgVjb3VudA==');

@$core.Deprecated('Use logsReplyDescriptor instead')
const LogsReply$json = {
  '1': 'LogsReply',
  '2': [
    {
      '1': 'entries',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.LogEntry',
      '10': 'entries'
    },
  ],
};

/// Descriptor for `LogsReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List logsReplyDescriptor = $convert.base64Decode(
    'CglMb2dzUmVwbHkSLwoHZW50cmllcxgBIAMoCzIVLm1lZ2FjaGF0LnYxLkxvZ0VudHJ5Ugdlbn'
    'RyaWVz');

@$core.Deprecated('Use iceServerDescriptor instead')
const IceServer$json = {
  '1': 'IceServer',
  '2': [
    {'1': 'urls', '3': 1, '4': 3, '5': 9, '10': 'urls'},
    {'1': 'username', '3': 2, '4': 1, '5': 9, '10': 'username'},
    {'1': 'credential', '3': 3, '4': 1, '5': 9, '10': 'credential'},
  ],
};

/// Descriptor for `IceServer`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List iceServerDescriptor = $convert.base64Decode(
    'CglJY2VTZXJ2ZXISEgoEdXJscxgBIAMoCVIEdXJscxIaCgh1c2VybmFtZRgCIAEoCVIIdXNlcm'
    '5hbWUSHgoKY3JlZGVudGlhbBgDIAEoCVIKY3JlZGVudGlhbA==');

@$core.Deprecated('Use callMediaDescriptor instead')
const CallMedia$json = {
  '1': 'CallMedia',
  '2': [
    {'1': 'muted', '3': 1, '4': 1, '5': 8, '10': 'muted'},
    {'1': 'video', '3': 2, '4': 1, '5': 8, '10': 'video'},
    {'1': 'screen', '3': 3, '4': 1, '5': 8, '10': 'screen'},
  ],
};

/// Descriptor for `CallMedia`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List callMediaDescriptor = $convert.base64Decode(
    'CglDYWxsTWVkaWESFAoFbXV0ZWQYASABKAhSBW11dGVkEhQKBXZpZGVvGAIgASgIUgV2aWRlbx'
    'IWCgZzY3JlZW4YAyABKAhSBnNjcmVlbg==');

@$core.Deprecated('Use callDescriptor instead')
const Call$json = {
  '1': 'Call',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'id', '3': 2, '4': 1, '5': 9, '10': 'id'},
    {'1': 'chat_id', '3': 3, '4': 1, '5': 9, '10': 'chatId'},
    {
      '1': 'peer',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.User',
      '10': 'peer'
    },
    {'1': 'outgoing', '3': 5, '4': 1, '5': 8, '10': 'outgoing'},
    {'1': 'video', '3': 6, '4': 1, '5': 8, '10': 'video'},
    {
      '1': 'state',
      '3': 7,
      '4': 1,
      '5': 14,
      '6': '.megachat.v1.CallState',
      '10': 'state'
    },
    {
      '1': 'end',
      '3': 8,
      '4': 1,
      '5': 14,
      '6': '.megachat.v1.CallEnd',
      '10': 'end'
    },
    {'1': 'error', '3': 9, '4': 1, '5': 9, '10': 'error'},
    {'1': 'started_ms', '3': 10, '4': 1, '5': 3, '10': 'startedMs'},
    {'1': 'verification', '3': 11, '4': 3, '5': 9, '10': 'verification'},
    {
      '1': 'ice',
      '3': 12,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.IceServer',
      '10': 'ice'
    },
    {'1': 'relay_only', '3': 13, '4': 1, '5': 8, '10': 'relayOnly'},
    {'1': 'polite', '3': 14, '4': 1, '5': 8, '10': 'polite'},
    {'1': 'channel', '3': 15, '4': 1, '5': 9, '10': 'channel'},
    {'1': 'open_channel', '3': 16, '4': 1, '5': 8, '10': 'openChannel'},
    {'1': 'offer', '3': 17, '4': 1, '5': 8, '10': 'offer'},
    {
      '1': 'remote',
      '3': 18,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.CallMedia',
      '10': 'remote'
    },
  ],
};

/// Descriptor for `Call`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List callDescriptor = $convert.base64Decode(
    'CgRDYWxsEh0KCmFjY291bnRfaWQYASABKAlSCWFjY291bnRJZBIOCgJpZBgCIAEoCVICaWQSFw'
    'oHY2hhdF9pZBgDIAEoCVIGY2hhdElkEiUKBHBlZXIYBCABKAsyES5tZWdhY2hhdC52MS5Vc2Vy'
    'UgRwZWVyEhoKCG91dGdvaW5nGAUgASgIUghvdXRnb2luZxIUCgV2aWRlbxgGIAEoCFIFdmlkZW'
    '8SLAoFc3RhdGUYByABKA4yFi5tZWdhY2hhdC52MS5DYWxsU3RhdGVSBXN0YXRlEiYKA2VuZBgI'
    'IAEoDjIULm1lZ2FjaGF0LnYxLkNhbGxFbmRSA2VuZBIUCgVlcnJvchgJIAEoCVIFZXJyb3ISHQ'
    'oKc3RhcnRlZF9tcxgKIAEoA1IJc3RhcnRlZE1zEiIKDHZlcmlmaWNhdGlvbhgLIAMoCVIMdmVy'
    'aWZpY2F0aW9uEigKA2ljZRgMIAMoCzIWLm1lZ2FjaGF0LnYxLkljZVNlcnZlclIDaWNlEh0KCn'
    'JlbGF5X29ubHkYDSABKAhSCXJlbGF5T25seRIWCgZwb2xpdGUYDiABKAhSBnBvbGl0ZRIYCgdj'
    'aGFubmVsGA8gASgJUgdjaGFubmVsEiEKDG9wZW5fY2hhbm5lbBgQIAEoCFILb3BlbkNoYW5uZW'
    'wSFAoFb2ZmZXIYESABKAhSBW9mZmVyEi4KBnJlbW90ZRgSIAEoCzIWLm1lZ2FjaGF0LnYxLkNh'
    'bGxNZWRpYVIGcmVtb3Rl');

@$core.Deprecated('Use callSignalDescriptor instead')
const CallSignal$json = {
  '1': 'CallSignal',
  '2': [
    {
      '1': 'kind',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.megachat.v1.SignalKind',
      '10': 'kind'
    },
    {'1': 'sdp', '3': 2, '4': 1, '5': 9, '10': 'sdp'},
    {'1': 'mid', '3': 3, '4': 1, '5': 9, '10': 'mid'},
    {'1': 'mline', '3': 4, '4': 1, '5': 5, '10': 'mline'},
    {'1': 'data', '3': 5, '4': 1, '5': 9, '10': 'data'},
    {'1': 'seq', '3': 6, '4': 1, '5': 3, '10': 'seq'},
  ],
};

/// Descriptor for `CallSignal`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List callSignalDescriptor = $convert.base64Decode(
    'CgpDYWxsU2lnbmFsEisKBGtpbmQYASABKA4yFy5tZWdhY2hhdC52MS5TaWduYWxLaW5kUgRraW'
    '5kEhAKA3NkcBgCIAEoCVIDc2RwEhAKA21pZBgDIAEoCVIDbWlkEhQKBW1saW5lGAQgASgFUgVt'
    'bGluZRISCgRkYXRhGAUgASgJUgRkYXRhEhAKA3NlcRgGIAEoA1IDc2Vx');

@$core.Deprecated('Use startCallRequestDescriptor instead')
const StartCallRequest$json = {
  '1': 'StartCallRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'chat_id', '3': 2, '4': 1, '5': 9, '10': 'chatId'},
    {'1': 'video', '3': 3, '4': 1, '5': 8, '10': 'video'},
  ],
};

/// Descriptor for `StartCallRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List startCallRequestDescriptor = $convert.base64Decode(
    'ChBTdGFydENhbGxSZXF1ZXN0Eh0KCmFjY291bnRfaWQYASABKAlSCWFjY291bnRJZBIXCgdjaG'
    'F0X2lkGAIgASgJUgZjaGF0SWQSFAoFdmlkZW8YAyABKAhSBXZpZGVv');

@$core.Deprecated('Use callRequestDescriptor instead')
const CallRequest$json = {
  '1': 'CallRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'call_id', '3': 2, '4': 1, '5': 9, '10': 'callId'},
    {'1': 'video', '3': 3, '4': 1, '5': 8, '10': 'video'},
  ],
};

/// Descriptor for `CallRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List callRequestDescriptor = $convert.base64Decode(
    'CgtDYWxsUmVxdWVzdBIdCgphY2NvdW50X2lkGAEgASgJUglhY2NvdW50SWQSFwoHY2FsbF9pZB'
    'gCIAEoCVIGY2FsbElkEhQKBXZpZGVvGAMgASgIUgV2aWRlbw==');

@$core.Deprecated('Use endCallRequestDescriptor instead')
const EndCallRequest$json = {
  '1': 'EndCallRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'call_id', '3': 2, '4': 1, '5': 9, '10': 'callId'},
    {
      '1': 'reason',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.megachat.v1.CallEnd',
      '10': 'reason'
    },
  ],
};

/// Descriptor for `EndCallRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List endCallRequestDescriptor = $convert.base64Decode(
    'Cg5FbmRDYWxsUmVxdWVzdBIdCgphY2NvdW50X2lkGAEgASgJUglhY2NvdW50SWQSFwoHY2FsbF'
    '9pZBgCIAEoCVIGY2FsbElkEiwKBnJlYXNvbhgDIAEoDjIULm1lZ2FjaGF0LnYxLkNhbGxFbmRS'
    'BnJlYXNvbg==');

@$core.Deprecated('Use callSignalRequestDescriptor instead')
const CallSignalRequest$json = {
  '1': 'CallSignalRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'call_id', '3': 2, '4': 1, '5': 9, '10': 'callId'},
    {
      '1': 'signal',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.CallSignal',
      '10': 'signal'
    },
  ],
};

/// Descriptor for `CallSignalRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List callSignalRequestDescriptor = $convert.base64Decode(
    'ChFDYWxsU2lnbmFsUmVxdWVzdBIdCgphY2NvdW50X2lkGAEgASgJUglhY2NvdW50SWQSFwoHY2'
    'FsbF9pZBgCIAEoCVIGY2FsbElkEi8KBnNpZ25hbBgDIAEoCzIXLm1lZ2FjaGF0LnYxLkNhbGxT'
    'aWduYWxSBnNpZ25hbA==');

@$core.Deprecated('Use callMediaRequestDescriptor instead')
const CallMediaRequest$json = {
  '1': 'CallMediaRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'call_id', '3': 2, '4': 1, '5': 9, '10': 'callId'},
    {
      '1': 'media',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.CallMedia',
      '10': 'media'
    },
  ],
};

/// Descriptor for `CallMediaRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List callMediaRequestDescriptor = $convert.base64Decode(
    'ChBDYWxsTWVkaWFSZXF1ZXN0Eh0KCmFjY291bnRfaWQYASABKAlSCWFjY291bnRJZBIXCgdjYW'
    'xsX2lkGAIgASgJUgZjYWxsSWQSLAoFbWVkaWEYAyABKAsyFi5tZWdhY2hhdC52MS5DYWxsTWVk'
    'aWFSBW1lZGlh');

@$core.Deprecated('Use callsReplyDescriptor instead')
const CallsReply$json = {
  '1': 'CallsReply',
  '2': [
    {
      '1': 'calls',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Call',
      '10': 'calls'
    },
  ],
};

/// Descriptor for `CallsReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List callsReplyDescriptor = $convert.base64Decode(
    'CgpDYWxsc1JlcGx5EicKBWNhbGxzGAEgAygLMhEubWVnYWNoYXQudjEuQ2FsbFIFY2FsbHM=');

@$core.Deprecated('Use callSignalsRequestDescriptor instead')
const CallSignalsRequest$json = {
  '1': 'CallSignalsRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'call_id', '3': 2, '4': 1, '5': 9, '10': 'callId'},
    {'1': 'after', '3': 3, '4': 1, '5': 3, '10': 'after'},
  ],
};

/// Descriptor for `CallSignalsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List callSignalsRequestDescriptor = $convert.base64Decode(
    'ChJDYWxsU2lnbmFsc1JlcXVlc3QSHQoKYWNjb3VudF9pZBgBIAEoCVIJYWNjb3VudElkEhcKB2'
    'NhbGxfaWQYAiABKAlSBmNhbGxJZBIUCgVhZnRlchgDIAEoA1IFYWZ0ZXI=');

@$core.Deprecated('Use callSignalsReplyDescriptor instead')
const CallSignalsReply$json = {
  '1': 'CallSignalsReply',
  '2': [
    {
      '1': 'signals',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.CallSignal',
      '10': 'signals'
    },
  ],
};

/// Descriptor for `CallSignalsReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List callSignalsReplyDescriptor = $convert.base64Decode(
    'ChBDYWxsU2lnbmFsc1JlcGx5EjEKB3NpZ25hbHMYASADKAsyFy5tZWdhY2hhdC52MS5DYWxsU2'
    'lnbmFsUgdzaWduYWxz');

@$core.Deprecated('Use accountsRequestDescriptor instead')
const AccountsRequest$json = {
  '1': 'AccountsRequest',
  '2': [
    {'1': 'account_ids', '3': 1, '4': 3, '5': 9, '10': 'accountIds'},
  ],
};

/// Descriptor for `AccountsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List accountsRequestDescriptor = $convert.base64Decode(
    'Cg9BY2NvdW50c1JlcXVlc3QSHwoLYWNjb3VudF9pZHMYASADKAlSCmFjY291bnRJZHM=');

@$core.Deprecated('Use accountErrorDescriptor instead')
const AccountError$json = {
  '1': 'AccountError',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'error', '3': 2, '4': 1, '5': 9, '10': 'error'},
  ],
};

/// Descriptor for `AccountError`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List accountErrorDescriptor = $convert.base64Decode(
    'CgxBY2NvdW50RXJyb3ISHQoKYWNjb3VudF9pZBgBIAEoCVIJYWNjb3VudElkEhQKBWVycm9yGA'
    'IgASgJUgVlcnJvcg==');

@$core.Deprecated('Use folderDescriptor instead')
const Folder$json = {
  '1': 'Folder',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'id', '3': 2, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '10': 'name'},
    {
      '1': 'role',
      '3': 4,
      '4': 1,
      '5': 14,
      '6': '.megachat.v1.FolderRole',
      '10': 'role'
    },
    {'1': 'unread', '3': 5, '4': 1, '5': 5, '10': 'unread'},
    {'1': 'total', '3': 6, '4': 1, '5': 5, '10': 'total'},
  ],
};

/// Descriptor for `Folder`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List folderDescriptor = $convert.base64Decode(
    'CgZGb2xkZXISHQoKYWNjb3VudF9pZBgBIAEoCVIJYWNjb3VudElkEg4KAmlkGAIgASgJUgJpZB'
    'ISCgRuYW1lGAMgASgJUgRuYW1lEisKBHJvbGUYBCABKA4yFy5tZWdhY2hhdC52MS5Gb2xkZXJS'
    'b2xlUgRyb2xlEhYKBnVucmVhZBgFIAEoBVIGdW5yZWFkEhQKBXRvdGFsGAYgASgFUgV0b3RhbA'
    '==');

@$core.Deprecated('Use foldersReplyDescriptor instead')
const FoldersReply$json = {
  '1': 'FoldersReply',
  '2': [
    {
      '1': 'folders',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Folder',
      '10': 'folders'
    },
    {
      '1': 'failed',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.AccountError',
      '10': 'failed'
    },
  ],
};

/// Descriptor for `FoldersReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List foldersReplyDescriptor = $convert.base64Decode(
    'CgxGb2xkZXJzUmVwbHkSLQoHZm9sZGVycxgBIAMoCzITLm1lZ2FjaGF0LnYxLkZvbGRlclIHZm'
    '9sZGVycxIxCgZmYWlsZWQYAiADKAsyGS5tZWdhY2hhdC52MS5BY2NvdW50RXJyb3JSBmZhaWxl'
    'ZA==');

@$core.Deprecated('Use addressDescriptor instead')
const Address$json = {
  '1': 'Address',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'email', '3': 2, '4': 1, '5': 9, '10': 'email'},
  ],
};

/// Descriptor for `Address`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addressDescriptor = $convert.base64Decode(
    'CgdBZGRyZXNzEhIKBG5hbWUYASABKAlSBG5hbWUSFAoFZW1haWwYAiABKAlSBWVtYWls');

@$core.Deprecated('Use mailDescriptor instead')
const Mail$json = {
  '1': 'Mail',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'id', '3': 2, '4': 1, '5': 9, '10': 'id'},
    {'1': 'folder_id', '3': 3, '4': 1, '5': 9, '10': 'folderId'},
    {'1': 'thread_id', '3': 4, '4': 1, '5': 9, '10': 'threadId'},
    {
      '1': 'from',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.Address',
      '10': 'from'
    },
    {
      '1': 'to',
      '3': 6,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Address',
      '10': 'to'
    },
    {
      '1': 'cc',
      '3': 7,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Address',
      '10': 'cc'
    },
    {
      '1': 'reply_to',
      '3': 8,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Address',
      '10': 'replyTo'
    },
    {'1': 'subject', '3': 9, '4': 1, '5': 9, '10': 'subject'},
    {'1': 'snippet', '3': 10, '4': 1, '5': 9, '10': 'snippet'},
    {'1': 'time_ms', '3': 11, '4': 1, '5': 3, '10': 'timeMs'},
    {'1': 'unread', '3': 12, '4': 1, '5': 8, '10': 'unread'},
    {'1': 'starred', '3': 13, '4': 1, '5': 8, '10': 'starred'},
    {'1': 'has_attachments', '3': 14, '4': 1, '5': 8, '10': 'hasAttachments'},
    {'1': 'text', '3': 15, '4': 1, '5': 9, '10': 'text'},
    {'1': 'html', '3': 16, '4': 1, '5': 9, '10': 'html'},
    {
      '1': 'attachments',
      '3': 17,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Attachment',
      '10': 'attachments'
    },
    {
      '1': 'format',
      '3': 18,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.TextSpan',
      '10': 'format'
    },
  ],
};

/// Descriptor for `Mail`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List mailDescriptor = $convert.base64Decode(
    'CgRNYWlsEh0KCmFjY291bnRfaWQYASABKAlSCWFjY291bnRJZBIOCgJpZBgCIAEoCVICaWQSGw'
    'oJZm9sZGVyX2lkGAMgASgJUghmb2xkZXJJZBIbCgl0aHJlYWRfaWQYBCABKAlSCHRocmVhZElk'
    'EigKBGZyb20YBSABKAsyFC5tZWdhY2hhdC52MS5BZGRyZXNzUgRmcm9tEiQKAnRvGAYgAygLMh'
    'QubWVnYWNoYXQudjEuQWRkcmVzc1ICdG8SJAoCY2MYByADKAsyFC5tZWdhY2hhdC52MS5BZGRy'
    'ZXNzUgJjYxIvCghyZXBseV90bxgIIAMoCzIULm1lZ2FjaGF0LnYxLkFkZHJlc3NSB3JlcGx5VG'
    '8SGAoHc3ViamVjdBgJIAEoCVIHc3ViamVjdBIYCgdzbmlwcGV0GAogASgJUgdzbmlwcGV0EhcK'
    'B3RpbWVfbXMYCyABKANSBnRpbWVNcxIWCgZ1bnJlYWQYDCABKAhSBnVucmVhZBIYCgdzdGFycm'
    'VkGA0gASgIUgdzdGFycmVkEicKD2hhc19hdHRhY2htZW50cxgOIAEoCFIOaGFzQXR0YWNobWVu'
    'dHMSEgoEdGV4dBgPIAEoCVIEdGV4dBISCgRodG1sGBAgASgJUgRodG1sEjkKC2F0dGFjaG1lbn'
    'RzGBEgAygLMhcubWVnYWNoYXQudjEuQXR0YWNobWVudFILYXR0YWNobWVudHMSLQoGZm9ybWF0'
    'GBIgAygLMhUubWVnYWNoYXQudjEuVGV4dFNwYW5SBmZvcm1hdA==');

@$core.Deprecated('Use mailListRequestDescriptor instead')
const MailListRequest$json = {
  '1': 'MailListRequest',
  '2': [
    {'1': 'account_ids', '3': 1, '4': 3, '5': 9, '10': 'accountIds'},
    {
      '1': 'role',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.megachat.v1.FolderRole',
      '10': 'role'
    },
    {'1': 'folder_id', '3': 3, '4': 1, '5': 9, '10': 'folderId'},
    {'1': 'cursor', '3': 4, '4': 1, '5': 9, '10': 'cursor'},
    {'1': 'limit', '3': 5, '4': 1, '5': 5, '10': 'limit'},
  ],
};

/// Descriptor for `MailListRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List mailListRequestDescriptor = $convert.base64Decode(
    'Cg9NYWlsTGlzdFJlcXVlc3QSHwoLYWNjb3VudF9pZHMYASADKAlSCmFjY291bnRJZHMSKwoEcm'
    '9sZRgCIAEoDjIXLm1lZ2FjaGF0LnYxLkZvbGRlclJvbGVSBHJvbGUSGwoJZm9sZGVyX2lkGAMg'
    'ASgJUghmb2xkZXJJZBIWCgZjdXJzb3IYBCABKAlSBmN1cnNvchIUCgVsaW1pdBgFIAEoBVIFbG'
    'ltaXQ=');

@$core.Deprecated('Use mailListReplyDescriptor instead')
const MailListReply$json = {
  '1': 'MailListReply',
  '2': [
    {
      '1': 'mails',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Mail',
      '10': 'mails'
    },
    {'1': 'more', '3': 2, '4': 1, '5': 8, '10': 'more'},
    {'1': 'next', '3': 3, '4': 1, '5': 9, '10': 'next'},
    {
      '1': 'failed',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.AccountError',
      '10': 'failed'
    },
  ],
};

/// Descriptor for `MailListReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List mailListReplyDescriptor = $convert.base64Decode(
    'Cg1NYWlsTGlzdFJlcGx5EicKBW1haWxzGAEgAygLMhEubWVnYWNoYXQudjEuTWFpbFIFbWFpbH'
    'MSEgoEbW9yZRgCIAEoCFIEbW9yZRISCgRuZXh0GAMgASgJUgRuZXh0EjEKBmZhaWxlZBgEIAMo'
    'CzIZLm1lZ2FjaGF0LnYxLkFjY291bnRFcnJvclIGZmFpbGVk');

@$core.Deprecated('Use getMailRequestDescriptor instead')
const GetMailRequest$json = {
  '1': 'GetMailRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'id', '3': 2, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetMailRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getMailRequestDescriptor = $convert.base64Decode(
    'Cg5HZXRNYWlsUmVxdWVzdBIdCgphY2NvdW50X2lkGAEgASgJUglhY2NvdW50SWQSDgoCaWQYAi'
    'ABKAlSAmlk');

@$core.Deprecated('Use sendMailRequestDescriptor instead')
const SendMailRequest$json = {
  '1': 'SendMailRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {
      '1': 'to',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Address',
      '10': 'to'
    },
    {
      '1': 'cc',
      '3': 3,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Address',
      '10': 'cc'
    },
    {
      '1': 'bcc',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Address',
      '10': 'bcc'
    },
    {'1': 'subject', '3': 5, '4': 1, '5': 9, '10': 'subject'},
    {'1': 'text', '3': 6, '4': 1, '5': 9, '10': 'text'},
    {'1': 'reply_to_id', '3': 7, '4': 1, '5': 9, '10': 'replyToId'},
  ],
};

/// Descriptor for `SendMailRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List sendMailRequestDescriptor = $convert.base64Decode(
    'Cg9TZW5kTWFpbFJlcXVlc3QSHQoKYWNjb3VudF9pZBgBIAEoCVIJYWNjb3VudElkEiQKAnRvGA'
    'IgAygLMhQubWVnYWNoYXQudjEuQWRkcmVzc1ICdG8SJAoCY2MYAyADKAsyFC5tZWdhY2hhdC52'
    'MS5BZGRyZXNzUgJjYxImCgNiY2MYBCADKAsyFC5tZWdhY2hhdC52MS5BZGRyZXNzUgNiY2MSGA'
    'oHc3ViamVjdBgFIAEoCVIHc3ViamVjdBISCgR0ZXh0GAYgASgJUgR0ZXh0Eh4KC3JlcGx5X3Rv'
    'X2lkGAcgASgJUglyZXBseVRvSWQ=');

@$core.Deprecated('Use updateMailRequestDescriptor instead')
const UpdateMailRequest$json = {
  '1': 'UpdateMailRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'ids', '3': 2, '4': 3, '5': 9, '10': 'ids'},
    {
      '1': 'action',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.megachat.v1.MailAction',
      '10': 'action'
    },
  ],
};

/// Descriptor for `UpdateMailRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateMailRequestDescriptor = $convert.base64Decode(
    'ChFVcGRhdGVNYWlsUmVxdWVzdBIdCgphY2NvdW50X2lkGAEgASgJUglhY2NvdW50SWQSEAoDaW'
    'RzGAIgAygJUgNpZHMSLwoGYWN0aW9uGAMgASgOMhcubWVnYWNoYXQudjEuTWFpbEFjdGlvblIG'
    'YWN0aW9u');

@$core.Deprecated('Use moveMailRequestDescriptor instead')
const MoveMailRequest$json = {
  '1': 'MoveMailRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'ids', '3': 2, '4': 3, '5': 9, '10': 'ids'},
    {'1': 'from_folder_id', '3': 3, '4': 1, '5': 9, '10': 'fromFolderId'},
    {'1': 'to_folder_id', '3': 4, '4': 1, '5': 9, '10': 'toFolderId'},
  ],
};

/// Descriptor for `MoveMailRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List moveMailRequestDescriptor = $convert.base64Decode(
    'Cg9Nb3ZlTWFpbFJlcXVlc3QSHQoKYWNjb3VudF9pZBgBIAEoCVIJYWNjb3VudElkEhAKA2lkcx'
    'gCIAMoCVIDaWRzEiQKDmZyb21fZm9sZGVyX2lkGAMgASgJUgxmcm9tRm9sZGVySWQSIAoMdG9f'
    'Zm9sZGVyX2lkGAQgASgJUgp0b0ZvbGRlcklk');

@$core.Deprecated('Use calendarDescriptor instead')
const Calendar$json = {
  '1': 'Calendar',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'id', '3': 2, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '10': 'name'},
    {'1': 'color', '3': 4, '4': 1, '5': 5, '10': 'color'},
    {'1': 'read_only', '3': 5, '4': 1, '5': 8, '10': 'readOnly'},
    {'1': 'primary', '3': 6, '4': 1, '5': 8, '10': 'primary'},
  ],
};

/// Descriptor for `Calendar`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List calendarDescriptor = $convert.base64Decode(
    'CghDYWxlbmRhchIdCgphY2NvdW50X2lkGAEgASgJUglhY2NvdW50SWQSDgoCaWQYAiABKAlSAm'
    'lkEhIKBG5hbWUYAyABKAlSBG5hbWUSFAoFY29sb3IYBCABKAVSBWNvbG9yEhsKCXJlYWRfb25s'
    'eRgFIAEoCFIIcmVhZE9ubHkSGAoHcHJpbWFyeRgGIAEoCFIHcHJpbWFyeQ==');

@$core.Deprecated('Use calendarsReplyDescriptor instead')
const CalendarsReply$json = {
  '1': 'CalendarsReply',
  '2': [
    {
      '1': 'calendars',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Calendar',
      '10': 'calendars'
    },
    {
      '1': 'failed',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.AccountError',
      '10': 'failed'
    },
  ],
};

/// Descriptor for `CalendarsReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List calendarsReplyDescriptor = $convert.base64Decode(
    'Cg5DYWxlbmRhcnNSZXBseRIzCgljYWxlbmRhcnMYASADKAsyFS5tZWdhY2hhdC52MS5DYWxlbm'
    'RhclIJY2FsZW5kYXJzEjEKBmZhaWxlZBgCIAMoCzIZLm1lZ2FjaGF0LnYxLkFjY291bnRFcnJv'
    'clIGZmFpbGVk');

@$core.Deprecated('Use attendeeDescriptor instead')
const Attendee$json = {
  '1': 'Attendee',
  '2': [
    {
      '1': 'address',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.Address',
      '10': 'address'
    },
    {
      '1': 'status',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.megachat.v1.Rsvp',
      '10': 'status'
    },
    {'1': 'self', '3': 3, '4': 1, '5': 8, '10': 'self'},
  ],
};

/// Descriptor for `Attendee`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List attendeeDescriptor = $convert.base64Decode(
    'CghBdHRlbmRlZRIuCgdhZGRyZXNzGAEgASgLMhQubWVnYWNoYXQudjEuQWRkcmVzc1IHYWRkcm'
    'VzcxIpCgZzdGF0dXMYAiABKA4yES5tZWdhY2hhdC52MS5Sc3ZwUgZzdGF0dXMSEgoEc2VsZhgD'
    'IAEoCFIEc2VsZg==');

@$core.Deprecated('Use calendarEventDescriptor instead')
const CalendarEvent$json = {
  '1': 'CalendarEvent',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'calendar_id', '3': 2, '4': 1, '5': 9, '10': 'calendarId'},
    {'1': 'id', '3': 3, '4': 1, '5': 9, '10': 'id'},
    {'1': 'title', '3': 4, '4': 1, '5': 9, '10': 'title'},
    {'1': 'location', '3': 5, '4': 1, '5': 9, '10': 'location'},
    {'1': 'description', '3': 6, '4': 1, '5': 9, '10': 'description'},
    {'1': 'start_ms', '3': 7, '4': 1, '5': 3, '10': 'startMs'},
    {'1': 'end_ms', '3': 8, '4': 1, '5': 3, '10': 'endMs'},
    {'1': 'all_day', '3': 9, '4': 1, '5': 8, '10': 'allDay'},
    {'1': 'recurring', '3': 10, '4': 1, '5': 8, '10': 'recurring'},
    {
      '1': 'organizer',
      '3': 11,
      '4': 1,
      '5': 11,
      '6': '.megachat.v1.Address',
      '10': 'organizer'
    },
    {
      '1': 'attendees',
      '3': 12,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.Attendee',
      '10': 'attendees'
    },
    {'1': 'link', '3': 13, '4': 1, '5': 9, '10': 'link'},
  ],
};

/// Descriptor for `CalendarEvent`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List calendarEventDescriptor = $convert.base64Decode(
    'Cg1DYWxlbmRhckV2ZW50Eh0KCmFjY291bnRfaWQYASABKAlSCWFjY291bnRJZBIfCgtjYWxlbm'
    'Rhcl9pZBgCIAEoCVIKY2FsZW5kYXJJZBIOCgJpZBgDIAEoCVICaWQSFAoFdGl0bGUYBCABKAlS'
    'BXRpdGxlEhoKCGxvY2F0aW9uGAUgASgJUghsb2NhdGlvbhIgCgtkZXNjcmlwdGlvbhgGIAEoCV'
    'ILZGVzY3JpcHRpb24SGQoIc3RhcnRfbXMYByABKANSB3N0YXJ0TXMSFQoGZW5kX21zGAggASgD'
    'UgVlbmRNcxIXCgdhbGxfZGF5GAkgASgIUgZhbGxEYXkSHAoJcmVjdXJyaW5nGAogASgIUglyZW'
    'N1cnJpbmcSMgoJb3JnYW5pemVyGAsgASgLMhQubWVnYWNoYXQudjEuQWRkcmVzc1IJb3JnYW5p'
    'emVyEjMKCWF0dGVuZGVlcxgMIAMoCzIVLm1lZ2FjaGF0LnYxLkF0dGVuZGVlUglhdHRlbmRlZX'
    'MSEgoEbGluaxgNIAEoCVIEbGluaw==');

@$core.Deprecated('Use calendarRefDescriptor instead')
const CalendarRef$json = {
  '1': 'CalendarRef',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'calendar_id', '3': 2, '4': 1, '5': 9, '10': 'calendarId'},
  ],
};

/// Descriptor for `CalendarRef`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List calendarRefDescriptor = $convert.base64Decode(
    'CgtDYWxlbmRhclJlZhIdCgphY2NvdW50X2lkGAEgASgJUglhY2NvdW50SWQSHwoLY2FsZW5kYX'
    'JfaWQYAiABKAlSCmNhbGVuZGFySWQ=');

@$core.Deprecated('Use eventsRequestDescriptor instead')
const EventsRequest$json = {
  '1': 'EventsRequest',
  '2': [
    {'1': 'account_ids', '3': 1, '4': 3, '5': 9, '10': 'accountIds'},
    {
      '1': 'calendars',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.CalendarRef',
      '10': 'calendars'
    },
    {'1': 'from_ms', '3': 3, '4': 1, '5': 3, '10': 'fromMs'},
    {'1': 'to_ms', '3': 4, '4': 1, '5': 3, '10': 'toMs'},
  ],
};

/// Descriptor for `EventsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List eventsRequestDescriptor = $convert.base64Decode(
    'Cg1FdmVudHNSZXF1ZXN0Eh8KC2FjY291bnRfaWRzGAEgAygJUgphY2NvdW50SWRzEjYKCWNhbG'
    'VuZGFycxgCIAMoCzIYLm1lZ2FjaGF0LnYxLkNhbGVuZGFyUmVmUgljYWxlbmRhcnMSFwoHZnJv'
    'bV9tcxgDIAEoA1IGZnJvbU1zEhMKBXRvX21zGAQgASgDUgR0b01z');

@$core.Deprecated('Use eventsReplyDescriptor instead')
const EventsReply$json = {
  '1': 'EventsReply',
  '2': [
    {
      '1': 'events',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.CalendarEvent',
      '10': 'events'
    },
    {
      '1': 'failed',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.megachat.v1.AccountError',
      '10': 'failed'
    },
  ],
};

/// Descriptor for `EventsReply`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List eventsReplyDescriptor = $convert.base64Decode(
    'CgtFdmVudHNSZXBseRIyCgZldmVudHMYASADKAsyGi5tZWdhY2hhdC52MS5DYWxlbmRhckV2ZW'
    '50UgZldmVudHMSMQoGZmFpbGVkGAIgAygLMhkubWVnYWNoYXQudjEuQWNjb3VudEVycm9yUgZm'
    'YWlsZWQ=');

@$core.Deprecated('Use deleteEventRequestDescriptor instead')
const DeleteEventRequest$json = {
  '1': 'DeleteEventRequest',
  '2': [
    {'1': 'account_id', '3': 1, '4': 1, '5': 9, '10': 'accountId'},
    {'1': 'calendar_id', '3': 2, '4': 1, '5': 9, '10': 'calendarId'},
    {'1': 'id', '3': 3, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteEventRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteEventRequestDescriptor = $convert.base64Decode(
    'ChJEZWxldGVFdmVudFJlcXVlc3QSHQoKYWNjb3VudF9pZBgBIAEoCVIJYWNjb3VudElkEh8KC2'
    'NhbGVuZGFyX2lkGAIgASgJUgpjYWxlbmRhcklkEg4KAmlkGAMgASgJUgJpZA==');
