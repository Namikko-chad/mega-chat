package dev.megachat

import android.Manifest
import android.app.Notification
import android.app.NotificationManager
import android.app.Service
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.content.pm.ServiceInfo
import android.os.Build
import android.os.IBinder

/**
 * Keeps the process (and so the core with its messenger sessions) alive
 * while the app is in the background, so new messages still arrive.
 */
class CoreService : Service() {
    override fun onBind(intent: Intent?): IBinder? = null

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        Core.start(this)
        Notifications.channels(this)
        running = true
        val n = notification(this)
        if (intent?.hasExtra(EXTRA_CALL) == true) call = intent.getBooleanExtra(EXTRA_CALL, false)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.UPSIDE_DOWN_CAKE) {
            var type = ServiceInfo.FOREGROUND_SERVICE_TYPE_REMOTE_MESSAGING
            if (call) {
                // Only types whose permission was granted; the others throw.
                if (checkSelfPermission(Manifest.permission.RECORD_AUDIO) == PackageManager.PERMISSION_GRANTED) {
                    type = type or ServiceInfo.FOREGROUND_SERVICE_TYPE_MICROPHONE
                }
                if (checkSelfPermission(Manifest.permission.CAMERA) == PackageManager.PERMISSION_GRANTED) {
                    type = type or ServiceInfo.FOREGROUND_SERVICE_TYPE_CAMERA
                }
            }
            try {
                startForeground(ID, n, type)
            } catch (e: RuntimeException) {
                // Not allowed from the background: keep messaging only.
                startForeground(ID, n, ServiceInfo.FOREGROUND_SERVICE_TYPE_REMOTE_MESSAGING)
            }
        } else {
            startForeground(ID, n)
        }
        return START_STICKY
    }

    override fun onDestroy() {
        running = false
        super.onDestroy()
    }

    companion object {
        private const val ID = 1
        private const val EXTRA_CALL = "call"

        @Volatile private var call = false
        @Volatile private var running = false

        private fun notification(context: Context): Notification = Notification.Builder(context, Notifications.SERVICE)
            .setContentTitle("MegaChat")
            .setContentText(Notifications.tr(context, "Connected to your messengers", "На связи с мессенджерами"))
            .setSmallIcon(R.drawable.ic_stat_chat)
            .setContentIntent(Notifications.openApp(context))
            .setOngoing(true)
            .build()

        /** The language changed: the ongoing notification shows the new text. */
        fun refresh(context: Context) {
            if (!running) return
            context.getSystemService(NotificationManager::class.java).notify(ID, notification(context))
        }

        /** A call started or ended: the service type follows it. */
        fun inCall(context: Context, on: Boolean) {
            if (on == call) return
            call = on
            runCatching {
                context.startForegroundService(Intent(context, CoreService::class.java).putExtra(EXTRA_CALL, on))
            }
        }

        fun start(context: Context) {
            context.startForegroundService(Intent(context, CoreService::class.java))
        }
    }
}
