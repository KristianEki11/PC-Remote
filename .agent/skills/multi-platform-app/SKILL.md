# Example Pattern: Multi-Platform App (Web + Mobile + Windows Desktop)

## Description
An architectural pattern for multi-platform systems encompassing a Web client (Next.js), a Mobile client (Flutter), and a Windows Desktop client (Tauri/C#), utilizing a centralized API Gateway and shared type contracts. Includes cross-platform authentication strategies, state synchronization, and deployment workflows.

## When to trigger
- App architectures that must run on web, mobile, and desktop concurrently
- Code sharing and API contract sharing across different platforms
- Unified single sign-on (SSO) secure authentication across web sandbox, mobile, and desktop environments
- Real-time data synchronization across multiple platforms
- Keywords: "multi-platform", "cross-platform", "unified API", "code sharing", "SSO", "Tauri Flutter Next.js"

## Agent persona
- **Name**: CO-DESIGNED BY SERA, MAYA & VIKTOR
- **Domain**: Cross-platform architecture, delivery optimization, unified API design.

## Core knowledge

### Architecture Overview
```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│  Next.js    │  │  Flutter    │  │  Tauri       │
│  (Web)      │  │  (Mobile)   │  │  (Desktop)   │
│  Port 3000  │  │  Android/iOS│  │  Windows     │
└──────┬──────┘  └──────┬──────┘  └──────┬──────┘
       │                │                │
       └────────┬───────┴────────┬───────┘
                │  HTTP/REST     │
                ▼                ▼
        ┌───────────────────────────┐
        │  API Gateway (FastAPI)    │
        │  - Auth (JWT)             │
        │  - Rate Limiting          │
        │  - CORS per-origin        │
        ├───────────────────────────┤
        │  PostgreSQL / SQLite      │
        │  Redis (cache + pub/sub)  │
        └───────────────────────────┘
```

### Token Storage per Platform
| Platform | Storage Method | Security Level |
|----------|---------------|----------------|
| **Web (Next.js)** | HttpOnly cookie + CSRF token | High — not accessible via client JavaScript |
| **Mobile (Flutter)** | `flutter_secure_storage` (Keychain/Keystore) | High — hardware-backed storage |
| **Desktop (Tauri)** | OS keyring via `tauri-plugin-store` | High — OS-managed credentials |
| ❌ **AVOID** | localStorage, SharedPreferences for tokens | Low — plaintext storage |

### API Contract Sharing Strategy
```
Single Source of Truth: OpenAPI schema generated from FastAPI server
                                 │
        ┌────────────────────────┼────────────────────────┐
        ▼                        ▼                        ▼
  TypeScript types         Dart classes             Rust types
  (openapi-ts)             (openapi-gen)            (openapi-gen)
        │                        │                        │
  Next.js client           Flutter client           Tauri client
```

### CORS Configuration per Client
```python
# FastAPI CORS — allow all 3 client origins
app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "http://localhost:3000",     # Next.js dev
        "https://app.example.com",   # Next.js production
        "tauri://localhost",         # Tauri desktop
    ],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
# Note: Flutter mobile doesn't need CORS (no browser sandbox restrictions)
```

### State Synchronization Patterns
| Pattern | When to Use | Latency |
|---------|-------------|---------|
| **Polling** (GET every N sec) | Dashboard status, slowly changing data | 1-10s |
| **SSE** (Server-Sent Events) | Unidirectional broadcasts (server → client) | Real-time |
| **WebSocket** | Bidirectional real-time sync (chat, collaborative editing) | Real-time |
| **Push Notifications** (FCM/APNs) | Mobile background updates, offline users | Variable |

### Platform-Specific Build & Deploy
| Platform | Build Tool | Distribution | Auto-Update |
|----------|-----------|--------------|-------------|
| **Web** | `next build` | Vercel / Docker | Instant (server-side updates) |
| **Android** | `flutter build apk` | Google Play / APK sideload | Play Store auto-updates |
| **iOS** | `flutter build ipa` | App Store / TestFlight | App Store auto-updates |
| **Windows** | `cargo tauri build` | NSIS installer / MSIX | Custom updater / Microsoft Store |

## Behavior rules

### MANDATORY (WAJIB):
1. The API contract MUST be defined in a single source of truth (OpenAPI/tRPC schema) and auto-generated for all platform clients.
2. Token storage MUST use platform-safe mechanisms (see table above) — NEVER use localStorage/SharedPreferences for secrets.
3. CORS MUST be configured explicitly per-origin in production — NEVER use `allow_origins=["*"]`.
4. Each platform client MUST have an independent build and deployment pipeline — deployment failures on one platform should not block others.

### FORBIDDEN (DILARANG):
1. DO NOT hardcode API keys or secrets in any client-side code — use environment variables and server-side proxies.
2. DO NOT compromise the security constraints of one platform for development convenience on another.
3. DO NOT assume all platforms share the same UI lifecycle or hardware capabilities — mobile platforms have different background constraints compared to Web.

## Invocation examples
1. "How do I share API schemas between a Next.js web application and a Flutter mobile app?"
2. "How do I design an SSO flow that functions correctly across browser sandboxes, Android, and Windows desktop?"
3. "What polling interval is optimal for a multi-platform dashboard?"
4. "How do I configure CORS on the backend so both Web and Tauri desktop apps can access the same API?"

## Output format
- **Architecture Diagram**: Relationships between client → gateway → database.
- **API Contract Schema**: Type-safe shared data transfer format.
- **Security Matrix**: Token storage and auth flows mapped per-platform.
- **Deployment Matrix**: Build tools, distribution channels, and auto-update mechanisms per-platform.

## Integration
Combines expertise from `skills/fullstack-web` (SERA — web client), `skills/mobile` (MAYA — Flutter client), and `skills/windows-desktop` (VIKTOR — Tauri/desktop client) into a cohesive architecture.
