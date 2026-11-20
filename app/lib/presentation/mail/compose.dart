import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/format.dart';
import '../../core/i18n.dart';
import '../../core/theme/layout.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../common/hm.dart';
import '../core/messenger_cubit.dart';
import 'mail_cubit.dart';
import 'mail_screen.dart';

/// Parses `Name <a@b>, c@d; …`; null when a part is not an address.
List<pb.Address>? parseAddresses(String s) {
  final out = <pb.Address>[];
  for (var part in s.split(RegExp('[,;]'))) {
    part = part.trim();
    if (part.isEmpty) continue;
    final m = RegExp(r'^(.*)<([^>]+)>$').firstMatch(part);
    final name = m == null ? '' : m.group(1)!.trim().replaceAll('"', '');
    final email = (m == null ? part : m.group(2)!).trim();
    if (!RegExp(r'^[^@\s]+@[^@\s]+$').hasMatch(email)) return null;
    out.add(pb.Address(name: name, email: email));
  }
  return out;
}

String formatAddresses(Iterable<pb.Address> l) => l.map((a) => a.name.isEmpty ? a.email : '${a.name} <${a.email}>').join(', ');

/// Opens the composer: a new message, a reply (to all) or a forward.
Future<void> showCompose(BuildContext context, {pb.Mail? reply, bool all = false, pb.Mail? forward}) {
  final mail = context.read<MailCubit>();
  final m = context.read<MessengerCubit>().state;
  return showHmModal<void>(
    context,
    title: reply != null
        ? tr('Reply', 'Ответ')
        : forward != null
        ? tr('Forward', 'Пересылка')
        : tr('New message', 'Новое письмо'),
    size: HmModalSize.lg,
    body: (ctx) => BlocProvider.value(
      value: mail,
      child: _Compose(accounts: m.accounts.where(hasMail).toList(), view: mail.state.view, reply: reply, all: all, forward: forward),
    ),
  );
}

class _Compose extends StatefulWidget {
  const _Compose({required this.accounts, required this.view, this.reply, this.all = false, this.forward});

  final List<pb.Account> accounts;
  final MailView view;
  final pb.Mail? reply;
  final bool all;
  final pb.Mail? forward;

  @override
  State<_Compose> createState() => _ComposeState();
}

class _ComposeState extends State<_Compose> {
  late String _account;
  final _to = TextEditingController();
  final _cc = TextEditingController();
  final _subject = TextEditingController();
  final _text = TextEditingController();
  String? _toError, _ccError, _error;
  bool _sending = false;

  @override
  void initState() {
    super.initState();
    final src = widget.reply ?? widget.forward;
    _account = src?.accountId ?? (widget.view.account.isNotEmpty ? widget.view.account : widget.accounts.first.id);
    final acct = widget.accounts.where((a) => a.id == _account).firstOrNull;
    // The account's own addresses: its login and the address it signed in as.
    final self = {
      for (final v in [acct?.login, acct?.self.username])
        if (v != null && v.contains('@')) v.toLowerCase(),
    };
    bool notSelf(pb.Address a) => !self.contains(a.email.toLowerCase());
    final r = widget.reply;
    if (r != null) {
      final to = r.replyTo.isNotEmpty ? r.replyTo : [r.from];
      _to.text = formatAddresses(to);
      if (widget.all) {
        final seen = {for (final a in to) a.email.toLowerCase()};
        _cc.text = formatAddresses([...r.to, ...r.cc].where(notSelf).where((a) => seen.add(a.email.toLowerCase())));
      }
      _subject.text = r.subject.toLowerCase().startsWith('re:') ? r.subject : 'Re: ${r.subject}';
      final quoted = r.text.split('\n').map((l) => '> $l').join('\n');
      _text.text = '\n\n${formatStamp(r.timeMs)}, ${addressName(r.from)} ${tr('wrote', 'пишет')}:\n$quoted';
    }
    final f = widget.forward;
    if (f != null) {
      _subject.text = f.subject.toLowerCase().startsWith('fwd:') ? f.subject : 'Fwd: ${f.subject}';
      _text.text =
          '\n\n---------- ${tr('Forwarded message', 'Пересланное письмо')} ----------\n'
          '${tr('From', 'От')}: ${formatAddresses([f.from])}\n'
          '${tr('Date', 'Дата')}: ${formatStamp(f.timeMs)}\n'
          '${tr('Subject', 'Тема')}: ${f.subject}\n'
          '${tr('To', 'Кому')}: ${formatAddresses(f.to)}\n\n${f.text}';
    }
  }

  @override
  void dispose() {
    for (final c in [_to, _cc, _subject, _text]) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _send() async {
    final to = parseAddresses(_to.text), cc = parseAddresses(_cc.text);
    setState(() {
      _toError = to == null ? tr('Not an address', 'Это не адрес') : (to.isEmpty ? tr('Who is it to?', 'Кому отправить?') : null);
      _ccError = cc == null ? tr('Not an address', 'Это не адрес') : null;
      _error = null;
    });
    if (_toError != null || _ccError != null) return;
    setState(() => _sending = true);
    try {
      await context.read<MailCubit>().send(
        pb.SendMailRequest(accountId: _account, to: to, cc: cc, subject: _subject.text.trim(), text: _text.text, replyToId: widget.reply?.id ?? ''),
      );
      if (!mounted) return;
      Toasts.success(tr('Sent', 'Отправлено'));
      Navigator.pop(context);
    } catch (e) {
      if (mounted) setState(() => _error = errorText(e));
    } finally {
      if (mounted) setState(() => _sending = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (widget.accounts.length > 1 && widget.reply == null && widget.forward == null) ...[
          HmSelect<String>(
            label: tr('From', 'От кого'),
            options: {for (final a in widget.accounts) a.id: '${accountName(a)}${a.login.isNotEmpty && a.login != a.name ? ' <${a.login}>' : ''}'},
            value: _account,
            onChanged: (v) => setState(() => _account = v),
          ),
          SizedBox(height: hm.gapSm),
        ],
        HmTextField(
          label: tr('To', 'Кому'),
          controller: _to,
          required: true,
          error: _toError,
          identifier: true,
          keyboard: TextInputType.emailAddress,
          autofocus: widget.reply == null,
        ),
        SizedBox(height: hm.gapSm),
        HmTextField(label: tr('Cc', 'Копия'), controller: _cc, error: _ccError, identifier: true, keyboard: TextInputType.emailAddress),
        SizedBox(height: hm.gapSm),
        HmTextField(label: tr('Subject', 'Тема'), controller: _subject),
        SizedBox(height: hm.gapSm),
        HmTextField(label: tr('Message', 'Письмо'), controller: _text, lines: 10, keyboard: TextInputType.multiline, autofocus: widget.reply != null),
        if (_error != null) ...[SizedBox(height: hm.gapSm), HmBanner(_error!, tone: HmTone.error)],
        SizedBox(height: hm.gapMd),
        Row(
          mainAxisAlignment: MainAxisAlignment.end,
          children: [
            HmButton(tr('Cancel', 'Отмена'), variant: HmVariant.neutral, onPressed: () => Navigator.pop(context)),
            const SizedBox(width: 8),
            HmButton(tr('Send', 'Отправить'), variant: HmVariant.primary, icon: Icons.send_outlined, loading: _sending, onPressed: _sending ? null : _send),
          ],
        ),
      ],
    );
  }
}
