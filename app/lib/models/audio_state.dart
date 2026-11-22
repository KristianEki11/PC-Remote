import 'package:flutter/material.dart';
import '../services/api_service.dart';
import 'audio_device.dart';
import '../utils/volume_helpers.dart';

class AudioState extends ChangeNotifier {
  double _masterVolume = 50.0;
  double _confirmedVolume = 50.0;
  bool _commandPending = false;
  bool _disposed = false;
  int _revision = 0;

  @override
  void notifyListeners() {
    if (!_disposed) super.notifyListeners();
  }

  @override
  void dispose() {
    _disposed = true;
    super.dispose();
  }

  final Map<String, double> _confirmedDevices = {};
  bool _isMasterMuted = false;
  bool _isLoadingMaster = false;

  List<AudioDevice> _devices = [];
  bool _isLoadingDevices = false;

  double get masterVolume => _masterVolume;
  bool get isMasterMuted => _isMasterMuted;
  bool get isLoadingMaster => _isLoadingMaster;

  List<AudioDevice> get devices => _devices;
  bool get isLoadingDevices => _isLoadingDevices;

  Future<void> fetchMasterVolume() async {
    if (_isLoadingMaster || _commandPending) return;
    final revision = _revision;
    _isLoadingMaster = true;
    notifyListeners();
    final data = await ApiService.getVolume();
    if (!_disposed && revision == _revision && data != null) {
      // Server returns 0.0–1.0, slider uses 0–100
      _masterVolume = VolumeHelpers.toSlider((data['level'] as num).toDouble());
      _confirmedVolume = _masterVolume;
      _isMasterMuted = data['muted'] as bool;
    }
    _isLoadingMaster = false;
    notifyListeners();
  }

  Future<void> fetchDevices() async {
    if (_isLoadingDevices || _commandPending) return;
    final revision = _revision;
    _isLoadingDevices = true;
    notifyListeners();
    final data = await ApiService.getAudioDevices();
    if (!_disposed && revision == _revision && data != null) {
      _devices = data
          .map((json) => AudioDevice.fromJson(json as Map<String, dynamic>))
          .toList();
      for (final device in _devices) {
        _confirmedDevices[device.id] = device.volume;
      }
    }
    _isLoadingDevices = false;
    notifyListeners();
  }

  /// Optimistic update for setting master volume level
  Future<bool> setMasterVolume(double level) async {
    if (_commandPending) return false;
    _revision++;
    _commandPending = true;
    final originalVolume = _confirmedVolume;
    _masterVolume = level;
    notifyListeners();

    // Server expects 0.0–1.0 float; slider gives 0–100
    final success = await ApiService.setVolume(VolumeHelpers.toScalar(level));
    _commandPending = false;
    if (success) _confirmedVolume = level;
    if (!success) {
      _masterVolume = originalVolume;
      notifyListeners();
      return false;
    }
    return true;
  }

  /// Optimistic update for toggling master mute state
  Future<bool> toggleMasterMute() async {
    if (_commandPending) return false;
    _revision++;
    _commandPending = true;
    final originalMute = _isMasterMuted;
    final newMute = !_isMasterMuted;
    _isMasterMuted = newMute;
    notifyListeners();

    // Pass explicit new mute state — server requires body {muted: bool}
    final success = await ApiService.setMute(newMute);
    _commandPending = false;
    if (!success) {
      _isMasterMuted = originalMute;
      notifyListeners();
      return false;
    }
    return true;
  }

  /// Optimistic update for setting individual device volume level
  Future<bool> setDeviceVolume(String deviceId, double level) async {
    final index = _devices.indexWhere((d) => d.id == deviceId);
    if (index == -1) return false;

    if (_commandPending) return false;
    _revision++;
    _commandPending = true;
    final originalDevice =
        _devices[index].copyWith(volume: _confirmedDevices[deviceId]);
    _devices[index] = originalDevice.copyWith(volume: level);
    notifyListeners();

    // Server expects volume as 0.0 - 1.0 (so toScalar is called in api_service or here,
    // let's follow the api_service pattern: it divides by 100.0, which matches toScalar).
    final success = await ApiService.setDeviceVolume(deviceId, level);
    if (success) _confirmedDevices[deviceId] = level;
    _commandPending = false;
    final currentIndex = _devices.indexWhere((d) => d.id == deviceId);
    if (!success) {
      if (currentIndex != -1) _devices[currentIndex] = originalDevice;
      notifyListeners();
      return false;
    }
    return true;
  }

  /// Optimistic update for toggling individual device mute state
  Future<bool> toggleDeviceMute(String deviceId, bool targetMute) async {
    final index = _devices.indexWhere((d) => d.id == deviceId);
    if (index == -1) return false;

    if (_commandPending) return false;
    _revision++;
    _commandPending = true;
    final originalDevice =
        _devices[index].copyWith(volume: _confirmedDevices[deviceId]);
    _devices[index] = originalDevice.copyWith(muted: targetMute);
    notifyListeners();

    final success = await ApiService.toggleDeviceMute(deviceId, targetMute);
    _commandPending = false;
    final currentIndex = _devices.indexWhere((d) => d.id == deviceId);
    if (!success) {
      if (currentIndex != -1) _devices[currentIndex] = originalDevice;
      notifyListeners();
      return false;
    }
    return true;
  }

  /// Local immediate state updates for smooth sliding (does not call API)
  void updateMasterVolumeLocally(double level) {
    if (_commandPending) return;
    _revision++;
    _masterVolume = level;
    notifyListeners();
  }

  /// Local immediate state updates for smooth sliding (does not call API)
  void updateDeviceVolumeLocally(String deviceId, double level) {
    if (_commandPending) return;
    _revision++;
    final index = _devices.indexWhere((d) => d.id == deviceId);
    if (index != -1) {
      _devices[index] = _devices[index].copyWith(volume: level);
      notifyListeners();
    }
  }
}
