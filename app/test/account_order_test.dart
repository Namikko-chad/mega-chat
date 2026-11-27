import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:megachat/core/theme/app_theme.dart';
import 'package:megachat/data/rpc/endpoint.dart';
import 'package:megachat/data/rpc/gen/megachat.pbgrpc.dart' as pb;
import 'package:megachat/presentation/chats/chats_screen.dart';
import 'package:megachat/presentation/core/core_cubit.dart';
import 'package:megachat/presentation/core/messenger_cubit.dart';

class _NoCore implements CoreLauncher {
  @override
  Future<Endpoint> endpoint() async => throw const SocketException('no core in tests');
}

/// Applies a move like the real cubit, without a core to save it to.
class _Messenger extends MessengerCubit {
  _Messenger(super.core);

  @override
  Future<void> moveAccount(int from, int to) async {
    if (to > from) to--;
    final list = [...state.accounts];
    list.insert(to, list.removeAt(from));
    emit(state.copyWith(accounts: list));
  }
}

void main() {
  for (final (name, size) in const [('compact rail', Size(1000, 800)), ('full rail', Size(1400, 800))]) {
    testWidgets('$name: dragging an account reorders, a tap still selects', (tester) async {
      tester.view.physicalSize = size;
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.reset);
      final core = CoreCubit(_NoCore());
      final m = _Messenger(core);
      for (final n in ['one', 'two', 'three']) {
        m.upsertAccount(pb.Account(id: n, name: n, provider: 'fake', enabled: true, capabilities: [pb.Capability.CAPABILITY_MESSAGES]));
      }
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.dark,
          home: Scaffold(
            body: BlocProvider<MessengerCubit>.value(value: m, child: const ChatsScreen()),
          ),
        ),
      );
      await tester.pump();

      await tester.tap(find.text('two'));
      await tester.pump(const Duration(milliseconds: 300));
      expect(m.state.filter, 'two');

      final from = tester.getCenter(find.text('one'));
      final to = tester.getCenter(find.text('three'));
      final touch = defaultTargetPlatform == TargetPlatform.android;
      final drag = await tester.startGesture(from, kind: touch ? PointerDeviceKind.touch : PointerDeviceKind.mouse);
      // Touch starts a drag with a long press, so that a swipe still scrolls.
      await tester.pump(touch ? kLongPressTimeout + const Duration(milliseconds: 100) : const Duration(milliseconds: 50));
      for (var i = 1; i <= 10; i++) {
        await drag.moveTo(Offset.lerp(from, to + const Offset(0, 30), i / 10)!);
        await tester.pump(const Duration(milliseconds: 50));
      }
      await drag.up();
      await tester.pump(const Duration(milliseconds: 500));
      await tester.pump(const Duration(milliseconds: 500));

      expect(m.state.accounts.map((a) => a.id), ['two', 'three', 'one']);
      expect(tester.takeException(), isNull);
      await tester.runAsync(() async {
        await m.close();
        await core.close();
      });
    }, variant: const TargetPlatformVariant({TargetPlatform.linux, TargetPlatform.android}));
  }
}
