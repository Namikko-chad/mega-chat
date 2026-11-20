import 'dart:async';

import 'package:flutter/material.dart';
import 'package:media_kit/media_kit.dart';

import 'app.dart';
import 'core/desktop/desktop_entry.dart';
import 'core/desktop/desktop_window.dart';
import 'core/settings/app_settings.dart';
import 'data/biometric.dart';
import 'data/rpc/endpoint.dart';
import 'presentation/core/core_cubit.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  MediaKit.ensureInitialized();
  await AppSettings.load();
  await Biometric.refresh();
  await DesktopWindow.instance.init();
  unawaited(registerDesktopEntry());
  final core = CoreCubit(CoreLauncher.forPlatform())..connect();
  runApp(MegaChatApp(core: core));
}
