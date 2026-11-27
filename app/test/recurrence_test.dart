import 'package:flutter_test/flutter_test.dart';
import 'package:parley/data/rpc/gen/parley.pbgrpc.dart' as pb;
import 'package:parley/presentation/calendar/recurrence.dart';

void main() {
  test('rules in words', () {
    expect(describeRule('FREQ=DAILY'), 'Every day');
    expect(describeRule('FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR'), 'Every weekday');
    expect(describeRule('FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE'), 'Every 2 weeks on Mon, Wed');
    expect(describeRule('FREQ=MONTHLY;COUNT=3'), 'Every month, 3 times');
    expect(describeRule('FREQ=WEEKLY;BYDAY=FR;UNTIL=20201113T065959Z'), 'Every week on Fri, until 13 November 2020');
    expect(describeRule(''), 'Repeats');
  });

  test('repeat choices follow the start and keep a rule of their own', () {
    final wed = DateTime(2026, 11, 25, 10);
    final c = repeatChoices(wed, current: 'FREQ=DAILY;COUNT=5');
    expect(c['FREQ=WEEKLY;BYDAY=WE'], 'Every week on Wed');
    expect(c['FREQ=MONTHLY'], 'Every month on day 25');
    expect(c['FREQ=DAILY;COUNT=5'], 'Every day, 5 times');
    expect(repeatChoices(wed, current: 'FREQ=DAILY').length, 7);
  });

  test('a new rule never goes to one occurrence', () {
    expect(scopeFits(pb.RecurrenceScope.RECURRENCE_SCOPE_THIS, ruleChanged: true), isFalse);
    expect(scopeFits(pb.RecurrenceScope.RECURRENCE_SCOPE_THIS, ruleChanged: false), isTrue);
    expect(scopeFits(pb.RecurrenceScope.RECURRENCE_SCOPE_ALL, ruleChanged: true), isTrue);
  });
}
