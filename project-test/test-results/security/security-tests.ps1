# Security Test Script
Write-Host "=== Security Extreme Tests ===" -ForegroundColor Cyan

# Test 1: /health endpoint (no PIN required)
Write-Host "`n[Test 1] /health without PIN:" -ForegroundColor Yellow
try {
    $r = Invoke-WebRequest -Uri 'http://localhost:8000/health' -TimeoutSec 3 -UseBasicParsing
    Write-Host "  Status: $($r.StatusCode)" -ForegroundColor Green
    Write-Host "  Body: $($r.Content)"
} catch {
    Write-Host "  Status: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
}

# Test 2: /health with wrong PIN
Write-Host "`n[Test 2] /health with wrong PIN (0000):" -ForegroundColor Yellow
try {
    $r = Invoke-WebRequest -Uri 'http://localhost:8000/health' -Headers @{'X-PIN'='0000'} -TimeoutSec 3 -UseBasicParsing
    Write-Host "  Status: $($r.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "  Status: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
}

# Test 3: /health with correct PIN
Write-Host "`n[Test 3] /health with correct PIN (1234):" -ForegroundColor Yellow
try {
    $r = Invoke-WebRequest -Uri 'http://localhost:8000/health' -Headers @{'X-PIN'='1234'} -TimeoutSec 3 -UseBasicParsing
    Write-Host "  Status: $($r.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "  Status: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
}

# Test 4: SQL injection on protected endpoint
Write-Host "`n[Test 4] SQL injection on /audio/volume:" -ForegroundColor Yellow
try {
    $r = Invoke-WebRequest -Uri 'http://localhost:8000/audio/volume' -Headers @{'X-PIN'="1234' OR '1'='1"} -Method POST -TimeoutSec 3 -UseBasicParsing
    Write-Host "  Status: $($r.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "  Status: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
}

# Test 5: Long PIN test
Write-Host "`n[Test 5] Long PIN (100 chars) on /audio/volume:" -ForegroundColor Yellow
try {
    $r = Invoke-WebRequest -Uri 'http://localhost:8000/audio/volume' -Headers @{'X-PIN'=('A' * 100)} -Method POST -TimeoutSec 3 -UseBasicParsing
    Write-Host "  Status: $($r.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "  Status: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
}

# Test 6: Path traversal
Write-Host "`n[Test 6] Path traversal on /debug:" -ForegroundColor Yellow
try {
    $r = Invoke-WebRequest -Uri 'http://localhost:8000/debug' -TimeoutSec 3 -UseBasicParsing
    Write-Host "  Status: $($r.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "  Status: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
}

# Test 7: Rate limit check (rapid requests)
Write-Host "`n[Test 7] Rate limit check (10 rapid requests):" -ForegroundColor Yellow
for ($i=0; $i -lt 10; $i++) {
    $pin = "000" + $i.ToString()
    try {
        $r = Invoke-WebRequest -Uri 'http://localhost:8000/audio/volume' -Headers @{'X-PIN'=$pin} -Method POST -TimeoutSec 2 -UseBasicParsing
        Write-Host "  Attempt $i - HTTP $($r.StatusCode)"
    } catch {
        Write-Host "  Attempt $i - HTTP $($_.Exception.Response.StatusCode)"
    }
    Start-Sleep -Milliseconds 50
}

Write-Host "`n=== Test Complete ===" -ForegroundColor Cyan