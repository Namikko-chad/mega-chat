import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../data/rpc/core_client.dart';
import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/layout.dart';
import '../../core/theme/typography.dart';
import '../../data/rpc/gen/parley.pbgrpc.dart' as pb;
import '../chats/media.dart';
import '../common/hm.dart';
import '../core/messenger_cubit.dart';
import 'login_wizard.dart';
import 'providers.dart';

/// Whether an account offers [cap]: what its session reports, else what its
/// messenger declares (an account that is off or signing in has no session).
bool accountOffers(MessengerState m, pb.Account a, pb.Capability cap) =>
    a.capabilities.contains(cap) || (m.provider(a.provider)?.capabilities.contains(cap) ?? false);

/// The accounts of one page (chats, mail, calendar): only those that offer
/// [cap], in the user's order.
List<pb.Account> accountsOffering(MessengerState m, pb.Capability cap) => m.accounts.where((a) => accountOffers(m, a, cap)).toList();

/// The side panel of accounts: a title, [child] (usually [ReorderableAccounts])
/// and "Add account" at the bottom; [top] goes above the title.
class AccountRail extends StatelessWidget {
  const AccountRail({super.key, required this.child, this.top});

  final Widget child;
  final Widget? top;

  @override
  Widget build(BuildContext context) => Container(
    decoration: const BoxDecoration(
      color: HmColors.surface,
      border: Border(right: BorderSide(color: HmColors.border)),
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        ?top,
        Padding(
          padding: EdgeInsets.fromLTRB(16, top == null ? 20 : 12, 16, 12),
          child: Text(tr('ACCOUNTS', 'АККАУНТЫ'), style: HmText.sectionTitle),
        ),
        Expanded(child: child),
        Padding(
          padding: const EdgeInsets.all(16),
          child: HmButton(tr('Add account', 'Добавить аккаунт'), icon: Icons.add, onPressed: () => showLoginWizard(context)),
        ),
      ],
    ),
  );
}

/// [accounts] (a part of all of them, in their order) after an optional
/// fixed first item ("All"). Mouse: drag an account; touch: long press, so
/// that a swipe still scrolls. An account's item may carry what is under it
/// (folders, calendars): it moves with it.
class ReorderableAccounts extends StatelessWidget {
  const ReorderableAccounts({super.key, required this.accounts, required this.item, this.first, this.axis = Axis.vertical, this.padding = EdgeInsets.zero});

  final List<pb.Account> accounts;
  final Widget Function(pb.Account) item;
  final Widget? first;
  final Axis axis;
  final EdgeInsets padding;

  @override
  Widget build(BuildContext context) {
    final all = context.watch<MessengerCubit>().state.accounts;
    final messenger = context.read<MessengerCubit>();
    final touch = defaultTargetPlatform == TargetPlatform.android || defaultTargetPlatform == TargetPlatform.iOS;
    final vertical = axis == Axis.vertical;
    int global(pb.Account a) => all.indexWhere((x) => x.id == a.id);
    return CustomScrollView(
      scrollDirection: axis,
      slivers: [
        if (first != null)
          SliverPadding(
            padding: vertical ? padding.copyWith(bottom: 0) : padding.copyWith(right: 0),
            sliver: SliverToBoxAdapter(child: first),
          ),
        SliverPadding(
          padding: first == null ? padding : (vertical ? padding.copyWith(top: 0) : padding.copyWith(left: 0)),
          sliver: SliverReorderableList(
            itemCount: accounts.length,
            proxyDecorator: (child, _, _) => Material(type: MaterialType.transparency, child: child),
            onReorder: (from, to) async {
              // Indexes of the part shown → indexes of all the accounts.
              final f = global(accounts[from]);
              final t = to < accounts.length ? global(accounts[to]) : global(accounts.last) + 1;
              if (f < 0 || t < 0) return;
              try {
                await messenger.moveAccount(f, t);
              } catch (e) {
                Toasts.error(tr('Could not save the order', 'Не удалось сохранить порядок'), errorText(e));
              }
            },
            itemBuilder: (_, i) {
              final key = ValueKey(accounts[i].id);
              final child = item(accounts[i]);
              return touch
                  ? ReorderableDelayedDragStartListener(key: key, index: i, child: child)
                  : ReorderableDragStartListener(key: key, index: i, child: child);
            },
          ),
        ),
      ],
    );
  }
}

/// An account in a rail: avatar with the messenger's mark, name, online
/// state and [count] (unread chats or mail).
class AccountTile extends StatelessWidget {
  const AccountTile({super.key, required this.account, required this.count, this.selected = false, this.onTap});

  final pb.Account account;
  final int count;
  final bool selected;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final a = account;
    final (status, _) = accountStatus(a);
    return HmListItem(
      leading: AccountAvatar(account: a),
      title: Text(a.name.isEmpty ? a.login : a.name),
      subtitle: Row(
        children: [
          if (a.enabled) ...[HmDot(ok: isOnline(a), label: status), const SizedBox(width: 6)],
          Flexible(child: Text(a.statusMessage.isNotEmpty && !isOnline(a) ? a.statusMessage : status, overflow: TextOverflow.ellipsis)),
        ],
      ),
      trailing: HmCount(a.enabled ? count : 0),
      selected: selected,
      onTap: onTap,
    );
  }
}

/// The account's own picture with the messenger's mark, else the
/// messenger's icon.
class AccountAvatar extends StatelessWidget {
  const AccountAvatar({super.key, required this.account, this.size});

  final pb.Account account;
  final double? size;

  @override
  Widget build(BuildContext context) {
    final a = account;
    return MediaAvatar(
      account: a.id,
      photo: a.self.avatarUrl,
      label: a.name,
      icon: providerIcon(a.provider),
      color: providerColor(a.provider),
      size: size ?? 36,
      corner: a.self.avatarUrl.isEmpty ? null : providerMarker(a.provider),
    );
  }
}

/// A line under an account in a rail (a mail folder, a calendar), indented
/// so that it reads as the account's.
class AccountSubItem extends StatelessWidget {
  const AccountSubItem({super.key, required this.leading, required this.title, this.trailing, this.selected = false, this.onTap});

  final Widget leading;
  final Widget title;
  final Widget? trailing;
  final bool selected;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) => HmListItem(
    leading: Padding(
      padding: EdgeInsets.only(left: Hm.of(context).mobile ? 16 : 20),
      child: leading,
    ),
    title: title,
    trailing: trailing,
    selected: selected,
    onTap: onTap,
    dense: true,
  );
}
