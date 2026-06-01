# ============================================================
# PC REMOTE CONTROLLER — Stress Test Script
# Tests: Baseline, Moderate, High Load, Extreme, Spike
# ============================================================

$ErrorActionPreference = "Stop"
$Script:ProjectRoot = "d:\remote-pc"
$Script:ServerPort = "8000"
$Script:PIN = "1234"
$Script:ReportPath = Join-Path $Script:ProjectRoot "server\test-results\stress\stress-test-report.md"

# Colors
$Colors = @{
    Green  = "`e[92m"
    Red    = "`e[91m"
    Yellow = "`e[93m"
    Cyan   = "`e[96m"
    Bold   = "`e[1m"
    Reset  = "`e[0m"
}

function Write-Step { param($msg) Write-Host "${Colors.Cyan}[TEST]${Colors.Reset} $msg" }
function Write-Success { param($msg) Write-Host "${Colors.Green}[PASS]${Colors.Reset} $msg" }
function Write-Fail { param($msg) Write-Host "${Colors.Red}[FAIL]${Colors.Reset} $msg" }

# ─────────────────────────────────────────────────────────────
# PREREQUISITES
# ─────────────────────────────────────────────────────────────
Write-Host "`n${Colors.Bold}═══════════════════════════════════════════════${Colors.Reset}"
Write-Host "${Colors.Bold}    PC REMOTE CONTROLLER — STRESS TEST SUITE    ${Colors.Reset}"
Write-Host "${Colors.Bold}═══════════════════════════════════════════════${Colors.Reset}`n"

Write-Step "Checking prerequisites..."
$heyPath = Get-Command hey -ErrorAction SilentlyContinue
if (-not $heyPath) {
    Write-Fail "hey not found. Install: go install github.com/rakyll/hey@latest"
    exit 1
}
Write-Success "hey: $($heyPath.Source)"

# Check server
try {
    $health = Invoke-WebRequest -Uri "http://localhost:$Script:ServerPort/health" -UseBasicParsing -TimeoutSec 5
    Write-Success "Server running on port $Script:ServerPort"
} catch {
    Write-Fail "Server not running on port $Script:ServerPort"
    Write-Host "Start server: .\dist\pcremote-server.exe"
    exit 1
}

# Ensure output directory
$null = New-Item -ItemType Directory -Path (Join-Path $Script:ProjectRoot "server\test-results\pprof") -Force

# ─────────────────────────────────────────────────────────────
# TEST FUNCTIONS
# ─────────────────────────────────────────────────────────────

function Run-HeyTest {
    param(
        [string]$Name,
        [int]$Requests,
        [int]$Concurrency,
        [string]$URL,
        [string]$Method = "GET"
    )

    Write-Host "`n${Colors.Bold}─── $Name ───${Colors.Reset}"
    Write-Host "Requests: $Requests, Concurrency: $Concurrency"

    $output = hey -n $Requests -c $Concurrency -H "X-PIN: $Script:PIN" $URL 2>&1

    # Parse results
    $results = @{
        Name = $Name
        Requests = $Requests
        Concurrency = $Concurrency
        Output = $output
    }

    # Extract metrics from hey output
    if ($output -match "Requests/sec:\s+([\d.]+)") {
        $results.RPS = [double]$matches[1]
    }
    if ($output -match "Average\s+([\d.]+) ms") {
        $results.AvgMs = [double]$matches[1]
    }
    if ($output -match "requests in\s+([\d.]+) s") {
        $results.Duration = [double]$matches[1]
    }
    if ($output -match "200\s+responses") {
        $results.Status200 = $true
    }
    if ($output -match "Non-2xx responses:\s+(\d+)") {
        $results.Errors = [int]$matches[1]
    }

    return $results
}

# ─────────────────────────────────────────────────────────────
# RUN TESTS
# ─────────────────────────────────────────────────────────────

$results = @()

# Test A — Baseline
Write-Step "Test A: Baseline (10 concurrent, 500 requests)"
$results += Run-HeyTest -Name "Test A: Baseline" -Requests 500 -Concurrency 10 -URL "http://localhost:$Script:ServerPort/health"
Start-Sleep -Seconds 2

# Test B — Moderate Load
Write-Step "Test B: Moderate Load (50 concurrent, 2000 requests)"
$results += Run-HeyTest -Name "Test B: Moderate Load" -Requests 2000 -Concurrency 50 -URL "http://localhost:$Script:ServerPort/health"
Start-Sleep -Seconds 2

# Test C — High Load
Write-Step "Test C: High Load (100 concurrent, 5000 requests)"
$results += Run-HeyTest -Name "Test C: High Load" -Requests 5000 -Concurrency 100 -URL "http://localhost:$Script:ServerPort/health"
Start-Sleep -Seconds 2

# Test D — Extreme Stress (if tests A-C pass)
Write-Step "Test D: Extreme Stress (300 concurrent, 10000 requests)"
$results += Run-HeyTest -Name "Test D: Extreme Stress" -Requests 10000 -Concurrency 300 -URL "http://localhost:$Script:ServerPort/health"
Start-Sleep -Seconds 2

# ─────────────────────────────────────────────────────────────
# SPIKE TEST (Test E)
# ─────────────────────────────────────────────────────────────

Write-Host "`n${Colors.Bold}═══════════════════════════════════════════════${Colors.Reset}"
Write-Host "${Colors.Bold}TEST E: SPIKE TEST${Colors.Reset}"
Write-Host "${Colors.Bold}═══════════════════════════════════════════════${Colors.Reset}`n"

# Round 1: Normal
Write-Step "Spike Round 1: Normal (10 concurrent, 200 requests)"
$spike1 = Run-HeyTest -Name "Spike R1: Normal" -Requests 200 -Concurrency 10 -URL "http://localhost:$Script:ServerPort/health"

# Round 2: Spike (immediately after)
Write-Step "Spike Round 2: Burst (200 concurrent, 2000 requests)"
$spike2 = Run-HeyTest -Name "Spike R2: Burst" -Requests 2000 -Concurrency 200 -URL "http://localhost:$Script:ServerPort/health"

# Round 3: Recovery
Write-Step "Spike Round 3: Recovery (10 concurrent, 200 requests)"
$spike3 = Run-HeyTest -Name "Spike R3: Recovery" -Requests 200 -Concurrency 10 -URL "http://localhost:$Script:ServerPort/health"

# ─────────────────────────────────────────────────────────────
# PPROF SNAPSHOTS
# ─────────────────────────────────────────────────────────────

Write-Step "Capturing pprof snapshots..."

try {
    $goroutineOutput = Invoke-WebRequest -Uri "http://localhost:6060/debug/pprof/goroutine?debug=1" -UseBasicParsing -TimeoutSec 10
    $goroutineOutput.Content | Out-File -FilePath (Join-Path $Script:ProjectRoot "server\test-results\pprof\goroutine_snapshot.txt") -Encoding UTF8
    Write-Success "Goroutine snapshot saved"
} catch {
    Write-Warn "Failed to capture goroutine snapshot: $_"
}

try {
    $heapOutput = Invoke-WebRequest -Uri "http://localhost:6060/debug/pprof/heap" -UseBasicParsing -TimeoutSec 10
    $heapOutput.Content | Out-File -FilePath (Join-Path $Script:ProjectRoot "server\test-results\pprof\heap_snapshot.txt") -Encoding UTF8
    Write-Success "Heap snapshot saved"
} catch {
    Write-Warn "Failed to capture heap snapshot: $_"
}

# ─────────────────────────────────────────────────────────────
# GENERATE REPORT
# ─────────────────────────────────────────────────────────────

Write-Host "`n${Colors.Bold}═══════════════════════════════════════════════${Colors.Reset}"
Write-Host "${Colors.Bold}GENERATING REPORT${Colors.Reset}"
Write-Host "${Colors.Bold}═══════════════════════════════════════════════${Colors.Reset}`n"

$report = @"
# PC Remote Controller — Stress Test Report

**Generated:** $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
**Server:** localhost:$Script:ServerPort
**PIN:** $Script:PIN

---

## Test Summary

| Test | Name | Requests | Concurrency | Avg (ms) | RPS | Errors |
|------|------|----------|-------------|----------|-----|--------|
"@

foreach ($r in $results) {
    $avgMs = if ($r.AvgMs) { $r.AvgMs } else { "N/A" }
    $rps = if ($r.RPS) { $r.RPS } else { "N/A" }
    $errors = if ($r.Errors) { $r.Errors } else { "0" }
    $report += "`n| $($r.Name) | | $($r.Requests) | $($r.Concurrency) | $avgMs | $rps | $errors |"
}

$report += @"

---

## Spike Test Results

| Round | Name | Requests | Concurrency | Avg (ms) | RPS |
|-------|------|----------|-------------|----------|-----|
| R1 | Normal | 200 | 10 | $($spike1.AvgMs ?? "N/A") | $($spike1.RPS ?? "N/A") |
| R2 | Burst | 2000 | 200 | $($spike2.AvgMs ?? "N/A") | $($spike2.RPS ?? "N/A") |
| R3 | Recovery | 200 | 10 | $($spike3.AvgMs ?? "N/A") | $($spike3.RPS ?? "N/A") |

---

## Flags & Analysis

"@

# Check for issues
$flags = @()
foreach ($r in $results) {
    if ($r.AvgMs -and $r.AvgMs -gt 500) {
        $flags += "- **HIGH LATENCY:** $($r.Name) avg $($r.AvgMs)ms > 500ms threshold"
    }
    if ($r.Errors -and $r.Errors -gt 0) {
        $flags += "- **ERRORS:** $($r.Name) had $($r.Errors) errors"
    }
}

if ($flags.Count -eq 0) {
    $report += "`n✅ **No critical flags detected.** Server handles load well.`n"
} else {
    $report += "`n" + ($flags -join "`n") + "`n"
}

# Recovery analysis
$spikeRecovery = $false
if ($spike1.AvgMs -and $spike3.AvgMs) {
    if ($spike3.AvgMs -gt ($spike1.AvgMs * 1.5)) {
        $spikeRecovery = $true
        $report += "`n⚠️ **RECOVERY ISSUE:** Post-spike latency is 1.5x+ baseline`n"
    } else {
        $report += "`n✅ **RECOVERY OK:** Server returns to baseline after spike`n"
    }
}

$report += @"

---

## Raw hey Output

"@

foreach ($r in $results) {
    $report += @"

### $($r.Name)

\`\`\`
$($r.Output)
\`\`\`

"@
}

$report += @"

---

*Report generated by run-stress-test.ps1*
"@

$report | Out-File -FilePath $Script:ReportPath -Encoding UTF8

Write-Success "Report saved: $Script:ReportPath"
Write-Host "`n${Colors.Bold}═══════════════════════════════════════════════${Colors.Reset}"
Write-Host "${Colors.Bold}STRESS TEST COMPLETE${Colors.Reset}"
Write-Host "${Colors.Bold}═══════════════════════════════════════════════${Colors.Reset}`n"

# Print summary
foreach ($r in $results) {
    $status = if ($r.Errors -eq 0) { "${Colors.Green}OK${Colors.Reset}" } else { "${Colors.Red}ERRORS${Colors.Reset}" }
    Write-Host "  $($r.Name): RPS=$($r.RPS), Avg=$($r.AvgMs)ms $status"
}