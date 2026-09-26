import 'dart:ui';
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
  String _versionText = 'v4.0.0';

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

      // Verify existing session token with server
      final isValid = await ApiService.verifySession(savedIp, savedToken);

      if (!mounted) return;

      if (isValid == true) {
        // Session is verified and active!
        Provider.of<AppState>(context, listen: false).setConnectionDetails(savedIp, savedToken);
        Provider.of<AppState>(context, listen: false).setConnectionStatus(true);

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
      } else if (isValid == null) {
        // Server is offline / network unreachable (temporary).
        // DO NOT wipe the token! Allow entering dashboard in offline mode.
        Provider.of<AppState>(context, listen: false).setConnectionDetails(savedIp, savedToken);
        Provider.of<AppState>(context, listen: false).setConnectionStatus(false);

        Navigator.pushReplacement(
          context,
          MaterialPageRoute(builder: (context) => const DashboardScreen()),
        );
      } else {
        // isValid == false -> Session was explicitly revoked on PC (401 Unauthorized)
        await prefs.remove('auth_token');
        setState(() {
          _isLoading = false;
          _errorMessage = 'Sesi telah diputus dari PC. Silakan pindai QR Code kembali.';
        });
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
          gradient: AppGradients.glassBackground,
        ),
        child: SafeArea(
          child: Center(
            child: TweenAnimationBuilder<Offset>(
              tween: Tween<Offset>(begin: const Offset(0, 40), end: Offset.zero),
              duration: const Duration(milliseconds: 600),
              curve: Curves.easeOutCubic,
              builder: (context, offset, child) {
                return Transform.translate(
                  offset: offset,
                  child: Opacity(
                    opacity: 1.0 - (offset.dy / 40).clamp(0.0, 1.0),
                    child: child,
                  ),
                );
              },
              child: SingleChildScrollView(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 24.0),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      // App icon with glassmorphism
                      Container(
                        width: 100,
                        height: 100,
                        decoration: AppGlass.cardDecoration().copyWith(
                          shape: BoxShape.circle,
                          borderRadius: null,
                        ),
                        child: const Icon(
                          Icons.computer_rounded,
                          size: 48,
                          color: AppColors.primary,
                        ),
                      ),
                      const SizedBox(height: 24),
                      const Text(
                        'PC Remote',
                        style: TextStyle(
                          fontSize: 32,
                          fontWeight: FontWeight.w800,
                          letterSpacing: -1,
                          color: AppColors.textPrimary,
                        ),
                      ),
                      const SizedBox(height: 8),
                      const Text(
                        'Kontrol PC dari genggaman tangan',
                        style: TextStyle(
                          fontSize: 16,
                          color: AppColors.textSecondary,
                          letterSpacing: 0.3,
                        ),
                      ),
                      const SizedBox(height: 40),

                      // Glassmorphism Card for form
                      ClipRRect(
                        borderRadius: BorderRadius.circular(24),
                        child: BackdropFilter(
                          filter: ImageFilter.blur(sigmaX: 20, sigmaY: 20),
                          child: Container(
                            constraints: const BoxConstraints(maxWidth: 400),
                            padding: const EdgeInsets.all(28.0),
                            decoration: AppGlass.cardDecoration(),
                            child: Column(
                              children: [
                                if (!kIsWeb) ...[
                                  SizedBox(
                                    width: double.infinity,
                                    height: 56, // Large touch target
                                    child: Container(
                                      decoration: BoxDecoration(
                                        gradient: AppGradients.primaryButton,
                                        borderRadius: BorderRadius.circular(16),
                                        boxShadow: [
                                          BoxShadow(
                                            color: AppColors.primary.withOpacity(0.3),
                                            blurRadius: 16,
                                            offset: const Offset(0, 6),
                                          ),
                                        ],
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
                                  Row(
                                    children: [
                                      Expanded(child: Divider(color: AppColors.textMuted.withOpacity(0.2))),
                                      Padding(
                                        padding: const EdgeInsets.symmetric(horizontal: 14),
                                        child: Text(
                                          'atau masukkan manual',
                                          style: TextStyle(
                                            fontSize: 13,
                                            color: AppColors.textMuted,
                                          ),
                                        ),
                                      ),
                                      Expanded(child: Divider(color: AppColors.textMuted.withOpacity(0.2))),
                                    ],
                                  ),
                                  const SizedBox(height: 24),
                                ],

                                // Input fields
                                TextField(
                                  controller: _ipController,
                                  keyboardType: TextInputType.url,
                                  decoration: const InputDecoration(
                                    labelText: 'IP Address / URL',
                                    hintText: '192.168.1.x',
                                    prefixIcon: Icon(Icons.wifi, color: AppColors.primary),
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
                                    prefixIcon: Icon(Icons.lock_outline, color: AppColors.primary),
                                    counterText: '',
                                  ),
                                ),
                                const SizedBox(height: 24),

                                if (_errorMessage != null) ...[
                                  Container(
                                    width: double.infinity,
                                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                                    decoration: BoxDecoration(
                                      color: AppColors.error.withOpacity(0.1),
                                      borderRadius: BorderRadius.circular(12),
                                      border: Border.all(color: AppColors.error.withOpacity(0.3)),
                                    ),
                                    child: Row(
                                      children: [
                                        const Icon(Icons.error_outline, color: AppColors.error, size: 20),
                                        const SizedBox(width: 12),
                                        Expanded(
                                          child: Text(
                                            _errorMessage!,
                                            style: const TextStyle(color: AppColors.error, fontSize: 14),
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                  const SizedBox(height: 16),
                                ],

                                SizedBox(
                                  width: double.infinity,
                                  height: 56, // Large touch target
                                  child: OutlinedButton(
                                    onPressed: _isLoading ? null : _handleLogin,
                                    style: OutlinedButton.styleFrom(
                                      foregroundColor: AppColors.primary,
                                      side: const BorderSide(color: AppColors.primary, width: 2),
                                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                                      backgroundColor: Colors.white.withOpacity(0.5),
                                    ),
                                    child: _isLoading
                                        ? const SizedBox(
                                            width: 24,
                                            height: 24,
                                            child: CircularProgressIndicator(
                                              color: AppColors.primary,
                                              strokeWidth: 2.5,
                                            ),
                                          )
                                        : const Text(
                                            'Hubungkan Manual',
                                            style: TextStyle(
                                              fontSize: 16,
                                              fontWeight: FontWeight.w700,
                                            ),
                                          ),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(height: 32),
                      Text(
                        _versionText,
                        style: TextStyle(
                          fontSize: 13,
                          fontWeight: FontWeight.w500,
                          color: AppColors.textMuted,
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
