# Playbook: Desktop (Windows)

Dimuat oleh: VP Desktop, Builder Desktop. Baca juga `ui-design.md` (bahasa desain fluent), `accessibility-i18n.md`, `devops-platform.md` bagian packaging.

## Daftar isi
1. Pilihan framework · 2. Arsitektur · 3. Data, izin, dan integrasi sistem · 4. UI dan aksesibilitas · 5. Keamanan · 6. Build, pengujian, packaging · 7. Anti-pola · 8. Selesai bila

## 1. Pilihan framework (ADR)
| Framework | Cocok bila | Perhatian |
|---|---|---|
| Wails (Go + UI web) | ingin biner kecil dan UI web | butuh WebView2 runtime; toolchain Go |
| Tauri (Rust + UI web) | biner kecil, keamanan ketat | butuh WebView2; toolchain Rust |
| Electron (JS) | ekosistem Node, UI web, banyak pustaka | ukuran besar; kelola pembaruan dan keamanan Chromium |
| .NET (WPF/WinUI 3) | tampilan native Windows, integrasi sistem dalam | satu platform; toolchain .NET |
| Python (PySide6/CustomTkinter + PyInstaller) | alat cepat, keahlian Python | ukuran, antivirus false positive, startup lambat |
Kriteria: ukuran, keahlian, kebutuhan akses sistem, kualitas UI, kemudahan packaging. Tambahkan flag `db`/`network` bila relevan.

## 2. Arsitektur
- **Inti tanpa UI** (logika, penyimpanan) terpisah dari shell UI; inti punya perintah `--selftest` (inisialisasi, cek konfigurasi, keluar 0) agar bisa diuji tanpa GUI.
- Kontrak UI-inti (perintah/IPC) dibekukan di G2. Jangan membocorkan akses sistem mentah ke lapisan UI web (batasi perintah yang diekspos).
- Satu instance (bila sesuai), penanganan galat global, log berkas dengan rotasi di `%LOCALAPPDATA%`.

## 3. Data, izin, dan integrasi sistem
- Data pengguna di `%APPDATA%` atau `%LOCALAPPDATA%`, bukan di folder instalasi; format punya versi dan migrasi.
- Jalankan sebagai pengguna biasa (manifest `asInvoker`); minta hak admin hanya untuk tindakan yang memang butuh.
- Path panjang, berspasi, dan Unicode harus berfungsi; hindari asumsi separator; gunakan API path framework.
- Tray, notifikasi, asosiasi berkas, mulai otomatis: opsional dan dapat dimatikan pengguna.

## 4. UI dan aksesibilitas
Tema terang, gelap, dan kontras tinggi (hormati pengaturan sistem); skala DPI 100-300% dan skala teks; semua kontrol dapat dijangkau keyboard dengan urutan fokus logis dan fokus terlihat; nama dan peran terbaca pembaca layar (UI Automation/ARIA di WebView); pintasan keyboard standar (Ctrl+S, Ctrl+Z, Esc, F1); jendela dapat diubah ukurannya.

## 5. Keamanan
Tanpa kode remote dieksekusi; pembaruan otomatis hanya lewat HTTPS dan verifikasi tanda tangan (gunakan updater framework, jangan buat sendiri); jangan menyimpan rahasia polos (pakai Windows Credential Manager/DPAPI); validasi semua masukan berkas/IPC; DLL/biner pihak ketiga ditinjau dan dikunci versinya; installer tidak menulis ke lokasi yang tak perlu.

## 6. Build, pengujian, packaging
- Tes unit pada inti; integrasi lewat CLI/`--selftest`; tes UI otomatis hanya bila lingkungan agen adalah Windows dengan desktop. **Agen non-Windows tidak bisa menjalankan GUI**: tandai `UNVERIFIED` dan sertakan perintah serta langkah manual untuk manusia.
- Packaging: installer (Inno Setup, WiX, NSIS, MSIX) atau zip portabel; hasilkan checksum SHA-256; uji install, upgrade, dan **uninstall bersih** (skrip PowerShell bila lingkungannya Windows).
- Penandatanganan kode dan reputasi SmartScreen: langkah manusia (sertifikat bukan milik repo). Dokumentasikan peringatan yang mungkin muncul pada biner tak bertanda tangan.
- Antivirus false positive (terutama PyInstaller/Go): catat di README dan sarankan penandatanganan.

## 7. Anti-pola
Logika bisnis di handler UI; data di folder instalasi; meminta hak admin tanpa alasan; updater buatan sendiri tanpa tanda tangan; menyatakan GUI "teruji" tanpa menjalankannya; mengabaikan DPI dan tema kontras tinggi; menyimpan rahasia polos.

## 8. Selesai bila
Build menghasilkan artefak + checksum; `--selftest` keluar 0 (bukti); DoD cluster terpenuhi; UX/UI Critic tanpa S1/S2; verifikasi GUI/instalasi yang tak bisa dijalankan agen ditandai `UNVERIFIED` dengan perintah untuk manusia.
