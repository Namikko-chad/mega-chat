import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import 'overlays.dart';

/// A link found in a text: where it is and what opening it opens.
typedef TextLink = ({int start, int end, Uri uri});

// A web address with a scheme or "www.", or an e-mail address. Stops at
// spaces, quotes and angle brackets; trailing punctuation is cut below.
final _linkRe = RegExp(r'''(?:https?://|www\.)[^\s<>"'«»]+|[\w.+-]+@[\w-]+(?:\.[\w-]+)*\.[a-zA-Z]{2,}''', caseSensitive: false);

/// The links in [text], in order. Punctuation ending a sentence and a
/// closing bracket without its opening one stay outside the link.
List<TextLink> findLinks(String text) {
  final out = <TextLink>[];
  for (final m in _linkRe.allMatches(text)) {
    var s = m.group(0)!;
    while (s.isNotEmpty) {
      final last = s[s.length - 1];
      if ('.,;:!?…'.contains(last)) {
        s = s.substring(0, s.length - 1);
      } else if (last == ')' && '('.allMatches(s).length < ')'.allMatches(s).length) {
        s = s.substring(0, s.length - 1);
      } else {
        break;
      }
    }
    final Uri? uri;
    if (s.contains('://')) {
      uri = Uri.tryParse(s);
    } else if (s.toLowerCase().startsWith('www.')) {
      uri = Uri.tryParse('https://$s');
    } else {
      uri = Uri(scheme: 'mailto', path: s);
    }
    if (uri == null || uri.host.isEmpty && uri.scheme != 'mailto') continue;
    out.add((start: m.start, end: m.start + s.length, uri: uri));
  }
  return out;
}

/// Opens a link in the browser or mail app of the device. Only web and
/// mail links: a message cannot make the app open anything else.
Future<void> openLink(Uri uri) async {
  if (!const {'http', 'https', 'mailto'}.contains(uri.scheme)) return;
  final ok = await launchUrl(uri, mode: LaunchMode.externalApplication).catchError((_) => false);
  if (!ok) Toasts.error(tr('Could not open the link', 'Не удалось открыть ссылку'), uri.toString());
}

/// [text] with its web and mail links underlined in the accent colour; a
/// tap opens one. Selectable inside a SelectionArea like a plain Text.
class HmLinkText extends StatefulWidget {
  const HmLinkText(this.text, {super.key, this.style});

  final String text;
  final TextStyle? style;

  @override
  State<HmLinkText> createState() => _HmLinkTextState();
}

class _HmLinkTextState extends State<HmLinkText> {
  final _taps = <TapGestureRecognizer>[];

  void _drop() {
    for (final r in _taps) {
      r.dispose();
    }
    _taps.clear();
  }

  @override
  void dispose() {
    _drop();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    _drop();
    final text = widget.text;
    final links = findLinks(text);
    if (links.isEmpty) return Text(text, style: widget.style);
    final linkStyle = TextStyle(color: HmColors.accent, decoration: TextDecoration.underline, decorationColor: HmColors.accent);
    final spans = <InlineSpan>[];
    var at = 0;
    for (final l in links) {
      if (l.start > at) spans.add(TextSpan(text: text.substring(at, l.start)));
      final tap = TapGestureRecognizer()..onTap = () => openLink(l.uri);
      _taps.add(tap);
      spans.add(TextSpan(text: text.substring(l.start, l.end), style: linkStyle, recognizer: tap, mouseCursor: SystemMouseCursors.click));
      at = l.end;
    }
    if (at < text.length) spans.add(TextSpan(text: text.substring(at)));
    return Text.rich(TextSpan(children: spans), style: widget.style);
  }
}
