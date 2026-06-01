# Rate Limit Test Script
Write-Host "=== Rate Limit Test ===" -ForegroundColor Cyan

for ($i = 1; $i -le 10; $i++) {
    $statusCode = $null
    $blocked = $false

    try {
        $null = Invoke-WebRequest -Uri 'http://localhost:8000/audio/volume' -Method POST -Headers @{'X-PIN'='wrong'} -Body '{"level":0.5}' -TimeoutSec 2 -UseBasicParsing -ErrorAction Stop
        $statusCode = 200
    } catch {
        $statusCode = $_.Exception.Response.StatusCode
        if ($statusCode -eq 429) {
            $blocked = $true
        }
    }

    $output = "Attempt $i : HTTP $statusCode"
    if ($blocked) {
        Write-Host $output " [BLOCKED]" -ForegroundColor Green
    } else {
        Write-Host $output
    }

    Start-Sleep -Milliseconds 50
}

Write-Host "`n=== Test Complete ===" -ForegroundColor Cyan