package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

const (
	// PairTokenLength is the number of random bytes for a pairing token (16 bytes = 32 hex chars).
	PairTokenLength = 16

	// PairTokenTTL is how long a one-time pairing token is valid after generation.
	PairTokenTTL = 5 * time.Minute
)

// QRPayload is the JSON structure encoded into the QR code.
// The Flutter app scans this and uses the pair_token to establish a session.
type QRPayload struct {
	Host           string   `json:"host"`
	Port           string   `json:"port"`
	PairToken      string   `json:"pair_token"`
	Protocol       string   `json:"protocol"`
	Fingerprint    string   `json:"fingerprint"`
	ServerName     string   `json:"server_name"`
	AlternateHosts []string `json:"alternate_hosts,omitempty"`
	PublicURL      string   `json:"public_url,omitempty"`
}

// PairingManager handles one-time QR code pairing tokens.
type PairingManager struct {
	mu           sync.Mutex
	currentToken string
	expiresAt    time.Time
	used         bool

	host           string
	port           string
	fingerprint    string
	serverName     string
	alternateHosts []string
	publicURL      string
}

// NewPairingManager creates a new pairing manager with server connection details.
func NewPairingManager(host, port, fingerprint, serverName string, alternateHosts []string) *PairingManager {
	return &PairingManager{
		host:           host,
		port:           port,
		fingerprint:    fingerprint,
		serverName:     serverName,
		alternateHosts: alternateHosts,
	}
}

// UpdatePublicURL updates the public tunnel URL (e.g. from Cloudflare tunnel).
func (pm *PairingManager) UpdatePublicURL(url string) {
	pm.mu.Lock()
	pm.publicURL = url
	pm.mu.Unlock()
}

// GetPublicURL returns the current public tunnel URL.
func (pm *PairingManager) GetPublicURL() string {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.publicURL
}

// UpdateHost updates the server host address (e.g., when network changes).
func (pm *PairingManager) UpdateHost(host string) {
	pm.mu.Lock()
	pm.host = host
	pm.mu.Unlock()
}

// UpdateFingerprint updates the TLS certificate fingerprint.
func (pm *PairingManager) UpdateFingerprint(fp string) {
	pm.mu.Lock()
	pm.fingerprint = fp
	pm.mu.Unlock()
}

// GenerateToken creates a new one-time pairing token, invalidating any previous one.
func (pm *PairingManager) GenerateToken() (string, error) {
	b := make([]byte, PairTokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	pm.mu.Lock()
	pm.currentToken = token
	pm.expiresAt = time.Now().Add(PairTokenTTL)
	pm.used = false
	pm.mu.Unlock()

	return token, nil
}

// ValidateAndConsume checks if the provided token matches the current pairing token.
// If valid, the token is consumed (single-use) and cannot be reused.
func (pm *PairingManager) ValidateAndConsume(token string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.currentToken == "" || pm.used {
		return false
	}
	if time.Now().After(pm.expiresAt) {
		pm.currentToken = ""
		return false
	}
	if token != pm.currentToken {
		return false
	}

	pm.used = true
	pm.currentToken = ""
	return true
}

// UpdateAlternateHosts updates the list of fallback LAN hosts.
func (pm *PairingManager) UpdateAlternateHosts(hosts []string) {
	pm.mu.Lock()
	pm.alternateHosts = hosts
	pm.mu.Unlock()
}

// GetQRPayload returns the active QR payload, creating a new token only if the current one is expired or used.
// This prevents token invalidation race conditions when rendering QR pages.
func (pm *PairingManager) GetQRPayload() (*QRPayload, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.currentToken == "" || pm.used || time.Now().After(pm.expiresAt) {
		b := make([]byte, PairTokenLength)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		pm.currentToken = hex.EncodeToString(b)
		pm.expiresAt = time.Now().Add(PairTokenTTL)
		pm.used = false
	}

	payload := &QRPayload{
		Host:           pm.host,
		Port:           pm.port,
		PairToken:      pm.currentToken,
		Protocol:       "https",
		Fingerprint:    pm.fingerprint,
		ServerName:     pm.serverName,
		AlternateHosts: pm.alternateHosts,
		PublicURL:      pm.publicURL,
	}

	return payload, nil
}

// GetEncryptedPayload returns the AES-256-CBC encrypted QR string "PCR3:<base64>".
func (pm *PairingManager) GetEncryptedPayload() (string, error) {
	payload, err := pm.GetQRPayload()
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return EncryptQRPayload(jsonBytes)
}

// GetQRImage generates a QR code PNG image containing the encrypted pairing payload.
// size controls the width/height in pixels.
func (pm *PairingManager) GetQRImage(size int) ([]byte, error) {
	encrypted, err := pm.GetEncryptedPayload()
	if err != nil {
		return nil, err
	}

	return qrcode.Encode(encrypted, qrcode.Medium, size)
}

// GetCurrentPayload returns the current active pairing payload without generating a new token.
// Returns nil if no valid token exists (expired or consumed).
func (pm *PairingManager) GetCurrentPayload() *QRPayload {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.currentToken == "" || pm.used || time.Now().After(pm.expiresAt) {
		return nil
	}

	return &QRPayload{
		Host:        pm.host,
		Port:        pm.port,
		PairToken:   pm.currentToken,
		Protocol:    "https",
		Fingerprint: pm.fingerprint,
		ServerName:  pm.serverName,
	}
}

// ExpiresAt returns when the current pairing token expires.
func (pm *PairingManager) ExpiresAt() time.Time {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.expiresAt
}
