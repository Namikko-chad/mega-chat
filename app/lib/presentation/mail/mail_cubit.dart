import 'dart:async';

import 'package:equatable/equatable.dart';
import 'package:fixnum/fixnum.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../core/core_cubit.dart';

/// Mail ids are unique only inside an account.
typedef MailKey = ({String account, String id});

MailKey mailKey(pb.Mail m) => (account: m.accountId, id: m.id);

/// What the list shows: the folders of one role of every account (account
/// empty: all inboxes, all sent, …) or one folder of one account.
typedef MailView = ({pb.FolderRole role, String account, String folder});

const MailView allInboxes = (role: pb.FolderRole.FOLDER_ROLE_INBOX, account: '', folder: '');

bool hasMail(pb.Account a) => a.capabilities.contains(pb.Capability.CAPABILITY_MAIL);

class MailState extends Equatable {
  const MailState({
    this.loading = false,
    this.error,
    this.folders = const [],
    this.failed = const [],
    this.view = allInboxes,
    this.mails = const [],
    this.next = '',
    this.more = false,
    this.loadingMore = false,
    this.open,
    this.body,
    this.bodyError,
    this.rev = 0,
  });

  final bool loading;
  final String? error;
  final List<pb.Folder> folders;

  /// Accounts whose part of the last read failed, and why.
  final List<pb.AccountError> failed;
  final MailView view;
  final List<pb.Mail> mails;
  final String next;
  final bool more;
  final bool loadingMore;

  /// The message open in the reader, and its body once loaded.
  final MailKey? open;
  final pb.Mail? body;
  final String? bodyError;
  final int rev;

  /// Unread of one role over all accounts.
  int unreadOf(pb.FolderRole role) => folders.where((f) => f.role == role).fold(0, (n, f) => n + f.unread);

  List<pb.Folder> foldersOf(String account) => folders.where((f) => f.accountId == account).toList();

  pb.Mail? summary(MailKey k) => mails.where((m) => mailKey(m) == k).firstOrNull;

  MailState copyWith({
    bool? loading,
    String? error,
    List<pb.Folder>? folders,
    List<pb.AccountError>? failed,
    MailView? view,
    List<pb.Mail>? mails,
    String? next,
    bool? more,
    bool? loadingMore,
    MailKey? open,
    pb.Mail? body,
    String? bodyError,
    bool clearError = false,
    bool close = false,
    bool clearBody = false,
  }) => MailState(
    loading: loading ?? this.loading,
    error: clearError ? null : error ?? this.error,
    folders: folders ?? this.folders,
    failed: failed ?? this.failed,
    view: view ?? this.view,
    mails: mails ?? this.mails,
    next: next ?? this.next,
    more: more ?? this.more,
    loadingMore: loadingMore ?? this.loadingMore,
    open: close ? null : open ?? this.open,
    body: close || clearBody ? null : body ?? this.body,
    bodyError: close || clearBody ? null : bodyError ?? this.bodyError,
    rev: rev + 1,
  );

  @override
  List<Object?> get props => [rev];
}

/// The merged mailbox of every mail account. Nothing is cached in the
/// core: the cubit keeps the pages it read and merges what changes.
class MailCubit extends Cubit<MailState> {
  MailCubit(this._core) : super(MailState()) {
    _coreSub = _core.stream.listen(_onCore);
    _eventSub = _core.events.listen(_onEvent);
    _onCore(_core.state);
  }

  static const pageSize = 40;

  final CoreCubit _core;
  late final StreamSubscription<CoreState> _coreSub;
  late final StreamSubscription<pb.Event> _eventSub;
  int? _session;
  Timer? _reload;
  int _gen = 0; // bumps when the view changes: late answers are dropped

  pb.MegaChatClient get api => _core.api;

  void _onCore(CoreState s) {
    if (!s.unlocked) {
      _session = null;
      if (state.mails.isNotEmpty || state.folders.isNotEmpty) emit(MailState());
      return;
    }
    if (s.conn != Conn.ready || s.session == _session) return;
    _session = s.session;
    unawaited(load());
  }

  void _onEvent(pb.Event ev) {
    if (!_core.state.unlocked) return;
    switch (ev.kind) {
      case pb.Event_Kind.KIND_MAIL || pb.Event_Kind.KIND_MAIL_NEW:
        _scheduleReload();
      case pb.Event_Kind.KIND_ACCOUNT:
        // An account came online or went away: its folders change.
        if (ev.account.provider.isEmpty || hasMail(ev.account)) _scheduleReload();
      default:
    }
  }

  void _scheduleReload() {
    _reload?.cancel();
    _reload = Timer(const Duration(milliseconds: 500), refresh);
  }

  pb.MailListRequest _request(MailView v, String cursor) =>
      pb.MailListRequest(accountIds: [if (v.account.isNotEmpty) v.account], role: v.role, folderId: v.folder, cursor: cursor, limit: pageSize);

  Future<void> load() async {
    final gen = ++_gen;
    emit(state.copyWith(loading: true, clearError: true));
    try {
      final r = await Future.wait([api.folders(pb.AccountsRequest()), api.mailList(_request(state.view, ''))]);
      if (isClosed || gen != _gen) return;
      final f = r[0] as pb.FoldersReply, l = r[1] as pb.MailListReply;
      emit(state.copyWith(loading: false, folders: f.folders, failed: [...f.failed, ...l.failed], mails: l.mails, next: l.next, more: l.more));
    } catch (e) {
      if (!isClosed && gen == _gen) emit(state.copyWith(loading: false, error: errorText(e)));
    }
  }

  /// Folder counts and the newest page, merged into what is shown: new
  /// messages come in, changed ones update, gone ones (in that range) go.
  Future<void> refresh() async {
    final gen = _gen;
    try {
      final r = await Future.wait([api.folders(pb.AccountsRequest()), api.mailList(_request(state.view, ''))]);
      if (isClosed || gen != _gen) return;
      final f = r[0] as pb.FoldersReply, l = r[1] as pb.MailListReply;
      final fresh = {for (final m in l.mails) mailKey(m): m};
      final oldest = l.mails.isEmpty || !l.more ? Int64.ZERO : l.mails.last.timeMs;
      final merged = <pb.Mail>[
        ...l.mails,
        for (final m in state.mails)
          if (!fresh.containsKey(mailKey(m)) && m.timeMs < oldest) m,
      ]..sort((a, b) => b.timeMs.compareTo(a.timeMs));
      // The reader stays open; a trashed message closes it.
      final gone = state.open != null && !merged.any((m) => mailKey(m) == state.open) && state.mails.any((m) => mailKey(m) == state.open);
      emit(
        state.copyWith(
          folders: f.folders,
          failed: [...f.failed, ...l.failed],
          mails: merged,
          next: state.mails.length > l.mails.length ? state.next : l.next,
          more: state.mails.length > l.mails.length ? state.more : l.more,
          close: gone,
        ),
      );
    } catch (_) {}
  }

  Future<void> loadMore() async {
    if (!state.more || state.loadingMore || state.next.isEmpty) return;
    final gen = _gen;
    emit(state.copyWith(loadingMore: true));
    try {
      final l = await api.mailList(_request(state.view, state.next));
      if (isClosed || gen != _gen) return;
      final have = {for (final m in state.mails) mailKey(m)};
      emit(
        state.copyWith(
          loadingMore: false,
          mails: [...state.mails, ...l.mails.where((m) => !have.contains(mailKey(m)))],
          next: l.next,
          more: l.more,
          failed: l.failed,
        ),
      );
    } catch (e) {
      if (!isClosed && gen == _gen) emit(state.copyWith(loadingMore: false, error: errorText(e)));
    }
  }

  void setView(MailView v) {
    if (v == state.view && state.mails.isNotEmpty) return;
    emit(MailState(folders: state.folders, view: v, loading: true));
    unawaited(load());
  }

  /// Opens a message: its body is fetched and it is marked read.
  Future<void> openMail(MailKey? k) async {
    if (k == null) {
      emit(state.copyWith(close: true));
      return;
    }
    emit(state.copyWith(open: k, clearBody: true));
    try {
      final m = await api.getMail(pb.GetMailRequest(accountId: k.account, id: k.id));
      if (isClosed || state.open != k) return;
      emit(state.copyWith(body: m));
      if (m.unread || (state.summary(k)?.unread ?? false)) unawaited(act(k, pb.MailAction.MAIL_ACTION_READ));
    } catch (e) {
      if (!isClosed && state.open == k) emit(state.copyWith(bodyError: errorText(e)));
    }
  }

  /// Marks, stars or trashes a message; the list follows at once, the
  /// core confirms with KIND_MAIL.
  Future<void> act(MailKey k, pb.MailAction a) async {
    void patch(void Function(pb.Mail m) fn) {
      pb.Mail changed(pb.Mail m) {
        final c = m.deepCopy();
        fn(c);
        return c;
      }

      final mails = [for (final m in state.mails) mailKey(m) == k ? changed(m) : m];
      final body = state.body != null && mailKey(state.body!) == k ? changed(state.body!) : null;
      emit(state.copyWith(mails: mails, body: body));
    }

    switch (a) {
      case pb.MailAction.MAIL_ACTION_READ:
        patch((m) => m.unread = false);
      case pb.MailAction.MAIL_ACTION_UNREAD:
        patch((m) => m.unread = true);
      case pb.MailAction.MAIL_ACTION_STAR:
        patch((m) => m.starred = true);
      case pb.MailAction.MAIL_ACTION_UNSTAR:
        patch((m) => m.starred = false);
      case pb.MailAction.MAIL_ACTION_TRASH:
        emit(state.copyWith(mails: state.mails.where((m) => mailKey(m) != k).toList(), close: state.open == k));
      default:
    }
    await api.updateMail(pb.UpdateMailRequest(accountId: k.account, ids: [k.id], action: a));
  }

  Future<void> send(pb.SendMailRequest r) async {
    await api.sendMail(r);
  }

  @override
  Future<void> close() async {
    await _coreSub.cancel();
    await _eventSub.cancel();
    _reload?.cancel();
    return super.close();
  }
}
