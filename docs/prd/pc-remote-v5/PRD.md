# PRD: PC Remote v5.0 — Native Windows & 24/7 Companion Edition

> Versi 1.0 - 2026-11-22 - Status: Approved
> Sumber: wawancara /prd + kerangka fitur yang disetujui (`docs/prd/pc-remote-v5/project-state.json`)

## 1. Ringkasan
PC Remote v5.0 adalah aplikasi kontrol jarak jauh PC Windows dari HP Android yang dibangun ulang secara native dengan Rust (Windows) dan Flutter (Android) menggunakan jaringan privat Tailscale milik pengguna masing-masing. Versi ini menggantikan dashboard web browser dengan aplikasi desktop native Windows, menghilangkan ritual *re-pairing* saat laptop mati-nyala atau berpindah jaringan, serta menambahkan servis latar belakang 24/7 bergaya *companion smartwatch* di Android lengkap dengan notifikasi status koneksi dan kontrol media interaktif.

## 2. Masalah & Pengguna
### 2.1 Persona utama
- **Pengguna Laptop Pribadi / Mahasiswa (Persona Utama):** Sering menggunakan laptop di kamar (Wi-Fi rumah) untuk menonton YouTube, mendengarkan musik, atau presentasi, lalu mematikan/menidurkan laptop di malam hari dan membawanya ke kampus atau bekerja dari luar rumah dengan 4G/5G.
- **Pengguna Baru Non-Teknis:** Mengunduh aplikasi dari halaman rilis, tidak memahami konsep tunnel/VPN/Tailscale, dan menginginkan panduan langkah demi langkah yang jelas tanpa menyentuh terminal.

### 2.2 Cara mereka mengatasi masalah sekarang
Pada v4.0.0 (Go + Cloudflare Quick Tunnel), setiap kali laptop dimatikan atau masuk mode sleep, Quick Tunnel mati dan URL berubah. Saat laptop dinyalakan kembali di pagi hari, meskipun pengguna membuka aplikasi HP di Wi-Fi rumah untuk memancing sinkronisasi, aplikasi gagal memperbarui alamat remote baru. Pengguna terpaksa memutus koneksi (*disconnect*), membuka dashboard web di laptop, membuat QR baru, dan melakukan scan ulang.

### 2.3 Pembeda
1. **Aplikasi Desktop Native Windows (Tanpa Web Browser):** Pengaturan, QR pairing, daftar perangkat, dan status Tailscale tampil langsung di jendela aplikasi Windows yang ringan dan menutup ke System Tray.
2. **Zero Re-Pairing Smart Roaming:** Alamat Tailscale bersifat permanen; ditambah *fast boot polling* dan *auto-sync* endpoint di background, HP selalu mendapatkan alamat LAN dan Tailscale terbaru secara otomatis tanpa pernah scan QR ulang.
3. **24/7 Smartwatch-Style Companion Service:** Aplikasi HP terus siaga di latar belakang dengan notifikasi status koneksi real-time dan notifikasi media interaktif (Judul, Prev, Play/Pause, Next) yang bisa ditekan dari Lock Screen maupun Notification Shade.
4. **Wizard Setup Tailscale 3-Langkah:** Memandu pengguna awam memasang dan login Tailscale di PC dan HP secara interaktif dengan deteksi status otomatis.

## 3. Tujuan & Ukuran Keberhasilan
- **Tujuan:**
  1. Menghapus kebutuhan scan QR ulang setelah laptop restart, sleep, atau pindah dari Wi-Fi ke seluler.
  2. Menghadirkan pengalaman kontrol instan dari panel notifikasi Android 24/7 seperti aplikasi pendamping *smartwatch*.
  3. Memudahkan pengguna awam mengaktifkan akses luar rumah (Tailscale) dalam 3 langkah terpandu.
  4. Menyelesaikan migrasi penuh ke Rust + Tailscale + Flutter (`All-In Cutover`).
- **Ukuran Keberhasilan:**
  - *(Usulan)* 0 kali re-pairing setelah laptop reboot/sleep dan berpindah dari Wi-Fi ke 4G/5G ketika PC & HP berada pada tailnet yang sama.
  - *(Usulan)* Pengguna baru dapat menyelesaikan pairing LAN pertama dalam < 30 detik dan mengaktifkan Tailscale dalam < 3 menit lewat wizard.
  - *(Usulan)* Tombol `Prev` / `Play/Pause` / `Next` pada notifikasi Android merespons tanpa membuka aplikasi utama.
- **Bukan Tujuan:**
  - Tidak membuat server relay cloud terpusat atau akun terpusat pengembang.
  - Tidak lagi mendukung Web Remote (`/connect`) melalui browser karena pembatasan origin keamanan.

## 4. Pengalaman Inti
### 4.1 Momen pertama
1. Pengguna memasang `PCRemoteRustSetup.exe` dan membuka **PC Remote** — jendela aplikasi native Windows langsung muncul dengan QR code pairing dan kartu status koneksi.
2. Pengguna membuka **PC Remote** di Android, memindai QR di layar laptop — HP langsung terhubung lewat Wi-Fi lokal dalam hitungan detik dan notifikasi *companion* aktif di status bar HP.
3. Jika ingin akses dari luar rumah (kampus/kantor), pengguna menekan kartu **"Aktifkan Mode Luar Rumah"** dan mengikuti Wizard 3 Langkah Tailscale tanpa mengetik perintah apa pun.

### 4.2 Alasan kembali
- Kontrol volume, mute, kecerahan layar (termasuk overlay OLED), dan media langsung dari HP atau dari notifikasi Android 24/7 tanpa perlu bangkit dari tempat tidur/sofa.
- Laptop dinyalakan pagi hari -> notifikasi HP otomatis berubah menjadi `Terhubung`, judul YouTube/Spotify muncul di notifikasi, dan siap dipakai di kampus tanpa ritual re-pairing.

### 4.3 Alur utama
1. PC menyala -> Server Rust otomatis aktif di System Tray dan melakukan *fast boot polling* untuk mendeteksi IP LAN dan Tailscale.
2. Companion Service di HP otomatis terhubung via WebSocket (`/events`) atau *parallel probe* LAN/Tailscale, lalu menyinkronkan `/connection`.
3. Saat media diputar di PC, event WinRT mendorong perubahan ke HP secara real-time dan memperbarui notifikasi media interaktif Android.

## 5. Kerangka Fitur
Identik dengan `docs/prd/pc-remote-v5/kerangka-fitur.md`:
- **Fase 1:** `M1` (M1.1–M1.4), `M2` (M2.1–M2.4), `M3` (M3.1–M3.4), `M4` (M4.1–M4.4), `M5` (M5.1–M5.4), `M6` (M6.1–M6.4).
- **Fase 2:** `M7` (M7.1–M7.3), `M8` (M8.1–M8.3).

## 6. Spesifikasi Fitur per Fase
### Fase 1 - Native Desktop App, 24/7 Companion & Smart Roaming

#### M1 Aplikasi Desktop Native Windows & Tray
**Tujuan modul:** Menyediakan antarmuka desktop native Windows tanpa browser untuk pairing QR, diagnostik jaringan, wizard Tailscale, dan kontrol server di System Tray.
| ID | Sub-fitur | Deskripsi | Kriteria penerimaan |
|---|---|---|---|
| M1.1 | Jendela GUI Native Windows | Jendela desktop native Win32/GDI dark mode yang muncul saat aplikasi dibuka langsung (atau via tray/instance kedua). | Membuka `pcremote-server.exe` menampilkan jendela aplikasi native tanpa membuka tab browser; menutup jendela menyembunyikannya ke System Tray. |
| M1.2 | Render QR Pairing Native | Merender QR pairing sekali pakai beserta hitung mundur 120 detik dan tombol buat QR baru langsung di jendela aplikasi. | Klik "Buat QR Pairing" menampilkan QR yang terbaca oleh kamera Flutter dan menghitung mundur 120 detik secara langsung. |
| M1.3 | Wizard Setup Tailscale PC | Panel status LAN/Tailscale dengan 3 langkah interaktif: cek/unduh Tailscale, tombol buka/login Tailscale, dan status siap luar rumah yang otomatis refresh. | Wizard mendeteksi perubahan status Tailscale dari `NotInstalled` / `NeedsLogin` ke `Running` dalam <= 3 detik tanpa restart aplikasi. |
| M1.4 | Tray, Close-to-Tray & Autostart | Ikon System Tray dengan menu buka jendela & hentikan server, daftar perangkat tepercaya dengan tombol cabut akses, atur PIN, dan toggle autostart Windows. | Klik kanan tray menampilkan menu Buka/Stop; tombol Cabut Akses menghapus perangkat; toggle Autostart menulis/menghapus entri Run user Windows. |

#### M2 Mesin Koneksi Zero Re-Pairing
**Tujuan modul:** Menjamin HP selalu memiliki alamat LAN dan Tailscale PC terbaru secara otomatis dan berpindah jalur mulus tanpa kehilangan atau menggandakan perintah.
| ID | Sub-fitur | Deskripsi | Kriteria penerimaan |
|---|---|---|---|
| M2.1 | Fast Boot Polling & Auto-Sync | Server melakukan polling cepat tiap 3 detik selama 90 detik awal startup dan refresh on-demand pada `GET /connection`; Flutter otomatis memanggil sinkronisasi endpoint saat connect/ping/WS snapshot. | Jika HP dipairing sebelum Tailscale aktif lalu Tailscale dinyalakan belakangan, HP otomatis menyimpan `tailscale_url` baru tanpa scan QR ulang. |
| M2.2 | Fast Parallel Probe LAN+Tailscale | Flutter menyimpan kandidat alamat (`last_ip`, `alternate_hosts`, `tailscale_url`, `tailscale_urls`) dan menguji kandidat secara paralel saat jalur aktif timeout/putus. | Saat Wi-Fi dimatikan dan pindah ke jaringan Tailscale, request `GET` berikutnya otomatis beralih ke endpoint Tailscale yang merespons. |
| M2.3 | Failover Command Idempotent | Perintah `POST` memakai `X-Request-Id` dan `X-Command-Deadline-Ms` yang sama saat failover ke kandidat endpoint lain agar server tidak mengeksekusi ulang perintah yang sudah diterima. | Mengirim `POST` dengan `X-Request-Id` yang sama dua kali ke server menghasilkan status tersimpan tanpa eksekusi ganda. |
| M2.4 | Wizard & Deep-Link Tailscale HP | Kartu status "Akses Luar Rumah" di Dashboard/Settings Android yang membuka panduan interaktif, deep-link ke aplikasi/Play Store Tailscale, dan tombol tes koneksi Tailscale. | Menekan tombol Tailscale membuka app/halaman Play Store Tailscale; tombol Tes Koneksi memverifikasi `tailscale_url/health` dan menampilkan status berhasil/gagal yang jelas. |

#### M3 Servis Companion 24/7 Android
**Tujuan modul:** Menjaga koneksi latar belakang 24/7 seperti aplikasi pendamping smartwatch dan menyediakan notifikasi status koneksi serta kontrol media interaktif.
| ID | Sub-fitur | Deskripsi | Kriteria penerimaan |
|---|---|---|---|
| M3.1 | Foreground Service 24/7 Background | Servis Android yang menjaga heartbeat dan auto-reconnect di latar belakang selama sesi login aktif. | Saat aplikasi diminimize, koneksi tetap terjaga dan otomatis terhubung kembali saat laptop menyala. |
| M3.2 | Notifikasi Status Koneksi Live | Notifikasi persisten menampilkan status real-time (`Terhubung • Wi-Fi Lokal`, `Terhubung • Tailscale`, atau `Menunggu Laptop Aktif...`). | Perubahan status koneksi langsung memperbarui teks notifikasi di panel notifikasi Android. |
| M3.3 | Notifikasi Media Play/Pause/Next/Prev | Saat media aktif di PC, notifikasi menampilkan judul/artis/app dan menyediakan 3 tombol aksi (`Prev`, `Play/Pause`, `Next`). | Menekan tombol Play/Pause/Next pada notifikasi memicu perintah media ke PC tanpa harus membuka aplikasi. |
| M3.4 | Tap-to-Open & Izin Notifikasi | Mengetuk notifikasi langsung membuka Dashboard aplikasi; tersedia pengaturan toggle servis 24/7 dan permintaan izin notifikasi Android 13+. | Tap pada notifikasi membuka layar utama aplikasi dalam kondisi aktif/terhubung. |

#### M4 Keamanan & Pairing Permanen
**Tujuan modul:** Melindungi identitas server, TLS fingerprint, dan token perangkat menggunakan penyimpanan aman OS tanpa ketergantungan cloud pihak ketiga.
| ID | Sub-fitur | Deskripsi | Kriteria penerimaan |
|---|---|---|---|
| M4.1 | Identitas DPAPI & TLS Pinning | Identitas server dan private key TLS dienkripsi DPAPI pada PC; Android memverifikasi SHA-256 sertifikat pada setiap koneksi LAN dan Tailscale. | Sertifikat yang tidak cocok dengan fingerprint hasil pairing selalu ditolak klien Android. |
| M4.2 | Pairing QR & Sesi Jangka Panjang | Token pairing berlaku 120 detik sekali pakai dan ditukar token Bearer per perangkat (berlaku 1 tahun) yang disimpan di Android Keystore (`FlutterSecureStorage`). | Token pairing yang sudah dipakai sekali ditolak dengan 401 pada percobaan kedua. |
| M4.3 | Manajemen & Pencabutan Perangkat | Pencabutan akses per perangkat dari GUI Windows atau logout dari HP langsung membatalkan token perangkat tersebut. | Perangkat yang dicabut langsung mendapat HTTP 401 pada request berikutnya. |
| M4.4 | Proteksi Brute-Force & Server-ID | Pembatasan 5 percobaan gagal per IP sumber (lockout 15 menit) dan penolakan origin browser (`Origin` header). | Request dari browser dengan header `Origin` ditolak dengan HTTP 403. |

#### M5 Kontrol Audio, Media & Layar
**Tujuan modul:** Memberikan kontrol hardware Windows yang cepat dan andal melalui API native COM dan WinRT tanpa proses PowerShell berkala.
| ID | Sub-fitur | Deskripsi | Kriteria penerimaan |
|---|---|---|---|
| M5.1 | Master, Device & Sonar COM | Baca dan atur volume/mute master, per-device output, serta kanal SteelSeries Sonar via COM. | Slider dan tombol mute di HP mengubah volume perangkat audio Windows terkait. |
| M5.2 | Kontrol Media WinRT Real-Time | Baca status media (judul, artis, app, status play) via WinRT GSMTC serta kontrol Play/Pause, Next, Prev, dan Seek +/-10 detik. | Event perubahan lagu di PC mendorong update ke WebSocket `/events` tanpa polling PowerShell. |
| M5.3 | Hybrid Brightness WMI + OLED | Kecerahan 50–100% mengatur WMI backlight; 0–49% menerapkan overlay hitam transparan Win32 di seluruh virtual screen. | Mengubah kecerahan ke 25% meredupkan layar melalui overlay tanpa memblokir klik mouse. |
| M5.4 | Power, Sleep, Lock & Shutdown | Lock PC, S3 Sleep, Restart, Display Off, dan Shutdown (langsung atau terjadwal + batal). | Setiap aksi power memvalidasi sesi interaktif non-Session-0 sebelum mengeksekusi API Windows. |

#### M6 Cutover All-In, Installer & Web
**Tujuan modul:** Menyelesaikan transisi penuh ke Rust v5.0 pada installer, aplikasi Flutter, dan situs web.
| ID | Sub-fitur | Deskripsi | Kriteria penerimaan |
|---|---|---|---|
| M6.1 | Installer NSIS & Migrasi Go | Installer `PCRemoteRust.nsi` menghentikan/membersihkan proses/service Go & `cloudflared` lama, membuat shortcut Start Menu/Desktop, dan opsi autostart. | Instalasi selesai bersih tanpa konflik dengan sisa service Go v4.x. |
| M6.2 | Deaktivasi /connect & Landing v5.0 | Mengganti `/connect` dengan halaman informasi native-only dan memperbarui landing page ke fitur v5.0 (Rust + Tailscale + 24/7 Companion) dengan `npm run lint` bersih. | `npm run lint` di `web/` keluar dengan kode 0 tanpa error. |
| M6.3 | Port 8443 & Cleanup Flutter | Menyeragamkan port default `:8443`, versi `v5.0.0`, dan menghapus dependency Flutter yang tidak dipakai. | `flutter analyze` dan `flutter test` lulus tanpa error. |
| M6.4 | Smoke Test, Roaming & Build | Menjalankan `cargo fmt/clippy/test/build --release`, `smoke.ps1`, `flutter test`, dan verifikasi build installer/APK. | Seluruh check CI lokal lulus tanpa kegagalan. |

### Fase 2 - Perluasan Kontrol PC Lebih Luas & Cepat
- **M7 Kontrol Audio & Aplikasi Lanjutan:**
  - `M7.1` Per-App Volume Mixer: mengatur volume masing-masing aplikasi yang sedang bersuara di Windows (mis. Chrome vs Game vs Discord). *(Kriteria: daftar sesi audio aktif tampil dan volume tiap aplikasi dapat diubah).*
  - `M7.2` Ganti Default Output Audio: memindahkan perangkat output utama Windows (Headset <-> Speaker) langsung dari HP. *(Kriteria: memilih perangkat mengubah default playback endpoint Windows).*
  - `M7.3` Preset Cepat Audio & Layar: tombol 1-ketuk (mis. Mode Malam / Mode Presentasi). *(Kriteria: 1 ketukan menerapkan kombinasi volume dan kecerahan).*
- **M8 Navigasi Cepat & Input Remote:**
  - `M8.1` Touchpad Virtual & Klik Mouse: menggerakkan kursor dan klik kiri/kanan dari layar HP. *(Kriteria: gesture geser menggerakkan kursor Windows).*
  - `M8.2` Shortcut Presentasi & Fullscreen: tombol Next/Prev slide, `F` fullscreen YouTube, dan `Space`/`Esc`. *(Kriteria: tombol mengirim input keyboard ke jendela aktif PC).*
  - `M8.3` Kirim Teks & Clipboard Cepat: mengirim teks/URL dari HP langsung ke clipboard/input aktif PC. *(Kriteria: teks dari HP dapat ditempel di PC).*

## 7. Kebutuhan Non-Fungsional
- **Kecepatan & Footprint:** Server Rust single-binary ringan tanpa spawn PowerShell berkala untuk status media/audio; startup GUI < 1 detik.
- **Perangkat Didukung:** Windows 10/11 (64-bit) dengan sesi desktop interaktif; Android 8.0+ (API 26+).
- **Bahasa Antarmuka:** Bahasa Indonesia untuk seluruh antarmuka aplikasi Android, jendela native Windows, dan panduan wizard.
- **Perilaku Offline:** Saat PC mati/offline, token kredensial dan daftar kandidat endpoint di HP tetap tersimpan; aplikasi masuk mode standby/menunggu dan otomatis pulih saat PC kembali online.

## 8. Rilis & Prioritas
Urutan pengerjaan di dalam Fase 1 berdasarkan ketergantungan:
1. `M1` & `M4` & `M5` (Fondasi Server Rust Native: GUI Win32, Tray, Wizard PC, Endpoint API, dan Audio/Media/Brightness).
2. `M2` (Mesin Koneksi Zero Re-Pairing di Rust + Flutter + Wizard HP).
3. `M3` (Servis Companion 24/7 & Notifikasi Media Interaktif Android).
4. `M6` (Installer NSIS, Pembersihan Web & Flutter, serta Pengujian Menyeluruh).

Fase 1 dianggap selesai apabila seluruh tes otomatis Rust, smoke test, `flutter analyze`/`flutter test`, dan `npm run lint` web lulus 100% serta dokumen `PRD.md` & `TECH-SPEC.md` konsisten.

## 9. Risiko
- **Pembatasan Background Android (Battery Optimization / Doze Mode):** Pabrikan HP tertentu membatasi servis latar belakang. *Mitigasi:* Menggunakan Foreground Service resmi Android dengan notifikasi persisten dan tombol panduan izin baterai di pengaturan.
- **Tailscale Belum Terpasang atau Key Kedaluwarsa pada Pengguna Awam:** Pengguna baru gagal terhubung dari luar rumah. *Mitigasi:* Wizard 3-langkah interaktif dengan deteksi status otomatis dan tombol langsung ke installer/Play Store Tailscale.

## 10. Asumsi & Pertanyaan Terbuka
1. Pengguna bersedia memasang aplikasi resmi Tailscale pada PC dan HP untuk akses internet luar jaringan lokal.
2. Kontrol tambahan Fase 2 (`M7`, `M8`) disiapkan arsitekturnya tanpa menghalangi rilis Fase 1.
3. Penandatanganan rilis Play Store (keystore produksi APK) tetap diaudit terpisah sebelum distribusi publik.
