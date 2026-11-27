import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:window_manager/window_manager.dart';

import 'core/desktop/desktop_window.dart';
import 'core/i18n.dart';
import 'core/theme/app_theme.dart';
import 'core/theme/layout.dart';
import 'presentation/call/call_cubit.dart';
import 'presentation/call/call_view.dart';
import 'presentation/common/overlays.dart';
import 'presentation/core/core_cubit.dart';
import 'presentation/core/messenger_cubit.dart';
import 'presentation/core/notifier.dart';
import 'presentation/shell/app_shell.dart';

class MegaChatApp extends StatefulWidget {
  const MegaChatApp({super.key, required this.core, this.home = const AppShell(), this.frame});

  final CoreCubit core;
  final Widget home;

  /// Wraps the whole app (dev tools render it at a fixed size).
  final TransitionBuilder? frame;

  @override
  State<MegaChatApp> createState() => _MegaChatAppState();
}

class _MegaChatAppState extends State<MegaChatApp> {
  // Above the navigator: pushed screens (a chat on phones) share it.
  late final _messenger = MessengerCubit(widget.core);
  late final _calls = CallCubit(widget.core);
  // Not under dev tools (frame): they draw the app, nothing to notify.
  late final _notifier = DesktopNotifier.supported && widget.frame == null ? DesktopNotifier(widget.core, _messenger) : null;

  @override
  void initState() {
    super.initState();
    lang.addListener(_relabel);
    _notifier;
  }

  @override
  void dispose() {
    lang.removeListener(_relabel);
    _notifier?.close();
    _messenger.close();
    _calls.close();
    super.dispose();
  }

  /// Strings come from tr() in build methods, so a language switch
  /// rebuilds every element; state (current page, forms) is kept.
  void _relabel() {
    setState(() {});
    void mark(Element e) {
      e.markNeedsBuild();
      e.visitChildren(mark);
    }

    (context as Element).visitChildren(mark);
  }

  CoreCubit get core => widget.core;
  Widget get home => widget.home;
  TransitionBuilder? get frame => widget.frame;

  @override
  Widget build(BuildContext context) => MultiBlocProvider(
    providers: [
      BlocProvider.value(value: core),
      BlocProvider.value(value: _messenger),
      BlocProvider.value(value: _calls),
    ],
    child: MaterialApp(
      title: 'MegaChat',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.dark,
      darkTheme: AppTheme.dark,
      themeMode: ThemeMode.dark,
      locale: currentLocale,
      supportedLocales: const [Locale('en'), Locale('ru')],
      localizationsDelegates: GlobalMaterialLocalizations.delegates,
      home: home,
      // Toasts are mounted after the navigator, so they stay above
      // modals and sheets and above the tab bar on phones.
      builder: (context, child) {
        final app = Stack(
          children: [
            ?child,
            const CallOverlay(),
            BlocBuilder<CoreCubit, CoreState>(
              buildWhen: (a, b) => a.unlocked != b.unlocked,
              builder: (context, s) => ToastHost(bottomOffset: Hm.of(context).mobile && s.unlocked ? 56 : 0),
            ),
          ],
        );
        // Frameless Linux window: the app gives it resize edges.
        final framed = frame == null && DesktopWindow.supported && Platform.isLinux ? DragToResizeArea(resizeEdgeSize: 6, child: app) : app;
        return frame == null ? framed : frame!(context, app);
      },
    ),
  );
}
