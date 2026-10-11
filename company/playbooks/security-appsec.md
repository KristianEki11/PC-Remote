# Playbook: Security (AppSec)

Dimuat oleh: CISO, Security Auditor; dibaca Builder Backend/Core, Critic, Auditor. Keamanan digeser ke kiri: threat model di P2 (sebelum kode), audit di P5, bukan baru di akhir.

## Daftar isi
1. Threat model (G2) · 2. OWASP Top 10:2025 · 3. ASVS · 4. Kontrol inti · 5. Platform khusus · 6. Alat · 7. Temuan dan severity · 8. Anti-pola · 9. Selesai bila

## 1. Threat model (`company/security/THREAT-MODEL.md`)
- **Aset** (data, kredensial, uang, reputasi), **aktor** (anonim, pengguna, admin, pihak ketiga, orang dalam), **batas kepercayaan** (browser-server, server-DB, server-layanan luar, aplikasi-OS).
- Diagram alur data (Mermaid) dengan batas kepercayaan.
- **STRIDE** per elemen/alur: Spoofing, Tampering, Repudiation, Information disclosure, Denial of service, Elevation of privilege. Tiap ancaman: kontrol (ada, akan ada, atau risiko diterima + alasan).
- **Kasus penyalahgunaan** dari alur kritis PRD (mis. "pengguna A mengganti ID untuk melihat data B").
- Keluaran: daftar kontrol wajib yang masuk kriteria penerimaan cluster dan tes.

## 2. OWASP Top 10:2025 (periksa bila edisi sudah berganti)
| ID | Kategori | Yang diperiksa |
|---|---|---|
| A01 | Broken Access Control (termasuk SSRF) | otorisasi deny-by-default, kepemilikan objek, isolasi tenant, tanpa IDOR; SSRF: daftar putih tujuan, blokir rentang internal dan metadata cloud |
| A02 | Security Misconfiguration | header keamanan, mode debug mati, tanpa kredensial default, CORS minimal, galat tidak membocorkan, panel/dokumen admin tidak terbuka |
| A03 | Software Supply Chain Failures | lockfile, audit dependensi, lisensi, sumber paket tepercaya, CI berizin minimal, versi dikunci |
| A04 | Cryptographic Failures | TLS, hash sandi modern, tanpa kripto buatan sendiri, kunci tidak di repo, data sensitif terenkripsi bila perlu |
| A05 | Injection | kueri terparameter/ORM, encoding keluaran, validasi masukan, tidak ada eval atau perintah shell dari masukan |
| A06 | Insecure Design | threat model, batas laju dan kuota, kasus penyalahgunaan dipertimbangkan |
| A07 | Authentication Failures | kebijakan sandi, batas percobaan, sesi aman, reset sandi aman, pesan galat netral |
| A08 | Software or Data Integrity Failures | verifikasi tanda tangan pembaruan dan webhook, deserialisasi aman, integritas artefak |
| A09 | Security Logging and Alerting Failures | log peristiwa keamanan (login gagal, perubahan hak) tanpa rahasia, alert dasar |
| A10 | Mishandling of Exceptional Conditions | gagal aman (fail closed), pengecualian ditangani konsisten, tanpa stack trace ke pengguna, tanpa state setengah jadi |

Setiap kategori mendapat status di laporan: **bersih (dengan bukti)** atau **temuan `S-###`**. "Tidak diperiksa" tidak boleh.

## 3. ASVS
Gunakan OWASP ASVS (versi terbaru, 5.x saat ini; verifikasi) sebagai daftar periksa terstruktur: **Level 1** default; **Level 2** untuk autentikasi, uang, atau data pribadi; Level 3 hanya untuk sistem sangat kritis.

## 4. Kontrol inti
- **Authn/z** dan sesi: lihat `backend-data.md` bagian 3.
- **CSRF/CORS/header**: token atau `SameSite` untuk aksi yang mengubah state; CORS dengan daftar asal eksplisit; header: HSTS (bila HTTPS), `X-Content-Type-Options: nosniff`, `frame-ancestors`/`X-Frame-Options`, `Referrer-Policy`, `Permissions-Policy`, CSP yang realistis.
- **Unggah berkas**: batas ukuran dan tipe berdasarkan isi, nama acak, simpan di luar webroot, `Content-Disposition` aman.
- **Rahasia**: hanya di env/secret manager; pemindai rahasia pada riwayat repo; rotasi bila bocor.
- **Batas laju** pada login, reset sandi, endpoint mahal. **Log** keamanan tanpa PII.

## 5. Platform khusus
- **Android**: `mobile-android.md` bagian 5 (MASVS). **Windows**: lokasi tulis, pembaruan bertanda tangan, urutan pencarian DLL, IPC yang diekspos, Credential Manager/DPAPI. **Ekstensi**: izin minimal, CSP, tanpa kode remote, validasi pengirim pesan, content script tidak mempercayai halaman. **Bot**: kredensial, batas aksi, kill switch, log bersih. **AI/LLM**: prompt injection, kebocoran data ke penyedia, keluaran tidak dipercaya (validasi skema), batas biaya; rujuk OWASP Top 10 for LLM Applications.

## 6. Alat (hanya terhadap kode dan instance lokal project)
SAST: semgrep, bandit (Python), eslint-plugin-security (JS). Rahasia: gitleaks atau trufflehog. Dependensi: `npm audit`, `pip-audit`, `composer audit`, osv-scanner, trivy (image dan dependensi). Alat yang butuh pemasangan baru: catat, minta persetujuan bila unduhan besar (R11). Alat tidak ada: tinjauan manual dan label `UNVERIFIED`.

## 7. Temuan dan severity (`S-###.md`, template 06)
- **S1**: mudah dieksploitasi dengan dampak data/akun/uang (bypass autentikasi, IDOR yang membuka data pengguna lain, injeksi, rahasia di repo).
- **S2**: butuh kondisi atau hak tertentu, atau dampak terbatas (CSRF pada aksi penting, header penting hilang di produksi, tanpa batas laju pada login).
- **S3**: hardening dan defense-in-depth.
Tiap temuan: bukti reproduksi lokal (tes gagal atau keluaran alat), dampak, rekomendasi, tautan ke `BUG-###`. Perbaikan lewat protokol bug dengan tes regresi.

## 8. Anti-pola
"Sudah aman" tanpa bukti; otorisasi hanya di front-end; rahasia di repo atau log; mematikan verifikasi TLS; kripto buatan sendiri; dependensi asing tanpa tinjauan; temuan tanpa cara reproduksi; menyimpan PoC eksploit lengkap alih-alih tes yang gagal.

## 9. Selesai bila
Threat model ada sebelum G2; tiap kategori A01-A10 punya status berbukti; audit dependensi dan pemindaian rahasia dijalankan (atau `UNVERIFIED`); semua S1/S2 tertutup dengan tes regresi; laporan `S-###` lengkap.
