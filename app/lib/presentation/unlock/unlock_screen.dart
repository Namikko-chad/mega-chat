import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:grpc/grpc.dart';

import '../../core/i18n.dart';
import '../../core/settings/app_settings.dart';
import '../../core/theme/layout.dart';
import '../../core/theme/typography.dart';
import '../../data/biometric.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../common/hm.dart';
import '../core/core_cubit.dart';

/// After the password was entered once: offers to unlock by fingerprint
/// next time (Android, a fingerprint is enrolled, not declined before).
Future<void> offerBiometric(BuildContext context, String password) async {
  await Biometric.refresh();
  if (!Biometric.available || Biometric.enabled || AppSettings.instance.biometricDeclined) return;
  if (!context.mounted) return;
  final yes = await confirm(
    context,
    title: tr('Unlock with fingerprint?', 'Входить по отпечатку?'),
    message: tr(
      'Next time the vault opens with your fingerprint. The password is kept encrypted on this phone, in the Android Keystore; it opens only after a fingerprint check.',
      'В следующий раз хранилище откроется по отпечатку пальца. Пароль хранится на этом телефоне в зашифрованном виде (Android Keystore) и открывается только после проверки отпечатка.',
    ),
    action: tr('Turn on', 'Включить'),
  );
  if (!yes) {
    await AppSettings.instance.setBiometricDeclined(true);
    return;
  }
  final r = await Biometric.enable(password);
  if (r == BioResult.ok) {
    Toasts.success(tr('Fingerprint unlock is on', 'Вход по отпечатку включён'));
  } else if (r == BioResult.failed) {
    Toasts.error(tr('Could not turn on fingerprint unlock', 'Не удалось включить вход по отпечатку'));
  }
}

/// Keeps the vault's key in the OS keychain after [password] opened it
/// (desktops, "Remember the password on this machine"). Returns the new
/// state, or null when the keychain refused (a toast says so).
Future<pb.State?> rememberPassword(pb.MegaChatClient api, String password) async {
  try {
    return await api.rememberPassword(pb.PasswordRequest(password: password));
  } catch (e) {
    Toasts.error('${tr('Could not remember the password', 'Не удалось запомнить пароль')}: ${errorText(e)}');
    return null;
  }
}

/// "Remember the password on this machine": the checkbox under a password
/// field, with what it means.
class _RememberBox extends StatelessWidget {
  const _RememberBox({required this.value, required this.onChanged});

  final bool value;
  final ValueChanged<bool> onChanged;

  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      HmCheckRow(label: tr('Remember the password on this machine', 'Запомнить пароль на этом компьютере'), value: value, onChanged: onChanged),
      if (value) ...[
        const SizedBox(height: 4),
        Text(rememberHint(), style: HmText.small.copyWith(height: 1.5)),
      ],
    ],
  );
}

/// What remembering the password means, for the checkbox and the setting.
String rememberHint() => tr(
  'The vault will open by itself when the app starts. A key made from the password is kept in the system keychain of your user account (${Platform.isWindows ? 'Windows Credential Manager' : 'Secret Service / GNOME Keyring / KWallet'}), not the password itself.',
  'Хранилище будет открываться само при запуске приложения. В системной связке ключей вашей учётной записи (${Platform.isWindows ? 'диспетчер учётных данных Windows' : 'Secret Service / GNOME Keyring / KWallet'}) хранится ключ, выведенный из пароля, а не сам пароль.',
);

/// Shown while the vault is locked, missing, or the core is unreachable.
class UnlockScreen extends StatelessWidget {
  const UnlockScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final core = context.watch<CoreCubit>().state;
    final hm = Hm.of(context);
    final Widget card = switch (core.conn) {
      // A retry keeps the error on screen instead of flashing a spinner.
      Conn.failed => const _NoCore(),
      Conn.connecting when core.error != null => const _NoCore(),
      Conn.connecting => const SizedBox(height: 240, child: HmLoading()),
      Conn.ready when core.state?.vaultExists ?? false => const _Unlock(),
      Conn.ready => const _Setup(),
    };
    return HmPage(
      children: [
        Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 520),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                SizedBox(height: hm.mobile ? 8 : 40),
                card,
              ],
            ),
          ),
        ),
      ],
    );
  }
}

class _NoCore extends StatelessWidget {
  const _NoCore();

  @override
  Widget build(BuildContext context) {
    final core = context.watch<CoreCubit>().state;
    final hint = Platform.isAndroid
        ? tr('The core runs inside the app; try again.', 'Ядро работает внутри приложения; попробуйте ещё раз.')
        : tr(
            'The core runs inside the app; megachat.log in the config folder says why it did not start.',
            'Ядро работает внутри приложения; почему оно не запустилось, видно в megachat.log в папке настроек.',
          );
    return HmCard(
      accent: HmAccent.error,
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          HmPageTitle(tr('No connection to the core', 'Нет связи с ядром')),
          const SizedBox(height: 16),
          HmBanner(core.error ?? tr('The core is not responding', 'Ядро не отвечает'), tone: HmTone.error),
          const SizedBox(height: 12),
          Text(hint, style: HmText.small.copyWith(height: 1.5)),
          const SizedBox(height: 20),
          Align(
            alignment: Alignment.centerRight,
            child: HmButton(
              tr('Retry', 'Повторить'),
              variant: HmVariant.primary,
              icon: Icons.refresh,
              loading: core.conn == Conn.connecting,
              onPressed: () => context.read<CoreCubit>().connect(),
            ),
          ),
        ],
      ),
    );
  }
}

class _Unlock extends StatefulWidget {
  const _Unlock();

  @override
  State<_Unlock> createState() => _UnlockState();
}

class _UnlockState extends State<_Unlock> {
  final _password = TextEditingController();
  bool _busy = false;
  bool _remember = false;
  String? _error;

  /// The remembered password is tried once, when the app starts; after a
  /// "Lock" the vault waits for a click.
  static bool _autoTried = false;

  bool get _bio => Biometric.available && Biometric.enabled;

  pb.State? get _state => context.read<CoreCubit>().state.state;

  @override
  void initState() {
    super.initState();
    if (_state?.remembered ?? false) {
      if (!_autoTried) {
        _autoTried = true;
        WidgetsBinding.instance.addPostFrameCallback((_) => _remembered());
      }
    } else if (_bio) {
      // Saved for fingerprint unlock: ask for the finger right away.
      WidgetsBinding.instance.addPostFrameCallback((_) => _fingerprint());
    }
    _autoTried = true;
  }

  @override
  void dispose() {
    _password.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (_password.text.isEmpty) {
      setState(() => _error = tr('Enter the password', 'Введите пароль'));
      return;
    }
    await _unlock(_password.text, byFingerprint: false);
  }

  Future<void> _fingerprint() async {
    if (_busy) return;
    setState(() => _error = null);
    final (password, r) = await Biometric.unlock();
    if (!mounted) return;
    switch (r) {
      case BioResult.ok:
        await _unlock(password!, byFingerprint: true);
      case BioResult.canceled:
        setState(() {});
      case BioResult.invalidated:
        setState(
          () => _error = tr(
            'Fingerprints on this phone have changed. Enter the password to turn fingerprint unlock on again.',
            'Отпечатки на телефоне изменились. Введите пароль, чтобы снова включить вход по отпечатку.',
          ),
        );
      case BioResult.failed:
        await Biometric.refresh();
        if (mounted) setState(() => _error = tr('Fingerprint check failed', 'Не удалось проверить отпечаток'));
    }
  }

  Future<void> _remembered() async {
    if (_busy) return;
    setState(() {
      _busy = true;
      _error = null;
    });
    final core = context.read<CoreCubit>();
    try {
      core.adopt(await core.api.unlockRemembered(pb.Empty()));
    } catch (e) {
      final code = e is GrpcError ? e.code : null;
      // Forgotten by the core (the password changed, the keychain lost it):
      // the state no longer says remembered.
      if (code == StatusCode.permissionDenied || code == StatusCode.notFound) {
        try {
          core.adopt(await core.api.getState(pb.Empty()));
        } catch (_) {}
      }
      if (mounted) {
        setState(
          () => _error = switch (code) {
            StatusCode.permissionDenied => tr(
              'The vault password has changed. Enter the new one.',
              'Пароль хранилища изменился. Введите новый.',
            ),
            StatusCode.notFound => tr('The password is no longer remembered. Enter it.', 'Пароль больше не запомнен. Введите его.'),
            _ => '${tr('The system keychain did not answer', 'Системная связка ключей не ответила')}: ${errorText(e)}',
          },
        );
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _unlock(String password, {required bool byFingerprint}) async {
    setState(() {
      _busy = true;
      _error = null;
    });
    final core = context.read<CoreCubit>();
    try {
      var st = await core.api.unlock(pb.PasswordRequest(password: password));
      if (_remember && st.canRemember) st = await rememberPassword(core.api, password) ?? st;
      if (!byFingerprint && mounted) await offerBiometric(context, password);
      core.adopt(st);
    } catch (e) {
      // The saved password no longer opens the vault (changed meanwhile).
      final stale = byFingerprint && e is GrpcError && e.code == StatusCode.permissionDenied;
      if (stale) await Biometric.disable();
      if (mounted) {
        setState(() => _error = stale ? tr('The vault password has changed. Enter the new one.', 'Пароль хранилища изменился. Введите новый.') : errorText(e));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    final st = context.watch<CoreCubit>().state.state;
    final path = st?.vaultPath ?? '';
    final remembered = st?.remembered ?? false;
    return HmCard(
      accent: HmAccent.total,
      padding: EdgeInsets.all(hm.mobile ? 16 : 24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          HmPageTitle(tr('Unlock vault', 'Разблокировать хранилище')),
          const SizedBox(height: 8),
          Text(path, style: HmText.small, overflow: TextOverflow.ellipsis),
          const SizedBox(height: 20),
          HmTextField(
            label: tr('Password', 'Пароль'),
            required: true,
            password: true,
            autofocus: !_bio,
            controller: _password,
            error: _error,
            onSubmitted: (_) => _submit(),
          ),
          if ((st?.canRemember ?? false) && !remembered) ...[
            SizedBox(height: hm.gapSm),
            _RememberBox(value: _remember, onChanged: (v) => setState(() => _remember = v)),
          ],
          const SizedBox(height: 24),
          _Actions(
            children: [
              if (remembered)
                HmButton(tr('Remembered password', 'Запомненный пароль'), icon: Icons.key, expand: hm.mobile, onPressed: _busy ? null : _remembered),
              if (_bio) HmButton(tr('Fingerprint', 'Отпечаток'), icon: Icons.fingerprint, expand: hm.mobile, onPressed: _busy ? null : _fingerprint),
              HmButton(tr('Unlock', 'Разблокировать'), variant: HmVariant.primary, loading: _busy, expand: hm.mobile, onPressed: _submit),
            ],
          ),
        ],
      ),
    );
  }
}

class _Setup extends StatefulWidget {
  const _Setup();

  @override
  State<_Setup> createState() => _SetupState();
}

class _SetupState extends State<_Setup> {
  final _password = TextEditingController();
  final _repeat = TextEditingController();
  bool _busy = false;
  bool _remember = false;
  String? _error;
  final _errors = <String, String>{};

  @override
  void dispose() {
    _password.dispose();
    _repeat.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    _errors.clear();
    if (_password.text.length < 8) _errors['password'] = tr('At least 8 characters', 'Не короче 8 символов');
    if (_repeat.text != _password.text) _errors['repeat'] = tr('Passwords don’t match', 'Пароли не совпадают');
    setState(() => _error = null);
    if (_errors.isNotEmpty) return;
    setState(() => _busy = true);
    final core = context.read<CoreCubit>();
    try {
      var st = await core.api.createVault(pb.PasswordRequest(password: _password.text));
      if (_remember && st.canRemember) st = await rememberPassword(core.api, _password.text) ?? st;
      if (mounted) await offerBiometric(context, _password.text);
      core.adopt(st);
    } catch (e) {
      if (mounted) setState(() => _error = errorText(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    return HmCard(
      accent: HmAccent.total,
      padding: EdgeInsets.all(hm.mobile ? 16 : 24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          HmPageTitle(tr('Welcome to MegaChat', 'Добро пожаловать в MegaChat')),
          const SizedBox(height: 8),
          Text(
            tr(
              'The vault keeps the sessions of your accounts on this device, encrypted with a password. The password is not stored anywhere: without it the accounts have to be added again.',
              'Хранилище держит сессии ваших аккаунтов на этом устройстве, зашифрованные паролем. Пароль нигде не сохраняется: без него аккаунты придётся добавлять заново.',
            ),
            style: HmText.small.copyWith(height: 1.5),
          ),
          const SizedBox(height: 20),
          if (_error != null) ...[HmBanner(_error!, tone: HmTone.error), SizedBox(height: hm.gapSm)],
          HmTextField(label: tr('Password', 'Пароль'), required: true, password: true, autofocus: true, controller: _password, error: _errors['password']),
          SizedBox(height: hm.gapSm),
          HmTextField(
            label: tr('Repeat password', 'Повторите пароль'),
            required: true,
            password: true,
            controller: _repeat,
            error: _errors['repeat'],
            onSubmitted: (_) => _submit(),
          ),
          if (context.watch<CoreCubit>().state.state?.canRemember ?? false) ...[
            SizedBox(height: hm.gapSm),
            _RememberBox(value: _remember, onChanged: (v) => setState(() => _remember = v)),
          ],
          const SizedBox(height: 24),
          _Actions(
            children: [HmButton(tr('Create vault', 'Создать хранилище'), variant: HmVariant.primary, loading: _busy, expand: hm.mobile, onPressed: _submit)],
          ),
        ],
      ),
    );
  }
}

class _Actions extends StatelessWidget {
  const _Actions({required this.children});

  final List<Widget> children;

  @override
  Widget build(BuildContext context) {
    if (Hm.of(context).mobile) return Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: children.reversed.toList());
    return Row(
      mainAxisAlignment: MainAxisAlignment.end,
      children: [
        for (final (i, c) in children.indexed) ...[if (i > 0) const SizedBox(width: 12), c],
      ],
    );
  }
}
