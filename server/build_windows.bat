@echo off
echo Building PCRemote Server for Windows...
set GOARCH=amd64
set GOOS=windows
set CGO_ENABLED=0

echo [1/2] Building sendkey.exe...
go build -ldflags="-s -w -H windowsgui" -o dist/sendkey.exe ./cmd/sendkey/
if %ERRORLEVEL% NEQ 0 (
    echo FAILED: sendkey.exe build failed
    pause
    exit /b 1
)

echo [2/2] Building pcremote-server.exe...
go build -ldflags="-s -w -H windowsgui" -o dist/pcremote-server.exe .
if %ERRORLEVEL% NEQ 0 (
    echo FAILED: pcremote-server.exe build failed
    pause
    exit /b 1
)

echo.
echo Build complete:
echo   dist/sendkey.exe
echo   dist/pcremote-server.exe
pause
