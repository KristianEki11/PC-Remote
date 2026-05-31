import 'package:flutter/material.dart';
import '../services/api_service.dart';
import '../utils/theme.dart';
import 'shared_card.dart';

class BrowserCard extends StatefulWidget {
  const BrowserCard({super.key});

  @override
  State<BrowserCard> createState() => _BrowserCardState();
}

class _BrowserCardState extends State<BrowserCard> {
  final TextEditingController _urlController = TextEditingController();
  bool _isLoading = false;
  int _selectedPresetIndex = -1;

  final List<Map<String, dynamic>> _presets = [
    {
      'name': 'YouTube',
      'url': 'youtube.com',
      'icon': Icons.play_circle_fill_rounded,
      'color': const Color(0xFFFF0000),
    },
    {
      'name': 'Netflix',
      'url': 'netflix.com',
      'icon': Icons.movie_filter_rounded,
      'color': const Color(0xFFE50914),
    },
    {
      'name': 'Twitch',
      'url': 'twitch.tv',
      'icon': Icons.live_tv_rounded,
      'color': const Color(0xFF9146FF),
    },
    {
      'name': 'Spotify',
      'url': 'open.spotify.com',
      'icon': Icons.library_music_rounded,
      'color': const Color(0xFF1DB954),
    },
    {
      'name': 'Google',
      'url': 'google.com',
      'icon': Icons.travel_explore_rounded,
      'color': const Color(0xFF4285F4),
    },
    {
      'name': 'ChatGPT',
      'url': 'chatgpt.com',
      'icon': Icons.auto_awesome_rounded,
      'color': const Color(0xFF10A37F),
    },
  ];

  Future<void> _openBrowser() async {
    String url = _urlController.text.trim();
    if (url.isEmpty || !url.contains('.')) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Masukkan domain yang valid, contoh: youtube.com'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    if (!url.startsWith('http://') && !url.startsWith('https://')) {
      url = 'https://$url';
    }

    setState(() => _isLoading = true);
    final success = await ApiService.openBrowser(url);
    if (mounted) {
      setState(() => _isLoading = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(success ? 'Berhasil dibuka di browser PC' : 'Gagal membuka URL'),
          backgroundColor: success ? Colors.green : Colors.red,
          duration: const Duration(seconds: 2),
        ),
      );
    }
  }

  void _selectPreset(int index) {
    setState(() {
      _selectedPresetIndex = index;
      _urlController.text = _presets[index]['url'] as String;
    });
  }

  @override
  Widget build(BuildContext context) {
    return SharedCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const CardHeader(
            icon: Icons.explore_rounded,
            title: 'Web Launcher',
          ),
          const SizedBox(height: 18),
          const Text(
            'Akses Cepat',
            style: TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.w600,
              color: AppColors.textSecondary,
            ),
          ),
          const SizedBox(height: 12),
          // Presets Grid
          GridView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: _presets.length,
            gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 3,
              crossAxisSpacing: 10,
              mainAxisSpacing: 10,
              childAspectRatio: 1.15,
            ),
            itemBuilder: (context, index) {
              final preset = _presets[index];
              final isSelected = _selectedPresetIndex == index;
              final accentColor = preset['color'] as Color;

              return InkWell(
                onTap: () => _selectPreset(index),
                borderRadius: BorderRadius.circular(14),
                child: Container(
                  decoration: BoxDecoration(
                    color: isSelected
                        ? accentColor.withValues(alpha: 0.15)
                        : AppColors.surfaceLight,
                    borderRadius: BorderRadius.circular(14),
                    boxShadow: AppClays.button(intensity: isSelected ? 0.8 : 0.4),
                  ),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(
                        preset['icon'] as IconData,
                        color: isSelected ? accentColor : AppColors.textSecondary,
                        size: 28,
                      ),
                      const SizedBox(height: 8),
                      Text(
                        preset['name'] as String,
                        style: TextStyle(
                          fontSize: 12,
                          fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                          color: isSelected ? Colors.white : AppColors.textSecondary,
                        ),
                      ),
                    ],
                  ),
                ),
              );
            },
          ),
          const SizedBox(height: 20),
          // Custom URL Input
          TextField(
            controller: _urlController,
            keyboardType: TextInputType.url,
            style: const TextStyle(fontSize: 14),
            decoration: InputDecoration(
              hintText: 'Masukkan URL manual, misal: google.com',
              prefixIcon: const Icon(Icons.link_rounded, color: AppColors.textSecondary),
              suffixIcon: _urlController.text.isNotEmpty
                  ? IconButton(
                      icon: const Icon(Icons.clear_rounded, color: AppColors.textSecondary),
                      onPressed: () {
                        setState(() {
                          _urlController.clear();
                          _selectedPresetIndex = -1;
                        });
                      },
                    )
                  : null,
            ),
            onChanged: (val) {
              // check if match preset
              int matchIndex = _presets.indexWhere((p) => p['url'] == val.trim());
              setState(() {
                _selectedPresetIndex = matchIndex;
              });
            },
          ),
          const SizedBox(height: 16),
          // Launch Button - claymorphic
          SizedBox(
            width: double.infinity,
            height: 52,
            child: Container(
              decoration: BoxDecoration(
                gradient: _urlController.text.isEmpty || _isLoading ? null : AppGradients.primaryButton,
                borderRadius: BorderRadius.circular(16),
                boxShadow: _urlController.text.isEmpty || _isLoading ? [] : AppClays.button(),
              ),
              child: ElevatedButton(
                onPressed: _urlController.text.isEmpty || _isLoading ? null : _openBrowser,
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.transparent,
                  shadowColor: Colors.transparent,
                  foregroundColor: Colors.white,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                ),
                child: _isLoading
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2),
                      )
                    : const Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.open_in_new_rounded, size: 18),
                          SizedBox(width: 8),
                          Text(
                            'Buka di PC',
                            style: TextStyle(fontWeight: FontWeight.bold, letterSpacing: 0.5),
                          ),
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
