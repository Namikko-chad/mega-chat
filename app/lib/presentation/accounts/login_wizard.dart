import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/layout.dart';
import '../../core/theme/typography.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../common/hm.dart';
import '../core/messenger_cubit.dart';
import 'providers.dart';

/// Adds an account, or signs [relogin] in again: provider, auth method,
/// then the steps the core asks for. Returns the account when done.
Future<pb.Account?> showLoginWizard(BuildContext context, {pb.Account? relogin}) => showHmModal<pb.Account>(
  context,
  title: relogin == null ? tr('Add account', 'Добавить аккаунт') : tr('Sign in again', 'Войти снова'),
  dismissible: true,
  body: (_) => _Wizard(relogin: relogin),
);

class _Wizard extends StatefulWidget {
  const _Wizard({this.relogin});

  final pb.Account? relogin;

  @override
  State<_Wizard> createState() => _WizardState();
}

class _WizardState extends State<_Wizard> {
  // Read once: dispose cancels the sign-in after the modal is gone.
  late final MessengerCubit _messenger = context.read<MessengerCubit>();

  List<pb.ProviderInfo> _providers = [];
  pb.ProviderInfo? _provider; // chosen, waiting for a method
  pb.LoginStepReply? _step;
  String _loginId = '';
  bool _busy = false;
  String? _error; // the call failed (not the core's step error)
  final _fields = <String, TextEditingController>{};
  final _fieldErrors = <String, String>{};

  // Waiting for a browser sign-in or a QR scan: LoginStep with no values
  // until the step changes (the core answers early on news). [_gen] bumps on every answer of the user, so a poll that
  // was in flight meanwhile does not overwrite its result.
  bool _polling = false;
  int _gen = 0;
  String _opened = '';

  pb.MegaChatClient get _api => _messenger.api;

  bool get _done => _step?.done ?? false;

  /// The step finishes outside the app (browser, phone scanning a QR).
  static bool _waits(pb.LoginStepReply? s) => s != null && (s.openUrl.isNotEmpty || s.qr.isNotEmpty);

  @override
  void initState() {
    super.initState();
    final re = widget.relogin;
    if (re != null) {
      _call(() => _api.relogin(pb.IdRequest(id: re.id)));
    } else {
      _providers = _messenger.state.providers;
      if (_providers.isEmpty) _loadProviders();
    }
  }

  @override
  void dispose() {
    if (_loginId.isNotEmpty && !_done) unawaited(_api.cancelLogin(pb.IdRequest(id: _loginId)).then((_) {}, onError: (_) {}));
    for (final c in _fields.values) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _loadProviders() async {
    setState(() => _busy = true);
    try {
      final r = await _api.providers(pb.Empty());
      if (mounted) setState(() => _providers = r.providers);
    } catch (e) {
      if (mounted) setState(() => _error = errorText(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  void _pickProvider(pb.ProviderInfo p) {
    if (p.authMethods.length > 1) {
      setState(() => _provider = p);
    } else {
      _start(p, p.authMethods.firstOrNull?.id ?? '');
    }
  }

  void _start(pb.ProviderInfo p, String method) =>
      _call(() => _api.startLogin(pb.StartLoginRequest(provider: p.type, method: method)));

  Future<void> _call(Future<pb.LoginStepReply> Function() rpc) async {
    final gen = ++_gen;
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final r = await rpc();
      if (mounted && gen == _gen) _apply(r);
    } catch (e) {
      if (mounted && gen == _gen) setState(() => _error = errorText(e));
    } finally {
      if (mounted && gen == _gen) setState(() => _busy = false);
    }
  }

  void _apply(pb.LoginStepReply r) {
    final retry = r.error.isNotEmpty;
    final old = Map.of(_fields);
    _fields.clear();
    for (final f in r.fields) {
      // A rejected answer keeps what was typed, except secrets.
      final keep = retry && !f.secret ? old[f.key]?.text : null;
      _fields[f.key] = TextEditingController(text: keep ?? f.defaultValue);
    }
    for (final c in old.values) {
      c.dispose();
    }
    _fieldErrors.clear();
    setState(() {
      _step = r;
      if (r.loginId.isNotEmpty) _loginId = r.loginId;
    });
    if (r.done && r.hasAccount()) _messenger.upsertAccount(r.account);
    if (!r.done && r.openUrl.isNotEmpty && r.openUrl != _opened) {
      _opened = r.openUrl;
      _open();
    }
    if (!r.done && _waits(r)) _poll();
  }

  Future<void> _open() async {
    final ok = await launchUrl(Uri.parse(_opened), mode: LaunchMode.externalApplication).catchError((_) => false);
    if (!ok) Toasts.error(tr('Could not open the browser', 'Не удалось открыть браузер'), tr('copy the link instead', 'скопируйте ссылку'));
  }

  Future<void> _poll() async {
    if (_polling) return;
    _polling = true;
    try {
      while (mounted && !_done && _waits(_step) && _error == null) {
        // The core holds each call until something changes, so a short
        // pause between calls is enough.
        await Future<void>.delayed(const Duration(milliseconds: 300));
        if (!mounted || _busy) continue;
        final gen = _gen;
        try {
          final r = await _api.loginStep(pb.LoginStepRequest(loginId: _loginId));
          if (!mounted || _done) return;
          if (gen == _gen || r.done) _apply(r);
        } catch (e) {
          if (mounted && gen == _gen) setState(() => _error = errorText(e));
        }
      }
    } finally {
      _polling = false;
    }
  }

  void _submit() {
    final step = _step;
    if (step == null) return;
    _fieldErrors.clear();
    for (final f in step.fields) {
      if (!f.optional && _fields[f.key]!.text.trim().isEmpty) _fieldErrors[f.key] = tr('Required', 'Обязательное поле');
    }
    setState(() {});
    if (_fieldErrors.isNotEmpty) return;
    final values = {for (final f in step.fields) f.key: f.secret ? _fields[f.key]!.text : _fields[f.key]!.text.trim()};
    _call(() => _api.loginStep(pb.LoginStepRequest(loginId: _loginId, values: values.entries)));
  }

  @override
  Widget build(BuildContext context) {
    final step = _step;
    final Widget body;
    if (step != null && step.done) {
      body = _doneView(step.account);
    } else if (step != null) {
      body = _stepView(step);
    } else if (_provider != null) {
      body = _methods(_provider!);
    } else if (widget.relogin == null) {
      body = _providerList();
    } else {
      body = const SizedBox(height: 120, child: HmLoading());
    }
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (_error != null) ...[
          HmBanner(
            _error!,
            tone: HmTone.error,
            action: _waits(step)
                ? HmButton(
                    tr('Check again', 'Проверить снова'),
                    small: true,
                    onPressed: () {
                      setState(() => _error = null);
                      _poll();
                    },
                  )
                : null,
          ),
          const SizedBox(height: 16),
        ],
        body,
      ],
    );
  }

  Widget _providerList() {
    if (_busy && _providers.isEmpty) return const SizedBox(height: 120, child: HmLoading());
    if (_providers.isEmpty) return HmEmptyState(tr('The core offers no providers', 'Ядро не предлагает провайдеров'));
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text(tr('Where is the account?', 'Где этот аккаунт?'), style: HmText.small),
        const SizedBox(height: 12),
        for (final p in _providers)
          HmListItem(
            leading: HmAvatar(label: p.name, icon: providerIcon(p.type), color: providerColor(p.type)),
            title: Text(providerName(p.type, p)),
            subtitle: Text(p.authMethods.map((m) => m.name).join(' · ')),
            onTap: _busy ? null : () => _pickProvider(p),
          ),
      ],
    );
  }

  Widget _methods(pb.ProviderInfo p) => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    children: [
      Text(tr('How to sign in to ${providerName(p.type, p)}?', 'Как войти в ${providerName(p.type, p)}?'), style: HmText.small),
      const SizedBox(height: 12),
      for (final m in p.authMethods)
        HmListItem(
          leading: HmAvatar(label: m.name, icon: providerIcon(p.type), color: providerColor(p.type)),
          title: Text(m.name),
          subtitle: m.description.isEmpty ? null : Text(m.description),
          onTap: _busy ? null : () => _start(p, m.id),
        ),
      const SizedBox(height: 16),
      _Actions(
        children: [
          HmButton(tr('Back', 'Назад'), variant: HmVariant.neutral, onPressed: () => setState(() => _provider = null)),
        ],
      ),
    ],
  );

  Widget _stepView(pb.LoginStepReply step) {
    final hm = Hm.of(context);
    final gap = SizedBox(height: hm.gapSm);
    final waiting = _waits(step);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (step.title.isNotEmpty) ...[HmSectionTitle(step.title), gap],
        if (step.message.isNotEmpty) ...[SelectableText(step.message, style: HmText.small.copyWith(height: 1.5)), gap],
        if (step.error.isNotEmpty) ...[HmBanner(step.error, tone: HmTone.error), gap],
        if (step.qr.isNotEmpty) ...[_qrCard(step.qr), gap],
        if (step.openUrl.isNotEmpty) ...[
          HmCard(
            recessed: true,
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Row(
                  children: [
                    const SizedBox.square(dimension: 16, child: HmSpinner()),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        tr('Waiting for the sign-in in the browser…', 'Ждём входа в браузере…'),
                        style: HmText.bodyStrong,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                SelectableText(step.openUrl, style: HmText.small, maxLines: 3),
                const SizedBox(height: 12),
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: [
                    HmButton(tr('Open again', 'Открыть снова'), small: true, icon: Icons.open_in_new, onPressed: _open),
                    HmButton(
                      tr('Copy link', 'Копировать ссылку'),
                      small: true,
                      icon: Icons.copy_outlined,
                      onPressed: () {
                        Clipboard.setData(ClipboardData(text: step.openUrl));
                        Toasts.success(tr('Copied', 'Скопировано'));
                      },
                    ),
                  ],
                ),
              ],
            ),
          ),
          gap,
        ],
        for (final f in step.fields) ...[
          HmTextField(
            label: f.label.isEmpty ? f.key : f.label,
            controller: _fields[f.key],
            required: !f.optional,
            password: f.secret,
            lines: f.multiline && !f.secret ? 3 : 1,
            hint: f.help.isEmpty ? null : f.help,
            error: _fieldErrors[f.key],
            identifier: true,
            autofocus: f == step.fields.first,
            onSubmitted: f.multiline ? null : (_) => _submit(),
          ),
          gap,
        ],
        const SizedBox(height: 8),
        _Actions(
          children: [
            HmButton(tr('Cancel', 'Отмена'), variant: HmVariant.neutral, expand: hm.mobile, onPressed: () => Navigator.of(context).pop()),
            if (step.fields.isNotEmpty || !waiting)
              HmButton(tr('Continue', 'Продолжить'), variant: HmVariant.primary, loading: _busy, expand: hm.mobile, onPressed: _submit),
          ],
        ),
      ],
    );
  }

  /// A code to scan with the phone. Light quiet zone around dark modules:
  /// scanners need the contrast, whatever the theme.
  Widget _qrCard(String data) => HmCard(
    recessed: true,
    padding: const EdgeInsets.all(16),
    child: Column(
      children: [
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(color: HmColors.text, borderRadius: BorderRadius.circular(8)),
          child: QrImageView(
            data: data,
            size: 220,
            padding: EdgeInsets.zero,
            backgroundColor: HmColors.text,
            eyeStyle: const QrEyeStyle(eyeShape: QrEyeShape.square, color: HmColors.bg),
            dataModuleStyle: const QrDataModuleStyle(dataModuleShape: QrDataModuleShape.square, color: HmColors.bg),
          ),
        ),
        const SizedBox(height: 12),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const SizedBox.square(dimension: 16, child: HmSpinner()),
            const SizedBox(width: 12),
            Flexible(child: Text(tr('Waiting for the scan…', 'Ждём сканирования…'), style: HmText.bodyStrong)),
          ],
        ),
      ],
    ),
  );

  Widget _doneView(pb.Account a) {
    final name = a.name.isNotEmpty ? a.name : a.login;
    final who = a.login.isNotEmpty && a.login != name ? '$name (${a.login})' : name;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        HmBanner(tr('Signed in: $who', 'Вход выполнен: $who'), tone: HmTone.ok),
        const SizedBox(height: 16),
        HmListItem(
          leading: HmAvatar(label: name, icon: providerIcon(a.provider), color: providerColor(a.provider)),
          title: Text(name),
          subtitle: Text(providerName(a.provider, _messenger.state.provider(a.provider))),
        ),
        const SizedBox(height: 16),
        _Actions(
          children: [
            HmButton(
              tr('Done', 'Готово'),
              variant: HmVariant.primary,
              expand: Hm.of(context).mobile,
              onPressed: () => Navigator.of(context).pop(a),
            ),
          ],
        ),
      ],
    );
  }
}

/// Buttons at the end of a step: right-aligned, stacked on phones.
class _Actions extends StatelessWidget {
  const _Actions({required this.children});

  final List<Widget> children;

  @override
  Widget build(BuildContext context) {
    if (Hm.of(context).mobile) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          for (final (i, c) in children.reversed.indexed) ...[if (i > 0) const SizedBox(height: 8), c],
        ],
      );
    }
    return Row(
      mainAxisAlignment: MainAxisAlignment.end,
      children: [
        for (final (i, c) in children.indexed) ...[if (i > 0) const SizedBox(width: 12), c],
      ],
    );
  }
}
