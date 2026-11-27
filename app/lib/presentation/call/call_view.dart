import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';

import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/layout.dart';
import '../../core/theme/typography.dart';
import '../../data/rpc/gen/parley.pbgrpc.dart' as pb;
import '../chats/media.dart';
import '../common/hm.dart';
import 'call_cubit.dart';

/// Over the whole app while there is a call: the call screen, or a bar on
/// top when it is minimised. Has its own Overlay: it sits above the
/// navigator, and tooltips need one.
class CallOverlay extends StatelessWidget {
  const CallOverlay({super.key});

  @override
  Widget build(BuildContext context) => BlocBuilder<CallCubit, CallUi>(
    buildWhen: (a, b) => (a.call == null) != (b.call == null) || a.minimized != b.minimized,
    builder: (context, s) {
      if (s.call == null) return const SizedBox.shrink();
      if (s.minimized) {
        // Above the status bar (desktop) and the tab bar (phones).
        return const Align(
          alignment: Alignment.bottomCenter,
          child: Padding(padding: EdgeInsets.only(bottom: 64), child: _CallBar()),
        );
      }
      return Positioned.fill(
        child: Overlay(initialEntries: [OverlayEntry(builder: (_) => const _CallScreen())]),
      );
    },
  );
}

String _status(CallUi s) {
  final c = s.call!;
  final video = c.video;
  return switch (c.state) {
    pb.CallState.CALL_STATE_WAITING => tr('Calling…', 'Вызов…'),
    pb.CallState.CALL_STATE_RINGING when c.outgoing => tr('Ringing…', 'Звонит…'),
    pb.CallState.CALL_STATE_RINGING => video ? tr('Incoming video call', 'Входящий видеозвонок') : tr('Incoming call', 'Входящий звонок'),
    pb.CallState.CALL_STATE_CONNECTING => tr('Connecting…', 'Соединение…'),
    pb.CallState.CALL_STATE_ACTIVE => '',
    pb.CallState.CALL_STATE_ENDED => switch (c.end) {
      pb.CallEnd.CALL_END_BUSY => tr('Busy', 'Занято'),
      pb.CallEnd.CALL_END_MISSED => c.outgoing ? tr('No answer', 'Нет ответа') : tr('Missed call', 'Пропущенный звонок'),
      pb.CallEnd.CALL_END_DECLINED => c.outgoing ? tr('Call declined', 'Звонок отклонён') : tr('Call ended', 'Звонок завершён'),
      pb.CallEnd.CALL_END_FAILED => tr('Call failed', 'Звонок не удался'),
      _ => tr('Call ended', 'Звонок завершён'),
    },
    _ => '',
  };
}

/// Time since the call became active, ticking.
class _Elapsed extends StatefulWidget {
  const _Elapsed({required this.startedMs, required this.style});

  final int startedMs;
  final TextStyle style;

  @override
  State<_Elapsed> createState() => _ElapsedState();
}

class _ElapsedState extends State<_Elapsed> {
  late final Timer _t;

  @override
  void initState() {
    super.initState();
    _t = Timer.periodic(const Duration(seconds: 1), (_) => setState(() {}));
  }

  @override
  void dispose() {
    _t.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final s = DateTime.now().difference(DateTime.fromMillisecondsSinceEpoch(widget.startedMs)).inSeconds.clamp(0, 1 << 30);
    String two(int v) => v.toString().padLeft(2, '0');
    final text = s >= 3600 ? '${s ~/ 3600}:${two(s % 3600 ~/ 60)}:${two(s % 60)}' : '${two(s ~/ 60)}:${two(s % 60)}';
    return Text(text, style: widget.style);
  }
}

Widget _statusText(CallUi s, TextStyle style) {
  final c = s.call!;
  if (c.state == pb.CallState.CALL_STATE_ACTIVE && c.startedMs > 0) {
    return _Elapsed(startedMs: c.startedMs.toInt(), style: style);
  }
  return Text(_status(s), style: style);
}

String _peerName(pb.Call c) => c.peer.name.isNotEmpty ? c.peer.name : tr('Unknown', 'Неизвестный');

class _CallBar extends StatelessWidget {
  const _CallBar();

  @override
  Widget build(BuildContext context) {
    final s = context.watch<CallCubit>().state;
    final c = s.call!;
    final cubit = context.read<CallCubit>();
    return SafeArea(
      child: Material(
        color: HmColors.tint(HmColors.ok, 0.9),
        borderRadius: BorderRadius.circular(8),
        child: InkWell(
          borderRadius: BorderRadius.circular(8),
          onTap: () => cubit.minimize(false),
          child: Padding(
            padding: const EdgeInsets.fromLTRB(14, 4, 4, 4),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(Icons.call, size: 16, color: HmColors.bg),
                const SizedBox(width: 8),
                ConstrainedBox(
                  constraints: const BoxConstraints(maxWidth: 220),
                  child: Text(
                    _peerName(c),
                    style: HmText.bodyStrong.copyWith(color: HmColors.bg),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                const SizedBox(width: 8),
                _statusText(s, HmText.footer.copyWith(color: HmColors.bg)),
                const SizedBox(width: 4),
                // No tooltip: the bar sits above the navigator's overlay.
                Semantics(
                  button: true,
                  label: tr('Hang up', 'Завершить'),
                  child: InkResponse(
                    onTap: cubit.end,
                    radius: 18,
                    child: const Padding(
                      padding: EdgeInsets.all(8),
                      child: Icon(Icons.call_end, size: 18, color: HmColors.bg),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _CallScreen extends StatelessWidget {
  const _CallScreen();

  @override
  Widget build(BuildContext context) {
    final s = context.watch<CallCubit>().state;
    final cubit = context.read<CallCubit>();
    final c = s.call;
    if (c == null) return const SizedBox.shrink();
    final hm = Hm.of(context);
    final active = c.state == pb.CallState.CALL_STATE_ACTIVE || c.state == pb.CallState.CALL_STATE_CONNECTING;
    final remoteVideo = active && s.remoteTrack && c.remote.video && cubit.remoteView != null;
    final localVideo = s.camera && cubit.localView != null && s.live;

    final buttons = <Widget>[
      if (s.incoming) ...[
        HmCallButton(icon: Icons.call_end, label: tr('Decline', 'Отклонить'), color: HmColors.error, onPressed: cubit.end),
        HmCallButton(icon: Icons.call, label: tr('Answer', 'Ответить'), color: HmColors.ok, onPressed: () => cubit.accept()),
        if (c.video) HmCallButton(icon: Icons.videocam, label: tr('With video', 'С видео'), color: HmColors.ok, onPressed: () => cubit.accept(video: true)),
      ] else if (s.live) ...[
        HmCallButton(
          icon: s.muted ? Icons.mic_off : Icons.mic,
          label: s.muted ? tr('Unmute', 'Включить') : tr('Mute', 'Микрофон'),
          active: s.muted,
          onPressed: cubit.toggleMute,
        ),
        HmCallButton(icon: s.camera ? Icons.videocam : Icons.videocam_off, label: tr('Camera', 'Камера'), active: s.camera, onPressed: cubit.toggleCamera),
        if (hm.mobile)
          HmCallButton(
            icon: s.speaker ? Icons.volume_up : Icons.phone_in_talk,
            label: tr('Speaker', 'Динамик'),
            active: s.speaker,
            onPressed: cubit.toggleSpeaker,
          ),
        HmCallButton(
          icon: Icons.call_end,
          label: c.state == pb.CallState.CALL_STATE_ACTIVE ? tr('Hang up', 'Завершить') : tr('Cancel', 'Отмена'),
          color: HmColors.error,
          onPressed: cubit.end,
        ),
      ],
    ];

    final peer = Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        MediaAvatar(account: c.accountId, photo: c.peer.avatarUrl, label: _peerName(c), size: hm.mobile ? 96 : 120),
        const SizedBox(height: 20),
        Text(_peerName(c), style: HmText.pageTitle, textAlign: TextAlign.center),
        const SizedBox(height: 8),
        _statusText(s, HmText.small),
      ],
    );

    return Material(
      color: HmColors.bg,
      child: SafeArea(
        child: Stack(
          children: [
            if (remoteVideo)
              Positioned.fill(child: RTCVideoView(cubit.remoteView!, objectFit: RTCVideoViewObjectFit.RTCVideoViewObjectFitContain))
            else
              Center(child: peer),
            if (remoteVideo)
              Positioned(
                left: 16,
                top: 12,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(_peerName(c), style: HmText.bodyStrong),
                    _statusText(s, HmText.footer),
                  ],
                ),
              ),
            Positioned(
              top: 8,
              right: 8,
              child: Row(
                children: [
                  if (c.verification.isNotEmpty && active)
                    Tooltip(
                      message: tr(
                        'If ${_peerName(c)} sees the same emoji, the call is encrypted end to end',
                        'Если у собеседника те же эмодзи, звонок зашифрован сквозным шифрованием',
                      ),
                      child: Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 8),
                        child: Text(c.verification.join(' '), style: const TextStyle(fontSize: 20)),
                      ),
                    ),
                  if (s.live && !s.incoming)
                    HmIconButton(icon: Icons.close_fullscreen, label: tr('Minimise', 'Свернуть'), onPressed: () => cubit.minimize(true)),
                ],
              ),
            ),
            if (localVideo)
              Positioned(
                right: 16,
                bottom: 140,
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(8),
                  child: Container(
                    width: hm.mobile ? 96 : 200,
                    height: hm.mobile ? 128 : 120,
                    color: HmColors.surface,
                    child: RTCVideoView(cubit.localView!, mirror: true, objectFit: RTCVideoViewObjectFit.RTCVideoViewObjectFitCover),
                  ),
                ),
              ),
            if (active && c.remote.muted)
              Positioned(
                left: 0,
                right: 0,
                bottom: 140,
                child: Center(
                  child: HmPill(tone: HmTone.warning, child: Text(tr('${_peerName(c)}: microphone off', '${_peerName(c)}: микрофон выключен'))),
                ),
              ),
            Positioned(
              left: 0,
              right: 0,
              bottom: 24,
              child: Wrap(alignment: WrapAlignment.center, spacing: hm.mobile ? 20 : 28, runSpacing: 12, children: buttons),
            ),
          ],
        ),
      ),
    );
  }
}
