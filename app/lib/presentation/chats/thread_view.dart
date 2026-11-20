import 'dart:async';

import 'package:equatable/equatable.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/format.dart';
import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/layout.dart';
import '../../core/theme/typography.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../common/hm.dart';
import '../core/core_cubit.dart';
import '../core/messenger_cubit.dart';
import 'chat_view.dart';
import 'forward.dart';
import 'messages.dart';

/// Opens the thread under [root]: next to the chat on desktops, as its
/// own screen on phones.
void showThread(BuildContext context, ThreadKey key, pb.Message root) {
  final messenger = context.read<MessengerCubit>();
  messenger.openThread(key, root);
  if (!Hm.of(context).mobile) return;
  Navigator.of(context)
      .push(
        MaterialPageRoute<void>(
          builder: (_) => Scaffold(
            backgroundColor: HmColors.bg,
            body: SafeArea(
              child: Builder(
                builder: (context) => ThreadPane(thread: key, onBack: () => Navigator.of(context).maybePop()),
              ),
            ),
          ),
        ),
      )
      .whenComplete(() {
        if (messenger.state.thread == key) messenger.openThread(null);
      });
}

class ThreadState extends Equatable {
  const ThreadState({this.root, this.replies = const [], this.loading = true, this.older = false, this.more = false, this.error, this.rev = 0});

  final pb.Message? root;

  /// Oldest first.
  final List<pb.Message> replies;
  final bool loading;
  final bool older;
  final bool more;
  final String? error;
  final int rev;

  ThreadState copyWith({pb.Message? root, List<pb.Message>? replies, bool? loading, bool? older, bool? more, String? error, bool clearError = false}) =>
      ThreadState(
        root: root ?? this.root,
        replies: replies ?? this.replies,
        loading: loading ?? this.loading,
        older: older ?? this.older,
        more: more ?? this.more,
        error: clearError ? null : error ?? this.error,
        rev: rev + 1,
      );

  @override
  List<Object?> get props => [loading, older, more, error, rev];
}

/// The replies of one thread, page by page, and the live ones. A forum's
/// topic is a thread too ([onRead] set): it is read while it is shown.
class ThreadCubit extends Cubit<ThreadState> {
  ThreadCubit(this._core, this.key, pb.Message? root, {this.onRead}) : super(ThreadState(root: root)) {
    _events = _core.events.listen(_onEvent);
    load();
  }

  static const _page = 50;

  final CoreCubit _core;
  final ThreadKey key;

  /// A topic: called after it was marked read.
  final VoidCallback? onRead;
  late final StreamSubscription<pb.Event> _events;
  Timer? _readTimer;

  void _readSoon() {
    if (onRead == null) return;
    _readTimer?.cancel();
    _readTimer = Timer(const Duration(milliseconds: 400), () async {
      if (isClosed) return;
      try {
        await _api.markRead(pb.MarkReadRequest(accountId: key.account, chatId: key.chat, threadId: key.thread));
        if (!isClosed) onRead!();
      } catch (_) {
        // The next message tries again.
      }
    });
  }

  pb.MegaChatClient get _api => _core.api;

  Future<void> load() async {
    emit(state.copyWith(loading: state.replies.isEmpty, clearError: true));
    try {
      final r = await _api.thread(pb.ThreadRequest(accountId: key.account, chatId: key.chat, threadId: key.thread, limit: _page));
      if (isClosed) return;
      emit(state.copyWith(replies: r.messages.reversed.toList(), more: r.more, loading: false));
      _readSoon();
    } catch (e) {
      if (!isClosed) emit(state.copyWith(loading: false, error: errorText(e)));
    }
  }

  Future<void> loadOlder() async {
    if (!state.more || state.older || state.loading || state.replies.isEmpty) return;
    emit(state.copyWith(older: true));
    try {
      final r = await _api.thread(
        pb.ThreadRequest(accountId: key.account, chatId: key.chat, threadId: key.thread, beforeId: state.replies.first.id, limit: _page),
      );
      if (isClosed) return;
      final known = {for (final m in state.replies) m.id};
      emit(state.copyWith(replies: [...r.messages.reversed.where((m) => !known.contains(m.id)), ...state.replies], more: r.more, older: false));
    } catch (e) {
      if (!isClosed) emit(state.copyWith(older: false, error: errorText(e)));
    }
  }

  /// Sends [text] into the thread; throws so the composer keeps the text.
  Future<void> send(String text) async {
    final m = await _api.send(pb.SendRequest(accountId: key.account, chatId: key.chat, text: text, threadId: key.thread));
    if (!isClosed) _put(m);
  }

  void _onEvent(pb.Event ev) {
    final m = ev.message;
    if (m.accountId != key.account || m.chatId != key.chat) return;
    switch (ev.kind) {
      case pb.Event_Kind.KIND_THREAD_MESSAGE when m.inThread == key.thread:
        _put(m);
        if (!m.outgoing) _readSoon();
      case pb.Event_Kind.KIND_MESSAGE_DELETE when onRead != null:
        final i = state.replies.indexWhere((x) => x.id == m.id);
        if (i >= 0) _put(state.replies[i].deepCopy()..deleted = true);
      case pb.Event_Kind.KIND_MESSAGE_EDIT when m.id == key.thread:
        emit(state.copyWith(root: m));
      default:
    }
  }

  void _put(pb.Message m) {
    final list = state.replies.where((x) => x.id != m.id).toList();
    var i = list.length;
    while (i > 0 && list[i - 1].timeMs > m.timeMs) {
      i--;
    }
    list.insert(i, m);
    emit(state.copyWith(replies: list));
  }

  @override
  Future<void> close() async {
    _readTimer?.cancel();
    await _events.cancel();
    return super.close();
  }
}

/// A thread: its root message on top, the replies, a composer. As a forum's
/// topic ([topic]) it has no header and no root, and sits inside the chat.
class ThreadPane extends StatelessWidget {
  const ThreadPane({super.key, required this.thread, this.onBack, this.onClose, this.topic, this.onTopicRead, this.readOnly = false});

  final ThreadKey thread;

  /// The forum topic shown (its title for the composer).
  final pb.Topic? topic;

  /// A topic was marked read.
  final VoidCallback? onTopicRead;

  /// No composer (the chat or the topic is closed for the account).
  final bool readOnly;

  /// Phone or a narrow window: back to the chat.
  final VoidCallback? onBack;

  /// Beside the chat: close the pane.
  final VoidCallback? onClose;

  @override
  Widget build(BuildContext context) => BlocProvider(
    key: ValueKey(thread),
    create: (c) => topic != null
        ? ThreadCubit(c.read<CoreCubit>(), thread, null, onRead: onTopicRead ?? () {})
        : ThreadCubit(c.read<CoreCubit>(), thread, c.read<MessengerCubit>().state.threadRoot),
    child: _ThreadBody(thread: thread, onBack: onBack, onClose: onClose, topic: topic, readOnly: readOnly),
  );
}

class _ThreadBody extends StatefulWidget {
  const _ThreadBody({required this.thread, this.onBack, this.onClose, this.topic, this.readOnly = false});

  final ThreadKey thread;
  final VoidCallback? onBack;
  final VoidCallback? onClose;
  final pb.Topic? topic;
  final bool readOnly;

  @override
  State<_ThreadBody> createState() => _ThreadBodyState();
}

class _ThreadBodyState extends State<_ThreadBody> {
  final _scroll = ScrollController();

  @override
  void initState() {
    super.initState();
    _scroll.addListener(() {
      final p = _scroll.position;
      if (p.pixels > p.maxScrollExtent - 400) context.read<ThreadCubit>().loadOlder();
    });
  }

  @override
  void dispose() {
    _scroll.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    final chat = context.select<MessengerCubit, pb.Chat?>((c) => c.state.chats[(account: widget.thread.account, chat: widget.thread.chat)]);
    final s = context.watch<ThreadCubit>().state;
    final cubit = context.read<ThreadCubit>();
    final direct = chat?.kind == pb.ChatKind.CHAT_KIND_DIRECT;
    final topic = widget.topic;

    // Newest first, as the list is reversed; the root closes it at the top.
    // A topic has no root; a day note goes above each of its days.
    // An album is one message.
    final replies = [for (final g in groupAlbums(s.replies)) g.message];
    final entries = <Object>[
      if (s.error != null) #error else if (s.loading && replies.isEmpty) #loading,
      for (var i = replies.length - 1; i >= 0; i--) ...[
        replies[i],
        if (topic != null && (i == 0 || !sameDay(fromMs(replies[i].timeMs), fromMs(replies[i - 1].timeMs)))) formatDay(fromMs(replies[i].timeMs)),
      ],
      if (s.more || s.older) #older,
      ?s.root,
    ];

    Widget body;
    if (s.loading && s.root == null) {
      body = const HmLoading();
    } else if (topic != null && s.replies.isEmpty && s.error == null) {
      body = Center(child: HmEmptyState(tr('No messages in this topic yet', 'В этой теме пока нет сообщений'), icon: Icons.tag));
    } else {
      body = ListView.builder(
        controller: _scroll,
        reverse: true,
        padding: EdgeInsets.symmetric(horizontal: hm.mobile ? 12 : 16, vertical: 12),
        itemCount: entries.length,
        itemBuilder: (context, i) {
          final e = entries[i];
          if (e == #error) {
            return Padding(
              padding: const EdgeInsets.symmetric(vertical: 12),
              child: HmBanner(
                s.error!,
                tone: HmTone.error,
                action: HmButton(tr('Retry', 'Повторить'), small: true, onPressed: cubit.load),
              ),
            );
          }
          if (e == #loading) {
            return const Padding(
              padding: EdgeInsets.symmetric(vertical: 12),
              child: Center(child: SizedBox.square(dimension: 20, child: HmSpinner())),
            );
          }
          if (e is String) return HmServiceNote(e);
          if (e == #older) {
            return Padding(
              padding: const EdgeInsets.symmetric(vertical: 12),
              child: Center(
                child: s.older
                    ? const SizedBox.square(dimension: 20, child: HmSpinner())
                    : HmButton(
                        topic != null ? tr('Load older', 'Загрузить ранние') : tr('Load earlier replies', 'Загрузить ранние ответы'),
                        small: true,
                        onPressed: cubit.loadOlder,
                      ),
              ),
            );
          }
          final m = e as pb.Message;
          if (identical(m, s.root)) {
            final n = m.hasThread() && m.thread.replies > 0 ? m.thread.replies : s.replies.length;
            return Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                MessageTile(message: m, quoted: null, showSender: !direct && !m.outgoing, gap: false, onForward: () => forwardMessages(context, [m])),
                HmServiceNote(n == 0 && !s.loading ? tr('No replies yet', 'Ответов пока нет') : '${tr('Replies', 'Ответы')}: $n'),
              ],
            );
          }
          final prev = i + 1 < entries.length ? entries[i + 1] : null;
          final first = prev is! pb.Message || identical(prev, s.root) || prev.service || prev.sender.id != m.sender.id;
          final quoted = topic != null && m.replyToId.isNotEmpty ? s.replies.where((x) => x.id == m.replyToId).firstOrNull : null;
          // A topic's messages are in the chat's history; a thread's replies
          // may live elsewhere (Telegram comments): those go as text.
          return MessageTile(
            message: m,
            quoted: quoted,
            showSender: !direct && !m.outgoing && first,
            gap: first,
            onForward: () => forwardMessages(context, [m], native: topic != null),
          );
        },
      );
    }

    if (topic != null) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Expanded(child: body),
          if (widget.readOnly)
            Container(
              padding: const EdgeInsets.symmetric(vertical: 10),
              decoration: const BoxDecoration(
                color: HmColors.surface,
                border: Border(top: BorderSide(color: HmColors.border)),
              ),
              child: Text(tr('You can only read this chat', 'В этом чате можно только читать'), textAlign: TextAlign.center, style: HmText.footer),
            )
          else
            _ThreadComposer(
              placeholder: topic.closed
                  ? tr('The topic is closed: only admins can write', 'Тема закрыта: писать могут только админы')
                  : tr('Message to «${topic.title}»', 'Сообщение в «${topic.title}»'),
            ),
        ],
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Container(
          height: hm.mobile ? 56 : 64,
          padding: EdgeInsets.symmetric(horizontal: hm.mobile ? 8 : 16),
          decoration: const BoxDecoration(
            color: HmColors.surface,
            border: Border(bottom: BorderSide(color: HmColors.border)),
          ),
          child: Row(
            children: [
              if (widget.onBack != null) ...[
                HmIconButton(icon: Icons.arrow_back, label: tr('Back to the chat', 'Назад к чату'), onPressed: widget.onBack),
                const SizedBox(width: 8),
              ],
              Expanded(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(tr('Thread', 'Тред'), style: HmText.bodyStrong),
                    const SizedBox(height: 2),
                    Text(chat?.title ?? '', style: HmText.footer, maxLines: 1, overflow: TextOverflow.ellipsis),
                  ],
                ),
              ),
              if (widget.onClose != null) HmIconButton(icon: Icons.close, label: tr('Close the thread', 'Закрыть тред'), onPressed: widget.onClose),
            ],
          ),
        ),
        Expanded(child: body),
        const _ThreadComposer(),
      ],
    );
  }
}

class _ThreadComposer extends StatefulWidget {
  const _ThreadComposer({this.placeholder});

  final String? placeholder;

  @override
  State<_ThreadComposer> createState() => _ThreadComposerState();
}

class _ThreadComposerState extends State<_ThreadComposer> {
  final _text = TextEditingController();
  final _focus = FocusNode();

  @override
  void initState() {
    super.initState();
    _text.addListener(() => setState(() {}));
  }

  @override
  void dispose() {
    _text.dispose();
    _focus.dispose();
    super.dispose();
  }

  Future<void> _send() async {
    final text = _text.text.trim();
    if (text.isEmpty) return;
    _text.clear();
    try {
      await context.read<ThreadCubit>().send(text);
    } catch (e) {
      if (_text.text.isEmpty) _text.text = text;
      Toasts.error(tr('Not sent', 'Не отправлено'), errorText(e));
    }
  }

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    return Container(
      padding: EdgeInsets.fromLTRB(hm.mobile ? 12 : 16, 10, hm.mobile ? 8 : 16, 10),
      decoration: const BoxDecoration(
        color: HmColors.surface,
        border: Border(top: BorderSide(color: HmColors.border)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          Expanded(
            child: HmMessageField(
              controller: _text,
              focusNode: _focus,
              onSend: _send,
              placeholder: widget.placeholder ?? tr('Reply in thread', 'Ответ в треде'),
            ),
          ),
          const SizedBox(width: 8),
          HmIconButton(icon: Icons.send, label: tr('Send', 'Отправить'), color: HmColors.accent, onPressed: _text.text.trim().isEmpty ? null : _send),
        ],
      ),
    );
  }
}
