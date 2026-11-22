# Audit kualitas kode — 13 November 2026

Audit terhadap working tree termasuk perubahan yang sudah ada. Tidak mengubah kode produk. Ini pemeriksaan statis terarah, bukan jaminan semua bug ditemukan atau pengujian perangkat Windows/Android menyeluruh. Laporan audit lama tidak dianggap bukti keadaan sekarang.

## Kesimpulan

Struktur dasar cukup jelas (handler, middleware, Windows API, model Flutter), tetapi belum rapi secara konsisten. Masalah terpenting berada di semantik request, state async, duplikasi, dan quality gate; bukan sekadar panjang file.

## Prioritas tinggi

1. `app/lib/services/api_service.dart:257`: POST otomatis dikirim ulang ke fallback setelah exception/timeout 2,5 detik. Timeout tidak membuktikan server belum menjalankan perintah. Play/pause bisa dieksekusi dua kali; next bisa melewati dua lagu. Pisahkan retry operasi idempotent dari command, atau gunakan request ID dan deduplikasi server.
2. `server/windows/media.go:494`: `mediaCacheTime` dan `mediaCacheVal` dibaca/ditulis tanpa mutex oleh handler HTTP konkuren. Ini data race secara struktur kode; belum direproduksi dengan race detector. Gunakan sinkronisasi dan gabungkan fetch simultan agar tidak meluncurkan banyak PowerShell.
3. `web/src/app/connect/page.tsx:164`: command menampilkan label sebelum hasil diketahui dan tidak memeriksa `res.ok`. Respons 401/500 dianggap selesai. Auto-reconnect di baris 38 juga menetapkan connected sebelum validasi, lalu langsung membaca JSON tanpa memeriksa status. Periksa status, bedakan unauthorized/offline, dan lakukan rollback UI.

## Prioritas sedang

4. `app/lib/models/audio_state.dart:46,79` dan `app/lib/widgets/audio_card.dart:202,311`: slider sudah menulis nilai baru lewat update lokal sebelum onChangeEnd memanggil setter. Akibatnya originalVolume/originalDevice berisi nilai baru; rollback ketika request gagal tidak mengembalikan nilai server sebelumnya. Pisahkan draft slider dan confirmed state.
5. `app/lib/models/audio_state.dart:79,99`: index perangkat disimpan melintasi await, sementara fetchDevices dapat mengganti daftar. Saat gagal, rollback dapat menimpa perangkat lain atau mengakses index yang hilang. Cari ulang berdasarkan ID dan gunakan versi operasi agar respons lama tidak menimpa perubahan baru.
6. `app/lib/models/media_state.dart:26` dan `app/lib/screens/dashboard_screen.dart:103`: timer periodik tidak menjaga satu request aktif. Latensi melebihi interval menimbulkan overlap dan respons usang. startPolling juga membatalkan/membuat timer pada setiap ping sukses. Buat start idempotent dan jadwalkan poll berikutnya setelah request selesai.
7. `app/lib/services/api_service.dart:86`: pemilihan client TLS memakai substring `.com/.net/.org/.site` pada URL. Domain di luar daftar diperlakukan sebagai LAN. Gunakan konfigurasi jenis endpoint dan parsing URI.

## Duplikasi dan struktur

8. `app/lib/services/api_service.dart:211,257`: _get/_post menggandakan pemilihan rute, timeout, header, 401, dan error. Ekstrak transport bersama dengan kebijakan retry eksplisit; jangan menyamakan retry semua metode.
9. `server/windows/media.go:30,99`: runInUserSession dan runInUserSessionStart menggandakan deteksi session serta create/run/delete scheduled task. Ekstrak bagian scheduled task; pertahankan perbedaan wait versus start yang memang diperlukan.
10. `server/handlers/media.go:18`: lima handler command hampir identik (method guard, SendMediaKey, log, JSON). Factory kecil dengan parameter action cukup; tidak perlu framework generik.
11. `app/lib/models/media_state.dart:66,79`: next/prev mengulang loading, request, notify dan delayed refresh. Helper command dapat menyatukan bagian tersebut.
12. ApiService (623 baris) mengurus transport, preferences, TLS, autentikasi, navigasi, snackbar dan pemeriksaan release. Pisahkan transport/session dari efek UI. Dashboard (484 baris), SystemCard (533), web connect (437) mencampur orchestration dengan UI; pecah berdasarkan tanggung jawab, bukan target jumlah baris.
13. `server/tray/qr_page.go:190`: HTML/CSS/JS besar ditanam dalam string Go (file 937 baris). Pindahkan ke asset embedded/template agar bisa diformat dan diperiksa terpisah.

## Kode mati dan kandidat cleanup

14. `web/src/app/page.tsx`: lint mengonfirmasi enam import tidak dipakai: ChevronRight, ShieldCheck, Terminal, Wifi, CheckCircle2, HardDrive.
15. `app/lib/models/audio_device.dart:40`: toJson tidak memiliki pemanggil pada source yang ditelusuri. Kandidat dihapus sampai serialization dibutuhkan.
16. `app/lib/services/api_service.dart:99`: parameter overrideProtocol tidak pernah diberikan pemanggil. Cabang tersebut tidak terpakai saat ini.
17. `app/pubspec.yaml:19`: media_kit, media_kit_video, media_kit_libs_video dan web_socket_channel terdaftar tanpa import/pemakaian pada app/lib. Kandidat dependency sisa; verifikasi build setelah penghapusan, khususnya plugin native.
18. `web/public/{next,vercel,window,globe,file}.svg`: tidak ditemukan referensi di source web. Kandidat asset scaffold, bukan bukti bahwa URL publiknya tidak pernah diakses.
19. authPin hanya alias authToken dan masih dipakai MediaState; bukan dead code, tetapi namanya menyesatkan. Wrapper Windows API dan mock platform juga bukan kode mati: memiliki tujuan dispatch/build tag.

## Konsistensi dan verifikasi

- `server/windows/types.go:12` memberi komentar volume 0–100 sementara AudioDevice menganggap payload 0–1. Seragamkan dokumentasi unit dengan implementasi API.
- `app/lib/services/api_service.dart:308` menyebut sesi 24 jam; cek konsistensi dengan TTL server sebelum menjadikan komentar kontrak.
- `npm run lint`: gagal, 3 error dan 6 warning. Error di connect/page.tsx:42 (set-state-in-effect), :164 (any), page.tsx:340 (unescaped apostrophe).
- Workflow web hanya build/deploy tanpa lint. Tidak ditemukan workflow Flutter dalam file CI terlacak. Test Flutter yang tersedia hanya smoke test login, belum mencakup failover/rollback/polling.
- `go test ./auth ./middleware`: auth lulus; middleware gagal pada TestIsLocalhost/Reverse_Proxy_(X-Forwarded-For) dan External_Host_Header_Targeting_Tunnel. Ini menunjukkan ketidaksesuaian implementasi/ekspektasi test yang perlu ditelaah, bukan otomatis bukti exploit.
- `go test ./handlers`: lulus. `go vet ./...`: gagal dengan tiga possible misuse of unsafe.Pointer pada windows/api.go:114,115,206; perlu review interop, bukan otomatis bukti pointer tidak valid.
- Flutter berhasil menyelesaikan dependency resolution tetapi analisis belum selesai saat laporan ditulis; hasil analyze/test belum tersedia. Tidak menjalankan remote shutdown/sleep/restart maupun tes perangkat nyata.
- Skill graphify dibaca, tetapi pipeline graph tidak dijalankan; temuan berdasarkan pembacaan source dan pencarian referensi, bukan graph yang dihasilkan.

## Urutan perbaikan

Perbaiki retry command, cache race dan respons web; kemudian rollback/polling. Setelah perilaku terlindungi test terarah, satukan duplikasi, hapus kandidat yang terbukti tidak dibutuhkan, pecah tanggung jawab file besar, dan jadikan lint/analyze/test sebagai quality gate.
