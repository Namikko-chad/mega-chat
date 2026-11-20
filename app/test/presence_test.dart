import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:megachat/core/theme/app_theme.dart';
import 'package:megachat/presentation/common/lists.dart';

Widget host(Widget child) => MaterialApp(
  theme: AppTheme.dark,
  home: Scaffold(body: Center(child: child)),
);

void main() {
  testWidgets('an online contact gets a dot on the avatar', (tester) async {
    await tester.pumpWidget(host(const HmAvatar(label: 'Ann Lee', online: true)));
    expect(find.bySemanticsLabel('online'), findsOneWidget);
    await tester.pumpWidget(host(const HmAvatar(label: 'Ann Lee')));
    expect(find.bySemanticsLabel('online'), findsNothing);
    // Beside the provider marker it moves to the top.
    await tester.pumpWidget(host(const HmAvatar(label: 'Ann Lee', online: true, corner: SizedBox.square(dimension: 16))));
    expect(find.bySemanticsLabel('online'), findsOneWidget);
  });
}
