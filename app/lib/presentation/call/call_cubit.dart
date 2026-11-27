import 'dart:async';
import 'dart:io';

import 'package:equatable/equatable.dart';
import 'package:fixnum/fixnum.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';
import 'package:media_kit/media_kit.dart';

import '../../core/desktop/desktop_window.dart';
import '../../core/i18n.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/parley.pbgrpc.dart' as pb;
import '../common/hm.dart';
import '../core/core_cubit.dart';

/// What the call screen shows.
class CallUi extends Equatable {
  const CallUi({this.call, this.muted = false, this.camera = false, this.speaker = false, this.minimized = false, this.remoteTrack = false, this.rev = 0});

  /// The call, live or just ended (shown for a moment); null: none.
  final pb.Call? call;
  final bool muted;
  final bool camera;
  final bool speaker;
  final bool minimized;

  /// Video arrives from the peer.
  final bool remoteTrack;
  final int rev;

  bool get live => call != null && call!.state != pb.CallState.CALL_STATE_ENDED;
  bool get incoming => call != null && !call!.outgoing && call!.state == pb.CallState.CALL_STATE_RINGING;

  CallUi copyWith({pb.Call? call, bool? muted, bool? camera, bool? speaker, bool? minimized, bool? remoteTrack}) => CallUi(
    call: call ?? this.call,
    muted: muted ?? this.muted,
    camera: camera ?? this.camera,
    speaker: speaker ?? this.speaker,
    minimized: minimized ?? this.minimized,
    remoteTrack: remoteTrack ?? this.remoteTrack,
    rev: rev + 1,
  );

  @override
  List<Object?> get props => [rev];
}

/// The one call of this device: its state from the core, and the WebRTC
/// peer connection that carries its audio and video. Messenger specifics
/// stay in the core: the app gets offers, answers, candidates and
/// data-channel messages and sends its own back.
class CallCubit extends Cubit<CallUi> {
  CallCubit(this._core) : super(const CallUi()) {
    _events = _core.events.listen(_onEvent);
    _coreState = _core.stream.listen(_onCore);
    _onCore(_core.state);
  }

  final CoreCubit _core;
  late final StreamSubscription<pb.Event> _events;
  late final StreamSubscription<CoreState> _coreState;
  int? _session;

  pb.ParleyClient get _api => _core.api;

  RTCPeerConnection? _pc;
  MediaStream? _local;
  MediaStreamTrack? _camera;
  RTCRtpSender? _videoSender;
  RTCDataChannel? _dc;
  final _toChannel = <String>[]; // data for the channel before it opens
  final _pendingCandidates = <RTCIceCandidate>[];
  bool _makingOffer = false;
  bool _ignoreOffer = false;
  bool _ready = false; // the peer connection is set up
  bool _negotiated = false; // a first offer/answer went through
  bool _connected = false;
  int _seq = 0; // last signal handled

  /// Signals and setup run one after the other.
  Future<void> _queue = Future.value();

  RTCVideoRenderer? localView;
  RTCVideoRenderer? remoteView;

  Player? _sound;
  String? _soundName;
  Timer? _clear;

  pb.Call? get _call => state.call;

  bool _isCall(String account, String id) => _call != null && _call!.accountId == account && _call!.id == id;

  void _run(Future<void> Function() step) {
    _queue = _queue.then((_) => step()).catchError((Object e, StackTrace st) {
      debugPrint('call: $e\n$st');
    });
  }

  // --- core events ---------------------------------------------------------

  void _onCore(CoreState s) {
    if (!s.unlocked) {
      _session = null;
      if (state.call != null) _run(() => _teardown(clear: true));
      return;
    }
    if (s.conn != Conn.ready || s.session == _session) return;
    _session = s.session;
    _run(_reload);
  }

  /// After a (re)connect: the live call, and the signals missed meanwhile.
  Future<void> _reload() async {
    final r = await _api.calls(pb.Empty());
    final mine = _call;
    final live = r.calls.where((c) => mine == null || (c.accountId == mine.accountId && c.id == mine.id)).firstOrNull;
    if (live == null) {
      if (mine != null && state.live) await _teardown(clear: true);
      return;
    }
    await _apply(live);
    if (mine != null) await _catchUp();
  }

  void _onEvent(pb.Event ev) {
    switch (ev.kind) {
      case pb.Event_Kind.KIND_CALL:
        final c = ev.call;
        if (_call == null || !state.live || _isCall(c.accountId, c.id)) _run(() => _apply(c));
      case pb.Event_Kind.KIND_CALL_SIGNAL:
        final c = ev.call;
        final sig = ev.signal;
        _run(() async {
          if (!_isCall(ev.accountId, c.id) || sig.seq <= _seq) return;
          if (sig.seq > _seq + 1) return _catchUp();
          _seq = sig.seq.toInt();
          await _handle(sig);
        });
      default:
    }
  }

  Future<void> _catchUp() async {
    final c = _call;
    if (c == null) return;
    final r = await _api.callSignals(pb.CallSignalsRequest(accountId: c.accountId, callId: c.id, after: Int64(_seq)));
    for (final s in r.signals) {
      if (s.seq <= _seq) continue;
      _seq = s.seq.toInt();
      await _handle(s);
    }
  }

  Future<void> _apply(pb.Call c) async {
    final fresh = !_isCall(c.accountId, c.id);
    if (fresh) {
      _clear?.cancel();
      if (state.call != null) await _teardown();
      if (c.state == pb.CallState.CALL_STATE_ENDED) return;
      emit(CallUi(call: c, camera: c.outgoing && c.video, speaker: c.video, rev: state.rev + 1));
      if (!c.outgoing && DesktopWindow.supported) unawaited(DesktopWindow.instance.show());
    } else {
      // The peer's camera, mic: as the core reports them.
      emit(state.copyWith(call: c));
    }
    switch (c.state) {
      case pb.CallState.CALL_STATE_RINGING:
        await _play(c.outgoing ? 'ringback' : 'ring');
      case pb.CallState.CALL_STATE_CONNECTING || pb.CallState.CALL_STATE_ACTIVE:
        await _play(null);
        if (_pc == null) {
          try {
            await _start();
          } catch (e) {
            // The core ends the call and logs why (else it waits for its timeout).
            _signal(pb.CallSignal(kind: pb.SignalKind.SIGNAL_KIND_FAILED, data: 'peer connection setup: $e'));
            rethrow;
          }
        }
      case pb.CallState.CALL_STATE_ENDED:
        await _ended(c);
      default:
        await _play(null);
    }
  }

  Future<void> _ended(pb.Call c) async {
    final wasLive = _pc != null || c.startedMs > 0;
    await _teardown();
    emit(state.copyWith(call: c, minimized: false));
    if (wasLive || c.end == pb.CallEnd.CALL_END_BUSY || c.end == pb.CallEnd.CALL_END_DECLINED) await _play('hangup', loop: false);
    if (c.end == pb.CallEnd.CALL_END_FAILED && c.error.isNotEmpty) Toasts.error(tr('The call failed', 'Звонок не удался'), c.error);
    _clear = Timer(const Duration(seconds: 2), () {
      if (!isClosed && state.call != null && !state.live) emit(const CallUi().copyWith());
    });
  }

  // --- the peer connection -------------------------------------------------

  Future<void> _start() async {
    final c = _call!;
    if (localView == null) {
      await (localView = RTCVideoRenderer()).initialize();
      await (remoteView = RTCVideoRenderer()).initialize();
    }
    final rv = remoteView!;
    final pc = _pc = await createPeerConnection({
      'iceServers': [
        for (final s in c.ice) {'urls': s.urls, if (s.username.isNotEmpty) 'username': s.username, if (s.credential.isNotEmpty) 'credential': s.credential},
      ],
      'iceTransportPolicy': c.relayOnly ? 'relay' : 'all',
      'sdpSemantics': 'unified-plan',
      'bundlePolicy': 'max-bundle',
      'rtcpMuxPolicy': 'require',
    });
    pc.onIceCandidate = (cand) {
      final c = _call;
      if (c == null || cand.candidate == null) return;
      _signal(pb.CallSignal(kind: pb.SignalKind.SIGNAL_KIND_CANDIDATE, sdp: cand.candidate, mid: cand.sdpMid ?? '', mline: cand.sdpMLineIndex ?? 0));
    };
    pc.onConnectionState = (s) {
      if (s == RTCPeerConnectionState.RTCPeerConnectionStateConnected && !_connected) {
        _connected = true;
        _signal(pb.CallSignal(kind: pb.SignalKind.SIGNAL_KIND_CONNECTED));
      } else if (s == RTCPeerConnectionState.RTCPeerConnectionStateFailed) {
        _signal(pb.CallSignal(kind: pb.SignalKind.SIGNAL_KIND_FAILED, data: 'connection failed'));
      }
    };
    pc.onTrack = (e) {
      if (e.track.kind != 'video') return;
      _run(() async {
        if (_pc != pc) return;
        if (e.streams.isNotEmpty) {
          rv.srcObject = e.streams.first;
        } else {
          final s = rv.srcObject ?? await createLocalMediaStream('remote');
          await s.addTrack(e.track);
          rv.srcObject = s;
        }
        emit(state.copyWith(remoteTrack: true));
      });
    };
    pc.onDataChannel = (dc) {
      if (dc.label == _call?.channel) _wire(dc);
    };
    pc.onRenegotiationNeeded = () {
      // The side that does not offer first waits for the first offer, as
      // the peer expects; later changes (a camera) are renegotiated.
      if (!_ready || (!_call!.offer && !_negotiated)) return;
      _run(_negotiate);
    };

    final local = _local = await navigator.mediaDevices.getUserMedia({'audio': true, 'video': false});
    for (final t in local.getAudioTracks()) {
      t.enabled = !state.muted;
      await pc.addTrack(t, local);
    }
    if (c.openChannel && c.channel.isNotEmpty) _wire(await pc.createDataChannel(c.channel, RTCDataChannelInit()));
    if (state.camera) await _cameraOn();
    if (!kIsWeb && (Platform.isAndroid || Platform.isIOS)) await Helper.setSpeakerphoneOn(state.speaker);
    _ready = true;
    if (c.offer) await _negotiate();
  }

  void _wire(RTCDataChannel dc) {
    _dc = dc;
    dc.onDataChannelState = (s) {
      if (s != RTCDataChannelState.RTCDataChannelOpen) return;
      for (final d in _toChannel) {
        dc.send(RTCDataChannelMessage(d));
      }
      _toChannel.clear();
      _sendMedia(); // the peer learns this side's state when the channel opens
    };
    dc.onMessage = (m) {
      if (!m.isBinary) _signal(pb.CallSignal(kind: pb.SignalKind.SIGNAL_KIND_DATA, data: m.text));
    };
  }

  Future<void> _negotiate() async {
    final pc = _pc;
    if (pc == null) return;
    try {
      _makingOffer = true;
      final offer = await pc.createOffer();
      if (await pc.getSignalingState() != RTCSignalingState.RTCSignalingStateStable) return;
      await pc.setLocalDescription(offer);
      _signal(pb.CallSignal(kind: pb.SignalKind.SIGNAL_KIND_OFFER, sdp: offer.sdp));
    } finally {
      _makingOffer = false;
    }
  }

  /// Perfect negotiation: a polite side rolls its offer back when offers
  /// collide, an impolite one ignores the other's.
  Future<void> _handle(pb.CallSignal s) async {
    final pc = _pc;
    switch (s.kind) {
      case pb.SignalKind.SIGNAL_KIND_DATA:
        final dc = _dc;
        if (dc != null && dc.state == RTCDataChannelState.RTCDataChannelOpen) {
          await dc.send(RTCDataChannelMessage(s.data));
        } else {
          _toChannel.add(s.data);
        }
        return;
      default:
    }
    if (pc == null) return;
    switch (s.kind) {
      case pb.SignalKind.SIGNAL_KIND_OFFER:
        final st = await pc.getSignalingState();
        final collision = _makingOffer || st != RTCSignalingState.RTCSignalingStateStable;
        _ignoreOffer = !_call!.polite && collision;
        if (_ignoreOffer) return;
        if (collision) await pc.setLocalDescription(RTCSessionDescription('', 'rollback'));
        await pc.setRemoteDescription(RTCSessionDescription(s.sdp, 'offer'));
        _negotiated = true;
        await _flushCandidates();
        final answer = await pc.createAnswer();
        await pc.setLocalDescription(answer);
        _signal(pb.CallSignal(kind: pb.SignalKind.SIGNAL_KIND_ANSWER, sdp: answer.sdp));
      case pb.SignalKind.SIGNAL_KIND_ANSWER:
        if (await pc.getSignalingState() != RTCSignalingState.RTCSignalingStateHaveLocalOffer) return;
        await pc.setRemoteDescription(RTCSessionDescription(s.sdp, 'answer'));
        _negotiated = true;
        await _flushCandidates();
      case pb.SignalKind.SIGNAL_KIND_CANDIDATE:
        if (s.sdp.isEmpty) return;
        final cand = RTCIceCandidate(s.sdp, s.mid, s.mline);
        if (await pc.getRemoteDescription() == null) {
          _pendingCandidates.add(cand);
          return;
        }
        try {
          await pc.addCandidate(cand);
        } catch (e) {
          if (!_ignoreOffer) debugPrint('call: candidate: $e');
        }
      default:
    }
  }

  Future<void> _flushCandidates() async {
    final pc = _pc;
    if (pc == null) return;
    final list = List.of(_pendingCandidates);
    _pendingCandidates.clear();
    for (final c in list) {
      try {
        await pc.addCandidate(c);
      } catch (_) {}
    }
  }

  void _signal(pb.CallSignal s) {
    final c = _call;
    if (c == null) return;
    unawaited(
      _api
          .callSignal(pb.CallSignalRequest(accountId: c.accountId, callId: c.id, signal: s))
          .then(
            (_) {},
            onError: (Object e) {
              debugPrint('call: signal ${s.kind}: $e');
            },
          ),
    );
  }

  void _sendMedia() {
    final c = _call;
    if (c == null || !state.live) return;
    unawaited(
      _api
          .setCallMedia(
            pb.CallMediaRequest(
              accountId: c.accountId,
              callId: c.id,
              media: pb.CallMedia(muted: state.muted, video: state.camera),
            ),
          )
          .then((_) {}, onError: (Object e) => debugPrint('call: media: $e')),
    );
  }

  Future<void> _cameraOn() async {
    final pc = _pc;
    if (pc == null) return;
    final s = await navigator.mediaDevices.getUserMedia({
      'audio': false,
      'video': {'facingMode': 'user', 'width': 1280, 'height': 720},
    });
    final t = _camera = s.getVideoTracks().first;
    localView?.srcObject = s;
    final sender = _videoSender;
    if (sender != null) {
      await sender.replaceTrack(t);
    } else {
      _videoSender = await pc.addTrack(t, _local ?? s);
    }
  }

  Future<void> _cameraOff() async {
    await _videoSender?.replaceTrack(null);
    final t = _camera;
    _camera = null;
    await t?.stop();
    localView?.srcObject = null;
  }

  Future<void> _teardown({bool clear = false}) async {
    await _play(null);
    _ready = _negotiated = _connected = _makingOffer = _ignoreOffer = false;
    _seq = 0;
    _toChannel.clear();
    _pendingCandidates.clear();
    final pc = _pc, dc = _dc, local = _local, cam = _camera;
    _pc = null;
    _dc = null;
    _local = null;
    _camera = null;
    _videoSender = null;
    try {
      await dc?.close();
      await pc?.close();
      for (final t in [...?local?.getTracks(), ?cam]) {
        await t.stop();
      }
      await local?.dispose();
      localView?.srcObject = null;
      remoteView?.srcObject = null;
    } catch (e) {
      debugPrint('call: teardown: $e');
    }
    if (clear && !isClosed) emit(const CallUi().copyWith());
  }

  // --- sounds --------------------------------------------------------------

  Future<void> _play(String? name, {bool loop = true}) async {
    if (name == _soundName && loop) return;
    _soundName = loop ? name : null;
    final p = _sound ??= Player();
    try {
      if (name == null) {
        await p.stop();
        return;
      }
      await p.setPlaylistMode(loop ? PlaylistMode.single : PlaylistMode.none);
      await p.open(Media('asset:///assets/sounds/$name.wav'));
    } catch (e) {
      debugPrint('call: sound: $e');
    }
  }

  // --- the user's actions --------------------------------------------------

  Future<void> start(String account, String chat, {bool video = false}) async {
    if (state.live) {
      Toasts.error(tr('You are in a call already', 'Вы уже в звонке'));
      return;
    }
    try {
      final c = await _api.startCall(pb.StartCallRequest(accountId: account, chatId: chat, video: video));
      _run(() => _apply(c));
    } catch (e) {
      Toasts.error(tr('Could not call', 'Не удалось позвонить'), errorText(e));
    }
  }

  Future<void> accept({bool video = false}) async {
    final c = _call;
    if (c == null) return;
    emit(state.copyWith(camera: video, speaker: video || state.speaker));
    await _play(null);
    try {
      await _api.acceptCall(pb.CallRequest(accountId: c.accountId, callId: c.id, video: video));
    } catch (e) {
      Toasts.error(tr('Could not answer', 'Не удалось ответить'), errorText(e));
    }
  }

  /// Hangs up, cancels, or declines a ringing call.
  Future<void> end() async {
    final c = _call;
    if (c == null) return;
    if (!state.live) {
      emit(const CallUi().copyWith());
      return;
    }
    final reason = state.incoming ? pb.CallEnd.CALL_END_DECLINED : pb.CallEnd.CALL_END_HANGUP;
    try {
      await _api.endCall(pb.EndCallRequest(accountId: c.accountId, callId: c.id, reason: reason));
    } catch (e) {
      // The core is gone or the call is: end it here.
      _run(
        () => _ended(
          c.deepCopy()
            ..state = pb.CallState.CALL_STATE_ENDED
            ..end = pb.CallEnd.CALL_END_HANGUP,
        ),
      );
    }
  }

  void toggleMute() {
    final muted = !state.muted;
    for (final t in [...?_local?.getAudioTracks()]) {
      t.enabled = !muted;
    }
    emit(state.copyWith(muted: muted));
    _sendMedia();
  }

  void toggleCamera() {
    final on = !state.camera;
    emit(state.copyWith(camera: on));
    _run(() async {
      try {
        on ? await _cameraOn() : await _cameraOff();
      } catch (e) {
        emit(state.copyWith(camera: false));
        Toasts.error(tr('No camera', 'Нет камеры'), '$e');
      }
      _sendMedia();
    });
  }

  Future<void> toggleSpeaker() async {
    final on = !state.speaker;
    emit(state.copyWith(speaker: on));
    await Helper.setSpeakerphoneOn(on);
  }

  void minimize(bool on) => emit(state.copyWith(minimized: on));

  @override
  Future<void> close() async {
    await _events.cancel();
    await _coreState.cancel();
    await _teardown();
    await _sound?.dispose();
    await localView?.dispose();
    await remoteView?.dispose();
    return super.close();
  }
}
