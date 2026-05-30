@echo off
NET SESSION >nul 2>&1
IF %ERRORLEVEL% NEQ 0 (
    echo This script must be run as Administrator.
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

pushd "%~dp0.."
set "FULL_PATH=%CD%"
popd

"%NSSM_EXE%" stop PCRemoteServer
timeout /t 2 /nobreak >nul

copy /Y "%FULL_PATH%\pcremote-server.exe" "%FULL_PATH%\pcremote-server.exe"
"%NSSM_EXE%" start PCRemoteServer

echo Service updated and restarted.
pause
