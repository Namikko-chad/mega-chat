// Renders every section at a desktop and a phone width against a running
// core and writes PNGs, for checking the UI against the design manifest:
//
//   SHOTS=/tmp/shots MC_PASSWORD=... flutter run -d linux -t tool/screenshots.dart
import 'dart:io';
import 'dart:ui' as ui;

import 'package:parley/app.dart';
import 'package:parley/core/settings/app_settings.dart';
import 'package:parley/data/rpc/endpoint.dart';
import 'package:parley/data/rpc/gen/parley.pbgrpc.dart' as pb;
import 'package:parley/presentation/core/core_cubit.dart';
import 'package:parley/presentation/chats/chats_screen.dart';
import 'package:parley/presentation/core/messenger_cubit.dart';
import 'package:parley/presentation/shell/app_shell.dart';
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

final _boundary = GlobalKey();

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await AppSettings.load();
  final out = Directory(Platform.environment['SHOTS'] ?? 'shots')..createSync(recursive: true);
  final core = CoreCubit(CoreLauncher.forPlatform());
  await core.connect();
  final frame = ValueNotifier<(Size, Section, int)>((const Size(1440, 900), Section.chats, 0));
  runApp(
    ValueListenableBuilder(
      valueListenable: frame,
      builder: (_, f, _) => ParleyApp(
        key: ValueKey(f),
        core: core,
        home: AppShell(initial: f.$2),
        frame: (context, child) {
          final mq = MediaQuery.of(context);
          return RepaintBoundary(
            key: _boundary,
            child: OverflowBox(
              alignment: Alignment.topLeft,
              minWidth: f.$1.width,
              maxWidth: f.$1.width,
              minHeight: f.$1.height,
              maxHeight: f.$1.height,
              child: MediaQuery(
                data: mq.copyWith(size: f.$1, padding: EdgeInsets.zero, viewPadding: EdgeInsets.zero),
                child: child!,
              ),
            ),
          );
        },
      ),
    ),
  );

  const sizes = {'desktop': Size(1440, 900), 'phone': Size(390, 844)};
  Future<void> shot(String name, Size size, Section s, {bool withChat = false}) async {
    frame.value = (size, s, frame.value.$3 + 1);
    await Future<void>.delayed(const Duration(milliseconds: 2500));
    if (withChat) {
      // The newest chat, opened the way a tap opens it (phones push a
      // route, so the context must be below the navigator).
      Element? shell;
      void find(Element e) {
        if (shell != null) return;
        if (e.widget is AppShell) {
          shell = e;
          return;
        }
        e.visitChildren(find);
      }

      (_boundary.currentContext! as Element).visitChildren(find);
      final first = shell?.read<MessengerCubit>().state.sorted.firstOrNull;
      if (first != null) openChat(shell!, chatKey(first));
      await Future<void>.delayed(const Duration(milliseconds: 2500));
    }
    final ro = _boundary.currentContext!.findRenderObject()! as RenderRepaintBoundary;
    final img = await ro.toImage(pixelRatio: 1);
    final png = await img.toByteData(format: ui.ImageByteFormat.png);
    await File('${out.path}/$name.png').writeAsBytes(png!.buffer.asUint8List());
    stdout.writeln('shot ${out.path}/$name.png');
  }

  if (!core.state.unlocked) {
    for (final e in sizes.entries) {
      await shot('unlock-${e.key}', e.value, Section.chats);
    }
    final pw = Platform.environment['MC_PASSWORD'] ?? '';
    core.adopt(await core.api.unlock(pb.PasswordRequest(password: pw)));
    await Future<void>.delayed(const Duration(seconds: 2));
  }
  for (final s in Section.values) {
    for (final e in sizes.entries) {
      await shot('${s.name}-${e.key}', e.value, s);
    }
  }
  for (final e in sizes.entries) {
    await shot('chat-${e.key}', e.value, Section.chats, withChat: true);
  }
  exit(0);
}
