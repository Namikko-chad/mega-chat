import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/format.dart';
import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/layout.dart';
import '../../core/theme/typography.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../accounts/account_rail.dart';
import '../accounts/login_wizard.dart';
import '../accounts/providers.dart';
import '../chats/media.dart';
import '../common/hm.dart';
import '../core/messenger_cubit.dart';
import 'compose.dart';
import 'mail_cubit.dart';

/// The folder roles shown for all accounts at once, in this order.
const unifiedRoles = [
  pb.FolderRole.FOLDER_ROLE_INBOX,
  pb.FolderRole.FOLDER_ROLE_SENT,
  pb.FolderRole.FOLDER_ROLE_DRAFTS,
  pb.FolderRole.FOLDER_ROLE_ARCHIVE,
  pb.FolderRole.FOLDER_ROLE_SPAM,
  pb.FolderRole.FOLDER_ROLE_TRASH,
];

String roleName(pb.FolderRole r) => switch (r) {
  pb.FolderRole.FOLDER_ROLE_INBOX => tr('Inbox', 'Входящие'),
  pb.FolderRole.FOLDER_ROLE_SENT => tr('Sent', 'Отправленные'),
  pb.FolderRole.FOLDER_ROLE_DRAFTS => tr('Drafts', 'Черновики'),
  pb.FolderRole.FOLDER_ROLE_ARCHIVE => tr('Archive', 'Архив'),
  pb.FolderRole.FOLDER_ROLE_SPAM => tr('Spam', 'Спам'),
  pb.FolderRole.FOLDER_ROLE_TRASH => tr('Trash', 'Корзина'),
  _ => tr('Folder', 'Папка'),
};

IconData roleIcon(pb.FolderRole r) => switch (r) {
  pb.FolderRole.FOLDER_ROLE_INBOX => Icons.inbox_outlined,
  pb.FolderRole.FOLDER_ROLE_SENT => Icons.send_outlined,
  pb.FolderRole.FOLDER_ROLE_DRAFTS => Icons.edit_note_outlined,
  pb.FolderRole.FOLDER_ROLE_ARCHIVE => Icons.archive_outlined,
  pb.FolderRole.FOLDER_ROLE_SPAM => Icons.report_outlined,
  pb.FolderRole.FOLDER_ROLE_TRASH => Icons.delete_outline,
  _ => Icons.folder_outlined,
};

/// A folder of an account by its own name; system folders by their role.
String folderName(pb.Folder f) => f.role == pb.FolderRole.FOLDER_ROLE_OTHER ? f.name : roleName(f.role);

String accountName(pb.Account? a) => a == null ? '' : (a.name.isEmpty ? a.login : a.name);

String addressName(pb.Address a) => a.name.isNotEmpty ? a.name : a.email;

String viewTitle(MailView v, MailState s, MessengerState m) {
  if (v.account.isEmpty) return '${roleName(v.role)} · ${tr('all accounts', 'все аккаунты')}';
  final f = s.folders.where((f) => f.accountId == v.account && f.id == v.folder).firstOrNull;
  return '${f == null ? roleName(v.role) : folderName(f)} · ${accountName(m.account(v.account))}';
}

/// Mail of every account in one place. Desktop: folders, list and the
/// open message side by side. Phone: a folder picker over the list; a
/// message is pushed.
class MailScreen extends StatelessWidget {
  const MailScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final m = context.watch<MessengerCubit>().state;
    final s = context.watch<MailCubit>().state;
    if (mailAccounts(m).isEmpty) {
      return Center(
        child: HmEmptyState(
          tr(
            'No mail accounts yet. Add Google or Mail.ru / VK WorkSpace to read all your mail here.',
            'Почтовых аккаунтов пока нет. Добавьте Google или Mail.ru / VK WorkSpace, чтобы читать здесь всю почту.',
          ),
          icon: Icons.mail_outline,
          action: HmButton(tr('Add account', 'Добавить аккаунт'), variant: HmVariant.primary, icon: Icons.add, onPressed: () => showLoginWizard(context)),
        ),
      );
    }
    if (Hm.of(context).mobile) {
      return const Column(
        children: [
          _FolderBar(),
          Expanded(child: _MailList()),
        ],
      );
    }
    return LayoutBuilder(
      builder: (context, box) {
        final railW = box.maxWidth >= 1280 ? 240.0 : 220.0;
        final listW = box.maxWidth >= 1100 ? 380.0 : 320.0;
        return Row(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            SizedBox(width: railW, child: const _FolderRail()),
            SizedBox(width: listW, child: const _MailList()),
            Expanded(
              child: s.open == null
                  ? HmEmptyState(tr('Choose a message', 'Выберите письмо'), icon: Icons.mail_outline)
                  : MailReader(key: ValueKey(s.open), mail: s.open!),
            ),
          ],
        );
      },
    );
  }
}

/// Opens a message: on the right on desktops, as a new screen on phones.
void openMail(BuildContext context, MailKey key) {
  final mail = context.read<MailCubit>();
  unawaited(mail.openMail(key));
  if (!Hm.of(context).mobile) return;
  Navigator.of(context)
      .push(
        MaterialPageRoute<void>(
          builder: (_) => BlocProvider.value(
            value: mail,
            child: Scaffold(
              backgroundColor: HmColors.bg,
              body: SafeArea(
                child: MailReader(mail: key, onBack: () => Navigator.of(context).maybePop()),
              ),
            ),
          ),
        ),
      )
      .whenComplete(() {
        if (mail.state.open == key) unawaited(mail.openMail(null));
      });
}

class _FolderRail extends StatelessWidget {
  const _FolderRail();

  @override
  Widget build(BuildContext context) {
    final s = context.watch<MailCubit>().state;
    final m = context.watch<MessengerCubit>().state;
    final mail = context.read<MailCubit>();
    final accounts = mailAccounts(m);
    Widget folder(String label, IconData icon, int count, bool selected, VoidCallback onTap) => AccountSubItem(
      leading: Icon(icon, size: 18, color: selected ? HmColors.accent : HmColors.muted),
      title: Text(label, overflow: TextOverflow.ellipsis),
      trailing: HmCount(count),
      selected: selected,
      onTap: onTap,
    );
    void inbox(String account) {
      final f = s.foldersOf(account).where((f) => f.role == pb.FolderRole.FOLDER_ROLE_INBOX).firstOrNull;
      if (f != null) mail.setView((role: f.role, account: account, folder: f.id));
    }

    return AccountRail(
      top: Padding(
        padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
        child: HmButton(tr('Write', 'Написать'), variant: HmVariant.primary, icon: Icons.edit_outlined, onPressed: () => showCompose(context)),
      ),
      child: ReorderableAccounts(
        accounts: accounts,
        first: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            HmListItem(
              leading: const HmAvatar(label: '', icon: Icons.all_inbox_outlined),
              title: Text(tr('All mail', 'Вся почта')),
              subtitle: Text(tr('${accounts.length} accounts', 'аккаунтов: ${accounts.length}')),
              trailing: HmCount(s.unreadOf(pb.FolderRole.FOLDER_ROLE_INBOX)),
              onTap: () => mail.setView((role: pb.FolderRole.FOLDER_ROLE_INBOX, account: '', folder: '')),
            ),
            for (final r in unifiedRoles)
              if (r == pb.FolderRole.FOLDER_ROLE_INBOX || s.folders.any((f) => f.role == r))
                folder(
                  roleName(r),
                  roleIcon(r),
                  r == pb.FolderRole.FOLDER_ROLE_INBOX ? s.unreadOf(r) : 0,
                  s.view.account.isEmpty && s.view.role == r,
                  () => mail.setView((role: r, account: '', folder: '')),
                ),
          ],
        ),
        item: (a) => Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            AccountTile(account: a, count: a.mailUnread, onTap: () => inbox(a.id)),
            for (final f in s.foldersOf(a.id))
              folder(
                folderName(f),
                roleIcon(f.role),
                f.role == pb.FolderRole.FOLDER_ROLE_SPAM || f.role == pb.FolderRole.FOLDER_ROLE_TRASH ? 0 : f.unread,
                s.view.account == a.id && s.view.folder == f.id,
                () => mail.setView((role: f.role, account: a.id, folder: f.id)),
              ),
          ],
        ),
      ),
    );
  }
}

/// The accounts with mail, in the user's order.
List<pb.Account> mailAccounts(MessengerState m) => accountsOffering(m, pb.Capability.CAPABILITY_MAIL);

/// Phone: the current folder as a picker, and the compose button.
class _FolderBar extends StatelessWidget {
  const _FolderBar();

  @override
  Widget build(BuildContext context) {
    final s = context.watch<MailCubit>().state;
    final m = context.watch<MessengerCubit>().state;
    final mail = context.read<MailCubit>();
    final options = <MailView, String>{
      for (final r in unifiedRoles)
        if (r == pb.FolderRole.FOLDER_ROLE_INBOX || s.folders.any((f) => f.role == r))
          (role: r, account: '', folder: ''): '${roleName(r)} · ${tr('all accounts', 'все аккаунты')}',
      for (final a in mailAccounts(m))
        for (final f in s.foldersOf(a.id)) (role: f.role, account: a.id, folder: f.id): '${accountName(a)} · ${folderName(f)}',
    };
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 8, 8, 8),
      decoration: const BoxDecoration(
        color: HmColors.surface,
        border: Border(bottom: BorderSide(color: HmColors.border)),
      ),
      child: Row(
        children: [
          Expanded(
            child: HmSelect<MailView>(
              label: tr('Folder', 'Папка'),
              showLabel: false,
              options: options,
              value: options.containsKey(s.view) ? s.view : null,
              placeholder: viewTitle(s.view, s, m),
              onChanged: mail.setView,
            ),
          ),
          const SizedBox(width: 4),
          HmIconButton(icon: Icons.edit_outlined, label: tr('Write', 'Написать'), onPressed: () => showCompose(context)),
        ],
      ),
    );
  }
}

class _MailList extends StatefulWidget {
  const _MailList();

  @override
  State<_MailList> createState() => _MailListState();
}

class _MailListState extends State<_MailList> {
  final _scroll = ScrollController();

  @override
  void initState() {
    super.initState();
    _scroll.addListener(() {
      if (_scroll.position.extentAfter < 400) context.read<MailCubit>().loadMore();
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
    final s = context.watch<MailCubit>().state;
    final m = context.watch<MessengerCubit>().state;
    final mail = context.read<MailCubit>();
    final unified = s.view.account.isEmpty;
    Widget list;
    if (s.loading && s.mails.isEmpty) {
      list = const HmLoading();
    } else if (s.mails.isEmpty) {
      list = ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        children: [
          const SizedBox(height: 40),
          HmEmptyState(s.error ?? tr('No messages', 'Писем нет'), icon: Icons.mail_outline),
        ],
      );
    } else {
      list = ListView.builder(
        controller: _scroll,
        physics: const AlwaysScrollableScrollPhysics(),
        itemCount: s.mails.length + (s.more ? 1 : 0),
        itemBuilder: (context, i) {
          if (i == s.mails.length) {
            return const Padding(
              padding: EdgeInsets.all(16),
              child: Center(child: HmSpinner()),
            );
          }
          final x = s.mails[i];
          final a = m.account(x.accountId);
          final k = mailKey(x);
          final selecting = s.selected.isNotEmpty;
          return _MailTile(
            mail: x,
            account: unified ? a : null,
            selected: !hm.mobile && s.open == k,
            selecting: selecting,
            checked: s.selected.contains(k),
            onCheck: () => mail.toggleSelected(k),
            onTap: selecting ? () => mail.toggleSelected(k) : () => openMail(context, k),
          );
        },
      );
    }
    if (hm.mobile) {
      list = RefreshIndicator(onRefresh: mail.refresh, color: HmColors.accent, backgroundColor: HmColors.surface, child: list);
    }
    return Container(
      decoration: BoxDecoration(
        color: hm.mobile ? null : HmColors.bg,
        border: hm.mobile ? null : const Border(right: BorderSide(color: HmColors.border)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          if (s.selected.isNotEmpty)
            const _MailSelectionBar()
          else if (!hm.mobile)
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 12, 12),
              child: Row(
                children: [
                  Expanded(
                    child: Text(viewTitle(s.view, s, m), style: HmText.modalTitle, overflow: TextOverflow.ellipsis),
                  ),
                  HmIconButton(icon: Icons.refresh, label: tr('Refresh', 'Обновить'), onPressed: mail.load),
                ],
              ),
            ),
          for (final f in s.failed)
            Padding(
              padding: const EdgeInsets.fromLTRB(12, 0, 12, 8),
              child: HmBanner('${accountName(m.account(f.accountId))}: ${f.error}', tone: HmTone.warning),
            ),
          if (s.error != null && s.mails.isNotEmpty)
            Padding(
              padding: const EdgeInsets.fromLTRB(12, 0, 12, 8),
              child: HmBanner(s.error!, tone: HmTone.error),
            ),
          const Divider(),
          Expanded(child: list),
        ],
      ),
    );
  }
}

class _MailTile extends StatelessWidget {
  const _MailTile({
    required this.mail,
    required this.account,
    required this.selected,
    required this.onTap,
    required this.onCheck,
    this.selecting = false,
    this.checked = false,
  });

  final pb.Mail mail;
  final pb.Account? account; // set in merged views: the account's marker shows
  final bool selected;
  final VoidCallback onTap;

  /// Ticks or unticks it: a tap on the avatar, a long press, a right click.
  final VoidCallback onCheck;

  /// Messages are being ticked: a check box instead of the avatar.
  final bool selecting;
  final bool checked;

  @override
  Widget build(BuildContext context) {
    final x = mail;
    final sent = x.folderId.isNotEmpty && x.to.isNotEmpty && (account?.login.toLowerCase() == x.from.email.toLowerCase());
    final who = sent ? '${tr('To', 'Кому')}: ${x.to.map(addressName).join(', ')}' : addressName(x.from);
    final strong = x.unread ? HmText.bodyStrong : HmText.body;
    final provider = account?.provider ?? '';
    final size = Hm.of(context).mobile ? 40.0 : 36.0;
    final avatar = selecting
        ? SizedBox.square(
            dimension: size,
            child: Icon(
              checked ? Icons.check_circle : Icons.radio_button_unchecked,
              size: 24,
              color: checked ? HmColors.accent : HmColors.muted,
              semanticLabel: checked ? tr('Selected', 'Выбрано') : tr('Not selected', 'Не выбрано'),
            ),
          )
        : HmAvatar(
            label: who,
            color: account == null ? HmColors.accent : providerColor(provider),
            size: size,
            corner: account == null ? null : providerMarker(provider, tooltip: accountName(account)),
          );
    return HmListItem(
      leading: MouseRegion(
        cursor: SystemMouseCursors.click,
        child: GestureDetector(
          behavior: HitTestBehavior.opaque,
          onTap: onCheck,
          child: Tooltip(message: tr('Select', 'Выбрать'), waitDuration: const Duration(milliseconds: 600), child: avatar),
        ),
      ),
      onLongPress: onCheck,
      title: Text(
        who,
        overflow: TextOverflow.ellipsis,
        style: strong.copyWith(color: x.unread ? HmColors.accent : null),
      ),
      meta: Text(formatChatTime(x.timeMs)),
      subtitle: Text.rich(
        TextSpan(
          children: [
            TextSpan(
              text: x.subject.isEmpty ? tr('(no subject)', '(без темы)') : x.subject,
              style: HmText.small.copyWith(color: HmColors.text),
            ),
            if (x.snippet.isNotEmpty) TextSpan(text: ' — ${x.snippet}'),
          ],
        ),
        maxLines: 2,
        overflow: TextOverflow.ellipsis,
      ),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (x.hasAttachments) const Icon(Icons.attach_file, size: 14, color: HmColors.muted),
          if (x.starred) const Icon(Icons.star, size: 14, color: HmColors.warning),
        ],
      ),
      selected: selected || checked,
      onTap: onTap,
    );
  }
}

/// Instead of the list's header while messages are ticked: how many, tick
/// all, and what to do with them (Escape drops the ticks).
class _MailSelectionBar extends StatelessWidget {
  const _MailSelectionBar();

  Future<void> _run(Future<void> Function() action, String done) async {
    try {
      await action();
      Toasts.success(done);
    } catch (e) {
      Toasts.error(tr('Not everything was changed', 'Изменилось не всё'), errorText(e));
    }
  }

  Future<void> _act(MailCubit cubit, pb.MailAction a, String done) {
    final keys = cubit.state.selected;
    if (a == pb.MailAction.MAIL_ACTION_TRASH) cubit.clearSelected();
    return _run(() => cubit.actMany(keys, a), done);
  }

  /// Picks where to: a folder of the account when all are of one, else a
  /// role every account has its own folder of.
  Future<void> _move(BuildContext context, MailCubit cubit) async {
    final s = cubit.state;
    final mails = s.selectedMails;
    final accounts = {for (final m in mails) m.accountId};
    final from = {for (final m in mails) m.folderId};
    const nowhere = {pb.FolderRole.FOLDER_ROLE_SENT, pb.FolderRole.FOLDER_ROLE_DRAFTS};
    final Map<String, String> options;
    if (accounts.length == 1) {
      options = {
        for (final f in s.foldersOf(accounts.first))
          if (!nowhere.contains(f.role) && !(from.length == 1 && from.first == f.id)) 'f:${f.id}': folderName(f),
      };
    } else {
      options = {
        for (final r in unifiedRoles)
          if (!nowhere.contains(r) && s.folders.any((f) => f.role == r && accounts.contains(f.accountId))) 'r:${r.value}': roleName(r),
      };
    }
    if (options.isEmpty) return;
    final v = await pickOption<String>(context, title: tr('Move to', 'Переместить в'), options: options);
    if (v == null) return;
    final role = v.startsWith('r:') ? pb.FolderRole.valueOf(int.parse(v.substring(2))) : null;
    final folder = v.startsWith('f:') ? s.foldersOf(accounts.first).where((f) => f.id == v.substring(2)).firstOrNull : null;
    cubit.clearSelected();
    var skipped = 0;
    await _run(() async => skipped = await cubit.move(mails, role: role, folder: folder), '${tr('Moved to', 'Перемещено в')} ${options[v]}');
    if (skipped > 0) {
      Toasts.error(tr('$skipped not moved', 'Не перемещено: $skipped'), tr('Their account has no such folder', 'В их аккаунте нет такой папки'));
    }
  }

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    final cubit = context.read<MailCubit>();
    final s = context.watch<MailCubit>().state;
    final picked = s.selectedMails;
    final n = s.selected.length;
    final anyUnread = picked.any((m) => m.unread);
    final anyUnstarred = picked.any((m) => !m.starred);
    return CallbackShortcuts(
      bindings: {const SingleActivator(LogicalKeyboardKey.escape): cubit.clearSelected},
      child: Focus(
        autofocus: true,
        child: Container(
          height: hm.mobile ? 56 : 60,
          padding: EdgeInsets.symmetric(horizontal: hm.mobile ? 4 : 8),
          decoration: const BoxDecoration(
            color: HmColors.surface,
            border: Border(bottom: BorderSide(color: HmColors.border)),
          ),
          child: Row(
            children: [
              HmIconButton(icon: Icons.close, label: tr('Cancel', 'Отмена'), onPressed: cubit.clearSelected),
              const SizedBox(width: 4),
              Expanded(
                child: Text('$n', style: HmText.bodyStrong, overflow: TextOverflow.ellipsis, semanticsLabel: tr('Selected: $n', 'Выбрано: $n')),
              ),
              HmIconButton(icon: Icons.select_all, label: tr('Select all', 'Выбрать все'), onPressed: cubit.selectAll),
              anyUnread
                  ? HmIconButton(
                      icon: Icons.mark_email_read_outlined,
                      label: tr('Mark read', 'Пометить прочитанными'),
                      onPressed: () => _act(cubit, pb.MailAction.MAIL_ACTION_READ, tr('Marked read', 'Помечено прочитанными')),
                    )
                  : HmIconButton(
                      icon: Icons.mark_email_unread_outlined,
                      label: tr('Mark unread', 'Пометить непрочитанными'),
                      onPressed: () => _act(cubit, pb.MailAction.MAIL_ACTION_UNREAD, tr('Marked unread', 'Помечено непрочитанными')),
                    ),
              anyUnstarred
                  ? HmIconButton(
                      icon: Icons.star_border,
                      label: tr('Star', 'Отметить'),
                      onPressed: () => _act(cubit, pb.MailAction.MAIL_ACTION_STAR, tr('Starred', 'Отмечено')),
                    )
                  : HmIconButton(
                      icon: Icons.star,
                      label: tr('Unstar', 'Снять отметку'),
                      color: HmColors.warning,
                      onPressed: () => _act(cubit, pb.MailAction.MAIL_ACTION_UNSTAR, tr('Unstarred', 'Отметка снята')),
                    ),
              Builder(
                builder: (context) =>
                    HmIconButton(icon: Icons.drive_file_move_outline, label: tr('Move to…', 'Переместить в…'), onPressed: () => _move(context, cubit)),
              ),
              HmIconButton(
                icon: Icons.delete_outline,
                label: tr('Delete', 'Удалить'),
                color: HmColors.error,
                onPressed: () => _act(cubit, pb.MailAction.MAIL_ACTION_TRASH, tr('Deleted: $n', 'Удалено: $n')),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// One message: actions, headers, attachments and the text.
class MailReader extends StatelessWidget {
  const MailReader({super.key, required this.mail, this.onBack});

  final MailKey mail;
  final VoidCallback? onBack;

  @override
  Widget build(BuildContext context) {
    final s = context.watch<MailCubit>().state;
    final m = context.watch<MessengerCubit>().state;
    final cubit = context.read<MailCubit>();
    final hm = Hm.of(context);
    final body = s.body != null && mailKey(s.body!) == mail ? s.body : null;
    final x = body ?? s.summary(mail);
    Future<void> act(pb.MailAction a) async {
      try {
        await cubit.act(mail, a);
        if (a == pb.MailAction.MAIL_ACTION_TRASH) {
          Toasts.success(tr('Moved to the trash', 'Перемещено в корзину'));
          if (onBack != null) onBack!();
        }
      } catch (e) {
        Toasts.error(tr('Could not change the message', 'Не удалось изменить письмо'), errorText(e));
      }
    }

    final toolbar = Container(
      height: hm.mobile ? 56 : 52,
      padding: EdgeInsets.symmetric(horizontal: hm.mobile ? 4 : 16),
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: HmColors.border)),
      ),
      child: Row(
        children: [
          if (onBack != null) HmIconButton(icon: Icons.arrow_back, label: tr('Back', 'Назад'), onPressed: onBack),
          const Spacer(),
          if (x != null) ...[
            HmIconButton(
              icon: Icons.reply,
              label: tr('Reply', 'Ответить'),
              onPressed: body == null ? null : () => showCompose(context, reply: body),
            ),
            HmIconButton(
              icon: Icons.reply_all,
              label: tr('Reply all', 'Ответить всем'),
              onPressed: body == null ? null : () => showCompose(context, reply: body, all: true),
            ),
            HmIconButton(
              icon: Icons.forward,
              label: tr('Forward', 'Переслать'),
              onPressed: body == null ? null : () => showCompose(context, forward: body),
            ),
            HmIconButton(
              icon: x.starred ? Icons.star : Icons.star_border,
              label: x.starred ? tr('Unstar', 'Снять отметку') : tr('Star', 'Отметить'),
              color: x.starred ? HmColors.warning : null,
              onPressed: () => act(x.starred ? pb.MailAction.MAIL_ACTION_UNSTAR : pb.MailAction.MAIL_ACTION_STAR),
            ),
            HmIconButton(
              icon: Icons.mark_email_unread_outlined,
              label: tr('Mark unread', 'Пометить непрочитанным'),
              onPressed: () => act(pb.MailAction.MAIL_ACTION_UNREAD),
            ),
            HmIconButton(icon: Icons.delete_outline, label: tr('Delete', 'Удалить'), onPressed: () => act(pb.MailAction.MAIL_ACTION_TRASH)),
          ],
        ],
      ),
    );
    if (x == null) {
      return Column(
        children: [
          toolbar,
          const Expanded(child: HmLoading()),
        ],
      );
    }
    final a = m.account(x.accountId);
    String list(List<pb.Address> l) => l.map((a) => a.name.isEmpty ? a.email : '${a.name} <${a.email}>').join(', ');
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        toolbar,
        Expanded(
          child: SelectionArea(
            child: ListView(
              padding: EdgeInsets.fromLTRB(hm.pageX, 20, hm.pageX, hm.pageBottom),
              children: [
                Text(x.subject.isEmpty ? tr('(no subject)', '(без темы)') : x.subject, style: HmText.dialogHeading),
                const SizedBox(height: 16),
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    HmAvatar(label: addressName(x.from), color: providerColor(a?.provider ?? '')),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(addressName(x.from), style: HmText.bodyStrong),
                          if (x.from.name.isNotEmpty) Text(x.from.email, style: HmText.small),
                          const SizedBox(height: 4),
                          if (x.to.isNotEmpty) Text('${tr('To', 'Кому')}: ${list(x.to)}', style: HmText.small),
                          if (x.cc.isNotEmpty) Text('${tr('Cc', 'Копия')}: ${list(x.cc)}', style: HmText.small),
                          Text('${formatStamp(x.timeMs)} · ${accountName(a)}', style: HmText.small),
                        ],
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                if (body != null && body.attachments.isNotEmpty) ...[
                  MessageAttachments(
                    message: pb.Message(accountId: x.accountId, attachments: body.attachments),
                    maxWidth: 360,
                  ),
                  const SizedBox(height: 16),
                ],
                const Divider(),
                const SizedBox(height: 16),
                if (s.bodyError != null && body == null)
                  HmBanner(s.bodyError!, tone: HmTone.error)
                else if (body == null)
                  const Padding(
                    padding: EdgeInsets.all(24),
                    child: Center(child: HmSpinner()),
                  )
                else
                  HmRichText(
                    body.text.isEmpty ? tr('(empty message)', '(пустое письмо)') : body.text,
                    spans: body.format,
                    style: HmText.body.copyWith(height: 1.6),
                  ),
              ],
            ),
          ),
        ),
      ],
    );
  }
}
