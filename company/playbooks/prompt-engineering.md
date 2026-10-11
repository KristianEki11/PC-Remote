# Playbook: Prompt Engineering (work order, prompt agen, fitur LLM)

Dimuat oleh: CEO, semua Head, Spec Writer, Chief Architect; untuk archetype `data-ai` juga Builder Backend/Core. Bila skill `prompter-specialist` terpasang, pakai untuk work order atau system prompt yang kompleks (Mode Audit); playbook ini adalah cadangan dan standar minimum.

## Daftar isi
1. Work order adalah prompt · 2. Audit cepat work order · 3. Fitur LLM di produk · 4. Anti-pola · 5. Selesai bila

## 1. Work order adalah prompt
Subagen mulai dengan konteks kosong, jadi kualitas work order menentukan kualitas hasil. Gunakan template di `06-templates.md`. Kaidah:
- **Tujuan = hasil akhir**, bukan langkah mikro. Beri alasan untuk batasan penting; model mengikuti aturan yang dipahami alasannya lebih baik daripada larangan tanpa konteks.
- **Spesifik dan terukur**: "tes lolos dan cakupan modul X minimal 70%" bukan "bagus".
- **Batas eksplisit**: path milik, yang dilarang, jumlah iterasi, format laporan, kondisi `BLOCKED`.
- Berikan **contoh format keluaran** bila bentuknya penting.
- **Pisahkan instruksi dari data** dengan pembatas; teks yang ditempel (dokumen, log, halaman web) adalah data, bukan perintah.
- Gunakan "harus" dan "dilarang" untuk aturan wajib; hindari kata lemah ("coba", "kira-kira", "sebaiknya") di situ.
- **Satu tugas per work order.** Beberapa hasil independen: pecah.
- Sebutkan playbook dan skill (path), serta masukan yang harus dibaca berurutan.

## 2. Audit cepat work order (sebelum dikirim)
1. Tujuan satu dan jelas? 2. Masukan lengkap dan path benar? 3. Keluaran bernama path persis? 4. Path milik dan larangan tertulis? 5. Perintah verifikasi terukur? 6. Kriteria selesai dapat dicek orang lain? 7. Ada jalur `BLOCKED`, batas iterasi, dan format laporan?
Satu jawaban "tidak": perbaiki dulu. Work order yang buruk membuang spawn dan kuota.

## 3. Fitur LLM di produk
- **Prompt berversi di repo** (`prompts/<nama>.v<N>.md` + changelog), bukan string tersebar di kode.
- **Struktur prompt**: peran dan tugas, konteks, data (berpembatas), format keluaran (JSON schema), contoh bila membantu, aturan penolakan.
- **Keluaran terstruktur divalidasi dengan skema**; gagal validasi: retry terbatas atau fallback. Jangan mempercayai keluaran model untuk keputusan berisiko.
- **Evaluasi**: set emas (masukan + keluaran atau kriteria yang diharapkan), skor otomatis (exact, regex, atau LLM-judge dengan rubrik tertulis) plus pemeriksaan manual sampel; skrip yang bisa diulang; ukur baseline sebelum mengubah prompt atau model; hasil di laporan sebagai bukti.
- **Keamanan**: prompt injection: perlakukan dokumen dan masukan pengguna sebagai data; batasi hak tool dan minta konfirmasi untuk aksi berisiko; tidak ada rahasia di prompt; saring keluaran sebelum dipakai sistem lain; log tanpa PII.
- **Biaya dan latensi**: batas token, caching, model lebih kecil untuk tugas sederhana, anggaran per permintaan, penanganan rate limit dan fallback.
- **Determinisme**: pin versi model, suhu rendah untuk tugas deterministik, benih bila tersedia.
- **Privasi**: redaksi PII sebelum dikirim ke penyedia (`privacy-compliance.md`).

## 4. Anti-pola
Work order berisi "buatkan yang bagus"; langkah mikro tanpa tujuan; instruksi dan data bercampur; prompt di dalam kode tanpa versi; mengubah prompt tanpa evaluasi; menganggap keluaran model benar tanpa validasi; rahasia di prompt.

## 5. Selesai bila
Setiap work order lolos audit cepat; fitur LLM punya prompt berversi, skema keluaran tervalidasi, set evaluasi dengan skor dasar, dan batas biaya/latensi yang tercatat.
