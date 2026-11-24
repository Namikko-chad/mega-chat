import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:megachat/core/theme/app_theme.dart';
import 'package:megachat/data/local_files.dart';
import 'package:megachat/presentation/chats/attach.dart';
import 'package:megachat/presentation/common/hm.dart';

Widget host(Widget child) => MaterialApp(
  theme: AppTheme.dark,
  home: Scaffold(body: Center(child: child)),
);

void main() {
  setUp(() {
    // The system clipboard holds some text.
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger.setMockMethodCallHandler(SystemChannels.platform, (call) async {
      if (call.method == 'Clipboard.getData') return {'text': 'from clipboard'};
      if (call.method == 'Clipboard.hasStrings') return {'value': true};
      return null;
    });
  });

  Future<TextEditingController> paste(WidgetTester tester, Future<bool> Function() onPaste) async {
    final text = TextEditingController();
    await tester.pumpWidget(host(HmMessageField(controller: text, onSend: () {}, onPaste: onPaste)));
    await tester.tap(find.byType(TextField));
    await tester.pump();
    await tester.sendKeyDownEvent(LogicalKeyboardKey.controlLeft);
    await tester.sendKeyEvent(LogicalKeyboardKey.keyV);
    await tester.sendKeyUpEvent(LogicalKeyboardKey.controlLeft);
    await tester.pumpAndSettle();
    return text;
  }

  testWidgets('Ctrl+V takes files from the clipboard instead of its text', (tester) async {
    var asked = 0;
    final text = await paste(tester, () async {
      asked++;
      return true;
    });
    expect(asked, 1);
    expect(text.text, '');
  });

  testWidgets('Ctrl+V pastes the text when the clipboard has no files', (tester) async {
    var asked = 0;
    final text = await paste(tester, () async {
      asked++;
      return false;
    });
    expect(asked, 1);
    expect(text.text, 'from clipboard');
  });

  testWidgets('pending files are listed and can be dropped', (tester) async {
    final files = PendingFiles();
    await tester.pumpWidget(host(PendingFilesStrip(files: files)));
    expect(find.byType(HmAttachment), findsNothing);
    expect(files.add([LocalFile('notes.txt', 'text/plain', Uint8List.fromList('hello'.codeUnits)), LocalFile('b.bin', '', Uint8List(2048))]), isTrue);
    await tester.pump();
    expect(find.text('notes.txt'), findsOneWidget);
    expect(find.text('b.bin'), findsOneWidget);
    await tester.tap(find.byIcon(Icons.close).first);
    await tester.pump();
    expect(find.text('notes.txt'), findsNothing);
    expect(files.value.single.name, 'b.bin');
    expect(files.take().length, 1);
    expect(files.isEmpty, isTrue);
  });

  testWidgets('a message carries at most maxUpload of files', (tester) async {
    final files = PendingFiles();
    expect(files.add([LocalFile('big', '', Uint8List(maxUpload - 10))]), isTrue);
    expect(files.add([LocalFile('more', '', Uint8List(11))]), isFalse);
    expect(files.value.length, 1);
  });
}
