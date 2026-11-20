import 'dart:async';

import 'package:equatable/equatable.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/layout.dart';
import '../../data/rpc/core_client.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../common/hm.dart';
import '../core/core_cubit.dart';
import '../core/messenger_cubit.dart';

class TopicsState extends Equatable {
  const TopicsState({this.topics = const [], this.selected, this.loading = true, this.error, this.rev = 0});

  /// Pinned first, then the latest activity first (as the core orders them).
  final List<pb.Topic> topics;

  /// The topic shown; null: all messages of the chat.
  final String? selected;
  final bool loading;
  final String? error;
  final int rev;

  pb.Topic? get current => topics.where((t) => t.id == selected).firstOrNull;

  TopicsState copyWith({List<pb.Topic>? topics, String? selected, bool clearSelected = false, bool? loading, String? error, bool clearError = false}) =>
      TopicsState(
        topics: topics ?? this.topics,
        selected: clearSelected ? null : selected ?? this.selected,
        loading: loading ?? this.loading,
        error: clearError ? null : error ?? this.error,
        rev: rev + 1,
      );

  @override
  List<Object?> get props => [selected, loading, error, rev];
}

/// The topics of a forum chat and which one is shown. The list reloads
/// when messages arrive (unread counts and order change with them).
class TopicsCubit extends Cubit<TopicsState> {
  TopicsCubit(this._core, this.key) : super(const TopicsState()) {
    _events = _core.events.listen(_onEvent);
    load();
  }

  final CoreCubit _core;
  final ChatKey key;
  late final StreamSubscription<pb.Event> _events;
  Timer? _reload;

  Future<void> load() async {
    try {
      final r = await _core.api.topics(pb.TopicsRequest(accountId: key.account, chatId: key.chat));
      if (isClosed) return;
      final open = state.selected;
      emit(
        state.copyWith(
          topics: r.topics,
          loading: false,
          clearError: true,
          // A topic that is gone (deleted) falls back to all messages.
          clearSelected: open != null && !r.topics.any((t) => t.id == open),
        ),
      );
    } catch (e) {
      if (!isClosed) emit(state.copyWith(loading: false, error: errorText(e)));
    }
  }

  void select(String? id) => emit(id == null ? state.copyWith(clearSelected: true) : state.copyWith(selected: id));

  /// The open topic was read: its badge goes now, not with the next reload.
  void markedRead(String id) {
    final list = [
      for (final t in state.topics)
        if (t.id == id && t.unread > 0) (t.deepCopy()..unread = 0) else t,
    ];
    emit(state.copyWith(topics: list));
  }

  void _onEvent(pb.Event ev) {
    final chat = switch (ev.kind) {
      pb.Event_Kind.KIND_MESSAGE_NEW || pb.Event_Kind.KIND_MESSAGE_DELETE => (account: ev.message.accountId, chat: ev.message.chatId),
      pb.Event_Kind.KIND_CHAT => (account: ev.chat.accountId, chat: ev.chat.id),
      _ => null,
    };
    if (chat != key) return;
    _reload?.cancel();
    _reload = Timer(const Duration(seconds: 1), () {
      if (!isClosed) load();
    });
  }

  @override
  Future<void> close() async {
    _reload?.cancel();
    await _events.cancel();
    return super.close();
  }
}

/// The strip of a forum's topics under the chat header: all messages, then
/// each topic with its unread count.
class TopicBar extends StatelessWidget {
  const TopicBar({super.key, required this.chat});

  final pb.Chat? chat;

  @override
  Widget build(BuildContext context) {
    final hm = Hm.of(context);
    final s = context.watch<TopicsCubit>().state;
    final cubit = context.read<TopicsCubit>();
    final pad = hm.mobile ? 12.0 : 24.0;
    final topics = s.topics.where((t) => !t.hidden || t.id == s.selected || t.unread > 0).toList();
    return Container(
      height: hm.buttonSm + 16,
      decoration: const BoxDecoration(
        color: HmColors.surface,
        border: Border(bottom: BorderSide(color: HmColors.border)),
      ),
      child: ListView(
        scrollDirection: Axis.horizontal,
        padding: EdgeInsets.symmetric(horizontal: pad, vertical: 8),
        children: [
          HmChip(
            label: tr('All messages', 'Все сообщения'),
            active: s.selected == null,
            leading: Icon(Icons.forum_outlined, size: 14, color: s.selected == null ? HmColors.accent : HmColors.muted),
            onTap: () => cubit.select(null),
          ),
          for (final t in topics) ...[
            const SizedBox(width: 8),
            HmChip(
              label: t.title,
              active: t.id == s.selected,
              count: t.unread,
              leading: Icon(
                t.closed ? Icons.lock_outline : (t.pinned ? Icons.push_pin_outlined : Icons.tag),
                size: 14,
                color: t.id == s.selected ? HmColors.accent : HmColors.muted,
              ),
              onTap: () => cubit.select(t.id),
            ),
          ],
          if (s.loading && s.topics.isEmpty) ...[const SizedBox(width: 12), const Center(child: SizedBox.square(dimension: 16, child: HmSpinner()))],
          if (s.error != null) ...[
            const SizedBox(width: 8),
            HmChip(
              label: tr('Topics not loaded — retry', 'Темы не загрузились — повторить'),
              active: false,
              leading: const Icon(Icons.refresh, size: 14, color: HmColors.muted),
              onTap: cubit.load,
            ),
          ],
        ],
      ),
    );
  }
}
