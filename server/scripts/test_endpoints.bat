@echo off
set SERVER=http://localhost:8000
set PIN=1234

if exist ..\.env (
    for /f "usebackq tokens=1,2 delims==" %%a in ("..\.env") do (
        if "%%a"=="PIN" set PIN=%%b
        if "%%a"=="APP_PIN" set PIN=%%b
    )
) else if exist .env (
    for /f "usebackq tokens=1,2 delims==" %%a in (".env") do (
        if "%%a"=="PIN" set PIN=%%b
        if "%%a"=="APP_PIN" set PIN=%%b
    )
)

echo ============================================================
echo   PCRemote Server — HTTP Endpoint Test Suite
echo   Server: %SERVER%   PIN: %PIN%
echo ============================================================

echo.
echo --- 1. HEALTH CHECK (no auth required) ---
curl -s %SERVER%/health
echo.

echo.
echo --- 2. AUTH: wrong pin (expect 401) ---
curl -s -w "  HTTP %%{http_code}" -X POST %SERVER%/audio/volume ^
  -H "X-PIN: wrongpin" ^
  -H "Content-Type: application/json" ^
  -d "{\"level\": 0.5}"
echo.

echo.
echo --- 3. AUTH: no pin header (expect 401) ---
curl -s -w "  HTTP %%{http_code}" -X POST %SERVER%/audio/volume ^
  -H "Content-Type: application/json" ^
  -d "{\"level\": 0.5}"
echo.

echo.
echo --- 4. AUDIO: set volume to 50%% ---
curl -s -w "  HTTP %%{http_code}" -X POST %SERVER%/audio/volume ^
  -H "X-PIN: %PIN%" ^
  -H "Content-Type: application/json" ^
  -d "{\"level\": 0.5}"
echo.

echo.
echo --- 5. AUDIO: invalid volume (expect 400) ---
curl -s -w "  HTTP %%{http_code}" -X POST %SERVER%/audio/volume ^
  -H "X-PIN: %PIN%" ^
  -H "Content-Type: application/json" ^
  -d "{\"level\": 1.5}"
echo.

echo.
echo --- 6. AUDIO: get status ---
curl -s -w "  HTTP %%{http_code}" -X GET %SERVER%/audio/status ^
  -H "X-PIN: %PIN%"
echo.

echo.
echo --- 7. AUDIO: mute ---
curl -s -w "  HTTP %%{http_code}" -X POST %SERVER%/audio/mute ^
  -H "X-PIN: %PIN%" ^
  -H "Content-Type: application/json" ^
  -d "{\"muted\": true}"
echo.

echo.
echo --- 8. AUDIO: unmute ---
curl -s -w "  HTTP %%{http_code}" -X POST %SERVER%/audio/mute ^
  -H "X-PIN: %PIN%" ^
  -H "Content-Type: application/json" ^
  -d "{\"muted\": false}"
echo.

echo.
echo --- 9. AUDIO: get all Sonar channels ---
curl -s -w "  HTTP %%{http_code}" -X GET %SERVER%/audio/channels ^
  -H "X-PIN: %PIN%"
echo.

echo.
echo --- 10. AUDIO: set channel volume (media 80%%) ---
curl -s -w "  HTTP %%{http_code}" -X POST %SERVER%/audio/channel/volume ^
  -H "X-PIN: %PIN%" ^
  -H "Content-Type: application/json" ^
  -d "{\"channel\": \"media\", \"level\": 0.8}"
echo.

echo.
echo --- 11. AUDIO: channel mute (gaming) ---
curl -s -w "  HTTP %%{http_code}" -X POST %SERVER%/audio/channel/mute ^
  -H "X-PIN: %PIN%" ^
  -H "Content-Type: application/json" ^
  -d "{\"channel\": \"gaming\", \"muted\": true}"
echo.

echo.
echo --- 12. MEDIA: play/pause ---
curl -s -w "  HTTP %%{http_code}" -X POST %SERVER%/media/play ^
  -H "X-PIN: %PIN%"
echo.

echo.
echo --- 13. MEDIA: next track ---
curl -s -w "  HTTP %%{http_code}" -X POST %SERVER%/media/next ^
  -H "X-PIN: %PIN%"
echo.

echo.
echo --- 14. MEDIA: prev track ---
curl -s -w "  HTTP %%{http_code}" -X POST %SERVER%/media/prev ^
  -H "X-PIN: %PIN%"
echo.

echo.
echo --- 15. BROWSER: open URL ---
curl -s -w "  HTTP %%{http_code}" -X POST %SERVER%/browser/open ^
  -H "X-PIN: %PIN%" ^
  -H "Content-Type: application/json" ^
  -d "{\"url\": \"https://example.com\"}"
echo.

echo.
echo --- 16. BROWSER: invalid URL (expect 400) ---
curl -s -w "  HTTP %%{http_code}" -X POST %SERVER%/browser/open ^
  -H "X-PIN: %PIN%" ^
  -H "Content-Type: application/json" ^
  -d "{\"url\": \"\"}"
echo.

echo.
echo --- 17. SYSTEM: lock (skipped — uncomment to test) ---
REM curl -s -w "  HTTP %%{http_code}" -X POST %SERVER%/system/lock -H "X-PIN: %PIN%"
echo   [SKIPPED] — would lock the workstation

echo.
echo --- 18. SYSTEM: shutdown cancel (safe to test) ---
curl -s -w "  HTTP %%{http_code}" -X POST %SERVER%/system/shutdown/cancel ^
  -H "X-PIN: %PIN%"
echo.

echo.
echo --- 19. SYSTEM: shutdown with negative delay (expect 400) ---
curl -s -w "  HTTP %%{http_code}" -X POST %SERVER%/system/shutdown ^
  -H "X-PIN: %PIN%" ^
  -H "Content-Type: application/json" ^
  -d "{\"delay_minutes\": -1}"
echo.

echo.
echo ============================================================
echo   Test suite complete. Review HTTP codes above.
echo   200 = success, 400 = bad request, 401 = unauthorized
echo ============================================================
pause
