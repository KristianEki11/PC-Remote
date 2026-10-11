# Playbook: Dokumentasi

Dimuat oleh: Docs Writer, Release & Docs (mode datar), Release Manager, Success Auditor. Prinsip: dokumentasi adalah produk; perintah di dalamnya diuji, bukan ditebak.

## Daftar isi
1. Dokumen wajib · 2. Prinsip penulisan · 3. README · 4. Dokumen lain · 5. Menjaga sinkron · 6. Anti-pola · 7. Selesai bila

## 1. Dokumen wajib
| Dokumen | Pembaca | Isi |
|---|---|---|
| `README.md` | pengguna/pengembang baru | apa ini, prasyarat, **maksimal 5 perintah setup**, cara menjalankan, cara tes, struktur singkat, lisensi |
| `.env.example` | pengembang | semua variabel dengan nilai fiktif dan penjelasan |
| Runbook (`docs/RUNBOOK.md`) | operator | jalankan, hentikan, cadangkan, pulihkan, rotasi rahasia, troubleshooting, rollback |
| `CHANGELOG.md` | semua | perubahan per versi + known issues |
| Dokumen API | pengembang | dari OpenAPI (di-generate), contoh permintaan/respons, kode galat |
| Panduan pengguna | pengguna akhir | tugas utama langkah demi langkah (tangkapan layar opsional via Playwright) |
| Arsitektur ringkas | pengembang | diagram C4 sederhana (Mermaid), keputusan penting (tautan ADR), peta modul |
| Panduan serah terima (proyek klien) | pemilik non-teknis | cara mengedit konten, mengganti kredensial, siapa dihubungi, apa yang perlu diperbarui berkala |

## 2. Prinsip penulisan
Bahasa Indonesia untuk dokumen manusia (istilah teknis tetap Inggris bila lazim). Kalimat pendek, urutan tugas, perintah yang bisa disalin, **hasil yang diharapkan** setelah setiap perintah penting. Satu fakta ditulis di satu tempat (tautkan, jangan salin). Tidak menyebut hal yang belum ada.

## 3. README yang baik
1. Satu kalimat tentang apa ini dan untuk siapa. 2. Prasyarat dengan versi (Node/Python/Docker). 3. Setup (<= 5 perintah) dan cara menjalankan. 4. Cara menjalankan tes. 5. Konfigurasi penting (tautan `.env.example`). 6. Cara kontribusi/struktur singkat. 7. Lisensi. Windows dan Unix: sertakan varian bila perintahnya berbeda.

## 4. Dokumen lain
- Dokumen API dihasilkan dari kontrak agar tidak menyimpang; contoh dijalankan sungguhan terhadap instance lokal.
- Panduan pengguna fokus pada tugas ("Cara membuat invoice"), bukan daftar tombol.
- Diagram dengan Mermaid agar bisa dirawat sebagai teks.

## 5. Menjaga sinkron
Docs Writer memperbarui bagian dokumen setiap cluster selesai. **Docs-as-tests**: Success Auditor menjalankan README dan runbook dari clone bersih; setiap langkah yang gagal atau ambigu menjadi bug dokumentasi. Periksa tautan internal (tidak ada tautan mati) dan versi yang disebut.

## 6. Anti-pola
README yang menyembunyikan langkah; perintah yang belum pernah dijalankan; tangkapan layar usang; dokumentasi yang menyalin kode; dokumen "akan datang"; menulis klaim fitur yang belum ada.

## 7. Selesai bila
Semua dokumen wajib ada; README lulus uji clone bersih; `.env.example` lengkap; tidak ada tautan mati; known issues dan langkah manusia tersisa tercantum.
