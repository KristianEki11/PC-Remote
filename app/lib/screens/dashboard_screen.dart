import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../services/api_service.dart';
import '../models/app_state.dart';
import '../models/media_state.dart';
import '../widgets/audio_card.dart';
import '../widgets/media_card.dart';
import '../widgets/browser_card.dart';
import '../widgets/system_card.dart';
import '../widgets/fade_in_stagger.dart';
import '../widgets/shared_card.dart';
import '../utils/theme.dart';
import 'login_screen.dart';
import 'settings_screen.dart';

class DashboardScreen extends StatefulWidget {
  const DashboardScreen({super.key});

  @override
  State<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends State<DashboardScreen> with TickerProviderStateMixin {
  int _currentIndex = 0;
  Timer? _pingTimer;
  int _failCount = 0;
  bool _isFirstPing = true;

  late AnimationController _pulseController;
  late Animation<double> _pulseAnimation;

  @override
  void initState() {
    super.initState();
    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    )..repeat(reverse: true);
    _pulseAnimation = Tween<double>(begin: 0.4, end: 1.0).animate(
      CurvedAnimation(parent: _pulseController, curve: Curves.easeInOut),
    );
    _startPing();
  }

  @override
  void dispose() {
    _pingTimer?.cancel();
    _pulseController.dispose();
    try {
      Provider.of<MediaState>(context, listen: false).stopPolling();
    } catch (_) {}
    super.dispose();
  }

  Future<void> _doPing() async {
    final health = await ApiService.healthCheck();
    if (!mounted) return;
    
    final appState = Provider.of<AppState>(context, listen: false);
    final wasConnected = appState.isConnected;

    if (health != null && health['status'] == 'ok') {
      _failCount = 0;
      appState.setConnectionStatus(true);
      if (mounted) {
        Provider.of<MediaState>(context, listen: false).startPolling();
      }
      if (!wasConnected && !_isFirstPing) {
        ScaffoldMessenger.of(context).hideCurrentSnackBar();
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Terhubung ke server (Online)'),
            backgroundColor: Colors.green,
            duration: Duration(seconds: 2),
          ),
        );
      }
    } else {
      _failCount++;
      if (_failCount >= 3) {
        appState.setConnectionStatus(false);
        if (mounted) {
          Provider.of<MediaState>(context, listen: false).stopPolling();
        }
        if (wasConnected && !_isFirstPing) {
          ScaffoldMessenger.of(context).hideCurrentSnackBar();
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Koneksi ke server terputus (Offline)'),
              backgroundColor: Colors.red,
              duration: Duration(seconds: 3),
            ),
          );
        }
      }
    }
    _isFirstPing = false;
  }

  void _startPing() {
    _pingTimer = Timer.periodic(const Duration(seconds: 10), (timer) async {
      await _doPing();
    });
    _doPing();
  }

  Future<void> _onRefresh() async {
    HapticFeedback.mediumImpact();
    await _doPing();
    if (mounted) setState(() {});
  }

  Future<void> _logout() async {
    HapticFeedback.mediumImpact();
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('auth_token');
    
    if (!mounted) return;
    Provider.of<AppState>(context, listen: false).clear();
    
    Navigator.pushReplacement(
      context,
      MaterialPageRoute(builder: (context) => const LoginScreen()),
    );
  }

  String _getGreeting() {
    final hour = DateTime.now().hour;
    if (hour >= 5 && hour < 12) {
      return 'Selamat Pagi';
    } else if (hour >= 12 && hour < 15) {
      return 'Selamat Siang';
    } else if (hour >= 15 && hour < 18) {
      return 'Selamat Sore';
    } else {
      return 'Selamat Malam';
    }
  }

  @override
  Widget build(BuildContext context) {
    final isConnected = context.watch<AppState>().isConnected;
    final appState = context.watch<AppState>();

    final List<Widget> tabs = [
      // Tab 0: Utama (Media & Browser)
      RefreshIndicator(
        onRefresh: _onRefresh,
        color: AppColors.primary,
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          padding: const EdgeInsets.all(16.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Dynamic greeting header
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 4.0, vertical: 8.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      '${_getGreeting()}, User!',
                      style: const TextStyle(
                        fontSize: 24,
                        fontWeight: FontWeight.w800,
                        color: AppColors.textPrimary,
                        letterSpacing: -0.5,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      'IP PC: ${appState.ipAddress}',
                      style: const TextStyle(
                        fontSize: 13,
                        color: AppColors.textSecondary,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16),
              const FadeInStagger(delayMs: 0, child: MediaCard()),
              const SizedBox(height: 16),
              const FadeInStagger(delayMs: 100, child: BrowserCard()),
              const SizedBox(height: 24),
            ],
          ),
        ),
      ),
      // Tab 1: Mixer (AudioCard handles its own layout and scroll)
      const AudioCard(),
      // Tab 2: Sistem (Power & Info)
      SingleChildScrollView(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          children: [
            const FadeInStagger(delayMs: 0, child: SystemCard()),
            const SizedBox(height: 16),
            // Server Info Detail Card
            FadeInStagger(
              delayMs: 100,
              child: SharedCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const CardHeader(
                      icon: Icons.info_outline_rounded,
                      title: 'Informasi Server',
                    ),
                    const SizedBox(height: 16),
                    _buildInfoRow('IP Address', appState.ipAddress),
                    _buildInfoRow('Platform', 'Windows'),
                    _buildInfoRow('Status Koneksi', isConnected ? 'Online' : 'Offline', isStatus: true, statusVal: isConnected),
                    _buildInfoRow('Versi Server', 'v2.2.4'),
                    const SizedBox(height: 20),
                    // Action Buttons inside card
                    Row(
                      children: [
                        Expanded(
                          child: ElevatedButton.icon(
                            onPressed: () {
                              Navigator.push(
                                context,
                                MaterialPageRoute(builder: (context) => const SettingsScreen()),
                              );
                            },
                            icon: const Icon(Icons.settings_outlined, size: 18),
                            label: const Text('Pengaturan'),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: ElevatedButton.icon(
                            onPressed: _logout,
                            style: ElevatedButton.styleFrom(
                              backgroundColor: Colors.red.withValues(alpha: 0.1),
                              foregroundColor: Colors.redAccent,
                            ),
                            icon: const Icon(Icons.logout_rounded, size: 18),
                            label: const Text('Keluar'),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),
          ],
        ),
      ),
    ];

    return Scaffold(
      appBar: AppBar(
        title: const Text('PC Remote'),
        actions: [
          // Animated connection status badge
          Container(
            margin: const EdgeInsets.only(right: 16),
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
            decoration: BoxDecoration(
              color: (isConnected ? Colors.green : Colors.red).withValues(alpha: 0.15),
              borderRadius: BorderRadius.circular(20),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                AnimatedBuilder(
                  animation: _pulseAnimation,
                  builder: (context, child) {
                    return Container(
                      width: 8,
                      height: 8,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: isConnected
                            ? Colors.green.withValues(alpha: _pulseAnimation.value)
                            : Colors.red,
                        boxShadow: isConnected
                            ? [BoxShadow(color: Colors.green.withValues(alpha: 0.4), blurRadius: 6, spreadRadius: 1)]
                            : [],
                      ),
                    );
                  },
                ),
                const SizedBox(width: 6),
                Text(
                  isConnected ? 'Online' : 'Offline',
                  style: TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w700,
                    color: isConnected ? Colors.green : Colors.red,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
      body: Column(
        children: [
          // Animated connection banner
          AnimatedContainer(
            duration: const Duration(milliseconds: 300),
            curve: Curves.easeInOut,
            height: isConnected ? 0 : 40,
            child: AnimatedOpacity(
              duration: const Duration(milliseconds: 200),
              opacity: isConnected ? 0.0 : 1.0,
              child: Container(
                width: double.infinity,
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: [Colors.red.shade700, Colors.red.shade400],
                  ),
                ),
                alignment: Alignment.center,
                child: const Text(
                  'Koneksi server terputus',
                  style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13),
                ),
              ),
            ),
          ),
          Expanded(
            child: IndexedStack(
              index: _currentIndex,
              children: tabs,
            ),
          ),
        ],
      ),
      bottomNavigationBar: Container(
        decoration: BoxDecoration(
          border: Border(
            top: BorderSide(
              color: Colors.white.withValues(alpha: 0.05),
              width: 1,
            ),
          ),
        ),
        child: BottomNavigationBar(
          currentIndex: _currentIndex,
          onTap: (index) {
            HapticFeedback.lightImpact();
            setState(() {
              _currentIndex = index;
            });
          },
          backgroundColor: AppColors.background,
          selectedItemColor: AppColors.primary,
          unselectedItemColor: AppColors.textSecondary,
          selectedLabelStyle: const TextStyle(fontWeight: FontWeight.bold, fontSize: 12),
          unselectedLabelStyle: const TextStyle(fontSize: 11),
          type: BottomNavigationBarType.fixed,
          items: const [
            BottomNavigationBarItem(
              icon: Icon(Icons.dashboard_rounded),
              activeIcon: Icon(Icons.dashboard_rounded, color: AppColors.primary),
              label: 'Utama',
            ),
            BottomNavigationBarItem(
              icon: Icon(Icons.volume_up_rounded),
              activeIcon: Icon(Icons.volume_up_rounded, color: AppColors.primary),
              label: 'Mixer',
            ),
            BottomNavigationBarItem(
              icon: Icon(Icons.settings_system_daydream_rounded),
              activeIcon: Icon(Icons.settings_system_daydream_rounded, color: AppColors.primary),
              label: 'Sistem',
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildInfoRow(String label, String value, {bool isStatus = false, bool statusVal = false}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8.0),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(color: AppColors.textSecondary, fontSize: 14)),
          if (isStatus)
            Row(
              children: [
                Container(
                  width: 8,
                  height: 8,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: statusVal ? Colors.green : Colors.red,
                  ),
                ),
                const SizedBox(width: 6),
                Text(
                  value,
                  style: TextStyle(
                    color: statusVal ? Colors.green : Colors.red,
                    fontWeight: FontWeight.bold,
                    fontSize: 14,
                  ),
                ),
              ],
            )
          else
            Text(
              value,
              style: const TextStyle(color: AppColors.textPrimary, fontWeight: FontWeight.w600, fontSize: 14),
            ),
        ],
      ),
    );
  }
}
