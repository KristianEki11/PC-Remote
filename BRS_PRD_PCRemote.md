# BRS / PRD — PC Remote Controller
**Business Requirements Specification & Product Requirements Document**

---

## Informasi Dokumen

| Field | Detail |
|---|---|
| Nama Proyek | PC Remote Controller |
| Versi Dokumen | 1.0.0 |
| Tanggal | 2026-05-25 |
| Platform Target | Android (Flutter) + Windows (Python Server) |
| IDE | Antigravity by Google |
| AI Models | Claude Opus, Gemini |
| Status | Draft — Tahap Perancangan |

---

## 1. Latar Belakang & Tujuan

### 1.1 Latar Belakang
Pengguna membutuhkan cara untuk mengontrol PC Windows dari smartphone Android melalui jaringan WiFi lokal, tanpa bergantung pada aplikasi pihak ketiga berbayar atau koneksi internet. Aplikasi ini dibuat untuk keperluan pribadi dengan fokus pada kemudahan, tampilan yang baik, dan keamanan dasar.

### 1.2 Tujuan Produk
Membangun aplikasi Android berbasis Flutter yang berkomunikasi dengan server Python di PC melalui jaringan WiFi lokal, memungkinkan pengguna mengontrol audio, membuka browser, mengunci PC, dan mengelola media tanpa menyentuh PC secara fisik.

### 1.3 Pengguna Target
- Pengguna tunggal (personal use)
- Familiar dengan teknologi dan pengembangan aplikasi
- Menggunakan Android dan PC Windows di jaringan WiFi yang sama

---

## 2. Ruang Lingkup

### 2.1 Dalam Lingkup (In-Scope)
- Kontrol volume audio Windows
- Lock PC
- Membuka URL di Microsoft Edge
- Kontrol media (play/pause/next/prev)
- Power management (sleep, shutdown, restart)
- Autentikasi berbasis PIN
- Dashboard UI satu layar

### 2.2 Di Luar Lingkup (Out-of-Scope)
- Unlock PC secara programatik (keterbatasan Windows API)
- Koneksi via internet / luar jaringan lokal
- Multi-user / multi-device
- Platform selain Android dan Windows
- Integrasi dengan layanan cloud

---

## 3. Business Requirements (BRS)

### BR-01 — Koneksi Lokal
Aplikasi harus dapat terhubung ke server PC melalui WiFi lokal menggunakan IP address dan port yang dikonfigurasi pengguna. Koneksi harus diverifikasi sebelum menampilkan dashboard.

### BR-02 — Keamanan PIN
Akses ke seluruh fitur harus dilindungi dengan PIN 4–6 digit. PIN disimpan secara lokal di server PC (file `.env`) dan tidak pernah dikirim ke luar jaringan lokal.

### BR-03 — Kontrol Audio
Pengguna harus dapat melihat dan mengubah volume master Windows, serta melakukan mute/unmute, langsung dari aplikasi Android.

### BR-04 — Kontrol Browser
Pengguna harus dapat mengirim URL dari aplikasi Android untuk dibuka di Microsoft Edge pada PC. Harus mendukung URL YouTube maupun URL umum lainnya.

### BR-05 — Lock PC
Pengguna harus dapat mengunci PC Windows dari aplikasi Android dengan satu ketukan tombol.

### BR-06 — Kontrol Media
Pengguna harus dapat mengontrol pemutaran media aktif di PC (play/pause, lagu berikutnya, lagu sebelumnya) dari aplikasi Android.

### BR-07 — Power Management
Pengguna harus dapat mengirim perintah sleep, restart, dan shutdown ke PC dari aplikasi Android.

### BR-08 — Status Koneksi Real-Time
Aplikasi harus menampilkan status koneksi ke server secara real-time (terhubung / terputus) pada dashboard.

---

## 4. Product Requirements (PRD)

### 4.1 Komponen Sistem

#### 4.1.1 Server PC (Python + FastAPI)
- Berjalan sebagai background process di PC Windows
- Menerima request HTTP dari aplikasi Android
- Mengeksekusi perintah sistem Windows via PowerShell / subprocess
- Validasi setiap request menggunakan Bearer token session

#### 4.1.2 Aplikasi Android (Flutter)
- Satu layar login untuk input IP dan PIN
- Satu layar dashboard utama dengan semua kontrol
- Komunikasi ke server via HTTP (REST)
- Menyimpan IP dan session token di `shared_preferences`

---

### 4.2 Functional Requirements

#### FR-01 — Login Screen
| ID | Requirement |
|---|---|
| FR-01.1 | Tampilkan field input untuk IP Address server |
| FR-01.2 | Tampilkan field input PIN (tersembunyi / obscured) |
| FR-01.3 | Tombol "Hubungkan" mengirim POST ke `/auth` |
| FR-01.4 | Jika berhasil, simpan token dan navigasi ke Dashboard |
| FR-01.5 | Jika gagal, tampilkan pesan error yang jelas |
| FR-01.6 | IP terakhir digunakan tersimpan otomatis |

#### FR-02 — Dashboard Screen
| ID | Requirement |
|---|---|
| FR-02.1 | Tampilkan status koneksi (ikon + teks) di header |
| FR-02.2 | Semua kontrol tampil dalam satu layar (scrollable jika perlu) |
| FR-02.3 | Setiap kartu fitur memiliki label jelas |
| FR-02.4 | Tombol logout tersedia di header |

#### FR-03 — Audio Control Card
| ID | Requirement |
|---|---|
| FR-03.1 | Slider volume (0–100) menampilkan nilai saat ini |
| FR-03.2 | Perubahan slider mengirim POST ke `/audio/volume` |
| FR-03.3 | Tombol mute/unmute toggle dengan ikon yang berubah |
| FR-03.4 | Volume saat ini diambil via GET `/audio/volume` saat load |

#### FR-04 — Media Control Card
| ID | Requirement |
|---|---|
| FR-04.1 | Tombol ⏮ Prev mengirim POST ke `/media/prev` |
| FR-04.2 | Tombol ⏯ Play/Pause mengirim POST ke `/media/playpause` |
| FR-04.3 | Tombol ⏭ Next mengirim POST ke `/media/next` |

#### FR-05 — Browser Card
| ID | Requirement |
|---|---|
| FR-05.1 | Field input teks untuk URL |
| FR-05.2 | Tombol cepat "YouTube" mengisi field dengan `https://youtube.com` |
| FR-05.3 | Tombol "Buka di Edge" mengirim POST ke `/browser/open` dengan URL |
| FR-05.4 | Validasi URL sebelum dikirim (harus diawali `http://` atau `https://`) |

#### FR-06 — System Control Card
| ID | Requirement |
|---|---|
| FR-06.1 | Tombol "Lock PC" mengirim POST ke `/system/lock` |
| FR-06.2 | Tombol "Sleep" mengirim POST ke `/system/sleep` |
| FR-06.3 | Tombol "Restart" menampilkan dialog konfirmasi sebelum POST ke `/system/restart` |
| FR-06.4 | Tombol "Shutdown" menampilkan dialog konfirmasi sebelum POST ke `/system/shutdown` |

---

### 4.3 API Specification

#### Base URL
```
http://{IP_ADDRESS}:8000
```

#### Authentication
Semua endpoint (kecuali `/auth`) membutuhkan header:
```
Authorization: Bearer {token}
```

#### Endpoints

| Method | Endpoint | Body | Response |
|---|---|---|---|
| POST | `/auth` | `{"pin": "1234"}` | `{"token": "abc123"}` |
| GET | `/audio/volume` | — | `{"level": 70, "muted": false}` |
| POST | `/audio/volume` | `{"level": 80}` | `{"success": true}` |
| POST | `/audio/mute` | — | `{"muted": true}` |
| POST | `/media/playpause` | — | `{"success": true}` |
| POST | `/media/next` | — | `{"success": true}` |
| POST | `/media/prev` | — | `{"success": true}` |
| POST | `/browser/open` | `{"url": "https://..."}` | `{"success": true}` |
| POST | `/system/lock` | — | `{"success": true}` |
| POST | `/system/sleep` | — | `{"success": true}` |
| POST | `/system/restart` | — | `{"success": true}` |
| POST | `/system/shutdown` | — | `{"success": true}` |

---

### 4.4 Non-Functional Requirements

| ID | Requirement |
|---|---|
| NFR-01 | Latensi respons server < 500ms untuk semua perintah |
| NFR-02 | Server hanya menerima koneksi dari subnet lokal (`192.168.x.x` / `10.x.x.x`) |
| NFR-03 | Token session kedaluwarsa setelah 24 jam atau saat server restart |
| NFR-04 | Aplikasi harus menampilkan error yang ramah pengguna jika server tidak dapat dijangkau |
| NFR-05 | Ukuran APK tidak melebihi 20MB |
| NFR-06 | Mendukung Android 8.0 (API level 26) ke atas |

---

### 4.5 UI / UX Requirements

| ID | Requirement |
|---|---|
| UX-01 | Tema dark mode sebagai default |
| UX-02 | Menggunakan Material Design 3 |
| UX-03 | Warna aksen: biru modern (`#2196F3` atau sejenisnya) |
| UX-04 | Setiap aksi destruktif (shutdown, restart) wajib konfirmasi dialog |
| UX-05 | Feedback visual (loading indicator) saat menunggu respons server |
| UX-06 | Ikon yang jelas dan intuitif di setiap tombol |

---

### 4.6 Tech Stack

| Komponen | Teknologi |
|---|---|
| Server bahasa | Python 3.10+ |
| Server framework | FastAPI + Uvicorn |
| Audio Windows | `pycaw` |
| Media keys | `keyboard` / `pyautogui` |
| Konfigurasi server | `python-dotenv` |
| App framework | Flutter (Dart) |
| HTTP client Flutter | `http` package |
| Penyimpanan lokal | `shared_preferences` |

---

### 4.7 Struktur Folder

```
pc-remote/
├── server/
│   ├── main.py
│   ├── auth.py
│   ├── .env                  ← PIN disimpan di sini
│   ├── requirements.txt
│   └── routes/
│       ├── audio.py
│       ├── media.py
│       ├── browser.py
│       └── system.py
│
└── app/
    └── lib/
        ├── main.dart
        ├── screens/
        │   ├── login_screen.dart
        │   └── dashboard_screen.dart
        ├── widgets/
        │   ├── audio_card.dart
        │   ├── media_card.dart
        │   ├── browser_card.dart
        │   └── system_card.dart
        └── services/
            └── api_service.dart
```

---

## 5. Batasan & Risiko

| # | Risiko | Mitigasi |
|---|---|---|
| 1 | Unlock PC tidak bisa dilakukan via API Windows | Fitur unlock dihapus; hanya Lock yang didukung |
| 2 | IP server berubah saat DHCP refresh | Gunakan IP statis di router untuk PC |
| 3 | Firewall Windows memblokir port 8000 | Tambahkan instruksi izin firewall di dokumentasi setup |
| 4 | `pycaw` tidak support semua versi Windows | Fallback ke `nircmd.exe` jika pycaw gagal |

---

## 6. Milestone

| Fase | Target | Deskripsi |
|---|---|---|
| Fase 1 | Setup | Server Python berjalan, endpoint `/auth` dan `/audio` berfungsi |
| Fase 2 | Core Features | Semua endpoint server selesai dan tertes via Postman |
| Fase 3 | Flutter App | Login screen + Dashboard dengan semua kartu |
| Fase 4 | Integrasi | App terhubung ke server, semua fitur berfungsi end-to-end |
| Fase 5 | Polish | UI refinement, error handling, testing menyeluruh |

---

*Dokumen ini adalah living document — perbarui setiap kali ada perubahan requirement selama pengembangan.*
