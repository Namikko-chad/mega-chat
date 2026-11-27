import 'package:fixnum/fixnum.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:parley/data/rpc/gen/parley.pbgrpc.dart' as pb;
import 'package:parley/presentation/calendar/calendar_cubit.dart';
import 'package:parley/presentation/mail/compose.dart';

void main() {
  test('addresses parse and print back', () {
    final l = parseAddresses('Anna <anna@example.com>; bob@x.org, ')!;
    expect(l.map((a) => '${a.name}|${a.email}'), ['Anna|anna@example.com', '|bob@x.org']);
    expect(formatAddresses(l), 'Anna <anna@example.com>, bob@x.org');
    expect(parseAddresses('not an address'), isNull);
    expect(parseAddresses(''), isEmpty);
  });

  test('an all-day event covers its local dates whatever the zone', () {
    final e = pb.CalendarEvent(allDay: true, startMs: dateMs(DateTime(2026, 11, 13)), endMs: dateMs(DateTime(2026, 11, 15)));
    final (s, end) = eventSpan(e);
    expect(s, DateTime(2026, 11, 13));
    expect(end, DateTime(2026, 11, 15));
    expect(onDay(e, DateTime(2026, 11, 14)), isTrue);
    expect(onDay(e, DateTime(2026, 11, 15)), isFalse);
  });

  test('a timed event is on the days it overlaps', () {
    final start = DateTime(2026, 11, 13, 23);
    final e = pb.CalendarEvent(startMs: Int64(start.millisecondsSinceEpoch), endMs: Int64(start.add(const Duration(hours: 2)).millisecondsSinceEpoch));
    expect(onDay(e, DateTime(2026, 11, 13)), isTrue);
    expect(onDay(e, DateTime(2026, 11, 14)), isTrue);
    expect(onDay(e, DateTime(2026, 11, 15)), isFalse);
  });

  test('the month grid starts on the Monday before the 1st', () {
    expect(gridStart(DateTime(2026, 11)), DateTime(2026, 10, 26));
    expect(gridStart(DateTime(2026, 6)), DateTime(2026, 6, 1)); // a Monday
  });

  test('every calendar gets a colour of its own', () {
    pb.Calendar cal(String a, String id, [int color = 0]) => pb.Calendar(accountId: a, id: id, color: color);
    final cals = [
      cal('g1', 'primary', 0x9fe1e7),
      cal('g2', 'primary', 0x9fe1e7), // same id and colour in another account
      cal('m1', '/cal/a/'),
      cal('m1', '/cal/b/'),
      for (var i = 0; i < 12; i++) cal('m2', 'c$i'),
    ];
    final m = calendarColors(cals);
    expect(m.length, cals.length);
    expect(m.values.map((c) => c.toARGB32()).toSet().length, cals.length);
    expect(m[(account: 'g1', calendar: 'primary')]!.toARGB32(), 0xFF9FE1E7);
    expect(CalendarState(calendars: cals).color((account: 'm1', calendar: '/cal/b/')), m[(account: 'm1', calendar: '/cal/b/')]);
  });
}
