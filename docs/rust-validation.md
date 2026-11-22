# Rust migration build and acceptance evidence

## Reproducible build

On Windows with Rust MSVC toolchain, C++ build tools, Flutter/Android SDK and optionally NSIS installed:

```powershell
./scripts/rust/build.ps1 -Android -Installer
```

The script stops on a failed formatter, Clippy, Rust test/build, Flutter analyzer/test/build, or installer compilation. It uses Cargo.lock and prints the executable SHA-256. CI runs the same checks and uploads the Windows binary/installer plus Android APK. APK release signing must be configured and audited before a public store release; a successful APK build alone is not proof of production signing.

For a local end-to-end check after building the debug executable:

```powershell
./scripts/rust/smoke.ps1
```

The smoke test launches its own server on ports 18443/18444 with an isolated data directory, verifies local setup protection, PIN login, bearer authentication, endpoint metadata and revocation, and makes one read-only audio request. It stops only the process it launched. TLS verification is bypassed solely for these explicit loopback test URLs; this is not a TLS pinning test or a production-client implementation. A native audio unavailable response is reported rather than counted as hardware success. No power, volume or media mutation commands are sent.

## Performance protocol

Run release binaries with equivalent features on the same PC and workload. Record Windows version, CPU, power mode, connected peripherals, installed media apps, commit hashes and Tailscale state. Stop the other implementation before each run. Use server process IDs, not a name shared by old and new executables.

```powershell
./scripts/rust/benchmark.ps1 -ServerProcessIds 1234 -DurationSeconds 900 -IncludeTailscale
```

The script samples root/descendant process CPU and memory, with Tailscale in a separate group. Sum per timestamp, not across timestamps. CPU is normalized to machine capacity. Samples include private bytes and working set, which measure different things. The sampler's own process is excluded. Sampling misses subprocesses that start and exit between intervals; use Windows Performance Recorder/ETW for reliable process-start and short-lived PowerShell counts. Do not compare cold compilation or debug binaries to release binaries.

Collect three runs each: 15-minute idle with phone disconnected, 15-minute connected idle, and 15-minute scripted read/control workload. Commands changing power state require a designated test machine. Measure client-side command latency p50/p95 separately and record reconnect delay after a controlled network interruption. Define numerical acceptance thresholds from Go baseline before judging Rust results. No performance gain is claimed until measurements exist.

## Release gate matrix

| Gate | Evidence required | Current interpretation |
|---|---|---|
| Rust compile/test/lint | Tool output and commit | Build evidence only; no hardware parity inference |
| Flutter analyze/test/APK | Tool output and artifact | Does not prove device connectivity |
| LAN pairing and commands | Physical phone/PC test | Required before release |
| Tailscale Wi-Fi to cellular | Physical phone/PC test | Required before release |
| User B in independent tailnet | Separate account and devices | Must work without developer credentials |
| Unpaired same-tailnet peer | Auth denial | Mandatory isolation check |
| Restart + Windows sign-in | Reconnect and stored pairing | User-session startup supported; before-login control excluded |
| No logged-in desktop session | Explicit unavailable behavior | Full Windows service migration is not yet established |
| Key expiry/revocation/VPN conflict | Clear recovery path | No automatic network-policy bypass |
| Audio/media/browser/system/display | Hardware feature inventory | Test individually; display dimming and screen-off semantics need explicit parity review |
| Installer upgrade/uninstall | Existing identity retained, old Go untouched | Per-user side-by-side installation |
| Performance | Comparable 15-minute runs | Unmeasured until artifacts supplied |

Do not label unexecuted rows passed. A local compile cannot establish reliability on another user's network or guarantee absence of major bugs.
