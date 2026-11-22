import 'dart:async';
import 'package:flutter/material.dart';
import '../services/api_service.dart';
import 'media_status.dart';

class MediaState extends ChangeNotifier {
  MediaStatus _currentStatus = MediaStatus.empty();
  bool _isLoading = false;
  Timer? _pollingTimer;
  bool _fetching = false;
  bool _disposed = false;
  int _revision = 0;

  MediaStatus get currentStatus => _currentStatus;
  bool get isLoading => _isLoading;

  Future<void> fetchMediaStatus() async {
    if (_disposed || _fetching) return;
    if (ApiService.authToken.isEmpty) {
      stopPolling();
      return;
    }
    final revision = _revision;
    _fetching = true;
    final data = await ApiService.getMediaStatus();
    _fetching = false;
    if (_disposed) return;
    if (revision == _revision && data != null) {
      _currentStatus = MediaStatus.fromJson(data);
      notifyListeners();
    }
  }

  void startPolling() {
    if (_pollingTimer != null || _disposed) return;
    fetchMediaStatus();
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
    if (_disposed || _isLoading) return false;
    _revision++;
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
    if (_disposed) return success;
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
    if (_disposed || _isLoading) return false;
    _revision++;
    _isLoading = true;
    notifyListeners();
    final success = await ApiService.mediaNext();
    if (_disposed) return success;
    _isLoading = false;
    notifyListeners();

    // Quick refresh to align state
    Future.delayed(const Duration(milliseconds: 500), fetchMediaStatus);
    return success;
  }

  /// Optimistic previous track trigger
  Future<bool> prev() async {
    if (_disposed || _isLoading) return false;
    _revision++;
    _isLoading = true;
    notifyListeners();
    final success = await ApiService.mediaPrev();
    if (_disposed) return success;
    _isLoading = false;
    notifyListeners();

    // Quick refresh to align state
    Future.delayed(const Duration(milliseconds: 500), fetchMediaStatus);
    return success;
  }

  /// Seek forward (+10s in YouTube / video)
  Future<bool> seekForward() async {
    final success = await ApiService.mediaSeekForward();
    return success;
  }

  /// Seek backward (-10s in YouTube / video)
  Future<bool> seekBackward() async {
    final success = await ApiService.mediaSeekBackward();
    return success;
  }

  @override
  void dispose() {
    _disposed = true;
    stopPolling();
    super.dispose();
  }
}
