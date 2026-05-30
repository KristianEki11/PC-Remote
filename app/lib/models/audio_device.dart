class AudioDevice {
  final String id;
  final String name;
  final double volume; // 0-100 to match slider
  final bool muted;

  AudioDevice({
    required this.id,
    required this.name,
    required this.volume,
    required this.muted,
  });

  factory AudioDevice.fromJson(Map<String, dynamic> json) {
    // Convert server volume scalar (0.0-1.0) to slider volume (0-100)
    // Server payload: {"id": "...", "name": "...", "volume": 0.5, "muted": false}
    final double rawVolume = (json['volume'] as num?)?.toDouble() ?? 0.0;
    return AudioDevice(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? 'Perangkat Audio',
      volume: (rawVolume * 100.0).clamp(0.0, 100.0),
      muted: json['muted'] as bool? ?? false,
    );
  }

  AudioDevice copyWith({
    String? id,
    String? name,
    double? volume,
    bool? muted,
  }) {
    return AudioDevice(
      id: id ?? this.id,
      name: name ?? this.name,
      volume: volume ?? this.volume,
      muted: muted ?? this.muted,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'volume': volume / 100.0, // convert back to scalar for server if needed
      'muted': muted,
    };
  }
}
