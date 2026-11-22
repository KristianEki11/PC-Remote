import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../models/media_status.dart';
import 'api_service.dart';

class CompanionService {
  CompanionService._();

  static const _channel = MethodChannel('com.pcremote/companion');
  static const _prefKey = 'companion_enabled';

  static bool _initialized = false;
  static String _lastSignature = '';

  static Future<bool> isEnabled() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getBool(_prefKey) ?? true;
  }

  static Future<void> setEnabled(bool enabled) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(_prefKey, enabled);
    if (enabled) {
      await requestNotificationPermission();
      await startService();
    } else {
      await stopService();
    }
  }

  static Future<void> init() async {
    if (_initialized) return;
    _initialized = true;
    _channel.setMethodCallHandler(_handleNativeCall);
    if (kIsWeb || defaultTargetPlatform != TargetPlatform.android) return;
    if (await isEnabled()) {
      await requestNotificationPermission();
      await startService();
    }
  }

  static Future<dynamic> _handleNativeCall(MethodCall call) async {
    switch (call.method) {
      case 'mediaPrev':
        return ApiService.mediaPrev();
      case 'mediaPlayPause':
        return ApiService.mediaPlayPause();
      case 'mediaNext':
        return ApiService.mediaNext();
      default:
        return null;
    }
  }

  static Future<void> startService() async {
    if (kIsWeb || defaultTargetPlatform != TargetPlatform.android) return;
    await _safeInvoke('startService');
  }

  static Future<void> stopService() async {
    if (kIsWeb || defaultTargetPlatform != TargetPlatform.android) return;
    _lastSignature = '';
    await _safeInvoke('stopService');
  }

  static Future<void> requestNotificationPermission() async {
    if (kIsWeb || defaultTargetPlatform != TargetPlatform.android) return;
    await _safeInvoke('requestNotificationPermission');
  }

  static Future<bool> openTailscaleApp() async {
    if (kIsWeb || defaultTargetPlatform != TargetPlatform.android) return false;
    final res = await _safeInvoke<bool>('openTailscaleApp');
    return res ?? false;
  }

  static Future<bool> downloadTailscaleApk() async {
    if (kIsWeb || defaultTargetPlatform != TargetPlatform.android) return false;
    final res = await _safeInvoke<bool>('downloadTailscaleApk');
    return res ?? false;
  }

  static Future<void> updateStatus({
    required bool connected,
    required String routeLabel,
    MediaStatus? media,
  }) async {
    if (kIsWeb || defaultTargetPlatform != TargetPlatform.android) return;
    if (!await isEnabled()) return;

    final title = media?.title.trim() ?? '';
    final artist = media?.artist.trim() ?? '';
    final app = media?.appId?.trim() ?? '';
    final status = media?.status ?? 'Closed';
    final sig = '$connected|$routeLabel|$status|$title|$artist|$app';
    if (sig == _lastSignature) return;
    _lastSignature = sig;

    await _safeInvoke('updateStatus', {
      'connected': connected,
      'routeLabel': routeLabel,
      'mediaTitle': title,
      'mediaArtist': artist,
      'mediaApp': app,
      'mediaStatus': status,
    });
  }

  static Future<T?> _safeInvoke<T>(String method, [dynamic arguments]) async {
    try {
      return await _channel.invokeMethod<T>(method, arguments);
    } on MissingPluginException {
      return null;
    } on PlatformException {
      return null;
    } catch (_) {
      return null;
    }
  }
}
