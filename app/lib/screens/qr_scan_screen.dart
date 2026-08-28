import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:provider/provider.dart';
import '../models/app_state.dart';
import '../services/api_service.dart';
import '../utils/theme.dart';
import 'dashboard_screen.dart';

class QRScanScreen extends StatefulWidget {
  const QRScanScreen({super.key});

  @override
  State<QRScanScreen> createState() => _QRScanScreenState();
}

class _QRScanScreenState extends State<QRScanScreen> {
  final MobileScannerController _controller = MobileScannerController(
    detectionSpeed: DetectionSpeed.normal,
    facing: CameraFacing.back,
    torchEnabled: false,
  );

  bool _isProcessing = false;
  bool _isTorchOn = false;
  String? _statusMessage;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Future<void> _handleBarcode(BarcodeCapture capture) async {
    if (_isProcessing) return;

    final barcodes = capture.barcodes;
    if (barcodes.isEmpty) return;

    final rawValue = barcodes.first.rawValue;
    if (rawValue == null || rawValue.isEmpty) return;

    setState(() {
      _isProcessing = true;
      _statusMessage = 'Menghubungkan ke PC...';
    });

    try {
      // 1. Parse JSON payload from QR code
      final Map<String, dynamic> data = jsonDecode(rawValue);

      final host = data['host'] as String?;
      final port = data['port']?.toString() ?? '8000';
      final pairToken = data['pair_token'] as String?;
      final protocol = data['protocol'] as String? ?? 'https';
      final fingerprint = data['fingerprint'] as String? ?? '';
      final serverName = data['server_name'] as String? ?? 'PC Remote';
      final alternateHosts = (data['alternate_hosts'] as List<dynamic>?)?.map((e) => e.toString()).toList() ?? [];

      if (host == null || host.isEmpty || pairToken == null || pairToken.isEmpty) {
        throw const FormatException('Format QR code tidak valid untuk PC Remote.');
      }

      // 2. Perform one-time token pairing exchange with primary host
      var token = await ApiService.pairWithQR(
        host: host,
        port: port,
        pairToken: pairToken,
        protocol: protocol,
        fingerprint: fingerprint,
        serverName: serverName,
      );

      // 3. If primary host fails, automatically try alternate LAN hosts
      String connectedHost = host;
      if (token == null && alternateHosts.isNotEmpty) {
        for (final altHost in alternateHosts) {
          debugPrint('Primary host ($host) failed. Trying alternate host: $altHost');
          token = await ApiService.pairWithQR(
            host: altHost,
            port: port,
            pairToken: pairToken,
            protocol: protocol,
            fingerprint: fingerprint,
            serverName: serverName,
          );
          if (token != null) {
            connectedHost = altHost;
            break;
          }
        }
      }

      if (!mounted) return;

      if (token != null) {
        // Success: update AppState and navigate to Dashboard
        final fullHost = '$connectedHost:$port';
        Provider.of<AppState>(context, listen: false).setConnectionDetails(fullHost, token);

        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Row(
              children: [
                const Icon(Icons.check_circle, color: Colors.greenAccent),
                const SizedBox(width: 8),
                Text('Berhasil terhubung ke $serverName!'),
              ],
            ),
            backgroundColor: const Color(0xFF1E293B),
            behavior: SnackBarBehavior.floating,
          ),
        );

        Navigator.pushAndRemoveUntil(
          context,
          MaterialPageRoute(builder: (context) => const DashboardScreen()),
          (route) => false,
        );
      } else {
        setState(() {
          _isProcessing = false;
          _statusMessage = null;
        });

        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Gagal pairing: Token kedaluwarsa atau server tidak dapat dijangkau.'),
            backgroundColor: Colors.redAccent,
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isProcessing = false;
        _statusMessage = null;
      });

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('QR Code tidak valid: $e'),
          backgroundColor: Colors.redAccent,
          behavior: SnackBarBehavior.floating,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new_rounded, color: AppColors.textPrimary),
          onPressed: () => Navigator.of(context).pop(),
        ),
        title: const Text(
          'Pindai QR Code',
          style: TextStyle(
            color: AppColors.textPrimary,
            fontWeight: FontWeight.w700,
            fontSize: 20,
          ),
        ),
        actions: [
          // Torch toggle
          IconButton(
            icon: Icon(
              _isTorchOn ? Icons.flash_on_rounded : Icons.flash_off_rounded,
              color: _isTorchOn ? Colors.amber : AppColors.textSecondary,
            ),
            onPressed: () async {
              try {
                await _controller.toggleTorch();
                setState(() {
                  _isTorchOn = !_isTorchOn;
                });
              } catch (_) {}
            },
          ),
          // Camera switch
          IconButton(
            icon: const Icon(Icons.cameraswitch_rounded, color: AppColors.textSecondary),
            onPressed: () => _controller.switchCamera(),
          ),
        ],
      ),
      body: Stack(
        alignment: Alignment.center,
        children: [
          // Camera scanner
          MobileScanner(
            controller: _controller,
            onDetect: _handleBarcode,
          ),

          // Custom dark overlay with scanning cutout
          ColorFiltered(
            colorFilter: ColorFilter.mode(
              Colors.black.withValues(alpha: 0.65),
              BlendMode.srcOut,
            ),
            child: Stack(
              fit: StackFit.expand,
              children: [
                Container(
                  decoration: const BoxDecoration(
                    color: Colors.black,
                    backgroundBlendMode: BlendMode.dstOut,
                  ),
                ),
                Align(
                  alignment: Alignment.center,
                  child: Container(
                    width: 260,
                    height: 260,
                    decoration: BoxDecoration(
                      color: Colors.red,
                      borderRadius: BorderRadius.circular(24),
                    ),
                  ),
                ),
              ],
            ),
          ),

          // Scan area framing border
          Container(
            width: 260,
            height: 260,
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(24),
              border: Border.all(
                color: AppColors.primary.withValues(alpha: 0.8),
                width: 3,
              ),
            ),
          ),

          // Instructions at the bottom
          Positioned(
            bottom: 60,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
              decoration: BoxDecoration(
                color: AppColors.surface.withValues(alpha: 0.9),
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: Colors.white.withValues(alpha: 0.1)),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.3),
                    blurRadius: 16,
                  ),
                ],
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (_isProcessing) ...[
                    const SizedBox(
                      width: 24,
                      height: 24,
                      child: CircularProgressIndicator(
                        strokeWidth: 2.5,
                        color: AppColors.primary,
                      ),
                    ),
                    const SizedBox(height: 10),
                    Text(
                      _statusMessage ?? 'Memproses...',
                      style: const TextStyle(
                        color: AppColors.textPrimary,
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ] else ...[
                    const Icon(Icons.qr_code_scanner_rounded, color: AppColors.primary, size: 28),
                    const SizedBox(height: 8),
                    const Text(
                      'Arahkan kamera ke QR Code di PC',
                      style: TextStyle(
                        color: AppColors.textPrimary,
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      'Buka QR dari System Tray di Taskbar PC Anda',
                      style: TextStyle(
                        color: AppColors.textSecondary.withValues(alpha: 0.8),
                        fontSize: 12,
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
