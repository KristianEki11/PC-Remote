class MediaStatus {
  final bool success;
  final String status; // e.g. "Playing", "Paused", "Stopped", "Closed"
  final String title;
  final String artist;
  final String album;
  final String? appId;

  MediaStatus({
    required this.success,
    required this.status,
    required this.title,
    required this.artist,
    required this.album,
    this.appId,
  });

  bool get isPlaying => status == 'Playing';

  factory MediaStatus.fromJson(Map<String, dynamic> json) {
    return MediaStatus(
      success: json['success'] as bool? ?? false,
      status: json['status'] as String? ?? 'Closed',
      title: json['title'] as String? ?? '',
      artist: json['artist'] as String? ?? '',
      album: json['album'] as String? ?? '',
      appId: json['app_id'] as String?,
    );
  }

  factory MediaStatus.empty() {
    return MediaStatus(
      success: true,
      status: 'Closed',
      title: '',
      artist: '',
      album: '',
    );
  }
}
