import 'dart:async';
import 'dart:io';

import 'package:path/path.dart' as p;

import '../../core/desktop/desktop_window.dart';
import '../../core/i18n.dart';
import '../../core/settings/app_settings.dart';
import '../../data/rpc/gen/parley.pbgrpc.dart' as pb;
import '../../data/rpc/native_core.dart';
import '../shell/app_shell.dart' show AppShell, Section;
import 'core_cubit.dart';
import 'messenger_cubit.dart';

/// System notifications on desktops: a new message the core marks as
/// worth one (Event.notify: someone else wrote, the chat is not muted or
/// the message mentions the user) and new mail, while the user does not
/// look at it — the window has no focus, or another chat is open. One
/// notification per chat, a newer message replaces it; opening the chat
/// takes it away, a click opens it. Android does this in Kotlin.
class DesktopNotifier {
  DesktopNotifier(this._core, this._messenger) {
    _events = _core.events.listen(_onEvent);
    _open = _messenger.stream.listen(_onMessenger);
    DesktopWindow.instance.focused.addListener(_onFocus);
    try {
      NativeCore.instance.onNotificationClick(_onClick);
    } catch (_) {
      // No native core (tests): nothing is shown either.
    }
  }

  final CoreCubit _core;
  final MessengerCubit _messenger;
  late final StreamSubscription<pb.Event> _events;
  late final StreamSubscription<MessengerState> _open;
  final _shown = <String>{};

  static bool get supported => Platform.isLinux || Platform.isWindows;

  /// The app's PNG icon next to the executable (Linux).
  static final String? _icon = () {
    if (!Platform.isLinux) return null;
    final f = File(p.join(p.dirname(Platform.resolvedExecutable), 'data', 'flutter_assets', 'assets', 'icon', 'parley.png'));
    return f.existsSync() ? f.path : null;
  }();

  // Tags name what a click opens: "chat\n<account>\n<chat>[\n<thread>]",
  // "mail\n<account>". Ids never hold a newline.
  static String _chatTag(String account, String chat, [String thread = '']) => ['chat', account, chat, if (thread.isNotEmpty) thread].join('\n');

  void _onEvent(pb.Event ev) {
    switch (ev.kind) {
      case pb.Event_Kind.KIND_MESSAGE_NEW || pb.Event_Kind.KIND_THREAD_MESSAGE:
        if (ev.notify) _message(ev);
      case pb.Event_Kind.KIND_MAIL_NEW:
        _mail(ev);
      case pb.Event_Kind.KIND_CHAT when ev.chat.unread == 0:
        // Read here or on another device: the notification is stale.
        _close(_chatTag(ev.chat.accountId, ev.chat.id));
      default:
    }
  }

  bool get _looking => DesktopWindow.instance.focused.value;

  void _message(pb.Event ev) {
    final m = ev.message;
    final mode = AppSettings.instance.notify;
    if (mode == NotifyMode.off || (mode == NotifyMode.mentions && !m.mentioned)) return;
    final key = (account: m.accountId, chat: m.chatId);
    final s = _messenger.state;
    final thread = ev.kind == pb.Event_Kind.KIND_THREAD_MESSAGE ? m.inThread : '';
    if (_looking && s.open == key && (thread.isEmpty || s.thread?.thread == thread)) return;
    final chat = ev.hasChat() ? ev.chat : s.chats[key];
    final title = chat?.title.isNotEmpty == true ? chatTitle(chat!) : m.sender.name;
    var text = m.text;
    if (text.isEmpty && m.attachments.isNotEmpty) text = '📎 ${m.attachments.first.name}'.trim();
    final from = m.sender.name;
    _show(
      _chatTag(m.accountId, m.chatId, thread),
      m.mentioned ? '$title · ${tr('mention', 'упоминание')}' : title,
      from.isEmpty || from == title ? text : '$from: $text',
    );
  }

  void _mail(pb.Event ev) {
    final mode = AppSettings.instance.notify;
    if (mode == NotifyMode.off || _looking) return;
    final m = ev.mail;
    _show('mail\n${ev.accountId}', m.from.name.isNotEmpty ? m.from.name : m.from.email, m.subject);
  }

  void _show(String tag, String title, String body) {
    try {
      final core = NativeCore.instance;
      if (!core.owner) return; // another instance of the app runs the core and notifies
      core.notify(tag: tag, title: title, body: body, icon: _icon, entry: 'dev.selfhold.parley');
      _shown.add(tag);
    } catch (_) {}
  }

  void _close(String tag) {
    if (!_shown.remove(tag)) return;
    try {
      NativeCore.instance.closeNotification(tag);
    } catch (_) {}
  }

  void _onFocus() => _onMessenger(_messenger.state);

  void _onMessenger(MessengerState s) {
    final open = s.open;
    if (open == null || !_looking) return;
    _close(_chatTag(open.account, open.chat));
    final t = s.thread;
    if (t != null) _close(_chatTag(t.account, t.chat, t.thread));
  }

  void _onClick(String tag) {
    _shown.remove(tag);
    final parts = tag.split('\n');
    unawaited(DesktopWindow.instance.show());
    switch (parts) {
      case ['chat', final account, final chat]:
        AppShell.requests.add(Section.chats);
        _messenger.openChat((account: account, chat: chat));
      case ['chat', final account, final chat, final thread]:
        AppShell.requests.add(Section.chats);
        _messenger.openThread((account: account, chat: chat, thread: thread));
      case ['mail', _]:
        AppShell.requests.add(Section.mail);
    }
  }

  Future<void> close() async {
    DesktopWindow.instance.focused.removeListener(_onFocus);
    await _events.cancel();
    await _open.cancel();
  }
}
