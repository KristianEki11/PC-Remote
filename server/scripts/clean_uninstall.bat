@echo off
NET SESSION >nul 2>&1
IF %ERRORLEVEL% NEQ 0 (
    echo ========================================================
    echo ERROR: script ini harus dijalankan sebagai Administrator!
    echo ========================================================
    echo Silakan klik kanan berkas ini dan pilih "Run as Administrator".
    echo.
    pause
    exit /b 1
)

echo ========================================================
echo Memulai Clean Uninstall PC Remote Server...
echo ========================================================

echo.
echo [1/5] Menghentikan dan menghapus Layanan Windows (Service)...
set "NSSM_EXE=nssm"
where nssm >nul 2>&1
if %ERRORLEVEL% neq 0 (
    if exist "%~dp0..\..\installer\tools\nssm.exe" (
        set "NSSM_EXE=%~dp0..\..\installer\tools\nssm.exe"
    ) else if exist "%~dp0..\installer\tools\nssm.exe" (
        set "NSSM_EXE=%~dp0..\installer\tools\nssm.exe"
    ) else (
        echo NSSM tidak ditemukan. Mencoba menghapus via sc.exe...
    )
)

if "%NSSM_EXE%"=="nssm" (
    sc stop PCRemoteServer >nul 2>&1
    sc delete PCRemoteServer >nul 2>&1
) else (
    "%NSSM_EXE%" stop PCRemoteServer >nul 2>&1
    "%NSSM_EXE%" remove PCRemoteServer confirm >nul 2>&1
)

echo.
echo [2/5] Memastikan tidak ada proses pcremote-server yang terkunci...
taskkill /F /IM pcremote-server.exe >nul 2>&1
taskkill /F /IM PCRemoteDashboard.exe >nul 2>&1
timeout /t 2 /nobreak >nul

echo.
echo [3/5] Menghapus folder distribusi dan instalasi...
if exist "%~dp0..\dist" (
    echo Menghapus folder dist lokal...
    rmdir /S /Q "%~dp0..\dist"
)
if exist "C:\Program Files\PCRemote" (
    echo Menghapus C:\Program Files\PCRemote...
    rmdir /S /Q "C:\Program Files\PCRemote"
)

echo.
echo [4/5] Menghapus aturan Firewall...
netsh advfirewall firewall delete rule name="PCRemote Server" >nul 2>&1

echo.
echo [5/5] Menghapus pintasan (Shortcuts)...
set "USER_PROFILE_DIR=%USERPROFILE%"
if exist "%USER_PROFILE_DIR%\Desktop\PCRemote Dashboard.lnk" (
    del /F "%USER_PROFILE_DIR%\Desktop\PCRemote Dashboard.lnk"
)
if exist "%USER_PROFILE_DIR%\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup\PCRemoteServer.lnk" (
    del /F "%USER_PROFILE_DIR%\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup\PCRemoteServer.lnk"
)
if exist "%USER_PROFILE_DIR%\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\PCRemote" (
    rmdir /S /Q "%USER_PROFILE_DIR%\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\PCRemote"
)

echo.
echo ========================================================
echo Uninstall Bersih Selesai!
echo Semua layanan, berkas kompilasi, dan konfigurasi telah dihapus.
echo ========================================================
pause
