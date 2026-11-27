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
import '../../data/rpc/gen/parley.pbgrpc.dart' as pb;
import '../common/hm.dart';
import '../core/core_cubit.dart';

typedef _S = pb.ShardState_Status;

/// "Sync via Shard" (CONNECTOR.md): the vault file kept in the user's
/// Shard vault. The switch shows only when Shard answers here or
/// sync is already on.
bool showShardSwitch(pb.ShardState cs) => cs.enabled || (cs.available && cs.status != _S.NO_VAULT);

/// Opens the Shard app (Android, `ShardState.can_open`).
Future<void> openShard() async {
  try {
    await const MethodChannel('dev.selfhold.parley/core').invokeMethod<void>('openShard');
  } catch (e) {
    Toasts.error('${tr('Could not open Shard', 'Не удалось открыть Shard')}: ${errorText(e)}');
  }
}

/// The Settings section: the switch, then [ShardStatus].
class ShardSection extends StatefulWidget {
  const ShardSection({super.key});

  @override
  State<ShardSection> createState() => _ShardSectionState();
}

class _ShardSectionState extends State<ShardSection> {
  late final CoreCubit _core = context.read<CoreCubit>();
  bool _busy = false;

  Future<void> _set(bool on, pb.ShardState cs) async {
    if (_busy) return;
    if (!on && cs.status == _S.PAIRING) return _run(() => _core.api.cancelShardPairing(pb.Empty()));
    if (on) return _run(() => _core.api.enableShard(pb.Empty()));
    final remove = await _confirmDisable();
    if (remove == null) return;
    await _run(() => _core.api.disableShard(pb.DisableShardRequest(removeSynced: remove)));
  }

  Future<void> _run(Future<pb.ShardState> Function() call) async {
    setState(() => _busy = true);
    try {
      _core.adoptShard(await call());
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
              label: tr('Also remove the synced vault from Shard', 'Также удалить синхронизированное хранилище из Shard'),
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
    final cs = context.watch<CoreCubit>().state.state?.shard ?? pb.ShardState();
    final gap = SizedBox(height: Hm.of(context).gapSm);
    return HmCard(
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          HmSectionTitle('Shard'),
          gap,
          if (showShardSwitch(cs)) ...[
            HmToggleRow(
              label: tr('Sync via Shard', 'Синхронизировать через Shard'),
              hint: tr(
                'The vault is kept in your Shard vault and every device gets the same accounts.',
                'Хранилище хранится в вашем хранилище Shard, и все устройства получают одни и те же аккаунты.',
              ),
              value: cs.enabled || cs.status == _S.PAIRING,
              onChanged: _busy ? null : (v) => _set(v, cs),
            ),
            gap,
          ],
          ShardStatus(cs),
        ],
      ),
    );
  }
}

/// What Shard is doing and what the user can do about it: the pairing
/// code, a conflict, a password, a message. Used in Settings and on the
/// welcome screen ("Get the vault from Shard").
class ShardStatus extends StatefulWidget {
  const ShardStatus(this.cs, {super.key});

  final pb.ShardState cs;

  @override
  State<ShardStatus> createState() => _ShardStatusState();
}

class _ShardStatusState extends State<ShardStatus> {
  late final CoreCubit _core = context.read<CoreCubit>();
  bool _busy = false;

  Future<void> _run(Future<pb.ShardState> Function() call) async {
    if (_busy) return;
    setState(() => _busy = true);
    try {
      _core.adoptShard(await call());
    } catch (e) {
      Toasts.error(errorText(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _resolve(bool canMerge) async {
    final choice = await showHmModal<pb.ResolveShardRequest_Choice>(
      context,
      title: tr('The synced vault differs', 'Синхронизированное хранилище отличается'),
      body: (ctx) {
        final mobile = Hm.of(ctx).mobile;
        Widget option(pb.ResolveShardRequest_Choice c, String label, String hint, HmVariant variant) => Padding(
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
                'Shard has a vault that is not the one on this device. Choose which to keep.',
                'В Shard лежит хранилище, отличное от хранилища на этом устройстве. Выберите, какое оставить.',
              ),
              style: HmText.body.copyWith(height: 1.5),
            ),
            const SizedBox(height: 16),
            if (canMerge)
              option(
                pb.ResolveShardRequest_Choice.MERGE,
                tr('Merge accounts', 'Объединить аккаунты'),
                tr('Both vaults keep their accounts.', 'Сохраняются аккаунты обоих хранилищ.'),
                HmVariant.primary,
              ),
            option(
              pb.ResolveShardRequest_Choice.USE_SYNCED,
              tr('Use the synced vault', 'Взять синхронизированное'),
              tr(
                'It replaces this one; a backup of this device’s vault is kept. Its password may differ.',
                'Оно заменит это; копия хранилища этого устройства сохранится. Его пароль может быть другим.',
              ),
              canMerge ? HmVariant.cyan : HmVariant.primary,
            ),
            option(
              pb.ResolveShardRequest_Choice.UPLOAD,
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
    await _run(() => _core.api.resolveShard(pb.ResolveShardRequest(choice: choice)));
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
              _core.adoptShard(await _core.api.shardPassword(pb.PasswordRequest(password: c.text)));
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
                  'The vault in Shard was made with another password. Enter it to sync.',
                  'Хранилище в Shard создано с другим паролем. Введите его, чтобы синхронизировать.',
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
        ? HmButton(tr('Open Shard', 'Открыть Shard'), icon: Icons.open_in_new, expand: mobile, onPressed: openShard)
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
            : tr('Shard no longer allows this device: sync is off.', 'Shard больше не разрешает это устройство: синхронизация выключена.'),
        HmTone.warning,
      ),
      _S.NOT_RUNNING when !cs.enabled => muted(
        tr(
          'Shard not found: install or start it to sync the vault between devices.',
          'Shard не найден: установите или запустите его, чтобы синхронизировать хранилище между устройствами.',
        ),
      ),
      _S.NOT_RUNNING => banner(
        tr('Shard is not running here; changes wait until it starts.', 'Shard здесь не запущен; изменения ждут его запуска.'),
        HmTone.warning,
      ),
      _S.NO_VAULT => muted(tr('Set up Shard first: it has no vault yet.', 'Сначала настройте Shard: в нём ещё нет хранилища.')),
      _S.LOCKED => banner(tr('Unlock Shard to sync.', 'Разблокируйте Shard для синхронизации.'), HmTone.warning),
      _S.OFFLINE => banner(tr('Shard has no storage yet; changes wait.', 'У Shard пока нет хранилища; изменения ждут.'), HmTone.info),
      _S.PAIRING => banner('${tr('Code', 'Код')} ${cs.pairingCode} — ${tr('confirm in Shard', 'подтвердите в Shard')}', HmTone.accent, [
        HmButton(
          tr('Cancel', 'Отмена'),
          variant: HmVariant.neutral,
          small: true,
          onPressed: _busy ? null : () => _run(() => _core.api.cancelShardPairing(pb.Empty())),
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
      _S.ERROR => banner(cs.message.isNotEmpty ? cs.message : tr('Shard failed', 'Ошибка Shard'), HmTone.error),
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
        if (showDevice) ...[const SizedBox(height: 4), muted('${tr('This device in Shard', 'Это устройство в Shard')}: ${cs.deviceName}')],
      ],
    );
  }
}
