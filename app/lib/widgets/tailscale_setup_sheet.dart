import 'package:flutter/material.dart';
import '../services/api_service.dart';
import '../services/companion_service.dart';
import '../utils/theme.dart';

Future<void> showTailscaleSetupSheet(BuildContext context,
    {VoidCallback? onUpdated}) {
  return showModalBottomSheet(
    context: context,
    isScrollControlled: true,
    backgroundColor: Colors.transparent,
    builder: (context) => TailscaleSetupSheet(onUpdated: onUpdated),
  );
}

class TailscaleSetupSheet extends StatefulWidget {
  final VoidCallback? onUpdated;

  const TailscaleSetupSheet({super.key, this.onUpdated});

  @override
  State<TailscaleSetupSheet> createState() => _TailscaleSetupSheetState();
}

class _TailscaleSetupSheetState extends State<TailscaleSetupSheet> {
  bool _isSyncing = false;
  bool? _testPassed;
  String? _statusNote;

  String _pcStateLabel(String state) {
    switch (state) {
      case 'Running':
        return 'Tailscale PC Aktif';
      case 'NeedsLogin':
        return 'Perlu Login di PC';
      case 'Stopped':
        return 'Tailscale PC Berhenti';
      case 'NotInstalled':
      case 'Unavailable':
        return 'Belum Terpasang di PC';
      default:
        return 'Belum Tersinkron ($state)';
    }
  }

  Color _pcStateColor(String state) {
    switch (state) {
      case 'Running':
        return const Color(0xFF059669);
      case 'NeedsLogin':
      case 'Stopped':
        return const Color(0xFFD97706);
      default:
        return AppColors.textSecondary;
    }
  }

  Future<void> _syncAndTest() async {
    setState(() {
      _isSyncing = true;
      _testPassed = null;
      _statusNote = 'Menyinkronkan endpoint dari PC...';
    });

    await ApiService.syncConnectionEndpoints(refresh: true);
    final hasUrl = ApiService.isTailscaleSynced;
    bool reachable = false;
    if (hasUrl) {
      setState(() {
        _statusNote = 'Menguji koneksi ke ${ApiService.tailscaleUrl}...';
      });
      reachable = await ApiService.testTailscaleConnection();
    }

    if (!mounted) return;
    setState(() {
      _isSyncing = false;
      _testPassed = reachable;
      if (!hasUrl) {
        _statusNote =
            'Alamat Tailscale PC belum tersedia. Pastikan Tailscale di PC sudah login dan aktif.';
      } else if (reachable) {
        _statusNote =
            'Koneksi luar rumah berhasil! (${ApiService.tailscaleUrl})';
      } else {
        _statusNote =
            'Alamat Tailscale tersimpan (${ApiService.tailscaleUrl}), namun belum terjangkau dari HP. Pastikan aplikasi Tailscale di HP aktif dan memakai akun yang sama.';
      }
    });
    widget.onUpdated?.call();
  }

  @override
  Widget build(BuildContext context) {
    final pcState = ApiService.tailscaleState;
    final tsUrl = ApiService.tailscaleUrl;

    return Container(
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: const BorderRadius.vertical(top: Radius.circular(24)),
        boxShadow: AppClays.card(),
      ),
      padding: EdgeInsets.only(
        left: 20,
        right: 20,
        top: 20,
        bottom: MediaQuery.of(context).viewInsets.bottom + 24,
      ),
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text(
                  'Wizard Akses Luar Rumah',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.w800,
                    color: AppColors.textPrimary,
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.close, color: AppColors.textSecondary),
                  onPressed: () => Navigator.pop(context),
                ),
              ],
            ),
            const SizedBox(height: 6),
            const Text(
              'Gunakan Tailscale pribadi Anda agar HP tetap terhubung ke PC dari kampus/kantor (4G/5G) tanpa scan QR ulang.',
              style: TextStyle(fontSize: 13, color: AppColors.textSecondary),
            ),
            const SizedBox(height: 16),

            // Step 1: PC Tailscale Status
            _StepTile(
              step: '1',
              title: 'Status Tailscale di PC',
              subtitle: tsUrl.isNotEmpty
                  ? '${_pcStateLabel(pcState)} • $tsUrl'
                  : _pcStateLabel(pcState),
              accentColor: _pcStateColor(pcState),
            ),
            const SizedBox(height: 12),

            // Step 2: Android Tailscale App
            _StepTile(
              step: '2',
              title: 'Pasang & Aktifkan Tailscale di HP',
              subtitle:
                  'Login dengan akun Tailscale yang sama dengan PC Anda.',
              accentColor: AppColors.primary,
              action: Column(
                children: [
                  SizedBox(
                    width: double.infinity,
                    child: OutlinedButton.icon(
                      onPressed: () => CompanionService.openTailscaleApp(),
                      icon: const Icon(Icons.open_in_new_rounded, size: 18),
                      label:
                          const Text('Buka / Install Tailscale (Play Store)'),
                      style: OutlinedButton.styleFrom(
                        foregroundColor: AppColors.primary,
                        side: BorderSide(
                            color: AppColors.primary.withValues(alpha: 0.4)),
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12)),
                      ),
                    ),
                  ),
                  const SizedBox(height: 8),
                  SizedBox(
                    width: double.infinity,
                    child: OutlinedButton.icon(
                      onPressed: () => CompanionService.downloadTailscaleApk(),
                      icon: const Icon(Icons.download_rounded, size: 18),
                      label: const Text('Direct Download APK Tailscale HP'),
                      style: OutlinedButton.styleFrom(
                        foregroundColor: const Color(0xFF059669),
                        side: BorderSide(
                            color: const Color(0xFF10B981)
                                .withValues(alpha: 0.5)),
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12)),
                      ),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 12),

            // Step 3: Sync & Test
            _StepTile(
              step: '3',
              title: 'Sinkronkan & Tes Koneksi Luar Rumah',
              subtitle:
                  'Ambil alamat Tailscale terbaru dari PC dan verifikasi koneksi.',
              accentColor: AppColors.success,
              action: SizedBox(
                width: double.infinity,
                child: ElevatedButton.icon(
                  onPressed: _isSyncing ? null : _syncAndTest,
                  icon: _isSyncing
                      ? const SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(
                              strokeWidth: 2, color: Colors.white),
                        )
                      : const Icon(Icons.sync_rounded, size: 18),
                  label: Text(_isSyncing
                      ? 'Menguji Koneksi...'
                      : 'Sinkronkan & Tes Koneksi Luar Rumah'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppColors.primary,
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12)),
                  ),
                ),
              ),
            ),

            if (_statusNote != null) ...[
              const SizedBox(height: 14),
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: (_testPassed == true
                          ? const Color(0xFF10B981)
                          : (_testPassed == false
                              ? const Color(0xFFF59E0B)
                              : AppColors.primary))
                      .withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(
                    color: (_testPassed == true
                            ? const Color(0xFF10B981)
                            : (_testPassed == false
                                ? const Color(0xFFF59E0B)
                                : AppColors.primary))
                        .withValues(alpha: 0.35),
                  ),
                ),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Icon(
                      _testPassed == true
                          ? Icons.check_circle_outline
                          : (_testPassed == false
                              ? Icons.warning_amber_rounded
                              : Icons.info_outline),
                      size: 18,
                      color: _testPassed == true
                          ? const Color(0xFF059669)
                          : (_testPassed == false
                              ? const Color(0xFFD97706)
                              : AppColors.primary),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        _statusNote!,
                        style: const TextStyle(
                            fontSize: 13, color: AppColors.textPrimary),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _StepTile extends StatelessWidget {
  final String step;
  final String title;
  final String subtitle;
  final Color accentColor;
  final Widget? action;

  const _StepTile({
    required this.step,
    required this.title,
    required this.subtitle,
    required this.accentColor,
    this.action,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: AppColors.surfaceLight,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFE5E5E5)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              CircleAvatar(
                radius: 12,
                backgroundColor: accentColor.withValues(alpha: 0.15),
                child: Text(
                  step,
                  style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                      color: accentColor),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  title,
                  style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w700,
                      color: AppColors.textPrimary),
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Text(
            subtitle,
            style:
                const TextStyle(fontSize: 12.5, color: AppColors.textSecondary),
          ),
          if (action != null) ...[
            const SizedBox(height: 10),
            action!,
          ],
        ],
      ),
    );
  }
}
