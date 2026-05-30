import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:provider/provider.dart';
import '../services/api_service.dart';
import '../models/app_state.dart';
import '../utils/theme.dart';
import 'dashboard_screen.dart';

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
  String _versionText = 'v2.2.6';

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
          // Server is online but credentials failed (PIN changed)
          await prefs.remove('auth_token');
          setState(() {
            _isLoading = false;
            _errorMessage = 'PIN telah berubah. Silakan login kembali.';
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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Container(
        decoration: const BoxDecoration(
          // Subtle radial gradient background
          gradient: RadialGradient(
            center: Alignment(0, -0.5),
            radius: 1.2,
            colors: [
              Color(0xFF1A2332), // Slightly lighter center
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
                  constraints: const BoxConstraints(maxWidth: 400),
                  padding: const EdgeInsets.all(24.0),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    // App icon with gradient glow background
                    Container(
                      width: 100,
                      height: 100,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        gradient: LinearGradient(
                          colors: [
                            AppColors.primary.withValues(alpha: 0.2),
                            AppColors.gradientEnd.withValues(alpha: 0.1),
                          ],
                          begin: Alignment.topLeft,
                          end: Alignment.bottomRight,
                        ),
                        boxShadow: [
                          BoxShadow(
                            color: AppColors.primary.withValues(alpha: 0.15),
                            blurRadius: 30,
                            spreadRadius: 5,
                          ),
                        ],
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
                    const SizedBox(height: 6),
                    const Text(
                      'Kontrol PC dari genggaman tangan',
                      style: TextStyle(
                        fontSize: 14,
                        color: AppColors.textSecondary,
                        letterSpacing: 0.3,
                      ),
                    ),
                    const SizedBox(height: 48),
                    // IP Address Input
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
                    // PIN Input
                    TextField(
                      controller: _pinController,
                      obscureText: true,
                      keyboardType: TextInputType.number,
                      maxLength: 8,
                      decoration: const InputDecoration(
                        labelText: 'PIN',
                        prefixIcon: Icon(Icons.lock_outline, color: AppColors.textSecondary),
                        counterText: '', // hide character counter
                      ),
                    ),
                    const SizedBox(height: 24),
                    // Error message
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
                    // Connect Button with gradient
                    SizedBox(
                      width: double.infinity,
                      height: 52,
                      child: Container(
                        decoration: BoxDecoration(
                          gradient: _isLoading ? null : AppGradients.accent,
                          borderRadius: BorderRadius.circular(12),
                          boxShadow: _isLoading ? [] : [
                            BoxShadow(
                              color: AppColors.primary.withValues(alpha: 0.3),
                              blurRadius: 12,
                              offset: const Offset(0, 4),
                            ),
                          ],
                        ),
                        child: ElevatedButton(
                          onPressed: _isLoading ? null : _handleLogin,
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.transparent,
                            shadowColor: Colors.transparent,
                            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                          ),
                          child: _isLoading
                              ? const SizedBox(
                                  width: 22,
                                  height: 22,
                                  child: CircularProgressIndicator(
                                    color: Colors.white,
                                    strokeWidth: 2.5,
                                  ),
                                )
                              : const Text(
                                  'Hubungkan',
                                  style: TextStyle(
                                    fontSize: 16,
                                    fontWeight: FontWeight.w600,
                                    color: Colors.white,
                                    letterSpacing: 0.5,
                                  ),
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
