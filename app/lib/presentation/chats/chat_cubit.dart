import 'dart:async';

import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../core/core_cubit.dart';
import '../core/messenger_cubit.dart';

class ChatState extends Equatable {
  const ChatState({
    this.messages = const [],
    this.loading = true,
    this.older = false,
    this.more = false,
    this.stale = false,
    this.firstUnread,
    this.error,
    this.replyTo,
    this.quote,
    this.selected = const {},
    this.rev = 0,
  });

  /// Oldest first.
  final List<pb.Message> messages;
  final bool loading; // the first page
  final bool older; // an older page is on its way
  final bool more; // older messages may exist

  /// What the cache had, shown while the real page loads (may have gaps).
  final bool stale;

  /// The oldest message that was unread when the chat was opened: the
  /// view opens there, under an "unread messages" line.
  final String? firstUnread;
  final String? error;
  final pb.Message? replyTo;

  /// The fragment of [replyTo]'s text the reply quotes.
  final String? quote;

  /// Messages picked to copy or forward together (the ids shown: an
  /// album's is its captioned part). Empty: not picking.
  final Set<String> selected;
  final int rev;

  pb.Message? byId(String id) => messages.where((m) => m.id == id).firstOrNull;

  ChatState copyWith({
    List<pb.Message>? messages,
    bool? loading,
    bool? older,
    bool? more,
    bool? stale,
    String? firstUnread,
    String? error,
    pb.Message? replyTo,
    String? quote,
    Set<String>? selected,
    bool clearError = false,
    bool clearReply = false,
  }) => ChatState(
    messages: messages ?? this.messages,
    loading: loading ?? this.loading,
    older: older ?? this.older,
    more: more ?? this.more,
    stale: stale ?? this.stale,
    firstUnread: firstUnread ?? this.firstUnread,
    error: clearError ? null : error ?? this.error,
    replyTo: clearReply ? null : replyTo ?? this.replyTo,
    quote: clearReply ? null : (replyTo != null ? quote : this.quote),
    selected: selected ?? this.selected,
    rev: rev + 1,
  );

  @override
  List<Object?> get props => [loading, older, more, stale, firstUnread, error, rev];
}

/// One open chat: its history page by page and live message events.
class ChatCubit extends Cubit<ChatState> {
  ChatCubit(this._core, this._messenger, this.key) : super(const ChatState()) {
    _session = _core.state.session;
    final acct = _messenger.state.accounts.where((a) => a.id == key.account).firstOrNull;
    if (acct != null) _online = acct.status == pb.AccountStatus.ACCOUNT_STATUS_ONLINE;
    _openUnread = _messenger.state.chats[key]?.unread ?? 0;
    _eventSub = _core.events.listen(_onEvent);
    // Events were missed while the core was away: start from the latest page.
    _coreSub = _core.stream.listen((s) {
      if (s.conn == Conn.ready && s.unlocked && s.session != _session) {
        _session = s.session;
        load();
      }
    });
    load();
  }

  static const _page = 50;

  final CoreCubit _core;
  final MessengerCubit _messenger;
  final ChatKey key;
  late final StreamSubscription<pb.Event> _eventSub;
  late final StreamSubscription<CoreState> _coreSub;
  int _session = 0;
  Timer? _readTimer;
  Timer? _reloadTimer;
  bool? _online; // the account's last known state

  // Read only while the user can see the chat (window in front, screen on),
  // and only up to messages the view has really shown (seen).
  bool _visible = true;
  int _openUnread = 0; // the chat's unread count when it was opened
  int? _readMs; // messages up to this time are read; null: not known yet
  bool _readSent = false; // the read point moved since opening
  pb.Message? _seen; // the newest message shown and not yet marked read

  pb.MegaChatClient get _api => _core.api;

  Future<void> load() async {
    emit(state.copyWith(loading: state.messages.isEmpty, clearError: true));
    // What the cache has shows at once; the real page (gaps filled from
    // the messenger) replaces it.
    if (state.messages.isEmpty) {
      try {
        final c = await _api.history(pb.HistoryRequest(accountId: key.account, chatId: key.chat, limit: _page, cachedOnly: true));
        if (isClosed) return;
        if (c.messages.isNotEmpty && state.messages.isEmpty) {
          final list = c.messages.reversed.toList();
          emit(state.copyWith(messages: list, stale: true, more: false, loading: false, firstUnread: _anchor(list)));
        }
      } catch (_) {
        // An older core without cached_only, or none: wait for the page.
      }
    }
    try {
      final r = await _api.history(pb.HistoryRequest(accountId: key.account, chatId: key.chat, limit: _page));
      if (isClosed) return;
      final list = r.messages.reversed.toList();
      emit(state.copyWith(messages: list, more: r.more, loading: false, stale: false, firstUnread: _anchor(list)));
    } catch (e) {
      if (!isClosed) emit(state.copyWith(loading: false, error: errorText(e)));
    }
  }

  Future<void> loadOlder() async {
    if (!state.more || state.older || state.loading || state.stale || state.messages.isEmpty) return;
    emit(state.copyWith(older: true));
    try {
      final r = await _api.history(pb.HistoryRequest(accountId: key.account, chatId: key.chat, beforeId: state.messages.first.id, limit: _page));
      if (isClosed) return;
      final known = {for (final m in state.messages) m.id};
      final older = r.messages.reversed.where((m) => !known.contains(m.id));
      emit(state.copyWith(messages: [...older, ...state.messages], more: r.more, older: false));
    } catch (e) {
      if (!isClosed) emit(state.copyWith(older: false, error: errorText(e)));
    }
  }

  /// Replies to [m] (null: cancels), quoting [quote] of its text.
  void reply(pb.Message? m, {String? quote}) {
    final q = quote?.trim();
    emit(m == null ? state.copyWith(clearReply: true) : state.copyWith(replyTo: m, quote: q == null || q.isEmpty ? null : q));
  }

  /// Picks [id] for copying or forwarding, or drops it.
  void toggleSelected(String id) {
    final s = {...state.selected};
    if (!s.remove(id)) s.add(id);
    emit(state.copyWith(selected: s));
  }

  void clearSelected() {
    if (state.selected.isNotEmpty) emit(state.copyWith(selected: const {}));
  }

  /// Sends [text]; throws so the composer can keep the text on failure.
  Future<void> send(String text) async {
    final reply = state.replyTo;
    final m = await _api.send(pb.SendRequest(accountId: key.account, chatId: key.chat, text: text, replyToId: reply?.id ?? '', quote: state.quote ?? ''));
    if (isClosed) return;
    _put(m);
    if (state.replyTo == reply) emit(state.copyWith(clearReply: true));
  }

  /// Adds or takes back the account's reaction [emoji] on [m].
  Future<void> react(pb.Message m, String emoji) async {
    final mine = m.reactions.any((r) => r.emoji == emoji && r.mine);
    final r = await _api.react(pb.ReactRequest(accountId: key.account, chatId: key.chat, messageId: m.id, emoji: emoji, on: !mine));
    if (isClosed) return;
    final cur = state.byId(m.id);
    if (cur == null) return;
    _put(
      cur.deepCopy()
        ..reactions.clear()
        ..reactions.addAll(r.reactions),
    );
  }

  void setVisible(bool v) {
    _visible = v;
    if (v) _readSoon();
  }

  /// A forum shows one of its topics instead of the whole history: the
  /// topic is read on its own, the chat is not marked read meanwhile.
  void setInTopic(bool v) {
    _inTopic = v;
    if (!v) _readSoon();
  }

  bool _inTopic = false;

  /// The first unread message of [list] (oldest first) by the count the
  /// chat had when opened; also sets where reading starts. Fixed once the
  /// user has read something here.
  String? _anchor(List<pb.Message> list) {
    if (_readSent || list.isEmpty) return state.firstUnread;
    if (_openUnread <= 0) {
      _readMs = list.last.timeMs.toInt();
      return null;
    }
    var n = 0;
    for (var i = list.length - 1; i >= 0; i--) {
      if (list[i].outgoing) continue;
      if (++n == _openUnread) {
        _readMs = i > 0 ? list[i - 1].timeMs.toInt() : list[i].timeMs.toInt() - 1;
        return list[i].id;
      }
    }
    // More unread than loaded: all of the page is unread.
    final first = list.firstWhere((m) => !m.outgoing, orElse: () => list.first);
    _readMs = first.timeMs.toInt() - 1;
    return first.id;
  }

  /// The view has shown [m] on screen: the chat is read up to it.
  void seen(pb.Message m) {
    final read = _readMs;
    if (m.outgoing || read == null) return;
    final t = m.timeMs.toInt();
    if (t <= read || (_seen != null && t <= _seen!.timeMs.toInt())) return;
    _seen = m;
    _readSoon();
  }

  void _readSoon() {
    if (_seen == null || !_visible || _inTopic) return;
    _readTimer?.cancel();
    _readTimer = Timer(const Duration(milliseconds: 300), () async {
      final m = _seen;
      if (m == null || !_visible || _inTopic || isClosed) return;
      try {
        await _api.markRead(pb.MarkReadRequest(accountId: key.account, chatId: key.chat, messageId: m.id));
        if (isClosed) return;
        _readSent = true;
        final t = m.timeMs.toInt();
        if (t > (_readMs ?? 0)) _readMs = t;
        if (_seen == m) _seen = null;
        final left = state.messages.where((x) => !x.outgoing && x.timeMs.toInt() > t).length;
        _messenger.markedRead(key, unread: left);
      } catch (_) {
        // The next message shown or focus tries again.
      }
    });
  }

  /// Reloads the latest page soon (several triggers may come together).
  void _reloadSoon() {
    _reloadTimer?.cancel();
    _reloadTimer = Timer(const Duration(milliseconds: 300), () {
      if (!isClosed) load();
    });
  }

  void _onEvent(pb.Event ev) {
    switch (ev.kind) {
      // Back online: the core fills what the session missed while down.
      case pb.Event_Kind.KIND_ACCOUNT when ev.account.id == key.account:
        final online = ev.account.status == pb.AccountStatus.ACCOUNT_STATUS_ONLINE;
        if (online && _online == false) _reloadSoon();
        _online = online;
        return;
      // The chat's latest message is newer than anything shown and was
      // never delivered here: the view has a gap at the bottom.
      case pb.Event_Kind.KIND_CHAT when ev.chat.accountId == key.account && ev.chat.id == key.chat:
        final last = ev.chat.hasLastMessage() ? ev.chat.lastMessage : null;
        if (last == null || state.loading || state.byId(last.id) != null) return;
        if (state.messages.isEmpty || last.timeMs >= state.messages.last.timeMs) _reloadSoon();
        return;
      default:
    }
    final m = ev.message;
    if (m.accountId != key.account || m.chatId != key.chat) return;
    switch (ev.kind) {
      case pb.Event_Kind.KIND_MESSAGE_NEW:
        _put(m); // read when the view shows it
      case pb.Event_Kind.KIND_MESSAGE_EDIT:
        if (state.byId(m.id) != null) _put(m);
      case pb.Event_Kind.KIND_MESSAGE_DELETE:
        final old = state.byId(m.id);
        if (old != null) _put(old.deepCopy()..deleted = true);
      default:
    }
  }

  /// Inserts or replaces a message, keeping the list in time order. The
  /// reply of Send and the NEW event of the same message meet here.
  void _put(pb.Message m) {
    final list = state.messages.where((x) => x.id != m.id).toList();
    var i = list.length;
    while (i > 0 && list[i - 1].timeMs > m.timeMs) {
      i--;
    }
    list.insert(i, m);
    emit(state.copyWith(messages: list));
  }

  @override
  Future<void> close() async {
    _readTimer?.cancel();
    _reloadTimer?.cancel();
    await _eventSub.cancel();
    await _coreSub.cancel();
    return super.close();
  }
}
