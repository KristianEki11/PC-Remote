import 'package:flutter/material.dart';

class AppState extends ChangeNotifier {
  String _ipAddress = '';
  String _token = '';
  bool _isConnected = false;

  String get ipAddress => _ipAddress;
  String get token => _token;
  bool get isConnected => _isConnected;

  void setConnectionDetails(String ip, String newToken) {
    _ipAddress = ip;
    _token = newToken;
    notifyListeners();
  }

  void setConnectionStatus(bool status) {
    if (_isConnected != status) {
      _isConnected = status;
      notifyListeners();
    }
  }

  void clear() {
    _ipAddress = '';
    _token = '';
    _isConnected = false;
    notifyListeners();
  }
}
