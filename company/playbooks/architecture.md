# Playbook: Arsitektur dan Keputusan

Dimuat oleh: Chief Architect, API Contract Designer, Interface Contract Designer; dibaca Head dan Release Manager. Menghasilkan ADR, kontrak, peta kepemilikan, dan graf dependensi yang menjadi dasar kerja paralel tanpa bentrok.

## Daftar isi
1. ADR · 2. Kontrak per archetype · 3. Ownership map · 4. Graf dependensi cluster · 5. NFR dan tata letak repo · 6. Pemilihan teknologi · 7. Freeze dan Change Request · 8. Anti-pola · 9. Selesai bila

## 1. ADR
Keputusan Type-1 lewat Board menjadi `ADR-####.md`; Type-2 menjadi ADR-lite (template 06). Minimal di G1: pilihan stack, model data/tenancy atau penyimpanan, model autentikasi (bila ada), bahasa desain dasar, strategi hosting/distribusi. Satu keputusan per ADR; ADR yang digantikan ditandai `SUPERSEDED`, tidak dihapus.

## 2. Kontrak per archetype (dibekukan di G2)
| Archetype | Kontrak |
|---|---|
| web-saas, web-app, api-service | skema DB + migrasi awal, OpenAPI, model galat, matriks RBAC (+ tenant), skema event/webhook |
| desktop-windows | perintah/IPC UI-inti, format penyimpanan + versi migrasi, lokasi data, izin sistem |
| android-app | model data + skema Room + migrasi, graf navigasi, kontrak API (bila ada), daftar izin |
| browser-extension | `manifest.json` (izin), skema pesan antar komponen, skema storage |
| bot-automation | skema konfigurasi, skema state, format fixture, batas risiko dan aksi yang diizinkan |
| cli-library | spesifikasi CLI (perintah, flag, kode keluar, format keluaran), tanda tangan API publik, kebijakan versi |
| game-browser | spesifikasi aturan (tabel keadaan/transisi), format simpanan, peta event |
| data-ai | skema data, format dan versi prompt, skema keluaran (JSON schema), set evaluasi |
| web-marketing | sitemap/IA, struktur konten, spesifikasi SEO |
Semua archetype ber-UI: token desain (`design-system/tokens.json`) ikut dibekukan.

## 3. Ownership map (`company/ownership.md`)
Tabel `| Path (glob) | Pemilik (agent-id) | Cluster |` (bentuk di `06-templates.md`). Aturan: satu path satu pemilik; cluster tidak berbagi file; kode bersama (util, tipe) dimiliki satu agen khusus (cluster fondasi) dan lainnya meminta lewat CR; konfigurasi root (package.json, pyproject, lockfile) dimiliki Platform. Jalankan `gate_check.py G2` untuk mendeteksi tumpang tindih.

## 4. Graf dependensi cluster
Diagram Mermaid (`graph TD`) cluster dan dependensinya, dengan jalur kritis ditandai. Dasar scheduler wave (03 bagian 4). Cluster yang bergantung pada kontrak yang belum beku tidak boleh dimulai.

## 5. NFR dan tata letak repo
- NFR terukur (kinerja, keamanan, aksesibilitas, ketersediaan, i18n, kompatibilitas) diturunkan dari PRD dan menjadi kriteria cluster.
- Tata letak repo sederhana dan konvensional untuk stack yang dipilih (monorepo vs terpisah diputuskan lewat ADR). Aturan batas modul (siapa boleh mengimpor siapa) ditulis dan, bila memungkinkan, diperiksa otomatis (lint impor).
- Strategi lingkungan: dev, test, prod; konfigurasi lewat env; migrasi data dan seed.

## 6. Pemilihan teknologi
Kriteria: kecocokan dengan keahlian pemilik (JS/TS, Python, PHP, SQL), kematangan dan pemeliharaan, lisensi, kesesuaian dengan hosting dan batasan klien, ukuran komunitas, keamanan rantai pasok, kemudahan diserahterimakan kepada orang lain. Pilih yang membosankan dan terbukti kecuali ada alasan tertulis. Fakta versi dan kompatibilitas dari riset bertanggal, bukan ingatan.

## 7. Freeze dan Change Request
Setelah kontrak lengkap: `gate_check.py freeze --files <semua kontrak>` lalu `G2`. Perubahan setelah itu: CR (template 06) yang dinilai Chief Architect: apakah akar masalah memang di kontrak, dampak ke cluster, rencana migrasi. Disetujui: ubah kontrak, `freeze` ulang, beri tahu agen terdampak. Ditolak: pemohon menyesuaikan implementasi.

## 8. Anti-pola
Kontrak ditulis setelah kode; kontrak yang tidak dibekukan; dua sumber kebenaran (skema di dua tempat); ownership berbagi file; menambah teknologi tanpa ADR; arsitektur berlebihan untuk proyek kecil; keputusan Type-1 tanpa Board.

## 9. Selesai bila
ADR G1 lengkap; kontrak lengkap dan dibekukan; ownership tanpa tumpang tindih; graf dependensi dan NFR ada; `gate_check.py G2` lulus.
