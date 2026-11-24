import 'package:fixnum/fixnum.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:megachat/core/theme/app_theme.dart';
import 'package:megachat/data/rpc/gen/megachat.pbgrpc.dart' as pb;
import 'package:megachat/presentation/chats/chat_view.dart';

pb.Message msg({String text = 'hello brave world'}) => pb.Message(
  id: '1',
  accountId: 'a',
  chatId: 'c',
  text: text,
  timeMs: Int64(1700000000000),
  sender: pb.User(id: 'u', name: 'Ann'),
);

Widget host(Widget child) => MaterialApp(
  theme: AppTheme.dark,
  // In a list, as in the chat: the tile is as tall as the message.
  home: Scaffold(
    body: ListView(padding: const EdgeInsets.only(top: 100), children: [child]),
  ),
);

void main() {
  // The test window (800 px) is a desktop one: text is selected right in
  // the bubble.
  testWidgets('the menu offers forwarding and picking', (tester) async {
    await tester.pumpWidget(host(MessageTile(message: msg(), quoted: null, showSender: true, gap: false, onReply: () {}, onForward: () {}, onSelect: () {})));
    await tester.longPress(find.text('Ann'));
    await tester.pumpAndSettle();
    for (final s in ['Reply', 'Copy text', 'Forward', 'Select']) {
      expect(find.text(s), findsOneWidget, reason: s);
    }
    expect(find.text('Copy selection'), findsNothing);
  });

  testWidgets('a right click on selected text offers to copy and quote it', (tester) async {
    String? quoted;
    await tester.pumpWidget(host(MessageTile(message: msg(), quoted: null, showSender: false, gap: false, onReply: () {}, onQuote: (q) => quoted = q)));
    final text = find.text('hello brave world');
    final box = tester.getRect(text);
    // Drag over the whole line with the mouse.
    final drag = await tester.startGesture(box.centerLeft + const Offset(1, 0), kind: PointerDeviceKind.mouse);
    await tester.pump();
    await drag.moveTo(box.centerRight - const Offset(1, 0));
    await tester.pump();
    await drag.up();
    await tester.pumpAndSettle();
    final click = await tester.startGesture(box.center, kind: PointerDeviceKind.mouse, buttons: kSecondaryMouseButton);
    await click.up();
    await tester.pumpAndSettle();
    expect(find.text('Copy selection'), findsOneWidget);
    expect(find.text('Forward selection'), findsOneWidget);
    await tester.tap(find.text('Quote'));
    await tester.pumpAndSettle();
    expect(quoted, 'hello brave world');
  });

  testWidgets('the menu opens at the pointer and stays inside the window', (tester) async {
    final own = msg()..outgoing = true;
    await tester.pumpWidget(host(MessageTile(message: own, quoted: null, showSender: false, gap: false, onReply: () {}, onForward: () {}, onSelect: () {})));
    // A right click on the own message (on the right side).
    final at = tester.getCenter(find.text('hello brave world'));
    final click = await tester.startGesture(at, kind: PointerDeviceKind.mouse, buttons: kSecondaryMouseButton);
    await click.up();
    await tester.pumpAndSettle();
    final reply = tester.getRect(find.text('Reply'));
    expect((reply.top - at.dy).abs(), lessThan(40), reason: 'opens next to the pointer, not at $reply');
    expect(reply.left, greaterThan(800 / 2 - 300), reason: 'not at the left edge');
    await tester.sendKeyEvent(LogicalKeyboardKey.escape);
    await tester.pumpAndSettle();
    // A long press near the bottom-right corner: it flips up and left.
    await tester.binding.setSurfaceSize(const Size(800, 600));
    addTearDown(() => tester.binding.setSurfaceSize(null));
    await tester.pumpWidget(host(MessageTile(message: own, quoted: null, showSender: false, gap: false, onReply: () {}, onForward: () {}, onSelect: () {})));
    final corner = tester.getBottomRight(find.text('hello brave world'));
    await tester.longPressAt(corner);
    await tester.pumpAndSettle();
    final select = tester.getRect(find.text('Select'));
    expect(select.right, lessThanOrEqualTo(800));
    expect(select.bottom, lessThanOrEqualTo(600));
    expect(find.text('Reply'), findsOneWidget);
  });

  testWidgets('picking: a tap picks, the check shows', (tester) async {
    var taps = 0;
    await tester.pumpWidget(host(MessageTile(message: msg(), quoted: null, showSender: false, gap: false, selecting: true, onSelect: () => taps++)));
    expect(find.byIcon(Icons.radio_button_unchecked), findsOneWidget);
    // The bubble ignores pointers while picking; the row takes the tap.
    await tester.tap(find.text('hello brave world'), warnIfMissed: false);
    expect(taps, 1);
    await tester.pumpWidget(host(MessageTile(message: msg(), quoted: null, showSender: false, gap: false, selecting: true, selected: true, onSelect: () {})));
    expect(find.byIcon(Icons.check_circle), findsOneWidget);
  });

  testWidgets('a forwarded message names its author, a quote shows the fragment', (tester) async {
    final m = msg()
      ..forwarded = true
      ..forwardedFrom = 'Carl'
      ..replyToId = '0'
      ..quote = 'the fragment';
    await tester.pumpWidget(
      host(
        MessageTile(
          message: m,
          quoted: msg(text: 'all of it, the fragment'),
          showSender: false,
          gap: false,
        ),
      ),
    );
    expect(find.text('Forwarded from Carl'), findsOneWidget);
    expect(find.text('the fragment'), findsOneWidget);
  });
}
