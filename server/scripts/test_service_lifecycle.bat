@echo off
NET SESSION >nul 2>&1
IF %ERRORLEVEL% NEQ 0 (
    echo This script must be run as Administrator.
    echo Please open an Elevated Command Prompt ^(Run as Administrator^) to run this lifecycle test.
    pause
    exit /b 1
)

set "NSSM_EXE=nssm"
where nssm >nul 2>&1
if %ERRORLEVEL% neq 0 (
    if exist "%~dp0..\..\installer\tools\nssm.exe" (
        set "NSSM_EXE=%~dp0..\..\installer\tools\nssm.exe"
    ) else if exist "%~dp0..\installer\tools\nssm.exe" (
        set "NSSM_EXE=%~dp0..\installer\tools\nssm.exe"
    ) else (
        echo NSSM is not installed or not in PATH, and installer tools not found.
        pause
        exit /b 1
    )
)

echo === 1. Checking service status ===
sc query PCRemoteServer

echo.
echo === 2. Verifying port 8000 listening ===
netstat -an | findstr :8000

echo.
echo === 3. Testing crash recovery ===
echo Finding PID of pcremote-server.exe (the Go application)...
set APP_PID=
for /f "usebackq tokens=2 delims=," %%i in (`tasklist /FI "IMAGENAME eq pcremote-server.exe" /FO CSV /NH 2^>nul`) do (
    set "APP_PID=%%~i"
)
if "%APP_PID%"=="" (
    echo pcremote-server.exe process not found. Is the service running?
    goto next_step
)
echo Found application PID: %APP_PID%
echo Killing application process to simulate a crash (NSSM wrapper will remain alive)...
taskkill /PID %APP_PID% /F
echo Waiting 5 seconds for NSSM auto-recovery delay...
timeout /t 5
echo Checking service status after crash recovery...
sc query PCRemoteServer
echo Testing health endpoint...
curl -s http://localhost:8000/health

:next_step
echo.
echo === 4. Testing Stop/Start via NSSM ===
echo Stopping PCRemoteServer...
"%NSSM_EXE%" stop PCRemoteServer
timeout /t 2
echo Starting PCRemoteServer...
"%NSSM_EXE%" start PCRemoteServer
timeout /t 3
echo Testing health endpoint...
curl -s http://localhost:8000/health

pause
