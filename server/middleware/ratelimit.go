package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// ── Rate Limiter ─────────────────────────────────────────────
// Token-bucket rate limiter per IP address.
// Two instances are used:
//   - authLimiter:    tight limits for /auth/* (5 req/min, lockout after 10 failures)
//   - generalLimiter: standard limits for all other endpoints (60 req/min)

// RateLimiter implements a per-IP token bucket algorithm with brute-force lockout.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor

	// Configuration
	requestsPerMinute float64
	burstSize         int
	authFailureLimit  int
	lockoutDuration   time.Duration
}

type visitor struct {
	tokens       float64
	maxTokens    float64
	refillRate   float64 // tokens per second
	lastCheck    time.Time
	authFailures int
	lockedUntil  time.Time
}

// NewRateLimiter creates a rate limiter for general API endpoints.
// 60 requests/minute with burst of 10.
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		visitors:          make(map[string]*visitor),
		requestsPerMinute: 60,
		burstSize:         10,
		authFailureLimit:  10,
		lockoutDuration:   15 * time.Minute,
	}
	go rl.cleanupLoop()
	return rl
}

// NewAuthRateLimiter creates a stricter rate limiter for authentication endpoints.
// 5 requests/minute with lockout after 10 consecutive failures.
func NewAuthRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		visitors:          make(map[string]*visitor),
		requestsPerMinute: 5,
		burstSize:         5,
		authFailureLimit:  10,
		lockoutDuration:   15 * time.Minute,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		refillRate := rl.requestsPerMinute / 60.0
		v = &visitor{
			tokens:     float64(rl.burstSize),
			maxTokens:  float64(rl.burstSize),
			refillRate: refillRate,
			lastCheck:  time.Now(),
		}
		rl.visitors[ip] = v
	}
	return v
}

// Allow checks whether the given IP has remaining request tokens.
// Also checks for brute-force lockout.
func (rl *RateLimiter) Allow(ip string) bool {
	v := rl.getVisitor(ip)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Check lockout from too many auth failures
	if now.Before(v.lockedUntil) {
		return false
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(v.lastCheck).Seconds()
	v.tokens += elapsed * v.refillRate
	if v.tokens > v.maxTokens {
		v.tokens = v.maxTokens
	}
	v.lastCheck = now

	if v.tokens < 1 {
		return false
	}

	v.tokens--
	return true
}

// RecordAuthFailure increments the failure counter for an IP.
// After authFailureLimit consecutive failures, the IP is locked out.
func (rl *RateLimiter) RecordAuthFailure(ip string) {
	v := rl.getVisitor(ip)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	v.authFailures++
	if v.authFailures >= rl.authFailureLimit {
		v.lockedUntil = time.Now().Add(rl.lockoutDuration)
		v.authFailures = 0
	}
}

// ResetAuthFailures clears the failure counter after a successful login.
func (rl *RateLimiter) ResetAuthFailures(ip string) {
	v := rl.getVisitor(ip)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	v.authFailures = 0
}

// IsLocked returns true if the IP is currently in lockout from too many failures.
func (rl *RateLimiter) IsLocked(ip string) bool {
	v := rl.getVisitor(ip)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	return time.Now().Before(v.lockedUntil)
}

// Middleware returns an http.Handler that enforces rate limiting on all requests.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := ExtractIP(r)

		if !rl.Allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"too many requests"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ExtractIP extracts the client IP from the request.
// Does NOT trust X-Forwarded-For since this server handles direct connections.
func ExtractIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// IsLocalhost returns true if the request comes from a loopback address.
// Used to protect internal-only endpoints like /internal/qr.
func IsLocalhost(r *http.Request) bool {
	ip := ExtractIP(r)
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return parsed.IsLoopback()
}

// LocalhostOnly is a middleware that rejects non-localhost requests with 403 Forbidden.
func LocalhostOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsLocalhost(r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"localhost only"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			// Remove visitors idle for more than 10 minutes and not locked out
			if now.Sub(v.lastCheck) > 10*time.Minute && now.After(v.lockedUntil) {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}
