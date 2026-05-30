import os
import secrets
import time
import logging
from typing import Dict
from fastapi import APIRouter, HTTPException, Security, Request
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from pydantic import BaseModel
from dotenv import load_dotenv
from slowapi import Limiter
from slowapi.util import get_remote_address
from slowapi.errors import RateLimitExceeded
from fastapi.responses import JSONResponse

load_dotenv()

logger = logging.getLogger("pc_remote.auth")

APP_PIN = os.getenv("APP_PIN", "1234")

# Token storage in memory
# Format: { token: timestamp }
active_tokens: Dict[str, float] = {}
TOKEN_EXPIRY = 24 * 60 * 60  # 24 hours in seconds

# ──────────────────────────────────────
# Rate Limiter (5 attempts per minute on /auth)
# ──────────────────────────────────────
limiter = Limiter(key_func=get_remote_address)

auth_router = APIRouter()
security = HTTPBearer()

class LoginRequest(BaseModel):
    pin: str

def get_client_ip(request: Request) -> str:
    """Helper to retrieve client IP address, checking forwarding headers first."""
    if request.headers.get("X-Forwarded-For"):
        return request.headers.get("X-Forwarded-For").split(",")[0].strip()
    return request.client.host if request.client else ""

def is_local_ip(ip: str) -> bool:
    """Checks if an IP address is part of the local machine or private network subnets."""
    if ip in ["127.0.0.1", "::1", "localhost", "testclient"]:
        return True
    
    # Private network checks (IPv4)
    # Class A: 10.0.0.0 to 10.255.255.255
    if ip.startswith("10."):
        return True
    # Class B: 172.16.0.0 to 172.31.255.255
    if ip.startswith("172."):
        try:
            parts = ip.split(".")
            second_octet = int(parts[1])
            if 16 <= second_octet <= 31:
                return True
        except (IndexError, ValueError):
            pass
    # Class C: 192.168.0.0 to 192.168.255.255
    if ip.startswith("192.168."):
        return True
        
    return False

def cleanup_tokens() -> int:
    """Remove expired tokens from active storage. Returns count of removed tokens."""
    current_time = time.time()
    expired = [t for t, ts in active_tokens.items() if current_time - ts > TOKEN_EXPIRY]
    for t in expired:
        del active_tokens[t]
    if expired:
        logger.info("Cleaned up %d expired token(s)", len(expired))
    return len(expired)

def verify_token(credentials: HTTPAuthorizationCredentials = Security(security)) -> str:
    """Verify that a bearer token is currently active and not expired."""
    token = credentials.credentials
    if token not in active_tokens:
        raise HTTPException(status_code=401, detail="Token invalid or expired")
    
    # Check expiry
    if time.time() - active_tokens[token] > TOKEN_EXPIRY:
        del active_tokens[token]
        raise HTTPException(status_code=401, detail="Token expired")
    
    return token

async def require_auth(request: Request, credentials: HTTPAuthorizationCredentials = Security(security)) -> str:
    """FastAPI dependency to enforce authentication and local network constraints."""
    client_ip = get_client_ip(request)
    if not is_local_ip(client_ip):
        raise HTTPException(status_code=403, detail="Access denied. Local network only.")
    
    return verify_token(credentials)

def rate_limit_exceeded_handler(request: Request, exc: RateLimitExceeded):
    """Custom exception handler for rate-limited requests."""
    client_ip = get_client_ip(request)
    logger.warning("Rate limit exceeded for IP %s on %s", client_ip, request.url.path)
    return JSONResponse(
        status_code=429,
        content={"detail": "Terlalu banyak percobaan. Coba lagi dalam 1 menit."},
    )

@auth_router.post("/auth")
@limiter.limit("5/minute")
async def login(request: Request, data: LoginRequest):
    """Authenticate client PIN and return a bearer token."""
    client_ip = get_client_ip(request)
    if not is_local_ip(client_ip):
         raise HTTPException(status_code=403, detail="Access denied. Local network only.")

    if data.pin != APP_PIN:
        logger.warning("Failed login attempt from %s", client_ip)
        raise HTTPException(status_code=401, detail="PIN salah")
    
    cleanup_tokens()
    
    token = secrets.token_hex(16)
    active_tokens[token] = time.time()
    
    logger.info("Successful login from %s", client_ip)
    return {"token": token, "success": True}
