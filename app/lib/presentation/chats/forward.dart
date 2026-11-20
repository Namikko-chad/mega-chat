import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/i18n.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../accounts/providers.dart';
import '../common/hm.dart';
import '../core/core_cubit.dart';
import '../core/messenger_cubit.dart';
import 'messages.dart';

/// Forwards [parts] (oldest first; an album's parts are all of it) or a
/// [fragment] of the only one to a chat the user picks, of any account.
/// Inside the account a messenger with CAPABILITY_FORWARD forwards them
/// itself ([native]: the messages are in their chat's history, not in a
/// thread elsewhere); otherwise the text goes with its author on top.
/// True when it was forwarded.
Future<bool> forwardMessages(BuildContext context, List<pb.Message> parts, {String? fragment, bool native = true}) async {
  final live = parts.where((p) => !p.deleted && !p.service).toList();
  if (live.isEmpty) return false;
  final m = context.read<MessengerCubit>().state;
  final core = context.read<CoreCubit>();
  final to = await _pickChat(context, m);
  if (to == null) return false;
  final src = live.first;
  final caps = m.account(src.accountId)?.capabilities ?? const [];
  final inside = native && fragment == null && to.account == src.accountId && caps.contains(pb.Capability.CAPABILITY_FORWARD);
  final title = m.chats[to]?.title ?? '';
  try {
    if (inside) {
      await core.api.forward(pb.ForwardRequest(accountId: src.accountId, chatId: src.chatId, messageIds: [for (final p in live) p.id], toChatId: to.chat));
    } else {
      for (final g in groupAlbums(live)) {
        await core.api.send(
          pb.SendRequest(
            accountId: to.account,
            chatId: to.chat,
            text: forwardText(g.message, fragment: fragment),
          ),
        );
      }
    }
  } catch (e) {
    Toasts.error(tr('Not forwarded', 'Не переслано'), errorText(e));
    return false;
  }
  final named = !inside && fragment == null && live.any((p) => p.attachments.isNotEmpty);
  Toasts.success(
    '${tr('Forwarded to', 'Переслано в')} «$title»',
    named
        ? tr('Files went only by name: the messenger cannot forward them there.', 'Файлы ушли только названиями: туда мессенджер их переслать не может.')
        : null,
  );
  return true;
}

/// Every chat the user can write in, latest activity first.
Future<ChatKey?> _pickChat(BuildContext context, MessengerState m) {
  final chats = m.sorted.where((c) => !c.readOnly).toList();
  String where(pb.Chat c) {
    final a = m.account(c.accountId);
    final provider = a?.provider ?? '';
    return [if (a != null && a.name.isNotEmpty) a.name, providerName(provider, m.provider(provider))].where((s) => s.isNotEmpty).join(' · ');
  }

  final keys = {for (final c in chats) (account: c.accountId, chat: c.id): c};
  return pickOption<ChatKey>(
    context,
    title: tr('Forward to', 'Переслать в'),
    options: {for (final e in keys.entries) e.key: e.value.title},
    describe: (k) => where(keys[k]!),
  );
}
