import 'dart:async';
import 'dart:math';

import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../data/rpc/core_client.dart';
import '../../data/rpc/endpoint.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../../core/i18n.dart';

enum Conn { connecting, ready, failed }

class CoreState extends Equatable {
  const CoreState({this.conn = Conn.connecting, this.state, this.error, this.session = 0});

  final Conn conn;
  final pb.State? state;
  final String? error; // connection error

  /// Bumps on every (re)connect: events may have been missed in between,
  /// so whoever caches core data reloads it.
  final int session;

  /// The last known vault state: a dropped connection keeps the screens
  /// (with a banner) until the core answers again.
  bool get unlocked => state?.unlocked ?? false;

  CoreState copyWith({Conn? conn, pb.State? state, String? error, int? session, bool clearError = false}) =>
      CoreState(conn: conn ?? this.conn, state: state ?? this.state, error: clearError ? null : error ?? this.error, session: session ?? this.session);

  @override
  List<Object?> get props => [conn, state, error, session];
}

/// The connection to the core and its one Watch stream. Screens and cubits
/// take live updates from [events]; the stream survives reconnects.
class CoreCubit extends Cubit<CoreState> {
  CoreCubit(this._launcher) : super(const CoreState());

  final CoreLauncher _launcher;
  CoreClient? _client;
  StreamSubscription<pb.Event>? _watch;
  Timer? _retry;
  int _failures = 0;
  final _events = StreamController<pb.Event>.broadcast();

  /// Every event of the core, across reconnects.
  Stream<pb.Event> get events => _events.stream;

  pb.MegaChatClient get api {
    final c = _client;
    if (c == null) throw StateError('not connected');
    return c.api;
  }

  Future<void> connect() async {
    _retry?.cancel();
    // The last error stays until the core answers: the screens show it while retrying.
    emit(state.copyWith(conn: Conn.connecting));
    try {
      await _watch?.cancel();
      await _client?.close();
      _client = await CoreClient.connect(_launcher);
      final st = await api.getState(pb.Empty());
      _failures = 0;
      _listen();
      emit(state.copyWith(conn: Conn.ready, state: st, session: state.session + 1, clearError: true));
    } catch (e) {
      if (isClosed) return;
      emit(state.copyWith(conn: Conn.failed, error: errorText(e)));
      _reconnectLater();
    }
  }

  void _listen() {
    _watch = api
        .watch(pb.Empty())
        .listen(
          (ev) {
            // State carries the CloudSync state too.
            if (ev.kind == pb.Event_Kind.KIND_STATE || ev.kind == pb.Event_Kind.KIND_CLOUDSYNC) unawaited(refreshState());
            _events.add(ev);
          },
          onError: (_) => _lost(),
          onDone: _lost,
          cancelOnError: true,
        );
  }

  // The stream broke (the other app instance whose core this one used
  // quit): reconnect, which starts the core here when none answers.
  void _lost() {
    if (isClosed) return;
    emit(state.copyWith(conn: Conn.failed, error: tr('No connection to the core', 'Нет связи с ядром')));
    _reconnectLater();
  }

  /// 1, 2, 4 … 30 s between attempts; [connect] by hand skips the wait.
  void _reconnectLater() {
    _retry?.cancel();
    final delay = Duration(seconds: min(30, 1 << min(_failures, 5)));
    _failures++;
    _retry = Timer(delay, () {
      if (!isClosed && state.conn == Conn.failed) connect();
    });
  }

  Future<void> refreshState() async {
    try {
      final st = await api.getState(pb.Empty());
      if (!isClosed) emit(state.copyWith(state: st));
    } catch (e) {
      if (!isClosed) emit(state.copyWith(error: errorText(e)));
    }
  }

  /// Applies a state returned by an unlock / create / lock call.
  void adopt(pb.State st) => emit(state.copyWith(state: st));

  /// Applies a CloudSync state returned by a CloudSync call.
  void adoptCloudSync(pb.CloudSyncState cs) => adopt((state.state ?? pb.State()).deepCopy()..cloudsync = cs);

  Future<void> lockVault() async => adopt(await api.lockVault(pb.Empty()));

  /// Moves the open vault to [path] (move), or opens the vault file at
  /// [path] from now on (locked only). Throws the core's error.
  Future<void> setVaultPath(String path, {bool move = false}) async =>
      adopt(await api.setVaultPath(pb.VaultPathRequest(path: path, move: move)));

  @override
  Future<void> close() async {
    _retry?.cancel();
    await _watch?.cancel();
    await _events.close();
    await _client?.close();
    return super.close();
  }
}
