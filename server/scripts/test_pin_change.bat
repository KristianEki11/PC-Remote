@echo off
setlocal enabledelayedexpansion
set SERVER=http://localhost:8000
set OLD_PIN=1234
set NEW_PIN=5678

echo === PIN CHANGE INTEGRATION TEST ===
echo.

REM 1. Get current status (should work with old PIN)
echo [1] Testing /audio/status with old PIN
curl -s -X GET %SERVER%/audio/status -H "X-PIN: %OLD_PIN%" | python -m json.tool

REM 2. Change PIN
echo.
echo [2] Changing PIN from %OLD_PIN% to %NEW_PIN%
curl -s -X POST %SERVER%/system/pin ^
  -H "X-PIN: %OLD_PIN%" ^
  -H "Content-Type: application/json" ^
  -d "{\"current_pin\":\"%OLD_PIN%\",\"new_pin\":\"%NEW_PIN%\"}" | python -m json.tool

REM 3. Verify old PIN no longer works
echo.
echo [3] Testing /audio/status with OLD PIN (should fail with 401)
curl -s -X GET %SERVER%/audio/status -H "X-PIN: %OLD_PIN%" | python -m json.tool

REM 4. Verify new PIN works
echo.
echo [4] Testing /audio/status with NEW PIN (should succeed)
curl -s -X GET %SERVER%/audio/status -H "X-PIN: %NEW_PIN%" | python -m json.tool

REM 5. Restart service (requires admin)
echo.
echo [5] Restarting service to verify PIN persisted in .env...
"%PROGRAMFILES%\PCRemote\nssm.exe" stop PCRemoteServer
timeout /t 2
"%PROGRAMFILES%\PCRemote\nssm.exe" start PCRemoteServer
timeout /t 3

REM 6. Test after restart with new PIN
echo.
echo [6] Testing /audio/status after restart with NEW PIN
curl -s -X GET %SERVER%/audio/status -H "X-PIN: %NEW_PIN%" | python -m json.tool

echo.
echo === TEST COMPLETE ===
pause
