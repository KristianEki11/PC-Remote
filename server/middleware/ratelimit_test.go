package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsLocalhost(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		host       string
		headers    map[string]string
		want       bool
	}{
		{
			name:       "Direct Localhost IPv4",
			remoteAddr: "127.0.0.1:54321",
			host:       "localhost:8000",
			headers:    nil,
			want:       true,
		},
		{
			name:       "Direct Localhost 127.0.0.1 IP",
			remoteAddr: "127.0.0.1:54321",
			host:       "127.0.0.1:8000",
			headers:    nil,
			want:       true,
		},
		{
			name:       "Direct Localhost IPv6",
			remoteAddr: "[::1]:54321",
			host:       "[::1]:8000",
			headers:    nil,
			want:       true,
		},
		{
			name:       "Remote LAN IP",
			remoteAddr: "192.168.1.105:54321",
			host:       "192.168.1.50:8000",
			headers:    nil,
			want:       false,
		},
		{
			name:       "Cloudflare Tunnel Request (CF-Ray Header)",
			remoteAddr: "127.0.0.1:54321",
			host:       "my-quick-tunnel.trycloudflare.com",
			headers: map[string]string{
				"CF-Ray": "8572abc1234-SIN",
			},
			want: false,
		},
		{
			name:       "Cloudflare Tunnel Request (CF-Connecting-IP)",
			remoteAddr: "127.0.0.1:54321",
			host:       "localhost:8000",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.195",
			},
			want: false,
		},
		{
			name:       "Reverse Proxy (X-Forwarded-For)",
			remoteAddr: "127.0.0.1:54321",
			host:       "localhost:8000",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195",
			},
			want: false,
		},
		{
			name:       "External Host Header Targeting Tunnel",
			remoteAddr: "127.0.0.1:54321",
			host:       "random-slug.trycloudflare.com",
			headers:    nil,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://"+tt.host+"/internal/qr", nil)
			req.RemoteAddr = tt.remoteAddr
			req.Host = tt.host
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := IsLocalhost(req)
			if got != tt.want {
				t.Errorf("IsLocalhost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocalhostOnlyMiddleware(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	protected := LocalhostOnly(dummyHandler)

	// Case 1: Valid direct localhost
	req1 := httptest.NewRequest("GET", "http://localhost:8000/internal/qr", nil)
	req1.RemoteAddr = "127.0.0.1:12345"
	rr1 := httptest.NewRecorder()
	protected.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Errorf("expected 200 OK for valid localhost, got %d", rr1.Code)
	}

	// Case 2: Blocked Cloudflare request
	req2 := httptest.NewRequest("GET", "http://tunnel.trycloudflare.com/internal/qr", nil)
	req2.RemoteAddr = "127.0.0.1:12345"
	req2.Header.Set("CF-Ray", "12345678")
	rr2 := httptest.NewRecorder()
	protected.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for Cloudflare tunnel request, got %d", rr2.Code)
	}
}
