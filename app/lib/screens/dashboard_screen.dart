import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../services/api_service.dart';
import '../models/app_state.dart';
import '../widgets/audio_card.dart';
import '../widgets/media_card.dart';
import '../widgets/browser_card.dart';
import '../widgets/system_card.dart';
import '../widgets/fade_in_stagger.dart';
import 'login_screen.dart';
import 'settings_screen.dart';

class DashboardScreen extends StatefulWidget {
  const DashboardScreen({super.key});

  @override
  State<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends State<DashboardScreen> with SingleTickerProviderStateMixin {
  Timer? _pingTimer;
  int _failCount = 0;

  // Animated pulse for connection indicator
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
    super.dispose();
  }

  Future<void> _doPing() async {
    final health = await ApiService.healthCheck();
    if (!mounted) return;
    
    if (health != null && health['status'] == 'ok') {
      _failCount = 0;
      Provider.of<AppState>(context, listen: false).setConnectionStatus(true);
    } else {
      _failCount++;
      if (_failCount >= 3) {
        Provider.of<AppState>(context, listen: false).setConnectionStatus(false);
      }
    }
  }

  void _startPing() {
    _pingTimer = Timer.periodic(const Duration(seconds: 10), (timer) async {
      await _doPing();
    });
    // Initial ping
    _doPing();
  }

  Future<void> _onRefresh() async {
    HapticFeedback.mediumImpact();
    await _doPing();
    // Trigger rebuild of all child cards by using setState
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

  @override
  Widget build(BuildContext context) {
    final isConnected = context.watch<AppState>().isConnected;

    return Scaffold(
      appBar: AppBar(
        title: const Text('PC Remote'),
        actions: [
          // Animated connection indicator with label
          Container(
            margin: const EdgeInsets.only(right: 8),
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
                      width: 10,
                      height: 10,
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
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: isConnected ? Colors.green : Colors.red,
                  ),
                ),
              ],
            ),
          ),
          IconButton(
            icon: const Icon(Icons.settings_outlined),
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(builder: (context) => const SettingsScreen()),
              );
            },
          ),
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: _logout,
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
                  'Server tidak terhubung',
                  style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13),
                ),
              ),
            ),
          ),
          Expanded(
            child: RefreshIndicator(
              onRefresh: _onRefresh,
              color: Theme.of(context).colorScheme.primary,
              child: SingleChildScrollView(
                physics: const AlwaysScrollableScrollPhysics(),
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  children: [
                    FadeInStagger(delayMs: 0, child: AudioCard(key: ValueKey('audio_${DateTime.now().millisecondsSinceEpoch ~/ 10000}'))),
                    const SizedBox(height: 12),
                    const FadeInStagger(delayMs: 100, child: MediaCard()),
                    const SizedBox(height: 12),
                    const FadeInStagger(delayMs: 200, child: BrowserCard()),
                    const SizedBox(height: 12),
                    const FadeInStagger(delayMs: 300, child: SystemCard()),
                    const SizedBox(height: 24),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

