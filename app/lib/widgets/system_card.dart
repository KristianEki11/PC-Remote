import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../services/api_service.dart';
import '../utils/theme.dart';
import '../utils/globals.dart';
import 'shared_card.dart';
import 'scale_button.dart';

class SystemCard extends StatefulWidget {
  const SystemCard({super.key});

  @override
  State<SystemCard> createState() => _SystemCardState();
}

class _SystemCardState extends State<SystemCard> {
  bool _isProcessing = false;
  double _brightnessLevel = 50.0;
  bool _isLoadingBrightness = true;

  @override
  void initState() {
    super.initState();
    _fetchBrightness();
  }

  Future<void> _fetchBrightness() async {
    final val = await ApiService.getBrightness();
    if (mounted && val != null) {
      setState(() {
        _brightnessLevel = val.clamp(0.0, 100.0);
        _isLoadingBrightness = false;
      });
    } else if (mounted) {
      setState(() => _isLoadingBrightness = false);
    }
  }

  Future<void> _updateBrightness(double val) async {
    // Snap to 50% if within ±3% range (47.0 - 53.0)
    if (val >= 47.0 && val <= 53.0) {
      val = 50.0;
    }
    setState(() {
      _brightnessLevel = val.clamp(0.0, 100.0);
    });
    HapticFeedback.selectionClick();
    await ApiService.setBrightness(_brightnessLevel);
  }

  void _stepBrightness(double delta) {
    double next = _brightnessLevel + delta;
    if ((_brightnessLevel < 50.0 && next >= 48.0 && next <= 52.0) ||
        (_brightnessLevel > 50.0 && next >= 48.0 && next <= 52.0)) {
      next = 50.0;
    }
    _updateBrightness(next.clamp(0.0, 100.0));
  }

  Future<void> _executeAction(String action, Future<bool> Function() apiCall) async {
    HapticFeedback.mediumImpact();
    setState(() => _isProcessing = true);
    final success = await apiCall();

    await Future.delayed(const Duration(milliseconds: 500));

    if (mounted) {
      setState(() => _isProcessing = false);
      snackbarKey.currentState?.showSnackBar(
        SnackBar(
          content: Text(success ? '$action berhasil' : 'Gagal mengeksekusi $action'),
          backgroundColor: success ? Colors.green : Colors.red,
          behavior: SnackBarBehavior.floating,
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
    final isDimmed = _brightnessLevel < 50.0;
    final displayPct = _brightnessLevel.toInt();

    return SharedCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const CardHeader(icon: Icons.settings_rounded, title: 'Kontrol Sistem'),
          const SizedBox(height: 16),

          // Brightness Control Section
          Container(
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: AppColors.surfaceLight,
              borderRadius: BorderRadius.circular(16),
              boxShadow: AppClays.card(),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Row(
                      children: [
                        Icon(
                          isDimmed
                              ? (_brightnessLevel == 0 ? Icons.brightness_2 : Icons.nightlight_round)
                              : Icons.wb_sunny_rounded,
                          size: 20,
                          color: isDimmed ? const Color(0xFF818CF8) : const Color(0xFFFBBF24),
                        ),
                        const SizedBox(width: 8),
                        Text(
                          'Kecerahan Layar',
                          style: TextStyle(
                            fontWeight: FontWeight.w600,
                            fontSize: 14,
                            color: Theme.of(context).colorScheme.onSurface,
                          ),
                        ),
                      ],
                    ),
                    Row(
                      children: [
                        if (!_isLoadingBrightness && _brightnessLevel != 50.0) ...[
                          GestureDetector(
                            onTap: () => _updateBrightness(50.0),
                            child: Container(
                              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                              decoration: BoxDecoration(
                                color: Theme.of(context).colorScheme.primary.withValues(alpha: 0.12),
                                borderRadius: BorderRadius.circular(6),
                                border: Border.all(
                                  color: Theme.of(context).colorScheme.primary.withValues(alpha: 0.3),
                                  width: 1,
                                ),
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Icon(
                                    Icons.restart_alt_rounded,
                                    size: 12,
                                    color: Theme.of(context).colorScheme.primary,
                                  ),
                                  const SizedBox(width: 2),
                                  Text(
                                    '50%',
                                    style: TextStyle(
                                      fontSize: 11,
                                      fontWeight: FontWeight.bold,
                                      color: Theme.of(context).colorScheme.primary,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                          const SizedBox(width: 6),
                        ],
                        _isLoadingBrightness
                            ? const SizedBox(
                                width: 14,
                                height: 14,
                                child: CircularProgressIndicator(strokeWidth: 2),
                              )
                            : Text(
                                isDimmed
                                    ? '$displayPct% (OLED -${((50 - _brightnessLevel) * 2).toInt()}%)'
                                    : '$displayPct% (HW ${((_brightnessLevel - 50) * 2).toInt()}%)',
                                style: TextStyle(
                                  fontSize: 12,
                                  fontWeight: FontWeight.w600,
                                  color: isDimmed ? const Color(0xFF818CF8) : AppColors.primary,
                                ),
                              ),
                      ],
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    // Button Minus 5%
                    Material(
                      color: Colors.transparent,
                      child: InkWell(
                        onTap: () => _stepBrightness(-5.0),
                        borderRadius: BorderRadius.circular(8),
                        child: Container(
                          padding: const EdgeInsets.all(6),
                          decoration: BoxDecoration(
                            color: Theme.of(context).colorScheme.surface.withValues(alpha: 0.5),
                            borderRadius: BorderRadius.circular(8),
                            border: Border.all(
                              color: Theme.of(context).colorScheme.onSurface.withValues(alpha: 0.1),
                            ),
                          ),
                          child: Icon(
                            Icons.remove_rounded,
                            size: 18,
                            color: Theme.of(context).colorScheme.onSurface,
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(width: 4),
                    Expanded(
                      child: SliderTheme(
                        data: SliderTheme.of(context).copyWith(
                          trackHeight: 6.0,
                          thumbShape: const RoundSliderThumbShape(enabledThumbRadius: 10.0),
                          overlayShape: const RoundSliderOverlayShape(overlayRadius: 20.0),
                          activeTrackColor: isDimmed ? const Color(0xFF818CF8) : const Color(0xFFFBBF24),
                          inactiveTrackColor: Theme.of(context).colorScheme.surface.withValues(alpha: 0.3),
                        ),
                        child: Slider(
                          value: _brightnessLevel,
                          min: 0,
                          max: 100,
                          divisions: 100,
                          onChanged: (val) {
                            // Magnetic snap to 50% during drag
                            if (val >= 47.0 && val <= 53.0) {
                              val = 50.0;
                            }
                            setState(() {
                              _brightnessLevel = val;
                            });
                          },
                          onChangeEnd: _updateBrightness,
                        ),
                      ),
                    ),
                    const SizedBox(width: 4),
                    // Button Plus 5%
                    Material(
                      color: Colors.transparent,
                      child: InkWell(
                        onTap: () => _stepBrightness(5.0),
                        borderRadius: BorderRadius.circular(8),
                        child: Container(
                          padding: const EdgeInsets.all(6),
                          decoration: BoxDecoration(
                            color: Theme.of(context).colorScheme.surface.withValues(alpha: 0.5),
                            borderRadius: BorderRadius.circular(8),
                            border: Border.all(
                              color: Theme.of(context).colorScheme.onSurface.withValues(alpha: 0.1),
                            ),
                          ),
                          child: Icon(
                            Icons.add_rounded,
                            size: 18,
                            color: Theme.of(context).colorScheme.onSurface,
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),

          Column(
            children: [
              Row(
                children: [
                  Expanded(child: _buildButton('Lock PC', Icons.lock_outline_rounded, Colors.blueAccent, () => _executeAction('Lock PC', ApiService.lockPc))),
                  const SizedBox(width: 12),
                  Expanded(child: _buildButton('Sleep', Icons.bedtime_rounded, Colors.purpleAccent, () => _executeAction('Sleep', ApiService.sleepPc))),
                ],
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  Expanded(child: _buildButton('Display Off', Icons.monitor_outlined, const Color(0xFF26C6DA), () => _executeAction('Display Off', ApiService.displayOff))),
                  const SizedBox(width: 12),
                  Expanded(child: _buildButton('Restart', Icons.restart_alt_rounded, Colors.orangeAccent, () => _showConfirmationDialog('Restart', ApiService.restartPc))),
                ],
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  Expanded(child: _buildButton('Shutdown', Icons.power_settings_new_rounded, Colors.redAccent, () => _showShutdownDialog())),
                ],
              ),
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
          borderRadius: BorderRadius.circular(16),
          splashColor: color.withValues(alpha: 0.2),
          highlightColor: color.withValues(alpha: 0.1),
          onTap: _isProcessing ? null : onPressed,
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
            decoration: BoxDecoration(
              color: AppColors.surfaceLight,
              borderRadius: BorderRadius.circular(16),
              boxShadow: AppClays.button(),
            ),
            child: Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: color.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Icon(icon, color: color, size: 22),
                ),
                const SizedBox(width: 12),
                Text(
                  label,
                  style: TextStyle(
                    fontWeight: FontWeight.w600,
                    fontSize: 14,
                    color: color,
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
