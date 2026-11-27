package dev.selfhold.parley

import android.app.Activity
import android.content.ClipboardManager
import android.content.ContentValues
import android.net.Uri
import android.os.Build
import android.os.Environment
import android.provider.MediaStore
import android.provider.OpenableColumns
import io.flutter.plugin.common.MethodCall
import io.flutter.plugin.common.MethodChannel
import java.io.File
import java.util.concurrent.Executor

// "dev.selfhold.parley/files" (lib/data/local_files.dart): files on the clipboard,
// and saving a file into the shared Downloads.
class Files(private val activity: Activity, private val worker: Executor) : MethodChannel.MethodCallHandler {
    // What one message may carry (the core's hub.MaxUpload).
    private val maxUpload = 100L shl 20

    override fun onMethodCall(call: MethodCall, result: MethodChannel.Result) {
        when (call.method) {
            "clipboard" -> clipboard(result)
            "saveDownload" -> {
                val name = call.argument<String>("name") ?: "file"
                val mime = call.argument<String>("mime") ?: ""
                val data = call.argument<ByteArray>("data") ?: ByteArray(0)
                run(result) { saveDownload(name, mime, data) }
            }
            else -> result.notImplemented()
        }
    }

    private fun run(result: MethodChannel.Result, block: () -> Any?) = worker.execute {
        try {
            val v = block()
            activity.runOnUiThread { result.success(v) }
        } catch (t: Throwable) {
            activity.runOnUiThread { result.error("files", t.message, null) }
        }
    }

    // The clipboard is read on the main thread (Android lets only the
    // focused app read it); the content of its URIs in the background.
    private fun clipboard(result: MethodChannel.Result) {
        val cm = activity.getSystemService(ClipboardManager::class.java)
        val clip = runCatching { cm?.primaryClip }.getOrNull()
        val uris = mutableListOf<Uri>()
        if (clip != null) for (i in 0 until clip.itemCount) clip.getItemAt(i).uri?.let { uris.add(it) }
        if (uris.isEmpty()) return result.success(emptyList<Any>())
        run(result) {
            uris.mapNotNull { uri ->
                val r = activity.contentResolver
                var name: String? = null
                var size = -1L
                runCatching {
                    r.query(uri, arrayOf(OpenableColumns.DISPLAY_NAME, OpenableColumns.SIZE), null, null, null)?.use { c ->
                        if (c.moveToFirst()) {
                            name = c.getString(0)
                            if (!c.isNull(1)) size = c.getLong(1)
                        }
                    }
                }
                if (size > maxUpload) throw IllegalArgumentException("too large: $name")
                val data = runCatching { r.openInputStream(uri)?.use { it.readBytes() } }.getOrNull() ?: return@mapNotNull null
                mapOf("name" to (name ?: uri.lastPathSegment ?: "file"), "mime" to (r.getType(uri) ?: ""), "data" to data)
            }
        }
    }

    // Android 10+: MediaStore's Downloads, no permission needed. Earlier:
    // the app's own Downloads folder on the shared storage.
    private fun saveDownload(name: String, mime: String, data: ByteArray): String {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            val r = activity.contentResolver
            val values = ContentValues().apply {
                put(MediaStore.Downloads.DISPLAY_NAME, name)
                if (mime.isNotEmpty()) put(MediaStore.Downloads.MIME_TYPE, mime)
                put(MediaStore.Downloads.IS_PENDING, 1)
            }
            val uri = r.insert(MediaStore.Downloads.EXTERNAL_CONTENT_URI, values) ?: throw IllegalStateException("no Downloads")
            try {
                r.openOutputStream(uri)?.use { it.write(data) } ?: throw IllegalStateException("cannot write")
                r.update(uri, ContentValues().apply { put(MediaStore.Downloads.IS_PENDING, 0) }, null, null)
            } catch (t: Throwable) {
                r.delete(uri, null, null)
                throw t
            }
            return "${Environment.DIRECTORY_DOWNLOADS}/$name"
        }
        val dir = activity.getExternalFilesDir(Environment.DIRECTORY_DOWNLOADS) ?: activity.filesDir
        dir.mkdirs()
        val base = name.substringBeforeLast('.', name)
        val ext = name.substringAfterLast('.', "")
        var f = File(dir, name)
        var i = 1
        while (f.exists()) f = File(dir, if (ext.isEmpty()) "$base (${i++})" else "$base (${i++}).$ext")
        f.writeBytes(data)
        return f.path
    }
}
