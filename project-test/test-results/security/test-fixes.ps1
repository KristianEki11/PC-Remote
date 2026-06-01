# Security Fix Verification Script
Write-Host "=== Security Fix Verification ===" -ForegroundColor Cyan

# Test 1: /health without PIN (should be minimal, no version)
Write-Host "`n[Test 1] /health without PIN:" -ForegroundColor Yellow
try {
    $r = Invoke-WebRequest -Uri 'http://localhost:8000/health' -TimeoutSec 3 -UseBasicParsing
    Write-Host "  Status: $($r.StatusCode)"
    Write-Host "  Body: $($r.Content)"
    if ($r.Content -match "version") {
        Write-Host "  ❌ FAIL: Version still disclosed" -ForegroundColor Red
    } else {
        Write-Host "  ✅ PASS: No version in response" -ForegroundColor Green
    }
} catch {
    Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 2: /health with correct PIN (should include version)
Write-Host "`n[Test 2] /health with correct PIN:" -ForegroundColor Yellow
try {
    $r = Invoke-WebRequest -Uri 'http://localhost:8000/health' -Headers @{'X-PIN'='1234'} -TimeoutSec 3 -UseBasicParsing
    Write-Host "  Status: $($r.StatusCode)"
    Write-Host "  Body: $($r.Content)"
    if ($r.Content -match "version") {
        Write-Host "  ✅ PASS: Version included for authenticated" -ForegroundColor Green
    } else {
        Write-Host "  ❌ FAIL: Version missing for authenticated" -ForegroundColor Red
    }
} catch {
    Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 3: Rate limiting on 5 failed attempts (should block at 6th)
Write-Host "`n[Test 3] Rate limiting on failed auth:" -ForegroundColor Yellow
Write-Host "  Sending 6 wrong PIN attempts..." -ForegroundColor Cyan
for ($i = 1; $i -le 6; $i++) {
    try {
        $r = Invoke-WebRequest -Uri 'http://localhost:8000/audio/volume' -Method POST -Headers @{'X-PIN'='wrong'} -Body '{"level":0.5}' -TimeoutSec 3 -UseBasicParsing
        Write-Host "  Attempt $i: $($r.StatusCode) (no block)" -ForegroundColor White
    } catch {
        $status = $_.Exception.Response.StatusCode
        if ($status -eq 429) {
            Write-Host "  Attempt $i: $status ✅ BLOCKED (rate limit)" -ForegroundColor Green
        } else {
            Write-Host "  Attempt $i: $status" -ForegroundColor White
        }
    }
    Start-Sleep -Milliseconds 100
}

Write-Host "`n=== Test Complete ===" -ForegroundColor Cyan