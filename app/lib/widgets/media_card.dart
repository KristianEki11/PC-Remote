import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import '../models/media_state.dart';
import '../utils/theme.dart';
import 'shared_card.dart';

class MediaCard extends StatefulWidget {
  const MediaCard({super.key});

  @override
  State<MediaCard> createState() => _MediaCardState();
}

class _MediaCardState extends State<MediaCard> with SingleTickerProviderStateMixin {
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
  }

  @override
  void dispose() {
    _breathingController.dispose();
    super.dispose();
  }

  Future<void> _executeMediaAction(String actionName, Future<bool> Function() actionCall) async {
    HapticFeedback.mediumImpact();
    final success = await actionCall();
    if (mounted) {
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
    final mediaState = context.watch<MediaState>();
    final mediaStatus = mediaState.currentStatus;
    final isPlaying = mediaStatus.isPlaying;
    final isLoading = mediaState.isLoading;

    // Handle breathing animation based on real play state
    if (isPlaying) {
      if (!_breathingController.isAnimating) {
        _breathingController.repeat(reverse: true);
      }
    } else {
      if (_breathingController.isAnimating) {
        _breathingController.stop();
        _breathingController.animateTo(0.5, duration: const Duration(milliseconds: 500));
      }
    }

    // Determine titles dynamically
    final displayTitle = mediaStatus.title.isNotEmpty ? mediaStatus.title : 'PC Media Output';
    String displayArtist = 'Dihentikan';
    if (isPlaying) {
      displayArtist = mediaStatus.artist.isNotEmpty ? mediaStatus.artist : 'Sedang Diputar';
    } else if (mediaStatus.status == 'Paused') {
      displayArtist = 'Dijeda';
    }

    return SharedCard(
      child: Column(
        children: [
          CardHeader(
            icon: Icons.music_note_rounded,
            title: 'Media Player',
            trailing: isLoading ? const CardLoadingIndicator() : null,
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
          // Track Info
          Text(
            displayTitle,
            style: const TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
            textAlign: TextAlign.center,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
          const SizedBox(height: 4),
          Text(
            displayArtist,
            style: const TextStyle(
              fontSize: 13,
              color: AppColors.textSecondary,
            ),
            textAlign: TextAlign.center,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
          const SizedBox(height: 16),
          // Visualizer bars with fixed height to prevent card jiggling
          SizedBox(
            height: 40,
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.end,
              children: List.generate(5, (index) {
                return _VisualizerBar(
                  isPlaying: isPlaying,
                  index: index,
                );
              }),
            ),
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
                onPressed: isLoading ? null : () => _executeMediaAction('Previous Track', mediaState.prev),
              ),
              const SizedBox(width: 24),
              // Play/Pause Center Button
              GestureDetector(
                onTap: isLoading ? null : () => _executeMediaAction('Play/Pause', mediaState.playPause),
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
                    isPlaying ? Icons.pause_rounded : Icons.play_arrow_rounded,
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
                onPressed: isLoading ? null : () => _executeMediaAction('Next Track', mediaState.next),
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
