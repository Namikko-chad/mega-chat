// This is a generated file - do not edit.
//
// Generated from parley.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'package:protobuf/protobuf.dart' as $pb;

import 'parley.pb.dart' as $0;

export 'parley.pb.dart';

/// Parley is served by the core, which runs inside the app's process on
/// every platform (unix socket, tcp on localhost on Windows). The Flutter
/// app and the parley CLI are clients. Every call carries "authorization: Bearer <token>".
///
/// Ids: an account is identified by account_id (never by provider type:
/// several accounts of one provider are normal). Chat and message ids are
/// the provider's own ids and are unique only inside one account, so every
/// call that names a chat or a message also names the account.
@$pb.GrpcServiceName('selfhold.parley.v1.Parley')
class ParleyClient extends $grpc.Client {
  /// The hostname for this service.
  static const $core.String defaultHost = '';

  /// OAuth scopes needed for the client.
  static const $core.List<$core.String> oauthScopes = [
    '',
  ];

  ParleyClient(super.channel, {super.options, super.interceptors});

  /// Vault lifecycle.
  $grpc.ResponseFuture<$0.State> getState(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getState, request, options: options);
  }

  $grpc.ResponseFuture<$0.State> createVault(
    $0.PasswordRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$createVault, request, options: options);
  }

  $grpc.ResponseFuture<$0.State> unlock(
    $0.PasswordRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$unlock, request, options: options);
  }

  $grpc.ResponseFuture<$0.State> lockVault(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$lockVault, request, options: options);
  }

  $grpc.ResponseFuture<$0.Empty> changePassword(
    $0.ChangePasswordRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$changePassword, request, options: options);
  }

  /// PermissionDenied unless the password opens the vault (e.g. before
  /// saving it for fingerprint unlock).
  $grpc.ResponseFuture<$0.Empty> checkPassword(
    $0.PasswordRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$checkPassword, request, options: options);
  }

  /// "Remember the password on this machine" (State.can_remember): the key
  /// derived from the password goes to the OS keychain (Secret Service,
  /// Windows Credential Manager), never the password itself.
  /// RememberPassword checks the password (PermissionDenied) and keeps it;
  /// UnlockRemembered opens the vault with it: NotFound when nothing is
  /// kept, PermissionDenied (and forgotten) when the password has changed
  /// elsewhere, Unavailable when the keychain does not answer. A password
  /// change here keeps what is remembered up to date.
  $grpc.ResponseFuture<$0.State> rememberPassword(
    $0.PasswordRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$rememberPassword, request, options: options);
  }

  $grpc.ResponseFuture<$0.State> forgetPassword(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$forgetPassword, request, options: options);
  }

  $grpc.ResponseFuture<$0.State> unlockRemembered(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$unlockRemembered, request, options: options);
  }

  /// Where the vault file lives (desktops; kept across runs). move: the open
  /// vault goes to path (FailedPrecondition while locked, AlreadyExists if a
  /// file is there), a remembered password follows it. Otherwise path is an
  /// existing vault to use from now on (FailedPrecondition while unlocked,
  /// NotFound if there is no file). A directory means "vault.parley" in it.
  $grpc.ResponseFuture<$0.State> setVaultPath(
    $0.VaultPathRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setVaultPath, request, options: options);
  }

  /// "Sync via Shard" (CONNECTOR.md): the vault file is kept in the
  /// user's Shard vault and every device gets the same one. It works
  /// only while Shard runs on this device (desktops: shardd; Android:
  /// the Shard app) and the app stays usable without it. Changes come as
  /// KIND_SHARD. EnableShard pairs this device (state PAIRING with a
  /// code the user compares in Shard) and syncs; it works while unlocked
  /// or when there is no vault yet (it then takes the synced one).
  /// CONFLICT: the synced vault differs from this one, ResolveShard
  /// picks. NEEDS_PASSWORD: the synced copy has another password,
  /// ShardPassword gives it. DisableShard keeps the local vault.
  $grpc.ResponseFuture<$0.ShardState> shard(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$shard, request, options: options);
  }

  $grpc.ResponseFuture<$0.ShardState> enableShard(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$enableShard, request, options: options);
  }

  $grpc.ResponseFuture<$0.ShardState> cancelShardPairing(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$cancelShardPairing, request, options: options);
  }

  $grpc.ResponseFuture<$0.ShardState> resolveShard(
    $0.ResolveShardRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$resolveShard, request, options: options);
  }

  $grpc.ResponseFuture<$0.ShardState> shardPassword(
    $0.PasswordRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$shardPassword, request, options: options);
  }

  $grpc.ResponseFuture<$0.ShardState> disableShard(
    $0.DisableShardRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$disableShard, request, options: options);
  }

  /// Providers and accounts.
  $grpc.ResponseFuture<$0.ProvidersReply> providers(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$providers, request, options: options);
  }

  /// Sign-in is a short dialog driven by the core: StartLogin returns the
  /// first step (fields to fill, a URL to open, ...), LoginStep answers it
  /// and returns the next one, until a step is done and carries the new
  /// account. CancelLogin drops an unfinished sign-in.
  $grpc.ResponseFuture<$0.LoginStepReply> startLogin(
    $0.StartLoginRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$startLogin, request, options: options);
  }

  $grpc.ResponseFuture<$0.LoginStepReply> loginStep(
    $0.LoginStepRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$loginStep, request, options: options);
  }

  $grpc.ResponseFuture<$0.Empty> cancelLogin(
    $0.IdRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$cancelLogin, request, options: options);
  }

  /// Signs an existing account in again (its session expired); same dialog,
  /// the result replaces the account's credentials and keeps its id.
  $grpc.ResponseFuture<$0.LoginStepReply> relogin(
    $0.IdRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$relogin, request, options: options);
  }

  $grpc.ResponseFuture<$0.AccountsReply> accounts(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$accounts, request, options: options);
  }

  $grpc.ResponseFuture<$0.Account> renameAccount(
    $0.RenameRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$renameAccount, request, options: options);
  }

  $grpc.ResponseFuture<$0.Account> setAccountEnabled(
    $0.SetEnabledRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setAccountEnabled, request, options: options);
  }

  /// Signs out and forgets the account and its cached messages.
  $grpc.ResponseFuture<$0.Empty> removeAccount(
    $0.IdRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$removeAccount, request, options: options);
  }

  /// Puts the accounts in this order; accounts not listed follow, as they were.
  $grpc.ResponseFuture<$0.AccountsReply> reorderAccounts(
    $0.ReorderAccountsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$reorderAccounts, request, options: options);
  }

  /// Chats and messages.
  /// All chats of the given accounts (all accounts when empty), newest
  /// activity first. Served from the local cache; Refresh asks providers.
  $grpc.ResponseFuture<$0.ChatsReply> chats(
    $0.ChatsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$chats, request, options: options);
  }

  $grpc.ResponseFuture<$0.Empty> refreshChats(
    $0.ChatsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$refreshChats, request, options: options);
  }

  /// Messages of one chat, newest first, before before_id (latest when
  /// empty). Missing ones are fetched from the provider and cached.
  $grpc.ResponseFuture<$0.HistoryReply> history(
    $0.HistoryRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$history, request, options: options);
  }

  $grpc.ResponseFuture<$0.Message> send(
    $0.SendRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$send, request, options: options);
  }

  $grpc.ResponseFuture<$0.ForwardReply> forward(
    $0.ForwardRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$forward, request, options: options);
  }

  $grpc.ResponseFuture<$0.Empty> markRead(
    $0.MarkReadRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$markRead, request, options: options);
  }

  /// Replies of a thread (Slack thread, Telegram channel comments), newest
  /// first like History. Not cached; live replies come as
  /// KIND_THREAD_MESSAGE.
  $grpc.ResponseFuture<$0.HistoryReply> thread(
    $0.ThreadRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$thread, request, options: options);
  }

  /// Topics of a forum chat (Chat.forum), pinned first. A topic's messages
  /// are the thread with the topic's id: Thread, SendRequest.thread_id,
  /// KIND_THREAD_MESSAGE; they are also in the chat's History. Not cached.
  $grpc.ResponseFuture<$0.TopicsReply> topics(
    $0.TopicsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$topics, request, options: options);
  }

  /// Adds or takes back the account's reaction; returns the message with
  /// its new reactions (a KIND_MESSAGE_EDIT follows).
  $grpc.ResponseFuture<$0.Message> react(
    $0.ReactRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$react, request, options: options);
  }

  /// Replaces the text of one of the account's own messages; returns the
  /// message as it is now (a KIND_MESSAGE_EDIT follows). CAPABILITY_EDIT.
  $grpc.ResponseFuture<$0.Message> edit(
    $0.EditRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$edit, request, options: options);
  }

  /// Deletes the account's own messages of one chat for everyone
  /// (CAPABILITY_DELETE) or only for this account (for_all false,
  /// CAPABILITY_DELETE_FOR_ME); KIND_MESSAGE_DELETE (KIND_THREAD_MESSAGE
  /// with deleted for thread replies) follows.
  $grpc.ResponseFuture<$0.Empty> delete(
    $0.DeleteRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$delete, request, options: options);
  }

  /// A link's preview from its page's Open Graph tags, fetched by the core
  /// (public addresses only; cached in memory). NotFound: nothing to show.
  $grpc.ResponseFuture<$0.LinkPreviewReply> linkPreview(
    $0.LinkPreviewRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$linkPreview, request, options: options);
  }

  /// People of a chat to mention, for the composer's "@" (CAPABILITY_MENTIONS).
  $grpc.ResponseFuture<$0.MembersReply> members(
    $0.MembersRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$members, request, options: options);
  }

  /// Content of a media ref: an avatar (User/Chat.avatar_url) or an
  /// attachment (Attachment.ref, .thumb). Cached in the core.
  $grpc.ResponseFuture<$0.MediaReply> media(
    $0.MediaRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$media, request, options: options);
  }

  /// What a chat is: the person of a direct chat, a group or a channel
  /// (CAPABILITY_PROFILES). Not cached.
  $grpc.ResponseFuture<$0.ChatProfile> profile(
    $0.ProfileRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$profile, request, options: options);
  }

  /// Mutes, pins, archives or leaves a chat (Account.chat_actions,
  /// CAPABILITY_MANAGE_CHATS) and returns it as it is now; a KIND_CHAT (or,
  /// after leaving, KIND_CHAT_REMOVED) follows. Leaving: groups and
  /// channels only.
  $grpc.ResponseFuture<$0.Chat> manageChat(
    $0.ManageChatRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$manageChat, request, options: options);
  }

  /// Chat folders of the accounts (CAPABILITY_CHAT_FOLDERS: Telegram).
  /// KIND_CHAT_FOLDERS tells they changed.
  $grpc.ResponseFuture<$0.ChatFoldersReply> chatFolders(
    $0.AccountsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$chatFolders, request, options: options);
  }

  /// 1:1 calls (CAPABILITY_CALLS). Media runs in the app as a standard
  /// WebRTC peer connection: the core passes its offer, answer, candidates
  /// and data-channel messages to the messenger (CallSignal) and back
  /// (KIND_CALL_SIGNAL). KIND_CALL reports state changes.
  $grpc.ResponseFuture<$0.Call> startCall(
    $0.StartCallRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$startCall, request, options: options);
  }

  $grpc.ResponseFuture<$0.Empty> acceptCall(
    $0.CallRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$acceptCall, request, options: options);
  }

  $grpc.ResponseFuture<$0.Empty> endCall(
    $0.EndCallRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$endCall, request, options: options);
  }

  $grpc.ResponseFuture<$0.Empty> callSignal(
    $0.CallSignalRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$callSignal, request, options: options);
  }

  $grpc.ResponseFuture<$0.Empty> setCallMedia(
    $0.CallMediaRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setCallMedia, request, options: options);
  }

  /// Live calls (after a reconnect), and the signals of one call after a
  /// seq (the ones a reconnect may have lost).
  $grpc.ResponseFuture<$0.CallsReply> calls(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$calls, request, options: options);
  }

  $grpc.ResponseFuture<$0.CallSignalsReply> callSignals(
    $0.CallSignalsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$callSignals, request, options: options);
  }

  /// Live events: new/edited/deleted messages, chat changes, account state,
  /// vault state. The stream ends when the client goes away.
  /// Mail (CAPABILITY_MAIL). Not cached: read from the providers, several
  /// accounts at once, merged. Accounts that failed are listed in the reply;
  /// the others still count. Attachments are media refs (Media).
  $grpc.ResponseFuture<$0.FoldersReply> folders(
    $0.AccountsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$folders, request, options: options);
  }

  /// One page of the merged mail of several accounts (all inboxes, ...),
  /// newest first.
  $grpc.ResponseFuture<$0.MailListReply> mailList(
    $0.MailListRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$mailList, request, options: options);
  }

  /// One message with its body; it is not marked read (UpdateMail READ).
  $grpc.ResponseFuture<$0.Mail> getMail(
    $0.GetMailRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getMail, request, options: options);
  }

  $grpc.ResponseFuture<$0.Empty> sendMail(
    $0.SendMailRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$sendMail, request, options: options);
  }

  $grpc.ResponseFuture<$0.Empty> updateMail(
    $0.UpdateMailRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$updateMail, request, options: options);
  }

  /// Moves messages of one account between its folders.
  $grpc.ResponseFuture<$0.Empty> moveMail(
    $0.MoveMailRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$moveMail, request, options: options);
  }

  /// Calendars (CAPABILITY_CALENDAR): all calendars of all accounts, and
  /// their events of a time range merged by start.
  $grpc.ResponseFuture<$0.CalendarsReply> calendars(
    $0.AccountsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$calendars, request, options: options);
  }

  $grpc.ResponseFuture<$0.EventsReply> events(
    $0.EventsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$events, request, options: options);
  }

  $grpc.ResponseFuture<$0.CalendarEvent> saveEvent(
    $0.CalendarEvent request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$saveEvent, request, options: options);
  }

  $grpc.ResponseFuture<$0.Empty> deleteEvent(
    $0.DeleteEventRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$deleteEvent, request, options: options);
  }

  $grpc.ResponseStream<$0.Event> watch(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(_$watch, $async.Stream.fromIterable([request]),
        options: options);
  }

  /// Recent core events (connection failures, sync errors).
  $grpc.ResponseFuture<$0.LogsReply> logs(
    $0.LogsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$logs, request, options: options);
  }

  $grpc.ResponseFuture<$0.Empty> clearLogs(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$clearLogs, request, options: options);
  }

  // method descriptors

  static final _$getState = $grpc.ClientMethod<$0.Empty, $0.State>(
      '/selfhold.parley.v1.Parley/GetState',
      ($0.Empty value) => value.writeToBuffer(),
      $0.State.fromBuffer);
  static final _$createVault = $grpc.ClientMethod<$0.PasswordRequest, $0.State>(
      '/selfhold.parley.v1.Parley/CreateVault',
      ($0.PasswordRequest value) => value.writeToBuffer(),
      $0.State.fromBuffer);
  static final _$unlock = $grpc.ClientMethod<$0.PasswordRequest, $0.State>(
      '/selfhold.parley.v1.Parley/Unlock',
      ($0.PasswordRequest value) => value.writeToBuffer(),
      $0.State.fromBuffer);
  static final _$lockVault = $grpc.ClientMethod<$0.Empty, $0.State>(
      '/selfhold.parley.v1.Parley/LockVault',
      ($0.Empty value) => value.writeToBuffer(),
      $0.State.fromBuffer);
  static final _$changePassword =
      $grpc.ClientMethod<$0.ChangePasswordRequest, $0.Empty>(
          '/selfhold.parley.v1.Parley/ChangePassword',
          ($0.ChangePasswordRequest value) => value.writeToBuffer(),
          $0.Empty.fromBuffer);
  static final _$checkPassword =
      $grpc.ClientMethod<$0.PasswordRequest, $0.Empty>(
          '/selfhold.parley.v1.Parley/CheckPassword',
          ($0.PasswordRequest value) => value.writeToBuffer(),
          $0.Empty.fromBuffer);
  static final _$rememberPassword =
      $grpc.ClientMethod<$0.PasswordRequest, $0.State>(
          '/selfhold.parley.v1.Parley/RememberPassword',
          ($0.PasswordRequest value) => value.writeToBuffer(),
          $0.State.fromBuffer);
  static final _$forgetPassword = $grpc.ClientMethod<$0.Empty, $0.State>(
      '/selfhold.parley.v1.Parley/ForgetPassword',
      ($0.Empty value) => value.writeToBuffer(),
      $0.State.fromBuffer);
  static final _$unlockRemembered = $grpc.ClientMethod<$0.Empty, $0.State>(
      '/selfhold.parley.v1.Parley/UnlockRemembered',
      ($0.Empty value) => value.writeToBuffer(),
      $0.State.fromBuffer);
  static final _$setVaultPath =
      $grpc.ClientMethod<$0.VaultPathRequest, $0.State>(
          '/selfhold.parley.v1.Parley/SetVaultPath',
          ($0.VaultPathRequest value) => value.writeToBuffer(),
          $0.State.fromBuffer);
  static final _$shard = $grpc.ClientMethod<$0.Empty, $0.ShardState>(
      '/selfhold.parley.v1.Parley/Shard',
      ($0.Empty value) => value.writeToBuffer(),
      $0.ShardState.fromBuffer);
  static final _$enableShard = $grpc.ClientMethod<$0.Empty, $0.ShardState>(
      '/selfhold.parley.v1.Parley/EnableShard',
      ($0.Empty value) => value.writeToBuffer(),
      $0.ShardState.fromBuffer);
  static final _$cancelShardPairing =
      $grpc.ClientMethod<$0.Empty, $0.ShardState>(
          '/selfhold.parley.v1.Parley/CancelShardPairing',
          ($0.Empty value) => value.writeToBuffer(),
          $0.ShardState.fromBuffer);
  static final _$resolveShard =
      $grpc.ClientMethod<$0.ResolveShardRequest, $0.ShardState>(
          '/selfhold.parley.v1.Parley/ResolveShard',
          ($0.ResolveShardRequest value) => value.writeToBuffer(),
          $0.ShardState.fromBuffer);
  static final _$shardPassword =
      $grpc.ClientMethod<$0.PasswordRequest, $0.ShardState>(
          '/selfhold.parley.v1.Parley/ShardPassword',
          ($0.PasswordRequest value) => value.writeToBuffer(),
          $0.ShardState.fromBuffer);
  static final _$disableShard =
      $grpc.ClientMethod<$0.DisableShardRequest, $0.ShardState>(
          '/selfhold.parley.v1.Parley/DisableShard',
          ($0.DisableShardRequest value) => value.writeToBuffer(),
          $0.ShardState.fromBuffer);
  static final _$providers = $grpc.ClientMethod<$0.Empty, $0.ProvidersReply>(
      '/selfhold.parley.v1.Parley/Providers',
      ($0.Empty value) => value.writeToBuffer(),
      $0.ProvidersReply.fromBuffer);
  static final _$startLogin =
      $grpc.ClientMethod<$0.StartLoginRequest, $0.LoginStepReply>(
          '/selfhold.parley.v1.Parley/StartLogin',
          ($0.StartLoginRequest value) => value.writeToBuffer(),
          $0.LoginStepReply.fromBuffer);
  static final _$loginStep =
      $grpc.ClientMethod<$0.LoginStepRequest, $0.LoginStepReply>(
          '/selfhold.parley.v1.Parley/LoginStep',
          ($0.LoginStepRequest value) => value.writeToBuffer(),
          $0.LoginStepReply.fromBuffer);
  static final _$cancelLogin = $grpc.ClientMethod<$0.IdRequest, $0.Empty>(
      '/selfhold.parley.v1.Parley/CancelLogin',
      ($0.IdRequest value) => value.writeToBuffer(),
      $0.Empty.fromBuffer);
  static final _$relogin = $grpc.ClientMethod<$0.IdRequest, $0.LoginStepReply>(
      '/selfhold.parley.v1.Parley/Relogin',
      ($0.IdRequest value) => value.writeToBuffer(),
      $0.LoginStepReply.fromBuffer);
  static final _$accounts = $grpc.ClientMethod<$0.Empty, $0.AccountsReply>(
      '/selfhold.parley.v1.Parley/Accounts',
      ($0.Empty value) => value.writeToBuffer(),
      $0.AccountsReply.fromBuffer);
  static final _$renameAccount =
      $grpc.ClientMethod<$0.RenameRequest, $0.Account>(
          '/selfhold.parley.v1.Parley/RenameAccount',
          ($0.RenameRequest value) => value.writeToBuffer(),
          $0.Account.fromBuffer);
  static final _$setAccountEnabled =
      $grpc.ClientMethod<$0.SetEnabledRequest, $0.Account>(
          '/selfhold.parley.v1.Parley/SetAccountEnabled',
          ($0.SetEnabledRequest value) => value.writeToBuffer(),
          $0.Account.fromBuffer);
  static final _$removeAccount = $grpc.ClientMethod<$0.IdRequest, $0.Empty>(
      '/selfhold.parley.v1.Parley/RemoveAccount',
      ($0.IdRequest value) => value.writeToBuffer(),
      $0.Empty.fromBuffer);
  static final _$reorderAccounts =
      $grpc.ClientMethod<$0.ReorderAccountsRequest, $0.AccountsReply>(
          '/selfhold.parley.v1.Parley/ReorderAccounts',
          ($0.ReorderAccountsRequest value) => value.writeToBuffer(),
          $0.AccountsReply.fromBuffer);
  static final _$chats = $grpc.ClientMethod<$0.ChatsRequest, $0.ChatsReply>(
      '/selfhold.parley.v1.Parley/Chats',
      ($0.ChatsRequest value) => value.writeToBuffer(),
      $0.ChatsReply.fromBuffer);
  static final _$refreshChats = $grpc.ClientMethod<$0.ChatsRequest, $0.Empty>(
      '/selfhold.parley.v1.Parley/RefreshChats',
      ($0.ChatsRequest value) => value.writeToBuffer(),
      $0.Empty.fromBuffer);
  static final _$history =
      $grpc.ClientMethod<$0.HistoryRequest, $0.HistoryReply>(
          '/selfhold.parley.v1.Parley/History',
          ($0.HistoryRequest value) => value.writeToBuffer(),
          $0.HistoryReply.fromBuffer);
  static final _$send = $grpc.ClientMethod<$0.SendRequest, $0.Message>(
      '/selfhold.parley.v1.Parley/Send',
      ($0.SendRequest value) => value.writeToBuffer(),
      $0.Message.fromBuffer);
  static final _$forward =
      $grpc.ClientMethod<$0.ForwardRequest, $0.ForwardReply>(
          '/selfhold.parley.v1.Parley/Forward',
          ($0.ForwardRequest value) => value.writeToBuffer(),
          $0.ForwardReply.fromBuffer);
  static final _$markRead = $grpc.ClientMethod<$0.MarkReadRequest, $0.Empty>(
      '/selfhold.parley.v1.Parley/MarkRead',
      ($0.MarkReadRequest value) => value.writeToBuffer(),
      $0.Empty.fromBuffer);
  static final _$thread = $grpc.ClientMethod<$0.ThreadRequest, $0.HistoryReply>(
      '/selfhold.parley.v1.Parley/Thread',
      ($0.ThreadRequest value) => value.writeToBuffer(),
      $0.HistoryReply.fromBuffer);
  static final _$topics = $grpc.ClientMethod<$0.TopicsRequest, $0.TopicsReply>(
      '/selfhold.parley.v1.Parley/Topics',
      ($0.TopicsRequest value) => value.writeToBuffer(),
      $0.TopicsReply.fromBuffer);
  static final _$react = $grpc.ClientMethod<$0.ReactRequest, $0.Message>(
      '/selfhold.parley.v1.Parley/React',
      ($0.ReactRequest value) => value.writeToBuffer(),
      $0.Message.fromBuffer);
  static final _$edit = $grpc.ClientMethod<$0.EditRequest, $0.Message>(
      '/selfhold.parley.v1.Parley/Edit',
      ($0.EditRequest value) => value.writeToBuffer(),
      $0.Message.fromBuffer);
  static final _$delete = $grpc.ClientMethod<$0.DeleteRequest, $0.Empty>(
      '/selfhold.parley.v1.Parley/Delete',
      ($0.DeleteRequest value) => value.writeToBuffer(),
      $0.Empty.fromBuffer);
  static final _$linkPreview =
      $grpc.ClientMethod<$0.LinkPreviewRequest, $0.LinkPreviewReply>(
          '/selfhold.parley.v1.Parley/LinkPreview',
          ($0.LinkPreviewRequest value) => value.writeToBuffer(),
          $0.LinkPreviewReply.fromBuffer);
  static final _$members =
      $grpc.ClientMethod<$0.MembersRequest, $0.MembersReply>(
          '/selfhold.parley.v1.Parley/Members',
          ($0.MembersRequest value) => value.writeToBuffer(),
          $0.MembersReply.fromBuffer);
  static final _$media = $grpc.ClientMethod<$0.MediaRequest, $0.MediaReply>(
      '/selfhold.parley.v1.Parley/Media',
      ($0.MediaRequest value) => value.writeToBuffer(),
      $0.MediaReply.fromBuffer);
  static final _$profile =
      $grpc.ClientMethod<$0.ProfileRequest, $0.ChatProfile>(
          '/selfhold.parley.v1.Parley/Profile',
          ($0.ProfileRequest value) => value.writeToBuffer(),
          $0.ChatProfile.fromBuffer);
  static final _$manageChat = $grpc.ClientMethod<$0.ManageChatRequest, $0.Chat>(
      '/selfhold.parley.v1.Parley/ManageChat',
      ($0.ManageChatRequest value) => value.writeToBuffer(),
      $0.Chat.fromBuffer);
  static final _$chatFolders =
      $grpc.ClientMethod<$0.AccountsRequest, $0.ChatFoldersReply>(
          '/selfhold.parley.v1.Parley/ChatFolders',
          ($0.AccountsRequest value) => value.writeToBuffer(),
          $0.ChatFoldersReply.fromBuffer);
  static final _$startCall = $grpc.ClientMethod<$0.StartCallRequest, $0.Call>(
      '/selfhold.parley.v1.Parley/StartCall',
      ($0.StartCallRequest value) => value.writeToBuffer(),
      $0.Call.fromBuffer);
  static final _$acceptCall = $grpc.ClientMethod<$0.CallRequest, $0.Empty>(
      '/selfhold.parley.v1.Parley/AcceptCall',
      ($0.CallRequest value) => value.writeToBuffer(),
      $0.Empty.fromBuffer);
  static final _$endCall = $grpc.ClientMethod<$0.EndCallRequest, $0.Empty>(
      '/selfhold.parley.v1.Parley/EndCall',
      ($0.EndCallRequest value) => value.writeToBuffer(),
      $0.Empty.fromBuffer);
  static final _$callSignal =
      $grpc.ClientMethod<$0.CallSignalRequest, $0.Empty>(
          '/selfhold.parley.v1.Parley/CallSignal',
          ($0.CallSignalRequest value) => value.writeToBuffer(),
          $0.Empty.fromBuffer);
  static final _$setCallMedia =
      $grpc.ClientMethod<$0.CallMediaRequest, $0.Empty>(
          '/selfhold.parley.v1.Parley/SetCallMedia',
          ($0.CallMediaRequest value) => value.writeToBuffer(),
          $0.Empty.fromBuffer);
  static final _$calls = $grpc.ClientMethod<$0.Empty, $0.CallsReply>(
      '/selfhold.parley.v1.Parley/Calls',
      ($0.Empty value) => value.writeToBuffer(),
      $0.CallsReply.fromBuffer);
  static final _$callSignals =
      $grpc.ClientMethod<$0.CallSignalsRequest, $0.CallSignalsReply>(
          '/selfhold.parley.v1.Parley/CallSignals',
          ($0.CallSignalsRequest value) => value.writeToBuffer(),
          $0.CallSignalsReply.fromBuffer);
  static final _$folders =
      $grpc.ClientMethod<$0.AccountsRequest, $0.FoldersReply>(
          '/selfhold.parley.v1.Parley/Folders',
          ($0.AccountsRequest value) => value.writeToBuffer(),
          $0.FoldersReply.fromBuffer);
  static final _$mailList =
      $grpc.ClientMethod<$0.MailListRequest, $0.MailListReply>(
          '/selfhold.parley.v1.Parley/MailList',
          ($0.MailListRequest value) => value.writeToBuffer(),
          $0.MailListReply.fromBuffer);
  static final _$getMail = $grpc.ClientMethod<$0.GetMailRequest, $0.Mail>(
      '/selfhold.parley.v1.Parley/GetMail',
      ($0.GetMailRequest value) => value.writeToBuffer(),
      $0.Mail.fromBuffer);
  static final _$sendMail = $grpc.ClientMethod<$0.SendMailRequest, $0.Empty>(
      '/selfhold.parley.v1.Parley/SendMail',
      ($0.SendMailRequest value) => value.writeToBuffer(),
      $0.Empty.fromBuffer);
  static final _$updateMail =
      $grpc.ClientMethod<$0.UpdateMailRequest, $0.Empty>(
          '/selfhold.parley.v1.Parley/UpdateMail',
          ($0.UpdateMailRequest value) => value.writeToBuffer(),
          $0.Empty.fromBuffer);
  static final _$moveMail = $grpc.ClientMethod<$0.MoveMailRequest, $0.Empty>(
      '/selfhold.parley.v1.Parley/MoveMail',
      ($0.MoveMailRequest value) => value.writeToBuffer(),
      $0.Empty.fromBuffer);
  static final _$calendars =
      $grpc.ClientMethod<$0.AccountsRequest, $0.CalendarsReply>(
          '/selfhold.parley.v1.Parley/Calendars',
          ($0.AccountsRequest value) => value.writeToBuffer(),
          $0.CalendarsReply.fromBuffer);
  static final _$events = $grpc.ClientMethod<$0.EventsRequest, $0.EventsReply>(
      '/selfhold.parley.v1.Parley/Events',
      ($0.EventsRequest value) => value.writeToBuffer(),
      $0.EventsReply.fromBuffer);
  static final _$saveEvent =
      $grpc.ClientMethod<$0.CalendarEvent, $0.CalendarEvent>(
          '/selfhold.parley.v1.Parley/SaveEvent',
          ($0.CalendarEvent value) => value.writeToBuffer(),
          $0.CalendarEvent.fromBuffer);
  static final _$deleteEvent =
      $grpc.ClientMethod<$0.DeleteEventRequest, $0.Empty>(
          '/selfhold.parley.v1.Parley/DeleteEvent',
          ($0.DeleteEventRequest value) => value.writeToBuffer(),
          $0.Empty.fromBuffer);
  static final _$watch = $grpc.ClientMethod<$0.Empty, $0.Event>(
      '/selfhold.parley.v1.Parley/Watch',
      ($0.Empty value) => value.writeToBuffer(),
      $0.Event.fromBuffer);
  static final _$logs = $grpc.ClientMethod<$0.LogsRequest, $0.LogsReply>(
      '/selfhold.parley.v1.Parley/Logs',
      ($0.LogsRequest value) => value.writeToBuffer(),
      $0.LogsReply.fromBuffer);
  static final _$clearLogs = $grpc.ClientMethod<$0.Empty, $0.Empty>(
      '/selfhold.parley.v1.Parley/ClearLogs',
      ($0.Empty value) => value.writeToBuffer(),
      $0.Empty.fromBuffer);
}

@$pb.GrpcServiceName('selfhold.parley.v1.Parley')
abstract class ParleyServiceBase extends $grpc.Service {
  $core.String get $name => 'selfhold.parley.v1.Parley';

  ParleyServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.State>(
        'GetState',
        getState_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.State value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.PasswordRequest, $0.State>(
        'CreateVault',
        createVault_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.PasswordRequest.fromBuffer(value),
        ($0.State value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.PasswordRequest, $0.State>(
        'Unlock',
        unlock_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.PasswordRequest.fromBuffer(value),
        ($0.State value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.State>(
        'LockVault',
        lockVault_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.State value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ChangePasswordRequest, $0.Empty>(
        'ChangePassword',
        changePassword_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.ChangePasswordRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.PasswordRequest, $0.Empty>(
        'CheckPassword',
        checkPassword_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.PasswordRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.PasswordRequest, $0.State>(
        'RememberPassword',
        rememberPassword_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.PasswordRequest.fromBuffer(value),
        ($0.State value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.State>(
        'ForgetPassword',
        forgetPassword_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.State value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.State>(
        'UnlockRemembered',
        unlockRemembered_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.State value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.VaultPathRequest, $0.State>(
        'SetVaultPath',
        setVaultPath_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.VaultPathRequest.fromBuffer(value),
        ($0.State value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.ShardState>(
        'Shard',
        shard_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.ShardState value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.ShardState>(
        'EnableShard',
        enableShard_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.ShardState value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.ShardState>(
        'CancelShardPairing',
        cancelShardPairing_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.ShardState value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ResolveShardRequest, $0.ShardState>(
        'ResolveShard',
        resolveShard_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.ResolveShardRequest.fromBuffer(value),
        ($0.ShardState value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.PasswordRequest, $0.ShardState>(
        'ShardPassword',
        shardPassword_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.PasswordRequest.fromBuffer(value),
        ($0.ShardState value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.DisableShardRequest, $0.ShardState>(
        'DisableShard',
        disableShard_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.DisableShardRequest.fromBuffer(value),
        ($0.ShardState value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.ProvidersReply>(
        'Providers',
        providers_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.ProvidersReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.StartLoginRequest, $0.LoginStepReply>(
        'StartLogin',
        startLogin_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.StartLoginRequest.fromBuffer(value),
        ($0.LoginStepReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.LoginStepRequest, $0.LoginStepReply>(
        'LoginStep',
        loginStep_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.LoginStepRequest.fromBuffer(value),
        ($0.LoginStepReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.IdRequest, $0.Empty>(
        'CancelLogin',
        cancelLogin_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.IdRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.IdRequest, $0.LoginStepReply>(
        'Relogin',
        relogin_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.IdRequest.fromBuffer(value),
        ($0.LoginStepReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.AccountsReply>(
        'Accounts',
        accounts_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.AccountsReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RenameRequest, $0.Account>(
        'RenameAccount',
        renameAccount_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.RenameRequest.fromBuffer(value),
        ($0.Account value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.SetEnabledRequest, $0.Account>(
        'SetAccountEnabled',
        setAccountEnabled_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.SetEnabledRequest.fromBuffer(value),
        ($0.Account value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.IdRequest, $0.Empty>(
        'RemoveAccount',
        removeAccount_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.IdRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ReorderAccountsRequest, $0.AccountsReply>(
        'ReorderAccounts',
        reorderAccounts_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.ReorderAccountsRequest.fromBuffer(value),
        ($0.AccountsReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ChatsRequest, $0.ChatsReply>(
        'Chats',
        chats_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.ChatsRequest.fromBuffer(value),
        ($0.ChatsReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ChatsRequest, $0.Empty>(
        'RefreshChats',
        refreshChats_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.ChatsRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.HistoryRequest, $0.HistoryReply>(
        'History',
        history_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.HistoryRequest.fromBuffer(value),
        ($0.HistoryReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.SendRequest, $0.Message>(
        'Send',
        send_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.SendRequest.fromBuffer(value),
        ($0.Message value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ForwardRequest, $0.ForwardReply>(
        'Forward',
        forward_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.ForwardRequest.fromBuffer(value),
        ($0.ForwardReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.MarkReadRequest, $0.Empty>(
        'MarkRead',
        markRead_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.MarkReadRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ThreadRequest, $0.HistoryReply>(
        'Thread',
        thread_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.ThreadRequest.fromBuffer(value),
        ($0.HistoryReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.TopicsRequest, $0.TopicsReply>(
        'Topics',
        topics_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.TopicsRequest.fromBuffer(value),
        ($0.TopicsReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ReactRequest, $0.Message>(
        'React',
        react_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.ReactRequest.fromBuffer(value),
        ($0.Message value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.EditRequest, $0.Message>(
        'Edit',
        edit_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.EditRequest.fromBuffer(value),
        ($0.Message value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.DeleteRequest, $0.Empty>(
        'Delete',
        delete_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.DeleteRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.LinkPreviewRequest, $0.LinkPreviewReply>(
        'LinkPreview',
        linkPreview_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.LinkPreviewRequest.fromBuffer(value),
        ($0.LinkPreviewReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.MembersRequest, $0.MembersReply>(
        'Members',
        members_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.MembersRequest.fromBuffer(value),
        ($0.MembersReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.MediaRequest, $0.MediaReply>(
        'Media',
        media_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.MediaRequest.fromBuffer(value),
        ($0.MediaReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ProfileRequest, $0.ChatProfile>(
        'Profile',
        profile_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.ProfileRequest.fromBuffer(value),
        ($0.ChatProfile value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ManageChatRequest, $0.Chat>(
        'ManageChat',
        manageChat_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.ManageChatRequest.fromBuffer(value),
        ($0.Chat value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.AccountsRequest, $0.ChatFoldersReply>(
        'ChatFolders',
        chatFolders_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.AccountsRequest.fromBuffer(value),
        ($0.ChatFoldersReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.StartCallRequest, $0.Call>(
        'StartCall',
        startCall_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.StartCallRequest.fromBuffer(value),
        ($0.Call value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.CallRequest, $0.Empty>(
        'AcceptCall',
        acceptCall_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.CallRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.EndCallRequest, $0.Empty>(
        'EndCall',
        endCall_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.EndCallRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.CallSignalRequest, $0.Empty>(
        'CallSignal',
        callSignal_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.CallSignalRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.CallMediaRequest, $0.Empty>(
        'SetCallMedia',
        setCallMedia_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.CallMediaRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.CallsReply>(
        'Calls',
        calls_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.CallsReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.CallSignalsRequest, $0.CallSignalsReply>(
        'CallSignals',
        callSignals_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.CallSignalsRequest.fromBuffer(value),
        ($0.CallSignalsReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.AccountsRequest, $0.FoldersReply>(
        'Folders',
        folders_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.AccountsRequest.fromBuffer(value),
        ($0.FoldersReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.MailListRequest, $0.MailListReply>(
        'MailList',
        mailList_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.MailListRequest.fromBuffer(value),
        ($0.MailListReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.GetMailRequest, $0.Mail>(
        'GetMail',
        getMail_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.GetMailRequest.fromBuffer(value),
        ($0.Mail value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.SendMailRequest, $0.Empty>(
        'SendMail',
        sendMail_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.SendMailRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.UpdateMailRequest, $0.Empty>(
        'UpdateMail',
        updateMail_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.UpdateMailRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.MoveMailRequest, $0.Empty>(
        'MoveMail',
        moveMail_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.MoveMailRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.AccountsRequest, $0.CalendarsReply>(
        'Calendars',
        calendars_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.AccountsRequest.fromBuffer(value),
        ($0.CalendarsReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.EventsRequest, $0.EventsReply>(
        'Events',
        events_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.EventsRequest.fromBuffer(value),
        ($0.EventsReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.CalendarEvent, $0.CalendarEvent>(
        'SaveEvent',
        saveEvent_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.CalendarEvent.fromBuffer(value),
        ($0.CalendarEvent value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.DeleteEventRequest, $0.Empty>(
        'DeleteEvent',
        deleteEvent_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.DeleteEventRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.Event>(
        'Watch',
        watch_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.Event value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.LogsRequest, $0.LogsReply>(
        'Logs',
        logs_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.LogsRequest.fromBuffer(value),
        ($0.LogsReply value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.Empty>(
        'ClearLogs',
        clearLogs_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
  }

  $async.Future<$0.State> getState_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return getState($call, await $request);
  }

  $async.Future<$0.State> getState($grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.State> createVault_Pre($grpc.ServiceCall $call,
      $async.Future<$0.PasswordRequest> $request) async {
    return createVault($call, await $request);
  }

  $async.Future<$0.State> createVault(
      $grpc.ServiceCall call, $0.PasswordRequest request);

  $async.Future<$0.State> unlock_Pre($grpc.ServiceCall $call,
      $async.Future<$0.PasswordRequest> $request) async {
    return unlock($call, await $request);
  }

  $async.Future<$0.State> unlock(
      $grpc.ServiceCall call, $0.PasswordRequest request);

  $async.Future<$0.State> lockVault_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return lockVault($call, await $request);
  }

  $async.Future<$0.State> lockVault($grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.Empty> changePassword_Pre($grpc.ServiceCall $call,
      $async.Future<$0.ChangePasswordRequest> $request) async {
    return changePassword($call, await $request);
  }

  $async.Future<$0.Empty> changePassword(
      $grpc.ServiceCall call, $0.ChangePasswordRequest request);

  $async.Future<$0.Empty> checkPassword_Pre($grpc.ServiceCall $call,
      $async.Future<$0.PasswordRequest> $request) async {
    return checkPassword($call, await $request);
  }

  $async.Future<$0.Empty> checkPassword(
      $grpc.ServiceCall call, $0.PasswordRequest request);

  $async.Future<$0.State> rememberPassword_Pre($grpc.ServiceCall $call,
      $async.Future<$0.PasswordRequest> $request) async {
    return rememberPassword($call, await $request);
  }

  $async.Future<$0.State> rememberPassword(
      $grpc.ServiceCall call, $0.PasswordRequest request);

  $async.Future<$0.State> forgetPassword_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return forgetPassword($call, await $request);
  }

  $async.Future<$0.State> forgetPassword(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.State> unlockRemembered_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return unlockRemembered($call, await $request);
  }

  $async.Future<$0.State> unlockRemembered(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.State> setVaultPath_Pre($grpc.ServiceCall $call,
      $async.Future<$0.VaultPathRequest> $request) async {
    return setVaultPath($call, await $request);
  }

  $async.Future<$0.State> setVaultPath(
      $grpc.ServiceCall call, $0.VaultPathRequest request);

  $async.Future<$0.ShardState> shard_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return shard($call, await $request);
  }

  $async.Future<$0.ShardState> shard($grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.ShardState> enableShard_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return enableShard($call, await $request);
  }

  $async.Future<$0.ShardState> enableShard(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.ShardState> cancelShardPairing_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return cancelShardPairing($call, await $request);
  }

  $async.Future<$0.ShardState> cancelShardPairing(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.ShardState> resolveShard_Pre($grpc.ServiceCall $call,
      $async.Future<$0.ResolveShardRequest> $request) async {
    return resolveShard($call, await $request);
  }

  $async.Future<$0.ShardState> resolveShard(
      $grpc.ServiceCall call, $0.ResolveShardRequest request);

  $async.Future<$0.ShardState> shardPassword_Pre($grpc.ServiceCall $call,
      $async.Future<$0.PasswordRequest> $request) async {
    return shardPassword($call, await $request);
  }

  $async.Future<$0.ShardState> shardPassword(
      $grpc.ServiceCall call, $0.PasswordRequest request);

  $async.Future<$0.ShardState> disableShard_Pre($grpc.ServiceCall $call,
      $async.Future<$0.DisableShardRequest> $request) async {
    return disableShard($call, await $request);
  }

  $async.Future<$0.ShardState> disableShard(
      $grpc.ServiceCall call, $0.DisableShardRequest request);

  $async.Future<$0.ProvidersReply> providers_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return providers($call, await $request);
  }

  $async.Future<$0.ProvidersReply> providers(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.LoginStepReply> startLogin_Pre($grpc.ServiceCall $call,
      $async.Future<$0.StartLoginRequest> $request) async {
    return startLogin($call, await $request);
  }

  $async.Future<$0.LoginStepReply> startLogin(
      $grpc.ServiceCall call, $0.StartLoginRequest request);

  $async.Future<$0.LoginStepReply> loginStep_Pre($grpc.ServiceCall $call,
      $async.Future<$0.LoginStepRequest> $request) async {
    return loginStep($call, await $request);
  }

  $async.Future<$0.LoginStepReply> loginStep(
      $grpc.ServiceCall call, $0.LoginStepRequest request);

  $async.Future<$0.Empty> cancelLogin_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.IdRequest> $request) async {
    return cancelLogin($call, await $request);
  }

  $async.Future<$0.Empty> cancelLogin(
      $grpc.ServiceCall call, $0.IdRequest request);

  $async.Future<$0.LoginStepReply> relogin_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.IdRequest> $request) async {
    return relogin($call, await $request);
  }

  $async.Future<$0.LoginStepReply> relogin(
      $grpc.ServiceCall call, $0.IdRequest request);

  $async.Future<$0.AccountsReply> accounts_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return accounts($call, await $request);
  }

  $async.Future<$0.AccountsReply> accounts(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.Account> renameAccount_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.RenameRequest> $request) async {
    return renameAccount($call, await $request);
  }

  $async.Future<$0.Account> renameAccount(
      $grpc.ServiceCall call, $0.RenameRequest request);

  $async.Future<$0.Account> setAccountEnabled_Pre($grpc.ServiceCall $call,
      $async.Future<$0.SetEnabledRequest> $request) async {
    return setAccountEnabled($call, await $request);
  }

  $async.Future<$0.Account> setAccountEnabled(
      $grpc.ServiceCall call, $0.SetEnabledRequest request);

  $async.Future<$0.Empty> removeAccount_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.IdRequest> $request) async {
    return removeAccount($call, await $request);
  }

  $async.Future<$0.Empty> removeAccount(
      $grpc.ServiceCall call, $0.IdRequest request);

  $async.Future<$0.AccountsReply> reorderAccounts_Pre($grpc.ServiceCall $call,
      $async.Future<$0.ReorderAccountsRequest> $request) async {
    return reorderAccounts($call, await $request);
  }

  $async.Future<$0.AccountsReply> reorderAccounts(
      $grpc.ServiceCall call, $0.ReorderAccountsRequest request);

  $async.Future<$0.ChatsReply> chats_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.ChatsRequest> $request) async {
    return chats($call, await $request);
  }

  $async.Future<$0.ChatsReply> chats(
      $grpc.ServiceCall call, $0.ChatsRequest request);

  $async.Future<$0.Empty> refreshChats_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.ChatsRequest> $request) async {
    return refreshChats($call, await $request);
  }

  $async.Future<$0.Empty> refreshChats(
      $grpc.ServiceCall call, $0.ChatsRequest request);

  $async.Future<$0.HistoryReply> history_Pre($grpc.ServiceCall $call,
      $async.Future<$0.HistoryRequest> $request) async {
    return history($call, await $request);
  }

  $async.Future<$0.HistoryReply> history(
      $grpc.ServiceCall call, $0.HistoryRequest request);

  $async.Future<$0.Message> send_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.SendRequest> $request) async {
    return send($call, await $request);
  }

  $async.Future<$0.Message> send(
      $grpc.ServiceCall call, $0.SendRequest request);

  $async.Future<$0.ForwardReply> forward_Pre($grpc.ServiceCall $call,
      $async.Future<$0.ForwardRequest> $request) async {
    return forward($call, await $request);
  }

  $async.Future<$0.ForwardReply> forward(
      $grpc.ServiceCall call, $0.ForwardRequest request);

  $async.Future<$0.Empty> markRead_Pre($grpc.ServiceCall $call,
      $async.Future<$0.MarkReadRequest> $request) async {
    return markRead($call, await $request);
  }

  $async.Future<$0.Empty> markRead(
      $grpc.ServiceCall call, $0.MarkReadRequest request);

  $async.Future<$0.HistoryReply> thread_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.ThreadRequest> $request) async {
    return thread($call, await $request);
  }

  $async.Future<$0.HistoryReply> thread(
      $grpc.ServiceCall call, $0.ThreadRequest request);

  $async.Future<$0.TopicsReply> topics_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.TopicsRequest> $request) async {
    return topics($call, await $request);
  }

  $async.Future<$0.TopicsReply> topics(
      $grpc.ServiceCall call, $0.TopicsRequest request);

  $async.Future<$0.Message> react_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.ReactRequest> $request) async {
    return react($call, await $request);
  }

  $async.Future<$0.Message> react(
      $grpc.ServiceCall call, $0.ReactRequest request);

  $async.Future<$0.Message> edit_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.EditRequest> $request) async {
    return edit($call, await $request);
  }

  $async.Future<$0.Message> edit(
      $grpc.ServiceCall call, $0.EditRequest request);

  $async.Future<$0.Empty> delete_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.DeleteRequest> $request) async {
    return delete($call, await $request);
  }

  $async.Future<$0.Empty> delete(
      $grpc.ServiceCall call, $0.DeleteRequest request);

  $async.Future<$0.LinkPreviewReply> linkPreview_Pre($grpc.ServiceCall $call,
      $async.Future<$0.LinkPreviewRequest> $request) async {
    return linkPreview($call, await $request);
  }

  $async.Future<$0.LinkPreviewReply> linkPreview(
      $grpc.ServiceCall call, $0.LinkPreviewRequest request);

  $async.Future<$0.MembersReply> members_Pre($grpc.ServiceCall $call,
      $async.Future<$0.MembersRequest> $request) async {
    return members($call, await $request);
  }

  $async.Future<$0.MembersReply> members(
      $grpc.ServiceCall call, $0.MembersRequest request);

  $async.Future<$0.MediaReply> media_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.MediaRequest> $request) async {
    return media($call, await $request);
  }

  $async.Future<$0.MediaReply> media(
      $grpc.ServiceCall call, $0.MediaRequest request);

  $async.Future<$0.ChatProfile> profile_Pre($grpc.ServiceCall $call,
      $async.Future<$0.ProfileRequest> $request) async {
    return profile($call, await $request);
  }

  $async.Future<$0.ChatProfile> profile(
      $grpc.ServiceCall call, $0.ProfileRequest request);

  $async.Future<$0.Chat> manageChat_Pre($grpc.ServiceCall $call,
      $async.Future<$0.ManageChatRequest> $request) async {
    return manageChat($call, await $request);
  }

  $async.Future<$0.Chat> manageChat(
      $grpc.ServiceCall call, $0.ManageChatRequest request);

  $async.Future<$0.ChatFoldersReply> chatFolders_Pre($grpc.ServiceCall $call,
      $async.Future<$0.AccountsRequest> $request) async {
    return chatFolders($call, await $request);
  }

  $async.Future<$0.ChatFoldersReply> chatFolders(
      $grpc.ServiceCall call, $0.AccountsRequest request);

  $async.Future<$0.Call> startCall_Pre($grpc.ServiceCall $call,
      $async.Future<$0.StartCallRequest> $request) async {
    return startCall($call, await $request);
  }

  $async.Future<$0.Call> startCall(
      $grpc.ServiceCall call, $0.StartCallRequest request);

  $async.Future<$0.Empty> acceptCall_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.CallRequest> $request) async {
    return acceptCall($call, await $request);
  }

  $async.Future<$0.Empty> acceptCall(
      $grpc.ServiceCall call, $0.CallRequest request);

  $async.Future<$0.Empty> endCall_Pre($grpc.ServiceCall $call,
      $async.Future<$0.EndCallRequest> $request) async {
    return endCall($call, await $request);
  }

  $async.Future<$0.Empty> endCall(
      $grpc.ServiceCall call, $0.EndCallRequest request);

  $async.Future<$0.Empty> callSignal_Pre($grpc.ServiceCall $call,
      $async.Future<$0.CallSignalRequest> $request) async {
    return callSignal($call, await $request);
  }

  $async.Future<$0.Empty> callSignal(
      $grpc.ServiceCall call, $0.CallSignalRequest request);

  $async.Future<$0.Empty> setCallMedia_Pre($grpc.ServiceCall $call,
      $async.Future<$0.CallMediaRequest> $request) async {
    return setCallMedia($call, await $request);
  }

  $async.Future<$0.Empty> setCallMedia(
      $grpc.ServiceCall call, $0.CallMediaRequest request);

  $async.Future<$0.CallsReply> calls_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return calls($call, await $request);
  }

  $async.Future<$0.CallsReply> calls($grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.CallSignalsReply> callSignals_Pre($grpc.ServiceCall $call,
      $async.Future<$0.CallSignalsRequest> $request) async {
    return callSignals($call, await $request);
  }

  $async.Future<$0.CallSignalsReply> callSignals(
      $grpc.ServiceCall call, $0.CallSignalsRequest request);

  $async.Future<$0.FoldersReply> folders_Pre($grpc.ServiceCall $call,
      $async.Future<$0.AccountsRequest> $request) async {
    return folders($call, await $request);
  }

  $async.Future<$0.FoldersReply> folders(
      $grpc.ServiceCall call, $0.AccountsRequest request);

  $async.Future<$0.MailListReply> mailList_Pre($grpc.ServiceCall $call,
      $async.Future<$0.MailListRequest> $request) async {
    return mailList($call, await $request);
  }

  $async.Future<$0.MailListReply> mailList(
      $grpc.ServiceCall call, $0.MailListRequest request);

  $async.Future<$0.Mail> getMail_Pre($grpc.ServiceCall $call,
      $async.Future<$0.GetMailRequest> $request) async {
    return getMail($call, await $request);
  }

  $async.Future<$0.Mail> getMail(
      $grpc.ServiceCall call, $0.GetMailRequest request);

  $async.Future<$0.Empty> sendMail_Pre($grpc.ServiceCall $call,
      $async.Future<$0.SendMailRequest> $request) async {
    return sendMail($call, await $request);
  }

  $async.Future<$0.Empty> sendMail(
      $grpc.ServiceCall call, $0.SendMailRequest request);

  $async.Future<$0.Empty> updateMail_Pre($grpc.ServiceCall $call,
      $async.Future<$0.UpdateMailRequest> $request) async {
    return updateMail($call, await $request);
  }

  $async.Future<$0.Empty> updateMail(
      $grpc.ServiceCall call, $0.UpdateMailRequest request);

  $async.Future<$0.Empty> moveMail_Pre($grpc.ServiceCall $call,
      $async.Future<$0.MoveMailRequest> $request) async {
    return moveMail($call, await $request);
  }

  $async.Future<$0.Empty> moveMail(
      $grpc.ServiceCall call, $0.MoveMailRequest request);

  $async.Future<$0.CalendarsReply> calendars_Pre($grpc.ServiceCall $call,
      $async.Future<$0.AccountsRequest> $request) async {
    return calendars($call, await $request);
  }

  $async.Future<$0.CalendarsReply> calendars(
      $grpc.ServiceCall call, $0.AccountsRequest request);

  $async.Future<$0.EventsReply> events_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.EventsRequest> $request) async {
    return events($call, await $request);
  }

  $async.Future<$0.EventsReply> events(
      $grpc.ServiceCall call, $0.EventsRequest request);

  $async.Future<$0.CalendarEvent> saveEvent_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.CalendarEvent> $request) async {
    return saveEvent($call, await $request);
  }

  $async.Future<$0.CalendarEvent> saveEvent(
      $grpc.ServiceCall call, $0.CalendarEvent request);

  $async.Future<$0.Empty> deleteEvent_Pre($grpc.ServiceCall $call,
      $async.Future<$0.DeleteEventRequest> $request) async {
    return deleteEvent($call, await $request);
  }

  $async.Future<$0.Empty> deleteEvent(
      $grpc.ServiceCall call, $0.DeleteEventRequest request);

  $async.Stream<$0.Event> watch_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async* {
    yield* watch($call, await $request);
  }

  $async.Stream<$0.Event> watch($grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.LogsReply> logs_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.LogsRequest> $request) async {
    return logs($call, await $request);
  }

  $async.Future<$0.LogsReply> logs(
      $grpc.ServiceCall call, $0.LogsRequest request);

  $async.Future<$0.Empty> clearLogs_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return clearLogs($call, await $request);
  }

  $async.Future<$0.Empty> clearLogs($grpc.ServiceCall call, $0.Empty request);
}
