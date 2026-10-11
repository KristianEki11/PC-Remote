# Playbook: UI Design

Dimuat oleh: Head of UI, Head of Design, UI Designer, UI Critic, Design Critic (juga dibaca Builder dan A11y Auditor).
Batas: UI **tidak** mengubah alur, urutan langkah, label navigasi, atau isi teks (itu UX). UI tidak menambah fitur.

## Daftar isi
1. Bahasa desain per platform · 2. Kalibrasi angka · 3. Prosedur · 4. Audit 5 lensa (UI Critic) · 5. Keluaran · 6. Checklist pra-pengiriman · 7. Anti-pola · 8. Skill terkait

## 1. Bahasa desain per platform (R9)
| Platform | Bahasa | Navigasi dan kontrol | Material | Ukuran target | Tipografi, satuan |
|---|---|---|---|---|---|
| Web, ekstensi, game | **apple** (Full of Design) | tab/bilah, sidebar, sheet; kontrol bergaya HIG | translusen/blur secukupnya, kedalaman halus | 44 px (sentuh), 28 px (pointer) sebagai standar; minimum WCAG 24 px | tumpukan font sistem atau Inter/Geist (lisensi terbuka); `rem` |
| Android | **material** (Material 3) | navigation bar/rail/drawer, top app bar, FAB, bottom sheet | tonal elevation, warna dinamis bila didukung | 48 dp, jarak antar-target 8 dp | tipografi Material 3 (Roboto atau font brand); `sp` |
| Windows | **fluent** (Fluent 2/WinUI) | NavigationView, bilah judul, menu dan pintasan | Mica untuk latar jendela, Acrylic untuk permukaan sementara | ikuti pedoman Fluent/WinUI terkini (verifikasi angkanya), semua kontrol dapat dijangkau keyboard | Segoe UI Variable; skala teks sistem |
| custom | token brand | sesuai ADR | sesuai ADR | tidak lebih kecil dari angka WCAG | sesuai ADR |

Fondasi Full of Design berlaku di semua platform: kejernihan dan fokus konten, hierarki tipografi, ritme spasi 4/8, kontras, motion 150-300 ms, kerapian ikon, paritas tema terang/gelap. Yang berbeda hanya konvensi platform. Putuskan di ADR-lite "bahasa desain".

## 2. Kalibrasi angka (dipakai audit bila catatan lain berbeda)
- **Kontras** (WCAG 2.2 AA): teks normal 4,5:1; teks besar (minimal 18 pt reguler atau 14 pt tebal) 3:1; komponen UI, indikator fokus, dan grafik bermakna 3:1 terhadap latar.
- **Target**: WCAG 2.5.8 minimal 24x24 CSS px; standar Apple-first 44x44 pt (sentuh) dan 28x28 pt (pointer); Android 48x48 dp.
- **Umpan balik** terlihat paling lama sekitar 100 ms; transisi 150-300 ms; hormati `prefers-reduced-motion`.
- **Teks** mengikuti preferensi pengguna (rem/sp/skala sistem), tetap terbaca pada perbesaran 200% tanpa kehilangan konten.

## 3. Prosedur (P2)
1. **Arah visual.** Jalankan design engine (`design_engine.py run -- "<produk> <industri> <kata kunci>" --design-system -p "<Project>" --persist`), tinjau opsi style, palet, tipografi (`-d style|color|typography`), pilih satu arah berdasarkan audiens, merek, dan konteks pakai (bukan selera). Catat ADR-lite.
2. **Layout.** Grid dan ritme 4/8, breakpoint (web), kepadatan, margin, area aman (seluler), skala spasi.
3. **Tipografi.** Skala (display, title, headline, body, callout, footnote, caption), font + fallback + lisensi, tinggi baris body 1,4-1,6, lebar baris 45-75 karakter.
4. **Warna.** Token semantik (nama di `design-system.md`), tema terang dan gelap sejajar, warna bukan satu-satunya pembawa makna. Jalankan `contrast.py --tokens`.
5. **Komponen dan state.** Inventaris dari wireframe UX; tiap komponen: varian, ukuran, state (default, hover, fokus, aktif, nonaktif, loading, galat, kosong), anatomi.
6. **Ikon dan gambar.** Satu keluarga ikon dengan lisensi jelas (Lucide, SF Symbols hanya untuk platform Apple, Material Symbols), stroke konsisten, tanpa emoji sebagai ikon; gambar dengan rasio tetap dan teks alternatif.
7. **Motion.** 150-300 ms, easing sesuai platform, alternatif untuk reduced motion.
8. **Prototipe hi-fi** statis (HTML/CSS memakai variabel token, terang dan gelap) per layar kritis di `company/design/ui/prototypes/`.
9. **`ui-spec.md`**: arah visual (rujuk ADR), ringkasan token, komponen dan state, aturan layout, daftar layar + prototipe, pengecualian. Serahkan token dan komponen ke Design System Engineer, serahkan spesifikasi ke Builder.

## 4. Audit 5 lensa (UI Critic, Design Critic)
Periksa urut lensa; gunakan versi lengkap di SKILL.md Full of Design dengan angka dari bagian 2.
| Lensa | Fokus | Severity default |
|---|---|---|
| 1 Aksesibilitas | skala teks, kontras, target, warna bukan satu-satunya makna, keyboard dan pembaca layar | S2 (S1 bila alur inti tak bisa dipakai pengguna keyboard/pembaca layar) |
| 2 Konvensi platform | navigasi, kontrol, perilaku sesuai bahasa desain (apple/material/fluent), tema terang/gelap berparitas | S2 bila merusak alur inti, selain itu S3 |
| 3 Hierarki dan layout | skala tipografi, ritme spasi, kepadatan, aksi utama jelas | S3 (S2 bila aksi utama tak terlihat) |
| 4 Material dan craft | blur/translusen dengan kontras cukup, ikon konsisten, garis halus di dua tema | S3 |
| 5 Motion | umpan balik cepat, durasi wajar, `prefers-reduced-motion` dihormati | S3 (S2 bila reduced motion diabaikan pada gerak besar) |

Bukti: tangkapan layar Playwright (3 viewport x 2 tema untuk web), keluaran `contrast.py`, daftar warna/ukuran hard-coded di luar token (cari hex dan px lepas), uji keyboard. Keluaran `company/design/ui/ui-review-###.md` (tabel: ID, lensa, layar, bukti, severity, rekomendasi).

## 5. Keluaran
`design-system/MASTER.md` dan token (lihat `design-system.md`), `company/design/ui/ui-spec.md`, `company/design/ui/prototypes/`, `company/design/ui/ui-review-###.md`. Dibekukan di G2.

## 6. Checklist pra-pengiriman (untuk Builder sebelum lapor DONE; versi Full of Design + tambahan)
- Ikon satu keluarga, tanpa emoji; stroke konsisten.
- Target interaktif sesuai angka platform; kontras terpenuhi di dua tema; token semantik, bukan hex.
- Area aman dihormati; tidak ada teks meluap saat skala teks diperbesar.
- Semua state komponen ada; tema terang dan gelap diuji; motion halus dan `prefers-reduced-motion` dihormati.
- Teks UI tahan ekspansi (+30% untuk terjemahan); tidak ada lorem ipsum; fokus keyboard terlihat.

## 7. Anti-pola
Mengubah alur atau isi teks; warna/ukuran hard-coded; emoji sebagai ikon; dekorasi yang menurunkan keterbacaan; mengabaikan tema gelap; makna hanya lewat warna; menyalin teks panduan pihak ketiga ke repo; satu arah visual dicampur dua `MASTER.md`.

## 8. Skill terkait
Full of Design lewat `design_engine.py` (`run`, `hig`), `references/cross-platform.md` di folder Full of Design (pemetaan ke Flutter/RN/Tauri/Electron/SwiftUI), dan `frontend-design` bila terpasang. Hasil kueri dipersist; keputusan ditulis dengan kata sendiri.
