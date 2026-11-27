package dev.selfhold.parley

import android.content.Context
import dev.selfhold.parley.mobilebind.Listener
import dev.selfhold.parley.mobilebind.Mobilebind
import java.io.File

/**
 * The Go core (gomobile AAR) runs in this process. Start is idempotent:
 * the activity and the foreground service both call it, whoever comes
 * first starts the core.
 */
object Core {
    data class Endpoint(val socket: String, val token: String)

    @Volatile private var endpoint: Endpoint? = null

    /** The app is on screen: it shows new messages itself, no notifications. */
    @Volatile var visible = false

    @Synchronized
    fun start(context: Context): Endpoint {
        endpoint?.let { return it }
        val app = context.applicationContext
        val dir = File(app.filesDir, "parley").apply { mkdirs() }
        // A unix socket path must stay under ~100 bytes: keep it short.
        val socket = File(app.filesDir, "parley.sock").absolutePath
        Shard.init(app)
        Mobilebind.setShard(Shard)
        val token = Mobilebind.start(dir.absolutePath, app.cacheDir.absolutePath, socket)
        Mobilebind.setListener(object : Listener {
            override fun onEvent(kind: Long, json: String?) {
                if ((kind == MESSAGE_NEW || kind == THREAD_MESSAGE || kind == MAIL_NEW) && json != null && !visible) {
                    Notifications.message(app, json)
                }
                if (kind == CALL && json != null) {
                    Notifications.call(app, json, visible)
                    // States CONNECTING (3) and ACTIVE (4): the microphone and
                    // camera stay usable while the app is in the background.
                    val state = runCatching { org.json.JSONObject(json).optInt("state") }.getOrDefault(0)
                    CoreService.inCall(app, state == 3 || state == 4)
                }
            }
        })
        return Endpoint(socket, token).also { endpoint = it }
    }

    // Event.Kind.KIND_MESSAGE_NEW in parley.proto.
    private const val MESSAGE_NEW = 4L

    // Event.Kind.KIND_THREAD_MESSAGE: comes here only for a followed thread or a mention.
    private const val THREAD_MESSAGE = 9L

    // Event.Kind.KIND_MAIL_NEW: a new message in an inbox.
    private const val MAIL_NEW = 20L

    // Event.Kind.KIND_CALL.
    private const val CALL = 10L
}
