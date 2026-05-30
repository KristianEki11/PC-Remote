# PC Remote Server — Migration Notes
**Source:** Python 3 + FastAPI (Uvicorn)  
**Target:** Go  
**Audit Date:** 2026-05-27  
**Audited Files:** `main.py`, `auth.py`, `tray_app.py`, `routes/__init__.py`, `routes/audio.py`, `routes/browser.py`, `routes/media.py`, `routes/system.py`, `.env.example`, `requirements.txt`

---

## 1. ENDPOINT INVENTORY

### `POST /auth`
| Field | Detail |
|---|---|
| **Auth required** | No |
| **Rate limit** | 5 requests / minute per IP |
| **Request body** | `{ "pin": string }` |
| **Response (200)** | `{ "token": string, "success": true }` |
| **Response (401)** | `{ "detail": "PIN salah" }` |
| **Response (403)** | `{ "detail": "Access denied. Local network only." }` |
| **Response (429)** | `{ "detail": "Terlalu banyak percobaan. Coba lagi dalam 1 menit." }` |
| **Library** | `secrets.token_hex(16)` for token generation |
| **Special logic** | On successful login, `cleanup_tokens()` is called first to purge expired tokens before issuing a new one. Tokens are stored in an **in-memory dict** (`active_tokens: Dict[str, float]`). |

---

### `GET /health`
| Field | Detail |
|---|---|
| **Auth required** | No |
| **Request body** | None |
| **Response (200)** | `{ "status": "ok", "uptime": int }` (uptime in seconds since server start) |
| **Library** | `time.time()` |

---

### `GET /audio/devices`
| Field | Detail |
|---|---|
| **Auth required** | Yes (Bearer token) |
| **Request body** | None |
| **Response (200)** | `{ "devices": [ { "id": string, "name": string, "volume": int (0–100), "muted": bool } ] }` |
| **Windows API** | `IMMDeviceEnumerator` (CLSID `{BCDE0395-E52F-467C-8E3D-C4579291692E}`), `IAudioEndpointVolume` via pycaw/comtypes |
| **Special logic** | Enumerates only **active render (output)** endpoints (`eRender=0, DEVICE_STATE_ACTIVE=1`). Cross-references `IMMDeviceEnumerator.EnumAudioEndpoints()` result with `AudioUtilities.GetAllDevices()` to filter to output-only. All COM calls are **dispatched to a dedicated background thread** that holds a single `CoInitialize()` apartment. |

---

### `POST /audio/device/volume`
| Field | Detail |
|---|---|
| **Auth required** | Yes |
| **Request body** | `{ "device_id": string, "level": int (0–100) }` |
| **Response (200)** | `{ "success": true, "device_id": string, "level": int }` |
| **Response (400)** | Level out of range |
| **Response (404)** | Device not found |
| **Windows API** | `IAudioEndpointVolume.SetMasterVolumeLevelScalar(level/100.0, None)` |
| **Special logic** | Looks up device by matching `dev.id == device_id` across `AudioUtilities.GetAllDevices()`. Executed in COM worker thread. |

---

### `POST /audio/device/mute`
| Field | Detail |
|---|---|
| **Auth required** | Yes |
| **Request body** | `{ "device_id": string, "mute": bool }` |
| **Response (200)** | `{ "success": true, "device_id": string, "muted": bool }` |
| **Response (404)** | Device not found |
| **Windows API** | `IAudioEndpointVolume.SetMute(mute, None)` |

---

### `GET /audio/volume`
| Field | Detail |
|---|---|
| **Auth required** | Yes |
| **Request body** | None |
| **Response (200)** | `{ "level": int (0–100), "muted": bool }` |
| **Windows API** | `IAudioEndpointVolume.GetMasterVolumeLevelScalar()`, `ISimpleAudioVolume.GetMute()` |
| **Special logic** | Master level comes from the **default Speakers** device. Mute status is computed from **application sessions**: if no sessions exist it falls back to `IAudioEndpointVolume.GetMute()`; if sessions exist, muted = `all sessions are muted`. |

---

### `POST /audio/volume`
| Field | Detail |
|---|---|
| **Auth required** | Yes |
| **Request body** | `{ "level": int (0–100) }` |
| **Response (200)** | `{ "success": true, "level": int }` |
| **Windows API** | `IAudioEndpointVolume.SetMasterVolumeLevelScalar()`, `ISimpleAudioVolume.SetMasterVolume()` |
| **Special logic** | Three-pass set: **(1)** sets default speakers endpoint, **(2)** sets every active render endpoint, **(3)** sets all application session volumes via `ISimpleAudioVolume`. Each pass has independent try/except to tolerate partial failures. |

---

### `POST /audio/mute`
| Field | Detail |
|---|---|
| **Auth required** | Yes |
| **Request body** | None |
| **Response (200)** | `{ "success": true, "muted": bool }` |
| **Windows API** | `ISimpleAudioVolume.GetMute()` / `SetMute()`, `IAudioEndpointVolume.SetMute()` |
| **Special logic** | Toggle logic: `new_mute = any(session NOT muted)` — i.e. mutes everything if **any** session is currently playing; unmutes if **all** are already muted. Applied to all sessions and all render endpoints in two passes. |

---

### `POST /media/playpause`
| Field | Detail |
|---|---|
| **Auth required** | Yes |
| **Request body** | None |
| **Response (200)** | `{ "success": true }` |
| **Windows API** | `user32.keybd_event(0xB3, 0, 0, 0)` then `keybd_event(0xB3, 0, 0x0002, 0)` (VK_MEDIA_PLAY_PAUSE key-down + key-up) |

---

### `POST /media/next`
| Field | Detail |
|---|---|
| **Auth required** | Yes |
| **Request body** | None |
| **Response (200)** | `{ "success": true }` |
| **Windows API** | `user32.keybd_event(0xB0, ...)` (VK_MEDIA_NEXT_TRACK) |

---

### `POST /media/prev`
| Field | Detail |
|---|---|
| **Auth required** | Yes |
| **Request body** | None |
| **Response (200)** | `{ "success": true }` |
| **Windows API** | `user32.keybd_event(0xB1, ...)` (VK_MEDIA_PREV_TRACK) |

---

### `POST /browser/open`
| Field | Detail |
|---|---|
| **Auth required** | Yes |
| **Request body** | `{ "url": string }` |
| **Response (200)** | `{ "success": true, "url": string }` |
| **Response (400)** | URL does not start with `http://` or `https://` |
| **Library** | Python `webbrowser.open(url)` — opens in OS-default browser |
| **Special logic** | URL scheme validation before opening. In Go this maps to `exec.Command("cmd", "/c", "start", url)` or `exec.Command("rundll32", "url.dll,FileProtocolHandler", url)`. |

---

### `POST /system/lock`
| Field | Detail |
|---|---|
| **Auth required** | Yes |
| **Request body** | None |
| **Response (200)** | `{ "success": true }` |
| **Windows API** | `subprocess.run(["rundll32.exe", "user32.dll,LockWorkStation"])` |

---

### `POST /system/sleep`
| Field | Detail |
|---|---|
| **Auth required** | Yes |
| **Request body** | None |
| **Response (200)** | `{ "success": true }` |
| **Windows API** | `subprocess.run(["rundll32.exe", "powrprof.dll,SetSuspendState", "0,1,0"])` |
| **Note** | In Go, prefer `SetSuspendState` via `syscall` / `golang.org/x/sys/windows` for cleaner error handling. |

---

### `POST /system/restart`
| Field | Detail |
|---|---|
| **Auth required** | Yes |
| **Request body** | None |
| **Response (200)** | `{ "success": true, "message": "PC akan restart dalam 5 detik" }` |
| **Windows API** | `subprocess.run(["shutdown", "/r", "/t", "5"])` |

---

### `POST /system/shutdown`
| Field | Detail |
|---|---|
| **Auth required** | Yes |
| **Request body** | None |
| **Response (200)** | `{ "success": true, "message": "PC akan shutdown dalam 5 detik" }` |
| **Windows API** | `subprocess.run(["shutdown", "/s", "/t", "5"])` |

---

## 2. AUTHENTICATION MECHANISM

### How PIN Auth Works
1. **Client** sends `POST /auth` with `{ "pin": "<user_pin>" }`.
2. Server compares `data.pin` to `APP_PIN` (loaded from env).
3. On success: generates a 32-character hex token via `secrets.token_hex(16)`, stores it in the in-memory dict `active_tokens[token] = time.time()`, and returns `{ "token": ..., "success": true }`.
4. **Subsequent requests** must send `Authorization: Bearer <token>` header.
5. The `require_auth` FastAPI dependency extracts the token via `HTTPBearer`, verifies it exists in `active_tokens` and has not expired.

### Where PIN is Stored
- Read from environment variable `APP_PIN` at startup via `python-dotenv` loading `.env`.
- **Default fallback**: `"1234"` (hardcoded in `auth.py` line 19).
- **Go equivalent**: Read from a `.env` file at startup (e.g., with `godotenv`) or embed as config.

### Token Storage
- **In-memory dict** `active_tokens: Dict[str, float]` — not persisted across restarts.
- Token TTL: **24 hours** (`TOKEN_EXPIRY = 86400` seconds).
- Cleanup runs: (a) on each successful login call, and (b) on a periodic background goroutine every **1 hour**.

### Per-Request Validation
- `verify_token()` checks: token present in dict AND `time.time() - timestamp <= 86400`.
- Expired tokens are deleted eagerly on access.
- Local-network IP check also runs inside `require_auth` (redundant with middleware, defensive pattern).

### Rate Limiting
- `slowapi` decorator: `@limiter.limit("5/minute")` on `POST /auth`, keyed by remote IP (`get_remote_address`).
- **Go equivalent**: Implement token bucket or sliding window in middleware (e.g., `golang.org/x/time/rate`).

---

## 3. WINDOWS API CALLS INVENTORY

### Audio — COM Interfaces
| Operation | Interface / Call |
|---|---|
| Enumerate render devices | `IMMDeviceEnumerator.EnumAudioEndpoints(eRender=0, DEVICE_STATE_ACTIVE=1)` |
| Get default speakers | `AudioUtilities.GetSpeakers()` → `IMMDevice` |
| Activate volume control | `IMMDevice.Activate(IAudioEndpointVolume._iid_, CLSCTX_ALL, None)` |
| Get master scalar volume | `IAudioEndpointVolume.GetMasterVolumeLevelScalar()` → float 0.0–1.0 |
| Set master scalar volume | `IAudioEndpointVolume.SetMasterVolumeLevelScalar(scalar, None)` |
| Get mute state | `IAudioEndpointVolume.GetMute()` / `ISimpleAudioVolume.GetMute()` |
| Set mute | `IAudioEndpointVolume.SetMute(bool, None)` / `ISimpleAudioVolume.SetMute(bool, None)` |
| List all audio sessions | `AudioUtilities.GetAllSessions()` → `ISimpleAudioVolume` per session |
| Set per-session volume | `ISimpleAudioVolume.SetMasterVolume(scalar, None)` |
| COM init/uninit | `CoInitialize()` / `CoUninitialize()` — **single dedicated OS thread** (STA model) |

#### Device Enumeration Approach
- CLSID for MMDeviceEnumerator: `{BCDE0395-E52F-467C-8E3D-C4579291692E}`
- pycaw `AudioUtilities.GetAllDevices()` wraps `IMMDeviceEnumerator` + `IPropertyStore` to get friendly names.
- Device identity uses the raw `IMMDevice.GetId()` string (a long GUID-like path string).

#### COM Threading Architecture (Critical for Go Port)
- A **single background daemon thread** (`PCRemoteAudioWorker`) calls `CoInitialize()` on startup and holds it for the process lifetime.
- All COM calls are serialized through a **task queue** (`queue.Queue`). HTTP handlers enqueue an `AudioTask` and block on `task.result_queue.get()`.
- **Go equivalent**: Run audio operations in a dedicated goroutine locked to its OS thread (`runtime.LockOSThread()` + `ole32.CoInitializeEx`), accepting tasks via a channel.

---

### Media Keys
| Key | VK Code | Dispatch |
|---|---|---|
| Play/Pause | `0xB3` | `user32.keybd_event(vk, 0, 0, 0)` then `keybd_event(vk, 0, 0x0002, 0)` |
| Next Track | `0xB0` | same pattern |
| Prev Track | `0xB1` | same pattern |

- **`keybd_event`** is the legacy Win32 API. Go equivalent: use `SendInput` with `INPUT_KEYBOARD` for correctness, or call `keybd_event` via `syscall` directly.

---

### Lock Screen
- `rundll32.exe user32.dll,LockWorkStation` via `subprocess.run`.
- **Go equivalent**: `windows.LockWorkStation()` from `golang.org/x/sys/windows` or direct `syscall.NewLazyDLL("user32.dll").NewProc("LockWorkStation").Call()`.

---

### Sleep
- `rundll32.exe powrprof.dll,SetSuspendState 0,1,0` via `subprocess.run`.
- Arguments: `hibernate=0, forceCritical=1, disableWakeEvent=0`.
- **Go equivalent**: `syscall` call to `powrprof.dll!SetSuspendState`.

---

### Restart / Shutdown
- `shutdown /r /t 5` and `shutdown /s /t 5` via `subprocess.run`.
- **Go equivalent**: `exec.Command("shutdown", "/r", "/t", "5").Run()` — same Win32 CLI, no special API needed.

---

### Browser Open
- Python `webbrowser.open(url)` resolves the default browser via Windows registry and spawns it.
- **Go equivalent**: `exec.Command("cmd", "/c", "start", "", url).Run()` or `exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Run()`.

---

### Single Instance Mutex (tray_app.py)
- `kernel32.CreateMutexW(None, True, "Global\\PCRemoteServer_SingleInstance_Mutex")`
- Error code `183` (ERROR_ALREADY_EXISTS) → show MessageBox + exit.
- **Go equivalent**: Same `CreateMutex` via `syscall` / `golang.org/x/sys/windows` in the system-tray wrapper.

---

## 4. KNOWN ISSUES LOG

No explicit `TODO`, `FIXME`, or `HACK` comments were found in any of the audited source files.

### Detected Workarounds / Implicit Issues

| Location | Pattern | Description |
|---|---|---|
| `auth.py:45` | `"testclient"` in allowed local IPs | Test client string is allowlisted as a local IP — a testing shim baked into production code. Remove or gate behind a `DEBUG` flag in Go. |
| `auth.py:53–60` | 172.x subnet check | The 172.16–31 check uses manual string parsing instead of a proper CIDR library. The logic is correct but brittle; use `net.ParseCIDR` in Go. |
| `auth.py:104` | Indonesian error message | Rate-limit error message `"Terlalu banyak percobaan..."` is hardcoded Indonesian — inconsistent with English-only errors elsewhere. Standardize in Go. |
| `audio.py:117` | `if not speakers or not speakers._dev` | Accesses private `_dev` attribute of pycaw — fragile, breaks if pycaw internals change. Go must directly use COM `IMMDevice` pointer. |
| `audio.py:213–216` | Walrus operator `vol :=` in comprehension | Python 3.8+ walrus operator used — works but obscures intent. Not relevant to Go, but worth noting for logic porting. |
| `audio.py:237–260` | Three-pass volume set with silent failures | Each pass silently swallows exceptions. If default speakers fail, the function still returns `success: true`. Go should log failures and consider returning partial-success status. |
| `tray_app.py:86–90` | `stop()` swallows all exceptions | `server_manager.stop()` has a bare `except Exception: pass` — shutdown errors are invisible. |
| `tray_app.py:67–75` | Port hardcoded to `8000` | `tray_app.py` ignores `APP_PORT` env var; it hardcodes port `8000` in `uvicorn.Config`. `main.py` does read `APP_PORT` when run standalone. Inconsistency — Go should read from config in one place. |
| `browser.py:14` | URL scheme validation only checks prefix | Only validates `http://` or `https://` prefix, not if the URL is otherwise well-formed. Low risk but worth a stricter `url.Parse` check in Go. |
| `system.py` | No confirmation / undo for restart/shutdown | `POST /system/restart` and `/shutdown` trigger with a 5-second delay — no cancel endpoint exists. Consider adding `/system/cancel-shutdown` in Go. |

---

## 5. FILES TO DELETE (Safe to Remove After Migration)

These are Python-ecosystem-specific and have no Go equivalent:

| File/Directory | Reason |
|---|---|
| `server/main.py` | FastAPI app entry point — replaced by Go `main.go` |
| `server/auth.py` | FastAPI auth logic — ported to Go middleware |
| `server/tray_app.py` | Python tray wrapper (pystray + Pillow) — replaced by Go systray library |
| `server/routes/__init__.py` | Python package marker — not needed in Go |
| `server/routes/audio.py` | Python/pycaw audio routes — ported to Go |
| `server/routes/browser.py` | Python webbrowser route — ported to Go |
| `server/routes/media.py` | Python ctypes media key route — ported to Go |
| `server/routes/system.py` | Python subprocess system route — ported to Go |
| `server/requirements.txt` | Python deps — replaced by `go.mod` / `go.sum` |
| `server/PCRemoteServer.spec` | PyInstaller build spec — not needed |
| `server/PCRemoteServerDebug.spec` | PyInstaller debug build spec — not needed |
| `server/build.bat` | Python build script — replace with Go build script |
| `server/build_exe.bat` | Python EXE packaging script — replace with `go build` script |
| `server/__pycache__/` | Python bytecode cache — delete |
| `server/build/` | PyInstaller build artifacts — delete |
| `server/dist/` | PyInstaller distribution output — delete |
| `server/stdout.txt` | Runtime stdout redirect artifact — delete |
| `server/stderr.txt` | Runtime stderr redirect artifact — delete |

---

## 6. FILES TO KEEP AS REFERENCE (Do Not Delete)

| File | Why Keep |
|---|---|
| `server/routes/audio.py` | Contains the complete COM audio control logic; the COM interface IDs, threading model, and three-pass volume set are non-trivial to reconstruct — **primary reference for Go audio port** |
| `server/auth.py` | Contains the full token lifecycle logic, IP allowlist rules (including the 172.x CIDR edge case), and rate-limiting integration |
| `server/routes/media.py` | Contains VK_MEDIA_* key codes and `keybd_event` call pattern |
| `server/routes/system.py` | Contains exact `rundll32`/`shutdown` command strings used for lock/sleep/restart/shutdown |
| `server/.env` | **Keep during migration** — contains real `APP_PIN` and `APP_PORT` values; Go server needs the same env vars |
| `server/.env.example` | Template showing required env var keys: `APP_PIN`, `APP_PORT` — copy to repo root for Go project |
| `server/favicon.ico` | Tray icon asset — reuse in Go systray implementation |
| `server/logs/` | Existing log files — useful for debugging reference during cutover |

---

## APPENDIX — Go Package Recommendations

| Python Library | Go Equivalent |
|---|---|
| `fastapi` + `uvicorn` | `net/http` stdlib or `github.com/gin-gonic/gin` / `github.com/go-chi/chi` |
| `slowapi` (rate limiting) | `golang.org/x/time/rate` |
| `python-dotenv` | `github.com/joho/godotenv` |
| `pydantic` (request models) | `encoding/json` + struct tags |
| `pycaw` + `comtypes` | `github.com/go-ole/go-ole` + direct `IMMDeviceEnumerator` COM calls |
| `ctypes.windll.user32` | `golang.org/x/sys/windows` or `syscall.NewLazyDLL` |
| `pystray` + `Pillow` | `github.com/getlantern/systray` |
| `webbrowser.open` | `exec.Command("cmd", "/c", "start", "", url)` |
| `secrets.token_hex` | `crypto/rand` + `encoding/hex` |
| PyInstaller | `go build -ldflags="-H windowsgui"` (no extra tooling needed) |
