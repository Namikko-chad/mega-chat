import 'dart:io';

import 'package:flutter/material.dart';
import 'package:tray_manager/tray_manager.dart';
import 'package:window_manager/window_manager.dart';

import '../../data/rpc/native_core.dart';
import '../i18n.dart';
import '../settings/app_settings.dart';

/// Desktop window chrome: frameless window drawn by the app (the header is
/// the title bar), and a tray icon so closing can keep the app (and with it
/// the core, which runs in this process) connected.
class DesktopWindow with WindowListener, TrayListener {
  DesktopWindow._();

  static final instance = DesktopWindow._();

  static bool get supported => Platform.isLinux || Platform.isWindows || Platform.isMacOS;

  bool _tray = false;

  /// The window has the keyboard focus: the user looks at the app.
  final focused = ValueNotifier(true);

  Future<void> init() async {
    if (!supported) return;
    await windowManager.ensureInitialized();
    const options = WindowOptions(
      title: 'MegaChat',
      size: Size(1440, 900),
      minimumSize: Size(420, 560),
      center: true,
      backgroundColor: Color(0xFF080C10),
      titleBarStyle: TitleBarStyle.hidden,
    );
    await windowManager.waitUntilReadyToShow(options, () async {
      await windowManager.show();
      await windowManager.focus();
    });
    await windowManager.setPreventClose(true);
    windowManager.addListener(this);
    try {
      await trayManager.setIcon(Platform.isWindows ? 'assets/tray/tray.ico' : 'assets/tray/tray.png');
      if (!Platform.isLinux) await trayManager.setToolTip('MegaChat');
      trayManager.addListener(this);
      _tray = true;
      await _menu();
      lang.addListener(_menu);
    } catch (_) {
      // No tray on this desktop: closing always quits.
      _tray = false;
    }
  }

  Future<void> _menu() async {
    if (!_tray) return;
    await trayManager.setContextMenu(
      Menu(
        items: [
          MenuItem(key: 'show', label: tr('Open MegaChat', 'Открыть MegaChat')),
          MenuItem.separator(),
          MenuItem(key: 'quit', label: tr('Quit', 'Выйти')),
        ],
      ),
    );
  }

  Future<void> show() async {
    await windowManager.show();
    await windowManager.focus();
  }

  /// Closes the window and the app; the core locks the vault and
  /// disconnects the accounts first.
  Future<void> quit() async {
    // Gone at once; the core may take a few seconds to disconnect.
    await windowManager.hide();
    NativeCore.stop();
    if (_tray) await trayManager.destroy();
    await windowManager.setPreventClose(false);
    await windowManager.destroy();
  }

  /// The close button: hide to the tray when asked to and possible.
  Future<void> close() async {
    if (_tray && AppSettings.instance.closeToTray) {
      await windowManager.hide();
      focused.value = false;
    } else {
      await quit();
    }
  }

  Future<void> minimize() => windowManager.minimize();

  Future<void> toggleMaximize() async {
    if (await windowManager.isMaximized()) {
      await windowManager.unmaximize();
    } else {
      await windowManager.maximize();
    }
  }

  @override
  void onWindowClose() => close();

  @override
  void onWindowFocus() => focused.value = true;

  @override
  void onWindowBlur() => focused.value = false;

  @override
  void onWindowMinimize() => focused.value = false;

  @override
  void onTrayIconMouseDown() => show();

  @override
  void onTrayIconRightMouseDown() => trayManager.popUpContextMenu();

  @override
  void onTrayMenuItemClick(MenuItem menuItem) {
    switch (menuItem.key) {
      case 'show':
        show();
      case 'quit':
        quit();
    }
  }
}
