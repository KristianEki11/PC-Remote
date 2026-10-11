# Template

Salinan lokal ada di `company/playbooks/templates.md`. Baris berformat tetap (`Severity:`, `Status:`, `STATUS:`, `Verdict:`, tabel ownership) dibaca oleh `gate_check.py`; jangan mengubah bentuknya.

## Daftar isi
1. Work order
2. Laporan agen
3. ADR dan ADR-lite
4. Change Request (CR)
5. Bug (`BUG-###`)
6. Temuan keamanan (`S-###`)
7. PRD dan backlog
8. SCOPE audit dan INVENTORY fix
9. HANDOFF
10. Laporan rilis dan laporan audit
11. Retro
12. Baris ownership dan ledger

## 1. Work order
```markdown
# WORK ORDER <id>
Dari: <CEO|Head> | Untuk: <agent-id> (<peran>) | Tier: pro|flash | Workspace: inherit|branch | offload: false
Batas: maksimal 3 iterasi | balasan paling banyak 200 kata | laporan: company/reports/<agent-id>.md

## Konteks (yang sudah diputuskan)
- ADR/PRD/kontrak terkait (path), keputusan yang tidak boleh dibuka lagi

## Tujuan
<satu paragraf: hasil akhir yang diinginkan, bukan langkah-langkahnya>

## Baca dulu (berurutan)
1. company/RULES_DIGEST.md
2. company/playbooks/<playbook>.md
3. <PRD/kontrak/laporan lain yang relevan>

## Skill dan alat (bila ada)
- <path SKILL.md> / python "<SKILL_DIR>/scripts/design_engine.py" run -- "<query>" ...   (hasil di-persist ke <path>)

## Keluaran (path persis)
- <path 1>
- <path 2>

## Path milik (boleh ditulis) dan terlarang
- Milik: <glob>
- Terlarang: selain milik di atas, kecuali company/scratch/<agent-id>/ dan laporanmu

## Perintah verifikasi (jalankan dan lampirkan hasilnya)
- <perintah> -> yang diharapkan: <hasil>

## Kriteria selesai
- <butir terukur 1>
- <butir terukur 2>

## Jika terblokir
Lapor STATUS: BLOCKED dengan diagnosis (apa dicoba, hipotesis, butuh apa). Jangan menebak dan jangan keluar dari batas path.
```

## 2. Laporan agen (`company/reports/<agent-id>.md`, maksimal 200 kata)
```markdown
STATUS: DONE | PARTIAL | BLOCKED
Ringkasan: <2-3 kalimat>
Berkas dibuat/diubah: <daftar path>
Bukti: <perintah> -> <hasil ringkas>   (ulangi per butir kriteria selesai)
Masalah/risiko: <atau "tidak ada">
Butuh dari Head: <atau "tidak ada">
```

## 3. ADR dan ADR-lite
```markdown
# ADR-0001: <judul keputusan>
Status: PROPOSED | ACCEPTED | SUPERSEDED by ADR-xxxx
Tanggal: <YYYY-MM-DD> | Diputuskan oleh: Board|<Head>|CEO
Konteks: <masalah dan kendala>
Opsi: A) ... B) ... C) ...
Keputusan: <opsi terpilih>
Alasan: <3-5 poin, kaitkan dengan kriteria>
Konsekuensi: (+) ... (-) ...
Pendapat berbeda: <kursi/pihak dan alasannya>
Tinjau ulang bila: <pemicu terukur>
```
ADR-lite (keputusan Type-2) = Konteks, Keputusan, Alasan dalam 3-5 baris; simpan sebagai `company/adr/ADR-lite-###.md`.

## 4. Change Request
```markdown
# CR-001: <judul>
Pemohon: <agent-id> | Tanggal: <...> | Kontrak terdampak: <path>
Perubahan yang diminta: <rinci>
Alasan dan bukti: <kenapa tidak bisa dengan kontrak sekarang>
Dampak ke cluster/agen: <daftar>
Keputusan: DISETUJUI | DITOLAK | DITUNDA (oleh <Chief Architect|Head>, tanggal)
Setelah disetujui: ubah kontrak, jalankan gate_check.py freeze ulang, beri tahu semua agen terdampak.
```

## 5. Bug
```markdown
# BUG-001: <judul singkat>
Severity: S1 | S2 | S3
Status: OPEN | IN-PROGRESS | FIXED | VERIFIED | DESCOPED
Cluster: C1
Ditemukan oleh: <agent-id> pada <fase/gate>
Langkah reproduksi: 1) ... 2) ... 3) ...
Aktual vs diharapkan: ...
Bukti: <perintah/log/path screenshot>
Akar masalah (SWAT): ...
Perbaikan: <cabang/commit>
Tes regresi: <path tes>
Verifikasi (penemu atau Tester independen): <perintah -> hasil>
```
Status selain `FIXED`, `VERIFIED`, `DESCOPED` dianggap terbuka oleh `gate_check.py`.

## 6. Temuan keamanan
```markdown
# S-001: <judul>
Severity: S1 | S2 | S3 | Kind: appsec | pentest | dependency | secrets | privacy
Kategori: OWASP A0x:2025 / ASVS V... / CWE-...
Lokasi: <path atau endpoint>
Bukti: <tes gagal atau keluaran pemindai + perintah reproduksi terhadap instance LOKAL>
Dampak: <apa yang bisa terjadi, siapa terdampak>
Rekomendasi: <perbaikan konkret>
Pelacakan: BUG-### | Status: OPEN | FIXED | VERIFIED
```

## 7. PRD dan backlog
```markdown
# PRD - <nama>
1. Masalah dan pengguna  2. Tujuan dan non-tujuan  3. Lingkup MVP (dan yang ditunda ke v2)
4. User story + kriteria penerimaan (Given/When/Then)  5. Alur kritis (daftar untuk e2e)
6. NFR (kinerja, keamanan, aksesibilitas, i18n, kompatibilitas)  7. Batasan dan asumsi
8. Metrik keberhasilan  9. Risiko  10. Pertanyaan terbuka (dengan default)
```
```markdown
# Backlog
| Cluster | Tujuan | User story | Kontrak disentuh | Dependensi | Ukuran | Path milik |
|---|---|---|---|---|---|---|
## v2 (di luar PRD)
```

## 8. SCOPE audit dan INVENTORY fix
```markdown
# SCOPE audit (company/audit/SCOPE.md)
Sistem dan versi: ... | Lingkungan uji (lokal/Docker, data fiktif): ... | Di luar lingkup: ...
Area audit: QA, keamanan, pentest sandbox, a11y, performa, UX/UI, privasi | ROE pentest disetujui: ya/tidak
Cara menjalankan sistem: <perintah> | Kontak manusia untuk keputusan: <...>
```
`company/bugs/INVENTORY.md` (misi FIX): tabel bug yang dilaporkan manusia + hasil reproduksi + severity awal.

## 9. HANDOFF (`company/state/HANDOFF.md`, ditulis Scribe setelah tiap gate)
```markdown
# HANDOFF setelah G<n> - <tanggal>
Status: <ringkas> | Spawn terpakai: x/cap
Sudah diputuskan: ADR-... (jangan dibuka lagi tanpa pemicu)
Selesai: <daftar> | Berikutnya: <daftar urut> | Risiko dan asumsi: <daftar>
Perintah penting: <build, tes, run> | Agen persisten aktif: <Head>
```

## 10. Laporan rilis dan laporan audit
```markdown
# RELEASE REPORT - <nama> <versi>
Verdict: GO | GO-dengan-known-issues | NO-GO
Ringkasan: ... | Kriteria selesai: <tabel butir -> bukti -> status>
Known issues (S3 dan S2 yang diterima dengan alasan): ...
Langkah manusia yang tersisa: <deploy, domain, tanda tangan kode, publikasi store, secret production>
Cara menjalankan: <perintah> | Artefak: <path + checksum SHA-256>
```
```markdown
# AUDIT REPORT - <sistem>
Verdict: GO | GO-dengan-known-issues | NO-GO
Ringkasan eksekutif (5 baris) | Lingkup dan metode | Temuan: tabel (ID, severity, area, bukti, rekomendasi, usaha)
Peta jalan remediasi: cepat (hari), menengah (minggu), panjang | Hal yang sudah baik | Lampiran bukti: company/audit/evidence/
```
Aturan verdict AUDIT: NO-GO bila ada S1; GO-dengan-known-issues bila hanya S2/S3 dengan rencana; GO bila bersih.

## 11. Retro (`company/RETRO.md`, P7)
Yang berhasil | Yang gagal atau lambat (dengan penyebab) | Agen/peran yang sia-sia atau kurang | Angka: spawn terpakai, respawn, bug per severity | Saran perbaikan skill ini.

## 12. Baris ownership dan ledger
Ownership (tabel di `company/ownership.md`): `| src/auth/** | builder-be-c1 | C1 |`
Ledger (ditulis `ledger.py`): `- <waktu> | spawn|respawn | <agent-id> | <peran> | <tier> | <workspace> | <catatan>`
