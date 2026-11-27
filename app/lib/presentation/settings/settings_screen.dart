import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:path/path.dart' as p;

import '../../core/desktop/desktop_window.dart';
import '../../core/format.dart';
import '../../core/i18n.dart';
import '../../core/settings/app_settings.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/layout.dart';
import '../../core/theme/typography.dart';
import '../../data/biometric.dart';
import '../../data/local_files.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../accounts/providers.dart' show isOnline;
import '../common/hm.dart';
import '../core/core_cubit.dart';
import '../core/messenger_cubit.dart';
import '../shell/app_shell.dart' show AppShell, BuildInfo, Section, StatusPills;
import '../unlock/unlock_screen.dart' show rememberHint, rememberPassword;
import 'cloudsync.dart';

/// This device, app preferences and the vault password: the "profile" of
/// the app (§ 4.6). On phones it also carries the footer's status (App. D).
class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key, required this.onLock});

  final VoidCallback onLock;

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  final _old = TextEditingController();
  final _new = TextEditingController();
  final _repeat = TextEditingController();
  final _errors = <String, String>{};
  String? _formError;
  bool _changing = false;

  // Read once in initState: handlers use it after awaits, when the screen
  // may already be gone and context.read would throw.
  late final CoreCubit _core;

  @override
  void initState() {
    super.initState();
    _core = context.read<CoreCubit>();
    // A fingerprint may have been enrolled since the app started.
    Biometric.refresh().then((_) {
      if (mounted) setState(() {});
    });
  }

  @override
  void dispose() {
    for (final c in [_old, _new, _repeat]) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _changePassword() async {
    _errors.clear();
    if (_old.text.isEmpty) _errors['old'] = tr('Enter the current password', 'Введите текущий пароль');
    if (_new.text.length < 8) _errors['new'] = tr('At least 8 characters', 'Не короче 8 символов');
    if (_repeat.text != _new.text) _errors['repeat'] = tr('Passwords do not match', 'Пароли не совпадают');
    setState(() => _formError = null);
    if (_errors.isNotEmpty) return;
    setState(() => _changing = true);
    try {
      await _core.api.changePassword(pb.ChangePasswordRequest(oldPassword: _old.text, newPassword: _new.text));
      // The fingerprint keeps the old password: seal the new one (or drop it).
      if (Biometric.enabled && await Biometric.enable(_new.text) != BioResult.ok) await Biometric.disable();
      if (mounted) setState(() {});
      for (final c in [_old, _new, _repeat]) {
        c.clear();
      }
      Toasts.success(tr('Password changed', 'Пароль изменён'));
    } catch (e) {
      setState(() => _formError = errorText(e));
    } finally {
      if (mounted) setState(() => _changing = false);
    }
  }

  /// Moves the vault file wherever the user likes (another disk, a synced
  /// folder); the core keeps the place for the next start.
  Future<void> _moveVault(String current) async {
    final path = await askSavePath(p.basename(current), directory: p.dirname(current));
    if (path == null) return;
    try {
      await _core.setVaultPath(path, move: true);
      Toasts.success(tr('The vault is moved', 'Хранилище перемещено'));
    } catch (e) {
      Toasts.error('${tr('Could not move the vault', 'Не удалось переместить хранилище')}: ${errorText(e)}');
    }
  }

  Future<void> _setRemember(bool on) async {
    if (!on) {
      try {
        _core.adopt(await _core.api.forgetPassword(pb.Empty()));
        Toasts.success(tr('The password is forgotten', 'Пароль забыт'));
      } catch (e) {
        Toasts.error('${tr('Could not forget the password', 'Не удалось забыть пароль')}: ${errorText(e)}');
      }
      return;
    }
    final password = await _askPassword(tr('Remember the password', 'Запомнить пароль'));
    if (password == null) return;
    final st = await rememberPassword(_core.api, password);
    if (st == null) return;
    _core.adopt(st);
    Toasts.success(tr('The password is remembered on this machine', 'Пароль запомнен на этом компьютере'));
  }

  Future<void> _setBiometric(bool on) async {
    if (!on) {
      await Biometric.disable();
      await AppSettings.instance.setBiometricDeclined(true);
      if (mounted) setState(() {});
      return;
    }
    final password = await _askPassword(tr('Unlock with fingerprint', 'Вход по отпечатку'));
    if (password == null || !mounted) return;
    final r = await Biometric.enable(password);
    if (r == BioResult.ok) {
      await AppSettings.instance.setBiometricDeclined(false);
      Toasts.success(tr('Fingerprint unlock is on', 'Вход по отпечатку включён'));
    } else if (r == BioResult.failed) {
      Toasts.error(tr('Could not turn on fingerprint unlock', 'Не удалось включить вход по отпечатку'));
    }
    if (mounted) setState(() {});
  }

  /// The vault password, to be sealed for fingerprint unlock or remembered.
  /// The core checks it first, so a typo is not saved.
  Future<String?> _askPassword(String title) {
    final c = TextEditingController();
    String? error;
    return showHmModal<String>(
      context,
      title: title,
      size: HmModalSize.sm,
      body: (ctx) => StatefulBuilder(
        builder: (ctx, set) {
          Future<void> submit() async {
            if (c.text.isEmpty) return set(() => error = tr('Enter the password', 'Введите пароль'));
            try {
              await _core.api.checkPassword(pb.PasswordRequest(password: c.text));
              if (ctx.mounted) Navigator.pop(ctx, c.text);
            } catch (e) {
              set(() => error = errorText(e));
            }
          }

          return Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              HmTextField(
                label: tr('Vault password', 'Пароль хранилища'),
                required: true,
                password: true,
                autofocus: true,
                controller: c,
                error: error,
                onSubmitted: (_) => submit(),
              ),
              const SizedBox(height: 16),
              HmButton(tr('Continue', 'Продолжить'), variant: HmVariant.primary, onPressed: submit),
            ],
          );
        },
      ),
    ).whenComplete(c.dispose);
  }

  Widget _preferences(pb.State s, Widget gap) {
    final settings = AppSettings.instance;
    return ListenableBuilder(
      listenable: settings,
      builder: (context, _) => HmCard(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            HmSectionTitle(tr('Settings', 'Настройки')),
            gap,
            HmField(
              tr('Language', 'Язык'),
              Align(
                alignment: Alignment.centerLeft,
                child: HmSegmented<Lang>(options: const {Lang.en: 'English', Lang.ru: 'Русский'}, value: settings.language, onChanged: settings.setLanguage),
              ),
            ),
            gap,
            HmField(
              tr('Notifications', 'Уведомления'),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  HmSegmented<NotifyMode>(
                    options: {
                      NotifyMode.all: tr('All messages', 'Все сообщения'),
                      NotifyMode.mentions: tr('Mentions', 'Упоминания'),
                      NotifyMode.off: tr('Off', 'Выключены'),
                    },
                    value: settings.notify,
                    onChanged: settings.setNotify,
                  ),
                  const SizedBox(height: 6),
                  Text(
                    tr(
                      'Shown while MegaChat is not on screen. Muted chats notify only when someone mentions you.',
                      'Показываются, когда MegaChat не на экране. Беззвучные чаты уведомляют, только если вас упомянули.',
                    ),
                    style: HmText.jb(11, color: HmColors.muted, height: 1.5),
                  ),
                ],
              ),
            ),
            gap,
            HmToggleRow(
              label: tr('Group the chat list', 'Группировать список чатов'),
              hint: tr(
                'Channels and groups first, then direct messages, each under its own heading.',
                'Сначала каналы и группы, затем личные сообщения — каждые под своим заголовком.',
              ),
              value: settings.groupChats,
              onChanged: settings.setGroupChats,
            ),
            gap,
            HmToggleRow(
              label: tr('Load pictures automatically', 'Загружать картинки автоматически'),
              hint: tr('Off: a picture in a chat loads on a tap.', 'Выключено: картинка в чате загружается по нажатию.'),
              value: settings.autoPhotos,
              onChanged: settings.setAutoPhotos,
            ),
            gap,
            HmToggleRow(
              label: tr('Link previews', 'Превью ссылок'),
              hint: tr(
                'A link in a message shows the title and picture of its page. The page is loaded from this device, so the site sees its address.',
                'Ссылка в сообщении показывает заголовок и картинку своей страницы. Страница загружается с этого устройства, поэтому сайт видит его адрес.',
              ),
              value: settings.linkPreviews,
              onChanged: settings.setLinkPreviews,
            ),
            gap,
            HmField(
              tr('Videos in chats', 'Видео в чатах'),
              Align(
                alignment: Alignment.centerLeft,
                child: HmSegmented<VideoAuto>(
                  options: {
                    VideoAuto.none: tr('On tap', 'По нажатию'),
                    VideoAuto.preview: tr('Preview', 'Превью'),
                    VideoAuto.full: tr('Preview and video', 'Превью и видео'),
                  },
                  value: settings.autoVideo,
                  onChanged: settings.setAutoVideo,
                ),
              ),
            ),
            gap,
            HmToggleRow(
              label: tr('Load voice and audio ahead', 'Загружать голосовые и аудио заранее'),
              hint: tr(
                'On: small voice messages and audio files are fetched as they are shown, so they play at once.',
                'Включено: небольшие голосовые и аудиофайлы скачиваются при показе и играют сразу.',
              ),
              value: settings.autoAudio,
              onChanged: settings.setAutoAudio,
            ),
            if (Biometric.available) ...[
              gap,
              HmToggleRow(
                label: tr('Unlock with fingerprint', 'Вход по отпечатку'),
                hint: tr(
                  'The password is kept encrypted in the Android Keystore and opens only after a fingerprint check.',
                  'Пароль хранится зашифрованным в Android Keystore и открывается только после проверки отпечатка.',
                ),
                value: Biometric.enabled,
                onChanged: _setBiometric,
              ),
            ],
            if (s.canRemember) ...[
              gap,
              HmToggleRow(
                label: tr('Remember the password on this machine', 'Запомнить пароль на этом компьютере'),
                hint: rememberHint(),
                value: s.remembered,
                onChanged: _setRemember,
              ),
            ],
            if (DesktopWindow.supported) ...[
              gap,
              HmToggleRow(
                label: tr('Minimize to tray on close', 'Сворачивать в трей при закрытии'),
                hint: tr(
                  'The close button hides the window and the accounts stay connected; quit from the tray icon menu.',
                  'Кнопка закрытия прячет окно, аккаунты остаются на связи; выйти можно из меню значка в трее.',
                ),
                value: settings.closeToTray,
                onChanged: settings.setCloseToTray,
              ),
            ],
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    final core = context.watch<CoreCubit>().state;
    final s = core.state ?? pb.State();
    // The live account list, as in the footer: State is a snapshot taken
    // at unlock, while the accounts were still connecting.
    final accounts = context.watch<MessengerCubit>().state.accounts.where((a) => a.enabled).toList();
    final gap = SizedBox(height: hm.gapSm);
    return HmPage(
      children: [
        HmPageTitle(tr('Settings', 'Настройки')),
        HmCard(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              HmSectionTitle(tr('This device', 'Это устройство')),
              gap,
              HmField(tr('Platform', 'Платформа'), Text(platformName(s.platform))),
              HmField(tr('Core', 'Ядро'), Text(s.coreVersion.isEmpty ? '—' : '${s.coreVersion} ${s.coreBuild}')),
              HmField(
                tr('Vault file', 'Файл хранилища'),
                Wrap(
                  spacing: 12,
                  runSpacing: 8,
                  crossAxisAlignment: WrapCrossAlignment.center,
                  children: [
                    SelectableText(s.vaultPath.isEmpty ? '—' : s.vaultPath),
                    if (canAskSavePath && s.vaultPath.isNotEmpty) HmButton(tr('Move…', 'Переместить…'), icon: Icons.drive_file_move_outline, onPressed: () => _moveVault(s.vaultPath)),
                  ],
                ),
              ),
              HmField(tr('Accounts', 'Аккаунты'), Text('${accounts.where(isOnline).length} / ${accounts.length} ${tr('online', 'в сети')}')),
              const SizedBox(height: 12),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  HmButton(tr('Accounts', 'Аккаунты'), icon: Section.accounts.icon, expand: hm.mobile, onPressed: () => AppShell.go(context, Section.accounts)),
                  HmButton(tr('Log', 'Журнал'), icon: Section.log.icon, expand: hm.mobile, onPressed: () => AppShell.go(context, Section.log)),
                ],
              ),
            ],
          ),
        ),
        if (s.cloudsync.status != pb.CloudSyncState_Status.STATUS_UNSPECIFIED) const CloudSyncSection(),
        _preferences(s, gap),
        HmCard(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              HmSectionTitle(tr('Password', 'Пароль')),
              gap,
              Text(
                tr(
                  'The password encrypts the vault with the sessions of all accounts on this device.',
                  'Пароль шифрует хранилище с сессиями всех аккаунтов на этом устройстве.',
                ),
                style: HmText.small.copyWith(height: 1.5),
              ),
              gap,
              if (_formError != null) ...[HmBanner(_formError!, tone: HmTone.error), gap],
              HmTextField(label: tr('Current password', 'Текущий пароль'), required: true, password: true, controller: _old, error: _errors['old']),
              gap,
              HmTextField(label: tr('New password', 'Новый пароль'), required: true, password: true, controller: _new, error: _errors['new']),
              gap,
              HmTextField(label: tr('Repeat', 'Повторите'), required: true, password: true, controller: _repeat, error: _errors['repeat']),
              const SizedBox(height: 16),
              Align(
                alignment: Alignment.centerRight,
                child: HmButton(
                  tr('Change password', 'Сменить пароль'),
                  variant: HmVariant.primary,
                  loading: _changing,
                  expand: hm.mobile,
                  onPressed: _changePassword,
                ),
              ),
            ],
          ),
        ),
        if (hm.mobile) ...[
          // The footer's slots live here on phones (M6, App. D).
          HmCard(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                HmSectionTitle(tr('Status', 'Статус')),
                gap,
                DefaultTextStyle(
                  style: HmText.footer,
                  child: BuildInfo(core: core),
                ),
                const SizedBox(height: 12),
                StatusPills(core: core),
              ],
            ),
          ),
          HmButton(tr('Lock', 'Заблокировать'), variant: HmVariant.red, icon: Icons.lock_outline, expand: true, onPressed: widget.onLock),
        ],
      ],
    );
  }
}
