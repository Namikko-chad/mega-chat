package dev.megachat

import android.content.ComponentName
import android.content.Context
import android.content.Intent
import android.content.ServiceConnection
import android.os.DeadObjectException
import android.os.IBinder
import android.os.SystemClock
import dev.cloudsync.connector.ICloudSyncConnector
import dev.cloudsync.connector.IConnectorWatcher
import dev.megachat.mobilebind.CloudSyncBinder
import dev.megachat.mobilebind.CloudSyncWatcher

/**
 * CloudSync's bound service (CONNECTOR.md § 3.2) for the Go core, which
 * syncs the vault through it. The core calls from its own threads, never the
 * main one; requests and replies are serialized connector.proto messages
 * the core makes and reads. CloudSync knows MegaChat from Binder (package and
 * signing key): there is no token here.
 */
object CloudSync : CloudSyncBinder {
    const val PACKAGE = "dev.cloudsync.cloudsync"
    private const val ACTION = "dev.cloudsync.connector.BIND"

    private lateinit var app: Context
    private val lock = Object()
    private var api: ICloudSyncConnector? = null
    private var bound = false

    fun init(context: Context) {
        app = context.applicationContext
    }

    private val conn = object : ServiceConnection {
        override fun onServiceConnected(name: ComponentName?, service: IBinder?) {
            synchronized(lock) {
                api = ICloudSyncConnector.Stub.asInterface(service)
                lock.notifyAll()
            }
        }

        // CloudSync's process died: Android binds again when it comes back.
        override fun onServiceDisconnected(name: ComponentName?) {
            synchronized(lock) { api = null }
        }

        override fun onBindingDied(name: ComponentName?) = drop()

        override fun onNullBinding(name: ComponentName?) = drop()
    }

    private fun drop() {
        synchronized(lock) {
            api = null
            if (bound) runCatching { app.unbindService(conn) }
            bound = false
            lock.notifyAll()
        }
    }

    /** The service, bound on first use; throws when CloudSync is missing. */
    private fun connect(): ICloudSyncConnector {
        synchronized(lock) {
            api?.let { return it }
            if (!bound) {
                val i = Intent(ACTION).setPackage(PACKAGE)
                bound = runCatching { app.bindService(i, conn, Context.BIND_AUTO_CREATE) }.getOrDefault(false)
                if (!bound) {
                    runCatching { app.unbindService(conn) }
                    throw IllegalStateException("CloudSync is not installed")
                }
            }
            val deadline = SystemClock.elapsedRealtime() + 5_000
            while (api == null) {
                val left = deadline - SystemClock.elapsedRealtime()
                if (left <= 0 || !bound) throw IllegalStateException("CloudSync does not answer")
                lock.wait(left)
            }
            return api!!
        }
    }

    override fun call(method: String, request: ByteArray): ByteArray {
        try {
            return connect().call(method, request)
        } catch (e: DeadObjectException) {
            synchronized(lock) { api = null }
            throw IllegalStateException("CloudSync stopped", e)
        }
    }

    override fun watch(w: CloudSyncWatcher): Long {
        val stub = object : IConnectorWatcher.Stub() {
            override fun onChange(change: ByteArray) = w.onChange(change)
            override fun onEnd(reply: ByteArray) = w.onEnd(reply)
        }
        try {
            val api = connect()
            // A dying CloudSync never calls onEnd: end the watch for it.
            val death = IBinder.DeathRecipient { w.onEnd(ByteArray(0)) }
            runCatching { api.asBinder().linkToDeath(death, 0) }
            return api.watch(stub)
        } catch (e: DeadObjectException) {
            synchronized(lock) { api = null }
            throw IllegalStateException("CloudSync stopped", e)
        }
    }

    override fun unwatch(handle: Long) {
        runCatching { synchronized(lock) { api }?.unwatch(handle) }
    }
}
