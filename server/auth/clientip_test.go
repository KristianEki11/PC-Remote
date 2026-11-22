package auth

import (
	"net/http/httptest"
	"testing"
)

func TestClientIP(t *testing.T) {
	cases := []struct {
		name, remote, cf, want string
	}{
		{"strips port", "192.168.1.5:5555", "", "192.168.1.5"},
		{"loopback with CF header", "127.0.0.1:40000", "203.0.113.7", "203.0.113.7"},
		{"loopback without CF header", "127.0.0.1:40000", "", "127.0.0.1"},
	}
	for _, c := range cases {
		r := httptest.NewRequest("POST", "/auth/login", nil)
		r.RemoteAddr = c.remote
		if c.cf != "" {
			r.Header.Set("CF-Connecting-IP", c.cf)
		}
		if got := ClientIP(r); got != c.want {
			t.Errorf("%s: ClientIP() = %q, want %q", c.name, got, c.want)
		}
	}
}
