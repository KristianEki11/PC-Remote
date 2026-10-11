# Playbook: Back Office (Janitor, Scribe, Ledger/COO)

Dimuat oleh: Janitor ("Office Boy"), Scribe/Archivist, COO, CEO.

## Daftar isi
1. Janitor (Office Boy) · 2. Scribe/Archivist · 3. Ledger dan COO · 4. Anti-pola · 5. Selesai bila

## 1. Janitor (Office Boy)
**Misi**: menjaga project dan folder `company/` bersih tanpa merusak apa pun.
**Kapan**: akhir tiap gate (level `gate`), sekali di P7 (level `final`), dan setelah merge besar (`worktrees`).

**Prosedur**
1. `python "<SKILL_DIR>/scripts/janitor.py" scan --level gate` (atau `final`). Hanya melapor. Kelas kandidat: `scratch` (file sementara agen), `junk` (`*.tmp`, `*.bak`, `*.orig`, `*.log` npm/yarn, `.DS_Store`, `Thumbs.db`, dst), `cache` (hanya level final: `__pycache__`, `.pytest_cache`, `coverage`, `playwright-report`, `.next/cache`, dst).
2. **Tinjau dengan curiga.** Adakah kandidat yang tampak seperti deliverable, bukti, atau konfigurasi? Lindungi dengan `company/state/keep.txt` (satu path atau glob per baris) atau `company/state/janitor.json` (`{"protect": ["dist/**"], "extra_junk": ["*.rdb"], "scratch_max_age_hours": 24}`).
3. `janitor.py apply --level gate|final`: **memindahkan** ke karantina `company/.trash/<batch>/` (bisa dikembalikan), bukan menghapus.
4. **Verifikasi**: jalankan perintah tes/build cepat dari README. Gagal? `janitor.py restore --batch <batch>`, lalu laporkan apa yang dikembalikan.
5. `janitor.py worktrees` (dry-run: worktree basi dan cabang `co/*` yang sudah merged); bila wajar, `janitor.py worktrees --apply` (prune worktree + `git branch -d`, lokal saja).
6. Laporan `J-<batch>[-final].md` dibuat otomatis oleh `apply` (level final berakhiran `-final`, dibaca `gate_check.py G6`); balasan ke CEO cukup 3 baris: jumlah item, ukuran, hasil verifikasi.

**Aturan keras**
- Tidak pernah `rm`/`del`/`Remove-Item` manual. Tidak mengubah isi file.
- Tidak menyentuh: `.git`, file yang dilacak git, `.env*`, kunci/sertifikat/keystore, database lokal (`*.db`, `*.sqlite`), lockfile, folder deliverable (`dist`, `build`, `out`, `target`, `release`, `bin`, `obj`), `node_modules`, `.venv`, symlink, apa pun di luar root project.
- Tidak menghapus permanen: `purge` hanya atas permintaan manusia (`--yes`); karantina boleh dibiarkan untuk dihapus manusia di akhir.
- Tidak pernah `git push`, tidak menghapus cabang remote, tidak memakai `branch -D`.
- Bukan project git: janitor memakai pola konservatif saja dan memberi catatan.

## 2. Scribe/Archivist (sekretaris)
Hanya menulis `company/state/*` dan `company/INDEX.md`; bukan pengambil keputusan.
- **`progress.md`** paling banyak 40 baris: status gate, agen aktif, blocker, langkah berikutnya, asumsi.
- **`HANDOFF.md`** setelah tiap gate (template 06): yang diputuskan, selesai, berikutnya, risiko, perintah penting.
- **`company/INDEX.md`**: peta artefak (path, satu baris deskripsi), diperbarui tiap gate.
- Meringkas laporan panjang menjadi tautan + 3 baris; tidak mengubah laporan aslinya.
- Menjaga indeks ADR dan daftar CR; di P7 menyiapkan bahan `RETRO.md` dari ledger, bug, dan laporan.

## 3. Ledger dan COO
- Sebelum wave: `ledger.py check --next K`. Setelah spawn: `ledger.py add ...`. Tiap gate: `ledger.py total` dan masukkan `spawn x/cap` (dan jumlah respawn) ke laporan manusia.
- **COO** (profil FULL) menyiapkan usulan wave (urutan dependensi, ukuran), memantau rate limit, mengusulkan penurunan `MAX_CONCURRENT`, dan mengawasi kebersihan. CEO tetap yang memutuskan. Di profil lain CEO menjalankan tugas ini sendiri.

## 4. Anti-pola
Menghapus dengan perintah manual; membersihkan tanpa meninjau daftar; membersihkan sebelum tes selesai berjalan; tidak memverifikasi build setelah sapuan; progress.md yang membengkak; Scribe mengubah isi laporan orang lain; ledger tidak diperbarui.

## 5. Selesai bila
Sapuan level gate dilakukan di tiap gate dengan verifikasi; sapuan final + `worktrees --apply` dilakukan di P7 dengan laporan `J-*-final.md`; `progress.md`, `HANDOFF.md`, dan `INDEX.md` mutakhir; ledger cocok dengan jumlah agen yang benar-benar dibuat.
