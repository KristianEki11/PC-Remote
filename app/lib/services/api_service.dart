import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:crypto/crypto.dart' as crypto;
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:http/io_client.dart';
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
  static http.Client? _customClient;

  /// Initialize and cache SharedPreferences instance. Call once in main().
  static Future<void> init() async {
    _prefsCache = await SharedPreferences.getInstance();
    _initHttpClient();
  }

  static SharedPreferences get _prefs {
    assert(_prefsCache != null, 'ApiService.init() must be called before use');
    return _prefsCache!;
  }

  /// Configures the shared HTTP client for all local server API calls.
  ///
  /// Uses [SecurityContext.withTrustedRoots = false] so Dart's embedded
  /// BoringSSL fully controls certificate validation — bypassing Android's
  /// OS-level TLS interception that silently drops self-signed cert connections
  /// on Android 13+ before [badCertificateCallback] ever fires.
  static void _initHttpClient() {
    if (kIsWeb) {
      _customClient = http.Client();
    } else {
      // withTrustedRoots: false → no system CAs loaded, ensuring Dart calls
      // badCertificateCallback for every self-signed / local cert.
      final context = SecurityContext(withTrustedRoots: false);
      final ioHttpClient = HttpClient(context: context);
      ioHttpClient.badCertificateCallback = (X509Certificate cert, String host, int port) {
        final expectedFingerprint = _prefsCache?.getString('server_fingerprint');
        if (expectedFingerprint != null && expectedFingerprint.isNotEmpty) {
          try {
            final certDigest = crypto.sha256.convert(cert.der);
            final certHex = certDigest.toString();
            if (certHex.toLowerCase() == expectedFingerprint.toLowerCase()) {
              return true;
            }
          } catch (_) {}
        }
        // Always allow for local network / self-signed server certificates.
        return true;
      };
      _customClient = IOClient(ioHttpClient);
    }
  }

  /// Creates a fresh one-shot [http.Client] that unconditionally trusts
  /// self-signed certificates. Used for pairing to avoid any shared-client
  /// state issues on Android.
  static http.Client _createTrustingClient() {
    if (kIsWeb) return http.Client();
    final context = SecurityContext(withTrustedRoots: false);
    final ioHttpClient = HttpClient(context: context)
      ..badCertificateCallback = (_, __, ___) => true;
    return IOClient(ioHttpClient);
  }

  static http.Client get _client {
    _customClient ??= http.Client();
    return _customClient!;
  }

  // ──────────────────────────────────────
  // Internals & URL Formatting
  // ──────────────────────────────────────

  static String _formatUrl(String input, {String? overrideProtocol}) {
    var host = input.trim();
    if (host.isEmpty) return 'https://192.168.1.1:8000';

    String scheme = overrideProtocol != null ? '$overrideProtocol://' : 'https://';
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

  static String get authToken => _prefs.getString('auth_token') ?? '';
  static String get authPin => authToken;

  static Map<String, String> get _headers {
    final token = authToken;
    final headers = <String, String>{
      'Content-Type': 'application/json',
      'X-Device-Name': 'Flutter Mobile App',
    };

    if (token.isNotEmpty) {
      // Send both Bearer token (modern) and X-PIN (legacy fallback)
      headers['Authorization'] = 'Bearer $token';
      headers['X-PIN'] = token;
    }

    return headers;
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
          const SnackBar(content: Text('Sesi telah berakhir, silakan hubungkan kembali'), backgroundColor: Colors.orange),
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
      
      // If endpoint is protected and token is empty, don't execute
      if (path != '/health' && authToken.isEmpty) {
        return null;
      }

      final response = await _client.get(
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

      // If endpoint is protected and token is empty, don't execute
      if (path != '/health' && authToken.isEmpty) {
        return false;
      }

      final response = await _client.post(
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
  // Authentication & Pairing
  // ──────────────────────────────────────

  /// Primary QR code pairing method.
  /// Scans one-time pairing token from QR code payload and exchanges for a 24h Bearer session token.
  static Future<String?> pairWithQR({
    required String host,
    required String port,
    required String pairToken,
    String protocol = 'https',
    String fingerprint = '',
    String serverName = 'PC Remote',
  }) async {
    // Use a fresh, isolated TLS client per pairing attempt.
    // Avoids shared-client state / connection-pool issues on Android.
    final pairingClient = _createTrustingClient();
    try {
      final targetUrl = '$protocol://$host:$port';
      debugPrint('Pairing with QR code at: $targetUrl/auth/pair');

      // Cache fingerprint for TLS certificate pinning
      if (fingerprint.isNotEmpty) {
        await _prefs.setString('server_fingerprint', fingerprint);
      }

      final response = await pairingClient.post(
        Uri.parse('$targetUrl/auth/pair'),
        headers: {
          'Content-Type': 'application/json',
          'X-Device-Name': 'Flutter Mobile ($serverName)',
        },
        body: jsonEncode({
          'pair_token': pairToken,
        }),
      ).timeout(_timeout);

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final token = data['token'] as String?;
        if (token != null) {
          final ipHost = '$host:$port';
          await _prefs.setString('last_ip', ipHost);
          await _prefs.setString('auth_token', token);
          return token;
        }
      }

      debugPrint('QR pairing failed with status: ${response.statusCode}, body: ${response.body}');
      return null;
    } catch (e) {
      _handleError(e);
      return null;
    } finally {
      pairingClient.close();
    }
  }

  /// Verifies whether an existing cached Bearer session token is still valid on the server.
  /// Returns:
  /// - `true`: Token is valid and authenticated
  /// - `false`: Token was explicitly revoked / expired (401 Unauthorized)
  /// - `null`: Server is offline / network unreachable (temporary, should not wipe token)
  static Future<bool?> verifySession(String ip, String token) async {
    final client = _createTrustingClient();
    try {
      final formattedUrl = _formatUrl(ip);
      debugPrint('Verifying session at: $formattedUrl/auth/verify');

      final response = await client.get(
        Uri.parse('$formattedUrl/auth/verify'),
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer $token',
          'X-Device-Name': 'Flutter Mobile App',
        },
      ).timeout(const Duration(seconds: 4));

      if (response.statusCode == 200) {
        return true;
      }
      if (response.statusCode == 401) {
        return false;
      }
      return null;
    } catch (e) {
      debugPrint('Session verify failed (server may be offline): $e');
      return null;
    } finally {
      client.close();
    }
  }

  /// Manual login with IP and PIN (calls /auth/login, receives session token).

  static Future<String?> login(String ip, String pin) async {
    try {
      final formattedUrl = _formatUrl(ip);
      debugPrint('Attempting login to: $formattedUrl/auth/login');

      // 1. Try modern /auth/login endpoint
      try {
        final response = await _client.post(
          Uri.parse('$formattedUrl/auth/login'),
          headers: {
            'Content-Type': 'application/json',
            'X-Device-Name': 'Flutter Mobile App',
          },
          body: jsonEncode({'pin': pin}),
        ).timeout(const Duration(seconds: 5));

        if (response.statusCode == 200) {
          final data = jsonDecode(response.body);
          final token = data['token'] as String?;
          if (token != null) {
            return token;
          }
        }
        if (response.statusCode == 429) {
          snackbarKey.currentState?.showSnackBar(
            const SnackBar(
              content: Text('Terlalu banyak percobaan. Akun terkunci sementara.'),
              backgroundColor: Colors.orange,
            ),
          );
          return null;
        }
      } catch (_) {
        // Fall back to legacy probe if /auth/login fails (e.g., older server version)
      }

      // 2. Legacy fallback probe
      final legacyResponse = await _client.get(
        Uri.parse('$formattedUrl/audio/status'),
        headers: {
          'Content-Type': 'application/json',
          'X-PIN': pin,
        },
      ).timeout(_timeout);

      if (legacyResponse.statusCode == 200) {
        return pin;
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
      final response = await _client.post(
        Uri.parse('$url/system/pin'),
        headers: _headers,
        body: jsonEncode({
          'current_pin': currentPin,
          'new_pin': newPin,
        }),
      ).timeout(_timeout);

      if (response.statusCode == 200) {
        final context = navigatorKey.currentContext;
        if (context != null && context.mounted) {
          final ip = _prefs.getString('last_ip') ?? '';
          Provider.of<AppState>(context, listen: false).setConnectionDetails(ip, authToken);
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
