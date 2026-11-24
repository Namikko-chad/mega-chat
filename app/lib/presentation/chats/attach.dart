import 'package:flutter/material.dart';

import '../../core/format.dart';
import '../../core/i18n.dart';
import '../../core/theme/colors.dart';
import '../../core/theme/layout.dart';
import '../../data/local_files.dart';
import '../../data/rpc/gen/megachat.pbgrpc.dart' as pb;
import '../common/hm.dart';

/// The files a composer sends with its next message, picked from the disk,
/// pasted (Ctrl+V) or inserted by the keyboard. In memory only.
class PendingFiles extends ValueNotifier<List<LocalFile>> {
  PendingFiles([super.value = const []]);

  bool get isEmpty => value.isEmpty;

  int get bytes => value.fold(0, (n, f) => n + f.data.length);

  /// Adds [files]; refuses them all (with a toast) when the message would
  /// carry more than [maxUpload].
  bool add(List<LocalFile> files) {
    if (files.isEmpty) return false;
    final total = bytes + files.fold(0, (n, f) => n + f.data.length);
    if (total > maxUpload) {
      Toasts.error(
        tr('Too much for one message', 'Слишком много для одного сообщения'),
        tr('Up to ${formatSize(maxUpload)} of files', 'Не больше ${formatSize(maxUpload)} файлов'),
      );
      return false;
    }
    value = [...value, ...files];
    return true;
  }

  void removeAt(int i) => value = [...value]..removeAt(i);

  /// Takes the files out (to send them).
  List<LocalFile> take() {
    final v = value;
    value = const [];
    return v;
  }
}

/// Opens the file dialog and adds what was picked.
Future<void> attachFromDisk(PendingFiles files) async {
  try {
    files.add(await pickLocalFiles());
  } on FileTooLarge catch (e) {
    _tooLarge(e.name);
  } catch (e) {
    Toasts.error(tr('Could not read the file', 'Не удалось прочитать файл'), '$e');
  }
}

/// Adds the files or the picture on the clipboard; false when there are
/// none (the paste is a text paste then). [tellEmpty] says so in a toast.
Future<bool> attachFromClipboard(PendingFiles files, {bool tellEmpty = false}) async {
  try {
    final got = await clipboardFiles();
    if (got.isEmpty) {
      if (tellEmpty) Toasts.show(tr('No files on the clipboard', 'В буфере обмена нет файлов'));
      return false;
    }
    files.add(got);
    return true;
  } on FileTooLarge catch (e) {
    _tooLarge(e.name);
    return true; // files were there: no text paste
  }
}

/// A picture the keyboard inserted (Android keyboards' clipboard, GIFs).
void attachInserted(PendingFiles files, KeyboardInsertedContent c) {
  final data = c.data;
  if (data == null || data.isEmpty) return;
  if (data.length > maxUpload) return _tooLarge('');
  var name = Uri.tryParse(c.uri)?.pathSegments.lastOrNull ?? '';
  if (!name.contains('.')) {
    final ext = c.mimeType.split('/').last.split('+').first;
    name = '${name.isEmpty ? 'image' : name}.$ext';
  }
  files.add([LocalFile(name, c.mimeType, data)]);
}

void _tooLarge(String name) => Toasts.error(
  tr('The file is too large', 'Файл слишком большой'),
  [if (name.isNotEmpty) name, tr('up to ${formatSize(maxUpload)}', 'не больше ${formatSize(maxUpload)}')].join(' · '),
);

/// The paper clip. Desktops open the file dialog right away (Ctrl+V pastes
/// files); phones offer the files or what the clipboard holds.
class AttachButton extends StatelessWidget {
  const AttachButton({super.key, required this.files, this.enabled = true});

  final PendingFiles files;
  final bool enabled;

  @override
  Widget build(BuildContext context) {
    final mobile = Hm.of(context).mobile;
    return HmIconButton(
      icon: Icons.attach_file,
      label: mobile ? tr('Attach', 'Прикрепить') : tr('Attach files · Ctrl+V pastes', 'Прикрепить файлы · Ctrl+V — вставить'),
      onPressed: !enabled
          ? null
          : () async {
              if (!mobile) return attachFromDisk(files);
              final v = await pickOption<String>(
                context,
                title: tr('Attach', 'Прикрепить'),
                options: {'disk': tr('Files', 'Файлы'), 'clipboard': tr('Paste from the clipboard', 'Вставить из буфера обмена')},
              );
              if (v == 'disk') await attachFromDisk(files);
              if (v == 'clipboard') await attachFromClipboard(files, tellEmpty: true);
            },
    );
  }
}

/// The pending files over the text field, each with a button to drop it.
class PendingFilesStrip extends StatelessWidget {
  const PendingFilesStrip({super.key, required this.files, this.enabled = true});

  final PendingFiles files;
  final bool enabled;

  @override
  Widget build(BuildContext context) => ValueListenableBuilder<List<LocalFile>>(
    valueListenable: files,
    builder: (context, list, _) {
      if (list.isEmpty) return const SizedBox.shrink();
      return Padding(
        padding: const EdgeInsets.only(bottom: 6),
        child: Wrap(
          spacing: 6,
          runSpacing: 2,
          children: [
            for (final (i, f) in list.indexed)
              ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 280),
                child: HmAttachment(
                  icon: f.isImage ? Icons.image_outlined : Icons.insert_drive_file_outlined,
                  leading: f.isImage
                      ? ClipRRect(
                          borderRadius: BorderRadius.circular(3),
                          child: Image.memory(
                            f.data,
                            width: 20,
                            height: 20,
                            fit: BoxFit.cover,
                            cacheWidth: 40,
                            errorBuilder: (_, _, _) => const Icon(Icons.image_outlined, size: 16, color: HmColors.accent),
                          ),
                        )
                      : null,
                  name: f.name,
                  detail: formatSize(f.data.length),
                  trailing: HmIconButton(dense: true, icon: Icons.close, label: tr('Remove', 'Убрать'), onPressed: enabled ? () => files.removeAt(i) : null),
                ),
              ),
          ],
        ),
      );
    },
  );
}

/// Whether the account can send files.
bool canSendFiles(pb.Account? a) => (a?.capabilities ?? const []).contains(pb.Capability.CAPABILITY_FILES);
