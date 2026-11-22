package com.example.pc_remote

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.content.pm.ServiceInfo
import android.os.Build
import android.os.IBinder
import androidx.core.app.NotificationCompat

class CompanionService : Service() {

    companion object {
        const val CHANNEL_ID = "pc_remote_companion"
        const val NOTIFICATION_ID = 5001

        const val ACTION_START = "com.example.pc_remote.ACTION_START"
        const val ACTION_STOP = "com.example.pc_remote.ACTION_STOP"
        const val ACTION_UPDATE = "com.example.pc_remote.ACTION_UPDATE"
        const val ACTION_MEDIA_PREV = "com.example.pc_remote.ACTION_MEDIA_PREV"
        const val ACTION_MEDIA_PLAY_PAUSE = "com.example.pc_remote.ACTION_MEDIA_PLAY_PAUSE"
        const val ACTION_MEDIA_NEXT = "com.example.pc_remote.ACTION_MEDIA_NEXT"

        const val EXTRA_CONNECTED = "connected"
        const val EXTRA_ROUTE_LABEL = "routeLabel"
        const val EXTRA_MEDIA_TITLE = "mediaTitle"
        const val EXTRA_MEDIA_ARTIST = "mediaArtist"
        const val EXTRA_MEDIA_APP = "mediaApp"
        const val EXTRA_MEDIA_STATUS = "mediaStatus"
    }

    private var connected: Boolean = false
    private var routeLabel: String = "Wi-Fi Lokal"
    private var mediaTitle: String = ""
    private var mediaArtist: String = ""
    private var mediaApp: String = ""
    private var mediaStatus: String = "Closed"

    override fun onCreate() {
        super.onCreate()
        createNotificationChannel()
    }

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_STOP -> {
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
                    stopForeground(STOP_FOREGROUND_REMOVE)
                } else {
                    @Suppress("DEPRECATION")
                    stopForeground(true)
                }
                stopSelf()
                return START_NOT_STICKY
            }
            ACTION_MEDIA_PREV -> {
                MainActivity.sendMediaAction("mediaPrev")
                return START_STICKY
            }
            ACTION_MEDIA_PLAY_PAUSE -> {
                MainActivity.sendMediaAction("mediaPlayPause")
                return START_STICKY
            }
            ACTION_MEDIA_NEXT -> {
                MainActivity.sendMediaAction("mediaNext")
                return START_STICKY
            }
            ACTION_UPDATE -> {
                connected = intent.getBooleanExtra(EXTRA_CONNECTED, connected)
                routeLabel = intent.getStringExtra(EXTRA_ROUTE_LABEL) ?: routeLabel
                mediaTitle = intent.getStringExtra(EXTRA_MEDIA_TITLE) ?: ""
                mediaArtist = intent.getStringExtra(EXTRA_MEDIA_ARTIST) ?: ""
                mediaApp = intent.getStringExtra(EXTRA_MEDIA_APP) ?: ""
                mediaStatus = intent.getStringExtra(EXTRA_MEDIA_STATUS) ?: "Closed"
            }
        }

        val notification = buildNotification()
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            startForeground(
                NOTIFICATION_ID,
                notification,
                ServiceInfo.FOREGROUND_SERVICE_TYPE_DATA_SYNC
            )
        } else {
            startForeground(NOTIFICATION_ID, notification)
        }
        return START_STICKY
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                "PC Remote Companion 24/7",
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = "Status koneksi PC Remote dan kontrol media cepat"
                setShowBadge(false)
            }
            val manager = getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
            manager.createNotificationChannel(channel)
        }
    }

    private fun serviceActionPendingIntent(action: String, requestCode: Int): PendingIntent {
        val intent = Intent(this, CompanionService::class.java).apply {
            this.action = action
        }
        return PendingIntent.getService(
            this,
            requestCode,
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
    }

    private fun buildNotification(): Notification {
        val openAppIntent = Intent(this, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_SINGLE_TOP or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        val contentIntent = PendingIntent.getActivity(
            this,
            0,
            openAppIntent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )

        val statusText = if (connected) {
            "Terhubung • $routeLabel"
        } else {
            "Menunggu Laptop Aktif..."
        }

        val hasActiveMedia =
            (mediaStatus == "Playing" || mediaStatus == "Paused") && mediaTitle.isNotBlank()

        val builder = NotificationCompat.Builder(this, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_stat_pc_remote)
            .setContentIntent(contentIntent)
            .setOngoing(true)
            .setOnlyAlertOnce(true)
            .setSilent(true)
            .setPriority(NotificationCompat.PRIORITY_LOW)

        if (hasActiveMedia) {
            val subInfo = listOf(mediaArtist, mediaApp, statusText)
                .filter { it.isNotBlank() }
                .joinToString(" • ")
            builder.setContentTitle(mediaTitle)
            builder.setContentText(subInfo)

            builder.addAction(
                android.R.drawable.ic_media_previous,
                "Prev",
                serviceActionPendingIntent(ACTION_MEDIA_PREV, 101)
            )
            val isPlaying = mediaStatus == "Playing"
            builder.addAction(
                if (isPlaying) android.R.drawable.ic_media_pause else android.R.drawable.ic_media_play,
                if (isPlaying) "Pause" else "Play",
                serviceActionPendingIntent(ACTION_MEDIA_PLAY_PAUSE, 102)
            )
            builder.addAction(
                android.R.drawable.ic_media_next,
                "Next",
                serviceActionPendingIntent(ACTION_MEDIA_NEXT, 103)
            )
        } else {
            builder.setContentTitle("PC Remote")
            builder.setContentText(statusText)
        }

        return builder.build()
    }
}
