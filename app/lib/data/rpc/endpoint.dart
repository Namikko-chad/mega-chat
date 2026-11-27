import 'dart:io';

import 'package:flutter/services.dart';
import 'package:path/path.dart' as p;

import 'native_core.dart';

/// Where the core listens and the token it wants.
class Endpoint {
  const Endpoint({required this.address, this.port = 0, required this.token});

  final InternetAddress address;
  final int port;
  final String token;
}

/// Starts the core of this platform (inside the app) and says where it listens.
abstract class CoreLauncher {
  Future<Endpoint> endpoint();

  static CoreLauncher forPlatform() => Platform.isAndroid ? AndroidLauncher() : DesktopLauncher();
}

/// Android: the core runs inside the app (gomobile AAR) in a foreground
/// service; Kotlin starts it and hands over the socket path and token.
class AndroidLauncher implements CoreLauncher {
  static const _ch = MethodChannel('dev.selfhold.parley/core');

  @override
  Future<Endpoint> endpoint() async {
    final r = await _ch.invokeMapMethod<String, String>('start');
    if (r == null) throw StateError('core did not start');
    return Endpoint(
      address: InternetAddress(r['socket']!, type: InternetAddressType.unix),
      token: r['token']!,
    );
  }
}

/// Desktop: the core runs inside the app ([NativeCore]) and listens on a
/// private socket. When another instance of the app already runs it, this
/// one connects to that core (endpoint.json in the config dir); when that
/// instance quits, the next reconnect starts the core here.
class DesktopLauncher implements CoreLauncher {
  static String get configDir {
    if (Platform.isWindows) return p.join(Platform.environment['APPDATA'] ?? '.', 'selfhold', 'parley');
    final xdg = Platform.environment['XDG_CONFIG_HOME'];
    final home = Platform.environment['HOME'] ?? '.';
    return p.join(xdg != null && xdg.isNotEmpty ? xdg : p.join(home, '.config'), 'selfhold', 'parley');
  }

  @override
  Future<Endpoint> endpoint() async {
    final r = NativeCore.instance.start();
    final err = r['error'];
    if (err != null) throw SocketException('the core did not start: $err');
    return _parse(r['address'] as String, r['token'] as String);
  }

  static Endpoint _parse(String addr, String token) {
    if (addr.startsWith('unix://')) {
      return Endpoint(
        address: InternetAddress(addr.substring('unix://'.length), type: InternetAddressType.unix),
        token: token,
      );
    }
    final i = addr.lastIndexOf(':');
    return Endpoint(address: InternetAddress(addr.substring(0, i)), port: int.parse(addr.substring(i + 1)), token: token);
  }
}
