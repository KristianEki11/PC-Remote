# Archetype: desktop-windows (aplikasi Windows)

Flags: ui, desktop, installable
Cocok untuk: aplikasi GUI Windows, alat utilitas, installer atau versi portabel.
Profil default: LEAN | Bahasa desain default: fluent (Fluent 2/WinUI) atau custom; UI berbasis web (Wails/Tauri/Electron) tetap mengikuti token proyek.
Stack: pilih lewat ADR dengan kriteria ukuran, keahlian, kebutuhan akses sistem, dan UI: Wails (Go + UI web), Tauri (Rust + UI web), Electron (JS), .NET (WPF/WinUI 3), Python (PySide6/CustomTkinter + PyInstaller). Tambahkan flag `db` bila ada penyimpanan lokal, `network` bila memanggil layanan luar.
Playbook kunci: `desktop-windows`, `ux-design`, `ui-design`, `design-system`, `devops-platform`, `security-appsec`, `qa-testing`.

## Pertanyaan intake
1. Portabel atau installer? Windows 10 dan 11, atau 11 saja?
2. Butuh hak admin, akses berkas luas, atau integrasi sistem (tray, ekstensi berkas, mulai otomatis)?
3. Offline saja atau memanggil layanan online; data disimpan di mana (SQLite/JSON)?

## Contoh cluster
- C1 Inti tanpa UI (logika, penyimpanan) + perintah `--selftest` yang keluar 0
- C2 Shell UI: jendela, navigasi, pengaturan, tema terang/gelap/kontras tinggi
- C3 Fitur utama
- C4 Integrasi sistem (tray, notifikasi, asosiasi berkas) dan penanganan galat/log
- C5 Packaging (installer/portabel), pembaruan, uninstall

## Kontrak yang dibekukan di G2
Perintah/IPC antara UI dan inti, format penyimpanan + versi migrasinya, lokasi data (`%APPDATA%`/`%LOCALAPPDATA%`), token desain, daftar izin sistem yang dibutuhkan.

## Alur kritis
Pasang atau jalankan, alur inti, tutup lalu buka lagi (data bertahan), mode gelap/kontras tinggi, uninstall bersih.

## Risiko khas
Antivirus/SmartScreen memberi peringatan pada biner baru (tanda tangan kode dan reputasi adalah langkah manusia), dependensi WebView2, skala DPI dan teks, path panjang atau berspasi/Unicode, hak admin yang tidak perlu, pembaruan otomatis tanpa verifikasi tanda tangan, GUI tidak bisa diuji di agen non-Windows (tandai `UNVERIFIED`, jangan dipalsukan).
