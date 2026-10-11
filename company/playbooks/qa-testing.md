# Playbook: QA dan Verifikasi

Dimuat oleh: VP QA, Tester, Critic (code review), Auditor, Success Auditor.
Prinsip: yang menguji bukan yang membangun. Bukti bisa dijalankan ulang oleh orang lain.

## Daftar isi
1. Strategi tes (G2) · 2. Tester · 3. Critic (code review) · 4. Auditor · 5. Success Auditor · 6. Pelaporan dan bug · 7. Anti-pola · 8. Selesai bila

## 1. Strategi tes (`company/qa/TEST-STRATEGY.md`, dibuat VP QA di P2)
- **Piramida**: banyak tes unit (logika murni), cukup tes integrasi (DB, API, batas), sedikit e2e (alur kritis dari PRD).
- Daftar alur kritis dan risiko tertinggi; target cakupan modul inti sekitar 70% atau lebih sebagai pemandu, bukan tujuan.
- Data uji: fixture dan factory fiktif; tidak ada data nyata.
- Kebijakan tes rapuh (flaky): tandai dan buat bug; jangan menghapus atau menonaktifkan diam-diam.
- Alat standar per stack: pytest, Vitest/Jest + Testing Library, Playwright (Chromium default), PHPUnit; pemindai aksesibilitas (axe) pada e2e UI; Lighthouse bila ada.
- Pembagian: tes inti ditulis Tester independen; Builder menulis tes unit untuk kodenya sendiri tetapi tidak menjadi satu-satunya penguji logika inti.

## 2. Tester
1. Baca kriteria penerimaan dan kontrak, bukan implementasi, saat merancang kasus; tambahkan kasus negatif dan batas (kosong, terpanjang, karakter khusus, bersamaan, tanpa izin, jaringan gagal).
2. Tulis tes, jalankan terhadap cabang cluster, lampirkan perintah dan keluaran.
3. Temukan bug: tulis `BUG-###.md` (template 06) dengan langkah reproduksi minimal. Jangan memperbaiki kode produk.
4. **Eksploratori**: piagam singkat ("Jelajahi X dengan Y untuk menemukan Z", 20-30 menit agen) lalu catat temuan.
5. e2e: satu alur kritis per tes, stabil (tunggu kondisi, bukan `sleep`), data terisolasi per tes.

## 3. Critic (code review)
Periksa: kebenaran terhadap kriteria, tes cukup dan bermakna, penanganan galat, keamanan dasar (input, otorisasi, rahasia), keterbacaan, kepatuhan kepemilikan dan kontrak (`gate_check.py drift`), kepatuhan design system (tanpa hard-coded). Keluaran: daftar temuan dengan severity dan saran; bukan menulis ulang kode.

## 4. Auditor (verifikasi ulang bukti)
- Pilih sampel perintah bukti dari laporan (minimal 30% dan semua yang kritis), jalankan ulang dari kondisi bersih, bandingkan hasil.
- **Cari anti-pola** (00 bagian 7): bandingkan tes terhadap cabang dasar: `git diff <base> -- <folder tes>` dan cari assert yang dihapus atau pola `skip`, `xfail`, `.only`, `todo` yang bertambah; cari fungsi yang hanya mengembalikan data statis; `except: pass`.
- Periksa bahwa angka di laporan (jumlah tes, cakupan) cocok dengan keluaran sebenarnya.
- Keluaran: laporan dengan status per butir (cocok, tidak cocok, tidak bisa diverifikasi).

## 5. Success Auditor (P6)
Mulai dari **clone bersih** di folder baru (`company/scratch/clean-clone` atau luar repo bila diizinkan), ikuti README tepat seperti pengguna baru, jalankan setiap kriteria di `08-exit-criteria-reporting.md`, catat tiap langkah yang gagal atau ambigu (itu bug dokumentasi). Tidak memakai file yang hanya ada di mesin pembangun.

## 6. Pelaporan dan bug
Laporan `company/reports/<id>.md` (template 06) dengan bukti. Bug lewat protokol (00 bagian 4); baris `Severity:`, `Status:`, `Cluster:` wajib karena dibaca `gate_check.py`. Verifikasi perbaikan oleh penemu atau Tester independen, bukan oleh pembuat perbaikan.

## 7. Anti-pola
Tes yang hanya memeriksa bahwa fungsi dipanggil; mock di dalam sistem; tes bergantung urutan atau waktu; `sleep` sebagai sinkronisasi; menandai lolos tanpa menjalankan; cakupan tinggi tanpa assert bermakna; Auditor yang hanya membaca laporan.

## 8. Selesai bila
Strategi tes ada (G2); tiap cluster punya tes independen dengan bukti; alur kritis e2e hijau; Auditor mengonfirmasi sampel; Success Auditor lulus dari clone bersih; tidak ada bug S1/S2 terbuka.
