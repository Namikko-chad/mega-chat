import 'package:flutter_test/flutter_test.dart';
import 'package:parley/data/rpc/gen/parley.pbgrpc.dart' as pb;
import 'package:parley/presentation/chats/media.dart';

pb.Attachment _photo(String ref, {String thumb = 't'}) => pb.Attachment(kind: pb.AttachmentKind.ATTACHMENT_KIND_IMAGE, ref: ref, thumb: thumb);

void main() {
  test('the gallery has the photos of all messages in order', () {
    final photos = galleryPhotos([
      pb.Message(id: '1', attachments: [_photo('a'), _photo('b')]),
      pb.Message(id: '2', text: 'no photos'),
      pb.Message(id: '3', attachments: [_photo('c')]),
    ]);
    expect(photos.map((p) => p.ref), ['a', 'b', 'c']);
  });

  test('the gallery skips deleted messages, other kinds and photos it cannot show', () {
    final photos = galleryPhotos([
      pb.Message(id: '1', deleted: true, attachments: [_photo('gone')]),
      pb.Message(
        id: '2',
        attachments: [
          pb.Attachment(kind: pb.AttachmentKind.ATTACHMENT_KIND_VIDEO, ref: 'video', thumb: 't'),
          pb.Attachment(kind: pb.AttachmentKind.ATTACHMENT_KIND_STICKER, ref: 'sticker', thumb: 't'),
          _photo('no-thumb', thumb: ''),
          _photo(''),
          _photo('kept'),
        ],
      ),
    ]);
    expect(photos.map((p) => p.ref), ['kept']);
  });
}
