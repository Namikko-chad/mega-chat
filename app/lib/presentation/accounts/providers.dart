import 'package:flutter/material.dart';

import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../common/hm.dart';

/// How a provider looks in lists. Unknown types (a newer core) still get
/// a neutral look and the name the core gives them.
IconData providerIcon(String type) => switch (type) {
  'telegram' => Icons.send_outlined,
  'slack' => Icons.tag,
  'vkteams' => Icons.forum_outlined,
  'vk' => Icons.people_outline,
  'google' => Icons.alternate_email,
  'mailru' => Icons.mail_outline,
  _ => Icons.chat_bubble_outline,
};

Color providerColor(String type) => switch (type) {
  'telegram' => HmColors.active,
  'slack' => HmColors.purpleLight,
  'vkteams' => HmColors.accent,
  'vk' => HmColors.active,
  'google' => HmColors.ok,
  'mailru' => HmColors.warning,
  _ => HmColors.muted,
};

String providerName(String type, pb.ProviderInfo? info) =>
    info?.name.isNotEmpty ?? false
    ? info!.name
    : switch (type) {
        'telegram' => 'Telegram',
        'slack' => 'Slack',
        'vkteams' => 'VK Teams',
        'vk' => 'VK',
        'google' => 'Google',
        'mailru' => 'Mail.ru / VK WorkSpace',
        _ => type,
      };

/// The provider marker in the corner of an avatar.
Widget providerMarker(String type, {String? tooltip}) =>
    HmMarker(icon: providerIcon(type), color: providerColor(type), tooltip: tooltip);

/// The label and tone of an account's status.
(String, HmTone) accountStatus(pb.Account a) => switch (a.status) {
  pb.AccountStatus.ACCOUNT_STATUS_ONLINE => (tr('online', 'в сети'), HmTone.ok),
  pb.AccountStatus.ACCOUNT_STATUS_CONNECTING => (tr('connecting', 'подключение'), HmTone.info),
  pb.AccountStatus.ACCOUNT_STATUS_OFFLINE => (tr('offline', 'нет связи'), HmTone.warning),
  pb.AccountStatus.ACCOUNT_STATUS_NEEDS_LOGIN => (tr('needs sign-in', 'нужен вход'), HmTone.error),
  pb.AccountStatus.ACCOUNT_STATUS_DISABLED => (tr('disabled', 'выключен'), HmTone.neutral),
  _ => (tr('unknown', 'неизвестно'), HmTone.neutral),
};

bool isOnline(pb.Account a) => a.status == pb.AccountStatus.ACCOUNT_STATUS_ONLINE;
