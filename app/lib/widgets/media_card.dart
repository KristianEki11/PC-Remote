import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../services/api_service.dart';
import '../utils/theme.dart';
import 'shared_card.dart';

class MediaCard extends StatefulWidget {
  const MediaCard({super.key});

  @override
  State<MediaCard> createState() => _MediaCardState();
}

class _MediaCardState extends State<MediaCard> with SingleTickerProviderStateMixin {
  bool _isLoading = false;
  bool _isPlaying = false; // local visual toggle
  late AnimationController _breathingController;
  late Animation<double> _breathingAnimation;

  @override
  void initState() {
    super.initState();
    _breathingController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 2500),
    );
    _breathingAnimation = Tween<double>(begin: 0.96, end: 1.04).animate(
      CurvedAnimation(parent: _breathingController, curve: Curves.easeInOut),
    );
    _breathingController.repeat(reverse: true);
  }

  @override
  void dispose() {
    _breathingController.dispose();
    super.dispose();
  }

  Future<void> _executeMediaAction(String actionName, Future<bool> Function() apiCall) async {
    HapticFeedback.mediumImpact();
    setState(() => _isLoading = true);
    final success = await apiCall();
    
    if (mounted) {
      setState(() {
        _isLoading = false;
        if (actionName == 'Play/Pause' && success) {
          _isPlaying = !_isPlaying;
        }
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(success ? '$actionName berhasil' : 'Gagal mengeksekusi $actionName'),
          backgroundColor: success ? Colors.green : Colors.red,
          duration: const Duration(milliseconds: 1500),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return SharedCard(
      child: Column(
        children: [
          CardHeader(
            icon: Icons.music_note_rounded,
            title: 'Media Player',
            trailing: _isLoading ? const CardLoadingIndicator() : null,
          ),
          const SizedBox(height: 24),
          // Cover art with breathing animation & glow
          ScaleTransition(
            scale: _breathingAnimation,
            child: Container(
              width: 130,
              height: 130,
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(24),
                gradient: const LinearGradient(
                  colors: [Color(0xFF8A2387), Color(0xFFE94057), Color(0xFFF27121)],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
                boxShadow: [
                  BoxShadow(
                    color: const Color(0xFFE94057).withValues(alpha: 0.35),
                    blurRadius: 20,
                    spreadRadius: 2,
                    offset: const Offset(0, 8),
                  ),
                ],
              ),
              child: Stack(
                alignment: Alignment.center,
                children: [
                  // Subtle rotating-like vinyl lines background
                  Container(
                    margin: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      border: Border.all(color: Colors.white.withValues(alpha: 0.1), width: 1.5),
                    ),
                  ),
                  Container(
                    margin: const EdgeInsets.all(28),
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      border: Border.all(color: Colors.white.withValues(alpha: 0.1), width: 1.5),
                    ),
                  ),
                  const Icon(
                    Icons.music_note_rounded,
                    size: 54,
                    color: Colors.white,
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 20),
          // Track Info Placeholder
          const Text(
            'PC Media Output',
            style: TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            _isPlaying ? 'Sedang Diputar' : 'Dihentikan',
            style: const TextStyle(
              fontSize: 13,
              color: AppColors.textSecondary,
            ),
          ),
          const SizedBox(height: 16),
          // Visualizer bars
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: List.generate(5, (index) {
              return _VisualizerBar(
                isPlaying: _isPlaying,
                index: index,
              );
            }),
          ),
          const SizedBox(height: 20),
          // Player Controls
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              // Prev Button
              _buildControlButton(
                icon: Icons.skip_previous_rounded,
                iconSize: 32,
                color: AppColors.textPrimary.withValues(alpha: 0.8),
                onPressed: _isLoading ? null : () => _executeMediaAction('Previous Track', ApiService.mediaPrev),
              ),
              const SizedBox(width: 24),
              // Play/Pause Center Button
              GestureDetector(
                onTap: _isLoading ? null : () => _executeMediaAction('Play/Pause', ApiService.mediaPlayPause),
                child: Container(
                  width: 68,
                  height: 68,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    gradient: AppGradients.accent,
                    boxShadow: [
                      BoxShadow(
                        color: AppColors.primary.withValues(alpha: 0.4),
                        blurRadius: 16,
                        spreadRadius: 2,
                        offset: const Offset(0, 4),
                      ),
                    ],
                  ),
                  child: Icon(
                    _isPlaying ? Icons.pause_rounded : Icons.play_arrow_rounded,
                    size: 38,
                    color: Colors.white,
                  ),
                ),
              ),
              const SizedBox(width: 24),
              // Next Button
              _buildControlButton(
                icon: Icons.skip_next_rounded,
                iconSize: 32,
                color: AppColors.textPrimary.withValues(alpha: 0.8),
                onPressed: _isLoading ? null : () => _executeMediaAction('Next Track', ApiService.mediaNext),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildControlButton({
    required IconData icon,
    required double iconSize,
    required Color color,
    required VoidCallback? onPressed,
  }) {
    return Container(
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        color: AppColors.surfaceLight.withValues(alpha: 0.5),
        border: Border.all(color: Colors.white.withValues(alpha: 0.05)),
      ),
      child: IconButton(
        iconSize: iconSize,
        color: color,
        disabledColor: color.withValues(alpha: 0.3),
        icon: Icon(icon),
        onPressed: onPressed,
      ),
    );
  }
}

// Custom Visualizer Bar Widget
class _VisualizerBar extends StatefulWidget {
  final bool isPlaying;
  final int index;

  const _VisualizerBar({required this.isPlaying, required this.index});

  @override
  State<_VisualizerBar> createState() => _VisualizerBarState();
}

class _VisualizerBarState extends State<_VisualizerBar> with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _animation;

  static const List<double> _heights = [24.0, 36.0, 18.0, 30.0, 20.0];
  static const List<int> _durations = [600, 800, 500, 700, 650];

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: Duration(milliseconds: _durations[widget.index % 5]),
    );
    _animation = Tween<double>(begin: 4.0, end: _heights[widget.index % 5]).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );

    if (widget.isPlaying) {
      _controller.repeat(reverse: true);
    }
  }

  @override
  void didUpdateWidget(covariant _VisualizerBar oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.isPlaying) {
      _controller.repeat(reverse: true);
    } else {
      _controller.stop();
      _controller.animateTo(0.0, duration: const Duration(milliseconds: 300));
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _animation,
      builder: (context, child) {
        return Container(
          margin: const EdgeInsets.symmetric(horizontal: 3),
          width: 4,
          height: _animation.value,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(2),
            gradient: const LinearGradient(
              colors: [AppColors.primary, AppColors.gradientEnd],
              begin: Alignment.bottomCenter,
              end: Alignment.topCenter,
            ),
          ),
        );
      },
    );
  }
}
