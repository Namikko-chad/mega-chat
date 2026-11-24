import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import 'package:window_manager/window_manager.dart';

import '../../core/config/app_config.dart';
import '../../core/desktop/desktop_window.dart';
import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/layout.dart';
import '../../core/theme/typography.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../accounts/accounts_screen.dart';
import '../accounts/providers.dart';
import '../calendar/calendar_cubit.dart';
import '../calendar/calendar_screen.dart';
import '../chats/chats_screen.dart';
import '../common/hm.dart';
import '../core/core_cubit.dart';
import '../core/messenger_cubit.dart';
import '../log/log_screen.dart';
import '../mail/mail_cubit.dart';
import '../mail/mail_screen.dart';
import '../settings/settings_screen.dart';
import '../unlock/unlock_screen.dart';

/// The sections of the app in navigation order; the header and the tab
/// bar are built from this list. Mail and calendar are shown only while
/// some account has their pb.Capability.
enum Section {
  chats(Icons.chat_bubble_outline),
  mail(Icons.mail_outline),
  calendar(Icons.calendar_month_outlined),
  accounts(Icons.manage_accounts_outlined, tab: false),
  log(Icons.receipt_long_outlined, tab: false),
  settings(Icons.settings_outlined);

  const Section(this.icon, {this.tab = true});

  final IconData icon;

  /// A phone tab and a header link; accounts and the log open from Settings.
  final bool tab;

  String get label => switch (this) {
    Section.chats => tr('Chats', 'Чаты'),
    Section.mail => tr('Mail', 'Почта'),
    Section.calendar => tr('Calendar', 'Календарь'),
    Section.accounts => tr('Accounts', 'Аккаунты'),
    Section.log => tr('Log', 'Журнал'),
    Section.settings => tr('Settings', 'Настройки'),
  };

  /// Settings is the "profile" of the app: an identity anchor (§ 4.6), not a nav link.
  static List<Section> nav(Set<Section> shown) => [
    for (final s in values)
      if (s != settings && s.tab && shown.contains(s)) s,
  ];

  static List<Section> tabs(Set<Section> shown) => [
    for (final s in values)
      if (s.tab && shown.contains(s)) s,
  ];

  /// The sections the accounts make available.
  static Set<Section> available(List<pb.Account> accounts) => {
    for (final s in values)
      if ((s != mail || accounts.any(hasMail)) && (s != calendar || accounts.any(hasCalendar))) s,
  };
}

/// The page shell (§ 2, § 4–6; MOBILITY M2, M4, M6): picks the section and
/// lays out header, canvas and footer or tab bar.
class AppShell extends StatefulWidget {
  const AppShell({super.key, this.initial = Section.chats});

  final Section initial;

  /// Opens a section.
  static void go(BuildContext context, Section s) => context.findAncestorStateOfType<_AppShellState>()?._go(s);

  /// Opens a section from outside the widget tree (a clicked notification).
  static final requests = StreamController<Section>.broadcast();

  /// Opens the Log section filtered to one account.
  static void openLog(BuildContext context, String accountId) => context.findAncestorStateOfType<_AppShellState>()?._openLog(accountId);

  @override
  State<AppShell> createState() => _AppShellState();
}

class _AppShellState extends State<AppShell> {
  late Section _section = widget.initial;
  String? _logFilter;
  late final StreamSubscription<Section> _requests = AppShell.requests.stream.listen(_go);

  @override
  void initState() {
    super.initState();
    _requests;
  }

  @override
  void dispose() {
    _requests.cancel();
    super.dispose();
  }

  void _openLog(String accountId) => setState(() {
    _section = Section.log;
    _logFilter = accountId;
  });

  void _go(Section s) {
    if (s != _section) HapticFeedback.selectionClick();
    setState(() {
      _section = s;
      _logFilter = null;
    });
  }

  Future<void> _lock() async {
    try {
      await context.read<CoreCubit>().lockVault();
    } catch (e) {
      Toasts.error(tr('Could not lock', 'Не удалось заблокировать'), '$e');
    }
  }

  // Mail and calendar keep their state while other sections are shown.
  @override
  Widget build(BuildContext context) => MultiBlocProvider(
    providers: [
      BlocProvider(create: (ctx) => MailCubit(ctx.read<CoreCubit>())),
      BlocProvider(create: (ctx) => CalendarCubit(ctx.read<CoreCubit>())),
    ],
    child: Builder(builder: _build),
  );

  Widget _build(BuildContext context) {
    final hm = Hm.of(context);
    final core = context.watch<CoreCubit>().state;
    final m = context.watch<MessengerCubit>().state;
    final unlocked = core.unlocked;
    final shown = Section.available(m.accounts);
    final tabs = Section.tabs(shown);
    final counts = {Section.chats: m.unread, Section.mail: m.accounts.fold(0, (n, a) => n + (a.enabled ? a.mailUnread : 0))};
    final Widget page = switch ((core.conn, unlocked)) {
      (Conn.connecting, false) when core.state == null => const HmLoading(),
      (_, false) => const UnlockScreen(),
      _ => switch (_section) {
        Section.chats => const ChatsScreen(),
        Section.mail => const MailScreen(),
        Section.calendar => const CalendarScreen(),
        Section.accounts => const AccountsScreen(),
        Section.log => LogScreen(key: ValueKey(_logFilter), accountId: _logFilter),
        Section.settings => SettingsScreen(onLock: _lock),
      },
    };
    final body = Stack(
      children: [
        const Positioned.fill(child: _GridBackground()),
        Positioned.fill(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              if (unlocked && core.conn != Conn.ready) _Reconnecting(core: core),
              Expanded(
                child: KeyedSubtree(key: ValueKey(unlocked ? _section : 'locked'), child: page),
              ),
            ],
          ),
        ),
      ],
    );

    if (hm.mobile) {
      // The log is not a tab: Settings stays lit while it is open.
      final lit = tabs.contains(_section) ? _section : Section.settings;
      return Scaffold(
        backgroundColor: HmColors.bg,
        body: SafeArea(
          bottom: false,
          child: Column(
            children: [
              _MobileHeader(online: core.conn == Conn.ready),
              Expanded(child: body),
            ],
          ),
        ),
        bottomNavigationBar: unlocked
            ? DecoratedBox(
                decoration: const BoxDecoration(
                  border: Border(top: BorderSide(color: HmColors.border)),
                ),
                child: NavigationBar(
                  selectedIndex: tabs.indexOf(lit),
                  onDestinationSelected: (i) => _go(tabs[i]),
                  labelBehavior: NavigationDestinationLabelBehavior.alwaysShow,
                  destinations: [
                    for (final s in tabs)
                      NavigationDestination(
                        icon: _TabIcon(icon: s.icon, count: counts[s] ?? 0),
                        label: s.label,
                        tooltip: '',
                      ),
                  ],
                ),
              )
            : null,
      );
    }
    return Scaffold(
      backgroundColor: HmColors.bezel,
      body: Container(
        margin: const EdgeInsets.all(8), // viewport bezel (§ 5.3)
        color: HmColors.bg,
        child: Column(
          children: [
            _Header(section: unlocked ? _section : null, nav: Section.nav(shown), counts: counts, onGo: _go, onLock: _lock),
            Expanded(child: body),
            _Footer(core: core),
          ],
        ),
      ),
    );
  }
}

/// The connection dropped while unlocked: the screens stay, this says why.
class _Reconnecting extends StatelessWidget {
  const _Reconnecting({required this.core});

  final CoreState core;

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    return Padding(
      padding: EdgeInsets.fromLTRB(hm.pageX, 12, hm.pageX, 0),
      child: HmBanner(
        '${core.error ?? tr('No connection to the core', 'Нет связи с ядром')} — ${tr('reconnecting…', 'переподключение…')}',
        tone: HmTone.warning,
        action: HmButton(tr('Retry', 'Повторить'), small: true, loading: core.conn == Conn.connecting, onPressed: () => context.read<CoreCubit>().connect()),
      ),
    );
  }
}

class _TabIcon extends StatelessWidget {
  const _TabIcon({required this.icon, required this.count});

  final IconData icon;
  final int count;

  @override
  Widget build(BuildContext context) => Stack(
    clipBehavior: Clip.none,
    children: [
      Icon(icon),
      if (count > 0) Positioned(left: 14, top: -6, child: HmCount(count)),
    ],
  );
}

class _GridBackground extends StatelessWidget {
  const _GridBackground();

  @override
  Widget build(BuildContext context) => const IgnorePointer(
    child: RepaintBoundary(child: CustomPaint(painter: _GridPainter())),
  );
}

/// Fixed 40 px cyan grid at 3 % (§ 2.2); it does not scroll with content.
class _GridPainter extends CustomPainter {
  const _GridPainter();

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = HmColors.tint(HmColors.accent, HmColors.subtle)
      ..strokeWidth = 1;
    for (double x = 0; x < size.width; x += 40) {
      canvas.drawLine(Offset(x, 0), Offset(x, size.height), paint);
    }
    for (double y = 0; y < size.height; y += 40) {
      canvas.drawLine(Offset(0, y), Offset(size.width, y), paint);
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

const _headerGradient = LinearGradient(
  begin: Alignment.topCenter,
  end: Alignment.bottomCenter,
  colors: [Color(0x0A00D4FF), Color(0x0000D4FF)], // accent @ 4 % → 0
);

/// Desktop header (§ 4): logo, nav links, then the identity anchors —
/// "Settings" (the profile of this app) and "Lock" (its sign-out).
class _Header extends StatelessWidget {
  const _Header({required this.section, required this.nav, required this.counts, required this.onGo, required this.onLock});

  final Section? section;
  final List<Section> nav;
  final Map<Section, int> counts;
  final ValueChanged<Section> onGo;
  final VoidCallback onLock;

  @override
  Widget build(BuildContext context) => _TitleBar(
    child: Container(
      padding: const EdgeInsets.fromLTRB(32, 14, 12, 14),
      decoration: const BoxDecoration(
        gradient: _headerGradient,
        border: Border(bottom: BorderSide(color: HmColors.border)),
      ),
      child: Row(
        children: [
          Text('MegaChat', style: HmText.logo),
          const SizedBox(width: 16),
          // A narrow window scrolls the links; the window buttons always stay.
          Expanded(
            child: section == null
                ? const SizedBox()
                : Align(
                    alignment: Alignment.centerRight,
                    child: SingleChildScrollView(
                      scrollDirection: Axis.horizontal,
                      reverse: true,
                      child: Row(
                        children: [
                          for (final s in nav) ...[_NavLink(label: s.label, active: section == s, count: counts[s] ?? 0, onTap: () => onGo(s)), const SizedBox(width: 12)],
                          Container(width: 1, height: 18, margin: const EdgeInsets.symmetric(horizontal: 4), color: HmColors.border),
                          const SizedBox(width: 12),
                          _NavLink(label: Section.settings.label, active: true, onTap: () => onGo(Section.settings)),
                          const SizedBox(width: 12),
                          _NavLink(label: tr('Lock', 'Заблокировать'), signOut: true, onTap: onLock),
                        ],
                      ),
                    ),
                  ),
          ),
          if (DesktopWindow.supported) ...[const SizedBox(width: 16), const _WindowButtons()],
        ],
      ),
    ),
  );
}

/// The app's header doubles as the window title bar on desktops: drag to
/// move, double-click to maximize.
class _TitleBar extends StatelessWidget {
  const _TitleBar({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) => DesktopWindow.supported ? DragToMoveArea(child: child) : child;
}

/// Minimize, maximize and close, drawn in the app's style.
class _WindowButtons extends StatelessWidget {
  const _WindowButtons();

  @override
  Widget build(BuildContext context) => Row(
    mainAxisSize: MainAxisSize.min,
    children: [
      _WindowButton(icon: Icons.remove, tooltip: tr('Minimize', 'Свернуть'), onTap: DesktopWindow.instance.minimize),
      _WindowButton(icon: Icons.crop_square, tooltip: tr('Maximize', 'Развернуть'), onTap: DesktopWindow.instance.toggleMaximize),
      _WindowButton(icon: Icons.close, tooltip: tr('Close', 'Закрыть'), danger: true, onTap: DesktopWindow.instance.close),
    ],
  );
}

class _WindowButton extends StatefulWidget {
  const _WindowButton({required this.tooltip, required this.onTap, required this.icon, this.danger = false});

  final IconData icon;
  final String tooltip;
  final VoidCallback onTap;
  final bool danger;

  @override
  State<_WindowButton> createState() => _WindowButtonState();
}

class _WindowButtonState extends State<_WindowButton> {
  bool _hot = false;

  @override
  Widget build(BuildContext context) {
    final base = widget.danger ? HmColors.error : HmColors.accent;
    final color = _hot ? base : HmColors.muted;
    return Tooltip(
      message: widget.tooltip,
      child: MouseRegion(
        cursor: SystemMouseCursors.click,
        onEnter: (_) => setState(() => _hot = true),
        onExit: (_) => setState(() => _hot = false),
        child: GestureDetector(
          onTap: widget.onTap,
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 150),
            width: 32,
            height: 28,
            margin: const EdgeInsets.only(left: 2),
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: HmColors.tint(base, _hot ? 0.12 : 0.0),
              borderRadius: BorderRadius.circular(6),
            ),
            child: Icon(widget.icon, size: 16, color: color),
          ),
        ),
      ),
    );
  }
}

class _NavLink extends StatefulWidget {
  const _NavLink({required this.label, required this.onTap, this.active = false, this.signOut = false, this.count = 0});

  final String label;
  final VoidCallback onTap;
  final bool active;
  final bool signOut;
  final int count; // unread, next to the label

  @override
  State<_NavLink> createState() => _NavLinkState();
}

class _NavLinkState extends State<_NavLink> {
  bool _hot = false;

  @override
  Widget build(BuildContext context) {
    final base = widget.signOut ? HmColors.error : HmColors.accent;
    final lit = widget.active || widget.signOut;
    final alpha = (lit ? 0.08 : 0.0) + (_hot ? 0.10 : 0.0);
    return Semantics(
      button: true,
      selected: widget.active,
      child: MouseRegion(
        cursor: SystemMouseCursors.click,
        onEnter: (_) => setState(() => _hot = true),
        onExit: (_) => setState(() => _hot = false),
        child: GestureDetector(
          onTap: widget.onTap,
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 150),
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
            decoration: BoxDecoration(
              color: HmColors.tint(base, alpha),
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: lit ? HmColors.tint(base, HmColors.ring) : Colors.transparent),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(widget.label.toUpperCase(), style: HmText.nav.copyWith(color: lit || _hot ? base : HmColors.muted)),
                if (widget.count > 0) ...[const SizedBox(width: 8), HmCount(widget.count)],
              ],
            ),
          ),
        ),
      ),
    );
  }
}

/// Phone header (M4.1): logo and a live connection dot (App. D).
class _MobileHeader extends StatelessWidget {
  const _MobileHeader({required this.online});

  final bool online;

  @override
  Widget build(BuildContext context) => _TitleBar(
    child: Container(
      height: 56,
      padding: const EdgeInsets.only(left: 16, right: 8),
      decoration: const BoxDecoration(
        gradient: _headerGradient,
        border: Border(bottom: BorderSide(color: HmColors.border)),
      ),
      child: Row(
        children: [
          Text('MegaChat', style: HmText.logo.copyWith(fontSize: 18)),
          const Spacer(),
          HmDot(ok: online, label: online ? tr('Online', 'В сети') : tr('Offline', 'Не в сети')),
          // On a narrow desktop window the window buttons take this place.
          if (!DesktopWindow.supported) ...[const SizedBox(width: 8), Text(online ? tr('online', 'в сети') : tr('offline', 'не в сети'), style: HmText.footer)],
          if (DesktopWindow.supported) ...[const SizedBox(width: 8), const _WindowButtons()],
        ],
      ),
    ),
  );
}

/// Desktop status bar (§ 6): version and build, online and env pills.
class _Footer extends StatelessWidget {
  const _Footer({required this.core});

  final CoreState core;

  @override
  Widget build(BuildContext context) {
    final vault = core.state?.vaultPath ?? '';
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 40, vertical: 14),
      decoration: const BoxDecoration(
        gradient: LinearGradient(begin: Alignment.bottomCenter, end: Alignment.topCenter, colors: [Color(0x0A00D4FF), Color(0x0000D4FF)]),
        border: Border(top: BorderSide(color: HmColors.border)),
      ),
      child: DefaultTextStyle(
        style: HmText.footer,
        child: Row(
          children: [
            // Long paths and build ids give way to the pills, never overflow.
            Flexible(flex: 3, child: BuildInfo(core: core)),
            const Spacer(),
            StatusPills(core: core),
            const Spacer(),
            if (vault.isNotEmpty)
              Flexible(
                flex: 3,
                child: Text(vault, maxLines: 1, overflow: TextOverflow.ellipsis, textAlign: TextAlign.end),
              ),
          ],
        ),
      ),
    );
  }
}

/// App and core versions; the core's turns amber when it is from
/// another build, e.g. an older app instance still running after an update.
class BuildInfo extends StatelessWidget {
  const BuildInfo({super.key, required this.core});

  final CoreState core;

  @override
  Widget build(BuildContext context) {
    final st = core.state;
    const dot = Opacity(opacity: 0.4, child: Text('  ·  '));
    final children = <Widget>[Text('MEGACHAT V${AppConfig.version}'), dot, Text('APP ${AppConfig.buildId}'.toUpperCase())];
    if (st != null && core.conn == Conn.ready) {
      final same = st.coreVersion == AppConfig.version && st.coreBuild == AppConfig.buildId;
      final coreBuild = st.coreBuild.isEmpty ? '?' : '${st.coreVersion} ${st.coreBuild}';
      children.addAll([
        dot,
        Text(
          'CORE $coreBuild${same ? '' : tr(' — differs from the app', ' — не совпадает с приложением')}'.toUpperCase(),
          style: same ? null : const TextStyle(color: HmColors.warning),
        ),
      ]);
    }
    return Wrap(crossAxisAlignment: WrapCrossAlignment.center, children: children);
  }
}

/// The online pill (§ 6.3); also shown on the phone's Settings (App. D).
/// Online: the core answers and the enabled accounts are connected.
class StatusPills extends StatelessWidget {
  const StatusPills({super.key, required this.core});

  final CoreState core;

  @override
  Widget build(BuildContext context) {
    final accounts = context.watch<MessengerCubit>().state.accounts.where((a) => a.enabled).toList();
    final up = accounts.where(isOnline).length;
    final online = core.conn == Conn.ready && up == accounts.length;
    final why = core.conn != Conn.ready
        ? tr('no connection to the core', 'нет связи с ядром')
        : accounts.isEmpty || !core.unlocked
        ? tr('online', 'в сети')
        : tr('$up of ${accounts.length} accounts online', 'в сети $up из ${accounts.length} аккаунтов');
    return HmPill(
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          HmDot(ok: online),
          const SizedBox(width: 6),
          Text(why.toUpperCase()),
        ],
      ),
    );
  }
}
