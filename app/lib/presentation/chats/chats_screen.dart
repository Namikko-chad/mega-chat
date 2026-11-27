import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/format.dart';
import '../../core/settings/app_settings.dart';
import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/layout.dart';
import '../../core/theme/typography.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/parley.pbgrpc.dart' as pb;
import '../accounts/account_rail.dart';
import '../accounts/login_wizard.dart';
import '../accounts/providers.dart';
import '../common/hm.dart';
import '../core/messenger_cubit.dart';
import 'chat_actions.dart';
import 'chat_view.dart';
import 'media.dart';
import 'messages.dart';
import 'thread_view.dart';

/// The unified inbox. Desktop: accounts rail, chat list and the open chat
/// side by side. Phone: an account strip over the list; a chat is pushed.
class ChatsScreen extends StatelessWidget {
  const ChatsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final m = context.watch<MessengerCubit>().state;
    if (m.loading && m.accounts.isEmpty) return const HmLoading();
    if (chatAccounts(m).isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: m.error != null
              ? HmBanner(m.error!, tone: HmTone.error)
              : HmEmptyState(
                  tr(
                    'No accounts yet. Add VK Teams, Slack or Telegram to see your chats here.',
                    'Аккаунтов пока нет. Добавьте VK Teams, Slack или Telegram, чтобы увидеть здесь свои чаты.',
                  ),
                  icon: Icons.forum_outlined,
                  action: HmButton(
                    tr('Add account', 'Добавить аккаунт'),
                    variant: HmVariant.primary,
                    icon: Icons.add,
                    onPressed: () => showLoginWizard(context),
                  ),
                ),
        ),
      );
    }
    if (Hm.of(context).mobile) {
      return const Column(
        children: [
          _AccountStrip(),
          Expanded(child: _ChatList()),
        ],
      );
    }
    return LayoutBuilder(
      builder: (context, box) {
        // A wide window shows the accounts with names; a narrower one (half
        // a screen) only their icons, like Telegram's folders.
        final w = box.maxWidth;
        final fullRail = w >= 1280;
        final railW = fullRail ? 240.0 : 76.0;
        final listW = w >= 1100 ? 360.0 : 300.0;
        final thread = m.thread;
        final hasChat = m.open != null && m.chats.containsKey(m.open);
        // The thread goes beside the chat when both fit, else over it.
        final threadBeside = thread != null && w - railW - listW >= 760;
        Widget main;
        if (!hasChat) {
          main = HmEmptyState(tr('Choose a chat', 'Выберите чат'), icon: Icons.chat_bubble_outline);
        } else if (thread != null && !threadBeside) {
          main = ThreadPane(thread: thread, onBack: () => context.read<MessengerCubit>().openThread(null));
        } else {
          main = ChatPane(chat: m.open!);
        }
        return Row(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            SizedBox(width: railW, child: fullRail ? const _AccountRail() : const _CompactRail()),
            SizedBox(width: listW, child: const _ChatList()),
            Expanded(child: main),
            if (hasChat && threadBeside)
              Container(
                width: 380,
                decoration: const BoxDecoration(
                  border: Border(left: BorderSide(color: HmColors.border)),
                ),
                child: ThreadPane(thread: thread, onClose: () => context.read<MessengerCubit>().openThread(null)),
              ),
          ],
        );
      },
    );
  }
}

/// Opens a chat: on the right on desktops, as a new screen on phones.
void openChat(BuildContext context, ChatKey key) {
  final messenger = context.read<MessengerCubit>();
  messenger.openChat(key);
  if (!Hm.of(context).mobile) return;
  Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => ChatPage(chat: key))).whenComplete(() {
    if (messenger.state.open == key) messenger.openChat(null);
  });
}

class _AccountRail extends StatelessWidget {
  const _AccountRail();

  @override
  Widget build(BuildContext context) {
    final m = context.watch<MessengerCubit>().state;
    final messenger = context.read<MessengerCubit>();
    final accounts = chatAccounts(m);
    return AccountRail(
      child: ReorderableAccounts(
        accounts: accounts,
        first: HmListItem(
          leading: const HmAvatar(label: '', icon: Icons.all_inbox_outlined),
          title: Text(tr('All chats', 'Все чаты')),
          subtitle: Text(tr('${accounts.length} accounts', 'аккаунтов: ${accounts.length}')),
          trailing: HmCount(m.unread),
          selected: m.filter.isEmpty,
          onTap: () => messenger.setFilter(''),
        ),
        item: (a) => AccountTile(account: a, count: a.unread, selected: m.filter == a.id, onTap: () => messenger.setFilter(a.id)),
      ),
    );
  }
}

/// The accounts with chats (mail and calendar accounts have pages of their own).
List<pb.Account> chatAccounts(MessengerState m) => accountsOffering(m, pb.Capability.CAPABILITY_MESSAGES);

/// A narrow window: the accounts as icons with their unread counts.
class _CompactRail extends StatelessWidget {
  const _CompactRail();

  @override
  Widget build(BuildContext context) {
    final m = context.watch<MessengerCubit>().state;
    final messenger = context.read<MessengerCubit>();
    return Container(
      decoration: const BoxDecoration(
        color: HmColors.surface,
        border: Border(right: BorderSide(color: HmColors.border)),
      ),
      child: Column(
        children: [
          Expanded(
            child: ReorderableAccounts(
              accounts: chatAccounts(m),
              padding: const EdgeInsets.symmetric(vertical: 8),
              first: _RailButton(
                icon: const HmAvatar(label: '', icon: Icons.all_inbox_outlined, size: 36),
                label: tr('All', 'Все'),
                tooltip: tr('All chats', 'Все чаты'),
                count: m.unread,
                selected: m.filter.isEmpty,
                onTap: () => messenger.setFilter(''),
              ),
              item: (a) => _RailButton(
                icon: AccountAvatar(account: a),
                dot: a.enabled ? HmDot(ok: isOnline(a), label: accountStatus(a).$1) : null,
                label: a.name.isEmpty ? a.login : a.name,
                tooltip: '${a.name.isEmpty ? a.login : a.name} · ${accountStatus(a).$1}',
                count: a.enabled ? a.unread : 0,
                selected: m.filter == a.id,
                onTap: () => messenger.setFilter(a.id),
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.only(bottom: 12),
            child: HmIconButton(icon: Icons.add, label: tr('Add account', 'Добавить аккаунт'), onPressed: () => showLoginWizard(context)),
          ),
        ],
      ),
    );
  }
}

class _RailButton extends StatefulWidget {
  const _RailButton({
    required this.icon,
    required this.label,
    required this.tooltip,
    required this.count,
    required this.selected,
    required this.onTap,
    this.dot,
  });

  final Widget icon;

  /// Online or not, at the bottom left (the messenger's mark takes the right).
  final Widget? dot;
  final String label;
  final String tooltip;
  final int count;
  final bool selected;
  final VoidCallback onTap;

  @override
  State<_RailButton> createState() => _RailButtonState();
}

class _RailButtonState extends State<_RailButton> {
  bool _hot = false;

  @override
  Widget build(BuildContext context) {
    final bg = widget.selected ? HmColors.tint(HmColors.accent, 0.08) : (_hot ? HmColors.tint(HmColors.text, HmColors.subtle) : Colors.transparent);
    return Tooltip(
      message: widget.tooltip,
      waitDuration: const Duration(milliseconds: 600),
      child: MouseRegion(
        cursor: SystemMouseCursors.click,
        onEnter: (_) => setState(() => _hot = true),
        onExit: (_) => setState(() => _hot = false),
        child: GestureDetector(
          behavior: HitTestBehavior.opaque,
          onTap: widget.onTap,
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 150),
            padding: const EdgeInsets.symmetric(vertical: 8),
            decoration: BoxDecoration(
              color: bg,
              border: Border(left: BorderSide(color: widget.selected ? HmColors.accent : Colors.transparent, width: 2)),
            ),
            child: Column(
              children: [
                Stack(
                  clipBehavior: Clip.none,
                  children: [
                    widget.icon,
                    if (widget.dot != null) Positioned(left: -2, bottom: -2, child: widget.dot!),
                    if (widget.count > 0) Positioned(right: -10, top: -6, child: HmCount(widget.count)),
                  ],
                ),
                const SizedBox(height: 4),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 4),
                  child: Text(
                    widget.label,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    textAlign: TextAlign.center,
                    style: HmText.footer.copyWith(color: widget.selected ? HmColors.accent : HmColors.muted),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

/// Phone: the accounts as a strip of chips above the list.
class _AccountStrip extends StatelessWidget {
  const _AccountStrip();

  @override
  Widget build(BuildContext context) {
    final m = context.watch<MessengerCubit>().state;
    final messenger = context.read<MessengerCubit>();
    return Container(
      height: 56,
      decoration: const BoxDecoration(
        color: HmColors.surface,
        border: Border(bottom: BorderSide(color: HmColors.border)),
      ),
      child: ReorderableAccounts(
        accounts: chatAccounts(m),
        axis: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        first: HmChip(label: tr('All', 'Все'), active: m.filter.isEmpty, count: m.unread, onTap: () => messenger.setFilter('')),
        item: (a) => Padding(
          padding: const EdgeInsets.only(left: 8),
          child: HmChip(
            label: a.name.isEmpty ? a.login : a.name,
            leading: Icon(providerIcon(a.provider), size: 14, color: providerColor(a.provider)),
            active: m.filter == a.id,
            count: a.enabled ? a.unread : 0,
            onTap: () => messenger.setFilter(a.id),
          ),
        ),
      ),
    );
  }
}

class _ChatList extends StatefulWidget {
  const _ChatList();

  @override
  State<_ChatList> createState() => _ChatListState();
}

class _ChatListState extends State<_ChatList> {
  final _search = TextEditingController();
  String _query = '';

  // Folded sections, while the app runs.
  static final _folded = <bool>{}; // isDirect of the folded sections

  @override
  void initState() {
    super.initState();
    AppSettings.instance.addListener(_rebuild);
  }

  void _rebuild() => setState(() {});

  @override
  void dispose() {
    AppSettings.instance.removeListener(_rebuild);
    _search.dispose();
    super.dispose();
  }

  Future<void> _refresh() async {
    try {
      await context.read<MessengerCubit>().refresh();
    } catch (e) {
      Toasts.error(tr('Could not refresh', 'Не удалось обновить'), errorText(e));
    }
  }

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    final m = context.watch<MessengerCubit>().state;
    final q = _query.toLowerCase();
    final chats = q.isEmpty ? m.shown : m.shown.where((c) => chatTitle(c).toLowerCase().contains(q)).toList();
    // The provider marker tells accounts apart only when each provider has one.
    final perProvider = <String, int>{};
    for (final a in chatAccounts(m)) {
      perProvider[a.provider] = (perProvider[a.provider] ?? 0) + 1;
    }
    final filtered = m.account(m.filter);
    Widget tile(pb.Chat c) {
      final key = chatKey(c);
      final a = m.account(c.accountId);
      return _ChatTile(
        chat: c,
        account: a,
        showAccount: m.filter.isEmpty,
        accountName: (perProvider[a?.provider] ?? 0) > 1 && m.filter.isEmpty ? (a?.name ?? '') : '',
        typing: m.typing[key],
        selected: !hm.mobile && m.open == key,
        onTap: () => openChat(context, key),
      );
    }

    // Grouped: channels and groups, then direct messages. A search
    // shows one flat list.
    final entries = <Object>[];
    if (AppSettings.instance.groupChats && q.isEmpty) {
      for (final direct in [false, true]) {
        final part = chats.where((c) => isDirect(c) == direct).toList();
        if (part.isEmpty) continue;
        entries.add(direct);
        if (!_folded.contains(direct)) entries.addAll(part);
      }
    } else {
      entries.addAll(chats);
    }
    int unreadIn(bool direct) => chats.where((c) => isDirect(c) == direct && !c.muted && c.unread > 0).length;
    Widget list = chats.isEmpty
        ? ListView(
            physics: const AlwaysScrollableScrollPhysics(),
            children: [
              const SizedBox(height: 40),
              HmEmptyState(
                q.isNotEmpty
                    ? tr('Nothing found', 'Ничего не найдено')
                    : m.unreadOnly
                    ? tr('No unread chats', 'Непрочитанных чатов нет')
                    : m.view == archiveView
                    ? tr('The archive is empty', 'Архив пуст')
                    : m.folder != null
                    ? tr('No chats in this folder', 'В этой папке нет чатов')
                    : tr('No chats yet', 'Чатов пока нет'),
                icon: m.unreadOnly && q.isEmpty
                    ? Icons.mark_chat_read_outlined
                    : m.view == archiveView
                    ? Icons.archive_outlined
                    : Icons.chat_bubble_outline,
              ),
            ],
          )
        : ListView.builder(
            physics: const AlwaysScrollableScrollPhysics(),
            itemCount: entries.length,
            itemBuilder: (context, i) {
              final e = entries[i];
              if (e is bool) {
                return HmListHeader(
                  e ? tr('Direct messages', 'Личные сообщения') : tr('Channels and groups', 'Каналы и группы'),
                  count: unreadIn(e),
                  collapsed: _folded.contains(e),
                  onTap: () => setState(() => _folded.contains(e) ? _folded.remove(e) : _folded.add(e)),
                );
              }
              return tile(e as pb.Chat);
            },
          );
    if (hm.mobile) {
      list = RefreshIndicator(onRefresh: _refresh, color: HmColors.accent, backgroundColor: HmColors.surface, child: list);
    }
    return Container(
      decoration: BoxDecoration(
        color: hm.mobile ? null : HmColors.bg,
        border: hm.mobile ? null : const Border(right: BorderSide(color: HmColors.border)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          if (!hm.mobile)
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 12, 12),
              child: Row(
                children: [
                  Expanded(
                    child: Text(filtered == null ? tr('All chats', 'Все чаты') : filtered.name, style: HmText.modalTitle, overflow: TextOverflow.ellipsis),
                  ),
                  HmIconButton(icon: Icons.refresh, label: tr('Refresh', 'Обновить'), onPressed: _refresh),
                ],
              ),
            ),
          Padding(
            padding: EdgeInsets.fromLTRB(16, hm.mobile ? 12 : 0, hm.mobile ? 8 : 12, 12),
            child: Row(
              children: [
                Expanded(
                  child: HmSearchField(
                    controller: _search,
                    placeholder: tr('Search chats', 'Поиск чатов'),
                    onChanged: (v) => setState(() => _query = v.trim()),
                  ),
                ),
                const SizedBox(width: 4),
                HmIconButton(
                  icon: m.unreadOnly ? Icons.mark_chat_unread : Icons.mark_chat_unread_outlined,
                  label: m.unreadOnly ? tr('Showing unread only — show all', 'Только непрочитанные — показать все') : tr('Unread only', 'Только непрочитанные'),
                  color: m.unreadOnly ? HmColors.accent : null,
                  onPressed: () => context.read<MessengerCubit>().setUnreadOnly(!m.unreadOnly),
                ),
              ],
            ),
          ),
          const _ViewStrip(),
          const Divider(),
          Expanded(child: list),
        ],
      ),
    );
  }
}

class _ChatTile extends StatelessWidget {
  const _ChatTile({
    required this.chat,
    required this.account,
    required this.showAccount,
    required this.accountName,
    required this.typing,
    required this.selected,
    required this.onTap,
  });

  final pb.Chat chat;
  final pb.Account? account;
  final bool showAccount;
  final String accountName;
  final String? typing;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final c = chat;
    final provider = account?.provider ?? '';
    final last = c.hasLastMessage() ? c.lastMessage : null;
    final direct = c.kind == pb.ChatKind.CHAT_KIND_DIRECT;
    final Widget subtitle = typing != null
        ? Text(tr('$typing is typing…', '$typing печатает…'), style: HmText.small.copyWith(color: HmColors.accent))
        : last == null
        ? const Text('')
        : Text.rich(
            TextSpan(
              children: [
                if (!last.service && (last.outgoing || !direct))
                  TextSpan(
                    text: '${senderName(last)}: ',
                    style: HmText.small.copyWith(color: HmColors.text),
                  ),
                TextSpan(text: messagePreview(last)),
              ],
            ),
          );
    return HmListItem(
      leading: MediaAvatar(
        account: c.accountId,
        photo: c.avatarUrl,
        label: chatTitle(c),
        icon: c.saved
            ? Icons.bookmark_outline
            : c.kind == pb.ChatKind.CHAT_KIND_CHANNEL
            ? Icons.campaign_outlined
            : null,
        color: providerColor(provider),
        size: Hm.of(context).mobile ? 44 : 40,
        corner: showAccount ? providerMarker(provider, tooltip: account?.name) : null,
        online: direct && c.online,
      ),
      title: Row(
        children: [
          Flexible(child: Text(chatTitle(c), overflow: TextOverflow.ellipsis)),
          if (c.muted) ...[const SizedBox(width: 4), const Icon(Icons.notifications_off_outlined, size: 12, color: HmColors.muted)],
          if (c.pinned) ...[const SizedBox(width: 4), const Icon(Icons.push_pin_outlined, size: 12, color: HmColors.muted)],
        ],
      ),
      meta: Text(accountName.isEmpty ? formatChatTime(c.updatedMs) : '$accountName · ${formatChatTime(c.updatedMs)}'),
      subtitle: subtitle,
      trailing: HmCount(c.unread, muted: c.muted),
      selected: selected,
      onTap: onTap,
      // Right click or long press: pin, mute, archive, leave.
      onLongPress: () => showChatMenu(context, c, anchor: context),
    );
  }
}

/// All chats, the chat folders (Telegram) and the archive, over the list.
/// Hidden while there are neither folders nor archived chats.
class _ViewStrip extends StatelessWidget {
  const _ViewStrip();

  @override
  Widget build(BuildContext context) {
    final m = context.watch<MessengerCubit>().state;
    final messenger = context.read<MessengerCubit>();
    final folders = m.shownFolders;
    if (folders.isEmpty && m.archived.isEmpty && m.view != archiveView) return const SizedBox.shrink();
    // Folders of several accounts are told apart by the account's name.
    final owners = {for (final f in folders) f.accountId};
    String label(pb.ChatFolder f) {
      final name = f.emoji.isEmpty ? f.title : '${f.emoji} ${f.title}';
      if (owners.length < 2) return name;
      final a = m.account(f.accountId);
      return '$name · ${a?.name.isNotEmpty == true ? a!.name : a?.login ?? ''}';
    }

    int unread(bool Function(pb.Chat) where) =>
        m.chats.values.where((c) => (m.filter.isEmpty || c.accountId == m.filter) && !c.muted && c.unread > 0 && where(c)).length;
    final current = m.view == archiveView || m.folder != null ? m.view : '';
    return SizedBox(
      height: 44,
      child: ListView(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.fromLTRB(16, 0, 16, 10),
        children: [
          HmChip(label: tr('All chats', 'Все чаты'), active: current.isEmpty, count: unread((c) => !c.archived), onTap: () => messenger.setView('')),
          for (final f in folders) ...[
            const SizedBox(width: 8),
            HmChip(label: label(f), active: current == folderView(f), count: unread((c) => inFolder(f, c)), onTap: () => messenger.setView(folderView(f))),
          ],
          if (m.archived.isNotEmpty || current == archiveView) ...[
            const SizedBox(width: 8),
            HmChip(
              label: tr('Archive', 'Архив'),
              leading: const Icon(Icons.archive_outlined, size: 14, color: HmColors.muted),
              active: current == archiveView,
              count: unread((c) => c.archived),
              onTap: () => messenger.setView(archiveView),
            ),
          ],
        ],
      ),
    );
  }
}

/// Phone: an open chat as its own screen, over the tab bar.
class ChatPage extends StatelessWidget {
  const ChatPage({super.key, required this.chat});

  final ChatKey chat;

  @override
  Widget build(BuildContext context) => BlocListener<MessengerCubit, MessengerState>(
    // The chat or its account was removed meanwhile.
    listenWhen: (a, b) => b.open == null && a.open == chat,
    listener: (context, _) {
      if (ModalRoute.of(context)?.isCurrent ?? false) Navigator.of(context).pop();
    },
    child: Scaffold(
      backgroundColor: HmColors.bg,
      body: SafeArea(
        child: ChatPane(chat: chat, onBack: () => Navigator.of(context).maybePop()),
      ),
    ),
  );
}
