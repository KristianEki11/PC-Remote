# Audit Report: PC Remote (v4.0.0)

Mode: AUDIT (read-only, no product code changed). Scope: server (Go), Android app (Flutter), web client (Next.js), docs.
Not covered: `server/handlers/*.go`, `server/windows/*.go`, installer NSIS, CI workflows, Flutter screens/widgets, running tests.

## Ranked findings

### High
1. **Pairing ignores the QR fingerprint (MITM on LAN).** `pairWithQR` stores `server_fingerprint` but sends `/auth/pair` with `_createTrustingClient()`, which accepts any certificate (`app/lib/services/api_service.dart:314`, `:71-77`). A rogue host on the same network can capture the one-time pair token. Fix: use the pinned client for pairing when the fingerprint is known.
2. **Manual login is trust-on-first-use.** `login()` (`api_service.dart:441-494`) never stores a fingerprint, so the PIN goes to whatever certificate answers first (`:61-62`). Fix: store and pin the fingerprint, or require pairing first.
3. **Login lockout never triggers.** `LoginHandler` keys failures on `r.RemoteAddr`, which includes the ephemeral port (`server/auth/handlers.go:60`), while `WithAuth` checks lockout by `ExtractIP` (`server/middleware/auth.go:39`). Failures are recorded under keys that are never checked. The 5 req/min token bucket still limits brute force (roughly 33 min for a 4-digit PIN, far longer for 8 digits), but the documented 10-failure/15-min lockout does not work. Fix: use `ExtractIP(r)` in `LoginHandler` and `PairHandler`, and record pair failures too (`PairHandler` currently only logs).
4. **cloudflared is downloaded and executed with no integrity check.** `findOrDownloadBinary` fetches `releases/latest` from GitHub and runs it (`server/tunnel/tunnel.go:311-341`). Fix: pin a version and verify a SHA-256, or bundle the binary with the installer.

### Medium
5. **Bearer token is stored in plaintext.** `session.json` holds the 1-year token (`server/auth/session.go:241-262`). The 0600 mode is ignored on Windows. Android keeps `auth_token` in SharedPreferences (`api_service.dart:345`), and the web client keeps it in `localStorage` (`web/src/app/connect/page.tsx:98-99`), so any XSS on the web origin gives full PC control. Fix: DPAPI for the server file, Keystore for Android (`flutter_secure_storage`), shorter TTL with refresh.
6. **Token lifetime is long and never rotates.** `DefaultTokenTTL` is 365 days (`session.go:18-19`). Consider 30 days with sliding renewal.
7. **Legacy `X-PIN` auth is still active** (`middleware/auth.go:58-91`). The PIN travels on every request, and the Android legacy probe sends it too (`api_service.dart:477-489`). Remove it after a deprecation window.
8. **Web client accepts `http://` and sends the token over it** (`page.tsx:68`). Reject non-HTTPS URLs except `localhost`, or warn.
9. **Public exposure contradicts the README.** README says "server rejects requests from outside local subnets" (`README.md:182`). The Cloudflare tunnel exposes `/auth/login`, `/auth/pair`, and all protected routes publicly. Only `/internal/*` is protected (by `Cf-Ray` detection, `middleware/ratelimit.go:184-194`, which works). Correct the README and state the public surface explicitly.

### Low / design and code
10. **Docs drift.** README says tokens expire after 24 h (`README.md:184`); code uses 1 year. README says constant-time comparison is used for tokens; `ValidateToken` uses `!=` (`session.go:124`).
11. **Web UI is not connected to state.** `isPlaying` is local and never syncs with the GSMTC status (`page.tsx:32`, `:354-357`). Play/pause shows the wrong icon after changes outside the app.
12. **Web swallows command errors.** `callBackend` drops failures (`page.tsx:183-185`) and the UI shows success. Show an error state and retry.
13. **Web volume slider sends only on mouse/touch release** (`page.tsx:328-335`). Keyboard changes are never sent. Use `onChange` with throttling, or `onKeyUp`.
14. **Web a11y.** Icon-only media and power buttons have no `aria-label` (`page.tsx:347-367`, `:375-398`). The 429 lockout response is shown as "Invalid PIN" (`page.tsx:93`). Use `alert()` replacements (toast).
15. **Flutter routing is by hostname heuristic.** `_getClient` checks `url.contains('.com')` etc. (`api_service.dart:82`). Choose the client from the stored configuration (LAN vs tunnel), not string matching.
16. **Duplicated failover logic.** `_get` and `_post` repeat the same two-route code (`api_service.dart:206-296`). Extract one `_request` helper. The server also duplicates `ErrorResponse` and `authSendJSON`.
17. **No `ReadHeaderTimeout`** on the HTTP server (`server/main.go:178-184`). Add it.
18. **Rate-limiter map grows** with each new IP between 10-minute cleanups (`ratelimit.go:209-224`). Acceptable for a LAN tool; note it.
19. **Quick Tunnel URL changes on every restart**, so the phone's `public_url` goes stale until re-pairing. Document the named-tunnel option (`TunnelToken`/`TunnelURL`).

## Suggested order
1. Fix findings 1-3 together (same pairing and lockout code, small diff).
2. Finding 4 (pin cloudflared version and hash).
3. Findings 5-9 (token storage, legacy auth, README correction).
4. Web UX fixes 11-14, then code cleanup 15-18.

## Verdict
Not safe for use on untrusted networks in its current state (findings 1-2). Acceptable for a home LAN with the tunnel off. Findings 3 and 9 need correcting before anyone relies on the documented behaviour.

## Outcomes (fixes applied, not committed)

| # | Finding | Status | Where |
|---|---------|--------|-------|
| 1 | Pairing ignores QR fingerprint | Fixed: LAN pairing pins the QR fingerprint; tunnel uses system trust | app/lib/services/api_service.dart (pairWithQR) |
| 2 | Manual login is trust-on-first-use | Fixed: no stored fingerprint = reject; login shows a "scan QR first" message | api_service.dart (_initHttpClient, login) |
| 3 | Login lockout never triggers | Fixed: ClientIP shared by lockout and login/pair handlers; pair failures recorded | server/auth/clientip.go, server/auth/handlers.go, server/middleware/ratelimit.go |
| 4 | cloudflared downloaded unverified | Fixed: silent download removed; clear error if the binary is missing | server/tunnel/tunnel.go |

Verification:
- `go build ./...`: pass.
- `go vet ./auth ./middleware ./tunnel`: pass. `go vet ./...` reports 3 pre-existing `unsafe.Pointer` warnings in server/windows/api.go (untouched file).
- `go test ./auth`: pass (includes new TestClientIP).
- `go test ./middleware`: TestIsLocalhost fails on two subcases (X-Forwarded-For, External Host Header). It fails identically on the untouched HEAD, so it is pre-existing and not fixed here.
- Dart: NOT compiled (Flutter not installed). Verified by reading only.

Behavior change to know: manual IP+PIN login now requires a prior QR pairing (stored fingerprint).
