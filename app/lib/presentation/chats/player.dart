import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:media_kit/media_kit.dart';
import 'package:media_kit_video/media_kit_video.dart';

import '../../core/i18n.dart';
import '../../core/settings/app_settings.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/typography.dart';
import '../../data/media_store.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../common/hm.dart';
import '../core/messenger_cubit.dart';
import 'media.dart';

/// Voice messages and audio files: one plays at a time, in any chat or
/// thread, and keeps playing while the user scrolls or switches chats.
class AudioPlayback extends ChangeNotifier {
  AudioPlayback._();

  static final instance = AudioPlayback._();

  Player? _player;
  final _subs = <StreamSubscription<Object?>>[];

  /// account + ref of the current track.
  String? current;
  bool loading = false;
  bool playing = false;
  Duration position = Duration.zero;
  Duration duration = Duration.zero;

  static String key(String account, String ref) => '$account\u0000$ref';

  Player get _p {
    final existing = _player;
    if (existing != null) return existing;
    final p = Player();
    _subs.addAll([
      p.stream.playing.listen((v) {
        playing = v;
        notifyListeners();
      }),
      p.stream.position.listen((v) {
        position = v;
        notifyListeners();
      }),
      p.stream.duration.listen((v) {
        duration = v;
        notifyListeners();
      }),
      p.stream.completed.listen((done) {
        if (!done) return;
        playing = false;
        position = Duration.zero;
        unawaited(p.pause());
        unawaited(p.seek(Duration.zero));
        notifyListeners();
      }),
    ]);
    return _player = p;
  }

  /// Plays or pauses [a]; another track stops.
  Future<void> toggle(MediaStore media, String account, pb.Attachment a, String name) async {
    final k = key(account, a.ref);
    if (k == current && !loading) {
      await _p.playOrPause();
      return;
    }
    current = k;
    loading = true;
    playing = false;
    position = Duration.zero;
    duration = Duration(milliseconds: a.durationMs);
    notifyListeners();
    try {
      await _player?.stop();
      final f = await media.toFile(account, a.ref, name);
      if (current != k) return; // another track was chosen meanwhile
      if (f == null) {
        Toasts.error(tr('Could not download the file', 'Не удалось скачать файл'));
        current = null;
        return;
      }
      await _p.open(Media(f.path));
    } on MediaTooLarge {
      Toasts.error(tr('The file is too large', 'Файл слишком большой'));
      current = null;
    } catch (e) {
      Toasts.error(tr('Could not play it', 'Не удалось воспроизвести'), '$e');
      current = null;
    } finally {
      if (current == null || current == k) {
        loading = false;
        notifyListeners();
      }
    }
  }

  Future<void> seek(String k, double fraction) async {
    if (k != current || duration == Duration.zero) return;
    await _p.seek(duration * fraction);
  }

  Future<void> pause() async => _player?.pause();

  /// Stops and forgets the track (the vault was locked).
  Future<void> stop() async {
    current = null;
    loading = playing = false;
    position = duration = Duration.zero;
    await _player?.stop();
    notifyListeners();
  }
}

bool playable(pb.Attachment a) => a.ref.isNotEmpty && (a.kind == pb.AttachmentKind.ATTACHMENT_KIND_AUDIO || a.kind == pb.AttachmentKind.ATTACHMENT_KIND_VOICE);

/// An audio attachment in a message, played in place.
class AudioAttachment extends StatelessWidget {
  const AudioAttachment({super.key, required this.account, required this.attachment, required this.name, this.width = 280});

  final String account;
  final pb.Attachment attachment;
  final String name;
  final double width;

  @override
  Widget build(BuildContext context) {
    final a = attachment;
    final k = AudioPlayback.key(account, a.ref);
    final voice = a.kind == pb.AttachmentKind.ATTACHMENT_KIND_VOICE;
    final title = voice ? tr('Voice message', 'Голосовое сообщение') : name;
    final media = context.read<MessengerCubit>().media;
    // Audio loads only when played, unless the settings say ahead.
    if (AppSettings.instance.autoAudio) media.prefetch(account, a.ref, size: a.size.toInt());
    return ListenableBuilder(
      listenable: AudioPlayback.instance,
      builder: (context, _) {
        final p = AudioPlayback.instance;
        final mine = p.current == k;
        final total = mine && p.duration > Duration.zero ? p.duration : Duration(milliseconds: a.durationMs);
        final pos = mine ? p.position : Duration.zero;
        final time = mine && (p.playing || pos > Duration.zero)
            ? '${formatDuration(pos.inMilliseconds)} / ${formatDuration(total.inMilliseconds)}'
            : total > Duration.zero
            ? formatDuration(total.inMilliseconds)
            : '';
        return GestureDetector(
          onSecondaryTap: () => saveAttachment(context, account, a),
          onLongPress: () => saveAttachment(context, account, a),
          child: HmAudio(
            title: title,
            voice: voice,
            width: width,
            playing: mine && p.playing,
            loading: mine && p.loading,
            position: total.inMilliseconds > 0 ? pos.inMilliseconds / total.inMilliseconds : 0,
            time: time,
            onToggle: () => AudioPlayback.instance.toggle(media, account, a, name),
            onSeek: mine ? (f) => AudioPlayback.instance.seek(k, f) : null,
          ),
        );
      },
    );
  }
}

/// A video over the whole window, played here.
Future<void> showVideoPlayer(BuildContext context, String account, pb.Attachment a, String name) {
  unawaited(AudioPlayback.instance.pause());
  return showGeneralDialog<void>(
    context: context,
    barrierDismissible: true,
    barrierLabel: tr('Close', 'Закрыть'),
    barrierColor: HmColors.bezel.withValues(alpha: 0.92),
    pageBuilder: (dialogContext, _, _) => BlocProvider.value(
      value: context.read<MessengerCubit>(),
      child: _VideoViewer(account: account, attachment: a, name: name),
    ),
  );
}

class _VideoViewer extends StatefulWidget {
  const _VideoViewer({required this.account, required this.attachment, required this.name});

  final String account;
  final pb.Attachment attachment;
  final String name;

  @override
  State<_VideoViewer> createState() => _VideoViewerState();
}

class _VideoViewerState extends State<_VideoViewer> {
  final _player = Player();
  late final _video = VideoController(_player);
  String? _error;
  bool _ready = false;

  @override
  void initState() {
    super.initState();
    unawaited(_open());
  }

  Future<void> _open() async {
    final media = context.read<MessengerCubit>().media;
    try {
      final f = await media.toFile(widget.account, widget.attachment.ref, widget.name);
      if (!mounted) return;
      if (f == null) {
        setState(() => _error = tr('Could not download the video', 'Не удалось скачать видео'));
        return;
      }
      await _player.open(Media(f.path));
      if (mounted) setState(() => _ready = true);
    } on MediaTooLarge {
      if (mounted) setState(() => _error = tr('The file is too large', 'Файл слишком большой'));
    } catch (e) {
      if (mounted) setState(() => _error = '${tr('Could not play it', 'Не удалось воспроизвести')}: $e');
    }
  }

  @override
  void dispose() {
    unawaited(_player.dispose());
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final a = widget.attachment;
    void close() => Navigator.of(context).pop();
    Widget body;
    if (_error != null) {
      body = Center(child: Text(_error!, style: HmText.small));
    } else if (!_ready) {
      // The preview while the file downloads.
      body = Stack(
        alignment: Alignment.center,
        children: [
          if (a.thumb.isNotEmpty)
            MediaBytes(
              account: widget.account,
              ref: a.thumb,
              builder: (context, data, _) => data == null ? const SizedBox() : Image.memory(data, fit: BoxFit.contain, gaplessPlayback: true),
            ),
          const HmSpinner(),
        ],
      );
    } else {
      body = Video(controller: _video, controls: AdaptiveVideoControls, fill: HmColors.bezel);
    }
    return SafeArea(
      child: Material(
        type: MaterialType.transparency,
        child: Stack(
          children: [
            Positioned.fill(child: body),
            Positioned(
              top: 8,
              right: 8,
              child: Row(
                children: [
                  HmIconButton(icon: Icons.download_outlined, label: tr('Save', 'Сохранить'), onPressed: () => saveAttachment(context, widget.account, a)),
                  HmIconButton(
                    icon: Icons.open_in_new,
                    label: tr('Open in another app', 'Открыть в другом приложении'),
                    onPressed: () {
                      unawaited(_player.pause());
                      unawaited(openAttachment(context, widget.account, a));
                    },
                  ),
                  HmIconButton(icon: Icons.close, label: tr('Close', 'Закрыть'), onPressed: close),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
