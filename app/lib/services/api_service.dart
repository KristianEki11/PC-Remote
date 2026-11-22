import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:math';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
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
  // â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
  // SharedPreferences Singleton Cache
  // â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

  static const _secureStorage = FlutterSecureStorage();
  static String _token = '';
  static Future<void> saveToken(String value) async {
    await _secureStorage.write(key: 'auth_token', value: value);
    _token = value;
    await _prefs.remove('auth_token');
  }

  static Future<void> clearToken() async {
    stopEvents();
    await _secureStorage.delete(key: 'auth_token');
    _token = '';
    await _prefs.remove('auth_token');
  }

  static SharedPreferences? _prefsCache;
  static http.Client? _customClient;
  static String? _activeOverrideUrl;
  static DateTime? _lastEndpointSyncAt;
  static Future<bool>? _syncInFlight;

  /// Initialize and cache SharedPreferences instance. Call once in main().
  static Future<void> init() async {
    _preferRemote = false;
    _activeOverrideUrl = null;
    _lastEndpointSyncAt = null;
    _syncInFlight = null;
    _prefsCache = await SharedPreferences.getInstance();
    _token = await _secureStorage.read(key: 'auth_token') ?? '';
    final legacy = _prefs.getString('auth_token');
    if (_token.isEmpty && legacy != null) await saveToken(legacy);
    if (legacy != null) await _prefs.remove('auth_token');
    _customClient?.close();
    _initHttpClient();
  }

  static SharedPreferences get _prefs {
    assert(_prefsCache != null, 'ApiService.init() must be called before use');
    return _prefsCache!;
  }

  /// Configures the shared HTTP client for all local server API calls.
  ///
  /// Uses [SecurityContext.withTrustedRoots = false] so Dart's embedded
  /// BoringSSL fully controls certificate validation â€” bypassing Android's
  /// OS-level TLS interception that silently drops self-signed cert connections
  /// on Android 13+ before [badCertificateCallback] ever fires.
  static void _initHttpClient() {
    if (kIsWeb) {
      _customClient = http.Client();
    } else {
      // withTrustedRoots: false â†’ no system CAs loaded, ensuring Dart calls
      // badCertificateCallback for every self-signed / local cert.
      final context = SecurityContext(withTrustedRoots: false);
      final ioHttpClient = HttpClient(context: context);
      ioHttpClient.badCertificateCallback =
          (X509Certificate cert, String host, int port) {
        final expectedFingerprint =
            _prefsCache?.getString('server_fingerprint');
        // Tolak jika fingerprint belum tersimpan (tidak ada trust-on-first-use)
        if (expectedFingerprint == null || expectedFingerprint.isEmpty) {
          return false;
        }
        return _certMatchesFingerprint(cert, expectedFingerprint);
      };
      _customClient = IOClient(ioHttpClient);
    }
  }

  /// Returns true only if the SHA-256 of [cert] (DER) equals [expectedFingerprint]
  /// (hex, case-insensitive). Empty fingerprint or any error â†’ reject.
  static bool _certMatchesFingerprint(
      X509Certificate cert, String expectedFingerprint) {
    if (expectedFingerprint.isEmpty) return false;
    try {
      final certHex = crypto.sha256.convert(cert.der).toString().toLowerCase();
      return certHex ==
          expectedFingerprint.toLowerCase(); // Tolak jika tidak cocok
    } catch (_) {
      return false;
    }
  }

  /// Creates a fresh one-shot [http.Client] that only trusts the self-signed
  /// certificate matching [fingerprint]. Used for pairing/verification to avoid
  /// any shared-client state issues on Android.
  static http.Client _createPinnedClient(String fingerprint) {
    if (kIsWeb) return http.Client();
    final context = SecurityContext(withTrustedRoots: false);
    final ioHttpClient = HttpClient(context: context)
      ..badCertificateCallback =
          (X509Certificate cert, String host, int port) =>
              _certMatchesFingerprint(cert, fingerprint);
    return IOClient(ioHttpClient);
  }

  @visibleForTesting
  static http.Client? testClient;
  static http.Client? _normalClient;

  static http.Client _getClient(String url) {
    if (testClient != null) return testClient!;
    // Only the explicitly paired legacy public endpoint uses system CAs.
    // LAN and Tailscale always use the certificate pinned during pairing.
    if (publicUrl.isNotEmpty &&
        Uri.parse(url).origin == Uri.parse(publicUrl).origin) {
      return _normalClient ??= http.Client();
    }
    if (_customClient == null) _initHttpClient();
    return _customClient!;
  }

  // â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
  // Internals & URL Formatting
  // â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

  static String _formatUrl(String input, {String? overrideProtocol}) {
    var host = input.trim();
    if (host.isEmpty) return 'https://192.168.1.1:8443';

    String scheme =
        overrideProtocol != null ? '$overrideProtocol://' : 'https://';
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
      host = '$host:8443';
    }

    return '$scheme$host';
  }

  static String get localIp =>
      _prefs.getString('last_ip') ?? '192.168.1.1:8443';
  static String get publicUrl => _prefs.getString('public_url') ?? '';
  static String get tailscaleUrl => _prefs.getString('tailscale_url') ?? '';
  static List<String> get alternateHosts =>
      _prefs.getStringList('alternate_hosts') ?? const [];
  static List<String> get tailscaleUrls =>
      _prefs.getStringList('tailscale_urls') ?? const [];
  static String get tailscaleState =>
      _prefs.getString('tailscale_state') ?? 'Unknown';
  static bool get tailscaleInstalled =>
      _prefs.getBool('tailscale_installed') ?? false;
  static String get serverId => _prefs.getString('server_id') ?? '';
  static bool get isTailscaleSynced =>
      tailscaleUrl.isNotEmpty || tailscaleUrls.isNotEmpty;

  static String get _remoteUrl {
    if (tailscaleUrl.isNotEmpty) return tailscaleUrl;
    if (tailscaleUrls.isNotEmpty) return tailscaleUrls.first;
    return publicUrl;
  }

  static bool _preferRemote = false;
  static bool get isUsingTailscale => _preferRemote && isTailscaleSynced;

  /// Returns whether the app is currently connected via local LAN WiFi or Public Internet Tunnel.
  static bool get isUsingPublicTunnel => _preferRemote && _remoteUrl.isNotEmpty;

  static String? _normalizeHttpsOrigin(String raw) {
    final trimmed = raw.trim();
    if (trimmed.isEmpty) return null;
    final uri = Uri.tryParse(trimmed);
    if (uri == null ||
        uri.scheme != 'https' ||
        uri.host.isEmpty ||
        uri.userInfo.isNotEmpty) {
      return null;
    }
    return uri.origin;
  }

  static bool _isRemoteEndpoint(String origin) {
    final localOrigins = <String>{
      _formatUrl(localIp),
      for (final h in alternateHosts) _formatUrl(h),
    };
    return !localOrigins.contains(origin);
  }

  static void _setActiveEndpoint(String origin) {
    _activeOverrideUrl = origin;
    _preferRemote = _isRemoteEndpoint(origin);
  }

  static List<String> _candidateEndpoints() {
    final seen = <String>{};
    void addCandidate(String? candidate) {
      if (candidate == null || candidate.isEmpty) return;
      final normalized = candidate.endsWith('/')
          ? candidate.substring(0, candidate.length - 1)
          : candidate;
      seen.add(normalized);
    }

    if (_activeOverrideUrl != null && _activeOverrideUrl!.isNotEmpty) {
      addCandidate(_activeOverrideUrl);
    }
    if (_preferRemote) {
      addCandidate(_normalizeHttpsOrigin(tailscaleUrl));
      for (final u in tailscaleUrls) {
        addCandidate(_normalizeHttpsOrigin(u));
      }
      addCandidate(_normalizeHttpsOrigin(publicUrl));
    }
    addCandidate(_formatUrl(localIp));
    for (final h in alternateHosts) {
      if (h.trim().isNotEmpty) {
        addCandidate(_formatUrl(h));
      }
    }
    addCandidate(_normalizeHttpsOrigin(tailscaleUrl));
    for (final u in tailscaleUrls) {
      addCandidate(_normalizeHttpsOrigin(u));
    }
    addCandidate(_normalizeHttpsOrigin(publicUrl));
    return seen.toList(growable: false);
  }

  static String get _activeBaseUrl {
    if (_activeOverrideUrl != null && _activeOverrideUrl!.isNotEmpty) {
      return _activeOverrideUrl!;
    }
    if (_preferRemote && _remoteUrl.isNotEmpty) {
      return _remoteUrl.endsWith('/')
          ? _remoteUrl.substring(0, _remoteUrl.length - 1)
          : _remoteUrl;
    }
    return _formatUrl(localIp);
  }

  static final _events = StreamController<Map<String, dynamic>>.broadcast();
  static Stream<Map<String, dynamic>> get events => _events.stream;
  static WebSocket? _socket;
  static bool get hasEventConnection => _socket != null;
  static String get serverVersion =>
      _prefs.getString('server_version') ?? 'Tidak diketahui';
  static HttpClient? _socketClient;
  static bool _connectingEvents = false;
  static int _eventGeneration = 0;

  static void stopEvents() {
    _eventGeneration++;
    _socket?.close();
    _socket = null;
    _socketClient?.close(force: true);
    _socketClient = null;
  }

  static Future<void> connectEvents() async {
    if (_socket != null || _connectingEvents || authToken.isEmpty || kIsWeb) {
      return;
    }
    _connectingEvents = true;
    final generation = _eventGeneration;
    final client = HttpClient(context: SecurityContext(withTrustedRoots: false))
      ..badCertificateCallback = (cert, host, port) => _certMatchesFingerprint(
          cert, _prefs.getString('server_fingerprint') ?? '');
    try {
      final endpoint =
          Uri.parse(_activeBaseUrl).replace(scheme: 'wss', path: '/events');
      final socket = await WebSocket.connect(endpoint.toString(),
              headers: _headers, customClient: client)
          .timeout(_timeout);
      if (generation != _eventGeneration) {
        await socket.close();
        client.close(force: true);
        return;
      }
      _socket = socket;
      _socketClient = client;
      socket.pingInterval = const Duration(seconds: 25);
      socket.listen((raw) async {
        if (generation != _eventGeneration || raw is! String) return;
        try {
          final event = jsonDecode(raw) as Map<String, dynamic>;
          final data = event['data'];
          if ((event['type'] == 'snapshot' || event['type'] == 'endpoints') &&
              data is Map) {
            await _applyConnectionData(Map<String, dynamic>.from(data));
          }
          if (generation != _eventGeneration) return;
          _events.add(event);
        } catch (_) {/* Ignore malformed events; REST remains available. */}
      }, onDone: () {
        if (identical(_socket, socket)) {
          _socket = null;
          client.close(force: true);
        }
      }, onError: (Object _) {
        if (identical(_socket, socket)) {
          _socket = null;
          client.close(force: true);
        }
      });
    } catch (_) {
      client.close(force: true);
    } finally {
      _connectingEvents = false;
    }
  }

  static String get authToken => _token;
  static String get authPin => authToken;

  static Map<String, String> get _headers {
    final token = authToken;
    final headers = <String, String>{
      'Content-Type': 'application/json',
      'X-Device-Name': 'Flutter Mobile App',
    };

    if (token.isNotEmpty) {
      headers['Authorization'] = 'Bearer $token';
    }

    return headers;
  }

  static bool _revoking = false;
  static void _check401(http.Response response) async {
    final sentToken = response.request?.headers['Authorization'] ??
        response.request?.headers['authorization'];
    if (sentToken != null && sentToken != 'Bearer $authToken') return;
    if (response.statusCode == 401 && !_revoking && authToken.isNotEmpty) {
      _revoking = true;
      try {
        await clearToken();
      } finally {
        _revoking = false;
      }

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
          const SnackBar(
              content: Text('Sesi telah berakhir, silakan hubungkan kembali'),
              backgroundColor: Colors.orange),
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

  static Future<void> _applyConnectionData(Map<String, dynamic> data) async {
    final hostsRaw = data['hosts'] ?? data['alternate_hosts'];
    if (hostsRaw is List) {
      final hosts = hostsRaw
          .map((e) => e.toString().trim())
          .where((e) => e.isNotEmpty)
          .toSet()
          .toList();
      await _prefs.setStringList('alternate_hosts', hosts);
    }

    List<String> validTailscaleUrls = [];
    if (data['tailscale_urls'] is List) {
      validTailscaleUrls = (data['tailscale_urls'] as List)
          .map((e) => _normalizeHttpsOrigin(e.toString()))
          .whereType<String>()
          .toSet()
          .toList();
      await _prefs.setStringList('tailscale_urls', validTailscaleUrls);
    }

    if (data['tailscale_url'] is String) {
      final normalized = _normalizeHttpsOrigin(data['tailscale_url'] as String);
      if (normalized != null) {
        await _prefs.setString('tailscale_url', normalized);
      } else if ((data['tailscale_url'] as String).trim().isEmpty) {
        await _prefs.setString(
          'tailscale_url',
          validTailscaleUrls.isNotEmpty ? validTailscaleUrls.first : '',
        );
      }
    } else if (validTailscaleUrls.isNotEmpty && tailscaleUrl.isEmpty) {
      await _prefs.setString('tailscale_url', validTailscaleUrls.first);
    }

    if (data['tailscale_state'] is String) {
      await _prefs.setString(
          'tailscale_state', data['tailscale_state'] as String);
    }
    if (data['tailscale_installed'] is bool) {
      await _prefs.setBool(
          'tailscale_installed', data['tailscale_installed'] as bool);
    }
    if (data['server_id'] is String &&
        (data['server_id'] as String).isNotEmpty) {
      await _prefs.setString('server_id', data['server_id'] as String);
    }
    if (data['version'] is String && (data['version'] as String).isNotEmpty) {
      await _prefs.setString('server_version', data['version'] as String);
    }
  }

  /// Synchronizes LAN and Tailscale endpoints from `GET /connection` without re-pairing.
  static Future<bool> syncConnectionEndpoints({bool refresh = false}) async {
    if (authToken.isEmpty) return false;
    if (!refresh && _syncInFlight != null) return _syncInFlight!;
    final future = () async {
      final path = refresh ? '/connection?refresh=1' : '/connection';
      final data = await _get(path,
          timeout: const Duration(seconds: 5), showSnackBar: false);
      if (data == null) return false;
      await _applyConnectionData(data);
      _lastEndpointSyncAt = DateTime.now();
      return true;
    }();
    _syncInFlight = future;
    try {
      return await future;
    } finally {
      if (identical(_syncInFlight, future)) {
        _syncInFlight = null;
      }
    }
  }

  static void _scheduleEndpointSync() {
    if (authToken.isEmpty || testClient != null) return;
    final last = _lastEndpointSyncAt;
    if (last != null &&
        DateTime.now().difference(last) < const Duration(seconds: 60)) {
      return;
    }
    unawaited(syncConnectionEndpoints());
  }

  static Future<({String url, Map<String, dynamic> health})?>
      _probeBestEndpointWithHealth({
    Set<String> exclude = const {},
    List<String>? candidatesOverride,
    Duration timeout = const Duration(seconds: 3),
  }) async {
    final candidates = (candidatesOverride ?? _candidateEndpoints())
        .where((c) => !exclude.contains(c))
        .toList();
    if (candidates.isEmpty) return null;

    final completer =
        Completer<({String url, Map<String, dynamic> health})?>();
    var remaining = candidates.length;
    final expectedServerId = serverId;

    for (final base in candidates) {
      () async {
        try {
          final response = await _getClient(base)
              .get(Uri.parse('$base/health'), headers: _headers)
              .timeout(timeout);
          if (response.statusCode == 200) {
            final data = jsonDecode(response.body) as Map<String, dynamic>;
            final respServerId = data['server_id'];
            if (expectedServerId.isEmpty ||
                respServerId is! String ||
                respServerId.isEmpty ||
                respServerId == expectedServerId) {
              if (!completer.isCompleted) {
                completer.complete((url: base, health: data));
              }
              return;
            }
          }
        } catch (_) {}
        remaining--;
        if (remaining == 0 && !completer.isCompleted) {
          completer.complete(null);
        }
      }();
    }

    final winner = await completer.future;
    if (winner != null) {
      _setActiveEndpoint(winner.url);
      if (winner.health['version'] is String) {
        await _prefs.setString(
            'server_version', winner.health['version'] as String);
      }
    }
    return winner;
  }

  static Future<String?> _probeBestEndpoint({
    Set<String> exclude = const {},
  }) async {
    final result = await _probeBestEndpointWithHealth(exclude: exclude);
    return result?.url;
  }

  /// Probes saved Tailscale endpoints (`tailscale_url` and `tailscale_urls`) directly.
  static Future<bool> testTailscaleConnection() async {
    final candidates = <String>{
      if (_normalizeHttpsOrigin(tailscaleUrl) != null)
        _normalizeHttpsOrigin(tailscaleUrl)!,
      for (final u in tailscaleUrls)
        if (_normalizeHttpsOrigin(u) != null) _normalizeHttpsOrigin(u)!,
    }.toList();
    if (candidates.isEmpty) return false;
    final winner =
        await _probeBestEndpointWithHealth(candidatesOverride: candidates);
    return winner != null;
  }

  // ──────────────────────────────────────
  // Generic HTTP helpers (Smart Roaming & Idempotent Failover)
  // ──────────────────────────────────────

  static const _timeout = Duration(seconds: 10);

  /// Generic GET request with automatic parallel candidate failover.
  static Future<Map<String, dynamic>?> _get(String path,
      {Duration? timeout, bool showSnackBar = true}) async {
    if (path != '/health' && authToken.isEmpty) return null;
    final headers = _headers;
    final active = _activeBaseUrl;
    Object? lastError;

    try {
      final response = await _getClient(active)
          .get(Uri.parse('$active$path'), headers: headers)
          .timeout(timeout ?? const Duration(seconds: 4));
      _check401(response);
      if (response.statusCode == 401 || response.statusCode == 403) {
        return null;
      }
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        _setActiveEndpoint(active);
        if (path == '/health') {
          if (data['version'] is String) {
            await _prefs.setString('server_version', data['version'] as String);
          }
          _scheduleEndpointSync();
        }
        return data;
      }
      if (response.statusCode < 500) return null;
    } catch (error) {
      lastError = error;
    }

    final fallback = await _probeBestEndpointWithHealth(exclude: {active});
    if (fallback != null) {
      if (path == '/health') {
        _scheduleEndpointSync();
        return fallback.health;
      }
      try {
        final response = await _getClient(fallback.url)
            .get(Uri.parse('${fallback.url}$path'), headers: headers)
            .timeout(timeout ?? const Duration(seconds: 4));
        _check401(response);
        if (response.statusCode == 200) {
          return jsonDecode(response.body) as Map<String, dynamic>;
        }
      } catch (error) {
        lastError = error;
      }
    }

    if (lastError != null) _handleError(lastError, showSnackBar: showSnackBar);
    return null;
  }

  static bool _isNetworkFailure(Object error) =>
      error is SocketException ||
      error is http.ClientException ||
      error is TimeoutException ||
      error is HandshakeException;

  /// Sends a POST command with idempotent `X-Request-Id` failover across reachable routes.
  static Future<bool> _post(String path,
      {Map<String, dynamic>? body, bool showSnackBar = true}) async {
    if (authToken.isEmpty) return false;
    final requestId = List.generate(
        16,
        (_) =>
            Random.secure().nextInt(256).toRadixString(16).padLeft(2, '0')).join();
    final deadlineMs =
        (DateTime.now().millisecondsSinceEpoch + 10000).toString();
    final postHeaders = {
      ..._headers,
      'X-Command-Deadline-Ms': deadlineMs,
      'X-Request-Id': requestId,
    };
    final encodedBody = jsonEncode(body ?? {});
    final url = _activeBaseUrl;

    try {
      final response = await _getClient(url)
          .post(Uri.parse('$url$path'), headers: postHeaders, body: encodedBody)
          .timeout(_timeout);
      _check401(response);
      return response.statusCode == 200;
    } catch (error) {
      if (_isNetworkFailure(error)) {
        final fallback = await _probeBestEndpoint(exclude: {url});
        if (fallback != null) {
          try {
            final retryResponse = await _getClient(fallback)
                .post(Uri.parse('$fallback$path'),
                    headers: postHeaders, body: encodedBody)
                .timeout(_timeout);
            _check401(retryResponse);
            return retryResponse.statusCode == 200;
          } catch (retryError) {
            _handleError(retryError, showSnackBar: showSnackBar);
            return false;
          }
        }
      }
      _handleError(error, showSnackBar: showSnackBar);
      return false;
    }
  }

  /// Primary QR code pairing method.
  /// Scans one-time pairing token from QR code payload and exchanges for a 24h Bearer session token.
  static Future<String?> pairWithQR({
    required String host,
    required String port,
    required String pairToken,
    String protocol = 'https',
    String fingerprint = '',
    String serverName = 'PC Remote',
    String publicUrl = '',
    String tailscaleUrl = '',
    String serverId = '',
    List<String> alternateHosts = const [],
  }) async {
    if (protocol != 'https' ||
        !RegExp(r'^[a-fA-F0-9]{64}$').hasMatch(fingerprint)) {
      return null;
    }
    final local =
        Uri(scheme: 'https', host: host, port: int.parse(port)).origin;
    final remote = tailscaleUrl.isNotEmpty ? tailscaleUrl : publicUrl;
    if (remote.isNotEmpty) {
      final uri = Uri.tryParse(remote);
      if (uri == null ||
          uri.scheme != 'https' ||
          uri.host.isEmpty ||
          uri.userInfo.isNotEmpty) {
        return null;
      }
    }
    final pinned = _createPinnedClient(fingerprint);
    final normal = http.Client();
    try {
      // Probe routes before spending the single-use token. Never replay pairing
      // after an ambiguous POST timeout (the first request may have succeeded).
      String? selected;
      final lanEndpoints = {
        local,
        for (final h in alternateHosts)
          Uri(scheme: 'https', host: h, port: int.parse(port)).origin
      };
      for (final endpoint in {
        ...lanEndpoints,
        if (remote.isNotEmpty) Uri.parse(remote).origin
      }) {
        final client = !lanEndpoints.contains(endpoint) && tailscaleUrl.isEmpty
            ? normal
            : pinned;
        try {
          final probe = await client
              .get(Uri.parse('$endpoint/health'))
              .timeout(const Duration(seconds: 3));
          if (probe.statusCode == 200) {
            selected = endpoint;
            break;
          }
        } catch (_) {}
      }
      if (selected == null) return null;
      final client = !lanEndpoints.contains(selected) && tailscaleUrl.isEmpty
          ? normal
          : pinned;
      final response = await client
          .post(
            Uri.parse('$selected/auth/pair'),
            headers: {
              'Content-Type': 'application/json',
              'X-Device-Name': 'Flutter Mobile'
            },
            body: jsonEncode(
                {'pair_token': pairToken, 'device_name': 'Flutter Mobile'}),
          )
          .timeout(_timeout);
      if (response.statusCode != 200) return null;
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      final token = data['token'] as String?;
      if (token == null || token.isEmpty) return null;
      if (serverId.isNotEmpty &&
          data['server_id'] != null &&
          data['server_id'] != serverId) {
        return null;
      }
      stopEvents();
      await _prefs.setString('server_fingerprint', fingerprint);
      await _prefs.setString(
          'last_ip', lanEndpoints.contains(selected) ? selected : local);
      await _prefs.setStringList('alternate_hosts', alternateHosts);
      await _prefs.setString(
          'public_url', tailscaleUrl.isEmpty ? publicUrl : '');
      await _prefs.setString('tailscale_url', tailscaleUrl);
      if (tailscaleUrl.isNotEmpty) {
        final normalized = _normalizeHttpsOrigin(tailscaleUrl);
        if (normalized != null) {
          await _prefs.setStringList('tailscale_urls', [normalized]);
        }
      }
      await _prefs.setString('server_id', serverId);
      await saveToken(token);
      _setActiveEndpoint(selected);
      _customClient?.close();
      _initHttpClient();
      unawaited(syncConnectionEndpoints());
      return token;
    } catch (error) {
      _handleError(error);
      return null;
    } finally {
      pinned.close();
      normal.close();
    }
  }

  /// Verifies whether an existing cached Bearer session token is still valid on the server.
  /// Returns:
  /// - `true`: Token is valid and authenticated
  /// - `false`: Token was explicitly revoked / expired (401 Unauthorized)
  /// - `null`: Server is offline / network unreachable (temporary, should not wipe token)
  static Future<bool?> verifySession(String ip, String token) async {
    final candidates = <String>{
      _formatUrl(ip),
      ..._candidateEndpoints(),
    };
    for (final base in candidates) {
      try {
        final response = await _getClient(base)
            .get(Uri.parse('$base/auth/verify'), headers: {
          'Authorization': 'Bearer $token'
        }).timeout(const Duration(seconds: 4));
        if (response.statusCode == 200) {
          _setActiveEndpoint(base);
          if (testClient == null) {
            unawaited(syncConnectionEndpoints());
          }
          return true;
        }
        if (response.statusCode == 401) return false;
      } catch (_) {}
    }
    return null;
  }

  /// Logs out the device by notifying the server to revoke the session,
  /// then clears all local authentication tokens and state.
  static Future<void> logout() async {
    try {
      await _post('/auth/logout', showSnackBar: false);
    } catch (_) {}
    await clearToken();
  }

  /// Manual login with IP and PIN (calls /auth/login, receives session token).

  static Future<String?> login(String ip, String pin) async {
    try {
      final storedFingerprint = _prefs.getString('server_fingerprint');
      if (storedFingerprint == null || storedFingerprint.isEmpty) {
        snackbarKey.currentState?.showSnackBar(const SnackBar(
            content: Text('Hubungkan dengan scan QR terlebih dahulu'),
            backgroundColor: Colors.orange));
        return null;
      }

      final formattedUrl = _formatUrl(ip);
      debugPrint('Attempting login to: $formattedUrl/auth/login');

      // 1. Try modern /auth/login endpoint
      try {
        final response = await _getClient(_activeBaseUrl)
            .post(
              Uri.parse('$formattedUrl/auth/login'),
              headers: {
                'Content-Type': 'application/json',
                'X-Device-Name': 'Flutter Mobile App',
              },
              body: jsonEncode({'pin': pin}),
            )
            .timeout(const Duration(seconds: 5));

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
              content:
                  Text('Terlalu banyak percobaan. Akun terkunci sementara.'),
              backgroundColor: Colors.orange,
            ),
          );
          return null;
        }
      } catch (_) {
        // Fall back to legacy probe if /auth/login fails (e.g., older server version)
      }

      // 2. Legacy fallback probe
      final legacyResponse = await _getClient(_activeBaseUrl).get(
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
      final url = _activeBaseUrl;
      final response = await _getClient(_activeBaseUrl)
          .post(
            Uri.parse('$url/system/pin'),
            headers: _headers,
            body: jsonEncode({
              'current_pin': currentPin,
              'new_pin': newPin,
            }),
          )
          .timeout(_timeout);

      if (response.statusCode == 200) {
        final context = navigatorKey.currentContext;
        if (context != null && context.mounted) {
          final ip = _prefs.getString('last_ip') ?? '';
          Provider.of<AppState>(context, listen: false)
              .setConnectionDetails(ip, authToken);
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

  // â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
  // Health Check (lightweight ping)
  // â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

  static Future<Map<String, dynamic>?> healthCheck() =>
      _get('/health', timeout: const Duration(seconds: 5), showSnackBar: false);

  static Future<String?> getLatestGitHubRelease() async {
    try {
      final response = await http
          .get(
            Uri.parse(
                'https://api.github.com/repos/KristianEki11/PC-Remote/releases/latest'),
          )
          .timeout(const Duration(seconds: 5));
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return data['tag_name'] as String?;
      }
    } catch (e) {
      debugPrint('Gagal mengambil rilis GitHub terbaru: $e');
    }
    return null;
  }

  // â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
  // Audio
  // â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

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

  static Future<bool> setDeviceVolume(String deviceId, double level) => _post(
      '/audio/device/volume',
      body: {'device_id': deviceId, 'level': VolumeHelpers.toScalar(level)});

  static Future<bool> toggleDeviceMute(String deviceId, bool newMute) =>
      _post('/audio/device/mute',
          body: {'device_id': deviceId, 'mute': newMute});

  // â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
  // Media
  // â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

  static Future<bool> mediaPlayPause() => _post('/media/play');
  static Future<bool> mediaNext() => _post('/media/next');
  static Future<bool> mediaPrev() => _post('/media/prev');
  static Future<bool> mediaSeekForward() => _post('/media/seek-forward');
  static Future<bool> mediaSeekBackward() => _post('/media/seek-backward');
  static Future<Map<String, dynamic>?> getMediaStatus() => _get('/media/status',
      timeout: const Duration(seconds: 5), showSnackBar: false);

  // â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
  // Browser
  // â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

  static Future<bool> openBrowser(String browserUrl) =>
      _post('/browser/open', body: {'url': browserUrl});

  // â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
  // System
  // â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

  static Future<bool> lockPc() => _post('/system/lock');
  static Future<bool> sleepPc() => _post('/system/sleep');
  static Future<bool> restartPc() => _post('/system/restart');
  static Future<bool> displayOff() => _post('/system/display/off');
  static Future<double?> getBrightness() async {
    final data = await _get('/system/brightness',
        timeout: const Duration(seconds: 5), showSnackBar: false);
    if (data != null && data['brightness'] != null) {
      return (data['brightness'] as num).toDouble();
    }
    return null;
  }

  static Future<bool> setBrightness(double level) =>
      _post('/system/brightness', body: {'level': level});
  static Future<bool> shutdownPc({int delaySeconds = 0}) =>
      _post('/system/shutdown', body: {'delay_seconds': delaySeconds});
}
