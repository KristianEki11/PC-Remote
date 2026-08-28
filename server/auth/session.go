package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// TokenLength is the number of random bytes used to generate a session token (32 bytes = 64 hex chars).
	TokenLength = 32

	// DefaultTokenTTL is how long a session token remains valid (1 year for seamless persistent pairing).
	DefaultTokenTTL = 365 * 24 * time.Hour

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

// SessionManager manages a single active session (max 1 device) with persistent disk backing.
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

// NewSessionManager creates a new session manager, restores any saved session from disk,
// and starts the cleanup loop.
func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		tokenTTL: DefaultTokenTTL,
		stopCh:   make(chan struct{}),
	}

	// Restore persisted session from disk if still valid
	if saved := loadSessionFromDisk(); saved != nil {
		if time.Now().Before(saved.ExpiresAt) {
			sm.current = saved
			slog.Info("Restored active paired session from disk",
				"device", saved.DeviceName,
				"paired_at", saved.CreatedAt.Format(time.RFC3339),
			)
		} else {
			deleteSessionDiskFile()
		}
	}

	go sm.cleanupLoop()
	return sm
}

// CreateSession generates a new cryptographic token, registers the device, and persists to disk.
// Since max 1 device is allowed, any existing session is automatically replaced.
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

	// Persist to disk so server restarts don't drop the paired device
	saveSessionToDisk(session)

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

// RevokeAll disconnects the currently connected device and removes the persisted session file.
func (sm *SessionManager) RevokeAll() {
	sm.mu.Lock()
	hadSession := sm.current != nil
	sm.current = nil
	sm.mu.Unlock()

	deleteSessionDiskFile()

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

// ──────────────────────────────────────
// Disk Persistence Helpers
// ──────────────────────────────────────

func getSessionFilePath() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		dir := filepath.Join(localAppData, "PCRemote")
		_ = os.MkdirAll(dir, 0700)
		return filepath.Join(dir, "session.json")
	}
	return "session.json"
}

func saveSessionToDisk(s *Session) {
	if s == nil {
		return
	}
	filePath := getSessionFilePath()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal session for disk", "error", err)
		return
	}

	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		slog.Error("Failed to write temp session file", "error", err)
		return
	}

	_ = os.Remove(filePath)
	if err := os.Rename(tmpPath, filePath); err != nil {
		slog.Error("Failed to commit session file to disk", "error", err)
	}
}

func loadSessionFromDisk() *Session {
	filePath := getSessionFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		slog.Warn("Failed to unmarshal session from disk", "error", err)
		return nil
	}

	return &session
}

func deleteSessionDiskFile() {
	filePath := getSessionFilePath()
	_ = os.Remove(filePath)
	_ = os.Remove(filePath + ".tmp")
}
