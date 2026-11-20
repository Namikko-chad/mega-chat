package dev.megachat

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.media.AudioAttributes
import android.media.RingtoneManager
import org.json.JSONObject

/** Notification channels and the "new message" notification. */
object Notifications {
    const val SERVICE = "service"
    private const val MESSAGES = "messages"
    private const val CALLS = "calls"
    private const val CALL_ID = 2

    fun channels(context: Context) {
        val nm = context.getSystemService(NotificationManager::class.java)
        nm.createNotificationChannel(
            NotificationChannel(SERVICE, "Фоновая работа", NotificationManager.IMPORTANCE_MIN).apply {
                description = "MegaChat держит соединения с мессенджерами"
            },
        )
        nm.createNotificationChannel(
            NotificationChannel(MESSAGES, "Сообщения", NotificationManager.IMPORTANCE_HIGH).apply {
                description = "Новые входящие сообщения"
            },
        )
        nm.createNotificationChannel(
            NotificationChannel(CALLS, "Звонки", NotificationManager.IMPORTANCE_HIGH).apply {
                description = "Входящие звонки"
                setSound(
                    RingtoneManager.getDefaultUri(RingtoneManager.TYPE_RINGTONE),
                    AudioAttributes.Builder().setUsage(AudioAttributes.USAGE_NOTIFICATION_RINGTONE)
                        .setContentType(AudioAttributes.CONTENT_TYPE_SONIFICATION).build(),
                )
                enableVibration(true)
            },
        )
    }

    /**
     * An incoming call rings with a full-screen notification while the app
     * is not on screen (on screen the app rings itself); anything else about
     * the call takes it away.
     */
    fun call(context: Context, json: String, visible: Boolean) {
        val c = runCatching { JSONObject(json) }.getOrNull() ?: return
        val nm = context.getSystemService(NotificationManager::class.java)
        val ringing = c.optInt("state") == CALL_RINGING && !c.optBoolean("outgoing")
        if (!ringing || visible) {
            nm.cancel(CALL_ID)
            return
        }
        channels(context)
        val video = c.optBoolean("video")
        val n = Notification.Builder(context, CALLS)
            .setContentTitle(c.optString("peer").ifEmpty { "MegaChat" })
            .setContentText(if (video) "Входящий видеозвонок" else "Входящий звонок")
            .setSmallIcon(R.drawable.ic_stat_chat)
            .setCategory(Notification.CATEGORY_CALL)
            .setContentIntent(openApp(context))
            .setFullScreenIntent(openApp(context), true)
            .setOngoing(true)
            .setAutoCancel(true)
            .setTimeoutAfter(90_000)
            .build()
        nm.notify(CALL_ID, n)
    }

    // CallState.CALL_STATE_RINGING in megachat.proto.
    private const val CALL_RINGING = 2

    fun openApp(context: Context): PendingIntent = PendingIntent.getActivity(
        context, 0,
        Intent(context, MainActivity::class.java).addFlags(Intent.FLAG_ACTIVITY_SINGLE_TOP),
        PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT,
    )

    private const val PREFS = "notifications"

    /** What message notifications show: "all", "mentions" or "off" (set by the app). */
    fun setMode(context: Context, mode: String) {
        context.getSharedPreferences(PREFS, Context.MODE_PRIVATE).edit().putString("mode", mode).apply()
    }

    private fun mode(context: Context) =
        context.getSharedPreferences(PREFS, Context.MODE_PRIVATE).getString("mode", "all") ?: "all"

    /** One notification per chat: a newer message replaces the older one. */
    fun message(context: Context, json: String) {
        val m = runCatching { JSONObject(json) }.getOrNull() ?: return
        val mention = m.optBoolean("mention")
        when (mode(context)) {
            "off" -> return
            // Mail is addressed to the user: it counts as a mention.
            "mentions" -> if (!mention && !m.optString("chat_id").startsWith("mail:")) return
        }
        val chat = m.optString("chat_title")
        val sender = m.optString("sender")
        val text = m.optString("text")
        val body = if (sender.isEmpty() || sender == chat) text else "$sender: $text"
        channels(context)
        val n = Notification.Builder(context, MESSAGES)
            .setContentTitle(chat.ifEmpty { sender })
            .setContentText(body)
            .setStyle(Notification.BigTextStyle().bigText(body))
            .setSmallIcon(R.drawable.ic_stat_chat)
            .setContentIntent(openApp(context))
            .setAutoCancel(true)
            .setCategory(Notification.CATEGORY_MESSAGE)
            .apply { if (mention) setSubText("Упоминание") }
            .build()
        val id = (m.optString("account_id") + "/" + m.optString("chat_id")).hashCode()
        context.getSystemService(NotificationManager::class.java).notify(id, n)
    }
}
