// This is a generated file - do not edit.
//
// Generated from megachat.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// What a provider can do. Only MESSAGES exists today; the others are
/// reserved so the app and the core agree on them when they arrive.
class Capability extends $pb.ProtobufEnum {
  static const Capability CAPABILITY_UNSPECIFIED =
      Capability._(0, _omitEnumNames ? '' : 'CAPABILITY_UNSPECIFIED');
  static const Capability CAPABILITY_MESSAGES =
      Capability._(1, _omitEnumNames ? '' : 'CAPABILITY_MESSAGES');
  static const Capability CAPABILITY_CALLS =
      Capability._(2, _omitEnumNames ? '' : 'CAPABILITY_CALLS');
  static const Capability CAPABILITY_GROUP_CALLS =
      Capability._(3, _omitEnumNames ? '' : 'CAPABILITY_GROUP_CALLS');
  static const Capability CAPABILITY_MAIL =
      Capability._(4, _omitEnumNames ? '' : 'CAPABILITY_MAIL');
  static const Capability CAPABILITY_CALENDAR =
      Capability._(5, _omitEnumNames ? '' : 'CAPABILITY_CALENDAR');

  /// Sessions: Account.capabilities tells what an account's session does.
  static const Capability CAPABILITY_THREADS =
      Capability._(6, _omitEnumNames ? '' : 'CAPABILITY_THREADS');
  static const Capability CAPABILITY_REACTIONS =
      Capability._(7, _omitEnumNames ? '' : 'CAPABILITY_REACTIONS');
  static const Capability CAPABILITY_MEDIA =
      Capability._(8, _omitEnumNames ? '' : 'CAPABILITY_MEDIA');
  static const Capability CAPABILITY_TOPICS =
      Capability._(9, _omitEnumNames ? '' : 'CAPABILITY_TOPICS');
  static const Capability CAPABILITY_FORWARD =
      Capability._(10, _omitEnumNames ? '' : 'CAPABILITY_FORWARD');

  static const $core.List<Capability> values = <Capability>[
    CAPABILITY_UNSPECIFIED,
    CAPABILITY_MESSAGES,
    CAPABILITY_CALLS,
    CAPABILITY_GROUP_CALLS,
    CAPABILITY_MAIL,
    CAPABILITY_CALENDAR,
    CAPABILITY_THREADS,
    CAPABILITY_REACTIONS,
    CAPABILITY_MEDIA,
    CAPABILITY_TOPICS,
    CAPABILITY_FORWARD,
  ];

  static final $core.List<Capability?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 10);
  static Capability? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const Capability._(super.value, super.name);
}

class AccountStatus extends $pb.ProtobufEnum {
  static const AccountStatus ACCOUNT_STATUS_UNSPECIFIED =
      AccountStatus._(0, _omitEnumNames ? '' : 'ACCOUNT_STATUS_UNSPECIFIED');
  static const AccountStatus ACCOUNT_STATUS_DISABLED =
      AccountStatus._(1, _omitEnumNames ? '' : 'ACCOUNT_STATUS_DISABLED');
  static const AccountStatus ACCOUNT_STATUS_CONNECTING =
      AccountStatus._(2, _omitEnumNames ? '' : 'ACCOUNT_STATUS_CONNECTING');
  static const AccountStatus ACCOUNT_STATUS_ONLINE =
      AccountStatus._(3, _omitEnumNames ? '' : 'ACCOUNT_STATUS_ONLINE');
  static const AccountStatus ACCOUNT_STATUS_OFFLINE =
      AccountStatus._(4, _omitEnumNames ? '' : 'ACCOUNT_STATUS_OFFLINE');
  static const AccountStatus ACCOUNT_STATUS_NEEDS_LOGIN =
      AccountStatus._(5, _omitEnumNames ? '' : 'ACCOUNT_STATUS_NEEDS_LOGIN');

  static const $core.List<AccountStatus> values = <AccountStatus>[
    ACCOUNT_STATUS_UNSPECIFIED,
    ACCOUNT_STATUS_DISABLED,
    ACCOUNT_STATUS_CONNECTING,
    ACCOUNT_STATUS_ONLINE,
    ACCOUNT_STATUS_OFFLINE,
    ACCOUNT_STATUS_NEEDS_LOGIN,
  ];

  static final $core.List<AccountStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 5);
  static AccountStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AccountStatus._(super.value, super.name);
}

class ChatKind extends $pb.ProtobufEnum {
  static const ChatKind CHAT_KIND_UNSPECIFIED =
      ChatKind._(0, _omitEnumNames ? '' : 'CHAT_KIND_UNSPECIFIED');
  static const ChatKind CHAT_KIND_DIRECT =
      ChatKind._(1, _omitEnumNames ? '' : 'CHAT_KIND_DIRECT');
  static const ChatKind CHAT_KIND_GROUP =
      ChatKind._(2, _omitEnumNames ? '' : 'CHAT_KIND_GROUP');
  static const ChatKind CHAT_KIND_CHANNEL =
      ChatKind._(3, _omitEnumNames ? '' : 'CHAT_KIND_CHANNEL');

  static const $core.List<ChatKind> values = <ChatKind>[
    CHAT_KIND_UNSPECIFIED,
    CHAT_KIND_DIRECT,
    CHAT_KIND_GROUP,
    CHAT_KIND_CHANNEL,
  ];

  static final $core.List<ChatKind?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static ChatKind? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ChatKind._(super.value, super.name);
}

class AttachmentKind extends $pb.ProtobufEnum {
  static const AttachmentKind ATTACHMENT_KIND_UNSPECIFIED =
      AttachmentKind._(0, _omitEnumNames ? '' : 'ATTACHMENT_KIND_UNSPECIFIED');
  static const AttachmentKind ATTACHMENT_KIND_FILE =
      AttachmentKind._(1, _omitEnumNames ? '' : 'ATTACHMENT_KIND_FILE');
  static const AttachmentKind ATTACHMENT_KIND_IMAGE =
      AttachmentKind._(2, _omitEnumNames ? '' : 'ATTACHMENT_KIND_IMAGE');
  static const AttachmentKind ATTACHMENT_KIND_VIDEO =
      AttachmentKind._(3, _omitEnumNames ? '' : 'ATTACHMENT_KIND_VIDEO');
  static const AttachmentKind ATTACHMENT_KIND_AUDIO =
      AttachmentKind._(4, _omitEnumNames ? '' : 'ATTACHMENT_KIND_AUDIO');
  static const AttachmentKind ATTACHMENT_KIND_VOICE =
      AttachmentKind._(5, _omitEnumNames ? '' : 'ATTACHMENT_KIND_VOICE');
  static const AttachmentKind ATTACHMENT_KIND_STICKER =
      AttachmentKind._(6, _omitEnumNames ? '' : 'ATTACHMENT_KIND_STICKER');

  static const $core.List<AttachmentKind> values = <AttachmentKind>[
    ATTACHMENT_KIND_UNSPECIFIED,
    ATTACHMENT_KIND_FILE,
    ATTACHMENT_KIND_IMAGE,
    ATTACHMENT_KIND_VIDEO,
    ATTACHMENT_KIND_AUDIO,
    ATTACHMENT_KIND_VOICE,
    ATTACHMENT_KIND_STICKER,
  ];

  static final $core.List<AttachmentKind?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 6);
  static AttachmentKind? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AttachmentKind._(super.value, super.name);
}

class CallState extends $pb.ProtobufEnum {
  static const CallState CALL_STATE_UNSPECIFIED =
      CallState._(0, _omitEnumNames ? '' : 'CALL_STATE_UNSPECIFIED');
  static const CallState CALL_STATE_WAITING =
      CallState._(1, _omitEnumNames ? '' : 'CALL_STATE_WAITING');
  static const CallState CALL_STATE_RINGING =
      CallState._(2, _omitEnumNames ? '' : 'CALL_STATE_RINGING');
  static const CallState CALL_STATE_CONNECTING =
      CallState._(3, _omitEnumNames ? '' : 'CALL_STATE_CONNECTING');
  static const CallState CALL_STATE_ACTIVE =
      CallState._(4, _omitEnumNames ? '' : 'CALL_STATE_ACTIVE');
  static const CallState CALL_STATE_ENDED =
      CallState._(5, _omitEnumNames ? '' : 'CALL_STATE_ENDED');

  static const $core.List<CallState> values = <CallState>[
    CALL_STATE_UNSPECIFIED,
    CALL_STATE_WAITING,
    CALL_STATE_RINGING,
    CALL_STATE_CONNECTING,
    CALL_STATE_ACTIVE,
    CALL_STATE_ENDED,
  ];

  static final $core.List<CallState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 5);
  static CallState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const CallState._(super.value, super.name);
}

class CallEnd extends $pb.ProtobufEnum {
  static const CallEnd CALL_END_UNSPECIFIED =
      CallEnd._(0, _omitEnumNames ? '' : 'CALL_END_UNSPECIFIED');
  static const CallEnd CALL_END_HANGUP =
      CallEnd._(1, _omitEnumNames ? '' : 'CALL_END_HANGUP');
  static const CallEnd CALL_END_DECLINED =
      CallEnd._(2, _omitEnumNames ? '' : 'CALL_END_DECLINED');
  static const CallEnd CALL_END_MISSED =
      CallEnd._(3, _omitEnumNames ? '' : 'CALL_END_MISSED');
  static const CallEnd CALL_END_BUSY =
      CallEnd._(4, _omitEnumNames ? '' : 'CALL_END_BUSY');
  static const CallEnd CALL_END_FAILED =
      CallEnd._(5, _omitEnumNames ? '' : 'CALL_END_FAILED');

  static const $core.List<CallEnd> values = <CallEnd>[
    CALL_END_UNSPECIFIED,
    CALL_END_HANGUP,
    CALL_END_DECLINED,
    CALL_END_MISSED,
    CALL_END_BUSY,
    CALL_END_FAILED,
  ];

  static final $core.List<CallEnd?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 5);
  static CallEnd? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const CallEnd._(super.value, super.name);
}

class SignalKind extends $pb.ProtobufEnum {
  static const SignalKind SIGNAL_KIND_UNSPECIFIED =
      SignalKind._(0, _omitEnumNames ? '' : 'SIGNAL_KIND_UNSPECIFIED');
  static const SignalKind SIGNAL_KIND_OFFER =
      SignalKind._(1, _omitEnumNames ? '' : 'SIGNAL_KIND_OFFER');
  static const SignalKind SIGNAL_KIND_ANSWER =
      SignalKind._(2, _omitEnumNames ? '' : 'SIGNAL_KIND_ANSWER');
  static const SignalKind SIGNAL_KIND_CANDIDATE =
      SignalKind._(3, _omitEnumNames ? '' : 'SIGNAL_KIND_CANDIDATE');
  static const SignalKind SIGNAL_KIND_DATA =
      SignalKind._(4, _omitEnumNames ? '' : 'SIGNAL_KIND_DATA');
  static const SignalKind SIGNAL_KIND_CONNECTED =
      SignalKind._(5, _omitEnumNames ? '' : 'SIGNAL_KIND_CONNECTED');
  static const SignalKind SIGNAL_KIND_FAILED =
      SignalKind._(6, _omitEnumNames ? '' : 'SIGNAL_KIND_FAILED');

  static const $core.List<SignalKind> values = <SignalKind>[
    SIGNAL_KIND_UNSPECIFIED,
    SIGNAL_KIND_OFFER,
    SIGNAL_KIND_ANSWER,
    SIGNAL_KIND_CANDIDATE,
    SIGNAL_KIND_DATA,
    SIGNAL_KIND_CONNECTED,
    SIGNAL_KIND_FAILED,
  ];

  static final $core.List<SignalKind?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 6);
  static SignalKind? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const SignalKind._(super.value, super.name);
}

class FolderRole extends $pb.ProtobufEnum {
  static const FolderRole FOLDER_ROLE_OTHER =
      FolderRole._(0, _omitEnumNames ? '' : 'FOLDER_ROLE_OTHER');
  static const FolderRole FOLDER_ROLE_INBOX =
      FolderRole._(1, _omitEnumNames ? '' : 'FOLDER_ROLE_INBOX');
  static const FolderRole FOLDER_ROLE_SENT =
      FolderRole._(2, _omitEnumNames ? '' : 'FOLDER_ROLE_SENT');
  static const FolderRole FOLDER_ROLE_DRAFTS =
      FolderRole._(3, _omitEnumNames ? '' : 'FOLDER_ROLE_DRAFTS');
  static const FolderRole FOLDER_ROLE_TRASH =
      FolderRole._(4, _omitEnumNames ? '' : 'FOLDER_ROLE_TRASH');
  static const FolderRole FOLDER_ROLE_SPAM =
      FolderRole._(5, _omitEnumNames ? '' : 'FOLDER_ROLE_SPAM');
  static const FolderRole FOLDER_ROLE_ARCHIVE =
      FolderRole._(6, _omitEnumNames ? '' : 'FOLDER_ROLE_ARCHIVE');

  static const $core.List<FolderRole> values = <FolderRole>[
    FOLDER_ROLE_OTHER,
    FOLDER_ROLE_INBOX,
    FOLDER_ROLE_SENT,
    FOLDER_ROLE_DRAFTS,
    FOLDER_ROLE_TRASH,
    FOLDER_ROLE_SPAM,
    FOLDER_ROLE_ARCHIVE,
  ];

  static final $core.List<FolderRole?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 6);
  static FolderRole? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const FolderRole._(super.value, super.name);
}

class MailAction extends $pb.ProtobufEnum {
  static const MailAction MAIL_ACTION_UNSPECIFIED =
      MailAction._(0, _omitEnumNames ? '' : 'MAIL_ACTION_UNSPECIFIED');
  static const MailAction MAIL_ACTION_READ =
      MailAction._(1, _omitEnumNames ? '' : 'MAIL_ACTION_READ');
  static const MailAction MAIL_ACTION_UNREAD =
      MailAction._(2, _omitEnumNames ? '' : 'MAIL_ACTION_UNREAD');
  static const MailAction MAIL_ACTION_STAR =
      MailAction._(3, _omitEnumNames ? '' : 'MAIL_ACTION_STAR');
  static const MailAction MAIL_ACTION_UNSTAR =
      MailAction._(4, _omitEnumNames ? '' : 'MAIL_ACTION_UNSTAR');
  static const MailAction MAIL_ACTION_TRASH =
      MailAction._(5, _omitEnumNames ? '' : 'MAIL_ACTION_TRASH');

  static const $core.List<MailAction> values = <MailAction>[
    MAIL_ACTION_UNSPECIFIED,
    MAIL_ACTION_READ,
    MAIL_ACTION_UNREAD,
    MAIL_ACTION_STAR,
    MAIL_ACTION_UNSTAR,
    MAIL_ACTION_TRASH,
  ];

  static final $core.List<MailAction?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 5);
  static MailAction? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const MailAction._(super.value, super.name);
}

class Rsvp extends $pb.ProtobufEnum {
  static const Rsvp RSVP_UNSPECIFIED =
      Rsvp._(0, _omitEnumNames ? '' : 'RSVP_UNSPECIFIED');
  static const Rsvp RSVP_NEEDS_ACTION =
      Rsvp._(1, _omitEnumNames ? '' : 'RSVP_NEEDS_ACTION');
  static const Rsvp RSVP_ACCEPTED =
      Rsvp._(2, _omitEnumNames ? '' : 'RSVP_ACCEPTED');
  static const Rsvp RSVP_TENTATIVE =
      Rsvp._(3, _omitEnumNames ? '' : 'RSVP_TENTATIVE');
  static const Rsvp RSVP_DECLINED =
      Rsvp._(4, _omitEnumNames ? '' : 'RSVP_DECLINED');

  static const $core.List<Rsvp> values = <Rsvp>[
    RSVP_UNSPECIFIED,
    RSVP_NEEDS_ACTION,
    RSVP_ACCEPTED,
    RSVP_TENTATIVE,
    RSVP_DECLINED,
  ];

  static final $core.List<Rsvp?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static Rsvp? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const Rsvp._(super.value, super.name);
}

class Event_Kind extends $pb.ProtobufEnum {
  static const Event_Kind KIND_UNSPECIFIED =
      Event_Kind._(0, _omitEnumNames ? '' : 'KIND_UNSPECIFIED');
  static const Event_Kind KIND_STATE =
      Event_Kind._(1, _omitEnumNames ? '' : 'KIND_STATE');
  static const Event_Kind KIND_ACCOUNT =
      Event_Kind._(2, _omitEnumNames ? '' : 'KIND_ACCOUNT');
  static const Event_Kind KIND_CHAT =
      Event_Kind._(3, _omitEnumNames ? '' : 'KIND_CHAT');
  static const Event_Kind KIND_MESSAGE_NEW =
      Event_Kind._(4, _omitEnumNames ? '' : 'KIND_MESSAGE_NEW');
  static const Event_Kind KIND_MESSAGE_EDIT =
      Event_Kind._(5, _omitEnumNames ? '' : 'KIND_MESSAGE_EDIT');
  static const Event_Kind KIND_MESSAGE_DELETE =
      Event_Kind._(6, _omitEnumNames ? '' : 'KIND_MESSAGE_DELETE');
  static const Event_Kind KIND_TYPING =
      Event_Kind._(7, _omitEnumNames ? '' : 'KIND_TYPING');
  static const Event_Kind KIND_CHAT_REMOVED =
      Event_Kind._(8, _omitEnumNames ? '' : 'KIND_CHAT_REMOVED');
  static const Event_Kind KIND_THREAD_MESSAGE =
      Event_Kind._(9, _omitEnumNames ? '' : 'KIND_THREAD_MESSAGE');

  /// (chat: the cached chat, when there is one)
  static const Event_Kind KIND_CALL =
      Event_Kind._(10, _omitEnumNames ? '' : 'KIND_CALL');
  static const Event_Kind KIND_CALL_SIGNAL =
      Event_Kind._(11, _omitEnumNames ? '' : 'KIND_CALL_SIGNAL');
  static const Event_Kind KIND_MAIL_NEW =
      Event_Kind._(20, _omitEnumNames ? '' : 'KIND_MAIL_NEW');
  static const Event_Kind KIND_MAIL =
      Event_Kind._(21, _omitEnumNames ? '' : 'KIND_MAIL');
  static const Event_Kind KIND_CALENDAR =
      Event_Kind._(22, _omitEnumNames ? '' : 'KIND_CALENDAR');

  static const $core.List<Event_Kind> values = <Event_Kind>[
    KIND_UNSPECIFIED,
    KIND_STATE,
    KIND_ACCOUNT,
    KIND_CHAT,
    KIND_MESSAGE_NEW,
    KIND_MESSAGE_EDIT,
    KIND_MESSAGE_DELETE,
    KIND_TYPING,
    KIND_CHAT_REMOVED,
    KIND_THREAD_MESSAGE,
    KIND_CALL,
    KIND_CALL_SIGNAL,
    KIND_MAIL_NEW,
    KIND_MAIL,
    KIND_CALENDAR,
  ];

  static final $core.Map<$core.int, Event_Kind> _byValue =
      $pb.ProtobufEnum.initByValue(values);
  static Event_Kind? valueOf($core.int value) => _byValue[value];

  const Event_Kind._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
