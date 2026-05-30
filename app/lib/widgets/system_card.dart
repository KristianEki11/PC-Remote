import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../services/api_service.dart';
import 'shared_card.dart';
import 'scale_button.dart';

class SystemCard extends StatefulWidget {
  const SystemCard({super.key});

  @override
  State<SystemCard> createState() => _SystemCardState();
}

class _SystemCardState extends State<SystemCard> {
  bool _isProcessing = false;

  Future<void> _executeAction(String action, Future<bool> Function() apiCall) async {
    HapticFeedback.mediumImpact();
    setState(() => _isProcessing = true);
    final success = await apiCall();

    await Future.delayed(const Duration(milliseconds: 500));

    if (mounted) {
      setState(() => _isProcessing = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(success ? '$action berhasil' : 'Gagal mengeksekusi $action'),
          backgroundColor: success ? Colors.green : Colors.red,
        ),
      );
    }
  }

  void _showConfirmationDialog(String action, Future<bool> Function() apiCall) {
    HapticFeedback.heavyImpact();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text('Konfirmasi $action', style: TextStyle(color: Theme.of(context).colorScheme.onSurface)),
        content: Text('Yakin ingin $action PC?', style: TextStyle(color: Theme.of(context).colorScheme.onSurface)),
        backgroundColor: Theme.of(context).cardColor,
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: Text('Batal', style: TextStyle(color: Theme.of(context).colorScheme.onSurface)),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(ctx);
              _executeAction(action, apiCall);
            },
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
            child: Text('Ya, lanjutkan', style: TextStyle(color: Theme.of(context).colorScheme.onError)),
          ),
        ],
      ),
    );
  }

  void _showShutdownDialog() {
    HapticFeedback.heavyImpact();
    int selectedPresetIndex = 0; // 0: Sekarang, 1: 1 Jam, 2: 3 Jam, 3: 5 Jam, 4: Manual
    String customUnit = 'Menit';
    final customController = TextEditingController();

    showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (dialogCtx, setDialogState) {
          return AlertDialog(
            title: Row(
              children: [
                const Icon(Icons.power_settings_new, color: Colors.red),
                const SizedBox(width: 8),
                Text(
                  'Shutdown PC',
                  style: TextStyle(color: Theme.of(dialogCtx).colorScheme.onSurface),
                ),
              ],
            ),
            content: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Pilih durasi tunda sebelum PC mati:',
                    style: TextStyle(
                      color: Theme.of(dialogCtx).colorScheme.onSurface.withValues(alpha: 0.7),
                      fontSize: 14,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: [
                      _buildPresetChip(0, 'Sekarang', selectedPresetIndex, (val) => setDialogState(() => selectedPresetIndex = val), dialogCtx),
                      _buildPresetChip(1, '1 Jam', selectedPresetIndex, (val) => setDialogState(() => selectedPresetIndex = val), dialogCtx),
                      _buildPresetChip(2, '3 Jam', selectedPresetIndex, (val) => setDialogState(() => selectedPresetIndex = val), dialogCtx),
                      _buildPresetChip(3, '5 Jam', selectedPresetIndex, (val) => setDialogState(() => selectedPresetIndex = val), dialogCtx),
                      _buildPresetChip(4, 'Manual', selectedPresetIndex, (val) => setDialogState(() => selectedPresetIndex = val), dialogCtx),
                    ],
                  ),
                  if (selectedPresetIndex == 4) ...[
                    const SizedBox(height: 16),
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.center,
                      children: [
                        Expanded(
                          flex: 3,
                          child: TextField(
                            controller: customController,
                            keyboardType: TextInputType.number,
                            inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                            decoration: const InputDecoration(
                              labelText: 'Durasi',
                              border: OutlineInputBorder(),
                              contentPadding: EdgeInsets.symmetric(horizontal: 10, vertical: 8),
                            ),
                            style: TextStyle(color: Theme.of(dialogCtx).colorScheme.onSurface),
                          ),
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          flex: 2,
                          child: Container(
                            padding: const EdgeInsets.symmetric(horizontal: 8),
                            decoration: BoxDecoration(
                              border: Border.all(color: Colors.grey.withValues(alpha: 0.5)),
                              borderRadius: BorderRadius.circular(4),
                            ),
                            child: DropdownButtonHideUnderline(
                              child: DropdownButton<String>(
                                value: customUnit,
                                items: const [
                                  DropdownMenuItem(value: 'Detik', child: Text('Detik')),
                                  DropdownMenuItem(value: 'Menit', child: Text('Menit')),
                                  DropdownMenuItem(value: 'Jam', child: Text('Jam')),
                                ],
                                onChanged: (val) {
                                  if (val != null) {
                                    setDialogState(() => customUnit = val);
                                  }
                                },
                                style: TextStyle(color: Theme.of(dialogCtx).colorScheme.onSurface),
                                dropdownColor: Theme.of(dialogCtx).cardColor,
                              ),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ],
                ],
              ),
            ),
            backgroundColor: Theme.of(dialogCtx).cardColor,
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(ctx),
                child: Text('Batal', style: TextStyle(color: Theme.of(dialogCtx).colorScheme.onSurface)),
              ),
              ElevatedButton(
                onPressed: () {
                  int seconds = 0;
                  if (selectedPresetIndex == 0) {
                    seconds = 0;
                  } else if (selectedPresetIndex == 1) {
                    seconds = 3600;
                  } else if (selectedPresetIndex == 2) {
                    seconds = 10800;
                  } else if (selectedPresetIndex == 3) {
                    seconds = 18000;
                  } else if (selectedPresetIndex == 4) {
                    final int val = int.tryParse(customController.text) ?? 0;
                    if (customUnit == 'Detik') {
                      seconds = val;
                    } else if (customUnit == 'Menit') {
                      seconds = val * 60;
                    } else if (customUnit == 'Jam') {
                      seconds = val * 3600;
                    }
                  }
                  Navigator.pop(ctx);
                  _executeAction('Shutdown', () => ApiService.shutdownPc(delaySeconds: seconds));
                },
                style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
                child: const Text('Ya, lanjutkan', style: TextStyle(color: Colors.white)),
              ),
            ],
          );
        },
      ),
    );
  }

  Widget _buildPresetChip(
    int index,
    String label,
    int selectedIndex,
    void Function(int) onSelected,
    BuildContext dialogCtx,
  ) {
    final isSelected = index == selectedIndex;
    return ChoiceChip(
      label: Text(
        label,
        style: TextStyle(
          color: isSelected 
              ? Colors.white
              : Theme.of(dialogCtx).colorScheme.onSurface,
          fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
        ),
      ),
      selected: isSelected,
      selectedColor: Colors.red,
      backgroundColor: Theme.of(dialogCtx).colorScheme.surface.withValues(alpha: 0.5),
      onSelected: (selected) {
        if (selected) {
          onSelected(index);
        }
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return SharedCard(
      child: Column(
        children: [
          const CardHeader(icon: Icons.settings_rounded, title: 'Kontrol Sistem'),
          const SizedBox(height: 16),
          GridView.count(
            crossAxisCount: 2,
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            mainAxisSpacing: 12,
            crossAxisSpacing: 12,
            childAspectRatio: 2.3,
            children: [
              _buildButton('Lock PC', Icons.lock_outline_rounded, Colors.blueAccent, () => _executeAction('Lock PC', ApiService.lockPc)),
              _buildButton('Sleep', Icons.bedtime_rounded, Colors.purpleAccent, () => _executeAction('Sleep', ApiService.sleepPc)),
              _buildButton('Restart', Icons.restart_alt_rounded, Colors.orangeAccent, () => _showConfirmationDialog('Restart', ApiService.restartPc)),
              _buildButton('Shutdown', Icons.power_settings_new_rounded, Colors.redAccent, () => _showShutdownDialog()),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildButton(String label, IconData icon, Color color, VoidCallback? onPressed) {
    return ScaleButtonWrapper(
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: _isProcessing ? null : onPressed,
          borderRadius: BorderRadius.circular(12),
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
            decoration: BoxDecoration(
              color: color.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: color.withValues(alpha: 0.25), width: 1.5),
            ),
            child: Row(
              children: [
                Icon(icon, color: color, size: 24),
                const SizedBox(width: 10),
                Text(
                  label,
                  style: const TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 13,
                    color: Colors.white,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
