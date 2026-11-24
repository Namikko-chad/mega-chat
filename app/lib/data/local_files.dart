import 'dart:ffi';
import 'dart:io';
import 'dart:typed_data';
import 'dart:ui' as ui;

import 'package:ffi/ffi.dart';
import 'package:file_selector/file_selector.dart';
import 'package:flutter/services.dart';
import 'package:path/path.dart' as p;

/// A file to send, held in memory until it is sent (nothing is written to
/// disk). [mime] may be empty: the core guesses it.
class LocalFile {
  const LocalFile(this.name, this.mime, this.data);

  final String name;
  final String mime;
  final Uint8List data;

  bool get isImage => const {'image/png', 'image/jpeg', 'image/gif', 'image/webp'}.contains(mime) || _imageExt.contains(p.extension(name).toLowerCase());

  static const _imageExt = {'.png', '.jpg', '.jpeg', '.gif', '.webp'};
}

/// What one message may carry in files, all together (the core's
/// hub.MaxUpload).
const maxUpload = 100 << 20;

/// A file over [maxUpload].
class FileTooLarge implements Exception {
  FileTooLarge(this.name);

  final String name;
}

/// Opens the system's file dialog; empty when the user cancelled.
Future<List<LocalFile>> pickLocalFiles() async {
  final picked = await openFiles();
  final out = <LocalFile>[];
  for (final x in picked) {
    if (await x.length() > maxUpload) throw FileTooLarge(x.name);
    out.add(LocalFile(x.name, x.mimeType ?? '', await x.readAsBytes()));
  }
  return out;
}

/// Asks where to save a file (desktops); null: cancelled.
Future<String?> askSavePath(String name, {String? directory}) async {
  final loc = await getSaveLocation(suggestedName: name, initialDirectory: directory);
  return loc?.path;
}

/// Asks for an existing vault file (desktops); null: cancelled.
Future<String?> pickVaultFile() async {
  final f = await openFile(acceptedTypeGroups: const [XTypeGroup(label: 'MegaChat vault', extensions: ['mc'])]);
  return f?.path;
}

/// Saving lets the user pick the place (desktops; Android saves to its
/// Downloads collection instead).
bool get canAskSavePath => Platform.isLinux || Platform.isWindows || Platform.isMacOS;

const _channel = MethodChannel('dev.megachat/files');

/// Android: puts a file into the shared Downloads collection and returns
/// its name there.
Future<String?> saveToAndroidDownloads(String name, String mime, Uint8List data) =>
    _channel.invokeMethod<String>('saveDownload', {'name': name, 'mime': mime, 'data': data});

/// Files copied in a file manager, or a picture (a screenshot, an image
/// copied in a browser) on the clipboard; empty when it holds neither, or
/// only text (then a paste is a text paste). A picture that comes with
/// text (cells of a spreadsheet, a part of a document) counts as text.
Future<List<LocalFile>> clipboardFiles() async {
  try {
    if (Platform.isWindows) return await _windowsClipboard();
    if (Platform.isLinux) return await _linuxClipboard();
    if (Platform.isAndroid) return await _androidClipboard();
  } on FileTooLarge {
    rethrow;
  } catch (_) {}
  return const [];
}

Future<List<LocalFile>> _filesAt(Iterable<String> paths) async {
  final out = <LocalFile>[];
  for (final path in paths) {
    final f = File(path);
    if (!await f.exists()) continue; // a folder, or gone
    if (await f.length() > maxUpload) throw FileTooLarge(p.basename(path));
    out.add(LocalFile(p.basename(path), '', await f.readAsBytes()));
  }
  return out;
}

/// The GTK runner answers with "uris" (files copied) or "png" (a picture).
Future<List<LocalFile>> _linuxClipboard() async {
  final r = await _channel.invokeMapMethod<String, Object?>('clipboard');
  if (r == null) return const [];
  final uris = (r['uris'] as List?)?.cast<String>() ?? const [];
  final paths = [
    for (final u in uris)
      if (u.startsWith('file://')) Uri.parse(u).toFilePath(),
  ];
  if (paths.isNotEmpty) return _filesAt(paths);
  final png = r['png'];
  if (png is Uint8List && png.isNotEmpty) return [LocalFile(_pastedName('png'), 'image/png', png)];
  return const [];
}

Future<List<LocalFile>> _androidClipboard() async {
  final r = await _channel.invokeListMethod<Map<Object?, Object?>>('clipboard');
  return [
    for (final m in r ?? const <Map<Object?, Object?>>[])
      if (m['data'] is Uint8List) LocalFile(m['name'] as String? ?? 'file', m['mime'] as String? ?? '', m['data'] as Uint8List),
  ];
}

/// "Pasted 2026-11-20 14.03.07.png", the way screenshots are named.
String _pastedName(String ext) {
  final t = DateTime.now();
  String two(int v) => v.toString().padLeft(2, '0');
  return 'Pasted ${t.year}-${two(t.month)}-${two(t.day)} ${two(t.hour)}.${two(t.minute)}.${two(t.second)}.$ext';
}

// --- Windows -------------------------------------------------------------

const _cfUnicodeText = 13, _cfHdrop = 15, _cfDib = 8;

typedef _OpenNative = Int32 Function(IntPtr);
typedef _Open = int Function(int);
typedef _CloseNative = Int32 Function();
typedef _Close = int Function();
typedef _FormatNative = Int32 Function(Uint32);
typedef _Format = int Function(int);
typedef _DataNative = IntPtr Function(Uint32);
typedef _Data = int Function(int);
typedef _RegisterNative = Uint32 Function(Pointer<Utf16>);
typedef _Register = int Function(Pointer<Utf16>);
typedef _LockNative = Pointer<Uint8> Function(IntPtr);
typedef _Lock = Pointer<Uint8> Function(int);
typedef _UnlockNative = Int32 Function(IntPtr);
typedef _Unlock = int Function(int);
typedef _SizeNative = IntPtr Function(IntPtr);
typedef _Size = int Function(int);
typedef _DragNative = Uint32 Function(IntPtr, Uint32, Pointer<Utf16>, Uint32);
typedef _Drag = int Function(int, int, Pointer<Utf16>, int);

/// Reads the clipboard through user32 directly: CF_HDROP (files copied in
/// Explorer), then a "PNG" (browsers, the Snipping Tool), then CF_DIB.
Future<List<LocalFile>> _windowsClipboard() async {
  final user32 = DynamicLibrary.open('user32.dll');
  final kernel32 = DynamicLibrary.open('kernel32.dll');
  final shell32 = DynamicLibrary.open('shell32.dll');
  final open = user32.lookupFunction<_OpenNative, _Open>('OpenClipboard');
  final close = user32.lookupFunction<_CloseNative, _Close>('CloseClipboard');
  final has = user32.lookupFunction<_FormatNative, _Format>('IsClipboardFormatAvailable');
  final get = user32.lookupFunction<_DataNative, _Data>('GetClipboardData');
  final register = user32.lookupFunction<_RegisterNative, _Register>('RegisterClipboardFormatW');
  final lock = kernel32.lookupFunction<_LockNative, _Lock>('GlobalLock');
  final unlock = kernel32.lookupFunction<_UnlockNative, _Unlock>('GlobalUnlock');
  final size = kernel32.lookupFunction<_SizeNative, _Size>('GlobalSize');
  final drag = shell32.lookupFunction<_DragNative, _Drag>('DragQueryFileW');

  Uint8List? copy(int handle) {
    if (handle == 0) return null;
    final ptr = lock(handle);
    if (ptr == nullptr) return null;
    try {
      return Uint8List.fromList(ptr.asTypedList(size(handle)));
    } finally {
      unlock(handle);
    }
  }

  final pngName = 'PNG'.toNativeUtf16();
  final paths = <String>[];
  Uint8List? png, dib;
  if (open(0) == 0) {
    calloc.free(pngName);
    return const [];
  }
  try {
    if (has(_cfHdrop) != 0) {
      final drop = get(_cfHdrop);
      final n = drop == 0 ? 0 : drag(drop, 0xFFFFFFFF, nullptr, 0);
      for (var i = 0; i < n; i++) {
        final len = drag(drop, i, nullptr, 0);
        final buf = calloc<Uint16>(len + 1).cast<Utf16>();
        try {
          drag(drop, i, buf, len + 1);
          paths.add(buf.toDartString());
        } finally {
          calloc.free(buf);
        }
      }
    } else if (has(_cfUnicodeText) == 0) {
      final fmt = register(pngName);
      if (fmt != 0 && has(fmt) != 0) {
        png = copy(get(fmt));
      } else if (has(_cfDib) != 0) {
        dib = copy(get(_cfDib));
      }
    }
  } finally {
    close();
    calloc.free(pngName);
  }
  if (paths.isNotEmpty) return _filesAt(paths);
  png ??= dib == null ? null : await _dibToPng(dib);
  if (png == null || png.isEmpty) return const [];
  return [LocalFile(_pastedName('png'), 'image/png', png)];
}

/// A CF_DIB (BITMAPINFOHEADER and pixels) as PNG. 24 and 32 bit pictures
/// are read here (a 32 bit one with no alpha at all is opaque); others go
/// through the engine's BMP decoder.
Future<Uint8List?> _dibToPng(Uint8List dib) async {
  if (dib.length < 40) return null;
  final h = ByteData.sublistView(dib);
  final hdr = h.getUint32(0, Endian.little);
  final w = h.getInt32(4, Endian.little), rawH = h.getInt32(8, Endian.little);
  final bits = h.getUint16(14, Endian.little), comp = h.getUint32(16, Endian.little);
  final colors = h.getUint32(32, Endian.little);
  final height = rawH.abs(), topDown = rawH < 0;
  if (w <= 0 || height == 0 || w * height > 64 << 20) return null;
  var offset = hdr;
  if (hdr == 40 && comp == 3) offset += 12; // BI_BITFIELDS masks after the header
  if (bits <= 8) offset += (colors == 0 ? 1 << bits : colors) * 4;
  ui.Image? img;
  if ((bits == 24 || bits == 32) && (comp == 0 || comp == 3)) {
    final stride = ((w * bits + 31) ~/ 32) * 4;
    if (offset + stride * height > dib.length) return null;
    final rgba = Uint8List(w * height * 4);
    var alpha = false;
    for (var y = 0; y < height; y++) {
      final src = offset + (topDown ? y : height - 1 - y) * stride;
      for (var x = 0; x < w; x++) {
        final s = src + x * (bits ~/ 8), d = (y * w + x) * 4;
        rgba[d] = dib[s + 2];
        rgba[d + 1] = dib[s + 1];
        rgba[d + 2] = dib[s];
        rgba[d + 3] = bits == 32 ? dib[s + 3] : 255;
        if (bits == 32 && dib[s + 3] != 0) alpha = true;
      }
    }
    if (bits == 32 && !alpha) {
      for (var i = 3; i < rgba.length; i += 4) {
        rgba[i] = 255;
      }
    }
    final buf = await ui.ImmutableBuffer.fromUint8List(rgba);
    final desc = ui.ImageDescriptor.raw(buf, width: w, height: height, pixelFormat: ui.PixelFormat.rgba8888);
    img = (await (await desc.instantiateCodec()).getNextFrame()).image;
  } else {
    // A BMP file is the DIB behind a 14 byte file header.
    final bmp = Uint8List(14 + dib.length);
    ByteData.sublistView(bmp)
      ..setUint8(0, 0x42)
      ..setUint8(1, 0x4D)
      ..setUint32(2, bmp.length, Endian.little)
      ..setUint32(10, 14 + offset, Endian.little);
    bmp.setRange(14, bmp.length, dib);
    final codec = await ui.instantiateImageCodec(bmp);
    img = (await codec.getNextFrame()).image;
  }
  final out = await img.toByteData(format: ui.ImageByteFormat.png);
  img.dispose();
  return out?.buffer.asUint8List();
}
