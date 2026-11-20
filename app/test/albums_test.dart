import 'package:flutter_test/flutter_test.dart';
import 'package:megachat/data/rpc/gen/megachat.pbgrpc.dart' as pb;
import 'package:megachat/presentation/chats/messages.dart';

pb.Message msg(String id, {String group = '', String text = '', String sender = 'a', bool deleted = false, int reactions = 0}) => pb.Message(
  id: id,
  group: group,
  text: text,
  deleted: deleted,
  sender: pb.User(id: sender),
  attachments: [pb.Attachment(kind: pb.AttachmentKind.ATTACHMENT_KIND_IMAGE, ref: 'r$id')],
  reactions: [for (var i = 0; i < reactions; i++) pb.Reaction(emoji: '👍', count: 1)],
);

void main() {
  test('an album is one message with every picture and the caption', () {
    final g = groupAlbums([
      msg('1'),
      msg('2', group: 'g'),
      msg('3', group: 'g', text: 'caption', reactions: 1),
      msg('4', group: 'g', deleted: true),
      msg('5', group: 'g'),
      msg('6', group: 'h'),
      msg('7', group: 'h', sender: 'b'),
    ]);
    expect([for (final x in g) x.parts.map((p) => p.id).join(',')], ['1', '2,3,4,5', '6', '7']);
    final album = g[1].message;
    expect(album.id, '3');
    expect(album.text, 'caption');
    expect(album.reactions, hasLength(1));
    expect([for (final a in album.attachments) a.ref], ['r2', 'r3', 'r5']);
    expect(g[1].newest.id, '5');
    expect(g[1].has('4'), isTrue);
  });

  test('a single message is shown as it is', () {
    final m = msg('1', group: 'g');
    expect(identical(groupAlbums([m]).single.message, m), isTrue);
  });
}
