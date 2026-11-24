import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/typography.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import 'media.dart';

/// Whether the account's messages can mention people (CAPABILITY_MENTIONS).
bool canMention(pb.Account? a) => (a?.capabilities ?? const []).contains(pb.Capability.CAPABILITY_MENTIONS);

/// "@" in a composer: finds the word being typed after it, asks the core
/// for the chat's people who fit, and puts the picked one into the text as
/// "@username" (or "@Name"). What was picked goes with the message as
/// mentions, so the messenger makes them real mentions (Slack <@id>,
/// Telegram's mention of a user, VK Teams @[id]).
class MentionPicker extends ChangeNotifier {
  MentionPicker({required this.api, required this.account, required this.chat, required this.text}) {
    text.addListener(_changed);
  }

  final pb.MegaChatClient Function() api; // the client of now (it changes after a reconnect)
  final String account;
  final String chat;
  final TextEditingController text;

  /// The people that fit what is typed; empty: the list is closed.
  List<pb.User> users = const [];

  /// The highlighted one (arrows move it, Enter or Tab picks it).
  int index = 0;

  int? _at; // where the "@" being completed is
  String _query = '';
  Timer? _wait;
  int _asked = 0;
  bool _dismissed = false; // Esc: closed until another "@"
  final _picked = <({String shown, pb.User user})>[];

  bool get open => users.isNotEmpty;

  @override
  void dispose() {
    text.removeListener(_changed);
    _wait?.cancel();
    super.dispose();
  }

  static bool _space(String c) => c.trim().isEmpty;

  void _changed() {
    final v = text.value;
    final cursor = v.selection.baseOffset;
    int? at;
    if (v.selection.isCollapsed && cursor > 0 && cursor <= v.text.length) {
      for (var i = cursor - 1; i >= 0 && cursor - i <= 40; i--) {
        final c = v.text[i];
        if (c == '@') {
          if (i == 0 || _space(v.text[i - 1])) at = i;
          break;
        }
        if (_space(c)) break;
      }
    }
    if (at == null) {
      _dismissed = false;
      return _close();
    }
    if (at != _at) _dismissed = false;
    final q = v.text.substring(at + 1, cursor);
    if (at == _at && q == _query) return;
    _at = at;
    _query = q;
    if (_dismissed) return;
    _wait?.cancel();
    _wait = Timer(const Duration(milliseconds: 150), () => _ask(q));
  }

  Future<void> _ask(String q) async {
    final n = ++_asked;
    try {
      final r = await api().members(pb.MembersRequest(accountId: account, chatId: chat, query: q, limit: 8));
      if (n != _asked || _at == null || _dismissed) return;
      users = r.users;
      index = 0;
      notifyListeners();
    } catch (_) {
      if (n == _asked) _close();
    }
  }

  void _close() {
    _wait?.cancel();
    _asked++;
    _at = null;
    _query = '';
    if (users.isEmpty) return;
    users = const [];
    notifyListeners();
  }

  /// Esc: closes the list until another "@".
  void dismiss() {
    _dismissed = true;
    _close();
  }

  void move(int by) {
    if (!open) return;
    index = (index + by) % users.length;
    notifyListeners();
  }

  /// Puts [users][i] (the highlighted one) in place of "@query".
  void pick([int? i]) {
    final at = _at;
    if (!open || at == null) return;
    final u = users[i ?? index];
    final shown = '@${u.username.isNotEmpty ? u.username : u.name}';
    final v = text.value;
    final end = v.selection.baseOffset.clamp(at, v.text.length);
    final next = '${v.text.substring(0, at)}$shown ${v.text.substring(end)}';
    _picked.add((shown: shown, user: u));
    _close();
    text.value = TextEditingValue(text: next, selection: TextSelection.collapsed(offset: at + shown.length + 1));
  }

  /// Keys while the list is open: arrows, Enter / Tab, Esc.
  KeyEventResult onKey(KeyEvent e) {
    if (!open || e is KeyUpEvent) return KeyEventResult.ignored;
    switch (e.logicalKey) {
      case LogicalKeyboardKey.arrowDown:
        move(1);
      case LogicalKeyboardKey.arrowUp:
        move(-1);
      case LogicalKeyboardKey.enter || LogicalKeyboardKey.numpadEnter || LogicalKeyboardKey.tab:
        if (e is KeyDownEvent) pick();
      case LogicalKeyboardKey.escape:
        dismiss();
      default:
        return KeyEventResult.ignored;
    }
    return KeyEventResult.handled;
  }

  /// The mentions of [sent] (the text as sent): each pick is found once,
  /// in order, as a whole word; a pick the user erased is dropped.
  List<pb.Mention> mentionsIn(String sent) {
    final out = <pb.Mention>[];
    final from = <String, int>{};
    for (final p in _picked) {
      var i = sent.indexOf(p.shown, from[p.shown] ?? 0);
      while (i >= 0) {
        final end = i + p.shown.length;
        final whole = end == sent.length || !RegExp(r'[\p{L}\p{N}_]', unicode: true).hasMatch(sent[end]);
        if (whole) break;
        i = sent.indexOf(p.shown, i + 1);
      }
      if (i < 0) continue;
      from[p.shown] = i + p.shown.length;
      out.add(pb.Mention(offset: i, length: p.shown.length, user: p.user));
    }
    out.sort((a, b) => a.offset.compareTo(b.offset));
    return out;
  }

  /// The message went: what was picked is done with.
  void sent() => _picked.clear();
}

/// The people that fit what is typed after "@", over the composer.
class MentionList extends StatelessWidget {
  const MentionList({super.key, required this.picker});

  final MentionPicker picker;

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: picker,
    builder: (context, _) {
      if (!picker.open) return const SizedBox.shrink();
      return Container(
        margin: const EdgeInsets.only(bottom: 8),
        constraints: const BoxConstraints(maxHeight: 280),
        decoration: BoxDecoration(
          color: HmColors.bg,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: HmColors.border),
        ),
        child: Semantics(
          label: tr('People to mention', 'Кого упомянуть'),
          child: ListView.builder(
            shrinkWrap: true,
            padding: const EdgeInsets.symmetric(vertical: 4),
            itemCount: picker.users.length,
            itemBuilder: (context, i) {
              final u = picker.users[i];
              final lit = i == picker.index;
              return MouseRegion(
                cursor: SystemMouseCursors.click,
                child: GestureDetector(
                  behavior: HitTestBehavior.opaque,
                  onTap: () => picker.pick(i),
                  child: Container(
                    color: lit ? HmColors.tint(HmColors.accent, 0.1) : Colors.transparent,
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                    child: Row(
                      children: [
                        MediaAvatar(account: picker.account, photo: u.avatarUrl, label: u.name, size: 24, online: u.online, showOffline: true),
                        const SizedBox(width: 10),
                        Flexible(
                          child: Text(
                            u.name.isNotEmpty ? u.name : u.username,
                            style: HmText.jb(13, weight: lit ? 600 : 400, color: lit ? HmColors.accent : HmColors.text),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        if (u.username.isNotEmpty && u.username != u.name) ...[
                          const SizedBox(width: 8),
                          Text('@${u.username}', style: HmText.footer, maxLines: 1),
                        ],
                      ],
                    ),
                  ),
                ),
              );
            },
          ),
        ),
      );
    },
  );
}
