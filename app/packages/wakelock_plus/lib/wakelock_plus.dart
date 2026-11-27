import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

/// Keeps the screen on while a video plays. Android: the activity's
/// FLAG_KEEP_SCREEN_ON (MainActivity, channel "dev.selfhold.parley/wakelock");
/// elsewhere nothing.
abstract final class WakelockPlus {
  static const _channel = MethodChannel('dev.selfhold.parley/wakelock');
  static bool _on = false;

  static Future<void> enable() => toggle(enable: true);

  static Future<void> disable() => toggle(enable: false);

  static Future<void> toggle({required bool enable}) async {
    _on = enable;
    if (kIsWeb || defaultTargetPlatform != TargetPlatform.android) return;
    try {
      await _channel.invokeMethod<void>('toggle', enable);
    } on MissingPluginException {
      // An engine without the channel (tests).
    }
  }

  static Future<bool> get enabled async => _on;
}
