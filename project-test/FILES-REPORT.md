# Project Test Files Report
# All Test Artifacts Consolidated

**Generated:** 2026-06-01 18:45
**Total Size:** ~8.57 MB
**Location:** `d:\remote-pc\project-test\`

---

## Tools Used During Testing (5 Test Suites)

### TEST-01: Stress & Spike Testing
| Tool | Purpose |
|------|---------|
| hey | HTTP load testing tool (-n, -c flags) |
| PowerShell | Test orchestration, process monitoring |
| curl | HTTP health checks |

### TEST-02: Security Extreme Testing
| Tool | Purpose |
|------|---------|
| govulncheck | Dependency vulnerability scanning |
| curl | HTTP endpoint testing |
| PowerShell | Security test scripts |
| strings (built-in) | Binary analysis for secrets |

### TEST-03: Optimization & Benchmark Testing
| Tool | Purpose |
|------|---------|
| go test | Go microbenchmark execution (-bench, -benchmem) |
| PowerShell | Benchmark orchestration |
| pprof | Go profiling (optional) |

### TEST-04: Flutter App Testing
| Tool | Purpose |
|------|---------|
| flutter test | Widget, integration, performance tests |
| flutter analyze | Static analysis |

### TEST-05: Windows System Integration Testing
| Tool | Purpose |
|------|---------|
| PowerShell | Service management, process monitoring |
| Get-Process | Resource monitoring |
| Get-EventLog | Event log analysis |
| curl | API endpoint testing |
| hey | Load generation for stability tests |

---

## All Generated Files

### Root Level (4 files)
```
project-test/
+-- MASTER-TEST-REPORT.md         (6.36 KB)  - Master test report
+-- run-all-tests.ps1             (25.95 KB) - Master test runner script
+-- test-api.exe                  (8,635 KB) - Test API utility
+-- test_smtc.ps1                 (0.84 KB)  - SMTC test script
```

### app-test-results/ (5 files)
```
+-- app-test-results/
    +-- accessibility-report.txt       (4.48 KB)
    +-- app-test-report.md            (6.21 KB)
    +-- integration-test-report.txt   (3.00 KB)
    +-- performance-timing.json       (2.05 KB)
    +-- widget-test-report.txt        (2.22 KB)
```

### test-results/ (Server Test Results - 22 files)

#### test-results/benchmark/
```
+-- benchmark/
    +-- baseline-idle.txt             (1.06 KB)
    +-- benchmark-report.md           (5.08 KB)
    +-- microbenchmarks.txt           (4.48 KB)
    +-- optimization-comparison.md    (2.87 KB)
    +-- raw-benchmarks.txt           (16.60 KB)
```

#### test-results/pprof/
```
+-- pprof/
    +-- goroutine_snapshot.txt        (2.61 KB)
    +-- README.md                     (0.72 KB)
```

#### test-results/security/
```
+-- security/
    +-- brute-force-test.csv          (0.31 KB)
    +-- dependency-audit.txt          (0.81 KB)
    +-- security-report.md             (7.71 KB)
    +-- security-tests.ps1             (3.07 KB)
    +-- test-brute-force.ps1          (0.85 KB)
    +-- test-fixes.ps1                (2.23 KB)
    +-- test-rate-limit.ps1           (0.79 KB)
```

#### test-results/stress/
```
+-- stress/
    +-- stress-test-report.md          (5.61 KB)
```

#### test-results/system/
```
+-- system/
    +-- event-log-extract.txt         (0.33 KB)
    +-- port-binding-test.txt         (0.39 KB)
    +-- resource-consumption.csv     (0.56 KB)
    +-- service-lifecycle-test.txt   (0.45 KB)
    +-- stability-30min.csv          (0.51 KB)
    +-- stability-30min.txt          (0.35 KB)
    +-- system-test-report.md         (7.19 KB)
    +-- test-crash-recovery.ps1      (1.79 KB)
    +-- test-log.txt                  (0.06 KB)
```

---

## Temporary Files Created During Testing

The following temp files were created but have been cleaned up:
```
/tmp/test_service_lifecycle.sh  - Service lifecycle test script
/tmp/test_crash_recovery.sh    - Crash recovery test script
/tmp/test_resource_monitoring.sh - Resource monitoring script
/tmp/test_load_monitoring.sh   - Load test monitoring script
/tmp/test_port_stability.sh    - Port binding test script
/tmp/test_stability_30min.sh  - 30-min stability test script
/tmp/hey_output.txt           - Hey load test output
```

Note: These temp files are stored in /tmp (Bash temp) and were cleaned up after use.

---

## File Count Summary

| Category | Files | Size |
|----------|-------|------|
| Root Scripts | 4 | ~8.67 MB |
| App Test Results | 5 | ~18 KB |
| Server Test Results | 22 | ~60 KB |
| **Total** | **31 files** | **~8.75 MB** |

---

## Test Reports (Key Outputs)

| Report | Purpose | Location |
|--------|---------|----------|
| MASTER-TEST-REPORT.md | All tests consolidated | Root |
| stress-test-report.md | Load & spike tests | test-results/stress/ |
| security-report.md | Security audit | test-results/security/ |
| benchmark-report.md | Performance benchmarks | test-results/benchmark/ |
| app-test-report.md | Flutter app tests | app-test-results/ |
| system-test-report.md | Windows integration | test-results/system/ |

---

## Test Scripts

| Script | Purpose |
|--------|---------|
| run-all-tests.ps1 | Master test runner (all 5 suites) |
| test_smtc.ps1 | Media transport test |
| test-api.exe | API testing utility |
| test-results/security/*.ps1 | Security test scripts |
| test-results/system/*.ps1 | System test scripts |

---

## Generated by: Claude Code
## Project: PC Remote Controller v2.2.10
