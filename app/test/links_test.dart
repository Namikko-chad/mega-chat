import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:megachat/presentation/common/links.dart';

void main() {
  List<String> found(String s) => [for (final l in findLinks(s)) '${s.substring(l.start, l.end)} -> ${l.uri}'];

  test('web links with a scheme or www.', () {
    expect(found('https://user.github.io/'), ['https://user.github.io/ -> https://user.github.io/']);
    expect(found('see www.example.org/a?b=1#c, ok'), ['www.example.org/a?b=1#c -> https://www.example.org/a?b=1#c']);
  });

  test('sentence punctuation and unbalanced brackets stay outside', () {
    expect(found('Look: http://a.io/x.'), ['http://a.io/x -> http://a.io/x']);
    expect(found('(see https://a.io/x)'), ['https://a.io/x -> https://a.io/x']);
    expect(found('https://en.wikipedia.org/wiki/Go_(game)!'), ['https://en.wikipedia.org/wiki/Go_(game) -> https://en.wikipedia.org/wiki/Go_(game)']);
  });

  test('mail addresses', () {
    expect(found('write to ann.lee+x@corp.example.com.'), ['ann.lee+x@corp.example.com -> mailto:ann.lee+x@corp.example.com']);
  });

  test('plain text has no links', () {
    expect(findLinks('message 3 of Contact 1, v1.2 and a.b'), isEmpty);
    expect(findLinks('https://'), isEmpty);
  });

  testWidgets('a link is a tappable span, the rest plain text', (tester) async {
    await tester.pumpWidget(const MaterialApp(home: Scaffold(body: HmLinkText('go to https://a.io now'))));
    final spans = <TextSpan>[];
    tester.widget<RichText>(find.byType(RichText)).text.visitChildren((s) {
      if (s is TextSpan && s.text != null) spans.add(s);
      return true;
    });
    expect([for (final s in spans) s.text], ['go to ', 'https://a.io', ' now']);
    expect([for (final s in spans) s.recognizer is TapGestureRecognizer], [false, true, false]);
  });
}
