from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from ctypes import cast, POINTER
from comtypes import CLSCTX_ALL, CoInitialize, CoUninitialize
import comtypes.client
from pycaw.pycaw import AudioUtilities, IAudioEndpointVolume, IMMDeviceEnumerator
import threading
import queue
import logging

router = APIRouter()
logger = logging.getLogger("pc_remote.audio")

# ──────────────────────────────────────
# Request Models
# ──────────────────────────────────────
class VolumeRequest(BaseModel):
    level: int

class DeviceVolumeRequest(BaseModel):
    device_id: str
    level: int

class DeviceMuteRequest(BaseModel):
    device_id: str
    mute: bool

# ──────────────────────────────────────
# Thread-safe COM Task Queue & Background Worker
# ──────────────────────────────────────
class AudioTask:
    def __init__(self, func, *args, **kwargs):
        self.func = func
        self.args = args
        self.kwargs = kwargs
        self.result_queue = queue.Queue()

    def run(self):
        try:
            res = self.func(*self.args, **self.kwargs)
            self.result_queue.put((True, res))
        except Exception as e:
            self.result_queue.put((False, e))

_audio_queue = queue.Queue()
_worker_should_run = True

def _audio_worker_loop():
    """Background worker loop that manages COM apartment lifecycle."""
    global _worker_should_run
    CoInitialize()
    logger.info("COM Audio worker thread initialized.")
    try:
        while _worker_should_run:
            try:
                task = _audio_queue.get(timeout=1.0)
            except queue.Empty:
                continue
                
            if task is None:
                logger.info("Poison pill received. Stopping audio worker thread loop.")
                break
                
            task.run()
            _audio_queue.task_done()
    except Exception as e:
        logger.error(f"Error in audio worker thread loop: {e}")
    finally:
        CoUninitialize()
        logger.info("COM Audio worker thread uninitialized and closed.")

# Start worker thread
_audio_worker_thread = threading.Thread(target=_audio_worker_loop, name="PCRemoteAudioWorker", daemon=True)
_audio_worker_thread.start()

def run_in_audio_thread(func, *args, **kwargs):
    """Executes a function inside the COM background thread and waits for its result."""
    if not _audio_worker_thread.is_alive():
        raise HTTPException(status_code=503, detail="Audio worker thread is dead.")
        
    task = AudioTask(func, *args, **kwargs)
    _audio_queue.put(task)
    success, res = task.result_queue.get()
    if success:
        return res
    else:
        if isinstance(res, HTTPException):
            raise res
        raise HTTPException(status_code=500, detail=str(res))

def stop_audio_worker():
    """Stops the audio worker loop and waits for the thread to terminate."""
    global _worker_should_run
    _worker_should_run = False
    _audio_queue.put(None)
    _audio_worker_thread.join(timeout=3.0)

# ──────────────────────────────────────
# Core COM Operations (Executed in COM worker thread)
# ──────────────────────────────────────
ENUMERATOR_CLSID = "{BCDE0395-E52F-467C-8E3D-C4579291692E}"

def _get_render_endpoints():
    """Return all active audio output devices via IMMDeviceEnumerator."""
    enumerator = comtypes.client.CreateObject(ENUMERATOR_CLSID, interface=IMMDeviceEnumerator)
    # eRender=0, DEVICE_STATE_ACTIVE=1
    return enumerator.EnumAudioEndpoints(0, 1)

def _get_endpoint_volume(device) -> IAudioEndpointVolume:
    """Activate and return IAudioEndpointVolume from a COM device object."""
    interface = device.Activate(IAudioEndpointVolume._iid_, CLSCTX_ALL, None)
    return cast(interface, POINTER(IAudioEndpointVolume))

def _get_master_volume():
    """Get the default Speakers' IAudioEndpointVolume interface."""
    speakers = AudioUtilities.GetSpeakers()
    if not speakers or not speakers._dev:
         raise HTTPException(status_code=404, detail="Default audio output device not found.")
    return _get_endpoint_volume(speakers._dev)

def _get_device_volume(device_id: str):
    """Find and return IAudioEndpointVolume for a specific device by ID."""
    for dev in AudioUtilities.GetAllDevices():
        if dev.id == device_id:
            return _get_endpoint_volume(dev._dev)
    raise HTTPException(status_code=404, detail="Perangkat audio tidak ditemukan")

def _validate_level(level: int):
    if not (0 <= level <= 100):
        raise HTTPException(status_code=400, detail="Level volume harus antara 0-100")

# ──────────────────────────────────────
# API Endpoint Handler Callbacks
# ──────────────────────────────────────
def _get_audio_devices_impl():
    collection = _get_render_endpoints()
    render_ids = {collection.Item(i).GetId().lower() for i in range(collection.GetCount())}

    result = []
    for dev in AudioUtilities.GetAllDevices():
        if dev.id.lower() not in render_ids:
            continue
        try:
            volume = _get_endpoint_volume(dev._dev)
            result.append({
                "id": dev.id,
                "name": dev.FriendlyName,
                "volume": round(volume.GetMasterVolumeLevelScalar() * 100),
                "muted": bool(volume.GetMute()),
            })
        except Exception:
            pass
    return result

@router.get("/devices")
def get_audio_devices():
    """List all active audio output devices."""
    try:
        devices = run_in_audio_thread(_get_audio_devices_impl)
        return {"devices": devices}
    except HTTPException as he:
        raise he
    except Exception as e:
        raise HTTPException(
            status_code=503,
            detail=f"Gagal mengambil daftar perangkat audio: {e}"
        )

def _set_device_volume_impl(device_id: str, level: int):
    _validate_level(level)
    volume = _get_device_volume(device_id)
    volume.SetMasterVolumeLevelScalar(level / 100.0, None)
    return {"success": True, "device_id": device_id, "level": level}

@router.post("/device/volume")
def set_device_volume(request: DeviceVolumeRequest):
    """Set volume level for a specific audio device."""
    try:
        return run_in_audio_thread(_set_device_volume_impl, request.device_id, request.level)
    except HTTPException as he:
        raise he
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Gagal mengubah volume perangkat: {e}"
        )

def _set_device_mute_impl(device_id: str, mute: bool):
    volume = _get_device_volume(device_id)
    volume.SetMute(mute, None)
    return {"success": True, "device_id": device_id, "muted": mute}

@router.post("/device/mute")
def set_device_mute(request: DeviceMuteRequest):
    """Mute or unmute a specific audio device."""
    try:
        return run_in_audio_thread(_set_device_mute_impl, request.device_id, request.mute)
    except HTTPException as he:
        raise he
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Gagal mengubah status mute perangkat: {e}"
        )

def _get_volume_impl():
    volume = _get_master_volume()
    level = round(volume.GetMasterVolumeLevelScalar() * 100)

    sessions = AudioUtilities.GetAllSessions()
    if not sessions:
        muted = bool(volume.GetMute())
    else:
        muted = all(
            (vol := s.SimpleAudioVolume) and vol.GetMute()
            for s in sessions if s.SimpleAudioVolume
        )
    return {"level": level, "muted": muted}

@router.get("/volume")
def get_volume():
    """Get current master volume level and mute status."""
    try:
        return run_in_audio_thread(_get_volume_impl)
    except HTTPException as he:
        raise he
    except Exception as e:
        raise HTTPException(
            status_code=503,
            detail=f"Gagal mengambil volume master: {e}"
        )

def _set_volume_impl(level: int):
    _validate_level(level)
    scalar = level / 100.0

    # 1. Default speakers
    try:
        _get_master_volume().SetMasterVolumeLevelScalar(scalar, None)
    except Exception:
        pass

    # 2. All render endpoints
    try:
        collection = _get_render_endpoints()
        for i in range(collection.GetCount()):
            try:
                _get_endpoint_volume(collection.Item(i)).SetMasterVolumeLevelScalar(scalar, None)
            except Exception:
                pass
    except Exception:
        pass

    # 3. All audio application sessions
    for session in AudioUtilities.GetAllSessions():
        if vol := session.SimpleAudioVolume:
            try:
                vol.SetMasterVolume(scalar, None)
            except Exception:
                pass

    return {"success": True, "level": level}

@router.post("/volume")
def set_volume(request: VolumeRequest):
    """Set master volume across all devices and application sessions."""
    try:
        return run_in_audio_thread(_set_volume_impl, request.level)
    except HTTPException as he:
        raise he
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Gagal mengubah volume: {e}"
        )

def _toggle_mute_impl():
    sessions = AudioUtilities.GetAllSessions()

    if not sessions:
        volume = _get_master_volume()
        new_mute = not volume.GetMute()
    else:
        new_mute = any(
            (vol := s.SimpleAudioVolume) and not vol.GetMute()
            for s in sessions if s.SimpleAudioVolume
        )

    # Apply to all sessions
    for session in sessions:
        if vol := session.SimpleAudioVolume:
            try:
                vol.SetMute(new_mute, None)
            except Exception:
                pass

    # Apply to all endpoints
    try:
        collection = _get_render_endpoints()
        for i in range(collection.GetCount()):
            try:
                _get_endpoint_volume(collection.Item(i)).SetMute(new_mute, None)
            except Exception:
                pass
    except Exception:
        pass

    return {"success": True, "muted": new_mute}

@router.post("/mute")
def toggle_mute():
    """Toggle mute/unmute globally."""
    try:
        return run_in_audio_thread(_toggle_mute_impl)
    except HTTPException as he:
        raise he
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Gagal melakukan toggle mute: {e}"
        )
