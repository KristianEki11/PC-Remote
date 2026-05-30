/// Utility helpers to convert volume between server scalar (0.0 - 1.0)
/// and local Flutter slider values (0.0 - 100.0).
class VolumeHelpers {
  /// Converts a server scalar volume (0.0 to 1.0) to a slider value (0.0 to 100.0).
  static double toSlider(double scalar) {
    return (scalar * 100.0).clamp(0.0, 100.0);
  }

  /// Converts a local slider value (0.0 to 100.0) to a server scalar volume (0.0 to 1.0).
  static double toScalar(double slider) {
    return (slider / 100.0).clamp(0.0, 1.0);
  }
}
