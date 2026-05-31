package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pcremote-server/config"
	"pcremote-server/middleware"
	winapi "pcremote-server/windows"
)

// ──────────────────────────────────────
// TestMockAPI — injectable mock for all Windows API calls
// ──────────────────────────────────────

type TestMockAPI struct {
	SetVolumeFunc            func(level float64) error
	GetVolumeStatusFunc      func() (float64, bool, error)
	SetMuteFunc              func(muted bool) error
	GetAllChannelVolumesFunc func() (map[string]winapi.ChannelStatus, error)
	SetChannelVolumeFunc     func(channelName string, level float64) error
	SetChannelMuteFunc       func(channelName string, muted bool) error
	GetDevicesFunc           func() ([]winapi.DeviceStatus, error)
	SetDeviceVolumeFunc      func(id string, level float64) error
	SetDeviceMuteFunc        func(id string, muted bool) error
	SendMediaKeyFunc         func(action string) error
	GetMediaStatusFunc       func() (map[string]any, error)
	LockWorkstationFunc      func() error
	OpenBrowserFunc          func(url string) error
	ScheduleShutdownFunc     func(delaySeconds int) error
	CancelShutdownFunc       func() error
	SleepFunc                func() error
	RestartFunc              func() error
	DebugListAllDevicesFunc  func() ([]string, error)
}

func (m *TestMockAPI) SetVolume(level float64) error {
	if m.SetVolumeFunc != nil {
		return m.SetVolumeFunc(level)
	}
	return nil
}
func (m *TestMockAPI) GetVolumeStatus() (float64, bool, error) {
	if m.GetVolumeStatusFunc != nil {
		return m.GetVolumeStatusFunc()
	}
	return 0.5, false, nil
}
func (m *TestMockAPI) SetMute(muted bool) error {
	if m.SetMuteFunc != nil {
		return m.SetMuteFunc(muted)
	}
	return nil
}
func (m *TestMockAPI) GetAllChannelVolumes() (map[string]winapi.ChannelStatus, error) {
	if m.GetAllChannelVolumesFunc != nil {
		return m.GetAllChannelVolumesFunc()
	}
	return map[string]winapi.ChannelStatus{
		"media": {Name: "Sonar - Media", Level: 0.75, Muted: false},
	}, nil
}
func (m *TestMockAPI) SetChannelVolume(channelName string, level float64) error {
	if m.SetChannelVolumeFunc != nil {
		return m.SetChannelVolumeFunc(channelName, level)
	}
	return nil
}
func (m *TestMockAPI) SetChannelMute(channelName string, muted bool) error {
	if m.SetChannelMuteFunc != nil {
		return m.SetChannelMuteFunc(channelName, muted)
	}
	return nil
}
func (m *TestMockAPI) GetDevices() ([]winapi.DeviceStatus, error) {
	if m.GetDevicesFunc != nil {
		return m.GetDevicesFunc()
	}
	return []winapi.DeviceStatus{
		{ID: "mock-device-id", Name: "Mock Speakers", Volume: 50.0, Muted: false},
	}, nil
}
func (m *TestMockAPI) SetDeviceVolume(id string, level float64) error {
	if m.SetDeviceVolumeFunc != nil {
		return m.SetDeviceVolumeFunc(id, level)
	}
	return nil
}
func (m *TestMockAPI) SetDeviceMute(id string, muted bool) error {
	if m.SetDeviceMuteFunc != nil {
		return m.SetDeviceMuteFunc(id, muted)
	}
	return nil
}
func (m *TestMockAPI) SendMediaKey(action string) error {
	if m.SendMediaKeyFunc != nil {
		return m.SendMediaKeyFunc(action)
	}
	return nil
}
func (m *TestMockAPI) GetMediaStatus() (map[string]any, error) {
	if m.GetMediaStatusFunc != nil {
		return m.GetMediaStatusFunc()
	}
	return map[string]any{
		"success": true,
		"status":  "playing",
		"title":   "Mock Title",
		"artist":  "Mock Artist",
	}, nil
}
func (m *TestMockAPI) LockWorkstation() error {
	if m.LockWorkstationFunc != nil {
		return m.LockWorkstationFunc()
	}
	return nil
}
func (m *TestMockAPI) OpenBrowser(url string) error {
	if m.OpenBrowserFunc != nil {
		return m.OpenBrowserFunc(url)
	}
	return nil
}
func (m *TestMockAPI) ScheduleShutdown(delaySeconds int) error {
	if m.ScheduleShutdownFunc != nil {
		return m.ScheduleShutdownFunc(delaySeconds)
	}
	return nil
}
func (m *TestMockAPI) CancelShutdown() error {
	if m.CancelShutdownFunc != nil {
		return m.CancelShutdownFunc()
	}
	return nil
}
func (m *TestMockAPI) Sleep() error {
	if m.SleepFunc != nil {
		return m.SleepFunc()
	}
	return nil
}
func (m *TestMockAPI) Restart() error {
	if m.RestartFunc != nil {
		return m.RestartFunc()
	}
	return nil
}
func (m *TestMockAPI) DebugListAllDevices() ([]string, error) {
	if m.DebugListAllDevicesFunc != nil {
		return m.DebugListAllDevicesFunc()
	}
	return nil, nil
}

// helper: inject mock and return cleanup func
func injectMock(m *TestMockAPI) func() {
	old := winapi.API
	winapi.API = m
	return func() { winapi.API = old }
}

// ──────────────────────────────────────
// Health
// ──────────────────────────────────────

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	HealthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["status"] != "ok" || resp["platform"] != "windows" {
		t.Errorf("unexpected body: %v", resp)
	}
}

// ──────────────────────────────────────
// Auth Middleware
// ──────────────────────────────────────

func setTestPIN(pin string) func() {
	old := config.App.PIN
	config.App.PIN = pin
	return func() { config.App.PIN = old }
}

func TestAuthMiddleware_NoPin(t *testing.T) {
	defer setTestPIN("1234")()
	handler := middleware.WithAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/protected", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_WrongPin(t *testing.T) {
	defer setTestPIN("1234")()
	handler := middleware.WithAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("X-PIN", "5678")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_CorrectPin(t *testing.T) {
	defer setTestPIN("1234")()
	handler := middleware.WithAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("X-PIN", "1234")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestAuthMiddleware_EmptyServerPin(t *testing.T) {
	defer setTestPIN("")()
	handler := middleware.WithAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/protected", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 when PIN unconfigured, got %d", rr.Code)
	}
}

// ──────────────────────────────────────
// Audio Volume
// ──────────────────────────────────────

func TestAudioVolumeHandler_InvalidLevel_TooHigh(t *testing.T) {
	defer injectMock(&TestMockAPI{})()
	body, _ := json.Marshal(VolumeRequest{Level: 1.5})
	req := httptest.NewRequest("POST", "/audio/volume", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	AudioVolumeHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestAudioVolumeHandler_InvalidLevel_Negative(t *testing.T) {
	defer injectMock(&TestMockAPI{})()
	body, _ := json.Marshal(VolumeRequest{Level: -0.1})
	req := httptest.NewRequest("POST", "/audio/volume", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	AudioVolumeHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestAudioVolumeHandler_ValidLevel(t *testing.T) {
	called := false
	defer injectMock(&TestMockAPI{
		SetVolumeFunc: func(level float64) error {
			if level == 0.5 {
				called = true
			}
			return nil
		},
	})()

	body, _ := json.Marshal(VolumeRequest{Level: 0.5})
	req := httptest.NewRequest("POST", "/audio/volume", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	AudioVolumeHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !called {
		t.Error("expected SetVolume to be called")
	}
}

func TestAudioVolumeHandler_WrongMethod(t *testing.T) {
	req := httptest.NewRequest("GET", "/audio/volume", nil)
	rr := httptest.NewRecorder()
	AudioVolumeHandler(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestAudioVolumeHandler_BadJSON(t *testing.T) {
	defer injectMock(&TestMockAPI{})()
	req := httptest.NewRequest("POST", "/audio/volume", strings.NewReader("not json"))
	rr := httptest.NewRecorder()
	AudioVolumeHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestAudioVolumeHandler_APIError(t *testing.T) {
	defer injectMock(&TestMockAPI{
		SetVolumeFunc: func(level float64) error {
			return fmt.Errorf("COM init failed")
		},
	})()
	body, _ := json.Marshal(VolumeRequest{Level: 0.5})
	req := httptest.NewRequest("POST", "/audio/volume", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	AudioVolumeHandler(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

// ──────────────────────────────────────
// Audio Mute
// ──────────────────────────────────────

func TestAudioMuteHandler_Valid(t *testing.T) {
	mutedVal := false
	defer injectMock(&TestMockAPI{
		SetMuteFunc: func(muted bool) error { mutedVal = muted; return nil },
	})()
	body, _ := json.Marshal(MuteRequest{Muted: true})
	req := httptest.NewRequest("POST", "/audio/mute", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	AudioMuteHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !mutedVal {
		t.Error("expected SetMute(true)")
	}
}

func TestAudioMuteHandler_WrongMethod(t *testing.T) {
	req := httptest.NewRequest("GET", "/audio/mute", nil)
	rr := httptest.NewRecorder()
	AudioMuteHandler(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

// ──────────────────────────────────────
// Audio Status
// ──────────────────────────────────────

func TestAudioStatusHandler_Valid(t *testing.T) {
	defer injectMock(&TestMockAPI{
		GetVolumeStatusFunc: func() (float64, bool, error) { return 0.75, true, nil },
	})()
	req := httptest.NewRequest("GET", "/audio/status", nil)
	rr := httptest.NewRecorder()
	AudioStatusHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var resp AudioStatusResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Level != 0.75 || !resp.Muted {
		t.Errorf("unexpected: level=%f muted=%v", resp.Level, resp.Muted)
	}
}

func TestAudioStatusHandler_WrongMethod(t *testing.T) {
	req := httptest.NewRequest("POST", "/audio/status", nil)
	rr := httptest.NewRecorder()
	AudioStatusHandler(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

// ──────────────────────────────────────
// Audio Channels
// ──────────────────────────────────────

func TestAudioChannelsHandler_Valid(t *testing.T) {
	defer injectMock(&TestMockAPI{})()
	req := httptest.NewRequest("GET", "/audio/channels", nil)
	rr := httptest.NewRecorder()
	AudioChannelsHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestAudioChannelVolumeHandler_MissingChannel(t *testing.T) {
	defer injectMock(&TestMockAPI{})()
	body, _ := json.Marshal(ChannelVolumeRequest{Channel: "", Level: 0.5})
	req := httptest.NewRequest("POST", "/audio/channel/volume", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	AudioChannelVolumeHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestAudioChannelVolumeHandler_InvalidLevel(t *testing.T) {
	defer injectMock(&TestMockAPI{})()
	body, _ := json.Marshal(ChannelVolumeRequest{Channel: "media", Level: 2.0})
	req := httptest.NewRequest("POST", "/audio/channel/volume", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	AudioChannelVolumeHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestAudioChannelMuteHandler_MissingChannel(t *testing.T) {
	defer injectMock(&TestMockAPI{})()
	body, _ := json.Marshal(ChannelMuteRequest{Channel: "", Muted: true})
	req := httptest.NewRequest("POST", "/audio/channel/mute", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	AudioChannelMuteHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

// ──────────────────────────────────────
// Browser
// ──────────────────────────────────────

func TestBrowserOpen_InvalidURL(t *testing.T) {
	defer injectMock(&TestMockAPI{
		OpenBrowserFunc: func(url string) error {
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				return fmt.Errorf("url must start with http:// or https://")
			}
			return nil
		},
	})()
	body, _ := json.Marshal(BrowserOpenRequest{URL: "ftp://example.com"})
	req := httptest.NewRequest("POST", "/browser/open", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	BrowserOpenHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestBrowserOpen_EmptyURL(t *testing.T) {
	defer injectMock(&TestMockAPI{})()
	body, _ := json.Marshal(BrowserOpenRequest{URL: ""})
	req := httptest.NewRequest("POST", "/browser/open", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	BrowserOpenHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestBrowserOpen_ValidURL(t *testing.T) {
	opened := ""
	defer injectMock(&TestMockAPI{
		OpenBrowserFunc: func(url string) error { opened = url; return nil },
	})()
	body, _ := json.Marshal(BrowserOpenRequest{URL: "https://google.com"})
	req := httptest.NewRequest("POST", "/browser/open", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	BrowserOpenHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if opened != "https://google.com" {
		t.Errorf("expected url to be passed, got %q", opened)
	}
}

// ──────────────────────────────────────
// Media
// ──────────────────────────────────────

func TestMediaPlayHandler_Valid(t *testing.T) {
	action := ""
	defer injectMock(&TestMockAPI{
		SendMediaKeyFunc: func(a string) error { action = a; return nil },
	})()
	req := httptest.NewRequest("POST", "/media/play", nil)
	rr := httptest.NewRecorder()
	MediaPlayHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if action != "play_pause" {
		t.Errorf("expected play_pause, got %q", action)
	}
}

func TestMediaNextHandler_Valid(t *testing.T) {
	action := ""
	defer injectMock(&TestMockAPI{
		SendMediaKeyFunc: func(a string) error { action = a; return nil },
	})()
	req := httptest.NewRequest("POST", "/media/next", nil)
	rr := httptest.NewRecorder()
	MediaNextHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if action != "next" {
		t.Errorf("expected next, got %q", action)
	}
}

func TestMediaPrevHandler_Valid(t *testing.T) {
	action := ""
	defer injectMock(&TestMockAPI{
		SendMediaKeyFunc: func(a string) error { action = a; return nil },
	})()
	req := httptest.NewRequest("POST", "/media/prev", nil)
	rr := httptest.NewRecorder()
	MediaPrevHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if action != "prev" {
		t.Errorf("expected prev, got %q", action)
	}
}

func TestMediaPlayHandler_WrongMethod(t *testing.T) {
	req := httptest.NewRequest("GET", "/media/play", nil)
	rr := httptest.NewRecorder()
	MediaPlayHandler(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

// ──────────────────────────────────────
// System
// ──────────────────────────────────────

func TestSystemLockHandler_Valid(t *testing.T) {
	called := false
	defer injectMock(&TestMockAPI{
		LockWorkstationFunc: func() error { called = true; return nil },
	})()
	req := httptest.NewRequest("POST", "/system/lock", nil)
	rr := httptest.NewRecorder()
	SystemLockHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !called {
		t.Error("expected LockWorkstation to be called")
	}
}

func TestSystemLockHandler_WrongMethod(t *testing.T) {
	req := httptest.NewRequest("GET", "/system/lock", nil)
	rr := httptest.NewRecorder()
	SystemLockHandler(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestShutdownHandler_NegativeDelay(t *testing.T) {
	defer injectMock(&TestMockAPI{})()
	body, _ := json.Marshal(ShutdownRequest{DelayMinutes: -5})
	req := httptest.NewRequest("POST", "/system/shutdown", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	SystemShutdownHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestShutdownHandler_ValidDelay(t *testing.T) {
	delayUsed := -1
	defer injectMock(&TestMockAPI{
		ScheduleShutdownFunc: func(d int) error { delayUsed = d; return nil },
	})()
	body, _ := json.Marshal(ShutdownRequest{DelayMinutes: 5})
	req := httptest.NewRequest("POST", "/system/shutdown", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	SystemShutdownHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if delayUsed != 300 {
		t.Errorf("expected delay 300, got %d", delayUsed)
	}
}

func TestShutdownCancelHandler_Valid(t *testing.T) {
	called := false
	defer injectMock(&TestMockAPI{
		CancelShutdownFunc: func() error { called = true; return nil },
	})()
	req := httptest.NewRequest("POST", "/system/shutdown/cancel", nil)
	rr := httptest.NewRecorder()
	SystemShutdownCancelHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !called {
		t.Error("expected CancelShutdown to be called")
	}
}

func TestShutdownCancelHandler_WrongMethod(t *testing.T) {
	req := httptest.NewRequest("GET", "/system/shutdown/cancel", nil)
	rr := httptest.NewRecorder()
	SystemShutdownCancelHandler(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}
