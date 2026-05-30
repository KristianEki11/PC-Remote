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

"%NSSM_EXE%" install PCRemoteServer "%FULL_PATH%\pcremote-server.exe"
"%NSSM_EXE%" set PCRemoteServer AppDirectory "%FULL_PATH%"
"%NSSM_EXE%" set PCRemoteServer AppStdout "%FULL_PATH%\logs\stdout.log"
"%NSSM_EXE%" set PCRemoteServer AppStderr "%FULL_PATH%\logs\stderr.log"
"%NSSM_EXE%" set PCRemoteServer AppRotateFiles 1
"%NSSM_EXE%" set PCRemoteServer AppRestartDelay 3000
"%NSSM_EXE%" start PCRemoteServer

echo Service installed and started. Check logs/ for output.
pause
