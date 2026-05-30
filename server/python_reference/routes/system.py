import subprocess
from fastapi import APIRouter, HTTPException

router = APIRouter()

@router.post("/lock")
async def lock_pc():
    """Lock the Windows workstation."""
    try:
        subprocess.run(["rundll32.exe", "user32.dll,LockWorkStation"], shell=False, check=True)
        return {"success": True}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Gagal mengunci PC: {str(e)}")

@router.post("/sleep")
async def sleep_pc():
    """Put the Windows PC to sleep."""
    try:
        subprocess.run(["rundll32.exe", "powrprof.dll,SetSuspendState", "0,1,0"], shell=False, check=True)
        return {"success": True}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Gagal sleep PC: {str(e)}")

@router.post("/restart")
async def restart_pc():
    """Restart the Windows system after a 5-second delay."""
    try:
        subprocess.run(["shutdown", "/r", "/t", "5"], shell=False, check=True)
        return {"success": True, "message": "PC akan restart dalam 5 detik"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Gagal restart PC: {str(e)}")

@router.post("/shutdown")
async def shutdown_pc():
    """Shutdown the Windows system after a 5-second delay."""
    try:
        subprocess.run(["shutdown", "/s", "/t", "5"], shell=False, check=True)
        return {"success": True, "message": "PC akan shutdown dalam 5 detik"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Gagal shutdown PC: {str(e)}")
