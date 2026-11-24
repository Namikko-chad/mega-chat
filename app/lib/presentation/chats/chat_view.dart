import 'dart:io';

import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/format.dart';
import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/layout.dart';
import '../../core/theme/typography.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../accounts/login_wizard.dart';
import '../accounts/providers.dart';
import '../common/hm.dart';
import '../call/call_cubit.dart';
import '../core/core_cubit.dart';
import '../core/messenger_cubit.dart';
import '../../data/local_files.dart';
import 'attach.dart';
import 'chat_cubit.dart';
import 'forward.dart';
import 'media.dart';
import 'messages.dart';
import 'link_preview.dart';
import 'mentions.dart';
import 'profile_view.dart';
import 'reactions.dart';
import 'thread_view.dart';
import 'topics.dart';
import 'chat_actions.dart';

/// One open chat: header, messages (newest at the bottom, older pages
/// load on scrolling up) and the composer.
class ChatPane extends StatelessWidget {
  const ChatPane({super.key, required this.chat, this.onBack});

  final ChatKey chat;

  /// Phone: the back button of the pushed screen.
  final VoidCallback? onBack;

  @override
  Widget build(BuildContext context) => BlocProvider(
    key: ValueKey(chat),
    create: (c) => ChatCubit(c.read<CoreCubit>(), c.read<MessengerCubit>(), chat),
    child: _ChatBody(chat: chat, onBack: onBack),
  );
}

class _ChatBody extends StatefulWidget {
  const _ChatBody({required this.chat, this.onBack});

  final ChatKey chat;
  final VoidCallback? onBack;

  @override
  State<_ChatBody> createState() => _ChatBodyState();
}

class _ChatBodyState extends State<_ChatBody> {
  final _scroll = ScrollController();
  late final ChatCubit _cubit = context.read<ChatCubit>();
  late final AppLifecycleListener _life;
  late final _seen = _SeenTracker(_cubit, _scroll);

  @override
  void initState() {
    super.initState();
    // The list is reversed: its end is the top, where older messages are.
    _scroll.addListener(() {
      final p = _scroll.position;
      if (p.pixels > p.maxScrollExtent - 400) _cubit.loadOlder();
      _seen.schedule();
    });
    final now = WidgetsBinding.instance.lifecycleState;
    _cubit.setVisible(now == null || now == AppLifecycleState.resumed);
    _life = AppLifecycleListener(
      onStateChange: (s) {
        _cubit.setVisible(s == AppLifecycleState.resumed);
        if (s == AppLifecycleState.resumed) _seen.schedule();
      },
    );
  }

  @override
  void dispose() {
    _seen.dispose();
    _life.dispose();
    _scroll.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final m = context.watch<MessengerCubit>().state;
    final chat = m.chats[widget.chat];
    final account = m.account(widget.chat.account);
    final forum = (chat?.forum ?? false) && (account?.capabilities ?? const []).contains(pb.Capability.CAPABILITY_TOPICS);
    if (!forum) return _body(context, m, chat, account, null);
    // A forum: a strip of its topics; a chosen topic replaces the history.
    return BlocProvider(
      create: (c) => TopicsCubit(c.read<CoreCubit>(), widget.chat),
      child: BlocConsumer<TopicsCubit, TopicsState>(
        listenWhen: (a, b) => a.selected != b.selected,
        listener: (_, s) => _cubit.setInTopic(s.selected != null),
        builder: (context, topics) => _body(context, m, chat, account, topics),
      ),
    );
  }

  Widget _body(BuildContext context, MessengerState m, pb.Chat? chat, pb.Account? account, TopicsState? topics) {
    final topic = topics?.current;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        BlocSelector<ChatCubit, ChatState, int>(
          selector: (s) => s.selected.length,
          builder: (context, n) =>
              n > 0 && topic == null ? _SelectionBar(count: n) : _Header(chat: chat, account: account, typing: m.typing[widget.chat], onBack: widget.onBack),
        ),
        if (topics != null) TopicBar(chat: chat),
        if (topic != null)
          Expanded(
            child: ThreadPane(
              thread: (account: widget.chat.account, chat: widget.chat.chat, thread: topic.id),
              topic: topic,
              readOnly: chat?.readOnly ?? false,
              onTopicRead: () => context.read<TopicsCubit>().markedRead(topic.id),
            ),
          ),
        if (topic == null) ...[
          Expanded(
            child: _Messages(
              scroll: _scroll,
              seen: _seen,
              direct: chat?.kind == pb.ChatKind.CHAT_KIND_DIRECT,
              canReact: (account?.capabilities ?? const []).contains(pb.Capability.CAPABILITY_REACTIONS) && !(chat?.readOnly ?? false),
              account: chat?.readOnly ?? false ? null : account,
              canThread: (account?.capabilities ?? const []).contains(pb.Capability.CAPABILITY_THREADS),
              openThread: m.thread,
            ),
          ),
          if (account != null && account.status == pb.AccountStatus.ACCOUNT_STATUS_NEEDS_LOGIN)
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
              child: HmBanner(
                tr('This account is signed out; messages are not sent.', 'Аккаунт вышел из сети; сообщения не отправляются.'),
                tone: HmTone.warning,
                action: HmButton(tr('Sign in again', 'Войти снова'), small: true, onPressed: () => showLoginWizard(context, relogin: account)),
              ),
            ),
          // A channel the account can only read has no composer.
          if (chat?.readOnly ?? false)
            Container(
              padding: const EdgeInsets.symmetric(vertical: 10),
              decoration: const BoxDecoration(
                color: HmColors.surface,
                border: Border(top: BorderSide(color: HmColors.border)),
              ),
              child: Text(tr('You can only read this chat', 'В этом чате можно только читать'), textAlign: TextAlign.center, style: HmText.footer),
            )
          else
            _Composer(chat: widget.chat),
        ],
      ],
    );
  }
}

class _Header extends StatelessWidget {
  const _Header({required this.chat, required this.account, required this.typing, required this.onBack});

  final pb.Chat? chat;
  final pb.Account? account;
  final String? typing;
  final VoidCallback? onBack;

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    final provider = account?.provider ?? '';
    final info = context.read<MessengerCubit>().state.provider(provider);
    final where = [if (account != null) account!.name, providerName(provider, info)].where((s) => s.isNotEmpty).join(' · ');
    final online = chat != null && chat!.kind == pb.ChatKind.CHAT_KIND_DIRECT && chat!.online;
    return Container(
      height: hm.mobile ? 56 : 64,
      padding: EdgeInsets.symmetric(horizontal: hm.mobile ? 8 : 20),
      decoration: const BoxDecoration(
        color: HmColors.surface,
        border: Border(bottom: BorderSide(color: HmColors.border)),
      ),
      child: Row(
        children: [
          if (onBack != null) ...[HmIconButton(icon: Icons.arrow_back, label: tr('Back', 'Назад'), onPressed: onBack), const SizedBox(width: 8)],
          Expanded(
            child: MouseRegion(
              cursor: chat == null ? MouseCursor.defer : SystemMouseCursors.click,
              child: GestureDetector(
                behavior: HitTestBehavior.opaque,
                onTap: chat == null ? null : () => showProfile(context, (account: chat!.accountId, chat: chat!.id)),
                child: Semantics(
                  button: chat != null,
                  label: tr('Chat info', 'Информация о чате'),
                  child: Row(
                    children: [
                      MediaAvatar(
                        account: chat?.accountId ?? '',
                        photo: chat?.avatarUrl ?? '',
                        label: chat == null ? '' : chatTitle(chat!),
                        icon: chat?.saved == true
                            ? Icons.bookmark_outline
                            : chat?.kind == pb.ChatKind.CHAT_KIND_CHANNEL
                            ? Icons.campaign_outlined
                            : null,
                        color: providerColor(provider),
                        size: 36,
                        online: online,
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(chat == null ? '' : chatTitle(chat!), style: HmText.bodyStrong, maxLines: 1, overflow: TextOverflow.ellipsis),
                            const SizedBox(height: 2),
                            typing != null
                                ? Text(tr('$typing is typing…', '$typing печатает…'), style: HmText.footer.copyWith(color: HmColors.accent))
                                : Text.rich(
                                    TextSpan(
                                      children: [
                                        if (online)
                                          TextSpan(
                                            text: where.isEmpty ? tr('online', 'в сети') : '${tr('online', 'в сети')} · ',
                                            style: HmText.footer.copyWith(color: HmColors.ok),
                                          ),
                                        TextSpan(text: where),
                                      ],
                                    ),
                                    style: HmText.footer,
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
          if (chat != null && chat!.kind == pb.ChatKind.CHAT_KIND_DIRECT && (account?.capabilities ?? const []).contains(pb.Capability.CAPABILITY_CALLS)) ...[
            HmIconButton(
              icon: Icons.call_outlined,
              label: tr('Call', 'Позвонить'),
              onPressed: () => context.read<CallCubit>().start(chat!.accountId, chat!.id),
            ),
            HmIconButton(
              icon: Icons.videocam_outlined,
              label: tr('Video call', 'Видеозвонок'),
              onPressed: () => context.read<CallCubit>().start(chat!.accountId, chat!.id, video: true),
            ),
          ],
          if (chat != null && chatActionItems(chat!, account).isNotEmpty)
            Builder(
              builder: (context) => HmIconButton(
                icon: Icons.more_vert,
                label: tr('Chat actions', 'Действия с чатом'),
                onPressed: () => showChatMenu(context, chat!, anchor: context),
              ),
            ),
        ],
      ),
    );
  }
}

/// The "older messages" slot at the top of the list.
class _Older {
  const _Older();
}

/// The "unread messages" line above the first message unread on opening.
class _UnreadLine {
  const _UnreadLine();
}

/// Tells the chat which messages are really on screen, and opens the chat
/// at its first unread message. A message counts as seen when at least
/// 40 px of it (or all of a smaller one) are inside the list's viewport
/// after a frame: built-but-offscreen items (the list's cache area), a
/// list not laid out yet, or the moment before the jump to the first
/// unread one read nothing.
class _SeenTracker {
  _SeenTracker(this._cubit, this._scroll);

  final ChatCubit _cubit;
  final ScrollController _scroll;
  final _probes = <_SeenProbeState>{};
  final anchorKey = GlobalKey();
  final viewportKey = GlobalKey();
  bool _scheduled = false;
  bool _disposed = false;

  /// The first unread message the list was moved to (null: not yet).
  String? _placedAt;

  /// The user scrolled: the list stays where they left it.
  bool userScrolled = false;

  /// The list still has to be moved to the first unread message: until then
  /// it builds every item, so the line can be found.
  bool get placing {
    final f = _cubit.state.firstUnread;
    return f != null && f != _placedAt && !userScrolled && _cubit.state.byId(f) != null;
  }

  void schedule() {
    if (_scheduled || _disposed) return;
    _scheduled = true;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _scheduled = false;
      if (!_disposed) _afterFrame();
    });
    WidgetsBinding.instance.ensureVisualUpdate();
  }

  void _afterFrame() {
    if (!_scroll.hasClients) return;
    if (placing) {
      final ctx = anchorKey.currentContext;
      final ro = ctx?.findRenderObject();
      if (ro == null || !ro.attached) return; // the next build brings it
      // The line at the top of the view, as far as the list goes.
      final p = _scroll.position;
      final to = RenderAbstractViewport.of(ro).getOffsetToReveal(ro, 1.0).offset;
      _placedAt = _cubit.state.firstUnread;
      p.jumpTo(to.clamp(p.minScrollExtent, p.maxScrollExtent));
      schedule(); // look at what is on screen after the jump
      return;
    }
    final vp = viewportKey.currentContext?.findRenderObject() as RenderBox?;
    if (vp == null || !vp.attached || !vp.hasSize) return;
    final view = vp.localToGlobal(Offset.zero) & vp.size;
    pb.Message? newest;
    for (final probe in _probes) {
      final box = probe.context.findRenderObject() as RenderBox?;
      if (box == null || !box.attached || !box.hasSize) continue;
      final r = (box.localToGlobal(Offset.zero) & box.size).intersect(view);
      if (r.height <= 0 || r.width <= 0) continue;
      if (r.height < (box.size.height < 40 ? box.size.height : 40) - 0.5) continue;
      final m = probe.widget.message;
      if (newest == null || m.timeMs > newest.timeMs) newest = m;
    }
    if (newest != null) _cubit.seen(newest);
  }

  void dispose() => _disposed = true;
}

/// Wraps a message so that [_SeenTracker] can tell whether it is on screen.
class _SeenProbe extends StatefulWidget {
  const _SeenProbe({super.key, required this.tracker, required this.message, required this.child});

  final _SeenTracker tracker;
  final pb.Message message;
  final Widget child;

  @override
  State<_SeenProbe> createState() => _SeenProbeState();
}

class _SeenProbeState extends State<_SeenProbe> {
  @override
  void initState() {
    super.initState();
    widget.tracker._probes.add(this);
    widget.tracker.schedule();
  }

  @override
  void didUpdateWidget(_SeenProbe old) {
    super.didUpdateWidget(old);
    if (old.message.id != widget.message.id) widget.tracker.schedule();
  }

  @override
  void dispose() {
    widget.tracker._probes.remove(this);
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => widget.child;
}

class _Messages extends StatelessWidget {
  const _Messages({
    required this.scroll,
    required this.seen,
    required this.direct,
    required this.canReact,
    required this.canThread,
    this.account,
    this.openThread,
  });

  final ScrollController scroll;
  final _SeenTracker seen;
  final bool direct;
  final bool canReact;

  /// For editing: the account's capabilities (null: never; a read-only chat).
  final pb.Account? account;
  final bool canThread;
  final ThreadKey? openThread;

  @override
  Widget build(BuildContext context) {
    final s = context.watch<ChatCubit>().state;
    final cubit = context.read<ChatCubit>();
    if (s.loading) return const HmLoading();
    if (s.messages.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: s.error != null
              ? HmBanner(
                  s.error!,
                  tone: HmTone.error,
                  action: HmButton(tr('Retry', 'Повторить'), small: true, onPressed: cubit.load),
                )
              : HmEmptyState(tr('No messages yet', 'Сообщений пока нет'), icon: Icons.chat_bubble_outline),
        ),
      );
    }
    // An album is one item; newest first, as the list is reversed; a day
    // note above each day.
    final groups = groupAlbums(s.messages);
    final msgs = [for (final g in groups) g.message];
    bool unread(int i) => s.firstUnread != null && groups[i].has(s.firstUnread!);
    final entries = <Object>[];
    for (var i = msgs.length - 1; i >= 0; i--) {
      entries.add(i);
      if (unread(i) && i > 0) entries.add(const _UnreadLine());
      final day = fromMs(msgs[i].timeMs);
      if (i == 0 || !sameDay(day, fromMs(msgs[i - 1].timeMs))) entries.add(formatDay(day));
    }
    if (s.more || s.older) entries.add(const _Older());
    final hm = Hm.of(context);
    final placing = seen.placing;
    if (placing) seen.schedule();
    final list = NotificationListener<UserScrollNotification>(
      onNotification: (_) {
        seen.userScrolled = true;
        return false;
      },
      child: Scrollbar(
        key: seen.viewportKey,
        controller: scroll,
        child: ListView.builder(
          controller: scroll,
          reverse: true,
          // Until the list is at the first unread message, every item is built
          // so that the line can be found (one page: 50 messages).
          cacheExtent: placing ? 100000 : null,
          padding: EdgeInsets.symmetric(horizontal: hm.mobile ? 12 : 24, vertical: 12),
          itemCount: entries.length,
          itemBuilder: (context, n) {
            final e = entries[n];
            if (e is String) return HmServiceNote(e);
            if (e is _UnreadLine) return KeyedSubtree(key: seen.anchorKey, child: HmServiceNote(tr('Unread messages', 'Непрочитанные сообщения')));
            if (e is _Older) {
              return Padding(
                padding: const EdgeInsets.symmetric(vertical: 12),
                child: Center(
                  child: s.older
                      ? const SizedBox.square(dimension: 20, child: HmSpinner())
                      : HmButton(tr('Load older', 'Загрузить ранние'), small: true, onPressed: cubit.loadOlder),
                ),
              );
            }
            final i = e as int;
            final m = msgs[i];
            final prev = i > 0 ? msgs[i - 1] : null;
            final first = prev == null || prev.service || prev.sender.id != m.sender.id || !sameDay(fromMs(prev.timeMs), fromMs(m.timeMs));
            final parts = groups[i].parts;
            final tile = MessageTile(
              message: m,
              quoted: m.replyToId.isEmpty ? null : s.byId(m.replyToId),
              showSender: !direct && !m.outgoing && first,
              gap: first,
              onReply: () => cubit.reply(m),
              onEdit: canEditMessage(account, m) ? () => cubit.edit(m) : null,
              onDelete: canDeleteMessage(account, m)
                  ? () => deleteMessages(context, account, [
                      for (final x in parts)
                        if (canDeleteMessage(account, x)) x,
                    ])
                  : null,
              onQuote: (q) => cubit.reply(m, quote: q),
              onForward: () => forwardMessages(context, parts),
              parts: parts,
              onSelect: () => cubit.toggleSelected(m.id),
              selecting: s.selected.isNotEmpty,
              selected: s.selected.contains(m.id),
              onReact: canReact ? (e) => _react(context, cubit, m, e) : null,
              onThread: canThread && m.hasThread() && m.thread.id.isNotEmpty
                  ? () => showThread(context, (account: m.accountId, chat: m.chatId, thread: m.thread.id), m)
                  : null,
              threadOpen: openThread?.thread == m.thread.id && openThread?.chat == m.chatId,
            );
            // The first unread message is the top one when it starts the list.
            final top = i == 0 && unread(i);
            final body = top ? KeyedSubtree(key: seen.anchorKey, child: tile) : tile;
            // Seeing an album reads all of it.
            final newest = groups[i].newest;
            return m.outgoing ? body : _SeenProbe(key: ValueKey(newest.id), tracker: seen, message: newest, child: body);
          },
        ),
      ),
    );
    return PhotoGallery(messages: s.messages, child: list);
  }
}

Future<void> _react(BuildContext context, ChatCubit cubit, pb.Message m, String emoji) async {
  try {
    await cubit.react(m, emoji);
  } catch (e) {
    Toasts.error(tr('Reaction not sent', 'Реакция не отправлена'), errorText(e));
  }
}

/// Text in a message is selected with the mouse right in the bubble on
/// desktops; on phones a long press opens the menu, and "Select text"
/// shows the text in a sheet to select from.
final bool _inlineSelect = !(Platform.isAndroid || Platform.isIOS);

/// One message: sender, quote, text, attachments, reactions and the
/// thread under it. Used by the chat and by the thread pane.
class MessageTile extends StatefulWidget {
  const MessageTile({
    super.key,
    required this.message,
    required this.quoted,
    required this.showSender,
    required this.gap,
    this.onReply,
    this.onEdit,
    this.onDelete,
    this.onQuote,
    this.onReact,
    this.onThread,
    this.onForward,
    this.forwardNative = true,
    this.parts,
    this.onSelect,
    this.selecting = false,
    this.selected = false,
    this.threadOpen = false,
  });

  final pb.Message message;
  final pb.Message? quoted;
  final bool showSender;
  final bool gap; // a new sender: a little more room above
  final VoidCallback? onReply;

  /// Changes the text of this message (null: it cannot be edited here).
  final VoidCallback? onEdit;

  /// Deletes this message (null: it cannot be deleted here).
  final VoidCallback? onDelete;

  /// Replies quoting a fragment of the text (null: no quoting here).
  final ValueChanged<String>? onQuote;

  /// Toggles a reaction (null: reactions are shown only).
  final ValueChanged<String>? onReact;

  /// Opens the thread of this message (null: no threads here).
  final VoidCallback? onThread;

  /// Forwards the whole message (an album: [parts]); null: no forwarding.
  final VoidCallback? onForward;

  /// The messenger may forward it itself (it is in its chat's history).
  final bool forwardNative;

  /// An album's parts, oldest first (null: just [message]).
  final List<pb.Message>? parts;

  /// Starts or toggles picking messages (null: no picking here).
  final VoidCallback? onSelect;

  /// Messages are being picked: a tap picks this one.
  final bool selecting;
  final bool selected;

  /// The thread of this message is open next to the chat.
  final bool threadOpen;

  @override
  State<MessageTile> createState() => _MessageTileState();
}

class _MessageTileState extends State<MessageTile> {
  /// The text selected in the bubble now (desktops).
  String _selection = '';

  pb.Message get m => widget.message;

  Future<void> _copy(String text) async {
    await Clipboard.setData(ClipboardData(text: text));
    Toasts.success(tr('Copied', 'Скопировано'));
  }

  /// The bubble: the message menu opened without a pointer sits next to it.
  final _bubbleKey = GlobalKey();

  /// The message menu at [at] (the pointer, global coordinates), or next to
  /// the bubble without one. Add items to [items] in menu order.
  Future<void> _menu(BuildContext context, {String selection = '', Offset? at}) async {
    final replies = m.hasThread() ? m.thread.replies : 0;
    final sel = m.deleted ? '' : selection.trim();
    final links = m.deleted ? const <TextLink>[] : formattedLinks(m.text, m.format).take(3).toList();
    final live = !m.deleted;
    final items = <HmMenuItem<String>>[
      if (sel.isNotEmpty) ...[
        (value: 'copySel', label: tr('Copy selection', 'Копировать выделенное')),
        if (widget.onQuote != null) (value: 'quote', label: tr('Quote', 'Цитировать')),
        (value: 'forwardSel', label: tr('Forward selection', 'Переслать выделенное')),
      ],
      if (live && widget.onReply != null) (value: 'reply', label: tr('Reply', 'Ответить')),
      if (live && widget.onEdit != null) (value: 'edit', label: tr('Edit', 'Редактировать')),
      if (live && widget.onThread != null)
        (value: 'thread', label: replies > 0 ? tr('Open thread', 'Открыть тред') : tr('Reply in thread', 'Ответить в треде')),
      if (live && widget.onReact != null) (value: 'react', label: tr('React', 'Реакция')),
      if (live && m.text.isNotEmpty) (value: 'copy', label: tr('Copy text', 'Копировать текст')),
      for (final (i, l) in links.indexed) (value: 'copyLink$i', label: '${tr('Copy link', 'Копировать ссылку')} · ${_linkLabel(l)}'),
      if (live && m.text.isNotEmpty && !_inlineSelect) (value: 'selectText', label: tr('Select text', 'Выделить текст')),
      if (live && widget.onForward != null) (value: 'forward', label: tr('Forward', 'Переслать')),
      if (widget.onSelect != null) (value: 'select', label: tr('Select', 'Выбрать')),
      if (live && widget.onDelete != null) (value: 'delete', label: tr('Delete', 'Удалить')),
      for (final (i, a) in m.attachments.indexed)
        if (a.ref.isNotEmpty) ...[
          (value: 'save$i', label: '${tr('Save to Downloads', 'Сохранить в «Загрузки»')} · ${attachmentText(a).$1}'),
          if (canAskSavePath) (value: 'saveAs$i', label: '${tr('Save as…', 'Сохранить как…')} · ${attachmentText(a).$1}'),
        ],
    ];
    final v = await showContextMenu<String>(context, title: tr('Message', 'Сообщение'), items: items, position: at, anchor: _bubbleKey.currentContext);
    if (!context.mounted) return;
    switch (v) {
      case 'copySel':
        await _copy(sel);
      case 'quote':
        widget.onQuote!(sel);
      case 'forwardSel':
        await forwardMessages(context, [m], fragment: sel);
      case 'reply':
        widget.onReply!();
      case 'edit':
        widget.onEdit!();
      case 'thread':
        widget.onThread!();
      case 'react':
        final e = await pickReaction(context);
        if (e != null) widget.onReact!(e);
      case 'copy':
        await _copy(m.text);
      case 'selectText':
        await _selectText(context);
      case 'forward':
        widget.onForward!();
      case 'select':
        widget.onSelect!();
      case 'delete':
        widget.onDelete!();
      case final String s when s.startsWith('copyLink'):
        await _copy(links[int.parse(s.substring(8))].uri.toString().replaceFirst('mailto:', ''));
      case final String s when s.startsWith('saveAs'):
        await saveAttachment(context, m.accountId, m.attachments[int.parse(s.substring(6))], ask: true);
      case final String s when s.startsWith('save'):
        await saveAttachment(context, m.accountId, m.attachments[int.parse(s.substring(4))]);
    }
  }

  /// A link as the menu names it: its host, or the address of a mail link.
  static String _linkLabel(TextLink l) => l.uri.scheme == 'mailto' ? l.uri.path : l.uri.host;

  /// Phones: the text in a sheet, to select a part of it and copy, quote
  /// or forward that part (nothing selected: all of it).
  Future<void> _selectText(BuildContext context) async {
    var sel = '';
    String picked() => sel.trim().isEmpty ? m.text : sel;
    final v = await showHmModal<String>(
      context,
      title: tr('Select text', 'Выделить текст'),
      body: (_) => SelectableText(m.text, style: HmText.body.copyWith(height: 1.45), onSelectionChanged: (s, _) => sel = s.textInside(m.text)),
      actions: (ctx) => [
        HmButton(tr('Forward', 'Переслать'), variant: HmVariant.neutral, onPressed: () => Navigator.pop(ctx, 'forward')),
        if (widget.onQuote != null) HmButton(tr('Quote', 'Цитировать'), variant: HmVariant.neutral, onPressed: () => Navigator.pop(ctx, 'quote')),
        HmButton(tr('Copy', 'Копировать'), onPressed: () => Navigator.pop(ctx, 'copy')),
      ],
    );
    if (!context.mounted) return;
    final text = picked();
    switch (v) {
      case 'copy':
        await _copy(text);
      case 'quote':
        widget.onQuote!(text);
      case 'forward':
        await forwardMessages(context, [m], fragment: text == m.text ? null : text);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (m.service) return HmServiceNote(m.text.isNotEmpty ? m.text : messagePreview(m));
    final edited = m.editedMs > 0 && !m.deleted;
    final hm = Hm.of(context);
    final mediaWidth = hm.mobile ? MediaQuery.sizeOf(context).width * 0.7 : 360.0;
    final thread = m.hasThread() && m.thread.replies > 0 && widget.onThread != null ? m.thread : null;
    final quoted = widget.quoted;
    Widget text = HmRichText(m.text, spans: m.format);
    // A link's page preview; a message with pictures or files shows those.
    final preview = !m.deleted && m.attachments.isEmpty && linkPreviewsOn ? previewLink(m.text) : null;
    if (_inlineSelect && !widget.selecting) {
      text = SelectionArea(
        onSelectionChanged: (c) => _selection = c?.plainText ?? '',
        // The message's own menu opens on a right click (below).
        contextMenuBuilder: (_, _) => const SizedBox.shrink(),
        child: text,
      );
    }
    final bubble = HmBubble(
      boxKey: _bubbleKey,
      outgoing: m.outgoing,
      dim: m.deleted,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          if (widget.showSender)
            Padding(
              padding: const EdgeInsets.only(bottom: 2),
              child: Text(senderName(m), style: HmText.jb(11, weight: 600, color: HmColors.accent)),
            ),
          if (m.forwarded && !m.deleted)
            Padding(
              padding: const EdgeInsets.only(bottom: 2),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(Icons.shortcut, size: 13, color: HmColors.muted),
                  const SizedBox(width: 4),
                  Flexible(
                    child: Text(
                      m.forwardedFrom.isNotEmpty ? '${tr('Forwarded from', 'Переслано от')} ${m.forwardedFrom}' : tr('Forwarded', 'Переслано'),
                      style: HmText.jb(11, color: HmColors.muted),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                ],
              ),
            ),
          if (m.replyToId.isNotEmpty)
            HmQuote(
              title: quoted != null ? senderName(quoted) : tr('Reply', 'Ответ'),
              text: m.quote.isNotEmpty
                  ? m.quote.replaceAll(RegExp(r'\s+'), ' ')
                  : (quoted != null ? messagePreview(quoted) : tr('an earlier message', 'более раннее сообщение')),
            ),
          if (m.deleted)
            Text(tr('Message deleted', 'Сообщение удалено'), style: HmText.small.copyWith(fontStyle: FontStyle.italic))
          else ...[
            if (m.attachments.isNotEmpty) MessageAttachments(message: m, maxWidth: mediaWidth),
            if (m.text.isNotEmpty)
              Padding(
                padding: EdgeInsets.only(top: m.attachments.isEmpty ? 0 : 4),
                child: text,
              ),
            if (preview != null) LinkPreviewCard(uri: preview, maxWidth: mediaWidth),
            if (m.reactions.isNotEmpty)
              HmReactions(items: [for (final r in m.reactions) (r.emoji, r.count, r.mine)], onTap: widget.selecting ? null : widget.onReact),
            if (thread != null)
              HmThreadSummary(
                label: '${thread.replies} ${_replies(thread.replies)}',
                detail: thread.lastMs > 0 ? '${tr('last', 'последний')} ${formatChatTime(thread.lastMs)}' : null,
                avatars: [for (final u in thread.users.take(3)) MediaAvatar(account: m.accountId, photo: u.avatarUrl, label: u.name, size: 20)],
                onTap: widget.onThread!,
              ),
          ],
          const SizedBox(height: 2),
          // Full date on hover over the time only: a tooltip on the whole message is
          // placed at its centre, over the content of a tall one.
          Tooltip(
            message: formatStamp(m.timeMs),
            waitDuration: const Duration(milliseconds: 500),
            child: Text(edited ? '${tr('edited', 'изменено')} · ${formatClock(m.timeMs)}' : formatClock(m.timeMs), style: HmText.footer),
          ),
        ],
      ),
    );
    final lit = widget.threadOpen || widget.selected;
    Widget body = lit
        ? DecoratedBox(
            decoration: BoxDecoration(color: HmColors.tint(HmColors.accent, HmColors.subtle)),
            child: bubble,
          )
        : bubble;
    if (widget.selecting) {
      // Picking: the whole row is a check box; nothing inside reacts.
      body = MouseRegion(
        cursor: SystemMouseCursors.click,
        child: GestureDetector(
          behavior: HitTestBehavior.opaque,
          onTap: widget.onSelect,
          child: Row(
            children: [
              Padding(
                padding: const EdgeInsets.only(right: 8),
                child: Icon(
                  widget.selected ? Icons.check_circle : Icons.radio_button_unchecked,
                  size: 20,
                  color: widget.selected ? HmColors.accent : HmColors.muted,
                  semanticLabel: widget.selected ? tr('Selected', 'Выбрано') : tr('Not selected', 'Не выбрано'),
                ),
              ),
              Expanded(child: IgnorePointer(child: body)),
            ],
          ),
        ),
      );
    }
    return Padding(
      padding: EdgeInsets.only(top: widget.gap ? 8 : 2, bottom: 2),
      child: Builder(
        builder: (context) => Listener(
          // Before the text's own handling of the click, which may drop the
          // selection: the menu gets what was selected.
          onPointerDown: (e) {
            if (widget.selecting || e.kind != PointerDeviceKind.mouse || e.buttons & kSecondaryMouseButton == 0) return;
            final sel = _selection;
            _menu(context, selection: sel, at: e.position);
          },
          child: GestureDetector(
            behavior: HitTestBehavior.opaque,
            onLongPress: widget.selecting ? widget.onSelect : null,
            onLongPressStart: widget.selecting ? null : (d) => _menu(context, at: d.globalPosition),
            onDoubleTap: widget.selecting || m.deleted || widget.onReply == null ? null : widget.onReply,
            child: body,
          ),
        ),
      ),
    );
  }
}

/// Instead of the header while messages are picked: how many, copy,
/// forward, cancel (Escape).
class _SelectionBar extends StatelessWidget {
  const _SelectionBar({required this.count});

  final int count;

  /// The picked messages with their albums, oldest first.
  List<MessageGroup> _picked(ChatState s) => [
    for (final g in groupAlbums(s.messages))
      if (s.selected.contains(g.message.id)) g,
  ];

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    final cubit = context.read<ChatCubit>();
    return CallbackShortcuts(
      bindings: {const SingleActivator(LogicalKeyboardKey.escape): cubit.clearSelected},
      child: Focus(
        autofocus: true,
        child: Container(
          height: hm.mobile ? 56 : 64,
          padding: EdgeInsets.symmetric(horizontal: hm.mobile ? 8 : 20),
          decoration: const BoxDecoration(
            color: HmColors.surface,
            border: Border(bottom: BorderSide(color: HmColors.border)),
          ),
          child: Row(
            children: [
              HmIconButton(icon: Icons.close, label: tr('Cancel', 'Отмена'), onPressed: cubit.clearSelected),
              const SizedBox(width: 8),
              Expanded(child: Text(tr('Selected: $count', 'Выбрано: $count'), style: HmText.bodyStrong)),
              HmIconButton(
                icon: Icons.copy_outlined,
                label: tr('Copy', 'Копировать'),
                onPressed: () async {
                  final text = copyText([for (final g in _picked(cubit.state)) g.message], (m) => formatStamp(m.timeMs));
                  await Clipboard.setData(ClipboardData(text: text));
                  Toasts.success(tr('Copied', 'Скопировано'));
                  cubit.clearSelected();
                },
              ),
              Builder(
                builder: (context) => HmIconButton(
                  icon: Icons.shortcut,
                  label: tr('Forward', 'Переслать'),
                  onPressed: () async {
                    final parts = [for (final g in _picked(cubit.state)) ...g.parts];
                    if (await forwardMessages(context, parts)) cubit.clearSelected();
                  },
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

String _replies(int n) {
  if (lang.value != Lang.ru) return n == 1 ? 'reply' : 'replies';
  final d = n % 10, h = n % 100;
  if (d == 1 && h != 11) return 'ответ';
  if (d >= 2 && d <= 4 && (h < 12 || h > 14)) return 'ответа';
  return 'ответов';
}

class _Composer extends StatefulWidget {
  const _Composer({required this.chat});

  final ChatKey chat;

  @override
  State<_Composer> createState() => _ComposerState();
}

class _ComposerState extends State<_Composer> {
  // Unsent text and files per chat, while the app runs.
  static final _drafts = <ChatKey, String>{};
  static final _draftFiles = <ChatKey, List<LocalFile>>{};

  late final _text = TextEditingController(text: _drafts[widget.chat]);
  late final _files = PendingFiles(_draftFiles[widget.chat] ?? const []);
  final _focus = FocusNode();
  late final ChatCubit _cubit = context.read<ChatCubit>();
  bool _sending = false; // files or an edit are on their way

  /// The draft put aside while a message is edited (null: not editing).
  String? _beforeEdit;

  /// People after "@" (null: the messenger has no mentions).
  MentionPicker? _mentions;

  @override
  void initState() {
    super.initState();
    if (canMention(context.read<MessengerCubit>().state.account(widget.chat.account))) {
      final core = context.read<CoreCubit>();
      _mentions = MentionPicker(api: () => core.api, account: widget.chat.account, chat: widget.chat.chat, text: _text);
    }
    _text.addListener(() => setState(() {}));
    _files.addListener(() => setState(() {}));
    // Desktop: type right away in the chat just opened.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted && !Hm.of(context).mobile) _focus.requestFocus();
    });
  }

  @override
  void dispose() {
    final t = _beforeEdit ?? _text.text;
    t.isEmpty ? _drafts.remove(widget.chat) : _drafts[widget.chat] = t;
    _files.isEmpty ? _draftFiles.remove(widget.chat) : _draftFiles[widget.chat] = _files.value;
    _mentions?.dispose();
    _text.dispose();
    _files.dispose();
    _focus.dispose();
    super.dispose();
  }

  Future<void> _send() async {
    final text = _text.text.trim();
    final files = _files.value;
    if (_sending) return;
    if (_cubit.state.editing != null) return _saveEdit(text);
    if (text.isEmpty && files.isEmpty) return;
    if (files.isNotEmpty) return _sendFiles(text, files);
    final mentions = _mentions?.mentionsIn(text) ?? const <pb.Mention>[];
    _text.clear();
    try {
      await _cubit.send(text, mentions: mentions);
      _mentions?.sent();
    } catch (e) {
      if (_text.text.isEmpty) _text.text = text;
      Toasts.error(tr('Not sent', 'Не отправлено'), errorText(e));
    }
  }

  /// The edited text goes in place of the message's; the draft comes back
  /// once it is saved (the listener below).
  Future<void> _saveEdit(String text) async {
    if (text.isEmpty) return;
    setState(() => _sending = true);
    try {
      await _cubit.saveEdit(text);
    } catch (e) {
      Toasts.error(tr('Not saved', 'Не сохранено'), errorText(e));
    } finally {
      if (mounted) setState(() => _sending = false);
    }
  }

  /// Editing started, moved to another message or ended: the composer
  /// holds the message's text meanwhile, the draft before and after.
  void _editingChanged(pb.Message? now) {
    if (now != null) {
      _beforeEdit ??= _text.text;
      final source = editSource(now);
      _text.value = TextEditingValue(
        text: source,
        selection: TextSelection.collapsed(offset: source.length),
      );
      _focus.requestFocus();
    } else {
      _text.text = _beforeEdit ?? '';
      _beforeEdit = null;
    }
  }

  /// Files take a while: they and the caption stay in the composer until
  /// they are sent, so a failure loses nothing.
  Future<void> _sendFiles(String text, List<LocalFile> files) async {
    final chat = widget.chat;
    setState(() => _sending = true);
    try {
      await _cubit.send(text, files: files, mentions: _mentions?.mentionsIn(text) ?? const []);
      _mentions?.sent();
      if (!mounted) {
        // The chat was left meanwhile: its draft is what was sent.
        _draftFiles.remove(chat);
        if (_drafts[chat]?.trim() == text) _drafts.remove(chat);
        return;
      }
      _files.value = const [];
      if (_text.text.trim() == text) _text.clear();
    } catch (e) {
      Toasts.error(tr('Not sent', 'Не отправлено'), errorText(e));
    } finally {
      if (mounted) setState(() => _sending = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    final reply = context.select<ChatCubit, pb.Message?>((c) => c.state.replyTo);
    final editing = context.select<ChatCubit, pb.Message?>((c) => c.state.editing);
    final files = editing == null && context.select<MessengerCubit, bool>((c) => canSendFiles(c.state.account(widget.chat.account)));
    final quote = context.select<ChatCubit, String?>((c) => c.state.quote);
    return MultiBlocListener(
      listeners: [
        BlocListener<ChatCubit, ChatState>(
          listenWhen: (a, b) => b.replyTo != null && (a.replyTo != b.replyTo || a.quote != b.quote),
          listener: (_, _) => _focus.requestFocus(),
        ),
        BlocListener<ChatCubit, ChatState>(listenWhen: (a, b) => a.editing?.id != b.editing?.id, listener: (_, s) => _editingChanged(s.editing)),
      ],
      child: Container(
        padding: EdgeInsets.fromLTRB(hm.mobile ? 12 : 24, 10, hm.mobile ? 8 : 24, 10),
        decoration: const BoxDecoration(
          color: HmColors.surface,
          border: Border(top: BorderSide(color: HmColors.border)),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (reply != null)
              HmQuote(
                title: tr('Reply to ${senderName(reply)}', 'Ответ: ${senderName(reply)}'),
                text: quote != null ? quote.replaceAll(RegExp(r'\s+'), ' ') : messagePreview(reply),
                trailing: HmIconButton(icon: Icons.close, label: tr('Cancel reply', 'Отменить ответ'), onPressed: () => _cubit.reply(null)),
              ),
            if (editing != null)
              HmQuote(
                title: tr('Editing', 'Редактирование'),
                text: messagePreview(editing),
                trailing: HmIconButton(icon: Icons.close, label: tr('Cancel editing', 'Отменить редактирование'), onPressed: () => _cubit.edit(null)),
              ),
            if (files) PendingFilesStrip(files: _files, enabled: !_sending),
            if (_mentions != null) MentionList(picker: _mentions!),
            Row(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                if (files) ...[AttachButton(files: _files, enabled: !_sending), const SizedBox(width: 8)],
                Expanded(
                  child: CallbackShortcuts(
                    bindings: {
                      const SingleActivator(LogicalKeyboardKey.escape): () {
                        _cubit.reply(null);
                        _cubit.edit(null);
                      },
                    },
                    child: HmMessageField(
                      controller: _text,
                      focusNode: _focus,
                      onSend: _send,
                      onKey: _mentions?.onKey,
                      onPaste: files ? () => attachFromClipboard(_files) : null,
                      onInsertContent: files ? (c) => attachInserted(_files, c) : null,
                      placeholder: hm.mobile
                          ? tr('Message', 'Сообщение')
                          : tr('Message — Enter to send, Shift+Enter for a new line', 'Сообщение — Enter отправить, Shift+Enter новая строка'),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                _sending
                    ? SizedBox(
                        width: hm.toolbar,
                        height: hm.toolbar,
                        child: const Center(child: SizedBox(width: 16, height: 16, child: HmSpinner())),
                      )
                    : HmIconButton(
                        icon: editing != null ? Icons.check : Icons.send,
                        label: editing != null ? tr('Save', 'Сохранить') : tr('Send', 'Отправить'),
                        color: HmColors.accent,
                        onPressed: _text.text.trim().isEmpty && (editing != null || _files.isEmpty) ? null : _send,
                      ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
