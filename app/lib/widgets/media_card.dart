import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../services/api_service.dart';
import 'shared_card.dart';

class MediaCard extends StatefulWidget {
  const MediaCard({super.key});

  @override
  State<MediaCard> createState() => _MediaCardState();
}

class _MediaCardState extends State<MediaCard> {
  bool _isLoading = false;

  Future<void> _executeMediaAction(String actionName, Future<bool> Function() apiCall) async {
    HapticFeedback.mediumImpact();
    setState(() => _isLoading = true);
    final success = await apiCall();
    
    if (mounted) {
      setState(() => _isLoading = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(success ? '$actionName berhasil' : 'Gagal mengeksekusi $actionName'),
          backgroundColor: success ? Colors.green : Colors.red,
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
            icon: Icons.music_note,
            title: 'Media',
            trailing: _isLoading ? const CardLoadingIndicator() : null,
          ),
          const SizedBox(height: 16),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              IconButton(
                iconSize: 36,
                color: Theme.of(context).colorScheme.onSurface,
                icon: const Icon(Icons.skip_previous),
                onPressed: _isLoading ? null : () => _executeMediaAction('Previous Track', ApiService.mediaPrev),
              ),
              const SizedBox(width: 16),
              IconButton(
                iconSize: 48,
                color: Theme.of(context).colorScheme.primary,
                icon: const Icon(Icons.play_circle_filled),
                onPressed: _isLoading ? null : () => _executeMediaAction('Play/Pause', ApiService.mediaPlayPause),
              ),
              const SizedBox(width: 16),
              IconButton(
                iconSize: 36,
                color: Theme.of(context).colorScheme.onSurface,
                icon: const Icon(Icons.skip_next),
                onPressed: _isLoading ? null : () => _executeMediaAction('Next Track', ApiService.mediaNext),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
