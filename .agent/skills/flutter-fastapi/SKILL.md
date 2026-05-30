# Example Pattern: Flutter → FastAPI (PC Remote Controller)

## Description
A battle-tested architectural pattern for local network (LAN) integration between a Flutter mobile client and a Windows FastAPI server for remote PC control. Includes COM thread safety, optimistic UI with rollback logic, localized error handling, and centralized state management.

## When to trigger
- LAN communication architectures between a mobile client and a local server
- Optimistic UI patterns for network-bound toggle/slider interactions
- Integrating Windows audio/system APIs from a Python backend
- State synchronization between multiple mobile clients and a single server
- Keywords: "PC Remote pattern", "Flutter FastAPI", "COM context manager", "local network control", "optimistic toggle"

## Agent persona
- **Name**: CO-DESIGNED BY RIKU, MAYA & VIKTOR
- **Domain**: Local network remote management — backend API ↔ mobile client ↔ Windows OS.

## Core knowledge

### Architecture Overview
```
┌──────────────────┐     HTTP/REST      ┌──────────────────────┐
│  Flutter App     │ ◄──────────────► │  FastAPI Server       │
│  (Android/iOS)   │     LAN WiFi      │  (Windows PC)         │
│                  │                    │                       │
│  ┌─────────────┐ │                    │  ┌─────────────────┐ │
│  │ AudioState  │ │  POST /audio/mute  │  │ COM Context Mgr │ │
│  │ (Provider)  │ │ ──────────────────►│  │ ┌─────────────┐ │ │
│  │             │ │                    │  │ │ pycaw        │ │ │
│  │ Optimistic  │ │  {"success": true} │  │ │ (Core Audio) │ │ │
│  │ Update +    │ │ ◄──────────────────│  │ └─────────────┘ │ │
│  │ Rollback    │ │                    │  └─────────────────┘ │
│  └─────────────┘ │                    │                       │
└──────────────────┘                    └──────────────────────┘
```

### Data Flow: Toggle Mute (Complete Lifecycle)
```
1. User taps the mute button on the Flutter app.
   ↓
2. HapticFeedback.lightImpact() — tactile confirmation.
   ↓
3. AudioState saves originalMuted = current value.
   ↓
4. AudioState sets muted = !current (OPTIMISTIC — UI changes instantly).
   ↓
5. notifyListeners() — all listening widgets rebuild with the new state.
   ↓
6. ApiService.toggleDeviceMute(deviceId, targetMute) is invoked.
   ↓
7. HTTP POST /audio/device/mute is sent to the FastAPI server.
   ↓
8. FastAPI route handler (def, NOT async def):
   a. com_context() → CoInitialize()
   b. _get_device_volume(device_id) — find device via pycaw
   c. volume.SetMute(mute, None) — apply to Windows core audio
   d. CoUninitialize() is called to release COM
   e. Return {"success": true, "muted": true}
   ↓
9. Flutter receives the response:
   ├── SUCCESS → State is already correct, show success SnackBar.
   └── FAILURE →
       a. AudioState rolls back: muted = originalMuted
       b. notifyListeners() — UI reverts back
       c. HapticFeedback.heavyImpact() — error tactile vibration
       d. Show red SnackBar: "Gagal mengubah mute perangkat" (localized error)
```

### Key Files in This Pattern
| Layer | File | Purpose |
|-------|------|---------|
| Server | `server/routes/audio.py` | COM-safe route handlers with context manager |
| Server | `server/main.py` | FastAPI app, middleware, rate limiting |
| Client | `app/lib/models/audio_state.dart` | Centralized state with optimistic updates |
| Client | `app/lib/services/api_service.dart` | HTTP client with timeout, 401 handling |
| Client | `app/lib/widgets/audio_card.dart` | UI consuming AudioState provider |

### Server Pattern: COM Context Manager
```python
from contextlib import contextmanager
from comtypes import CoInitialize, CoUninitialize
from fastapi import HTTPException
import logging

logger = logging.getLogger(__name__)

@contextmanager
def com_context():
    """Thread-safe COM lifecycle — MUST wrap every pycaw call."""
    try:
        CoInitialize()
        yield
    finally:
        try:
            CoUninitialize()
        except Exception:
            pass

# CRITICAL: Use 'def' (NOT 'async def') for COM routes
# FastAPI runs 'def' handlers in thread pool, avoiding event loop blocking
@router.post("/device/mute")
def set_device_mute(request: DeviceMuteRequest):
    with com_context():
        try:
            volume = _get_device_volume(request.device_id)
            volume.SetMute(request.mute, None)
            return {"success": True, "device_id": request.device_id, "muted": request.mute}
        except HTTPException:
            raise
        except Exception as e:
            logger.error("Mute failed for %s: %s", request.device_id, e)
            raise HTTPException(status_code=503, detail=f"Failed to toggle mute: {e}")
```

### Client Pattern: Optimistic State with Rollback
```dart
class AudioState extends ChangeNotifier {
  List<dynamic> _devices = [];

  Future<bool> toggleDeviceMute(String deviceId, bool targetMute) async {
    final index = _devices.indexWhere((d) => d['id'] == deviceId);
    if (index == -1) return false;

    // Save for rollback
    final originalMuted = _devices[index]['muted'] as bool;

    // Optimistic update — UI changes INSTANTLY
    _devices[index]['muted'] = targetMute;
    notifyListeners();

    // Server call
    final success = await ApiService.toggleDeviceMute(deviceId, targetMute);

    // Rollback on failure
    if (!success) {
      _devices[index]['muted'] = originalMuted;
      notifyListeners();
      return false;
    }
    return true;
  }

  // Local-only update for smooth slider dragging (no API call)
  void updateDeviceVolumeLocally(String deviceId, double level) {
    final index = _devices.indexWhere((d) => d['id'] == deviceId);
    if (index != -1) {
      _devices[index]['volume'] = level.toInt();
      notifyListeners();
    }
  }
}
```

### Error Response Convention
| HTTP Code | Server Response | Flutter SnackBar (Indonesian) | Flutter SnackBar (English) |
|-----------|----------------|--------------------------------|----------------------------|
| 400 | `"Level volume harus antara 0-100"` | Message direct from detail | Message direct from detail |
| 404 | `"Perangkat audio tidak ditemukan"` | "Perangkat tidak ditemukan" | "Device not found" |
| 429 | `"Terlalu banyak percobaan"` | "Coba lagi dalam 1 menit" | "Too many attempts. Try again in 1 min." |
| 503 | `"Gagal mengakses audio service"` | "Server audio bermasalah" | "Audio service error" |

## Behavior rules
- Server routes accessing the Windows API MUST be synchronous (`def`, not `async def`).
- Mobile client MUST isolate audio state inside a `ChangeNotifierProvider`.
- Error messages sent to clients should be localized (e.g., Indonesian for Indonesian users) for consumption in UI elements like SnackBars.
- Slider `onChanged` updates state locally (no API call). `onChangeEnd` triggers the actual API call to avoid spamming the network.

## Invocation examples
1. "How do I handle pycaw in FastAPI to prevent blocking the client connection ping?"
2. "How do I write a Slider callback to prevent stuttering while dragging?"
3. "What is the FastAPI JSON response format for listing Windows audio outputs?"
4. "How do I implement rollback logic if the server goes down while a user toggles mute?"

## Output format
- **Architecture Diagram**: Data flow from client → server → OS → response.
- **Code Patterns**: Concrete Server (Python) and Client (Dart) implementations.
- **Error Pipeline**: Mappings of status codes → response message → UI behavior.

## Integration
Integrates patterns from `skills/backend` (RIKU), `skills/mobile` (MAYA), and `skills/windows-desktop` (VIKTOR) into a unified, production-tested use case.
