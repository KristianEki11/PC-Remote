# Playbook: UX Design

Dimuat oleh: Head of UX, Head of Design, UX Researcher, UX Designer, UX Writer, UX Critic, Design Critic.
Batas: UX **tidak** memilih warna, font, radius, dan tidak membuat token (itu UI). UX tidak menambah fitur di luar PRD.

## Daftar isi
1. Masukan wajib · 2. Prosedur · 3. Konvensi navigasi per platform · 4. UX Critic · 5. Keluaran · 6. Checklist · 7. Anti-pola · 8. Skill terkait

## 1. Masukan wajib
`company/RULES_DIGEST.md`, PRD dan backlog, catatan riset `R-###`, ADR bahasa desain bila sudah ada.

## 2. Prosedur (P1-P2)
1. **Pengguna dan tugas.** Paling banyak 3 persona ringkas (siapa, tujuan, konteks, hambatan) dan 5-10 pernyataan tugas ("Ketika ..., saya ingin ..., agar ..."). Berbasis riset dan brief, bukan karangan; yang diasumsikan diberi label asumsi.
2. **Journey.** Tahap, tindakan, pikiran, titik nyeri, peluang untuk jalur utama + dua jalur gagal terpenting.
3. **IA.** Hirarki halaman/layar, model navigasi (bagian 3), penamaan label dengan bahasa pengguna, struktur URL/rute. Navigasi mengikuti tugas pengguna, bukan struktur organisasi.
4. **Alur.** Untuk tiap alur kritis (diagram Mermaid): langkah, titik keputusan, dan **semua state**: kosong/awal, memuat, sukses, galat (validasi, jaringan, izin, server), parsial, offline. Setiap galat punya jalur pemulihan.
5. **Wireframe low-fi** (ASCII, Mermaid, atau HTML tanpa gaya). Per layar: tujuan, elemen utama, hirarki, aksi utama dan sekunder, state. Pakai nama komponen design system bila sudah ada.
6. **Aturan interaksi.** Validasi (saat blur dan saat kirim), simpan otomatis atau eksplisit, konfirmasi aksi destruktif + opsi urungkan, indikator progres, pesan sukses, debounce pencarian, pintasan keyboard.
7. **Konten dan microcopy.** Nada suara; daftar teks tombol, label, placeholder, galat, kosong, sukses (ID, plus EN bila perlu). Rumus pesan galat: apa yang terjadi + penyebab + apa yang bisa dilakukan. Satu istilah satu makna; hindari jargon.
8. **Perilaku aksesibilitas.** Urutan tab dan fokus, pengelolaan fokus setelah dialog/navigasi, pengumuman perubahan dinamis, label formulir, alternatif untuk drag dan gestur.
9. **Kriteria penerimaan UX terukur.** Contoh: "tugas X selesai dalam paling banyak N langkah", "setiap galat punya jalur pulih", "layar Y punya 5 state", "navigasi utama paling banyak 7 item".

## 3. Konvensi navigasi per platform
| Platform | Navigasi utama | Kembali | Catatan |
|---|---|---|---|
| Web | bilah atas atau sidebar, breadcrumb bila dalam | tombol kembali browser harus bekerja; URL dapat dibagikan | mobile-first; pencarian di tempat yang dapat diprediksi |
| Android | navigation bar (3-5 tujuan), rail untuk layar lebar, top app bar | gestur/tombol kembali sistem selalu bermakna | konfirmasi keluar hanya bila ada data belum tersimpan |
| Windows | menu/NavigationView, pintasan keyboard, bilah judul | Alt+Kiri atau tombol kembali di app | semua aksi dapat dijangkau keyboard; dukung jendela diubah ukurannya |
| Ekstensi | popup sempit (satu tugas utama), halaman opsi untuk sisanya | - | jangan menumpuk fitur di popup |
| Game | menu, mode, permainan, jeda | jeda alih-alih keluar | onboarding bertahap, bukan teks panjang |

## 4. UX Critic (ulasan hasil build)
Metode: **cognitive walkthrough** tiap alur kritis (apakah pengguna tahu apa yang harus dilakukan, melihat kontrolnya, mengaitkannya dengan tujuan, mendapat umpan balik?) + **10 heuristik Nielsen** (status sistem terlihat, kecocokan dengan dunia nyata, kontrol dan kebebasan pengguna, konsistensi, pencegahan galat, mengenali bukan mengingat, fleksibilitas dan efisiensi, desain minimalis, bantu pulih dari galat, bantuan dan dokumentasi) + cek 5 state + uji jalur galat.
Bukti: urutan langkah, tangkapan layar (Playwright) atau snapshot DOM. Severity: **S1** pengguna tidak bisa menyelesaikan tugas inti atau ada jebakan kehilangan data; **S2** membingungkan sehingga sering gagal, atau galat tanpa jalur pulih; **S3** gesekan dan inkonsistensi.
Keluaran `company/design/ux/ux-review-###.md`: ringkasan, tabel temuan (ID, alur/layar, heuristik, bukti, severity, rekomendasi). Critic bukan perancang spesifikasi yang sama.

## 5. Keluaran
`company/design/ux/`: `research.md`, `personas.md`, `journeys.md`, `ia.md`, `flows.md`, `wireframes/`, `content.md`, `ux-review-###.md`. Dibekukan di G2 bersama kontrak lain.

## 6. Checklist mutu
- Setiap alur kritis punya 5 state dan jalur pulih untuk tiap galat.
- Setiap formulir punya aturan validasi, teks galat, dan perilaku saat gagal kirim.
- Aksi destruktif punya konfirmasi atau urungkan; tidak ada jalan buntu.
- Alur pertama kali (onboarding) terdefinisi; label konsisten dengan komponen design system.
- Tidak ada lorem ipsum atau teks sementara.

## 7. Anti-pola
Mendesain tampilan (itu UI); fitur di luar PRD; navigasi berdasar struktur organisasi; alur tanpa state galat atau kosong; dark pattern (opsi keluar disembunyikan, hitung mundur palsu, centang pra-pilih yang merugikan); persona karangan tanpa dasar.

## 8. Skill terkait
Full of Design (via `design_engine.py`): aturan UX dari basis data (nama domain di `--help`), dokumen HIG tentang navigasi/modalitas/input (`hig <kata kunci>`); persist ke `company/design/ux/`. Ringkas dengan kata sendiri, jangan menyalin dokumen.
