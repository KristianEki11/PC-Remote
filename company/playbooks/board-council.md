# Board (Dewan Penasihat)

Board mencegah keputusan mahal diambil oleh satu sudut pandang. Anggotanya persisten, tidak menulis kode, hanya berpendapat dan memutuskan.

## Daftar isi
1. Kursi
2. Kapan Board bersidang
3. Prosedur (3 ronde)
4. Board Lite (BLITZ)
5. Penasihat tak berhak suara
6. Tribunal bug
7. Format keputusan (ADR)

## 1. Kursi
| Kursi | Sudut pandang | Wajib menghasilkan |
|---|---|---|
| **ATLAS (Chair)** | netral; menyintesis, memastikan kriteria keputusan eksplisit | keputusan akhir + catatan perbedaan pendapat |
| **Contrarian** | mencari cacat fatal, mode kegagalan, biaya tersembunyi | risiko terbesar + mitigasi atau alternatif (kritik tanpa alternatif tidak dihitung) |
| **First-Principles** | menanggalkan asumsi: apa yang sebenarnya dibutuhkan? | versi paling sederhana yang tetap memenuhi tujuan |
| **Expansionist** | peluang, opsi masa depan, jalur skala | peluang + tanda bila ini scope creep |
| **Outsider** | pemula/klien/pengguna awam: istilah membingungkan? mudah dirawat orang lain? | hal yang tidak masuk akal bagi non-ahli |
| **Executor** | kelayakan, urutan, waktu, langkah konkret | rencana terkecil yang bisa dikirim dan risiko eksekusinya |

Jumlah kursi per profil (dari roster): BLITZ 3 (Chair, Contrarian, Executor), LEAN 4 (+Outsider), STANDARD 5 (+First-Principles), FULL 6 (+Expansionist). Kursi bisa terpangkas oleh cap; Chair dan Contrarian dipangkas paling akhir.

## 2. Kapan Board bersidang
- **Type-1** (00 bagian 3): stack, model data dan tenancy, model auth, bahasa desain dasar, hosting, arsitektur inti, strategi migrasi.
- **G2** (STANDARD, FULL): tinjauan kontrak satu ronde (Contrarian + Executor menantang kontrak sebelum dibekukan).
- **G6**: verdict akhir (GO, GO-dengan-known-issues, NO-GO).
- **Tribunal bug** (bagian 6).
- Tidak untuk keputusan Type-2 atau soal teknis kecil: itu urusan Head.

## 3. Prosedur (3 ronde, paling banyak 400 kata per posisi)
1. **Posisi independen.** CEO mengirim *satu* ringkasan masalah yang sama (konteks, opsi, kendala, kriteria) ke semua kursi. Tiap kursi menulis posisi tanpa melihat kursi lain. Hasil diberi label acak (A, B, C...) agar tidak bias pada siapa yang berbicara.
2. **Kritik silang.** Setiap kursi membaca semua posisi berlabel, menunjuk satu kelemahan terkuat dan satu hal yang ia ubah dari pikirannya sendiri.
3. **Sintesis.** Chair menulis keputusan: opsi terpilih, alasan, risiko yang diterima dan mitigasinya, pendapat berbeda yang dicatat, pemicu untuk meninjau ulang. Hasilnya menjadi ADR.

Kuorum: Chair + dua kursi lain. Chair tidak dapat memutuskan bertentangan dengan kesepakatan seluruh kursi lain tanpa alasan tertulis. Deadlock: CEO memutuskan dengan alasan tertulis di ADR; konsekuensi besar dan tidak bisa dibalik: E1.

## 4. Board Lite (BLITZ)
Chair, Contrarian, Executor. Satu ronde posisi + sintesis (tanpa kritik silang). Hanya dipanggil di G1 untuk keputusan Type-1 dan di G6.

## 5. Penasihat tak berhak suara
CISO, Head of UX, Head of UI, VP QA, atau VP lain yang relevan boleh dimintai masukan lewat `send_message` pada keputusan di domain mereka. Mereka bukan spawn baru dan tidak ikut memilih; pendapatnya dilampirkan ke ringkasan masalah.

## 6. Tribunal bug
Dipanggil bila perbaikan gagal 3 kali. Anggota: Chair, Contrarian, Executor (+ First-Principles bila ada). Input: laporan investigasi (gejala, langkah reproduksi, hipotesis yang sudah diuji beserta hasilnya). Keluaran satu dari:
1. hipotesis baru + penugasan ulang (dengan batas 2 percobaan lagi);
2. perubahan desain (CR) karena akar masalahnya di kontrak;
3. descope (hanya S3, dicatat sebagai known issue);
4. eskalasi E1 bila S1 tak terpecahkan.

## 7. Format keputusan (ADR)
Simpan di `company/adr/ADR-####.md` (template di `06-templates.md`): konteks, opsi, keputusan, alasan, konsekuensi, pendapat berbeda, pemicu tinjau ulang, tanggal.
