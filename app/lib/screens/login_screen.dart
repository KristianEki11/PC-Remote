import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:provider/provider.dart';
import '../services/api_service.dart';
import '../models/app_state.dart';
import '../utils/theme.dart';
import 'dashboard_screen.dart';
import 'qr_scan_screen.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final TextEditingController _ipController = TextEditingController();
  final TextEditingController _pinController = TextEditingController();
  bool _isLoading = false;
  String? _errorMessage;
  String _versionText = 'v3.0.0';

  @override
  void initState() {
    super.initState();
    _checkAutoLogin();
    _loadVersionInfo();
  }

  Future<void> _checkAutoLogin() async {
    final prefs = await SharedPreferences.getInstance();
    final savedIp = prefs.getString('last_ip');
    final savedToken = prefs.getString('auth_token');

    if (savedIp != null) {
      _ipController.text = savedIp;
    }

    if (savedIp != null && savedToken != null && savedToken.isNotEmpty) {
      setState(() {
        _isLoading = true;
      });

      // Try login/ping to verify token
      final token = await ApiService.login(savedIp, savedToken);
      if (token != null) {
        if (!mounted) return;
        Provider.of<AppState>(context, listen: false).setConnectionDetails(savedIp, token);

        // Update server version cache in background
        ApiService.healthCheck().then((health) {
          if (health != null && health['version'] != null) {
            prefs.setString('server_version', health['version'] as String);
          }
        });

        Navigator.pushReplacement(
          context,
          MaterialPageRoute(builder: (context) => const DashboardScreen()),
        );
      } else {
        // If login fails, check if the server is just offline
        final health = await ApiService.healthCheck();
        if (health == null) {
          // Server is unreachable (offline). We still auto-login using cached details
          // so the user can access the dashboard in offline mode.
          if (!mounted) return;
          Provider.of<AppState>(context, listen: false).setConnectionDetails(savedIp, savedToken);
          Provider.of<AppState>(context, listen: false).setConnectionStatus(false);
          Navigator.pushReplacement(
            context,
            MaterialPageRoute(builder: (context) => const DashboardScreen()),
          );
        } else {
          // Server is online but credentials failed (PIN/session expired)
          await prefs.remove('auth_token');
          setState(() {
            _isLoading = false;
            _errorMessage = 'Sesi telah berakhir atau PIN salah. Silakan hubungkan kembali.';
          });
        }
      }
    }
  }

  Future<void> _loadVersionInfo() async {
    final prefs = await SharedPreferences.getInstance();
    
    // 1. Jika di Web, coba ambil rilis terbaru dari GitHub
    if (kIsWeb) {
      final gitHubVersion = await ApiService.getLatestGitHubRelease();
      if (gitHubVersion != null && mounted) {
        setState(() {
          _versionText = 'GitHub Release: $gitHubVersion';
        });
        return;
      }
    }
    
    // 2. Jika di APK/Mobile (atau fetch GitHub gagal), pakai cache versi server terakhir
    final savedServerVersion = prefs.getString('server_version');
    if (savedServerVersion != null && mounted) {
      setState(() {
        _versionText = 'Server v$savedServerVersion';
      });
    }
    
    // 3. Coba ping server secara asinkron untuk update versi terbaru di latar belakang
    final savedIp = prefs.getString('last_ip');
    if (savedIp != null) {
      try {
        final health = await ApiService.healthCheck();
        if (health != null && health['version'] != null && mounted) {
          final version = health['version'] as String;
          await prefs.setString('server_version', version);
          setState(() {
            _versionText = 'Server v$version';
          });
        }
      } catch (e) {
        debugPrint('Gagal ping server untuk ambil versi: $e');
      }
    }
  }

  Future<void> _handleLogin() async {
    final ip = _ipController.text.trim();
    final pin = _pinController.text.trim();

    if (ip.isEmpty || pin.isEmpty) {
      setState(() {
        _errorMessage = 'IP dan PIN tidak boleh kosong';
      });
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    final token = await ApiService.login(ip, pin);

    if (token != null) {
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString('last_ip', ip);
      await prefs.setString('auth_token', token);

      // Update server version cache in background
      ApiService.healthCheck().then((health) {
        if (health != null && health['version'] != null) {
          prefs.setString('server_version', health['version'] as String);
        }
      });

      if (!mounted) return;

      Provider.of<AppState>(context, listen: false).setConnectionDetails(ip, token);

      Navigator.pushReplacement(
        context,
        MaterialPageRoute(builder: (context) => const DashboardScreen()),
      );
    } else {
      if (!mounted) return;
      setState(() {
        _isLoading = false;
        _errorMessage = 'PIN salah atau server tidak dapat dijangkau';
      });
    }
  }

  void _openQRScanner() {
    Navigator.of(context).push(
      MaterialPageRoute(builder: (context) => const QRScanScreen()),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Container(
        decoration: const BoxDecoration(
          // Minimal gradient background
          gradient: RadialGradient(
            center: Alignment(0, -0.3),
            radius: 1.0,
            colors: [
              Color(0xFF252630), // Slightly lighter center
              AppColors.background,
            ],
          ),
        ),
        child: SafeArea(
          child: Center(
            child: TweenAnimationBuilder<Offset>(
              tween: Tween<Offset>(begin: const Offset(0, 50), end: Offset.zero),
              duration: const Duration(milliseconds: 400),
              curve: Curves.easeOutCubic,
              builder: (context, offset, child) {
                return Transform.translate(
                  offset: offset,
                  child: Opacity(
                    opacity: 1.0 - (offset.dy / 50).clamp(0.0, 1.0),
                    child: child,
                  ),
                );
              },
              child: SingleChildScrollView(
                child: Container(
                  constraints: const BoxConstraints(maxWidth: 380),
                  padding: const EdgeInsets.all(28.0),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      // App icon with claymorphic appearance
                      Container(
                        width: 90,
                        height: 90,
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          color: AppColors.surfaceLight,
                          boxShadow: AppClays.card(),
                        ),
                        child: const Icon(
                          Icons.computer_rounded,
                          size: 44,
                          color: AppColors.primary,
                        ),
                      ),
                      const SizedBox(height: 20),
                      const Text(
                        'PC Remote',
                        style: TextStyle(
                          fontSize: 30,
                          fontWeight: FontWeight.w800,
                          letterSpacing: -1,
                          color: AppColors.textPrimary,
                        ),
                      ),
                      const SizedBox(height: 6),
                      const Text(
                        'Kontrol PC dari genggaman tangan',
                        style: TextStyle(
                          fontSize: 14,
                          color: AppColors.textSecondary,
                          letterSpacing: 0.3,
                        ),
                      ),
                      const SizedBox(height: 36),

                      // ── Primary Action: Scan QR Code Button ─────────
                      if (!kIsWeb) ...[
                        SizedBox(
                          width: double.infinity,
                          height: 56,
                          child: Container(
                            decoration: BoxDecoration(
                              gradient: AppGradients.primaryButton,
                              borderRadius: BorderRadius.circular(16),
                              boxShadow: AppClays.button(),
                            ),
                            child: ElevatedButton.icon(
                              onPressed: _isLoading ? null : _openQRScanner,
                              icon: const Icon(Icons.qr_code_scanner_rounded, color: Colors.white, size: 24),
                              label: const Text(
                                'Pindai QR Code di PC',
                                style: TextStyle(
                                  fontSize: 16,
                                  fontWeight: FontWeight.w700,
                                  color: Colors.white,
                                  letterSpacing: 0.3,
                                ),
                              ),
                              style: ElevatedButton.styleFrom(
                                backgroundColor: Colors.transparent,
                                shadowColor: Colors.transparent,
                                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                              ),
                            ),
                          ),
                        ),
                        const SizedBox(height: 24),

                        // Divider with text
                        Row(
                          children: [
                            Expanded(child: Divider(color: Colors.white.withValues(alpha: 0.15))),
                            Padding(
                              padding: const EdgeInsets.symmetric(horizontal: 14),
                              child: Text(
                                'atau masukkan manual',
                                style: TextStyle(
                                  fontSize: 12,
                                  color: AppColors.textSecondary.withValues(alpha: 0.7),
                                ),
                              ),
                            ),
                            Expanded(child: Divider(color: Colors.white.withValues(alpha: 0.15))),
                          ],
                        ),
                        const SizedBox(height: 24),
                      ],

                      // ── Secondary Action: Manual IP & PIN Input ─────
                      TextField(
                        controller: _ipController,
                        keyboardType: TextInputType.url,
                        decoration: const InputDecoration(
                          labelText: 'IP Address / URL',
                          hintText: '192.168.1.x',
                          prefixIcon: Icon(Icons.wifi, color: AppColors.textSecondary),
                        ),
                      ),
                      const SizedBox(height: 16),
                      TextField(
                        controller: _pinController,
                        obscureText: true,
                        keyboardType: TextInputType.number,
                        maxLength: 8,
                        decoration: const InputDecoration(
                          labelText: 'PIN',
                          prefixIcon: Icon(Icons.lock_outline, color: AppColors.textSecondary),
                          counterText: '',
                        ),
                      ),
                      const SizedBox(height: 20),

                      // Error message display
                      if (_errorMessage != null) ...[
                        Container(
                          width: double.infinity,
                          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                          decoration: BoxDecoration(
                            color: AppColors.error.withValues(alpha: 0.1),
                            borderRadius: BorderRadius.circular(8),
                            border: Border.all(color: AppColors.error.withValues(alpha: 0.3)),
                          ),
                          child: Row(
                            children: [
                              const Icon(Icons.error_outline, color: AppColors.error, size: 18),
                              const SizedBox(width: 8),
                              Expanded(
                                child: Text(
                                  _errorMessage!,
                                  style: const TextStyle(color: AppColors.error, fontSize: 13),
                                ),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 16),
                      ],

                      // Manual Connect Button
                      SizedBox(
                        width: double.infinity,
                        height: 50,
                        child: OutlinedButton(
                          onPressed: _isLoading ? null : _handleLogin,
                          style: OutlinedButton.styleFrom(
                            foregroundColor: AppColors.textPrimary,
                            side: BorderSide(color: AppColors.primary.withValues(alpha: 0.6), width: 1.5),
                            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                          ),
                          child: _isLoading
                              ? const SizedBox(
                                  width: 20,
                                  height: 20,
                                  child: CircularProgressIndicator(
                                    color: AppColors.primary,
                                    strokeWidth: 2,
                                  ),
                                )
                              : const Text(
                                  'Hubungkan Manual',
                                  style: TextStyle(
                                    fontSize: 15,
                                    fontWeight: FontWeight.w600,
                                    color: AppColors.textPrimary,
                                  ),
                                ),
                        ),
                      ),
                      const SizedBox(height: 32),

                      // Version text
                      Text(
                        _versionText,
                        style: TextStyle(
                          fontSize: 12,
                          color: AppColors.textSecondary.withValues(alpha: 0.85),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
