import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/format.dart';
import '../../core/i18n.dart';
import '../../core/theme/typography.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../common/hm.dart';
import '../core/core_cubit.dart';

IconData attachmentIcon(pb.AttachmentKind k) => switch (k) {
  pb.AttachmentKind.ATTACHMENT_KIND_IMAGE => Icons.image_outlined,
  pb.AttachmentKind.ATTACHMENT_KIND_VIDEO => Icons.movie_outlined,
  pb.AttachmentKind.ATTACHMENT_KIND_AUDIO => Icons.audiotrack_outlined,
  pb.AttachmentKind.ATTACHMENT_KIND_VOICE => Icons.mic_none_outlined,
  pb.AttachmentKind.ATTACHMENT_KIND_STICKER => Icons.emoji_emotions_outlined,
  _ => Icons.insert_drive_file_outlined,
};

String attachmentKind(pb.AttachmentKind k) => switch (k) {
  pb.AttachmentKind.ATTACHMENT_KIND_IMAGE => tr('Photo', 'Фото'),
  pb.AttachmentKind.ATTACHMENT_KIND_VIDEO => tr('Video', 'Видео'),
  pb.AttachmentKind.ATTACHMENT_KIND_AUDIO => tr('Audio', 'Аудио'),
  pb.AttachmentKind.ATTACHMENT_KIND_VOICE => tr('Voice message', 'Голосовое сообщение'),
  pb.AttachmentKind.ATTACHMENT_KIND_STICKER => tr('Sticker', 'Стикер'),
  _ => tr('File', 'Файл'),
};

/// Name and size of an attachment; a nameless one shows its kind.
(String, String) attachmentText(pb.Attachment a) => (a.name.isNotEmpty ? a.name : attachmentKind(a.kind), a.size > 0 ? formatSize(a.size.toInt()) : '');

/// One line for the chat list and reply quotes.
String messagePreview(pb.Message m) {
  if (m.deleted) return tr('Message deleted', 'Сообщение удалено');
  if (m.text.isNotEmpty) return m.text.replaceAll(RegExp(r'\s+'), ' ');
  if (m.attachments.isNotEmpty) return attachmentText(m.attachments.first).$1;
  return '';
}

String senderName(pb.Message m) =>
    m.outgoing ? tr('You', 'Вы') : (m.sender.name.isNotEmpty ? m.sender.name : (m.sender.username.isNotEmpty ? m.sender.username : '?'));

/// Messages sent together as one album (an unbroken run with the same
/// `group`), shown as one message.
class MessageGroup {
  MessageGroup(this.parts) : message = parts.length == 1 ? parts.single : _merge(parts);

  /// Oldest first; usually one.
  final List<pb.Message> parts;

  /// What is shown: the part with the caption, carrying every attachment.
  /// Replies and reactions go to that part, as in Telegram.
  final pb.Message message;

  pb.Message get newest => parts.last;

  bool has(String id) => parts.any((p) => p.id == id);

  static pb.Message _merge(List<pb.Message> parts) {
    final live = parts.where((p) => !p.deleted).toList();
    if (live.isEmpty) return parts.last;
    final main =
        live.where((p) => p.text.isNotEmpty).firstOrNull ??
        live.where((p) => p.reactions.isNotEmpty || (p.hasThread() && p.thread.replies > 0)).firstOrNull ??
        live.first;
    // The captions are joined: each one's styles move with it.
    final texts = <String>[];
    final format = <pb.TextSpan>[];
    var at = 0;
    for (final p in live) {
      if (p.text.isEmpty) continue;
      for (final s in p.format) {
        format.add(s.deepCopy()..offset = s.offset + at);
      }
      texts.add(p.text);
      at += p.text.length + 1;
    }
    return main.deepCopy()
      ..attachments.clear()
      ..attachments.addAll([for (final p in live) ...p.attachments])
      ..text = texts.join('\n')
      ..format.clear()
      ..format.addAll(format);
  }
}

/// Folds each album in [msgs] (oldest first) into one [MessageGroup].
List<MessageGroup> groupAlbums(List<pb.Message> msgs) {
  final out = <MessageGroup>[];
  var run = <pb.Message>[];
  for (final m in msgs) {
    final last = run.lastOrNull;
    if (last != null && m.group.isNotEmpty && m.group == last.group && !m.service && m.sender.id == last.sender.id) {
      run.add(m);
      continue;
    }
    if (run.isNotEmpty) out.add(MessageGroup(run));
    run = [m];
  }
  if (run.isNotEmpty) out.add(MessageGroup(run));
  return out;
}

/// The author's own name (not "You"), for text that leaves the app.
String authorName(pb.Message m) => m.sender.name.isNotEmpty ? m.sender.name : (m.sender.username.isNotEmpty ? m.sender.username : senderName(m));

/// The text of a message with its attachments named, for copying.
String messageBody(pb.Message m) => [if (m.text.isNotEmpty) m.text, for (final a in m.attachments) '[${attachmentText(a).$1}]'].join('\n');

/// What copying messages puts on the clipboard: one message is its text;
/// several (oldest first) get their author and time above each.
String copyText(List<pb.Message> msgs, String Function(pb.Message) stamp) {
  final live = msgs.where((m) => !m.deleted && !m.service).toList();
  if (live.length == 1) return messageBody(live.single);
  return [for (final m in live) '${authorName(m)}, ${stamp(m)}\n${messageBody(m)}'].join('\n\n');
}

/// A forward sent as text (to another account, or where the messenger
/// cannot forward): the original author on top, then the text or
/// [fragment] of it; attachments only by name.
String forwardText(pb.Message m, {String? fragment}) {
  final from = m.forwarded && m.forwardedFrom.isNotEmpty ? m.forwardedFrom : authorName(m);
  return '${tr('Forwarded from', 'Переслано от')} $from:\n${fragment ?? messageBody(m)}';
}

/// What the composer shows when [m] is edited: its text with the formatting
/// written as Markdown.
String editSource(pb.Message m) => m.markdown.isNotEmpty ? m.markdown : m.text;

/// The account may change the text of [m] here: its own, not forwarded,
/// with text, and the messenger edits (CAPABILITY_EDIT).
bool canEditMessage(pb.Account? account, pb.Message m) =>
    (account?.capabilities ?? const []).contains(pb.Capability.CAPABILITY_EDIT) &&
    m.outgoing &&
    !m.deleted &&
    !m.service &&
    !m.forwarded &&
    m.id.isNotEmpty &&
    m.text.isNotEmpty;

/// The account's own message that it can delete (CAPABILITY_DELETE).
bool canDeleteMessage(pb.Account? account, pb.Message m) =>
    (account?.capabilities ?? const []).contains(pb.Capability.CAPABILITY_DELETE) &&
    m.outgoing &&
    !m.deleted &&
    !m.service &&
    m.id.isNotEmpty;

/// Deletes the account's own messages of one chat (an album: all its
/// parts) after asking how: for everyone, or only here when the messenger
/// can (CAPABILITY_DELETE_FOR_ME). [thread]: they are replies in it.
Future<void> deleteMessages(BuildContext context, pb.Account? account, List<pb.Message> msgs, {String thread = ''}) async {
  if (msgs.isEmpty) return;
  final forMe = (account?.capabilities ?? const []).contains(pb.Capability.CAPABILITY_DELETE_FOR_ME);
  final api = context.read<CoreCubit>().api;
  final v = await showHmModal<String>(
    context,
    title: tr('Delete the message?', 'Удалить сообщение?'),
    size: HmModalSize.sm,
    body: (_) => Text(
      forMe
          ? tr('Delete it for everyone in the chat, or only for you.', 'Удалить его у всех в чате или только у себя.')
          : tr('It is deleted for everyone in the chat.', 'Оно удалится у всех в чате.'),
      style: HmText.body.copyWith(height: 1.5),
    ),
    actions: (ctx) => [
      HmButton(tr('Cancel', 'Отмена'), variant: HmVariant.neutral, onPressed: () => Navigator.pop(ctx)),
      if (forMe) HmButton(tr('Only for me', 'Только у себя'), variant: HmVariant.neutral, onPressed: () => Navigator.pop(ctx, 'me')),
      HmButton(tr('For everyone', 'У всех'), variant: HmVariant.red, onPressed: () => Navigator.pop(ctx, 'all')),
    ],
  );
  if (v == null) return;
  try {
    await api.delete(
      pb.DeleteRequest(
        accountId: msgs.first.accountId,
        chatId: msgs.first.chatId,
        messageIds: [for (final m in msgs) m.id],
        forAll: v == 'all',
        threadId: thread,
      ),
    );
  } catch (e) {
    Toasts.error('${tr('Could not delete the message', 'Не удалось удалить сообщение')}: ${errorText(e)}');
  }
}
