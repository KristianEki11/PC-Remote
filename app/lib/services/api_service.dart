import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import 'package:provider/provider.dart';
import '../models/app_state.dart';
import '../utils/globals.dart';
import '../screens/login_screen.dart';
import '../utils/volume_helpers.dart';
import '../models/media_state.dart';

class ApiService {
  // ──────────────────────────────────────
  // SharedPreferences Singleton Cache
  // ──────────────────────────────────────

  static SharedPreferences? _prefsCache;

  /// Initialize and cache SharedPreferences instance. Call once in main().
  static Future<void> init() async {
    _prefsCache = await SharedPreferences.getInstance();
  }

  static SharedPreferences get _prefs {
    assert(_prefsCache != null, 'ApiService.init() must be called before use');
    return _prefsCache!;
  }

  // ──────────────────────────────────────
  // Internals
  // ──────────────────────────────────────

  static String _formatUrl(String input) {
    var host = input.trim();
    if (host.isEmpty) return 'http://192.168.1.1:8000';

    String scheme = 'http://';
    if (host.startsWith('http://')) {
      scheme = 'http://';
      host = host.substring(7);
    } else if (host.startsWith('https://')) {
      scheme = 'https://';
      host = host.substring(8);
    }

    bool hasPort = false;
    if (host.contains(']')) {
      // IPv6 address
      int closingBracket = host.indexOf(']');
      if (host.substring(closingBracket).contains(':')) {
        hasPort = true;
      }
    } else {
      // IPv4 or hostname
      if (host.contains(':')) {
        hasPort = true;
      }
    }

    if (!hasPort) {
      host = '$host:8000';
    }

    return '$scheme$host';
  }

  static String get _baseUrl {
    final ip = _prefs.getString('last_ip') ?? '192.168.1.1:8000';
    return _formatUrl(ip);
  }

  static String get authPin => _prefs.getString('auth_token') ?? '';

  static Map<String, String> get _headers {
    return {
      'Content-Type': 'application/json',
      'X-PIN': authPin,
    };
  }

  static void _check401(http.Response response) async {
    if (response.statusCode == 401) {
      await _prefs.remove('auth_token');

      final context = navigatorKey.currentContext;
      if (context != null && context.mounted) {
        Provider.of<AppState>(context, listen: false).clear();
        try {
          Provider.of<MediaState>(context, listen: false).stopPolling();
        } catch (_) {}
        Navigator.of(context).pushAndRemoveUntil(
          MaterialPageRoute(builder: (context) => const LoginScreen()),
          (route) => false,
        );
        snackbarKey.currentState?.showSnackBar(
          const SnackBar(content: Text('Sesi telah berakhir, silakan login kembali'), backgroundColor: Colors.orange),
        );
      }
    }
  }

  static void _handleError(Object e, {bool showSnackBar = true}) {
    debugPrint('API Error: $e');
    if (!showSnackBar) return;
    final msg = e is TimeoutException
        ? 'Waktu koneksi habis (Timeout)'
        : 'Network Error: $e';

    snackbarKey.currentState?.showSnackBar(
      SnackBar(content: Text(msg), backgroundColor: Colors.red),
    );
  }

  // ──────────────────────────────────────
  // Generic HTTP helpers (DRY)
  // ──────────────────────────────────────

  static const _timeout = Duration(seconds: 10);

  /// Generic GET request. Returns decoded JSON body or null.
  static Future<Map<String, dynamic>?> _get(String path, {Duration? timeout, bool showSnackBar = true}) async {
    try {
      final url = _baseUrl;
      final headers = _headers;
      
      // If endpoint is protected and token/PIN is empty, don't execute
      if (path != '/health' && (headers['X-PIN'] ?? '').isEmpty) {
        return null;
      }

      final response = await http.get(
        Uri.parse('$url$path'),
        headers: headers,
      ).timeout(timeout ?? _timeout);

      _check401(response);
      if (response.statusCode == 200) {
        return jsonDecode(response.body);
      }
      return null;
    } catch (e) {
      _handleError(e, showSnackBar: showSnackBar);
      return null;
    }
  }

  /// Generic POST request. Returns true if status 200.
  static Future<bool> _post(String path, {Map<String, dynamic>? body, bool showSnackBar = true}) async {
    try {
      final url = _baseUrl;
      final headers = _headers;

      // If endpoint is protected and token/PIN is empty, don't execute
      if (path != '/health' && (headers['X-PIN'] ?? '').isEmpty) {
        return false;
      }

      final response = await http.post(
        Uri.parse('$url$path'),
        headers: headers,
        body: body != null ? jsonEncode(body) : null,
      ).timeout(_timeout);

      _check401(response);
      return response.statusCode == 200;
    } catch (e) {
      _handleError(e, showSnackBar: showSnackBar);
      return false;
    }
  }

  // ──────────────────────────────────────
  // Auth
  // ──────────────────────────────────────

  static Future<String?> login(String ip, String pin) async {
    try {
      final formattedUrl = _formatUrl(ip);
      debugPrint('Attempting login verification to: $formattedUrl/audio/status');
      final response = await http.get(
        Uri.parse('$formattedUrl/audio/status'),
        headers: {
          'Content-Type': 'application/json',
          'X-PIN': pin,
        },
      ).timeout(_timeout);

      if (response.statusCode == 200) {
        return pin; // Return the pin itself to be saved in SharedPreferences as the token
      }
      if (response.statusCode == 401) {
        return null;
      }
      if (response.statusCode == 429) {
        snackbarKey.currentState?.showSnackBar(
          const SnackBar(
            content: Text('Terlalu banyak percobaan. Coba lagi dalam 1 menit.'),
            backgroundColor: Colors.orange,
          ),
        );
      }
      return null;
    } catch (e) {
      _handleError(e);
      return null;
    }
  }

  static Future<String?> changePIN(String currentPin, String newPin) async {
    try {
      final url = _baseUrl;
      final response = await http.post(
        Uri.parse('$url/system/pin'),
        headers: {
          'Content-Type': 'application/json',
          'X-PIN': currentPin,
        },
        body: jsonEncode({
          'current_pin': currentPin,
          'new_pin': newPin,
        }),
      ).timeout(_timeout);

      if (response.statusCode == 200) {
        // Save the new PIN so future requests succeed
        await _prefs.setString('auth_token', newPin);
        
        final context = navigatorKey.currentContext;
        if (context != null && context.mounted) {
          final ip = _prefs.getString('last_ip') ?? '';
          Provider.of<AppState>(context, listen: false).setConnectionDetails(ip, newPin);
        }
        return null; // success
      }
      
      final body = jsonDecode(response.body);
      return body['error'] ?? 'Gagal mengubah PIN';
    } catch (e) {
      _handleError(e);
      return 'Koneksi error: $e';
    }
  }

  // ──────────────────────────────────────
  // Health Check (lightweight ping)
  // ──────────────────────────────────────

  static Future<Map<String, dynamic>?> healthCheck() =>
      _get('/health', timeout: const Duration(seconds: 5), showSnackBar: false);

  static Future<String?> getLatestGitHubRelease() async {
    try {
      final response = await http.get(
        Uri.parse('https://api.github.com/repos/KristianEki11/PC-Remote/releases/latest'),
      ).timeout(const Duration(seconds: 5));
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return data['tag_name'] as String?;
      }
    } catch (e) {
      debugPrint('Gagal mengambil rilis GitHub terbaru: $e');
    }
    return null;
  }

  // ──────────────────────────────────────
  // Audio
  // ──────────────────────────────────────

  static Future<Map<String, dynamic>?> getVolume() =>
      _get('/audio/status', timeout: const Duration(seconds: 15));

  static Future<bool> setVolume(double level) =>
      _post('/audio/volume', body: {'level': level});

  static Future<bool> setMute(bool muted) =>
      _post('/audio/mute', body: {'muted': muted});

  static Future<List<dynamic>?> getAudioDevices() async {
    final data = await _get('/audio/devices');
    return data?['devices'];
  }

  static Future<bool> setDeviceVolume(String deviceId, double level) =>
      _post('/audio/device/volume', body: {'device_id': deviceId, 'level': VolumeHelpers.toScalar(level)});

  static Future<bool> toggleDeviceMute(String deviceId, bool newMute) =>
      _post('/audio/device/mute', body: {'device_id': deviceId, 'mute': newMute});


  // ──────────────────────────────────────
  // Media
  // ──────────────────────────────────────

  static Future<bool> mediaPlayPause() => _post('/media/play');
  static Future<bool> mediaNext()      => _post('/media/next');
  static Future<bool> mediaPrev()      => _post('/media/prev');
  static Future<Map<String, dynamic>?> getMediaStatus() =>
      _get('/media/status', timeout: const Duration(seconds: 5), showSnackBar: false);

  // ──────────────────────────────────────
  // Browser
  // ──────────────────────────────────────

  static Future<bool> openBrowser(String browserUrl) =>
      _post('/browser/open', body: {'url': browserUrl});

  // ──────────────────────────────────────
  // System
  // ──────────────────────────────────────

  static Future<bool> lockPc()      => _post('/system/lock');
  static Future<bool> sleepPc()     => _post('/system/sleep');
  static Future<bool> restartPc()   => _post('/system/restart');
  static Future<bool> displayOff()  => _post('/system/display/off');
  static Future<bool> shutdownPc({int delaySeconds = 0}) => _post('/system/shutdown', body: {'delay_seconds': delaySeconds});
}
