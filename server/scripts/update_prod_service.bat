@echo off
NET SESSION >nul 2>&1
if errorlevel 1 goto :no_admin

set "PROD_DIR=C:\Program Files\PCRemote"
set "BIN_NAME=pcremote-server.exe"

echo [1/4] Menghentikan service PCRemoteServer...
"%PROD_DIR%\nssm.exe" stop PCRemoteServer

timeout /t 2 /nobreak >nul

echo [2/4] Mengompilasi program Go terbaru...
cd /d "%~dp0.."
go build -o "%BIN_NAME%"
if errorlevel 1 goto :compile_err
go build -o PCRemoteDashboard.exe cmd\test_api\main.go
if errorlevel 1 goto :compile_err

echo [3/4] Menyalin binary baru ke %PROD_DIR%...
copy /Y "%BIN_NAME%" "%PROD_DIR%\%BIN_NAME%"
if errorlevel 1 goto :copy_err
copy /Y PCRemoteDashboard.exe "%PROD_DIR%\PCRemoteDashboard.exe"
if errorlevel 1 goto :copy_err

echo [4/4] Memulai kembali service PCRemoteServer...
"%PROD_DIR%\nssm.exe" start PCRemoteServer
if errorlevel 1 goto :start_err

echo [SUKSES] Service berhasil diperbarui ke versi terbaru dengan dukungan multi-device audio!
pause
exit /b 0

:no_admin
echo [ERROR] Script ini harus dijalankan sebagai Administrator (Run as Administrator).
pause
exit /b 1

:compile_err
echo [ERROR] Gagal mengompilasi kode Go.
pause
exit /b 1

:copy_err
echo [ERROR] Gagal menyalin binary ke %PROD_DIR%. Pastikan file tidak sedang dikunci.
pause
exit /b 1

:start_err
echo [ERROR] Gagal memulai service PCRemoteServer.
pause
exit /b 1
