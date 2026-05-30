import webbrowser
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

router = APIRouter()

class BrowserRequest(BaseModel):
    url: str

@router.post("/open")
async def open_browser(request: BrowserRequest):
    """Opens a URL in the system's default browser."""
    url = request.url
    if not (url.startswith("http://") or url.startswith("https://")):
        raise HTTPException(status_code=400, detail="URL harus diawali http:// atau https://")

    try:
        webbrowser.open(url)
        return {"success": True, "url": url}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Gagal membuka browser: {e}")
