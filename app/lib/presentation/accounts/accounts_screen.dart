import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/layout.dart';
import '../../core/theme/typography.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../common/hm.dart';
import '../core/messenger_cubit.dart';
import '../shell/app_shell.dart' show AppShell;
import 'login_wizard.dart';
import 'providers.dart';

/// Messenger accounts, grouped by provider: status, rename, enable or
/// disable, sign in again, remove; and adding new ones.
class AccountsScreen extends StatelessWidget {
  const AccountsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    final m = context.watch<MessengerCubit>().state;
    // Providers in the core's order; accounts of unknown ones after them.
    final types = [for (final p in m.providers) p.type];
    for (final a in m.accounts) {
      if (!types.contains(a.provider)) types.add(a.provider);
    }
    final online = m.accounts.where(isOnline).length;
    final needLogin = m.accounts.where((a) => a.status == pb.AccountStatus.ACCOUNT_STATUS_NEEDS_LOGIN).length;
    final add = HmButton(
      tr('Add account', 'Добавить аккаунт'),
      variant: HmVariant.primary,
      icon: Icons.add,
      expand: hm.mobile,
      onPressed: () => showLoginWizard(context),
    );
    return HmPage(
      onRefresh: context.read<MessengerCubit>().load,
      children: [
        HmPageTitle(tr('Accounts', 'Аккаунты'), trailing: hm.mobile ? null : add),
        Text(
          tr(
            'Each account keeps its own session in the core. Several accounts of one messenger are fine.',
            'Каждый аккаунт держит свою сессию в ядре. Несколько аккаунтов одного мессенджера — это нормально.',
          ),
          style: HmText.small.copyWith(height: 1.5),
        ),
        if (hm.mobile) add,
        HmStatCards(
          children: [
            HmStatCard(label: tr('Accounts', 'Аккаунтов'), value: '${m.accounts.length}', accent: HmAccent.total),
            HmStatCard(label: tr('Online', 'В сети'), value: '$online', accent: HmAccent.ok),
            HmStatCard(
              label: tr('Need sign-in', 'Нужен вход'),
              value: '$needLogin',
              accent: needLogin == 0 ? HmAccent.paused : HmAccent.error,
            ),
          ],
        ),
        if (m.error != null) HmBanner(m.error!, tone: HmTone.error),
        if (m.loading && m.accounts.isEmpty) const SizedBox(height: 160, child: HmLoading()),
        if (!m.loading && m.accounts.isEmpty)
          HmEmptyState(tr('No accounts yet', 'Аккаунтов пока нет'), icon: Icons.manage_accounts_outlined),
        for (final t in types)
          if (m.accounts.any((a) => a.provider == t))
            Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                HmSectionTitle(providerName(t, m.provider(t)).toUpperCase()),
                for (final a in m.accounts.where((a) => a.provider == t)) ...[SizedBox(height: hm.gapSm), _AccountCard(account: a)],
              ],
            ),
      ],
    );
  }
}

class _AccountCard extends StatelessWidget {
  const _AccountCard({required this.account});

  final pb.Account account;

  Future<void> _rename(BuildContext context) async {
    final messenger = context.read<MessengerCubit>();
    final c = TextEditingController(text: account.name);
    String? error;
    await showHmModal<void>(
      context,
      title: tr('Rename account', 'Переименовать аккаунт'),
      size: HmModalSize.sm,
      body: (ctx) => StatefulBuilder(
        builder: (ctx, set) {
          Future<void> save() async {
            final name = c.text.trim();
            if (name.isEmpty) return set(() => error = tr('Enter a name', 'Введите имя'));
            try {
              await messenger.rename(account.id, name);
              if (ctx.mounted) Navigator.pop(ctx);
            } catch (e) {
              set(() => error = errorText(e));
            }
          }

          return Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              HmTextField(
                label: tr('Name', 'Имя'),
                required: true,
                autofocus: true,
                controller: c,
                error: error,
                hint: tr('Shown in the chat list; only on your devices', 'Видно в списке чатов; только на ваших устройствах'),
                onSubmitted: (_) => save(),
              ),
              const SizedBox(height: 16),
              HmButton(tr('Save', 'Сохранить'), variant: HmVariant.primary, onPressed: save),
            ],
          );
        },
      ),
    ).whenComplete(c.dispose);
  }

  Future<void> _setEnabled(BuildContext context, bool on) async {
    final messenger = context.read<MessengerCubit>();
    try {
      await messenger.setEnabled(account.id, on);
    } catch (e) {
      Toasts.error(on ? tr('Could not enable', 'Не удалось включить') : tr('Could not disable', 'Не удалось выключить'), errorText(e));
    }
  }

  Future<void> _remove(BuildContext context) async {
    final messenger = context.read<MessengerCubit>();
    final ok = await confirm(
      context,
      title: tr('Remove ${account.name}?', 'Удалить ${account.name}?'),
      message: tr(
        'The app signs out of this account and forgets its cached chats and messages. Nothing is deleted in the messenger itself.',
        'Приложение выйдет из этого аккаунта и забудет его чаты и сообщения в кэше. В самом мессенджере ничего не удаляется.',
      ),
      action: tr('Remove', 'Удалить'),
      destructive: true,
    );
    if (!ok) return;
    try {
      await messenger.remove(account.id);
      Toasts.success(tr('Account removed', 'Аккаунт удалён'), account.name);
    } catch (e) {
      Toasts.error(tr('Could not remove', 'Не удалось удалить'), errorText(e));
    }
  }

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    final a = account;
    final (label, tone) = accountStatus(a);
    final needsLogin = a.status == pb.AccountStatus.ACCOUNT_STATUS_NEEDS_LOGIN;
    final status = switch (tone) {
      HmTone.ok => HmStatus.ok,
      HmTone.error => HmStatus.error,
      HmTone.warning => HmStatus.warning,
      _ => HmStatus.paused,
    };
    final info = Row(
      children: [
        HmAvatar(label: a.name, icon: providerIcon(a.provider), color: providerColor(a.provider), size: 40),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Wrap(
                spacing: 8,
                runSpacing: 4,
                crossAxisAlignment: WrapCrossAlignment.center,
                children: [
                  Text(a.name.isEmpty ? a.login : a.name, style: HmText.bodyStrong),
                  HmBadge(label, tone: tone),
                  if (a.enabled && a.unread > 0) HmCount(a.unread),
                ],
              ),
              const SizedBox(height: 4),
              Text(
                [if (a.login.isNotEmpty) a.login, if (a.self.name.isNotEmpty && a.self.name != a.name) a.self.name].join(' · '),
                style: HmText.small,
                overflow: TextOverflow.ellipsis,
              ),
              if (a.statusMessage.isNotEmpty) ...[
                const SizedBox(height: 4),
                Text(a.statusMessage, style: HmText.small.copyWith(color: needsLogin ? HmColors.error : HmColors.muted)),
              ],
            ],
          ),
        ),
      ],
    );
    final actions = Wrap(
      spacing: 8,
      runSpacing: 8,
      crossAxisAlignment: WrapCrossAlignment.center,
      children: [
        if (needsLogin)
          HmButton(
            tr('Sign in again', 'Войти снова'),
            small: true,
            variant: HmVariant.primary,
            icon: Icons.login,
            onPressed: () => showLoginWizard(context, relogin: a),
          ),
        HmActionButton(icon: Icons.edit_outlined, label: tr('Rename', 'Переименовать'), onPressed: () => _rename(context)),
        a.enabled
            ? HmActionButton(
                icon: Icons.pause_circle_outline,
                label: tr('Disable', 'Выключить'),
                kind: HmAction.warning,
                onPressed: () => _setEnabled(context, false),
              )
            : HmActionButton(
                icon: Icons.play_circle_outline,
                label: tr('Enable', 'Включить'),
                kind: HmAction.success,
                onPressed: () => _setEnabled(context, true),
              ),
        HmActionButton(
          icon: Icons.receipt_long_outlined,
          label: tr('Log', 'Журнал'),
          kind: HmAction.neutral,
          onPressed: () => AppShell.openLog(context, a.id),
        ),
        HmActionButton(
          icon: Icons.delete_outline,
          label: tr('Remove', 'Удалить'),
          kind: HmAction.delete,
          onPressed: () => _remove(context),
        ),
      ],
    );
    return HmCard(
      status: status,
      padding: const EdgeInsets.all(16),
      child: hm.mobile
          ? Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [info, const SizedBox(height: 12), actions])
          : Row(
              children: [
                Expanded(child: info),
                const SizedBox(width: 16),
                actions,
              ],
            ),
    );
  }
}
