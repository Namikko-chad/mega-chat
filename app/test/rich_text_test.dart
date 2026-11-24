import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:megachat/data/rpc/gen/megachat.pbgrpc.dart' as pb;
import 'package:megachat/presentation/chats/messages.dart';
import 'package:megachat/presentation/common/hm.dart';

pb.TextSpan span(pb.TextStyle s, int o, int l, {String url = ''}) => pb.TextSpan(style: s, offset: o, length: l, url: url);

Future<void> show(WidgetTester t, Widget w) => t.pumpWidget(MaterialApp(home: Scaffold(body: w)));

TextSpan? styled(WidgetTester t, String piece) {
  for (final r in t.widgetList<RichText>(find.byType(RichText))) {
    TextSpan? hit;
    r.text.visitChildren((s) {
      if (s is TextSpan && s.text == piece) hit = s;
      return hit == null;
    });
    if (hit != null) return hit;
  }
  return null;
}

void main() {
  testWidgets('styles the parts of a text', (t) async {
    await show(
      t,
      HmRichText(
        'a bold site',
        spans: [
          span(pb.TextStyle.TEXT_STYLE_BOLD, 2, 4),
          span(pb.TextStyle.TEXT_STYLE_LINK, 7, 4, url: 'https://x.io'),
        ],
      ),
    );
    expect(styled(t, 'bold')!.style!.fontWeight, FontWeight.w700);
    expect(styled(t, 'site')!.recognizer, isNotNull);
  });

  testWidgets('a spoiler opens on a tap', (t) async {
    await show(t, HmRichText('see secret', spans: [span(pb.TextStyle.TEXT_STYLE_SPOILER, 4, 6)]));
    expect(styled(t, 'secret')!.style!.backgroundColor, isNotNull);
    await t.runAsync(() async => (styled(t, 'secret')!.recognizer as dynamic).onTap());
    await t.pump();
    expect(styled(t, 'secret')!.style!.backgroundColor, isNull);
  });

  testWidgets('quotes and code are blocks; broken spans are ignored', (t) async {
    await show(
      t,
      HmRichText(
        'q\nx=1\nend',
        spans: [span(pb.TextStyle.TEXT_STYLE_QUOTE, 0, 1), span(pb.TextStyle.TEXT_STYLE_PRE, 2, 3), span(pb.TextStyle.TEXT_STYLE_BOLD, 50, 9)],
      ),
    );
    expect(find.text('x=1'), findsOneWidget);
    expect(find.text('end'), findsOneWidget);
    expect(t.takeException(), isNull);
  });

  test('links: named ones and bare ones, none inside code', () {
    final links = formattedLinks('site https://a.io `https://b.io`', [
      span(pb.TextStyle.TEXT_STYLE_LINK, 0, 4, url: 'https://x.io'),
      span(pb.TextStyle.TEXT_STYLE_CODE, 18, 14),
    ]);
    expect(links.map((l) => l.uri.host), ['x.io', 'a.io']);
  });

  test('an album keeps each caption\'s styles at its place', () {
    final g = MessageGroup([
      pb.Message(id: '1', group: 'g', text: 'one', format: [span(pb.TextStyle.TEXT_STYLE_BOLD, 0, 3)]),
      pb.Message(id: '2', group: 'g', text: 'two', format: [span(pb.TextStyle.TEXT_STYLE_ITALIC, 0, 3)]),
    ]);
    expect(g.message.text, 'one\ntwo');
    expect(g.message.format.map((s) => s.offset), [0, 4]);
  });

  test('editing starts from the Markdown source', () {
    expect(editSource(pb.Message(text: 'bold', markdown: '**bold**')), '**bold**');
    expect(editSource(pb.Message(text: 'plain')), 'plain');
  });
}
