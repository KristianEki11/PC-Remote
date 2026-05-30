import 'package:flutter/material.dart';
import '../services/api_service.dart';

class AudioState extends ChangeNotifier {
  double _masterVolume = 50.0;
  bool _isMasterMuted = false;
  bool _isLoadingMaster = false;

  List<dynamic> _devices = [];
  bool _isLoadingDevices = false;

  double get masterVolume => _masterVolume;
  bool get isMasterMuted => _isMasterMuted;
  bool get isLoadingMaster => _isLoadingMaster;

  List<dynamic> get devices => _devices;
  bool get isLoadingDevices => _isLoadingDevices;

  Future<void> fetchMasterVolume() async {
    _isLoadingMaster = true;
    notifyListeners();
    final data = await ApiService.getVolume();
    if (data != null) {
      // Server returns 0.0–1.0, slider uses 0–100
      _masterVolume = ((data['level'] as num).toDouble() * 100).clamp(0.0, 100.0);
      _isMasterMuted = data['muted'] as bool;
    }
    _isLoadingMaster = false;
    notifyListeners();
  }

  Future<void> fetchDevices() async {
    _isLoadingDevices = true;
    notifyListeners();
    final data = await ApiService.getAudioDevices();
    if (data != null) {
      _devices = List<dynamic>.from(data);
    }
    _isLoadingDevices = false;
    notifyListeners();
  }

  /// Optimistic update for setting master volume level
  Future<bool> setMasterVolume(double level) async {
    final originalVolume = _masterVolume;
    _masterVolume = level;
    notifyListeners();

    // Server expects 0.0–1.0 float; slider gives 0–100
    final success = await ApiService.setVolume(level / 100.0);
    if (!success) {
      _masterVolume = originalVolume;
      notifyListeners();
      return false;
    }
    return true;
  }

  /// Optimistic update for toggling master mute state
  Future<bool> toggleMasterMute() async {
    final originalMute = _isMasterMuted;
    final newMute = !_isMasterMuted;
    _isMasterMuted = newMute;
    notifyListeners();

    // Pass explicit new mute state — server requires body {muted: bool}
    final success = await ApiService.setMute(newMute);
    if (!success) {
      _isMasterMuted = originalMute;
      notifyListeners();
      return false;
    }
    return true;
  }

  /// Optimistic update for setting individual device volume level
  Future<bool> setDeviceVolume(String deviceId, double level) async {
    final index = _devices.indexWhere((d) => d['id'] == deviceId);
    if (index == -1) return false;

    final originalVolume = (_devices[index]['volume'] as num).toDouble();
    _devices[index]['volume'] = level;
    notifyListeners();

    final success = await ApiService.setDeviceVolume(deviceId, level);
    if (!success) {
      _devices[index]['volume'] = originalVolume;
      notifyListeners();
      return false;
    }
    return true;
  }

  /// Optimistic update for toggling individual device mute state
  Future<bool> toggleDeviceMute(String deviceId, bool targetMute) async {
    final index = _devices.indexWhere((d) => d['id'] == deviceId);
    if (index == -1) return false;

    final originalMuted = _devices[index]['muted'] as bool;
    _devices[index]['muted'] = targetMute;
    notifyListeners();

    final success = await ApiService.toggleDeviceMute(deviceId, targetMute);
    if (!success) {
      _devices[index]['muted'] = originalMuted;
      notifyListeners();
      return false;
    }
    return true;
  }

  /// Local immediate state updates for smooth sliding (does not call API)
  void updateMasterVolumeLocally(double level) {
    _masterVolume = level;
    notifyListeners();
  }

  /// Local immediate state updates for smooth sliding (does not call API)
  void updateDeviceVolumeLocally(String deviceId, double level) {
    final index = _devices.indexWhere((d) => d['id'] == deviceId);
    if (index != -1) {
      _devices[index]['volume'] = level;
      notifyListeners();
    }
  }
}
