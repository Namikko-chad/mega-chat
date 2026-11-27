import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:grpc/grpc.dart';

import '../../core/format.dart';
import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/layout.dart';
import '../../core/theme/typography.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../common/hm.dart';
import '../core/core_cubit.dart';

typedef _S = pb.CloudSyncState_Status;

/// "Sync via CloudSync" (CONNECTOR.md): the vault file kept in the user's
/// CloudSync vault. The switch shows only when CloudSync answers here or
/// sync is already on.
bool showCloudSyncSwitch(pb.CloudSyncState cs) => cs.enabled || (cs.available && cs.status != _S.NO_VAULT);

/// Opens the CloudSync app (Android, `CloudSyncState.can_open`).
Future<void> openCloudSync() async {
  try {
    await const MethodChannel('dev.megachat/core').invokeMethod<void>('openCloudSync');
  } catch (e) {
    Toasts.error('${tr('Could not open CloudSync', 'Не удалось открыть CloudSync')}: ${errorText(e)}');
  }
}

/// The Settings section: the switch, then [CloudSyncStatus].
class CloudSyncSection extends StatefulWidget {
  const CloudSyncSection({super.key});

  @override
  State<CloudSyncSection> createState() => _CloudSyncSectionState();
}

class _CloudSyncSectionState extends State<CloudSyncSection> {
  late final CoreCubit _core = context.read<CoreCubit>();
  bool _busy = false;

  Future<void> _set(bool on, pb.CloudSyncState cs) async {
    if (_busy) return;
    if (!on && cs.status == _S.PAIRING) return _run(() => _core.api.cancelCloudSyncPairing(pb.Empty()));
    if (on) return _run(() => _core.api.enableCloudSync(pb.Empty()));
    final remove = await _confirmDisable();
    if (remove == null) return;
    await _run(() => _core.api.disableCloudSync(pb.DisableCloudSyncRequest(removeSynced: remove)));
  }

  Future<void> _run(Future<pb.CloudSyncState> Function() call) async {
    setState(() => _busy = true);
    try {
      _core.adoptCloudSync(await call());
    } catch (e) {
      Toasts.error(errorText(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  /// Null: cancelled; else whether to remove the synced copy too.
  Future<bool?> _confirmDisable() {
    var remove = false;
    return showHmModal<bool>(
      context,
      title: tr('Turn off sync?', 'Выключить синхронизацию?'),
      size: HmModalSize.sm,
      body: (ctx) => StatefulBuilder(
        builder: (ctx, set) => Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              tr(
                'The vault stays on this device; changes no longer reach your other devices.',
                'Хранилище остаётся на этом устройстве; изменения больше не дойдут до других ваших устройств.',
              ),
              style: HmText.body.copyWith(height: 1.5),
            ),
            const SizedBox(height: 12),
            HmCheckRow(
              label: tr('Also remove the synced vault from CloudSync', 'Также удалить синхронизированное хранилище из CloudSync'),
              value: remove,
              onChanged: (v) => set(() => remove = v),
            ),
          ],
        ),
      ),
      actions: (ctx) => [
        HmButton(tr('Cancel', 'Отмена'), variant: HmVariant.neutral, expand: Hm.of(ctx).mobile, onPressed: () => Navigator.pop(ctx)),
        HmButton(tr('Turn off', 'Выключить'), variant: HmVariant.red, expand: Hm.of(ctx).mobile, onPressed: () => Navigator.pop(ctx, remove)),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    final cs = context.watch<CoreCubit>().state.state?.cloudsync ?? pb.CloudSyncState();
    final gap = SizedBox(height: Hm.of(context).gapSm);
    return HmCard(
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          HmSectionTitle('CloudSync'),
          gap,
          if (showCloudSyncSwitch(cs)) ...[
            HmToggleRow(
              label: tr('Sync via CloudSync', 'Синхронизировать через CloudSync'),
              hint: tr(
                'The vault is kept in your CloudSync vault and every device gets the same accounts.',
                'Хранилище хранится в вашем хранилище CloudSync, и все устройства получают одни и те же аккаунты.',
              ),
              value: cs.enabled || cs.status == _S.PAIRING,
              onChanged: _busy ? null : (v) => _set(v, cs),
            ),
            gap,
          ],
          CloudSyncStatus(cs),
        ],
      ),
    );
  }
}

/// What CloudSync is doing and what the user can do about it: the pairing
/// code, a conflict, a password, a message. Used in Settings and on the
/// welcome screen ("Get the vault from CloudSync").
class CloudSyncStatus extends StatefulWidget {
  const CloudSyncStatus(this.cs, {super.key});

  final pb.CloudSyncState cs;

  @override
  State<CloudSyncStatus> createState() => _CloudSyncStatusState();
}

class _CloudSyncStatusState extends State<CloudSyncStatus> {
  late final CoreCubit _core = context.read<CoreCubit>();
  bool _busy = false;

  Future<void> _run(Future<pb.CloudSyncState> Function() call) async {
    if (_busy) return;
    setState(() => _busy = true);
    try {
      _core.adoptCloudSync(await call());
    } catch (e) {
      Toasts.error(errorText(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _resolve(bool canMerge) async {
    final choice = await showHmModal<pb.ResolveCloudSyncRequest_Choice>(
      context,
      title: tr('The synced vault differs', 'Синхронизированное хранилище отличается'),
      body: (ctx) {
        final mobile = Hm.of(ctx).mobile;
        Widget option(pb.ResolveCloudSyncRequest_Choice c, String label, String hint, HmVariant variant) => Padding(
          padding: const EdgeInsets.only(bottom: 12),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              HmButton(label, variant: variant, expand: mobile, onPressed: () => Navigator.pop(ctx, c)),
              const SizedBox(height: 4),
              Text(hint, style: HmText.small.copyWith(height: 1.5)),
            ],
          ),
        );
        return Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              tr(
                'CloudSync has a vault that is not the one on this device. Choose which to keep.',
                'В CloudSync лежит хранилище, отличное от хранилища на этом устройстве. Выберите, какое оставить.',
              ),
              style: HmText.body.copyWith(height: 1.5),
            ),
            const SizedBox(height: 16),
            if (canMerge)
              option(
                pb.ResolveCloudSyncRequest_Choice.MERGE,
                tr('Merge accounts', 'Объединить аккаунты'),
                tr('Both vaults keep their accounts.', 'Сохраняются аккаунты обоих хранилищ.'),
                HmVariant.primary,
              ),
            option(
              pb.ResolveCloudSyncRequest_Choice.USE_SYNCED,
              tr('Use the synced vault', 'Взять синхронизированное'),
              tr(
                'It replaces this one; a backup of this device’s vault is kept. Its password may differ.',
                'Оно заменит это; копия хранилища этого устройства сохранится. Его пароль может быть другим.',
              ),
              canMerge ? HmVariant.cyan : HmVariant.primary,
            ),
            option(
              pb.ResolveCloudSyncRequest_Choice.UPLOAD,
              tr('Upload this device’s vault', 'Загрузить хранилище этого устройства'),
              tr('It replaces the synced one on every device.', 'Оно заменит синхронизированное на всех устройствах.'),
              HmVariant.cyan,
            ),
          ],
        );
      },
      actions: (ctx) => [HmButton(tr('Later', 'Позже'), variant: HmVariant.neutral, expand: Hm.of(ctx).mobile, onPressed: () => Navigator.pop(ctx))],
    );
    if (choice == null) return;
    await _run(() => _core.api.resolveCloudSync(pb.ResolveCloudSyncRequest(choice: choice)));
  }

  Future<void> _password() {
    final c = TextEditingController();
    String? error;
    var busy = false;
    return showHmModal<void>(
      context,
      title: tr('Password of the synced vault', 'Пароль синхронизированного хранилища'),
      size: HmModalSize.sm,
      body: (ctx) => StatefulBuilder(
        builder: (ctx, set) {
          Future<void> submit() async {
            if (c.text.isEmpty) return set(() => error = tr('Enter the password', 'Введите пароль'));
            set(() => busy = true);
            try {
              _core.adoptCloudSync(await _core.api.cloudSyncPassword(pb.PasswordRequest(password: c.text)));
              if (ctx.mounted) Navigator.pop(ctx);
            } catch (e) {
              final wrong = e is GrpcError && e.code == StatusCode.permissionDenied;
              if (ctx.mounted) set(() => error = wrong ? tr('Wrong password', 'Неверный пароль') : errorText(e));
            } finally {
              if (ctx.mounted) set(() => busy = false);
            }
          }

          return Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                tr(
                  'The vault in CloudSync was made with another password. Enter it to sync.',
                  'Хранилище в CloudSync создано с другим паролем. Введите его, чтобы синхронизировать.',
                ),
                style: HmText.small.copyWith(height: 1.5),
              ),
              const SizedBox(height: 12),
              HmTextField(
                label: tr('Password of the synced vault', 'Пароль синхронизированного хранилища'),
                required: true,
                password: true,
                autofocus: true,
                controller: c,
                error: error,
                onSubmitted: (_) => submit(),
              ),
              const SizedBox(height: 16),
              HmButton(tr('Continue', 'Продолжить'), variant: HmVariant.primary, loading: busy, onPressed: submit),
            ],
          );
        },
      ),
    ).whenComplete(c.dispose);
  }

  @override
  Widget build(BuildContext context) {
    final cs = widget.cs;
    final mobile = Hm.of(context).mobile;
    final open = cs.canOpen && (cs.status == _S.LOCKED || cs.status == _S.PAIRING)
        ? HmButton(tr('Open CloudSync', 'Открыть CloudSync'), icon: Icons.open_in_new, expand: mobile, onPressed: openCloudSync)
        : null;

    Widget muted(String text) => Text(text, style: HmText.jb(11, color: HmColors.muted, height: 1.5));
    Widget banner(String text, HmTone tone, [List<Widget> actions = const []]) => Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        HmBanner(text, tone: tone),
        if (actions.isNotEmpty || open != null) ...[
          const SizedBox(height: 8),
          Wrap(spacing: 8, runSpacing: 8, children: [...actions, ?open]),
        ],
      ],
    );

    final Widget main = switch (cs.status) {
      _ when cs.unpaired => banner(
        cs.message.isNotEmpty
            ? cs.message
            : tr('CloudSync no longer allows this device: sync is off.', 'CloudSync больше не разрешает это устройство: синхронизация выключена.'),
        HmTone.warning,
      ),
      _S.NOT_RUNNING when !cs.enabled => muted(
        tr(
          'CloudSync not found: install or start it to sync the vault between devices.',
          'CloudSync не найден: установите или запустите его, чтобы синхронизировать хранилище между устройствами.',
        ),
      ),
      _S.NOT_RUNNING => banner(
        tr('CloudSync is not running here; changes wait until it starts.', 'CloudSync здесь не запущен; изменения ждут его запуска.'),
        HmTone.warning,
      ),
      _S.NO_VAULT => muted(tr('Set up CloudSync first: it has no vault yet.', 'Сначала настройте CloudSync: в нём ещё нет хранилища.')),
      _S.LOCKED => banner(tr('Unlock CloudSync to sync.', 'Разблокируйте CloudSync для синхронизации.'), HmTone.warning),
      _S.OFFLINE => banner(tr('CloudSync has no storage yet; changes wait.', 'У CloudSync пока нет хранилища; изменения ждут.'), HmTone.info),
      _S.PAIRING => banner('${tr('Code', 'Код')} ${cs.pairingCode} — ${tr('confirm in CloudSync', 'подтвердите в CloudSync')}', HmTone.accent, [
        HmButton(
          tr('Cancel', 'Отмена'),
          variant: HmVariant.neutral,
          small: true,
          onPressed: _busy ? null : () => _run(() => _core.api.cancelCloudSyncPairing(pb.Empty())),
        ),
      ]),
      _S.SYNCING => banner(tr('Syncing…', 'Синхронизация…'), HmTone.info),
      _S.SYNCED => muted(cs.syncedMs > 0 ? '${tr('Synced', 'Синхронизировано')} ${formatChatTime(cs.syncedMs)}' : tr('Synced', 'Синхронизировано')),
      _S.CONFLICT => banner(tr('The synced vault differs from this one.', 'Синхронизированное хранилище отличается от этого.'), HmTone.warning, [
        HmButton(tr('Choose…', 'Выбрать…'), variant: HmVariant.primary, small: true, onPressed: _busy ? null : () => _resolve(cs.canMerge)),
      ]),
      _S.NEEDS_PASSWORD => banner(tr('The synced vault has another password.', 'У синхронизированного хранилища другой пароль.'), HmTone.warning, [
        HmButton(tr('Enter the password…', 'Ввести пароль…'), variant: HmVariant.primary, small: true, onPressed: _password),
      ]),
      _S.WAITING_UNLOCK => banner(tr('Unlock the vault here to merge.', 'Разблокируйте хранилище здесь, чтобы объединить.'), HmTone.info),
      _S.ERROR => banner(cs.message.isNotEmpty ? cs.message : tr('CloudSync failed', 'Ошибка CloudSync'), HmTone.error),
      _ when cs.message.isNotEmpty => banner(cs.message, HmTone.info),
      _ => const SizedBox.shrink(),
    };
    final showDevice = cs.enabled && cs.deviceName.isNotEmpty;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        main,
        // A message beside a status that has its own text (SYNCED with a note).
        if (cs.message.isNotEmpty && !cs.unpaired && cs.status == _S.SYNCED) ...[const SizedBox(height: 4), muted(cs.message)],
        if (showDevice) ...[const SizedBox(height: 4), muted('${tr('This device in CloudSync', 'Это устройство в CloudSync')}: ${cs.deviceName}')],
      ],
    );
  }
}
