# Playbook: Platform dan DevOps

Dimuat oleh: VP Platform, Platform Engineer (mode datar), Release Manager; dibaca Builder dan Success Auditor.

## Daftar isi
1. Tujuan · 2. Scaffold repo · 3. Lingkungan dan rahasia · 4. Docker dan Compose (web/API) · 5. CI · 6. Build per archetype · 7. Cadangan, pemulihan, rollback · 8. Rantai pasok dan artefak · 9. Anti-pola · 10. Selesai bila

## 1. Tujuan
Siapapun (termasuk Success Auditor dari clone bersih) dapat membangun, menjalankan, dan menguji proyek dengan perintah yang sedikit dan terdokumentasi. Dikerjakan di P2 (scaffold), P4 (build bersih), P6 (artefak).

## 2. Scaffold repo (P2)
Tata letak sesuai ADR; file kunci: `README.md` (paling banyak 5 perintah setup), `.gitignore` (termasuk `company/scratch/`, `company/.trash/`, `.env`), `.env.example`, `.editorconfig`, lockfile terkomit, versi runtime dikunci (`.nvmrc`/`.python-version`/`go.mod`/`rust-toolchain`). Skrip tugas seragam (`npm run`, `make`, atau `scripts/*.ps1|sh`): `lint`, `typecheck`, `test`, `build`, `dev`.

## 3. Lingkungan dan rahasia
Konfigurasi dari variabel lingkungan yang divalidasi saat start; hanya `.env.example` yang di-commit (nilai fiktif). Tidak ada rahasia di image, log, atau laporan. Rahasia produksi adalah langkah manusia. Pisahkan konfigurasi dev, test, dan prod.

## 4. Docker dan Compose (web/API)
- Image minimal, pengguna non-root, tanpa rahasia di layer, multi-stage untuk memperkecil ukuran, versi dasar dikunci.
- `docker compose up` dari clone bersih menghasilkan sistem sehat: `healthcheck` pada tiap layanan, `depends_on` dengan kondisi sehat, migrasi dan seed fiktif berjalan otomatis atau lewat satu perintah.
- Data di volume bernama; port tidak bentrok; log ke stdout.
- Docker tidak tersedia: bukti alternatif (menjalankan layanan langsung) dan tandai `UNVERIFIED` pada butir compose.

## 5. CI
Berkas alur (mis. GitHub Actions) yang menjalankan lint, typecheck, test, build pada setiap push/PR, dengan cache dependensi dan izin minimal (`permissions:` seminimal mungkin). Tanpa rahasia di berkas alur. Skrip lokal yang sama dapat dijalankan manusia tanpa CI.

## 6. Build per archetype
| Archetype | Keluaran build |
|---|---|
| web-saas, web-app, api-service | image/compose + migrasi + seed; dokumentasi deploy agnostik hosting |
| web-marketing | build statis/SSG + konfigurasi hosting; DNS dan SSL adalah langkah manusia |
| desktop-windows | installer atau zip portabel + checksum (lihat `desktop-windows.md`) |
| android-app | APK/AAB debug, konfigurasi rilis tanpa kunci produksi (lihat `mobile-android.md`) |
| browser-extension | zip siap unggah + `manifest` tervalidasi |
| bot-automation | proses/kontainer dengan mode dry-run default, unit layanan (systemd/Task Scheduler) contoh |
| cli-library | paket (wheel/tarball/npm pack) dan biner bila perlu, dites instal dari artefak bersih |
| game-browser, data-ai | build statis / lingkungan terkunci dengan data sampel |

## 7. Cadangan, pemulihan, rollback
Skrip cadangan dan pulihkan DB diuji; migrasi punya rencana rollback; dokumentasi singkat "jika gagal setelah deploy, lakukan ini". Untuk MIGRATE: latihan rollback terdokumentasi.

## 8. Rantai pasok dan artefak
Dependensi dikunci; audit paket dijalankan; lisensi diperiksa (Privacy Reviewer atau Security Auditor); artefak rilis disertai SHA-256 dan (opsional) SBOM. Jangan menjalankan skrip instalasi dari sumber tak dikenal.

## 9. Anti-pola
README yang butuh lebih dari 5 langkah atau langkah tersembunyi; hanya berjalan di mesin pembuat; rahasia di repo/image; image berjalan sebagai root tanpa alasan; dependensi tanpa lockfile; CI dengan izin berlebihan; "works on my machine" tanpa uji clone bersih.

## 10. Selesai bila
Dari clone bersih: install, build, test, dan run mengikuti README berhasil (bukti perintah dan keluaran); `.env.example` lengkap; CI lokal/berkas alur ada; artefak + checksum tersedia; butir yang tak bisa dijalankan agen ditandai `UNVERIFIED`.
