import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:open_filex/open_filex.dart';

import '../../core/format.dart';
import '../../core/i18n.dart';
import '../../core/settings/app_settings.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/typography.dart';
import '../../data/media_store.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../common/hm.dart';
import '../core/messenger_cubit.dart';
import 'messages.dart';
import 'player.dart';

/// Loads a media ref of an account and builds with its bytes (null while
/// loading or when it failed).
class MediaBytes extends StatefulWidget {
  const MediaBytes({super.key, required this.account, required this.ref, required this.builder, this.maxBytes = 0});

  final String account;
  final String ref;
  final int maxBytes;
  final Widget Function(BuildContext context, Uint8List? data, bool failed) builder;

  @override
  State<MediaBytes> createState() => _MediaBytesState();
}

class _MediaBytesState extends State<MediaBytes> {
  Uint8List? _data;
  bool _failed = false;

  MediaStore get _media => context.read<MessengerCubit>().media;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void didUpdateWidget(MediaBytes old) {
    super.didUpdateWidget(old);
    if (old.ref != widget.ref || old.account != widget.account) _load();
  }

  void _load() {
    _failed = false;
    if (widget.ref.isEmpty) {
      _data = null;
      return;
    }
    _data = _media.peek(widget.account, widget.ref);
    if (_data != null) return;
    final ref = widget.ref;
    _media
        .load(widget.account, ref, maxBytes: widget.maxBytes)
        .then(
          (d) {
            if (!mounted || ref != widget.ref) return;
            setState(() {
              _data = d;
              _failed = d == null;
            });
          },
          onError: (_) {
            if (mounted && ref == widget.ref) setState(() => _failed = true);
          },
        );
  }

  @override
  Widget build(BuildContext context) => widget.builder(context, _data, _failed);
}

/// An avatar with the photo of a user or chat when there is one.
class MediaAvatar extends StatelessWidget {
  const MediaAvatar({
    super.key,
    required this.account,
    required this.photo,
    required this.label,
    this.icon,
    this.color = HmColors.accent,
    this.size = 36,
    this.corner,
    this.online = false,
  });

  final String account;
  final String photo; // media ref; empty: initials only
  final String label;
  final IconData? icon;
  final Color color;
  final double size;
  final Widget? corner;
  final bool online; // see HmAvatar.online

  @override
  Widget build(BuildContext context) {
    HmAvatar plain([ImageProvider? image]) =>
        HmAvatar(label: label, icon: image == null ? icon : null, color: color, size: size, corner: corner, image: image, online: online);
    if (photo.isEmpty) return plain();
    return MediaBytes(account: account, ref: photo, maxBytes: 4 << 20, builder: (context, data, _) => plain(data == null ? null : MemoryImage(data)));
  }
}

bool _visual(pb.Attachment a) =>
    a.thumb.isNotEmpty &&
    (a.kind == pb.AttachmentKind.ATTACHMENT_KIND_IMAGE ||
        a.kind == pb.AttachmentKind.ATTACHMENT_KIND_VIDEO ||
        a.kind == pb.AttachmentKind.ATTACHMENT_KIND_STICKER);

String formatDuration(int ms) {
  final s = ms ~/ 1000;
  String two(int v) => v.toString().padLeft(2, '0');
  return s >= 3600 ? '${s ~/ 3600}:${two(s % 3600 ~/ 60)}:${two(s % 60)}' : '${s ~/ 60}:${two(s % 60)}';
}

/// The attachments of a message: pictures and videos as previews, other
/// files as rows; a tap opens them.
class MessageAttachments extends StatelessWidget {
  const MessageAttachments({super.key, required this.message, this.maxWidth = 320});

  final pb.Message message;
  final double maxWidth;

  @override
  Widget build(BuildContext context) {
    final m = message;
    // An album: its pictures and videos as a grid, two to a row; with an
    // odd number the first one takes a whole row.
    final grid = m.attachments.where((a) => !playable(a) && _visual(a) && a.kind != pb.AttachmentKind.ATTACHMENT_KIND_STICKER).toList();
    if (grid.length < 2) grid.clear();
    const gap = 4.0;
    final cell = ((maxWidth - gap) / 2).floorToDouble();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        if (grid.isNotEmpty)
          Padding(
            padding: const EdgeInsets.only(top: 4, bottom: 2),
            child: SizedBox(
              width: cell * 2 + gap,
              child: Wrap(
                spacing: gap,
                runSpacing: gap,
                children: [
                  for (final (i, a) in grid.indexed)
                    i == 0 && grid.length.isOdd
                        ? _Preview(account: m.accountId, attachment: a, maxWidth: cell * 2 + gap, cell: (cell * 2 + gap, cell))
                        : _Preview(account: m.accountId, attachment: a, maxWidth: cell, cell: (cell, cell)),
                ],
              ),
            ),
          ),
        for (final a in m.attachments)
          if (grid.contains(a))
            const SizedBox.shrink()
          else if (playable(a))
            AudioAttachment(account: m.accountId, attachment: a, name: attachmentFileName(a), width: maxWidth.clamp(200, 320))
          else if (_visual(a))
            Padding(
              padding: const EdgeInsets.only(top: 4, bottom: 2),
              child: _Preview(account: m.accountId, attachment: a, maxWidth: maxWidth),
            )
          else
            _FileRow(account: m.accountId, attachment: a),
      ],
    );
  }
}

/// A picture, video or sticker preview. What loads without a tap follows
/// the settings (Settings → pictures, videos), as in Telegram.
class _Preview extends StatefulWidget {
  const _Preview({required this.account, required this.attachment, required this.maxWidth, this.cell});

  final String account;
  final pb.Attachment attachment;
  final double maxWidth;

  /// A fixed box in an album grid (the picture is cropped to fill it).
  final (double, double)? cell;

  @override
  State<_Preview> createState() => _PreviewState();
}

class _PreviewState extends State<_Preview> {
  bool _asked = false; // the user tapped to load it

  @override
  Widget build(BuildContext context) {
    final account = widget.account;
    final a = widget.attachment;
    final cell = widget.cell;
    final sticker = a.kind == pb.AttachmentKind.ATTACHMENT_KIND_STICKER;
    final video = a.kind == pb.AttachmentKind.ATTACHMENT_KIND_VIDEO;
    final side = sticker ? 160.0 : widget.maxWidth;
    final media = context.read<MessengerCubit>().media;
    final settings = AppSettings.instance;
    final auto = sticker || (video ? settings.autoVideo != VideoAuto.none : settings.autoPhotos);
    final badge = video && a.durationMs > 0 ? formatDuration(a.durationMs) : null;
    if (!auto && !_asked && media.peek(account, a.thumb) == null) {
      return HmPhoto(
        image: null,
        download: true,
        width: cell?.$1.round() ?? a.width,
        height: cell?.$2.round() ?? a.height,
        maxWidth: cell?.$1 ?? side,
        maxHeight: cell?.$2 ?? 360,
        badge: badge ?? (a.size > 0 ? formatSize(a.size.toInt()) : null),
        onTap: () => setState(() => _asked = true),
      );
    }
    if (video && settings.autoVideo == VideoAuto.full) media.prefetch(account, a.ref, size: a.size.toInt());
    return MediaBytes(
      account: account,
      ref: a.thumb,
      maxBytes: 8 << 20,
      builder: (context, data, failed) => HmPhoto(
        image: data == null ? null : MemoryImage(data),
        failed: failed,
        width: cell?.$1.round() ?? a.width,
        height: cell?.$2.round() ?? a.height,
        maxWidth: cell?.$1 ?? side,
        maxHeight: cell?.$2 ?? (sticker ? 160 : 360),
        plain: sticker,
        play: video,
        badge: badge,
        onTap: sticker || a.ref.isEmpty
            ? null
            : video
            ? () => showVideoPlayer(context, account, a, attachmentFileName(a))
            : () => showImageViewer(context, account, a),
      ),
    );
  }
}

class _FileRow extends StatelessWidget {
  const _FileRow({required this.account, required this.attachment});

  final String account;
  final pb.Attachment attachment;

  @override
  Widget build(BuildContext context) {
    final a = attachment;
    final (name, size) = attachmentText(a);
    final detail = [if (a.durationMs > 0) formatDuration(a.durationMs), if (size.isNotEmpty) size].join(' · ');
    final row = HmAttachment(icon: attachmentIcon(a.kind), name: name, detail: detail);
    if (a.ref.isEmpty) return row;
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: GestureDetector(
        onTap: a.kind == pb.AttachmentKind.ATTACHMENT_KIND_VIDEO
            ? () => showVideoPlayer(context, account, a, attachmentFileName(a))
            : () => openAttachment(context, account, a),
        onLongPress: () => saveAttachment(context, account, a),
        onSecondaryTap: () => saveAttachment(context, account, a),
        child: Tooltip(
          message: tr('Open · right click or hold to save', 'Открыть · правый клик или удержание — сохранить'),
          waitDuration: const Duration(seconds: 1),
          child: row,
        ),
      ),
    );
  }
}

/// A file name for an attachment that has none of its own.
String attachmentFileName(pb.Attachment a) {
  if (a.name.isNotEmpty) return a.name;
  final ext = switch (a.mime) {
    'image/jpeg' => '.jpg',
    'image/png' => '.png',
    'image/webp' => '.webp',
    'image/gif' => '.gif',
    'video/mp4' => '.mp4',
    'audio/ogg' => '.ogg',
    'audio/mpeg' => '.mp3',
    'application/pdf' => '.pdf',
    _ => '',
  };
  return '${attachmentKind(a.kind)}$ext';
}

/// Downloads an attachment and opens it with the system's app.
Future<void> openAttachment(BuildContext context, String account, pb.Attachment a) async {
  final media = context.read<MessengerCubit>().media;
  Toasts.show(tr('Opening…', 'Открываю…'), detail: attachmentFileName(a));
  try {
    final f = await media.toFile(account, a.ref, attachmentFileName(a));
    if (f == null) {
      Toasts.error(tr('Could not download the file', 'Не удалось скачать файл'));
      return;
    }
    final r = await OpenFilex.open(f.path, type: a.mime.isEmpty ? null : a.mime);
    if (r.type != ResultType.done) Toasts.error(tr('No app to open it', 'Нет приложения, чтобы открыть'), r.message);
  } on MediaTooLarge {
    Toasts.error(tr('The file is too large', 'Файл слишком большой'));
  }
}

/// Saves an attachment to Downloads.
Future<void> saveAttachment(BuildContext context, String account, pb.Attachment a) async {
  final media = context.read<MessengerCubit>().media;
  try {
    final f = await media.save(account, a.ref, attachmentFileName(a));
    if (f == null) {
      Toasts.error(tr('Could not download the file', 'Не удалось скачать файл'));
    } else {
      Toasts.success(tr('Saved', 'Сохранено'), f.path);
    }
  } on MediaTooLarge {
    Toasts.error(tr('The file is too large', 'Файл слишком большой'));
  }
}

/// A picture over the whole window: zoom and pan, save, open elsewhere.
Future<void> showImageViewer(BuildContext context, String account, pb.Attachment a) => showGeneralDialog<void>(
  context: context,
  barrierDismissible: true,
  barrierLabel: tr('Close', 'Закрыть'),
  barrierColor: HmColors.bezel.withValues(alpha: 0.92),
  pageBuilder: (dialogContext, _, _) => BlocProvider.value(
    value: context.read<MessengerCubit>(),
    child: _Viewer(account: account, attachment: a),
  ),
);

class _Viewer extends StatelessWidget {
  const _Viewer({required this.account, required this.attachment});

  final String account;
  final pb.Attachment attachment;

  @override
  Widget build(BuildContext context) {
    final a = attachment;
    void close() => Navigator.of(context).pop();
    return SafeArea(
      child: Stack(
        children: [
          Positioned.fill(
            child: GestureDetector(
              onTap: close,
              child: MediaBytes(
                account: account,
                ref: a.ref,
                maxBytes: 32 << 20,
                // The preview stays until the full picture arrives.
                builder: (context, full, failed) => MediaBytes(
                  account: account,
                  ref: a.thumb,
                  builder: (context, thumb, _) {
                    final data = full ?? thumb;
                    if (data == null) {
                      return Center(
                        child: failed ? Text(tr('Could not load the picture', 'Не удалось загрузить картинку'), style: HmText.small) : const HmSpinner(),
                      );
                    }
                    return InteractiveViewer(
                      maxScale: 8,
                      child: Center(
                        child: GestureDetector(
                          onTap: () {}, // a tap on the picture does not close
                          child: Image.memory(data, gaplessPlayback: true, filterQuality: FilterQuality.high),
                        ),
                      ),
                    );
                  },
                ),
              ),
            ),
          ),
          Positioned(
            top: 8,
            right: 8,
            child: Row(
              children: [
                if (a.size > 0)
                  Padding(
                    padding: const EdgeInsets.only(right: 8),
                    child: Text(formatSize(a.size.toInt()), style: HmText.footer),
                  ),
                HmIconButton(icon: Icons.download_outlined, label: tr('Save', 'Сохранить'), onPressed: () => saveAttachment(context, account, a)),
                HmIconButton(
                  icon: Icons.open_in_new,
                  label: tr('Open in another app', 'Открыть в другом приложении'),
                  onPressed: () => openAttachment(context, account, a),
                ),
                HmIconButton(icon: Icons.close, label: tr('Close', 'Закрыть'), onPressed: close),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
