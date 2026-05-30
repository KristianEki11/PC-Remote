import 'dart:async';
import 'package:flutter/material.dart';
import '../services/api_service.dart';
import 'media_status.dart';

class MediaState extends ChangeNotifier {
  MediaStatus _currentStatus = MediaStatus.empty();
  bool _isLoading = false;
  Timer? _pollingTimer;

  MediaStatus get currentStatus => _currentStatus;
  bool get isLoading => _isLoading;

  Future<void> fetchMediaStatus() async {
    final data = await ApiService.getMediaStatus();
    if (data != null) {
      _currentStatus = MediaStatus.fromJson(data);
      notifyListeners();
    }
  }

  void startPolling() {
    _pollingTimer?.cancel();
    _pollingTimer = Timer.periodic(const Duration(seconds: 3), (timer) {
      fetchMediaStatus();
    });
  }

  void stopPolling() {
    _pollingTimer?.cancel();
    _pollingTimer = null;
  }

  /// Optimistic play/pause trigger
  Future<bool> playPause() async {
    final originalStatus = _currentStatus;
    // Optimistic toggle
    _currentStatus = MediaStatus(
      success: true,
      status: originalStatus.isPlaying ? 'Paused' : 'Playing',
      title: originalStatus.title,
      artist: originalStatus.artist,
      album: originalStatus.album,
      appId: originalStatus.appId,
    );
    _isLoading = true;
    notifyListeners();

    final success = await ApiService.mediaPlayPause();
    _isLoading = false;
    if (!success) {
      _currentStatus = originalStatus;
    }
    notifyListeners();
    
    // Quick refresh to align state
    Future.delayed(const Duration(milliseconds: 300), fetchMediaStatus);
    return success;
  }

  /// Optimistic next track trigger
  Future<bool> next() async {
    _isLoading = true;
    notifyListeners();
    final success = await ApiService.mediaNext();
    _isLoading = false;
    notifyListeners();
    
    // Quick refresh to align state
    Future.delayed(const Duration(milliseconds: 500), fetchMediaStatus);
    return success;
  }

  /// Optimistic previous track trigger
  Future<bool> prev() async {
    _isLoading = true;
    notifyListeners();
    final success = await ApiService.mediaPrev();
    _isLoading = false;
    notifyListeners();
    
    // Quick refresh to align state
    Future.delayed(const Duration(milliseconds: 500), fetchMediaStatus);
    return success;
  }

  @override
  void dispose() {
    stopPolling();
    super.dispose();
  }
}
