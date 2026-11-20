import 'package:fixnum/fixnum.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/format.dart';
import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/layout.dart';
import '../../core/theme/typography.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../common/hm.dart';
import '../core/messenger_cubit.dart';
import '../mail/compose.dart' show formatAddresses, parseAddresses;
import '../mail/mail_screen.dart' show accountName, addressName;
import 'calendar_cubit.dart';
import 'calendar_screen.dart' show eventTime;

String rsvpName(pb.Rsvp r) => switch (r) {
  pb.Rsvp.RSVP_ACCEPTED => tr('accepted', 'придёт'),
  pb.Rsvp.RSVP_DECLINED => tr('declined', 'отказ'),
  pb.Rsvp.RSVP_TENTATIVE => tr('maybe', 'возможно'),
  pb.Rsvp.RSVP_NEEDS_ACTION => tr('no answer', 'нет ответа'),
  _ => '',
};

Widget _field(String label, String value) => HmField(label, SelectableText(value));

/// An event with what is known about it; edit and delete when its
/// calendar may be written.
Future<void> showEventDetails(BuildContext context, pb.CalendarEvent e) {
  final cubit = context.read<CalendarCubit>();
  final m = context.read<MessengerCubit>().state;
  final s = cubit.state;
  final cal = s.calendar(eventCal(e));
  final writable = cal != null && !cal.readOnly;
  return showHmModal<void>(
    context,
    title: e.title,
    body: (ctx) => Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          children: [
            Container(
              width: 10,
              height: 10,
              decoration: BoxDecoration(color: s.color(eventCal(e)), borderRadius: BorderRadius.circular(3)),
            ),
            const SizedBox(width: 8),
            Expanded(child: Text('${cal?.name ?? ''} · ${accountName(m.account(e.accountId))}', style: HmText.small)),
          ],
        ),
        const SizedBox(height: 12),
        _field(tr('When', 'Когда'), '${formatDay(eventSpan(e).$1)}, ${eventTime(e)}${e.recurring ? tr(' · repeats', ' · повторяется') : ''}'),
        if (e.location.isNotEmpty) _field(tr('Where', 'Где'), e.location),
        if (e.link.isNotEmpty) _field(tr('Link', 'Ссылка'), e.link),
        if (e.hasOrganizer() && e.organizer.email.isNotEmpty) _field(tr('Organizer', 'Организатор'), addressName(e.organizer)),
        if (e.attendees.isNotEmpty)
          _field(
            tr('Guests', 'Участники'),
            e.attendees.map((a) => '${addressName(a.address)}${rsvpName(a.status).isEmpty ? '' : ' (${rsvpName(a.status)})'}').join('\n'),
          ),
        if (e.description.isNotEmpty) ...[const SizedBox(height: 8), SelectableText(e.description, style: HmText.body.copyWith(height: 1.5))],
        if (writable) ...[
          const SizedBox(height: 20),
          Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              HmButton(
                tr('Delete', 'Удалить'),
                variant: HmVariant.red,
                icon: Icons.delete_outline,
                onPressed: () async {
                  final ok = await confirm(
                    ctx,
                    title: tr('Delete the event?', 'Удалить событие?'),
                    message: e.recurring
                        ? tr('Only this occurrence is deleted.', 'Будет удалено только это повторение.')
                        : tr('It is deleted from the calendar for everyone.', 'Оно будет удалено из календаря у всех.'),
                    action: tr('Delete', 'Удалить'),
                    destructive: true,
                  );
                  if (!ok) return;
                  try {
                    await cubit.delete(e);
                    if (ctx.mounted) Navigator.pop(ctx);
                  } catch (err) {
                    Toasts.error(tr('Could not delete', 'Не удалось удалить'), errorText(err));
                  }
                },
              ),
              const SizedBox(width: 8),
              HmButton(
                tr('Edit', 'Изменить'),
                variant: HmVariant.primary,
                icon: Icons.edit_outlined,
                onPressed: () {
                  Navigator.pop(ctx);
                  showEventEditor(context, event: e);
                },
              ),
            ],
          ),
        ],
      ],
    ),
  );
}

/// Creates an event (on [day]) or edits [event].
Future<void> showEventEditor(BuildContext context, {pb.CalendarEvent? event, DateTime? day}) {
  final cubit = context.read<CalendarCubit>();
  final writable = cubit.state.calendars.where((c) => !c.readOnly).toList();
  if (writable.isEmpty) {
    Toasts.error(tr('No calendar to write to', 'Нет календаря для записи'));
    return Future.value();
  }
  return showHmModal<void>(
    context,
    title: event == null ? tr('New event', 'Новое событие') : tr('Edit event', 'Изменить событие'),
    body: (ctx) => BlocProvider.value(
      value: cubit,
      child: _Editor(calendars: writable, event: event, day: day ?? dayOf(DateTime.now())),
    ),
  );
}

class _Editor extends StatefulWidget {
  const _Editor({required this.calendars, required this.event, required this.day});

  final List<pb.Calendar> calendars;
  final pb.CalendarEvent? event;
  final DateTime day;

  @override
  State<_Editor> createState() => _EditorState();
}

class _EditorState extends State<_Editor> {
  final _title = TextEditingController();
  final _location = TextEditingController();
  final _description = TextEditingController();
  final _guests = TextEditingController();
  late CalKey _cal;
  late bool _allDay;
  late DateTime _start, _end; // all day: dates, end inclusive here
  String? _error, _titleError, _guestsError;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    final e = widget.event;
    if (e != null) {
      _title.text = e.title;
      _location.text = e.location;
      _description.text = e.description;
      _guests.text = formatAddresses(e.attendees.where((a) => !a.self).map((a) => a.address));
      _cal = eventCal(e);
      _allDay = e.allDay;
      final (s, end) = eventSpan(e);
      _start = s;
      _end = e.allDay ? end.subtract(const Duration(days: 1)) : end;
    } else {
      final primary = widget.calendars.where((c) => c.primary).firstOrNull ?? widget.calendars.first;
      _cal = calKey(primary);
      _allDay = false;
      final now = DateTime.now();
      final hour = sameDay(widget.day, now) ? now.hour + 1 : 10;
      _start = DateTime(widget.day.year, widget.day.month, widget.day.day, hour);
      _end = _start.add(const Duration(hours: 1));
    }
  }

  @override
  void dispose() {
    for (final c in [_title, _location, _description, _guests]) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _pickDate(bool start) async {
    final cur = start ? _start : _end;
    final d = await showDatePicker(context: context, initialDate: cur, firstDate: DateTime(2000), lastDate: DateTime(2100), locale: currentLocale);
    if (d == null) return;
    setState(() {
      final moved = DateTime(d.year, d.month, d.day, cur.hour, cur.minute);
      if (start) {
        final len = _end.difference(_start);
        _start = moved;
        _end = moved.add(len);
      } else {
        _end = moved;
      }
    });
  }

  Future<void> _pickTime(bool start) async {
    final cur = start ? _start : _end;
    final t = await showTimePicker(context: context, initialTime: TimeOfDay.fromDateTime(cur));
    if (t == null) return;
    setState(() {
      final moved = DateTime(cur.year, cur.month, cur.day, t.hour, t.minute);
      if (start) {
        final len = _end.difference(_start);
        _start = moved;
        _end = moved.add(len);
      } else {
        _end = moved;
      }
    });
  }

  Future<void> _save() async {
    final guests = parseAddresses(_guests.text);
    setState(() {
      _titleError = _title.text.trim().isEmpty ? tr('Enter a title', 'Введите название') : null;
      _guestsError = guests == null ? tr('Not an address', 'Это не адрес') : null;
      _error = null;
      if (_allDay ? _end.isBefore(dayOf(_start)) : !_end.isAfter(_start)) _error = tr('It must end after it starts', 'Конец должен быть позже начала');
    });
    if (_titleError != null || _guestsError != null || _error != null) return;
    final e = pb.CalendarEvent(
      accountId: _cal.account,
      calendarId: _cal.calendar,
      id: widget.event != null && eventCal(widget.event!) == _cal ? widget.event!.id : '',
      title: _title.text.trim(),
      location: _location.text.trim(),
      description: _description.text.trim(),
      allDay: _allDay,
      startMs: _allDay ? dateMs(_start) : Int64(_start.millisecondsSinceEpoch),
      endMs: _allDay ? dateMs(DateTime(_end.year, _end.month, _end.day + 1)) : Int64(_end.millisecondsSinceEpoch),
      attendees: [for (final g in guests!) pb.Attendee(address: g)],
    );
    setState(() => _saving = true);
    final cubit = context.read<CalendarCubit>();
    try {
      final moved = widget.event != null && eventCal(widget.event!) != _cal;
      await cubit.save(e);
      if (moved) await cubit.delete(widget.event!); // to another calendar: a new one, the old goes
      if (!mounted) return;
      Toasts.success(tr('Saved', 'Сохранено'));
      Navigator.pop(context);
    } catch (err) {
      if (mounted) setState(() => _error = errorText(err));
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    final m = context.read<MessengerCubit>().state;
    String two(int v) => v.toString().padLeft(2, '0');
    Widget when(String label, DateTime t, bool start) => HmFormField(
      label: label,
      child: Row(
        children: [
          Expanded(
            child: HmButton(formatDay(t), variant: HmVariant.neutral, icon: Icons.event_outlined, onPressed: () => _pickDate(start)),
          ),
          if (!_allDay) ...[
            const SizedBox(width: 8),
            HmButton('${two(t.hour)}:${two(t.minute)}', variant: HmVariant.neutral, icon: Icons.schedule, onPressed: () => _pickTime(start)),
          ],
        ],
      ),
    );
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        HmTextField(label: tr('Title', 'Название'), controller: _title, required: true, error: _titleError, autofocus: widget.event == null),
        SizedBox(height: hm.gapSm),
        HmSelect<CalKey>(
          label: tr('Calendar', 'Календарь'),
          options: {for (final c in widget.calendars) calKey(c): '${c.name} · ${accountName(m.account(c.accountId))}'},
          value: _cal,
          onChanged: (v) => setState(() => _cal = v),
        ),
        SizedBox(height: hm.gapSm),
        HmToggleRow(label: tr('All day', 'Весь день'), value: _allDay, onChanged: (v) => setState(() => _allDay = v)),
        SizedBox(height: hm.gapSm),
        when(tr('Starts', 'Начало'), _start, true),
        SizedBox(height: hm.gapSm),
        when(tr('Ends', 'Конец'), _end, false),
        SizedBox(height: hm.gapSm),
        HmTextField(label: tr('Where', 'Где'), controller: _location),
        SizedBox(height: hm.gapSm),
        HmTextField(
          label: tr('Guests', 'Участники'),
          controller: _guests,
          error: _guestsError,
          identifier: true,
          keyboard: TextInputType.emailAddress,
          hint: tr('Addresses, separated by commas; they get an invitation', 'Адреса через запятую; им придёт приглашение'),
        ),
        SizedBox(height: hm.gapSm),
        HmTextField(label: tr('Description', 'Описание'), controller: _description, lines: 3, keyboard: TextInputType.multiline),
        if (_error != null) ...[SizedBox(height: hm.gapSm), HmBanner(_error!, tone: HmTone.error)],
        SizedBox(height: hm.gapMd),
        Row(
          mainAxisAlignment: MainAxisAlignment.end,
          children: [
            HmButton(tr('Cancel', 'Отмена'), variant: HmVariant.neutral, onPressed: () => Navigator.pop(context)),
            const SizedBox(width: 8),
            HmButton(tr('Save', 'Сохранить'), variant: HmVariant.primary, loading: _saving, onPressed: _saving ? null : _save),
          ],
        ),
        if (widget.event?.recurring ?? false) ...[
          const SizedBox(height: 8),
          Text(
            tr('Changes apply to this occurrence only.', 'Изменения коснутся только этого повторения.'),
            style: HmText.small.copyWith(color: HmColors.muted),
          ),
        ],
      ],
    );
  }
}
