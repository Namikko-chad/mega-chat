import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:grpc/grpc.dart';

import '../../core/settings/app_settings.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/typography.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../common/links.dart';
import '../core/core_cubit.dart';

/// Previews of links, fetched by the core (Open Graph tags of the page and
/// its picture) and kept for this run; a page without one is remembered too.
class LinkPreviews {
  LinkPreviews._();

  static const _cap = 300;
  static final _cache = <String, Future<pb.LinkPreviewReply?>>{};

  static Future<pb.LinkPreviewReply?> of(pb.MegaChatClient api, String url) {
    final hit = _cache.remove(url);
    if (hit != null) return _cache[url] = hit; // most recent last
    if (_cache.length >= _cap) _cache.remove(_cache.keys.first);
    final f = _fetch(api, url);
    _cache[url] = f;
    return f;
  }

  static Future<pb.LinkPreviewReply?> _fetch(pb.MegaChatClient api, String url) async {
    try {
      return await api.linkPreview(pb.LinkPreviewRequest(url: url));
    } on GrpcError catch (e) {
      // No preview: keep the answer. Anything else (no core): try later.
      if (e.code != StatusCode.notFound) _cache.remove(url);
      return null;
    } catch (_) {
      _cache.remove(url);
      return null;
    }
  }
}

/// The first web link of a message text to preview, if any.
Uri? previewLink(String text) {
  for (final l in findLinks(text)) {
    if (l.uri.scheme == 'http' || l.uri.scheme == 'https') return l.uri;
  }
  return null;
}

/// A link's preview under a message: site, title, description, picture.
/// Nothing while it loads or when the page has none; off in Settings.
class LinkPreviewCard extends StatefulWidget {
  const LinkPreviewCard({super.key, required this.uri, required this.maxWidth});

  final Uri uri;
  final double maxWidth;

  @override
  State<LinkPreviewCard> createState() => _LinkPreviewCardState();
}

class _LinkPreviewCardState extends State<LinkPreviewCard> {
  Future<pb.LinkPreviewReply?>? _preview;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _preview ??= LinkPreviews.of(context.read<CoreCubit>().api, widget.uri.toString());
  }

  @override
  void didUpdateWidget(LinkPreviewCard old) {
    super.didUpdateWidget(old);
    if (old.uri != widget.uri) _preview = LinkPreviews.of(context.read<CoreCubit>().api, widget.uri.toString());
  }

  @override
  Widget build(BuildContext context) => FutureBuilder<pb.LinkPreviewReply?>(
    future: _preview,
    builder: (context, snap) {
      final p = snap.data;
      if (p == null) return const SizedBox.shrink();
      return _card(p);
    },
  );

  Widget _card(pb.LinkPreviewReply p) => Padding(
    padding: const EdgeInsets.only(top: 6),
    child: ConstrainedBox(
      constraints: BoxConstraints(maxWidth: widget.maxWidth),
      child: MouseRegion(
        cursor: SystemMouseCursors.click,
        child: GestureDetector(
          onTap: () => openLink(widget.uri),
          child: Container(
            padding: const EdgeInsets.fromLTRB(10, 6, 8, 8),
            decoration: BoxDecoration(
              color: HmColors.tint(HmColors.accent, 0.06),
              borderRadius: BorderRadius.circular(6),
              border: const Border(left: BorderSide(color: HmColors.accent, width: 2)),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                if (p.siteName.isNotEmpty)
                  Text(p.siteName, style: HmText.jb(11, weight: 600, color: HmColors.accent), maxLines: 1, overflow: TextOverflow.ellipsis),
                if (p.title.isNotEmpty)
                  Padding(
                    padding: const EdgeInsets.only(top: 2),
                    child: Text(p.title, style: HmText.jb(13, weight: 600, color: HmColors.text), maxLines: 2, overflow: TextOverflow.ellipsis),
                  ),
                if (p.description.isNotEmpty)
                  Padding(
                    padding: const EdgeInsets.only(top: 2),
                    child: Text(p.description, style: HmText.small.copyWith(height: 1.4), maxLines: 3, overflow: TextOverflow.ellipsis),
                  ),
                if (p.image.isNotEmpty)
                  Padding(
                    padding: const EdgeInsets.only(top: 6),
                    child: ClipRRect(
                      borderRadius: BorderRadius.circular(4),
                      child: ConstrainedBox(
                        constraints: const BoxConstraints(maxHeight: 200),
                        child: Image.memory(
                          p.image is Uint8List ? p.image as Uint8List : Uint8List.fromList(p.image),
                          fit: BoxFit.cover,
                          width: widget.maxWidth,
                          gaplessPlayback: true,
                          errorBuilder: (_, _, _) => const SizedBox.shrink(),
                        ),
                      ),
                    ),
                  ),
              ],
            ),
          ),
        ),
      ),
    ),
  );
}

/// Whether link previews are on (Settings).
bool get linkPreviewsOn => AppSettings.instance.linkPreviews;
