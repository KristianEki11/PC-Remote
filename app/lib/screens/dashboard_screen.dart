import 'dart:async';
import 'dart:ui';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
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
    await ApiService.logout();
    
    if (!mounted) return;
    Provider.of<AppState>(context, listen: false).clear();
    try {
      Provider.of<MediaState>(context, listen: false).stopPolling();
    } catch (_) {}
    
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
                        fontSize: 28,
                        fontWeight: FontWeight.w800,
                        color: AppColors.textPrimary,
                        letterSpacing: -1,
                      ),
                    ),
                    const SizedBox(height: 6),
                    Row(
                      children: [
                        Expanded(
                          child: Text(
                            'IP PC: ${appState.ipAddress}',
                            style: const TextStyle(
                              fontSize: 14,
                              color: AppColors.textSecondary,
                              fontWeight: FontWeight.w500,
                            ),
                            overflow: TextOverflow.ellipsis,
                            maxLines: 1,
                          ),
                        ),
                        const SizedBox(width: 8),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                          decoration: BoxDecoration(
                            color: ApiService.isUsingPublicTunnel
                                ? const Color(0xFFF59E0B).withOpacity(0.15)
                                : const Color(0xFF10B981).withOpacity(0.15),
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(
                              color: ApiService.isUsingPublicTunnel
                                  ? const Color(0xFFF59E0B).withOpacity(0.5)
                                  : const Color(0xFF10B981).withOpacity(0.5),
                              width: 1,
                            ),
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Icon(
                                ApiService.isUsingPublicTunnel ? Icons.public : Icons.wifi,
                                size: 14,
                                color: ApiService.isUsingPublicTunnel
                                    ? const Color(0xFFD97706)
                                    : const Color(0xFF059669),
                              ),
                              const SizedBox(width: 6),
                              Text(
                                ApiService.isUsingPublicTunnel ? 'Internet (4G/WAN)' : 'WiFi Lokal',
                                style: TextStyle(
                                  fontSize: 12,
                                  fontWeight: FontWeight.w600,
                                  color: ApiService.isUsingPublicTunnel
                                      ? const Color(0xFFD97706)
                                      : const Color(0xFF059669),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 20),
              const FadeInStagger(delayMs: 0, child: MediaCard()),
              const SizedBox(height: 20),
              const FadeInStagger(delayMs: 100, child: BrowserCard()),
              const SizedBox(height: 32),
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
            const SizedBox(height: 20),
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
                    const SizedBox(height: 20),
                    _buildInfoRow('IP Address', appState.ipAddress),
                    _buildInfoRow('Platform', 'Windows'),
                    _buildInfoRow('Status Koneksi', isConnected ? 'Online' : 'Offline', isStatus: true, statusVal: isConnected),
                    _buildInfoRow('Versi Server', 'v4.0.0'),
                    const SizedBox(height: 24),
                    // Action Buttons inside card
                    Row(
                      children: [
                        Expanded(
                          child: OutlinedButton.icon(
                            onPressed: () {
                              Navigator.push(
                                context,
                                MaterialPageRoute(builder: (context) => const SettingsScreen()),
                              );
                            },
                            style: OutlinedButton.styleFrom(
                              foregroundColor: AppColors.primary,
                              side: BorderSide(color: AppColors.primary.withOpacity(0.5), width: 1.5),
                              padding: const EdgeInsets.symmetric(vertical: 14),
                              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                            ),
                            icon: const Icon(Icons.settings_outlined, size: 20),
                            label: const Text('Pengaturan', style: TextStyle(fontWeight: FontWeight.w600)),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: ElevatedButton.icon(
                            onPressed: _logout,
                            style: ElevatedButton.styleFrom(
                              backgroundColor: Colors.red.withOpacity(0.1),
                              foregroundColor: Colors.redAccent,
                              elevation: 0,
                              padding: const EdgeInsets.symmetric(vertical: 14),
                              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                            ),
                            icon: const Icon(Icons.logout_rounded, size: 20),
                            label: const Text('Keluar', style: TextStyle(fontWeight: FontWeight.w600)),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 32),
          ],
        ),
      ),
    ];

    return Scaffold(
      extendBodyBehindAppBar: true,
      extendBody: true,
      appBar: AppBar(
        title: const Text('PC Remote'),
        flexibleSpace: ClipRRect(
          child: BackdropFilter(
            filter: ImageFilter.blur(sigmaX: 10, sigmaY: 10),
            child: Container(color: Colors.white.withOpacity(0.3)),
          ),
        ),
        actions: [
          // Animated connection status badge - liquid glass
          Container(
            margin: const EdgeInsets.only(right: 16),
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(0.6),
              borderRadius: BorderRadius.circular(20),
              border: Border.all(color: Colors.white, width: 1.5),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                AnimatedBuilder(
                  animation: _pulseAnimation,
                  builder: (context, child) {
                    return Container(
                      width: 10,
                      height: 10,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: isConnected
                            ? AppColors.success.withOpacity(_pulseAnimation.value)
                            : AppColors.error,
                        boxShadow: isConnected
                            ? [BoxShadow(color: AppColors.success.withOpacity(0.4), blurRadius: 6, spreadRadius: 1)]
                            : [],
                      ),
                    );
                  },
                ),
                const SizedBox(width: 8),
                Text(
                  isConnected ? 'Online' : 'Offline',
                  style: TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w700,
                    color: isConnected ? AppColors.success : AppColors.error,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
      body: Container(
        decoration: const BoxDecoration(
          gradient: AppGradients.glassBackground,
        ),
        child: Column(
          children: [
            SizedBox(height: MediaQuery.of(context).padding.top + kToolbarHeight),
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
                  color: AppColors.error.withOpacity(0.9),
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
      ),
      // Liquid glass bottom navigation bar
      bottomNavigationBar: ClipRRect(
        child: BackdropFilter(
          filter: ImageFilter.blur(sigmaX: 20, sigmaY: 20),
          child: Container(
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(0.5),
              border: Border(top: BorderSide(color: Colors.white.withOpacity(0.6), width: 1.5)),
            ),
            child: BottomNavigationBar(
              currentIndex: _currentIndex,
              onTap: (index) {
                HapticFeedback.lightImpact();
                setState(() {
                  _currentIndex = index;
                });
              },
              backgroundColor: Colors.transparent,
              elevation: 0,
              selectedItemColor: AppColors.primary,
              unselectedItemColor: AppColors.textMuted,
              selectedLabelStyle: const TextStyle(fontWeight: FontWeight.w700, fontSize: 12),
              unselectedLabelStyle: const TextStyle(fontSize: 11, fontWeight: FontWeight.w500),
              type: BottomNavigationBarType.fixed,
              items: const [
                BottomNavigationBarItem(
                  icon: Padding(padding: EdgeInsets.only(bottom: 4), child: Icon(Icons.dashboard_outlined)),
                  activeIcon: Padding(padding: EdgeInsets.only(bottom: 4), child: Icon(Icons.dashboard_rounded)),
                  label: 'Utama',
                ),
                BottomNavigationBarItem(
                  icon: Padding(padding: EdgeInsets.only(bottom: 4), child: Icon(Icons.tune_outlined)),
                  activeIcon: Padding(padding: EdgeInsets.only(bottom: 4), child: Icon(Icons.tune_rounded)),
                  label: 'Mixer',
                ),
                BottomNavigationBarItem(
                  icon: Padding(padding: EdgeInsets.only(bottom: 4), child: Icon(Icons.desktop_windows_outlined)),
                  activeIcon: Padding(padding: EdgeInsets.only(bottom: 4), child: Icon(Icons.desktop_windows_rounded)),
                  label: 'Sistem',
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildInfoRow(String label, String value, {bool isStatus = false, bool statusVal = false}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 10.0),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(color: AppColors.textMuted, fontSize: 14, fontWeight: FontWeight.w500)),
          if (isStatus)
            Row(
              children: [
                Container(
                  width: 8,
                  height: 8,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: statusVal ? AppColors.success : AppColors.error,
                  ),
                ),
                const SizedBox(width: 8),
                Text(
                  value,
                  style: TextStyle(
                    color: statusVal ? AppColors.success : AppColors.error,
                    fontWeight: FontWeight.w700,
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
