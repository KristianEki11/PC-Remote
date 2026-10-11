# Playbook: Privasi dan Kepatuhan

Dimuat oleh: Privacy & Compliance Reviewer, CISO, Security Auditor, Release Manager.
Penting: ini panduan kerja teknis, **bukan nasihat hukum**. Hal yang berimplikasi hukum ditandai "perlu ditinjau manusia/penasihat hukum".

## Daftar isi
1. Inventaris data · 2. Minimisasi dan klasifikasi · 3. Transparansi dan hak pengguna · 4. Regulasi · 5. Log dan pihak ketiga · 6. Lisensi · 7. Kebijakan store · 8. Pembayaran dan data anak · 9. Keluaran · 10. Anti-pola · 11. Selesai bila

## 1. Inventaris data (`company/security/PRIVACY-INVENTORY.md`)
Tabel: data apa, untuk apa (tujuan), dari siapa, disimpan di mana, berapa lama (retensi), siapa yang bisa mengakses, dikirim ke pihak ketiga mana. Tidak ada data yang dikumpulkan tanpa tujuan di PRD.

## 2. Minimisasi dan klasifikasi
Kumpulkan seminimal mungkin. Klasifikasi: publik, internal, pribadi, sensitif (kesehatan, biometrik, keuangan, data anak, lokasi presisi). Data sensitif butuh alasan kuat, enkripsi, akses terbatas, dan retensi pendek. Seed dan tes selalu data fiktif.

## 3. Transparansi dan hak pengguna
- Kebijakan privasi dan ketentuan layanan: **draf** yang mencerminkan inventaris data apa adanya, ditandai untuk ditinjau manusia/hukum; jangan menyalin teks dari situs lain.
- Cookie dan analitik: pilih analitik ramah privasi bila memungkinkan; spanduk persetujuan hanya bila ada pelacakan yang memerlukannya.
- Bila ada akun: fitur ekspor data dan hapus akun (termasuk data turunan dan cadangan sesuai retensi), perubahan email/sandi aman.

## 4. Regulasi (verifikasi dari sumber resmi, bertanggal: R15)
Indonesia: UU Pelindungan Data Pribadi No. 27 Tahun 2022 beserta peraturan pelaksanaannya; lihat ketentuan terkini. Bila pengguna di Uni Eropa: GDPR. Sektor tertentu (keuangan, kesehatan) punya aturan sendiri. Periksa persyaratan konkret (dasar pemrosesan, hak subjek data, notifikasi insiden, transfer lintas negara) lalu terjemahkan menjadi tugas teknis; sisanya untuk ditinjau manusia.

## 5. Log dan pihak ketiga
Log tanpa PII dan rahasia (redaksi nilai sensitif). Daftar semua pihak ketiga dan SDK (analitik, penyedia email, penyedia AI, CDN, font): data yang dikirim, lokasi, dasar. **Penyedia AI/LLM: jangan mengirim data pribadi tanpa dasar dan persetujuan**; redaksi sebelum dikirim.

## 6. Lisensi (`company/security/LICENSES.md`)
Daftar dependensi dan lisensinya (SBOM sederhana dari manajer paket). Daftar putih awal: MIT, Apache-2.0, BSD, ISC. Tandai GPL/AGPL/LGPL, lisensi "source-available" atau non-komersial untuk ditinjau bila produk didistribusikan tertutup/komersial. Font, ikon, gambar, audio, aset game: sumber dan lisensinya dicatat; sertakan atribusi bila disyaratkan. Konflik lisensi tak terselesaikan: E4.

## 7. Kebijakan store
Google Play: formulir Data Safety dan kebijakan privasi harus sesuai perilaku aplikasi; izin dijelaskan. Chrome Web Store: tujuan tunggal ekstensi, justifikasi setiap izin, kebijakan data pengguna. Windows: penandatanganan kode dan peringatan SmartScreen (langkah manusia). Kebijakan berubah: Explorer memverifikasi dokumen resmi terbaru dengan tanggal.

## 8. Pembayaran dan data anak
Jangan menyimpan atau memproses data kartu sendiri; gunakan bidang yang di-host gateway agar cakupan PCI minimal. Produk yang menyasar anak atau mengumpulkan data anak: tandai untuk tinjauan hukum sebelum rilis.

## 9. Keluaran
`PRIVACY-INVENTORY.md`, `LICENSES.md`, draf kebijakan privasi dan ketentuan layanan (bila relevan), daftar "butuh peninjauan manusia/hukum" di `RELEASE_REPORT.md`.

## 10. Anti-pola
Mengumpulkan data "kalau-kalau butuh"; log yang memuat PII; mengirim data pribadi ke penyedia AI tanpa dasar; kebijakan privasi yang tidak sesuai perilaku; mengabaikan lisensi aset; menyatakan "patuh hukum" (tidak ada agen yang boleh menyatakannya).

## 11. Selesai bila
Inventaris data dan daftar lisensi lengkap; tidak ada PII di log; pihak ketiga terdaftar; butir hukum ditandai untuk manusia; draf dokumen kebijakan (bila relevan) tersedia.
