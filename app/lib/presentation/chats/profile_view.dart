import 'package:fixnum/fixnum.dart';
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
import '../accounts/providers.dart';
import '../common/hm.dart';
import '../core/core_cubit.dart';
import '../core/messenger_cubit.dart';
import 'media.dart';

/// Opens the info page of a chat (its person, group or channel): a screen
/// of its own on phones, a dialog on desktops.
void showProfile(BuildContext context, ChatKey key) {
  if (Hm.of(context).mobile) {
    Navigator.of(context).push(
      MaterialPageRoute<void>(
        builder: (context) => Scaffold(
          backgroundColor: HmColors.bg,
          body: SafeArea(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Container(
                  height: 56,
                  padding: const EdgeInsets.symmetric(horizontal: 8),
                  decoration: const BoxDecoration(
                    color: HmColors.surface,
                    border: Border(bottom: BorderSide(color: HmColors.border)),
                  ),
                  child: Row(
                    children: [
                      HmIconButton(icon: Icons.arrow_back, label: tr('Back', 'Назад'), onPressed: () => Navigator.of(context).maybePop()),
                      const SizedBox(width: 8),
                      Text(tr('Info', 'Информация'), style: HmText.bodyStrong),
                    ],
                  ),
                ),
                Expanded(
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.fromLTRB(16, 24, 16, 32),
                    child: ProfileView(chat: key),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
    return;
  }
  showHmModal<void>(
    context,
    title: tr('Info', 'Информация'),
    body: (_) => ProfileView(chat: key),
  );
}

/// What is known of a chat: what the chat list has at once, then what the
/// messenger tells (CAPABILITY_PROFILES): about, contacts, members.
class ProfileView extends StatefulWidget {
  const ProfileView({super.key, required this.chat});

  final ChatKey chat;

  @override
  State<ProfileView> createState() => _ProfileViewState();
}

class _ProfileViewState extends State<ProfileView> {
  pb.ChatProfile? _profile;
  String? _error;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    final account = context.read<MessengerCubit>().state.account(widget.chat.account);
    if ((account?.capabilities ?? const []).contains(pb.Capability.CAPABILITY_PROFILES)) _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final p = await context.read<CoreCubit>().api.profile(pb.ProfileRequest(accountId: widget.chat.account, chatId: widget.chat.chat));
      if (mounted) setState(() => _profile = p);
    } catch (e) {
      if (mounted) setState(() => _error = errorText(e));
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _copy(String text) async {
    await Clipboard.setData(ClipboardData(text: text));
    Toasts.success(tr('Copied', 'Скопировано'));
  }

  @override
  Widget build(BuildContext context) {
    final m = context.watch<MessengerCubit>().state;
    final chat = m.chats[widget.chat];
    final account = m.account(widget.chat.account);
    final p = _profile;
    final provider = account?.provider ?? '';
    final kind = chat?.kind ?? pb.ChatKind.CHAT_KIND_UNSPECIFIED;
    final title = chat?.title.isNotEmpty ?? false ? chatTitle(chat!) : (p?.title ?? '');
    final photo = p != null && p.avatarUrl.isNotEmpty ? p.avatarUrl : (chat?.avatarUrl ?? '');
    final online = (chat?.online ?? false) || (p?.online ?? false);
    final where = [if (account != null) account.name, providerName(provider, m.provider(provider))].where((s) => s.isNotEmpty).join(' · ');

    final status = <String>[
      switch (kind) {
        pb.ChatKind.CHAT_KIND_DIRECT => p?.bot ?? false ? tr('bot', 'бот') : tr('person', 'контакт'),
        pb.ChatKind.CHAT_KIND_CHANNEL => tr('channel', 'канал'),
        _ => tr('group', 'группа'),
      },
      if (kind == pb.ChatKind.CHAT_KIND_DIRECT && online)
        tr('online', 'в сети')
      else if (kind == pb.ChatKind.CHAT_KIND_DIRECT && p != null && p.lastSeenMs > Int64.ZERO)
        '${tr('last seen', 'был(а) в сети')} ${formatChatTime(p.lastSeenMs)}',
      if (kind != pb.ChatKind.CHAT_KIND_DIRECT && p != null && p.members > 0)
        '${p.members} ${members(p.members, channel: kind == pb.ChatKind.CHAT_KIND_CHANNEL)}',
    ];

    final rows = <Widget>[
      if (p != null && p.about.isNotEmpty)
        _Row(
          label: kind == pb.ChatKind.CHAT_KIND_DIRECT ? tr('About', 'О себе') : tr('Description', 'Описание'),
          child: HmLinkText(p.about, style: HmText.body.copyWith(height: 1.45)),
        ),
      if (p != null && p.username.isNotEmpty) _Row(label: tr('Username', 'Имя пользователя'), value: '@${p.username}', onTap: () => _copy('@${p.username}')),
      if (p != null && p.position.isNotEmpty) _Row(label: tr('Position', 'Должность'), value: p.position),
      if (p != null && p.phone.isNotEmpty) _Row(label: tr('Phone', 'Телефон'), value: p.phone, onTap: () => _copy(p.phone)),
      if (p != null && p.email.isNotEmpty)
        _Row(
          label: tr('E-mail', 'Почта'),
          value: p.email,
          link: Uri(scheme: 'mailto', path: p.email),
          onTap: () => _copy(p.email),
        ),
      if (p != null && p.link.isNotEmpty) _Row(label: tr('Link', 'Ссылка'), value: p.link, link: Uri.tryParse(p.link), onTap: () => _copy(p.link)),
    ];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      mainAxisSize: MainAxisSize.min,
      children: [
        Center(
          child: MediaAvatar(
            account: widget.chat.account,
            photo: photo,
            label: title,
            icon: kind == pb.ChatKind.CHAT_KIND_CHANNEL ? Icons.campaign_outlined : null,
            color: providerColor(provider),
            size: 96,
            online: online && kind == pb.ChatKind.CHAT_KIND_DIRECT,
          ),
        ),
        const SizedBox(height: 16),
        SelectableText(title, textAlign: TextAlign.center, style: HmText.dialogHeading),
        const SizedBox(height: 4),
        Text(
          status.join(' · '),
          textAlign: TextAlign.center,
          style: HmText.small.copyWith(color: online && kind == pb.ChatKind.CHAT_KIND_DIRECT ? HmColors.ok : null),
        ),
        if (where.isNotEmpty) ...[const SizedBox(height: 2), Text(where, textAlign: TextAlign.center, style: HmText.footer)],
        const SizedBox(height: 24),
        if (_loading)
          const Padding(
            padding: EdgeInsets.all(16),
            child: Center(child: HmSpinner()),
          ),
        if (_error != null) HmBanner('${tr('Could not load the details', 'Не удалось загрузить сведения')}: $_error', tone: HmTone.warning),
        if (rows.isNotEmpty)
          HmCard(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                for (final (i, r) in rows.indexed) ...[if (i > 0) const Divider(height: 1), r],
              ],
            ),
          ),
        if (p != null && p.memberList.isNotEmpty) ...[
          const SizedBox(height: 24),
          HmSectionTitle(tr('Members', 'Участники')),
          const SizedBox(height: 8),
          for (final u in p.memberList)
            HmListItem(
              dense: true,
              leading: MediaAvatar(account: widget.chat.account, photo: u.avatarUrl, label: u.name, color: providerColor(provider), size: 28),
              title: Text(u.name.isNotEmpty ? u.name : u.id),
              trailing: u.username.isNotEmpty && u.username != u.name ? Text('@${u.username}', style: HmText.footer) : null,
            ),
          if (p.members > p.memberList.length)
            Padding(
              padding: const EdgeInsets.only(top: 8),
              child: Text(tr('and ${p.members - p.memberList.length} more', 'и ещё ${p.members - p.memberList.length}'), style: HmText.footer),
            ),
        ],
      ],
    );
  }
}

/// "members" or "subscribers" after a number.
String members(int n, {bool channel = false}) {
  if (lang.value != Lang.ru) return channel ? (n == 1 ? 'subscriber' : 'subscribers') : (n == 1 ? 'member' : 'members');
  final d = n % 10, h = n % 100;
  final one = d == 1 && h != 11, few = d >= 2 && d <= 4 && (h < 12 || h > 14);
  if (channel) return one ? 'подписчик' : (few ? 'подписчика' : 'подписчиков');
  return one ? 'участник' : (few ? 'участника' : 'участников');
}

/// One line of the details: a label above its value. A tap copies it; a
/// link opens instead.
class _Row extends StatelessWidget {
  const _Row({required this.label, this.value, this.child, this.link, this.onTap});

  final String label;
  final String? value;
  final Widget? child;
  final Uri? link;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final content = Padding(
      padding: const EdgeInsets.symmetric(vertical: 10),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: HmText.footer),
          const SizedBox(height: 2),
          child ??
              Text(
                value ?? '',
                style: link != null
                    ? HmText.body.copyWith(color: HmColors.accent, decoration: TextDecoration.underline, decorationColor: HmColors.accent)
                    : HmText.body,
              ),
        ],
      ),
    );
    if (child != null || (onTap == null && link == null)) return content;
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: GestureDetector(
        behavior: HitTestBehavior.opaque,
        onTap: link != null ? () => openLink(link!) : onTap,
        onLongPress: onTap,
        onSecondaryTap: onTap,
        child: content,
      ),
    );
  }
}
