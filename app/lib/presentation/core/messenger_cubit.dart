import 'dart:async';

import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/i18n.dart';
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

/// The name a chat is shown with: the saved chat has the app's own.
String chatTitle(pb.Chat c) => c.saved ? tr('Saved Messages', 'Избранное') : c.title;

/// The chat-list view of the archive ([MessengerState.view]).
const archiveView = 'archive';

/// The chat-list view of a chat folder ([MessengerState.view]).
String folderView(pb.ChatFolder f) => 'folder:${f.accountId}/${f.id}';

/// Whether [c] is in the chat folder [f] (rules in the proto's ChatFolder).
bool inFolder(pb.ChatFolder f, pb.Chat c) {
  if (c.accountId != f.accountId) return false;
  if (f.always.contains(c.id)) return true;
  return c.folders.contains(f.id) && !(f.excludeMuted && c.muted) && !(f.excludeRead && c.unread == 0) && !(f.excludeArchived && c.archived);
}

class MessengerState extends Equatable {
  MessengerState({
    this.loading = false,
    this.error,
    this.accounts = const [],
    this.chats = const {},
    this.providers = const [],
    this.typing = const {},
    this.filter = '',
    this.view = '',
    this.folders = const [],
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

  /// What part of the chat list is shown: empty: the chats outside the
  /// archive, [archiveView], or a chat folder ([folderView]).
  final String view;

  /// The accounts' chat folders (CAPABILITY_CHAT_FOLDERS), each account's
  /// in its order.
  final List<pb.ChatFolder> folders;

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

  /// The chat folder shown ([view]), when it still exists.
  late final pb.ChatFolder? folder = folders.where((f) => folderView(f) == view).firstOrNull;

  /// The folders the list offers: those of the shown account (all of them
  /// when every account is shown).
  late final List<pb.ChatFolder> shownFolders = folders.where((f) => filter.isEmpty || f.accountId == filter).toList();

  /// Chats of the shown accounts in the archive.
  late final List<pb.Chat> archived = sorted.where((c) => c.archived && (filter.isEmpty || c.accountId == filter)).toList();

  /// The chat list as shown: pinned chats first (a folder's own pins
  /// first of all, in its order), then by activity.
  late final List<pb.Chat> shown = () {
    final f = folder;
    final list = sorted.where((c) {
      if (filter.isNotEmpty && c.accountId != filter) return false;
      if (unreadOnly && c.unread == 0 && chatKey(c) != open) return false;
      if (view == archiveView) return c.archived;
      if (f != null) return inFolder(f, c);
      return !c.archived || chatKey(c) == open;
    }).toList();
    final pins = f == null ? const <String>[] : f.pinned;
    final first = [for (final id in pins) ...list.where((c) => c.id == id)];
    final rest = list.where((c) => !first.contains(c));
    return [...first, ...rest.where((c) => c.pinned && view != archiveView), ...rest.where((c) => !c.pinned || view == archiveView)];
  }();

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
    String? view,
    List<pb.ChatFolder>? folders,
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
    view: view ?? this.view,
    folders: folders ?? this.folders,
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
    await _loadFolders();
  }

  /// Chat folders; an account that fails keeps none (they come again with
  /// its KIND_CHAT_FOLDERS).
  Future<void> _loadFolders() async {
    try {
      final r = await api.chatFolders(pb.AccountsRequest());
      if (!isClosed) emit(state.copyWith(folders: r.folders));
    } catch (_) {}
  }

  /// Shows all chats, the archive or a folder ([MessengerState.view]).
  void setView(String view) => emit(state.copyWith(view: view));

  /// Mutes, pins, archives or leaves a chat (Account.chat_actions).
  Future<void> manageChat(ChatKey key, pb.ChatAction action) async {
    final c = await api.manageChat(pb.ManageChatRequest(accountId: key.account, chatId: key.chat, action: action));
    if (isClosed) return;
    if (action == pb.ChatAction.CHAT_ACTION_LEAVE) {
      emit(state.copyWith(chats: Map.of(state.chats)..remove(key), close: state.open == key));
    } else if (c.id.isNotEmpty && state.chats.containsKey(key)) {
      emit(state.copyWith(chats: {...state.chats, key: c}));
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

  /// Shows one account (empty: all); a folder of another account closes.
  void setFilter(String accountId) {
    final f = state.folder;
    emit(state.copyWith(filter: accountId, view: f != null && accountId.isNotEmpty && f.accountId != accountId ? '' : null));
  }

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
        folders: state.folders.where((f) => f.accountId != id).toList(),
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

  /// Moves an account (indexes as in onReorder: [to] counts the moved one).
  /// Shown at once; a failed save brings the previous order back.
  Future<void> moveAccount(int from, int to) async {
    if (to > from) to--;
    if (to == from) return;
    final before = state.accounts;
    final list = [...before];
    list.insert(to, list.removeAt(from));
    emit(state.copyWith(accounts: list));
    try {
      await api.reorderAccounts(pb.ReorderAccountsRequest(ids: list.map((a) => a.id)));
    } catch (_) {
      if (!isClosed && identical(state.accounts, list)) emit(state.copyWith(accounts: before));
      rethrow;
    }
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
      case pb.Event_Kind.KIND_CHAT_FOLDERS:
        unawaited(_loadFolders());
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
