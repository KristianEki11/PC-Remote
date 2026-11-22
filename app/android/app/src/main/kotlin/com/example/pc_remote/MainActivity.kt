package com.example.pc_remote

import android.Manifest
import android.content.Intent
import android.content.pm.PackageManager
import android.net.Uri
import android.os.Build
import android.os.Handler
import android.os.Looper
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel

class MainActivity : FlutterActivity() {

    companion object {
        private const val CHANNEL = "com.pcremote/companion"
        private const val REQ_POST_NOTIFICATIONS = 5002
        private var methodChannel: MethodChannel? = null
        private val mainHandler = Handler(Looper.getMainLooper())

        fun sendMediaAction(action: String) {
            mainHandler.post {
                methodChannel?.invokeMethod(action, null)
            }
        }
    }

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        val channel = MethodChannel(flutterEngine.dartExecutor.binaryMessenger, CHANNEL)
        methodChannel = channel

        channel.setMethodCallHandler { call, result ->
            when (call.method) {
                "startService" -> {
                    val intent = Intent(this, CompanionService::class.java).apply {
                        action = CompanionService.ACTION_START
                    }
                    startCompanionService(intent)
                    result.success(true)
                }
                "stopService" -> {
                    val intent = Intent(this, CompanionService::class.java).apply {
                        action = CompanionService.ACTION_STOP
                    }
                    startService(intent)
                    result.success(true)
                }
                "updateStatus" -> {
                    val args = call.arguments as? Map<*, *>
                    val intent = Intent(this, CompanionService::class.java).apply {
                        action = CompanionService.ACTION_UPDATE
                        putExtra(CompanionService.EXTRA_CONNECTED, args?.get("connected") as? Boolean ?: false)
                        putExtra(CompanionService.EXTRA_ROUTE_LABEL, args?.get("routeLabel") as? String ?: "Wi-Fi Lokal")
                        putExtra(CompanionService.EXTRA_MEDIA_TITLE, args?.get("mediaTitle") as? String ?: "")
                        putExtra(CompanionService.EXTRA_MEDIA_ARTIST, args?.get("mediaArtist") as? String ?: "")
                        putExtra(CompanionService.EXTRA_MEDIA_APP, args?.get("mediaApp") as? String ?: "")
                        putExtra(CompanionService.EXTRA_MEDIA_STATUS, args?.get("mediaStatus") as? String ?: "Closed")
                    }
                    startCompanionService(intent)
                    result.success(true)
                }
                "requestNotificationPermission" -> {
                    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                        if (ContextCompat.checkSelfPermission(
                                this,
                                Manifest.permission.POST_NOTIFICATIONS
                            ) != PackageManager.PERMISSION_GRANTED
                        ) {
                            ActivityCompat.requestPermissions(
                                this,
                                arrayOf(Manifest.permission.POST_NOTIFICATIONS),
                                REQ_POST_NOTIFICATIONS
                            )
                        }
                    }
                    result.success(true)
                }
                "openTailscaleApp" -> {
                    result.success(openTailscaleOrPlayStore())
                }
                "downloadTailscaleApk" -> {
                    result.success(openUrl("https://tailscale.com/download/android"))
                }
                else -> result.notImplemented()
            }
        }
    }

    private fun startCompanionService(intent: Intent) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            startForegroundService(intent)
        } else {
            startService(intent)
        }
    }

    private fun openTailscaleOrPlayStore(): Boolean {
        val launchIntent = packageManager.getLaunchIntentForPackage("com.tailscale.ipn")
        if (launchIntent != null) {
            launchIntent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
            startActivity(launchIntent)
            return true
        }
        return try {
            val marketIntent = Intent(
                Intent.ACTION_VIEW,
                Uri.parse("market://details?id=com.tailscale.ipn")
            ).apply { addFlags(Intent.FLAG_ACTIVITY_NEW_TASK) }
            startActivity(marketIntent)
            true
        } catch (_: Exception) {
            openUrl("https://play.google.com/store/apps/details?id=com.tailscale.ipn")
        }
    }

    private fun openUrl(url: String): Boolean {
        return try {
            val webIntent = Intent(
                Intent.ACTION_VIEW,
                Uri.parse(url)
            ).apply { addFlags(Intent.FLAG_ACTIVITY_NEW_TASK) }
            startActivity(webIntent)
            true
        } catch (_: Exception) {
            false
        }
    }

    override fun onDestroy() {
        methodChannel = null
        super.onDestroy()
    }
}
