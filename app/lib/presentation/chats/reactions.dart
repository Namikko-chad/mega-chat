import 'package:flutter/material.dart';

import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../common/hm.dart';

/// Reactions offered in the picker: ones every messenger accepts.
const quickReactions = [
  '👍', '👎', '❤️', '🔥', '😂', '😮', '😢', '🙏', //
  '👏', '🎉', '🤔', '👀', '💯', '✅', '🚀', '😁', //
];

/// Asks for a reaction emoji.
Future<String?> pickReaction(BuildContext context) => showHmModal<String>(
  context,
  title: tr('Reaction', 'Реакция'),
  size: HmModalSize.sm,
  body: (ctx) => Wrap(
    spacing: 4,
    runSpacing: 4,
    children: [for (final e in quickReactions) _EmojiButton(emoji: e, onTap: () => Navigator.of(ctx).pop(e))],
  ),
);

class _EmojiButton extends StatefulWidget {
  const _EmojiButton({required this.emoji, required this.onTap});

  final String emoji;
  final VoidCallback onTap;

  @override
  State<_EmojiButton> createState() => _EmojiButtonState();
}

class _EmojiButtonState extends State<_EmojiButton> {
  bool _hot = false;

  @override
  Widget build(BuildContext context) => MouseRegion(
    cursor: SystemMouseCursors.click,
    onEnter: (_) => setState(() => _hot = true),
    onExit: (_) => setState(() => _hot = false),
    child: GestureDetector(
      onTap: widget.onTap,
      child: Container(
        width: 44,
        height: 44,
        alignment: Alignment.center,
        decoration: BoxDecoration(color: _hot ? HmColors.tint(HmColors.accent, HmColors.resting) : Colors.transparent, borderRadius: BorderRadius.circular(8)),
        child: Text(widget.emoji, style: const TextStyle(fontSize: 24)),
      ),
    ),
  );
}
