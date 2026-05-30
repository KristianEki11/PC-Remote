import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import '../models/app_state.dart';
import '../models/audio_state.dart';
import '../models/audio_device.dart';
import '../utils/theme.dart';
import 'shared_card.dart';

class AudioCard extends StatefulWidget {
  const AudioCard({super.key});

  @override
  State<AudioCard> createState() => _AudioCardState();
}

class _AudioCardState extends State<AudioCard> {
  Timer? _pollingTimer;
  bool _isDragging = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _fetchInitialData();
      _startPolling();
    });
  }

  @override
  void dispose() {
    _pollingTimer?.cancel();
    super.dispose();
  }

  void _fetchInitialData() {
    if (!mounted) return;
    final appState = Provider.of<AppState>(context, listen: false);
    if (appState.isConnected) {
      final audioState = Provider.of<AudioState>(context, listen: false);
      audioState.fetchMasterVolume();
      audioState.fetchDevices();
    }
  }

  void _startPolling() {
    _pollingTimer = Timer.periodic(const Duration(seconds: 10), (timer) {
      if (!mounted) return;
      final appState = Provider.of<AppState>(context, listen: false);
      if (appState.isConnected && !_isDragging) {
        final audioState = Provider.of<AudioState>(context, listen: false);
        audioState.fetchMasterVolume();
        audioState.fetchDevices();
      }
    });
  }

  Future<void> _toggleMasterMute() async {
    HapticFeedback.mediumImpact();
    final audioState = Provider.of<AudioState>(context, listen: false);
    final success = await audioState.toggleMasterMute();
    if (!mounted) return;
    if (success) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Master mute diubah'),
          backgroundColor: Colors.green,
          duration: Duration(milliseconds: 1500),
        ),
      );
    } else {
      HapticFeedback.heavyImpact();
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Gagal mengubah master mute'),
          backgroundColor: Colors.red,
          duration: Duration(milliseconds: 1500),
        ),
      );
    }
  }

  Future<void> _setMasterVolume(double val) async {
    setState(() => _isDragging = false);
    final audioState = Provider.of<AudioState>(context, listen: false);
    final success = await audioState.setMasterVolume(val);
    if (!success && mounted) {
      HapticFeedback.heavyImpact();
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Gagal mengubah volume master'), backgroundColor: Colors.red),
      );
    }
  }

  Future<void> _toggleDeviceMute(String deviceId, bool targetMute) async {
    HapticFeedback.lightImpact();
    final audioState = Provider.of<AudioState>(context, listen: false);
    final success = await audioState.toggleDeviceMute(deviceId, targetMute);
    if (!success && mounted) {
      HapticFeedback.heavyImpact();
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Gagal mengubah mute perangkat'), backgroundColor: Colors.red),
      );
    }
  }

  Future<void> _setDeviceVolume(String deviceId, double val) async {
    setState(() => _isDragging = false);
    final audioState = Provider.of<AudioState>(context, listen: false);
    final success = await audioState.setDeviceVolume(deviceId, val);
    if (!success && mounted) {
      HapticFeedback.heavyImpact();
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Gagal mengubah volume perangkat'), backgroundColor: Colors.red),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final isConnected = context.watch<AppState>().isConnected;
    final audioState = context.watch<AudioState>();
    final displayDevices = audioState.devices.take(6).toList();

    final masterVolume = audioState.masterVolume;
    final isMasterMuted = audioState.isMasterMuted;
    final isLoadingMaster = audioState.isLoadingMaster;
    final isLoadingDevices = audioState.isLoadingDevices;

    if (!isConnected) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.volume_off_rounded, size: 64, color: AppColors.textSecondary.withValues(alpha: 0.5)),
            const SizedBox(height: 16),
            const Text(
              'Mixer Audio Offline',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: AppColors.textPrimary),
            ),
            const SizedBox(height: 8),
            const Text(
              'Hubungkan server untuk mengontrol volume.',
              style: TextStyle(fontSize: 14, color: AppColors.textSecondary),
            ),
          ],
        ),
      );
    }

    return SingleChildScrollView(
      padding: const EdgeInsets.all(16.0),
      child: Column(
        children: [
          // Master Volume Card
          SharedCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                CardHeader(
                  icon: Icons.volume_up_rounded,
                  title: 'Volume Master',
                  trailing: (isLoadingMaster || isLoadingDevices) ? const CardLoadingIndicator() : null,
                ),
                const SizedBox(height: 24),
                Row(
                  children: [
                    GestureDetector(
                      onTap: isLoadingMaster ? null : _toggleMasterMute,
                      child: Container(
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          color: isMasterMuted
                              ? Colors.red.withValues(alpha: 0.15)
                              : AppColors.primary.withValues(alpha: 0.1),
                          border: Border.all(
                            color: isMasterMuted
                                ? Colors.red.withValues(alpha: 0.3)
                                : AppColors.primary.withValues(alpha: 0.2),
                          ),
                        ),
                        child: Icon(
                          isMasterMuted ? Icons.volume_off_rounded : Icons.volume_up_rounded,
                          color: isMasterMuted ? Colors.red : AppColors.primary,
                          size: 26,
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: SliderTheme(
                        data: SliderTheme.of(context).copyWith(
                          trackHeight: 6.0,
                          thumbShape: const RoundSliderThumbShape(enabledThumbRadius: 10.0),
                          overlayShape: const RoundSliderOverlayShape(overlayRadius: 20.0),
                          activeTrackColor: AppColors.primary,
                          inactiveTrackColor: AppColors.surfaceLight,
                        ),
                        child: Slider(
                          value: masterVolume,
                          min: 0,
                          max: 100,
                          divisions: 100,
                          onChanged: (val) {
                            setState(() => _isDragging = true);
                            audioState.updateMasterVolumeLocally(val);
                          },
                          onChangeEnd: _setMasterVolume,
                        ),
                      ),
                    ),
                    Container(
                      width: 50,
                      alignment: Alignment.centerRight,
                      child: Text(
                        '${masterVolume.toInt()}%',
                        style: const TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: AppColors.textPrimary,
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),
          // Output Devices Card
          SharedCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const CardHeader(
                  icon: Icons.speaker_group_rounded,
                  title: 'Saluran Output',
                ),
                const SizedBox(height: 18),
                if (displayDevices.isEmpty)
                  Padding(
                    padding: const EdgeInsets.symmetric(vertical: 24.0),
                    child: Center(
                      child: Text(
                        'Tidak ada perangkat audio ditemukan',
                        style: TextStyle(color: AppColors.textSecondary.withValues(alpha: 0.7)),
                      ),
                    ),
                  )
                else
                  Column(
                    children: List.generate(displayDevices.length, (index) {
                      final AudioDevice device = displayDevices[index];
                      final deviceId = device.id;
                      final name = device.name;
                      final double level = device.volume;
                      final isMuted = device.muted;

                      return Container(
                        margin: const EdgeInsets.only(bottom: 12.0),
                        padding: const EdgeInsets.all(12.0),
                        decoration: BoxDecoration(
                          color: AppColors.surfaceLight.withValues(alpha: 0.3),
                          borderRadius: BorderRadius.circular(14),
                          border: Border.all(color: Colors.white.withValues(alpha: 0.03)),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              children: [
                                const Icon(Icons.audiotrack_rounded, size: 16, color: AppColors.textSecondary),
                                const SizedBox(width: 8),
                                Expanded(
                                  child: Text(
                                    name,
                                    style: const TextStyle(
                                      fontSize: 13,
                                      fontWeight: FontWeight.w600,
                                      color: AppColors.textPrimary,
                                    ),
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                ),
                                GestureDetector(
                                  onTap: () => _toggleDeviceMute(deviceId, !isMuted),
                                  child: Icon(
                                    isMuted ? Icons.volume_off_rounded : Icons.volume_up_rounded,
                                    color: isMuted ? Colors.redAccent : AppColors.textSecondary,
                                    size: 20,
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 8),
                            Row(
                              children: [
                                Expanded(
                                  child: SliderTheme(
                                    data: SliderTheme.of(context).copyWith(
                                      trackHeight: 4.0,
                                      thumbShape: const RoundSliderThumbShape(enabledThumbRadius: 8.0),
                                      overlayShape: const RoundSliderOverlayShape(overlayRadius: 16.0),
                                      activeTrackColor: Colors.blueAccent,
                                      inactiveTrackColor: AppColors.surfaceLight,
                                    ),
                                    child: Slider(
                                      value: level,
                                      min: 0,
                                      max: 100,
                                      divisions: 100,
                                      onChanged: (val) {
                                        setState(() => _isDragging = true);
                                        audioState.updateDeviceVolumeLocally(deviceId, val);
                                      },
                                      onChangeEnd: (val) => _setDeviceVolume(deviceId, val),
                                    ),
                                  ),
                                ),
                                const SizedBox(width: 8),
                                SizedBox(
                                  width: 36,
                                  child: Text(
                                    '${level.toInt()}%',
                                    style: const TextStyle(
                                      fontSize: 12,
                                      fontWeight: FontWeight.w500,
                                      color: AppColors.textSecondary,
                                    ),
                                    textAlign: TextAlign.right,
                                  ),
                                ),
                              ],
                            ),
                          ],
                        ),
                      );
                    }),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
