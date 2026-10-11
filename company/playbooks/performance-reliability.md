# Playbook: Performa dan Keandalan

Dimuat oleh: Performance & Reliability Engineer, VP QA, Builder, Platform. Prinsip: ukur dulu, optimalkan yang terbukti lambat, buktikan dengan angka sebelum dan sesudah.

## Daftar isi
1. Anggaran · 2. Cara mengukur · 3. Pola optimasi · 4. Keandalan · 5. Platform non-web · 6. Laporan · 7. Anti-pola · 8. Selesai bila

## 1. Anggaran (ditetapkan di G2, dicatat di `TEST-STRATEGY.md`)
- **Web (batas "baik" Core Web Vitals, persentil ke-75)**: LCP paling lama 2,5 detik; INP paling lama 200 ms; CLS paling besar 0,1. Tambahan: ukuran JS awal dan gambar per halaman ditentukan per proyek.
- **API**: p95 latensi per kelas endpoint (patokan awal: baca sederhana beberapa ratus ms; sesuaikan di PRD), tingkat galat, throughput target.
- **Mobile/desktop/game**: waktu mulai dingin, jank (frame time), memori, penggunaan baterai/CPU saat idle; game: anggaran frame (mis. 16,7 ms untuk 60 fps).
Angka tanpa dasar di PRD adalah asumsi dan ditandai begitu.

## 2. Cara mengukur
- Web: Lighthouse atau alat sejenis di lingkungan terkendali (build produksi), beberapa kali lari dan ambil median; Playwright trace untuk interaksi (INP). Bila alat tidak tersedia: `UNVERIFIED`.
- API: tes beban ringan (k6, Locust, atau skrip sederhana) pada instance lokal dengan data seed yang realistis; catat p50/p95/p99, galat.
- DB: `EXPLAIN ANALYZE` untuk query lambat, deteksi N+1 lewat log query.
- Profil: profiler bahasa (cProfile, py-spy, Node `--prof`/Chrome DevTools), bukan tebakan.
- Hasil di lingkungan lokal bersifat relatif: bandingkan sebelum/sesudah pada mesin yang sama, jangan menjadikannya janji produksi.

## 3. Pola optimasi
Web: kurangi JS di jalur kritis, pecah bundle, gambar berukuran dan format tepat dengan dimensi tetap (cegah CLS), pemuatan font efisien, cache dan prefetch yang masuk akal, hindari render ulang berlebih. API/DB: indeks yang tepat, paginasi, query terpilih (bukan `SELECT *`), caching dengan kebijakan invalidasi jelas, batch dan antrean untuk kerja lama. Mobile/desktop: pekerjaan berat di luar thread UI, lazy loading, hindari alokasi per-frame.

## 4. Keandalan
Timeout pada semua panggilan keluar; retry dengan backoff dan jitter hanya untuk operasi idempoten; pembatas laju dan antrean untuk melindungi dependensi; circuit breaker sederhana bila bergantung pada layanan rapuh; penutupan anggun (graceful shutdown); health dan readiness benar; degradasi anggun saat dependensi gagal; pemulihan setelah crash tanpa efek ganda; batas ukuran masukan dan memori.

## 5. Platform non-web
Android: waktu mulai, ukuran APK, jank, penggunaan baterai (profil bila ada perangkat; selain itu `UNVERIFIED`). Windows: waktu mulai, memori saat idle, kebocoran handle/memori pada sesi panjang. Bot: konsumsi sumber daya per siklus, penanganan rate limit, stabil dalam run panjang (uji soak singkat).

## 6. Laporan (`company/reports/perf-*.md`)
Tabel: metrik, anggaran, sebelum, sesudah, perintah/alat, kondisi uji. Rekomendasi bertingkat (cepat, menengah). Temuan di atas anggaran pada alur kritis = bug S2 (S3 bila tidak kritis).

## 7. Anti-pola
Optimasi tanpa pengukuran; satu kali lari sebagai bukti; membandingkan mesin berbeda; cache tanpa invalidasi; retry pada operasi tidak idempoten; menyembunyikan lambat dengan spinner; mengejar skor Lighthouse tanpa memperbaiki pengalaman nyata.

## 8. Selesai bila
Anggaran ada dan terukur; angka sebelum/sesudah pada alur kritis ada; tidak ada temuan di atas anggaran yang tak ditangani; butir yang butuh alat/perangkat yang tidak ada ditandai `UNVERIFIED` dengan perintah untuk manusia.
