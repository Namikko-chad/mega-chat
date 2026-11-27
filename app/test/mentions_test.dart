import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:parley/data/rpc/gen/parley.pbgrpc.dart' as pb;
import 'package:parley/presentation/chats/mentions.dart';

void main() {
  test('a picked person goes into the text and out as a mention', () {
    final c = TextEditingController();
    final p = MentionPicker(api: () => throw StateError('no core'), account: 'a', chat: 'c', text: c);
    addTearDown(p.dispose);
    c.value = const TextEditingValue(text: '😀 hi @an', selection: TextSelection.collapsed(offset: 9));
    p.users = [pb.User(id: 'U1', name: 'Ann Lee', username: 'ann'), pb.User(id: 'U2', name: 'Andy')];
    p.move(1);
    p.move(1); // wraps to the first
    p.pick();
    expect(c.text, '😀 hi @ann ');
    expect(c.selection.baseOffset, 11);
    expect(p.open, isFalse);
    // Typed on: "@annette" is not the pick, the second "@ann" is.
    c.text = '😀 hi @annette and @ann!';
    final ms = p.mentionsIn(c.text);
    expect(ms.length, 1);
    expect(ms.single.offset, 19);
    expect(ms.single.length, 4);
    expect(ms.single.user.id, 'U1');
    p.sent();
    expect(p.mentionsIn(c.text), isEmpty);
  });

  test('no list for an "@" inside a word', () {
    final c = TextEditingController();
    final p = MentionPicker(api: () => throw StateError('no core'), account: 'a', chat: 'c', text: c);
    addTearDown(p.dispose);
    c.value = const TextEditingValue(text: 'mail@an', selection: TextSelection.collapsed(offset: 7));
    p.users = [pb.User(id: 'U1', name: 'Ann')];
    p.pick(); // nothing to complete
    expect(c.text, 'mail@an');
  });
}
