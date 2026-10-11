# Playbook: Mobile (Android)

Dimuat oleh: VP Mobile, Builder Android. Baca juga `ui-design.md` (bahasa desain material), `accessibility-i18n.md`, `security-appsec.md` bagian mobile.

## Daftar isi
1. Pilihan stack · 2. Arsitektur · 3. Siklus hidup, izin, dan offline · 4. UI dan aksesibilitas · 5. Keamanan (MASVS) · 6. Build dan pengujian · 7. Rilis dan Play Store · 8. Anti-pola · 9. Selesai bila

## 1. Pilihan stack (ADR)
- **Kotlin + Jetpack Compose** (native): default untuk akses perangkat keras, performa, dan konvensi Material 3.
- **Flutter** atau **React Native/Expo**: bila berbagi kode lintas platform atau keahlian JS lebih penting. Putuskan lewat ADR; token desain dipetakan sesuai `design-system.md`.
- Periksa lebih dulu apakah Antigravity menyediakan alat/skill Android bawaan di lingkunganmu dan catat di `tools.md`.

## 2. Arsitektur (versi native)
MVVM dengan aliran data satu arah: UI (Compose) → ViewModel → repositori → sumber data (Room/DataStore/jaringan). DI dengan Hilt atau Koin. Navigasi lewat Navigation Compose dengan graf yang dibekukan di G2. Modularisasi bila proyek besar. Logika bisnis bebas dari Android framework agar bisa dites unit.

## 3. Siklus hidup, izin, dan offline
- State harus selamat dari rotasi dan **proses dimatikan sistem** (simpan state penting, uji).
- Izin diminta saat dibutuhkan (runtime), dengan penjelasan; aplikasi tetap berfungsi bermakna bila izin ditolak. Izin sesedikit mungkin.
- Offline-first bila diminta: sumber kebenaran lokal (Room), sinkronisasi lewat WorkManager dengan batas retry; indikator status koneksi.
- Kerja latar memakai WorkManager atau foreground service sesuai kebijakan platform terbaru (verifikasi).

## 4. UI dan aksesibilitas
Tema Material 3 dari token (`lightColorScheme/darkColorScheme`), dukung edge-to-edge dan insets, gestur/tombol kembali selalu bermakna, target minimal 48 dp, teks dalam `sp` (skala teks sistem), label semantik untuk TalkBack, urutan fokus logis, kontras sesuai kalibrasi. Layar lebar/lipat: gunakan kelas ukuran jendela (compact/medium/expanded) bila relevan.

## 5. Keamanan (OWASP MASVS)
Penyimpanan rahasia di Android Keystore (bukan SharedPreferences polos); lalu lintas hanya HTTPS (network security config, tanpa cleartext); komponen yang tidak perlu tidak di-`export`; WebView dikeraskan (tanpa JS bridge yang tidak perlu, domain dibatasi); tidak ada rahasia atau PII di log; aturan backup dibatasi untuk data sensitif; validasi deep link/intent; obfuscation/shrink (R8) untuk rilis. Rujukan: MASVS dan MASTG.

## 6. Build dan pengujian
Perintah standar (variant `debug`): `./gradlew assembleDebug testDebugUnitTest lintDebug` (Windows: `gradlew.bat`). Tes unit (JUnit) untuk logika dan ViewModel; tes UI Compose; tes instrumented dan emulator **hanya bila tersedia**. Tanpa emulator/perangkat: tandai `UNVERIFIED` dan tulis perintah persis untuk manusia (`./gradlew connectedDebugAndroidTest`). Jangan memalsukan hasil UI.

## 7. Rilis dan Play Store
Kunci penandatanganan rilis adalah rahasia manusia (tidak di repo; `.gitignore`). Persyaratan Play Console (level API target, arsitektur 64-bit, dukungan ukuran halaman memori 16 KB, formulir Data Safety, kebijakan privasi) **berubah berkala**: Explorer memverifikasi dari dokumentasi resmi dengan tanggal sebelum G5. Siapkan aset store (ikon, tangkapan layar, deskripsi) dan daftar izin beserta alasannya. Publikasi adalah langkah manusia.

## 8. Anti-pola
Logika bisnis di Activity/composable; state hanya di memori; izin diminta semua di awal; hard-coded string/warna/ukuran; mengabaikan tombol kembali; rahasia di repo; mengandalkan satu ukuran layar; menyatakan UI "lolos" tanpa emulator.

## 9. Selesai bila
`assembleDebug`, `testDebugUnitTest`, `lintDebug` lolos dengan bukti; DoD cluster terpenuhi; UX/UI Critic tanpa S1/S2; butir yang butuh emulator ditandai `UNVERIFIED` dengan perintah untuk manusia.
