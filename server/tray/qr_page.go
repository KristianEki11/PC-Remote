package tray

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pcremote-server/auth"
	tlsutil "pcremote-server/tls"
	"pcremote-server/tunnel"
)

// QRPageHandler serves the Steam-inspired dashboard for pairing and managing connected devices.
// This endpoint is restricted to localhost.
type QRPageHandler struct {
	Pairing  *auth.PairingManager
	Sessions *auth.SessionManager
	Tunnel   *tunnel.TunnelManager
}

// ServeQRPage handles GET /internal/qr.
func (h *QRPageHandler) ServeQRPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Discover all usable LAN interfaces
	interfaces := tlsutil.GetAllLANInterfaces()
	interfaces = append(interfaces, tlsutil.InterfaceInfo{
		Name:      "USB Cable / Localhost",
		IP:        "127.0.0.1",
		IsPrimary: false,
	})

	if len(interfaces) > 0 {
		h.Pairing.UpdateHost(interfaces[0].IP)
		var alts []string
		for _, iface := range interfaces[1:] {
			alts = append(alts, iface.IP)
		}
		h.Pairing.UpdateAlternateHosts(alts)
	}

	// 2. Check active pairing session
	session := h.Sessions.GetActiveSession()
	isPaired := session != nil
	pairedDeviceName := ""
	pairedAddr := ""
	pairedSince := ""
	isOnline := false

	if isPaired {
		pairedDeviceName = session.DeviceName
		pairedAddr = session.RemoteAddr
		pairedSince = session.CreatedAt.Format("02 Jan 2006, 15:04 WIB")
		isOnline = session.IsOnline()
	}

	// 3. Generate QR payload and AES-256 encrypted string (single-pass)
	rawPayload, err := h.Pairing.GetQRPayload()
	if err != nil {
		slog.Error("Failed to get QR payload", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	jsonBytes, err := json.Marshal(rawPayload)
	if err != nil {
		slog.Error("Failed to marshal QR payload", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	encryptedPayload, err := auth.EncryptQRPayload(jsonBytes)
	if err != nil {
		slog.Error("Failed to encrypt QR payload", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	interfacesJSON, _ := json.Marshal(interfaces)
	expiresAt := h.Pairing.ExpiresAt()
	expiresIn := int(time.Until(expiresAt).Seconds())

	statusText := "Siap Pairing"
	statusClass := "disconnected"
	if isPaired {
		if isOnline {
			statusText = fmt.Sprintf("Terhubung: %s", pairedDeviceName)
			statusClass = "connected"
		} else {
			statusText = fmt.Sprintf("Tersimpan (Standby): %s", pairedDeviceName)
			statusClass = "standby"
		}
	}

	publicURL := ""
	tunnelStatus := "disabled"
	if h.Tunnel != nil {
		st, pURL, _ := h.Tunnel.GetStatus()
		tunnelStatus = string(st)
		publicURL = pURL
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")

	page := qrPageHTML
	page = strings.ReplaceAll(page, "{{HOST}}", rawPayload.Host)
	page = strings.ReplaceAll(page, "{{PORT}}", rawPayload.Port)
	page = strings.ReplaceAll(page, "{{STATUS_TEXT}}", statusText)
	page = strings.ReplaceAll(page, "{{STATUS_CLASS}}", statusClass)
	page = strings.ReplaceAll(page, "{{ENCRYPTED_PAYLOAD}}", encryptedPayload)
	page = strings.ReplaceAll(page, "{{FINGERPRINT}}", rawPayload.Fingerprint)
	page = strings.ReplaceAll(page, "{{EXPIRES_IN}}", strconv.Itoa(expiresIn))
	page = strings.ReplaceAll(page, "{{INTERFACES_JSON}}", string(interfacesJSON))
	page = strings.ReplaceAll(page, "{{SERVER_NAME}}", rawPayload.ServerName)
	page = strings.ReplaceAll(page, "{{PUBLIC_URL}}", publicURL)
	page = strings.ReplaceAll(page, "{{TUNNEL_STATUS}}", tunnelStatus)

	if isPaired {
		page = strings.ReplaceAll(page, "{{IS_PAIRED}}", "true")
	} else {
		page = strings.ReplaceAll(page, "{{IS_PAIRED}}", "false")
	}
	page = strings.ReplaceAll(page, "{{PAIRED_DEVICE_NAME}}", pairedDeviceName)
	page = strings.ReplaceAll(page, "{{PAIRED_ADDR}}", pairedAddr)
	page = strings.ReplaceAll(page, "{{PAIRED_SINCE}}", pairedSince)
	if isOnline {
		page = strings.ReplaceAll(page, "{{IS_ONLINE}}", "true")
	} else {
		page = strings.ReplaceAll(page, "{{IS_ONLINE}}", "false")
	}

	w.Write([]byte(page))
}

// ServeQRImage handles GET /internal/qr/image.
func (h *QRPageHandler) ServeQRImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pngData, err := h.Pairing.GetQRImage(300)
	if err != nil {
		slog.Error("Failed to generate QR image", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(pngData)
}

// ServeStatus handles GET /internal/status.
func (h *QRPageHandler) ServeStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := h.Sessions.GetActiveSession()
	connected, deviceName := h.Sessions.IsDeviceConnected()

	var publicURL string
	var tunnelStatus string
	if h.Tunnel != nil {
		st, pURL, _ := h.Tunnel.GetStatus()
		tunnelStatus = string(st)
		publicURL = pURL
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"has_session":   session != nil,
		"connected":     connected,
		"device_name":   deviceName,
		"public_url":    publicURL,
		"tunnel_status": tunnelStatus,
	})
}

const qrPageHTML = `<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>PC Remote — Dashboard</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;600;700&family=Nunito:wght@400;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
<style>
  :root {
    --color-bg-base: #FAFAFA;
    --color-text-main: #0A0A0A;
    --color-text-muted: #525252;
    --color-text-light: #525252;
    
    --color-primary: #171717;
    --color-primary-hover: #0A0A0A;
    
    --color-indicator-green: #3B82F6;
    --color-indicator-green-bg: rgba(59, 130, 246, 0.15);
    --color-indicator-red: #EF4444;
    --color-indicator-red-bg: rgba(239, 68, 68, 0.15);
    --color-indicator-amber: #F59E0B;
    --color-indicator-amber-bg: rgba(245, 158, 11, 0.15);

    --radius-modal: 24px;
    --radius-btn: 12px;
    --radius-input: 12px;
  }

  * { box-sizing: border-box; margin: 0; padding: 0; }

  body {
    font-family: 'DM Sans', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    background: var(--color-bg-base);
    color: var(--color-text-main);
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
  }
  
  h1, h2, h3, .qr-heading, .section-label, .steam-logo {
    font-family: 'Nunito', sans-serif;
  }

  /* Main Modal Container - Liquid Glass Apple Minimalist */
  .steam-modal {
    width: 100%;
    max-width: 1100px;
    background: rgba(255, 255, 255, 0.6);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border: 1px solid rgba(0, 0, 0, 0.05);
    border-radius: var(--radius-modal);
    box-shadow: 0 8px 32px 0 rgba(0, 0, 0, 0.05);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  /* Header Section */
  .steam-header {
    background: rgba(255, 255, 255, 0.5);
    padding: 24px 32px;
    border-bottom: 1px solid rgba(0, 0, 0, 0.05);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .header-brand {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .steam-logo {
    width: 44px;
    height: 44px;
    border-radius: 14px;
    background: var(--color-primary);
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-weight: 800;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }
  .steam-logo svg { width: 24px; height: 24px; fill: #fff; }

  .header-title-wrap h1 {
    font-size: 22px;
    font-weight: 800;
    color: var(--color-text-main);
    letter-spacing: -0.5px;
  }
  .header-title-wrap p {
    font-size: 13px;
    color: var(--color-text-muted);
    font-weight: 500;
    margin-top: 2px;
  }

  /* Status Badge */
  .status-badge {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    border-radius: 999px;
    font-size: 13px;
    font-weight: 700;
    letter-spacing: 0.3px;
    transition: all 0.3s ease;
    background: rgba(255, 255, 255, 0.7);
    border: 1px solid rgba(0, 0, 0, 0.05);
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.02);
  }
  .status-badge.connected {
    background: var(--color-indicator-green-bg);
    border: 1px solid rgba(59, 130, 246, 0.2);
    color: var(--color-indicator-green);
  }
  .status-badge.standby {
    background: var(--color-indicator-amber-bg);
    border: 1px solid rgba(245, 158, 11, 0.2);
    color: var(--color-indicator-amber);
  }
  .status-badge.disconnected {
    color: var(--color-text-muted);
  }
  .badge-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: currentColor;
    box-shadow: 0 0 8px currentColor;
  }

  /* Body Layout */
  .steam-body {
    padding: 48px;
    display: grid;
    grid-template-columns: 1.2fr 1fr;
    gap: 48px;
  }
  @media (max-width: 780px) {
    .steam-body {
      grid-template-columns: 1fr;
      padding: 24px;
      gap: 28px;
    }
  }

  .left-col {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .section-label {
    font-size: 12px;
    font-weight: 800;
    color: var(--color-text-main);
    text-transform: uppercase;
    letter-spacing: 1.5px;
    margin-bottom: 6px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .form-label {
    font-size: 13px;
    font-weight: 600;
    color: var(--color-text-muted);
  }

  .steam-select {
    width: 100%;
    padding: 12px 16px;
    background: rgba(255, 255, 255, 0.8);
    border: 1px solid rgba(0, 0, 0, 0.1);
    border-radius: var(--radius-input);
    color: var(--color-text-main);
    font-size: 14px;
    font-weight: 500;
    font-family: inherit;
    outline: none;
    cursor: pointer;
    transition: all 0.2s ease;
    box-shadow: inset 0 2px 4px rgba(0,0,0,0.02);
  }
  .steam-select:focus {
    border-color: var(--color-text-main);
    box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.1);
  }

  /* Info Tiles */
  .info-tiles-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }
  @media (max-width: 500px) {
    .info-tiles-grid { grid-template-columns: 1fr; }
  }

  .steam-tile {
    background: rgba(255, 255, 255, 0.6);
    border: 1px solid rgba(0, 0, 0, 0.05);
    border-radius: var(--radius-input);
    padding: 14px 16px;
    display: flex;
    flex-direction: column;
    gap: 6px;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.02);
  }
  .tile-k { font-size: 11px; font-weight: 700; color: var(--color-text-muted); text-transform: uppercase; letter-spacing: 0.5px; font-family: 'Nunito', sans-serif; }
  .tile-v { font-family: 'JetBrains Mono', monospace; font-size: 13px; color: var(--color-text-main); font-weight: 600; }

  /* Fingerprint Section */
  .fp-card {
    background: rgba(255, 255, 255, 0.6);
    border: 1px solid rgba(0, 0, 0, 0.05);
    border-radius: var(--radius-input);
    padding: 14px 16px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.02);
  }
  .fp-content {
    display: flex;
    flex-direction: column;
    gap: 4px;
    overflow: hidden;
  }
  .fp-value {
    font-family: 'JetBrains Mono', monospace;
    font-size: 12px;
    color: var(--color-text-main);
    word-break: break-all;
  }
  .fp-value.masked {
    letter-spacing: 3px;
    color: var(--color-text-muted);
  }

  .btn-gold {
    padding: 8px 16px;
    background: var(--color-primary);
    border: none;
    border-radius: var(--radius-btn);
    color: #fff;
    font-size: 13px;
    font-weight: 700;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 8px;
    white-space: nowrap;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }
  .btn-gold:hover {
    background: var(--color-primary-hover);
    transform: translateY(-2px) scale(1.02);
    box-shadow: 0 6px 16px rgba(0, 0, 0, 0.25);
  }

  /* Right Column */
  .right-col {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }

  .qr-steam-box {
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
  }

  .qr-heading {
    font-size: 14px;
    font-weight: 800;
    color: var(--color-text-main);
    text-transform: uppercase;
    letter-spacing: 1.5px;
    text-align: center;
  }

  /* QR Card */
  .qr-card-white {
    background: #FFFFFF;
    padding: 20px;
    border-radius: 20px;
    box-shadow: 0 16px 40px rgba(0, 0, 0, 0.05);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    border: 1px solid rgba(0, 0, 0, 0.05);
  }
  .qr-card-white:hover {
    transform: scale(1.03);
    box-shadow: 0 20px 48px rgba(0, 0, 0, 0.08);
  }
  .qr-card-white svg {
    display: block;
    width: 240px;
    height: 240px;
  }

  .qr-footer-hint {
    font-size: 13px;
    color: var(--color-text-muted);
    text-align: center;
    line-height: 1.5;
    font-weight: 500;
  }
  .qr-footer-hint span {
    color: var(--color-text-main);
    font-weight: 700;
  }

  /* Paired State Card */
  .paired-state-box {
    width: 100%;
    background: rgba(255, 255, 255, 0.7);
    border: 1px solid rgba(59, 130, 246, 0.2);
    border-radius: var(--radius-modal);
    padding: 32px 24px;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 16px;
    box-shadow: 0 16px 40px rgba(0, 0, 0, 0.04);
    backdrop-filter: blur(20px);
  }
  .paired-state-icon {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    background: var(--color-indicator-green-bg);
    border: 1px solid rgba(59, 130, 246, 0.2);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--color-indicator-green);
  }
  .paired-state-icon svg { width: 32px; height: 32px; }
  
  .paired-state-title {
    font-size: 20px;
    font-family: 'Nunito', sans-serif;
    font-weight: 800;
    color: var(--color-text-main);
  }
  .paired-state-meta {
    font-size: 13px;
    color: var(--color-text-muted);
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-top: 8px;
  }

  .btn-unpair-danger {
    width: 100%;
    margin-top: 12px;
    padding: 12px 20px;
    border-radius: var(--radius-btn);
    background: #fff;
    border: 1px solid var(--color-indicator-red);
    color: var(--color-indicator-red);
    font-size: 14px;
    font-weight: 700;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    box-shadow: 0 4px 12px rgba(239, 68, 68, 0.1);
  }
  .btn-unpair-danger:hover {
    background: var(--color-indicator-red-bg);
    transform: translateY(-2px) scale(1.02);
    box-shadow: 0 6px 16px rgba(239, 68, 68, 0.2);
  }

  /* PIN Modal */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.4);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.3s ease;
    z-index: 1000;
  }
  .modal-overlay.active {
    opacity: 1;
    pointer-events: auto;
  }
  .modal-card {
    background: rgba(255, 255, 255, 0.9);
    backdrop-filter: blur(20px);
    border: 1px solid rgba(0, 0, 0, 0.1);
    border-radius: var(--radius-modal);
    padding: 32px;
    width: 100%;
    max-width: 400px;
    box-shadow: 0 24px 48px rgba(0, 0, 0, 0.1);
    display: flex;
    flex-direction: column;
    gap: 20px;
  }
  .modal-card h3 { font-size: 18px; font-weight: 800; color: var(--color-text-main); }
  .modal-card p { font-size: 14px; color: var(--color-text-muted); line-height: 1.5; }
  .pin-field {
    width: 100%;
    padding: 14px;
    font-size: 20px;
    text-align: center;
    letter-spacing: 8px;
    border-radius: var(--radius-input);
    background: rgba(255, 255, 255, 0.8);
    border: 1px solid rgba(0, 0, 0, 0.2);
    color: var(--color-text-main);
    outline: none;
    transition: all 0.2s ease;
    font-family: 'JetBrains Mono', monospace;
    font-weight: 600;
  }
  .pin-field:focus { 
    border-color: var(--color-text-main); 
    box-shadow: 0 0 0 4px rgba(0, 0, 0, 0.1);
  }
  .modal-btns { display: flex; gap: 12px; margin-top: 8px; }
  .btn-modal-cancel {
    flex: 1;
    padding: 12px;
    background: rgba(255, 255, 255, 0.6);
    border: 1px solid rgba(0, 0, 0, 0.1);
    color: var(--color-text-muted);
    border-radius: var(--radius-btn);
    font-weight: 700;
    cursor: pointer;
    transition: all 0.2s;
  }
  .btn-modal-cancel:hover {
    background: rgba(255, 255, 255, 1);
  }
  .btn-modal-submit {
    flex: 1;
    padding: 12px;
    background: var(--color-primary);
    border: none;
    color: #fff;
    border-radius: var(--radius-btn);
    font-weight: 700;
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }
  .btn-modal-submit:hover {
    background: var(--color-primary-hover);
    transform: translateY(-2px);
    box-shadow: 0 6px 16px rgba(0, 0, 0, 0.25);
  }
  .error-msg { color: var(--color-indicator-red); font-size: 13px; font-weight: 500; text-align: center; display: none; }
</style>
<script src="https://cdn.jsdelivr.net/npm/qrcode-generator@1.4.4/qrcode.min.js"></script>
</head>
<body>

<div class="steam-modal">
  <header class="steam-header">
    <div class="header-brand">
      <div class="steam-logo">
        <svg viewBox="0 0 24 24"><path d="M4 6h16v10H4z M2 18h20v2H2z"/></svg>
      </div>
      <div class="header-title-wrap">
        <h1>PC Remote Connection</h1>
        <p>{{SERVER_NAME}} • Windows PC Service v4.1.0</p>
      </div>
    </div>
    <div class="status-badge {{STATUS_CLASS}}" id="statusBadge">
      <span class="badge-dot"></span>
      <span id="statusText">{{STATUS_TEXT}}</span>
    </div>
  </header>

  <div class="steam-body">
    <div class="left-col">
      <div>
        <div class="section-label">PENGATURAN KONEKSI SERVER</div>
        <div class="form-group" style="margin-top: 8px;">
          <label class="form-label">Jalur Adapter Jaringan (IP)</label>
          <select class="steam-select" id="networkSelect" onchange="switchAdapter(this.value)">
          </select>
        </div>
      </div>

      <div class="info-tiles-grid">
        <div class="steam-tile">
          <span class="tile-k">Server Host (LAN)</span>
          <span class="tile-v" id="tileHost">{{HOST}}:{{PORT}}</span>
        </div>
        <div class="steam-tile">
          <span class="tile-k">Protokol Keamanan</span>
          <span class="tile-v" style="color: var(--color-indicator-green);">HTTPS (TLS Encrypted)</span>
        </div>
      </div>

      <div class="steam-tile">
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span class="tile-k" style="color: var(--color-text-main);">Akses Publik / Internet (4G/5G)</span>
          <span id="tunnelBadge" class="status-badge connected" style="padding: 4px 10px; font-size: 11px;">🟢 Aktif</span>
        </div>
        <span class="tile-v" id="tilePublicUrl" style="font-size: 12px; word-break: break-all; color: var(--color-text-muted);">{{PUBLIC_URL}}</span>
      </div>

      <div>
        <div class="section-label">KEAMANAN & SIDIK JARI TLS</div>
        <div class="fp-card">
          <div class="fp-content">
            <span class="tile-k">SHA-256 Fingerprint</span>
            <span class="fp-value masked" id="fpVal">••••••••••••••••••••••••••••••••••••••••</span>
          </div>
          <button class="btn-gold" id="btnUnlockFp" onclick="openPinModal()">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
            Buka Kunci
          </button>
        </div>
      </div>

      <div style="font-size: 13px; color: var(--color-text-muted); line-height: 1.6; margin-top: auto; padding: 16px; background: rgba(255, 255, 255, 0.5); border-radius: 12px; border: 1px solid rgba(0, 0, 0, 0.05);">
        💡 <strong>Petunjuk:</strong> Buka aplikasi <strong>PC Remote</strong> di HP Anda, lalu scan QR Code di sebelah kanan. Aplikasi mendukung <strong>Dual-Stack</strong> (otomatis via WiFi saat di rumah, via Internet publik saat di luar).
      </div>
    </div>

    <div class="right-col">
      <div class="qr-steam-box" id="unpairedView" style="display: {{IS_PAIRED}} ? 'none' : 'flex';">
        <div class="qr-heading">HUBUNGKAN DENGAN QR CODE</div>
        <div class="qr-card-white">
          <div id="qrCanvas"></div>
        </div>
        <div class="qr-footer-hint">
          Gunakan <span>Aplikasi PC Remote</span><br>untuk pairing otomatis via kode QR
        </div>
      </div>

      <div class="paired-state-box" id="pairedView" style="display: {{IS_PAIRED}} ? 'flex' : 'none';">
        <div class="paired-state-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="2" width="14" height="20" rx="2" ry="2"/><line x1="12" y1="18" x2="12.01" y2="18"/></svg>
        </div>
        <div>
          <div class="paired-state-title" id="pairedDeviceName">{{PAIRED_DEVICE_NAME}}</div>
          <div class="paired-state-meta">
            <span>IP: <strong style="color: var(--color-text-main);" id="pairedAddr">{{PAIRED_ADDR}}</strong></span>
            <span>Tersambung: <span id="pairedSince">{{PAIRED_SINCE}}</span></span>
          </div>
        </div>
        <div style="font-size: 12px; font-weight: 600; color: var(--color-indicator-green); background: var(--color-indicator-green-bg); padding: 6px 16px; border-radius: 999px; border: 1px solid rgba(59, 130, 246, 0.2);">
          🔒 Sesi Aktif & Terdaftar
        </div>
        <button class="btn-unpair-danger" onclick="unpairDevice()">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M18.36 6.64a9 9 0 1 1-12.73 0"/><line x1="12" y1="2" x2="12" y2="12"/></svg>
          Putuskan Perangkat
        </button>
      </div>
    </div>
  </div>
</div>

<div class="modal-overlay" id="pinModal">
  <div class="modal-card">
    <h3>Verifikasi PIN Server</h3>
    <p>Masukkan Master PIN PC Remote untuk membuka detail sidik jari sertifikat TLS.</p>
    <input type="password" id="pinInput" class="pin-field" placeholder="••••" maxlength="8" autofocus onkeydown="if(event.key==='Enter') submitPinVerification()">
    <div class="error-msg" id="pinError">PIN yang dimasukkan salah.</div>
    <div class="modal-btns">
      <button class="btn-modal-cancel" onclick="closePinModal()">Batal</button>
      <button class="btn-modal-submit" onclick="submitPinVerification()">Buka Kunci</button>
    </div>
  </div>
</div>

<script>
  const isPairedInit = {{IS_PAIRED}};
  const realFingerprint = "{{FINGERPRINT}}";
  let encryptedPayload = "{{ENCRYPTED_PAYLOAD}}";
  const networkInterfaces = {{INTERFACES_JSON}};
  let currentPublicURL = "{{PUBLIC_URL}}";

  function initPage() {
    const isPaired = isPairedInit;
    document.getElementById('pairedView').style.display = isPaired ? 'flex' : 'none';
    document.getElementById('unpairedView').style.display = isPaired ? 'none' : 'flex';

    updateTunnelDisplay("{{TUNNEL_STATUS}}", currentPublicURL);

    if (!isPaired) {
      const select = document.getElementById('networkSelect');
      select.innerHTML = '';
      networkInterfaces.forEach(iface => {
        const opt = document.createElement('option');
        opt.value = iface.ip;
        opt.textContent = iface.name + ' (' + iface.ip + ')' + (iface.is_primary ? ' — Default' : '');
        select.appendChild(opt);
      });
      renderQR(encryptedPayload);
    }
  }

  function updateTunnelDisplay(status, url) {
    const badge = document.getElementById('tunnelBadge');
    const urlEl = document.getElementById('tilePublicUrl');

    if (url && url.length > 0) {
      badge.textContent = '🟢 Aktif';
      badge.className = 'status-badge connected';
      urlEl.textContent = url;
    } else if (status === 'starting') {
      badge.textContent = '⏳ Menghubungkan';
      badge.className = 'status-badge standby';
      urlEl.textContent = 'Menghubungkan ke Cloudflare Edge...';
    } else {
      badge.textContent = '⚪ Offline';
      badge.className = 'status-badge disconnected';
      urlEl.textContent = 'Menunggu koneksi internet...';
    }
  }

  function renderQR(text) {
    try {
      const qr = qrcode(0, 'M');
      qr.addData(text);
      qr.make();
      document.getElementById('qrCanvas').innerHTML = qr.createSvgTag({ scalable: true, margin: 0 });
      const svg = document.querySelector('#qrCanvas svg');
      if (svg) {
        svg.style.width = '240px';
        svg.style.height = '240px';
      }
    } catch (e) {
      console.error('QR render error:', e);
    }
  }

  function switchAdapter(ip) {
    document.getElementById('tileHost').textContent = ip + ':{{PORT}}';
  }

  async function unpairDevice() {
    if (!confirm('Putuskan sambungan perangkat ini? HP harus scan QR ulang untuk menghubungkan kembali.')) return;
    try {
      const res = await fetch('/internal/unpair', { method: 'POST' });
      if (res.ok) {
        window.location.reload();
      }
    } catch (e) {
      alert('Gagal: ' + e);
    }
  }

  function openPinModal() {
    document.getElementById('pinModal').classList.add('active');
    document.getElementById('pinInput').value = '';
    document.getElementById('pinError').style.display = 'none';
    document.getElementById('pinInput').focus();
  }

  function closePinModal() {
    document.getElementById('pinModal').classList.remove('active');
  }

  async function submitPinVerification() {
    const pin = document.getElementById('pinInput').value.trim();
    if (!pin) return;

    try {
      const res = await fetch('/internal/verify-pin', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ pin: pin })
      });

      if (res.ok) {
        const fpEl = document.getElementById('fpVal');
        fpEl.textContent = realFingerprint;
        fpEl.classList.remove('masked');
        document.getElementById('btnUnlockFp').style.display = 'none';
        closePinModal();
      } else {
        document.getElementById('pinError').style.display = 'block';
      }
    } catch (e) {
      document.getElementById('pinError').textContent = 'Error: ' + e;
      document.getElementById('pinError').style.display = 'block';
    }
  }

  setInterval(async () => {
    try {
      const res = await fetch('/internal/status');
      if (res.ok) {
        const data = await res.json();
        if (data.has_session !== isPairedInit) {
          window.location.reload();
        }
        if (data.public_url !== currentPublicURL) {
          currentPublicURL = data.public_url;
          updateTunnelDisplay(data.tunnel_status, data.public_url);
          if (!isPairedInit) {
            window.location.reload();
          }
        }
      }
    } catch (_) {}
  }, 3000);

  initPage();
</script>

</body>
</html>`
