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
)

// QRPageHandler serves the Steam-inspired dashboard for pairing and managing connected devices.
// This endpoint is restricted to localhost.
type QRPageHandler struct {
	Pairing  *auth.PairingManager
	Sessions *auth.SessionManager
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

	// 3. Generate AES-256 encrypted QR payload
	encryptedPayload, err := h.Pairing.GetEncryptedPayload()
	if err != nil {
		slog.Error("Failed to generate encrypted QR payload", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	rawPayload, _ := h.Pairing.GetQRPayload()
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"has_session": session != nil,
		"connected":   connected,
		"device_name": deviceName,
	})
}

const qrPageHTML = `<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>PC Remote — Steam-Style Dashboard</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
<style>
  :root {
    /* Steam-Style Custom Palette */
    --color-bg-base: #171D25;
    --color-bg-surface: #2C3947;
    --color-bg-elevated: #1E2630;
    --color-accent-slate: #547A95;
    --color-accent-gold: #C2A56D;
    --color-text-light: #E8EDF2;
    --color-text-white: #FFFFFF;
    --color-text-muted: #93A8B8;
    --color-border: rgba(84, 122, 149, 0.35);
    --color-border-focus: #547A95;
    
    /* Indicator Colors */
    --color-indicator-green: #22C55E;
    --color-indicator-green-bg: rgba(34, 197, 94, 0.15);
    --color-indicator-red: #EF4444;
    --color-indicator-red-bg: rgba(239, 68, 68, 0.15);
    --color-indicator-amber: #F59E0B;
    --color-indicator-amber-bg: rgba(245, 158, 11, 0.15);

    --radius-modal: 16px;
    --radius-btn: 8px;
    --radius-input: 6px;
  }

  * { box-sizing: border-box; margin: 0; padding: 0; }

  body {
    font-family: 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    background-color: var(--color-bg-base);
    color: var(--color-text-light);
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    background-image: 
      radial-gradient(circle at 20% 15%, rgba(84, 122, 149, 0.18), transparent 45%),
      radial-gradient(circle at 80% 85%, rgba(194, 165, 109, 0.12), transparent 45%),
      linear-gradient(180deg, #18202A 0%, #12161C 100%);
    background-attachment: fixed;
  }

  /* Main Steam-Like Modal Container */
  .steam-modal {
    width: 100%;
    max-width: 860px;
    background: var(--color-bg-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-modal);
    box-shadow: 0 30px 60px -15px rgba(0, 0, 0, 0.8), 0 0 0 1px rgba(255, 255, 255, 0.05);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  /* Header Section */
  .steam-header {
    background: #1F2833;
    padding: 20px 28px;
    border-bottom: 1px solid rgba(84, 122, 149, 0.25);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .header-brand {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .steam-logo {
    width: 36px;
    height: 36px;
    border-radius: 8px;
    background: linear-gradient(135deg, var(--color-accent-gold) 0%, #8E7445 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    color: #171D25;
    font-weight: 800;
    box-shadow: 0 4px 12px rgba(194, 165, 109, 0.3);
  }
  .steam-logo svg { width: 20px; height: 20px; fill: #171D25; }

  .header-title-wrap h1 {
    font-size: 20px;
    font-weight: 800;
    color: var(--color-text-white);
    letter-spacing: -0.3px;
  }
  .header-title-wrap p {
    font-size: 12px;
    color: var(--color-text-muted);
    font-weight: 500;
    margin-top: 1px;
  }

  /* Status Badge */
  .status-badge {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 14px;
    border-radius: 999px;
    font-size: 12px;
    font-weight: 600;
    letter-spacing: 0.3px;
    transition: all 0.3s ease;
  }
  .status-badge.connected {
    background: var(--color-indicator-green-bg);
    border: 1px solid rgba(34, 197, 94, 0.4);
    color: var(--color-indicator-green);
  }
  .status-badge.standby {
    background: var(--color-indicator-amber-bg);
    border: 1px solid rgba(245, 158, 11, 0.4);
    color: var(--color-indicator-amber);
  }
  .status-badge.disconnected {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--color-border);
    color: var(--color-text-muted);
  }
  .badge-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: currentColor;
    box-shadow: 0 0 8px currentColor;
  }

  /* Body: Two Column Steam Layout */
  .steam-body {
    padding: 32px 36px;
    display: grid;
    grid-template-columns: 1fr 340px;
    gap: 36px;
  }
  @media (max-width: 780px) {
    .steam-body {
      grid-template-columns: 1fr;
      padding: 24px;
      gap: 28px;
    }
  }

  /* Left Column: Form & Connection Info */
  .left-col {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .section-label {
    font-size: 11px;
    font-weight: 800;
    color: var(--color-accent-gold);
    text-transform: uppercase;
    letter-spacing: 1px;
    margin-bottom: 6px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .form-label {
    font-size: 12px;
    font-weight: 600;
    color: var(--color-text-muted);
  }

  .steam-select {
    width: 100%;
    padding: 12px 14px;
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-input);
    color: var(--color-text-white);
    font-size: 13px;
    font-family: inherit;
    outline: none;
    cursor: pointer;
    transition: all 0.2s ease;
  }
  .steam-select:focus {
    border-color: var(--color-accent-slate);
    box-shadow: 0 0 0 2px rgba(84, 122, 149, 0.25);
  }

  /* Info Tiles */
  .info-tiles-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }
  @media (max-width: 500px) {
    .info-tiles-grid { grid-template-columns: 1fr; }
  }

  .steam-tile {
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-input);
    padding: 12px 14px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .tile-k { font-size: 10px; font-weight: 700; color: var(--color-text-muted); text-transform: uppercase; letter-spacing: 0.5px; }
  .tile-v { font-family: 'JetBrains Mono', monospace; font-size: 12.5px; color: var(--color-text-white); font-weight: 600; }

  /* Fingerprint Section with PIN Lock */
  .fp-card {
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-input);
    padding: 14px 16px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
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
    color: var(--color-text-light);
    word-break: break-all;
  }
  .fp-value.masked {
    letter-spacing: 3px;
    color: var(--color-text-muted);
  }

  .btn-gold {
    padding: 8px 14px;
    background: linear-gradient(135deg, var(--color-accent-gold) 0%, #A58852 100%);
    border: none;
    border-radius: var(--radius-btn);
    color: #171D25;
    font-size: 12px;
    font-weight: 700;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
    transition: all 0.2s ease;
  }
  .btn-gold:hover {
    filter: brightness(1.1);
    transform: translateY(-1px);
  }

  /* Right Column: Steam-Style QR Box (Primary) */
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
    font-size: 12px;
    font-weight: 800;
    color: var(--color-accent-gold);
    text-transform: uppercase;
    letter-spacing: 1px;
    text-align: center;
  }

  /* Crisp White QR Card (Steam Style) */
  .qr-card-white {
    background: #FFFFFF;
    padding: 18px;
    border-radius: 12px;
    box-shadow: 0 16px 32px rgba(0, 0, 0, 0.45);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: transform 0.2s ease;
  }
  .qr-card-white:hover {
    transform: scale(1.02);
  }
  .qr-card-white svg {
    display: block;
    width: 220px;
    height: 220px;
  }

  .qr-footer-hint {
    font-size: 12px;
    color: var(--color-text-muted);
    text-align: center;
    line-height: 1.5;
  }
  .qr-footer-hint a {
    color: var(--color-accent-gold);
    text-decoration: underline;
    font-weight: 600;
  }

  /* Paired State Card (When phone is active) */
  .paired-state-box {
    width: 100%;
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-indicator-green);
    border-radius: var(--radius-modal);
    padding: 24px 20px;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 14px;
    box-shadow: 0 8px 24px rgba(34, 197, 94, 0.15);
  }
  .paired-state-icon {
    width: 56px;
    height: 56px;
    border-radius: 50%;
    background: var(--color-indicator-green-bg);
    border: 1px solid rgba(34, 197, 94, 0.4);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--color-indicator-green);
  }
  .paired-state-title {
    font-size: 16px;
    font-weight: 700;
    color: var(--color-text-white);
  }
  .paired-state-meta {
    font-size: 12px;
    color: var(--color-text-muted);
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .btn-unpair-danger {
    width: 100%;
    margin-top: 8px;
    padding: 10px 16px;
    border-radius: var(--radius-btn);
    background: var(--color-indicator-red-bg);
    border: 1px solid rgba(239, 68, 68, 0.4);
    color: var(--color-indicator-red);
    font-size: 12.5px;
    font-weight: 700;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    transition: all 0.2s ease;
  }
  .btn-unpair-danger:hover {
    background: rgba(239, 68, 68, 0.25);
    transform: translateY(-1px);
  }

  /* PIN Modal */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.75);
    backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.2s ease;
    z-index: 1000;
  }
  .modal-overlay.active {
    opacity: 1;
    pointer-events: auto;
  }
  .modal-card {
    background: var(--color-bg-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-modal);
    padding: 26px;
    width: 100%;
    max-width: 360px;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.85);
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .modal-card h3 { font-size: 16px; font-weight: 700; color: var(--color-text-white); }
  .modal-card p { font-size: 12.5px; color: var(--color-text-muted); line-height: 1.4; }
  .pin-field {
    width: 100%;
    padding: 12px;
    font-size: 18px;
    text-align: center;
    letter-spacing: 5px;
    border-radius: var(--radius-input);
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
    color: var(--color-text-white);
    outline: none;
  }
  .pin-field:focus { border-color: var(--color-accent-gold); }
  .modal-btns { display: flex; gap: 10px; }
  .btn-modal-cancel {
    flex: 1;
    padding: 10px;
    background: transparent;
    border: 1px solid var(--color-border);
    color: var(--color-text-muted);
    border-radius: var(--radius-btn);
    font-weight: 600;
    cursor: pointer;
  }
  .btn-modal-submit {
    flex: 1;
    padding: 10px;
    background: linear-gradient(135deg, var(--color-accent-gold) 0%, #A58852 100%);
    border: none;
    color: #171D25;
    border-radius: var(--radius-btn);
    font-weight: 700;
    cursor: pointer;
  }
  .error-msg { color: var(--color-indicator-red); font-size: 12px; text-align: center; display: none; }
</style>
<script src="https://cdn.jsdelivr.net/npm/qrcode-generator@1.4.4/qrcode.min.js"></script>
</head>
<body>

<div class="steam-modal">
  <!-- Steam Header -->
  <header class="steam-header">
    <div class="header-brand">
      <div class="steam-logo">
        <svg viewBox="0 0 24 24"><path d="M4 6h16v10H4z M2 18h20v2H2z"/></svg>
      </div>
      <div class="header-title-wrap">
        <h1>PC Remote Connection</h1>
        <p>{{SERVER_NAME}} • Windows PC Service v3.0.0</p>
      </div>
    </div>
    <div class="status-badge {{STATUS_CLASS}}" id="statusBadge">
      <span class="badge-dot"></span>
      <span id="statusText">{{STATUS_TEXT}}</span>
    </div>
  </header>

  <!-- Steam Body -->
  <div class="steam-body">
    <!-- LEFT: Connection Info & Network Controls -->
    <div class="left-col">
      <div>
        <div class="section-label">PENGATURAN KONEKSI SERVER</div>
        <div class="form-group" style="margin-top: 8px;">
          <label class="form-label">Jalur Adapter Jaringan (IP)</label>
          <select class="steam-select" id="networkSelect" onchange="switchAdapter(this.value)">
            <!-- Populated via JS -->
          </select>
        </div>
      </div>

      <!-- Quick Server Info -->
      <div class="info-tiles-grid">
        <div class="steam-tile">
          <span class="tile-k">Server Host</span>
          <span class="tile-v" id="tileHost">{{HOST}}:{{PORT}}</span>
        </div>
        <div class="steam-tile">
          <span class="tile-k">Protokol Keamanan</span>
          <span class="tile-v" style="color: var(--color-indicator-green);">HTTPS (TLS Encrypted)</span>
        </div>
      </div>

      <!-- Certificate Fingerprint (PIN Protected) -->
      <div>
        <div class="section-label">KEAMANAN & SIDIK JARI TLS</div>
        <div class="fp-card">
          <div class="fp-content">
            <span class="tile-k">SHA-256 Fingerprint</span>
            <span class="fp-value masked" id="fpVal">••••••••••••••••••••••••••••••••••••••••</span>
          </div>
          <button class="btn-gold" id="btnUnlockFp" onclick="openPinModal()">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
            Buka Kunci
          </button>
        </div>
      </div>

      <div style="font-size: 12px; color: var(--color-text-muted); line-height: 1.5; margin-top: auto;">
        💡 <strong>Petunjuk:</strong> Buka aplikasi <strong>PC Remote</strong> di HP Anda, lalu arahkan pemindai ke QR Code di sebelah kanan. Sesi akan otomatis tersimpan secara permanen.
      </div>
    </div>

    <!-- RIGHT: QR Code (Primary Steam-Style) -->
    <div class="right-col">
      <!-- When Unpaired: Show QR -->
      <div class="qr-steam-box" id="unpairedView" style="display: {{IS_PAIRED}} ? 'none' : 'flex';">
        <div class="qr-heading">HUBUNGKAN DENGAN QR CODE</div>
        <div class="qr-card-white">
          <div id="qrCanvas"></div>
        </div>
        <div class="qr-footer-hint">
          Gunakan <span style="color: var(--color-text-white); font-weight: 600;">Aplikasi PC Remote</span><br>untuk pairing otomatis via kode QR
        </div>
      </div>

      <!-- When Paired: Show Active Connected Device Card -->
      <div class="paired-state-box" id="pairedView" style="display: {{IS_PAIRED}} ? 'flex' : 'none';">
        <div class="paired-state-icon">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="5" y="2" width="14" height="20" rx="2" ry="2"/><line x1="12" y1="18" x2="12.01" y2="18"/></svg>
        </div>
        <div>
          <div class="paired-state-title" id="pairedDeviceName">{{PAIRED_DEVICE_NAME}}</div>
          <div class="paired-state-meta" style="margin-top: 6px;">
            <span>IP: <strong style="color: #fff;" id="pairedAddr">{{PAIRED_ADDR}}</strong></span>
            <span>Tersambung: <span id="pairedSince">{{PAIRED_SINCE}}</span></span>
          </div>
        </div>
        <div style="font-size: 11px; color: var(--color-indicator-green); background: var(--color-indicator-green-bg); padding: 4px 12px; border-radius: 999px; border: 1px solid rgba(34, 197, 94, 0.4);">
          🔒 Sesi Aktif & Terdaftar
        </div>
        <button class="btn-unpair-danger" onclick="unpairDevice()">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18.36 6.64a9 9 0 1 1-12.73 0"/><line x1="12" y1="2" x2="12" y2="12"/></svg>
          Putuskan Perangkat (Unpair)
        </button>
      </div>
    </div>
  </div>
</div>

<!-- Master PIN Modal -->
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
  const encryptedPayload = "{{ENCRYPTED_PAYLOAD}}";
  const networkInterfaces = {{INTERFACES_JSON}};

  function initPage() {
    const isPaired = isPairedInit;
    document.getElementById('pairedView').style.display = isPaired ? 'flex' : 'none';
    document.getElementById('unpairedView').style.display = isPaired ? 'none' : 'flex';

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

  function renderQR(text) {
    try {
      const qr = qrcode(0, 'M');
      qr.addData(text);
      qr.make();
      document.getElementById('qrCanvas').innerHTML = qr.createSvgTag({ scalable: true, margin: 0 });
      const svg = document.querySelector('#qrCanvas svg');
      if (svg) {
        svg.style.width = '220px';
        svg.style.height = '220px';
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
      }
    } catch (_) {}
  }, 4000);

  initPage();
</script>

</body>
</html>`
