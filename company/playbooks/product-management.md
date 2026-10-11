# Playbook: Product Management

Dimuat oleh: CPO, Head of Strategy & Research, Spec Writer. Menghasilkan PRD dan backlog yang menjadi sumber kebenaran lingkup (R7).

## Daftar isi
1. Masukan · 2. PRD · 3. User story dan kriteria penerimaan · 4. Backlog dan cluster · 5. Prioritas dan scope guard · 6. Misi non-BUILD · 7. Anti-pola · 8. Selesai bila

## 1. Masukan
Brief (`COMPANY.md`/chat), `CONSTRAINTS`, `REFERENCES`, catatan riset `R-###`, asumsi intake di `progress.md`. Bila brief ambigu, tulis asumsi dan pertanyaan terbuka dengan default, bukan menunggu (E0 hanya untuk yang benar-benar terblokir).

## 2. PRD (`company/product/PRD.md`, template 06)
- **Masalah dan pengguna**: siapa, apa yang sulit, bagaimana mereka melakukannya sekarang.
- **Tujuan dan non-tujuan**: yang akan dan tidak akan dicapai.
- **Lingkup MVP**: daftar fitur wajib; sisanya ke v2. MVP = versi terkecil yang menyelesaikan masalah inti dan bisa dipakai nyata.
- **Alur kritis**: 3-7 alur yang wajib berjalan (menjadi dasar tes e2e dan Success Auditor).
- **NFR**: kinerja, keamanan, aksesibilitas, i18n, kompatibilitas, ketersediaan, batas biaya.
- **Asumsi, risiko, pertanyaan terbuka** (masing-masing dengan default dan pemilik keputusan).
- **Konten faktual yang belum ada**: tandai `[ISI: ...]`; jangan dikarang.

## 3. User story dan kriteria penerimaan
Format: "Sebagai <peran>, saya ingin <tujuan>, agar <manfaat>." Kriteria penerimaan Given/When/Then yang **terukur dan dapat dites**, termasuk kasus negatif dan batas. Setiap story menyebut peran yang berhak dan data yang disentuh. Kriteria menjadi dasar tes Tester dan bagian DoD cluster.

## 4. Backlog dan cluster (`company/product/backlog.md`)
Cluster = irisan vertikal yang bisa dibangun dan diuji end-to-end (aturan di `02-intake-staffing.md` bagian 6). Per cluster: ID, tujuan, story, kontrak yang disentuh, dependensi, ukuran S/M/L, path milik. Cluster fondasi lebih dulu. Bagian **v2** menampung semua ide di luar MVP.

## 5. Prioritas dan scope guard
Urutkan dengan MoSCoW (Must, Should, Could, Won't-now) dan dependensi. Ide baru di tengah jalan: tulis ke v2; masuk MVP hanya lewat keputusan tertulis yang menyebut apa yang dikeluarkan sebagai gantinya. Head menolak pekerjaan yang tidak ada di PRD.

## 6. Misi non-BUILD
- **ENHANCE**: PRD berupa delta (fitur baru, area yang disentuh, risiko regresi); backlog memetakan ke modul yang ada.
- **REDESIGN**: dokumen "apa yang berubah dan apa yang tetap" (alur, konten, fitur, data), alasan, metrik keberhasilan visual/UX.
- **MIGRATE**: tujuan, versi asal dan tujuan, batas waktu henti yang diterima, kriteria sukses (paritas), rencana rollback.
- **FIX**: `INVENTORY.md` (bug dilaporkan, hasil reproduksi, severity awal). **AUDIT**: `SCOPE.md` (template 06).

## 7. Anti-pola
Fitur tanpa pengguna atau tanpa kriteria penerimaan; kriteria samar ("cepat", "mudah"); MVP yang membengkak; mengarang data bisnis; backlog tanpa dependensi; mengubah lingkup tanpa catatan.

## 8. Selesai bila
PRD dan backlog lengkap dan konsisten; setiap story punya kriteria terukur; cluster dan dependensi jelas; asumsi dan pertanyaan terbuka tercatat dengan default; `gate_check.py G1` lulus.
