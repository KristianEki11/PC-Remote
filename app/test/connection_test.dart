import 'dart:async';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:pc_remote/services/api_service.dart';
import 'package:pc_remote/services/companion_service.dart';
import 'package:pc_remote/models/audio_state.dart';
import 'package:pc_remote/models/media_status.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  setUp(() async {
    FlutterSecureStorage.setMockInitialValues({});
    SharedPreferences.setMockInitialValues({
      'last_ip': '192.168.1.2:8443',
      'tailscale_url': 'https://100.101.102.103:8443',
      'auth_token': 'legacy-token',
    });
    await ApiService.init();
  });
  tearDown(() {
    ApiService.testClient = null;
    ApiService.stopEvents();
  });

  test('migrates existing credential out of preferences', () async {
    expect(ApiService.authToken, 'legacy-token');
    expect((await SharedPreferences.getInstance()).getString('auth_token'),
        isNull);
    expect(await const FlutterSecureStorage().read(key: 'auth_token'),
        'legacy-token');
  });

  test('read fails over to paired Tailscale endpoint', () async {
    final hosts = <String>[];
    ApiService.testClient = MockClient((request) async {
      hosts.add(request.url.host);
      if (request.url.host.startsWith('192.')) {
        throw http.ClientException('offline');
      }
      return http.Response('{"status":"ok"}', 200);
    });
    expect(await ApiService.healthCheck(), {'status': 'ok'});
    expect(hosts, ['192.168.1.2', '100.101.102.103']);
    expect(ApiService.isUsingTailscale, isTrue);
  });

  test('POST timeout does not double-send when no fallback is reachable',
      () async {
    var postAttempts = 0;
    ApiService.testClient = MockClient((request) async {
      if (request.method == 'POST') {
        postAttempts++;
        expect(request.headers['X-Request-Id'],
            matches(RegExp(r'^[a-f0-9]{32}$')));
        final deadline = int.parse(request.headers['X-Command-Deadline-Ms']!);
        expect(deadline, greaterThan(DateTime.now().millisecondsSinceEpoch));
      }
      throw TimeoutException('response lost after execution');
    });
    expect(await ApiService.mediaNext(), isFalse);
    expect(postAttempts, 1);
  });

  test('server command error is not reported as success and does not failover',
      () async {
    var attempts = 0;
    ApiService.testClient = MockClient((_) async {
      attempts++;
      return http.Response('{}', 500);
    });
    expect(await ApiService.setVolume(.4), isFalse);
    expect(attempts, 1);
  });

  test('volume rollback uses confirmed value before slider drag', () async {
    ApiService.testClient = MockClient((request) async =>
        request.method == 'GET'
            ? http.Response('{"level":0.25,"muted":false}', 200)
            : http.Response('{}', 500));
    final state = AudioState();
    await state.fetchMasterVolume();
    state.updateMasterVolumeLocally(80);
    expect(await state.setMasterVolume(80), isFalse);
    expect(state.masterVolume, 25);
    state.dispose();
  });

  test('syncConnectionEndpoints persists hosts and tailscale metadata',
      () async {
    ApiService.testClient = MockClient((request) async {
      expect(request.url.path, '/connection');
      expect(request.url.queryParameters['refresh'], '1');
      return http.Response(
        '{"hosts":["192.168.1.2","192.168.50.10"],"tailscale_url":"https://100.88.99.10:8443","tailscale_urls":["https://100.88.99.10:8443","https://laptop.tail.ts.net:8443"],"tailscale_state":"Running","tailscale_installed":true,"server_id":"srv-123","version":"5.0.0"}',
        200,
      );
    });

    final ok = await ApiService.syncConnectionEndpoints(refresh: true);
    expect(ok, isTrue);
    expect(ApiService.alternateHosts, ['192.168.1.2', '192.168.50.10']);
    expect(ApiService.tailscaleUrl, 'https://100.88.99.10:8443');
    expect(ApiService.tailscaleUrls, [
      'https://100.88.99.10:8443',
      'https://laptop.tail.ts.net:8443',
    ]);
    expect(ApiService.tailscaleState, 'Running');
    expect(ApiService.tailscaleInstalled, isTrue);
    expect(ApiService.serverId, 'srv-123');
    expect(ApiService.serverVersion, '5.0.0');
  });

  test('parallel probe verifies server_id and fails over LAN -> Tailscale',
      () async {
    SharedPreferences.setMockInitialValues({
      'last_ip': '192.168.1.2:8443',
      'alternate_hosts': ['192.168.1.99'],
      'tailscale_url': 'https://100.101.102.103:8443',
      'server_id': 'expected-srv',
      'auth_token': 'legacy-token',
    });
    await ApiService.init();

    ApiService.testClient = MockClient((request) async {
      if (request.url.host == '192.168.1.2') {
        throw http.ClientException('lan down');
      }
      if (request.url.host == '192.168.1.99') {
        return http.Response(
            '{"status":"ok","server_id":"wrong-srv"}', 200);
      }
      if (request.url.host == '100.101.102.103') {
        return http.Response(
            '{"status":"ok","server_id":"expected-srv"}', 200);
      }
      return http.Response('{}', 404);
    });

    final health = await ApiService.healthCheck();
    expect(health?['server_id'], 'expected-srv');
    expect(ApiService.isUsingTailscale, isTrue);
  });

  test('idempotent POST failover reuses exact same X-Request-Id on fallback',
      () async {
    final postRequests = <http.Request>[];
    ApiService.testClient = MockClient((request) async {
      if (request.method == 'POST' && request.url.host == '192.168.1.2') {
        postRequests.add(request);
        throw http.ClientException('wifi dropped');
      }
      if (request.method == 'GET' &&
          request.url.host == '100.101.102.103' &&
          request.url.path == '/health') {
        return http.Response('{"status":"ok"}', 200);
      }
      if (request.method == 'POST' && request.url.host == '100.101.102.103') {
        postRequests.add(request);
        return http.Response('{"success":true}', 200);
      }
      return http.Response('{}', 404);
    });

    final ok = await ApiService.mediaPlayPause();
    expect(ok, isTrue);
    expect(postRequests.length, 2);
    expect(postRequests[0].url.host, '192.168.1.2');
    expect(postRequests[1].url.host, '100.101.102.103');
    expect(
      postRequests[0].headers['X-Request-Id'],
      postRequests[1].headers['X-Request-Id'],
    );
    expect(
      postRequests[0].headers['X-Command-Deadline-Ms'],
      postRequests[1].headers['X-Command-Deadline-Ms'],
    );
  });

  test('CompanionService no-ops gracefully without platform handler', () async {
    await CompanionService.init();
    await CompanionService.updateStatus(
      connected: true,
      routeLabel: 'Wi-Fi Lokal',
      media: MediaStatus(
        success: true,
        status: 'Playing',
        title: 'Test Song',
        artist: 'Test Artist',
        album: '',
        appId: 'Spotify',
      ),
    );
    await CompanionService.stopService();
    expect(await CompanionService.isEnabled(), isTrue);
  });
}
