import 'package:fixnum/fixnum.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:parley/data/rpc/gen/parley.pbgrpc.dart' as pb;
import 'package:parley/presentation/chats/messages.dart';

pb.Message msg(String id, String text, {String name = 'Ann', bool outgoing = false, List<String> files = const [], bool deleted = false}) => pb.Message(
  id: id,
  text: text,
  outgoing: outgoing,
  deleted: deleted,
  timeMs: Int64(int.parse(id)),
  sender: pb.User(id: name, name: name),
  attachments: [for (final f in files) pb.Attachment(kind: pb.AttachmentKind.ATTACHMENT_KIND_FILE, name: f)],
);

void main() {
  String stamp(pb.Message m) => 't${m.timeMs}';

  test('one copied message is its text, several carry author and time', () {
    expect(copyText([msg('1', 'hi')], stamp), 'hi');
    expect(copyText([msg('1', 'hi'), msg('2', '', deleted: true)], stamp), 'hi');
    expect(
      copyText([
        msg('1', 'hi'),
        msg('2', 'see', name: 'Me', outgoing: true, files: ['a.pdf']),
      ], stamp),
      'Ann, t1\nhi\n\nMe, t2\nsee\n[a.pdf]',
    );
  });

  test('a forward as text names the original author', () {
    expect(forwardText(msg('1', 'hi', files: ['a.pdf'])), 'Forwarded from Ann:\nhi\n[a.pdf]');
    expect(forwardText(msg('1', 'hello there'), fragment: 'there'), 'Forwarded from Ann:\nthere');
    final again = msg('1', 'hi', name: 'Bob')
      ..forwarded = true
      ..forwardedFrom = 'Carl';
    expect(forwardText(again), 'Forwarded from Carl:\nhi');
  });
}
