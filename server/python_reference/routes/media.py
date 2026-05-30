from fastapi import APIRouter
import ctypes

router = APIRouter()

# Windows Virtual-Key Codes for Media Control
VK_MEDIA_NEXT_TRACK = 0xB0
VK_MEDIA_PREV_TRACK = 0xB1
VK_MEDIA_PLAY_PAUSE = 0xB3

KEYEVENTF_KEYUP = 0x0002

def _send_media_key(key_code: int):
    """Sends a hardware-level keystroke to Windows user32 library."""
    # Key down
    ctypes.windll.user32.keybd_event(key_code, 0, 0, 0)
    # Key up
    ctypes.windll.user32.keybd_event(key_code, 0, KEYEVENTF_KEYUP, 0)

@router.post("/playpause")
async def media_playpause():
    """Simulates media Play/Pause keystroke."""
    _send_media_key(VK_MEDIA_PLAY_PAUSE)
    return {"success": True}

@router.post("/next")
async def media_next():
    """Simulates media Next Track keystroke."""
    _send_media_key(VK_MEDIA_NEXT_TRACK)
    return {"success": True}

@router.post("/prev")
async def media_prev():
    """Simulates media Previous Track keystroke."""
    _send_media_key(VK_MEDIA_PREV_TRACK)
    return {"success": True}
