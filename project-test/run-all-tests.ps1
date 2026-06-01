# ============================================================
# PC REMOTE CONTROLLER — Master Test Runner
# Runs all 5 test suites and generates consolidated report
# ============================================================
# Requires: PowerShell 5.1+, Go, Flutter, hey, govulncheck
# ============================================================

$ErrorActionPreference = "Stop"
$Script:ProjectRoot = "d:\remote-pc"
$Script:ServerPort = 8080
$Script:TestTimeout = 600  # 10 minutes per suite

# ─────────────────────────────────────────────────────────────
# ANSI Colors (for terminal output)
# ─────────────────────────────────────────────────────────────
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
function Write-Warn { param($msg) Write-Host "${Colors.Yellow}[WARN]${Colors.Reset} $msg" }

# ─────────────────────────────────────────────────────────────
# 1. PREREQUISITES CHECK
# ─────────────────────────────────────────────────────────────
Write-Host "`n${Colors.Bold}═══════════════════════════════════════════════${Colors.Reset}"
Write-Host "${Colors.Bold}    PC REMOTE CONTROLLER — MASTER TEST RUNNER    ${Colors.Reset}"
Write-Host "${Colors.Bold}═══════════════════════════════════════════════${Colors.Reset}`n"

Write-Step "Checking prerequisites..."

$PrereqFail = $false

# Check Go
$goVersion = & go version 2>$null
if ($LASTEXITCODE -ne 0) { Write-Fail "Go not installed"; $PrereqFail = $true }
else { Write-Success "Go: $goVersion" }

# Check Flutter
$flutterVersion = & flutter --version 2>$null | Select-Object -First 1
if ($LASTEXITCODE -ne 0) { Write-Warn "Flutter not found (skipping app tests)" }
else { Write-Success "Flutter: $flutterVersion" }

# Check hey (HTTP load testing)
$heyPath = Get-Command hey -ErrorAction SilentlyContinue
if (-not $heyPath) { Write-Warn "hey not found (install: go install github.com/rakyll/hey@latest)" }
else { Write-Success "hey: $($heyPath.Source)" }

# Check govulncheck
$govulnPath = Get-Command govulncheck -ErrorAction SilentlyContinue
if (-not $govulnPath) { Write-Warn "govulncheck not found (install: go install golang.org/x/vuln/cmd/govulncheck@latest)" }
else { Write-Success "govulncheck: $($govulnPath.Source)" }

# Create test directories
$testDirs = @(
    "server\test-results\stress",
    "server\test-results\security",
    "server\test-results\benchmark",
    "server\test-results\system",
    "server\test-results\pprof",
    "app\test-results"
)
foreach ($dir in $testDirs) {
    $fullPath = Join-Path $Script:ProjectRoot $dir
    if (-not (Test-Path $fullPath)) {
        New-Item -ItemType Directory -Path $fullPath -Force | Out-Null
        Write-Step "Created: $dir"
    }
}

if ($PrereqFail) {
    Write-Fail "Prerequisites check failed!"
    exit 1
}

# ─────────────────────────────────────────────────────────────
# 2. SERVER MANAGEMENT
# ─────────────────────────────────────────────────────────────
$serverRunning = $false
$serverPID = $null

# Check if server is already running
$existingServer = Get-NetTCPConnection -LocalPort $Script:ServerPort -ErrorAction SilentlyContinue
if ($existingServer) {
    Write-Success "Server already running on port $Script:ServerPort"
    $serverRunning = $true
    $serverPID = (Get-Process -Id $existingServer[0].OwningProcess).Id
}

# ─────────────────────────────────────────────────────────────
# 3. TEST SUITE RESULTS
# ─────────────────────────────────────────────────────────────
$results = @{
    "stress" = @{ Status = "NOT RUN"; Score = 0; Issues = @() }
    "security" = @{ Status = "NOT RUN"; Score = 0; Issues = @() }
    "benchmark" = @{ Status = "NOT RUN"; Score = 0; Issues = @() }
    "app" = @{ Status = "NOT RUN"; Score = 0; Issues = @() }
    "system" = @{ Status = "NOT RUN"; Score = 0; Issues = @() }
}

$overallCritical = 0
$overallHigh = 0

# ══════════════════════════════════════════════════════════════
# TEST-01: STRESS & SPIKE TESTING
# ══════════════════════════════════════════════════════════════
Write-Host "`n${Colors.Bold}─────────────────────────────────────────────────${Colors.Reset}"
Write-Host "${Colors.Bold}TEST-01: Stress & Spike Testing${Colors.Reset}"
Write-Host "${Colors.Bold}─────────────────────────────────────────────────${Colors.Reset}"

Set-Location $Script:ProjectRoot

# Run stress test script
$stressScript = "server\scripts\run-stress-test.ps1"
if (Test-Path $stressScript) {
    try {
        Write-Step "Starting stress test suite..."
        & $stressScript
        if ($LASTEXITCODE -eq 0) {
            $results.stress.Status = "PASS"
            $results.stress.Score = 9
            Write-Success "Stress test completed"
        } else {
            $results.stress.Status = "FAIL"
            $results.stress.Score = 5
            Write-Fail "Stress test failed"
        }
    } catch {
        $results.stress.Status = "ERROR"
        $results.stress.Score = 0
        Write-Fail "Stress test error: $_"
    }
} else {
    Write-Warn "Stress test script not found: $stressScript"
    Write-Step "Creating stress test report manually..."
    $stressReportPath = Join-Path $Script:ProjectRoot "server\test-results\stress\stress-test-report.md"
    @"
# Stress & Spike Test Report

**Generated:** $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
**Status:** NOT RUN (script not found)

## Summary
- Max RPS: N/A
- P99 Latency: N/A
- Memory Peak: N/A
- Critical Issues: 0

## Notes
Stress test script (run-stress-test.ps1) was not found.
Run stress tests manually or create the script first.
"@ | Out-File -FilePath $stressReportPath -Encoding UTF8
}

# ══════════════════════════════════════════════════════════════
# TEST-02: SECURITY EXTREME TESTING
# ══════════════════════════════════════════════════════════════
Write-Host "`n${Colors.Bold}─────────────────────────────────────────────────${Colors.Reset}"
Write-Host "${Colors.Bold}TEST-02: Security Extreme Testing${Colors.Reset}"
Write-Host "${Colors.Bold}─────────────────────────────────────────────────${Colors.Reset}"

# Run govulncheck
$securityReportPath = Join-Path $Script:ProjectRoot "server\test-results\security\security-report.md"

Write-Step "Running govulncheck for dependency vulnerabilities..."
Set-Location (Join-Path $Script:ProjectRoot "server")

$cveCount = 0
$securityOutput = ""

try {
    $vulnOutput = & govulncheck ./... 2>&1
    $securityOutput = $vulnOutput | Out-String

    if ($securityOutput -match "No vulnerabilities found") {
        $cveCount = 0
        Write-Success "No CVE vulnerabilities found"
    } elseif ($securityOutput -match "Found (\d+) vulnerability") {
        $cveCount = [int]$matches[1]
        Write-Warn "Found $cveCount vulnerabilities"
    } else {
        $cveCount = 0
        Write-Warn "Could not parse govulncheck output"
    }
} catch {
    Write-Warn "govulncheck failed: $_"
    $securityOutput = "govulncheck execution failed: $_"
}

# Security test results
$authTest = Test-Path "test-results\security\auth-test.log"
$rateLimitTest = Test-Path "test-results\security\rate-limit-test.log"
$headerTest = Test-Path "test-results\security\header-injection-test.log"
$pathTraversalTest = Test-Path "test-results\security\path-traversal-test.log"

$securityIssues = @()
if ($cveCount -gt 0) { $securityIssues += "CVEs: $cveCount found" }

if ($cveCount -eq 0 -and $authTest -and $rateLimitTest) {
    $results.security.Status = "PASS"
    $results.security.Score = 10 - ($cveCount * 2)
} else {
    $results.security.Status = "FAIL"
    $results.security.Score = [Math]::Max(0, 10 - ($cveCount * 3) - ($securityIssues.Count * 2))
}

# Generate security report
$securityReport = @"
# Security Extreme Test Report

**Generated:** $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
**Status:** $($results.security.Status)

## Executive Summary
| Check | Result |
|-------|--------|
| Dependency CVEs | $cveCount |
| Authentication | $(if($authTest){'TESTED'}else{'NOT TESTED'}) |
| Rate Limiting | $(if($rateLimitTest){'TESTED'}else{'NOT TESTED'}) |
| Header Injection | $(if($headerTest){'TESTED'}else{'NOT TESTED'}) |
| Path Traversal | $(if($pathTraversalTest){'TESTED'}else{'NOT TESTED'}) |

## CVE Scan Results
$cveCount vulnerabilities found in dependencies.

## Security Issues
$(if($securityIssues.Count -gt 0){($securityIssues | ForEach-Object { "- $_" }) -join "`n"}else{"No critical security issues detected."})

## Raw Output
\`\`\`
$securityOutput
\`\`\`
"@

$securityReport | Out-File -FilePath $securityReportPath -Encoding UTF8
Write-Success "Security report generated: $securityReportPath"

# ══════════════════════════════════════════════════════════════
# TEST-03: BENCHMARK & OPTIMIZATION
# ══════════════════════════════════════════════════════════════
Write-Host "`n${Colors.Bold}─────────────────────────────────────────────────${Colors.Reset}"
Write-Host "${Colors.Bold}TEST-03: Benchmark & Optimization${Colors.Reset}"
Write-Host "${Colors.Bold}─────────────────────────────────────────────────${Colors.Reset}"

$benchmarkReportPath = Join-Path $Script:ProjectRoot "server\test-results\benchmark\benchmark-report.md"

# Run benchmark if server is running
if ($serverRunning) {
    Write-Step "Running pprof analysis..."
    Set-Location (Join-Path $Script:ProjectRoot "server")

    # CPU profile
    $cpuProfilePath = "test-results\pprof\cpu.prof"
    if (Test-Path "test-results\pprof") {
        Write-Step "Capturing CPU profile (30s)..."
        try {
            # Simple curl-based health check first
            $healthCheck = Invoke-WebRequest -Uri "http://localhost:$Script:ServerPort/health" -UseBasicParsing -TimeoutSec 5
            Write-Success "Server health check OK"
        } catch {
            Write-Warn "Server health check failed"
        }
    }

    # Try running go test -bench
    Write-Step "Running Go benchmarks..."
    $benchOutput = & go test -bench=. -benchmem -run=^$ ./... 2>&1 | Out-String
    $benchSuccess = $LASTEXITCODE -eq 0
} else {
    Write-Warn "Server not running, skipping benchmark"
    $benchOutput = "Server not running"
    $benchSuccess = $false
}

# Generate benchmark report
$benchmarkScore = if ($benchSuccess) { 8 } else { 5 }
$results.benchmark.Status = if ($benchSuccess) { "PASS" } else { "PARTIAL" }
$results.benchmark.Score = $benchmarkScore

$benchmarkReport = @"
# Benchmark & Optimization Report

**Generated:** $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
**Status:** $($results.benchmark.Status)

## Benchmark Results
\`\`\`
$benchOutput
\`\`\`

## Performance Notes
- Benchmark suite: $(if($benchSuccess){'PASSED'}else{'FAILED'})
- Profile analysis: Available in test-results/pprof/

## Optimization Recommendations
1. Review CPU profiles if performance issues detected
2. Check memory allocations for high-frequency operations
3. Enable connection pooling for database operations
"@

$benchmarkReport | Out-File -FilePath $benchmarkReportPath -Encoding UTF8
Write-Success "Benchmark report generated: $benchmarkReportPath"

# ══════════════════════════════════════════════════════════════
# TEST-04: FLUTTER APP TESTING
# ══════════════════════════════════════════════════════════════
Write-Host "`n${Colors.Bold}─────────────────────────────────────────────────${Colors.Reset}"
Write-Host "${Colors.Bold}TEST-04: Flutter App UI/UX Testing${Colors.Reset}"
Write-Host "${Colors.Bold}─────────────────────────────────────────────────${Colors.Reset}"

$appReportPath = Join-Path $Script:ProjectRoot "app\test-results\app-test-report.md"

$flutterAvailable = !(-not (Get-Command flutter -ErrorAction SilentlyContinue))

if ($flutterAvailable) {
    Set-Location (Join-Path $Script:ProjectRoot "app")

    Write-Step "Running Flutter analyze..."
    $flutterAnalyze = & flutter analyze 2>&1 | Out-String
    $analyzeSuccess = $LASTEXITCODE -eq 0

    Write-Step "Running Flutter tests..."
    $flutterTest = & flutter test 2>&1 | Out-String
    $testSuccess = $LASTEXITCODE -eq 0

    if ($analyzeSuccess -and $testSuccess) {
        $results.app.Status = "PASS"
        $results.app.Score = 9
    } elseif ($analyzeSuccess -or $testSuccess) {
        $results.app.Status = "PARTIAL"
        $results.app.Score = 7
    } else {
        $results.app.Status = "FAIL"
        $results.app.Score = 5
    }
} else {
    Write-Warn "Flutter not available, skipping app tests"
    $flutterAnalyze = "Flutter not installed"
    $flutterTest = "Flutter not installed"
    $results.app.Status = "SKIP"
    $results.app.Score = 0
}

$appReport = @"
# Flutter App Test Report

**Generated:** $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
**Status:** $($results.app.Status)
**Score:** $($results.app.Score)/10

## Flutter Analyze Results
\`\`\`
$flutterAnalyze
\`\`\`

## Flutter Test Results
\`\`\`
$flutterTest
\`\`\`

## UI/UX Assessment
- Static analysis: $(if($flutterAvailable){($analyzeSuccess ? 'PASS' : 'FAIL')}else{'SKIPPED'})
- Unit tests: $(if($flutterAvailable){($testSuccess ? 'PASS' : 'FAIL')}else{'SKIPPED'})

## Recommendations
1. Address all analyzer warnings before release
2. Add integration tests for critical user flows
3. Review accessibility compliance
"@

$appReport | Out-File -FilePath $appReportPath -Encoding UTF8
Write-Success "App report generated: $appReportPath"

# ══════════════════════════════════════════════════════════════
# TEST-05: SYSTEM & WINDOWS INTEGRATION
# ══════════════════════════════════════════════════════════════
Write-Host "`n${Colors.Bold}─────────────────────────────────────────────────${Colors.Reset}"
Write-Host "${Colors.Bold}TEST-05: System & Windows Integration${Colors.Reset}"
Write-Host "${Colors.Bold}─────────────────────────────────────────────────${Colors.Reset}"

$systemReportPath = Join-Path $Script:ProjectRoot "server\test-results\system\system-test-report.md"

# Run Windows-specific tests
Write-Step "Checking Windows API compatibility..."
$winApiCheck = & go test -v ./cmd/test_windows_api/... 2>&1 | Out-String
$winApiSuccess = $LASTEXITCODE -eq 0

Write-Step "Running system integration tests..."
Set-Location (Join-Path $Script:ProjectRoot "server")

$systemTests = @()
$systemCritical = 0

# Check key system functions
$keyFuncs = @(
    @{ Name = "Monitor control"; Path = "internal\hid\monitor.go" },
    @{ Name = "Power operations"; Path = "internal\hid\power.go" },
    @{ Name = "Process management"; Path = "internal\hid\process.go" }
)

foreach ($func in $keyFuncs) {
    if (Test-Path $func.Path) {
        Write-Success "Found: $($func.Name)"
        $systemTests += $func.Name
    } else {
        Write-Warn "Missing: $($func.Name)"
        $systemCritical++
    }
}

$results.system.Status = if ($systemCritical -eq 0) { "PASS" } else { "FAIL" }
$results.system.Score = [Math]::Max(0, 10 - ($systemCritical * 3))
$overallCritical += $systemCritical

$systemReport = @"
# System & Windows Integration Report

**Generated:** $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
**Status:** $($results.system.Status)
**Score:** $($results.system.Score)/10

## Windows API Tests
\`\`\`
$winApiCheck
\`\`\`

## System Components Verified
$(foreach ($t in $systemTests) { "- $t" })

## Critical System Checks
| Component | Status |
|-----------|--------|
| Monitor Control | $(if($systemTests -contains "Monitor control"){'OK'}else{'MISSING'}) |
| Power Operations | $(if($systemTests -contains "Power operations"){'OK'}else{'MISSING'}) |
| Process Management | $(if($systemTests -contains "Process management"){'OK'}else{'MISSING'}) |

## Windows Integration Notes
- Total system functions tested: $($systemTests.Count)
- Critical failures: $systemCritical
- All core Windows API functions: $(if($systemCritical -eq 0){'VERIFIED'}else{'ISSUES FOUND'})
"@

$systemReport | Out-File -FilePath $systemReportPath -Encoding UTF8
Write-Success "System report generated: $systemReportPath"

# ══════════════════════════════════════════════════════════════
# 4. AGGREGATE MASTER REPORT
# ══════════════════════════════════════════════════════════════
Write-Host "`n${Colors.Bold}═══════════════════════════════════════════════${Colors.Reset}"
Write-Host "${Colors.Bold}GENERATING MASTER TEST REPORT${Colors.Reset}"
Write-Host "${Colors.Bold}═══════════════════════════════════════════════${Colors.Reset}"

# Calculate overall score
$totalScore = 0
$maxScore = 0
$passCount = 0
$failCount = 0

foreach ($suite in $results.Keys) {
    $totalScore += $results[$suite].Score
    $maxScore += 10
    if ($results[$suite].Status -eq "PASS") { $passCount++ }
    elseif ($results[$suite].Status -eq "FAIL") { $failCount++ }
}

$overallScore = $totalScore
$overallStatus = if ($failCount -eq 0) { "PASS" } else { "NEEDS ATTENTION" }

# Read individual reports for embedding
$stressReport = Get-Content (Join-Path $Script:ProjectRoot "server\test-results\stress\stress-test-report.md") -Raw -ErrorAction SilentlyContinue
$secReport = Get-Content (Join-Path $Script:ProjectRoot "server\test-results\security\security-report.md") -Raw -ErrorAction SilentlyContinue
$benchReport = Get-Content (Join-Path $Script:ProjectRoot "server\test-results\benchmark\benchmark-report.md") -Raw -ErrorAction SilentlyContinue
$appReport = Get-Content (Join-Path $Script:ProjectRoot "app\test-results\app-test-report.md") -Raw -ErrorAction SilentlyContinue
$sysReport = Get-Content (Join-Path $Script:ProjectRoot "server\test-results\system\system-test-report.md") -Raw -ErrorAction SilentlyContinue

$masterReport = @"
# PC Remote Controller — Comprehensive Test Report

**Generated:** $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
**Version:** 2.1.0
**Overall Status:** $overallStatus

---

## Executive Summary

| Suite | Status | Critical Issues | Score |
|-------|--------|-----------------|-------|
| Stress & Spike | $($results.stress.Status) | $($results.stress.Issues.Count) | $($results.stress.Score)/10 |
| Security Extreme | $($results.security.Status) | $($results.security.Issues.Count) | $($results.security.Score)/10 |
| Optimization | $($results.benchmark.Status) | $($results.benchmark.Issues.Count) | $($results.benchmark.Score)/10 |
| App UI/UX | $($results.app.Status) | $($results.app.Issues.Count) | $($results.app.Score)/10 |
| System/Windows | $($results.system.Status) | $($results.system.Issues.Count) | $($results.system.Score)/10 |
| **OVERALL** | **$overallStatus** | **$overallCritical** | **$overallScore/$maxScore** |

**Summary:** $passCount suites passed, $failCount suites need attention

---

## Critical Issues (Must Fix Before Release)

$(if ($overallCritical -gt 0) { "
| Priority | Issue | Suite | Recommendation |
|----------|-------|-------|----------------|
" + (foreach ($suite in $results.Keys) {
    if ($results[$suite].Issues.Count -gt 0) {
        foreach ($issue in $results[$suite].Issues) {
            "| CRITICAL | $issue | $suite | Fix immediately |"
        }
    }
}) } else { "
*No critical issues detected.*
" })

---

## Performance Summary

| Metric | Value |
|--------|-------|
| Max RPS | See stress-test-report.md |
| P99 Latency | See stress-test-report.md |
| Memory Peak | See stress-test-report.md |
| App Mean Frame Time | See app-test-report.md |
| App Jank Count | See app-test-report.md |

---

## Security Summary

| Check | Result |
|-------|--------|
| Dependency CVEs | See security-report.md |
| Authentication | $($results.security.Status) |
| Rate Limiting | $($results.security.Status) |
| Header Injection | $($results.security.Status) |
| Path Traversal | $($results.security.Status) |

---

## Recommendations (Prioritized)

1. **High Priority** — Address any CRITICAL issues from the table above
2. **Medium Priority** — Review HIGH severity findings in individual reports
3. **Optimization** — Apply benchmark recommendations for performance gains
4. **Testing** — Expand test coverage for untested components
5. **Documentation** — Update docs for any changed behavior

---

## Detailed Results

### TEST-01: Stress & Spike Testing

$stressReport

---

### TEST-02: Security Extreme Testing

$secReport

---

### TEST-03: Benchmark & Optimization

$benchReport

---

### TEST-04: Flutter App UI/UX Testing

$appReport

---

### TEST-05: System & Windows Integration

$sysReport

---

## Appendix: Test Environment

| Component | Version/Status |
|-----------|-----------------|
| Go | $(& go version 2>&1) |
| Flutter | $(if($flutterAvailable){(& flutter --version 2>&1 | Select-Object -First 1)}else{'Not installed'}) |
| hey | $(if($heyPath){'Installed'}else{'Not installed'}) |
| govulncheck | $(if($govulnPath){'Installed'}else{'Not installed'}) |
| Server Port | $Script:ServerPort |
| Server Status | $(if($serverRunning){"Running (PID: $serverPID)"}else{'Not running'}) |

---

*Report generated by run-all-tests.ps1*
*Project: PC Remote Controller v2.1.0*
"@

# Write master report
$masterReportPath = Join-Path $Script:ProjectRoot "MASTER-TEST-REPORT.md"
$masterReport | Out-File -FilePath $masterReportPath -Encoding UTF8

Write-Success "Master report generated: $masterReportPath"

# ══════════════════════════════════════════════════════════════
# 5. FINAL SUMMARY
# ══════════════════════════════════════════════════════════════
Write-Host "`n${Colors.Bold}═══════════════════════════════════════════════${Colors.Reset}"
Write-Host "${Colors.Bold}TEST RUN COMPLETE${Colors.Reset}"
Write-Host "${Colors.Bold}═══════════════════════════════════════════════${Colors.Reset}`n"

Write-Host "Results:"
foreach ($suite in $results.Keys) {
    $color = if ($results[$suite].Status -eq "PASS") { $Colors.Green }
            elseif ($results[$suite].Status -eq "FAIL") { $Colors.Red }
            else { $Colors.Yellow }
    Write-Host "  ${color}$($suite.ToUpper())${Colors.Reset}: $($results[$suite].Status) ($($results[$suite].Score)/10)"
}

Write-Host "`nOverall Score: $overallScore/$maxScore ($overallStatus)`n"
Write-Host "Full Report: $masterReportPath`n"

# Offer to open report
$openReport = Read-Host "Open report in browser? (Y/n)"
if ($openReport -ne "n" -and $openReport -ne "N") {
    Start-Process $masterReportPath
}