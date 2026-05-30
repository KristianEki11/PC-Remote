package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	winapi "pcremote-server/windows"
)

// ──────────────────────────────────────
// Request / Response Models
// ──────────────────────────────────────

type VolumeRequest struct {
	Level float64 `json:"level"`
}

type MuteRequest struct {
	Muted bool `json:"muted"`
}

type AudioStatusResponse struct {
	Level float64 `json:"level"`
	Muted bool    `json:"muted"`
}

type ChannelVolumeRequest struct {
	Channel string  `json:"channel"`
	Level   float64 `json:"level"`
}

type ChannelMuteRequest struct {
	Channel string `json:"channel"`
	Muted   bool   `json:"muted"`
}

type DeviceVolumeRequest struct {
	DeviceID string  `json:"device_id"`
	Level    float64 `json:"level"`
}

type DeviceMuteRequest struct {
	DeviceID string `json:"device_id"`
	Mute     bool   `json:"mute"`
	Muted    bool   `json:"muted"`
}

// ──────────────────────────────────────
// Handlers — master volume (targets Sonar Media)
// ──────────────────────────────────────

func AudioVolumeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}
	var req VolumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "invalid request body"})
		return
	}
	if req.Level < 0.0 || req.Level > 1.0 {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "level must be between 0.0 and 1.0"})
		return
	}
	if err := winapi.SetVolume(req.Level); err != nil {
		slog.Error("SetVolume failed", "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}
	sendJSON(w, http.StatusOK, map[string]any{"success": true, "level": req.Level})
}

func AudioMuteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}
	var req MuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "invalid request body"})
		return
	}
	if err := winapi.SetMute(req.Muted); err != nil {
		slog.Error("SetMute failed", "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}
	sendJSON(w, http.StatusOK, map[string]any{"success": true, "muted": req.Muted})
}

func AudioStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}
	level, muted, err := winapi.GetVolumeStatus()
	if err != nil {
		slog.Error("GetVolumeStatus failed", "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}
	sendJSON(w, http.StatusOK, AudioStatusResponse{Level: level, Muted: muted})
}

// ──────────────────────────────────────
// Handlers — Sonar multi-channel
// ──────────────────────────────────────

// AudioChannelsHandler handles GET /audio/channels
// Returns volume/mute for all discovered Sonar channels.
func AudioChannelsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}
	channels, err := winapi.GetAllChannelVolumes()
	if err != nil {
		slog.Error("GetAllChannelVolumes failed", "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}
	sendJSON(w, http.StatusOK, map[string]any{"channels": channels})
}

// AudioChannelVolumeHandler handles POST /audio/channel/volume
// Sets volume on a specific Sonar channel. Body: {"channel":"media","level":0.8}
func AudioChannelVolumeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}
	var req ChannelVolumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "invalid request body"})
		return
	}
	if req.Channel == "" {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "channel is required"})
		return
	}
	if req.Level < 0.0 || req.Level > 1.0 {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "level must be between 0.0 and 1.0"})
		return
	}
	if err := winapi.SetChannelVolume(req.Channel, req.Level); err != nil {
		slog.Error("SetChannelVolume failed", "channel", req.Channel, "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}
	sendJSON(w, http.StatusOK, map[string]any{"success": true, "channel": req.Channel, "level": req.Level})
}

// AudioChannelMuteHandler handles POST /audio/channel/mute
// Sets mute on a specific Sonar channel. Body: {"channel":"media","muted":true}
func AudioChannelMuteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}
	var req ChannelMuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "invalid request body"})
		return
	}
	if req.Channel == "" {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "channel is required"})
		return
	}
	if err := winapi.SetChannelMute(req.Channel, req.Muted); err != nil {
		slog.Error("SetChannelMute failed", "channel", req.Channel, "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}
	sendJSON(w, http.StatusOK, map[string]any{"success": true, "channel": req.Channel, "muted": req.Muted})
}

// AudioDevicesHandler handles GET /audio/devices
func AudioDevicesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}
	devices, err := winapi.GetDevices()
	if err != nil {
		slog.Error("GetDevices failed", "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}
	sendJSON(w, http.StatusOK, map[string]any{"devices": devices})
}

// AudioDeviceVolumeHandler handles POST /audio/device/volume
func AudioDeviceVolumeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}
	var req DeviceVolumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "invalid request body"})
		return
	}
	if req.DeviceID == "" {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "device_id is required"})
		return
	}
	if req.Level < 0.0 || req.Level > 1.0 {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "level must be between 0.0 and 1.0"})
		return
	}
	if err := winapi.SetDeviceVolume(req.DeviceID, req.Level); err != nil {
		slog.Error("SetDeviceVolume failed", "device_id", req.DeviceID, "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}
	sendJSON(w, http.StatusOK, map[string]any{"success": true, "device_id": req.DeviceID, "level": req.Level})
}

// AudioDeviceMuteHandler handles POST /audio/device/mute
func AudioDeviceMuteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}
	var req DeviceMuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "invalid request body"})
		return
	}
	if req.DeviceID == "" {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "device_id is required"})
		return
	}
	muted := req.Mute || req.Muted
	if err := winapi.SetDeviceMute(req.DeviceID, muted); err != nil {
		slog.Error("SetDeviceMute failed", "device_id", req.DeviceID, "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}
	sendJSON(w, http.StatusOK, map[string]any{"success": true, "device_id": req.DeviceID, "muted": muted})
}
