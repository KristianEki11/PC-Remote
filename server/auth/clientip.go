package auth

import (
	"net"
	"net/http"
)

// ClientIP extracts the client IP from the request.
// Does NOT trust X-Forwarded-For since this server handles direct connections.
// CF-Connecting-IP is only trusted when the request arrives via loopback (cloudflared tunnel).
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	if host == "127.0.0.1" || host == "::1" {
		if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
			return cfIP
		}
	}
	return host
}
