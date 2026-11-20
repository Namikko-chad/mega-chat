import 'dart:async';

import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/settings/app_settings.dart';
import '../../data/media_store.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../chats/player.dart';
import 'core_cubit.dart';

/// Chat ids are unique only inside an account, so a chat is named by both.
typedef ChatKey = ({String account, String chat});

ChatKey chatKey(pb.Chat c) => (account: c.accountId, chat: c.id);

ChatKey messageChat(pb.Message m) => (account: m.accountId, chat: m.chatId);

/// A thread open next to its chat: the chat and the thread's id.
typedef ThreadKey = ({String account, String chat, String thread});

/// Channels and groups, or direct messages: the sections of the list.
bool isDirect(pb.Chat c) => c.kind == pb.ChatKind.CHAT_KIND_DIRECT;

class MessengerState extends Equatable {
  MessengerState({
    this.loading = false,
    this.error,
    this.accounts = const [],
    this.chats = const {},
    this.providers = const [],
    this.typing = const {},
    this.filter = '',
    this.unreadOnly = false,
    this.open,
    this.thread,
    this.threadRoot,
    this.rev = 0,
  });

  final bool loading;
  final String? error;
  final List<pb.Account> accounts;
  final Map<ChatKey, pb.Chat> chats;
  final List<pb.ProviderInfo> providers;

  /// Who is typing in a chat right now (KIND_TYPING, expires by itself).
  final Map<ChatKey, String> typing;

  /// The account the chat list shows; empty: all of them.
  final String filter;

  /// Only chats with unread messages are listed (and the open one).
  final bool unreadOnly;

  /// The chat open on the right (desktop) or pushed (phone).
  final ChatKey? open;

  /// The thread open next to the chat (desktop) or pushed (phone).
  final ThreadKey? thread;

  /// The message the open thread hangs under, as it was when opened.
  final pb.Message? threadRoot;

  /// Bumps on every change: the collections are replaced, never mutated.
  final int rev;

  /// Newest activity first.
  late final List<pb.Chat> sorted = chats.values.toList()..sort((a, b) => b.updatedMs.compareTo(a.updatedMs));

  late final List<pb.Chat> shown = sorted
      .where((c) => (filter.isEmpty || c.accountId == filter) && (!unreadOnly || c.unread > 0 || chatKey(c) == open))
      .toList();

  int get unread => accounts.fold(0, (n, a) => n + (a.enabled ? a.unread : 0));

  pb.Account? account(String id) => accounts.where((a) => a.id == id).firstOrNull;

  pb.ProviderInfo? provider(String type) => providers.where((p) => p.type == type).firstOrNull;

  MessengerState copyWith({
    bool? loading,
    String? error,
    List<pb.Account>? accounts,
    Map<ChatKey, pb.Chat>? chats,
    List<pb.ProviderInfo>? providers,
    Map<ChatKey, String>? typing,
    String? filter,
    bool? unreadOnly,
    ChatKey? open,
    ThreadKey? thread,
    pb.Message? threadRoot,
    bool clearError = false,
    bool close = false,
    bool closeThread = false,
  }) => MessengerState(
    loading: loading ?? this.loading,
    error: clearError ? null : error ?? this.error,
    accounts: accounts ?? this.accounts,
    chats: chats ?? this.chats,
    providers: providers ?? this.providers,
    typing: typing ?? this.typing,
    filter: filter ?? this.filter,
    unreadOnly: unreadOnly ?? this.unreadOnly,
    open: close ? null : open ?? this.open,
    // A thread belongs to its chat: another chat closes it.
    thread: close || closeThread || (open != null && open != this.open) ? null : thread ?? this.thread,
    threadRoot: close || closeThread || (open != null && open != this.open) ? null : threadRoot ?? this.threadRoot,
    rev: rev + 1,
  );

  @override
  List<Object?> get props => [loading, error, rev];
}

/// Accounts and the unified chat list, kept live from the core's events.
/// Loaded on unlock and after every reconnect (events may have been
/// missed), dropped on lock.
class MessengerCubit extends Cubit<MessengerState> {
  MessengerCubit(this._core) : super(MessengerState(unreadOnly: AppSettings.instance.unreadOnly)) {
    _coreSub = _core.stream.listen(_onCore);
    _eventSub = _core.events.listen(_onEvent);
    _onCore(_core.state);
  }

  final CoreCubit _core;

  /// Avatars, previews and files, kept while the vault is open.
  late final MediaStore media = MediaStore(() => _core.api);
  late final StreamSubscription<CoreState> _coreSub;
  late final StreamSubscription<pb.Event> _eventSub;
  int? _session; // the core session the data was loaded in
  final _typingTimers = <ChatKey, Timer>{};
  Timer? _reloadChats;

  pb.MegaChatClient get api => _core.api;

  void _onCore(CoreState s) {
    if (!s.unlocked) {
      if (_session != null) {
        unawaited(AudioPlayback.instance.stop());
        unawaited(media.clear());
      }
      _session = null;
      if (state.accounts.isNotEmpty || state.chats.isNotEmpty) emit(MessengerState(unreadOnly: state.unreadOnly));
      return;
    }
    if (s.conn != Conn.ready || s.session == _session) return;
    _session = s.session;
    unawaited(load());
  }

  Future<void> load() async {
    emit(state.copyWith(loading: true, clearError: true));
    try {
      final r = await Future.wait([api.accounts(pb.Empty()), api.chats(pb.ChatsRequest()), api.providers(pb.Empty())]);
      if (isClosed) return;
      final chats = (r[1] as pb.ChatsReply).chats;
      emit(
        state.copyWith(
          loading: false,
          accounts: (r[0] as pb.AccountsReply).accounts,
          chats: {for (final c in chats) chatKey(c): c},
          providers: (r[2] as pb.ProvidersReply).providers,
        ),
      );
    } catch (e) {
      if (!isClosed) emit(state.copyWith(loading: false, error: errorText(e)));
    }
  }

  Future<void> _loadChats() async {
    try {
      final r = await api.chats(pb.ChatsRequest());
      if (!isClosed) emit(state.copyWith(chats: {for (final c in r.chats) chatKey(c): c}));
    } catch (_) {}
  }

  /// Asks the providers for fresh chat lists; the result comes back as
  /// KIND_CHAT events, the reload covers a core that sends none.
  Future<void> refresh() async {
    await api.refreshChats(pb.ChatsRequest(accountIds: [if (state.filter.isNotEmpty) state.filter]));
    await _loadChats();
  }

  void setFilter(String accountId) => emit(state.copyWith(filter: accountId));

  void setUnreadOnly(bool v) {
    emit(state.copyWith(unreadOnly: v));
    unawaited(AppSettings.instance.setUnreadOnly(v));
  }

  void openChat(ChatKey? key) => emit(key == null ? state.copyWith(close: true) : state.copyWith(open: key));

  void openThread(ThreadKey? key, [pb.Message? root]) =>
      emit(key == null ? state.copyWith(closeThread: true) : state.copyWith(open: (account: key.account, chat: key.chat), thread: key, threadRoot: root));

  /// The open chat was read: clear its badge now, the core confirms later.
  /// A chat was read here: [unread] messages after the read one are left.
  void markedRead(ChatKey key, {int unread = 0}) {
    final c = state.chats[key];
    if (c == null || c.unread <= unread) return;
    emit(state.copyWith(chats: {...state.chats, key: c.deepCopy()..unread = unread}));
  }

  void upsertAccount(pb.Account a) {
    final list = [...state.accounts];
    final i = list.indexWhere((x) => x.id == a.id);
    i < 0 ? list.add(a) : list[i] = a;
    emit(state.copyWith(accounts: list));
  }

  void _removeAccount(String id) {
    emit(
      state.copyWith(
        accounts: state.accounts.where((a) => a.id != id).toList(),
        chats: Map.of(state.chats)..removeWhere((k, _) => k.account == id),
        filter: state.filter == id ? '' : null,
        close: state.open?.account == id,
      ),
    );
  }

  Future<void> rename(String id, String name) async => upsertAccount(await api.renameAccount(pb.RenameRequest(id: id, name: name)));

  Future<void> setEnabled(String id, bool enabled) async => upsertAccount(await api.setAccountEnabled(pb.SetEnabledRequest(id: id, enabled: enabled)));

  Future<void> remove(String id) async {
    await api.removeAccount(pb.IdRequest(id: id));
    _removeAccount(id);
  }

  void _onEvent(pb.Event ev) {
    if (!_core.state.unlocked) return;
    switch (ev.kind) {
      case pb.Event_Kind.KIND_ACCOUNT:
        // A removed account comes with its id only.
        if (ev.account.provider.isEmpty) {
          _removeAccount(ev.account.id.isNotEmpty ? ev.account.id : ev.accountId);
        } else {
          upsertAccount(ev.account);
        }
      case pb.Event_Kind.KIND_CHAT:
        emit(state.copyWith(chats: {...state.chats, chatKey(ev.chat): ev.chat}));
      case pb.Event_Kind.KIND_CHAT_REMOVED:
        final key = chatKey(ev.chat);
        emit(state.copyWith(chats: Map.of(state.chats)..remove(key), close: state.open == key));
      case pb.Event_Kind.KIND_MESSAGE_NEW:
        _newMessage(ev.message);
      case pb.Event_Kind.KIND_MESSAGE_EDIT || pb.Event_Kind.KIND_MESSAGE_DELETE:
        final key = messageChat(ev.message);
        final c = state.chats[key];
        if (c == null || !c.hasLastMessage() || c.lastMessage.id != ev.message.id) return;
        final m = ev.kind == pb.Event_Kind.KIND_MESSAGE_EDIT ? ev.message : (c.lastMessage.deepCopy()..deleted = true);
        emit(state.copyWith(chats: {...state.chats, key: c.deepCopy()..lastMessage = m}));
      case pb.Event_Kind.KIND_TYPING:
        _typing(messageChat(ev.message), ev.message.sender.name);
      default:
    }
  }

  // The core also sends KIND_CHAT (it owns the unread count); this keeps
  // the preview and the order right even if that one comes later.
  void _newMessage(pb.Message m) {
    final key = messageChat(m);
    if (state.typing.containsKey(key)) _typing(key, null);
    final c = state.chats[key];
    if (c == null) {
      // A chat we have not seen yet: the core knows its title.
      _reloadChats?.cancel();
      _reloadChats = Timer(const Duration(milliseconds: 300), _loadChats);
      return;
    }
    final next = c.deepCopy()
      ..lastMessage = m
      ..updatedMs = m.timeMs > c.updatedMs ? m.timeMs : c.updatedMs;
    emit(state.copyWith(chats: {...state.chats, key: next}));
  }

  void _typing(ChatKey key, String? name) {
    _typingTimers.remove(key)?.cancel();
    final typing = Map.of(state.typing);
    if (name == null) {
      typing.remove(key);
    } else {
      typing[key] = name;
      _typingTimers[key] = Timer(const Duration(seconds: 6), () => _typing(key, null));
    }
    emit(state.copyWith(typing: typing));
  }

  @override
  Future<void> close() async {
    await _coreSub.cancel();
    await _eventSub.cancel();
    _reloadChats?.cancel();
    for (final t in _typingTimers.values) {
      t.cancel();
    }
    return super.close();
  }
}
