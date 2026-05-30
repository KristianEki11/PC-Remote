# PC Remote Server Installer

## Persyaratan
- Python & pip (untuk PyInstaller)
- NSIS (Nullsoft Scriptable Install System)

## 1. Cara Instalasi NSIS di Windows
1. Download installer NSIS dari situs resminya: [https://nsis.sourceforge.io/Download](https://nsis.sourceforge.io/Download)
2. Jalankan file `.exe` yang di-download dan ikuti instruksi instalasi default (Next > Next > Install).
3. Setelah instalasi selesai, NSIS akan terintegrasi dengan Windows Explorer sehingga Anda bisa melakukan klik kanan pada file `.nsi` untuk melakukan *Compile*.

## 2. Langkah-langkah Build Installer
1. Buka terminal atau command prompt.
2. Pindah ke direktori `server`: 
   ```cmd
   cd d:\remote-pc\server
   ```
3. Jalankan script batch untuk mem-build executable dari Python:
   ```cmd
   build_exe.bat
   ```
   *(Script ini akan menginstal PyInstaller dan membuat single executable di `server\dist\PCRemoteServer.exe`)*
4. Buka Windows Explorer dan arahkan ke direktori `d:\remote-pc\installer\`.
5. Klik kanan pada file `pc_remote.nsi` dan pilih **Compile NSIS Script**.
6. Tunggu proses kompilasi selesai. Hasilnya berupa file installer `PCRemoteInstaller.exe` akan muncul di folder yang sama.

## 3. Lokasi File APK
Untuk aplikasi klien di Android, APK sudah di-build dan tersedia di lokasi berikut:
```
d:\remote-pc\app\build\app\outputs\flutter-apk\app-release.apk
```
Anda dapat memindahkan file APK ini ke perangkat Android Anda dan menginstalnya secara langsung.
