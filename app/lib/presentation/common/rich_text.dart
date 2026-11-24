import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';

import '../../core/theme/colors.dart';
import '../../core/theme/typography.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import 'links.dart';

/// The links of a formatted text: those its spans name, then the addresses
/// written in it outside code and named links.
List<TextLink> formattedLinks(String text, List<pb.TextSpan> spans) {
  final out = <TextLink>[];
  for (final s in spans) {
    if (s.style != pb.TextStyle.TEXT_STYLE_LINK) continue;
    final uri = Uri.tryParse(s.url);
    if (uri != null) out.add((start: s.offset, end: s.offset + s.length, uri: uri));
  }
  bool covered(TextLink l) => spans.any(
    (s) =>
        const {pb.TextStyle.TEXT_STYLE_LINK, pb.TextStyle.TEXT_STYLE_CODE, pb.TextStyle.TEXT_STYLE_PRE}.contains(s.style) &&
        s.offset < l.end &&
        s.offset + s.length > l.start,
  );
  out.addAll(findLinks(text).where((l) => !covered(l)));
  out.sort((a, b) => a.start.compareTo(b.start));
  return out;
}

/// [text] styled by [spans]: emphasis, code, links, spoilers (a tap shows
/// one), headings, quotes and code blocks. Selectable inside a SelectionArea.
class HmRichText extends StatefulWidget {
  const HmRichText(this.text, {super.key, this.spans = const [], this.style});

  final String text;
  final List<pb.TextSpan> spans;
  final TextStyle? style;

  @override
  State<HmRichText> createState() => _HmRichTextState();
}

class _HmRichTextState extends State<HmRichText> {
  final _taps = <TapGestureRecognizer>[];
  final _shown = <int>{}; // offsets of the spoilers opened

  void _drop() {
    for (final r in _taps) {
      r.dispose();
    }
    _taps.clear();
  }

  @override
  void didUpdateWidget(HmRichText old) {
    super.didUpdateWidget(old);
    if (old.text != widget.text) _shown.clear();
  }

  @override
  void dispose() {
    _drop();
    super.dispose();
  }

  static bool _isBlock(pb.TextStyle s) => s == pb.TextStyle.TEXT_STYLE_PRE || s == pb.TextStyle.TEXT_STYLE_QUOTE;

  @override
  Widget build(BuildContext context) {
    _drop();
    final text = widget.text;
    final spans = [
      for (final s in widget.spans)
        if (s.length > 0 && s.offset >= 0 && s.offset + s.length <= text.length) s,
    ];
    if (spans.isEmpty) return HmLinkText(text, style: widget.style);
    final links = formattedLinks(text, spans);
    final blocks = spans.where((s) => _isBlock(s.style)).toList()..sort((a, b) => a.offset.compareTo(b.offset));
    final children = <Widget>[];
    var at = 0;
    void plain(int lo, int hi) {
      // The line breaks around a block are the block's own margins.
      if (lo < hi && text[lo] == '\n') lo++;
      if (lo < hi && text[hi - 1] == '\n') hi--;
      if (lo < hi) children.add(Text.rich(_inline(lo, hi, spans, links), style: widget.style));
    }

    for (final b in blocks) {
      if (b.offset < at) continue; // overlaps the block before it
      plain(at, b.offset);
      final end = b.offset + b.length;
      children.add(b.style == pb.TextStyle.TEXT_STYLE_PRE ? _code(text.substring(b.offset, end)) : _quote(_inline(b.offset, end, spans, links)));
      at = end;
    }
    plain(at, text.length);
    if (children.length == 1 && blocks.isEmpty) return children.single;
    return Column(crossAxisAlignment: CrossAxisAlignment.start, mainAxisSize: MainAxisSize.min, children: children);
  }

  Widget _code(String body) => Container(
    margin: const EdgeInsets.symmetric(vertical: 4),
    padding: const EdgeInsets.all(8),
    decoration: BoxDecoration(
      color: HmColors.bg,
      border: Border.all(color: HmColors.border),
      borderRadius: BorderRadius.circular(4),
    ),
    child: SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Text(body, style: (widget.style ?? HmText.body).copyWith(fontFamily: HmText.mono, height: 1.4)),
    ),
  );

  Widget _quote(InlineSpan body) => Container(
    margin: const EdgeInsets.symmetric(vertical: 4),
    padding: const EdgeInsets.only(left: 8),
    decoration: const BoxDecoration(
      border: Border(left: BorderSide(color: HmColors.accent, width: 2)),
    ),
    child: Text.rich(body, style: (widget.style ?? HmText.body).copyWith(color: HmColors.muted)),
  );

  /// text[lo:hi] cut wherever a span or a link starts or ends.
  TextSpan _inline(int lo, int hi, List<pb.TextSpan> spans, List<TextLink> links) {
    final text = widget.text;
    final cuts = <int>{lo, hi};
    void cut(int a, int b) {
      if (b <= lo || a >= hi) return;
      cuts.add(a < lo ? lo : a);
      cuts.add(b > hi ? hi : b);
    }

    for (final s in spans) {
      if (!_isBlock(s.style)) cut(s.offset, s.offset + s.length);
    }
    for (final l in links) {
      cut(l.start, l.end);
    }
    final points = cuts.toList()..sort();
    final out = <InlineSpan>[];
    for (var i = 0; i + 1 < points.length; i++) {
      final a = points[i], b = points[i + 1];
      final on = spans.where((s) => !_isBlock(s.style) && s.offset <= a && s.offset + s.length >= b).toList();
      final link = links.where((l) => l.start <= a && l.end >= b).firstOrNull;
      final spoiler = on.where((s) => s.style == pb.TextStyle.TEXT_STYLE_SPOILER).firstOrNull;
      final hidden = spoiler != null && !_shown.contains(spoiler.offset);
      GestureRecognizer? tap;
      if (hidden) {
        tap = TapGestureRecognizer()..onTap = () => setState(() => _shown.add(spoiler.offset));
      } else if (link != null) {
        tap = TapGestureRecognizer()..onTap = () => openLink(link.uri);
      }
      if (tap is TapGestureRecognizer) _taps.add(tap);
      out.add(
        TextSpan(
          text: text.substring(a, b),
          style: _styleOf(on, link != null, hidden),
          recognizer: tap,
          mouseCursor: tap != null ? SystemMouseCursors.click : MouseCursor.defer,
        ),
      );
    }
    return TextSpan(children: out);
  }

  TextStyle _styleOf(List<pb.TextSpan> on, bool link, bool hidden) {
    var st = const TextStyle();
    final lines = <TextDecoration>[];
    for (final s in on) {
      switch (s.style) {
        case pb.TextStyle.TEXT_STYLE_BOLD:
          st = st.copyWith(fontWeight: FontWeight.w700);
        case pb.TextStyle.TEXT_STYLE_ITALIC:
          st = st.copyWith(fontStyle: FontStyle.italic);
        case pb.TextStyle.TEXT_STYLE_UNDERLINE:
          lines.add(TextDecoration.underline);
        case pb.TextStyle.TEXT_STYLE_STRIKE:
          lines.add(TextDecoration.lineThrough);
        case pb.TextStyle.TEXT_STYLE_CODE:
          st = st.copyWith(fontFamily: HmText.mono, backgroundColor: HmColors.border);
        case pb.TextStyle.TEXT_STYLE_HEADING:
          st = st.copyWith(fontWeight: FontWeight.w700, fontSize: (widget.style?.fontSize ?? HmText.body.fontSize ?? 13) + 3);
        default:
      }
    }
    if (link) {
      lines.add(TextDecoration.underline);
      st = st.copyWith(color: HmColors.accent, decorationColor: HmColors.accent);
    }
    if (lines.isNotEmpty) st = st.copyWith(decoration: TextDecoration.combine(lines));
    if (hidden) st = st.copyWith(color: HmColors.muted, backgroundColor: HmColors.muted, decoration: TextDecoration.none);
    return st;
  }
}
