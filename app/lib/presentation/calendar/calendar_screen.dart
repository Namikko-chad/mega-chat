import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/format.dart';
import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/layout.dart';
import '../../core/theme/typography.dart';
import '../../data/rpc/gen/parley.pbgrpc.dart' as pb;
import '../accounts/account_rail.dart';
import '../accounts/login_wizard.dart';
import '../common/hm.dart';
import '../core/messenger_cubit.dart';
import '../mail/mail_screen.dart' show accountName;
import 'calendar_cubit.dart';
import 'event_editor.dart';

String monthName(int m) => tr(
  'January February March April May June July August September October November December',
  'Январь Февраль Март Апрель Май Июнь Июль Август Сентябрь Октябрь Ноябрь Декабрь',
).split(' ')[m - 1];

List<String> weekdayNames() => tr('Mon Tue Wed Thu Fri Sat Sun', 'Пн Вт Ср Чт Пт Сб Вс').split(' ');

String _two(int v) => v.toString().padLeft(2, '0');

String clock(DateTime t) => '${_two(t.hour)}:${_two(t.minute)}';

/// "10:00–11:30", "all day", or with dates when it spans days.
String eventTime(pb.CalendarEvent e, {DateTime? on}) {
  final (s, end) = eventSpan(e);
  if (e.allDay) {
    final last = end.subtract(const Duration(days: 1));
    return sameDay(s, last) ? tr('all day', 'весь день') : '${formatDay(s)} – ${formatDay(last)}';
  }
  if (sameDay(s, end) || (end.difference(s).inHours < 24 && end.hour == 0 && end.minute == 0)) {
    return '${clock(s)}–${clock(end)}';
  }
  if (on != null && !sameDay(on, s)) return '… ${clock(end)}';
  return '${formatDay(s)} ${clock(s)} – ${formatDay(end)} ${clock(end)}';
}

/// All calendars of every account as one month. Desktop: calendars,
/// month grid and the selected day side by side. Phone: the month over
/// the day's agenda.
class CalendarScreen extends StatelessWidget {
  const CalendarScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final m = context.watch<MessengerCubit>().state;
    if (calendarAccounts(m).isEmpty) {
      return Center(
        child: HmEmptyState(
          tr(
            'No calendar accounts yet. Add Google or Mail.ru / VK WorkSpace to see all your calendars here.',
            'Аккаунтов с календарём пока нет. Добавьте Google или Mail.ru / VK WorkSpace, чтобы видеть здесь все календари.',
          ),
          icon: Icons.calendar_month_outlined,
          action: HmButton(tr('Add account', 'Добавить аккаунт'), variant: HmVariant.primary, icon: Icons.add, onPressed: () => showLoginWizard(context)),
        ),
      );
    }
    if (Hm.of(context).mobile) {
      return const Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _MonthHeader(),
          _MonthGrid(compact: true),
          Divider(),
          Expanded(child: _DayAgenda()),
        ],
      );
    }
    return LayoutBuilder(
      builder: (context, box) {
        final wide = box.maxWidth >= 1280;
        return Row(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const SizedBox(width: 240, child: _CalendarList()),
            const Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  _MonthHeader(),
                  Expanded(child: _MonthGrid(compact: false)),
                ],
              ),
            ),
            Container(
              width: wide ? 340 : 280,
              decoration: const BoxDecoration(
                border: Border(left: BorderSide(color: HmColors.border)),
              ),
              child: const _DayAgenda(),
            ),
          ],
        );
      },
    );
  }
}

class _CalendarList extends StatelessWidget {
  const _CalendarList();

  @override
  Widget build(BuildContext context) {
    final s = context.watch<CalendarCubit>().state;
    final m = context.watch<MessengerCubit>().state;
    final cubit = context.read<CalendarCubit>();
    return AccountRail(
      top: Padding(
        padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
        child: HmButton(
          tr('New event', 'Новое событие'),
          variant: HmVariant.primary,
          icon: Icons.add,
          onPressed: () => showEventEditor(context, day: s.day),
        ),
      ),
      child: ReorderableAccounts(
        accounts: calendarAccounts(m),
        item: (a) {
          final keys = [for (final c in s.calendars.where((c) => c.accountId == a.id)) calKey(c)];
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // A tap shows all of the account's calendars, or hides them when all are shown.
              AccountTile(account: a, count: 0, onTap: keys.isEmpty ? null : () => cubit.setShown(keys, keys.any(s.hidden.contains))),
              for (final k in keys)
                _CalendarToggle(
                  name: s.calendar(k)!.name,
                  color: s.color(k),
                  on: !s.hidden.contains(k),
                  readOnly: s.calendar(k)!.readOnly,
                  onTap: () => cubit.toggle(k),
                ),
            ],
          );
        },
      ),
    );
  }
}

/// The accounts with calendars, in the user's order.
List<pb.Account> calendarAccounts(MessengerState m) => accountsOffering(m, pb.Capability.CAPABILITY_CALENDAR);

class _CalendarToggle extends StatelessWidget {
  const _CalendarToggle({required this.name, required this.color, required this.on, required this.readOnly, required this.onTap});

  final String name;
  final Color color;
  final bool on;
  final bool readOnly;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) => AccountSubItem(
    leading: AnimatedContainer(
      duration: const Duration(milliseconds: 150),
      width: 16,
      height: 16,
      decoration: BoxDecoration(
        color: on ? color : Colors.transparent,
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: color, width: 1.5),
      ),
      child: on ? const Icon(Icons.check, size: 12, color: HmColors.onAccent) : null,
    ),
    title: Text(
      name,
      overflow: TextOverflow.ellipsis,
      style: on ? null : HmText.body.copyWith(color: HmColors.muted),
    ),
    trailing: readOnly ? const Icon(Icons.lock_outline, size: 12, color: HmColors.muted) : null,
    onTap: onTap,
  );
}

class _MonthHeader extends StatelessWidget {
  const _MonthHeader();

  @override
  Widget build(BuildContext context) {
    final s = context.watch<CalendarCubit>().state;
    final m = context.watch<MessengerCubit>().state;
    final cubit = context.read<CalendarCubit>();
    final hm = Hm.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Padding(
          padding: EdgeInsets.fromLTRB(hm.mobile ? 16 : 20, hm.mobile ? 8 : 16, hm.mobile ? 8 : 16, 8),
          child: Row(
            children: [
              Expanded(
                child: Text('${monthName(s.month.month)} ${s.month.year}', style: HmText.modalTitle, maxLines: 1, overflow: TextOverflow.ellipsis),
              ),
              if (s.loading)
                const Padding(
                  padding: EdgeInsets.only(right: 8),
                  child: SizedBox(width: 14, height: 14, child: HmSpinner()),
                ),
              HmIconButton(icon: Icons.chevron_left, label: tr('Previous month', 'Предыдущий месяц'), onPressed: () => cubit.shiftMonth(-1)),
              if (hm.mobile)
                HmIconButton(icon: Icons.today_outlined, label: tr('Today', 'Сегодня'), onPressed: cubit.today)
              else
                HmButton(tr('Today', 'Сегодня'), small: true, variant: HmVariant.neutral, onPressed: cubit.today),
              HmIconButton(icon: Icons.chevron_right, label: tr('Next month', 'Следующий месяц'), onPressed: () => cubit.shiftMonth(1)),
              if (hm.mobile) ...[
                HmIconButton(icon: Icons.layers_outlined, label: tr('Calendars', 'Календари'), onPressed: () => _pickCalendars(context)),
                HmIconButton(
                  icon: Icons.add,
                  label: tr('New event', 'Новое событие'),
                  onPressed: () => showEventEditor(context, day: s.day),
                ),
              ],
            ],
          ),
        ),
        for (final f in s.failed)
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
            child: HmBanner('${accountName(m.account(f.accountId))}: ${f.error}', tone: HmTone.warning),
          ),
        if (s.error != null)
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
            child: HmBanner(s.error!, tone: HmTone.error),
          ),
      ],
    );
  }

  /// Phone: which calendars are shown, in a sheet.
  void _pickCalendars(BuildContext context) {
    final cubit = context.read<CalendarCubit>();
    final m = context.read<MessengerCubit>().state;
    showHmModal<void>(
      context,
      title: tr('Calendars', 'Календари'),
      body: (ctx) => BlocBuilder<CalendarCubit, CalendarState>(
        bloc: cubit,
        builder: (ctx, s) => Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            for (final c in s.calendars)
              _CalendarToggle(
                name: '${c.name} · ${accountName(m.account(c.accountId))}',
                color: s.color(calKey(c)),
                on: !s.hidden.contains(calKey(c)),
                readOnly: c.readOnly,
                onTap: () => cubit.toggle(calKey(c)),
              ),
          ],
        ),
      ),
    );
  }
}

class _MonthGrid extends StatelessWidget {
  const _MonthGrid({required this.compact});

  /// Phone: dots instead of event titles.
  final bool compact;

  @override
  Widget build(BuildContext context) {
    final s = context.watch<CalendarCubit>().state;
    final cubit = context.read<CalendarCubit>();
    final start = gridStart(s.month);
    final today = dayOf(DateTime.now());
    final days = [for (var i = 0; i < 42; i++) DateTime(start.year, start.month, start.day + i)];
    // Six weeks unless the month fits in five.
    final weeks = days[35].month != s.month.month ? 5 : 6;
    Widget cell(DateTime d) {
      final evs = s.on(d);
      final inMonth = d.month == s.month.month;
      final selected = d == s.day;
      final isToday = d == today;
      final number = Container(
        width: 24,
        height: 24,
        alignment: Alignment.center,
        decoration: BoxDecoration(color: isToday ? HmColors.accent : Colors.transparent, borderRadius: BorderRadius.circular(12)),
        child: Text(
          '${d.day}',
          style: HmText.jb(12, weight: isToday ? 700 : 400, color: isToday ? HmColors.onAccent : (inMonth ? HmColors.text : HmColors.muted)),
        ),
      );
      return MouseRegion(
        cursor: SystemMouseCursors.click,
        child: GestureDetector(
          behavior: HitTestBehavior.opaque,
          onTap: () => cubit.selectDay(d),
          onDoubleTap: compact ? null : () => showEventEditor(context, day: d),
          child: Container(
            decoration: BoxDecoration(
              color: selected ? HmColors.tint(HmColors.accent, 0.08) : null,
              border: Border(
                right: const BorderSide(color: HmColors.border),
                bottom: const BorderSide(color: HmColors.border),
                left: selected ? const BorderSide(color: HmColors.accent, width: 2) : BorderSide.none,
              ),
            ),
            padding: const EdgeInsets.all(4),
            child: compact
                ? Column(
                    children: [
                      number,
                      const SizedBox(height: 2),
                      Wrap(
                        spacing: 2,
                        runSpacing: 2,
                        alignment: WrapAlignment.center,
                        children: [
                          for (final e in evs.take(4))
                            Container(
                              width: 5,
                              height: 5,
                              decoration: BoxDecoration(color: s.color(eventCal(e)), shape: BoxShape.circle),
                            ),
                        ],
                      ),
                    ],
                  )
                : LayoutBuilder(
                    builder: (context, box) {
                      final fit = ((box.maxHeight - 28) / 18).floor().clamp(0, 20);
                      final more = evs.length > fit ? evs.length - (fit - 1).clamp(0, fit) : 0;
                      final shown = more > 0 ? evs.take((fit - 1).clamp(0, fit)) : evs;
                      return Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          Align(alignment: Alignment.topLeft, child: number),
                          const SizedBox(height: 2),
                          for (final e in shown) _EventChip(event: e, color: s.color(eventCal(e)), day: d),
                          if (more > 0) Text(tr('+$more more', 'ещё $more'), style: HmText.footer, overflow: TextOverflow.clip, maxLines: 1),
                        ],
                      );
                    },
                  ),
          ),
        ),
      );
    }

    final names = weekdayNames();
    final grid = Column(
      children: [
        Container(
          decoration: const BoxDecoration(
            border: Border(bottom: BorderSide(color: HmColors.border)),
          ),
          child: Row(
            children: [
              for (final n in names)
                Expanded(
                  child: Padding(
                    padding: const EdgeInsets.symmetric(vertical: 6),
                    child: Text(n.toUpperCase(), textAlign: TextAlign.center, style: HmText.th),
                  ),
                ),
            ],
          ),
        ),
        for (var w = 0; w < weeks; w++)
          (compact ? (Widget c) => SizedBox(height: 52, child: c) : (Widget c) => Expanded(child: c))(
            Row(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [for (var i = 0; i < 7; i++) Expanded(child: cell(days[w * 7 + i]))],
            ),
          ),
      ],
    );
    return Container(
      decoration: const BoxDecoration(
        border: Border(top: BorderSide(color: HmColors.border)),
      ),
      child: GestureDetector(
        // Swipe between months on touch screens.
        onHorizontalDragEnd: (d) {
          final v = d.primaryVelocity ?? 0;
          if (v.abs() > 300) cubit.shiftMonth(v < 0 ? 1 : -1);
        },
        child: grid,
      ),
    );
  }
}

class _EventChip extends StatelessWidget {
  const _EventChip({required this.event, required this.color, required this.day});

  final pb.CalendarEvent event;
  final Color color;
  final DateTime day;

  @override
  Widget build(BuildContext context) {
    final (s, _) = eventSpan(event);
    final label = event.allDay || !sameDay(s, day) ? event.title : '${clock(s)} ${event.title}';
    return Tooltip(
      message: '${event.title}\n${eventTime(event)}',
      waitDuration: const Duration(milliseconds: 600),
      child: GestureDetector(
        onTap: () => showEventDetails(context, event),
        child: Container(
          height: 16,
          margin: const EdgeInsets.only(bottom: 2),
          padding: const EdgeInsets.symmetric(horizontal: 4),
          decoration: BoxDecoration(
            color: HmColors.tint(color, event.allDay ? 0.30 : 0.12),
            borderRadius: BorderRadius.circular(3),
            border: Border(left: BorderSide(color: color, width: 2)),
          ),
          alignment: Alignment.centerLeft,
          child: Text(
            label,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: HmText.jb(10, color: HmColors.text),
          ),
        ),
      ),
    );
  }
}

class _DayAgenda extends StatelessWidget {
  const _DayAgenda();

  @override
  Widget build(BuildContext context) {
    final s = context.watch<CalendarCubit>().state;
    final m = context.watch<MessengerCubit>().state;
    final evs = s.on(s.day);
    final hm = Hm.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Padding(
          padding: EdgeInsets.fromLTRB(16, hm.mobile ? 12 : 16, 12, 8),
          child: Row(
            children: [
              Expanded(
                child: Text('${weekdayNames()[s.day.weekday - 1]}, ${formatDay(s.day)}', style: HmText.modalTitle, overflow: TextOverflow.ellipsis),
              ),
              if (!hm.mobile)
                HmIconButton(
                  icon: Icons.add,
                  label: tr('New event', 'Новое событие'),
                  onPressed: () => showEventEditor(context, day: s.day),
                ),
            ],
          ),
        ),
        Expanded(
          child: evs.isEmpty
              ? ListView(children: [HmEmptyState(tr('Nothing planned', 'Ничего не запланировано'), icon: Icons.event_available_outlined)])
              : ListView(
                  children: [
                    for (final e in evs)
                      HmListItem(
                        leading: Container(
                          width: 4,
                          height: 36,
                          decoration: BoxDecoration(color: s.color(eventCal(e)), borderRadius: BorderRadius.circular(2)),
                        ),
                        title: Text(e.title, overflow: TextOverflow.ellipsis),
                        meta: Text(eventTime(e, on: s.day)),
                        subtitle: Text(
                          [
                            if (e.location.isNotEmpty) e.location,
                            '${s.calendar(eventCal(e))?.name ?? ''} · ${accountName(m.account(e.accountId))}',
                          ].join(' · '),
                          overflow: TextOverflow.ellipsis,
                        ),
                        trailing: e.recurring ? const Icon(Icons.repeat, size: 12, color: HmColors.muted) : null,
                        onTap: () => showEventDetails(context, e),
                      ),
                  ],
                ),
        ),
      ],
    );
  }
}
