import 'package:flutter/material.dart';

import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/layout.dart';
import '../../core/theme/typography.dart';

/// A message bubble: own messages on the right in an accent tint, others
/// on the left on the surface. The corner next to the sender is sharp.
class HmBubble extends StatelessWidget {
  const HmBubble({super.key, required this.child, required this.outgoing, this.dim = false, this.boxKey});

  final Widget child;

  /// The key of the bubble's own box (not of the full-width row it is
  /// aligned in): menus anchor to it.
  final Key? boxKey;
  final bool outgoing;

  /// A deleted message: muted and without the tint.
  final bool dim;

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    final max = MediaQuery.sizeOf(context).width * (hm.mobile ? 0.82 : 0.5);
    const r = Radius.circular(12), sharp = Radius.circular(4);
    final tinted = outgoing && !dim;
    return Align(
      alignment: outgoing ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        key: boxKey,
        constraints: BoxConstraints(maxWidth: max.clamp(240.0, 720.0)),
        padding: const EdgeInsets.fromLTRB(12, 8, 12, 6),
        decoration: BoxDecoration(
          color: tinted ? HmColors.tint(HmColors.accent, HmColors.resting) : HmColors.surface,
          borderRadius: BorderRadius.only(topLeft: r, topRight: r, bottomLeft: outgoing ? r : sharp, bottomRight: outgoing ? sharp : r),
          border: Border.all(color: tinted ? HmColors.tint(HmColors.accent, HmColors.ring) : HmColors.border),
        ),
        child: DefaultTextStyle(
          style: HmText.body.copyWith(height: 1.45, color: dim ? HmColors.muted : HmColors.text),
          child: child,
        ),
      ),
    );
  }
}

/// A centred line between messages: a service message or a day.
class HmServiceNote extends StatelessWidget {
  const HmServiceNote(this.text, {super.key});

  final String text;

  @override
  Widget build(BuildContext context) => Center(
    child: Container(
      margin: const EdgeInsets.symmetric(vertical: 8),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      decoration: BoxDecoration(
        color: HmColors.tint(HmColors.text, HmColors.subtle),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: HmColors.border),
      ),
      child: Text(text, textAlign: TextAlign.center, style: HmText.footer),
    ),
  );
}

/// An attachment inside a bubble: kind icon, name and size.
class HmAttachment extends StatelessWidget {
  const HmAttachment({super.key, required this.icon, required this.name, this.detail, this.leading, this.trailing});

  final IconData icon;
  final String name;
  final String? detail;

  /// Shown instead of the icon (a small preview of a picture).
  final Widget? leading;

  /// After the size: an action on the file (download, remove).
  final Widget? trailing;

  @override
  Widget build(BuildContext context) => Container(
    margin: const EdgeInsets.only(top: 4, bottom: 2),
    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
    decoration: BoxDecoration(
      color: HmColors.bg,
      borderRadius: BorderRadius.circular(8),
      border: Border.all(color: HmColors.border),
    ),
    child: Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        leading ?? Icon(icon, size: 16, color: HmColors.accent),
        const SizedBox(width: 8),
        Flexible(
          child: Text(name, style: HmText.jb(12), maxLines: 1, overflow: TextOverflow.ellipsis),
        ),
        if (detail != null && detail!.isNotEmpty) ...[const SizedBox(width: 8), Text(detail!, style: HmText.footer)],
        if (trailing != null) ...[const SizedBox(width: 4), trailing!],
      ],
    ),
  );
}

/// The quoted message of a reply (in a bubble and above the composer).
class HmQuote extends StatelessWidget {
  const HmQuote({super.key, required this.title, required this.text, this.trailing});

  final String title;
  final String text;
  final Widget? trailing;

  @override
  Widget build(BuildContext context) => Container(
    margin: const EdgeInsets.only(bottom: 6),
    padding: const EdgeInsets.fromLTRB(8, 4, 4, 4),
    decoration: BoxDecoration(
      color: HmColors.tint(HmColors.accent, HmColors.subtle),
      border: const Border(left: BorderSide(color: HmColors.accent, width: 2)),
    ),
    child: Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Flexible(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                title,
                style: HmText.jb(11, weight: 600, color: HmColors.accent),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              Text(text, style: HmText.small, maxLines: 1, overflow: TextOverflow.ellipsis),
            ],
          ),
        ),
        ?trailing,
      ],
    ),
  );
}

/// A picture in a message (photo, video preview, sticker): sized by its
/// aspect ratio within [maxWidth] × [maxHeight], rounded, with a spinner
/// while [image] is null and an optional overlay (play button, duration).
class HmPhoto extends StatelessWidget {
  const HmPhoto({
    super.key,
    required this.image,
    this.width = 0,
    this.height = 0,
    this.maxWidth = 320,
    this.maxHeight = 320,
    this.failed = false,
    this.play = false,
    this.badge,
    this.plain = false,
    this.download = false,
    this.onTap,
  });

  final ImageProvider? image;

  /// The picture's own size in pixels (0: unknown, a square is used).
  final int width;
  final int height;
  final double maxWidth;
  final double maxHeight;

  /// It could not be loaded: an icon instead of the spinner.
  final bool failed;

  /// A video: a play button over it.
  final bool play;

  /// A short text in the corner (duration, size).
  final String? badge;

  /// A sticker: no frame and no background.
  final bool plain;

  /// Not loaded until a tap ([onTap] loads it): a download sign instead of
  /// the spinner.
  final bool download;
  final VoidCallback? onTap;

  Size _size() {
    final w = width > 0 ? width.toDouble() : maxWidth;
    final h = height > 0 ? height.toDouble() : maxWidth;
    final scale = [maxWidth / w, maxHeight / h, 1.0].reduce((a, b) => a < b ? a : b);
    var sw = w * scale, sh = h * scale;
    // Very thin pictures still get a usable box.
    if (sw < 120) sw = 120;
    if (sh < 80) sh = 80;
    return Size(sw, sh);
  }

  @override
  Widget build(BuildContext context) {
    final size = _size();
    Widget body;
    if (image != null) {
      body = Image(
        image: image!,
        width: size.width,
        height: size.height,
        fit: plain ? BoxFit.contain : BoxFit.cover,
        gaplessPlayback: true,
        filterQuality: FilterQuality.medium,
        errorBuilder: (_, _, _) => const Center(child: Icon(Icons.broken_image_outlined, color: HmColors.muted)),
      );
    } else {
      body = Center(
        child: download
            ? Container(
                width: 44,
                height: 44,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: HmColors.surface,
                  border: Border.all(color: HmColors.tint(HmColors.accent, HmColors.ring)),
                ),
                child: const Icon(Icons.download, color: HmColors.accent),
              )
            : failed
            ? Icon(play ? Icons.movie_outlined : Icons.image_outlined, size: 28, color: HmColors.muted)
            : const SizedBox.square(dimension: 20, child: CircularProgressIndicator(strokeWidth: 2, color: HmColors.accent)),
      );
    }
    final box = ClipRRect(
      borderRadius: BorderRadius.circular(plain ? 0 : 8),
      child: Container(
        width: size.width,
        height: size.height,
        color: plain ? null : HmColors.bg,
        child: Stack(
          fit: StackFit.expand,
          children: [
            body,
            if (play && !download)
              Center(
                child: Container(
                  width: 44,
                  height: 44,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: HmColors.bg.withValues(alpha: 0.7),
                    border: Border.all(color: HmColors.tint(HmColors.accent, HmColors.ring)),
                  ),
                  child: const Icon(Icons.play_arrow, color: HmColors.accent),
                ),
              ),
            if (badge != null && badge!.isNotEmpty)
              Positioned(
                left: 6,
                top: 6,
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                  decoration: BoxDecoration(color: HmColors.bg.withValues(alpha: 0.75), borderRadius: BorderRadius.circular(6)),
                  child: Text(badge!, style: HmText.footer.copyWith(color: HmColors.text)),
                ),
              ),
          ],
        ),
      ),
    );
    if (onTap == null) return box;
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: GestureDetector(onTap: onTap, child: box),
    );
  }
}

/// The reactions under a message: emoji with counts; the account's own
/// are lit. A tap toggles one ([onTap] null: display only).
class HmReactions extends StatelessWidget {
  const HmReactions({super.key, required this.items, this.onTap});

  /// (emoji, count, mine)
  final List<(String, int, bool)> items;
  final ValueChanged<String>? onTap;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(top: 6),
    child: Wrap(
      spacing: 4,
      runSpacing: 4,
      children: [
        for (final (emoji, count, mine) in items)
          MouseRegion(
            cursor: onTap == null ? MouseCursor.defer : SystemMouseCursors.click,
            child: GestureDetector(
              onTap: onTap == null ? null : () => onTap!(emoji),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 2),
                decoration: BoxDecoration(
                  color: mine ? HmColors.tint(HmColors.accent, HmColors.badge) : HmColors.tint(HmColors.text, HmColors.subtle),
                  borderRadius: BorderRadius.circular(10),
                  border: Border.all(color: mine ? HmColors.tint(HmColors.accent, HmColors.ring) : HmColors.border),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(emoji, style: emoji.startsWith(':') ? HmText.jb(11, color: HmColors.muted) : const TextStyle(fontSize: 14, height: 1.2)),
                    if (count > 0) ...[const SizedBox(width: 4), Text('$count', style: HmText.badge.copyWith(color: mine ? HmColors.accent : HmColors.muted))],
                  ],
                ),
              ),
            ),
          ),
      ],
    ),
  );
}

/// The line under a message that has a thread: who answered, how many
/// replies, when the last one came. A tap opens the thread.
class HmThreadSummary extends StatefulWidget {
  const HmThreadSummary({super.key, required this.label, required this.avatars, this.detail, required this.onTap});

  final String label;
  final List<Widget> avatars;
  final String? detail;
  final VoidCallback onTap;

  @override
  State<HmThreadSummary> createState() => _HmThreadSummaryState();
}

class _HmThreadSummaryState extends State<HmThreadSummary> {
  bool _hot = false;

  @override
  Widget build(BuildContext context) => MouseRegion(
    cursor: SystemMouseCursors.click,
    onEnter: (_) => setState(() => _hot = true),
    onExit: (_) => setState(() => _hot = false),
    child: GestureDetector(
      behavior: HitTestBehavior.opaque,
      onTap: widget.onTap,
      child: Container(
        margin: const EdgeInsets.only(top: 6),
        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
        decoration: BoxDecoration(
          color: _hot ? HmColors.tint(HmColors.accent, HmColors.resting) : Colors.transparent,
          borderRadius: BorderRadius.circular(6),
          border: Border.all(color: _hot ? HmColors.tint(HmColors.accent, HmColors.ring) : Colors.transparent),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            for (final a in widget.avatars) Padding(padding: const EdgeInsets.only(right: 4), child: a),
            if (widget.avatars.isNotEmpty) const SizedBox(width: 2),
            Text(widget.label, style: HmText.jb(12, weight: 600, color: HmColors.accent)),
            if (widget.detail != null && widget.detail!.isNotEmpty) ...[
              const SizedBox(width: 8),
              Flexible(
                child: Text(widget.detail!, style: HmText.footer, overflow: TextOverflow.ellipsis),
              ),
            ],
            const SizedBox(width: 4),
            const Icon(Icons.chevron_right, size: 14, color: HmColors.muted),
          ],
        ),
      ),
    ),
  );
}

/// A voice message or an audio file in a bubble: play / pause, a bar that
/// seeks on tap or drag, the time.
class HmAudio extends StatelessWidget {
  const HmAudio({
    super.key,
    required this.title,
    required this.playing,
    required this.position,
    required this.time,
    required this.onToggle,
    this.loading = false,
    this.onSeek,
    this.voice = false,
    this.width = 280,
  });

  final String title;
  final bool playing;
  final bool loading;

  /// 0..1 of the length.
  final double position;
  final String time;
  final VoidCallback onToggle;
  final ValueChanged<double>? onSeek;
  final bool voice;
  final double width;

  @override
  Widget build(BuildContext context) {
    void seek(double dx, double w) => onSeek?.call((dx / w).clamp(0.0, 1.0));
    return Container(
      width: width,
      margin: const EdgeInsets.only(top: 4, bottom: 2),
      padding: const EdgeInsets.fromLTRB(4, 4, 10, 4),
      decoration: BoxDecoration(
        color: HmColors.bg,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: HmColors.border),
      ),
      child: Row(
        children: [
          SizedBox.square(
            dimension: 36,
            child: loading
                ? const Padding(padding: EdgeInsets.all(9), child: CircularProgressIndicator(strokeWidth: 2))
                : IconButton(
                    padding: EdgeInsets.zero,
                    tooltip: playing ? tr('Pause', 'Пауза') : tr('Play', 'Слушать'),
                    icon: Icon(playing ? Icons.pause_circle_filled : Icons.play_circle_fill, size: 30, color: HmColors.accent),
                    onPressed: onToggle,
                  ),
          ),
          const SizedBox(width: 6),
          Expanded(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Row(
                  children: [
                    if (voice)
                      const Padding(
                        padding: EdgeInsets.only(right: 4),
                        child: Icon(Icons.mic_none, size: 13, color: HmColors.muted),
                      ),
                    Expanded(
                      child: Text(title, style: HmText.jb(12), maxLines: 1, overflow: TextOverflow.ellipsis),
                    ),
                    const SizedBox(width: 8),
                    Text(time, style: HmText.footer),
                  ],
                ),
                const SizedBox(height: 6),
                LayoutBuilder(
                  builder: (context, c) => MouseRegion(
                    cursor: onSeek == null ? MouseCursor.defer : SystemMouseCursors.click,
                    child: GestureDetector(
                      behavior: HitTestBehavior.opaque,
                      onTapDown: (d) => seek(d.localPosition.dx, c.maxWidth),
                      onHorizontalDragUpdate: (d) => seek(d.localPosition.dx, c.maxWidth),
                      child: Padding(
                        padding: const EdgeInsets.symmetric(vertical: 4),
                        child: ClipRRect(
                          borderRadius: BorderRadius.circular(2),
                          child: LinearProgressIndicator(value: position.clamp(0, 1), minHeight: 3, color: HmColors.accent, backgroundColor: HmColors.border),
                        ),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
