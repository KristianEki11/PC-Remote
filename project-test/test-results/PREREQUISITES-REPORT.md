# Prerequisites Check Report
Generated: 2026-06-01 02:31:00+07:00

## Base Requirements
| Tool | Status | Version |
|---|---|---|
| Go | ✅ | 1.26.3 |
| Flutter | ✅ | 3.44.0 |
| Git | ✅ | 2.54.0.windows.1 |
| PowerShell | ✅ | 5.1.26100.8457 |
| Python | ✅ | 3.14.5 |

## Go Test Tools
| Tool | Status | Notes |
|---|---|---|
| hey | ✅ | path: C:\Users\KristianEki\go\bin\hey.exe |
| govulncheck | ✅ | version: v1.3.0 |
| osv-scanner | ✅ | version: 1.9.2 |
| pprof | ✅ ENABLED | Profiler listener added on localhost:6060 and server binary rebuilt |

## Flutter Test Tools
| Tool | Status | Notes |
|---|---|---|
| flutter_test | ✅ | sdk: flutter |
| mockito | ✅ | version: ^5.7.0 |
| build_runner | ✅ | version: ^2.15.0 |
| integration_test | ✅ | sdk: flutter |

## Windows System Tools
| Tool | Status | Notes |
|---|---|---|
| AudioDeviceCmdlets | ✅ | Module installed and verified. Active playback device: SteelSeries Sonar - Gaming |
| NSSM | ⚠️ NOT IN PATH | Found at C:\Program Files\PCRemote\nssm.exe and d:\remote-pc\installer\tools\nssm.exe |
| PCRemoteServer service | ❌ NOT REGISTERED | Service not registered — TEST-05 service lifecycle tests will be skipped |
| Log directory | ✅ | Writable log directory at d:\remote-pc\server\logs |

## Test Directory Structure
| Directory | Status |
|---|---|
| server/test-results/stress | ✅ |
| server/test-results/security | ✅ |
| server/test-results/benchmark | ✅ |
| server/test-results/system | ✅ |
| server/test-results/pprof | ✅ |
| app/test-results | ✅ |

## Overall Status
[PARTIAL — 5/6 tests can run] — TEST-01 (load), TEST-02 (vuln), TEST-03 (profile), TEST-04 (Flutter unit/integration), and TEST-06 (system tools/APIs) can run. TEST-05 service lifecycle tests will be skipped because PCRemoteServer service is not registered.

## Action Required (if any)
1. To run TEST-05 service lifecycle tests, register the PCRemoteServer service first:
   - Run `nssm install PCRemoteServer [Path to executable]`
   - Or install it via the installer package.
