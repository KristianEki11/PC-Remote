# Crash Recovery Test Script
$LOGFILE = "d:\remote-pc\server\test-results\system\crash-recovery-test.txt"

"=== CRASH RECOVERY TEST ===" | Out-File $LOGFILE -Encoding utf8
"Start Time: $(Get-Date)" | Out-File $LOGFILE -Append

# Get current PID
$process = Get-Process pcremote* -ErrorAction SilentlyContinue
$PID = $process.Id
"Current PID: $PID" | Out-File $LOGFILE -Append

# Simulate crash by killing the process
"Simulating crash by killing process $PID..." | Out-File $LOGFILE -Append
"Kill time: $(Get-Date)" | Out-File $LOGFILE -Append

# Kill the process
try {
    Stop-Process -Id $PID -Force -ErrorAction Stop
    "Kill result: SUCCESS" | Out-File $LOGFILE -Append
} catch {
    "Kill result: FAILED - $($_.Exception.Message)" | Out-File $LOGFILE -Append
}

# Wait 5 seconds for recovery
"Waiting 5 seconds for recovery..." | Out-File $LOGFILE -Append
Start-Sleep 5

# Check if server is back
"Recovery check time: $(Get-Date)" | Out-File $LOGFILE -Append

try {
    $health = Invoke-WebRequest -Uri "http://localhost:8000/health" -UseBasicParsing -TimeoutSec 5
    "Health after recovery: $($health.Content)" | Out-File $LOGFILE -Append
    if ($health.Content -match 'ok') {
        "RESULT: SUCCESS - Server recovered" | Out-File $LOGFILE -Append
    } else {
        "RESULT: UNCLEAR - Unexpected health response" | Out-File $LOGFILE -Append
    }
} catch {
    "Health after recovery: FAILED" | Out-File $LOGFILE -Append
    "RESULT: FAILED - Server did not recover" | Out-File $LOGFILE -Append
}

# Get new PID if server recovered
$newProcess = Get-Process pcremote* -ErrorAction SilentlyContinue
if ($newProcess) {
    "New PID after recovery: $($newProcess.Id)" | Out-File $LOGFILE -Append
}

"" | Out-File $LOGFILE -Append
"=== TEST COMPLETED ===" | Out-File $LOGFILE -Append

# Output to console
Get-Content $LOGFILE
