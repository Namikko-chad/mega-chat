import 'dart:convert';
import 'dart:ffi';
import 'dart:io';

import 'package:ffi/ffi.dart';

/// The desktop core inside this process (Go, `desktopbind`): linked into
/// the executable on Linux, parley_core.dll next to it on Windows.
class NativeCore {
  NativeCore._(DynamicLibrary lib)
    : _start = lib.lookupFunction<Pointer<Uint8> Function(), Pointer<Uint8> Function()>('ParleyStart'),
      _stop = lib.lookupFunction<Void Function(), void Function()>('ParleyStop'),
      _free = lib.lookupFunction<Void Function(Pointer<Uint8>), void Function(Pointer<Uint8>)>('ParleyFree'),
      _notify = lib.lookupFunction<Void Function(Pointer<Utf8>), void Function(Pointer<Utf8>)>('ParleyNotify'),
      _notifyClose = lib.lookupFunction<Void Function(Pointer<Utf8>), void Function(Pointer<Utf8>)>('ParleyNotifyClose'),
      _notifyClicks = lib.lookupFunction<Void Function(Pointer<NativeFunction<_ClickFn>>), void Function(Pointer<NativeFunction<_ClickFn>>)>(
        'ParleyNotifyClicks',
      );

  static NativeCore? _instance;

  static NativeCore get instance => _instance ??= NativeCore._(Platform.isWindows ? DynamicLibrary.open('parley_core.dll') : DynamicLibrary.process());

  final Pointer<Uint8> Function() _start;
  final void Function() _stop;
  final void Function(Pointer<Uint8>) _free;
  final void Function(Pointer<Utf8>) _notify;
  final void Function(Pointer<Utf8>) _notifyClose;
  final void Function(Pointer<NativeFunction<_ClickFn>>) _notifyClicks;
  NativeCallable<_ClickFn>? _clicks;
  bool _started = false;

  /// The core runs in this process (not in another instance of the app).
  bool get owner => _started;

  /// Shows a desktop notification (D-Bus on Linux, a toast on Windows);
  /// a newer one with the same [tag] replaces it.
  void notify({required String tag, required String title, required String body, String? icon, String? entry}) =>
      _withString(jsonEncode({'tag': tag, 'title': title, 'body': body, 'icon': ?icon, 'entry': ?entry}), _notify);

  /// Takes away the notification with this [tag].
  void closeNotification(String tag) => _withString(tag, _notifyClose);

  /// [onClick] gets the tag of a clicked notification (Linux).
  void onNotificationClick(void Function(String tag) onClick) {
    _clicks?.close();
    final cb = NativeCallable<_ClickFn>.listener((Pointer<Utf8> p) {
      final tag = p.toDartString();
      _free(p.cast());
      onClick(tag);
    });
    _clicks = cb;
    _notifyClicks(cb.nativeFunction);
  }

  static void _withString(String s, void Function(Pointer<Utf8>) f) {
    final p = s.toNativeUtf8();
    try {
      f(p);
    } finally {
      malloc.free(p);
    }
  }

  /// Starts the core (or returns the running one): {"address", "token"},
  /// with "external" when another instance of the app runs it, or {"error"}.
  Map<String, dynamic> start() {
    final p = _start();
    try {
      var n = 0;
      while (p[n] != 0) {
        n++;
      }
      final r = jsonDecode(utf8.decode(p.asTypedList(n))) as Map<String, dynamic>;
      if (r['error'] == null && r['external'] != true) _started = true;
      return r;
    } finally {
      _free(p);
    }
  }

  /// Locks the vault and closes the sessions (before the app quits).
  static void stop() {
    final c = _instance;
    if (c == null || !c._started) return;
    c._started = false;
    c._stop();
  }
}

typedef _ClickFn = Void Function(Pointer<Utf8>);
