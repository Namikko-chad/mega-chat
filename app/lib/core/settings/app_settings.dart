import 'dart:convert';
import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';

import '../../data/rpc/endpoint.dart';
import '../i18n.dart';

/// How much of a video in a chat loads without a tap.
enum VideoAuto {
  none, // nothing: a tap loads the preview
  preview, // the preview picture; the video itself on play
  full, // the preview, and small videos are fetched ahead into the core's cache
}

/// Which new messages show a system notification while the app is not
/// looked at. Muted chats notify only about mentions in any case.
enum NotifyMode {
  all, // every message worth one (Event.notify)
  mentions, // only those that mention the user (and new mail)
  off,
}

/// App preferences of this device (not part of the vault): kept in
/// `<config dir>/megachat/app.json` on desktops, in app support on Android.
class AppSettings extends ChangeNotifier {
  AppSettings._(this._file);

  final File? _file;
  bool _closeToTray = false;
  bool _biometricDeclined = false;
  bool _groupChats = true;
  bool _unreadOnly = false;
  bool _autoPhotos = true;
  VideoAuto _autoVideo = VideoAuto.preview;
  bool _autoAudio = false;
  bool _linkPreviews = true;
  NotifyMode _notify = NotifyMode.all;

  static AppSettings? _instance;

  /// The loaded settings; defaults (not saved) until [load] ran, e.g. in
  /// widget tests.
  static AppSettings get instance => _instance ??= AppSettings._(null);

  /// The chat list is split into channels and groups, and direct messages.
  bool get groupChats => _groupChats;

  /// The chat list shows only chats with unread messages.
  bool get unreadOnly => _unreadOnly;

  /// Pictures in chats load as they are shown (off: on a tap).
  bool get autoPhotos => _autoPhotos;

  /// Links in messages show their page's preview (the core fetches the
  /// page, so the site sees this machine's address).
  bool get linkPreviews => _linkPreviews;

  /// What of a video in a chat loads as it is shown.
  VideoAuto get autoVideo => _autoVideo;

  /// Voice messages and audio files are fetched ahead as they are shown
  /// (off: only when played).
  bool get autoAudio => _autoAudio;

  /// Which new messages show a notification.
  NotifyMode get notify => _notify;

  /// Closing the window hides it to the tray instead of quitting (desktop).
  bool get closeToTray => _closeToTray;

  Lang get language => lang.value;

  /// The user said no to unlocking by fingerprint: don't offer it again
  /// after each password unlock (Android; can be turned on in Settings).
  bool get biometricDeclined => _biometricDeclined;

  static Future<AppSettings> load() async {
    File? file;
    try {
      final dir = Platform.isAndroid ? (await getApplicationSupportDirectory()).path : DesktopLauncher.configDir;
      file = File(p.join(dir, 'app.json'));
    } catch (_) {}
    final s = AppSettings._(file);
    try {
      if (file != null && await file.exists()) {
        final j = jsonDecode(await file.readAsString()) as Map<String, dynamic>;
        s._closeToTray = j['close_to_tray'] == true;
        s._biometricDeclined = j['biometric_declined'] == true;
        s._groupChats = j['group_chats'] != false;
        s._unreadOnly = j['unread_only'] == true;
        s._autoPhotos = j['auto_photos'] != false;
        s._autoVideo = VideoAuto.values.firstWhere((v) => v.name == j['auto_video'], orElse: () => VideoAuto.preview);
        s._autoAudio = j['auto_audio'] == true;
        s._linkPreviews = j['link_previews'] != false;
        s._notify = NotifyMode.values.firstWhere((v) => v.name == j['notify'], orElse: () => NotifyMode.all);
        lang.value = Lang.values.firstWhere((l) => l.name == j['language'], orElse: () => Lang.en);
      }
    } catch (_) {
      // A broken file means defaults.
    }
    s._tellAndroid();
    return _instance = s;
  }

  // Android shows notifications in Kotlin, also while the UI is not running:
  // it gets the notification mode and the chosen language (its strings follow
  // the app's language, never the system locale).
  void _tellAndroid() {
    if (!Platform.isAndroid) return;
    const channel = MethodChannel('dev.megachat/core');
    channel.invokeMethod<void>('notifications', _notify.name).catchError((_) {});
    channel.invokeMethod<void>('language', lang.value.name).catchError((_) {});
  }

  Future<void> setNotify(NotifyMode v) async {
    _notify = v;
    _tellAndroid();
    notifyListeners();
    await _save();
  }

  Future<void> setLanguage(Lang l) async {
    lang.value = l;
    _tellAndroid();
    notifyListeners();
    await _save();
  }

  Future<void> setCloseToTray(bool v) async {
    _closeToTray = v;
    notifyListeners();
    await _save();
  }

  Future<void> setBiometricDeclined(bool v) async {
    _biometricDeclined = v;
    notifyListeners();
    await _save();
  }

  Future<void> setGroupChats(bool v) async {
    _groupChats = v;
    notifyListeners();
    await _save();
  }

  Future<void> setUnreadOnly(bool v) async {
    _unreadOnly = v;
    notifyListeners();
    await _save();
  }

  Future<void> setAutoPhotos(bool v) async {
    _autoPhotos = v;
    notifyListeners();
    await _save();
  }

  Future<void> setAutoVideo(VideoAuto v) async {
    _autoVideo = v;
    notifyListeners();
    await _save();
  }

  Future<void> setLinkPreviews(bool v) async {
    _linkPreviews = v;
    notifyListeners();
    await _save();
  }

  Future<void> setAutoAudio(bool v) async {
    _autoAudio = v;
    notifyListeners();
    await _save();
  }

  Future<void> _save() async {
    final f = _file;
    if (f == null) return;
    try {
      await f.parent.create(recursive: true);
      await f.writeAsString(
        jsonEncode({
          'language': lang.value.name,
          'close_to_tray': _closeToTray,
          'biometric_declined': _biometricDeclined,
          'group_chats': _groupChats,
          'unread_only': _unreadOnly,
          'auto_photos': _autoPhotos,
          'auto_video': _autoVideo.name,
          'auto_audio': _autoAudio,
          'link_previews': _linkPreviews,
          'notify': _notify.name,
        }),
      );
    } catch (_) {}
  }
}
