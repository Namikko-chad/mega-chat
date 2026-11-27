import 'package:fixnum/fixnum.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:parley/data/rpc/gen/parley.pbgrpc.dart' as pb;
import 'package:parley/presentation/core/messenger_cubit.dart';

pb.Chat chat(String id, int at, {bool archived = false, bool pinned = false, bool muted = false, int unread = 0, List<String> folders = const []}) =>
    pb.Chat(accountId: 'a', id: id, title: id, updatedMs: Int64(at), archived: archived, pinned: pinned, muted: muted, unread: unread, folders: folders);

void main() {
  final chats = [
    chat('old', 1),
    chat('new', 5),
    chat('pin', 2, pinned: true),
    chat('arch', 9, archived: true, folders: ['work']),
    chat('work', 3, folders: ['work'], unread: 1),
    chat('hand', 0),
  ];
  final work = pb.ChatFolder(accountId: 'a', id: 'work', always: ['hand'], pinned: ['hand'], excludeArchived: true);
  MessengerState state(String view) => MessengerState(chats: {for (final c in chats) chatKey(c): c}, folders: [work], view: view);

  test('the main list leaves the archive out and puts pinned chats first', () {
    expect(state('').shown.map((c) => c.id), ['pin', 'new', 'work', 'old', 'hand']);
    expect(state('').archived.map((c) => c.id), ['arch']);
  });

  test('the archive shows only archived chats', () {
    expect(state(archiveView).shown.map((c) => c.id), ['arch']);
  });

  test('a folder: its pins first, rules and exclusions applied', () {
    expect(state(folderView(work)).shown.map((c) => c.id), ['hand', 'work']);
    expect(inFolder(work, chats[3]), isFalse); // archived, excluded
  });

  test('the saved chat has the app’s name', () {
    expect(chatTitle(pb.Chat(title: 'Me', saved: true)), isNot('Me'));
  });
}
