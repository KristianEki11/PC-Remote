# Brute Force Rate Limiting Test
$results = @()
for ($i=0; $i -lt 10; $i++) {
    $pin = "000$i"
    try {
        $r = Invoke-WebRequest -Uri 'http://localhost:8000/health' -Headers @{'X-PIN'=$pin} -SkipHttpErrorCheck -TimeoutSec 3
        $results += [PSCustomObject]@{
            Attempt = $i
            PIN = $pin
            StatusCode = $r.StatusCode
            Success = ($r.StatusCode -eq 200)
        }
        Write-Host "Attempt $i ($pin): $($r.StatusCode)"
    } catch {
        $results += [PSCustomObject]@{
            Attempt = $i
            PIN = $pin
            StatusCode = "ERROR"
            Success = $false
        }
        Write-Host "Attempt $i ($pin): ERROR"
    }
    Start-Sleep -Milliseconds 100
}
$results | Format-Table
$results | Export-Csv -Path "d:/remote-pc/server/test-results/security/brute-force-test.csv" -NoTypeInformation