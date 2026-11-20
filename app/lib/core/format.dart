import 'package:fixnum/fixnum.dart';

import 'i18n.dart';

DateTime fromMs(Int64 ms) => DateTime.fromMillisecondsSinceEpoch(ms.toInt());

String _two(int v) => v.toString().padLeft(2, '0');

/// `YYYY-MM-DD HH:MM:SS` in local time.
String formatStamp(Int64 ms) {
  if (ms == Int64.ZERO) return '—';
  final t = fromMs(ms);
  return '${t.year}-${_two(t.month)}-${_two(t.day)} ${_two(t.hour)}:${_two(t.minute)}:${_two(t.second)}';
}

/// `HH:MM`.
String formatClock(Int64 ms) {
  final t = fromMs(ms);
  return '${_two(t.hour)}:${_two(t.minute)}';
}

bool sameDay(DateTime a, DateTime b) => a.year == b.year && a.month == b.month && a.day == b.day;

/// Chat list time: the clock today, the weekday this week, else the date.
String formatChatTime(Int64 ms) {
  if (ms == Int64.ZERO) return '';
  final t = fromMs(ms), now = DateTime.now();
  if (sameDay(t, now)) return formatClock(ms);
  final days = DateTime(now.year, now.month, now.day).difference(DateTime(t.year, t.month, t.day)).inDays;
  if (days < 7) return _weekdays()[t.weekday - 1];
  if (t.year == now.year) return '${_two(t.day)}.${_two(t.month)}';
  return '${_two(t.day)}.${_two(t.month)}.${_two(t.year % 100)}';
}

/// The day separator in a chat: Today, Yesterday or `12 March 2026`.
String formatDay(DateTime t) {
  final now = DateTime.now();
  if (sameDay(t, now)) return tr('Today', 'Сегодня');
  if (sameDay(t, now.subtract(const Duration(days: 1)))) return tr('Yesterday', 'Вчера');
  final month = tr(
    'January February March April May June July August September October November December',
    'января февраля марта апреля мая июня июля августа сентября октября ноября декабря',
  ).split(' ')[t.month - 1];
  return t.year == now.year ? '${t.day} $month' : '${t.day} $month ${t.year}';
}

List<String> _weekdays() => tr('Mon Tue Wed Thu Fri Sat Sun', 'Пн Вт Ср Чт Пт Сб Вс').split(' ');

String formatSize(num bytes) {
  final units = lang.value == Lang.ru ? const ['Б', 'КБ', 'МБ', 'ГБ', 'ТБ'] : const ['B', 'KB', 'MB', 'GB', 'TB'];
  var v = bytes.toDouble();
  var i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  return i == 0 ? '${v.toInt()} ${units[i]}' : '${v.toStringAsFixed(v < 10 ? 1 : 0)} ${units[i]}';
}

String platformName(String p) => switch (p) {
  'linux' => 'Linux',
  'windows' => 'Windows',
  'android' => 'Android',
  'darwin' => 'macOS',
  _ => p,
};
