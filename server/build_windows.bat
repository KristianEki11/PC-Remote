@echo off
echo Building PCRemote Server for Windows...
set GOARCH=amd64
set GOOS=windows
set CGO_ENABLED=0
go build -ldflags="-s -w -H windowsgui" -o dist/pcremote-server.exe .
echo Build complete: dist/pcremote-server.exe
pause
