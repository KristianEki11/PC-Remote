package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const (
	// TokenLength is the number of random bytes used to generate a session token (32 bytes = 64 hex chars).
	TokenLength = 32

	// DefaultTokenTTL is how long a session token remains valid.
	DefaultTokenTTL = 24 * time.Hour

	// HeartbeatTimeout defines how long without activity before a device is considered offline.
	// The Flutter app polls /audio/status every 10s and /media/status every 5s,
	// so 90 seconds covers ~9 missed polls before showing the device as disconnected.
	HeartbeatTimeout = 90 * time.Second

	// CleanupInterval is how often the background goroutine checks for expired sessions.
	CleanupInterval = time.Minute
)

// Session represents an authenticated device connection.
type Session struct {
	Token      string    `json:"token"`
	DeviceName string    `json:"device_name"`
	RemoteAddr string    `json:"remote_addr"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

// IsOnline returns true if the device has been seen within the heartbeat window.
func (s *Session) IsOnline() bool {
	return time.Since(s.LastSeenAt) < HeartbeatTimeout
}

// SessionManager manages a single active session (max 1 device).
// A new login automatically revokes the previous session.
type SessionManager struct {
	mu       sync.RWMutex
	current  *Session
	tokenTTL time.Duration
	stopCh   chan struct{}

	// OnStatusChange is called whenever connection status changes.
	// connected=true means a device just paired/logged in.
	// connected=false means the device went offline or session expired.
	OnStatusChange func(connected bool, deviceName string)
}

// NewSessionManager creates a new session manager and starts the cleanup loop.
func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		tokenTTL: DefaultTokenTTL,
		stopCh:   make(chan struct{}),
	}
	go sm.cleanupLoop()
	return sm
}

// CreateSession generates a new cryptographic token and registers the device.
// Since max 1 device is allowed, any existing session is automatically revoked.
func (sm *SessionManager) CreateSession(deviceName, remoteAddr string) (*Session, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &Session{
		Token:      token,
		DeviceName: deviceName,
		RemoteAddr: remoteAddr,
		CreatedAt:  now,
		ExpiresAt:  now.Add(sm.tokenTTL),
		LastSeenAt: now,
	}

	sm.mu.Lock()
	sm.current = session
	sm.mu.Unlock()

	if sm.OnStatusChange != nil {
		sm.OnStatusChange(true, deviceName)
	}

	return session, nil
}

// ValidateToken checks if the provided token matches the active session.
// Returns the session and true if valid, nil and false otherwise.
// Also updates LastSeenAt to serve as an implicit heartbeat.
func (sm *SessionManager) ValidateToken(token string) (*Session, bool) {
	sm.mu.RLock()
	session := sm.current
	sm.mu.RUnlock()

	if session == nil || session.Token != token {
		return nil, false
	}

	if time.Now().After(session.ExpiresAt) {
		sm.RevokeAll()
		return nil, false
	}

	// Update last seen (implicit heartbeat from any authenticated request)
	sm.mu.Lock()
	session.LastSeenAt = time.Now()
	sm.mu.Unlock()

	return session, true
}

// RevokeAll disconnects the currently connected device.
func (sm *SessionManager) RevokeAll() {
	sm.mu.Lock()
	hadSession := sm.current != nil
	sm.current = nil
	sm.mu.Unlock()

	if hadSession && sm.OnStatusChange != nil {
		sm.OnStatusChange(false, "")
	}
}

// GetActiveSession returns the current session if it exists and hasn't expired.
func (sm *SessionManager) GetActiveSession() *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.current != nil && time.Now().Before(sm.current.ExpiresAt) {
		return sm.current
	}
	return nil
}

// IsDeviceConnected returns whether a device is actively online (seen recently)
// and the device name. Used by the system tray to show connection status.
func (sm *SessionManager) IsDeviceConnected() (bool, string) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.current == nil {
		return false, ""
	}
	if time.Now().After(sm.current.ExpiresAt) {
		return false, ""
	}
	return sm.current.IsOnline(), sm.current.DeviceName
}

// Stop terminates the cleanup goroutine.
func (sm *SessionManager) Stop() {
	close(sm.stopCh)
}

func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	var wasOnline bool

	for {
		select {
		case <-ticker.C:
			connected, name := sm.IsDeviceConnected()

			// Fire callback when device goes from online to offline
			if wasOnline && !connected {
				if sm.OnStatusChange != nil {
					sm.OnStatusChange(false, name)
				}
			}
			wasOnline = connected

			// Clean expired sessions
			sm.mu.Lock()
			if sm.current != nil && time.Now().After(sm.current.ExpiresAt) {
				sm.current = nil
			}
			sm.mu.Unlock()

		case <-sm.stopCh:
			return
		}
	}
}

// generateToken creates a cryptographically random hex-encoded token.
func generateToken() (string, error) {
	b := make([]byte, TokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
