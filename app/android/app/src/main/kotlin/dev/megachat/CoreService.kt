package dev.megachat

import android.Manifest
import android.app.Notification
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
        val n = Notification.Builder(this, Notifications.SERVICE)
            .setContentTitle("MegaChat")
            .setContentText("На связи с мессенджерами")
            .setSmallIcon(R.drawable.ic_stat_chat)
            .setContentIntent(Notifications.openApp(this))
            .setOngoing(true)
            .build()
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

    companion object {
        private const val ID = 1
        private const val EXTRA_CALL = "call"

        @Volatile private var call = false

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
