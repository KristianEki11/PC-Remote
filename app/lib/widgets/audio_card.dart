import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import '../models/audio_state.dart';
import 'shared_card.dart';

class AudioCard extends StatefulWidget {
  const AudioCard({super.key});

  @override
  State<AudioCard> createState() => _AudioCardState();
}

class _AudioCardState extends State<AudioCard> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
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
        const SnackBar(content: Text('Master mute diubah'), backgroundColor: Colors.green),
      );
    } else {
      HapticFeedback.heavyImpact();
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Gagal mengubah master mute'), backgroundColor: Colors.red),
      );
    }
  }

  Future<void> _setMasterVolume(double val) async {
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
    } else if (success && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Mute perangkat diubah'), backgroundColor: Colors.green),
      );
    }
  }

  Future<void> _setDeviceVolume(String deviceId, double val) async {
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
    final audioState = context.watch<AudioState>();
    final displayDevices = audioState.devices.take(6).toList();
    final bool isScrollable = displayDevices.length > 4;

    final masterVolume = audioState.masterVolume;
    final isMasterMuted = audioState.isMasterMuted;
    final isLoadingMaster = audioState.isLoadingMaster;
    final isLoadingDevices = audioState.isLoadingDevices;

    return SharedCard(
      child: Column(
        children: [
          CardHeader(
            icon: Icons.volume_up,
            title: 'Audio Master',
            trailing: (isLoadingMaster || isLoadingDevices) ? const CardLoadingIndicator() : null,
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              IconButton(
                icon: Icon(
                  isMasterMuted ? Icons.volume_off : Icons.volume_up,
                  color: isMasterMuted ? Colors.red : Theme.of(context).colorScheme.onSurface,
                ),
                onPressed: isLoadingMaster ? null : _toggleMasterMute,
              ),
              Expanded(
                child: Slider(
                  value: masterVolume,
                  min: 0,
                  max: 100,
                  divisions: 100,
                  activeColor: Colors.blue,
                  onChanged: (val) => audioState.updateMasterVolumeLocally(val),
                  onChangeEnd: _setMasterVolume,
                ),
              ),
              SizedBox(
                width: 40,
                child: Text('${masterVolume.toInt()}%'),
              ),
            ],
          ),
          if (displayDevices.isNotEmpty) ...[
            const Divider(height: 24),
            Align(
              alignment: Alignment.centerLeft,
              child: Text(
                'Devices',
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: Theme.of(context).colorScheme.onSurface.withValues(alpha: 0.7),
                ),
              ),
            ),
            const SizedBox(height: 12),
            Container(
              constraints: isScrollable ? const BoxConstraints(maxHeight: 220) : null,
              child: isScrollable
                  ? SingleChildScrollView(child: _buildDeviceList(displayDevices, audioState))
                  : _buildDeviceList(displayDevices, audioState),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildDeviceList(List<dynamic> devices, AudioState audioState) {
    return Column(
      children: List.generate(devices.length, (index) {
        final device = devices[index];
        final deviceId = device['id'] as String;
        final name = device['name'] as String;
        final double level = (device['volume'] as num).toDouble();
        final isMuted = device['muted'] as bool;

        return Card(
          margin: const EdgeInsets.only(bottom: 8.0),
          color: Theme.of(context).colorScheme.onSurface.withValues(alpha: 0.05),
          elevation: 0,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12.0, vertical: 8.0),
            child: Row(
              children: [
                Expanded(
                  flex: 3,
                  child: Text(
                    name,
                    style: TextStyle(fontSize: 12, color: Theme.of(context).colorScheme.onSurface),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                Expanded(
                  flex: 4,
                  child: SliderTheme(
                    data: SliderTheme.of(context).copyWith(
                      trackHeight: 2.0,
                      thumbShape: const RoundSliderThumbShape(enabledThumbRadius: 6.0),
                      overlayShape: const RoundSliderOverlayShape(overlayRadius: 14.0),
                    ),
                    child: Slider(
                      value: level,
                      min: 0,
                      max: 100,
                      divisions: 100,
                      activeColor: Colors.blueAccent,
                      inactiveColor: Theme.of(context).colorScheme.onSurface.withValues(alpha: 0.24),
                      onChanged: (val) => audioState.updateDeviceVolumeLocally(deviceId, val),
                      onChangeEnd: (val) => _setDeviceVolume(deviceId, val),
                    ),
                  ),
                ),
                IconButton(
                  icon: Icon(
                    isMuted ? Icons.volume_off : Icons.volume_up,
                    color: isMuted ? Colors.redAccent : Theme.of(context).colorScheme.onSurface.withValues(alpha: 0.7),
                    size: 20,
                  ),
                  onPressed: () => _toggleDeviceMute(deviceId, !isMuted),
                  constraints: const BoxConstraints(),
                  padding: EdgeInsets.zero,
                ),
              ],
            ),
          ),
        );
      }),
    );
  }
}
