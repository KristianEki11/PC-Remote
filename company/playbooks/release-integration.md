# Playbook: Release dan Integrasi

Dimuat oleh: Release Manager, Release & Docs (mode datar), Platform Engineer, Auditor, Success Auditor. Release Manager adalah satu-satunya yang menggabungkan cabang ke main.

## Daftar isi
1. Merge berurutan (P4) · 2. Konflik · 3. Versi dan changelog · 4. Artefak per archetype · 5. Checklist rilis · 6. Rollback · 7. Laporan rilis · 8. Anti-pola · 9. Selesai bila

## 1. Merge berurutan (P4)
1. Urutkan cabang `co/*` mengikuti graf dependensi (fondasi dulu).
2. Per cabang: pastikan laporan `STATUS: DONE`, ownership tidak dilanggar, `gate_check.py drift` bersih. Merge ke main (non-fast-forward bila ingin riwayat jelas), jalankan **seluruh suite** (lint, typecheck, test, build).
3. Merah setelah merge: batalkan merge itu (cabang kembali ke pemiliknya) dan buat bug; **main selalu hijau**.
4. Setelah semua merge: jalankan e2e smoke, build/run dari nol (compose up, build artefak, atau `--help`/selftest), dan `drift` kontrak. Lanjut `janitor.py worktrees --apply`.

## 2. Konflik
Konflik berarti pelanggaran ownership atau kontrak yang tidak sinkron. Selesaikan dengan pemilik path yang bersangkutan (lewat Head); jangan menyelesaikan konflik logika tanpa mereka. Catat penyebab di laporan agar kepemilikan diperjelas.

## 3. Versi dan changelog
Semver (`MAJOR.MINOR.PATCH`); rilis awal `0.1.0` atau `1.0.0` sesuai PRD. `CHANGELOG.md` ala Keep a Changelog (Added, Changed, Fixed, Security, Known issues). Tag **lokal** (`git tag`), tidak di-push.

## 4. Artefak per archetype
| Archetype | Artefak rilis |
|---|---|
| web-saas, web-app, api-service | build + `docker compose` terverifikasi, migrasi, panduan deploy agnostik |
| web-marketing | build statis + panduan hosting, daftar placeholder `[ISI: ...]` |
| desktop-windows | installer atau zip portabel + SHA-256 |
| android-app | APK/AAB debug + panduan penandatanganan rilis (kunci oleh manusia) |
| browser-extension | zip siap unggah + teks deskripsi store + justifikasi izin |
| bot-automation | paket/kontainer + panduan mode live (langkah manusia) |
| cli-library | paket (wheel/tarball/npm pack) + biner bila ada |
| game-browser, data-ai | build statis / lingkungan terkunci + data sampel |
Semua artefak disertai checksum SHA-256 dan cara memverifikasinya.

## 5. Checklist rilis
Kriteria selesai (08) lulus; tidak ada S1/S2 terbuka; README diuji dari clone bersih oleh Success Auditor; `.env.example` lengkap; lisensi dan atribusi dipenuhi; dokumen kebijakan (bila relevan) berstatus draf untuk manusia; laporan Janitor final ada; daftar **langkah manusia** tersisa (deploy, domain/DNS, secret produksi, tanda tangan kode, publikasi store, mode live).

## 6. Rollback
Setiap rilis punya cara kembali ke versi sebelumnya (tag sebelumnya, migrasi mundur, atau cadangan data). Untuk MIGRATE, latihan rollback dijalankan dan dicatat. Migrasi tanpa rencana rollback tidak boleh dirilis.

## 7. Laporan rilis (`company/RELEASE_REPORT.md`, template 06)
Verdict (`GO` / `GO-dengan-known-issues` / `NO-GO`) + tabel kriteria dengan bukti + known issues + langkah manusia tersisa + cara menjalankan + artefak dan checksum. Verdict ditetapkan Board setelah Auditor dan Success Auditor.

## 8. Anti-pola
Merge tanpa suite penuh; main merah dibiarkan; memaksa fast-forward di atas konflik yang belum dipahami; artefak tanpa checksum; rilis tanpa rencana rollback; `git push`; menyembunyikan known issues.

## 9. Selesai bila
Semua cabang merged dan main hijau; artefak dan checksum ada; `RELEASE_REPORT.md` dengan baris `Verdict:`; `gate_check.py G4` dan `G6` lulus.
