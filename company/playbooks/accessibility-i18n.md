# Playbook: Aksesibilitas dan i18n

Dimuat oleh: A11y Auditor, UX Designer, UI Critic, Builder FE/Android/Desktop, Tester.
Standar acuan: **WCAG 2.2 level AA** untuk web dan sebagai pedoman untuk Android/Windows. Fondasi Full of Design (kontras, target, skala teks, keyboard) sejalan dengan ini.

## Daftar isi
1. Checklist WCAG 2.2 AA · 2. Cara menguji · 3. Platform non-web · 4. i18n dan l10n · 5. Severity · 6. Anti-pola · 7. Selesai bila

## 1. Checklist WCAG 2.2 AA (ringkas, per prinsip)
**Dapat dipersepsi**
- Teks alternatif untuk gambar bermakna; dekoratif dikosongkan. Video/audio punya teks atau transkrip bila relevan.
- Kontras teks 4,5:1 (teks besar 3:1); komponen UI dan indikator fokus 3:1 (1.4.3, 1.4.11). Info tidak hanya lewat warna (1.4.1).
- Teks dapat diperbesar 200% tanpa kehilangan fungsi, dan konten tetap terbaca pada lebar 320 CSS px tanpa gulir dua arah (1.4.4, 1.4.10). Jarak teks dapat diubah (1.4.12).
**Dapat dioperasikan**
- Semua fungsi dapat dijangkau keyboard, tanpa jebakan fokus (2.1.1, 2.1.2); urutan fokus logis; fokus terlihat dan tidak tertutup elemen lain (2.4.7, 2.4.11).
- Tautan lompat ke konten utama; judul halaman dan heading bermakna (2.4.1, 2.4.2, 2.4.6).
- Target minimal 24x24 CSS px (2.5.8); alternatif non-drag untuk gerakan seret (2.5.7); tidak ada kilatan berbahaya, gerak dapat dihentikan (2.3.1, 2.2.2).
**Dapat dipahami**
- Bahasa halaman diset; navigasi dan bantuan konsisten (3.1.1, 3.2.3, 3.2.6).
- Formulir: label jelas, galat dijelaskan dan disarankan perbaikannya, tidak meminta ulang info yang sama (3.3.1, 3.3.3, 3.3.7). Autentikasi tidak mengandalkan uji kognitif tanpa alternatif (3.3.8).
**Kokoh**
- HTML valid dan semantik; nama, peran, nilai tersedia untuk teknologi bantu; perubahan status diumumkan (4.1.2, 4.1.3).

Aturan ARIA pertama: gunakan elemen native bila ada. ARIA yang salah lebih buruk daripada tanpa ARIA.

## 2. Cara menguji (web)
1. **Otomatis**: axe-core lewat Playwright pada tiap layar kritis (terang dan gelap); Lighthouse aksesibilitas bila tersedia. Otomatis hanya menangkap sebagian masalah.
2. **Keyboard**: telusuri seluruh alur kritis hanya dengan Tab, Shift+Tab, Enter, Space, Esc, panah; catat jebakan fokus dan fokus tak terlihat.
3. **Kontras**: `python "<SKILL_DIR>/scripts/contrast.py" --tokens design-system/tokens.json` plus periksa teks di atas gambar/gradien secara manual.
4. **Zoom dan reflow**: 200% dan lebar 320 px; teks diperbesar sistem.
5. **Pembaca layar** (bila ada) atau pemeriksaan pohon aksesibilitas (snapshot Playwright): nama dan peran tiap kontrol, urutan baca, live region.
6. **Reduced motion** dan mode kontras tinggi (`forced-colors`).

## 3. Platform non-web
- **Android**: target 48 dp, `contentDescription`/semantics untuk TalkBack, urutan fokus, skala teks sistem (sp) sampai maksimal, tidak bergantung warna, uji dengan Accessibility Scanner bila ada perangkat/emulator (`UNVERIFIED` bila tidak).
- **Windows**: nama dan peran UI Automation, navigasi keyboard penuh, tema kontras tinggi, skala DPI/teks, tes dengan Narrator bila bisa.
- **Ekstensi/game**: popup dan menu dapat dijangkau keyboard; game: kontrol dapat dipetakan ulang, pilihan kurangi gerak, tidak bergantung warna.

## 4. i18n dan l10n
- Semua teks lewat katalog pesan (ICU/format pustaka proyek); tanpa konkatenasi kalimat; plural dan gender mengikuti aturan bahasa.
- Format dengan `Intl`/API platform: tanggal ("6 Oktober 2026"), angka (1.234.567,89), mata uang (Rp 1.234.567), zona waktu WIB/WITA/WIT; simpan UTC.
- Desain tahan ekspansi teks 30% dan kata panjang; ikon tidak mengandung teks; dukung bahasa kedua (EN) bila diminta; siapkan RTL bila target membutuhkannya.
- Atribut bahasa (`lang`) benar; deteksi bahasa dengan fallback jelas; tautan pindah bahasa dapat dijangkau.

## 5. Severity
S1: alur inti tidak bisa diselesaikan dengan keyboard atau teknologi bantu. S2: pelanggaran AA pada alur kritis (kontras, label, fokus) atau pesan galat tidak aksesibel. S3: pelanggaran minor di layar sekunder, perbaikan kosmetik.

## 6. Anti-pola
`div` yang dibuat seperti tombol tanpa peran/keyboard; menghilangkan outline fokus tanpa pengganti; placeholder sebagai satu-satunya label; hanya mengandalkan skor Lighthouse; teks dalam gambar; string UI di-hard-code; format tanggal/angka ditulis tangan.

## 7. Selesai bila
Tidak ada temuan S1/S2 pada alur kritis; laporan memuat hasil otomatis + uji keyboard + kontras; i18n: tidak ada string hard-coded dan format lokal benar; butir yang butuh perangkat/pembaca layar yang tidak tersedia ditandai `UNVERIFIED`.
