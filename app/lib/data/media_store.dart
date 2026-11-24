import 'dart:async';
import 'dart:io';
import 'dart:typed_data';

import 'package:fixnum/fixnum.dart';
import 'package:grpc/grpc.dart';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';

import 'local_files.dart';
import 'rpc/gen/megachat.pbgrpc.dart' as pb;

/// Avatars, previews and files of messages, fetched from the core (which
/// keeps its own encrypted cache) and held in memory while the app runs.
/// Nothing is written to disk except a file the user opens.
class MediaStore {
  MediaStore(this._api);

  final pb.MegaChatClient Function() _api;

  /// Memory kept for recently shown pictures.
  static const _cap = 96 << 20;

  final _cache = <String, Uint8List>{}; // insertion order: least recently used first
  int _bytes = 0;
  final _pending = <String, Future<Uint8List?>>{};
  // A ref that failed is not asked again for a while (an avatar in a long
  // list would otherwise be retried on every scroll).
  final _failed = <String, DateTime>{};

  static String _key(String account, String ref) => '$account\u0000$ref';

  /// The content if it is in memory already (no flicker on rebuilds).
  Uint8List? peek(String account, String ref) {
    final k = _key(account, ref);
    final v = _cache.remove(k);
    if (v != null) _cache[k] = v; // most recently used last
    return v;
  }

  /// The content of a media ref; null when it cannot be had (offline,
  /// unsupported, deleted). Larger than [maxBytes] throws [MediaTooLarge].
  Future<Uint8List?> load(String account, String ref, {int maxBytes = 0, bool keep = true}) {
    if (ref.isEmpty) return Future.value();
    final hit = peek(account, ref);
    if (hit != null) return Future.value(hit);
    final k = _key(account, ref);
    final failed = _failed[k];
    if (failed != null && DateTime.now().difference(failed) < const Duration(minutes: 2)) return Future.value();
    // The action must not return the removed future: whenComplete would
    // wait for it, that is for itself.
    return _pending[k] ??= _fetch(account, ref, maxBytes, keep).whenComplete(() {
      _pending.remove(k);
    });
  }

  Future<Uint8List?> _fetch(String account, String ref, int maxBytes, bool keep) async {
    final k = _key(account, ref);
    try {
      final r = await _api().media(pb.MediaRequest(accountId: account, ref: ref, maxBytes: Int64(maxBytes)));
      final data = Uint8List.fromList(r.data);
      _failed.remove(k);
      if (keep && data.length < _cap ~/ 8) _put(k, data);
      return data;
    } on GrpcError catch (e) {
      if (e.code == StatusCode.resourceExhausted) throw MediaTooLarge();
      _failed[k] = DateTime.now();
      return null;
    } catch (_) {
      _failed[k] = DateTime.now();
      return null;
    }
  }

  /// Media the core caches (larger ones are fetched each time, so they
  /// are not worth fetching ahead).
  static const prefetchMax = 8 << 20;

  final _prefetched = <String>{};

  /// Has the core fetch a media ref into its own cache, so that playing or
  /// opening it later needs no download; nothing comes to the app. [size]
  /// is the file's size when known (0: unknown).
  void prefetch(String account, String ref, {int size = 0}) {
    if (ref.isEmpty || size > prefetchMax) return;
    final k = _key(account, ref);
    if (_cache.containsKey(k) || !_prefetched.add(k)) return;
    _api()
        .media(pb.MediaRequest(accountId: account, ref: ref, maxBytes: Int64(prefetchMax), prefetch: true))
        .then((_) {}, onError: (_) => _prefetched.remove(k));
  }

  void _put(String k, Uint8List v) {
    _cache[k] = v;
    _bytes += v.length;
    while (_bytes > _cap && _cache.isNotEmpty) {
      final first = _cache.keys.first;
      _bytes -= _cache.remove(first)!.length;
    }
  }

  /// Writes a file to open it with another app: a temporary folder of
  /// the app, removed by [clear] and on the next start.
  Future<File?> toFile(String account, String ref, String name, {int maxBytes = 512 << 20}) async {
    final data = await load(account, ref, maxBytes: maxBytes, keep: false);
    if (data == null) return null;
    final dir = Directory(p.join((await getTemporaryDirectory()).path, 'megachat-open', '${ref.hashCode.toUnsigned(32)}'));
    await dir.create(recursive: true);
    final f = File(p.join(dir.path, _safeName(name)));
    await f.writeAsBytes(data, flush: true);
    return f;
  }

  /// Saves a file to [path], or without one to the Downloads folder (the
  /// shared Downloads on Android), under a name not taken yet. Returns
  /// where it went.
  Future<String?> save(String account, String ref, String name, {String? path, String mime = '', int maxBytes = 512 << 20}) async {
    final data = await load(account, ref, maxBytes: maxBytes, keep: false);
    if (data == null) return null;
    if (path != null) {
      await File(path).writeAsBytes(data, flush: true);
      return path;
    }
    if (Platform.isAndroid) return saveToAndroidDownloads(_safeName(name), mime, data);
    Directory? dir;
    try {
      dir = await getDownloadsDirectory();
    } catch (_) {}
    dir ??= await getApplicationDocumentsDirectory();
    await dir.create(recursive: true);
    var f = File(p.join(dir.path, _safeName(name)));
    for (var i = 1; await f.exists(); i++) {
      final base = p.basenameWithoutExtension(name), ext = p.extension(name);
      f = File(p.join(dir.path, _safeName('$base ($i)$ext')));
    }
    await f.writeAsBytes(data, flush: true);
    return f.path;
  }

  /// Where downloads go by default (desktops), to start a "Save as" there.
  static Future<String?> downloadsDir() async {
    try {
      return (await getDownloadsDirectory())?.path;
    } catch (_) {
      return null;
    }
  }

  static String _safeName(String name) {
    final n = name.replaceAll(RegExp(r'[\\/:*?"<>|\x00-\x1f]'), '_').trim();
    return n.isEmpty ? 'file' : n;
  }

  /// Forgets everything (the vault was locked).
  Future<void> clear() async {
    _cache.clear();
    _bytes = 0;
    _failed.clear();
    _prefetched.clear();
    try {
      final dir = Directory(p.join((await getTemporaryDirectory()).path, 'megachat-open'));
      if (await dir.exists()) await dir.delete(recursive: true);
    } catch (_) {}
  }
}

class MediaTooLarge implements Exception {
  @override
  String toString() => 'file is too large';
}
