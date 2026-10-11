# Playbook: Bug Response

Dimuat oleh: Bug Chief, SWAT (investigator/fixer), VP QA (LEAN: triase), Tester. Protokol dan definisi severity ada di Konstitusi (00 bagian 4); ini adalah tekniknya.

## Daftar isi
1. Triase · 2. Reproduksi · 3. Akar masalah · 4. Perbaikan · 5. Regresi dan verifikasi · 6. Percobaan gagal dan Tribunal · 7. Pasca-insiden S1 · 8. Anti-pola · 9. Selesai bila

## 1. Triase (Bug Chief)
Baca `BUG-###.md`; tetapkan severity (S1/S2/S3), cluster/pemilik path (dari `ownership.md`), dan prioritas. Kelompokkan bug yang berakar sama. S1 menghentikan merge dan didahulukan. Satu SWAT per kelompok akar masalah; SWAT bekerja di cabang `co/<cluster>-swat-N`.
Bug yang menyentuh kontrak beku: buat CR, jangan menambal diam-diam.

## 2. Reproduksi (selalu pertama)
Tulis reproduksi **minimal dan otomatis**: tes yang gagal (lebih baik) atau skrip satu perintah. Tanpa reproduksi, jangan mengubah kode: kumpulkan log, versi, data, langkah; bila tetap tidak reproduksibel, laporkan `BLOCKED` dengan apa yang sudah dicoba. Reproduksi disimpan sebagai tes regresi.

## 3. Akar masalah
- Persempit: `git bisect` (otomatis dengan `git bisect run <skrip>` bila tes cepat), log berlapis, bagi dua jalur data.
- Bedakan **gejala** dan **penyebab**: tanyakan "mengapa" sampai ke kondisi yang dapat diperbaiki (5 whys secukupnya).
- Hipotesis tertulis satu per satu, tiap hipotesis diuji dengan eksperimen yang bisa menyangkalnya. Catat yang disangkal di bug.
- Periksa apakah akar masalahnya pola yang berulang di tempat lain (cari dengan grep pola serupa).

## 4. Perbaikan
- Perubahan sekecil mungkin yang menghilangkan akar masalah; bukan refaktor besar. Di luar path milik: CR.
- Tidak melemahkan tes yang ada. Tidak menambal gejala (mis. menelan pengecualian) tanpa menangani akar.
- Pertimbangkan efek samping: siapa lagi memakai kode ini; migrasi data yang terdampak; kompatibilitas kontrak.
- Bila perbaikan berisiko, sembunyikan di balik feature flag atau siapkan rollback.

## 5. Regresi dan verifikasi
Tes dari reproduksi harus **gagal sebelum** dan **lolos sesudah** (buktikan keduanya). Jalankan seluruh suite terkait. Status `FIXED` oleh SWAT; `VERIFIED` hanya oleh penemu atau Tester independen setelah menjalankan reproduksi dari cabang perbaikan. Cari bug serupa di modul tetangga dan catat yang ditemukan sebagai bug baru.

## 6. Percobaan gagal dan Tribunal
Hitung percobaan per bug. Gagal 3 kali: serahkan ke Tribunal (04 bagian 6) dengan laporan investigasi: gejala, reproduksi, hipotesis yang diuji dan hasilnya, apa yang dicurigai berikutnya. Jangan terus menebak.

## 7. Pasca-insiden S1
Tulis ringkasan 5 baris di bug: apa yang terjadi, dampak, akar masalah, perbaikan, apa yang mencegah terulang (tes, validasi, kontrak). Masukkan pelajaran ke `RETRO.md`.

## 8. Anti-pola
Memperbaiki tanpa reproduksi; menghapus atau melonggarkan tes agar hijau; menambal gejala; refaktor besar saat memperbaiki bug; menutup bug tanpa tes regresi; memverifikasi perbaikan sendiri; mengubah file di luar path milik.

## 9. Selesai bila
Tes regresi ada (gagal sebelum, lolos sesudah), suite terkait hijau, bug berstatus `VERIFIED` oleh pihak independen, laporan memuat perintah dan keluaran, dan tidak ada bug terkait yang terlewat.
