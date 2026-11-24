import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/i18n.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../common/hm.dart';
import '../core/messenger_cubit.dart';

/// The actions the chat menu offers for [c]: what its account supports
/// (Account.chat_actions), each in the direction that changes something.
/// Leaving is for groups and channels only.
List<HmMenuItem<pb.ChatAction>> chatActionItems(pb.Chat c, pb.Account? account) {
  final can = account?.chatActions ?? const <pb.ChatAction>[];
  final channel = c.kind == pb.ChatKind.CHAT_KIND_CHANNEL;
  return [
    if (c.pinned && can.contains(pb.ChatAction.CHAT_ACTION_UNPIN))
      (value: pb.ChatAction.CHAT_ACTION_UNPIN, label: tr('Unpin', 'Открепить'))
    else if (!c.pinned && can.contains(pb.ChatAction.CHAT_ACTION_PIN))
      (value: pb.ChatAction.CHAT_ACTION_PIN, label: tr('Pin', 'Закрепить')),
    if (c.muted && can.contains(pb.ChatAction.CHAT_ACTION_UNMUTE))
      (value: pb.ChatAction.CHAT_ACTION_UNMUTE, label: tr('Turn notifications on', 'Включить уведомления'))
    else if (!c.muted && can.contains(pb.ChatAction.CHAT_ACTION_MUTE))
      (value: pb.ChatAction.CHAT_ACTION_MUTE, label: tr('Mute', 'Отключить уведомления')),
    if (c.archived && can.contains(pb.ChatAction.CHAT_ACTION_UNARCHIVE))
      (value: pb.ChatAction.CHAT_ACTION_UNARCHIVE, label: tr('Move out of the archive', 'Вернуть из архива'))
    else if (!c.archived && can.contains(pb.ChatAction.CHAT_ACTION_ARCHIVE))
      (value: pb.ChatAction.CHAT_ACTION_ARCHIVE, label: tr('Archive', 'В архив')),
    if (!isDirect(c) && !c.saved && can.contains(pb.ChatAction.CHAT_ACTION_LEAVE))
      (value: pb.ChatAction.CHAT_ACTION_LEAVE, label: channel ? tr('Leave channel', 'Покинуть канал') : tr('Leave group', 'Выйти из группы')),
  ];
}

/// Opens the chat menu of [c] (at [position], or next to [anchor]) and does
/// what is picked. Leaving asks first.
Future<void> showChatMenu(BuildContext context, pb.Chat c, {Offset? position, BuildContext? anchor}) async {
  final messenger = context.read<MessengerCubit>();
  final items = chatActionItems(c, messenger.state.account(c.accountId));
  if (items.isEmpty) return;
  final action = await showContextMenu<pb.ChatAction>(context, title: chatTitle(c), items: items, position: position, anchor: anchor);
  if (action == null || !context.mounted) return;
  await runChatAction(context, c, action);
}

/// Does [action] to [c]; leaving asks first. Errors become a toast.
Future<void> runChatAction(BuildContext context, pb.Chat c, pb.ChatAction action) async {
  final messenger = context.read<MessengerCubit>();
  if (action == pb.ChatAction.CHAT_ACTION_LEAVE) {
    final channel = c.kind == pb.ChatKind.CHAT_KIND_CHANNEL;
    final ok = await confirm(
      context,
      title: channel ? tr('Leave the channel?', 'Покинуть канал?') : tr('Leave the group?', 'Выйти из группы?'),
      message: tr(
        '“${chatTitle(c)}” disappears from the chat list. You can come back only if the chat is public or someone adds you.',
        '«${chatTitle(c)}» пропадёт из списка чатов. Вернуться можно, только если чат открытый или вас снова добавят.',
      ),
      action: channel ? tr('Leave', 'Покинуть') : tr('Leave', 'Выйти'),
      destructive: true,
    );
    if (!ok) return;
  }
  try {
    await messenger.manageChat(chatKey(c), action);
    switch (action) {
      case pb.ChatAction.CHAT_ACTION_ARCHIVE:
        Toasts.success(tr('Moved to the archive', 'Чат перенесён в архив'));
      case pb.ChatAction.CHAT_ACTION_UNARCHIVE:
        Toasts.success(tr('Moved out of the archive', 'Чат возвращён из архива'));
      case pb.ChatAction.CHAT_ACTION_LEAVE:
        Toasts.success(tr('You left “${chatTitle(c)}”', 'Вы покинули «${chatTitle(c)}»'));
      default:
    }
  } catch (e) {
    Toasts.error(tr('Could not change the chat', 'Не удалось изменить чат'), errorText(e));
  }
}
