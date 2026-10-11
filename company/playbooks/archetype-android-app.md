# Archetype: android-app (aplikasi Android)

Flags: ui, mobile, installable
Cocok untuk: aplikasi Android native atau lintas platform, dengan atau tanpa backend (tambahkan archetype `api-service`/`web-app` bila backend ikut dibangun).
Profil default: LEAN | Bahasa desain default: material (Material 3).
Stack: Kotlin + Jetpack Compose (native, disarankan) atau Flutter atau React Native/Expo bila keahlian JS diutamakan. Putuskan lewat ADR: kebutuhan perangkat keras, performa, pemeliharaan. Periksa dulu apakah Antigravity menyediakan alat/skill Android bawaan.
Playbook kunci: `mobile-android`, `ux-design`, `ui-design`, `design-system`, `accessibility-i18n`, `security-appsec`, `qa-testing`.

## Pertanyaan intake
1. Target minSdk/perangkat, dan apakah harus offline-first?
2. Izin sensitif yang dibutuhkan (kamera, lokasi, notifikasi, penyimpanan)?
3. Distribusi: Play Store atau APK langsung? Butuh backend?

## Contoh cluster
- C1 Fondasi: proyek Gradle, tema Material 3 dari token, navigasi, DI, tata letak modul
- C2 Lapisan data: Room/DataStore, jaringan, repositori
- C3 Fitur inti (UI + ViewModel + tes)
- C4 Fitur pendukung: notifikasi, pekerjaan latar (WorkManager), pengaturan
- C5 Rilis: R8/shrink, ikon dan aset store, konfigurasi penandatanganan (kunci produksi adalah rahasia manusia)

## Kontrak yang dibekukan di G2
Model data + skema Room dan versi migrasi, graf navigasi, kontrak API bila ada, token yang dipetakan ke tema Material 3, daftar izin.

## Alur kritis
Onboarding pertama kali, alur inti, mode offline/online, rotasi layar dan proses dimatikan sistem (state pulih), izin ditolak, tombol kembali/gestur kembali.

## Risiko khas
Persyaratan Play Console dan level API target yang berubah (cek sumber resmi, bertanggal), kebijakan data dan izin, siklus hidup dan konfigurasi, ukuran APK, tes instrumented butuh emulator/perangkat (jika tidak ada: `UNVERIFIED` dengan perintah untuk manusia), kunci rilis tidak boleh masuk repo.
