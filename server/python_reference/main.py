import os
import time
import logging
import asyncio
from logging.handlers import RotatingFileHandler
from contextlib import asynccontextmanager
from fastapi import FastAPI, Depends, Request
from fastapi.responses import JSONResponse
from dotenv import load_dotenv
import uvicorn

from auth import auth_router, require_auth, get_client_ip, is_local_ip, cleanup_tokens, limiter, rate_limit_exceeded_handler
from slowapi.errors import RateLimitExceeded
from routes.audio import router as audio_router, stop_audio_worker
from routes.media import router as media_router
from routes.browser import router as browser_router
from routes.system import router as system_router

load_dotenv()

# ──────────────────────────────────────
# Structured Logging (Saved in the executable's directory)
# ──────────────────────────────────────
LOG_DIR = os.path.join(os.getcwd(), "logs")
os.makedirs(LOG_DIR, exist_ok=True)

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
    handlers=[
        RotatingFileHandler(
            os.path.join(LOG_DIR, "server.log"),
            maxBytes=5 * 1024 * 1024,  # 5 MB per file
            backupCount=3,
            encoding="utf-8",
        ),
        logging.StreamHandler(),
    ],
)
logger = logging.getLogger("pc_remote")

# ──────────────────────────────────────
# Startup / Shutdown Lifespan
# ──────────────────────────────────────
SERVER_START_TIME = time.time()

async def _token_cleanup_loop():
    """Periodic background task: clean expired tokens every hour."""
    try:
        while True:
            await asyncio.sleep(3600)  # 1 hour
            count = cleanup_tokens()
            if count:
                logger.info("Periodic cleanup: removed %d expired token(s)", count)
    except asyncio.CancelledError:
        pass

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan: manage token cleaner task and COM background thread."""
    logger.info("PC Remote Server starting up...")
    task = asyncio.create_task(_token_cleanup_loop())
    yield
    # Cleanup on shutdown
    task.cancel()
    try:
        await task
    except Exception:
        pass
    
    # Clean terminate the COM worker thread in routes.audio
    logger.info("Stopping COM audio worker thread...")
    stop_audio_worker()
    
    logger.info("PC Remote Server shutting down.")

# ──────────────────────────────────────
# FastAPI App Initialization
# ──────────────────────────────────────
app = FastAPI(title="PC Remote Controller API", lifespan=lifespan)
app.state.limiter = limiter
app.add_exception_handler(RateLimitExceeded, rate_limit_exceeded_handler)

@app.middleware("http")
async def local_network_middleware(request: Request, call_next):
    """Middleware to enforce that all HTTP requests must originate from the local subnet."""
    client_ip = get_client_ip(request)
    if not is_local_ip(client_ip):
        logger.warning("Blocked non-local request from %s", client_ip)
        return JSONResponse(status_code=403, content={"detail": "Access denied. Local network only."})
    response = await call_next(request)
    return response

# ──────────────────────────────────────
# Health Check Endpoint
# ──────────────────────────────────────
@app.get("/health", tags=["Health"])
async def health_check():
    """Lightweight health check endpoint for connection ping."""
    uptime_seconds = round(time.time() - SERVER_START_TIME)
    return {"status": "ok", "uptime": uptime_seconds}

# ──────────────────────────────────────
# Modular Routers
# ──────────────────────────────────────
app.include_router(auth_router, tags=["Authentication"])
app.include_router(audio_router, prefix="/audio", tags=["Audio"], dependencies=[Depends(require_auth)])
app.include_router(media_router, prefix="/media", tags=["Media"], dependencies=[Depends(require_auth)])
app.include_router(browser_router, prefix="/browser", tags=["Browser"], dependencies=[Depends(require_auth)])
app.include_router(system_router, prefix="/system", tags=["System"], dependencies=[Depends(require_auth)])

if __name__ == "__main__":
    port = int(os.getenv("APP_PORT", 8000))
    # Direct script execution uses signal handlers for easy CTRL+C handling
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=True)
