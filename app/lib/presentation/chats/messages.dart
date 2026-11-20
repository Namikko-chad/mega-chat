import 'package:flutter/material.dart';

import '../../core/format.dart';
import '../../core/i18n.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;

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
    return main.deepCopy()
      ..attachments.clear()
      ..attachments.addAll([for (final p in live) ...p.attachments])
      ..text = [
        for (final p in live)
          if (p.text.isNotEmpty) p.text,
      ].join('\n');
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
