import 'dart:convert';
import 'dart:typed_data';
import 'package:crypto/crypto.dart' as crypto;
import 'package:encrypt/encrypt.dart' as enc;

class QRCrypto {
  static final Uint8List _secretKey = Uint8List.fromList(
    crypto.sha256.convert(utf8.encode('PCRemote-v3-Secure-QR-Envelope-Protocol-2026')).bytes,
  );
  static const prefix = 'PCR3:';

  /// Decrypts a QR code raw string.
  /// If the raw string starts with "PCR3:", decrypts with AES-256-CBC.
  /// If it is already a plain JSON string (legacy fallback), returns it as-is.
  static String? decrypt(String raw) {
    final trimmed = raw.trim();
    if (trimmed.startsWith(prefix)) {
      try {
        final base64Cipher = trimmed.substring(prefix.length);
        final combined = base64.decode(base64Cipher);
        if (combined.length < 16) return null;

        final ivBytes = combined.sublist(0, 16);
        final cipherBytes = combined.sublist(16);

        final key = enc.Key(_secretKey);
        final iv = enc.IV(Uint8List.fromList(ivBytes));
        final encrypter = enc.Encrypter(enc.AES(key, mode: enc.AESMode.cbc, padding: 'PKCS7'));

        final decrypted = encrypter.decrypt(enc.Encrypted(Uint8List.fromList(cipherBytes)), iv: iv);
        return decrypted;
      } catch (_) {
        return null;
      }
    }

    // Fallback: if it's already a JSON payload
    if (trimmed.startsWith('{') && trimmed.endsWith('}')) {
      return trimmed;
    }

    return null;
  }
}
