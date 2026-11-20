package dev.megachat

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.view.WindowManager
import io.flutter.embedding.android.FlutterFragmentActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel
import java.util.concurrent.Executors

// A FragmentActivity: BiometricPrompt needs one.
class MainActivity : FlutterFragmentActivity() {
    private val worker = Executors.newSingleThreadExecutor()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        // Android 13+: message notifications need the user's consent.
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU &&
            checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED
        ) {
            requestPermissions(arrayOf(Manifest.permission.POST_NOTIFICATIONS), 1)
        }
    }

    override fun onResume() {
        super.onResume()
        Core.visible = true
    }

    override fun onPause() {
        Core.visible = false
        super.onPause()
    }

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, "dev.megachat/biometric").setMethodCallHandler(Biometric(this))
        // The video player keeps the screen on while it plays.
        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, "dev.megachat/wakelock").setMethodCallHandler { call, result ->
            if (call.method != "toggle") return@setMethodCallHandler result.notImplemented()
            if (call.arguments == true) {
                window.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
            } else {
                window.clearFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
            }
            result.success(null)
        }
        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, "dev.megachat/core").setMethodCallHandler { call, result ->
            when (call.method) {
                // Starts the core (if needed) and returns where the UI connects.
                "start" -> worker.execute {
                    try {
                        val e = Core.start(this)
                        runCatching { CoreService.start(this) }
                        runOnUiThread { result.success(mapOf("socket" to e.socket, "token" to e.token)) }
                    } catch (t: Throwable) {
                        runOnUiThread { result.error("core", t.message, null) }
                    }
                }
                // "all", "mentions" or "off": what message notifications show.
                "notifications" -> {
                    Notifications.setMode(this, call.arguments as? String ?: "all")
                    result.success(null)
                }
                else -> result.notImplemented()
            }
        }
    }
}
