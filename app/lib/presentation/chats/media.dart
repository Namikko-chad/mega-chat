import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:open_filex/open_filex.dart';

import '../../core/format.dart';
import '../../core/i18n.dart';
import '../../core/settings/app_settings.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/typography.dart';
import '../../data/local_files.dart';
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
    this.showOffline = false,
  });

  final String account;
  final String photo; // media ref; empty: initials only
  final String label;
  final IconData? icon;
  final Color color;
  final double size;
  final Widget? corner;
  final bool online; // see HmAvatar.online
  final bool showOffline; // see HmAvatar.showOffline

  @override
  Widget build(BuildContext context) {
    HmAvatar plain([ImageProvider? image]) => HmAvatar(
      label: label,
      icon: image == null ? icon : null,
      color: color,
      size: size,
      corner: corner,
      image: image,
      online: online,
      showOffline: showOffline,
    );
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
    if (a.ref.isEmpty) return HmAttachment(icon: attachmentIcon(a.kind), name: name, detail: detail);
    final row = HmAttachment(
      icon: attachmentIcon(a.kind),
      name: name,
      detail: detail,
      trailing: HmIconButton(
        dense: true,
        icon: Icons.download_outlined,
        label: canAskSavePath ? tr('Save as…', 'Сохранить как…') : tr('Save to Downloads', 'Сохранить в «Загрузки»'),
        onPressed: () => saveAttachment(context, account, a, ask: true),
      ),
    );
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: GestureDetector(
        onTap: a.kind == pb.AttachmentKind.ATTACHMENT_KIND_VIDEO
            ? () => showVideoPlayer(context, account, a, attachmentFileName(a))
            : () => openAttachment(context, account, a),
        onLongPress: () => saveAttachment(context, account, a),
        onSecondaryTap: () => saveAttachment(context, account, a),
        child: Tooltip(
          message: tr('Open · right click or hold to save to Downloads', 'Открыть · правый клик или удержание — сохранить в «Загрузки»'),
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

/// Saves an attachment to Downloads; with [ask] (desktops) the user picks
/// where.
Future<void> saveAttachment(BuildContext context, String account, pb.Attachment a, {bool ask = false}) async {
  final media = context.read<MessengerCubit>().media;
  final name = attachmentFileName(a);
  String? path;
  if (ask && canAskSavePath) {
    path = await askSavePath(name, directory: await MediaStore.downloadsDir());
    if (path == null) return; // cancelled
  }
  if (a.size > MediaStore.prefetchMax) Toasts.show(tr('Downloading…', 'Скачиваю…'), detail: name);
  try {
    final where = await media.save(account, a.ref, name, path: path, mime: a.mime);
    if (where == null) {
      Toasts.error(tr('Could not download the file', 'Не удалось скачать файл'));
    } else {
      Toasts.success(tr('Saved', 'Сохранено'), where);
    }
  } on MediaTooLarge {
    Toasts.error(tr('The file is too large', 'Файл слишком большой'));
  } catch (e) {
    Toasts.error(tr('Could not save the file', 'Не удалось сохранить файл'), '$e');
  }
}

/// The messages whose photos the picture viewer pages through: the loaded
/// history of the chat or thread around it.
class PhotoGallery extends InheritedWidget {
  const PhotoGallery({super.key, required this.messages, required super.child});

  /// Oldest first.
  final List<pb.Message> messages;

  @override
  bool updateShouldNotify(PhotoGallery old) => false; // read once, on opening
}

/// The photos of [messages] in order, without deleted messages.
List<pb.Attachment> galleryPhotos(Iterable<pb.Message> messages) => [
  for (final m in messages)
    if (!m.deleted)
      for (final a in m.attachments)
        if (a.kind == pb.AttachmentKind.ATTACHMENT_KIND_IMAGE && a.ref.isNotEmpty && a.thumb.isNotEmpty) a,
];

/// A picture over the whole window: zoom and pan, save, open elsewhere, and
/// paging through the other photos of the chat. The list is fixed on
/// opening, so that a new message does not move it.
Future<void> showImageViewer(BuildContext context, String account, pb.Attachment a) {
  final gallery = context.getInheritedWidgetOfExactType<PhotoGallery>();
  var photos = galleryPhotos(gallery?.messages ?? const []);
  var start = photos.indexWhere((p) => p.ref == a.ref);
  if (start < 0) {
    photos = [a];
    start = 0;
  }
  return showGeneralDialog<void>(
    context: context,
    barrierDismissible: true,
    barrierLabel: tr('Close', 'Закрыть'),
    barrierColor: HmColors.bezel.withValues(alpha: 0.92),
    pageBuilder: (dialogContext, _, _) => BlocProvider.value(
      value: context.read<MessengerCubit>(),
      child: _Viewer(account: account, photos: photos, start: start),
    ),
  );
}

class _Viewer extends StatefulWidget {
  const _Viewer({required this.account, required this.photos, required this.start});

  final String account;
  final List<pb.Attachment> photos;
  final int start;

  @override
  State<_Viewer> createState() => _ViewerState();
}

class _ViewerState extends State<_Viewer> {
  // Swiping is for touch screens: on desktops the wheel zooms the picture
  // and the arrows page.
  static final bool _swipes = defaultTargetPlatform == TargetPlatform.android || defaultTargetPlatform == TargetPlatform.iOS;

  late final PageController _pages = PageController(initialPage: widget.start);
  late int _index = widget.start;
  bool _zoomed = false;
  int _pointers = 0;

  @override
  void dispose() {
    _pages.dispose();
    super.dispose();
  }

  void _go(int delta) {
    final to = _index + delta;
    if (to < 0 || to >= widget.photos.length) return;
    _pages.animateToPage(to, duration: const Duration(milliseconds: 200), curve: Curves.easeOut);
  }

  void _touch(int delta) {
    final before = _pointers > 1;
    _pointers = (_pointers + delta).clamp(0, 10);
    // A second finger is a pinch: the page must not take the gesture.
    if (before != _pointers > 1) setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    final account = widget.account;
    final photos = widget.photos;
    final a = photos[_index];
    final many = photos.length > 1;
    void close() => Navigator.of(context).pop();
    return CallbackShortcuts(
      bindings: {const SingleActivator(LogicalKeyboardKey.arrowLeft): () => _go(-1), const SingleActivator(LogicalKeyboardKey.arrowRight): () => _go(1)},
      child: Focus(
        autofocus: true,
        child: SafeArea(
          child: Stack(
            children: [
              Positioned.fill(
                child: GestureDetector(
                  onTap: close,
                  child: Listener(
                    behavior: HitTestBehavior.translucent,
                    onPointerDown: (_) => _touch(1),
                    onPointerUp: (_) => _touch(-1),
                    onPointerCancel: (_) => _touch(-1),
                    child: PageView.builder(
                      controller: _pages,
                      physics: _swipes && !_zoomed && _pointers < 2 ? null : const NeverScrollableScrollPhysics(),
                      itemCount: photos.length,
                      onPageChanged: (i) => setState(() {
                        _index = i;
                        _zoomed = false;
                      }),
                      itemBuilder: (context, i) => _ViewerPage(
                        key: ValueKey(photos[i].ref),
                        account: account,
                        attachment: photos[i],
                        onZoom: (z) {
                          if (i == _index && z != _zoomed) setState(() => _zoomed = z);
                        },
                      ),
                    ),
                  ),
                ),
              ),
              if (many && _index > 0)
                Positioned(
                  left: 8,
                  top: 0,
                  bottom: 0,
                  child: Center(
                    child: HmIconButton(icon: Icons.chevron_left, label: tr('Previous', 'Предыдущее'), onPressed: () => _go(-1)),
                  ),
                ),
              if (many && _index < photos.length - 1)
                Positioned(
                  right: 8,
                  top: 0,
                  bottom: 0,
                  child: Center(
                    child: HmIconButton(icon: Icons.chevron_right, label: tr('Next', 'Следующее'), onPressed: () => _go(1)),
                  ),
                ),
              if (many) Positioned(top: 16, left: 16, child: Text('${_index + 1} / ${photos.length}', style: HmText.footer)),
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
                    if (canAskSavePath)
                      HmIconButton(
                        icon: Icons.save_as_outlined,
                        label: tr('Save as…', 'Сохранить как…'),
                        onPressed: () => saveAttachment(context, account, a, ask: true),
                      ),
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
        ),
      ),
    );
  }
}

/// One picture of the viewer; reports whether it is zoomed in.
class _ViewerPage extends StatefulWidget {
  const _ViewerPage({super.key, required this.account, required this.attachment, required this.onZoom});

  final String account;
  final pb.Attachment attachment;
  final ValueChanged<bool> onZoom;

  @override
  State<_ViewerPage> createState() => _ViewerPageState();
}

class _ViewerPageState extends State<_ViewerPage> {
  final _transform = TransformationController();
  bool _zoomed = false;

  @override
  void initState() {
    super.initState();
    _transform.addListener(_changed);
  }

  @override
  void dispose() {
    _transform.dispose();
    super.dispose();
  }

  void _changed() {
    final z = _transform.value.getMaxScaleOnAxis() > 1.01;
    if (z == _zoomed) return;
    setState(() => _zoomed = z);
    widget.onZoom(z);
  }

  @override
  Widget build(BuildContext context) {
    final a = widget.attachment;
    return MediaBytes(
      account: widget.account,
      ref: a.ref,
      maxBytes: 32 << 20,
      // The preview stays until the full picture arrives.
      builder: (context, full, failed) => MediaBytes(
        account: widget.account,
        ref: a.thumb,
        builder: (context, thumb, _) {
          final data = full ?? thumb;
          if (data == null) {
            return Center(child: failed ? Text(tr('Could not load the picture', 'Не удалось загрузить картинку'), style: HmText.small) : const HmSpinner());
          }
          return InteractiveViewer(
            transformationController: _transform,
            maxScale: 8,
            panEnabled: _zoomed, // not zoomed in, a drag turns the page
            child: Center(
              child: GestureDetector(
                onTap: () {}, // a tap on the picture does not close
                child: Image.memory(data, gaplessPlayback: true, filterQuality: FilterQuality.high),
              ),
            ),
          );
        },
      ),
    );
  }
}
