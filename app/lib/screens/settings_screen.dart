import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../services/api_service.dart';
import '../models/app_state.dart';
import '../utils/theme.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  final _formKey = GlobalKey<FormState>();
  final _currentPinController = TextEditingController();
  final _newPinController = TextEditingController();
  final _confirmPinController = TextEditingController();
  bool _isLoading = false;
  String? _errorMessage;

  @override
  void dispose() {
    _currentPinController.dispose();
    _newPinController.dispose();
    _confirmPinController.dispose();
    super.dispose();
  }

  void _showChangePinBottomSheet() {
    _currentPinController.clear();
    _newPinController.clear();
    _confirmPinController.clear();
    setState(() {
      _errorMessage = null;
    });

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: AppColors.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setModalState) {
            Future<void> handleChangePin() async {
              if (!_formKey.currentState!.validate()) return;

              final currentPin = _currentPinController.text.trim();
              final newPin = _newPinController.text.trim();
              final navigator = Navigator.of(context);
              final scaffoldMessenger = ScaffoldMessenger.of(this.context);

              setModalState(() {
                _isLoading = true;
                _errorMessage = null;
              });

              final error = await ApiService.changePIN(currentPin, newPin);

              if (error == null) {
                navigator.pop(); // close bottom sheet
                scaffoldMessenger.showSnackBar(
                  const SnackBar(
                    content: Text('PIN server berhasil diubah!'),
                    backgroundColor: AppColors.success,
                  ),
                );
              } else {
                setModalState(() {
                  _isLoading = false;
                  _errorMessage = error;
                });
              }
            }

            return Padding(
              padding: EdgeInsets.only(
                left: 20,
                right: 20,
                top: 20,
                bottom: MediaQuery.of(context).viewInsets.bottom + 20,
              ),
              child: Form(
                key: _formKey,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text(
                          'Ganti PIN Server',
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                            color: AppColors.textPrimary,
                          ),
                        ),
                        IconButton(
                          icon: const Icon(Icons.close, color: AppColors.textSecondary),
                          onPressed: () => Navigator.pop(context),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    TextFormField(
                      controller: _currentPinController,
                      obscureText: true,
                      keyboardType: TextInputType.number,
                      maxLength: 8,
                      decoration: const InputDecoration(
                        labelText: 'PIN Sekarang',
                        prefixIcon: Icon(Icons.lock_outline, color: AppColors.textSecondary),
                        counterText: '',
                      ),
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return 'PIN sekarang tidak boleh kosong';
                        }
                        if (value.length < 4 || value.length > 8) {
                          return 'PIN harus 4-8 digit';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 12),
                    TextFormField(
                      controller: _newPinController,
                      obscureText: true,
                      keyboardType: TextInputType.number,
                      maxLength: 8,
                      decoration: const InputDecoration(
                        labelText: 'PIN Baru',
                        prefixIcon: Icon(Icons.lock_reset, color: AppColors.textSecondary),
                        counterText: '',
                      ),
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return 'PIN baru tidak boleh kosong';
                        }
                        if (value.length < 4 || value.length > 8) {
                          return 'PIN harus 4-8 digit';
                        }
                        if (value == _currentPinController.text) {
                          return 'PIN baru tidak boleh sama dengan PIN sekarang';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 12),
                    TextFormField(
                      controller: _confirmPinController,
                      obscureText: true,
                      keyboardType: TextInputType.number,
                      maxLength: 8,
                      decoration: const InputDecoration(
                        labelText: 'Konfirmasi PIN Baru',
                        prefixIcon: Icon(Icons.lock_clock, color: AppColors.textSecondary),
                        counterText: '',
                      ),
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return 'Konfirmasi PIN baru tidak boleh kosong';
                        }
                        if (value != _newPinController.text) {
                          return 'Konfirmasi PIN baru tidak cocok';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 16),
                    if (_errorMessage != null) ...[
                      Container(
                        width: double.infinity,
                        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                        decoration: BoxDecoration(
                          color: AppColors.error.withValues(alpha: 0.1),
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(color: AppColors.error.withValues(alpha: 0.3)),
                        ),
                        child: Row(
                          children: [
                            const Icon(Icons.error_outline, color: AppColors.error, size: 18),
                            const SizedBox(width: 8),
                            Expanded(
                              child: Text(
                                _errorMessage!,
                                style: const TextStyle(color: AppColors.error, fontSize: 13),
                              ),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(height: 16),
                    ],
                    SizedBox(
                      width: double.infinity,
                      height: 50,
                      child: Container(
                        decoration: BoxDecoration(
                          gradient: _isLoading ? null : AppGradients.accent,
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: ElevatedButton(
                          onPressed: _isLoading ? null : handleChangePin,
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.transparent,
                            shadowColor: Colors.transparent,
                            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                          ),
                          child: _isLoading
                              ? const SizedBox(
                                  width: 20,
                                  height: 20,
                                  child: CircularProgressIndicator(
                                    color: Colors.white,
                                    strokeWidth: 2,
                                  ),
                                )
                              : const Text(
                                  'Simpan PIN Baru',
                                  style: TextStyle(
                                    fontSize: 15,
                                    fontWeight: FontWeight.w600,
                                    color: Colors.white,
                                  ),
                                ),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            );
          },
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final ipAddress = context.watch<AppState>().ipAddress;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Pengaturan'),
      ),
      body: ListView(
        children: [
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 16, 16, 8),
            child: Text(
              'KEAMANAN',
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: AppColors.textSecondary,
                letterSpacing: 1.0,
              ),
            ),
          ),
          ListTile(
            leading: const Icon(Icons.lock_outline, color: AppColors.primary),
            title: const Text('Ganti PIN Server'),
            subtitle: const Text('Ubah PIN akses kontrol PC'),
            trailing: const Icon(Icons.chevron_right, color: AppColors.textSecondary),
            onTap: _showChangePinBottomSheet,
          ),
          const Divider(),
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 16, 16, 8),
            child: Text(
              'INFORMASI SERVER',
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: AppColors.textSecondary,
                letterSpacing: 1.0,
              ),
            ),
          ),
          ListTile(
            leading: const Icon(Icons.dns_outlined, color: AppColors.primary),
            title: const Text('IP Server Terhubung'),
            subtitle: Text(ipAddress),
          ),
          const Divider(),
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 16, 16, 8),
            child: Text(
              'TENTANG',
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: AppColors.textSecondary,
                letterSpacing: 1.0,
              ),
            ),
          ),
          const ListTile(
            leading: Icon(Icons.info_outline, color: AppColors.primary),
            title: Text('Versi Aplikasi'),
            subtitle: Text('v2.2.5'),
          ),
        ],
      ),
    );
  }
}
