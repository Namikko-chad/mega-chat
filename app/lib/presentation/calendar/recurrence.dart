import '../../core/format.dart';
import '../../core/i18n.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import 'calendar_screen.dart' show weekdayNames;

/// Repeat rules of events: RFC 5545 RRULE values without "RRULE:", as the
/// core carries them in `CalendarEvent.recurrence`.

const _days = ['MO', 'TU', 'WE', 'TH', 'FR', 'SA', 'SU'];

/// The rule's parts, keys upper-cased.
Map<String, String> ruleParts(String rule) => {
  for (final p in rule.split(';'))
    if (p.contains('=')) p.substring(0, p.indexOf('=')).trim().toUpperCase(): p.substring(p.indexOf('=') + 1).trim(),
};

/// The choices of the editor's "Repeat" field for an event starting at
/// [start]: rule → name. '' does not repeat; [current] (a rule the
/// presets do not cover) is kept as it is.
Map<String, String> repeatChoices(DateTime start, {String current = ''}) {
  final day = _days[start.weekday - 1];
  final out = <String, String>{
    '': tr('Does not repeat', 'Не повторяется'),
    'FREQ=DAILY': tr('Every day', 'Каждый день'),
    'FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR': tr('Every weekday (Mon–Fri)', 'По будням (пн–пт)'),
    'FREQ=WEEKLY;BYDAY=$day': tr('Every week on ', 'Каждую неделю: ') + weekdayNames()[start.weekday - 1],
    'FREQ=WEEKLY;INTERVAL=2;BYDAY=$day': tr('Every 2 weeks on ', 'Раз в 2 недели: ') + weekdayNames()[start.weekday - 1],
    'FREQ=MONTHLY': '${tr('Every month on day ', 'Каждый месяц, ')}${start.day}',
    'FREQ=YEARLY': tr('Every year', 'Каждый год'),
  };
  if (current.isNotEmpty && !out.containsKey(current)) out[current] = describeRule(current);
  return out;
}

/// The rule in words ("Every week on Mon, Wed, until 1 December").
String describeRule(String rule) {
  if (rule.isEmpty) return tr('Repeats', 'Повторяется');
  final p = ruleParts(rule);
  final n = int.tryParse(p['INTERVAL'] ?? '') ?? 1;
  final days = (p['BYDAY'] ?? '')
      .split(',')
      .map((d) => d.length >= 2 ? _days.indexOf(d.substring(d.length - 2).toUpperCase()) : -1)
      .where((i) => i >= 0)
      .toList();
  String text;
  switch (p['FREQ']) {
    case 'DAILY':
      text = n == 1 ? tr('Every day', 'Каждый день') : tr('Every $n days', 'Раз в $n дн.');
    case 'WEEKLY':
      final weekdays = days.length == 5 && days.every((d) => d < 5);
      text = weekdays && n == 1
          ? tr('Every weekday', 'По будням')
          : (n == 1 ? tr('Every week', 'Каждую неделю') : tr('Every $n weeks', 'Раз в $n нед.')) +
                (days.isEmpty ? '' : tr(' on ', ': ') + days.map((d) => weekdayNames()[d]).join(', '));
    case 'MONTHLY':
      text = n == 1 ? tr('Every month', 'Каждый месяц') : tr('Every $n months', 'Раз в $n мес.');
    case 'YEARLY':
      text = n == 1 ? tr('Every year', 'Каждый год') : tr('Every $n years', 'Раз в $n г.');
    default:
      text = tr('Repeats', 'Повторяется');
  }
  if (p['COUNT'] case final c?) text += tr(', $c times', ', $c раз');
  if (p['UNTIL'] case final u? when u.length >= 8) {
    final d = DateTime.tryParse(u.substring(0, 8));
    if (d != null) text += tr(', until ', ', до ') + formatDay(d);
  }
  return text;
}

/// Whether a scope may be offered: a single occurrence cannot take a new
/// rule.
bool scopeFits(pb.RecurrenceScope s, {required bool ruleChanged}) => !(ruleChanged && s == pb.RecurrenceScope.RECURRENCE_SCOPE_THIS);
