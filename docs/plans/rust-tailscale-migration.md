# Rencana migrasi server Rust + Tailscale

Tanggal: 2026-11-13. Status: rencana awal, belum implementasi.
Baseline: origin/main 99fc6d0449e4ae2c07702e4944552e67db9d618c.

## Tujuan dan batas scope

- Server Windows ditulis ulang bertahap dalam Rust; aplikasi Android tetap Flutter.
- Turunkan penggunaan CPU idle, RAM total dan proses sementara; buktikan dengan benchmark pada mesin yang sama. Rust bukan jaminan penghematan tanpa pengukuran.
- LAN langsung tetap bisa digunakan tanpa Tailscale atau internet. Akses internet menggunakan jaringan privat Tailscale.
- Mendukung pengguna independen, bukan hanya laptop/account pengembang.
- Tidak membangun relay cloud sendiri pada fase ini. Quick Tunnel dipertahankan hanya selama transisi dan dihentikan setelah jalur baru lolos pengujian.
- Branch ini mulai dari kode yang committed. Perubahan lokal auth/TLS/rate-limit dan laporan audit yang belum committed tidak otomatis masuk baseline; review dan port perbaikannya secara eksplisit sebelum rilis.

## Keputusan: jaringan milik masing-masing pengguna

Setiap pengguna memasang Tailscale resmi di PC dan HP, lalu menghubungkan keduanya ke tailnet miliknya sendiri. Tidak ada akun pusat pengembang, shared auth key dalam installer, atau kewajiban masuk tailnet Kristian. Tidak memerlukan VPS, domain publik, exit node atau subnet router untuk mengontrol PC sendiri.

Untuk pengguna yang sudah memiliki Tailscale (termasuk akses SSH kampus/VPS), gunakan profil aktif yang ada jika kedua perangkat dapat bergabung dan kebijakan mengizinkan. Jangan menjalankan reset, logout, force-reauth, mengganti exit node, advertised routes, DNS, atau aturan SSH secara diam-diam. Tailnet organisasi dapat memerlukan izin admin; app harus memberi diagnosis, bukan mengubah policy organisasi.

Anggota tailnet tidak otomatis berhak mengontrol PC. Pairing aplikasi tetap wajib, dengan identitas dan kredensial terpisah per perangkat, daftar perangkat terpercaya, pencabutan individual, dan pembatasan brute force. Berbagi PC lintas akun/tailnet bukan syarat MVP; dokumentasikan sebagai fitur lanjutan dengan izin eksplisit pemilik.

Tailscale Personal gratis hanya sesuai ketentuan penggunaan nonkomersial. Jangan menjanjikan semua penggunaan bisnis gratis. Setiap pengguna mengelola akun/plan sendiri. Android hanya mendukung satu VPN aktif pada satu waktu: tampilkan konflik ini dalam setup dan sediakan LAN sebagai fallback, bukan menonaktifkan VPN lain otomatis.

## Onboarding dan lifecycle

1. Installer memasang server/agent PC Remote. Deteksi Tailscale yang sudah ada; jika belum ada, pandu instalasi resmi dan login oleh pengguna.
2. Pandu Run Unattended untuk Windows agar koneksi tersedia setelah restart/logout. Konfigurasi administratif hanya lewat tindakan pengguna yang jelas.
3. Jelaskan node key expiry terpisah dari SSH check mode. Pemilik boleh memilih menonaktifkan expiry untuk perangkat tepercaya; jika tetap aktif, tampilkan status perlu autentikasi ulang. Jangan menjanjikan tidak pernah login ulang.
4. HP bergabung ke tailnet yang sama; wizard memeriksa konektivitas port aplikasi dan membedakan offline, blocked policy, Tailscale tidak aktif, serta pairing dicabut bila bukti tersedia. Kegagalan reachability saja tidak cukup untuk menyimpulkan sebab spesifik.
5. QR pairing berisi server_id, versi protokol, endpoint LAN/Tailscale, fingerprint atau public key identitas, serta token sekali pakai dengan TTL. Setelah pairing, simpan rahasia HP di penyimpanan aman platform dan PC dengan proteksi Windows.
6. Startup/reconnect mengambil snapshot state terbaru. HP tidak perlu terus hidup di background; saat resume, sambung kembali dan sinkronisasi.

## Arsitektur server Rust

Usulan awal untuk divalidasi spike: Tokio + HTTP/WebSocket server (misalnya Axum), serde untuk kontrak, bindings Windows resmi untuk COM/WinRT. Versi dipilih/dikunci saat spike; ukur footprint konfigurasi minimal.

- `core`: model domain, command dispatcher, pairing/session, identitas perangkat dan penyimpanan.
- `transport`: HTTPS/WSS, autentikasi, pembatasan request, snapshot/events, reconnect/resume protocol.
- `windows-agent`: berjalan pada sesi pengguna; akses audio/media/display langsung melalui API Windows dan event subscription bila tersedia.
- `service`: fungsi yang benar-benar perlu hidup sebelum login dan supervisi agent. Tentukan kebutuhan hak akses minimum; jangan menjalankan seluruh UI/session agent sebagai SYSTEM.
- `ipc`: named pipe dengan ACL terbatas dan validasi pesan antara service dan agent; client tidak boleh mengirim arbitrary shell command.
- `tray/installer`: status, pairing, perangkat terpercaya, diagnostik, update dan uninstall.

Spike harus membuktikan batas Windows Session 0. Saat tidak ada user session, endpoint tetap memberi status jelas tetapi fitur desktop yang membutuhkan sesi tidak dijanjikan tersedia. Hindari PowerShell/schtasks berulang pada jalur polling media. Fallback yang masih diperlukan harus terukur dan terdokumentasi.

## Koneksi lokal dan internet

```text
Flutter -- HTTPS/WSS LAN langsung ----------------> PC Rust
Flutter -- Tailscale (direct / relay otomatis) ---> PC Rust
```

Transport aplikasi sama pada kedua jalur. Tailscale dijalankan sebagai komponen eksternal resmi, bukan di-embed atau ditulis ulang dalam Rust pada MVP. Ukur overhead Tailscale bersama server, agent, dan proses anak. Tidak memakai Tailscale SSH untuk command aplikasi dan tidak memakai Funnel untuk mempublikasikan endpoint.

Identitas server stabil dipisahkan dari IP/hostname. Simpan kandidat endpoint hasil pairing; preferensi LAN ditentukan dengan probe terbatas dan verifikasi identitas, fallback Tailscale bila tidak reachable. MagicDNS/IP Tailscale tidak dianggap identitas permanen: remove/re-enroll perangkat dapat mengubahnya. Pemulihan kasus tersebut dapat memerlukan pairing ulang, jangan dijanjikan discovery ajaib.

TLS dengan pin identitas hasil QR berlaku konsisten untuk kedua jalur. Definisikan rotasi sertifikat/identitas, dan jangan menerima sertifikat sembarang atau memilih trust policy berdasarkan substring domain. Bind interface dan aturan firewall hanya sesuai kebutuhan LAN/Tailscale; jangan membuka port router publik.

## Protokol dua arah v1

- Handshake: versi protokol, server_id, device_id, session, last_seen_revision.
- Server mengirim snapshot, kemudian event perubahan audio/media, status agent dan capabilities. Polling fallback dibatasi jika API Windows tertentu tidak menyediakan event.
- Command: request_id, nama command allowlist, parameter tervalidasi, deadline. Respons: request_id, hasil/error terstruktur, revisi state.
- Deduplikasi command pada server dengan batas ukuran/TTL dan scope perangkat. Setelah server restart dengan hasil command tidak diketahui, client tidak otomatis mengulang command non-idempotent.
- Reconnect memakai exponential backoff dengan jitter; heartbeat secukupnya dan satu koneksi aktif per sesi. Abaikan respons/revisi usang.
- Jangan antre shutdown/restart/aksi interaktif untuk dieksekusi setelah perangkat kembali online. Perintah yang hasilnya tidak diketahui ditampilkan demikian, bukan sukses palsu.
- Perubahan endpoint/capabilities tersinkronisasi ketika tersambung; bukan syarat utama pemulihan internet karena jalur Tailscale tidak bergantung pada Quick Tunnel.

## Tahapan implementasi dan gerbang kelulusan

### P0 — baseline dan kontrak
- Inventaris endpoint/fitur Go, termasuk audio master/device/channel, media, browser, power, brightness/overlay, pairing, tray dan installer.
- Review perubahan lokal keamanan yang belum committed, lalu tentukan commit sumber perbaikannya.
- Ukur idle 15 menit, sesi aktif 15 menit dan reconnect: private bytes/working set seluruh proses, CPU, jumlah proses PowerShell, latensi command p50/p95, ukuran instalasi. Pisahkan Tailscale yang sudah aktif dari tambahan beban PC Remote.
- Catat batas target numerik setelah baseline tersedia, sebelum hasil Rust dinilai.

### P1 — spike Rust + Flutter + Tailscale
- Server Rust minimal di port terpisah, health/capabilities, pairing dan satu command aman (misalnya pembacaan status volume).
- Flutter mengakses melalui LAN serta Tailscale; TLS pin dan auth yang sama.
- Jalankan pada mesin pengembang DAN pengguna kedua dengan tailnet independen.
- Lulus bila reboot, perpindahan Wi-Fi/seluler, HP resume dan Tailscale reconnect dapat pulih tanpa scan QR ulang ketika identitas/kredensial masih valid.

### P2 — state/events dan Windows agent
- Implementasikan snapshot, subscription, request ID, timeout/dedup serta IPC sesi pengguna.
- Buktikan status media/audio tidak meluncurkan proses shell berkala; uji logout/login Windows dan pergantian sesi.
- Tambahkan test unit/protokol untuk unauthorized, replay, respons usang, timeout setelah command diterima, reconnect dan pencabutan perangkat.

### P3 — paritas fitur dan migrasi Flutter
- Port fitur berdasarkan inventaris dengan contract tests. Uji operasi power hanya pada mesin uji dengan tindakan eksplisit.
- Pisahkan transport/session Flutter dari UI; perbaiki rollback volume dan overlapping polling selama migrasi.
- Versi protokol/capabilities mengatur kompatibilitas. Existing credential Go tidak otomatis diimpor tanpa review keamanan; wizard pairing ulang menjadi fallback yang jelas.

### P4 — installer, uji pengguna lain, rilis
- Uji clean install tanpa Tailscale, existing Tailscale kampus, key expired, VPN konflik, firewall/ACL blocked, offline, uninstall/update dan user tanpa admin.
- Dua user pada dua tailnet: masing-masing hanya dapat mengendalikan PC yang dipasangkan; peer satu tailnet tanpa pairing tetap ditolak.
- Bandingkan benchmark Go/Rust pada workload sama. Tunda klaim lebih ringan jika data belum mendukung.
- Installer mengganti service Go secara atomik setelah healthcheck; backup konfigurasi secara aman, dokumentasikan rollback, jangan menjalankan dua server produksi pada port yang sama.
- Matikan Quick Tunnel default dan keluarkan cloudflared dari instalasi baru setelah semua acceptance tests lulus. Pengguna lama mendapat jalur transisi eksplisit.

## CI dan bukti

Rust: cargo fmt --check, clippy, unit/contract tests dan build Windows; integrasi platform di Windows runner/mesin uji. Flutter: analyze, unit/widget/protocol tests dan build Android. Pisahkan tes yang memerlukan desktop interaktif dari CI headless. Catat hasil benchmark dan matrix acceptance sebagai artefak review.

## Referensi (diperiksa 2026-11-13)

- https://tailscale.com/docs/how-to/run-unattended
- https://tailscale.com/docs/features/access-control/key-expiry
- https://tailscale.com/docs/features/tailscale-ssh
- https://tailscale.com/docs/reference/connection-types
- https://tailscale.com/docs/reference/faq/other-vpns
- https://tailscale.com/pricing

Keputusan berikutnya ditentukan melalui spike: apakah service terpisah diperlukan untuk seluruh target penggunaan, matriks Windows minimum, ambang performa berdasarkan baseline, dan pilihan detail library. Tidak perlu menunggu keputusan ini untuk memulai P0.
