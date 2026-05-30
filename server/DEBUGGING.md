# PCRemote Server — Debugging Runbook

> Last updated from live log analysis: 2026-05-29

---

## Quick Reference

```
# Run unit tests
go test ./... -v

# Run specific test
go test ./handlers/ -run TestAuthMiddleware -v

# Build binary
go build -o pcremote-server.exe .

# HTTP endpoint tests (from scripts/ dir)
test_endpoints.bat

# Service lifecycle tests (requires Admin)
test_service_lifecycle.bat
```

---

## Log Analysis — Known Issues Found

### ISSUE 1: Port-Bind Crash Loop (FIXED)

**Observed in** `logs/server.log` at 21:18–21:19 on 2026-05-29:

```
{"level":"INFO","msg":"PCRemote Server listening on :8000"}
{"level":"ERROR","msg":"Failed to start server","error":"listen tcp :8000: bind: Only one usage of each socket address..."}
{"level":"INFO","msg":"PCRemote Server listening on :8000"}
{"level":"ERROR","msg":"Failed to start server","error":"listen tcp :8000: bind: Only one usage of each socket address..."}
```

**Root cause:** When the Go process was killed during crash recovery testing, the old `main.go` called `os.Exit(1)` from the listener goroutine. This killed the Go process but orphaned the port binding. NSSM restarted the binary, which then failed to bind port 8000 because the orphaned socket was still held by the OS.

**Fix applied:** Replaced `os.Exit(1)` with `quit <- syscall.SIGTERM`, which signals the main goroutine to perform graceful shutdown (releasing the socket) before the process exits. NSSM's configured `AppRestartDelay` of 3000ms then gives the OS time to release the port.

### ISSUE 2: `go vet` Warnings — `unsafe.Pointer` Misuse

**Observed:** 3 warnings from `go vet ./...`:

```
windows\api.go:133: possible misuse of unsafe.Pointer
windows\api.go:134: possible misuse of unsafe.Pointer
windows\api.go:225: possible misuse of unsafe.Pointer
```

**Location:** The `vtbl()` helper and `readUTF16()` functions in `windows/api.go`.

**Assessment:** These are **false positives**. The code correctly dereferences COM vtable pointers and reads null-terminated UTF-16 strings from COM PROPVARIANT data. The Go vet tool flags any `unsafe.Pointer` arithmetic that doesn't follow the exact `unsafe.Pointer` → `uintptr` → `unsafe.Pointer` one-expression pattern, but COM interop requires this pattern. The code is functionally correct and has been verified via `test_api.exe` and live endpoint testing.

**Mitigation:** Add `//nolint` comments if using a linter in CI, or suppress via `-gcflags` if needed.

### ISSUE 3: Config ENV Variable Mismatch (FIXED)

**Observed:** The `.env` file uses `APP_PIN=1234` and `APP_PORT=8000`, but `config.go` only checked for `PORT` (not `APP_PORT`).

**Fix applied:** Added `APP_PORT` as a fallback lookup in `config/config.go`, matching the existing `APP_PIN` fallback pattern.

---

## Diagnostic Runbooks

### PROBLEM: Service shows "Running" but `/health` returns connection refused

**Diagnose:**

1. Check if port is actually open:

   ```cmd
   netstat -an | findstr :8000
   ```

2. Inspect logs for startup errors:

   ```cmd
   type "d:\remote-pc\server\logs\stderr.log"
   type "d:\remote-pc\server\logs\server.log"
   ```

3. Run the binary manually (stop service first):

   ```cmd
   nssm stop PCRemoteServer
   cd /d "d:\remote-pc\server"
   pcremote-server.exe
   ```

4. If manual run works → service environment issue. Verify NSSM working directory:
   ```cmd
   nssm get PCRemoteServer AppDirectory
   ```
   Must return the directory containing `.env` and the binary.

**Resolve:**

```cmd
nssm set PCRemoteServer AppDirectory "d:\remote-pc\server"
nssm restart PCRemoteServer
```

If the port is already in use (orphaned process from crash):

```cmd
taskkill /IM pcremote-server.exe /F
nssm restart PCRemoteServer
```

---

### PROBLEM: Audio control returns 200 but volume doesn't change

**Diagnose:**

1. Run the CLI test tool:

   ```cmd
   d:\remote-pc\server\test_api.exe
   ```

   Select option `[d]` to dump all detected audio devices. Verify Sonar channels appear.

2. Check if the correct device is targeted. The server looks for devices with names containing:
   - `Sonar - Gaming` (gaming channel)
   - `Sonar - Chat` (chat channel)
   - `Sonar - Media` (media channel, also the default volume target)
   - `Sonar - Microphone` (mic channel)

3. Check `server.log` for COM errors (HRESULT codes).

4. If running as a Windows Service, the COM audio API requires an interactive user session.

**Resolve:**

- If COM error `CoInitialize failed`: the service is running as SYSTEM account which has no audio session:

  ```cmd
  nssm set PCRemoteServer ObjectName ".\YourUsername" "YourPassword"
  nssm restart PCRemoteServer
  ```

- If Sonar devices not found: verify SteelSeries GG / Sonar is running.

---

### PROBLEM: Flutter app shows "disconnected" immediately

**Diagnose:**

1. Test from the server PC first:

   ```cmd
   curl http://localhost:8000/health
   ```

2. Test from the network address:

   ```cmd
   curl http://192.168.0.100:8000/health
   ```

3. Check Windows Firewall rule:

   ```cmd
   netsh advfirewall firewall show rule name="PCRemote Server"
   ```

4. Check Flutter app logs for specific error:
   - `Connection refused` → server not running or firewall blocked
   - `Timeout` → wrong IP / network issue
   - `401 Unauthorized` → PIN mismatch

**Resolve:**

- Add firewall rule if missing:

  ```cmd
  netsh advfirewall firewall add rule name="PCRemote Server" dir=in action=allow protocol=TCP localport=8000
  ```

- If 401: verify `.env` PIN matches the PIN in the Flutter app.

- If timeout: ensure the PC's WiFi network profile is set to **Private** (not Public). Public networks block incoming connections even with firewall rules.

---

### PROBLEM: Service crashes on Windows startup (before user login)

**Diagnose:**

1. Check Event Viewer → Windows Logs → System for service errors.
2. Audio COM interfaces require a logged-in desktop session to work.

**Resolve:**

Set the service to delayed auto-start so Windows has time to initialize the user session:

```cmd
nssm set PCRemoteServer Start SERVICE_DELAYED_AUTO_START
nssm restart PCRemoteServer
```

---

### PROBLEM: Port bind error after crash / taskkill

**Diagnose:**

This happens when the Go server is killed without graceful shutdown. The OS may hold the TCP socket in `TIME_WAIT` state for up to 30 seconds.

```cmd
netstat -an | findstr :8000
```

If you see `TIME_WAIT` entries, the port is temporarily held by the OS.

**Resolve:**

Wait for the NSSM restart delay (3 seconds configured) plus TIME_WAIT timeout. If urgent:

```cmd
taskkill /IM pcremote-server.exe /F
timeout /t 5
nssm start PCRemoteServer
```

---

## Architecture Quick Reference

```
main.go
  ├── config/config.go          → Loads .env (PORT/APP_PORT, PIN/APP_PIN)
  ├── middleware/auth.go         → X-PIN header check, request logging
  ├── handlers/
  │   ├── health.go              → GET /health (no auth)
  │   ├── audio.go               → POST /audio/volume, /mute, GET /status, channels
  │   ├── browser.go             → POST /browser/open
  │   ├── media.go               → POST /media/play, /next, /prev
  │   ├── system.go              → POST /system/lock, /shutdown, /shutdown/cancel
  │   └── response.go            → sendJSON helper
  └── windows/
      ├── types.go               → APIInterface + ChannelStatus (platform-agnostic)
      ├── api.go                 → RealAPI{} (Windows COM, //go:build windows)
      └── windows_api_mock.go    → MockAPI{} (stubs, //go:build !windows)
```
