package middleware

import (
	"net"
	"net/http"
)

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS).
// Restricts allowed origins to local network subnets and explicitly paired origins.
// Rejects wildcard origin (*) to prevent CSRF attacks from arbitrary websites.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin != "" && isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-PIN, X-Device-Name")
			w.Header().Set("Access-Control-Max-Age", "3600")
			w.Header().Set("Vary", "Origin")
		}

		// Handle preflight request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// isAllowedOrigin checks if the origin is from a trusted source:
// - Localhost (development / QR page)
// - Private network subnets (RFC 1918: 192.168.x.x, 10.x.x.x, 172.16-31.x.x)
// - Flutter web deployed on GitHub Pages (for PWA support)
func isAllowedOrigin(origin string) bool {
	// Always allow localhost origins (any port)
	if isLocalhostOrigin(origin) {
		return true
	}

	// Allow private network IP origins
	if isPrivateNetworkOrigin(origin) {
		return true
	}

	// Allow known deployment origins (Flutter Web PWA on GitHub Pages)
	allowedDomains := []string{
		"https://kristianeki11.github.io",
	}
	for _, domain := range allowedDomains {
		if origin == domain {
			return true
		}
	}

	return false
}

// isLocalhostOrigin checks if the origin is a localhost address.
func isLocalhostOrigin(origin string) bool {
	localhostPrefixes := []string{
		"http://localhost",
		"https://localhost",
		"http://127.0.0.1",
		"https://127.0.0.1",
		"http://[::1]",
		"https://[::1]",
	}
	for _, prefix := range localhostPrefixes {
		if len(origin) >= len(prefix) && origin[:len(prefix)] == prefix {
			// Must be exactly the prefix or followed by : (port) or / (path)
			rest := origin[len(prefix):]
			if rest == "" || rest[0] == ':' || rest[0] == '/' {
				return true
			}
		}
	}
	return false
}

// isPrivateNetworkOrigin checks if the origin IP belongs to an RFC 1918 private network.
func isPrivateNetworkOrigin(origin string) bool {
	// Extract host from origin URL
	host := extractHostFromOrigin(origin)
	if host == "" {
		return false
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	// RFC 1918 private ranges
	privateRanges := []struct {
		network string
		mask    string
	}{
		{"10.0.0.0", "255.0.0.0"},      // 10.0.0.0/8
		{"172.16.0.0", "255.240.0.0"},  // 172.16.0.0/12
		{"192.168.0.0", "255.255.0.0"}, // 192.168.0.0/16
	}

	for _, r := range privateRanges {
		network := net.ParseIP(r.network)
		mask := net.IPMask(net.ParseIP(r.mask).To4())
		if network != nil && mask != nil {
			subnet := &net.IPNet{IP: network, Mask: mask}
			if subnet.Contains(ip) {
				return true
			}
		}
	}

	return false
}

// extractHostFromOrigin parses the host portion from an origin URL like "http://192.168.1.100:8000".
func extractHostFromOrigin(origin string) string {
	// Remove scheme
	s := origin
	if idx := findSchemeEnd(s); idx >= 0 {
		s = s[idx:]
	}

	// Remove port
	if host, _, err := net.SplitHostPort(s); err == nil {
		return host
	}
	return s
}

func findSchemeEnd(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' && i+2 < len(s) && s[i+1] == '/' && s[i+2] == '/' {
			return i + 3
		}
	}
	return -1
}
