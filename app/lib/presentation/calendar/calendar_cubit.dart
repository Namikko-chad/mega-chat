import 'dart:async';

import 'package:equatable/equatable.dart';
import 'package:fixnum/fixnum.dart';
import 'package:flutter/painting.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/theme/colors.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../core/core_cubit.dart';

/// Calendar ids are unique only inside an account.
typedef CalKey = ({String account, String calendar});

CalKey calKey(pb.Calendar c) => (account: c.accountId, calendar: c.id);

CalKey eventCal(pb.CalendarEvent e) => (account: e.accountId, calendar: e.calendarId);

bool hasCalendar(pb.Account a) => a.capabilities.contains(pb.Capability.CAPABILITY_CALENDAR);

DateTime dayOf(DateTime t) => DateTime(t.year, t.month, t.day);

/// Local start and end of an event. All-day events are dates: their
/// midnights UTC become local midnights of the same dates.
(DateTime, DateTime) eventSpan(pb.CalendarEvent e) {
  DateTime at(Int64 ms) {
    if (!e.allDay) return DateTime.fromMillisecondsSinceEpoch(ms.toInt());
    final u = DateTime.fromMillisecondsSinceEpoch(ms.toInt(), isUtc: true);
    return DateTime(u.year, u.month, u.day);
  }

  return (at(e.startMs), at(e.endMs));
}

/// The midnight UTC of a local date: how all-day events are sent.
Int64 dateMs(DateTime d) => Int64(DateTime.utc(d.year, d.month, d.day).millisecondsSinceEpoch);

bool onDay(pb.CalendarEvent e, DateTime day) {
  final (s, end) = eventSpan(e);
  final d0 = dayOf(day), d1 = DateTime(day.year, day.month, day.day + 1);
  return s.isBefore(d1) && (end.isAfter(d0) || (s == end && !s.isBefore(d0)));
}

/// The first day of the 6-week grid of a month (weeks start on Monday).
DateTime gridStart(DateTime month) {
  final first = DateTime(month.year, month.month);
  return DateTime(first.year, first.month, first.day - (first.weekday - 1));
}

const _palette = [HmColors.accent, HmColors.ok, HmColors.purpleLight, HmColors.warning, HmColors.active, HmColors.error];

/// Distinct colours for [calendars] in their order (see [CalendarState.color]).
/// Past the palette, hues go round the colour wheel by the golden angle.
Map<CalKey, Color> calendarColors(List<pb.Calendar> calendars) {
  final out = <CalKey, Color>{};
  final used = <int>{};
  var next = 0;
  Color fresh() {
    while (true) {
      final i = next++;
      final c = i < _palette.length ? _palette[i] : HSVColor.fromAHSV(1, ((i - _palette.length) * 137.508 + 20) % 360, 0.65, 0.95).toColor();
      if (!used.contains(c.toARGB32() & 0xFFFFFF)) return c;
    }
  }

  for (final cal in calendars) {
    final own = cal.color & 0xFFFFFF;
    final c = own != 0 && !used.contains(own) ? Color(0xFF000000 | own) : fresh();
    used.add(c.toARGB32() & 0xFFFFFF);
    out[calKey(cal)] = c;
  }
  return out;
}

class CalendarState extends Equatable {
  CalendarState({
    this.loading = false,
    this.error,
    this.calendars = const [],
    this.failed = const [],
    this.hidden = const {},
    DateTime? month,
    DateTime? day,
    this.events = const [],
    this.rev = 0,
  }) : month = month ?? DateTime(DateTime.now().year, DateTime.now().month),
       day = day ?? dayOf(DateTime.now());

  final bool loading;
  final String? error;
  final List<pb.Calendar> calendars;
  final List<pb.AccountError> failed;
  final Set<CalKey> hidden;

  /// The month shown (its first day) and the selected day.
  final DateTime month;
  final DateTime day;

  /// The events of the month's grid, every calendar.
  final List<pb.CalendarEvent> events;
  final int rev;

  late final List<pb.CalendarEvent> shown = events.where((e) => !hidden.contains(eventCal(e))).toList();

  List<pb.CalendarEvent> on(DateTime d) => shown.where((e) => onDay(e, d)).toList();

  pb.Calendar? calendar(CalKey k) => calendars.where((c) => calKey(c) == k).firstOrNull;

  /// One colour per calendar, the same for its mark and its events: the
  /// provider's own unless an earlier calendar already has it, else the next
  /// palette colour no calendar uses.
  late final Map<CalKey, Color> _colors = calendarColors(calendars);

  Color color(CalKey k) => _colors[k] ?? HmColors.muted;

  CalendarState copyWith({
    bool? loading,
    String? error,
    List<pb.Calendar>? calendars,
    List<pb.AccountError>? failed,
    Set<CalKey>? hidden,
    DateTime? month,
    DateTime? day,
    List<pb.CalendarEvent>? events,
    bool clearError = false,
  }) => CalendarState(
    loading: loading ?? this.loading,
    error: clearError ? null : error ?? this.error,
    calendars: calendars ?? this.calendars,
    failed: failed ?? this.failed,
    hidden: hidden ?? this.hidden,
    month: month ?? this.month,
    day: day ?? this.day,
    events: events ?? this.events,
    rev: rev + 1,
  );

  @override
  List<Object?> get props => [rev];
}

/// All calendars of every account as one: the month grid's events,
/// reloaded when a calendar changes.
class CalendarCubit extends Cubit<CalendarState> {
  CalendarCubit(this._core) : super(CalendarState()) {
    _coreSub = _core.stream.listen(_onCore);
    _eventSub = _core.events.listen(_onEvent);
    _onCore(_core.state);
  }

  final CoreCubit _core;
  late final StreamSubscription<CoreState> _coreSub;
  late final StreamSubscription<pb.Event> _eventSub;
  int? _session;
  Timer? _reload;
  int _gen = 0;
  Set<String> _online = {}; // calendar accounts that were online at the last load

  pb.MegaChatClient get api => _core.api;

  void _onCore(CoreState s) {
    if (!s.unlocked) {
      _session = null;
      if (state.calendars.isNotEmpty) emit(CalendarState());
      return;
    }
    if (s.conn != Conn.ready || s.session == _session) return;
    _session = s.session;
    unawaited(load());
  }

  void _onEvent(pb.Event ev) {
    if (!_core.state.unlocked) return;
    switch (ev.kind) {
      case pb.Event_Kind.KIND_CALENDAR:
        _schedule(calendars: false);
      case pb.Event_Kind.KIND_ACCOUNT:
        final a = ev.account;
        final online = a.status == pb.AccountStatus.ACCOUNT_STATUS_ONLINE && hasCalendar(a);
        if (a.provider.isEmpty || online != _online.contains(a.id)) {
          online ? _online.add(a.id) : _online.remove(a.id);
          _schedule(calendars: true);
        }
      default:
    }
  }

  bool _withCalendars = false;

  void _schedule({required bool calendars}) {
    _withCalendars |= calendars;
    _reload?.cancel();
    _reload = Timer(const Duration(milliseconds: 500), () {
      final all = _withCalendars;
      _withCalendars = false;
      unawaited(all ? load() : loadEvents());
    });
  }

  Future<void> load() async {
    emit(state.copyWith(loading: true, clearError: true));
    try {
      final r = await api.calendars(pb.AccountsRequest());
      if (isClosed) return;
      _online = {for (final c in r.calendars) c.accountId};
      emit(state.copyWith(calendars: r.calendars, failed: r.failed));
      await loadEvents();
    } catch (e) {
      if (!isClosed) emit(state.copyWith(loading: false, error: errorText(e)));
    }
  }

  Future<void> loadEvents() async {
    final gen = ++_gen;
    final from = gridStart(state.month);
    final to = DateTime(from.year, from.month, from.day + 42);
    try {
      final r = await api.events(pb.EventsRequest(fromMs: Int64(from.millisecondsSinceEpoch), toMs: Int64(to.millisecondsSinceEpoch)));
      if (isClosed || gen != _gen) return;
      emit(state.copyWith(loading: false, events: r.events, failed: r.failed, clearError: true));
    } catch (e) {
      if (!isClosed && gen == _gen) emit(state.copyWith(loading: false, error: errorText(e)));
    }
  }

  void showMonth(DateTime month) {
    final m = DateTime(month.year, month.month);
    final keepDay = state.day.year == m.year && state.day.month == m.month;
    emit(state.copyWith(month: m, day: keepDay ? state.day : m, events: const []));
    unawaited(loadEvents());
  }

  void shiftMonth(int by) => showMonth(DateTime(state.month.year, state.month.month + by));

  void today() {
    final now = DateTime.now();
    showMonth(now);
    emit(state.copyWith(day: dayOf(now)));
  }

  void selectDay(DateTime d) {
    if (d.year != state.month.year || d.month != state.month.month) {
      showMonth(d);
    }
    emit(state.copyWith(day: dayOf(d)));
  }

  void toggle(CalKey k) {
    final hidden = Set.of(state.hidden);
    hidden.contains(k) ? hidden.remove(k) : hidden.add(k);
    emit(state.copyWith(hidden: hidden));
  }

  /// Shows or hides several calendars at once (all of an account's).
  void setShown(Iterable<CalKey> keys, bool on) {
    final hidden = Set.of(state.hidden);
    on ? hidden.removeAll(keys) : hidden.addAll(keys);
    emit(state.copyWith(hidden: hidden));
  }

  Future<pb.CalendarEvent> save(pb.CalendarEvent e) async {
    final out = await api.saveEvent(e);
    await loadEvents();
    return out;
  }

  /// Deletes an event; of a repeating one, what [scope] says.
  Future<void> delete(pb.CalendarEvent e, {pb.RecurrenceScope scope = pb.RecurrenceScope.RECURRENCE_SCOPE_THIS}) async {
    await api.deleteEvent(pb.DeleteEventRequest(accountId: e.accountId, calendarId: e.calendarId, id: e.id, scope: scope));
    emit(state.copyWith(events: state.events.where((x) => !(x.accountId == e.accountId && x.id == e.id)).toList()));
    await loadEvents();
  }

  @override
  Future<void> close() async {
    await _coreSub.cancel();
    await _eventSub.cancel();
    _reload?.cancel();
    return super.close();
  }
}
